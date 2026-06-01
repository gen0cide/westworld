    // ===== mainloop =====
    //
    // Per-tick game logic, connection keep-alive, logout flow, and the three
    // top-right UI tabs whose obfuscated bodies the skeleton grouped under
    // "mainloop".
    //
    // CORRECTNESS-AUDIT NOTES (re-verified method-by-method against the CLEAN
    // Vineflower base `decompiled/normalized-clean/client.java`, cross-checked
    // vs `decompiled/cfr/client.java`, the mudclient204 oracle, and the
    // OpenRSC Payload235 parser). Several reconstructed bodies in the previous
    // pass were WRONG; the fixes are flagged inline with `// FIX:`.
    //
    // CROSS-CLASS NAMING (per docs/NAMING.md — k=World, lb=Scene):
    //   * field `Ek` has type `lb` => Ek is the Scene  (client.scene).
    //   * field `Hh` has type `k`  => Hh is the World   (client.world).
    //   The skeleton's field table lists these swapped; NAMING.md is honored here.
    //
    // SKELETON-MISLABEL NOTES (verified against the oracle):
    //   * a(boolean,byte)@minimap the skeleton calls "tick"            -> drawUiTabMinimap.
    //   * b(boolean,byte)        the skeleton calls "updateGameState"  -> drawUiTabMagic.
    //   * J(int)                 the skeleton calls "drawSleepScreen"  -> handleGameInput (the real tick).
    // Methods are named for what they actually do; obf signatures are kept in comments.

    /**
     * GameShell stop hook. Tears down the game session: confirms logout to the
     * server (opcode 31) and stops the active sound voice.
     *
     * @param fromShell true when invoked by the shell's stop path (sets the tick marker).
     */
    // obf: final void a(boolean)   [client.SA(]   proposed: onStopGame
    final void onStopGame(boolean fromShell) {
        if (fromShell) {
            tickMarker = -103L;                 // ze: scratch timing marker
        }
        // a(true, 31) -> sendConfirmLogoutAck: opcode 31 (CONFIRM_LOGOUT) + stream teardown.
        sendConfirmLogoutAck(true, 31);
        if (soundChannel == null) {             // ni: active audio voice
            return;
        }
        soundChannel.d();                       // stop/close the sound channel
    }

    /**
     * Send a server keep-alive and pump one inbound packet. Called once per tick
     * by {@link #handleGameInput}. If no packet has arrived for >5s a PING
     * (opcode 67) is sent; then any pending writes are flushed and one inbound
     * packet is read and dispatched.
     *
     * The parameter is an obfuscation magic: it is called as {@code K(0 - 26345)},
     * i.e. {@code magic == -26345}, which is what makes the embedded arg
     * arithmetic (read length {@code magic+26345 == 0}, dispatch tag
     * {@code magic ^ -26304 == 87}) resolve correctly.
     */
    // obf: void K(int)   [client.SB(]   proposed: sendHeartbeat
    private final void sendHeartbeat(int magic) {
        long now = Timer.a(0);                  // p.a(0) = System.currentTimeMillis()

        // Wi = packetLastRead timestamp (activity timer reused for net liveness).
        if (clientStream.a((byte) 34)) {        // Jh.a(34) = hasPacket(): data arrived
            lastActionTime = now;
        }
        // clean: if (-5001 > ~(now - lastActionTime))  <=>  (now - lastActionTime) > 5000ms idle
        if ((now - lastActionTime) > 5000L) {
            lastActionTime = now;
            clientStream.b(67, 0);              // opcode 67 (HEARTBEAT / CL_PING)
            clientStream.b(21294);              // flush packet
        }

        try {
            clientStream.a(20, true);           // writePacket(20): flush queued writes
        } catch (IOException ex) {
            closeConnection(123);               // u(...) = "Lost connection" teardown
            return;
        }

        if (!hasInboundData((byte) -125)) {     // f(...) = data ready to read?
            return;
        }
        // readPacket(incomingPacket); arg (magic+26345)==0 selects the real read path.
        int size = clientStream.a(magic + 26345, incomingPacket);
        if (size <= 0) {                        // clean: ~size < -1  <=>  size > 0
            return;
        }
        // Dispatch one server->client packet. First arg (magic ^ -26304 == 87) is a
        // dispatch magic; mg.a(104) reads the (de-ISAAC'd) opcode byte.
        handlePacket(magic ^ -26304, size, incomingPacket.a((byte) 104));
    }

    /**
     * Request a normal logout (opcode 102, LOGOUT). Refused while in / shortly
     * after combat. On success an internal logout timer is armed so the
     * "Logging out..." dialog shows until the server drops us.
     *
     * @param combatGrace usually 0 (the post-combat grace threshold to compare
     *        the combat timer against); passed through from the call site.
     */
    // obf: void B(int)   [client.T(]   proposed: requestLogout
    private final void requestLogout(int combatGrace) {
        if (loggedIn == 0) {                    // clean: ~qg == -1  =>  qg == 0  =>  not logged in
            return;
        }
        if (combatTimeout > 450) {              // ai > 450: in combat
            showServerMessage(STRINGS[421], 3); // "@cya@You can't logout during combat!"
            return;
        }
        if (combatGrace < combatTimeout) {      // clean: var1 < ai; var1 is 0 -> within 10s grace
            showServerMessage(STRINGS[420], 3); // "@cya@You can't logout for 10 seconds after combat"
            return;
        }
        clientStream.b(102, 0);                 // opcode 102 (LOGOUT)
        clientStream.b(21294);                  // flush
        logoutTimeout = 1000;                   // bj: arm "Logging out..." dialog
    }

    /**
     * Abort an in-progress logout: clears the logout timer and shows the
     * "can't logout" notice (server told us the request was rejected).
     */
    // obf: void g(byte)   [client.CB(]   proposed: sendConfirmLogout
    private final void sendConfirmLogout(byte unused) {
        logoutTimeout = 0;                      // bj: cancel "Logging out..." dialog
        showServerMessage(STRINGS[64], 3);      // "@cya@Sorry, you can't logout at the moment"
    }

    /**
     * Draw the small modal "Logging out..." dialog box in the centre of the screen.
     */
    // obf: void d(byte)   [client.SD(]   proposed: doLogout
    private final void doLogout(byte unused) {
        surface.a(126, (byte) 52, 0, 137, 60, 260);          // drawBox(126,137,260,60, black)
        surface.e(126, 260, 137, 27785, 60, 16777215);       // drawBoxEdge(126,137,260,60, white)
        surface.a(256, STRINGS[679], 16777215, 0, 5, 173);   // drawStringCenter("Logging out...",256,173)
    }

    /**
     * Render + service the "Map" (minimap) UI tab in the top-right panel.
     *
     * Draws the rotated minimap landscape sprite plus coloured dots for nearby
     * scenery (cyan), ground items (red), NPCs (yellow) and players (white, or
     * green for friends), then the white centre dot and compass. When menus are
     * enabled, a left-click inside the map walks the player to the clicked tile.
     *
     * NOTE: skeleton mislabels this as "tick". It is drawUiTabMinimap.
     * The byte param is an obf magic (called with 125); only the dead
     * {@code if (var2 <= 119)} anti-tamper branch reads it.
     */
    // obf: void a(boolean,byte)   proposed (skeleton): tick   actual: drawUiTabMinimap
    private final void drawUiTabMinimap(boolean handleMenus, byte unused) {
        int uiX = surface.u - 199;              // li.u = surface.width2 (right edge)
        int uiWidth = 156;
        int uiHeight = 152;
        surface.b(-1, 2 + spriteMedia, 3, -49 + uiX);        // drawSprite tab background
        uiX += 40;
        surface.a(uiX, (byte) -125, 0, 36, uiHeight, uiWidth);          // drawBox(uiX,36,w,h,black)
        surface.a(uiX, uiWidth + uiX, 36 + uiHeight, 36, (byte) 76);    // setBounds(clip rect)

        // Rotation/zoom for the minimap. sd=minimapRandom2 (zoom jitter), ug=cameraRotation,
        // Df=minimapRandom1 (rotation offset). cc = SurfaceSprite.sin2048Cache (fixed-point sin/cos).
        // GameCharacter (ta) accessors: .currentX=obf .i, .currentY=obf .K, .hash=obf .C.
        int zoom = 192 + minimapRandom2;
        int rot = (cameraRotation + minimapRandom1) & 255;
        int px = (localPlayer.currentX - 6040) * zoom * 3 / 2048;
        int py = (localPlayer.currentY - 6040) * zoom * 3 / 2048;
        int sinR = SurfaceSprite.cc[(1024 - 4 * rot) & 0x3ff];
        int cosR = SurfaceSprite.cc[((1024 - 4 * rot) & 0x3ff) + 1024];
        int rx = px * cosR + py * sinR >> 18;   // >>18: 2048*2048 -> divide back (junk shift masked to 18)
        py = -(px * sinR) + py * cosR >> 18;    // (2D rotate the point by -rot)
        px = rx;
        // FIX: landscape minimap sprite id is `spriteMedia - 1`, NOT `uiX - 1`.
        //      obf: this.li.a(-1 + this.tg, ...)   tg = spriteMedia.
        surface.a(spriteMedia - 1, 36 - (-(uiHeight / 2) + -py), uiWidth / 2 + uiX - px, 842218000, zoom, (64 + rot) & 255);

        // Scenery dots (cyan = 0x00FFFF). eh=objectCount, ye/Se = objectX/objectY, Ug=magicLoc.
        for (int i = 0; i < objectCount; i++) {
            int dy = zoom * (64 + (magicLoc * objectY[i] - localPlayer.currentY)) * 3 / 2048;
            int dx = 3 * ((magicLoc * objectX[i] - (-64 - -localPlayer.currentX)) * zoom) / 2048;
            int rdx = cosR * dx + dy * sinR >> 18;
            dy = cosR * dy + -(sinR * dx) >> 18;
            dx = rdx;
            drawMinimapEntity(65535, dx + uiX + uiWidth / 2, (byte) -61, -dy + 36 - -(uiHeight / 2));
        }

        // Ground-item dots (red = 0xFF0000). Ah=groundItemCount, Zf/Ni = groundItemX/Y.
        for (int i = 0; i < groundItemCount; i++) {
            int dx = zoom * ((-localPlayer.currentX + (64 + groundItemX[i] * magicLoc)) * 3) / 2048;
            int dy = zoom * 3 * (-localPlayer.currentY + (64 + magicLoc * groundItemY[i])) / 2048;
            int rdx = cosR * dx + sinR * dy >> 18;
            dy = cosR * dy + -(dx * sinR) >> 18;
            dx = rdx;
            drawMinimapEntity(0xFF0000, uiX - (-(uiWidth / 2) + -dx), (byte) -53, uiHeight / 2 + 36 - dy);
        }

        // NPC dots (yellow = 0xFFFF00). Tb=npcsLast, de=npcsLastCount.
        for (int i = 0; i < npcsLastCount; i++) {
            GameCharacter npc = npcsLast[i];
            int dy = zoom * ((npc.currentY + -localPlayer.currentY) * 3) / 2048;
            int dx = 3 * ((npc.currentX + -localPlayer.currentX) * zoom) / 2048;
            int rdx = dy * sinR - -(dx * cosR) >> 18;
            dy = -(dx * sinR) + cosR * dy >> 18;
            dx = rdx;
            drawMinimapEntity(0xFFFF00, uiWidth / 2 + (uiX - -dx), (byte) -93, -dy + uiHeight / 2 + 36);
        }

        // Player dots (white = 0xFFFFFF, green = 0x00FF00 if on the friends list).
        // rg=playersLast, Yc=playersLastCount.
        for (int i = 0; i < playersLastCount; i++) {
            GameCharacter player = playersLast[i];
            int dx = 3 * ((-localPlayer.currentX + player.currentX) * zoom) / 2048;
            int dy = zoom * (player.currentY + -localPlayer.currentY) * 3 / 2048;
            int rdx = dx * cosR + sinR * dy >> 18;
            dy = cosR * dy - dx * sinR >> 18;
            dx = rdx;
            int colour = 0xFFFFFF;
            String name = WorldEntity.a(player.hash, (byte) 82);    // hashed name of this player
            if (name != null) {
                for (int f = 0; f < FontWidths.g; f++) {            // n.g = friendListCount
                    boolean isFriend = name.equals(WorldEntity.a(Surface.h[f], (byte) 107));
                    if (isFriend && (friendOnlineState[f] & 2) != 0) {   // Fj[f]&2 = friend online
                        colour = 0x00FF00;
                        break;
                    }
                }
            }
            drawMinimapEntity(colour, dx + (uiX - -(uiWidth / 2)), (byte) -67, -dy + 36 - -(uiHeight / 2));
        }

        // Centre marker (local player) + compass sprite, then restore the full-screen clip.
        surface.c(255, -1057205208, 2, uiHeight / 2 + 36, 0xFFFFFF, uiX - -(uiWidth / 2));   // drawCircle
        surface.a(spriteMedia + 24, 55, uiX - -19, 842218000, 128, (cameraRotation + 128) & 255);
        surface.a(0, gameWidth, gameHeight + 12, 0, (byte) 119);     // setBounds(full screen)

        if (!handleMenus) {
            return;
        }
        // Left-click inside the map area -> walk to the corresponding world tile.
        int mx = mouseX - (surface.u - 199);
        int my = mouseY - 36;
        if (mx >= 40 && my >= 0 && mx < 196 && my < 152) {
            int z = 192 + minimapRandom2;
            int r = (cameraRotation + minimapRandom1) & 255;
            int base = (surface.u - 199) + 40;
            // unproject screen offset -> world delta (16384 = 1<<14 fixed point; >>15 == /32768)
            int wy = 16384 * (mouseY - uiHeight / 2 - 36) / (z * 3);
            int wx = 16384 * (mouseX - (base + uiWidth / 2)) / (z * 3);
            int s2 = SurfaceSprite.cc[(1024 - r * 4) & 0x3ff];
            int c2 = SurfaceSprite.cc[((1024 - r * 4) & 0x3ff) + 1024];
            int rwx = wy * s2 - -(c2 * wx) >> 15;
            wy = c2 * wy - s2 * wx >> 15;
            wx = rwx + localPlayer.currentX;
            wy = localPlayer.currentY - wy;
            if (mouseButtonClick == 1) {        // Cf == 1: a fresh left-click this tick
                // obf: a(worldY>>7, worldX>>7, localRegionX, localRegionY, false, 8)
                walkToActionSource(wy / 128, wx / 128, localRegionX, localRegionY, false, 8);
            }
            mouseButtonClick = 0;
        }
    }

    /**
     * Render + service the "Magic" / "Prayers" UI tab (toggled by tabMagicPrayer).
     *
     * Lists castable spells (colour-coded by rune availability and level) or
     * prayers (colour-coded by level and prayer points), shows the hovered
     * entry's description, and on click either selects a spell to cast or
     * toggles a prayer, sending opcode 60 (PRAYER_ACTIVATED) / 254
     * (PRAYER_DEACTIVATED).
     *
     * NOTE: skeleton mislabels this as "updateGameState". It is drawUiTabMagic.
     * The byte param is an obf magic (called with -74): the embedded
     * {@code var2+74 == 0}, {@code var2^-74 == 0}, {@code var2+88 == 14},
     * {@code var2+17124 == 17050} arithmetic only resolves with -74.
     */
    // obf: void b(boolean,byte)   proposed (skeleton): updateGameState   actual: drawUiTabMagic
    private final void drawUiTabMagic(boolean handleMenus, byte unused) {
        int uiX = -199 + surface.u;
        int uiY = 36;
        surface.b(-1, spriteMedia + 4, 3, -49 + uiX);        // drawSprite tab background
        int uiWidth = 196;
        int uiHeight = 182;

        // Highlight the active sub-tab header brighter (220) than the inactive (160).
        // leftShade = Magic header (bright when tabMagicPrayer==0); rightShade = Prayer header.
        int leftShade, rightShade;
        leftShade = rightShade = ISAAC.a(160, 9570, 160, 160);     // o.a(...) = Surface.rgb2long
        if (tabMagicPrayer != 0) {
            rightShade = ISAAC.a(220, 9570, 220, 220);             // prayers tab active
        } else {
            leftShade = ISAAC.a(220, 9570, 220, 220);              // magic tab active
        }
        surface.c(128, uiX, 24, 0, uiY, uiWidth / 2, leftShade);
        surface.c(128, uiWidth / 2 + uiX, 24, 0, uiY, uiWidth / 2, rightShade);
        surface.c(128, uiX, 90, 0, uiY + 24, uiWidth, ISAAC.a(220, 9570, 220, 220));
        surface.c(128, uiX, uiHeight - 24 - 90, 0, uiY + 24 + 90, uiWidth, ISAAC.a(160, 9570, 160, 160));
        surface.b(uiWidth, 0, uiX, uiY + 24, (byte) 70);     // drawLineHoriz under headers
        surface.b(uiX - -(uiWidth / 2), 0 + uiY, 0, 24, 0);  // drawLineVert between headers
        surface.b(uiWidth, 0, uiX, uiY + 113, (byte) -92);   // drawLineHoriz under list
        surface.a(uiWidth / 4 + uiX, STRINGS[16], 0, 0, 4, 16 + uiY);                 // "Magic"
        surface.a(uiX + uiWidth / 4 + uiWidth / 2, STRINGS[21], 0, 0, 4, 16 + uiY);   // "Prayers"

        if (tabMagicPrayer == 0) {
            // --- Spell list ---
            panelMagic.c((byte) 118, controlListMagic);    // clearList
            int row = 0;
            for (int spell = 0; spell < EntityDef.b; spell++) {       // spellCount
                String colour = STRINGS[20];     // "@yel@" (have all runes)
                for (int rune = 0; rune < ISAAC.p[spell]; rune++) {   // spellRunesRequired
                    int runeId = NameHash.d[spell][rune];             // spellRunesId
                    if (!hasInventoryItems(ClientStream.J[spell][rune], runeId)) {
                        colour = STRINGS[15];     // "@whi@" (missing a rune)
                        break;
                    }
                }
                if (ImageLoader.f[spell] > skillCurrent[6]) {         // spellLevel > magic level
                    colour = STRINGS[19];        // "@bla@" (level too low)
                }
                panelMagic.a(row++, null, -116, 0, null,
                        colour + STRINGS[18] + ImageLoader.f[spell] + STRINGS[12] + BitBuffer.L[spell], controlListMagic);
            }
            panelMagic.a((byte) -92);            // drawPanel
            int sel = panelMagic.b(controlListMagic, 17050);          // getListEntryIndex
            if (sel != -1) {
                surface.a(STRINGS[18] + ImageLoader.f[sel] + STRINGS[12] + BitBuffer.L[sel], 2 + uiX, uiY + 124, 0xFFFF00, false, 1);
                surface.a(NameHash.a[sel], 2 + uiX, 136 + uiY, 0xFFFFFF, false, 0);       // spellDescription
                for (int rune = 0; rune < ISAAC.p[sel]; rune++) {
                    int runeId = NameHash.d[sel][rune];
                    surface.b(-1, Surface.Bb[runeId] + spriteItem, uiY + 150, 2 + uiX + rune * 44);   // rune icon
                    int have = getInventoryCount(runeId);
                    int need = ClientStream.J[sel][rune];
                    String s = hasInventoryItems(need, runeId) ? STRINGS[27] : STRINGS[10]; // "@gre@" : "@red@"
                    surface.a(s + have + "/" + need, 2 + (uiX + rune * 44), uiY + 150, 0xFFFFFF, false, 1);
                }
            } else {
                surface.a(STRINGS[14], uiX + 2, uiY + 124, 0, false, 1);   // "Point at a spell for a description"
            }
        }

        if (tabMagicPrayer == 1) {
            // --- Prayer list ---
            panelMagic.c((byte) 90, controlListMagic);    // clearList
            int row = 0;
            for (int prayer = 0; prayer < EntityDef.g; prayer++) {    // prayerCount
                String colour = STRINGS[15];     // "@whi@"
                if (skillBase[5] < GameModel.B[prayer]) {             // prayer base < prayerLevel
                    colour = STRINGS[19];        // "@bla@"
                }
                if (prayerOn[prayer]) {          // bk[]: prayer currently active
                    colour = STRINGS[27];        // "@gre@"
                }
                panelMagic.a(row++, null, -113, 0, null,
                        colour + STRINGS[18] + GameModel.B[prayer] + STRINGS[12] + EntityDef.h[prayer], controlListMagic);
            }
            panelMagic.a((byte) -7);             // drawPanel
            int sel = panelMagic.b(controlListMagic, 17050);
            if (sel != -1) {
                surface.a(uiX - -(uiWidth / 2), STRINGS[18] + GameModel.B[sel] + STRINGS[12] + EntityDef.h[sel], 0xFFFF00, 0, 1, uiY + 130);
                surface.a(uiX - -(uiWidth / 2), TextEncoder.e[sel], 0xFFFFFF, 0, 0, 145 + uiY);    // prayerDescription
                surface.a(uiX - -(uiWidth / 2), STRINGS[26] + ClientIOException.c[sel], 0, 0, 1, 160 + uiY);   // "Drain rate: "
            } else {
                surface.a(STRINGS[11], uiX - -2, uiY + 124, 0, false, 1);   // "Point at a prayer for a description"
            }
        }

        if (!handleMenus) {
            return;
        }
        int mx = mouseX - (surface.u - 199);
        int my = mouseY - 36;
        if (mx < 0 || my < 0 || mx >= 196 || my >= 182) {
            return;
        }
        // handleMouse(mouseButton, mouseY, junk, mouseLastButton, mouseX) on the magic panel.
        // obf: Mc.b(Bb, my+36, -9989, Qb, mx + (surface.u-199)).
        panelMagic.b(mouseButton, mouseY, -9989, mouseLastButton, mouseX);

        // Header click toggles between Magic (left) and Prayers (right).
        if (my <= 24 && mouseButtonClick == 1) {
            if (mx < 98 && tabMagicPrayer == 1) {
                tabMagicPrayer = 0;
                panelMagic.e(controlListMagic, 14);    // resetListProps
            } else if (mx > 98 && tabMagicPrayer == 0) {
                tabMagicPrayer = 1;
                panelMagic.e(controlListMagic, 14);
            }
        }

        // Click a spell -> select it (level + rune checks first).
        if (mouseButtonClick == 1 && tabMagicPrayer == 0) {
            int sel = panelMagic.b(controlListMagic, 17050);
            if (sel != -1) {
                if (skillCurrent[6] >= ImageLoader.f[sel]) {     // magic level OK
                    int rune;
                    for (rune = 0; rune < ISAAC.p[sel]; rune++) {
                        int runeId = NameHash.d[sel][rune];
                        if (!hasInventoryItems(ClientStream.J[sel][rune], runeId)) {
                            // FIX: missing reagents is il[25], not il[24] (indices were swapped).
                            showServerMessage(STRINGS[25], 3);   // "You don't have all the reagents you need for this spell"
                            rune = -1;
                            break;
                        }
                    }
                    if (rune == ISAAC.p[sel]) {
                        selectedSpell = sel;
                        selectedItemInventoryIndex = -1;
                    }
                } else {
                    // FIX: magic-level-too-low is il[24], not il[25].
                    showServerMessage(STRINGS[24], 3);   // "Your magic ability is not high enough for this spell"
                }
            }
        }

        // Click a prayer -> toggle it; sends PRAYER_ACTIVATED (60) / PRAYER_DEACTIVATED (254).
        if (mouseButtonClick == 1 && tabMagicPrayer == 1) {
            int sel = panelMagic.b(controlListMagic, 17050);
            if (sel != -1) {
                if (skillBase[5] < GameModel.B[sel]) {
                    showServerMessage(STRINGS[23], 3);   // "Your prayer ability is not high enough for this prayer"
                } else if (skillCurrent[5] == 0) {
                    showServerMessage(STRINGS[28], 3);   // "You have run out of prayer points..."
                } else if (!prayerOn[sel]) {
                    clientStream.b(60, 0);              // opcode 60 (PRAYER_ACTIVATED)
                    clientStream.f.c(sel, 57);          // putByte(prayerId)
                    clientStream.b(21294);              // flush
                    prayerOn[sel] = true;
                    playSoundFile(STRINGS[22]);         // "prayeron"
                } else {
                    clientStream.b(254, 0);             // opcode 254 (PRAYER_DEACTIVATED)
                    clientStream.f.c(sel, 37);          // putByte(prayerId)
                    clientStream.b(21294);              // flush
                    prayerOn[sel] = false;
                    playSoundFile(STRINGS[17]);         // "prayeroff"
                }
            }
        }
        mouseButtonClick = 0;
    }

    /**
     * Render + service the generic "enter amount / answer" modal dialog and,
     * once the player confirms, flush the queued client action it was gathering.
     *
     * The dialog kind is held in `inputDialogType` (gc). On confirm the typed
     * number is parsed and the matching packet is sent (gc -> action):
     *   3 -> bank withdraw (opcode 22), 4 -> bank deposit (23),
     *   5 -> shop buy (236), 6 -> shop sell (221), 9 -> skip tutorial (84);
     *   1/2/7/8 route through intra-class quantity/item-action wrappers.
     *
     * IMPORTANT: this method is dual-purpose.
     *   * Called as c(-43) from the panel render loop: renders the box AND
     *     services the Ok/Cancel buttons (the {@code if (var1 == -43)} gate).
     *   * Called as c(-97) elsewhere: only flushes a confirmed action; the
     *     bottom button hit-test is skipped.
     * The embedded packet keys (393 for putShort, -422797528 for putInt,
     * 21294 for flush) are anti-tamper guards in Buffer.e/Buffer.b and only
     * resolve to those exact values when var1 == -43 — see Buffer.e:
     * {@code if (n2 != 393) return;} and Buffer.b: {@code if (n2 != -422797528) inject byte}.
     */
    // obf: void c(byte)   [client.AB(]   proposed: sendQueuedActions   (dialogKind == gc)
    private final void sendQueuedActions(byte var1) {
        // --- Confirmation path: a value has been submitted (Cb non-empty) or OK was
        //     latched last tick (vk = inputDialogConfirmed). clean: !(Cb.length()<=0 && !vk). ---
        if (inputTextFinal.length() > 0 || inputDialogConfirmed) {
            String value = inputTextFinal.trim();
            inputTextCurrent = "";
            inputTextFinal = "";

            // gc (inputDialogType) selects which queued action to flush. The bare a()/b()/c()
            // wrappers build their own packets. ae[Rd]=bank slot item id;
            // Rj[Di]/Jf[Di]=selected shop slot item id / price.
            if (inputDialogType == 1) {             // generic "enter amount" wrapper
                try {
                    sendItemAction(Integer.parseInt(value), (byte) 9, dialogItemId);
                } catch (NumberFormatException ignored) {
                }
            } else if (inputDialogType == 2) {      // -> c(amount, 124, itemId) wrapper
                try {
                    sendItemActionAlt(Integer.parseInt(value), (byte) 124, dialogItemId);
                } catch (NumberFormatException ignored) {
                }
            } else if (inputDialogType == 3) {      // Bank withdraw: opcode 22 (BANK_WITHDRAW).
                try {
                    int itemId = (bankSelectedSlot >= 0) ? bankItems[bankSelectedSlot] : -1;
                    int amount = Integer.parseInt(value);
                    clientStream.b(22, 0);
                    clientStream.f.e(393, itemId);              // putShort(itemId)
                    clientStream.f.b(-422797528, amount);       // putInt(amount)
                    clientStream.f.b(-422797528, 0x12345678);   // putInt(magic/checksum)
                    clientStream.b(21294);                      // flush
                } catch (NumberFormatException ignored) {
                }
            } else if (inputDialogType == 4) {      // Bank deposit: opcode 23 (BANK_DEPOSIT).
                try {
                    // clean inverts the ternary: (Rd < 0) ? -1 : ae[Rd]  (same as withdraw).
                    int itemId = (bankSelectedSlot < 0) ? -1 : bankItems[bankSelectedSlot];
                    int amount = Integer.parseInt(value);
                    clientStream.b(23, 0);
                    clientStream.f.e(393, itemId);              // putShort(itemId)   [var1+436 == 393]
                    clientStream.f.b(-422797528, amount);       // putInt(amount)
                    clientStream.f.b(-422797528, 0x87654321);   // putInt(magic/checksum)
                    clientStream.b(21294);                      // flush
                } catch (NumberFormatException ignored) {
                }
            } else if (inputDialogType == 6) {      // Shop sell: opcode 221 (SHOP_SELL).
                try {
                    // FIX: guard is `!= -1` (clean: ~Rj[Di] != 0), not `!= 0`.
                    if (shopSelectedItemId[shopSelectedSlot] != -1) {
                        int amount = Integer.parseInt(value);
                        clientStream.b(221, 0);
                        clientStream.f.e(393, shopSelectedItemId[shopSelectedSlot]);    // item id
                        clientStream.f.e(393, shopSelectedItemPrice[shopSelectedSlot]); // price
                        clientStream.f.e(393, amount);          // amount   [var1+436 == 393]
                        clientStream.b(21294);
                    }
                } catch (NumberFormatException ignored) {
                }
            } else if (inputDialogType == 7) {      // -> b(109, amount, itemId) wrapper
                try {
                    sendItemActionB(109, Integer.parseInt(value), dialogItemId2);
                } catch (NumberFormatException ignored) {
                }
            } else if (inputDialogType == 8) {      // -> a(itemId, amount, -78) wrapper
                try {
                    sendItemActionC(dialogItemId2, Integer.parseInt(value), (byte) -78);
                } catch (NumberFormatException ignored) {
                }
            } else if (inputDialogType == 9) {      // Skip tutorial: opcode 84 (SKIP_TUTORIAL), no payload.
                clientStream.b(84, 0);
                clientStream.b(21294);
            } else {                                // case 5 (and clean's fall-through): Shop buy, opcode 236.
                try {
                    // FIX: guard is `!= -1` (clean: ~Rj[Di] != 0), not `!= 0`.
                    if (shopSelectedItemId[shopSelectedSlot] != -1) {
                        int amount = Integer.parseInt(value);
                        clientStream.b(236, 0);
                        clientStream.f.e(393, shopSelectedItemId[shopSelectedSlot]);    // item id   [var1^-420 == 393]
                        clientStream.f.e(393, shopSelectedItemPrice[shopSelectedSlot]); // price
                        clientStream.f.e(393, amount);
                        clientStream.b(21294);                  // flush   [var1+21337 == 21294]
                    }
                } catch (NumberFormatException ignored) {
                }
            }
            inputDialogType = 0;
            return;
        }

        // --- Render path (still waiting on input) ---
        // For numeric dialog kinds (gc 1..8) strip any non-digit chars from the live text.
        int boxX;
        if (inputDialogType >= 1 && inputDialogType <= 8) {
            StringBuilder digits = new StringBuilder();
            for (int n = 0; n < inputTextCurrent.length(); n++) {
                char ch = inputTextCurrent.charAt(n);
                if (Character.isDigit(ch)) {
                    digits.append(ch);
                }
            }
            inputTextCurrent = digits.toString();
        }
        boxX = 256 - inputDialogWidth / 2;

        // Draw the modal box, the prompt lines, and (only when called as c(-43)) Ok / Cancel.
        int boxY = 180 - inputDialogHeight / 2;
        surface.a(boxX, (byte) -103, 0, boxY, inputDialogHeight, inputDialogWidth);       // drawBox
        surface.e(boxX, inputDialogWidth, boxY, 27785, inputDialogHeight, 0xFFFFFF);       // drawBoxEdge   [var1^-27812 == 27785]
        int lineH = surface.a(1, 1);            // text height   [var1+508305395 == 508305352]
        int btnH = surface.a(4, 4);
        int step = lineH + 2;
        for (int n = 0; n < inputDialogLines.length; n++) {
            surface.a(256, inputDialogLines[n], 0xFFFF00, 0, 1, step * n + (5 + boxY) - -lineH);
        }
        if (inputDialogMask) {                   // Bd: password-style masking
            surface.a(256, inputTextCurrent + "*", 0xFFFFFF, 0, 4, boxY + (5 + step * inputDialogLines.length) - (-3 + -btnH));
        }

        // The Ok/Cancel buttons and the click-outside dismiss only run for the c(-43) call.
        if (var1 != -43) {
            return;
        }
        int btnY = lineH + (8 + boxY) - (-(inputDialogLines.length * step) + (-btnH - 2));
        // "Ok" button (left @ x=230..248).
        int colour = 0xFFFFFF;
        if (mouseX > 230 && mouseX < 248 && btnY - lineH < mouseY && btnY > mouseY) {
            colour = 0xFFFF00;
            if (mouseButtonClick != 0) {
                inputDialogConfirmed = true;     // vk: latch confirm
                mouseButtonClick = 0;
                inputTextFinal = inputTextCurrent;
            }
        }
        surface.a(STRINGS[122], 230, btnY, colour, false, 1);   // "Ok"
        // "Cancel" button (right @ x=264..304).
        colour = 0xFFFFFF;
        if (mouseX > 264 && mouseX < 304 && btnY - lineH < mouseY && btnY > mouseY) {
            colour = 0xFFFF00;
            if (mouseButtonClick != 0) {
                mouseButtonClick = 0;
                inputDialogType = 0;
            }
        }
        surface.a(STRINGS[121], 264, btnY, colour, false, 1);   // "Cancel"

        // A left-click outside the box also dismisses the dialog.
        if (mouseButtonClick == 1
                && (mouseX < boxX || mouseX > inputDialogWidth + boxX || mouseY < boxY || mouseY > inputDialogHeight + boxY)) {
            inputDialogType = 0;
            mouseButtonClick = 0;
        }
    }

    /**
     * The real per-tick game logic (one call per client tick, invoked as J(0)).
     * Drives the connection keep-alive, the logout/combat/idle timers, player &
     * NPC movement interpolation along their waypoint buffers, camera
     * auto-rotate and zoom, the sleep-CAPTCHA word entry (opcode 45, SLEEP_WORD),
     * the chat message tabs / chat-command parsing, mouse-button repeat
     * acceleration, and world object animations.
     *
     * NOTE: skeleton mislabels this as "drawSleepScreen". The sleep handling is
     * only one branch; this is handleGameInput (the tick).
     */
    // obf: void J(int)   [client.HD(]   proposed (skeleton): drawSleepScreen   actual: handleGameInput / tick
    private final void handleGameInput(int magic) {
        // 1) System-update countdown (server restart timer).
        if (systemUpdate > 1) {
            systemUpdate--;
        }
        // 2) Connection keep-alive + inbound packet pump.  K(magic - 26345) == K(-26345).
        sendHeartbeat(magic + -26345);
        // 3) Logout timer.
        if (logoutTimeout > 0) {
            logoutTimeout--;
        }
        // 4) Auto-logout after long inactivity (idle > 15000, not in combat / logging out).
        if (mouseActionTimeout > 15000 && combatTimeout == 0 && logoutTimeout == 0) {
            mouseActionTimeout -= 15000;
            requestLogout(magic ^ 0);           // B(0)
            return;
        }
        // 5) Local-player combat state: anim 8/9 means fighting -> hold combat timer high.
        if (localPlayer.animationCurrent == 8 || localPlayer.animationCurrent == 9) {
            combatTimeout = 500;
        }
        if (combatTimeout > 0) {
            combatTimeout--;
        }
        // 6) Character-design panel takes over input while open.
        //    F(86) services the panel and, on accept, sends opcode 235 (PLAYER_APPEARANCE_CHANGE).
        if (showAppearanceChange) {             // Kg
            sendAppearance(86);                 // F(86)
            return;
        }

        // 7) Interpolate nearby players toward their next waypoint and tick their timers.
        //    GameCharacter (ta) fields: o=waypointCurrent, e=movingStep, y=animationCurrent,
        //    D=animationNext, i=currentX, K=currentY, k[]=waypointsX, F[]=waypointsY, x=stepCount,
        //    E=messageTimeout, d=bubbleTimeout, I=combatTimer, w=projectileRange.
        for (int i = 0; i < playersLastCount; i++) {        // Yc over rg
            GameCharacter c = playersLast[i];
            int target = (c.waypointCurrent + 1) % 10;
            if (c.movingStep == target) {
                c.animationCurrent = c.animationNext;
            } else {
                int facing = -1;
                int step = c.movingStep;
                int remaining = (step < target) ? (target - step) : ((10 + target) - step);
                int speed = 4;
                if (remaining > 2) {
                    speed = (remaining - 1) * 4;
                }
                // Snap if the next waypoint is too far (teleport) or too many steps queued.
                if (c.waypointsX[step] - c.currentX > magicLoc * 3 || c.waypointsY[step] - c.currentY > magicLoc * 3
                        || c.waypointsX[step] - c.currentX < -magicLoc * 3 || c.waypointsY[step] - c.currentY < -magicLoc * 3
                        || remaining > 8) {
                    c.currentX = c.waypointsX[step];
                    c.currentY = c.waypointsY[step];
                } else {
                    if (c.currentX < c.waypointsX[step]) {
                        facing = 2; c.currentX += speed; c.stepCount++;
                    } else if (c.currentX > c.waypointsX[step]) {
                        c.stepCount++; facing = 6; c.currentX -= speed;
                    }
                    if (c.currentX - c.waypointsX[step] < speed && c.currentX - c.waypointsX[step] > -speed) {
                        c.currentX = c.waypointsX[step];
                    }
                    if (c.currentY < c.waypointsY[step]) {
                        facing = (facing == -1) ? 4 : (facing == 2 ? 3 : 5);
                        c.currentY += speed; c.stepCount++;
                    } else if (c.currentY > c.waypointsY[step]) {
                        c.stepCount++; c.currentY -= speed;
                        facing = (facing == -1) ? 0 : (facing == 2 ? 1 : 7);
                    }
                    if (c.currentY - c.waypointsY[step] < speed && c.currentY - c.waypointsY[step] > -speed) {
                        c.currentY = c.waypointsY[step];
                    }
                }
                if (facing != -1) {
                    c.animationCurrent = facing;
                }
                if (c.currentX == c.waypointsX[step] && c.currentY == c.waypointsY[step]) {
                    c.movingStep = (step + 1) % 10;
                }
            }
            if (c.messageTimeout > 0) c.messageTimeout--;
            if (c.bubbleTimeout > 0) c.bubbleTimeout--;
            if (c.combatTimer > 0) c.combatTimer--;
            // Death-screen respawn message (decremented inside this loop, matching the oracle).
            if (deathScreenTimeout > 0) {
                deathScreenTimeout--;
                if (deathScreenTimeout == 0) {
                    showServerMessage(STRINGS[629], 3);   // "You have been granted another life..."
                }
                if (deathScreenTimeout == 0) {
                    showServerMessage(STRINGS[628], 3);   // "You retain your skills..."
                }
            }
        }

        // 8) Interpolate nearby NPCs likewise (NPC id 43 spins continuously while idle).
        for (int i = 0; i < npcsLastCount; i++) {           // de over Tb
            GameCharacter c = npcsLast[i];
            int target = (c.waypointCurrent + 1) % 10;
            if (c.movingStep == target) {
                if (c.npcId == 43) {
                    c.stepCount++;
                }
                c.animationCurrent = c.animationNext;
            } else {
                int facing = -1;
                int step = c.movingStep;
                int remaining = (step < target) ? (target - step) : ((10 + target) - step);
                int speed = 4;
                if (remaining > 2) {
                    speed = (remaining - 1) * 4;
                }
                if (c.waypointsX[step] - c.currentX > magicLoc * 3 || c.waypointsY[step] - c.currentY > magicLoc * 3
                        || c.waypointsX[step] - c.currentX < -magicLoc * 3 || c.waypointsY[step] - c.currentY < -magicLoc * 3
                        || remaining > 8) {
                    c.currentX = c.waypointsX[step];
                    c.currentY = c.waypointsY[step];
                } else {
                    if (c.currentX < c.waypointsX[step]) {
                        c.stepCount++; c.currentX += speed; facing = 2;
                    } else if (c.currentX > c.waypointsX[step]) {
                        facing = 6; c.stepCount++; c.currentX -= speed;
                    }
                    if (c.currentX - c.waypointsX[step] < speed && c.currentX - c.waypointsX[step] > -speed) {
                        c.currentX = c.waypointsX[step];
                    }
                    if (c.currentY < c.waypointsY[step]) {
                        facing = (facing == -1) ? 4 : (facing == 2 ? 3 : 5);
                        c.currentY += speed; c.stepCount++;
                    } else if (c.currentY > c.waypointsY[step]) {
                        facing = (facing == -1) ? 0 : (facing == 2 ? 1 : 7);
                        c.currentY -= speed; c.stepCount++;
                    }
                    if (c.currentY - c.waypointsY[step] < speed && c.currentY - c.waypointsY[step] > -speed) {
                        c.currentY = c.waypointsY[step];
                    }
                }
                if (facing != -1) {
                    c.animationCurrent = facing;
                }
                if (c.currentX == c.waypointsX[step] && c.currentY == c.waypointsY[step]) {
                    c.movingStep = (step + 1) % 10;
                }
            }
            if (c.bubbleTimeout > 0) c.bubbleTimeout--;
            if (c.messageTimeout > 0) c.messageTimeout--;
            if (c.combatTimer > 0) c.combatTimer--;
        }

        // 9) Sleep-word delay bookkeeping (key-activity counters off the sleep tab).
        //    obf: nb.g = DataStore key-typed counter, da.M = ClientStream special-key counter.
        if (showUiTab != 2) {
            if (DataStore.g > 0) sleepWordDelayTimer++;
            if (ClientStream.M > 0) sleepWordDelayTimer = 0;
            DataStore.g = 0;
            ClientStream.M = 0;
        }
        // Tick projectile ranges on players.
        for (int i = 0; i < playersLastCount; i++) {
            GameCharacter c = playersLast[i];
            if (c.projectileRange > 0) c.projectileRange--;
        }
        if (sleepWordDelayTimer > 20) {
            sleepWordDelayTimer = 0;
            sleepWordDelay = false;
        }

        // 10) Camera smooth-follow + auto-rotate of the local player.
        //     clean: if (!Td) { snap; autorotate; followY; followX } else { snap }.
        //     Td == cameraAutoAngleDebug (when set, only the hard snap happens).
        if (!cameraAutoAngleDebug) {
            if (Math.abs(cameraFollowX - localPlayer.currentX) > 500 || Math.abs(cameraFollowY - localPlayer.currentY) > 500) {
                cameraFollowX = localPlayer.currentX;
                cameraFollowY = localPlayer.currentY;
            }
            if (optionCameraModeAuto) {
                int target = cameraAngle * 32;
                int delta = target - cameraRotation;
                int dir = 1;
                if (delta != 0) {
                    cameraRotateSpeed++;
                    if (delta > 128) {
                        dir = -1;
                        delta = 256 - delta;
                    } else if (delta > 0) {
                        dir = 1;
                    } else if (delta < -128) {
                        delta = 256 + delta;
                        dir = 1;
                    } else if (delta < 0) {
                        dir = -1;
                        delta = -delta;
                    }
                    cameraRotation += ((delta * cameraRotateSpeed + 255) / 256) * dir;
                    cameraRotation &= 255;
                } else {
                    cameraRotateSpeed = 0;
                }
            }
            if (localPlayer.currentY != cameraFollowY) {
                cameraFollowY += (localPlayer.currentY - cameraFollowY) / ((cameraZoom - 500) / 15 + 16);
            }
            if (localPlayer.currentX != cameraFollowX) {
                cameraFollowX += (localPlayer.currentX - cameraFollowX) / ((cameraZoom - 500) / 15 + 16);
            }
        } else if (cameraFollowX - localPlayer.currentX < -500 || cameraFollowX - localPlayer.currentX > 500
                || cameraFollowY - localPlayer.currentY < -500 || cameraFollowY - localPlayer.currentY > 500) {
            cameraFollowX = localPlayer.currentX;
            cameraFollowY = localPlayer.currentY;
        }

        if (!isSleeping) {                          // clean: if (!Qk)  (Qk = isSleeping)
            // 11) Chat message tab strip along the bottom of the screen.
            //     I=mouseX, xb=mouseY, Qb=mouseLastButton, Bb=mouseButton, Oi=gameHeight,
            //     Zh=messageTabSelected, yd=panelMessageTabs, yd.j[]=controlFlashText.
            if (mouseY > gameHeight - 4) {
                if (mouseX > 15 && mouseX < 96 && mouseLastButton == 1) {
                    messageTabSelected = 0;
                }
                if (mouseX > 110 && mouseX < 194 && mouseLastButton == 1) {
                    messageTabSelected = 1;
                    panelMessageTabs.flashText[controlListChat] = 999999;
                }
                if (mouseX > 215 && mouseX < 295 && mouseLastButton == 1) {
                    messageTabSelected = 2;
                    panelMessageTabs.flashText[controlListQuest] = 999999;
                }
                if (mouseX > 315 && mouseX < 395 && mouseLastButton == 1) {
                    messageTabSelected = 3;
                    panelMessageTabs.flashText[controlListPrivate] = 999999;
                }
                if (mouseX > 417 && mouseX < 497 && mouseLastButton == 1) {
                    // FIX: clean (rev 235) sets only these three; the rev-204 `reportAbuseOffence = 0` was removed.
                    inputTextFinal = "";
                    showDialogReportAbuseStep = 1;
                    inputTextCurrent = "";
                }
                mouseLastButton = 0;
                mouseButton = 0;
            }
            // handleMouse(mouseButton, mouseY, junk, mouseLastButton, mouseX) on the chat-tabs panel.
            // obf: yd.b(Bb, xb, magic-9989, Qb, I).
            panelMessageTabs.b(mouseButton, mouseY, magic + -9989, mouseLastButton, mouseX);

            if (messageTabSelected > 0 && mouseX >= 494 && mouseY >= gameHeight - 66) {
                mouseLastButton = 0;
            }

            // 12) A chat line was entered -> parse "::" commands or send as chat.
            if (panelMessageTabs.a((byte) -128, controlListInput)) {     // isClicked(bh)
                String text = panelMessageTabs.g(controlListInput, 4);   // getText
                panelMessageTabs.a(controlListInput, "", 27642);         // updateText("")
                if (text.startsWith(STRINGS[627])) {                     // "::"
                    // hj = appletMode; these debug commands are disabled in applet mode.
                    if (text.equalsIgnoreCase(STRINGS[626]) && !appletMode) {        // "::logout"
                        // FIX: clean sends a(true, 31) (sendConfirmLogoutAck), NOT closeConnection.
                        sendConfirmLogoutAck(true, magic ^ 31);   // a(true, 31)
                    } else if (text.equalsIgnoreCase(STRINGS[630]) && !appletMode) { // "::lostcon"
                        closeConnection(116);               // u(116)
                    } else if (text.equalsIgnoreCase(STRINGS[623]) && !appletMode) { // "::closecon"
                        clientStream.a(true);               // closeStream()
                    } else {
                        sendCommand(text.substring(2), 120); // opcode 38 (COMMAND): "::" command
                    }
                } else {
                    sendChatMessage(text, magic + 216);     // b(...) -> chat send
                }
            }

            // 13) Decay the chat-message fade timers (100-slot ring, ImageLoader.g[]).
            for (int i = 0; i < 100; i++) {
                if (messageHistoryTimeout[i] > 0) messageHistoryTimeout[i]--;
            }
            if (deathScreenTimeout != 0) {              // rk != 0
                mouseLastButton = 0;                   // Qb = 0
            }

            // 14) Trade/duel quantity buttons: accelerate the increment the longer held.
            //     Ti=mouseButtonDownTime, Tk=mouseButtonItemCountIncrement, Bb=mouseButton.
            if (!showDialogTrade && !showDialogDuel) {  // !Hk && !Pj
                mouseButtonDownTime = 0;
                mouseButtonItemCountIncrement = 0;
            } else {
                if (mouseButton == 0) {
                    mouseButtonDownTime = 0;
                } else {
                    mouseButtonDownTime++;
                }
                if (mouseButtonDownTime <= 600) {
                    if (mouseButtonDownTime > 450) {
                        mouseButtonItemCountIncrement += 500;
                    } else if (mouseButtonDownTime > 300) {
                        mouseButtonItemCountIncrement += 50;
                    } else if (mouseButtonDownTime <= 150) {
                        if (mouseButtonDownTime <= 50) {
                            if (mouseButtonDownTime > 20 && (mouseButtonDownTime & 5) == 0) {
                                mouseButtonItemCountIncrement++;
                            }
                        } else {
                            mouseButtonItemCountIncrement++;          // 50 < t <= 150
                        }
                    } else {
                        mouseButtonItemCountIncrement += 5;           // 150 < t <= 300
                    }
                } else {
                    mouseButtonItemCountIncrement += 5000;            // t > 600
                }
            }

            // 15) Latch this tick's click (1 = left, 2 = right) for the UI handlers.
            if (mouseLastButton == 1) {                 // ~Qb == -2
                mouseButtonClick = 1;
            }
            if (mouseLastButton == 2) {                 // ~Qb == -3
                mouseButtonClick = 2;
            }
            scene.a(0, mouseX, mouseY);             // Ek.a(0, mouseX, mouseY): setMouseLoc (Ek = Scene)
            mouseLastButton = 0;                    // Qb = 0

            // 16) Camera angle via arrow keys (auto mode steps the discrete 8-way angle,
            //     manual mode nudges the continuous rotation). Z=keyLeft, E=keyRight,
            //     si=cameraAngle, ug=cameraRotation, zf=fogOfWar, Wc=cameraRotateSpeed.
            if (optionCameraModeAuto) {                 // Kh
                if (cameraRotateSpeed == 0 || cameraAutoAngleDebug) {   // !(Wc!=0 && !Td)
                    if (keyLeft) {
                        keyLeft = false;
                        cameraAngle = cameraAngle + 1 & 7;
                        if (!fogOfWar) {
                            if ((cameraAngle & 1) == 0) cameraAngle = 1 + cameraAngle & 7;
                            for (int i = 0; i < 8; i++) {
                                if (isValidCameraAngle((byte) -125, cameraAngle)) break;
                                cameraAngle = 1 + cameraAngle & 7;
                            }
                        }
                    }
                    if (keyRight) {
                        keyRight = false;
                        cameraAngle = 7 + cameraAngle & 7;
                        if (!fogOfWar) {
                            if ((cameraAngle & 1) == 0) cameraAngle = cameraAngle + 7 & 7;
                            for (int i = 0; i < 8; i++) {
                                if (isValidCameraAngle((byte) -116, cameraAngle)) break;
                                cameraAngle = cameraAngle + 7 & 7;
                            }
                        }
                    }
                }
            } else {
                if (keyLeft) {
                    cameraRotation = 0xFF & cameraRotation + 2;
                }
                if (keyRight) {
                    cameraRotation = 0xFF & -2 + cameraRotation;
                }
            }

            // 17) Decay the minimap click-walk step counter toward zero (xh = mouseClickXStep).
            if (mouseClickXStep > 0) {
                mouseClickXStep--;
            } else if (mouseClickXStep < 0) {
                mouseClickXStep++;
            }

            // 18) Camera zoom drifts in (in fog-of-war / wilderness) or out otherwise (ac=cameraZoom).
            if (fogOfWar && cameraZoom > 550) {
                cameraZoom -= 4;
            } else if (!fogOfWar && cameraZoom < 750) {
                cameraZoom += 4;
            }

            // 19) Animated world scenery.
            scene.d(25013, 17);                     // Ek.d(25013, 17): animate fountain (model id 17)
            objectAnimationCount++;                 // qk
            if (objectAnimationCount > 5) {
                objectAnimationCount = 0;
                objectAnimationTorch = (objectAnimationTorch + 1) % 4;   // Nc %4
                objectAnimationFire = (objectAnimationFire + 1) % 3;     // Mg %3
                objectAnimationClaw = (objectAnimationClaw + 1) % 5;     // pj %5
            }
            for (int i = 0; i < objectCount; i++) {
                int ox = objectX[i];                // ye
                int oy = objectY[i];                // Se
                if (oy >= 0 && ox >= 0 && oy < 96 && ox < 96 && objectId[i] == 74) {
                    objectModel[i].f(0, -31616, 0, 1);  // hg[i].f(...): rotate windmill sails (yaw += 1)
                }
            }

            // 20) Age out expired teleport "bubble" effects (compacting the parallel arrays).
            for (int i = 0; i < teleportBubbleCount; i++) {
                teleportBubbleTime[i]++;
                if (teleportBubbleTime[i] > 50) {
                    teleportBubbleCount--;
                    for (int j = i; j < teleportBubbleCount; j++) {
                        teleportBubbleX[j] = teleportBubbleX[j + 1];
                        teleportBubbleY[j] = teleportBubbleY[j + 1];
                        teleportBubbleTime[j] = teleportBubbleTime[j + 1];
                        teleportBubbleType[j] = teleportBubbleType[j + 1];
                    }
                }
            }
        } else {
            // 21) While asleep (Qk): handle the sleep-word CAPTCHA submit (opcode 45, SLEEP_WORD).
            //     Protocol (Payload235): putByte(delayFlag) THEN putString(word).
            if (inputTextFinal.length() > 0) {
                if (inputTextFinal.equalsIgnoreCase(STRINGS[630]) && !appletMode) {        // "::lostcon"
                    clientStream.a(true);               // closeStream()
                } else if (inputTextFinal.equalsIgnoreCase(STRINGS[623]) && !appletMode) { // "::closecon"
                    // FIX: clean sends a(true, 31) (sendConfirmLogoutAck), NOT closeConnection.
                    sendConfirmLogoutAck(true, magic + 31);   // a(true, 31)
                } else {
                    clientStream.b(45, 0);              // opcode 45 (SLEEP_WORD)
                    // FIX: delay byte is written FIRST (1 if delay engaged, else 0), then the word.
                    if (sleepWordDelay) {
                        clientStream.f.c(1, -75);       // putByte(1)
                    } else {
                        clientStream.f.c(0, -100);      // putByte(0)
                        sleepWordDelay = true;
                    }
                    clientStream.f.a(inputTextFinal, 116);  // putString(word)
                    clientStream.b(21294);
                    inputTextCurrent = "";
                    sleepingStatusText = STRINGS[436];  // "Please wait..."
                    inputTextFinal = "";
                }
            }
            // Clicking the "type the word" box submits "-null-".
            if (mouseLastButton == 1 && mouseY > 275 && mouseY < 310 && mouseX > 56 && mouseX < 456) {
                clientStream.b(45, 0);                  // opcode 45 (SLEEP_WORD)
                // FIX: write the delay byte first (0 the first time, 1 thereafter), then the word.
                if (!sleepWordDelay) {
                    clientStream.f.c(0, 35);            // putByte(0)
                    sleepWordDelay = true;
                } else {
                    clientStream.f.c(1, 123);           // putByte(1)
                }
                clientStream.f.a(STRINGS[625], magic ^ -74);    // putString("-null-")
                clientStream.b(21294);
                sleepingStatusText = STRINGS[436];      // "Please wait..."
                inputTextFinal = "";
                inputTextCurrent = "";
            }
            mouseLastButton = 0;
        }
    }
