#!/usr/bin/env bash

set -euo pipefail

NS="upbound-system"
SECRET_NAME="capt-aws-creds"

info()  { echo "[INFO]  $*"; }
warn()  { echo "[WARN]  $*" >&2; }
error() { echo "[ERROR] $*" >&2; exit 1; }

require_cmd() {
  command -v "$1" >/dev/null 2>&1 || error "'$1' が見つかりません。インストールしてください。"
}

main() {
  require_cmd kubectl

  # 既存の環境変数を尊重。未設定ならプロンプトで入力。
  local region="${AWS_REGION:-ap-northeast-1}"  # 参照のみ（Secretには含めない）
  local akid="${AWS_ACCESS_KEY_ID:-}"
  local secret="${AWS_SECRET_ACCESS_KEY:-}"
  local token="${AWS_SESSION_TOKEN:-}"

  if [[ -z "${akid}" ]]; then
    read -r -p "AWS_ACCESS_KEY_ID: " akid
  fi
  if [[ -z "${secret}" ]]; then
    read -r -s -p "AWS_SECRET_ACCESS_KEY: " secret
    echo ""
  fi
  # セッショントークンは任意。環境変数があれば自動で含める。

  info "Namespace '${NS}' を確認/作成"
  kubectl get ns "${NS}" >/dev/null 2>&1 || kubectl create ns "${NS}" >/dev/null

  info "Secret '${SECRET_NAME}' を適用します (context region=${region})"
  cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: Secret
metadata:
  name: ${SECRET_NAME}
  namespace: ${NS}
type: Opaque
stringData:
  credentials: |
    [default]
    aws_access_key_id = ${akid}
    aws_secret_access_key = ${secret}
$( [[ -n "${token}" ]] && printf "    aws_session_token = %s\n" "${token}" )
EOF

  info "適用が完了しました。"
}

main "$@"


