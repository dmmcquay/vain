package vain

import "net/http"

type Server struct {
}

func NewServer(sm *http.ServeMux) *Server {
	s := &Server{}
	addRoutes(sm, s)
	return s
}

func addRoutes(sm *http.ServeMux, s *Server) {
}
