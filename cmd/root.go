// Package cmd wires the tfvars-lint CLI.
package cmd

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/moveeeax/tfvars-lint/internal/lint"
	"github.com/moveeeax/tfvars-lint/internal/report"
	"github.com/moveeeax/tfvars-lint/internal/schema"
	"github.com/moveeeax/tfvars-lint/internal/tfvars"
)

// Exit codes: 0 = clean, 1 = lint findings, 2 = usage/internal error.
const (
	exitOK       = 0
	exitFindings = 1
	exitError    = 2
)

// Run executes the CLI with the given args and streams, returning the process
// exit code. Keeping exit-code logic here (instead of cobra's Execute) makes CI
// gating explicit and testable.
func Run(args []string, stdout, stderr io.Writer) (int, error) {
	var (
		varsFile string
		module   string
		asJSON   bool
	)
	code := exitOK

	root := &cobra.Command{
		Use:           "tfvars-lint",
		Short:         "Validate .tfvars files against a module's variable schema",
		Long:          "tfvars-lint checks a .tfvars file against the variables declared in a\nTerraform module: type conformance, missing required variables, and unknown keys.",
		Example:       "  tfvars-lint --vars prod.tfvars --module ./\n  tfvars-lint --vars prod.tfvars --module ./ --json",
		SilenceUsage:  true,
		SilenceErrors: true,
		// Everything is a flag; a stray positional argument is a typo we must
		// not silently ignore (e.g. `tfvars-lint --vars a.tfvars b.tfvars`
		// would otherwise lint only a.tfvars and exit 0).
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			vars, err := schema.LoadModule(module)
			if err != nil {
				return err
			}
			vals, err := tfvars.LoadFile(varsFile)
			if err != nil {
				return err
			}
			findings := lint.Check(vars, vals)
			if len(findings) > 0 {
				code = exitFindings
			}
			if asJSON {
				return report.JSON(cmd.OutOrStdout(), varsFile, module, findings)
			}
			report.Pretty(cmd.OutOrStdout(), varsFile, findings)
			return nil
		},
	}

	root.SetOut(stdout)
	root.SetErr(stderr)
	root.SetArgs(args)
	root.Flags().StringVar(&varsFile, "vars", "", "path to the .tfvars file (required)")
	root.Flags().StringVar(&module, "module", ".", "path to the Terraform module directory")
	root.Flags().BoolVar(&asJSON, "json", false, "emit machine-readable JSON")
	_ = root.MarkFlagRequired("vars")

	if err := root.Execute(); err != nil {
		return exitError, fmt.Errorf("%w", err)
	}
	return code, nil
}
