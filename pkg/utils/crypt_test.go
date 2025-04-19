package utils

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"reflect"
	"testing"
)

var testPrivateKey = []byte(`
-----BEGIN RSA PRIVATE KEY-----
MIICXAIBAAKBgQDDIt7hoHP4Ip+dUlcQbst+qggL2OX4Gz7eM4Ah6+i8MJUEN6DE
H5wr3Sy9Tf42/UGss/p8Ump2kY5uMZpDEIP4LsLC8S64vruFryoDopkOC4V1huRJ
bN1+R+J+AmlV+bOBww3gv+REYic3AG7toDzm0DmjtLkqNtXwNatosadAOwIDAQAB
AoGAdMGvlFb6LLUixfIXkAiLD+3SxKvx5cL/mdo8x48tATUcZJqYQeEcA47iDx7U
hbiBDEHgFXUGqI0tKLfbMld2bgSN81HNPFGuOF/QoSRAGLAIWc+AzvoHq90B7RDi
PSC3DFSaCzcUaF0jRhPRs19DZAUH/PEkcl33BhFYdDafrtkCQQDmXWQkxTcWz993
dd7E/Y+21KTUrR+DtriklgEcQIgzP/tGWfdwZy2gu0//wJ8GxdRalPJGzx461pwc
ZJSWURklAkEA2NnkDwtb47ECqq0X9ALv58a+y82OcYAYTZ1B4hVqDARq2r/IAhH6
9DBmkw6AoJAnDdWeEKwkCFBir/nXaOAl3wJBANfXN0aQlh5EpM/MW+7c2TPoJ4yx
rS5/HJ/xgJbVDAhg8XGoSAREWGcaOkmaVCZHY8F/f0XDOELO5DRiNSpmUBUCQHrI
xS0PjXQbIhtp7wonL5fZHOdg+KqjkR9BT7Cn12f+iFJcDO+/Jo1lam8R4xsHBFX9
AocGMVDT00049hNX95kCQGFIMWEtQscEfKv3fUZ6lemjm6uMy3DSPoaWjE+89EMq
sGPE3EFS+q1NzqqMKVZVkigKbFQQXbcVbVHv45QYYXM=
-----END RSA PRIVATE KEY-----`)

var testPublicKey = []byte(`
-----BEGIN RSA PUBLIC KEY-----
MIGJAoGBAMMi3uGgc/gin51SVxBuy36qCAvY5fgbPt4zgCHr6LwwlQQ3oMQfnCvd
LL1N/jb9Qayz+nxSanaRjm4xmkMQg/guwsLxLri+u4WvKgOimQ4LhXWG5Els3X5H
4n4CaVX5s4HDDeC/5ERiJzcAbu2gPObQOaO0uSo21fA1q2ixp0A7AgMBAAE=
-----END RSA PUBLIC KEY-----`)

var testErrorKey = []byte(`zzzz`)

func TestLoadRSAPrivateKey(t *testing.T) {
	block, _ := pem.Decode(testPrivateKey)
	key, _ := x509.ParsePKCS1PrivateKey(block.Bytes)
	type args struct {
		key []byte
	}
	tests := []struct {
		want    *rsa.PrivateKey
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				key: testPrivateKey,
			},
			want:    key,
			wantErr: false,
		},
		{
			name: "error",
			args: args{
				key: testErrorKey,
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := LoadRSAPrivateKey(tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadRSAPrivateKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LoadRSAPrivateKey() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLoadRSAPublicKey(t *testing.T) {
	block, _ := pem.Decode(testPublicKey)
	key, _ := x509.ParsePKCS1PublicKey(block.Bytes)
	type args struct {
		key []byte
	}
	tests := []struct {
		want    *rsa.PublicKey
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				key: testPublicKey,
			},
			want:    key,
			wantErr: false,
		},
		{
			name: "error",
			args: args{
				key: testErrorKey,
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := LoadRSAPublicKey(tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadRSAPublicKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LoadRSAPublicKey() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEncode(t *testing.T) {
	block, _ := pem.Decode(testPublicKey)
	key, _ := x509.ParsePKCS1PublicKey(block.Bytes)
	type args struct {
		key  *rsa.PublicKey
		data []byte
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				data: []byte("hello world"),
				key:  key,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Encode(tt.args.data, tt.args.key)
			// fmt.Println(base64.StdEncoding.EncodeToString(res))
			if (err != nil) != tt.wantErr {
				t.Errorf("Encode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestDecode(t *testing.T) {
	block, _ := pem.Decode(testPrivateKey)
	key, _ := x509.ParsePKCS1PrivateKey(block.Bytes)
	encodedBase64 := "bzU0dksrWhjPjJFET7PiX+xwzu2zDNe7t4JMr0E2yKMWRUw3h9eupU4OzakcVv+G62ZcicyescsIk2yrTWReeZeBoISvFdiP8PB87kXYLBFIJ4VNYavDKc+mKQdNYdbMnCDHsPJtb1aF60Q6C0gxIRD01eKl1WrTDUd5HJjM+rI="
	encoded, _ := base64.StdEncoding.DecodeString(encodedBase64)
	type args struct {
		key  *rsa.PrivateKey
		data []byte
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				data: encoded,
				key:  key,
			},
			want:    []byte("hello world"),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Decode(tt.args.data, tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("Decode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Decode() got = %v, want %v", got, tt.want)
			}
		})
	}
}
