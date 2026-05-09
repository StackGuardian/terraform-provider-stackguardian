---
name: Create Resource
description: Use this skill when asked to create a new resource for the `terraform-provider-stackguardian` provider using the Terraform Plugin Framework v1.
---

## Overview

A resource lives under `internal/resource/<name>/` and consists of four files:

| File               | Purpose                                                             |
| ------------------ | ------------------------------------------------------------------- |
| `resource.go`      | Struct, `NewResource`, `Metadata`, `Configure`, `ImportState`, CRUD |
| `schema.go`        | `Schema()` method with all attribute definitions                    |
| `model.go`         | Terraform model struct + `ToAPIModel` / `convertXxxFromAPI`         |
| `resource_test.go` | Acceptance tests                                                    |

After creating the files, register `NewResource` in `internal/provider/provider.go` inside the `Resources()` function.

---

## Step-by-step

### 1. Ask for the SDK struct

Before writing any code, ask the user:
- Which SDK struct represents the **Create** request?
- Which SDK struct represents the **Update** (Patched) request?
- Which SDK client method group handles this resource (e.g. `r.client.Workflows`)?
- What is the resource name suffix (e.g. `_workflow_group`)?

Locate the SDK under the path referenced by the `replace` directive in `go.mod`.

---

### 2. `resource.go` (Methods other than CRUD)

```go
package <name>

import (
    "context"
    "fmt"

    sgclient "github.com/StackGuardian/sg-sdk-go/client"
    "github.com/StackGuardian/terraform-provider-stackguardian/internal/customTypes"
    "github.com/hashicorp/terraform-plugin-framework/path"
    "github.com/hashicorp/terraform-plugin-framework/resource"
)

var (
    _ resource.Resource                = &<name>Resource{}
    _ resource.ResourceWithConfigure   = &<name>Resource{}
    _ resource.ResourceWithImportState = &<name>Resource{}
)

type <name>Resource struct {
    client   *sgclient.Client
    org_name string
}

func NewResource() resource.Resource {
    return &<name>Resource{}
}

func (r *<name>Resource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
    resp.TypeName = req.ProviderTypeName + "_<terraform_name>"
}

func (r *<name>Resource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
    if req.ProviderData == nil {
        return
    }
    provider, ok := req.ProviderData.(*customTypes.ProviderInfo)
    if !ok {
        resp.Diagnostics.AddError(
            "Unexpected Resource Configure Type",
            fmt.Sprintf("Expected *customTypes.ProviderInfo, got: %T.", req.ProviderData),
        )
        return
    }
    r.client = provider.Client
    r.org_name = provider.Org_name
}

func (r *<name>Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
    resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}

func (r *<name>Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
}

func (r *<name>Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
}

func (r *<name>Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
}

func (r *<name>Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
}
```

---

### 3. `schema.go`

Implement `Schema()` as a method on the resource struct. Use the type mapping and model struct rules in CLAUDE.md as the reference for mapping SDK types to schema attributes and `AttributeTypes()` values.

- `id` is always `Required: true`.
- Required inputs from the user → `Required: true`.
- Add `MarkdownDescription` from `internal/constants` or an inline string if no constant matches.
- After writing the schema, ask the user to confirm that all SDK fields are represented, types match, and docs are accurate (a string that contains JSON needs marshal/unmarshal in the converters).

```go
func (r *<name>Resource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
    resp.Schema = schema.Schema{
        Attributes: map[string]schema.Attribute{
            // ... attributes
        },
    }
}
```

---

### 4. `model.go`

```go
package <name>

// Top-level model — one field per schema attribute.
type <Name>ResourceModel struct {
    Id           types.String `tfsdk:"id"`
    // ... other fields
}

// ToAPIModel converts the Terraform plan/state to the SDK Create request type.
func (m *<Name>ResourceModel) ToAPIModel(ctx context.Context) (*sdkpkg.<CreateStruct>, diag.Diagnostics) {
    apiModel := &sdkpkg.<CreateStruct>{
        Field: m.Field.ValueStringPointer(),
    }
    // Handle optional/null fields with sgsdkgo.Optional / sgsdkgo.Null
    return apiModel, nil
}

// ToUpdateAPIModel converts for the SDK Update/Patch request type.
// Check the SDK struct — some fields present in Create may be absent in Update.
func (m *<Name>ResourceModel) ToUpdateAPIModel(ctx context.Context) (*sdkpkg.<UpdateStruct>, diag.Diagnostics) {
    apiModel := &sdkpkg.<UpdateStruct>{}
    if !m.Field.IsNull() && !m.Field.IsUnknown() {
        apiModel.Field = sgsdkgo.Optional(m.Field.ValueString())
    } else {
        apiModel.Field = sgsdkgo.Null[string]()
    }
    return apiModel, nil
}

// convert<Name>FromAPI converts an SDK response to the Terraform model.
func convert<Name>FromAPI(ctx context.Context, resp *sdkpkg.<ResponseStruct>) (<Name>ResourceModel, diag.Diagnostics) {
    var diags diag.Diagnostics
    model := <Name>ResourceModel{
        Id:    flatteners.String(resp.Data.Id),
        Field: flatteners.StringPtr(resp.Data.Field),
    }
    return model, diags
}
```

For complex nested structs, converter patterns, `AttributeTypes()` rules, diagnostics handling, and null guards see CLAUDE.md.

---

### 5. `resource.go` (CRUD methods)

The SDK does not follow a consistent naming convention for CRUD methods or client groups. Before implementing:
1. Inspect the SDK source (via `go.mod` replace path) to find the correct client group field on `r.client` and the exact method names for create, read, update, and delete.
2. Make your best guess at the group name and method names, then confirm with the user before writing the code.

Each method follows this logical flow:

**Create:** read plan → `ToAPIModel` → SDK create call → SDK read call (using the returned id) → `convert<Name>FromAPI` → `resp.State.Set`

**Read:** read state → SDK read call (using `state.Id`) → `convert<Name>FromAPI` → `resp.State.Set`

**Update:** read plan + state → `ToUpdateAPIModel` → SDK update call → SDK read call → `convert<Name>FromAPI` → `resp.State.Set`

**Delete:** read state → SDK delete call

Apply the same guard pattern (`HasError` check after every `Append`) across all four methods.

---


### 6. `resource_test.go`

```go
package <name>_test

import (
    "fmt"
    "net/http"
    "testing"

    "github.com/StackGuardian/terraform-provider-stackguardian/internal/acctest"
    "github.com/hashicorp/terraform-plugin-testing/helper/resource"
    "github.com/hashicorp/terraform-plugin-testing/tfversion"
)

// testAccResource returns a Terraform config for the resource.
// resourceName  — the Terraform resource label (also used as the logical name)
// requiredField — value for any required root-level attribute
// additional_config         — any additional attribute lines to append inside the block (pass "" for none)
func testAccResource(resourceName, requiredField, additional_config string) string {
    return fmt.Sprintf(`
resource "stackguardian_<terraform_name>" %q {
  required_field = %q
  %s
}`, resourceName, requiredField, additional_config)
}

func TestAcc<Name>(t *testing.T) {
    resourceName := "<name>-example"
    requiredField := "<value>"

    resource.ParallelTest(t, resource.TestCase{
        PreCheck: func() { acctest.TestAccPreCheck(t) },
        TerraformVersionChecks: []tfversion.TerraformVersionCheck{
            tfversion.SkipBelow(tfversion.Version1_1_0),
        },
        ProtoV6ProviderFactories: acctest.ProviderFactories(http.Header{}),
        Steps: []resource.TestStep{
            {Config: testAccResource(resourceName, requiredField, "")},
            // Add an update step passing additional_config config via the third argument
            // {Config: testAccResource(resourceName, requiredField, `optional_field = "updated"`)},
        },
    })
}
```

---

### 7. Register in the provider

In `internal/provider/provider.go`, add to the `Resources()` function:

```go
import <name> "github.com/StackGuardian/terraform-provider-stackguardian/internal/resource/<name>"

// inside Resources():
<name>.NewResource,
```

---

## Checklist

- [ ] SDK struct names, client group, and CRUD method names confirmed with user
- [ ] `resource.go` — struct, `NewResource`, `Metadata`, `Configure`, `ImportState`, empty CRUD placeholders
- [ ] `schema.go` — `Schema()` with all attributes; `id` is `Required: true`; user has confirmed all SDK fields are represented and docs are accurate
- [ ] `model.go` — `<Name>ResourceModel`, `ToAPIModel`, `ToUpdateAPIModel`, `convert<Name>FromAPI`; nested structs have `AttributeTypes()` and `ToAPIModel()`
- [ ] `resource.go` CRUD methods — filled in after schema + model are confirmed
- [ ] `resource_test.go` — `testAccResource` function with `resourceName`, required field params, and `additional_config`; at least one create step
- [ ] Provider registered in `internal/provider/provider.go`
- [ ] `go build ./...` passes with no errors

See CLAUDE.md → Common Gotchas for the full list of pitfalls.
