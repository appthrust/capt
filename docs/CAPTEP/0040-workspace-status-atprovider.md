# CAPTEP-0040: WorkspaceStatus AtProvider Management

## Summary

CAPTControlPlaneのWorkspaceStatusが最終更新時に消失する問題に対する分析と解決策を提案します。

## Motivation

CAPTControlPlaneコントローラーにおいて、WorkspaceStatusのatProviderフィールドが最終更新時に消失する問題が発生しています。この問題により、Workspaceの状態が正しく反映されず、クラスターの状態管理に支障をきたす可能性があります。

### Goals

- WorkspaceStatusの永続性を確保する
- ステータス更新処理の信頼性を向上させる
- オブジェクト更新とステータス更新の適切な分離を実現する

### Non-Goals

- WorkspaceStatusの構造自体の変更
- 既存のAPIとの互換性の変更

## Proposal

### User Stories

#### Story 1: 管理者としてWorkspaceの状態を確認したい

管理者として、CAPTControlPlaneのステータスを確認する際に、WorkspaceのatProviderフィールドを含む完全な状態情報を確認できるようにしたい。

#### Story 2: 開発者としてWorkspaceの状態を利用したい

開発者として、CAPTControlPlaneのコントローラーロジックの中で、WorkspaceのatProviderフィールドを含む状態情報を利用して、適切な制御を行いたい。

### Implementation Details

#### 問題の分析

1. 状態管理の複雑性：
```go
type CAPTControlPlaneStatus struct {
    WorkspaceStatus *WorkspaceStatus `json:"workspaceStatus,omitempty"`
}

type WorkspaceStatus struct {
    AtProvider *runtime.RawExtension `json:"atProvider,omitempty"`
}
```

- 複数階層のポインタ型による状態管理の複雑さ
- 更新時の状態保持が不完全

2. 更新処理の問題：
```go
// 更新時に状態が消失
controlPlane.Status.WorkspaceStatus = status
if err := r.Status().Update(ctx, controlPlane); err != nil {
    return err
}
```

#### 解決策

1. パッチベースの更新：
```go
// パッチベースの更新に変更
patchBase := controlPlane.DeepCopy()
controlPlane.Status.WorkspaceStatus = status
if err := r.Status().Patch(ctx, controlPlane, client.MergeFrom(patchBase)); err != nil {
    return err
}
```

2. 状態の保持：
```go
// 現在の状態を保持
currentStatus := controlPlane.Status.WorkspaceStatus.DeepCopy()

// エラー時は状態を復元
if err := r.Status().Patch(ctx, controlPlane, client.MergeFrom(patchBase)); err != nil {
    controlPlane.Status.WorkspaceStatus = currentStatus
    return err
}
```

### 実装結果

1. WorkspaceStatusの保持：
```go
INFO    Status before final update      {"workspaceStatus": {"ready":true,"state":"Available","atProvider":...}}
```

2. atProviderフィールドの保持：
```go
"atProvider": {
    "checksum": "fff6ccdeaf6c49ea185e47e6cf18dc47",
    "outputs": {
        "cluster_endpoint": "https://...",
        ...
    }
}
```

### Risks and Mitigations

1. パフォーマンスへの影響：
   - DeepCopyによる若干のオーバーヘッド
   - 必要最小限の更新で影響を最小化

2. 競合状態：
   - パッチ操作による競合の可能性
   - リトライロジックで対応

3. 互換性：
   - 既存のAPIとの互換性は維持
   - 更新ロジックの変更のみ

## Design Details

### Test Plan

1. ユニットテスト：
   - WorkspaceStatus更新の各段階での状態保持を確認
   - エラー時の状態復元を確認

2. 統合テスト：
   - 実際のWorkspaceとの連携
   - 完全なステータス更新フローの確認

3. E2Eテスト：
   - クラスター作成フローでの状態保持確認
   - エラーケースの確認

### Graduation Criteria

1. Alpha:
   - パッチベース更新の実装
   - 基本的なテストの追加

2. Beta:
   - 統合テストの追加
   - エラーハンドリングの改善

3. GA:
   - E2Eテストの追加
   - パフォーマンス最適化

### Upgrade / Downgrade Strategy

- アップグレード：
  - コントローラーの更新のみ
  - CRDの変更なし

- ダウングレード：
  - 互換性維持により可能
  - 状態の永続性は保証

## Implementation History

- [x] 2024-11-22: 初期提案
- [x] 2024-11-22: workspace_status.goの修正
- [x] 2024-11-22: 実装の検証
- [ ] テストの追加
- [ ] ドキュメント更新

## Alternatives

### 代替案1: 状態更新の分離

WorkspaceStatusの更新を独立した関数に分離：
```go
func (r *Reconciler) updateWorkspaceStatus(ctx context.Context, controlPlane *controlplanev1beta1.CAPTControlPlane) error {
    patchBase := controlPlane.DeepCopy()
    // 更新処理
    return r.Status().Patch(ctx, controlPlane, client.MergeFrom(patchBase))
}
```

利点：
- 責務の明確な分離
- テストの容易さ

欠点：
- 処理の複雑化
- パフォーマンスへの影響

### 代替案2: イミュータブルな状態管理

新しい状態オブジェクトを作成して更新：
```go
func (r *Reconciler) createNewStatus(old *WorkspaceStatus) *WorkspaceStatus {
    new := old.DeepCopy()
    // 更新内容を設定
    return new
}
```

利点：
- 副作用の防止
- デバッグの容易さ

欠点：
- メモリ使用量の増加
- パフォーマンスへの影響
