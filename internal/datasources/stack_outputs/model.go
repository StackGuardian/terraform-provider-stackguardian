package stackoutputs

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

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

	dataMap := map[string]*string{}
	for key, value := range stackOutputsMap {
		if value != "" {
			reqResp, err := http.Get(value)
			if err != nil {
				return nil, diag.Diagnostics{diag.NewErrorDiagnostic("Failed to fetch stack outputs", "")}
			}
			defer reqResp.Body.Close()

			body, err := io.ReadAll(reqResp.Body)
			if err != nil {
				return nil, diag.Diagnostics{diag.NewErrorDiagnostic("Failed to parse stack outputs", "")}
			}

			bodyStr := string(body)

			dataMap[key] = &bodyStr
		} else {
			dataMap[key] = nil
		}
	}
	dataTerraType, diags := types.MapValueFrom(context.TODO(), types.StringType, &dataMap)
	if diags.HasError() {
		return nil, diags
	}

	stackOutputsDataSourceModel.Data = dataTerraType

	dataMapJson := map[string]any{}
	for workflow, outputs := range dataMap {
		var jsonOutputs any
		err := json.Unmarshal([]byte(*outputs), &jsonOutputs)
		if err != nil {
			return nil, diag.Diagnostics{diag.NewErrorDiagnostic("Failed to parse stack outputs", "")}
		}

		dataMapJson[workflow] = jsonOutputs
	}

	dataString, err := json.Marshal(dataMapJson)
	if err != nil {
		return nil, diag.Diagnostics{diag.NewErrorDiagnostic("Failed to read stack outputs", err.Error())}
	}
	stackOutputsDataSourceModel.DataJson = types.StringValue(string(dataString))

	return &stackOutputsDataSourceModel, nil
}
