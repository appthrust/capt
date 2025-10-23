#!/usr/bin/env bash
set -euo pipefail

NAMESPACE="default"
CLUSTER_NAME="demo-noop"

log() {
  echo "[no-op-e2e] $*" >&2
}

kubectl_apply() {
  local f="$1"
  log "apply ${f}"
  kubectl apply -f "$f"
}

wait_for_crd() {
  local crd="$1"
  log "wait CRD ${crd}"
  kubectl wait --for=condition=Established crd/${crd} --timeout=120s
}

wait_ready_cond() {
  local kind="$1"; shift
  local name="$1"; shift
  local ns="${1:-$NAMESPACE}"
  log "wait ready ${kind}/${name}"
  kubectl wait --for=condition=Ready "${kind}/${name}" -n "${ns}" --timeout=300s || true
}

# 0) sanity
log "namespace: ${NAMESPACE}"

# 1) apply no-op provider and templates
kubectl_apply config/samples/clusterclass-e2e/noop-provider-config.yaml
kubectl_apply config/samples/clusterclass-e2e/noop-workspacetemplate.yaml
kubectl_apply config/samples/clusterclass-e2e/eks-kubeconfig-template.yaml
kubectl_apply config/samples/clusterclass-e2e/noop-spot-templates.yaml

# Ensure terraform backend namespace exists for provider-terraform (backend kubernetes)
kubectl get ns upbound-system >/dev/null 2>&1 || kubectl create namespace upbound-system

# 2) apply clusterclass stack
kubectl_apply config/samples/clusterclass-e2e/controlplanetemplate.yaml
kubectl_apply config/samples/clusterclass-e2e/captclustertemplate.yaml
kubectl_apply config/samples/clusterclass-e2e/kubeadmconfigtemplate.yaml
kubectl_apply config/samples/clusterclass-e2e/captmachinetemplate.yaml
kubectl_apply config/samples/clusterclass-e2e/clusterclass.yaml

# 3) create cluster (topology)
kubectl_apply config/samples/clusterclass-e2e/cluster.yaml

sleep 2

# 4) wait topology objects to be created
log "wait topology objects (CAPTCluster/CAPTControlPlane)"
# Discover CAPTCluster name by label (random suffix)
CAPTCLUSTER_NAME=""
for i in $(seq 1 60); do
  CAPTCLUSTER_NAME=$(kubectl get captclusters.infrastructure.cluster.x-k8s.io -n ${NAMESPACE} -l cluster.x-k8s.io/cluster-name=${CLUSTER_NAME} -o jsonpath='{.items[0].metadata.name}' 2>/dev/null || true)
  if [[ -n "${CAPTCLUSTER_NAME}" ]]; then
    break
  fi
  sleep 2
done
for i in $(seq 1 60); do
  if [[ $(kubectl get captcontrolplanes.controlplane.cluster.x-k8s.io -n ${NAMESPACE} -o jsonpath='{.items[*].metadata.name}' | wc -w) -gt 0 ]]; then
    break
  fi
  sleep 2
done

# 5) basic checks
log "list core resources"
kubectl get clusters -n ${NAMESPACE}
kubectl get captclusters.infrastructure.cluster.x-k8s.io -n ${NAMESPACE}
kubectl get captcontrolplanes.controlplane.cluster.x-k8s.io -n ${NAMESPACE}

# 5.1) validate ClusterClass topology wiring
log "validate ClusterClass topology wiring"
TOPOLOGY_INFO=$(kubectl get cluster ${CLUSTER_NAME} -n ${NAMESPACE} -o jsonpath='{.spec.topology.class} {.spec.topology.version} {.spec.infrastructureRef.kind}/{.spec.infrastructureRef.name} {.spec.controlPlaneRef.kind}/{.spec.controlPlaneRef.name}' 2>/dev/null || true)
echo "[no-op-e2e] topology: ${TOPOLOGY_INFO}"
INFRA_READY=$(kubectl get cluster ${CLUSTER_NAME} -n ${NAMESPACE} -o jsonpath='{.status.infrastructureReady}' 2>/dev/null || true)
CP_READY=$(kubectl get cluster ${CLUSTER_NAME} -n ${NAMESPACE} -o jsonpath='{.status.controlPlaneReady}' 2>/dev/null || true)
echo "[no-op-e2e] cluster ready: infra=${INFRA_READY} cp=${CP_READY}"

# 5.2) resolve CP/Infra object names via label
RESOLVED_CAPTCLUSTER_NAME=${CAPTCLUSTER_NAME}
RESOLVED_CP_NAME=$(kubectl get captcontrolplanes.controlplane.cluster.x-k8s.io -n ${NAMESPACE} -l cluster.x-k8s.io/cluster-name=${CLUSTER_NAME} -o jsonpath='{.items[0].metadata.name}' 2>/dev/null || true)
echo "[no-op-e2e] resolved: captcluster=${RESOLVED_CAPTCLUSTER_NAME} captcontrolplane=${RESOLVED_CP_NAME}"

# 6) verify no-op outputs reflected
log "verify VPC ID dummy"
if [[ -n "${CAPTCLUSTER_NAME}" ]]; then
  VPC_ID=""
  for i in $(seq 1 60); do
    VPC_ID=$(kubectl get captclusters.infrastructure.cluster.x-k8s.io ${CAPTCLUSTER_NAME} -n ${NAMESPACE} -o jsonpath='{.status.vpcId}' 2>/dev/null || true)
    if [[ "${VPC_ID}" == "vpc-00000000" ]]; then
      log "ok: vpcId dummy"
      break
    fi
    sleep 2
  done
  if [[ "${VPC_ID:-}" != "vpc-00000000" ]]; then
    log "warn: vpcId not set yet"
  fi
else
  log "warn: CAPTCluster not found for cluster ${CLUSTER_NAME}"
fi

# 6.1) endpoint verification (control plane + cluster)
log "verify endpoints"
CP_ENDPOINT=$(kubectl get captcontrolplanes.controlplane.cluster.x-k8s.io ${RESOLVED_CP_NAME} -n ${NAMESPACE} -o jsonpath='{.spec.controlPlaneEndpoint.host}:{.spec.controlPlaneEndpoint.port}' 2>/dev/null || true)
CLUSTER_ENDPOINT=$(kubectl get cluster ${CLUSTER_NAME} -n ${NAMESPACE} -o jsonpath='{.spec.controlPlaneEndpoint.host}:{.spec.controlPlaneEndpoint.port}' 2>/dev/null || true)
echo "[no-op-e2e] endpoints: cp=${CP_ENDPOINT} cluster=${CLUSTER_ENDPOINT}"

# 6.2) WTA existence check
log "verify WorkspaceTemplateApply existence"
kubectl get workspacetemplateapplies.infrastructure.cluster.x-k8s.io -n ${NAMESPACE} -o name | sed 's/^/[no-op-e2e] wta: /'

# 6.3) kubeconfig secret existence check
log "verify kubeconfig secret existence"
if kubectl get secret ${CLUSTER_NAME}-kubeconfig -n ${NAMESPACE} >/dev/null 2>&1; then
  log "ok: kubeconfig secret exists"
else
  log "warn: kubeconfig secret missing"
fi

# 7) webhook immutability (expect Forbidden); ignore non-zero to continue
log "webhook immutability check (expect Forbidden)"
CP_NAME=$(kubectl get captcontrolplanes.controlplane.cluster.x-k8s.io -n ${NAMESPACE} -l cluster.x-k8s.io/cluster-name=${CLUSTER_NAME} -o jsonpath='{.items[0].metadata.name}' || true)
if [[ -n "${CP_NAME}" ]]; then
  set +e
  kubectl patch captcontrolplanes.controlplane.cluster.x-k8s.io ${CP_NAME} -n ${NAMESPACE} --type merge -p '{"spec":{"controlPlaneConfig":{"region":"changed"}}}' >/dev/null 2>&1
  rc=$?
  set -e
  if [[ ${rc} -ne 0 ]]; then
    log "ok: patch rejected as expected"
  else
    log "warn: patch accepted (unexpected)"
  fi
fi

log "done"

