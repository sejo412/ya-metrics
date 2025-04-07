package main

import (
	"github.com/sejo412/ya-metrics/pkg/analyzers"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
)

func main() {
	checks := make([]*analysis.Analyzer, 0)
	checks = append(checks, analyzers.PassesAnalyzers()...)
	checks = append(checks, analyzers.StaticCheckSAAnalyzers()...)
	checks = append(checks, analyzers.StaticCheckSAnalyzer())
	checks = append(checks, analyzers.StaticCheckSTAnalyzer())
	checks = append(checks, analyzers.StaticCheckQFAnalyzer())
	checks = append(checks, analyzers.ErrCheckAnalyzer())
	checks = append(checks, analyzers.SQLPassCtxCheckAnalyzer())
	checks = append(checks, analyzers.ExitCheckAnalyzer)

	multichecker.Main(
		checks...,
	)
}
