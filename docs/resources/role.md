---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "stackguardian_role Resource - terraform-provider-stackguardian"
subcategory: ""
description: |-

---

# stackguardian_role (Resource)

## Example Usage

```terraform
resource "stackguardian_role" "TPS-Example-Role" {
  data = jsonencode({
    "ResourceName" : "TPS-Example-Role",
    "Description" : "Example of terraform-provider-stackguardian for Role",
    "Tags" : ["tf-provider-example"],
    "Actions" : [
      "Org-Name-1"
    ],
    "AllowedPermissions" : {
      "Permission-key-1" : "Permission-val-1",
      "Permission-key-2" : "Permission-val-2"
    }
  })
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `data` (String) Valid JSON data that this provider will manage with the API server. Please refer to the API Docs: https://docs.stackguardian.io/api#tag/Role

### Read-Only

- `api_data` (Map of String) After data from the API server is read, this map will include k/v pairs usable in other terraform resources as readable objects. Currently the value is the golang fmt package's representation of the value (simple primitives are set as expected, but complex types like arrays and maps contain golang formatting).
- `api_response` (String) The raw body of the HTTP response from the last read of the object.
- `id` (String) The ID of this resource.

