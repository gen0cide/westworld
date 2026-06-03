package client;

import client.data.*;     // CacheFile, RecordLoader, NameHash, DataStore, FontWidths ...
import client.net.*;      // ClientStream, TextEncoder, ISAAC ...
import client.scene.*;    // SurfaceSprite, Surface ...
import client.shell.*;    // InputState ...
import client.ui.*;       // MessageList, CharTable, FontBuilder ...
import client.util.*;     // ClientIOException, Utility, DecodeBuffer ...

/**
 * GameInterface — the modal-panel renderers extracted from Mudclient (DE-GOD seam 3).
 *
 * <p>The full set of open-panel renderers that draw onto {@code m.li}, hit-test against
 * {@code m.mouseX}/{@code m.mouseY}, mutate panel-state flags / selection scalars, and emit
 * shop / bank / trade / duel packets. Each is dispatched from
 * {@code Mudclient.drawActiveInterface} / {@code drawGameFrame}, which stay in Mudclient and
 * re-route to {@code m.gameInterface.drawShop(...)} etc.
 *
 * <p>All field access and shared-routine callbacks (menuHitTest, drawMenuOptions, walkTo,
 * showServerMessage, sendTradeOffer / sendDuelOffer / bankSend via the trade-packets delegate,
 * resetPanels, drawOptionsTab, requestLogout, formatNumber, the {@code packets} delegate, …)
 * are reached through the {@code m} back-reference. Cross-package statics
 * (ClientIOException.STRING_TABLE, DataStore.strayUiStrings, Mudclient.STRINGS, ISAAC.*, …)
 * are kept as-is.
 *
 * <p>This is pure code motion: every opcode number, putShort/putInt order, the 49x34 grid
 * hit-test math, and the exact click/render interleave are reproduced unchanged. The only
 * rewrites are the {@code this.} / bare-identifier → {@code m.} qualifier changes and statics
 * → {@code Mudclient.}.
 *
 * <p>Methods (obf): drawShop M, drawBank r, drawTrade n, drawTradeConfirm a(int,byte,int),
 * drawTradeConfirmWindow hasPainted, drawDuelConfirm h, drawDuel q,
 * drawGameSettings b(int,boolean), drawWildernessWarning H, drawHelpMenu C,
 * drawCloseButton l(byte), drawWelcome j.
 */
class GameInterface {
    final Mudclient m;

    GameInterface(Mudclient m) {
        this.m = m;
    }

    /** "Warning! Proceed with caution" wilderness-entry dialog. Sets le=2 to enter wilderness
     *  mode on click (either on the "Click here to proceed" line or outside the panel bounds). */
    void drawWildernessWarning(int param) {
        m.li.drawBox(86, (byte) -115, 0, 77, 180, 340);   // panel fill at (86,77) 180x340
        int y = 97;
        if (param <= 90) {              // (anti-tamper-safe path) also render the options tab beneath
            m.drawOptionsTab(true);
        }
        m.li.drawBoxEdge(86, 340, 77, 27785, 180, 0xFFFFFF);  // white border

        m.li.drawStringCenter(256, Mudclient.STRINGS[307], 0xFF0000, 0, 4, y);          // "Warning!"
        m.li.drawStringCenter(256, Mudclient.STRINGS[305], 0xFFFFFF, 0, 1, y += 26);
        m.li.drawStringCenter(256, Mudclient.STRINGS[300], 0xFFFFFF, 0, 1, y += 13);
        m.li.drawStringCenter(256, Mudclient.STRINGS[306], 0xFFFFFF, 0, 1, y += 13);
        m.li.drawStringCenter(256, Mudclient.STRINGS[308], 0xFFFFFF, 0, 1, y += 22);
        m.li.drawStringCenter(256, Mudclient.STRINGS[301], 0xFFFFFF, 0, 1, y += 13);
        m.li.drawStringCenter(256, Mudclient.STRINGS[302], 0xFFFFFF, 0, 1, y += 22);
        m.li.drawStringCenter(256, Mudclient.STRINGS[303], 0xFFFFFF, 0, 1, y += 13);

        // "Click here to proceed" — red on hover
        int colour = 0xFFFFFF;
        y += 22;
        if (m.mouseY > y - 12 && m.mouseY <= y && m.mouseX > 181 && m.mouseX < 331) {
            colour = 0xFF0000;
        }
        m.li.drawStringCenter(256, Mudclient.STRINGS[126], colour, 0, 1, y);

        if (m.Cf != 0) {
            if (m.mouseY > y - 12 && m.mouseY <= y && m.mouseX > 181 && m.mouseX < 331) {
                m.le = 2;
            }
            m.Cf = 0;
            // click anywhere outside the panel rect also confirms
            if (m.mouseX < 86 || m.mouseX > 426 || m.mouseY < 77 || m.mouseY > 257) {
                m.le = 2;
            }
        }
    }

    /** Shop buy/sell panel ("Buying and selling items"). Hit-tests the 4x5 item grid plus the
     *  buy and sell quantity buttons (1/5/10/50/X). Opcodes: 236 BUY_ITEM, 221 SELL_ITEM,
     *  166 CLOSE_SHOP.
     *
     *  FIX vs old: grid hit-test had a duplicated `relY > cellY` (should be `relY < cellY+34`);
     *  the buy "X" dialog used ClientIOException.a (fa.a) not "ArchiveReader.u". */
    void drawShop(int param) {
        if (m.Cf != 0 && m.gc == 0) {
            m.Cf = 0;
            int relX = m.mouseX - 52;
            int relY = m.mouseY - 44;
            // click outside the shop grid → CLOSE_SHOP (clean: break label565 → opcode 166)
            if (relX < 0 || relY < 12 || relX >= 408 || relY >= 246) {
                m.Jh.newPacket(166, 0);   // CLOSE_SHOP
                m.Jh.finishPacket(21294);
                m.uk = false;
                return;
            }

            // item-grid hit-test (4 rows x 5 cols)
            int slot = 0;
            for (int row = 0; row < 4; row++) {
                for (int col = 0; col < 5; col++) {
                    int cellX = col * 49 + 7;
                    int cellY = row * 34 + 28;
                    if (relX > cellX && cellX + 49 > relX
                            && relY > cellY && relY < cellY + 34    // FIX: was `relY > cellY` twice
                            && m.Rj[slot] != -1) {
                        m.Di = slot;
                        m.fh = m.Rj[slot];
                    }
                    slot++;
                }
            }

            // Di >= 0  (the clean `var44(=0) <= Di`) → an item is selected
            if (m.Di >= 0) {
                int itemId = m.Rj[m.Di];
                if (itemId != -1) {
                    int stock = m.Jf[m.Di];
                    // --- buy row (y 204..215) ---
                    if (stock > 0 && relY >= 204 && relY <= 215) {
                        int qty = 0;
                        if (relX > 318 && relX < 330) qty = 1;
                        if (stock >= 5  && relX > 333 && relX < 345) qty = 5;
                        if (stock >= 10 && relX > 348 && relX < 365) qty = 10;
                        if (stock >= 50 && relX > 368 && relX < 385) qty = 50;
                        if (relX > 388 && relX < 400) {
                            m.drawMenuOptions(ClientIOException.STRING_TABLE, 12, 5, true);   // "X" → quantity entry
                        }
                        if (qty > 0) {
                            m.Jh.newPacket(236, 0);               // BUY_ITEM
                            m.Jh.outBuffer.putShort(m.Rj[m.Di]);
                            m.Jh.outBuffer.putShort(stock);
                            m.Jh.outBuffer.putShort(qty);
                            m.Jh.finishPacket(21294);
                        }
                    }
                    // --- sell row (y 229..240) ---
                    int held = m.menus.menuHitTest(102, itemId);   // how many the player owns
                    if (held > 0 && relY >= 229 && relY <= 240) {
                        int qty = 0;
                        if (relX > 318 && relX < 330) qty = 1;
                        if (held >= 5  && relX > 333 && relX < 345) qty = 5;
                        if (held >= 10 && relX > 348 && relX < 365) qty = 10;
                        if (relX > 388 && relX < 400) {
                            m.drawMenuOptions(DataStore.strayUiStrings, 12, 6, true);   // "X" → quantity entry
                        }
                        if (held >= 50 && relX > 368 && relX < 385) qty = 50;
                        if (qty > 0) {
                            m.Jh.newPacket(221, 0);               // SELL_ITEM
                            m.Jh.outBuffer.putShort(m.Rj[m.Di]);
                            m.Jh.outBuffer.putShort(stock);
                            m.Jh.outBuffer.putShort(qty);
                            m.Jh.finishPacket(21294);
                        }
                    }
                }
            }
        }

        // --- draw panel ---
        final int px = 52, py = 44;
        m.li.drawBox(px, (byte) 101, 192, py, 12, 408);
        int grey = 0x989898;
        m.li.drawBoxAlpha(160, px, 17, 0, py + 12, 408, grey);
        m.li.drawBoxAlpha(160, px, 170, 0, py + 29, 8, grey);
        m.li.drawBoxAlpha(160, px + 399, 170, 0, py + 29, 9, grey);
        m.li.drawBoxAlpha(160, px, 47, 0, py + 199, 408, grey);
        m.li.drawstring(Mudclient.STRINGS[640], px + 1, py + 10, 0xFFFFFF, false, 1);   // title

        int closeCol = 0xFFFFFF;
        if (m.mouseX > px + 320 && m.mouseY >= py && m.mouseX < px + 408 && m.mouseY < py + 12) {
            closeCol = 0xFF0000;
        }
        m.li.drawstringRightSimple(px + 406, Mudclient.STRINGS[620], py + 10, closeCol, -92, 1);   // "Close window"
        m.li.drawstring(Mudclient.STRINGS[637], px + 2, py + 24, 0x00FF00, false, 1);   // "Buy"
        m.li.drawstring(Mudclient.STRINGS[635], px + 135, py + 24, 0x00FFFF, false, 1); // "Sell"
        m.li.drawstring(Mudclient.STRINGS[643] + m.menus.menuHitTest(84, 10) + Mudclient.STRINGS[631], px + 280, py + 24, 0xFFFF00, false, 1);

        // item grid 4x5
        int grey2 = 0xD0D0D0;
        int slot = 0;
        for (int row = 0; row < 4; row++) {
            for (int col = 0; col < 5; col++) {
                int cellX = col * 49 + 7 + px;
                int cellY = py + 28 + 34 * row;
                if (m.Di == slot) {
                    m.li.drawBoxAlpha(160, cellX, 34, 0, cellY, 49, 0xFF0000);
                } else {
                    m.li.drawBoxAlpha(160, cellX, 34, 0, cellY, 49, grey2);
                }
                m.li.drawBoxEdge(cellX, 50, cellY, 27785, 35, 0);
                if (m.Rj[slot] != -1) {
                    m.li.spriteClipping(cellY, TextEncoder.scratchIntArray2[m.Rj[slot]], 0, false, 0,
                        Surface.unusedIntsBb[m.Rj[slot]] + m.sg, 32, 48, cellX, 1);
                    m.li.drawstring("" + m.Jf[slot], cellX + 1, cellY + 10, 0x00FF00, false, 1);
                    m.li.drawstringRightSimple(cellX + 47, "" + m.menus.menuHitTest(85, m.Rj[slot]), cellY + 10, 0x00FFFF, -80, 1);
                }
                slot++;
            }
        }
        m.li.drawLineHoriz(398, 0, px + 5, py + 222, (byte) -103);   // scrollbar

        // selected-item detail (buy / sell rows)
        if (m.Di != -1) {
            int itemId = m.Rj[m.Di];
            if (itemId != -1) {
                int stock = m.Jf[m.Di];
                // buy line
                if (stock <= 0) {
                    m.li.drawStringCenter(px + 204, Mudclient.STRINGS[641], 0xFFFF00, 0, 3, py + 214); // "out of stock"
                } else {
                    int buyPrice = ISAAC.scaledPercentage(InputState.pixelData[itemId], m.vi[m.Di], m.xk, -30910, true, 1, stock, m.Pf);
                    m.li.drawstring(DecodeBuffer.chatFilterCache[itemId] + Mudclient.STRINGS[639] + buyPrice + Mudclient.STRINGS[636],
                        px + 2, py + 214, 0xFFFF00, false, 1);
                    boolean inBuyRow = m.mouseY >= py + 204 && m.mouseY <= py + 215;
                    m.li.drawstring(Mudclient.STRINGS[642], px + 285, py + 214, 0xFFFFFF, false, 3);
                    int c = 0xFFFFFF;
                    if (inBuyRow && m.mouseX > px + 318 && m.mouseX < px + 330) c = 0xFF0000;
                    m.li.drawstring("1", px + 320, py + 214, c, false, 3);
                    if (stock >= 5) {
                        c = 0xFFFFFF;
                        if (inBuyRow && m.mouseX > px + 333 && m.mouseX < px + 345) c = 0xFF0000;
                        m.li.drawstring("5", px + 335, py + 214, c, false, 3);
                    }
                    if (stock >= 10) {
                        c = 0xFFFFFF;
                        if (inBuyRow && m.mouseX > px + 348 && m.mouseX < px + 365) c = 0xFF0000;
                        m.li.drawstring(Mudclient.STRINGS[612], px + 350, py + 214, c, false, 3);
                    }
                    if (stock >= 50) {
                        c = 0xFFFFFF;
                        if (inBuyRow && m.mouseX > px + 368 && m.mouseX < px + 385) c = 0xFF0000;
                        m.li.drawstring(Mudclient.STRINGS[605], px + 370, py + 214, c, false, 3);
                    }
                    c = 0xFFFFFF;
                    if (inBuyRow && m.mouseX > px + 388 && m.mouseX < px + 400) c = 0xFF0000;
                    m.li.drawstring("X", px + 390, py + 214, c, false, 3);
                }
                // sell line
                int held = m.menus.menuHitTest(88, itemId);
                if (held <= 0) {
                    m.li.drawStringCenter(px + 204, Mudclient.STRINGS[632], 0xFFFF00, 0, 3, py + 239); // "shop won't buy"
                } else {
                    int sellPrice = ISAAC.scaledPercentage(InputState.pixelData[itemId], m.vi[m.Di], m.Nh, -30910, false, 1, stock, m.Pf);
                    m.li.drawstring(DecodeBuffer.chatFilterCache[itemId] + Mudclient.STRINGS[638] + sellPrice + Mudclient.STRINGS[636],
                        px + 2, py + 239, 0xFFFF00, false, 1);
                    boolean inSellRow = m.mouseY >= py + 229 && m.mouseY <= py + 240;
                    m.li.drawstring(Mudclient.STRINGS[634], px + 285, py + 239, 0xFFFFFF, false, 3);
                    int c = 0xFFFFFF;
                    if (inSellRow && m.mouseX > px + 318 && m.mouseX < px + 330) c = 0xFF0000;
                    m.li.drawstring("1", px + 320, py + 239, c, false, 3);
                    if (held >= 5) {
                        c = 0xFFFFFF;
                        if (inSellRow && m.mouseX > px + 333 && m.mouseX < px + 345) c = 0xFF0000;
                        m.li.drawstring("5", px + 335, py + 239, c, false, 3);
                    }
                    if (held >= 10) {
                        c = 0xFFFFFF;
                        if (inSellRow && m.mouseX > px + 348 && m.mouseX < px + 365) c = 0xFF0000;
                        m.li.drawstring(Mudclient.STRINGS[612], px + 350, py + 239, c, false, 3);
                    }
                    if (held >= 50) {
                        c = 0xFFFFFF;
                        if (inSellRow && m.mouseX > px + 368 && m.mouseX < px + 385) c = 0xFF0000;
                        m.li.drawstring(Mudclient.STRINGS[605], px + 370, py + 239, c, false, 3);
                    }
                    c = 0xFFFFFF;
                    if (inSellRow && m.mouseX > px + 388 && m.mouseX < px + 400) c = 0xFF0000;
                    m.li.drawstring("X", px + 390, py + 239, c, false, 3);
                }
                return;
            }
        }
        // nothing selected
        m.li.drawStringCenter(px + 204, Mudclient.STRINGS[644], 0xFFFF00, 0, 3, py + 214);
    }

    /** Bank deposit/withdraw panel with up to 4 page tabs. Opcodes: 22 WITHDRAW, 23 DEPOSIT,
     *  212 CLOSE_BANK.  Selected slot is Rd / item sj; page index is xg; vj = items used.
     *
     *  FIX vs old: the withdraw/deposit packet sends live INSIDE the click-handling block
     *  (clean lines 2221-2309), not interleaved with the render of the quantity buttons as the
     *  old version had it. Page-1/2 tabs are only drawn when vj>48 (single page → no tabs).
     *  Withdraw-X uses CacheFile.m (d.m), deposit-X uses RecordLoader.c (f.c). */
    void drawBank(int param) {
        final int PANEL_W = 408;
        final int PANEL_H = 334;

        // clamp page index against item count vj
        if (m.xg < 0 && m.vj <= 48)  m.xg = 0;
        if (m.xg > 1 && m.vj <= 96)  m.xg = 1;
        if (m.vj <= m.Rd || m.Rd < 0) m.Rd = -1;
        if (m.xg > 3 && m.vj <= 144) m.xg = 2;   // (clean: xg<-3 && vj<-145 idiom)
        if (m.Rd != -1 && m.sj != m.ae[m.Rd]) {
            m.Rd = -1;
            m.sj = -2;
        }

        // --- click handling ---
        // Returns early (CLOSE_BANK at the bottom) when the click lands outside both the item
        // grid and the page tabs; clicking a tab switches xg; clicking the grid selects a slot
        // and may fire a withdraw/deposit. (clean source label984 / label929 structure.)
        if (m.gc == 0 && m.Cf != 0) {
            m.Cf = 0;
            int relX = PANEL_W / 2 - 256 + m.mouseX;
            int relY = m.mouseY - (-(PANEL_H / 2) + 170);

            boolean inGrid = !(relX < 0 || relY < 12 || relX >= 408 || relY >= 280);
            boolean closeBank = false;
            if (!inGrid) {
                // page-tab row / close: relY <= 12, columns 50px wide
                if (m.vj > 48 && relX >= 50 && relX <= 115 && relY <= 12) {
                    m.xg = 0;
                } else if (m.vj > 48 && relX >= 115 && relX <= 180 && relY <= 12) {
                    m.xg = 1;
                } else if (m.vj > 96 && relX >= 180 && relX <= 245 && relY <= 12) {
                    m.xg = 2;
                } else if (m.vj > 144 && relX >= 245 && relX <= 311 && relY <= 12) {
                    m.xg = 3;
                } else {
                    closeBank = true;   // any other out-of-grid click closes the bank
                }
            }

            if (closeBank) {
                m.Jh.newPacket(212, 0);   // CLOSE_BANK
                m.Jh.finishPacket(21294);
                m.Fe = false;
                return;
            }

            if (inGrid) {
                // item-grid hit-test (8 rows x 6 cols, current page = xg*48)
                int srcSlot = m.xg * 48;
                for (int row = 0; row < 8; row++) {
                    for (int col = 0; col < 6; col++) {
                        int cellX = 7 + 49 * col;
                        int cellY = 34 * row + 28;
                        if (cellX < relX && cellX + 49 > relX && cellY < relY && cellY + 34 > relY
                                && srcSlot < m.vj && m.ae[srcSlot] != 0) {
                            m.sj = m.ae[srcSlot];
                            m.Rd = srcSlot;
                        }
                        srcSlot++;
                    }
                }

                // --- withdraw / deposit quantity dispatch for the selected slot ---
                // (uses absolute mouseX/mouseY against the restored render offsets px/py)
                int px2 = 256 - PANEL_W / 2;
                int py2 = -(PANEL_H / 2) + 170;
                if (m.Rd != -1) {
                    int selItem = m.Rd >= 0 ? m.ae[m.Rd] : -1;
                    if (selItem != 0 && selItem != -1) {
                        int qty = m.di[m.Rd];
                        // withdraw 1 / 5 / 10 / 50 / X / All
                        if (qty >= 1 && m.mouseX >= px2 + 220 && m.mouseY >= py2 + 238
                                && m.mouseX < px2 + 250 && m.mouseY <= py2 + 249) {
                            m.bankSend(22, selItem, 1, 0x12345678);
                        }
                        if (qty >= 5 && m.mouseX >= px2 + 250 && m.mouseY >= py2 + 238
                                && m.mouseX < px2 + 280 && m.mouseY <= py2 + 249) {
                            m.bankSend(22, selItem, 5, 0x12345678);
                        }
                        if (qty >= 10 && m.mouseX >= px2 + 280 && m.mouseY >= py2 + 238
                                && m.mouseX < px2 + 305 && m.mouseY <= py2 + 249) {
                            m.bankSend(22, selItem, 10, 0x12345678);
                        }
                        if (qty >= 50 && m.mouseX >= px2 + 305 && m.mouseY >= py2 + 238
                                && m.mouseX < px2 + 335 && m.mouseY <= py2 + 249) {
                            m.bankSend(22, selItem, 50, 0x12345678);
                        }
                        if (m.mouseX >= px2 + 335 && m.mouseY >= py2 + 238
                                && m.mouseX < px2 + 368 && m.mouseY <= py2 + 249) {
                            m.drawMenuOptions(CacheFile.sharedStrings, 12, 3, true);   // withdraw X dialog (CacheFile.m)
                        }
                        if (m.mouseX >= px2 + 370 && m.mouseY >= py2 + 238
                                && m.mouseX < px2 + 400 && m.mouseY <= py2 + 249) {
                            m.bankSend(22, selItem, qty, 0x12345678);   // withdraw All
                        }
                        // deposit 1 / 5 / 10 / 50 / X / All  (b(...) = count held in inventory)
                        if (m.menus.menuHitTest(93, selItem) >= 1 && m.mouseX >= px2 + 220 && m.mouseY >= py2 + 263
                                && m.mouseX < px2 + 250 && m.mouseY <= py2 + 274) {
                            m.bankSend(23, selItem, 1, 0x87654321);
                        }
                        if (m.menus.menuHitTest(90, selItem) >= 5 && m.mouseX >= px2 + 250 && m.mouseY >= py2 + 263
                                && m.mouseX < px2 + 280 && m.mouseY <= py2 + 274) {
                            m.bankSend(23, selItem, 5, 0x87654321);
                        }
                        if (m.menus.menuHitTest(108, selItem) >= 10 && m.mouseX >= px2 + 280 && m.mouseY >= py2 + 263
                                && m.mouseX < px2 + 305 && m.mouseY <= py2 + 274) {
                            m.bankSend(23, selItem, 10, 0x87654321);
                        }
                        if (m.menus.menuHitTest(109, selItem) >= 50 && m.mouseX >= px2 + 305 && m.mouseY >= py2 + 263
                                && m.mouseX < px2 + 335 && m.mouseY <= py2 + 274) {
                            m.bankSend(23, selItem, 50, 0x87654321);
                        }
                        if (m.mouseX >= px2 + 335 && m.mouseY >= py2 + 263
                                && m.mouseX < px2 + 368 && m.mouseY <= py2 + 274) {
                            m.drawMenuOptions(RecordLoader.DEPOSIT_PROMPT_STRINGS, 12, 4, true);   // deposit X dialog (RecordLoader.c)
                        }
                        if (m.mouseX >= px2 + 370 && m.mouseY >= py2 + 263
                                && m.mouseX < px2 + 400 && m.mouseY <= py2 + 274) {
                            m.bankSend(23, selItem, m.menus.menuHitTest(85, selItem), 0x87654321);   // deposit All
                        }
                    }
                }
            }
        }

        // --- render panel ---
        int px = 256 - PANEL_W / 2;
        int py = 170 - PANEL_H / 2;
        m.li.drawBox(px, (byte) -126, 192, py, 12, 408);
        int grey = 0x989898;
        m.li.drawBoxAlpha(160, px, 17, 0, py + 12, 408, grey);
        m.li.drawBoxAlpha(160, px, 204, 0, py + 29, 8, grey);
        m.li.drawBoxAlpha(160, px + 399, 204, 0, py + 29, 9, grey);
        m.li.drawBoxAlpha(160, px, 47, 0, py + 233, 408, grey);
        m.li.drawstring(Mudclient.STRINGS[610], px + 1, py + 10, 0xFFFFFF, false, 1);   // "Bank"

        int tabX = 50;
        if (m.vj > 48) {
            // page-1 tab
            int col = 0xFFFFFF;
            if (m.xg == 0) col = 0xFF0000;
            if (px + tabX < m.mouseX && m.mouseY >= py && px + tabX + 65 > m.mouseX && m.mouseY < py + 12) {
                col = 0xFFFF00;
            }
            m.li.drawstring(Mudclient.STRINGS[607], px + tabX, py + 10, col, false, 1);
            tabX += 65;
            // page-2 tab
            col = 0xFFFFFF;
            if (m.xg == 1) col = 0xFF0000;
            if (px + tabX < m.mouseX && m.mouseY >= py && px + tabX + 65 > m.mouseX && m.mouseY < py + 12) {
                col = 0xFFFF00;
            }
            m.li.drawstring(Mudclient.STRINGS[618], px + tabX, py + 10, col, false, 1);
            tabX += 65;
        }
        if (m.vj > 96) {
            int col = 0xFFFFFF;
            if (m.xg == 2) {
                col = 0xFF0000;
            } else if (px + tabX < m.mouseX && m.mouseY >= py && px + tabX + 65 > m.mouseX && m.mouseY < py + 12) {
                col = 0xFFFF00;
            }
            m.li.drawstring(Mudclient.STRINGS[616], px + tabX, py + 10, col, false, 1);
            tabX += 65;
        }
        if (m.vj > 144) {
            int col = 0xFFFFFF;
            if (m.xg == 3) col = 0xFF0000;
            if (px + tabX < m.mouseX && m.mouseY >= py && px + tabX + 65 > m.mouseX && m.mouseY > py + 12) {
                col = 0xFFFF00;
            }
            m.li.drawstring(Mudclient.STRINGS[621], px + tabX, py + 10, col, false, 1);
            tabX += 65;
        }

        int closeCol = 0xFFFFFF;
        if (m.mouseX > px + 320 && m.mouseY >= py && m.mouseX < px + 408 && m.mouseY < py + 12) {
            closeCol = 0xFF0000;
        }
        m.li.drawstringRightSimple(px + 406, Mudclient.STRINGS[620], py + 10, closeCol, -69, 1);
        m.li.drawstring(Mudclient.STRINGS[608], px + 7, py + 24, 0x00FF00, false, 1);    // "Withdraw"
        m.li.drawstring(Mudclient.STRINGS[606], px + 289, py + 24, 0x00FFFF, false, 1);  // "Deposit"

        // item grid 8x6
        int grey2 = 0xD0D0D0;
        int srcSlot = m.xg * 48;
        for (int row = 0; row < 8; row++) {
            for (int col = 0; col < 6; col++) {
                int cellX = col * 49 + px + 7;
                int cellY = row * 34 + py + 28;
                if (srcSlot == m.Rd) {
                    m.li.drawBoxAlpha(160, cellX, 34, 0, cellY, 49, 0xFF0000);
                } else {
                    m.li.drawBoxAlpha(160, cellX, 34, 0, cellY, 49, grey2);
                }
                m.li.drawBoxEdge(cellX, 50, cellY, 27785, 35, 0);
                if (srcSlot < m.vj && m.ae[srcSlot] != 0) {
                    m.li.spriteClipping(cellY, TextEncoder.scratchIntArray2[m.ae[srcSlot]], 0, false, 0,
                        Surface.unusedIntsBb[m.ae[srcSlot]] + m.sg, 32, 48, cellX, 1);
                    m.li.drawstring("" + m.di[srcSlot], cellX + 1, cellY + 10, 0x00FF00, false, 1);
                    m.li.drawstringRightSimple(cellX + 47, "" + m.menus.menuHitTest(87, m.ae[srcSlot]), cellY + 29, 0x00FFFF, 127, 1);
                }
                srcSlot++;
            }
        }
        m.li.drawLineHoriz(398, 0, px + 5, py + 256, (byte) -87);   // scrollbar

        // selected-slot quantity rows
        if (m.Rd != -1) {
            int selItem = m.Rd >= 0 ? m.ae[m.Rd] : -1;
            if (selItem != 0) {
                int qty = m.di[m.Rd];
                if (ClientIOException.itemY[selItem] == 1 && qty > 1) qty = 1;   // non-stackable cap (~e[]==-2 idiom)
                if (qty > 0) {
                    // "Withdraw <item>" + 1/5/10/50/X/All buttons
                    int c = 0xFFFFFF;
                    m.li.drawstring(Mudclient.STRINGS[611] + DecodeBuffer.chatFilterCache[selItem], px + 2, py + 248, 0xFFFFFF, false, 1);
                    if (m.mouseX >= px + 220 && m.mouseY >= py + 238 && m.mouseX < px + 250 && m.mouseY <= py + 249) c = 0xFF0000;
                    m.li.drawstring(Mudclient.STRINGS[617], px + 222, py + 248, c, false, 1);     // "1"
                    if (qty >= 5) {
                        c = 0xFFFFFF;
                        if (m.mouseX >= px + 250 && m.mouseY >= py + 238 && m.mouseX < px + 280 && m.mouseY <= py + 249) c = 0xFF0000;
                        m.li.drawstring(Mudclient.STRINGS[619], px + 252, py + 248, c, false, 1); // "5"
                    }
                    if (qty >= 10) {
                        c = 0xFFFFFF;
                        if (m.mouseX >= px + 280 && m.mouseY >= py + 238 && m.mouseX < px + 305 && m.mouseY <= py + 249) c = 0xFF0000;
                        m.li.drawstring(Mudclient.STRINGS[612], px + 282, py + 248, c, false, 1); // "10"
                    }
                    if (qty >= 50) {
                        c = 0xFFFFFF;
                        if (m.mouseX >= px + 305 && m.mouseY >= py + 238 && m.mouseX < px + 335 && m.mouseY <= py + 249) c = 0xFF0000;
                        m.li.drawstring(Mudclient.STRINGS[605], px + 307, py + 248, c, false, 1); // "50"
                    }
                    c = 0xFFFFFF;
                    if (m.mouseX >= px + 335 && m.mouseY >= py + 238 && m.mouseX < px + 368 && m.mouseY <= py + 249) c = 0xFF0000;
                    m.li.drawstring("X", px + 337, py + 248, c, false, 1);
                    c = 0xFFFFFF;
                    if (m.mouseX >= px + 370 && m.mouseY >= py + 238 && m.mouseX < px + 400 && m.mouseY <= py + 249) c = 0xFF0000;
                    m.li.drawstring(Mudclient.STRINGS[615], px + 370, py + 248, c, false, 1);     // "All"
                }
                // "Deposit <item>" + 1/5/10/50/X/All buttons (only if the player owns any)
                if (m.menus.menuHitTest(126, selItem) > 0) {
                    m.li.drawstring(Mudclient.STRINGS[614] + DecodeBuffer.chatFilterCache[selItem], px + 2, py + 273, 0xFFFFFF, false, 1);
                    int c = 0xFFFFFF;
                    if (m.mouseX >= px + 220 && m.mouseY >= py + 263 && m.mouseX < px + 250 && m.mouseY <= py + 274) c = 0xFF0000;
                    m.li.drawstring(Mudclient.STRINGS[617], px + 222, py + 273, c, false, 1);
                    if (m.menus.menuHitTest(88, selItem) >= 5) {
                        c = 0xFFFFFF;
                        if (m.mouseX >= px + 250 && m.mouseY >= py + 263 && m.mouseX < px + 280 && m.mouseY <= py + 274) c = 0xFF0000;
                        m.li.drawstring(Mudclient.STRINGS[619], px + 252, py + 273, c, false, 1);
                    }
                    if (m.menus.menuHitTest(93, selItem) >= 10) {
                        c = 0xFFFFFF;
                        if (m.mouseX >= px + 280 && m.mouseY >= py + 263 && m.mouseX < px + 305 && m.mouseY <= py + 274) c = 0xFF0000;
                        m.li.drawstring(Mudclient.STRINGS[612], px + 282, py + 273, c, false, 1);
                    }
                    if (m.menus.menuHitTest(98, selItem) >= 50) {
                        c = 0xFFFFFF;
                        if (m.mouseX >= px + 305 && m.mouseY >= py + 263 && m.mouseX < px + 335 && m.mouseY <= py + 274) c = 0xFF0000;
                        m.li.drawstring(Mudclient.STRINGS[605], px + 307, py + 273, c, false, 1);
                    }
                    c = 0xFFFFFF;
                    if (m.mouseX >= px + 335 && m.mouseY >= py + 263 && m.mouseX < px + 368 && m.mouseY <= py + 274) c = 0xFF0000;
                    m.li.drawstring("X", px + 337, py + 273, c, false, 1);
                    c = 0xFFFFFF;
                    if (m.mouseX >= px + 370 && m.mouseY >= py + 263 && m.mouseX < px + 400 && m.mouseY <= py + 274) c = 0xFF0000;
                    m.li.drawstring(Mudclient.STRINGS[615], px + 370, py + 273, c, false, 1);
                }
                return;
            }
        }
        m.li.drawStringCenter(px + 204, Mudclient.STRINGS[613], 0xFFFF00, 0, 3, py + 248);   // "Select an item"
    }

    /** Trade offer window: your inventory (left, lc items in vf/xe drawn from px+217), their
     *  current offer (Qf/jj, mf items) and your committed offer (zj/Dd, Lk items). Handles a
     *  right-click "offer hasPainted" sub-menu via the ignoreList MessageList (Wf). Opcodes: 55
     *  ACCEPT_TRADE, 230 DECLINE_TRADE; offers go through sendTradeOffer/sendDuelOffer.
     *
     *  FIX vs old: old version had only a stub render with the wrong arrays (zc/of/wj are the
     *  DUEL buffers) and omitted the Cf==2 right-click menu builder and the third (zj/Dd) grid.
     *  Rewritten in full from the clean source. */
    void drawTrade(byte param) {
        int menuPick = -1;
        if (m.Cf != 0 && m.lh) {
            menuPick = m.ignoreList.hitTestNoRender(m.mouseX, m.Gf, m.Bf, (byte) -40, m.mouseY);
        }

        if (menuPick < 0) {
            if (m.gc == 0) {
                if (m.Cf == 1 && m.Tk == 0) m.Tk = 1;
                int relX = m.mouseX - 22;
                int relY = m.mouseY - 36;
                boolean inPanel = !(relX < 0 || relY < 0 || relX >= 469 || relY >= 262);
                if (!inPanel) {
                    if (m.Cf == 1) {          // click outside → decline
                        m.Hk = false;
                        m.Jh.newPacket(230, 0);
                        m.Jh.finishPacket(21294);
                    }
                } else {
                    // --- left mouse: remove an offered item / accept / decline ---
                    if (m.Tk > 0) {
                        // your-offer grid (217..462 x, 31..235 y, 5 cols) → remove
                        if (relX > 216 && relY > 30 && relX < 462 && relY < 235) {
                            int slot = 5 * ((relY - 31) / 34) + (relX - 217) / 49;
                            if (slot >= 0 && slot < m.lc) {
                                this.drawTradeConfirm(-1, (byte) 9, slot);  // a(int,byte,int): remove 1
                            }
                        }
                        // their-offer grid (8..205 x, 31..133 y, 4 cols) → remove
                        if (relX > 8 && relY > 30 && relX < 205 && relY < 133) {
                            int slot = (relY - 31) / 34 * 4 + (relX - 9) / 49;
                            if (slot >= 0 && slot < m.mf) {
                                m.sendTradeOffer(-1, (byte) 125, slot); // c(int,byte,int)
                            }
                        }
                        // accept button (217..286 x, 238..259 y)
                        if (relX >= 217 && relY >= 238 && relX <= 286 && relY <= 259) {
                            m.Mi = true;
                            m.Jh.newPacket(55, 0);
                            m.Jh.finishPacket(21294);
                        }
                        // decline button (394..462 x, 238..258 y)
                        if (relX >= 394 && relY >= 238 && relX < 463 && relY < 259) {
                            m.Hk = false;
                            m.Jh.newPacket(230, param - 8);
                            m.Jh.finishPacket(21294);
                        }
                        m.Tk = 0;
                        m.Cf = 0;
                    }

                    // --- right mouse (Cf==2): open an "offer 1/5/10/all/cancel" sub-menu ---
                    if (m.Cf == 2) {
                        // over your-offer grid → menu for an inventory item
                        if (relX > 216 && relY > 30 && relX < 462 && relY < 235) {
                            int w = m.zh.getPanelWidth(16256);
                            int hgt = m.zh.getPanelHeight(-21224);
                            m.fg = m.mouseY - 7;
                            m.rh = m.mouseX - w / 2;
                            m.se = true;
                            if (m.fg < 0) m.fg = 0;
                            if (m.rh < 0) m.rh = 0;
                            if (m.rh + w > 510) m.rh = 510 - w;
                            if (m.fg + hgt > 316) m.fg = 315 - hgt;

                            int slot = (relY - 31) / 34 * 5 + (relX - 217) / 49;
                            if (slot >= 0 && slot < m.lc) {
                                int itemId = m.vf[slot];
                                m.lh = true;
                                m.ignoreList.setCount(0);
                                m.ignoreList.addEntryWithColor(itemId, Mudclient.STRINGS[34] + DecodeBuffer.chatFilterCache[itemId], 1, Mudclient.STRINGS[172], 1,  param + 3288);
                                m.ignoreList.addEntryWithColor(itemId, Mudclient.STRINGS[34] + DecodeBuffer.chatFilterCache[itemId], 1, Mudclient.STRINGS[169], 5,  3296);
                                m.ignoreList.addEntryWithColor(itemId, Mudclient.STRINGS[34] + DecodeBuffer.chatFilterCache[itemId], 1, Mudclient.STRINGS[158], 10, 3296);
                                m.ignoreList.addEntryWithColor(itemId, Mudclient.STRINGS[34] + DecodeBuffer.chatFilterCache[itemId], 1, Mudclient.STRINGS[174], -1, 3296);
                                m.ignoreList.addEntryWithColor(itemId, Mudclient.STRINGS[34] + DecodeBuffer.chatFilterCache[itemId], 1, Mudclient.STRINGS[166], -2, param ^ 3304);
                                int mw = m.ignoreList.getPanelWidth(param ^ 16264);
                                int mh = m.ignoreList.getPanelHeight(-21224);
                                m.Gf = m.mouseX - mw / 2;
                                m.Bf = m.mouseY - 7;
                                if (m.Gf < 0) m.Gf = 0;
                                if (m.Bf < 0) m.Bf = 0;
                                if (m.Bf + mh > 316) m.Bf = 315 - mh;
                                if (m.Gf + mw > 511) m.Gf = 510 - mw;
                            }
                        }
                        // over their-offer grid → menu for a removable offered item
                        if (relX > 8 && relY > 30 && relX < 205 && relY < 133) {
                            int slot = (relX - 9) / 49 + (relY - 31) / 34 * 4;
                            if (slot >= 0 && slot < m.mf) {
                                int itemId = m.Qf[slot];
                                m.lh = true;
                                m.ignoreList.setCount(0);
                                m.ignoreList.addEntryWithColor(itemId, Mudclient.STRINGS[34] + DecodeBuffer.chatFilterCache[itemId], 2, Mudclient.STRINGS[163], 1,  3296);
                                m.ignoreList.addEntryWithColor(itemId, Mudclient.STRINGS[34] + DecodeBuffer.chatFilterCache[itemId], 2, Mudclient.STRINGS[173], 5,  param ^ 3304);
                                m.ignoreList.addEntryWithColor(itemId, Mudclient.STRINGS[34] + DecodeBuffer.chatFilterCache[itemId], 2, Mudclient.STRINGS[161], 10, 3296);
                                m.ignoreList.addEntryWithColor(itemId, Mudclient.STRINGS[34] + DecodeBuffer.chatFilterCache[itemId], 2, Mudclient.STRINGS[177], -1, 3296);
                                m.ignoreList.addEntryWithColor(itemId, Mudclient.STRINGS[34] + DecodeBuffer.chatFilterCache[itemId], 2, Mudclient.STRINGS[170], -2, param ^ 3304);
                                int mw = m.ignoreList.getPanelWidth(16256);
                                int mh = m.ignoreList.getPanelHeight(-21224);
                                m.Gf = m.mouseX - mw / 2;
                                m.Bf = m.mouseY - 7;
                                if (m.Gf < 0) m.Gf = 0;
                                if (m.Bf < 0) m.Bf = 0;
                                if (mh + m.Bf > 315) m.Bf = 315 - mh;
                                if (mw + m.Gf > 511) m.Gf = 510 - mw;
                            }
                        }
                        m.Cf = 0;
                    }

                    // dismiss the sub-menu when the cursor leaves its bounds
                    if (m.lh) {
                        int mw = m.ignoreList.getPanelWidth(16256);
                        int mh = m.ignoreList.getPanelHeight(-21224);
                        if (m.mouseX < m.Gf - 10 || m.mouseX > m.Gf + mw + 10
                                || m.mouseY < m.Bf - 10 || m.mouseY > m.Bf + mh + 10) {
                            m.lh = false;
                        }
                    }
                }
            }
        } else {
            // --- a sub-menu entry was clicked: resolve it to an offer ---
            m.lh = false;
            m.Cf = 0;
            int action = m.ignoreList.getEntryXPos(-91, menuPick);   // 1 = inventory item, else offered item
            int itemId = m.ignoreList.getEntryColorE(true, menuPick);
            int slot = -1;
            int total = 0;
            if (action == 1) {
                for (int i = 0; i < m.lc; i++) {
                    if (m.vf[i] == itemId) {
                        if (slot < 0) slot = i;
                        if (ClientIOException.itemY[itemId] == 0) { total = m.xe[i]; break; }
                        total++;
                    }
                }
            } else {
                for (int i = 0; i < m.mf; i++) {
                    if (m.Qf[i] == itemId) {
                        if (slot < 0) slot = i;
                        if (ClientIOException.itemY[itemId] == 0) { total = m.jj[i]; break; }
                        total++;
                    }
                }
            }
            if (slot >= 0) {
                int amount = m.ignoreList.getEntryColorCode((byte) 97, menuPick);
                if (amount == -2) {
                    m.ji = slot;                                    // "X" → open qty entry
                    if (action == 1) {
                        m.drawMenuOptions(FontBuilder.injectedStrings, 12, 1, true);
                    } else {
                        m.drawMenuOptions(Surface.decoyString1, param ^ 4, 2, true);
                    }
                } else {
                    if (amount == 0) amount = total;                   // "All"
                    if (action == 1) {
                        this.drawTradeConfirm(amount, (byte) 9, slot); // add to your offer
                    } else {
                        m.sendTradeOffer(amount, (byte) 124, slot); // remove from your offer
                    }
                }
            }
        }

        // --- draw panel ---
        if (m.Hk) {
            final int px = 22, py = 36;
            m.li.drawBox(px, (byte) 117, 192, py, 12, 468);
            int grey = 0x989898;
            m.li.drawBoxAlpha(160, px, 18, param - 8, py + 12, 468, grey);
            m.li.drawBoxAlpha(160, px, 248, 0, py + 30, 8, grey);
            m.li.drawBoxAlpha(160, px + 205, 248, param - 8, py + 30, 11, grey);
            m.li.drawBoxAlpha(160, px + 462, 248, param - 8, py + 30, 6, grey);
            m.li.drawBoxAlpha(160, px + 8, 22, 0, py + 133, 197, grey);
            m.li.drawBoxAlpha(160, px + 8, 20, 0, py + 258, 197, grey);
            m.li.drawBoxAlpha(160, px + 216, 43, 0, py + 235, 246, grey);
            int lgrey = 0xD0D0D0;
            m.li.drawBoxAlpha(160, px + 8, 103, param - 8, py + 30, 197, lgrey);
            m.li.drawBoxAlpha(160, px + 8, 103, 0, py + 155, 197, lgrey);
            m.li.drawBoxAlpha(160, px + 216, 205, param - 8, py + 30, 246, lgrey);

            for (int r = 0; r < 4; r++) m.li.drawLineHoriz(197, 0, px + 8, py + 30 + 34 * r, (byte) -98);
            for (int r = 0; r < 4; r++) m.li.drawLineHoriz(197, 0, px + 8, 34 * r + 155 + py, (byte) -29);
            for (int r = 0; r < 7; r++) m.li.drawLineHoriz(246, 0, px + 216, py + 30 + r * 34, (byte) 60);
            for (int c = 0; c < 6; c++) {
                m.li.drawLineVert(px + 8 + c * 49, py + 30, 0, 103, 0);
                m.li.drawLineVert(c * 49 + 8 + px, py + 155, 0, 103, param ^ 8);
                m.li.drawLineVert(px + 216 + c * 49, py + 30, 0, 205, 0);
            }

            m.li.drawstring(Mudclient.STRINGS[175] + m.cj, px + 1, py + 10, 0xFFFFFF, false, 1);  // "Trade with <name>"
            m.li.drawstring(Mudclient.STRINGS[164], px + 9, py + 27, 0xFFFFFF, false, 4);            // "Your offer"
            m.li.drawstring(Mudclient.STRINGS[167], px + 9, py + 152, 0xFFFFFF, false, 4);           // "Opponent's offer"
            m.li.drawstring(Mudclient.STRINGS[171], px + 216, py + 27, 0xFFFFFF, false, 4);          // "Options"
            if (!m.Mi) {
                m.li.drawSprite(-1, m.tg + 25, py + 238, px + 217);   // accept button sprite
            }
            m.li.drawSprite(-1, m.tg + 26, py + 238, px + 394);       // decline button sprite
            if (m.md) {
                m.li.drawStringCenter(px + 341, Mudclient.STRINGS[168], 0xFFFFFF, param ^ 8, 1, py + 246);
                m.li.drawStringCenter(px + 341, Mudclient.STRINGS[165], 0xFFFFFF, 0, 1, py + 256);
            }
            if (m.Mi) {
                m.li.drawStringCenter(px + 217 + 35, Mudclient.STRINGS[176], 0xFFFFFF, param - 8, 1, py + 246);
                m.li.drawStringCenter(px + 252, Mudclient.STRINGS[160], 0xFFFFFF, param - 8, 1, py + 256);
            }

            // your inventory grid (vf/xe, 5 cols, starts at px+217)
            for (int i = 0; i < m.lc; i++) {
                int cellX = px + 217 + 49 * (i % 5);
                int cellY = py + 31 + i / 5 * 34;
                m.li.spriteClipping(cellY, TextEncoder.scratchIntArray2[m.vf[i]], 0, false, 0,
                    m.sg + Surface.unusedIntsBb[m.vf[i]], 32, 48, cellX, 1);
                if (ClientIOException.itemY[m.vf[i]] == 0) {
                    m.li.drawstring("" + m.xe[i], cellX + 1, cellY + 10, 0xFFFF00, false, 1);
                }
            }
            // their current offer (Qf/jj, 4 cols, starts at px+9)
            for (int i = 0; i < m.mf; i++) {
                int cellX = i % 4 * 49 + 9 + px;
                int cellY = i / 4 * 34 + py + 31;
                m.li.spriteClipping(cellY, TextEncoder.scratchIntArray2[m.Qf[i]], 0, false, 0,
                    Surface.unusedIntsBb[m.Qf[i]] + m.sg, 32, 48, cellX, 1);
                if (ClientIOException.itemY[m.Qf[i]] == 0) {
                    m.li.drawstring("" + m.jj[i], cellX + 1, cellY + 10, 0xFFFF00, false, 1);
                }
                if (m.mouseX > cellX && cellX + 48 > m.mouseX && cellY < m.mouseY && m.mouseY < cellY + 32) {
                    m.li.drawstring(DecodeBuffer.chatFilterCache[m.Qf[i]] + Mudclient.STRINGS[159] + CharTable.itemDescriptions[m.Qf[i]],
                        px + 8, py + 273, 0xFFFF00, false, 1);   // examine tooltip
                }
            }
            // your committed offer (zj/Dd, 4 cols, starts at px+9, second block at py+156)
            for (int i = 0; i < m.Lk; i++) {
                int cellX = px + 9 + i % 4 * 49;
                int cellY = py + 156 + 34 * (i / 4);
                m.li.spriteClipping(cellY, TextEncoder.scratchIntArray2[m.zj[i]], 0, false, 0,
                    Surface.unusedIntsBb[m.zj[i]] + m.sg, 32, 48, cellX, 1);
                if (ClientIOException.itemY[m.zj[i]] == 0) {
                    m.li.drawstring("" + m.Dd[i], cellX + 1, cellY + 10, 0xFFFF00, false, 1);
                }
                if (m.mouseX > cellX && cellX + 48 > m.mouseX && m.mouseY > cellY && m.mouseY < cellY + 32) {
                    m.li.drawstring(DecodeBuffer.chatFilterCache[m.zj[i]] + Mudclient.STRINGS[159] + CharTable.itemDescriptions[m.zj[i]],
                        px + 8, py + 273, 0xFFFF00, false, 1);
                }
            }

            if (m.lh) {
                m.ignoreList.hitTest(m.Bf, m.Gf, m.mouseY, (byte) -12, m.mouseX);   // render sub-menu
            }
        }
    }

    /** Add/adjust items in the local trade-offer buffer (Qf/jj) from inventory slot invSlot,
     *  then push the whole offer with opcode 46 (PLAYER_ADDED_ITEMS_TO_TRADE_OFFER).
     *  `count` is the requested amount (-1 = single decrement via Tk loop).
     *
     *  FIX vs old: stackable cap test is kb.c[itemId] (InputState.c), not "EntityDef.c";
     *  stackable flag is fa.e[itemId] (ClientIOException.e). */
    void drawTradeConfirm(int count, byte action, int invSlot) {
        if (action != 9) {
            m.resetPanels((byte) -38);   // p(byte)
        }
        boolean changed = false;
        int dupes = 0;
        int itemId = m.vf[invSlot];

        int offerCount;
        for (int i = 0; ; i++) {
            if (i < m.mf) {
                if (itemId == m.Qf[i]) {
                    if (ClientIOException.itemY[itemId] == 0) {     // stackable
                        if (count >= 0) {
                            m.jj[i] += count;
                            if (m.jj[i] > m.xe[invSlot]) m.jj[i] = m.xe[invSlot];
                            changed = true;
                        } else {
                            for (int k = 0; k < m.Tk; k++) {
                                changed = true;
                                if (m.jj[i] < m.xe[invSlot]) m.jj[i]++;
                            }
                        }
                    }
                    dupes++;
                }
                continue;
            }
            offerCount = m.menus.menuHitTest(99, itemId);   // current copies already offered
            break;
        }
        if (offerCount <= dupes) changed = true;
        if (InputState.slotFlags[itemId] == 1) {               // FIX: kb.c (InputState), not EntityDef.c
            changed = true;
            m.showServerMessage(false, null, action ^ 9, Mudclient.STRINGS[215], 0, 0, null, null);
        }

        if (!changed) {
            if (count < 0) {
                if (m.mf < 12) {
                    m.Qf[m.mf] = itemId;
                    m.jj[m.mf] = 1;
                    changed = true;
                    m.mf++;
                }
            } else {
                for (int k = 0; k < count; k++) {
                    if (m.mf >= 12 || offerCount <= dupes) break;
                    m.Qf[m.mf] = itemId;
                    m.jj[m.mf] = 1;
                    changed = true;
                    dupes++;
                    m.mf++;
                    if (k == 0 && ClientIOException.itemY[itemId] == 0) {   // first add of a stackable → take min(count,have)
                        m.jj[m.mf - 1] = count <= m.xe[invSlot] ? count : m.xe[invSlot];
                        break;
                    }
                }
            }
        }
        if (!changed) return;

        m.Jh.newPacket(46, 0);
        m.Jh.outBuffer.putByte(m.mf);
        for (int k = 0; k < m.mf; k++) {
            m.Jh.outBuffer.putShort(m.Qf[k]);
            m.Jh.outBuffer.putInt(m.jj[k]);
        }
        m.Jh.finishPacket(21294);
        m.md = false;
        m.Mi = false;
    }

    /** "Please confirm your trade" window: your final items (Vb/Me, count Ui) and theirs
     *  (Lc/Bi, count nh), with Accept/Decline. Vi = you have accepted. Opcodes: 104
     *  CONFIRM_TRADE, 230 DECLINE_TRADE.
     *
     *  FIX vs old: the previous part file's "drawTradeConfirmWindow" was tied to the wrong obf
     *  method (a(boolean,boolean), the social-list panel) and only a stub. This is the real
     *  hasPainted(int) body (clean line 13749). drawActiveInterface dispatches it via the Xj flag. */
    void drawTradeConfirmWindow(int param) {
        final int px = 22, py = 36;
        m.li.drawBox(px, (byte) -117, 192, py, 16, 468);
        int grey = 0x989898;
        m.li.drawBoxAlpha(160, px, 246, 0, py + 16, 468, grey);
        m.li.drawStringCenter(px + 234, Mudclient.STRINGS[204] + m.re, 0xFFFFFF, 0, 1, py + 12);  // "Trade with <name>"
        m.li.drawStringCenter(px + 117, Mudclient.STRINGS[210], 0xFFFF00, 0, 1, py + 30);            // "You are about to give:"

        // your final offer (Vb ids / Me counts, Ui of them)
        for (int i = 0; i < m.Ui; i++) {
            String name = DecodeBuffer.chatFilterCache[m.Vb[i]];
            if (ClientIOException.itemY[m.Vb[i]] == 0) {     // stackable → append count
                name = name + Mudclient.STRINGS[211] + Utility.formatNumber(m.Me[i], 131071);
            }
            m.li.drawStringCenter(px + 117, name, 0xFFFFFF, 0, 1, i * 12 + 42 + py);
        }
        if (m.Ui == 0) {
            m.li.drawStringCenter(px + 117, Mudclient.STRINGS[213], 0xFFFFFF, 0, 1, py + 42);
        }

        m.li.drawStringCenter(px + 351, Mudclient.STRINGS[209], 0xFFFF00, 0, 1, py + 30);            // "In return you will receive:"
        for (int i = 0; i < m.nh; i++) {
            String name = DecodeBuffer.chatFilterCache[m.Lc[i]];
            if (ClientIOException.itemY[m.Lc[i]] == 0) {
                name = name + Mudclient.STRINGS[211] + Utility.formatNumber(m.Bi[i], 131071);
            }
            m.li.drawStringCenter(px + 351, name, 0xFFFFFF, 0, 1, 42 + py + 12 * i);
        }
        if (m.nh == 0) {
            m.li.drawStringCenter(px + 351, Mudclient.STRINGS[213], 0xFFFFFF, 0, 1, py + 42);
        }
        if (param >= -6) return;   // anti-tamper guard (clean: var10000>=var10001 → b(true))

        m.li.drawStringCenter(px + 234, Mudclient.STRINGS[206], 0x00FFFF, 0, 4, py + 200);   // confirm hint lines
        m.li.drawStringCenter(px + 234, Mudclient.STRINGS[207], 0xFFFFFF, 0, 1, py + 215);
        m.li.drawStringCenter(px + 234, Mudclient.STRINGS[205], 0xFFFFFF, 0, 1, py + 230);
        if (m.Vi) {
            m.li.drawStringCenter(px + 234, Mudclient.STRINGS[212], 0xFFFF00, 0, 1, py + 250); // "Waiting..."
        } else {
            m.li.drawSprite(-1, m.tg + 25, py + 238, px + 118 - 35);        // accept sprite
            m.li.drawSprite(-1, m.tg + 26, py + 238, px + 352 - 35);        // decline sprite
        }

        if (m.Cf == 1) {   // clean hasPainted(int) L13845: ~m.Cf==-2 == Cf==1 (left-click Accept/Decline gate)
            // click outside the panel → decline
            if (m.mouseX < px || m.mouseY < py || m.mouseX > px + 468 || m.mouseY > py + 262) {
                m.Xj = false;
                m.Jh.newPacket(230, 0);
                m.Jh.finishPacket(21294);
            }
            // accept button: x 83..188, y 238..259
            if (m.mouseX >= px + 118 - 35 && m.mouseX <= px + 118 + 70 && m.mouseY >= py + 238 && m.mouseY <= py + 259) {
                m.Vi = true;
                m.Jh.newPacket(104, 0);   // CONFIRM_TRADE
                m.Jh.finishPacket(21294);
            }
            // decline button: x 317..423, y 238..259
            if (m.mouseX >= px + 352 - 35 && m.mouseX <= px + 423 && m.mouseY >= py + 238 && m.mouseY <= py + 259) {
                m.Xj = false;
                m.Jh.newPacket(230, 0);
                m.Jh.finishPacket(21294);
            }
            m.Cf = 0;
        }
    }

    /** Duel confirm window: your stake (Nj items in xi/th) vs theirs (Ve items in xj/kf),
     *  the four rule flags, and accept/decline. Opcodes: 77 ACCEPT_DUEL, 197 CONFIRM_DUEL,
     *  230 DECLINE_DUEL.
     *
     *  FIX vs old: the two accept/decline hit-test rectangles had wrong right edges
     *  (old used `+70-35`; clean is left `83..188`, right `317..423`). */
    void drawDuelConfirm(int param) {
        final int px = 22, py = 36;
        m.li.drawBox(px, (byte) -108, 192, py, 16, 468);
        int grey = 0x989898;
        m.li.drawBoxAlpha(160, px, 246, 0, py + 16, 468, grey);
        m.li.drawStringCenter(px + 234, Mudclient.STRINGS[522] + m.Uc, 0xFFFFFF, 0, 1, py + 12);  // "Duel with <name>"
        m.li.drawStringCenter(px + 117, Mudclient.STRINGS[524], 0xFFFF00, 0, 1, py + 30);            // "Your stake:"

        for (int i = 0; i < m.Nj; i++) {
            String name = DecodeBuffer.chatFilterCache[m.xi[i]];
            if (ClientIOException.itemY[m.xi[i]] == 0) {     // stackable → append count
                name = name + Mudclient.STRINGS[211] + Utility.formatNumber(m.th[i], 131071);
            }
            m.li.drawStringCenter(px + 117, name, 0xFFFFFF, 0, 1, 42 + py + 12 * i);
        }
        if (param > -10) return;   // anti-tamper guard (clean: var10000<=var10001 path)

        if (m.Nj == 0) {
            m.li.drawStringCenter(px + 117, Mudclient.STRINGS[213], 0xFFFFFF, 0, 1, 42 + py);
        }
        m.li.drawStringCenter(px + 351, Mudclient.STRINGS[527], 0xFFFF00, 0, 1, py + 30);            // "Their stake:"
        for (int i = 0; i < m.Ve; i++) {
            String name = DecodeBuffer.chatFilterCache[m.xj[i]];
            if (ClientIOException.itemY[m.xj[i]] == 0) {
                name = name + Mudclient.STRINGS[211] + Utility.formatNumber(m.kf[i], 131071);
            }
            m.li.drawStringCenter(px + 351, name, 0xFFFFFF, 0, 1, i * 12 + 42 + py);
        }
        if (m.Ve == 0) {
            m.li.drawStringCenter(px + 351, Mudclient.STRINGS[213], 0xFFFFFF, 0, 1, 42 + py);
        }

        // rule flags (Sh retreat / gh magic / Cc prayer / Rc weapons)
        if (m.Sh == 0) {
            m.li.drawStringCenter(px + 234, Mudclient.STRINGS[528], 0x00FF00, 0, 1, py + 180);   // "Retreat allowed"
        } else {
            m.li.drawStringCenter(px + 234, Mudclient.STRINGS[517], 0xFF0000, 0, 1, py + 180);   // "No retreat"
        }
        if (m.gh == 0) {
            m.li.drawStringCenter(px + 234, Mudclient.STRINGS[526], 0x00FF00, 0, 1, py + 192);   // "Magic allowed"
        } else {
            m.li.drawStringCenter(px + 234, Mudclient.STRINGS[519], 0xFF0000, 0, 1, py + 192);   // "No magic"
        }
        if (m.Cc == 0) {
            m.li.drawStringCenter(px + 234, Mudclient.STRINGS[516], 0x00FF00, 0, 1, py + 204);   // "Prayer allowed"
        } else {
            m.li.drawStringCenter(px + 234, Mudclient.STRINGS[521], 0xFF0000, 0, 1, py + 204);   // "No prayer"
        }
        if (m.Rc != 0) {
            m.li.drawStringCenter(px + 234, Mudclient.STRINGS[518], 0xFF0000, 0, 1, py + 216);   // "No weapons"
        } else {
            m.li.drawStringCenter(px + 234, Mudclient.STRINGS[525], 0x00FF00, 0, 1, py + 216);   // "Weapons allowed"
        }
        m.li.drawStringCenter(px + 234, Mudclient.STRINGS[520], 0xFFFFFF, 0, 1, py + 230);       // "Both must confirm"

        if (!m.Cd) {
            m.li.drawSprite(-1, m.tg + 25, py + 238, px + 83);                // accept sprite
            m.li.drawSprite(-1, m.tg + 26, py + 238, px + 352 - 35);          // decline sprite
        } else {
            m.li.drawStringCenter(px + 234, Mudclient.STRINGS[212], 0xFFFF00, 0, 1, py + 250);   // "Waiting..."
        }

        if (m.Cf == 1) {   // clean h(int) L5771: ~m.Cf==-2 == Cf==1 (left-click Accept/Decline gate)
            // click outside panel → decline
            if (m.mouseX < px || m.mouseY < py || m.mouseX > px + 468 || m.mouseY > py + 262) {
                m.dd = false;
                m.Jh.newPacket(230, 0);
                m.Jh.finishPacket(21294);
            }
            // accept button: x 83..188, y 238..259   (FIX: was 83..153)
            if (m.mouseX >= px + 118 - 35 && m.mouseX < px + 118 + 70 && m.mouseY >= py + 238 && m.mouseY <= py + 259) {
                m.Cd = true;
                m.Jh.newPacket(77, 0);
                m.Jh.finishPacket(21294);
            }
            // decline button: x 317..423, y 238..259   (FIX: was 317..388)
            if (m.mouseX >= px + 352 - 35 && m.mouseX <= px + 353 + 70 && m.mouseY >= py + 238 && m.mouseY <= py + 259) {
                m.dd = false;
                m.Jh.newPacket(197, 0);
                m.Jh.finishPacket(21294);
            }
            m.Cf = 0;
        }
    }

    /** Duel setup window: your stake (lc inventory items vf/xe), opponent's offered items
     *  (Ke items Uf/df), opponent's committed stake (wj items zc/of), and the four rule
     *  checkboxes (No retreat fd / No magic Yi / No prayer vd / No weapons ff). A right-click
     *  on a stake cell opens an "offer hasPainted" sub-menu via the chatList MessageList (He).
     *  Opcodes: 8 DUEL_SETTINGS, 176 DUEL_ACCEPT, 197 DUEL_DECLINE.
     *
     *  FIX vs old: old version stubbed the right-click sub-menu builders and the menu-pick
     *  resolution, and mixed up the three render grids. Rewritten in full from the clean source. */
    void drawDuel(int param) {
        int menuPick = -1;
        if (m.Cf != 0 && m.Je) {
            menuPick = m.chatList.hitTestNoRender(m.mouseX, m.ad, m.Uk, (byte) -40, m.mouseY);
        }

        if (menuPick >= 0) {
            // --- a sub-menu entry was clicked: resolve to a stake change ---
            m.Cf = 0;
            m.Je = false;
            int action = m.chatList.getEntryXPos(-26, menuPick);   // 3 = your inventory, 4 = their offer
            int itemId = m.chatList.getEntryColorE(true, menuPick);
            int slot = -1;
            int total = 0;
            if (action != 3) {
                for (int i = 0; i < m.Ke; i++) {
                    if (m.Uf[i] == itemId) {
                        if (slot < 0) slot = i;
                        if (ClientIOException.itemY[itemId] == 0) { total = m.df[i]; break; }
                        total++;
                    }
                }
            }
            for (int i = 0; i < m.lc; i++) {
                if (m.vf[i] == itemId) {
                    if (slot < 0) slot = i;
                    if (ClientIOException.itemY[itemId] == 0) { total = m.xe[i]; break; }
                    total++;
                }
            }
            if (slot >= 0) {
                int amount = m.chatList.getEntryColorCode((byte) 97, menuPick);
                if (amount != -2) {
                    if (amount == 0) amount = total;   // "All"
                    if (action == 3) {
                        m.sendTradeOffer(amount, (byte) 124, slot);  // add to your stake (obf c(var28,(byte)124,var21))
                    } else {
                        m.sendDuelOffer(slot, amount, (byte) -78);   // remove from their offer
                    }
                } else {
                    m.ck = slot;                    // "X" → quantity entry dialog
                    if (action == 4) {
                        m.drawMenuOptions(NameHash.uiStrings, 12, 1, true);
                    } else {
                        m.drawMenuOptions(FontWidths.PROMPTS, param ^ 4, 2, true);
                    }
                }
            }
        } else if (m.gc == 0) {
            if (m.Cf == 1 && m.Tk == 0) m.Tk = 1;
            int relX = m.mouseX - 22;
            int relY = m.mouseY - 36;
            if (relX >= 0 && relY >= 0 && relX < 469 && relY < 262) {
                // --- left mouse: remove a staked item / toggle rules / accept / decline ---
                if (m.Tk > 0) {
                    // your stake grid (217..462 x, 31..235 y, 5 cols) → remove
                    if (relX > 216 && relY > 30 && relX < 462 && relY < 235) {
                        int slot = (relX - 217) / 49 + 5 * ((relY - 31) / 34);
                        if (slot >= 0 && slot < m.lc) {
                            this.drawTradeConfirm(-1, (byte) 9, slot);   // obf a(-1,(byte)9,var5) -> your-stake remove
                        }
                    }
                    // their offer grid (8..205 x, 30..129 y, 4 cols) → remove
                    if (relX > 8 && relY > 30 && relX < 205 && relY < 129) {
                        int slot = (relX - 9) / 49 + (relY - 31) / 34 * 4;
                        if (slot >= 0 && slot < m.Ke) {
                            m.sendDuelOffer(slot, -1, (byte) -78);
                        }
                    }
                    // rule checkboxes
                    boolean rulesChanged = false;
                    if (relX >= 93 && relY >= 221 && relX <= 104 && relY <= 232) { m.fd = !m.fd; rulesChanged = true; } // No retreat
                    if (relX >= 93 && relY >= 240 && relX <= 104 && relY <= 251) { m.Yi = !m.Yi; rulesChanged = true; } // No magic
                    if (relX >= 191 && relY >= 221 && relX <= 202 && relY <= 232) { m.vd = !m.vd; rulesChanged = true; } // No prayer
                    if (relX >= 191 && relY >= 240 && relX <= 202 && relY <= 251) { m.ff = !m.ff; rulesChanged = true; } // No weapons
                    if (rulesChanged) {
                        m.Jh.newPacket(8, 0);   // DUEL_SETTINGS
                        m.Jh.outBuffer.putByte(m.fd ? 1 : 0);
                        m.Jh.outBuffer.putByte(m.Yi ? 1 : 0);
                        m.Jh.outBuffer.putByte(m.vd ? 1 : 0);
                        m.Jh.outBuffer.putByte(m.ff ? 1 : 0);
                        m.Jh.finishPacket(param ^ 21254);
                        m.ki = false;
                        m.ke = false;
                    }
                    // accept button (218..287 x, 238..259 y)
                    if (relX >= 218 && relY >= 238 && relX <= 287 && relY <= 259) {
                        m.ke = true;
                        m.Jh.newPacket(176, param - 40);   // DUEL_ACCEPT
                        m.Jh.finishPacket(param + 21254);
                    }
                    // decline button (394..463 x, 238..259 y)
                    if (relX >= 394 && relY >= 238 && relX < 463 && relY < 259) {
                        m.Pj = false;
                        m.Jh.newPacket(197, 0);            // DUEL_DECLINE
                        m.Jh.finishPacket(21294);
                    }
                    m.Tk = 0;
                    m.Cf = 0;
                }

                // --- right mouse (Cf==2): open an "offer 1/5/10/all/cancel" sub-menu ---
                if (m.Cf == 2) {   // clean q(int) L16499: -3==~m.Cf == Cf==2 (Cf only ever 0/1/2)
                    // over your-stake grid → inventory item menu
                    if (relX > 216 && relY > 30 && relX < 462 && relY < 235) {
                        int w = m.zh.getPanelWidth(16256);
                        int hgt = m.zh.getPanelHeight(param - 21264);
                        m.rh = m.mouseX - w / 2;
                        m.fg = m.mouseY - 7;
                        m.se = true;
                        if (m.fg < 0) m.fg = 0;
                        if (m.rh < 0) m.rh = 0;
                        if (m.rh + w > 510) m.rh = 510 - w;
                        if (m.fg + hgt > 315) m.fg = 315 - hgt;

                        int slot = (relX - 217) / 49 + 5 * ((relY - 31) / 34);
                        if (slot >= 0 && slot < m.lc) {
                            int itemId = m.vf[slot];
                            m.Je = true;
                            m.chatList.setCount(0);
                            m.chatList.addEntryWithColor(itemId, Mudclient.STRINGS[34] + DecodeBuffer.chatFilterCache[itemId], 3, Mudclient.STRINGS[502], 1,  param + 3256);
                            m.chatList.addEntryWithColor(itemId, Mudclient.STRINGS[34] + DecodeBuffer.chatFilterCache[itemId], 3, Mudclient.STRINGS[509], 5,  param ^ 3272);
                            m.chatList.addEntryWithColor(itemId, Mudclient.STRINGS[34] + DecodeBuffer.chatFilterCache[itemId], 3, Mudclient.STRINGS[505], 10, 3296);
                            m.chatList.addEntryWithColor(itemId, Mudclient.STRINGS[34] + DecodeBuffer.chatFilterCache[itemId], 3, Mudclient.STRINGS[501], -1, 3296);
                            m.chatList.addEntryWithColor(itemId, Mudclient.STRINGS[34] + DecodeBuffer.chatFilterCache[itemId], 3, Mudclient.STRINGS[503], -2, 3296);
                            int mw = m.chatList.getPanelWidth(16256);
                            int mh = m.chatList.getPanelHeight(-21224);
                            m.Uk = m.mouseY - 7;
                            m.ad = m.mouseX - mw / 2;
                            if (m.ad < 0) m.ad = 0;
                            if (m.Uk < 0) m.Uk = 0;
                            if (m.ad + mw > 510) m.ad = 510 - mw;
                            if (m.Uk + mh > 316) m.Uk = 315 - mh;
                        }
                    }
                    // over their-offer grid → removable item menu
                    if (relX > 8 && relY > 30 && relX < 205 && relY < 133) {
                        int slot = (relX - 9) / 49 + 4 * ((relY - 31) / 34);
                        if (slot >= 0 && slot < m.Ke) {
                            int itemId = m.Uf[slot];
                            m.Je = true;
                            m.chatList.setCount(0);
                            m.chatList.addEntryWithColor(itemId, Mudclient.STRINGS[34] + DecodeBuffer.chatFilterCache[itemId], 4, Mudclient.STRINGS[163], 1,  param ^ 3272);
                            m.chatList.addEntryWithColor(itemId, Mudclient.STRINGS[34] + DecodeBuffer.chatFilterCache[itemId], 4, Mudclient.STRINGS[173], 5,  3296);
                            m.chatList.addEntryWithColor(itemId, Mudclient.STRINGS[34] + DecodeBuffer.chatFilterCache[itemId], 4, Mudclient.STRINGS[161], 10, 3296);
                            m.chatList.addEntryWithColor(itemId, Mudclient.STRINGS[34] + DecodeBuffer.chatFilterCache[itemId], 4, Mudclient.STRINGS[177], -1, 3296);
                            m.chatList.addEntryWithColor(itemId, Mudclient.STRINGS[34] + DecodeBuffer.chatFilterCache[itemId], 4, Mudclient.STRINGS[170], -2, param + 3256);
                            int mw = m.chatList.getPanelWidth(16256);
                            int mh = m.chatList.getPanelHeight(-21224);
                            m.Uk = m.mouseY - 7;
                            m.ad = m.mouseX - mw / 2;
                            if (m.ad < 0) m.ad = 0;
                            if (m.Uk < 0) m.Uk = 0;
                            if (m.ad + mw > 511) m.ad = 510 - mw;
                            if (mh + m.Uk > 315) m.Uk = 315 - mh;
                        }
                    }
                    m.Cf = 0;
                }

                // dismiss the sub-menu when the cursor leaves its bounds
                if (m.Je) {
                    int mw = m.chatList.getPanelWidth(16256);
                    int mh = m.chatList.getPanelHeight(-21224);
                    if (m.ad - 10 > m.mouseX || m.Uk - 10 > m.mouseY
                            || m.ad + mw + 10 < m.mouseX || m.Uk + mh + 10 < m.mouseY) {
                        m.Je = false;
                    }
                }
            } else if (m.Cf != 0) {
                // click outside the panel → decline duel
                m.Pj = false;
                m.Jh.newPacket(197, 0);
                m.Jh.finishPacket(21294);
            }
        }

        // --- draw panel ---
        if (m.Pj) {
            final int px = 22, py = 36;
            m.li.drawBox(px, (byte) 112, 0xC90B1D, py, 12, 468);   // maroon panel bg
            int grey = 0x989898;
            m.li.drawBoxAlpha(160, px, 18, 0, py + 12, 468, grey);
            m.li.drawBoxAlpha(160, px, 248, 0, py + 30, 8, grey);
            m.li.drawBoxAlpha(160, px + 205, 248, 0, py + 30, 11, grey);
            m.li.drawBoxAlpha(160, px + 462, 248, 0, py + 30, 6, grey);
            m.li.drawBoxAlpha(160, px + 8, 24, param ^ 40, py + 99, 197, grey);
            m.li.drawBoxAlpha(160, px + 8, 23, 0, py + 192, 197, grey);
            m.li.drawBoxAlpha(160, px + 8, 20, 0, py + 258, 197, grey);
            m.li.drawBoxAlpha(160, px + 216, 43, 0, py + 235, 246, grey);
            int lgrey = 0xD0D0D0;
            m.li.drawBoxAlpha(160, px + 8, 69, 0, py + 30, 197, lgrey);
            m.li.drawBoxAlpha(160, px + 8, 69, 0, py + 123, 197, lgrey);
            m.li.drawBoxAlpha(160, px + 8, 43, param - 40, py + 215, 197, lgrey);
            m.li.drawBoxAlpha(160, px + 216, 205, 0, py + 30, 246, lgrey);

            for (int r = 0; r < 3; r++) m.li.drawLineHoriz(197, 0, px + 8, py + 30 + 34 * r, (byte) 58);
            for (int r = 0; r < 4; r++) m.li.drawLineHoriz(197, 0, px + 8, r * 34 + py + 123, (byte) -88);
            for (int r = 0; r < 7; r++) m.li.drawLineHoriz(246, 0, px + 216, r * 34 + py + 30, (byte) -40);
            for (int c = 0; c < 6; c++) {
                m.li.drawLineVert(49 * c + 8 + px, py + 30, 0, 69, 0);
                if (c < 5) m.li.drawLineVert(49 * c + px + 8, py + 123, 0, 69, 0);
                m.li.drawLineVert(c * 49 + px + 216, py + 30, 0, 205, 0);
            }
            m.li.drawLineHoriz(197, 0, px + 8, py + 215, (byte) 97);
            m.li.drawLineHoriz(197, 0, px + 8, py + 257, (byte) 99);
            m.li.drawLineVert(px + 8, py + 215, 0, 43, 0);
            m.li.drawLineVert(px + 204, py + 215, 0, 43, 0);

            m.li.drawstring(Mudclient.STRINGS[508] + m.Lg, px + 1, py + 10, 0xFFFFFF, false, 1);  // "Duel with <name>"
            m.li.drawstring(Mudclient.STRINGS[498], px + 9, py + 27, 0xFFFFFF, false, 4);            // "Your stake"
            m.li.drawstring(Mudclient.STRINGS[500], px + 9, py + 120, 0xFFFFFF, false, 4);           // "Opponent's stake"
            m.li.drawstring(Mudclient.STRINGS[499], px + 9, py + 212, 0xFFFFFF, false, 4);           // "Their offer"
            m.li.drawstring(Mudclient.STRINGS[171], px + 216, py + 27, 0xFFFFFF, false, 4);          // "Options"
            m.li.drawstring(Mudclient.STRINGS[506], px + 8 + 1, py + 215 + 16, 0xFFFF00, false, 3);
            m.li.drawstring(Mudclient.STRINGS[496], px + 8 + 1, py + 250, 0xFFFF00, false, 3);
            m.li.drawstring(Mudclient.STRINGS[507], px + 8 + 102, py + 231, 0xFFFF00, false, 3);
            m.li.drawstring(Mudclient.STRINGS[497], px + 8 + 102, py + 215 + 35, 0xFFFF00, false, 3);

            // rule checkboxes (box + tick)
            m.li.drawBoxEdge(px + 93, 11, py + 215 + 6, param + 27745, 11, 0xFFFF00);
            if (m.fd) m.li.drawBox(px + 95, (byte) -109, 0xFFFF00, py + 215 + 8, 7, 7);
            m.li.drawBoxEdge(px + 93, 11, py + 215 + 25, 27785, 11, 0xFFFF00);
            if (m.Yi) m.li.drawBox(px + 95, (byte) -127, 0xFFFF00, py + 215 + 27, 7, 7);
            m.li.drawBoxEdge(px + 191, 11, py + 215 + 6, 27785, 11, 0xFFFF00);
            if (m.vd) m.li.drawBox(px + 193, (byte) -106, 0xFFFF00, py + 215 + 8, 7, 7);
            m.li.drawBoxEdge(px + 191, 11, py + 215 + 25, param + 27745, 11, 0xFFFF00);
            if (m.ff) m.li.drawBox(px + 193, (byte) 59, 0xFFFF00, py + 215 + 27, 7, 7);

            if (!m.ke) m.li.drawSprite(-1, m.tg + 25, py + 238, px + 217);   // accept sprite
            m.li.drawSprite(-1, m.tg + 26, py + 238, px + 394);                 // decline sprite
            if (m.ki) {
                m.li.drawstringRight(px + 341, Mudclient.STRINGS[168], 0xFFFFFF, 0, 1, py + 246);
                m.li.drawstringRight(px + 341, Mudclient.STRINGS[165], 0xFFFFFF, 0, 1, py + 256);
            }
            if (m.ke) {
                m.li.drawstringRight(px + 217 + 35, Mudclient.STRINGS[176], 0xFFFFFF, 0, 1, py + 246);
                m.li.drawstringRight(px + 252, Mudclient.STRINGS[160], 0xFFFFFF, 0, 1, py + 256);
            }

            // your stake grid (vf/xe, 5 cols, starts at px+217)
            for (int i = 0; i < m.lc; i++) {
                int cellX = px + 217 + i % 5 * 49;
                int cellY = py + 31 + 34 * (i / 5);
                m.li.spriteClipping(cellY, TextEncoder.scratchIntArray2[m.vf[i]], 0, false, 0,
                    m.sg + Surface.unusedIntsBb[m.vf[i]], 32, 48, cellX, 1);
                if (ClientIOException.itemY[m.vf[i]] == 0) {
                    m.li.drawstring("" + m.xe[i], cellX + 1, cellY + 10, 0xFFFF00, false, 1);
                }
            }
            // opponent's offered items (Uf/df, 4 cols, starts at px+9)
            for (int i = 0; i < m.Ke; i++) {
                int cellX = px + 9 + i % 4 * 49;
                int cellY = py + 31 + i / 4 * 34;
                m.li.spriteClipping(cellY, TextEncoder.scratchIntArray2[m.Uf[i]], 0, false, 0,
                    m.sg + Surface.unusedIntsBb[m.Uf[i]], 32, 48, cellX, param - 39);
                if (ClientIOException.itemY[m.Uf[i]] == 0) {
                    m.li.drawstring("" + m.df[i], cellX + 1, cellY + 10, 0xFFFF00, false, 1);
                }
                if (cellX < m.mouseX && cellX + 48 > m.mouseX && cellY < m.mouseY && cellY + 32 > m.mouseY) {
                    m.li.drawstring(DecodeBuffer.chatFilterCache[m.Uf[i]] + Mudclient.STRINGS[159] + CharTable.itemDescriptions[m.Uf[i]],
                        px + 8, py + 273, 0xFFFF00, false, 1);   // examine tooltip
                }
            }
            // opponent's committed stake (zc/of, 4 cols, second block at py+124)
            for (int i = 0; i < m.wj; i++) {
                int cellX = px + 9 + i % 4 * 49;
                int cellY = py + 124 + i / 4 * 34;
                m.li.spriteClipping(cellY, TextEncoder.scratchIntArray2[m.zc[i]], 0, false, 0,
                    Surface.unusedIntsBb[m.zc[i]] + m.sg, 32, 48, cellX, param ^ 41);
                if (ClientIOException.itemY[m.zc[i]] == 0) {
                    m.li.drawstring("" + m.of[i], cellX + 1, cellY + 10, 0xFFFF00, false, 1);
                }
                if (m.mouseX > cellX && cellX + 48 > m.mouseX && m.mouseY > cellY && cellY + 32 > m.mouseY) {
                    m.li.drawstring(DecodeBuffer.chatFilterCache[m.zc[i]] + Mudclient.STRINGS[159] + CharTable.itemDescriptions[m.zc[i]],
                        px + 8, py + 273, 0xFFFF00, false, 1);
                }
            }

            if (m.Je) {
                m.chatList.hitTest(m.Uk, m.ad, m.mouseY, (byte) -12, m.mouseX);   // render sub-menu
            }
        }
    }

    /** "Welcome to RuneScape" box on login: last-login, recovery-questions reminder, unread
     *  messages, subscription/members status, and a "Click here to play" dismiss button.
     *  Sb = subscription-days marker (201 = none); ce = recovery set time; id = unread count. */
    void drawWelcome(int param) {
        int h = 65;
        if (m.Sb != 201) h += 60;
        if (m.id > 0)    h += 30;
        if (m.ce != 0)   h += 45;

        int top = 167 - h / 2;
        m.li.drawBox(56, (byte) 77, 0, top, h, 400);
        m.li.drawBoxEdge(56, 400, top, 27785, h, 0xFFFFFF);

        int y = top + 20;
        m.li.drawstringRight(256, Mudclient.STRINGS[667] + m.wi.message, 0xFFFF00, 0, 4, y);   // "Welcome <name>"
        y += 30;

        // last-login line
        String last;
        if (m.hi == 0)      last = Mudclient.STRINGS[658];               // "first time"
        else if (m.hi == 1) last = Mudclient.STRINGS[665];               // "yesterday"
        else                   last = m.hi + Mudclient.STRINGS[652];     // "hasPainted days ago"

        // recovery-questions reminder block
        if (m.ce != 0) {
            m.li.drawstringRight(256, Mudclient.STRINGS[655] + last, 0xFFFFFF, 0, 1, y);
            y += 15;
            if (m.ve == null) {
                m.ve = m.formatNumber(param ^ 0x128B, m.ce);   // c(int,int): date string
            }
            m.li.drawstringRight(256, Mudclient.STRINGS[662] + m.ve, 0xFFFFFF, param ^ -4853, 1, y);
            y += 15;
            y += 15;
        }

        // unread messages block
        if (m.id > 0) {
            if (m.id == 1) {
                m.li.drawstringRight(256, Mudclient.STRINGS[656], 0xFFFFFF, 0, 1, y);                // "no unread"
            } else {
                m.li.drawstringRight(256, Mudclient.STRINGS[668] + (m.id - 1) + Mudclient.STRINGS[661], 0xFFFFFF, param + 4853, 1, y);
            }
            y += 15;
            y += 15;
        }

        // subscription/members status block
        if (m.Sb != 201) {
            if (m.Sb == 200) {   // ~Sb == -201
                m.li.drawstringRight(256, Mudclient.STRINGS[660], 0xFF8000, 0, 1, y);
                y += 15;
                m.li.drawstringRight(256, Mudclient.STRINGS[657], 0xFF8000, param ^ -4853, 1, y);
                y += 15;
                m.li.drawstringRight(256, Mudclient.STRINGS[663], 0xFF8000, 0, 1, y);
                y += 15;
            } else {
                String sub;
                if (m.Sb == 0)      sub = Mudclient.STRINGS[654];
                else if (m.Sb == 1) sub = Mudclient.STRINGS[659];
                else                   sub = m.Sb + Mudclient.STRINGS[652];
                m.li.drawstringRight(256, sub + Mudclient.STRINGS[666], 0xFF8000, 0, 1, y);
                y += 15;
                m.li.drawstringRight(256, Mudclient.STRINGS[664], 0xFF8000, 0, 1, y);
                y += 15;
                m.li.drawstringRight(256, Mudclient.STRINGS[663], 0xFF8000, param + 4853, 1, y);
                y += 15;
            }
            y += 15;
        }

        // "Click here to play" — red on hover
        int colour = 0xFFFFFF;
        if (m.mouseY > y - 12 && m.mouseY <= y && m.mouseX > 106 && m.mouseX < 406) {
            colour = 0xFF0000;
        }
        m.li.drawstringRight(256, Mudclient.STRINGS[126], colour, param ^ param, 1, y);

        // dismiss on click (on the button, or anywhere outside the panel)
        if (m.Cf == 2) {
            if (colour == 0xFF0000) {
                m.Oh = false;
            }
            if ((m.mouseX < 86 || m.mouseX > 426) && (m.mouseY < top || m.mouseY > top + h)) {
                m.Oh = false;
            }
        }
        m.Cf = 0;
    }

    /**
     * Rebuild the inventory item draw-cache from the previous frame's draw list and the
     * current inventory, de-duplicating item ids and appending new ones up to the cache
     * capacity (m.Gi). (The trailing junk divide in the clean output is an opaque-predicate
     * artifact and is dropped.)
     * obf: void C(int)
     */
    void drawHelpMenu(int param) {
        m.drawListCount = m.drawListSize;   // obf: vj = fj
        for (int i = 0; i < m.drawListSize; i++) {
            m.drawListIds[i] = m.drawListCurrent[i];   // obf: ae[i] = ci[i]
            m.drawListY[i] = m.drawListYShadow[i];     // obf: di[i] = Xe[i]
        }

        for (int n = 0; n < m.lc; n++) {
            if (m.Gi <= m.drawListCount) {   // obf: m.Gi <= vj — cache full
                break;
            }
            int itemId = m.vf[n];   // obf: m.vf[n]
            boolean found = false;
            for (int j = 0; j < m.drawListCount; j++) {
                if (m.drawListIds[j] == itemId) {
                    found = true;
                    break;
                }
            }
            if (!found) {
                m.drawListIds[m.drawListCount] = itemId;
                m.drawListY[m.drawListCount] = 0;
                m.drawListCount++;
            }
        }
    }

    /**
     * Draw the "Click here to close window" modal button (and the help panel frame).
     * Highlights the close text on hover; on a LEFT click (m.Cf==1) over the text, or a
     * LEFT click fully outside the modal box, closes it (m.mh=false).
     * obf: void l(byte)
     */
    void drawCloseButton(byte param) {
        int w = 400;
        if (param != -115) {
            m.qd = 64;   // obf: m.qd
        }
        int h = 100;
        if (m.Wk) {   // obf: m.Wk
            h = 300;          // (clean assigns 450 then 300; first is dead)
        }
        m.li.drawBox(256 - w / 2, (byte)122, 0, 167 - h / 2, h, w);
        m.li.drawBoxEdge(256 - w / 2, w, 167 - h / 2, 27785, h, 0xFFFFFF);
        m.li.centrepara(w - 40, m.Cj, 256, 92, 1, 167 - (h / 2 - 20), true, 0xFFFFFF);

        int labelY = 157 + h / 2;
        int labelColor = 0xFFFFFF;
        // obf: ~m.mouseY < ~(var4-12) && ~m.mouseY >= ~var4 && ~m.mouseX < -107 && -407 < ~m.mouseX
        if (m.mouseY > labelY - 12 && m.mouseY <= labelY && m.mouseX > 106 && m.mouseX < 406) {
            labelColor = 0xFF0000;
        }
        m.li.drawstringRight(256, Mudclient.STRINGS[126], labelColor, 0, 1, labelY);   // "Click here to close window"

        if (m.Cf == 1) {   // obf: ~m.Cf == -2 → m.Cf == 1 (LEFT click)
            if (labelColor == 0xFF0000) {
                m.mh = false;   // obf: m.mh
            }
            // close when the click lands fully outside the modal box (X outside AND Y outside)
            if ((m.mouseX < 256 - w / 2 || m.mouseX > 256 + w / 2) && (m.mouseY < 167 - h / 2 || m.mouseY > 167 + h / 2)) {
                m.mh = false;
            }
        }
        m.Cf = 0;   // obf: m.Cf = 0
    }

    /**
     * Game settings / privacy "Configuration" tab.
     * Renders the privacy toggles (private chat / trade / members) and the option
     * toggles (auto-retaliate / mouse buttons / camera / members-only), plus the
     * logout link(s). When processClicks is set, hit-tests the rows and, on a LEFT
     * click (m.Cf==1 for every row), flips the relevant flag and emits the matching
     * packet (opcode 111 GAME_SETTINGS_CHANGED for the privacy rows, opcode 64
     * PRIVACY_SETTINGS_CHANGED for the option rows).
     * obf: void b(int,boolean)
     */
    void drawGameSettings(int param, boolean processClicks) {
        int panelX = -199 + m.li.width;          // obf: var3 = -199 + m.li.u  (right panel left edge)
        m.li.drawSprite(-1, m.tg + 6, 3, panelX - 49);  // header border line
        int panelY = 36;                        // obf: var4
        int panelW = 196;                       // obf: var5

        // Section background fills (colour from ISAAC scratch helper o.a).
        m.li.drawBoxAlpha(160, panelX, 65, param ^ 15, 36, panelW, ISAAC.packColor(181, param + 9555, 181, 181));
        m.li.drawBoxAlpha(160, panelX, 65, 0, 101, panelW, ISAAC.packColor(201, param ^ 9581, 201, 201));
        m.li.drawBoxAlpha(160, panelX, 95, 0, 166, panelW, ISAAC.packColor(181, 9570, 181, 181));
        m.li.drawBoxAlpha(160, panelX, m.Kd ? 55 : 40, 0, 261, panelW, ISAAC.packColor(201, 9570, 201, 201));

        int textX = panelX + 3;     // obf: var6
        int textY = panelY + 15;    // obf: var18

        // "Private chat:" header
        m.li.drawstring(Mudclient.STRINGS[138], textX, textY, 0, false, 1);
        textY += 15;
        // chat-private on/off
        m.li.drawstring(m.privacyChatOn ? Mudclient.STRINGS[151] : Mudclient.STRINGS[136], textX, textY, 0xFFFFFF, false, 1);

        textY += 15;
        // trade/duel-private on/off
        m.li.drawstring(m.privacyTradeOn ? Mudclient.STRINGS[144] : Mudclient.STRINGS[146], textX, textY, 0xFFFFFF, false, 1);

        textY += 15;
        // members-private on/off (members accounts only)
        if (m.Pg) {
            m.li.drawstring(m.privacyMembersOn ? Mudclient.STRINGS[155] : Mudclient.STRINGS[141], textX, textY, 0xFFFFFF, false, 1);
        }

        // static labels (obf: var18 += param, where param is normally 0)
        textY += param;
        m.li.drawstring(Mudclient.STRINGS[145], textX, textY, 0xFFFFFF, false, 0);
        textY += 15;
        m.li.drawstring(Mudclient.STRINGS[143], textX, textY, 0xFFFFFF, false, 0);
        textY += 15;
        m.li.drawstring(Mudclient.STRINGS[130], textX, textY, 0xFFFFFF, false, 0);
        textY += 15;

        // combat-mode 3-state display (obf: m.Yd): 0→[135], 1→[137], else→[132]
        if (m.Yd == 0) {
            m.li.drawstring(Mudclient.STRINGS[135], textX, textY, 0xFFFFFF, false, 0);
        } else if (m.Yd == 1) {
            m.li.drawstring(Mudclient.STRINGS[137], textX, textY, 0xFFFFFF, false, 0);
        } else {
            m.li.drawstring(Mudclient.STRINGS[132], textX, textY, 0xFFFFFF, false, 0);
        }

        textY += 15;
        textY += 5;
        // "Sound effects:" / "Camera:" headers
        m.li.drawstring(Mudclient.STRINGS[139], panelX + 3, textY, 0, false, 1);
        textY += 15;
        m.li.drawstring(Mudclient.STRINGS[133], panelX + 3, textY, 0, false, 1);
        textY += 15;

        // auto-retaliate on/off (obf: m.De): m.De!=0 → [153] "On", else [131] "Off"
        m.li.drawstring(m.De != 0 ? Mudclient.STRINGS[153] : Mudclient.STRINGS[131], panelX + 3, textY, 0xFFFFFF, false, 1);

        textY += 15;
        // mouse buttons one/two (obf: m.dc): 0 → [142], else [150]
        m.li.drawstring(m.dc == 0 ? Mudclient.STRINGS[142] : Mudclient.STRINGS[150], panelX + 3, textY, 0xFFFFFF, false, 1);

        textY += 15;
        // camera auto/manual (obf: m.Vg): 0 → [152], else [140]
        m.li.drawstring(m.Vg == 0 ? Mudclient.STRINGS[140] : Mudclient.STRINGS[152], panelX + 3, textY, 0xFFFFFF, false, 1);

        textY += 15;
        // members-only option (obf: m.ui), members accounts only: !=0 → [154], else [129]
        if (m.Pg) {
            m.li.drawstring(m.ui != 0 ? Mudclient.STRINGS[154] : Mudclient.STRINGS[129], panelX + 3, textY, 0xFFFFFF, false, 1);
        }

        textY += 15;
        // members-world logout shortcut row (highlight on hover)
        if (m.Kd) {
            int color = 0xFFFFFF;
            textY += 5;
            if (m.mouseX > textX && m.mouseX < textX + panelW && m.mouseY > textY - 12 && m.mouseY < textY + 4) {
                color = 0xFFFF00;
            }
            m.li.drawstring(Mudclient.STRINGS[134], textX, textY, color, false, 1);
            textY += 15;
        }

        textY += 5;
        // "Logout" label
        m.li.drawstring(Mudclient.STRINGS[147], textX, textY, 0, false, 1);
        int logoutColor = 0xFFFFFF;
        textY += 15;
        if (m.mouseX > textX && m.mouseX < textX + panelW && m.mouseY > textY - 12 && m.mouseY < textY + 4) {
            logoutColor = 0xFFFF00;
        }
        m.li.drawstring(Mudclient.STRINGS[149], panelX + 3, textY, logoutColor, false, 1);

        if (!processClicks) {
            return;
        }

        // --- click handling ---
        // relX/relY measured from the panel's top-left; gate to the panel rectangle.
        // obf: var3 = 199 - m.li.u + m.mouseX ; var15 = -36 + m.mouseY
        int relX = 199 - m.li.width + m.mouseX;
        int relY = m.mouseY - 36;
        if (relX < 0 || relY < 0 || relX >= 196 || relY >= 265) {
            return;
        }

        int px = m.li.width - 199 + 3;   // obf: var6 (re-derived for hit-tests)
        int pw = 196;                   // obf: var5
        int rowY = 66;                  // obf: var18 = 30 + 36

        // Private chat toggle (LEFT click)
        if (m.mouseX > px && m.mouseX < px + pw && m.mouseY > rowY - 12 && m.mouseY < rowY + 4 && m.Cf == 1) {
            m.privacyChatOn = !m.privacyChatOn;
            m.Jh.newPacket(111, 0);              // opcode 111 GAME_SETTINGS_CHANGED
            m.Jh.outBuffer.putByte(0);             // setting id 0 = private chat
            m.Jh.outBuffer.putByte(m.privacyChatOn ? 1 : 0);
            m.Jh.finishPacket(21294);               // flush
        }
        rowY += 15;
        // Trade/duel privacy toggle (LEFT click)
        if (m.mouseX > px && m.mouseX < px + pw && m.mouseY > rowY - 12 && m.mouseY < rowY + 4 && m.Cf == 1) {
            m.privacyTradeOn = !m.privacyTradeOn;
            m.Jh.newPacket(111, param - 15);
            m.Jh.outBuffer.putByte(2);     // setting id 2 = trade
            m.Jh.outBuffer.putByte(m.privacyTradeOn ? 1 : 0);
            m.Jh.finishPacket(param ^ 21281);
        }
        rowY += 15;
        // Members privacy toggle (members account, LEFT click)
        if (m.Pg && m.mouseX > px && m.mouseX < px + pw && m.mouseY > rowY - 12 && m.mouseY < rowY + 4 && m.Cf == 1) {
            m.privacyMembersOn = !m.privacyMembersOn;
            m.Jh.newPacket(111, 0);
            m.Jh.outBuffer.putByte(3);    // setting id 3 = members
            m.Jh.outBuffer.putByte(m.privacyMembersOn ? 1 : 0);
            m.Jh.finishPacket(21294);
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
        if (m.mouseX > px && m.mouseX < px + pw && m.mouseY > rowY - 12 && m.mouseY < rowY + 4 && m.Cf == 1) {
            m.De = 1 - m.De;
            optionsChanged = true;
        }
        rowY += 15;
        // Mouse-buttons toggle (LEFT click)
        if (m.mouseX > px && m.mouseX < px + pw && m.mouseY > rowY - 12 && m.mouseY < rowY + 4 && m.Cf == 1) {
            m.dc = 1 - m.dc;
            optionsChanged = true;
        }
        rowY += 15;
        // Camera toggle (LEFT click)
        if (m.mouseX > px && m.mouseX < px + pw && m.mouseY > rowY - 12 && m.mouseY < rowY + 4 && m.Cf == 1) {
            m.Vg = 1 - m.Vg;
            optionsChanged = true;
        }
        rowY += 15;
        // Members-only option toggle (members account, LEFT click)
        if (m.Pg && m.mouseX > px && m.mouseX < px + pw && m.mouseY > rowY - 12 && m.mouseY < rowY + 4 && m.Cf == 1) {
            optionsChanged = true;
            m.ui = 1 - m.ui;
        }
        rowY += 15;
        if (optionsChanged) {
            // opcode 64 PRIVACY_SETTINGS_CHANGED: (m.Vg camera, m.dc mouse, m.De retaliate, m.ui members)
            m.packets.sendPrivacySettings(m.Vg, m.dc, m.De, param + 64, m.ui);
        }

        // Members-world quick-logout shortcut
        if (m.Kd) {
            rowY += 5;
            if (m.mouseX > px && m.mouseX < px + pw && m.mouseY > rowY - 12 && m.mouseY < rowY + 4 && m.Cf == 1) {
                // obf: a(o.g, param-3, 9, false) — o.g is a String[] menu-option list,
                // so this is m.drawMenuOptions(String[],int,int,boolean), NOT an inventory click.
                // o.g (obf static String[]) is deob ISAAC.decoyStrings (the renamed decoy table).
                m.drawMenuOptions(ISAAC.decoyStrings, param - 3, 9, false);
                m.qc = 0;
            }
            rowY += 15;
        }

        // Logout link
        rowY += 20;
        if (m.mouseX > px && m.mouseX < px + pw && m.mouseY > rowY - 12 && m.mouseY < rowY + 4 && m.Cf == 1) {
            m.requestLogout(0);   // obf: B(0) — opcode 102 LOGOUT
        }
        m.Cf = 0;
    }

}
