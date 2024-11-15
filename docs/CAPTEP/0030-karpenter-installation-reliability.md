# CAPTEP-0030: Karpenter Installation Using ClusterResourceSet

## Summary
現在のTerraform Helm Providerを使用したKarpenterのインストール方法をCluster APIのClusterResourceSet機能を使用する方式に変更し、インストールの信頼性を向上させることを提案します。

## Motivation
現在、eks-controlplane-template-v2.yamlではTerraformのhelm_releaseリソースを使用してKarpenterをインストールしていますが、以下の問題が報告されています：

1. Terraform Helm Providerの不安定性
   - クラスター作成時のHelmインストールが信頼性に欠ける
   - インストール失敗時のリカバリーが困難
2. タイミングの問題
   - EKSクラスター作成直後のHelmインストールは、クラスターの準備が完全に整う前に実行される可能性がある

### Goals
- Karpenterインストールの信頼性を向上させる
- クラスター作成とアドオンインストールの適切な分離
- より堅牢なエラーハンドリングとリカバリーメカニズムの実現
- FluxCDとKarpenterのインストール順序の適切な管理

### Non-Goals
- Karpenter以外のアドオンインストール方法の変更
- クラスター作成プロセス全体の見直し

## Proposal
Cluster APIのClusterResourceSet機能を使用して、FluxCDとKarpenterのインストールを管理します。

### Implementation Details

1. FluxCD用のClusterResourceSet
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

2. FluxCDインストール用ConfigMap
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
    apiVersion: v1
    kind: ServiceAccount
    metadata:
      name: flux-installer
      namespace: flux-system
    ---
    apiVersion: batch/v1
    kind: Job
    metadata:
      name: flux-installer
      namespace: flux-system
    spec:
      template:
        spec:
          serviceAccountName: flux-installer
          containers:
          - name: flux
            image: fluxcd/flux-cli:v2.1.0
            command:
            - /bin/sh
            - -c
            - |
              flux install \
                --namespace=flux-system \
                --network-policy=false \
                --components=source-controller,helm-controller
          restartPolicy: OnFailure
```

3. Karpenter用のClusterResourceSet（FluxCD依存）
```yaml
apiVersion: addons.cluster.x-k8s.io/v1alpha3
kind: ClusterResourceSet
metadata:
  name: karpenter-installer
  namespace: default
spec:
  clusterSelector:
    matchLabels:
      karpenter.sh/enabled: "true"
  resources:
    - name: karpenter-resources
      kind: ConfigMap
```

4. Karpenterインストール用ConfigMap
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: karpenter-resources
  namespace: default
data:
  karpenter.yaml: |
    apiVersion: v1
    kind: Namespace
    metadata:
      name: karpenter
    ---
    apiVersion: source.toolkit.fluxcd.io/v1beta2
    kind: HelmRepository
    metadata:
      name: karpenter
      namespace: karpenter
    spec:
      interval: 1m
      url: oci://public.ecr.aws/karpenter
    ---
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
      values:
        serviceAccount:
          annotations:
            eks.amazonaws.com/role-arn: ${KARPENTER_IAM_ROLE_ARN}
        settings:
          clusterName: ${CLUSTER_NAME}
          clusterEndpoint: ${CLUSTER_ENDPOINT}
          interruptionQueue: ${INTERRUPTION_QUEUE_NAME}
```

### インストール順序の管理

1. ラベルによる制御
   - クラスター作成時に`fluxcd.io/enabled=true`ラベルを付与
   - FluxCDのインストール完了を確認後、`karpenter.sh/enabled=true`ラベルを付与

2. 依存関係の確認
   - Karpenterインストール前にFluxCDの正常動作を確認
   - HelmRepositoryとHelmReleaseリソースの状態チェック

### 利点

1. 宣言的な管理
   - ClusterResourceSetを使用することで、インストールプロセスを宣言的に管理
   - クラスターのライフサイクルとアドオンのインストールが適切に分離

2. 信頼性の向上
   - Cluster APIのネイティブ機能を使用
   - インストール順序の明確な管理
   - FluxCDによる信頼性の高いHelmリリース管理

3. 状態管理の改善
   - ClusterResourceSetBindingによってインストール状態を追跡可能
   - 失敗時の再試行が自動的に行われる

4. 柔軟性
   - 必要に応じてインストール設定をカスタマイズ可能
   - 複数のクラスターに対して一貫した設定を適用可能

### 移行計画

1. 既存のTerraform Helm Provider設定の削除
   - eks-controlplane-template-v2.yamlからhelm_releaseブロックを削除
   - 関連するIAMロールと権限は維持

2. ClusterResourceSetの導入
   - FluxCD用とKarpenter用のClusterResourceSetとConfigMapの作成
   - 既存クラスターへのラベル付け

3. 検証
   - 新規クラスター作成時のインストールフロー確認
   - 既存クラスターへの適用確認
   - インストール順序の検証

## Risks and Mitigations

### リスク1: インストール順序の制御
- リスク: FluxCDのインストール完了前にKarpenterのインストールが開始される
- 緩和策: 
  - ラベルベースの段階的インストール
  - FluxCDの状態確認ジョブの実装
  - リトライメカニズムの実装

### リスク2: 変数展開
- リスク: ConfigMap内の変数展開が必要
- 緩和策: 
  - 環境変数の注入メカニズムの実装
  - 変数置換用のイニシャライザーコンテナの使用

### リスク3: バージョン互換性
- リスク: FluxCD、Karpenter、EKSの互換性
- 緩和策:
  - バージョンマトリックスの管理
  - 互換性チェックの実装

## Alternatives Considered

### 1. Kubernetes Job Based Installation
- 否定理由: ClusterResourceSetの方が宣言的で管理が容易

### 2. EKS Addon Based Installation
- 否定理由: 柔軟性が低く、カスタマイズが制限される

## References

- [Cluster API ClusterResourceSet Proposal](https://github.com/kubernetes-sigs/cluster-api/blob/main/docs/proposals/20200220-cluster-resource-set.md)
- [FluxCD Installation](https://fluxcd.io/docs/installation/)
- [Karpenter Installation Guide](https://karpenter.sh/docs/getting-started/installing/)
- [Cluster API Documentation](https://cluster-api.sigs.k8s.io/)
