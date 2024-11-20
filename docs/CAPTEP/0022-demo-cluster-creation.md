# CAPTEP-0022: デモクラスタの作成プロセス

## 概要

demo-cluster4を参考に、demo-cluster5を作成するプロセスと、その過程で得られた知見をまとめます。

## 背景

### 現状の実装

1. クラスタ構成の基本要素：
   - cluster.yaml: 基本的なクラスタ設定
   - captcluster.yaml: インフラ設定
   - controlplane.yaml: コントロールプレーン設定
   - get-kubeconfig.sh: kubeconfig取得スクリプト

2. 重要な設定パラメータ：
   - リージョン: ap-northeast-1
   - Kubernetesバージョン: 1.31
   - ネットワーク設定:
     - サービスCIDR: 10.96.0.0/12
     - ポッドCIDR: 192.168.0.0/16
   - エンドポイントアクセス: パブリック/プライベート両対応

### 考慮すべき点

1. 変数展開の形式
   - CAPTEP-0021で修正された`${var.変数名}`形式を採用
   - WorkspaceTemplateでの変数参照が正しく設定されていることを確認

2. テンプレート参照
   - VPCテンプレート: vpc-template
   - コントロールプレーンテンプレート: eks-controlplane-template

## 実装手順

1. ディレクトリ構造の作成
```bash
mkdir -p config/samples/demo-cluster5
```

2. 基本マニフェストの作成
   - cluster.yaml: クラスタ基本設定
   - captcluster.yaml: インフラ設定
   - controlplane.yaml: コントロールプレーン設定

3. 補助スクリプトの作成
   - get-kubeconfig.sh: kubeconfig取得用スクリプト
   - 実行権限の付与

## 検証項目

1. マニフェスト設定の確認
   - 各種参照（infrastructureRef, controlPlaneRef）の整合性
   - ネットワーク設定の妥当性
   - バージョン設定の確認

2. 変数展開の確認
   - kubeconfigでの変数参照形式が正しいこと
   - WorkspaceTemplateでの変数展開が機能すること

3. アクセス設定の確認
   - エンドポイントアクセス設定（パブリック/プライベート）
   - セキュリティ設定の妥当性

## 今後の課題

1. テンプレート管理
   - テンプレートのバージョン管理方法の検討
   - テンプレート更新時の互換性確保

2. 設定値の標準化
   - CIDRレンジの標準化
   - タグ付けルールの統一

3. 運用管理
   - クラスタライフサイクル管理の改善
   - モニタリング/ロギング設定の標準化

## 参考資料

- [CAPTEP-0021](docs/CAPTEP/0021-kubeconfig-variable-expansion.md)
- [WorkspaceTemplate設計](docs/workspace-template-design.md)
- [EKSクラスタ設計](docs/eks-cluster-design.md)
