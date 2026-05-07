package flatteners

import (
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func JSONInterfaceToString(v interface{}) types.String {
	if v == nil {
		return types.StringNull()
	}
	b, err := json.Marshal(v)
	if err != nil {
		return types.StringNull()
	}
	return types.StringValue(string(b))
}

func JSONInterfaceToStringDefault(v interface{}) types.String {
	if v == nil {
		return types.StringValue("")
	}
	b, err := json.Marshal(v)
	if err != nil {
		return types.StringNull()
	}
	return types.StringValue(string(b))
}
