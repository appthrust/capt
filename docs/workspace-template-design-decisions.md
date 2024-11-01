# WorkspaceTemplate Design Decisions

## Overview

このドキュメントでは、WorkspaceTemplateの設計に関する重要な判断と、その背景について説明します。

## Terraform Provider との関係

### 現状の設計

1. WorkspaceTemplateは、Crossplane Terraform Providerの`Workspace`リソースを直接参照
   - `WorkspaceTemplateSpec`は`tfv1beta1.WorkspaceSpec`を使用
   - これにより、Terraform Providerの機能を完全に活用可能

2. メリット
   - Terraform Providerの新機能を自動的に利用可能
   - 型の整合性が保証される
   - コードの重複を避けられる

3. デメリット
   - Terraform Providerの内部実装に依存
   - 一部の設定（例：entrypoint）が不要でも含まれる

## 設計上の課題

### 1. Entrypointの問題

現在、以下の問題が確認されています：

1. 問題の詳細
   - Terraform Providerの`Workspace`には`entrypoint`フィールドが含まれる
   - このフィールドはデフォルト値`""`を持つ
   - WorkspaceTemplateを通じて作成されるWorkspaceにもこの値が設定される

2. 影響
   - `.terraformrc`ファイルに関するエラーが発生
   - Terraformの実行に影響を与える可能性がある

### 2. 考えられる解決策

1. WorkspaceTemplateの型を変更
   ```go
   // 新しい型を定義して不要なフィールドを除外
   type WorkspaceTemplateSpec struct {
       // entrypointを除外した必要なフィールドのみを含む
   }
   ```
   - メリット：型レベルで問題を解決
   - デメリット：Terraform Providerとの整合性が失われる

2. コントローラーでフィールドを制御
   ```go
   // Workspaceを作成時に特定のフィールドを除外
   workspace := &tfv1beta1.Workspace{
       // 必要なフィールドのみをコピー
   }
   ```
   - メリット：既存の型定義を維持
   - デメリット：保守性が低下

## 設計判断

以下の理由により、現状の設計を維持することを決定：

1. Terraform Providerとの整合性
   - 型の一貫性を保つことで、長期的な保守性を確保
   - Providerの更新に追従しやすい構造を維持

2. 拡張性
   - 将来的なTerraform Providerの機能拡張に対応可能
   - カスタマイズが必要な場合は、コントローラーレベルで対応

3. 責任の分離
   - WorkspaceTemplateは、Terraform Workspaceの定義に専念
   - 特定の実装の詳細（entrypointなど）はProviderの責任範囲

## 今後の方向性

1. Terraform Providerとの協業
   - entrypointの問題について、Terraform Providerチームと協議
   - 必要に応じて、Providerレベルでの改善を提案

2. ドキュメントの充実
   - 既知の問題と回避策を明確に文書化
   - ベストプラクティスの提供

3. モニタリングの強化
   - 同様の問題が発生していないか監視
   - ユーザーフィードバックの収集

## 参考資料

- [Terraform Provider Workspace Types](../references/provider-terraform/apis/v1beta1/workspace_types.go)
- [WorkspaceTemplate Types](../api/v1beta1/workspacetemplate_types.go)
- [WorkspaceTemplate Sample](../config/samples/infrastructure_v1beta1_workspacetemplate_eks.yaml)
