package workflowsteptemplate

import (
	"context"
	"fmt"

	sgclient "github.com/StackGuardian/sg-sdk-go/client"
	"github.com/StackGuardian/terraform-provider-stackguardian/internal/customTypes"
	workflowsteptemplateresource "github.com/StackGuardian/terraform-provider-stackguardian/internal/resource/workflow_step_template"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

var (
	_ datasource.DataSource              = &workflowStepTemplateDatasource{}
	_ datasource.DataSourceWithConfigure = &workflowStepTemplateDatasource{}
)

// NewDataSource is a helper function to simplify the provider implementation.
func NewDataSource() datasource.DataSource {
	return &workflowStepTemplateDatasource{}
}

type workflowStepTemplateDatasource struct {
	client  *sgclient.Client
	orgName string
}

func (d *workflowStepTemplateDatasource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workflow_step_template"
}

func (d *workflowStepTemplateDatasource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *workflowStepTemplateDatasource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config workflowsteptemplateresource.WorkflowStepTemplateResourceModel

	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := config.Id.ValueString()
	if id == "" {
		resp.Diagnostics.AddError("id is required", "The id attribute must be provided to look up a workflow step template.")
		return
	}

	readResp, err := d.client.WorkflowStepTemplate.ReadWorkflowStepTemplate(ctx, d.orgName, id)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read workflow step template.", err.Error())
		return
	}

	templateModel, diags := workflowsteptemplateresource.BuildAPIModelToWorkflowStepTemplateModel(&readResp.Msg)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, templateModel)...)
}
