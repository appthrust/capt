# Workspace Template Design Document

## Overview

WorkspaceTemplateは、Terraformのワークスペースを定義し、再利用可能なテンプレートとして管理するための機能です。この機能は、インフラストラクチャのプロビジョニングを効率的に行い、一貫性を保つことを目的としています。

## Architecture

### Core Components

1. WorkspaceTemplate (V2)
   - テンプレートの定義を管理
   - Kubernetes Deploymentライクな構造を採用
   - Workspaceの型を直接活用

2. WorkspaceTemplateApply
   - テンプレートの適用を管理
   - 依存関係の制御
   - 変数のカスタマイズ

### Resource Definitions

#### WorkspaceTemplate

```go
type WorkspaceTemplateSpec struct {
    // Template defines the workspace template
    Template WorkspaceTemplateDefinition `json:"template"`

    // WriteConnectionSecretToRef specifies the namespace and name of a
    // Secret to which any connection details for this managed resource should
    // be written.
    WriteConnectionSecretToRef *xpv1.SecretReference `json:"writeConnectionSecretToRef,omitempty"`
}

type WorkspaceTemplateDefinition struct {
    // Metadata contains template-specific metadata
    Metadata *WorkspaceTemplateMetadata `json:"metadata,omitempty"`

    // Spec defines the desired state of the workspace
    Spec tfv1beta1.WorkspaceParameters `json:"spec"`
}
```

#### WorkspaceTemplateApply

```go
type WorkspaceTemplateApplySpec struct {
    // TemplateRef references the WorkspaceTemplate to be applied
    TemplateRef WorkspaceTemplateReference `json:"templateRef"`

    // Variables for customization
    Variables map[string]string `json:"variables,omitempty"`

    // WriteConnectionSecretToRef specifies the output secret
    WriteConnectionSecretToRef *xpv1.SecretReference `json:"writeConnectionSecretToRef,omitempty"`

    // WaitForWorkspaces defines dependencies on other workspaces
    WaitForWorkspaces []WorkspaceReference `json:"waitForWorkspaces,omitempty"`
}
```

## Features

### 1. Template Management
- HCLモジュールの直接定義
- 変数の動的解決
- プロバイダー設定の柔軟な参照
- 接続情報の出力管理

### 2. Template Application
- テンプレート定義と適用の分離
- 変数のオーバーライド
- 依存関係の管理
- 状態監視

### 3. Dependency Management
- 他のWorkspaceへの依存関係定義
- Ready状態の監視
- 循環依存の防止

## Usage Examples

### 1. Basic Template Definition

```yaml
apiVersion: infrastructure.cluster.x-k8s.io/v1beta1
kind: WorkspaceTemplate
metadata:
  name: vpc-template-sample
spec:
  template:
    metadata:
      description: "Template for creating AWS VPC"
    spec:
      forProvider:
        source: Inline
        module: |
          module "vpc" {
            source = "terraform-aws-modules/vpc/aws"
            version = "~> 5.0"
            name = var.name
            cidr = "10.0.0.0/16"
            # ... additional configuration
          }
```

### 2. Template Application with Dependencies

```yaml
apiVersion: infrastructure.cluster.x-k8s.io/v1beta1
kind: WorkspaceTemplateApply
metadata:
  name: vpc-template-apply-with-deps
spec:
  templateRef:
    name: vpc-template-sample
    namespace: default
  variables:
    name: "production-vpc"
  writeConnectionSecretToRef:
    name: production-vpc-connection
    namespace: default
  waitForWorkspaces:
    - name: base-vpc-workspace
      namespace: default
```

## Best Practices

1. Template Organization
   - 明確な命名規則の使用
   - 適切な説明とメタデータの提供
   - モジュールの再利用性を考慮した設計

2. Variable Management
   - デフォルト値の適切な設定
   - 変数の型と制約の明確な定義
   - 機密情報の適切な取り扱い

3. Dependency Management
   - 明示的な依存関係の定義
   - 循環依存の回避
   - タイムアウトの適切な設定

## Future Improvements

1. Enhanced Variable Support
   - より高度な変数オーバーライド機能
   - 型バリデーションの強化

2. Template Versioning
   - テンプレートのバージョン管理
   - バージョン互換性チェック

3. Status Enhancement
   - より詳細なステータス情報
   - 進捗トラッキング
   - 失敗詳細の改善

4. Validation
   - テンプレートの事前検証
   - 変数バリデーション
   - クロスリソースバリデーション
