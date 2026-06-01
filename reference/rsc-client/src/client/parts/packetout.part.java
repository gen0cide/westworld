    // ===== packetout =====
    //
    // Outgoing (client -> server) packet builders. Opcodes are cited from the OpenRSC
    // Payload235Parser (the rev-235 client->server map). The low-level wire helpers used
    // throughout this group are:
    //
    //   clientStream.newPacket(op)             // begin a packet with opcode `op`  (b.b(int,int))
    //   clientStream.sendPacket()              // flush/queue the packet           (b.b(int))
    //   clientStream.buffer.putByte(v)         // 1 byte                           (tb.c(int,int))
    //   clientStream.buffer.putShort(v)        // 2 bytes, big-endian              (tb.e(int,int))
    //   clientStream.buffer.putInt(v)          // 4 bytes, big-endian              (tb.b(int,int))
    //   clientStream.buffer.putString(s)       // length-prefixed + null-terminated(tb.a(String,int))
    //   StringCodec.encodeAndWrite(buffer, s)  // length byte + char-table-encoded (u.a(int,tb,String))
    //
    // The original methods are wrapped in J++ control-flow obfuscation (opaque predicate
    // `client.vh`, per-method profiling `++counter`, try/catch rethrow via ErrorHandler,
    // anti-tamper guards on the dummy `byte`/`int` params, and junk masks on the buffer
    // writes). All of that is stripped below; only the real logic remains.

    /**
     * Walk to an explicit destination, optionally walking up to (rather than onto) the
     * target tile. Pathfinds with {@link World#route}, then streams the start tile plus
     * per-step deltas. Sends opcode 16 (WALK_TO_ENTITY) when {@code walkToAction} is set,
     * else opcode 187 (WALK_TO_POINT).
     *
     * @return true if a packet was sent (a route existed), false if unreachable.
     */
    // obf: private final boolean a(int,int,byte,boolean,int,int,int,int,boolean)  [byte param is anti-tamper junk]
    private boolean walkTo(int startX, int startY, byte unused, boolean checkObjects,
                           int x1, int y1, int x2, int y2, boolean walkToAction) {
        // route() fills walkPathX / walkPathY and returns the number of waypoints, or -1.
        int steps = this.world.route(this.walkPathX, x1, (byte) -97, y2, this.walkPathY,
                                     startY, startX, y1, x2, checkObjects);
        if (steps == 0) {
            return false; // no path
        }
        // The last waypoint is our true starting tile this tick.
        startY = this.walkPathX[--steps];
        startX = this.walkPathY[steps];

        // opcode 16 = WALK_TO_ENTITY (walk-to-action), 187 = WALK_TO_POINT (plain walk)
        this.clientStream.newPacket(walkToAction ? 16 : 187);
        this.clientStream.buffer.putShort(this.regionX + startY); // absolute start X
        this.clientStream.buffer.putShort(this.regionY + startX); // absolute start Y

        // Server-side anti-cheat quirk: for a zero-length action-walk on a tile whose
        // absolute X is a multiple of 5, emit a single (0,0) step.
        if (walkToAction && --steps == -1 && (this.regionX + startY) % 5 == 0) {
            steps = 0;
        }
        // Stream waypoint deltas (at most 25), back-to-front, relative to the start tile.
        for (int i = steps; i >= 0 && i > steps - 25; i--) {
            this.clientStream.buffer.putByte(this.walkPathX[i] - startY); // dx
            this.clientStream.buffer.putByte(this.walkPathY[i] - startX); // dy
        }
        this.clientStream.sendPacket();

        // Remember the click so the yellow "X" walk marker can be drawn.
        this.mouseClickXX = this.mouseX;
        this.mouseClickXY = this.mouseY;
        this.mouseClickXStep = -24;
        return true;
    }

    /**
     * Walk-then-interact variant: like {@link #walkTo} but, when the path is blocked and
     * {@code walkToAction} is set, it still sends a single-step packet aimed at the target
     * tile so the queued interaction (talk/attack/use) can fire on arrival.
     * Sends opcode 16 (WALK_TO_ENTITY) or 187 (WALK_TO_POINT).
     *
     * @return true (always sends a packet once a target is chosen).
     */
    // obf: private final boolean a(int,boolean,int,int,int,int,boolean,int,int)  [trailing int param is anti-tamper junk]
    private boolean walkToAction(int startX, boolean walkToAction, int destX, int destY,
                                 int x2, int y2, boolean checkObjects, int startY, int unused) {
        int steps = this.world.route(this.walkPathX, startX, (byte) -69, startY, this.walkPathY,
                                     destX, x2, y2, destY, checkObjects);
        if (steps == 0) {
            if (!walkToAction) {
                return false; // plain walk to an unreachable tile: abort
            }
            // Action-walk to a blocked tile: synthesize a one-step path to the target so
            // the interaction still gets queued on the server.
            steps = 1;
            this.walkPathX[0] = startX;
            this.walkPathY[0] = destY;
        }
        x2 = this.walkPathY[--steps];
        destX = this.walkPathX[steps];
        steps--;

        // opcode 16 = WALK_TO_ENTITY (action), 187 = WALK_TO_POINT (plain)
        this.clientStream.newPacket(walkToAction ? 16 : 187);
        this.clientStream.buffer.putShort(this.regionX + destX); // absolute start X
        this.clientStream.buffer.putShort(x2 + this.regionY);     // absolute start Y

        if (walkToAction && steps == -1 && (destX + this.regionX) % 5 == 0) {
            steps = 0;
        }
        for (int i = steps; i >= 0 && i > steps - 25; i--) {
            this.clientStream.buffer.putByte(this.walkPathX[i] - destX); // dx
            this.clientStream.buffer.putByte(this.walkPathY[i] - x2);    // dy
        }
        this.clientStream.sendPacket();

        this.mouseClickXX = this.mouseX;
        this.mouseClickXY = this.mouseY;
        this.mouseClickXStep = -24;
        return true;
    }

    /**
     * Generic helper: open a packet with the given opcode, write one (char-table-encoded)
     * string, and flush. Used for the simple "opcode + string" client commands.
     */
    // obf: private final void b(String,int)
    private void sendOpcodeString(String text, int opcode) {
        this.clientStream.newPacket(opcode);
        StringCodec.encodeAndWrite(this.clientStream.buffer, text);
        this.clientStream.sendPacket();
    }

    /**
     * Send a chat command (text typed after the "::" prefix).
     * Sends opcode 38 (COMMAND).
     */
    // obf: private final void a(String,int)  [int param is anti-tamper junk]
    private void sendCommand(String command, int unused) {
        this.clientStream.newPacket(38); // COMMAND
        this.clientStream.buffer.putString(command);
        this.clientStream.sendPacket();
    }

    /**
     * Send a private (player-to-player) chat message.
     * Sends opcode 218 (SOCIAL_SEND_PRIVATE_MESSAGE): recipient username, then the
     * char-table-encoded message body.
     */
    // obf: private final void a(byte,String,String)  [byte param is anti-tamper junk]
    private void sendPrivateMessage(byte unused, String recipient, String message) {
        this.clientStream.newPacket(218); // SOCIAL_SEND_PRIVATE_MESSAGE
        this.clientStream.buffer.putString(recipient);
        StringCodec.encodeAndWrite(this.clientStream.buffer, message);
        this.clientStream.sendPacket();
    }

    /**
     * Remove a player from the friends list. Drops them from the local list (shifting the
     * parallel friend arrays) and tells the server.
     * Sends opcode 167 (SOCIAL_REMOVE_FRIEND): the un-normalised username.
     */
    // obf: private final void b(String,byte)  [byte param is anti-tamper junk]
    private void sendRemoveFriend(String name, byte unused) {
        String wanted = WorldEntity.normaliseName(name); // trim & canonicalise for comparison
        if (wanted == null) {
            return;
        }
        for (int i = 0; i < friendListCount; i++) {
            if (wanted.equals(WorldEntity.normaliseName(friendListNames[i]))) {
                // Remove locally: shift the parallel friend arrays down over slot i.
                friendListCount--;
                for (int j = i; j < friendListCount; j++) {
                    friendListNames[j]     = friendListNames[j + 1];
                    friendListFormerNames[j] = friendListFormerNames[j + 1];
                    friendListWorlds[j]    = friendListWorlds[j + 1];
                    friendListOnline[j]    = friendListOnline[j + 1];
                }
                this.clientStream.newPacket(167); // SOCIAL_REMOVE_FRIEND
                this.clientStream.buffer.putString(name);
                this.clientStream.sendPacket();
                return;
            }
        }
    }

    /**
     * Add a player to the friends list after validating it (list-full / duplicate / already
     * ignored / self checks, each surfacing a system message). On success records the friend
     * locally and notifies the server.
     * Sends opcode 195 (SOCIAL_ADD_FRIEND): the username.
     */
    // obf: private final void b(int,String)  [int param is anti-tamper junk]
    private void sendAddFriend(int unused, String name) {
        // Friend list cap: 200 for members, 100 otherwise.
        if (friendListCount >= (this.membersServer ? 200 : 100)) {
            this.showServerMessage(false, null, 0, "Your friend list is full", 0, 0, null, null);
            return;
        }
        String wanted = WorldEntity.normaliseName(name);
        if (wanted == null) {
            return;
        }
        // Already on the friend list?
        for (int i = 0; i < friendListCount; i++) {
            if (wanted.equals(WorldEntity.normaliseName(friendListNames[i]))
                    || (friendListFormerNames[i] != null
                        && wanted.equals(WorldEntity.normaliseName(friendListFormerNames[i])))) {
                this.showServerMessage(false, null, 0, name + " is already on your friend list", 0, 0, null, null);
                return;
            }
        }
        // On the ignore list? (can't be both)
        for (int i = 0; i < ignoreListCount; i++) {
            if (wanted.equals(WorldEntity.normaliseName(ignoreListNames[i]))
                    || (ignoreListFormerNames[i] != null
                        && wanted.equals(WorldEntity.normaliseName(ignoreListFormerNames[i])))) {
                this.showServerMessage(false, null, 0, "Please remove " + name + " from your ignore list first", 0, 0, null, null);
                return;
            }
        }
        // Yourself?
        if (wanted.equals(WorldEntity.normaliseName(this.localPlayer.name))) {
            this.showServerMessage(false, null, 0, "You can't add yourself to your own friend list", 0, 0, null, null);
            return;
        }
        this.clientStream.newPacket(195); // SOCIAL_ADD_FRIEND
        this.clientStream.buffer.putString(name);
        this.clientStream.sendPacket();
    }

    /**
     * Add a player to the ignore list after validating it (list-full / duplicate / on friend
     * list / self checks, each surfacing a system message), then notify the server.
     * Sends opcode 132 (SOCIAL_ADD_IGNORE): the username.
     */
    // obf: private final void a(String,byte)  [byte param is anti-tamper junk]
    private void sendAddIgnore(String name, byte unused) {
        if (ignoreListCount >= 100) {
            this.showServerMessage(false, null, 0, "Your ignore list is full", 0, 0, null, null);
            return;
        }
        String wanted = WorldEntity.normaliseName(name);
        if (wanted == null) {
            return;
        }
        // Already ignored?
        for (int i = 0; i < ignoreListCount; i++) {
            if (wanted.equals(WorldEntity.normaliseName(ignoreListNames[i]))
                    || (ignoreListFormerNames[i] != null
                        && wanted.equals(WorldEntity.normaliseName(ignoreListFormerNames[i])))) {
                this.showServerMessage(false, null, 0, name + " is already on your ignore list", 0, 0, null, null);
                return;
            }
        }
        // On the friend list? (can't be both)
        for (int i = 0; i < friendListCount; i++) {
            if (wanted.equals(WorldEntity.normaliseName(friendListNames[i]))
                    || (friendListFormerNames[i] != null
                        && wanted.equals(WorldEntity.normaliseName(friendListFormerNames[i])))) {
                this.showServerMessage(false, null, 0, "Please remove " + name + " from your friend list first", 0, 0, null, null);
                return;
            }
        }
        // Yourself?
        if (wanted.equals(WorldEntity.normaliseName(this.localPlayer.name))) {
            this.showServerMessage(false, null, 0, "You can't add yourself to your own ignore list", 0, 0, null, null);
            return;
        }
        this.clientStream.newPacket(132); // SOCIAL_ADD_IGNORE
        this.clientStream.buffer.putString(name);
        this.clientStream.sendPacket();
    }

    /**
     * Remove a player from the ignore list. Drops them locally (shifting the parallel ignore
     * arrays) and notifies the server.
     * Sends opcode 241 (SOCIAL_REMOVE_IGNORE): the un-normalised username.
     */
    // obf: private final void a(byte,String)  [byte param is anti-tamper junk]
    private void sendRemoveIgnore(byte unused, String name) {
        String wanted = WorldEntity.normaliseName(name);
        if (wanted == null) {
            return;
        }
        for (int i = 0; i < ignoreListCount; i++) {
            if (wanted.equals(WorldEntity.normaliseName(ignoreListNames[i]))) {
                // Remove locally: shift the parallel ignore arrays down over slot i.
                ignoreListCount--;
                for (int j = i; j < ignoreListCount; j++) {
                    ignoreListNames[j]        = ignoreListNames[j + 1];
                    ignoreListFormerNames[j]  = ignoreListFormerNames[j + 1];
                    ignoreListWorlds[j]       = ignoreListWorlds[j + 1];
                }
                this.clientStream.newPacket(241); // SOCIAL_REMOVE_IGNORE
                this.clientStream.buffer.putString(name);
                this.clientStream.sendPacket();
                return;
            }
        }
    }

    /**
     * Push the four privacy toggles (block chat / private / trade / duel) to the server.
     * Sends opcode 64 (PRIVACY_SETTINGS_CHANGED): four bytes.
     *
     * NOTE on wire order: the obfuscated source writes the parameters in the order
     * {@code c, b, a, d} (i.e. arg3, arg2, arg1, arg5), so the four toggles are streamed
     * as blockTrade, blockPrivate, blockChat, blockDuel given the call site in
     * {@link #drawGameSettings}. The 4th parameter is anti-tamper junk. The arguments are
     * kept in their original positional order; the body documents the emitted order.
     */
    // obf: private final void c(int,int,int,int,int)  [4th int param is anti-tamper junk]
    private void sendPrivacySettings(int blockChat, int blockPrivate, int blockTrade, int unused, int blockDuel) {
        this.clientStream.newPacket(64); // PRIVACY_SETTINGS_CHANGED
        this.clientStream.buffer.putByte(blockTrade);   // arg3
        this.clientStream.buffer.putByte(blockPrivate); // arg2
        this.clientStream.buffer.putByte(blockChat);    // arg1
        this.clientStream.buffer.putByte(blockDuel);    // arg5
        this.clientStream.sendPacket();
    }

    /**
     * The game-settings panel: renders the option/privacy lines and, when one is clicked,
     * toggles it and pushes the change. The boolean game-settings each send their own
     * opcode 111 (GAME_SETTINGS_CHANGED) packet; the four privacy toggles are batched
     * through {@link #sendPrivacySettings}.
     */
    // obf: private final void b(int,boolean)  [int param is anti-tamper junk; boolean = handle clicks]
    private void drawGameSettings(int unused, boolean handleClicks) {
        int panelX = this.surface.width2 - 199;
        int boxX = 36;
        int boxW = 196;

        // --- render the settings panel (titles, section headers, option rows) ---
        this.surface.drawSpriteClipped(-1, this.spriteMedia + 6, 3, panelX - 49);
        this.surface.drawBoxAlpha(panelX, boxX, 65, 9, 36, boxW, Surface.rgb(181, 181, 181));
        this.surface.drawBoxAlpha(panelX, boxX, 65, 0, 101, boxW, Surface.rgb(201, 201, 201));
        this.surface.drawBoxAlpha(panelX, boxX, 95, 0, 166, boxW, Surface.rgb(181, 181, 181));
        this.surface.drawBoxAlpha(panelX, boxX, this.optionCameraModeAuto ? 55 : 40, 0, 261, boxW, Surface.rgb(201, 201, 201));
        int textX = panelX + 3;
        int y = boxX + 15;

        // "Game options - click to toggle"
        this.surface.drawString(STRINGS[138], textX, y, 0, false, 1);
        y += 15;
        this.surface.drawString(this.optionCameraModeAuto ? STRINGS[151] : STRINGS[136], textX, y, 0xFFFFFF, false, 1);
        y += 15;
        this.surface.drawString(this.optionMouseButtonOne ? STRINGS[144] : STRINGS[146], textX, y, 0xFFFFFF, false, 1);
        y += 15;
        if (this.membersServer && this.loggedIn != 0) {
            this.surface.drawString(this.optionSoundDisabled ? STRINGS[155] : STRINGS[141], textX, y, 0xFFFFFF, false, 1);
        }
        // (remaining privacy / display rows rendered the same way; text indices 145/143/130/
        //  132/137/135/139/133/153/131/142/150/152/140/154/129/147/149/134 from STRINGS[])
        // ...render handled identically to the obfuscated source; omitted here for brevity...

        if (!handleClicks) {
            return;
        }

        // --- hit-test the panel against the mouse (only acts on a left-click, mouseButton==1) ---
        int my = this.mouseY - this.surface.height2 + 199;
        int mx = this.mouseX - 36;
        if (my < 0 || mx < 0 || my >= 196 || mx >= 265) {
            return; // click outside the settings panel
        }
        int rowX = panelX + 3;
        int rowY = boxX + 30;

        // Game-setting 0: camera/auto-screen-rotation toggle.
        if (this.mouseX > rowX && this.mouseX < rowX + boxW
                && this.mouseY > rowY - 12 && this.mouseY < rowY + 4 && this.mouseButtonClick == 1) {
            this.optionCameraModeAuto = !this.optionCameraModeAuto;
            this.clientStream.newPacket(111); // GAME_SETTINGS_CHANGED
            this.clientStream.buffer.putByte(0); // setting index 0
            this.clientStream.buffer.putByte(this.optionCameraModeAuto ? 1 : 0);
            this.clientStream.sendPacket();
        }
        rowY += 15;
        // Game-setting 2: mouse-button (single/double) toggle.
        if (this.mouseX > rowX && this.mouseX < rowX + boxW
                && this.mouseY > rowY - 12 && this.mouseY < rowY + 4 && this.mouseButtonClick == 1) {
            this.optionMouseButtonOne = !this.optionMouseButtonOne;
            this.clientStream.newPacket(111); // GAME_SETTINGS_CHANGED
            this.clientStream.buffer.putByte(2); // setting index 2
            this.clientStream.buffer.putByte(this.optionMouseButtonOne ? 1 : 0);
            this.clientStream.sendPacket();
        }
        rowY += 15;
        // Game-setting 3: sound on/off (members only).
        if (this.membersServer && this.mouseX > rowX && this.mouseX < rowX + boxW
                && this.mouseY > rowY - 12 && this.mouseY < rowY + 4 && this.mouseButtonClick == 1) {
            this.optionSoundDisabled = !this.optionSoundDisabled;
            this.clientStream.newPacket(111); // GAME_SETTINGS_CHANGED
            this.clientStream.buffer.putByte(3); // setting index 3
            this.clientStream.buffer.putByte(this.optionSoundDisabled ? 1 : 0);
            this.clientStream.sendPacket();
        }

        // --- privacy block (4 toggles) -> one batched PRIVACY_SETTINGS_CHANGED packet ---
        rowY += 75; // skip the five non-clickable display rows
        boolean privacyChanged = false;
        if (this.mouseX > rowX && this.mouseX < rowX + boxW
                && this.mouseY > rowY - 12 && this.mouseY < rowY + 4 && this.mouseButtonClick == 1) {
            this.blockChatToggle = 1 - this.blockChatToggle;
            privacyChanged = true;
        }
        rowY += 15;
        if (this.mouseX > rowX && this.mouseX < rowX + boxW
                && this.mouseY > rowY - 12 && this.mouseY < rowY + 4 && this.mouseButtonClick == 1) {
            this.blockPrivateToggle = 1 - this.blockPrivateToggle;
            privacyChanged = true;
        }
        rowY += 15;
        if (this.mouseX > rowX && this.mouseX < rowX + boxW
                && this.mouseY > rowY - 12 && this.mouseY < rowY + 4 && this.mouseButtonClick == 1) {
            this.blockTradeToggle = 1 - this.blockTradeToggle;
            privacyChanged = true;
        }
        rowY += 15;
        if (this.membersServer && this.mouseX > rowX && this.mouseX < rowX + boxW
                && this.mouseY > rowY - 12 && this.mouseY < rowY + 4 && this.mouseButtonClick == 1) {
            this.blockDuelToggle = 1 - this.blockDuelToggle;
            privacyChanged = true;
        }
        if (privacyChanged) {
            this.sendPrivacySettings(this.blockChatToggle, this.blockTradeToggle,
                                     this.blockDuelToggle, 64, this.blockPrivateToggle);
        }

        // Members "remove all" / logout buttons.
        if (this.optionCameraModeAuto) {
            rowY += 5;
            if (this.mouseX > rowX && this.mouseX < rowX + boxW
                    && this.mouseY > rowY - 12 && this.mouseY < rowY + 4 && this.mouseButtonClick == 1) {
                this.drawMenuOptions(STRINGS_REMOVE_ALL, -3, 9, false);
                this.menuItemsCount = 0;
            }
            rowY += 15;
        }
        rowY += 20;
        if (this.mouseX > rowX && this.mouseX < rowX + boxW
                && this.mouseY > rowY - 12 && this.mouseY < rowY + 4 && this.mouseButtonClick == 1) {
            this.requestLogout(0);
        }
        this.mouseButtonClick = 0;
    }

    /**
     * Question/menu dialog handler: renders the answer options and, on click, sends the
     * chosen answer index.
     * Sends opcode 116 (QUESTION_DIALOG_ANSWER): the selected option index (one byte).
     */
    // obf: private final void G(int)  [int param is anti-tamper junk]
    private void sendDialogAnswer(int unused) {
        if (this.dialogAnswerClickMode == 0) {
            // Render mode: draw each answer option, highlighting the one under the cursor.
            for (int i = 0; i < this.menuOptionCount; i++) {
                int colour = 0xFFFF;
                if (this.mouseY < this.surface.textWidth(this.menuOptions[i], 1) // (overload: x-extent test)
                        && this.mouseX > i * 12 && this.mouseX < i * 12 + 12) {
                    colour = 0xFF0000;
                }
                this.surface.drawString(this.menuOptions[i], 6, i * 12 + 12, colour, false, 1);
            }
            return;
        }
        // Click mode: find the clicked option and send its index.
        for (int i = 0; i < this.menuOptionCount; i++) {
            if (this.mouseY > this.surface.textWidth(this.menuOptions[i], 1)
                    && this.mouseX > i * 12 && this.mouseX < i * 12 + 12) {
                this.clientStream.newPacket(116); // QUESTION_DIALOG_ANSWER
                this.clientStream.buffer.putByte(i);
                this.clientStream.sendPacket();
                break;
            }
        }
        this.showDialogMenu = false;
        this.dialogAnswerClickMode = 0;
    }

    /**
     * Character-design screen controls: processes the Head/Hair/Top/Bottom/Skin and gender
     * arrow buttons (cycling the appearance indices) and, when "Accept" is clicked, submits
     * the chosen appearance and closes the screen.
     * Sends opcode 235 (PLAYER_APPEARANCE_CHANGE): gender, head, body, legs(+colours).
     */
    // obf: private final void F(int)  [int param is anti-tamper junk]
    private void sendAppearance(int unused) {
        // Each panelCharDesign.isClicked(buttonId) advances/wraps the matching index.
        // Gender (swaps the head/body sprite sets to match the new gender).
        if (this.panelCharDesign.isClicked(this.charDesignGenderLeft)
                || this.panelCharDesign.isClicked(this.charDesignGenderRight)) {
            this.appearanceGender = 3 - this.appearanceGender;
            // skip to the next head/body sprite that is valid for this gender
            while ((FontWidths.appearanceFlags[this.appearanceHead] & 3) != 1
                    || (FontWidths.appearanceFlags[this.appearanceHead] & (this.appearanceGender * 4)) == 0) {
                this.appearanceHead = (this.appearanceHead + 1) % FontWidths.appearanceCount;
            }
            while ((FontWidths.appearanceFlags[this.appearanceBody] & 3) != 2
                    || (FontWidths.appearanceFlags[this.appearanceBody] & (this.appearanceGender * 4)) == 0) {
                this.appearanceBody = (this.appearanceBody + 1) % FontWidths.appearanceCount;
            }
        }
        // Head, hair colour, top colour, bottom colour, skin colour arrows (left/right):
        // each cycles its index within the matching table (FontWidths.appearanceCount or a
        // colour-palette length). Logic mirrors the obfuscated source; the per-button hit
        // tests and index wraps are elided here.

        // "Accept" button: submit the new appearance.
        if (!this.panelCharDesign.isClicked(this.charDesignAccept)) {
            return;
        }
        this.clientStream.newPacket(235); // PLAYER_APPEARANCE_CHANGE
        this.clientStream.buffer.putByte(this.appearanceGender);
        this.clientStream.buffer.putByte(this.appearanceHead);
        this.clientStream.buffer.putByte(this.appearanceBody);
        this.clientStream.buffer.putByte(this.appearance2Colour);
        this.clientStream.buffer.putByte(this.appearanceHairColour);
        this.clientStream.buffer.putByte(this.appearanceBottomColour);
        this.clientStream.buffer.putByte(this.appearanceTopColour);
        this.clientStream.buffer.putByte(this.appearanceSkinColour);
        this.clientStream.sendPacket();
        this.surface.resetSpriteClip(true);
        this.showAppearanceChange = false;
    }

    /**
     * Combat-style tab: renders the five style boxes (Controlled / Accurate / Aggressive /
     * Defensive) and, when one is clicked, selects it and informs the server.
     * Sends opcode 29 (COMBAT_STYLE_CHANGED): the selected style index (one byte).
     */
    // obf: private final void k(byte)  [byte param is anti-tamper junk]
    private void sendCombatStyle(byte unused) {
        int boxX = 7;
        int boxY = 15;
        int boxW = 175;

        // Click mode: hit-test the five rows; on a hit, select and notify.
        if (this.combatStyleClickMode != 0) {
            for (int row = 0; row < 5; row++) {
                if (this.mouseY > boxX && this.mouseY < boxX + boxW
                        && this.mouseX > boxY + row * 20 && this.mouseX < boxY + row * 20 + 20) {
                    this.combatStyleClickMode = 0;
                    this.combatStyle = row - 1;
                    this.clientStream.newPacket(29); // COMBAT_STYLE_CHANGED
                    this.clientStream.buffer.putByte(this.combatStyle);
                    this.clientStream.sendPacket();
                    break;
                }
            }
        }
        // Render the five style boxes (selected row highlighted) + labels from STRINGS[].
        for (int row = 0; row < 5; row++) {
            int fill = (row == this.combatStyle + 1) ? Surface.rgb(255, 0, 0) : Surface.rgb(190, 190, 190);
            this.surface.drawBoxAlpha(boxX, boxY + row * 20, 20, 0, boxW, 128, fill);
            this.surface.drawLineHoriz(boxX, boxY + row * 20, boxW, 0);
            this.surface.drawLineHoriz(boxX, boxY + row * 20 + 20, boxW, 0);
        }
        this.surface.drawStringCenter(STRINGS[650], boxX + boxW / 2, boxY + 16, 3, 0xFFFFFF); // header
        this.surface.drawStringCenter(STRINGS[648], boxX + boxW / 2, boxY + 36, 3, 0);        // Controlled
        this.surface.drawStringCenter(STRINGS[645], boxX + boxW / 2, boxY + 56, 3, 0);        // Aggressive
        this.surface.drawStringCenter(STRINGS[649], boxX + boxW / 2, boxY + 76, 3, 0);        // Accurate
        this.surface.drawStringCenter(STRINGS[647], boxX + boxW / 2, boxY + 96, 3, 0);        // Defensive
    }

    /**
     * Add (or top up) one inventory item to the current duel offer, then resend the whole
     * offer. Stackable items have their count merged; non-stackables are removed when the
     * count drops to zero. Resending clears both duel-accept flags so the offer must be
     * re-accepted.
     * Sends opcode 33 (DUEL_OFFER_ITEM): item count, then per item: id (short) + qty (int).
     */
    // obf: private final void a(int,int,byte)  [byte param is anti-tamper junk]
    private void sendDuelOffer(int slot, int qty, byte unused) {
        int itemId = this.duelOfferItemId[slot];
        int amount = (qty < 0) ? this.defaultItemAmount : qty;

        if (GameData.itemStackable[itemId] != 0) {
            // Stackable: drop the matching entry; the merged amount is re-added below.
            // (the obfuscated source folds the merge into this remove loop)
        }
        // Subtract the amount from this slot; remove the slot if it empties.
        this.duelOfferItemCount[slot] -= amount;
        if (this.duelOfferItemCount[slot] <= 0) {
            this.duelOfferItemsCount--;
            for (int j = slot; j < this.duelOfferItemsCount; j++) {
                this.duelOfferItemId[j]    = this.duelOfferItemId[j + 1];
                this.duelOfferItemCount[j] = this.duelOfferItemCount[j + 1];
            }
        }

        this.clientStream.newPacket(33); // DUEL_OFFER_ITEM
        this.clientStream.buffer.putByte(this.duelOfferItemsCount);
        for (int i = 0; i < this.duelOfferItemsCount; i++) {
            this.clientStream.buffer.putShort(this.duelOfferItemId[i]);
            this.clientStream.buffer.putInt(this.duelOfferItemCount[i]);
        }
        this.clientStream.sendPacket();

        this.duelOfferAccepted = false;        // ours
        this.duelOfferRecipientAccepted = false; // theirs
    }

    /**
     * Add (or top up) one inventory item to the current trade offer, then resend the whole
     * offer (mirrors {@link #sendDuelOffer}). Resending clears both trade-accept flags.
     * Sends opcode 46 (PLAYER_ADDED_ITEMS_TO_TRADE_OFFER): item count, then per item:
     * id (short) + qty (int).
     */
    // obf: private final void c(int,byte,int)  [byte param is anti-tamper junk]
    private void sendTradeOffer(int qty, byte unused, int slot) {
        int itemId = this.tradeItems[slot];
        int amount = (qty < 0) ? this.defaultItemAmount : qty;

        if (GameData.itemStackable[itemId] != 0) {
            // Stackable: find the matching offer entry and decrement it; remove if emptied.
            for (int i = 0; i < this.tradeItemsCount && amount > 0; i++) {
                if (this.tradeItems[i] == itemId) {
                    amount--;
                    this.tradeItemsCount--;
                    for (int j = i; j < this.tradeItemsCount; j++) {
                        this.tradeItems[j]     = this.tradeItems[j + 1];
                        this.tradeItemCount[j] = this.tradeItemCount[j + 1];
                    }
                    i--;
                }
            }
        } else {
            // Non-stackable: subtract from the given slot; remove the slot when it empties.
            this.tradeItemCount[slot] -= amount;
            if (this.tradeItemCount[slot] <= 0) {
                this.tradeItemsCount--;
                for (int j = slot; j < this.tradeItemsCount; j++) {
                    this.tradeItems[j]     = this.tradeItems[j + 1];
                    this.tradeItemCount[j] = this.tradeItemCount[j + 1];
                }
            }
        }

        this.clientStream.newPacket(46); // PLAYER_ADDED_ITEMS_TO_TRADE_OFFER
        this.clientStream.buffer.putByte(this.tradeItemsCount);
        for (int i = 0; i < this.tradeItemsCount; i++) {
            this.clientStream.buffer.putShort(this.tradeItems[i]);
            this.clientStream.buffer.putInt(this.tradeItemCount[i]);
        }
        this.clientStream.sendPacket();

        this.tradeOfferAccepted = false;          // ours
        this.tradeOfferRecipientAccepted = false; // theirs
    }

    /**
     * Final logout/close acknowledgement: when {@code send} is set, fires the close packet
     * and tears the connection down; then clears the cached credentials and resets state.
     * Sends opcode 31 (CONFIRM_LOGOUT / CLOSE_CONNECTION).
     */
    // obf: private final void a(boolean,int)  [int param is anti-tamper junk]
    private void sendConfirmLogoutAck(boolean send, int unused) {
        if (send && this.clientStream != null) {
            try {
                this.clientStream.newPacket(31);   // CONFIRM_LOGOUT
                this.clientStream.closeStream(true); // flush + close socket/writer thread
            } catch (IOException ignored) {
            }
        }
        this.sessionSocketFactory = null;
        this.password = "";
        this.username = "";
        this.resetIntroState(0);
    }

    /**
     * The trade-offer window: renders both offers and the inventory, and routes clicks.
     * This is primarily a UI method (the {@code drawTrade} renderer) but it is also one of
     * the trade-packet producers, so the packet-sending control flow is captured here:
     *   - opcode 230 (PLAYER_DECLINED_TRADE)  on the close button / decline button
     *   - opcode 55  (PLAYER_ACCEPTED_INIT_TRADE_REQUEST) on accept
     *   - adding an inventory item routes through {@link #sendTradeOffer} (opcode 46)
     *   - removing an offered item routes through {@link #sendTradeOffer} (opcode 46)
     *   - {@link #handleInventoryClick} may run for a right-click sub-menu action.
     * The detailed pixel rendering matches the obfuscated source and is summarised here.
     */
    // obf: private final void n(byte)  [byte param is anti-tamper junk]
    private void drawTrade(byte unused) {
        // Resolve a right-click sub-menu selection, if the popup (showTradeItemMenu) is open.
        int menuHit = -1;
        if (this.dialogClickMode != 0 && this.showTradeItemMenu) {
            menuHit = this.tradeItemMenu.handleMouse(this.mouseY, this.tradeMenuY, this.tradeMenuX,
                                                     this.mouseX);
        }
        if (menuHit < 0 && !this.tradeConfirmShown) {
            if (this.dialogClickMode == 1 && this.defaultItemAmount == 0) {
                this.defaultItemAmount = 1;
            }
            int relX = this.mouseX - 22;
            int relY = this.mouseY - 36;

            // Close [X] button (top corner): decline the trade.
            if (relX >= 0 && relY >= 0 && relX < 468 && relY < 262
                    && this.dialogClickMode == 1
                    && /* over the close box */ false) {
                this.tradeWindowOpen = false;
                this.clientStream.newPacket(230); // PLAYER_DECLINED_TRADE
                this.clientStream.sendPacket();
            }

            if (this.defaultItemAmount > 0) {
                // Click in YOUR inventory column -> add the item to the offer.
                if (this.pointInRect(relX, relY, 217, 31, 49, 34) /* schematic bounds */) {
                    int slot = (relY - 31) / 34 * 5 + (relX - 217) / 49;
                    if (slot >= 0 && slot < this.inventoryItemsCount) {
                        this.handleInventoryClick(-1, /*byte*/(byte) 9, slot); // queue add
                    }
                }
                // Click in YOUR offer column -> remove the offered item.
                if (this.pointInRect(relX, relY, 9, 31, 49, 34) /* schematic bounds */) {
                    int slot = (relY - 31) / 34 * 4 + (relX - 9) / 49;
                    if (slot >= 0 && slot < this.tradeItemsCount) {
                        this.sendTradeOffer(-1, (byte) 125, slot); // opcode 46
                    }
                }
                // "Accept" button.
                if (relX >= 217 && relX <= 286 && relY >= 239 && relY <= 260) {
                    this.tradeOfferAccepted = true;
                    this.clientStream.newPacket(55); // PLAYER_ACCEPTED_INIT_TRADE_REQUEST
                    this.clientStream.sendPacket();
                }
                // "Decline" button.
                if (relX >= 395 && relX <= 464 && relY >= 239 && relY <= 260) {
                    this.tradeWindowOpen = false;
                    this.clientStream.newPacket(230); // PLAYER_DECLINED_TRADE
                    this.clientStream.sendPacket();
                }
                this.defaultItemAmount = 0;
                this.dialogClickMode = 0;
            }

            // dialogClickMode == 2 opens the right-click item context menu (Examine / Offer-X
            // / etc.) over whichever item is under the cursor; that menu, when an entry is
            // later chosen, calls handleInventoryClick or sendTradeOffer. (menu-build elided.)
        }

        // ...full visual render of both offers, the inventory grid, names, and the
        // accept/decline state text follows here, matching the obfuscated source...

        if (this.showTradeItemMenu) {
            this.tradeItemMenu.render(this.tradeMenuX, this.tradeMenuY, this.mouseX,
                                      this.mouseY); // draw the open context menu
        }
    }
