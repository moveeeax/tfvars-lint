# tfvars-lint

> Catch bad `.tfvars` before `terraform plan` even runs.

**Status:** 🚧 In development

## Overview

Validate .tfvars files against a module's variable schema: types, required, unknown vars.

## Features

- Reads `variables.tf` to build the expected schema
- Checks type conformance (string/number/bool/list/map/object)
- Flags missing required variables and unknown keys
- Machine-readable (JSON) and pretty output
- Exit codes suitable for CI gating

## Stack

Go + hashicorp/hcl v2 for parsing.

## Usage

```bash
tfvars-lint --vars prod.tfvars --module ./
```

## License

MIT
