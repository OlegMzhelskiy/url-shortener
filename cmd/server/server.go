package server

import (
	"net/http"
)

//var Host = "localhost:8080"
//var UrlCollection storage.Storager

type Server struct {
	httpServer *http.Server
}

func (s *Server) Run(port string, h http.Handler) error {
	//UrlCollection = storage.NewMemoryRep()
	s.httpServer = &http.Server{
		Addr:    ":" + port,
		Handler: h,
	}
	return s.httpServer.ListenAndServe()
}
