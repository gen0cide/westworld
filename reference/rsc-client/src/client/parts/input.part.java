// ===== input =====
// Methods: handleGameClick, buildClickMenu, handleInventoryClick,
//          menuHitTest (getInventoryCount), pointInRect (isItemEquipped),
//          pointInPanel (isEquipSlotActive), pollInput
//
// Field naming follows MUDCLIENT_SKELETON.md; class names follow NAMING.md.
// Obfuscation strips applied to every method:
//   - `boolean bl = client.vh;` opaque-predicate removed (always false)
//   - `++<counter>;` profiling counter calls removed
//   - `if (bl) return;` / `while (!bl)` dead branches removed
//   - `try { BODY } catch (RuntimeException e) { throw ErrorHandler.a(e,"sig"); }` unwrapped
//   - anti-tamper `if (param != CONST) this.someOtherCall(junkArgs);` guards removed
//   - junk bitwise masks before arithmetic (`~x != ~y` → `x != y`, etc.) simplified
//
// Field notes (named here by behaviour, cross-checked vs oracle mudclient204):
//   Kk  (int[8192])  clickHistoryX    – ring buffer of recent click X values
//   uj  (int[8192])  clickHistoryY    – ring buffer of recent click Y values
//                                       (oracle: mouseClickXHistory/mouseClickYHistory)
//   nk  (int)        clickRingHead    – ring buffer write index (masked 0x1FFF)
//   oh  (int[])      equipBonusDisplay – equipment stat array (cleared on top-bar click)
//   ai  (int)        combatTimeout    – >0 while in/after combat; logout blocked unless 0
//                                       (oracle: combatTimeout; set to 500 by idle-logout path)
//   bj  (int)        logoutTimeout    – set to 1000 after requestLogout(0) fires
//   af  (int)        selectedSpell    – selected spell index (-1 = none)
//   Bh  (int)        selectedItemInventoryIndex – selected inventory slot (-1 = none)
//   ig  (String)     selectedItemName
//   Pg  (boolean)    isMember
//   rg  (ta[500])    playersLast      – player entities from previous tick
//   zg,sh,sk,Qg,Lf,Ki (int) – camera / region offset fields (localRegionX/Y, planeW/H, regionX/Y)
//
// Cross-class game-data tables (scattered across obf classes by the obfuscator;
// names reflect oracle GameData.*, obf traces kept for accuracy — NOTE these
// CONTRADICT a naive "EntityDef/BitBuffer holds everything" reading):
//   ac.x   (String[]) itemName       – oracle GameData.itemName    (NAMING: ac=DecodeBuffer)
//   fa.e   (int[])    itemStackable  – oracle GameData.itemStackable (NAMING: fa=ClientIOException; static table reused)
//   mb.k   (int[])    itemWearable   – oracle GameData.itemWearable  (NAMING: mb=Utility)
//   qb.e   (int[])    spellType      – oracle GameData.spellType     (NAMING: qb=GameFrame; static table reused)
//   ja.L   (String[]) spellName      – oracle GameData.spellName     (NAMING: ja=BitBuffer)
//   lb.ac  (String[]) itemCommand    – oracle GameData.itemCommand   (NAMING: lb=Scene)
//   h.c    (int[])    itemPicture    – oracle GameData.itemPicture   (NAMING: h=TextEncoder)
//   ua.Bb  (int[])    itemMask       – oracle GameData.itemMask      (NAMING: ua=Surface)
//   ta.K   (int)      currentY       – GameCharacter currentY (pairs with waypointsY F[]); ta.i = currentX
//   ta.s   (int)      level          – GameCharacter combat level
//   ta.b   (int)      serverIndex    – GameCharacter serverIndex
//   ta.c   (String)   name ; ta.C    (String) displayName
//
// Key method cross-references (using skeleton proposed names):
//   requestLogout(0)  →  void B(int)  (opcode 102 LOGOUT)
//   resetChatInput(x) →  void o(byte)
//   resetPanels(x)    →  void p(byte)
//   drawSocialDialog  →  void h(byte)
//   friendsList       →  zh  (wb / MessageList, reused as shared right-click option-menu list)

    // -----------------------------------------------------------------
    // handleGameClick  obf: final void a(int,int,int,int)
    // (oracle: protected void handleMouseDown(int button, int x, int y))
    // -----------------------------------------------------------------

    /**
     * Records a world click in the click-history ring buffer and triggers an
     * auto-logout if a bot-like repeated-identical-click pattern is detected.
     *
     * Ring buffers {@code Kk}/{@code uj} hold up to 8191 (0x1FFF) recent
     * click (x, y) pairs.  After storing the new click, the method walks
     * forward over strides 10..3999, looking for the current (x,y) pair at
     * exactly that stride back in history.  If found, it verifies the whole
     * preceding run is consistent; if it is, the run is non-trivial, and no
     * combat / logout is pending, opcode 102 (LOGOUT) is sent.
     *
     * Also clears the equipment-stat display overlay ({@code oh}) when the
     * click Y is in the top-bar area (param2 &le; 87).
     *
     * <p>Verified vs clean base: the logout guard is {@code ai == 0} (combat
     * timeout idle) — the earlier reconstruction had {@code ai == -1}, which
     * was wrong (clean base: {@code -1 == ~this.ai} ⟺ {@code ai == 0}).</p>
     *
     * obf: final void a(int param1, int param2, int param3, int param4)
     */
    final void handleGameClick(int clickX, int topBarY, int unused, int clickY) {
        // Store this click in the ring buffer at the current head position.
        // obf: Kk[nk] = param1; uj[nk] = param4;  (param4, not param2, is the stored Y)
        Kk[nk] = clickX;
        uj[nk] = clickY;
        nk = (nk + 1) & 0x1FFF; // advance write head, wraps at 8191

        // Clicking in the top HUD area (y <= 87) dismisses the equipment stat overlay.
        if (topBarY <= 87) {
            oh = null;
        }

        // Scan recent ring history for a bot-like "exact same click N ticks in a row" pattern.
        // If detected with no combat/logout pending, trigger automatic logout (opcode 102).
        for (int stride = 10; stride < 4000; stride++) {
            int slotA = (nk - stride) & 0x1FFF; // ring slot exactly 'stride' clicks ago
            // obf: if (Kk[slotA] == param1 && ~uj[slotA] == ~param4)
            if (Kk[slotA] != clickX || uj[slotA] != clickY) {
                continue; // that historic slot does not carry the current (x,y)
            }
            // Found a match at distance 'stride'.  Walk the inner range [1..stride-1],
            // comparing each recent slot (slotB) against its stride-offset mirror (slotC).
            //   slotB: position (nk    - inner)  – what we clicked 'inner' ticks ago
            //   slotC: position (slotA - inner)  – that slot's partner 'stride' clicks earlier
            boolean hasMismatch = false;
            for (int inner = 1; inner < stride; inner++) {
                int slotB = (nk    - inner) & 0x1FFF; // recent entry at offset 'inner'
                int slotC = (slotA - inner) & 0x1FFF; // mirror at (nk - stride - inner)

                // If slotC doesn't match the current click → the run is non-trivial.
                // obf: if (param1 != Kk[slotC] || ~uj[slotC] != ~param4) hasMismatch = true;
                if (Kk[slotC] != clickX || uj[slotC] != clickY) {
                    hasMismatch = true;
                }
                // If slotB and slotC differ → the stride-based pattern is broken; abandon stride.
                if (Kk[slotB] != Kk[slotC] || uj[slotB] != uj[slotC]) {
                    break; // inner loop exits, outer loop increments stride
                }
                // Last inner iteration: a solid repeated-click run was detected.
                // obf: if (~(stride-1) == ~inner && hasMismatch && -1 == ~ai && 0 == bj)
                //   ~(stride-1) == ~inner  ⟺  inner == stride - 1
                //   -1 == ~ai              ⟺  ai == 0   (combat timeout idle)
                if (inner == stride - 1 && hasMismatch && ai == 0 && bj == 0) {
                    requestLogout(0); // obf: this.B(0) — opcode 102 LOGOUT, anti-bot measure
                    return;
                }
            }
        }
    }

    // -----------------------------------------------------------------
    // buildClickMenu  obf: private final void a(int,int)
    // (oracle: createRightClickMenu(), type==1 / player branch)
    // -----------------------------------------------------------------

    /**
     * Builds the right-click context menu for a player entity that was
     * scene-picked under the cursor.
     *
     * Appends menu entries to the shared right-click option list ({@code zh},
     * a reused {@code wb} MessageList) via {@code zh.a(...)} calls.  Entries
     * depend on whether a spell or inventory item is currently selected:
     *
     * <pre>
     *  selectedSpell &ge; 0 (type 1/2) → 800  Cast [spell] on [player]
     *  selectedItem  &ge; 0            → 810  Use [item] with [player]
     *  in attack range                → 805/2805  Attack
     *  else if members                → 2806  Duel with
     *  (then always)                  → 2810  Trade with
     *                                 → 2820  Follow
     *                                 → 2833  Report abuse
     * </pre>
     *
     * <p>Verified vs clean base — fixes vs prior reconstruction:</p>
     * <ul>
     *   <li>{@code levelDiff = localPlayer.level - player.level}
     *       (obf {@code -ta.s + wi.s}); was reversed.</li>
     *   <li>Attack-range test uses {@code player.currentY} (obf {@code ta.K});
     *       was {@code currentX}.  And the bound is strict {@code < 2203}
     *       (obf {@code ~(...) > -2204}); was {@code <= 2203}.</li>
     *   <li>Attack / Duel are mutually exclusive (else-if via {@code break
     *       label102}); the prior code emitted both and gated Duel only on
     *       members.</li>
     *   <li>{@code qb.e} = spell-type table, {@code ja.L} = spell-name table
     *       (oracle {@code GameData.spellType/spellName}); were mislabelled
     *       {@code InputState.mouseButtons}.</li>
     * </ul>
     *
     * obf: private final void a(int param1, int param2)
     */
    private final void buildClickMenu(int playerIndex, int dummy) {
        // Anti-tamper: if (param2 != -12) this.o(-32); — stripped.

        GameCharacter player = playersLast[playerIndex]; // obf: rg[playerIndex] (ta)
        String playerName  = player.name;               // obf: ta.c

        // Wilderness boundary flag (oracle: int i = 2203 - (localRegionY+planeHeight+regionY)).
        // obf: -zg - sh + -sk + 2203.  Negative ⇒ wilderness combat zone.
        int wildY = -zg - sh + -sk + 2203;
        // obf: if (2640 <= Qg + Lf - -Ki)  (oracle: localRegionX+planeWidth+regionX >= 2640)
        if (2640 <= Qg + Lf - (-Ki)) {
            wildY = -50; // deep wilderness / above the 2640 map boundary
        }

        // --- Level-difference colour tag (oracle: "@or1@".."@gre@") ---
        // levelDiff = localPlayer.level - player.level   (obf: -ta.s + wi.s).
        // Negative ⇒ target out-levels us (orange "@or*@"); positive ⇒ we out-level them (green "@gr*@").
        String colourTag = "";
        int levelDiff    = 0;
        // obf: if (~wi.s < -1 && -1 > ~ta.s)  ⟺  wi.s > 0 && ta.s > 0
        if (localPlayer.level > 0 && player.level > 0) { // obf: wi.s, ta.s
            levelDiff = localPlayer.level - player.level; // obf: -ta.s + wi.s
        }
        // Orange severity (target outranks us): more negative ⇒ more dangerous.
        if (levelDiff <  0)  colourTag = STRINGS[40]; // "@or1@"
        if (levelDiff < -3)  colourTag = STRINGS[39]; // "@or2@"  (obf: 2 < ~levelDiff)
        if (levelDiff < -6)  colourTag = STRINGS[49]; // "@or3@"  (obf: 5 < ~levelDiff)
        if (levelDiff < -9)  colourTag = STRINGS[10]; // "@red@"  (obf: 8 < ~levelDiff)
        // Green severity (we outrank target).
        if (levelDiff >  0)  colourTag = STRINGS[35]; // "@gr1@"  (obf: -1 > ~levelDiff)
        if (levelDiff >  3)  colourTag = STRINGS[37]; // "@gr2@"
        if (levelDiff >  6)  colourTag = STRINGS[47]; // "@gr3@"  (obf: ~levelDiff < -7)
        if (levelDiff >  9)  colourTag = STRINGS[27]; // "@gre@"  (obf: -10 > ~levelDiff)

        // Suffix string e.g. " @or2@(level-64)".  STRINGS[42] = "(level-"
        String levelSuffix = " " + colourTag + STRINGS[42] + player.level + ")";

        // ── Case 1: Spell selected — Cast [spell] on [player] (action 800) ──────
        if (af >= 0) { // af = selectedSpell
            // Only player-targeting spells (type 1 or 2) get a menu entry.
            // obf: if (1 != qb.e[af] && -3 != ~qb.e[af]) return;  ⟺  spellType not in {1,2}
            if (GameData.spellType[af] != 1 && GameData.spellType[af] != 2) {
                return; // not a player-target spell — no entry
            }
            // STRINGS[15] = "@whi@"  STRINGS[46] = "Cast "  STRINGS[50] = " on"
            zh.a(player.serverIndex,                                  // obf: ta.b
                    STRINGS[15] + playerName + levelSuffix,           // menuText2
                    800,                                              // CAST_SPELL_ON_PLAYER
                    STRINGS[46] + GameData.spellName[af] + STRINGS[50], // menuText1
                    af, 3296);
            return;
        }

        // ── Case 2: Inventory item selected — Use [item] with [player] (810) ───
        if (Bh >= 0) { // Bh = selectedItemInventoryIndex
            // STRINGS[38] = "Use "  STRINGS[53] = " with"
            zh.a(player.serverIndex,
                    STRINGS[15] + playerName + levelSuffix,
                    810,                                              // USE_ITEM_WITH_PLAYER
                    STRINGS[38] + ig + STRINGS[53],                   // "Use [selectedItemName] with"
                    Bh, 3296);
            return;
        }

        // ── Case 3: No selection — standard player menu ────────────────────────
        // Attack and Duel are MUTUALLY EXCLUSIVE (obf: `break label102` after Attack).
        attackOrDuel: {
            // Attack range test (oracle: i > 0 && (player.currentY-64)/magicLoc + planeHeight + regionY < 2203).
            // obf: if (0 < wildY && ~((-64 + ta.K)/Ug - (-sk + -zg)) > -2204)
            //   ~X > -2204  ⟺  X < 2203 ;   -(-sk + -zg) = sk + zg = planeHeight + regionY
            //   ta.K = currentY ;  Ug = magicLoc (tile scale)
            boolean inRange = wildY > 0
                    && ((-64 + player.currentY) / Ug - (-sk + -zg)) < 2203;
            if (inRange) {
                // 805 in true PvP wilderness, 2805 in a safe/duel zone.
                // obf: levelDiff >= 0 && -6 < ~levelDiff ? 805 : 2805  ⟺  (0<=levelDiff<5) ? 805 : 2805
                int attackActionId = (levelDiff >= 0 && levelDiff < 5) ? 805 : 2805;
                zh.a(player.serverIndex, attackActionId, false,
                        STRINGS[48],                                  // "Attack"
                        STRINGS[15] + playerName + levelSuffix);
                break attackOrDuel; // Attack was added → skip Duel (else-if)
            }
            // Duel with (members server only).  STRINGS[118] = "Duel with"
            if (isMember) { // obf: Pg
                zh.a(player.serverIndex, 2806, false,
                        STRINGS[118],
                        STRINGS[15] + playerName + levelSuffix);
            }
        }

        // Trade with (2810) and Follow (2820) — always present when no selection.
        // STRINGS[116] = "Trade with"  STRINGS[119] = "Follow"
        zh.a(player.serverIndex, 2810, false,
                STRINGS[116],
                STRINGS[15] + playerName + levelSuffix);
        zh.a(player.serverIndex, 2820, false,
                STRINGS[119],
                STRINGS[15] + playerName + levelSuffix);

        // Report-abuse link (action 2833).  Passes name + display name for the
        // abuse-report dialog.  STRINGS[120] = "Report abuse"
        // zh.a(menuText1, menuText2, name, actionId, displayName, flags)
        zh.a(STRINGS[120],
                STRINGS[15] + playerName + levelSuffix,
                player.name,        // obf: ta.c – lookup key for report-abuse dialog
                2833,
                player.displayName, // obf: ta.C – display name shown in report form
                (byte)103);
    }

    // -----------------------------------------------------------------
    // handleInventoryClick  obf: private final void a(int,boolean)
    // (oracle: private void drawUiTabInventory(boolean nomenus))
    // -----------------------------------------------------------------

    /**
     * Renders the inventory panel and, when {@code buildMenu} is {@code true},
     * processes mouse-hover over inventory slots to populate the right-click
     * option list.
     *
     * <p>Layout: 5 columns × {@code inventoryMaxSlots/5} rows of 49×34 px
     * slots anchored at x = {@code surface.width − 248}.</p>
     *
     * <p>Equipped slots ({@code inventoryEquipped[slot] == 1}) draw a red
     * (0xFF0000) box; other slots a 181-grey box (both alpha 128).</p>
     *
     * Menu action IDs (match oracle mudclient204):
     * <pre>
     *   600  Cast [spell] on item   (spell type 3)
     *   610  Use [selected] with    (item-on-item)
     *   620  Remove                 (worn item)
     *   630  Wear / Wield
     *   640  Item custom command    (e.g. Eat, Read)
     *   650  Use
     *   660  Drop
     *   3600 Examine
     * </pre>
     *
     * <p>Verified vs clean base — the menu structure was substantially wrong in
     * the prior reconstruction.  Correct control flow (matching oracle):</p>
     * <ul>
     *   <li>spell selected → only a 600 entry (if spellType==3), then return;</li>
     *   <li>{@code Bh < 0} (NO item selected, obf {@code -1 < ~Bh}) → the full
     *       Remove/Wear/Use/Drop/Examine menu;</li>
     *   <li>{@code Bh >= 0} (item selected) → fall through to a single 610
     *       "Use [selected] with" entry.</li>
     * </ul>
     * The prior code inverted the {@code Bh} test, invented an "equip-tab" case,
     * and emitted Wear/Use/Drop/Examine in both branches.
     *
     * obf: private final void a(int param1, boolean param2)
     */
    private final void handleInventoryClick(int dummy, boolean buildMenu) {
        // Anti-tamper: if (param1 != -15252) this.b(-79,(byte)75,-83); — stripped.

        // Inventory panel left edge: surfaceWidth − 248.  obf: -248 + li.u
        int invX = surface.width - 248;
        // Draw the inventory panel background sprite (oracle: drawSprite(uiX, 3, spriteMedia+1)).
        // obf: li.b(-1, tg+1, 3, invX) — args (guard=-1, spriteIndex=tg+1, y=3, x=invX); tg = spriteMediaBase
        surface.drawSprite(invX, 3, tg + 1);

        // ── Render each slot ─────────────────────────────────────────────────
        for (int slot = 0; slot < cl; slot++) { // cl = inventoryMaxSlots (e.g. 30)
            int slotX = invX + (slot % 5) * 49;
            int slotY = slot / 5 * 34 + 36;

            // obf: if (lc > slot && -2 == ~Aj[slot])  ⟺  slot < lc && Aj[slot] == 1 (equipped)
            if (slot < lc && Aj[slot] == 1) { // lc = inventoryItemsCount; Aj = inventoryEquipped
                // Equipped item slot: red box (alpha 128).
                // obf: li.c(128, slotX, 34, 0, slotY, 49, 0xFF0000)
                surface.drawBoxAlpha(128, slotX, 34, 0, slotY, 49, 0xFF0000);
            } else {
                // Normal slot: 181-grey box (alpha 128).
                // o.a(181, junk, 181, 181) → grey ARGB (oracle: Surface.rgb2long(181,181,181)).
                surface.drawBoxAlpha(128, slotX, 34, 0, slotY, 49,
                        ISAAC.a(181, dummy ^ -7922, 181, 181));
            }

            if (slot < lc) { // lc = inventoryItemsCount
                int itemId = vf[slot]; // vf = inventoryItemId[]
                // Draw item sprite (oracle: spriteClipping(x,y,48,32,spriteItem+itemPicture,itemMask,0,0,false)).
                // obf: li.a(slotY, h.c[itemId], 0, false, 0, sg + ua.Bb[itemId], 32, 48, slotX, junk)
                //   h.c = itemPicture ; ua.Bb = itemMask ; sg = spriteItem base
                surface.spriteClipping(slotY,
                        GameData.itemPicture[itemId],
                        0, false, 0,
                        sg + GameData.itemMask[itemId],
                        32, 48, slotX, dummy ^ -15251);
                // Stack-count label for non-stackable items (itemStackable == 0).
                // obf: if (fa.e[itemId] == 0) li.a(""+xe[slot], slotX+1, slotY+10, 0xFFFF00, false, 1)
                if (GameData.itemStackable[itemId] == 0) {
                    surface.drawString("" + xe[slot],            // xe = inventoryStackCount[]
                            slotX + 1, slotY + 10, 0xFFFF00, false, 1);
                }
            }
        }

        // ── Draw grid dividers ───────────────────────────────────────────────
        // Vertical lines between the 5 columns.  obf: li.b(invX+49*col, 36, 0, cl/5*34, 0)
        for (int col = 1; col <= 4; col++) {
            surface.drawLineVert(invX + 49 * col, 36, 0, cl / 5 * 34, 0);
        }
        // Horizontal lines between rows.  obf: li.b(245, 0, invX, 36+34*row, (byte)76)
        for (int row = 1; row <= cl / 5 - 1; row++) {
            surface.drawLineHoriz(245, 0, invX, 36 + 34 * row, (byte)76);
        }

        // ── Mouse / menu handling ────────────────────────────────────────────
        if (!buildMenu) return;

        // Convert raw screen coords to inventory-relative coords.
        // obf: invX(reused) = 248 + -li.u + I ;  row = xb - 36   (I = mouseX, xb = mouseY)
        int mouseRelX = 248 - surface.width + I;
        int mouseRelY = xb - 36;

        // obf: if (mouseRelX >= 0 && -1 >= ~mouseRelY && 248 > mouseRelX && cl/5*34 > mouseRelY)
        //   -1 >= ~mouseRelY  ⟺  mouseRelY >= 0
        if (mouseRelX < 0 || mouseRelY < 0 || mouseRelX >= 248 || mouseRelY >= cl / 5 * 34) {
            return;
        }

        // Note clean-base index order: (mouseRelY/34)*5 + mouseRelX/49.
        int hoveredSlot = mouseRelY / 34 * 5 + mouseRelX / 49;
        if (hoveredSlot >= lc) return; // lc = inventoryItemsCount

        int itemId = vf[hoveredSlot]; // vf = inventoryItemId[]

        // ── Case 1: Spell selected → Cast [spell] on item (600) ───────────────
        if (af >= 0) {
            // Only item-target spells (type 3) get an entry.
            // obf: if (~qb.e[af] != -4) return;  ⟺  if (spellType[af] != 3) return;
            if (GameData.spellType[af] != 3) {
                return;
            }
            // STRINGS[34] = "@lre@"  STRINGS[46] = "Cast "  STRINGS[50] = " on"
            zh.a(hoveredSlot,
                    STRINGS[34] + GameData.itemName[itemId],          // obf: ac.x[itemId]
                    600,                                              // CAST_SPELL_ON_ITEM
                    STRINGS[46] + GameData.spellName[af] + STRINGS[50],
                    af, 3296);
            return;
        }

        // ── Case 2: No item selected (Bh < 0) → full slot menu ────────────────
        // obf: if (-1 < ~Bh)  ⟺  ~Bh >= 0  ⟺  Bh < 0
        if (Bh < 0) {
            slotMenu: {
                // Equipped item → Remove (620); else wearable → Wear/Wield (630).
                // obf: if (-2 == ~Aj[hoveredSlot])  ⟺  Aj[hoveredSlot] == 1
                if (Aj[hoveredSlot] == 1) {
                    zh.a(hoveredSlot, 620, false,
                            STRINGS[69],                              // "Remove"
                            STRINGS[34] + GameData.itemName[itemId]);
                    break slotMenu; // else-if: equipped items are not also Wear/Wield-able here
                }
                // obf: if (-1 != ~mb.k[itemId])  ⟺  itemWearable[itemId] != -1
                if (GameData.itemWearable[itemId] != -1) {
                    // (24 & wearable) selects Wield vs Wear.  obf: if (0 == (24 & mb.k[itemId])) "Wear" else "Wield"
                    String wearText = (24 & GameData.itemWearable[itemId]) == 0
                            ? STRINGS[68]   // "Wield"
                            : STRINGS[72];  // "Wear"
                    zh.a(hoveredSlot, 630, false,
                            wearText,
                            STRINGS[34] + GameData.itemName[itemId]);
                }
            }

            // Custom item command (e.g. "Eat", "Read") if non-empty.  obf: lb.ac = Scene-held itemCommand[]
            if (!GameData.itemCommand[itemId].equals("")) {
                zh.a(hoveredSlot, 640, false,
                        GameData.itemCommand[itemId],
                        STRINGS[34] + GameData.itemName[itemId]);
            }
            zh.a(hoveredSlot, 650, false, STRINGS[71],   // "Use"
                    STRINGS[34] + GameData.itemName[itemId]);
            zh.a(hoveredSlot, 660, false, STRINGS[67],   // "Drop"
                    STRINGS[34] + GameData.itemName[itemId]);
            zh.a(itemId, 3600, false, STRINGS[51],       // "Examine" (target = itemId, not slot)
                    STRINGS[34] + GameData.itemName[itemId]);
            return;
        }

        // ── Case 3: Item selected (Bh >= 0) → Use [selected] with (610) ───────
        // STRINGS[38] = "Use "  STRINGS[53] = " with"
        zh.a(hoveredSlot,
                STRINGS[34] + GameData.itemName[itemId],
                610,
                STRINGS[38] + ig + STRINGS[53],          // "Use [selectedItemName] with"
                Bh, dummy ^ -14196);
    }

    // -----------------------------------------------------------------
    // menuHitTest (getInventoryCount)  obf: private final int b(int,int)
    // (oracle: getInventoryCount(int id))
    // -----------------------------------------------------------------

    /**
     * Counts how many of item {@code itemId} are held across all inventory slots.
     *
     * For non-stackable items ({@code itemStackable[itemId] == 0}): sums the
     * per-slot stack-count values ({@code xe[slot]}) across matching slots.
     * For stackable items: each matching slot counts as +1.
     *
     * <p>Note: skeleton name "menuHitTest" is a placeholder; semantics match
     * oracle {@code getInventoryCount(int id)}.</p>
     *
     * obf: private final int b(int param1, int param2)
     */
    private final int menuHitTest(int dummy, int itemId) {
        int count = 0;
        for (int slot = 0; slot < lc; slot++) { // obf: while (~slot > ~lc); lc = inventoryItemsCount
            if (vf[slot] == itemId) {            // obf: ~vf[slot] == ~param2; vf = inventoryItemId[]
                // obf: if (~fa.e[itemId] != -2)  ⟺  itemStackable[itemId] != 1
                if (GameData.itemStackable[itemId] != 1) {
                    // Non-stackable: accumulate individual quantities.
                    count += xe[slot];           // xe = inventoryStackCount[]
                } else {
                    // Stackable: count matching slots.
                    count++;
                }
            }
        }
        // Anti-tamper guard: if (param1 < 83) this.h((byte)87); — stripped.
        return count;
    }

    // -----------------------------------------------------------------
    // pointInRect (isItemEquipped)  obf: private final boolean e(int,int)
    // -----------------------------------------------------------------

    /**
     * Returns {@code true} if the inventory contains item {@code itemId} in an
     * equipped slot ({@code Aj[slot] == 1}).
     *
     * When no slot matches, returns {@code mustBeActive != 1}.
     *
     * <p>Note: skeleton name "pointInRect" is a placeholder; the logic is a
     * per-slot equipped-state lookup for a specific item ID.</p>
     *
     * obf: private final boolean e(int param1, int param2)
     */
    private final boolean pointInRect(int itemId, int mustBeActive) {
        for (int slot = 0; slot < lc; slot++) { // obf: while (~slot > ~lc); lc = inventoryItemsCount
            // obf: if (~vf[slot] == ~param1 && 1 == Aj[slot])
            if (vf[slot] == itemId               // vf = inventoryItemId[]
                    && Aj[slot] == 1) {           // Aj = inventoryEquipped[]
                return true;
            }
        }
        // No matching equipped slot: return (mustBeActive != 1).  obf: return param2 != 1;
        return mustBeActive != 1;
    }

    // -----------------------------------------------------------------
    // pointInPanel (isEquipSlotActive)  obf: private final boolean a(byte,int,int)
    // -----------------------------------------------------------------

    /**
     * Tests whether an equipment-slot group ({@code slotType}) currently has an
     * item equipped, by probing a fixed set of item IDs per group via
     * {@link #pointInRect}.  Used by the equipment-tab HUD to highlight slots.
     *
     * <pre>
     *   31 → weapon / shield row  (item IDs 197, 615, 682)
     *   32 → body armour row      (item IDs 102, 616, 683)
     *   33 → leg armour row       (item IDs 101, 617, 684)
     *   34 → head / special row   (item IDs 103, 618, 685)
     *   other → getInventoryCount(94, slotType) &ge; minCount
     * </pre>
     *
     * <p>Note: skeleton name "pointInPanel" is a placeholder.  The junk args
     * {@code dummy ^ -69} / {@code dummy + 71} all evaluate to {@code 1} because
     * the anti-tamper guard pins {@code dummy == -70}.</p>
     *
     * obf: private final boolean a(byte param1, int param2, int param3)
     */
    private final boolean pointInPanel(byte dummy, int minCount, int slotType) {
        // Anti-tamper: if (param1 != -70) return true; — stripped (dummy is pinned to -70).

        // obf: -32 == ~slotType  ⟺  slotType == 31
        if (slotType == 31) {
            if (pointInRect(197, 1)) return true;
            if (pointInRect(615, 1 /* obf: dummy ^ -69 */)) return true;
            if (pointInRect(682, 1 /* obf: dummy + 71  */)) return true;
        }
        // obf: -33 == ~slotType  ⟺  slotType == 32
        if (slotType == 32) {
            if (pointInRect(102, 1)) return true;
            if (pointInRect(616, 1)) return true;
            if (pointInRect(683, 1)) return true;
        }
        if (slotType == 33) {
            if (pointInRect(101, 1)) return true;
            if (pointInRect(617, 1)) return true;
            if (pointInRect(684, 1)) return true;
        }
        if (slotType == 34) {
            if (pointInRect(103, 1)) return true;
            if (pointInRect(618, 1 /* obf: dummy + 71 */)) return true;
            if (pointInRect(685, 1)) return true;
        }
        // Fallback: count of item 94 across inventory vs required minCount.
        // obf: return this.b(94, slotType) >= param2;  (menuHitTest(dummy=94, itemId=slotType))
        return menuHitTest(94, slotType) >= minCount;
    }

    // -----------------------------------------------------------------
    // pollInput  obf: private final void n(int)
    // -----------------------------------------------------------------

    /**
     * Per-tick window-size poll and layout reset.  Detects host-window resize
     * and rebuilds the panel layout.
     *
     * <p>Selects the AWT {@link java.awt.Component} to query for dimensions:</p>
     * <ul>
     *   <li>{@code hj} set (desktop) + socket open → {@code ClientStream.socket} ({@code da.gb})</li>
     *   <li>{@code hj} set but no socket → {@code this} (the Applet itself)</li>
     *   <li>Applet mode → {@code InputState.applet} ({@code kb.a})</li>
     * </ul>
     *
     * <p>Sets {@code Rh} (screenWidth), {@code Hf} (screenHeight), zeroes the Y
     * origin {@code K}, recomputes centering offset {@code Eb = (screenWidth -
     * gameWidth)/2}, then calls {@link #resetPanels}({@code 49}).</p>
     *
     * obf: private final void n(int param1)
     */
    private final void pollInput(int dummy) {
        // Select which AWT Component hosts the display surface.
        Object hostComponent;
        if (hj) {                          // hj = isDesktopMode
            if (da.gb != null) {           // da.gb = ClientStream.socket
                hostComponent = da.gb;
            } else {
                hostComponent = this;      // standalone without active socket
            }
        } else {
            hostComponent = kb.a;          // applet mode: InputState.applet
        }

        // Anti-tamper: if (param1 > -77) Ee = 30; — keep side effect (brief debounce timer).
        if (dummy > -77) {
            Ee = 30;
        }

        // Query display dimensions from the host component.
        Rh = ((java.awt.Component) hostComponent).getSize().width;  // screenWidth
        Hf = ((java.awt.Component) hostComponent).getSize().height; // screenHeight
        K  = 0;                                                     // screenOriginY
        // Horizontal centering: (screenWidth - gameWidth) / 2.  obf: (-Wd + Rh) / 2
        Eb = (-Wd + Rh) / 2;
        // Rebuild all Panel widgets for the current window size.
        resetPanels((byte) 49); // obf: this.p((byte)49)
    }
