package main

import (
	"github.com/desepticon55/metrics-collector/cmd/staticlint"
	"github.com/gostaticanalysis/nilerr"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/unusedresult"
	"honnef.co/go/tools/staticcheck"
)

func main() {
	checks := []*analysis.Analyzer{
		shadow.Analyzer,
		unusedresult.Analyzer,
	}

	for _, v := range staticcheck.Analyzers {
		if v.Analyzer.Name != "" && (v.Analyzer.Name[:2] == "SA" || v.Analyzer.Name[:2] == "ST") {
			checks = append(checks, v.Analyzer)
		}
	}

	checks = append(checks, nilerr.Analyzer)
	checks = append(checks, staticlint.Analyzer)

	multichecker.Main(checks...)
}
