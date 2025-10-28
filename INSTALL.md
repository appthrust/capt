# CAPT Installation Guide

This guide provides step-by-step instructions for installing CAPT (Cluster API Provider Terraform).

## Prerequisites

Before you begin, ensure you have:

1. kubectl installed and configured
2. AWS credentials properly configured
3. Crossplane with Terraform Provider installed

## Installation Steps

### Step 1: Create Kind Cluster

1. Create a kind cluster:
   ```bash
   kind create cluster --name capt-test
   ```

2. Verify cluster status:
   ```bash
   kubectl cluster-info
   ```

### Step 2: Install cert-manager (required for automatic CA injection)

```
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.16.1/cert-manager.yaml
```
Wait for CRDs to be established:
```
kubectl wait --for=condition=Established crd/certificates.cert-manager.io --timeout=300s || true
```

### Step 3: Install Cluster API (management cluster v1beta2)

1. Install clusterctl (v1.11.x):
   ```bash
   # Match the version used by the Makefile (CLUSTERCTL_VERSION)
   curl -L https://github.com/kubernetes-sigs/cluster-api/releases/download/v1.11.2/clusterctl-linux-amd64 -o clusterctl
   chmod +x clusterctl
   sudo mv clusterctl /usr/local/bin/
   ```

2. Initialize Cluster API (Topology requires kubeadm bootstrap):
   ```bash
   clusterctl init --core cluster-api --bootstrap kubeadm
   ```

3. Verify the installation:
   ```bash
   # Check core components
   kubectl get pods -n capi-system
   kubectl get pods -n capi-kubeadm-bootstrap-system
   kubectl get pods -n capi-kubeadm-control-plane-system

   # Verify CRDs
   kubectl get crds | grep cluster.x-k8s.io
   ```

### Step 4: Install CAPT (with automatic CA injection)

1. Download and apply the installer:
   ```bash
   # Latest stable release
   curl -LO https://github.com/appthrust/capt/releases/latest/download/capt.yaml
   
   # Or specific version
   curl -LO https://github.com/appthrust/capt/releases/download/v0.1.9/capt.yaml

   # Apply the installer
   kubectl apply -f capt.yaml
   ```

3. If you build from source via kustomize (recommended one-shot bootstrap is `make setup`):
   ```bash
   # Option A: one-shot bootstrap for local kind env
   make setup

   # Option B: deploy only CAPT manifests with kustomize
   # (Issuer/Certificate are in config/certmanager; CA injection for Webhooks/CRDs
   #  is enabled via cert-manager annotations in config/default/kustomization.yaml)
   make deploy
   ```

2. Verify the installation:
   ```bash
   # Check controller pod
   kubectl get pods -n capt-system

   # Verify CAPT CRDs
   kubectl get crds | grep infrastructure.cluster.x-k8s.io
   ```

   Expected CRDs:
   - captclusters.infrastructure.cluster.x-k8s.io
   - captcontrolplanes.controlplane.cluster.x-k8s.io
   - captcontrolplanetemplates.controlplane.cluster.x-k8s.io
   - captmachinedeployments.infrastructure.cluster.x-k8s.io
   - captmachines.infrastructure.cluster.x-k8s.io
   - captmachinesets.infrastructure.cluster.x-k8s.io
   - captmachinetemplates.infrastructure.cluster.x-k8s.io
   - workspacetemplateapplies.infrastructure.cluster.x-k8s.io

## Troubleshooting

### Image Pull Errors

If you encounter image pull errors (ErrImagePull or ImagePullBackOff):

1. Check image accessibility:
   ```bash
   # Check pod status
   kubectl get pods -n capt-system
   
   # Check detailed events
   kubectl describe pod -n capt-system <pod-name>
   ```

2. For authentication errors:
   - Ensure the GitHub Container Registry package is set to public
   - Or verify proper authentication credentials are configured

3. If needed, recreate the pod:
   ```bash
   kubectl delete pod -n capt-system -l control-plane=controller-manager
   ```

## Next Steps

After successful installation, you can:

1. Create EKS clusters
2. Manage infrastructure using WorkspaceTemplates
3. Use ClusterClass for standardized cluster deployments

For detailed usage instructions, refer to the main README.md.
