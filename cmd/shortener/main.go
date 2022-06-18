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

	var host string
	var baseUrl string
	var storagePath string

	flagHost := flag.String("a", "", "server address")        //SERVER_ADDRESS
	flagBaseUrl := flag.String("b", "", "base url")           //BASE_URL
	flagFilePath := flag.String("f", "", "file storage path") //FILE_STORAGE_PATH
	flag.Parse()

	//Проверка корректности заполнения
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
	host = getVarValue(*flagHost, "SERVER_ADDRESS", "localhost:8080")
	baseUrl = getVarValue(*flagBaseUrl, "BASE_URL", "http://"+host)
	storagePath = getVarValue(*flagFilePath, "FILE_STORAGE_PATH", "")

	//if len(*flagHost) == 0 {
	//	host, ok = os.LookupEnv("SERVER_ADDRESS")
	//	if ok == false || host == "" {
	//		fmt.Println("Не задано значение переменной окружения SERVER_ADDRESS")
	//		host = "localhost:8080"
	//	}
	//} else {
	//	host = *flagHost
	//}
	//
	//if len(*flagBaseUrl) == 0 {
	//	baseUrl, ok = os.LookupEnv("BASE_URL")
	//	if ok == false || baseUrl == "" {
	//		fmt.Println(("Не задано значение переменной окружения BASE_URL"))
	//		baseUrl = "http://" + host
	//	}
	//} else {
	//	baseUrl = *flagBaseUrl
	//}
	//
	//if len(*flagFilePath) == 0 {
	//	storagePath, ok = os.LookupEnv("FILE_STORAGE_PATH") //URLdata.json
	//	if ok == false || storagePath == "" {
	//		fmt.Println(("Не задано значение переменной окружения FILE_STORAGE_PATH"))
	//	}
	//} else {
	//	storagePath = *flagFilePath
	//}

	if strings.HasSuffix(baseUrl, "/") == false {
		baseUrl += "/"
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

//Проверяет заполнение флага и получает значение переменной окружения если он не заполнен
func getVarValue(flagValue, envVarName, defValue string) string {
	var ok bool
	varVal := flagValue
	if len(flagValue) == 0 {
		varVal, ok = os.LookupEnv(envVarName) //URLdata.json
		if ok == false || varVal == "" {
			fmt.Printf("Не задано значение переменной окружения %s\n", envVarName)
			varVal = defValue
		}
	}
	return varVal
}
