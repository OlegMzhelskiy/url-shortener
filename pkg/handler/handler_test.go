package handler

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"url-shortener/storage"
)

var filePath = ""
var host = "http://localhost:8080/"
var baseUrl = "http://" + host + "/"
var url1 = "https://zen.yandex.ru/media/1fx_online/advcash-chto-eto-takoe-i-dlia-kogo-dlia-chego-nujen-etot-elektronnyi-koshelek-6061c1d814931c44e89c923b"
var short1 = host + "bhgaedbedj"
var url2 = "https://proglib.io/p/go-programming"
var short2 = host + "bdifachdif"
var dbDSN = ""

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
	Id  string `json:"correlation_id"`
	Url string `json:"original_url"`
}

type testCase struct {
	name string
	args args
	want response
}

func TestHandler_ShortenerHandler(t *testing.T) {
	//type fields struct {
	//	storage storage.Storager
	//}

	var handl *Handler
	configHandler := &Config{host, dbDSN}
	configStore := &storage.StoreConfig{baseUrl, dbDSN, filePath}

	store := storage.ConfigurateStorage(configStore)
	defer store.Close()
	handl = NewHandler(store, configHandler)

	router := handl.NewRouter()

	//Init testcases
	pingStat := http.StatusInternalServerError
	if store.Ping() {
		pingStat = http.StatusOK
	}

	b := []elemBatch{
		{"1ya", "https://practicum.yandex.ru/learn/go-developer/courses/9908027e-ac38-4005-a7c9-30f61f5ed23f/sprints/51370/topics/dd5c3680-6603-4f17-957a-6991147bf14c/lessons/e7f410af-7304-4a6e-9c7f-6e109813e16f/"},
		{"2cgo", "https://habr.com/ru/company/intel/blog/275709/"},
	}
	bodyBatch, _ := json.Marshal(b)

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
		{name: "GET Error",
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
		{name: "GET Empty ID",
			want: response{
				code:    400,
				body:    "The query parameter id is missing\n",
				headers: map[string]string{"Content-Type": "text/plain; charset=utf-8"},
			},
			args: args{
				http.MethodGet,
				new(bytes.Buffer),
				host,
				map[string]string{},
			},
		},
		{name: "GET Empty Request",
			want: response{
				code:    400,
				body:    "The query parameter id is missing\n",
				headers: map[string]string{"Content-Type": "text/plain; charset=utf-8"},
			},
			args: args{
				http.MethodGet,
				new(bytes.Buffer),
				host,
				map[string]string{},
			},
		},
		//{name: "GET ALL",
		//	want: response{
		//		code:    200,
		//		body:    "{\"bhgaedbedj\":\"https://zen.yandex.ru/media/1fx_online/advcash-chto-eto-takoe-i-dlia-kogo-dlia-chego-nujen-etot-elektronnyi-koshelek-6061c1d814931c44e89c923b\"}",
		//		headers: map[string]string{"Content-Type": "application/json; charset=utf-8"},
		//	},
		//	args: args{
		//		http.MethodGet,
		//		bytes.NewBuffer([]byte("")),
		//		host + "all",
		//		map[string]string{},
		//	},
		//},
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
		{name: "GET /api/user/urls",
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
		{name: "GET /api/shorten/batch",
			want: response{
				code:    201,
				body:    "[{\"correlation_id\":\"1ya\",\"original_url\":\"https://practicum.yandex.ru/learn/go-developer/courses/9908027e-ac38-4005-a7c9-30f61f5ed23f/sprints/51370/topics/dd5c3680-6603-4f17-957a-6991147bf14c/lessons/e7f410af-7304-4a6e-9c7f-6e109813e16f/\",\"short_url\":\"ghafjfgeb\"},{\"correlation_id\":\"2cgo\",\"original_url\":\"https://habr.com/ru/company/intel/blog/275709/\",\"short_url\":\"badbgeicic\"}]",
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

	cookie := &http.Cookie{
		Name:  "userId",
		Value: "270cc75709f72a3b3457d838fcc4c5a4.d9aedd6479de522652e585ddb308c2ce3842db212ed2be5584e6c9e6d7fc076f",
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

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
		})
	}
}

//func testRequest(t *testing.T, ts *httptest.Server, method, path string, body *bytes.Buffer) (*http.Response, string) {
//	req, err := http.NewRequest(method, path, body)
//	require.NoError(t, err)
//	//req := httptest.NewRequest(method, path, body)
//
//	resp, err := http.DefaultClient.Do(req)
//	require.NoError(t, err)
//
//	respBody, err := ioutil.ReadAll(resp.Body)
//	require.NoError(t, err)
//
//	defer resp.Body.Close()
//
//	return resp, string(respBody)
//}

//func TestHandler_Ping(t *testing.T) {
//	//var handl *Handler
//	//configHandler := &Config{host, dbDSN}
//	//configStore := &storage.StoreConfig{baseUrl, dbDSN, filePath}
//	//
//	//store := storage.ConfigurateStorage(configStore)
//	//defer store.Close()
//	//handl = NewHandler(store, configHandler)
//	//router := handl.NewRouter()
//	//
//	//t.Run("Ping", func(t *testing.T) {
//	//	request := httptest.NewRequest("GET", "", bytes.NewBuffer([]byte("")))
//	//
//	//	w := httptest.NewRecorder()
//	//
//	//	router.ServeHTTP(w, request)
//	//
//	//	assert.Equal(t, 200, w.Code)
//	//})
//
//
//
//}

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
