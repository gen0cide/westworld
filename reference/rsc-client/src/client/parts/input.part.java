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
// Field notes (undocumented in skeleton – named here by behaviour):
//   Kk  (int[8192])  clickHistoryX    – ring buffer of recent click screen-X values
//   uj  (int[8192])  clickHistoryType – ring buffer of recent click type/flags
//   nk  (int)        clickRingHead    – ring buffer write index (masked 0x1FFF)
//   oh  (int[])      equipBonusDisplay – equipment stat array (cleared on top-bar click)
//   ai  (int)        selectedAction   – currently selected action index (-1 = none)
//   bj  (int)        walkPendingTimer – set to 1000 after requestLogout(0) fires
//   af  (int)        selectedSpell    – selected spell index (-1 = none)
//   Bh  (int)        selectedItemInventoryIndex – selected inventory slot (-1 = none)
//   ig  (String)     selectedItemName
//   Pg  (boolean)    isMember
//   rg  (ta[500])    playersLast      – player entities from previous tick
//   zg,sh,sk,Qg,Lf,Ki (int) – camera / region offset fields
//
// Key method cross-references (using skeleton proposed names):
//   requestLogout(0)  →  void B(int)  (opcode 102 LOGOUT)
//   resetChatInput(x) →  void o(byte)
//   resetPanels(x)    →  void p(byte)
//   friendsList       →  zh  (wb / MessageList, reused as shared context-menu list)
//                         NOTE: skeleton labels zh="friendsList" but at runtime it
//                         acts as the right-click option-menu accumulator.

    // -----------------------------------------------------------------
    // handleGameClick  obf: void a(int,int,int,int)  @L19905
    // -----------------------------------------------------------------

    /**
     * Records a world click in the click-history ring buffer and triggers an
     * auto-logout if a bot-like repeated-identical-click pattern is detected.
     *
     * Ring buffers {@code Kk}/{@code uj} hold up to 8191 (0x1FFF) recent
     * click (x, type) pairs.  After storing the new click, the method walks
     * backwards from index 10 to 3999, looking for the current (x,type) pair
     * at exactly that stride.  If found, it verifies the whole preceding run
     * [1..stride-1] is also consistent; if it is and no action/walk is pending,
     * opcode 102 (LOGOUT) is sent.
     *
     * Also clears the equipment-stat display overlay ({@code oh}) when the
     * click Y is in the top-bar area (y &le; 87).
     *
     * obf: final void a(int param1, int param2, int param3, int param4)
     */
    final void handleGameClick(int clickX, int clickY, int unused, int clickType) {
        // Store this click in the ring buffer at the current head position.
        Kk[nk] = clickX;
        uj[nk] = clickType;
        nk = (nk + 1) & 0x1FFF; // advance write head, wraps at 8191

        // Clicking in the top HUD area (y <= 87) dismisses the equipment stat overlay.
        if (clickY <= 87) {
            oh = null;
        }

        // Scan recent ring history for a bot-like "exact same click N ticks in a row" pattern.
        // If detected with no pending walk or action, trigger automatic logout (opcode 102).
        for (int stride = 10; stride < 4000; stride++) {
            int slotA = (nk - stride) & 0x1FFF; // ring slot exactly 'stride' ticks ago
            // Does that historic slot carry the same (x, type) as the current click?
            if (Kk[slotA] != clickX || uj[slotA] != clickType) {
                continue;
            }
            // Found a match at distance 'stride'.  Now walk the inner range [1..stride-1],
            // comparing each recent slot (slotB) against its stride-offset mirror (slotC).
            // slotB: position (nk - inner)     – what we clicked 'inner' ticks ago
            // slotC: position (slotA - inner)  – that same slot's partner 'stride' ticks earlier
            boolean hasMismatch = false;
            for (int inner = 1; inner < stride; inner++) {
                int slotB = (nk    - inner) & 0x1FFF; // recent entry at offset 'inner'
                int slotC = (slotA - inner) & 0x1FFF; // mirror at (nk - stride - inner)

                // If slotC doesn't match current click → note a mismatch (non-trivial run).
                if (Kk[slotC] != clickX || uj[slotC] != clickType) {
                    hasMismatch = true;
                }
                // If slotB and slotC differ → the stride-based pattern is broken; skip this stride.
                if (Kk[slotB] != Kk[slotC] || uj[slotB] != uj[slotC]) {
                    break; // inner loop exits, outer loop increments stride
                }
                // At the last inner iteration: solid repeated-click run detected.
                // Conditions for logout: non-trivial run, no active spell/selection, no walk pending.
                if (inner == stride - 1 && hasMismatch && ai == -1 && bj == 0) {
                    requestLogout(0); // opcode 102 LOGOUT – anti-bot measure
                    return;
                }
            }
        }
    }

    // -----------------------------------------------------------------
    // buildClickMenu  obf: void a(int,int)  @L35711
    // -----------------------------------------------------------------

    /**
     * Builds the right-click context menu for a player entity that was
     * scene-picked under the cursor.
     *
     * Appends menu entries to the shared context-menu list ({@code zh /
     * friendsList}) via {@code wb.a()} calls.  The menu items added depend
     * on whether a spell or inventory item is currently selected:
     *
     * <pre>
     *  selectedSpell &ge; 0   → 800  Cast [spell] on [player]
     *  selectedItem &ge; 0    → 810  Use [item] with [player]
     *  (default, in range)    → 805/2805  Attack
     *  members, non-wild      → 2806  Duel with
     *  always (non-wild)      → 2810  Trade with
     *                         → 2820  Follow
     *                         → 2833  Report abuse (via chat-markup link)
     * </pre>
     *
     * The level-difference colour codes embedded in {@code STRINGS[]}:
     * {@code @or1..@red@} (we outmatch them) / {@code @gr1..@gre@} (they outmatch us).
     *
     * obf: private final void a(int param1, int param2)
     */
    private final void buildClickMenu(int playerIndex, int dummy) {
        // Anti-tamper: if (dummy != -12) resetChatInput(-32); — stripped.

        GameCharacter player = playersLast[playerIndex]; // obf: rg[playerIndex] (ta)
        String playerName  = player.name;               // obf: ta.c

        // Compute wilderness boundary flag.
        // worldYBound < 0 signals the wilderness region (combat always allowed).
        int worldYBound = -zg - sh + -sk + 2203; // local-region Y boundary
        if (2640 <= Qg + Lf - (-Ki)) {
            worldYBound = -50; // in wilderness / above 2640 map boundary
        }

        // --- Level-difference colour tag (oracle: "@or1@".."@gre@") ---
        String colourTag = "";
        int levelDiff    = 0;
        if (localPlayer.level > 0 && player.level > 0) { // obf: wi.s, ta.s
            levelDiff = player.level - localPlayer.level;
        }
        // Ascending colour severity: slight → extreme difference.
        if (levelDiff <  0)  colourTag = STRINGS[40]; // "@or1@" – we have slight advantage
        if (levelDiff < -3)  colourTag = STRINGS[39]; // "@or2@"
        if (levelDiff < -6)  colourTag = STRINGS[49]; // "@or3@"
        if (levelDiff < -9)  colourTag = STRINGS[10]; // "@red@" – big advantage
        if (levelDiff >  0)  colourTag = STRINGS[35]; // "@gr1@" – they have slight advantage
        if (levelDiff >  3)  colourTag = STRINGS[37]; // "@gr2@"
        if (levelDiff >  6)  colourTag = STRINGS[47]; // "@gr3@"
        if (levelDiff >  9)  colourTag = STRINGS[27]; // "@gre@" – they massively outmatch us

        // Suffix string e.g. " @gr2@(level-64)"
        // STRINGS[42] = "(level-"
        String levelSuffix = " " + colourTag + STRINGS[42] + player.level + ")";

        // ── Case 1: Spell selected ──────────────────────────────────────────
        if (af >= 0) { // af = selectedSpell
            // Only build a menu entry if mouse button is left (1) or right (2).
            // qb.e = InputState.mouseButtons[];  ~qb.e[af] == -3 → qb.e[af] == 2
            if (InputState.mouseButtons[af] != 1 && InputState.mouseButtons[af] != 2) {
                return; // wrong button state – no menu entry
            }
            // STRINGS[46] = "Cast "  STRINGS[50] = " on"  (concatenated: "Cast [spell] on")
            // BitBuffer.spellNames (ja.L) = spell name lookup table
            friendsList.a(player.serverIndex,              // target entity id
                    STRINGS[15] + playerName + levelSuffix, // menuText2: "@whi@Name (level-X)"
                    800,                                   // actionId: CAST_SPELL_ON_PLAYER
                    STRINGS[46] + BitBuffer.spellNames[af] + STRINGS[50], // menuText1
                    af, 3296);
            return;
        }

        // ── Case 2: Inventory item selected ────────────────────────────────
        if (Bh >= 0) { // Bh = selectedItemInventoryIndex
            // STRINGS[38] = "Use "  STRINGS[53] = " with"  → "Use [item] with"
            friendsList.a(player.serverIndex,
                    STRINGS[15] + playerName + levelSuffix,
                    810,                                   // actionId: USE_ITEM_WITH_PLAYER
                    STRINGS[38] + ig + STRINGS[53],        // "Use [selectedItemName] with"
                    Bh, 3296);
            return;
        }

        // ── Case 3: No selection — standard player menu ─────────────────────
        // Attack option: only when player is in world range AND in a combat zone.
        // Tile-range test: player currentX within ~2203 tiles of local region.
        boolean inRange = worldYBound > 0
                && ((-64 + player.currentX) / magicLoc   // obf: Ug = magicLoc (tile scale)
                        - (-sk + -zg)) <= 2203;           // obf: sk/zg = local region offset

        if (inRange) {
            // Attack actionId: 805 if wilderness PvP (level diff < 0 or diff >= 5),
            //                 2805 otherwise (safe zone – e.g. duel arena vicinity).
            int attackActionId = (levelDiff < 0 || levelDiff >= 5) ? 2805 : 805;
            // STRINGS[48] = "Attack"
            friendsList.a(player.serverIndex, attackActionId, false,
                    STRINGS[48],
                    STRINGS[15] + playerName + levelSuffix);
        }

        // Duel with (members server, non-wilderness only).
        // STRINGS[118] = "Duel with"  actionId 2806
        if (isMember) { // obf: Pg
            friendsList.a(player.serverIndex, 2806, false,
                    STRINGS[118],
                    STRINGS[15] + playerName + levelSuffix);
        }

        // Trade with (2810), Follow (2820) — always available outside combat.
        // STRINGS[116] = "Trade with"  STRINGS[119] = "Follow"
        friendsList.a(player.serverIndex, 2810, false,
                STRINGS[116],
                STRINGS[15] + playerName + levelSuffix);
        friendsList.a(player.serverIndex, 2820, false,
                STRINGS[119],
                STRINGS[15] + playerName + levelSuffix);

        // Report-abuse link (action 2833); passes player name separately for the
        // abuse-report dialog.  STRINGS[120] = "Report abuse"
        // wb.a(String menuText1, String menuText2, String name, int actionId, int/Object displayName, byte flags)
        friendsList.a(STRINGS[120],
                STRINGS[15] + playerName + levelSuffix,
                player.name,      // ta.c – lookup key for report-abuse dialog
                2833,
                player.displayName, // ta.C – display name shown in report form
                (byte)103);
    }

    // -----------------------------------------------------------------
    // handleInventoryClick  obf: void a(int,boolean)  @L40645
    // -----------------------------------------------------------------

    /**
     * Renders the inventory panel and, when {@code buildMenu} is {@code true},
     * processes mouse-hover/click over inventory slots to populate the
     * right-click option list.
     *
     * <p>Layout: 5 columns × {@code inventoryMaxSlots/5} rows of 49×34 px
     * slots anchored at x = {@code surface.width − 248}.</p>
     *
     * <p>Equipped slots (where {@code inventoryEquipped[slot] == 1}) are
     * drawn with a red tint (0xFF0000); unequipped slots use a grey
     * alpha-blend derived from ISAAC's pseudo-random 181-grey.</p>
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
     * obf: private final void a(int param1, boolean param2)
     */
    private final void handleInventoryClick(int dummy, boolean buildMenu) {
        // Anti-tamper: if (dummy != -15252) this.b(-79,(byte)75,-83); — stripped.

        // Inventory panel left edge: surfaceWidth − 248.
        // li.u = SurfaceSprite.width (screen pixel width of surface).
        int invX = surface.width - 248; // obf: -248 + this.li.u

        // Draw border box around the inventory panel region.
        // li.b(width, x2, y, x1) — exact surface call semantics; tg = panel handle/offset.
        surface.drawBorder(-1, tg + 1, 3, invX); // obf: this.li.b(-1, this.tg+1, 3, invX)

        // ── Render each slot ─────────────────────────────────────────────────
        for (int slot = 0; slot < cl; slot++) { // cl = inventoryMaxSlots (35)
            int slotX = invX + (slot % 5) * 49;
            int slotY  = 36 + (slot / 5) * 34;

            if (slot < lc && Aj[slot] == 1) { // lc = inventoryItemsCount; Aj = inventoryEquipped
                // Equipped item slot: solid red tint.
                surface.drawBoxAlpha(slotX, slotY, 34, 0, 49, 0xFF0000);
            } else {
                // Normal slot: grey alpha blend.
                // ISAAC.a(181, junk, 181, 181) → stable grey ARGB; junk from dummy param.
                surface.drawBoxAlpha(slotX, slotY, 34, 0, 49,
                        ISAAC.a(181, dummy ^ -7922, 181, 181));
            }

            if (slot < lc) { // lc = inventoryItemsCount
                int itemId = vf[slot]; // vf = inventoryItemId[]
                // Draw item sprite: li.a(y, pictureIdx, mask, mirrorH, mirrorV,
                //                       maskOffset, width, height, x, dummy)
                // h.c[itemId] = item picture index (TextEncoder.itemPictures)
                // sg + ua.Bb[itemId] = spriteBase + Surface.itemMasks[itemId]
                surface.spriteClipping(slotY,
                        h.c[itemId],                        // item sprite picture index
                        0, false, 0,
                        sg + ua.Bb[itemId],                 // mask + base offset
                        32, 48, slotX, dummy ^ -15251);
                // Draw stack-count label for non-stackable items
                // (fa.e[itemId] == 0 → not stackable → show quantity xe[slot])
                if (fa.e[itemId] == 0) { // fa = EntityDef; e[]=itemStackable[]
                    surface.drawString("" + xe[slot],       // xe = inventoryStackCount[]
                            slotX + 1, slotY + 10, 0xFFFF00, false, 1);
                }
            }
        }

        // ── Draw grid dividers ───────────────────────────────────────────────
        for (int col = 1; col <= 4; col++) {
            surface.drawLineVert(invX + col * 49, 36, 0, cl / 5 * 34, 0);
        }
        for (int row = 1; row <= cl / 5 - 1; row++) {
            surface.drawLineHoriz(245, 0, invX, 36 + row * 34, (byte)76);
        }

        // ── Mouse / click handling ───────────────────────────────────────────
        if (!buildMenu) return;

        // Convert raw screen coords to inventory-relative coords.
        // I = mouseX;  li.u = surfaceWidth;  xb = mouseY.
        int mouseRelX = 248 + (I - surface.width); // obf: 248 + (-li.u + I)
        int mouseRelY = xb - 36;                   // obf: xb - 36

        if (mouseRelX < 0)  return;
        if (mouseRelY < 0)  return;
        if (mouseRelX >= 248) return;
        if (mouseRelY >= cl / 5 * 34) return;

        int hoveredSlot = (mouseRelY / 34) * 5 + (mouseRelX / 49);
        if (hoveredSlot >= lc) return; // lc = inventoryItemsCount

        int itemId = vf[hoveredSlot]; // vf = inventoryItemId[]

        // ── Case 1: Spell selected (af >= 0) → Cast [spell] on item ─────────
        if (af >= 0) {
            // Only allow if mouse is using button state "Use" (qb.e[af] != -4 → not drop).
            if (InputState.mouseButtons[af] != -4) {
                // STRINGS[34] = "@lre@"  (item name colour prefix)
                // STRINGS[46] = "Cast "  STRINGS[50] = " on"
                friendsList.a(hoveredSlot,
                        STRINGS[34] + EntityDef.itemName[itemId],
                        600,                                 // CAST_SPELL_ON_ITEM
                        STRINGS[46] + BitBuffer.spellNames[af] + STRINGS[50],
                        af, 3296);
            }
            return;
        }

        // ── Case 2: Item selected (Bh >= 0) → Use [selected] with ───────────
        if (Bh >= 0) {
            // If the hovered slot is a worn item (Aj[slot] == -2?) or in range:
            if (Aj[hoveredSlot] == -2) {
                // Hovered slot is equipped → show Remove action.
                friendsList.a(hoveredSlot, 620, false,
                        STRINGS[69],                         // "Remove"
                        STRINGS[34] + EntityDef.itemName[itemId]);
                return;
            }
            // Wearable item → Wear/Wield action (630).
            if (mb.k[itemId] != -1) { // mb = Utility; k[] = itemEquipType?
                if ((24 & mb.k[itemId]) != 0) {             // wield-type bit
                    friendsList.a(hoveredSlot, 630, false,
                            STRINGS[68],                     // "Wield"
                            STRINGS[34] + EntityDef.itemName[itemId]);
                } else {
                    friendsList.a(hoveredSlot, 630, false,
                            STRINGS[72],                     // "Wear"
                            STRINGS[34] + EntityDef.itemName[itemId]);
                }
            }
            // ── Standard entries when item is selected ─
            // Custom item command (e.g., "Eat", "Read") if non-empty.
            // lb.ac = Scene.itemCommands[]
            if (!lb.ac[itemId].equals("")) {
                friendsList.a(hoveredSlot, 640, false,
                        lb.ac[itemId],
                        STRINGS[34] + EntityDef.itemName[itemId]);
            }
            friendsList.a(hoveredSlot, 650, false, STRINGS[71],  // "Use"
                    STRINGS[34] + EntityDef.itemName[itemId]);
            friendsList.a(hoveredSlot, 660, false, STRINGS[67],  // "Drop"
                    STRINGS[34] + EntityDef.itemName[itemId]);
            friendsList.a(itemId, 3600, false, STRINGS[51],      // "Examine" (itemId, not slot)
                    STRINGS[34] + EntityDef.itemName[itemId]);

            // Use-with entry (action 610).
            // STRINGS[38] = "Use "  STRINGS[53] = " with"
            friendsList.a(hoveredSlot,
                    STRINGS[34] + EntityDef.itemName[itemId],
                    610,
                    STRINGS[38] + ig + STRINGS[53],          // "Use [selectedItemName] with"
                    Bh, dummy ^ -14196);
            return;
        }

        // ── Case 3: No selection — standard inventory slot menu ──────────────
        if (Aj[hoveredSlot] == 1) {
            // Item is currently equipped → offer Remove.
            friendsList.a(hoveredSlot, 620, false,
                    STRINGS[69],                             // "Remove"
                    STRINGS[34] + EntityDef.itemName[itemId]);
        } else if (mb.k[itemId] != -1) {
            // Item is wearable but not equipped → Wear or Wield.
            if ((24 & mb.k[itemId]) != 0) {
                friendsList.a(hoveredSlot, 630, false,
                        STRINGS[68],                         // "Wield"
                        STRINGS[34] + EntityDef.itemName[itemId]);
            } else {
                friendsList.a(hoveredSlot, 630, false,
                        STRINGS[72],                         // "Wear"
                        STRINGS[34] + EntityDef.itemName[itemId]);
            }
        }

        // Custom item command (e.g., "Eat", "Read").
        if (!lb.ac[itemId].equals("")) { // lb = Scene; ac = itemCommands[]
            friendsList.a(hoveredSlot, 640, false,
                    lb.ac[itemId],
                    STRINGS[34] + EntityDef.itemName[itemId]);
        }
        friendsList.a(hoveredSlot, 650, false, STRINGS[71],  // "Use"
                STRINGS[34] + EntityDef.itemName[itemId]);
        friendsList.a(hoveredSlot, 660, false, STRINGS[67],  // "Drop"
                STRINGS[34] + EntityDef.itemName[itemId]);
        friendsList.a(itemId, 3600, false, STRINGS[51],      // "Examine"
                STRINGS[34] + EntityDef.itemName[itemId]);
    }

    // -----------------------------------------------------------------
    // menuHitTest (getInventoryCount)  obf: int b(int,int)  @L3648  (client.HB()
    // -----------------------------------------------------------------

    /**
     * Counts how many of item {@code itemId} are held across all inventory slots.
     *
     * For non-stackable items ({@code fa.e[itemId] == 0}): sums the individual
     * stack-count values ({@code xe[slot]}) across matching slots, giving the
     * total number of that item regardless of how many stacks it spans.
     * For stackable items ({@code fa.e[itemId] == 1}): each matching slot
     * counts as +1 (one distinct stack).
     *
     * <p>Note: the skeleton name "menuHitTest" is a placeholder; the actual
     * semantics match oracle {@code mudclient.getInventoryCount(int id)}.</p>
     *
     * obf: private final int b(int param1, int param2)
     */
    private final int menuHitTest(int dummy, int itemId) {
        int count = 0;
        for (int slot = 0; slot < lc; slot++) { // lc = inventoryItemsCount
            if (vf[slot] == itemId) {            // vf = inventoryItemId[]
                if (fa.e[itemId] != 1) {         // fa = EntityDef; e[] = itemStackable[]
                    // Non-stackable: accumulate individual quantities.
                    count += xe[slot];            // xe = inventoryStackCount[]
                } else {
                    // Stackable (coins, arrows, etc.): count matching slots.
                    count++;
                }
            }
        }
        // Anti-tamper guard: if (dummy < 83) drawSocialDialog((byte)87); — stripped.
        return count;
    }

    // -----------------------------------------------------------------
    // pointInRect (isItemEquipped)  obf: boolean e(int,int)  @L38766
    // -----------------------------------------------------------------

    /**
     * Returns {@code true} if the inventory contains item {@code itemId} in an
     * equipped slot ({@code Aj[slot] == 1}).
     *
     * When no slot matches, the fallback behaviour depends on {@code mustBeActive}:
     * if {@code mustBeActive == 1} returns {@code false}; otherwise {@code true}.
     *
     * <p>Note: the skeleton name "pointInRect" is a placeholder; the actual
     * logic is a per-slot equipped-state lookup for a specific item ID.</p>
     *
     * obf: private final boolean e(int param1, int param2)
     */
    private final boolean pointInRect(int itemId, int mustBeActive) {
        for (int slot = 0; slot < lc; slot++) { // lc = inventoryItemsCount
            if (vf[slot] == itemId              // vf = inventoryItemId[]
                    && Aj[slot] == 1) {          // Aj = inventoryEquipped[]
                return true;
            }
        }
        // If no matching equipped slot found: return (mustBeActive != 1).
        return mustBeActive != 1;
    }

    // -----------------------------------------------------------------
    // pointInPanel (isEquipSlotActive)  obf: boolean a(byte,int,int)  @L38848
    // -----------------------------------------------------------------

    /**
     * Tests whether a given equipment-slot group ({@code slotType}) has any
     * item currently equipped in the inventory.  Used by the equipment-tab HUD
     * to decide which slot-overlay icons should be highlighted.
     *
     * <p>Slot-type constants (param {@code slotType}):</p>
     * <pre>
     *   31 → weapon / shield row  (item IDs: 197, 615, 682)
     *   32 → body armour row      (item IDs: 102, 616, 683)
     *   33 → leg armour row       (item IDs: 101, 617, 684)
     *   34 → head / special row   (item IDs: 103, 618+dummy, 685)
     *   other → getInventoryCount(94, slotType) &ge; minCount  (ammo/rune threshold)
     * </pre>
     *
     * <p>Note: the skeleton name "pointInPanel" is a placeholder; the actual
     * logic is an equipment-slot occupancy check.</p>
     *
     * obf: private final boolean a(byte param1, int param2, int param3)
     */
    private final boolean pointInPanel(byte dummy, int minCount, int slotType) {
        // Anti-tamper: if (dummy != -70) return true; — stripped.

        // Slot 31 (~slotType == -32 → slotType == 31): weapon/shield positions.
        if (slotType == 31) {
            if (pointInRect(197, 1))                return true;
            if (pointInRect(615, dummy ^ 0xBB))     return true; // dummy XOR 187
            if (pointInRect(682, dummy + 71))        return true;
        }
        // Slot 32: body-armour positions.
        if (slotType == 32) {
            if (pointInRect(102, 1)) return true;
            if (pointInRect(616, 1)) return true;
            if (pointInRect(683, 1)) return true;
        }
        // Slot 33: leg-armour positions.
        if (slotType == 33) {
            if (pointInRect(101, 1)) return true;
            if (pointInRect(617, 1)) return true;
            if (pointInRect(684, 1)) return true;
        }
        // Slot 34: head / special slot positions.
        if (slotType == 34) {
            if (pointInRect(103, 1))             return true;
            if (pointInRect(618, dummy + 71))    return true;
            if (pointInRect(685, 1))             return true;
        }
        // Fallback: count occurrences of item 94 (runes/ammo?) vs required minCount.
        // Uses menuHitTest(dummy=94, itemId=slotType) — note swapped roles here.
        return menuHitTest(94, slotType) >= minCount;
    }

    // -----------------------------------------------------------------
    // pollInput  obf: void n(int)  @L7738  (client.PD()
    // -----------------------------------------------------------------

    /**
     * Per-tick window-size poll and layout reset.  Called once per game tick
     * from the main loop to detect host-window resize events and rebuild the
     * panel layout accordingly.
     *
     * <p>Selects the AWT {@link java.awt.Component} to query for dimensions:</p>
     * <ul>
     *   <li>{@code hj} set (standalone/desktop) + socket open →
     *       {@code ClientStream.socket} ({@code da.gb})</li>
     *   <li>{@code hj} set but no socket → {@code this} (the Applet itself)</li>
     *   <li>Applet mode → {@code InputState.applet} ({@code kb.a})</li>
     * </ul>
     *
     * <p>Sets {@link #Rh screenWidth}, {@link #Hf screenHeight}, zeroes the
     * Y origin ({@code K}), recomputes the horizontal centering offset
     * ({@code Eb = (screenWidth - gameWidth) / 2}), then calls
     * {@link #resetPanels}({@code 49}) to tear down and rebuild all widgets.</p>
     *
     * obf: private final void n(int param1)
     */
    private final void pollInput(int dummy) {
        // Select which AWT Component hosts the display surface.
        Object hostComponent;
        if (hj && da.gb != null) {        // hj = isDesktopMode; da.gb = ClientStream.socket
            hostComponent = da.gb;
        } else if (hj) {
            hostComponent = this;         // standalone applet without active socket
        } else {
            hostComponent = kb.a;         // applet mode: InputState.applet
        }

        // Anti-tamper: if (dummy > -77) Ee = 30; — keep side effect (minor frame timer).
        if (dummy > -77) {
            Ee = 30; // obf: Ee – brief cooldown/debounce counter
        }

        // Query display dimensions from the host component.
        Rh = ((java.awt.Component) hostComponent).getSize().width;  // screenWidth
        Hf = ((java.awt.Component) hostComponent).getSize().height; // screenHeight
        K  = 0;                                                     // screenOriginY = 0
        // Horizontal centering: (screenWidth - gameViewportWidth) / 2.
        // Wd = gameWidth (fixed design width).
        Eb = (-Wd + Rh) / 2;  // layoutOffsetX
        // Rebuild all Panel widgets for the current window size.
        resetPanels((byte) 49); // obf: this.p((byte)49)
    }
