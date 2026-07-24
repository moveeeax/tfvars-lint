# tfvars-lint

[![ci](https://github.com/moveeeax/tfvars-lint/actions/workflows/ci.yml/badge.svg)](https://github.com/moveeeax/tfvars-lint/actions/workflows/ci.yml)
[![go](https://img.shields.io/badge/go-1.22%2B-00ADD8)](https://go.dev)
[![license](https://img.shields.io/badge/license-MIT-blue)](LICENSE)

> Catch bad `.tfvars` before `terraform plan` even runs.

`tfvars-lint` reads the variable declarations in a Terraform module and checks a
`.tfvars` file against them — **before** you burn a plan/apply cycle finding out
a value has the wrong type or a key is misspelled.

```console
$ tfvars-lint --module ./ --vars prod.tfvars
✗ prod.tfvars: 3 issue(s)
  [type_mismatch]:1 expected number, got string
  [unknown_variable]:4 variable "subnetss" is not declared by the module; did you mean "subnets"?
  [missing_required] required variable "name" is not set
```

## What it checks

- **Type conformance** — every value is converted against the declared type
  (`string`, `number`, `bool`, `list`/`set`/`map`/`object`/`tuple`, and nested
  combinations) using the same `cty` engine Terraform uses.
- **Missing required variables** — declared with no `default` but absent from the
  `.tfvars`.
- **Unknown keys** — set in the `.tfvars` but not declared by the module, with a
  nearest-name suggestion for typos.

Variables typed `any` (or with no `type`) accept any value, matching Terraform.

## Install

```bash
go install github.com/moveeeax/tfvars-lint@latest
```

Produces a single static binary (`tfvars-lint`).

## Usage

```
tfvars-lint --vars <file.tfvars> [--module <dir>] [--json]
```

| Flag | Default | Description |
|------|---------|-------------|
| `--vars` | — | Path to the `.tfvars` file (required) |
| `--module` | `.` | Terraform module directory to read `variables.tf` from |
| `--json` | `false` | Emit machine-readable JSON |

### Exit codes

| Code | Meaning |
|------|---------|
| `0` | No issues |
| `1` | Lint findings |
| `2` | Usage or internal error |

The `0/1` split makes it a drop-in CI gate.

### JSON output

```json
{
  "vars_file": "prod.tfvars",
  "module": "./",
  "ok": false,
  "findings": [
    { "kind": "type_mismatch", "variable": "instance_count", "message": "expected number, got string", "line": 1 }
  ]
}
```

## GitHub Action

This repo ships a composite action ([`action.yml`](action.yml)):

```yaml
- uses: moveeeax/tfvars-lint@v0
  with:
    module: ./infra
    vars: ./infra/prod.tfvars
    json: "true"
```

## How it works

1. Every `*.tf` in the module dir is parsed with `hashicorp/hcl/v2`; `variable`
   blocks yield a name → type-constraint schema (via `hcl/ext/typeexpr`).
2. The `.tfvars` file is parsed to `cty` values.
3. Each value is `convert.Convert`-ed against its declared type; failures,
   unknown keys, and unset required variables become findings.

## Development

```bash
make build   # go build -o tfvars-lint .
make test    # go test -race ./...
make lint    # gofmt -l . && go vet ./...
```

## License

MIT — see [LICENSE](LICENSE).
