# CAPTEP-0031: FluxCD Integration with ClusterResourceSet

## Summary
Cluster APIのClusterResourceSet機能を使用してFluxCDをインストールし、Helmリリースを管理する方法を提案します。

## Motivation
KarpenterなどのHelmチャートベースのアドオンを安定的にインストールするために、FluxCDを使用する必要があります。しかし、FluxCD自体のインストールも信頼性高く行う必要があります。

### Goals
- FluxCDの信頼性の高いインストール方法の確立
- ClusterResourceSetを使用したFluxCDの管理
- FluxCDのインストール状態の追跡

### Non-Goals
- FluxCDの一般的な設定や使用方法の定義
- GitOpsワークフローの確立
- FluxCD以外のGitOpsツールの検討

## Proposal

### FluxCDインストール用ClusterResourceSet

```yaml
apiVersion: addons.cluster.x-k8s.io/v1alpha3
kind: ClusterResourceSet
metadata:
  name: fluxcd-installer
  namespace: default
spec:
  clusterSelector:
    matchLabels:
      fluxcd.io/enabled: "true"
  resources:
    - name: fluxcd-install
      kind: ConfigMap
```

### インストール用マニフェスト

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: fluxcd-install
  namespace: default
data:
  fluxcd.yaml: |
    apiVersion: v1
    kind: Namespace
    metadata:
      name: flux-system
    ---
    apiVersion: source.toolkit.fluxcd.io/v1beta2
    kind: HelmRepository
    metadata:
      name: fluxcd
      namespace: flux-system
    spec:
      interval: 1h
      url: https://fluxcd-community.github.io/helm-charts
    ---
    apiVersion: helm.toolkit.fluxcd.io/v2beta1
    kind: HelmRelease
    metadata:
      name: flux2
      namespace: flux-system
    spec:
      interval: 5m
      chart:
        spec:
          chart: flux2
          version: "2.1.0"
          sourceRef:
            kind: HelmRepository
            name: fluxcd
            namespace: flux-system
      values:
        components:
          - source-controller
          - helm-controller
```

### インストール状態の管理

1. ClusterResourceSetBindingによる状態追跡
```yaml
apiVersion: addons.cluster.x-k8s.io/v1alpha3
kind: ClusterResourceSetBinding
metadata:
  name: cluster-1
spec:
  bindings:
    - clusterResourceSetName: fluxcd-installer
      resources:
        - kind: ConfigMap
          name: fluxcd-install
          applied: true
          hash: sha256:...
```

2. Readinessチェック
```yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: fluxcd-readiness-check
  namespace: flux-system
spec:
  template:
    spec:
      containers:
      - name: kubectl
        image: bitnami/kubectl
        command:
        - /bin/sh
        - -c
        - |
          until kubectl wait --for=condition=Ready pods -n flux-system -l app.kubernetes.io/instance=flux2 --timeout=300s; do
            sleep 5
          done
      restartPolicy: OnFailure
```

### インストール順序の制御

1. ラベル管理
```yaml
apiVersion: cluster.x-k8s.io/v1beta1
kind: Cluster
metadata:
  name: cluster-1
  labels:
    fluxcd.io/enabled: "true"
```

2. 依存関係の管理
- FluxCDのインストール完了を確認後、他のアドオンのインストールを開始
- Readinessチェックジョブの完了を待機

## Implementation Details

### Phase 1: 基本インストール
1. FluxCD用のClusterResourceSetとConfigMapの作成
2. インストール状態の追跡機能の実装
3. 基本的なテストの実装

### Phase 2: 高度な機能
1. Readinessチェックの実装
2. リトライメカニズムの追加
3. エラーハンドリングの改善

### Phase 3: 統合とテスト
1. Karpenterインストールとの統合
2. エンドツーエンドテストの実装
3. ドキュメントの整備

## Risks and Mitigations

### リスク1: インストール失敗
- リスク: ネットワーク問題などによるインストール失敗
- 緩和策:
  - リトライメカニズムの実装
  - タイムアウト設定の最適化
  - エラー通知の実装

### リスク2: バージョン互換性
- リスク: Kubernetes、FluxCD、Helmのバージョン互換性
- 緩和策:
  - バージョンマトリックスの管理
  - 互換性テストの自動化
  - アップグレードガイドラインの整備

## References

- [FluxCD Installation](https://fluxcd.io/docs/installation/)
- [Cluster API ClusterResourceSet](https://cluster-api.sigs.k8s.io/tasks/experimental-features/cluster-resource-set)
- [Helm Controller](https://fluxcd.io/docs/components/helm/)
