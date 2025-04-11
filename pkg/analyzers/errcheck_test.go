package analyzers

import (
	"reflect"
	"testing"

	"golang.org/x/tools/go/analysis"
)

func TestErrCheckAnalyzer(t *testing.T) {
	tests := []struct {
		name string
		want *analysis.Analyzer
	}{
		{
			name: "ErrCheckAnalyzer",
			want: &analysis.Analyzer{
				Name: "errcheck",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ErrCheckAnalyzer(); !reflect.DeepEqual(got.Name, tt.want.Name) {
				t.Errorf("ErrCheckAnalyzer() = %v, want %v", got.Name, tt.want.Name)
			}
		})
	}
}
