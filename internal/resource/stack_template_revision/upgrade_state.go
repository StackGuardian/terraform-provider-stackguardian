package stacktemplaterevision

import (
	"context"
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

var _ resource.ResourceWithUpgradeState = &stackTemplateRevisionResource{}

// UpgradeState migrates state written before schema version 1 (provider <= 1.12.x).
//
// The only type change is deployment_platform_config[*].config, which was a JSON string and is
// now an object, in both workflows_config.workflows[*] and actions[*].order[*].parameters.
// Attributes removed in version 1 need no handling: the framework ignores attributes in raw
// state that the current schema no longer defines.
//
// No PriorSchema is given, so the version 0 schema does not have to be kept around: the raw
// JSON is rewritten in place and decoded against the current schema. Terraform refreshes the
// resource after the upgrade, so the converted config only has to be decodable, not exact.
func (r *stackTemplateRevisionResource) UpgradeState(_ context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{
		0: {StateUpgrader: r.upgradeStateV0},
	}
}

func (r *stackTemplateRevisionResource) upgradeStateV0(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
	if req.RawState == nil || req.RawState.JSON == nil {
		resp.Diagnostics.AddError("Unable to upgrade stack template revision state", "The previous state is missing or not in JSON form.")
		return
	}

	var state map[string]any
	if err := json.Unmarshal(req.RawState.JSON, &state); err != nil {
		resp.Diagnostics.AddError("Unable to upgrade stack template revision state", "Could not parse the previous state: "+err.Error())
		return
	}

	if wfc, ok := state["workflows_config"].(map[string]any); ok {
		if workflows, ok := wfc["workflows"].([]any); ok {
			for _, wf := range workflows {
				if wfMap, ok := wf.(map[string]any); ok {
					upgradeDeploymentPlatformConfigListV0(wfMap)
				}
			}
		}
	}
	if actions, ok := state["actions"].(map[string]any); ok {
		for _, action := range actions {
			actionMap, ok := action.(map[string]any)
			if !ok {
				continue
			}
			order, ok := actionMap["order"].(map[string]any)
			if !ok {
				continue
			}
			for _, entry := range order {
				entryMap, ok := entry.(map[string]any)
				if !ok {
					continue
				}
				if params, ok := entryMap["parameters"].(map[string]any); ok {
					upgradeDeploymentPlatformConfigListV0(params)
				}
			}
		}
	}

	upgraded, err := json.Marshal(state)
	if err != nil {
		resp.Diagnostics.AddError("Unable to upgrade stack template revision state", "Could not encode the upgraded state: "+err.Error())
		return
	}

	var schemaResp resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &schemaResp)
	if schemaResp.Diagnostics.HasError() {
		resp.Diagnostics.Append(schemaResp.Diagnostics...)
		return
	}
	stateType := schemaResp.Schema.Type().TerraformType(ctx)

	value, err := tftypes.ValueFromJSONWithOpts(upgraded, stateType, tftypes.ValueFromJSONOpts{IgnoreUndefinedAttributes: true})
	if err != nil {
		resp.Diagnostics.AddError("Unable to upgrade stack template revision state", "Could not decode the upgraded state: "+err.Error())
		return
	}
	dynamicValue, err := tfprotov6.NewDynamicValue(stateType, value)
	if err != nil {
		resp.Diagnostics.AddError("Unable to upgrade stack template revision state", "Could not build the upgraded state: "+err.Error())
		return
	}
	resp.DynamicValue = &dynamicValue
}

// upgradeDeploymentPlatformConfigListV0 rewrites each deployment_platform_config[*].config held by
// parent from the version 0 JSON string to the version 1 object.
func upgradeDeploymentPlatformConfigListV0(parent map[string]any) {
	list, ok := parent["deployment_platform_config"].([]any)
	if !ok {
		return
	}
	for _, item := range list {
		if itemMap, ok := item.(map[string]any); ok {
			itemMap["config"] = deploymentPlatformConfigObjectV0(itemMap["config"])
		}
	}
}

// deploymentPlatformConfigObjectV0 converts a version 0 config value — the API's config object
// marshalled to a JSON string, e.g. {"integrationId":"/integrations/aws","profileName":"p"} — to
// the version 1 {integration_id, profile_name} object. An empty or unparseable string becomes
// null; the refresh that follows the upgrade fills it from the API.
func deploymentPlatformConfigObjectV0(v any) any {
	if obj, ok := v.(map[string]any); ok {
		return obj // already in the version 1 shape
	}
	s, ok := v.(string)
	if !ok || s == "" {
		return nil
	}
	var cfg map[string]any
	if err := json.Unmarshal([]byte(s), &cfg); err != nil {
		return nil
	}
	return map[string]any{
		"integration_id": firstString(cfg, "integrationId", "integration_id"),
		"profile_name":   firstString(cfg, "profileName", "profile_name"),
	}
}

func firstString(m map[string]any, keys ...string) any {
	for _, k := range keys {
		if s, ok := m[k].(string); ok {
			return s
		}
	}
	return nil
}
