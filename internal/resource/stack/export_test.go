package stack

// ComputeWorkflowId exposes computeWorkflowId to stack_test so acceptance tests can
// derive the exact expected workflow_id instead of duplicating the formula.
var ComputeWorkflowId = computeWorkflowId

// OverlayConfigUnknowns exposes overlayConfigUnknowns to stack_test so its value
// shapes (nested objects, lists, maps, length mismatch) can be unit tested
// directly; the acceptance test needs a live API and only covers one path.
var OverlayConfigUnknowns = overlayConfigUnknowns

// WorkflowsMismatchDetail exposes workflowsMismatchDetail to stack_test so each
// wording of the workflows_config mismatch error can be unit tested.
var WorkflowsMismatchDetail = workflowsMismatchDetail
