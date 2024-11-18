# CAPTEP-0026: リリースワークフローの最適化

## 概要

このCAPTEPは、CAPTのリリースプロセスの最適化と改善について説明します。特に、GitHub Container Registryへのプッシュ権限の設定、リリースワークフローの効率化、リリースプロセスの自動化に焦点を当てています。

## 動機

リリースプロセスの最適化には以下の目的があります：

1. リリースプロセスの信頼性向上
2. 権限管理の適切な設定
3. ワークフローの効率化
4. 開発者の生産性向上

## 実装の詳細

### 1. GitHub Container Registry権限設定

#### 問題点
- GITHUB_TOKENの権限不足
- パッケージの作成と公開の権限設定
- Organization levelの設定との整合性
- イメージ名による権限の違い

#### 発見された課題
1. イメージ名による権限の違い
   - テストイメージ（capt-test）: プッシュ成功
   - 本番イメージ（capt）: 403 Forbidden
   
2. 詳細な分析結果
   - 同一のワークフロー権限設定:
     ```yaml
     permissions:
       contents: write
       packages: write
     ```
   - 同一のGITHUB_TOKENを使用
   - テストイメージと本番イメージで異なる動作

3. 根本原因の特定
   - リポジトリリンクの欠如が主な原因
   - `org.opencontainers.image.source`ラベルが必要
   - パッケージとリポジトリの明示的な接続が重要

#### 実験と検証
1. 実験1: イメージ名の変更（v0.1.7）
   - アプローチ：イメージ名を`capt-stable`に変更
   - 結果：プッシュ成功
   - 観察：新しいパッケージ名でも動作

2. 実験2: リポジトリリンクの追加（v0.1.9）
   - アプローチ：Dockerfileにラベルを追加
     ```dockerfile
     LABEL org.opencontainers.image.source=https://github.com/appthrust/capt
     LABEL org.opencontainers.image.description="Cluster API Provider for Tofu/Terraform"
     LABEL org.opencontainers.image.licenses=MIT
     ```
   - イメージ名を元の`capt`に戻して検証
   - 結果：プッシュ成功
   ```
   pushing manifest for ghcr.io/appthrust/capt:v0.1.9
   pushing manifest for ghcr.io/appthrust/capt:latest
   DONE 7.4s
   ```

#### 解決策
1. Dockerfileの修正
   ```dockerfile
   LABEL org.opencontainers.image.source=https://github.com/appthrust/capt
   ```
   - このラベルにより、パッケージとリポジトリが正しく接続される
   - GITHUB_TOKENに適切な権限が付与される

2. パッケージとリポジトリの接続
   - リポジトリリンクラベルを使用して明示的に接続
   - 新しいパッケージは自動的にリポジトリにリンク
   - 既存のパッケージも正しく権限を継承

3. イメージ名の戦略
   - 元の`capt`イメージ名を使用可能
   - リポジトリリンクラベルが適切に設定されていれば十分

### 2. リリースワークフローの最適化

#### 2段階アプローチ
1. Release Candidate (RC)
   - タグパターン: `v*-rc*`
   - 目的: 初期検証とテスト
   - 成果物: テスト用イメージ
   - 特徴:
     - シンプルなDockerfile.rc使用
     - 単一アーキテクチャ（linux/amd64）

2. 本番リリース
   - タグパターン: `v*`
   - 目的: 安定版のリリース
   - 成果物:
     - マルチアーキテクチャイメージ（linux/amd64,linux/arm64）
     - インストーラー（capt.yaml）
     - GitHubリリース
   - 特徴:
     - 本番用Dockerfile使用
     - マルチステージビルド
     - マルチアーキテクチャサポート

### 3. リリースプロセスの自動化

#### ワークフロー設定
```yaml
name: Release
on:
  push:
    tags: ['v*']

jobs:
  release:
    permissions:
      contents: write
      packages: write
    steps:
      - name: Build and push Docker image
        with:
          platforms: linux/amd64,linux/arm64
      - name: Generate installer
        run: make build-installer
      - name: Create GitHub Release
        run: gh release create
```

#### 主要コンポーネント
1. マルチアーキテクチャビルド
   - amd64とarm64のサポート
   - buildxを使用した効率的なビルド

2. インストーラー生成
   - CRDとデプロイメント設定の生成
   - バージョン情報の自動挿入

3. リリース作成
   - CHANGELOGの自動添付
   - アセットのアップロード

## 推奨事項

1. 権限管理
   - リポジトリリンクラベルを必ず使用
   - パッケージとリポジトリの接続を確認
   - 明示的な権限設定を優先
   - 定期的な権限監査を実施

2. リリースプロセス
   - セマンティックバージョニングの厳守
   - CHANGELOGの適切な管理
   - リリースノートの充実

3. 自動化
   - エラーハンドリングの強化
   - ログ出力の改善
   - 再試行メカニズムの実装

## 次のステップ

1. イメージ名の統一
   - `capt`イメージ名に統一
   - リポジトリリンクラベルの維持

2. ドキュメントの更新
   - パッケージ権限のベストプラクティスを追加
   - リポジトリリンクの重要性を強調

3. ワークフローの改善
   - エラーメッセージの詳細化
   - 再試行ロジックの実装
   - ログ出力の強化

## 参考資料

- [GitHub Actions Workflow Syntax](https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions)
- [GitHub Container Registry Documentation](https://docs.github.com/en/packages/working-with-a-github-packages-registry/working-with-the-container-registry)
- [CAPTEP-0025: CAPT Cluster API Provider Release](./0025-capt-cluster-api-provider-release.md)
