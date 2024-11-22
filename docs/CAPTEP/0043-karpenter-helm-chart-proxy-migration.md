# CAPTEP-0043: Karpenter Installation Migration to HelmChartProxy

## Summary

KarpenterのインストールをTerraformベースの実装からCluster APIのHelmChartProxyベースの実装に移行します。この変更により、Kubernetesネイティブな方法でKarpenterのライフサイクル管理を行うことが可能になります。

## Motivation

### Goals

- KarpenterのHelmチャートインストールをTerraformからHelmChartProxyに移行
- EC2NodeClassとNodePoolの設定をKubernetesマニフェストとして管理
- Karpenterの設定とNodePool設定の分離による柔軟性の向上
- より宣言的なアプローチでのKarpenter管理の実現

### Non-Goals

- Karpenterの基本的な機能や設定の変更
- AWS IAMロールやポリシーの設定方法の変更
- 既存のクラスターの移行プロセスの自動化

## Proposal

### User Stories

#### Story 1: Kubernetesネイティブな管理

運用者として、KarpenterをKubernetesネイティブな方法で管理したい。これにより、他のKubernetesリソースと同様の方法でKarpenterのライフサイクルを管理できる。

#### Story 2: 設定の分離と柔軟性

開発者として、KarpenterのコアコンポーネントとNodePoolの設定を分離して管理したい。これにより、NodePoolの設定を独立して更新できる。

### Implementation Details

#### 1. WorkspaceTemplateの分離

既存のeks-controlplane-templateを維持しつつ、新しいeks-controlplane-template-without-karpenterを作成：

```yaml
module "karpenter" {
  source  = "terraform-aws-modules/eks/aws//modules/karpenter"
  version = "~> 20.29"

  cluster_name          = module.eks.cluster_name
  enable_v1_permissions = true
  namespace             = "karpenter"
  node_iam_role_name    = "${module.eks.cluster_name}-node"
  enable_irsa           = true
  irsa_oidc_provider_arn = module.eks.oidc_provider_arn

  tags = local.tags
}
```

Terraformでは必要なIAMロールとポリシーのみを作成し、Helmインストールは行わない。

#### 2. HelmChartProxyの実装

2つの独立したHelmChartProxyを作成：

1. Karpenterコアコンポーネント
```yaml
apiVersion: addons.cluster.x-k8s.io/v1alpha1
kind: HelmChartProxy
metadata:
  name: karpenter
spec:
  namespace: karpenter
  releaseName: karpenter
  valuesTemplate: |
    settings:
      clusterName: {{ $outputs.cluster_name }}
      clusterEndpoint: {{ $outputs.cluster_endpoint }}
```

2. デフォルトNodePool設定
```yaml
apiVersion: addons.cluster.x-k8s.io/v1alpha1
kind: HelmChartProxy
metadata:
  name: karpenter-aws-default-nodepool
spec:
  namespace: karpenter
  releaseName: karpenter-nodepool
  dependsOn:
    - name: karpenter
```

### 設計上の考慮点

#### 1. 名前空間の分離

- Karpenterリソースを専用の`karpenter`名前空間に配置
- 名前空間の分離によるリソース管理の明確化

#### 2. リリース名の固定

- 予測可能な動作のためにリリース名を固定
- リソース名の競合を防止

#### 3. 依存関係の管理

- NodePoolはKarpenterコアに依存
- `dependsOn`を使用して明示的に依存関係を定義

### Risks and Mitigations

#### リスク1: 移行時の互換性

- リスク: 既存クラスターとの互換性問題
- 対策: 既存テンプレートの維持と段階的な移行

#### リスク2: リソースの競合

- リスク: 既存のHelmリリースとの競合
- 対策: 固定のリリース名と名前空間の使用

## Design Details

### Test Plan

1. 単体テスト
- HelmChartProxy設定のバリデーション
- 変数展開のテスト

2. 結合テスト
- クラスター作成からKarpenterインストールまでの一連のフロー
- NodePool設定の適用と検証

3. 移行テスト
- 既存クラスターへの影響確認
- アップグレードパスの検証

### Graduation Criteria

1. すべてのテストが成功
2. ドキュメントの完備
3. 移行手順の確立

## Implementation History

- 2024-01-25: 初期提案
- 2024-01-25: 設計レビュー
- 2024-01-25: 実装開始
- 2024-01-25: 実装完了

## Alternatives

### 代替案1: ClusterResourceSet

ClusterResourceSetを使用してKarpenterをインストールする案：
- 利点: シンプルな実装
- 欠点: 柔軟性の不足、バージョン管理の難しさ

### 代替案2: カスタムコントローラー

Karpenterのインストールを専用のカスタムコントローラーで管理する案：
- 利点: より細かい制御が可能
- 欠点: 実装の複雑化、メンテナンスコストの増加

## Infrastructure Needed

- HelmChartProxy対応のCluster API環境
- テスト用EKSクラスター
- CI/CD環境の更新

## References

- [Cluster API Documentation](https://cluster-api.sigs.k8s.io/)
- [HelmChartProxy Documentation](https://cluster-api.sigs.k8s.io/tasks/experimental-features/addons/helm-chart-proxy.html)
- [Karpenter Documentation](https://karpenter.sh/)
