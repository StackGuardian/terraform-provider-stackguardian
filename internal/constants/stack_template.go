package constants

// Stack Template - Common documentation
const (
	StackTemplateSourceConfigKindCommon = "What the stack contains. Use `MIXED` — a stack groups workflows built from different tools, and the tool used by each individual workflow is recorded on that workflow rather than here. The other values (`TERRAFORM`, `OPENTOFU`, `ANSIBLE_PLAYBOOK`, `HELM`, `KUBECTL`, `CLOUDFORMATION`, `CUSTOM`) belong on a `stackguardian_workflow_template`, which describes a single workflow."

	StackTemplateIsPublicCommon = `Whether the stack template is publicly available. Valid values:
	<span style="background-color: #eff0f0; color: #e53835;">0</span> (false),
	<span style="background-color: #eff0f0; color: #e53835;">1</span> (true)`
)

// Stack Template Resource documentation
const (
	StackTemplateName        = "Name of the stack template. Must be less than 100 characters."
	StackTemplateOwnerOrg    = "Organization that owns the stack template."
	StackTemplateSharedOrgs  = "List of organization IDs with which this template is shared."
	StackTemplateDescription = "A brief description of the stack template."
	StackTemplateTags        = "A list of tags associated with the stack template."
	StackTemplateContextTags = "Contextual key-value tags that provide additional context to the main tags."
)

// Stack Template Revision Resource documentation
const (
	StackTemplateRevisionId              = "Identifier of the revision, in the form `<template-name>:<revision>` (e.g. `my-stack-template:1`)."
	StackTemplateRevisionTemplateId      = "The revision's `id` qualified with its owning organization, in the form `/<org>/<template-name>:<revision>` (e.g. `/my-org/my-stack-template:1`)."
	StackTemplateRevisionIacTemplateId   = "Workflow template this stack workflow is created from, as the full id of an existing revision, including the owning organization: `/<org>/<template-id>:<revision>` (e.g. `/my-org/my-workflow-template:1`). Give the organization even for a template in your own organization; a template shared by another organization, or StackGuardian's own (`/stackguardian/…`), is referenced the same way."
	StackTemplateRevisionAlias           = "Human-readable alias for the revision (e.g., `v1.0.0`)."
	StackTemplateRevisionNotes           = "Release notes or changelog for this revision."
	StackTemplateRevisionDescription     = "Long description for the stack template revision."
	StackTemplateRevisionTags            = "A list of tags associated with the revision."
	StackTemplateRevisionContextTags     = "Contextual key-value tags for the revision."
	StackTemplateRevisionWorkflowsConfig = "JSON-encoded workflows configuration for the stack template revision."
	StackTemplateRevisionActions         = "JSON-encoded map of actions for the stack template revision."
)

// Stack Resource documentation
const (
	StackTemplateGroupId = "Stack template revision this stack is created from, as its full id including the owning organization: `/<org>/<template-id>:<revision>` (e.g. `/my-org/my-stack-template:1`) — the stack template revision's `template_id`. Give the organization even for a template in your own organization; a stack template shared by another organization is referenced the same way. Change the revision to upgrade the stack."
)
