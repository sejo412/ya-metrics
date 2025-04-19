package utils

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
)

// Hash returns string hashed by key.
func Hash(data []byte, key string) string {
	h := hmac.New(sha256.New, []byte(key))
	h.Write(data)
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

// Encode encodes data with rsa public key.
func Encode(data []byte, key *rsa.PublicKey) ([]byte, error) {
	h := sha256.New()
	random := rand.Reader
	dataLen := len(data)
	ks := key.Size()
	hs := h.Size()
	step := ks - 2*hs - 2
	if step < 1 {
		encryptedBytes, err := rsa.EncryptOAEP(h, random, key, data, nil)
		if err != nil {
			return nil, fmt.Errorf("error encrypt data: %w", err)
		}
		return encryptedBytes, nil
	}

	var encryptedBytes []byte
	for start := 0; start < dataLen; start += step {
		finish := start + step
		if finish > dataLen {
			finish = dataLen
		}
		encryptedBlockBytes, err := rsa.EncryptOAEP(h, random, key, data[start:finish], nil)
		if err != nil {
			return nil, fmt.Errorf("error encrypt block: %w", err)
		}
		encryptedBytes = append(encryptedBytes, encryptedBlockBytes...)
	}
	return encryptedBytes, nil
}

// Decode decodes data with rsa private key.
func Decode(data []byte, key *rsa.PrivateKey) ([]byte, error) {
	h := sha256.New()
	random := rand.Reader
	dataLen := len(data)
	step := key.PublicKey.Size()
	if step < 1 {
		decryptedBytes, err := rsa.DecryptOAEP(h, random, key, data, nil)
		if err != nil {
			return nil, fmt.Errorf("error decrypt data: %w", err)
		}
		return decryptedBytes, nil
	}

	var decryptedBytes []byte
	for start := 0; start < dataLen; start += step {
		finish := start + step
		if finish > dataLen {
			finish = dataLen
		}
		decryptedBlockBytes, err := rsa.DecryptOAEP(h, random, key, data[start:finish], nil)
		if err != nil {
			return nil, fmt.Errorf("error decrypt block: %w", err)
		}
		decryptedBytes = append(decryptedBytes, decryptedBlockBytes...)
	}
	return decryptedBytes, nil
}

// LoadRSAPrivateKey loads rsa private key.
func LoadRSAPrivateKey(key []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(key)
	if block == nil || block.Type != "RSA PRIVATE KEY" {
		return nil, fmt.Errorf("failed to decode PEM block %+v", block)
	}
	return x509.ParsePKCS1PrivateKey(block.Bytes)
}

// LoadRSAPublicKey loads rsa public key.
func LoadRSAPublicKey(key []byte) (*rsa.PublicKey, error) {
	block, _ := pem.Decode(key)
	if block == nil || block.Type != "RSA PUBLIC KEY" {
		return nil, fmt.Errorf("failed to decode PEM block")
	}
	return x509.ParsePKCS1PublicKey(block.Bytes)
}
