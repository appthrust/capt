# CAPTEP-0037: Kubeconfig Secret Updates

## Summary

このプロポーザルでは、kubeconfigシークレットの更新メカニズムを改善し、WorkspaceTemplateから生成される最新のkubeconfigが常にCluster API互換のシークレットに反映されるようにします。

## Motivation

現在の実装では、kubeconfigシークレットは初回作成時のみ生成され、その後の更新が行われません。これにより以下の問題が発生しています：

1. EKSトークンの更新が反映されない
2. クラスターエンドポイントやCA証明書の変更が反映されない
3. WorkspaceTemplateで生成される最新のkubeconfigが使用されない

### Goals

- kubeconfigシークレットの自動更新メカニズムの実装
- 既存のシークレット管理との互換性維持
- 適切なエラーハンドリングとログ記録

### Non-Goals

- kubeconfig生成ロジックの変更
- シークレット命名規則の変更
- 認証メカニズムの変更

## Proposal

### Implementation Details

`reconcileKubeconfigSecret`関数を以下のように修正します：

```go
func (r *Reconciler) reconcileKubeconfigSecret(ctx context.Context, controlPlane *controlplanev1beta1.CAPTControlPlane, cluster *clusterv1.Cluster) error {
    logger := log.FromContext(ctx)

    // Get outputs secret first
    outputsSecretName := fmt.Sprintf("%s-outputs-kubeconfig", cluster.Name)
    outputsSecret := &corev1.Secret{}
    if err := r.Get(ctx, client.ObjectKey{
        Name:      outputsSecretName,
        Namespace: "default",
    }, outputsSecret); err != nil {
        if !apierrors.IsNotFound(err) {
            logger.Error(err, "Failed to get outputs secret")
            return err
        }
        logger.Info("Waiting for outputs secret to be created", "name", outputsSecretName)
        return nil
    }

    // Prepare kubeconfig secret
    kubeconfigSecretName := fmt.Sprintf("%s-kubeconfig", cluster.Name)
    kubeconfigSecret := &corev1.Secret{
        ObjectMeta: metav1.ObjectMeta{
            Name:      kubeconfigSecretName,
            Namespace: controlPlane.Namespace,
            Labels: map[string]string{
                "cluster.x-k8s.io/cluster-name": cluster.Name,
            },
        },
        Type: "cluster.x-k8s.io/secret",
        Data: map[string][]byte{
            "value": outputsSecret.Data["kubeconfig"],
        },
    }

    // Set controller reference
    if err := controllerutil.SetControllerReference(controlPlane, kubeconfigSecret, r.Scheme); err != nil {
        logger.Error(err, "Failed to set controller reference for kubeconfig secret")
        return err
    }

    // Create or update kubeconfig secret
    existingKubeconfigSecret := &corev1.Secret{}
    err := r.Get(ctx, client.ObjectKey{
        Name:      kubeconfigSecretName,
        Namespace: controlPlane.Namespace,
    }, existingKubeconfigSecret)

    if err != nil {
        if !apierrors.IsNotFound(err) {
            logger.Error(err, "Failed to get existing kubeconfig secret")
            return err
        }
        // Create new secret
        if err := r.Create(ctx, kubeconfigSecret); err != nil {
            logger.Error(err, "Failed to create kubeconfig secret")
            return err
        }
        logger.Info("Created kubeconfig secret")
    } else {
        // Update existing secret
        existingKubeconfigSecret.Data = kubeconfigSecret.Data
        existingKubeconfigSecret.Labels = kubeconfigSecret.Labels
        if err := r.Update(ctx, existingKubeconfigSecret); err != nil {
            logger.Error(err, "Failed to update kubeconfig secret")
            return err
        }
        logger.Info("Updated kubeconfig secret")
    }

    return nil
}
```

### リスクと対策

1. パフォーマンスリスク：
   - 定期的な更新によるAPIサーバーの負荷増加
   - 対策：必要な場合のみ更新を実行（データの変更検知）

2. 互換性リスク：
   - 既存のシークレット参照への影響
   - 対策：同じ命名規則とラベルを維持

3. セキュリティリスク：
   - シークレット更新中のデータ露出
   - 対策：適切なRBACとowner reference設定

### テストプラン

1. ユニットテスト：
   - シークレット作成・更新ロジックのテスト
   - エラーハンドリングのテスト
   - owner referenceの設定テスト

2. 統合テスト：
   - 完全なクラスター作成フローでのテスト
   - トークン更新シナリオのテスト
   - エラー回復シナリオのテスト

## Implementation History

- 2024-01-25: 初期プロポーザル作成
- 2024-01-25: 実装完了

## Alternatives Considered

1. ポーリングベースの更新：
   - 却下：リソース消費が大きい
   - 代わりにReconcileループを利用

2. イベントベースの更新：
   - 却下：複雑性が増加
   - 現在のReconcileパターンで十分

3. 差分検出による更新：
   - 却下：オーバーヘッドが大きい
   - 単純な更新の方が信頼性が高い

## References

- [CAPTEP-0034: Dedicated WorkspaceTemplate for kubeconfig generation](./0034-kubeconfig-generation-workspace.md)
- [CAPTEP-0036: Kubeconfig Generation Improvements](./0036-kubeconfig-generation-improvements.md)
