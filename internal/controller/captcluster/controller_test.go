package captcluster

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	infrastructurev1beta1 "github.com/appthrust/capt/api/v1beta1"
)

func TestReconciler_Reconcile(t *testing.T) {
	testCases := []struct {
		name           string
		existingObjs   []runtime.Object
		expectedResult reconcile.Result
		expectedError  error
		validate       func(t *testing.T, c client.Client)
	}{
		{
			name:           "CAPTCluster not found",
			existingObjs:   []runtime.Object{},
			expectedResult: reconcile.Result{},
			expectedError:  nil,
		},
		{
			name: "Parent Cluster not found",
			existingObjs: []runtime.Object{
				&infrastructurev1beta1.CAPTCluster{
					ObjectMeta: metav1.ObjectMeta{
						Name:       "test-cluster",
						Namespace:  "default",
						Finalizers: []string{CAPTClusterFinalizer},
					},
					Spec: infrastructurev1beta1.CAPTClusterSpec{
						Region:        "us-west-2",
						ExistingVPCID: "vpc-123456",
					},
				},
			},
			expectedResult: reconcile.Result{
				RequeueAfter: requeueInterval,
			},
			expectedError: nil,
			validate: func(t *testing.T, c client.Client) {
				captCluster := &infrastructurev1beta1.CAPTCluster{}
				err := c.Get(context.Background(), types.NamespacedName{
					Name:      "test-cluster",
					Namespace: "default",
				}, captCluster)
				assert.NoError(t, err)

				// Verify WaitingForCluster condition is set
				var waitingCondition *metav1.Condition
				for i := range captCluster.Status.Conditions {
					if captCluster.Status.Conditions[i].Type == WaitingForClusterCondition {
						waitingCondition = &captCluster.Status.Conditions[i]
						break
					}
				}
				if assert.NotNil(t, waitingCondition) {
					assert.Equal(t, metav1.ConditionTrue, waitingCondition.Status)
					assert.Equal(t, "ClusterNotFound", waitingCondition.Reason)
				}
				assert.False(t, captCluster.Status.Ready)
			},
		},
		{
			name: "Parent Cluster exists",
			existingObjs: []runtime.Object{
				&infrastructurev1beta1.CAPTCluster{
					ObjectMeta: metav1.ObjectMeta{
						Name:       "test-cluster",
						Namespace:  "default",
						Finalizers: []string{CAPTClusterFinalizer},
					},
					Spec: infrastructurev1beta1.CAPTClusterSpec{
						Region:        "us-west-2",
						ExistingVPCID: "vpc-123456",
					},
				},
				&clusterv1.Cluster{
					TypeMeta: metav1.TypeMeta{
						APIVersion: clusterv1.GroupVersion.String(),
						Kind:       "Cluster",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-cluster",
						Namespace: "default",
					},
				},
			},
			expectedResult: reconcile.Result{},
			expectedError:  nil,
			validate: func(t *testing.T, c client.Client) {
				captCluster := &infrastructurev1beta1.CAPTCluster{}
				err := c.Get(context.Background(), types.NamespacedName{
					Name:      "test-cluster",
					Namespace: "default",
				}, captCluster)
				assert.NoError(t, err)

				// Verify owner reference is set
				assert.True(t, controllerutil.ContainsFinalizer(captCluster, CAPTClusterFinalizer))
				assert.Len(t, captCluster.OwnerReferences, 1)
				assert.Equal(t, "Cluster", captCluster.OwnerReferences[0].Kind)
				assert.Equal(t, clusterv1.GroupVersion.String(), captCluster.OwnerReferences[0].APIVersion)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			scheme := runtime.NewScheme()
			_ = infrastructurev1beta1.AddToScheme(scheme)
			_ = clusterv1.AddToScheme(scheme)
			_ = corev1.AddToScheme(scheme)

			fakeClient := fake.NewClientBuilder().
				WithScheme(scheme).
				WithRuntimeObjects(tc.existingObjs...).
				WithStatusSubresource(&infrastructurev1beta1.CAPTCluster{}).
				WithStatusSubresource(&clusterv1.Cluster{}).
				Build()

			reconciler := &Reconciler{
				Client: fakeClient,
				Scheme: scheme,
			}

			result, err := reconciler.Reconcile(context.Background(), reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      "test-cluster",
					Namespace: "default",
				},
			})

			if tc.expectedError == nil {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.expectedError.Error())
			}

			assert.Equal(t, tc.expectedResult, result)

			if tc.validate != nil {
				tc.validate(t, fakeClient)
			}
		})
	}
}
