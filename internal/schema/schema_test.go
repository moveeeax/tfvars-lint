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
