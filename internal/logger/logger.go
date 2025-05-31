// Package logger implements logging functions.
package logger

import (
	"net/http"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger struct {
	Logger *zap.SugaredLogger
	Level  zap.AtomicLevel
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

func (l *Logger) WithLogging(h http.Handler) http.Handler {
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
		l.Logger.Infow(
			"incoming request",
			"uri", r.RequestURI,
			"method", r.Method,
			"status", responseData.status,
			"duration", duration,
			"size", responseData.size)
	}
	return http.HandlerFunc(fn)
}

func (l *Logger) IntToLevel(level int) zapcore.Level {
	return zapcore.Level(level)
}

func MustNewLogger(debug bool) *Logger {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	lvl := zap.NewAtomicLevel()
	if debug {
		lvl.SetLevel(zap.DebugLevel)
	} else {
		lvl.SetLevel(zap.InfoLevel)
	}
	res := &Logger{Logger: logger.Sugar(), Level: lvl}
	return res
}
