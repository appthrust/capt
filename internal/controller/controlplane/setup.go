package controlplane

import (
	"context"
	"strings"

	controlplanev1beta1 "github.com/appthrust/capt/api/controlplane/v1beta1"
	infrastructurev1beta1 "github.com/appthrust/capt/api/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// ControllerVersion indicates the running version of the CAPTControlPlane controller
const ControllerVersion = "v0.4.5"

// SetupWithManager sets up the controller with the Manager.
func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	// Log controller version at startup
	log.Log.WithName("captcontrolplane").Info("Starting CAPTControlPlane controller", "version", ControllerVersion)
	// Watch filter for outputs kubeconfig Secret (created by Crossplane, not owned by CP)
	outputsKubeconfigFilter := predicate.NewPredicateFuncs(func(obj client.Object) bool {
		sec, ok := obj.(*corev1.Secret)
		if !ok {
			return false
		}
		return strings.HasSuffix(sec.GetName(), "-outputs-kubeconfig")
	})
	// Watch filter for Workspace connection Secret used for EKS outputs
	eksConnectionFilter := predicate.NewPredicateFuncs(func(obj client.Object) bool {
		sec, ok := obj.(*corev1.Secret)
		if !ok {
			return false
		}
		return strings.HasSuffix(sec.GetName(), "-eks-connection")
	})
	// Map outputs Secret -> CAPTControlPlane reconcile requests
	mapOutputsSecretToCP := handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []reconcile.Request {
		sec, ok := obj.(*corev1.Secret)
		if !ok {
			return nil
		}
		const suffix = "-outputs-kubeconfig"
		name := sec.GetName()
		if !strings.HasSuffix(name, suffix) {
			return nil
		}
		clusterName := strings.TrimSuffix(name, suffix)
		// Find CAPTControlPlane(s) in the same namespace with matching cluster label
		cpList := &controlplanev1beta1.CAPTControlPlaneList{}
		if err := r.Client.List(ctx, cpList,
			client.InNamespace(sec.GetNamespace()),
			client.MatchingLabels(map[string]string{
				"cluster.x-k8s.io/cluster-name": clusterName,
			}),
		); err != nil {
			return nil
		}
		reqs := make([]reconcile.Request, 0, len(cpList.Items))
		for _, cp := range cpList.Items {
			reqs = append(reqs, reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      cp.Name,
					Namespace: cp.Namespace,
				},
			})
		}
		return reqs
	})
	// Map EKS connection Secret -> CAPTControlPlane reconcile requests
	mapEKSConnSecretToCP := handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []reconcile.Request {
		sec, ok := obj.(*corev1.Secret)
		if !ok {
			return nil
		}
		const suffix = "-eks-connection"
		name := sec.GetName()
		if !strings.HasSuffix(name, suffix) {
			return nil
		}
		trimmed := strings.TrimSuffix(name, suffix) // "<cluster>-<rand>"
		// Derive cluster name by removing the last hyphen and subsequent random segment if present
		clusterName := trimmed
		if idx := strings.LastIndex(trimmed, "-"); idx > 0 {
			clusterName = trimmed[:idx]
		}
		cpList := &controlplanev1beta1.CAPTControlPlaneList{}
		if err := r.Client.List(ctx, cpList,
			client.InNamespace(sec.GetNamespace()),
			client.MatchingLabels(map[string]string{
				"cluster.x-k8s.io/cluster-name": clusterName,
			}),
		); err != nil {
			return nil
		}
		reqs := make([]reconcile.Request, 0, len(cpList.Items))
		for _, cp := range cpList.Items {
			reqs = append(reqs, reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      cp.Name,
					Namespace: cp.Namespace,
				},
			})
		}
		return reqs
	})

	return ctrl.NewControllerManagedBy(mgr).
		// For: 世代変更時に再キュー（Spec変更）
		For(&controlplanev1beta1.CAPTControlPlane{}, builder.WithPredicates(predicate.GenerationChangedPredicate{})).
		// Owns: 子リソース（WTA/kubeconfig Secret）イベントで再キュー
		Owns(&infrastructurev1beta1.WorkspaceTemplateApply{}).
		Owns(&corev1.Secret{}).
		// Watches: Crossplaneが生成する outputs-kubeconfig Secret の更新で再キュー
		Watches(&corev1.Secret{}, mapOutputsSecretToCP, builder.WithPredicates(outputsKubeconfigFilter)).
		// Watches: Workspace connection Secret (-eks-connection) の更新で再キュー
		Watches(&corev1.Secret{}, mapEKSConnSecretToCP, builder.WithPredicates(eksConnectionFilter)).
		Complete(r)
}
