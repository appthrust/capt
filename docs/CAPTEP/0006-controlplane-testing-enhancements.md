# CAPTEP-0006: Control Plane Testing Enhancements

## Summary

本提案は、controlplaneコントローラーのテストカバレッジを改善し、Cluster API規約への準拠を強化するための具体的な改善案を提示します。

## Motivation

現状のcontrolplaneコントローラーのテストには以下の課題があります：

1. WorkspaceTemplateの存在確認と処理のテストケースが不足
2. ステータス更新の詳細な検証が不十分
3. エラーシナリオのカバレッジが限定的
4. リソースクリーンアップの検証が不完全

これらの課題に対応し、より信頼性の高いコントローラー実装を実現する必要があります。

## Proposal

### 1. ステータス管理テストの強化

#### フェーズ遷移の検証
```go
{
    name: "Status phase transition test",
    existingObjs: []runtime.Object{
        &controlplanev1beta1.CAPTControlPlane{
            ObjectMeta: metav1.ObjectMeta{
                Name: "test-controlplane",
                Namespace: "default",
            },
            Spec: controlplanev1beta1.CAPTControlPlaneSpec{
                WorkspaceTemplateRef: controlplanev1beta1.WorkspaceTemplateReference{
                    Name: "test-template",
                },
            },
        },
    },
    validate: func(t *testing.T, client client.Client, result Result, err error) {
        controlPlane := &controlplanev1beta1.CAPTControlPlane{}
        err = client.Get(context.Background(), types.NamespacedName{
            Name: "test-controlplane",
            Namespace: "default",
        }, controlPlane)
        
        assert.NoError(t, err)
        
        // フェーズ遷移の検証
        assert.Equal(t, controlplanev1beta1.ControlPlanePhaseProvisioning, controlPlane.Status.Phase)
        
        // 条件の検証
        validateCondition(t, controlPlane.Status.Conditions, 
            controlplanev1beta1.ControlPlaneReadyCondition,
            metav1.ConditionFalse,
            controlplanev1beta1.ReasonCreating)
    },
}
```

### 2. エラーハンドリングテストの追加

#### WorkspaceTemplate不在のケース
```go
{
    name: "WorkspaceTemplate not found error",
    existingObjs: []runtime.Object{
        &controlplanev1beta1.CAPTControlPlane{
            ObjectMeta: metav1.ObjectMeta{
                Name: "test-controlplane",
                Namespace: "default",
            },
            Spec: controlplanev1beta1.CAPTControlPlaneSpec{
                WorkspaceTemplateRef: controlplanev1beta1.WorkspaceTemplateReference{
                    Name: "non-existent-template",
                },
            },
        },
    },
    validate: func(t *testing.T, client client.Client, result Result, err error) {
        assert.Error(t, err)
        assert.True(t, apierrors.IsNotFound(err))
        
        controlPlane := &controlplanev1beta1.CAPTControlPlane{}
        err = client.Get(context.Background(), types.NamespacedName{
            Name: "test-controlplane",
            Namespace: "default",
        }, controlPlane)
        
        assert.NoError(t, err)
        validateCondition(t, controlPlane.Status.Conditions,
            controlplanev1beta1.ControlPlaneReadyCondition,
            metav1.ConditionFalse,
            "WorkspaceTemplateNotFound")
    },
}
```

### 3. リソース管理テストの改善

#### WorkspaceTemplateApply作成の検証
```go
{
    name: "WorkspaceTemplateApply creation",
    existingObjs: []runtime.Object{
        &controlplanev1beta1.CAPTControlPlane{
            ObjectMeta: metav1.ObjectMeta{
                Name: "test-controlplane",
                Namespace: "default",
            },
            Spec: controlplanev1beta1.CAPTControlPlaneSpec{
                WorkspaceTemplateRef: controlplanev1beta1.WorkspaceTemplateReference{
                    Name: "test-template",
                },
            },
        },
        &infrastructurev1beta1.WorkspaceTemplate{
            ObjectMeta: metav1.ObjectMeta{
                Name: "test-template",
                Namespace: "default",
            },
        },
    },
    validate: func(t *testing.T, client client.Client, result Result, err error) {
        assert.NoError(t, err)
        
        // WorkspaceTemplateApplyの検証
        workspaceApply := &infrastructurev1beta1.WorkspaceTemplateApply{}
        err = client.Get(context.Background(), types.NamespacedName{
            Name: "test-controlplane-eks-controlplane-apply",
            Namespace: "default",
        }, workspaceApply)
        
        assert.NoError(t, err)
        assert.Equal(t, "test-template", workspaceApply.Spec.TemplateRef.Name)
    },
}
```

### 4. 削除フローの詳細な検証

#### リソースクリーンアップの順序検証
```go
{
    name: "Resource cleanup order during deletion",
    existingObjs: []runtime.Object{
        &controlplanev1beta1.CAPTControlPlane{
            ObjectMeta: metav1.ObjectMeta{
                Name: "test-controlplane",
                Namespace: "default",
                DeletionTimestamp: &metav1.Time{Time: time.Now()},
                Finalizers: []string{CAPTControlPlaneFinalizer},
            },
            Spec: controlplanev1beta1.CAPTControlPlaneSpec{
                WorkspaceTemplateApplyName: "test-apply",
            },
        },
        &infrastructurev1beta1.WorkspaceTemplateApply{
            ObjectMeta: metav1.ObjectMeta{
                Name: "test-apply",
                Namespace: "default",
            },
        },
    },
    validate: func(t *testing.T, client client.Client, result Result, err error) {
        assert.NoError(t, err)
        
        // WorkspaceTemplateApplyの削除確認
        workspaceApply := &infrastructurev1beta1.WorkspaceTemplateApply{}
        err = client.Get(context.Background(), types.NamespacedName{
            Name: "test-apply",
            Namespace: "default",
        }, workspaceApply)
        assert.True(t, apierrors.IsNotFound(err))
        
        // Finalizerの削除確認
        controlPlane := &controlplanev1beta1.CAPTControlPlane{}
        err = client.Get(context.Background(), types.NamespacedName{
            Name: "test-controlplane",
            Namespace: "default",
        }, controlPlane)
        assert.True(t, apierrors.IsNotFound(err))
    },
}
```

### 5. テストヘルパー関数の導入

```go
// 条件検証用ヘルパー関数
func validateCondition(t *testing.T, conditions []metav1.Condition, 
    conditionType string, status metav1.ConditionStatus, reason string) {
    found := false
    for _, condition := range conditions {
        if condition.Type == conditionType {
            assert.Equal(t, status, condition.Status)
            assert.Equal(t, reason, condition.Reason)
            found = true
            break
        }
    }
    assert.True(t, found, "Expected condition not found")
}

// リソース削除検証用ヘルパー関数
func validateResourceDeletion(t *testing.T, client client.Client, 
    name types.NamespacedName, obj client.Object) {
    err := client.Get(context.Background(), name, obj)
    assert.True(t, apierrors.IsNotFound(err), 
        "Expected resource to be deleted, but it still exists")
}
```

## Benefits

1. テストカバレッジの向上
   - ステータス管理の完全な検証
   - エラーシナリオの網羅的なテスト
   - リソース管理フローの詳細な検証

2. コードの品質向上
   - 標準化されたテストパターン
   - 再利用可能なヘルパー関数
   - 明確なエラーメッセージ

3. 保守性の向上
   - 構造化されたテストケース
   - 理解しやすいテストコード
   - 効率的なデバッグ

## Implementation Plan

1. ヘルパー関数の実装
2. 既存テストケースの改善
3. 新規テストケースの追加
4. テストドキュメントの更新

## References

1. [CAPTEP-0005: Control Plane Testing Improvements](0005-controlplane-testing-improvements.md)
2. [Testing Best Practices Part 2: Controller Deletion Testing](../design/testing-best-practices2.md)
3. [Cluster API Testing Best Practices](https://cluster-api.sigs.k8s.io/developer/testing.html)
