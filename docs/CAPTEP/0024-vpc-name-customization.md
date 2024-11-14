# CAPTEP-0024: VPC名のカスタマイズ対応

## 概要

CaptClusterの作成時に生成されるVPCの名前を、クラスター名に基づいてデフォルトで設定し、さらにCaptClusterのspecで指定可能にする機能を追加します。

## 背景

### 現状の実装

1. 現在のVPC名は以下のように設定されています：
   - vpc.yamlテンプレートで`${cluster_name}`変数を使用
   - WorkspaceTemplateApplyコントローラーが変数を置換

2. 問題点：
   - VPC名がクラスター名と連動していない
   - ユーザーがVPC名を指定する方法がない

## 目標

1. デフォルトでVPC名を`{cluster-name}-vpc`形式にする
2. CaptClusterのspecでVPC名を指定可能にする

## 実装詳細

### 1. CaptCluster APIの拡張

```go
// CaptClusterSpec defines the desired state of CaptCluster
type CaptClusterSpec struct {
    // 既存のフィールド...

    // VPCConfig contains VPC-specific configuration
    // +optional
    VPCConfig *VPCConfig `json:"vpcConfig,omitempty"`
}

// VPCConfig contains configuration for the VPC
type VPCConfig struct {
    // Name is the name of the VPC
    // If not specified, defaults to {cluster-name}-vpc
    // +optional
    Name string `json:"name,omitempty"`
}
```

### 2. WorkspaceTemplateApply作成時の変数設定

CaptClusterコントローラーで以下のロジックを実装：

```go
func (r *CaptClusterReconciler) getVPCName(cluster *v1beta1.CaptCluster) string {
    if cluster.Spec.VPCConfig != nil && cluster.Spec.VPCConfig.Name != "" {
        return cluster.Spec.VPCConfig.Name
    }
    return fmt.Sprintf("%s-vpc", cluster.Name)
}

func (r *CaptClusterReconciler) createVPCWorkspaceTemplateApply(ctx context.Context, cluster *v1beta1.CaptCluster) error {
    vpcName := r.getVPCName(cluster)
    
    apply := &v1beta1.WorkspaceTemplateApply{
        // ... 他の設定 ...
        Spec: v1beta1.WorkspaceTemplateApplySpec{
            Variables: map[string]string{
                "cluster_name": vpcName,
            },
        },
    }
    
    return r.Client.Create(ctx, apply)
}
```

### 3. サンプルとドキュメントの更新

1. サンプルマニフェストの追加：

```yaml
apiVersion: infrastructure.cluster.x-k8s.io/v1beta1
kind: CaptCluster
metadata:
  name: custom-vpc-cluster
spec:
  vpcConfig:
    name: my-custom-vpc
```

2. 既存のサンプルファイルの更新：
   - `config/samples/cluster-with-machine/vpc.yaml`
   - `config/samples/cluster.yaml`
   - その他の関連サンプル

## 代替案

### 1. VPC名を完全に自動生成

- メリット：シンプルな実装
- デメリット：柔軟性に欠ける

### 2. WorkspaceTemplateApplyで直接指定

- メリット：既存の仕組みを利用可能
- デメリット：CaptClusterとの統合が弱くなる

## 実装計画

1. APIの更新
   - CaptClusterのCRDに新しいフィールドを追加
   - 生成されたコードの更新

2. コントローラーの更新
   - VPC名の生成ロジックの実装
   - WorkspaceTemplateApply作成時の変数設定

3. テスト
   - デフォルトのVPC名生成のテスト
   - カスタムVPC名指定のテスト
   - 既存のテストの更新

4. ドキュメントとサンプルの更新
   - READMEの更新
   - サンプルマニフェストの更新

## 影響範囲

1. 既存のクラスター
   - 既存のクラスターには影響なし（後方互換性を維持）
   - 新しい機能はオプショナル

2. アップグレード時の考慮事項
   - CRDの更新が必要
   - 既存のWorkspaceTemplateApplyは影響を受けない

## 参考資料

- [Kubernetes API Conventions](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md)
- [Cluster API Provider AWS VPC Management](https://github.com/kubernetes-sigs/cluster-api-provider-aws/blob/main/docs/book/src/topics/vpc-management.md)
