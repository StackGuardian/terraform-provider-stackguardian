# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).


## [Unreleased]

### Added

- Centralized configuration via `github.com/spf13/viper` in `internal/config/config.go`, replacing all `os.Getenv` calls throughout the codebase
- Minimum supported Terraform version of 1.5.7, enforced at runtime: the provider now returns a clear error diagnostic on an older CLI instead of failing obscurely later (`internal/provider/version_check.go`)
- `constants.MinTerraformVersion`, `constants.MaxTerraformVersion`, and `constants.SupportedTerraformVersions` as the single definition of the supported Terraform range
- `acctest.VersionChecks()` — shared `TerraformVersionChecks` for every acceptance test case, pinned to the support floor
- `acctest.SkipUnlessAcceptance()` — skips an acceptance test when `TF_ACC` is unset, before any API fixture is created
- `scripts/tf-compat-check.sh` and `make tf-compat-check` — credential-free check that the provider loads, its schema decodes, and every `docs-examples/` configuration validates on a given Terraform CLI
- `make test-acc-min-terraform` — runs the acceptance suite against the oldest supported Terraform release
- `.github/workflows/compat.yaml` — compatibility gate on pull requests to `develop`: build, `go vet`, unit tests, docs validation, and the Terraform compatibility matrix
- Unit tests covering the version floor and guarding the CI matrix against drift: `TestCheckTerraformVersion`, `TestSupportedTerraformVersions`, `TestTerraformMatrixMatchesWorkflows`, `TestMakefileUsesMinTerraformVersion`
- `required_version = ">= 1.5.7"` in the quickstart and onboarding example configurations
- "Terraform version support" and "Continuous integration" sections in the README

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
- Acceptance tests now use `tfversion.RequireAbove` at the support floor instead of `tfversion.SkipBelow(Version1_1_0)`, so a run below the floor fails loudly rather than skipping every test and reporting green
- Add the missing `TerraformVersionChecks` to the three acceptance test cases that declared none
- `.github/workflows/test.yaml` runs the acceptance suite across the full supported Terraform matrix (1.5.7 through 1.16.2, the latest patch of every minor series) on pull requests to `main`, replacing the single pinned 1.14.0 run
- Rename the CI development branch from `devel` to `develop`
- `.github/workflows/release.yaml` now requires the compatibility gate in addition to the acceptance suite
- Remove all `schedule` (cron) triggers; the API high-load workflows `test-api-prd.yaml` and `test-api-stg.yaml` are now manual (`workflow_dispatch`) only
- Acceptance jobs run one Terraform version at a time (`max-parallel: 1`) under a `concurrency` group, since they all share a single StackGuardian organization

### Fixed

- Fix non-constant format string vet errors in `workflow_template_revision` tests for Go 1.26 compatibility
- Fix unchecked error returns (errcheck) from SDK `Delete*`/`Update*` calls in test cleanup functions
- Fix unchecked `defer Body.Close()` in 4 datasource files
- Fix staticcheck SA4006 unused `diags` in `workflow_template_revision/model.go` and `workflow_template/model.go`
- Remove unused `charSetAlphaNum` constant in `internal/acctest/random_acc_test_name.go`
- Fix gofmt alignment in `constants/template.go`, `constants/workflow.go`, `datasources/workflow_template_revision/schema.go`, and `resource/workflow_from_template/model.go`
- Fix `make test` failing without API credentials: 84 acceptance tests created API fixtures before `resource.Test` reached its `TF_ACC` check, so a credential-free run failed on 401 errors instead of skipping
- Fix `docs-examples/datasources/workflow_group` referencing the managed resource instead of the data source in its output
- Fix `docs-examples/datasources/workflow_template` setting the computed `template_name` attribute instead of the required `id`
- Fix `docs-examples/resources/workflow_template_revision` omitting the required `user_job_cpu` and `user_job_memory` arguments from its basic example


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
