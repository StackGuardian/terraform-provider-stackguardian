package workflowsteptemplaterevision

import (
	"context"
	"fmt"
	"strings"

	sgclient "github.com/StackGuardian/sg-sdk-go/client"
	"github.com/StackGuardian/terraform-provider-stackguardian/internal/customTypes"
	workflowsteptemplaterevisionresource "github.com/StackGuardian/terraform-provider-stackguardian/internal/resource/workflow_step_template_revision"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

var (
	_ datasource.DataSource              = &workflowStepTemplateRevisionDatasource{}
	_ datasource.DataSourceWithConfigure = &workflowStepTemplateRevisionDatasource{}
)

// NewDataSource is a helper function to simplify the provider implementation.
func NewDataSource() datasource.DataSource {
	return &workflowStepTemplateRevisionDatasource{}
}

type workflowStepTemplateRevisionDatasource struct {
	client  *sgclient.Client
	orgName string
}

func (d *workflowStepTemplateRevisionDatasource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workflow_step_template_revision"
}

func (d *workflowStepTemplateRevisionDatasource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *workflowStepTemplateRevisionDatasource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config workflowsteptemplaterevisionresource.WorkflowStepTemplateRevisionResourceModel

	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := config.Id.ValueString()
	if id == "" {
		resp.Diagnostics.AddError("id is required", "The id attribute must be provided in the format `templateId:revisionNumber`.")
		return
	}

	// Extract templateId from the ID (format: templateId:revisionNumber)
	templateId := ""
	parts := strings.SplitN(id, ":", 2)
	if len(parts) == 2 {
		templateId = parts[0]
	}

	readResp, err := d.client.WorkflowStepTemplateRevision.ReadWorkflowStepTemplateRevision(ctx, d.orgName, id)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read workflow step template revision.", err.Error())
		return
	}

	if readResp == nil || readResp.Msg == nil {
		resp.Diagnostics.AddError("Unable to read workflow step template revision.", "API response is empty")
		return
	}

	revisionModel, diags := workflowsteptemplaterevisionresource.BuildAPIModelToRevisionModel(readResp.Msg, id, templateId)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, revisionModel)...)
}
