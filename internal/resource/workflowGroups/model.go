package workflowGroups

import (
	"context"

	sgsdkgo "github.com/StackGuardian/sg-sdk-go"
	"github.com/StackGuardian/terraform-provider-stackguardian/internal/expanders"
	"github.com/StackGuardian/terraform-provider-stackguardian/internal/flatteners"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type WorkflowGroupResourceModel struct {
	ResourceName types.String `tfsdk:"resource_name"`
	Description  types.String `tfsdk:"description"`
	Tags         types.List   `tfsdk:"tags"`
}

func (m *WorkflowGroupResourceModel) ToAPIModel(ctx context.Context) (*sgsdkgo.WorkflowGroup, diag.Diagnostics) {
	apiModel := sgsdkgo.WorkflowGroup{
		ResourceName: m.ResourceName.ValueStringPointer(),
		Description:  m.Description.ValueStringPointer(),
	}

	if !m.Tags.IsUnknown() {
		tags, diags := expanders.StringList(context.TODO(), m.Tags)
		if diags.HasError() {
			return nil, diags
		}
		apiModel.Tags = tags
	}

	return &apiModel, nil
}

func (m *WorkflowGroupResourceModel) ToPatchedAPIModel(ctx context.Context) (*sgsdkgo.PatchedWorkflowGroup, diag.Diagnostics) {
	apiModel := sgsdkgo.PatchedWorkflowGroup{
		ResourceName: sgsdkgo.Optional(m.ResourceName.ValueString()),
	}

	if m.Description.IsUnknown() {
		apiModel.Description = sgsdkgo.Null[string]()
	}

	// Convert Tags from types.List to []string
	if m.Tags.IsUnknown() {
		tags, diags := expanders.StringList(context.TODO(), m.Tags)
		if diags.HasError() {
			return nil, diags
		}
		apiModel.Tags = sgsdkgo.Optional(tags)
	} else {
		apiModel.Tags = sgsdkgo.Null[[]string]()
	}

	return &apiModel, nil
}

func buildAPIModelToWorkflowGroupModel(apiResponse *sgsdkgo.WorkflowGroupDataResponse) (*WorkflowGroupResourceModel, diag.Diagnostics) {
	diag := diag.Diagnostics{}
	WorkflowGroupModel := &WorkflowGroupResourceModel{
		ResourceName: flatteners.String(*apiResponse.ResourceName),
		Description:  flatteners.String(*apiResponse.Description),
	}

	// Convert Tags from []string to types.List
	if apiResponse.Tags != nil {
		var tags []types.String
		for _, tag := range apiResponse.Tags {
			tags = append(tags, flatteners.String(tag))
		}
		tagsList, diags := types.ListValueFrom(context.Background(), types.StringType, tags)
		diag.Append(diags...)
		if diag.HasError() {
			return nil, diag
		}

		WorkflowGroupModel.Tags = tagsList
	} else {
		WorkflowGroupModel.Tags = types.ListNull(types.StringType)
	}

	return WorkflowGroupModel, nil
}
