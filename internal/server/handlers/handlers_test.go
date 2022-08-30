package handlers

import (
	"bytes"
	"compress/gzip"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AlekseyKas/metrics/internal/config"
	"github.com/AlekseyKas/metrics/internal/storage"
	"github.com/fatih/structs"
	"github.com/go-chi/chi"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
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
		key    string
		body   []byte
		want   want
	}{
		{
			name:   "fist sample#",
			url:    "/value/",
			method: "POST",
			body:   []byte(`{"ID": "Alloc", "type": "gauge"}`),
			key:    "key",
			want: want{
				contentType: "application/json",
				statusCode:  200,
			},
		},
		{
			name:   "fist sample#",
			url:    "/value/",
			method: "POST",
			key:    "",
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
			key:    "3333",
			body:   []byte(`{"ID": "PollCount", "type": "gouge"}`),
			want: want{
				contentType: "application/json",
				statusCode:  500,
			},
		},
		{
			name:   "4 sample#",
			url:    "/value/",
			method: "POST",
			body:   []byte(`{"ID": "MetricName", "type": "test"}`),
			want: want{
				contentType: "application/json",
				statusCode:  500,
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
				statusCode:  400,
			},
		},
		// {
		// 	name:   "7 sample#",
		// 	url:    "/update/",
		// 	method: "POST",
		// 	key:    "ssd",
		// 	body:   []byte(`{"ID": "Alloc", "type": "gauge", "value": 3.1}`),
		// 	want: want{
		// 		contentType: "application/json",
		// 		statusCode:  200,
		// 	},
		// },

		{
			name:   "saveMetricsSlice success 1#",
			url:    "/updates/",
			method: "POST",
			key:    "ssd",
			body:   []byte(`[{"ID": "Alloc", "type": "gauge", "value": 3.1}]`),
			want: want{
				contentType: "application/json",
				statusCode:  200,
			},
		},

		{
			name:   "saveMetricsSlice success 2#",
			url:    "/updates/",
			method: "POST",
			body:   []byte(`[{"ID": "PollCount", "type": "counter", "delta": 102}]`),
			want: want{
				contentType: "application/json",
				statusCode:  200,
			},
		},
		{
			name:   "saveMetricsSlice success 2#",
			url:    "/updates/",
			method: "POST",
			key:    "lll",
			body:   []byte(`[{"ID": "PollCount", "type": "counter", "delta": 102}]`),
			want: want{
				contentType: "application/json",
				statusCode:  200,
			},
		},
		{
			name:   "saveMetricsSlice success 2#",
			url:    "/updates/",
			method: "POST",
			body:   []byte(`[{"ID": "TestCount", "type": "counter", "delta": 12202}]`),
			want: want{
				contentType: "application/json",
				statusCode:  200,
			},
		},
		{
			name:   "saveMetricsSlice success 2#",
			url:    "/updates/",
			method: "POST",
			key:    "pppp",
			body:   []byte(`[{"ID": "TestCount", "type": "counter", "delta": 12202}]`),
			want: want{
				contentType: "application/json",
				statusCode:  200,
			},
		},
		// {
		// 	name:   "saveMetricsSlice bad request 1#",
		// 	url:    "/updates/",
		// 	method: "POST",
		// 	body:   []byte(`[{"ID": "Alloc", "type": "counter", "delta": 3}]`),
		// 	want: want{
		// 		contentType: "application/json",
		// 		statusCode:  400,
		// 	},
		// },

		// {
		// 	name:   "saveMetricsSlice bad request 2#",
		// 	url:    "/updates/",
		// 	method: "POST",
		// 	body:   []byte(`[{"ID": "Alloc", "type": "counter", "delta": 3.1}]`),
		// 	want: want{
		// 		contentType: "application/json",
		// 		statusCode:  400,
		// 	},
		// },
		// {
		// 	name:   "saveMetricsSlice bad request 3#",
		// 	url:    "/updates/",
		// 	method: "POST",
		// 	body:   []byte(`[{"ID": "Ametrics", "type": "gauge", "value": 3.2221}]`),
		// 	want: want{
		// 		contentType: "application/json",
		// 		statusCode:  200,
		// 	},
		// },

		// {
		// 	name:   "saveMetricsSlice bad request 4#",
		// 	url:    "/updates/",
		// 	method: "POST",
		// 	body:   []byte(`[{"ID": "Ametrics", "type": "gauge", "delta": 3333.1}]`),
		// 	want: want{
		// 		contentType: "plain/text",
		// 		statusCode:  200,
		// 	},
		// },

		// {
		// 	name:   "saveMetricsSlice bad request 5#",
		// 	url:    "/updates/",
		// 	method: "POST",
		// 	body:   []byte(`[{"ID": "PollCount", "type": "gauge", "delta": 3333.1}]`),
		// 	want: want{
		// 		contentType: "plain/text",
		// 		statusCode:  200,
		// 	},
		// },
	}
	logger, _ := zap.NewProduction()

	InitLogger(logger)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := chi.NewRouter()
			r.Route("/", Router)

			ts := httptest.NewServer(r)
			defer ts.Close()
			body = tt.body
			config.ArgsM.Key = tt.key
			buff := bytes.NewBuffer(body)

			req, err := http.NewRequest(tt.method, ts.URL+tt.url, buff)
			require.NoError(t, err)

			resp, errr := http.DefaultClient.Do(req)
			require.Equal(t, tt.want.statusCode, resp.StatusCode)
			require.NoError(t, errr)

			defer resp.Body.Close()

		})
	}
}

func Test_compareHash(t *testing.T) {
	f := float64(99.1)

	jm := storage.JSONMetrics{
		ID:    "Alloc",
		MType: "gauge",
		Value: &f,
		Hash:  "6521182e1b27f1efe5d43f7e1a438eeaaff4e89bf656d28a801d0d16e6b28557",
	}
	jm2 := storage.JSONMetrics{
		ID:    "Alloc",
		MType: "gauge",
		Value: &f,
		Hash:  "7c1ce04447600a7ede550e33a9133102e8706755d86205774fa5e8ca2fe5e352",
	}
	key := []byte("key")

	type args struct {
		s   *storage.JSONMetrics
		key []byte
	}
	tests := []struct {
		name     string
		args     args
		wantBool bool
	}{
		{
			name: "first",
			args: args{
				s:   &jm,
				key: key,
			},
			wantBool: false,
		},
		{
			name: "second",
			args: args{
				s:   &jm2,
				key: key,
			},
			wantBool: true,
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotB, err := compareHash(tt.args.s, tt.args.key)
			require.NoError(t, err)
			if gotB != tt.wantBool {
				t.Errorf("compareHash() = %v, want %v", gotB, tt.wantBool)
			}
		})
	}
}

func Test_calculateHash(t *testing.T) {
	f := float64(99.1)
	c := int64(99)

	type args struct {
		s   *storage.JSONMetrics
		key []byte
	}

	tests := []struct {
		name     string
		wantHash string
		args     args
	}{
		{
			name: "Success calculate",
			args: args{
				s: &storage.JSONMetrics{
					ID:    "Alloc",
					MType: "gauge",
					Value: &f,
				},
				key: []byte("key"),
			},
			wantHash: "7c1ce04447600a7ede550e33a9133102e8706755d86205774fa5e8ca2fe5e352",
		},
		{
			name: "Success calculate",
			args: args{
				s: &storage.JSONMetrics{
					ID:    "PollCount",
					MType: "counter",
					Delta: &c,
				},
				key: []byte("key"),
			},
			wantHash: "10cf641702fb80988f18a68a913dd980d0b10a9e24332be9edd5f4da92b12a22",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calculateHash(tt.args.s, tt.args.key)
			require.Equal(t, tt.wantHash, tt.args.s.Hash)
		})
	}
}

func Test_getMetrics(t *testing.T) {

	type want struct {
		contentType string
		statusCode  int
	}

	tests := []struct {
		name   string
		url    string
		method string
		want   want
	}{
		{
			name:   "fist sample#",
			url:    "/",
			method: "GET",
			want: want{
				contentType: "text/html",
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

			req, err := http.NewRequest(tt.method, ts.URL+tt.url, nil)
			require.NoError(t, err)

			resp, errr := http.DefaultClient.Do(req)
			require.Equal(t, tt.want.statusCode, resp.StatusCode)
			require.NoError(t, errr)

			defer resp.Body.Close()

		})
	}
}

func Test_getMetric(t *testing.T) {

	type want struct {
		contentType string
		statusCode  int
	}

	tests := []struct {
		name   string
		url    string
		method string
		want   want
	}{
		{
			name:   "fist sample#",
			url:    "/value/gauge/Alloc",
			method: "GET",
			want: want{
				contentType: "text/html",
				statusCode:  200,
			},
		},
		{
			name:   "second sample#",
			url:    "/value/unknown/Alloc",
			method: "GET",
			want: want{
				contentType: "text/html",
				statusCode:  500,
			},
		},
		{
			name:   "third sample#",
			url:    "/value/gauge/PollCount",
			method: "GET",
			want: want{
				contentType: "text/html",
				statusCode:  404,
			},
		},

		{
			name:   "4th sample#",
			url:    "/value/gauge/unknown",
			method: "GET",
			want: want{
				contentType: "text/html",
				statusCode:  404,
			},
		},
		{
			name:   "5th sample#",
			url:    "/value/counter/PollCount",
			method: "GET",
			want: want{
				contentType: "text/html",
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

			req, err := http.NewRequest(tt.method, ts.URL+tt.url, nil)
			require.NoError(t, err)

			resp, errr := http.DefaultClient.Do(req)
			require.Equal(t, tt.want.statusCode, resp.StatusCode)
			require.NoError(t, errr)

			defer resp.Body.Close()

		})
	}
}

func Test_saveMetrics(t *testing.T) {
	type want struct {
		contentType string
		statusCode  int
	}

	tests := []struct {
		name   string
		url    string
		method string
		want   want
	}{
		{
			name:   "fist sample#",
			url:    "/update/gauge/Alloc/2.1",
			method: "POST",
			want: want{
				contentType: "text/plain",
				statusCode:  200,
			},
		},

		{
			name:   "second sample#",
			url:    "/update/counter/PollCounter/2",
			method: "POST",
			want: want{
				contentType: "text/plain",
				statusCode:  200,
			},
		},

		{
			name:   "third sample#",
			url:    "/update/unknown/PollCounter/2",
			method: "POST",
			want: want{
				contentType: "text/plain",
				statusCode:  501,
			},
		},

		{
			name:   "4th sample#",
			url:    "/update/gauge/unknown/aaa",
			method: "POST",
			want: want{
				contentType: "text/plain",
				statusCode:  400,
			},
		},

		{
			name:   "5th sample#",
			url:    "/update/counter/unknown/aaa",
			method: "POST",
			want: want{
				contentType: "text/plain",
				statusCode:  400,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := chi.NewRouter()
			r.Route("/", Router)

			ts := httptest.NewServer(r)
			defer ts.Close()

			req, err := http.NewRequest(tt.method, ts.URL+tt.url, nil)
			require.NoError(t, err)

			resp, errr := http.DefaultClient.Do(req)
			require.Equal(t, tt.want.statusCode, resp.StatusCode)
			require.NoError(t, errr)

			defer resp.Body.Close()

		})
	}
}

func Test_checkConnection(t *testing.T) {

	type want struct {
		contentType string
		statusCode  int
	}

	tests := []struct {
		name   string
		url    string
		method string
		want   want
	}{
		{
			name:   "fist sample#",
			url:    "/ping",
			method: "GET",
			want: want{
				contentType: "text/html",
				statusCode:  500,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := chi.NewRouter()
			r.Route("/", Router)

			ts := httptest.NewServer(r)
			defer ts.Close()

			req, err := http.NewRequest(tt.method, ts.URL+tt.url, nil)
			require.NoError(t, err)

			resp, errr := http.DefaultClient.Do(req)
			require.Equal(t, tt.want.statusCode, resp.StatusCode)
			require.NoError(t, errr)

			defer resp.Body.Close()
		})
	}
}

func TestRouterCompressDecompress(t *testing.T) {

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
				statusCode:  500,
			},
		},
		{
			name:   "4 sample#",
			url:    "/value/",
			method: "POST",
			body:   []byte(`{"ID": "MetricName", "type": "test"}`),
			want: want{
				contentType: "application/json",
				statusCode:  500,
			},
		},
		// {
		// 	name:   "5 sample#",
		// 	url:    "/update/",
		// 	method: "POST",
		// 	body:   []byte(`{"ID": "Pollcount", "type": "counter", "delta": 45}`),
		// 	want: want{
		// 		contentType: "application/json",
		// 		statusCode:  200,
		// 	},
		// },
		// {
		// 	name:   "6 sample#",
		// 	url:    "/update/",
		// 	method: "POST",
		// 	body:   []byte(`{"ID": "Pollcount", "type": "gauge", "delta": "12"}`),
		// 	want: want{
		// 		contentType: "application/json",
		// 		statusCode:  400,
		// 	},
		// },
		// {
		// 	name:   "7 sample#",
		// 	url:    "/update/",
		// 	method: "POST",
		// 	body:   []byte(`{"ID": "Alloc", "type": "gauge", "value": 3.1}`),
		// 	want: want{
		// 		contentType: "application/json",
		// 		statusCode:  200,
		// 	},
		// },

		// {
		// 	name:   "saveMetricsSlice success#",
		// 	url:    "/updates/",
		// 	method: "POST",
		// 	body:   []byte(`[{"ID": "Alloc", "type": "gauge", "value": 3.1}]`),
		// 	want: want{
		// 		contentType: "application/json",
		// 		statusCode:  200,
		// 	},
		// },

		// {
		// 	name:   "saveMetricsSlice#",
		// 	url:    "/updates/",
		// 	method: "POST",
		// 	body:   []byte(`[{"ID": "Alloc", "type": "counter", "delta": 3}]`),
		// 	want: want{
		// 		contentType: "application/json",
		// 		statusCode:  400,
		// 	},
		// },
		// {
		// 	name:   "saveMetricsSlice delta float64#",
		// 	url:    "/updates/",
		// 	method: "POST",
		// 	body:   []byte(`[{"ID": "Alloc", "type": "counter", "delta": 3.1}]`),
		// 	want: want{
		// 		contentType: "application/json",
		// 		statusCode:  400,
		// 	},
		// },
	}
	logger, _ := zap.NewProduction()

	InitLogger(logger)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := chi.NewRouter()
			r.Route("/", Router)

			ts := httptest.NewServer(r)
			defer ts.Close()
			body = tt.body

			var buf bytes.Buffer

			gz, _ := gzip.NewWriterLevel(&buf, gzip.BestCompression)

			gz.Write(tt.body)
			gz.Close()

			req, err := http.NewRequest(tt.method, ts.URL+tt.url, &buf)
			req.Header.Set("Content-Encoding", "gzip")
			req.Header.Set("Content-Type", "application/json")
			require.NoError(t, err)
			resp, errr := http.DefaultClient.Do(req)
			require.NoError(t, errr)
			require.Equal(t, tt.want.statusCode, resp.StatusCode)

			defer resp.Body.Close()

		})
	}
}
