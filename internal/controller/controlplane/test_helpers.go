package controlplane

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"

	controlplanev1beta1 "github.com/appthrust/capt/api/controlplane/v1beta1"
	infrastructurev1beta1 "github.com/appthrust/capt/api/v1beta1"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
)

func setupScheme() *runtime.Scheme {
	scheme := runtime.NewScheme()

	// Register standard Kubernetes types
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(corev1.AddToScheme(scheme))

	// Register CAPI types
	utilruntime.Must(clusterv1.AddToScheme(scheme))

	// Register CAPT types
	utilruntime.Must(controlplanev1beta1.AddToScheme(scheme))
	utilruntime.Must(infrastructurev1beta1.AddToScheme(scheme))

	// Register version priorities
	// Core Kubernetes APIs
	utilruntime.Must(scheme.SetVersionPriority(corev1.SchemeGroupVersion))

	// CAPI APIs
	utilruntime.Must(scheme.SetVersionPriority(clusterv1.GroupVersion))

	// CAPT APIs
	utilruntime.Must(scheme.SetVersionPriority(controlplanev1beta1.GroupVersion))
	utilruntime.Must(scheme.SetVersionPriority(infrastructurev1beta1.GroupVersion))

	return scheme
}

// Helper function to create test conditions
func createTestCondition(conditionType string, status metav1.ConditionStatus, reason, message string) metav1.Condition {
	return metav1.Condition{
		Type:               conditionType,
		Status:             status,
		LastTransitionTime: metav1.Now(),
		Reason:             reason,
		Message:            message,
	}
}

// Helper function to verify conditions
func containsCondition(conditions []metav1.Condition, conditionType string, status metav1.ConditionStatus) bool {
	for _, condition := range conditions {
		if condition.Type == conditionType && condition.Status == status {
			return true
		}
	}
	return false
}
