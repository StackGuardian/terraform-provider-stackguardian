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

// JSONStringToMapStrict parses jsonStr as a JSON object. Unlike JSONStringToMap
// it returns the parse error instead of an empty map, so a malformed value is
// reported rather than sent to the API as {}. An empty string or JSON null
// yields a nil map.
func JSONStringToMapStrict(jsonStr string) (map[string]interface{}, error) {
	if jsonStr == "" {
		return nil, nil
	}
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, err
	}
	return result, nil
}
