# CAPTEP-0027: リリーステストと検証プロセス

## 概要

このCAPTEPは、CAPTのリリースプロセスにおけるテストと検証手順について説明します。特に、リリースされたアーティファクトを使用したローカル環境でのインストールと動作確認に焦点を当てています。

## 動機

リリースプロセスの検証には以下の目的があります：

1. リリースアーティファクトの品質保証
2. インストール手順の正確性確認
3. ユーザー体験の向上
4. 潜在的な問題の早期発見

## 検証プロセス

### 1. リリース前の確認事項

1. バージョン情報
   - VERSION ファイルの更新
   - CHANGELOGの更新
   - タグの作成

2. ドキュメント
   - インストール手順の正確性
   - 設定例の更新
   - トラブルシューティングガイドの更新

### 2. リリースアーティファクト

1. コンテナイメージ
   - ghcr.io/appthrust/capt:vX.Y.Z
   - マルチアーキテクチャサポート（amd64/arm64）
   - リポジトリリンクラベルの確認
   - パッケージの可視性設定（パブリック）

2. インストーラー
   - capt.yaml
   - CRDとコントローラーの設定を含む
   - RBACルールの正確性

### 3. インストール検証

1. kindクラスタでのテスト
   ```bash
   # kindクラスタの作成
   kind create cluster --name capt-test

   # インストーラーのダウンロード
   curl -LO https://github.com/appthrust/capt/releases/download/vX.Y.Z/capt.yaml

   # インストール
   kubectl apply -f capt.yaml

   # 検証
   kubectl get pods -n capt-system
   kubectl get crds | grep cluster.x-k8s.io
   ```

2. コンポーネントの確認
   - コントローラーPodの状態
   - CRDのインストール
   - RBACの設定

### 4. トラブルシューティング

1. イメージプルエラー
   - エラー: ErrImagePull または ImagePullBackOff
   - 原因: GitHub Container Registry の認証またはパッケージの可視性設定
   - 解決策:
     1. パッケージの可視性をパブリックに設定
     2. Podの再作成
     ```bash
     kubectl delete pod -n capt-system -l control-plane=controller-manager
     ```

2. CRDの確認
   必要なCRDが正しくインストールされていることを確認：
   - captclusters.infrastructure.cluster.x-k8s.io
   - captcontrolplanes.controlplane.cluster.x-k8s.io
   - captcontrolplanetemplates.controlplane.cluster.x-k8s.io
   - captmachinedeployments.infrastructure.cluster.x-k8s.io
   - captmachines.infrastructure.cluster.x-k8s.io
   - captmachinesets.infrastructure.cluster.x-k8s.io
   - captmachinetemplates.infrastructure.cluster.x-k8s.io
   - workspacetemplateapplies.infrastructure.cluster.x-k8s.io

## 推奨事項

1. リリース前チェックリスト
   - バージョン番号の一貫性
   - ドキュメントの更新
   - テスト結果の確認
   - パッケージの可視性設定の確認

2. 検証環境
   - 複数のKubernetesバージョン
   - 異なるクラウドプロバイダー
   - 様々なインストール方法

3. フィードバックループ
   - 問題の報告方法
   - 修正プロセス
   - ドキュメントの改善

## 次のステップ

1. 自動テストの強化
   - E2Eテストの追加
   - 継続的な検証
   - kindクラスタでの自動テスト

2. ドキュメントの改善
   - トラブルシューティングガイドの拡充
   - ユースケースの追加
   - インストール手順の詳細化

3. リリースプロセスの改善
   - パッケージ可視性の自動設定
   - インストール検証の自動化
   - リリースノートの自動生成

## 参考資料

- [CAPTEP-0025: CAPT Cluster API Provider Release](./0025-capt-cluster-api-provider-release.md)
- [CAPTEP-0026: Release Workflow Optimization](./0026-release-workflow-optimization.md)
- [Cluster API Testing Guidelines](https://cluster-api.sigs.k8s.io/developer/testing.html)
