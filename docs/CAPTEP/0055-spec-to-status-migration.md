# CAPTEP-0055: Spec-to-Status Migration for WorkspaceTemplateApply Naming

## Status

Accepted

## Summary

Align CAPT with ClusterTopology immutability by avoiding controller writes to spec after creation. Migrate internal linkage from `spec.workspaceTemplateApplyName` to status-based observability and deterministic naming.

## Motivation

- Under ClusterTopology, `spec` is considered desired state and must be immutable post-creation (except via topology patches).
- Controllers writing back to `spec.workspaceTemplateApplyName` conflicts with this principle and may be rejected by admission in future.
- Operators still need to observe the actual Terraform workspace used by the controller.

## Goals

- Stop mutating `spec.workspaceTemplateApplyName` in controllers when managed by ClusterTopology.
- Keep deterministic naming for WorkspaceTemplateApply and surface the created workspace via status.
- Preserve compatibility for existing users while enabling stricter immutability validation.

## Non-Goals

- Removing the `spec.workspaceTemplateApplyName` field immediately.
- Introducing breaking API changes in this release.

## Design

### Deterministic Naming

- Control plane WTA name: `<captcontrolplane-name>-eks-controlplane-apply`.
- Cluster VPC WTA name: `<captcluster-name>-vpc`.
- Names are derived at reconcile time and used for get/create/update without persisting back to `spec`.

### Status Observability

- Add `status.workspaceTemplateStatus.workspaceName` to `CAPTControlPlane` and `CAPTCluster`.
- Controllers set this field from the associated `WorkspaceTemplateApply.status.workspaceName`.

### API Compatibility

- The `spec.workspaceTemplateApplyName` field remains present but controllers no longer write to it under topology.
- Future releases may remove the field once downstream users migrate.

## Alternatives Considered

1) Admission defaulting on CREATE to populate `spec.workspaceTemplateApplyName`.
   - Pros: value present in spec from the start
   - Cons: hard to evolve naming; still exposes spec mutation risk if changed later

2) Annotations for internal references
   - Pros: no API change
   - Cons: less structured than status; weaker contract

## Risks and Mitigations

- Risk: External automation relying on `spec.workspaceTemplateApplyName` updates.
  - Mitigation: Field is retained (readable), documentation updated to use status instead.

## Upgrade / Migration

- No data migration required. Controllers stop writing to spec on upgrade.
- Users should update dashboards and tooling to prefer `status.workspaceTemplateStatus.workspaceName`.

## Test Plan

- Unit tests updated to verify deterministic names and status population.
- E2E no-op topology validates expected WTA names and generated kubeconfig.

## Follow-ups

- Add `v1beta2` webhook immutability validation for `CAPTControlPlane`.
- Consider deprecating and eventually removing the spec field in a major version.
