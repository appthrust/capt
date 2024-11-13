package captcluster

import (
	"time"

	infrastructurev1beta1 "github.com/appthrust/capt/api/v1beta1"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	tfv1beta1 "github.com/upbound/provider-terraform/apis/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
)

var _ = Describe("CAPTCluster Controller", func() {
	const (
		ClusterName      = "test-cluster"
		ClusterNamespace = "default"
		timeout          = time.Second * 30
		interval         = time.Second * 1
	)

	Context("When creating CAPTCluster", func() {
		It("Should create WorkspaceTemplateApply for VPC when VPCTemplateRef is specified", func() {
			By("Creating Cluster")
			cluster := &clusterv1.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      ClusterName,
					Namespace: ClusterNamespace,
				},
				Spec: clusterv1.ClusterSpec{},
			}
			Expect(k8sClient.Create(ctx, cluster)).Should(Succeed())

			By("Creating VPC WorkspaceTemplate")
			vpcTemplate := &infrastructurev1beta1.WorkspaceTemplate{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "vpc-template",
					Namespace: ClusterNamespace,
				},
				Spec: infrastructurev1beta1.WorkspaceTemplateSpec{
					Template: infrastructurev1beta1.WorkspaceTemplateDefinition{
						Metadata: &infrastructurev1beta1.WorkspaceTemplateMetadata{
							Description: "VPC template for testing",
							Version:     "1.0.0",
						},
						Spec: tfv1beta1.WorkspaceSpec{
							ForProvider: tfv1beta1.WorkspaceParameters{
								Source: "Inline",
							},
						},
					},
				},
			}
			Expect(k8sClient.Create(ctx, vpcTemplate)).Should(Succeed())

			By("Creating CAPTCluster")
			captCluster := &infrastructurev1beta1.CAPTCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      ClusterName,
					Namespace: ClusterNamespace,
					OwnerReferences: []metav1.OwnerReference{
						{
							APIVersion: clusterv1.GroupVersion.String(),
							Kind:       "Cluster",
							Name:       cluster.Name,
							UID:        cluster.UID,
						},
					},
				},
				Spec: infrastructurev1beta1.CAPTClusterSpec{
					Region: "us-west-2",
					VPCTemplateRef: &infrastructurev1beta1.WorkspaceTemplateReference{
						Name:      "vpc-template",
						Namespace: ClusterNamespace,
					},
				},
				Status: infrastructurev1beta1.CAPTClusterStatus{
					WorkspaceTemplateStatus: &infrastructurev1beta1.CAPTClusterWorkspaceStatus{},
				},
			}
			Expect(k8sClient.Create(ctx, captCluster)).Should(Succeed())

			By("Verifying WorkspaceTemplateApply creation")
			vpcApplyName := types.NamespacedName{
				Name:      ClusterName + "-vpc",
				Namespace: ClusterNamespace,
			}
			createdVPCApply := &infrastructurev1beta1.WorkspaceTemplateApply{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, vpcApplyName, createdVPCApply)
				if err != nil {
					By("WorkspaceTemplateApply not found yet")
					return false
				}
				By("WorkspaceTemplateApply found")
				return true
			}, timeout, interval).Should(BeTrue())

			By("Verifying WorkspaceTemplateApply properties")
			Expect(createdVPCApply.Spec.TemplateRef.Name).Should(Equal("vpc-template"))
			Expect(createdVPCApply.Spec.Variables["name"]).Should(Equal(ClusterName + "-vpc"))

			By("Creating Terraform Workspace")
			workspace := &unstructured.Unstructured{}
			workspace.SetGroupVersionKind(schema.GroupVersionKind{
				Group:   "tf.upbound.io",
				Version: "v1beta1",
				Kind:    "Workspace",
			})
			workspace.SetName(ClusterName + "-vpc-workspace")
			workspace.SetNamespace(ClusterNamespace)
			workspace.Object["spec"] = map[string]interface{}{
				"forProvider": map[string]interface{}{
					"source": "Inline",
					"module": "./modules/vpc",
					"variables": map[string]interface{}{
						"name":   ClusterName + "-vpc",
						"region": "us-west-2",
					},
				},
			}
			Expect(k8sClient.Create(ctx, workspace)).Should(Succeed())

			By("Updating WorkspaceTemplateApply status")
			now := metav1.Now()
			createdVPCApply.Status.Applied = true
			createdVPCApply.Status.WorkspaceName = ClusterName + "-vpc-workspace"
			createdVPCApply.Status.LastAppliedTime = &now
			createdVPCApply.Status.Conditions = []xpv1.Condition{
				{
					Type:               xpv1.TypeSynced,
					Status:             corev1.ConditionTrue,
					LastTransitionTime: now,
					Reason:             "Synced",
					Message:            "Workspace is synced",
				},
				{
					Type:               xpv1.TypeReady,
					Status:             corev1.ConditionTrue,
					LastTransitionTime: now,
					Reason:             "Ready",
					Message:            "Workspace is ready",
				},
			}
			Expect(k8sClient.Status().Update(ctx, createdVPCApply)).Should(Succeed())

			By("Updating Workspace status")
			workspace.Object["status"] = map[string]interface{}{
				"atProvider": map[string]interface{}{
					"outputs": map[string]interface{}{
						"vpc_id": "vpc-12345",
					},
				},
				"conditions": []interface{}{
					map[string]interface{}{
						"type":               "Ready",
						"status":             "True",
						"lastTransitionTime": metav1.Now().Format(time.RFC3339),
						"reason":             "Available",
						"message":            "Workspace is ready",
					},
				},
			}
			Expect(k8sClient.Status().Update(ctx, workspace)).Should(Succeed())

			By("Verifying CAPTCluster status update")
			Eventually(func() bool {
				err := k8sClient.Get(ctx, types.NamespacedName{
					Name:      ClusterName,
					Namespace: ClusterNamespace,
				}, captCluster)
				if err != nil {
					return false
				}
				return captCluster.Status.WorkspaceTemplateStatus != nil &&
					captCluster.Status.WorkspaceTemplateStatus.Ready &&
					captCluster.Status.VPCID == "vpc-12345"
			}, timeout, interval).Should(BeTrue())
		})
	})
})
