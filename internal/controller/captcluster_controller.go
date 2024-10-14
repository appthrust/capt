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

package controller

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	infrastructurev1beta1 "github.com/appthrust/capt/api/v1beta1"
	tfv1beta1 "github.com/upbound/provider-terraform/apis/v1beta1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

const (
	captClusterFinalizer = "infrastructure.cluster.x-k8s.io/finalizer"
)

// CAPTClusterReconciler reconciles a CAPTCluster object
type CAPTClusterReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=captclusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=captclusters/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=captclusters/finalizers,verbs=update
// +kubebuilder:rbac:groups=tf.upbound.io,resources=workspaces,verbs=get;list;watch;create;update;patch;delete

func (r *CAPTClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// Fetch the CAPTCluster instance
	captCluster := &infrastructurev1beta1.CAPTCluster{}
	if err := r.Get(ctx, req.NamespacedName, captCluster); err != nil {
		if apierrors.IsNotFound(err) {
			log.Info("CAPTCluster resource not found. Ignoring since object must be deleted.")
			return ctrl.Result{}, nil
		}
		log.Error(err, "Failed to get CAPTCluster")
		return ctrl.Result{}, err
	}

	// Check if the CAPTCluster instance is marked to be deleted
	if captCluster.ObjectMeta.DeletionTimestamp.IsZero() {
		// The object is not being deleted, so if it does not have our finalizer,
		// then lets add the finalizer and update the object.
		if !controllerutil.ContainsFinalizer(captCluster, captClusterFinalizer) {
			controllerutil.AddFinalizer(captCluster, captClusterFinalizer)
			if err := r.Update(ctx, captCluster); err != nil {
				return ctrl.Result{}, err
			}
		}
	} else {
		// The object is being deleted
		if controllerutil.ContainsFinalizer(captCluster, captClusterFinalizer) {
			// our finalizer is present, so lets handle any external dependency
			if err := r.deleteExternalResources(ctx, captCluster); err != nil {
				// if fail to delete the external dependency here, return with error
				// so that it can be retried
				return ctrl.Result{}, err
			}

			// remove our finalizer from the list and update it.
			controllerutil.RemoveFinalizer(captCluster, captClusterFinalizer)
			if err := r.Update(ctx, captCluster); err != nil {
				return ctrl.Result{}, err
			}
		}

		// Stop reconciliation as the item is being deleted
		return ctrl.Result{}, nil
	}

	// Check if the Terraform Workspace already exists
	workspace := &tfv1beta1.Workspace{}
	workspaceName := types.NamespacedName{
		Name:      fmt.Sprintf("%s-workspace", captCluster.Name),
		Namespace: captCluster.Namespace,
	}

	if err := r.Get(ctx, workspaceName, workspace); err != nil {
		if apierrors.IsNotFound(err) {
			log.Info("Terraform Workspace not found. Creating a new one.")
			if err := r.createWorkspace(ctx, captCluster, workspaceName); err != nil {
				log.Error(err, "Failed to create Terraform Workspace")
				return ctrl.Result{}, err
			}
			log.Info("Created Terraform Workspace", "workspace", workspaceName)
		} else {
			log.Error(err, "Failed to get Terraform Workspace")
			return ctrl.Result{}, err
		}
	} else {
		log.Info("Updating existing Terraform Workspace", "workspace", workspaceName)
		if err := r.updateWorkspace(ctx, captCluster, workspace); err != nil {
			log.Error(err, "Failed to update Terraform Workspace")
			return ctrl.Result{}, err
		}
		log.Info("Updated Terraform Workspace", "workspace", workspaceName)
	}

	// Update CAPTCluster status
	captCluster.Status.WorkspaceName = workspaceName.Name
	if err := r.Status().Update(ctx, captCluster); err != nil {
		log.Error(err, "Failed to update CAPTCluster status")
		return ctrl.Result{}, err
	}

	log.Info("Successfully reconciled CAPTCluster", "name", captCluster.Name)
	return ctrl.Result{}, nil
}

func (r *CAPTClusterReconciler) deleteExternalResources(ctx context.Context, captCluster *infrastructurev1beta1.CAPTCluster) error {
	log := log.FromContext(ctx)

	// Delete the associated Terraform Workspace
	workspace := &tfv1beta1.Workspace{}
	workspaceName := types.NamespacedName{
		Name:      fmt.Sprintf("%s-workspace", captCluster.Name),
		Namespace: captCluster.Namespace,
	}

	if err := r.Get(ctx, workspaceName, workspace); err != nil {
		if apierrors.IsNotFound(err) {
			log.Info("Terraform Workspace not found. Skipping deletion.")
			return nil
		}
		return err
	}

	if err := r.Delete(ctx, workspace); err != nil {
		log.Error(err, "Failed to delete Terraform Workspace")
		return err
	}

	log.Info("Deleted Terraform Workspace", "workspace", workspaceName)
	return nil
}

func (r *CAPTClusterReconciler) createWorkspace(ctx context.Context, cluster *infrastructurev1beta1.CAPTCluster, name types.NamespacedName) error {
	workspace := &tfv1beta1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name.Name,
			Namespace: name.Namespace,
		},
		Spec: tfv1beta1.WorkspaceSpec{
			ForProvider: tfv1beta1.WorkspaceParameters{
				Module:       r.generateTerraformCode(cluster),
				Source:       tfv1beta1.ModuleSourceInline,
				InlineFormat: tfv1beta1.FileFormatHCL,
			},
		},
	}

	return r.Create(ctx, workspace)
}

func (r *CAPTClusterReconciler) updateWorkspace(ctx context.Context, cluster *infrastructurev1beta1.CAPTCluster, workspace *tfv1beta1.Workspace) error {
	workspace.Spec.ForProvider.Module = r.generateTerraformCode(cluster)
	return r.Update(ctx, workspace)
}

func (r *CAPTClusterReconciler) generateTerraformCode(cluster *infrastructurev1beta1.CAPTCluster) string {
	return fmt.Sprintf(`
module "eks" {
  source  = "terraform-aws-modules/eks/aws"
  version = "~> 19.0"

  cluster_name    = "%s"
  cluster_version = "%s"

  vpc_id     = module.vpc.vpc_id
  subnet_ids = module.vpc.private_subnets

  cluster_endpoint_public_access  = %t
  cluster_endpoint_private_access = %t

  # Add other EKS configurations based on CAPTCluster spec
  %s
}

module "vpc" {
  source  = "terraform-aws-modules/vpc/aws"
  version = "~> 5.0"

  name = "%s-vpc"
  cidr = "%s"

  azs             = ["${data.aws_availability_zones.available.names[0]}", "${data.aws_availability_zones.available.names[1]}", "${data.aws_availability_zones.available.names[2]}"]
  private_subnets = [for k, v in data.aws_availability_zones.available.names : cidrsubnet(var.vpc_cidr, 4, k)]
  public_subnets  = [for k, v in data.aws_availability_zones.available.names : cidrsubnet(var.vpc_cidr, 8, k+48)]

  enable_nat_gateway   = %t
  single_nat_gateway   = %t
  enable_dns_hostnames = true

  public_subnet_tags = %s
  private_subnet_tags = %s
}

data "aws_availability_zones" "available" {}

variable "vpc_cidr" {
  default = "%s"
}
`, cluster.Name, cluster.Spec.EKS.Version,
		cluster.Spec.EKS.PublicAccess,
		cluster.Spec.EKS.PrivateAccess,
		r.generateNodeGroupsConfig(cluster.Spec.EKS.NodeGroups),
		cluster.Name, cluster.Spec.VPC.CIDR,
		cluster.Spec.VPC.EnableNatGateway,
		cluster.Spec.VPC.SingleNatGateway,
		formatTags(cluster.Spec.VPC.PublicSubnetTags),
		formatTags(cluster.Spec.VPC.PrivateSubnetTags),
		cluster.Spec.VPC.CIDR)
}

func (r *CAPTClusterReconciler) generateNodeGroupsConfig(nodeGroups []infrastructurev1beta1.NodeGroupConfig) string {
	if len(nodeGroups) == 0 {
		return ""
	}

	result := "eks_managed_node_groups = {\n"
	for _, ng := range nodeGroups {
		result += fmt.Sprintf(`    %s = {
      instance_types = ["%s"]
      min_size     = %d
      max_size     = %d
      desired_size = %d
    }
`, ng.Name, ng.InstanceType, ng.MinSize, ng.MaxSize, ng.DesiredSize)
	}
	result += "  }"
	return result
}

func formatTags(tags map[string]string) string {
	if len(tags) == 0 {
		return "{}"
	}

	result := "{\n"
	for k, v := range tags {
		result += fmt.Sprintf("    %s = \"%s\"\n", k, v)
	}
	result += "  }"

	return result
}

// SetupWithManager sets up the controller with the Manager.
func (r *CAPTClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&infrastructurev1beta1.CAPTCluster{}).
		Complete(r)
}
