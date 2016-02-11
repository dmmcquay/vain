package vain

import (
	"fmt"
	"net/http"
)

type Server struct {
}

func (s *Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case "GET":
	case "POST":
	case "PATCH":
	default:
		http.Error(w, fmt.Sprintf("unsupported method %q; accepted: POST, GET, PATCH", req.Method), http.StatusMethodNotAllowed)
	}
}

func NewServer(sm *http.ServeMux) *Server {
	s := &Server{}
	sm.Handle("/", s)
	return s
}
