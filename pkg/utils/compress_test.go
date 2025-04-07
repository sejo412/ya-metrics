package utils

import (
	"reflect"
	"testing"
)

var testCompressedData = &[]byte{}

func TestCompress(t *testing.T) {
	type args struct {
		data []byte
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "TestCompress",
			args: args{
				data: []byte("hello world"),
			},
			wantErr: false,
		},
	}
	var err error
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			*testCompressedData, err = Compress(tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("Compress() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestDecompress(t *testing.T) {
	type args struct {
		data []byte
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "TestDecompress",
			args: args{
				data: *testCompressedData,
			},
			want:    []byte("hello world"),
			wantErr: false,
		},
		{
			name: "TestDecompress error",
			args: args{
				data: []byte("hello world"),
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Decompress(tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("Decompress() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Decompress() got = %v, want %v", got, tt.want)
			}
		})
	}
}
