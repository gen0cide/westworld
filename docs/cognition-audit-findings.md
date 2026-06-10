# Cognition Stack Audit — Opus Review (Phases 1–5a + orsc-ctl)

**Audit date:** _(to be filled by operator)_
**Scope:** the full host cognition stack — knowledge ledger, goal graph, perception writers,
director reasoning/lifecycle/anti-stuck, reactive tier, speech, limbic/relationships,
host persistence, mesa consolidation + insight crons, LTM storage, mesa act prompts,
mesa server handlers, wire round-trip consistency, the `orsc-ctl` admin CLI, and the
cross-cutting concerns (concurrency, persistence spine, tiering/cost, host-light,
goal-graph lifecycle, doc fidelity).
**Method:** every line of Phases 1–5a was previously reconned only on Haiku/Sonnet — never
Opus. This pass was run by 23 parallel Opus auditors, then deduplicated/prioritized here.
Findings cite exact `file:line`; many were confirmed empirically (`go test -race`, focused
probe tests, standalone logic replicas). The known CRITICAL the weaker tiers missed
(relationships table dropping affinity/grievance) is **closed** and recorded below as a
positive control.

**Severity counts (post-dedup):** 2 critical · 18 high · 14 medium · 16 low.
**Dropped low-confidence nits:** 3 (see end).

---

## RESOLUTION — ALL FINDINGS FIXED (2026-06-08)

Every CRITICAL / HIGH / MEDIUM / LOW finding below is **fixed + regression-tested**, applied
in three Opus-driven waves (fan-out fix → cross-package integration closeout → hand-finished
concurrency), each adversarially Opus-verified.

- **Gate (whole tree, on the settled branch):** `go build ./...` ✅ · `go vet` ✅ · `gofmt -l` clean ✅ ·
  `go test -race ./runtime/... ./cognition/... ./limbic/...` ✅ (runtime ~30s, no DATA RACE) ·
  `go test -race ./mesa/mesad/... ./cmd/...` ✅ against a **live Postgres** (validates the DB-backed
  cron/LTM/server fixes + H13/H14, not skipped). Only non-green marks are two PRE-EXISTING,
  out-of-scope items untouched by the fixers: `proto/v235/buffer.go` (vet) and `cmd/dronegen/main.go` (gofmt).
- **Confirmed-real races, fixed + `-race`-verified:** C1 (Affect re-baseline → in-place `SetBaseline`
  under mutex), C2 (duel-fight window), H11 (combat round counters under `combatMu`), H12 (keyword
  ladder → `atomic.Pointer`), H13 (registry entry copy-on-write / guarded). A confirmed regression
  the fix wave introduced (H8's new speech-tier `effectiveGoalView` read) was caught in verify and
  closed with a `MesaDirector.goalMu` RWMutex (adversarially reverted → race test fails without it).
- **Latent traps (recorded so they don't become the next hidden 3b) — both WIRED end-to-end:**
  `relationships.value_traded` (limbic accumulator + converter + atomic `LTM.Add` fold) and episode
  `entity`+`importance` on the journal cold-start round-trip (new `KnowledgeItem` proto fields +
  `buf generate` + Recall population).
- **H17/M17 (two-writer goal-graph/knowledge LWW clobber):** non-destructive `ImportMerge`
  primitives + a periodic host warm re-import of the cron-reconciled snapshot + a strictly-newer
  (`maxNodeAt`) guard in `SyncGoalGraph` (host owns status nodes, cron owns chain/closure nodes).
- **Bonus (beyond the H11 finding):** `lastAttacked*Index` — a benign-but-real cross-goroutine
  read/write (the old "single-goroutine" comment was factually wrong) → `atomic.Int32`. Zero
  behavior change; keeps the `-race` gate trustworthy.

The doc fidelity / disclosed-stub notes at the end remain as design-doc follow-ups (not code bugs).
Findings below are kept verbatim as the durable ledger of what was found and why.

---

## CRITICAL

### C1 · Data race on the `h.affect` pointer (genesis path)
- **Location:** `runtime/limbic.go:28` (`SetAffectBaseline`) vs readers at `runtime/limbic.go:107-111`, `runtime/pearl.go:36-38`, `runtime/decision_cache.go:31-33`, `runtime/mesa_director.go:520`; trigger at `runtime/runhost_bootstrap.go:159` (`go host.Run`) then `:192` (`SetAffectBaseline`).
- **Problem:** `SetAffectBaseline` does `h.affect = limbic.NewAffect(...)` — a whole-pointer reassignment with no lock. The genesis path starts `host.Run` (which launches `runLimbic` + the pearl/director readers) and ~1s later calls `SetAffectBaseline`, swapping the pointer while the limbic goroutine calls `h.affect.OnXPGain/OnLevelUp/OnDeath` on the login burst and pearl/decision-cache/director read `h.affect.Snapshot()`. Confirmed with `go test -race`: `WARNING: DATA RACE` between `limbic.go:28` and `limbic.go:107` and inside `limbic/affect.go`. A reader can observe a half-initialized `Affect` or a nil pointer (panic). The `applyPersona` path (`runhost_bootstrap.go:485`) is pre-`Run` and safe; only the genesis post-`Run` call races. The internal `Affect` mutex does not help — the race is on the field holding the `*Affect`, not its contents.
- **Fix:** Do not reassign the pointer at runtime. Add `func (a *Affect) SetBaseline(s,c,v float64)` that updates `baseS/baseC/baseV` (and current values) under `a.mu`, and have `SetAffectBaseline` call `h.affect.SetBaseline(...)`. Alternatively, move the genesis `SetAffectBaseline` call to *before* `go host.Run` so no reader is live.
- **Confidence:** high. _(Merged: limbic-relationships + XC-concurrency.)_

### C2 · Duel grievance gate is OPEN during the actual fight (duel ≠ PK violated)
- **Location:** `runtime/limbic.go:117-125` & `172-199` (the `!inDuel()` gates) vs `runtime/limbic.go:296-298` (`inDuel`); `world/world.go:1061-1072`; `world/duel.go:73-81,248-262`; event def `event/events.go:749-755`.
- **Problem:** `DuelClosed{Completed:true}` fires when the fight *starts*, not when it ends (`events.go:749-755`; `duel.go:248-249`). `world.Apply` runs `MarkClosed(true)` → `Phase="completed"` BEFORE `bus.Publish`, so by the time a `Death`/`OtherPlayerProjectile` lands DURING the fight, `inDuel()` (Phase != completed/cancelled) returns FALSE. Confirmed: `IsActive()==false` after `MarkClosed(true)`. Consequence: a death or hostile projectile during a *skulled* duel (e.g. a wilderness right-click challenge where `SkullType!=0`) passes the `!inDuel()` gate and is recorded as a wilderness PK gank — `ObserveGrievance(opponent, 3.0)` + `"ganked-me"` tag (`limbic.go:120-121`) — directly violating the cardinal duel≠PK invariant. The existing regression `TestDeathInDuelNoGrievance` only asserts the gate during the OFFER window (`Phase="open"`), never during the fight, so it gives false confidence. (Arena staked duels don't skull, so `pkKiller` returns false there — the bite is the skulled wilderness-duel case.)
- **Fix:** Track an explicit "in a sporting fight until the post-fight result is observed" window rather than keying off the duel-UI phase. On `DuelClosed{Completed:true}`, set a host flag/timestamp (`duelFightUntil`) and have `inDuel()`/a new `inSportingFight()` also return true while live, cleared on the next respawn/`StatsSnapshot`; OR keep the `DuelRecord` in a non-terminal "fighting" phase until a death/respawn is seen. Add a test firing `Death` AFTER `DuelClosed{Completed:true}` asserting zero grievance.
- **Confidence:** high.

---

## HIGH

### Director — anti-stuck / lifecycle (graph-write family)

> These five share a root: the anti-stuck paths perform **durable goal-graph writes** that
> are not guarded against terminal states or the displacement override, and the spin signal
> is fed by no-op idles. Fixing the precedence + guards closes most of them together.

### H1 · Completed goal is resurrected and re-looped (no Done-guard on the blocked write)
- **Location:** `runtime/mesa_director.go:551-558` (spin → `markGoalBlocked`), `:143-160` (`markGoalBlocked`, no Done guard), `:434-441` (un-block recovery, no Done guard); contrast guarded progress path at `:215`.
- **Problem:** A planner-COMPLETED goal (`goal_op:"done"` → `markGoalDone` → `StatusDone`) with no successor leaves `effectiveGoal` returning the Done goal as `sit.Goal` every turn (`:105-111` only advances when `selectNextGoal != ""`). The planner re-emits the same plan and SUCCEEDS; `spinCount` climbs (the inventory-net at `:472` doesn't fire for a non-acquire goal so it never resets `spinCount`). At `spinCount>=3`, `markGoalBlocked` (no Done/Abandoned guard, `:149`) flips the Done goal to Blocked; the next OK turn the un-block recovery (`:434`, gated only on `Status==Blocked`) flips it Blocked→Active at progress 0.25. The completed goal is now ACTIVE again and loops forever. Reproduced end-to-end at the director level (turns 0-2 done, turn 3 blocked, turn 4 active). `TestLifecycleAdvancementStaysPut` only checks ONE `effectiveGoal` call so it passes while the bug is live. Note the asymmetry: `applyGoalOp`'s progress path IS Done-guarded (`:215`); the blocked-write path is not.
- **Fix:** Guard the non-monotonic status writes against terminal states. In `markGoalBlocked`, bail when `n.Status==StatusDone||StatusAbandoned`. In the spin branch (`:551`), skip entirely when `effectiveGoal`'s node is Done/Abandoned. Long-term, stay-put-on-a-done-goal should idle (or surface a distinct "goal complete, awaiting next objective" trigger) rather than handing the planner a Done goal.
- **Confidence:** high.

### H2 · Anti-stuck graph writes run BEFORE the displacement override (precedence bug)
- **Location:** `runtime/mesa_director.go:551-558` (SPINNING) and `:529-540` (BLOCKED hard-fail) relative to `:564-579` (displacement block).
- **Problem:** On a turn that is both spinning/hard-failing AND displaced, the SPINNING/BLOCKED blocks see `trigger == baseTrigger` (true) and perform their durable writes — `markGoalBlocked(g)` (goal→Blocked + "how-to-progress" open question + "enabler:g" subgoal) and `Tag(g,"spinning")` — and THEN `h.displacement.take()` overrides the trigger to DISPLACED. Net: the planner correctly gets the DISPLACED re-orient text, but the goal is durably mis-marked Blocked+spinning by a spin that was about the OLD plan/location no longer applicable after the jump. The block's own comment claims it fires only if "death/displacement override BELOW" hasn't changed the base — but displacement is BELOW, so it cannot have run yet. Reproduced: `trigger=DISPLACED` while goal status='blocked' tags=[spinning]. `TestMesaDirectorDisplacementResetsFailStreak` only checks the trigger string + `failStreak==0`, never the node status, so it missed this.
- **Fix:** Peek displacement FIRST (split `take()` into peek+commit, or compute `disp, displaced := h.displacement.take()` at the top of the trigger section) and gate the STUCK/BLOCKED/SPINNING graph writes on `!displaced`, applying the displacement trigger string last.
- **Confidence:** high. _(Merged: director-antistuck findings 1 + 3.)_

### H3 · Displacement does not reset `spinCount`/`lastPlanFP`
- **Location:** `runtime/mesa_director.go:564-579` (displacement block resets only `failStreak` at `:571`).
- **Problem:** A death/teleport is a hard context change. The displacement block resets `failStreak` ("was about the OLD situation") but leaves `spinCount` and `lastPlanFP` dirty by the identical reasoning. Reproduced: `spinCount=1` before a `dispDeath`, then the post-respawn turn re-ran the same fingerprint and `spinCount` climbed to 2 — the spin counter can reach threshold and fire on executions begun in a different location, exactly the false-positive class displacement is supposed to suppress.
- **Fix:** In the `if displaced {` branch add `d.spinCount, d.lastPlanFP = 0, 0` alongside the existing `d.failStreak = 0`.
- **Confidence:** high.

### H4 · Spin detector counts IDLE waits as a stuck loop
- **Location:** `runtime/mesa_director.go:452-464` (spin detector) + `:268-296` (`planFingerprint`) + `:544-558` (SPINNING fire); idle sources at `:355`, `:1244`, `:1258-1261`.
- **Problem:** The spin detector fingerprints EVERY OK intent, including `wait(N)` idles. An idle is the ABSENCE of a plan, not a re-derived one; `d.idle(secs)` emits a stable Source/fingerprint. A host that legitimately idles a few turns — OR the director's OWN fallbacks (planner error → `d.idle(3)` at `:355`; empty-verb direct action → `d.idle(2)` at `:1244`) — accumulates `spinCount`. At `antiStuckSpinTurns(3)` the SPINNING path fires: marks the active goal `StatusBlocked`, tags it "spinning", and nudges abandon/false-complete. So a valid goal is corrupted while the host merely waited — and it fires *exactly* when the LLM is unavailable (3 consecutive planner errors). Verified empirically with the host moving (so STUCK doesn't pre-empt).
- **Fix:** Exclude idle/no-op intents from the spin signal. Simplest: `planFingerprint` returns 0 for an idle intent (`fp==0` already resets `spinCount` at `:459-463`). Also consider not counting the director's error/empty-verb idle fallbacks as goal-work.
- **Confidence:** high.

### Director — reasoning / completion

### H5 · "spinning" tag is permanent (no untag path)
- **Location:** `runtime/mesa_director.go:555` (`Tag(g,"spinning")`), `cognition/goalgraph/goalgraph.go:167-174` (`Tag`, append-only), `runtime/mesa_director.go:899-901` (`renderIntention`).
- **Problem:** goalgraph exposes `Tag` (append-if-absent) but no `Untag` anywhere (grep-confirmed). Once set, `renderIntention` forever appends "(you've been SPINNING on this — if it's done, say so)" to that goal's planner-facing prose, even after the host recovers, un-blocks, and makes real progress (`:432-441` flip status Active but never touch Tags). Combined with H2/H3, a single spurious spin-mark during a displaced turn permanently pollutes that goal's prose for the rest of the persisted life, and injects a stale "find what is preventing X" enabler subgoal into the very re-orient context where the planner was told to decide fresh.
- **Fix:** Add `Graph.Untag(id, tag)` and clear "spinning" wherever spin recovers (the OK-turn un-block recovery `:432-441`, and `markGoalDone`/abandon). Avoiding the spurious tag (H2) is the primary fix; an untag path is the durable backstop.
- **Confidence:** high.

### Reactive / speech — the ask→learn→close loop

### H6 · ASK path latches the target AFTER emitting (Q→A pairing never happens)
- **Location:** `runtime/speech.go:205-213`; `runtime/host.go:1380` (`Say`→`reactiveObserveSelf`); `runtime/reactive.go:169-187` (`tryLatch`), `:275-284` (`reactiveObserveSelf`).
- **Problem:** The ASK loop's central mechanism — "speaks via `host.Say` which fans the line into the reactive windows so the answer pairs with the question" — is broken by ordering. `emit(ctx,text)` runs at `:205` (→ `host.Say` → `reactiveObserveSelf`) but the target is only latched at `:213`, AFTER emit. `reactiveObserveSelf` early-returns when `latchCount()==0` (`reactive.go:280`), so at emit time (zero latches) the host's own question line is silently DROPPED and never fanned into the target window — when the answer later arrives, `snapshot(target)` contains only the answer, not the question, so the Q→A pairing never happens. In the cold case (host never previously heard the target — the normal case for asking a shop NPC) `tryLatch` also returns false (`reactive.go:172-175` requires a pre-existing window). Both proven: cold → window=[], latchCount=0; warm → latch=1 but window holds only the old NPC line, not the asked question.
- **Fix:** Latch + ensure the target window exists BEFORE emitting. (a) Have `tryLatch` create the window when `windows[key]==nil` instead of returning false; (b) move `tryLatch(normalizeSpeaker(tgt.name), now)` ABOVE the `emit()` call so the self-question is fanned in. On send failure, unlatch/drop to avoid an orphaned latch. Add a regression test asserting the target window contains a `"Me: ..."` line after a successful `tryAsk`.
- **Confidence:** high.

### H7 · Open-question closure gates on the RAW LLM confidence, not the role-derived value
- **Location:** `runtime/reactive.go:436-459` (`closeResolvedQuestions`, raw `c.Confidence` at `:443`) vs `:406-422` (`writebackClaims` defaulting); confidence is also unclamped upstream at `mesa/mesad/act.go:455-460` (`parseExtractedDialog`) and `mesa/client/grpc.go:158`.
- **Problem:** The don't-know→ask→learn→**close** loop (the §3.8a design centerpiece) silently fails to close for authoritative answers. `writebackClaims` DEFAULTS an out-of-range confidence to 0.85 for npc/server and 0.5 for player BEFORE writing the belief, but `closeResolvedQuestions` reads the RAW `c.Confidence` against `closeQConf=0.6`. The extractor passes the model's confidence straight through with no clamping, and the prompt lets the LLM omit/emit 0. So an NPC/server answer with `Confidence:0` (modeled by `reactive_test.go:199`) is written at 0.85 (authoritative) yet SKIPPED by the closer (raw 0 < 0.6) — the question stays OPEN and its goal stays BLOCKED. Symmetrically, a player hearsay claim the LLM tagged 0.7 DOES close the question (0.7 ≥ 0.6) because the closer has no role parameter — letting the host act on a confident-sounding lie as game-authoritative, violating the "provenance from ROLE not LLM self-report" invariant `writebackClaims` itself enforces. This is the same class as the affinity/grievance drop: the contract is honored in one place (the belief's Provenance) and dropped in the adjacent decision (whether to close).
- **Fix:** Compute the effective `(conf, provenance, role)` exactly as `writebackClaims` does, ONCE, and gate the closer on that: only `role==npc||role==server` may auto-close, and use the defaulted 0.85 (not raw 0) for the floor test. Extract `effectiveConf(role, raw)` and call it in both writeback and closer. Durable backstop: clamp confidence to `[0,1]` in `parseExtractedDialog` so both consumers see one value.
- **Confidence:** high. _(Merged: reactive-tier H + speech M + mesa-act-prompts M + XC-doc-fidelity H.)_

### H8 · `pickAskQuestion` slices blockers off `LiveGoal()` (empty in the normal case)
- **Location:** `runtime/speech.go:260` (`Blockers(h.LiveGoal())`).
- **Problem:** `LiveGoal()` returns the operator-pushed goal override and is `""` in the normal autonomous case (`host.go:362-364`; only set on `GOAL_REVISION`). `Blockers(id)` is a node-ID lookup. The rest of the codebase slices blockers off `d.effectiveGoal(h)` (`mesa_director.go:683→723`), which falls back to the persona/genesis goal. So speech queries `Blockers("")` → empty, and Pass-1 ("prefer the question blocking the live goal") NEVER engages in the common path — the flagship "ask about what blocks your goal" prioritization is effectively dead and inconsistent with the director.
- **Fix:** Use the same effective goal the director uses: expose `effectiveGoal`/`d.goal` to the host and call `Blockers(effectiveGoal)`, or fall back to the genesis/persona goal node ID when `LiveGoal()==""`. Add a test where `d.goal` (not `LiveGoal`) names a goal node with a `blocked_by` open question and assert Pass-1 selects it.
- **Confidence:** high.

### Perception writers

### H9 · `NpcNearby` movement events credit a phantom "Unicorn" (NPC type 0)
- **Location:** `runtime/perception.go:55-60` (NpcNearby handler); `proto/v235/npccoords.go:102`.
- **Problem:** The handler resolves the NPC via `h.npcNameByType(e.TypeID)`, but ONLY the new-NPC wire section carries a real `TypeID`. Movement updates — the bulk of NpcNearby traffic (opcode-79 emits `NpcNearby{Index,DX,DY,Sprite,IsNew:false}` with `TypeID` zero) — have `e.TypeID==0`. In the OpenRSC catalog NPC type 0 == "Unicorn", so every step of every moving NPC calls `h.knowledge.Seen("Unicorn","npc")`, inflating a phantom subject hundreds of times per session while the NPC that actually moved gets familiarity credit only on its one-time `IsNew` spawn. Corrupts the persisted, mesa-mirrored knowledge ledger on essentially every session and grows unbounded. Proven: a Giant rat (idx 7, type 3) movement credits Unicorn=1, Giant rat=0. `perceiveKill` does it correctly (reads `rec.TypeID` from the world model); this handler is the lone outlier trusting the raw event's type.
- **Fix:** Resolve by scene INDEX through the already-applied world model: `if n := h.npcNameByIndex(e.Index); n != ""` (`npcNameByIndex` exists at `perception.go:254`). `world.Apply` runs before `bus.Publish` and `MoveBy` preserves the stored `TypeID`, so index resolution names the moving NPC correctly on both spawn and movement.
- **Confidence:** high.

### Host persistence — unbounded growth

### H10 · Host-side knowledge ledger + goal graph have NO cap and NO host-side pruning
- **Location:** `cognition/knowledge/knowledge.go:78-101,128-129` and `cognition/goalgraph/goalgraph.go:82-326` (writers: `runtime/perception.go:58/139/162/166/197/209/219`, `runtime/reactive.go:422`, `runtime/mesa_director.go:148-158,209,446`); contrast `memory/journal.go:54,90-92` (`DefaultJournalCap=64`).
- **Problem:** Both host-side structures grow MONOTONICALLY with no upper bound, violating the package docs (`goalgraph.go:11-13` "the distillation crons grow AND prune — NOT unbounded growth") and the host-LIGHT invariant. There is NO `Delete/Remove/Prune` API on `*Graph` (verified — only status-flips to done/abandoned, which STAY in `g.nodes`/`g.edges` forever). The director Upserts a brand-new `open_goal` node for EVERY planner "adopt" keyed by free-text (`mesa_director.go:209`) — paraphrases accumulate as distinct permanent nodes; it adds an open_question + subgoal + enabler node on every block. `knowledge.Note` (`reactive.go:422`) appends a row per distinct LLM-derived subject and a Belief per distinct claim with no cap. The only host GC is `gcLatches`/`gcSpeech` — neither touches these ledgers. Worse, it compounds: `flushKnowledge/flushGoalGraph` + the `…ToMesa` mirrors run every 30s and `memory.Manager.Put` re-serializes the ENTIRE growing JSON to bbolt AND re-uploads to mesa each cycle, so RAM and per-flush I/O both grow without bound. `Link`'s dedup is O(E) per call so bulk re-linking is O(E²). The distillation crons that "grow and prune" run on MESA, not on the host's in-RAM copy — nothing ever shrinks the host structure.
- **Fix:** Bound both like `Journal`. Knowledge: max-subjects cap (evict by oldest `LastSeen`/lowest `Encounters`) + per-Entry max-beliefs cap (drop lowest-confidence/oldest). Goalgraph: add a bounded `PruneTerminal(maxAge, keep)` (drop done/abandoned older than a horizon + dangling edges, cap total nodes by evicting oldest terminal by `At`) and call it from the limbic flush tick alongside `gcLatches`/`gcSpeech`. Dedup adopt-goals by normalized label before Upsert. Keep deterministic + O(1)-amortized.
- **Confidence:** high. _(Merged: goalgraph-pkg H + host-persistence H.)_

### Limbic / combat concurrency

### H11 · Data race on the combat-tracking int/time fields
- **Location:** `runtime/host.go:779-797` (`noteCombatRound`) + `:920-924` (`emitTargetDeathEdge`) vs `runtime/combat.go:138,419,440,462-468` (`confirmEngaged`/retreat) + `runtime/actions_combat.go:22,28,58` + `runtime/limbic.go:316` (`pkKiller`).
- **Problem:** `combatRoundTarget`, `combatRounds`, `outgoingHits`, `combatStartedAt`, `lastAttackedNpcIndex` are plain fields with no synchronization. WRITTEN by the frame-pump goroutine (`noteCombatRound` `combatRounds++`/`outgoingHits++` at `:787-794`; `emitTargetDeathEdge` zeroes at `:920-924`), concurrently WRITTEN+READ by the conductor/routine goroutine (`confirmEngaged` `combat.go:465-468` set, `:138` read, retreat gate `:419/:440`), and READ by the limbic goroutine (`pkKiller` `limbic.go:316`). Three goroutines, no mutex/atomic. Confirmed with `go test -race`: `WARNING: DATA RACE … host.go:787` vs the routine-goroutine read. Effect: `confirmEngaged` can mis-read `combatRounds` and re-attack a live fight (the exact regression its comment warns against — resets the server combat event and stalls the kill), or the retreat anti-kite gate fires on a torn count.
- **Fix:** Guard the combat-tracking block with a dedicated `sync.Mutex` taken in `noteCombatRound`, `emitTargetDeathEdge`, `confirmEngaged`, retreat, attack-dispatch, `pkKiller`, and the `views_combat` accessors; or move the counters into a small mutex-guarded struct with `Note/Reset/Snapshot` (mirrors `reactiveState`). A single mutex keeps the host light.
- **Confidence:** high. _(Merged: limbic-relationships combat-race + XC-concurrency.)_

### H12 · Data race on `h.keywordLadder` (bootstrap after Run)
- **Location:** `runtime/host.go:369` (`SetKeywordLadder`, unlocked `h.keywordLadder = l`) vs `runtime/reactive.go:305` (`triggerHit` ranges over `h.keywordLadder`); call site `runtime/runhost_bootstrap.go:208` (after `go host.Run` at `:159`).
- **Problem:** `SetKeywordLadder` is called at bootstrap AFTER `host.Run` started the limbic goroutine, which ranges `h.keywordLadder` for every dialog line in `triggerHit`. A chat/server event between Run start and `:208` races the slice-header write against the range. Confirmed with `go test -race`: `reactive.go:305` vs `host.go:369`. Torn slice header → possible OOB/garbage read.
- **Fix:** Hoist the `SetKeywordLadder` call to BEFORE `go host.Run` (it's a one-time bootstrap snapshot), or guard the field with a mutex / `atomic.Pointer`.
- **Confidence:** high.

### Mesa server handlers

### H13 · Data race on `entry.goals` / `entry.goalPushed` (PushGoal vs Subscribe/Fetch)
- **Location:** `mesa/mesad/server.go:600-601` (Subscribe read) + `:671-672` (PushGoal write) + `:583` (Fetch read); `lookup()` at `:234-239`.
- **Problem:** `lookup()` takes `s.mu.RLock`, reads the `*entry`, RUnlocks, and RETURNS the pointer — every caller reads the entry's mutable fields with NO lock. `Admin.PushGoal` mutates those same fields under `s.mu.Lock()` (`e.goals = []string{goal}`, `e.goalPushed = true`). Concurrently Subscribe reads `e.goalPushed && len(e.goals)` and `json.Marshal(e.goals)` outside the lock, and Fetch returns `e.goals` straight into the proto. A slice header is 3 words and its assignment is not atomic, so a reader can observe a TORN header (new ptr + old len) → OOB read/panic, not merely a stale read. Reachable in production: any `mesa-ctl push-goal` while any host holds an open Subscribe stream (the steady state). Confirmed by the project's OWN `TestAdminPushGoal` under `-race` (`Write at PushGoal server.go:672` vs `Previous read Subscribe server.go:600`).
- **Fix:** Make `PushGoal` copy-on-write like `registerLocal` — under `s.mu.Lock` build a NEW `*entry` (copy + new goals/goalPushed) and reassign `s.reg[id]=ne`, so every looked-up pointer is an immutable snapshot. Alternatively have `lookup()` copy out the needed value fields under the RLock, or guard Subscribe/Fetch reads with `s.mu.RLock` for the duration.
- **Confidence:** high. _(Merged: mesa-server-handlers + XC-concurrency.)_

### H14 · Graceful shutdown hangs forever while any host holds a Subscribe stream
- **Location:** `cmd/mesad/main.go:199` (`gs.GracefulStop`) interacting with `mesa/mesad/server.go:589-616` (Subscribe).
- **Problem:** On SIGTERM, `gs.GracefulStop()` sends GOAWAY but does NOT cancel the server-side handler context; it blocks on `for len(s.conns) != 0 { s.cv.Wait() }` until every connection closes. A connection closes only when its in-flight streams end. Subscribe's select returns only on `stream.Context().Done()`, which under a graceful drain fires solely when the CLIENT disconnects. So with a live host (or a 200-drone fleet, each holding a Subscribe stream), `GracefulStop` never returns and mesad never exits cleanly — requires a second SIGKILL.
- **Fix:** Give Subscribe a server-driven exit. Add a `shutdown chan struct{}` (or context) on `Server`, close it in `StopCrons`/a new `Server.Shutdown()`, and add `case <-s.shutdown: return nil` to Subscribe's select. Alternatively bound the drain in `main.go`: run `gs.GracefulStop()` in a goroutine and fall back to `gs.Stop()` after a timeout.
- **Confidence:** high.

### Cron — consolidation / insight

### H15 · `entriesToKnowledge` α/β coercion corrupts valid confidence-1.0 and -0.0 beliefs
- **Location:** `mesa/mesad/cron.go:546-560` (mesa converter, coerces `alpha<=0`→1 and `beta<=0`→1); contrast `runtime/knowledge.go:114-117` (host converter, no coercion); source writes `runtime/perception.go:166,209,219`.
- **Problem:** The guard fires not only on a malformed row but on the LEGITIMATE single-zero case. A confidence-1.0 belief is `Alpha=1,Beta=0` (`Note` sets `Beta=1-conf=0`; `Observe(...,true)` on a fresh claim sets `Beta=0`); a confidence-0.0 disconfirming belief is `Alpha=0,Beta=1` (`perception.go:166` `Observe(subject,claim,false,1.0)` for out-of-stock = "Bob's does NOT sell pickaxes"). On the FIRST consolidation/insight round-trip (Knowledge → `entriesToKnowledge` → Import → fold → Export → SyncKnowledge), `Beta 0→1` drops `Confidence` from 1.0 to 0.5, and `Alpha 0→1` lifts confidence-0.0 to 0.5 — silently corrupting both authoritative server-msg claims and host-pushed "killable/obtainable by me" beliefs, AND inverting the "guessing vs knowing" distinction the ledger exists to encode. Confirmed: `Note(conf=1.0)` round-trip yields 0.5. Same class as the affinity/grievance drop (a value mutated mid-round-trip by an asymmetric converter), violating the lossless round-trip invariant.
- **Fix:** Coerce ONLY the genuinely-degenerate case: `if alpha<=0 && beta<=0 { alpha=1; beta=1 }` (same guard `knowledge.Belief.Confidence` uses), and separately clamp only NEGATIVE values to 0. Never coerce a single zero side. Extract ONE shared converter so host and mesa can never disagree on the same wire bytes.
- **Confidence:** high. _(Merged: cron-consolidation + XC-persistence-spine.)_

### H16 · Insight cap permanently loses overflow escalations (cursor over-advances)
- **Location:** `mesa/mesad/cron_insight.go:218-221` (sort newest-first + cap to `InsightMaxPerHost`) + `:278` (`cur.LastUnix = maxAt`).
- **Problem:** Candidates are sorted NEWEST-first and truncated to the cap (default 6), keeping the 6 newest admitted items and discarding the older admitted ones. But the cursor is then advanced to `maxAt` (newest AtUnix of the FULL read), which is `>=` every dropped item's AtUnix. Next tick all cap-dropped items satisfy `it.AtUnix <= cur.LastUnix` and are silently skipped FOREVER (the cron never writes back to the queue). The escalation queue is exactly the pre-filtered high-value slice destined for expensive Tier-2 reasoning, so this is a silent loss of the most important inputs. Worsened by second-granularity `AtUnix` (`recordEscalations` stamps every flag from one consolidation with the same `time.Now().Unix()`, `cron.go:376`). Replicated in a standalone harness: 20 admitted, cap=6 → tick 1 processes 6, tick 2 admits nothing, 14 lost. `TestInsightAdmitCapBounded` only asserts the cap, never that overflow is later processed.
- **Fix:** Advance the cursor only to the max AtUnix among ADMITTED/PROCESSED items (sort OLDEST-first so the cap keeps the oldest and the cursor advances to the newest kept candidate's AtUnix, leaving newer overflow for next tick). The persona-drop skip-advance branch is fine; the CAP branch must not. Because coarse-second `AtUnix` still loses same-second siblings, also give the queue a monotonic per-item id (or processed-set keyed by fingerprint). Add a test seeding `>cap` admitted items asserting a second tick processes the remainder.
- **Confidence:** high. _(Merged: cron-consolidation + cron-insight — identical root.)_

### Cross-cutting — persistence spine, tiering

### H17 · Goal graph has two whole-row LWW writers; the host clobbers the insight cron's growth
- **Location:** `runtime/goalgraph.go:87-101` (`flushGoalGraphToMesa`, every 30s, full local graph) + `mesa/mesad/ltm.go:631-645` (`SyncGoalGraph`, whole-blob `ON CONFLICT (host_id) DO UPDATE SET snapshot = EXCLUDED.snapshot`) + `mesa/mesad/cron_insight.go:272` + `runtime/limbic.go:50-56` (Empty()-guarded bootstrap).
- **Problem:** The single `goal_graphs` row has TWO independent whole-blob LWW writers — the host (every 30s, full local graph) and the insight cron (after `applyChains`/`applyCloseQuestions`). No merge/version/dirty-check. Because the host writes a goal root node on EVERY Act turn (`mesa_director.go:404-407`), every autonomous host's local graph is non-empty, so the host re-imports the cron's growth ONLY at cold-start (the `Empty()` guard). Loop: (1) cron loads the row, adds cross-entity chains + closes open questions, writes back; (2) within ≤30s the host's flush overwrites the row with its local graph (which lacks the cron's nodes) → chains and closures gone. The cron's growth NEVER reaches a warm host AND is continuously clobbered. Contrast `SyncKnowledge` which upserts PER-SUBJECT so distinct subjects survive — the goal graph lacks that granularity. Net: the entire Phase-4b insight→host goal-graph loop is dead for any non-cold-started host.
- **Fix:** Stop treating the graph as a clobberable single blob. (a) Node/edge-level MERGE — separate `goal_graph_nodes`/`goal_graph_edges` tables keyed by `(host_id,id)`/`(host_id,from,to,rel)`, upsert per row like knowledge; or (b) give the host a mid-run MERGE-import (non-replacing) that adds cron nodes/edges + applies closures without clobbering live host status, and gate the host's flush so it never deletes nodes it doesn't know about; partition authority (host owns goal/subgoal/status nodes, cron owns chain/closure nodes). At minimum add an `updated_at`/version guard so the host cannot overwrite a strictly-newer cron snapshot.
- **Confidence:** high. _(Merged: XC-persistence-spine + XC-goalgraph-lifecycle — identical root.)_

### H18 · Distillation crons' `Cache:true` prefixes are below the model minimum — never cached
- **Location:** `mesa/mesad/cron.go:423-435,472-484` (`consolidateSystem`, `Cache:true`, on Haiku) and `mesa/mesad/cron_insight.go:392-409,459-471` (`insightSystem`, `Cache:true`, on Sonnet); model wiring `cmd/mesad/main.go:53-54`.
- **Problem:** Both crons mark their stable analyzer prefix `Cache:true` expecting prompt-caching to carry fleet cost. But the cached prefixes are far too SHORT to cache. `consolidateSystem` is ~381-509 tokens on `claude-haiku-4-5` (4096-token minimum cacheable prefix); `insightSystem` is ~417-556 tokens on `claude-sonnet-4-6` (2048 minimum). A `cache_control` breakpoint below the model minimum is SILENTLY ignored by Anthropic (no error; `cache_creation_input_tokens` stays 0). So neither cron ever gets a cache write or read — every consolidation (one Haiku call per active host per 60s tick, up to `MaxHostsPerSweep=64`) and every insight call (one Sonnet call per host per 180s) pays full input price on the analyzer prefix forever. The §3.6 "cheap × volume" cost model the 200-host go-live rests on does not hold. (Not a tiering violation — Haiku/Sonnet are correct and there's no Opus-on-bulk and no per-item fan-out; it's a silent cost-model defeat the SDK-free client emits without any min-length check.)
- **Fix:** (a) Restore the cache by making the cached prefix exceed the model minimum — prepend the large deterministic `dslManual` (or a shared static manual) as `block[0]` so the prefix clears 4096 (Haiku) / 2048 (Sonnet), the same trick Act already uses (`act.go:30`); or (b) drop `Cache:true`, stop pretending, and re-derive the 200-host budget from full-input pricing (re-tune `ConsolidateEvery`/`BatchSize`/`InsightEvery`). Independently, instrument the llm client to surface `usage.cache_read_input_tokens`/`cache_creation_input_tokens` so a silent zero-hit prefix shows up in telemetry, not the bill.
- **Confidence:** high.

### Cross-cutting — doc fidelity

### H19 · §4 "failure → cause-attributed blocked_by + enabling sub-goal" is a generic placeholder host-side
- **Location:** `runtime/limbic.go:110-125` (own-death handler — no goalgraph writer) + `runtime/mesa_director.go:143-160` (`markGoalBlocked`, generic).
- **Problem:** §4's worked example promises "Dying to the bear marks `mine --blocked_by--> bear` → spawns `train-combat --enables--> survive-at-mine`," and §9 Phase 2 lists "failure → spawn enabling sub-goal (bear → combat)" as a host-side deliverable. On the host's OWN death, the only writes are affect (`OnDeath`) + grievance to the PK killer; nothing touches the goal graph. The only deterministic blocked-edge writer, `markGoalBlocked`, fires on a generic fail-STREAK (not on death) and creates a generic "how-to-progress:" open question + a placeholder enabler labelled "Find and remove what is preventing …" — it never attributes the cause (the bear) nor writes a cause-named `blocked_by` edge or a `train-combat` enabling sub-goal. So the keystone "failures become generative" behavior materializes host-side only if the Sonnet insight cron happens to emit that exact chain.
- **Fix:** Either build the host-side death→goalgraph writer §9 Phase 2 calls for (on own-death while pursuing a survival/grind goal, write `goal --blocked_by--> <cause>` using the engaged-NPC/last-damager attribution `pkKiller`-style logic already gathers, and spawn the enabling sub-goal), or update §4/§9 to state cause-attributed failure edges are produced only by the mesa insight cron and the host-side path is a generic placeholder.
- **Confidence:** high.

### orsc-ctl

### H20 · Flag-after-positional is silently broken across the whole CLI
- **Location:** `cmd/orsc-ctl/player_presence.go:35,67,129,178,227,261,316`; `player_stats.go:28,61,106,143,176`; `player_items.go:134,174,204` — every per-command `fs.Parse(args)`.
- **Problem:** Every handler calls `fs.Parse(args)` directly, and Go's `flag` package STOPS parsing at the first non-flag token. Because every command takes `<username>` (or another positional) FIRST, any flags AFTER the positional are never parsed — they land in `fs.Args()` and are dropped. The tool's OWN documented invocation (`main.go:12`: `orsc-ctl player mute Delores -minutes 30`) hits this. Proven: `mute Delores -minutes 5 -shadow` sends `{minutes:30, shadow:false}` with NO error (user asked for a 5-min shadow mute, got a 30-min normal mute); `kick Delores -reason spamming` sends the default reason; `restore Delores -all-skills=false` sends `allSkills:true`. These are silent wrong-privileged-action bugs against a live admin API. Required-flag commands fail the other way (`teleport Delores -x 122 -y 503` errors "flag is required" though the flag was passed). Only world announce/system-message escape it.
- **Fix:** Parse flags independent of position — split the leading positional(s) off the front before parsing (or use a permutation helper / the `pflag` library which does POSIX interspersed parsing). A shared helper that pulls required positional(s) then ParseFlags over the rest is cleanest. The silent-wrong-body commands (mute/kick/restore/summon/item) must be FIXED, not just documented.
- **Confidence:** high.

---

## MEDIUM

### M1 · `Observe` cannot upgrade a belief's provenance (verified hearsay stays "hearsay")
- **Location:** `cognition/knowledge/knowledge.go:135-162` (`Observe`) vs `:121-122` (`Note` provenance upgrade).
- **Problem:** `Observe` never touches `Provenance`, so direct observations can't upgrade a belief's provenance the way restated `Note`s do. A hearsay belief reinforced by many DIRECT observations keeps reporting `provenance=hearsay`. Confirmed: `Note("sells X", ProvHearsay, 0.6)` + 10× `Observe(true)` leaves `Provenance="hearsay"` at confidence 0.96. `mesa_director.go:856-859/1022-1040` surfaces `best.Provenance` to the LLM as the literal prose "someone told you" vs "you saw it yourself," so a thing the host personally verified dozens of times is presented as unverified hearsay — the exact failure the provenance ranking exists to prevent.
- **Fix:** Give `Observe` a provenance argument, or on `good && existing belief` upgrade via `provenanceRank`: `if provenanceRank(ProvObserved) > provenanceRank(e.Beliefs[i].Provenance) { e.Beliefs[i].Provenance = ProvObserved }`.
- **Confidence:** medium.

### M2 · `knowledge.Import` is not a deep copy (slice aliasing)
- **Location:** `cognition/knowledge/knowledge.go:248-256` (Import).
- **Problem:** `e := rows[i]` copies the Entry by value but the stored `*Entry` shares the SAME backing array for `e.Beliefs`/`e.Tags` as the caller's `rows[i]`. Confirmed: after `l.Import(src)`, mutating `src[0].Beliefs[0].Alpha` changes the value the ledger returns. Asymmetric with the rest of the package (`view()`/`Export()` clone via `append(nil, …)`). Latent today (fed by a fresh `json.Unmarshal`), but a real foot-gun for any future caller that Imports an `Export()`'d snapshot from a still-live ledger.
- **Fix:** After `e := rows[i]`, clone: `e.Beliefs = append([]Belief(nil), rows[i].Beliefs...)` and `e.Tags = append([]string(nil), rows[i].Tags...)`.
- **Confidence:** high.

### M3 · `goalgraph.Import` is not a deep copy (Tags aliasing)
- **Location:** `cognition/goalgraph/goalgraph.go:309-318`, esp. `:315` (`g.nodes[norm(n.ID)] = &n`).
- **Problem:** `n := s.Nodes[i]` copies the Node by value but the copied `Tags` slice header ALIASES the caller's `Snapshot.Nodes[i].Tags`. Confirmed via probe: importing a snapshot then mutating `snap.Nodes[0].Tags[0]` changes the graph node's tag (and two graphs importing the same snapshot share backing arrays). Violates the package's own deep-copy invariant that `cloneNode` (`:285-292`) guarantees everywhere else.
- **Fix:** `n := s.Nodes[i]; n.Tags = append([]string(nil), n.Tags...); g.nodes[norm(n.ID)] = &n`. (Edges are pure value types, fine as-is.)
- **Confidence:** high.

### M4 · Firehose idempotency key collapses distinct same-second lines
- **Location:** `runtime/observation.go:45` (`IdempotencyKey = kind|subject|now.Unix()`) + `mesa/mesad/ltm.go:326-334` (`ON CONFLICT … DO NOTHING`).
- **Problem:** Two DISTINCT perception lines sharing kind+subject within one wall-clock second collapse to a single row — only the first text survives; the rest are silently dropped server-side. Worst case: `SystemMessage`/`ChatReceived(no speaker)` always emit `subject="server"` (`perception.go:96,99`), so any burst of system/quest messages in the same second (tutorial prompts, multi-line quest text) loses all but the first from the speed-3 firehose — the exact substrate the Phase-4 distillation crons chew on. (The reactive speed-2 path keeps all lines, so reactive extraction is unaffected.)
- **Fix:** Append a short stable hash of the text for capture kinds: `fmt.Sprintf("%s|%s|%d|%x", kind, subject, now.Unix(), fnvHash(text))` for dialog/chat/server; keep the coarse key for genuinely once-per-event outcome/transaction/location. Update the mesad fallback key (`ltm.go:320`) to match.
- **Confidence:** high.

### M5 · `nearestNamedNpc` returns a non-deterministic NPC (random map order)
- **Location:** `runtime/perception.go:292-302` (`nearestNamedNpc`); `world/world.go:354-362` (`NpcsState.All`).
- **Problem:** Ranges over `h.world.Npcs.All()` which iterates the underlying map in Go's randomized order — returns NO stable element despite the "takes the FIRST named hit" comment. With multiple named NPCs in view the result is non-deterministic across calls. This feeds shop-attribution (`lastNpc`/`perceiveShop`'s subject, `perception.go:69,124`) and the npc_dialog firehose subject + reactive window key, so "where do I buy X" beliefs and per-speaker dialog windows scatter across whoever the map yielded — violating the host-LIGHT deterministic invariant.
- **Fix:** Iterate a deterministic order (`NpcsState.Order()` exists at `world.go:184`), or better compute the TRUE nearest by Chebyshev distance to `w.Self.Position()` so "nearest named NPC" actually is the nearest.
- **Confidence:** high.

### M6 · False auto-completion of unmet acquire goals (loose substring match)
- **Location:** `runtime/mesa_director.go:227` (`acquireVerbs`) + `:236-266` (`goalSatisfiedByInventory`).
- **Problem:** `acquireVerbs` uses bare substrings ("have a", "have an", "find a", "find an") matched with `strings.Contains` over the whole goal text, so they fire on unrelated words: "have all the bronze bars" matches "have a" (inside "have all"), "have any reason" matches "have an", "find another spot" matches "find an". Combined with the loose item-name substring test (`:261`), "have all the bronze bars" with a single "Bronze Bar" in inventory satisfies → `markGoalDone` fires with the goal NOT met. Verified all four phrasings classified as acquire goals.
- **Fix:** Anchor on word boundaries: match `" have a "`, `" find a "` with trailing space, or a regex `\b(have|find|own)\b a\b`, or a small set of leading anchors (`acquire `/`obtain `/`buy `/`get `). Drop the bare "have a"/"find a"/"have an"/"find an".
- **Confidence:** high.

### M7 · `goal_op` accepted unvalidated + unlogged (done-synonyms silently ignored)
- **Location:** `mesa/mesad/act.go:641` (`parseMove`) + `:54-56` (act log); `runtime/mesa_director.go:191-220` (`applyGoalOp`).
- **Problem:** `goal_op` is a free-form string with NO validation/logging. `parseMove` lowercases/trims it; `applyGoalOp`'s switch recognizes only "done"/"abandoned"/"adopt" and silently routes EVERYTHING else into the still-active default branch. The Act prompt says the completion token is exactly "done" (`act.go:563`), but Sonnet routinely emits near-synonyms ("complete"/"completed"/"finished"/"satisfied"). Any of those means the goal is NEVER marked done — the host re-plans a finished goal, and because `goal_progress` is ~1.0 that turn the default branch writes progress 1.0 while leaving Status=open, which trips the spin/SPINNING detector (the symptom that triggers the "goal is already done" nudge is *caused* by the host mis-handling the done-declaration). The `act` log omits `goal_op`/`goal_text`/`goal_progress`, so the failure is invisible in mesa logs.
- **Fix:** Validate `goal_op` against `{"","done","abandoned","adopt"}`; map common synonyms (complete/completed/finished/satisfied → done) or reset an unrecognized value to "" and Warn. Add `goal_op`/`goal_text`/`goal_progress` to the act Info log line.
- **Confidence:** high.

### M8 · `markGoalBlocked` re-OPENs an already-resolved open question (+ un-reset attempts cap mutes re-asking)
- **Location:** `runtime/mesa_director.go:143-160` (`markGoalBlocked`) vs `runtime/reactive.go:436-459` (`closeResolvedQuestions`) / `mesa/mesad/cron_insight.go:595-617` (`applyCloseQuestions`); attempts map in `runtime/speech.go:228-237`.
- **Problem:** `markGoalBlocked` unconditionally re-OPENs its open question every fire: `Upsert(qid, KindOpenQuestion, reason, StatusOpen)` (`:148`) and `Upsert` forces status when status!='' (`goalgraph.go:121-123`). `qid` is deterministic (`"how-to-progress:"+g`). So after the reactive/cron closers flip it `StatusDone`, a later hard-fail/spin re-sets it `StatusOpen` and re-blocks the goal. Worse, the speech `attempts` counter (keyed by qid) is NOT reset on re-open, and `gcSpeech` only clears it after 15m — so a re-opened question whose attempts already hit `maxAskAttempts(3)` is permanently ask-ineligible for up to 15 minutes even though the host flagged it a fresh blocker.
- **Fix:** Have `markGoalBlocked` NOT force `StatusOpen` on an existing resolved question (pass `''` for status and `SetStatus(StatusOpen)` only when the node is absent/abandoned); if intentional re-open is desired, reset `attempts[qid]` and strip resolved-by-ask/ask-exhausted tags on re-open so it's actually re-askable.
- **Confidence:** medium.

### M9 · Server-side question closure can fire on TOPIC OVERLAP, not an ANSWER
- **Location:** `runtime/reactive.go:436-459` (`closeResolvedQuestions`).
- **Problem:** The closer flips an open question Done whenever any claim with `Confidence>=0.6` shares a single significant (≥4-char, non-stopword) token with the question label via `goalTouch(low, c.Claim) || goalTouch(low, c.Subject)`. NPC/server claims default to 0.85, so they clear the floor. An authoritative NPC line that merely MENTIONS the topic — "don't trust the pickaxe seller, he scams" or any `subject=="pickaxe"` claim — closes "where to buy a pickaxe" and marks its goal Active, a false resolution that silently ends the ask-drive. Because writeback sets `subject:=speaker` when empty, a claim about the NPC itself can also accidentally match.
- **Fix:** Require overlap on the question's salient topic specifically (reuse `salientTopic` on `q.Label`) AND that the claim reads as an answer (require the topic word in the CLAIM not just the subject, and ≥2 overlapping significant tokens). Add a test that a topical-but-non-answering authoritative claim does NOT close the question. (Pairs with H7 — same function.)
- **Confidence:** high.

### M10 · `combatRoundTarget` PvE fallback can blame a player at the same numeric index
- **Location:** `runtime/limbic.go:307-330` (`pkKiller`); index lookup `:321`; writes `runtime/host.go:779-797,920-924`, `runtime/combat.go:461-468`.
- **Problem:** Beyond the race (see H11), `pkKiller` passes `combatRoundTarget` — which holds an NPC index for the common PvE case (`actions_combat.go:23/36/59/65` set it from `npcView.record.Index`) — to `h.world.Players.Get(idx)` (`:321`), looking up an NPC index in the PLAYER store. RSC NPC and player indices are separate spaces sharing numeric ranges, so on a normal PvE death a skulled bystander player occupying that same numeric index gets blamed: `ObserveGrievance` + "ganked-me" against an innocent party. The `SkullType` guard only narrows, doesn't prevent the collision.
- **Fix:** Prefer dropping the fallback — attribute melee PK ONLY from `self.EngagedPlayerIndex` (genuinely player-scoped). If a fallback is kept, track the last engaged PLAYER index separately (`lastAttackedPlayerIndex` exists) and use only that.
- **Confidence:** medium.

### M11 · Reactive cron firehose skips high-salience non-firehose kinds (false fallback comment)
- **Location:** `mesa/mesad/cron.go:277-294,350-352`.
- **Problem:** The comment claims "if the batch is entirely non-firehose kinds, fall back to processing it all so the cursor still advances," but the code NEVER falls back — it filters to `firehoseKinds` and, on an empty `batch`, skips the LLM block and advances the cursor to `maxAt` of the FULL read. In a MIXED batch the non-firehose rows are dropped from `batch` AND the cursor steps past them. The host emits high-salience non-firehose kinds the cron should distil: `transaction`(0.8 — the shop catalogue, the literal "where to buy X" signal), `outcome`(0.7 — kills), `location`(0.5). These never get the Haiku distillation pass. Mitigated (not fully lost) because those kinds also write the host's local knowledge ledger, but the cron's cross-line dedup/triage over them is skipped and the comment is materially false.
- **Fix:** Either add transaction/outcome/location to `firehoseKinds`, or actually implement the documented fallback (process `obs` wholesale when the filtered batch is empty but `obs` is non-empty). At minimum fix the comment.
- **Confidence:** high.

### M12 · Episode embeddings are write-once with no back-fill — unembedded rows invisible to semantic recall
- **Location:** `mesa/mesad/ltm.go:303-312` (`Add`, `ON CONFLICT DO NOTHING`) + `:356-367` (`recallCosine`, `WHERE … embedding IS NOT NULL`).
- **Problem:** When `l.embedder==nil` OR the Voyage Embed call fails, the episode is stored with `embedding=NULL` and stays NULL forever (no re-embed path anywhere). `Recall()` takes the cosine path EXCLUSIVELY for any non-empty query when an embedder is wired (`:346-352`, no union with lexical). The host's primary runtime recall passes non-empty query text (`runtime/memory_journal.go:290`), so any episode written before an embedder was wired or that hit a transient Voyage error becomes permanently invisible to semantic recall. The "never lose a memory to a transient API error" comment is misleading — the row survives but is functionally lost for semantic recall.
- **Fix:** On Add re-send refresh a missing embedding: `ON CONFLICT (host_id,idem_key) DO UPDATE SET embedding = COALESCE(EXCLUDED.embedding, episodes.embedding)`; OR add a Tier-0 backfill cron for `embedding IS NULL`; OR make `recallCosine` union with `recallLexical` so unembedded rows are reachable.
- **Confidence:** high.

### M13 · Confidence unclamped at the act-prompt boundary (downstream consumers disagree)
- **Location:** `mesa/mesad/act.go:455-460` (`parseExtractedDialog`, unclamped) feeding `runtime/reactive.go:410-417` vs `:443`.
- **Problem:** `parseExtractedDialog` copies the model's claim Confidence straight into the proto with no clamping. `reactive.go:411` RE-CLAMPS for the ledger write but `:443` (`closeResolvedQuestions`) tests the ORIGINAL unclamped value. (a) A claim returned with confidence 0 is written at 0.85 yet REFUSED by closure (0 < 0.6); (b) a bogus out-of-range 5.0 is written sanely (clamped) but PASSES closure (5.0 ≥ 0.6). The single source of truth should be clamped once at the boundary. (Root cause shared with H7; this is the durable upstream fix.)
- **Fix:** Clamp confidence to `[0,1]` in `parseExtractedDialog` before writing into the `DialogClaim` proto, so both ledger-write and question-closure see the same value.
- **Confidence:** high.

### M14 · `MesaDirector.visited` is an unbounded lifetime tile-history map
- **Location:** `runtime/mesa_director.go:1283` (write), `:74` (decl), `:1395-1402` (read).
- **Problem:** `visited` is an in-RAM `map[[2]int]bool` recording EVERY unique tile the host has ever stood on (`describeArea` runs every turn). Never pruned/capped/reset (death/displacement do not clear it). The host is an open-ended wanderer, so over a long session it walks tens of thousands of distinct tiles and the map grows without bound for the host's lifetime — a direct "no unbounded in-RAM growth" violation on the per-turn director path. Its only use, `doorUsed()`, needs only a small recency window.
- **Fix:** Replace the unbounded map with a bounded recency structure (ring buffer of the last ~64 visited tiles, or an LRU), or retain only tiles within a radius / since the last displacement. `doorUsed` only checks the 5 tiles around a door.
- **Confidence:** high. _(Merged into XC-host-light.)_

### M15 · Firehose `emitObservation` spawns unbounded goroutines/RPCs (no concurrency cap)
- **Location:** `runtime/observation.go:52-63` (`emitObservation`); callers `runtime/perception.go:174,198,220,235`.
- **Problem:** `emitObservation` launches one `go func()` per observation that clears the 0.3 salience floor, each opening a fresh client-streaming gRPC `RecordObservations` call, with NO concurrency cap — unlike the sibling `spawnExtract` which IS semaphore-capped at `reactiveMaxInflight=3` (and whose comment misleadingly claims this path "mirrors" it). The firehose is only salience-gated, so a fast-movement / chat-storm burst fans out into many concurrent goroutines/streams at once. Each is bounded by a 5s timeout so it self-drains (hence medium not high), but under sustained load at the 200-drone scale it's unbounded concurrency against mesa.
- **Fix:** Add a buffered semaphore on the host (mirroring `reactive.inflight`); non-blocking acquire, DROP the observation when full (a single firehose line isn't worth queueing). Fix the misleading `reactive.go:346` comment.
- **Confidence:** high. _(Merged: host-persistence + XC-concurrency + XC-host-light.)_

### M16 · Analysis-mode dry-run writes to the live goal graph (freeze leak)
- **Location:** `runtime/analysis.go:335-378` (`runHypothetical`/`analysisSituation`) → `runtime/mesa_director.go:404-408` (`situation` Upsert) and `:660` (SetObjective).
- **Problem:** Analysis mode is specified to FREEZE all learning I/O, and every host-side learning writer gates on `h.AnalysisActive()`. But the HYPOTHETICAL dry-run path does not: `runHypothetical → analysisSituation` builds a throwaway `MesaDirector` and calls `d.situation(h, Outcome{})`, and `situation()` unconditionally WRITES to the shared `h.goalGraph` (`:406` Upsert when the goal node is absent) and `h.journal.SetObjective` (`:660`), with no `AnalysisActive()` guard. `effectiveGoal()` can also mutate `d.goal` and call `selectNextGoal → SetStatus(Active)`. So an operator "what would you do" during a freeze creates a phantom active goal node / flips an open_goal active in the in-RAM graph the host resumes with — a write the operator was told would not happen. (Flushes are AnalysisActive-gated so it doesn't reach durable storage, but the in-RAM state is corrupted.)
- **Fix:** Guard the goal-graph writes in `situation()` (at minimum the Upsert at `:406` and SetObjective at `:660`) behind `if !h.AnalysisActive()`, or give `analysisSituation` a read-only "preview" path that builds the Situation without any goalGraph/journal writes.
- **Confidence:** high. _(Merged: XC-host-light + XC-goalgraph-lifecycle.)_

### M17 · Knowledge per-subject LWW: warm host clobbers cron reconciliation
- **Location:** `runtime/knowledge.go:87-101` (`flushKnowledgeToMesa`) + `mesa/mesad/ltm.go:546-586` (`SyncKnowledge`, per-subject replace) + `runtime/limbic.go:44-49` (Empty-guarded bootstrap).
- **Problem:** Knowledge has the same two-writer LWW collision as the goal graph (H17), scoped per-subject. Host and crons write the SAME subjects (the reactive tier and cron extract claims about the same shops/NPCs). `SyncKnowledge` is `ON CONFLICT (host_id,subject) DO UPDATE SET beliefs_json = EXCLUDED.beliefs_json` — a full per-subject replace. The cron does read-modify-write (including the insight reconcile β-bumps); the host then pushes its OWN local copy of that subject (which never saw the reconciliation) and LWW-overwrites the cron's work. The host re-imports cron knowledge only at cold-start (`All()==0`), so a warm host never benefits AND clobbers it.
- **Fix:** Merge α/β per `(subject,claim)` on the SQL upsert instead of replacing `beliefs_json` wholesale (read the current row in the same tx and fold), and have the host periodically re-import the merged ledger (not just Empty-guarded). Until the 4b merge lands, at least don't let the host's flush overwrite a subject the cron more-recently touched (compare `updated_at`).
- **Confidence:** high.

### M18 · Persona escalation gate is self-referential — `EscalateThreshold` is a near-inert knob
- **Location:** `mesa/mesad/cron_insight.go:299-302`.
- **Problem:** `drive = 0.6*bulk + 0.4*flavor`; `adjusted = EscalateThreshold*(1.5 - drive)`; `return drive >= adjusted`. Because `drive` appears on BOTH sides, the inequality collapses to `drive >= EscalateThreshold*1.5/(1+EscalateThreshold)` — for the default 0.5 it is exactly `drive >= 0.5`, independent of any item property. There is no independent item-salience term measured against the bar (`qItem` has no salience field), so the comment's framing never actually governs admission. The knob shifts the cutoff monotonically but weakly/non-linearly and does not implement the documented contract. Verified across threshold 0.2/0.5/0.9 in a standalone replica.
- **Fix:** Decide the intended contract and implement it without `drive` on both sides. Either compare a FIXED reference (a constant, or a real per-item salience once 4c lands) against the persona-scaled bar, or make it an explicit `drive >= EscalateThreshold` gate and drop the `(1.5-drive)` scaling. Update the comment + the `cron.go:41` "gate center" docstring. Add a mid-drive test case.
- **Confidence:** high.

### M19 · Cron LTM read errors are swallowed → reconcile overwrites a subject's real DB state
- **Location:** `mesa/mesad/cron_insight.go:234-242,259-271`.
- **Problem:** `Knowledge()`/`GoalGraph()` errors are ignored (the `kerr==nil`/`gerr==nil` guards skip the import), so on a transient DB read error `led` starts EMPTY. `applyReconcile` then does `led.Observe(subj, loser, false, …)` on the empty ledger, creating a fresh belief carrying only the negative evidence, and `SyncKnowledge` (UPSERT keyed on host_id,subject) OVERWRITES that subject's row with the impoverished single-belief version — silently dropping the subject's richer DB state. The cursor then advances, so the loss is permanent for that subject. Low probability (needs a transient read error coinciding with a reconcile op) but real corruption.
- **Fix:** Treat a `Knowledge()`/`GoalGraph()` read error as a hard, retryable failure for the tick (bump FailCount, do NOT advance the cursor, return the error) — mirroring the LLM/persist failure handling. Only proceed when the prior state loaded successfully.
- **Confidence:** medium.

### M20 · Distillation crons send an empty system block on a cold host → 400 (wedges consolidation)
- **Location:** `mesa/mesad/cron.go:464-475` (`extractClaims` ctxBlock) and `mesa/mesad/cron_insight.go:448-462` (`runInsight` ctxBlock).
- **Problem:** Both crons build a second uncached system block from `ctxBlock.String()` and pass it unconditionally. For a host with no persona prose AND no existing knowledge — a freshly-registered cold host on its FIRST consolidation — `ctxBlock` is empty, so the request sends `{type:"text", text:""}`. The Anthropic Messages API rejects an empty text block with 400. The cron treats this as retryable, bumping `fail_count` and re-reading the same batch each tick, so a cold host wedges its consolidation for `maxFailBeforeSkip=3` ticks (~3 min) burning a failing Haiku call each tick before the poison-batch guard advances the cursor PAST real un-consolidated observations (data loss for that batch).
- **Fix:** Append the second `SystemBlock` only when `ctxBlock.Len() > 0`, or guard the llm client to drop any block whose trimmed Text is empty. Mirror in both `extractClaims` and `runInsight`.
- **Confidence:** medium.

---

## LOW

### L1 · `knowledge.Import` silently last-writer-wins on normalized-subject collisions
- **Location:** `cognition/knowledge/knowledge.go:251-255` + `:90-101`.
- **Problem:** Import keys by `normalize(e.Subject)` with no collision handling — two snapshot rows normalizing to the same key (e.g. a merged cross-host snapshot) silently overwrite each other, no merge/error/log. Only bites externally-merged/hand-built snapshots (mesa bootstrap), so a lossy import is unobservable.
- **Fix:** Merge on collision (append/merge Beliefs, max Beta evidence + Encounters) or at least `if _, dup := l.rows[key]; dup { /* log/merge */ }`.
- **Confidence:** high.

### L2 · `goalgraph.Upsert` silently merges on normalized-id collision (Kind/Label/Status clobber)
- **Location:** `cognition/goalgraph/goalgraph.go:96-107` (`nodeLocked`), `:111-126` (`Upsert`).
- **Problem:** Two distinct ids collapsing to the same normalized key merge into ONE node with last-writer-wins on Kind/Label/Status and no log. Because adopt-goals are keyed on raw LLM free-text, a re-adopt of "Mine ore" vs existing "Mine Ore" silently mutates the existing node's kind/label/status (a goal can be silently demoted Goal→Subgoal). Usually benign dedup, but the silent Kind/Status overwrite is surprising and unlogged.
- **Fix:** Don't downgrade Kind on collision (only set Kind when transitioning to a more-specific kind), or log/skip a kind change. At minimum document the silent merge.
- **Confidence:** high.

### L3 · `goalgraph` edge endpoint strings can differ from the node's canonical ID
- **Location:** `cognition/goalgraph/goalgraph.go:141` (Link stores `TrimSpace(from/to)`) vs `:100` (Node.ID stores first-writer's `TrimSpace(id)`).
- **Problem:** `Edge.From/To` persist the caller's case/spacing while `Node.ID` persists the first creator's, so for the same normalized key the two display strings can DIFFER. All in-package traversal re-normalizes, so it's internally tolerated — risk is only a future SQL/LTM string-JOIN of `Edge.From/To` against `Node.ID` literally, which would orphan edges. Latent persistence foot-gun.
- **Fix:** Have `Link` resolve from/to through `nodeLocked` and store the node's canonical `ID` on the Edge.
- **Confidence:** high.

### L4 · ShopClosed cleanup is short-circuited by the analysis-mode early return (stale slate)
- **Location:** `runtime/perception.go:41-52`.
- **Problem:** `perceptionHandle` returns immediately under `AnalysisActive()` (`:42`), short-circuiting the `ShopClosed` case (`:50-52`) that clears `shopStock`/`shopSubject`. If analysis is entered while a shop is open and `ShopClosed` arrives during the freeze, the snapshot is never cleared; after analysis exits, the next `ShopOpened` with the same subject inherits the previous shop's per-item dedup slate — suppressing legitimate "sells X" writes. Narrow (requires an analysis toggle straddling a shop session + matching subject).
- **Fix:** Handle `ShopClosed` state-reset before/outside the `AnalysisActive()` early-return (or clear the snapshot on `EnterAnalysis`).
- **Confidence:** medium.

### L5 · Stale-goal narration gap + repeated inventory-satisfied log on a no-successor done goal
- **Location:** `runtime/mesa_director.go:119-132` (`selectNextGoal`), `:100-112` (`effectiveGoal`), `:472-477` (satisfaction net), `mesa/mesad/act.go:477`.
- **Problem:** When the only active goal completes and there's no successor, `effectiveGoal` leaves `d.goal` pointing at the now-Done goal. `act.go:477` renders `GOAL: <done goal>` with no done annotation (the planner is told its goal is still active and may keep trying it), and `goalSatisfiedByInventory` re-fires every turn re-logging "goal completed (inventory-satisfied)" indefinitely (markGoalDone is idempotent, the log line is not).
- **Fix:** When `effectiveGoal` would return a Done/Abandoned goal with no successor, return "" for `sit.Goal` (so act.go omits the GOAL line) or surface an explicit DONE intention hint. Guard the satisfaction-net log to fire once (check status transition before `markGoalDone`). (Closely related to H1.)
- **Confidence:** high.

### L6 · `hints["displacement"]` is written + transported but never read by mesad
- **Location:** `runtime/mesa_director.go:597-605` vs `mesa/mesad/act.go` (no reader).
- **Problem:** `situation()` writes `hints["displacement"]` and grpc copies the whole Hints map, but `act.go` never reads the key (every other hint key is read). The displacement info DOES still reach the planner via `Trigger`, so this is dead/redundant code, not data loss — but misleading: a future refactor trimming the trigger could silently drop the signal believing the hint covers it.
- **Fix:** Delete the `hints["displacement"]` block (Trigger already carries it) or add a reader and remove the Trigger duplication. Pick one source of truth.
- **Confidence:** high.

### L7 · Self-chat fans into ALL latched windows (cross-contaminates multi-conversation extraction)
- **Location:** `runtime/reactive.go:212-224` (`appendToLatched`) / `:275-284` (`reactiveObserveSelf`).
- **Problem:** The host's own outbound chat fans into EVERY currently-latched window (up to `reactiveMaxLatches=4`), not just the speaker being replied to, and refreshes all their latch TTLs. A reply intended for A is appended to B/C/D's windows; the next `ExtractDialog` for B mis-pairs A's reply as part of B's exchange. The single-conversation case is correct; multi-latch cross-contaminates.
- **Fix:** Pass the target speaker through `Say→reactiveObserveSelf` and fan the self-line only into that speaker's window; at minimum don't refresh other speakers' latch TTLs off a reply aimed at one.
- **Confidence:** medium.

### L8 · `groundReply` teach branch burns the per-player cooldown even when grudge-suppressed
- **Location:** `runtime/speech.go:433-440`.
- **Problem:** `teachable(from, now)` is evaluated inside the `&&` chain at `:433` and has a side effect (records `lastTeach[from]=now`, returns true). The grudge suppression check is applied AFTER at `:437`. So when a teachable belief exists and the cooldown is clear but the sender is resented, the per-player teach cooldown is BURNED for 5 min though no teach line is emitted; if the grievance decays below 0.5 within that window the host is wrongly suppressed.
- **Fix:** Compute the suppression decision first and call the side-effecting `teachable()` only when the teach will actually be emitted.
- **Confidence:** high.

### L9 · Goal-graph bootstrap Import clobbers a non-empty live graph (guard is caller-side only)
- **Location:** `runtime/goalgraph.go:67-79` + `runtime/limbic.go:50-56`; `cognition/goalgraph/goalgraph.go:308-318`.
- **Problem:** `goalgraph.Import` REPLACES the graph wholesale; `bootstrapGoalGraphFromMesa` calls it directly. Its only self-protection is the empty-FETCH bail — it will happily clobber a non-empty live graph if mesa returns a non-empty snapshot. The sole protection is the CALLER's `Empty()` guard at `limbic.go:50` (enforced by convention + comment only). A future second caller that forgets the guard silently wipes live intention state. Same latent shape for knowledge bootstrap (`All()==0` at `limbic.go:44`).
- **Fix:** Move the `Empty()`/`All()==0` guard INSIDE `bootstrapGoalGraphFromMesa`/`bootstrapKnowledgeFromMesa` so the no-clobber invariant is self-enforcing. Add a regression test that a non-empty live graph survives a non-empty mesa fetch.
- **Confidence:** medium.

### L10 · Curiosity tangent signal includes the goal's own required sub-goals
- **Location:** `runtime/mesa_director.go:746` (call site) + `:943-964` (`curiosityBias`).
- **Problem:** The explore/exploit gate is fed `hasTangent = len(serves) > 0 || len(subgoals) > 0`. Sub-goals are the goal's OPEN requires-children — the core-path steps it NEEDS — the OPPOSITE of a tangent. So a curious persona with pending required sub-goals (and no downstream `serves` edge) gets the "a short detour to learn something new is worth it" nudge, pulling it OFF its required steps. `TestCuriosityReadAtDecisionTime` only wires a `serves` edge, so the subgoals path was unexercised.
- **Fix:** Drop subgoals from the tangent signal: `hasTangent = len(serves) > 0`. If a tangent heuristic is desired, derive it from genuinely off-path nodes, not requires-children.
- **Confidence:** high.

### L11 · `capChat` caps on BYTE length, not RUNE count (truncates valid multibyte replies)
- **Location:** `mesa/mesad/act.go:573-586` (`capChat`).
- **Problem:** `capChat` caps on byte length (const max=80; `len(s)<=max`), but the RSC wire layer enforces the limit on RUNE count (`action/chat.go:40-43`, per commit `4d58de4`). Because `capChat` runs in mesa BEFORE host-side `sanitizeChat` folds multibyte chars to single ASCII bytes, a legal ≤80-rune reply with a few em-dashes/curly-quotes gets needlessly truncated. Verified: a 75-rune reply with 6 em-dashes (87 bytes) chopped to 72 runes though after sanitize it would be 75 single-byte chars. Not wire-unsafe, but silently drops the tail of valid replies.
- **Fix:** Cap on runes (walk `[]rune(s)`, cut at 80), ideally folding the same multibyte chars `sanitizeChat` handles, or move capping after sanitize. Fix the comment.
- **Confidence:** high.

### L12 · `hasGameAction` substring-scans raw DSL source (false positive from string literals/comments)
- **Location:** `mesa/mesad/act.go:169-195` (`hasGameAction`/`containsCall`).
- **Problem:** `validateMove` already parses `dsl_source` into an AST (`:84`), but `hasGameAction` re-scans the RAW source text with a substring matcher that doesn't strip string literals or comments. A narration-only routine — `note("then I attack( the goblin")` or a comment mentioning `go_to(` — yields a false positive for "takes a real game action" and passes the no-op guard, defeating the check meant to reject note/wait-only routines. Low impact (the routine no-ops on the host).
- **Fix:** Walk the already-parsed AST for actual call expressions whose callee is a PrimaryAction, instead of substring-scanning. Drops the brittle `containsCall`/`isIdentRune` helpers.
- **Confidence:** high.

### L13 · gcKnowledge "open-question"/"pinned" tag-pin branch is dead (nothing writes those tags)
- **Location:** `mesa/mesad/cron.go:605-609`.
- **Problem:** GC pins entries whose tags prefix "open-question" or equal "pinned", but NOTHING writes those tags onto a knowledge Entry (grep-confirmed; open_questions live in the goal graph). The pin branch is dead for those two tags — only `ProvSystem` provenance actually pins. Latent: an operator later relying on a "pinned" knowledge tag would be surprised it does nothing.
- **Fix:** Drop the dead tag branch, or document/emit the convention end-to-end and add a test that a "pinned"-tagged stale entry survives GC.
- **Confidence:** high.

### L14 · `SyncKnowledge` overcounts written rows; `nil` beliefs stored as JSON `null` not `[]`
- **Location:** `mesa/mesad/ltm.go:546-586` (count) + `:560` (`json.Marshal` of nil beliefs); column DEFAULT `'[]'` at `:153`.
- **Problem:** (a) `SyncKnowledge` returns `len(entries)` though it `continue`s past every blank-subject entry — the logged "subjects" count overstates rows written and masks a host pushing empty-subject knowledge. (b) `json.Marshal([]*KnowledgeBelief(nil))` returns literal `null`, not `[]`, so a belief-less entry is stored as `'null'::jsonb`, contradicting the column's `DEFAULT '[]'`; reads self-heal but downstream JSONB array queries would error on the `null` rows.
- **Fix:** (a) Track an explicit `n` incremented only on a successful Exec. (b) Normalize: `if e.GetBeliefs()==nil { bj=[]byte("[]") }`.
- **Confidence:** high.

### L15 · Persona generated-column migration can brick startup on a malformed `schema_version`
- **Location:** `mesa/mesad/ltm.go:226-264` (personas GENERATED columns) in `migrate()`.
- **Problem:** The persona projection adds STORED generated columns with hard casts (`schema_version int = (persona_json #>> '{schema_version}')::int`, the `::real` curiosity columns). On a live DB with a persona row whose JSON carries a non-numeric value, the `ALTER TABLE … ADD COLUMN … GENERATED ALWAYS AS ((…)::int) STORED` evaluates over existing rows and fails the whole `migrate()` (each stmt runs un-transactioned), bricking `OpenLTM` startup. Missing keys are safe; a present-but-malformed value is fatal. Latent landmine as persona JSON evolves.
- **Fix:** Use a NULL-tolerant cast (`CASE WHEN persona_json #>> '{schema_version}' ~ '^-?\d+$' THEN … END`) or store as text + cast at read time. Run `migrate()` statements in a transaction.
- **Confidence:** medium.

### L16 · orsc-ctl partial-coordinate footgun on summon
- **Location:** `cmd/orsc-ctl/player_presence.go:272-274` (summon), `:323-325` (summon-all).
- **Problem:** The coordinate branch is taken on `flagSet(fs,"x") || flagSet(fs,"y")` (OR). If the operator passes only `-x`, the missing coordinate is silently sent as 0 — `summon -x 50 Delores` sends `{x:50, y:0}`, teleporting to y=0 instead of erroring. Proven via test.
- **Fix:** Require both: if `flagSet(fs,"x") != flagSet(fs,"y")` return an error; only send x/y when both are set.
- **Confidence:** high.

---

## Positive control — the previously-missed CRITICAL is CLOSED

The 3b-class drop (affinity/grievance silently lost) is **closed at every layer** of this slice:
proto carries `affinity=7`/`grievance=8` (`mesa/proto/mesa.proto:332-333`); the mesaclient struct
carries them (`client.go:326-327`); `grpc.go` `SyncRelationships` SETS both (`:179-180`) and
`FetchRelationships` READS both (`:202-203`); the runtime converters map
`limbic.Entry.AffinitySum/GrievanceSum ↔ wire Affinity/Grievance` in BOTH directions
(`runtime/limbic.go:288`, `:338`); and the SQL layer persists+scans both columns
(`mesa/mesad/ltm.go:130-131,421-427,441,452`). The full host→proto→client→SQL→fetch→host loop
is lossless for the multi-axis fields. No residual drop found.

Two related lossy-round-trip traps remain LATENT (not yet biting because nothing host-side
consumes them) and are recorded so they don't become the next hidden 3b:
- **`relationships.value_traded` is a dead/clobbered field** — `ledgerToRelationships` always
  writes 0 (limbic.Entry has no `ValueTraded`), and `LTM.Add` never reads `ep.Relation` at all,
  so the episode `RelationDelta` (incl. `TotalValueTraded`) is silently discarded server-side.
  (`runtime/limbic.go:283-292`, `mesa/mesad/ltm.go:287-312,420-435`, `mesa/client/grpc.go:516-536`.)
  Fix: either delete the field, or wire it (limbic accumulator + `LTM.Add` applies `ep.Relation`).
- **Episode entity + importance dropped on the journal cold-start round-trip** — `Recall` builds
  `KnowledgeItem` with only Kind/Text/Provenance, and `bootstrapJournalFromMesa` hardcodes
  `Importance=0.6` and no entity, so a cold-start recall loses real importance weighting (affects
  `Salient()` ordering) and entity attribution. (`mesa/mesad/server.go:291-298`,
  `runtime/memory_journal.go:300-309`.) Fix: carry entity+importance back through `Recall`.

---

## Dropped low-confidence nits (3)

1. **`MetricsReport.at_unix` never set host-side** (wire-consistency) — the server substitutes
   `time.Now()` on `at_unix==0`, so the only effect is per-batch timestamps collapse to server-receive
   time, a minor analytics skew. Structurally harmless; documentable rather than a bug. Dropped.
2. **`Goal.Progress` fetched but discarded by `bootstrapGoalFromMesa`** (wire-consistency) — the
   auditor itself confirmed this is a deliberate design choice (progress is journal-derived) and the
   wire contract is consistent. Not a bug. Dropped.
3. **`AnalysisInterpret` returns a hard error on LLM failure** vs Chat/ExtractDialog's safe-value
   contract (mesa-act-prompts) — the host consumer already degrades gracefully
   (`runtime/analysis.go:225-226`), so it's a latent contract wart with no live impact. Dropped.

Additionally NOT counted as a product finding: the cron-consolidation auditor's **audit-process note**
about a broken untracked `zz_race_audit_test.go` (a `fakeStream` stub missing `SendHeader` that broke
the `mesa/mesad` test package's compilation). It is a CI/test-infra hygiene item, not a product bug:
ensure any `fakeStream` test stub implements `grpc.ServerStreamingServer[mesapb.Directive]` so the
package compiles and the test signal isn't silently disabled.

Also de-scoped to documentation-only (no code finding): two §-fidelity oversells —
the salience firehose is still the Phase-1 static `observationFloor=0.3` STUB (no novelty seen-set /
persona-curiosity weighting; `runtime/observation.go:18-29`), and the ASK drive's "host-owned
proactive clock" is in practice just the Act-turn `AgentThought` cadence (`runtime/speech.go:240`,
`mesa_director.go:378`). Both are correctly disclosed as stubs in code comments; the design doc
should be corrected to stop claiming a built capability. The `mind` inspector also omits
Affinity/Grievance (`debughttp/server.go:514-520`), which an operator needs to debug
"why won't this host ask the player next to it" — add the squashed read values to `mindRel`.
