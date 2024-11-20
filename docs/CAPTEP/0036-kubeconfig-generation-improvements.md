# CAPTEP-0036: Kubeconfig Generation Improvements

## Summary

このプロポーザルでは、kubeconfigの生成プロセスを改善するために行った変更について説明します。主な変更点は、kubeconfigの生成を専用のWorkspaceTemplateに分離し、シークレット管理を改善したことです。

## Motivation

以前の実装では、以下のような課題がありました：

1. kubeconfigの生成がEKSコントロールプレーンのWorkspaceTemplateに組み込まれており、責務が混在していた
2. YAMLフォーマットの保持が難しく、生成されるkubeconfigの形式が崩れることがあった
3. シークレットの管理が複雑で、名前の衝突が発生する可能性があった

### Goals

- kubeconfigの生成を独立したコンポーネントとして分離
- YAMLフォーマットの正確な保持
- シークレット管理の簡素化と明確化
- 再利用可能なWorkspaceTemplateの作成

### Non-Goals

- kubeconfigの形式自体の変更
- 認証メカニズムの変更
- 既存のシークレット命名規則の大幅な変更

## Proposal

### Implementation Details

1. 専用WorkspaceTemplateの作成：
```yaml
apiVersion: infrastructure.cluster.x-k8s.io/v1beta1
kind: WorkspaceTemplate
metadata:
  name: eks-kubeconfig-template
spec:
  template:
    spec:
      writeConnectionSecretToRef:
        name: "${WORKSPACE_NAME}-outputs-kubeconfig"
```

2. heredocを使用したYAMLフォーマットの保持：
```hcl
output "kubeconfig" {
  value = <<-EOT
  apiVersion: v1
  clusters:
  - cluster:
      certificate-authority-data: ${var.cluster_certificate_authority_data}
      server: ${var.cluster_endpoint}
    name: ${var.cluster_name}.${var.region}.eksctl.io
  EOT
}
```

3. 2段階のシークレット管理：
   - WorkspaceTemplateが`{cluster-name}-outputs-kubeconfig`を生成
   - コントローラーがCluster API互換の`{cluster-name}-kubeconfig`を生成

### リスクと対策

1. 移行リスク：
   - 既存のクラスターは再作成なしで新しい方式に移行可能
   - 古い形式のkubeconfigも引き続きサポート

2. パフォーマンスリスク：
   - 追加のWorkspaceTemplateApplyによる若干のオーバーヘッド
   - 必要な場合のみkubeconfig生成を実行

3. 互換性リスク：
   - Cluster APIの仕様に準拠したシークレット形式を維持
   - 既存のツールやスクリプトとの互換性を確保

## Design Details

### テストプラン

1. ユニットテスト：
   - YAMLフォーマットの検証
   - シークレット生成の検証
   - エラー処理の検証

2. 統合テスト：
   - 完全なクラスター作成フロー
   - kubeconfigを使用したクラスターアクセス
   - 異常系のテスト

### 卒業基準

1. すべてのテストが成功
2. YAMLフォーマットが正しく保持される
3. シークレットが適切に生成される
4. 既存のクラスターとの互換性が確保される

## Implementation History

- 2024-01-25: 初期プロポーザル
- 2024-01-25: 実装完了
- 2024-01-25: v0.1.11でリリース

## Alternatives Considered

1. YAMLエンコーディングの使用：
   - 却下：フォーマットの制御が難しい
   - heredocの方が可読性が高く、メンテナンスが容易

2. 単一シークレットの使用：
   - 却下：責務の分離が不明確
   - 2段階のアプローチの方が柔軟性が高い

3. コントローラーでの直接生成：
   - 却下：Terraformの機能を活用できない
   - WorkspaceTemplateの方が管理が容易

## References

- [Cluster API Kubeconfig Specification](https://cluster-api.sigs.k8s.io/developer/architecture/controllers/cluster.html#kubeconfig)
- [EKS Authentication](https://docs.aws.amazon.com/eks/latest/userguide/cluster-auth.html)
- [CAPTEP-0034: Dedicated WorkspaceTemplate for kubeconfig generation](./0034-kubeconfig-generation-workspace.md)
