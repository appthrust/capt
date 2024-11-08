package captcluster

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	infrastructurev1beta1 "github.com/appthrust/capt/api/v1beta1"
)

func TestReconciler_reconcileVPC(t *testing.T) {
	testCases := []struct {
		name          string
		existingObjs  []runtime.Object
		captCluster   *infrastructurev1beta1.CAPTCluster
		cluster       *clusterv1.Cluster
		expectedError error
		validate      func(t *testing.T, captCluster *infrastructurev1beta1.CAPTCluster)
	}{
		{
			name: "Using existing VPC",
			captCluster: &infrastructurev1beta1.CAPTCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-cluster",
					Namespace: "default",
				},
				Spec: infrastructurev1beta1.CAPTClusterSpec{
					Region:        "us-west-2",
					ExistingVPCID: "vpc-123456",
				},
			},
			cluster: &clusterv1.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-cluster",
					Namespace: "default",
				},
			},
			expectedError: nil,
			validate: func(t *testing.T, captCluster *infrastructurev1beta1.CAPTCluster) {
				assert.Equal(t, "vpc-123456", captCluster.Status.VPCID)
				assert.True(t, captCluster.Status.Ready)

				var vpcReadyCondition *metav1.Condition
				for i := range captCluster.Status.Conditions {
					if captCluster.Status.Conditions[i].Type == infrastructurev1beta1.VPCReadyCondition {
						vpcReadyCondition = &captCluster.Status.Conditions[i]
						break
					}
				}
				if assert.NotNil(t, vpcReadyCondition) {
					assert.Equal(t, metav1.ConditionTrue, vpcReadyCondition.Status)
					assert.Equal(t, infrastructurev1beta1.ReasonExistingVPCUsed, vpcReadyCondition.Reason)
				}
			},
		},
		{
			name: "Create new VPC with template",
			existingObjs: []runtime.Object{
				&infrastructurev1beta1.WorkspaceTemplate{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "vpc-template",
						Namespace: "default",
					},
				},
			},
			captCluster: &infrastructurev1beta1.CAPTCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-cluster",
					Namespace: "default",
				},
				Spec: infrastructurev1beta1.CAPTClusterSpec{
					Region: "us-west-2",
					VPCTemplateRef: &infrastructurev1beta1.WorkspaceTemplateReference{
						Name:      "vpc-template",
						Namespace: "default",
					},
				},
			},
			cluster: &clusterv1.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-cluster",
					Namespace: "default",
				},
			},
			expectedError: nil,
			validate: func(t *testing.T, captCluster *infrastructurev1beta1.CAPTCluster) {
				assert.NotEmpty(t, captCluster.Spec.WorkspaceTemplateApplyName)

				var vpcReadyCondition *metav1.Condition
				for i := range captCluster.Status.Conditions {
					if captCluster.Status.Conditions[i].Type == infrastructurev1beta1.VPCReadyCondition {
						vpcReadyCondition = &captCluster.Status.Conditions[i]
						break
					}
				}
				if assert.NotNil(t, vpcReadyCondition) {
					assert.Equal(t, metav1.ConditionFalse, vpcReadyCondition.Status)
					assert.Equal(t, infrastructurev1beta1.ReasonVPCCreating, vpcReadyCondition.Reason)
				}
			},
		},
		{
			name: "Parent cluster is nil",
			captCluster: &infrastructurev1beta1.CAPTCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-cluster",
					Namespace: "default",
				},
				Spec: infrastructurev1beta1.CAPTClusterSpec{
					Region:        "us-west-2",
					ExistingVPCID: "vpc-123456",
				},
			},
			cluster:       nil,
			expectedError: fmt.Errorf("parent cluster is required for VPC reconciliation"),
			validate: func(t *testing.T, captCluster *infrastructurev1beta1.CAPTCluster) {
				assert.False(t, captCluster.Status.Ready)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			scheme := runtime.NewScheme()
			_ = infrastructurev1beta1.AddToScheme(scheme)
			_ = clusterv1.AddToScheme(scheme)
			_ = corev1.AddToScheme(scheme)

			objs := tc.existingObjs
			if tc.captCluster != nil {
				objs = append(objs, tc.captCluster)
			}
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

			_, err := reconciler.reconcileVPC(context.Background(), tc.captCluster, tc.cluster)

			if tc.expectedError == nil {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.expectedError.Error())
			}

			if tc.validate != nil {
				tc.validate(t, tc.captCluster)
			}
		})
	}
}
