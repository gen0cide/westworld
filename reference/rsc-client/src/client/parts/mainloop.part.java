    // ===== mainloop =====
    //
    // Per-tick game logic, connection keep-alive, logout flow, and the three
    // top-right UI tabs whose obfuscated bodies the skeleton grouped under
    // "mainloop".
    //
    // IMPORTANT SKELETON-MISLABEL NOTES (verified against mudclient204 oracle):
    //   * The skeleton calls a(boolean,byte)@13050 "tick"            -> it is actually
    //     the MINIMAP tab renderer (drawUiTabMinimap).
    //   * The skeleton calls b(boolean,byte)@11918 "updateGameState" -> it is actually
    //     the MAGIC/PRAYER tab renderer (drawUiTabMagic).
    //   * The skeleton calls J(int)@26226 "drawSleepScreen"          -> it is actually
    //     the REAL per-tick update (handleGameInput); the sleep-CAPTCHA send is just
    //     one branch of it.
    // Methods are named for what they actually do; obf signatures are kept in comments.

    /**
     * GameShell stop hook. Tears down the game session: confirms logout to the
     * server (opcode 31) and stops the active sound voice.
     *
     * @param fromShell true when invoked by the shell's stop path (sets the tick marker).
     */
    // obf: void a(boolean)   [client.SA(]   proposed: onStopGame
    final void onStopGame(boolean fromShell) {
        if (fromShell) {
            tickMarker = -103L;                 // ze: scratch timing marker
        }
        // sendConfirmLogoutAck(true, 31) -> opcode 31 (CONFIRM_LOGOUT)
        sendConfirmLogoutAck(true, 31);
        if (soundChannel == null) {             // ni: active audio voice
            return;
        }
        soundChannel.d();                       // stop/close the sound channel
    }

    /**
     * Send a server keep-alive and pump one inbound packet. Called once per tick
     * by {@link #handleGameInput}. If no packet has been received for >5s a PING
     * (opcode 67) is sent; then any pending writes are flushed and one inbound
     * packet is read and dispatched.
     */
    // obf: void K(int)   [client.SB(]   proposed: sendHeartbeat
    private final void sendHeartbeat(int magic) {
        long now = Timer.a(0);                  // p.a(0) = System.currentTimeMillis()

        // Wi here = packetLastRead timestamp (activity timer reused for net liveness).
        if (clientStream.a((byte) 34)) {        // Jh.a(34) = hasPacket(): data arrived
            lastActionTime = now;
        }
        // ~(now - lastActionTime) >= -5001  <=>  (now - lastActionTime) > 5000ms idle
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
        int size = clientStream.a(0, incomingPacket);   // readPacket(incomingPacket) (magic+26345==0)
        if (size <= 0) {
            return;
        }
        // Dispatch one server->client packet. First payload byte (de-ISAAC'd) is the opcode.
        handlePacket(23, size, incomingPacket.a((byte) 104));
    }

    /**
     * Request a normal logout (opcode 102, LOGOUT). Refused while in / shortly
     * after combat. On success an internal logout timer is armed so the
     * "Logging out..." dialog shows until the server drops us.
     */
    // obf: void B(int)   [client.T(]   proposed: requestLogout
    private final void requestLogout(int unused) {
        if (loggedIn == 0) {                    // ~qg: only when logged in
            return;
        }
        if (combatTimeout > 450) {              // ai: combat countdown
            showServerMessage(STRINGS[421], 3); // "@cya@You can't logout during combat!"
            return;
        }
        if (combatTimeout > 0) {                // within 10s grace after combat
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

        // Rotation/zoom for the minimap. sd=minimapRandom (zoom jitter), ug=cameraRotation,
        // Df=minimapRandom rotation offset. cc = Scene.sin2048Cache (fixed-point sin/cos).
        // GameCharacter (ta) accessors used below: .currentX=obf .i, .currentY=obf .K, .hash=obf .C.
        int zoom = 192 + minimapRandom2;
        int rot = (cameraRotation + minimapRandom1) & 255;
        int px = (localPlayer.currentX - 6040) * zoom * 3 / 2048;
        int py = (localPlayer.currentY - 6040) * 3 * zoom / 2048;
        int sinR = SurfaceSprite.cc[(1024 - 4 * rot) & 0x3ff];
        int cosR = SurfaceSprite.cc[((1024 - 4 * rot) & 0x3ff) + 1024];
        int rx = px * cosR + py * sinR >> 18;   // >>18: 2048*2048 -> divide back (junk shift masked to 18)
        py = -(px * sinR) + py * cosR >> 18;    // (2D rotate the point by -rot)
        px = rx;
        // landscape sprite, rotated/scaled
        surface.a(uiX - 1, 36 - (-(uiHeight / 2) + -py), uiWidth / 2 + uiX - px, 842218000, zoom, (64 + rot) & 255);

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
                for (int f = 0; f < FontWidths.g; f++) {         // friends list count
                    boolean isFriend = name.equals(WorldEntity.a(Surface.h[f], (byte) 107));
                    if (isFriend && (keyState[f] & 2) != 0) {    // friend online
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
            int wx = ((mouseX - (base + uiWidth / 2)) * 16384) / (z * 3);
            int wy = ((mouseY - (36 + uiHeight / 2)) * 16384) / (z * 3);
            int s2 = SurfaceSprite.cc[(1024 - r * 4) & 0x3ff];
            int c2 = SurfaceSprite.cc[((1024 - r * 4) & 0x3ff) + 1024];
            int rwx = wy * s2 + wx * c2 >> 15;
            wy = wy * c2 - wx * s2 >> 15;
            wx = rwx + localPlayer.currentX;
            wy = localPlayer.currentY - wy;
            if (mouseButtonClick == 1) {        // Cf == 1: a fresh left-click this tick
                // obf arg order: a(worldY>>7, worldX>>7, localRegionX, localRegionY, false, 8)
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
     */
    // obf: void b(boolean,byte)   proposed (skeleton): updateGameState   actual: drawUiTabMagic
    private final void drawUiTabMagic(boolean handleMenus, byte unused) {
        int uiX = -199 + surface.u;
        int uiY = 36;
        surface.b(-1, spriteMedia + 4, 3, -49 + uiX);        // drawSprite tab background
        int uiWidth = 196;
        int uiHeight = 182;

        // Highlight the active sub-tab header brighter (220,220,220) than the inactive (160).
        int magicShade, prayerShade;
        magicShade = prayerShade = ISAAC.a(160, 9570, 160, 160);     // o.a(...) = Surface.rgb2long
        if (tabMagicPrayer == 0) {
            magicShade = ISAAC.a(220, 9570, 220, 220);
        } else {
            prayerShade = ISAAC.a(220, 9570, 220, 220);
        }
        surface.c(128, uiX, 24, 0, uiY, uiWidth / 2, magicShade);
        surface.c(128, uiWidth / 2 + uiX, 24, 0, uiY, uiWidth / 2, prayerShade);
        surface.c(128, uiX, 90, uiY + 24, uiWidth, ISAAC.a(220, 9570, 220, 220));
        surface.c(128, uiX, uiHeight - 24 - 90, uiY + 24 + 90, uiWidth, ISAAC.a(160, 9570, 160, 160));
        surface.b(uiWidth, 0, uiX, uiY + 24, (byte) 70);     // drawLineHoriz under headers
        surface.b(uiX - -(uiWidth / 2), uiY, 0, 24, 0);      // drawLineVert between headers
        surface.b(uiWidth, 0, uiX, uiY + 113, (byte) -92);   // drawLineHoriz under list
        surface.a(uiWidth / 4 + uiX, STRINGS[16], 0, uiY + 16, 4, 36);   // "Magic"
        surface.a(uiX + uiWidth / 4 + uiWidth / 2, STRINGS[21], 0, 0, 4, uiY + 16);   // "Prayers"

        if (tabMagicPrayer == 0) {
            // --- Spell list ---
            panelGame2d.c(controlListMagic);    // clearList
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
                panelGame2d.a(row++, controlListMagic,
                        colour + STRINGS[18] + ImageLoader.f[spell] + STRINGS[12] + BitBuffer.L[spell]);
            }
            panelGame2d.a();                     // drawPanel
            int sel = panelGame2d.b(controlListMagic, 17050);         // getListEntryIndex
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
            panelGame2d.c(controlListMagic);    // clearList
            int row = 0;
            for (int prayer = 0; prayer < EntityDef.g; prayer++) {    // prayerCount
                String colour = STRINGS[15];     // "@whi@"
                if (skillBase[5] < GameModel.B[prayer]) {             // prayerLevel > prayer base
                    colour = STRINGS[19];        // "@bla@"
                }
                if (prayerOn[prayer]) {          // bk[]: prayer currently active
                    colour = STRINGS[27];        // "@gre@"
                }
                panelGame2d.a(row++, controlListMagic,
                        colour + STRINGS[18] + GameModel.B[prayer] + STRINGS[12] + EntityDef.h[prayer]);
            }
            panelGame2d.a();                     // drawPanel
            int sel = panelGame2d.b(controlListMagic, 17124);
            if (sel != -1) {
                surface.a(uiX - -(uiWidth / 2), STRINGS[18] + GameModel.B[sel] + STRINGS[12] + EntityDef.h[sel], 0xFFFF00, 0, 1, uiY + 130);
                surface.a(uiX - -(uiWidth / 2), TextEncoder.e[sel], 0xFFFFFF, 0, 0, 145 + uiY);    // prayerDescription
                surface.a(uiX - -(uiWidth / 2), STRINGS[26] + ClientIOException.c[sel], 0, 0, 1, 160 + uiY);   // "Drain rate: "
            } else {
                surface.a(STRINGS[11], uiX + 2, uiY + 124, 0, false, 1);   // "Point at a prayer for a description"
            }
        }

        if (!handleMenus) {
            return;
        }
        int mx = mouseX - (surface.u - 199);
        int my = mouseY - 36;
        if (mx < 0 || my < 0 || mx >= 196 || my >= 182) {
            mouseButtonClick = 0;
            return;
        }
        panelGame2d.b(panelHandleArg, my + 36, mouseLastButton, mouseButton, mx + (surface.u - 199));  // handleMouse

        // Header click toggles between Magic (left) and Prayers (right).
        if (my <= 24 && mouseButtonClick == 1) {
            if (mx < 98 && tabMagicPrayer == 1) {
                tabMagicPrayer = 0;
                panelGame2d.e(controlListMagic);    // resetListProps
            } else if (mx > 98 && tabMagicPrayer == 0) {
                tabMagicPrayer = 1;
                panelGame2d.e(controlListMagic);
            }
        }

        // Click a spell -> select it (level + rune checks first).
        if (mouseButtonClick == 1 && tabMagicPrayer == 0) {
            int sel = panelGame2d.b(controlListMagic, 17050);
            if (sel != -1) {
                if (ImageLoader.f[sel] > skillCurrent[6]) {
                    showServerMessage(STRINGS[25], 3);   // "Your magic ability is not high enough for this spell"
                } else {
                    int rune;
                    for (rune = 0; rune < ISAAC.p[sel]; rune++) {
                        int runeId = NameHash.d[sel][rune];
                        if (!hasInventoryItems(ClientStream.J[sel][rune], runeId)) {
                            showServerMessage(STRINGS[24], 3);   // "You don't have all the reagents you need for this spell"
                            rune = -1;
                            break;
                        }
                    }
                    if (rune == ISAAC.p[sel]) {
                        selectedSpell = sel;
                        selectedItemInventoryIndex = -1;
                    }
                }
            }
        }

        // Click a prayer -> toggle it; sends PRAYER_ACTIVATED (60) / PRAYER_DEACTIVATED (254).
        if (mouseButtonClick == 1 && tabMagicPrayer == 1) {
            int sel = panelGame2d.b(controlListMagic, 17050);
            if (sel != -1) {
                if (GameModel.B[sel] > skillBase[5]) {
                    showServerMessage(STRINGS[23], 3);   // "Your prayer ability is not high enough for this prayer"
                } else if (skillCurrent[5] == 0) {
                    showServerMessage(STRINGS[28], 3);   // "You have run out of prayer points..."
                } else if (prayerOn[sel]) {
                    clientStream.b(254, 0);              // opcode 254 (PRAYER_DEACTIVATED)
                    clientStream.f.c(sel, 37);          // putByte(prayerId)
                    clientStream.b(21294);              // flush
                    prayerOn[sel] = false;
                    playSoundFile(STRINGS[17]);         // "prayeroff"
                } else {
                    clientStream.b(60, 0);              // opcode 60 (PRAYER_ACTIVATED)
                    clientStream.f.c(sel, 57);          // putByte(prayerId)
                    clientStream.b(21294);              // flush
                    prayerOn[sel] = true;
                    playSoundFile(STRINGS[22]);         // "prayeron"
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
     * NOTE: skeleton calls this "sendQueuedActions"; it also draws the dialog box.
     */
    // obf: void c(byte)   proposed: sendQueuedActions
    private final void sendQueuedActions(byte unused) {
        // --- Confirmation path: a value has been submitted (Cb non-empty) or OK was
        //     latched last tick (vk = inputDialogConfirmed). Otherwise fall through to render. ---
        if (inputTextFinal.length() > 0 || inputDialogConfirmed) {
            String value = inputTextFinal.trim();
            inputTextCurrent = "";
            inputTextFinal = "";

            // gc (inputDialogType) selects which queued action to flush. The bare a()/b()/c()
            // calls are intra-class wrappers that build their own packets.
            // ae[Rd] = item id of the selected bank slot; Rj[Di]/Jf[Di] = id/price of the
            // selected shop slot.
            switch (inputDialogType) {
                case 1: // generic "enter amount" -> a(amount, 9, itemId) wrapper (e.g. drop/quantity)
                    try {
                        sendItemAction(Integer.parseInt(value), (byte) 9, dialogItemId);
                    } catch (NumberFormatException ignored) {
                    }
                    break;
                case 2: // -> c(amount, 124, itemId) wrapper
                    try {
                        sendItemActionAlt(Integer.parseInt(value), (byte) 124, dialogItemId);
                    } catch (NumberFormatException ignored) {
                    }
                    break;
                case 4: { // Bank deposit: opcode 23 (BANK_DEPOSIT).
                    try {
                        int itemId = (bankSelectedSlot >= 0) ? bankItems[bankSelectedSlot] : -1;
                        int amount = Integer.parseInt(value);
                        clientStream.b(23, 0);
                        clientStream.f.e(436, itemId);          // putShort(itemId)
                        clientStream.f.b(0, amount);            // putInt(amount)
                        clientStream.f.b(0, 0x87654321);        // putInt(magic/checksum)
                        clientStream.b(21294);
                    } catch (NumberFormatException ignored) {
                    }
                    break;
                }
                case 6: // Shop sell: opcode 221 (SHOP_SELL).
                    try {
                        if (shopSelectedItemId[shopSelectedSlot] != 0) {
                            int amount = Integer.parseInt(value);
                            clientStream.b(221, 0);
                            clientStream.f.e(393, shopSelectedItemId[shopSelectedSlot]);   // item id
                            clientStream.f.e(393, shopSelectedItemPrice[shopSelectedSlot]); // price
                            clientStream.f.e(436, amount);
                            clientStream.b(21294);
                        }
                    } catch (NumberFormatException ignored) {
                    }
                    break;
                case 7: // -> b(109, amount, itemId) wrapper
                    try {
                        sendItemActionB(109, Integer.parseInt(value), dialogItemId2);
                    } catch (NumberFormatException ignored) {
                    }
                    break;
                case 8: // -> a(itemId, amount, -78) wrapper
                    try {
                        sendItemActionC(dialogItemId2, Integer.parseInt(value), (byte) -78);
                    } catch (NumberFormatException ignored) {
                    }
                    break;
                case 9: // Skip tutorial: opcode 84 (SKIP_TUTORIAL), no payload.
                    clientStream.b(84, 0);
                    clientStream.b(21294);
                    break;
                case 5: // Shop buy: opcode 236 (SHOP_BUY).
                    try {
                        if (shopSelectedItemId[shopSelectedSlot] != 0) {
                            int amount = Integer.parseInt(value);
                            clientStream.b(236, 0);
                            clientStream.f.e(420, shopSelectedItemId[shopSelectedSlot]);   // item id (436^... -> 420)
                            clientStream.f.e(393, shopSelectedItemPrice[shopSelectedSlot]); // price
                            clientStream.f.e(393, amount);
                            clientStream.b(21337);
                        }
                    } catch (NumberFormatException ignored) {
                    }
                    break;
                case 3: // Bank withdraw: opcode 22 (BANK_WITHDRAW).
                default:
                    try {
                        int itemId = (bankSelectedSlot >= 0) ? bankItems[bankSelectedSlot] : -1;
                        int amount = Integer.parseInt(value);
                        clientStream.b(22, 0);
                        clientStream.f.e(393, itemId);
                        clientStream.f.b(0, amount);
                        clientStream.f.b(0, 0x12345678);
                        clientStream.b(21294);
                    } catch (NumberFormatException ignored) {
                    }
                    break;
            }
            inputDialogType = 0;
            return;
        }

        // --- Render path (still waiting on input) ---
        // For numeric dialog kinds (gc 1..8) strip any non-digit chars from the live text.
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

        // Draw the modal box, the prompt lines, and OK / Cancel.
        int boxX = 256 - inputDialogWidth / 2;
        int boxY = 180 - inputDialogHeight / 2;
        surface.a(boxX, (byte) -103, 0, boxY, inputDialogHeight, inputDialogWidth);       // drawBox
        surface.e(boxX, inputDialogWidth, boxY, 27812, inputDialogHeight, 0xFFFFFF);       // drawBoxEdge
        int lineH = surface.a(1, 1);            // text height
        int btnH = surface.a(4, 4);
        int step = lineH + 2;
        for (int n = 0; n < inputDialogLines.length; n++) {
            surface.a(256, inputDialogLines[n], 0xFFFF00, 1, 1, step * n + (5 + boxY) - -lineH);
        }
        if (inputDialogMask) {                   // Bd: password-style masking
            surface.a(256, inputTextCurrent + "*", 0xFFFFFF, 1, 4, boxY + (5 + step * inputDialogLines.length) - (-3 + -btnH));
        }

        int btnY = lineH + (8 + boxY) - (-(inputDialogLines.length * step) + (-btnH - 2));
        // "Ok" button (left @ x=230).
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
        // "Cancel" button (right @ x=264).
        colour = 0xFFFFFF;
        if (mouseX > 264 && mouseX < 304 && btnY - lineH < mouseY && btnY > mouseY) {
            colour = 0xFFFF00;
            if (mouseButtonClick != 0) {
                mouseButtonClick = 0;
                inputDialogType = 0;
            }
        }
        surface.a(STRINGS[121], 264, btnY, colour, false, 1);   // "Cancel"

        // Clicking outside the box also dismisses the dialog.
        if (mouseButtonClick != 1) {
            return;
        }
        if (mouseX >= boxX && mouseX <= inputDialogWidth + boxX && mouseY >= boxY && mouseY <= inputDialogHeight + boxY) {
            return;
        }
        inputDialogType = 0;
        mouseButtonClick = 0;
    }

    /**
     * The real per-tick game logic (one call per client tick). Drives the
     * connection keep-alive, the logout/combat/idle timers, player & NPC
     * movement interpolation along their waypoint buffers, camera auto-rotate
     * and zoom, the sleep-CAPTCHA word entry (opcode 45, SLEEP_WORD), the chat
     * message tabs / chat-command parsing, mouse-button repeat acceleration,
     * and world object animations.
     *
     * NOTE: skeleton mislabels this as "drawSleepScreen". The sleep handling is
     * only one branch; this is handleGameInput (the tick).
     */
    // obf: void J(int)   proposed (skeleton): drawSleepScreen   actual: handleGameInput / tick
    private final void handleGameInput(int magic) {
        // 1) System-update countdown (server restart timer).
        if (systemUpdate > 1) {
            systemUpdate--;
        }
        // 2) Connection keep-alive + inbound packet pump.
        sendHeartbeat(magic + -26345);          // K(0)
        // 3) Logout timer.
        if (logoutTimeout > 0) {
            logoutTimeout--;
        }
        // 4) Auto-logout after long inactivity (idle > 15000 ticks, not in combat / logging out).
        if (mouseActionTimeout > 15000 && combatTimeout == 0 && logoutTimeout == 0) {
            mouseActionTimeout -= 15000;
            requestLogout(magic);               // B(...)
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
        //    F(86) = sendAppearance (skeleton name); it both services the panel and,
        //    on accept, sends opcode 235 (PLAYER_APPEARANCE_CHANGE).
        if (showAppearanceChange) {             // Kg
            sendAppearance(86);                 // F(86)
            return;
        }

        // 7) Interpolate nearby players toward their next waypoint and tick their timers.
        //    GameCharacter (ta) fields: o=waypointCurrent, e=movingStep, y=animationCurrent,
        //    D=animationNext, i=currentX, K=currentY, k[]=waypointsX, F[]=waypointsY, x=stepCount.
        for (int i = 0; i < playersLastCount; i++) {        // Yc over rg
            GameCharacter c = playersLast[i];
            int target = (c.waypointCurrent + 1) % 10;
            if (c.movingStep != target) {
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
                        c.currentX += speed; c.stepCount++; facing = 2;
                    } else if (c.currentX > c.waypointsX[step]) {
                        c.currentX -= speed; c.stepCount++; facing = 6;
                    }
                    if (c.currentX - c.waypointsX[step] < speed && c.currentX - c.waypointsX[step] > -speed) {
                        c.currentX = c.waypointsX[step];
                    }
                    if (c.currentY < c.waypointsY[step]) {
                        c.currentY += speed; c.stepCount++;
                        facing = (facing == -1) ? 4 : (facing == 2 ? 3 : 5);
                    } else if (c.currentY > c.waypointsY[step]) {
                        c.currentY -= speed; c.stepCount++;
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
            } else {
                c.animationCurrent = c.animationNext;
            }
            if (c.messageTimeout > 0) c.messageTimeout--;
            if (c.bubbleTimeout > 0) c.bubbleTimeout--;
            if (c.combatTimer > 0) c.combatTimer--;
            if (deathScreenTimeout > 0) {
                deathScreenTimeout--;
                if (deathScreenTimeout == 0) {
                    showServerMessage(STRINGS[629], 3);   // "You have been granted another life..."
                    showServerMessage(STRINGS[628], 3);   // "You retain your skills..."
                }
            }
        }

        // 8) Interpolate nearby NPCs likewise (NPC id 43 spins continuously while idle).
        for (int i = 0; i < npcsLastCount; i++) {           // de over Tb
            GameCharacter c = npcsLast[i];
            int target = (c.waypointCurrent + 1) % 10;
            if (c.movingStep != target) {
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
                        c.currentX += speed; c.stepCount++; facing = 2;
                    } else if (c.currentX > c.waypointsX[step]) {
                        c.currentX -= speed; c.stepCount++; facing = 6;
                    }
                    if (c.currentX - c.waypointsX[step] < speed && c.currentX - c.waypointsX[step] > -speed) {
                        c.currentX = c.waypointsX[step];
                    }
                    if (c.currentY < c.waypointsY[step]) {
                        c.currentY += speed; c.stepCount++;
                        facing = (facing == -1) ? 4 : (facing == 2 ? 3 : 5);
                    } else if (c.currentY > c.waypointsY[step]) {
                        c.currentY -= speed; c.stepCount++;
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
            } else {
                c.animationCurrent = c.animationNext;
                if (c.npcId == 43) {
                    c.stepCount++;
                }
            }
            if (c.messageTimeout > 0) c.messageTimeout--;
            if (c.bubbleTimeout > 0) c.bubbleTimeout--;
            if (c.combatTimer > 0) c.combatTimer--;
        }

        // 9) Sleep-word delay bookkeeping (Surface tracks key activity off the sleep tab).
        if (showUiTab != 2) {
            if (Surface.anInt346 > 0) sleepWordDelayTimer++;
            if (Surface.anInt347 > 0) sleepWordDelayTimer = 0;
            Surface.anInt346 = 0;
            Surface.anInt347 = 0;
        }
        // Tick projectile ranges on players.
        for (int i = 0; i < playersLastCount; i++) {
            GameCharacter c = playersLast[i];
            if (c.projectileRange > 0) c.projectileRange--;
        }

        // 10) Camera auto-rotate / smooth-follow of the local player.
        if (cameraAutoAngleDebug) {
            if (Math.abs(cameraFollowX - localPlayer.currentX) > 500 || Math.abs(cameraFollowY - localPlayer.currentY) > 500) {
                cameraFollowX = localPlayer.currentX;
                cameraFollowY = localPlayer.currentY;
            }
        } else {
            if (Math.abs(cameraFollowX - localPlayer.currentX) > 500 || Math.abs(cameraFollowY - localPlayer.currentY) > 500) {
                cameraFollowX = localPlayer.currentX;
                cameraFollowY = localPlayer.currentY;
            }
            if (cameraFollowX != localPlayer.currentX) {
                cameraFollowX += (localPlayer.currentX - cameraFollowX) / (16 + (cameraZoom - 500) / 15);
            }
            if (cameraFollowY != localPlayer.currentY) {
                cameraFollowY += (localPlayer.currentY - cameraFollowY) / (16 + (cameraZoom - 500) / 15);
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
                        dir = 1;
                        delta = 256 + delta;
                    } else if (delta < 0) {
                        dir = -1;
                        delta = -delta;
                    }
                    cameraRotation += ((cameraRotateSpeed * delta + 255) / 256) * dir;
                    cameraRotation &= 255;
                } else {
                    cameraRotateSpeed = 0;
                }
            }
        }
        if (sleepWordDelayTimer > 20) {
            sleepWordDelay = false;
            sleepWordDelayTimer = 0;
        }

        // 11) While asleep: handle the sleep-word CAPTCHA submit (opcode 45, SLEEP_WORD).
        if (isSleeping) {
            if (inputTextFinal.length() > 0) {
                if (inputTextFinal.equalsIgnoreCase(STRINGS[630]) && !appletMode) {       // "::lostcon"
                    clientStream.a(true);               // closeStream()
                } else if (inputTextFinal.equalsIgnoreCase(STRINGS[623]) && !appletMode) { // "::closecon"
                    closeConnection(116);
                } else {
                    clientStream.b(45, 0);              // opcode 45 (SLEEP_WORD)
                    clientStream.f.a(inputTextFinal, 116);
                    if (!sleepWordDelay) {
                        clientStream.f.c(0, 35);        // putByte(0): include delay flag once
                        sleepWordDelay = true;
                    }
                    clientStream.b(21294);
                    inputTextCurrent = "";
                    inputTextFinal = "";
                    sleepingStatusText = STRINGS[436];  // "Please wait..."
                }
            }
            // Clicking the "type the word" box submits "-null-".
            if (mouseLastButton == 1 && mouseY > 275 && mouseY < 310 && mouseX > 56 && mouseX < 456) {
                clientStream.b(45, 0);                  // opcode 45 (SLEEP_WORD)
                clientStream.f.a(STRINGS[625], 116);    // "-null-"
                if (!sleepWordDelay) {
                    clientStream.f.c(1, 123);           // putByte
                    sleepWordDelay = true;
                }
                clientStream.b(21294);
                sleepingStatusText = STRINGS[436];      // "Please wait..."
                inputTextFinal = "";
                inputTextCurrent = "";
            }
            mouseLastButton = 0;
            return;
        }

        // 12) Chat message tab strip along the bottom of the screen.
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
                showDialogReportAbuseStep = 1;          // open report-abuse dialog
                reportAbuseOffence = 0;
                inputTextCurrent = "";
                inputTextFinal = "";
            }
            mouseLastButton = 0;
            mouseButton = 0;
        }
        panelMessageTabs.b(mouseX, mouseY, mouseLastButton, mouseButton);   // handleMouse

        if (messageTabSelected > 0 && mouseX >= 494 && mouseY >= gameHeight - 66) {
            mouseLastButton = 0;
        }

        // 13) A chat line was entered -> parse "::" commands or send as chat.
        if (panelMessageTabs.d(controlListInput)) {     // isClicked
            String text = panelMessageTabs.g(controlListInput, 4);   // getText
            panelMessageTabs.a(controlListInput, "");                // updateText("")
            if (text.startsWith(STRINGS[627])) {         // "::"
                // hj = appletMode; these debug commands are disabled in applet mode.
                if (text.equalsIgnoreCase(STRINGS[623]) && !appletMode) {        // "::closecon"
                    clientStream.a(true);               // closeStream()
                } else if (text.equalsIgnoreCase(STRINGS[626]) && !appletMode) { // "::logout"
                    closeConnection(116);               // u(116)
                } else if (text.equalsIgnoreCase(STRINGS[630]) && !appletMode) { // "::lostcon"
                    closeConnection(116);               // lostConnection path (also via u(...))
                } else {
                    sendCommand(text.substring(2), 120); // opcode 38 (COMMAND): "::" command
                }
            } else {
                sendChatMessage(text, magic + 216);     // b(...) -> opcode 4 (chat send)
            }
        }

        // 14) Tick the on-screen chat-history fade timers when the "All" tab is up.
        if (messageTabSelected == 0) {
            for (int i = 0; i < 5; i++) {
                if (messageHistoryTimeout[i] > 0) messageHistoryTimeout[i]--;
            }
        }
        if (deathScreenTimeout != 0) {
            mouseLastButton = 0;
        }

        // 15) Trade/duel quantity buttons: accelerate the increment the longer the button is held.
        if (showDialogTrade || showDialogDuel) {
            if (mouseButton != 0) {
                mouseButtonDownTime++;
            } else {
                mouseButtonDownTime = 0;
            }
            if (mouseButtonDownTime > 600) mouseButtonItemCountIncrement += 5000;
            else if (mouseButtonDownTime > 450) mouseButtonItemCountIncrement += 500;
            else if (mouseButtonDownTime > 300) mouseButtonItemCountIncrement += 50;
            else if (mouseButtonDownTime > 150) mouseButtonItemCountIncrement += 5;
            else if (mouseButtonDownTime > 50) mouseButtonItemCountIncrement++;
            else if (mouseButtonDownTime > 20 && (mouseButtonDownTime & 5) == 0) mouseButtonItemCountIncrement++;
        } else {
            mouseButtonDownTime = 0;
            mouseButtonItemCountIncrement = 0;
        }

        // 16) Latch this tick's click (1 = left, 2 = right) for the UI handlers, then consume it.
        if (mouseLastButton == 1) {
            mouseButtonClick = 1;
        } else if (mouseLastButton == 2) {
            mouseButtonClick = 2;
        }
        scene.a(mouseX, mouseY);                 // setMouseLoc
        mouseLastButton = 0;

        // 17) Camera angle via arrow keys (auto mode steps the discrete 8-way angle,
        //     manual mode nudges the continuous rotation).
        if (optionCameraModeAuto) {
            if (cameraRotateSpeed == 0 || cameraAutoAngleDebug) {
                if (keyLeft) {
                    cameraAngle = cameraAngle + 1 & 7;
                    keyLeft = false;
                    if (!fogOfWar) {
                        if ((cameraAngle & 1) == 0) cameraAngle = cameraAngle + 1 & 7;
                        for (int i = 0; i < 8; i++) {
                            if (isValidCameraAngle(cameraAngle)) break;
                            cameraAngle = cameraAngle + 1 & 7;
                        }
                    }
                }
                if (keyRight) {
                    cameraAngle = cameraAngle + 7 & 7;
                    keyRight = false;
                    if (!fogOfWar) {
                        if ((cameraAngle & 1) == 0) cameraAngle = cameraAngle + 7 & 7;
                        for (int i = 0; i < 8; i++) {
                            if (isValidCameraAngle(cameraAngle)) break;
                            cameraAngle = cameraAngle + 7 & 7;
                        }
                    }
                }
            }
        } else if (keyLeft) {
            cameraRotation = cameraRotation + 2 & 255;
        } else if (keyRight) {
            cameraRotation = cameraRotation - 2 & 255;
        }

        // 18) Camera zoom drifts in (in fog-of-war / wilderness) or out otherwise.
        if (fogOfWar && cameraZoom > 550) {
            cameraZoom -= 4;
        } else if (!fogOfWar && cameraZoom < 750) {
            cameraZoom += 4;
        }
        // Decay the minimap click-walk step counter toward zero.
        if (mouseClickXStep > 0) {
            mouseClickXStep--;
        } else if (mouseClickXStep < 0) {
            mouseClickXStep++;
        }

        // 19) Animated world scenery.
        scene.h(17);                            // animate fountain (model id 17)
        objectAnimationCount++;
        if (objectAnimationCount > 5) {
            objectAnimationCount = 0;
            objectAnimationFire = (objectAnimationFire + 1) % 3;
            objectAnimationTorch = (objectAnimationTorch + 1) % 4;
            objectAnimationClaw = (objectAnimationClaw + 1) % 5;
        }
        for (int i = 0; i < objectCount; i++) {
            int ox = objectX[i];
            int oy = objectY[i];
            if (ox >= 0 && oy >= 0 && ox < 96 && oy < 96 && objectId[i] == 74) {
                objectModel[i].rotate(1, 0, 0);     // spin windmill sails etc.
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
    }
