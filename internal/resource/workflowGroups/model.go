package workflowGroups

import (
	"context"

	sgsdkgo "github.com/StackGuardian/sg-sdk-go"
	flatteners "github.com/StackGuardian/terraform-provider-stackguardian/internal/flattners"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type WorkflowGroupResourceModel struct {
	ResourceName types.String `tfsdk:"resource_name"`
	Description  types.String `tfsdk:"description"`
	Tags         types.List   `tfsdk:"tags"`
}

func (m *WorkflowGroupResourceModel) ToAPIModel(ctx context.Context) (*sgsdkgo.WorkflowGroup, diag.Diagnostics) {
	diag := diag.Diagnostics{}
	apiModel := sgsdkgo.WorkflowGroup{
		ResourceName: m.ResourceName.ValueStringPointer(),
		Description:  m.Description.ValueStringPointer(),
	}

	// Convert Tags from types.List to []string
	elements := make([]types.String, 0, len(m.Tags.Elements()))
	diags := m.Tags.ElementsAs(ctx, &elements, false)
	diag.Append(diags...)
	if diag.HasError() {
		return nil, diag
	}
	var tags []string
	for _, tag := range elements {
		tags = append(tags, tag.ValueString())
	}

	apiModel.Tags = tags

	return &apiModel, nil
}

func (m *WorkflowGroupResourceModel) ToPatchedAPIModel(ctx context.Context) (*sgsdkgo.PatchedWorkflowGroup, diag.Diagnostics) {
	diag := diag.Diagnostics{}
	apiModel := sgsdkgo.PatchedWorkflowGroup{
		Description: m.Description.ValueStringPointer(),
	}

	// Convert Tags from types.List to []string
	elements := make([]types.String, 0, len(m.Tags.Elements()))
	diags := m.Tags.ElementsAs(ctx, &elements, false)
	diag.Append(diags...)
	if diag.HasError() {
		return nil, diag
	}
	var tags []string
	for _, tag := range elements {
		tags = append(tags, tag.ValueString())
	}

	apiModel.Tags = tags

	return &apiModel, nil
}

func buildAPIModelToWorkflowGroupModel(apiResponse *sgsdkgo.WorkflowGroup) (*WorkflowGroupResourceModel, diag.Diagnostics) {
	diag := diag.Diagnostics{}
	WorkflowGroupModel := &WorkflowGroupResourceModel{
		ResourceName: flatteners.String(*apiResponse.ResourceName),
		Description:  flatteners.String(*apiResponse.Description),
	}

	// Convert Tags from []string to types.List
	var tags []attr.Value
	for _, tag := range apiResponse.Tags {
		tags = append(tags, flatteners.String(tag))
	}
	tagsList, diags := types.ListValueFrom(context.Background(), types.StringType, tags)
	diag.Append(diags...)
	if diag.HasError() {
		return nil, diag
	}

	WorkflowGroupModel.Tags = tagsList

	return WorkflowGroupModel, nil
}
