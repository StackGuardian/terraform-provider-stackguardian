# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).


## [Unreleased]

### Added

- Centralized configuration via `github.com/spf13/viper` in `internal/config/config.go`, replacing all `os.Getenv` calls throughout the codebase

### Changed

- Update Go version from 1.21.4 to 1.26.7
- Update `terraform-plugin-framework` from v1.11.0 to v1.19.0
- Update `terraform-plugin-framework-validators` from v0.12.0 to v0.19.0
- Update `terraform-plugin-go` from v0.23.0 to v0.31.0
- Update `terraform-plugin-log` from v0.9.0 to v0.11.0
- Update `terraform-plugin-testing` from v1.10.0 to v1.16.0
- Update `terraform-plugin-docs` (tools) from v0.18.0 to v0.25.0
- Rename `ProviderInfo.Org_name` to `ProviderInfo.OrgName` for Go naming consistency
- Replace hardcoded `os.Getenv` calls with `config.Get()` singleton in provider, acctest, and all resource/datasource test files

### Fixed

- Fix non-constant format string vet errors in `workflow_template_revision` tests for Go 1.26 compatibility
- Fix unchecked error returns (errcheck) from SDK `Delete*`/`Update*` calls in test cleanup functions
- Fix unchecked `defer Body.Close()` in 4 datasource files
- Fix staticcheck SA4006 unused `diags` in `workflow_template_revision/model.go` and `workflow_template/model.go`
- Remove unused `charSetAlphaNum` constant in `internal/acctest/random_acc_test_name.go`
- Fix gofmt alignment in `constants/template.go`, `constants/workflow.go`, `datasources/workflow_template_revision/schema.go`, and `resource/workflow_from_template/model.go`


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
