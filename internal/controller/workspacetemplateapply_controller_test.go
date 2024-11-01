package controller

import (
	"context"
	"testing"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	tfv1beta1 "github.com/upbound/provider-terraform/apis/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/appthrust/capt/api/v1beta1"
)

func TestWaitForDependentWorkspaces(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = v1beta1.AddToScheme(scheme)
	_ = tfv1beta1.SchemeBuilder.AddToScheme(scheme)

	tests := []struct {
		name          string
		cr            *v1beta1.WorkspaceTemplateApply
		workspaces    []runtime.Object
		expectedError bool
	}{
		{
			name: "no dependent workspaces",
			cr: &v1beta1.WorkspaceTemplateApply{
				Spec: v1beta1.WorkspaceTemplateApplySpec{},
			},
			workspaces:    []runtime.Object{},
			expectedError: false,
		},
		{
			name: "dependent workspace not found",
			cr: &v1beta1.WorkspaceTemplateApply{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
				},
				Spec: v1beta1.WorkspaceTemplateApplySpec{
					WaitForWorkspaces: []v1beta1.WorkspaceReference{
						{
							Name: "non-existent",
						},
					},
				},
			},
			workspaces:    []runtime.Object{},
			expectedError: true,
		},
		{
			name: "dependent workspace not ready",
			cr: &v1beta1.WorkspaceTemplateApply{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
				},
				Spec: v1beta1.WorkspaceTemplateApplySpec{
					WaitForWorkspaces: []v1beta1.WorkspaceReference{
						{
							Name: "test-workspace",
						},
					},
				},
			},
			workspaces: []runtime.Object{
				&tfv1beta1.Workspace{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-workspace",
						Namespace: "default",
					},
					Status: tfv1beta1.WorkspaceStatus{
						ResourceStatus: xpv1.ResourceStatus{
							ConditionedStatus: xpv1.ConditionedStatus{
								Conditions: []xpv1.Condition{
									{
										Type:   xpv1.TypeReady,
										Status: corev1.ConditionFalse,
									},
								},
							},
						},
					},
				},
			},
			expectedError: true,
		},
		{
			name: "dependent workspace ready",
			cr: &v1beta1.WorkspaceTemplateApply{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
				},
				Spec: v1beta1.WorkspaceTemplateApplySpec{
					WaitForWorkspaces: []v1beta1.WorkspaceReference{
						{
							Name: "test-workspace",
						},
					},
				},
			},
			workspaces: []runtime.Object{
				&tfv1beta1.Workspace{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-workspace",
						Namespace: "default",
					},
					Status: tfv1beta1.WorkspaceStatus{
						ResourceStatus: xpv1.ResourceStatus{
							ConditionedStatus: xpv1.ConditionedStatus{
								Conditions: []xpv1.Condition{
									{
										Type:   xpv1.TypeReady,
										Status: corev1.ConditionTrue,
									},
								},
							},
						},
					},
				},
			},
			expectedError: false,
		},
		{
			name: "multiple dependent workspaces ready",
			cr: &v1beta1.WorkspaceTemplateApply{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
				},
				Spec: v1beta1.WorkspaceTemplateApplySpec{
					WaitForWorkspaces: []v1beta1.WorkspaceReference{
						{
							Name: "test-workspace-1",
						},
						{
							Name: "test-workspace-2",
						},
					},
				},
			},
			workspaces: []runtime.Object{
				&tfv1beta1.Workspace{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-workspace-1",
						Namespace: "default",
					},
					Status: tfv1beta1.WorkspaceStatus{
						ResourceStatus: xpv1.ResourceStatus{
							ConditionedStatus: xpv1.ConditionedStatus{
								Conditions: []xpv1.Condition{
									{
										Type:   xpv1.TypeReady,
										Status: corev1.ConditionTrue,
									},
								},
							},
						},
					},
				},
				&tfv1beta1.Workspace{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-workspace-2",
						Namespace: "default",
					},
					Status: tfv1beta1.WorkspaceStatus{
						ResourceStatus: xpv1.ResourceStatus{
							ConditionedStatus: xpv1.ConditionedStatus{
								Conditions: []xpv1.Condition{
									{
										Type:   xpv1.TypeReady,
										Status: corev1.ConditionTrue,
									},
								},
							},
						},
					},
				},
			},
			expectedError: false,
		},
		{
			name: "one of multiple workspaces not ready",
			cr: &v1beta1.WorkspaceTemplateApply{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
				},
				Spec: v1beta1.WorkspaceTemplateApplySpec{
					WaitForWorkspaces: []v1beta1.WorkspaceReference{
						{
							Name: "test-workspace-1",
						},
						{
							Name: "test-workspace-2",
						},
					},
				},
			},
			workspaces: []runtime.Object{
				&tfv1beta1.Workspace{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-workspace-1",
						Namespace: "default",
					},
					Status: tfv1beta1.WorkspaceStatus{
						ResourceStatus: xpv1.ResourceStatus{
							ConditionedStatus: xpv1.ConditionedStatus{
								Conditions: []xpv1.Condition{
									{
										Type:   xpv1.TypeReady,
										Status: corev1.ConditionTrue,
									},
								},
							},
						},
					},
				},
				&tfv1beta1.Workspace{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-workspace-2",
						Namespace: "default",
					},
					Status: tfv1beta1.WorkspaceStatus{
						ResourceStatus: xpv1.ResourceStatus{
							ConditionedStatus: xpv1.ConditionedStatus{
								Conditions: []xpv1.Condition{
									{
										Type:   xpv1.TypeReady,
										Status: corev1.ConditionFalse,
									},
								},
							},
						},
					},
				},
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := fake.NewClientBuilder().
				WithScheme(scheme).
				WithRuntimeObjects(tt.workspaces...).
				Build()

			r := &workspaceTemplateApplyReconciler{
				client: client,
				log:    logging.NewNopLogger(),
				record: event.NewNopRecorder(),
			}

			err := r.waitForDependentWorkspaces(context.Background(), tt.cr)
			if (err != nil) != tt.expectedError {
				t.Errorf("waitForDependentWorkspaces() error = %v, expectedError %v", err, tt.expectedError)
			}
		})
	}
}
