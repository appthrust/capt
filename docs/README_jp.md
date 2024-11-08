# Cluster API Terraform Provider

## 概要

Cluster API Terraform Providerは、KubernetesクラスターのインフラストラクチャをTerraformを使用して宣言的に管理するためのツールです。このプロバイダーは、インフラストラクチャの構築、管理、運用を効率化し、一貫性のある方法でクラスターリソースを提供します。

## 主な利点

### 1. 宣言的なインフラストラクチャ管理

WorkspaceTemplateを使用することで、インフラストラクチャをコードとして管理できます：

```yaml
apiVersion: infrastructure.cluster.x-k8s.io/v1beta1
kind: WorkspaceTemplate
metadata:
  name: vpc-template
spec:
  template:
    metadata:
      description: "Template for creating AWS VPC"
      version: "1.0.0"
    spec:
      forProvider:
        source: Inline
        module: |
          module "vpc" {
            source = "terraform-aws-modules/vpc/aws"
            # VPC configuration
          }
```

- バージョン管理とタグ付けによる明確な構成管理
- 状態追跡による設定のドリフト検出
- 標準的なTerraformモジュールの活用

### 2. 堅牢な依存関係管理

コンポーネント間の依存関係を明示的に定義し、安全に管理します：

```yaml
spec:
  waitForSecret:
    name: vpc-connection
    namespace: default
```

- VPCとEKSなどのコンポーネント間の明示的な依存関係定義
- シークレットベースの安全な設定伝播
- コンポーネントごとの独立したライフサイクル管理

### 3. セキュアな設定管理

セキュリティを重視した設定管理機能を提供します：

- Kubernetesシークレットによる機密情報の安全な管理
- OIDC認証やIAMロールの自動設定
- セキュリティグループとネットワークポリシーの一元管理
- 環境間でのセキュアな設定の移行

### 4. 高い運用性と再利用性

効率的な運用と設定の再利用を実現します：

```yaml
spec:
  template:
    metadata:
      tags:
        environment: "dev"
        provider: "aws"
    spec:
      forProvider:
        vars:
          - key: cluster_name
            value: "demo-cluster"
```

- 再利用可能なインフラストラクチャテンプレート
- 環境固有の変数とタグによるカスタマイズ
- HelmチャートやEKSアドオンの自動管理
- 既存のTerraformモジュールとの互換性

### 5. モダンなKubernetes機能との統合

最新のKubernetes機能を簡単に統合できます：

- Fargateプロファイルの自動設定
- Karpenterによる効率的なノードスケーリング
- EKSアドオンの統合管理
- カスタムリソース定義（CRD）による拡張性

## 使用例

1. VPCの作成:
```yaml
apiVersion: infrastructure.cluster.x-k8s.io/v1beta1
kind: WorkspaceTemplateApply
metadata:
  name: demo-vpc-apply
spec:
  templateRef:
    name: vpc-template
  variables:
    name: demo-cluster-vpc
```

2. EKSクラスターの作成:
```yaml
apiVersion: controlplane.cluster.x-k8s.io/v1beta1
kind: CAPTControlPlane
metadata:
  name: demo-cluster
spec:
  version: "1.31"
  workspaceTemplateRef:
    name: eks-controlplane-template
```

## ベストプラクティス

1. リソース管理
- 関連リソースは同じネームスペースで管理
- 一貫性のある命名規則の使用
- 明確な依存関係の定義

2. セキュリティ
- 機密情報はシークレットとして管理
- 最小権限の原則に従ったIAM設定
- セキュリティグループの適切な設定

3. 運用管理
- 環境ごとの設定分離
- バージョン管理の活用
- 定期的な設定のドリフトチェック
