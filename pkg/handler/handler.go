package handler

import (
	"encoding/json"
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

//Gin
func (h *Handler) New() *gin.Engine {
	router := gin.New()

	//group := router.Group("/")
	//{
	//	group.POST("", h.addLink)
	//	group.GET("", h.getEmptyID)
	//	group.GET("/:id", h.getLinkByID)
	//	group.GET("/all", h.PrintAll)
	//}

	router.POST("/", h.addLink)
	router.GET("/", h.getEmptyID)
	router.GET("/:id", h.getLinkByID)
	router.GET("/all", h.PrintAll)
	router.POST("/api/shorten", h.GetShorten)

	return router
}

func (h *Handler) addLink(c *gin.Context) {
	r := c.Request
	w := c.Writer
	fmt.Printf("Получен запрос POST %s\n", r.RequestURI)
	body, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if len(body) == 0 {
		http.Error(w, "The query must contain a link", http.StatusBadRequest)
		return
	}
	shortURL, err := storage.AddToCollection(h.storage, string(body))
	if err != nil {
		fmt.Printf("Ошибка при добавлении в коллекцию: %s \n", err)
	}
	w.Header().Set("Content-Type", "text/html; charset=UTF-8")
	//w.WriteHeader(http.StatusCreated) //201
	//w.Write([]byte("http://" + h.host + "/" + shortURL))
	//c.String(http.StatusCreated, "http://"+h.host+"/"+shortURL)
	c.String(http.StatusCreated, h.host+shortURL)

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
		id = c.Param("id") //id = c.Params.ByName("id")
	}
	//if id == "" {
	//	http.Error(w, "The query parameter id is missing", http.StatusBadRequest)
	//	return
	//}
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

func (h *Handler) GetShorten(c *gin.Context) {
	fmt.Printf("Получен запрос POST %s\n", c.Request.RequestURI)
	body, err := io.ReadAll(c.Request.Body)
	defer c.Request.Body.Close()
	if err != nil {
		http.Error(c.Writer, err.Error(), http.StatusBadRequest)
		return
	}
	if len(body) == 0 {
		http.Error(c.Writer, "The query must contain a short URL", http.StatusBadRequest)
		return
	}
	value := struct {
		Url string `json:"url"`
	}{}
	if err := json.Unmarshal(body, &value); err != nil {
		panic(err)
	}
	//r := strings.NewReplacer(h.host+"/", "", "http://", "")
	//idUrl := r.Replace(value.Url)
	//idUrl := strings.Replace(value.Url, h.host+"/", "", 1)
	//longURL, err := h.storage.GetByID(idUrl)

	shortURL, err := storage.AddToCollection(h.storage, value.Url)

	if err != nil {
		http.Error(c.Writer, err.Error(), http.StatusNotFound)
		return
	}
	result := struct {
		Url string `json:"result"`
	}{h.host + shortURL}
	//json.Marshal(result)
	c.JSON(http.StatusCreated, result)
}

func (h *Handler) getEmptyID(c *gin.Context) {
	http.Error(c.Writer, "The query parameter id is missing", http.StatusBadRequest)
	return
}

func (h *Handler) PrintAll(c *gin.Context) {
	c.JSON(http.StatusOK, h.storage.GetAll())
}
