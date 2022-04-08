package main

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AlekseyKas/metrics/cmd/server/handlers"
	"github.com/AlekseyKas/metrics/internal/storage"
	"github.com/fatih/structs"
	"github.com/go-chi/chi"
	"github.com/stretchr/testify/require"
)

var body []byte

func TestRouter(t *testing.T) {

	// var MapMetrics map[string]interface{} = structs.Map(storage.Metrics{})
	s := &storage.MetricsStore{
		MM: structs.Map(storage.Metrics{}),
	}
	handlers.SetStorage(s)
	type want struct {
		contentType string
		statusCode  int
	}

	tests := []struct {
		name   string
		url    string
		method string
		body   []byte
		want   want
	}{
		{
			name:   "fist sample#",
			url:    "/value/",
			method: "POST",
			body:   []byte(`{"ID": "Alloc", "type": "gauge"}`),
			want: want{
				contentType: "application/json",
				statusCode:  200,
			},
			// },
			// {
			// 	name:    "second sample#",
			// 	request: "/update/gauge/TestCount/200",
			// 	method:  "POST",
			// 	want: want{
			// 		contentType: "application/json",
			// 		statusCode:  200,
			// 	},
			// },
			// {
			// 	name:    "third sample#",
			// 	request: "/update/unknown/TestCount/200",
			// 	method:  "POST",
			// 	want: want{
			// 		contentType: "application/json",
			// 		statusCode:  501,
			// 	},
			// },
			// {
			// 	name:    "third sample#",
			// 	request: "/value/gauge/Alloc",
			// 	method:  "GET",
			// 	want: want{
			// 		contentType: "application/json",
			// 		statusCode:  200,
			// 	},
			// },
			// {
			// 	name:    "third sample#",
			// 	request: "/",
			// 	method:  "GET",
			// 	want: want{
			// 		contentType: "application/json",
			// 		statusCode:  200,
			// 	},
			// },
			// {
			// 	name:    "third sample#",
			// 	request: "/value/gauge/PollCount",
			// 	method:  "GET",
			// 	want: want{
			// 		contentType: "application/json",
			// 		statusCode:  404,
			// 	},
			// },
			// {
			// 	name:    "third sample#",
			// 	request: "/value/",
			// 	method:  "GET",
			// 	want: want{
			// 		contentType: "application/json",
			// 		statusCode:  404,
			// 	},
		},

		// TODO: Add test cases.
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			r := chi.NewRouter()
			r.Route("/", handlers.Router)
			// Post("http://127.0.0.1:8080/update/")
			ts := httptest.NewServer(r)
			defer ts.Close()
			body = tt.body

			resp, _ := testRequest(t, ts, tt.method, tt.url)

			defer resp.Body.Close()
			// assert.Equal(t, tt.want.statusCode, resp.StatusCode)

		})
	}
}

func testRequest(t *testing.T, ts *httptest.Server, method, path string) (*http.Response, string) {

	// out, err := json.Marshal(body)
	// if err != nil {
	// 	logrus.Error("Error marshaling metric: ", err)
	// }
	// fmt.Println(out)
	buff := bytes.NewBuffer(body)

	req, err := http.NewRequest(method, ts.URL+path, buff)
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	respBody, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)

	defer resp.Body.Close()
	return resp, string(respBody)
}
