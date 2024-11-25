package main

import (
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
				response: messageNotFloat,
			},
		},
		{
			name:    "bad counter",
			request: "http://localhost:8080/update/counter/Frees/10.55",
			want: want{
				code:     http.StatusBadRequest,
				response: messageNotInt,
			},
		},
		{
			name:    "bad type",
			request: "http://localhost:8080/update/preved/Frees/10",
			want: want{
				code:     http.StatusBadRequest,
				response: messageNotSupported,
			},
		},
		{
			name:    "too short",
			request: "http://localhost:8080/update/gauge/10",
			want: want{
				code:     http.StatusNotFound,
				response: messageNotFound,
			},
		},
		{
			name:    "too long",
			request: "http://localhost:8080/update/gauge/Frees/subfree/10",
			want: want{
				code:     http.StatusNotFound,
				response: messageNotFound,
			},
		},
		{
			name:    "index not found",
			request: "http://localhost:8080",
			want: want{
				code:     http.StatusNotFound,
				response: messageNotFound,
			},
		},
		{
			name:    "other not found",
			request: "http://localhost:8080/updateeee/qwe/asd",
			want: want{
				code:     http.StatusNotFound,
				response: messageNotFound,
			},
		}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, tt.request, nil)
			w := httptest.NewRecorder()
			handleUpdate(w, request)
			res := w.Result()
			assert.Equal(t, tt.want.code, res.StatusCode)
			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)
			require.NoError(t, err)
			assert.Equal(t, tt.want.response, strings.TrimSuffix(string(resBody), "\n"))
		})
	}
}
