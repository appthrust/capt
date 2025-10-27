package e2e

import (
	"fmt"
	"path/filepath"
	"runtime"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var (
	testEnv   *envtest.Environment
	k8sCfg    *rest.Config
	dynClient dynamic.Interface
)

func TestEnvtestE2E(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CAPT Envtest E2E Suite")
}

var _ = BeforeSuite(func() {
	binDir := filepath.Join("..", "..", "bin", "k8s", fmt.Sprintf("1.31.0-%s-%s", runtime.GOOS, runtime.GOARCH))
	testEnv = &envtest.Environment{
		CRDDirectoryPaths: []string{
			"/home/reoring/dev/capt/config/clusterapi/infrastructure/bases",
			"/home/reoring/dev/capt/config/clusterapi/controlplane/bases",
		},
		BinaryAssetsDirectory: binDir,
	}
	var err error
	k8sCfg, err = testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sCfg).NotTo(BeNil())

	dynClient, err = dynamic.NewForConfig(k8sCfg)
	Expect(err).NotTo(HaveOccurred())

	// Quiet controller-runtime logs in tests
	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))
})

var _ = AfterSuite(func() {
	if testEnv != nil {
		_ = testEnv.Stop()
	}
})
