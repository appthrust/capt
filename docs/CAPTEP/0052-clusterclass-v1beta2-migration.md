# CAPTEP-0052: Migrate ClusterClass/Cluster/Bootstrap to v1beta2

## Summary

Migrate key Cluster API resources from v1beta1 to v1beta2, updating field names and structures in samples to remove warnings and avoid strict decoding errors.

## Motivation

v1beta1 produced warnings (e.g., for Cluster and KubeadmConfigTemplate). v1beta2 introduces field and structural changes that we must adopt to remain current and error-free.

### Goals

- Update samples to v1beta2 API versions and field structures
- Eliminate warnings and strict decoding errors

### Non-Goals

- Broader refactors beyond required API structure updates

## Proposal

Key changes:

- Cluster.topology: `class` → `classRef.name`
- ClusterClass.workers.machineDeployments[].template structure split:
  - `template.bootstrap.ref` → `bootstrap.templateRef`
  - `template.infrastructure.ref` → `infrastructure.templateRef`
- KubeadmConfigTemplate v1beta2: tighten to a minimal no-op sample that only defines `preKubeadmCommands`

### Implementation Details

- Update samples:
  - `config/samples/clusterclass-e2e/cluster.yaml`
  - `config/samples/clusterclass-e2e/clusterclass.yaml`
  - `config/samples/clusterclass-e2e/kubeadmconfigtemplate.yaml`
- Validate with `kubectl apply --dry-run=client` and ensure no warnings.

### Risks and Mitigations

- Risk: Missed field renames
  - Mitigation: Cross-check against CAPI v1beta2 docs and CRD schemas.

## Alternatives Considered

- Remain on v1beta1 (incurs warnings and future incompatibility)

## Upgrade Strategy

- Update manifests and re-run e2e samples; no runtime migration logic required.

## References

- CAPI v1beta2 docs
- Project sample manifests listed above
