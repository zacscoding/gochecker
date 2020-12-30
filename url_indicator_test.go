package gochecker

import (
	"context"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestUrlIndicator_Health(t *testing.T) {
	cases := []struct {
		Name       string
		Method     string
		Headers    http.Header
		Path       string
		Timeout    time.Duration
		HandleFunc func(t *testing.T, rw http.ResponseWriter, req *http.Request)
		// expected
		IsUp bool
		Err  string
		Code int
	}{
		{
			Name:   "Status OK",
			Method: "GET",
			Path:   "/health",
			Headers: http.Header{
				"key1": []string{"value1"},
			},
			Timeout: time.Second,
			HandleFunc: func(t *testing.T, rw http.ResponseWriter, req *http.Request) {
				// assert method, path
				assert.Equal(t, "/health", req.URL.Path)
				assert.Equal(t, "GET", req.Method)
				assert.Equal(t, "value1", req.Header.Get("key1"))
				rw.WriteHeader(http.StatusOK)
			},
			IsUp: true,
			Code: http.StatusOK,
		}, {
			Name:    "Status OK But Timeout",
			Method:  "GET",
			Path:    "/health",
			Timeout: time.Millisecond * 100,
			HandleFunc: func(t *testing.T, rw http.ResponseWriter, req *http.Request) {
				// assert method, path
				assert.Equal(t, "/health", req.URL.Path)
				assert.Equal(t, "GET", req.Method)
				time.Sleep(110 * time.Millisecond)
				rw.WriteHeader(http.StatusOK)
			},
			IsUp: false,
			Err:  "context deadline exceeded",
		}, {
			Name:   "Internal Server Error",
			Method: "GET",
			Path:   "/health",
			Headers: http.Header{
				"key1": []string{"value1"},
			},
			Timeout: time.Second,
			HandleFunc: func(t *testing.T, rw http.ResponseWriter, req *http.Request) {
				// assert method, path
				assert.Equal(t, "/health", req.URL.Path)
				assert.Equal(t, "GET", req.Method)
				assert.Equal(t, "value1", req.Header.Get("key1"))
				rw.WriteHeader(http.StatusInternalServerError)
			},
			IsUp: false,
			Code: http.StatusInternalServerError,
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				tc.HandleFunc(t, rw, req)
			}))
			indicator := UrlIndicator{
				Name:    "indicator",
				Method:  tc.Method,
				Headers: tc.Headers,
				Url:     server.URL + tc.Path,
				timeout: tc.Timeout,
			}
			status := indicator.Health(context.Background())

			assert.Equal(t, tc.IsUp, status.IsUp())
			if tc.Code == 0 {
				_, ok := status.details["status"]
				assert.False(t, ok)
			} else {
				assert.Equal(t, tc.Code, status.details["status"].(int))
			}

			if tc.Err == "" {
				_, ok := status.details["err"]
				assert.False(t, ok)
			} else {
				assert.Contains(t, status.details["err"].(string), tc.Err)
			}

		})
	}
}
