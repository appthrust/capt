# CAPTEP-0013: Control Plane Endpoint Update

## Summary

CAPTControlPlaneのエンドポイント更新機能が正しく動作していない問題に対する改善提案です。

## Motivation

現在、CAPTControlPlaneコントローラーはWorkspaceからクラスターのエンドポイントを取得し、CAPTControlPlaneリソースのSpecに反映する処理が実装されていません。これにより、クラスターの接続情報が正しく更新されず、クラスターへのアクセスに問題が発生する可能性があります。

### Goals

- Workspaceからクラスターエンドポイントを取得する処理の実装
- CAPTControlPlaneのエンドポイント情報を更新する処理の実装
- エンドポイント更新のエラーハンドリングの実装

### Non-Goals

- エンドポイントの検証機能の実装
- エンドポイントの自動リトライ機能の実装

## Proposal

### User Stories

#### Story 1: エンドポイントの自動更新

クラスター管理者として、CAPTControlPlaneリソースを作成した際に、自動的にクラスターのエンドポイントが設定されることを期待します。これにより、クラスターへのアクセスが容易になります。

### Implementation Details

1. `handleReadyStatus`メソッドの拡張
```go
func (r *Reconciler) handleReadyStatus(
    ctx context.Context,
    controlPlane *controlplanev1beta1.CAPTControlPlane,
    cluster *clusterv1.Cluster,
    workspaceApply *infrastructurev1beta1.WorkspaceTemplateApply,
) (ctrl.Result, error) {
    // 既存の処理...

    // エンドポイントの更新
    if endpoint, err := endpoint.GetEndpointFromWorkspace(ctx, r.Client, workspaceApply.Status.WorkspaceName); err != nil {
        return ctrl.Result{}, fmt.Errorf("failed to get endpoint: %w", err)
    } else if endpoint != nil {
        controlPlane.Spec.ControlPlaneEndpoint = *endpoint
        if err := r.Update(ctx, controlPlane); err != nil {
            return ctrl.Result{}, fmt.Errorf("failed to update endpoint: %w", err)
        }
    }

    // 残りの処理...
    return ctrl.Result{}, nil
}
```

2. エラーハンドリングの追加
- エンドポイント取得に失敗した場合のリトライ処理
- エンドポイントが見つからない場合のログ出力

### Risks and Mitigations

1. エンドポイント更新の失敗
   - 対策: エラーハンドリングとログ出力の強化
   - 対策: 次回の調整で再試行できるようにする

2. 不正なエンドポイント情報
   - 対策: 基本的な検証（形式、必須フィールド）の実装

## Design Details

### Test Plan

1. ユニットテスト
   - エンドポイント取得のテスト
   - エンドポイント更新のテスト
   - エラーケースのテスト

2. E2Eテスト
   - クラスター作成からエンドポイント設定までの統合テスト

### Graduation Criteria

1. すべてのテストが成功すること
2. エンドポイント更新の成功率が100%であること
3. エラーハンドリングが適切に機能すること

### Upgrade Strategy

この変更は後方互換性があり、特別なアップグレード手順は必要ありません。

## Implementation History

- [ ] 2024-03-XX: CAPTEP提案
- [ ] 実装
- [ ] テスト
- [ ] レビュー
- [ ] マージ

## Alternatives

1. Webhookによる検証
   - メリット: エンドポイントの検証をより厳密に行える
   - デメリット: 複雑性が増す

2. 非同期更新
   - メリット: パフォーマンスの向上
   - デメリット: 実装の複雑化、状態管理の難しさ
