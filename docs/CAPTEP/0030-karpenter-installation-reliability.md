# CAPTEP-0030: Karpenter Installation Strategy

## Summary
現在のTerraform Helm Providerを使用したKarpenterのインストール方法を改善し、より信頼性の高いアプローチを提案します。

## Motivation
現在、eks-controlplane-template-v2.yamlではTerraformのhelm_releaseリソースを使用してKarpenterをインストールしていますが、以下の問題が報告されています：

1. Terraform Helm Providerの不安定性
   - クラスター作成時のHelmインストールが信頼性に欠ける
   - インストール失敗時のリカバリーが困難
2. タイミングの問題
   - EKSクラスター作成直後のHelmインストールは、クラスターの準備が完全に整う前に実行される可能性がある

### Goals
- Karpenterインストールの信頼性を向上させる
- クラスター作成とアドオンインストールの適切な分離
- より堅牢なエラーハンドリングとリカバリーメカニズムの実現

### Non-Goals
- Karpenter以外のアドオンインストール方法の変更
- クラスター作成プロセス全体の見直し

## Proposal

### アプローチの概要

1. Cluster APIのClusterResourceSet機能を使用
   - クラスター作成とアドオンインストールの分離
   - 宣言的な管理
   - 状態の追跡が可能

2. FluxCDの活用
   - HelmリリースをFluxCDで管理
   - より信頼性の高いHelmチャートの管理
   - 詳細は[CAPTEP-0031](./0031-fluxcd-integration.md)を参照

3. 変数解決の改善
   - WorkspaceTemplateのOutputsとSecretの活用
   - 詳細は[CAPTEP-0032](./0032-variable-resolution.md)を参照

### 利点

1. 信頼性の向上
   - クラスター準備完了後のインストール
   - 自動リカバリー機能
   - 状態管理の改善

2. 保守性
   - 宣言的な設定
   - バージョン管理の容易さ
   - 標準的なアップグレードパス

3. 柔軟性
   - 必要に応じた設定のカスタマイズ
   - 複数クラスターへの一貫した適用

## Implementation Plan

1. Phase 1: 基本構造の確立
   - eks-controlplane-template-v2.yamlからhelm_releaseブロックの削除
   - 必要なOutputsの追加
   - 基本的なClusterResourceSetの作成

2. Phase 2: FluxCDの統合
   - FluxCDのインストール機能の実装
   - HelmリリースのFluxCDへの移行
   - 詳細は[CAPTEP-0031](./0031-fluxcd-integration.md)を参照

3. Phase 3: 変数解決の実装
   - 変数解決メカニズムの実装
   - テストとバリデーション
   - 詳細は[CAPTEP-0032](./0032-variable-resolution.md)を参照

## Risks and Mitigations

### リスク1: 移行の複雑さ
- リスク: 既存のクラスターへの影響
- 緩和策: 
  - 段階的な移行計画
  - 詳細なテスト計画
  - ロールバック手順の整備

### リスク2: 運用の複雑さ
- リスク: 新しいアプローチの学習コスト
- 緩和策:
  - 詳細なドキュメント作成
  - 運用手順の整備
  - トラブルシューティングガイドの提供

## References

- [Cluster API ClusterResourceSet Proposal](https://github.com/kubernetes-sigs/cluster-api/blob/main/docs/proposals/20200220-cluster-resource-set.md)
- [Karpenter Installation Guide](https://karpenter.sh/docs/getting-started/installing/)
- [Cluster API Documentation](https://cluster-api.sigs.k8s.io/)
