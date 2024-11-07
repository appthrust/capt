# Secret Naming Design in CAPTControlPlane

## Overview

このドキュメントでは、CAPTControlPlaneにおけるSecret命名の設計と実装について説明します。特に、WorkspaceTemplateから作成されるWorkspaceのsecretsの名前にcontrolplane名を含める実装に関する設計判断と知見をまとめます。

## Background

以前の実装では、WorkspaceTemplateから作成されるWorkspaceのsecretsの名前が固定値（"eks-connection"）でした。これには以下の問題がありました：

1. 複数のControlPlaneが同じNamespaceに存在する場合、Secret名が衝突する可能性がある
2. Secretとそれを所有するControlPlaneの関係が名前から分かりにくい
3. Secret名の一貫性が保たれていない（他のSecretは`{controlplane-name}-{purpose}`の形式）

## Design

### Secret命名規則

Secret名を以下の形式に統一することを決定しました：

```
{controlplane-name}-{purpose}
```

例：
- `test-controlplane-eks-connection`
- `test-controlplane-ca`
- `test-controlplane-kubeconfig`

この命名規則には以下の利点があります：

1. Secret名の衝突を防ぐ
2. Secret名から所有者（ControlPlane）を特定できる
3. 命名規則の一貫性を保つ

### 実装の重要なポイント

1. **Secret名の生成**
   - ControlPlaneの名前をベースに生成
   - 目的を示すサフィックスを追加
   ```go
   secretName := fmt.Sprintf("%s-eks-connection", controlPlane.Name)
   ```

2. **Secret管理の責任分離**
   - SecretManagerをControlPlaneベースに変更
   - GetAndValidateSecretメソッドの引数をControlPlaneに変更
   ```go
   func (m *SecretManager) GetAndValidateSecret(ctx context.Context, controlPlane *controlplanev1beta1.CAPTControlPlane) (*corev1.Secret, error)
   ```

3. **WorkspaceTemplateApplyとの連携**
   - WriteConnectionSecretToRefの名前もControlPlaneの名前から生成
   ```go
   WriteConnectionSecretToRef: &xpv1.SecretReference{
       Name:      fmt.Sprintf("%s-eks-connection", controlPlane.Name),
       Namespace: controlPlane.Namespace,
   }
   ```

## Testing Strategy

テストでは以下の点を重点的に確認：

1. **基本的な動作確認**
   - WriteConnectionSecretToRefが正しく設定されることを確認
   - Secret名がControlPlaneの名前を含むことを確認

2. **更新時の動作確認**
   - ControlPlaneの更新後もSecret名が維持されることを確認
   - 既存のSecretとの整合性が保たれることを確認

3. **エラーケースの処理**
   - WorkspaceTemplateApplyのStatusが設定されていない場合の処理
   - Secret取得失敗時のエラーハンドリング

## Lessons Learned

1. **テストの重要性**
   - 複雑な依存関係（Workspace、Secret、ControlPlane）を持つシステムでは、テストの順序が重要
   - モックやスタブの代わりに実際のリソースを使用することで、より現実的なテストが可能

2. **エラーハンドリング**
   - 初期状態（StatusなしのWorkspaceTemplateApply）を適切に処理することの重要性
   - エラーメッセージの明確化による問題の早期発見

3. **設計の一貫性**
   - 命名規則の統一による保守性の向上
   - リソース間の関係を名前に反映させることの利点

## Future Considerations

1. **名前の衝突回避**
   - 将来的に必要になる可能性のある追加のプレフィックスやサフィックス
   - 名前の長さ制限への対応

2. **Secret管理の拡張**
   - 他のタイプのSecretへの対応
   - Secret名のカスタマイズオプションの提供

3. **テストの改善**
   - より多くのエッジケースのカバー
   - テストヘルパー関数の整理と共通化
