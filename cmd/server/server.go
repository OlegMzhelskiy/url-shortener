package server

import (
	"net/http"
)

type Server struct {
	httpServer *http.Server
}

func (s Server) Run(port string, h http.Handler) error {
	s.httpServer = &http.Server{
		Addr:    ":" + port,
		Handler: h,
	}
	return s.httpServer.ListenAndServe()
}
