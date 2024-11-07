/*
Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controlplane

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	tfv1beta1 "github.com/upbound/provider-terraform/apis/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	controlplanev1beta1 "github.com/appthrust/capt/api/controlplane/v1beta1"
	infrastructurev1beta1 "github.com/appthrust/capt/api/v1beta1"
)

var _ = Describe("CAPTControlPlane Controller", func() {
	Context("When reconciling a resource", func() {
		const resourceName = "test-resource"
		const namespace = "default"

		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: namespace,
		}
		captcontrolplane := &controlplanev1beta1.CAPTControlPlane{}

		BeforeEach(func() {
			By("creating the custom resource for the Kind CAPTControlPlane")
			err := k8sClient.Get(ctx, typeNamespacedName, captcontrolplane)
			if err != nil && errors.IsNotFound(err) {
				// Create WorkspaceTemplate
				template := &infrastructurev1beta1.WorkspaceTemplate{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-template",
						Namespace: namespace,
					},
					Spec: infrastructurev1beta1.WorkspaceTemplateSpec{
						Template: infrastructurev1beta1.WorkspaceTemplateDefinition{
							Spec: tfv1beta1.WorkspaceSpec{
								ForProvider: tfv1beta1.WorkspaceParameters{
									Source: "Inline",
									Module: "test module",
								},
							},
						},
					},
				}
				Expect(k8sClient.Create(ctx, template)).To(Succeed())

				// Create CAPTControlPlane
				resource := &controlplanev1beta1.CAPTControlPlane{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceName,
						Namespace: namespace,
					},
					Spec: controlplanev1beta1.CAPTControlPlaneSpec{
						Version: "1.24",
						WorkspaceTemplateRef: controlplanev1beta1.WorkspaceTemplateReference{
							Name:      "test-template",
							Namespace: namespace,
						},
					},
				}
				Expect(k8sClient.Create(ctx, resource)).To(Succeed())

				// Create Secret
				secret := &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      fmt.Sprintf("%s-eks-connection", resourceName),
						Namespace: namespace,
					},
					Data: map[string][]byte{
						"kubeconfig":                         []byte("test-kubeconfig"),
						"cluster_certificate_authority_data": []byte("test-ca-data"),
						"cluster_endpoint":                   []byte("https://test-endpoint"),
					},
				}
				Expect(k8sClient.Create(ctx, secret)).To(Succeed())

				// Create Workspace
				workspace := &unstructured.Unstructured{}
				workspace.SetGroupVersionKind(schema.GroupVersionKind{
					Group:   "tf.upbound.io",
					Version: "v1beta1",
					Kind:    "Workspace",
				})
				workspace.SetName(fmt.Sprintf("%s-eks-controlplane", resourceName))
				workspace.SetNamespace(namespace)
				workspace.Object["spec"] = map[string]interface{}{
					"forProvider": map[string]interface{}{
						"source": "Inline",
						"module": "test module",
					},
				}
				workspace.Object["status"] = map[string]interface{}{
					"atProvider": map[string]interface{}{
						"outputs": map[string]interface{}{
							"cluster_endpoint": "https://test-endpoint",
						},
					},
				}
				Expect(k8sClient.Create(ctx, workspace)).To(Succeed())
			}
		})

		AfterEach(func() {
			// Cleanup CAPTControlPlane
			resource := &controlplanev1beta1.CAPTControlPlane{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			if err == nil {
				Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
			}

			// Cleanup WorkspaceTemplate
			template := &infrastructurev1beta1.WorkspaceTemplate{}
			err = k8sClient.Get(ctx, types.NamespacedName{Name: "test-template", Namespace: namespace}, template)
			if err == nil {
				Expect(k8sClient.Delete(ctx, template)).To(Succeed())
			}

			// Cleanup WorkspaceTemplateApply
			apply := &infrastructurev1beta1.WorkspaceTemplateApply{}
			applyName := fmt.Sprintf("%s-eks-controlplane-apply", resourceName)
			err = k8sClient.Get(ctx, types.NamespacedName{Name: applyName, Namespace: namespace}, apply)
			if err == nil {
				Expect(k8sClient.Delete(ctx, apply)).To(Succeed())
			}

			// Cleanup Secret
			secret := &corev1.Secret{}
			err = k8sClient.Get(ctx, types.NamespacedName{Name: fmt.Sprintf("%s-eks-connection", resourceName), Namespace: namespace}, secret)
			if err == nil {
				Expect(k8sClient.Delete(ctx, secret)).To(Succeed())
			}

			// Cleanup Workspace
			workspace := &unstructured.Unstructured{}
			workspace.SetGroupVersionKind(schema.GroupVersionKind{
				Group:   "tf.upbound.io",
				Version: "v1beta1",
				Kind:    "Workspace",
			})
			err = k8sClient.Get(ctx, types.NamespacedName{Name: fmt.Sprintf("%s-eks-controlplane", resourceName), Namespace: namespace}, workspace)
			if err == nil {
				Expect(k8sClient.Delete(ctx, workspace)).To(Succeed())
			}
		})

		It("should set correct WriteConnectionSecretToRef", func() {
			By("Reconciling the created resource")
			controllerReconciler := &CAPTControlPlaneReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			By("Getting WorkspaceTemplateApply")
			apply := &infrastructurev1beta1.WorkspaceTemplateApply{}
			applyName := fmt.Sprintf("%s-eks-controlplane-apply", resourceName)
			err = k8sClient.Get(ctx, types.NamespacedName{Name: applyName, Namespace: namespace}, apply)
			Expect(err).NotTo(HaveOccurred())

			By("Verifying WriteConnectionSecretToRef configuration")
			Expect(apply.Spec.WriteConnectionSecretToRef).NotTo(BeNil())
			Expect(apply.Spec.WriteConnectionSecretToRef.Name).To(Equal(fmt.Sprintf("%s-eks-connection", resourceName)))
			Expect(apply.Spec.WriteConnectionSecretToRef.Namespace).To(Equal(namespace))

			By("Setting WorkspaceTemplateApply status")
			apply.Status.WorkspaceName = fmt.Sprintf("%s-eks-controlplane", resourceName)
			err = k8sClient.Status().Update(ctx, apply)
			Expect(err).NotTo(HaveOccurred())

			_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())
		})

		It("should maintain WriteConnectionSecretToRef after update", func() {
			By("Creating initial WorkspaceTemplateApply")
			controllerReconciler := &CAPTControlPlaneReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			By("Getting WorkspaceTemplateApply")
			apply := &infrastructurev1beta1.WorkspaceTemplateApply{}
			applyName := fmt.Sprintf("%s-eks-controlplane-apply", resourceName)
			err = k8sClient.Get(ctx, types.NamespacedName{Name: applyName, Namespace: namespace}, apply)
			Expect(err).NotTo(HaveOccurred())

			By("Verifying initial WriteConnectionSecretToRef configuration")
			Expect(apply.Spec.WriteConnectionSecretToRef).NotTo(BeNil())
			Expect(apply.Spec.WriteConnectionSecretToRef.Name).To(Equal(fmt.Sprintf("%s-eks-connection", resourceName)))
			Expect(apply.Spec.WriteConnectionSecretToRef.Namespace).To(Equal(namespace))

			By("Setting WorkspaceTemplateApply status")
			apply.Status.WorkspaceName = fmt.Sprintf("%s-eks-controlplane", resourceName)
			err = k8sClient.Status().Update(ctx, apply)
			Expect(err).NotTo(HaveOccurred())

			By("Updating CAPTControlPlane version")
			controlPlane := &controlplanev1beta1.CAPTControlPlane{}
			err = k8sClient.Get(ctx, typeNamespacedName, controlPlane)
			Expect(err).NotTo(HaveOccurred())

			controlPlane.Spec.Version = "1.25"
			err = k8sClient.Update(ctx, controlPlane)
			Expect(err).NotTo(HaveOccurred())

			By("Reconciling the updated resource")
			_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			By("Verifying WriteConnectionSecretToRef remains correct")
			err = k8sClient.Get(ctx, types.NamespacedName{Name: applyName, Namespace: namespace}, apply)
			Expect(err).NotTo(HaveOccurred())

			Expect(apply.Spec.WriteConnectionSecretToRef).NotTo(BeNil())
			Expect(apply.Spec.WriteConnectionSecretToRef.Name).To(Equal(fmt.Sprintf("%s-eks-connection", resourceName)))
			Expect(apply.Spec.WriteConnectionSecretToRef.Namespace).To(Equal(namespace))
		})
	})
})
