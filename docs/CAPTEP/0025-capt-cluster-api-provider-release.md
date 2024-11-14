# CAPTEP-0025: CAPTのCluster API Provider（Infrastructure & Controlplane）としてのリリース

## 概要

このCAPTEPは、CAPTをCluster APIのInfrastructure ProviderとControlplane Providerとしてリリースするための計画と分析を提供します。最近の変更、特にEC2 Spot Service-Linked Roleの管理改善を踏まえて、リリースに向けた準備状況を評価します。

## 動機

CAPTをCluster APIのProviderとしてリリースすることで、以下の利点が得られます：

1. Kubernetes クラスタの管理を標準化された方法で行うことができる
2. AWS EKSクラスタの作成と管理を自動化できる
3. Infrastructure-as-Codeの原則に基づいたクラスタ管理が可能になる
4. コミュニティとの連携や貢献の機会が増える

## 設計の詳細

### カスタムリソース定義（CRD）

1. Infrastructure Provider CRDs
   - CAPTCluster
     - クラスタ全体のインフラストラクチャを管理
     - VPCの設定と管理
     - リージョン設定
     - WorkspaceTemplateとの連携
   
   - CAPTMachine
     - 個々のノードインスタンスを表現
     - EC2インスタンスタイプの設定
     - ノードグループとの関連付け
     - ラベルとタグの管理

   - CAPTMachineDeployment
     - ノードのデプロイメント戦略を管理
     - ローリングアップデートの設定
     - スケーリング設定
     - レプリカ数の管理

   - CAPTMachineSet
     - 同一設定のノードグループを管理
     - デプロイメントの下位リソース
     - レプリカの同期を担当

2. Controlplane Provider CRDs
   - CAPTControlPlane
     - EKSコントロールプレーンの管理
     - Kubernetesバージョン管理
     - エンドポイントアクセス設定
     - アドオン管理

   - CAPTControlPlaneTemplate
     - コントロールプレーンのテンプレート
     - 再利用可能な設定の定義
     - バージョニングとメタデータ管理

3. WorkspaceTemplate関連CRDs
   - WorkspaceTemplate
     - Terraformモジュールの定義
     - 変数とプロバイダー設定
     - メタデータと説明の管理

   - WorkspaceTemplateApply
     - テンプレートの適用を管理
     - 依存関係の制御
     - 状態の追跡

### リソース間の関係性

```
CAPTCluster
  ├── WorkspaceTemplate (VPC)
  └── CAPTControlPlane
       ├── WorkspaceTemplate (EKS)
       └── CAPTMachineDeployment
            └── CAPTMachineSet
                 └── CAPTMachine
```

### RBAC設定

1. ServiceAccount
   - capt-controller-manager
   - システムネームスペースで動作
   - 最小権限の原則に基づく設定

2. ClusterRoles
   - manager-role: コントローラーの主要な操作権限
   - metrics-auth-role: メトリクス認証
   - leader-election-role: リーダー選出

3. RoleBindings
   - システムコンポーネント間の権限バインディング
   - 名前空間スコープの制御

### セキュリティ考慮事項

1. 権限の最小化
   - 必要最小限の権限のみを付与
   - 名前空間による分離
   - リソースごとの詳細なRBAC制御

2. Secrets管理
   - クレデンシャルの安全な保管
   - シークレットの自動ローテーション
   - 暗号化されたコミュニケーション

### コントローラー設定

1. リソース制限
   ```yaml
   resources:
     limits:
       cpu: 500m
       memory: 128Mi
     requests:
       cpu: 10m
       memory: 64Mi
   ```

2. ヘルスチェック
   - Liveness Probe
     - パス: /healthz
     - 初期遅延: 15秒
     - 間隔: 20秒
   
   - Readiness Probe
     - パス: /readyz
     - 初期遅延: 5秒
     - 間隔: 10秒

3. メトリクス設定
   - ポート: 8443
   - 認証付きエンドポイント
   - Prometheusフォーマット

4. セキュリティコンテキスト
   ```yaml
   securityContext:
     allowPrivilegeEscalation: false
     capabilities:
       drop: ["ALL"]
   ```

## リリースプロセス

### コンテナイメージの管理

1. イメージレジストリ
   - GitHub Container Registry (ghcr.io) を使用
   - イメージ名: `ghcr.io/appthrust/capt`
   - タグ形式: `vX.Y.Z` (セマンティックバージョニング)

2. マルチアーキテクチャサポート
   - linux/amd64
   - linux/arm64
   - linux/s390x
   - linux/ppc64le

### リリース手順

1. バージョン番号の更新
   ```bash
   make update-version VERSION=X.Y.Z
   ```

2. CHANGELOGの更新
   ```bash
   make update-changelog VERSION=X.Y.Z
   ```

3. イメージのビルドとプッシュ
   ```bash
   make docker-buildx VERSION=X.Y.Z
   ```

4. インストーラーの生成
   ```bash
   make build-installer VERSION=X.Y.Z
   ```

5. GitHubリリースの作成
   ```bash
   make create-github-release VERSION=X.Y.Z
   ```

## 影響

1. ユーザーへの影響
   - Cluster API準拠のツールとの互換性向上
   - AWS EKSクラスタ管理の自動化と標準化
   - より柔軟なクラスタ設定オプション

2. 開発プロセスへの影響
   - Cluster API仕様に準拠した開発が必要
   - コミュニティとの連携強化
   - CI/CDパイプラインの整備

3. 運用への影響
   - クラスタライフサイクル管理の改善
   - マルチクラスタ環境の統一的な管理
   - モニタリングとアラートの標準化

## 参考資料

- [Cluster API Provider仕様](https://cluster-api.sigs.k8s.io/developer/providers/implementers-guide/overview)
- [AWS EKS ドキュメント](https://docs.aws.amazon.com/eks/latest/userguide/what-is-eks.html)
- [Terraform AWS Provider](https://registry.terraform.io/providers/hashicorp/aws/latest/docs)
- [CAPTEP-0023: EC2 Spot Service-Linked Role Management](./0023-spot-instance-service-linked-role.md)
- [GitHub Container Registry ドキュメント](https://docs.github.com/ja/packages/working-with-a-github-packages-registry/working-with-the-container-registry)
- [GitHub Actions ワークフローの構文](https://docs.github.com/ja/actions/reference/workflow-syntax-for-github-actions)
