package main

import (
	"log"
	"net/http"
	"url-shortener/pkg/handler"
	"url-shortener/storage"
)

var host = "localhost:8080"

//var UrlCollection storage.Storager //map[string]string

func main() {
	strg := storage.NewMemoryRep()
	handl := handler.NewHandler(strg, host)
	//urlCollection = make(map[string]string)
	http.HandleFunc("/", handl.ShortenerHandler)
	log.Fatal(http.ListenAndServe(host, nil))
	//s := new(Server)
	//log.Fatal(s.Run())
}

//func ShortenerHandler(w http.ResponseWriter, r *http.Request) {
//	if r.Method == http.MethodPost {
//		fmt.Printf("Получен запрос POST %s\n", r.RequestURI)
//		body, err := io.ReadAll(r.Body)
//		if err != nil {
//			http.Error(w, err.Error(), http.StatusBadRequest)
//			return
//		}
//		shortURL, err := storage.AddToCollection(urlCollection, string(body))
//		w.WriteHeader(http.StatusCreated) //201
//		w.Write([]byte("http://" + host + "/" + shortURL))
//	} else if r.Method == http.MethodGet {
//		fmt.Printf("Получен запрос GET %s\n", r.RequestURI)
//		//id := r.URL.Query().Get("id")
//		id := r.URL.Path[1:]
//		if id == "" {
//			http.Error(w, "The query parameter id is missing", http.StatusBadRequest)
//			return
//		}
//		longURL, err := urlCollection.GetByID(id)
//		if err != nil {
//			http.Error(w, err.Error(), http.StatusNotFound)
//			//w.Write([]byte(err.Error()))
//			return
//		}
//		//w.Header().Add("Location", longURL)
//		w.Header().Set("Location", longURL)
//		w.WriteHeader(http.StatusTemporaryRedirect) //307
//	} else {
//		//http.NotFoundHandler()
//		http.Error(w, "Requests not allowed!", http.StatusMethodNotAllowed)
//		return
//	}
//}
