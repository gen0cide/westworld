# External agent charters

> **STATUS: CURRENT** — verified 2026-06-10 against branch HEAD `0bfa818`. One live
> charter remains (the OpenRSC steward). The dev-partner charter is archived: the
> cognitive layers it scaffolded are built — see the status matrix in
> [docs/index.md](../index.md).

Onboarding briefs for the non-Claude AI agents collaborating on westworld. Each file
is a self-contained charter for a zero-context agent. The human (Alex) pastes a short
bootstrap line (below) into the external tool; the agent reads its charter and reports
back before acting.

| Charter | Role | Recommended tool | Territory |
|---|---|---|---|
| [openrsc-steward.md](openrsc-steward.md) | Run the OpenRSC server + be the authentic-client reference oracle | OpenAI Codex | `~/Code/openrsc` (read-only on `~/Code/westworld`) |
| [westworld-dev-partner.md](../archive/initial-brainstorming/westworld-dev-partner.md) | ARCHIVED 2026-06-10 — the layers it scaffolded are built | (historical) | — |

**Lanes today.** Claude owns the westworld Go monorepo — the cognitive stack, the
cradle daemon (`cmd/cradle-server` + `cmd/cradle-ctl`), and the software renderer
(`render/`, `cmd/legacy-cradle/spectate.go`) — and is not covered by a charter here.
The steward (Codex) owns the OpenRSC server and the authentic Java reference sources:
a focused job (one live server + targeted source lookups) that fits a single-territory
tool. The dev-partner lane (formerly Google Antigravity, cross-package scaffolding of
`cognition`/`brain`/`persona`/`memory`/`mesa`) is closed; its charter is kept in the
archive for the conventions record.

## Bootstrap line

**OpenRSC steward (Codex):**
> You are joining a project with zero prior context. Read
> `~/Code/westworld/docs/agents/openrsc-steward.md` in full — it is your charter —
> then follow its "First task" and report back before changing anything.
