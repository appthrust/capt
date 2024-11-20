# ClusterResourceSet PoC

このPoCは、ClusterResourceSetを使用してFluxCDとKarpenterをインストールする方法を検証します。

## 構成

1. `00-test-cluster.yaml`
   - テスト用のクラスター定義
   - 必要なラベル（`fluxcd.io/enabled: "true"`, `karpenter.sh/enabled: "true"`）を含む

2. `01-fluxcd-installer.yaml`
   - FluxCDインストール用のClusterResourceSet
   - source-controllerとhelm-controllerのみをインストール

3. `02-test-secret-and-helmrelease.yaml`
   - テスト用のダミーSecret（WorkspaceTemplateの出力を模倣）
   - Karpenterインストール用のHelmRelease
   - SecretからHelmReleaseへの変数参照のテスト

## テスト手順

1. クラスターの作成
```bash
kubectl apply -f 00-test-cluster.yaml
```

2. FluxCDインストーラーの適用
```bash
kubectl apply -f 01-fluxcd-installer.yaml
```

3. Karpenter設定の適用
```bash
kubectl apply -f 02-test-secret-and-helmrelease.yaml
```

## 確認項目

1. FluxCDのインストール
```bash
# FluxCDのPodが起動していることを確認
kubectl get pods -n flux-system

# HelmRepositoryとHelmReleaseが作成されていることを確認
kubectl get helmrepository -A
kubectl get helmrelease -A
```

2. Secretの転送
```bash
# Secretが転送されていることを確認
kubectl get secret test-eks-connection -n default
```

3. Karpenterのインストール
```bash
# HelmReleaseの状態を確認
kubectl get helmrelease -n karpenter

# 変数が正しく解決されていることを確認
kubectl get helmrelease karpenter -n karpenter -o yaml
```

## 期待される結果

1. FluxCDが正常にインストールされる
2. Secretが正しく転送される
3. HelmReleaseで変数が正しく解決される
4. Karpenterが正常にインストールされる

## クリーンアップ

```bash
# リソースの削除
kubectl delete -f 02-test-secret-and-helmrelease.yaml
kubectl delete -f 01-fluxcd-installer.yaml
kubectl delete -f 00-test-cluster.yaml
