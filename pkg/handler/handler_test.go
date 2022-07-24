package handler

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"url-shortener/storage"
)

var filePath = ""
var host = "http://localhost:8080/"
var baseURL = "http://" + host + "/"
var url1 = "https://zen.yandex.ru/media/1fx_online/advcash-chto-eto-takoe-i-dlia-kogo-dlia-chego-nujen-etot-elektronnyi-koshelek-6061c1d814931c44e89c923b"
var short1 = host + "bhgaedbedj"
var url2 = "https://proglib.io/p/go-programming"
var short2 = host + "bdifachdif"
var dbDSN = ""
var batch = []elemBatch{
	{"1ya", "https://practicum.yandex.ru/learn/go-developer/courses/9908027e-ac38-4005-a7c9-30f61f5ed23f/sprints/51370/topics/dd5c3680-6603-4f17-957a-6991147bf14c/lessons/e7f410af-7304-4a6e-9c7f-6e109813e16f/"},
	{"2cgo", "https://habr.com/ru/company/intel/blog/275709/"},
}

var cookie = &http.Cookie{
	Name:  "userId",
	Value: "270cc75709f72a3b3457d838fcc4c5a4.d9aedd6479de522652e585ddb308c2ce3842db212ed2be5584e6c9e6d7fc076f",
}

type args struct {
	method  string
	body    *bytes.Buffer
	url     string
	headers map[string]string
}
type response struct {
	code    int
	body    string
	headers map[string]string
}

type elemBatch struct {
	ID  string `json:"correlation_id"`
	URL string `json:"original_url"`
}

type testCase struct {
	name string
	args args
	want response
}

func TestHandler_addLink(t *testing.T) {

	var handl *Handler
	configHandler := &Config{host, dbDSN}
	configStore := &storage.StoreConfig{BaseURL: baseURL, DBDNS: dbDSN, FilestoragePath: filePath}

	ch := make(chan *storage.UserArrayURL, 100)

	store := storage.ConfigurateStorage(configStore)
	defer store.Close()
	handl = NewHandler(store, ch, configHandler)

	router := handl.NewRouter()

	//bodyBatch, _ := json.Marshal(batch)

	tests := []testCase{
		{name: "POST",
			want: response{
				code: 201,
				//body: "http://localhost:8080/bhgaedbedj",
				body: short1,
				//contentType: "application/json",
				headers: map[string]string{
					"Content-Type": "text/html; charset=UTF-8",
				},
			},
			args: args{
				http.MethodPost,
				bytes.NewBuffer([]byte(url1)),
				host,
				map[string]string{},
			},
		},
		{name: "POST Empty Body",
			want: response{
				code: 400,
				body: "The query must contain a link\n",
				//contentType: "application/json",
				headers: map[string]string{},
			},
			args: args{
				http.MethodPost,
				bytes.NewBuffer([]byte("")),
				host,
				map[string]string{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testRequest(t, tt, router, cookie)
		})
		//t.Run(tt.name, func(t *testing.T) {
		//
		//	request := httptest.NewRequest(tt.args.method, tt.args.url, tt.args.body)
		//	for key, value := range tt.args.headers {
		//		request.Header.Add(key, value)
		//	}
		//	w := httptest.NewRecorder()
		//
		//	request.AddCookie(cookie)
		//
		//	// запускаем сервер
		//	router.ServeHTTP(w, request)
		//
		//	// проверяем код ответа
		//	assert.Equal(t, tt.want.code, w.Code)
		//
		//	//получаем и проверяем тело запроса
		//	var body string
		//	if strings.Contains(w.Header().Get("Content-Encoding"), "gzip") {
		//		bodyByte, err := Decompress(w.Body.Bytes())
		//		//if !assert.Error(t, err, "Ошибка декомпрессии тела ответа") {
		//		//	body = string(bodyByte)
		//		//}
		//		if err != nil {
		//			assert.Fail(t, "Ошибка декомпрессии: "+err.Error())
		//		} else {
		//			body = string(bodyByte)
		//		}
		//	} else {
		//		body = w.Body.String()
		//	}
		//	//body := w.Body.String()
		//	assert.Equal(t, tt.want.body, body) //w.Body.String()
		//
		//	// заголовки ответа
		//	for key, value := range tt.want.headers {
		//		assert.Equal(t, value, w.Header().Get(key))
		//	}
		//})
	}
}

func TestHandler_Ping(t *testing.T) {

	var handl *Handler
	configHandler := &Config{host, dbDSN}
	configStore := &storage.StoreConfig{BaseURL: baseURL, DBDNS: dbDSN, FilestoragePath: filePath}

	ch := make(chan *storage.UserArrayURL, 100)

	store := storage.ConfigurateStorage(configStore)
	defer store.Close()
	handl = NewHandler(store, ch, configHandler)

	router := handl.NewRouter()

	pingStat := http.StatusInternalServerError
	if store.Ping() {
		pingStat = http.StatusOK
	}

	tests := []testCase{
		{name: "GET Ping",
			want: response{
				code:    pingStat,
				body:    "",
				headers: map[string]string{},
			},
			args: args{
				http.MethodGet,
				bytes.NewBuffer([]byte("")),
				host + "ping",
				map[string]string{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testRequest(t, tt, router, cookie)
		})
	}
}

func TestHandler_getLinkByID(t *testing.T) {

	var handl *Handler
	configHandler := &Config{host, dbDSN}
	configStore := &storage.StoreConfig{BaseURL: baseURL, DBDNS: dbDSN, FilestoragePath: filePath}

	ch := make(chan *storage.UserArrayURL, 100)

	store := storage.ConfigurateStorage(configStore)
	defer store.Close()
	handl = NewHandler(store, ch, configHandler)

	router := handl.NewRouter()

	tests := []testCase{
		{name: "POST",
			want: response{
				code: 201,
				//body: "http://localhost:8080/bhgaedbedj",
				body: short1,
				//contentType: "application/json",
				headers: map[string]string{
					"Content-Type": "text/html; charset=UTF-8",
				},
			},
			args: args{
				http.MethodPost,
				bytes.NewBuffer([]byte(url1)),
				host,
				map[string]string{},
			},
		},
		{name: "GET",
			want: response{
				code: 307,
				body: "",
				headers: map[string]string{
					"Content-Type": "text/html; charset=UTF-8",
					"Location":     url1},
			},
			args: args{
				http.MethodGet,
				new(bytes.Buffer),
				host + "bhgaedbedj",
				map[string]string{},
			},
		},
		{name: "GET non-existent URL",
			want: response{
				code:    404,
				body:    "not found",
				headers: map[string]string{},
			},
			args: args{
				http.MethodGet,
				new(bytes.Buffer),
				host + "aaaaeeeedr",
				map[string]string{},
			},
		},
		{name: "GET Empty Request",
			want: response{
				code:    400,
				body:    "the query parameter id is missing\n",
				headers: map[string]string{"Content-Type": "text/plain; charset=utf-8"},
			},
			args: args{
				http.MethodGet,
				new(bytes.Buffer),
				host,
				map[string]string{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testRequest(t, tt, router, cookie)
		})
	}
}

func TestHandler_GetShorten(t *testing.T) {

	var handl *Handler
	configHandler := &Config{host, dbDSN}
	configStore := &storage.StoreConfig{BaseURL: baseURL, DBDNS: dbDSN, FilestoragePath: filePath}

	ch := make(chan *storage.UserArrayURL, 100)

	store := storage.ConfigurateStorage(configStore)
	defer store.Close()
	handl = NewHandler(store, ch, configHandler)

	router := handl.NewRouter()

	tests := []testCase{
		{name: "POST /api/shorten",
			want: response{
				code: 201,
				body: `{"result":"` + short1 + `"}`,
				headers: map[string]string{
					"Content-Type": "application/json; charset=utf-8",
				},
			},
			args: args{
				http.MethodPost,
				//bytes.NewBuffer([]byte(`{"url":"https://zen.yandex.ru/media/1fx_online/advcash-chto-eto-takoe-i-dlia-kogo-dlia-chego-nujen-etot-elektronnyi-koshelek-6061c1d814931c44e89c923b"}`)),
				bytes.NewBuffer([]byte(`{"url":"` + url1 + `"}`)),
				host + "api/shorten",
				map[string]string{},
			},
		},
		{name: "POST /api/shorten gzip-json",
			want: response{
				code: 201,
				body: `{"result":"` + short2 + `"}`,
				headers: map[string]string{
					"Content-Type": "application/json; charset=utf-8",
				},
			},
			args: args{
				http.MethodPost,
				//bytes.NewBuffer([]byte(`{"url":"https://zen.yandex.ru/media/1fx_online/advcash-chto-eto-takoe-i-dlia-kogo-dlia-chego-nujen-etot-elektronnyi-koshelek-6061c1d814931c44e89c923b"}`)),
				bytes.NewBuffer(Compress([]byte(`{"url":"` + url2 + `"}`))),
				host + "api/shorten",
				map[string]string{
					"Content-Encoding": "gzip",
				},
			},
		},
		{name: "POST /api/shorten gzip-gzip",
			want: response{
				code: 201,
				body: `{"result":"` + short2 + `"}`,
				headers: map[string]string{
					"Content-Type":     "application/json; charset=utf-8",
					"Content-Encoding": "gzip",
				},
			},
			args: args{
				http.MethodPost,
				//bytes.NewBuffer([]byte(`{"url":"https://zen.yandex.ru/media/1fx_online/advcash-chto-eto-takoe-i-dlia-kogo-dlia-chego-nujen-etot-elektronnyi-koshelek-6061c1d814931c44e89c923b"}`)),
				bytes.NewBuffer(Compress([]byte(`{"url":"` + url2 + `"}`))),
				host + "api/shorten",
				map[string]string{
					"Content-Encoding": "gzip",
					"Accept-Encoding":  "gzip",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testRequest(t, tt, router, cookie)
		})
	}
}

func TestHandler_GetUserUrls(t *testing.T) {

	var handl *Handler
	configHandler := &Config{host, dbDSN}
	configStore := &storage.StoreConfig{BaseURL: baseURL, DBDNS: dbDSN, FilestoragePath: filePath}

	ch := make(chan *storage.UserArrayURL, 100)

	store := storage.ConfigurateStorage(configStore)
	defer store.Close()
	handl = NewHandler(store, ch, configHandler)

	router := handl.NewRouter()

	bodyBatch, _ := json.Marshal(batch)

	tests := []testCase{
		{name: "GET /api/user/urls Empty",
			want: response{
				code:    204,
				body:    "", //"[]"
				headers: map[string]string{
					//	"Content-Type": "application/json; charset=utf-8",
				},
			},
			args: args{
				http.MethodGet,
				bytes.NewBuffer([]byte("")),
				host + "api/user/urls",
				map[string]string{},
			},
		},
		//Добавляем данные для тестирования
		{name: "---GET /api/shorten/batch",
			want: response{
				code:    201,
				body:    "[{\"correlation_id\":\"1ya\",\"original_url\":\"https://practicum.yandex.ru/learn/go-developer/courses/9908027e-ac38-4005-a7c9-30f61f5ed23f/sprints/51370/topics/dd5c3680-6603-4f17-957a-6991147bf14c/lessons/e7f410af-7304-4a6e-9c7f-6e109813e16f/\",\"short_url\":\"http://localhost:8080/ghafjfgeb\"},{\"correlation_id\":\"2cgo\",\"original_url\":\"https://habr.com/ru/company/intel/blog/275709/\",\"short_url\":\"http://localhost:8080/badbgeicic\"}]",
				headers: map[string]string{
					//	"Content-Type": "application/json; charset=utf-8",
				},
			},
			args: args{
				http.MethodPost,
				bytes.NewBuffer(bodyBatch),
				host + "api/shorten/batch",
				map[string]string{},
			},
		},
		{name: "GET /api/user/urls",
			want: response{
				code:    200,
				body:    `[{"short_url":"http://http://localhost:8080//ghafjfgeb","original_url":"https://practicum.yandex.ru/learn/go-developer/courses/9908027e-ac38-4005-a7c9-30f61f5ed23f/sprints/51370/topics/dd5c3680-6603-4f17-957a-6991147bf14c/lessons/e7f410af-7304-4a6e-9c7f-6e109813e16f/"},{"short_url":"http://http://localhost:8080//badbgeicic","original_url":"https://habr.com/ru/company/intel/blog/275709/"}]`,
				headers: map[string]string{
					//	"Content-Type": "application/json; charset=utf-8",
				},
			},
			args: args{
				http.MethodGet,
				bytes.NewBuffer([]byte("")),
				host + "api/user/urls",
				map[string]string{},
			},
		},
	}

	newCookie := &http.Cookie{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			//testRequest(t, tt, router, newCookie)
			newCookie = testRequestCookie(t, tt, router, newCookie)
		})
	}
}

func TestHandler_GetShortenBatch(t *testing.T) {

	var handl *Handler
	configHandler := &Config{host, dbDSN}
	configStore := &storage.StoreConfig{BaseURL: baseURL, DBDNS: dbDSN, FilestoragePath: filePath}

	//ch := make(chan string, 100)
	ch := make(chan *storage.UserArrayURL, 100)

	store := storage.ConfigurateStorage(configStore)
	defer store.Close()
	handl = NewHandler(store, ch, configHandler)

	router := handl.NewRouter()

	bodyBatch, _ := json.Marshal(batch)

	tests := []testCase{
		{name: "GET /api/shorten/batch",
			want: response{
				code:    201,
				body:    "[{\"correlation_id\":\"1ya\",\"original_url\":\"https://practicum.yandex.ru/learn/go-developer/courses/9908027e-ac38-4005-a7c9-30f61f5ed23f/sprints/51370/topics/dd5c3680-6603-4f17-957a-6991147bf14c/lessons/e7f410af-7304-4a6e-9c7f-6e109813e16f/\",\"short_url\":\"http://localhost:8080/ghafjfgeb\"},{\"correlation_id\":\"2cgo\",\"original_url\":\"https://habr.com/ru/company/intel/blog/275709/\",\"short_url\":\"http://localhost:8080/badbgeicic\"}]",
				headers: map[string]string{
					//	"Content-Type": "application/json; charset=utf-8",
				},
			},
			args: args{
				http.MethodPost,
				bytes.NewBuffer(bodyBatch),
				host + "api/shorten/batch",
				map[string]string{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testRequest(t, tt, router, cookie)
		})
	}
}

func TestHandler_PrintAll(t *testing.T) {

	var handl *Handler
	configHandler := &Config{host, dbDSN}
	configStore := &storage.StoreConfig{BaseURL: baseURL, DBDNS: dbDSN, FilestoragePath: filePath}

	ch := make(chan *storage.UserArrayURL, 100)

	store := storage.ConfigurateStorage(configStore)
	defer store.Close()
	handl = NewHandler(store, ch, configHandler)

	router := handl.NewRouter()

	bodyBatch, _ := json.Marshal(batch)

	tests := []testCase{
		{name: "---", //вставка данных для следующего кейса
			want: response{
				code:    201,
				body:    "[{\"correlation_id\":\"1ya\",\"original_url\":\"https://practicum.yandex.ru/learn/go-developer/courses/9908027e-ac38-4005-a7c9-30f61f5ed23f/sprints/51370/topics/dd5c3680-6603-4f17-957a-6991147bf14c/lessons/e7f410af-7304-4a6e-9c7f-6e109813e16f/\",\"short_url\":\"http://localhost:8080/ghafjfgeb\"},{\"correlation_id\":\"2cgo\",\"original_url\":\"https://habr.com/ru/company/intel/blog/275709/\",\"short_url\":\"http://localhost:8080/badbgeicic\"}]",
				headers: map[string]string{
					//	"Content-Type": "application/json; charset=utf-8",
				},
			},
			args: args{
				http.MethodPost,
				bytes.NewBuffer(bodyBatch),
				host + "api/shorten/batch",
				map[string]string{},
			},
		},
		{name: "GET ALL",
			want: response{
				code:    200,
				body:    `{"badbgeicic":{"originUrl":"https://habr.com/ru/company/intel/blog/275709/","userId":"5502d0741bd614878a8815c4930b0686","deleted":false},"ghafjfgeb":{"originUrl":"https://practicum.yandex.ru/learn/go-developer/courses/9908027e-ac38-4005-a7c9-30f61f5ed23f/sprints/51370/topics/dd5c3680-6603-4f17-957a-6991147bf14c/lessons/e7f410af-7304-4a6e-9c7f-6e109813e16f/","userId":"5502d0741bd614878a8815c4930b0686","deleted":false}}`,
				headers: map[string]string{"Content-Type": "application/json; charset=utf-8"},
			},
			args: args{
				http.MethodGet,
				bytes.NewBuffer([]byte("")),
				host + "all",
				map[string]string{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			request := httptest.NewRequest(tt.args.method, tt.args.url, tt.args.body)
			for key, value := range tt.args.headers {
				request.Header.Add(key, value)
			}
			w := httptest.NewRecorder()

			//request.AddCookie(cookie)

			// запускаем сервер
			router.ServeHTTP(w, request)

			if tt.name == "---" {
				return
			}

			// проверяем код ответа
			assert.Equal(t, tt.want.code, w.Code)

			var masbat map[string]storage.UserURL
			var expect map[string]storage.UserURL

			//получаем и проверяем тело запроса
			var bodyByte []byte
			var err error
			if strings.Contains(w.Header().Get("Content-Encoding"), "gzip") {
				bodyByte, err = Decompress(w.Body.Bytes())
				//if !assert.Error(t, err, "Ошибка декомпрессии тела ответа") {
				//	body = string(bodyByte)
				//}
				if err != nil {
					assert.Fail(t, "Ошибка декомпрессии: "+err.Error())
				}
			} else {
				bodyByte = w.Body.Bytes()
			}

			err = json.Unmarshal(bodyByte, &masbat)
			if err != nil {
				assert.Fail(t, "Ошибка декодирования: "+err.Error())
			}

			err = json.Unmarshal([]byte(tt.want.body), &expect)
			if err != nil {
				assert.Fail(t, "Ошибка декодирования: "+err.Error())
			}

			//проверяем элементы мап
			for key, val := range expect {
				el := masbat[key]
				assert.Equal(t, el.OriginURL, val.OriginURL)
			}

			// заголовки ответа
			for key, value := range tt.want.headers {
				assert.Equal(t, value, w.Header().Get(key))
			}

		})
	}
}

func testRequest(t *testing.T, tt testCase, router *gin.Engine, cookie *http.Cookie) {
	request := httptest.NewRequest(tt.args.method, tt.args.url, tt.args.body)
	for key, value := range tt.args.headers {
		request.Header.Add(key, value)
	}
	w := httptest.NewRecorder()

	request.AddCookie(cookie)

	// запускаем сервер
	router.ServeHTTP(w, request)

	// проверяем код ответа
	assert.Equal(t, tt.want.code, w.Code)

	//получаем и проверяем тело запроса
	var body string
	if strings.Contains(w.Header().Get("Content-Encoding"), "gzip") {
		bodyByte, err := Decompress(w.Body.Bytes())
		//if !assert.Error(t, err, "Ошибка декомпрессии тела ответа") {
		//	body = string(bodyByte)
		//}
		if err != nil {
			assert.Fail(t, "Ошибка декомпрессии: "+err.Error())
		} else {
			body = string(bodyByte)
		}
	} else {
		body = w.Body.String()
	}
	//body := w.Body.String()
	assert.Equal(t, tt.want.body, body) //w.Body.String()

	// заголовки ответа
	for key, value := range tt.want.headers {
		assert.Equal(t, value, w.Header().Get(key))
	}
}

func testRequestCookie(t *testing.T, tt testCase, router *gin.Engine, pCookie *http.Cookie) *http.Cookie {
	request := httptest.NewRequest(tt.args.method, tt.args.url, tt.args.body)
	for key, value := range tt.args.headers {
		request.Header.Add(key, value)
	}
	w := httptest.NewRecorder()

	request.AddCookie(pCookie)

	// запускаем сервер
	router.ServeHTTP(w, request)

	// проверяем код ответа
	assert.Equal(t, tt.want.code, w.Code)

	//получаем и проверяем тело запроса
	var body string
	if strings.Contains(w.Header().Get("Content-Encoding"), "gzip") {
		bodyByte, err := Decompress(w.Body.Bytes())
		//if !assert.Error(t, err, "Ошибка декомпрессии тела ответа") {
		//	body = string(bodyByte)
		//}
		if err != nil {
			assert.Fail(t, "Ошибка декомпрессии: "+err.Error())
		} else {
			body = string(bodyByte)
		}
	} else {
		body = w.Body.String()
	}
	//body := w.Body.String()
	assert.Equal(t, tt.want.body, body) //w.Body.String()

	// заголовки ответа
	for key, value := range tt.want.headers {
		assert.Equal(t, value, w.Header().Get(key))
	}

	cooStr := w.Header().Get("Set-Cookie")
	mas := strings.Split(cooStr, ";")
	for _, elmas := range mas {
		if strings.HasPrefix(elmas, "userId=") {
			pc := strings.Split(elmas, "=")
			newCookie := &http.Cookie{
				Name:  "userId",
				Value: pc[1],
			}
			return newCookie
		}
	}
	return nil
}

// Compress сжимает слайс байт.
func Compress(data []byte) []byte {
	var b bytes.Buffer
	// создаём переменную w — в неё будут записываться входящие данные,
	// которые будут сжиматься и сохраняться в bytes.Buffer
	w := gzip.NewWriter(&b)
	// запись данных
	_, err := w.Write(data)
	if err != nil {
		return nil //, fmt.Errorf("failed write data to compress temporary buffer: %v", err)
	}
	// обязательно нужно вызвать метод Close() — в противном случае часть данных
	// может не записаться в буфер b; если нужно выгрузить все упакованные данные
	// в какой-то момент сжатия, используйте метод Flush()
	err = w.Close()
	if err != nil {
		return nil //, fmt.Errorf("failed compress data: %v", err)
	}
	// переменная b содержит сжатые данные
	return b.Bytes() //, nil
}

// Decompress распаковывает слайс байт.
func Decompress(data []byte) ([]byte, error) {
	// переменная r будет читать входящие данные и распаковывать их
	r, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer r.Close()

	var b bytes.Buffer
	// в переменную b записываются распакованные данные
	_, err = b.ReadFrom(r)
	if err != nil {
		return nil, fmt.Errorf("failed decompress data: %v", err)
	}

	return b.Bytes(), nil
}
