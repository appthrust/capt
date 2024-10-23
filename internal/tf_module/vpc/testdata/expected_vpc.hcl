module "vpc" {
  source  = "terraform-aws-modules/vpc/aws"
  version = "5.0.0"

  name = "eks-vpc"
  cidr = "10.0.0.0/16"


  enable_nat_gateway = true
  single_nat_gateway = true

  public_subnet_tags = {
    "kubernetes.io/role/elb" = "1"
  }
  
  tags = {
    Environment = "dev"
    Terraform   = "true"
  }
  azs             = local.azs
  private_subnets = [for k, v in local.azs : cidrsubnet(var.vpc_cidr, 4, k)]
  public_subnets  = [for k, v in local.azs : cidrsubnet(var.vpc_cidr, 8, k + 48)]
  private_subnet_tags = merge(
    {
      "kubernetes.io/role/internal-elb" = "1"
      "karpenter.sh/discovery" = local.name
    },
    var.private_subnet_tags
  )
}
