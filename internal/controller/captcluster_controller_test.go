package controller

import (
	"context"
	"testing"
	"time"

	infrastructurev1beta1 "github.com/appthrust/capt/api/v1beta1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("CAPTCluster Controller", func() {
	const (
		ClusterName      = "test-cluster"
		ClusterNamespace = "default"
		timeout          = time.Second * 10
		interval         = time.Millisecond * 250
	)

	Context("When creating CAPTCluster", func() {
		It("Should create WorkspaceTemplateApply for VPC when VPCTemplateRef is specified", func() {
			ctx := context.Background()

			// Create VPC WorkspaceTemplate
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
					},
				},
			}
			Expect(k8sClient.Create(ctx, vpcTemplate)).Should(Succeed())

			// Create CAPTCluster
			cluster := &infrastructurev1beta1.CAPTCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      ClusterName,
					Namespace: ClusterNamespace,
				},
				Spec: infrastructurev1beta1.CAPTClusterSpec{
					Region: "us-west-2",
					VPCTemplateRef: &infrastructurev1beta1.WorkspaceTemplateReference{
						Name:      "vpc-template",
						Namespace: ClusterNamespace,
					},
				},
			}
			Expect(k8sClient.Create(ctx, cluster)).Should(Succeed())

			// Verify WorkspaceTemplateApply is created
			vpcApplyName := types.NamespacedName{
				Name:      ClusterName + "-vpc",
				Namespace: ClusterNamespace,
			}
			createdVPCApply := &infrastructurev1beta1.WorkspaceTemplateApply{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, vpcApplyName, createdVPCApply)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			// Verify WorkspaceTemplateApply properties
			Expect(createdVPCApply.Spec.TemplateRef.Name).Should(Equal("vpc-template"))
			Expect(createdVPCApply.Spec.Variables["name"]).Should(Equal(ClusterName + "-vpc"))
		})
	})
})

func TestController(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Controller Suite")
}
