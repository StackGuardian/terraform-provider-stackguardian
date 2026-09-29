package stack

// ComputeWorkflowId exposes computeWorkflowId to stack_test so acceptance tests can
// derive the exact expected workflow_id instead of duplicating the formula.
var ComputeWorkflowId = computeWorkflowId
