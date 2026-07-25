package tfvars

import (
	"testing"

	"github.com/zclconf/go-cty/cty"
)

func TestParseOrdersByLine(t *testing.T) {
	vals, err := Parse("t.tfvars", []byte("b = 2\na = 1\n"))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(vals) != 2 || vals[0].Name != "b" || vals[1].Name != "a" {
		t.Fatalf("expected source order b,a; got %+v", vals)
	}
}

func TestParseTypes(t *testing.T) {
	vals, err := Parse("t.tfvars", []byte(`s = "x"
n = 3
b = true
l = ["a", "b"]
m = { k = "v" }`))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	got := map[string]cty.Type{}
	for _, v := range vals {
		got[v.Name] = v.Val.Type()
	}
	if got["s"] != cty.String || got["n"] != cty.Number || got["b"] != cty.Bool {
		t.Errorf("primitive types wrong: %+v", got)
	}
	if !got["l"].IsTupleType() {
		t.Errorf("l should be tuple, got %s", got["l"].FriendlyName())
	}
	if !got["m"].IsObjectType() {
		t.Errorf("m should be object, got %s", got["m"].FriendlyName())
	}
}

func TestParseInvalid(t *testing.T) {
	if _, err := Parse("t.tfvars", []byte("a = =")); err == nil {
		t.Fatal("expected parse error")
	}
}

func TestLoadFileMissing(t *testing.T) {
	if _, err := LoadFile("/no/such/file.tfvars"); err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestParseJSONSyntax(t *testing.T) {
	vals, err := Parse("t.tfvars.json", []byte(`{
  "s": "x",
  "n": 3,
  "b": true,
  "l": ["a"],
  "m": {"k": "v"}
}`))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	got := map[string]cty.Value{}
	for _, v := range vals {
		got[v.Name] = v.Val
	}
	if len(got) != 5 {
		t.Fatalf("want 5 assignments, got %d: %+v", len(got), got)
	}
	if got["s"].Type() != cty.String || got["n"].Type() != cty.Number || got["b"].Type() != cty.Bool {
		t.Errorf("primitive types wrong: %+v", got)
	}
	if !got["l"].Type().IsTupleType() {
		t.Errorf("l should be a tuple, got %s", got["l"].Type().FriendlyName())
	}
	if !got["m"].Type().IsObjectType() {
		t.Errorf("m should be an object, got %s", got["m"].Type().FriendlyName())
	}
}

func TestParseJSONInvalid(t *testing.T) {
	if _, err := Parse("t.tfvars.json", []byte(`{"a": }`)); err == nil {
		t.Fatal("expected parse error for malformed JSON")
	}
}

// A .json file must not be fed to the native-syntax parser (and vice versa).
func TestParseSyntaxSelectedBySuffix(t *testing.T) {
	if _, err := Parse("t.tfvars", []byte(`{"a": 1}`)); err == nil {
		t.Error("JSON body in a native-syntax file should not parse")
	}
	if _, err := Parse("t.TFVARS.JSON", []byte(`{"a": 1}`)); err != nil {
		t.Errorf("suffix match should be case-insensitive: %v", err)
	}
}
