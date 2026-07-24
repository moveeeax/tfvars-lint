package cmd

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/moveeeax/tfvars-lint/internal/report"
)

func TestRunValidExit0(t *testing.T) {
	var out, errb bytes.Buffer
	code, err := Run([]string{"--module", "../testdata/module", "--vars", "../testdata/valid.tfvars"}, &out, &errb)
	if err != nil {
		t.Fatalf("Run: %v (%s)", err, errb.String())
	}
	if code != exitOK {
		t.Errorf("exit = %d, want 0; out=%s", code, out.String())
	}
	if !strings.Contains(out.String(), "no issues") {
		t.Errorf("unexpected output: %s", out.String())
	}
}

func TestRunBadExit1(t *testing.T) {
	var out, errb bytes.Buffer
	code, err := Run([]string{"--module", "../testdata/module", "--vars", "../testdata/bad.tfvars"}, &out, &errb)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if code != exitFindings {
		t.Errorf("exit = %d, want 1", code)
	}
	if !strings.Contains(out.String(), "issue(s)") {
		t.Errorf("expected issue summary, got: %s", out.String())
	}
}

func TestRunJSON(t *testing.T) {
	var out, errb bytes.Buffer
	code, err := Run([]string{"--module", "../testdata/module", "--vars", "../testdata/bad.tfvars", "--json"}, &out, &errb)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if code != exitFindings {
		t.Errorf("exit = %d, want 1", code)
	}
	var res report.Result
	if err := json.Unmarshal(out.Bytes(), &res); err != nil {
		t.Fatalf("JSON invalid: %v\n%s", err, out.String())
	}
	if res.OK {
		t.Error("OK should be false for bad tfvars")
	}
	if len(res.Findings) == 0 {
		t.Error("expected findings in JSON")
	}
}

func TestRunMissingVarsFlag(t *testing.T) {
	var out, errb bytes.Buffer
	code, err := Run([]string{"--module", "../testdata/module"}, &out, &errb)
	if err == nil {
		t.Fatal("expected error when --vars is absent")
	}
	if code != exitError {
		t.Errorf("exit = %d, want 2", code)
	}
}

func TestRunUnparseableModule(t *testing.T) {
	var out, errb bytes.Buffer
	code, _ := Run([]string{"--module", "/no/such/module", "--vars", "../testdata/valid.tfvars"}, &out, &errb)
	if code != exitError {
		t.Errorf("exit = %d, want 2", code)
	}
}
