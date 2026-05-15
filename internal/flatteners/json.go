package flatteners

import (
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

// IsEmptyObject reports whether v marshals to "{}".
// SDK structs use omitempty on every field, so a struct populated entirely with
// zero-values (what you get when the API returns "{}") marshals back to exactly
// two bytes. Use this in convertXxxFromAPI alongside the nil check:
//
//	if cfg == nil || flatteners.IsEmptyObject(cfg) {
//	    return nullObj, nil
//	}
func IsEmptyObject(v any) bool {
	b, err := json.Marshal(v)
	if err != nil {
		return false
	}
	return len(b) == 2 && b[0] == '{' && b[1] == '}'
}

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
