# CAPTEP-0050: Unify CA Injection via cert-manager inject-ca-from Annotation

## Summary

Standardize CA bundle injection for Admission and CRD conversions using `cert-manager.io/inject-ca-from` annotations pointing to the serving certificate Secret.

## Motivation

We require consistent CA injection for admission webhooks and CRD conversion. Some patches previously produced `"cert-manager.io/inject-ca-from: null"`, leading to missing `caBundle` on the CRD side.

### Goals

- Ensure reliable CA injection into MWC/VWC and CRD conversion
- Avoid null or missing annotations in rendered manifests

### Non-Goals

- Changing certificate issuance mechanisms beyond the required annotations

## Proposal

- In `config/default/kustomization.yaml`, add `cert-manager.io/inject-ca-from: capt-system/capt-serving-cert` via JSON6902 patches to:
  - `ValidatingWebhookConfiguration`
  - `MutatingWebhookConfiguration`
  - `CustomResourceDefinition` (conversion)
- In `config/certmanager`, align Issuer/Certificate with name prefixes and ensure the certificate SubjectAltName matches the final Service FQDN: `capt-webhook-service.capt-system.svc`.

### Implementation Details

- Keep patches concise and idempotent; run `make wait-ca` to verify bundle injection.
- Confirm that rendered objects contain the `caBundle` after cert-manager reconciliation.

### Risks and Mitigations

- Risk: Annotation key drift or namespace/name changes
  - Mitigation: Centralize names; validate in CI with a manifest linter.

## Alternatives Considered

- Manual CA bundle management (higher operational burden, error-prone)

## Upgrade Strategy

- Apply the annotation patches and re-render manifests; no downtime expected.

## References

- `config/default/kustomization.yaml`
- `config/certmanager/certificate.yaml`
