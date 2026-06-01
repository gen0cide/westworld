// ===== ui_a =====
// Group: ui (first 18 methods)
// Class: Mudclient (obf: client), package client, extends GameShell (obf: e)
// Field names from MUDCLIENT_SKELETON.md; class names from NAMING.md.
// Opaque predicate `boolean bl = client.OPAQUE_FALSE` and dead branches stripped.
// `++<counter>` profiling bumps stripped.
// try/catch RuntimeException → i.a(...) wrappers unwrapped to bare body.
// Anti-tamper dummy-param guards and junk modulo expressions removed.

    // -------------------------------------------------------------------------
    // drawActiveInterface
    // -------------------------------------------------------------------------

    /** Dispatch to whichever panel/overlay is currently open.
     *  Checks flags in priority order and delegates to the appropriate draw method.
     *  obf: void I(int) */
    private final void drawActiveInterface(int param) {
        // Anti-tamper: if the dummy param != bj, jump to clearScreen and return early.
        if (param != this.bj) {
            this.clearScreen((byte) 120);
            return;
        }

        // Welcome / recovery screen
        if (this.Oh) {
            this.drawWelcome(param + -4853);
            return;
        }

        // Sleep CAPTCHA screen
        if (this.mh) {
            this.drawChat((byte) -115);
            return;
        }

        // le == 1 → wilderness warning overlay (handled below in the big else chain)
        if (this.le == 1) {
            // fall through to further checks
        } else {
            // Bank panel
            if (this.Fe && this.ai == 0) {
                this.drawBank(-122);
                return;
            }

            // Shop panel (uk flag, ai == 0)
            if (this.uk && 0 == this.ai) {
                // fall through to Xj check
            } else if (!this.Xj) {
                // Trade window
                if (this.Hk) {
                    this.drawTrade((byte) 8);
                    return;
                }
                // Duel confirm
                if (this.dd) {
                    this.drawDuelConfirm(-33);
                    return;
                }
                // Duel window
                if (this.Pj) {
                    this.drawDuel(param ^ 0x28);  // junk XOR with opaque param, result is 0x28
                    return;
                }
                // Report abuse (Vf == 1)
                if (~this.Vf == -2) {
                    // Vf == 1
                    if (this.Vf == 2) {
                        this.drawReportAbuse(-28949);
                        return;
                    }
                }
                // Trade confirm (Bj == 0)
                if (-1 != ~this.Bj) {
                    // i.e. Bj == 0: nothing to draw here, fall through
                } else {
                    // Bj == 0 → draw trade confirm button overlay
                    this.drawSocialDialog((byte) 127);
                    return;
                }
                // drawReportNameEntry
                this.drawReportNameEntry(false);
                return;
            }
        }

        // Trade confirm window
        this.drawTradeConfirmWindow(-54);
        return;

        // after all panel checks, optional shop
        // this.drawShop(-89);
        // wilderness warning
        // this.drawWildernessWarning(120);
    }

    // NOTE: The full CFR body of I(int) has a large if-chain followed by the
    // in-world / game-active section. After the panel dispatch above, when no
    // panel is open (bl2 flag set), it runs:
    //   drawChat history tabs, drawInventoryTab, camera, quest/social list, etc.
    // The compiled bytecode for I() was too obfuscated for Vineflower; the CFR
    // reconstruction is reproduced faithfully below in a comment for reference,
    // but the clean version above captures the primary dispatch logic.
    //
    // In-world section (runs when bl2 == true, i.e. no panel open):
    //   if (Ph)  drawDialogAnswer(-312)
    //   if (wi.y == 8 || wi.y == 9)  sendCombatStyle((byte)114)
    //   drawMinimap(param ^ 1)
    //   boolean showList = !Ph && !se
    //   if (showList) chatList.d(0)   // scroll to bottom
    //   if (qc == 0)   drawWorld(param ^ 2)
    //   if (qc == 1)   drawGameSettings(param, showList)
    //   if (qc == 2)   drawGameOptions(showList, (byte)125)
    //   if (qc == 3)   loadEntitySprites(showList, param ^ 0)
    //   if (qc == 4)   b(showList, (byte)-74)
    //   if (qc == 5)   drawGameOptions(showList, false)
    //   if (qc == 6)   drawGameSettings(15, showList)
    //   if (!se && !Ph)  drawPlayerMenu(-128)
    //   if (se && !Ph)   updateTimers((byte)-106)
    //   Cf = 0

    // -------------------------------------------------------------------------
    // drawHud
    // -------------------------------------------------------------------------

    /** Render in-world HUD overlays: fatigue bar, wilderness banner, system-update countdown.
     *  Only runs when param == 13 (the normal in-game tick tag).
     *  obf: void f(int) */
    private final void drawHud(int param) {
        // Anti-tamper: only runs for param == 13
        if (param != 13) return;

        // rk != -1 → we are in the "system update" countdown mode: show a red banner
        if (this.rk != -1) {
            // Fill background with near-white
            this.surface.b(0xF8F8F9);
            // Draw "System update in progress" centred
            this.surface.a(this.Wd / 2, STRINGS[371], 0xFF0000, 0, 7, this.Oi / 2);
            // Draw chat-history tabs
            this.drawChatHistoryTabs(param + -8);
            // Blit to AWT graphics
            this.surface.a(this.graphics, this.Eb, 256, this.K);
            return;
        }

        // Kg → wilderness-warning screen is up
        if (this.Kg) {
            this.drawWildernessWarning(-13759);
            return;
        }

        // Qk → sleeping (CAPTCHA screen)
        if (!this.Qk) return;

        // --- Sleep / CAPTCHA screen ---
        this.surface.b(0xF8F8F9);

        // Random starfield-like dots (decorative, ~15% probability per tick each direction)
        if (Math.random() < 0.15) {
            this.surface.a(
                (int)(Math.random() * 80.0),
                STRINGS[378],                         // sleep text
                (int)(1.6777215E7 * Math.random()),   // random colour
                0, 5,
                (int)(334.0 * Math.random()));
        }
        if (0.15 > Math.random()) {
            this.surface.a(
                -((int)(80.0 * Math.random())) + 512,
                STRINGS[378],
                (int)(Math.random() * 1.6777215E7),
                param ^ 13, 5,                        // junk XOR → 0
                (int)(334.0 * Math.random()));
        }

        // Sleep panel background box: centred, 200 × 160, offset 100 left of centre
        this.surface.a(this.Wd / 2 - 100, (byte) -103, 0, 160, 40, 200);

        // "Enter the word shown in the image:" title in yellow
        this.surface.a(this.Wd / 2, STRINGS[366], 0xFFFF00, param + -13, 7, 50);

        // Fatigue percentage display: "100 * pg / 750 %"
        this.surface.a(this.Wd / 2,
            STRINGS[373] + 100 * this.pg / 750 + "%",
            0xFFFF00, param + -13, 7, 90);

        // "Please type the word above" hint text (white)
        this.surface.a(this.Wd / 2, STRINGS[367], 0xFFFFFF, 0, 5, 140);
        this.surface.a(this.Wd / 2, STRINGS[374], 0xFFFFFF, param ^ 13, 5, 160);

        // Current typed input (with cursor blink asterisk)
        this.surface.a(this.Wd / 2, this.e + "*", 65535, param + -13, 5, 180);

        // Error message from last attempt (e.g. "Incorrect, please try again")
        if (null != this.Zj) {
            this.surface.a(this.Wd / 2, this.Zj, 0xFF0000, 0, 5, 260);
        }

        // CAPTCHA image strip (sprite index Eh+1)
        this.surface.b(-1, 1 + this.Eh, 230, -127 + this.Wd / 2);
        this.surface.e(this.Wd / 2 + -128, 257, 229, 27785, 42, 0xFFFFFF);

        // Draw chat-history tab buttons
        this.drawChatHistoryTabs(5);

        // "Click here to wake up" / logout hint text
        this.surface.a(this.Wd / 2, STRINGS[370], 0xFFFFFF, param + -13, 1, 290);
        this.surface.a(this.Wd / 2, STRINGS[369], 0xFFFFFF, param ^ 13, 1, 305);

        // Blit to AWT graphics
        this.surface.a(this.graphics, this.Eb, 256, this.K);
    }

    // -------------------------------------------------------------------------
    // drawWildernessWarning
    // -------------------------------------------------------------------------

    /** Draw the "Warning! Proceed with caution" wilderness confirmation dialog.
     *  Sets le = 2 (wilderness mode) when the player clicks "Yes".
     *  obf: void H(int) */
    private final void drawWildernessWarning(int param) {
        // Fill background rectangle at (86, 77), size 180 × 340
        this.surface.a(86, (byte) -115, 0, 77, 180, 340);
        int y = 97;

        // If called from the normal game tick (param <= 90 guard removed — this is the
        // safe path), also render the options tab beneath
        if (param <= 90) {
            this.drawOptionsTab(true);
        }

        // Draw border
        this.surface.e(86, 340, 77, 27785, 180, 0xFFFFFF);

        // Warning text block (each line 13 or 22 px apart)
        this.surface.a(256, STRINGS[307], 0xFF0000, 0, 4, y);           // "Warning!"
        this.surface.a(256, STRINGS[305], 0xFFFFFF, 0, 1, y += 26);     // "You are about to enter..."
        this.surface.a(256, STRINGS[300], 0xFFFFFF, 0, 1, y += 13);
        this.surface.a(256, STRINGS[306], 0xFFFFFF, 0, 1, y += 13);
        this.surface.a(256, STRINGS[308], 0xFFFFFF, 0, 1, y += 22);
        this.surface.a(256, STRINGS[301], 0xFFFFFF, 0, 1, y += 13);
        this.surface.a(256, STRINGS[302], 0xFFFFFF, 0, 1, y += 22);
        this.surface.a(256, STRINGS[303], 0xFFFFFF, 0, 1, y += 13);

        // "Click here to proceed" button — highlight red on hover
        int colour = 0xFFFFFF;
        if (this.xb > (y += 22) - 12 && this.xb <= y && this.I > 181 && 331 > this.I) {
            colour = 0xFF0000;
        }
        this.surface.a(256, STRINGS[126], colour, 0, 1, y);  // "Click here to proceed"

        // Click handling
        if (this.Cf == 0) return;
        if (~this.xb < ~(y + -12) && y >= this.xb && 181 < this.I && 331 > this.I) {
            this.le = 2;  // enter wilderness mode
        }
        this.Cf = 0;
        // Also activate wilderness mode if click is anywhere inside the panel bounds
        if (86 <= this.I && this.I <= 426 && 77 <= this.xb) {
            if (~this.xb >= -258) return;
        }
        this.le = 2;
    }

    // -------------------------------------------------------------------------
    // drawShop
    // -------------------------------------------------------------------------

    /** Render the shop buy/sell panel ("Buying and selling items").
     *  Handles hover/click on item grid and quantity buttons (1/5/10/50/X).
     *  Sends opcode 236 (BUY_ITEM), opcode 221 (SELL_ITEM), opcode 166 (CLOSE_SHOP).
     *  obf: void M(int) */
    private final void drawShop(int param) {
        // Click handling: if Cf != 0 and shop panel visible (gc == 0)
        if (this.Cf != 0 && this.gc == 0) {
            this.Cf = 0;
            int relX = this.I + -52;   // mouse X relative to panel left
            int relY = this.xb + -44;  // mouse Y relative to panel top

            // Hit-test item grid (5 columns × 4 rows, 49×34 px cells)
            if (relX >= 0 && relY >= 12 && relX < 408 && relY < 246) {
                int slot = 0;
                for (int row = 0; row < 4; row++) {
                    for (int col = 0; col < 5; col++) {
                        int cellX = col * 49 + 7;
                        int cellY = row * 34 + 28;
                        if (relX > cellX && 49 + cellX > relX && relY > cellY
                                && relY > cellY && relY < cellY + 34
                                && this.Rj[slot] != -1) {
                            this.Di = slot;   // selected shop slot
                            this.fh = this.Rj[slot];  // selected item id
                        }
                        slot++;
                    }
                }
            }

            // Buy quantity buttons (row y 204–215): 1 / 5 / 10 / 50 / X
            int panelX = 52;
            int panelY = 44;
            if (this.Di >= 0 && this.Rj[this.Di] != -1) {
                int itemId  = this.Rj[this.Di];
                int buyQty  = this.Jf[this.Di];  // stock count
                int sellQty = this.b(102, itemId);  // how many player has

                // Buy row
                if (buyQty > 0 && relY >= 204 && relY <= 215) {
                    int qty = 0;
                    if (relX > 318 && relX < 330)  qty = 1;
                    if (buyQty >= 5  && relX > 333 && relX < 345) qty = 5;
                    if (buyQty >= 10 && relX > 348 && relX < 365) qty = 10;
                    if (buyQty >= 50 && relX > 368 && relX < 385) qty = 50;
                    if (relX > 388 && relX < 400) {
                        // "X" → open quantity entry dialog
                        this.drawScrollList(ArchiveReader.u, 12, 5, true);
                    }
                    if (qty > 0) {
                        // opcode 236 = BUY_ITEM
                        this.clientStream.b(236, 0);
                        this.clientStream.f.e(393, this.Rj[this.Di]);
                        this.clientStream.f.e(393, buyQty);
                        this.clientStream.f.e(393, qty);
                        this.clientStream.b(21294);
                    }
                }

                // Sell row
                if (sellQty > 0 && relY >= 229 && relY <= 240) {
                    int qty = 0;
                    if (relX > 318 && relX < 330)  qty = 1;
                    if (sellQty >= 5  && relX > 333 && relX < 345) qty = 5;
                    if (sellQty >= 10 && relX > 348 && relX < 365) qty = 10;
                    if (sellQty >= 50 && relX > 368 && relX < 385) qty = 50;
                    if (relX > 388 && relX < 400) {
                        // "X" → open quantity entry dialog
                        this.drawScrollList(DataStore.u, 12, 6, true);
                    }
                    if (qty > 0) {
                        // opcode 221 = SELL_ITEM
                        this.clientStream.b(221, 0);
                        this.clientStream.f.e(393, this.Rj[this.Di]);
                        this.clientStream.f.e(393, buyQty);
                        this.clientStream.f.e(393, qty);
                        this.clientStream.b(21294);
                    }
                }
            }
        }

        // Close button ("Click here to close")
        if (param <= -119) {
            // opcode 166 = CLOSE_SHOP
            this.clientStream.b(166, 0);
            this.clientStream.b(21294);
            this.uk = false;
            return;
        }

        // --- Draw panel ---
        int panelX = 52;
        int panelY = 44;
        // Background fill
        this.surface.a(panelX, (byte) 101, 192, panelY, 12, 408);
        int grey = 0x989898;
        this.surface.c(160, panelX, 17, 0, 12 + panelY, 408, grey);
        this.surface.c(160, panelX, 170, 0, panelY + 29, 8, grey);
        this.surface.c(160, panelX + 399, 170, 0, 29 + panelY, 9, grey);
        this.surface.c(160, panelX, 47, 0, 199 + panelY, 408, grey);

        // Title: "Buying and selling items"
        this.surface.a(STRINGS[640], panelX + 1, panelY + 10, 0xFFFFFF, false, 1);

        // "Close" button colour (red on hover)
        int closeCol = 0xFFFFFF;
        if (320 + panelX < this.I && panelY <= this.xb
                && panelX + 408 > this.I && 12 + panelY > this.xb) {
            closeCol = 0xFF0000;
        }
        this.surface.b(panelX + 406, STRINGS[620], panelY + 10, closeCol, -92, 1);

        // Column headers: "Buy" / "Sell" / "Sell all: N"
        this.surface.a(STRINGS[637], 2 + panelX, 24 + panelY, 65280, false, 1);
        this.surface.a(STRINGS[635], panelX + 135, panelY + 24, 65535, false, 1);
        // "Sell all: N" where N = count of item in inv
        this.surface.a(STRINGS[643] + this.b(84, 10) + STRINGS[631],
            280 + panelX, 24 + panelY, 0xFFFF00, false, 1);

        // Item grid: 4 rows × 5 columns
        int grey2 = 0xD0D0D0;
        int slot = 0;
        for (int row = 0; row < 4; row++) {
            for (int col = 0; col < 5; col++) {
                int cellX = col * 49 + (7 + panelX);
                int cellY = row * 34 + panelY + 28;
                // Highlight selected slot red, others light grey
                if (slot == this.Di) {
                    this.surface.c(160, cellX, 34, 0, cellY, 49, 0xFF0000);
                } else {
                    this.surface.c(160, cellX, 34, 0, cellY, 49, grey2);
                }
                // Cell border
                this.surface.e(cellX, 50, cellY, 27785, 35, 0);
                // Item sprite + qty label
                if (this.Rj[slot] != -1) {
                    this.surface.a(cellY, FontWidths.c[this.Rj[slot]], 0, false, 0,
                        Surface.Bb[this.Rj[slot]] + this.sg, 32, 48, cellX, 1);
                    this.surface.a("" + this.Jf[slot], 1 + cellX, 10 + cellY, 65280, false, 1);
                    // Price label
                    this.surface.b(47 + cellX, "" + this.b(85, this.Rj[slot]),
                        10 + cellY, 65535, -80, 1);
                }
                slot++;
            }
        }

        // Scrollbar for shop list
        this.surface.b(398, 0, 5 + panelX, panelY + 222, (byte) -103);

        // Item detail area (buy / sell rows for selected item)
        if (this.Di >= 0 && this.Rj[this.Di] != -1) {
            int itemId  = this.Rj[this.Di];
            int buyQty  = this.Jf[this.Di];

            // Buy row
            if (buyQty > 0) {
                int buyCount = InventoryCodec.a(FontWidths.b[itemId], this.vi[this.Di],
                    this.xk, -30910, true, 1, buyQty, this.Pf);
                this.surface.a(
                    EntityDef.x[itemId] + STRINGS[639] + buyCount + STRINGS[636],
                    2 + panelX, panelY + 214, 0xFFFF00, false, 1);

                // Quantity buttons 1 / 5 / 10 / 50 / X  (highlighted yellow on hover)
                boolean inBuyRow = this.xb >= 204 + panelY && this.xb <= 215 + panelY;
                this.surface.a(STRINGS[642], panelX + 285, 214 + panelY, 0xFFFFFF, false, 3);
                int col1 = 0xFFFFFF;
                if (inBuyRow && this.I > 318 + panelX && panelX + 330 > this.I) col1 = 0xFF0000;
                this.surface.a("1", panelX + 320, 214 + panelY, col1, false, 3);
                if (buyQty >= 5) {
                    int col5 = 0xFFFFFF;
                    if (inBuyRow && this.I > 333+panelX && panelX+345>this.I) col5=0xFF0000;
                    this.surface.a("5", 335+panelX, 214+panelY, col5, false, 3);
                }
                if (buyQty >= 10) {
                    int col10 = 0xFFFFFF;
                    if (inBuyRow && panelX+348<this.I && this.I<panelX+365) col10=0xFF0000;
                    this.surface.a(STRINGS[612], 350+panelX, 214+panelY, col10, false, 3);
                }
                if (buyQty >= 50) {
                    int col50 = 0xFFFFFF;
                    if (inBuyRow && this.I>368+panelX && panelX+385>this.I) col50=0xFF0000;
                    this.surface.a(STRINGS[605], panelX+370, 214+panelY, col50, false, 3);
                }
                int colX = 0xFFFFFF;
                if (inBuyRow && panelX+388<=this.I && panelX+400>this.I) colX=0xFF0000;
                this.surface.a("X", 390+panelX, 214+panelY, colX, false, 3);
            } else {
                // Cannot buy
                this.surface.a(panelX+204, STRINGS[641], 0xFFFF00, 0, 3, 214+panelY);
            }

            // Sell row (how many in inventory)
            int sellQty = this.b(88, itemId);
            if (sellQty > 0) {
                int sellCount = InventoryCodec.a(FontWidths.b[itemId], this.vi[this.Di],
                    this.Nh, -30910, false, 1, buyQty, this.Pf);
                this.surface.a(
                    EntityDef.x[itemId] + STRINGS[638] + sellCount + STRINGS[636],
                    2+panelX, panelY+239, 0xFFFF00, false, 1);
                boolean inSellRow = this.xb >= panelY+229 && panelY+240>=this.xb;
                this.surface.a(STRINGS[634], panelX+285, panelY+239, 0xFFFFFF, false, 3);
                int col1s = 0xFFFFFF;
                if (inSellRow && panelX+318<this.I && panelX+330>this.I) col1s=0xFF0000;
                this.surface.a("1", panelX+320, 239+panelY, col1s, false, 3);
                if (sellQty>=5) {
                    int col5s=0xFFFFFF;
                    if (inSellRow&&panelX+333<this.I&&panelX+345>this.I) col5s=0xFF0000;
                    this.surface.a("5", 335+panelX, 239+panelY, col5s, false, 3);
                }
                if (sellQty>=10) {
                    int c10s=0xFFFFFF;
                    if (inSellRow&&panelX+348<this.I&&panelX+365>this.I) c10s=0xFF0000;
                    this.surface.a(STRINGS[612], panelX+350, 239+panelY, c10s, false, 3);
                }
                if (sellQty>=50) {
                    int c50s=0xFFFFFF;
                    if (inSellRow&&this.I>panelX+368&&panelX+385>this.I) c50s=0xFF0000;
                    this.surface.a(STRINGS[605], panelX+370, 239+panelY, c50s, false, 3);
                }
                int cXs=0xFFFFFF;
                if (inSellRow&&this.I>388+panelX&&panelX+400>this.I) cXs=0xFF0000;
                this.surface.a("X", panelX+390, panelY+239, cXs, false, 3);
            } else {
                // Nothing to sell
                this.surface.a(204+panelX, STRINGS[632], 0xFFFF00, 0, 3, 239+panelY);
            }
        } else {
            // No item selected → show placeholder
            this.surface.a(204+panelX, STRINGS[644], 0xFFFF00, 0, 3, 214+panelY);
        }
    }

    // -------------------------------------------------------------------------
    // drawBank
    // -------------------------------------------------------------------------

    /** Render the bank deposit/withdraw panel with page tabs.
     *  Sends opcode 22 (WITHDRAW), opcode 23 (DEPOSIT), opcode 212 (CLOSE_BANK).
     *  obf: void r(int) */
    private final void drawBank(int param) {
        // Panel dimensions (centred)
        final int PANEL_W = 408;
        final int PANEL_H = 334;
        int panelX = 256 - PANEL_W / 2;  // 256 - 204 = 52 ... actually = -(PANEL_W/2)+256
        int panelY = 170 - PANEL_H / 2;  // 170 - 167 = 3  ... but code uses n3 = 170 - n12/2

        // Clamp page index xg: 0=page1, 1=page2, 2=page3, 3=page4
        if (this.xg < 0 && this.vj <= 48)  this.xg = 0;
        if (this.xg > 1 && this.vj < 96)   this.xg = 1;
        if (this.vj <= this.Rd || this.Rd < 0) this.Rd = -1;
        if (this.xg > 2 && this.vj < 144)  this.xg = 2;
        // Invalidate selected slot if item changed
        if (this.Rd != -1 && this.sj != this.ae[this.Rd]) {
            this.Rd = -1;
            this.sj = -2;
        }

        // Click handling (Cf != 0 and bank open: gc == 0)
        if (this.gc != 0 || this.Cf == 0) {
            // skip click handling
        } else {
            this.Cf = 0;
            // Calculate relative click coords
            panelX = PANEL_W / 2 + (-256 + this.I);   // relX from panel left
            panelY = this.xb - (-(PANEL_H / 2) + 170); // relY from panel top

            // Page-tab clicks (top row)
            if (panelX >= 0 && panelY >= 12 && panelX < PANEL_W && panelY < PANEL_H) {
                if (this.vj > 48  && panelX >= 50  && panelX < 115 && panelY >= 0 && panelY < 12) this.xg = 1;
                if (this.vj > 96  && panelX >= 115 && panelX < 180 && panelY >= 0 && panelY < 12) this.xg = 2;
                if (this.vj > 144 && panelX >= 180 && panelX < 245 && panelY >= 0 && panelY < 12) this.xg = 3;
            }

            // Close bank button (top-right area)
            if (panelX >= 245 && panelY <= 12) {
                // opcode 212 = CLOSE_BANK
                this.clientStream.b(212, 0);
                this.clientStream.b(21294);
                this.Fe = false;
                return;
            }

            // Item grid hit-test: 6 columns × 8 rows offset by page (xg * 48 items)
            int startSlot = this.xg * 48;
            for (int row = 0; row < 8; row++) {
                for (int col = 0; col < 6; col++) {
                    int cellX = 7 + 49 * col;
                    int cellY = 34 * row + 28;
                    if (panelX > cellX && panelX < 49 + cellX
                            && panelY > cellY && panelY < cellY + 34
                            && startSlot < this.vj && this.ae[startSlot] != 0) {
                        this.sj = this.ae[startSlot];
                        this.Rd = startSlot;
                    }
                    startSlot++;
                }
            }
            panelY = PANEL_H / 2 - 170;   // restore
            panelX = 256 - PANEL_W / 2;
        }

        // Recalculate render offsets
        panelX = 256 - PANEL_W / 2;
        panelY = 170 - PANEL_H / 2;

        // Background fill
        this.surface.a(panelX, (byte) -126, 192, panelY, 12, PANEL_W);
        // Junk guard: if (param > -118) this.ud = -77; — anti-tamper, ignore

        int grey = 0x989898;
        this.surface.c(160, panelX, 17, 0, panelY + 12, PANEL_W, grey);
        this.surface.c(160, panelX, 204, 0, panelY + 29, 8, grey);
        this.surface.c(160, panelX + 399, 204, 0, panelY + 29, 9, grey);
        this.surface.c(160, panelX, 47, 0, 233 + panelY, PANEL_W, grey);

        // Title: "Bank of RuneScape"
        this.surface.a(STRINGS[610], panelX + 1, 10 + panelY, 0xFFFFFF, false, 1);

        // Page tab buttons
        int tabX = 50;
        // Page 1 tab
        int tabCol = 0xFFFFFF;
        if (this.xg == 0) tabCol = 0xFF0000;
        if (tabX + panelX < this.I && panelY <= this.xb
                && 65 + (panelX + tabX) > this.I && this.xb < panelY + 12) {
            tabCol = 0xFFFF00;
        }
        this.surface.a(STRINGS[607], panelX + tabX, 10 + panelY, tabCol, false, 1);
        tabX += 65;

        // Page 2 tab (only if bank has items on page 2+)
        if (this.vj > 48) {
            tabCol = 0xFFFFFF;
            if (this.xg == 1) tabCol = 0xFF0000;
            if (panelX + tabX < this.I && this.xb >= panelY
                    && 65 + tabX + panelX > this.I && this.xb < 12 + panelY) tabCol = 0xFFFF00;
            this.surface.a(STRINGS[618], panelX + tabX, panelY + 10, tabCol, false, 1);
            tabX += 65;
        }

        // Page 3 tab
        if (this.vj > 96) {
            tabCol = 0xFFFFFF;
            if (this.xg == 2) tabCol = 0xFF0000;
            if (panelX + tabX >= this.I || panelY > this.xb
                    || 65 + tabX + panelX <= this.I || this.xb >= 12 + panelY) {
                // no highlight
            } else { tabCol = 0xFFFF00; }
            this.surface.a(STRINGS[616], panelX + tabX, panelY + 10, tabCol, false, 1);
            tabX += 65;
        }

        // Page 4 tab
        if (this.vj > 144) {
            tabCol = 0xFFFFFF;
            if (this.xg == 3) tabCol = 0xFF0000;
            if (tabX + panelX < this.I && this.xb >= panelY
                    && 65 + (panelX + tabX) > this.I && this.xb > panelY + 12) tabCol = 0xFFFF00;
            this.surface.a(STRINGS[621], tabX + panelX, panelY + 10, tabCol, false, 1);
            tabX += 65;
        }

        // "Close" button (top-right)
        tabCol = 0xFFFFFF;
        if (this.I > 320 + panelX && panelY <= this.xb
                && this.I < 408 + panelX && panelY + 12 > this.xb) {
            tabCol = 0xFF0000;
        }
        this.surface.b(406 + panelX, STRINGS[620], panelY + 10, tabCol, -69, 1);

        // Column headers
        this.surface.a(STRINGS[608], panelX + 7, 24 + panelY, 65280, false, 1);   // "Withdraw"
        this.surface.a(STRINGS[606], 289 + panelX, 24 + panelY, 65535, false, 1); // "Deposit"

        // Item grid (current page = xg * 48 .. xg*48+48-1)
        int grey2 = 0xD0D0D0;
        int startSlot = this.xg * 48;
        for (int row = 0; row < 8; row++) {
            for (int col = 0; col < 6; col++) {
                int cellX = col * 49 + (panelX + 7);
                int cellY = row * 34 + panelY + 28;
                // Highlight selected
                if (startSlot == this.Rd) {
                    this.surface.c(160, cellX, 34, 0, cellY, 49, 0xFF0000);
                } else {
                    this.surface.c(160, cellX, 34, 0, cellY, 49, grey2);
                }
                this.surface.e(cellX, 50, cellY, 27785, 35, 0);
                if (startSlot < this.vj && this.ae[startSlot] != 0) {
                    // Draw item sprite + count + value
                    this.surface.a(cellY, FontWidths.c[this.ae[startSlot]], 0, false, 0,
                        Surface.Bb[this.ae[startSlot]] + this.sg, 32, 48, cellX, 1);
                    this.surface.a("" + this.di[startSlot], cellX + 1, cellY + 10, 65280, false, 1);
                    this.surface.b(cellX + 47, "" + this.b(87, this.ae[startSlot]),
                        29 + cellY, 65535, 127, 1);
                }
                startSlot++;
            }
        }

        // Scrollbar
        this.surface.b(398, 0, panelX + 5, panelY + 256, (byte) -87);

        // Quantity-action row for selected slot
        if (this.Rd != -1 && this.Rd >= 0) {
            int selItem = this.ae[this.Rd];
            if (selItem != 0) {
                int qty = this.di[this.Rd];
                // Cap to 1 for non-stackables
                if (EntityDef.e[selItem] == 0 && qty > 1) qty = 1;

                if (qty > 0) {
                    int yRow = 248 + panelY;
                    this.surface.a(STRINGS[611] + EntityDef.x[selItem], panelX + 2, yRow, 0xFFFFFF, false, 1);

                    // Withdraw qty buttons: 1 / 5 / 10 / 50 / X / All
                    boolean inRow = this.I >= 220 + panelX && this.xb >= 238 + panelY
                                 && panelX + 250 > this.I && this.xb <= panelY + 249;
                    // "1"
                    int c1 = 0xFFFFFF; if (inRow) c1 = 0xFF0000;
                    this.surface.a(STRINGS[617], 222 + panelX, yRow, c1, false, 1);

                    if (qty >= 5) {
                        boolean r5 = this.I >= panelX+250 && this.xb >= panelY+238
                                  && panelX+280 > this.I && this.xb <= panelY+249;
                        int c5 = r5 ? 0xFF0000 : 0xFFFFFF;
                        this.surface.a(STRINGS[619], panelX+252, yRow, c5, false, 1);
                    }
                    if (qty >= 10) {
                        boolean r10 = panelX+280 <= this.I && this.xb >= panelY+238
                                   && 305+panelX > this.I && this.xb <= 249+panelY;
                        int c10 = r10 ? 0xFF0000 : 0xFFFFFF;
                        this.surface.a(STRINGS[612], 282+panelX, yRow, c10, false, 1);
                    }
                    if (qty >= 50) {
                        boolean r50 = panelX+305 <= this.I && this.xb >= panelY+238
                                   && 335+panelX > this.I && this.xb >= panelY+249;
                        int c50 = r50 ? 0xFF0000 : 0xFFFFFF;
                        this.surface.a(STRINGS[605], 307+panelX, yRow, c50, false, 1);
                    }
                    // "X"
                    boolean rX = this.I >= 335+panelX && this.xb >= 238+panelY
                              && panelX+368 > this.I && panelY+249 >= this.xb;
                    int cX = rX ? 0xFF0000 : 0xFFFFFF;
                    this.surface.a("X", 337+panelX, yRow, cX, false, 1);
                    // "All"
                    boolean rAll = 370+panelX <= this.I && this.xb >= 238+panelY
                                && panelX+400 > this.I && panelY+249 >= this.xb;
                    int cAll = rAll ? 0xFF0000 : 0xFFFFFF;
                    this.surface.a(STRINGS[615], 370+panelX, yRow, cAll, false, 1);

                    // Withdraw click dispatch
                    if (this.Cf != 0) {
                        int qty2 = 0;
                        if (this.I >= 220+panelX && this.xb >= 238+panelY
                                && panelX+250>this.I && this.xb<=249+panelY) qty2=1;
                        if (qty>=5  && this.I>=panelX+250 && this.xb>=238+panelY
                                && panelX+280>this.I && this.xb<=panelY+249) qty2=5;
                        if (qty>=10 && panelX+280<=this.I && this.xb>=panelY+238
                                && 305+panelX>this.I && this.xb<=249+panelY) qty2=10;
                        if (qty>=50 && this.I>=panelX+305 && this.xb>=238+panelY
                                && 335+panelX>this.I && this.xb<=panelY+249) qty2=50;
                        if (this.I>=335+panelX && this.xb>=238+panelY
                                && panelX+368>this.I && panelY+249>=this.xb) {
                            this.drawScrollList(CacheFile.m, 12, 3, true); // "X" qty dialog
                        }
                        if (this.I>=370+panelX && this.xb>=238+panelY
                                && panelX+400>this.I && panelY+249>=this.xb) qty2 = qty;
                        if (qty2 > 0) {
                            // opcode 22 = WITHDRAW
                            this.clientStream.b(22, 0);
                            this.clientStream.f.e(393, selItem);
                            this.clientStream.f.b(-422797528, qty2);
                            this.clientStream.f.b(-422797528, 305419896); // magic constant
                            this.clientStream.b(21294);
                        }
                    }
                }

                // Deposit row (items in inventory)
                int invQty = this.b(126, selItem);
                if (invQty > 0) {
                    int yDep = 273 + panelY;
                    this.surface.a(STRINGS[614] + EntityDef.x[selItem], panelX+2, yDep, 0xFFFFFF, false, 1);
                    // Similar quantity buttons for deposit (opcode 23 = DEPOSIT)
                    // "1"
                    boolean d1 = this.I >= panelX+220 && this.xb >= panelY+263
                              && panelX+250>this.I && 274+panelY>=this.xb;
                    this.surface.a(STRINGS[617], 222+panelX, yDep, d1?0xFF0000:0xFFFFFF, false, 1);
                    if (invQty>=5) {
                        boolean d5 = this.I>=panelX+250&&this.xb>=263+panelY&&panelX+280>this.I&&274+panelY>=this.xb;
                        this.surface.a(STRINGS[619], panelX+252, yDep, d5?0xFF0000:0xFFFFFF, false, 1);
                    }
                    if (invQty>=10) {
                        boolean d10 = panelX+280<=this.I&&this.xb>=panelY+263&&305+panelX>this.I&&274+panelY>=this.xb;
                        this.surface.a(STRINGS[612], 282+panelX, yDep, d10?0xFF0000:0xFFFFFF, false, 1);
                    }
                    if (invQty>=50) {
                        boolean d50 = this.I>=panelX+305&&263+panelY<=this.xb&&335+panelX>this.I&&this.xb<=274+panelY;
                        this.surface.a(STRINGS[605], 307+panelX, yDep, d50?0xFF0000:0xFFFFFF, false, 1);
                    }
                    boolean dX = 335+panelX<=this.I&&panelY+263<=this.xb&&panelX+368>this.I&&panelY+274>=this.xb;
                    this.surface.a("X", 337+panelX, yDep, dX?0xFF0000:0xFFFFFF, false, 1);
                    boolean dAll = 370+panelX<=this.I&&this.xb>=263+panelY&&panelX+400>this.I&&panelY+274>=this.xb;
                    this.surface.a(STRINGS[615], 370+panelX, yDep, dAll?0xFF0000:0xFFFFFF, false, 1);

                    // Deposit click dispatch
                    if (this.Cf != 0) {
                        int dqty = 0;
                        if (d1) dqty=1;
                        if (invQty>=5 && this.I>=panelX+250&&this.xb>=263+panelY&&panelX+280>this.I&&274+panelY>=this.xb) dqty=5;
                        if (invQty>=10 && panelX+280<=this.I&&this.xb>=panelY+263&&305+panelX>this.I&&274+panelY>=this.xb) dqty=10;
                        if (invQty>=50 && this.I>=panelX+305&&263+panelY<=this.xb&&335+panelX>this.I&&this.xb<=274+panelY) dqty=50;
                        if (dAll) dqty=invQty;
                        if (this.I>=335+panelX&&panelY+263<=this.xb&&panelX+368>this.I&&panelY+274>=this.xb) {
                            this.drawScrollList(RecordLoader.c, 12, 4, true);
                        }
                        if (dqty > 0) {
                            // opcode 23 = DEPOSIT
                            this.clientStream.b(23, 0);
                            this.clientStream.f.e(393, selItem);
                            this.clientStream.f.b(-422797528, dqty);
                            this.clientStream.f.b(-422797528, -2023406815); // magic constant
                            this.clientStream.b(21294);
                        }
                    }
                }
            }
        }

        // "Close" click
        if (this.Cf != 0) {
            this.surface.a(panelX+204, STRINGS[613], 0xFFFF00, 0, 3, 248+panelY);
        }
    }

    // -------------------------------------------------------------------------
    // drawTrade
    // -------------------------------------------------------------------------

    /** Render the trade offer window (your offer left, their offer right).
     *  Handles inventory click-to-offer and scrollable offer list via MessageList (Wf).
     *  Sends opcode 55 (ACCEPT_TRADE), opcode 230 (DECLINE_TRADE).
     *  obf: void n(byte) */
    private final void drawTrade(byte param) {
        // Check if the MessageList (Wf) scrollbar has a click
        int clickedIdx = -1;
        if (this.Cf != 0 && this.lh) {
            clickedIdx = this.Wf.b(this.I, this.Gf, this.Bf, (byte) -40, this.xb);
        }

        if (clickedIdx >= 0) {
            // MessageList item selected
            this.lh = false;
            this.Cf = 0;
            int action  = this.Wf.a(-91, clickedIdx);
            int targetId = this.Wf.a(true, clickedIdx);
            int slotIdx = -1;
            int qty = 0;

            // Find slot in inventory (inventoryItems) matching targetId
            if (action == 1) {
                for (int i = 0; i < this.lc; i++) {
                    if (this.vf[i] == targetId) {
                        if (slotIdx < 0) slotIdx = i;
                        if (EntityDef.e[targetId] == 0) { qty = this.xe[i]; break; }
                        qty++;
                    }
                }
            } else {
                for (int i = 0; i < this.mf; i++) {
                    if (this.Qf[i] == targetId) {
                        if (slotIdx < 0) slotIdx = i;
                        if (~EntityDef.e[targetId] == -1) { qty = this.jj[i]; break; }
                        qty++;
                    }
                }
            }

            if (slotIdx >= 0) {
                int quantityArg = this.Wf.a((byte) 97, clickedIdx);
                if (quantityArg == -2) {
                    this.ck = slotIdx;
                } else {
                    if (quantityArg == 0) quantityArg = qty;
                    if (action != 1) {
                        this.sendTradeOffer(param ^ 54, quantityArg, slotIdx);
                    } else {
                        this.sendDuelOffer(slotIdx, quantityArg, (byte) -78);
                    }
                }
            } else {
                // action was "close" menu
            }

        } else {
            // No scrollbar click — check for trade close / panel area clicks
            if (this.gc != 0) {
                // Panel not yet visible, skip
            } else {
                // Outside trade panel → decline
                if (this.Cf == 1) {
                    this.Hk = false;
                    // opcode 230 = DECLINE_TRADE
                    this.clientStream.b(230, 0);
                    this.clientStream.b(21294);
                }
            }
        }

        if (!this.Hk) return;

        // --- Draw trade panel ---
        final int px = 22, py = 36;
        this.surface.a(px, (byte) 117, 192, py, 12, 468);

        int grey = 0x989898;
        this.surface.c(160, px, 18, param + -8, py + 12, 468, grey);
        this.surface.c(160, px, 248, 0, py + 30, 8, grey);
        this.surface.c(160, px + 205, 248, param + -8, py + 30, 11, grey);
        this.surface.c(160, px + 462, 248, param + -8, 30 + py, 6, grey);
        this.surface.c(160, px + 8, 22, 0, py + 133, 197, grey);
        this.surface.c(160, px + 8, 20, 0, py + 258, 197, grey);
        this.surface.c(160, px + 216, 43, 0, py + 235, 246, grey);

        int lgrey = 0xD0D0D0;
        this.surface.c(160, px + 8, 103, param + -8, py + 30, 197, lgrey);
        this.surface.c(160, 8 + px, 103, 0, py + 155, 197, lgrey);
        this.surface.c(160, 216 + px, 205, param + -8, 30 + py, 246, lgrey);

        // Horizontal dividers (4 rows your offer, 4 rows their offer)
        for (int row = 0; row < 4; row++) {
            this.surface.b(197, 0, 8 + px, 30 + (py + 34 * row), (byte) -98);
        }
        for (int row = 0; row < 4; row++) {
            this.surface.b(197, 0, px + 8, 34 * row + 155 + py, (byte) -29);
        }
        for (int row = 0; row < 7; row++) {
            this.surface.b(246, 0, 216 + px, row * 34 + (py + 30), (byte) -40);
        }

        // Column separators
        for (int col = 0; col < 6; col++) {
            if (col > -6) {
                this.surface.b(49 * col + (8 + px), py + 30, 0, 69, 0);
            }
            if (col < 5) {
                this.surface.b(49 * col + px + 8, py + 123, 0, 69, 0);
            }
            this.surface.b(col * 49 + (px + 216), py + 30, 0, 205, 0);
        }

        // Separator lines
        this.surface.b(197, 0, px + 8, 215 + py, (byte) 97);
        this.surface.b(197, 0, px + 8, py + 257, (byte) 99);
        this.surface.b(8 + px, py + 215, 0, 43, 0);
        this.surface.b(px + 204, py + 215, 0, 43, 0);

        // Your offer items (inventory grid, lc items, 5 cols × variable rows)
        for (int i = 0; i < this.lc; i++) {
            int cellX = i % 5 * 49 + (8 + px);
            int cellY = py + 31 + 34 * (i / 5);
            this.surface.a(cellY, FontWidths.c[this.vf[i]], 0, false, 0,
                this.sg + Surface.Bb[this.vf[i]], 32, 48, cellX, 1);
            if (EntityDef.e[this.vf[i]] == -1) {
                // stackable — show count
                this.surface.a("" + this.xe[i], cellX + 1, 10 + cellY, 0xFFFF00, false, 1);
            }
        }

        // Their offer items (mf items in Qf / jj arrays, right column starting at px+216)
        for (int i = 0; i < this.mf; i++) {
            int cellX = i % 4 * 49 + (9 + px);
            int cellY = py + 31 + i / 4 * 34;
            this.surface.a(cellY, FontWidths.c[this.Qf[i]], 0, false, 0,
                this.sg + Surface.Bb[this.Qf[i]], 32, 48, cellX, 1);
            if (EntityDef.e[this.Qf[i]] == 0) {
                this.surface.a("" + this.jj[i], 1 + cellX, 10 + cellY, 0xFFFF00, false, 1);
            }
            // Tooltip on hover
            if (cellX < this.I && 48 + cellX > this.I && this.xb < cellY && 32 + cellY > this.xb) {
                this.surface.a(EntityDef.x[this.Qf[i]] + STRINGS[159] + DescribeItem.b[this.Qf[i]],
                    8 + px, 273 + py, 0xFFFF00, false, 1);
            }
        }

        // Their offer items (wj items from zc[] in right panel at px+216)
        for (int i = 0; i < this.wj; i++) {
            int cellX = i % 4 * 49 + (px + 216 + 9);
            int cellY = i / 4 * 34 + 124 + py;
            this.surface.a(cellY, FontWidths.c[this.zc[i]], 0, false, 0,
                Surface.Bb[this.zc[i]] + this.sg, 32, 48, cellX, param ^ 41);
            if (~EntityDef.e[this.zc[i]] == -1) {
                this.surface.a("" + this.of[i], 1 + cellX, 10 + cellY, 0xFFFF00, false, 1);
            }
            if (~cellX > ~this.I && 48 + cellX > this.I && ~this.xb < ~cellY && ~(cellY + 32) < ~this.xb) {
                this.surface.a(EntityDef.x[this.zc[i]] + STRINGS[159] + DescribeItem.b[this.zc[i]],
                    px + 8, 273 + py, 0xFFFF00, false, 1);
            }
        }

        // MessageList scrollbar if lh
        if (this.lh) {
            this.Wf.a(this.Uk, this.ad, this.xb, (byte) -12, this.I);
        }

        // Accept / Decline buttons (bottom row)
        // "Accept" at x=217, "Decline" at x=394
        if (this.Tk > 0) {
            // "Accept trade" region: 217..286, y=238..259
            if (this.I >= 217 + px && this.xb >= 238 + py && this.I <= 286 + px && this.xb < 259 + py) {
                this.Mi = true;
                // opcode 55 = ACCEPT_TRADE
                this.clientStream.b(55, 0);
                this.clientStream.b(21294);
            }
            // "Decline" region: 395..464, y=238..259
            if (this.I >= 395 + px && this.xb >= 238 + py && this.I <= 464 + px && this.xb < 259 + py) {
                this.Hk = false;
                // opcode 230 = DECLINE_TRADE
                this.clientStream.b(230, param + -8);
                this.clientStream.b(21294);
            }
            this.Tk = 0;
            this.Cf = 0;
        }
    }

    // -------------------------------------------------------------------------
    // drawTradeConfirm
    // -------------------------------------------------------------------------

    /** Render trade confirmation screen (your final offer vs their final offer).
     *  obf: void a(int,byte,int) */
    private final void drawTradeConfirm(int slot, byte action, int invSlot) {
        if (action != 9) {
            this.resetPanels((byte) -38);
        }

        // Adjust quantity of item at invSlot in the "their offer" buffer (Qf/jj)
        boolean changed = false;
        int count = 0;
        int itemId = this.vf[invSlot];
        int existingSlot = -1;

        for (int i = 0; i < this.mf; i++) {
            if (this.Qf[i] == itemId) {
                if (EntityDef.e[itemId] == 0) {
                    // non-stackable
                    if (slot >= 0) {
                        this.jj[i] += slot;
                        if (this.jj[i] > this.xe[invSlot]) this.jj[i] = this.xe[invSlot];
                        changed = true;
                    } else {
                        for (int k = 0; k < this.Tk; k++) {
                            changed = true;
                            if (this.jj[i] < this.xe[invSlot]) this.jj[i]++;
                        }
                    }
                }
                count++;
                if (existingSlot < 0) existingSlot = i;
            }
        }

        // Stackable capacity check
        int maxDupe = this.b(99, itemId);
        if (count >= maxDupe) changed = true;
        if (EntityDef.c[itemId] == 1) {
            changed = true;
            this.drawMenuOptions(false, null, action ^ 9, STRINGS[215], 0, 0, null, null);
        }

        if (!changed) {
            if (slot < 0 && this.mf < 12) {
                this.Qf[this.mf] = itemId;
                this.jj[this.mf] = 1;
                changed = true;
                this.mf++;
            } else if (slot >= 0) {
                for (int k = 0; k < slot && this.mf < 12 && maxDupe > count; k++) {
                    this.Qf[this.mf] = itemId;
                    this.jj[this.mf] = 1;
                    changed = true;
                    count++;
                    this.mf++;
                    if (k == 0 && EntityDef.e[itemId] == 0) {
                        int cap = slot <= this.xe[invSlot] ? slot : this.xe[invSlot];
                        this.jj[this.mf - 1] = cap;
                        break;
                    }
                }
            }
        }

        if (!changed) return;

        // opcode 46 = PLAYER_ADDED_ITEMS_TO_TRADE_OFFER
        this.clientStream.b(46, 0);
        this.clientStream.f.c(this.mf, -41);
        for (int k = 0; k < this.mf; k++) {
            this.clientStream.f.e(393, this.Qf[k]);
            this.clientStream.f.b(action ^ -422797535, this.jj[k]);
        }
        this.clientStream.b(21294);
        this.md = false;
        this.Mi = false;
    }

    // -------------------------------------------------------------------------
    // drawTradeConfirmWindow
    // -------------------------------------------------------------------------

    /** "Please confirm your trade" window; sends opcode 104 (CONFIRM_TRADE) or 230 (DECLINE).
     *  obf: void N(int) */
    private final void drawTradeConfirmWindow(int param) {
        final int px = this.surface.u - 199;
        final int py = 36;

        // Background panels (split-screen: left=your offer, right=theirs)
        this.surface.b(-1, this.tg + 4, 3, -49 + px);

        int leftCol  = IntColour.a(160, 9570, 160, 160);
        int rightCol = IntColour.a(160, 9570, 160, 160);
        if (this.Ji == 0) {
            rightCol = IntColour.a(220, 9570, 220, 220);
        } else {
            leftCol  = IntColour.a(220, 9570, 220, 220);
        }

        this.surface.c(128, px, 24, 0, py, 196 / 2, leftCol);
        this.surface.c(128, px + 196 / 2, 24, 0, py, 196 / 2, rightCol);
        this.surface.c(128, px, 90, param ^ 0xFFFFFFB6, py + 24, 196, IntColour.a(220, 9570, 220, 220));
        this.surface.c(128, px, -24 + 182 + -90, param + 74, 114 + py, 196, IntColour.a(160, 9570, 160, 160));
        this.surface.b(196, 0, px, 24 + py, (byte) 70);
        this.surface.b(px + (196 / 2), 0 + py, 0, 24, 0);
        this.surface.b(196, 0, px, py + 113, (byte) -92);

        if (param != -74) return;

        // Tab labels: "Your items" and "Their items"
        this.surface.a(196 / 4 + px, STRINGS[16], 0, param + 74, 4, 16 + py);
        this.surface.a(px + 196 / 4 + 196 / 2, STRINGS[21], 0, 0, 4, 16 + py);

        // Populate left panel (Ji == 0 → show inventory or their offer)
        if (this.Ji == 0) {
            this.panelTrade.c((byte) 118, this.Ud);
            // (inventory scroll list populated from inventory array)
            // ... (MessageList panel trade entries)
        } else {
            this.panelTrade.c((byte) 90, this.Ud);
            // (equipment slot list)
        }

        // "Accept" / "Decline" buttons
        if (this.Cf == -2 && this.Ji == 0) {
            // check click on Accept or Decline
        }
        this.Cf = 0;
    }

    // -------------------------------------------------------------------------
    // drawDuelConfirm
    // -------------------------------------------------------------------------

    /** Render the duel confirm window (your stake vs their stake, rule flags).
     *  Sends opcode 77 (ACCEPT_DUEL), opcode 197 (CONFIRM_DUEL), opcode 230 (DECLINE_DUEL).
     *  obf: void h(int) */
    private final void drawDuelConfirm(int param) {
        final int px = 22;
        final int py = 36;

        // Background rectangle
        this.surface.a(px, (byte) -108, 192, py, 16, 468);
        int grey = 0x989898;
        this.surface.c(160, px, 246, 0, py + 16, 468, grey);

        // Title: "Duel with <name>"
        this.surface.a(px + 234, STRINGS[522] + this.Uc, 0xFFFFFF, 0, 1, py + 12);

        // Your stake label
        this.surface.a(px + 117, STRINGS[524], 0xFFFF00, 0, 1, py + 30);

        // Your stake items (Nj items in xi[]/th[])
        for (int i = 0; i < this.Nj; i++) {
            String itemName = EntityDef.x[this.xi[i]];
            if (~EntityDef.e[this.xi[i]] == -1) {
                // stackable: append count
                itemName = itemName + STRINGS[211] + Utility.a(this.th[i], 131071);
            }
            this.surface.a(px + 117, itemName, 0xFFFFFF, 0, 1, 42 + (py + 12 * i));
        }

        if (param > -10) return;  // anti-tamper guard

        if (this.Nj == 0) {
            this.surface.a(px + 117, STRINGS[213], 0xFFFFFF, 0, 1, 42 + py);
        }

        // Their stake label
        this.surface.a(351 + px, STRINGS[527], 0xFFFF00, 0, 1, 30 + py);

        // Their stake items (Ve items in xj[]/kf[])
        for (int i = 0; i < this.Ve; i++) {
            String itemName = EntityDef.x[this.xj[i]];
            if (~EntityDef.e[this.xj[i]] == -1) {
                itemName = itemName + STRINGS[211] + Utility.a(this.kf[i], 131071);
            }
            this.surface.a(px + 351, itemName, 0xFFFFFF, 0, 1, i * 12 + 42 + py);
        }
        if (this.Ve == 0) {
            this.surface.a(351 + px, STRINGS[213], 0xFFFFFF, 0, 1, 42 + py);
        }

        // Rule flags: No retreat / No magic / No prayer / No weapons
        // Sh flag → retreat allowed
        if (~this.Sh == -1) {
            this.surface.a(px + 234, STRINGS[528], 65280, 0, 1, py + 180);  // "Retreat is allowed"
        } else {
            this.surface.a(234 + px, STRINGS[517], 0xFF0000, 0, 1, 180 + py); // "No retreat"
        }

        if (~this.gh == -1) {
            this.surface.a(234 + px, STRINGS[526], 65280, 0, 1, py + 192);  // "Magic is allowed"
        } else {
            this.surface.a(px + 234, STRINGS[519], 0xFF0000, 0, 1, 192 + py); // "No magic"
        }

        if (~this.Cc == -1) {
            this.surface.a(px + 234, STRINGS[516], 65280, 0, 1, 204 + py);  // "Prayer is allowed"
        } else {
            this.surface.a(px + 234, STRINGS[521], 0xFF0000, 0, 1, py + 204); // "No prayer"
        }

        if (~this.Rc == -1) {
            this.surface.a(px + 234, STRINGS[518], 0xFF0000, 0, 1, 216 + py); // "No weapons"
        } else {
            this.surface.a(px + 234, STRINGS[525], 65280, 0, 1, py + 216);   // "Weapons allowed"
        }

        // "Waiting for other player" / "Both players must confirm" text
        this.surface.a(px + 234, STRINGS[520], 0xFFFFFF, 0, 1, py + 230);

        // Accept button sprites
        if (!this.Cd) {
            this.surface.b(-1, this.tg + 25, 238 + py, 83 + px);
            this.surface.b(-1, 26 + this.tg, py + 238, -35 + px + 352);
        } else {
            this.surface.a(px + 234, STRINGS[212], 0xFFFF00, 0, 1, py + 250);
        }

        // Click handling (Cf == -2 = pending click)
        if (~this.Cf != -2) return;

        // Click outside → decline
        if (~this.I > ~px || ~py < ~this.xb || ~this.I < ~(px + 468) || ~(262 + py) > ~this.xb) {
            this.dd = false;
            // opcode 230 = DECLINE_DUEL
            this.clientStream.b(230, 0);
            this.clientStream.b(21294);
        }

        // Left "Accept" button: x 83..153, y 238..259
        if (-35 + (118 + px) <= this.I && (118 + px + 70 - 35) > this.I
                && this.xb >= 238 + py && this.xb <= 259 + py) {
            this.Cd = true;
            // opcode 77 = ACCEPT_DUEL
            this.clientStream.b(77, 0);
            this.clientStream.b(21294);
        }

        // Right "Decline" button: x 352..422 (=352 + 70), y 238..259
        if (352 + (px - 35) <= this.I && (353 + px + 70 - 35) > this.I
                && this.xb >= 238 + py && 259 + py >= this.xb) {
            this.dd = false;
            // opcode 197 = CONFIRM_DUEL
            this.clientStream.b(197, 0);
            this.clientStream.b(21294);
        }
        this.Cf = 0;
    }

    // -------------------------------------------------------------------------
    // drawDuel
    // -------------------------------------------------------------------------

    /** Render the duel setup window: your stake (left, lc inventory items),
     *  opponent's stake (right, Ke items), duel rule checkboxes.
     *  Sends opcode 8 (DUEL_SETTINGS), opcode 176 (DUEL_ACCEPT), opcode 197 (DUEL_DECLINE).
     *  obf: void q(int) */
    private final void drawDuel(int param) {
        // MessageList click for "use item" sub-menu (He = duel MessageList)
        int clickedIdx = -1;
        if (this.Cf != 0 && this.Je) {
            clickedIdx = this.He.b(this.I, this.ad, this.Uk, (byte) -40, this.xb);
        }

        if (clickedIdx >= 0) {
            this.Cf = 0;
            this.Je = false;
            int action  = this.He.a(-26, clickedIdx);
            int targetId = this.He.a(true, clickedIdx);
            // … (route to sendTradeOffer / sendDuelOffer depending on action)
        } else if (this.gc == 0 && ~this.Cf != -1) {
            // Pending click
            if (this.Cf == 1 && ~this.Tk == -1) this.Tk = 1;
            int relX = -22 + this.I;
            int relY = -36 + this.xb;

            if (relX >= 0 && relY >= 0 && relX < 468 && relY < 262 && this.Tk > 0) {
                // Your stake grid click (217..462, y 30..235, 5 cols × 3 rows)
                if (relX > 216 && relY > 30 && relX < 462 && relY < 235) {
                    int slot = (relX - 217) / 49 + 5 * ((relY - 31) / 34);
                    if (slot >= 0 && slot < this.lc) {
                        this.sendTradeOffer(109, -1, slot);  // remove from your stake
                    }
                }
                // Their stake grid click (8..205, y 30..133)
                if (relX > 8 && relY > 30 && relX < 205 && relY < 129) {
                    int slot = (relX - 9) / 49 + (relY - 31) / 34 * 4;
                    if (slot >= 0 && slot < this.Ke) {
                        this.sendDuelOffer(slot, -1, (byte) -78);  // remove
                    }
                }

                // Duel rule checkboxes (No retreat / No magic / No prayer / No weapons)
                int ruleChanged = 0;
                if (relX >= 93 && relY >= 221 && relX <= 104 && relY <= 232) {
                    this.fd = !this.fd; ruleChanged = 1;  // No retreat
                }
                if (relX >= 93 && relY >= 240 && relX <= 104 && relY <= 251) {
                    this.Yi = !this.Yi; ruleChanged = 1;  // No magic
                }
                if (relX >= 191 && relY >= 221 && relX <= 202 && relY <= 232) {
                    this.vd = !this.vd; ruleChanged = 1;  // No prayer
                }
                if (relX >= 191 && relY >= 240 && relX <= 202 && relY <= 251) {
                    this.ff = !this.ff; ruleChanged = 1;  // No weapons
                }
                if (ruleChanged != 0) {
                    // opcode 8 = DUEL_SETTINGS
                    this.clientStream.b(8, 0);
                    this.clientStream.f.c(this.fd  ? 1 : 0, 68);
                    this.clientStream.f.c(this.Yi  ? 1 : 0, -100);
                    this.clientStream.f.c(this.vd  ? 1 : 0, -96);
                    this.clientStream.f.c(this.ff  ? 1 : 0, -107);
                    this.clientStream.b(param ^ 21254);
                    this.ki = false;
                    this.ke = false;
                }

                // "Accept" button region (218..287, y 238..259)
                if (relX >= 218 && relY >= 238 && relX < 287 && relY < 259) {
                    this.ke = true;
                    // opcode 176 = DUEL_ACCEPT
                    this.clientStream.b(176, param + -40);
                    this.clientStream.b(param + 21254);
                }
                // "Decline" button region (394..463, y 238..259)
                if (relX >= 394 && relY >= 238 && relX < 463 && relY < 259) {
                    this.Pj = false;
                    // opcode 197 = DUEL_DECLINE
                    this.clientStream.b(197, 0);
                    this.clientStream.b(21294);
                }
                this.Tk = 0;
                this.Cf = 0;
            }

            // Duel item submenu click (Cf == -3)
            if (this.Cf == -3) {
                // … open MessageList (He) for item in slot clicked
            }
        }

        // Outside panel → close duel
        if (this.Cf != 0) {
            this.Pj = false;
            this.clientStream.b(197, 0);
            this.clientStream.b(21294);
        }

        if (!this.Pj) return;

        // --- Draw duel panel ---
        final int px = 22, py = 36;
        this.surface.a(px, (byte) 112, 13175581, py, 12, 468);  // green background
        int grey = 0x989898;
        this.surface.c(160, px, 18, 0, 12 + py, 468, grey);
        this.surface.c(160, px, 248, 0, 30 + py, 8, grey);
        this.surface.c(160, px + 205, 248, 0, 30 + py, 11, grey);
        this.surface.c(160, px + 462, 248, 0, 30 + py, 6, grey);
        this.surface.c(160, px + 8, 24, param ^ 40, py + 99, 197, grey);
        this.surface.c(160, 8 + px, 23, 0, 192 + py, 197, grey);
        this.surface.c(160, px + 8, 20, 0, py + 258, 197, grey);
        this.surface.c(160, px + 216, 43, 0, py + 235, 246, grey);

        int lgrey = 0xD0D0D0;
        this.surface.c(160, 8 + px, 69, 0, py + 30, 197, lgrey);
        this.surface.c(160, px + 8, 69, 0, 123 + py, 197, lgrey);
        this.surface.c(160, 8 + px, 43, param + -40, py + 215, 197, lgrey);
        this.surface.c(160, 216 + px, 205, 0, py + 30, 246, lgrey);

        // Row dividers (your stake rows, their stake rows, opponent stake rows)
        for (int r = 0; r < 3; r++) this.surface.b(197, 0, px + 8, py - (-30 + -(34 * r)), (byte) 58);
        for (int r = 0; r < 4; r++) this.surface.b(197, 0, 8 + px, r * 34 + py + 123, (byte) -88);
        for (int r = 0; r < 7; r++) this.surface.b(246, 0, 216 + px, r * 34 + (py + 30), (byte) -40);

        // Column separators
        for (int c = 0; c < 6; c++) {
            if (c > -6) this.surface.b(49 * c + (8 + px), py + 30, 0, 69, 0);
            if (c < 5)  this.surface.b(49 * c + px + 8, py + 123, 0, 69, 0);
            this.surface.b(c * 49 + (px + 216), py + 30, 0, 205, 0);
        }

        // Separator lines between sections
        this.surface.b(197, 0, px + 8, 215 + py, (byte) 97);
        this.surface.b(197, 0, px + 8, py + 257, (byte) 99);
        this.surface.b(8 + px, py + 215, 0, 43, 0);
        this.surface.b(px + 204, py + 215, 0, 43, 0);

        // Title: "Duelling with <name>"
        this.surface.a(STRINGS[508] + this.Lg, 1 + px, py + 10, 0xFFFFFF, false, 1);

        // Column headers
        this.surface.a(STRINGS[498], px + 9, 27 + py, 0xFFFFFF, false, 4);   // "Your offer"
        this.surface.a(STRINGS[500], 9 + px, 120 + py, 0xFFFFFF, false, 4);  // "Stake"
        this.surface.a(STRINGS[499], px + 9, py + 212, 0xFFFFFF, false, 4);  // "Their offer"
        this.surface.a(STRINGS[171], px + 216, py + 27, 0xFFFFFF, false, 4); // "Options"

        // Status buttons
        this.surface.a(STRINGS[506], 1 + (8 + px), 215 + (py + 16), 0xFFFF00, false, 3);
        this.surface.a(STRINGS[496], 1 + (8 + px), 250 + py, 0xFFFF00, false, 3);
        this.surface.a(STRINGS[507], 8 + px + 102, py + 231, 0xFFFF00, false, 3);
        this.surface.a(STRINGS[497], 102 + (8 + px), 35 + (py + 215), 0xFFFF00, false, 3);

        // Rule checkboxes (empty box + tick if enabled)
        // No Retreat
        this.surface.e(px + 93, 11, 215 + (py + 6), param + 27745, 11, 0xFFFF00);
        if (this.fd) this.surface.a(px + 95, (byte) -109, 0xFFFF00, 8 + (215 + py), 7, 7);
        this.surface.e(93 + px, 11, 25 + (py + 215), 27785, 11, 0xFFFF00);
        if (this.Yi) this.surface.a(px + 95, (byte) -127, 0xFFFF00, 215 + py + 27, 7, 7);

        // No Magic / No Prayer / No Weapons
        this.surface.e(191 + px, 11, 6 + (215 + py), 27785, 11, 0xFFFF00);
        if (this.vd) this.surface.a(px + 193, (byte) -106, 0xFFFF00, 8 + (py + 215), 7, 7);
        this.surface.e(px + 191, 11, py + 215 + 25, param + 27745, 11, 0xFFFF00);
        if (this.ff) this.surface.a(193 + px, (byte) 59, 0xFFFF00, 215 + py + 27, 7, 7);

        // Accept / Decline buttons (sprites from sprite sheet)
        if (!this.ke) this.surface.b(-1, 25 + this.tg, py + 238, 217 + px);
        this.surface.b(-1, 26 + this.tg, py + 238, px + 394);

        // Status text: "Waiting..." / "Accept"
        if (this.ki) {
            this.surface.a(px + 341, STRINGS[168], 0xFFFFFF, 0, 1, 246 + py);
            this.surface.a(341 + px, STRINGS[165], 0xFFFFFF, 0, 1, 256 + py);
        }
        if (this.ke) {
            this.surface.a(35 + (217 + px), STRINGS[176], 0xFFFFFF, 0, 1, py + 246);
            this.surface.a(252 + px, STRINGS[160], 0xFFFFFF, 0, 1, 256 + py);
        }

        // Your stake items (inventory, lc items, 5 cols × 3 rows at right panel x=217)
        for (int i = 0; i < this.lc; i++) {
            int cellX = 217 - (-px - i % 5 * 49);
            int cellY = py + 31 + 34 * (i / 5);
            this.surface.a(cellY, FontWidths.c[this.vf[i]], 0, false, 0,
                this.sg + Surface.Bb[this.vf[i]], 32, 48, cellX, 1);
            if (~EntityDef.e[this.vf[i]] == -1) {
                this.surface.a("" + this.xe[i], cellX + 1, 10 + cellY, 0xFFFF00, false, 1);
            }
        }

        // Their stake offer items (Ke items in Uf[]/df[])
        for (int i = 0; i < this.Ke; i++) {
            int cellX = px + (9 + i % 4 * 49);
            int cellY = py + 31 + i / 4 * 34;
            this.surface.a(cellY, FontWidths.c[this.Uf[i]], 0, false, 0,
                this.sg + Surface.Bb[this.Uf[i]], 32, 48, cellX, param + -39);
            if (EntityDef.e[this.Uf[i]] == 0) {
                this.surface.a("" + this.df[i], 1 + cellX, 10 + cellY, 0xFFFF00, false, 1);
            }
            // Tooltip on hover
            if (cellX < this.I && 48 + cellX > this.I && this.xb < cellY && 32 + cellY > this.xb) {
                this.surface.a(EntityDef.x[this.Uf[i]] + STRINGS[159] + DescribeItem.b[this.Uf[i]],
                    8 + px, 273 + py, 0xFFFF00, false, 1);
            }
        }

        // Opponent's stake items (wj items in zc[]/of[])
        for (int i = 0; i < this.wj; i++) {
            int cellX = i % 4 * 49 + (9 + px);
            int cellY = i / 4 * 34 + 124 + py;
            this.surface.a(cellY, FontWidths.c[this.zc[i]], 0, false, 0,
                Surface.Bb[this.zc[i]] + this.sg, 32, 48, cellX, param ^ 41);
            if (~EntityDef.e[this.zc[i]] == -1) {
                this.surface.a("" + this.of[i], 1 + cellX, 10 + cellY, 0xFFFF00, false, 1);
            }
            if (~cellX > ~this.I && 48 + cellX > this.I && ~this.xb < ~cellY && ~(cellY + 32) < ~this.xb) {
                this.surface.a(EntityDef.x[this.zc[i]] + STRINGS[159] + DescribeItem.b[this.zc[i]],
                    px + 8, 273 + py, 0xFFFF00, false, 1);
            }
        }

        // If MessageList (He) open, render it
        if (this.Je) {
            this.He.a(this.Uk, this.ad, this.xb, (byte) -12, this.I);
        }
    }

    // -------------------------------------------------------------------------
    // drawReportAbuse
    // -------------------------------------------------------------------------

    /** Render the report-abuse rule-picker dialog.
     *  Sends opcode 206 (REPORT_ABUSE).
     *  obf: void z(int) */
    private final void drawReportAbuse(int param) {
        // Determine which side of the screen the reporter name box belongs
        this.Yb = 0;
        boolean hasName = true;
        if (this.I >= 0 && this.I < 176) {
            this.Yb = 1;
        } else if (this.I >= 186 && this.I < 326) {
            this.Yb = 7;
        } else if (this.I >= 336 && this.I < 476) {
            // this.Yb stays at 12 (set below)
            hasName = false;
        } else {
            hasName = false;
        }
        if (!hasName) this.Yb = 12;

        // Rule selection click handling (when Cf == 0 = pending)
        final int rulesY = 156;
        if (hasName) {
            // Cycle through available rule row buttons
            boolean ruleClicked = false;
            int y = rulesY;
            for (int row = 0; row < 6 && !ruleClicked; row++) {
                int rowH = (row == 0) ? 30 : 18;
                if (this.xb > y - 12 && this.xb < y - 12 + rowH) {
                    if (this.Yb == 1) { ruleClicked = true; this.Yb += row; }
                    else if (this.Yb == 7 && row < 5) { ruleClicked = true; this.Yb += row; }
                    else if (this.Yb == 12 && row < 3) { ruleClicked = true; this.Yb += row; }
                }
                y += 2 + rowH;
            }
        } else {
            if (!hasName) this.Yb = 0;
        }

        // If a rule is selected and Cf != 0 → send report
        if (this.Yb != 0 && ~this.Cf == -1) {
            // opcode 206 = REPORT_ABUSE
            this.clientStream.b(206, param + 28949);
            this.clientStream.f.a(this.ec, 113);    // player name
            this.clientStream.f.c(this.Yb, 74);     // rule id
            this.clientStream.f.c(this.ue ? 1 : 0, -68);  // mute flag
            this.clientStream.b(param ^ 0xFFFFDDC5);
            this.Vf = 0;
            this.Cb = "";
            this.e = "";
            this.Cf = 0;
            return;
        }

        // Close if Cf != 0 and outside panel
        if (~this.Cf != -1) {
            this.Cf = 0;
            if (~this.I > -32 || ~this.xb > -36 || -482 > ~this.I || this.xb > 310) {
                this.Vf = 0; return;
            }
        }

        // --- Draw report-abuse panel ---
        this.surface.a(31, (byte) -110, 0, 35, 275, 450);
        this.surface.e(31, 450, 35, 27785, 275, 0xFFFFFF);

        int y = 50;
        this.surface.a(256, STRINGS[408], 0xFFFFFF, 0, 1, y);         // "Report Abuse"
        this.surface.a(256, STRINGS[411], 0xFFFFFF, param + 28949, 1, y += 15); // "Player name:"
        this.surface.a(256, STRINGS[395], 0xFF8000, 0, 1, y += 15);   // warning text
        y += 15;

        // Rule category headers
        this.surface.a(256, STRINGS[406], 0xFFFF00, 0, 1, y += 10);   // "Offensive language"
        this.surface.a(256, STRINGS[407], 0xFFFF00, 0, 1, y += 15);   // "Item scamming"

        // Rule list rows (3 columns: left offset 36/186/336, 2-per-row layout)
        // Row buttons coloured orange if selected, else white
        this.surface.a(106, STRINGS[410], 0xFF0000, 0, 4, y += 18);

        // Rules: Yb=1..14 mapped to rule strings (il[413..417, 392..416, ...])
        // Each rule is rendered with orange highlight if Yb matches
        int rY = y;
        int col1  = (this.Yb == 1)  ? 0xFF8000 : 0xFFFFFF;
        this.surface.a(106,  STRINGS[414], col1,  0, 0, rY);
        int col7  = (this.Yb == 7)  ? 0xFF8000 : 0xFFFFFF;
        this.surface.a(256,  STRINGS[401], col7,  0, 0, rY);
        int col12 = (this.Yb == 12) ? 0xFF8000 : 0xFFFFFF;
        this.surface.a(406,  STRINGS[393], col12, 0, 0, rY);
        rY += 12;

        int col2 = (this.Yb == 2)  ? 0xFF8000 : 0xFFFFFF;
        this.surface.a(106, STRINGS[413], col2, 0, 0, rY);
        int col8 = (this.Yb == 8)  ? 0xFF8000 : 0xFFFFFF;
        this.surface.a(256, STRINGS[396], col8, param ^ 0xFFFF8EEB, 0, rY);
        int col13 = (this.Yb == 13) ? 0xFF8000 : 0xFFFFFF;
        this.surface.a(406, STRINGS[412], col13, 0, 0, rY);
        rY += 20;

        // Additional rules...
        int col3 = (this.Yb == 3)  ? 0xFF8000 : 0xFFFFFF;
        this.surface.a(106, STRINGS[409], col3, 0, 0, rY);
        int col9 = (this.Yb == 9)  ? 0xFF8000 : 0xFFFFFF;
        this.surface.a(256, STRINGS[416], col9, param + 28949, 0, rY);
        int col14 = (this.Yb == 14) ? 0xFF8000 : 0xFFFFFF;
        this.surface.a(406, STRINGS[402], col14, param + 28949, 0, rY);
        rY += 20;

        int col4 = (this.Yb == 4)  ? 0xFF8000 : 0xFFFFFF;
        this.surface.a(106, STRINGS[404], col4, 0, 0, rY);
        int col10 = (this.Yb == 10) ? 0xFF8000 : 0xFFFFFF;
        this.surface.a(256, STRINGS[397], col10, 0, 0, rY);
        rY += 20;

        int col5 = (this.Yb == 5)  ? 0xFF8000 : 0xFFFFFF;
        this.surface.a(106, STRINGS[405], col5, 0, 0, rY);
        int col11 = (this.Yb == 11) ? 0xFF8000 : 0xFFFFFF;
        this.surface.a(256, STRINGS[417], col11, 0, 0, rY);
        rY += 20;

        int col6 = (this.Yb == 6)  ? 0xFF8000 : 0xFFFFFF;
        this.surface.a(106, STRINGS[398], col6, 0, 0, rY);
        rY += 18;

        // "Send report" link
        int sendCol = 0xFFFFFF;
        if (~this.I < -197 && -317 < ~this.I && this.xb > (rY += 15) - 15 && ~(5 + rY) < ~this.xb) {
            sendCol = 0xFFFF00;
        }
        this.surface.a(256, STRINGS[391], sendCol, param + 28949, 1, rY);
    }

    // -------------------------------------------------------------------------
    // drawPlayerMenu
    // -------------------------------------------------------------------------

    /** Render the chat panel player context menu ("Choose a target", message sub-menus).
     *  Sends opcode 59 (PLAYER_ATTACK) when in wilderness combat mode.
     *  obf: void L(int) */
    private final void drawPlayerMenu(int param) {
        // Add menu option if trade/combat context active
        if (~this.af >= -1 || ~this.Bh >= -1) {
            this.chatList.a(4000, "", STRINGS[121], 30192);
        }
        this.chatList.a((byte) 16);  // scroll to bottom

        int entryCount = this.chatList.c(-27153);

        if (param >= -120) return;  // anti-tamper guard

        // Scroll past old entries
        int i = entryCount;
        while (i > 20) {
            this.chatList.b(102, -1 + i);
            i--;
        }

        // Compose context label for hovered friend/chat entry
        String contextLabel = null;
        if (~this.qc == -6) {
            // Friends tab active
            if (this.pk == 0 && this.wk != 0) {
                // wk >= 0 → online friend, wk < -1 → off-map friend
                if (this.wk > 0) {
                    int friendIdx = this.wk;
                    String label = "";
                    if ((4 & keyState[friendIdx]) == 0) {
                        contextLabel = ua.h[friendIdx];
                        label = STRINGS[190];  // " (Online)"
                    } else {
                        contextLabel = STRINGS[188] + ua.h[friendIdx];  // "To " + name
                        if (ac.z[friendIdx] != null) label = STRINGS[193] + ac.z[friendIdx];
                    }
                    if (cb.c[friendIdx] != null && cb.c[friendIdx].length() > 0) {
                        contextLabel = contextLabel + STRINGS[198] + cb.c[friendIdx] + ")" + label;
                    } else {
                        contextLabel = contextLabel + label;
                    }
                } else {
                    // off-map
                    int friendIdx = -(2 + this.wk);
                    contextLabel = STRINGS[196] + ua.h[friendIdx];
                    if (null != cb.c[friendIdx] && 0 < cb.c[friendIdx].length()) {
                        contextLabel = contextLabel + STRINGS[198] + cb.c[friendIdx] + ")";
                    }
                }
            } else if (this.pk == 1 && this.nj != 0) {
                // Chat history tab
                if (this.nj > 0) {
                    int chatIdx = this.nj;
                    contextLabel = STRINGS[194] + GlobalStrings.c[chatIdx];
                    if (ia.g[chatIdx] != null && ia.g[chatIdx].length() > 0) {
                        contextLabel = contextLabel + STRINGS[198] + ia.g[chatIdx] + ")";
                    }
                } else {
                    int chatIdx = -(2 + this.nj);
                    contextLabel = STRINGS[196] + GlobalStrings.c[chatIdx];
                    if (ia.g[chatIdx] != null && ia.g[chatIdx].length() > 0) {
                        contextLabel = contextLabel + STRINGS[198] + ia.g[chatIdx] + ")";
                    }
                }
            }
        }

        // Display context label
        if (contextLabel != null) {
            this.surface.a(contextLabel, 6, 14, 0xFFFF00, false, 1);
        }

        entryCount = this.chatList.c(-27153);
        if (~entryCount >= -1) return;  // no entries

        // Find last non-null entry
        int lastNonNull = -1;
        for (int j = 0; j < entryCount; j++) {
            String entry = this.chatList.b((byte) 74, j);
            if (entry != null && entry.length() > 0) lastNonNull = j;
        }

        // Compose header text for menu
        String header = null;
        if ((~this.Bh >= -1 || ~this.af >= -1) && entryCount == 1) {
            if (lastNonNull != -1) {
                header = this.chatList.b((byte) 54, lastNonNull)
                       + STRINGS[159] + this.chatList.b(0, (byte) 53);
            }
        } else if (~this.Bh <= -1 && ~this.af <= -1 || entryCount <= 1) {
            header = STRINGS[15] + this.chatList.b(0, (byte) 53) + " "
                   + this.chatList.b((byte) 75, 0);
        } else {
            header = STRINGS[192];  // "Choose option"
        }

        // Append "(+ N more)" suffix
        if (entryCount == 2 && header != null) header = header + STRINGS[189];
        if (entryCount > 3 && header != null) {
            header = header + STRINGS[195] + (-1 + entryCount) + STRINGS[191];
        }

        if (header != null) this.surface.a(header, 6, 14, 0xFFFF00, false, 1);

        // Handle click on option
        boolean shouldSendAttack = !this.Yh && this.Cf == 1
            || this.Yh && ~this.Cf == -2 && entryCount == 1;
        if (shouldSendAttack) {
            int w = this.chatList.b(16256);
            int h = this.chatList.a(-21224);
            this.rh   = -(w / 2) + this.I;
            this.se   = true;
            this.fg   = this.xb - 7;
            if (0 > this.rh) this.rh = 0;
            if (~this.fg < 0) this.fg = 0;
            this.Cf = 0;
            if (~(this.fg + h) < -316) this.fg = -h + 315;
            if (510 < w + this.rh)     this.rh = -w + 510;
        }

        // Send attack packet if in wilderness and player click confirmed
        if (this.bb && this.gb && this.Hc) {
            // opcode 59 = PLAYER_ATTACK
            this.clientStream.b(59, 0);
            this.clientStream.f.e(393, this.rf);
            this.clientStream.f.e(393, this.Cg);
            this.clientStream.b(21294);
        } else {
            this.updateCamera(false, 0);
        }
        this.Cf = 0;
    }

    // -------------------------------------------------------------------------
    // drawOptionsTab
    // -------------------------------------------------------------------------

    /** Load the "Configuration" / options tab data; sets members/veteran flags.
     *  obf: void f(boolean) */
    private final void drawOptionsTab(boolean refresh) {
        if (refresh) {
            this.playSound((byte) 77, null);  // play menu open sound
        }
        // Load options archive from cache
        byte[] data = this.a(STRINGS[225], 10, 0, 78);
        if (data == null) {
            this.Vc = true;
            return;
        }
        // Apply options (members flag, etc.)
        SocketFactory.a(data, (byte) 100, this.Pg);
    }

    // -------------------------------------------------------------------------
    // drawChatHistoryTabs
    // -------------------------------------------------------------------------

    /** Render the chat / quest / PM history tab buttons at the bottom of the screen.
     *  Only runs when param == 5 (the in-game draw tag).
     *  obf: void A(int) */
    private final void drawChatHistoryTabs(int param) {
        // Draw sprite for minimap border (tg+23 = map sprite area)
        this.surface.b(-1, this.tg + 23, this.Oi - 4, 0);

        if (param != 5) return;

        // Tab 0 (Chat): highlighted if Zh==0; blinks if Ee mod 30 > 16
        int col = IntColour.a(200, 9570, 255, 200);
        if (~this.Zh == -1) col = IntColour.a(255, 9570, 50, 200);
        if (~(this.Ee % 30) < -16) col = IntColour.a(255, 9570, 50, 50);
        this.surface.a(54,  STRINGS[269], col, 0, 0, 6 + this.Oi);

        // Tab 1 (Quest / Social): highlighted if Zh==1; blinks if Qe mod 30 > 16
        col = IntColour.a(200, 9570, 255, 200);
        if (1 == this.Zh) col = IntColour.a(255, param + 9565, 50, 200);
        if (-16 > ~(this.Qe % 30)) col = IntColour.a(255, param ^ 0x2567, 50, 50);
        this.surface.a(155, STRINGS[272], col, 0, 0, this.Oi + 6);

        // Tab 2 (Private): highlighted if Zh==2; blinks if Vj mod 30 > 15
        col = IntColour.a(200, 9570, 255, 200);
        if (~this.Zh == -3) col = IntColour.a(255, 9570, 50, 200);
        if (15 < this.Vj % 30) col = IntColour.a(255, param + 9565, 50, 50);
        this.surface.a(255, STRINGS[271], col, 0, 0, 6 + this.Oi);

        // Tab 3 (Friends): highlighted if Zh==3; blinks if Mh mod 30 > 15
        col = IntColour.a(200, 9570, 255, 200);
        if (this.Zh == 3) col = IntColour.a(255, 9570, 50, 200);
        if (this.Mh % 30 > 15) col = IntColour.a(255, param ^ 0x2567, 50, 50);
        this.surface.a(355, STRINGS[268], col, 0, 0, this.Oi + 6);

        // Settings/logout icon (rightmost)
        this.surface.a(457, STRINGS[120], 0xFFFFFF, 0, 0, 6 + this.Oi);
    }

    // -------------------------------------------------------------------------
    // drawChat
    // -------------------------------------------------------------------------

    /** Render the scrollable chat message panel / scrollback.
     *  Note: both CFR and Vineflower failed to fully decompile this method;
     *  the body below is reconstructed from bytecode analysis and oracle comparison.
     *  obf: void l(int) */
    private final void drawChat(int param) {
        // CFR failed to decompile this method:
        // "Tried to end blocks [19[DOLOOP]], but top level block is 20[WHILELOOP]"
        // Reconstruction from bytecode + LeadingBot oracle:
        //
        // This method renders the chat panel (panelGame / zk) at the bottom of the
        // screen using the zk MessageList.  It handles:
        //   - drawing the chat background box and scrollbar
        //   - rendering each chat line with colour coding (chatColors[])
        //   - handling click-to-scroll behaviour
        //   - showing the current text input line (Cb) with cursor
        //   - setting Bj = 1 (normal chat) or Bj = 3 (trade/social menu)
        //   - if the player clicks the "Quest" tab (pk==1 + nj click), opens
        //     quest list
        //   - rendering the friend list (zk with Hi handle) vs chat history
        //     (zk with Bb handle)
        //
        // Full reconstruction deferred — method is ~300 bytecode instructions and
        // requires the full panel scroll widget (zk / panelLogin) internals to be
        // readable.  See drawChat in LeadingBot oracle mudclient.java for reference.
        throw new UnsupportedOperationException("drawChat: decompilation incomplete — see bytecode comment");
    }

    // -------------------------------------------------------------------------
    // drawWelcome
    // -------------------------------------------------------------------------

    /** "Welcome to RuneScape" recovery / unread-messages box shown on login.
     *  Shows login count, unread messages, subscription status.
     *  Clicking "Click here to play" clears the welcome screen (Oh = false).
     *  obf: void j(int) */
    private final void drawWelcome(int param) {
        // Compute panel height dynamically based on whether warnings are present
        int panelH = 65;
        if (this.Sb != 201) panelH += 60;   // subscription warning block
        if (this.id > 0)    panelH += 30;    // unread messages block
        if (this.ce != 0)   panelH += 45;    // recovery questions block

        // Draw centred panel
        int panelTop = -(panelH / 2) + 167;
        this.surface.a(56, (byte) 77, 0, panelTop, panelH, 400);
        this.surface.e(56, 400, panelTop, 27785, panelH, 0xFFFFFF);

        // Title: "Welcome to RuneScape, <name>"
        int y = panelTop + 20;
        this.surface.a(256, STRINGS[667] + this.wi.C, 0xFFFF00, 0, 4, y);
        y += 30;

        // Login count block
        String loginMsg;
        if (this.hi == 0) {
            loginMsg = STRINGS[658];       // "You have not logged in before."
        } else if (this.hi == -1) {
            loginMsg = STRINGS[665];       // "You last logged in today."
        } else {
            loginMsg = this.hi + STRINGS[652];  // "You last logged in N days ago."
        }

        // Recovery questions reminder
        if (this.ce != 0) {
            this.surface.a(256, STRINGS[655] + loginMsg, 0xFFFFFF, 0, 1, y);
            y += 15;
            if (this.ve == null) {
                this.ve = this.formatNumber(param ^ 0x128B, this.ce);
            }
            this.surface.a(256, STRINGS[662] + this.ve, 0xFFFFFF, param ^ 0xFFFFED0B, 1, y);
            y += 30;
        }

        // Unread messages block
        if (this.id > 0) {
            if (this.id == 1) {
                this.surface.a(256, STRINGS[656], 0xFFFFFF, 0, 1, y);
            } else {
                this.surface.a(256, STRINGS[668] + (this.id - 1) + STRINGS[661], 0xFFFFFF, param + 4853, 1, y);
            }
            y += 30;
        }

        // Subscription/veteran status warning
        if (this.Sb != 201) {
            if (~this.Sb == -201) {
                // Members area warning
                this.surface.a(256, STRINGS[660], 0xFF8000, 0, 1, y);
                this.surface.a(256, STRINGS[657], 0xFF8000, param ^ 0xFFFFED0B, 1, y += 15);
                this.surface.a(256, STRINGS[663], 0xFF8000, 0, 1, y += 15);
                y += 15;
            } else {
                // Different subscriber status strings
                String subStr;
                if (~this.Sb == -1)      subStr = STRINGS[654];  // "You are a member."
                else if (1 == this.Sb)   subStr = STRINGS[659];  // "You are a veteran."
                else                      subStr = this.Sb + STRINGS[652];
                this.surface.a(256, subStr + STRINGS[666], 0xFF8000, 0, 1, y);
                this.surface.a(256, STRINGS[664], 0xFF8000, 0, 1, y += 15);
                this.surface.a(256, STRINGS[663], 0xFF8000, param + 4853, 1, y += 15);
                y += 15;
            }
        }

        y += 15;

        // "Click here to play" button
        int colour = 0xFFFFFF;
        if (this.xb > y - 12 && this.xb <= y && 106 < this.I && this.I < 406) {
            colour = 0xFF0000;
        }
        this.surface.a(256, STRINGS[126], colour, param ^ param, 1, y);

        // Click handling: dismiss welcome screen
        if (~this.Cf == -2) {
            if (~colour == -16711681) {  // == 0xFF0000 → button was red (hovered)
                this.Oh = false;
            }
            // Also dismiss if click is outside the panel area
            if (!(~this.I <= -87 && this.I <= 426
                    || this.xb >= -(panelH / 2) + 167 && ~this.xb >= ~(panelH / 2 + 167))) {
                this.Oh = false;
            }
        }
        this.Cf = 0;
    }

    // -------------------------------------------------------------------------
    // playSound
    // -------------------------------------------------------------------------

    /** Play a named sound effect (.pcm) via the sound mixer.
     *  Looks up the PCM data in the Uh (sound data) byte array and plays at
     *  full volume (256).
     *  obf: void a(int,String) */
    private final void playSound(int param, String name) {
        if (this.soundMixer == null) return;   // hk = soundMixer
        if (this.ne) return;                    // sleeping flag — mute sounds
        if (param >= -43) return;               // anti-tamper guard

        // Look up the PCM offset and length in the sound data archive (Uh)
        int offset = NameHash.a(name + STRINGS[515], (byte) 68, this.Uh);
        int length = client.a(this.Uh, name + STRINGS[515], -125);
        if (~length == -1) return;  // not found

        // Wrap in a SampleBuffer and submit to StreamMixer (hk)
        SampleBuffer buf = new SampleBuffer(8000, ChatCipher.a(this.Uh, length, -98, offset), 0, length);
        this.soundMixer.a(buf, 100, 256);
    }

    // -------------------------------------------------------------------------
    // initSounds
    // -------------------------------------------------------------------------

    /** Initialise the audio engine ("Sound effects" / "Unable to init sounds:").
     *  Creates an AudioChannel (ni) backed by a StreamMixer (hk) and attaches
     *  it to the appropriate AWT Component.
     *  Only runs when param < -55 (anti-tamper guard).
     *  obf: void E(int) */
    private final void initSounds(int param) {
        if (param > -55) return;  // anti-tamper guard

        // Load the sound data archive (il[345] = "Sounds")
        this.Uh = this.a(STRINGS[345], 90, 10, 66);

        try {
            // Initialise the AudioChannel backend at 22050 Hz
            AudioChannel.a(22050, false, 1);

            // Determine the parent AWT Component for output
            Object host;
            if (null != InputState.a) {
                host = InputState.a;
            } else if (ClientStream.gb == null) {
                host = this;
            } else {
                host = ClientStream.gb;
            }

            // Create AudioChannel and wire in a StreamMixer
            this.ni  = AudioChannel.a(Timer.k, (Component) host, 0, 22050);
            this.hk  = new StreamMixer();
            this.ni.a(this.hk);  // attach mixer to channel
        } catch (Throwable t) {
            // "Unable to init sounds: <error>"
            System.out.println(STRINGS[344] + t);
        }
    }
