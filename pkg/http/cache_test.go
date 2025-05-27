package http

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	gomock "github.com/golang/mock/gomock"
	"github.com/neurocode-io/cache-offloader/config"
	"github.com/neurocode-io/cache-offloader/pkg/model"
	"github.com/stretchr/testify/assert"
)

type responseMatcher struct {
	body   []byte
	status int
}

func newResponseMatcher(status int, body []byte) *responseMatcher {
	return &responseMatcher{body, status}
}

func (m *responseMatcher) String() string {
	return fmt.Sprintf("Response Matcher: %d %s", m.status, string(m.body))
}

func (m *responseMatcher) Matches(x interface{}) bool {
	resp, ok := x.(*model.Response)
	if !ok {
		return false
	}
	if string(resp.Body) != string(m.body) {
		return false
	}
	if resp.Status != m.status {
		return false
	}

	return true
}

func mustRequest(t *testing.T, url, method string) *http.Request {
	if method == "" {
		method = "GET"
	}
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		t.Fatal(err)
	}

	return req
}

func mustURL(t *testing.T, downstreamURL string) *url.URL {
	u, err := url.Parse(downstreamURL)
	if err != nil {
		t.Fatal(err)
	}

	return u
}

func TestCacheHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	proxied := http.StatusUseProxy
	endpoint := "/status/200?q=1"
	downstreamServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(proxied)
	}))
	defer downstreamServer.Close()

	downstreamServerNok := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer downstreamServerNok.Close()

	webSocketReq := mustRequest(t, "/connect", "")
	webSocketReq.Header.Set("connection", "upgrade")

	tests := []struct {
		name    string
		handler handler
		req     *http.Request
		want    int
	}{
		{
			name: "cacheLookup error should still forward request to downstream and store response",
			handler: newCacheHandler(
				func() Cacher {
					mock := NewMockCacher(ctrl)
					mock.EXPECT().LookUp(gomock.Any(), gomock.Any()).Return(nil, errors.New("test-error"))
					mock.EXPECT().Store(gomock.Any(), gomock.Any(), gomock.Any())
					return mock
				}(),
				func() MetricsCollector {
					mock := NewMockMetricsCollector(ctrl)
					mock.EXPECT().CacheMiss("GET", proxied)
					return mock
				}(),
				nil,
				config.CacheConfig{
					DownstreamHost: mustURL(t, downstreamServer.URL),
				},
			),
			req:  mustRequest(t, endpoint, ""),
			want: proxied,
		},
		{
			name: "cache miss",
			handler: newCacheHandler(
				func() Cacher {
					mock := NewMockCacher(ctrl)
					mock.EXPECT().LookUp(gomock.Any(), gomock.Any()).Return(nil, nil).Times(1)
					mock.EXPECT().Store(gomock.Any(), gomock.Any(), newResponseMatcher(proxied, nil)).Return(nil).Times(1)
					return mock
				}(),
				func() MetricsCollector {
					mock := NewMockMetricsCollector(ctrl)
					mock.EXPECT().CacheMiss("GET", proxied).Times(1)
					return mock
				}(),
				nil,
				config.CacheConfig{
					DownstreamHost: mustURL(t, downstreamServer.URL),
				},
			),
			req:  mustRequest(t, endpoint, ""),
			want: proxied,
		},
		{
			name: "websockets will not be stored in cache",
			handler: newCacheHandler(
				func() Cacher {
					mock := NewMockCacher(ctrl)
					mock.EXPECT().LookUp(nil, nil).Times(0)
					mock.EXPECT().Store(nil, nil, nil).Times(0)
					return mock
				}(),
				func() MetricsCollector {
					mock := NewMockMetricsCollector(ctrl)
					mock.EXPECT().CacheMiss(nil, nil).Times(0)
					return mock
				}(),
				nil,
				config.CacheConfig{
					DownstreamHost: mustURL(t, downstreamServer.URL),
				},
			),
			req:  webSocketReq,
			want: proxied,
		},
		{
			name: "POST requests will not be stored in cache",
			handler: newCacheHandler(
				func() Cacher {
					mock := NewMockCacher(ctrl)
					mock.EXPECT().LookUp(nil, nil).Times(0)
					mock.EXPECT().Store(nil, nil, nil).Times(0)
					return mock
				}(),
				func() MetricsCollector {
					mock := NewMockMetricsCollector(ctrl)
					mock.EXPECT().CacheMiss(nil, nil).Times(0)
					return mock
				}(),
				nil,
				config.CacheConfig{
					DownstreamHost: mustURL(t, downstreamServer.URL),
				},
			),
			req:  mustRequest(t, endpoint, "POST"),
			want: proxied,
		},
		{
			name: "PUT requests will not be stored in cache",
			handler: newCacheHandler(
				func() Cacher {
					mock := NewMockCacher(ctrl)
					mock.EXPECT().LookUp(nil, nil).Times(0)
					mock.EXPECT().Store(nil, nil, nil).Times(0)
					return mock
				}(),
				func() MetricsCollector {
					mock := NewMockMetricsCollector(ctrl)
					mock.EXPECT().CacheMiss(nil, nil).Times(0)
					return mock
				}(),
				nil,
				config.CacheConfig{
					DownstreamHost: mustURL(t, downstreamServer.URL),
				},
			),
			req:  mustRequest(t, endpoint, "PUT"),
			want: proxied,
		},
		{
			name: "PATCH requests will not be stored in cache",
			handler: newCacheHandler(
				func() Cacher {
					mock := NewMockCacher(ctrl)
					mock.EXPECT().LookUp(nil, nil).Times(0)
					mock.EXPECT().Store(nil, nil, nil).Times(0)
					return mock
				}(),
				func() MetricsCollector {
					mock := NewMockMetricsCollector(ctrl)
					mock.EXPECT().CacheMiss(nil, nil).Times(0)
					return mock
				}(),
				nil,
				config.CacheConfig{
					DownstreamHost: mustURL(t, downstreamServer.URL),
				},
			),
			req:  mustRequest(t, endpoint, "PATCH"),
			want: proxied,
		},
		{
			name: "DELETE requests will not be stored in cache",
			handler: newCacheHandler(
				func() Cacher {
					mock := NewMockCacher(ctrl)
					mock.EXPECT().LookUp(nil, nil).Times(0)
					mock.EXPECT().Store(nil, nil, nil).Times(0)
					return mock
				}(),
				func() MetricsCollector {
					mock := NewMockMetricsCollector(ctrl)
					mock.EXPECT().CacheMiss(nil, nil).Times(0)
					return mock
				}(),
				nil,
				config.CacheConfig{
					DownstreamHost: mustURL(t, downstreamServer.URL),
				},
			),
			req:  mustRequest(t, endpoint, "DELETE"),
			want: proxied,
		},
		{
			name: "statusCode > 500 from downstream will not be stored in cache",
			handler: newCacheHandler(
				func() Cacher {
					mock := NewMockCacher(ctrl)
					mock.EXPECT().LookUp(gomock.Any(), gomock.Any()).Return(nil, nil).Times(1)
					mock.EXPECT().Store(nil, nil, nil).Times(0)
					return mock
				}(),
				func() MetricsCollector {
					mock := NewMockMetricsCollector(ctrl)
					mock.EXPECT().CacheMiss("GET", http.StatusInternalServerError).Times(1)
					return mock
				}(),
				nil,
				config.CacheConfig{
					DownstreamHost: mustURL(t, downstreamServerNok.URL),
				},
			),
			req:  mustRequest(t, "/status/500?q=1", ""),
			want: http.StatusInternalServerError,
		},
		{
			name: "cache hit",
			handler: newCacheHandler(
				func() Cacher {
					mock := NewMockCacher(ctrl)
					mock.EXPECT().LookUp(gomock.Any(), gomock.Any()).Return(&model.Response{
						Status: http.StatusOK,
						Body:   []byte("hello"),
					}, nil)
					return mock
				}(),
				func() MetricsCollector {
					mock := NewMockMetricsCollector(ctrl)
					mock.EXPECT().CacheHit("GET", http.StatusOK)
					return mock
				}(),
				func() Worker {
					mock := NewMockWorker(ctrl)
					mock.EXPECT().Start(gomock.Any(), gomock.Any())
					return mock
				}(),
				config.CacheConfig{
					DownstreamHost: mustURL(t, downstreamServer.URL),
				},
			),
			req:  mustRequest(t, endpoint, ""),
			want: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			want := httptest.NewRecorder()
			tt.handler.ServeHTTP(want, tt.req)
			assert.Equal(t, tt.want, want.Code)
		})
	}
}

func TestGetCacheKey(t *testing.T) {
	tests := []struct {
		name            string
		path            string
		method          string
		query           string
		headers         map[string]string
		shouldHashQuery bool
		ignoreParams    []string
		hashHeaders     []string
		want            string
	}{
		{
			name:            "simple path without query",
			path:            "/api/users",
			method:          "GET",
			query:           "",
			shouldHashQuery: true,
			want:            "GET:/api/users",
		},
		{
			name:            "path with query parameters",
			path:            "/api/users",
			method:          "GET",
			query:           "name=john&age=30",
			shouldHashQuery: true,
			want:            "GET:/api/users&age=30&name=john",
		},
		{
			name:            "query parameters in different order",
			path:            "/api/users",
			method:          "GET",
			query:           "age=30&name=john",
			shouldHashQuery: true,
			want:            "GET:/api/users&age=30&name=john",
		},
		{
			name:            "multiple values for same parameter",
			path:            "/api/users",
			method:          "GET",
			query:           "role=admin&role=user",
			shouldHashQuery: true,
			want:            "GET:/api/users&role=admin&role=user",
		},
		{
			name:            "different HTTP method",
			path:            "/api/users",
			method:          "POST",
			query:           "name=john",
			shouldHashQuery: true,
			want:            "POST:/api/users&name=john",
		},
		{
			name:            "query parameters disabled",
			path:            "/api/users",
			method:          "GET",
			query:           "name=john&age=30",
			shouldHashQuery: false,
			want:            "GET:/api/users",
		},
		{
			name:            "ignored query parameters",
			path:            "/api/users",
			method:          "GET",
			query:           "name=john&age=30&timestamp=123",
			shouldHashQuery: true,
			ignoreParams:    []string{"timestamp"},
			want:            "GET:/api/users&age=30&name=john",
		},
		{
			name:            "special characters in path and query",
			path:            "/api/users/123/profile",
			method:          "GET",
			query:           "filter=active&sort=name",
			shouldHashQuery: true,
			want:            "GET:/api/users/123/profile&filter=active&sort=name",
		},
		{
			name:   "with authorization header",
			path:   "/api/users",
			method: "GET",
			query:  "name=john",
			headers: map[string]string{
				"Authorization": "Bearer token123",
			},
			shouldHashQuery: true,
			hashHeaders:     []string{"Authorization"},
			want:            "GET:/api/users&name=john|Authorization=Bearer token123",
		},
		{
			name:   "multiple headers",
			path:   "/api/users",
			method: "GET",
			headers: map[string]string{
				"Authorization": "Bearer token123",
				"X-User-ID":     "user456",
				"Accept":        "application/json",
			},
			shouldHashQuery: true,
			hashHeaders:     []string{"Authorization", "X-User-ID", "Accept"},
			want:            "GET:/api/users|Accept=application/json|Authorization=Bearer token123|X-User-ID=user456",
		},
		{
			name:   "multiple values for same header",
			path:   "/api/users",
			method: "GET",
			headers: map[string]string{
				"Accept": "application/json,text/plain",
			},
			shouldHashQuery: true,
			hashHeaders:     []string{"Accept"},
			want:            "GET:/api/users|Accept=application/json,text/plain",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			req, err := http.NewRequest(tt.method, tt.path, nil)
			if err != nil {
				t.Fatal(err)
			}

			// Add query parameters if any
			if tt.query != "" {
				req.URL.RawQuery = tt.query
			}

			// Add headers
			for key, value := range tt.headers {
				req.Header.Set(key, value)
			}

			// Create handler with config
			h := handler{
				cfg: config.CacheConfig{
					ShouldHashQuery: tt.shouldHashQuery,
					HashQueryIgnore: make(map[string]bool),
					HashHeaders:     tt.hashHeaders,
				},
			}

			// Add ignored parameters
			for _, param := range tt.ignoreParams {
				h.cfg.HashQueryIgnore[param] = true
			}

			// Get cache key
			got := h.getCacheKey(req)

			// Calculate expected hash
			expectedHash := sha256.New()
			expectedHash.Write([]byte(tt.want))
			expected := fmt.Sprintf("%x", expectedHash.Sum(nil))

			assert.Equal(t, expected, got, "cache key mismatch")
		})
	}
}
