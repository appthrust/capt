# CAPTEP-0051: Use Enhancement Proposals for Significant Changes

## Summary

Adopt a standardized Enhancement Proposal format (CAPTEP) for proposing and documenting significant changes to CAPT, inspired by KEPs and tailored to project needs.

## Motivation

As the project grows, we need a structured approach to propose, discuss, and record major enhancements before implementation, beyond brief ADRs.

### Goals

- Standardize the way we document significant enhancements
- Provide comprehensive sections covering motivation, user stories, and risks
- Improve review quality and maintain a clear historical record

### Non-Goals

- Replacing ADRs or design docs for minor decisions
- Adding unnecessary bureaucracy to small changes or bug fixes

## Proposal

- Use the CAPTEP format defined in CAPTEP-0029 for significant changes:
  - Clear metadata (number, title, status)
  - Sections: Summary, Motivation, Goals/Non-Goals, Proposal (User Stories, Implementation Details, Risks), Alternatives, Upgrade Strategy
- Write CAPTEPs in English and store them under `docs/CAPTEP/` as `NNNN-descriptive-title.md`.

### Implementation Details

- Introduce a template based on CAPTEP-0029.
- Reference CAPTEP-0029 in new proposals; evolve the format as needed.
- Encourage contributors to use CAPTEPs for impactful changes.

### Risks and Mitigations

- Risk: Process overhead slows development
  - Mitigation: Limit CAPTEPs to significant changes; keep the template pragmatic.
- Risk: Format drift over time
  - Mitigation: Add lightweight format validation and reviewer checklists.

## Alternatives Considered

- Continue using ADRs only (insufficient depth for major enhancements)
- GitHub issues only (lacks structure and traceability)

## Upgrade Strategy

- Start using CAPTEPs for new significant proposals immediately; no migration required for minor past decisions.

## References

- CAPTEP-0029: Standardizing Enhancement Proposal Format (`docs/CAPTEP/0029-enhancement-proposal-format.md`)
- Kubernetes Enhancement Proposals: https://github.com/kubernetes/enhancements/tree/master/keps
