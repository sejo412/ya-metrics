package analyzers

import (
	"reflect"
	"testing"

	"golang.org/x/tools/go/analysis"
)

func TestSQLPassCtxCheckAnalyzer(t *testing.T) {
	tests := []struct {
		want *analysis.Analyzer
		name string
	}{
		{
			name: "sql passctx check",
			want: &analysis.Analyzer{
				Name: "sqlpassctxcheck",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SQLPassCtxCheckAnalyzer(); !reflect.DeepEqual(got.Name, tt.want.Name) {
				t.Errorf("SQLPassCtxCheckAnalyzer() = %v, want %v", got.Name, tt.want.Name)
			}
		})
	}
}
