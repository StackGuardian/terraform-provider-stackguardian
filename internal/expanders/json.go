package expanders

import (
	"encoding/json"
	"fmt"
)

func JSONStringToInterface(s string) any {
	if s == "" {
		return nil
	}
	var v any
	_ = json.Unmarshal([]byte(s), &v)
	return v
}

func JSONStringToMap(jsonStr string) map[string]interface{} {
	var result map[string]interface{}
	if jsonStr == "" {
		return result
	}
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return make(map[string]interface{})
	}
	return result
}

// JSONStringToObject parses a JSON-object string attribute (such as
// iac_input_data.data) for the API. Unlike JSONStringToMap it returns the parse
// error instead of replacing invalid input with an empty object: a silent {}
// sent the user's inputs nowhere and only surfaced as "Provider produced
// inconsistent result after apply". "" means no data and returns nil. Valid
// JSON that isn't an object (an array, a string, null) is an error too, since
// the API expects an object.
func JSONStringToObject(s string) (map[string]interface{}, error) {
	if s == "" {
		return nil, nil
	}
	var v interface{}
	if err := json.Unmarshal([]byte(s), &v); err != nil {
		return nil, fmt.Errorf("not valid JSON: %w", err)
	}
	obj, ok := v.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("must be a JSON object, e.g. jsonencode({ key = \"value\" })")
	}
	return obj, nil
}
