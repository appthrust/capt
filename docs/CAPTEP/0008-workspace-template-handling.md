# CAPTEP-0008: WorkspaceTemplate Handling Improvements

## Summary

本提案は、CAPTControlPlaneコントローラーにおけるWorkspaceTemplateの処理を改善し、より効率的で信頼性の高い実装を実現するための具体的な改善案を提示します。

## Motivation

現状のWorkspaceTemplate処理には以下の課題があります：

1. 不必要なVPCワークスペース依存関係によるタイムアウト
2. 接続情報の管理が不十分
3. テストカバレッジの不足

これらの課題に対応し、より効率的で信頼性の高い実装を実現する必要があります。

## Goals

- WorkspaceTemplateApply作成処理の最適化
- 接続情報管理の改善
- テストカバレッジの向上

## Non-Goals

- WorkspaceTemplateのAPI仕様の変更
- 既存の依存関係管理の完全な見直し
- パフォーマンスの最適化（タイムアウト対策以外）

## Proposal

### 1. 条件付きVPC依存関係

```go
func (r *Reconciler) generateWorkspaceTemplateApplySpec(controlPlane *controlplanev1beta1.CAPTControlPlane) infrastructurev1beta1.WorkspaceTemplateApplySpec {
    spec := infrastructurev1beta1.WorkspaceTemplateApplySpec{
        // 基本設定
        TemplateRef: infrastructurev1beta1.WorkspaceTemplateReference{
            Name:      controlPlane.Spec.WorkspaceTemplateRef.Name,
            Namespace: controlPlane.Spec.WorkspaceTemplateRef.Namespace,
        },
        Variables: map[string]string{
            "cluster_name":       controlPlane.Name,
            "kubernetes_version": controlPlane.Spec.Version,
        },
        WriteConnectionSecretToRef: &xpv1.SecretReference{
            Name:      fmt.Sprintf("%s-eks-connection", controlPlane.Name),
            Namespace: controlPlane.Namespace,
        },
    }

    // VPC依存関係の条件付き追加
    vpcWorkspaceName := fmt.Sprintf("%s-vpc", controlPlane.Name)
    vpcWorkspace := &infrastructurev1beta1.WorkspaceTemplateApply{}
    if err := r.Get(context.Background(), types.NamespacedName{
        Name:      vpcWorkspaceName,
        Namespace: controlPlane.Namespace,
    }, vpcWorkspace); err == nil {
        spec.WaitForWorkspaces = []infrastructurev1beta1.WorkspaceReference{
            {
                Name:      vpcWorkspaceName,
                Namespace: controlPlane.Namespace,
            },
        }
    }

    return spec
}
```

### 2. テストの改善

1. VPC依存関係のテスト
```go
func TestReconcileWorkspace(t *testing.T) {
    tests := []struct {
        name           string
        controlPlane   *controlplanev1beta1.CAPTControlPlane
        vpcWorkspace   *infrastructurev1beta1.WorkspaceTemplateApply
        expectedResult ctrl.Result
        validate      func(t *testing.T, workspaceApply *infrastructurev1beta1.WorkspaceTemplateApply)
    }{
        {
            name: "With VPC dependency",
            // ...
            validate: func(t *testing.T, workspaceApply *infrastructurev1beta1.WorkspaceTemplateApply) {
                assert.Len(t, workspaceApply.Spec.WaitForWorkspaces, 1)
                assert.Equal(t, "test-controlplane-vpc", workspaceApply.Spec.WaitForWorkspaces[0].Name)
            },
        },
        {
            name: "Without VPC dependency",
            // ...
            validate: func(t *testing.T, workspaceApply *infrastructurev1beta1.WorkspaceTemplateApply) {
                assert.Empty(t, workspaceApply.Spec.WaitForWorkspaces)
            },
        },
    }
    // ...
}
```

## Implementation Details

### Phase 1: 基盤整備

1. WorkspaceTemplateApply生成ロジックの改善
2. 接続情報管理の実装
3. テストヘルパー関数の導入

### Phase 2: テスト強化

1. 新規テストケースの追加
2. エラーケースの改善
3. テストカバレッジの確認

### Phase 3: 検証と改善

1. パフォーマンステスト
2. エラーハンドリングの検証
3. ドキュメントの更新

## Migration Plan

1. コードの変更
   - WorkspaceTemplateApply生成ロジックの更新
   - テストの追加と更新
   - エラーハンドリングの改善

2. ドキュメントの更新
   - 設計ドキュメントの作成
   - テスト戦略の文書化
   - 運用ガイドの更新

## Risks and Mitigations

### リスク

1. 既存の依存関係への影響
   - リスク: 既存のワークスペース連携の破損
   - 対策: 段階的な導入とテスト

2. パフォーマンスへの影響
   - リスク: VPC確認による遅延
   - 対策: キャッシュの検討

3. エラーハンドリング
   - リスク: 新しいエラーケースの発生
   - 対策: 包括的なテストカバレッジ

## Alternatives Considered

1. 静的依存関係の維持
   - 却下理由: タイムアウト問題の継続
   - 柔軟性の欠如

2. 依存関係の完全な削除
   - 却下理由: 一部のユースケースで必要
   - 運用上の制約

## References

1. [CAPTEP-0007: Control Plane Event Recording](0007-controlplane-event-recording.md)
2. [Workspace Template Design](../design/controlplane/workspace-template-improvements.md)
3. [Testing Best Practices](../design/testing-best-practices.md)

## Implementation History

- 2024-11-12: 初期提案
- (Future dates to be added as implementation progresses)
