# VPC維持機能の設計

## 概要

このドキュメントでは、CAPTClusterにおけるVPC維持機能の設計と実装について説明します。この機能は、親クラスタが削除された場合でも、必要に応じてVPCリソースを維持することを可能にします。

## 背景

### 課題

1. VPCリソースは複数のプロジェクトで共有されることがある
2. 親クラスタの削除時に、共有されているVPCも削除されてしまう
3. VPCの削除は他のプロジェクトに影響を与える可能性がある

### 要件

1. VPCの維持/削除を制御できる機能が必要
2. この制御は明示的に設定可能である必要がある
3. 既存のVPCを使用する場合は影響を受けない

## 設計の決定

### API設計

#### CaptCluster

```go
type CaptClusterSpec struct {
    // RetainVPCOnDelete は親クラスタが削除された時にVPCを維持するかどうかを指定します
    // これは、VPCが複数のプロジェクトで共有されている場合に有用です
    // このフィールドはVPCTemplateRefが設定されている場合のみ有効です
    // +optional
    RetainVPCOnDelete bool `json:"retainVpcOnDelete,omitempty"`
}
```

#### WorkspaceTemplateApply

```go
type WorkspaceTemplateApplySpec struct {
    // RetainWorkspaceOnDelete はこのWorkspaceTemplateApplyが削除された時にWorkspaceを維持するかどうかを指定します
    // これは、Workspaceが共有リソースを管理しており、このWorkspaceTemplateApplyよりも長く存続させる必要がある場合に有用です
    // +optional
    RetainWorkspaceOnDelete bool `json:"retainWorkspaceOnDelete,omitempty"`
}
```

### 設計上の考慮点

1. **明示的な設定**
   - CaptCluster: デフォルトではfalse（VPCは削除される）
   - WorkspaceTemplateApply: デフォルトではfalse（Workspaceは削除される）
   - 明示的にtrueを設定した場合のみリソースが維持される

2. **適用範囲の制限**
   - CaptCluster: VPCTemplateRefを使用する場合のみ有効
   - WorkspaceTemplateApply: すべてのケースで有効

3. **バリデーション**
   - RetainVPCOnDeleteはVPCTemplateRefが設定されている場合のみ有効
   - 不正な組み合わせは早期に検出される

### WorkspaceTemplateApplyの保持機能

1. **設計の背景**
   - WorkspaceTemplateApplyの削除時に、関連するWorkspaceも自動的に削除される現在の動作
   - この動作は共有リソース（VPCなど）の保持が必要なケースで問題となる

2. **新しい設計方針**
   - WorkspaceTemplateApplyに直接Workspace保持設定を追加
   - 親リソースの設定に依存せず、各WorkspaceTemplateApplyで個別に制御可能

3. **利点**
   - 柔軟性：各WorkspaceTemplateApplyで個別に制御可能
   - 明確性：リソースの保持意図が明示的に記述される
   - 再利用性：同じWorkspaceTemplateを使用する異なるユースケースで、異なる保持戦略を適用可能

### 実装の重要ポイント

1. **削除処理の制御**
```go
func (r *workspaceTemplateApplyReconciler) reconcileDelete(ctx context.Context, workspaceApply *infrastructurev1beta1.WorkspaceTemplateApply) (ctrl.Result, error) {
    if workspaceApply.Spec.RetainWorkspaceOnDelete {
        // Workspaceを維持する場合は、削除処理をスキップ
        return ctrl.Result{}, nil
    }
    // 通常の削除処理
    ...
}
```

2. **バリデーション**
```go
func (s *WorkspaceTemplateApplySpec) ValidateConfiguration() error {
    // 必要に応じてバリデーションを追加
    return nil
}
```

3. **ログ記録**
   - リソースの維持/削除の判断を明確に記録
   - トラブルシューティングを容易にする

## 使用例

### VPC保持の設定

```yaml
apiVersion: infrastructure.cluster.x-k8s.io/v1beta1
kind: CAPTCluster
metadata:
  name: example-retained-vpc
spec:
  region: ap-northeast-1
  vpcTemplateRef:
    name: vpc-template
  retainVpcOnDelete: true  # VPCを維持する設定
```

### Workspace保持の設定

```yaml
apiVersion: infrastructure.cluster.x-k8s.io/v1beta1
kind: WorkspaceTemplateApply
metadata:
  name: shared-vpc
spec:
  templateRef:
    name: vpc-template
  retainWorkspaceOnDelete: true  # Workspaceを保持する設定
```

## 学んだ教訓

1. **機能の分離**
   - リソースの維持/削除の制御は、各リソースレベルで独立して管理する必要がある
   - これにより、設定の意図が明確になり、誤用を防ぐことができる

2. **明示的な設定の重要性**
   - デフォルトでは安全側（リソースを削除）に倒す
   - 維持が必要な場合は明示的な設定を要求する

3. **バリデーションの重要性**
   - 早期のバリデーションにより、設定ミスを防ぐ
   - エラーメッセージは具体的で理解しやすいものにする

4. **ドキュメントとサンプル**
   - 機能の使用方法を明確に示すサンプルが重要
   - 設定の意図と影響を理解しやすくする

## 今後の検討事項

1. **拡張性**
   - 他のリソースタイプにも同様の維持機能が必要になる可能性
   - 共通のパターンとして抽象化を検討

2. **モニタリング**
   - リソースの維持/削除の決定を監視可能にする
   - メトリクスの収集を検討

3. **ライフサイクル管理**
   - 維持されたリソースの管理方法
   - クリーンアップポリシーの検討

## 参考資料

- [Terraform Outputs Management](./terraform-outputs-management.md)
- [Cluster Status Management](./cluster-status-management.md)
