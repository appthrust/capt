# Workspace Template Implementation Details

## コントローラーの実装

### 1. WorkspaceTemplate Controller

WorkspaceTemplate Controllerは、WorkspaceTemplateリソースを監視し、テンプレートの検証と管理を行います。

#### 主な責務
- テンプレートの検証
- メタデータの管理
- 状態の追跡

#### 実装のポイント

1. テンプレートの検証
```go
func (r *WorkspaceTemplateReconciler) validateTemplate(template *WorkspaceTemplateDefinition) error {
    // テンプレートの構文チェック
    // 必須フィールドの確認
    // モジュールソースの検証
    // など
}
```

2. メタデータの管理
```go
func (r *WorkspaceTemplateReconciler) updateMetadata(template *WorkspaceTemplate) error {
    // バージョン情報の更新
    // タグの管理
    // 説明の更新
    // など
}
```

3. 状態の追跡
```go
func (r *WorkspaceTemplateReconciler) updateStatus(template *WorkspaceTemplate) error {
    // ステータス条件の更新
    // ワークスペース名の記録
    // など
}
```

### 2. WorkspaceTemplateApply Controller

WorkspaceTemplateApply Controllerは、テンプレートの適用と依存関係の管理を担当します。

#### 主な責務
- テンプレートの適用
- 依存関係の管理
- 変数の解決
- 状態の監視

#### 実装のポイント

1. 依存関係の管理
```go
func (r *WorkspaceTemplateApplyReconciler) waitForDependentWorkspaces(ctx context.Context, cr *v1beta1.WorkspaceTemplateApply) error {
    for _, workspaceRef := range cr.Spec.WaitForWorkspaces {
        workspace := &tfv1beta1.Workspace{}
        namespace := workspaceRef.Namespace
        if namespace == "" {
            namespace = cr.Namespace
        }

        // ワークスペースの存在確認
        err := r.client.Get(ctx, types.NamespacedName{
            Name:      workspaceRef.Name,
            Namespace: namespace,
        }, workspace)
        if err != nil {
            return err
        }

        // Ready状態の確認
        if !isWorkspaceReady(workspace) {
            return fmt.Errorf("workspace %s/%s is not ready", namespace, workspaceRef.Name)
        }
    }
    return nil
}
```

2. 変数の解決
```go
func (r *WorkspaceTemplateApplyReconciler) resolveVariables(ctx context.Context, cr *v1beta1.WorkspaceTemplateApply) (map[string]string, error) {
    // テンプレートのデフォルト変数の取得
    // オーバーライド変数の適用
    // 変数の検証
    // など
}
```

3. ワークスペースの作成
```go
func (r *WorkspaceTemplateApplyReconciler) createWorkspace(ctx context.Context, cr *v1beta1.WorkspaceTemplateApply, template *v1beta1.WorkspaceTemplate, vars map[string]string) error {
    // ワークスペースの作成
    // 変数の適用
    // プロバイダー設定の適用
    // など
}
```

## エラーハンドリング

### 1. 一般的なエラー処理

```go
const (
    errGetTemplate     = "cannot get template"
    errCreateWorkspace = "cannot create workspace"
    errResolveVars    = "cannot resolve variables"
)

func handleError(err error, msg string) error {
    return fmt.Errorf("%s: %w", msg, err)
}
```

### 2. リトライ戦略

```go
func (r *WorkspaceTemplateApplyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    // エラー時のリトライ設定
    if err != nil {
        return ctrl.Result{
            Requeue:      true,
            RequeueAfter: time.Second * 30,
        }, nil
    }
    return ctrl.Result{}, nil
}
```

## テスト戦略

### 1. ユニットテスト

```go
func TestWorkspaceTemplateReconciler_validateTemplate(t *testing.T) {
    tests := []struct {
        name     string
        template *WorkspaceTemplateDefinition
        wantErr  bool
    }{
        // テストケース
    }
    // ...
}
```

### 2. 統合テスト

```go
func TestWorkspaceTemplateApplyReconciler_Integration(t *testing.T) {
    // テスト環境のセットアップ
    // テストケースの実行
    // クリーンアップ
}
```

### 3. E2Eテスト

```go
func TestWorkspaceTemplateE2E(t *testing.T) {
    // クラスターのセットアップ
    // リソースの作成
    // 状態の確認
    // クリーンアップ
}
```

## パフォーマンス最適化

### 1. キャッシュの活用

```go
func (r *WorkspaceTemplateApplyReconciler) getTemplateFromCache(name, namespace string) (*v1beta1.WorkspaceTemplate, error) {
    // キャッシュからテンプレートを取得
    // キャッシュミス時の処理
    // など
}
```

### 2. バッチ処理

```go
func (r *WorkspaceTemplateApplyReconciler) processBatch(ctx context.Context, items []workItem) error {
    // バッチ処理の実装
    // エラーハンドリング
    // など
}
```

## 監視とロギング

### 1. メトリクス

```go
var (
    workspaceCreationDuration = prometheus.NewHistogram(
        prometheus.HistogramOpts{
            Name: "workspace_creation_duration_seconds",
            Help: "Duration of workspace creation in seconds",
        },
    )
)
```

### 2. ロギング

```go
func (r *WorkspaceTemplateApplyReconciler) logReconciliation(ctx context.Context, cr *v1beta1.WorkspaceTemplateApply) {
    logger := log.FromContext(ctx)
    logger.Info("reconciling workspace template apply",
        "name", cr.Name,
        "namespace", cr.Namespace,
        "template", cr.Spec.TemplateRef.Name,
    )
}
```

## セキュリティ考慮事項

### 1. RBAC

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: workspace-template-manager
rules:
  - apiGroups: ["infrastructure.cluster.x-k8s.io"]
    resources: ["workspacetemplates", "workspacetemplateapplies"]
    verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
```

### 2. シークレット管理

```go
func (r *WorkspaceTemplateApplyReconciler) handleSecrets(ctx context.Context, cr *v1beta1.WorkspaceTemplateApply) error {
    // シークレットの取得
    // シークレットの検証
    // シークレットの更新
    // など
}
