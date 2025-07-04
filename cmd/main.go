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

package main

import (
	"crypto/md5"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/metrics/server"

	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"

	controlplanev1beta1 "github.com/appthrust/capt/api/controlplane/v1beta1"
	infrastructurev1beta1 "github.com/appthrust/capt/api/v1beta1"
	"github.com/appthrust/capt/internal/controller"
	controlplanecontroller "github.com/appthrust/capt/internal/controller/controlplane"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	tfv1beta1 "github.com/upbound/provider-terraform/apis/v1beta1"
	//+kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

const (
	controlPlaneController   = "control-plane"
	infrastructureController = "infrastructure"
)

var allControllers = []string{
	controlPlaneController,
	infrastructureController,
}

// controllerFlag is a custom flag type to handle multiple controller names.
type controllerFlag []string

func (f *controllerFlag) String() string {
	return strings.Join(*f, ",")
}

func (f *controllerFlag) Set(value string) error {
	*f = append(*f, value)
	return nil
}

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(infrastructurev1beta1.AddToScheme(scheme))
	utilruntime.Must(controlplanev1beta1.AddToScheme(scheme))
	utilruntime.Must(tfv1beta1.SchemeBuilder.AddToScheme(scheme))
	utilruntime.Must(clusterv1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	var probeAddr string
	var enabledControllers controllerFlag
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.Var(&enabledControllers, "enable-controller", "The controller to enable. Can be specified multiple times. Valid options: "+strings.Join(allControllers, ", "))
	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	leaderElectionSource := enabledControllers
	if len(leaderElectionSource) == 0 {
		leaderElectionSource = allControllers
	}
	sort.Strings(leaderElectionSource)
	leaderElectionID := fmt.Sprintf("%x.capt.ai", md5.Sum([]byte(strings.Join(leaderElectionSource, ""))))

	enabledControllerMap := make(map[string]bool)
	if len(enabledControllers) == 0 {
		for _, c := range allControllers {
			enabledControllerMap[c] = true
		}
	} else {
		for _, c := range enabledControllers {
			valid := false
			for _, validController := range allControllers {
				if c == validController {
					valid = true
					break
				}
			}
			if !valid {
				setupLog.Error(nil, "invalid controller specified", "controller", c)
				os.Exit(1)
			}
			enabledControllerMap[c] = true
		}
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme: scheme,
		Metrics: server.Options{
			BindAddress: metricsAddr,
		},
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       leaderElectionID,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if enabledControllerMap[infrastructureController] {
		setupLog.Info("setting up infrastructure controllers")
		if err = controller.SetupCAPTClusterController(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", "CAPTCluster")
			os.Exit(1)
		}

		if err = (&controller.WorkspaceTemplateReconciler{
			Client: mgr.GetClient(),
			Scheme: mgr.GetScheme(),
		}).SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", "WorkspaceTemplate")
			os.Exit(1)
		}

		if err = controller.SetupWorkspaceTemplateApply(mgr, logging.NewLogrLogger(ctrl.Log.WithName("controllers").WithName("WorkspaceTemplateApply"))); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", "WorkspaceTemplateApply")
			os.Exit(1)
		}

		if err = (&controller.CaptMachineReconciler{
			Client: mgr.GetClient(),
			Scheme: mgr.GetScheme(),
		}).SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", "CaptMachine")
			os.Exit(1)
		}

		if err = (&controller.CaptMachineSetReconciler{
			Client:   mgr.GetClient(),
			Scheme:   mgr.GetScheme(),
			Recorder: mgr.GetEventRecorderFor("captmachineset-controller"),
		}).SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", "CaptMachineSet")
			os.Exit(1)
		}

		if err = (&controller.CaptMachineDeploymentReconciler{
			Client:   mgr.GetClient(),
			Scheme:   mgr.GetScheme(),
			Recorder: mgr.GetEventRecorderFor("captmachinedeployment-controller"),
		}).SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", "CaptMachineDeployment")
			os.Exit(1)
		}

		if err = (&controller.CaptMachineTemplateReconciler{
			Client: mgr.GetClient(),
			Scheme: mgr.GetScheme(),
		}).SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", "CaptMachineTemplate")
			os.Exit(1)
		}
	}

	if enabledControllerMap[controlPlaneController] {
		setupLog.Info("setting up control-plane controllers")
		if err = (&controlplanecontroller.Reconciler{
			Client:   mgr.GetClient(),
			Scheme:   mgr.GetScheme(),
			Recorder: mgr.GetEventRecorderFor("captcontrolplane-controller"),
		}).SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", "CAPTControlPlane")
			os.Exit(1)
		}

		if err = (&controlplanecontroller.CaptControlPlaneTemplateReconciler{
			Client:   mgr.GetClient(),
			Scheme:   mgr.GetScheme(),
			Recorder: mgr.GetEventRecorderFor("captcontrolplanetemplate-controller"),
		}).SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", "CaptControlPlaneTemplate")
			os.Exit(1)
		}
	}

	//+kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
