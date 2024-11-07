# EKS CA Certificate Secret Management

## 概要

このドキュメントでは、EKSクラスターにおけるCluster API (CAPI) CA証明書シークレット管理の設計と実装の詳細を説明します。
特に、EKSの制約とCAPIの要件の間のギャップ、およびその解決方法に焦点を当てます。

## 背景

### CAPIの要件

CAPIは、コントロールプレーンのCA証明書シークレットに対して以下の要件を定めています：

```yaml
data:
  # 必須キー
  tls.crt: <base64-encoded-cert>        # API Server CA証明書
  tls.key: <base64-encoded-key>         # API Server CA秘密鍵
  
  # オプショナルだが推奨されるキー
  ca.crt: <base64-encoded-ca-cert>      # CA証明書（tls.crtと同じ）
  ca.key: <base64-encoded-ca-key>       # CA秘密鍵（tls.keyと同じ）
```

### EKSの制約

EKSマネージドクラスターでは：
1. CA証明書（公開鍵）のみが提供される
2. 秘密鍵はAWSによって管理され、アクセス不可
3. 証明書データはWorkspaceTemplateApplyによって生成されるシークレットに格納される

## プロバイダー互換性

### CAPDとの互換性

EKSの制約（秘密鍵の非可用性）は、以下の理由により他のCAPIプロバイダーとの互換性に影響を与えます：

1. CAPDの要件：
   - CAPDは完全なCA証明書と秘密鍵のペアを必要とする
   - これらは新しい証明書の生成や検証に使用される
   - 特にコントロールプレーンコンポーネント間の通信のセキュリティに重要

2. 互換性の制限：
   - EKSクラスターをCAPDのターゲットクラスターとして使用することは困難
   - 秘密鍵がないため、CAPDが必要とする証明書操作が実行できない

3. 影響を受ける操作：
   - 新しい証明書の生成
   - 証明書の更新
   - クライアント証明書の署名
   - コントロールプレーンコンポーネントの認証

### ノードの制限

EKSの証明書管理の制約は、クラスターに参加できるノードの種類にも直接影響を与えます：

1. 参加可能なノード：
   - EKS Managed Node Groups
   - EKS Fargate Pods
   - Self-managed nodes（AWS EC2上でのみ動作）

2. 技術的な制限の理由：
   - 証明書の制限：
     * CA秘密鍵へのアクセス不可により、外部ノード用の証明書が発行できない
     * AWS管理下のノードの認証にのみCA証明書を使用
   - ネットワークの制限：
     * AWS VPC内での動作が前提
     * APIサーバーエンドポイントはAWSのネットワークインフラを通じて提供
     * セキュリティグループやVPCエンドポイントの設定が必要
   - IAM認証：
     * AWS IAMと密接に統合
     * ノードの認証にIAM roleを使用
     * 外部ノードではこの認証メカニズムが使用不可

3. 影響：
   - オンプレミスサーバーをノードとして追加不可
   - 他のクラウドプロバイダーのVMをノードとして追加不可
   - ハイブリッドクラウド構成の制限

### 他のプロバイダーとの互換性

1. マネージドKubernetesサービス：
   - AKS、GKEなども同様の制限を持つ可能性が高い
   - マネージドサービスは一般的に証明書管理を内部で行う

2. セルフマネージドプロバイダー：
   - 完全な証明書管理機能を必要とするプロバイダーとは互換性が制限される
   - 例：オンプレミスプロバイダー、ベアメタルプロバイダー

### 対応策と推奨事項

1. 用途の制限：
   - EKSクラスターは独立したエンドポイントとして使用
   - 他のプロバイダーとの直接的な統合は避ける
   - ノードはAWS環境内に限定

2. アーキテクチャ設計：
   - EKSクラスターを独立したコンポーネントとして扱う
   - 必要な場合は、アプリケーションレベルでの統合を検討
   - マルチクラスター構成の場合は、クラスター間連携に別のメカニズムを使用

3. 代替アプローチ：
   - クラスター間通信が必要な場合は、別の認証メカニズムを使用
   - 例：サービスアカウント、IAMロール、相互TLS（別の証明書を使用）

## 設計上の決定

### 1. シークレット構造

EKSクラスター用のCAシークレットは以下の構造を採用します：

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: {cluster-name}-ca
  namespace: {namespace}
  ownerReferences:
    - apiVersion: controlplane.cluster.x-k8s.io/v1beta1
      kind: CAPTControlPlane
      name: {cluster-name}
      uid: {controller-uid}
data:
  tls.crt: {base64-encoded-cert}  # EKS CA証明書
  ca.crt: {base64-encoded-cert}   # 同じCA証明書（推奨キー）
```

決定理由：
1. 必須の`tls.crt`を提供
2. 推奨される`ca.crt`も提供して可能な限りCAPIの推奨に従う
3. 秘密鍵フィールド（`tls.key`、`ca.key`）は省略
   - EKSでは利用できない
   - 不完全なデータを含めるよりも、明示的に除外する方が安全

### 2. データソース

CA証明書データは以下のソースから取得：

```yaml
WorkspaceTemplateApply Secret:
  name: {cluster-name}-eks-connection
  data:
    cluster_certificate_authority_data: {base64-encoded-cert}
```

このデータを`tls.crt`と`ca.crt`の両方に使用します。

### 3. シークレット管理

1. シークレットの作成タイミング：
   - WorkspaceTemplateApplyシークレットが利用可能になった後
   - CAPTControlPlaneのreconcileSecretsフェーズで

2. 所有権：
   - CAPTControlPlaneリソースを所有者として設定
   - クラスター削除時に自動的にクリーンアップされる

3. 更新戦略：
   - 証明書データの変更を検知
   - 既存のシークレットを更新

## エラー処理

以下のケースに対するエラー処理を実装：

1. WorkspaceTemplateApplyシークレットが見つからない
2. 証明書データが欠落している
3. 証明書データが無効
4. シークレットの作成/更新に失敗

## 監視とロギング

以下のイベントを記録：

1. シークレットの作成/更新
2. 証明書データの取得
3. エラー状態
4. 所有者参照の設定

## 制限事項

1. 秘密鍵の非可用性：
   - EKSの制約により、秘密鍵は提供できない
   - CAPIの完全な要件は満たせない
   - 他のプロバイダーとの互換性が制限される
   - 外部ノードの追加が不可能

2. 証明書の更新：
   - EKSによって管理される
   - 手動での更新は不可能

3. ノードの制限：
   - AWS環境内のノードのみ追加可能
   - ハイブリッドクラウド構成が制限される

## 将来の改善点

1. 証明書の有効期限監視：
   - 証明書の有効期限を監視
   - 期限切れ前に警告を発行

2. 証明書の検証強化：
   - 証明書フォーマットの検証
   - 有効期限の確認
   - 使用目的の確認

3. メトリクス：
   - 証明書の有効期限までの残り時間
   - シークレット作成/更新の成功率
   - エラー発生頻度

4. プロバイダー互換性：
   - 他のプロバイダーとの統合方法の研究
   - 代替認証メカニズムの検討
   - ユースケースに基づいた推奨構成の提供

5. ノード管理：
   - ノード追加の制限に関する明確なドキュメント
   - 代替アプローチの提案
   - マルチクラスター構成のベストプラクティス

## 関連ドキュメント

- [Secret Management](secret-management.md)
- [Endpoint Management](endpoint-management.md)
- [ControlPlane Status Management](controlplane-status-management.md)
