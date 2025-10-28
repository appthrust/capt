# CAPTEP-0049: Resolve Webhook Service References with Kustomize and Harden with Name-Based Merge Patches

## Summary

Use Kustomize name/namespace references to resolve webhook `clientConfig.service` and harden the configuration with name-based merge patches, making it resilient to webhook list order changes.

## Motivation

Previously, JSON6902 patches targeted array indices (e.g., `/webhooks/0/...`) in `MutatingWebhookConfiguration` and `ValidatingWebhookConfiguration`. These were brittle as webhook additions or reordering broke the patches.

### Goals

- Make webhook Service name/namespace references robust against list changes
- Avoid index-based JSON patches that are order-sensitive

### Non-Goals

- Redesigning webhook content or admission logic

## Proposal

- Configure Kustomize name/namespace resolution for Service references:
  - `config/webhook/kustomizeconfig.yaml`
  - `config/default/kustomizeconfig.yaml`
- Rely on `namePrefix: capt-` so that final names resolve to `capt-webhook-service` in the `capt-system` namespace.
- As defense-in-depth, apply name-merge YAML patches to force the final `clientConfig.service` values without depending on array indices:
  - `config/default/webhook_service_patch_mutating.yaml`
  - `config/default/webhook_service_patch_validating.yaml`

### Implementation Details

- Ensure Kustomize nameReference/namespace mappings cover both MWC and VWC objects.
- Keep patches minimal and name-scoped to avoid unintended changes.
- Validate final rendered manifests to confirm correct service FQDN resolution.

### Risks and Mitigations

- Risk: Name prefix changes or refactors could break assumptions
  - Mitigation: Keep name prefix centralized; CI validates rendered webhook targets.

## Alternatives Considered

- Continue using index-based JSON6902 patches (rejected due to fragility)

## Upgrade Strategy

- Replace existing index-based patches with the name-based approach.
- Verify production clusters render identical final webhook service targets.

## References

- `config/webhook/kustomizeconfig.yaml`
- `config/default/kustomizeconfig.yaml`
- `config/default/webhook_service_patch_mutating.yaml`
- `config/default/webhook_service_patch_validating.yaml`
