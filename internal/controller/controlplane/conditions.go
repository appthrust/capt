package controlplane

import (
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
)

// FindStatusCondition finds the condition that has the matching condition type.
func FindStatusCondition(conditions []xpv1.Condition, conditionType xpv1.ConditionType) *xpv1.Condition {
	for i := range conditions {
		if conditions[i].Type == conditionType {
			return &conditions[i]
		}
	}
	return nil
}
