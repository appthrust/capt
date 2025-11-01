#!/usr/bin/env bash

set -euo pipefail

############################
# Configurable parameters
############################
CROSSPLANE_HELM_REPO_NAME="crossplane-stable"
CROSSPLANE_HELM_REPO_URL="https://charts.crossplane.io/stable"
CROSSPLANE_CHART_NAME="${CROSSPLANE_HELM_REPO_NAME}/crossplane"
CROSSPLANE_CHART_VERSION="1.18"
CROSSPLANE_RELEASE_NAME="crossplane"
CROSSPLANE_NAMESPACE="upbound-system"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
CROSSPLANE_DIR="${REPO_ROOT}/setup/crossplane"

info()  { echo "[INFO]  $*"; }
warn()  { echo "[WARN]  $*" >&2; }
error() { echo "[ERROR] $*" >&2; exit 1; }

require_cmd() {
  command -v "$1" >/dev/null 2>&1 || error "'$1' が見つかりません。インストールしてください。"
}

wait_for_crd() {
  local crd="$1"; local timeout_seconds="${2:-180}"; local start_ts end_ts
  start_ts=$(date +%s)
  info "CRD '${crd}' の作成を待機しています (timeout: ${timeout_seconds}s)..."
  until kubectl get crd "${crd}" >/dev/null 2>&1; do
    sleep 3
    end_ts=$(date +%s)
    if (( end_ts - start_ts > timeout_seconds )); then
      error "CRD '${crd}' がタイムアウトまでに見つかりませんでした。"
    fi
  done
  info "CRD '${crd}' が確認できました。"
}

apply_manifest() {
  local file="$1"
  info "kubectl apply -f ${file}"
  kubectl apply -f "${file}"
}

main() {
  require_cmd kubectl
  require_cmd helm

  info "Namespace を事前作成 (idempotent)"
  kubectl get ns mgmt >/dev/null 2>&1 || kubectl create ns mgmt >/dev/null
  kubectl get ns "${CROSSPLANE_NAMESPACE}" >/dev/null 2>&1 || kubectl create ns "${CROSSPLANE_NAMESPACE}" >/dev/null

  info "Helm レポジトリ登録/更新 (${CROSSPLANE_HELM_REPO_NAME})"
  if ! helm repo list | grep -q "^${CROSSPLANE_HELM_REPO_NAME}\b"; then
    helm repo add "${CROSSPLANE_HELM_REPO_NAME}" "${CROSSPLANE_HELM_REPO_URL}"
  fi
  helm repo update "${CROSSPLANE_HELM_REPO_NAME}" >/dev/null

  info "Crossplane Chart をインストール/アップグレード (${CROSSPLANE_CHART_NAME} ${CROSSPLANE_CHART_VERSION})"
  helm upgrade --install "${CROSSPLANE_RELEASE_NAME}" "${CROSSPLANE_CHART_NAME}" \
    --namespace "${CROSSPLANE_NAMESPACE}" \
    --create-namespace \
    --version "${CROSSPLANE_CHART_VERSION}" \
    --wait

  info "Crossplane デプロイメントのロールアウト完了を待機"
  kubectl -n "${CROSSPLANE_NAMESPACE}" rollout status deploy/crossplane --timeout=5m
  # rbac-manager はチャート構成により存在しない場合があるため best-effort
  kubectl -n "${CROSSPLANE_NAMESPACE}" rollout status deploy/crossplane-rbac-manager --timeout=5m || \
    warn "deploy/crossplane-rbac-manager は見つかりませんでした (スキップ)"

  info "Crossplane 関連 CRD の準備を待機"
  wait_for_crd deploymentruntimeconfigs.pkg.crossplane.io 300

  info "基礎マニフェストを適用 (Namespace/Secret)"
  apply_manifest "${CROSSPLANE_DIR}/00-upbound-namespace.yaml"
  apply_manifest "${CROSSPLANE_DIR}/01-aws-creds-secret.yaml"

  info "Runtime Config を適用"
  apply_manifest "${CROSSPLANE_DIR}/10-flake-config.yaml"
  apply_manifest "${CROSSPLANE_DIR}/11-aws-cli-runtime-config.yaml"

  info "Provider (provider-terraform) を適用"
  apply_manifest "${CROSSPLANE_DIR}/12-provider-terraform.yaml"

  info "Provider の状態を待機 (Healthy/Installed)"
  # Healthy 条件が無い場合に備えて Installed でも待機
  if ! kubectl wait --for=condition=Healthy --timeout=10m provider.pkg.crossplane.io/provider-terraform 2>/dev/null; then
    kubectl wait --for=condition=Installed --timeout=10m provider.pkg.crossplane.io/provider-terraform
  fi

  info "Terraform Provider の CRD 登場を待機"
  wait_for_crd providerconfigs.tf.upbound.io 300
  wait_for_crd workspaces.tf.upbound.io 300

  info "ProviderConfig を適用"
  apply_manifest "${CROSSPLANE_DIR}/20-aws-provider-config.yaml"

  info "WorkspaceTemplate 群を適用"
  apply_manifest "${CROSSPLANE_DIR}/30-workspacetemplate-vpc.yaml"
  apply_manifest "${CROSSPLANE_DIR}/31-workspacetemplate-eks-controlplane.yaml"
  apply_manifest "${CROSSPLANE_DIR}/32-workspacetemplate-eks-kubeconfig.yaml"
  apply_manifest "${CROSSPLANE_DIR}/40-workspacetemplate-spot-role-check.yaml"
  apply_manifest "${CROSSPLANE_DIR}/41-workspacetemplate-spot-role-create.yaml"

  info "完了しました。"
}

main "$@"


