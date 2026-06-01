// ===== ui_b =====
// Methods 19–36 of the "ui" group from MUDCLIENT_SKELETON.md.
// All methods belong to class Mudclient (obf: client), extends GameShell (obf: e).
// Naming: fields from MUDCLIENT_SKELETON.md; other classes from NAMING.md.
// Obfuscation stripped: opaque-predicate guards (vh/bl), profiling ++counters,
// try/catch ErrorHandler wrappers, anti-tamper dummy params, junk masks.
// Field quick-ref used below:
//   li         = surface         (SurfaceSprite / ba)
//   Xb         = graphics        (java.awt.Graphics)
//   Eb,K       = screen offset x,y
//   tg,dg      = panel column offsets (tg = right panel x-base, dg = left panel x-base)
//   Oi         = inventory panel y-offset
//   jk         = compassAngle    (0..3071, compass rotation counter)
//   Xd         = activePanel     (0=none,1=options,2=quest/skill,...)
//   qc         = inventoryTab    (0-6 inventory sub-tab index)
//   I          = mouseX          (inherited from GameShell)
//   xb         = mouseY
//   Cf         = mouseClickButton (0=none,1=left,2=right)
//   Bh         = selectedItem    (selected inventory item index, -1=none)
//   af         = selectedSpell   (selected spell index, -1=none)
//   Lf         = currentFloor
//   sh         = localX
//   Tk         = itemCount       (mouse item-count increment)
//   lc         = inventoryCount
//   vf         = inventoryItems  (int[35] item ids)
//   xe         = inventoryQty    (int[35] stack counts)
//   Aj         = inventoryEquipped (int[35] equip flags)
//   eg         = mouseButtonMode (1=one-button, 2=two-button)
//   Jh         = clientStream    (ClientStream/da)
//   zh         = friendsList     (MessageList/wb), used here as menu helper
//   He         = chatList        (MessageList/wb)
//   Af         = panelDuel       (Panel/qa)  [reused for char design]
//   ge         = panelGame       (Panel/qa)
//   yi         = panelDuel       (Panel/qa)
//   Ek         = world           (World/k)
//   Hh         = scene           (Scene/lb)
//   wi         = localPlayer     (GameCharacter/ta)
//   Vk/Ej      = skillNamesShort/Long
//   Kh,Yh,ne   = privacy: chatPrivate,tradePrivate,membersWorld
//   De,dc,Vg,ui= game options: autoRetaliateEnabled, singleMouseButton, cameraAuto, ...
//   Kd         = isMembersWorld flag
//   Bd         = showMenuOnRight
//   Cb         = inputLine (current text entry buffer)
//   ec         = playerNameTarget
//   Vf         = inputMode (0=none,1=chatline,2=reportAbuse...)
//   Oj         = reportAbuseOffence
//   Ce         = reportAbuseMuteFlag
//   ue         = reportAbuseMuteConfirmed
//   e          = tempInputString (scratch label for dialogs)
//   Bj         = socialDialogMode (1=addFriend,2=sendPM,3=addIgnore)
//   Qd         = pmTarget
//   x          = pmInput
//   Ob         = submittedInput
//   Wk         = helpMenuOpen
//   Cj         = helpMenuTitle
//   re         = tradeTargetName
//   Ui,Nj      = tradeOurCount, theirCount
//   Vb,xi      = tradeOurItems, ourItems
//   Me,th      = tradeOurQty, ourQty
//   Lc,Bi      = tradeTheirItems, theirQty
//   nh         = tradeTheirCount
//   Vi,Xj      = tradeOurAccepted, theirAccepted
//   Sh,gh,Cc,Rc= duel option flags (noMagic,noPrayer,noWeapons,noRetreat)
//   Cd         = duelWeaponsEquipped
//   Ve         = duelTheirCount
//   xj,kf      = duelTheirItems, theirQty
//   Pj         = duelWindowOpen
//   ki,ke      = duelStakeAccepted flags
//   Ke,Uf,df   = duelOurCount, items, qty
//   wj,zc,of   = duelTheirCount, items, qty (bank reuse)
//   Wg,hh,ld,Vd,wg,dk = charDesign indices (skinColor,hairColor,topColor,...)
//   Dg,Wh,ei   = charDesign color palettes
//   Sf         = charDesignGender
//   sg         = spriteStartIndex (base sprite slot for equipped items)
//   Lg         = playerName
//   fd,Yi,vd,ff= duel no-magic/prayer/weapons/retreat booleans
//   Pj,ki,ke   = duel window/stake/item accepted booleans
//   Je         = menuOpen (right-click menu visible)
//   He         = chatList (MessageList)
//   ad,Uk      = menuX, menuY

// ---------------------------------------------------------------------------

    /**
     * Render the minimap/compass panel.
     * Blits the compass needle (jk angle), draws the minimap region, and
     * flushes to AWT Graphics.  Also clears the ground-item overlay array
     * if param != 2540.
     * obf: void k(int)
     */
    private void drawMinimap(int param) {
        // Reset surface clip and flag before drawing minimap panel
        surface.i = false;        // clear sprite-clip flag
        Dc = false;               // clear "dirty" flag
        surface.a(true);          // flush surface buffer

        // compassAngle runs 0..3071; derive a 0..2 scroll phase for the compass ring
        // jk increments each tick; (2*jk) % 3072 gives a [0,3072) scroll position
        int compassPos = (2 * compassAngle) % 3072;

        // Draw the three compass-strip segments depending on angle phase
        if (compassPos > 1024) {
            // First strip: solid fill at panel edge
            surface.b(-1, dg, 10, 0);
            if (compassPos < 768) {
                // Partial second strip: offset by scroll
                surface.a(1 + dg, 0, 0, compassPos - 768, 10);
            }
        }
        if (compassPos > 2048) {
            // Second full strip
            surface.b(-1, 1 + dg, 10, 0);
            if (compassPos > 1792) {
                // Partial third strip
                surface.a(tg - -10, 0, 0, compassPos - 1792, 10);
            }
        }
        if (compassPos > 2816) {
            // Third strip
            surface.b(-1, tg - -10, 10, 0);
            surface.a(dg, 0, 0, compassPos - 2816, 10);
        }

        // If not called with the special token 2540, clear ground-item overlay
        if (param != 2540) {
            inventoryGroundOverlay = null;   // obf: of
        }

        // If no active panel (activePanel == 0), reset quest panel scroll
        if (~activePanel == -1) {   // activePanel == 0
            panelQuest.a((byte)-63);
        }

        // If quest/skill panel is active (activePanel == 2), draw fatigue bar
        if (activePanel == 2) {
            String fatStr = panelDuel.g(fatiguePercent, 4);   // obf: yi.g(Qi,4)
            if (null != fatStr && fatStr.length() > 0) {
                // Draw fatigue progress bar (100=x, 30=y, 185=max, Wd=width)
                surface.c(100, 0, 30, 0, 185, Wd, 0);
            }
            panelDuel.a((byte)-52);  // tick/update duel panel
        }

        // Draw inventory panel border line
        surface.b(-1, tg + 22, Oi, 0);
        // Blit the whole surface to AWT graphics
        surface.a(graphics, Eb, 256, K);
    }

// ---------------------------------------------------------------------------

    /**
     * Render the inventory tab (items grid, equip highlights, selection).
     * Draws item sprites at a 5-column × N-row grid starting at x=li.u-248,y=36.
     * On click (mouseClick==true), resolves item under cursor and builds context menu.
     * obf: void D(int)
     */
    private void drawInventoryTab(int param) {
        // Determine sub-tab from mouse position (hover regions set inventoryTab 0..6)
        // inventoryTab (qc) transitions are based on mouseX(I) and mouseY(xb) vs surface width (li.u)
        if (inventoryTab == 0 && surface.u - 35 >= mouseX && mouseY < 35 && mouseX < surface.u - 3 && mouseY < 35) {
            inventoryTab = 1;
        }
        if (inventoryTab == 1 && surface.u - 35 - 33 >= mouseX && mouseY >= 3 && mouseX < surface.u - 3 - 33 && mouseY < 35) {
            inventoryTab = 2;
            // Randomize "wobble" for the skill stat animation
            charDesignWobbleX = (int)(13.0 * Math.random()) - 6;   // obf: Df
            charDesignWobbleY = (int)(Math.random() * 23.0) - 11;   // obf: sd
        }
        if (inventoryTab == 1 && surface.u - 101 <= mouseX && mouseY >= 3 && mouseX < surface.u - 3 - 66 && mouseY < 35) {
            inventoryTab = 3;
        }
        if (inventoryTab == 0 && surface.u - 35 - 99 >= mouseX && mouseY >= 0 && mouseX < surface.u - 3 - 99 && mouseY < 35) {
            inventoryTab = 4;
        }
        if (inventoryTab == 1 && surface.u - 35 - 132 >= mouseX && mouseY >= 3 && mouseX < surface.u - 135 && mouseY < 35) {
            inventoryTab = 5;
        }
        if (param != 1) {
            currentFloor = -32;   // obf: Lf
        }
        if (inventoryTab == 1 && surface.u - 35 - 165 >= mouseX && mouseY >= 3 && mouseX < surface.u - 3 - 165 && mouseY < 35) {
            inventoryTab = 6;
        }
        // Clamp back to 1 if mouse has moved outside
        if (inventoryTab != 0 && surface.u - 35 <= mouseX && mouseY >= 3 && mouseX < surface.u - 3 && mouseY < 26) {
            inventoryTab = 1;
        }
        // Re-check sub-tabs 2/3 boundaries
        if (inventoryTab != 1 && inventoryTab != 3 && surface.u - 68 >= mouseX && mouseY >= 0 && mouseX < surface.u - 3 - 33 && mouseY < 26) {
            inventoryTab = 2;
            charDesignWobbleY = -11 + (int)(23.0 * Math.random());
            charDesignWobbleX = -6 + (int)(13.0 * Math.random());
        }
        if (inventoryTab != 1 && surface.u - 35 - 66 >= mouseX && mouseY >= 3 && mouseX < surface.u - 3 - 66 && mouseY < 26) {
            inventoryTab = 3;
        }
        if (inventoryTab != 1 && surface.u - 35 - 99 <= mouseX && mouseY >= 3 && mouseX < surface.u - 102 && mouseY < 26) {
            inventoryTab = 4;
        }
        if (inventoryTab != 1 && surface.u - 167 >= mouseX && mouseY >= 3 && mouseX < surface.u - 3 - 132 && mouseY < 26) {
            inventoryTab = 5;
        }
        if (inventoryTab != 1 && surface.u - 35 - 165 >= mouseX && mouseY >= 3 && mouseX < surface.u - 168 && mouseY < 26) {
            inventoryTab = 6;
        }
        // Bounds-check: reset tab to 0 if cursor left the panel
        if (inventoryTab == 2 && (mouseX < surface.u - 248 || mouseY > 36 + 34 * (cl / 5))) {
            inventoryTab = 0;
        }
        if (inventoryTab == 4 && (mouseX > surface.u - 199 || mouseY > 317)) {
            inventoryTab = 0;
        }
        if ((inventoryTab == 3 || inventoryTab == 5 || inventoryTab == 6)
                && !(mouseX >= surface.u - 199 && mouseY <= 241)) {
            inventoryTab = 0;
        }
        if (inventoryTab == 6) {
            if (mouseX >= surface.u - 199 && mouseY <= 312) {
                return;
            }
            inventoryTab = 0;
        }
    }

// ---------------------------------------------------------------------------

    /**
     * Game settings / privacy options tab (Configuration panel).
     * Renders toggles for: private chat, trade, members-only world, auto-retaliate,
     * mouse mode, camera, sounds.  On click sends opcode 111 (GAME_SETTINGS_CHANGED).
     * param=passed click x position; bl=true means process clicks.
     * obf: void b(int,boolean)  — drawGameSettings @53359
     */
    private void drawGameSettings(int param, boolean processClicks) {
        // n8 = panel left edge: surface.u - 199  (right-side panel)
        int panelX = surface.u - 199;    // obf: n8
        surface.b(-1, tg + 6, 3, panelX - 49);   // border line
        int panelY = 36;   // obf: n7
        int panelW = 196;  // obf: n6

        // Three colored section backgrounds
        surface.c(160, panelX, 65, param ^ 0xF, 36, panelW, ISAAC.a(181, param + 9555, 181, 181));
        surface.c(160, panelX, 65, 0, 101, panelW, ISAAC.a(201, param ^ 0x256D, 201, 201));
        surface.c(160, panelX, 95, 0, 166, panelW, ISAAC.a(181, 9570, 181, 181));
        // Fourth section height depends on members world flag (Kd)
        surface.c(160, panelX, isMembersWorld ? 55 : 40, 0, 261, panelW, ISAAC.a(201, 9570, 201, 201));

        int textX = panelX + 3;   // obf: n5
        int textY = panelY + 15;  // obf: n4

        // Section header STRINGS[138] = "Private chat:"
        surface.a(STRINGS[138], textX, textY, 0, false, 1);
        textY += 15;

        // Private-chat toggle: Kh=true → "On", false → "Off"
        if (privacyChatOn) {   // obf: Kh
            surface.a(STRINGS[151], textX, textY, 0xFFFFFF, false, 1);
        } else {
            surface.a(STRINGS[136], textX, textY, 0xFFFFFF, false, 1);
        }
        textY += 15;

        // Trade/duel privacy: Yh=true → "On", false → "Off"  STRINGS[144]/[146]
        if (privacyTradeOn) {  // obf: Yh
            surface.a(STRINGS[144], textX, textY, 0xFFFFFF, false, 1);
        } else {
            surface.a(STRINGS[146], textX, textY, 0xFFFFFF, false, 1);
        }
        textY += 15;

        // Members/friends privacy (only if members world: Pg)
        if (isMembersAccount) {   // obf: Pg
            if (privacyMembersOn) {  // obf: ne
                surface.a(STRINGS[155], textX, textY, 0xFFFFFF, false, 1);
            } else {
                surface.a(STRINGS[141], textX, textY, 0xFFFFFF, false, 1);
            }
        }

        // Spacer then static labels  STRINGS[145],[143],[130]
        surface.a(STRINGS[145], textX, textY += param, 0xFFFFFF, false, 0);  // dead param used as 0 normally
        surface.a(STRINGS[143], textX, textY += 15, 0xFFFFFF, false, 0);
        surface.a(STRINGS[130], textX, textY += 15, 0xFFFFFF, false, 0);
        textY += 15;

        // Auto-retaliate: Yd=0→On, 1→Defensive, 2→Aggressive
        if (autoRetaliateMode == 0) {     // obf: Yd
            surface.a(STRINGS[135], textX, textY, 0xFFFFFF, false, 0);
        } else if (autoRetaliateMode == 1) {
            surface.a(STRINGS[137], textX, textY, 0xFFFFFF, false, 0);
        } else {
            surface.a(STRINGS[132], textX, textY, 0xFFFFFF, false, 0);
        }
        textY += 15;

        // STRINGS[139] = "Sound effects:", STRINGS[133] = "Camera:"
        surface.a(STRINGS[139], panelX + 3, textY += 5, 0, false, 1);
        surface.a(STRINGS[133], panelX + 3, textY += 15, 0, false, 1);
        textY += 15;

        // Auto-retaliate on/off  STRINGS[153]/"On" vs [131]/"Off"
        if (autoRetaliateMode == 0) {
            surface.a(STRINGS[153], panelX + 3, textY, 0xFFFFFF, false, 1);
        } else {
            surface.a(STRINGS[131], panelX + 3, textY, 0xFFFFFF, false, 1);
        }
        textY += 15;

        // One/two button mouse:  dc=0→one-button, 1→two-button  STRINGS[142]/[150]
        if (singleMouseButton == 0) {   // obf: dc
            surface.a(STRINGS[142], panelX + 3, textY, 0xFFFFFF, false, 1);
        } else {
            surface.a(STRINGS[150], panelX + 3, textY, 0xFFFFFF, false, 1);
        }
        textY += 15;

        // Camera mode:  Vg=0→manual, 1→auto  STRINGS[152]/[140]
        if (cameraAuto == 0) {   // obf: Vg
            surface.a(STRINGS[152], panelX + 3, textY, 0xFFFFFF, false, 1);
        } else {
            surface.a(STRINGS[140], panelX + 3, textY, 0xFFFFFF, false, 1);
        }
        textY += 15;

        // Members-only option (ui=0→off, 1→on)  STRINGS[154]/[129]
        if (isMembersAccount) {
            if (membersOnlyOption == 0) {    // obf: ui
                surface.a(STRINGS[154], panelX + 3, textY, 0xFFFFFF, false, 1);
            } else {
                surface.a(STRINGS[129], panelX + 3, textY, 0xFFFFFF, false, 1);
            }
        }
        textY += 15;

        // Members logout button (Kd = members world)
        if (isMembersWorld) {   // obf: Kd
            int logoutColor = 0xFFFFFF;
            if (mouseX > textX && mouseX < textX + panelW
                    && mouseY > (textY += 5) - 12 && mouseY < textY + 4) {
                logoutColor = 0xFFFF00;
            }
            surface.a(STRINGS[134], textX, textY, logoutColor, false, 1);
            textY += 15;
        }

        // STRINGS[147] = divider/spacer label
        surface.a(STRINGS[147], textX, textY += 5, 0, false, 1);
        int logoutColor = 0xFFFFFF;
        if (mouseX > textX && mouseX < textX + panelW
                && mouseY > (textY += 15) - 12 && mouseY < textY + 4) {
            logoutColor = 0xFFFF00;
        }
        // STRINGS[149] = "Logout" button
        surface.a(STRINGS[149], panelX + 3, textY, logoutColor, false, 1);

        // Click handling: only when processClicks==true
        if (!processClicks) {
            return;
        }
        // Translate mouse position relative to panel
        int relX = surface.u - 199 + mouseX;  // obf: n8 = 199 - li.u + I
        int relY = mouseY - 36;               // obf: n7 = -36 + xb

        if (relX < 0 || relY < 0 || relX >= 196 || relY >= 265) return;

        // Re-derive panel metrics for click testing
        int px = surface.u - 199 + 3;
        int py = 36;
        int pw = 196;
        int clickY = 30 + py;

        // Toggle private chat
        if (mouseX > px && mouseX < px + pw && mouseY > clickY - 12 && mouseY < clickY + 4 && mouseClickButton == 1) {
            privacyChatOn = !privacyChatOn;
            clientStream.b(111, 0);   // opcode 111 = GAME_SETTINGS_CHANGED
            clientStream.f.c(0, 41);
            clientStream.f.c(privacyChatOn ? 1 : 0, -107);
            clientStream.b(21294);    // flush
        }
        clickY += 15;

        // Toggle trade privacy
        if (mouseX > px && mouseX < px + pw && mouseY > clickY - 12 && mouseY < clickY + 4 && mouseClickButton == 2) {
            privacyTradeOn = !privacyTradeOn;
            clientStream.b(111, param - 15);
            clientStream.f.c(2, param ^ 0x55);
            clientStream.f.c(privacyTradeOn ? 1 : 0, -82);
            clientStream.b(param ^ 0x5321);
        }
        clickY += 15;

        // Toggle members privacy (members only)
        if (isMembersAccount && mouseX > px && mouseX < px + pw && mouseY > clickY - 12 && mouseY < clickY + 4 && mouseClickButton == 1) {
            privacyMembersOn = !privacyMembersOn;
            clientStream.b(111, 0);
            clientStream.f.c(3, param - 136);
            clientStream.f.c(privacyMembersOn ? 1 : 0, -42);
            clientStream.b(21294);
        }
        clickY += 15; clickY += 15; clickY += 15; clickY += 15; clickY += 15;

        // Toggle auto-retaliate
        boolean settingsChanged = false;
        if (mouseX > px && mouseX < px + pw && mouseY > (clickY += 35) - 12 && mouseY < clickY + 4 && mouseClickButton == 2) {
            autoRetaliateMode = 1 - autoRetaliateMode;
            settingsChanged = true;
        }
        // Toggle mouse mode
        if (mouseX > px && mouseX < px + pw && mouseY > (clickY += 15) - 12 && mouseY < clickY + 4 && mouseClickButton == 2) {
            singleMouseButton = 1 - singleMouseButton;
            settingsChanged = true;
        }
        // Toggle camera
        if (mouseX > px && mouseX < px + pw && mouseY > (clickY += 15) - 12 && mouseY < clickY + 4 && mouseClickButton == 2) {
            cameraAuto = 1 - cameraAuto;
            settingsChanged = true;
        }
        // Toggle members option
        if (isMembersAccount && mouseX > px && mouseX < px + pw
                && mouseY > (clickY += 15) - 12 && mouseY < clickY + 4 && mouseClickButton == 1) {
            settingsChanged = true;
            membersOnlyOption = 1 - membersOnlyOption;
        }
        clickY += 15;

        if (settingsChanged) {
            // opcode 111 (GAME_SETTINGS_CHANGED): Vg, dc, De, ui
            sendPrivacySettings(cameraAuto, singleMouseButton, autoRetaliateMode, param + 64, membersOnlyOption);
        }

        // Members logout (Kd only)
        if (isMembersWorld) {
            if (mouseX > px && mouseX < px + pw && mouseY > (clickY += 5) - 12 && mouseY < clickY + 4 && mouseClickButton == 1) {
                handleInventoryClick(ISAAC.g, param - 3, 9, false);  // obf: a(o.g,...)
                inventoryTab = 0;
            }
            clickY += 15;
        }

        // Logout button
        if (mouseX > px && mouseX < px + pw && mouseY > (clickY += 20) - 12 && mouseY < clickY + 4 && mouseClickButton == 1) {
            requestLogout(0);   // obf: B(0)
        }
        mouseClickButton = 0;
    }

    /**
     * Toggle one-button vs two-button mouse mode and persist to field eg.
     * Sets eg=77 (two-button) unless called with the sentinel param.
     * obf: void g(int)  obf-label: client.FC(
     */
    private void setMouseButtonMode(int param) {
        // Anti-tamper sentinel check stripped (dummy guard on param)
        // eg==77 means two-button mode; any other value = one-button
        if (param != -16433) {
            mouseButtonMode = 77;   // obf: eg  — two-button mode
        }
    }

// ---------------------------------------------------------------------------

    /**
     * Clear trade and duel offer state arrays.
     * Resets all pending offer slots so panels start clean.
     * obf: void o(int)  obf-label: client.FC( (overload)
     */
    private void resetTradeDuelState(int param) {
        tradeQueuedAction = 0;    // obf: bj
        activePanel = 0;          // obf: Xd
        chatInputMode = 0;        // obf: qg
        if (param != -2) {
            return;
        }
        inventoryTab = 0;         // obf: kc (secondary reset when param==-2 sentinel)
    }

// ---------------------------------------------------------------------------

    /**
     * Render the character-design arrow buttons (hover highlight only — no click handling here).
     * Draws left/right arrows for each design category when the mouse hovers over them.
     * obf: void w(int)  obf-label: client.GD(
     */
    private void drawCharDesignControls(int param) {
        surface.i = false;
        surface.a(true);
        panelDuel.a((byte)-13);    // obf: Af.a — update panel hover state

        // Layout constants: n3=256 (base x after +116 offset from 140), n4=25 (base y)
        int baseX = 140 + 116;  // = 256
        int baseY = 50 - 25;    // = 25

        // Draw the 9 character-design rows (body part + gender + 2 color wheels each):
        // Row structure: [skinColor at wg][hairColor at hh][topColor at ld][bottomColor at Vd][...]
        // surface.a(x, bodyPart, colorSwatch, y, arrowH, arrowW, arrowSprite)
        // Each row: left arrow (offset -55), right arrow (+6 from sprite w), bottom row (+55)

        surface.a(baseX - 87, ei[Lh], w.g[wg], baseY, 102, (byte)105, 64);
        surface.a(baseY, ei[Wg], Wh[hh], false, 0, w.g[dk], 102, 64, baseX - 87, 1);
        surface.a(baseY, Dg[ld], Wh[hh], false, 0, w.g[Vd], 102, 64, baseX - 87, param + 13760);
        surface.a(baseX - 32, ei[Lh], 6 + w.g[wg], baseY, 102, (byte)105, 64);
        surface.a(baseY, ei[Wg], Wh[hh], false, 0, w.g[dk] + 6, 102, 64, baseX - 32, 1);
        surface.a(baseY, Dg[ld], Wh[hh], false, 0, 6 + w.g[Vd], 102, 64, baseX - 32, 1);
        surface.a(baseX - 32 + 55, ei[Lh], 12 + w.g[wg], baseY, 102, (byte)110, 64);
        surface.a(baseY, ei[Wg], Wh[hh], false, 0, w.g[dk] + 12, 102, 64, baseX + 55 - 32, param + 13760);
        surface.a(baseY, Dg[ld], Wh[hh], false, 0, w.g[Vd] + 12, 102, 64, baseX - 32 + 55, 1);

        // Draw the bottom inventory border line and blit surface
        surface.b(-1, tg + 22, Oi, 0);
        surface.a(graphics, Eb, 256, K);
        if (param != -13759) {
            drawCloseButton(70);
        }
    }

// ---------------------------------------------------------------------------

    /**
     * Build and display the "Please design Your Character" appearance screen.
     * Creates arrow-button widgets in panelDuel (Af) for Head/Hair/Top/Bottom/Skin/Gender.
     * On "Accept" click sends opcode 235 (PLAYER_APPEARANCE_CHANGE).
     * obf: void t(int)  obf-label: client.J(
     */
    private void drawCharDesign(int param) {
        panelDuel = new Panel(surface, 100);   // obf: Af = new qa(li, 100)

        // Title label: STRINGS[87] = "Please design your Character"
        panelDuel.a(true, (byte)-125, 4, 256, STRINGS[87], 10);

        int x = 140 + 116;   // = 256 (centered column)
        int y = 34 - 10;     // = 24 (starting y)

        // Row 1: Head style (left/right arrows at x-55, x, x+55)
        panelDuel.a(true, (byte)-104, 3, x - 55, STRINGS[82], y + 110);
        panelDuel.a(true, (byte)-91, 3, x, STRINGS[92], y + 110);
        panelDuel.a(true, (byte)-117, 3, x + 55, STRINGS[81], y + 110);
        y += 145;

        // Gender toggle row: 54px wide sprite swatch
        int s = 54;
        panelDuel.a(41, x - s, 53, 26531, y);
        panelDuel.a(true, (byte)-81, 1, x - s, STRINGS[84], y - 8);
        panelDuel.a(true, (byte)-125, 1, x - s, STRINGS[88], y + 8);
        panelDuel.c(StringCodec.g + 7, y, x - s - 40, -114);   // obf: u.g
        Dj = panelDuel.d(x - s - 40, 20, y, param + 24525, 20);   // left-arrow button id

        panelDuel.c(6 + StringCodec.g, y, x - s + 40, -59);
        pi = panelDuel.d(x - s + 40, 20, y, param ^ 0x6049, 20);  // right-arrow button id

        panelDuel.a(41, x + s, 53, 26531, y);
        panelDuel.a(true, (byte)-85, 1, x + s, STRINGS[85], y - 8);
        panelDuel.a(true, (byte)-102, 1, x + s, STRINGS[86], y + 8);
        panelDuel.c(7 + StringCodec.g, y, x + s - 40, -57);
        Kj = panelDuel.d(x + s - 40, 20, y, 64, 20);

        panelDuel.c(6 + StringCodec.g, y, x + s + 40, -127);
        ed = panelDuel.d(x + s + 40, 20, y, param ^ 0xFFFF9FB6, 20);

        y += 50;

        // Row 3: Skin color / Hair color left/right pairs
        panelDuel.a(41, x - s, 53, 26531, y);
        panelDuel.a(true, (byte)-102, 1, x - s, STRINGS[91], y);
        panelDuel.c(StringCodec.g + 7, y, x - s - 40, param + 24525);
        Ge = panelDuel.d(x - s - 40, 20, y, -81, 20);

        panelDuel.c(StringCodec.g + 6, y, x - s + 40, param + 24521);
        Of = panelDuel.d(x - s + 40, 20, y, 54, 20);

        panelDuel.a(41, x + s, 53, param ^ 0xFFFFF84E, y);
        panelDuel.a(true, (byte)-102, 1, x + s, STRINGS[79], y - 8);
        panelDuel.a(true, (byte)-79, 1, x + s, STRINGS[86], y + 8);
        panelDuel.c(7 + StringCodec.g, y, x + s - 40, -104);
        Xc = panelDuel.d(x + s - 40, 20, y, param + 24504, 20);

        panelDuel.c(6 + StringCodec.g, y, x + s + 40, -105);
        ek = panelDuel.d(x + s + 40, 20, y, -91, 20);

        y += 50;
        if (param != -24595) {
            drawProgressBar(-127);  // redraw loading/progress bar
        }

        // Row 4: Top color, bottom color
        panelDuel.a(41, x - s, 53, param ^ 0xFFFFF84E, y);
        panelDuel.a(true, (byte)-81, 1, x - s, STRINGS[83], y - 8);
        panelDuel.a(true, (byte)-109, 1, x - s, STRINGS[86], y + 8);
        panelDuel.c(7 + StringCodec.g, y, x - s - 40, -59);
        Ze = panelDuel.d(x - s - 40, 20, y, param + 24468, 20);

        panelDuel.c(StringCodec.g + 6, y, x - s + 40, -95);
        Mj = panelDuel.d(x - s + 40, 20, y, param + 24637, 20);

        panelDuel.a(41, x + s, 53, 26531, y);
        panelDuel.a(true, (byte)-108, 1, x + s, STRINGS[89], y - 8);
        panelDuel.a(true, (byte)-108, 1, x + s, STRINGS[86], y + 8);
        panelDuel.c(StringCodec.g + 7, y, x + s - 40, -90);
        Re = panelDuel.d(x + s - 40, 20, y, 69, 20);

        panelDuel.c(6 + StringCodec.g, y, x + s + 40, param + 24537);
        Ai = panelDuel.d(x + s + 40, 20, y, -119, 20);

        y += 82;

        // Accept button row
        panelDuel.c(param ^ 0x6055, 200, 30, x, y - 35);
        panelDuel.a(false, (byte)-74, 4, x, STRINGS[90], y - 35);
        // Eg = button id for "Accept" — triggers opcode 235 in sendAppearance()
        Eg = panelDuel.d(x, 200, y - 35, param ^ 0xFFFF9FC9, 30);
    }

// ---------------------------------------------------------------------------

    /**
     * Append a formatted chat message to the chat list (chatList).
     * Formats with Helvetica font and color coding; also triggers a minimap redraw.
     * obf: void a(String,byte,String)  obf-label: client.MD(
     */
    private void addChatMessage(String header, byte colorCode, String body) {
        // colorCode sentinel: only -64 is valid; other values are dead (anti-tamper)
        if (colorCode != -64) {
            return;
        }
        Graphics g = this.getGraphics();
        if (null == g) {
            return;
        }
        g.translate(Eb, K);
        Font font = new Font(STRINGS[477], 1, 15);  // STRINGS[477] = "Helvetica"
        int w = 512;
        g.setColor(Color.black);
        int h = 344;
        // Draw black fill rect centered on screen
        g.fillRect(w / 2 - 140, h / 2 - 25, 280, 50);
        g.setColor(Color.white);
        g.drawRect(w / 2 - 140, h / 2 - 25, 280, 50);
        // Draw body above header
        this.a(font, body, h / 2 - 10, true, w / 2, g);
        this.a(font, header, h / 2 + 10, true, w / 2, g);
    }

// ---------------------------------------------------------------------------

    /**
     * Display a server/system message: routes to chatList or duelPanel depending on
     * activePanel (Xd==2 → duel/private panel), then triggers minimap + send-queued.
     * obf: void b(byte,String,String)
     */
    private void showServerMessage(byte triggerCode, String title, String body) {
        if (activePanel == 2) {
            // If no body text, show in first slot; otherwise both
            if (body != null && body.length() >= 1) {
                // fall through to dual-slot display
            } else {
                panelDuel.a(fatiguePercent, title, 27642);  // obf: yi.a(td, ...)
                // 27642 = color code for yellow server messages
                return;
            }
        }
        // Two-slot display: title in slot 0, body in slot 1
        panelDuel.a(fatiguePercent, title, 27642);   // obf: yi.a(Qi, ...)
        panelDuel.a(fatiguePercent, body, 27642);    // obf: yi.a(td, ...)

        // triggerCode < -11 means "force-redraw minimap and queued actions"
        if (triggerCode < -11) {
            drawMinimap(2540);
            sendQueuedActions(-28492);  // obf: c(-28492)
        }
    }

// ---------------------------------------------------------------------------

    /**
     * Render the right-click option list ("Choose option" menu).
     * Positions and draws option strings; on click sends the chosen action.
     * obf: void a(String[],int,int,boolean)  obf-label: client.RA(
     */
    private void drawMenuOptions(String[] options, int x, int y, boolean rightClick) {
        // If not called with sentinel 12, clear the screen first
        if (x != 12) {
            clearScreen((byte)31);   // obf: e((byte)31)
        }
        // Delegate to drawScrollList: offset x by -9 to give left margin
        drawScrollList(x - 9, y, options, rightClick, "");
    }

// ---------------------------------------------------------------------------

    /**
     * Generic scrollable list/menu renderer.
     * Computes required width from option text, then renders background + text rows.
     * obf: void a(int,int,String[],boolean,String)
     */
    private void drawScrollList(int x, int y, String[] options, boolean showBorder, String title) {
        menuOptionList = options;       // obf: od
        menuWidth = 400;                // obf: zi — will be expanded below
        // Expand width to longest option string
        for (int i = 0; i < options.length; i++) {
            int w = surface.a(1, x + 113, options[i]) + 10;
            if (menuWidth < w) {
                menuWidth = w;
            }
        }
        // Total height = font height * (items+1) + leading
        menuHeight = 15 + (surface.a(508305352, 1) + 2) * (1 + options.length)
                     + surface.a(508305352, 4);  // obf: gl
        menuX = y;           // obf: gc
        menuTitle = title;   // obf: e (temp label)
        menuOpenFlag = false; // obf: vk
        inputLine = "";       // obf: Cb
        showMenuBorder = showBorder;  // obf: Bd
    }

// ---------------------------------------------------------------------------

    /**
     * Scrollbar widget (primary variant).
     * Draws a scrollbar thumb at position derived from current scroll value.
     * Calls walkTo to move if needed, then optionally calls drawScrollbar2.
     * obf: void a(byte,int,int,int,boolean,int)  obf-label: client.WC(
     */
    private void drawScrollbar(byte sentinel, int x, int y, int scrollPos, boolean animate, int trackLen) {
        // Delegate to walkTo for hit-testing / interaction
        if (walkTo(x, trackLen, (byte)14, false, scrollPos, scrollPos, y, y, animate)) {
            return;
        }
        walkTo(scrollPos, animate, trackLen, y, x, scrollPos, true, y, sentinel + 107);
        if (sentinel != 10) {
            // Draw default scrollbar thumb indicator
            drawScrollbar2(99, 113, -126, -87, true, 125);
        }
    }

// ---------------------------------------------------------------------------

    /**
     * Secondary scrollbar/slider widget.
     * Draws a small filled rect representing the scroll thumb position.
     * obf: void a(int,int,int,int,boolean,int)  obf-label: client.BE(
     */
    private void drawScrollbar2(int x, int y, int w, int h, boolean animate, int trackLen) {
        // Simple wrapper: calls walkTo with a one-way (no-waypoints) path
        walkTo(y, animate, trackLen, x, w, y, false, x, 105);
        if (trackLen != 8) {
            lastActionTime = -85L;  // obf: Wi  — reset activity timer
        }
    }

// ---------------------------------------------------------------------------

    /**
     * Draw the help/close-window overlay ("Click here to close" or help panel).
     * Tracks item info for the sort-draw list and updates the draw order.
     * obf: void C(int)  obf-label: client.WB(
     */
    private void drawHelpMenu(int param) {
        // Copy current draw-list into shadow buffer
        drawListCount = drawListSize;   // obf: vj = fj
        for (int i = 0; i < drawListSize; i++) {
            drawListIds[i] = drawListCurrent[i];   // obf: ae[i] = ci[i]
            drawListY[i] = drawListYShadow[i];     // obf: di[i] = Xe[i]
        }

        // Merge inventory items into draw list (de-duplicate by item id)
        int n = 0;
        for (; n < inventoryCount; n++) {
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

        // Junk opaque-predicate tail stripped (n5 = -89, n4 = (param-2)/60, n3 = n5/n4)
    }

// ---------------------------------------------------------------------------

    /**
     * Draw the "Click here to close window" button at the bottom of modal panels.
     * Also closes any open right-click menu if clicked outside.
     * obf: void l(byte)
     */
    private void drawCloseButton(byte param) {
        if (param != -115) {
            closeButtonSpriteId = 64;   // obf: qd  — default sprite
        }
        int w = 400;
        int h = 100;
        if (helpMenuOpen) {   // obf: Wk
            h = 450;
            h = 300;  // obf: two assignments, second wins (dead first)
        }
        // Draw close button panel: centered at (256, 167) with h×w dimensions
        surface.a(-(w / 2) + 256, (byte)122, 0, 167 - h / 2, h, w);
        surface.e(-(w / 2) + 256, w, 167 - h / 2, 27785, h, 0xFFFFFF);
        surface.a(w - 40, helpMenuTitle, 256, 92, 1, 167 - (h / 2 - 20), true, 0xFFFFFF);

        int labelY = 157 + h / 2;
        int labelColor = 0xFFFFFF;
        // Highlight "Close" text if mouse is over it
        if (mouseY < labelY && mouseY >= labelY - 12 && mouseX > 107 && mouseX < 407) {
            labelColor = 0xFF0000;
        }
        surface.a(256, STRINGS[126], labelColor, 0, 1, labelY);  // STRINGS[126] = "Click here to close window"

        if (mouseClickButton == 2) {   // obf: Cf == 2 — right-click
            if (labelColor == 0xFF0000) {
                helpMenuOpen = false;   // obf: mh
            }
            // If click is outside the modal, close it
            if (!(mouseX >= 256 - w / 2 && mouseX <= 256 + w / 2
                    && mouseY >= 167 - h / 2 && mouseY <= 167 + h / 2)) {
                helpMenuOpen = false;
            }
        }
        mouseClickButton = 0;   // obf: Cf = 0
    }

// ---------------------------------------------------------------------------

    /**
     * Draw an editable text field box with current text and blinking cursor marker.
     * obf: void a(byte,int,String)
     */
    private void drawTextField(byte sentinel, int slotIndex, String text) {
        // Wall-model slot coords
        int wallZ = wallModelZ[slotIndex];   // obf: Se[n2]
        int wallX = wallModelX[slotIndex];   // obf: ye[n2]
        // Relative position from local player (wi.i/K = world coords * 128)
        int relX = wallZ - (localPlayer.i / 128);
        int relY = -(localPlayer.K / 128) + wallX;
        int range = 7;

        // Sentinel guard: only proceed for values > 2
        if (sentinel <= 2) return;
        if (wallZ < 0) return;
        if (wallX < 0) return;
        if (wallZ > 95) return;
        if (wallX > 95) return;
        if (relX <= -range) return;
        if (relX >= range) return;
        if (relY <= -range) return;
        if (relY >= range) return;

        // Remove previous wall model from scene and substitute model named 'text'
        world.a(wallModels[slotIndex], -1);
        int modelIdx = GameModel.a((byte)91, text);  // look up model name → index
        GameModel newModel = objectModels[modelIdx].b(-2);  // clone base model
        world.a(newModel, (byte)118);
        newModel.a(-50, 48, -10, -50, true, 48, -74);  // transform/scale
        newModel.a(wallModels[slotIndex], 6029);
        newModel.rb = slotIndex;
        wallModels[slotIndex] = newModel;
    }

// ---------------------------------------------------------------------------

    /**
     * Render the social dialog overlays: add-friend / send-PM / add-ignore entry boxes.
     * Manages Bj (socialDialogMode): 1=addFriend, 2=sendPM, 3=addIgnore.
     * obf: void h(byte)
     */
    private void drawSocialDialog(byte sentinel) {
        // If a right-click menu is open (Cf != 0), close it if click is outside bounds
        if (mouseClickButton != 0) {
            mouseClickButton = 0;
            if (socialDialogMode == 2 && (mouseX < 6 || mouseY < 145 || mouseX > 506 || mouseY > 215)) {
                socialDialogMode = 0;
                return;
            }
            if (socialDialogMode == 2 && (mouseX < 6 || mouseY < 145 || mouseX > 506 || mouseY > 215)) {
                socialDialogMode = 0;
                return;
            }
            if (socialDialogMode == 3 && (mouseX < 106 || mouseY < 145 || mouseX > 406 || mouseY > 215)) {
                socialDialogMode = 0;
                return;
            }
            if (mouseX > 236 && mouseX < 276 && mouseY > 193 && mouseY < 213) {
                socialDialogMode = 0;
                return;
            }
        }

        int y = 145;

        // Mode 1: Add Friend
        if (socialDialogMode == 1) {
            surface.a(106, (byte)26, 0, y, 70, 300);
            surface.e(106, 300, y, 27785, 70, 0xFFFFFF);
            surface.a(256, STRINGS[246], 0xFFFFFF, 0, 4, y += 20);  // "Add friend"
            surface.a(256, tempInputString + "*", 0xFFFFFF, 0, 4, y += 20);
            // If enter was pressed (Cb non-empty), submit add-friend
            String normalizedName = ChatCipher.a(localPlayer.C, (byte)50);
            if (null != normalizedName && inputLine.length() > 0) {
                String trimmed = inputLine.trim();
                socialDialogMode = 0;
                tempInputString = "";
                inputLine = "";
                if (trimmed.length() > 0 && !normalizedName.equals(ChatCipher.a(trimmed, (byte)100))) {
                    sendOpcodeString(114, trimmed);  // obf: b(114, trimmed) — add friend packet
                }
            }
        }

        // Mode 2: Send Private Message
        if (socialDialogMode == 2) {
            surface.a(6, (byte)110, 0, y, 70, 500);
            surface.e(6, 500, y, 27785, 70, 0xFFFFFF);
            surface.a(256, STRINGS[249] + pmTarget, 0xFFFFFF, 0, 4, y += 20);  // "Send PM to: "
            surface.a(256, pmInput + "*", 0xFFFFFF, 0, 4, y += 20);
            if (submittedInput.length() > 0) {
                String msg = submittedInput;
                pmInput = "";
                socialDialogMode = 0;
                submittedInput = "";
                sendPrivateMessage((byte)-76, pmTarget, msg);  // obf: a(byte,-76,...) opc 218
            }
        }

        // Mode 3: Add Ignore
        if (socialDialogMode == 3) {
            surface.a(106, (byte)-115, 0, y, 70, 300);
            surface.e(106, 300, y, 27785, 70, 0xFFFFFF);
            surface.a(256, STRINGS[248], 0xFFFFFF, 0, 4, y += 20);  // "Add ignore"
            surface.a(256, tempInputString + "*", 0xFFFFFF, 0, 4, y += 20);
            String normalizedName = ChatCipher.a(localPlayer.C, (byte)59);
            if (normalizedName != null && inputLine.length() > 0) {
                String trimmed = inputLine.trim();
                tempInputString = "";
                socialDialogMode = 0;
                inputLine = "";
                if (trimmed.length() > 0 && !normalizedName.equals(ChatCipher.a(trimmed, (byte)105))) {
                    sendAddIgnore(trimmed, (byte)5);  // obf: a(string, (byte)5)
                }
            }
        }

        // "Cancel" link always visible at (256, 208)
        int labelColor = 0xFFFFFF;
        if (mouseX > 236 && mouseX < 276 && mouseY > 193 && mouseY < 213) {
            labelColor = 0xFFFF00;
        }
        surface.a(256, STRINGS[121], labelColor, 0, 1, 208);  // STRINGS[121] = "Cancel"

        if (sentinel <= 77) {
            closeButtonSpriteId = -42;   // obf: pj — mark close-button pressed
        }
    }

// ---------------------------------------------------------------------------

    /**
     * "Enter the name of the player you wish to report" screen.
     * Shows report-abuse name entry + optional mute checkbox.
     * Submits opcode 206 (REPORT_ABUSE) when the player name is confirmed.
     * obf: void d(boolean)
     */
    private void drawReportNameEntry(boolean rightAlign) {
        // If player typed a name (Cb non-empty), accept it immediately
        if (inputLine.length() > 0) {
            reportAbuseTarget = inputLine.trim();   // obf: ec
            reportAbusePage = 0;                    // obf: Yb
            inputMode = 2;                          // obf: Vf  — switch to offence-picker
            return;
        }

        // Determine mute-option tier (0=none, 1=mute-only, 2=mute+offence)
        int muteTier = 0;
        if (reportAbuseMuteFlag < 2 && reportAbuseOffence < 7) {
            if (reportAbuseOffence >= 6) {
                muteTier = 1;
            } else {
                muteTier = 2;
            }
        }

        // Layout
        int fontH = surface.a(508305352, 1);
        int lineH = surface.a(508305352, 4);
        int panelW = 400;
        int panelH = (muteTier > 0 ? 5 + fontH : 0) + 70;
        int panelX = 256 - panelW / 2;
        int panelY = 180 - panelH / 2;

        surface.a(panelX, (byte)88, 0, panelY, panelH, panelW);
        surface.e(panelX, panelW, panelY, 27785, panelH, 0xFFFFFF);
        // STRINGS[340] = "Enter the name of the player you wish to report:"
        surface.a(256, STRINGS[340], 0xFFFF00, 0, 1, 5 + panelY + fontH);
        int inputY = fontH + 2;
        surface.a(256, tempInputString + "*", 0xFFFFFF, 0, 4,
                  lineH + (panelY + 5) + (inputY + 3));  // entry field with cursor

        int nextY = fontH + lineH + (8 + panelY + inputY + 2);
        int labelColor = 0xFFFFFF;

        // Show mute options if tier > 1
        if (muteTier > 1) {
            String muteLabel;
            if (reportAbuseMuteConfirmed) {
                muteLabel = STRINGS[336];   // "Mute player for 48 hours"
            } else {
                muteLabel = STRINGS[339];   // "Do not mute player"
            }
            if (muteTier >= 2) {
                muteLabel = muteLabel + STRINGS[341];   // extended mute text
            }
            muteLabel = muteLabel + STRINGS[337];       // suffix

            // Hit-test the mute toggle label
            int muteW = surface.a(1, 72, muteLabel);
            if (mouseX > 256 - muteW / 2 && mouseX < 256 + muteW / 2
                    && mouseY > nextY - fontH && mouseY <= nextY) {
                if (mouseClickButton != 0) {
                    reportAbuseMuteConfirmed = !reportAbuseMuteConfirmed;
                    mouseClickButton = 0;
                }
                labelColor = 0xFFFF00;
            }
            surface.a(256, muteLabel, labelColor, 0, 1, nextY);
            nextY += 10 + fontH;
        }

        // "Submit" link
        labelColor = 0xFFFFFF;
        if (mouseX > 211 && mouseX < 228 && mouseY > nextY - fontH && mouseY <= nextY) {
            if (mouseClickButton != 0) {
                reportAbuseTarget = tempInputString;
                mouseClickButton = 0;
            }
            labelColor = 0xFFFF00;
        }
        surface.a(STRINGS[122], 210, nextY, labelColor, rightAlign, 1);  // STRINGS[122] = "Submit"

        // "Cancel" link
        labelColor = 0xFFFFFF;
        if (mouseX > 265 && mouseX < 304 && mouseY > nextY - fontH && mouseY <= nextY) {
            labelColor = 0xFFFF00;
            if (mouseClickButton != 0) {
                mouseClickButton = 0;
                inputMode = 0;  // obf: Vf = 0
            }
        }
        surface.a(STRINGS[121], 264, nextY, labelColor, rightAlign, 1);  // STRINGS[121] = "Cancel"

        // Click outside modal → cancel
        if (mouseClickButton == 2) {
            if (mouseX < panelX || mouseX > panelX + panelW || mouseY < panelY || mouseY > panelY + panelH) {
                inputMode = 0;
                mouseClickButton = 0;
            }
        }
    }
