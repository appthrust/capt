# CAPTEP-0038: Kubeconfig Secret RBAC Permissions

## Summary

CAPTコントローラーがクラスター用のkubeconfigを含むSecretを作成/更新する際に、必要なRBAC権限が不足していることが原因でエラーが発生する問題に対処します。

## Motivation

### Goals

- kubeconfigのSecretを適切に作成/更新できるようにRBAC権限を設定する
- 権限の範囲を必要最小限に保ちながら、必要な操作を可能にする

### Non-Goals

- 既存のSecret管理ロジックの変更
- 他のリソースに対するRBAC権限の変更

## Proposal

### User Stories

#### Story 1: クラスター作成時のkubeconfig生成

クラスター管理者として、新しいクラスターを作成する際に、そのクラスターにアクセスするためのkubeconfigが自動的に生成され、Secretとして保存されることを期待します。

#### Story 2: クラスター更新時のkubeconfig更新

クラスター管理者として、クラスターの設定が変更された際に、kubeconfigが適切に更新され、Secretが最新の状態に保たれることを期待します。

### Implementation Details

#### RBAC権限の追加

capt-manager-roleのClusterRoleに以下の権限を追加します：

```yaml
- apiGroups:
  - ""
  resources:
  - secrets
  verbs:
  - create
  - update
  - patch
```

これにより、コントローラーは以下の操作が可能になります：
- 新しいkubeconfigのSecretの作成
- 既存のSecretの更新
- 必要に応じたSecretの部分的な更新（patch）

### Risks and Mitigations

#### セキュリティリスク

- **リスク**: Secretsに対する追加の権限付与によるセキュリティ上の懸念
- **対策**: 
  - 権限をkubeconfig Secretsに限定することを検討（今後の改善として）
  - 監査ログを活用して、Secret操作の監視を強化

#### 運用リスク

- **リスク**: 権限変更による既存の動作への影響
- **対策**:
  - 変更を段階的に適用
  - テスト環境での十分な検証

## Design Details

### Test Plan

1. ユニットテスト
   - Secret作成/更新の権限チェック
   - エラーハンドリングの確認

2. E2Eテスト
   - クラスター作成時のSecret生成
   - クラスター更新時のSecret更新
   - 権限エラーが発生しないことの確認

### Upgrade Strategy

1. 既存のクラスターへの影響
   - 既存のクラスターは再起動後に新しい権限で動作
   - 手動での再適用は不要

2. ダウングレード考慮事項
   - 互換性の問題なし
   - 以前のバージョンへの復帰も可能

## Implementation History

- [ ] 2024-01-XX: 初版作成
- [ ] RBAC権限の追加
- [ ] テストの実装と実行
- [ ] ドキュメントの更新

## Alternatives

### 代替案1: Secretsに特化したサービスアカウントの作成

新しいサービスアカウントを作成し、Secrets管理専用の権限を付与する方法も検討しましたが、以下の理由で採用しませんでした：
- 複雑性の増加
- 運用負荷の増加
- 既存のアーキテクチャとの整合性

### 代替案2: Secretsの代わりにConfigMapの使用

機密性の低い情報をConfigMapに保存する方法も検討しましたが、以下の理由で採用しませんでした：
- kubeconfigには機密情報が含まれるため、Secretsの使用が適切
- Kubernetes標準の実装方法との整合性

## References

- [Kubernetes RBAC Documentation](https://kubernetes.io/docs/reference/access-authn-authz/rbac/)
- [Cluster API Documentation](https://cluster-api.sigs.k8s.io/)
