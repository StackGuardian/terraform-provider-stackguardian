package stacktemplatedatasource

import (
	"context"
	"fmt"

	sgclient "github.com/StackGuardian/sg-sdk-go/client"
	"github.com/StackGuardian/terraform-provider-stackguardian/internal/constants"
	"github.com/StackGuardian/terraform-provider-stackguardian/internal/customTypes"
	stacktemplate "github.com/StackGuardian/terraform-provider-stackguardian/internal/resource/stack_template"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &stackTemplateDataSource{}
	_ datasource.DataSourceWithConfigure = &stackTemplateDataSource{}
)

// NewDataSource is a helper function to simplify the provider implementation.
func NewDataSource() datasource.DataSource {
	return &stackTemplateDataSource{}
}

type stackTemplateDataSource struct {
	client  *sgclient.Client
	orgName string
}

func (d *stackTemplateDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_stack_template"
}

func (d *stackTemplateDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	provInfo, ok := req.ProviderData.(*customTypes.ProviderInfo)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *customTypes.ProviderInfo, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = provInfo.Client
	d.orgName = provInfo.Org_name
}

func (d *stackTemplateDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "> **Note:** This data source is currently in **BETA**. Features and behavior may change.\n\nUse this data source to read a stack template.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: constants.DatasourceId,
				Required:            true,
			},
			"owner_org": schema.StringAttribute{
				MarkdownDescription: constants.StackTemplateOwnerOrg,
				Computed:            true,
			},
			"template_name": schema.StringAttribute{
				MarkdownDescription: constants.StackTemplateName,
				Computed:            true,
			},
			"source_config_kind": schema.StringAttribute{
				MarkdownDescription: constants.StackTemplateSourceConfigKindCommon,
				Computed:            true,
			},
			"is_active": schema.StringAttribute{
				MarkdownDescription: constants.StackTemplateIsActiveCommon,
				Computed:            true,
			},
			"is_public": schema.StringAttribute{
				MarkdownDescription: constants.StackTemplateIsPublicCommon,
				Computed:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: fmt.Sprintf(constants.Description, "stack template"),
				Computed:            true,
			},
			"tags": schema.ListAttribute{
				MarkdownDescription: fmt.Sprintf(constants.Tags, "stack template"),
				ElementType:         types.StringType,
				Computed:            true,
			},
			"context_tags": schema.MapAttribute{
				MarkdownDescription: fmt.Sprintf(constants.ContextTags, "stack template"),
				ElementType:         types.StringType,
				Computed:            true,
			},
			"shared_orgs_list": schema.ListAttribute{
				MarkdownDescription: constants.StackTemplateSharedOrgs,
				ElementType:         types.StringType,
				Computed:            true,
			},
		},
	}
}

func (d *stackTemplateDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config stacktemplate.StackTemplateResourceModel

	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	templateID := config.Id.ValueString()
	if templateID == "" {
		resp.Diagnostics.AddError("id must be provided", "")
		return
	}

	readResp, err := d.client.StackTemplates.ReadStackTemplate(ctx, d.orgName, templateID)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read stack template.", err.Error())
		return
	}

	if readResp == nil {
		resp.Diagnostics.AddError("Error reading stack template", "API response is empty")
		return
	}

	model, diags := stacktemplate.BuildAPIModelToStackTemplateModel(&readResp.Msg)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, model)...)
}
