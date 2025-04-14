package analyzers

import (
	"reflect"
	"testing"

	"golang.org/x/tools/go/analysis"
)

func TestStaticCheckQFAnalyzer(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		want *analysis.Analyzer
		name string
		args args
	}{
		{
			name: "Static Check QF Analyzer",
			args: args{
				name: "",
			},
			want: &analysis.Analyzer{
				Name: defaultStaticCheckQFAnalyzer,
			},
		},
		{
			name: "nil Static Check QF Analyzer",
			args: args{
				name: "zzz",
			},
			want: &analysis.Analyzer{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StaticCheckQFAnalyzer(tt.args.name); !reflect.DeepEqual(got.Name, tt.want.Name) {
				t.Errorf("StaticCheckQFAnalyzer() = %v, want %v", got.Name, tt.want.Name)
			}
		})
	}
}

func TestStaticCheckSAAnalyzers(t *testing.T) {
	tests := []struct {
		name string
		want []*analysis.Analyzer
	}{
		{
			name: "Static Check SA Analyzer",
			want: []*analysis.Analyzer{
				{
					Name: "SA1000",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StaticCheckSAAnalyzers(); !reflect.DeepEqual(got[0].Name, tt.want[0].Name) {
				t.Errorf("StaticCheckSAAnalyzers() = %v, want %v", got[0].Name, tt.want[0].Name)
			}
		})
	}
}

func TestStaticCheckSAnalyzer(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		want *analysis.Analyzer
		name string
		args args
	}{
		{
			name: "Static Check S Analyzer",
			args: args{
				name: "",
			},
			want: &analysis.Analyzer{
				Name: defaultStaticCheckSAnalyzer,
			},
		},
		{
			name: "nil Static Check S Analyzer",
			args: args{
				name: "zzz",
			},
			want: &analysis.Analyzer{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StaticCheckSAnalyzer(tt.args.name); !reflect.DeepEqual(got.Name, tt.want.Name) {
				t.Errorf("StaticCheckSAnalyzer() = %v, want %v", got.Name, tt.want.Name)
			}
		})
	}
}

func TestStaticCheckSTAnalyzer(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		want *analysis.Analyzer
		name string
		args args
	}{
		{
			name: "Static Check ST Analyzer",
			args: args{
				name: "",
			},
			want: &analysis.Analyzer{
				Name: "ST1005",
			},
		},
		{
			name: "nil Static Check ST Analyzer",
			args: args{
				name: "zzz",
			},
			want: &analysis.Analyzer{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StaticCheckSTAnalyzer(tt.args.name); !reflect.DeepEqual(got.Name, tt.want.Name) {
				t.Errorf("StaticCheckSTAnalyzer() = %v, want %v", got.Name, tt.want.Name)
			}
		})
	}
}
