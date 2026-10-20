## UpdateState

Without it, anyone already using stackguardian_stack_template_revision with a deployment_platform_config would hit an error on their first plan after upgrading the
provider. The only fix they'd have would be editing their state by hand.

### What changed

The attribute's type changed between releases:

hcl

#### provider 1.12.5 — config was a string holding JSON
```terraform
deployment_platform_config = [{
    kind   = "AWS_RBAC"
    config = jsonencode({ integrationId = "/integrations/aws-prod" })
}]
```

#### provider 1.13.0 — config is an object
```terraform
deployment_platform_config = [{
  kind   = "AWS_RBAC"
  config = { integration_id = "/integrations/aws-prod" }
}]
```

### Why that breaks existing users

Terraform saves every resource's attributes to the state file, in the format of the provider version that wrote it. A user on 1.12.5 has this in terraform.tfstate:

```text
"config": "{\"integrationId\":\"/integrations/aws-prod\"}"
```

On the next plan with 1.13.0, the provider must decode that state against its current schema before doing anything else. The schema now expects an object and finds a
string, so it fails:

invalid JSON, expected "{", got "{\"integrationId\":...}"

This happens before any API call, so terraform refresh and terraform apply hit the same error. The user
can't fix it by editing their .tf files either, because the problem is in the saved state, not the configuration. Their only way out would be editing state by hand or
removing the resource from state and re-importing it.

### What UpgradeState does

The Plugin Framework's built-in mechanism for this has two parts:

1. A schema version number. The schema now declares Version: 1; before it was implicitly 0. Terraform records that number in state next to each resource.
2. A converter from version 0. When Terraform loads state recorded at version 0, it doesn't decode it directly. It first passes the raw JSON to our upgrader, which rewrites
   the string into the object:

```text
"config": "{\"integrationId\":\"/integrations/aws-prod\",\"profileName\":\"default\"}"
↓
"config": { "integration_id": "/integrations/aws-prod", "profile_name": "default" }
```

Terraform then stores the result as version 1, so this happens only once per resource. Right after the upgrade it also refreshes the resource from the API, so any value the
converter couldn't read (an empty or invalid string becomes null) is filled back in.

### Why only this attribute needed it

This commit also removed several attributes (template_id, custom_source, input_schemas[*].is_committed…). Those don't need conversion: the framework skips attributes in old
state that the schema no longer defines. A changed type is what can't be decoded, and the full schema comparison found exactly two places: config under
workflows_config.workflows[*] and under actions[*].order[*].parameters.

### What users still have to do

UpgradeState fixes the state, not the configuration. Users still need to rewrite config = jsonencode({...}) as config = { integration_id = ... } in their .tf files, because
the old form is now a type error. That's why it's listed as a breaking change in the CHANGELOG. The difference is that it's an ordinary edit followed by plan, with no
manual state surgery.
