# `client` (Mudclient) — structural index

Obfuscated class **`client`** (deob **`Mudclient`**, package `client`), `extends e` (`GameShell`).
~55.6k normalized lines, originally **123 declared methods** (120 game methods + ctor + 2 `z`
string-pool decoders), **484 fields**. As of the delegate refactor (two waves), **43 of those
methods now live in seven sibling classes** — wave 1: `ClientPackets` ×8, `ClientSound` ×2;
wave 2: `TradeDuelBankPackets` ×4, `WidgetRenderer` ×8, `GameInterface` ×12, `MenuController` ×6,
`IncomingPackets` ×4 (see the section below). `Mudclient.java` shrank from 13168 → 8977 lines;
each subtree compiles (JDK17, 0 errors, 78 `.class`) and logs in + renders the live baseline
frame. Names below cross-reference `docs/NAMING.md` for all other classes.

Sources: `decompiled/normalized/client.java` (primary), `decompiled/cfr/client.java` (used where
Vineflower emits raw bytecode — the packet-in dispatch, login, and many render methods), the
decoded XOR string pool `il[0..659]`, and the OpenRSC `Payload235{Parser,Generator}` opcode maps.

---

## Extracted delegate classes (refactor — pure code motion)

Some standalone subsystems have been **moved out of `Mudclient.java` into sibling
package-private classes** in `package client`, using the **extract-delegate** idiom: each
new class holds a `final Mudclient m;` back-reference set in its constructor, the moved
method bodies are byte-for-byte identical (only `this.x` / unqualified `x` references were
rewritten to `m.x`), and `Mudclient` keeps a `final` delegate field through which every
former call site now routes. No logic, constants, operators, or control flow changed (one
documented behaviour-preserving exception, noted below). See
`docs/build/REFACTOR_RESULTS.md` for the full verification.

| delegate field on `Mudclient` | class (new file) | methods moved | former group |
|---|---|---|---|
| `packets` (`final ClientPackets packets = new ClientPackets(this)`) | `client.ClientPackets` (`src/client/ClientPackets.java`) | `sendOpcodeString`, `sendCommand`, `sendPrivateMessage`, `sendRemoveFriend`, `sendAddFriend`, `sendAddIgnore`, `sendRemoveIgnore`, `sendPrivacySettings` | packetout |
| `sound` (`final ClientSound sound = new ClientSound(this)`) | `client.ClientSound` (`src/client/ClientSound.java`) | `playSound`, `initSounds` | ui (audio) |
| `tradePackets` (`final TradeDuelBankPackets tradePackets = new TradeDuelBankPackets(this)`) | `client.TradeDuelBankPackets` (`src/client/TradeDuelBankPackets.java`) | `bankSend`, `sendDuelOffer`, `sendTradeOffer`, `sendDuelItems` | packetout |
| `widgetRenderer` (`final WidgetRenderer widgetRenderer = new WidgetRenderer(this)`) | `client.WidgetRenderer` (`src/client/WidgetRenderer.java`) | `drawBox`, `clearScreen`, `drawSprite`, `drawIcon`, `drawScrollList`, `drawMenuOptions`, `drawScrollbar`, `drawScrollbar2` | ui (render) |
| `gameInterface` (`final GameInterface gameInterface = new GameInterface(this)`) | `client.GameInterface` (`src/client/GameInterface.java`) | `drawWildernessWarning`, `drawShop`, `drawBank`, `drawTrade`, `drawTradeConfirm`, `drawTradeConfirmWindow`, `drawDuelConfirm`, `drawDuel`, `drawWelcome`, `drawHelpMenu`, `drawCloseButton`, `drawGameSettings` | ui (interfaces) |
| `menus` (`final MenuController menus = new MenuController(this)`) | `client.MenuController` (`src/client/MenuController.java`) | `handleGameClick`, `buildClickMenu`, `handleInventoryClick`, `menuHitTest`, `pointInRect` (private), `pointInPanel` | ui (menu/hit-test) |
| `incoming` (`final IncomingPackets incoming = new IncomingPackets(this)`) | `client.IncomingPackets` (`src/client/IncomingPackets.java`) | `handlePacket` (master opcode dispatch), `handleSceneUpdates`, `onFriendUpdate`, `applyAppearanceUpdate` | packetin |

Call sites in `Mudclient` now read `packets.sendAddFriend(...)`, `sound.playSound(...)`,
`incoming.handlePacket(...)`, `menus.menuHitTest(...)`, `gameInterface.drawBank(...)`, etc.
For wave-2 methods, `Mudclient` additionally keeps thin private forwarder stubs with the
**identical original signature** (e.g. `void drawBank(int p){ gameInterface.drawBank(p); }`) so
the existing dispatchers (`drawActiveInterface`, `drawGameFrame`) and all other call sites stay
byte-identical and re-route through the delegate.

Fields/methods widened from `private` (wave 1) so the delegates can reach them (all stay
package-private, same package): `Jh`, `Uh`, `hk`, `ni`, `ne`, `wi`, `username`,
`membersServer`, the `friendList*`/`ignoreList*` parallel arrays, plus the methods
`showServerMessage` and the static `findStringInData`. Wave 2 widened a further compiler-driven
set (~100+ fields plus shared methods such as `loadRegion`, `getPlayer`/`addPlayer`/`addNpc`,
`buildEntityModel`, `requestLogout`, `resetPanels`, `formatNumber`) — all visibility-only flips.

**Round-2 verification (de-god, 2026-06-02): functionalDivergence = false.** All 33 wave-2
extracted bodies are byte-for-byte identical to HEAD after stripping only the sanctioned
`m.`/`this.`/`Mudclient.` qualifier rewrites; all 18 kept-in-`Mudclient` bodies and the
constructor are byte-identical; all 24 pure-delegate wrappers forward args 1:1 and preserve the
HEAD signature. The wave-2 delegate ctors are side-effect-free (`this.m = m;` only). No
behaviour-preserving deviation was needed in wave 2 (unlike the wave-1 `host = this;` →
`host = m;` note below). Highest-risk targets confirmed exact: `handlePacket` opcode dispatch
(19038 normalized chars), the trade/duel/bank packet writers' byte order, the bank
withdraw/deposit send inside `drawBank`'s click block, and the menu hit-test arithmetic.

**One behaviour-preserving deviation (wave 1 only)** (in `ClientSound.initSounds`): the
AudioChannel host fallback `host = this;` became `host = m;`. In the original, `this` was the
`Mudclient` (an `Applet`, hence a `Component`); in the delegate `this` is the non-`Component`
`ClientSound`, so `m` keeps the same runtime host object passed to the `(Component)` cast.
Identical control flow, constants, argument order, and cast otherwise.

## How methods were identified

* The exception wrappers `throw i.a(e, il[N])` embed each method's **original obfuscated label**
  (e.g. `il[183]`=`"client.init()"`, `il[202]`=`"client.U("`). These are recovered per method below
  as `obf-label`. They are *not* descriptive (still single/double-letter), but anchor the mapping.
* Behaviour was read from the descriptive strings each method references (`il[]` pool) plus the
  fields/helpers it touches (`mg`=BitBuffer in, `Jh`=ClientStream out, `li`=SurfaceSprite draws,
  `Ek`=World, `Hh`=Scene, `kh/hg/rd`=GameModel[], `te/We/Zg`=GameCharacter[]).
* Incoming opcodes (`b(int,byte,int)`) and outgoing opcodes (`Jh.b(op,0)`) were matched to
  `Payload235Generator` (server→client) and `Payload235Parser` (client→server).

---

## Fields (484)

The field set is dominated by two machine-generated categories; the table after them lists every
**typed / semantically-meaningful** field by role.

| category | count | notes |
|----------|-------|-------|
| `static int` per-method **profiling counters** | ~119 | one bumped at each method entry (`++Qj;`, `++Ec;`, `++pd;` …). Obfuscation artifact — **not game state**. Examples: `Ic Pd Ci Xg lf hd Od Yg Hd Tj Bg Pi lj yh Yf Yj Ae gk Hg Th ej Dk Gc ic Kf Tf gj nd Qc bi hl ph wf Zi Ii pd wd Rf wc oi Gk Ec ih gf Ij qf Ei Zk mk Wb ie jh kk Bk jg Gh he Qj me aj dl tc al Gd ch Qh Lj ik rj ri Nk Xk pc Ck Tc Ie lg Ac bd Mf dh Nd Sd Mk Ye sc mj tk Sk Ch Fk Ad qi og ag Ik uh Uj qh ok cc Ri cd we Gg Nf kj yc Jc jf xc mi gg Li Oe yf Md fk Pe xf Ig` (and the inherited `e.Ab`). A few of these names are *also* read as real counters (`Ig`, `ef`) — disambiguate at use site. |
| `private int` **game-state scalars** | ~189 | cursors, screen coords, panel ids, selected-item indices, timers, counts, flags-as-int. Named individually only where load-bearing (below). |

### Opaque-predicate / special static fields

| obf | proposed name | type | role |
|-----|---------------|------|------|
| `vh` | `OPAQUE_FALSE` | `public static boolean` | **opaque predicate**, always `false`; gates dead control-flow throughout. Strip when reading. |
| `il` | `STRINGS` | `private static final String[660]` | XOR-encrypted string pool, decoded by the two `z` helpers. Holds UI text, login responses, method labels, skill/quest names, asset filenames. |
| `Fj` | `keyState` | `private static int[200]` | per-keycode held/pressed bitfield (`Fj[k]&2` = held, `&4` = pressed); shared with `e`/input. |
| `Jk` | `loginScreenBg` | `static int[]` | static scratch pixel/colour array used by login/welcome screens. |
| `ze` | `tickMarker` | `static long` | scratch timing marker set in the tick hook (`ze = -103L`). |

### Network / streams

| obf | proposed name | type | role |
|-----|---------------|------|------|
| `Jh` | `clientStream` | `da` (ClientStream) | server connection; all outgoing packets (`Jh.b(op,0)` begin, `Jh.b(21294)` flush). |
| `mg` | `incomingPacket` | `ja` (BitBuffer) | inbound packet buffer (`mg.f(...)` bit reads) consumed by the packet-in dispatcher. |
| `Dh` | `serverHost` | `String` | login/world host (built from `runescape.com` params). |
| `Uh` | `sessionBytes` | `byte[]` | session/handshake scratch bytes. |

### Scene / world / models

| obf | proposed name | type | role |
|-----|---------------|------|------|
| `Hh` | `scene` | `k` (Scene) | 3D scene/camera/render. |
| `Ek` | `world` | `lb` (World) | terrain/wall/object manager. |
| `li` | `surface` | `ba` (SurfaceSprite) | 2D software renderer / UI blitter. |
| `Xb` | `graphics` | `Graphics` | AWT graphics target. |
| `kh` | `objectModels` | `ca[1000]` | scene object GameModels. |
| `hg` | `wallModels` | `ca[1500]` | wall/boundary GameModels. |
| `rd` | `npcModelCache` | `ca[500]` | per-NPC/animation GameModels. |
| `Hh`-fed arrays `ye vc Se vi Ni Le Gj …` | model/terrain index buffers | `int[]` | object ids, coords, dirs, anim frame indices. |

### Entities (players / NPCs)

| obf | proposed name | type | role |
|-----|---------------|------|------|
| `wi` | `localPlayer` | `ta` (GameCharacter) | the local player. |
| `te` | `playersCache` | `ta[5000]` | id→player cache. |
| `Zg` | `players` | `ta[500]` | players in view (this tick). |
| `rg` | `playersLast` | `ta[500]` | players from previous tick (double-buffer). |
| `We` | `npcsCache` | `ta[4000]` | id→npc cache. |
| `Ff` | `npcs` | `ta[500]` | npcs in view (this tick). |
| `Tb` | `npcsLast` | `ta[500]` | npcs previous tick. |
| `If Yc Mg …` | view/entity counts | `int` | counts for the above arrays. |

### UI panels & chat

| obf | proposed name | type | role |
|-----|---------------|------|------|
| `zk fe Af yd Mc ge yi` | `panelLogin / panelGame / panelQuest / panelShop / panelBank / panelTrade / panelDuel` (7 `qa` Panel instances) | `qa` | interface/widget containers (Helvetica text). |
| `He zh Wf` | `chatList / friendsList / ignoreList` (3 `wb` MessageList) | `wb` | scrolling message/social lists. |
| `Kc gd ak je …` (`int[50]`/`String[50]` families) | menu-option buffers | arrays | right-click "Choose option" text, action ids, target ids, x/y. |
| `Vk Ej` | `skillNamesShort / skillNamesLong` | `String[18]` | skill labels (Attack…Hits) from `il[]`. |
| `Te` | `questNames` | `String[50]` | quest list (`il[529..598]`). |
| `Ld` | `combatStyleNames` | `String[5]` | Controlled/Accurate/Aggressive/Defensive. |
| `ei Dg Wh` | `chatColors / itemColors / fatigueColors` | `int[]` | RGB palettes. |
| `Lg cj wh Xf ec ig ve re Zj Qd Cj Xf` | text-entry / login / message buffers | `String` | username (`Xf`), password (`wh`), current input lines, pm targets. |

### Skills / stats / inventory / equipment

| obf | proposed name | type | role |
|-----|---------------|------|------|
| `Bi Me Lc Qf jj Dd Vb zj` | `skillCurrent / skillBase / skillXp …` | `int[14]` | per-skill current/base levels and xp. |
| `ti` | `experienceTable` | `int[99]` | xp-per-level lookup. |
| `oh Ak cg` | equipment-bonus stats | `int[18]` | armour/weapon aim/power etc. |
| `xe Aj vf` | `inventoryItems / inventoryCount / inventoryEquipped` | `int[35]` | inventory slots. |
| `xi kf of df zc xj th` | trade/duel/bank slot buffers | `int[8]` | offer/stake item ids & counts. |

### Misc state flags / timers

| obf | proposed name | type | role |
|-----|---------------|------|------|
| `Zd` | `mouseClickTimes` | `long[100]` | input timing ring. |
| `Wi` | `lastActionTime` | `long` | activity timer (logout/fatigue). |
| `ni hk` | `soundChannel / soundMixer` | `sa` / `ra` | active audio voice + mixer. |
| `Zb Vh nc cl Sf eg ld wg Wg dk` | connection/mode/state machine ints | `int` | login stage (`Zb`), world index (`Vh`), screen mode, fps cap, mouse-button mode (`eg`). |
| booleans `Bd Kh Yk Hk vk uk Ph Td Hc Kg Vi md ki Yi Je dd ne Pj Wk Qk Oh mh ff fd se vd lh zf cf Vc Pg Ue Mi Xh hj Xj Dc Fe Cd Ub ke …` | feature/dialog/render flags | `boolean` | members flag, sleeping, dead, panel-open toggles, "loaded" guards, privacy toggles. |

---

## Methods (123) grouped

`L` = first line in `decompiled/normalized/client.java`. `obf-label` = original obfuscated method
name recovered from the `i.a(e, il[N])` catch wrapper (`?` = wrapper stripped by decompiler).

### bootstrap

| L | obf signature | obf-label | proposed name | purpose |
|---|---------------|-----------|---------------|---------|
| 36183 | `static void main(String[])` | `?` | `main` | applet/standalone entry; reads `live/classic/members/veterans` mode args, builds frame, starts client. |
| 29418 | `void init()` | `client.init()` | `init` | applet init: reads `nodeid/modewhat/modewhere` params, sizes window, kicks off load. |
| 55231 | `client()` | `?` | `<init>` | constructor: allocates all state arrays (skills/quests/entities/models) and defaults. |
| 21741 | `void e(int)` | `client.MA(` | `startGame` | GameShell hook: start a game session (mouse/audio init), called from `e.run`. |
| 34505 | `void a(byte)` | `?` | `loadGameConfig` | resolve world host (`servertype/referid/*.runescape.com`), "Starting game...", build sockets. |
| 39059 | `void e(boolean)` | `?` | `loadModels` | load 3D model defs (`../content/src/models/*.ob2/.ob3`, "Loading 3d models"). |
| 38265 | `void m(byte)` | `?` | `loadMedia2d` | load 2D sprite archives (`inv1/inv2/bubble/buttons/…dat`, "2d graphics"). |
| 39375 | `void c(boolean,int)` | `?` | `loadEntitySprites` | load people/monster frames (`a.dat/f.dat`, "people and monsters", " frames of animation"). |
| 17447 | `void m(int)` | `client.ED(` | `loadMaps` | load landscape/map archives (`landscape/map/members landscape/members map`). |
| 3368 | `void j(byte)` | `?` | `loadTextures` | load `Textures` / `index.dat` texture archive into Scene. |
| 26079 | `void b(boolean)` | `client.JD(` | `drawLoadError` | fatal "Error - out of memory!/unable to load game!" help screen. |

### login

| L | obf signature | obf-label | proposed name | purpose |
|---|---------------|-----------|---------------|---------|
| 24803 | `void a(int,String,String,boolean)` | `?` | `loginUser` | open `da` stream, send opcode 0 (LOGIN), build ISAAC seed (4 randoms) + RSA session block (`ja.K`), send user/pass; handle reconnect. |
| 35141 | `void a(boolean,String,int,String,int,int,String,String)` | `?` | `registerAccount` | send opcode 2 (REGISTER_ACCOUNT); account-creation flow. |
| 48939 | `void p(int)` | `client.B(` | `drawLoginScreen` | render login UI (Username/Password fields, members/veteran gating text). |
| 36508 | `void x(int)` | `?` | `drawLoginInput` | login text-entry handling ("Please enter your username and password"). |
| 35117 | `void u(int)` | `client.CC(` | `closeConnection` | "Lost connection"; tear down stream/session. |
| 1735 | `void M(int)` | `?` | `shopRender` | (see ui — large shop panel) |

### mainloop

| L | obf signature | obf-label | proposed name | purpose |
|---|---------------|-----------|---------------|---------|
| 24459 | `void a(boolean)` | `client.SA(` | `onStopGame` | GameShell stop hook: `a(true,31)` cleanup + stop sound channel. |
| 13050 | `void a(boolean,byte)` | `?` | `tick` | per-tick update: poll input, run mob/camera updates, drive sub-systems. |
| 11918 | `void b(boolean,byte)` | `?` | `updateGameState` | high-level state-machine step (login↔game↔sleep↔dead transitions). |
| 30205 | `void c(byte)` | `?` | `sendQueuedActions` | flush queued client actions (bank/shop/tutorial: opc 22/23/84/221/236). |
| 52630 | `void K(int)` | `client.SB(` | `sendHeartbeat` | send opcode 67 (HEARTBEAT) keep-alive. |
| 3311 | `void B(int)` | `client.T(` | `requestLogout` | send opcode 102 (LOGOUT); blocks during/after combat. |
| 43288 | `void g(byte)` | `client.CB(` | `sendConfirmLogout` | confirm logout ("Sorry, you can't logout at the moment"). |
| 16163 | `void d(byte)` | `client.SD(` | `doLogout` | "Logging out..."; finalize disconnect. |
| 26226 | `void J(int)` | `?` | `drawSleepScreen` | sleep CAPTCHA loop: send opcode 45 (SLEEPWORD), handle ::logout/::lostcon. |

### packetout

| L | obf signature | obf-label | proposed name | purpose |
|---|---------------|-----------|---------------|---------|
| 11251 | `boolean a(int,int,byte,boolean,int,int,int,int,boolean)` | `?` | `walkTo` | send opcode 16 (WALK_TO_ENTITY) / 187 (WALK_TO_POINT). |
| 19133 | `boolean a(int,boolean,int,int,int,int,boolean,int,int)` | `?` | `walkToAction` | walk-then-action variant (opc 16/187 + queued interaction). |
| moved | `void b(String,int)` | `client.AA(` | `sendOpcodeString` | **→ `ClientPackets`.** generic helper: begin `opcode`, write a string, flush. |
| moved | `void a(String,int)` | `client.UC(` | `sendCommand` | **→ `ClientPackets`.** send opcode 38 (COMMAND) chat-command (`::`). |
| moved | `void a(byte,String,String)` | `client.KB(` | `sendPrivateMessage` | **→ `ClientPackets`.** send opcode 218 (SOCIAL_SEND_PRIVATE_MESSAGE). |
| moved | `void b(String,byte)` | `?` | `sendRemoveFriend` | **→ `ClientPackets`.** send opcode 167 (SOCIAL_REMOVE_FRIEND). |
| moved | `void b(int,String)` | `?` | `sendAddFriend` | **→ `ClientPackets`.** send opcode 195 (SOCIAL_ADD_FRIEND); friend-list-full checks. |
| moved | `void a(String,byte)` | `?` | `sendAddIgnore` | **→ `ClientPackets`.** send opcode 132 (SOCIAL_ADD_IGNORE); ignore-list checks. |
| moved | `void a(byte,String)` | `client.E(` | `sendRemoveIgnore` | **→ `ClientPackets`.** send opcode 241 (SOCIAL_REMOVE_IGNORE). |
| moved | `void c(int,int,int,int,int)` | `client.L(` | `sendPrivacySettings` | **→ `ClientPackets`.** send opcode 64 (PRIVACY_SETTINGS_CHANGED). |
| 31018 | `void a(int,int,int)` | `?` | `onFriendUpdate` | (packetin — friend/login state, see below) |
| 53359 | `void b(int,boolean)` | `?` | `drawGameOptions` | game-settings panel; sends opcode 111 (GAME_SETTINGS_CHANGED). |
| 18451 | `void G(int)` | `?` | `sendDialogAnswer` | send opcode 116 (QUESTION_DIALOG_ANSWER). |
| 18657 | `void F(int)` | `?` | `sendAppearance` | send opcode 235 (PLAYER_APPEARANCE_CHANGE) from char-design. |
| 24485 | `void k(byte)` | `?` | `sendCombatStyle` | send opcode 29 (COMBAT_STYLE_CHANGED). |
| 19436 | `void a(int,int,byte)` | `client.IC(` | `sendDuelOffer` | send opcode 33 (DUEL_OFFER_ITEM); "cannot be added to a duel offer". |
| 52692 | `void c(int,byte,int)` | `client.A(` | `sendTradeOffer` | send opcode 46 (PLAYER_ADDED_ITEMS_TO_TRADE_OFFER); "cannot be traded". |
| 43337 | `void a(boolean,int)` | `?` | `sendConfirmLogoutAck` | send opcode 31 (CONFIRM_LOGOUT). |
| 7792 | `void n(byte)` | `?` | `drawTradeWindow` | trade UI; sends opc 55/230 (accept/decline trade). |

### packetin

| L | obf signature | obf-label | proposed name | purpose |
|---|---------------|-----------|---------------|---------|
| 43853 | `void b(int,byte,int)` | `?` | `handlePacket` | **MASTER server→client dispatch** (≈42 opcodes): player/npc coords (191/79), updates (234/104), inventory (53/90/123), trade (15/20/92/97/128/162/253), duel (6/30/172/176/210/225/253), shop/bank (42/101/203/249/137), stats/xp (33/153/156/159), appearance (59/213), sound (204), prayers active (206), boxes/messages (131/222), system update (52), sleep (84/117/194/244), options menu (245/252), privacy (51/240). Uses `mg` bit reads. |
| 14175 | `void b(boolean,int)` | `?` | `handleSceneUpdates` | per-region mob/scene packet stream: boundaries (91), scenery (48/211), ground items (99), bubbles (36), projectiles, walking/animation deltas (opc 4/14/50/53/79/90/91/99/103/113/115/135/137/142/153/158/161/165/169/170/171/180/190/202/229/246/249). |
| 31018 | `void a(int,int,int)` | `?` | `onFriendUpdate` | apply SEND_FRIEND_UPDATE / login-state ("has logged in/out", display-name-change). |
| 6745 | `void a(boolean,boolean)` | `?` | `applyAppearanceUpdate` | apply player appearance/equip update from stream. |

### scene

| L | obf signature | obf-label | proposed name | purpose |
|---|---------------|-----------|---------------|---------|
| 31882 | `void s(int)` | `?` | `drawWorld` | main 3D world render: build Scene from World, place models, camera, project, blit to surface. |
| 36734 | `boolean a(int,int,boolean)` | `?` | `loadRegion` | (re)load terrain region around player ("Loading... Please wait", "Loc/Bound Error"). |
| 19602 | `ca a(boolean,int,int,int,int,int)` | `?` | `buildEntityModel` | build a GameModel for a scene/wall/door entity from World data. |
| 51839 | `void a(int,int,int,int,int,int,int,int)` | `?` | `addSceneObject` | register an object/scenery GameModel into the Scene at tile coords. |
| 40626 | `void b(int,int,int,int,int,int,int)` | `client.CA(` | `addWallModel` | add a wall/boundary GameModel to the Scene. |
| 19403 | `void a(int,int,int,int,int,int,int)` | `client.LB(` | `addGroundObject` | add ground/floor decoration model. |
| 10186 | `void b(int,int,int,int,int,int,int,int)` | `?` | `buildTerrainTile` | build terrain tile geometry/lighting. |
| 11219 | `ta b(int,byte)` | `client.AC(` | `getPlayer` | resolve/alloc a player `ta` by server index from cache. |
| 42171 | `ta a(int,int,int,byte,int,int)` | `client.U(` | `addPlayer` | create/update a player entity at coords with anim. |
| 42893 | `ta d(int,int,int,int,int)` | `?` | `addNpc` | create/update an NPC entity. |
| 37439 | `ta d(int,int)` | `client.K(` | `getNpc` | resolve/alloc an NPC `ta` by index. |
| 30190 | `void a(int,int,byte,int)` | `client.D(` | `setCamera` | position/orient the Scene camera. |
| 23985 | `void b(int,int,int)` | `?` | `updateCamera` | per-tick camera follow/auto-rotate. |
| 19886 | `void o(byte)` | `client.NA(` | `resetChatInput` | clear chat input buffers (`x`,`Ob`). |
| 30003 | `void v(int)` | `?` | `sortDrawList` | depth-sort the menu/sprite draw list (`ac.z[]` swap). |
| 9963 | `static int a(byte[],String,int)` | `client.ND(` | `findStringInData` | scan a data blob for a string/record offset. |
| 17005 | `static String a(int,tb,int)` | `client.CD(` | `readDefString` | read a length-prefixed def string from a `tb` ("Cabbage" default). |

### ui

| L | obf signature | obf-label | proposed name | purpose |
|---|---------------|-----------|---------------|---------|
| 1335 | `void I(int)` | `?` | `drawActiveInterface` | dispatch to the open panel renderer (calls `d/j/l/r/n/h/q/z…`). |
| 21910 | `void f(int)` | `?` | `drawHud` | in-world HUD overlays: sleep/dead/wilderness banners, fatigue, system-update countdown. |
| 21475 | `void H(int)` | `?` | `drawWildernessWarning` | "Warning! Proceed with caution" wilderness dialog. |
| 1735 | `void M(int)` | `?` | `drawShop` | shop panel: buy/sell list, prices ("Buying and selling items"); opc 166/221/236. |
| 3909 | `void r(int)` | `?` | `drawBank` | bank panel: withdraw/deposit, pages; opc 22/23/212. |
| 7792 | `void n(byte)` | `?` | `drawTrade` | trade offer window (your/their offer, accept/decline). |
| 11538 | `void a(int,byte,int)` | `?` | `drawTradeConfirm` | trade confirmation screen. |
| 42311 | `void N(int)` | `?` | `drawTradeConfirmWindow` | "Please confirm your trade" window; opc 104/230. |
| 17736 | `void h(int)` | `?` | `drawDuelConfirm` | duel confirm window; opc 77/197/230. |
| 49168 | `void q(int)` | `?` | `drawDuel` | duel setup window (stakes, options No magic/prayer/weapons/retreat); opc 8/176/197. |
| 20121 | `void z(int)` | `?` | `drawReportAbuse` | report-abuse rule-picker; sends opcode 206 (REPORT_ABUSE). |
| 33699 | `void L(int)` | `?` | `drawPlayerMenu` | right-click player context menu ("Choose a target", "Click to message"); opc 59. |
| 30986 | `void f(boolean)` | `client.CE(` | `drawOptionsTab` | "Configuration" / options tab. |
| 52885 | `void A(int)` | `client.QB(` | `drawChatHistoryTabs` | chat/quest/private history tab switcher. |
| 17040 | `void l(int)` | `?` | `drawChat` | render chat message panel/scrollback. |
| 28904 | `void j(int)` | `?` | `drawWelcome` | "Welcome to RuneScape" recovery/unread-messages box. |
| moved | `void a(int,String)` | `client.SC(` | `playSound` | **→ `ClientSound`.** play a named sound effect (`.pcm`). |
| moved | `void E(int)` | `client.LA(` | `initSounds` | **→ `ClientSound`.** "Sound effects"/"Unable to init sounds:" audio engine bring-up. |
| 13908 | `void k(int)` | `?` | `drawMinimap` | minimap/compass panel render. |
| 16182 | `void D(int)` | `?` | `drawInventoryTab` | inventory tab render. |
| 16182-adj `c(boolean,int)` @39375 see bootstrap | | | |
| 53359 | `void b(int,boolean)` | `?` | `drawGameSettings` | settings panel (privacy/mouse/sound toggles); opc 111. |
| 31871 | `void g(int)` | `client.EA(` | `setMouseButtonMode` | toggle one/two-button mouse (`eg`). |
| 42297 | `void o(int)` | `client.FC(` | `resetTradeDuelState` | clear trade/duel offer state. |
| 30205-see mainloop | | | |
| 38732 | `void w(int)` | `client.GD(` | `drawCharDesignControls` | character-design arrow buttons (mouse hover). |
| 52554 | `void t(int)` | `client.J(` | `drawCharDesign` | "Please design Your Character" appearance screen (Head/Hair/Top/…). |
| 37761 | `void a(String,byte,String)` | `client.MD(` | `addChatMessage` | append a formatted (Helvetica) message to the chat list. |
| 38156 | `void b(byte,String,String)` | `?` | `showServerMessage` | display a server/system message. |
| 43306 | `void a(String[],int,int,boolean)` | `client.RA(` | `drawMenuOptions` | render the right-click option list. |
| 3754 | `void a(int,int,String[],boolean,String)` | `?` | `drawScrollList` | generic scrollable list/menu renderer. |
| 13042 | `void a(byte,int,int,int,boolean,int)` @49154 | `client.WC(` | `drawScrollbar` | scrollbar widget. |
| 3342 | `void a(int,int,int,int,boolean,int)` | `client.BE(` | `drawScrollbar2` | secondary scrollbar/slider. |
| 41259 | `void C(int)` | `client.WB(` | `drawHelpMenu` | help/close-window overlay. |
| 41920 | `void l(byte)` | `?` | `drawCloseButton` | "Click here to close window" button. |
| 40445 | `void a(byte,int,String)` | `?` | `drawTextField` | draw an editable text field. |
| 41389 | `void h(byte)` | `?` | `drawSocialDialog` | add-friend/add-ignore/send-message entry dialogs. |
| 54737 | `void d(boolean)` | `?` | `drawReportNameEntry` | "Enter the name of the player you wish to report" + mute options. |
| 13050-see mainloop | | | |

### input

| L | obf signature | obf-label | proposed name | purpose |
|---|---------------|-----------|---------------|---------|
| 19905 | `void a(int,int,int,int)` | `?` | `handleGameClick` | route an in-world mouse click to walk/interact/menu. |
| 35711 | `void a(int,int)` | `?` | `buildClickMenu` | assemble right-click options for clicked tile/entity ("Walk here", "Examine", …). |
| 31882-see scene | | | |
| 40645 | `void a(int,boolean)` | `?` | `handleInventoryClick` | inventory item click (Use/Wear/Wield/Drop/Remove). |
| 3648 | `int b(int,int)` | `client.HB(` | `menuHitTest` | hit-test the menu/option region for a mouse point. |
| 38766 | `boolean e(int,int)` | `?` | `pointInRect` | mouse-in-rectangle test helper. |
| 38848 | `boolean a(byte,int,int)` | `?` | `pointInPanel` | mouse-in-panel test helper. |
| 36734-see scene | | | |
| 7738 | `void n(int)` | `client.PD(` | `pollInput` | per-tick input poll / cursor update. |

### util

| L | obf signature | obf-label | proposed name | purpose |
|---|---------------|-----------|---------------|---------|
| 24370 | `String c(int,int)` | `client.GC(` | `formatNumber` | format an int (e.g. coins/level) to display string. |
| 37813 | `void i(int)` | `?` | `updateTimers` | advance per-tick counters/timers. |
| 37604 | `void i(byte)` | `?` | `resetGameState` | clear transient game state on (re)entry. |
| 15931 | `void y(int)` | `client.HC(` | `drawProgressBar` | loading progress bar. |
| 15754 | `void b(int,int,int,int,int)` | `?` | `drawBox` | filled/bordered box primitive. |
| 17480 | `void q(byte)` | `?` | `clearScreen` | clear/reset the surface buffer. |
| 43113 | `void p(byte)` | `?` | `resetPanels` | tear down / re-init panel widgets. |
| 19380 | `void a(int,Runnable)` | `client.S(` | `runOnQueue` | run a Runnable (deferred/EventQueue helper). |
| 49022 | `void a(byte,int)` | `?` | `setPanelVisible` | show/hide a panel by id. |
| 13009 `a(byte,int)` mirror | | | |
| 10065 | `void a(boolean,int,int,int)` | `?` | `drawSprite` | blit a sprite with flags at x/y. |
| 7868 `a(int,int,byte,int)` @3342-family | | `drawIcon` | draw a small UI icon. |
| 26216 | `boolean f(byte)` | `client.LC(` | `isLoaded` | "ready/loaded" guard check. |
| 38732-see ui | | | |
| 9963-see scene | | | |
| 55613 | `static char[] z(String)` | `?` | `xorDecode1` | string-pool decoder stage 1 (XOR `~` if len<2). |
| 55622 | `static String z(char[])` | `?` | `xorDecode2` | string-pool decoder stage 2 (mod-5 key XOR). |

---

### Notes on duplicate `obf signature` rows

Several obfuscated names are overloaded by parameter list (`a(int,int)`, `a(int,String)`, `b(int,int)`
…). Rows are keyed by `L` (line); where a method is referenced from two groups, the canonical row is the
first occurrence and the others say "see <group>". The packet-in master dispatcher (`b(int,byte,int)`
@43853) and the scene-update reader (`b(boolean,int)` @14175) only fully decompile in the **CFR**
output; the opcode names above are taken from `Payload235Generator`.
