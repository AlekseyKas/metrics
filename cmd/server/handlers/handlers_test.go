package handlers

import (
	"testing"

	"github.com/go-chi/chi"
)

func TestRouter(t *testing.T) {
	type args struct {
		r chi.Router
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Router(tt.args.r)
		})
	}
}
