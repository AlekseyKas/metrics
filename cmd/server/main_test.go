package main

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AlekseyKas/metrics/cmd/server/handlers"
	"github.com/AlekseyKas/metrics/internal/storage"
	"github.com/fatih/structs"
	"github.com/go-chi/chi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRouter(t *testing.T) {

	// mm = structs.Map(metr)

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
	var MapMetrics map[string]interface{} = structs.Map(storage.Metrics{})
	s := &storage.MetricsStore{
		MM: MapMetrics,
	}
	handlers.SetStorage(s)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			r := chi.NewRouter()
			r.Route("/", handlers.Router)
			ts := httptest.NewServer(r)
			defer ts.Close()

			resp, _ := testRequest(t, ts, tt.method, tt.request)
			defer resp.Body.Close()
			assert.Equal(t, tt.want.statusCode, resp.StatusCode)

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
