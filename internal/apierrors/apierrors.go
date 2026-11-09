// Package apierrors classifies errors returned by sg-sdk-go.
package apierrors

import (
	"errors"
	"strings"

	"github.com/StackGuardian/sg-sdk-go/core"
)

// IsNotFound reports whether err says the object does not exist.
//
// Most APIs answer with a 404. The template APIs (stack, workflow and workflow
// step templates and their revisions) answer a missing template with
// 400 {"msg":"Template ... does not exist"}, and deleting a missing workflow
// returns 400 {"msg":"Workflow does not exist"}, so a 400 carrying
// "does not exist" counts as well.
func IsNotFound(err error) bool {
	var apiErr *core.APIError
	if !errors.As(err, &apiErr) {
		return false
	}
	return apiErr.StatusCode == 404 ||
		(apiErr.StatusCode == 400 && strings.Contains(err.Error(), "does not exist"))
}
