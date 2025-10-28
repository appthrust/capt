# CAPTEP-0048: Adopt "Keep a Changelog" Format

## Summary

Adopt the Keep a Changelog format for `CHANGELOG.md` to clearly communicate project changes with consistent categorization and version/date clarity, while remaining friendly to automation.

## Motivation

We need a human-readable, well-structured, and tool-parseable way to track project changes and communicate them to users and contributors.

### Goals

- Provide a standard, human-friendly changelog
- Clearly distinguish change types (features, fixes, etc.)
- Record release versions and dates consistently
- Support automation and tooling

### Non-Goals

- Redefining the project versioning policy
- Building full release automation in this proposal

## Proposal

Adopt the Keep a Changelog format:

1. Maintain the changelog at `CHANGELOG.md`.
2. Use the following sections to classify changes:
   - Added: New features
   - Changed: Changes to existing functionality
   - Deprecated: Soon-to-be removed features
   - Removed: Removed features
   - Fixed: Bug fixes
   - Security: Security-related fixes
3. Record each release in the following format:

```markdown
## [version] - YYYY-MM-DD

### Added
- Description of new features

### Fixed
- Description of bug fixes
```

### Implementation Details

- Create and maintain `CHANGELOG.md` in the repository root.
- Require changelog updates as part of the release process and major PRs.
- Optionally add a CI check to ensure a changelog entry exists for release PRs.

### Risks and Mitigations

- Risk: Inconsistent use across contributors
  - Mitigation: Provide a short template in `CONTRIBUTING.md` and enforce via PR checklist.

## Alternatives Considered

- Ad-hoc changelog entries without a standard format
- Using only GitHub Releases notes (less structured for version control)

## Upgrade Strategy

- Start using this format from the next release onward.
- Backfill recent releases as time permits to improve historical consistency.

## References

- Keep a Changelog: https://keepachangelog.com/
- Semantic Versioning: https://semver.org/
