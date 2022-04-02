package main

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// var mm map[string]interface{} = make(map[string]interface{})
// var metr Metrics = Metrics{}

// func testRequest(t *testing.T, ts *httptest.Server, method, path string) (*http.Response, string) {
// 	req, err := http.NewRequest(method, ts.URL+path, nil)
// 	require.NoError(t, err)

// 	resp, err := http.DefaultClient.Do(req)
// 	require.NoError(t, err)

// 	respBody, err := ioutil.ReadAll(resp.Body)
// 	require.NoError(t, err)

// 	defer resp.Body.Close()

// 	return resp, string(respBody)
// }

func TestClient(t *testing.T) {
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
	// mm = structs.Map(metr)

	type want struct {
		contentType string
		statusCode  int
	}
	tests := []struct {
		name    string
		request string
		warnErr error
	}{
		{
			name:    "fist sample#",
			request: "/",
			warnErr: nil,
		},
		// TODO: Add test cases.
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ctx, cancel := context.WithCancel(context.Background())
			err := saveMetrics(ctx, M)

			// UpdateMetrics(ctx, &M, pollInterval)
			// fmt.Println(&M.Alloc)
			require.Error(t, err)
			time.AfterFunc(3*time.Second, cancel)
			// assert.Equal(t, tt.warnErr, err)

			// assert.Equal(t, string(jsonString), body)
		})
	}

}
