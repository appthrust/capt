# CAPTEP-0054: Ensure Webhook Readiness Before Admission in e2e

## Summary

In e2e tests, wait for the webhook server to be fully ready before applying CRs to avoid `connect: connection refused` errors and stabilize test runs.

## Motivation

Applying CRs while the Admission Webhook is not yet ready leads to intermittent failures. Adding readiness checks improves reliability.

### Goals

- Ensure webhook readiness prior to running admission-dependent tests
- Eliminate flaky connection errors in e2e

### Non-Goals

- Modifying webhook logic beyond readiness waiting

## Proposal

Add the following waits to the e2e script:

- `kubectl -n capt-system rollout status deploy/capt-controller-manager`
- Wait for Secret `webhook-server-cert` to exist
- Wait for Service `capt-webhook-service` Endpoints to exist
- Add a short settle time (2s)

### Implementation Details

- Update `test/e2e/scripts/no-op-clusterclass.sh` to include the waits.
- Keep waits bounded with reasonable timeouts.

### Risks and Mitigations

- Risk: Longer e2e duration
  - Mitigation: Keep waits minimal and only at the start.

## Alternatives Considered

- None; skipping readiness checks is flaky.

## Upgrade Strategy

- Land the script updates; no cluster changes required.

## References

- `test/e2e/scripts/no-op-clusterclass.sh`
