package main

import (
	"fmt"
	. "github.com/sejo412/ya-metrics/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func Test_handleUpdate(t *testing.T) {
	type want struct {
		code     int
		response string
	}
	tests := []struct {
		name    string
		request string
		want    want
	}{
		{
			name:    "ok gauge",
			request: "http://localhost:8080/update/gauge/Frees/10.909",
			want: want{
				code:     http.StatusOK,
				response: "",
			},
		},
		{
			name:    "ok counter",
			request: "http://localhost:8080/update/counter/Frees/10",
			want: want{
				code:     http.StatusOK,
				response: "",
			},
		},
		{
			name:    "bad gauge",
			request: "http://localhost:8080/update/gauge/Frees/preved",
			want: want{
				code:     http.StatusBadRequest,
				response: fmt.Sprintf("%s: %s", ErrHttpBadRequest, MessageNotFloat),
			},
		},
		{
			name:    "bad counter",
			request: "http://localhost:8080/update/counter/Frees/10.55",
			want: want{
				code:     http.StatusBadRequest,
				response: fmt.Sprintf("%s: %s", ErrHttpBadRequest, MessageNotInteger),
			},
		},
		{
			name:    "bad type",
			request: "http://localhost:8080/update/preved/Frees/10",
			want: want{
				code:     http.StatusBadRequest,
				response: fmt.Sprintf("%s: %s", ErrHttpBadRequest, MessageNotSupported),
			},
		},
		{
			name:    "too short",
			request: "http://localhost:8080/update/gauge/10",
			want: want{
				code:     http.StatusNotFound,
				response: fmt.Sprintf("%s", ErrHttpNotFound),
			},
		},
		{
			name:    "too long",
			request: "http://localhost:8080/update/gauge/Frees/subfree/10",
			want: want{
				code:     http.StatusNotFound,
				response: fmt.Sprintf("%s", ErrHttpNotFound),
			},
		},
		{
			name:    "index not found",
			request: "http://localhost:8080",
			want: want{
				code:     http.StatusNotFound,
				response: fmt.Sprintf("%s", ErrHttpNotFound),
			},
		},
		{
			name:    "other not found",
			request: "http://localhost:8080/updateeee/qwe/asd",
			want: want{
				code:     http.StatusNotFound,
				response: fmt.Sprintf("%s", ErrHttpNotFound),
			},
		}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, tt.request, nil)
			w := httptest.NewRecorder()
			postUpdate(w, request)
			res := w.Result()
			assert.Equal(t, tt.want.code, res.StatusCode)
			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)
			require.NoError(t, err)
			assert.Equal(t, tt.want.response, strings.TrimSuffix(string(resBody), "\n"))
		})
	}
}
