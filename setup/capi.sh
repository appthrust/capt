#!/usr/bin/env bash

set -euo pipefail

############################
# Configurable parameters
############################
# 上書き可: 環境変数で指定 (例: CAPI_VERSION=v1.7.3 SKIP_CERT_MANAGER=true ./setup/capi.sh)
CAPI_VERSION="${CAPI_VERSION:-v1.7.3}"
CERT_MANAGER_VERSION="${CERT_MANAGER_VERSION:-v1.14.4}"
CAPI_NAMESPACE="${CAPI_NAMESPACE:-capi-system}"
SKIP_CERT_MANAGER="${SKIP_CERT_MANAGER:-false}"
CLUSTERCTL_VERSION="${CLUSTERCTL_VERSION:-}"  # 空なら自動選定
CLUSTERCTL_CACHE_DIR="${CLUSTERCTL_CACHE_DIR:-${PWD}/setup/bin}"

info()  { echo "[INFO]  $*"; }
warn()  { echo "[WARN]  $*" >&2; }
error() { echo "[ERROR] $*" >&2; exit 1; }

require_cmd() {
  command -v "$1" >/dev/null 2>&1 || error "'$1' が見つかりません。インストールしてください。"
}

usage() {
  cat <<EOF
使い方: $0 [install|uninstall]

環境変数:
  CAPI_VERSION           (default: ${CAPI_VERSION})  # Cluster API Core のバージョン
  CERT_MANAGER_VERSION   (default: ${CERT_MANAGER_VERSION})
  CAPI_NAMESPACE         (default: ${CAPI_NAMESPACE})
  SKIP_CERT_MANAGER      (default: ${SKIP_CERT_MANAGER})  # true で cert-manager 導入スキップ
  CLUSTERCTL_VERSION     (default: auto)  # 管理クラスタのAPI版に適合する clusterctl を自動選定。明示指定可 (例: v1.6.7)
  CLUSTERCTL_CACHE_DIR   (default: ${CLUSTERCTL_CACHE_DIR})  # clusterctl を保存するローカルディレクトリ

例:
  $0 install                            # cert-manager も自動導入 (未導入時)
  SKIP_CERT_MANAGER=true $0 install     # cert-manager をスキップ
  CAPI_VERSION=v1.7.3 $0 install
  $0 uninstall                          # CAPI Core のみ削除 (cert-manager は削除しない)
EOF
}

wait_for_deploy_ready() {
  local ns="$1"; local name="$2"; local timeout="${3:-5m}"
  info "Deployment ${ns}/${name} のロールアウト完了を待機 (${timeout})"
  kubectl -n "${ns}" rollout status deploy/"${name}" --timeout="${timeout}"
}

# 管理クラスタの API バージョンを検出 (v1beta1|v1beta2|none|unknown) を echo
detect_mgmt_apiversion() {
  if ! kubectl get crd clusters.cluster.x-k8s.io >/dev/null 2>&1; then
    echo "none"
    return 0
  fi
  local versions
  versions=$(kubectl get crd clusters.cluster.x-k8s.io -o jsonpath='{.spec.versions[*].name}' 2>/dev/null || true)
  if echo "${versions}" | grep -qw "v1beta2"; then
    echo "v1beta2"
  elif echo "${versions}" | grep -qw "v1beta1"; then
    echo "v1beta1"
  else
    echo "unknown"
  fi
}

# 使用する clusterctl のバージョンを決定して echo
pick_clusterctl_version() {
  if [[ -n "${CLUSTERCTL_VERSION}" ]]; then
    echo "${CLUSTERCTL_VERSION}"
    return 0
  fi
  local api
  api=$(detect_mgmt_apiversion)
  case "${api}" in
    v1beta1)
      # v1beta1 管理クラスタは clusterctl v1.6 系が適合
      echo "v1.6.7"
      ;;
    v1beta2|none|unknown)
      # v1beta2 または未導入の場合は CAPI_VERSION に合わせる
      echo "${CAPI_VERSION}"
      ;;
  esac
}

# OS/ARCH 正規化
_norm_os() { uname -s | tr '[:upper:]' '[:lower:]'; }
_norm_arch() {
  local m; m=$(uname -m)
  case "${m}" in
    x86_64|amd64) echo "amd64";;
    aarch64|arm64) echo "arm64";;
    *) echo "${m}";;
  esac
}

# clusterctl をローカルに確保し、グローバル変数 CLUSTERCTL_CMD に設定
ensure_clusterctl() {
  local ver os arch url path
  ver=$(pick_clusterctl_version)
  os=$(_norm_os)
  arch=$(_norm_arch)
  mkdir -p "${CLUSTERCTL_CACHE_DIR}"
  path="${CLUSTERCTL_CACHE_DIR}/clusterctl-${ver}-${os}-${arch}"

  if [[ -x "${path}" ]]; then
    CLUSTERCTL_CMD="${path}"
    info "clusterctl (${ver}) を再利用します: ${CLUSTERCTL_CMD}"
    return 0
  fi

  url="https://github.com/kubernetes-sigs/cluster-api/releases/download/${ver}/clusterctl-${os}-${arch}"
  info "clusterctl (${ver}) を取得: ${url}"
  if ! curl -fL -o "${path}" "${url}"; then
    error "clusterctl (${ver}) のダウンロードに失敗しました。URL/ネットワークを確認してください。"
  fi
  chmod +x "${path}"
  CLUSTERCTL_CMD="${path}"
  info "clusterctl (${ver}) 準備完了: ${CLUSTERCTL_CMD}"
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

# ClusterTopology feature gate を CAPI の全コア/プロバイダ(bootstrap/control-plane)デプロイメントに付与
enable_cluster_topology_feature() {
  local ns="${CAPI_NAMESPACE}"
  local deployments=(
    "capi-controller-manager"
    "capi-kubeadm-bootstrap-controller-manager"
    "capi-kubeadm-control-plane-controller-manager"
  )

  for d in "${deployments[@]}"; do
    if ! kubectl -n "${ns}" get deploy "${d}" >/dev/null 2>&1; then
      warn "${ns}/${d} が見つかりません。スキップします。"
      continue
    fi

    # 既に付与済みか確認
    if kubectl -n "${ns}" get deploy "${d}" -o jsonpath='{range .spec.template.spec.containers[*].args[*]}{.}{"\n"}{end}' \
      | grep -q "ClusterTopology=true"; then
      info "${ns}/${d} は既に ClusterTopology=true が設定されています。"
    else
      info "${ns}/${d} に ClusterTopology=true を付与します。"
      # args の末尾に追記 (containers[0] を想定)
      kubectl -n "${ns}" patch deploy "${d}" \
        --type='json' \
        -p='[{"op":"add","path":"/spec/template/spec/containers/0/args/-","value":"--feature-gates=ClusterTopology=true"}]'
      # ロールアウト待機
      wait_for_deploy_ready "${ns}" "${d}" 5m || true
    fi
  done
}

ensure_namespace() {
  local ns="$1"
  if ! kubectl get ns "${ns}" >/dev/null 2>&1; then
    info "Namespace '${ns}' を作成"
    kubectl create ns "${ns}" >/dev/null
  else
    info "Namespace '${ns}' は既に存在します"
  fi
}

is_capi_installed() {
  # 代表的な CRD またはコントローラ Deployment の存在で判定
  if kubectl get crd clusters.cluster.x-k8s.io >/dev/null 2>&1; then
    return 0
  fi
  if kubectl -n "${CAPI_NAMESPACE}" get deploy capi-controller-manager >/dev/null 2>&1; then
    return 0
  fi
  return 1
}

install_cert_manager() {
  if [[ "${SKIP_CERT_MANAGER}" == "true" ]]; then
    warn "SKIP_CERT_MANAGER=true のため cert-manager 導入をスキップします。"
    return 0
  fi

  if kubectl get ns cert-manager >/dev/null 2>&1; then
    info "cert-manager は既に導入済みのためスキップします。"
    return 0
  fi

  local url="https://github.com/cert-manager/cert-manager/releases/download/${CERT_MANAGER_VERSION}/cert-manager.yaml"
  info "cert-manager (${CERT_MANAGER_VERSION}) をインストール: ${url}"
  kubectl apply -f "${url}"

  # webhook/deployment の起動待ち
  wait_for_deploy_ready cert-manager cert-manager-webhook 5m || true
  wait_for_deploy_ready cert-manager cert-manager-cainjector 5m || true
  wait_for_deploy_ready cert-manager cert-manager 5m || true
}

install_capi_core() {
  ensure_namespace "${CAPI_NAMESPACE}"

  if is_capi_installed; then
    info "CAPI Core は既に導入済みのため clusterctl init をスキップします。"
  else
    ensure_clusterctl
    info "${CLUSTERCTL_CMD} で Cluster API Core (${CAPI_VERSION}) をインストール"
    # 既定では bootstrap/control-plane の最新版が選ばれ契約不整合が起きうるため、明示的に同一バージョンを指定
    # (インフラプロバイダは導入しない)
    "${CLUSTERCTL_CMD}" init \
      --core "cluster-api:${CAPI_VERSION}" \
      --bootstrap "kubeadm:${CAPI_VERSION}" \
      --control-plane "kubeadm:${CAPI_VERSION}" \
      --wait-providers
  fi

  # 代表的な CRD が登録されるまで待機
  wait_for_crd clusters.cluster.x-k8s.io 300
  wait_for_crd machines.cluster.x-k8s.io 300
  wait_for_crd machinedeployments.cluster.x-k8s.io 300
  wait_for_crd machinesets.cluster.x-k8s.io 300

  # コントローラの起動待ち
  wait_for_deploy_ready "${CAPI_NAMESPACE}" capi-controller-manager 5m

  # ClusterClass / Topology を利用するための feature gate を有効化
  enable_cluster_topology_feature
}

uninstall_capi_core() {
  if ! is_capi_installed; then
    info "CAPI Core は未導入または既に削除済みのためスキップします。"
    return 0
  fi
  ensure_clusterctl
  info "${CLUSTERCTL_CMD} で Cluster API Core を削除"
  # core + (固定で導入した) kubeadm bootstrap/control-plane を削除
  "${CLUSTERCTL_CMD}" delete --core cluster-api --bootstrap kubeadm --control-plane kubeadm

  warn "cert-manager は他コンポーネントで利用される可能性が高いため削除しません。必要に応じて手動削除してください。"
}

main() {
  local cmd="${1:-install}"
  case "${cmd}" in
    -h|--help|help)
      usage; exit 0 ;;
    install|uninstall)
      ;;
    *)
      error "未知のコマンド: ${cmd}" ;;
  esac

  require_cmd kubectl

  if [[ "${cmd}" == "install" ]]; then
    info "=== CAPI Core v1beta1 をインストールします (namespace=${CAPI_NAMESPACE}) ==="
    install_cert_manager
    install_capi_core
    info "CAPI Core のインストールが完了しました。"
  else
    info "=== CAPI Core をアンインストールします (namespace=${CAPI_NAMESPACE}) ==="
    uninstall_capi_core
    info "CAPI Core のアンインストールが完了しました。"
  fi
}

main "$@"


