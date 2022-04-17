package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AlekseyKas/metrics/internal/storage"
	"github.com/fatih/structs"
	"github.com/go-chi/chi"
	"github.com/stretchr/testify/require"
)

var body []byte

func TestRouter(t *testing.T) {

	s := &storage.MetricsStore{
		MM: structs.Map(storage.Metrics{}),
	}
	SetStorage(s)
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
		},
		{
			name:   "fist sample#",
			url:    "/value/",
			method: "POST",
			body:   []byte(`{"ID": "PollCount", "type": "counter"}`),
			want: want{
				contentType: "application/json",
				statusCode:  200,
			},
		},
		{
			name:   "second sample#",
			url:    "/value/",
			method: "POST",
			body:   []byte(`{"ID": "PollCount", "type": "gouge"}`),
			want: want{
				contentType: "application/json",
				statusCode:  200,
			},
		},
		{
			name:   "4 sample#",
			url:    "/value/",
			method: "POST",
			body:   []byte(`{"ID": "MetricName", "type": "test"}`),
			want: want{
				contentType: "application/json",
				statusCode:  200,
			},
		},
		{
			name:   "5 sample#",
			url:    "/update/",
			method: "POST",
			body:   []byte(`{"ID": "Pollcount", "type": "counter", "delta": 45}`),
			want: want{
				contentType: "application/json",
				statusCode:  200,
			},
		},
		{
			name:   "6 sample#",
			url:    "/update/",
			method: "POST",
			body:   []byte(`{"ID": "Pollcount", "type": "gauge", "delta": "12"}`),
			want: want{
				contentType: "application/json",
				statusCode:  200,
			},
		},
		{
			name:   "7 sample#",
			url:    "/update/",
			method: "POST",
			body:   []byte(`{"ID": "Alloc", "type": "gauge", "value": 3.1}`),
			want: want{
				contentType: "application/json",
				statusCode:  200,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := chi.NewRouter()
			r.Route("/", Router)

			ts := httptest.NewServer(r)
			defer ts.Close()
			body = tt.body

			buff := bytes.NewBuffer(body)
			req, err := http.NewRequest(tt.method, ts.URL+tt.url, buff)
			require.NoError(t, err)

			resp, errr := http.DefaultClient.Do(req)
			require.Equal(t, tt.want.statusCode, resp.StatusCode)
			require.NoError(t, errr)

			// respb, err := ioutil.ReadAll(resp.Body)
			// require.NoError(t, err)
			// assert.Equal(t, string(respb), string(tt.body))
			defer resp.Body.Close()

		})
	}
}

// func testRequest(t *testing.T, ts *httptest.Server, method, path string) (*http.Response, string) {

// 	// assert.NoError(t, err)

// }
