# CAPTEP-0007: Control Plane Event Recording Enhancement

## Summary

本提案は、Control Plane Controllerにイベント記録機能を追加し、運用性とデバッグ性を向上させるための具体的な改善案を提示します。

## Motivation

現状のControl Plane Controllerには以下の課題があります：

1. 状態変更の可視性が低い
2. トラブルシューティングが困難
3. 標準的なKubernetesパターンからの逸脱

これらの課題に対応し、より運用性の高いコントローラー実装を実現する必要があります。

## Goals

- イベント記録機能の実装によるコントローラーの状態変更の可視化
- Kubernetes標準パターンへの準拠
- 運用性とデバッグ性の向上

## Non-Goals

- 既存の機能の変更
- パフォーマンスの最適化
- 外部監視システムとの統合

## Proposal

### イベント記録の実装

1. ReconcilerへのEventRecorder追加
```go
type Reconciler struct {
    client.Client
    Scheme    *runtime.Scheme
    Recorder  record.EventRecorder
}
```

2. 主要なイベントの定義
- ControlPlaneCreating: コントロールプレーン作成開始
- ControlPlaneReady: コントロールプレーン準備完了
- ControlPlaneFailed: コントロールプレーン作成失敗
- WorkspaceTemplateApplyCreated: WorkspaceTemplateApply作成
- WorkspaceTemplateApplyFailed: WorkspaceTemplateApply作成失敗

### 標準パターンへの移行

1. Result型の標準化
- カスタムResult型の削除
- controller-runtimeのctrl.Resultの使用

2. コントローラー初期化の改善
- setup.goによる初期化ロジックの集中化
- RBACの適切な設定

## Implementation Details

### Phase 1: 基盤整備

1. EventRecorderの追加
2. Result型の移行
3. RBACの更新

### Phase 2: イベント実装

1. 各状態変更でのイベント発行
2. テストケースの追加
3. ドキュメントの更新

### Phase 3: 検証と改善

1. 運用テスト
2. フィードバックの収集
3. 必要に応じた調整

## Migration Plan

1. コードの変更
   - EventRecorderの追加
   - Result型の移行
   - テストの更新

2. ドキュメントの更新
   - 設計ドキュメントの作成
   - APIドキュメントの更新
   - 運用ガイドの更新

## Risks and Mitigations

### リスク

1. パフォーマンスへの影響
   - リスク: イベント記録によるオーバーヘッド
   - 対策: 重要なイベントのみを記録

2. 互換性の問題
   - リスク: 既存の統合テストへの影響
   - 対策: 段階的な導入とテストの更新

3. 運用の複雑化
   - リスク: イベントログの増加による混乱
   - 対策: 明確なイベント分類とドキュメント化

## Alternatives Considered

1. ログレベルの調整
   - 却下理由: Kubernetesネイティブな方法でない
   - イベントの方が運用者にとって扱いやすい

2. メトリクスの拡張
   - 却下理由: イベントほど詳細な情報を提供できない
   - 状態変更の文脈が失われる

## References

1. [Kubernetes Events](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#event-v1-core)
2. [Controller Runtime Event Recording](https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/event)
3. [CAPTEP-0004: Control Plane Refactoring](0004-controlplane-refactoring.md)
4. [CAPTEP-0005: Control Plane Testing Improvements](0005-controlplane-testing-improvements.md)

## Implementation History

- 2024-11-12: 初期提案
- (Future dates to be added as implementation progresses)
