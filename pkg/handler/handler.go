package handler

import (
	"fmt"
	"github.com/gin-gonic/gin"
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
		strbody := string(body)
		if len(strbody) == 0 {
			http.Error(w, "The query must contain a link", http.StatusBadRequest)
			return
		}
		shortURL, err := storage.AddToCollection(h.storage, strbody)
		w.Header().Set("Content-Type", "text/html; charset=UTF-8")
		w.WriteHeader(http.StatusCreated) //201
		w.Write([]byte("http://" + h.host + "/" + shortURL))
	} else if r.Method == http.MethodGet {
		fmt.Printf("Получен запрос GET %s\n", r.RequestURI)

		id := ""
		if len(r.URL.Path) > 0 {
			//id := r.URL.Query().Get("id")
			id = r.URL.Path[1:]
		}
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
		w.Header().Set("Content-Type", "text/html; charset=UTF-8")
		w.Header().Set("Location", longURL)
		w.WriteHeader(http.StatusTemporaryRedirect) //307
	} else {
		//http.NotFoundHandler()
		http.Error(w, "Requests not allowed!", http.StatusMethodNotAllowed)
		return
	}
}

//Gin
func (h *Handler) Init() *gin.Engine {
	router := gin.New()
	//router.GET("", h.getEmptyID)
	group := router.Group("/")
	{
		group.POST("", h.addLink)
		group.GET("", h.getEmptyID)
		group.GET("/:id", h.getLinkByID)
		group.GET("/all", h.PrintAll)
	}
	return router
}

func (h *Handler) addLink(c *gin.Context) {
	r := c.Request
	w := c.Writer
	fmt.Printf("Получен запрос POST %s\n", r.RequestURI)
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	strbody := string(body)
	if len(strbody) == 0 {
		http.Error(w, "The query must contain a link", http.StatusBadRequest)
		return
	}
	shortURL, err := storage.AddToCollection(h.storage, strbody)
	w.Header().Set("Content-Type", "text/html; charset=UTF-8")
	w.WriteHeader(http.StatusCreated) //201
	w.Write([]byte("http://" + h.host + "/" + shortURL))

	fmt.Println("Это Post")
	//c.Writer.Write([]byte("<h1>Привет это POST!</h1>"))
	//c.IndentedJSON(http.StatusOK, struct{ Status string }{Status: "ok"})
}

func (h *Handler) getLinkByID(c *gin.Context) {
	r := c.Request
	w := c.Writer
	fmt.Printf("Получен запрос GET %s\n", r.RequestURI)
	id := ""
	if len(r.URL.Path) > 0 {
		id = c.Params.ByName("id")
	}
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
	w.Header().Set("Content-Type", "text/html; charset=UTF-8")
	w.Header().Set("Location", longURL)
	w.WriteHeader(http.StatusTemporaryRedirect) //307
}

func (h *Handler) getEmptyID(c *gin.Context) {
	http.Error(c.Writer, "The query parameter id is missing", http.StatusBadRequest)
	return
}

func (h *Handler) PrintAll(c *gin.Context) {
	c.JSON(http.StatusOK, h.storage.GetAll())
}
