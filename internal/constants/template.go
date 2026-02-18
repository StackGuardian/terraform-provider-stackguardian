package constants

// common template attributes
const (
	TemplateRevisionAlias                    string = "Alias for the template revision"
	TemplateRevisionNotes                    string = "Notes for the revision"
	TemplateRevisionIsPublic                 string = `Whether a revision is published to be used. Options: <span style="background-color: #eff0f0; color: #e53835;">\"1\"</span>, <span style="background-color: #eff0f0; color: #e53835;">\"0\"</span>`
	TemplateRevisionDeprecation              string = "Marking a template revision for deprecation"
	TemplateRevisionDeprecationEffectiveDate string = "Effective date for after which revision will be deprecated"
	TemplateRevisionDeprecationMessage       string = "Deprecation message"
	DeprecationMessage                       string = "Deprecation message"
)

// WorkflowTemplateAttributes
const (
	TemplateRevisionSourceConfigKind string = `Kind of Workflow Template. Options: <span style="background-color: #eff0f0; color: #e53835;">TERRAFORM</span>,
		<span style="background-color: #eff0f0; color: #e53835;">OPENTOFU</span>,<span style="background-color: #eff0f0; color: #e53835;">ANSIBLE_PLAYBOOK</span>,<span style="background-color: #eff0f0; color: #e53835;">HELM</span>,<span style="background-color: #eff0f0; color: #e53835;">KUBECTL</span>,<span style="background-color: #eff0f0; color: #e53835;">CLOUDFORMATION</span>,<span style="background-color: #eff0f0; color: #e53835;">CUSTOM</span>
	`
)
