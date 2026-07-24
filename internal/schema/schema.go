// Package schema reads a Terraform module's variable declarations into a typed
// schema used to validate .tfvars files.
package schema

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/ext/typeexpr"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

// Variable is one declared input variable.
type Variable struct {
	Name       string
	Type       cty.Type
	TypeString string
	HasDefault bool
	Nullable   bool
	Sensitive  bool
	DeclRange  hcl.Range
}

// Required reports whether a value must be supplied (no default declared).
func (v Variable) Required() bool { return !v.HasDefault }

// LoadModule parses every *.tf file in dir and returns the declared variables,
// sorted by name. Files that fail to parse produce diagnostics, not a panic.
func LoadModule(dir string) ([]Variable, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read module dir: %w", err)
	}
	var files []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if strings.HasSuffix(e.Name(), ".tf") && !strings.HasSuffix(e.Name(), ".tftpl") {
			files = append(files, filepath.Join(dir, e.Name()))
		}
	}
	sort.Strings(files)

	byName := map[string]Variable{}
	var diags hcl.Diagnostics
	for _, f := range files {
		src, err := os.ReadFile(f)
		if err != nil {
			return nil, err
		}
		vars, d := parse(f, src)
		diags = append(diags, d...)
		for _, v := range vars {
			byName[v.Name] = v
		}
	}
	if diags.HasErrors() {
		return nil, fmt.Errorf("parse module: %s", diags.Error())
	}

	out := make([]Variable, 0, len(byName))
	for _, v := range byName {
		out = append(out, v)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}

// parse extracts variable blocks from a single file's source.
func parse(filename string, src []byte) ([]Variable, hcl.Diagnostics) {
	file, diags := hclsyntax.ParseConfig(src, filename, hcl.InitialPos)
	if diags.HasErrors() {
		return nil, diags
	}
	body, ok := file.Body.(*hclsyntax.Body)
	if !ok {
		return nil, diags
	}

	var vars []Variable
	for _, block := range body.Blocks {
		if block.Type != "variable" || len(block.Labels) != 1 {
			continue
		}
		v := Variable{
			Name:       block.Labels[0],
			Type:       cty.DynamicPseudoType,
			TypeString: "any",
			Nullable:   true,
			DeclRange:  block.DefRange(),
		}
		attrs := block.Body.Attributes
		if ta, ok := attrs["type"]; ok {
			ty, d := typeexpr.TypeConstraint(ta.Expr)
			diags = append(diags, d...)
			if !d.HasErrors() {
				v.Type = ty
				v.TypeString = typeexpr.TypeString(ty)
			}
		}
		if _, ok := attrs["default"]; ok {
			v.HasDefault = true
		}
		if na, ok := attrs["nullable"]; ok {
			if val, d := na.Expr.Value(nil); !d.HasErrors() && val.Type() == cty.Bool {
				v.Nullable = val.True()
			}
		}
		if sa, ok := attrs["sensitive"]; ok {
			if val, d := sa.Expr.Value(nil); !d.HasErrors() && val.Type() == cty.Bool {
				v.Sensitive = val.True()
			}
		}
		vars = append(vars, v)
	}
	return vars, diags
}
