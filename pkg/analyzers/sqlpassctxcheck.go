package analyzers

import (
	"github.com/400f/sqlpassctxcheck"
	"golang.org/x/tools/go/analysis"
)

// SQLPassCtxCheckAnalyzer returns analyzer from `github.com/400f/sqlpassctxcheck` module.
// sqlpassctxcheck is a program for checking for sql module method call without ctx.
// Using this tool, you can avoid falling outside of distributed tracing by forgetting to pass the context.
func SQLPassCtxCheckAnalyzer() *analysis.Analyzer {
	return sqlpassctxcheck.Analyzer
}
