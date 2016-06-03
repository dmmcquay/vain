package vain

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/mail"
	"strings"
	"time"

	"github.com/elazarl/go-bindata-assetfs"

	verrors "mcquay.me/vain/errors"
	"mcquay.me/vain/static"
)

const apiPrefix = "/api/v0/"
const emailSubject = "your api token"

var prefix map[string]string

func init() {
	prefix = map[string]string{
		"pkgs":     apiPrefix + "db/",
		"register": apiPrefix + "register/",
		"confirm":  apiPrefix + "confirm/",
		"forgot":   apiPrefix + "forgot/",
		"static":   "/_static/",
	}
}

// Server serves up the http.
type Server struct {
	db           *DB
	static       string
	emailTimeout time.Duration
	mail         Mailer
	insecure     bool
}

// NewServer populates a server, adds the routes, and returns it for use.
func NewServer(sm *http.ServeMux, store *DB, m Mailer, static string, emailTimeout time.Duration, insecure bool) *Server {
	s := &Server{
		db:           store,
		static:       static,
		emailTimeout: emailTimeout,
		mail:         m,
		insecure:     insecure,
	}
	addRoutes(sm, s)
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if req.Method == "GET" {
		req.ParseForm()
		if _, ok := req.Form["go-get"]; !ok {
			http.Redirect(w, req, prefix["static"], http.StatusTemporaryRedirect)
			return
		}
		if req.URL.Path == "/" {
			fmt.Fprintf(w, "<!DOCTYPE html>\n<html><head>\n")
			for _, p := range s.db.Pkgs() {
				fmt.Fprintf(w, "%s\n", p)
			}
			fmt.Fprintf(w, "</head>\n<body><p>go tool metadata in head</p></body>\n</html>\n")
		} else {
			p, err := s.db.Package(req.Host + req.URL.Path)
			if err := verrors.ToHTTP(err); err != nil {
				http.Error(w, err.Message, err.Code)
				return
			}
			fmt.Fprintf(w, "<!DOCTYPE html>\n<html><head>\n%s\n</head>\n<body><p>go tool metadata in head</p></body>\n</html>\n", p)
		}
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

	addr, err := mail.ParseAddress(email[0])
	if err != nil {
		http.Error(w, fmt.Sprintf("invalid email detected: %v", err), http.StatusBadRequest)
		return
	}

	tok, err := s.db.Register(addr.Address)
	if err := verrors.ToHTTP(err); err != nil {
		http.Error(w, err.Message, err.Code)
		return
	}

	proto := "https"
	if s.insecure {
		proto = "http"
	}
	resp := struct {
		Msg string `json:"msg"`
	}{
		Msg: "please check your email\n",
	}

	err = s.mail.Send(
		*addr,
		"your api string",
		fmt.Sprintf("%s://%s/api/v0/confirm/%+v", proto, req.Host, tok),
	)
	if err != nil {
		resp.Msg = fmt.Sprintf("problem sending email: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.Header().Set("Content-type", "application/json")
	json.NewEncoder(w).Encode(resp)
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

	addr, err := mail.ParseAddress(email[0])
	if err != nil {
		http.Error(w, fmt.Sprintf("invalid email detected: %v", err), http.StatusBadRequest)
		return
	}

	tok, err := s.db.forgot(addr.Address, s.emailTimeout)
	if err := verrors.ToHTTP(err); err != nil {
		http.Error(w, err.Message, err.Code)
		return
	}
	proto := "https"
	if s.insecure {
		proto = "http"
	}
	resp := struct {
		Msg string `json:"msg"`
	}{
		Msg: "please check your email\n",
	}

	err = s.mail.Send(
		*addr,
		emailSubject,
		fmt.Sprintf("%s://%s/api/v0/confirm/%+v", proto, req.Host, tok),
	)
	if err != nil {
		resp.Msg = fmt.Sprintf("problem sending email: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.Header().Set("Content-type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) pkgs(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-type", "application/json")
	json.NewEncoder(w).Encode(s.db.Pkgs())
}

func addRoutes(sm *http.ServeMux, s *Server) {
	sm.Handle("/", s)

	if s.static == "" {
		sm.Handle(
			prefix["static"],
			http.FileServer(
				&assetfs.AssetFS{
					Asset:     static.Asset,
					AssetDir:  static.AssetDir,
					AssetInfo: static.AssetInfo,
				},
			),
		)
	} else {
		sm.Handle(
			prefix["static"],
			http.StripPrefix(
				prefix["static"],
				http.FileServer(http.Dir(s.static)),
			),
		)
	}

	sm.HandleFunc(prefix["pkgs"], s.pkgs)
	sm.HandleFunc(prefix["register"], s.register)
	sm.HandleFunc(prefix["confirm"], s.confirm)
	sm.HandleFunc(prefix["forgot"], s.forgot)
}
