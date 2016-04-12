package vain

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	verrors "mcquay.me/vain/errors"
)

const apiPrefix = "/api/v0/"

var prefix map[string]string

func init() {
	prefix = map[string]string{
		"pkgs":     apiPrefix + "db/",
		"register": apiPrefix + "register/",
		"confirm":  apiPrefix + "confirm/",
		"forgot":   apiPrefix + "forgot/",
	}
}

// NewServer populates a server, adds the routes, and returns it for use.
func NewServer(sm *http.ServeMux, store *DB) *Server {
	s := &Server{
		db: store,
	}
	addRoutes(sm, s)
	return s
}

// Server serves up the http.
type Server struct {
	db *DB
}

func (s *Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if req.Method == "GET" {
		// TODO: perhaps have a nicely formatted page with info as root if
		// go-get=1 not in request?
		fmt.Fprintf(w, "<!DOCTYPE html>\n<html><head>\n")
		for _, p := range s.db.Pkgs() {
			fmt.Fprintf(w, "%s\n", p)
		}
		fmt.Fprintf(w, "</head>\n<body><p>go tool metadata in head</p></body>\n</html>\n")
		return
	}

	const prefix = "Bearer "
	var tok string
	auth := req.Header.Get("Authorization")
	if strings.HasPrefix(auth, prefix) {
		tok = strings.TrimPrefix(auth, prefix)
	}
	if tok == "" {
		http.Error(w, "missing token", http.StatusUnauthorized)
		return
	}
	ns, err := parseNamespace(req.URL.Path)
	if err != nil {
		http.Error(w, fmt.Sprintf("could not parse namespace:%v", err), http.StatusBadRequest)
		return
	}

	if err := verrors.ToHTTP(s.db.NSForToken(ns, tok)); err != nil {
		http.Error(w, err.Message, err.Code)
		return
	}

	switch req.Method {
	case "POST":
		if req.URL.Path == "/" {
			http.Error(w, fmt.Sprintf("invalid path %q", req.URL.Path), http.StatusBadRequest)
			return
		}
		p := Package{}
		if err := json.NewDecoder(req.Body).Decode(&p); err != nil {
			http.Error(w, fmt.Sprintf("unable to parse json from body: %v", err), http.StatusBadRequest)
			return
		}
		if p.Repo == "" {
			http.Error(w, fmt.Sprintf("invalid repository %q", p.Repo), http.StatusBadRequest)
			return
		}
		if p.Vcs == "" {
			p.Vcs = "git"
		}
		if !valid(p.Vcs) {
			http.Error(w, fmt.Sprintf("invalid vcs %q", p.Vcs), http.StatusBadRequest)
			return
		}
		p.Path = fmt.Sprintf("%s/%s", req.Host, strings.Trim(req.URL.Path, "/"))
		p.Ns = ns
		if !Valid(p.Path, s.db.Pkgs()) {
			http.Error(w, fmt.Sprintf("invalid path; prefix already taken %q", req.URL.Path), http.StatusConflict)
			return
		}
		if err := s.db.AddPackage(p); err != nil {
			http.Error(w, fmt.Sprintf("unable to add package: %v", err), http.StatusInternalServerError)
			return
		}
	case "DELETE":
		p := fmt.Sprintf("%s/%s", req.Host, strings.Trim(req.URL.Path, "/"))
		if !s.db.PackageExists(p) {
			http.Error(w, fmt.Sprintf("package %q not found", p), http.StatusNotFound)
			return
		}

		if err := s.db.RemovePackage(p); err != nil {
			http.Error(w, fmt.Sprintf("unable to delete package: %v", err), http.StatusInternalServerError)
			return
		}
	default:
		http.Error(w, fmt.Sprintf("unsupported method %q; accepted: POST, GET, DELETE", req.Method), http.StatusMethodNotAllowed)
	}
}

func (s *Server) register(w http.ResponseWriter, req *http.Request) {
	req.ParseForm()
	email, ok := req.Form["email"]
	if !ok || len(email) != 1 {
		http.Error(w, "must provide one email parameter", http.StatusBadRequest)
		return
	}
	tok, err := s.db.Register(email[0])
	if err := verrors.ToHTTP(err); err != nil {
		http.Error(w, err.Message, err.Code)
		return
	}
	log.Printf("http://%s/api/v0/confirm/%+v", req.Host, tok)
	fmt.Fprintf(w, "please check your email\n")
}

func (s *Server) confirm(w http.ResponseWriter, req *http.Request) {
	tok := req.URL.Path[len(prefix["confirm"]):]
	tok = strings.TrimRight(tok, "/")
	if tok == "" {
		http.Error(w, "must provide one email parameter", http.StatusBadRequest)
		return
	}
	tok, err := s.db.Confirm(tok)
	if err := verrors.ToHTTP(err); err != nil {
		http.Error(w, err.Message, err.Code)
		return
	}
	fmt.Fprintf(w, "new token: %s\n", tok)
}

func (s *Server) forgot(w http.ResponseWriter, req *http.Request) {
	req.ParseForm()
	email, ok := req.Form["email"]
	if !ok || len(email) != 1 {
		http.Error(w, "must provide one email parameter", http.StatusBadRequest)
		return
	}
	tok, err := s.db.forgot(email[0])
	if err := verrors.ToHTTP(err); err != nil {
		http.Error(w, err.Message, err.Code)
		return
	}
	log.Printf("http://%s/api/v0/confirm/%+v", req.Host, tok)
	fmt.Fprintf(w, "please check your email\n")
}
func (s *Server) pkgs(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-type", "application/json")
	json.NewEncoder(w).Encode(s.db.Pkgs())
}

func addRoutes(sm *http.ServeMux, s *Server) {
	sm.Handle("/", s)

	sm.HandleFunc(prefix["pkgs"], s.pkgs)
	sm.HandleFunc(prefix["register"], s.register)
	sm.HandleFunc(prefix["confirm"], s.confirm)
	sm.HandleFunc(prefix["forgot"], s.forgot)
}
