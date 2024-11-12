package controlplane

import (
	"context"
	"fmt"
	"testing"

	controlplanev1beta1 "github.com/appthrust/capt/api/controlplane/v1beta1"
	infrastructurev1beta1 "github.com/appthrust/capt/api/v1beta1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestReconcileWorkspace(t *testing.T) {
	scheme := setupScheme()

	tests := []struct {
		name           string
		controlPlane   *controlplanev1beta1.CAPTControlPlane
		template       *infrastructurev1beta1.WorkspaceTemplate
		cluster        *clusterv1.Cluster
		workspaceApply *infrastructurev1beta1.WorkspaceTemplateApply
		vpcWorkspace   *infrastructurev1beta1.WorkspaceTemplateApply
		expectedError  bool
		expectedResult ctrl.Result
		validate       func(t *testing.T, c client.Client)
	}{
		{
			name: "Successfully reconcile workspace without VPC dependency",
			controlPlane: &controlplanev1beta1.CAPTControlPlane{
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
			template: &infrastructurev1beta1.WorkspaceTemplate{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-template",
					Namespace: "default",
				},
				Spec: infrastructurev1beta1.WorkspaceTemplateSpec{},
			},
			cluster: &clusterv1.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-controlplane",
					Namespace: "default",
				},
			},
			expectedError:  false,
			expectedResult: ctrl.Result{RequeueAfter: requeueInterval},
			validate: func(t *testing.T, c client.Client) {
				workspaceApply := &infrastructurev1beta1.WorkspaceTemplateApply{}
				err := c.Get(context.Background(), types.NamespacedName{
					Name:      "test-controlplane-eks-controlplane-apply",
					Namespace: "default",
				}, workspaceApply)
				assert.NoError(t, err)
				assert.Equal(t, "test-template", workspaceApply.Spec.TemplateRef.Name)
				assert.Empty(t, workspaceApply.Spec.WaitForWorkspaces, "Should not have VPC dependency")
				assert.NotNil(t, workspaceApply.Spec.WriteConnectionSecretToRef, "Should have connection secret ref")
				assert.Equal(t, "test-controlplane-eks-connection", workspaceApply.Spec.WriteConnectionSecretToRef.Name)
			},
		},
		{
			name: "Successfully reconcile workspace with VPC dependency",
			controlPlane: &controlplanev1beta1.CAPTControlPlane{
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
			template: &infrastructurev1beta1.WorkspaceTemplate{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-template",
					Namespace: "default",
				},
				Spec: infrastructurev1beta1.WorkspaceTemplateSpec{},
			},
			vpcWorkspace: &infrastructurev1beta1.WorkspaceTemplateApply{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-controlplane-vpc",
					Namespace: "default",
				},
			},
			expectedError:  false,
			expectedResult: ctrl.Result{RequeueAfter: requeueInterval},
			validate: func(t *testing.T, c client.Client) {
				workspaceApply := &infrastructurev1beta1.WorkspaceTemplateApply{}
				err := c.Get(context.Background(), types.NamespacedName{
					Name:      "test-controlplane-eks-controlplane-apply",
					Namespace: "default",
				}, workspaceApply)
				assert.NoError(t, err)
				assert.Equal(t, "test-template", workspaceApply.Spec.TemplateRef.Name)
				assert.Len(t, workspaceApply.Spec.WaitForWorkspaces, 1, "Should have VPC dependency")
				assert.Equal(t, "test-controlplane-vpc", workspaceApply.Spec.WaitForWorkspaces[0].Name)
			},
		},
		{
			name: "Template not found",
			controlPlane: &controlplanev1beta1.CAPTControlPlane{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-controlplane",
					Namespace: "default",
				},
				Spec: controlplanev1beta1.CAPTControlPlaneSpec{
					WorkspaceTemplateRef: controlplanev1beta1.WorkspaceTemplateReference{
						Name:      "non-existent-template",
						Namespace: "default",
					},
				},
				Status: controlplanev1beta1.CAPTControlPlaneStatus{
					WorkspaceTemplateStatus: &controlplanev1beta1.WorkspaceTemplateStatus{},
				},
			},
			template:       nil,
			cluster:        nil,
			expectedError:  true,
			expectedResult: ctrl.Result{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var objects []runtime.Object
			objects = append(objects, tt.controlPlane)
			if tt.template != nil {
				objects = append(objects, tt.template)
			}
			if tt.cluster != nil {
				objects = append(objects, tt.cluster)
			}
			if tt.workspaceApply != nil {
				objects = append(objects, tt.workspaceApply)
			}
			if tt.vpcWorkspace != nil {
				objects = append(objects, tt.vpcWorkspace)
			}

			client := fake.NewClientBuilder().
				WithScheme(scheme).
				WithRuntimeObjects(objects...).
				WithStatusSubresource(&controlplanev1beta1.CAPTControlPlane{}, &clusterv1.Cluster{}).
				Build()

			r := &Reconciler{
				Client: client,
				Scheme: scheme,
			}

			result, err := r.reconcileWorkspace(context.Background(), tt.controlPlane, tt.cluster)
			if tt.expectedError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "failed to get WorkspaceTemplate")
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.expectedResult, result)

			if tt.validate != nil {
				tt.validate(t, client)
			}
		})
	}
}

func TestGetOrCreateWorkspaceTemplateApply(t *testing.T) {
	scheme := setupScheme()

	tests := []struct {
		name          string
		controlPlane  *controlplanev1beta1.CAPTControlPlane
		template      *infrastructurev1beta1.WorkspaceTemplate
		existingApply *infrastructurev1beta1.WorkspaceTemplateApply
		vpcWorkspace  *infrastructurev1beta1.WorkspaceTemplateApply
		expectCreate  bool
		validate      func(t *testing.T, workspaceApply *infrastructurev1beta1.WorkspaceTemplateApply)
	}{
		{
			name: "Create new WorkspaceTemplateApply without VPC dependency",
			controlPlane: &controlplanev1beta1.CAPTControlPlane{
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
			},
			template: &infrastructurev1beta1.WorkspaceTemplate{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-template",
					Namespace: "default",
				},
			},
			existingApply: nil,
			expectCreate:  true,
			validate: func(t *testing.T, workspaceApply *infrastructurev1beta1.WorkspaceTemplateApply) {
				assert.NotEmpty(t, workspaceApply.Name)
				assert.Equal(t, "default", workspaceApply.Namespace)
				assert.Equal(t, "test-template", workspaceApply.Spec.TemplateRef.Name)
				assert.Equal(t, "1.21", workspaceApply.Spec.Variables["kubernetes_version"])
				assert.Empty(t, workspaceApply.Spec.WaitForWorkspaces, "Should not have VPC dependency")
				assert.NotNil(t, workspaceApply.Spec.WriteConnectionSecretToRef)
			},
		},
		{
			name: "Create new WorkspaceTemplateApply with VPC dependency",
			controlPlane: &controlplanev1beta1.CAPTControlPlane{
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
			},
			template: &infrastructurev1beta1.WorkspaceTemplate{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-template",
					Namespace: "default",
				},
			},
			vpcWorkspace: &infrastructurev1beta1.WorkspaceTemplateApply{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-controlplane-vpc",
					Namespace: "default",
				},
			},
			existingApply: nil,
			expectCreate:  true,
			validate: func(t *testing.T, workspaceApply *infrastructurev1beta1.WorkspaceTemplateApply) {
				assert.NotEmpty(t, workspaceApply.Name)
				assert.Equal(t, "default", workspaceApply.Namespace)
				assert.Equal(t, "test-template", workspaceApply.Spec.TemplateRef.Name)
				assert.Equal(t, "1.21", workspaceApply.Spec.Variables["kubernetes_version"])
				assert.Len(t, workspaceApply.Spec.WaitForWorkspaces, 1, "Should have VPC dependency")
				assert.Equal(t, "test-controlplane-vpc", workspaceApply.Spec.WaitForWorkspaces[0].Name)
				assert.NotNil(t, workspaceApply.Spec.WriteConnectionSecretToRef)
			},
		},
		{
			name: "Update existing WorkspaceTemplateApply",
			controlPlane: &controlplanev1beta1.CAPTControlPlane{
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
					WorkspaceTemplateApplyName: "test-apply",
				},
			},
			template: &infrastructurev1beta1.WorkspaceTemplate{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-template",
					Namespace: "default",
				},
			},
			existingApply: &infrastructurev1beta1.WorkspaceTemplateApply{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-apply",
					Namespace: "default",
				},
			},
			expectCreate: false,
			validate: func(t *testing.T, workspaceApply *infrastructurev1beta1.WorkspaceTemplateApply) {
				assert.Equal(t, "test-apply", workspaceApply.Name)
				assert.Equal(t, "default", workspaceApply.Namespace)
				assert.Equal(t, "test-template", workspaceApply.Spec.TemplateRef.Name)
				assert.Equal(t, "1.21", workspaceApply.Spec.Variables["kubernetes_version"])
				assert.NotNil(t, workspaceApply.Spec.WriteConnectionSecretToRef)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var objects []runtime.Object
			objects = append(objects, tt.controlPlane, tt.template)
			if tt.existingApply != nil {
				objects = append(objects, tt.existingApply)
			}
			if tt.vpcWorkspace != nil {
				objects = append(objects, tt.vpcWorkspace)
			}

			client := fake.NewClientBuilder().
				WithScheme(scheme).
				WithRuntimeObjects(objects...).
				WithStatusSubresource(&controlplanev1beta1.CAPTControlPlane{}).
				Build()

			r := &Reconciler{
				Client: client,
				Scheme: scheme,
			}

			workspaceApply, err := r.getOrCreateWorkspaceTemplateApply(context.Background(), tt.controlPlane, tt.template)
			assert.NoError(t, err)
			assert.NotNil(t, workspaceApply)

			if tt.validate != nil {
				tt.validate(t, workspaceApply)
			}
		})
	}
}

func TestGenerateWorkspaceTemplateApplySpec(t *testing.T) {
	scheme := setupScheme()

	tests := []struct {
		name         string
		controlPlane *controlplanev1beta1.CAPTControlPlane
		expectedVars map[string]string
	}{
		{
			name: "Generate spec with basic configuration",
			controlPlane: &controlplanev1beta1.CAPTControlPlane{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-controlplane",
				},
				Spec: controlplanev1beta1.CAPTControlPlaneSpec{
					Version: "1.21",
					WorkspaceTemplateRef: controlplanev1beta1.WorkspaceTemplateReference{
						Name: "test-template",
					},
				},
			},
			expectedVars: map[string]string{
				"cluster_name":       "test-controlplane",
				"kubernetes_version": "1.21",
			},
		},
		{
			name: "Generate spec with endpoint access",
			controlPlane: &controlplanev1beta1.CAPTControlPlane{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-controlplane",
				},
				Spec: controlplanev1beta1.CAPTControlPlaneSpec{
					Version: "1.21",
					WorkspaceTemplateRef: controlplanev1beta1.WorkspaceTemplateReference{
						Name: "test-template",
					},
					ControlPlaneConfig: &controlplanev1beta1.ControlPlaneConfig{
						EndpointAccess: &controlplanev1beta1.EndpointAccess{
							Public:  true,
							Private: false,
						},
					},
				},
			},
			expectedVars: map[string]string{
				"cluster_name":            "test-controlplane",
				"kubernetes_version":      "1.21",
				"endpoint_public_access":  "true",
				"endpoint_private_access": "false",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := fake.NewClientBuilder().
				WithScheme(scheme).
				Build()

			r := &Reconciler{
				Client: client,
				Scheme: scheme,
			}

			spec := r.generateWorkspaceTemplateApplySpec(tt.controlPlane)

			assert.Equal(t, tt.controlPlane.Spec.WorkspaceTemplateRef.Name, spec.TemplateRef.Name)
			for k, v := range tt.expectedVars {
				assert.Equal(t, v, spec.Variables[k])
			}
			assert.NotNil(t, spec.WriteConnectionSecretToRef, "Should have connection secret ref")
			assert.Equal(t, fmt.Sprintf("%s-eks-connection", tt.controlPlane.Name), spec.WriteConnectionSecretToRef.Name)
		})
	}
}
