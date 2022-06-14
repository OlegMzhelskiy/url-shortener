package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"url-shortener/cmd/server"
	"url-shortener/pkg/handler"
	"url-shortener/storage"
)

func main() {

	//host := "localhost:8080"
	//baseUrl := "http://" + host

	host, ok := os.LookupEnv("SERVER_ADDRESS")
	if ok == false || host == "" {
		//panic("Не задано значение переменной окружения SERVER_ADDRESS")
		fmt.Println("Не задано значение переменной окружения SERVER_ADDRESS")
		host = "localhost:8080"
	}
	baseUrl, ok := os.LookupEnv("BASE_URL")
	if ok == false || baseUrl == "" {
		//panic("Не задано значение переменной окружения BASE_URL")
		fmt.Println(("Не задано значение переменной окружения BASE_URL"))
		baseUrl = "http://" + host
	}
	if strings.HasSuffix(baseUrl, "/") == false {
		baseUrl += "/"
	}
	storagePath, ok := os.LookupEnv("FILE_STORAGE_PATH")
	if ok == false || baseUrl == "" {
		fmt.Println(("Не задано значение переменной окружения FILE_STORAGE_PATH"))
	}

	strg := storage.NewMemoryRep(storagePath)
	handl := handler.NewHandler(strg, baseUrl)
	router := handl.New()

	serv := new(server.Server) //нужна ли мне вообще структура Server?
	log.Fatal(serv.Run(host, router))

	//log.Fatal(router.Run(host)) //или можно так запустить?
}
