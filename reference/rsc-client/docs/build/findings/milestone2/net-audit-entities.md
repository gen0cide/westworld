> _Verbatim report from the deob-compile-and-net-depth workflow (2026-06-01)._

## Per-Opcode Fidelity Audit — INCOMING dispatch, "Player & NPC Synchronization" group

**Scope:** server→client entity opcodes decoded in `handlePacket` (obf `client.a(III)V` / `b(int,byte,int)`), deob `Mudclient.handlePacket` (Mudclient.java:4105). Compared against the clean Vineflower decompile (`decompiled/normalized-clean/client.java`, method `b(int,byte,int)` @L14213), with opcode meanings cross-referenced to OpenRSC `Payload235Generator`/`GameStateUpdater`.

**Structural confirmations (read before the per-opcode table):**
- The skeleton's `handleSceneUpdates` (`b(boolean,int)` @clean L4696) is NOT an incoming entity-stream reader — it is the OUTGOING right-click menu-action dispatcher (sends opcodes 249/53/247/180/161/14/127/99/115/136…). It parses no incoming packet. The deob documents this (Mudclient.java:4047-4050). Correct.
- The skeleton's `applyAppearanceUpdate` (`a(boolean,boolean)` @clean L2612) reads neither `mg` (incoming) nor sends anything — it is the social-entry DIALOG renderer. Not an entity-packet handler. Correct (Mudclient.java:4055-4057).
- Therefore ALL incoming player/NPC sync lives in `handlePacket`, which is fully audited below.

### Player & NPC opcodes

**191 SEND_PLAYER_COORDS** (Mudclient.java:4110-4182 vs clean L14220-14367) — local-player reposition + nearby-player movement stream. Double-buffer `Zg[]←rg[]` (count `If=Yc`); `mg.i(-2231)` align; `Lf=f(-106,11)`, `sh=f(-106,13)`, localAnim`=f(-82,4)`; `loadRegion(sh,Lf,false)`; subtract `Qg`/`zg`; worldX/Y `*Ug+64`; `wi=addNpc(worldY,Zc,worldX,-56,anim)` (player-create via obf `d`); `f(-69,8)` other-count; per-player 1-bit hasUpdate, 1-bit moved (turn: `f(-69,2)` dir, dir==3⇒skip, `D=(dir<<2)+f(-98,2)`; walk: `f(-87,3)` dir, 8-way waypoint ring %10); new-player loop `while k(-31874)+24 < len*8`: `f(-120,11)` idx, signed `f(-96,5)` dx, `f(-90,5)` dy, `f(-97,4)` anim, `addNpc(...,-112,...)`; `mg.j(25505)`. Every read width/order, sign-extension (`>15 ⇒ -=32`), branch and field write matches. **MATCH.**

**234 SEND_UPDATE_PLAYERS** (Mudclient.java:4277-4377 vs clean L14694-14898) — per-player update stream from the `We` cache. `count=f(255)`; per entry `idx=f(255)`, `player=We[idx]`, `type=h(20869)`. Sub-type cascade verified against OpenRSC `GameStateUpdater` type tags: 0=bubble-item (`f(255)`), 1=chat (`a()` icon + `ia.a` msg, ignore-hash filtered, read-skipped when null), 2=combat (`a,a,a`=damage/curHits/maxHits; local⇒`oh[3]/cg[3]`), 3=projectile (`f,f`; `.h=tgt,.z=-1`), 4=projectile-alt (`f,f`; `.h=-1,.z=tgt`), 5=full appearance (`f` echo, `c` name, `c` fname, `a` n, n×`a` equip, fill-to-12, `a`×6 colours/level/icon; null-skip reads `f+c+c+a` then `w+=6+n`), 6=self-speech, 7=no-op. The 234-uses-`We` / 104-uses-`te` role-swap is correctly applied. Read widths, the `h(20869)` type tag (vs 104's `a()` type tag), and the null-skip byte arithmetic on type 5 all match. **DIVERGENCE on sub-type 6 only** (see below); all other sub-types **MATCH.**

**104 SEND_UPDATE_NPC** (Mudclient.java:4505-4535 vs clean L15169-15215) — per-NPC update stream from the `te` cache. `count=f(255)`; `idx=f(255)`, `npc=te[idx]`, `type=a((byte)104)` (1-byte, NOT `h`). type 1: `speakerIdx=f(255)` read OUTSIDE the null guard, `ia.a` message read INSIDE — matches clean exactly; type 2: `a,a,a`=damage/curHits/maxHits with field order `u,G,d,B`. Caches/widths/order all match. **MATCH.**

**79 SEND_NPC_COORDS** (Mudclient.java:4446-4502 vs clean L15030-15166) — nearby-NPC movement stream. `qj=de; de=0`; `Ff[]←Tb[]`; `mg.i(-2231)`; inView`=f(-87,8)`; per-NPC 1-bit hasUpdate, 1-bit moved (note: for NPCs, moved==0⇒walk `f(-114,3)`, moved==1⇒turn `f(-109,2)` dir/dir==3⇒skip/`D=f(-127,2)+(dir<<2)` — the walk/turn bit polarity is opposite to opcode 191, and the deob preserves this correctly); new-NPC loop `while k(-31874)+34 < len*8`: `f(-104,12)` idx (12-bit), signed `f(-68,5)` dx, `f(-111,5)` dy, `f(-74,4)` anim, `f(-108,10)` type (10-bit), clamp `>=la.d ⇒ 24`, `addPlayer(...,127,...)` (npc-create via obf `a`); `mg.j(25505)`. All widths/order/budget(+34)/clamp match. **MATCH.**

**83 SEND_DEATH** (Mudclient.java:4581-4584 vs clean L15322-15324) — `rk=250` death-anim timer. **MATCH.**

**Entity helper methods (called by the above):**
- **getPlayer** `ta b(int,byte)` (Mudclient.java:6583 vs clean L3832): sentinel≠-123⇒`Bf=-116`; loop `de` over `Tb[].b`. **MATCH.**
- **addPlayer** (npc-create) `ta a(int,int,int,byte,int,int)` (Mudclient.java:6605 vs clean L13654): alloc `te[idx]`; known-check over `qj`/`Ff[].b`; sentinel≠127⇒sound `a(-81,-15,null)`; known⇒`D,t`+waypoint advance; new⇒full init `b,o,e,k[0]=i,D=y,x,t,F[0]=K`; `Tb[de++]`. **MATCH.**
- **addNpc** (player-create) `ta d(int,int,int,int,int)` (Mudclient.java:6658 vs clean L13871): alloc `We[idx]`; known-check over `If`/`Zg[].b`; NO sound call (correctly asymmetric vs addPlayer); known⇒`D`+waypoint; new⇒full init; dead `-98%((0-junk)/39)`; `rg[Yc++]`. **MATCH.**
- **getNpc** `ta d(int,int)` (Mudclient.java:6706 vs clean L12247): loop `Yc` over `rg[].b`; else sentinel≠220⇒`wi=null`. **MATCH.**

### Adjacent stat/inventory/scene opcodes decoded in the same dispatch (entity-state-adjacent)

**53 SEND_INVENTORY** (4256-4268 vs clean L14666): `lc=a()`; per item `raw=f(255)`, `vf=raw&0x7FFF`, `Aj=raw>>15`, stackable test `fa.e[raw&32767]==0`⇒`xe=c(103)` else `1`. The "==0 ⇒ stackable ⇒ read count" sense matches clean (the previously-inverted OLD-part bug is fixed). **MATCH.**
**156 SEND_STATS** (4566-4571 vs clean L15251): 18×`a()` curr, 18×`a()` base, 18×`b(-129)` xp (signed 16-bit), 1×`a()` quest-points. **MATCH.**
**153 SEND_EQUIPMENT_STATS** (4575-4577 vs clean L15308): 5×`a()` into `Fc`. **MATCH.**
**25 SEND_WORLD_INFO** (4554-4562 vs clean L15240): `Zc,Ki,sk,bc,rc = f(255)` ×5, `sk -= bc*rc`. Sets local player server index. **MATCH.**
**59 SEND_APPEARANCE_SCREEN** (4658-4660 vs clean L? `var1==59`): `Kg=true`. **MATCH.**
**99 SEND_BOUNDARY (walls)** (4185-4246 vs clean L14370): writes `Zf/Ni/Gj/Le`. Note OpenRSC's enum labels 99=GROUND_ITEM, but in THIS client 99=walls/boundary (the deob follows the jar, which is ground truth). **MATCH (vs jar).**
**91 SEND_GROUND_ITEM** (4380-4442 vs clean L14901): writes `Jd/yk/Hj/Ng`+`rd[]`. OpenRSC labels 91=BOUNDARY; this client uses 91=ground-items. Deob follows jar. **MATCH (vs jar).**
**211 SEND_REMOVE_WORLD_ENTITY** (4590-4655 vs clean L15327): `(len-1)/4` anchors; re-culls walls, scenery, ground-items; anchor reads use `mg.a(false)` (short) not `h(20869)`; ground-items removed when `(rx==0&&ry==0)`. Clean nests the three loops; deob writes three sequential `kept=0` loops — semantically identical. **MATCH.**

### DIVERGENCE detail

**Opcode 234, sub-type 6 (self-speech) — unconditional message read before the null guard.**

DEOB (Mudclient.java:4333-4341):
```java
} else if (type == 6) {                    // self speech (local player only)
    String message = ia.a(this.mg, false); // <-- consumed UNCONDITIONALLY
    if (player != null) {
        player.n = message; player.I = 150;
        if (this.wi == player) { this.showServerMessage(false, player.c, 0, player.n, 3, 0, player.C, null); }
    }
}
```
JAR/CLEAN (client.java L14833-14842):
```java
if (~var117 != -7 || var103 == null) { break label1692; } // type!=6 OR null -> read NOTHING
String var139 = ia.a(this.mg, false);                     // reached only if type==6 AND non-null
var103.n = var139; var103.I = 150;
if (this.wi == var103) { this.a(false, var103.c, 0, var103.n, 3, 0, var103.C, null); }
```
Why it matters: when a type-6 update is sent for an index whose `We[idx]` cache slot is null, the jar consumes **zero** bytes for that entry, but the deob consumes one variable-length scrambled string (`ia.a`). That advances the `mg` byte cursor past data the jar leaves in place, **desyncing every remaining player-update entry in the same 234 packet** (the `count` loop). This is a buffer-read-order break from byte-faithful ground truth. Real-world trigger probability is low (type 6 is the local player's own chat echo, whose cache slot is normally populated), so it rarely fires in practice — but it is a genuine semantic divergence, not cosmetic. Severity: **LOW-MEDIUM** (narrow trigger; catastrophic-per-packet desync if it does fire). All other 234 sub-types (0–5, 7) and opcode-104 type 1 correctly mirror the jar's null-skip behavior.

**Note (out-of-group, flagged for completeness):** Opcode **48 SEND_SCENERY_HANDLER** is present in the jar/clean base (`~var1==-49`, client.java L14510-14659: builds scenery `kh[]` models, writes `Se/ye/vc/bg/hg[]`) but its dispatch branch is **entirely absent from the deob `handlePacket`** — the deob jumps from opcode 99 (`return` @4247) straight to opcode 111 (@4250). The deob's own header (Mudclient.java:4063) documents that 48 should map to SCENERY, but the handler body was not transcribed. Scenery is a scene/object opcode (not a player/NPC entity), so it is outside this group's strict scope; recording it here because it sits inside the audited dispatch and is a missing-handler (not just a divergence). Severity (for the scene group): **HIGH** — without opcode 48 the client never adds/removes server-driven scenery models; flagging for the scenery/objects auditor.

## NUMBERS

- **Opcodes audited in this group (player/NPC entity sync):** 8 core — 191, 234, 104, 79, 83 + 4 entity helpers (getPlayer/addPlayer/addNpc/getNpc), plus 7 entity-adjacent decoded in the same dispatch (53, 156, 153, 25, 59, 99, 91, 211).
- **Core entity opcodes:** 191 MATCH · 79 MATCH · 104 MATCH · 83 MATCH · 234 = MATCH except sub-type 6 DIVERGENCE.
- **Entity helper methods:** getPlayer / addPlayer / addNpc / getNpc — all 4 MATCH (incl. the player-create/npc-create obf-name swap, the addPlayer-only "appeared" sound, and the dead junk-guard expression).
- **Entity-adjacent opcodes:** 53, 156, 153, 25, 59, 99, 91, 211 — all 8 MATCH (vs jar). Note 91/99 are swapped relative to OpenRSC's enum labels; the deob correctly follows the jar.
- **Tally:** MATCH = 19 · MINOR = 0 · DIVERGENCE = 1 (in-group) + 1 missing-handler (out-of-group/scenery).
  - **Verdicts MATCH (19):** 191, 79, 104, 83, 53, 156, 153, 25, 59, 99, 91, 211, getPlayer, addPlayer, addNpc, getNpc, and 234 sub-types {0,1,2,3,4,5,7}, plus the two structural non-handlers (handleSceneUpdates, applyAppearanceUpdate) confirmed to carry no incoming entity decode.
  - **DIVERGENCE list:**
    - **Opcode 234, sub-type 6 (self-speech) — buffer-read-order divergence — Severity LOW-MEDIUM.** Deob reads `ia.a(mg)` (a var-length scrambled string) before the `player != null` check; jar reads nothing when the cache slot is null, so a type-6 update for an out-of-view player desyncs the rest of the 234 packet. Mudclient.java:4334 vs clean client.java:14833-14837. Narrow trigger (local-player echo, normally non-null), but violates byte-faithfulness.
    - **Opcode 48 SEND_SCENERY_HANDLER — missing handler — Severity HIGH (out of this group's scope; scenery/objects).** Present in jar/clean (client.java L14510-14659) but no branch in deob `handlePacket` (jumps 99→111 at Mudclient.java:4247-4250). Flagged for the scene/scenery auditor.
