module "eks_blueprints_addons" {
    source  = "aws-ia/eks-blueprints-addons/aws"
    version = "~> 1.16"
    cluster_name      = module.eks.cluster_name
    cluster_endpoint  = module.eks.cluster_endpoint
    cluster_version   = module.eks.cluster_version
    oidc_provider_arn = module.eks.oidc_provider_arn
    create_delay_dependencies = [for prof in module.eks.fargate_profiles : prof.fargate_profile_arn]
    eks_addons = {
      coredns = {
        configuration_values = jsonencode({
          computeType = "Fargate"
          resources = {
            limits = {
              cpu = "0.25"
              memory = "256M"
            }
            requests = {
              cpu = "0.25"
              memory = "256M"
            }
          }
        })
      }
      vpc-cni    = {}
      kube-proxy = {}
    }
    enable_karpenter = true
    karpenter = {
      helm_config = {
        cacheDir = "/tmp/.helmcache"
      }
    }
    karpenter_node = {
      iam_role_use_name_prefix = false
    }
    tags = local.tags
}
