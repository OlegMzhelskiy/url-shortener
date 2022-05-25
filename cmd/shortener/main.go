package main

import (
	"errors"
	"fmt"
	"hash/crc32"
	"io"
	"log"
	"net/http"
	"strconv"
)

var host = "localhost:8080"
var urlCollection map[string]string

func main() {
	urlCollection = make(map[string]string)
	http.HandleFunc("/", ShortenerHandler)
	log.Fatal(http.ListenAndServe(host, nil))
}

func ShortenerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		fmt.Printf("Получен запрос POST %s\n", r.RequestURI)
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		shortURL := addToCollection(string(body))
		w.WriteHeader(http.StatusCreated) //201
		w.Write([]byte("http://" + host + "/" + shortURL))
	} else if r.Method == http.MethodGet {
		fmt.Printf("Получен запрос GET %s\n", r.RequestURI)
		//id := r.URL.Query().Get("id")
		id := r.URL.Path[1:]
		if id == "" {
			http.Error(w, "The query parameter id is missing", http.StatusBadRequest)
			return
		}
		longURL, err := getURLByID(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			//w.Write([]byte(err.Error()))
			return
		}
		//w.Header().Add("Location", longURL)
		w.Header().Set("Location", longURL)
		w.WriteHeader(http.StatusTemporaryRedirect) //307
	} else {
		//http.NotFoundHandler()
		http.Error(w, "Requests not allowed!", http.StatusMethodNotAllowed)
		return
	}
}

func addToCollection(longURL string) string {
	shortURL := generateIdentify(longURL)
	//Проверим идентификатор на наличие в колекции
	_, ok := urlCollection[shortURL]
	if !ok {
		urlCollection[shortURL] = longURL
	}
	return shortURL
}

func getURLByID(id string) (string, error) {
	val, ok := urlCollection[id]
	if !ok {
		return "", errors.New("not found")
	}
	return val, nil
}

func generateIdentify(s string) string {
	h := crc32.NewIEEE()
	h.Write([]byte(s))
	hashSum := h.Sum32()
	sHash := strconv.Itoa(int(hashSum))
	masRune := []rune(sHash)
	for i, value := range masRune {
		//masRune[i] = rune(int(value) + 49 + rand.Intn(15))
		masRune[i] = value + 49
	}
	res := string(masRune)
	return res
}
