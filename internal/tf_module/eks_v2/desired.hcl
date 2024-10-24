module "eks" {
  source  = "terraform-aws-modules/eks/aws"
  version = "~> 20.11"
  cluster_name                   = local.name
  cluster_version                = "1.31"
  cluster_endpoint_public_access = true
  vpc_id     = module.vpc.vpc_id
  subnet_ids = module.vpc.private_subnets
  create_cluster_security_group = false
  create_node_security_group    = false
  enable_cluster_creator_admin_permissions = true
  fargate_profiles = {
    karpenter = {
      selectors = [
        { namespace = "karpenter" }
      ]
    }
    kube_system = {
      name = "kube-system"
      selectors = [
        { namespace = "kube-system" }
      ]
    }
  }
  tags = merge(local.tags, { "karpenter.sh/discovery" = local.name })
}
