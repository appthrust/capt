# CAPTEP-0019: Control Plane Endpoint Deletion

## Summary
親のClusterを削除した際に、CaptControlPlaneが削除される時点で、親クラスタに設定されているEndpointを適切に削除する処理を追加します。

## Motivation

### 現状の問題
1. CaptControlPlaneの削除時に、親クラスタのエンドポイントが適切に削除されない
2. これにより、親クラスタのリソースに不要なエンドポイント設定が残存する
3. リソースの完全なクリーンアップが行われない

### 影響
- リソースリーク：不要なエンドポイント設定が残る
- 一貫性の欠如：クラスタ削除後もエンドポイントが残存
- 運用上の問題：手動クリーンアップが必要

## Goals
- CaptControlPlaneの削除時に、親クラスタのエンドポイントを適切に削除する
- エンドポイント削除の失敗を適切にハンドリングする
- 削除処理の順序を適切に管理する

## Non-Goals
- 既存の削除済みエンドポイントの自動クリーンアップ
- エンドポイント管理の一般的な改善
- 他のリソースタイプの削除処理の変更

## Proposal

### 実装の概要
1. cleanupResources関数にエンドポイント削除処理を追加
2. エラーハンドリングの実装
3. 削除順序の管理

### 詳細設計

#### 1. エンドポイント削除処理の追加
```go
func (r *Reconciler) cleanupResources(ctx context.Context, controlPlane *controlplanev1beta1.CAPTControlPlane) error {
    logger := log.FromContext(ctx)

    // 親クラスタを取得
    cluster := &clusterv1.Cluster{}
    if err := r.Get(ctx, types.NamespacedName{
        Name:      controlPlane.Name,
        Namespace: controlPlane.Namespace,
    }, cluster); err != nil {
        if !apierrors.IsNotFound(err) {
            return fmt.Errorf("failed to get parent cluster: %v", err)
        }
        // 親クラスタが既に削除されている場合は処理を続行
        logger.Info("Parent cluster already deleted")
    } else {
        // エンドポイントを削除
        cluster.Spec.ControlPlaneEndpoint = clusterv1.APIEndpoint{}
        if err := r.Update(ctx, cluster); err != nil {
            return fmt.Errorf("failed to update cluster endpoint: %v", err)
        }
        logger.Info("Successfully cleared control plane endpoint")
    }

    // 既存のWorkspaceTemplateApplyの削除処理
    // ... (既存のコード)

    return nil
}
```

#### 2. エラーハンドリング
- クラスタ取得エラーの適切な処理
- エンドポイント更新エラーの処理
- リソース削除の順序管理

### テスト計画

1. 正常系テスト
   - 親クラスタ存在時の削除処理
   - エンドポイント削除の確認
   - リソース削除順序の確認

2. 異常系テスト
   - 親クラスタ不在時の処理
   - エンドポイント更新エラー時の処理
   - 部分的な削除失敗時のリカバリー

## Alternatives Considered

1. CaptClusterでの実装
   - Pros: インフラストラクチャー層での一元管理
   - Cons: コントロールプレーンの責務が不適切
   - 決定: 採用しない（コントロールプレーンの責務）

2. 非同期削除処理
   - Pros: 削除処理の柔軟性向上
   - Cons: 複雑性の増加、状態管理の困難さ
   - 決定: 採用しない（シンプルな同期処理で十分）

## Implementation History

- [ ] 2024-11-13: 初期提案
- [ ] Implementation
- [ ] Testing
- [ ] Documentation

## References

- [Cluster API Control Plane Provider](https://cluster-api.sigs.k8s.io/developer/providers/implementers-guide/controllers_and_reconciliation.html)
- [Kubernetes Finalizers](https://kubernetes.io/docs/concepts/overview/working-with-objects/finalizers/)
- [Controller Runtime Client](https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/client)
