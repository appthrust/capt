# CAPTEP-0023: EC2 Spotインスタンス用Service-Linked Role作成の自動化

## 概要

CAPTControlPlaneが作成したEKSでFargateからEC2 Spotインスタンスの調達が失敗する問題が発生しています。
これは、EC2 Spotインスタンス用のService-Linked Roleが存在せず、作成権限もない状態で発生しています。

## 背景

### 現状の実装

1. CAPTControlPlaneは、KarpenterをEKSクラスターにデプロイし、必要に応じてEC2インスタンスを自動でプロビジョニングします。

2. Karpenterの設定では、コスト最適化のためにSpotインスタンスを使用するように設定されています：
```yaml
apiVersion: karpenter.sh/v1
kind: NodePool
metadata:
  name: default
spec:
  template:
    spec:
      requirements:
        - key: "karpenter.k8s.aws/instance-category"
          operator: In
          values: ["c", "m", "r"]
```

### 問題点

1. EC2 Spotインスタンスを使用するには、AWSアカウントに`AWSServiceRoleForEC2Spot`というService-Linked Roleが必要です。

2. このRoleが存在しない場合、以下のエラーが発生します：
```
AuthFailure.ServiceLinkedRoleCreationNotPermitted: The provided credentials do not have permission to create the service-linked role for EC2 Spot Instances.
```

## 解決策

### 1. Service-Linked Roleの条件付き作成

eks-controlplane-template-v2.yamlのTerraformモジュールに以下を追加：

```hcl
# Check if the service-linked role already exists
data "aws_iam_role" "spot" {
  name = "AWSServiceRoleForEC2Spot"
  
  # Roleが存在しない場合はエラーを無視
  count = 0
}

# Create the service-linked role only if it doesn't exist
resource "aws_iam_service_linked_role" "spot" {
  aws_service_name = "spot.amazonaws.com"
  description      = "Service-linked role for EC2 Spot Instances"

  # Roleが存在する場合は作成をスキップ
  lifecycle {
    ignore_changes = [aws_service_name]
  }
}
```

### 2. 依存関係の設定

1. KarpenterのNodePool作成前にService-Linked Roleが作成されるように依存関係を設定：

```hcl
resource "kubectl_manifest" "node_pool" {
  # ... existing configuration ...

  depends_on = [
    kubectl_manifest.ec2_node_class,
    aws_iam_service_linked_role.spot
  ]
}
```

## 実装の注意点

1. Service-Linked Roleの重複作成について：
   - Service-Linked Roleがすでに存在する場合、Terraformは`EntityAlreadyExists`エラーを返します
   - このエラーは無害で、実行に影響を与えません
   - `lifecycle { ignore_changes = [aws_service_name] }`を使用することで、既存のRoleに対する変更を無視します

2. エラーハンドリング：
   - Roleの作成に失敗した場合でも、すでにRoleが存在する場合は処理を継続できます
   - 権限不足などの他のエラーの場合は、適切なエラーメッセージを表示します

## 実装計画

1. eks-controlplane-template-v2.yamlの修正
   - Service-Linked Role作成用のリソースを追加
   - 依存関係の設定を追加
   - エラーハンドリングの実装

2. テスト
   - 新規アカウントでの動作確認（Roleが存在しない場合）
   - 既存アカウントでの動作確認（Roleが存在する場合）
   - エラーケースのテスト

## 代替案

### 1. 手動でのService-Linked Role作成

```bash
aws iam create-service-linked-role --aws-service-name spot.amazonaws.com
```

- メリット：一度だけの作業で済む
- デメリット：自動化されない、新規アカウントでの展開時に手動作業が必要

### 2. Spotインスタンスを使用しない設定

- NodePoolの設定からSpotインスタンスの使用を除外
- メリット：Service-Linked Roleが不要
- デメリット：コスト最適化の機会を失う

## 実装履歴

- [x] 2024-11-13: 初期提案
- [x] 2024-11-13: 問題の原因特定（Service-Linked Role不足）
- [x] 2024-11-13: 既存Roleの処理方法を追加
- [ ] eks-controlplane-template-v2.yamlの修正
- [ ] テスト実施
- [ ] ドキュメント更新

## 参考資料

- [AWS Service-Linked Roles](https://docs.aws.amazon.com/IAM/latest/UserGuide/using-service-linked-roles.html)
- [EC2 Spot Instances](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/using-spot-instances.html)
- [Terraform aws_iam_service_linked_role](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/iam_service_linked_role)
- [Terraform Lifecycle Configuration](https://www.terraform.io/docs/language/meta-arguments/lifecycle.html)
