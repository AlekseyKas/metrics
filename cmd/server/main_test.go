package main

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fatih/structs"
	"github.com/go-chi/chi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var mm map[string]interface{} = make(map[string]interface{})
var metr Metrics = Metrics{}

func TestRouter(t *testing.T) {
	// var nameMetInt counter
	//init typs
	type gauge float64
	type counter int64

	//Struct for metrics
	type Metrics struct {
		Alloc         gauge
		BuckHashSys   gauge
		Frees         gauge
		GCCPUFraction gauge
		GCSys         gauge
		HeapAlloc     gauge
		HeapIdle      gauge
		HeapInuse     gauge
		HeapObjects   gauge
		HeapReleased  gauge
		HeapSys       gauge
		LastGC        gauge
		Lookups       gauge
		MCacheInuse   gauge
		MCacheSys     gauge
		MSpanInuse    gauge
		MSpanSys      gauge
		Mallocs       gauge
		NextGC        gauge
		NumForcedGC   gauge
		NumGC         gauge
		OtherSys      gauge
		PauseTotalNs  gauge
		StackInuse    gauge
		StackSys      gauge
		Sys           gauge
		TotalAlloc    gauge

		PollCount   counter
		RandomValue gauge
	}

	mm = structs.Map(metr)

	// jsonString, _ := json.Marshal(mm)

	type want struct {
		contentType string
		statusCode  int
	}
	tests := []struct {
		name    string
		request string
		method  string
		want    want
	}{
		{
			name:    "fist sample#",
			request: "/update/gauge/HeapAlloc/0.112",
			method:  "POST",
			want: want{
				contentType: "text/plain",
				statusCode:  200,
			},
		},
		{
			name:    "second sample#",
			request: "/update/gauge/TestCount/200",
			method:  "POST",
			want: want{
				contentType: "text/plain",
				statusCode:  200,
			},
		},
		{
			name:    "third sample#",
			request: "/update/unknown/TestCount/200",
			method:  "POST",
			want: want{
				contentType: "text/plain",
				statusCode:  501,
			},
		},
		{
			name:    "third sample#",
			request: "/value/gauge/Alloc",
			method:  "GET",
			want: want{
				contentType: "text/plain",
				statusCode:  200,
			},
		},
		{
			name:    "third sample#",
			request: "/",
			method:  "GET",
			want: want{
				contentType: "text/plain",
				statusCode:  200,
			},
		},
		{
			name:    "third sample#",
			request: "/value/gauge/PollCount",
			method:  "GET",
			want: want{
				contentType: "text/plain",
				statusCode:  404,
			},
		},
		{
			name:    "third sample#",
			request: "/value/gauge/PollCount/none",
			method:  "GET",
			want: want{
				contentType: "text/plain",
				statusCode:  404,
			},
		},

		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := chi.NewRouter()
			r.Route("/", metrics.Router)
			ts := httptest.NewServer(r)
			defer ts.Close()

			resp, _ := testRequest(t, ts, tt.method, tt.request)
			defer resp.Body.Close()
			assert.Equal(t, tt.want.statusCode, resp.StatusCode)
			// assert.Equal(t, tt.want.contentType, resp.Header.Get("Content-Type"))
			// assert.Equal(t, string(jsonString), body)

		})
	}
}

func testRequest(t *testing.T, ts *httptest.Server, method, path string) (*http.Response, string) {
	req, err := http.NewRequest(method, ts.URL+path, nil)
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	respBody, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)

	defer resp.Body.Close()
	return resp, string(respBody)
}
