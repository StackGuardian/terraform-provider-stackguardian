package acctest

import (
	"context"
	"fmt"
	"os"
	"strings"

	sgsdkgo "github.com/StackGuardian/sg-sdk-go"
	sgclient "github.com/StackGuardian/sg-sdk-go/client"
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
		_, err := client.WorkflowGroups.DeleteWorkflowGroup(ctx, org, name)
		r.note("workflow group", name, err)
	}
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
	kinds := []struct {
		kind     sgsdkgo.ListAllTemplatesRequestTemplateType
		label    string
		deleteFn func(name string) error
	}{
		{
			kind:  sgsdkgo.ListAllTemplatesRequestTemplateTypeIac,
			label: "workflow template",
			deleteFn: func(name string) error {
				return client.WorkflowTemplates.DeleteWorkflowTemplate(ctx, org, name)
			},
		},
		{
			kind:  sgsdkgo.ListAllTemplatesRequestTemplateTypeWorkflowStep,
			label: "workflow step template",
			deleteFn: func(name string) error {
				return client.WorkflowStepTemplate.DeleteWorkflowStepTemplate(ctx, org, name)
			},
		},
	}

	for _, k := range kinds {
		resp, err := client.Templates.ListAllTemplates(ctx, k.kind, &sgsdkgo.ListAllTemplatesRequest{})
		if err != nil {
			r.Errors = append(r.Errors, fmt.Errorf("list %s: %w", k.label, err))
			continue
		}
		for _, tmpl := range resp.GetMsg() {
			if !r.consider(tmpl.TemplateName) {
				continue
			}
			if r.DryRun {
				r.note(k.label, tmpl.TemplateName, nil)
				continue
			}
			r.note(k.label, tmpl.TemplateName, k.deleteFn(tmpl.TemplateName))
		}
	}
}
