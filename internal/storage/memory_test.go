package storage

import (
	"reflect"
	"testing"
)

func TestMemoryStorage_Add(t *testing.T) {
	type fields struct {
		Metrics []Metric
	}
	type args struct {
		metric Metric
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &MemoryStorage{
				Metrics: tt.fields.Metrics,
			}
			s.Add(tt.args.metric)
		})
	}
}

func TestMemoryStorage_Last(t *testing.T) {
	type fields struct {
		Metrics []Metric
	}
	type args struct {
		metric Metric
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   Metric
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &MemoryStorage{
				Metrics: tt.fields.Metrics,
			}
			if got := s.Last(tt.args.metric); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Last() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMemoryStorage_LastAll(t *testing.T) {
	type fields struct {
		Metrics []Metric
	}
	tests := []struct {
		name   string
		fields fields
		want   []Metric
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &MemoryStorage{
				Metrics: tt.fields.Metrics,
			}
			if got := s.LastAll(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LastAll() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewMemoryStorage(t *testing.T) {
	tests := []struct {
		name string
		want *MemoryStorage
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewMemoryStorage(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewMemoryStorage() = %v, want %v", got, tt.want)
			}
		})
	}
}
