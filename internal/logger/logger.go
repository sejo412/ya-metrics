// Package logger implements logging functions.
package logger

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

type Logger struct {
	*zap.SugaredLogger
}

type Middleware struct {
	Logger *zap.SugaredLogger
}

type responseData struct {
	status int
	size   int
}

type LoggingResponseWriter struct {
	http.ResponseWriter
	ResponseData *responseData
}

func (r *LoggingResponseWriter) Write(data []byte) (int, error) {
	size, err := r.ResponseWriter.Write(data)
	r.ResponseData.size += size
	return size, err
}

func (r *LoggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.ResponseData.status = statusCode
}

func NewMiddleware(logger *zap.SugaredLogger) *Middleware {
	return &Middleware{Logger: logger}
}

func (lm *Middleware) WithLogging(h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		responseData := &responseData{
			status: 200,
			size:   0,
		}
		lw := LoggingResponseWriter{
			ResponseWriter: w,
			ResponseData:   responseData,
		}

		h.ServeHTTP(&lw, r)
		duration := time.Since(start)
		lm.Logger.Infow(
			"incoming request",
			"uri", r.RequestURI,
			"method", r.Method,
			"status", responseData.status,
			"duration", duration,
			"size", responseData.size)
	}
	return http.HandlerFunc(fn)
}

func NewLogger() (*Logger, error) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		return nil, err
	}
	res := &Logger{logger.Sugar()}
	return res, nil
}
