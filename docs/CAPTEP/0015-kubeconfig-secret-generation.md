# CAPTEP-0015: kubeconfigシークレット生成の実装

## Summary

CAPTControlPlaneコントローラーにおいて、{controlplane-name}-control-plane-kubeconfigシークレットが生成されない問題の解決提案です。

## Motivation

1. CAPTControlPlaneコントローラーには、kubeconfigを生成するためのreconcileSecrets関数が実装されていますが、この関数がReconcileループ内で呼び出されていません。
2. この問題により、クラスターへのアクセスに必要なkubeconfigが生成されず、クラスターにアクセスできない状態となっています。

### Goals

- Reconcileループ内でのreconcileSecrets関数の呼び出し実装
- エラーハンドリングの実装
- 適切なログ出力の追加
- 既存のテストケースの拡張

### Non-Goals

- reconcileSecrets関数自体の変更
- Secret生成ロジックの変更
- 新しいSecret形式の追加

## Proposal

### User Stories

#### Story 1: kubeconfigの自動生成

クラスター管理者として、CAPTControlPlaneリソースを作成した際に、自動的にkubeconfigがSecretとして生成されることを期待します。

### Implementation Details

1. controller.goの修正
```go
func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    // ... existing code ...

    // Update status based on WorkspaceTemplateApply conditions
    result, err := r.updateStatus(ctx, controlPlane, workspaceApply, cluster)
    if err != nil {
        return result, err
    }

    // Add secret reconciliation after status update
    if err := r.reconcileSecrets(ctx, controlPlane, cluster, workspaceApply); err != nil {
        logger.Error(err, "Failed to reconcile secrets")
        return ctrl.Result{}, err
    }

    // Fetch the final updated object
    if err := r.Get(ctx, req.NamespacedName, controlPlane); err != nil {
        return ctrl.Result{}, err
    }

    return result, nil
}
```

### Risks and Mitigations

1. Secret生成の失敗
   - 対策: エラーハンドリングとログ出力の強化
   - 対策: 次回の調整で再試行できるようにする

2. 既存のSecretとの競合
   - 対策: 適切なオーナー参照の設定
   - 対策: 更新処理の適切な実装

### Test Plan

1. ユニットテストの拡張
   - reconcileSecretsの呼び出しテスト
   - エラーケースのテスト
   - Secret生成の成功確認テスト

2. E2Eテスト
   - クラスター作成からSecret生成までの統合テスト
   - エラー状態からの回復テスト

### Graduation Criteria

1. すべてのテストが成功すること
2. Secret生成の成功率が100%であること
3. エラーハンドリングが適切に機能すること

### Upgrade Strategy

この変更は後方互換性があり、特別なアップグレード手順は必要ありません。既存のクラスターは、次回の調整時にSecretが生成されます。

## Implementation History

- [x] 2024-03-XX: CAPTEP提案
- [ ] controller.goの修正実装
- [ ] テストの拡張実装
- [ ] レビュー
- [ ] マージ

## Alternatives

1. 別のReconcileフェーズでの実装
   - メリット: ステータス更新との分離
   - デメリット: 複雑性の増加

2. 非同期Secret生成
   - メリット: パフォーマンスの向上
   - デメリット: 状態管理の複雑化

## Lessons Learned

1. Reconcileループの重要性
   - 各フェーズの適切な順序付け
   - 必要な処理の漏れ防止

2. エラー状態管理の原則
   - 適切なエラーハンドリング
   - 明確なエラーメッセージの提供
   - リトライ戦略の重要性

3. テストの重要性
   - 機能の完全性の確認
   - エラーケースの網羅
   - 統合テストの必要性
