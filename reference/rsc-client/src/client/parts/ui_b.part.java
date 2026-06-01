// ===== ui_b =====
// Methods 19–36 of the "ui" group from MUDCLIENT_SKELETON.md.
// RE-AUDITED against decompiled/normalized-clean/client.java (ground truth) +
// cfr/client.java cross-ref. The previous version was written against a DEFECTIVE
// decompilation (72 client methods missing) and had numerous reconstructed-logic
// bugs (inverted compass strips, scrambled inventory-tab bounds, wrong mouse-button
// constants in the settings panel, missing draw-list capacity guard, etc.). Those
// are now fixed against the clean base.
//
// All methods belong to class Mudclient (obf: client), extends GameShell (obf: e).
// Naming: fields from MUDCLIENT_SKELETON.md; other classes from NAMING.md.
// Obfuscation stripped: opaque-predicate guards (vh/bl), profiling ++counters,
// try/catch ErrorHandler wrappers, anti-tamper dummy params, junk masks, and the
// ~x>~y / ~x==-N sign idioms (rewritten to plain >,<,==).
//
// Field quick-ref used below:
//   li         = surface         (SurfaceSprite / ba)
//   Xb         = graphics        (java.awt.Graphics)
//   Eb,K       = screen offset x,y
//   tg,dg      = panel column offsets (tg = right panel x-base, dg = left panel x-base)
//   Oi         = inventoryPanelH  (inventory panel height / bottom border y)
//   jk         = compassAngle    (0..3071, compass rotation counter)
//   Xd         = activePanel     (0=none,1=options,2=quest/skill,3=...)
//   qc         = inventoryTab    (0-6 inventory sub-tab index)
//   I          = mouseX          (inherited from GameShell)
//   xb         = mouseY
//   Cf         = mouseClickButton (0=none,1=left,2=right)
//   Bh         = selectedItem    (selected inventory item index, -1=none)
//   af         = selectedSpell   (selected spell index, -1=none)
//   Lf         = currentFloor
//   cl         = inventorySize   (inventory capacity, e.g. 30)
//   lc         = inventoryCount
//   vf         = inventoryItems  (int[] item ids)
//   xe         = inventoryQty    (int[] stack counts)
//   Aj         = inventoryEquipped (int[] equip flags)
//   eg         = mouseButtonMode (1=one-button, 77=two-button)
//   Jh         = clientStream    (ClientStream/da); Jh.f = outbound Buffer
//   Wf         = scrollMessageList (MessageList/wb) cleared by drawScrollList
//   zh         = friendsList     (MessageList/wb), used as menu builder
//   He         = chatList        (MessageList/wb)
//   ge         = panelGame       (Panel/qa)
//   yi         = panelDuel       (Panel/qa)  [reused: fatigue + server msgs]
//   Af         = panelQuest      (Panel/qa)  [reused as char-design widget container]
//   Qi,td      = panelDuel control ids (Qi=fatigue/title slot, td=server-msg slot)
//   Wd         = fatigueBarWidth
//   Ek         = world           (World/k)
//   Hh         = scene           (Scene/lb)
//   wi         = localPlayer     (GameCharacter/ta)
//   Kh,Yh,ne   = privacy: chatPrivateOn, tradePrivateOn, membersPrivateOn
//   Pg         = isMembersAccount
//   Kd         = isMembersWorld
//   De         = autoRetaliateOn (toggled+sent in opcode 64)
//   Yd         = combatModeSetting (3-state display 0/1/2)
//   dc         = mouseButtonsOne  (0=two,1=one) game option
//   Vg         = cameraModeAuto   (0=manual,1=auto) game option
//   ui         = membersOption    (0/1) game option
//   Bd         = showMenuBorder
//   Cb         = inputLine        (current text entry buffer)
//   e          = tempInputString  (scratch label for dialogs/menus)
//   ec         = reportAbuseTarget
//   Yb         = reportAbusePage
//   Vf         = inputMode (0=none,2=report-offence-picker,...)
//   Oj         = reportAbuseOffence
//   Ce         = reportAbuseMuteFlag
//   ue         = reportAbuseMuteConfirmed
//   Bj         = socialDialogMode (1=addFriend,2=sendPM,3=addIgnore)
//   Qd         = pmTarget
//   x          = pmInput
//   Ob         = submittedPmInput
//   Wk         = helpMenuOpen     (controls close-window panel height)
//   mh         = showCloseWindow  (modal close-window dialog visible)
//   Cj         = helpMenuTitle
//   qd         = closeButtonSpriteId
//   pj         = closeButtonPressed
//   od,zi,gl,gc,vk = scroll-list: options, width, height, x, openFlag
//   vj,fj      = drawListCount, drawListSize
//   ae,ci      = drawListIds, drawListCurrent
//   di,Xe      = drawListY, drawListYShadow
//   Gi         = drawListCapacity
//   Se,ye,vc   = wall-model coordinate arrays (z, x, type) ; hg = wallModels
//   kh         = objectModels (GameModel[]) ; sg = spriteBaseSlot
//   ei,Dg,Wh   = char-design colour palettes (int[]) ; Lh,Wg,hh,ld = palette indices ;
//                wg,dk,Vd = equip-sprite slot bases (obf names kept — see drawCharDesignControls)
//   Sf         = charDesignGender
//   o          = ISAAC (o.a = colour/scratch int helper, o.g = static String[] menu list)
//   u          = StringCodec (u.g = scratch)
//   w          = WorldEntity (w.a = name normalize/format, w.g = equip-sprite slot table)
//   STRINGS    = il[] decoded string pool

// ---------------------------------------------------------------------------

    /**
     * Render the minimap/compass panel (and fatigue bar on the stats tab).
     * Only draws the rotating 3-segment compass ring when a side panel is open
     * (activePanel in {0,1,2,3}); otherwise the panel area is left to the world view.
     * obf: void k(int)
     */
    private void drawMinimap(int param) {
        surface.i = false;        // clear sprite-clip flag (obf: li.i)
        Dc = false;               // clear dirty flag
        surface.a(true);          // flush surface buffer

        // Compass ring only rotates while a side panel is showing.
        // obf: if (~Xd==-1 || ~Xd==-2 || Xd==2 || ~Xd==-4)  →  Xd in {0,1,2,3}
        if (activePanel == 0 || activePanel == 1 || activePanel == 2 || activePanel == 3) {
            // compassAngle (jk) runs 0..3071; doubled+wrapped into a [0,3072) phase.
            int compassPos = 2 * compassAngle % 3072;
            // The ring is drawn in three 1024-wide slices; which slice we're in
            // selects one solid strip + an optional partial seam.
            if (compassPos < 1024) {
                surface.b(-1, dg, 10, 0);                       // first strip
                if (compassPos > 768) {                         // seam into next slice
                    surface.a(1 + dg, 0, 0, compassPos - 768, 10);
                }
            } else if (compassPos < 2048) {
                surface.b(-1, 1 + dg, 10, 0);                   // second strip
                if (compassPos > 1792) {
                    surface.a(tg - -10, 0, 0, compassPos - 1792, 10);
                }
            } else {
                surface.b(-1, tg - -10, 10, 0);                 // third strip
                if (compassPos > 2816) {
                    surface.a(dg, 0, 0, compassPos - 2816, 10);
                }
            }
        }

        // Special token 2540: keep the ground-item overlay; otherwise clear it.
        if (param != 2540) {
            inventoryGroundOverlay = null;   // obf: of
        }

        // No active panel → tick the game panel (resets quest-list scroll).
        if (activePanel == 0) {
            panelGame.a((byte)-63);          // obf: ge.a
        }

        // Stats/skills tab open → draw the fatigue bar.
        if (activePanel == 2) {
            String fatStr = panelDuel.g(fatigueControlId, 4);   // obf: yi.g(Qi,4)
            if (fatStr != null && fatStr.length() > 0) {
                surface.c(100, 0, 30, 0, 185, fatigueBarWidth, 0);   // obf: li.c(...,Wd,0)
            }
            panelDuel.a((byte)-52);          // obf: yi.a — tick panel
        }

        surface.b(-1, tg + 22, inventoryPanelH, 0);   // bottom border (obf: Oi)
        surface.a(graphics, Eb, 256, K);              // blit panel to AWT
    }

// ---------------------------------------------------------------------------

    /**
     * Inventory-area sub-tab hover/leave tracking (no drawing here — it only mutates
     * the active sub-tab index qc based on the cursor position over the right-side
     * tab strip). The tab strip lives at the right edge: x ∈ [li.u-200, li.u-3].
     *
     * All conditions below are the de-obfuscated forms of the clean ~-idioms, e.g.
     *   ~(li.u-35) >= ~I        →  I >= li.u-35
     *   ~xb <= -4               →  xb >= 3
     *   ~I > ~(li.u-3)          →  I < li.u-3
     * obf: void D(int)
     */
    private void drawInventoryTab(int param) {
        // --- enter a sub-tab from the closed state (qc==0) or the row-1 state (qc==1) ---
        if (qc == 0 && I >= surface.u - 35 && xb >= 3 && I < surface.u - 3 && xb < 35) {
            qc = 1;
        }
        if (qc == 1 && I >= surface.u - 68 && xb >= 3 && I < surface.u - 36 && xb < 35) {
            qc = 2;
            charDesignWobbleX = (int)(13.0 * Math.random()) - 6;    // obf: Df
            charDesignWobbleY = (int)(Math.random() * 23.0) - 11;   // obf: sd
        }
        if (qc == 1 && I >= surface.u - 101 && xb >= 3 && I < surface.u - 69 && xb < 35) {
            qc = 3;
        }
        if (qc == 0 && I >= surface.u - 134 && xb >= 3 && I < surface.u - 102 && xb < 35) {
            qc = 4;
        }
        if (qc == 1 && I >= surface.u - 167 && xb >= 3 && I < surface.u - 135 && xb < 35) {
            qc = 5;
        }
        if (param != 1) {
            currentFloor = -32;   // obf: Lf  (dummy reset when not called with sentinel 1)
        }
        if (qc == 1 && I >= surface.u - 200 && xb >= 3 && I < surface.u - 168 && xb < 35) {
            qc = 6;
        }
        // --- re-select from any open sub-tab when cursor is over the narrower (26px) header ---
        if (qc != 0 && I >= surface.u - 35 && xb >= 3 && I < surface.u - 3 && xb < 26) {
            qc = 1;
        }
        if (qc != 0 && qc != 2 && I >= surface.u - 68 && xb >= 3 && I < surface.u - 36 && xb < 26) {
            qc = 2;
            charDesignWobbleY = -11 + (int)(23.0 * Math.random());
            charDesignWobbleX = -6 + (int)(13.0 * Math.random());
        }
        if (qc != 0 && I >= surface.u - 101 && xb >= 3 && I < surface.u - 69 && xb < 26) {
            qc = 3;
        }
        if (qc != 0 && I >= surface.u - 134 && xb >= 3 && I < surface.u - 102 && xb < 26) {
            qc = 4;
        }
        if (qc != 0 && I >= surface.u - 167 && xb >= 3 && I < surface.u - 135 && xb < 26) {
            qc = 5;
        }
        if (qc != 0 && I >= surface.u - 200 && xb >= 3 && I < surface.u - 168 && xb < 26) {
            qc = 6;
        }
        // --- leave a sub-tab when the cursor drops out of its body region (→ qc=0) ---
        // Inventory grid (qc==1): below the item rows.
        if (qc == 1 && (I < surface.u - 248 || xb > 36 + 34 * (inventorySize / 5))) {
            qc = 0;
        }
        // Stats body (qc==3): obf condition ~qc==-4 → qc==3.
        if (qc == 3 && (I < surface.u - 199 || xb > 316)) {
            qc = 0;
        }
        // Quest / friends / ignore bodies (qc==2 || qc==4 || qc==5).
        if ((qc == 2 || qc == 4 || qc == 5) && (I < surface.u - 199 || xb > 240)) {
            qc = 0;
        }
        // Settings body (qc==6).
        if (qc == 6 && (I < surface.u - 199 || xb > 311)) {
            qc = 0;
        }
    }

// ---------------------------------------------------------------------------

    /**
     * Game settings / privacy "Configuration" tab.
     * Renders the privacy toggles (private chat / trade / members) and the option
     * toggles (auto-retaliate / mouse buttons / camera / members-only), plus the
     * logout link(s). When processClicks is set, hit-tests the rows and, on a LEFT
     * click (Cf==1 for every row), flips the relevant flag and emits the matching
     * packet (opcode 111 GAME_SETTINGS_CHANGED for the privacy rows, opcode 64
     * PRIVACY_SETTINGS_CHANGED for the option rows).
     * obf: void b(int,boolean)
     */
    private void drawGameSettings(int param, boolean processClicks) {
        int panelX = -199 + surface.u;          // obf: var3 = -199 + li.u  (right panel left edge)
        surface.b(-1, tg + 6, 3, panelX - 49);  // header border line
        int panelY = 36;                        // obf: var4
        int panelW = 196;                       // obf: var5

        // Section background fills (colour from ISAAC scratch helper o.a).
        surface.c(160, panelX, 65, param ^ 15, 36, panelW, ISAAC.a(181, param + 9555, 181, 181));
        surface.c(160, panelX, 65, 0, 101, panelW, ISAAC.a(201, param ^ 9581, 201, 201));
        surface.c(160, panelX, 95, 0, 166, panelW, ISAAC.a(181, 9570, 181, 181));
        surface.c(160, panelX, isMembersWorld ? 55 : 40, 0, 261, panelW, ISAAC.a(201, 9570, 201, 201));

        int textX = panelX + 3;     // obf: var6
        int textY = panelY + 15;    // obf: var18

        // "Private chat:" header
        surface.a(STRINGS[138], textX, textY, 0, false, 1);
        textY += 15;
        // chat-private on/off
        surface.a(privacyChatOn ? STRINGS[151] : STRINGS[136], textX, textY, 0xFFFFFF, false, 1);

        textY += 15;
        // trade/duel-private on/off
        surface.a(privacyTradeOn ? STRINGS[144] : STRINGS[146], textX, textY, 0xFFFFFF, false, 1);

        textY += 15;
        // members-private on/off (members accounts only)
        if (isMembersAccount) {
            surface.a(privacyMembersOn ? STRINGS[155] : STRINGS[141], textX, textY, 0xFFFFFF, false, 1);
        }

        // static labels (obf: var18 += param, where param is normally 0)
        textY += param;
        surface.a(STRINGS[145], textX, textY, 0xFFFFFF, false, 0);
        textY += 15;
        surface.a(STRINGS[143], textX, textY, 0xFFFFFF, false, 0);
        textY += 15;
        surface.a(STRINGS[130], textX, textY, 0xFFFFFF, false, 0);
        textY += 15;

        // combat-mode 3-state display (obf: Yd): 0→[135], 1→[137], else→[132]
        if (combatModeSetting == 0) {
            surface.a(STRINGS[135], textX, textY, 0xFFFFFF, false, 0);
        } else if (combatModeSetting == 1) {
            surface.a(STRINGS[137], textX, textY, 0xFFFFFF, false, 0);
        } else {
            surface.a(STRINGS[132], textX, textY, 0xFFFFFF, false, 0);
        }

        textY += 15;
        textY += 5;
        // "Sound effects:" / "Camera:" headers
        surface.a(STRINGS[139], panelX + 3, textY, 0, false, 1);
        textY += 15;
        surface.a(STRINGS[133], panelX + 3, textY, 0, false, 1);
        textY += 15;

        // auto-retaliate on/off (obf: De): De!=0 → [153] "On", else [131] "Off"
        surface.a(autoRetaliateOn != 0 ? STRINGS[153] : STRINGS[131], panelX + 3, textY, 0xFFFFFF, false, 1);

        textY += 15;
        // mouse buttons one/two (obf: dc): 0 → [142], else [150]
        surface.a(mouseButtonsOne == 0 ? STRINGS[142] : STRINGS[150], panelX + 3, textY, 0xFFFFFF, false, 1);

        textY += 15;
        // camera auto/manual (obf: Vg): 0 → [152], else [140]
        surface.a(cameraModeAuto == 0 ? STRINGS[140] : STRINGS[152], panelX + 3, textY, 0xFFFFFF, false, 1);

        textY += 15;
        // members-only option (obf: ui), members accounts only: !=0 → [154], else [129]
        if (isMembersAccount) {
            surface.a(membersOption != 0 ? STRINGS[154] : STRINGS[129], panelX + 3, textY, 0xFFFFFF, false, 1);
        }

        textY += 15;
        // members-world logout shortcut row (highlight on hover)
        if (isMembersWorld) {
            int color = 0xFFFFFF;
            textY += 5;
            if (I > textX && I < textX + panelW && xb > textY - 12 && xb < textY + 4) {
                color = 0xFFFF00;
            }
            surface.a(STRINGS[134], textX, textY, color, false, 1);
            textY += 15;
        }

        textY += 5;
        // "Logout" label
        surface.a(STRINGS[147], textX, textY, 0, false, 1);
        int logoutColor = 0xFFFFFF;
        textY += 15;
        if (I > textX && I < textX + panelW && xb > textY - 12 && xb < textY + 4) {
            logoutColor = 0xFFFF00;
        }
        surface.a(STRINGS[149], panelX + 3, textY, logoutColor, false, 1);

        if (!processClicks) {
            return;
        }

        // --- click handling ---
        // relX/relY measured from the panel's top-left; gate to the panel rectangle.
        // obf: var3 = 199 - li.u + I ; var15 = -36 + xb
        int relX = 199 - surface.u + I;
        int relY = xb - 36;
        if (relX < 0 || relY < 0 || relX >= 196 || relY >= 265) {
            return;
        }

        int px = surface.u - 199 + 3;   // obf: var6 (re-derived for hit-tests)
        int pw = 196;                   // obf: var5
        int rowY = 66;                  // obf: var18 = 30 + 36

        // Private chat toggle (LEFT click)
        if (I > px && I < px + pw && xb > rowY - 12 && xb < rowY + 4 && mouseClickButton == 1) {
            privacyChatOn = !privacyChatOn;
            clientStream.b(111, 0);              // opcode 111 GAME_SETTINGS_CHANGED
            clientStream.f.c(0, 41);             // setting id 0 = private chat
            clientStream.f.c(privacyChatOn ? 1 : 0, -107);
            clientStream.b(21294);               // flush
        }
        rowY += 15;
        // Trade/duel privacy toggle (LEFT click)
        if (I > px && I < px + pw && xb > rowY - 12 && xb < rowY + 4 && mouseClickButton == 1) {
            privacyTradeOn = !privacyTradeOn;
            clientStream.b(111, param - 15);
            clientStream.f.c(2, param ^ 85);     // setting id 2 = trade
            clientStream.f.c(privacyTradeOn ? 1 : 0, -82);
            clientStream.b(param ^ 21281);
        }
        rowY += 15;
        // Members privacy toggle (members account, LEFT click)
        if (isMembersAccount && I > px && I < px + pw && xb > rowY - 12 && xb < rowY + 4 && mouseClickButton == 1) {
            privacyMembersOn = !privacyMembersOn;
            clientStream.b(111, 0);
            clientStream.f.c(3, param - 136);    // setting id 3 = members
            clientStream.f.c(privacyMembersOn ? 1 : 0, -42);
            clientStream.b(21294);
        }
        // five blank rows before the option toggles
        rowY += 15;
        rowY += 15;
        rowY += 15;
        rowY += 15;
        rowY += 15;

        boolean optionsChanged = false;
        rowY += 35;
        // Auto-retaliate toggle (LEFT click)
        if (I > px && I < px + pw && xb > rowY - 12 && xb < rowY + 4 && mouseClickButton == 1) {
            autoRetaliateOn = 1 - autoRetaliateOn;
            optionsChanged = true;
        }
        rowY += 15;
        // Mouse-buttons toggle (LEFT click)
        if (I > px && I < px + pw && xb > rowY - 12 && xb < rowY + 4 && mouseClickButton == 1) {
            mouseButtonsOne = 1 - mouseButtonsOne;
            optionsChanged = true;
        }
        rowY += 15;
        // Camera toggle (LEFT click)
        if (I > px && I < px + pw && xb > rowY - 12 && xb < rowY + 4 && mouseClickButton == 1) {
            cameraModeAuto = 1 - cameraModeAuto;
            optionsChanged = true;
        }
        rowY += 15;
        // Members-only option toggle (members account, LEFT click)
        if (isMembersAccount && I > px && I < px + pw && xb > rowY - 12 && xb < rowY + 4 && mouseClickButton == 1) {
            optionsChanged = true;
            membersOption = 1 - membersOption;
        }
        rowY += 15;
        if (optionsChanged) {
            // opcode 64 PRIVACY_SETTINGS_CHANGED: (Vg camera, dc mouse, De retaliate, ui members)
            sendPrivacySettings(cameraModeAuto, mouseButtonsOne, autoRetaliateOn, param + 64, membersOption);
        }

        // Members-world quick-logout shortcut
        if (isMembersWorld) {
            rowY += 5;
            if (I > px && I < px + pw && xb > rowY - 12 && xb < rowY + 4 && mouseClickButton == 1) {
                // obf: a(o.g, param-3, 9, false) — o.g is a String[] menu-option list,
                // so this is drawMenuOptions(String[],int,int,boolean), NOT an inventory click.
                drawMenuOptions(ISAAC.g, param - 3, 9, false);
                qc = 0;
            }
            rowY += 15;
        }

        // Logout link
        rowY += 20;
        if (I > px && I < px + pw && xb > rowY - 12 && xb < rowY + 4 && mouseClickButton == 1) {
            requestLogout(0);   // obf: B(0) — opcode 102 LOGOUT
        }
        mouseClickButton = 0;
    }

// ---------------------------------------------------------------------------

    /**
     * Toggle to two-button mouse mode (eg=77). Called with a sentinel param so the
     * anti-tamper guard (if param != -16433) gates the assignment.
     * obf: void g(int)
     */
    private void setMouseButtonMode(int param) {
        if (param != -16433) {
            mouseButtonMode = 77;   // obf: eg
        }
    }

// ---------------------------------------------------------------------------

    /**
     * Reset the trade/duel/panel transient state. The secondary inventory-tab reset
     * only fires for the sentinel param (-2).
     * obf: void o(int)
     */
    private void resetTradeDuelState(int param) {
        tradeQueuedAction = 0;    // obf: bj
        activePanel = 0;          // obf: Xd
        chatInputMode = 0;        // obf: qg
        if (param == -2) {
            inventoryTab = 0;     // obf: kc
        }
    }

// ---------------------------------------------------------------------------

    /**
     * Character-design arrow/colour-wheel control row renderer (hover highlight only).
     * Three categories per pass (body-part sprite + two colour wheels), drawn at three
     * column anchors (x-87, x-32, x+23 from the centred base x=256).
     * obf: void w(int)
     */
    private void drawCharDesignControls(int param) {
        surface.i = false;
        surface.a(true);
        panelQuest.a((byte)-13);    // obf: Af.a — update panel hover state

        int baseX = 140 + 116;   // = 256 (centred column)
        int baseY = 50 - 25;     // = 25

        // Per-category indices into the colour palettes (ei/Wh/Dg) and the equip-sprite
        // slot table (WorldEntity.g): Lh,Wg = skin/colour ; hh = colour ; ld = colour ;
        // wg,dk,Vd = sprite-slot bases. (Obf field names kept — semantic split is approximate.)
        // Column 1 (baseX-87)
        surface.a(baseX - 87, ei[Lh], WorldEntity.g[wg], baseY, 102, (byte)105, 64);
        surface.a(baseY, ei[Wg], Wh[hh], false, 0, WorldEntity.g[dk], 102, 64, baseX - 87, 1);
        surface.a(baseY, Dg[ld], Wh[hh], false, 0, WorldEntity.g[Vd], 102, 64, baseX - 87, param + 13760);
        // Column 2 (baseX-32)
        surface.a(baseX - 32, ei[Lh], 6 + WorldEntity.g[wg], baseY, 102, (byte)105, 64);
        surface.a(baseY, ei[Wg], Wh[hh], false, 0, WorldEntity.g[dk] + 6, 102, 64, baseX - 32, 1);
        surface.a(baseY, Dg[ld], Wh[hh], false, 0, 6 + WorldEntity.g[Vd], 102, 64, baseX - 32, 1);
        // Column 3 (baseX+23)
        surface.a(baseX - 32 + 55, ei[Lh], 12 + WorldEntity.g[wg], baseY, 102, (byte)110, 64);
        surface.a(baseY, ei[Wg], Wh[hh], false, 0, WorldEntity.g[dk] + 12, 102, 64, baseX - 32 + 55, param + 13760);
        surface.a(baseY, Dg[ld], Wh[hh], false, 0, WorldEntity.g[Vd] + 12, 102, 64, baseX - 32 + 55, 1);

        surface.b(-1, tg + 22, inventoryPanelH, 0);   // bottom border (obf: Oi)
        surface.a(graphics, Eb, 256, K);
        if (param != -13759) {
            drawCloseButton((byte)70);   // obf: l(70)
        }
    }

// ---------------------------------------------------------------------------

    /**
     * Build the "Please design your Character" appearance screen widgets into the
     * char-design panel (Af). Lays out the head/hair toggles, gender swatches and the
     * five colour-wheel arrow pairs (top/bottom/hair/skin), then the Accept button.
     * The recorded button ids (Dj/pi/Kj/ed/Ge/Of/Xc/ek/Ze/Mj/Re/Ai/Eg) drive
     * drawCharDesignControls + sendAppearance.
     * obf: void t(int)
     */
    private void drawCharDesign(int param) {
        panelQuest = new Panel(surface, 100);   // obf: Af = new qa(li, 100)
        panelQuest.a(true, (byte)-125, 4, 256, STRINGS[87], 10);   // title

        int x = 140 + 116;   // = 256
        int y = 34 - 10;     // = 24

        // Head-style toggle row (left / current / right)
        panelQuest.a(true, (byte)-104, 3, x - 55, STRINGS[82], y + 110);
        panelQuest.a(true, (byte)-91, 3, x, STRINGS[92], y + 110);
        panelQuest.a(true, (byte)-117, 3, x + 55, STRINGS[81], y + 110);
        y += 145;

        int s = 54;
        // Gender swatch (left)
        panelQuest.a(41, x - s, 53, 26531, y);
        panelQuest.a(true, (byte)-81, 1, x - s, STRINGS[84], y - 8);
        panelQuest.a(true, (byte)-125, 1, x - s, STRINGS[88], y + 8);
        panelQuest.c(StringCodec.g - -7, y, x - s - 40, -114);
        Dj = panelQuest.d(x - s - 40, 20, y, param + 24525, 20);
        panelQuest.c(6 + StringCodec.g, y, x - s + 40, -59);
        pi = panelQuest.d(x - s + 40, 20, y, param ^ 24649, 20);
        // Gender swatch (right)
        panelQuest.a(41, x - -s, 53, 26531, y);
        panelQuest.a(true, (byte)-85, 1, x - -s, STRINGS[85], y - 8);
        panelQuest.a(true, (byte)-102, 1, s + x, STRINGS[86], y + 8);
        panelQuest.c(7 + StringCodec.g, y, s + (x - 40), -57);
        Kj = panelQuest.d(x - -s - 40, 20, y, 64, 20);
        panelQuest.c(6 + StringCodec.g, y, 40 + s + x, -127);
        ed = panelQuest.d(40 + s + x, 20, y, param ^ -24650, 20);
        y += 50;

        // Hair colour (left swatch) + Top colour pair (right swatch)
        panelQuest.a(41, x - s, 53, 26531, y);
        panelQuest.a(true, (byte)-102, 1, x - s, STRINGS[91], y);
        panelQuest.c(StringCodec.g - -7, y, -40 + x - s, param + 24525);
        Ge = panelQuest.d(x - s - 40, 20, y, -81, 20);
        panelQuest.c(StringCodec.g - -6, y, 40 - s + x, param + 24521);
        Of = panelQuest.d(40 - s + x, 20, y, 54, 20);
        panelQuest.a(41, s + x, 53, param ^ -1970, y);
        panelQuest.a(true, (byte)-102, 1, s + x, STRINGS[79], y - 8);
        panelQuest.a(true, (byte)-79, 1, s + x, STRINGS[86], y + 8);
        panelQuest.c(7 + StringCodec.g, y, s + x - 40, -104);
        Xc = panelQuest.d(s + x - 40, 20, y, param + 24504, 20);
        panelQuest.c(6 + StringCodec.g, y, 40 + s + x, -105);
        ek = panelQuest.d(x - (-s - 40), 20, y, -91, 20);
        y += 50;
        if (param != -24595) {
            drawProgressBar(-127);   // obf: y(-127)
        }

        // Top colour pair (left swatch) + Bottom colour pair (right swatch)
        panelQuest.a(41, x - s, 53, param ^ -1970, y);
        panelQuest.a(true, (byte)-81, 1, x - s, STRINGS[83], y - 8);
        panelQuest.a(true, (byte)-109, 1, x - s, STRINGS[86], y + 8);
        panelQuest.c(7 + StringCodec.g, y, -40 + x - s, -59);
        Ze = panelQuest.d(-40 + x - s, 20, y, param + 24468, 20);
        panelQuest.c(StringCodec.g + 6, y, x - s - -40, -95);
        Mj = panelQuest.d(x - s + 40, 20, y, param + 24637, 20);
        panelQuest.a(41, s + x, 53, 26531, y);
        panelQuest.a(true, (byte)-108, 1, s + x, STRINGS[89], y - 8);
        panelQuest.a(true, (byte)-108, 1, s + x, STRINGS[86], y + 8);
        panelQuest.c(StringCodec.g + 7, y, -40 + s + x, -90);
        Re = panelQuest.d(s + x - 40, 20, y, 69, 20);
        panelQuest.c(6 + StringCodec.g, y, x + s + 40, param + 24537);
        Ai = panelQuest.d(40 + s + x, 20, y, -119, 20);
        y += 82;

        // Accept button
        y -= 35;
        panelQuest.c(param ^ 24661, 200, 30, x, y);
        panelQuest.a(false, (byte)-74, 4, x, STRINGS[90], y);
        Eg = panelQuest.d(x, 200, y, param ^ -24631, 30);   // Accept → opcode 235
    }

// ---------------------------------------------------------------------------

    /**
     * Overlay a centred two-line message box directly onto the AWT Graphics (used for
     * transient "saving/loading" notices that must paint outside the normal surface
     * blit). Header text below, body text above, inside a black 280x50 box.
     * The colorCode sentinel (-64) is an anti-tamper guard.
     * obf: void a(String,byte,String)
     */
    private void addChatMessage(String header, byte colorCode, String body) {
        if (colorCode != -64) {
            return;
        }
        Graphics g = this.getGraphics();
        if (g == null) {
            return;
        }
        g.translate(Eb, K);
        Font font = new Font(STRINGS[477], 1, 15);   // STRINGS[477] = "Helvetica"
        int w = 512;
        g.setColor(Color.black);
        int h = 344;
        g.fillRect(w / 2 - 140, h / 2 - 25, 280, 50);
        g.setColor(Color.white);
        g.drawRect(w / 2 - 140, h / 2 - 25, 280, 50);
        this.a(font, body, h / 2 - 10, true, w / 2, g);
        this.a(font, header, h / 2 + 10, true, w / 2, g);
    }

// ---------------------------------------------------------------------------

    /**
     * Display a server/system message on the duel/stats panel (yi). When the body is
     * empty it goes to a single slot (td); otherwise the title goes to slot Qi and the
     * body to slot td. A trigger code < -11 forces a minimap redraw + action flush.
     * obf: void b(byte,String,String)
     */
    private void showServerMessage(byte triggerCode, String title, String body) {
        if (activePanel == 2) {
            if (body == null || body.length() < 1) {
                panelDuel.a(serverMsgControlId, title, 27642);   // obf: yi.a(td, ...)
            } else {
                panelDuel.a(fatigueControlId, title, 27642);     // obf: yi.a(Qi, ...)
                panelDuel.a(serverMsgControlId, body, 27642);    // obf: yi.a(td, ...)
            }
        }
        if (triggerCode < -11) {
            drawMinimap(2540);            // obf: k(2540)
            sendQueuedActions(-28492);    // obf: c(-28492)
        }
    }

// ---------------------------------------------------------------------------

    /**
     * Render the right-click "Choose option" list. Clears the screen first unless
     * called with sentinel 12, then delegates to drawScrollList with a 9px left margin.
     * obf: void a(String[],int,int,boolean)
     */
    private void drawMenuOptions(String[] options, int x, int y, boolean rightClick) {
        if (x != 12) {
            // obf: this.e((byte)31) — NOT clearScreen (q(byte)@17480); e(byte)@12828 resets
            // entity counts + panel id + username/password buffers. Named per behaviour.
            resetMenuState((byte)31);
        }
        drawScrollList(x - 9, y, options, rightClick, "");
    }

// ---------------------------------------------------------------------------

    /**
     * Configure the generic scrollable list/menu: stash the option strings, compute the
     * widest row (min 400) and the total height from the font metrics, and reset the
     * list state. Clears the bound message list unless x==3.
     * obf: void a(int,int,String[],boolean,String)
     */
    private void drawScrollList(int x, int y, String[] options, boolean showBorder, String title) {
        menuOptionList = options;       // obf: od
        menuWidth = 400;                // obf: zi
        if (x != 3) {
            scrollMessageList = null;   // obf: Wf
        }
        for (int i = 0; i < options.length; i++) {
            int w = surface.a(1, x + 113, options[i]) + 10;
            if (menuWidth < w) {
                menuWidth = w;
            }
        }
        menuHeight = 15 + (surface.a(508305352, 1) + 2) * (1 + options.length) + surface.a(508305352, 4);   // obf: gl
        menuX = y;            // obf: gc
        menuTitle = title;    // obf: e
        menuOpenFlag = false; // obf: vk
        inputLine = "";       // obf: Cb
        showMenuBorder = showBorder;   // obf: Bd
    }

// ---------------------------------------------------------------------------

    /**
     * Scrollbar widget (primary variant). Hit-tests via the walkTo dispatch helper; if
     * the first probe misses, runs the second probe and (unless sentinel==10) draws the
     * secondary slider.
     * obf: void a(byte,int,int,int,boolean,int)
     */
    private void drawScrollbar(byte sentinel, int x, int y, int scrollPos, boolean animate, int trackLen) {
        if (!walkTo(x, trackLen, (byte)14, false, scrollPos, scrollPos, y, y, animate)) {
            walkTo(scrollPos, animate, trackLen, y, x, scrollPos, true, y, sentinel + 107);
            if (sentinel != 10) {
                drawScrollbar2(99, 113, -126, -87, true, 125);
            }
        }
    }

// ---------------------------------------------------------------------------

    /**
     * Secondary scrollbar/slider widget. Single walkTo probe; resets the activity timer
     * unless trackLen==8.
     * obf: void a(int,int,int,int,boolean,int)
     */
    private void drawScrollbar2(int x, int y, int w, int h, boolean animate, int trackLen) {
        // obf: this.a(var2, var5, var4, var1, var3, var2, false, var1, 105)
        //   = walkTo(y, animate, h, x, w, y, false, x, 105)   [3rd arg is h (var4), not trackLen]
        walkTo(y, animate, h, x, w, y, false, x, 105);
        if (trackLen != 8) {
            lastActionTime = -85L;   // obf: Wi
        }
    }

// ---------------------------------------------------------------------------

    /**
     * Rebuild the inventory item draw-cache from the previous frame's draw list and the
     * current inventory, de-duplicating item ids and appending new ones up to the cache
     * capacity (Gi). (The trailing junk divide in the clean output is an opaque-predicate
     * artifact and is dropped.)
     * obf: void C(int)
     */
    private void drawHelpMenu(int param) {
        drawListCount = drawListSize;   // obf: vj = fj
        for (int i = 0; i < drawListSize; i++) {
            drawListIds[i] = drawListCurrent[i];   // obf: ae[i] = ci[i]
            drawListY[i] = drawListYShadow[i];     // obf: di[i] = Xe[i]
        }

        for (int n = 0; n < inventoryCount; n++) {
            if (drawListCapacity <= drawListCount) {   // obf: Gi <= vj — cache full
                break;
            }
            int itemId = inventoryItems[n];   // obf: vf[n]
            boolean found = false;
            for (int j = 0; j < drawListCount; j++) {
                if (drawListIds[j] == itemId) {
                    found = true;
                    break;
                }
            }
            if (!found) {
                drawListIds[drawListCount] = itemId;
                drawListY[drawListCount] = 0;
                drawListCount++;
            }
        }
    }

// ---------------------------------------------------------------------------

    /**
     * Draw the "Click here to close window" modal button (and the help panel frame).
     * Highlights the close text on hover; on a LEFT click (Cf==1) over the text, or a
     * LEFT click fully outside the modal box, closes it (mh=false).
     * obf: void l(byte)
     */
    private void drawCloseButton(byte param) {
        int w = 400;
        if (param != -115) {
            closeButtonSpriteId = 64;   // obf: qd
        }
        int h = 100;
        if (helpMenuOpen) {   // obf: Wk
            h = 300;          // (clean assigns 450 then 300; first is dead)
        }
        surface.a(256 - w / 2, (byte)122, 0, 167 - h / 2, h, w);
        surface.e(256 - w / 2, w, 167 - h / 2, 27785, h, 0xFFFFFF);
        surface.a(w - 40, helpMenuTitle, 256, 92, 1, 167 - (h / 2 - 20), true, 0xFFFFFF);

        int labelY = 157 + h / 2;
        int labelColor = 0xFFFFFF;
        // obf: ~xb < ~(var4-12) && ~xb >= ~var4 && ~I < -107 && -407 < ~I
        if (xb > labelY - 12 && xb <= labelY && I > 106 && I < 406) {
            labelColor = 0xFF0000;
        }
        surface.a(256, STRINGS[126], labelColor, 0, 1, labelY);   // "Click here to close window"

        if (mouseClickButton == 1) {   // obf: ~Cf == -2 → Cf == 1 (LEFT click)
            if (labelColor == 0xFF0000) {
                showCloseWindow = false;   // obf: mh
            }
            // close when the click lands fully outside the modal box (X outside AND Y outside)
            if ((I < 256 - w / 2 || I > 256 + w / 2) && (xb < 167 - h / 2 || xb > 167 + h / 2)) {
                showCloseWindow = false;
            }
        }
        mouseClickButton = 0;   // obf: Cf = 0
    }

// ---------------------------------------------------------------------------

    /**
     * Substitute a named wall/boundary GameModel into the scene at wall-slot index
     * slotIndex, provided the slot's tile (Se/ye) is in-bounds and within 7 tiles of the
     * local player. Used to swap in special boundary geometry (e.g. doors). The sentinel
     * guard (var1 > 2) is anti-tamper.
     * obf: void a(byte,int,String)
     */
    private void drawTextField(byte sentinel, int slotIndex, String text) {
        int wallZ = wallModelZ[slotIndex];   // obf: Se[n2]
        int wallX = wallModelX[slotIndex];   // obf: ye[n2]
        int relX = wallZ - localPlayer.i / 128;
        int relY = -(localPlayer.K / 128) + wallX;
        int range = 7;

        if (sentinel <= 2) return;
        if (wallZ < 0) return;
        if (wallX < 0) return;
        if (wallZ > 95) return;
        if (wallX > 95) return;
        if (relX <= -range) return;
        if (relX >= range) return;
        if (relY <= -range) return;
        if (relY >= range) return;

        world.a(wallModels[slotIndex], -1);             // remove existing wall model (obf: Ek.a)
        int modelIdx = GameModel.a((byte)91, text);     // resolve model by name
        GameModel newModel = objectModels[modelIdx].b(-2);   // clone base model
        world.a(newModel, (byte)118);                   // register into scene
        newModel.a(-50, 48, -10, -50, true, 48, -74);   // transform/scale
        newModel.a(wallModels[slotIndex], 6029);        // copy placement from old model
        newModel.rb = slotIndex;
        wallModels[slotIndex] = newModel;
    }

// ---------------------------------------------------------------------------

    /**
     * Add-friend / send-PM / add-ignore entry dialog renderer + submit handler.
     * socialDialogMode (Bj): 1=Add friend, 2=Send PM, 3=Add ignore. A click outside the
     * active dialog box (or on the Cancel link) dismisses it. Each mode renders its box
     * and, once the input line is committed, normalises the name (WorldEntity.a) and
     * fires the matching social packet — but never when the target equals the local
     * player's own name.
     * obf: void h(byte)
     */
    private void drawSocialDialog(byte sentinel) {
        if (mouseClickButton != 0) {
            mouseClickButton = 0;
            // Add-friend box: 106..406 x, 145..215 y
            if (socialDialogMode == 1 && (I < 106 || xb < 145 || I > 406 || xb > 215)) {
                socialDialogMode = 0;
                return;
            }
            // Send-PM box: 6..506 x, 145..215 y
            if (socialDialogMode == 2 && (I < 6 || xb < 145 || I > 506 || xb > 215)) {
                socialDialogMode = 0;
                return;
            }
            // Add-ignore box: 106..406 x, 145..215 y
            if (socialDialogMode == 3 && (I < 106 || xb < 145 || I > 406 || xb > 215)) {
                socialDialogMode = 0;
                return;
            }
            // Cancel link region
            if (I > 236 && I < 276 && xb > 193 && xb < 213) {
                socialDialogMode = 0;
                return;
            }
        }

        int y = 145;

        // Mode 1: Add friend
        if (socialDialogMode == 1) {
            surface.a(106, (byte)26, 0, y, 70, 300);
            surface.e(106, 300, y, 27785, 70, 0xFFFFFF);
            y += 20;
            surface.a(256, STRINGS[246], 0xFFFFFF, 0, 4, y);   // "Enter name to add to friends list"
            y += 20;
            surface.a(256, tempInputString + "*", 0xFFFFFF, 0, 4, y);
            String self = WorldEntity.a(localPlayer.C, (byte)50);   // normalise local name
            if (self != null && inputLine.length() > 0) {
                String typed = inputLine.trim();
                socialDialogMode = 0;
                tempInputString = "";
                inputLine = "";
                if (typed.length() > 0 && !self.equals(WorldEntity.a(typed, (byte)100))) {
                    sendAddFriend(114, typed);   // obf: b(114, typed) — opcode 195
                }
            }
        }

        // Mode 2: Send private message
        if (socialDialogMode == 2) {
            surface.a(6, (byte)110, 0, y, 70, 500);
            surface.e(6, 500, y, 27785, 70, 0xFFFFFF);
            y += 20;
            surface.a(256, STRINGS[249] + pmTarget, 0xFFFFFF, 0, 4, y);   // "Sending message to "
            y += 20;
            surface.a(256, pmInput + "*", 0xFFFFFF, 0, 4, y);
            if (submittedPmInput.length() > 0) {
                String msg = submittedPmInput;
                pmInput = "";
                socialDialogMode = 0;
                submittedPmInput = "";
                sendPrivateMessage((byte)-76, pmTarget, msg);   // obf: a((byte)-76, ...) opcode 218
            }
        }

        // Mode 3: Add ignore
        if (socialDialogMode == 3) {
            surface.a(106, (byte)-115, 0, y, 70, 300);
            surface.e(106, 300, y, 27785, 70, 0xFFFFFF);
            y += 20;
            surface.a(256, STRINGS[248], 0xFFFFFF, 0, 4, y);   // "Enter name to add to ignore list"
            y += 20;
            surface.a(256, tempInputString + "*", 0xFFFFFF, 0, 4, y);
            String self = WorldEntity.a(localPlayer.C, (byte)59);
            if (self != null && inputLine.length() > 0) {
                String typed = inputLine.trim();
                tempInputString = "";
                socialDialogMode = 0;
                inputLine = "";
                if (typed.length() > 0 && !self.equals(WorldEntity.a(typed, (byte)105))) {
                    sendAddIgnore(typed, (byte)5);   // obf: a(typed, (byte)5) — opcode 132
                }
            }
        }

        // Cancel link (always drawn)
        int color = 0xFFFFFF;
        if (I > 236 && I < 276 && xb > 193 && xb < 213) {
            color = 0xFFFF00;
        }
        surface.a(256, STRINGS[121], color, 0, 1, 208);   // STRINGS[121] = "Cancel"

        if (sentinel <= 77) {
            closeButtonPressed = -42;   // obf: pj
        }
    }

// ---------------------------------------------------------------------------

    /**
     * "Enter the name of the player you wish to report" entry screen (report-abuse step 1).
     * If a name has been typed, accept it and advance to the offence picker (Vf=2).
     * Otherwise render the entry box (plus an optional mute toggle), the Submit link
     * (commits tempInputString into the input line) and the Cancel link. A LEFT click
     * outside the box cancels.
     *
     * Mute tier (var2): (Ce>=2 || Oj>=7) → 2 ; else Oj<5 → 0 ; else (Oj 5/6) → 1.
     * obf: void d(boolean)
     */
    private void drawReportNameEntry(boolean rightAlign) {
        if (inputLine.length() > 0) {
            reportAbuseTarget = inputLine.trim();   // obf: ec
            reportAbusePage = 0;                    // obf: Yb
            inputMode = 2;                          // obf: Vf
            return;
        }

        // mute-option tier
        int muteTier;
        if (reportAbuseMuteFlag >= 2 || reportAbuseOffence >= 7) {
            muteTier = 2;
        } else if (reportAbuseOffence < 5) {        // obf: ~Oj > -6
            muteTier = 0;
        } else {
            muteTier = 1;
        }

        int fontH = surface.a(508305352, 1);
        int lineH = surface.a(508305352, 4);
        int panelW = 400;
        int panelH = (muteTier > 0 ? 5 + fontH : 0) + 70;
        int panelX = 256 - panelW / 2;
        int panelY = 180 - panelH / 2;

        surface.a(panelX, (byte)88, 0, panelY, panelH, panelW);
        surface.e(panelX, panelW, panelY, 27785, panelH, 0xFFFFFF);
        surface.a(256, STRINGS[340], 0xFFFF00, 0, 1, 5 + panelY + fontH);   // title prompt

        int inputPad = fontH + 2;   // obf: var9
        surface.a(256, tempInputString + "*", 0xFFFFFF, 0, 4, lineH + (panelY + 5) + (inputPad + 3));

        int nextY = fontH + lineH + (8 + panelY + inputPad + 2);   // obf: var10
        int color = 0xFFFFFF;

        // mute toggle row (only when tier > 0)
        if (muteTier > 0) {
            String muteLabel = reportAbuseMuteConfirmed ? STRINGS[336] : STRINGS[339];
            if (muteTier > 1) {
                muteLabel = muteLabel + STRINGS[341];
            }
            muteLabel = muteLabel + STRINGS[337];

            int muteW = surface.a(1, 72, muteLabel);
            if (I > 256 - muteW / 2 && I < 256 + muteW / 2 && xb > nextY - fontH && xb < nextY) {
                if (mouseClickButton != 0) {
                    reportAbuseMuteConfirmed = !reportAbuseMuteConfirmed;
                    mouseClickButton = 0;
                }
                color = 0xFFFF00;
            }
            surface.a(256, muteLabel, color, 0, 1, nextY);
            nextY += 10 + fontH;
        }

        // Submit link (obf: I > 210 && I < 228) — commits the typed name into the input line
        color = 0xFFFFFF;
        if (I > 210 && I < 228 && xb > nextY - fontH && xb < nextY) {
            if (mouseClickButton != 0) {
                inputLine = tempInputString;   // obf: Cb = e
                mouseClickButton = 0;
            }
            color = 0xFFFF00;
        }
        surface.a(STRINGS[122], 210, nextY, color, rightAlign, 1);   // "Submit"

        // Cancel link (obf: I > 264 && I < 304)
        color = 0xFFFFFF;
        if (I > 264 && I < 304 && xb > nextY - fontH && xb < nextY) {
            color = 0xFFFF00;
            if (mouseClickButton != 0) {
                mouseClickButton = 0;
                inputMode = 0;   // obf: Vf = 0
            }
        }
        surface.a(STRINGS[121], 264, nextY, color, rightAlign, 1);   // "Cancel"

        // LEFT click fully outside the box → cancel
        if (mouseClickButton == 1) {   // obf: ~Cf == -2 → Cf == 1
            if (I < panelX || I > panelX + panelW || xb < panelY || xb > panelY + panelH) {
                inputMode = 0;
                mouseClickButton = 0;
            }
        }
    }
