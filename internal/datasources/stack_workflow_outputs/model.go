package stackworkflowoutputs

import (
	"context"
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	sgsdkgo "github.com/StackGuardian/sg-sdk-go"
)

type stackWorkflowOutputsDataSourceModel struct {
	Stack         types.String `tfsdk:"stack"`
	Workflow      types.String `tfsdk:"workflow"`
	WorkflowGroup types.String `tfsdk:"workflow_group"`
	DataJson      types.String `tfsdk:"data_json"`
	Data          types.Map    `tfsdk:"data"`
}

func buildAPIModelToTerraformModel(stackOutputs *sgsdkgo.GeneratedWorkflowOutputsResponse) (*stackWorkflowOutputsDataSourceModel, diag.Diagnostics) {
	stackWorkflowOutputsDataSourceModel := stackWorkflowOutputsDataSourceModel{}
	stackWorkflowOutputsMap := stackOutputs.Data.Outputs

	if stackWorkflowOutputsMap == nil {
		stackWorkflowOutputsDataSourceModel.Data = types.MapNull(types.StringType)
		stackWorkflowOutputsDataSourceModel.DataJson = types.StringNull()
		return &stackWorkflowOutputsDataSourceModel, nil
	}

	dataString, err := json.Marshal(stackWorkflowOutputsMap)
	if err != nil {
		return nil, diag.Diagnostics{diag.NewErrorDiagnostic("Faile to convert stack outputs map to string", err.Error())}
	}
	stackWorkflowOutputsDataSourceModel.DataJson = types.StringValue(string(dataString))

	dataMap := map[string]types.String{}
	for key, value := range stackWorkflowOutputsMap {
		if value != nil {
			valueString, err := json.Marshal(value)
			if err != nil {
				return nil, diag.Diagnostics{diag.NewErrorDiagnostic("Fail to convert stack output to string", err.Error())}
			}

			dataMap[key] = types.StringValue(string(valueString))
		} else {
			dataMap[key] = types.StringNull()
		}
	}
	dataTerraType, diags := types.MapValueFrom(context.TODO(), types.StringType, &dataMap)
	if diags.HasError() {
		return nil, diags
	}

	stackWorkflowOutputsDataSourceModel.Data = dataTerraType

	return &stackWorkflowOutputsDataSourceModel, nil
}
