// Package lint validates parsed .tfvars values against a module schema.
package lint

import (
	"fmt"
	"sort"

	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/convert"

	"github.com/moveeeax/tfvars-lint/internal/schema"
	"github.com/moveeeax/tfvars-lint/internal/tfvars"
)

// Kind classifies a finding.
type Kind string

const (
	TypeMismatch    Kind = "type_mismatch"
	MissingRequired Kind = "missing_required"
	UnknownVariable Kind = "unknown_variable"
	NullNotAllowed  Kind = "null_not_allowed"
)

// Finding is a single lint problem.
type Finding struct {
	Kind       Kind   `json:"kind"`
	Variable   string `json:"variable"`
	Message    string `json:"message"`
	Suggestion string `json:"suggestion,omitempty"`
	Line       int    `json:"line,omitempty"`
}

// Check compares the supplied values against the schema and returns findings,
// ordered: unknown/type problems by source line first, then missing-required by
// name. An empty result means the file is valid.
func Check(vars []schema.Variable, vals []tfvars.Value) []Finding {
	byName := make(map[string]schema.Variable, len(vars))
	names := make([]string, 0, len(vars))
	for _, v := range vars {
		byName[v.Name] = v
		names = append(names, v.Name)
	}

	var findings []Finding
	seen := map[string]bool{}
	for _, val := range vals {
		seen[val.Name] = true
		v, ok := byName[val.Name]
		if !ok {
			f := Finding{
				Kind:     UnknownVariable,
				Variable: val.Name,
				Message:  fmt.Sprintf("variable %q is not declared by the module", val.Name),
				Line:     val.Range.Start.Line,
			}
			if s := nearest(val.Name, names); s != "" {
				f.Suggestion = s
				f.Message += fmt.Sprintf("; did you mean %q?", s)
			}
			findings = append(findings, f)
			continue
		}
		// Terraform rejects an explicit null only for a nullable = false
		// variable that has no default; when a default exists it silently
		// substitutes the default instead, so that case is not a finding.
		if !v.Nullable && !v.HasDefault && val.Val.IsNull() {
			findings = append(findings, Finding{
				Kind:     NullNotAllowed,
				Variable: val.Name,
				Message: fmt.Sprintf("required variable %q is declared nullable = false and may not be set to null",
					val.Name),
				Line: val.Range.Start.Line,
			})
			continue
		}
		if v.Type == cty.DynamicPseudoType || v.Type == cty.NilType {
			continue // "any" accepts everything
		}
		if _, err := convert.Convert(val.Val, v.Type); err != nil {
			findings = append(findings, Finding{
				Kind:     TypeMismatch,
				Variable: val.Name,
				Message: fmt.Sprintf("expected %s, got %s",
					v.TypeString, val.Val.Type().FriendlyName()),
				Line: val.Range.Start.Line,
			})
		}
	}

	// Missing required variables (declared, no default, not set).
	var missing []Finding
	for _, v := range vars {
		if v.Required() && !seen[v.Name] {
			missing = append(missing, Finding{
				Kind:     MissingRequired,
				Variable: v.Name,
				Message:  fmt.Sprintf("required variable %q is not set", v.Name),
			})
		}
	}

	sort.SliceStable(findings, func(i, j int) bool { return findings[i].Line < findings[j].Line })
	sort.SliceStable(missing, func(i, j int) bool { return missing[i].Variable < missing[j].Variable })
	return append(findings, missing...)
}

// nearest returns the closest candidate name within a small edit distance, for
// typo suggestions; "" when nothing is close enough.
func nearest(target string, candidates []string) string {
	best, bestDist := "", 1<<30
	for _, c := range candidates {
		d := levenshtein(target, c)
		if d < bestDist {
			best, bestDist = c, d
		}
	}
	// Only suggest when the edit distance is a small fraction of the length.
	limit := (len(target) + 3) / 3
	if limit < 1 {
		limit = 1
	}
	if bestDist <= limit {
		return best
	}
	return ""
}

func levenshtein(a, b string) int {
	ra, rb := []rune(a), []rune(b)
	prev := make([]int, len(rb)+1)
	for j := range prev {
		prev[j] = j
	}
	for i := 1; i <= len(ra); i++ {
		cur := make([]int, len(rb)+1)
		cur[0] = i
		for j := 1; j <= len(rb); j++ {
			cost := 1
			if ra[i-1] == rb[j-1] {
				cost = 0
			}
			cur[j] = min3(prev[j]+1, cur[j-1]+1, prev[j-1]+cost)
		}
		prev = cur
	}
	return prev[len(rb)]
}

func min3(a, b, c int) int {
	m := a
	if b < m {
		m = b
	}
	if c < m {
		m = c
	}
	return m
}
