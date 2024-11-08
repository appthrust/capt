# Cluster API Control Plane Secrets Design Document

## Overview

This document defines the design specifications for Control Plane Secrets in Cluster API providers.
Following these specifications ensures secure and manageable Control Plane implementation and compatibility with other Cluster API providers.

## Table of Contents

1. [Basic Design](#basic-design)
2. [Secret Specifications](#secret-specifications)
3. [Certificate Requirements](#certificate-requirements)
4. [Implementation Guidelines](#implementation-guidelines)
5. [Security Considerations](#security-considerations)
6. [Operational Management](#operational-management)

## Basic Design

### Naming Conventions

Control Plane related Secrets must follow these naming conventions:

```
<cluster-name>-control-plane-kubeconfig  # Cluster kubeconfig
<cluster-name>-etcd                      # etcd related certificates
<cluster-name>-ca                        # Kubernetes API Server CA
<cluster-name>-sa                        # Service Account related
<cluster-name>-proxy                     # Proxy configuration related
```

### Required Components

Each Control Plane requires the following Secrets:

- etcd certificates and keys
- Kubernetes API Server certificates and keys
- Service Account certificates and keys
- Cluster CA certificates

## Secret Specifications

### etcd Secret

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: ${CLUSTER_NAME}-etcd
  namespace: ${NAMESPACE}
type: Opaque
data:
  tls.crt: <base64-encoded-cert>    # etcd server certificate
  tls.key: <base64-encoded-key>     # etcd server private key
  ca.crt: <base64-encoded-ca-cert>  # etcd CA certificate
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
  tls.crt: <base64-encoded-cert>  # API Server CA certificate
  tls.key: <base64-encoded-key>   # API Server CA private key
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
  tls.crt: <base64-encoded-cert>  # SA signing certificate
  tls.key: <base64-encoded-key>   # SA signing private key
```

## Certificate Requirements

### General Requirements

- Format: X.509
- Recommended validity period: 1 year or more
- Encoding: PEM format

### SANs Requirements

API server certificates must include the following SANs:

- DNSNames:
  - kubernetes
  - kubernetes.default
  - kubernetes.default.svc
  - kubernetes.default.svc.cluster.local
  - `<cluster-name>-apiserver`
- IPAddresses:
  - First IP of Cluster IP range
  - Control Plane Endpoint IP
  - Localhost (127.0.0.1)

## Implementation Guidelines

### Certificate Generation

```go
// Basic implementation of certificate generation
func generateCertificates(cluster *clusterv1.Cluster) (*certificates.CertificateAuthority, error) {
    caConfig := certificates.CAConfig{
        CommonName: fmt.Sprintf("%s-ca", cluster.Name),
        Duration:   time.Hour * 24 * 365, // 1 year
    }
    
    return certificates.NewCertificateAuthority(caConfig)
}
```

### Secret Creation

```go
// Basic implementation of secret creation
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

## Security Considerations

### Access Control

- Apply RBAC policies
- Follow principle of least privilege
- Monitor access logs to Secrets

### Encryption Requirements

- Enable etcd encryption
- Use TLS 1.2 or higher
- Select strong cipher suites

## Operational Management

### Certificate Renewal

Certificates should be renewed at the following times:

1. Start renewal preparation at 70% of validity period
2. Issue renewal warning at 80% of validity period
3. Force renewal at 90% of validity period

```go
func shouldRotateCertificates(cert *x509.Certificate) bool {
    lifetime := cert.NotAfter.Sub(cert.NotBefore)
    threshold := cert.NotBefore.Add(lifetime * 7 / 10) // 70%
    return time.Now().After(threshold)
}
```

### Monitoring

The following items need to be monitored:

- Certificate expiration
- Secret creation/update success rate
- Certificate rotation status

### Troubleshooting

General diagnostic procedures:

1. Verify certificate validity
2. Validate SANs
3. Check private key pairs
4. Verify permissions

```go
func validateCertificateSecret(secret *corev1.Secret) error {
    requiredKeys := []string{"tls.crt", "tls.key"}
    for _, key := range requiredKeys {
        if _, ok := secret.Data[key]; !ok {
            return fmt.Errorf("missing required key: %s", key)
        }
    }
    
    // Certificate validation
    cert, err := certificates.DecodeCertPEM(secret.Data["tls.crt"])
    if err != nil {
        return fmt.Errorf("invalid certificate: %v", err)
    }
    
    // Other validations...
    return nil
}
```

## References

- [Kubernetes Certificates API](https://kubernetes.io/docs/reference/access-authn-authz/certificate-signing-requests/)
- [etcd Security Model](https://etcd.io/docs/latest/op-guide/security/)
- [Cluster API Book](https://cluster-api.sigs.k8s.io/)
