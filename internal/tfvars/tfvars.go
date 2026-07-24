// Package tfvars parses a .tfvars file into its assigned values.
package tfvars

import (
	"fmt"
	"os"
	"sort"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

// Value is a single assignment from a .tfvars file.
type Value struct {
	Name  string
	Val   cty.Value
	Range hcl.Range
}

// LoadFile parses path and returns its assignments, ordered by position so
// output is deterministic. tfvars are literal, so expressions evaluate with no
// context.
func LoadFile(path string) ([]Value, error) {
	src, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return Parse(path, src)
}

// Parse reads assignments from tfvars source.
func Parse(filename string, src []byte) ([]Value, error) {
	file, diags := hclsyntax.ParseConfig(src, filename, hcl.InitialPos)
	if diags.HasErrors() {
		return nil, fmt.Errorf("parse %s: %s", filename, diags.Error())
	}
	attrs, diags := file.Body.JustAttributes()
	if diags.HasErrors() {
		return nil, fmt.Errorf("parse %s: %s", filename, diags.Error())
	}

	out := make([]Value, 0, len(attrs))
	for name, attr := range attrs {
		val, d := attr.Expr.Value(nil)
		if d.HasErrors() {
			return nil, fmt.Errorf("evaluate %q in %s: %s", name, filename, d.Error())
		}
		out = append(out, Value{Name: name, Val: val, Range: attr.Range})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Range.Start.Line != out[j].Range.Start.Line {
			return out[i].Range.Start.Line < out[j].Range.Start.Line
		}
		return out[i].Name < out[j].Name
	})
	return out, nil
}
