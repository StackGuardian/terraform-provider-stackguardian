package flatteners

import (
	"encoding/json"
	"reflect"

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

// isNil reports whether v is nil, including a typed nil — a nil map, pointer, slice,
// interface, func or chan wrapped in an interface{} is not == nil, and json.Marshal would
// encode it as the string "null" rather than leaving the value unset.
func isNil(v interface{}) bool {
	if v == nil {
		return true
	}
	switch rv := reflect.ValueOf(v); rv.Kind() {
	case reflect.Map, reflect.Ptr, reflect.Slice, reflect.Interface, reflect.Func, reflect.Chan:
		return rv.IsNil()
	}
	return false
}

func JSONInterfaceToString(v interface{}) types.String {
	if isNil(v) {
		return types.StringNull()
	}
	b, err := json.Marshal(v)
	if err != nil {
		return types.StringNull()
	}
	return types.StringValue(string(b))
}

func JSONInterfaceToStringDefault(v interface{}) types.String {
	if isNil(v) {
		return types.StringValue("")
	}
	b, err := json.Marshal(v)
	if err != nil {
		return types.StringNull()
	}
	return types.StringValue(string(b))
}
