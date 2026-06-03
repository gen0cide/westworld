package client;

import client.scene.*;     // Scene, GameModel, ImageLoader, SurfaceSprite, Surface, SpriteScaler ...
import client.net.*;       // ClientStream (Jh), BitBuffer (mg), TextEncoder, ISAAC ...
import client.world.*;     // World (Hh), Scene (Ek), GameCharacter, GameModel ...
import client.data.*;      // GameData, DataStore, CacheUpdater, CharTable, NameTable ...
import client.ui.*;        // Panel, FontBuilder, MessageList, CharTable ...
import client.util.*;      // Utility (mb), ErrorHandler (i), ClientIOException, DecodeBuffer ...
import client.shell.*;     // GameShell, GameFrame, InputState ...
import client.audio.*;     // SoundVoice, SoundMixer ...
import client.nativeapi.*; // platform glue

/**
 * IncomingPackets — the master server->client packet dispatch + scene/menu/social/dialog
 * subsystem extracted from Mudclient (de-god extraction 5/5).
 *
 * <p>All instance state lives on Mudclient and is reached through the {@code m} back-reference.
 * Static string pools are {@code Mudclient.il} / {@code Mudclient.STRINGS}. Outgoing packets go
 * through {@code m.packets.*}; sound through {@code m.sound}. Shared callbacks
 * (showServerMessage, getPlayer/getNpc/addPlayer/addNpc, loadRegion, buildEntityModel,
 * onStopGame, requestLogout, drawHelpMenu, the drawBox/drawSprite/drawScrollbar(2) panel
 * wrappers, sortFriendsList, sendConfirmLogout(Ack), resetTradeDuelState, ...) stay in
 * Mudclient and are invoked via {@code m.<method>}.
 *
 * <p>NO this-identity: {@code mg} is a Mudclient field, not {@code this}. The one self-call
 * ({@code onFriendUpdate -> this.handlePacket}) stays on this delegate instance.
 *
 * <p>WIRE-FORMAT CRITICAL: every mg read/getUnsigned/getString read order and bit-width is
 * the protocol-235 stream. The bodies below are byte-for-byte the originals from Mudclient;
 * only the {@code m.} / {@code Mudclient.} qualifiers were rewritten.
 */
class IncomingPackets {
    final Mudclient m;

    IncomingPackets(Mudclient m) {
        this.m = m;
    }

    // =========================================================================
    // ===== packetin =====
    // =========================================================================
    //
    // Incoming server->client packet handling. Opcode numbers cited below are the
    // RSC protocol-235 SEND_* values from OpenRSC's Payload235Generator.
    //
    // IMPORTANT NAMING NOTE (read before trusting the method names):
    // The skeleton's *proposed names* for several methods in this group are inaccurate
    // (the bodies only decompiled in the CFR/clean base). To keep cross-method call sites
    // resolvable by the orchestrator, the skeleton's proposed name is kept as the Java
    // method name, but each doc comment states what the method ACTUALLY does:
    //   * handlePacket          (b(int,byte,int))  IS the master server->client dispatch.   [correct]
    //   * handleSceneUpdates    (b(boolean,int))   is really the right-click MENU-ACTION
    //                                              dispatcher (menuItemClick): it turns the
    //                                              selected context-menu entry into an
    //                                              OUTGOING action packet. (misnamed)
    //   * onFriendUpdate        (a(int,int,int))   is the social/private-message packet
    //                                              sub-dispatcher; it handles the social
    //                                              opcodes and delegates everything else to
    //                                              handlePacket. (broadly correct)
    //   * applyAppearanceUpdate (a(boolean,boolean)) is really the social-entry DIALOG
    //                                              renderer (add-friend / add-ignore /
    //                                              send-message list). It parses no packet. (misnamed)
    //
    // CORRECTNESS-AUDIT NOTES (vs the OLD part written against the defective base — the
    // clean Vineflower base at decompiled/normalized-clean/client.java is now ground truth):
    //   * The region-stream opcodes were re-decoded from the clean base. The OLD part had
    //     the 48/91/99 handler->array mapping SCRAMBLED. Truth:
    //         opcode 48 -> SCENERY     (eh, hg[], Se/ye/vc/bg)         [SEND_SCENERY_HANDLER]
    //         opcode 91 -> GROUND ITEM (hf, rd[], Jd/yk/Hj/Ng)         [SEND_GROUND_ITEM_HANDLER]
    //         opcode 99 -> BOUNDARY    (Ah, Zf/Ni/Gj/Le)              [SEND_BOUNDARY_HANDLER]
    //   * opcode 5  -> PRAYERS-ACTIVE (fi[])   and opcode 206 -> QUEST flags (bk[])
    //     were SWAPPED in the OLD part. The clean base proves 5=prayers, 206=quests.
    //   * the inventory "stackable" test in opcodes 53 and 90 was INVERTED: the count is
    //     read when fa.e[id] == 0 (stackable), not != 0.
    //   * opcode 104 (NPC update) reads from the `te` cache, not `We` (the two entity
    //     caches are role-swapped vs the skeleton's guess: opcode 234 player-update uses
    //     `We`, opcode 104 npc-update uses `te`).
    //   * opcode 30 reads four bytes; ALL four flags are (byte == 1) — the OLD part had
    //     fd/Yi as (==2).
    //   * opcode 234 has SEVEN sub-types (0..6), decoded below; the OLD part only had 0..3
    //     with the wrong semantics.
    //   * opcode 211 fully re-culls walls+scenery+ground-items; the OLD part stubbed it.
    //
    // Stripped per instructions: `boolean var17 = client.vh;` opaque predicate and all the
    // dead `if(var17)`/`if(!var17)`/`break`/`continue` control flow it gates; `++<counter>;`
    // profiling bumps; the `~x` sign-test idiom (rewritten to plain comparisons); junk
    // shift masks. The bit-stream reader `incomingPacket` (obf `mg`, type ja/BitBuffer) is
    // read via:
    //   .a((byte)104)  -> read one unsigned byte
    //   .f(255)        -> read one unsigned short (16 bits)
    //   .h(20869)      -> read one unsigned short (returned as byte/int)
    //   .b(-129)       -> read one signed 16-bit
    //   .c((byte)-44)  -> read a zero-terminated string
    //   .f(bias, nBits)-> read nBits as an unsigned bitfield
    //   .c(103)        -> read a var-length int (1 or 4 bytes by high bit)
    //   .g(0)          -> read an 8-byte long
    //   .w            -> the byte cursor; .k(...) -> the bit cursor; .i/.j -> align/finalize.

    /**
     * MASTER server->client packet dispatch (protocol 235). Reads the already-buffered
     * packet (opcode in {@code opcode}, body length in {@code length}) from
     * {@code incomingPacket} (obf mg) and applies it to game state. Roughly 50 opcodes
     * plus four bulk region streams (players/walls/scenery/ground-items/npcs) decode here.
     *
     * On any RuntimeException the original logs the packet bytes and forces a clean
     * disconnect ({@code onStopGame(true)}); that recovery path is preserved.
     *
     * obf: void b(int,byte,int)   params: (opcode, <anti-tamper dummy byte>, length)
     */
    void handlePacket(int opcode, byte unused, int length) {
        try {
            try {

                // ---- 191 SEND_PLAYER_COORDS: local player position + nearby-player movement stream ----
                if (opcode == 191) {
                    // double-buffer the in-view player list (this tick <- last tick)
                    m.If = m.Yc;                          // playersThisTick count = previous view count
                    for (int i = 0; i < m.If; i++) {
                        m.Zg[i] = m.rg[i];                // players[] <- playersLast[]
                    }
                    m.mg.initBitAccess();                           // align bit reader to byte boundary
                    m.Lf = m.mg.readBits(11);              // region origin X (11 bits)   (obf: Lf = localRegionX)
                    m.sh = m.mg.readBits(13);              // region origin Y (13 bits)   (obf: sh = localRegionY)
                    int localAnim = m.mg.readBits(4);          // local player facing/anim (4 bits)
                    boolean regionChanged = m.loadRegion(m.sh, m.Lf, false); // obf: a(sh,Lf,false)
                    m.Lf -= m.Qg;                         // subtract scene origin -> tile-local coords
                    m.sh -= m.zg;
                    int worldX = m.Lf * m.Ug + 64;        // Ug = tile size in world units; +64 = tile centre
                    int worldY = m.sh * m.Ug + 64;
                    m.Yc = 0;                                // reset this-tick player count
                    if (regionChanged) {                        // snap camera-follow target to new position
                        m.wi.stepCount = 0;
                        m.wi.waypointCurrent = 0;
                        m.wi.currentX = m.wi.waypointsX[0] = worldX;
                        m.wi.currentY = m.wi.waypointsY[0] = worldY;
                    }
                    // (re)create the local player entity at the new tile.
                    // NB: the player-create method is obf `d(int,int,int,int,int)`, which
                    // scene.part.java named `addNpc` (the scene group's addPlayer/addNpc names
                    // are swapped vs behaviour); call it by obf-signature, not the misleading name.
                    m.wi = m.addNpc(worldY, m.Zc, worldX, -56, localAnim); // obf: d(worldY,Zc,worldX,-56,anim)
                    int otherPlayers = m.mg.readBits(8);       // count of other visible players (8 bits)

                    // --- movement deltas for the players already in view ---
                    for (int p = 0; p < otherPlayers; p++) {
                        GameCharacter player = m.Zg[p + 1];
                        int hasUpdate = m.mg.readBits(1);     // 1 bit: did this player move/turn?
                        if (hasUpdate != 0) {
                            int moved = m.mg.readBits(1);      // 1 bit: walked (1) vs. only-turned (0)
                            if (moved == 0) {                   // only changed facing direction
                                int dir = m.mg.readBits(2);    // 2 bits
                                if (dir == 3) continue;          // 3 = "removed from view" (skip carry-over)
                                player.animationNext = (dir << 2) + m.mg.readBits(2); // pack facing: hi 2 bits + lo 2 bits
                            } else {
                                // walked one tile in one of 8 compass directions
                                int dir = m.mg.readBits(3);
                                int wp = player.waypointCurrent;              // current waypoint slot
                                int wx = player.waypointsX[wp];
                                int wy = player.waypointsY[wp];
                                if (dir == 2 || dir == 1 || dir == 3) wx += m.Ug;  // E component
                                if (dir == 6 || dir == 5 || dir == 7) wx -= m.Ug;  // W component
                                if (dir == 4 || dir == 3 || dir == 5) wy += m.Ug;  // S component
                                player.waypointCurrent = wp = (wp + 1) % 10;  // advance ring of 10 waypoints
                                player.animationNext = dir;
                                if (dir == 0 || dir == 1 || dir == 7) wy -= m.Ug;  // hasPainted component
                                player.waypointsX[wp] = wx;
                                player.waypointsY[wp] = wy;
                            }
                        }
                        m.rg[m.Yc++] = player;             // carry the (possibly updated) player to this tick
                    }

                    // --- newly-appeared players (absolute coords) ---
                    while (m.mg.getBitPosition() + 24 < length * 8) {  // while bits remain in this packet
                        int serverIndex = m.mg.readBits(11);     // 11-bit player server index
                        int dx = m.mg.readBits(5);                // signed 5-bit X offset from local
                        if (dx > 15) dx -= 32;
                        int dy = m.mg.readBits(5);
                        if (dy > 15) dy -= 32;
                        int anim = m.mg.readBits(4);
                        int ay = (dy + m.sh) * m.Ug + 64;
                        int ax = (m.Lf + dx) * m.Ug + 64;
                        m.addNpc(ay, serverIndex, ax, -112, anim);  // obf: d(ay,idx,ax,-112,anim) — player-create
                    }
                    m.mg.finishBitAccess();                              // finalize bit reader
                    return;
                }

                // ---- 99 SEND_BOUNDARY_HANDLER: add/remove wall (boundary) models for this region ----
                if (opcode == 99) {
                    while (m.mg.offset < length) {                   // walk packet payload (byte cursor mg.w)
                        if (m.mg.getUnsignedByte() == 255) {         // removal run marker (consumed; jar does NOT un-read it — client.java:14471-14473)
                            // remove walls at this region-tile anchor; compact the arrays.
                            int anchorX = m.Lf + m.mg.readRawByte() >> 3;
                            int anchorY = m.sh + m.mg.readRawByte() >> 3;
                            int kept = 0;
                            for (int w = 0; w < m.Ah; w++) {    // Ah = active wall count
                                int rx = (m.Zf[w] >> 3) - anchorX;
                                int ry = (m.Ni[w] >> 3) - anchorY;
                                if (rx != 0 || ry != 0) {          // keep -> compact down
                                    if (kept != w) {
                                        m.Zf[kept] = m.Zf[w];
                                        m.Ni[kept] = m.Ni[w];
                                        m.Gj[kept] = m.Gj[w];
                                        m.Le[kept] = m.Le[w];
                                    }
                                    kept++;
                                }
                            }
                            m.Ah = kept;
                            continue;
                        }
                        // --- add (or single-remove) a wall/boundary ---
                        m.mg.offset--;
                        int dir = m.mg.getUnsignedShort();                  // wall direction/type (bit15 = remove flag)
                        int x = m.Lf + m.mg.readRawByte();
                        int y = m.sh + m.mg.readRawByte();
                        if ((dir & 0x8000) != 0) {                 // high bit set -> "remove this wall"
                            dir &= 0x7FFF;
                            int kept = 0;
                            for (int w = 0; w < m.Ah; w++) {
                                if (m.Zf[w] == x && m.Ni[w] == y && m.Gj[w] == dir) {
                                    // matched -> drop (do not copy through)
                                } else {
                                    if (kept != w) {
                                        m.Zf[kept] = m.Zf[w];
                                        m.Ni[kept] = m.Ni[w];
                                        m.Gj[kept] = m.Gj[w];
                                        m.Le[kept] = m.Le[w];
                                    }
                                    kept++;
                                }
                            }
                            m.Ah = kept;
                        } else {                                   // add
                            m.Zf[m.Ah] = x;
                            m.Ni[m.Ah] = y;
                            m.Gj[m.Ah] = dir;
                            m.Le[m.Ah] = 0;
                            // inherit lighting from a scenery model sharing this tile, if any
                            for (int s = 0; s < m.eh; s++) {
                                if (m.Se[s] == x && m.ye[s] == y) {
                                    m.Le[m.Ah] = TextEncoder.scratchIntArray1[m.vc[s]];
                                    break;
                                }
                            }
                            m.Ah++;
                        }
                    }
                    return;
                }

                // ---- 48 SEND_SCENERY_HANDLER: add/remove scene objects (trees/signs/fences/door-objects) ----
                // jar: client.java:14510-14659 (encoded `if(-49==~var1)`); sits between 99 and 111.
                if (opcode == 48) {
                    while (length > m.mg.offset) {                   // walk packet payload (byte cursor mg.w)
                        if (m.mg.getUnsignedByte() != 255) {         // peeked marker is NOT a removal run
                            m.mg.offset--;                           // un-read the peeked byte
                            int objType = m.mg.getUnsignedShort();          // scene-object type (15-bit; 60000 = pure remove)
                            int x = m.Lf + m.mg.readRawByte();
                            int y = m.sh + m.mg.readRawByte();
                            // remove any existing scenery on this exact tile (de-dup), compacting the arrays
                            int kept = 0;
                            for (int s = 0; s < m.eh; s++) {    // eh = active scenery count
                                if (m.Se[s] != x || m.ye[s] != y) {   // keep -> compact down
                                    if (s != kept) {
                                        m.hg[kept] = m.hg[s];
                                        m.hg[kept].key = kept;
                                        m.Se[kept] = m.Se[s];
                                        m.ye[kept] = m.ye[s];
                                        m.vc[kept] = m.vc[s];
                                        m.bg[kept] = m.bg[s];
                                    }
                                    kept++;
                                } else {                           // matched -> drop from scene + world
                                    m.Ek.removeModel(m.hg[s]);
                                    m.Hh.removeObject(m.vc[s], m.Se[s], m.ye[s], 4081);
                                }
                            }
                            m.eh = kept;
                            if (objType != 60000) {                // add gate (jar: 60000 != objType)
                                int orient = m.Hh.getTileDirection(x, y);  // tile orientation code
                                int dimW, dimH;
                                if (orient != 0 && orient != 4) {
                                    dimW = NameTable.sortKeys[objType];
                                    dimH = RecordLoader.intArray[objType];
                                } else {
                                    dimH = NameTable.sortKeys[objType];
                                    dimW = RecordLoader.intArray[objType];
                                }
                                int midX = m.Ug * (x + (x + dimW)) / 2;
                                int midZ = (y + (y + dimH)) * m.Ug / 2;
                                int modelIdx = SurfaceImageProducer.entityIndexTableF[objType];  // obf fb.f -> entityIndexTableF
                                GameModel model = m.objectModels[modelIdx].copy(-2);  // obf ca model; obf kh -> objectModels; clone base
                                m.Ek.addModel(model);       // add to scene
                                model.key = m.eh;           // obf .rb -> key
                                model.rotate(0, -31616, orient * 32, 0); // orient the model (yaw = orient*32)
                                model.translate(midX, midZ, -m.Hh.getElevation(midX, midZ), true); // drop to terrain height
                                model.setLight(-50, 48, -10, -50, true, 48, 117); // lighting/colour defaults
                                m.Hh.removeObject2(x, objType, false, y);   // place object in world
                                if (74 == objType) {               // special case: floats 480 up
                                    model.translate(0, 0, -480, true);
                                }
                                m.Se[m.eh] = x;
                                m.ye[m.eh] = y;
                                m.vc[m.eh] = objType;
                                m.bg[m.eh] = orient;
                                m.hg[m.eh++] = model;
                            }
                        } else {                                   // marker == 255 -> removal run for a tile anchor
                            int anchorX = m.Lf + m.mg.readRawByte() >> 3;
                            int anchorY = m.sh + m.mg.readRawByte() >> 3;
                            int kept = 0;
                            for (int s = 0; s < m.eh; s++) {
                                int rx = (m.Se[s] >> 3) - anchorX;
                                int ry = (m.ye[s] >> 3) - anchorY;
                                if (rx == 0 && ry == 0) {          // matched -> drop from scene + world
                                    m.Ek.removeModel(m.hg[s]);
                                    m.Hh.removeObject(m.vc[s], m.Se[s], m.ye[s], 4081);
                                } else {                           // keep -> compact down
                                    if (s != kept) {
                                        m.hg[kept] = m.hg[s];
                                        m.hg[kept].key = kept;
                                        m.Se[kept] = m.Se[s];
                                        m.ye[kept] = m.ye[s];
                                        m.vc[kept] = m.vc[s];
                                        m.bg[kept] = m.bg[s];
                                    }
                                    kept++;
                                }
                            }
                            m.eh = kept;
                        }
                    }
                    return;
                }

                // ---- 111 SEND_TRADE_OPEN_CONFIRM... here: tutorial flag (Kd) ----
                if (opcode == 111) {
                    m.Kd = m.mg.getUnsignedByte() != 0;
                    return;
                }

                // ---- 53 SEND_INVENTORY: full inventory contents ----
                if (opcode == 53) {
                    m.lc = m.mg.getUnsignedByte();                // item count
                    for (int i = 0; i < m.lc; i++) {
                        int raw = m.mg.getUnsignedShort();                  // bit15 = wielded, bits0..14 = item id
                        m.vf[i] = StreamBase.bitwiseAnd(raw, 32767);             // inventoryItems[i] = raw & 0x7FFF
                        m.Aj[i] = raw / 32768;                  // inventoryEquipped[i] = (raw >> 15)
                        if (ClientIOException.itemY[raw & 32767] == 0) {              // obf fa.e -> itemY; ==0 => stackable -> read a quantity
                            m.xe[i] = m.mg.getSmartSigned();           // inventoryCount[i] (var-length int)
                        } else {
                            m.xe[i] = 1;
                        }
                    }
                    return;
                }

                // ---- 234 SEND_UPDATE_PLAYERS: per-player update stream ----
                // Sub-type cascade (read from `We` player cache, index `idx`):
                //   0 = bubble holding an item        4 = projectile (id+target)
                //   1 = chat message (ignore-checked) 5 = full appearance/equipment
                //   2 = combat (damage+hits)          6 = self speech (local player)
                //   3 = projectile (alt id+target)    7 = no-op
                if (opcode == 234) {
                    int count = m.mg.getUnsignedShort();
                    for (int u = 0; u < count; u++) {
                        int idx = m.mg.getUnsignedShort();                  // player server index
                        GameCharacter player = m.We[idx];                  // obf: We = player cache (this client)
                        int type = m.mg.readRawByte();               // update sub-type
                        if (type == 1) {                           // chat message
                            // clean base gates the whole block on a visible player (null -> skip,
                            // reading nothing else); the server only sends type-1 for known players.
                            if (player != null) {
                                int icon = m.mg.getUnsignedByte();
                                String message = SpriteScaler.readPacketString(m.mg, false); // decode scrambled chat
                                boolean ignored = false;
                                String hashed = WorldEntity.trimAndValidateString(player.message, (byte)109); // name -> ignore-hash
                                if (hashed != null) {
                                    for (int i = 0; i < LinkedQueue.DEAD_G; i++) {   // db.g = ignore count
                                        if (hashed.equals(WorldEntity.trimAndValidateString(SpriteScaler.playerNames[i], (byte)100))) { ignored = true; break; }
                                    }
                                }
                                if (!ignored) {
                                    player.messageTimeout = 150;
                                    player.name = message;
                                    m.showServerMessage(icon == 2, player.chatSenderName, 0, player.name, 4, icon, player.message, null);
                                }
                            }
                        } else if (type == 3) {                    // projectile (id + target index, 16-bit)
                            int sprite = m.mg.getUnsignedShort();
                            int target = m.mg.getUnsignedShort();
                            if (player != null) {
                                player.attackingNpcServerIndex = target;
                                player.projectileRange = m.nc;
                                player.attackingPlayerServerIndex = -1;
                                player.incomingProjectileSprite = sprite;
                            }
                        } else if (type == 5) {                    // full appearance / equipment update
                            if (player == null) {                  // not in view -> skip the block
                                m.mg.getUnsignedShort();
                                m.mg.getString();
                                m.mg.getString();
                                int n = m.mg.getUnsignedByte();
                                m.mg.offset += 6 + n;
                            } else {
                                m.mg.getUnsignedShort();                    // (server index echo, discarded)
                                player.chatSenderName = m.mg.getString();   // name
                                player.message = m.mg.getString();   // formatted name
                                int n = m.mg.getUnsignedByte();      // equipment slot count
                                int s = 0;
                                for (; s < n; s++) player.equippedItem[s] = m.mg.getUnsignedByte();
                                for (; s < 12; s++) player.equippedItem[s] = 0;
                                player.colourHair = m.mg.getUnsignedByte();   // hair colour
                                player.colourTop = m.mg.getUnsignedByte();   // top colour
                                player.npcIdOrColourBottom = m.mg.getUnsignedByte();   // trouser colour
                                player.colourSkin = m.mg.getUnsignedByte();   // skin colour
                                player.level = m.mg.getUnsignedByte();   // combat level
                                player.skullVisible = m.mg.getUnsignedByte();   // skull/icon
                            }
                        } else if (type == 6) {                    // self speech (local player only)
                            // jar reads the scrambled string ONLY when the slot is populated
                            // (client.java:14833-14837: `if (type != 6 || player == null) break;`);
                            // a null slot consumes nothing, keeping the 234 stream in sync.
                            if (player != null) {
                                String message = SpriteScaler.readPacketString(m.mg, false);
                                player.name = message;
                                player.messageTimeout = 150;
                                if (m.wi == player) {
                                    m.showServerMessage(false, player.chatSenderName, 0, player.name, 3, 0, player.message, null);
                                }
                            }
                        } else if (type == 4) {                    // projectile (alt id + target, 16-bit)
                            int sprite = m.mg.getUnsignedShort();
                            int target = m.mg.getUnsignedShort();
                            if (player != null) {
                                player.projectileRange = m.nc;
                                player.attackingNpcServerIndex = -1;
                                player.attackingPlayerServerIndex = target;
                                player.incomingProjectileSprite = sprite;
                            }
                        } else if (type == 2) {                    // combat: damage + current/max hits
                            int damage = m.mg.getUnsignedByte();
                            int curHits = m.mg.getUnsignedByte();
                            int maxHits = m.mg.getUnsignedByte();
                            if (player != null) {
                                player.healthMax = maxHits;
                                player.healthCurrent = curHits;
                                player.damageTaken = damage;
                                if (m.wi == player) {           // local player took damage
                                    m.oh[3] = curHits;          // skillCurrent[Hits]
                                    m.cg[3] = maxHits;          // skillBase[Hits]
                                    m.mh = false;               // close any open dialog box
                                    m.Oh = false;
                                }
                                player.combatTimer = 200;                    // combat timer
                            }
                        } else if (type == 0) {                    // bubble holding an item
                            int itemId = m.mg.getUnsignedShort();
                            if (player != null) {
                                player.bubbleTimeout = 150;
                                player.bubbleItem = itemId;
                            }
                        }
                        // type == 7: no-op (server index/data consumed nowhere)
                    }
                    return;
                }

                // ---- 91 SEND_GROUND_ITEM_HANDLER: add/remove ground items for this region ----
                if (opcode == 91) {
                    while (length > m.mg.offset) {
                        if (m.mg.getUnsignedByte() == 255) {         // removal run
                            int anchorX = m.Lf + m.mg.readRawByte() >> 3;
                            int anchorY = m.sh + m.mg.readRawByte() >> 3;
                            int kept = 0;
                            for (int g = 0; g < m.hf; g++) {    // hf = active ground-item count
                                int rx = (m.Jd[g] >> 3) - anchorX;
                                int ry = (m.yk[g] >> 3) - anchorY;
                                if (rx != 0 || ry != 0) {          // keep -> compact
                                    if (kept != g) {
                                        m.rd[kept] = m.rd[g];
                                        m.rd[kept].key = kept + 10000;
                                        m.Jd[kept] = m.Jd[g];
                                        m.yk[kept] = m.yk[g];
                                        m.Hj[kept] = m.Hj[g];
                                        m.Ng[kept] = m.Ng[g];
                                    }
                                    kept++;
                                } else {                           // removed
                                    m.Ek.removeModel(m.rd[g]);
                                    m.Hh.clearWallObjectAdjacency(true, m.Hj[g], m.yk[g], m.Jd[g], m.Ng[g]);
                                }
                            }
                            m.hf = kept;
                            continue;
                        }
                        // --- add or single-remove a ground item ---
                        m.mg.offset--;
                        int itemId = m.mg.getUnsignedShort();
                        int x = m.Lf + m.mg.readRawByte();
                        int y = m.sh + m.mg.readRawByte();
                        int dir = m.mg.readRawByte();
                        boolean placed = false;
                        int kept = 0;
                        for (int g = 0; g < m.hf; g++) {        // remove a matching item if present
                            if (m.Jd[g] != x || m.yk[g] != y || m.Hj[g] != dir) {
                                if (kept != g) {
                                    m.rd[kept] = m.rd[g];
                                    m.rd[kept].key = kept + 10000;
                                    m.Jd[kept] = m.Jd[g];
                                    m.yk[kept] = m.yk[g];
                                    m.Hj[kept] = m.Hj[g];
                                    m.Ng[kept] = m.Ng[g];
                                }
                                kept++;
                            } else {
                                placed = true;                     // sentinel: this item already existed
                                m.Ek.removeModel(m.rd[g]);
                                m.Hh.clearWallObjectAdjacency(true, m.Hj[g], m.yk[g], m.Jd[g], m.Ng[g]);
                            }
                        }
                        m.hf = kept;
                        if (itemId != 65535) {                     // jar gate: add unless pure-remove sentinel (client.java:15009-15022)
                            m.Hh.setWallObjectAdjacency(y, itemId, dir, x, 11715);   // scene.placeGroundItem
                            m.rd[m.hf] = m.buildEntityModel(true, y, itemId, x, dir, m.hf);
                            m.Jd[m.hf] = x;
                            m.yk[m.hf] = y;
                            m.Ng[m.hf] = itemId;
                            m.Hj[m.hf++] = dir;
                        }
                    }
                    return;
                }

                // ---- 79 SEND_NPC_COORDS: nearby-NPC movement stream ----
                if (opcode == 79) {
                    // double-buffer the in-view NPC list
                    m.qj = m.de;                             // npcsLastCount = previous view count
                    m.de = 0;
                    for (int i = 0; i < m.qj; i++) {
                        m.Ff[i] = m.Tb[i];                   // npcs[] <- npcsLast[]
                    }
                    m.mg.initBitAccess();                              // align reader
                    int inView = m.mg.readBits(8);                // count of NPCs already in view

                    // --- movement deltas for NPCs in view ---
                    for (int n = 0; n < inView; n++) {
                        GameCharacter npc = m.Ff[n];
                        int hasUpdate = m.mg.readBits(1);
                        if (hasUpdate != 0) {
                            int moved = m.mg.readBits(1);
                            if (moved == 0) {                      // walked one tile
                                int dir = m.mg.readBits(3);
                                int wp = npc.waypointCurrent;
                                int wx = npc.waypointsX[wp];
                                int wy = npc.waypointsY[wp];
                                if (dir == 2 || dir == 1 || dir == 3) wx += m.Ug;
                                if (dir == 6 || dir == 5 || dir == 7) wx -= m.Ug;
                                if (dir == 4 || dir == 3 || dir == 5) wy += m.Ug;
                                npc.animationNext = dir;
                                npc.waypointCurrent = wp = (wp + 1) % 10;
                                if (dir == 0 || dir == 1 || dir == 7) wy -= m.Ug;
                                npc.waypointsX[wp] = wx;
                                npc.waypointsY[wp] = wy;
                            } else {                               // only turned
                                int dir = m.mg.readBits(2);
                                if (dir == 3) continue;            // removed from view
                                npc.animationNext = m.mg.readBits(2) + (dir << 2);
                            }
                        }
                        m.Tb[m.de++] = npc;
                    }

                    // --- newly-appeared NPCs (absolute coords + type) ---
                    while (m.mg.getBitPosition() + 34 < length * 8) {
                        int serverIndex = m.mg.readBits(12);     // 12-bit npc server index
                        int dx = m.mg.readBits(5);
                        if (dx > 15) dx -= 32;
                        int dy = m.mg.readBits(5);
                        if (dy > 15) dy -= 32;
                        int anim = m.mg.readBits(4);
                        int ax = (dx + m.Lf) * m.Ug + 64;
                        int ay = m.Ug * (m.sh + dy) + 64;
                        int npcTypeId = m.mg.readBits(10);       // 10-bit NPC type id
                        if (npcTypeId >= ClientRuntimeException.intCounter) npcTypeId = 24;     // clamp to valid range (la.d = npc-def count)
                        // NB: the npc-create method is obf `a(int,int,int,byte,int,int)`, which
                        // scene.part.java named `addPlayer` (names swapped vs behaviour) — call by obf-signature.
                        m.addPlayer(anim, npcTypeId, ax, (byte)127, ay, serverIndex); // obf: a(anim,type,ax,127,ay,idx) — npc-create
                    }
                    m.mg.finishBitAccess();
                    return;
                }

                // ---- 104 SEND_UPDATE_NPC: per-NPC update stream (chat / combat) ----
                if (opcode == 104) {
                    int count = m.mg.getUnsignedShort();
                    for (int u = 0; u < count; u++) {
                        int idx = m.mg.getUnsignedShort();
                        GameCharacter npc = m.te[idx];                     // obf: te = npc cache (this client)
                        int type = m.mg.getUnsignedByte();
                        if (type == 1) {                           // NPC said something
                            int speakerIdx = m.mg.getUnsignedShort();       // who it spoke to (for filtering)
                            if (npc != null) {
                                String message = SpriteScaler.readPacketString(m.mg, false); // decode scrambled chat
                                npc.messageTimeout = 150;                       // message timeout
                                npc.name = message;
                                if (m.wi.serverIndex == speakerIdx) {     // spoken to the local player
                                    m.showServerMessage(false, null, 0,
                                        GameShell.equipMb[npc.serverId] + Mudclient.il[12] + npc.name, 3, 0, null, Mudclient.il[20]); // obf e.Mb -> equipMb; "<npcName>: <msg>"
                                }
                            }
                        } else if (type == 2) {                    // NPC combat: damage + current/max hits
                            int damage = m.mg.getUnsignedByte();
                            int curHits = m.mg.getUnsignedByte();
                            int maxHits = m.mg.getUnsignedByte();
                            if (npc != null) {
                                npc.damageTaken = damage;
                                npc.healthMax = maxHits;
                                npc.combatTimer = 200;
                                npc.healthCurrent = curHits;
                            }
                        }
                    }
                    return;
                }

                // ---- 245 SEND_OPTIONS_MENU_OPEN: an in-game multiple-choice question dialog ----
                if (opcode == 245) {
                    m.Ph = true;                               // options menu visible
                    int n = m.Id = m.mg.getUnsignedByte();       // number of options
                    for (int i = 0; i < n; i++) {
                        m.ah[i] = m.mg.getString();         // option text
                    }
                    return;
                }

                // ---- 252 SEND_OPTIONS_MENU_CLOSE ----
                if (opcode == 252) {
                    m.Ph = false;
                    return;
                }

                // ---- 25 SEND_WORLD_INFO: world/membership/region metadata at login ----
                if (opcode == 25) {
                    m.Ub = true;
                    m.Zc = m.mg.getUnsignedShort();                     // local player server index
                    m.Ki = m.mg.getUnsignedShort();                     // plane width
                    m.sk = m.mg.getUnsignedShort();                     // plane index base
                    m.bc = m.mg.getUnsignedShort();                     // plane height
                    m.rc = m.mg.getUnsignedShort();                     // planes-per-region
                    m.sk -= m.bc * m.rc;                 // compute origin plane
                    return;
                }

                // ---- 156 SEND_STATS: all skill levels + xp + quest points ----
                if (opcode == 156) {
                    for (int s = 0; s < 18; s++) m.oh[s] = m.mg.getUnsignedByte(); // skillCurrent
                    for (int s = 0; s < 18; s++) m.cg[s] = m.mg.getUnsignedByte(); // skillBase
                    for (int s = 0; s < 18; s++) m.Ak[s] = m.mg.getInt();      // skillXp (signed 16-bit)
                    m.ii = m.mg.getUnsignedByte();               // quest points
                    return;
                }

                // ---- 153 SEND_EQUIPMENT_STATS: armour/weapon aim/power bonuses ----
                if (opcode == 153) {
                    for (int i = 0; i < 5; i++) m.Fc[i] = m.mg.getUnsignedByte();
                    return;
                }

                // ---- 83 SEND_DEATH: player died ----
                if (opcode == 83) {
                    m.rk = 250;                                // death animation timer
                    return;
                }

                // ---- 211 SEND_REMOVE_WORLD_ENTITY: bulk re-cull of walls + scenery + ground items ----
                // For each of (length-1)/4 anchor tiles the client re-derives which boundary,
                // scenery and ground-item models are still in view and compacts the parallel
                // arrays. Entries whose region-tile delta from the anchor is (0,0) are removed.
                if (opcode == 211) {
                    int count = (length - 1) / 4;
                    for (int u = 0; u < count; u++) {
                        int anchorX = m.Lf + m.mg.getSignedShort() >> 3; // obf: mg.a(false) = read short
                        int anchorY = m.sh + m.mg.getSignedShort() >> 3;
                        // walls
                        int kept = 0;
                        for (int w = 0; w < m.Ah; w++) {
                            int rx = (m.Zf[w] >> 3) - anchorX;
                            int ry = (m.Ni[w] >> 3) - anchorY;
                            if (rx != 0 || ry != 0) {
                                if (kept != w) {
                                    m.Zf[kept] = m.Zf[w];
                                    m.Ni[kept] = m.Ni[w];
                                    m.Gj[kept] = m.Gj[w];
                                    m.Le[kept] = m.Le[w];
                                }
                                kept++;
                            }
                        }
                        m.Ah = kept;
                        // scenery
                        kept = 0;
                        for (int s = 0; s < m.eh; s++) {
                            int rx = (m.Se[s] >> 3) - anchorX;
                            int ry = (m.ye[s] >> 3) - anchorY;
                            if (rx != 0 || ry != 0) {
                                if (kept != s) {
                                    m.hg[kept] = m.hg[s];
                                    m.hg[kept].key = kept;
                                    m.Se[kept] = m.Se[s];
                                    m.ye[kept] = m.ye[s];
                                    m.vc[kept] = m.vc[s];
                                    m.bg[kept] = m.bg[s];
                                }
                                kept++;
                            } else {
                                m.Ek.removeModel(m.hg[s]);
                                m.Hh.removeObject(m.vc[s], m.Se[s], m.ye[s], 4081);
                            }
                        }
                        m.eh = kept;
                        // ground items
                        kept = 0;
                        for (int g = 0; g < m.hf; g++) {
                            int rx = (m.Jd[g] >> 3) - anchorX;
                            int ry = (m.yk[g] >> 3) - anchorY;
                            if (rx == 0 && ry == 0) {
                                m.Ek.removeModel(m.rd[g]);
                                m.Hh.clearWallObjectAdjacency(true, m.Hj[g], m.yk[g], m.Jd[g], m.Ng[g]);
                            } else {
                                if (kept != g) {
                                    m.rd[kept] = m.rd[g];
                                    m.rd[kept].key = kept + 10000;
                                    m.Jd[kept] = m.Jd[g];
                                    m.yk[kept] = m.yk[g];
                                    m.Hj[kept] = m.Hj[g];
                                    m.Ng[kept] = m.Ng[g];
                                }
                                kept++;
                            }
                        }
                        m.hf = kept;
                    }
                    return;
                }

                // ---- 59 SEND_APPEARANCE_SCREEN: open "design your character" editor ----
                if (opcode == 59) {
                    m.Kg = true;                               // show char-design screen
                    return;
                }

                // ---- 92: open the DUEL window (Hk=duel-open) ----
                if (opcode == 92) {
                    int idx = m.mg.getUnsignedShort();
                    if (m.We[idx] != null) m.cj = m.We[idx].chatSenderName; // obf .c -> chatSenderName; opponent name
                    m.Hk = true;                               // duel window open
                    m.Lk = 0;                                  // their stake count
                    m.mf = 0;
                    m.Mi = false;                              // their-accepted
                    m.md = false;                              // your-accepted
                    return;
                }

                // ---- 128: close shop + duel windows (Xj/Hk) ----
                if (opcode == 128) {
                    m.Xj = false;
                    m.Hk = false;
                    return;
                }

                // ---- 97: the opponent's DUEL stake (zj/Dd) ----
                if (opcode == 97) {
                    m.Lk = m.mg.getUnsignedByte();               // their stake count
                    for (int i = 0; i < m.Lk; i++) {
                        m.zj[i] = m.mg.getUnsignedShort();              // item id
                        m.Dd[i] = m.mg.getInt();            // amount
                    }
                    m.md = false;                              // reset accepted flags (stake changed)
                    m.Mi = false;
                    return;
                }

                // ---- 162: DUEL your-accepted flag (md) ----
                if (opcode == 162) {
                    m.md = m.mg.getUnsignedByte() == 1;
                    return;
                }

                // ---- 101: searchable item-list / bank-search tab (Rj/Jf/vi) ----
                if (opcode == 101) {
                    m.uk = true;
                    int n = m.mg.getUnsignedByte();                 // entry count
                    int mode = m.mg.readRawByte();                  // 1 = also append owned-but-missing items
                    m.Nh = m.mg.getUnsignedByte();
                    m.xk = m.mg.getUnsignedByte();
                    m.Pf = m.mg.getUnsignedByte();
                    for (int i = 0; i < 40; i++) m.Rj[i] = -1; // clear all 40 slots
                    int slot = 0;
                    for (; slot < n; slot++) {
                        m.Rj[slot] = m.mg.getUnsignedShort();
                        m.Jf[slot] = m.mg.getUnsignedShort();
                        m.vi[slot] = m.mg.getUnsignedShort();
                    }
                    // jar leaves uk == true after a list-open packet (client.java:15534); 137 closes it
                    if (mode == 1) {
                        // append inventory items not already present (counting from slot 39 down)
                        slot = 39;
                        for (int inv = 0; inv < m.lc; inv++) {
                            if (slot < n) break;
                            boolean present = false;
                            for (int i = 0; i < 40; i++) {
                                if (m.vf[inv] == m.Rj[i]) { present = true; break; }
                            }
                            if (m.vf[inv] == 10) present = true;
                            if (!present) {
                                m.Rj[slot] = StreamBase.bitwiseAnd(32767, m.vf[inv]); // vf[inv] & 0x7FFF
                                m.Jf[slot] = 0;
                                m.vi[slot] = 0;
                                slot--;
                            }
                        }
                    }
                    // clear the selected entry if its item changed out from under us
                    if (m.Di >= 0 && m.Di < 40 && m.fh != m.Rj[m.Di]) {
                        m.Di = -1;
                        m.fh = -2;
                    }
                    return;
                }

                // ---- 137 SEND_SHOP_CLOSE: clear shop/search-list-open flag; reads ZERO bytes (client.java:15661-15663) ----
                if (opcode == 137) {
                    m.uk = false;
                    return;
                }

                // ---- 15: accepted flag for the open trade/duel (Mi) ----
                if (opcode == 15) {
                    m.Mi = m.mg.readRawByte() == 1;
                    return;
                }

                // ---- 240 SEND_GAME_SETTINGS: server-pushed camera/mouse/sound toggles ----
                if (opcode == 240) {
                    m.Kh = m.mg.getUnsignedByte() == 1;          // auto-camera
                    m.Yh = m.mg.getUnsignedByte() == 1;          // one-mouse-button
                    m.ne = m.mg.getUnsignedByte() == 1;          // sound on
                    return;
                }

                // ---- 206 SEND_QUESTS: quest-completion flags (plays a jingle on a change) ----
                if (opcode == 206) {
                    for (int i = 0; i < length - 1; i++) {
                        boolean complete = m.mg.readRawByte() == 1;
                        if (!m.bk[i] && complete) {
                            m.sound.playSound(-127, Mudclient.il[22]);         // obf: a(-127, name) — quest-complete jingle
                        }
                        if (m.bk[i] && !complete) {
                            m.sound.playSound(-66, Mudclient.il[17]);          // obf: a(-66, name)
                        }
                        m.bk[i] = complete;                    // bk[] = quest-complete flags
                    }
                    return;
                }

                // ---- 5 SEND_PRAYERS_ACTIVE: which prayers are currently on ----
                if (opcode == 5) {
                    for (int i = 0; i < 50; i++) {
                        m.fi[i] = m.mg.readRawByte() == 1;       // fi[] = prayer-on flags
                    }
                    return;
                }

                // ---- 42 SEND_BANK_OPEN: open the bank ----
                if (opcode == 42) {
                    m.Fe = true;                               // bank open
                    m.fj = m.mg.getUnsignedByte();               // stored item count
                    m.Gi = m.mg.getUnsignedByte();               // bank capacity
                    for (int i = 0; i < m.fj; i++) {
                        m.ci[i] = m.mg.getUnsignedShort();             // item id
                        m.Xe[i] = m.mg.getSmartSigned();            // amount (var-length)
                    }
                    m.drawHelpMenu(108);                       // obf: C(108) — refresh panel
                    return;
                }

                // ---- 203 SEND_BANK_CLOSE ----
                if (opcode == 203) {
                    m.Fe = false;
                    return;
                }

                // ---- 33 SEND_EXPERIENCE: a single skill's raw xp ----
                if (opcode == 33) {
                    int s = m.mg.getUnsignedByte();
                    m.Ak[s] = m.mg.getInt();
                    return;
                }

                // ---- 176: open the TRADE window (Pj=trade-open; Lg=partner) ----
                if (opcode == 176) {
                    int idx = m.mg.getUnsignedShort();
                    if (m.We[idx] != null) m.Lg = m.We[idx].chatSenderName; // obf .c -> chatSenderName; trade partner name
                    m.ke = false;                              // their-accepted
                    m.vd = false;
                    m.ki = false;                              // your-accepted
                    m.ff = false;
                    m.fd = false;
                    m.Pj = true;                               // trade window open
                    m.Yi = false;
                    m.wj = 0;                                  // their offer count
                    m.Ke = 0;                                  // your offer count
                    return;
                }

                // ---- 225: close the trade-confirm + trade windows (Pj/dd) ----
                if (opcode == 225) {
                    m.Pj = false;
                    m.dd = false;
                    return;
                }

                // ---- 20: open the SHOP window (Xj=shop-open; Lc/Bi stock, Vb/Me base amounts) ----
                if (opcode == 20) {
                    m.Hk = false;
                    m.Xj = true;                               // shop open
                    m.Vi = false;
                    m.re = m.mg.getString();              // shop header/flags string
                    m.nh = m.mg.getUnsignedByte();              // stock size
                    for (int i = 0; i < m.nh; i++) {
                        m.Lc[i] = m.mg.getUnsignedShort();             // item id
                        m.Bi[i] = m.mg.getInt();           // amount in stock
                    }
                    m.Ui = m.mg.getUnsignedByte();              // base-amount list size
                    for (int i = 0; i < m.Ui; i++) {
                        m.Vb[i] = m.mg.getUnsignedShort();
                        m.Me[i] = m.mg.getInt();
                    }
                    return;
                }

                // ---- 6: the other player's TRADE offer (zc/of) ----
                if (opcode == 6) {
                    m.wj = m.mg.getUnsignedByte();               // their item count
                    for (int i = 0; i < m.wj; i++) {
                        m.zc[i] = m.mg.getUnsignedShort();              // item id
                        m.of[i] = m.mg.getInt();            // amount
                    }
                    m.ke = false;                              // reset accepted flags (offer changed)
                    m.ki = false;
                    return;
                }

                // ---- 30: the 4 trade-confirm boolean flags (fd/Yi/vd/ff) — each is (byte == 1) ----
                if (opcode == 30) {
                    m.fd = m.mg.getUnsignedByte() == 1;
                    m.Yi = m.mg.getUnsignedByte() == 1;
                    m.vd = m.mg.getUnsignedByte() == 1;
                    m.ff = m.mg.getUnsignedByte() == 1;
                    m.ke = false;
                    m.ki = false;
                    return;
                }

                // ---- 249 SEND_BANK_UPDATE: single bank slot changed ----
                if (opcode == 249) {
                    int slot = m.mg.getUnsignedByte();
                    int itemId = m.mg.getUnsignedShort();
                    int amount = m.mg.getSmartSigned();
                    if (amount == 0) {                            // removed -> shift down
                        m.fj--;
                        for (int i = slot; i < m.fj; i++) {
                            m.ci[i] = m.ci[i + 1];
                            m.Xe[i] = m.Xe[i + 1];
                        }
                    } else {
                        m.ci[slot] = itemId;
                        m.Xe[slot] = amount;
                        if (slot >= m.fj) m.fj = slot + 1;
                    }
                    m.drawHelpMenu(-103);                       // obf: C(-103)
                    return;
                }

                // ---- 90 SEND_INVENTORY_UPDATEITEM: single inventory slot changed ----
                if (opcode == 90) {
                    int amount = 1;
                    int slot = m.mg.getUnsignedByte();
                    int raw = m.mg.getUnsignedShort();                      // bit15 = wielded
                    if (ClientIOException.itemY[raw & 32767] == 0) {                  // obf fa.e -> itemY; ==0 => stackable -> read amount
                        amount = m.mg.getSmartSigned();
                    }
                    m.vf[slot] = StreamBase.bitwiseAnd(raw, 32767);              // item id
                    m.Aj[slot] = raw / 32768;                   // wielded flag
                    m.xe[slot] = amount;
                    if (slot >= m.lc) m.lc = slot + 1;       // grow inventory count if needed
                    return;
                }

                // ---- 123 SEND_INVENTORY_REMOVE_ITEM: remove a slot, shift the rest down ----
                if (opcode == 123) {
                    int slot = m.mg.getUnsignedByte();
                    m.lc--;
                    for (int i = slot; i < m.lc; i++) {
                        m.vf[i] = m.vf[i + 1];
                        m.xe[i] = m.xe[i + 1];
                        m.Aj[i] = m.Aj[i + 1];
                    }
                    return;
                }

                // ---- 159 SEND_STAT: a single skill changed ----
                if (opcode == 159) {
                    int s = m.mg.getUnsignedByte();
                    m.oh[s] = m.mg.getUnsignedByte();            // current level
                    m.cg[s] = m.mg.getUnsignedByte();            // base level
                    m.Ak[s] = m.mg.getInt();                // xp
                    return;
                }

                // ---- 253: your-accepted flag (ki) ----
                if (opcode == 253) {
                    m.ki = m.mg.readRawByte() == 1;
                    return;
                }

                // ---- 210: their-accepted flag (ke) ----
                if (opcode == 210) {
                    m.ke = m.mg.readRawByte() == 1;
                    return;
                }

                // ---- 172: the second "confirm" window (your/their items + the 4 trade stats) ----
                if (opcode == 172) {
                    m.Cd = false;
                    m.dd = true;                               // show confirm screen
                    m.Pj = false;
                    m.Uc = m.mg.getString();              // confirmation text
                    // your side
                    m.Ve = m.mg.getUnsignedByte();
                    for (int i = 0; i < m.Ve; i++) {
                        m.xj[i] = m.mg.getUnsignedShort();
                        m.kf[i] = m.mg.getInt();
                    }
                    // their side
                    m.Nj = m.mg.getUnsignedByte();
                    for (int i = 0; i < m.Nj; i++) {
                        m.xi[i] = m.mg.getUnsignedShort();
                        m.th[i] = m.mg.getInt();
                    }
                    m.Sh = m.mg.getUnsignedByte();
                    m.gh = m.mg.getUnsignedByte();
                    m.Cc = m.mg.getUnsignedByte();
                    m.Rc = m.mg.getUnsignedByte();
                    return;
                }

                // ---- 204 SEND_PLAY_SOUND: play a named sound effect ----
                if (opcode == 204) {
                    String soundName = m.mg.getString();
                    m.sound.playSound(-73, soundName);               // obf: a(-73, name)
                    return;
                }

                // ---- 36 SEND_BUBBLE: teleport / telegrab / iban-magic bubble effect ----
                if (opcode == 36) {
                    if (m.el < 50) {                           // bubble ring capacity
                        int itemId = m.mg.getUnsignedByte();
                        int x = m.mg.readRawByte() + m.Lf;
                        int y = m.mg.readRawByte() + m.sh;
                        m.Oc[m.el] = itemId;                // bubble item
                        m.oe[m.el] = 0;                     // bubble timer
                        m.Sc[m.el] = x;
                        m.gi[m.el] = y;
                        m.el++;                                // active bubble count
                    }
                    return;
                }

                // ---- 182 SEND_WELCOME_INFO: "Welcome" box (last login IP/date + unread messages) ----
                if (opcode == 182) {
                    if (!m.Dc) {                               // only the first time
                        m.ce = m.mg.getInt();                // days since last login
                        m.hi = m.mg.getUnsignedShort();                 // unread-messages count
                        m.Sb = m.mg.getUnsignedByte();           // recovery-set days
                        m.id = m.mg.getUnsignedShort();                 // last-login IP (packed)
                        m.Oh = true;
                        m.ve = null;
                        m.Dc = true;
                    }
                    return;
                }

                // ---- 89 SEND_BOX2: server message box (not closeable) ----
                if (opcode == 89) {
                    m.Cj = m.mg.getString();              // box text
                    m.mh = true;                               // box visible
                    m.Wk = false;                             // not "closeable" style
                    return;
                }

                // ---- 222 SEND_BOX: server message box (closeable) ----
                if (opcode == 222) {
                    m.Cj = m.mg.getString();
                    m.mh = true;
                    m.Wk = true;
                    return;
                }

                // ---- 114 SEND_FATIGUE: current fatigue value ----
                if (opcode == 114) {
                    m.vg = m.mg.getUnsignedShort();                     // fatigue (0..7500)
                    return;
                }

                // ---- 117 SEND_SLEEPSCREEN: enter the sleep CAPTCHA screen ----
                if (opcode == 117) {
                    if (!m.Qk) m.pg = m.vg;              // seed sleep-fatigue from current
                    m.inputTextCurrent = "";                                  // clear sleep-word input
                    m.Qk = true;                               // sleeping
                    m.inputTextFinal = "";
                    m.li.readSleepWord((byte)-118, m.mg.data, m.Eh + 1); // load CAPTCHA bitmap from packet bytes
                    m.Zj = null;                               // clear "incorrect" prompt
                    return;
                }

                // ---- 244 SEND_SLEEP_FATIGUE: fatigue while sleeping ----
                if (opcode == 244) {
                    m.pg = m.mg.getUnsignedShort();
                    return;
                }

                // ---- 84 SEND_STOPSLEEP: wake up ----
                if (opcode == 84) {
                    m.Qk = false;
                    return;
                }

                // ---- 194 SEND_SLEEPWORD_INCORRECT ----
                if (opcode == 194) {
                    m.Zj = Mudclient.il[55];                             // "...you entered the wrong word..."
                    return;
                }

                // ---- 52 SEND_SYSTEM_UPDATE: countdown to server restart ----
                if (opcode == 52) {
                    m.kc = m.mg.getUnsignedShort() * 32;                // ticks until update (×32)
                    return;
                }

                // ---- 213 SEND_APPEARANCE_KEEPALIVE: no-op keepalive ----
                if (opcode == 213) {
                    return;
                }
                // Unknown opcode -> falls through to the "log + drop" tail below.
            } catch (RuntimeException badPacket) {
                // Authentic recovery (the WITH-body path): log opcode + length + region
                // + a dump of the packet bytes, then force a clean disconnect.
                String dump = Mudclient.il[59] + opcode + Mudclient.il[60] + length + Mudclient.il[56] + m.Lf
                        + Mudclient.il[58] + m.sh + Mudclient.il[62] + m.eh + Mudclient.il[60];
                for (int i = 0; i < length; i++) {
                    dump = dump + m.mg.data[i] + ",";
                }
                Utility.reportError(0x1FFFFF, badPacket, dump);
                m.onStopGame(true);                            // obf: a(true,31)
                return;
            }
            // Reached only for an unhandled opcode: log and drop, then disconnect.
            Utility.reportError(0x1FFFFF, null, Mudclient.il[57] + opcode + Mudclient.il[60] + length);
            m.onStopGame(true);                                // obf: a(true,31)
        } catch (RuntimeException e) {
            throw ErrorHandler.wrap(e, Mudclient.il[61] + opcode + ',' + unused + ',' + length + ')');
        }
    }

    /**
     * Right-click MENU-ACTION dispatcher (the real role of this method; the skeleton's
     * "handleSceneUpdates" name is a misnomer). The selected context-menu entry encodes
     * an action code in {@code zh} (the menu MessageList) plus up to five operands and a
     * string; this turns it into the matching OUTGOING action packet via {@code Jh}.
     *
     * Action codes: 200/210/220 = object op (1st/2nd/examine), 300/310/320/2300 = wall op,
     * 400/410/420/2400 = ground-item-target op, 600..660 = ground-item op, 700..725/2715/715
     * = player op (attack/trade/follow/duel/cast), 800/810/2805/805/2806/2810/2820 = npc op,
     * 900/920 = walk, 1000 = close-shop, 2830..2833 = social, 3xxx = examine-text.
     *
     * obf: void b(boolean,int)   params: (signedShortFlag, menuIndex) — both consumed as
     * MessageList read selectors.
     *
     * WALK-WRAPPER NAMING: the two "walk toward target, then send op" helpers used here are
     * NOT the 9-arg pathfinders in packetout. They are the two 6-arg wrappers:
     *   drawScrollbar  = obf a(byte,int,int,int,boolean,int) [WC] — wraps walkTo  (object/Hh)
     *   drawScrollbar2 = obf a(int,int,int,int,boolean,int)  [BE] — wraps walkToAction (entities)
     * ui_b.part.java named these obf signatures `drawScrollbar`/`drawScrollbar2` (a mislabel
     * there — they actually call walkTo/walkToAction), so we call them by those bound names to
     * keep the assembled class resolvable. Read them as walkTo/walkToAction wrappers.
     */
    void handleSceneUpdates(boolean signedFlag, int menuIndex) {
        try {
            // Pull the selected menu entry's action code + operands out of the menu list.
            int action = m.zh.getEntryXPos(-110, menuIndex);          // action code (200,300,…)
            int a1 = m.zh.getEntryColorE(true, menuIndex);              // operand 1 (id / dx)
            int a2 = m.zh.getEntryColorCode((byte)97, menuIndex);          // operand 2 (dy)
            int a3 = m.zh.getEntrySprite(menuIndex, (byte)22);          // operand 3
            int a4 = m.zh.getEntryMessageColor(menuIndex, signedFlag);        // operand 4
            int a5 = m.zh.getEntryLayer(true, menuIndex);              // operand 5 (item slot)
            String str = m.zh.getEntryName(menuIndex, -4126);         // string operand (target name)

            if (action == 200) {                              // object: use 1st option
                m.drawScrollbar((byte)10, m.sh, a2, a1, true, m.Lf); // obf: a((byte)10,sh,a2,a1,true,Lf) — walkTo wrapper
                m.Jh.newPacket(249, 0);                            // -> opcode 249 (OP_OBJECT_1) [outgoing]
                m.Jh.outBuffer.putShort(a1 + m.Qg);
                m.Jh.outBuffer.putShort(a2 + m.zg);
                m.Jh.outBuffer.putShort(a3);
                m.Jh.outBuffer.putShort(a4);
                m.Jh.finishPacket(21294);                             // flush
                m.af = -1;
            }
            if (action == 210) {                              // object: use 2nd option
                m.drawScrollbar((byte)10, m.sh, a2, a1, true, m.Lf);
                m.Jh.newPacket(53, 0);                             // -> outgoing object-action 2
                m.Jh.outBuffer.putShort(m.Qg + a1);
                m.Jh.outBuffer.putShort(m.zg + a2);
                m.Jh.outBuffer.putShort(a3);
                m.Jh.outBuffer.putShort(a4);
                m.Jh.finishPacket(21294);
                m.Bh = -1;
            }
            if (action == 220) {                              // object: examine (path then op)
                m.drawScrollbar((byte)10, m.sh, a2, a1, true, m.Lf);
                m.Jh.newPacket(247, 0);
                m.Jh.outBuffer.putShort(a1 + m.Qg);
                m.Jh.outBuffer.putShort(m.zg + a2);
                m.Jh.outBuffer.putShort(a3);
                m.Jh.finishPacket(21294);
            }
            if (action == 3600 || action == 3200) {           // object/scenery examine -> show def text
                m.showServerMessage(false, null, 0, CharTable.itemDescriptions[a1], 0, 0, null, null); // ga.b
            }
            if (action == 300) {                              // wall/boundary: use 1st option
                // drawSprite = obf a(boolean,int,int,int): the wall-walk helper (walks toward
                // the boundary tile then sends the op). util.part.java bound this obf signature
                // to `drawSprite`; it is NOT the 4-int handleGameClick.
                m.drawSprite(false, a1, a2, a3);      // obf: a(false,a1,a2,a3) — wall-walk wrapper
                m.Jh.newPacket(180, 0);
                m.Jh.outBuffer.putShort(m.Qg + a1);
                m.Jh.outBuffer.putShort(m.zg + a2);
                m.Jh.outBuffer.putByte(a3);
                m.Jh.outBuffer.putShort(a4);
                m.Jh.finishPacket(21294);
                m.af = -1;
            }
            if (action == 310) {                              // wall: use 2nd option
                m.drawSprite(false, a1, a2, a3);
                m.Jh.newPacket(161, 0);
                m.Jh.outBuffer.putShort(a1 + m.Qg);
                m.Jh.outBuffer.putShort(a2 + m.zg);
                m.Jh.outBuffer.putByte(a3);
                m.Jh.outBuffer.putShort(a4);
                m.Jh.finishPacket(21294);
                m.Bh = -1;
            }
            if (action == 320) {                              // wall: examine
                m.drawSprite(signedFlag, a1, a2, a3);
                m.Jh.newPacket(14, 0);
                m.Jh.outBuffer.putShort(a1 + m.Qg);
                m.Jh.outBuffer.putShort(a2 + m.zg);
                m.Jh.outBuffer.putByte(a3);
                m.Jh.finishPacket(21294);
            }
            if (action == 2300) {                             // wall: use-item-on
                m.drawSprite(false, a1, a2, a3);
                m.Jh.newPacket(127, 0);
                m.Jh.outBuffer.putShort(m.Qg + a1);
                m.Jh.outBuffer.putShort(a2 + m.zg);
                m.Jh.outBuffer.putByte(a3);
                m.Jh.finishPacket(21294);
            }
            if (action == 3300) {                             // wall examine -> show def text
                m.showServerMessage(false, null, 0, NameTable.textureNames[a1], 0, 0, null, null);
            }
            if (action == 400) {                              // ground item: 1st option
                m.drawBox(5126, a4, a1, a2, a3);           // obf: b(5126,a4,a1,a2,a3) — path/select helper
                m.Jh.newPacket(99, 0);
                m.Jh.outBuffer.putShort(a1 + m.Qg);
                m.Jh.outBuffer.putShort(m.zg + a2);
                m.Jh.outBuffer.putShort(a5);
                m.Jh.finishPacket(21294);
                m.af = -1;
            }
            if (action == 410) {                              // ground item: use-with-item
                m.drawBox(5126, a4, a1, a2, a3);
                m.Jh.newPacket(115, 0);
                m.Jh.outBuffer.putShort(a1 + m.Qg);
                m.Jh.outBuffer.putShort(a2 + m.zg);
                m.Jh.outBuffer.putShort(a5);
                m.Jh.finishPacket(21294);
                m.Bh = -1;
            }
            if (action == 420) {                              // ground item: examine target
                m.drawBox(5126, a4, a1, a2, a3);
                m.Jh.newPacket(136, 0);
                m.Jh.outBuffer.putShort(m.Qg + a1);
                m.Jh.outBuffer.putShort(a2 + m.zg);
                m.Jh.finishPacket(21294);
            }
            if (action == 2400) {                             // ground item: cast spell on
                m.drawBox(5126, a4, a1, a2, a3);
                m.Jh.newPacket(79, 0);
                m.Jh.outBuffer.putShort(m.Qg + a1);
                m.Jh.outBuffer.putShort(m.zg + a2);
                m.Jh.finishPacket(21294);
            }
            if (action == 3400) {                             // ground item examine -> def text
                m.showServerMessage(false, null, 0, ClientRuntimeException.stringScratch[a1], 0, 0, null, null);
            }
            if (action == 600) {                              // ground item (simple): 1st option
                m.Jh.newPacket(4, 0);
                m.Jh.outBuffer.putShort(a1);
                m.Jh.outBuffer.putShort(a2);
                m.Jh.finishPacket(21294);
                m.af = -1;
            }
            if (action == 610) {
                m.Jh.newPacket(91, 0);
                m.Jh.outBuffer.putShort(a1);
                m.Jh.outBuffer.putShort(a2);
                m.Jh.finishPacket(21294);
                m.Bh = -1;
            }
            if (action == 620) {
                m.Jh.newPacket(170, 0);
                m.Jh.outBuffer.putShort(a1);
                m.Jh.finishPacket(21294);
            }
            if (action == 630) {
                m.Jh.newPacket(169, 0);
                m.Jh.outBuffer.putShort(a1);
                m.Jh.finishPacket(21294);
            }
            if (action == 640) {
                m.Jh.newPacket(90, 0);
                m.Jh.outBuffer.putShort(a1);
                m.Jh.finishPacket(21294);
            }
            if (action == 650) {                              // inventory item: examine (local def)
                m.Bh = a1;
                m.qc = 0;
                m.ig = DecodeBuffer.chatFilterCache[m.vf[m.Bh]];
            }
            if (action == 660) {                              // inventory item: examine -> show text
                m.Jh.newPacket(246, 0);
                m.Jh.outBuffer.putShort(a1);
                m.Jh.finishPacket(21294);
                m.qc = 0;
                m.Bh = -1;
                m.showServerMessage(false, null, 0, Mudclient.il[511] + DecodeBuffer.chatFilterCache[m.vf[a1]], 7, 0, null, null);
            }
            // 700..725/715: player actions (attack/trade/follow/duel/cast). Walk toward the
            // target's tile first, then send the op.
            if (action == 700) {                              // attack player
                GameCharacter target = m.getPlayer(a1, (byte)-123);   // obf: b(a1,-123)
                int ty = (target.currentX - 64) / m.Ug;
                int tx = (target.currentY - 64) / m.Ug;
                m.drawScrollbar2(tx, ty, m.sh, m.Lf, true, 8); // obf: a(tx,ty,sh,Lf,true,8) — walkToAction wrapper
                m.Jh.newPacket(50, 0);
                m.Jh.outBuffer.putShort(a1);
                m.Jh.outBuffer.putShort(a2);
                m.Jh.finishPacket(21294);
                m.af = -1;
            }
            if (action == 710) {                              // trade player
                GameCharacter target = m.getPlayer(a1, (byte)-123);
                int ty = (target.currentX - 64) / m.Ug;
                int tx = (target.currentY - 64) / m.Ug;
                m.drawScrollbar2(tx, ty, m.sh, m.Lf, true, 8);
                m.Jh.newPacket(135, 0);
                m.Jh.outBuffer.putShort(a1);
                m.Jh.outBuffer.putShort(a2);
                m.Jh.finishPacket(21294);
                m.Bh = -1;
            }
            if (action == 720) {                              // follow player
                GameCharacter target = m.getPlayer(a1, (byte)-123);
                int ty = (target.currentX - 64) / m.Ug;
                int tx = (target.currentY - 64) / m.Ug;
                m.drawScrollbar2(tx, ty, m.sh, m.Lf, true, 8);
                m.Jh.newPacket(153, 0);
                m.Jh.outBuffer.putShort(a1);
                m.Jh.finishPacket(21294);
            }
            if (action == 725) {                              // duel player
                GameCharacter target = m.getPlayer(a1, (byte)-123);
                int ty = (target.currentX - 64) / m.Ug;
                int tx = (target.currentY - 64) / m.Ug;
                m.drawScrollbar2(tx, ty, m.sh, m.Lf, true, 8);
                m.Jh.newPacket(202, 0);
                m.Jh.outBuffer.putShort(a1);
                m.Jh.finishPacket(21294);
            }
            if (action == 2715 || action == 715) {            // cast spell / use item on player
                GameCharacter target = m.getPlayer(a1, (byte)-123);
                int ty = (target.currentX - 64) / m.Ug;
                int tx = (target.currentY - 64) / m.Ug;
                m.drawScrollbar2(tx, ty, m.sh, m.Lf, true, 8);
                m.Jh.newPacket(190, 0);
                m.Jh.outBuffer.putShort(a1);
                m.Jh.finishPacket(21294);
            }
            if (action == 3700) {                             // player examine -> def text
                m.showServerMessage(false, null, 0, SurfaceSprite.equipAc[a1], 0, 0, null, null); // ba.ac
            }
            // 800..2820: npc actions
            if (action == 800) {                              // attack npc
                GameCharacter npc = m.getNpc(a1, 220);                // obf: d(a1,220)
                int ty = (npc.currentX - 64) / m.Ug;
                int tx = (npc.currentY - 64) / m.Ug;
                m.drawScrollbar2(tx, ty, m.sh, m.Lf, true, 8);
                m.Jh.newPacket(229, 0);
                m.Jh.outBuffer.putShort(a1);
                m.Jh.outBuffer.putShort(a2);
                m.Jh.finishPacket(21294);
                m.af = -1;
            }
            if (action == 810) {                              // talk-to npc
                GameCharacter npc = m.getNpc(a1, 220);
                int ty = (npc.currentX - 64) / m.Ug;
                int tx = (npc.currentY - 64) / m.Ug;
                m.drawScrollbar2(tx, ty, m.sh, m.Lf, true, 8);
                m.Jh.newPacket(113, 0);
                m.Jh.outBuffer.putShort(a1);
                m.Jh.outBuffer.putShort(a2);
                m.Jh.finishPacket(21294);
                m.Bh = -1;
            }
            if (action == 2805 || action == 805) {            // cast spell / use item on npc
                GameCharacter npc = m.getNpc(a1, 220);
                int ty = (npc.currentX - 64) / m.Ug;
                int tx = (npc.currentY - 64) / m.Ug;
                m.drawScrollbar2(tx, ty, m.sh, m.Lf, true, 8);
                m.Jh.newPacket(171, 0);
                m.Jh.outBuffer.putShort(a1);
                m.Jh.finishPacket(21294);
            }
            if (action == 2806) {                             // npc: 1st option
                m.Jh.newPacket(103, 0);
                m.Jh.outBuffer.putShort(a1);
                m.Jh.finishPacket(21294);
            }
            if (action == 2810) {                             // npc: 2nd option
                m.Jh.newPacket(142, 0);
                m.Jh.outBuffer.putShort(a1);
                m.Jh.finishPacket(21294);
            }
            if (action == 2820) {                             // npc: examine
                m.Jh.newPacket(165, 0);
                m.Jh.outBuffer.putShort(a1);
                m.Jh.finishPacket(21294);
            }
            // 28xx social actions (operate on the picked player name `str`)
            if (action == 2833) {                             // send public/quick message
                m.inputTextFinal = "";
                m.Vf = 1;
                m.inputTextCurrent = str;
            }
            if (action == 2831) {                             // add friend
                m.packets.sendAddFriend(97, str);                  // obf: b(97,str)
            }
            if (action == 2832) {                             // add ignore
                m.packets.sendAddIgnore(str, (byte)5);             // obf: a(str,5)
            }
            if (action == 2830) {                             // send private message (open entry)
                m.Qd = str;
                m.inputPmCurrent = "";
                m.Bj = 2;
                m.inputPmFinal = "";
            }
            if (action == 900) {                              // walk to clicked tile (then face)
                m.drawScrollbar2(a2, a1, m.sh, m.Lf, true, 8);
                m.Jh.newPacket(158, 0);
                m.Jh.outBuffer.putShort(a1 + m.Qg);
                m.Jh.outBuffer.putShort(m.zg + a2);
                m.Jh.outBuffer.putShort(a3);
                m.Jh.finishPacket(21294);
                m.af = -1;
            }
            if (action == 920) {                              // walk to clicked tile (no path-send)
                m.drawScrollbar2(a2, a1, m.sh, m.Lf, false, 8);
                if (m.xh == -24) m.xh = 24;             // tutorial walk-acknowledged
            }
            if (action == 1000) {                             // close shop
                m.Jh.newPacket(137, 0);
                m.Jh.outBuffer.putShort(a1);
                m.Jh.finishPacket(21294);
                m.af = -1;
            }
            if (action == 4000) {                             // cancel / clear pending action
                m.af = -1;
                m.Bh = -1;
            }
        } catch (RuntimeException e) {
            throw ErrorHandler.wrap(e, Mudclient.il[510] + signedFlag + ',' + menuIndex + ')');
        }
    }

    /**
     * Social / private-message packet sub-dispatcher (broadly matches the skeleton's
     * "onFriendUpdate" intent). Re-reads the opcode from the stream header, applies the
     * friend-list, ignore-list, private-message and server-message packets, and forwards
     * any opcode it does not own to {@link #handlePacket}.
     *
     * The original is a fall-through cascade: every path that does not early-{@code return}
     * lands on a shared tail that, unless the caller's mode operand {@code a} is 87, sends a
     * LOGOUT request. The 87 sentinel suppresses the auto-logout for the PM-flush path.
     *
     * obf: void a(int,int,int)   params: (a, b, opcode). `b` is a residual operand forwarded
     * to handlePacket; the opcode is re-resolved via {@code Jh.a(507,opcode)}.
     */
    void onFriendUpdate(int a, int b, int opcode) {
        try {
            opcode = m.Jh.isaacCommand(507, opcode);      // re-resolve opcode from stream header (da.a(int,int))

            if (opcode == 131) {                             // ---- 131 SEND_SERVER_MESSAGE ----
                int msgType = m.mg.getUnsignedByte();     // chat tab / message-type id (ja.a(byte))
                int infoFlags = m.mg.getUnsignedByte();   // bit0 = has sender, bit1 = has colour
                String message = m.mg.getString();        // ja.c(byte)
                String sender = null, senderDup = null, colour = null;
                if ((infoFlags & 1) != 0) sender = m.mg.getString();
                if ((infoFlags & 1) != 0) senderDup = m.mg.getString(); // authentic duplicate read
                if ((infoFlags & 2) != 0) colour = m.mg.getString();
                m.showServerMessage(false, sender, 0, message, msgType, 0, senderDup, colour);

            } else if (opcode == 4) {                        // ---- 4 SEND_LOGOUT_REQUEST_CONFIRM ----
                // server allows logout -> CONFIRM_LOGOUT + tear down
                m.sendConfirmLogoutAck(true, a - 56);     // obf: a(true, a-56)

            } else if (opcode == 183) {                      // ---- 183 SEND_CANT_LOGOUT ----
                m.sendConfirmLogout((byte)-65);           // obf: g(-65)

            } else if (opcode == 189) {                      // ---- 189 SEND_28_BYTES_UNUSED ----
                m.mg.offset += 28;                             // skip a fixed 28-byte block
                if (m.mg.verifyCrc(-422797528)) {         // CRC/length check (ja.e(int))
                    Packet.telemetry(m.mg, 26628, m.mg.offset - 28); // obf: b.a(tb,int,int) static telemetry
                }

            } else if (opcode == 165) {                      // ---- 165 SEND_LOGOUT ----
                m.sendConfirmLogoutAck(false, 31);        // obf: a(false, 31) — reset session only

            } else if (opcode == 149) {                      // ---- 149 SEND_FRIEND_UPDATE ----
                String name = m.mg.getString();           // current name
                String formerName = m.mg.getString();      // former name (for rename match)
                int flags = m.mg.getUnsignedByte();        // bit0: 1 => match-by-former (rename); bit2: online
                boolean matchByFormer = (flags & 1) != 0;
                boolean nowOnline = (flags & 4) != 0;
                String onlineWorld = null;
                if (nowOnline) onlineWorld = m.mg.getString();
                for (int f = 0; f < FontWidths.listEntryCount; f++) { // n.g = friends count
                    if (!matchByFormer) {
                        if (Surface.decoyStrings200[f].equals(name)) {          // matched by current name -> update status
                            if (m.friendListWorlds[f] == null && nowOnline)
                                m.showServerMessage(false, null, 0, name + Mudclient.STRINGS[9], 5, 0, null, null); // "has logged in"
                            if (m.friendListWorlds[f] != null && !nowOnline)
                                m.showServerMessage(false, null, 0, name + Mudclient.STRINGS[8], 5, 0, null, null); // "has logged out"
                            CacheUpdater.archiveNames[f] = formerName;
                            m.friendListWorlds[f] = onlineWorld;  // null = offline marker (ac.z)
                            Mudclient.Fj[f] = flags;
                            b = 0;
                            m.sortFriendsList(51);           // obf: v(51) — re-sort friends list
                            return;
                        }
                    } else if (Surface.decoyStrings200[f].equals(formerName)) {  // matched by former name -> rename in place
                        if (m.friendListWorlds[f] == null && nowOnline)
                            m.showServerMessage(false, null, 0, name + Mudclient.STRINGS[9], 5, 0, null, null);
                        if (m.friendListWorlds[f] != null && !nowOnline)
                            m.showServerMessage(false, null, 0, name + Mudclient.STRINGS[8], 5, 0, null, null);
                        Surface.decoyStrings200[f] = name;
                        CacheUpdater.archiveNames[f] = formerName;
                        m.friendListWorlds[f] = onlineWorld;  // ac.z
                        Mudclient.Fj[f] = flags;
                        b = 0;
                        m.sortFriendsList(50);               // obf: v(50)
                        return;
                    }
                }
                if (matchByFormer) {                          // rename target not present -> log + drop
                    System.out.println(Mudclient.STRINGS[4] + formerName + Mudclient.STRINGS[3]);
                    return;
                }
                // insert-if-missing -> append a new friend
                Surface.decoyStrings200[FontWidths.listEntryCount] = name;
                CacheUpdater.archiveNames[FontWidths.listEntryCount] = formerName;
                m.friendListWorlds[FontWidths.listEntryCount] = onlineWorld; // ac.z
                Mudclient.Fj[FontWidths.listEntryCount] = flags;
                FontWidths.listEntryCount++;                    // n.g++
                m.sortFriendsList(66);                       // obf: v(66)

            } else if (opcode == 237) {                      // ---- 237 SEND_IGNORE_LIST_RENAME ----
                String newName = m.mg.getString();
                String newName2 = m.mg.getString();
                if (newName2.length() == 0) newName2 = newName;
                String oldWorld = m.mg.getString();
                String oldName = m.mg.getString();
                if (oldName.length() == 0) oldName = newName;
                boolean matchExisting = m.mg.getUnsignedByte() == 1;
                for (int idx = 0; idx < LinkedQueue.DEAD_G; idx++) {       // db.g = ignore count
                    if (matchExisting) {
                        if (SpriteScaler.playerNames[idx].equals(oldName)) {     // rename an existing ignore entry
                            Globals.strings[idx] = newName;
                            SpriteScaler.playerNames[idx] = newName2;
                            SpriteScaler.playerTitles[idx] = oldWorld;
                            Surface.decoyStrings100[idx] = oldName;
                            return;
                        }
                    } else if (SpriteScaler.playerNames[idx].equals(newName2)) {
                        return;                              // already present
                    }
                }
                if (matchExisting) {                          // rename target not present -> log + drop
                    System.out.println(Mudclient.STRINGS[7] + oldName + Mudclient.STRINGS[5]);
                    return;
                }
                // append a new ignore entry
                Globals.strings[LinkedQueue.DEAD_G] = newName;
                SpriteScaler.playerNames[LinkedQueue.DEAD_G] = newName2;
                SpriteScaler.playerTitles[LinkedQueue.DEAD_G] = oldWorld;
                Surface.decoyStrings100[LinkedQueue.DEAD_G] = oldName;
                LinkedQueue.DEAD_G++;

            } else if (opcode == 109) {                      // ---- 109 SEND_IGNORE_LIST ----
                LinkedQueue.DEAD_G = m.mg.getUnsignedByte();                  // ignore count
                for (int idx = 0; idx < LinkedQueue.DEAD_G; idx++) {
                    Globals.strings[idx] = m.mg.getString();
                    SpriteScaler.playerNames[idx] = m.mg.getString();
                    SpriteScaler.playerTitles[idx] = m.mg.getString();
                    Surface.decoyStrings100[idx] = m.mg.getString();
                }

            } else if (opcode == 120) {                      // ---- 120 SEND_PRIVATE_MESSAGE ----
                String fromName = m.mg.getString();
                String fromFormer = m.mg.getString();
                int icon = m.mg.getUnsignedByte();         // moderator/icon sprite
                long messageId = m.mg.getLong();           // 8-byte message id (world+counter) (ja.g(int))
                String message = SpriteScaler.readPacketString(m.mg, false); // decode scrambled body (ia.a(tb,bool))
                // drop if we've already seen this exact message id (anti-duplicate ring)
                for (int i = 0; i < 100; i++) {
                    if (m.Zd[i] == messageId) return;
                }
                m.Zd[m.Ag] = messageId;                 // record id in the ring
                m.Ag = (m.Ag + 1) % 100;
                m.showServerMessage(icon == 2, fromName, 0, message, 1, icon, fromFormer, null);

            } else if (opcode == 51) {                       // ---- 51 SEND_PRIVACY_SETTINGS ----
                m.De = m.mg.getUnsignedByte();              // block-chat privacy
                m.dc = m.mg.getUnsignedByte();              // public-chat privacy
                m.Vg = m.mg.getUnsignedByte();              // private-chat privacy
                m.ui = m.mg.getUnsignedByte();              // trade/duel privacy

            } else if (opcode == 87) {                       // ---- 87 SEND_PRIVATE_MESSAGE_SENT ----
                String toName = m.mg.getString();
                String message = SpriteScaler.readPacketString(m.mg, false);
                m.showServerMessage(false, toName, 0, message, 2, 0, toName, null);

            } else {                                         // everything else -> master dispatcher
                this.handlePacket(opcode, (byte)41, b);       // obf: b(opcode,41,b)
            }

            // Shared tail: unless the mode operand `a` is 87, request a logout.
            if (a != 87) {
                m.requestLogout(56);                       // obf: B(56) — send opcode 102 (LOGOUT)
            }
        } catch (RuntimeException e) {
            throw ErrorHandler.wrap(e, Mudclient.STRINGS[6] + a + ',' + b + ',' + opcode + ')');
        }
    }

    /**
     * Social-entry DIALOG renderer (the real role; the skeleton's "applyAppearanceUpdate"
     * name is a misnomer — this parses no packet). Draws the add-friend / add-ignore /
     * send-private-message popup: a titled box with a Friends tab and an Ignore tab, plus
     * the appropriate name list, and handles clicks/typing inside it. {@code handleInput}
     * enables click handling; {@code suppressInput} pins the panel ({@code Be = -88}).
     *
     * {@code pk} selects the active tab (0 = friends, 1 = ignore). Drawing uses
     * {@code li} (obf li) and the {@code panelLogin} widget container (obf zk); text
     * comes from the STRINGS pool (obf il).
     *
     * obf: void a(boolean,boolean)   params: (handleInput, suppressInput)
     */
    void applyAppearanceUpdate(boolean handleInput, boolean suppressInput) {
        try {
            int boxX = m.li.width - 199;                       // surface width - 199 (centre the box)
            int titleY = 36;
            m.li.drawSprite(-1, m.tg + 5, 3, boxX - 49); // li.b(int,int,int,int) draw backdrop sprite
            int boxW = 196;
            int boxH = 182;
            if (suppressInput) m.Be = -88;

            // header gradient: highlight the active tab (pk 0 -> right bright, pk 1 -> left bright)
            int leftColour = ISAAC.packColor(160, 9570, 160, 160);     // o.a(int,int,int,int)
            int rightColour;
            if (m.pk == 0) {
                rightColour = ISAAC.packColor(220, 9570, 220, 220);
            } else {
                leftColour = ISAAC.packColor(220, 9570, 220, 220);
                rightColour = ISAAC.packColor(220, 9570, 220, 220);
            }
            m.li.drawBoxAlpha(128, boxX, 24, 0, titleY, boxW / 2, leftColour);   // li.c(int x7)
            m.li.drawBoxAlpha(128, boxX + boxW / 2, 24, 0, titleY, boxW / 2, rightColour);
            m.li.drawBoxAlpha(128, boxX, boxH - 24, 0, titleY + 24, boxW, ISAAC.packColor(220, 9570, 220, 220));
            // borders
            m.li.drawLineHoriz(boxW, 0, boxX, titleY + 24, (byte)95);
            m.li.drawLineVert(boxX + boxW / 2, titleY, 0, 24, 0); // li.b(int x5)
            m.li.drawLineHoriz(boxW, 0, boxX, titleY + boxH - 16, (byte)-113);
            // tab captions
            m.li.drawstringRight(boxX + boxW / 4, Mudclient.STRINGS[260], 0, 0, 4, titleY + 16);          // li.a(int,String,int,int,int,int)
            m.li.drawstringRight(boxX + boxW / 4 + boxW / 2, Mudclient.STRINGS[258], 0, 0, 4, titleY + 16);

            m.zk.resetItemCount((byte)-82, m.Hi);       // reset the dialog's list rows (qa.c(byte,int))

            // --- populate the list with the appropriate names ---
            if (m.pk == 0) {                               // FRIENDS list
                for (int f = 0; f < FontWidths.listEntryCount; f++) { // n.g = friends count
                    String statusColour;
                    if ((Mudclient.Fj[f] & 2) == 0) {                   // offline
                        if ((Mudclient.Fj[f] & 4) == 0) statusColour = Mudclient.STRINGS[10]; // grey
                        else statusColour = Mudclient.STRINGS[20];           // intermediate
                    } else {
                        statusColour = Mudclient.STRINGS[27];                // green (online)
                    }
                    // truncate the name so the row fits 120px
                    String name = Surface.decoyStrings200[f];
                    int len = Surface.decoyStrings200[f].length();
                    int cut = 0;
                    while (m.li.textWidth(1, 111, name) > 120) {
                        name = Surface.decoyStrings200[f].substring(0, len - (++cut)) + Mudclient.STRINGS[261]; // "..."
                    }
                    m.zk.setListItem(f, null, 49, 0, null, statusColour + name + Mudclient.STRINGS[262], m.Hi); // qa.a(int,String,int,int,String,String,int)
                }
            }
            if (m.pk == 1) {                               // IGNORE list
                for (int i = 0; i < LinkedQueue.DEAD_G; i++) {              // db.g = ignore count
                    String name = Globals.strings[i];
                    int len = Globals.strings[i].length();
                    int cut = 0;
                    while (m.li.textWidth(1, 100, name) > 120) {
                        name = Globals.strings[i].substring(0, len - (++cut)) + Mudclient.STRINGS[261];
                    }
                    m.zk.setListItem(i, null, 60, 0, null, Mudclient.STRINGS[20] + name + Mudclient.STRINGS[262], m.Hi);
                }
            }
            m.zk.render((byte)-43);                        // finalize list layout (qa.a(byte))
            m.nj = -1;                                     // hovered friends row
            m.wk = -1;                                     // hovered ignore row

            // --- IGNORE tab caption + (when friends active) ignore-row hover highlight ---
            if (m.pk == 0) {
                int row = m.zk.getHoveredItem(m.Hi, 17050);
                if (row >= 0 && m.mouseX < 489) {
                    if (m.mouseX > 430) m.wk = -(row + 2);   // hovering the "remove" zone
                    else m.wk = row;
                }
                m.li.drawstringRight(boxX + boxW / 2, Mudclient.STRINGS[266], 0xFFFFFF, 0, 1, titleY + 35);
                int ignoreColour;
                if (m.mouseX > boxX && m.mouseX < boxX + boxW
                        && m.mouseY > boxH + titleY - 16 && m.mouseY < boxH + titleY) {
                    ignoreColour = 0xFFFF00;                  // yellow when hovered
                } else {
                    ignoreColour = 0xFFFFFF;
                }
                m.li.drawstringRight(boxX + boxW / 2, Mudclient.STRINGS[259], ignoreColour, 0, 1, boxH + titleY - 3); // "Ignore"
            }

            // --- FRIENDS tab caption + (when ignore active) friends-row hover highlight ---
            if (m.pk == 1) {
                int row = m.zk.getHoveredItem(m.Hi, 17050);
                if (row >= 0 && m.mouseX < 489) {
                    if (m.mouseX > 430) m.nj = row;
                    else m.nj = -(row + 2);
                }
                m.li.drawstringRight(boxX + boxW / 2, Mudclient.STRINGS[263], 0xFFFFFF, 0, 1, titleY + 35);
                int friendsColour;
                if (m.mouseX <= boxX || m.mouseX >= boxX + boxW
                        || boxH + titleY - 16 >= m.mouseY || m.mouseY >= boxH + titleY) {
                    friendsColour = 0xFFFFFF;
                } else {
                    friendsColour = 0xFFFF00;                 // yellow when hovered
                }
                m.li.drawstringRight(boxX + boxW / 2, Mudclient.STRINGS[265], friendsColour, 0, 1, titleY + boxH - 3); // "Friends"
            }

            // --- input handling (skipped when only redrawing) ---
            if (!handleInput) return;
            int my = m.mouseY - 36;                            // mouse Y relative to box
            int mx = m.mouseX + 199 - m.li.width;                // mouse X relative to box
            if (mx < 0 || my < 0 || mx >= 196 || my >= 182) return;

            m.zk.handleMouseInput(m.lastMouseButtonDown, my + 36, -9989, m.Cf, mx + m.li.width - 199); // route mouse into panel (qa.b(int x5))

            // tab switching by clicking the header tabs (top 24px)
            if (my <= 24 && m.Cf == 1) {
                if (mx < 98 && m.pk == 1) {                // left tab -> friends
                    m.pk = 0;
                    m.zk.clearList(m.Hi, 14);
                }
                if (mx > 98 && m.pk == 0) {                // right tab -> ignore
                    m.pk = 1;
                    m.zk.clearList(m.Hi, 14);
                }
            }
            // friends tab: clicking a row -> remove (right zone) or open PM (online friend)
            if (m.Cf == 1 && m.pk == 0) {
                int row = m.zk.getHoveredItem(m.Hi, 17050);
                if (row >= 0 && m.mouseX < 489) {
                    if (m.mouseX > 429) {
                        m.packets.sendRemoveFriend(Surface.decoyStrings200[row], (byte)69); // obf: b(name,69)
                    }
                    if ((Mudclient.Fj[row] & 4) != 0) {                 // open PM entry for this friend
                        m.Bj = 2;
                        m.Qd = Surface.decoyStrings200[row];
                        m.inputPmFinal = "";
                        m.inputPmCurrent = "";
                    }
                }
            }
            // ignore tab: clicking a row in the right zone -> remove that ignore
            if (m.Cf == 1 && m.pk == 1) {
                int row = m.zk.getHoveredItem(m.Hi, 17050);
                if (row >= 0 && m.mouseX < 489 && m.mouseX > 430) {
                    m.packets.sendRemoveIgnore((byte)-15, SpriteScaler.playerNames[row]); // obf: a(-15, name)
                }
            }
            // bottom button -> open add-friend (friends tab) / add-ignore (ignore tab) entry
            if (my > 166 && m.Cf == 1 && m.pk == 0) {
                m.inputTextFinal = "";
                m.inputTextCurrent = "";
                m.Bj = 1;
            }
            if (my > 166 && m.Cf == 1 && m.pk == 1) {
                m.inputTextFinal = "";
                m.Bj = 3;
                m.inputTextCurrent = "";
            }
            m.Cf = 0;                                       // consume the click
        } catch (RuntimeException e) {
            throw ErrorHandler.wrap(e, Mudclient.STRINGS[264] + handleInput + ',' + suppressInput + ')');
        }
    }
}
