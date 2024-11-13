# CAPTEP-0021: kubeconfigの変数展開の問題

## 概要

CAPTControlPlaneが作成するkubeconfigのsecretsで、Regionの引数部分が`${Region}`といったように展開されていない問題が発生しています。

## 背景

### 現状の実装

1. CAPTControlPlaneは、WorkspaceTemplateApplyを作成する際に、以下のように変数を設定しています：
```go
spec.Variables["region"] = controlPlane.Spec.ControlPlaneConfig.Region  // 小文字のregion
```

2. WorkspaceTemplateのvarsセクションでは、以下のように変数が設定されています：
```yaml
vars:
- key: region
  value: "${region}"  # WorkspaceTemplateApply用の変数展開形式
```

3. Terraformモジュール内では、以下のように変数を参照する必要があります：
```hcl
name = "${module.eks.cluster_name}.${var.region}.eksctl.io"  # Terraform変数参照形式
```

### 問題点

1. 変数参照形式の混在
   - WorkspaceTemplateApply: `${変数名}`形式
   - Terraformモジュール: `${var.変数名}`形式

2. この形式の違いにより、kubeconfigの`region`パラメータが正しく展開されていませんでした。

## 解決策

### 1. 変数参照形式の明確な区別

1. WorkspaceTemplateApplyの変数展開用（varsセクション）：
```yaml
vars:
- key: region
  value: "${region}"  # ${変数名}形式を使用
```

2. Terraformモジュール内の変数参照：
```hcl
name = "${module.eks.cluster_name}.${var.region}.eksctl.io"  # ${var.変数名}形式を使用
```

### 2. 命名規則の統一

1. すべての変数名を小文字に統一
2. 複数単語の場合はアンダースコアを使用（例：`cluster_name`）
3. この命名規則をドキュメントに記載

## 実装計画

1. WorkspaceTemplateの修正
   - kubeconfigの出力部分で`${var.region}`形式を使用
   - 他のTerraform変数参照も同様に修正

2. テスト
   - 新しいクラスタを作成し、kubeconfigが正しく生成されることを確認
   - 既存のクラスタに影響がないことを確認

## 代替案

### 1. 大文字小文字を区別しない変数展開

```go
// 変数名の大文字小文字を無視して展開
func replaceVariables(template string, vars map[string]string) string {
    for k, v := range vars {
        template = strings.ReplaceAll(template, "${" + strings.ToUpper(k) + "}", v)
        template = strings.ReplaceAll(template, "${" + strings.ToLower(k) + "}", v)
    }
    return template
}
```

- メリット：既存の変数名との互換性を維持
- デメリット：実装が複雑化、予期しない置換の可能性

### 2. 変数名のバリデーション追加

- WorkspaceTemplateApplyコントローラーで変数名の形式をチェック
- メリット：一貫性の強制
- デメリット：既存のテンプレートの修正が必要

## 実装履歴

- [x] 2024-11-13: 初期提案
- [x] 2024-11-13: 問題の原因特定（変数参照形式の混在）
- [x] 2024-11-13: WorkspaceTemplateの修正（`var.変数名`形式の使用）
- [ ] テスト実施
- [ ] ドキュメント更新

## 参考資料

- [WorkspaceTemplateApply実装](internal/controller/workspacetemplateapply_controller.go)
- [CAPTControlPlane実装](internal/controller/controlplane/workspace.go)
- [変数展開の仕様](docs/workspace-template-design.md)
- [Terraform変数参照](https://www.terraform.io/docs/language/expressions/references.html)
