# Corpus-mined gap-finding scenario candidates

> Source: wave-2 botting-corpus mining (IdleRSC / AutoRune / SBot / APOS-lineage
> TBoT+CS). These ~22 multi-step scenarios recreate canonical framework loops
> and are deliberately chosen to either (a) exercise a real DSL gap (see
> `build-backlog.md`) or (b) cover a perceive→decide→act loop **not** already in
> `cmd/scenariogen/scenarios_bots_proposed.yaml` (the existing 60-entry bot
> catalog).
>
> Each is a few-minute scenario with an assertable outcome; format follows
> `scenarios_bots_proposed.yaml` (setup admin commands + body + pass-assert +
> timeout). Candidates flagged **(FAILS today)** are intentional gap-finders —
> they should fail until the named gap is built, surfacing it concretely.
>
> **Already covered — do NOT re-propose** (exist as `bot-*` ids): power
> woodcut/mine/fish-drop, smelt-bronze, fletch chains, cook-bulk, spin-wool,
> cut-gems, gold-ring, clean/mix-potion, high-alch, autofighter
> (rat/eat/bury/loot/retreat/ranged/magic/prayer/style/multi-kill), stall-steal,
> pickpocket, chest-loot, kill-and-sell, buy-low, bank-runs, deposit-all-keep,
> agility, door-key.

---

## Group A — Fatigue / sleep (uncovered theme; targets the NEW fatigue→sleep gap)

1. **fatigue-sleep-resume-woodcut** — set fatigue high (`fatigue {self.name} 99`),
   give a sleeping bag, power-chop; assert the loop sleeps and resumes.
   **(FAILS today → surfaces the missing sleep verb + `is_sleeping` stub.)**
2. **fatigue-poll-stops-skilling** — gate a mining loop on `self.fatigue < 90`;
   assert it halts near the cap. Exercises the `self.fatigue` read path under load.
18. **alch-while-fatigue-managed** — alch a stack of items while sleeping when
   fatigued; combines the (covered) alch loop with the sleep gap → composite
   gap-finder for fatigue→sleep.

## Group B — Bot-to-bot trade / duel (uncovered theme; two-host orchestration)

3. **bot-to-bot-trade-give-item** — drone A summons drone B, `trade.request`,
   offer 100 coins, both accept+confirm; assert A's coins decreased and B's
   increased. Exercises the full twin-accept flow end-to-end between two live
   hosts (committed set only has decline/initiate, not a completed exchange).
4. **bot-to-bot-trade-swap** — A offers logs, B offers ore; assert both
   inventories swapped. Two-sided offer + `on trade_other_offer` reactor.
20. **two-drone-duel-stake-and-payout** — A and B duel with a coin stake, A wins;
   assert A gained B's stake. Full duel offer+confirm+combat between hosts
   (committed set has duel-initiate only, no staked completion).

## Group C — Combat / health perception (targets the opcode-234/104 combat gap)

5. **death-recovery-relog-resume** — fighter at 1 HP attacked to death; assert
   `on death()` fires, respawn at Lumbridge, then re-acquire and resume.
   Death-loop is an uncovered theme.
10. **fighter-eat-at-hp-fraction** — attack an aggressive NPC, eat when
   `self.hp_fraction < 0.5`; assert never died and food consumed. Confirms the
   **player-HP** path works even while combat-engagement is stubbed (isolates
   gap #1 — auto-eat is possible without combat-engagement decode).
11. **fighter-detect-kill-via-retarget** — kill 3 NPCs detecting each kill by
   target-loss + re-acquire; assert 3 distinct kills. Stress-tests the
   `combat.target==null` polling race → motivates `on target_died`.

## Group D — Shop (targets the shop body-surface gap, §4)

7. **runecraft-buy-essence-then-craft** — buy rune/pure essence from a shop, walk
   to altar, craft runes; assert rune count rose. **Depends on `shop.buy`** →
   gap-finder for §3/§4 shop.
8. **kill-loot-sell-banktrip** — kill men, loot coins+drops, sell loot to a shop,
   bank coins; assert bank coin balance rose. Chains loot + **`shop.sell`** (gap)
   + bank.
15. **shop-restock-wait-buy** — buy item until stock hits 0, wait for restock, buy
   again; needs `shop.stock` + `on shop_stock_update`. Gap-finder.

## Group E — Ground-item / loot selection (targets the ground_items.nearest gap)

9. **loot-most-valuable-nearby** — spawn several different ground items, pick up
   the highest-value one. **(FAILS today → surfaces the missing
   `ground_items.nearest` / value-sort.)**

## Group F — Inventory / equipment convenience (targets find_any + per-slot equip)

13. **multi-id-food-eat-loop** — hold mixed food ids, eat whichever is present
   using one `find_any`/loop; surfaces the missing
   `inventory.find_all`/`find_any`.
14. **is-bow-equipped-gate** — equip a shortbow, then assert via an equipment read
   that a bow is worn before ranged-attacking; surfaces the missing per-slot
   `self.equipped` accessor.

## Group G — Bank circuit / deposit-all (targets bank.deposit_all wiring + dialog open)

12. **deposit-all-keep-tools** (negative gap test) — call a single
   deposit-all-but-keep verb. **(FAILS today — unwired → forces the
   `bank.deposit_all` wiring win.)**
17. **cook-bank-circuit-with-burn-handling** — withdraw 30 raw, cook on range,
   deposit cooked + burnt separately, repeat 2 cycles; assert cooked-count in
   bank. Multi-cycle banker loop (committed set has single-pass cook only).
21. **bank-dialog-multi-answer-open** — open the bank via the full `talk_to` +
   multi-option dialog chain (not the shortcut); assert `bank.is_open`. Targets
   the backlog note "open_bank does NOT navigate the multi() menu".

## Group H — Skilling combos / reactor model (uncovered themes)

6. **smelt-then-superheat-combo** — mine/withdraw ore, smelt bars at the furnace,
   then `cast_on_item("Superheat")` for the remainder; assert smithing + magic
   xp both rose. Combo loop uncovered.
19. **event-driven-mine-on-respawn** — use `on coords_changed` / poll the rock
   object-id to mine only on respawn (AutoRune `simpleminer.txt` pattern);
   assert ore-count rose without fixed sleeps. Validates the reactor model for
   skilling.

## Group I — Prayer / magic timing reads (confirms recently-built accessors)

16. **prayer-flick-strength-during-kill** — toggle Burst of Strength on for one
   tick per attack during a kill; assert prayer points drained + activation
   observed. Confirms `prayer.active(N)` read (opcode 206) under a flick pattern.

## Group J — Chained menu fragility (targets §10 menu robustness)

22. **anvil-chained-product-menu** — smith a specific product through the
   two-level anvil menu (category→item) via `answer(find_option(...))`; assert
   the right item was made. Stress-tests the §10 chained-menu fragility.

---

## Gap → scenario cross-reference

| Gap (build-backlog) | Gap-finding scenarios |
|---|---|
| **fatigue→sleep** (NEW, §4) | 1, 2, 18 |
| **combat/health perception** (§1 opcode-234/104; §3) | 5, 10, 11 |
| **shop body surface** (§4) | 7, 8, 15 |
| **`bank.deposit_all` wiring** (§4) | 12 |
| **`ground_items.nearest`/value-sort** (§3) | 9 |
| **`inventory.find_any`** (§3) | 13 |
| **per-slot `self.equipped`** (§3) | 14 |
| **`on npc_killed`/`target_died`** (§5) | 11 |
| **`on shop_stock_update`** (§5) | 15 |
| bank-open multi-dialog navigation (§4 note) | 21 |
| chained-menu robustness (§10) | 22 |
| two-host orchestration (open question) | 3, 4, 20 |
| reactor model validation | 19 |
| recently-built `prayer.active` (206) | 16 |
