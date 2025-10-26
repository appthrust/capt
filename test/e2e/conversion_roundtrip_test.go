package e2e

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var _ = Describe("CRD basic operations (envtest)", Ordered, func() {
	ctx := context.Background()

	It("Create/Get v1beta2 CAPTCluster works", func() {
		infraV1beta2 := schema.GroupVersionResource{
			Group:    "infrastructure.cluster.x-k8s.io",
			Version:  "v1beta2",
			Resource: "captclusters",
		}

		ns := "test-e2e"
		// ensure namespace exists
		_, _ = dynClient.Resource(schema.GroupVersionResource{Group: "", Version: "v1", Resource: "namespaces"}).Create(ctx,
			&unstructured.Unstructured{Object: map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "Namespace",
				"metadata":   map[string]interface{}{"name": ns},
			}}, metav1.CreateOptions{})

		name := fmt.Sprintf("e2e-%d", time.Now().UnixNano())
		u := map[string]interface{}{
			"apiVersion": "infrastructure.cluster.x-k8s.io/v1beta2",
			"kind":       "CAPTCluster",
			"metadata":   map[string]interface{}{"name": name, "namespace": ns},
			"spec":       map[string]interface{}{"region": "us-east-1"},
		}
		created, err := dynClient.Resource(infraV1beta2).Namespace(ns).Create(ctx, toUnstructured(u), metav1.CreateOptions{})
		Expect(err).NotTo(HaveOccurred())
		got, err := dynClient.Resource(infraV1beta2).Namespace(ns).Get(ctx, created.GetName(), metav1.GetOptions{})
		Expect(err).NotTo(HaveOccurred())
		spec, _, _ := unstructured.NestedMap(got.Object, "spec")
		Expect(spec["region"]).To(Equal("us-east-1"))
	})
})

func toUnstructured(m map[string]interface{}) *unstructured.Unstructured {
	return &unstructured.Unstructured{Object: m}
}
