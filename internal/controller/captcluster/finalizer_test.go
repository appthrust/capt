package captcluster

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	infrastructurev1beta1 "github.com/appthrust/capt/api/v1beta1"
)

func TestReconciler_reconcileDelete(t *testing.T) {
	testCases := []struct {
		name          string
		captCluster   *infrastructurev1beta1.CAPTCluster
		existingObjs  []runtime.Object
		expectedError error
		validate      func(t *testing.T, c client.Client)
	}{
		{
			name: "Retain VPC on delete",
			captCluster: &infrastructurev1beta1.CAPTCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:              "test-cluster",
					Namespace:         "default",
					DeletionTimestamp: &metav1.Time{Time: time.Now()},
					Finalizers:        []string{CAPTClusterFinalizer},
				},
				Spec: infrastructurev1beta1.CAPTClusterSpec{
					RetainVPCOnDelete: true,
					VPCTemplateRef: &infrastructurev1beta1.WorkspaceTemplateReference{
						Name: "vpc-template",
					},
					WorkspaceTemplateApplyName: "test-workspace",
				},
				Status: infrastructurev1beta1.CAPTClusterStatus{
					VPCID: "vpc-123456",
				},
			},
			existingObjs: []runtime.Object{
				&infrastructurev1beta1.WorkspaceTemplateApply{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-workspace",
						Namespace: "default",
					},
				},
			},
			expectedError: nil,
			validate: func(t *testing.T, c client.Client) {
				// Verify finalizer is removed
				captCluster := &infrastructurev1beta1.CAPTCluster{}
				err := c.Get(context.Background(), types.NamespacedName{
					Name:      "test-cluster",
					Namespace: "default",
				}, captCluster)
				assert.NoError(t, err)
				assert.False(t, controllerutil.ContainsFinalizer(captCluster, CAPTClusterFinalizer))

				// Verify WorkspaceTemplateApply still exists
				workspaceApply := &infrastructurev1beta1.WorkspaceTemplateApply{}
				err = c.Get(context.Background(), types.NamespacedName{
					Name:      "test-workspace",
					Namespace: "default",
				}, workspaceApply)
				assert.NoError(t, err)
			},
		},
		{
			name: "Delete VPC and WorkspaceTemplateApply",
			captCluster: &infrastructurev1beta1.CAPTCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:              "test-cluster",
					Namespace:         "default",
					DeletionTimestamp: &metav1.Time{Time: time.Now()},
					Finalizers:        []string{CAPTClusterFinalizer},
				},
				Spec: infrastructurev1beta1.CAPTClusterSpec{
					RetainVPCOnDelete:          false,
					WorkspaceTemplateApplyName: "test-workspace",
				},
			},
			existingObjs: []runtime.Object{
				&infrastructurev1beta1.WorkspaceTemplateApply{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-workspace",
						Namespace: "default",
					},
				},
			},
			expectedError: nil,
			validate: func(t *testing.T, c client.Client) {
				// Verify finalizer is removed
				captCluster := &infrastructurev1beta1.CAPTCluster{}
				err := c.Get(context.Background(), types.NamespacedName{
					Name:      "test-cluster",
					Namespace: "default",
				}, captCluster)
				assert.NoError(t, err)
				assert.False(t, controllerutil.ContainsFinalizer(captCluster, CAPTClusterFinalizer))

				// Verify WorkspaceTemplateApply is deleted
				workspaceApply := &infrastructurev1beta1.WorkspaceTemplateApply{}
				err = c.Get(context.Background(), types.NamespacedName{
					Name:      "test-workspace",
					Namespace: "default",
				}, workspaceApply)
				assert.Error(t, err)
			},
		},
		{
			name: "No WorkspaceTemplateApply to delete",
			captCluster: &infrastructurev1beta1.CAPTCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:              "test-cluster",
					Namespace:         "default",
					DeletionTimestamp: &metav1.Time{Time: time.Now()},
					Finalizers:        []string{CAPTClusterFinalizer},
				},
				Spec: infrastructurev1beta1.CAPTClusterSpec{
					RetainVPCOnDelete: false,
				},
			},
			existingObjs:  []runtime.Object{},
			expectedError: nil,
			validate: func(t *testing.T, c client.Client) {
				// Verify finalizer is removed
				captCluster := &infrastructurev1beta1.CAPTCluster{}
				err := c.Get(context.Background(), types.NamespacedName{
					Name:      "test-cluster",
					Namespace: "default",
				}, captCluster)
				assert.NoError(t, err)
				assert.False(t, controllerutil.ContainsFinalizer(captCluster, CAPTClusterFinalizer))
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			scheme := runtime.NewScheme()
			_ = infrastructurev1beta1.AddToScheme(scheme)
			_ = clusterv1.AddToScheme(scheme)
			_ = corev1.AddToScheme(scheme)

			// Create the fake client with all objects
			fakeClient := fake.NewClientBuilder().
				WithScheme(scheme).
				WithStatusSubresource(&infrastructurev1beta1.CAPTCluster{}).
				WithStatusSubresource(&infrastructurev1beta1.WorkspaceTemplateApply{}).
				Build()

			// Create a new context for the test
			ctx := context.Background()

			// Create CAPTCluster
			err := fakeClient.Create(ctx, tc.captCluster.DeepCopy())
			assert.NoError(t, err)

			// Create other objects
			for _, obj := range tc.existingObjs {
				err = fakeClient.Create(ctx, obj.(client.Object))
				assert.NoError(t, err)
			}

			reconciler := &Reconciler{
				Client: fakeClient,
				Scheme: scheme,
			}

			// Get the CAPTCluster from the fake client
			captCluster := &infrastructurev1beta1.CAPTCluster{}
			err = fakeClient.Get(ctx, types.NamespacedName{
				Name:      tc.captCluster.Name,
				Namespace: tc.captCluster.Namespace,
			}, captCluster)
			assert.NoError(t, err)

			// Call reconcileDelete
			result, err := reconciler.reconcileDelete(ctx, captCluster)

			if tc.expectedError == nil {
				assert.NoError(t, err)
				assert.Equal(t, Result{}, result)
			} else {
				assert.EqualError(t, err, tc.expectedError.Error())
			}

			if tc.validate != nil {
				tc.validate(t, fakeClient)
			}
		})
	}
}

func TestReconciler_cleanupWorkspaceTemplateApply(t *testing.T) {
	testCases := []struct {
		name          string
		existingObjs  []runtime.Object
		expectedError error
		validate      func(t *testing.T, c client.Client)
	}{
		{
			name: "Cleanup existing WorkspaceTemplateApply",
			existingObjs: []runtime.Object{
				&infrastructurev1beta1.WorkspaceTemplateApply{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-workspace",
						Namespace: "default",
					},
				},
				&infrastructurev1beta1.CAPTCluster{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-cluster",
						Namespace: "default",
					},
					Spec: infrastructurev1beta1.CAPTClusterSpec{
						WorkspaceTemplateApplyName: "test-workspace",
					},
				},
			},
			expectedError: nil,
			validate: func(t *testing.T, c client.Client) {
				workspaceApply := &infrastructurev1beta1.WorkspaceTemplateApply{}
				err := c.Get(context.Background(), types.NamespacedName{
					Name:      "test-workspace",
					Namespace: "default",
				}, workspaceApply)
				assert.Error(t, err)

				captCluster := &infrastructurev1beta1.CAPTCluster{}
				err = c.Get(context.Background(), types.NamespacedName{
					Name:      "test-cluster",
					Namespace: "default",
				}, captCluster)
				assert.NoError(t, err)
				assert.Empty(t, captCluster.Spec.WorkspaceTemplateApplyName)
			},
		},
		{
			name: "No WorkspaceTemplateApply to cleanup",
			existingObjs: []runtime.Object{
				&infrastructurev1beta1.CAPTCluster{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-cluster",
						Namespace: "default",
					},
				},
			},
			expectedError: nil,
			validate: func(t *testing.T, c client.Client) {
				// Nothing to verify
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
				WithStatusSubresource(&infrastructurev1beta1.CAPTCluster{}).
				WithStatusSubresource(&infrastructurev1beta1.WorkspaceTemplateApply{}).
				Build()

			// Create objects
			for _, obj := range tc.existingObjs {
				err := fakeClient.Create(context.Background(), obj.(client.Object))
				assert.NoError(t, err)
			}

			reconciler := &Reconciler{
				Client: fakeClient,
				Scheme: scheme,
			}

			captCluster := &infrastructurev1beta1.CAPTCluster{}
			err := fakeClient.Get(context.Background(), types.NamespacedName{
				Name:      "test-cluster",
				Namespace: "default",
			}, captCluster)
			assert.NoError(t, err)

			err = reconciler.cleanupWorkspaceTemplateApply(context.Background(), captCluster)

			if tc.expectedError == nil {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.expectedError.Error())
			}

			if tc.validate != nil {
				tc.validate(t, fakeClient)
			}
		})
	}
}
