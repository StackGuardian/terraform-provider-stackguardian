# CLAUDE.md — terraform-provider-stackguardian

## Project Overview

Terraform provider for the StackGuardian Orchestrator platform. Built with the [Terraform Plugin Framework v1](https://developer.hashicorp.com/terraform/plugin/framework).

**Module:** `github.com/StackGuardian/terraform-provider-stackguardian`
**Go version:** 1.21.4
**SDK:** `github.com/StackGuardian/sg-sdk-go` look for replace if in go.mod and use that path for the sdk

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

```bash
TF_ACC=1
STACKGUARDIAN_API_KEY=<key>
STACKGUARDIAN_API_URI=<uri>
STACKGUARDIAN_ORG_NAME=<org>
```

---

## Architecture

### Directory Layout

```text
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

| File               | Purpose                                                                                   |
| ------------------ | ----------------------------------------------------------------------------------------- |
| `resource.go`      | CRUD operations, `Configure()`, `ImportState()`, `Metadata`                               |
| `schema.go`        | Schema definitions (`schema.SingleNestedAttribute`, `schema.ListNestedAttribute`, etc.)   |
| `model.go`         | Go structs with `tfsdk:` tags + `AttributeTypes()` methods + expander/flattener functions |
| `resource_test.go` | Acceptance tests                                                                          |

### Per-datasource file pattern

| File            | Purpose                                                                               |
| --------------- | ------------------------------------------------------------------------------------- |
| `datasource.go` | Read operation, `Configure()`, `Metadata()`, `NewDataSource()` constructor            |
| `schema.go`     | Schema definitions (read-only, typically mirrors the resource schema with `Computed`) |

Datasources do not have their own `model.go` — they import and reuse the model from the corresponding resource package (e.g., `workflowresource "…/internal/resource/workflow"`).

### Type conversion conventions

These are required while writing the implementing or making changes in model.go

- **Expanders** (`expanders/` + inside `model.go`): convert Terraform `types.*` → SDK Go types
  - Simple/leaf conversions use package-level helpers in `expanders/`
- **Flatteners** (`flatteners/` + inside `model.go`): convert SDK Go types → Terraform `types.*`
  - Simple/leaf conversions use package-level helpers in `flatteners/`
- While editing model.go or applying conversions check the expanders and flatteners for compatibility.
- **`ToAPIModel` (Terraform → SDK):** Every model struct that has a complex type (i.e. not a plain scalar) must implement a `ToAPIModel` method. This is the canonical way to convert from Terraform `types.*` to SDK Go types for nested objects and lists — do not use standalone `convertXxxToAPI` functions for this. Signature conventions:
  - Leaf structs with no nested `types.Object`/`types.List` fields: `func (m XxxModel) ToAPIModel() SdkType`
  - Structs that unpack nested objects/lists from Terraform state: `func (m XxxModel) ToAPIModel(ctx context.Context) (SdkType, diag.Diagnostics)`
  - Package-level `convertXxxToAPI` helpers are thin wrappers that handle the null/unknown guard and delegate to `m.ToAPIModel(ctx)`:

    ```go
    func convertXxxToAPI(ctx context.Context, obj types.Object) (*SdkType, diag.Diagnostics) {
        if obj.IsNull() || obj.IsUnknown() { return nil, nil }
        var m XxxModel
        diags := obj.As(ctx, &m, basetypes.ObjectAsOptions{})
        if diags.HasError() { return nil, diags }
        return m.ToAPIModel(ctx)
    }
    ```

- **`convertXxxFromAPI` (SDK → Terraform):** Package-level functions named `convert<Xxx>FromAPI` that accept the SDK Go type and return `types.Object`, `types.List`, or a model struct. For lists, iterate the SDK slice and convert each element before collecting into a `types.List`.

### Model struct rules

- Every model struct if it is types.Object must implement `AttributeTypes() map[string]attr.Type`
- `AttributeTypes()` must exactly match the schema — type mismatches cause runtime panics
- Schema `SingleNestedAttribute` → `types.Object` field
- Schema `ListNestedAttribute` → `types.List` field with `ElemType: types.ObjectType{...}`
- Schema `BoolAttribute` → `types.BoolType` in `AttributeTypes()` (not `StringType`)
- If a `schema.StringAttribute` description says the value is a JSON string, marshal it in `ToAPIModel`/`ToUpdateAPIModel` (Go value → `json.Marshal` → string) and unmarshal it in `convertXxxFromAPI` (string → `json.Unmarshal` → Go value). Treat a marshal/unmarshal error as a diagnostic error and return early.
- For any attribute that is both `Optional` and `Computed`, guard against null/unknown in `ToAPIModel` before setting its value. `ValueBoolPointer()` returns `&false` for unknown and `ValueStringPointer()` returns `&""` for unknown — both send unintended values to the API. Use: `if !m.Field.IsNull() && !m.Field.IsUnknown() { cfg.Field = m.Field.ValueBoolPointer() }`

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

Define `nullObj` / `nullList` at the top of every `convertXxxFromAPI` function and return it on every error path — never return the zero value (`types.Object{}` / `types.List{}`). Check diags after each step and return early with the null value if there is an error. This prevents cascading: if an inner converter returns a zero value and the parent passes it to `types.ObjectValueFrom` or `types.ListValueFrom`, the framework produces a second wave of diags that obscures the original error. Returning the null value stops that chain.

**Object-returning converters** — check nil and `IsEmptyObject` at the top:

```go
func convertXxxFromAPI(ctx context.Context, input *SomeType) (types.Object, diag.Diagnostics) {
    nullObj := types.ObjectNull(XxxModel{}.AttributeTypes())

    if input == nil || flatteners.IsEmptyObject(input) {
        return nullObj, nil
    }

    nested, diags := convertNestedFromAPI(ctx, input.Nested)
    if diags.HasError() {
        return nullObj, diags
    }

    obj, diags := types.ObjectValueFrom(ctx, XxxModel{}.AttributeTypes(), XxxModel{
        Field:  flatteners.String(input.Field),
        Nested: nested,
    })
    if diags.HasError() {
        return nullObj, diags
    }

    return obj, diags
}
```

`flatteners.IsEmptyObject` marshals the struct to JSON and returns `true` if the result is `{}`. This requires every JSON-tagged field on the SDK struct to use `omitempty` on a pointer, slice, or map type — bare value fields (e.g. `Type string \`json:"type"\``) will not be omitted and must be changed to pointers in the SDK repo.

**List-returning converters** — check `len == 0` at the top (covers both nil and empty slice):

```go
func convertXxxListFromAPI(ctx context.Context, items []SdkType) (types.List, diag.Diagnostics) {
    nullList := types.ListNull(types.ObjectType{AttrTypes: XxxModel{}.AttributeTypes()})

    if len(items) == 0 {
        return nullList, nil
    }

    models := make([]XxxModel, 0, len(items))
    for _, item := range items {
        if flatteners.IsEmptyObject(item) {
            continue
        }
        models = append(models, XxxModel{ /* ... */ })
    }
    if len(models) == 0 {
        return nullList, nil
    }

    list, diags := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: XxxModel{}.AttributeTypes()}, models)
    if diags.HasError() {
        return nullList, diags
    }
    return list, nil
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

### Reusable components in schemas

If some attributes occur many times in the schema create functions for the
schemas of those attributes and re-use them.

---

## Common Gotchas

1. **`AttributeTypes()` must match schema exactly.** Runtime errors like `Expected framework type … / Received framework type` always mean a mismatch between the schema definition and the corresponding model's `AttributeTypes()`. Fix the side that is wrong — usually the model's `AttributeTypes()` return value.

2. **`SingleNestedAttribute` vs `ListNestedAttribute`:** If the API returns a list, use `schema.ListNestedAttribute` + `types.List` in the model. If it returns a single object, use `schema.SingleNestedAttribute` + `types.Object`.

3. **SDK value slices:** `WfStepsConfig`, `EnvVars`, `InputSchemas`, etc. are value slices (`[]T`), not pointer slices (`[]*T`).

4. **`UpdateRequest` vs `CreateRequest`:** Some fields present in Create are absent in Update (and vice versa). Always check the SDK struct for the specific request type.

5. **`go build` is the source of truth** for compilation errors. IDE diagnostics can be stale.

6. **Diagnostics convention in `model.go`:** Check diags after every step in all converter functions (`ToAPIModel`, `ToUpdateAPIModel`, `convertXxxFromAPI`) and return early if there is an error. For `convertXxxFromAPI`, return `nullObj` / `nullList` (never the zero value) on every error path — see the null guards pattern above.

7. **`Computed`-only fields inside `ListNestedAttribute` need `UseStateForUnknown()`.**  For top-level `Computed` attributes Terraform carries the state value forward automatically. But inside a `ListNestedAttribute`, Terraform plans the entire list element object fresh each time — any `Computed` field with no config value is marked `(known after apply)` even when state already has a value. This produces a perpetual diff after the first apply. Fix: add `stringplanmodifier.UseStateForUnknown()` (or the equivalent for other types) to every `Computed`-only field inside a nested list object.

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
| `stackguardian_workflow_git`               | `workflow_git/`               |

---
