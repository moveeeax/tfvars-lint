package schema

import (
	"testing"

	"github.com/zclconf/go-cty/cty"
)

func TestLoadModule(t *testing.T) {
	vars, err := LoadModule("../../testdata/module")
	if err != nil {
		t.Fatalf("LoadModule: %v", err)
	}
	byName := map[string]Variable{}
	for _, v := range vars {
		byName[v.Name] = v
	}

	name, ok := byName["name"]
	if !ok {
		t.Fatal("missing variable 'name'")
	}
	if name.Type != cty.String {
		t.Errorf("name type = %s, want string", name.TypeString)
	}
	if !name.Required() {
		t.Error("name should be required (no default)")
	}

	if ic := byName["instance_count"]; ic.Required() || ic.Type != cty.Number {
		t.Errorf("instance_count wrong: required=%v type=%s", ic.Required(), ic.TypeString)
	}

	tags := byName["tags"]
	if !tags.Type.IsMapType() {
		t.Errorf("tags should be a map, got %s", tags.TypeString)
	}

	settings := byName["settings"]
	if !settings.Type.IsObjectType() {
		t.Errorf("settings should be object, got %s", settings.TypeString)
	}

	if any := byName["anything"]; any.Type != cty.DynamicPseudoType {
		t.Errorf("anything should be dynamic/any, got %s", any.TypeString)
	}
}

func TestLoadModuleMissingDir(t *testing.T) {
	if _, err := LoadModule("../../testdata/does-not-exist"); err == nil {
		t.Fatal("expected error for missing dir")
	}
}

func TestParseInvalidHCL(t *testing.T) {
	_, diags := parse("bad.tf", []byte(`variable "x" { type = `))
	if !diags.HasErrors() {
		t.Fatal("expected parse diagnostics for malformed HCL")
	}
}

// Regression: typeexpr.TypeConstraint rejects the two-argument
// optional(<type>, <default>) form that Terraform 1.3+ accepts, which made
// LoadModule fail outright on any module using it.
func TestLoadModuleOptionalObjectAttributes(t *testing.T) {
	vars, err := LoadModule("../../testdata/module")
	if err != nil {
		t.Fatalf("LoadModule: %v", err)
	}
	var scaling *Variable
	for i := range vars {
		if vars[i].Name == "scaling" {
			scaling = &vars[i]
		}
	}
	if scaling == nil {
		t.Fatal("missing variable 'scaling'")
	}
	if !scaling.Type.IsObjectType() {
		t.Fatalf("scaling should be an object, got %s", scaling.TypeString)
	}
	for _, attr := range []string{"max", "strategy"} {
		if !scaling.Type.AttributeOptional(attr) {
			t.Errorf("attribute %q should be optional, type is %s", attr, scaling.TypeString)
		}
	}
	if scaling.Type.AttributeOptional("min") {
		t.Errorf("attribute \"min\" should be required, type is %s", scaling.TypeString)
	}
}

func TestLoadModuleNullable(t *testing.T) {
	vars, err := LoadModule("../../testdata/strict")
	if err != nil {
		t.Fatalf("LoadModule: %v", err)
	}
	byName := map[string]Variable{}
	for _, v := range vars {
		byName[v.Name] = v
	}
	if v := byName["account_id"]; v.Nullable || !v.Required() {
		t.Errorf("account_id: nullable=%v required=%v, want false/true", v.Nullable, v.Required())
	}
	if v := byName["region"]; v.Nullable || v.Required() {
		t.Errorf("region: nullable=%v required=%v, want false/false", v.Nullable, v.Required())
	}
	// A variable with no nullable argument defaults to nullable, per Terraform.
	mod, err := LoadModule("../../testdata/module")
	if err != nil {
		t.Fatalf("LoadModule: %v", err)
	}
	for _, v := range mod {
		if v.Name == "name" && !v.Nullable {
			t.Error("name should default to nullable = true")
		}
	}
}
