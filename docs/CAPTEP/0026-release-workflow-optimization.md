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
   
2. 考えられる原因
   - パッケージ名の命名規則による制限
   - リポジトリとパッケージの権限の不一致
   - Organization levelでの追加設定の必要性

#### 解決策
1. Workflow permissionsの適切な設定
   ```yaml
   permissions:
     contents: write
     packages: write
   ```

2. Organization設定での対応
   - パッケージの作成権限を有効化
   - Public/Privateパッケージの許可
   - リポジトリからの権限継承設定

3. パッケージ固有の設定
   - 本番パッケージの明示的な権限設定
   - 命名規則の見直し
   - アクセス制御の再確認

### 2. リリースワークフローの最適化

#### 2段階アプローチ
1. Release Candidate (RC)
   - タグパターン: `v*-rc*`
   - 目的: 初期検証とテスト
   - 成果物: テスト用イメージ

2. 本番リリース
   - タグパターン: `v*`
   - 目的: 安定版のリリース
   - 成果物:
     - マルチアーキテクチャイメージ
     - インストーラー（capt.yaml）
     - GitHubリリース

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
   - 最小権限の原則に従う
   - 明示的な権限設定を優先
   - Organization levelの設定を活用
   - パッケージ固有の権限を確認

2. リリースプロセス
   - セマンティックバージョニングの厳守
   - CHANGELOGの適切な管理
   - リリースノートの充実

3. 自動化
   - エラーハンドリングの強化
   - ログ出力の改善
   - 再試行メカニズムの実装

## 影響

1. 開発プロセス
   - 信頼性の高いリリース
   - 効率的なワークフロー
   - 明確な責任分担

2. メンテナンス
   - シンプルな設定による保守性向上
   - トラブルシューティングの容易化
   - ドキュメントの一元管理

## 参考資料

- [GitHub Actions Workflow Syntax](https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions)
- [GitHub Container Registry Documentation](https://docs.github.com/en/packages/working-with-a-github-packages-registry/working-with-the-container-registry)
- [CAPTEP-0025: CAPT Cluster API Provider Release](./0025-capt-cluster-api-provider-release.md)
