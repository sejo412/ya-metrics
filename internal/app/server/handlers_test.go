package server

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sejo412/ya-metrics/internal/config"
	logger2 "github.com/sejo412/ya-metrics/internal/logger"
	m "github.com/sejo412/ya-metrics/internal/models"
	"github.com/sejo412/ya-metrics/internal/storage"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

var cfg = config.ServerConfig{
	Address:       ":8080",
	StoreInterval: 30,
	StoreFile:     "/tmp/testing_metrics.json",
	Restore:       new(bool),
}

const notFound = "404 page not found"

func testRequest(t *testing.T, ts *httptest.Server, method, path string, header http.Header,
	body io.Reader) (*http.Response, string) {
	ctx := context.TODO()
	req, err := http.NewRequestWithContext(ctx, method, ts.URL+path, body)
	req.Header = header
	if err != nil {
		t.Fatal(err)
		return nil, ""
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
		return nil, ""
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
		return nil, ""
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	result, _ := strings.CutSuffix(string(respBody), "\n")
	return resp, result
}

func Test_handleUpdate(t *testing.T) {
	type want struct {
		response string
		code     int
	}
	tests := []struct {
		name    string
		request string
		header  http.Header
		want    want
	}{
		{
			name:    "ok gauge",
			request: "/update/gauge/testGauge/10.909",
			want: want{
				code:     http.StatusOK,
				response: "",
			},
		},
		{
			name:    "ok counter",
			request: "/update/counter/testCounter/10",
			want: want{
				code:     http.StatusOK,
				response: "",
			},
		},
		{
			name:    "bad gauge",
			request: "/update/gauge/Frees/preved",
			want: want{
				code:     http.StatusBadRequest,
				response: fmt.Sprintf("%s: %s", m.ErrHTTPBadRequest, m.MessageNotFloat),
			},
		},
		{
			name:    "bad counter",
			request: "/update/counter/Frees/10.55",
			want: want{
				code:     http.StatusBadRequest,
				response: fmt.Sprintf("%s: %s", m.ErrHTTPBadRequest, m.MessageNotInteger),
			},
		},
		{
			name:    "bad type",
			request: "/update/preved/Frees/10",
			want: want{
				code:     http.StatusBadRequest,
				response: fmt.Sprintf("%s: %s", m.ErrHTTPBadRequest, m.MessageNotSupported),
			},
		},
		{
			name:    "too short",
			request: "/update/gauge/10",
			want: want{
				code:     http.StatusNotFound,
				response: notFound,
			},
		},
		{
			name:    "too long",
			request: "/update/gauge/Frees/subfree/10",
			want: want{
				code:     http.StatusNotFound,
				response: notFound,
			},
		},
		{
			name:    "POST 405",
			request: "",
			want: want{
				code:     http.StatusMethodNotAllowed,
				response: "",
			},
		},
		{
			name:    "other not found",
			request: "/updateeee/qwe/asd",
			want: want{
				code:     http.StatusNotFound,
				response: notFound,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, _ := zap.NewDevelopment()
			defer func() {
				_ = logger.Sync()
			}()
			sugar := logger.Sugar()
			lm := logger2.NewMiddleware(sugar)
			store := storage.NewMemoryStorage()

			r := NewRouterWithConfig(&config.Options{
				Config:  cfg,
				Storage: store,
			}, lm)
			ts := httptest.NewServer(r)
			defer ts.Close()
			resp, body := testRequest(t, ts, http.MethodPost, tt.request, nil, nil)
			defer func() {
				_ = resp.Body.Close()
			}()
			assert.Equal(t, tt.want.code, resp.StatusCode, tt.name)
			assert.Equal(t, tt.want.response, body, tt.name)
		})
	}
}
func Test_getIndex(t *testing.T) {
	// notFound := "404 page not found"
	type want struct {
		response string
		code     int
	}
	tests := []struct {
		name    string
		request string
		want    want
	}{
		{
			name:    "ok index",
			request: "/",
			want: want{
				code:     http.StatusOK,
				response: "<!DOCTYPE html>",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, _ := zap.NewDevelopment()
			defer func() {
				_ = logger.Sync()
			}()
			sugar := logger.Sugar()
			lm := logger2.NewMiddleware(sugar)
			store := storage.NewMemoryStorage()

			r := NewRouterWithConfig(&config.Options{
				Config:  cfg,
				Storage: store,
			}, lm)
			ts := httptest.NewServer(r)
			defer ts.Close()
			resp, body := testRequest(t, ts, http.MethodGet, tt.request, nil, nil)
			defer func() {
				_ = resp.Body.Close()
			}()
			assert.Equal(t, tt.want.code, resp.StatusCode, tt.name)
			assert.Contains(t, body, tt.want.response, tt.name)
		})
	}
}

func Test_postUpdateJSON(t *testing.T) {
	gzippedBase64invalid := "H4sICDwDBWgAA3Rlc3QAKyhKLUtN4QIAlOowhwcAAAA="
	gzippedInvalid, _ := base64.StdEncoding.DecodeString(gzippedBase64invalid)
	type want struct {
		response string
		code     int
	}
	tests := []struct {
		name    string
		request string
		header  http.Header
		body    io.Reader
		want    want
	}{
		{
			name:    "not json",
			request: "/update/",
			header: http.Header{
				"Content-Type": []string{"text/plain"},
			},
			body: bytes.NewBuffer([]byte(`foo=bar`)),
			want: want{
				code:     http.StatusBadRequest,
				response: m.ErrHTTPBadRequest.Error(),
			},
		},
		{
			name:    "invalid json",
			request: "/update/",
			header: http.Header{
				"Content-Type": []string{"application/json"},
			},
			body: bytes.NewBuffer([]byte(`{"value": "foo""}`)),
			want: want{
				code:     http.StatusBadRequest,
				response: m.ErrHTTPBadRequest.Error(),
			},
		},
		{
			name:    "valid gzip invalid json",
			request: "/update/",
			header: http.Header{
				m.HTTPHeaderContentEncoding: []string{"gzip"},
			},
			body: bytes.NewBuffer(gzippedInvalid),
			want: want{
				code:     http.StatusBadRequest,
				response: m.ErrHTTPBadRequest.Error(),
			},
		},
		{
			name:    "invalid gzip",
			request: "/update/",
			header: http.Header{
				m.HTTPHeaderContentEncoding: []string{"gzip"},
			},
			body: bytes.NewBuffer([]byte(`zzz`)),
			want: want{
				code:     http.StatusBadRequest,
				response: "decompress error: unexpected EOF",
			},
		},
		{
			name:    "valid json",
			request: "/update/",
			header: http.Header{
				m.HTTPHeaderContentType: []string{"application/json"},
			},
			body: bytes.NewBuffer([]byte(`{"type": "gauge", "value": 99.11, "id": "testGauge90"}`)),
			want: want{
				code:     http.StatusOK,
				response: `{"value":99.11,"id":"testGauge90","type":"gauge"}`,
			},
		},
	}
	logger, _ := zap.NewDevelopment()
	defer func() {
		_ = logger.Sync()
	}()
	sugar := logger.Sugar()
	lm := logger2.NewMiddleware(sugar)
	r := NewRouterWithConfig(&config.Options{
		Config:  cfg,
		Storage: storage.NewMemoryStorage(),
	}, lm)
	ts := httptest.NewServer(r)
	defer ts.Close()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, body := testRequest(t, ts, http.MethodPost, tt.request, tt.header, tt.body)
			defer func() {
				_ = resp.Body.Close()
			}()
			assert.Equal(t, tt.want.code, resp.StatusCode, tt.name)
			assert.Equal(t, tt.want.response, body, tt.name)
		})
	}
}
func Test_postUpdatesJSON(t *testing.T) {
	type want struct {
		code int
	}
	tests := []struct {
		body    io.Reader
		header  http.Header
		name    string
		request string
		want    want
	}{
		{
			name:    "not json",
			request: "/updates/",
			header: http.Header{
				"Content-Type": []string{"text/plain"},
			},
			body: bytes.NewBuffer([]byte(`foo=bar`)),
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name:    "invalid json",
			request: "/updates/",
			header: http.Header{
				"Content-Type": []string{"application/json"},
			},
			body: bytes.NewBuffer([]byte(`{"value": "foo""}`)),
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name:    "valid json",
			request: "/updates/",
			header: http.Header{
				m.HTTPHeaderContentType: []string{"application/json"},
			},
			body: bytes.NewBuffer([]byte(`[{"type": "gauge", "value": 99.11, "id": "testGauge90"}]`)),
			want: want{
				code: http.StatusOK,
			},
		},
	}
	logger, _ := zap.NewDevelopment()
	defer func() {
		_ = logger.Sync()
	}()
	sugar := logger.Sugar()
	lm := logger2.NewMiddleware(sugar)
	r := NewRouterWithConfig(&config.Options{
		Config:  cfg,
		Storage: storage.NewMemoryStorage(),
	}, lm)
	ts := httptest.NewServer(r)
	defer ts.Close()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, _ := testRequest(t, ts, http.MethodPost, tt.request, tt.header, tt.body)
			defer func() {
				_ = resp.Body.Close()
			}()
			assert.Equal(t, tt.want.code, resp.StatusCode, tt.name)
		})
	}
}

func TestRouter_pingStorage(t *testing.T) {
	type want struct {
		response string
		code     int
	}
	tests := []struct {
		name    string
		request string
		want    want
	}{
		{
			name:    "ping",
			request: "/ping",
			want: want{
				code: http.StatusOK,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, _ := zap.NewDevelopment()
			defer func() {
				_ = logger.Sync()
			}()
			sugar := logger.Sugar()
			lm := logger2.NewMiddleware(sugar)
			store := storage.NewMemoryStorage()

			r := NewRouterWithConfig(&config.Options{
				Config:  cfg,
				Storage: store,
			}, lm)
			ts := httptest.NewServer(r)
			defer ts.Close()
			resp, body := testRequest(t, ts, http.MethodGet, tt.request, nil, nil)
			defer func() {
				_ = resp.Body.Close()
			}()
			assert.Equal(t, tt.want.code, resp.StatusCode, tt.name)
			assert.Contains(t, body, tt.want.response, tt.name)
		})
	}
}

func TestRouter_getMetricJSON(t *testing.T) {
	type want struct {
		response string
		code     int
	}
	tests := []struct {
		name   string
		header http.Header
		body   io.Reader
		want   want
	}{
		{
			name: "ok",
			header: http.Header{
				"Content-Type": []string{"application/json"},
			},
			body: bytes.NewBuffer([]byte(`{"id": "testGauge90", "type": "gauge"}`)),
			want: want{
				code:     http.StatusOK,
				response: `{"value":99.11,"id":"testGauge90","type":"gauge"}`,
			},
		},
		{
			name: "404 not found",
			header: http.Header{
				"Content-Type": []string{"application/json"},
			},
			body: bytes.NewBuffer([]byte(`{"id": "testGauge9", "type": "gauge"}`)),
			want: want{
				code:     http.StatusNotFound,
				response: "not found",
			},
		},
	}
	logger, _ := zap.NewDevelopment()
	defer func() {
		_ = logger.Sync()
	}()
	sugar := logger.Sugar()
	lm := logger2.NewMiddleware(sugar)
	store := storage.NewMemoryStorage()
	r := NewRouterWithConfig(&config.Options{
		Config:  cfg,
		Storage: store,
	}, lm)
	ts := httptest.NewServer(r)
	defer ts.Close()
	_ = store.Upsert(context.Background(), m.Metric{
		Kind:  "gauge",
		Name:  "testGauge90",
		Value: "99.11",
	})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, body := testRequest(t, ts, http.MethodPost, "/value/", tt.header, tt.body)
			defer func() {
				_ = resp.Body.Close()
			}()
			assert.Equal(t, tt.want.code, resp.StatusCode, tt.name)
			assert.Equal(t, tt.want.response, body, tt.name)
		})
	}
}

func TestRouter_getValue(t *testing.T) {
	type want struct {
		response string
		code     int
	}
	tests := []struct {
		name    string
		request string
		want    want
	}{
		{
			name:    "ok",
			request: "/value/gauge/testGauge90",
			want: want{
				code:     http.StatusOK,
				response: "99.11",
			},
		},
		{
			name:    "404 not found",
			request: "/value/gauge/testGauge91",
			want: want{
				code:     http.StatusNotFound,
				response: "not found",
			},
		},
	}
	logger, _ := zap.NewDevelopment()
	defer func() {
		_ = logger.Sync()
	}()
	sugar := logger.Sugar()
	lm := logger2.NewMiddleware(sugar)
	store := storage.NewMemoryStorage()
	r := NewRouterWithConfig(&config.Options{
		Config:  cfg,
		Storage: store,
	}, lm)
	ts := httptest.NewServer(r)
	defer ts.Close()
	_ = store.Upsert(context.Background(), m.Metric{
		Kind:  "gauge",
		Name:  "testGauge90",
		Value: "99.11",
	})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, body := testRequest(t, ts, http.MethodGet, tt.request, nil, nil)
			defer func() {
				_ = resp.Body.Close()
			}()
			assert.Equal(t, tt.want.code, resp.StatusCode, tt.name)
			assert.Equal(t, tt.want.response, body, tt.name)
		})
	}
}

func TestRouter_checkXRealIPHandler(t *testing.T) {
	type args struct {
		xRealIPHeader string
		request       string
	}
	type want struct {
		code int
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "ok",
			args: args{
				xRealIPHeader: "127.0.0.1",
			},
			want: want{
				code: http.StatusOK,
			},
		},
		{
			name: "forbidden",
			args: args{
				xRealIPHeader: "192.168.2.1",
			},
			want: want{
				code: http.StatusForbidden,
			},
		},
	}
	logger, _ := zap.NewDevelopment()
	defer func() {
		_ = logger.Sync()
	}()
	sugar := logger.Sugar()
	lm := logger2.NewMiddleware(sugar)
	store := storage.NewMemoryStorage()
	r := NewRouterWithConfig(&config.Options{
		Config:  cfg,
		Storage: store,
		TrustedSubnets: &[]net.IPNet{
			{
				IP:   []byte{127, 0, 0, 0},
				Mask: []byte{255, 0, 0, 0},
			},
		},
	}, lm)
	ts := httptest.NewServer(r)
	defer ts.Close()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			header := http.Header{
				"X-Real-IP":    []string{tt.args.xRealIPHeader},
				"content-type": []string{"application/json"},
			}
			body := bytes.NewBuffer([]byte(`{"id": "testGauge90", "type": "gauge", "value": 99.11}`))
			resp, _ := testRequest(t, ts, http.MethodPost, "/update/", header, body)
			assert.Equal(t, tt.want.code, resp.StatusCode, tt.name)
		})
	}
}
