# CAPTEP-0020: CAPTClusterのOwnerReferences設定の問題

## 概要

demo-cluster4を作成した際に、CAPTClusterリソースにClusterのownerReferencesが設定されない問題が発生しています。
これにより、以下のエラーが発生しています：

```
ERROR   Reconciler error        {"controller": "captcluster", "controllerGroup": "infrastructure.cluster.x-k8s.io", "controllerKind": "CAPTCluster", "CAPTCluster": {"name":"demo-cluster4","namespace":"default"}, "namespace": "default", "name": "demo-cluster4", "reconcileID": "3633366c-9430-4a55-92bd-8de4851e531b", "error": "no owner cluster found"}
```

## 背景

### Cluster APIの要件
- InfrastructureClusterリソース（CAPTCluster）は、CAPIのClusterリソースによって所有される必要があります
- この所有関係は、Kubernetes OwnerReferencesを通じて表現されます
- OwnerReferencesは、リソース間の階層関係を定義し、ガベージコレクションを制御します

### 最近の変更
- CAPTEP-0018で、ownerReferencesの処理が改善されました
- cluster.x-k8s.io/cluster-nameラベルの設定が追加されました
- Cluster削除イベントのWatchが追加されました

## 問題の分析

### 現在の実装の問題点

1. CAPTClusterコントローラーでの複雑なownerReferences処理
   - 複数の検索パターン
   - 複雑な条件分岐
   - 不必要な検証

2. CAPTControlPlaneとの実装の違い
   - CAPTControlPlaneはシンプルな実装
   - 名前ベースの検索のみ
   - 標準的なownerReference設定

### 想定される原因

1. Infrastructure Providerの実装パターンの誤解
   - 複雑な実装を追加
   - Cluster APIの規約から逸脱

2. ownerReferences設定の責任分担の混乱
   - Cluster APIコントローラーとの役割分担が不明確

## 解決策

### 1. シンプルな実装への移行

```go
// getOwnerCluster returns the owner Cluster for a CAPTCluster
func (r *Reconciler) getOwnerCluster(ctx context.Context, captCluster *infrastructurev1beta1.CAPTCluster) (*clusterv1.Cluster, error) {
    logger := log.FromContext(ctx)

    // Get Cluster by name
    cluster := &clusterv1.Cluster{}
    key := types.NamespacedName{
        Namespace: captCluster.Namespace,
        Name:      captCluster.Name,
    }
    if err := r.Get(ctx, key, cluster); err != nil {
        if !apierrors.IsNotFound(err) {
            logger.Error(err, "Failed to get Cluster")
            return nil, err
        }
        return nil, fmt.Errorf("no owner cluster found")
    }

    return cluster, nil
}
```

### 2. 標準的なownerReference設定

```go
// Set owner reference if cluster exists
if err := controllerutil.SetControllerReference(cluster, captCluster, r.Scheme); err != nil {
    logger.Error(err, "Failed to set owner reference")
    return Result{}, err
}
```

## 設計判断

1. シンプルな実装の採用
   - 名前ベースの検索のみ
   - 標準的なKubernetesパターンの使用
   - CAPTControlPlaneと同様のアプローチ

2. Cluster APIの規約への準拠
   - Infrastructure Providerの責任範囲の明確化
   - Cluster APIコントローラーとの適切な役割分担

3. 複雑な実装の回避
   - 不必要な検証の削除
   - 条件分岐の最小化
   - メンテナンス性の向上

## 実装履歴

- [x] 2024-11-13: 初期提案
- [x] 2024-11-13: CAPTClusterコントローラーの実装修正
- [x] 2024-11-13: 動作確認完了

## 参考資料

- [Cluster API Provider実装ガイド](https://cluster-api.sigs.k8s.io/developer/providers/implementers-guide/controllers_and_reconciliation.html)
- [Kubernetes OwnerReferences](https://kubernetes.io/docs/concepts/overview/working-with-objects/owners-dependents/)
- [CAPTEP-0018: Cluster Owner Reference Handling](docs/CAPTEP/0018-cluster-owner-reference-handling.md)
