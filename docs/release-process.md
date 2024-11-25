# リリースプロセス

このドキュメントでは、CAPTのリリースプロセスについて説明します。

## リリースの種類

1. リリース候補（Release Candidate）
   - タグ形式: `v{major}.{minor}.{patch}-rc{n}` (例: v0.2.0-rc1)
   - テスト用のイメージがビルドされます

2. 正式リリース
   - タグ形式: `v{major}.{minor}.{patch}` (例: v0.2.0)
   - 本番用のマルチアーキテクチャイメージがビルドされます

## リリース手順

### 1. リリース準備

1. CAPTEP作成
   ```bash
   # リリース計画のCAPTEPを作成
   vim docs/CAPTEP/XXXX-release-vX.Y.Z-planning.md
   ```

2. CHANGELOGの更新
   ```bash
   # Unreleasedセクションを新しいバージョンに変更
   vim CHANGELOG.md
   ```

3. バージョン更新
   ```bash
   # VERSIONファイルの更新
   echo "VERSION = X.Y.Z" > VERSION
   
   # kustomization.yamlのイメージタグ更新
   vim config/manager/kustomization.yaml
   ```

4. 変更のコミット
   ```bash
   git add docs/CAPTEP/XXXX-release-vX.Y.Z-planning.md CHANGELOG.md VERSION config/manager/kustomization.yaml
   git commit -m "release: prepare for vX.Y.Z release"
   ```

### 2. リリースタグの作成

1. リリース候補タグの作成（必要な場合）
   ```bash
   git tag -a vX.Y.Z-rc1 -m "Release Candidate 1 for vX.Y.Z"
   git push origin vX.Y.Z-rc1
   ```

2. 正式リリースタグの作成
   ```bash
   git tag -a vX.Y.Z -m "Release vX.Y.Z"
   git push origin vX.Y.Z
   ```

### 3. GitHub Actionsの実行

タグをプッシュすると、以下の処理が自動的に実行されます：

1. Dockerイメージのビルドとプッシュ
   - RC: linux/amd64のみ
   - 正式リリース: linux/amd64, linux/arm64

2. capt.yamlの生成

3. GitHubリリースの作成
   - リリースノート（CHANGELOG.md）の添付
   - capt.yamlの添付

### 4. リリース確認

1. GitHub Actionsの完了を確認
2. GitHubリリースページの確認
   - リリースノートの内容
   - capt.yamlの添付
3. Dockerイメージの確認
   - イメージタグ
   - アーキテクチャ（amd64/arm64）

## トラブルシューティング

### GitHub Actionsとの競合

手動でリリースを作成した場合、GitHub Actionsのリリース作成と競合する可能性があります。
以下のいずれかの方法で対処してください：

1. GitHub Actionsにリリース作成を任せる（推奨）
   - タグをプッシュし、GitHub Actionsの完了を待つ

2. 手動リリースを作成する場合
   - GitHub Actionsのワークフローをキャンセル
   - リリースの作成とアセットのアップロードを手動で行う

## 注意事項

- リリースは必ずメインブランチから行う
- リリース前にすべてのテストが通過していることを確認
- CHANGELOGの更新を忘れずに行う
- バージョン番号は[セマンティックバージョニング](https://semver.org/)に従う
