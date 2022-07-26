package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"log"
	"os"
	"strings"
	"time"
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

	//ch := make(chan string, 100)
	ch := make(chan *storage.UserArrayURL, 100)

	var h *handler.Handler
	configHandler := &handler.Config{Host: baseURL, DBDSN: dbDSN}
	configStore := &storage.StoreConfig{BaseURL: baseURL, DBDNS: dbDSN, FilestoragePath: storagePath}

	store := storage.ConfigurateStorage(configStore)
	defer store.Close()
	h = handler.NewHandler(store, ch, configHandler)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go DeleteUserArrayURL(ctx, ch, store, 5)

	router := h.NewRouter()

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

func deleteURLs(ctx context.Context, store storage.Storager, masID []string, dur time.Duration) {
	err := store.DeleteURLs(ctx, masID)
	if err != nil {
		fmt.Println(err.Error())
	}
	time.Sleep(dur)
}

func DeleteUserArrayURL(ctx context.Context, ch chan *storage.UserArrayURL, store storage.Storager, intervalMin int) {
	dur := time.Duration(intervalMin) * time.Minute
	//ticker := time.NewTicker(dur)
	var masID []string
	for {
		select {
		case <-ctx.Done():
			return
		case arr, ok := <-ch:
			if ok {
				//получаем все ссылки пользователя и проверяем полученные в запросе на соответствие
				userURLs, err := store.GetUserMapURLs(ctx, arr.UserID)
				if err != nil {
					continue
				}
				for _, el := range arr.ArrayURL {
					_, ok := userURLs[el]
					if !ok {
						fmt.Printf("%s is no exist or belongs to another user\n", el)
						continue
					}
				}
				masID = append(masID, arr.ArrayURL...)
			} else {
				if len(masID) > 0 {
					deleteURLs(ctx, store, masID, dur)
					masID = []string{} //обнуляем слайс
				}
			}
		default:
			if len(masID) > 0 {
				deleteURLs(ctx, store, masID, dur)
				masID = []string{} //обнуляем слайс
			}
		}
	}
}
