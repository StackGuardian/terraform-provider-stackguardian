# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).
Versions use the `MAJOR.MINOR.PATCH` format. While the provider is under active development,
breaking changes can ship in any release and are marked **Breaking**.

Entries describe changes that affect provider users: resources, data sources, attributes,
behavior, documentation and security. Internal refactors, tests and CI changes are left out.
A **Breaking** change can require edits to existing configurations.

## [Unreleased]

## [1.12.4] - 2026-10-06

### Added

- Approvers guide, and corrected `approvers` attribute descriptions covering the entry formats for local users, SSO users and SSO groups, and the user pool ID each region takes ([#151])

### Changed

- **Breaking:** The provider now refuses to run on Terraform CLI versions older than 1.5.7 and returns an error explaining how to upgrade. The provider is tested against Terraform 1.5.7 through 1.16.4 ([#134])
- **Breaking:** Several credential and secret attributes are now marked sensitive (see Security below). Terraform redacts them in plan output, and any `output` that references one must now set `sensitive = true` ([#147])

### Fixed

- `stackguardian_role_assignment`: creating an assignment for a user who already exists now adopts the existing user when the API responds with `409 Conflict`, the same way it already handled `400` ([#145])
- The missing API key error now says "Missing API Key" and names the correct environment variable, instead of "Missing Organization Name" ([#147])
- `stackguardian_runner_group_token` data source: corrected a configuration error message that referred to the wrong client type ([#147])

### Security

- Stop logging the full API key in provider debug output; it is now redacted ([#147])
- Mark connector credential attributes as sensitive in the `stackguardian_connector` resource and data source: `github_app_webhook_secret`, `github_app_client_secret`, `github_app_pem_file_content`, `gitlab_creds`, `azure_creds`, `bitbucket_creds`, `aws_access_key_id`, `aws_secret_access_key`, `arm_client_secret` and `gcp_config_file_content`. These were previously shown in plaintext in plan output ([#147])
- Mark the `stackguardian_runner_group_token` data source token as sensitive ([#147])
- Mark `webhook_secret` as sensitive on `stackguardian_workflow_git`, `stackguardian_workflow_from_template`, `stackguardian_stack_template_revision`, `stackguardian_workflow_template_revision` and their data sources ([#147])
- Mark the `text_value` environment variable field as sensitive on all resources and data sources that have it ([#147])
- Mark `azure_blob_storage_access_key` as sensitive on the `stackguardian_runner_group` resource and data source ([#147])
- Mark `docker_registry_username` as sensitive on `stackguardian_workflow_step_template`, `stackguardian_runner_group` and their data sources ([#147])
- `stackguardian_runner_group_token` data source: add a 30-second timeout to the API request, URL-encode the organization and runner group IDs in the request path, and stop including the raw response body in error messages ([#147])
- Cancelling a run (Ctrl+C) now cancels in-flight create requests for `stackguardian_workflow_git` and `stackguardian_runner_group` ([#147])
- Update `golang.org/x/crypto`, `golang.org/x/net`, `google.golang.org/grpc` and related dependencies to patched versions ([#145], [#146])

## [1.12.3] - 2026-09-24

### Changed

- **Breaking:** `stackguardian_workflow_template`: removed the `vcs_triggers` attribute from the resource and the data source ([#143])
- **Breaking:** Changing an identity attribute on an existing resource is now rejected at plan time with an error, instead of destroying and recreating the resource. This applies to `id`, `source_config_kind` and `runtime_source.config.repo` on `stackguardian_workflow_template`, and to `template_id`, `source_config_kind` and `runtime_source.config.repo` on `stackguardian_workflow_template_revision` ([#143])
- Attribute docs now state which `workflow_template` and `workflow_template_revision` attributes cannot change after creation, and that a published revision only accepts changes to `description`, `alias`, `notes` and `deprecation` ([#143])

### Fixed

- `stackguardian_workflow_template` and `stackguardian_workflow_template_revision`: optional attributes left out of the configuration are no longer sent to the API as `false` or `""`. Affects `runtime_source.config.is_private`, `runtime_source.config.git_core_auto_crlf`, `runtime_source.config.ref` and `runtime_source.config_dest_kind`, and every `terraform_config` attribute on the revision ([#143])
- `stackguardian_workflow_template_revision`: changes to `runtime_source.config.auth` are now sent on update; previously they had no effect ([#143])
- `stackguardian_workflow_template`: a configured `id` is now used on create, instead of being replaced by a server-generated ID and causing "Provider produced inconsistent result after apply" ([#143])
- An explicitly empty list attribute (`[]`) no longer turns into `null` after apply, which caused "Provider produced inconsistent result after apply" ([#143])
- `stackguardian_workflow_git` `tags` and `approvers`, and `stackguardian_workflow_template` `tags`, no longer show a perpetual diff after apply ([#143])

## [1.12.2] - 2026-09-17

### Added

- `stackguardian_workflow_template` and `stackguardian_workflow_template_revision`: plan-time validation of the `runtime_source` `auth` and `is_private` rules ([#132])
- `stackguardian_workflow_template_revision`: `wf_steps_config` is rejected at plan time when `source_config_kind` is `TERRAFORM` or `OPENTOFU`, which use built-in run steps ([#132])
- Every resource and data source attribute is now documented ([#117])
- New guides: Getting Started, Object Model, Templates, Policies, Access Management, Resource IDs, Runtime References, Importing Resources and Troubleshooting ([#117])

### Changed

- **Breaking:** `stackguardian_workflow_template_revision`: `input_schemas.type` is now required, to match the API ([#132])
- `stackguardian_workflow_template_revision`: `mount_points.read_only` is now computed when not set, to match the API ([#132])
- Provider log and error messages now spell the product name "StackGuardian"
- Built with Go 1.26.7 and `terraform-plugin-framework` v1.19.0

### Fixed

- `stackguardian_workflow_template_revision`: `description`, `notes`, `is_public`, `number_of_approvals_required` and `mount_points.read_only` are no longer sent to the API as zero values on create when left out of the configuration ([#132])

## Releases 0.2.0 to 1.12.1

These releases are not recorded in this file. See the
[GitHub releases page](https://github.com/StackGuardian/terraform-provider-stackguardian/releases)
for their notes.

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

[Unreleased]: https://github.com/StackGuardian/terraform-provider-stackguardian/compare/v1.12.4...HEAD
[1.12.4]: https://github.com/StackGuardian/terraform-provider-stackguardian/compare/v1.12.3...v1.12.4
[1.12.3]: https://github.com/StackGuardian/terraform-provider-stackguardian/compare/v1.12.2...v1.12.3
[1.12.2]: https://github.com/StackGuardian/terraform-provider-stackguardian/compare/v1.12.1...v1.12.2
[0.1.0]: https://github.com/StackGuardian/terraform-provider-stackguardian/compare/v0.1.0-rc4...v0.1.0
[0.1.0-rc4]: https://github.com/StackGuardian/terraform-provider-stackguardian/compare/v0.1.0-rc3...v0.1.0-rc4
[0.1.0-rc3]: https://github.com/StackGuardian/terraform-provider-stackguardian/compare/v0.1.0-rc2...v0.1.0-rc3
[0.1.0-rc2]: https://github.com/StackGuardian/terraform-provider-stackguardian/compare/v0.1.0-rc1...v0.1.0-rc2
[0.1.0-rc1]: https://github.com/StackGuardian/terraform-provider-stackguardian/compare/v0.1.0-beta1...v0.1.0-rc1
[0.1.0-beta1]: https://github.com/StackGuardian/terraform-provider-stackguardian/compare/v0.1.0-alpha1...v0.1.0-beta1
[0.1.0-alpha1]: https://github.com/StackGuardian/terraform-provider-stackguardian/releases/tag/v0.1.0-alpha1

[#117]: https://github.com/StackGuardian/terraform-provider-stackguardian/pull/117
[#132]: https://github.com/StackGuardian/terraform-provider-stackguardian/pull/132
[#134]: https://github.com/StackGuardian/terraform-provider-stackguardian/pull/134
[#143]: https://github.com/StackGuardian/terraform-provider-stackguardian/pull/143
[#145]: https://github.com/StackGuardian/terraform-provider-stackguardian/pull/145
[#146]: https://github.com/StackGuardian/terraform-provider-stackguardian/pull/146
[#147]: https://github.com/StackGuardian/terraform-provider-stackguardian/pull/147
[#151]: https://github.com/StackGuardian/terraform-provider-stackguardian/pull/151
