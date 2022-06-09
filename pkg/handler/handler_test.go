package handler

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"url-shortener/storage"
)

var host = "http://localhost:8080/"
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
				host + "bhgaedbedj",
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
				host + "aaaaeeeedr",
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
			},
		},
		{name: "GET ALL",
			want: response{
				code:    200,
				body:    "{\"bhgaedbedj\":\"https://zen.yandex.ru/media/1fx_online/advcash-chto-eto-takoe-i-dlia-kogo-dlia-chego-nujen-etot-elektronnyi-koshelek-6061c1d814931c44e89c923b\"}",
				headers: map[string]string{"Content-Type": "application/json; charset=utf-8"},
			},
			args: args{
				http.MethodGet,
				bytes.NewBuffer([]byte("")),
				host + "all",
			},
		},
	}

	strg := storage.NewMemoryRep()

	h := NewHandler(strg, "localhost:8080")
	serv := h.New()

	//ts := httptest.NewServer(handl)
	//defer ts.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			request := httptest.NewRequest(tt.args.method, tt.args.url, tt.args.body)
			//
			////h := NewHandler(strg, host)
			w := httptest.NewRecorder()
			//
			////handl := http.HandlerFunc(h.ShortenerHandler)
			////handl := h.Init()
			//
			//// запускаем сервер
			serv.ServeHTTP(w, request)
			//res := w.Result()

			//res, body := testRequest(t, ts, tt.args.method, tt.args.url, tt.args.body)

			// проверяем код ответа
			assert.Equal(t, tt.want.code, w.Code)
			//assert.Equal(t, tt.want.code, res.StatusCode)
			//require.Equal(t, tt.want.code, res.StatusCode)
			//if res.StatusCode != tt.want.code {
			//	t.Errorf("Expected status code %d, got %d", tt.want.code, w.Code)
			//}

			//t.Log(res.StatusCode)

			//получаем и проверяем тело запроса
			//defer res.Body.Close()
			//body, err := io.ReadAll(res.Body)
			//if err != nil {
			//	t.Fatal(err)
			//}
			//assert.Equal(t, tt.want.body, string(body))
			assert.Equal(t, tt.want.body, w.Body.String())

			// заголовки ответа
			for key, value := range tt.want.headers {
				assert.Equal(t, value, w.Header().Get(key))
				//if res.Header.Get(key) != value {
				//	t.Errorf("Expected haeder '%s' %s, got %s", key, value, res.Header.Get(key))
				//}
			}
		})
	}
}

func testRequest(t *testing.T, ts *httptest.Server, method, path string, body *bytes.Buffer) (*http.Response, string) {
	req, err := http.NewRequest(method, path, body)
	require.NoError(t, err)
	//req := httptest.NewRequest(method, path, body)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	respBody, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)

	defer resp.Body.Close()

	return resp, string(respBody)
}
