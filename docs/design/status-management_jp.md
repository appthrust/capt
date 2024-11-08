# ステータス管理の設計

## 概要

このドキュメントでは、Cluster API統合の文脈におけるCaptClusterとCAPTControlPlaneリソースのステータス管理の実装について説明します。この設計は、インフラストラクチャとコントロールプレーンプロバイダーからCluster APIのClusterリソースへの適切なステータス伝播を確保します。

## 要件

### インフラストラクチャプロバイダーの要件

Cluster APIの仕様に従い、インフラストラクチャプロバイダーは以下を提供する必要があります：

1. ステータスフィールド：
   - 必須：
     - `controlPlaneEndpoint` - APIサーバーエンドポイント
     - `ready` - インフラストラクチャの準備状態
   - オプション：
     - `failureReason` - プログラム的なエラートークン
     - `failureMessage` - 人間が読めるエラーメッセージ
     - `failureDomains` - マシン配置に利用可能な障害ドメイン

2. オーナー参照の処理：
   - Clusterコントローラーからのオーナー参照を待機
   - オーナー参照が設定されるまでアクションを取らない

3. コントロールプレーンエンドポイントの処理：
   - 独自のエンドポイントを提供するか、Clusterのエンドポイントを使用
   - エンドポイントが利用できない場合は調整を終了

### コントロールプレーンプロバイダーの要件

コントロールプレーンプロバイダーは以下を行う必要があります：

1. コントロールプレーンエンドポイントの管理：
   - エンドポイント情報の提供または消費
   - クラスター操作のためのエンドポイント可用性の確保

2. ステータスの維持：
   - 準備状態
   - 初期化状態
   - エラー状態

## 実装の詳細

### CaptClusterステータス管理

```go
type CAPTClusterStatus struct {
    // VPC固有のステータス
    VPCWorkspaceName string
    VPCID           string
    
    // Cluster APIの必須フィールド
    Ready           bool
    FailureReason   *string
    FailureMessage  *string
    FailureDomains  clusterv1.FailureDomains
    
    // 詳細な状態追跡のための条件
    Conditions      []metav1.Condition
}
```

主要な実装ポイント：

1. VPCステータス追跡：
   - WorkspaceTemplateApplyを通じてVPC作成を追跡
   - ワークスペースの状態（SyncedとReady）を監視
   - 接続シークレットからVPC IDを抽出

2. エラー処理：
   - 理由とメッセージを含む詳細なエラー状態
   - Clusterリソースへの適切なエラー伝播
   - 終端エラー状態の処理

3. ステータス伝播：
   ```go
   func (r *CAPTClusterReconciler) updateClusterStatus(ctx context.Context, cluster *clusterv1.Cluster, captCluster *infrastructurev1beta1.CAPTCluster) error {
       cluster.Status.InfrastructureReady = captCluster.Status.Ready
       if captCluster.Status.FailureReason != nil {
           reason := capierrors.ClusterStatusError(*captCluster.Status.FailureReason)
           cluster.Status.FailureReason = &reason
       }
       // ...
   }
   ```

### CAPTControlPlaneステータス管理

```go
type CAPTControlPlaneStatus struct {
    Ready                  bool
    Initialized           bool
    Phase                 string
    FailureReason         *string
    FailureMessage        *string
    WorkspaceTemplateStatus *WorkspaceTemplateStatus
    Conditions            []metav1.Condition
}
```

主要な実装ポイント：

1. フェーズ管理：
   - 明確なフェーズ遷移：Creating → Ready/Failed
   - フェーズ固有の条件とメッセージ
   - 各フェーズのタイムアウト処理

2. ワークスペース統合：
   ```go
   type WorkspaceTemplateStatus struct {
       Ready              bool
       State             string
       LastAppliedRevision string
       LastFailedRevision string
       LastFailureMessage string
       Outputs           map[string]string
   }
   ```

3. 条件タイプ：
   ```go
   const (
       ControlPlaneReadyCondition = "Ready"
       ControlPlaneInitializedCondition = "Initialized"
       ControlPlaneFailedCondition = "Failed"
       ControlPlaneCreatingCondition = "Creating"
   )
   ```

## 実装の分析

### ステータス更新フロー

1. インフラストラクチャステータス管理：
   - CaptClusterは`reconcileVPC`メソッドを通じて包括的なステータス更新メカニズムを実装
   - 主要なステージでステータス更新が発生：
     * 初期VPC作成/検証
     * WorkspaceTemplateApplyの監視
     * シークレットからのVPC ID取得
     * 最終的な準備状態の確認

2. ステータス伝播チェーン：
   ```
   WorkspaceTemplateApply Status
          ↓
   CaptCluster Status
          ↓
   Cluster API Cluster Status
   ```

3. 条件管理：
   - 詳細な条件追跡を実装：
     * VPCReadyCondition
     * VPCCreatingCondition
     * VPCFailedCondition
   - 各条件には以下が含まれます：
     * ステータス（True/False）
     * 理由（例：ReasonVPCCreated、ReasonVPCCreating）
     * 詳細なメッセージ
     * 遷移タイムスタンプ

4. エラー処理戦略：
   - 一貫したエラー状態管理のための`setFailedStatus`を実装
   - 複数レベルでエラーを伝播：
     * WorkspaceTemplateApplyエラー
     * VPC設定検証エラー
     * リソース作成/更新エラー
   - FailureReasonとFailureMessageを通じてエラーコンテキストを維持

### ステータス同期

1. インフラストラクチャ準備状態：
   ```go
   // CaptClusterからClusterへのステータス同期
   cluster.Status.InfrastructureReady = captCluster.Status.Ready
   ```

2. コントロールプレーンエンドポイント：
   ```go
   // エンドポイント伝播
   if captCluster.Spec.ControlPlaneEndpoint.Host != "" {
       cluster.Spec.ControlPlaneEndpoint = captCluster.Spec.ControlPlaneEndpoint
   }
   ```

3. 障害ドメイン管理：
   ```go
   // 障害ドメイン伝播
   if len(captCluster.Status.FailureDomains) > 0 {
       cluster.Status.FailureDomains = captCluster.Status.FailureDomains
   }
   ```

## タイムアウト処理

両方のコントローラーが重要な操作のタイムアウト処理を実装：

1. CaptCluster：
   - VPC作成タイムアウト
   - シークレット可用性タイムアウト
   - ワークスペース準備タイムアウト

2. CAPTControlPlane：
   - コントロールプレーン作成タイムアウト
   - VPC準備タイムアウト
   - ワークスペース準備タイムアウト

例：
```go
const (
    controlPlaneTimeout = 30 * time.Minute
    vpcReadyTimeout    = 15 * time.Minute
    secretTimeout      = 5 * time.Minute
)
```

## エラーリカバリー

実装には堅牢なエラーリカバリーメカニズムが含まれています：

1. 一時的なエラー：
   - バックオフを伴う自動再試行
   - 明確なエラー条件
   - リカバリー中のステータス保持

2. 終端エラー：
   - 明確な失敗表示
   - 詳細なエラーメッセージ
   - 自動リカバリーなし（手動介入が必要）

## ステータスの可視性

kubectlを通じた強化されたステータスの可視性：

1. CaptCluster：
   ```bash
   $ kubectl get captcluster
   NAME            VPC-ID          READY   ENDPOINT        AGE
   test-cluster    vpc-123456789   true    10.0.0.1:6443   10m
   ```

2. CAPTControlPlane：
   ```bash
   $ kubectl get captcontrolplane
   NAME            READY   PHASE      VERSION   ENDPOINT        AGE
   test-cluster    true    Ready      1.31      10.0.0.1:6443   10m
   ```

## ベストプラクティス

1. ステータス更新：
   - 競合状態を避けるための原子的更新
   - 更新中の適切なエラー処理
   - 明確なステータス遷移

2. エラー処理：
   - 詳細なエラーメッセージ
   - 適切なエラー分類
   - 明確なリカバリーパス

3. 条件管理：
   - 標準的な条件タイプの使用
   - 遷移タイムスタンプの含有
   - 明確なメッセージの提供

## テストの考慮事項

1. ステータス遷移：
   - すべての可能な状態遷移をテスト
   - タイムアウト処理の検証
   - エラーリカバリーパスの確認

2. 統合テスト：
   - Clusterへのステータス伝播の検証
   - オーナー参照処理のテスト
   - エンドポイント管理の検証

3. エラーシナリオ：
   - 様々なエラー条件のテスト
   - エラーメッセージ伝播の検証
   - リカバリーメカニズムの確認

## 将来の改善点

1. 強化されたステータス：
   - より詳細な進捗情報
   - リソース使用量メトリクス
   - ヘルスインジケーター

2. エラー処理：
   - 特定のエラーに対する自動リカバリー
   - より詳細なエラー分類
   - より良いエラー相関

3. モニタリング：
   - ステータスベースのアラート
   - パフォーマンスメトリクス
   - リソースヘルスモニタリング
