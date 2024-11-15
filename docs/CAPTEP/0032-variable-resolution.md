# CAPTEP-0032: Variable Resolution for ClusterResourceSet

## Summary
WorkspaceTemplateが生成するSecretの値をHelmReleaseのtargetPath機能を使用して直接参照する方法を提案します。

## Motivation
WorkspaceTemplateは必要な設定値をSecretとして生成します。これらの値をKarpenterなどのHelmチャートで使用する際、FluxCDのHelmReleaseのtargetPath機能を活用することで、値を直接マージできます。

### Goals
- WorkspaceTemplateが生成するSecretの値を直接利用
- 形式変換なしでの値の参照
- セキュアな変数管理

### Non-Goals
- WorkspaceTemplateのSecret生成形式の変更
- 複雑な変数解決メカニズムの実装
- Secret管理方法の変更

## Proposal

### WorkspaceTemplateのSecret形式

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: ${WORKSPACE_NAME}-eks-connection
  namespace: default
type: Opaque
data:
  cluster_endpoint: <base64_encoded_value>
  cluster_name: <base64_encoded_value>
  karpenter_iam_role_arn: <base64_encoded_value>
  karpenter_queue_name: <base64_encoded_value>
```

### HelmReleaseでのSecret参照

```yaml
apiVersion: helm.toolkit.fluxcd.io/v2beta1
kind: HelmRelease
metadata:
  name: karpenter
  namespace: karpenter
spec:
  interval: 5m
  chart:
    spec:
      chart: karpenter
      version: 1.0.7
      sourceRef:
        kind: HelmRepository
        name: karpenter
        namespace: karpenter
  valuesFrom:
    - kind: Secret
      name: ${WORKSPACE_NAME}-eks-connection
      valuesKey: cluster_name
      targetPath: settings.clusterName
    - kind: Secret
      name: ${WORKSPACE_NAME}-eks-connection
      valuesKey: cluster_endpoint
      targetPath: settings.clusterEndpoint
    - kind: Secret
      name: ${WORKSPACE_NAME}-eks-connection
      valuesKey: karpenter_iam_role_arn
      targetPath: serviceAccount.annotations.eks\.amazonaws\.com/role-arn
    - kind: Secret
      name: ${WORKSPACE_NAME}-eks-connection
      valuesKey: karpenter_queue_name
      targetPath: settings.interruptionQueue
```

### ClusterResourceSetでのSecret転送

```yaml
apiVersion: addons.cluster.x-k8s.io/v1alpha3
kind: ClusterResourceSet
metadata:
  name: karpenter-config
  namespace: default
spec:
  clusterSelector:
    matchLabels:
      karpenter.sh/enabled: "true"
  resources:
    - name: ${WORKSPACE_NAME}-eks-connection
      kind: Secret
```

## Implementation Details

### Phase 1: 基本設定
1. ClusterResourceSetの設定
2. HelmReleaseマニフェストの作成
3. 基本的なテストの実装

### Phase 2: 検証
1. 値の参照テスト
2. エラーケースの確認
3. 統合テストの実装

### Phase 3: ドキュメント化
1. 設定ガイドの作成
2. トラブルシューティングガイドの作成
3. 運用手順の整備

## Risks and Mitigations

### リスク1: Secret値の更新
- リスク: Secret値が更新された場合の反映
- 緩和策:
  - HelmReleaseのinterval設定の最適化
  - 更新状態の監視
  - 手動更新トリガーの提供

### リスク2: パス指定のエラー
- リスク: targetPathの指定ミス
- 緩和策:
  - バリデーション機能の実装
  - エラーメッセージの改善
  - デフォルト値の設定

## References

- [FluxCD HelmRelease Values](https://fluxcd.io/flux/components/helm/helmreleases/#values)
- [Cluster API ClusterResourceSet](https://cluster-api.sigs.k8s.io/tasks/experimental-features/cluster-resource-set)
- [Kubernetes Secrets](https://kubernetes.io/docs/concepts/configuration/secret/)
