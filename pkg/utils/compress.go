package utils

import (
	"bytes"
	"compress/gzip"
	"fmt"
)

// Decompress returns unzipped data or error.
func Decompress(data []byte) ([]byte, error) {
	reader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("decompress error: %w", err)
	}
	defer func() {
		_ = reader.Close()
	}()
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(reader); err != nil {
		return nil, fmt.Errorf("decompress error: %w", err)
	}
	return buf.Bytes(), nil
}

// Compress returns gzipped data or error.
func Compress(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	w, err := gzip.NewWriterLevel(&buf, gzip.BestSpeed)
	if err != nil {
		return nil, fmt.Errorf("compress error: %w", err)
	}
	if _, err := w.Write(data); err != nil {
		return nil, fmt.Errorf("compress error: %w", err)
	}
	if err := w.Close(); err != nil {
		return nil, fmt.Errorf("compress error: %w", err)
	}
	return buf.Bytes(), nil
}
