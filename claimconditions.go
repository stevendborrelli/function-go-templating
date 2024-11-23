package main

import (
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/function-sdk-go/errors"
	fnv1 "github.com/crossplane/function-sdk-go/proto/v1"
	"github.com/crossplane/function-sdk-go/response"
	corev1 "k8s.io/api/core/v1"
)

// A CompositionTarget is the target of a composition event or condition.
type CompositionTarget string

// A TargetedCondition represents a condition produced by the composition
// process. It can target either the XR only, or both the XR and the claim.
type TargetedCondition struct {
	xpv1.Condition `json:",inline"`
	Target         CompositionTarget `json:"target"`
}

// Composition event and condition targets.
const (
	CompositionTargetComposite         CompositionTarget = "Composite"
	CompositionTargetCompositeAndClaim CompositionTarget = "CompositeAndClaim"
)

// UpdateClaimConditions updates Conditions in the Claim and Composite
func UpdateClaimConditions(rsp *fnv1.RunFunctionResponse, conditions ...TargetedCondition) (*fnv1.RunFunctionResponse, error) {
	for _, c := range conditions {
		if xpv1.IsSystemConditionType(xpv1.ConditionType(c.Type)) {
			response.Fatal(rsp, errors.Errorf("cannot set ClaimCondition type: %s is a reserved Crossplane Condition", c.Type))
			return rsp, nil
		}
		co := transformCondition(c)
		UpdateResponseWithCondition(rsp, co)
	}
	return rsp, nil
}

// transformCondition converts a TargetedCondition to be compatible with the Protobuf SDK
func transformCondition(tc TargetedCondition) *fnv1.Condition {
	c := &fnv1.Condition{
		Type:   string(tc.Condition.Type),
		Reason: string(tc.Condition.Reason),
		Target: transformTarget(tc.Target),
	}

	switch tc.Condition.Status {
	case corev1.ConditionTrue:
		c.Status = fnv1.Status_STATUS_CONDITION_TRUE
	case corev1.ConditionFalse:
		c.Status = fnv1.Status_STATUS_CONDITION_FALSE
	case corev1.ConditionUnknown:
		fallthrough
	default:
		c.Status = fnv1.Status_STATUS_CONDITION_UNKNOWN
	}

	if tc.Message != "" {
		c.Message = &tc.Message
	}
	return c
}

// transformTarget converts the input into a target Go SDK Enum
// Default to TARGET_COMPOSITE_AND_CLAIM
func transformTarget(ct CompositionTarget) *fnv1.Target {
	if ct == CompositionTargetComposite {
		return fnv1.Target_TARGET_COMPOSITE.Enum()
	}
	return fnv1.Target_TARGET_COMPOSITE_AND_CLAIM.Enum()
}

func UpdateResponseWithCondition(rsp *fnv1.RunFunctionResponse, c *fnv1.Condition) {
	if rsp.GetConditions() == nil {
		rsp.Conditions = make([]*fnv1.Condition, 0, 1)
	}
	rsp.Conditions = append(rsp.GetConditions(), c)
}