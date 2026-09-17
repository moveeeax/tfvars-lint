# examples

A small module (`module/variables.tf`) and a matching `prod.tfvars`.

```bash
# passes
tfvars-lint --module ./module --vars ./prod.tfvars

# JSON for CI
tfvars-lint --module ./module --vars ./prod.tfvars --json
```

Introduce a mistake to see it caught:

```bash
echo 'instance_count = "three"' >> prod.tfvars
tfvars-lint --module ./module --vars ./prod.tfvars
# ✗ prod.tfvars: 1 issue(s)
#   [type_mismatch]:N expected number, got string
```
