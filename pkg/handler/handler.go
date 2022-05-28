package handler

import (
	"fmt"
	"io"
	"net/http"
	//"url-shortener/cmd/shortener"
	"url-shortener/storage"
)

type Handler struct {
	storage storage.Storager
	host    string
}

func NewHandler(s storage.Storager, host string) *Handler {
	return &Handler{storage: s, host: host}
}

func (h *Handler) ShortenerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		fmt.Printf("Получен запрос POST %s\n", r.RequestURI)
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		shortURL, err := storage.AddToCollection(h.storage, string(body))
		w.WriteHeader(http.StatusCreated) //201
		w.Write([]byte("http://" + h.host + "/" + shortURL))
	} else if r.Method == http.MethodGet {
		fmt.Printf("Получен запрос GET %s\n", r.RequestURI)
		//id := r.URL.Query().Get("id")
		id := r.URL.Path[1:]
		if id == "" {
			http.Error(w, "The query parameter id is missing", http.StatusBadRequest)
			return
		}
		longURL, err := h.storage.GetByID(id)
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
