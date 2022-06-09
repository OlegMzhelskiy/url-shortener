package main

import (
	"log"
	"url-shortener/pkg/handler"
	"url-shortener/storage"
)

var host = "localhost:8080"

func main() {
	strg := storage.NewMemoryRep()
	handl := handler.NewHandler(strg, host)
	router := handl.New()

	//serv := new(server.Server) //нужна ли мне вообще структура Server?
	//log.Fatal(serv.Run("8080", router))

	log.Fatal(router.Run(host))
}
