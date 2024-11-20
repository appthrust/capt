# CAPTEP-0033: Terraform Provider Migration

## Summary

このCAPTEPは、CrossplaneのTerraform ProviderからUpboundのTerraform Providerへの移行に関する変更を記録します。

## Motivation

CAPTは、Terraformのワークスペースを管理するためにTerraform Providerを使用しています。これまでは、CrossplaneのTerraform Provider (`tf.crossplane.io`)を使用していましたが、以下の理由により、UpboundのTerraform Provider (`tf.upbound.io`)への移行が必要となりました：

1. より安定したサポート
2. 最新のTerraform機能のサポート
3. 商用サポートの利用可能性

## Proposal

### 技術的な変更

1. RBACアノテーションの更新
   - `tf.crossplane.io`から`tf.upbound.io`へのAPIグループの変更
   - WorkspaceTemplateApplyコントローラーのRBACアノテーションを更新

2. ClusterRoleの追加
   - `capt-tf-upbound-role`の追加
   - 対応するClusterRoleBindingの追加

### 影響範囲

この変更は以下のコンポーネントに影響を与えます：

1. WorkspaceTemplateApplyコントローラー
2. RBACの設定
3. CRDの定義

### 互換性

この変更は後方互換性のない変更となります。既存のクラスターは、新しいProviderに移行する必要があります。

## Implementation

1. WorkspaceTemplateApplyコントローラーのRBACアノテーションを更新
```go
//+kubebuilder:rbac:groups=tf.upbound.io,resources=workspaces,verbs=get;list;watch;create;update;patch;delete
```

2. 必要なClusterRoleとClusterRoleBindingの追加
```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: capt-tf-upbound-role
rules:
- apiGroups:
  - tf.upbound.io
  resources:
  - workspaces
  - workspaces/status
  verbs:
  - create
  - delete
  - get
  - list
  - patch
  - update
  - watch
```

## Testing

1. 既存のE2Eテストスイートを使用して、新しいProviderでの動作を確認
2. 新しいProviderを使用したワークスペースの作成、更新、削除をテスト
3. RBACの権限が正しく設定されていることを確認

## Risks and Mitigations

1. 既存のクラスターへの影響
   - 移行ガイドを提供
   - 段階的な移行プロセスの推奨

2. 権限の問題
   - 詳細なRBACテストの実施
   - 必要最小限の権限の確認

## Alternatives Considered

1. CrossplaneのTerraform Providerの継続使用
   - 長期的なサポートの懸念
   - 機能の制限

2. 独自のTerraform統合の実装
   - 開発コストが高い
   - メンテナンスの負担

## References

- [Upbound Terraform Provider Documentation](https://docs.upbound.io/providers/terraform/)
- [Crossplane to Upbound Migration Guide](https://docs.upbound.io/concepts/migration-guide/)
