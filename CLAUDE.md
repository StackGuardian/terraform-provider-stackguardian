# CLAUDE.md — terraform-provider-stackguardian

## Project Overview

Terraform provider for the StackGuardian Orchestrator platform. Built with the [Terraform Plugin Framework v1](https://developer.hashicorp.com/terraform/plugin/framework).

**Module:** `github.com/StackGuardian/terraform-provider-stackguardian`
**Go version:** 1.21.4
**SDK:** `github.com/StackGuardian/sg-sdk-go` (currently on `feat-workflow-templates` feature branch)

---

## Common Commands

```bash
# Build
make build          # compile provider binary
make install        # build + install to ~/.terraform.d/plugins

# Tests
make test           # unit tests (4 parallel, 30s timeout)
make test-acc       # acceptance tests (requires env vars, 1 parallel, 15m timeout)

# Documentation
make docs-generate  # run tfplugindocs generate
make docs-validate  # validate generated docs

# Clean
make clean
```

### Required environment variables for acceptance tests

```
TF_ACC=1
STACKGUARDIAN_API_KEY=<key>
STACKGUARDIAN_API_URI=<uri>
STACKGUARDIAN_ORG_NAME=<org>
```

---

## Architecture

### Directory Layout

```
internal/
  acctest/                  # Shared acceptance test helpers
  constants/                # Shared documentation strings and enums
  customTypes/              # ProviderInfo struct (API client + credentials)
  datasources/<name>/       # Data source implementations
  expanders/                # Terraform types → SDK types
  flatteners/               # SDK types → Terraform types
  provider/                 # Provider registration
  resource/<name>/          # Resource implementations
```

### Per-resource file pattern

Each resource in `internal/resource/<name>/` contains:

| File | Purpose |
|------|---------|
| `resource.go` | CRUD operations, `Configure()`, `ImportState()` |
| `schema.go` | Schema definitions (`schema.SingleNestedAttribute`, `schema.ListNestedAttribute`, etc.) |
| `model.go` | Go structs with `tfsdk:` tags + `AttributeTypes()` methods + expander/flattener functions |
| `resource_test.go` | Acceptance tests |

### Type conversion conventions

- **Expanders** (`expanders/` + inside `model.go`): convert Terraform `types.*` → SDK Go types
  - Named `convertXxxToAPI`
- **Flatteners** (`flatteners/` + inside `model.go`): convert SDK Go types → Terraform `types.*`
  - Named `convertXxxFromAPI`
- Use `flatteners.String()`, `flatteners.StringPtr()`, `flatteners.BoolPtr()`, `flatteners.Int64Ptr()` for scalar conversions
- Use `flatteners.ListOfStringToTerraformList()` for `[]string → types.List`
- Use `expanders.MapStringString()` for `types.Map → map[string]string`
- Use `sgsdkgo.Optional(value)` and `sgsdkgo.Null[T]()` for optional SDK fields

### Model struct rules

- Every model struct must implement `AttributeTypes() map[string]attr.Type`
- `AttributeTypes()` must exactly match the schema — type mismatches cause runtime panics
- Schema `SingleNestedAttribute` → `types.Object` field
- Schema `ListNestedAttribute` → `types.List` field with `ElemType: types.ObjectType{...}`
- Schema `BoolAttribute` → `types.BoolType` in `AttributeTypes()` (not `StringType`)

---

## Key Patterns

### Handling optional SDK fields (UpdateRequest)

```go
if value != nil {
    apiModel.Field = sgsdkgo.Optional(value)
} else {
    apiModel.Field = sgsdkgo.Null[FieldType]()
}
```

### Null guards in fromAPI converters

```go
func convertXxxFromAPI(ctx context.Context, input *SomeType) (types.Object, diag.Diagnostics) {
    nullObj := types.ObjectNull(XxxModel{}.AttributeTypes())
    if input == nil {
        return nullObj, nil
    }
    // ... convert fields
}
```

### Loop variable addresses (safe in Go 1.22+)

The SDK uses value slices (`[]T`, not `[]*T`). Taking `&item` inside a range loop is safe because converters use the pointer synchronously:

```go
for i, item := range items {
    obj, diags := convertItemFromAPI(ctx, &item)
    // ...
}
```

---

## Common Gotchas

1. **`AttributeTypes()` must match schema exactly.** Runtime errors like `Expected framework type … / Received framework type` always mean a mismatch between the schema definition and the corresponding model's `AttributeTypes()`. Fix the side that is wrong — usually the model's `AttributeTypes()` return value.

2. **`SingleNestedAttribute` vs `ListNestedAttribute`:** If the API returns a list, use `schema.ListNestedAttribute` + `types.List` in the model. If it returns a single object, use `schema.SingleNestedAttribute` + `types.Object`.

3. **SDK value slices:** `WfStepsConfig`, `EnvVars`, `InputSchemas`, etc. are value slices (`[]T`), not pointer slices (`[]*T`).

4. **`UpdateRequest` vs `CreateRequest`:** Some fields present in Create are absent in Update (and vice versa). Always check the SDK struct for the specific request type.

5. **`go build` is the source of truth** for compilation errors. IDE diagnostics can be stale.

---

## Resources

| Resource | `internal/resource/` path |
|----------|--------------------------|
| `stackguardian_connector` | `connector/` |
| `stackguardian_workflow_group` | `workflow_group/` |
| `stackguardian_role` | `role/` |
| `stackguardian_role_v4` | `role_v4/` |
| `stackguardian_role_assignment` | `role_assignment/` |
| `stackguardian_policy` | `policy/` |
| `stackguardian_runner_group` | `runner_group/` |
| `stackguardian_workflow_template` | `workflow_template/` |
| `stackguardian_workflow_template_revision` | `workflow_template_revision/` |

---

## SDK Location

During development the SDK is replaced with a local path in `go.mod`:

```
replace github.com/StackGuardian/sg-sdk-go => /path/to/sg-sdk-go.git/feat-workflow-templates
```

Key SDK packages:
- `github.com/StackGuardian/sg-sdk-go/sgsdkgo` — shared types (`EnvVars`, `WfStepsConfig`, `MountPoint`, `Optional`, `Null`, etc.)
- `github.com/StackGuardian/sg-sdk-go/workflowtemplaterevisions` — revision-specific types and request structs
