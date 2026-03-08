package stacktemplate

import (
	"context"
	"fmt"

	"github.com/StackGuardian/terraform-provider-stackguardian/internal/constants"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Schema defines the schema for the resource.
func (r *stackTemplateResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "> **Note:** This resource is currently in **BETA**. Features and behavior may change.\n\nManages a stack template resource.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: constants.Id,
				Computed:            true,
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"owner_org": schema.StringAttribute{
				MarkdownDescription: constants.StackTemplateOwnerOrg,
				Computed:            true,
			},
			"template_name": schema.StringAttribute{
				MarkdownDescription: constants.StackTemplateName,
				Required:            true,
			},
			"source_config_kind": schema.StringAttribute{
				MarkdownDescription: constants.StackTemplateSourceConfigKindCommon,
				Required:            true,
			},
			"is_active": schema.StringAttribute{
				MarkdownDescription: constants.StackTemplateIsActiveCommon,
				Optional:            true,
				Computed:            true,
			},
			"is_public": schema.StringAttribute{
				MarkdownDescription: constants.StackTemplateIsPublicCommon,
				Optional:            true,
				Computed:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: fmt.Sprintf(constants.Description, "stack template"),
				Optional:            true,
				Computed:            true,
			},
			"tags": schema.ListAttribute{
				MarkdownDescription: fmt.Sprintf(constants.Tags, "stack template"),
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
			},
			"context_tags": schema.MapAttribute{
				MarkdownDescription: fmt.Sprintf(constants.ContextTags, "stack template"),
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
			},
			"shared_orgs_list": schema.ListAttribute{
				MarkdownDescription: constants.StackTemplateSharedOrgs,
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
			},
		},
	}
}
