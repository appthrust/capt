# CAPTEP-0014: エンドポイント更新のエラーハンドリング改善

## Summary

CAPTControlPlaneのエンドポイント更新機能が正しく動作していない問題と、CAPTClusterのステータス管理の問題に対する改善提案です。

## Motivation

1. CAPTControlPlaneコントローラーはWorkspaceからクラスターのエンドポイントを取得し、CAPTControlPlaneリソースのSpecに反映していますが、親のClusterリソースのSpec.ControlPlane.Endpointが更新されていません。これはClusterAPI仕様に準拠していない状態です。
2. CAPTClusterコントローラーのステータス管理において、エラー状態とReady状態の整合性が取れていない問題があります。

### Goals

- Workspaceからクラスターエンドポイントを取得する処理の実装
- CAPTControlPlaneのエンドポイント情報を更新する処理の実装
- 親のClusterリソースのSpec.ControlPlane.Endpointを更新する処理の実装
- エンドポイント更新のエラーハンドリングの実装
- CAPTClusterのステータス管理の改善
- ClusterAPI規約に準拠したステータス更新の実装

### Non-Goals

- エンドポイントの検証機能の実装
- エンドポイントの自動リトライ機能の実装
- ClusterAPIのphase管理の変更

## Proposal

### User Stories

#### Story 1: エンドポイントの自動更新

クラスター管理者として、CAPTControlPlaneリソースを作成した際に、自動的にクラスターのエンドポイントが設定されることを期待します。これにより、クラスターへのアクセスが容易になります。

#### Story 2: 適切なエラー状態管理

クラスター管理者として、エラーが発生した場合に適切なエラー状態が設定され、問題が解決した際にはエラー状態がクリアされることを期待します。

### Implementation Details

1. CAPTControlPlaneのエンドポイント更新
```go
func (r *Reconciler) handleReadyStatus(
    ctx context.Context,
    controlPlane *controlplanev1beta1.CAPTControlPlane,
    cluster *clusterv1.Cluster,
    workspaceApply *infrastructurev1beta1.WorkspaceTemplateApply,
) (ctrl.Result, error) {
    // エンドポイントの更新
    if workspaceApply.Status.WorkspaceName != "" {
        if apiEndpoint, err := endpoint.GetEndpointFromWorkspace(ctx, r.Client, workspaceApply.Status.WorkspaceName); err != nil {
            errMsg := fmt.Sprintf("Failed to get endpoint from workspace: %v", err)
            return r.setFailedStatus(ctx, controlPlane, cluster, ReasonEndpointUpdateFailed, errMsg)
        } else if apiEndpoint != nil {
            logger.Info("Updating control plane endpoint", "endpoint", apiEndpoint)
            
            // Update CAPTControlPlane endpoint
            controlPlane.Spec.ControlPlaneEndpoint = *apiEndpoint
            if err := r.Update(ctx, controlPlane); err != nil {
                errMsg := fmt.Sprintf("Failed to update control plane endpoint: %v", err)
                return r.setFailedStatus(ctx, controlPlane, cluster, ReasonEndpointUpdateFailed, errMsg)
            }

            // Update parent Cluster endpoint
            if cluster != nil {
                patchBase := cluster.DeepCopy()
                cluster.Spec.ControlPlaneEndpoint = *apiEndpoint
                if err := r.Patch(ctx, cluster, client.MergeFrom(patchBase)); err != nil {
                    errMsg := fmt.Sprintf("Failed to update cluster endpoint: %v", err)
                    return r.setFailedStatus(ctx, controlPlane, cluster, ReasonEndpointUpdateFailed, errMsg)
                }
            }
        }
    }
}
```

2. CAPTClusterのステータス管理改善
```go
func (r *Reconciler) updateStatus(ctx context.Context, captCluster *infrastructurev1beta1.CAPTCluster, cluster *v1beta1.Cluster) error {
    // Ready状態の場合はエラー状態をクリア
    if captCluster.Status.Ready {
        cluster.Status.FailureReason = nil
        cluster.Status.FailureMessage = nil
    } else if captCluster.Status.FailureReason != nil {
        // エラー状態の更新
        reason := capierrors.ClusterStatusError(*captCluster.Status.FailureReason)
        cluster.Status.FailureReason = &reason
        cluster.Status.FailureMessage = captCluster.Status.FailureMessage
    }
}
```

### Risks and Mitigations

1. エンドポイント更新の失敗
   - 対策: エラーハンドリングとログ出力の強化
   - 対策: 次回の調整で再試行できるようにする

2. 不正なエンドポイント情報
   - 対策: 基本的な検証（形式、必須フィールド）の実装

3. ステータス更新の競合
   - 対策: 適切なリトライ処理の実装
   - 対策: 最新バージョンの取得と更新

4. 親のClusterリソース更新の失敗
   - 対策: パッチ操作による競合の最小化
   - 対策: エラー時の適切なロールバック処理

### ClusterAPI規約への準拠

1. インフラストラクチャプロバイダーの責務
- InfrastructureReady状態の管理
- 条件（Conditions）の設定
- エラー状態（FailureReason, FailureMessage）の管理

2. ControlPlaneプロバイダーの責務
- 親のClusterリソースのSpec.ControlPlane.Endpointの更新
- ControlPlaneReady状態の管理
- エンドポイント情報の管理

3. Cluster API側の責務
- Phase（Pending, Provisioning, Provisioned, Failed等）の管理
- 全体的なクラスターのライフサイクル管理

## Design Details

### Test Plan

1. ユニットテスト
   - エンドポイント取得のテスト
   - エンドポイント更新のテスト（CAPTControlPlaneとCluster両方）
   - エラーケースのテスト
   - ステータス更新のテスト

2. E2Eテスト
   - クラスター作成からエンドポイント設定までの統合テスト
   - エラー状態からの回復テスト
   - 親のClusterリソース更新の検証テスト

### Graduation Criteria

1. すべてのテストが成功すること
2. エンドポイント更新の成功率が100%であること
3. エラーハンドリングが適切に機能すること
4. ClusterAPI規約に準拠していること

### Upgrade Strategy

この変更は後方互換性があり、特別なアップグレード手順は必要ありません。

## Implementation History

- [x] 2024-03-XX: CAPTEP提案
- [x] エンドポイント更新の実装
- [x] ステータス管理の改善
- [x] ClusterAPI規約への準拠
- [x] テスト
- [x] レビュー
- [x] マージ

## Alternatives

1. Webhookによる検証
   - メリット: エンドポイントの検証をより厳密に行える
   - デメリット: 複雑性が増す

2. 非同期更新
   - メリット: パフォーマンスの向上
   - デメリット: 実装の複雑化、状態管理の難しさ

## Lessons Learned

1. ClusterAPI規約の重要性
- インフラストラクチャプロバイダーとCluster APIの責務を明確に分離することの重要性
- 適切なステータス管理による一貫性の確保
- ControlPlaneプロバイダーの責務の理解と実装

2. エラー状態管理の原則
- Ready状態とエラー状態の整合性確保
- 適切なエラー状態のクリア
- 明確なエラーメッセージの提供

3. 状態管理のベストプラクティス
- 最新バージョンの取得と更新
- 競合状態の適切な処理
- ログ出力による追跡可能性の確保

4. エンドポイント更新の教訓
- 親リソースの更新の重要性
- パッチ操作による競合の最小化
- エラーハンドリングの階層的な実装
