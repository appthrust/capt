# Cluster API Control Plane Secrets Design Document

## 概要

このドキュメントは、Cluster APIプロバイダーにおけるControl Plane Secretsの設計仕様を定義します。
この仕様に従うことで、セキュアで管理しやすいControl Planeの実装と、他のCluster APIプロバイダーとの互換性を確保することができます。

## 目次

1. [基本設計](#基本設計)
2. [Secret仕様](#secret仕様)
3. [証明書要件](#証明書要件)
4. [実装ガイドライン](#実装ガイドライン)
5. [セキュリティ考慮事項](#セキュリティ考慮事項)
6. [運用管理](#運用管理)

## 基本設計

### 命名規則

Control Plane関連のSecretは以下の命名規則に従う必要があります：

```
<cluster-name>-control-plane-kubeconfig  # クラスターのkubeconfig
<cluster-name>-etcd                      # etcd関連の証明書
<cluster-name>-ca                        # Kubernetes API Server CA
<cluster-name>-sa                        # Service Account関連
<cluster-name>-proxy                     # プロキシ設定関連
```

### 必須コンポーネント

各Control Planeは以下のSecretを必要とします：

- etcd証明書とキー
- Kubernetes API Server証明書とキー
- Service Account証明書とキー
- クラスターCA証明書

## Secret仕様

### etcd Secret

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: ${CLUSTER_NAME}-etcd
  namespace: ${NAMESPACE}
type: Opaque
data:
  tls.crt: <base64-encoded-cert>    # etcdサーバー証明書
  tls.key: <base64-encoded-key>     # etcdサーバー秘密鍵
  ca.crt: <base64-encoded-ca-cert>  # etcd CA証明書
```

### API Server CA Secret

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: ${CLUSTER_NAME}-ca
  namespace: ${NAMESPACE}
type: Opaque
data:
  tls.crt: <base64-encoded-cert>  # API Server CA証明書
  tls.key: <base64-encoded-key>   # API Server CA秘密鍵
```

### Service Account Secret

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: ${CLUSTER_NAME}-sa
  namespace: ${NAMESPACE}
type: Opaque
data:
  tls.crt: <base64-encoded-cert>  # SA署名用証明書
  tls.key: <base64-encoded-key>   # SA署名用秘密鍵
```

## 証明書要件

### 一般要件

- フォーマット: X.509
- 推奨有効期間: 1年以上
- エンコーディング: PEM形式

### SANs要件

APIサーバー証明書には以下のSANsを含める必要があります：

- DNSNames:
  - kubernetes
  - kubernetes.default
  - kubernetes.default.svc
  - kubernetes.default.svc.cluster.local
  - `<cluster-name>-apiserver`
- IPAddresses:
  - Cluster IP範囲の最初のIP
  - Control PlaneのEndpoint IP
  - Localhost (127.0.0.1)

## 実装ガイドライン

### 証明書生成

```go
// 証明書生成の基本実装
func generateCertificates(cluster *clusterv1.Cluster) (*certificates.CertificateAuthority, error) {
    caConfig := certificates.CAConfig{
        CommonName: fmt.Sprintf("%s-ca", cluster.Name),
        Duration:   time.Hour * 24 * 365, // 1年
    }
    
    return certificates.NewCertificateAuthority(caConfig)
}
```

### Secret作成

```go
// Secret作成の基本実装
func createSecret(ctx context.Context, cluster *clusterv1.Cluster, name string, data map[string][]byte) error {
    secret := &corev1.Secret{
        ObjectMeta: metav1.ObjectMeta{
            Name:      name,
            Namespace: cluster.Namespace,
            OwnerReferences: []metav1.OwnerReference{
                *metav1.NewControllerRef(cluster, clusterv1.GroupVersion.WithKind("Cluster")),
            },
        },
        Type: corev1.SecretTypeOpaque,
        Data: data,
    }
    
    return client.Create(ctx, secret)
}
```

## セキュリティ考慮事項

### アクセス制御

- RBACポリシーの適用
- 最小権限の原則の遵守
- Secretsへのアクセスログの監視

### 暗号化要件

- etcd暗号化の有効化
- TLS 1.2以上の使用
- 強力な暗号スイートの選択

## 運用管理

### 証明書更新

証明書の更新は以下のタイミングで実施します：

1. 有効期限の70%経過時に更新準備開始
2. 有効期限の80%経過時に更新警告
3. 有効期限の90%経過時に強制更新

```go
func shouldRotateCertificates(cert *x509.Certificate) bool {
    lifetime := cert.NotAfter.Sub(cert.NotBefore)
    threshold := cert.NotBefore.Add(lifetime * 7 / 10) // 70%
    return time.Now().After(threshold)
}
```

### モニタリング

以下の項目を監視する必要があります：

- 証明書の有効期限
- Secret作成・更新の成功率
- 証明書のローテーション状態

### トラブルシューティング

一般的な問題の診断手順：

1. 証明書の有効性確認
2. SANsの検証
3. 秘密鍵とペアの確認
4. 権限の確認

```go
func validateCertificateSecret(secret *corev1.Secret) error {
    requiredKeys := []string{"tls.crt", "tls.key"}
    for _, key := range requiredKeys {
        if _, ok := secret.Data[key]; !ok {
            return fmt.Errorf("missing required key: %s", key)
        }
    }
    
    // 証明書の検証
    cert, err := certificates.DecodeCertPEM(secret.Data["tls.crt"])
    if err != nil {
        return fmt.Errorf("invalid certificate: %v", err)
    }
    
    // その他の検証...
    return nil
}
```

## 参考資料

- [Kubernetes Certificates API](https://kubernetes.io/docs/reference/access-authn-authz/certificate-signing-requests/)
- [etcd Security Model](https://etcd.io/docs/latest/op-guide/security/)
- [Cluster API Book](https://cluster-api.sigs.k8s.io/)