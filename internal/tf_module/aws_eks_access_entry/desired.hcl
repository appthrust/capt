resource "aws_eks_access_entry" "karpenter_node_access_entry" {
  cluster_name      = module.eks.cluster_name
  principal_arn     = module.eks_blueprints_addons.karpenter.node_iam_role_arn
  kubernetes_groups = []
  type              = "EC2_LINUX"
  lifecycle {
    ignore_changes = [kubernetes_groups]
  }
}
