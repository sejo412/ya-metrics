package config

import (
	"reflect"
	"testing"
)

func TestGetVersion(t *testing.T) {
	tests := []struct {
		name string
		want Version
	}{
		{
			name: "should return N/A",
			want: Version{
				BuildVersion: defaultVersion,
				BuildDate:    defaultVersion,
				BuildCommit:  defaultVersion,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetVersion(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}
