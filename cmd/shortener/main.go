package main

import (
	"flag"
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

	//Значения по умолчанию
	var (
		host        = "localhost:8080"
		baseUrl     = "http://" + host
		storagePath = ""
	)

	flagHost := flag.String("a", "", "a string")     //SERVER_ADDRESS
	flagBaseUrl := flag.String("b", "", "b string")  //BASE_URL
	flagFilePath := flag.String("f", "", "f string") //FILE_STORAGE_PATH
	flag.Parse()

	//Проверка корректности заоплнения
	if len(*flagHost) > 0 {
		//Если задан только порт в формате :8080
		if strings.HasPrefix(*flagHost, ":") {
			*flagHost = "localhost" + *flagHost
		} else {
			st := strings.Split(*flagHost, ":")
			if len(st) == 1 {
				*flagHost = st[0] + ":" + "8080"
			}
		}
	}
	if len(*flagHost) == 0 {
		host, ok := os.LookupEnv("SERVER_ADDRESS")
		if ok == false || host == "" {
			//panic("Не задано значение переменной окружения SERVER_ADDRESS")
			fmt.Println("Не задано значение переменной окружения SERVER_ADDRESS")
			//host = "localhost:8080"
		}
	} else {
		host = *flagHost
	}

	if len(*flagBaseUrl) == 0 {
		baseUrl, ok := os.LookupEnv("BASE_URL")
		if ok == false || baseUrl == "" {
			//panic("Не задано значение переменной окружения BASE_URL")
			fmt.Println(("Не задано значение переменной окружения BASE_URL"))
			//baseUrl = "http://" + host
		}
	} else {
		baseUrl = *flagBaseUrl
	}

	if strings.HasSuffix(baseUrl, "/") == false {
		baseUrl += "/"
	}

	storagePath, ok := os.LookupEnv("FILE_STORAGE_PATH") //URLdata.json
	if ok == false || baseUrl == "" {
		fmt.Println(("Не задано значение переменной окружения FILE_STORAGE_PATH"))
	}
	if len(*flagFilePath) > 0 {
		storagePath = *flagFilePath
	}

	strg := storage.NewMemoryRep(storagePath)
	handl := handler.NewHandler(strg, baseUrl)
	router := handl.New()

	fmt.Printf("Host: %s\n", host)
	//defer strg.WriteRepoFromFile() //Запишем в файл по завершении работы сервера

	serv := new(server.Server) //нужна ли мне вообще структура Server?
	log.Fatal(serv.Run(host, router))

	//log.Fatal(router.Run(host)) //или можно так запустить?
}
