package main

import (
	"context"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var M Metrics

func TestUpdateMetrics(t *testing.T) {

	tests := []struct {
		name    string
		metrics *Metrics
		want    *Metrics
	}{
		{
			name: "first #",
			metrics: &Metrics{
				Alloc:         gauge(1.2),
				BuckHashSys:   gauge(2.1),
				Frees:         gauge(2.1),
				GCCPUFraction: gauge(2.1),
				GCSys:         gauge(2.1),
				HeapAlloc:     gauge(0.00001),
				HeapIdle:      gauge(3),
				HeapInuse:     gauge(4),
				HeapObjects:   gauge(5),
				HeapReleased:  gauge(6),
				HeapSys:       gauge(7),
				LastGC:        gauge(8),
				Lookups:       gauge(9),
				MCacheInuse:   gauge(10),
				MCacheSys:     gauge(11),
				MSpanInuse:    gauge(12),
				MSpanSys:      gauge(13),
				Mallocs:       gauge(14),
				NextGC:        gauge(15),
				NumForcedGC:   gauge(16),
				NumGC:         gauge(17),
				OtherSys:      gauge(0),
				PauseTotalNs:  gauge(0),
				StackInuse:    gauge(0),
				StackSys:      gauge(8),
				Sys:           gauge(8),
				TotalAlloc:    gauge(5),
				RandomValue:   gauge(rand.Float64()),
				PollCount:     counter(0),
			},

			want: &Metrics{
				Alloc:         gauge(1.2),
				BuckHashSys:   gauge(2.1),
				Frees:         gauge(2.1),
				GCCPUFraction: gauge(2.1),
				GCSys:         gauge(2.1),
				HeapAlloc:     gauge(0.00001),
				HeapIdle:      gauge(3),
				HeapInuse:     gauge(4),
				HeapObjects:   gauge(5),
				HeapReleased:  gauge(6),
				HeapSys:       gauge(7),
				LastGC:        gauge(8),
				Lookups:       gauge(9),
				MCacheInuse:   gauge(10),
				MCacheSys:     gauge(11),
				MSpanInuse:    gauge(12),
				MSpanSys:      gauge(13),
				Mallocs:       gauge(14),
				NextGC:        gauge(15),
				NumForcedGC:   gauge(16),
				NumGC:         gauge(17),
				OtherSys:      gauge(0),
				PauseTotalNs:  gauge(0),
				StackInuse:    gauge(0),
				StackSys:      gauge(8),
				Sys:           gauge(8),
				TotalAlloc:    gauge(5),
				RandomValue:   gauge(rand.Float64()),
				PollCount:     counter(3),
			},
		},
	}
	for _, tt := range tests {
		pollint := 2 * time.Second
		ctx, cancel := context.WithCancel(context.Background())
		time.AfterFunc(4*time.Second, cancel)
		t.Run(tt.name, func(t *testing.T) {
			UpdateMetrics(ctx, tt.metrics, pollint)
			M = *tt.metrics
			assert.Equal(t, tt.want.PollCount, M.PollCount)
		})
	}
}

func Test_saveMetrics(t *testing.T) {

	tests := []struct {
		name       string
		metrics    Metrics
		statusCode int
	}{
		{
			name:       "first test",
			metrics:    M,
			statusCode: 400,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ctx, cancel := context.WithCancel(context.Background())
			time.AfterFunc(6*time.Second, cancel)
			err := saveMetrics(ctx, tt.metrics)
			require.NoError(t, err)
			assert.Equal(t, tt.statusCode, 400)

		})
	}
}

// func Test_saveMetrics(t *testing.T) {

// 	tests := []struct {
// 		name    string
// 		metrics Metrics
// 	}{
// 		{
// 			name:    "first test",
// 			metrics: M,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			err := sendMetrics(tt.metrics)
// 			require.Error(t, err)

// 		})
// 	}
// }
