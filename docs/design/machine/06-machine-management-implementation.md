# Machine Management Implementation Guide

## Controller Implementation

### 1. MachineDeploymentコントローラー

```go
// 主要な責務
func (r *CaptMachineDeploymentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    // 1. MachineSetの管理
    // 2. 更新戦略の実装
    // 3. ステータスの更新
}

// 更新戦略の実装例（RollingUpdate）
func (r *CaptMachineDeploymentReconciler) rolloutRolling(ctx context.Context, deployment *v1beta1.CaptMachineDeployment) error {
    // 1. 新しいMachineSetの作成
    // 2. 段階的なスケーリング
    // 3. 古いMachineSetの削除
}
```

### 2. MachineSetコントローラー

```go
// 主要な責務
func (r *CaptMachineSetReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    // 1. レプリカ数の管理
    // 2. Machineの作成/削除
    // 3. ステータスの更新
}

// Machineの管理
func (r *CaptMachineSetReconciler) reconcileMachines(ctx context.Context, machineSet *v1beta1.CaptMachineSet) error {
    // 1. 必要なMachine数の計算
    // 2. Machineの作成/削除
    // 3. ステータスの同期
}
```

### 3. Machineコントローラー

```go
// 主要な責務
func (r *CaptMachineReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    // 1. WorkspaceTemplateApplyの管理
    // 2. NodeGroupとの連携
    // 3. ステータスの更新
}

// WorkspaceTemplateApplyの管理
func (r *CaptMachineReconciler) reconcileWorkspaceTemplateApply(ctx context.Context, machine *v1beta1.CaptMachine) error {
    // 1. WorkspaceTemplateApplyの作成/更新
    // 2. 変数の設定
    // 3. ステータスの同期
}
```

## WorkspaceTemplate Usage

### 1. Machine用のWorkspaceTemplate

```yaml
apiVersion: infrastructure.cluster.x-k8s.io/v1beta1
kind: WorkspaceTemplate
metadata:
  name: machine-template
spec:
  template:
    spec:
      providerConfigRef:
        name: aws-provider
      forProvider:
        source: Inline
        module: |
          # Machine固有の設定
          variable "instance_type" { type = string }
          variable "node_group" { type = string }
          
          # NodeGroupとの連携
          data "aws_eks_node_group" "target" {
            cluster_name    = var.cluster_name
            node_group_name = var.node_group
          }
          
          # ノードの作成
          resource "aws_instance" "machine" {
            instance_type = var.instance_type
            subnet_id     = data.aws_eks_node_group.target.subnet_ids[0]
            # ... その他の設定
          }
```

### 2. 変数の管理

```go
// Machineコントローラーでの変数設定
apply.Spec.Variables = map[string]string{
    "instance_type": machine.Spec.InstanceType,
    "node_group":    machine.Spec.NodeGroupRef.Name,
    "labels":        fmt.Sprintf("%v", machine.Spec.Labels),
    "tags":          fmt.Sprintf("%v", machine.Spec.Tags),
}
```

## Sample Manifests

### 1. MachineDeployment

```yaml
apiVersion: infrastructure.cluster.x-k8s.io/v1beta1
kind: CaptMachineDeployment
metadata:
  name: worker-deployment
spec:
  replicas: 3
  selector:
    matchLabels:
      role: worker
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxUnavailable: 0
      maxSurge: 1
  template:
    metadata:
      labels:
        role: worker
    spec:
      nodeGroupRef:
        name: managed-ng-1
      workspaceTemplateRef:
        name: machine-template
      instanceType: "t3.medium"
```

### 2. MachineSet

```yaml
apiVersion: infrastructure.cluster.x-k8s.io/v1beta1
kind: CaptMachineSet
metadata:
  name: worker-set
spec:
  replicas: 3
  selector:
    matchLabels:
      role: worker
  template:
    metadata:
      labels:
        role: worker
    spec:
      nodeGroupRef:
        name: managed-ng-1
      workspaceTemplateRef:
        name: machine-template
      instanceType: "t3.medium"
```

## Implementation Best Practices

### 1. エラー処理

```go
// 再試行可能なエラーの処理
if err != nil {
    if isRetryableError(err) {
        return ctrl.Result{RequeueAfter: time.Second * 30}, nil
    }
    return ctrl.Result{}, err
}

// 条件付きの更新
if !reflect.DeepEqual(currentStatus, newStatus) {
    if err := r.Status().Update(ctx, resource); err != nil {
        return ctrl.Result{}, fmt.Errorf("failed to update status: %w", err)
    }
}
```

### 2. ステータス管理

```go
// ステータスの更新
func updateStatus(ctx context.Context, resource *v1beta1.Resource) error {
    // 1. 現在の状態の取得
    // 2. 新しい状態の計算
    // 3. 条件の更新
    // 4. メトリクスの更新
    return nil
}

// 条件の更新
func updateConditions(resource *v1beta1.Resource, condition *metav1.Condition) {
    // 1. 既存の条件の検索
    // 2. 条件の更新
    // 3. タイムスタンプの更新
}
```

### 3. リソース管理

```go
// オーナーシップの設定
if err := controllerutil.SetControllerReference(owner, resource, r.Scheme); err != nil {
    return fmt.Errorf("failed to set controller reference: %w", err)
}

// ファイナライザーの管理
if !controllerutil.ContainsFinalizer(resource, finalizerName) {
    controllerutil.AddFinalizer(resource, finalizerName)
    if err := r.Update(ctx, resource); err != nil {
        return fmt.Errorf("failed to add finalizer: %w", err)
    }
}
```

### 4. イベント記録

```go
// イベントの記録
r.Recorder.Event(resource, corev1.EventTypeNormal, "Created", "Created new resource")
r.Recorder.Eventf(resource, corev1.EventTypeWarning, "Failed", "Failed to create resource: %v", err)
```

## Testing

### 1. ユニットテスト

```go
func TestReconcile(t *testing.T) {
    // 1. テストケースの定義
    // 2. モックの設定
    // 3. コントローラーのテスト
    // 4. 結果の検証
}
```

### 2. 統合テスト

```go
func TestIntegration(t *testing.T) {
    // 1. テスト環境のセットアップ
    // 2. リソースの作成
    // 3. 状態の検証
    // 4. クリーンアップ
}
```

## Monitoring and Debugging

### 1. メトリクス

```go
// メトリクスの定義
var (
    reconcileErrors = prometheus.NewCounter(prometheus.CounterOpts{
        Name: "controller_reconcile_errors_total",
        Help: "Total number of reconciliation errors",
    })
)

// メトリクスの記録
reconcileErrors.Inc()
```

### 2. ロギング

```go
// 構造化ログ
logger := log.FromContext(ctx)
logger.Info("reconciling resource",
    "name", resource.Name,
    "namespace", resource.Namespace,
    "generation", resource.Generation)
```

## Troubleshooting Guide

1. 一般的な問題
- リソースが作成されない
- 更新が進まない
- ステータスが更新されない

2. デバッグ手順
- ログの確認
- イベントの確認
- ステータスの確認
- メトリクスの確認

3. 解決策
- コントローラーの再起動
- リソースの再作成
- 手動での状態リセット
