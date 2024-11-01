# WorkspaceTemplate 開発タスク

このドキュメントは、WorkspaceTemplate機能の実装における具体的なタスクを定義します。現在の実装状況を踏まえ、必要な改善と追加機能を段階的に実装していくためのガイドラインとなります。

## Phase 1: テンプレート検証機能の強化

1. **テンプレート構文検証の実装**
   - [ ] `WorkspaceTemplateReconciler`にテンプレート構文検証メソッドを追加
   ```go
   func (r *WorkspaceTemplateReconciler) validateTemplateStructure(template *WorkspaceTemplateDefinition) error
   ```
   - [ ] HCLモジュールの構文チェック機能の実装
   - [ ] 必須フィールドの存在チェック
   - [ ] モジュールソースの有効性検証

2. **変数バリデーションの実装**
   - [ ] `api/v1beta1/workspacetemplate_types.go`に変数スキーマ定義を追加
   ```go
   type VariableSchema struct {
       Type        string   `json:"type"`
       Required    bool     `json:"required,omitempty"`
       Default     string   `json:"default,omitempty"`
       Validation  string   `json:"validation,omitempty"`
   }
   ```
   - [ ] 変数型チェックの実装
   - [ ] デフォルト値の検証ロジック追加
   - [ ] カスタムバリデーションルールのサポート

## Phase 2: 依存関係管理の改善

3. **依存関係グラフの実装**
   - [ ] `internal/controller/workspacetemplateapply_controller.go`に依存関係グラフ構造体を追加
   ```go
   type DependencyGraph struct {
       Nodes map[string]*WorkspaceNode
       Edges map[string][]string
   }
   ```
   - [ ] 循環依存検出ロジックの実装
   - [ ] 依存関係の並列解決機能
   - [ ] 依存関係の状態追跡機能

4. **依存関係の状態管理**
   - [ ] WorkspaceTemplateApplyStatusに詳細な依存関係状態フィールドを追加
   ```go
   type DependencyStatus struct {
       Name      string
       Status    string
       Message   string
       LastCheck metav1.Time
   }
   ```
   - [ ] 依存関係の状態更新ロジック実装
   - [ ] タイムアウト処理の追加
   - [ ] リトライメカニズムの実装

## Phase 3: 変数解決機能の拡張

5. **変数オーバーライドメカニズム**
   - [ ] `internal/controller/workspacetemplateapply_controller.go`に変数解決ロジックを実装
   ```go
   func (r *WorkspaceTemplateApplyReconciler) resolveVariables(ctx context.Context, apply *v1beta1.WorkspaceTemplateApply, template *v1beta1.WorkspaceTemplate) (map[string]string, error)
   ```
   - [ ] 変数の優先順位付けロジック
   - [ ] 環境変数からの値取得サポート
   - [ ] シークレットからの値取得機能

6. **動的変数解決**
   - [ ] 参照による変数解決の実装
   ```go
   type VariableReference struct {
       Kind      string `json:"kind"`
       Name      string `json:"name"`
       Namespace string `json:"namespace"`
       Path      string `json:"path"`
   }
   ```
   - [ ] クロスワークスペース変数参照
   - [ ] 外部システムからの変数取得
   - [ ] 変数解決のキャッシング機能

## Phase 4: ステータス管理とメトリクス

7. **詳細なステータス管理**
   - [ ] WorkspaceTemplateStatusに詳細情報を追加
   ```go
   type DetailedStatus struct {
       Phase           string
       Message         string
       LastTransition  metav1.Time
       Conditions      []metav1.Condition
   }
   ```
   - [ ] フェーズ遷移の実装
   - [ ] 詳細なエラー情報の記録
   - [ ] 進捗状況の追跡

8. **メトリクスとモニタリング**
   - [ ] Prometheusメトリクスの実装
   ```go
   var (
       templateProcessingDuration = prometheus.NewHistogramVec(
           prometheus.HistogramOpts{
               Name: "workspace_template_processing_duration_seconds",
               Help: "Duration of workspace template processing in seconds",
           },
           []string{"template", "phase"},
       )
   )
   ```
   - [ ] カスタムメトリクスの追加
   - [ ] アラート定義の作成
   - [ ] ダッシュボードテンプレートの作成

## Phase 5: エラーハンドリングとリカバリー

9. **エラーハンドリングの強化**
   - [ ] エラー種別の定義
   ```go
   type ErrorCategory string

   const (
       ValidationError   ErrorCategory = "Validation"
       DependencyError  ErrorCategory = "Dependency"
       ResourceError    ErrorCategory = "Resource"
       SystemError      ErrorCategory = "System"
   )
   ```
   - [ ] エラーリカバリー戦略の実装
   - [ ] バックオフリトライの実装
   - [ ] エラーイベントの記録

10. **自動リカバリー機能**
    - [ ] リソースの自動修復ロジック
    - [ ] 状態の自動同期機能
    - [ ] クリーンアップ処理の強化
    - [ ] 手動介入フラグの実装

## Phase 6: テストとドキュメント

11. **テストカバレッジの向上**
    - [ ] ユニットテストの追加
    ```go
    func TestWorkspaceTemplateReconciler_validateTemplate(t *testing.T)
    func TestWorkspaceTemplateApplyReconciler_resolveDependencies(t *testing.T)
    ```
    - [ ] 統合テストシナリオの追加
    - [ ] E2Eテストの実装
    - [ ] パフォーマンステストの追加

12. **ドキュメントの更新**
    - [ ] APIリファレンスの更新
    - [ ] ユースケース例の追加
    - [ ] トラブルシューティングガイドの作成
    - [ ] 運用マニュアルの作成

## 成功基準

各タスクは以下の基準を満たす必要があります：

1. 単体テストカバレッジ80%以上
2. 統合テストが全て成功
3. E2Eテストシナリオが実装され成功
4. コードレビューが完了
5. ドキュメントが更新されている

## 依存関係

- Kubernetes cluster-api
- Terraform Provider
- AWS Provider
- CAPT Core Components

## 注意事項

- 後方互換性を維持すること
- 各フェーズでの変更は段階的に行うこと
- セキュリティ考慮事項を常に確認すること
- パフォーマンスへの影響を考慮すること
