package controlplane

import (
	"context"
	"testing"

	controlplanev1beta1 "github.com/appthrust/capt/api/controlplane/v1beta1"
	infrastructurev1beta1 "github.com/appthrust/capt/api/v1beta1"
	"github.com/appthrust/capt/internal/controller/controlplane/secrets"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestReconcileSecrets(t *testing.T) {
	scheme := setupScheme()

	tests := []struct {
		name           string
		controlPlane   *controlplanev1beta1.CAPTControlPlane
		cluster        *clusterv1.Cluster
		workspaceApply *infrastructurev1beta1.WorkspaceTemplateApply
		workspace      *unstructured.Unstructured
		secret         *corev1.Secret
		expectedError  bool
		validate       func(t *testing.T, controlPlane *controlplanev1beta1.CAPTControlPlane, client fake.ClientBuilder)
	}{
		{
			name: "Successfully reconcile secrets",
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
				Status: clusterv1.ClusterStatus{},
			},
			workspaceApply: &infrastructurev1beta1.WorkspaceTemplateApply{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-workspace-apply",
					Namespace: "default",
				},
				Status: infrastructurev1beta1.WorkspaceTemplateApplyStatus{
					WorkspaceName: "test-workspace",
					Applied:       true,
					Conditions: []xpv1.Condition{
						{
							Type:   xpv1.TypeReady,
							Status: corev1.ConditionTrue,
						},
					},
				},
				Spec: infrastructurev1beta1.WorkspaceTemplateApplySpec{
					WriteConnectionSecretToRef: &xpv1.SecretReference{
						Name:      "test-connection-secret",
						Namespace: "default",
					},
				},
			},
			workspace: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "tf.upbound.io/v1beta1",
					"kind":       "Workspace",
					"metadata": map[string]interface{}{
						"name":      "test-workspace",
						"namespace": "default",
					},
					"status": map[string]interface{}{
						"outputs": map[string]interface{}{
							"endpoint": map[string]interface{}{
								"type":  "string",
								"value": "https://test-endpoint:6443",
							},
						},
					},
				},
			},
			secret: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-connection-secret",
					Namespace: "default",
				},
				Data: map[string][]byte{
					"kubeconfig": []byte("test-kubeconfig"),
					"ca.crt":     []byte("test-ca-data"),
					"endpoint":   []byte("https://test-endpoint:6443"),
				},
			},
			expectedError: false,
			validate: func(t *testing.T, controlPlane *controlplanev1beta1.CAPTControlPlane, client fake.ClientBuilder) {
				// Create fake client with test objects
				c := client.Build()

				// Verify kubeconfig secret was created
				kubeconfigSecret := &corev1.Secret{}
				err := c.Get(context.Background(), types.NamespacedName{
					Name:      controlPlane.Name + "-kubeconfig",
					Namespace: controlPlane.Namespace,
				}, kubeconfigSecret)
				assert.NoError(t, err)
				assert.Equal(t, []byte("test-kubeconfig"), kubeconfigSecret.Data["value"])

				// Verify CA secret was created
				caSecret := &corev1.Secret{}
				err = c.Get(context.Background(), types.NamespacedName{
					Name:      controlPlane.Name + "-ca",
					Namespace: controlPlane.Namespace,
				}, caSecret)
				assert.NoError(t, err)
				assert.Equal(t, []byte("test-ca-data"), caSecret.Data["tls.crt"])
				assert.Equal(t, []byte("test-ca-data"), caSecret.Data["ca.crt"])

				// Verify endpoint was set in status
				assert.Equal(t, "test-endpoint", controlPlane.Spec.ControlPlaneEndpoint.Host)
				assert.Equal(t, int32(6443), controlPlane.Spec.ControlPlaneEndpoint.Port)
			},
		},
		{
			name: "Missing workspace name",
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
				Status: clusterv1.ClusterStatus{},
			},
			workspaceApply: &infrastructurev1beta1.WorkspaceTemplateApply{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-workspace-apply",
					Namespace: "default",
				},
				Status: infrastructurev1beta1.WorkspaceTemplateApplyStatus{
					WorkspaceName: "",
				},
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var objects []runtime.Object
			objects = append(objects, tt.controlPlane)
			if tt.cluster != nil {
				objects = append(objects, tt.cluster)
			}
			if tt.workspaceApply != nil {
				objects = append(objects, tt.workspaceApply)
			}
			if tt.secret != nil {
				objects = append(objects, tt.secret)
			}
			if tt.workspace != nil {
				objects = append(objects, tt.workspace)
			}

			// Create fake client with test objects and status subresource
			client := fake.NewClientBuilder().
				WithScheme(scheme).
				WithRuntimeObjects(objects...).
				WithStatusSubresource(&controlplanev1beta1.CAPTControlPlane{}, &clusterv1.Cluster{}).
				Build()

			r := &Reconciler{
				Client: client,
				Scheme: scheme,
			}

			err := r.reconcileSecrets(context.Background(), tt.controlPlane, tt.cluster, tt.workspaceApply)

			if tt.expectedError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)

			if tt.validate != nil {
				// Create new client with updated objects for validation
				updatedObjects := []runtime.Object{tt.controlPlane}
				if tt.cluster != nil {
					updatedObjects = append(updatedObjects, tt.cluster)
				}
				if tt.secret != nil {
					updatedObjects = append(updatedObjects, tt.secret)
				}

				// Add expected secrets
				kubeconfigSecret := &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      tt.controlPlane.Name + "-kubeconfig",
						Namespace: tt.controlPlane.Namespace,
					},
					Data: map[string][]byte{
						"value": []byte("test-kubeconfig"),
					},
				}
				caSecret := &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      tt.controlPlane.Name + "-ca",
						Namespace: tt.controlPlane.Namespace,
					},
					Data: map[string][]byte{
						"tls.crt": []byte("test-ca-data"),
						"ca.crt":  []byte("test-ca-data"),
					},
				}
				updatedObjects = append(updatedObjects, kubeconfigSecret, caSecret)

				tt.validate(t, tt.controlPlane, *fake.NewClientBuilder().
					WithScheme(scheme).
					WithRuntimeObjects(updatedObjects...).
					WithStatusSubresource(&controlplanev1beta1.CAPTControlPlane{}, &clusterv1.Cluster{}))
			}
		})
	}
}

func TestSecretManager(t *testing.T) {
	scheme := setupScheme()

	tests := []struct {
		name          string
		workspace     *unstructured.Unstructured
		secret        *corev1.Secret
		expectedHost  string
		expectedPort  int32
		expectedError bool
	}{
		{
			name: "Get endpoint from workspace outputs",
			workspace: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"status": map[string]interface{}{
						"outputs": map[string]interface{}{
							"endpoint": map[string]interface{}{
								"value": "https://test-endpoint:6443",
							},
						},
					},
				},
			},
			secret: &corev1.Secret{
				Data: map[string][]byte{
					"ca.crt": []byte("test-ca-data"),
				},
			},
			expectedHost:  "test-endpoint",
			expectedPort:  6443,
			expectedError: false,
		},
		{
			name: "Get endpoint from secret",
			workspace: &unstructured.Unstructured{
				Object: map[string]interface{}{},
			},
			secret: &corev1.Secret{
				Data: map[string][]byte{
					"endpoint": []byte("https://test-endpoint:6443"),
					"ca.crt":   []byte("test-ca-data"),
				},
			},
			expectedHost:  "test-endpoint",
			expectedPort:  6443,
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := fake.NewClientBuilder().
				WithScheme(scheme).
				Build()

			manager := secrets.NewManager(client)

			endpoint, err := manager.GetClusterEndpoint(context.Background(), tt.workspace, tt.secret)
			if tt.expectedError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedHost, endpoint.Host)
			assert.Equal(t, tt.expectedPort, endpoint.Port)

			caData, err := manager.GetCertificateAuthorityData(context.Background(), tt.secret)
			assert.NoError(t, err)
			assert.Equal(t, "test-ca-data", caData)
		})
	}
}
