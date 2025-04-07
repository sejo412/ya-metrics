package analyzers

import (
	"github.com/kisielk/errcheck/errcheck"
	"golang.org/x/tools/go/analysis"
)

// ErrCheckAnalyzer returns `errcheck` analyzer from `github.com/kisielk/errcheck/errcheck` module.
func ErrCheckAnalyzer() *analysis.Analyzer {
	return errcheck.Analyzer
}
