# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

> **Releases from 0.1.0 onward are documented on the
> [GitHub releases page](https://github.com/StackGuardian/terraform-provider-stackguardian/releases),
> which is generated from the merged pull requests for each tag. The entries below cover the
> initial release series and are kept for historical reference.**


## [Unreleased]

### Added

- Approvers guide and corrected `approvers` attribute descriptions: the entry formats for local users, SSO users and SSO groups, and the user pool ID each region takes
- Centralized configuration via `github.com/spf13/viper` in `internal/config/config.go`, replacing all `os.Getenv` calls throughout the codebase
- `ValidateConfig` on `workflow_template` and `workflow_template_revision` resources to enforce `runtime_source` auth/is_private rules at plan time
- `wf_steps_config` validation on `workflow_template_revision` to reject usage when `source_config_kind` is `TERRAFORM` or `OPENTOFU`
- `internal/constants/vcs.go` with shared enum constants for VCS provider kinds (`GITHUB_COM`, `GIT_OTHER`, etc.)
- `acctest.TFStandardErrorPattern` helper that builds regex patterns tolerating Terraform CLI line-wrapping in expected error messages
- Comprehensive acceptance tests covering every schema attribute on `workflow_template` and `workflow_template_revision`
- `internal/constants/terraform_version.go` with `MinTerraformVersion` (1.5.7), `MaxTerraformVersion` (1.16.4), and `SupportedTerraformVersions` CI matrix
- `internal/provider/version_check.go` runtime check that refuses to configure below `MinTerraformVersion` with an actionable error message
- `internal/acctest/skip.go` with `SkipUnlessAcceptance(t)` helper for early `TF_ACC` guard before fixture-creation API calls
- `internal/acctest/tfversion.go` with `VersionChecks()` returning `tfversion.RequireAbove(MinTerraformVersion)`
- `scripts/tf-compat-check.sh` hermetic Terraform CLI compatibility gate (schema decode + example validation, no credentials)
- `compat.yaml` workflow — hermetic compatibility gate targeting `main`: build, vet, unit tests, docs validation, and TF compatibility matrix
- Terraform version matrix (1.5.7–1.16.4) in `test.yaml` acceptance workflow
- `concurrency: acceptance-tests` group in `test.yaml` to prevent overlapping acceptance runs
- `Makefile` targets `tf-compat-check` and `test-acc-min-terraform`

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
- All CI workflows now target `main` only (removed `develop` branch triggers from `compat.yaml`, `test-api-stg.yaml`, `test-api.yaml`)
- `test-api-stg.yaml` and `test-api.yaml` default `gitref` changed from `develop` to `main`
- `release.yaml` now calls `compat.yaml` as a pre-release gate
- `test.yaml` installs Terraform CLI before the docs check step

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
- Fix `make test` failing in CI with `401: Unauthorized` — acceptance test fixtures made API calls before the `TF_ACC` check; added `SkipUnlessAcceptance(t)` guard to 54 `TestAcc*` functions
- Fix `make docs-validate-examples` failing in `test.yaml` CI with `terraform: command not found` — Terraform CLI is now installed before the docs check step

### Security

- **CRITICAL:** Stop logging the full API key in debug output (`provider.go`); the key is now redacted to `***redacted***`
- **CRITICAL:** Mark all connector credential fields as `Sensitive: true` in both the `stackguardian_connector` resource and data source schemas — `github_app_webhook_secret`, `github_app_client_secret`, `github_app_pem_file_content`, `gitlab_creds`, `azure_creds`, `bitbucket_creds`, `aws_access_key_id`, `aws_secret_access_key`, `arm_client_secret`, `gcp_config_file_content` were previously written in plaintext to plan output and state
- **HIGH:** Mark `runner_group_token` data source output as `Sensitive: true` — the registration token was previously exposed in plaintext in plan output and state
- **HIGH:** Mark `webhook_secret` as `Sensitive: true` across all resources and data sources that use it (`workflow_git`, `workflow_from_template`, `stack_template_revision`, `workflow_template_revision` and their data sources)
- **HIGH:** Mark `text_value` env var field as `Sensitive: true` across all resources and data sources — the field description itself warns that values are visible in configuration and state
- **HIGH:** Mark `azure_blob_storage_access_key` as `Sensitive: true` in `stackguardian_runner_group` resource and data source
- **MEDIUM:** Add 30-second HTTP client timeout to `runner_group_token` data source API call, replacing `http.DefaultClient` which had no timeout
- **MEDIUM:** Replace `context.TODO()` with the request context (`ctx`) in `workflow_git` and `runner_group` resource `Create` operations so Ctrl+C can cancel in-flight API calls
- **MEDIUM:** URL-encode `orgName` and `runnerGroupID` in the `runner_group_token` data source HTTP request to prevent URL path injection
- **MEDIUM:** Stop dumping the raw HTTP response body in error messages from the `runner_group_token` data source; only the status code is now included, and response body reads are capped at 1 MB
- **LOW:** Mark `docker_registry_username` as `Sensitive: true` across `workflow_step_template`, `runner_group`, and their data sources
- **LOW:** Fix copy-paste error in `runner_group_token` data source configure error message (`"*hashicups.Client"` → `"*customTypes.ProviderInfo"`)
- **LOW:** Fix incorrect error summary `"Missing Organization Name"` → `"Missing API Key"` and wrong env var name in the provider's missing API key error


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
