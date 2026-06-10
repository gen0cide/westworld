# External agent charters

Onboarding briefs for the non-Claude AI agents collaborating on westworld. Each file
is a self-contained charter for a zero-context agent. The human (Alex) pastes a short
bootstrap line (below) into the external tool; the agent reads its charter and reports
back before acting.

| Charter | Role | Recommended tool | Territory |
|---|---|---|---|
| [openrsc-steward.md](openrsc-steward.md) | Run the OpenRSC server + be the authentic-client reference oracle | OpenAI Codex | `~/Code/openrsc` (read-only on `~/Code/westworld`) |
| [westworld-dev-partner.md](../archive/initial-brainstorming/westworld-dev-partner.md) | ARCHIVED 2026-06-10 — the layers it scaffolded are built | (historical) | — |

> **Tool assignment rationale.** Antigravity takes the **dev-partner** role
> because scaffolding the layer cake means holding many westworld packages
> (`cognition`/`brain`/`persona`/`memory`/`reveries`/`mesa`/`runtime`) in context
> at once — it reads more folders simultaneously, so it's the better fit for
> cross-package work. Codex takes the **steward** role: a focused job (one live
> server + targeted lookups into the Java reference sources). Swap if your usage
> patterns suggest otherwise — both charters are written tool-agnostic.

Claude owns the Go software renderer (`render/`, `cmd/cradle/spectate.go`) and overall
architecture — it is not covered by a charter here.

## Bootstrap lines

**OpenRSC steward (Codex):**
> You are joining a project with zero prior context. Read
> `~/Code/westworld/docs/agents/openrsc-steward.md` in full — it is your charter —
> then follow its "First task" and report back before changing anything.

**westworld dev partner (Antigravity):**
> You are joining a project with zero prior context. Read
> `~/Code/westworld/docs/archive/initial-brainstorming/westworld-dev-partner.md` in full — it is your charter (ARCHIVED: the scaffold work completed; kept for the conventions record)
> — then follow its "First task" and report back before writing any code.
