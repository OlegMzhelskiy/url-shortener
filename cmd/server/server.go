package server

import (
	"net/http"
)

type Server struct {
	httpServer *http.Server
}

func (s Server) Run(host string, h http.Handler) error {
	s.httpServer = &http.Server{
		Addr:    host,
		Handler: h,
	}
	return s.httpServer.ListenAndServe()
}
