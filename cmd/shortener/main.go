package main

import (
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"log"
	"os"
	"strings"
	"url-shortener/cmd/server"
	"url-shortener/pkg/handler"
	"url-shortener/storage"
)

func main() {

	if !gin.IsDebugging() {
		gin.SetMode(gin.ReleaseMode)
	}

	//host := "localhost:8080"
	//baseUrl := "http://" + host

	var host string
	var baseURL string
	var storagePath string
	var dbDSN string

	flagHost := flag.String("a", "", "server address")        //SERVER_ADDRESS
	flagBaseURL := flag.String("b", "", "base url")           //BASE_URL
	flagFilePath := flag.String("f", "", "file storage path") //FILE_STORAGE_PATH
	flagDBDSN := flag.String("d", "", "DB connection")        //DATABASE_DSN
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
	baseURL = getVarValue(*flagBaseURL, "BASE_URL", "http://"+host)
	storagePath = getVarValue(*flagFilePath, "FILE_STORAGE_PATH", "")
	dbDSN = getVarValue(*flagDBDSN, "DATABASE_DSN", "host=localhost dbname=shortener user=postgres password=123 sslmode=disable")

	if !strings.HasSuffix(baseURL, "/") {
		baseURL += "/"
	}

	var handl *handler.Handler
	configHandler := &handler.Config{baseURL, dbDSN}
	configStore := &storage.StoreConfig{baseURL, dbDSN, storagePath}

	store := storage.ConfigurateStorage(configStore)
	defer store.Close()
	handl = handler.NewHandler(store, configHandler)

	//postgreDB, err := storage.NewStoreDB(configStore)
	//if err != nil || postgreDB.Ping() == false {
	//	memoryDB := storage.NewMemoryRep(storagePath, baseUrl)
	//	handl = handler.NewHandler(memoryDB, configHandler)
	//} else {
	//	defer postgreDB.Close()
	//	handl = handler.NewHandler(postgreDB, configHandler)
	//}

	router := handl.NewRouter()

	fmt.Printf("Host: %s\n", host)
	//defer strg.WriteRepoFromFile() //Запишем в файл по завершении работы сервера

	srv := new(server.Server) //нужна ли мне вообще структура Server?
	log.Fatal(srv.Run(host, router))

	//log.Fatal(router.Run(host)) //или можно так запустить?
}

//Проверяет заполнение флага и получает значение переменной окружения если он не заполнен
func getVarValue(flagValue, envVarName, defValue string) string {
	var ok bool
	varVal := flagValue
	if len(flagValue) == 0 {
		varVal, ok = os.LookupEnv(envVarName) //URLdata.json
		if !ok || varVal == "" {
			fmt.Printf("Не задано значение переменной окружения %s\n", envVarName)
			varVal = defValue
		}
	}
	return varVal
}

func Configurate() {

}

//func getStorageDB(storagePath, baseUrl string) *storage.Storager{
//storageDB, err := storage.NewStoreDB(dbDSN)
//if err != nil || storageDB.Ping() == false {
//	storageDB := storage.NewMemoryRep(storagePath, baseUrl)
//}
//	return storageDB
//}
