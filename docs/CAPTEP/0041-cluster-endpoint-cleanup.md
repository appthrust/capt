# CAPTEP-0041: Cluster Endpoint Cleanup and Resource Deletion Order

## Summary

現在、Clusterを削除する際にCAPTControlPlaneとCAPTClusterが削除されますが、CAPTControlPlaneが削除されたときに、親のClusterリソースのEndpointを削除しないと、Clusterリソースが削除されない問題があります。この問題は、特に航空管制システムのような重要なシステムにおいて、リソースの適切なクリーンアップと安全な削除を確保するために解決する必要があります。

## Motivation

### Goals

- Workspaceの削除完了を確実に確認してからCAPTControlPlaneを削除する
- Clusterリソースのエンドポイントを適切にクリーンアップする
- リソース削除の順序を明確に制御する
- 削除処理中のエラーハンドリングを改善する

### Non-Goals

- 既存のWorkspaceの管理方法の変更
- クラスター作成プロセスの変更
- 他のリソースタイプの削除処理の変更

## Proposal

### User Stories

#### Story 1: クラスター削除時の安全な処理

管理者として、クラスターを削除する際に、すべての関連リソースが適切な順序で確実に削除されることを確認したい。これにより、リソースのリークや不整合を防ぎ、システムの信頼性を維持できる。

#### Story 2: エラー発生時の適切な処理

運用者として、削除処理中にエラーが発生した場合でも、システムが一貫性のある状態を維持し、適切なエラーメッセージとリトライ処理を提供することを期待する。

### Implementation Details

#### リソース削除の順序制御

1. Workspaceの削除確認
```go
// Get workspace name before deleting WorkspaceTemplateApply
workspaceName := workspaceApply.Status.WorkspaceName

// Delete WorkspaceTemplateApply
if err := r.Delete(ctx, workspaceApply); err != nil {
    logger.Error(err, "Failed to delete WorkspaceTemplateApply")
    return fmt.Errorf("failed to delete WorkspaceTemplateApply: %v", err)
}

// If we have a workspace name, check if it's fully deleted
if workspaceName != "" {
    workspace := &unstructured.Unstructured{}
    workspace.SetGroupVersionKind(schema.GroupVersionKind{
        Group:   "tf.upbound.io",
        Version: "v1beta1",
        Kind:    "Workspace",
    })

    err := r.Get(ctx, types.NamespacedName{
        Name:      workspaceName,
        Namespace: controlPlane.Namespace,
    }, workspace)

    if err == nil {
        // Workspace still exists, requeue
        logger.Info("Waiting for workspace to be deleted", 
            "name", workspaceName,
            "namespace", controlPlane.Namespace)
        return fmt.Errorf("waiting for workspace deletion")
    }
}
```

2. エンドポイントのクリーンアップ
```go
// エンドポイントを削除
cluster.Spec.ControlPlaneEndpoint = clusterv1.APIEndpoint{}
if err := r.Update(ctx, cluster); err != nil {
    return fmt.Errorf("failed to update cluster endpoint: %v", err)
}
```

#### エラーハンドリングとリトライ

1. 削除待ち時のリトライ制御
```go
if err := r.cleanupResources(ctx, controlPlane); err != nil {
    logger.Error(err, "Failed to cleanup resources")
    return ctrl.Result{RequeueAfter: defaultRequeueInterval}, err
}
```

2. 詳細なログ出力
```go
logger.Info("Waiting for workspace to be deleted",
    "name", workspaceName,
    "namespace", controlPlane.Namespace)
```

### 実装の分析と改善点

#### 問題点1: 削除順序の制御不足

**問題:**
- CAPTControlPlaneが削除される前にWorkspaceが完全に削除されていない
- リソース間の依存関係が適切に考慮されていない

**解決策:**
- Workspaceの削除状態を確認する処理を追加
- 削除が完了していない場合はリトライする仕組みを実装
- 削除の順序を明確に制御（エンドポイント → WorkspaceTemplateApply → Workspace）

#### 問題点2: エラーハンドリングの不足

**問題:**
- エラー発生時の状態管理が不十分
- リトライ処理が適切に実装されていない

**解決策:**
- エラー発生時の詳細なログ出力
- 適切なリトライ間隔の設定
- エラー状態の明確な管理

### Risks and Mitigations

#### リスク1: 削除処理の遅延

- リスク: Workspaceの削除確認による処理の遅延
- 対策: 適切なリトライ間隔の設定とタイムアウト処理の実装

#### リスク2: 不完全な削除

- リスク: エラー発生時のリソース残存
- 対策: 各ステップでの詳細なエラーチェックとロギング

#### リスク3: 並行処理の影響

- リスク: 複数のリソース削除が同時に実行される場合の整合性
- 対策: 適切なロックメカニズムとステータス管理

## Design Details

### Test Plan

1. 単体テスト
- Workspaceの削除確認処理のテスト
- エンドポイントクリーンアップのテスト
- エラーケースのテスト

2. 結合テスト
- 完全な削除シーケンスのテスト
- エラー発生時のリカバリーテスト

3. 負荷テスト
- 複数クラスター同時削除時の動作確認

### Graduation Criteria

1. すべてのテストが成功すること
2. エラーハンドリングが適切に機能すること
3. リソースリークが発生しないこと

## Implementation History

- 2024-01-24: 初期提案
- 2024-01-24: 設計レビュー
- 2024-01-24: 実装開始
- 2024-01-25: 実装完了
  - Workspaceの削除確認処理を追加
  - エラーハンドリングを改善
  - ログ出力を詳細化

## Alternatives

### 代替案1: ファイナライザーの順序制御

Workspaceのファイナライザーを使用して削除順序を制御する方法も検討しましたが、以下の理由で採用しませんでした：
- 既存のWorkspace管理との整合性
- 実装の複雑化
- エラーハンドリングの難しさ

### 代替案2: 非同期削除

削除処理を非同期で行う方法も検討しましたが、以下の理由で採用しませんでした：
- 状態管理の複雑化
- デバッグの難しさ
- エラーリカバリーの複雑化

## Infrastructure Needed

- 既存のKubernetesクラスター
- テスト環境
- CI/CD環境

## Security Considerations

1. 認証・認可
- 適切な権限チェック
- セキュアな削除処理

2. 監査
- 削除操作のログ記録
- 異常検知

3. データ保護
- 機密情報の適切な削除
- バックアップとリストア

## References

- [Cluster API Documentation](https://cluster-api.sigs.k8s.io/)
- [Kubernetes Garbage Collection](https://kubernetes.io/docs/concepts/workloads/controllers/garbage-collection/)
- [Finalizer Documentation](https://kubernetes.io/docs/concepts/overview/working-with-objects/finalizers/)
