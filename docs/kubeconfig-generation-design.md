# Kubeconfig Generation Design

## 目的
EKSクラスターに接続するためのkubeconfigを生成し、既存の`.kube/config`にマージ可能な形式で出力する。

## 要件
1. `.kube/config`と同じYAML形式であること
2. AWS EKS認証情報が正しく設定されていること
3. eksctlと同じ命名規則に従うこと

## 実装の選択肢と検討

### 1. yamlencode + jsonencode
```hcl
output "kubeconfig" {
  value = yamlencode(jsonencode({
    apiVersion = "v1"
    // ...
  }))
}
```
- 利点:
  - Terraformの型システムを活用可能
  - 構造化されたデータとして扱える
- 欠点:
  - JSONスタイルのクォートが残る
  - `.kube/config`と形式が異なる

### 2. locals + yamlencode
```hcl
locals {
  kubeconfig_json = {
    apiVersion = "v1"
    // ...
  }
}
output "kubeconfig" {
  value = yamlencode(local.kubeconfig_json)
}
```
- 利点:
  - コードの整理が容易
  - 中間データの再利用が可能
- 欠点:
  - JSONスタイルのクォートの問題は解決されない

### 3. Heredoc (採用)
```hcl
output "kubeconfig" {
  value = <<-EOT
apiVersion: v1
clusters:
- cluster:
    // ...
EOT
}
```
- 利点:
  - 純粋なYAML形式を維持
  - クォートの問題がない
  - `.kube/config`と完全に同じ形式
- 欠点:
  - インデントの管理が必要
  - エディタのYAMLサポートが効きにくい

## 最終的な実装方法

1. Heredocを使用して純粋なYAML形式を生成
2. 必要な箇所でTerraformの変数展開を使用
3. eksctlの命名規則に従ったクラスター名とユーザー名の生成
   - クラスター名: `${cluster_name}.${region}.eksctl.io`
   - ユーザー名: `rancher-installer@${cluster_name}.${region}.eksctl.io`

```hcl
output "kubeconfig" {
  description = "Kubeconfig in YAML format"
  sensitive   = true
  value = <<-EOT
apiVersion: v1
clusters:
- cluster:
    certificate-authority-data: ${module.eks.cluster_certificate_authority_data}
    server: ${module.eks.cluster_endpoint}
  name: ${module.eks.cluster_name}.ap-northeast-1.eksctl.io
// ...
EOT
}
```

## 認証設定
AWS EKS認証に必要な設定：
1. aws eks get-tokenコマンドの使用
2. 必要なパラメータの設定
   - cluster-name
   - region
   - output format (json)
3. AWS STSリージョナルエンドポイントの設定

## 注意点
1. kubeconfigには機密情報が含まれるため、`sensitive = true`を設定
2. インデントは4スペースを使用（`.kube/config`の標準）
3. リージョン情報は`data.aws_region.current`から動的に取得
4. 生成されたkubeconfigは既存の`.kube/config`にそのままマージ可能

## 検証方法
1. 生成されたkubeconfigを`.kube/config`にマージ
2. `kubectl get nodes`等のコマンドでクラスターへの接続を確認
3. 認証情報が正しく機能することを確認
