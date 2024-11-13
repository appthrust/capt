# Testing Best Practices

このドキュメントでは、CAPTプロジェクトのテスト作成に関するベストプラクティスと知見をまとめています。

## コントローラーのテスト設計

### 1. テストケースの構造化

```go
type testCase struct {
    name          string           // テストケースの説明
    captCluster   *CAPTCluster     // テスト対象のリソース
    existingObjs  []runtime.Object // 事前に存在するリソース
    expectedError error            // 期待されるエラー
    validate      func(t *testing.T, c client.Client) // 検証ロジック
}
```

このような構造化により：
- テストケースの意図が明確になる
- 前提条件と期待される結果が分かりやすい
- 検証ロジックを柔軟に記述できる

### 2. フェイククライアントの設定

```go
fakeClient := fake.NewClientBuilder().
    WithScheme(scheme).
    WithStatusSubresource(&CAPTCluster{}).
    WithStatusSubresource(&WorkspaceTemplateApply{}).
    Build()
```

重要な点：
- StatusSubresourceを有効にする
- 必要なスキーマを登録する
- テストケースごとに新しいクライアントを作成する

### 3. リソースの作成順序

1. まずCAPTClusterを作成
2. その後で依存リソースを作成
3. DeepCopyを使用して元のオブジェクトを変更しない

```go
// Create CAPTCluster
err := fakeClient.Create(ctx, tc.captCluster.DeepCopy())

// Create other objects
for _, obj := range tc.existingObjs {
    err = fakeClient.Create(ctx, obj.(client.Object))
}
```

### 4. 検証パターン

以下のパターンを考慮してテストケースを作成：

1. 正常系
   - リソースが正しく作成される
   - ステータスが適切に更新される
   - フィナライザーが正しく処理される

2. 異常系
   - リソースが存在しない
   - 依存リソースが存在しない
   - 無効な設定値

3. エッジケース
   - 削除中のリソース
   - 親リソースがnil
   - 競合状態

## テストの実装パターン

### 1. リソースの状態検証

```go
validate: func(t *testing.T, c client.Client) {
    captCluster := &CAPTCluster{}
    err := c.Get(context.Background(), types.NamespacedName{
        Name:      "test-cluster",
        Namespace: "default",
    }, captCluster)
    assert.NoError(t, err)
    assert.False(t, controllerutil.ContainsFinalizer(captCluster, CAPTClusterFinalizer))
}
```

### 2. 依存リソースの検証

```go
// Verify WorkspaceTemplateApply exists
workspaceApply := &WorkspaceTemplateApply{}
err = c.Get(context.Background(), types.NamespacedName{
    Name:      "test-workspace",
    Namespace: "default",
}, workspaceApply)
assert.NoError(t, err)
```

### 3. エラー処理の検証

```go
if tc.expectedError == nil {
    assert.NoError(t, err)
    assert.Equal(t, Result{}, result)
} else {
    assert.EqualError(t, err, tc.expectedError.Error())
}
```

## テストの分割方針

1. 機能単位でテストファイルを分割
   - controller_test.go: 基本的なReconcile処理
   - finalizer_test.go: 削除処理
   - vpc_test.go: VPC関連の処理
   - status_test.go: ステータス更新処理

2. 各ファイルで以下をテスト
   - 正常系の基本フロー
   - エラーケース
   - エッジケース
   - 状態遷移

## テストメンテナンス

1. テストコードの可読性
   - 明確なテストケース名
   - コメントによる意図の説明
   - 検証ロジックの分離

2. テストの独立性
   - テストケース間で状態を共有しない
   - 各テストケースで必要なリソースを明示的に作成
   - クリーンアップを適切に行う

3. テストの拡張性
   - 新しいテストケースの追加が容易
   - 検証ロジックの再利用
   - 共通処理のヘルパー関数化

## 学んだ教訓

1. StatusSubresourceの重要性
   - ステータス更新のテストには必須
   - 明示的な有効化が必要

2. リソースの作成順序
   - 依存関係を考慮した順序が重要
   - DeepCopyによる元オブジェクトの保護

3. エラー処理の網羅
   - エラーケースの明示的なテスト
   - エラーメッセージの検証

4. 検証の粒度
   - 適切な粒度での検証
   - 必要十分な検証項目の選定
