package util

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFetchPageTitle(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/valid":
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Write([]byte(`<html><head><title>Test Page Title</title></head><body></body></html>`))
		case "/no-title":
			w.Header().Set("Content-Type", "text/html")
			w.Write([]byte(`<html><body>No Title Here</body></html>`))
		case "/utf8-meta":
			w.Header().Set("Content-Type", "text/html")
			w.Write([]byte(`<html><head><meta charset="utf-8"><title>UTF-8 Title</title></head></html>`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()

	tests := []struct {
		name    string
		url     string
		want    string
		wantErr bool
	}{
		{
			name:    "Valid Title",
			url:     ts.URL + "/valid",
			want:    "Test Page Title",
			wantErr: false,
		},
		{
			name:    "No Title",
			url:     ts.URL + "/no-title",
			wantErr: true,
		},
		{
			name:    "404",
			url:     ts.URL + "/not-found",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FetchPageTitle(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("FetchPageTitle() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("FetchPageTitle() = %v, want %v", got, tt.want)
			}
		})
	}
}
