package logger

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"go.uber.org/zap"
)

var testSugaredLogger, _ = zap.NewDevelopment()

func TestLoggingResponseWriter_Write(t *testing.T) {
	var testData = []byte("hello world")
	type fields struct {
		ResponseWriter http.ResponseWriter
		ResponseData   *responseData
	}
	type args struct {
		data []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    int
		wantErr bool
	}{
		{
			name: "success",
			fields: fields{
				ResponseWriter: httptest.NewRecorder(),
				ResponseData:   &responseData{},
			},
			args: args{
				data: testData,
			},
			want:    len(testData),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &LoggingResponseWriter{
				ResponseWriter: tt.fields.ResponseWriter,
				ResponseData:   tt.fields.ResponseData,
			}
			got, err := r.Write(tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("Write() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Write() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLoggingResponseWriter_WriteHeader(t *testing.T) {
	type fields struct {
		ResponseWriter http.ResponseWriter
		ResponseData   *responseData
	}
	type args struct {
		statusCode int
	}
	tests := []struct {
		fields fields
		name   string
		args   args
	}{
		{
			name: "status code 200",
			fields: fields{
				ResponseWriter: httptest.NewRecorder(),
				ResponseData:   &responseData{},
			},
			args: args{
				statusCode: http.StatusOK,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &LoggingResponseWriter{
				ResponseWriter: tt.fields.ResponseWriter,
				ResponseData:   tt.fields.ResponseData,
			}
			r.WriteHeader(tt.args.statusCode)
			if r.ResponseData.status != tt.args.statusCode {
				t.Errorf("WriteHeader() got = %v, want %v", r.ResponseData.status, tt.args.statusCode)
			}
		})
	}
}

func TestNewLogger(t *testing.T) {
	tests := []struct {
		want *Logger
		name string
	}{
		{
			name: "MustNewLogger",
			want: &Logger{
				Logger: testSugaredLogger.Sugar(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MustNewLogger(false)
			if !reflect.DeepEqual(got.Logger.Level(), tt.want.Logger.Level()) {
				t.Errorf("MustNewLogger() got = %v, want %v", got.Logger.Level(), tt.want.Logger.Level())
			}
		})
	}
}
