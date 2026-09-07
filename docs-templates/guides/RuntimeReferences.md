---
page_title: "Runtime References"
subcategory: "Concepts"
description: |-
  Values the platform resolves when a workflow runs — secrets, outputs of other workflows, and stack template outputs.
---

# Runtime References

Some values cannot be known when Terraform runs. A database password lives in a vault; a subnet
ID is produced by a workflow that has not run yet. For these, you write a **reference** — a
`${...}` token that Terraform stores verbatim and StackGuardian resolves later, when the workflow
actually runs.

This is the opposite of a
[Resource ID](https://registry.terraform.io/providers/StackGuardian/stackguardian/latest/docs/guides/ResourceIDs),
which Terraform resolves while planning. A reference is still a reference in your state file.
That is the point: a secret reference is safe to put in a `PLAIN_TEXT` environment variable,
because what lands in configuration and state is the token, never the secret.

## Writing one in Terraform

Terraform reads `${` as the start of its **own** interpolation. Write a reference with a bare
`${` and Terraform will try to evaluate it, and fail — or worse, quietly produce something else.

Double the dollar sign. `$${` sends a literal `${` through to StackGuardian:

```terraform
resource "stackguardian_workflow_git" "api" {
  # ...
  environment_variables = [
    {
      kind = "PLAIN_TEXT"
      config = {
        var_name   = "DATADOG_API_KEY"
        text_value = "$${secret::datadog-api-key}"
      }
    },
  ]
}
```

The value the platform receives is `${secret::datadog-api-key}`. Inside `jsonencode`, or any
other Terraform string, the same rule applies.

## The forms

| Form | Resolves to |
| --- | --- |
| `${secret::<name>}` | a secret in your organization |
| `${ext-secret::azure-kv::<vault>.<secret>}` | a secret in an external Azure Key Vault |
| `${workflow::<group>.<workflow>.<output-key>}` | an output of another workflow |
| `${reference::<template-group>.<index>.outputs.<key>}` | an output of another template — stack templates only |

The separator after the type is a **double colon**. A single dot resolves to nothing and fails
silently at run time.

### Secrets

```
${secret::db-password}
```

### External secrets

Azure Key Vault only. The version segment is optional — omit it for the latest version:

```
${ext-secret::azure-kv::<vault-name>.<secret-name>}
${ext-secret::azure-kv::<vault-name>.<secret-name>.<version>}
```

Access is authorized through the workflow's deployment and environment configuration, not
through a connector you name in the reference.

### Workflow outputs

The shape depends on what produced the output.

A Terraform workflow's outputs keep Terraform's own envelope, so the key is followed by
`.value`:

```
${workflow::<group>.<workflow>.outputs.<variable>.value}
```

A custom workflow writes its outputs directly, with no envelope:

```
${workflow::<group>.<workflow>.outputs.<variable>}
```

Index into a list output by position:

```
${workflow::<group>.<workflow>.outputs.<variable>.value.1}
```

A nested workflow group keeps the `/` that separates parent from child, exactly as in the
group's `id`:

```
${workflow::<parent-group>/<group>.<workflow>.<output-key>}
```

A workflow **inside a stack** takes the stack as an extra segment between the group and the
workflow:

```
${workflow::<group>.<stack>.<workflow>.<output-key>}
```

There is no separate `stack` reference type — a stack's workflows are addressed the way
shown above.

## Where you can write one

References are resolved in:

- `environment_variables[].config.text_value` — on workflows, stack template revisions, workflow
  template revisions, and per-step under `wf_steps_config`
- `iac_input_data.data` — the JSON string of inputs handed to the IaC
- `wf_steps_config[].wf_step_input_data.data` — the same, per step

Attributes that accept a reference say so in their description.

Everywhere else, a `${...}` token is stored as a literal string. If you are unsure whether a
field resolves, check the description before relying on it — an unresolved reference does not
error, it simply arrives at your workflow as the token you typed.

## `${reference::}` is for stack templates

A stack runs several templates in order, and one template usually needs a value the previous one
produced. Inside a **stack template**, that is a `${reference::}` token rather than a
`${workflow::}` one, because the templates are being wired together before any of them has a
workflow to name:

```
${reference::template-group.0.outputs.cluster_name}
```

The `.0` selects a template by position. In the provider, that ordering is the
`workflows_config.workflows` list on `stackguardian_stack_template_revision` — the first entry is
`.0`. The leading `template-group` segment is the platform's own grouping of the templates in the
stack; the provider has no attribute of that name.

## Run time or plan time

If Terraform itself needs the value — to pass into another provider's resource, to build a
string, to make a decision — a reference is the wrong tool, because Terraform never sees what it
resolves to. Read the value instead:

- [`stackguardian_workflow_outputs`](https://registry.terraform.io/providers/StackGuardian/stackguardian/latest/docs/data-sources/workflow_outputs)
- [`stackguardian_stack_outputs`](https://registry.terraform.io/providers/StackGuardian/stackguardian/latest/docs/data-sources/stack_outputs)
- [`stackguardian_stack_workflow_outputs`](https://registry.terraform.io/providers/StackGuardian/stackguardian/latest/docs/data-sources/stack_workflow_outputs)

These read the outputs of the most recent run at plan time, as ordinary Terraform data.

The trade-off is what each one waits for. A data source needs the run to have already happened,
and re-plans when its outputs change. A reference needs nothing at plan time and picks up
whatever the referenced workflow last produced, at the moment your workflow runs.

Reach for a data source when Terraform needs the value. Reach for a reference when only the
workflow does.

## Where to go next

- [Resource IDs](https://registry.terraform.io/providers/StackGuardian/stackguardian/latest/docs/guides/ResourceIDs) — the references Terraform resolves, and how to build them.
- [Object Model](https://registry.terraform.io/providers/StackGuardian/stackguardian/latest/docs/guides/ObjectModel) — how workflows, stacks and templates relate.
- [Templates and Revisions](https://registry.terraform.io/providers/StackGuardian/stackguardian/latest/docs/guides/Templates) — where stack templates fit.
