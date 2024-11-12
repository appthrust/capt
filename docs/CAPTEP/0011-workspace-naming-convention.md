# CAPTEP-0011: Workspace Naming Convention

## Summary

本提案は、WorkspaceTemplateApplyとWorkspaceの命名規則の不一致による問題と、その解決策を提示します。

## Motivation

CAPTControlPlaneの動作確認中に、以下の問題が発見されました：

1. WorkspaceTemplateApplyが作成するWorkspaceの名前と、コントローラーが期待する名前が一致していない
   - 作成されたWorkspace名: "demo-cluster-vpc-workspace"
   - コントローラーが期待する名前: "demo-cluster-vpc"

この不一致により、以下の問題が発生しています：
- コントローラーがWorkspaceを見つけられない
- ControlPlaneの作成が失敗する

## Goals

- WorkspaceTemplateApplyとWorkspaceの命名規則の統一
- 既存の実装との互換性の維持
- 明確な命名規則のドキュメント化

## Non-Goals

- 既存のWorkspaceの名前変更
- 命名規則以外の機能変更

## Proposal

### 1. 命名規則の統一

```go
// WorkspaceTemplateApply側の命名規則
func generateWorkspaceName(clusterName, component string) string {
    return fmt.Sprintf("%s-%s", clusterName, component)
}

// Workspace側の命名規則も同じ規則を使用
func (r *Reconciler) reconcileWorkspace(ctx context.Context, workspaceApply *infrastructurev1beta1.WorkspaceTemplateApply) error {
    workspaceName := generateWorkspaceName(workspaceApply.Spec.ClusterName, workspaceApply.Spec.Component)
    // ...
}
```

### 2. 移行戦略

1. 新しい命名規則の導入
```go
const (
    WorkspaceNameFormat = "%s-%s"  // clusterName-component
)
```

2. 既存のWorkspaceの処理
```go
func (r *Reconciler) findWorkspace(ctx context.Context, name string) (*tfv1beta1.Workspace, error) {
    // 新しい命名規則で検索
    workspace := &tfv1beta1.Workspace{}
    err := r.Get(ctx, types.NamespacedName{Name: name}, workspace)
    if err == nil {
        return workspace, nil
    }

    // 古い命名規則で検索（後方互換性のため）
    legacyName := fmt.Sprintf("%s-workspace", name)
    err = r.Get(ctx, types.NamespacedName{Name: legacyName}, workspace)
    if err == nil {
        return workspace, nil
    }

    return nil, err
}
```

## Implementation Details

### Phase 1: コードの修正

1. 命名規則の定義
```go
// pkg/util/naming/workspace.go
package naming

const (
    WorkspaceNameFormat = "%s-%s"
)

func GenerateWorkspaceName(clusterName, component string) string {
    return fmt.Sprintf(WorkspaceNameFormat, clusterName, component)
}
```

2. コントローラーの更新
```go
func (r *Reconciler) reconcileWorkspace(ctx context.Context, apply *infrastructurev1beta1.WorkspaceTemplateApply) error {
    workspaceName := naming.GenerateWorkspaceName(apply.Spec.ClusterName, apply.Spec.Component)
    // ...
}
```

### Phase 2: テストの更新

```go
func TestWorkspaceNaming(t *testing.T) {
    tests := []struct {
        name        string
        clusterName string
        component   string
        expected    string
    }{
        {
            name:        "VPC workspace",
            clusterName: "demo-cluster",
            component:   "vpc",
            expected:    "demo-cluster-vpc",
        },
        // ...
    }
    // ...
}
```

## Risks and Mitigations

### リスク

1. 既存のWorkspaceとの互換性
   - リスク: 既存のWorkspaceが見つからない
   - 対策: 両方の命名規則をサポート

2. 移行中の混乱
   - リスク: 異なる命名規則の共存
   - 対策: 明確なドキュメント化と段階的な移行

## Alternatives Considered

1. 現状の命名規則を維持
   - 却下理由: 問題の根本的な解決にならない
   - 継続的な問題の可能性

2. 完全に新しい命名規則の導入
   - 却下理由: 移行コストが高い
   - 既存システムへの影響が大きい

## References

1. [VPC Creation Test Results](../design/vpc-creation-test-results.md)
2. [CAPTEP-0010: VPC Creation Validation](0010-vpc-creation-validation.md)

## Implementation History

- 2024-11-12: 問題の発見と初期提案
- 2024-11-12: ドキュメント作成
