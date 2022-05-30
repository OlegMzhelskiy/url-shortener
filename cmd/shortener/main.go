package main

import (
	"log"
	"url-shortener/cmd/server"
	"url-shortener/pkg/handler"
	"url-shortener/storage"
)

var host = "localhost:8080"

func main() {
	strg := storage.NewMemoryRep()
	handl := handler.NewHandler(strg, host)

	//http.HandleFunc("/", handl.ShortenerHandler)
	//log.Fatal(http.ListenAndServe(host, nil))

	serv := new(server.Server)
	log.Fatal(serv.Run("8080", handl.Init()))
}
