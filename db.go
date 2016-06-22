package vain

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/jmoiron/sqlx"
	// for side effects
	_ "github.com/mattn/go-sqlite3"

	verrors "mcquay.me/vain/errors"
	vsql "mcquay.me/vain/sql"
)

// DB wraps a sqlx.DB connection and provides methods for interating with
// a vain database.
type DB struct {
	conn *sqlx.DB
}

// NewDB opens a sqlite3 file, sets options, and reports errors.
func NewDB(path string) (*DB, error) {
	conn, err := sqlx.Open("sqlite3", fmt.Sprintf("file:%s?cache=shared&mode=rwc", path))
	if _, err := conn.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return nil, err
	}
	return &DB{conn}, err
}

// Init runs the embedded sql to initialize tables.
func (db *DB) Init() error {
	content, err := vsql.Asset("sql/init.sql")
	if err != nil {
		return err
	}
	_, err = db.conn.Exec(string(content))
	return err
}

// Close the underlying connection.
func (db *DB) Close() error {
	return db.conn.Close()
}

// AddPackage adds p into packages table.
func (db *DB) AddPackage(p Package) error {
	_, err := db.conn.NamedExec(
		"INSERT INTO packages(vcs, repo, path, ns) VALUES (:vcs, :repo, :path, :ns)",
		&p,
	)
	return err
}

// RemovePackage removes package with given path
func (db *DB) RemovePackage(path string) error {
	_, err := db.conn.Exec("DELETE FROM packages WHERE path = ?", path)
	return err
}

// Pkgs returns all packages from the database
func (db *DB) Pkgs() []Package {
	r := []Package{}
	rows, err := db.conn.Queryx("SELECT * FROM packages")
	if err != nil {
		log.Printf("%+v", err)
		return nil
	}
	for rows.Next() {
		var p Package
		err = rows.StructScan(&p)
		if err != nil {
			log.Printf("%+v", err)
			return nil
		}
		r = append(r, p)
	}
	return r
}

// PackageExists tells if a package with path is in the database.
func (db *DB) PackageExists(path string) bool {
	var count int
	if err := db.conn.Get(&count, "SELECT COUNT(*) FROM packages WHERE path = ?", path); err != nil {
		log.Printf("%+v", err)
	}

	r := false
	switch count {
	case 1:
		r = true
	default:
		log.Printf("unexpected count of packages matching %q: %d", path, count)
	}
	return r
}

// Package fetches the package associated with path.
func (db *DB) Package(path string) (Package, error) {
	r := Package{}
	err := db.conn.Get(&r, "SELECT * FROM packages WHERE path = ?", path)
	if err == sql.ErrNoRows {
		return r, verrors.HTTP{
			Message: fmt.Sprintf("couldn't find package %q", path),
			Code:    http.StatusNotFound,
		}
	}
	return r, err
}

// NSForToken creates an entry namespaces with a relation to the token.
func (db *DB) NSForToken(ns string, tok string) error {
	var err error
	txn, err := db.conn.Beginx()
	if err != nil {
		return verrors.HTTP{
			Message: fmt.Sprintf("problem creating transaction: %v", err),
			Code:    http.StatusInternalServerError,
		}
	}
	defer func() {
		if err != nil {
			txn.Rollback()
		} else {
			txn.Commit()
		}
	}()

	var count int
	if err = txn.Get(&count, "SELECT COUNT(*) FROM namespaces WHERE namespaces.ns = ?", ns); err != nil {
		return verrors.HTTP{
			Message: fmt.Sprintf("problem matching fetching namespaces matching %q", ns),
			Code:    http.StatusInternalServerError,
		}
	}

	if count == 0 {
		var email string
		if err = txn.Get(&email, "SELECT email FROM users WHERE token = $1", tok); err != nil {
			return verrors.HTTP{
				Message: fmt.Sprintf("could not find user for token %q", tok),
				Code:    http.StatusInternalServerError,
			}
		}
		if _, err = txn.Exec(
			"INSERT INTO namespaces(ns, email) VALUES ($1, $2)",
			ns,
			email,
		); err != nil {
			return verrors.HTTP{
				Message: fmt.Sprintf("problem inserting %q into namespaces for token %q: %v", ns, tok, err),
				Code:    http.StatusInternalServerError,
			}
		}
		return err
	}

	if err = txn.Get(&count, "SELECT COUNT(*) FROM namespaces JOIN users ON namespaces.email = users.email WHERE users.token = ? AND namespaces.ns = ?", tok, ns); err != nil {
		return verrors.HTTP{
			Message: fmt.Sprintf("ns: %q, tok: %q; %v", ns, tok, err),
			Code:    http.StatusInternalServerError,
		}
	}

	switch count {
	case 1:
		err = nil
	case 0:
		err = verrors.HTTP{
			Message: fmt.Sprintf("not authorized against namespace %q", ns),
			Code:    http.StatusUnauthorized,
		}
	default:
		err = verrors.HTTP{
			Message: fmt.Sprintf("inconsistent db; found %d results with ns (%s) with token (%s)", count, ns, tok),
			Code:    http.StatusInternalServerError,
		}
	}
	return err
}

// Register adds email to the database, returning an error if there was one.
func (db *DB) Register(email string) (string, error) {
	var err error
	txn, err := db.conn.Beginx()
	if err != nil {
		return "", verrors.HTTP{
			Message: fmt.Sprintf("problem creating transaction: %v", err),
			Code:    http.StatusInternalServerError,
		}
	}
	defer func() {
		if err != nil {
			txn.Rollback()
		} else {
			txn.Commit()
		}
	}()

	var count int
	if err = txn.Get(&count, "SELECT COUNT(*) FROM users WHERE email = ?", email); err != nil {
		return "", verrors.HTTP{
			Message: fmt.Sprintf("could not search for email %q in db: %v", email, err),
			Code:    http.StatusInternalServerError,
		}
	}

	if count != 0 {
		return "", verrors.HTTP{
			Message: fmt.Sprintf("duplicate email %q", email),
			Code:    http.StatusConflict,
		}
	}

	tok := FreshToken()
	_, err = txn.Exec(
		"INSERT INTO users(email, token, requested) VALUES (?, ?, ?)",
		email,
		tok,
		time.Now(),
	)
	return tok, err
}

// Confirm  modifies the user with the given token. Used on register confirmation.
func (db *DB) Confirm(token string) (string, error) {
	var err error
	txn, err := db.conn.Beginx()
	if err != nil {
		return "", verrors.HTTP{
			Message: fmt.Sprintf("problem creating transaction: %v", err),
			Code:    http.StatusInternalServerError,
		}
	}
	defer func() {
		if err != nil {
			txn.Rollback()
		} else {
			txn.Commit()
		}
	}()

	var count int
	if err = txn.Get(&count, "SELECT COUNT(*) FROM users WHERE token = ?", token); err != nil {
		return "", verrors.HTTP{
			Message: fmt.Sprintf("could not perform search for user with token %q in db: %v", token, err),
			Code:    http.StatusInternalServerError,
		}
	}

	if count != 1 {
		return "", verrors.HTTP{
			Message: fmt.Sprintf("bad token: %s", token),
			Code:    http.StatusNotFound,
		}
	}

	newToken := FreshToken()

	_, err = txn.Exec(
		"UPDATE users SET token = ?, registered = 1 WHERE token = ?",
		newToken,
		token,
	)
	if err != nil {
		return "", verrors.HTTP{
			Message: fmt.Sprintf("couldn't update user with token %q: %v", token, err),
			Code:    http.StatusInternalServerError,
		}
	}
	return newToken, nil
}

func (db *DB) forgot(email string, window time.Duration) (string, error) {
	txn, err := db.conn.Beginx()
	if err != nil {
		return "", verrors.HTTP{
			Message: fmt.Sprintf("problem creating transaction: %v", err),
			Code:    http.StatusInternalServerError,
		}
	}
	defer func() {
		if err != nil {
			txn.Rollback()
		} else {
			txn.Commit()
		}
	}()

	out := struct {
		Token     string
		Requested time.Time
	}{}
	if err = txn.Get(&out, "SELECT token, requested FROM users WHERE email = ?", email); err != nil {
		return "", verrors.HTTP{
			Message: fmt.Sprintf("could not find email %q in db", email),
			Code:    http.StatusNotFound,
		}
	}

	if out.Requested.After(time.Now()) {
		return "", verrors.HTTP{
			Message: fmt.Sprintf("rate limit hit for %q; try again in %0.2f mins", email, out.Requested.Sub(time.Now()).Minutes()),
			Code:    http.StatusTooManyRequests,
		}
	}
	_, err = txn.Exec("UPDATE users SET requested = ? WHERE email = ?", time.Now().Add(window), email)
	if err != nil {
		return "", verrors.HTTP{
			Message: fmt.Sprintf("could not update last requested time for %q: %v", email, err),
			Code:    http.StatusInternalServerError,
		}
	}
	return out.Token, nil
}

func (db *DB) addUser(email string) (string, error) {
	tok := FreshToken()
	_, err := db.conn.Exec(
		"INSERT INTO users(email, token, requested) VALUES (?, ?, ?)",
		email,
		tok,
		time.Now(),
	)
	return tok, err
}

func (db *DB) user(email string) (User, error) {
	u := User{}
	err := db.conn.Get(
		&u,
		"SELECT email, token, registered, requested FROM users WHERE email = ?",
		email,
	)
	if err == sql.ErrNoRows {
		return User{}, verrors.HTTP{
			Message: fmt.Sprintf("could not find requested user's email: %q: %v", email, err),
			Code:    http.StatusNotFound,
		}
	}
	return u, err
}
