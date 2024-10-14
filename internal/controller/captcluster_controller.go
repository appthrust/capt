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
	"sigs.k8s.io/controller-runtime/pkg/log"

	infrastructurev1beta1 "github.com/appthrust/capt/api/v1beta1"
	tfv1beta1 "github.com/upbound/provider-terraform/apis/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
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
		log.Error(err, "Unable to fetch CAPTCluster")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Check if the Terraform Workspace already exists
	workspace := &tfv1beta1.Workspace{}
	workspaceName := types.NamespacedName{
		Name:      fmt.Sprintf("%s-workspace", captCluster.Name),
		Namespace: captCluster.Namespace,
	}

	if err := r.Get(ctx, workspaceName, workspace); err != nil {
		if client.IgnoreNotFound(err) != nil {
			log.Error(err, "Unable to fetch Terraform Workspace")
			return ctrl.Result{}, err
		}

		// Workspace doesn't exist, create a new one
		workspace = &tfv1beta1.Workspace{
			ObjectMeta: metav1.ObjectMeta{
				Name:      workspaceName.Name,
				Namespace: workspaceName.Namespace,
			},
			Spec: tfv1beta1.WorkspaceSpec{
				ForProvider: tfv1beta1.WorkspaceParameters{
					Module:       r.generateTerraformCode(captCluster),
					Source:       tfv1beta1.ModuleSourceInline,
					InlineFormat: tfv1beta1.FileFormatHCL,
				},
			},
		}

		if err := r.Create(ctx, workspace); err != nil {
			log.Error(err, "Unable to create Terraform Workspace")
			return ctrl.Result{}, err
		}

		log.Info("Created Terraform Workspace", "workspace", workspaceName)
	} else {
		// Workspace exists, update it
		workspace.Spec.ForProvider.Module = r.generateTerraformCode(captCluster)
		if err := r.Update(ctx, workspace); err != nil {
			log.Error(err, "Unable to update Terraform Workspace")
			return ctrl.Result{}, err
		}

		log.Info("Updated Terraform Workspace", "workspace", workspaceName)
	}

	// Update CAPTCluster status
	captCluster.Status.WorkspaceName = workspace.Name
	if err := r.Status().Update(ctx, captCluster); err != nil {
		log.Error(err, "Unable to update CAPTCluster status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *CAPTClusterReconciler) generateTerraformCode(cluster *infrastructurev1beta1.CAPTCluster) string {
	// TODO: Implement the logic to generate Terraform code based on the CAPTCluster spec
	return fmt.Sprintf(`
module "eks" {
  source  = "terraform-aws-modules/eks/aws"
  version = "~> 19.0"

  cluster_name    = "%s"
  cluster_version = "%s"

  vpc_id     = module.vpc.vpc_id
  subnet_ids = module.vpc.private_subnets

  cluster_endpoint_public_access = %t

  # Add other EKS configurations based on CAPTCluster spec
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
`, cluster.Name, cluster.Spec.Version, cluster.Spec.PublicAccess, cluster.Name, cluster.Spec.VpcCIDR,
		cluster.Spec.PublicAccess, true, // Assuming single NAT gateway for now
		formatTags(cluster.Spec.PublicSubnetTags),
		formatTags(cluster.Spec.PrivateSubnetTags),
		cluster.Spec.VpcCIDR)
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
