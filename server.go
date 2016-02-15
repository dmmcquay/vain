package vain

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// NewServer populates a server, adds the routes, and returns it for use.
func NewServer(sm *http.ServeMux, store Storage, hostname string) *Server {
	s := &Server{
		storage:  store,
		hostname: hostname,
	}
	sm.Handle("/", s)
	return s
}

// Server serves up the http.
type Server struct {
	hostname string
	storage  Storage
}

func (s *Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case "GET":
		fmt.Fprintf(w, "<!DOCTYPE html>\n<html><head>\n")
		for _, p := range s.storage.All() {
			fmt.Fprintf(w, "%s\n", p)
		}
		fmt.Fprintf(w, "</head>\n</html>\n")
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
			http.Error(w, fmt.Sprintf("invalid repository %q", req.URL.Path), http.StatusBadRequest)
			return
		}
		p.path = fmt.Sprintf("%s/%s", s.hostname, strings.Trim(req.URL.Path, "/"))
		if !Valid(p.path, s.storage.All()) {
			http.Error(w, fmt.Sprintf("invalid path; prefix already taken %q", req.URL.Path), http.StatusConflict)
			return
		}
		if err := s.storage.Add(p); err != nil {
			http.Error(w, fmt.Sprintf("unable to add package: %v", err), http.StatusInternalServerError)
			return
		}
	default:
		http.Error(w, fmt.Sprintf("unsupported method %q; accepted: POST, GET", req.Method), http.StatusMethodNotAllowed)
	}
}
