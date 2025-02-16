# CAPTEP-0047: Helm Chart Management

## Summary

このCAPTEPでは、CAPTコントローラーのHelm Chart提供とその管理の自動化について提案します。主な変更点は、Helm Chartの新規提供、バージョン管理の自動化、および既存のリリースワークフローの活用です。

## Motivation

### 現状の課題

1. デプロイ方法
- CAPTコントローラーのHelm Chartが未提供
- ユーザーが独自にマニフェストを管理する必要がある
- バージョン管理やカスタマイズが困難

2. バージョン管理
- チャートのバージョンとアプリケーションのバージョンの同期が必要
- バージョン更新時のミスのリスク

### Goals

- Helm Chartによるデプロイ方法の提供
- バージョン管理の自動化
- 既存のリリースワークフローの活用
- ユーザーエクスペリエンスの向上

### Non-Goals

- 既存のチャート構造の大幅な変更
- 他のチャート（karpenter）の移動や変更
- リリースプロセス全体の見直し

## Changes

### 1. ディレクトリ構造

```bash
/charts/
  ├── capt/           # コントローラーのチャート（新規追加）
  └── karpenter/      # 既存のkarpenterチャート
```

### 2. バージョン管理の自動化

Makefileに以下の機能を追加：

```makefile
.PHONY: update-version
update-version: ## Update version number
	@echo "Updating version number to $(VERSION)"
	@sed -i 's/^VERSION = .*/VERSION = $(VERSION)/' VERSION
	@yq e -i '.version = "$(VERSION)"' charts/capt/Chart.yaml
	@yq e -i '.appVersion = "v$(VERSION)"' charts/capt/Chart.yaml
```

### 3. Helm Chart生成の改善

```makefile
.PHONY: helm
helm: manifests kustomize helmify ## Generate helm charts
	@echo "Generating Helm chart in charts/capt"
	@mkdir -p charts/capt
	$(KUSTOMIZE) build config/default | $(HELMIFY) charts/capt
```

### 4. リリースワークフローの活用

既存の`release-helm.yml`ワークフローを活用：
- karpenterチャートで実績のあるワークフローを使用
- mainブランチへのプッシュで自動的にチャートをパブリッシュ
- バージョンの重複チェック
- GitHub Container Registryへの保存

## Implementation

1. チャートの生成
```bash
make helm  # charts/captに直接生成
```

2. Makefileの更新
- `update-version`ターゲットの拡張
- `helm`ターゲットの更新
- `helmify`ターゲットのドキュメント改善

3. バージョンの同期
- VERSIONファイル
- Chart.yamlのversion
- Chart.yamlのappVersion

## Migration

新機能の追加であるため、既存のユーザーへの影響は最小限：
- 既存の機能に影響なし
- 新しいデプロイオプションとして提供
- バージョン管理の自動化による信頼性の向上

## Alternatives Considered

### 1. 独立したバージョン管理

- チャートとアプリケーションのバージョンを別々に管理
- メリット：より柔軟なバージョニング
- デメリット：同期の複雑さ、ミスのリスク

### 2. チャートの独立リポジトリ化

- チャートを別リポジトリで管理
- メリット：関心の分離
- デメリット：メンテナンスの複雑さ、同期の困難さ

## References

- [Helm Best Practices](https://helm.sh/docs/chart_best_practices/)
- [Semantic Versioning](https://semver.org/)
- [GitHub Container Registry](https://docs.github.com/ja/packages/working-with-a-github-packages-registry/working-with-the-container-registry)
