package server

import (
	"net/http"

	"go.uber.org/zap"
)

type LoggerMiddleware struct {
	Logger *zap.SugaredLogger
}

type responseData struct {
	status int
	size   int
}

type loggingResponseWriter struct {
	http.ResponseWriter
	responseData *responseData
}

func (r *loggingResponseWriter) Write(data []byte) (int, error) {
	size, err := r.ResponseWriter.Write(data)
	r.responseData.size += size
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}

func NewLoggerMiddleware(logger *zap.SugaredLogger) *LoggerMiddleware {
	return &LoggerMiddleware{Logger: logger}
}
