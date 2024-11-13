# CAPTEP-0018: Cluster Owner Reference Handling

## Summary
親のClusterリソースを削除しても、CAPTClusterの削除が開始されない問題の解決策を提案します。

## Motivation

### 現状の問題
1. CAPTClusterコントローラーが親Clusterを探す際に、OwnerReferenceではなく名前ベースで検索している
2. これにより、親Clusterが削除されても、CAPTClusterは異なる名前のClusterを探し続ける
3. 結果として、親Clusterの削除イベントを正しく検知できず、CAPTClusterの削除が開始されない

### 影響
- リソースリーク：親Clusterが削除されても、CAPTClusterとその配下のリソースが残存
- 運用性：手動でのクリーンアップが必要
- Cluster API仕様との不整合：インフラストラクチャープロバイダーの標準的な動作から逸脱

## Goals
- 親Clusterの削除を正しく検知し、CAPTClusterの削除を開始する
- Cluster APIの標準的なライフサイクル管理パターンに準拠する
- リソースの適切なクリーンアップを保証する

## Non-Goals
- 既存の削除済みリソースの自動クリーンアップ
- 一般的なリソース管理の最適化
- 削除順序の変更

## Proposal

### 実装の概要
1. CAPTClusterコントローラーでの親Cluster検索ロジックを修正
2. OwnerReferenceに基づいた親Clusterの検索を実装
3. 削除イベントの適切な処理を確保

### 詳細設計

#### 1. 親Cluster検索ロジックの修正
```go
func (r *Reconciler) getOwnerCluster(ctx context.Context, captCluster *infrastructurev1beta1.CAPTCluster) (*clusterv1.Cluster, error) {
    // OwnerReferencesから親Clusterの参照を取得
    for _, ref := range captCluster.OwnerReferences {
        if ref.APIVersion == clusterv1.GroupVersion.String() && ref.Kind == "Cluster" {
            // 親Clusterを取得
            cluster := &clusterv1.Cluster{}
            key := types.NamespacedName{
                Namespace: captCluster.Namespace,
                Name:     ref.Name,
            }
            if err := r.Get(ctx, key, cluster); err != nil {
                return nil, err
            }
            return cluster, nil
        }
    }
    return nil, fmt.Errorf("no owner cluster found")
}
```

#### 2. Reconcileロジックの更新
```go
// Reconcileメソッドの修正
func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (Result, error) {
    // ...

    // 親Clusterの取得（修正後）
    cluster, err := r.getOwnerCluster(ctx, captCluster)
    if err != nil {
        if apierrors.IsNotFound(err) {
            return r.handleMissingCluster(ctx, captCluster)
        }
        return Result{}, err
    }

    // ...
}
```

### テスト計画

1. 正常系テスト
   - 親Cluster削除時のCAPTCluster削除開始の確認
   - OwnerReference変更時の動作確認
   - 削除の伝播確認

2. 異常系テスト
   - 親Clusterが存在しない場合の処理
   - 不正なOwnerReferenceの処理
   - 削除失敗時のリカバリー

## Alternatives Considered

1. 現在の名前ベース検索の維持
   - Pros: 既存コードの変更が最小限
   - Cons: 根本的な問題が解決されない
   - 決定: 採用しない

2. Watch対象の拡張
   - Pros: より細かい制御が可能
   - Cons: 複雑性が増加
   - 決定: 現時点では不要

## Implementation History

- [ ] 2024-11-13: 初期提案
- [ ] Implementation
- [ ] Testing
- [ ] Documentation

## References

- [Cluster API Owner References](https://cluster-api.sigs.k8s.io/developer/providers/implementers-guide/controllers_and_reconciliation.html#owner-references)
- [Kubernetes Garbage Collection](https://kubernetes.io/docs/concepts/architecture/garbage-collection/)
- [Controller Runtime Client](https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/client)
