package stackoutputs

import (
	"context"
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	sgsdkgo "github.com/StackGuardian/sg-sdk-go"
)

type stackOutputsDataSourceModel struct {
	Stack         types.String `tfsdk:"stack"`
	WorkflowGroup types.String `tfsdk:"workflow_group"`
	DataJson      types.String `tfsdk:"data_json"`
	Data          types.Map    `tfsdk:"data"`
}

func buildAPIModelToTerraformModel(stackOutputs *sgsdkgo.GeneratedStackOutputsResponse) (*stackOutputsDataSourceModel, diag.Diagnostics) {
	stackOutputsDataSourceModel := stackOutputsDataSourceModel{}
	stackOutputsMap := stackOutputs.Data

	if stackOutputsMap == nil {
		stackOutputsDataSourceModel.Data = types.MapNull(types.StringType)
		stackOutputsDataSourceModel.DataJson = types.StringNull()
		return &stackOutputsDataSourceModel, nil
	}

	dataString, err := json.Marshal(stackOutputsMap)
	if err != nil {
		return nil, diag.Diagnostics{diag.NewErrorDiagnostic("Faile to convert stack outputs map to string", err.Error())}
	}
	stackOutputsDataSourceModel.DataJson = types.StringValue(string(dataString))

	dataMap := map[string]types.String{}
	for key, value := range stackOutputsMap {
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

	stackOutputsDataSourceModel.Data = dataTerraType

	return &stackOutputsDataSourceModel, nil
}
