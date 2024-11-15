#!/bin/bash

set -euo pipefail

# Function to check if a command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to check if cert-manager is installed
check_cert_manager() {
    echo "Checking cert-manager installation..."
    if ! kubectl get namespace cert-manager >/dev/null 2>&1; then
        echo "cert-manager is not installed. Installing..."
        kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.16.1/cert-manager.yaml
        echo "Waiting for cert-manager to be ready..."
        kubectl wait --for=condition=Ready pod -l app.kubernetes.io/instance=cert-manager -n cert-manager --timeout=120s
    else
        echo "cert-manager is already installed"
    fi
}

# Function to check if clusterctl is installed
check_clusterctl() {
    if ! command_exists clusterctl; then
        echo "clusterctl is not installed. Please install it first."
        echo "Visit: https://cluster-api.sigs.k8s.io/user/quick-start.html#installation"
        exit 1
    fi
}

# Function to install Cluster API
install_cluster_api() {
    echo "Installing Cluster API..."
    
    # Remove any existing Cluster API installation
    if kubectl get namespace capi-system >/dev/null 2>&1; then
        echo "Removing existing Cluster API installation..."
        kubectl delete namespace capi-system capi-kubeadm-bootstrap-system capi-kubeadm-control-plane-system --timeout=60s || true
        echo "Waiting for namespaces to be deleted..."
        sleep 10
    fi

    # Initialize Cluster API
    echo "Initializing Cluster API..."
    clusterctl init
}

# Function to verify installation
verify_installation() {
    echo "Verifying installation..."
    
    # Check if all required namespaces exist
    for ns in capi-system capi-kubeadm-bootstrap-system capi-kubeadm-control-plane-system; do
        if ! kubectl get namespace "$ns" >/dev/null 2>&1; then
            echo "Error: Namespace $ns not found"
            exit 1
        fi
    done

    # Check if core CRDs are installed
    echo "Checking core CRDs..."
    required_crds=(
        "clusters.cluster.x-k8s.io"
        "machinedeployments.cluster.x-k8s.io"
        "machines.cluster.x-k8s.io"
        "machinesets.cluster.x-k8s.io"
    )

    for crd in "${required_crds[@]}"; do
        if ! kubectl get crd "$crd" >/dev/null 2>&1; then
            echo "Error: CRD $crd not found"
            exit 1
        fi
    done

    # Check if controllers are running
    echo "Checking controller deployments..."
    kubectl wait --for=condition=Available deployment -n capi-system capi-controller-manager --timeout=60s
    kubectl wait --for=condition=Available deployment -n capi-kubeadm-bootstrap-system capi-kubeadm-bootstrap-controller-manager --timeout=60s
    kubectl wait --for=condition=Available deployment -n capi-kubeadm-control-plane-system capi-kubeadm-control-plane-controller-manager --timeout=60s

    echo "Cluster API installation verified successfully!"
}

main() {
    echo "Starting Cluster API installation..."
    
    # Check prerequisites
    check_clusterctl
    check_cert_manager
    
    # Install and verify
    install_cluster_api
    verify_installation
    
    echo "Cluster API installation completed successfully!"
}

main "$@"
