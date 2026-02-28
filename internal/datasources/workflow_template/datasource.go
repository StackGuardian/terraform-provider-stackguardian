package workflowtemplate

import (
	"context"
	"fmt"

	sgclient "github.com/StackGuardian/sg-sdk-go/client"
	"github.com/StackGuardian/terraform-provider-stackguardian/internal/customTypes"
	workflowtemplate "github.com/StackGuardian/terraform-provider-stackguardian/internal/resource/workflow_template"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

var (
	_ datasource.DataSource              = &workflowTemplateDataSource{}
	_ datasource.DataSourceWithConfigure = &workflowTemplateDataSource{}
)

// NewDataSource is a helper function to simplify the provider implementation.
func NewDataSource() datasource.DataSource {
	return &workflowTemplateDataSource{}
}

type workflowTemplateDataSource struct {
	client  *sgclient.Client
	orgName string
}

func (d *workflowTemplateDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workflow_template"
}

func (d *workflowTemplateDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *workflowTemplateDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config workflowtemplate.WorkflowTemplateResourceModel

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

	readResp, err := d.client.WorkflowTemplates.ReadWorkflowTemplate(ctx, d.orgName, templateID)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read workflow template.", err.Error())
		return
	}

	if readResp == nil {
		resp.Diagnostics.AddError("Error reading workflow template", "API response is empty")
		return
	}

	model, diags := workflowtemplate.BuildAPIModelToWorkflowTemplateModel(&readResp.Msg)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, model)...)
}
