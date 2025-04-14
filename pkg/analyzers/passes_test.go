package analyzers

import (
	"reflect"
	"testing"

	"golang.org/x/tools/go/analysis"
)

func TestPassesAnalyzers(t *testing.T) {
	tests := []struct {
		name string
		want []*analysis.Analyzer
	}{
		{
			name: "passes",
			want: []*analysis.Analyzer{
				{
					Name: "asmdecl",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := PassesAnalyzers(); !reflect.DeepEqual(got[0].Name, tt.want[0].Name) {
				t.Errorf("PassesAnalyzers() = %v, want %v", got[0], tt.want[0].Name)
			}
		})
	}
}
