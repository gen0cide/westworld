package client;

import client.util.ClientIOException;
import client.shell.InputState;

/**
 * TradeDuelBankPackets — outgoing trade / duel / bank packet writers extracted from Mudclient.
 *
 * <p>Three self-contained outgoing-packet methods that open m.Jh, write fixed args, and flush.
 * All field access is via the {@code m} back-reference; shared callbacks (menuHitTest,
 * showServerMessage) are invoked via {@code m} as well.  No this-identity, no AWT, no
 * synchronized.
 *
 * <p>Opcodes:
 * <ul>
 *   <li>22 / 23 — BANK_WITHDRAW / BANK_DEPOSIT  (bankSend)</li>
 *   <li>33       — DUEL_OFFER_ITEM               (sendDuelOffer, sendDuelItems)</li>
 *   <li>46       — PLAYER_ADDED_ITEMS_TO_TRADE_OFFER (sendTradeOffer)</li>
 * </ul>
 */
class TradeDuelBankPackets {
    final Mudclient m;

    TradeDuelBankPackets(Mudclient m) {
        this.m = m;
    }

    /** Helper used by drawBank's click dispatch: begin a bank op (22 withdraw / 23 deposit),
     *  write the item id, the amount, and the obfuscated session "magic" word, then flush.
     *  (In the obfuscated source these were five inlined Jh writes per button.) */
    void bankSend(int opcode, int itemId, int amount, int magic) {
        m.Jh.newPacket(opcode, 0);
        m.Jh.outBuffer.putShort(itemId);
        m.Jh.outBuffer.putInt(amount);
        m.Jh.outBuffer.putInt(magic);
        m.Jh.finishPacket(21294);
    }

    /**
     * Remove items of one type from the current DUEL offer, then resend the whole offer.
     * Stackable -> decrement the entry's count; non-stackable -> drop up to {@code amount}
     * matching entries. Resending clears both duel-accept flags so the offer must be re-accepted.
     * Sends opcode 33 (DUEL_OFFER_ITEM): item count, then per item: id (short) + qty (int).
     */
    // obf: private final void a(int,int,byte)   [byte param var3 is anti-tamper guard: send only if var3 == -78]
    void sendDuelOffer(int slot, int qty, byte unused) {
        int itemId = m.Uf[slot];                       // obf: Uf[var1]
        int amount = (qty >= 0) ? qty : m.Tk;        // obf: ~var2<=-1 ? var2 : Tk

        if (ClientIOException.itemY[itemId] == 0) {                     // obf: fa.e[var4] == 0  -> STACKABLE
            // Stackable: one entry, decrement its count; remove the slot if it empties.
            m.df[slot] -= amount;                   // obf: df[var1] -= var5
            if (m.df[slot] <= 0) {                  // obf: !(0 < df[var1])
                m.Ke--;                            // obf: Ke--
                for (int j = slot; j < m.Ke; j++) {
                    m.Uf[j]    = m.Uf[j + 1];
                    m.df[j] = m.df[j + 1];
                }
            }
        } else {
            // Non-stackable: each unit is its own entry; drop up to `amount` matching entries.
            int removed = 0;                                           // obf: var11
            for (int i = 0; i < m.Ke && removed < amount; i++) { // obf: var7<Ke && ~var5>=~var11
                if (m.Uf[i] == itemId) {               // obf: Uf[var7] == var4
                    m.Ke--;
                    removed++;
                    for (int j = i; j < m.Ke; j++) {
                        m.Uf[j]    = m.Uf[j + 1];
                        m.df[j] = m.df[j + 1];
                    }
                    i--;
                }
            }
        }

        m.Jh.newPacket(33, 0); // DUEL_OFFER_ITEM
        m.Jh.outBuffer.putByte(m.Ke); // obf: Jh.f.c(Ke, ..)
        for (int i = 0; i < m.Ke; i++) {
            m.Jh.outBuffer.putShort(m.Uf[i]);     // obf: Jh.f.e(.., Uf[var12])
            m.Jh.outBuffer.putInt((int) m.df[i]); // obf: Jh.f.b(.., df[var12])
        }
        m.Jh.finishPacket(21294);

        m.ke = false;          // obf: ke = false (ours)
        m.ki = false; // obf: ki = false (theirs)
    }

    /**
     * Remove items of one type from the current TRADE offer, then resend the whole offer
     * (mirrors {@link #sendDuelOffer}). Stackable -> decrement the entry's count; non-stackable
     * -> drop up to {@code amount} matching entries. Resending clears both trade-accept flags.
     * Sends opcode 46 (PLAYER_ADDED_ITEMS_TO_TRADE_OFFER): item count, then per item:
     * id (short) + qty (int).
     */
    // obf: private final void c(int,byte,int)   [byte param var2 is anti-tamper guard: send only if var2 > 120]
    void sendTradeOffer(int qty, byte unused, int slot) {
        int itemId = m.Qf[slot];                      // obf: Qf[var3]
        int amount = (qty < 0) ? m.Tk : qty;   // obf: var1<0 ? Tk : var1

        if (ClientIOException.itemY[itemId] != 0) {               // obf: fa.e[var4] != 0  -> NON-stackable
            // Non-stackable: drop up to `amount` matching offer entries.
            int removed = 0;                                     // obf: var6
            for (int i = 0; i < m.tradeItemsCount && removed < amount; i++) { // obf: var7<mf && var6<var5
                if (m.Qf[i] == itemId) {              // obf: ~Qf[var7] == ~var4
                    removed++;
                    m.tradeItemsCount--;
                    for (int j = i; j < m.tradeItemsCount; j++) {
                        m.Qf[j]     = m.Qf[j + 1];
                        m.jj[j] = m.jj[j + 1];
                    }
                    i--;
                }
            }
        } else {
            // Stackable: one entry, decrement its count; remove the slot if it empties.
            m.jj[slot] -= amount;                 // obf: jj[var3] -= var5
            if (m.jj[slot] <= 0) {                // obf: -1 <= ~jj[var3]
                m.tradeItemsCount--;
                for (int j = slot; j < m.tradeItemsCount; j++) {
                    m.Qf[j]     = m.Qf[j + 1];
                    m.jj[j] = m.jj[j + 1];
                }
            }
        }

        m.Jh.newPacket(46, 0); // PLAYER_ADDED_ITEMS_TO_TRADE_OFFER
        m.Jh.outBuffer.putByte(m.tradeItemsCount); // obf: Jh.f.c(mf, ..)
        for (int i = 0; i < m.tradeItemsCount; i++) {
            m.Jh.outBuffer.putShort(m.Qf[i]);          // obf: Jh.f.e(393, Qf[var12])
            m.Jh.outBuffer.putInt((int) m.jj[i]);  // obf: Jh.f.b(.., jj[var12])
        }
        m.Jh.finishPacket(21294);

        m.Mi = false;          // obf: Mi = false (ours)
        m.md = false; // obf: md = false (theirs)
    }

    /**
     * Add/remove an inventory item to/from the local duel-stake offer and resend it
     * (opcode 33, DUEL_OFFER_ITEM).  Maintains the parallel Uf[]/df[] offer slots,
     * clamping non-stackable items to the held count (xe[slot]).
     *
     * Skeleton mislabels this "updateCamera" (L23985 in normalized) — it is a network
     * offer-update, nothing to do with camera.  Faithful to the clean base.
     *
     * obf: void b(int p1, int delta, int invSlot)
     *   p1: when < 2, fire the offer-confirmation callback; delta: +add / -remove;
     *   invSlot: inventory slot whose item id (vf[]) is being offered.
     */
    void sendDuelItems(int p1, int delta, int invSlot) {
        boolean changed = false;
        int matched = 0;                                      // count of stackable duplicates seen
        int itemId = m.vf[invSlot];

        // pass over the existing offer slots looking for this item
        for (int i = 0; i < m.Ke; i++) {                        // Ke = current offer slot count
            if (itemId == m.Uf[i]) {
                if (ClientIOException.itemY[itemId] == 0) {                      // non-stackable
                    if (delta < 0) {                          // remove: tick df[] up to held count, Tk times
                        for (int n = 0; n < m.Tk; n++) {
                            if (m.df[i] < m.xe[invSlot]) {
                                m.df[i]++;
                            }
                            changed = true;
                        }
                    } else {                                  // add: bump df[] by delta, clamp to held
                        m.df[i] += delta;
                        if (m.xe[invSlot] < m.df[i]) {
                            m.df[i] = m.xe[invSlot];
                        }
                        changed = true;
                    }
                    // non-stackable match handled — do NOT count it
                } else {
                    matched++;                                // stackable duplicate
                }
            }
        }

        if (p1 < 2) {
            m.packets.sendRemoveFriend((String)null, (byte)-34);                  // offer-confirmation callback
        }

        int slotsForItem = m.menus.menuHitTest(103, itemId);               // slots this item is allowed to occupy
        if (matched >= slotsForItem) {
            changed = true;
        }
        if (InputState.slotFlags[itemId] == 1) {                               // item flagged non-offerable
            changed = true;
            m.showServerMessage(false, null, 0, Mudclient.STRINGS[217], 0, 0, null, null);  // "cannot be added" message
        }

        // item not yet in the offer: add it
        if (!changed) {
            if (delta < 0) {
                // remove path with no existing slot: add a single df=1 slot (if room)
                if (m.Ke < 8) {
                    m.Uf[m.Ke] = itemId;
                    m.df[m.Ke] = 1;
                    m.Ke++;
                    changed = true;
                }
            } else {
                // add path: append df=1 slots while room remains and we are still under
                // the item's allowed slot count; the first slot of a non-stackable item
                // is clamped to min(heldCount, delta)
                for (int n = 0; delta > n; n++) {
                    if (m.Ke >= 8 || matched <= slotsForItem) {
                        break;
                    }
                    m.Uf[m.Ke] = itemId;
                    m.df[m.Ke] = 1;
                    matched++;
                    m.Ke++;
                    changed = true;
                    if (n == 0 && ClientIOException.itemY[itemId] == 0) {
                        m.df[m.Ke - 1] = Math.min(m.xe[invSlot], delta);
                        break;                                // clean: breaks after first clamp
                    }
                }
            }
        }

        if (!changed) {
            return;
        }

        // send opcode 33 (DUEL_OFFER_ITEM): count + (itemId, qty) pairs
        m.Jh.newPacket(33, 0);
        m.Jh.outBuffer.putByte(m.Ke);                             // obf: Jh.f (Packet.outBuffer = BitBuffer)
        for (int j = 0; j < m.Ke; j++) {
            m.Jh.outBuffer.putShort(m.Uf[j]);
            m.Jh.outBuffer.putInt(m.df[j]);
        }
        m.Jh.finishPacket(21294);
        m.ki = false;
        m.ke = false;
    }
}
