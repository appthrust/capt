# CAPTEP-0017: Helm Release Destroy Error Handling

## Summary
WorkspaceTemplateのHelm Releaseリソースの削除時に発生するエラーを適切に処理する方法について提案します。

## Motivation
現在、workspacetemplates/eks-controlplane-template-v2.yamlで定義されているhelm_releaseリソースを削除する際に、以下のエラーが発生することがあります：

```
Error: uninstallation completed with 1 error(s): uninstall: Failed to purge the release: release: not found
```

このエラーは、Helmリリースが既に存在しない状態でアンインストールを試みた際に発生します。これは実質的には無害なエラーですが、現在のワークフローを中断させる原因となっています。

### Goals
- Helm Releaseの削除時のエラーを適切に処理し、ワークフローの中断を防ぐ
- エラーが発生しても安全に処理を継続できるようにする

### Non-Goals
- Helm Release全般のエラーハンドリングの改善
- 他のリソースタイプの削除エラーへの対応

## Proposal
Terraformのlifecycleブロックを使用して、リソースの削除時のエラーを適切に処理します。

### Implementation Details
eks-controlplane-template-v2.yamlのhelm_releaseリソースにlifecycleブロックを追加します：

```hcl
resource "helm_release" "karpenter" {
  # 既存の設定
  lifecycle {
    ignore_changes = [
      namespace,  # namespaceが既に削除されている場合を考慮
      status     # リリースのステータスの変更を無視
    ]
  }
}
```

この設定により、以下の効果が期待できます：
1. namespaceが既に削除されている場合でも処理を継続
2. リリースのステータス変更による影響を最小化

### Risks and Mitigations
- リスク: 重要な変更も無視される可能性
  - 緩和策: ignore_changesの対象を必要最小限に限定
- リスク: リソースの不完全な削除
  - 緩和策: 実際のリソース状態を定期的に確認するメカニズムの導入を検討

## Design Details

### 修正内容
1. eks-controlplane-template-v2.yamlの修正
   - helm_releaseリソースにlifecycleブロックを追加
   - 既存の設定（cleanup_on_fail, atomic等）は維持

### テスト計画
1. 正常系テスト
   - 通常のHelm Releaseの削除が正常に動作することを確認
2. 異常系テスト
   - namespaceが既に削除されている状態での削除を確認
   - リリースが存在しない状態での削除が正常に完了することを確認

## Alternatives Considered

1. Helmのアンインストールオプションの使用
   - Pros: Helm側で制御可能
   - Cons: Terraform Providerで直接サポートされていない
   - 決定: 現時点では実装が困難なため不採用

2. カスタムプロバイダーの開発
   - Pros: より柔軟なエラーハンドリングが可能
   - Cons: 開発・メンテナンスコストが高い
   - 決定: オーバーエンジニアリングになるため不採用

## Implementation History

- [ ] 2024-11-13: 初期提案
- [ ] Implementation
- [ ] Testing
- [ ] Documentation

## References

- [Terraform Lifecycle Configuration](https://developer.hashicorp.com/terraform/language/meta-arguments/lifecycle)
- [Terraform Helm Provider Documentation](https://registry.terraform.io/providers/hashicorp/helm/latest/docs/resources/release)
- [Helm Release Resource](https://registry.terraform.io/providers/hashicorp/helm/latest/docs/resources/release)
