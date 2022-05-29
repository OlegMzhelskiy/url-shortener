package handler

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"url-shortener/storage"
)

var host = "localhost:8080"
var url1 = "https://zen.yandex.ru/media/1fx_online/advcash-chto-eto-takoe-i-dlia-kogo-dlia-chego-nujen-etot-elektronnyi-koshelek-6061c1d814931c44e89c923b"

func TestHandler_ShortenerHandler(t *testing.T) {
	type fields struct {
		storage storage.Storager
	}
	type args struct {
		method string
		body   *bytes.Buffer
		url    string
	}
	type response struct {
		code    int
		body    string
		headers map[string]string
	}

	tests := []struct {
		name string
		args args
		want response
	}{
		{name: "POST",
			want: response{
				code: 201,
				body: "http://localhost:8080/bhgaedbedj",
				//contentType: "application/json",
				headers: map[string]string{
					"Content-Type": "text/html; charset=UTF-8",
				},
			},
			args: args{
				http.MethodPost,
				bytes.NewBuffer([]byte(url1)),
				host,
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
				"http://localhost:8080/bhgaedbedj",
			},
		},
		{name: "GET Error",
			want: response{
				code:    404,
				body:    "not found\n",
				headers: map[string]string{},
			},
			args: args{
				http.MethodGet,
				new(bytes.Buffer),
				"http://localhost:8080/aaaaeeeedr",
			},
		},
	}

	strg := storage.NewMemoryRep()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			request := httptest.NewRequest(tt.args.method, tt.args.url, tt.args.body)

			h := NewHandler(strg, host)
			w := httptest.NewRecorder()

			handl := http.HandlerFunc(h.ShortenerHandler)

			// запускаем сервер
			handl.ServeHTTP(w, request)
			res := w.Result()

			// проверяем код ответа
			assert.Equal(t, tt.want.code, res.StatusCode)
			//if res.StatusCode != tt.want.code {
			//	t.Errorf("Expected status code %d, got %d", tt.want.code, w.Code)
			//}

			//t.Log(res.StatusCode)

			// получаем и проверяем тело запроса
			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, tt.want.body, string(resBody))
			//if string(resBody) != tt.want.body {
			//	t.Errorf("Expected body %s, got %s", tt.want.body, w.Body.String())
			//}

			// заголовки ответа
			for key, value := range tt.want.headers {
				assert.Equal(t, value, res.Header.Get(key))
				//if res.Header.Get(key) != value {
				//	t.Errorf("Expected haeder '%s' %s, got %s", key, value, res.Header.Get(key))
				//}
			}
		})
	}
}
