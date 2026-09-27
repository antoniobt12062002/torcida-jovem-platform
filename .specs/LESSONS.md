# LESSONS - auto-maintained by scripts/lessons.py

> Machine-owned. Do NOT hand-edit. Changes are overwritten on the next `lessons.py` write.
> Canonical state lives in `.specs/lessons.json`. Edit lessons only via the script.
> promote_threshold=2 distinct features · window_days=45 · quarantine_threshold=2

## Confirmed (load these at Specify/Design)

Corroborated across multiple features. Safe to apply as guidance.

_none_

## Candidates (under observation - do NOT load as guidance yet)

Seen once or not yet corroborated. Tracked, not trusted.

### L-001 - For a service that writes to an external store before a DB transaction, add a test that injects a failing store double and asserts zero rows were written at the service level, not just that the adapter itself returns an error.
- signal: `ac_gap` · recurrence: 1 feature(s) · scope: `platform/documents` · harmful: 0
- features: fundacao-documentos
- evidence: spec.md Edge Cases ("storage unreachable during a store operation"); api/internal/platform/documents/service.go:220-224 (platform/documents)
- last seen: 2026-09-27T18:13:59Z

## Quarantined (failed when applied - ignore)

A confirmed lesson that recurred alongside failure. Kept for the maintainer to review.

_none_
