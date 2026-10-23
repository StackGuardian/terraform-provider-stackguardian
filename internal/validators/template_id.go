// Package validators holds schema validators shared across resources.
package validators

import (
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// Template revision ids are taken in the API's own fully qualified form,
// "/<owner-org>/<template-id>:<revision>", with the owning organization
// always spelled out, even for a template in the caller's own organization,
// and the revision always given (it must already exist).
//
// Why the org is mandatory rather than added by the provider: a template can
// be owned by another organization (shared one-to-one, published publicly, or
// StackGuardian's own marketplace templates under "stackguardian"), and the
// API identifies it only by the full id. A provider that qualifies a bare id
// with the caller's org can never reference such a template, and one that
// strips the org on read loses which org the template belongs to, so the next
// update silently points at a same-named template in the caller's org.
// Requiring the full id everywhere gives one canonical form that is sent
// and stored exactly as the API returns it, with no rewriting in either
// direction and no ambiguity about which form state holds.
var qualifiedTemplateRevisionIdRegexp = regexp.MustCompile(`^/[^/\s]+/[^/:\s]+:[^/:\s]+$`)

// QualifiedTemplateRevisionId validates "/<owner-org>/<template-id>:<revision>".
func QualifiedTemplateRevisionId() validator.String {
	return stringvalidator.RegexMatches(qualifiedTemplateRevisionIdRegexp,
		"must be the template revision's full id, /<org>/<template-id>:<revision> "+
			"(e.g. /my-org/my-template:1), including the organization even when it is your own")
}
