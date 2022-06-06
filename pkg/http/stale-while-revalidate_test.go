package http

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"neurocode.io/cache-offloader/pkg/model"
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

func mustURL(t *testing.T, downstreamURL string) url.URL {
	u, err := url.Parse(downstreamURL)
	if err != nil {
		t.Fatal(err)
	}

	return *u
}

// nolint:funlen
func TestStaleWhileRevalidate(t *testing.T) {
	ctrl := gomock.NewController(t)

	proxied := http.StatusUseProxy
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
			name: "cache miss",
			handler: handler{
				cacher: func() Cacher {
					mock := NewMockCacher(ctrl)
					mock.EXPECT().LookUp(gomock.Any(), gomock.Any()).Return(nil, nil).Times(1)
					mock.EXPECT().Store(gomock.Any(), gomock.Any(), newResponseMatcher(proxied, nil)).Return(nil).Times(1)

					return mock
				}(),
				metricsCollector: func() MetricsCollector {
					mock := NewMockMetricsCollector(ctrl)
					mock.EXPECT().CacheMiss("GET", proxied).Times(1)

					return mock
				}(),
				downstreamURL: mustURL(t, downstreamServer.URL),
			},

			req:  mustRequest(t, "/status/200?q=1", ""),
			want: proxied,
		},
		{
			name: "websockets will not be stored in cache",
			handler: handler{
				cacher: func() Cacher {
					mock := NewMockCacher(ctrl)
					mock.EXPECT().LookUp(nil, nil).Times(0)
					mock.EXPECT().Store(nil, nil, nil).Times(0)

					return mock
				}(),
				metricsCollector: func() MetricsCollector {
					mock := NewMockMetricsCollector(ctrl)
					mock.EXPECT().CacheMiss(nil, nil).Times(0)

					return mock
				}(),
				downstreamURL: mustURL(t, downstreamServer.URL),
			},

			req:  webSocketReq,
			want: proxied,
		},
		{
			name: "POST requests will not be stored in cache",
			handler: handler{
				cacher: func() Cacher {
					mock := NewMockCacher(ctrl)
					mock.EXPECT().LookUp(nil, nil).Times(0)
					mock.EXPECT().Store(nil, nil, nil).Times(0)

					return mock
				}(),
				metricsCollector: func() MetricsCollector {
					mock := NewMockMetricsCollector(ctrl)
					mock.EXPECT().CacheMiss(nil, nil).Times(0)

					return mock
				}(),
				downstreamURL: mustURL(t, downstreamServer.URL),
			},

			req:  mustRequest(t, "/status/200?q=1", "POST"),
			want: proxied,
		},
		{
			name: "PUT requests will not be stored in cache",
			handler: handler{
				cacher: func() Cacher {
					mock := NewMockCacher(ctrl)
					mock.EXPECT().LookUp(nil, nil).Times(0)
					mock.EXPECT().Store(nil, nil, nil).Times(0)

					return mock
				}(),
				metricsCollector: func() MetricsCollector {
					mock := NewMockMetricsCollector(ctrl)
					mock.EXPECT().CacheMiss(nil, nil).Times(0)

					return mock
				}(),
				downstreamURL: mustURL(t, downstreamServer.URL),
			},

			req:  mustRequest(t, "/status/200?q=1", "PUT"),
			want: proxied,
		},
		{
			name: "PATCH requests will not be stored in cache",
			handler: handler{
				cacher: func() Cacher {
					mock := NewMockCacher(ctrl)
					mock.EXPECT().LookUp(nil, nil).Times(0)
					mock.EXPECT().Store(nil, nil, nil).Times(0)

					return mock
				}(),
				metricsCollector: func() MetricsCollector {
					mock := NewMockMetricsCollector(ctrl)
					mock.EXPECT().CacheMiss(nil, nil).Times(0)

					return mock
				}(),
				downstreamURL: mustURL(t, downstreamServer.URL),
			},

			req:  mustRequest(t, "/status/200?q=1", "PATCH"),
			want: proxied,
		},
		{
			name: "DELETE requests will not be stored in cache",
			handler: handler{
				cacher: func() Cacher {
					mock := NewMockCacher(ctrl)
					mock.EXPECT().LookUp(nil, nil).Times(0)
					mock.EXPECT().Store(nil, nil, nil).Times(0)

					return mock
				}(),
				metricsCollector: func() MetricsCollector {
					mock := NewMockMetricsCollector(ctrl)
					mock.EXPECT().CacheMiss(nil, nil).Times(0)

					return mock
				}(),
				downstreamURL: mustURL(t, downstreamServer.URL),
			},

			req:  mustRequest(t, "/status/200?q=1", "DELETE"),
			want: proxied,
		},
		{
			name: "statusCode > 500 from downstream will not be stored in cache",
			handler: handler{
				cacher: func() Cacher {
					mock := NewMockCacher(ctrl)
					mock.EXPECT().LookUp(gomock.Any(), gomock.Any()).Return(nil, nil).Times(1)
					mock.EXPECT().Store(nil, nil, nil).Times(0)

					return mock
				}(),
				metricsCollector: func() MetricsCollector {
					mock := NewMockMetricsCollector(ctrl)
					mock.EXPECT().CacheMiss("GET", http.StatusInternalServerError).Times(1)

					return mock
				}(),
				downstreamURL: mustURL(t, downstreamServerNok.URL),
			},

			req:  mustRequest(t, "/status/500?q=1", ""),
			want: http.StatusInternalServerError,
		},
		{
			name: "cache hit",
			handler: handler{
				cacher: func() Cacher {
					mock := NewMockCacher(ctrl)
					mock.EXPECT().LookUp(gomock.Any(), gomock.Any()).Return(&model.Response{
						Status: http.StatusOK,
						Body:   []byte("hello"),
					}, nil).Times(1)
					mock.EXPECT().Store(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

					return mock
				}(),
				metricsCollector: func() MetricsCollector {
					mock := NewMockMetricsCollector(ctrl)
					mock.EXPECT().CacheHit("GET", http.StatusOK).Times(1)

					return mock
				}(),
				downstreamURL: mustURL(t, downstreamServer.URL),
			},

			req:  mustRequest(t, "/status/200?q=1", ""),
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
