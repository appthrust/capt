# VPC Retention Design

## Overview

このドキュメントでは、CAPTClusterにおけるVPC維持機能の設計と実装について説明します。この機能は、親クラスタが削除された場合でも、必要に応じてVPCリソースを維持することを可能にします。

## Background

### 課題

1. VPCリソースは複数のプロジェクトで共有されることがある
2. 親クラスタの削除時に、共有されているVPCも削除されてしまう
3. VPCの削除は他のプロジェクトに影響を与える可能性がある

### 要件

1. VPCの維持/削除を制御できる機能が必要
2. この制御は明示的に設定可能である必要がある
3. 既存のVPCを使用する場合は影響を受けない

## Design Decision

### API設計

```go
type CaptClusterSpec struct {
    // RetainVPCOnDelete specifies whether to retain the VPC when the parent cluster is deleted
    // This is useful when the VPC is shared among multiple projects
    // This field is only effective when VPCTemplateRef is set
    // +optional
    RetainVPCOnDelete bool `json:"retainVpcOnDelete,omitempty"`
}
```

### 設計上の考慮点

1. **明示的な設定**
   - デフォルトではfalse（VPCは削除される）
   - 明示的にtrueを設定した場合のみVPCが維持される

2. **適用範囲の制限**
   - VPCTemplateRefを使用する場合のみ有効
   - ExistingVPCIDを使用する場合は無関係

3. **バリデーション**
   - RetainVPCOnDeleteはVPCTemplateRefが設定されている場合のみ有効
   - 不正な組み合わせは早期に検出される

### 実装の重要ポイント

1. **削除処理の制御**
```go
func (r *CAPTClusterReconciler) reconcileDelete(ctx context.Context, captCluster *infrastructurev1beta1.CAPTCluster) (ctrl.Result, error) {
    if captCluster.Spec.RetainVPCOnDelete && captCluster.Spec.VPCTemplateRef != nil {
        // VPCを維持する場合は、WorkspaceTemplateApplyを削除しない
        return ctrl.Result{}, nil
    }
    // 通常の削除処理
    ...
}
```

2. **バリデーション**
```go
func (s *CAPTClusterSpec) ValidateVPCConfiguration() error {
    if s.RetainVPCOnDelete && s.VPCTemplateRef == nil {
        return fmt.Errorf("retainVpcOnDelete can only be set when VPCTemplateRef is specified")
    }
    ...
}
```

3. **ログ記録**
   - VPCの維持/削除の判断を明確に記録
   - トラブルシューティングを容易にする

## Usage Example

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

## Lessons Learned

1. **機能の分離**
   - VPCの維持/削除の制御は、他の設定と独立して管理する必要がある
   - これにより、設定の意図が明確になり、誤用を防ぐことができる

2. **明示的な設定の重要性**
   - デフォルトでは安全側（VPCを削除）に倒す
   - 維持が必要な場合は明示的な設定を要求する

3. **バリデーションの重要性**
   - 早期のバリデーションにより、設定ミスを防ぐ
   - エラーメッセージは具体的で理解しやすいものにする

4. **ドキュメントとサンプル**
   - 機能の使用方法を明確に示すサンプルが重要
   - 設定の意図と影響を理解しやすくする

## Future Considerations

1. **拡張性**
   - 他のリソースタイプにも同様の維持機能が必要になる可能性
   - 共通のパターンとして抽象化を検討

2. **モニタリング**
   - VPCの維持/削除の決定を監視可能にする
   - メトリクスの収集を検討

3. **ライフサイクル管理**
   - 維持されたVPCの管理方法
   - クリーンアップポリシーの検討

## References

- [Terraform Outputs Management](./terraform-outputs-management.md)
- [Cluster Status Management](./cluster-status-management.md)
