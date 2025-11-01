#!/usr/bin/env bash

set -euo pipefail

# Karpenter installer for EKS (no Sveltos)
# - Discovers the EKS cluster name and connection secret from WorkspaceTemplateApply
# - Waits for EKS ACTIVE, updates kubeconfig
# - Installs Karpenter via Helm (OCI chart)
# - Applies default EC2NodeClass/NodePool from samples/karpenter/default-nodepool.yaml

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

MGMT_NS="mgmt"
REGION="ap-northeast-1"
CHART_VERSION="1.6.3"
HELM_CHART="oci://public.ecr.aws/karpenter/karpenter"
CLUSTER_NAME=""
MGMT_CONTEXT=""
EKS_CONTEXT=""
NODEPOOL_FILE="$REPO_ROOT/samples/karpenter/default-nodepool.yaml"
SKIP_NODEPOOL="false"
WAIT_TIMEOUT_SEC=900
ROLE_ARN_OVERRIDE=""
QUEUE_NAME_OVERRIDE=""

log() { echo "[karpenter] $*"; }
err() { echo "[karpenter][ERROR] $*" 1>&2; }

usage() {
  cat <<EOF
Usage: $0 [--cluster <name>] [--region <aws-region>] [--mgmt-namespace <ns>] \
          [--chart-version <ver>] [--nodepool-file <path>] [--skip-nodepool] \
          [--role-arn <arn>] [--queue-name <name>]

Options:
  --cluster           EKS cluster name (defaults to auto-detect from WTA)
  --region            AWS region (default: ${REGION})
  --mgmt-namespace    Namespace where WTA/Secrets live (default: ${MGMT_NS})
  --chart-version     Karpenter Helm chart version (default: ${CHART_VERSION})
  --nodepool-file     Path to NodePool manifest with __CLUSTER__ placeholders
                      (default: ${NODEPOOL_FILE})
  --skip-nodepool     Install controller only, skip NodePool apply
  --role-arn          IRSA role ARN for Karpenter service account (override)
  --queue-name        SQS queue name for Karpenter interruption handling (override)

Prereqs:
  - aws, kubectl, helm (v3 with OCI support)
  - AWS credentials configured for the target account
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --cluster)
      CLUSTER_NAME="$2"; shift 2 ;;
    --region)
      REGION="$2"; shift 2 ;;
    --mgmt-namespace)
      MGMT_NS="$2"; shift 2 ;;
    --chart-version)
      CHART_VERSION="$2"; shift 2 ;;
    --nodepool-file)
      NODEPOOL_FILE="$2"; shift 2 ;;
    --skip-nodepool)
      SKIP_NODEPOOL="true"; shift 1 ;;
    --role-arn)
      ROLE_ARN_OVERRIDE="$2"; shift 2 ;;
    --queue-name)
      QUEUE_NAME_OVERRIDE="$2"; shift 2 ;;
    -h|--help)
      usage; exit 0 ;;
    *)
      err "Unknown argument: $1"; usage; exit 1 ;;
  esac
done

command -v kubectl >/dev/null 2>&1 || { err "kubectl not found"; exit 1; }
command -v helm >/dev/null 2>&1 || { err "helm not found"; exit 1; }
command -v aws >/dev/null 2>&1 || { err "aws cli not found"; exit 1; }
command -v jq >/dev/null 2>&1 || { err "jq not found (required for AWS discovery). Provide --role-arn/--queue-name or install jq"; }

# Capture current management context if not preset (may be unused when --cluster specified)
if [[ -z "$MGMT_CONTEXT" ]]; then
  MGMT_CONTEXT="$(kubectl config current-context)"
fi

# Discover EKS WorkspaceTemplateApply and cluster name if not provided
discover_wta() {
  local pattern='eks-controlplane-apply$'
  local names
  names=$(kubectl --context "$MGMT_CONTEXT" -n "$MGMT_NS" get workspacetemplateapply -o name 2>/dev/null | grep -E "$pattern" || true)
  if [[ -z "$names" ]]; then
    err "No WorkspaceTemplateApply matching '*-eks-controlplane-apply' found in namespace $MGMT_NS (context: $MGMT_CONTEXT)"
    return 1
  fi

  # If cluster name provided, pick matching WTA; else require uniqueness
  if [[ -n "$CLUSTER_NAME" ]]; then
    local wta
    wta=$(echo "$names" | awk -F/ '{print $2}' | grep -E "^${CLUSTER_NAME}-.*-eks-controlplane-apply$" || true)
    if [[ -z "$wta" ]]; then
      err "No matching WTA for cluster '$CLUSTER_NAME' in $MGMT_NS (context: $MGMT_CONTEXT)"
      return 1
    fi
    echo "$wta"
    return 0
  fi

  local count
  count=$(echo "$names" | wc -l | awk '{print $1}')
  if [[ "$count" -ne 1 ]]; then
    err "Multiple EKS WTAs found, specify --cluster. Candidates:\n$names"
    return 1
  fi
  echo "$names" | awk -F/ '{print $2}'
}

# Determine WTA and cluster name
WTA_NAME=""
if [[ -z "$CLUSTER_NAME" ]]; then
  WTA_NAME="$(discover_wta)"
  CLUSTER_NAME="$(kubectl --context "$MGMT_CONTEXT" -n "$MGMT_NS" get workspacetemplateapply "$WTA_NAME" -o jsonpath='{.spec.variables.cluster_name}')"
  if [[ -z "$CLUSTER_NAME" ]]; then
    err "Failed to determine cluster name"; exit 1
  fi
fi

log "Using cluster: $CLUSTER_NAME (mgmt context: $MGMT_CONTEXT, region: $REGION)"

discover_role_via_aws() {
  # Prefer IRSA trust policy match: federated OIDC provider for this cluster and SA subject
  local issuer issuer_noproto account provider_arn roles_json arn name
  issuer=$(aws eks describe-cluster --region "$REGION" --name "$CLUSTER_NAME" --query 'cluster.identity.oidc.issuer' --output text 2>/dev/null || true)
  account=$(aws sts get-caller-identity --query 'Account' --output text 2>/dev/null || true)
  issuer_noproto=${issuer#https://}
  provider_arn="arn:aws:iam::${account}:oidc-provider/${issuer_noproto}"
  roles_json=$(aws iam list-roles --query 'Roles[*].{RoleName:RoleName,Arn:Arn}' --output json 2>/dev/null || echo '[]')
  # Use jq if available to select role trusted for SA karpenter/karpenter via IRSA
  if command -v jq >/dev/null 2>&1; then
    arn=$(aws iam list-roles --output json \
      | jq -r --arg prov "$provider_arn" --arg sub "system:serviceaccount:karpenter:karpenter" '
        .Roles[]
        | select(.AssumeRolePolicyDocument != null)
        | select([.AssumeRolePolicyDocument.Statement[]?]
            | map((.Principal.Federated // "") | tostring | contains($prov))
            | any)
        | select([.AssumeRolePolicyDocument.Statement[]?]
            | map((.Condition?.StringEquals? // {} | to_entries[]?) | select(.key | endswith(":sub")) | .value == $sub)
            | any)
        | .Arn' 2>/dev/null | head -n1)
    if [[ -n "$arn" ]]; then
      echo "$arn"; return 0; fi
  fi
  # Fallback: name heuristic
  name=$(aws iam list-roles --query 'Roles[].RoleName' --output text 2>/dev/null | tr '\t' '\n' | grep -Ei "(karpenter).*(${CLUSTER_NAME})" | head -n1 || true)
  if [[ -n "$name" ]]; then
    aws iam get-role --role-name "$name" --query 'Role.Arn' --output text 2>/dev/null || true
  fi
}

discover_queue_via_aws() {
  # Try to infer from Karpenter controller role policy resources (SQS ARN -> name)
  local role_arn role_name attached arns version doc qname urls pick=""
  role_arn="$ROLE_ARN"
  if [[ -z "$role_arn" ]]; then
    role_arn="$(discover_role_via_aws)"
  fi
  if [[ -n "$role_arn" ]]; then
    role_name=${role_arn##*/}
    # Managed policies
    attached=$(aws iam list-attached-role-policies --role-name "$role_name" --query 'AttachedPolicies[].PolicyArn' --output text 2>/dev/null | tr '\t' '\n' || true)
    while read -r p; do
      [[ -z "$p" ]] && continue
      version=$(aws iam get-policy --policy-arn "$p" --query 'Policy.DefaultVersionId' --output text 2>/dev/null || true)
      doc=$(aws iam get-policy-version --policy-arn "$p" --version-id "$version" --query 'PolicyVersion.Document' --output json 2>/dev/null || echo '{}')
      qname=$(echo "$doc" | jq -r '.. | objects | select(has("Action")) | select((.Action|tostring)|test("sqs")) | .Resource? | if type=="array" then .[] else . end | select(type=="string") | select(test(":sqs:")) | split(":")[-1]' 2>/dev/null | head -n1)
      [[ -n "$qname" ]] && { echo "$qname"; return 0; }
    done <<<"$attached"
    # Inline policies
    attached=$(aws iam list-role-policies --role-name "$role_name" --query 'PolicyNames[]' --output text 2>/dev/null | tr '\t' '\n' || true)
    while read -r pn; do
      [[ -z "$pn" ]] && continue
      doc=$(aws iam get-role-policy --role-name "$role_name" --policy-name "$pn" --query 'PolicyDocument' --output json 2>/dev/null || echo '{}')
      qname=$(echo "$doc" | jq -r '.. | objects | select(has("Action")) | select((.Action|tostring)|test("sqs")) | .Resource? | if type=="array" then .[] else . end | select(type=="string") | select(test(":sqs:")) | split(":")[-1]' 2>/dev/null | head -n1)
      [[ -n "$qname" ]] && { echo "$qname"; return 0; }
    done <<<"$attached"
  fi
  # Fallback heuristics
  urls=$(aws sqs list-queues --region "$REGION" --queue-name-prefix karpenter --query 'QueueUrls[]' --output text 2>/dev/null || true)
  if [[ -n "$urls" ]]; then
    pick=$(echo "$urls" | tr '\t' '\n' | awk -F/ '{print $NF}' | grep -Ei "(karpenter).*(${CLUSTER_NAME})" | head -n1 || true)
  fi
  if [[ -z "$pick" ]]; then
    urls=$(aws sqs list-queues --region "$REGION" --queue-name-prefix "$CLUSTER_NAME" --query 'QueueUrls[]' --output text 2>/dev/null || true)
    [[ -n "$urls" ]] && pick=$(echo "$urls" | tr '\t' '\n' | awk -F/ '{print $NF}' | head -n1)
  fi
  [[ -n "$pick" ]] && echo "$pick" || true
}

# Ensure EKS is ACTIVE and kubeconfig is set first
log "Waiting for EKS cluster ACTIVE: $CLUSTER_NAME"
aws eks wait cluster-active --region "$REGION" --name "$CLUSTER_NAME"
EKS_CONTEXT="eks-${CLUSTER_NAME}"
aws eks update-kubeconfig --region "$REGION" --name "$CLUSTER_NAME" --alias "$EKS_CONTEXT"

# Resolve cluster endpoint via AWS
CLUSTER_ENDPOINT=$(aws eks describe-cluster --region "$REGION" --name "$CLUSTER_NAME" --query 'cluster.endpoint' --output text || true)

# Resolve Karpenter role ARN and SQS queue via AWS (or CLI overrides)
ROLE_ARN="$ROLE_ARN_OVERRIDE"
QUEUE_NAME="$QUEUE_NAME_OVERRIDE"
[[ -z "$ROLE_ARN" ]] && ROLE_ARN="$(discover_role_via_aws)"
[[ -z "$QUEUE_NAME" ]] && QUEUE_NAME="$(discover_queue_via_aws)"

if [[ -z "$ROLE_ARN" || -z "$QUEUE_NAME" ]]; then
  err "Could not resolve Karpenter ROLE_ARN/QUEUE_NAME. Provide --role-arn and --queue-name explicitly."
  exit 1
fi

log "EKS endpoint: ${CLUSTER_ENDPOINT:-<resolved via AWS>}"
log "Karpenter IRSA role: $ROLE_ARN"
log "Karpenter SQS queue: $QUEUE_NAME"

# Ensure namespace exists
kubectl get ns karpenter >/dev/null 2>&1 || kubectl create namespace karpenter

# Install/upgrade Karpenter via Helm (values aligned with provided examples)
log "Installing/Upgrading Karpenter chart ${CHART_VERSION}"
helm --kube-context "$EKS_CONTEXT" upgrade --install karpenter "$HELM_CHART" \
  --version "$CHART_VERSION" \
  --namespace karpenter \
  --set dnsPolicy=Default \
  --set priorityClassName=system-cluster-critical \
  --set settings.clusterName="$CLUSTER_NAME" \
  --set settings.clusterEndpoint="$CLUSTER_ENDPOINT" \
  --set settings.featureGates.spotToSpotConsolidation=true \
  --set settings.interruptionQueue="$QUEUE_NAME" \
  --set serviceAccount.annotations."eks\.amazonaws\.com/role-arn"="$ROLE_ARN" \
  --wait --timeout 10m

log "Karpenter controller installed. Verifying pods (this may run on Fargate)."
kubectl --context "$EKS_CONTEXT" -n karpenter get pods -o wide || true

if [[ "$SKIP_NODEPOOL" == "true" ]]; then
  log "--skip-nodepool specified; skipping NodePool/EC2NodeClass apply."
  exit 0
fi

# Apply default NodePool/EC2NodeClass, replacing __CLUSTER__ placeholders
if [[ ! -f "$NODEPOOL_FILE" ]]; then
  err "NodePool file not found: $NODEPOOL_FILE"; exit 1
fi

log "Applying NodePool from: $NODEPOOL_FILE (cluster=$CLUSTER_NAME, context=$EKS_CONTEXT)"
sed "s/__CLUSTER__/${CLUSTER_NAME}/g" "$NODEPOOL_FILE" | kubectl --context "$EKS_CONTEXT" apply -f -

log "Done. You can verify nodes after a short while with: kubectl get nodes"


