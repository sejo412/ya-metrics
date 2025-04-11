package analyzers

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func Test_run(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), ExitCheckAnalyzer, "./...")
}
