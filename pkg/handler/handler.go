package handler

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"strings"

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
func (h *Handler) NewRouter() *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(gzipHandle())

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
	//w.Header().Set("Content-Type", "text/html; charset=UTF-8")
	c.Header("Content-Type", "text/html; charset=UTF-8")
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
	c.Header("Content-Type", "text/html; charset=UTF-8")
	c.Header("Location", longURL)
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

type gzipWriter struct {
	gin.ResponseWriter
	Writer io.Writer
}

func (w gzipWriter) Write(b []byte) (int, error) {
	// w.Writer будет отвечать за gzip-сжатие, поэтому пишем в него
	return w.Writer.Write(b)
}

type gzipReader struct {
}

func gzipHandle() gin.HandlerFunc {
	fn := func(c *gin.Context) {
		fmt.Println("Это gzipHandle")

		//Распаковка тела запроса
		if strings.Contains(c.GetHeader("Content-Encoding"), "gzip") {
			fmt.Println("Переданы сжатые данные")
			gz, err := gzip.NewReader(c.Request.Body)
			if err != nil {
				http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
				return
			}
			//defer gz.Close()
			////Распаковываем Body и поомещаем обратно в запрос
			//body, err := io.ReadAll(gz)
			//if err != nil {
			//	http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
			//	return
			//}
			//c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(body))
			c.Request.Body = gz
		}

		//Сжатие результата
		if strings.Contains(c.GetHeader("Accept-Encoding"), "gzip") {
			fmt.Println("Поддерживает сжатие")
			gz, err := gzip.NewWriterLevel(c.Writer, gzip.BestSpeed)
			if err != nil {
				io.WriteString(c.Writer, err.Error())
				return
			}
			defer gz.Close()
			c.Writer = gzipWriter{ResponseWriter: c.Writer, Writer: gz}
			c.Header("Content-Encoding", "gzip")
		}
		//next.ServeHTTP(c)
		c.Next()
	}
	return fn //gin.HandlerFunc(fn)
}

// Compress сжимает слайс байт.
func Compress(data []byte) ([]byte, error) {
	var b bytes.Buffer
	// создаём переменную w — в неё будут записываться входящие данные,
	// которые будут сжиматься и сохраняться в bytes.Buffer
	w := gzip.NewWriter(&b)
	// запись данных
	_, err := w.Write(data)
	if err != nil {
		return nil, fmt.Errorf("failed write data to compress temporary buffer: %v", err)
	}
	// обязательно нужно вызвать метод Close() — в противном случае часть данных
	// может не записаться в буфер b; если нужно выгрузить все упакованные данные
	// в какой-то момент сжатия, используйте метод Flush()
	err = w.Close()
	if err != nil {
		return nil, fmt.Errorf("failed compress data: %v", err)
	}
	// переменная b содержит сжатые данные
	return b.Bytes(), nil
}

// Decompress распаковывает слайс байт.
func Decompress(data []byte) ([]byte, error) {
	// переменная r будет читать входящие данные и распаковывать их
	r := flate.NewReader(bytes.NewReader(data))
	defer r.Close()

	var b bytes.Buffer
	// в переменную b записываются распакованные данные
	_, err := b.ReadFrom(r)
	if err != nil {
		return nil, fmt.Errorf("failed decompress data: %v", err)
	}

	return b.Bytes(), nil
}
