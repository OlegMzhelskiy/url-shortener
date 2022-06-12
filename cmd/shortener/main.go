package main

import (
	"log"
	"os"
	"url-shortener/pkg/handler"
	"url-shortener/storage"
)

//var host string //"localhost:8080"

func main() {
	host, ok := os.LookupEnv("SERVER_ADDRESS")
	if ok == false || host == "" {
		panic("Не задано значение переменной окружения SERVER_ADDRESS")
	}
	baseUrl, ok := os.LookupEnv("BASE_URL")
	if ok == false || baseUrl == "" {
		panic("Не задано значение переменной окружения BASE_URL")
	}

	strg := storage.NewMemoryRep()
	handl := handler.NewHandler(strg, baseUrl)
	router := handl.New()

	//serv := new(server.Server) //нужна ли мне вообще структура Server?
	//log.Fatal(serv.Run("8080", router))

	log.Fatal(router.Run(host))
}
