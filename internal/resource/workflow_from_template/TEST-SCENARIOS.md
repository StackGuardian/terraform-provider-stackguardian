# Test Scenarios — `workflow_from_template` + `workflow_template_revision` data source

## Unit tests (pure Go, no API)

| Test | Scenario covered |
| --- | --- |
| `TestSplitTemplateOrg` | Parsing `iac_template_id` into (owning-org, revision-id): own fully-qualified path, shared cross-org path, bare id, empty id. |
| `TestConvertEnvironmentVariablesFromAPI_NilConfig` | Flattening env vars where an entry's `Config` pointer is nil must not panic (nil-safety regression guard). |

## Acceptance tests (`TF_ACC`, live API)

### Create / resolution / merge

| Test | Scenario covered |
| --- | --- |
| `..._Basic` | Create from template; re-plan is clean (no perpetual diff). |
| `..._FullResolution` | User declares a subset; all remaining fields resolve from the template onto state. |
| `..._TemplateDefaultsResolved` | A field the user never declared is resolved from the template; second plan is clean. |
| `..._WithDescription` | `description` inherited/overridden from template. |
| `..._WithEnvironmentVariables` | `environment_variables` create + round-trip. |
| `..._WithTagsAndContextTags` | `tags` + `context_tags` create + round-trip. |
| `..._WithApprovers` | `approvers` create + round-trip. |
| `..._WithRunnerConstraints` | `runner_constraints` nested object create + round-trip. |
| `..._WithUserSchedules` | `user_schedules` create + round-trip. |
| `..._WithIacInputData` | `iac_input_data` default derived from the template's `FORM_JSONSCHEMA`. |
| `..._WithTerraformConfig` | `terraform_config` create + round-trip. |
| `..._WithWfStepsConfig` | Top-level `wf_steps_config` (CUSTOM-only) with required `wf_step_template_id`; round-trip stable. |
| `..._LifecycleWfStepsConfig` | Lifecycle hook list `terraform_config.post_apply_wf_steps_config`; round-trip stable. |
| `..._InNestedWorkflowGroup` | Create inside a nested workflow group. |

### Update

| Test | Scenario covered |
| --- | --- |
| `..._NormalUpdate` | Create → modify a field → apply → plan clean (in-place update). |
| `..._WithWfStepTemplateRevisionId` | `terraform_config.wf_step_template_revision_id` set in config + updated. |
| `..._WfStepRevisionInheritedFromTemplate` | `wf_step_template_revision_id` inherited from template; explicit `""` suppresses inheritance. |

### Empty-value / suppression semantics

| Test | Scenario covered |
| --- | --- |
| `..._EmptyAllowBlankFalse` | Empty values on `allow_blank=false` fields handled across create/update/refresh. |
| `..._ExplicitEmptySuppressesTemplateDefault` | An explicit empty list/dict suppresses a template default (vs. null → inherit). |

### Drift

| Test | Scenario covered |
| --- | --- |
| `..._DriftDetection` | An out-of-band change (e.g. dev portal) is detected as a non-empty plan on refresh. |
| `..._DriftCronDroppedWhenCheckFalse` | `drift_cron` is dropped whenever `drift_check` is false (coupling), stored as `""` in state. |

### Revision upgrade (`iac_template_id` → new revision)

| Test | Scenario covered |
| --- | --- |
| `..._RevisionUpgrade` | Upgrade to a new revision; user-set fields preserved, unset fields adopt the new default. |
| `..._RevisionUpgradeNoSpuriousDependentUpdate` | Upgrade where derived fields (desc/tags/approvers/context_tags/env_vars/ints/runner_constraints/deployment_platform_config/mini_steps) are unchanged → a dependent resource referencing them stays NoOp. |
| `..._WfStepsConfigRevisionUpgradeNoSpuriousDependentUpdate` | Same, for `wf_steps_config` (concrete resolution keeps a dependent NoOp on upgrade). |

### `workflow_template_revision` data source

| Test | Scenario covered |
| --- | --- |
| `TestAccWorkflowTemplateRevisionDataSource_Custom` | Read a CUSTOM revision (top-level `wf_steps_config`, populated `deployment_platform_config` list, env vars) — all flatten correctly. |
| `TestAccWorkflowTemplateRevisionDataSource_Terraform` | Read a TERRAFORM revision (populated `terraform_config` + git `runtime_source`); `terraform_version` normalized to the bare form. |
