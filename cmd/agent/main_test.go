package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_incPollCount(t *testing.T) {
	tests := []struct {
		name string
		want int64
	}{
		{
			name: "just increment",
			want: 413,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var i int64
			i = 412
			incPollCount(&i)
			assert.Equal(t, tt.want, i)
		})
	}
}

// Не очень понятно как можно протестировать остальное. Там по-сути никакого интерактива