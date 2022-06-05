package http

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	gomock "github.com/golang/mock/gomock"
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

func mustRequest(t *testing.T, url string) *http.Request {
	req, err := http.NewRequest("GET", url, nil)
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

func TestStaleWhileRevalidate(t *testing.T) {
	// ctx := context.Background()
	ctrl := gomock.NewController(t)
	downstreamServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer downstreamServer.Close()

	tests := []struct {
		name    string
		handler handler
		req     *http.Request
		want    *httptest.ResponseRecorder
	}{
		{
			name: "stale-while-revalidate",
			handler: handler{
				cacher: func() Cacher {
					mock := NewMockCacher(ctrl)
					mock.EXPECT().LookUp(gomock.Any(), gomock.Any()).Return(nil, nil).Times(1)
					mock.EXPECT().Store(gomock.Any(), gomock.Any(), newResponseMatcher(http.StatusOK, nil)).Return(nil).Times(1)

					return mock
				}(),
				metricsCollector: func() MetricsCollector {
					mock := NewMockMetricsCollector(ctrl)
					mock.EXPECT().CacheMiss("GET", http.StatusOK).Times(1)

					return mock
				}(),
				downstreamURL: mustURL(t, downstreamServer.URL),
			},

			req:  mustRequest(t, "/status/200?q=1"),
			want: httptest.NewRecorder(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.handler.ServeHTTP(tt.want, tt.req)
		})
	}
}
