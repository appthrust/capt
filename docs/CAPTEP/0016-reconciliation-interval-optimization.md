# CAPTEP-0016: リコンサイル間隔の最適化

## Summary

CAPTControlPlaneコントローラーにおいて、リコンサイルが高頻度で実行される問題の解決提案です。

## Motivation

1. 現在のCAPTControlPlaneコントローラーは、リコンサイルが非常に高頻度（100ミリ秒程度）で実行されています。
2. この問題により、以下の影響が発生しています：
   - システムリソースの過剰な消費
   - ログの大量出力
   - 不要な状態更新の発生

### Goals

- リコンサイル間隔の適切な制御
- システムリソースの効率的な利用
- ログ出力の最適化
- 状態更新の適切な管理

### Non-Goals

- リコンサイルロジックの変更
- ステータス管理ロジックの変更
- エラーハンドリングの変更

## Proposal

### User Stories

#### Story 1: システムリソースの効率的な利用

クラスター管理者として、CAPTControlPlaneコントローラーが適切な間隔でリコンサイルを実行し、システムリソースを効率的に利用することを期待します。

#### Story 2: ログの適切な管理

クラスター管理者として、ログが適切な頻度で出力され、必要な情報を容易に確認できることを期待します。

### Implementation Details

1. リコンサイル間隔の制御
```go
// 現在の実装
return ctrl.Result{RequeueAfter: requeueInterval}, nil

// 提案される改善
const (
    // 通常のリコンサイル間隔
    defaultRequeueInterval = 30 * time.Second
    // エラー時のリコンサイル間隔
    errorRequeueInterval = 10 * time.Second
    // 初期化時のリコンサイル間隔
    initializationRequeueInterval = 5 * time.Second
)
```

2. 状態に応じたリコンサイル間隔の調整
```go
func (r *Reconciler) determineRequeueInterval(controlPlane *controlplanev1beta1.CAPTControlPlane) time.Duration {
    // 初期化中は短い間隔
    if !controlPlane.Status.Initialized {
        return initializationRequeueInterval
    }

    // エラー状態の場合は中間の間隔
    if controlPlane.Status.FailureMessage != nil {
        return errorRequeueInterval
    }

    // 通常状態は長い間隔
    return defaultRequeueInterval
}
```

3. リコンサイル条件の最適化
```go
func (r *Reconciler) shouldRequeue(controlPlane *controlplanev1beta1.CAPTControlPlane) bool {
    // 状態が変更された場合のみリコンサイル
    return controlPlane.Status.LastTransitionTime.Add(defaultRequeueInterval).Before(time.Now())
}
```

### Risks and Mitigations

1. リコンサイル間隔が長すぎる場合
   - 対策: 状態に応じた適切な間隔の設定
   - 対策: 重要なイベントの即時処理

2. 状態更新の遅延
   - 対策: 重要な状態変更の即時反映
   - 対策: エラー状態での適切な間隔設定

3. リソース状態の同期ずれ
   - 対策: 適切なキャッシュ無効化
   - 対策: 重要な更新の即時反映

### Test Plan

1. ユニットテスト
   - リコンサイル間隔の計算テスト
   - 状態に応じた間隔調整テスト
   - エラー状態のハンドリングテスト

2. E2Eテスト
   - 長期実行時のリソース使用量テスト
   - 状態変更の反映タイミングテスト
   - エラー状態からの回復テスト

### Graduation Criteria

1. CPU使用率の改善
2. メモリ使用量の安定化
3. ログ出力量の適正化
4. 状態更新の適切な反映

### Upgrade Strategy

この変更は後方互換性があり、特別なアップグレード手順は必要ありません。

## Implementation History

- [ ] 2024-03-XX: CAPTEP提案
- [ ] リコンサイル間隔の最適化実装
- [ ] テストの実装
- [ ] レビュー
- [ ] マージ

## Alternatives

1. イベントベースの実装
   - メリット: リソース使用量の削減
   - デメリット: 実装の複雑化

2. 状態変更の監視
   - メリット: 即時反映が可能
   - デメリット: オーバーヘッドの増加

## Lessons Learned

1. リコンサイル間隔の重要性
   - システムリソースへの影響
   - ログ管理の重要性
   - 状態更新の適切なタイミング

2. Kubernetes Controllerの設計原則
   - レベルトリガー vs エッジトリガー
   - キャッシュの重要性
   - リソース効率の考慮

3. 監視とデバッグ
   - 適切なログレベル
   - メトリクスの重要性
   - トレーサビリティの確保
