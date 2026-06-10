# Lessons learned

These chapters capture problems we hit and the fixes we determined, so we never
re-fight a settled battle. New chapters are added when a saga closes, with the
durable rules up front.

## Chapters

| Chapter | Hook |
| --- | --- |
| [1_PATHING.md](1_PATHING.md) | Walk-through-walls, mishandled doors, and "boxed upstairs" were one root cause wearing five costumes: a static-only, plane-0-locked, dynamic-blind collision grid. |
| [2_WIRE-PROTOCOL.md](2_WIRE-PROTOCOL.md) | Byte-identical payloads are still two opcodes (50 cast vs 135 use-on-NPC), and single-entry reads silently hid the rest of multi-entry ground-item frames — decoders fail loudly or they hide cascades. |
| [3_CONFIG-AND-ENVIRONMENT-DRIFT.md](3_CONFIG-AND-ENVIRONMENT-DRIFT.md) | `westworld.conf` was maintained as copies instead of a symlink and drifted twice on schedule — copies drift; symlink or generate. |
| [4_DOCS-AND-PROCESS.md](4_DOCS-AND-PROCESS.md) | Hand-maintained status matrices and banners go stale in days, not months — a status assertion without enforcement is a future lie. |
| [5_LLM-FACING-DOCS.md](5_LLM-FACING-DOCS.md) | The DSL manual's flagship example taught `&&` — a guaranteed parse error — and six reviewers missed it; every example in an LLM-facing doc must be machine-validated, because prose presence predicts usage. |
| [6_COGNITION-DESIGN.md](6_COGNITION-DESIGN.md) | A 6,000-line, 44-diagram v2 architecture was rejected on one paragraph ("there is no tick; there never will be") — quote the real code before designing, and reject framings while salvaging mechanisms. |
