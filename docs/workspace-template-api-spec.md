# Workspace Template API Specification

## Overview

このドキュメントでは、WorkspaceTemplateおよびWorkspaceTemplateApplyのAPI仕様について詳細に説明します。

## API Resources

### 1. WorkspaceTemplate

#### Resource Definition

```yaml
apiVersion: infrastructure.cluster.x-k8s.io/v1beta1
kind: WorkspaceTemplate
```

#### Spec

```go
type WorkspaceTemplateSpec struct {
    // Template defines the workspace template
    Template WorkspaceTemplateDefinition `json:"template"`

    // WriteConnectionSecretToRef specifies the namespace and name of a
    // Secret to which any connection details for this managed resource should
    // be written.
    WriteConnectionSecretToRef *xpv1.SecretReference `json:"writeConnectionSecretToRef,omitempty"`
}
```

#### Template Definition

```go
type WorkspaceTemplateDefinition struct {
    // Metadata contains template-specific metadata
    Metadata *WorkspaceTemplateMetadata `json:"metadata,omitempty"`

    // Spec defines the desired state of the workspace
    Spec tfv1beta1.WorkspaceSpec `json:"spec"`
}
```

#### Metadata

```go
type WorkspaceTemplateMetadata struct {
    // Description provides a human-readable description of the template
    Description string `json:"description,omitempty"`

    // Version specifies the version of this template
    Version string `json:"version,omitempty"`

    // Tags are key-value pairs that can be used to organize and categorize templates
    Tags map[string]string `json:"tags,omitempty"`
}
```

#### Status

```go
type WorkspaceTemplateStatus struct {
    // WorkspaceName is the name of the created Terraform Workspace
    WorkspaceName string `json:"workspaceName,omitempty"`

    // Conditions of the resource
    Conditions []xpv1.Condition `json:"conditions,omitempty"`
}
```

### 2. WorkspaceTemplateApply

#### Resource Definition

```yaml
apiVersion: infrastructure.cluster.x-k8s.io/v1beta1
kind: WorkspaceTemplateApply
```

#### Spec

```go
type WorkspaceTemplateApplySpec struct {
    // TemplateRef references the WorkspaceTemplate to be applied
    TemplateRef WorkspaceTemplateReference `json:"templateRef"`

    // WriteConnectionSecretToRef specifies the namespace and name of a
    // Secret to which any connection details for this managed resource should
    // be written.
    WriteConnectionSecretToRef *xpv1.SecretReference `json:"writeConnectionSecretToRef,omitempty"`

    // Variables are used to override or provide additional variables to the workspace
    Variables map[string]string `json:"variables,omitempty"`

    // WaitForSecret specifies a secret that must exist before creating the workspace
    WaitForSecret *xpv1.SecretReference `json:"waitForSecret,omitempty"`

    // WaitForWorkspaces specifies a list of workspaces that must be ready before creating this workspace
    WaitForWorkspaces []WorkspaceReference `json:"waitForWorkspaces,omitempty"`
}
```

#### References

```go
type WorkspaceTemplateReference struct {
    // Name of the referenced WorkspaceTemplate
    Name string `json:"name"`

    // Namespace of the referenced WorkspaceTemplate
    Namespace string `json:"namespace,omitempty"`
}

type WorkspaceReference struct {
    // Name of the referenced Workspace
    Name string `json:"name"`

    // Namespace of the referenced Workspace
    Namespace string `json:"namespace,omitempty"`
}
```

#### Status

```go
type WorkspaceTemplateApplyStatus struct {
    // WorkspaceName is the name of the created Terraform Workspace
    WorkspaceName string `json:"workspaceName,omitempty"`

    // Applied indicates whether the template has been successfully applied
    Applied bool `json:"applied,omitempty"`

    // LastAppliedTime is the last time this template was applied
    LastAppliedTime *metav1.Time `json:"lastAppliedTime,omitempty"`

    // Conditions of the resource
    Conditions []xpv1.Condition `json:"conditions,omitempty"`
}
```

## フィールドの説明

### WorkspaceTemplate

| フィールド | 説明 | 必須 |
|------------|------|------|
| `spec.template` | ワークスペーステンプレートの定義 | Yes |
| `spec.template.metadata` | テンプレートのメタデータ | No |
| `spec.template.metadata.description` | テンプレートの説明 | No |
| `spec.template.metadata.version` | テンプレートのバージョン | No |
| `spec.template.metadata.tags` | テンプレートの分類用タグ | No |
| `spec.template.spec` | Terraformワークスペースの仕様 | Yes |
| `spec.writeConnectionSecretToRef` | 接続情報を書き込むSecret | No |

### WorkspaceTemplateApply

| フィールド | 説明 | 必須 |
|------------|------|------|
| `spec.templateRef` | 適用するWorkspaceTemplateの参照 | Yes |
| `spec.templateRef.name` | WorkspaceTemplateの名前 | Yes |
| `spec.templateRef.namespace` | WorkspaceTemplateの名前空間 | No |
| `spec.variables` | オーバーライドする変数 | No |
| `spec.writeConnectionSecretToRef` | 接続情報を書き込むSecret | No |
| `spec.waitForSecret` | 待機するSecret | No |
| `spec.waitForWorkspaces` | 待機するWorkspace | No |

## ステータス情報

### WorkspaceTemplate Status

| フィールド | 説明 |
|------------|------|
| `status.workspaceName` | 作成されたTerraformワークスペースの名前 |
| `status.conditions` | リソースの状態を示すCondition |

### WorkspaceTemplateApply Status

| フィールド | 説明 |
|------------|------|
| `status.workspaceName` | 作成されたTerraformワークスペースの名前 |
| `status.applied` | テンプレートが正常に適用されたかどうか |
| `status.lastAppliedTime` | 最後に適用された時刻 |
| `status.conditions` | リソースの状態を示すCondition |

## 使用例

### 基本的なWorkspaceTemplate

```yaml
apiVersion: infrastructure.cluster.x-k8s.io/v1beta1
kind: WorkspaceTemplate
metadata:
  name: vpc-template
spec:
  template:
    metadata:
      description: "AWS VPC Template"
      version: "1.0.0"
      tags:
        type: "network"
        environment: "production"
    spec:
      forProvider:
        source: Inline
        module: |
          module "vpc" {
            source = "terraform-aws-modules/vpc/aws"
            version = "~> 5.0"
            # ... module configuration
          }
```

### 依存関係を持つWorkspaceTemplateApply

```yaml
apiVersion: infrastructure.cluster.x-k8s.io/v1beta1
kind: WorkspaceTemplateApply
metadata:
  name: vpc-apply
spec:
  templateRef:
    name: vpc-template
  variables:
    vpc_name: "production-vpc"
    environment: "prod"
  waitForWorkspaces:
    - name: base-network
      namespace: default
  writeConnectionSecretToRef:
    name: vpc-connection
    namespace: default
