package main

import (
	"encoding/json"
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

	r := chi.NewRouter()
	r.Route("/", metrics.Router)
	ts := httptest.NewServer(r)
	defer ts.Close()

	mm = structs.Map(metr)

	jsonString, _ := json.Marshal(mm)

	resp, bodygetMetrics := testRequest(t, ts, "GET", "/")

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, string(jsonString), bodygetMetrics)
	defer resp.Body.Close()
	respVal, bodygetMetric := testRequest(t, ts, "GET", "/value/gauge/HeapSys")
	assert.Equal(t, http.StatusOK, respVal.StatusCode)
	assert.Equal(t, "0", bodygetMetric)
	defer respVal.Body.Close()

	respSave, _ := testRequest(t, ts, "POST", "/update/gauge/HeapSys/0.2201")
	assert.Equal(t, http.StatusOK, respSave.StatusCode)
	assert.Equal(t, "text/plain", respSave.Header.Get("Content-Type"))
	defer respSave.Body.Close()

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
