package workflowoutputs

import (
	"context"
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type workflowOutputsDataSourceModel struct {
	Workflow      types.String `tfsdk:"workflow"`
	WorkflowGroup types.String `tfsdk:"workflow_group"`
	DataJson      types.String `tfsdk:"data_json"`
	Data          types.Map    `tfsdk:"data"`
}

func buildAPIModelToTerraformModel(workflowOutputs []byte) (*workflowOutputsDataSourceModel, diag.Diagnostics) {
	workflowOutputsDataSourceModel := workflowOutputsDataSourceModel{}

	if workflowOutputs == nil {
		workflowOutputsDataSourceModel.Data = types.MapNull(types.StringType)
		workflowOutputsDataSourceModel.DataJson = types.StringNull()
		return &workflowOutputsDataSourceModel, nil
	}

	workflowOutputsDataSourceModel.DataJson = types.StringValue(string(workflowOutputs))

	var workflowOutputsMap map[string]any
	err := json.Unmarshal(workflowOutputs, &workflowOutputsMap)
	if err != nil {
		return nil, diag.Diagnostics{diag.NewErrorDiagnostic("Workflow output is not valid JSON", "")}
	}

	dataMap := map[string]types.String{}
	for key, value := range workflowOutputsMap {
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

	workflowOutputsDataSourceModel.Data = dataTerraType

	return &workflowOutputsDataSourceModel, nil
}
