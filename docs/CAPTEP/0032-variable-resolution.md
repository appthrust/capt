# CAPTEP-0032: Variable Resolution for ClusterResourceSet

## Summary
ClusterResourceSetで使用する変数を、WorkspaceTemplateのOutputsとSecretから解決する方法を提案します。

## Motivation
ClusterResourceSetを使用してアドオンをインストールする際、設定値の多くはWorkspaceTemplateによって作成されたリソースから取得する必要があります。これらの値を安全かつ効率的に解決する方法が必要です。

### Goals
- WorkspaceTemplateのOutputsとSecretからの変数解決
- 変数解決プロセスの自動化
- セキュアな変数管理

### Non-Goals
- 一般的な設定管理の方法の定義
- WorkspaceTemplate以外の変数ソースの対応
- 動的な変数解決の実装

## Proposal

### WorkspaceTemplateのOutputs定義

```hcl
# eks-controlplane-template-v2.yaml内のOutputs
output "cluster_endpoint" {
  description = "Endpoint for EKS control plane"
  value       = module.eks.cluster_endpoint
}

output "cluster_name" {
  description = "The name of the EKS cluster"
  value       = module.eks.cluster_name
}

output "karpenter_iam_role_arn" {
  description = "IAM role ARN for Karpenter"
  value       = module.karpenter.iam_role_arn
}

output "karpenter_queue_name" {
  description = "SQS queue name for Karpenter"
  value       = module.karpenter.queue_name
}
```

### ConnectionSecret構造

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: ${WORKSPACE_NAME}-eks-connection
  namespace: default
type: Opaque
stringData:
  cluster_endpoint: ${cluster_endpoint}
  cluster_name: ${cluster_name}
  karpenter_iam_role_arn: ${karpenter_iam_role_arn}
  karpenter_queue_name: ${karpenter_queue_name}
```

### 変数解決の実装

1. CAPTコントローラーでの変数解決
```go
// CAPTコントローラー内での実装
type VariableResolver struct {
    Client client.Client
}

func (r *VariableResolver) ResolveVariables(ctx context.Context, cluster *clusterv1.Cluster, configMap *corev1.ConfigMap) error {
    // WorkspaceTemplateApplyから関連するSecretを取得
    secret := &corev1.Secret{}
    secretName := fmt.Sprintf("%s-eks-connection", workspaceName)
    if err := r.Client.Get(ctx, types.NamespacedName{
        Namespace: cluster.Namespace,
        Name: secretName,
    }, secret); err != nil {
        return fmt.Errorf("failed to get connection secret: %w", err)
    }

    // 変数マッピングの定義
    varMapping := map[string]string{
        "${CLUSTER_NAME}": string(secret.Data["cluster_name"]),
        "${CLUSTER_ENDPOINT}": string(secret.Data["cluster_endpoint"]),
        "${KARPENTER_IAM_ROLE_ARN}": string(secret.Data["karpenter_iam_role_arn"]),
        "${KARPENTER_QUEUE_NAME}": string(secret.Data["karpenter_queue_name"]),
    }

    // ConfigMap内の変数を置換
    for k, v := range varMapping {
        for key, content := range configMap.Data {
            configMap.Data[key] = strings.ReplaceAll(content, k, v)
        }
    }

    return nil
}
```

2. 変数解決のトリガー
```go
func (r *ClusterResourceSetReconciler) reconcileConfigMap(ctx context.Context, cluster *clusterv1.Cluster, configMap *corev1.ConfigMap) error {
    // 変数解決が必要かチェック
    if needsVariableResolution(configMap) {
        resolver := &VariableResolver{Client: r.Client}
        if err := resolver.ResolveVariables(ctx, cluster, configMap); err != nil {
            return err
        }
    }

    // 解決済みConfigMapをクラスターに適用
    return r.applyConfigMap(ctx, cluster, configMap)
}
```

### エラーハンドリング

1. 変数が見つからない場合
```go
func (r *VariableResolver) handleMissingVariable(varName string) error {
    return &MissingVariableError{
        Variable: varName,
        Message: fmt.Sprintf("variable %s not found in connection secret", varName),
    }
}
```

2. 変数解決の再試行
```go
func (r *VariableResolver) resolveWithRetry(ctx context.Context, cluster *clusterv1.Cluster, configMap *corev1.ConfigMap) error {
    return retry.OnError(retry.DefaultRetry, func() error {
        return r.ResolveVariables(ctx, cluster, configMap)
    })
}
```

## Implementation Details

### Phase 1: 基本実装
1. WorkspaceTemplateのOutputs定義の追加
2. ConnectionSecret構造の実装
3. 基本的な変数解決機能の実装

### Phase 2: エラーハンドリング
1. エラー検出と報告の実装
2. リトライメカニズムの追加
3. エラーメッセージの改善

### Phase 3: テストと検証
1. ユニットテストの実装
2. 統合テストの実装
3. エッジケースの検証

## Risks and Mitigations

### リスク1: 変数解決の失敗
- リスク: 必要な変数が見つからない
- 緩和策:
  - デフォルト値の設定
  - 明確なエラーメッセージ
  - リトライメカニズム

### リスク2: パフォーマンス
- リスク: 大量の変数解決による遅延
- 緩和策:
  - キャッシング機能の実装
  - バッチ処理の最適化
  - タイムアウト設定

## References

- [Cluster API Variable Substitution](https://cluster-api.sigs.k8s.io/developer/architecture/controllers/variable-substitution.html)
- [Kubernetes Secrets](https://kubernetes.io/docs/concepts/configuration/secret/)
- [Go Client for Kubernetes](https://github.com/kubernetes/client-go)
