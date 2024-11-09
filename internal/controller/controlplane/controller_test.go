package controlplane

import (
	"context"
	"testing"
	"time"

	controlplanev1beta1 "github.com/appthrust/capt/api/controlplane/v1beta1"
	infrastructurev1beta1 "github.com/appthrust/capt/api/v1beta1"
	"github.com/stretchr/testify/assert"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func TestReconcile(t *testing.T) {
	scheme := setupScheme()

	tests := []struct {
		name           string
		existingObjs   []runtime.Object
		expectedResult Result
		expectedError  bool
		validate       func(t *testing.T, client client.Client, result Result, err error)
	}{
		{
			name:           "ControlPlane not found",
			existingObjs:   []runtime.Object{},
			expectedResult: Result{},
			expectedError:  false,
		},
		{
			name: "ControlPlane being deleted",
			existingObjs: []runtime.Object{
				&controlplanev1beta1.CAPTControlPlane{
					ObjectMeta: metav1.ObjectMeta{
						Name:              "test-controlplane",
						Namespace:         "default",
						DeletionTimestamp: &metav1.Time{Time: time.Now()},
						Finalizers:        []string{CAPTControlPlaneFinalizer},
					},
					Status: controlplanev1beta1.CAPTControlPlaneStatus{
						WorkspaceTemplateStatus: &controlplanev1beta1.WorkspaceTemplateStatus{},
					},
				},
			},
			expectedResult: Result{},
			expectedError:  false,
			validate: func(t *testing.T, client client.Client, result Result, err error) {
				// Verify the result
				assert.Equal(t, Result{}, result)
				assert.NoError(t, err)

				// Verify the object is deleted
				controlPlane := &controlplanev1beta1.CAPTControlPlane{}
				err = client.Get(context.Background(), types.NamespacedName{
					Name:      "test-controlplane",
					Namespace: "default",
				}, controlPlane)

				assert.True(t, apierrors.IsNotFound(err), "Expected NotFound error, got %v", err)
			},
		},
		{
			name: "Missing owner cluster",
			existingObjs: []runtime.Object{
				&controlplanev1beta1.CAPTControlPlane{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-controlplane",
						Namespace: "default",
					},
					Spec: controlplanev1beta1.CAPTControlPlaneSpec{
						WorkspaceTemplateRef: controlplanev1beta1.WorkspaceTemplateReference{
							Name:      "test-template",
							Namespace: "default",
						},
					},
					Status: controlplanev1beta1.CAPTControlPlaneStatus{
						WorkspaceTemplateStatus: &controlplanev1beta1.WorkspaceTemplateStatus{},
					},
				},
				&infrastructurev1beta1.WorkspaceTemplate{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-template",
						Namespace: "default",
					},
				},
			},
			expectedResult: Result{RequeueAfter: requeueInterval},
			expectedError:  false,
			validate: func(t *testing.T, client client.Client, result Result, err error) {
				controlPlane := &controlplanev1beta1.CAPTControlPlane{}
				err = client.Get(context.Background(), types.NamespacedName{
					Name:      "test-controlplane",
					Namespace: "default",
				}, controlPlane)
				assert.NoError(t, err)
				assert.True(t, controllerutil.ContainsFinalizer(controlPlane, CAPTControlPlaneFinalizer))

				// Verify waiting condition
				found := false
				for _, condition := range controlPlane.Status.Conditions {
					if condition.Type == controlplanev1beta1.ControlPlaneReadyCondition {
						assert.Equal(t, metav1.ConditionFalse, condition.Status)
						assert.Equal(t, controlplanev1beta1.ReasonCreating, condition.Reason)
						found = true
						break
					}
				}
				assert.True(t, found, "Expected condition not found")
			},
		},
		{
			name: "Normal reconciliation",
			existingObjs: []runtime.Object{
				&controlplanev1beta1.CAPTControlPlane{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-controlplane",
						Namespace: "default",
					},
					Spec: controlplanev1beta1.CAPTControlPlaneSpec{
						Version: "1.21",
						WorkspaceTemplateRef: controlplanev1beta1.WorkspaceTemplateReference{
							Name:      "test-template",
							Namespace: "default",
						},
					},
					Status: controlplanev1beta1.CAPTControlPlaneStatus{
						WorkspaceTemplateStatus: &controlplanev1beta1.WorkspaceTemplateStatus{},
					},
				},
				&clusterv1.Cluster{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-controlplane",
						Namespace: "default",
						UID:       "test-uid",
					},
				},
				&infrastructurev1beta1.WorkspaceTemplate{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-template",
						Namespace: "default",
					},
				},
				&infrastructurev1beta1.WorkspaceTemplateApply{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-workspace",
						Namespace: "default",
					},
					Spec: infrastructurev1beta1.WorkspaceTemplateApplySpec{
						TemplateRef: infrastructurev1beta1.WorkspaceTemplateReference{
							Name:      "test-template",
							Namespace: "default",
						},
					},
				},
			},
			expectedResult: Result{RequeueAfter: requeueInterval},
			expectedError:  false,
			validate: func(t *testing.T, client client.Client, result Result, err error) {
				controlPlane := &controlplanev1beta1.CAPTControlPlane{}
				err = client.Get(context.Background(), types.NamespacedName{
					Name:      "test-controlplane",
					Namespace: "default",
				}, controlPlane)
				assert.NoError(t, err)

				// Verify owner reference
				found := false
				for _, ref := range controlPlane.OwnerReferences {
					if ref.Kind == "Cluster" {
						found = true
						break
					}
				}
				assert.True(t, found, "Expected owner reference not found")

				// Verify finalizer
				assert.True(t, controllerutil.ContainsFinalizer(controlPlane, CAPTControlPlaneFinalizer))

				// Verify workspace template apply creation
				assert.NotEmpty(t, controlPlane.Spec.WorkspaceTemplateApplyName)

				// Verify conditions
				found = false
				for _, condition := range controlPlane.Status.Conditions {
					if condition.Type == controlplanev1beta1.ControlPlaneReadyCondition {
						assert.Equal(t, metav1.ConditionFalse, condition.Status)
						assert.Equal(t, controlplanev1beta1.ReasonCreating, condition.Reason)
						found = true
						break
					}
				}
				assert.True(t, found, "Expected condition not found")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := fake.NewClientBuilder().
				WithScheme(scheme).
				WithRuntimeObjects(tt.existingObjs...).
				WithStatusSubresource(&controlplanev1beta1.CAPTControlPlane{}, &clusterv1.Cluster{}).
				Build()

			r := &Reconciler{
				Client: client,
				Scheme: scheme,
			}

			namespacedName := types.NamespacedName{
				Name:      "test-controlplane",
				Namespace: "default",
			}

			result, err := r.Reconcile(context.Background(), namespacedName)

			if tt.validate != nil {
				tt.validate(t, client, result, err)
			} else {
				if tt.expectedError {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}

				assert.Equal(t, tt.expectedResult.Requeue, result.Requeue)
				if tt.expectedResult.RequeueAfter > 0 {
					assert.Equal(t, tt.expectedResult.RequeueAfter, result.RequeueAfter)
				}
			}
		})
	}
}

func TestSetOwnerReference(t *testing.T) {
	scheme := setupScheme()

	tests := []struct {
		name          string
		controlPlane  *controlplanev1beta1.CAPTControlPlane
		cluster       *clusterv1.Cluster
		expectedError bool
		validate      func(t *testing.T, controlPlane *controlplanev1beta1.CAPTControlPlane)
	}{
		{
			name: "Set owner reference successfully",
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
					UID:       "test-uid",
				},
			},
			expectedError: false,
			validate: func(t *testing.T, controlPlane *controlplanev1beta1.CAPTControlPlane) {
				assert.Len(t, controlPlane.OwnerReferences, 1)
				assert.Equal(t, "Cluster", controlPlane.OwnerReferences[0].Kind)
				assert.Equal(t, "test-cluster", controlPlane.OwnerReferences[0].Name)
				assert.Equal(t, "test-uid", string(controlPlane.OwnerReferences[0].UID))
			},
		},
		{
			name: "Owner reference already set",
			controlPlane: &controlplanev1beta1.CAPTControlPlane{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-controlplane",
					Namespace: "default",
					OwnerReferences: []metav1.OwnerReference{
						{
							APIVersion: clusterv1.GroupVersion.String(),
							Kind:       "Cluster",
							Name:       "test-cluster",
							UID:        "test-uid",
						},
					},
				},
				Status: controlplanev1beta1.CAPTControlPlaneStatus{
					WorkspaceTemplateStatus: &controlplanev1beta1.WorkspaceTemplateStatus{},
				},
			},
			cluster: &clusterv1.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-cluster",
					Namespace: "default",
					UID:       "test-uid",
				},
			},
			expectedError: false,
			validate: func(t *testing.T, controlPlane *controlplanev1beta1.CAPTControlPlane) {
				assert.Len(t, controlPlane.OwnerReferences, 1)
				assert.Equal(t, "Cluster", controlPlane.OwnerReferences[0].Kind)
				assert.Equal(t, "test-cluster", controlPlane.OwnerReferences[0].Name)
				assert.Equal(t, "test-uid", string(controlPlane.OwnerReferences[0].UID))
			},
		},
		{
			name: "Nil cluster",
			controlPlane: &controlplanev1beta1.CAPTControlPlane{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-controlplane",
					Namespace: "default",
				},
				Status: controlplanev1beta1.CAPTControlPlaneStatus{
					WorkspaceTemplateStatus: &controlplanev1beta1.WorkspaceTemplateStatus{},
				},
			},
			cluster:       nil,
			expectedError: false,
			validate: func(t *testing.T, controlPlane *controlplanev1beta1.CAPTControlPlane) {
				assert.Empty(t, controlPlane.OwnerReferences)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := fake.NewClientBuilder().
				WithScheme(scheme).
				WithObjects(tt.controlPlane).
				WithStatusSubresource(tt.controlPlane).
				Build()

			r := &Reconciler{
				Client: client,
				Scheme: scheme,
			}

			err := r.setOwnerReference(context.Background(), tt.controlPlane, tt.cluster)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if tt.validate != nil {
				tt.validate(t, tt.controlPlane)
			}
		})
	}
}
