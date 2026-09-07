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
	StackTemplateRevisionTemplateId      = "Parent stack template, as its bare `template_name` (which is also its `id`) — not a path."
	StackTemplateRevisionIacTemplateId   = "Workflow template this stack workflow is created from, as the bare `template_name` of a template in your own organization (e.g. `my-workflow-template`). The provider qualifies it with your organization, so do not give the `/<org>/…` form."
	StackTemplateRevisionAlias           = "Human-readable alias for the revision (e.g., `v1.0.0`)."
	StackTemplateRevisionNotes           = "Release notes or changelog for this revision."
	StackTemplateRevisionDescription     = "Long description for the stack template revision."
	StackTemplateRevisionTags            = "A list of tags associated with the revision."
	StackTemplateRevisionContextTags     = "Contextual key-value tags for the revision."
	StackTemplateRevisionWorkflowsConfig = "JSON-encoded workflows configuration for the stack template revision."
	StackTemplateRevisionActions         = "JSON-encoded map of actions for the stack template revision."
)
