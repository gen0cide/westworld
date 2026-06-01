    // ===== packetout =====
    //
    // Outgoing (client -> server) packet builders. Opcodes are cited from the OpenRSC
    // Payload235Parser (the rev-235 client->server map). The low-level wire helpers used
    // throughout this group are:
    //
    //   clientStream.newPacket(op)             // begin a packet with opcode `op`  (da.b(int,int))
    //   clientStream.sendPacket()              // flush/queue the packet           (da.b(int) -> b(21294))
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
    //
    // CROSS-CLASS NAMES (verified against docs/NAMING.md + clean base):
    //   Hh (type k=World)   -> this.world      [route()/pathfinding lives on World]
    //   Ek (type lb=Scene)  -> this.scene
    //   Cf                  -> mouseButtonClick (1=left, 2=right this tick; cleared after use)
    //   The social lists are scattered across static arrays on unrelated obf classes; the
    //   readable instance-style names below mirror the rest of the deob (see oracle GameConnection):
    //     friends: ua.h=names, cb.c=formerNames, ac.z=worlds(str), Fj=online/flags, n.g=count
    //     ignores: l.c=names, ia.a=displayNames, ia.g=formerNames, ua.wb=worlds(str), db.g=count
    //   fa.e (int[]) = GameData.itemStackable;  VALUE SEMANTICS: ==0 => STACKABLE, !=0 => not.

    /**
     * Walk to an explicit destination, optionally walking up to (rather than onto) the
     * target tile. Pathfinds with {@link World#route}, then streams the start tile plus
     * per-step deltas. Sends opcode 16 (WALK_TO_ENTITY) when {@code walkToAction} is set,
     * else opcode 187 (WALK_TO_POINT).
     *
     * @return true if a packet was sent (a route existed), false if unreachable.
     */
    // obf: private final boolean a(int,int,byte,boolean,int,int,int,int,boolean)  [byte param var3 is anti-tamper junk]
    private boolean walkTo(int startX, int startY, byte unused, boolean checkObjects,
                           int x1, int y1, int x2, int y2, boolean walkToAction) {
        // route() fills walkPathX (Rg) / walkPathY (pf) and returns the waypoint count, or -1.
        int steps = this.world.route(this.walkPathX, x1, (byte) -97, y2, this.walkPathY,
                                     startY, startX, y1, x2, checkObjects);
        if (steps == -1) {            // obf: ~steps == 0  ⟺  steps == -1  (no path)
            return false;
        }

        // The last waypoint is our true starting tile this tick (read both arrays at the
        // same index, then drop into the per-step stream).
        int curX = this.walkPathX[--steps]; // obf: var2 = Rg[--var10]
        int curY = this.walkPathY[steps];   // obf: var1 = pf[var10]

        // opcode 16 = WALK_TO_ENTITY (walk-to-action), 187 = WALK_TO_POINT (plain walk)
        this.clientStream.newPacket(walkToAction ? 16 : 187);
        this.clientStream.buffer.putShort(this.regionX + curX); // obf: Qg + var2 (absolute start X)
        this.clientStream.buffer.putShort(this.regionY + curY); // obf: zg + var1 (absolute start Y)

        steps--; // obf: var10-- (UNCONDITIONAL second decrement, before the loop bound)

        // Server-side anti-cheat quirk: for a zero-length action-walk on a tile whose
        // absolute X is a multiple of 5, emit a single (0,0) step.
        if (walkToAction && steps == -1 && (this.regionX + curX) % 5 == 0) {
            steps = 0;
        }
        // Stream waypoint deltas (at most 25), back-to-front, relative to the start tile.
        for (int i = steps; i >= 0 && i > steps - 25; i--) {
            this.clientStream.buffer.putByte(this.walkPathX[i] - curX); // dx
            this.clientStream.buffer.putByte(this.walkPathY[i] - curY); // dy
        }
        this.clientStream.sendPacket();

        // Remember the click so the yellow "X" walk marker can be drawn.
        this.mouseClickXX = this.mouseX;   // obf: tj = I
        this.mouseClickXY = this.mouseY;   // obf: Fd = xb
        this.mouseClickXStep = -24;        // obf: xh = -24
        return true;
    }

    /**
     * Walk-then-interact variant: like {@link #walkTo} but, when the path is blocked and
     * {@code walkToAction} is set, it still synthesises a single-step packet aimed at the
     * target tile so the queued interaction (talk/attack/use) can fire on arrival.
     * Sends opcode 16 (WALK_TO_ENTITY) or 187 (WALK_TO_POINT).
     *
     * @return true once a target is chosen (and a packet is sent).
     */
    // obf: private final boolean a(int,boolean,int,int,int,int,boolean,int,int)  [trailing int param var9 is anti-tamper junk]
    private boolean walkToAction(int startX, boolean walkToAction, int destX, int destY,
                                 int x2, int y2, boolean checkObjects, int startY, int unused) {
        int steps = this.world.route(this.walkPathX, startX, (byte) -69, startY, this.walkPathY,
                                     destX, x2, y2, destY, checkObjects);
        if (steps == -1) {            // obf: ~steps == 0  ⟺  steps == -1  (blocked)
            if (!walkToAction) {
                return false; // plain walk to an unreachable tile: abort
            }
            // Action-walk to a blocked tile: synthesise a one-step path to the target so
            // the interaction still gets queued on the server.
            steps = 1;
            this.walkPathX[0] = startX; // obf: Rg[0] = var1
            this.walkPathY[0] = destY;  // obf: pf[0] = var4
        }
        int curY = this.walkPathY[--steps]; // obf: var5 = pf[--var10]
        int curX = this.walkPathX[steps];   // obf: var3 = Rg[var10]
        steps--;                            // obf: var10-- (unconditional)

        // opcode 16 = WALK_TO_ENTITY (action), 187 = WALK_TO_POINT (plain)
        this.clientStream.newPacket(walkToAction ? 16 : 187);
        this.clientStream.buffer.putShort(this.regionX + curX); // obf: Qg + var3
        this.clientStream.buffer.putShort(curY + this.regionY); // obf: var5 + zg

        if (walkToAction && steps == -1 && (curX + this.regionX) % 5 == 0) {
            steps = 0;
        }
        for (int i = steps; i >= 0 && i > steps - 25; i--) {
            this.clientStream.buffer.putByte(this.walkPathX[i] - curX); // dx
            this.clientStream.buffer.putByte(this.walkPathY[i] - curY); // dy
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
    // obf: private final void b(String,int)   [b.b(op, op^216): 2nd newPacket arg is junk]
    private void sendOpcodeString(String text, int opcode) {
        this.clientStream.newPacket(opcode);
        StringCodec.encodeAndWrite(this.clientStream.buffer, text); // obf: u.a(99, Jh.f, var1)
        this.clientStream.sendPacket();
    }

    /**
     * Send a chat command (text typed after the "::" prefix).
     * Sends opcode 38 (COMMAND).
     */
    // obf: private final void a(String,int)   [int param var2 is anti-tamper junk]
    private void sendCommand(String command, int unused) {
        this.clientStream.newPacket(38); // COMMAND
        this.clientStream.buffer.putString(command); // obf: Jh.f.a(var1, 104)
        this.clientStream.sendPacket();
    }

    /**
     * Send a private (player-to-player) chat message.
     * Sends opcode 218 (SOCIAL_SEND_PRIVATE_MESSAGE): recipient username, then the
     * char-table-encoded message body.
     */
    // obf: private final void a(byte,String,String)   [byte param var1 is anti-tamper junk]
    private void sendPrivateMessage(byte unused, String recipient, String message) {
        this.clientStream.newPacket(218); // SOCIAL_SEND_PRIVATE_MESSAGE
        this.clientStream.buffer.putString(recipient);                  // obf: Jh.f.a(var2, 124)
        StringCodec.encodeAndWrite(this.clientStream.buffer, message);  // obf: u.a(103, Jh.f, var3)
        this.clientStream.sendPacket();
    }

    /**
     * Remove a player from the friends list. Drops them from the local list (shifting the
     * parallel friend arrays) and tells the server.
     * Sends opcode 167 (SOCIAL_REMOVE_FRIEND): the un-normalised username.
     */
    // obf: private final void b(String,byte)   [byte param var2 is anti-tamper junk]
    private void sendRemoveFriend(String name, byte unused) {
        String wanted = WorldEntity.normaliseName(name); // obf: w.a(var1, ..)  trim & canonicalise
        if (wanted == null) {
            return;
        }
        for (int i = 0; i < friendListCount; i++) { // obf: var4 < n.g
            if (wanted.equals(WorldEntity.normaliseName(friendListNames[i]))) { // obf: ua.h[var4]
                // Remove locally: shift the four parallel friend arrays down over slot i.
                friendListCount--;
                for (int j = i; j < friendListCount; j++) {
                    friendListNames[j]       = friendListNames[j + 1]; // obf: ua.h
                    friendListFormerNames[j] = friendListFormerNames[j + 1]; // obf: cb.c
                    friendListWorlds[j]      = friendListWorlds[j + 1]; // obf: ac.z
                    friendListOnline[j]      = friendListOnline[j + 1]; // obf: Fj
                }
                this.clientStream.newPacket(167); // SOCIAL_REMOVE_FRIEND
                this.clientStream.buffer.putString(name); // obf: Jh.f.a(var1, 110)  (raw, un-normalised)
                this.clientStream.sendPacket();
                return;
            }
        }
    }

    /**
     * Add a player to the friends list after validating it (list-full / duplicate / already
     * ignored / self checks, each surfacing a system message). On success notifies the server
     * (the local list is filled by the SEND_FRIEND_UPDATE reply, not here).
     * Sends opcode 195 (SOCIAL_ADD_FRIEND): the username.
     */
    // obf: private final void b(int,String)   [int param var1 is anti-tamper junk]
    private void sendAddFriend(int unused, String name) {
        // Friend list cap: 200 for members, 100 otherwise.  obf: ~(Pg?200:100) >= ~n.g
        if (friendListCount >= (this.membersServer ? 200 : 100)) {
            this.showServerMessage(false, null, 0, "Your friend list is full", 0, 0, null, null); // il[384]
            return;
        }
        String wanted = WorldEntity.normaliseName(name);
        if (wanted == null) {
            return;
        }
        // Already on the friend list? (match either current or former name)
        for (int i = 0; i < friendListCount; i++) {
            if (wanted.equals(WorldEntity.normaliseName(friendListNames[i]))
                    || (friendListFormerNames[i] != null
                        && wanted.equals(WorldEntity.normaliseName(friendListFormerNames[i])))) {
                this.showServerMessage(false, null, 0, name + " is already on your friend list", 0, 0, null, null); // il[386]
                return;
            }
        }
        // On the ignore list? (can't be both)  obf: var4 < db.g over l.c / ia.g
        for (int i = 0; i < ignoreListCount; i++) {
            if (wanted.equals(WorldEntity.normaliseName(ignoreListNames[i]))
                    || (ignoreListFormerNames[i] != null
                        && wanted.equals(WorldEntity.normaliseName(ignoreListFormerNames[i])))) {
                this.showServerMessage(false, null, 0, "Please remove " + name + " from your ignore list first", 0, 0, null, null); // il[251]+name+il[383]
                return;
            }
        }
        // Yourself?  obf: w.a(wi.C, ..)
        if (wanted.equals(WorldEntity.normaliseName(this.localPlayer.name))) {
            this.showServerMessage(false, null, 0, "You can't add yourself to your own friend list", 0, 0, null, null); // il[385]
            return;
        }
        this.clientStream.newPacket(195); // SOCIAL_ADD_FRIEND
        this.clientStream.buffer.putString(name); // obf: Jh.f.a(var2, -23)
        this.clientStream.sendPacket();
    }

    /**
     * Add a player to the ignore list after validating it (list-full / duplicate / on friend
     * list / self checks, each surfacing a system message), then notify the server.
     * Sends opcode 132 (SOCIAL_ADD_IGNORE): the username.
     */
    // obf: private final void a(String,byte)   [byte param var2 is anti-tamper junk]
    private void sendAddIgnore(String name, byte unused) {
        // Ignore list cap: 100.  obf: ~db.g <= -101  ⟺  db.g >= 100
        if (ignoreListCount >= 100) {
            this.showServerMessage(false, null, 0, "Your ignore list is full", 0, 0, null, null); // il[254]
            return;
        }
        String wanted = WorldEntity.normaliseName(name);
        if (wanted == null) {
            return;
        }
        // Already ignored? (match current or former name)  obf: var4 < db.g over l.c / ia.g
        for (int i = 0; i < ignoreListCount; i++) {
            if (wanted.equals(WorldEntity.normaliseName(ignoreListNames[i]))
                    || (ignoreListFormerNames[i] != null
                        && wanted.equals(WorldEntity.normaliseName(ignoreListFormerNames[i])))) {
                this.showServerMessage(false, null, 0, name + " is already on your ignore list", 0, 0, null, null); // il[252]
                return;
            }
        }
        // On the friend list? (can't be both)  obf: var4 < n.g over ua.h / cb.c
        for (int i = 0; i < friendListCount; i++) {
            if (wanted.equals(WorldEntity.normaliseName(friendListNames[i]))
                    || (friendListFormerNames[i] != null
                        && wanted.equals(WorldEntity.normaliseName(friendListFormerNames[i])))) {
                this.showServerMessage(false, null, 0, "Please remove " + name + " from your friend list first", 0, 0, null, null); // il[251]+name+il[255]
                return;
            }
        }
        // Yourself?
        if (wanted.equals(WorldEntity.normaliseName(this.localPlayer.name))) {
            this.showServerMessage(false, null, 0, "You can't add yourself to your own ignore list", 0, 0, null, null); // il[253]
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
    // obf: private final void a(byte,String)   [byte param var1 is anti-tamper junk: if(var1<-7){..whole body}]
    private void sendRemoveIgnore(byte unused, String name) {
        String wanted = WorldEntity.normaliseName(name);
        if (wanted == null) {
            return;
        }
        for (int i = 0; i < ignoreListCount; i++) { // obf: var4 < db.g
            // NOTE: the match is against ignoreListDisplayNames (ia.a), not ignoreListNames (l.c).
            if (wanted.equals(WorldEntity.normaliseName(ignoreListDisplayNames[i]))) { // obf: ia.a[var4]
                // Remove locally: shift the FOUR parallel ignore arrays down over slot i.
                ignoreListCount--;
                for (int j = i; j < ignoreListCount; j++) {
                    ignoreListNames[j]        = ignoreListNames[j + 1];        // obf: l.c
                    ignoreListDisplayNames[j] = ignoreListDisplayNames[j + 1]; // obf: ia.a
                    ignoreListFormerNames[j]  = ignoreListFormerNames[j + 1];  // obf: ia.g
                    // obf: ua.wb[j] = ua.wb[j]  — the original client's shift omits the +1 here
                    // (both Vineflower and CFR agree); reproduced faithfully as a no-op self-assign.
                    ignoreListWorlds[j]       = ignoreListWorlds[j];           // obf: ua.wb (self-assign, original bug)
                }
                this.clientStream.newPacket(241); // SOCIAL_REMOVE_IGNORE
                this.clientStream.buffer.putString(name); // obf: Jh.f.a(var2, -78)
                this.clientStream.sendPacket();
                return;
            }
        }
    }

    /**
     * Push the four privacy toggles (block chat / private / trade / duel) to the server.
     * Sends opcode 64 (PRIVACY_SETTINGS_CHANGED): four bytes in wire order chat, priv, trade, duel.
     *
     * Wire order note: the obfuscated body writes the parameters as {@code var3, var2, var1, var5}.
     * Combined with the (only) call site {@code c(Vg, dc, De, 64, ui)} that means the params land as
     * var1=blockTrade(Vg), var2=blockPrivate(dc), var3=blockChat(De), var5=blockDuel(ui); the four
     * bytes emitted are therefore chat, priv, trade, duel — matching oracle GameConnection.
     * The 4th parameter is anti-tamper junk.
     */
    // obf: private final void c(int,int,int,int,int)   [4th int param var4 is anti-tamper guard: if(var4>=62)]
    private void sendPrivacySettings(int blockTrade, int blockPrivate, int blockChat, int unused, int blockDuel) {
        this.clientStream.newPacket(64); // PRIVACY_SETTINGS_CHANGED
        this.clientStream.buffer.putByte(blockChat);    // obf: Jh.f.c(var3, ..)
        this.clientStream.buffer.putByte(blockPrivate); // obf: Jh.f.c(var2, ..)
        this.clientStream.buffer.putByte(blockTrade);   // obf: Jh.f.c(var1, ..)
        this.clientStream.buffer.putByte(blockDuel);    // obf: Jh.f.c(var5, ..)
        this.clientStream.sendPacket();
    }

    /**
     * The game-settings panel: renders the option / display / privacy rows and, when one is
     * clicked, toggles it and pushes the change. The three boolean game-settings each send
     * their own opcode 111 (GAME_SETTINGS_CHANGED) packet; the four privacy toggles are
     * batched through {@link #sendPrivacySettings} (opcode 64).
     */
    // obf: private final void b(int,boolean)   [int param var1 is anti-tamper junk; boolean var2 = handle clicks]
    private void drawGameSettings(int unused, boolean handleClicks) {
        int panelX = this.surface.width2 - 199; // obf: var3 = -199 + li.u
        int boxY = 36;                           // obf: var4
        int boxW = 196;                          // obf: var5
        int textX = panelX + 3;                  // obf: var6
        int y = boxY + 15;                       // obf: var18

        // --- render the settings panel (section boxes + option/display/privacy rows) ---
        this.surface.drawSprite(panelX - 49, 3, this.spriteMedia + 6);          // obf: li.b(-1, tg+6, 3, var3-49)
        this.surface.drawBoxAlpha(panelX, boxY, 65, 9, 36, boxW, Surface.rgb(181, 181, 181));
        this.surface.drawBoxAlpha(panelX, boxY, 65, 0, 101, boxW, Surface.rgb(201, 201, 201));
        this.surface.drawBoxAlpha(panelX, boxY, 95, 0, 166, boxW, Surface.rgb(181, 181, 181));
        this.surface.drawBoxAlpha(panelX, boxY, this.optionExtraRow ? 55 : 40, 0, 261, boxW, Surface.rgb(201, 201, 201)); // obf: Kd

        // Game options header + the three toggle rows.
        this.surface.drawString(STRINGS[138], textX, y, 0, false, 1);                                           // header
        y += 15;
        this.surface.drawString(this.optionCameraModeAuto ? STRINGS[151] : STRINGS[136], textX, y, 0xFFFFFF, false, 1); // obf: Kh
        y += 15;
        this.surface.drawString(this.optionMouseButtonOne ? STRINGS[144] : STRINGS[146], textX, y, 0xFFFFFF, false, 1); // obf: Yh
        y += 15;
        if (this.membersServer) {
            this.surface.drawString(this.optionSoundDisabled ? STRINGS[155] : STRINGS[141], textX, y, 0xFFFFFF, false, 1); // obf: ne
        }
        y += 0; // obf: var18 += var1 (var1 == 0 junk) — display block starts here
        // Five non-clickable display rows (account-management blurb + referral line).
        this.surface.drawString(STRINGS[145], textX, y, 0xFFFFFF, false, 0);
        y += 15;
        this.surface.drawString(STRINGS[143], textX, y, 0xFFFFFF, false, 0);
        y += 15;
        this.surface.drawString(STRINGS[130], textX, y, 0xFFFFFF, false, 0);
        y += 15;
        // Referral line: referId==0 -> STRINGS[135], ==1 -> STRINGS[137], >=2 -> STRINGS[132].  obf: Yd
        if (this.referId == 0) {
            this.surface.drawString(STRINGS[135], textX, y, 0xFFFFFF, false, 0);
        } else if (this.referId == 1) {
            this.surface.drawString(STRINGS[137], textX, y, 0xFFFFFF, false, 0);
        } else {
            this.surface.drawString(STRINGS[132], textX, y, 0xFFFFFF, false, 0);
        }
        y += 15;
        y += 5;
        // Privacy section header + four toggle labels (on/off).
        this.surface.drawString(STRINGS[139], panelX + 3, y, 0, false, 1);
        y += 15;
        this.surface.drawString(STRINGS[133], panelX + 3, y, 0, false, 1);
        y += 15;
        this.surface.drawString(this.blockChatToggle != 0 ? STRINGS[153] : STRINGS[131], panelX + 3, y, 0xFFFFFF, false, 1);    // obf: De
        y += 15;
        this.surface.drawString(this.blockPrivateToggle == 0 ? STRINGS[142] : STRINGS[150], panelX + 3, y, 0xFFFFFF, false, 1); // obf: dc
        y += 15;
        this.surface.drawString(this.blockTradeToggle != 0 ? STRINGS[152] : STRINGS[140], panelX + 3, y, 0xFFFFFF, false, 1);   // obf: Vg
        y += 15;
        if (this.membersServer) {
            this.surface.drawString(this.blockDuelToggle != 0 ? STRINGS[154] : STRINGS[129], panelX + 3, y, 0xFFFFFF, false, 1); // obf: ui
        }
        y += 15;
        // Members-only extra row ("Remove all" style action), highlighted on hover.
        if (this.optionExtraRow) { // obf: Kd
            int colour = 0xFFFFFF;
            y += 5;
            if (this.mouseX > textX && this.mouseX < textX + boxW && this.mouseY > y - 12 && this.mouseY < y + 4) {
                colour = 0xFFFF00;
            }
            this.surface.drawString(STRINGS[134], textX, y, colour, false, 1);
            y += 15;
        }
        y += 5;
        // "Always logout when you finish" + clickable "Click here to logout" row.
        this.surface.drawString(STRINGS[147], textX, y, 0, false, 1);
        int logoutColour = 0xFFFFFF;
        y += 15;
        if (this.mouseX > textX && this.mouseX < textX + boxW && this.mouseY > y - 12 && this.mouseY < y + 4) {
            logoutColour = 0xFFFF00;
        }
        this.surface.drawString(STRINGS[149], panelX + 3, y, logoutColour, false, 1);

        if (!handleClicks) {
            return;
        }

        // --- hit-test the panel against the mouse (acts on a left-click, mouseButtonClick==1) ---
        int mx = this.mouseX - (this.surface.width2 - 199); // obf: var3
        int my = this.mouseY - 36;                          // obf: var15
        if (mx < 0 || my < 0 || mx >= 196 || my >= 265) {   // obf: 0<=var3 && var15>=0 && var3<196 && var15<265
            return; // click outside the settings panel
        }
        int rowX = (this.surface.width2 - 199) + 3; // obf: var6 = var9 + 3
        int rowY = 36 + 30;                          // obf: var18 = 30 + var10

        // Game-setting 0: camera / auto-screen-rotation toggle.
        if (this.mouseX > rowX && this.mouseX < rowX + boxW
                && this.mouseY > rowY - 12 && this.mouseY < rowY + 4 && this.mouseButtonClick == 1) {
            this.optionCameraModeAuto = !this.optionCameraModeAuto; // obf: Kh
            this.clientStream.newPacket(111); // GAME_SETTINGS_CHANGED
            this.clientStream.buffer.putByte(0); // setting index 0
            this.clientStream.buffer.putByte(this.optionCameraModeAuto ? 1 : 0);
            this.clientStream.sendPacket();
        }
        rowY += 15;
        // Game-setting 2: mouse-button (single/double) toggle.
        if (this.mouseX > rowX && this.mouseX < rowX + boxW
                && this.mouseY > rowY - 12 && this.mouseY < rowY + 4 && this.mouseButtonClick == 1) {
            this.optionMouseButtonOne = !this.optionMouseButtonOne; // obf: Yh
            this.clientStream.newPacket(111);
            this.clientStream.buffer.putByte(2); // setting index 2
            this.clientStream.buffer.putByte(this.optionMouseButtonOne ? 1 : 0);
            this.clientStream.sendPacket();
        }
        rowY += 15;
        // Game-setting 3: sound on/off (members only).
        if (this.membersServer && this.mouseX > rowX && this.mouseX < rowX + boxW
                && this.mouseY > rowY - 12 && this.mouseY < rowY + 4 && this.mouseButtonClick == 1) {
            this.optionSoundDisabled = !this.optionSoundDisabled; // obf: ne
            this.clientStream.newPacket(111);
            this.clientStream.buffer.putByte(3); // setting index 3
            this.clientStream.buffer.putByte(this.optionSoundDisabled ? 1 : 0);
            this.clientStream.sendPacket();
        }

        // Skip the five non-clickable display rows: obf does +=15 (x5) then +=35.
        rowY += 15;
        rowY += 15;
        rowY += 15;
        rowY += 15;
        rowY += 15;
        rowY += 35;

        // --- privacy block (4 toggles) -> one batched PRIVACY_SETTINGS_CHANGED packet ---
        boolean privacyChanged = false;
        if (this.mouseX > rowX && this.mouseX < rowX + boxW
                && this.mouseY > rowY - 12 && this.mouseY < rowY + 4 && this.mouseButtonClick == 1) {
            this.blockChatToggle = 1 - this.blockChatToggle; // obf: De
            privacyChanged = true;
        }
        rowY += 15;
        if (this.mouseX > rowX && this.mouseX < rowX + boxW
                && this.mouseY > rowY - 12 && this.mouseY < rowY + 4 && this.mouseButtonClick == 1) {
            this.blockPrivateToggle = 1 - this.blockPrivateToggle; // obf: dc
            privacyChanged = true;
        }
        rowY += 15;
        if (this.mouseX > rowX && this.mouseX < rowX + boxW
                && this.mouseY > rowY - 12 && this.mouseY < rowY + 4 && this.mouseButtonClick == 1) {
            this.blockTradeToggle = 1 - this.blockTradeToggle; // obf: Vg
            privacyChanged = true;
        }
        rowY += 15;
        if (this.membersServer && this.mouseX > rowX && this.mouseX < rowX + boxW
                && this.mouseY > rowY - 12 && this.mouseY < rowY + 4 && this.mouseButtonClick == 1) {
            this.blockDuelToggle = 1 - this.blockDuelToggle; // obf: ui
            privacyChanged = true;
        }
        rowY += 15;
        if (privacyChanged) {
            // Args land as (blockTrade=Vg, blockPrivate=dc, blockChat=De, junk=64, blockDuel=ui).
            this.sendPrivacySettings(this.blockTradeToggle, this.blockPrivateToggle,
                                     this.blockChatToggle, 64, this.blockDuelToggle);
        }

        // Members-only extra "remove all" row.
        if (this.optionExtraRow) { // obf: Kd
            rowY += 5;
            if (this.mouseX > rowX && this.mouseX < rowX + boxW
                    && this.mouseY > rowY - 12 && this.mouseY < rowY + 4 && this.mouseButtonClick == 1) {
                this.drawMenuOptions(ISAAC.g, -3, 9, false); // obf: a(o.g, var1-3, 9, false)  (o.g = static String[] menu list)
                this.menuItemsCount = 0; // obf: qc = 0
            }
            rowY += 15;
        }
        // "Click here to logout" row.
        rowY += 20;
        if (this.mouseX > rowX && this.mouseX < rowX + boxW
                && this.mouseY > rowY - 12 && this.mouseY < rowY + 4 && this.mouseButtonClick == 1) {
            this.requestLogout(0); // obf: B(0)  -> opcode 102 LOGOUT
        }
        this.mouseButtonClick = 0;
    }

    /**
     * Question/menu dialog handler: renders the answer options and, on click, sends the
     * chosen answer index.
     * Sends opcode 116 (QUESTION_DIALOG_ANSWER): the selected option index (one byte).
     *
     * Hit-test (both render and click modes, matching oracle drawOptionMenu):
     *   mouseX < textWidth(option) && mouseY > 12*i && mouseY < 12*i + 12
     */
    // obf: private final void G(int)   [int param var1 is anti-tamper junk]
    private void sendDialogAnswer(int unused) {
        if (this.mouseButtonClick == 0) { // obf: Cf == 0  -> render mode
            for (int i = 0; i < this.menuOptionCount; i++) { // obf: var6 < Id
                int colour = 0xFFFF; // yellow
                if (this.mouseX < this.surface.textWidth(this.menuOptions[i], 1)
                        && this.mouseY > i * 12 && this.mouseY < i * 12 + 12) {
                    colour = 0xFF0000; // red highlight under cursor
                }
                this.surface.drawString(this.menuOptions[i], 6, i * 12 + 12, colour, false, 1);
            }
            return;
        }
        // Click mode: find the clicked option and send its index.
        for (int i = 0; i < this.menuOptionCount; i++) {
            if (this.mouseX < this.surface.textWidth(this.menuOptions[i], 1)
                    && this.mouseY > i * 12 && this.mouseY < i * 12 + 12) {
                this.clientStream.newPacket(116); // QUESTION_DIALOG_ANSWER
                this.clientStream.buffer.putByte(i);
                this.clientStream.sendPacket();
                break;
            }
        }
        this.showDialogMenu = false;    // obf: Ph = false
        this.mouseButtonClick = 0;      // obf: Cf = 0
    }

    /**
     * Character-design screen controls: processes the Head / Hair / Top / Bottom / Skin and
     * gender arrow buttons (cycling the appearance indices, wrapping within each table) and,
     * when "Accept" is clicked, submits the chosen appearance and closes the screen.
     * Sends opcode 235 (PLAYER_APPEARANCE_CHANGE): gender, head, bodyGender, 2colour, hair,
     * top, bottom, skin (one byte each).
     *
     * obf class names for the appearance tables: n.m = FontWidths.appearanceFlags (per-sprite
     * gender/slot flags), na.e = StreamFactory.appearanceCount (#sprites). The colour palettes
     * are Dg (hair), ei (top+bottom), Wh (skin) — see oracle GameData.character*Colours.
     */
    // obf: private final void F(int)   [int param var1 is anti-tamper junk]
    private void sendAppearance(int unused) {
        // panelCharDesign.handleMouse(lastMouseButtonDown, mouseY, junk, mouseButtonDown, mouseX).
        this.panelCharDesign.handleMouse(this.lastMouseButtonDown, this.mouseY, -9989, this.mouseButtonDown, this.mouseX); // obf: Af.b(Bb, xb, -9989, Qb, I)

        // Head arrows: cycle appearanceHead to the next sprite valid for the current gender
        // (flag&3 == 1 means "head" slot; flag & 4*gender must be set).
        if (this.panelCharDesign.isClicked(this.charDesignHeadLeft)) {   // obf: Af.a(.., Dj)
            do {
                this.appearanceHead = (StreamFactory.appearanceCount + this.appearanceHead - 1) % StreamFactory.appearanceCount;
            } while ((FontWidths.appearanceFlags[this.appearanceHead] & 3) != 1
                     || (FontWidths.appearanceFlags[this.appearanceHead] & (4 * this.appearanceGender)) == 0);
        }
        if (this.panelCharDesign.isClicked(this.charDesignHeadRight)) {  // obf: Af.a(.., pi)
            do {
                this.appearanceHead = (this.appearanceHead + 1) % StreamFactory.appearanceCount;
            } while ((FontWidths.appearanceFlags[this.appearanceHead] & 3) != 1
                     || (FontWidths.appearanceFlags[this.appearanceHead] & (4 * this.appearanceGender)) == 0);
        }
        // Hair colour arrows.
        if (this.panelCharDesign.isClicked(this.charDesignHairLeft)) {   // obf: Af.a(.., Kj)
            this.appearanceHairColour = (this.charHairColours.length + this.appearanceHairColour - 1) % this.charHairColours.length;
        }
        if (this.panelCharDesign.isClicked(this.charDesignHairRight)) {  // obf: Af.a(.., ed)
            this.appearanceHairColour = (this.appearanceHairColour + 1) % this.charHairColours.length;
        }
        // Gender arrows: flip gender, then re-seek head (flag&3==1) and bodyGender (flag&3==2).
        if (this.panelCharDesign.isClicked(this.charDesignGenderLeft)   // obf: Af.a(.., Ge)
                || this.panelCharDesign.isClicked(this.charDesignGenderRight)) { // obf: Af.a(.., Of)
            this.appearanceGender = 3 - this.appearanceGender;
            while ((FontWidths.appearanceFlags[this.appearanceHead] & 3) != 1
                    || (FontWidths.appearanceFlags[this.appearanceHead] & (4 * this.appearanceGender)) == 0) {
                this.appearanceHead = (this.appearanceHead + 1) % StreamFactory.appearanceCount;
            }
            while ((FontWidths.appearanceFlags[this.appearanceBodyGender] & 3) != 2
                    || (FontWidths.appearanceFlags[this.appearanceBodyGender] & (4 * this.appearanceGender)) == 0) {
                this.appearanceBodyGender = (this.appearanceBodyGender + 1) % StreamFactory.appearanceCount;
            }
        }
        // Top colour arrows.
        if (this.panelCharDesign.isClicked(this.charDesignTopLeft)) {    // obf: Af.a(.., Xc)
            this.appearanceTopColour = (this.appearanceTopColour - 1 + this.charTopBottomColours.length) % this.charTopBottomColours.length;
        }
        if (this.panelCharDesign.isClicked(this.charDesignTopRight)) {   // obf: Af.a(.., ek)
            this.appearanceTopColour = (this.appearanceTopColour + 1) % this.charTopBottomColours.length;
        }
        // Skin colour arrows.
        if (this.panelCharDesign.isClicked(this.charDesignSkinLeft)) {   // obf: Af.a(.., Ze)
            this.appearanceSkinColour = (this.charSkinColours.length + this.appearanceSkinColour - 1) % this.charSkinColours.length;
        }
        if (this.panelCharDesign.isClicked(this.charDesignSkinRight)) {  // obf: Af.a(.., Mj)
            this.appearanceSkinColour = (this.appearanceSkinColour + 1) % this.charSkinColours.length;
        }
        // Bottom colour arrows.
        if (this.panelCharDesign.isClicked(this.charDesignBottomLeft)) { // obf: Af.a(.., Re)
            this.appearanceBottomColour = (this.charTopBottomColours.length + this.appearanceBottomColour - 1) % this.charTopBottomColours.length;
        }
        if (this.panelCharDesign.isClicked(this.charDesignBottomRight)) {// obf: Af.a(.., Ai)
            this.appearanceBottomColour = (this.appearanceBottomColour + 1) % this.charTopBottomColours.length;
        }

        // "Accept" button: submit the new appearance.
        if (!this.panelCharDesign.isClicked(this.charDesignAccept)) {    // obf: Af.a(.., Eg)
            return;
        }
        this.clientStream.newPacket(235); // PLAYER_APPEARANCE_CHANGE
        this.clientStream.buffer.putByte(this.appearanceGender);      // obf: Sf
        this.clientStream.buffer.putByte(this.appearanceHead);        // obf: Vd
        this.clientStream.buffer.putByte(this.appearanceBodyGender);  // obf: dk
        this.clientStream.buffer.putByte(this.appearance2Colour);     // obf: wg
        this.clientStream.buffer.putByte(this.appearanceHairColour);  // obf: ld
        this.clientStream.buffer.putByte(this.appearanceTopColour);   // obf: Wg
        this.clientStream.buffer.putByte(this.appearanceBottomColour);// obf: Lh
        this.clientStream.buffer.putByte(this.appearanceSkinColour);  // obf: hh
        this.clientStream.sendPacket();
        this.surface.blackScreen();          // obf: li.a(true)
        this.showAppearanceChange = false;   // obf: Kg = false
    }

    /**
     * Combat-style tab: renders the five rows (header + Controlled / Aggressive / Accurate /
     * Defensive) and, when one of the four style rows is clicked, selects it and informs the
     * server.
     * Sends opcode 29 (COMBAT_STYLE_CHANGED): the selected style index (one byte, 0..3).
     *
     * Hit-test (matching oracle drawDialogCombatStyle): row index 1..4 only,
     *   mouseX > boxX && mouseX < boxX+boxW && mouseY > boxY+20*row && mouseY < boxY+20*row+20
     */
    // obf: private final void k(byte)   [byte param var1 is anti-tamper junk]
    private void sendCombatStyle(byte unused) {
        int boxX = 7;    // obf: var2
        int boxY = 15;   // obf: var3
        int boxW = 175;  // obf: var4

        // Click mode: hit-test the four style rows (row 0 is the header).
        if (this.mouseButtonClick != 0) { // obf: Cf != 0
            for (int row = 0; row < 5; row++) {
                if (row > 0
                        && this.mouseX > boxX && this.mouseX < boxX + boxW
                        && this.mouseY > boxY + row * 20 && this.mouseY < boxY + row * 20 + 20) {
                    this.combatStyle = row - 1; // obf: Fg = var5 - 1
                    this.mouseButtonClick = 0;  // obf: Cf = 0
                    this.clientStream.newPacket(29); // COMBAT_STYLE_CHANGED
                    this.clientStream.buffer.putByte(this.combatStyle);
                    this.clientStream.sendPacket();
                    break;
                }
            }
        }
        // Render the five rows (selected style row highlighted red) + labels.
        for (int row = 0; row < 5; row++) {
            int fill = (row == this.combatStyle + 1) ? Surface.rgb(255, 0, 0) : Surface.rgb(190, 190, 190);
            this.surface.drawBoxAlpha(boxX, boxY + row * 20, boxW, 20, fill, 128);
            this.surface.drawLineHoriz(boxX, boxY + row * 20, boxW, 0);
            this.surface.drawLineHoriz(boxX, boxY + row * 20 + 20, boxW, 0);
        }
        this.surface.drawStringCenter(STRINGS[650], boxX + boxW / 2, boxY + 16, 3, 0xFFFFFF); // header "Select combat style"
        this.surface.drawStringCenter(STRINGS[648], boxX + boxW / 2, boxY + 36, 3, 0);        // Controlled
        this.surface.drawStringCenter(STRINGS[645], boxX + boxW / 2, boxY + 56, 3, 0);        // Aggressive
        this.surface.drawStringCenter(STRINGS[649], boxX + boxW / 2, boxY + 76, 3, 0);        // Accurate
        this.surface.drawStringCenter(STRINGS[647], boxX + boxW / 2, boxY + 96, 3, 0);        // Defensive
    }

    /**
     * Remove items of one type from the current DUEL offer, then resend the whole offer.
     * For a stackable item (fa.e==0) the single offer entry's count is decremented; for a
     * non-stackable item (fa.e!=0) up to {@code amount} matching entries are dropped.
     * Resending clears both duel-accept flags so the offer must be re-accepted.
     * Sends opcode 33 (DUEL_OFFER_ITEM): item count, then per item: id (short) + qty (int).
     */
    // obf: private final void a(int,int,byte)   [byte param var3 is anti-tamper guard: send only if var3 == -78]
    private void sendDuelOffer(int slot, int qty, byte unused) {
        int itemId = this.duelOfferItemId[slot];                       // obf: Uf[var1]
        int amount = (qty >= 0) ? qty : this.defaultItemAmount;        // obf: ~var2<=-1 ? var2 : Tk

        if (GameData.itemStackable[itemId] == 0) {                     // obf: fa.e[var4] == 0  -> STACKABLE
            // Stackable: one entry, decrement its count; remove the slot if it empties.
            this.duelOfferItemCount[slot] -= amount;                   // obf: df[var1] -= var5
            if (this.duelOfferItemCount[slot] <= 0) {                  // obf: !(0 < df[var1])
                this.duelOfferItemsCount--;                            // obf: Ke--
                for (int j = slot; j < this.duelOfferItemsCount; j++) {
                    this.duelOfferItemId[j]    = this.duelOfferItemId[j + 1];
                    this.duelOfferItemCount[j] = this.duelOfferItemCount[j + 1];
                }
            }
        } else {
            // Non-stackable: each unit is its own entry; drop up to `amount` matching entries.
            int removed = 0;                                           // obf: var11
            for (int i = 0; i < this.duelOfferItemsCount && removed < amount; i++) { // obf: var7<Ke && ~var5>=~var11
                if (this.duelOfferItemId[i] == itemId) {               // obf: Uf[var7] == var4
                    this.duelOfferItemsCount--;
                    removed++;
                    for (int j = i; j < this.duelOfferItemsCount; j++) {
                        this.duelOfferItemId[j]    = this.duelOfferItemId[j + 1];
                        this.duelOfferItemCount[j] = this.duelOfferItemCount[j + 1];
                    }
                    i--;
                }
            }
        }

        this.clientStream.newPacket(33); // DUEL_OFFER_ITEM
        this.clientStream.buffer.putByte(this.duelOfferItemsCount); // obf: Jh.f.c(Ke, ..)
        for (int i = 0; i < this.duelOfferItemsCount; i++) {
            this.clientStream.buffer.putShort(this.duelOfferItemId[i]);     // obf: Jh.f.e(.., Uf[var12])
            this.clientStream.buffer.putInt((int) this.duelOfferItemCount[i]); // obf: Jh.f.b(.., df[var12])
        }
        this.clientStream.sendPacket();

        this.duelOfferAccepted = false;          // obf: ke = false (ours)
        this.duelOfferRecipientAccepted = false; // obf: ki = false (theirs)
    }

    /**
     * Remove items of one type from the current TRADE offer, then resend the whole offer
     * (mirrors {@link #sendDuelOffer}). Stackable -> decrement the entry's count; non-stackable
     * -> drop up to {@code amount} matching entries. Resending clears both trade-accept flags.
     * Sends opcode 46 (PLAYER_ADDED_ITEMS_TO_TRADE_OFFER): item count, then per item:
     * id (short) + qty (int).
     */
    // obf: private final void c(int,byte,int)   [byte param var2 is anti-tamper guard: send only if var2 > 120]
    private void sendTradeOffer(int qty, byte unused, int slot) {
        int itemId = this.tradeItems[slot];                      // obf: Qf[var3]
        int amount = (qty < 0) ? this.defaultItemAmount : qty;   // obf: var1<0 ? Tk : var1

        if (GameData.itemStackable[itemId] != 0) {               // obf: fa.e[var4] != 0  -> NON-stackable
            // Non-stackable: drop up to `amount` matching offer entries.
            int removed = 0;                                     // obf: var6
            for (int i = 0; i < this.tradeItemsCount && removed < amount; i++) { // obf: var7<mf && var6<var5
                if (this.tradeItems[i] == itemId) {              // obf: ~Qf[var7] == ~var4
                    removed++;
                    this.tradeItemsCount--;
                    for (int j = i; j < this.tradeItemsCount; j++) {
                        this.tradeItems[j]     = this.tradeItems[j + 1];
                        this.tradeItemCount[j] = this.tradeItemCount[j + 1];
                    }
                    i--;
                }
            }
        } else {
            // Stackable: one entry, decrement its count; remove the slot if it empties.
            this.tradeItemCount[slot] -= amount;                 // obf: jj[var3] -= var5
            if (this.tradeItemCount[slot] <= 0) {                // obf: -1 <= ~jj[var3]
                this.tradeItemsCount--;
                for (int j = slot; j < this.tradeItemsCount; j++) {
                    this.tradeItems[j]     = this.tradeItems[j + 1];
                    this.tradeItemCount[j] = this.tradeItemCount[j + 1];
                }
            }
        }

        this.clientStream.newPacket(46); // PLAYER_ADDED_ITEMS_TO_TRADE_OFFER
        this.clientStream.buffer.putByte(this.tradeItemsCount); // obf: Jh.f.c(mf, ..)
        for (int i = 0; i < this.tradeItemsCount; i++) {
            this.clientStream.buffer.putShort(this.tradeItems[i]);          // obf: Jh.f.e(393, Qf[var12])
            this.clientStream.buffer.putInt((int) this.tradeItemCount[i]);  // obf: Jh.f.b(.., jj[var12])
        }
        this.clientStream.sendPacket();

        this.tradeOfferAccepted = false;          // obf: Mi = false (ours)
        this.tradeOfferRecipientAccepted = false; // obf: md = false (theirs)
    }

    /**
     * Final logout/close acknowledgement: when {@code send} is set, fires the close packet
     * and tears the connection down; then clears the cached credentials and resets the
     * username/password entry state.
     * Sends opcode 31 (CONFIRM_LOGOUT / CLOSE_CONNECTION).
     */
    // obf: private final void a(boolean,int)   [int param var2 is anti-tamper junk: if(var2!=31){sf=null}]
    private void sendConfirmLogoutAck(boolean send, int unused) {
        if (send && this.clientStream != null) { // obf: var1 && Jh != null
            try {
                this.clientStream.newPacket(31);     // CONFIRM_LOGOUT
                this.clientStream.closeStream(true); // obf: Jh.a(-6924)  flush + close socket/writer thread
            } catch (IOException ignored) {
            }
        }
        this.password = "";  // obf: wh = ""
        this.username = "";  // obf: Xf = ""
        this.resetIntroState(-2); // obf: o(var2 ^ -31) -> o(-2): clears entry-cursor state (incl. kc)
    }

    /**
     * The trade-offer window: renders both offers + your inventory and routes clicks. This is
     * primarily a UI method (the {@code drawTrade} renderer) but it is also one of the
     * trade-packet producers, so the packet-sending control flow is captured exactly here:
     *   - opcode 230 (PLAYER_DECLINED_TRADE) on a click outside the window or the Decline button
     *   - opcode 55  (PLAYER_ACCEPTED_INIT_TRADE_REQUEST) on the Accept button
     *   - adding an inventory item routes through {@link #drawTradeConfirm} (opcode 46)
     *   - removing an offered item routes through {@link #sendTradeOffer} (opcode 46)
     *   - a right-click over an item opens the {@code tradeItemMenu} (Wf) context menu, whose
     *     chosen entry then calls drawTradeConfirm / sendTradeOffer.
     */
    // obf: private final void n(byte)   [byte param var1 is anti-tamper junk]
    private void drawTrade(byte unused) {
        // Resolve a right-click context-menu selection, if its popup (showTradeItemMenu) is open.
        int menuHit = -1; // obf: var2
        if (this.mouseButtonClick != 0 && this.showTradeItemMenu) { // obf: Cf != 0 && lh
            menuHit = this.tradeItemMenu.hitTestNoRender(this.mouseX, this.tradeMenuX, this.tradeMenuY, (byte) -40, this.mouseY); // obf: Wf.b(I, Gf, Bf, .., xb)
        }

        if (menuHit < 0 && this.tradeConfirmShown == 0) { // obf: var2 < 0 && gc == 0
            // Fresh left-click with no amount set yet -> default to 1.
            if (this.mouseButtonClick == 1 && this.defaultItemAmount == 0) { // obf: Cf == 1 && Tk == 0
                this.defaultItemAmount = 1;
            }
            int relX = this.mouseX - 22; // obf: var3
            int relY = this.mouseY - 36; // obf: var4

            // Click OUTSIDE the 468x262 window -> decline the trade.
            if (relX < 0 || relY < 0 || relX >= 468 || relY >= 262) {
                if (this.mouseButtonClick != 0) { // obf: ~Cf != -1
                    this.tradeWindowOpen = false;       // obf: Hk = false
                    this.clientStream.newPacket(230);   // PLAYER_DECLINED_TRADE
                    this.clientStream.sendPacket();
                }
            } else if (this.defaultItemAmount > 0) { // obf: ~Tk < -1  ⟺  Tk > 0  (left-click actions)
                // Click in YOUR INVENTORY column (x 216..462, y 30..235) -> add the item.
                if (relX > 216 && relY > 30 && relX < 462 && relY < 235) {
                    int slot = 5 * ((relY - 31) / 34) + (relX - 217) / 49;
                    if (slot >= 0 && slot < this.inventoryItemsCount) { // obf: var5 < lc
                        this.drawTradeConfirm(-1, (byte) 9, slot); // obf: a(-1, 9, var5)  -> add 1, opcode 46
                    }
                }
                // Click in YOUR OFFER column (x 8..205, y 30..133) -> remove the offered item.
                if (relX > 8 && relY > 30 && relX < 205 && relY < 133) {
                    int slot = (relY - 31) / 34 * 4 + (relX - 9) / 49;
                    if (slot >= 0 && slot < this.tradeItemsCount) { // obf: var17 < mf
                        this.sendTradeOffer(-1, (byte) 125, slot); // obf: c(-1, 125, var17)  -> remove 1, opcode 46
                    }
                }
                // "Accept" button (x 217..286, y 238..259).
                if (relX >= 217 && relY >= 238 && relX <= 286 && relY <= 259) {
                    this.tradeOfferAccepted = true; // obf: Mi = true
                    this.clientStream.newPacket(55); // PLAYER_ACCEPTED_INIT_TRADE_REQUEST
                    this.clientStream.sendPacket();
                }
                // "Decline" button (x 394..462, y 238..258).
                if (relX >= 394 && relY >= 238 && relX < 463 && relY < 259) {
                    this.tradeWindowOpen = false;     // obf: Hk = false
                    this.clientStream.newPacket(230); // PLAYER_DECLINED_TRADE
                    this.clientStream.sendPacket();
                }
                this.defaultItemAmount = 0; // obf: Tk = 0
                this.mouseButtonClick = 0;  // obf: Cf = 0
            }

            // Right-click (mouseButtonClick == 2) over an item -> open its context menu.
            if (this.mouseButtonClick == 2) { // obf: Cf == 2
                // Right-click in the inventory column -> "Offer / Offer-5 / Offer-10 / Offer-X / Offer-All".
                if (relX > 216 && relY > 30 && relX < 462 && relY < 235) {
                    // Pre-size a context-menu box off the shared friendsList probe (obf: zh) — the
                    // actual menu is built on tradeItemMenu (Wf) just below.
                    int menuW = this.friendsList.getPanelWidth(16256);   // obf: zh.b(16256)
                    int menuH = this.friendsList.getPanelHeight(-21224); // obf: zh.a(-21224)
                    this.contextMenuY = this.mouseY - 7;       // obf: fg
                    this.contextMenuX = this.mouseX - menuW / 2;// obf: rh
                    this.contextMenuOpen = true;               // obf: se = true
                    if (this.contextMenuY < 0) this.contextMenuY = 0;
                    if (this.contextMenuX < 0) this.contextMenuX = 0;
                    if (this.contextMenuX + menuW > 510) this.contextMenuX = 510 - menuW;
                    if (this.contextMenuY + menuH > 315) this.contextMenuY = 315 - menuH;
                    int slot = (relY - 31) / 34 * 5 + (relX - 217) / 49;
                    if (slot >= 0 && slot < this.inventoryItemsCount) {
                        int id = this.inventoryItemId[slot]; // obf: vf[var7]
                        this.showTradeItemMenu = true;       // obf: lh = true
                        this.tradeItemMenu.setCount(0);      // obf: Wf.d(0)
                        // addEntryWithColor(color=id, prefix, x=layer?, message, layer, guard).
                        this.tradeItemMenu.addEntryWithColor(id, STRINGS[34] + GameData.itemName[id], 1, STRINGS[172], 1, 3288);  // Offer
                        this.tradeItemMenu.addEntryWithColor(id, STRINGS[34] + GameData.itemName[id], 1, STRINGS[169], 5, 3296);  // Offer-5
                        this.tradeItemMenu.addEntryWithColor(id, STRINGS[34] + GameData.itemName[id], 1, STRINGS[158], 10, 3296); // Offer-10
                        this.tradeItemMenu.addEntryWithColor(id, STRINGS[34] + GameData.itemName[id], 1, STRINGS[174], -1, 3296); // Offer-X
                        this.tradeItemMenu.addEntryWithColor(id, STRINGS[34] + GameData.itemName[id], 1, STRINGS[166], -2, 3304); // Offer-All
                        int w = this.tradeItemMenu.getPanelWidth(16256);
                        int h = this.tradeItemMenu.getPanelHeight(-21224);
                        this.tradeMenuX = this.mouseX - w / 2; // obf: Gf
                        this.tradeMenuY = this.mouseY - 7;     // obf: Bf
                        if (this.tradeMenuX < 0) this.tradeMenuX = 0;
                        if (this.tradeMenuY < 0) this.tradeMenuY = 0;
                        if (this.tradeMenuY + h > 315) this.tradeMenuY = 315 - h;
                        if (this.tradeMenuX + w > 510) this.tradeMenuX = 510 - w;
                    }
                }
                // Right-click in YOUR offer column -> "Remove / Remove-5 / Remove-10 / Remove-X / Remove-All".
                if (relX > 8 && relY > 30 && relX < 205 && relY < 133) {
                    int slot = (relX - 9) / 49 + (relY - 31) / 34 * 4;
                    if (slot >= 0 && slot < this.tradeItemsCount) {
                        int id = this.tradeItems[slot]; // obf: Qf[var19]
                        this.showTradeItemMenu = true;
                        this.tradeItemMenu.setCount(0);
                        this.tradeItemMenu.addEntryWithColor(id, STRINGS[34] + GameData.itemName[id], 2, STRINGS[163], 1, 3296);  // Remove
                        this.tradeItemMenu.addEntryWithColor(id, STRINGS[34] + GameData.itemName[id], 2, STRINGS[173], 5, 3304);  // Remove-5
                        this.tradeItemMenu.addEntryWithColor(id, STRINGS[34] + GameData.itemName[id], 2, STRINGS[161], 10, 3296); // Remove-10
                        this.tradeItemMenu.addEntryWithColor(id, STRINGS[34] + GameData.itemName[id], 2, STRINGS[177], -1, 3296); // Remove-X
                        this.tradeItemMenu.addEntryWithColor(id, STRINGS[34] + GameData.itemName[id], 2, STRINGS[170], -2, 3304); // Remove-All
                        int w = this.tradeItemMenu.getPanelWidth(16256);
                        int h = this.tradeItemMenu.getPanelHeight(-21224);
                        this.tradeMenuX = this.mouseX - w / 2;
                        this.tradeMenuY = this.mouseY - 7;
                        if (this.tradeMenuX < 0) this.tradeMenuX = 0;
                        if (this.tradeMenuY < 0) this.tradeMenuY = 0;
                        if (this.tradeMenuY + h > 315) this.tradeMenuY = 315 - h;
                        if (this.tradeMenuX + w > 510) this.tradeMenuX = 510 - w;
                    }
                }
                this.mouseButtonClick = 0; // obf: Cf = 0
            }

            // Auto-close the context menu when the cursor leaves its bounds.
            if (this.showTradeItemMenu) {
                int w = this.tradeItemMenu.getPanelWidth(16256);
                int h = this.tradeItemMenu.getPanelHeight(-21224);
                if (this.mouseX < this.tradeMenuX - 10 || this.mouseY < this.tradeMenuY - 10
                        || this.mouseX > this.tradeMenuX + w + 10 || this.mouseY > this.tradeMenuY + h + 10) {
                    this.showTradeItemMenu = false;
                }
            }
        } else if (menuHit >= 0) {
            // A context-menu entry was clicked: resolve it (its stored params decide add vs remove).
            this.showTradeItemMenu = false; // obf: lh = false
            this.mouseButtonClick = 0;      // obf: Cf = 0
            int action = this.tradeItemMenu.getEntryXPos(-91, menuHit);    // obf: Wf.a(-91, var2)  -> sortKey (1=offer, 2=remove)
            int id = this.tradeItemMenu.getEntryColorE(true, menuHit);     // obf: Wf.a(true, var2)  -> color field holds the item id
            int firstSlot = -1; // obf: var21
            int totalCount = 0; // obf: var25
            if (action == 1) {
                // Offer: find the inventory slot(s) holding this item; total the count.
                for (int s = 0; s < this.inventoryItemsCount; s++) { // obf: var28 < lc
                    if (this.inventoryItemId[s] == id) {
                        if (firstSlot < 0) {
                            firstSlot = s;
                        }
                        if (GameData.itemStackable[id] == 0) { // stackable: count is the stack size
                            totalCount = this.inventoryItemStackCount[s];
                        }
                        totalCount++;
                    }
                }
            } else {
                // Remove: find the offer slot(s) holding this item; total the offered count.
                for (int s = 0; s < this.tradeItemsCount; s++) { // obf: var28 < mf
                    if (id == this.tradeItems[s]) {
                        if (firstSlot < 0) {
                            firstSlot = s;
                        }
                        if (GameData.itemStackable[id] == 0) { // ~fa.e == -1 ⟺ stackable
                            totalCount = this.tradeItemCount[s];
                        }
                        totalCount++;
                    }
                }
            }
            if (firstSlot >= 0) {
                int requested = this.tradeItemMenu.getEntryColorCode((byte) 97, menuHit); // obf: Wf.a((byte)97, var2)  -> colorCode holds the amount
                if (requested == -2) { // "X": open the quantity-entry list
                    this.menuTargetSlot = firstSlot; // obf: ji = var21
                    if (action == 1) {
                        this.drawMenuOptions(s.e, 12, 1, true);     // obf: a(s.e, 12, 1, true) — "Offer X?" entry list
                    } else {
                        this.drawMenuOptions(Surface.Kb, 4, 2, true); // obf: a(ua.Kb, var1^4, 2, true) — "Remove X?" entry list
                    }
                }
                if (requested == -1) {
                    requested = totalCount; // "X" with no amount -> use the full count
                }
                if (action != 1) {
                    this.sendTradeOffer(requested, (byte) 124, firstSlot); // obf: c(var28, 124, var21) — remove
                }
                this.drawTradeConfirm(requested, (byte) 9, firstSlot); // obf: a(var28, 9, var21) — add
            }
        }

        if (this.tradeWindowOpen) { // obf: if (Hk)
            int x = 22; // obf: var14
            int y = 36; // obf: var16
            this.surface.drawBox(x, y, 468, 12, 192);                              // obf: li.a(var14, .., 192, var16, 12, 468)
            this.surface.drawBoxAlpha(x, y + 12, 468, 18, 0x989898, 160);
            this.surface.drawBoxAlpha(x, y + 30, 8, 248, 0x989898, 160);
            this.surface.drawBoxAlpha(x + 205, y + 30, 11, 248, 0x989898, 160);
            this.surface.drawBoxAlpha(x + 462, y + 30, 6, 248, 0x989898, 160);
            this.surface.drawBoxAlpha(x + 8, y + 133, 197, 22, 0x989898, 160);
            this.surface.drawBoxAlpha(x + 8, y + 258, 197, 20, 0x989898, 160);
            this.surface.drawBoxAlpha(x + 216, y + 235, 246, 43, 0x989898, 160);
            this.surface.drawBoxAlpha(x + 8, y + 30, 197, 103, 0xD0D0D0, 160);
            this.surface.drawBoxAlpha(x + 8, y + 155, 197, 103, 0xD0D0D0, 160);
            this.surface.drawBoxAlpha(x + 216, y + 30, 246, 205, 0xD0D0D0, 160);
            for (int r = 0; r < 4; r++) {
                this.surface.drawLineHoriz(x + 8, y + 30 + r * 34, 197, 0);   // obf: li.b(197, 0, 8+var14, 30+var16+34*var29)
            }
            for (int r = 0; r < 4; r++) {
                this.surface.drawLineHoriz(x + 8, y + 155 + r * 34, 197, 0);
            }
            for (int r = 0; r < 7; r++) {
                this.surface.drawLineHoriz(x + 216, y + 30 + r * 34, 246, 0);
            }
            for (int c = 0; c < 6; c++) {
                if (c < 5) {
                    this.surface.drawLineVert(x + 8 + c * 49, y + 30, 103, 0);
                }
                if (c < 5) {
                    this.surface.drawLineVert(x + 8 + c * 49, y + 155, 103, 0);
                }
                this.surface.drawLineVert(x + 216 + c * 49, y + 30, 205, 0);
            }
            this.surface.drawString(STRINGS[175] + this.tradeRecipientName, x + 1, y + 10, 0xFFFFFF, false, 1); // "Trading with: "
            this.surface.drawString(STRINGS[164], x + 9, y + 27, 0xFFFFFF, false, 4);  // "Your Offer"
            this.surface.drawString(STRINGS[167], x + 9, y + 152, 0xFFFFFF, false, 4); // "Opponent's Offer"
            this.surface.drawString(STRINGS[171], x + 216, y + 27, 0xFFFFFF, false, 4);// "Your Inventory"
            if (!this.tradeOfferAccepted) { // obf: if (!Mi)
                this.surface.drawSprite(x + 217, y + 238, this.spriteMedia + 25); // Accept button
            }
            this.surface.drawSprite(x + 394, y + 238, this.spriteMedia + 26);     // Decline button
            if (this.tradeRecipientAccepted) { // obf: if (md)
                this.surface.drawStringCenter(STRINGS[168], x + 341, y + 246, 1, 0xFFFFFF); // "Other player"
                this.surface.drawStringCenter(STRINGS[165], x + 341, y + 256, 1, 0xFFFFFF); // "has accepted"
            }
            if (this.tradeOfferAccepted) { // obf: if (Mi)
                this.surface.drawStringCenter(STRINGS[176], x + 217 + 35, y + 246, 1, 0xFFFFFF); // "Waiting for"
                this.surface.drawStringCenter(STRINGS[160], x + 252, y + 256, 1, 0xFFFFFF);      // "other player"
            }
            // Your inventory grid (5 columns).
            for (int i = 0; i < this.inventoryItemsCount; i++) { // obf: var40 = ~lc loop
                int slotX = x + 217 + (i % 5) * 49;
                int slotY = y + 31 + (i / 5) * 34;
                this.surface.drawItemSprite(slotX, slotY, 48, 32,
                        this.spriteItem + GameData.itemPicture[this.inventoryItemId[i]],
                        GameData.itemMask[this.inventoryItemId[i]], 0, 0, false); // obf: li.a(..h.c[vf], ua.Bb[vf]+sg..)
                if (GameData.itemStackable[this.inventoryItemId[i]] == 0) {
                    this.surface.drawString("" + this.inventoryItemStackCount[i], slotX + 1, slotY + 10, 0xFFFF00, false, 1);
                }
            }
            // Your offer grid (4 columns).
            for (int i = 0; i < this.tradeItemsCount; i++) { // obf: var40 = mf loop
                int slotX = x + 9 + (i % 4) * 49;
                int slotY = y + 31 + (i / 4) * 34;
                this.surface.drawItemSprite(slotX, slotY, 48, 32,
                        this.spriteItem + GameData.itemPicture[this.tradeItems[i]],
                        GameData.itemMask[this.tradeItems[i]], 0, 0, false);
                if (GameData.itemStackable[this.tradeItems[i]] == 0) {
                    this.surface.drawString("" + this.tradeItemCount[i], slotX + 1, slotY + 10, 0xFFFF00, false, 1);
                }
                if (this.mouseX > slotX && this.mouseX < slotX + 48 && this.mouseY > slotY && this.mouseY < slotY + 32) {
                    this.surface.drawString(GameData.itemName[this.tradeItems[i]] + STRINGS[159] + GameData.itemDescription[this.tradeItems[i]],
                            x + 8, y + 273, 0xFFFF00, false, 1);
                }
            }
            // Opponent's offer grid (4 columns).
            for (int i = 0; i < this.tradeRecipientItemsCount; i++) { // obf: var40 = Lk loop
                int slotX = x + 9 + (i % 4) * 49;
                int slotY = y + 156 + (i / 4) * 34;
                this.surface.drawItemSprite(slotX, slotY, 48, 32,
                        this.spriteItem + GameData.itemPicture[this.tradeRecipientItems[i]],
                        GameData.itemMask[this.tradeRecipientItems[i]], 0, 0, false);
                if (GameData.itemStackable[this.tradeRecipientItems[i]] == 0) {
                    this.surface.drawString("" + this.tradeRecipientItemCount[i], slotX + 1, slotY + 10, 0xFFFF00, false, 1);
                }
                if (this.mouseX > slotX && this.mouseX < slotX + 48 && this.mouseY > slotY && this.mouseY < slotY + 32) {
                    this.surface.drawString(GameData.itemName[this.tradeRecipientItems[i]] + STRINGS[159] + GameData.itemDescription[this.tradeRecipientItems[i]],
                            x + 8, y + 273, 0xFFFF00, false, 1);
                }
            }

            if (this.showTradeItemMenu) { // obf: if (lh)
                // hitTest doubles as the renderer: it draws the panel and returns the hovered index.
                this.tradeItemMenu.hitTest(this.tradeMenuY, this.tradeMenuX, this.mouseY, (byte) -12, this.mouseX); // obf: Wf.a(Bf, Gf, xb, .., I)
            }
        }
    }
