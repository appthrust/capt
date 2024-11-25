# CAPTEP-0045: Release v0.2.0 Planning

## Summary

CAPTの次期リリース（v0.2.0）の計画と、主要な変更点の整理を行います。このリリースでは、Terraform Providerの移行、FluxCD統合、Karpenterインストールの改善など、重要なアーキテクチャ変更と機能追加が含まれています。

## Motivation

### 現状の課題

1. Terraform Provider
- CrossplaneからUpbound Terraform Providerへの移行が必要
- RBAC権限の更新が必要

2. クラスタ管理機能
- Kubeconfig生成の改善が必要
- WorkspaceStatusによる状態管理の強化が必要

3. アドオン管理
- KarpenterのインストールプロセスをHelmChartProxyベースに移行
- FluxCDとの統合による設定管理の改善

### Goals

- Upbound Terraform Providerへの完全な移行
- Kubeconfig生成メカニズムの改善
- Karpenterインストールの信頼性向上
- FluxCD統合によるアドオン管理の強化
- WorkspaceStatusによる状態管理の改善

### Non-Goals

- 既存のTerraformモジュールの大幅な変更
- 新しいインフラストラクチャプロバイダーの追加
- 既存のクラスター管理機能の削除

## Changes

### Major Changes

1. Terraform Provider Migration (CAPTEP-0033)
- CrossplaneからUpbound Terraform Providerへの移行
- RBAC権限の更新
- Workspace管理の改善

2. Kubeconfig Generation (CAPTEP-0034, 0036, 0037)
- 専用WorkspaceTemplateの導入
- 生成プロセスの改善
- 自動更新メカニズムの実装

3. Karpenter Installation (CAPTEP-0042, 0043)
- HelmChartProxyベースの実装
- インストール信頼性の向上
- 名前空間分離の改善

4. FluxCD Integration (CAPTEP-0031)
- ClusterResourceSetとの統合
- 変数解決メカニズムの実装
- アドオン管理の改善

5. Status Management (CAPTEP-0035, 0040)
- WorkspaceStatusの導入
- 状態追跡の改善
- atProvider詳細の管理

### Breaking Changes

1. Terraform Provider
- tf.crossplane.io APIグループから tf.upbound.io への移行
- RBAC設定の更新が必要

2. Karpenter Installation
- Terraform経由のインストールから HelmChartProxy への移行
- 新しい設定形式の導入

### Migration Guide

1. Terraform Provider
```yaml
# Before
apiVersion: tf.crossplane.io/v1alpha1
kind: Workspace

# After
apiVersion: tf.upbound.io/v1alpha1
kind: Workspace
```

2. Karpenter Installation
```yaml
# Before
spec:
  template:
    spec:
      module:
        source: "terraform-aws-modules/eks/aws"
        values:
          karpenter_enabled: true

# After
spec:
  helmChartProxy:
    name: karpenter
    namespace: karpenter
    chart:
      name: karpenter
      version: "v0.33.0"
```

## Implementation History

- 2024-01-25: Initial proposal
- 2024-01-25: Review started

## References

- [Upbound Terraform Provider](https://docs.upbound.io/providers/terraform/)
- [HelmChartProxy Documentation](https://cluster-api.sigs.k8s.io/tasks/experimental-features/addons/helm-chart-proxy.html)
- [FluxCD Documentation](https://fluxcd.io/docs/)
