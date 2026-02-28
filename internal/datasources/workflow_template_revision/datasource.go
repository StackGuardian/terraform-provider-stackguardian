package workflowtemplaterevision

import (
	"context"
	"fmt"

	sgclient "github.com/StackGuardian/sg-sdk-go/client"
	"github.com/StackGuardian/terraform-provider-stackguardian/internal/customTypes"
	workflowtemplaterevision "github.com/StackGuardian/terraform-provider-stackguardian/internal/resource/workflow_template_revision"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

var (
	_ datasource.DataSource              = &workflowTemplateRevisionDataSource{}
	_ datasource.DataSourceWithConfigure = &workflowTemplateRevisionDataSource{}
)

// NewDataSource is a helper function to simplify the provider implementation.
func NewDataSource() datasource.DataSource {
	return &workflowTemplateRevisionDataSource{}
}

type workflowTemplateRevisionDataSource struct {
	client  *sgclient.Client
	orgName string
}

func (d *workflowTemplateRevisionDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workflow_template_revision"
}

func (d *workflowTemplateRevisionDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *workflowTemplateRevisionDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config workflowtemplaterevision.WorkflowTemplateRevisionResourceModel

	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	revisionID := config.Id.ValueString()
	if revisionID == "" {
		resp.Diagnostics.AddError("id must be provided", "")
		return
	}

	readResp, err := d.client.WorkflowTemplatesRevisions.ReadWorkflowTemplateRevision(ctx, d.orgName, revisionID)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read workflow template revision.", err.Error())
		return
	}

	if readResp == nil {
		resp.Diagnostics.AddError("Error reading workflow template revision", "API response is empty")
		return
	}

	model, diags := workflowtemplaterevision.BuildAPIModelToWorkflowTemplateRevisionModel(ctx, &readResp.Msg)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Preserve template_id from config if provided
	if !config.TemplateId.IsNull() && !config.TemplateId.IsUnknown() {
		model.TemplateId = config.TemplateId
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, model)...)
}
