package stack

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// actionDefaultModifier plans actions[*].default, the platform's flag that tells
// a stack template action (default=true) from a custom one (default=false). In
// core, an upgrade that does not send Actions (e.g. from the UI) keeps custom
// actions and replaces or drops template ones (merge_actions_for_upgrade), so
// the provider must re-send each action with the right flag rather than stamping
// every action as custom.
//
//   - actions declared in config: the user owns them, so default=false.
//   - actions not declared: keep the state value (true for actions the platform
//     generated or took from the template, false for actions the user declared
//     earlier). A null state value (state written by an older provider version,
//     with no refresh) plans as unknown, since the real value is only known
//     after apply.
//
// On a template_group_id change with actions not declared, ModifyPlan
// (reResolveOnRevisionChange) replaces the whole map with the new revision's
// actions, flags included, after this modifier runs.
//
// Neither UseStateForUnknown alone nor a schema Default fits: both would plan
// the state's true for an action the user now declares under the same name
// (e.g. "apply"), while a declared action must be sent as false.
type actionDefaultModifier struct{}

func (actionDefaultModifier) Description(_ context.Context) string {
	return "false when actions is declared in config, otherwise the prior state value."
}

func (m actionDefaultModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (actionDefaultModifier) PlanModifyBool(ctx context.Context, req planmodifier.BoolRequest, resp *planmodifier.BoolResponse) {
	var actions types.Map
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("actions"), &actions)...)
	if resp.Diagnostics.HasError() || actions.IsUnknown() {
		return
	}

	if !actions.IsNull() {
		resp.PlanValue = types.BoolValue(false)
		return
	}

	if !req.StateValue.IsNull() {
		resp.PlanValue = req.StateValue
		return
	}
	resp.PlanValue = types.BoolUnknown()
}
