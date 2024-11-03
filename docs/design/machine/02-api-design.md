# Machine API設計

## CAPTMachineTemplate

```yaml
apiVersion: infrastructure.cluster.x-k8s.io/v1beta1
kind: CAPTMachineTemplate
metadata:
  name: eks-managed-node-template
spec:
  template:
    spec:
      workspaceTemplateRef:
        name: eks-node-template
        namespace: default
      nodeType: ManagedNodeGroup  # または Fargate
      instanceType: t3.medium     # nodeType: ManagedNodeGroup の場合のみ
      scaling:
        minSize: 1
        maxSize: 5
        desiredSize: 3
      labels:
        role: worker
      taints: []
      additionalTags:
        Environment: "dev"
```

## WorkspaceTemplate for Machines

```yaml
apiVersion: infrastructure.cluster.x-k8s.io/v1beta1
kind: WorkspaceTemplate
metadata:
  name: eks-node-template
spec:
  template:
    spec:
      providerConfigRef:
        name: aws-provider
      forProvider:
        source: Inline
        module: |
          module "eks_managed_node_group" {
            source = "terraform-aws-modules/eks/aws//modules/eks-managed-node-group"
            
            name            = var.node_group_name
            cluster_name    = var.cluster_name
            cluster_version = var.cluster_version
            
            subnet_ids = var.subnet_ids
            
            min_size     = var.min_size
            max_size     = var.max_size
            desired_size = var.desired_size
            
            instance_types = [var.instance_type]
            
            labels = var.labels
            taints = var.taints
            
            tags = var.tags
          }
```

## CAPTMachineDeployment

```yaml
apiVersion: cluster.x-k8s.io/v1beta1
kind: CAPTMachineDeployment
metadata:
  name: worker-deployment
spec:
  replicas: 3
  selector:
    matchLabels:
      role: worker
  template:
    spec:
      infrastructureRef:
        name: eks-managed-node-template
        kind: CAPTMachineTemplate
      version: "1.31"
```

## 移行戦略

1. 既存のControlPlaneからコンピュートリソース設定を抽出
2. 新しいMachine関連のCRDを導入
3. 既存のWorkspaceTemplateを分割
   - ControlPlane用とMachine用に分離
   - Terraform moduleの依存関係を整理
4. 段階的な移行を可能にするための互換性レイヤーの提供
