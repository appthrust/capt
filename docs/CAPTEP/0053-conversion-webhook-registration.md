# CAPTEP-0053: Avoid Double Registration of Conversion Webhook (/convert)

## Summary

Remove manual registration of the conversion webhook path `/convert` and rely on controller-runtime's automatic registration when Hub/Convertible is implemented and `SetupWebhookWithManager` is invoked.

## Motivation

Manual registration in `cmd/main.go` caused a panic due to duplicate path registration when types also enabled conversion via controller-runtime.

### Goals

- Prevent duplicate registration panics
- Centralize webhook registration in one place

### Non-Goals

- Changing type definitions beyond Hub/Convertible requirements

## Proposal

- Delete `mgr.GetWebhookServer().Register("/convert", ...)` from `cmd/main.go`.
- Ensure types implement Hub/Convertible as needed and call `SetupWebhookWithManager` once.
- Establish a convention: webhook registrations occur in exactly one place (main), zero or one time.

### Implementation Details

- Audit for duplicate registration sites and remove them.
- Add a short comment near the setup code to document the convention.

### Risks and Mitigations

- Risk: Multiple managers or packages accidentally re-register
  - Mitigation: Keep registration in main only; add code review checklist.

## Alternatives Considered

- Keep manual registration (unsafe; causes duplicate path panic)

## Upgrade Strategy

- Remove manual registration and rebuild; no cluster migration required.

## References

- `cmd/main.go`
