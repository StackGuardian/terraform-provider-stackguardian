package stacktemplaterevision_test

import (
	"context"
	"testing"

	stacktemplaterevision "github.com/StackGuardian/terraform-provider-stackguardian/internal/resource/stack_template_revision"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// v0State is state as provider 1.12.x wrote it: deployment_platform_config[*].config is a JSON
// string, and it carries attributes version 1 removed (template_id, iac_input_data,
// input_schemas[*].is_committed, iac_vcs_config.use_marketplace_template).
const v0State = `{
  "id": "my-stack-template:1",
  "parent_template_id": "my-stack-template",
  "workflows_config": {
    "workflows": [
      {
        "id": "d8dfaf15-2ad9-da29-8af0-c6b288b12089",
        "template_id": "my-workflow-template",
        "iac_input_data": {"schema_type": "RAW_JSON", "data": "{}"},
        "input_schemas": [{"type": "FORM_JSONSCHEMA", "is_committed": true}],
        "vcs_config": {"iac_vcs_config": {"use_marketplace_template": true, "iac_template_id": "my-workflow-template"}},
        "deployment_platform_config": [
          {"kind": "AWS_RBAC", "config": "{\"integrationId\":\"/integrations/aws-prod\",\"profileName\":\"default\"}"},
          {"kind": "AWS_RBAC", "config": "{\"integrationId\":\"/integrations/aws-dev\"}"},
          {"kind": "AWS_RBAC", "config": "not json"},
          {"kind": "AWS_RBAC", "config": null}
        ]
      }
    ]
  },
  "actions": {
    "apply": {
      "name": "apply",
      "order": {
        "d8dfaf15-2ad9-da29-8af0-c6b288b12089": {
          "parameters": {
            "deployment_platform_config": [
              {"kind": "AWS_OIDC", "config": "{\"integrationId\":\"/integrations/aws-oidc\"}"}
            ]
          }
        }
      }
    }
  }
}`

func TestStackTemplateRevisionUpgradeStateV0(t *testing.T) {
	ctx := context.Background()

	r, ok := stacktemplaterevision.NewResource().(resource.ResourceWithUpgradeState)
	if !ok {
		t.Fatal("stack_template_revision must implement resource.ResourceWithUpgradeState")
	}

	var schemaResp resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &schemaResp)
	if got := schemaResp.Schema.Version; got != 1 {
		t.Fatalf("schema version = %d, want 1", got)
	}

	upgrader, ok := r.UpgradeState(ctx)[0]
	if !ok {
		t.Fatal("no state upgrader registered for version 0")
	}

	var resp resource.UpgradeStateResponse
	upgrader.StateUpgrader(ctx, resource.UpgradeStateRequest{
		RawState: &tfprotov6.RawState{JSON: []byte(v0State)},
	}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("upgrade failed: %v", resp.Diagnostics)
	}
	if resp.DynamicValue == nil {
		t.Fatal("upgrade did not set DynamicValue")
	}

	stateType := schemaResp.Schema.Type().TerraformType(ctx)
	value, err := resp.DynamicValue.Unmarshal(stateType)
	if err != nil {
		t.Fatalf("upgraded state does not decode against the current schema: %s", err)
	}

	wfDpc := tftypes.NewAttributePath().WithAttributeName("workflows_config").WithAttributeName("workflows").
		WithElementKeyInt(0).WithAttributeName("deployment_platform_config")
	actionDpc := tftypes.NewAttributePath().WithAttributeName("actions").WithElementKeyString("apply").
		WithAttributeName("order").WithElementKeyString("d8dfaf15-2ad9-da29-8af0-c6b288b12089").
		WithAttributeName("parameters").WithAttributeName("deployment_platform_config")

	for _, tc := range []struct {
		name               string
		config             *tftypes.AttributePath
		wantNull           bool
		wantIntegrationId  string
		wantProfileName    string
		wantProfileNameSet bool
	}{
		{"both fields", wfDpc.WithElementKeyInt(0).WithAttributeName("config"), false, "/integrations/aws-prod", "default", true},
		{"integration only", wfDpc.WithElementKeyInt(1).WithAttributeName("config"), false, "/integrations/aws-dev", "", false},
		{"unparseable string", wfDpc.WithElementKeyInt(2).WithAttributeName("config"), true, "", "", false},
		{"null", wfDpc.WithElementKeyInt(3).WithAttributeName("config"), true, "", "", false},
		{"actions parameters", actionDpc.WithElementKeyInt(0).WithAttributeName("config"), false, "/integrations/aws-oidc", "", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			raw, _, err := tftypes.WalkAttributePath(value, tc.config)
			if err != nil {
				t.Fatalf("walk %s: %s", tc.config, err)
			}
			cfg := raw.(tftypes.Value)
			if tc.wantNull {
				if !cfg.IsNull() {
					t.Fatalf("config = %s, want null", cfg)
				}
				return
			}
			var attrs map[string]tftypes.Value
			if err := cfg.As(&attrs); err != nil {
				t.Fatalf("config is not an object: %s", err)
			}
			var integrationId string
			if err := attrs["integration_id"].As(&integrationId); err != nil || integrationId != tc.wantIntegrationId {
				t.Errorf("integration_id = %q (err %v), want %q", integrationId, err, tc.wantIntegrationId)
			}
			if !tc.wantProfileNameSet {
				if !attrs["profile_name"].IsNull() {
					t.Errorf("profile_name = %s, want null", attrs["profile_name"])
				}
				return
			}
			var profileName string
			if err := attrs["profile_name"].As(&profileName); err != nil || profileName != tc.wantProfileName {
				t.Errorf("profile_name = %q (err %v), want %q", profileName, err, tc.wantProfileName)
			}
		})
	}
}
