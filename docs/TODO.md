# TODO — single source of truth

This is the ONE list. Work against it; add to it; cull into Appendix A with evidence.
Built from the 2026-06-10 repo-wide TODO consolidation (every docs/ backlog section, code
markers, and the _research backlogs — see `docs/_research/docs-audit/`), every status
verified against code / `git log --all`, not just docs.

**How to read this file**

- Items marked **†** were added during the docs-audit reconciliation (surfaced by a per-doc audit record,
  missing from the base sweep). Everything else is carried from the sweep, sometimes with corrected status.
- Provenance tags name the source doc(s). Beware the **`#N` namespace collision**: `#19–#123` = the canonical
  `docs/tasks.md` ladder; `#14–#36` (2026-06 era) = the in-session tracker; `#1–#15` = the plutonium backlog;
  `#1–#11` = the dsl-gap-analysis redesigns; `A/B/C/D-n` = persona-open-decisions. Tags below always name the
  namespace.
- Status legend: **still-valid** (verified open) · **partial** (some shipped, remainder named) ·
  **ambiguous** (verify-then-close) · **NEEDS OPERATOR** (decision, not engineering).
- Nothing was silently dropped: every done/stale/superseded item from any source lives in the
  **CULLED ledger** (Appendix A) with one line of evidence.

---

## 1. Movement / protocol / wire

| ID | Item | Provenance | Status |
|---|---|---|---|
| MP-1 | **Per-tick coherent world snapshot** — NPC/player ordered-list rebuild per tick so a routine never reads a half-updated world (cross-substore torn reads; NPC slot desync is WARN-only today) | plutonium backlog **#10**; dsl-synthesis Phase E | still-valid |
| MP-2 | **Appearance (op235) + skipTutorial (op84) decode** + tutorial-graduation/ignore-list bootstrap surface | plutonium backlog **#12**; build-backlog §1; host-bootstrap §5 | still-valid (no decode in `proto/v235/inbound.go`) |
| MP-3 | **Inbound opcode-validity gate + permanent server-reference ISAAC regression test** (`proto/v235/isaac_test.go`) — silent ISAAC desync masquerades as perception corruption; promote the throwaway differential test | plutonium backlog **#14**; plutonium/findings.md; build-backlog §1 | still-valid |
| MP-4 | **Plutonium #15 residue** — `SceneryRecord` lacks a `dir` field for approach geometry (only decodable under 10010); confirm peek-ahead reads live stores end-to-end; per-dir door approach-rect regression test | plutonium findings #15; dimensions.md | ambiguous — bulk done via movement steps 6–10 |
| MP-5 | **10010 custom-wire migration program** — EXECUTION-PLAN Phases 4–6 / steps 14–21 (dialect spine, custom login, pcap golden DECODE/ENCODE, live 10010 session) + its open questions. Recommendation exists (protocol-migration.md: switch); the **decision was never ratified**; zero code (`dialect.go` absent) | plutonium/EXECUTION-PLAN.md; protocol-migration.md | **NEEDS OPERATOR** — gates MP-4/MP-6 and half this section |
| MP-6 | **EXECUTION-PLAN Phase 3 / step 13** — westworld.conf ↔ uranium.conf content alignment + verified `want_pcap_logging` capture recipe (prereq for MP-5's golden corpus; protocol-independent) | EXECUTION-PLAN Phase 3; protocol-and-server-config.md | still-valid, out-of-repo |
| MP-7 | **Cradle defs-browser** — `GET /api/defs` + filterable Defs UI tab ("what id is a bronze pickaxe" as one lookup), incl. its open questions | EXECUTION-PLAN Phase 2 / steps 10–12 | still-valid (`cradle/defs_api.go` absent) |
| MP-8 | **defsgen completion** — emit `prayers_gen`/`spells_gen` (retire `facts/prayers.go` hand literal + vendored `SpellDef.xml` embed — both live drift sources, verified present); `resources_gen.go` (feeds the epistemic "is this minable" layer) + `recipes_gen.go` (production graph); switch the remaining hand-literal consumers (`views_self.go:174` skillIDs map, `views_magic.go` magicSkillID → `facts.SkillID`) | EXECUTION-PLAN Phase 1 steps 6/8/9; plutonium/defs-indexing.md P1/P2 | partial — core codegen shipped `af8b17e` |
| MP-9 | **Movement step-10 LIVE gauntlet completeness** — unit tests exist (`pathfind/overlay_test.go`, `plane_test.go`); `gate_test.go` absent; the cradle `-debug-http` live JSONL gauntlet unverified | plutonium/movement-design.md step 10 | ambiguous — verify-then-close |
| MP-10 | **Low-pri protocol research tails** — XTEA/ISAAC key-derivation confirm; dialog colour-tag format (222); PM icon→rank mapping; plus the long protocol.md open-questions ledger (bank >255 slots, deposit magicNumber, shop stockSensitivity, Huffman tree ref, damage>255 capping, opcode-conflict post-login disambiguation, …) | build-backlog §1; docs/lang/protocol.md §open-questions | still-valid, low |
| MP-11 | **`use(item, sceneryView)` direct dispatch** — the one real runtime code-TODO (`runtime/actions_ambient.go:1157`, op 115 view-target form); coordinate form works | code marker | still-valid, small |
| MP-12 † | **Sleep/wire hardening set** — decode opcode **194** SleepwordIncorrect (latent infinite-resleep trap if server config changes); explicit keep/decode decision on opcode **244** SEND_SLEEP_FATIGUE (still UnknownPacket); session-id 800ms-vs-640ms race hardening; config-robustness guard/log beside `const sleepWord` (correctness depends on server `usingPrerenderedSleepwords`) | plutonium/findings.md recommendations | still-valid |
| MP-13 † | **World-mirror removal aging** — TTL sweep for missed-removal ground-item/scenery records (known limitation, no sweep exists) | docs/world-state.md §known-limitations | still-valid |

## 2. DSL surface, spec & manual

Base namespace: dsl-gap-analysis redesigns **#1–#11** (plutonium family). Done already: #1 (`&&`/`||`), #7a (phantom verbs), the full manual rewrite `b42a52b` — see Appendix A.

| ID | Item | Provenance | Status |
|---|---|---|---|
| DSL-1 | **Gap #2 [P0]** — default-ON `reachable=` filter on every action-intent selector (`nearest_npc`, `world.ground_items.*`, `world.scenery.*`, `world.boundaries.near`, `list.nearest/filter`); only `scan_for` is gated today. Highest-impact behavioral fix ("act on a target behind a wall, loop") | dsl-gap-analysis #2 | **SHIPPED 2026-06-10** (feat/grounding-and-metacog) |
| DSL-2 | **Gap #3 [P0]** — spec = honest superset: add `dsl/spec/accessors.go` rows for `world.boundaries`/`world.scenery`/`world.dialog` (0 rows, verified), `inventory.capacity/used`, `self.equipped.bonuses`, `bank.has/count`, `shop.is_general`, trade `they_/both_` family, + per-entity-view FIELD tables | dsl-gap-analysis #3; dsl-manual-analysis rec #3 | **SHIPPED 2026-06-10** (feat/grounding-and-metacog; bank.has name-resolution stays DSL-22) |
| DSL-3 | **Gap #4 [P1]** — unify perception: one selector shape (`world.<entity>.nearest(pred\|ids=\|names=, reachable=, radius=)`), one return shape; reconcile `scan_for` (bare list) vs `search_map` (Result) | dsl-gap-analysis #4 | still-valid |
| DSL-4 | **Gap #5 [P1]** — `path_distance(target)->Int` + rank "nearest" by real path cost (BFS length computed then discarded today) | dsl-gap-analysis #5; dsl-crossref | still-valid |
| DSL-5 | **Gap #6 [P1]** — `.distance` / `.is_reachable` view fields (composable in lambdas) + `.id` alias on npcView, `.equipped` on itemView | dsl-gap-analysis #6 | still-valid |
| DSL-6 | **Gap #7b/c [P1]** — rename bare-bool `is_reachable`→`can_reach` / oracle `reachable`→`explain_reach`; **validator REJECTS `NotYetImplemented` verbs** (mine/fish/chop/cook still pass validation, verified) | dsl-gap-analysis #7; dsl-manual-analysis rec #5 | still-valid |
| DSL-7 | **Gap #8 [P1]** — `shop.open(shopkeeper)` (absent, verified) + shopkeeper case in `resolveKnownDialogChoice`; `toward=`/`approach=` resolve-or-approach fallback on `interact_at`/`use`/`pick_up`; `inventory.count_any/has_any`; `bank.ensure([[item,qty]])` | dsl-gap-analysis #8; dsl-synthesis C8 | still-valid |
| DSL-8 | **Gap #9 [P2]** — scriptable Path value (`plan_path` → `.next/.step()/.done/.cost/.remaining`) for tolls/quest-gates | dsl-gap-analysis #9 | still-valid, low |
| DSL-9 | **Gap #10 [P2]** — geometry/POI predicates: `in_area("Varrock")` / `at_poi("bank")` / `in_radius_of` / `distance(a,b)` | dsl-gap-analysis #10 | still-valid, low |
| DSL-10 | **Gap #11 residue** — `trade.offer` demands raw Int id pairs, violating the names-everywhere rule | dsl-gap-analysis #11 | still-valid |
| DSL-11 | **Structured outcomes / `do_until`** — fire-and-forget world verbs return Ok(Null) on a miss; add structured action lifecycle + a `do_until` primitive so routines stop scraping `world.messages` for skilling outcomes | plutonium/dsl-usage-best-practices.md | still-valid |
| DSL-12 | **`host.idle_ticks`** self/meta accessor (AutoRune `%IdleC`) — no spec row | build-backlog §9; primitives-backlog | still-valid, low (stall detector covers the runtime need out-of-band) |
| DSL-13 | **#93 (tasks.md) — persona reflex tier**: per-handler `extends host` + `super()` + `host.defaults.<event>`; 3-tier handler stack (persona base reflex → learned specialization → routine override) + `handler_versions` + per-handler punt_count. Design sources: event-study `synth-handlers.md`/`synth-interface.md` + chat-interruption tier ladder. **Now UNBLOCKED** — the persona tier it was parked on exists | tasks.md #93; event-handler-meta §3; event-study synths | still-valid (parser/validator still reject; verified) |
| DSL-14 | **#95 (tasks.md)** — combat-style toggle **live** XP-split verify (read getter shipped `accessors.go:231`; corpus scenario `combat_style_controlled_splits_xp.routine` encodes the assert; live PASS unrecorded) | tasks.md #95 | still-valid (live task) |
| DSL-15 | **DSL testing Tiers 3–4** — property tests (never-panic, determinism, 100ms-cancel) + `go test -fuzz` on lexer/parser (zero `Fuzz*` funcs in dsl/, verified) | docs/dsl.md §testing | still-valid |
| DSL-16 | **Validator-time `IsAdmin` gate** so production hosts reject `command()` at parse time | tasks.md/phases.md Phase 2.7 remainder | still-valid, optional |
| DSL-17 † | **`cmd/specdoc` generator** — regenerate api.md §8 from `dsl/spec/*` "so it can never drift". The doc's own header predicted the hand-tag drift that then happened (a dozen `(to build)` tags now false, internal combat-notes contradictions). Single highest-leverage docs/spec fix | docs/lang/api.md header; api.md audit record | still-valid (no `cmd/specdoc` exists) |
| DSL-18 † | **Runtime-version arrears** — `spec.RuntimeVersion` still `1.0.0` while ~9 surface-changing commits shipped (scan_for, go_to/converse rework `3c0ab33` — arguably MAJOR, C-style aliases, 5 new events, remember/recollect/forget, search_map/survey_map…): perform the overdue bump(s) + CHANGELOG catch-up entries + fix the "224 files" count (now 294/299); **add the spec surface-snapshot test** that fails when builtin/event/accessor tables change without a bump (§4 "planned", now proven necessary) | docs/lang/versioning.md §4; docs/lang/CHANGELOG.md | still-valid — discipline decayed within days of being built |
| DSL-19 † | **Event-vocabulary gaps** (consolidated) — quick fix: flip `damage_taken` `NotYetImplemented:true` in `dsl/spec/events.go:89` (it IS emitted, verified). Surface-only adds: `item_lost` (anti-admin reflex precedent), `npc_chat`/`dialog_opened`, `shop_opened`/`shop_closed`, `shop_stock_update` (only blocker for the shop-restock scenario), `sleep_captcha` (note: cradle auto-answers — decide want). When-predicates: `combat.under_attack` / `attacker_is_player` / `being_followed` / `can_eat`, `self.poisoned`, `hp_below` safety watcher (the dsl_events.go header names a watcher that doesn't exist). Qualifiers: `increases/decreases/added/removed` + "by N+" thresholds (AST has only becomes_true/false/changes) | docs/lang/events.md §planned; event-study synth-events/survey-wiki-mechanics; scenario-candidates D-15; reference/diagrams/s6-ambush | still-valid |
| DSL-20 † | **Perception-field gaps** — `GroundItem.quantity` (world/world.go:50 has no Amount, verified) + `.is_mine` loot ownership; friend-list packet mirror to de-stub `player.is_friend` (always-false stub, verified views_world.go:409); per-player `in_combat_with` tracking; `self.prayer.available`; frozen trade-event param renames (`other_index`→`other: Player`) | api.md corrections + tag drift; lang/state.md; build-backlog §10 | still-valid |
| DSL-21 † | **`examine(target)` verb** — never built (runtime/examine.go is the gazetteer, not the verb; no spec entry; def-side only, no wire work needed) | build-backlog §4 | still-valid, small |
| DSL-22 † | **Fix `bank.has`/`bank.count` to resolve item names** — Int-only today (`views_bank.go:78` `interp.AsInt`, no `resolveItemID`), violating names-everywhere | dsl-manual-analysis rec #2 | still-valid |
| DSL-23 † | **Manual prose-block validation as a unit test** — parse+validate every code block in `dslManualBase` beside `dsl/spec/consistency_test.go` (the rewrite's harness was one-off; the manual taught a guaranteed parse error for weeks once already) | dsl-manual-analysis rec #4 | still-valid |
| DSL-24 † | **Relational/reputation DSL reads** — `trust()`, `trust_confidence()`, `reputation()`, `is_rival/is_ally/is_stranger`, `rel.tag`/`rel.has_tag` (zero spec rows; the ledger exists, only `relation_with` reads it) | lang/actions.md Phase-4 stdlib; social-graph-and-trust-ledger §9 | still-valid (pairs with C-24) |
| DSL-25 † | **bounds `{}` deferred set** — polygon shape; dynamic (re-evaluated) bounds; entity-relative bounds ("within 4 of alex"); scenery-resolved `near(scenery=…, radius=…)`; validator warning for bound handlers ignoring their (x,y) params | lang/proposals/bounds-scope.md (shipped core; these deferred) | still-valid, low |
| DSL-26 † | **Syntax niceties** — multi-arg lambdas `(a,b) =>`; duration suffixes in `repeat`/`wait` (done for `select`); docstring convention; `import "lib/…"` namespacing | docs/lang/syntax.md | still-valid, low |
| DSL-27 † | **REPL niceties** — multi-line brace-depth input; `-repl-on-fail` + `.resume`/`.cancel`; `.routines`; `.watchers` (blocker note stale — when-watchers exist, only the meta-command is missing); `.log <level>`; readline + history file | docs/lang/repl.md | still-valid, low — daemon-era `/eval` covers much of the need; consider archive-grade |

## 3. Cognition

| ID | Item | Provenance | Status |
|---|---|---|---|
| C-1 | **#30/#30a (in-session tracker) — promotion-on-progress gate**: `HybridDirector` promotes on `last.OK() && !OneShot` only (`hybrid_director.go:116-117`, verified) — a completed-but-zero-progress routine still gets learned. Gate `lib.Promote` on the already-computed `progressKey` delta; bigger #30: classify errored (broken→re-author) vs completes-no-progress (wrong tool→block goal) | SESSION-STATE-2026-06-09 §5 | **DONE 2026-06-10** — in main via 7945ee1 |
| C-2 | **#33 — confidence-scaled routine commitment**: derive confidence from `failStreak+worldStall+spinCount+progress`, feed the Act prompt (LOW → one small step then reassess). Ties decision-brain-tiering R2 (self-rated confidence is uncalibrated — log vs outcomes) | SESSION-STATE §5; reference/decision-brain-tiering R2 | **SHIPPED 2026-06-10** (feat/grounding-and-metacog; calibration analysis itself still open) |
| C-3 | **#32 — stuck-breaker with distilled CONTEXT**: on STALL, same-model call with HOST-precomputed reachability conclusions + knowledge/goals/decision-history/failed-routines-with-why; output = correct plan or unachievable-here→abandon. Trigger shipped; rich payload not. Much "stuck" was the lying grid (now fixed) — re-evaluate need after a soak | SESSION-STATE §5 | **SHIPPED 2026-06-10** (feat/grounding-and-metacog) |
| C-4 | **#27 — analysis-console REFLECT kind**: mesad analysisSystem has no persona-voiced REFLECT path for abstract operator questions (verified, `act.go:408` flat interpreter unchanged) | SESSION-STATE §5 | **SHIPPED 2026-06-10** (feat/grounding-and-metacog) |
| C-5 | **Phase 5b-4 — forage termination**: forage-local box-escape exhaustion counter (count `source-spent:place:*` tags, NOT `d.spinCount`) + global give-up `forage-exhausted` tag past cap K → goal ABANDON. ARM-1 otherwise bounded only by "untried reachable POIs" | phase-5b-1-impl §13; phase-5b-revision-backlog P1 #5 | **SHIPPED 2026-06-10** (feat/grounding-and-metacog) |
| C-6 | **5b-6 dials** — box-escape radius / nearest-vs-likely-town tuning + RAG-only fallback; left to the 200-host load test | phase-5b-foraging-plan §10 | still-valid, gated on O-5 |
| C-7 | **ARM-2 trust-decay for verifiable-but-wrong tips** (lie → real POI passes verification → wasted detour) + confirm the load test co-locates hosts so the player-oracle path is exercised | phase-5b-revision-backlog P1 #7 | still-valid |
| C-8 | **ARM-3 — `ProposeSources` LLM RPC** (Haiku-only, safe-empty degrade; hypotheses-to-test, never plans). Includes revision #10 (proposeCooldown durability), moot until this ships | phase-5b-foraging-plan; w-k-l §6 | still-valid, optional/droppable |
| C-9 | **Goal-graph-aware Act loop** — conductor/director traverses the graph, spawns sub-goals, detects dead ends, instead of one flat goal ("the deepest refactor") | w-k-l §9 cross-cutting | still-valid (large) |
| C-10 | **Observation/sync at 200 hosts** — knowledge + goal-graph sync batching/backpressure | w-k-l §9 cross-cutting | still-valid |
| C-11 | **Full interrupt ladder** — remaining: directedness classification beyond the keyword ladder, full T0–T3 tier set, committed-region model, two-phase ORIENT→REPLY as DSL-level machinery (detour tiers/pause/reactive tier/displacement arbiter shipped; `detour.go:125` marks the seam) | cognition-and-autonomy §7.4; _research/chat-interruption §8 | partial |
| C-12 | **Progressive disclosure (G2)** — `Tier=core` on dsl/spec, `bot_vocabulary` earned ledger, `DSLSurface` assembly, `ERR_VERB_NOT_LEARNED`, + §9 calibration questions. Never built (zero code, verified). **Direct tension with shipped reality**: the manual now teaches the full surface, and the dsl-manual analysis proved "prose presence predicts usage" (full disclosure is what made hosts competent) | SESSION-STATE.md next-steps #4; _research/host-bootstrap-and-knowledge-gating.md; brain.md earned-DSLSurface | **NEEDS OPERATOR** — keep/kill; research-goals §1a may still want it as a cohort experiment |
| C-13 | **Event-handler open design questions** — idle/tick scheduler home; armed-predicate cost gating; specialization promotion threshold; rollback hysteresis; fold `message` into `server_message`; most-specific-match tie-break | event-handler-meta-analysis §6 | still-valid (activates with DSL-13) |
| C-14 | **cognition design opens** — retrieval query generation (heuristic vs LLM expansion); cross-source prompt budgeting; dense-area summarization/truncation (questions-and-decisions D6) | cognition.md §open; questions-and-decisions D6 | still-valid (design) |
| C-15 | **memory design opens** — cross-host relational asymmetry policy (de-facto answered "asymmetric is fine" by AuthLocal — ratify); recall ranking beyond pure cosine (similarity × recency × salience, weights logged per call) | memory.md §open; knowledge-pipeline §7; reference/decision-memory-mesa-tiers | still-valid (research/tuning) |
| C-16 | **brain design opens** — persona-drift audit cadence (PersonaAudit never built); cost-cap behavior (blocked on O-9 ledger); rate-limit tier fallback (ties O-10) | brain.md §open | still-valid (design) |
| C-17 | **w-k-l §6 "still open"** — cron cadence/trigger mix; sync-escalation build-vs-defer; ledger GC/decay specifics; persona dials (`BulkApperception`, `Curiosity.*`) feeding the escalation threshold | world-knowledge-and-learning §6 | still-valid (design) |
| C-18 | **Quick-wins residue** — reinforce autonomous looping routines; cache persona block (ties P-1 A13) | cognition-and-autonomy §7.6 | partial |
| C-19 † | **Genesis remainder** — `GenesisResult` carries only Goal+Mood+KeywordLadder; the designed HANDLER SET / pearl-policy deltas / routine-library picks + delivery via Provision directives never shipped (all five Directive kinds exist in proto, unused for this) | cognition-and-autonomy §2 vs mesa/client/client.go:392 | still-valid |
| C-20 † | **Observation novelty/curiosity gate** — replace the static `observationFloor=0.3` stub (`runtime/observation.go:17-22`, comment says "novelty/curiosity gate replaces this constant later"); plus host-owned **proactive ask clock** (ask cadence is Act-turn-driven today, contradicting the design doc's claim) | cognition-audit-findings end-notes | still-valid |
| C-21 † | **Ask/teach deferred set** (w-k-l §3.8a) — drive real NPC dialog-tree menus for asks; unprompted first-class TEACH; LLM-judged fuzzy closure + cron re-arm of ask-exhausted + ask reformulation; multi-turn follow-ups / multi-target coordination; trust-weighted target choice (3b relationships shipped — wiring into target choice unverified) | world-knowledge-and-learning §3.8a deferred | still-valid |
| C-22 † | **Forage residual-risk watchlist** (shipped-code defect watch, tracked nowhere else) — R1 destinationLabel collision over-suppression (re-key tags on SnapX|SnapY if lossy); R3 5b-6 cron re-arms `source-spent` on restock; R5 per-Q cursor map before 5b-5 multi-source; R7 check interrupts-queue before burning lastAsk floor | phase-5b-23-impl §7 | still-valid (watch items) |
| C-23 † | **At-least-once path for highest-importance events** (Death/Scammed) past the lossy bus — flagged twice in the decision records, never resolved; consider deriving grievance redundantly from the Death diff | reference/decision-cognitive-loop §9; decision-cradle-modifications §5 | still-valid |
| C-25 † | **Chat-reply knowledge slice** — `mesad.Chat` (act.go Chat RPC) answers player chat from persona prose + recent lines ONLY, under an explicit never-bluff order — so a host honestly denies knowing mechanics its own ledger holds at confidence 1.0 (live exhibit 2026-06-10: bernard "not familiar with 'prospect'" while his ledger carried "The examined rock contains tin ore" ×22 beliefs; vocabulary mismatch: server messages say "examine", UI label says "Prospect"). Fix: host attaches top-K ledger subjects keyword-matched against the incoming message to ChatTurn (proto already carries Recent; same shape), so the grounding instruction has facts to ground in. Optional: harvest the verb LABEL alongside message-vocabulary subjects. | operator exhibit 2026-06-10 | still-valid |
| C-26 † | **Dialog trees as first-class knowledge sources — the smithing-anvil exemplar.** A multi-level server menu IS the curriculum (operator: "It asks what category, then what specifically, and if you pick something you don't have the level for the server tells you"). Perception already exists (`world.dialog.is_open/.options/.find_option/.answer`, views_world.go; options rendered in situation hints) and `npc_dialog`/`server_msg` are observation-capture kinds — but nothing records the MENU STRUCTURE as durable knowledge. Build: (a) harvest opened option-sets as observed claims bound to the invoking CONTEXT subject (station+item: "anvil + bronze bar → dagger/sword/axe/…"), not transient hints; (b) bind level-rejection facts to the ITEM subject ("mithril dagger requires smithing 50") — mind the C-25 vocabulary-binding lesson; (c) manual idiom: a new menu is a free lesson — read every option, note what you learn, then choose; a rejection reason is knowledge, remember it; (d) `dialog_opened(options)` event (fold from DSL-19). Design tension to honor: MP-8 `recipes_gen` would SEED the production graph omnisciently — discover-don't-seed says hosts should learn it from these trees in-world; recipes_gen stays validation tooling. | operator design note 2026-06-10; pairs C-21/C-25/DSL-19 | still-valid |
| C-24 † | **Trust-ledger remainder** — decide the prior: shipped uniform Beta(1,1) vs designed H-shaped α₀=2+4H (never plumbed); daily decay-toward-prior; LLM TrustGrade overlay tier + TRUST_UPDATE directive (`ledger.go:144` punts to "Phase-4 LLM overlay"); insult/hostile-chat sentiment classification for negative deltas. DSL read surface = DSL-24 | social-graph-and-trust-ledger §9; reference/decision-appraisal-affect §5; memory.md | still-valid — prior question is **NEEDS OPERATOR** |

## 4. Persona & reveries

| ID | Item | Provenance | Status |
|---|---|---|---|
| P-1 | **Open decisions A2–A15** (A1 answered: cohort = launch batch, archetype = personality; A2 needs re-think in its light). Top-3 unblockers per the worksheet: **A4 compiler causality path (quirks-first vs EventPolicy templates)**, **A5 quirk Bucket-C closed registries**; then A3 band cut-points, A6 emergent-quirk promotion, A7 revision triggers, A8/A9 seeding, A10 typo/voice, A11 mood→prompt table, A12 curiosity/attention formula home, A13 cache-prefix, A14 prefs dormancy, A15 trust projection | _research/persona-open-decisions.md §A | **NEEDS OPERATOR** |
| P-2 | **Authoring B1–B5** (deferred paired session) — archetype prototype table (~13); cohorts.yaml weights/forbidden pairs/name pools (reframed per A1); north-star catalog + conflict rules; **reverie catalog** (enums.go explicitly defers it; validate every gesture for server-visibility first); 500-host population mix | persona-open-decisions §B; reveries.md; personas.md; reference/digest-claude-persona | **NEEDS OPERATOR** (paired session) |
| P-3 | **Build backlog C residue** — offline **genpop sampler** (priors → rejection → quirks → seal+hash; external mixture-model sampling becomes necessary at Phase-7 scale); cognition.Bundle Card/Pinned/Affect cache-prefix re-inject; drift update rules (κ/ρ, Fleeson state); fill reserved `mood()`/`motivation()` builtins (still NotYetImplemented, verified); `reveries/` package | persona-open-decisions §C; claude-persona-implementation | partial — remainder still-valid |
| P-4 | **Code TODOs in persona/** — `compile.go:33` TODO(band→policy session): compile `EventPolicy`/`ChoiceWeights`/`Directives`/`ReverieKernel` into PersonaCard (only quirks+thresholds compile today); `policy.go:192` soft directives → biases | code markers | still-valid — concrete form of P-1 A4 |
| P-5 | **Doc-consistency E items** — fix reference-YAML Appendix A per A1; audience-tag (LLM/CODE) pass + pin Bucket-C registries; preseed Tier 0/1/2 into the YAML; remaining quantitative-persona-models §10 GAP items | persona-open-decisions §E | still-valid (cheap, some gated on P-1) |
| P-6 | **Phase 5 — Reveries (full build)** — engine as a **sibling-goroutine scheduler driving public effector methods with a busy-flag** (the decision record REFUTES docs/reveries.md's `reverie.tick()` interpreter hook — merge that correction when this starts); catalog ~15–20 (server-visible only); trait-derived weights + per-host jitter (ReverieSeed shipped; engine not); weight drift + REVERIE_REBASELINE cron (directive kind reserved, no cron); cross-host reverie recognition; the jitter/naturalism set (per-host interpreter seed — interp is fixed `0xC0DE_BABE` today, npc-pick shuffle, corner skew, settling pauses) | phases.md Phase 5; reveries.md; reference/decision-reverie-believability; primitives-backlog §jitter | still-valid — not started |
| P-7 | Parked (keep parked): gossip/reputation propagation; full prospect-theory/copula forms; expression_drift | persona-open-decisions §D | parked by design |
| P-8 † | **Seal integrity** — `cornerstone_hash = sha256(canonical_json(cornerstone))` (exists only as a comment, `persona/policy.go:262`); mesa-side guard rejecting cornerstone UPDATE without a persona_revisions row (no such table); optional automated regen-below-threshold loop in personacook (operator-in-the-loop today) | persona-compile.md §6; reference/decision-lifecycle-identity-sampling | still-valid |
| P-9 † | **ActionArbiter + real `is_busy`** via a per-host current-action registry — Phase-5 prerequisite (reverie emission must be gated on effector ownership) and independently useful to routines (`is_busy` is a hard-stubbed false, verified `views_self.go:88-95`) | reference/diagrams/s7-reverie | still-valid |

## 5. Mesa

| ID | Item | Provenance | Status |
|---|---|---|---|
| M-1 | **Knowledge-pipeline Stage 1** — worldbrief generator (RAG-query-from-persona + Opus, `game_familiarity`-scaled); chain Stage 1→Stage 2; `game_familiarity` field (persona field vs runtime knob — open); `WorldBrief` struct → generated-prose-artifact refactor; seed→memory initialization (no Pinned/worldbrief seeding into episodes, verified) | docs/mesa/knowledge-pipeline.md §6 | still-valid (Stage 2 cook + per-host RAG built) |
| M-2 | **Knowledge-pipeline opens** — misconception dial; Stage-1 query derivation mapping; recall ranking weights (current impl is pure cosine); NULL-embedding backlog alerting once embedding is on the live path | knowledge-pipeline §7; reference/decision-memory-mesa-tiers | still-valid (design/tuning) |
| M-3 | **Admin control plane v2** — send live `PERSONA_REVISION` from `putOnePersona` over `Provision.Subscribe`. **Now cheap**: the per-host subscriber registry it was blocked on EXISTS (built for goal push, `server.go:609-662`); host-side handler is logs-only (`runhost_bootstrap.go:574`). Includes pearl/policy recompile-on-revision (today: restart-only) | docs/mesa/admin-control-plane.md §phasing; recheck | still-valid — unblocked |
| M-4 | **Admin control plane v3 remainder** — `mesa-ctl hosts` (live roster), `persona recook`, persona history/rollback (+ optional persona_history table), `genesis <id>` via ctl. *(Goal push + fleet gen: DONE — see Appendix A)* | admin-control-plane §phasing | still-valid |
| M-5 | **RSC account provisioning for arbitrary batch names** — hard external prerequisite (drone1..200 exist; arbitrary persona names don't); no account-creation tooling in `cmd/orsc-ctl` (verified) | admin-control-plane open Q4 | **NEEDS OPERATOR / external** |
| M-6 | Admin-plane minor opens — mTLS later (env token now, fail-closed: fine); `ListPersonas` paging when it bites (no cursor, verified `admin.go:88-105`) | admin-control-plane Q1–Q3 | still-valid, lean |
| M-7 | **mesa.md opens** — persona JSON schema migration policy (jsonb, no migration story); backup/restore (`pg_dump`, trivial); cross-host routine-influence detection (research aspiration) | mesa.md §open | still-valid (design/research) |
| M-8 | **AutoRune script-corpus ingest** (Phase 2.6 item 3) — never done; the w-k-l "discover, don't seed" decision forbids it as live host knowledge; dev-namespace gate exists unused | tasks.md/phases.md 2.6; lang/knowledge.md slice 3 | ambiguous — repurpose as OFFLINE reference or drop |
| M-9 | **`-knowledge-query` corpus sanity CLI** (Phase 2.6 item 5) — never built; `mesa/wikirag` CLI partially covers | same; lang/knowledge.md slice 4 | ambiguous — fold-or-drop |
| M-10 | **Hot-swappable prompt registry** — kill the mesad-redeploy-per-prompt-tweak loop. Named prompts (dslManualBase prose, genesis, decide, analysis system) each = compiled-in default + optional live override: Postgres `prompt_overrides` table (name/body/version/set_by/set_at + history), `atomic.Pointer` snapshot reads, `Admin.{Get,Set,Reset,List}Prompt` RPCs, `mesa-ctl prompt get/set/diff/reset/history`. SetPrompt parses+validates every DSL code block in a candidate manual IN-PROCESS before the atomic swap (a manual that teaches a parse error becomes uninstallable — closes DSL-23 as a side effect). The spec-generated reference appendix stays compiled-in (must match the running engine). Hot swap = one prompt-cache invalidation, same as a restart. Later, same pattern for model ids / cron knobs / RPC timeouts (`Admin.SetConfig`). | operator request 2026-06-10 (manual-prose iteration cadence: "we do a lot of tweaking to how we tell the LLMs to write routines") | still-valid |

## 6. Control-plane / fleet ops / runtime robustness / observability

| ID | Item | Provenance | Status |
|---|---|---|---|
| O-1 | **Leak-audit deferred #1** — `routineCtx` per-view binding (bug #3): detour park/resume poisons resumed namespace verbs with a cancelled ctx; fires on every fatigue nap (`dsl_bridge.go:253` write / `namespace_actions.go:36` read, verified unchanged). Land with `-race` tests | runtime-leak-audit-2026-06-10 §deferred | **DONE 2026-06-10** — in main via 7945ee1 |
| O-2 | **Leak-audit deferred #2** — Pause/ANALYSIS dual-cancel of in-flight detours ("frozen" host keeps acting ≤3min). Trap: do NOT re-parent dctx to turnCtx (2m turn deadline kills 3m naps) | same | **DONE 2026-06-10** — in main via 7945ee1 |
| O-3 | **Leak-audit deferred #3** — size-capped JSONL rotation for the debughttp event log + `decisions.jsonl` (allowlist stopgap landed; rotation didn't) | same | still-valid |
| O-4 | **Leak-audit deferred #4** — `limbic.Ledger` entry cap + inclusion in `pruneCognition` (covers knowledge+goalGraph only, verified); watch the `ledger.relationships` gauge | same | still-valid |
| O-5 | **The 200-drone soak / load test** — instrumentation all landed (bus `*` gauge, pprof, telemetry); the RUN + verdicts are the open work. Soak matrix must include one never-restarting host (supervisor restarts MASK leaks) + a crash-looper; acceptance: bus-subscriber delta == 0 over any steady-state hour. Also exercises C-6/C-7 dials, ARM-2 co-location, the 2×N `*`-subscriber fan-out cost, and Postgres write-throughput / pgvector HNSW insert latency at fleet scale | leak-audit instrumentation plan; cradle memory #9; reference/decision-transport-storage | still-valid — standing milestone |
| O-6 | **#20 (in-session) — drone state reset** — orsc-ctl client done; needs the Codex-side OpenRSC admin API + drones online | SESSION-STATE §5 | ambiguous — verify with operator |
| O-7 | **Phase 6 Delos residue** — cohort/archetype analytics (re-scope post cohort-reframe: per-archetype over the mesa metrics table), cost dashboard, perturbation toolkit (drop rare item, pause cohorts — formal capability beyond admin `::commands`); observability minor opens (historic prompt retention policy) | phases.md Phase 6; observability.md; research-goals technician § | still-valid — much intent pre-delivered by the cradle UI; re-scope before building |
| O-8 | **Server/ops decisions** — multiple worlds per experiment; production F2P flip (`member_world:false` + gate-loc confinement re-verify); public-exposure limit re-enable; perf tuning at 500 hosts; production launch boundary (world reset + non-admin account re-mint at Phase 7) | server-config.md §open; development-workflow launch boundary | still-valid (ops decisions) |
| O-9 † | **LLM cost/token ledger** — NOTHING tracks tokens or dollars today (no brain_calls table, no usage accounting anywhere, verified): per-call model/tokens/cost/latency rows + per-host budget surface. "We need to know exactly where the money is going" is still unmet; C-16's cost-cap question is blocked on this | observability.md; brain.md §cost; reference/digest-claude-arch | still-valid |
| O-10 † | **Fleet LLM throttling/degradation** — bounded semaphore/worker pool in front of mesad LLM calls sized to account TPM (no throttle found, verified); ctx-deadline degrade to local default; Sonnet→Haiku load-shed fallback. Becomes a verdict item of O-5 | reference/diagrams/mr6-deployment; decision-brain-tiering R4 | still-valid |
| O-11 † | **Two leftover hardening fixes from the leak-audit body** — bounded duration/success-return for `Follow` (behavioral half of leak #7); 2-minute `WithTimeout` on ANALYSIS `runCommand` | runtime-leak-audit body (fix #5 + leak #7) | still-valid (not evidenced as landed) |
| O-12 † | **Conf drift — actually apply the fix** — establish the `inc/westworld.conf` → `~/Code/openrsc/server/westworld.conf` symlink (chartered, never executed; drift has now recurred TWICE — 2026-06-09 deployed `want_runecraft: false` vs repo `true`, verified non-symlink + diff); reconcile want_runecraft and update server-config.md's diff table | docs/agents/openrsc-steward.md first task; server-config.md | still-valid — proven necessary |
| O-13 † | **Pearl ops remainder** — Hit/Miss/Veto telemetry + hit-rate dashboard (hit-rate is THE fleet cost lever; not evidenced); veto-recoverable vs hard-abort contract (operator veto ⇒ abort, quirk veto ⇒ recoverable); stricter Gate for brain-authored fragments (Facts source flag); benchmark pearlFacts on a worst-case combat routine | reference/decision-pearl-local-core; decision-policy-quirk-engine | still-valid |
| O-14 † | **Render lighting polish vs RSCPlus** — optional, human-in-the-loop tune against reference screenshots; castle-door re-open recipe only if it recurs | render-engine.md §7 | still-valid, optional/low |

## 7. Docs & cleanup

| ID | Item | Provenance | Status |
|---|---|---|---|
| D-1 | Fix **stale Phase-2.6 claims** — `docs/phases.md:181` + `docs/tasks.md:198` + `docs/state.md` (×3) say "NOT STARTED / recall() remains the stub", disproven by `ccbc220` + mesa Recall/wikirag/LTM | decoder §16.1; phases/tasks/state audit records | still-valid |
| D-2 | Re-baseline **`docs/phases.md`** forward plan (anchored `fd0731c`): Phase 3 substantially delivered out-of-order (bots→hosts rename MOOT — schema born host_id); Phase 4 leftovers landed; Phase 6 partially pre-delivered; extend the what-actually-happened diagram past `fd0731c`; 196→263 scenario count | phases.md audit record | still-valid |
| D-3 | Full re-snapshot of **`docs/state.md`** (banner's own "Outstanding" list is itself done; repo tree missing ~14 packages) + **`cognition-and-autonomy.md` §3/§6/§7** (every listed "gap" except committed-region/directedness has shipped) | state.md / cognition-and-autonomy audit records | still-valid |
| D-4 | Retire/annotate **`.claude/HANDOFF.md`** (2026-05-31 render handoff, dead workstream); confirm render-port-plan.md's SUPERSEDED banner remains the only live pointer | sweep | still-valid |
| D-5 | Mark **SESSION-STATE-2026-06-09** superseded (its IMMEDIATE NEXT TASK + P0 queue are done; hashes rebased) — §5 opens absorbed here as C-1..C-4/O-6; same banner treatment for **SESSION-STATE.md** (2026-06-02: "Pearl REJECTED" reversed, "nothing built" inverted) | both SESSION-STATE audit records | still-valid |
| D-6 | **`facts/prayers.go:5` stale comment** ("OpenRSC adds Protect from Melee at 14" — false for this build; verified present) | build-backlog §8 | still-valid (one line) |
| D-7 | **primitives-backlog.md + tasks.md checkmark fix** — archive primitives-backlog (per audit verdict) AND fix the duplicated **false checkmarks** in `docs/tasks.md:102-103` (#89 `by_type` / #90 `last_attacked_*` claim surfaces that never existed in any commit; shipped as list `.filter/.random` + `combat.last_npc/last_player`) | primitives-backlog audit + recheck | still-valid |
| D-8 | knowledge-pipeline §7 last bullet — merge settled parts into canonical `memory.md` + `persona-compile.md` | knowledge-pipeline §7 | still-valid |
| D-9 | Drain-then-archive the **temporary worksheets**: `persona-open-decisions.md` (keep A1's verbatim answer as primary source), `phase-5b-revision-backlog.md`, `recovered-scope/*.json` (all four closed) | the docs themselves | still-valid |
| D-10 | **lang/state.md "Open questions for tomorrow"** — all four de-facto settled by the 2.9 freeze; answer-or-strike | lang/state.md:435 | stale-but-unmarked — close out |
| D-11 | Reconcile **build-backlog §10** spec/impl drift list against the frozen surface (most closed in #115; verify `GroundItem.quantity` → DSL-20, trade phase-enum names → done, `find_all` → done) | build-backlog §10 | ambiguous — verify-then-close |
| D-12 † | **Execute the docs-grounding audit verdicts** (the umbrella for this whole cleanup): per-doc rewrites (index.md matrix, architecture.md, brain/cognition/mesa/personas/reveries banners, scenarios.md §8/§9/§11 inversion, server-config DB section sqlite→MySQL, protocol.md port + ground-item decode, world-state self-index/sprite-map fixes, events.md catalog claims, api.md tags or DSL-17, README/repl/development-workflow path fixes); archives → `docs/archive/initial-brainstorming/` (memory.md w/ fresh replacement, observability.md → short rewrite per recheck, render-port-plan, primitives-backlog, scenario-candidates, gemini/gpt55 persona+arch external reports per their verdicts, dev-partner charter — **fix the 7+ inbound links in the same change**); promote gitignored-but-code-cited specs to tracked (`decision-policy-quirk-engine.md` — cited by `pearl/doc.go:24`; `decision-persona-schema.md` — linked by tracked persona docs); promote movement-design wire-facts + trade/duel FSMs into tracked protocol/world-state docs; write `docs/lessons-learned/` chapters from the corpus's lessonsMaterial; comment-rewrite pass per the runtime comment-audit guidance | docs-audit corpus (verdicts + rechecks) | still-valid — the audit's own action plan |
| D-13 † | **Delete `docs/_research/reference.bak/`** wholesale (~36 files: byte-identical or rename-only pre-Limbic snapshot of `reference/`; one exception: move `reference.bak/REVIEW-GUIDE.md` → `reference/REVIEW-GUIDE.v2.md` first — it is the only surviving v2 explanation). **Delete `docs/_research/cognitive-architecture-design-v2-backup.md`** (byte-identical to `reference/cognitive-architecture-design.v2-rejected.md`, cmp-verified) | reference.bak audit records | still-valid |
| D-14 † | Trivia ticks — fill cognition-audit-findings "Audit date" placeholder + tick its now-closed end-notes (mindRel Affinity/Grievance exported); fix CHANGELOG "224 files" → 294/299 (folds into DSL-18); `worldmap.go:127` stray "Plane 0 only in v1" comment (code loops 4 planes) | cognition-audit-findings; CHANGELOG; movement-design records | still-valid (minutes) |

## 8. Content / scenarios

| ID | Item | Provenance | Status |
|---|---|---|---|
| S-1 | **Phase 2.8 run-to-ground campaign** — open-ended corpus-vs-live sweep; 4 known engine gaps logged (confirmEngaged kiting; melee food-lock r3; admin death-drop bypass; +1) | phases.md 2.8 | still-valid (mature maintenance) |
| S-2 | **Merge `scenarios_proposed.yaml` (47) + `scenarios_bots_proposed.yaml` (60)** — verified-but-unmerged, old embedded-code schema (scenarios.md §3 retained for the conversion); re-verify ids vs defs; mind the trailing NEEDS-VERB blocks | scenarios.md file table | still-valid (content) |
| S-3 | **Corpus-mined gap-finder candidates** — the few genuinely-unreplicated shapes from scenario-candidates: completed two-sided item SWAP; staked-duel payout assertion; multi-cycle cook-bank; event-driven mine-on-respawn; shop-restock-wait-buy (blocked on `shop_stock_update`, DSL-19) | lang/scenario-candidates.md (archived); recheck | still-valid (content; most of the 22 superseded — see Appendix A) |
| S-4 | Phase 2.8 validation bar (3+ F2P quests E2E, all skills to 20, 30-min tour) — never formally recorded; superseded in spirit by autonomous play | phases.md 2.8 | ambiguous — **NEEDS OPERATOR**: does the bar still matter? |
| S-5 † | **Scenario-catalog CI check** — parse scenarios.yaml and validate referenced ids/coords against facts (now cheap: defs are code-genned + TestStaticParity) | scenarios.md §11 | still-valid |

## 9. Research goals & experiments (long-horizon, keep visible)

| ID | Item | Provenance |
|---|---|---|
| R-1 | **Phase 4.5** — wiki-RAG cohort experiment (wiki_enabled vs no_wiki, ~1 week) — not started; substrate (cohort_id, corpus) exists | phases.md; questions-and-decisions E5 |
| R-2 | **Phase 7** — formal scale waves 7a 10 / 7b 100 / 7c 500 (ad-hoc 200-drone tests don't count) | phases.md |
| R-3 | **Phase 8+** — long-horizon accomplishment, community formation, ethics-without-observers, cohort experiments, inter-host learning; **punt-rate instrumentation** (the named learning observable — zero code) | phases.md; research-goals.md |
| R-4 | **5-Properties / success-metrics harness** — skill/wealth checkpoints, network-analysis vs null model, categorical ethics incident logging, detection-survival rate; persona flattening battery (population entropy, pairwise JSD, trait re-administration CI) — nothing instrumented | research-goals.md; gpt55/claude persona reports |

---

## Operator-decision roll-up (blocking calls, in one place)

1. **MP-5** — ratify or reject the 235→10010 migration (gates MP-4/MP-6 and the §1 back-half).
2. **C-12** — progressive disclosure (G2): keep as a Phase-4.5-style cohort experiment, or kill (full-disclosure manual won in practice).
3. **C-24** — trust prior: ratify uniform Beta(1,1) or implement the designed α₀=2+4H.
4. **P-1 (A4+A5) / P-2** — the persona paired sessions; A4/A5 block `compile.go:33`.
5. **M-5** — who/what mints RSC accounts for arbitrary persona names.
6. **M-8 / M-9** — AutoRune ingest + knowledge-query CLI: fold or drop.
7. **S-4** — does the Phase-2.8 validation bar still matter post-autonomous-play?
8. **DSL-27** — REPL feature set: keep or declare archive-grade (debug-http `/eval` era).

## Top of the queue (if priority is wanted)

O-1/O-2 (correctness bugs firing nightly) → C-1 (quick win, exact fix located) → C-5 (forage termination, the open obligation) → DSL-1/DSL-2 (highest-impact authoring fixes) + DSL-17/DSL-18 (stop the two proven drift engines: hand-tagged reference + unbumped runtime version) → O-5 (the soak that validates everything, with O-9/O-10 riding along) → P-1 A4/A5 + M-5 (operator calls blocking persona code and fleet naming) → MP-5 decision.

---

# Appendix A — CULLED ledger (done / stale / superseded; nothing silently disappears)

One line + evidence each. Grouped by family.

## Done — master ladder & body (tasks.md #19–#123 namespace)

- tasks **#19–#92, #94, #85, #114–#123** — done; commit trail in tasks.md (`93aad62` DSL 9 steps; `38ef5a0` api freeze; `18ac18b` + r3 rounds).
- Phase 2.7 admin actions — done differently via generic `command()` builtin (`b876c2d`); dedicated `admin_*` verbs never needed.
- Render engine + spectator — shipped (`render/orsc`; classic rasteriser deleted `76b1a48`; cmd/rendertest deleted with it).
- `combat.style` read getter (#117) — done (`dsl/spec/accessors.go:231`).
- `resolve()` + alias store (#120) — done (`dsl/spec/actions.go:448`, `cognition/resolve`).
- Shop namespace (api.md "(to build)") — done (`accessors.go:163-170`, `runtime/{shop,actions_shop,views_shop,shop_register}.go`).
- `magic.cast` unification, `bank.deposit_all`/`withdraw_all`, `combat.retreat` (RETREAT_TOO_EARLY), duel event family, `self.position.plane`, `is_sleeping` de-stub, `find_any` — all done (api.md tag drift, each verified in spec/views).
- trade exemplar naming (`trade.accepted` family) — done (`accessors.go:177`).
- Opcode-234 decode + bitpack research (build-backlog §1) — done (`proto/v235/updateplayers.go`; combat faculties de-stubbed exactly as predicted).
- Boundary dir-byte enum + opcode 50/135 disambiguation — RESOLVED in-doc and verified in `action/use.go:20` / `action/magic.go:14`.
- build-backlog §8 content table (bucket=21, gold-ore=152, def-655, pottery wheel, Druidic gate, protect-from-missiles rename, iron-scimitar `item 170 2` fix) — all applied (scenario files verified).
- `bury(item)` — covered by use/eat's inventory-command path (`runtime/items.go:66`); closed as done-differently.
- fatigue→sleep design questions (captcha; sleep-as-body-faculty) — answered & built (cradle auto-answer, `outSleepWord=45`, fatigue detour + arbiter).
- writing-routines §10: wipeinv mirror (`world/inventory.go:45-57` shift-delete), shop verbs, `self.position.plane` — all fixed.
- protocol.md opens now closed: ground-item 0xFF = multi-entry `[0xFF][offX][offY]` clear (`proto/v235/ground_item_test.go` — the hidden Barbarian-Village pickaxe); GROUND_ITEM_TAKE=247 (`action/interaction.go`); boundary 14/127 verified (`action/boundary.go`); heartbeat 5s live-validated; 8-byte session id; full 191 decode.
- examine wire-opcode question — moot (examine info is def-side, `runtime/examine.go`; the *verb* remains open as DSL-21).
- lang open language-design questions (range syntax, f-strings, truthiness, equality, pattern matching) — all decided + shipped (dsl.md "Resolved" section + lexer).
- DSL error-model implementation order 1–6 — done (`dsl/interp/error.go`, BangCallable `dsl_bridge.go:293`); actions.md's "stringly-typed. Ugly." status is the stale side.
- bounds `{}` proposal — accepted + implemented same day (`23477ec` → `5aebf1f`); canonical surface in api.md §spatial-primitives.
- REPL `.load`/`.run` — built (`runtime/repl.go` metaLoad/metaRun); `.watchers` blocker note ("when doesn't exist") dead — when shipped.
- `level_up` event — wired (`runtime/dsl_events.go:126`); trade/duel lifecycle events surfaced since 2026-05-28 (`dsl/spec/events.go:44-71`, `e6c78fa`/`28e1165`) — events.md's "planned" bullet was wrong at merge time.
- Versioning machinery itself (directive + CheckTarget + policy + CHANGELOG skeleton) — built `898b3fc` (the lapse is DSL-18, not the mechanism).
- scenario-candidates: ~19 of 22 superseded — gaps closed in body build-out; near-equivalents exist (edges-fatigue-set-zero-via-sleep-bag, edges-shop-buy-bronze-axe, magic-superheat-iron-ore, trade-confirm-screen-reached-on-both-accept, social-duel-stake-items, bot-autofighter-prayer-flick-strength, bot-bank-dialog-gate ≈ #21, bot-smith-bronze-bar-to-dagger ≈ #22); residue = S-3.

## Done — cognition (family-7, all commit-verified)

- Phase 0 `106ac72`+`c5aca73` · Phase 1 `e493c4e`..`95409a0` · 2a `d129589` / 2b `8a18c24` · 3a `82f82f7` / 3b `2daffee` · 4a `e376aba` / 4b `8a2e7d7` · 5a `eb999c4` · 5b-1 `6372684` · 5b-2/3 `2b31068` · stall→markGoalBlocked (`09d9216` pre-rebase).
- All 54 cognition-audit findings (2 critical / 18 high / …) — fixed + regression-tested (resolution header verified in code: SetBaseline race, H16 oldest-first cron cursor, H20 parseFlags, H5 Untag, mindRel Affinity/Grievance export, value_traded + episode entity/importance wired).
- Revision-backlog P0 #1 (tags-not-nodes), P0 #2a (foragingInFlight guard), P1 #3 (act.go:88 claim retracted), P1 #6 (suppress how-to-progress), P2 #8/#9, P2 #11 (`detour.go:141 case "forage"`) — absorbed into the shipped 5b specs.
- cognition-and-autonomy §6 "gaps to build" — ALL built except the §7-item-4 ladder remainder (= C-11): memory loop (`runtime/memory_journal.go`), mesa LTM (`mesa/mesad/ltm.go`), session genesis reduced-scope (`mesa/mesad/genesis.go` + bootstrap wiring), detour suspend/resume (`runtime/detour.go:64-120`), routine library + promotion (`runtime/library.go`, `hybrid_director.go`, `161c1a2`), TTL decision cache (`runtime/decision_cache.go`).
- §8.2 OtherPlayerChat companion-recognition fix — committed (`runtime/limbic.go:203`).
- Tutorial-Island bootstrap mode (brain.md "design, not built") — superseded: the generic MesaDirector Act loop completed Tutorial Island (`dacc425`) with no special mode.
- Trust ledger "designed but not built" (layers.md/cognition.md banners) — built as `limbic/ledger.go` Beta(α,β) + mesa mirror; the *remainder* is C-24.
- chat-interruption §8.2 cache-verb gap — done differently (tiered `remember`/`recollect`/`forget`, `dsl/spec/actions.go:488-491` + memory package).

## Done — plutonium / movement / soak (backlog #1–#15 namespace)

- #1 live-overlay grid, #2 DoorType==1 blocking, #3 plane-aware, #4 delete raw-Walk fallback, #6 opcode-211 decode — movement steps 1–5, `b24012c`.
- #8 approach rects, #9 reachability-filtered queries (done-differently via `a270236` + `worldmap/reachable.go`), #11 live peek-ahead, #13 BlockedEdge ledger (op-194 half still open → MP-12), #15 multi-plane oracle — steps 6–10, `6a329fd` ("Completes movement-design.md").
- #5 fatigue→sleep arbiter (`detour.go:279`) + #7 clean-EOF supervisor restart (`session.ErrServerClosed` + `conn_eof_test.go`) — soak trio `0214a0b`.
- In-session #31 door-blind reachability fix + "bug A" door pathing + interact_at(bed)→sleep — all superseded/closed by the above (SESSION-STATE-2026-06-09's IMMEDIATE NEXT TASK is finished).
- #34 — superseded outright by the plutonium study.
- EXECUTION-PLAN Phase 0 (typed door state: `IsOpenable`/`BlocksWhenClosed` in `facts/defs.go`, traversal consumes them) + Phase 1 core (`af8b17e` #36: `static_defs_gen`/`static_locs_gen`/`idnames_gen` + TestStaticParity; TileDef/OverlayBlocks in `facts/defs.go:177-194` incl. the 250→2 remap) — done in modified shape; remainder = MP-8.
- PlayerCoords bitpack widths — byte-verified correct against server source (5-bit/12-bit confirmed); the proposed 5→6-bit "fix" proven WRONG (would corrupt the stream; `GameStateUpdater.java:273 forAuthentic ? 5 : 6`).
- protocol-and-server-config caveats — fatigue-arbiter prerequisite shipped; EOF exposure fixed (both `0214a0b`).
- dsl-gap #1 (`&&`/`||` lexer aliases, `dsl/lex/lex.go:369` + manual teaches them) and #7a (phantom `accept_trade`/`accept_duel` purged, `events.go:44-47`) — done in `0214a0b`/`b42a52b`.
- The full **DSL-manual ground-up rewrite** + companion truth-fixes (validator `timeout 30` message, equipment_changed docstring, host.name NYI pointer) — done `b42a52b`, 17/17 examples adversarially validated (record: `dsl-manual-analysis-2026-06-10.md`).
- dsl-synthesis Phases A/B/C1/D — done in exactly its prescribed order (grid → arbiter/EOF → lexer → manual); residue = §2 C-items + Phase E (= MP-1/2/3, DSL-1..9).

## Done — infra / mesa / control plane

- hostkv (bbolt→pebble `a9c9e36`) + conductor turn-loop + `cmd/host` + pearl seams (TryDecide in dslDecide, Gate at dsl_bridge, POLICY_VETO) + memory tiered Manager + DSL memory verbs — all built (memory-spine workstream).
- mesa gRPC daemon (Act/Decide/Chat/Recall/Remember), LTM Postgres+pgvector+Voyage, consolidation + insight crons — built (`mesa/mesad/*`); mesa.md/architecture.md/index.md banners are the stale side (→ D-12).
- **Admin control plane v1** — shipped near-verbatim (PR #5, merge `1c2423e`): Admin proto + per-item bulk import + fail-closed ADMIN_TOKEN + `cmd/mesa-ctl` persona put/import/ls/get/set/rm.
- **Goal push** (`Admin.PushGoal` + glob fan-out, `server.go:665-729`, `cmd/mesa-ctl/goal.go`) and **fleet gen** (`cmd/mesa-ctl/fleet.go`) — v3 items shipped EARLY; **Provision.Subscribe live registry** (`server.go:609-662`) shipped, unblocking M-3.
- Cradle daemon + web UI + cradle-ctl (`730ae9a`) + per-host bounce `d0ff736` + dronegen families (`19d4ad0`/`c37d41c`) — built; delivers most of the Phase-6/observability intent out-of-band (residue = O-7/O-9).
- Leak-audit quick-patch set `478a690` — bus `*` Unsubscribe fixes (dsl_events/follow), Subscribe-after-Close guard, per-run WithCancel, dctx 5s bounded join, leak gauges, pprof on fleet mux, JSONL allowlist, Resume→drainInterrupts, nil-ctx guard. (Deferred 1–4 = O-1..O-4.)
- Decision-stream persistence (`runtime/decisions.go`, `4377c73`); runhost extraction steps 1–4 + 6–8 (cradle-server + legacy-cradle rename) done; step-5 pooling deliberately dropped (one conn per host).
- In-session tracker: #14 displacement arbiter `d698be3`; #19 f-string tolerance `5cda557`; #22 scan_for `937f707`; #23 analysis-UI premise FALSE (console already shipped `00a63fa`); #26 forage ✓; #28 ✓; #29 → pebble ✓; #36 ✓; recovered-scope JSONs all four closed.
- bots→hosts schema rename — MOOT (`mesa/mesad/ltm.go` born host_id-native).

## Done — persona

- Persona schema (Cornerstone/Trajectory, HEXACO+Schwartz+prefs+quirks+ReverieSeed, schema-version guard) — built (`persona/persona.go`); compile-first thesis shipped (`persona/compile.go` PersonaCard + `render.go` deterministic floor citing persona-compile.md §8 + `persona/policy.go` CompilePolicy→pearl).
- personacook Opus best-of-N cook + judge (rubric matches persona-compile §4/§5 nearly verbatim) — built; persona-open-decisions A1 answered (cohort=launch batch); §C items 1-2 + A5-adjacent enum registries built (`enums.go` SSOT).
- "Persona design session gates Phase 4" — historically falsified: the brain shipped first on a hand-authored persona; sessions proceed via the worksheet (P-1/P-2).
- E4 hybrid persona generation decision — done (personacook); mood/curiosity/attention designs from reveries.md — shipped as limbic.Affect + persona enums (engine remainder = P-6).

## Stale / superseded (do NOT carry as work)

- render-port-plan Phases 2–6 — SUPERSEDED banner correct; orsc port delivered differently; classic .jag cross-check oracle dropped (pipeline reads only OpenRSC .orsc).
- docs/protocol.md "Phases 0–3" — wire-handshake sequence, shipped; "Opcodes we need for Phase 0" sections are history.
- dsl.md "revisit Starlark if >4k LOC" tripwire — dead (dsl/ ≈ 12.7k LOC, decision long stood); stdlib cost tables + `routine_versions` punt-rate logging — design-era, never built as specced.
- brain.md body — Brain/DecisionRequest/Decision-union, MockBrain, brain_calls (as written), per-host budgets, 5-chunk caching, DSLGrammar/DSLSurface earned vocabulary — never manifested; archive per audit (cost ledger survives as O-9; G2 survives as C-12).
- cognition.md PrepareDecision/DSLSurface assembly + `GET /knowledge?q=` + knowledge_chunks table + cradle RAG cache (5-min TTL) — never built / architecture moved (mesa-side assembly); wiki shipped as cognition/corpus + offline wikirag.
- mesa.md REST API + bots/auth_token/reflections/routines/brain_calls schema + mesa-graded TrustGrade/decay — never built; authority direction REVERSED (host-authoritative AuthLocal, mesa mirrors); routine-version GC moot.
- observability.md cmd/delos + obs package + bot_id vocabulary + Phases table — never built as named; intent delivered by cradle UI/debughttp (residue = O-7/O-9); "force reflection" + "adjust cohort" admin actions stale (cohort reframed).
- memory.md four-category lifecycle (ImportanceOf scorer, compression stages, reinforcement formula, tau tuning, reflections table, bot_vocabulary, cache.\* verbs) — never built / shipped differently (storage tiers + kind-importance + crons); archive w/ fresh replacement (D-12).
- knowledge.md (lang) Slice 2 mesa-backed corpus.Corpus — superseded by mesa Recall RPC + wikirag + LTM; Slice 4 CLI = M-9.
- questions-and-decisions recorded-decisions drift — F2P posture REVERSED (deployed P2P `member_world:true`); E1 reflections-table + D5 agent-decided cadence manifested as fixed crons; D1 in-process AnthropicBrain → mesa gRPC; G1 MockClient never existed; A2′/F2 mandatory reverie.tick() unbuilt (→ P-6, as scheduler not tick); C1 resolved. (Annotation pass = D-12.)
- SESSION-STATE.md (2026-06-02) next-steps — persona/trust/agent-driver/tutorial all done or done-differently; "Pearl REJECTED" REVERSED (pearl built + wired); only G2 survives (C-12); /tmp/mmv tooling dead.
- host-bootstrap §5 station-sequence scaffolding — overtaken by genesis + autonomous play; §8/§9 checklist survives only inside the C-12 decision.
- v2 component-zoo design (398KB) — formally rejected, then Pearl/Limbic rebuilt in seam-respecting shape from the decision records; reference/ keeps the archive; reference.bak/ deletion = D-13.
- decision-transport-storage body (LLM/Voyage in cradle, process-muxed conn, Redis at Phase B) — corrected by Alex's §0 compute-locality block; Redis/River never adopted anywhere; Phase-B gRPC split happened.
- diagrams: c1-pearl PearlImage/hot-slice/decision-cache-in-Pearl, c2-director brain.Director, c5-guard decision-seam-only veto, s4 goal-ladder, mr3 Sink pipeline, mr4 "no durable host store" premise — all manifested differently (pearl rule engine, conductor/HybridDirector, Gate at action seam, goalgraph picker, limbic+journal, hostkv local tier).
- external research reports (claude/gemini/gpt55 ×2 each) — provenance, not work; their adopted ideas live in code; their unadopted specifics (Bartle quotas, M0–M4 stages, five-service proto, SNN hardware, IPIP/QLoRA pipelines, 13-archetype sampler as specced) are not TODOs (sampler/eval-battery survive as P-3/R-4).
- event-study tier-1 adds sleep_captcha/disconnected/idle as DSL events — needs met at other layers (cradle auto-answer, supervisor restart, stall detector); only the DSL-19 residue survives.
- "Plutonium is a byte-oracle" framing + "different servers" myth — both corrected in the study; protocol verdict: ours authentic-235, FINE.
- Two of the five code TODOs (`action/sleep.go:33`, `action/shop.go:26`) — describe *server*-side TODOs, informational only.
- lang/state.md region_name accessor + self.spells.selected — superseded (gazetteer verbs; no client-side selected-spell state worth mirroring).
- repl.md `<repl-line-N>` identity + `exec:<hash>` minting (syntax.md) — fiction; actual names `<repl>` / `mesa/authored`; exec()/improvise() remain deliberate NYI stubs (capability landed via the Act/WriteRoutine path).
- agents/dev-partner first-task + mission — completed differently (in-repo cognition/mesa workstreams); charter archives per audit (reveries pointer survives as P-6).
- steward first-task items 1–4 — one-time bootstrap whose specific claims (server-config F2P/43596 inaccuracy) were since fixed; the symlink survives as O-12.
