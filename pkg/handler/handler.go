package handler

import (
	"compress/gzip"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgerrcode"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
	"io"
	"net/http"
	"strings"
	"time"

	//"url-shortener/cmd/shortener"
	"url-shortener/storage"
)

var secretkey = []byte("It is a secret key")

type Handler struct {
	storage    storage.Storager
	host       string
	dbDSN      string
	timeout    time.Duration
	chanDelURL chan string
	//config *Config
}

type Config struct {
	Host  string
	DBDSN string
}

type Link struct {
	URL string `json:"url"`
}

type gzipWriter struct {
	gin.ResponseWriter
	Writer io.Writer
}

//func NewHandler(s storage.Storager, host string) *Handler {
func NewHandler(s storage.Storager, ch chan string, config *Config) *Handler {
	return &Handler{
		storage:    s,
		host:       config.Host,
		dbDSN:      config.DBDSN,
		timeout:    5,
		chanDelURL: ch,
	}
}

func (h *Handler) NewRouter() *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(gzipHandle())
	router.Use(h.cookiesHandle())

	router.POST("/", h.addLink)
	router.GET("/", h.getEmptyID)
	router.GET("/:id", h.getLinkByID)
	router.GET("/all", h.PrintAll)
	//router.POST("/api/shorten", h.GetShorten)
	//router.GET("/api/user/urls", h.GetUserUrls)
	router.GET("/ping", h.Ping)

	apiGroup := router.Group("/api")
	{
		apiGroup.POST("/shorten", h.GetShorten)
		apiGroup.GET("/user/urls", h.GetUserUrls)
		apiGroup.POST("/shorten/batch", h.GetShortenBatch)
		apiGroup.DELETE("/user/urls", h.deleteURL)
	}
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

	userId, ex := c.Get("userId")
	if !ex {
		http.Error(c.Writer, "Отсутствует user id в контексте", http.StatusNoContent)
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), h.timeout*time.Second)
	defer cancel()

	statusCode := http.StatusCreated
	shortURL, err := storage.AddToCollection(ctx, h.storage, string(body), userId.(string))
	if err != nil {
		if isUniqueViolationError(err) {
			statusCode = http.StatusConflict
		} else {
			fmt.Printf("Ошибка при добавлении в коллекцию: %s \n", err)
			statusCode = http.StatusInternalServerError
		}
	}
	c.Header("Content-Type", "text/html; charset=UTF-8")
	c.String(statusCode, h.host+shortURL)
	//c.IndentedJSON(http.StatusOK, struct{ Status string }{Status: "ok"})
}

func (h *Handler) getLinkByID(c *gin.Context) {
	r := c.Request
	//w := c.Writer
	fmt.Printf("Получен запрос GET %s\n", r.RequestURI)
	id := ""
	if len(r.URL.Path) > 0 {
		id = c.Param("id") //id = c.Params.ByName("id")
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), h.timeout*time.Second)
	defer cancel()

	longURL, err := h.storage.GetByID(ctx, id)
	if err != nil {
		if err == storage.ErrURLDeleted {
			c.Status(http.StatusGone)
			return
		}
		//http.Error(c.Writer, err.Error(), http.StatusNotFound)
		c.String(http.StatusNotFound, err.Error())
		return
	}
	c.Header("Content-Type", "text/html; charset=UTF-8")
	c.Header("Location", longURL)
	//w.WriteHeader(http.StatusTemporaryRedirect) //307
	c.Status(http.StatusTemporaryRedirect)
}

func (h *Handler) GetShorten(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	defer c.Request.Body.Close()
	if err != nil {
		http.Error(c.Writer, err.Error(), http.StatusBadRequest)
		return
	}
	if len(body) == 0 {
		http.Error(c.Writer, "the query must contain a short URL", http.StatusBadRequest)
		return
	}

	var value Link
	if err := json.Unmarshal(body, &value); err != nil {
		http.Error(c.Writer, "error: unmarshal body ", http.StatusInternalServerError) //panic(err)
	}

	userId, ex := c.Get("userId")
	if !ex {
		http.Error(c.Writer, "Отсутствует user id в контексте", http.StatusNoContent)
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), h.timeout*time.Second)
	defer cancel()

	statusCode := http.StatusCreated
	shortURL, err := storage.AddToCollection(ctx, h.storage, value.URL, userId.(string))
	if err != nil {
		if isUniqueViolationError(err) {
			statusCode = http.StatusConflict
		} else {
			http.Error(c.Writer, err.Error(), http.StatusInternalServerError) //http.StatusNotFound)
			return
		}
	}
	result := struct {
		Url string `json:"result"`
	}{h.host + shortURL}
	//json.Marshal(result)
	c.JSON(statusCode, result)
}

func isUniqueViolationError(err error) bool {
	var pqError *pq.Error
	if errors.As(err, &pqError) { //&& errors.Is(err, pqErr3) { //err.Code == pgerrcode.UniqueViolation {
		pqerr, _ := err.(*pq.Error)
		if pqerr.Code == pgerrcode.UniqueViolation {
			return true
		}
	}
	return false
}

func (h *Handler) getEmptyID(c *gin.Context) {
	http.Error(c.Writer, "the query parameter id is missing", http.StatusBadRequest)
	return
}

func (h *Handler) PrintAll(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), h.timeout*time.Second)
	defer cancel()
	c.JSON(http.StatusOK, h.storage.GetAll(ctx))
}

func (w gzipWriter) Write(b []byte) (int, error) {
	// w.Writer будет отвечать за gzip-сжатие, поэтому пишем в него
	return w.Writer.Write(b)
}

func gzipHandle() gin.HandlerFunc {
	return func(c *gin.Context) {
		fmt.Println("Это gzipHandle")

		//Распаковка тела запроса
		if strings.Contains(c.GetHeader("Content-Encoding"), "gzip") {
			fmt.Println("Переданы сжатые данные")
			gz, err := gzip.NewReader(c.Request.Body)
			if err != nil {
				http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
				return
			}
			c.Request.Body = gz

			//defer gz.Close()
			////Распаковываем Body и поомещаем обратно в запрос
			//body, err := io.ReadAll(gz)
			//if err != nil {
			//	http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
			//	return
			//}
			//c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(body))
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
	//return fn //gin.HandlerFunc(fn)
}

func (h *Handler) GetUserUrls(c *gin.Context) {
	userId, ex := c.Get("userId")
	if !ex {
		http.Error(c.Writer, "cookies doesn't content user id", http.StatusNoContent)
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), h.timeout*time.Second)
	defer cancel()

	masURLs := h.storage.GetUserURLs(ctx, userId.(string))
	if len(masURLs) == 0 {
		c.String(http.StatusNoContent, "")
		return
	}
	c.JSON(http.StatusOK, masURLs)
}

func (h *Handler) GetShortenBatch(c *gin.Context) {
	var batch []storage.ElemBatch

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
	if err := json.Unmarshal(body, &batch); err != nil {
		http.Error(c.Writer, "error: unmarshal body ", http.StatusInternalServerError) //panic(err)
	}

	userId, ex := c.Get("userId")
	if !ex {
		http.Error(c.Writer, "cookies doesn't content user id", http.StatusNoContent)
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), h.timeout*time.Second)
	defer cancel()

	err = storage.AddToCollectionBatch(ctx, h.storage, batch, userId.(string))

	//Добавляем хост к идентификатору
	for ind, el := range batch {
		batch[ind].ShortURL = h.host + el.ShortURL
	}

	if err != nil {
		http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusCreated, batch)
}

func (h *Handler) Ping(c *gin.Context) {
	if !h.storage.Ping() {
		c.Status(http.StatusInternalServerError)
		return
	}
	c.Status(http.StatusOK)
}

func (h *Handler) deleteURL(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	defer c.Request.Body.Close()
	if err != nil {
		http.Error(c.Writer, err.Error(), http.StatusBadRequest)
		return
	}
	if len(body) == 0 {
		http.Error(c.Writer, "the query must contain a short URL", http.StatusBadRequest)
		return
	}
	var mas []string //var mas Link
	if err := json.Unmarshal(body, &mas); err != nil {
		http.Error(c.Writer, "error: unmarshal body ", http.StatusInternalServerError) //panic(err)
	}

	userId, ex := c.Get("userId")
	if !ex {
		http.Error(c.Writer, "cookies doesn't content user id", http.StatusNoContent)
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), h.timeout*time.Second)
	defer cancel()

	//получаем все ссылки пользователя и проверяем полученные в запросе на соответствие
	userURLs := h.storage.GetUserMapURLs(ctx, userId.(string))
	for _, el := range mas {
		_, ok := userURLs[el]
		if !ok {
			http.Error(c.Writer, el+" is no exist or belongs to another user", http.StatusBadRequest)
			return
		}
		h.chanDelURL <- el
	}

	//h.storage.deleteURL(ctx, mas)

	c.Status(http.StatusAccepted)
	return
}

func (h *Handler) cookiesHandle() gin.HandlerFunc {
	return func(c *gin.Context) {
		cookieUserID := ""
		userID, idMAC, err := h.getUserIDFromCookie(c)

		ctx, cancel := context.WithTimeout(c.Request.Context(), h.timeout*time.Second)
		defer cancel()

		if err != nil {
			cookieUserID, userID = h.NewCookieUserID(ctx)
		} else {
			//Проверяем полученные куки
			if !ValidMAC([]byte(userID), []byte(idMAC), secretkey) || !h.storage.UserIdIsExist(ctx, userID) {
				cookieUserID, userID = h.NewCookieUserID(ctx)
			}
		}

		if len(cookieUserID) == 0 {
			cookieUserID = userID + "." + idMAC
		}

		c.SetCookie("userId", cookieUserID, 3600, "/", "localhost", false, true)
		c.Set("userId", userID) //Передаем значение userId через контекст запроса
		c.Next()
	}
}

func (h *Handler) getUserIDFromCookie(c *gin.Context) (string, string, error) {

	cookieUserID, err := c.Cookie("userId")
	if err != nil {
		return "", "", errors.New("no cookie user id")
	}

	//Проверяем полученные куки
	str := strings.Split(cookieUserID, ".")
	if len(str) != 2 {
		return "", "", errors.New("incorrect cookie format user id")
	}
	return str[0], str[1], nil
}

//генерируем новый id, подписываем его и устанавливаем в Cookie
func (h *Handler) NewCookieUserID(ctx context.Context) (string, string) {
	id := h.storage.NewUserID(ctx)

	hashf := hmac.New(sha256.New, secretkey)
	hashf.Write([]byte(id))
	hsum := hex.EncodeToString(hashf.Sum(nil))
	valueId := id + "." + hsum

	return valueId, id
}

//Проверка подписи userId
func ValidMAC(message, messageMAC, key []byte) bool {
	mac := hmac.New(sha256.New, key)
	mac.Write(message)
	expectedMAC := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal(messageMAC, []byte(expectedMAC))
}
