package stackworkflowoutputs

import (
	"context"
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type stackWorkflowOutputsDataSourceModel struct {
	Stack         types.String `tfsdk:"stack"`
	Workflow      types.String `tfsdk:"workflow"`
	WorkflowGroup types.String `tfsdk:"workflow_group"`
	DataJson      types.String `tfsdk:"data_json"`
	Data          types.Map    `tfsdk:"data"`
}

func buildAPIModelToTerraformModel(stackWorkflowOutputs []byte) (*stackWorkflowOutputsDataSourceModel, diag.Diagnostics) {
	stackWorkflowOutputsDataSourceModel := stackWorkflowOutputsDataSourceModel{}

	if stackWorkflowOutputs == nil {
		stackWorkflowOutputsDataSourceModel.Data = types.MapNull(types.StringType)
		stackWorkflowOutputsDataSourceModel.DataJson = types.StringNull()
		return &stackWorkflowOutputsDataSourceModel, nil
	}

	stackWorkflowOutputsDataSourceModel.DataJson = types.StringValue(string(stackWorkflowOutputs))

	var stackWorkflowOutputsMap map[string]any
	err := json.Unmarshal(stackWorkflowOutputs, &stackWorkflowOutputsMap)
	if err != nil {
		return nil, diag.Diagnostics{diag.NewErrorDiagnostic("Stack outputs is not valid JSON", "")}
	}

	dataMap := map[string]types.String{}
	for key, value := range stackWorkflowOutputsMap {
		if value != nil {
			valueString, err := json.Marshal(value)
			if err != nil {
				return nil, diag.Diagnostics{diag.NewErrorDiagnostic("Stack outputs is not valid json", err.Error())}
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
