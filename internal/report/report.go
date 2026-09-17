// Package report renders lint findings as human or machine output.
package report

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/moveeeax/tfvars-lint/internal/lint"
)

// Result is the machine-readable envelope emitted in JSON mode.
type Result struct {
	VarsFile string         `json:"vars_file"`
	Module   string         `json:"module"`
	OK       bool           `json:"ok"`
	Findings []lint.Finding `json:"findings"`
}

// JSON writes findings as a stable JSON document.
func JSON(w io.Writer, varsFile, module string, findings []lint.Finding) error {
	if findings == nil {
		findings = []lint.Finding{}
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(Result{
		VarsFile: varsFile,
		Module:   module,
		OK:       len(findings) == 0,
		Findings: findings,
	})
}

// Pretty writes a human-readable summary and returns nothing; callers decide the
// exit code from len(findings).
func Pretty(w io.Writer, varsFile string, findings []lint.Finding) {
	if len(findings) == 0 {
		fmt.Fprintf(w, "✓ %s: no issues\n", varsFile)
		return
	}
	fmt.Fprintf(w, "✗ %s: %d issue(s)\n", varsFile, len(findings))
	for _, f := range findings {
		loc := ""
		if f.Line > 0 {
			loc = fmt.Sprintf(":%d", f.Line)
		}
		fmt.Fprintf(w, "  [%s]%s %s\n", f.Kind, loc, f.Message)
	}
}
