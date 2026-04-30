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
  docs/                     # Directory for compiled docs. Not to be edited manually it is generated using tfplugindocs
  docs-templates/           # Directory where you define what docs for resources and datasources will contain
  docs-examples/            # Examples to be used in docs are defined here. These are referenced by the docs-templates
```

### Per-resource file pattern

Each resource in `internal/resource/<name>/` contains:

| File               | Purpose                                                                                   |
| ------------------ | ----------------------------------------------------------------------------------------- |
| `resource.go`      | CRUD operations, `Configure()`, `ImportState()`                                           |
| `schema.go`        | Schema definitions (`schema.SingleNestedAttribute`, `schema.ListNestedAttribute`, etc.)   |
| `model.go`         | Go structs with `tfsdk:` tags + `AttributeTypes()` methods + expander/flattener functions |
| `resource_test.go` | Acceptance tests                                                                          |

### Creating a Resource

- Create a struct for a resource with `client sgclient.Client` and `org_name string`
- Then create functions `NewResource` where you return the struct created in the previous resource
- Create functions Metadata, Configure and ImportState
- Then create the schema.go and define the schema for the resource based on a struct in the sdk. Ask for which struct to use.
- Then create model.go where you will create a struct with terraform `types.*` as per the schema. This file will also contain conversions from SDK Go types to terraform `types.*` and vice versa. Use `Type conversion conventions` in this file as a guide.
- Then define the CRUD methods Create, Update, Read and Delete.
- Create function will first read the plan and convert the plan to api model, use the response from create to make the read call for the resource to build the terraform `types.*`. Similarly will be the case for Update.
- Read and Delete are quite straight forward.

### Type conversion conventions

- **Expanders** (`expanders/` + inside `model.go`): convert Terraform `types.*` → SDK Go types
  - Named `convertXxxToAPI`
- **Flatteners** (`flatteners/` + inside `model.go`): convert SDK Go types → Terraform `types.*`
  - Named `convertXxxFromAPI`
- Use `flatteners.String()`, `flatteners.StringPtr()`, `flatteners.BoolPtr()`, `flatteners.Int64Ptr()` for scalar conversions
- Use `flatteners.ListOfStringToTerraformList()` for `[]string → types.List`
- Use `expanders.MapStringString()` for `types.Map → map[string]string`
- Use `sgsdkgo.Optional(value)` and `sgsdkgo.Null[T]()` for optional SDK fields
- There is a Terraform model struct for each resource. Each attribute in that resource if it is a complex type has it's own struct. That struct should have a ToAPIModel method to it and it should be used while converting from terraform `types.*` to SDK Go types. For simple types we use expanders
- For the converting complex structs from SDK Go types to terraform `types.*` create a function with naming convention like `policiesConfigToTerraType` which accepts the SDK Go types and returns terraform `types.*` or terraform model in case it is a list where the SDK Go type needs to iterated and each value from it needs to be converted to terraform model before converting the list to terraform `types.*`.

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

| Resource                                   | `internal/resource/` path     |
| ------------------------------------------ | ----------------------------- |
| `stackguardian_connector`                  | `connector/`                  |
| `stackguardian_workflow_group`             | `workflow_group/`             |
| `stackguardian_role`                       | `role/`                       |
| `stackguardian_role_v4`                    | `role_v4/`                    |
| `stackguardian_role_assignment`            | `role_assignment/`            |
| `stackguardian_policy`                     | `policy/`                     |
| `stackguardian_runner_group`               | `runner_group/`               |
| `stackguardian_workflow_template`          | `workflow_template/`          |
| `stackguardian_workflow_template_revision` | `workflow_template_revision/` |

---

## SDK Location

- During development the SDK is replaced with a local path in `go.mod`: `/Users/taherkathanawala/workspace/sg-sdk-go.git/main`
- Strictly while development use the sdk on this path and not the actual sdk being used.
