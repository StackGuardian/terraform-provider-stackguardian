# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

> **Releases from 0.1.0 onward are documented on the
> [GitHub releases page](https://github.com/StackGuardian/terraform-provider-stackguardian/releases),
> which is generated from the merged pull requests for each tag. The entries below cover the
> initial release series and are kept for historical reference.**


## [1.12.4]

### Breaking

- `stackguardian_workflow_template_revision`: every `user_schedules` entry now requires an
  `inputs` block with `inputs.terraform_action.action` set. Existing configurations whose
  schedules have no `inputs` fail validation after upgrading; add, for example:

  ```hcl
  user_schedules = [
    {
      cron  = "0 8 ? * MON *"
      state = "ENABLED"
      inputs = {
        terraform_action = { action = "apply" }
      }
    }
  ]
  ```

### Added

- `user_schedules[*].inputs` on `stackguardian_workflow_template_revision` and its data source:
  the run inputs a schedule uses when it triggers a run — `terraform_action.action` and
  `vcs_config.iac_input_data` (`schema_type` and `data`, a JSON string
  such as `jsonencode({ test = "value" })`). `enable_chaining` is read-only: set by
  StackGuardian and sent back unchanged on update
- Plan-time validation that `user_schedules[*].inputs.vcs_config.iac_input_data.schema_type`
  is `RAW_JSON` or `FORM_JSONSCHEMA`
- `terraform_config.run_pre_plan_hooks_on_drift` and `terraform_config.run_post_plan_hooks_on_drift`
  on `stackguardian_workflow_template_revision` and its data source

### Changed

- `terraform_config.run_pre_init_hooks_on_drift`, `run_pre_plan_hooks_on_drift` and
  `run_post_plan_hooks_on_drift` on `stackguardian_workflow_template_revision` default to
  `false`. Omitting one (or setting it to `null`) now sends `false`; previously an omitted
  `run_pre_init_hooks_on_drift` kept its last value.

### Fixed

- Importing a `stackguardian_workflow_template_revision` (`<template_id>:<revision>`) now sets
  `template_id`. It was left null, so the first plan after an import failed with
  "template_id cannot be changed". A malformed import ID now returns a clear error.
- `user_schedules[*].name` and `desc` on `stackguardian_workflow_template_revision` are now
  Optional + Computed. Schedules created in the UI store `""` for both; leaving them out of the
  config after an import planned a change to null on every run, and the update was then
  rejected for published revisions ("Cannot update … UserSchedules … for a published template").

## [Unreleased]

### Added

- Centralized configuration via `github.com/spf13/viper` in `internal/config/config.go`, replacing all `os.Getenv` calls throughout the codebase
- `ValidateConfig` on `workflow_template` and `workflow_template_revision` resources to enforce `runtime_source` auth/is_private rules at plan time
- `wf_steps_config` validation on `workflow_template_revision` to reject usage when `source_config_kind` is `TERRAFORM` or `OPENTOFU`
- `internal/constants/vcs.go` with shared enum constants for VCS provider kinds (`GITHUB_COM`, `GIT_OTHER`, etc.)
- `acctest.TFStandardErrorPattern` helper that builds regex patterns tolerating Terraform CLI line-wrapping in expected error messages
- Comprehensive acceptance tests covering every schema attribute on `workflow_template` and `workflow_template_revision`

### Changed

- Update Go version from 1.21.4 to 1.26.7
- Update `terraform-plugin-framework` from v1.11.0 to v1.19.0
- Update `terraform-plugin-framework-validators` from v0.12.0 to v0.19.0
- Update `terraform-plugin-go` from v0.23.0 to v0.31.0
- Update `terraform-plugin-log` from v0.9.0 to v0.11.0
- Update `terraform-plugin-testing` from v1.10.0 to v1.16.0
- Update `terraform-plugin-docs` (tools) from v0.18.0 to v0.25.0
- Rename `ProviderInfo.Org_name` to `ProviderInfo.OrgName` for Go naming consistency
- Rename `stackguardianProviderModel` fields (`Api_key` → `APIKey`, `Api_uri` → `APIUri`, `Org_name` → `OrgName`) and local variables in `Configure()` from snake_case to camelCase for Go naming consistency
- Fix "Stackguardian" → "StackGuardian" casing in provider log messages, error messages, and comments
- Replace hardcoded `os.Getenv` calls with `config.Get()` singleton in provider, acctest, and all resource/datasource test files
- Make `mount_point.read_only` `Computed` with `UseStateForUnknown()` on `workflow_template_revision` to match API behavior
- Make `input_schemas.type` `Required` (was `Optional`) on `workflow_template_revision` to match actual API behavior
- Shorten import alias `workflowtemplate` → `wft` across `provider.go`, `datasource.go`, `resource.go`, `model.go`, and `schema.go`

### Fixed

- Fix non-constant format string vet errors in `workflow_template_revision` tests for Go 1.26 compatibility
- Fix unchecked error returns (errcheck) from SDK `Delete*`/`Update*` calls in test cleanup functions
- Fix unchecked `defer Body.Close()` in 4 datasource files
- Fix staticcheck SA4006 unused `diags` in `workflow_template_revision/model.go` and `workflow_template/model.go`
- Remove unused `charSetAlphaNum` constant in `internal/acctest/random_acc_test_name.go`
- Fix gofmt alignment in `constants/template.go`, `constants/workflow.go`, `datasources/workflow_template_revision/schema.go`, and `resource/workflow_from_template/model.go`
- Guard Optional+Computed fields (`LongDescription`, `Notes`, `IsPublic`, `NumberOfApprovalsRequired`) in `WorkflowTemplateRevisionResourceModel.ToAPIModel` against Unknown values, preventing unintended zero-value sends on Create
- Guard `mount_point.read_only` in `ConvertMountPointsListToAPI` against Unknown values to prevent sending an explicit `false`
- Replace `os.Getenv("STACKGUARDIAN_ORG_NAME")` with `config.Get().OrgName` in `sweep_test.go`


## [0.1.0] - 2024-03-14

- First GA Release on the Terraform Registry

### Added

- Initial Terraform Provider for StackGuardian
- Resource and Data-Source for StackGuardian Workflow
- Resource and Data-Source for StackGuardian Stack
- Resource and Data-Source for StackGuardian Policy
- Resource and Data-Source for StackGuardian Integration
- Data-Source for StackGuardian Workflow Outputs
- Tests for Resources
- Examples for Resources
- Quickstart guide
- Documentation
- GH workflows for test & release


## [0.1.0-rc4] - 2024-03-08

### Added

- Validation for provider docs

### Fixed

- Provider docs
- Cleanup repo


## [0.1.0-rc3] - 2024-03-07

### Fixed

- _Nihil ad rem_ release for Terraform Registry


## [0.1.0-rc2] - 2024-03-07

### Added

- Release Test with quickstart example
- CLI Test with quickstart example


## [0.1.0-rc1] - 2024-01-15

### Added

- Cleanup & Tests for each resource


## [0.1.0-beta1] - 2023-12-22

### Added

- TF Data-Source for StackGuardian Workflow Outputs


## [0.1.0-alpha1] - 2023-10-16

### Added

- Initial TF provider for StackGuardian
- TF Resource and Data-Source for StackGuardian Workflow
- TF Resource and Data-Source for StackGuardian Stack
- TF Resource and Data-Source for StackGuardian Policy
- TF Resource and Data-Source for StackGuardian Integration
