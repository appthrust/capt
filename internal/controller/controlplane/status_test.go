package controlplane

import (
	"context"
	"testing"

	controlplanev1beta1 "github.com/appthrust/capt/api/controlplane/v1beta1"
	infrastructurev1beta1 "github.com/appthrust/capt/api/v1beta1"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestUpdateStatus(t *testing.T) {
	scheme := setupScheme()
	now := metav1.Now()

	tests := []struct {
		name               string
		controlPlane       *controlplanev1beta1.CAPTControlPlane
		workspaceApply     *infrastructurev1beta1.WorkspaceTemplateApply
		cluster            *clusterv1.Cluster
		expectedPhase      string
		expectedReady      bool
		expectRequeue      bool
		expectedConditions []metav1.Condition
		validate           func(t *testing.T, controlPlane *controlplanev1beta1.CAPTControlPlane)
	}{
		{
			name: "Initial state - workspace not applied",
			controlPlane: &controlplanev1beta1.CAPTControlPlane{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-controlplane",
					Namespace: "default",
				},
				Status: controlplanev1beta1.CAPTControlPlaneStatus{
					WorkspaceTemplateStatus: &controlplanev1beta1.WorkspaceTemplateStatus{},
				},
			},
			workspaceApply: &infrastructurev1beta1.WorkspaceTemplateApply{
				Status: infrastructurev1beta1.WorkspaceTemplateApplyStatus{
					Applied: false,
					Conditions: []xpv1.Condition{
						{
							Type:   xpv1.TypeReady,
							Status: corev1.ConditionFalse,
						},
					},
				},
			},
			cluster: &clusterv1.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-cluster",
					Namespace: "default",
				},
			},
			expectedPhase: "Creating",
			expectedReady: false,
			expectRequeue: true,
			expectedConditions: []metav1.Condition{
				{
					Type:   controlplanev1beta1.ControlPlaneReadyCondition,
					Status: metav1.ConditionFalse,
					Reason: controlplanev1beta1.ReasonCreating,
				},
			},
			validate: func(t *testing.T, controlPlane *controlplanev1beta1.CAPTControlPlane) {
				assert.NotNil(t, controlPlane.Status.WorkspaceTemplateStatus)
				assert.False(t, controlPlane.Status.WorkspaceTemplateStatus.Ready)
				assert.Empty(t, controlPlane.Status.WorkspaceTemplateStatus.LastAppliedRevision)
			},
		},
		{
			name: "Workspace applied - transition to ready state",
			controlPlane: &controlplanev1beta1.CAPTControlPlane{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-controlplane",
					Namespace: "default",
				},
				Status: controlplanev1beta1.CAPTControlPlaneStatus{
					WorkspaceTemplateStatus: &controlplanev1beta1.WorkspaceTemplateStatus{},
				},
			},
			workspaceApply: &infrastructurev1beta1.WorkspaceTemplateApply{
				Status: infrastructurev1beta1.WorkspaceTemplateApplyStatus{
					Applied:         true,
					LastAppliedTime: &now,
					Conditions: []xpv1.Condition{
						{
							Type:   xpv1.TypeReady,
							Status: corev1.ConditionTrue,
						},
					},
				},
			},
			cluster: &clusterv1.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-cluster",
					Namespace: "default",
				},
			},
			expectedPhase: "Ready",
			expectedReady: true,
			expectRequeue: false,
			expectedConditions: []metav1.Condition{
				{
					Type:   controlplanev1beta1.ControlPlaneReadyCondition,
					Status: metav1.ConditionTrue,
					Reason: controlplanev1beta1.ReasonReady,
				},
			},
			validate: func(t *testing.T, controlPlane *controlplanev1beta1.CAPTControlPlane) {
				assert.NotNil(t, controlPlane.Status.WorkspaceTemplateStatus)
				assert.True(t, controlPlane.Status.WorkspaceTemplateStatus.Ready)
				assert.NotEmpty(t, controlPlane.Status.WorkspaceTemplateStatus.LastAppliedRevision)
			},
		},
		{
			name: "Error state - workspace failed",
			controlPlane: &controlplanev1beta1.CAPTControlPlane{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-controlplane",
					Namespace: "default",
				},
				Status: controlplanev1beta1.CAPTControlPlaneStatus{
					WorkspaceTemplateStatus: &controlplanev1beta1.WorkspaceTemplateStatus{},
				},
			},
			workspaceApply: &infrastructurev1beta1.WorkspaceTemplateApply{
				Status: infrastructurev1beta1.WorkspaceTemplateApplyStatus{
					Applied: false,
					Conditions: []xpv1.Condition{
						{
							Type:    xpv1.TypeReady,
							Status:  corev1.ConditionFalse,
							Message: "Workspace creation failed",
						},
					},
				},
			},
			cluster: &clusterv1.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-cluster",
					Namespace: "default",
				},
			},
			expectedPhase: "Failed",
			expectedReady: false,
			expectRequeue: true,
			expectedConditions: []metav1.Condition{
				{
					Type:    controlplanev1beta1.ControlPlaneReadyCondition,
					Status:  metav1.ConditionFalse,
					Reason:  controlplanev1beta1.ReasonWorkspaceError,
					Message: "Workspace creation failed",
				},
			},
			validate: func(t *testing.T, controlPlane *controlplanev1beta1.CAPTControlPlane) {
				assert.NotNil(t, controlPlane.Status.WorkspaceTemplateStatus)
				assert.False(t, controlPlane.Status.WorkspaceTemplateStatus.Ready)
				assert.Equal(t, "Workspace creation failed", controlPlane.Status.WorkspaceTemplateStatus.LastFailureMessage)
			},
		},
		{
			name: "Recovery from error state",
			controlPlane: &controlplanev1beta1.CAPTControlPlane{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-controlplane",
					Namespace: "default",
				},
				Status: controlplanev1beta1.CAPTControlPlaneStatus{
					Phase: "Failed",
					WorkspaceTemplateStatus: &controlplanev1beta1.WorkspaceTemplateStatus{
						Ready:              false,
						LastFailureMessage: "Previous failure",
					},
					Conditions: []metav1.Condition{
						{
							Type:    controlplanev1beta1.ControlPlaneReadyCondition,
							Status:  metav1.ConditionFalse,
							Reason:  controlplanev1beta1.ReasonWorkspaceError,
							Message: "Previous failure",
						},
					},
				},
			},
			workspaceApply: &infrastructurev1beta1.WorkspaceTemplateApply{
				Status: infrastructurev1beta1.WorkspaceTemplateApplyStatus{
					Applied:         true,
					LastAppliedTime: &now,
					Conditions: []xpv1.Condition{
						{
							Type:   xpv1.TypeReady,
							Status: corev1.ConditionTrue,
						},
					},
				},
			},
			cluster: &clusterv1.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-cluster",
					Namespace: "default",
				},
			},
			expectedPhase: "Ready",
			expectedReady: true,
			expectRequeue: false,
			expectedConditions: []metav1.Condition{
				{
					Type:   controlplanev1beta1.ControlPlaneReadyCondition,
					Status: metav1.ConditionTrue,
					Reason: controlplanev1beta1.ReasonReady,
				},
			},
			validate: func(t *testing.T, controlPlane *controlplanev1beta1.CAPTControlPlane) {
				assert.NotNil(t, controlPlane.Status.WorkspaceTemplateStatus)
				assert.True(t, controlPlane.Status.WorkspaceTemplateStatus.Ready)
				assert.Empty(t, controlPlane.Status.WorkspaceTemplateStatus.LastFailureMessage)
				assert.NotEmpty(t, controlPlane.Status.WorkspaceTemplateStatus.LastAppliedRevision)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := fake.NewClientBuilder().
				WithScheme(scheme).
				WithObjects(tt.controlPlane, tt.cluster).
				WithStatusSubresource(tt.controlPlane, tt.cluster).
				Build()

			r := &Reconciler{
				Client: client,
				Scheme: scheme,
			}

			result, err := r.updateStatus(context.Background(), tt.controlPlane, tt.workspaceApply, tt.cluster)
			assert.NoError(t, err)

			updatedControlPlane := &controlplanev1beta1.CAPTControlPlane{}
			err = client.Get(context.Background(), types.NamespacedName{
				Name:      tt.controlPlane.Name,
				Namespace: tt.controlPlane.Namespace,
			}, updatedControlPlane)
			assert.NoError(t, err)

			assert.Equal(t, tt.expectedPhase, updatedControlPlane.Status.Phase)
			assert.Equal(t, tt.expectedReady, updatedControlPlane.Status.Ready)
			assert.Equal(t, tt.expectRequeue, result.RequeueAfter > 0)

			// Verify conditions
			for _, expectedCond := range tt.expectedConditions {
				found := false
				for _, actualCond := range updatedControlPlane.Status.Conditions {
					if actualCond.Type == expectedCond.Type {
						assert.Equal(t, expectedCond.Status, actualCond.Status)
						assert.Equal(t, expectedCond.Reason, actualCond.Reason)
						if expectedCond.Message != "" {
							assert.Equal(t, expectedCond.Message, actualCond.Message)
						}
						found = true
						break
					}
				}
				assert.True(t, found, "Expected condition %s not found", expectedCond.Type)
			}

			// Run additional validations
			if tt.validate != nil {
				tt.validate(t, updatedControlPlane)
			}

			// Verify cluster status
			updatedCluster := &clusterv1.Cluster{}
			err = client.Get(context.Background(), types.NamespacedName{
				Name:      tt.cluster.Name,
				Namespace: tt.cluster.Namespace,
			}, updatedCluster)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedReady, updatedCluster.Status.ControlPlaneReady)
		})
	}
}

func TestSetFailedStatus(t *testing.T) {
	scheme := setupScheme()

	tests := []struct {
		name          string
		controlPlane  *controlplanev1beta1.CAPTControlPlane
		cluster       *clusterv1.Cluster
		reason        string
		message       string
		expectedPhase string
		validate      func(t *testing.T, controlPlane *controlplanev1beta1.CAPTControlPlane, cluster *clusterv1.Cluster)
	}{
		{
			name: "Set initial failure status",
			controlPlane: &controlplanev1beta1.CAPTControlPlane{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-controlplane",
					Namespace: "default",
				},
				Status: controlplanev1beta1.CAPTControlPlaneStatus{
					WorkspaceTemplateStatus: &controlplanev1beta1.WorkspaceTemplateStatus{},
				},
			},
			cluster: &clusterv1.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-cluster",
					Namespace: "default",
				},
			},
			reason:        "InitialFailure",
			message:       "Initial failure message",
			expectedPhase: "Failed",
			validate: func(t *testing.T, controlPlane *controlplanev1beta1.CAPTControlPlane, cluster *clusterv1.Cluster) {
				assert.False(t, controlPlane.Status.Ready)
				assert.Equal(t, "Initial failure message", controlPlane.Status.WorkspaceTemplateStatus.LastFailureMessage)
				assert.False(t, cluster.Status.ControlPlaneReady)
				assert.Equal(t, "Initial failure message", *cluster.Status.FailureMessage)
			},
		},
		{
			name: "Update existing failure status",
			controlPlane: &controlplanev1beta1.CAPTControlPlane{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-controlplane",
					Namespace: "default",
				},
				Status: controlplanev1beta1.CAPTControlPlaneStatus{
					Phase: "Failed",
					WorkspaceTemplateStatus: &controlplanev1beta1.WorkspaceTemplateStatus{
						LastFailureMessage: "Previous failure",
					},
					Conditions: []metav1.Condition{
						{
							Type:    controlplanev1beta1.ControlPlaneReadyCondition,
							Status:  metav1.ConditionFalse,
							Reason:  "PreviousFailure",
							Message: "Previous failure message",
						},
					},
				},
			},
			cluster: &clusterv1.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-cluster",
					Namespace: "default",
				},
				Status: clusterv1.ClusterStatus{
					FailureMessage: func() *string {
						s := "Previous failure"
						return &s
					}(),
				},
			},
			reason:        "UpdatedFailure",
			message:       "Updated failure message",
			expectedPhase: "Failed",
			validate: func(t *testing.T, controlPlane *controlplanev1beta1.CAPTControlPlane, cluster *clusterv1.Cluster) {
				assert.False(t, controlPlane.Status.Ready)
				assert.Equal(t, "Updated failure message", controlPlane.Status.WorkspaceTemplateStatus.LastFailureMessage)
				assert.False(t, cluster.Status.ControlPlaneReady)
				assert.Equal(t, "Updated failure message", *cluster.Status.FailureMessage)

				// Verify condition update
				var failedCondition *metav1.Condition
				for i := range controlPlane.Status.Conditions {
					if controlPlane.Status.Conditions[i].Type == controlplanev1beta1.ControlPlaneReadyCondition {
						failedCondition = &controlPlane.Status.Conditions[i]
						break
					}
				}
				assert.NotNil(t, failedCondition)
				assert.Equal(t, metav1.ConditionFalse, failedCondition.Status)
				assert.Equal(t, "UpdatedFailure", failedCondition.Reason)
				assert.Equal(t, "Updated failure message", failedCondition.Message)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := fake.NewClientBuilder().
				WithScheme(scheme).
				WithObjects(tt.controlPlane, tt.cluster).
				WithStatusSubresource(tt.controlPlane, tt.cluster).
				Build()

			r := &Reconciler{
				Client: client,
				Scheme: scheme,
			}

			result, err := r.setFailedStatus(context.Background(), tt.controlPlane, tt.cluster, tt.reason, tt.message)
			assert.NoError(t, err)
			assert.True(t, result.RequeueAfter > 0)

			updatedControlPlane := &controlplanev1beta1.CAPTControlPlane{}
			err = client.Get(context.Background(), types.NamespacedName{
				Name:      tt.controlPlane.Name,
				Namespace: tt.controlPlane.Namespace,
			}, updatedControlPlane)
			assert.NoError(t, err)

			updatedCluster := &clusterv1.Cluster{}
			err = client.Get(context.Background(), types.NamespacedName{
				Name:      tt.cluster.Name,
				Namespace: tt.cluster.Namespace,
			}, updatedCluster)
			assert.NoError(t, err)

			assert.Equal(t, tt.expectedPhase, updatedControlPlane.Status.Phase)

			if tt.validate != nil {
				tt.validate(t, updatedControlPlane, updatedCluster)
			}
		})
	}
}
