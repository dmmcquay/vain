package vain

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type Server struct {
	hostname string
	storage  *MemStore
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
			http.Error(w, fmt.Sprintf("unable to parse json from body: %v", err), http.StatusInternalServerError)
			return
		}
		p.Path = fmt.Sprintf("%s/%s", s.hostname, strings.Trim(req.URL.Path, "/"))
		if !Valid(p.Path, s.storage.All()) {
			http.Error(w, fmt.Sprintf("invalid path; prefix already taken %q", req.URL.Path), http.StatusConflict)
			return
		}
		s.storage.Add(p)
	default:
		http.Error(w, fmt.Sprintf("unsupported method %q; accepted: POST, GET", req.Method), http.StatusMethodNotAllowed)
	}
}

func NewServer(sm *http.ServeMux, ms *MemStore, hostname string) *Server {
	s := &Server{
		storage:  ms,
		hostname: hostname,
	}
	sm.Handle("/", s)
	return s
}
