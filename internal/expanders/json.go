package expanders

import (
	"encoding/json"
	"fmt"
)

// JSONStringToObject parses a JSON-object string attribute (iac_input_data.data,
// wf_step_input_data.data, workflow_run_payload, stack_run_payload) for the API.
// It returns the parse error instead of replacing invalid input with an empty
// object, as the helpers it replaced did: a silent {} sent the user's inputs
// nowhere and only surfaced as "Provider produced inconsistent result after
// apply". "" means no data and returns nil. Valid JSON that isn't an object (an
// array, a string, null) is an error too, since the API expects an object.
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
