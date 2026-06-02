package client;

import client.net.ClientStream;
import client.net.StringCodec;
import client.world.WorldEntity;

/**
 * ClientPackets — outgoing social / chat-command / privacy packet writers extracted from
 * Mudclient. These are the only truly standalone packet-out methods: each opens a packet on
 * m.Jh (ClientStream), writes args, flushes, and (for friend/ignore) validates against the
 * local friend/ignore arrays.
 *
 * <p>The only Mudclient method they call is showServerMessage (stays in Mudclient; called as
 * m.showServerMessage(...)). No this-identity, no synchronized, no this-as-listener.
 */
class ClientPackets {
    final Mudclient m;

    ClientPackets(Mudclient m) {
        this.m = m;
    }

    /**
     * Generic helper: open a packet with the given opcode, write one (char-table-encoded)
     * string, and flush. Used for the simple "opcode + string" client commands.
     */
    // obf: private final void b(String,int)   [b.b(op, op^216): 2nd newPacket arg is junk]
    void sendOpcodeString(String text, int opcode) {
        m.Jh.newPacket(opcode, 0);
        StringCodec.writeString(m.Jh.outBuffer, text); // obf: u.a(99, Jh.f, var1)
        m.Jh.finishPacket(21294);
    }

    /**
     * Send a chat command (text typed after the "::" prefix).
     * Sends opcode 38 (COMMAND).
     */
    // obf: private final void a(String,int)   [int param var2 is anti-tamper junk]
    void sendCommand(String command, int unused) {
        m.Jh.newPacket(38, 0); // COMMAND
        m.Jh.outBuffer.putString(command); // obf: Jh.f.a(var1, 104)
        m.Jh.finishPacket(21294);
    }

    /**
     * Send a private (player-to-player) chat message.
     * Sends opcode 218 (SOCIAL_SEND_PRIVATE_MESSAGE): recipient username, then the
     * char-table-encoded message body.
     */
    // obf: private final void a(byte,String,String)   [byte param var1 is anti-tamper junk]
    void sendPrivateMessage(byte unused, String recipient, String message) {
        m.Jh.newPacket(218, 0); // SOCIAL_SEND_PRIVATE_MESSAGE
        m.Jh.outBuffer.putString(recipient);                  // obf: Jh.f.a(var2, 124)
        StringCodec.writeString(m.Jh.outBuffer, message);  // obf: u.a(103, Jh.f, var3)
        m.Jh.finishPacket(21294);
    }

    /**
     * Remove a player from the friends list. Drops them from the local list (shifting the
     * parallel friend arrays) and tells the server.
     * Sends opcode 167 (SOCIAL_REMOVE_FRIEND): the un-normalised username.
     */
    // obf: private final void b(String,byte)   [byte param var2 is anti-tamper junk]
    void sendRemoveFriend(String name, byte unused) {
        String wanted = WorldEntity.trimAndValidateString(name, (byte)127); // obf: w.a(var1, ..)  trim & canonicalise
        if (wanted == null) {
            return;
        }
        for (int i = 0; i < m.friendListCount; i++) { // obf: var4 < n.g
            if (wanted.equals(WorldEntity.trimAndValidateString(m.friendListNames[i], (byte)127))) { // obf: ua.h[var4]
                // Remove locally: shift the four parallel friend arrays down over slot i.
                m.friendListCount--;
                for (int j = i; j < m.friendListCount; j++) {
                    m.friendListNames[j]       = m.friendListNames[j + 1]; // obf: ua.h
                    m.friendListFormerNames[j] = m.friendListFormerNames[j + 1]; // obf: cb.c
                    m.friendListWorlds[j]      = m.friendListWorlds[j + 1]; // obf: ac.z
                    m.friendListOnline[j]      = m.friendListOnline[j + 1]; // obf: Fj
                }
                m.Jh.newPacket(167, 0); // SOCIAL_REMOVE_FRIEND
                m.Jh.outBuffer.putString(name); // obf: Jh.f.a(var1, 110)  (raw, un-normalised)
                m.Jh.finishPacket(21294);
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
    void sendAddFriend(int unused, String name) {
        // Friend list cap: 200 for members, 100 otherwise.  obf: ~(Pg?200:100) >= ~n.g
        if (m.friendListCount >= (m.membersServer ? 200 : 100)) {
            m.showServerMessage(false, null, 0, "Your friend list is full", 0, 0, null, null); // il[384]
            return;
        }
        String wanted = WorldEntity.trimAndValidateString(name, (byte)127);
        if (wanted == null) {
            return;
        }
        // Already on the friend list? (match either current or former name)
        for (int i = 0; i < m.friendListCount; i++) {
            if (wanted.equals(WorldEntity.trimAndValidateString(m.friendListNames[i], (byte)127))
                    || (m.friendListFormerNames[i] != null
                        && wanted.equals(WorldEntity.trimAndValidateString(m.friendListFormerNames[i], (byte)127)))) {
                m.showServerMessage(false, null, 0, name + " is already on your friend list", 0, 0, null, null); // il[386]
                return;
            }
        }
        // On the ignore list? (can't be both)  obf: var4 < db.g over l.c / ia.g
        for (int i = 0; i < m.ignoreListCount; i++) {
            if (wanted.equals(WorldEntity.trimAndValidateString(m.ignoreListNames[i], (byte)127))
                    || (m.ignoreListFormerNames[i] != null
                        && wanted.equals(WorldEntity.trimAndValidateString(m.ignoreListFormerNames[i], (byte)127)))) {
                m.showServerMessage(false, null, 0, "Please remove " + name + " from your ignore list first", 0, 0, null, null); // il[251]+name+il[383]
                return;
            }
        }
        // Yourself?  obf: w.a(wi.C, ..)
        if (wanted.equals(WorldEntity.trimAndValidateString(m.wi.name, (byte)127))) {
            m.showServerMessage(false, null, 0, "You can't add yourself to your own friend list", 0, 0, null, null); // il[385]
            return;
        }
        m.Jh.newPacket(195, 0); // SOCIAL_ADD_FRIEND
        m.Jh.outBuffer.putString(name); // obf: Jh.f.a(var2, -23)
        m.Jh.finishPacket(21294);
    }

    /**
     * Add a player to the ignore list after validating it (list-full / duplicate / on friend
     * list / self checks, each surfacing a system message), then notify the server.
     * Sends opcode 132 (SOCIAL_ADD_IGNORE): the username.
     */
    // obf: private final void a(String,byte)   [byte param var2 is anti-tamper junk]
    void sendAddIgnore(String name, byte unused) {
        // Ignore list cap: 100.  obf: ~db.g <= -101  ⟺  db.g >= 100
        if (m.ignoreListCount >= 100) {
            m.showServerMessage(false, null, 0, "Your ignore list is full", 0, 0, null, null); // il[254]
            return;
        }
        String wanted = WorldEntity.trimAndValidateString(name, (byte)127);
        if (wanted == null) {
            return;
        }
        // Already ignored? (match current or former name)  obf: var4 < db.g over l.c / ia.g
        for (int i = 0; i < m.ignoreListCount; i++) {
            if (wanted.equals(WorldEntity.trimAndValidateString(m.ignoreListNames[i], (byte)127))
                    || (m.ignoreListFormerNames[i] != null
                        && wanted.equals(WorldEntity.trimAndValidateString(m.ignoreListFormerNames[i], (byte)127)))) {
                m.showServerMessage(false, null, 0, name + " is already on your ignore list", 0, 0, null, null); // il[252]
                return;
            }
        }
        // On the friend list? (can't be both)  obf: var4 < n.g over ua.h / cb.c
        for (int i = 0; i < m.friendListCount; i++) {
            if (wanted.equals(WorldEntity.trimAndValidateString(m.friendListNames[i], (byte)127))
                    || (m.friendListFormerNames[i] != null
                        && wanted.equals(WorldEntity.trimAndValidateString(m.friendListFormerNames[i], (byte)127)))) {
                m.showServerMessage(false, null, 0, "Please remove " + name + " from your friend list first", 0, 0, null, null); // il[251]+name+il[255]
                return;
            }
        }
        // Yourself?
        if (wanted.equals(WorldEntity.trimAndValidateString(m.wi.name, (byte)127))) {
            m.showServerMessage(false, null, 0, "You can't add yourself to your own ignore list", 0, 0, null, null); // il[253]
            return;
        }
        m.Jh.newPacket(132, 0); // SOCIAL_ADD_IGNORE
        m.Jh.outBuffer.putString(name);
        m.Jh.finishPacket(21294);
    }

    /**
     * Remove a player from the ignore list. Drops them locally (shifting the parallel ignore
     * arrays) and notifies the server.
     * Sends opcode 241 (SOCIAL_REMOVE_IGNORE): the un-normalised username.
     */
    // obf: private final void a(byte,String)   [byte param var1 is anti-tamper junk: if(var1<-7){..whole body}]
    void sendRemoveIgnore(byte unused, String name) {
        String wanted = WorldEntity.trimAndValidateString(name, (byte)127);
        if (wanted == null) {
            return;
        }
        for (int i = 0; i < m.ignoreListCount; i++) { // obf: var4 < db.g
            // NOTE: the match is against ignoreListDisplayNames (ia.a), not ignoreListNames (l.c).
            if (wanted.equals(WorldEntity.trimAndValidateString(m.ignoreListDisplayNames[i], (byte)127))) { // obf: ia.a[var4]
                // Remove locally: shift the FOUR parallel ignore arrays down over slot i.
                m.ignoreListCount--;
                for (int j = i; j < m.ignoreListCount; j++) {
                    m.ignoreListNames[j]        = m.ignoreListNames[j + 1];        // obf: l.c
                    m.ignoreListDisplayNames[j] = m.ignoreListDisplayNames[j + 1]; // obf: ia.a
                    m.ignoreListFormerNames[j]  = m.ignoreListFormerNames[j + 1];  // obf: ia.g
                    // obf: ua.wb[j] = ua.wb[j]  — the original client's shift omits the +1 here
                    // (both Vineflower and CFR agree); reproduced faithfully as a no-op self-assign.
                    m.ignoreListWorlds[j]       = m.ignoreListWorlds[j];           // obf: ua.wb (self-assign, original bug)
                }
                m.Jh.newPacket(241, 0); // SOCIAL_REMOVE_IGNORE
                m.Jh.outBuffer.putString(name); // obf: Jh.f.a(var2, -78)
                m.Jh.finishPacket(21294);
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
    void sendPrivacySettings(int blockTrade, int blockPrivate, int blockChat, int unused, int blockDuel) {
        m.Jh.newPacket(64, 0); // PRIVACY_SETTINGS_CHANGED
        m.Jh.outBuffer.putByte(blockChat);    // obf: Jh.f.c(var3, ..)
        m.Jh.outBuffer.putByte(blockPrivate); // obf: Jh.f.c(var2, ..)
        m.Jh.outBuffer.putByte(blockTrade);   // obf: Jh.f.c(var1, ..)
        m.Jh.outBuffer.putByte(blockDuel);    // obf: Jh.f.c(var5, ..)
        m.Jh.finishPacket(21294);
    }
}
