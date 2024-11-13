# CAPT Testing Best Practices

このドキュメントでは、CAPTプロジェクトの単体テストに特化したベストプラクティスと知見をまとめています。

## ClusterAPIリソースのテスト

### CAPTClusterのテスト

#### 1. 親Clusterとの関係

```go
// 親Clusterが存在しない場合
cluster := &clusterv1.Cluster{}
err := r.Get(ctx, types.NamespacedName{
    Namespace: captCluster.Namespace,
    Name:      captCluster.Name,
}, cluster)
assert.True(t, apierrors.IsNotFound(err))

// WaitingForCluster条件の検証
var waitingCondition *metav1.Condition
for i := range captCluster.Status.Conditions {
    if captCluster.Status.Conditions[i].Type == WaitingForClusterCondition {
        waitingCondition = &captCluster.Status.Conditions[i]
        break
    }
}
assert.NotNil(t, waitingCondition)
assert.Equal(t, metav1.ConditionTrue, waitingCondition.Status)
```

#### 2. VPC設定の検証

```go
// 既存VPCの使用
if captCluster.Spec.ExistingVPCID != "" {
    assert.Equal(t, captCluster.Spec.ExistingVPCID, captCluster.Status.VPCID)
    assert.True(t, captCluster.Status.Ready)
}

// VPCテンプレートの使用
if captCluster.Spec.VPCTemplateRef != nil {
    assert.NotEmpty(t, captCluster.Spec.WorkspaceTemplateApplyName)
    // VPC作成中の状態を検証
    var vpcReadyCondition *metav1.Condition
    for i := range captCluster.Status.Conditions {
        if captCluster.Status.Conditions[i].Type == infrastructurev1beta1.VPCReadyCondition {
            vpcReadyCondition = &captCluster.Status.Conditions[i]
            break
        }
    }
    assert.NotNil(t, vpcReadyCondition)
    assert.Equal(t, metav1.ConditionFalse, vpcReadyCondition.Status)
}
```

### WorkspaceTemplateApplyの処理

#### 1. 作成と削除

```go
// WorkspaceTemplateApplyの存在確認
workspaceApply := &infrastructurev1beta1.WorkspaceTemplateApply{}
err := r.Get(ctx, types.NamespacedName{
    Name:      captCluster.Spec.WorkspaceTemplateApplyName,
    Namespace: captCluster.Namespace,
}, workspaceApply)

// 削除時の処理
if !captCluster.Spec.RetainVPCOnDelete {
    assert.Error(t, err)
    assert.True(t, apierrors.IsNotFound(err))
} else {
    assert.NoError(t, err)
}
```

#### 2. 状態の検証

```go
// WorkspaceTemplateApplyの状態検証
assert.Equal(t, workspaceApply.Spec.TemplateName, "vpc-template")
assert.Equal(t, workspaceApply.Spec.TemplateNamespace, captCluster.Namespace)
```

### フィナライザーの処理

#### 1. 追加と削除

```go
// フィナライザーの追加
if !controllerutil.ContainsFinalizer(captCluster, CAPTClusterFinalizer) {
    controllerutil.AddFinalizer(captCluster, CAPTClusterFinalizer)
    err := r.Update(ctx, captCluster)
    assert.NoError(t, err)
}

// フィナライザーの削除
controllerutil.RemoveFinalizer(captCluster, CAPTClusterFinalizer)
err := r.Update(ctx, captCluster)
assert.NoError(t, err)
```

#### 2. 削除処理の検証

```go
// 削除マークされたリソースの処理
if !captCluster.DeletionTimestamp.IsZero() {
    // VPC保持設定の検証
    if captCluster.Spec.RetainVPCOnDelete {
        assert.NotEmpty(t, captCluster.Status.VPCID)
        // WorkspaceTemplateApplyが残っていることを確認
        workspaceApply := &infrastructurev1beta1.WorkspaceTemplateApply{}
        err := r.Get(ctx, types.NamespacedName{
            Name:      captCluster.Spec.WorkspaceTemplateApplyName,
            Namespace: captCluster.Namespace,
        }, workspaceApply)
        assert.NoError(t, err)
    }
}
```

### ステータス管理

#### 1. 条件の設定と検証

```go
// 条件の検証ヘルパー関数
func assertCondition(t *testing.T, conditions []metav1.Condition, conditionType string, status metav1.ConditionStatus, reason string) {
    var condition *metav1.Condition
    for i := range conditions {
        if conditions[i].Type == conditionType {
            condition = &conditions[i]
            break
        }
    }
    assert.NotNil(t, condition)
    assert.Equal(t, status, condition.Status)
    assert.Equal(t, reason, condition.Reason)
}

// 使用例
assertCondition(t, captCluster.Status.Conditions,
    infrastructurev1beta1.VPCReadyCondition,
    metav1.ConditionTrue,
    infrastructurev1beta1.ReasonExistingVPCUsed)
```

#### 2. Ready状態の管理

```go
// Ready状態の検証
if captCluster.Status.Ready {
    // 必要な条件がすべて満たされていることを確認
    assertCondition(t, captCluster.Status.Conditions,
        infrastructurev1beta1.VPCReadyCondition,
        metav1.ConditionTrue,
        "")
    assert.NotEmpty(t, captCluster.Status.VPCID)
} else {
    // エラー状態の検証
    var failureCondition *metav1.Condition
    for i := range captCluster.Status.Conditions {
        if captCluster.Status.Conditions[i].Status == metav1.ConditionFalse {
            failureCondition = &captCluster.Status.Conditions[i]
            break
        }
    }
    assert.NotNil(t, failureCondition)
}
```

### エラー処理とリトライ

#### 1. リトライ処理の検証

```go
// リトライが必要な場合の結果検証
result, err := reconciler.Reconcile(ctx, req)
assert.NoError(t, err)
assert.Equal(t, requeueInterval, result.RequeueAfter)

// 即時リトライが必要な場合
assert.True(t, result.Requeue)
assert.Equal(t, time.Duration(0), result.RequeueAfter)
```

#### 2. エラー状態の検証

```go
// エラー状態の設定と検証
result, err := reconciler.setFailedStatus(ctx, captCluster, cluster, "TestFailure", "Test error message")
assert.NoError(t, err)
assert.False(t, result.Requeue)

// エラー状態の条件を検証
var failureCondition *metav1.Condition
for i := range captCluster.Status.Conditions {
    if captCluster.Status.Conditions[i].Type == "Failed" {
        failureCondition = &captCluster.Status.Conditions[i]
        break
    }
}
assert.NotNil(t, failureCondition)
assert.Equal(t, "TestFailure", failureCondition.Reason)
assert.Equal(t, "Test error message", failureCondition.Message)
```

## テストケースの構造化

### 1. テストケースの定義

```go
type testCase struct {
    name          string           // テストケースの説明
    captCluster   *CAPTCluster     // テスト対象のリソース
    existingObjs  []runtime.Object // 事前に存在するリソース
    expectedError error            // 期待されるエラー
    validate      func(t *testing.T, c client.Client) // 検証ロジック
}
```

### 2. フェイククライアントの設定

```go
fakeClient := fake.NewClientBuilder().
    WithScheme(scheme).
    WithStatusSubresource(&CAPTCluster{}).
    WithStatusSubresource(&WorkspaceTemplateApply{}).
    Build()
```

### 3. リソースの作成順序

```go
// Create CAPTCluster
err := fakeClient.Create(ctx, tc.captCluster.DeepCopy())

// Create other objects
for _, obj := range tc.existingObjs {
    err = fakeClient.Create(ctx, obj.(client.Object))
}
```

## テストケースのパターン

1. 基本的なReconcileフロー
   - CAPTClusterが存在しない
   - 親Clusterが存在しない
   - 正常なReconcile

2. VPC関連の処理
   - 既存VPCの使用
   - VPCテンプレートの使用
   - VPC作成エラー

3. 削除処理
   - VPCを保持して削除
   - VPCとWorkspaceTemplateApplyを削除
   - WorkspaceTemplateApplyが存在しない

4. ステータス更新
   - Ready状態への更新
   - エラー状態への更新
   - 条件の更新

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
