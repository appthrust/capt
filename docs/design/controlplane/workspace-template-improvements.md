# WorkspaceTemplate Handling Improvements

## 概要

このドキュメントでは、CAPTControlPlaneコントローラーにおけるWorkspaceTemplate処理の改善点について説明します。

## 背景

以前の実装では以下の課題がありました：

1. WorkspaceTemplateApplyの作成時に常にVPCワークスペースへの依存関係が追加され、不要なタイムアウトが発生
2. WriteConnectionSecretToRefの設定が不十分
3. テストケースが一部の機能をカバーしていない

## 改善点

### 1. VPCワークスペース依存関係の最適化

```go
// VPC workspace dependency is now conditional
vpcWorkspaceName := fmt.Sprintf("%s-vpc", controlPlane.Name)
vpcWorkspace := &infrastructurev1beta1.WorkspaceTemplateApply{}
err := r.Get(context.Background(), types.NamespacedName{
    Name:      vpcWorkspaceName,
    Namespace: controlPlane.Namespace,
}, vpcWorkspace)
if err == nil {
    spec.WaitForWorkspaces = []infrastructurev1beta1.WorkspaceReference{
        {
            Name:      vpcWorkspaceName,
            Namespace: controlPlane.Namespace,
        },
    }
}
```

- VPCワークスペースが存在する場合のみ依存関係を追加
- 不要なタイムアウトを防止
- より柔軟な構成をサポート

### 2. WriteConnectionSecretToRefの改善

```go
spec := infrastructurev1beta1.WorkspaceTemplateApplySpec{
    // ...
    WriteConnectionSecretToRef: &xpv1.SecretReference{
        Name:      fmt.Sprintf("%s-eks-connection", controlPlane.Name),
        Namespace: controlPlane.Namespace,
    },
}
```

- 接続情報の保存先を明示的に指定
- 一貫性のある命名規則の適用
- セキュリティ考慮事項の明確化

### 3. テストカバレッジの向上

新しいテストケースの追加：

1. VPC依存関係の検証
```go
"Successfully reconcile workspace with VPC dependency"
"Successfully reconcile workspace without VPC dependency"
```

2. WriteConnectionSecretToRefの検証
```go
assert.NotNil(t, spec.WriteConnectionSecretToRef, "Should have connection secret ref")
assert.Equal(t, fmt.Sprintf("%s-eks-connection", controlPlane.Name), spec.WriteConnectionSecretToRef.Name)
```

3. エラーケースの改善
```go
"Template not found"
```

## 影響と利点

1. パフォーマンスの向上
   - 不要な待機時間の削減
   - より効率的なリソース管理

2. 信頼性の向上
   - エラーケースの適切な処理
   - より堅牢なテスト

3. 保守性の向上
   - コードの可読性向上
   - 明確な設計意図の文書化

## 今後の課題

1. パフォーマンスモニタリング
   - タイムアウトの監視
   - リソース使用状況の追跡

2. エラーハンドリング
   - より詳細なエラーメッセージ
   - リカバリーメカニズムの改善

3. テスト強化
   - エッジケースのカバレッジ向上
   - 統合テストの追加

## 参考資料

- [CAPTEP-0007: Control Plane Event Recording](../../CAPTEP/0007-controlplane-event-recording.md)
- [Workspace Template API Specification](../../workspace-template-api-spec.md)
