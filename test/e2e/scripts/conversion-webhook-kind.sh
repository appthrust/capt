#!/usr/bin/env bash
set -euo pipefail

# Kind-based E2E for conversion webhook
# - creates a kind cluster
# - installs CRDs and manager + webhook service
# - generates TLS (self-signed) and injects caBundle into CRDs
# - verifies conversions:
#   * v1beta1 create → v1beta2 GET
#   * v1beta2 create → v1beta1 GET

CLUSTER_NAME="capt-conv"
NAMESPACE="capt-system"
IMG="${IMG:-ghcr.io/appthrust/capt:dev-e2e}"
KIND_NODE_IMAGE="${KIND_NODE_IMAGE:-kindest/node:v1.31.0}"
KIND_RETAIN_FLAG="${KIND_RETAIN_FLAG:---retain}"
KIND_WAIT_FLAG="${KIND_WAIT_FLAG:---wait 300s}"

ROOT_DIR="$(cd "$(dirname "$0")/../../.." && pwd)"

log() { echo "[conv-e2e] $*" >&2; }
has_jq() { command -v jq >/dev/null 2>&1; }

kind_up() {
  if ! kind get clusters | grep -q "^${CLUSTER_NAME}$"; then
    log "creating kind cluster ${CLUSTER_NAME} (image=${KIND_NODE_IMAGE})"
    cat <<EOF | kind create cluster --name "${CLUSTER_NAME}" --image "${KIND_NODE_IMAGE}" --config - ${KIND_RETAIN_FLAG} ${KIND_WAIT_FLAG}
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
kubeadmConfigPatches:
- |
  kind: KubeletConfiguration
  apiVersion: kubelet.config.k8s.io/v1beta1
  cgroupDriver: systemd
containerdConfigPatches:
- |-
  [plugins."io.containerd.grpc.v1.cri".containerd.runtimes.runc.options]
    SystemdCgroup = true
EOF
  else
    log "reusing existing kind cluster ${CLUSTER_NAME}"
  fi
  # Ensure kubeconfig context is present/selected
  kind export kubeconfig --name "${CLUSTER_NAME}"
  kubectl config use-context "kind-${CLUSTER_NAME}" >/dev/null 2>&1 || true
  kubectl cluster-info --context "kind-${CLUSTER_NAME}" >/dev/null
}

install_crds() {
  log "generate CRDs"
  ( cd "${ROOT_DIR}" && make clusterapi-manifests )
  log "apply CRDs"
  "${ROOT_DIR}/bin/kustomize" build "${ROOT_DIR}/config/clusterapi" | kubectl apply -f -
  # wait a few key CRDs
  kubectl wait --for=condition=Established crd/captclusters.infrastructure.cluster.x-k8s.io --timeout=120s || true
  kubectl wait --for=condition=Established crd/captcontrolplanes.controlplane.cluster.x-k8s.io --timeout=120s || true
}

build_and_deploy_manager() {
  log "ensure namespace"
  kubectl get ns "${NAMESPACE}" >/dev/null 2>&1 || kubectl create namespace "${NAMESPACE}"

  log "build controller image ${IMG}"
  ( cd "${ROOT_DIR}" && make docker-build IMG="${IMG}" )
  log "load image into kind"
  kind load docker-image "${IMG}" --name "${CLUSTER_NAME}"

  log "apply RBAC"
  ( cd "${ROOT_DIR}/config/rbac" && "${ROOT_DIR}/bin/kustomize" edit set namespace "${NAMESPACE}" )
  "${ROOT_DIR}/bin/kustomize" build "${ROOT_DIR}/config/rbac" | kubectl apply -f -

  log "apply manager manifests"
  ( cd "${ROOT_DIR}/config/manager" && "${ROOT_DIR}/bin/kustomize" edit set image controller="${IMG}" && "${ROOT_DIR}/bin/kustomize" edit set namespace "${NAMESPACE}" )
  "${ROOT_DIR}/bin/kustomize" build "${ROOT_DIR}/config/manager" | kubectl apply -f -

  log "apply webhook service"
  kubectl apply -f "${ROOT_DIR}/config/webhook/service.yaml"
  # Wait for webhook service endpoints to be ready
  log "wait webhook service endpoints"
  for i in $(seq 1 60); do
    eps=$(kubectl -n "${NAMESPACE}" get endpoints webhook-service -o jsonpath='{.subsets[*].addresses[*].ip}' 2>/dev/null || true)
    if [ -n "${eps}" ]; then
      break
    fi
    sleep 2
  done
}

generate_tls_and_patch() {
  local tmp
  tmp="$(mktemp -d)"
  trap '[ -n "${tmp:-}" ] && rm -rf "${tmp}"' RETURN

  log "generate CA and server certs"
  openssl genrsa -out "${tmp}/ca.key" 2048 >/dev/null 2>&1
  openssl req -x509 -new -nodes -key "${tmp}/ca.key" -subj "/CN=capt-webhook-ca" -days 365 -out "${tmp}/ca.crt" >/dev/null 2>&1

  cat >"${tmp}/openssl.cnf" <<EOF
[ req ]
distinguished_name = req_distinguished_name
req_extensions = v3_req
prompt = no
[ req_distinguished_name ]
CN = webhook-service.${NAMESPACE}.svc
[ v3_req ]
subjectAltName = @alt_names
[ alt_names ]
DNS.1 = webhook-service.${NAMESPACE}.svc
DNS.2 = webhook-service.${NAMESPACE}.svc.cluster.local
EOF

  openssl genrsa -out "${tmp}/tls.key" 2048 >/dev/null 2>&1
  openssl req -new -key "${tmp}/tls.key" -out "${tmp}/tls.csr" -config "${tmp}/openssl.cnf" >/dev/null 2>&1
  openssl x509 -req -in "${tmp}/tls.csr" -CA "${tmp}/ca.crt" -CAkey "${tmp}/ca.key" -CAcreateserial -out "${tmp}/tls.crt" -days 365 -extensions v3_req -extfile "${tmp}/openssl.cnf" >/dev/null 2>&1

  log "create webhook-server-cert secret"
  kubectl -n "${NAMESPACE}" create secret tls webhook-server-cert \
    --cert="${tmp}/tls.crt" --key="${tmp}/tls.key" \
    --dry-run=client -o yaml | kubectl apply -f -

  log "mount cert secret into controller-manager"
  kubectl -n "${NAMESPACE}" patch deploy/controller-manager --type='json' -p='[
    {"op":"add","path":"/spec/template/spec/volumes","value":[{"name":"webhook-certs","secret":{"secretName":"webhook-server-cert"}}]},
    {"op":"add","path":"/spec/template/spec/containers/0/volumeMounts","value":[{"name":"webhook-certs","mountPath":"/tmp/k8s-webhook-server/serving-certs","readOnly":true}]}
  ]' || true

  log "inject caBundle into CRDs conversion.webhook"
  local cabundle
  cabundle=$(base64 -w0 "${tmp}/ca.crt")
  for crd in \
    captclusters.infrastructure.cluster.x-k8s.io \
    captclustertemplates.infrastructure.cluster.x-k8s.io \
    captmachines.infrastructure.cluster.x-k8s.io \
    captmachinetemplates.infrastructure.cluster.x-k8s.io \
    captmachinesets.infrastructure.cluster.x-k8s.io \
    captmachinedeployments.infrastructure.cluster.x-k8s.io \
    workspacetemplates.infrastructure.cluster.x-k8s.io \
    workspacetemplateapplies.infrastructure.cluster.x-k8s.io \
    captcontrolplanes.controlplane.cluster.x-k8s.io \
    captcontrolplanetemplates.controlplane.cluster.x-k8s.io; do
    kubectl patch crd "$crd" --type='json' -p="[
      {\"op\":\"add\",\"path\":\"/spec/conversion/webhook/clientConfig/caBundle\",\"value\":\"${cabundle}\"}
    ]" || true
  done
}

assert_conversion_roundtrip() {
  log "create v1beta1 CAPTCluster"
  local tmpf
  tmpf="$(mktemp)"
  trap '[ -n "${tmpf:-}" ] && rm -f "${tmpf}"' RETURN
  cat >"${tmpf}" <<EOF
apiVersion: infrastructure.cluster.x-k8s.io/v1beta1
kind: CAPTCluster
metadata:
  name: conv-sample
  namespace: default
spec:
  region: us-west-2
EOF
  for i in $(seq 1 10); do
    if kubectl apply -f "${tmpf}" >/dev/null 2>&1; then
      break
    fi
    log "retry apply conv-sample ($i)"
    sleep 3
  done

  log "get v1beta2 CAPTCluster via raw endpoint"
  local region
  for i in $(seq 1 20); do
    if has_jq; then
      region=$(kubectl get --raw \
        "/apis/infrastructure.cluster.x-k8s.io/v1beta2/namespaces/default/captclusters/conv-sample" 2>/dev/null | jq -r '.spec.region // empty') || true
    else
      region=$(kubectl get --raw \
        "/apis/infrastructure.cluster.x-k8s.io/v1beta2/namespaces/default/captclusters/conv-sample" 2>/dev/null \
        | sed -n 's/.*"region":"\([^"]*\)".*/\1/p') || true
    fi
    if [ "${region:-}" = "us-west-2" ]; then
      break
    fi
    sleep 3
  done
  if [[ "${region:-}" != "us-west-2" ]]; then
    log "FAIL: expected region=us-west-2, got '${region:-<empty>}'"
    exit 1
  fi
  log "PASS: conversion roundtrip succeeded (region=${region})"
}

# Create v1beta1 resources and GET via v1beta2 to verify conversion works for other CRDs
assert_conversion_roundtrip_more() {
  # CaptMachineSet
  log "create v1beta1 CaptMachineSet"
  cat <<EOF | kubectl apply -f -
apiVersion: infrastructure.cluster.x-k8s.io/v1beta1
kind: CaptMachineSet
metadata:
  name: conv-ms
  namespace: default
spec:
  template:
    spec:
      instanceType: m5.large
      nodeGroupRef:
        name: ng1
        namespace: default
      workspaceTemplateRef:
        name: wt
        namespace: default
EOF
  if has_jq; then
    ms_it=$(kubectl get --raw \
      "/apis/infrastructure.cluster.x-k8s.io/v1beta2/namespaces/default/captmachinesets/conv-ms" | jq -r '.spec.template.spec.instanceType // empty' || true)
  else
    ms_it=$(kubectl get --raw \
      "/apis/infrastructure.cluster.x-k8s.io/v1beta2/namespaces/default/captmachinesets/conv-ms" \
      | sed -n 's/.*"instanceType":"\([^"]*\)".*/\1/p' || true)
  fi
  if [ "${ms_it:-}" != "m5.large" ]; then
    log "FAIL: expected CaptMachineSet.spec.template.spec.instanceType=m5.large, got '${ms_it:-<empty>}'"
    exit 1
  fi

  # CaptMachineDeployment
  log "create v1beta1 CaptMachineDeployment"
  cat <<EOF | kubectl apply -f -
apiVersion: infrastructure.cluster.x-k8s.io/v1beta1
kind: CaptMachineDeployment
metadata:
  name: conv-md
  namespace: default
spec:
  template:
    spec:
      instanceType: t3.medium
      nodeGroupRef:
        name: ng2
        namespace: default
      workspaceTemplateRef:
        name: wt
        namespace: default
EOF
  if has_jq; then
    md_it=$(kubectl get --raw \
      "/apis/infrastructure.cluster.x-k8s.io/v1beta2/namespaces/default/captmachinedeployments/conv-md" | jq -r '.spec.template.spec.instanceType // empty' || true)
  else
    md_it=$(kubectl get --raw \
      "/apis/infrastructure.cluster.x-k8s.io/v1beta2/namespaces/default/captmachinedeployments/conv-md" \
      | sed -n 's/.*"instanceType":"\([^"]*\)".*/\1/p' || true)
  fi
  if [ "${md_it:-}" != "t3.medium" ]; then
    log "FAIL: expected CaptMachineDeployment.spec.template.spec.instanceType=t3.medium, got '${md_it:-<empty>}'"
    exit 1
  fi

  # WorkspaceTemplate
  log "create v1beta1 WorkspaceTemplate"
  cat <<EOF | kubectl apply -f -
apiVersion: infrastructure.cluster.x-k8s.io/v1beta1
kind: WorkspaceTemplate
metadata:
  name: conv-wt
  namespace: default
spec:
  template:
    spec:
      forProvider:
        source: Inline
        module: |
          terraform {}
EOF
  if has_jq; then
    wt_source=$(kubectl get --raw \
      "/apis/infrastructure.cluster.x-k8s.io/v1beta2/namespaces/default/workspacetemplates/conv-wt" | jq -r '.spec.template.spec.forProvider.source // empty' || true)
  else
    wt_source=$(kubectl get --raw \
      "/apis/infrastructure.cluster.x-k8s.io/v1beta2/namespaces/default/workspacetemplates/conv-wt" \
      | sed -n 's/.*"source":"\([^"]*\)".*/\1/p' || true)
  fi
  if [ "${wt_source:-}" != "Inline" ]; then
    log "FAIL: expected WorkspaceTemplate.spec.template.spec.forProvider.source=Inline, got '${wt_source:-<empty>}'"
    exit 1
  fi

  # WorkspaceTemplateApply
  log "create v1beta1 WorkspaceTemplateApply"
  cat <<EOF | kubectl apply -f -
apiVersion: infrastructure.cluster.x-k8s.io/v1beta1
kind: WorkspaceTemplateApply
metadata:
  name: conv-wta
  namespace: default
spec:
  templateRef:
    name: conv-wt
EOF
  if has_jq; then
    wta_tpl=$(kubectl get --raw \
      "/apis/infrastructure.cluster.x-k8s.io/v1beta2/namespaces/default/workspacetemplateapplies/conv-wta" | jq -r '.spec.templateRef.name // empty' || true)
  else
    wta_tpl=$(kubectl get --raw \
      "/apis/infrastructure.cluster.x-k8s.io/v1beta2/namespaces/default/workspacetemplateapplies/conv-wta" \
      | sed -n 's/.*"templateRef":{[^}]*"name":"\([^"]*\)".*/\1/p' || true)
  fi
  if [ "${wta_tpl:-}" != "conv-wt" ]; then
    log "FAIL: expected WorkspaceTemplateApply.spec.templateRef.name=conv-wt, got '${wta_tpl:-<empty>}'"
    exit 1
  fi

  log "PASS: additional resource conversions succeeded"
}

# Create v1beta2 resources and GET via v1beta1 to verify reverse conversion
assert_conversion_roundtrip_reverse() {
  # CAPTCluster
  log "create v1beta2 CAPTCluster (reverse)"
  cat <<EOF | kubectl apply -f -
apiVersion: infrastructure.cluster.x-k8s.io/v1beta2
kind: CAPTCluster
metadata:
  name: conv2-sample
  namespace: default
spec:
  region: eu-central-1
EOF
  local region
  for i in $(seq 1 20); do
    if has_jq; then
      region=$(kubectl get --raw \
        "/apis/infrastructure.cluster.x-k8s.io/v1beta1/namespaces/default/captclusters/conv2-sample" 2>/dev/null | jq -r '.spec.region // empty') || true
    else
      region=$(kubectl get --raw \
        "/apis/infrastructure.cluster.x-k8s.io/v1beta1/namespaces/default/captclusters/conv2-sample" 2>/dev/null \
        | sed -n 's/.*"region":"\([^"]*\)".*/\1/p') || true
    fi
    if [ "${region:-}" = "eu-central-1" ]; then
      break
    fi
    sleep 3
  done
  if [ "${region:-}" != "eu-central-1" ]; then
    log "FAIL: reverse CAPTCluster region expected eu-central-1, got '${region:-<empty>}'"
    exit 1
  fi

  # CaptMachineSet
  log "create v1beta2 CaptMachineSet (reverse)"
  cat <<EOF | kubectl apply -f -
apiVersion: infrastructure.cluster.x-k8s.io/v1beta2
kind: CaptMachineSet
metadata:
  name: conv2-ms
  namespace: default
spec:
  template:
    spec:
      instanceType: c6a.large
      nodeGroupRef:
        name: ng1
        namespace: default
      workspaceTemplateRef:
        name: wt
        namespace: default
EOF
  local ms_it
  if has_jq; then
    ms_it=$(kubectl get --raw \
      "/apis/infrastructure.cluster.x-k8s.io/v1beta1/namespaces/default/captmachinesets/conv2-ms" | jq -r '.spec.template.spec.instanceType // empty' || true)
  else
    ms_it=$(kubectl get --raw \
      "/apis/infrastructure.cluster.x-k8s.io/v1beta1/namespaces/default/captmachinesets/conv2-ms" \
      | sed -n 's/.*"instanceType":"\([^"]*\)".*/\1/p' || true)
  fi
  if [ "${ms_it:-}" != "c6a.large" ]; then
    log "FAIL: reverse CaptMachineSet.spec.template.spec.instanceType=c6a.large, got '${ms_it:-<empty>}'"
    exit 1
  fi

  # CaptMachineDeployment
  log "create v1beta2 CaptMachineDeployment (reverse)"
  cat <<EOF | kubectl apply -f -
apiVersion: infrastructure.cluster.x-k8s.io/v1beta2
kind: CaptMachineDeployment
metadata:
  name: conv2-md
  namespace: default
spec:
  template:
    spec:
      instanceType: t3a.medium
      nodeGroupRef:
        name: ng2
        namespace: default
      workspaceTemplateRef:
        name: wt
        namespace: default
EOF
  local md_it
  if has_jq; then
    md_it=$(kubectl get --raw \
      "/apis/infrastructure.cluster.x-k8s.io/v1beta1/namespaces/default/captmachinedeployments/conv2-md" | jq -r '.spec.template.spec.instanceType // empty' || true)
  else
    md_it=$(kubectl get --raw \
      "/apis/infrastructure.cluster.x-k8s.io/v1beta1/namespaces/default/captmachinedeployments/conv2-md" \
      | sed -n 's/.*"instanceType":"\([^"]*\)".*/\1/p' || true)
  fi
  if [ "${md_it:-}" != "t3a.medium" ]; then
    log "FAIL: reverse CaptMachineDeployment.spec.template.spec.instanceType=t3a.medium, got '${md_it:-<empty>}'"
    exit 1
  fi

  # WorkspaceTemplate
  log "create v1beta2 WorkspaceTemplate (reverse)"
  cat <<EOF | kubectl apply -f -
apiVersion: infrastructure.cluster.x-k8s.io/v1beta2
kind: WorkspaceTemplate
metadata:
  name: conv2-wt
  namespace: default
spec:
  template:
    spec:
      forProvider:
        source: Inline
        module: |
          terraform {}
EOF
  local wt_source
  if has_jq; then
    wt_source=$(kubectl get --raw \
      "/apis/infrastructure.cluster.x-k8s.io/v1beta1/namespaces/default/workspacetemplates/conv2-wt" | jq -r '.spec.template.spec.forProvider.source // empty' || true)
  else
    wt_source=$(kubectl get --raw \
      "/apis/infrastructure.cluster.x-k8s.io/v1beta1/namespaces/default/workspacetemplates/conv2-wt" \
      | sed -n 's/.*"source":"\([^"]*\)".*/\1/p' || true)
  fi
  if [ "${wt_source:-}" != "Inline" ]; then
    log "FAIL: reverse WorkspaceTemplate.spec.template.spec.forProvider.source=Inline, got '${wt_source:-<empty>}'"
    exit 1
  fi

  # WorkspaceTemplateApply
  log "create v1beta2 WorkspaceTemplateApply (reverse)"
  cat <<EOF | kubectl apply -f -
apiVersion: infrastructure.cluster.x-k8s.io/v1beta2
kind: WorkspaceTemplateApply
metadata:
  name: conv2-wta
  namespace: default
spec:
  templateRef:
    name: conv2-wt
EOF
  local wta_tpl
  if has_jq; then
    wta_tpl=$(kubectl get --raw \
      "/apis/infrastructure.cluster.x-k8s.io/v1beta1/namespaces/default/workspacetemplateapplies/conv2-wta" | jq -r '.spec.templateRef.name // empty' || true)
  else
    wta_tpl=$(kubectl get --raw \
      "/apis/infrastructure.cluster.x-k8s.io/v1beta1/namespaces/default/workspacetemplateapplies/conv2-wta" \
      | sed -n 's/.*"templateRef":{[^}]*"name":"\([^"]*\)".*/\1/p' || true)
  fi
  if [ "${wta_tpl:-}" != "conv2-wt" ]; then
    log "FAIL: reverse WorkspaceTemplateApply.spec.templateRef.name=conv2-wt, got '${wta_tpl:-<empty>}'"
    exit 1
  fi

  log "PASS: reverse conversions succeeded"
}

main() {
  kind_up
  install_crds
  build_and_deploy_manager
  generate_tls_and_patch
  # Trigger new rollout to pick up cert mounts
  kubectl -n "${NAMESPACE}" rollout restart deploy/controller-manager || true
  # Give webhook server a moment to start with mounted certs
  sleep 5
  # Wait for controller-manager after mounting certs
  kubectl -n "${NAMESPACE}" rollout status deploy/controller-manager --timeout=180s
  assert_conversion_roundtrip
  assert_conversion_roundtrip_more
  assert_conversion_roundtrip_reverse
  log "done"
}

main "$@"


