package acctest

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	sgsdkgo "github.com/StackGuardian/sg-sdk-go"
	sgclient "github.com/StackGuardian/sg-sdk-go/client"
	"github.com/StackGuardian/sg-sdk-go/stacktemplaterevisions"
	"github.com/StackGuardian/sg-sdk-go/workflowsteptemplaterevision"
	"github.com/StackGuardian/sg-sdk-go/workflowtemplaterevisions"
)

// Sweeping deletes what an earlier run created and failed to remove.
//
// Unique names stop a leftover from blocking the next run, but they do not make
// it go away: a run killed between create and destroy leaves its resources
// behind for good, and the organisation fills with residue nobody can attribute.
// Every identifier these tests create carries NamePrefix, and a sweep deletes
// exactly what carries it.
//
// The prefix check is the only thing standing between this and real
// infrastructure, so every candidate goes through consider() and there is no way
// to opt out of it. SweepDryRun defaults to on, so a first or accidental
// invocation reports what it would remove instead of removing it.
//
// Not everything can be swept. The SDK's ListAllPolicies and ListAllRoles return
// only an error and discard the response body, and RunnerGroups has no list
// method at all, so those three cannot be enumerated and are left to
// SweepUnsupported to report.

// SweepDryRun reports whether the sweep should only describe what it would do.
func SweepDryRun() bool {
	return os.Getenv("SG_ACC_SWEEP_APPLY") == ""
}

// SweepUnsupported names the resource types a sweep cannot reach, so the caller
// can say so rather than implying the organisation is clean.
var SweepUnsupported = []string{
	"policies (sg-sdk-go ListAllPolicies discards the response body)",
	"roles (sg-sdk-go ListAllRoles discards the response body)",
	"runner groups (sg-sdk-go exposes no list method)",
}

// SweepReport accumulates what a sweep found.
type SweepReport struct {
	DryRun  bool
	Swept   []string
	Skipped int
	Errors  []error
}

func (r *SweepReport) note(kind, name string, err error) {
	if err != nil {
		r.Errors = append(r.Errors, fmt.Errorf("%s %q: %w", kind, name, err))
		return
	}
	r.Swept = append(r.Swept, kind+" "+name)
}

// consider decides whether a candidate belongs to the test suite. Everything
// routes through here, and anything without the prefix is left alone.
func (r *SweepReport) consider(name string) bool {
	if name == "" || !IsTestResourceName(name) {
		r.Skipped++
		return false
	}
	return true
}

func (r *SweepReport) String() string {
	verb := "would delete"
	if !r.DryRun {
		verb = "deleted"
	}
	var b strings.Builder
	fmt.Fprintf(&b, "sweep: %s %d resource(s), left %d untouched",
		verb, len(r.Swept), r.Skipped)
	for _, s := range r.Swept {
		fmt.Fprintf(&b, "\n  %s %s", verb, s)
	}
	for _, e := range r.Errors {
		fmt.Fprintf(&b, "\n  ERROR %s", e)
	}
	for _, u := range SweepUnsupported {
		fmt.Fprintf(&b, "\n  NOT SWEPT %s", u)
	}
	return b.String()
}

// Sweep removes every test-created resource it can reach in org. It continues
// past individual failures: one resource that refuses to delete should not hide
// the rest.
func Sweep(ctx context.Context, client *sgclient.Client, org string) *SweepReport {
	r := &SweepReport{DryRun: SweepDryRun()}

	// Workflow groups first: deleting one takes its workflows with it, so the
	// workflows these tests create need no sweeper of their own.
	sweepWorkflowGroups(ctx, client, org, r)
	sweepConnectors(ctx, client, org, r)
	sweepTemplates(ctx, client, org, r)

	return r
}

func sweepWorkflowGroups(ctx context.Context, client *sgclient.Client, org string, r *SweepReport) {
	resp, err := client.WorkflowGroups.ListAllWorkflowGroups(ctx, org, &sgsdkgo.ListAllWorkflowGroupsRequest{})
	if err != nil {
		r.Errors = append(r.Errors, fmt.Errorf("list workflow groups: %w", err))
		return
	}
	for _, item := range resp.GetMsg() {
		name := ""
		if item.ResourceName != nil {
			name = *item.ResourceName
		}
		if !r.consider(name) {
			continue
		}
		if r.DryRun {
			r.note("workflow group", name, nil)
			continue
		}
		// A group that still holds workflows refuses to delete (400 "Workflow
		// Group not empty"), so remove its workflows first.
		if err := deleteWorkflowsIn(ctx, client, org, name); err != nil {
			r.note("workflow group", name, err)
			continue
		}
		_, err := client.WorkflowGroups.DeleteWorkflowGroup(ctx, org, name)
		r.note("workflow group", name, err)
	}
}

// deleteWorkflowsIn deletes every workflow in the workflow group wfGrp.
func deleteWorkflowsIn(ctx context.Context, client *sgclient.Client, org, wfGrp string) error {
	resp, err := client.Workflows.ListAllWorkflows(ctx, org, wfGrp, &sgsdkgo.ListAllWorkflowsRequest{})
	if err != nil {
		return fmt.Errorf("listing its workflows: %w", err)
	}
	for _, wf := range resp.Msg {
		if wf == nil {
			continue
		}
		id := wf.ResourceName
		if i := strings.LastIndex(wf.ResourceId, "/"); i >= 0 && i < len(wf.ResourceId)-1 {
			id = wf.ResourceId[i+1:]
		}
		if _, err := client.Workflows.DeleteWorkflow(ctx, org, id, wfGrp); err != nil {
			return fmt.Errorf("deleting its workflow %s: %w", id, err)
		}
	}
	return nil
}

func sweepConnectors(ctx context.Context, client *sgclient.Client, org string, r *SweepReport) {
	resp, err := client.Connectors.ListAllConnectors(ctx, org, &sgsdkgo.ListAllConnectorsRequest{})
	if err != nil {
		r.Errors = append(r.Errors, fmt.Errorf("list connectors: %w", err))
		return
	}
	for _, item := range resp.GetMsg() {
		// Connectors report an Id rather than a resource name; it is derived from
		// the name, so it still carries the prefix.
		msg := item.GetMsg()
		if msg == nil || !r.consider(msg.Id) {
			continue
		}
		if r.DryRun {
			r.note("connector", msg.Id, nil)
			continue
		}
		_, err := client.Connectors.DeleteConnector(ctx, org, msg.Id)
		r.note("connector", msg.Id, err)
	}
}

// Templates are the ones that bit hardest: a leftover template made every later
// run fail at create with 409 "already exists".
func sweepTemplates(ctx context.Context, client *sgclient.Client, org string, r *SweepReport) {
	// The per-type /templatetypes/<type>/templates/listall/ endpoints reject org
	// API keys with 401, so list the org's own templates and dispatch on the
	// TemplateType the API returns (sg-sdk-go only exposes it, and
	// LatestRevision, as extra properties).
	kinds := templateKinds(ctx, client, org)

	// Read every page before deleting anything: deleting while paging moves
	// the pagination cursor, and about half the templates were skipped.
	var templates []*sgsdkgo.Template
	req := &sgsdkgo.ListAllTemplatesBasedOnOwnerOrgRequest{}
	for {
		resp, err := client.Templates.ListAllTemplatesBasedOnOwnerOrg(ctx, org, req)
		if err != nil {
			r.Errors = append(r.Errors, fmt.Errorf("list templates: %w", err))
			return
		}
		templates = append(templates, resp.GetMsg()...)
		next := resp.GetLastevaluatedkey()
		if next == "" {
			break
		}
		req.Lastevaluatedkey = &next
	}

	for _, tmpl := range templates {
		if tmpl == nil || !r.consider(tmpl.TemplateName) {
			continue
		}
		extra := tmpl.GetExtraProperties()
		typ, _ := extra["TemplateType"].(string)
		k, ok := kinds[typ]
		if !ok {
			r.note(fmt.Sprintf("template (type %q)", typ), tmpl.TemplateName, fmt.Errorf("no sweeper for this template type"))
			continue
		}
		if r.DryRun {
			r.note(k.label, tmpl.TemplateName, nil)
			continue
		}
		latest, _ := extra["LatestRevision"].(float64)
		r.note(k.label, tmpl.TemplateName, k.delete(tmpl.TemplateName, int(latest)))
	}
}

type templateKind struct {
	label          string
	deleteRevision func(revisionID string)
	deleteTemplate func(name string) error
}

// delete removes revisions 1..latest of the template, then the template.
func (k templateKind) delete(name string, latest int) error {
	for i := 1; i <= latest; i++ {
		k.deleteRevision(fmt.Sprintf("%s:%d", name, i))
	}
	return k.deleteTemplate(name)
}

// templateKinds maps the API's TemplateType to how to delete such a template.
func templateKinds(ctx context.Context, client *sgclient.Client, org string) map[string]templateKind {
	// A template with revisions cannot be deleted, and a published revision
	// must be deprecated before it can be. Errors from those two steps are
	// ignored: the revision may never have been published, or may already be
	// gone. The template delete reports whether it all worked.
	effective := fmt.Sprintf("%d", time.Now().AddDate(0, 1, 0).Unix())
	message := "deprecated by the acceptance test sweeper"

	return map[string]templateKind{
		"IAC": {
			label: "workflow template",
			deleteRevision: func(id string) {
				_, _ = client.WorkflowTemplatesRevisions.UpdateWorkflowTemplateRevision(ctx, org, id, &workflowtemplaterevisions.UpdateWorkflowTemplateRevisionRequest{
					Deprecation: sgsdkgo.Optional(workflowtemplaterevisions.Deprecation{EffectiveDate: &effective, Message: &message}),
				})
				_ = client.WorkflowTemplatesRevisions.DeleteWorkflowTemplateRevision(ctx, org, id, true)
			},
			deleteTemplate: func(name string) error {
				return client.WorkflowTemplates.DeleteWorkflowTemplate(ctx, org, name)
			},
		},
		"IAC_GROUP": {
			label: "stack template",
			deleteRevision: func(id string) {
				_, _ = client.StackTemplateRevisions.UpdateStackTemplateRevision(ctx, org, id, &stacktemplaterevisions.UpdateStackTemplateRevisionRequest{
					Deprecation: sgsdkgo.Optional(stacktemplaterevisions.Deprecation{EffectiveDate: &effective, Message: &message}),
				})
				_ = client.StackTemplateRevisions.DeleteStackTemplateRevision(ctx, org, id, true)
			},
			deleteTemplate: func(name string) error {
				return client.StackTemplates.DeleteStackTemplate(ctx, org, name)
			},
		},
		"WORKFLOW_STEP": {
			label: "workflow step template",
			deleteRevision: func(id string) {
				_, _ = client.WorkflowStepTemplateRevision.UpdateWorkflowStepTemplateRevision(ctx, org, id, &workflowsteptemplaterevision.UpdateWorkflowStepTemplateRevisionModel{
					Deprecation: sgsdkgo.Optional(workflowsteptemplaterevision.Deprecation{EffectiveDate: &effective, Message: &message}),
				})
				_ = client.WorkflowStepTemplateRevision.DeleteWorkflowStepTemplateRevision(ctx, org, id, true)
			},
			deleteTemplate: func(name string) error {
				return client.WorkflowStepTemplate.DeleteWorkflowStepTemplate(ctx, org, name)
			},
		},
	}
}

// DeleteTemplate deletes a template of templateType ("IAC", "IAC_GROUP" or
// "WORKFLOW_STEP") together with its revisions 1..revisions, deprecating each
// first. Use it in test cleanup: deleting only the template fails while any
// revision exists, and that failure is easy to miss when the error is dropped.
func DeleteTemplate(ctx context.Context, client *sgclient.Client, org, templateType, name string, revisions int) error {
	k, ok := templateKinds(ctx, client, org)[templateType]
	if !ok {
		return fmt.Errorf("unknown template type %q", templateType)
	}
	return k.delete(name, revisions)
}
