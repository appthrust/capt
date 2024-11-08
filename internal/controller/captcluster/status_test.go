package captcluster

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/util/conditions"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	infrastructurev1beta1 "github.com/appthrust/capt/api/v1beta1"
)

func TestReconciler_updateStatus(t *testing.T) {
	testCases := []struct {
		name        string
		captCluster *infrastructurev1beta1.CAPTCluster
		cluster     *clusterv1.Cluster
		validate    func(t *testing.T, captCluster *infrastructurev1beta1.CAPTCluster, cluster *clusterv1.Cluster)
	}{
		{
			name: "Update ready status",
			captCluster: &infrastructurev1beta1.CAPTCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-cluster",
					Namespace: "default",
				},
				Status: infrastructurev1beta1.CAPTClusterStatus{
					Ready: true,
				},
			},
			cluster: &clusterv1.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-cluster",
					Namespace: "default",
				},
				Status: clusterv1.ClusterStatus{
					Conditions: []clusterv1.Condition{},
				},
			},
			validate: func(t *testing.T, captCluster *infrastructurev1beta1.CAPTCluster, cluster *clusterv1.Cluster) {
				assert.True(t, cluster.Status.InfrastructureReady)

				// Verify ControlPlaneInitialized condition
				cpInitCondition := conditions.Get(cluster, ControlPlaneInitializedCondition)
				if assert.NotNil(t, cpInitCondition) {
					assert.Equal(t, corev1.ConditionTrue, cpInitCondition.Status)
				}

				// Verify InfrastructureReady condition
				infraReadyCondition := conditions.Get(cluster, InfrastructureReadyCondition)
				if assert.NotNil(t, infraReadyCondition) {
					assert.Equal(t, corev1.ConditionTrue, infraReadyCondition.Status)
				}
			},
		},
		{
			name: "Update failure status",
			captCluster: &infrastructurev1beta1.CAPTCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-cluster",
					Namespace: "default",
				},
				Status: infrastructurev1beta1.CAPTClusterStatus{
					Ready: false,
					FailureReason: func() *string {
						s := "TestFailure"
						return &s
					}(),
					FailureMessage: func() *string {
						s := "Test failure message"
						return &s
					}(),
				},
			},
			cluster: &clusterv1.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-cluster",
					Namespace: "default",
				},
			},
			validate: func(t *testing.T, captCluster *infrastructurev1beta1.CAPTCluster, cluster *clusterv1.Cluster) {
				assert.False(t, cluster.Status.InfrastructureReady)
				assert.NotNil(t, cluster.Status.FailureReason)
				assert.NotNil(t, cluster.Status.FailureMessage)
				assert.Equal(t, "TestFailure", string(*cluster.Status.FailureReason))
				assert.Equal(t, "Test failure message", *cluster.Status.FailureMessage)
			},
		},
		{
			name: "Update with nil cluster",
			captCluster: &infrastructurev1beta1.CAPTCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-cluster",
					Namespace: "default",
				},
				Status: infrastructurev1beta1.CAPTClusterStatus{
					Ready: true,
				},
			},
			cluster: nil,
			validate: func(t *testing.T, captCluster *infrastructurev1beta1.CAPTCluster, cluster *clusterv1.Cluster) {
				assert.True(t, captCluster.Status.Ready)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			scheme := runtime.NewScheme()
			_ = infrastructurev1beta1.AddToScheme(scheme)
			_ = clusterv1.AddToScheme(scheme)
			_ = corev1.AddToScheme(scheme)

			objs := []runtime.Object{tc.captCluster}
			if tc.cluster != nil {
				objs = append(objs, tc.cluster)
			}

			fakeClient := fake.NewClientBuilder().
				WithScheme(scheme).
				WithRuntimeObjects(objs...).
				WithStatusSubresource(&infrastructurev1beta1.CAPTCluster{}).
				WithStatusSubresource(&clusterv1.Cluster{}).
				Build()

			reconciler := &Reconciler{
				Client: fakeClient,
				Scheme: scheme,
			}

			err := reconciler.updateStatus(context.Background(), tc.captCluster, tc.cluster)
			assert.NoError(t, err)

			if tc.validate != nil {
				tc.validate(t, tc.captCluster, tc.cluster)
			}
		})
	}
}

func TestReconciler_setFailedStatus(t *testing.T) {
	testCases := []struct {
		name        string
		captCluster *infrastructurev1beta1.CAPTCluster
		cluster     *clusterv1.Cluster
		reason      string
		message     string
		validate    func(t *testing.T, captCluster *infrastructurev1beta1.CAPTCluster, cluster *clusterv1.Cluster)
	}{
		{
			name: "Set failed status",
			captCluster: &infrastructurev1beta1.CAPTCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-cluster",
					Namespace: "default",
				},
				Status: infrastructurev1beta1.CAPTClusterStatus{
					Ready: true,
					WorkspaceTemplateStatus: &infrastructurev1beta1.CAPTClusterWorkspaceStatus{
						Ready: true,
					},
				},
			},
			cluster: &clusterv1.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-cluster",
					Namespace: "default",
				},
			},
			reason:  "TestFailure",
			message: "Test failure message",
			validate: func(t *testing.T, captCluster *infrastructurev1beta1.CAPTCluster, cluster *clusterv1.Cluster) {
				assert.False(t, captCluster.Status.Ready)
				assert.Equal(t, "TestFailure", *captCluster.Status.FailureReason)
				assert.Equal(t, "Test failure message", *captCluster.Status.FailureMessage)

				assert.NotNil(t, captCluster.Status.WorkspaceTemplateStatus)
				assert.False(t, captCluster.Status.WorkspaceTemplateStatus.Ready)
				assert.Equal(t, "Test failure message", captCluster.Status.WorkspaceTemplateStatus.LastFailureMessage)

				var failedCondition *metav1.Condition
				for i := range captCluster.Status.Conditions {
					if captCluster.Status.Conditions[i].Type == infrastructurev1beta1.VPCFailedCondition {
						failedCondition = &captCluster.Status.Conditions[i]
						break
					}
				}
				if assert.NotNil(t, failedCondition) {
					assert.Equal(t, metav1.ConditionTrue, failedCondition.Status)
					assert.Equal(t, "TestFailure", failedCondition.Reason)
					assert.Equal(t, "Test failure message", failedCondition.Message)
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			scheme := runtime.NewScheme()
			_ = infrastructurev1beta1.AddToScheme(scheme)
			_ = clusterv1.AddToScheme(scheme)
			_ = corev1.AddToScheme(scheme)

			objs := []runtime.Object{tc.captCluster}
			if tc.cluster != nil {
				objs = append(objs, tc.cluster)
			}

			fakeClient := fake.NewClientBuilder().
				WithScheme(scheme).
				WithRuntimeObjects(objs...).
				WithStatusSubresource(&infrastructurev1beta1.CAPTCluster{}).
				WithStatusSubresource(&clusterv1.Cluster{}).
				Build()

			reconciler := &Reconciler{
				Client: fakeClient,
				Scheme: scheme,
			}

			_, err := reconciler.setFailedStatus(context.Background(), tc.captCluster, tc.cluster, tc.reason, tc.message)
			assert.Error(t, err)
			assert.Equal(t, tc.message, err.Error())

			if tc.validate != nil {
				tc.validate(t, tc.captCluster, tc.cluster)
			}
		})
	}
}
