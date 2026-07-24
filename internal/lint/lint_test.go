package lint

import (
	"testing"

	"github.com/moveeeax/tfvars-lint/internal/schema"
	"github.com/moveeeax/tfvars-lint/internal/tfvars"
)

func load(t *testing.T, varsFile string) ([]schema.Variable, []tfvars.Value) {
	t.Helper()
	vars, err := schema.LoadModule("../../testdata/module")
	if err != nil {
		t.Fatalf("schema: %v", err)
	}
	vals, err := tfvars.LoadFile(varsFile)
	if err != nil {
		t.Fatalf("tfvars: %v", err)
	}
	return vars, vals
}

func byKind(findings []Finding) map[Kind][]Finding {
	m := map[Kind][]Finding{}
	for _, f := range findings {
		m[f.Kind] = append(m[f.Kind], f)
	}
	return m
}

func TestCheckValid(t *testing.T) {
	vars, vals := load(t, "../../testdata/valid.tfvars")
	if f := Check(vars, vals); len(f) != 0 {
		t.Fatalf("valid tfvars should have no findings, got %+v", f)
	}
}

func TestCheckBad(t *testing.T) {
	vars, vals := load(t, "../../testdata/bad.tfvars")
	findings := Check(vars, vals)
	m := byKind(findings)

	// Three type mismatches: instance_count, enabled, tags.
	if len(m[TypeMismatch]) != 3 {
		t.Errorf("want 3 type_mismatch, got %d: %+v", len(m[TypeMismatch]), m[TypeMismatch])
	}
	// Two unknowns: subnetss, unknown_key.
	if len(m[UnknownVariable]) != 2 {
		t.Errorf("want 2 unknown_variable, got %d: %+v", len(m[UnknownVariable]), m[UnknownVariable])
	}
	// name is required and unset.
	if len(m[MissingRequired]) != 1 || m[MissingRequired][0].Variable != "name" {
		t.Errorf("want missing_required name, got %+v", m[MissingRequired])
	}

	// subnetss should suggest subnets.
	var suggested bool
	for _, f := range m[UnknownVariable] {
		if f.Variable == "subnetss" && f.Suggestion == "subnets" {
			suggested = true
		}
	}
	if !suggested {
		t.Errorf("expected typo suggestion subnetss -> subnets, got %+v", m[UnknownVariable])
	}
}

func TestCheckComplexObjectAccepted(t *testing.T) {
	vars, _ := load(t, "../../testdata/valid.tfvars")
	vals, err := tfvars.Parse("t.tfvars", []byte(`name = "x"
settings = {
  tier     = "a"
  replicas = 1
}`))
	if err != nil {
		t.Fatal(err)
	}
	if f := Check(vars, vals); len(f) != 0 {
		t.Errorf("well-typed object should pass, got %+v", f)
	}
}

func TestCheckObjectMissingAttrIsMismatch(t *testing.T) {
	vars, _ := load(t, "../../testdata/valid.tfvars")
	vals, err := tfvars.Parse("t.tfvars", []byte(`name = "x"
settings = { tier = "a" }`))
	if err != nil {
		t.Fatal(err)
	}
	f := Check(vars, vals)
	if len(f) != 1 || f[0].Kind != TypeMismatch || f[0].Variable != "settings" {
		t.Errorf("object missing required attr should be a type_mismatch, got %+v", f)
	}
}

func TestNearest(t *testing.T) {
	cands := []string{"instance_count", "subnets", "tags"}
	if got := nearest("subnetss", cands); got != "subnets" {
		t.Errorf("nearest(subnetss) = %q, want subnets", got)
	}
	if got := nearest("zzzzzzzz", cands); got != "" {
		t.Errorf("nearest(zzzzzzzz) = %q, want empty", got)
	}
}
