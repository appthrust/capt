# CAPTEP-0035: Status Condition Helper Functions

## Summary

このプロポーザルでは、CAPTコントローラーでのステータス条件の処理を改善するために、ヘルパー関数の実装と再構成について説明します。特に、`FindStatusCondition`関数の実装と、関連する定数の整理に焦点を当てています。

## Motivation

kubeconfigの生成プロセスやその他のコントローラーの処理において、WorkspaceTemplateApplyやWorkspaceのステータス条件をチェックする必要があります。現在、この機能は複数の場所で重複して実装されており、コードの保守性と一貫性を損なっています。

### Goals

- `FindStatusCondition`関数の一元化
- ステータス条件の処理を標準化
- コードの重複を排除
- 定数定義の整理

### Non-Goals

- 既存の条件タイプの変更
- ステータス条件の設定方法の変更
- 新しい条件タイプの追加

## Proposal

### Implementation Details

1. `conditions.go`の新規作成：

```go
package controlplane

import (
    xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
)

// FindStatusCondition finds the condition that has the matching condition type.
func FindStatusCondition(conditions []xpv1.Condition, conditionType xpv1.ConditionType) *xpv1.Condition {
    for i := range conditions {
        if conditions[i].Type == conditionType {
            return &conditions[i]
        }
    }
    return nil
}
```

### 変更点

1. `FindStatusCondition`関数の移動
   - `service_linked_role.go`から削除
   - 新しい`conditions.go`ファイルに移動
   - 全てのコントローラーで共通のヘルパー関数として使用

### リスクと対策

1. リスク：既存のコードへの影響
   - 対策：既存の呼び出し箇所の動作確認
   - 対策：ユニットテストの追加

2. リスク：パフォーマンスへの影響
   - 対策：条件チェックのロジックを最適化
   - 対策：必要な場合のみ条件チェックを実行

## Design Details

### テストプラン

1. ユニットテスト
   - `FindStatusCondition`関数の基本的な動作
   - 条件が見つからない場合の処理
   - 複数の条件が存在する場合の処理

2. 統合テスト
   - WorkspaceTemplateApplyの状態チェック
   - Workspaceの状態チェック
   - エラー処理の確認

### 卒業基準

1. 全てのテストが成功
2. コードの重複が排除されている
3. 既存の機能に影響がない
4. ドキュメントが更新されている

## Implementation History

- 2024-01-25: 初期プロポーザル
- 2024-01-25: 実装完了

## Alternatives Considered

1. 各コントローラーで独自の実装を維持
   - 却下：コードの重複と保守性の問題

2. メソッドとしての実装
   - 却下：単純な関数で十分

3. インターフェースとしての実装
   - 却下：現時点では過剰な抽象化
