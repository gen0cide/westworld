    // ===== login =====
    //
    // Session / login / registration block: opens the ClientStream, performs the
    // ISAAC-seed + RSA-encrypted login handshake, decodes the server's login
    // response code, renders the login UI (welcome screen + username/password
    // entry), and tears the connection down on loss.
    //
    // NOTE on STRINGS[] (`il[]`): the obfuscated client stores all UI / login-response
    // text XOR-encrypted in STRINGS[]. The plaintext beside each index below was
    // recovered by cross-referencing the canonical RSC login flow
    // (mudclient204 GameConnection.login and OpenRSC mudclient.showMessage), whose
    // response-code -> message mapping is byte-identical to this client. Indices are
    // exact; the quoted text is the canonical wording.
    //
    // Helper Mudclient methods called below that are outside the login group (obf -> name):
    //   b(byte,String,String)   -> showLoginScreenStatus  (two-line login banner)
    //   a(String,byte,String)   -> drawTextBox            (overlay box for reconnect)
    //   o(int)                  -> resetLoginState        (clears bj/Xd/qg, and kc when -2)
    //   i(int)                  -> resetGameState         (post-login scene/UI reset, qg=1)
    //   g(int)                  -> onSessionNeedsVerify    (resp==1 verify/sleep hook; mostly stub)
    //   c(byte)                 -> sendQueuedActions      (skeleton mainloop name)
    //   G(int)                  -> sendDialogAnswer       (skeleton packetout name)
    //   a(int,int,String)       -> createSocket           (open world socket)

    /**
     * Connect to the world server and perform the full login handshake for
     * {@code username}/{@code password}.
     *
     * Flow: open a ClientStream socket, send opcode 0 (LOGIN) carrying an
     * RSA-encrypted block (4 random session-key words + username) plus an
     * XTEA-encrypted tail (UID + password), seed the ISAAC stream cipher, then read
     * and decode the one-byte login response. {@code reconnecting} drives the silent
     * auto-relogin path used after a dropped connection.
     *
     * @param dummy        anti-tamper constant (callers pass -12); only used masked,
     *                     e.g. {@code dummy ^ 0xFFFFFFF4 == 0} yields the LOGIN opcode
     * @param username     account name (raw; trimmed + truncated to 20 chars)
     * @param password     account password (truncated to 20 chars)
     * @param reconnecting true for a silent re-establish after lost connection
     *
     * obf: void a(int,String,String,boolean)  [proposed: loginUser]
     */
    private final void loginUser(int dummy, String username, String password, boolean reconnecting) {
        // If this world reported "currently full" recently, refuse to even try and
        // show the full-world banner after a short pause.
        if (this.worldFullTimeout > 0) {
            this.showLoginScreenStatus(STRINGS[436], STRINGS[432]); // "Please wait..." / "Connecting to server"
            try {
                Utility.sleep(11200, 2000L);
            } catch (Exception ignored) {
            }
            this.showLoginScreenStatus(STRINGS[422], STRINGS[454]); // "Sorry! The server is currently full." / "Please try again later"
            return;
        }

        // Retry/attempt loop. NOTE: the obf field `worldIndex` (Vh) is overloaded —
        // it is both the world/port selector (see port pick below) and the remaining
        // auto-login attempt counter. The reconnect path (closeConnection) presets it
        // to 10; the entry-panel submit path sets it to 2; each failure decrements it.
        while (this.worldIndex > 0) {
            try {
                this.password = password;
                this.username = username;
                // Truncate username to 20 chars / strip illegal characters for the auth packet.
                String authUsername = Packet.formatAuthString(20, (byte) -5, username);

                if (this.password.trim().length() == 0) {
                    this.showLoginScreenStatus(STRINGS[474], STRINGS[471]); // "You must enter both a username" / "and a password - Please try again"
                    return;
                }

                if (reconnecting) {
                    // Silent reconnect: overlay a "lost connection" box rather than the
                    // normal connecting status.
                    this.drawTextBox(STRINGS[460], (byte) -64, STRINGS[446]); // "Connection lost! Please wait..." / "Attempting to re-establish"
                } else {
                    this.showLoginScreenStatus(STRINGS[436], STRINGS[432]); // "Please wait..." / "Connecting to server"
                }

                // World index <= 1 uses the primary login port, otherwise the alternate.
                int port = this.worldIndex <= 1 ? this.loginPort : this.loginPortAlt;
                this.clientStream = new ClientStream(this.createSocket(dummy, port, this.serverHost), this);
                this.clientStream.d = CacheFile.l; // max read-retry count

                // "limit30" applet param caps the frame rate to 30fps; flagged into the login block.
                int limit30 = 0;
                try {
                    if (InputState.a == null && this.getParameter(STRINGS[462]).equals("1")) {
                        limit30 = 1;
                    }
                } catch (Exception ignored) {
                }

                // Four random 32-bit words. These are both the ISAAC stream-cipher key
                // (applied to all subsequent traffic) and the XTEA key that encrypts the
                // login packet's plaintext tail (username/UID/password).
                int[] sessionKey = new int[]{
                    (int) (9.9999999E7 * Math.random()),
                    (int) (9.9999999E7 * Math.random()),
                    (int) (9.9999999E7 * Math.random()),
                    (int) (9.9999999E7 * Math.random()),
                };

                // ---- opcode 0: LOGIN ----  (dummy ^ 0xFFFFFFF4 reduces to opcode 0)
                // (clientStream.f is the ClientStream's outgoing BitBuffer.)
                this.clientStream.newPacket(0, dummy ^ 0xFFFFFFF4);
                // reconnect flag: 1 = re-establish existing session, 0 = fresh login.
                if (reconnecting) {
                    this.clientStream.f.putByte(1);
                } else {
                    this.clientStream.f.putByte(0);
                }
                this.clientStream.f.putInt(ClientIOException.CLIENT_VERSION); // client/protocol version

                // Build the RSA block in a scratch Buffer: leading byte 10, the four
                // session-key words, then the username, padding and a random tail byte.
                Buffer rsaBlock = new Buffer(500);
                rsaBlock.putByte(10);                  // RSA block marker
                rsaBlock.putInt(sessionKey[0]);
                rsaBlock.putInt(sessionKey[1]);
                rsaBlock.putInt(sessionKey[2]);
                rsaBlock.putInt(sessionKey[3]);
                rsaBlock.putString((byte) -39, authUsername);
                // Random padding ints (anti-replay filler before encryption).
                for (int i = 0; i < 5; i++) {
                    rsaBlock.putInt((int) (Math.random() * 9.9999999E7));
                }
                rsaBlock.putByte((int) (9.9999999E7 * Math.random()));
                // RSA-encrypt the whole block in place (modulus, exponent).
                rsaBlock.rsaEncrypt(BitBuffer.RSA_MODULUS, -118, FontBuilder.RSA_EXPONENT);

                // Emit the encrypted RSA block bytes, then a 2-byte length placeholder
                // for the XTEA tail that follows (patched after the tail is written).
                this.clientStream.f.putBytes(0, -123, rsaBlock.position, rsaBlock.buffer);
                this.clientStream.f.putShort(0); // placeholder for XTEA tail length
                int tailStart = this.clientStream.f.position;
                this.clientStream.f.putByte(limit30);
                RecordLoader.a(22607, this.clientStream.f); // append 24-byte client UID record
                this.clientStream.f.putString((byte) -39, this.password);
                // XTEA-encrypt the plaintext tail [tailStart, position) with sessionKey.
                this.clientStream.f.xteaEncrypt((byte) 87, tailStart, sessionKey, this.clientStream.f.position);
                // Back-patch the placeholder with the XTEA tail's actual length.
                this.clientStream.f.patchBlockLength(this.clientStream.f.position - tailStart, 1);
                this.clientStream.flushPacket(-6924);
                // Initialise the ISAAC stream cipher for all subsequent traffic.
                this.clientStream.seedIsaac((byte) -119, sessionKey);

                // ---- read one-byte login response ----
                int response = this.clientStream.readResponse(true);
                System.out.println(STRINGS[439] + response); // "login response:"

                // Bit 0x40 set => login succeeded; low bits carry account flags + rank.
                if ((response & 0x40) != 0) {
                    this.accountFlags = response & 3;            // low 2 bits = account flags
                    this.worldIndex = 0;
                    this.moderatorLevel = (response & 0x3F) >> 2; // bits 2..5 = staff rank
                    this.resetGameState(-109);
                    return;
                }
                if (response == 1) {
                    // Session needs verification (sleep word / recovery); enter that flow.
                    this.worldIndex = 0;
                    this.onSessionNeedsVerify(-16433);
                    return;
                }

                // Non-success: on the silent reconnect path, do not surface error text.
                if (!reconnecting) {
                    if (response == -1) {
                        this.showLoginScreenStatus(STRINGS[429], STRINGS[442]); // "Error unable to login." / "Server timed out"
                    } else if (response == 3) {
                        this.showLoginScreenStatus(STRINGS[431], STRINGS[473]); // "Invalid username or password." / "Try again, or create a new account"
                    } else if (response == 4) {
                        this.showLoginScreenStatus(STRINGS[450], STRINGS[453]); // "That username is already logged in." / "Wait 60 seconds then retry"
                    } else if (response == 5) {
                        this.showLoginScreenStatus(STRINGS[430], STRINGS[467]); // "The client has been updated." / "Please reload this page"
                    } else if (response == 6) {
                        this.showLoginScreenStatus(STRINGS[438], STRINGS[470]); // "You may only use 1 character at once." / "Your ip-address is already in use"
                    } else if (response == 7) {
                        this.showLoginScreenStatus(STRINGS[458], STRINGS[469]); // "Login attempts exceeded!" / "Please try again in 5 minutes"
                    } else if (response == 8) {
                        this.showLoginScreenStatus(STRINGS[429], STRINGS[445]); // "Error unable to login." / "Server rejected session"
                    } else if (response == 9) {
                        this.showLoginScreenStatus(STRINGS[425], STRINGS[453]); // "Error unable to login." / "Loginserver rejected session"
                    } else if (response == 10) {
                        this.showLoginScreenStatus(STRINGS[457], STRINGS[426]); // "That username is already in use." / "Wait 60 seconds then retry"
                    } else if (response == 11) {
                        this.showLoginScreenStatus(STRINGS[466], STRINGS[426]); // "Account temporarily disabled." / "Check your message inbox for details"
                    } else if (response == 12) {
                        this.showLoginScreenStatus(STRINGS[443], STRINGS[423]); // "Account permanently disabled." / "Check your message inbox for details"
                    } else if (response == 14) {
                        this.showLoginScreenStatus(STRINGS[444], STRINGS[449]); // "Sorry! This world is currently full." / "Please try a different world"
                        this.worldFullTimeout = 1500;
                    } else if (response == 15) {
                        this.showLoginScreenStatus(STRINGS[459], STRINGS[455]); // "You need a members account" / "to login to this world"
                    } else if (response == 16) {
                        this.showLoginScreenStatus(STRINGS[440], STRINGS[468]); // "Error - no reply from loginserver." / "Please try again"
                    } else if (response == 17) {
                        this.showLoginScreenStatus(STRINGS[463], STRINGS[435]); // "Error - failed to decode profile." / "Contact customer support"
                    } else if (response == 18) {
                        this.showLoginScreenStatus(STRINGS[451], STRINGS[428]); // "Account suspected stolen." / "Press 'recover a locked account' on front page."
                    } else if (response == 20) {
                        this.showLoginScreenStatus(STRINGS[464], STRINGS[449]); // "Error - loginserver mismatch" / "Please try a different world"
                    } else if (response == 21) {
                        this.showLoginScreenStatus(STRINGS[461], STRINGS[456]); // "Unable to login." / "That is not an RS-Classic account"
                    } else if (response == 22) {
                        this.showLoginScreenStatus(STRINGS[424], STRINGS[465]); // "Password suspected stolen." / "Press 'change your password' on front page."
                    } else if (response == 23) {
                        this.showLoginScreenStatus(STRINGS[434], STRINGS[435]); // (extended code) account-support message
                    } else if (response == 24) {
                        this.showLoginScreenStatus(STRINGS[472], STRINGS[448]); // (extended code) account-support message
                    } else {
                        this.showLoginScreenStatus(STRINGS[429], STRINGS[452]); // "Error unable to login." / "Unrecognised response code"
                    }
                    return;
                }

                // Reconnect path with a non-success response: silently give up this attempt.
                authUsername = "";
                this.password = "";
                this.resetLoginState(-2);
                return;
            } catch (Exception e) {
                System.out.println("" + e);
                // On exception, decrement the retry counter and (if any left) loop to retry.
                if (this.worldIndex > 0) {
                    try {
                        Utility.sleep(11200, 5000L);
                    } catch (Exception ignored) {
                    }
                    --this.worldIndex;
                    continue;
                }
                if (reconnecting) {
                    // Reconnect ran out of retries: forget credentials, drop to login screen.
                    this.username = "";
                    this.password = "";
                    this.resetLoginState(-2);
                } else {
                    Utility.reportError(0x1FFFFF, e, STRINGS[427]);
                    this.showLoginScreenStatus(STRINGS[441], STRINGS[433]); // "Sorry! Unable to connect." / "Check internet settings or try another world"
                }
                continue;
            }
        }
        // Loop exited (worldIndex <= 0). For the normal call path (dummy == -12) we are
        // already done; otherwise flush any queued client actions before returning.
        if (dummy == -12) {
            return;
        }
        this.sendQueuedActions((byte) -97);
    }

    /**
     * Append a server/system message to the chat history and route it to the
     * correct message tab.
     *
     * NOTE: the skeleton labelled this `registerAccount`, but the actual bytecode
     * (identical to OpenRSC's {@code showMessage} / mudclient's {@code fdj}) is the
     * chat-message display routine, not account registration. It pushes the message
     * onto the 100-entry rolling history buffers and adds it to the chat/quest/
     * private message panels. There is no opcode-2 REGISTER traffic here. Named to
     * match true behaviour; original skeleton name noted for traceability.
     *
     * @param crownEnabled   show the sender's rank crown (else crownId forced to 0)
     * @param sender         display name of the sender (may be null for system text)
     * @param messageSlot    index into the message-text history array for the insert
     *                       (callers pass 0 = newest slot)
     * @param message        the message body
     * @param type           internal MessageType id (0 game, 1 private-recv,
     *                       2 private-send, 3 quest, 4 chat, 6 friend-status,
     *                       7 inventory; NOT the OpenRSC enum order)
     * @param crownId        rank/crown id of the sender
     * @param formerName     clan/former display name; used for ignore filtering
     * @param colourOverride explicit colour code, or null to use the type's default
     *
     * obf: void a(boolean,String,int,String,int,int,String,String)  [skeleton: registerAccount; actual: showServerMessage]
     */
    private final void showServerMessage(boolean crownEnabled, String sender, int messageSlot, String message,
                                         int type, int crownId, String formerName, String colourOverride) {
        // This client's internal MessageType ids (do NOT match OpenRSC's enum order):
        //   0 = GAME, 1 = PRIVATE_RECV, 2 = PRIVATE_SEND, 3 = QUEST, 4 = CHAT,
        //   6 = FRIEND_STATUS, 7 = INVENTORY (others unused here).
        //
        // Ignore filtering: for a player-originated message (chat/private-recv/friend)
        // that is NOT showing a crown, drop it if the sender's display key is on the
        // ignore list. The display-key derived here is scratch used only for that
        // comparison; the actual render colour is always the per-type default below.
        if ((type == 1 || type == 4 || type == 6) && formerName != null && !crownEnabled) {
            String senderKey = WorldEntity.displayNameToKey(formerName, (byte) 93);
            if (senderKey == null) {
                return;
            }
            for (int i = 0; i < LinkedQueue.ignoreListCount; i++) {
                if (senderKey.equals(WorldEntity.displayNameToKey(SpriteScaler.ignoreList[i], (byte) 78))) {
                    return;
                }
            }
        }

        // Render colour for this message type (a "@xxx@" colour-code string).
        String colour = StreamFactory.messageTypeColors[type];

        // Flash the destination message tab (activity timer = 200) when the message
        // is NOT landing on the currently-viewed tab. `messageTabSelected` (Zh):
        // 0 = All/Game, 1 = Chat, 2 = Quest, 3 = Private.
        if (this.messageTabSelected != 0) {
            if ((type == 5 || type == 1 || type == 2) && this.messageTabSelected != 3) {
                this.tabActivityPrivate = 200;  // private tab
            }
            if (type == 4 && this.messageTabSelected != 1) {
                this.tabActivityChat = 200;     // chat tab
            }
            if (type == 3 && this.messageTabSelected != 2) {
                this.tabActivityQuest = 200;    // quest tab
            }
            if (type == 0 || type == 7) {
                this.tabActivityGame = 200;     // game/inventory -> "All/Game" tab
            }
            if (type == 1 && this.messageTabSelected != 0) {
                this.messageTabSelected = 0;
            }
            if ((type == 5 || type == 1 || type == 2) && this.messageTabSelected != 3 && this.messageTabSelected != 0) {
                this.messageTabSelected = 0;
            }
        }

        // An explicit colour override replaces the per-type default.
        if (colourOverride != null) {
            colour = colourOverride;
        }

        // Shift the 100-entry rolling message history down by one and insert at slot 0.
        // NB: the obfuscator scatters this one logical record's parallel arrays across
        // unrelated classes as static storage (FontWidths/ImageLoader/BitBuffer/World/
        // SurfaceSprite/BZip/NameTable) — those host classes are just opaque slots here,
        // not their NAMING.md semantic roles.
        for (int i = 99; i > 0; i--) {
            FontWidths.messageHistoryType[i] = FontWidths.messageHistoryType[i - 1];
            ImageLoader.messageHistoryTimeout[i] = ImageLoader.messageHistoryTimeout[i - 1];
            BitBuffer.messageHistoryCrownId[i] = BitBuffer.messageHistoryCrownId[i - 1];
            World.messageHistorySender[i] = World.messageHistorySender[i - 1];
            SurfaceSprite.messageHistoryClan[i] = SurfaceSprite.messageHistoryClan[i - 1];
            BZip.messageHistoryMessage[i] = BZip.messageHistoryMessage[i - 1];
            NameTable.messageHistoryColor[i] = NameTable.messageHistoryColor[i - 1];
        }
        FontWidths.messageHistoryType[0] = type;
        ImageLoader.messageHistoryTimeout[0] = 300; // frames the message stays in the in-world overlay
        World.messageHistorySender[0] = sender;
        BitBuffer.messageHistoryCrownId[0] = crownId;
        SurfaceSprite.messageHistoryClan[0] = formerName;
        BZip.messageHistoryMessage[messageSlot] = message;
        NameTable.messageHistoryColor[0] = colour;

        // Build the colour-prefixed, fully formatted message string.
        String formatted = colour + Utility.formatMessage(message, sender, true, type);

        // Route into the chat tab list. type 4 (CHAT) auto-scrolls only if already at
        // the bottom (controlScroll == listSize - 4); every other type is appended.
        if (type == 4) {
            boolean chatAtBottom =
                this.messagePanel.controlScrollAmount[this.tabChat] == this.messagePanel.controlListSize[this.tabChat] - 4;
            this.messagePanel.addToList(formatted, chatAtBottom, crownId, sender, formerName, (byte) -100, this.tabChat);
        } else {
            this.messagePanel.addToList(formatted, true, crownId, sender, formerName, (byte) -69, this.tabChat);
        }

        // QUEST (type 3) also goes to the quest tab.
        if (type == 3) {
            boolean questAtBottom =
                this.messagePanel.controlScrollAmount[this.tabQuest] == this.messagePanel.controlListSize[this.tabQuest] - 4;
            this.messagePanel.addToList(formatted, questAtBottom, 0, null, null, (byte) -64, this.tabQuest);
        }

        // PRIVATE messages (type 1 = received, 2 = sent) also go to the private tab.
        // The received-crown is only shown for received messages.
        if (type == 1 || type == 2) {
            int privCrown = crownId;
            if (type != 1) {
                privCrown = 0;
            }
            boolean privAtBottom =
                this.messagePanel.controlScrollAmount[this.tabPrivate] == this.messagePanel.controlListSize[this.tabPrivate] - 4;
            this.messagePanel.addToList(formatted, privAtBottom, privCrown, sender, formerName, (byte) -87, this.tabPrivate);
        }
    }

    /**
     * Build the login screen UI: the welcome panel (with account-type gating text
     * and a "Click here to login" button) and the username/password entry panel.
     *
     * Two {@code qa} (Panel) widgets are constructed each time the login screen is
     * shown:
     *   - {@code loginWelcomePanel} (obf {@code ge}) — title, gating text, login button.
     *   - {@code loginEntryPanel}   (obf {@code yi}) — username + password fields,
     *     Ok / Cancel buttons.
     *
     * obf: void p(int)  [proposed: drawLoginScreen]
     */
    private final void drawLoginScreen(int n) {
        // --- welcome panel ---
        this.loginWelcomePanel = new Panel(this.surface, 50);
        int y = 40;
        // Centered title at (256, 240).
        this.loginWelcomePanel.drawText(true, (byte) -79, 4, 256, STRINGS[237], 200 + y); // "Welcome to RuneScape"

        // Account-type gating sub-line. Selection depends on two flags:
        //   membersWorld (Pg) — this is a members world.
        //   veteranWorld (cf) — this is a veteran/classic world.
        // members && veteran -> STRINGS[233]; members && !veteran -> STRINGS[230];
        // !members && veteran -> STRINGS[238]; !members && !veteran -> no sub-line.
        String gatingText;
        if (this.membersWorld) {
            gatingText = this.veteranWorld ? STRINGS[233] : STRINGS[230];
        } else {
            gatingText = this.veteranWorld ? STRINGS[238] : null;
        }
        if (gatingText != null) {
            this.loginWelcomePanel.drawText(true, (byte) -109, 4, 256, gatingText, 215 + y);
        }

        // "Click here to login" button at (256, 250).
        this.loginWelcomePanel.drawButtonBackground(n - 3917, 200, 35, 256, y + 250);
        this.loginWelcomePanel.drawText(false, (byte) -96, 5, 256, STRINGS[232], y + 250); // "Click here to login"
        this.loginButton = this.loginWelcomePanel.addButton(256, 200, 250 + y, 91, 35);

        // --- username / password entry panel ---
        this.loginEntryPanel = new Panel(this.surface, 50);
        y = 230;
        this.loginTitleControl = this.loginEntryPanel.drawText(true, (byte) -107, 4, 256, "", y - 30);
        // Instruction line: "Please enter your username and password".
        this.loginPromptControl = this.loginEntryPanel.drawText(true, (byte) -125, 4, 256, STRINGS[65], y - 10);

        // First entry row. NOTE: field identity here is fixed by where drawLoginInput()
        // reads it back (obf ng -> password, Ih -> username); the displayed label index
        // and the masked flag in this obfuscated build do not follow the usual ordering.
        // The first input (obf ng) is the one read into the password.
        this.loginEntryPanel.drawButtonBackground(-87, 200, 40, 140, y += 28);
        this.loginEntryPanel.drawText(false, (byte) -126, 4, 140, STRINGS[235], y - 10);
        // addTextInput(..., masked = false, ...)
        this.passwordField = this.loginEntryPanel.addTextInput(n - 3845, 320, 200, false, 10 + y, 4, 40, false, 140);

        // Second entry row (obf Ih) -> read into the username.
        this.loginEntryPanel.drawButtonBackground(-120, 200, 40, 190, y += 47);
        this.loginEntryPanel.drawText(false, (byte) -93, 4, 190, STRINGS[234], y - 10);
        // addTextInput(..., masked = true, ...)
        this.usernameField = this.loginEntryPanel.addTextInput(n - 3845, 20, 200, false, 10 + y, 4, 40, true, 190);

        // Ok button (back at the higher row).
        this.loginEntryPanel.drawButtonBackground(-90, 120, 25, 410, y -= 55);
        this.loginEntryPanel.drawText(false, (byte) -127, 4, 410, STRINGS[231], y); // "Ok"
        this.loginOkButton = this.loginEntryPanel.addButton(410, 120, y, -94, 25);

        // Cancel button.
        this.loginEntryPanel.drawButtonBackground(n - 3952, 120, 25, 410, y += 30);
        this.loginEntryPanel.drawText(false, (byte) -89, 4, 410, STRINGS[121], y); // "Cancel"
        this.loginCancelButton = this.loginEntryPanel.addButton(410, 120, y, -120, 25);

        // Give the first entry field (obf ng -> passwordField) initial keyboard focus.
        this.loginEntryPanel.setFocus(this.passwordField, -105);
        y += 30;
    }

    /**
     * Per-frame login-screen input handler: drives the welcome screen and the
     * username/password entry sub-state, and fires {@link #loginUser} when the user
     * submits credentials.
     *
     * {@code loginScreenMode} (obf {@code Xd}) selects the sub-screen:
     *   0 = welcome screen (click to begin), 2 = username/password entry.
     *
     * obf: void x(int)  [proposed: drawLoginInput]
     */
    private final void drawLoginInput(int n) {
        // Mark login screen as needing a redraw (skip on the no-op param value).
        if (n != 2) {
            this.loginScreenRedraw = true;
        }
        // Count down the "world full" cooldown.
        if (this.worldFullTimeout > 0) {
            --this.worldFullTimeout;
        }

        if (this.loginScreenMode != 0) {
            // --- username/password entry sub-screen ---
            if (this.loginScreenMode != 2) {
                return;
            }
            // Forward mouse state to the entry panel.
            this.loginEntryPanel.handleMouse(this.lastMouseButtonDown, this.mouseX, -9989, this.mouseButtonClick, this.mouseY);

            // Cancel button -> back to the welcome screen.
            if (this.loginEntryPanel.isClicked((byte) -104, this.loginCancelButton)) {
                this.loginScreenMode = 0;
            }
            // Enter pressed in the first field -> advance focus to the second field.
            if (this.loginEntryPanel.isClicked((byte) -100, this.passwordField)) {
                this.loginEntryPanel.setFocus(this.usernameField, -88);
            }
            // Submit when Enter is pressed in the second field, or the Ok button clicked.
            if (!this.loginEntryPanel.isClicked((byte) -114, this.usernameField)
                && !this.loginEntryPanel.isClicked((byte) -105, this.loginOkButton)) {
                return;
            }
            // Field identity comes from the bytecode's read targets (Xf=username, wh=password).
            this.password = this.loginEntryPanel.getText(this.passwordField, n + 2);
            this.username = this.loginEntryPanel.getText(this.usernameField, 4);
            this.worldIndex = 2; // try alternate-port world by default on manual login
            this.loginUser(-12, this.username, this.password, false);
            return;
        }

        // --- welcome sub-screen: wait for the "Click here to login" button ---
        this.loginWelcomePanel.handleMouse(this.lastMouseButtonDown, this.mouseX, -9989, this.mouseButtonClick, this.mouseY);
        if (!this.loginWelcomePanel.isClicked((byte) -98, this.loginButton)) {
            return;
        }
        // Enter the username/password entry sub-screen and clear all login fields.
        this.loginScreenMode = 2;
        this.loginEntryPanel.setText(this.loginTitleControl, "", n ^ 0x6BF8);
        this.loginEntryPanel.setText(this.loginPromptControl, STRINGS[65], n + 27640); // "Please enter your username and password"
        this.loginEntryPanel.setText(this.passwordField, "", n ^ 0x6BF8);
        this.loginEntryPanel.setText(this.usernameField, "", 27642);
        this.loginEntryPanel.setFocus(this.passwordField, n ^ 0xFFFFFFA4);
    }

    /**
     * Handle a dropped/closed server connection (the SV_CLOSE_CONNECTION /
     * lost-socket path). Cancels the system-update countdown, then either resets to
     * the login screen (if mid-session) or starts a silent auto-relogin.
     *
     * If we are still logged in ({@code loggedInState} != 0) just reset the login
     * bookkeeping; otherwise log "Lost connection", arm 10 auto-relogin attempts and
     * call {@link #loginUser} in reconnect mode.
     *
     * @param dummy anti-tamper constant (callers pass 116/123, both > 59)
     *
     * obf: void u(int)  [proposed: closeConnection]
     */
    private final void closeConnection(int dummy) {
        this.systemUpdateTimer = 0; // cancel any pending "system update" countdown
        if (dummy <= 59) {
            // Dead for the real call paths (dummy is always > 59); kept for fidelity.
            this.sendDialogAnswer(-85);
        }
        if (this.loggedInState != 0) {
            // Mid-session: reset login state, do not auto-reconnect.
            this.resetLoginState(-2);
        } else {
            System.out.println(STRINGS[76]); // "Lost connection"
            this.worldIndex = 10;
            this.loginUser(-12, this.username, this.password, true);
        }
    }
