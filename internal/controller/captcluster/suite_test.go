package captcluster

import (
	"context"
	"path/filepath"
	"testing"

	infrastructurev1beta1 "github.com/appthrust/capt/api/v1beta1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	tfv1beta1 "github.com/upbound/provider-terraform/apis/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var (
	cfg       *rest.Config
	k8sClient client.Client
	testEnv   *envtest.Environment
	ctx       context.Context
	cancel    context.CancelFunc
)

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Controller Suite")
}

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	ctx, cancel = context.WithCancel(context.TODO())

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		CRDDirectoryPaths: []string{
			filepath.Join("..", "..", "..", "config", "crd", "bases"),
			filepath.Join("..", "..", "..", "third_party", "cluster-api", "config", "crd", "bases"),
		},
		ErrorIfCRDPathMissing: false, // Set to false to allow missing CRDs
	}

	var err error
	// cfg is defined in this file globally.
	cfg, err = testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	err = infrastructurev1beta1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	err = clusterv1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	// Register Terraform types
	schemeBuilder := runtime.NewSchemeBuilder(func(scheme *runtime.Scheme) error {
		scheme.AddKnownTypes(tfv1beta1.SchemeGroupVersion,
			&tfv1beta1.Workspace{},
			&tfv1beta1.WorkspaceList{},
		)
		metav1.AddToGroupVersion(scheme, tfv1beta1.SchemeGroupVersion)
		return nil
	})
	err = schemeBuilder.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	//+kubebuilder:scaffold:scheme

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())

	// Create Cluster CRD
	clusterCRD := &clusterv1.ClusterList{}
	err = k8sClient.List(ctx, clusterCRD)
	if err != nil {
		By("Installing Cluster API CRDs")
		// Create the Cluster CRD
		cluster := &clusterv1.Cluster{
			TypeMeta: metav1.TypeMeta{
				APIVersion: clusterv1.GroupVersion.String(),
				Kind:       "Cluster",
			},
		}
		err = k8sClient.Create(ctx, cluster)
		Expect(err).NotTo(HaveOccurred())
	}

	// Create Terraform Workspace CRD
	workspaceCRD := &tfv1beta1.WorkspaceList{}
	err = k8sClient.List(ctx, workspaceCRD)
	if err != nil {
		By("Installing Terraform Workspace CRDs")
		// Create the Workspace CRD
		workspace := &tfv1beta1.Workspace{
			TypeMeta: metav1.TypeMeta{
				APIVersion: tfv1beta1.SchemeGroupVersion.String(),
				Kind:       "Workspace",
			},
		}
		err = k8sClient.Create(ctx, workspace)
		Expect(err).NotTo(HaveOccurred())
	}

	// Start the controller
	k8sManager, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme: scheme.Scheme,
	})
	Expect(err).ToNot(HaveOccurred())

	err = (&Reconciler{
		Client: k8sManager.GetClient(),
		Scheme: k8sManager.GetScheme(),
	}).SetupWithManager(k8sManager)
	Expect(err).ToNot(HaveOccurred())

	go func() {
		defer GinkgoRecover()
		err = k8sManager.Start(ctx)
		Expect(err).ToNot(HaveOccurred(), "failed to run manager")
	}()
})

var _ = AfterSuite(func() {
	cancel()
	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).NotTo(HaveOccurred())
})
