package expanders

import "encoding/json"

func JSONStringToInterface(s string) any {
	if s == "" {
		return nil
	}
	var v any
	_ = json.Unmarshal([]byte(s), &v)
	return v
}

func ParseJSONToMap(jsonStr string) map[string]interface{} {
	var result map[string]interface{}
	if jsonStr == "" {
		return result
	}
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return make(map[string]interface{})
	}
	return result
}
