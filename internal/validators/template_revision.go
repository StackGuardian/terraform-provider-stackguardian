package validators

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var _ validator.String = unpinnedTemplateRevisionValidator{}

type unpinnedTemplateRevisionValidator struct{}

// WarnUnpinnedTemplateRevision warns, without failing the plan, when a template
// id names a template but no revision ("my-template" rather than
// "my-template:1"). StackGuardian then reads the template itself, which holds
// none of a revision's content. Null and unknown values are not checked.
func WarnUnpinnedTemplateRevision() validator.String {
	return unpinnedTemplateRevisionValidator{}
}

func (v unpinnedTemplateRevisionValidator) Description(_ context.Context) string {
	return "should name a revision, as <template-name>:<revision> or <template-name>:latest"
}

func (v unpinnedTemplateRevisionValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v unpinnedTemplateRevisionValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	value := req.ConfigValue.ValueString()
	if value == "" || HasTemplateRevision(value) {
		return
	}
	resp.Diagnostics.AddAttributeWarning(req.Path, "Workflow template has no revision",
		"\""+value+"\" names a workflow template but no revision. StackGuardian then uses the template itself, "+
			"which holds none of a revision's content: stacks created from this stack template revision get no "+
			"inputs, Terraform settings or other defaults from it. Name a revision instead, for example \""+value+
			":1\" to pin one, or \""+value+":latest\" to use the newest.")
}

// HasTemplateRevision reports whether a template id ("<name>:<revision>", optionally
// "/<org>/"-qualified) names a revision, an alias or "latest".
func HasTemplateRevision(id string) bool {
	name := id[strings.LastIndex(id, "/")+1:]
	_, revision, found := strings.Cut(name, ":")
	return found && revision != ""
}
