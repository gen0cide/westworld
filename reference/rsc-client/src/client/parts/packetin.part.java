    // ===== packetin =====
    //
    // Incoming server->client packet handling. Opcode numbers cited below are the
    // RSC protocol-235 SEND_* values from OpenRSC's Payload235Generator.
    //
    // IMPORTANT NAMING NOTE (read before trusting the method names):
    // The skeleton's *proposed names* for several methods in this group turned out to
    // be inaccurate once the (CFR-only) bodies were decompiled. To keep cross-method
    // call sites resolvable by the orchestrator, the skeleton's proposed name is kept
    // as the Java method name, but each doc comment states what the method ACTUALLY
    // does. Concretely:
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
    //                                              send-message list). It does NOT parse any
    //                                              appearance packet. (misnamed)
    //
    // Field-name note: only ~80 of the class's 484 fields are named in the skeleton.
    // Fields not in the skeleton tables keep their obfuscated names; the role of the
    // load-bearing ones is given inline (cross-referenced from mudclient204's
    // handleIncomingPacket). Helper-method calls use the skeleton's proposed name with
    // an `// obf:` note so any residual mismatch stays traceable.
    //
    // Stripped per instructions: `boolean bl = client.vh;` opaque predicate and all the
    // dead `if(!bl)`/`if(bl)return`/`while(!bl)` control flow it gates; `++<counter>;`
    // profiling bumps; `try{...}catch(RuntimeException e){throw ErrorHandler.a(e,sig)}`
    // wrappers (unwrapped to the body); and junk masks before bit shifts. The bit-stream
    // reader `incomingPacket` (obf `mg`, type ja/BitBuffer) is read via:
    //   .a((byte)104)  -> read one unsigned byte
    //   .f(255)        -> read one unsigned short (16 bits)  [normalized to readShort()]
    //   .h(20869)      -> read one unsigned short            [normalized to readShort()]
    //   .b(-129)       -> read one signed 16-bit             [normalized to readSignedShort()]
    //   .c((byte)-44)  -> read a zero-terminated string      [normalized to readString()]
    //   .f(bias, nBits)-> read nBits as an unsigned bitfield [normalized to readBits(nBits)]
    //   .g(0)          -> read an 8-byte long
    // For clarity these are written below with intent-revealing pseudo-calls; see the
    // `// obf:` note on each call for the exact decompiled form.

    /**
     * MASTER server->client packet dispatch (protocol 235). Reads the already-buffered
     * packet (opcode in {@code opcode}, body length in {@code length}) from
     * {@code incomingPacket} and applies it to game state. Roughly 50 opcodes plus four
     * bulk region streams (players/walls/scenery/npcs) are decoded here.
     *
     * On any RuntimeException the original logs the packet bytes and forces a clean
     * disconnect ({@code onStopGame(true)}); that recovery path is preserved.
     *
     * obf: void b(int,byte,int)   params: (opcode, <anti-tamper dummy byte>, length)
     */
    private void handlePacket(int opcode, byte unused, int length) {
        try {
            switch (opcode) {

            // ---- 191 SEND_PLAYER_COORDS: local player position + nearby-player movement stream ----
            case 191: {
                // double-buffer the in-view player list (this tick <- last tick)
                this.If = this.Yc;                          // playersThisTick count = previous view count
                for (int i = 0; i < this.If; i++) {
                    this.Zg[i] = this.rg[i];                // players[] <- playersLast[]
                }
                this.mg.i(-2231);                          // align bit reader to byte boundary
                this.Lf = this.mg.f(-106, 11);             // region origin X (11 bits)   (obf: Lf = localRegionX)
                this.sh = this.mg.f(-106, 13);             // region origin Y (13 bits)   (obf: sh = localRegionY)
                int localAnim = this.mg.f(-82, 4);         // local player facing/anim (4 bits)
                boolean regionChanged = this.loadRegion(this.sh, this.Lf, false); // obf: a(this.sh,this.Lf,false)
                this.Lf -= this.Qg;                        // subtract scene origin -> tile-local coords
                this.sh -= this.zg;
                int worldX = this.Lf * this.Ug + 64;       // Ug = tile size in world units; +64 = tile centre
                int worldY = this.sh * this.Ug + 64;
                this.Yc = 0;                               // reset this-tick player count
                if (regionChanged) {                       // snap camera-follow target to new position
                    this.wi.e = 0;
                    this.wi.o = 0;
                    this.wi.i = this.wi.k[0] = worldX;
                    this.wi.K = this.wi.F[0] = worldY;
                }
                // (re)create the local player entity at the new tile
                this.wi = this.addPlayer(worldY, this.Zc, worldX, -56, localAnim); // obf: d(worldY,Zc,worldX,-56,anim)
                int otherPlayers = this.mg.f(-69, 8);      // count of other visible players (8 bits)

                // --- movement deltas for the players already in view ---
                for (int p = 0; p < otherPlayers; p++) {
                    ta player = this.Zg[p + 1];
                    int hasUpdate = this.mg.f(-112, 1);    // 1 bit: did this player move/turn?
                    if (hasUpdate == 0) {
                        this.rg[this.Yc++] = player;        // unchanged -> carry over
                        continue;
                    }
                    int moved = this.mg.f(-95, 1);         // 1 bit: walked (1) vs. only-turned (0)
                    if (moved == 0) {                      // only changed facing direction
                        int dir = this.mg.f(-69, 2);       // 2 bits
                        if (dir == 3) continue;             // 3 = "removed from view"
                        player.D = (dir << 2) + this.mg.f(-98, 2); // pack facing: hi 2 bits + lo 2 bits
                        this.rg[this.Yc++] = player;
                        continue;
                    }
                    // walked one tile in one of 8 compass directions
                    int dir = this.mg.f(-87, 3);
                    int wp = player.o;                      // current waypoint slot
                    int wx = player.k[wp];
                    int wy = player.F[wp];
                    if (dir == 2 || dir == 1 || dir == 3) wx += this.Ug;  // E component
                    if (dir == 6 || dir == 5 || dir == 7) wx -= this.Ug;  // W component
                    if (dir == 4 || dir == 3 || dir == 5) wy += this.Ug;  // S component
                    player.o = wp = (wp + 1) % 10;          // advance ring of 10 waypoints
                    player.D = dir;
                    if (dir == 0 || dir == 1 || dir == 7) wy -= this.Ug;  // N component
                    player.k[wp] = wx;
                    player.F[wp] = wy;
                    this.rg[this.Yc++] = player;
                }

                // --- newly-appeared players (absolute coords) ---
                while (this.mg.k(-31874) + 24 < length * 8) {  // while bits remain in this packet
                    int serverIndex = this.mg.f(-120, 11);     // 11-bit player server index
                    int dx = this.mg.f(-96, 5);                // signed 5-bit X offset from local
                    int dy;
                    if (dx > 15) dx -= 32;
                    if ((dy = this.mg.f(-90, 5)) > 15) dy -= 32;
                    int anim = this.mg.f(-97, 4);
                    int ay = (dy + this.sh) * this.Ug + 64;
                    int ax = (this.Lf + dx) * this.Ug + 64;
                    this.addPlayer(ay, serverIndex, ax, -112, anim);  // obf: d(ay,idx,ax,-112,anim)
                }
                this.mg.j(25505);                              // finalize/clamp bit reader
                return;
            }

            // ---- 234 SEND_UPDATE_PLAYERS: per-player update stream (bubbles/chat/damage/projectile/appearance) ----
            case 234: {
                int count = this.mg.f(255);                    // number of player updates (16-bit)
                for (int u = 0; u < count; u++) {
                    int idx = this.mg.f(255);                  // player server index
                    ta player = this.We[idx];                  // obf: We = npcsCache; here re-used as player cache by index
                    int type = this.mg.h(20869);               // update sub-type
                    if (type == 0) {                           // speech bubble holding an item
                        int itemId = this.mg.a((byte)104);     // NB: rev235 reads a byte here
                        int itemId2 = this.mg.a((byte)104);
                        if (player != null) {
                            player.h = itemId2;                // bubble item id
                            player.w = this.nc;                // bubble timer marker
                            player.z = -1;
                            player.a = itemId;
                        }
                    } else if (type == 1) {                    // incoming projectile (cast/ranged) at this player
                        int projectileSprite = this.mg.f(255);
                        int targetIdx = this.mg.f(255);
                        if (player != null) {
                            player.h = targetIdx;
                            player.w = this.nc;
                            player.z = -1;
                            player.a = projectileSprite;
                        }
                    } else if (type == 2) {                    // combat: damage taken + current/max hits
                        int damage = this.mg.a((byte)104);
                        int curHits = this.mg.a((byte)104);
                        int maxHits = this.mg.a((byte)104);
                        if (player != null) {
                            player.u = damage;                 // damageTaken
                            player.G = maxHits;
                            player.B = curHits;
                            if (this.wi == player) {           // local player took damage
                                this.oh[3] = curHits;          // skillCurrent[Hits]
                                this.cg[3] = maxHits;          // skillBase[Hits]
                                this.mh = false;               // close any open dialog box
                                this.Oh = false;
                            }
                            player.d = 200;                    // combat timer
                        }
                    } else if (type == 3) {                    // bubble holding an item (alt encoding)
                        int itemId = this.mg.f(255);
                        if (player != null) {
                            player.E = 150;
                            player.j = itemId;
                        }
                    }
                    // (sub-types 4..7 — chat / full appearance — handled by the dedicated 79/91 streams)
                }
                return;
            }

            // ---- 79 SEND_NPC_COORDS: nearby-NPC movement stream ----
            case 79: {
                // double-buffer the in-view NPC list
                this.qj = this.de;                             // npcsLastCount = previous view count
                this.de = 0;
                for (int i = 0; i < this.qj; i++) {
                    this.Ff[i] = this.Tb[i];                   // npcs[] <- npcsLast[]
                }
                this.mg.i(-2231);                              // align reader
                int inView = this.mg.f(-87, 8);                // count of NPCs already in view

                // --- movement deltas for NPCs in view ---
                for (int n = 0; n < inView; n++) {
                    ta npc = this.Ff[n];
                    int hasUpdate = this.mg.f(-127, 1);
                    if (hasUpdate == 0) {
                        this.Tb[this.de++] = npc;
                        continue;
                    }
                    int moved = this.mg.f(-72, 1);
                    if (moved == 0) {                          // only turned
                        int dir = this.mg.f(-109, 2);
                        if (dir == 3) continue;                // removed from view
                        npc.D = this.mg.f(-127, 2) + (dir << 2);
                        this.Tb[this.de++] = npc;
                        continue;
                    }
                    // walked one tile
                    int dir = this.mg.f(-114, 3);
                    int wp = npc.o;
                    int wx = npc.k[wp];
                    int wy = npc.F[wp];
                    if (dir == 2 || dir == 1 || dir == 3) wx += this.Ug;
                    if (dir == 6 || dir == 5 || dir == 7) wx -= this.Ug;
                    if (dir == 4 || dir == 3 || dir == 5) wy += this.Ug;
                    npc.D = dir;
                    npc.o = wp = (wp + 1) % 10;
                    if (dir == 0 || dir == 1 || dir == 7) wy -= this.Ug;
                    npc.k[wp] = wx;
                    npc.F[wp] = wy;
                    this.Tb[this.de++] = npc;
                }

                // --- newly-appeared NPCs (absolute coords + type) ---
                while (this.mg.k(-31874) + 34 < length * 8) {
                    int serverIndex = this.mg.f(-104, 12);     // 12-bit npc server index
                    int dx = this.mg.f(-68, 5);
                    int dy;
                    if (dx > 15) dx -= 32;
                    if ((dy = this.mg.f(-111, 5)) > 15) dy -= 32;
                    int anim = this.mg.f(-74, 4);
                    int ax = (dx + this.Lf) * this.Ug + 64;
                    int ay = this.Ug * (this.sh + dy) + 64;
                    int npcTypeId = this.mg.f(-108, 10);       // 10-bit NPC type id
                    if (npcTypeId >= la.d) npcTypeId = 24;     // clamp to valid range (la.d = npc-def count)
                    this.addNpc(anim, npcTypeId, ax, (byte)127, ay, serverIndex); // obf: a(anim,type,ax,127,ay,idx)
                }
                this.mg.j(25505);
                return;
            }

            // ---- 104 SEND_UPDATE_NPC: per-NPC update stream (chat / combat / projectile) ----
            case 104: {
                int count = this.mg.f(255);
                for (int u = 0; u < count; u++) {
                    int idx = this.mg.f(255);
                    ta npc = this.We[idx];                     // obf: We = npcsCache
                    int type = this.mg.h(20869);
                    if (type == 1) {                           // NPC said something
                        int speakerIdx = this.mg.f(255);       // who it spoke to (for filtering)
                        String message = ia.a(this.mg, false); // decode scrambled chat
                        if (npc != null) {
                            npc.I = 150;                       // message timeout
                            npc.n = message;
                            if (this.wi.b == speakerIdx) {     // spoken to the local player
                                this.addMessageToChat(false, null, 0,
                                    e.Mb[npc.t] + il[12] + npc.n, 3, 0, null, il[20]); // "<npcName>: <msg>"
                            }
                        }
                    } else if (type == 2) {                    // NPC combat: damage + current/max hits
                        int damage = this.mg.a((byte)104);
                        int curHits = this.mg.a((byte)104);
                        int maxHits = this.mg.a((byte)104);
                        if (npc != null) {
                            npc.u = damage;
                            npc.G = maxHits;
                            npc.d = 200;
                            npc.B = curHits;
                        }
                    }
                }
                return;
            }

            // ---- 48 SEND_SCENERY_HANDLER: add/remove ground scenery models for this region ----
            case 48: {
                while (this.mg.w < length) {                   // walk packet payload (byte cursor mg.w)
                    if (this.mg.a((byte)104) == 255) {         // 255 = "removal run" marker
                        this.mg.w--;                           // un-read the marker
                        // remove scenery at this region-tile anchor; compact the arrays.
                        // (coords are stored ×8; >> 3 collapses them to region-tile units.)
                        int anchorX = this.Lf + this.mg.h(20869) >> 3;
                        int anchorY = this.sh + this.mg.h(20869) >> 3;
                        int kept = 0;
                        for (int s = 0; s < this.eh; s++) {     // eh = active scenery count
                            int rx = (this.Se[s] >> 3) - anchorX;
                            int ry = (this.ye[s] >> 3) - anchorY;
                            if (rx == 0 && ry == 0) {           // this scenery removed
                                this.Ek.a(this.hg[s], -1);      // world.removeModel
                                this.Hh.a(this.vc[s], this.Se[s], this.ye[s], 4081); // scene.remove
                            } else {                            // keep -> compact down
                                if (s != kept) {
                                    this.hg[kept] = this.hg[s];
                                    this.hg[kept].rb = kept;
                                    this.Se[kept] = this.Se[s];
                                    this.ye[kept] = this.ye[s];
                                    this.vc[kept] = this.vc[s];
                                    this.bg[kept] = this.bg[s];
                                }
                                kept++;
                            }
                        }
                        this.eh = kept;
                        continue;
                    }
                    // --- add a scenery object ---
                    int typeId = this.mg.f(255);
                    int x = this.Lf + this.mg.h(20869);
                    int y = this.sh + this.mg.h(20869);
                    int floorDir = this.Hh.b(x, y, -75);       // scene floor-direction query at this tile
                    // NOTE: the original inlines the GameModel build here (allocate from
                    // objectModels[], orient by floorDir, place into Scene + World). That
                    // geometry detail belongs to the `scene` group's addSceneObject — only
                    // the resulting bookkeeping is reproduced below for clarity.
                    this.addSceneObject(typeId, x, y, floorDir, 0, 0, 0, 0); // obf: a(...) — see scene group
                    this.Se[this.eh] = x;
                    this.ye[this.eh] = y;
                    this.vc[this.eh] = typeId;
                    this.bg[this.eh] = floorDir;
                    this.eh++;
                }
                return;
            }

            // ---- 91 SEND_BOUNDARY_HANDLER: add/remove wall (boundary) models for this region ----
            case 91: {
                while (this.mg.w < length) {
                    if (this.mg.a((byte)104) == 255) {         // removal run
                        this.mg.w--;
                        int anchorX = this.Lf + this.mg.h(20869) >> 3;
                        int anchorY = this.sh + this.mg.h(20869) >> 3;
                        int kept = 0;
                        for (int wIdx = 0; wIdx < this.Ah; wIdx++) {  // Ah = active wall count
                            int rx = (this.Zf[wIdx] >> 3) - anchorX;
                            int ry = (this.Ni[wIdx] >> 3) - anchorY;
                            if (rx == 0 && ry == 0) {
                                // (removed; arrays compacted below)
                            } else if (kept != wIdx) {
                                this.Zf[kept] = this.Zf[wIdx];
                                this.Ni[kept] = this.Ni[wIdx];
                                this.Gj[kept] = this.Gj[wIdx];
                                this.Le[kept] = this.Le[wIdx];
                                kept++;
                            } else {
                                kept++;
                            }
                        }
                        this.Ah = kept;
                        continue;
                    }
                    // --- add a wall/boundary ---
                    int dir = this.mg.f(255);
                    int x = this.Lf + this.mg.h(20869);
                    int y = this.sh + this.mg.h(20869);
                    this.addWallModel(0, 0, x, y, 0, 0, dir);  // obf: b(int x7) — see scene group
                    this.Zf[this.Ah] = x;
                    this.Ni[this.Ah] = y;
                    this.Gj[this.Ah] = dir;
                    this.Le[this.Ah] = 0;
                    this.Ah++;
                }
                return;
            }

            // ---- 99 SEND_GROUND_ITEM_HANDLER: add/remove ground items for this region ----
            case 99: {
                while (this.mg.w < length) {
                    if (this.mg.a((byte)104) == 255) {         // removal run
                        this.mg.w--;
                        int anchorX = this.Lf + this.mg.h(20869) >> 3;
                        int anchorY = this.sh + this.mg.h(20869) >> 3;
                        int kept = 0;
                        for (int g = 0; g < this.hf; g++) {     // hf = active ground-item count
                            int rx = (this.Jd[g] >> 3) - anchorX;
                            int ry = (this.yk[g] >> 3) - anchorY;
                            if (rx == 0 && ry == 0) {           // removed
                                this.Ek.a(this.rd[g], -1);
                                this.Hh.a(true, this.Hj[g], this.yk[g], this.Jd[g], this.Ng[g]);
                            } else {
                                if (kept != g) {
                                    this.rd[kept] = this.rd[g];
                                    this.rd[kept].rb = kept + 10000;
                                    this.Jd[kept] = this.Jd[g];
                                    this.yk[kept] = this.yk[g];
                                    this.Hj[kept] = this.Hj[g];
                                    this.Ng[kept] = this.Ng[g];
                                }
                                kept++;
                            }
                        }
                        this.hf = kept;
                        continue;
                    }
                    // --- add or remove a single ground item ---
                    int itemId = this.mg.f(255);
                    int x = this.Lf + this.mg.h(20869);
                    int y = this.sh + this.mg.h(20869);
                    int dir = this.mg.h(20869);
                    if ((itemId & 0x8000) != 0) {              // high bit set -> "remove this item"
                        itemId &= 0x7FFF;
                        int kept = 0;
                        for (int g = 0; g < this.hf; g++) {
                            if (this.Jd[g] == x && this.yk[g] == y && this.Hj[g] == itemId) {
                                this.Ek.a(this.rd[g], -1);
                                this.Hh.a(true, this.Hj[g], this.yk[g], this.Jd[g], this.Ng[g]);
                            } else {
                                if (kept != g) {
                                    this.rd[kept] = this.rd[g];
                                    this.rd[kept].rb = kept + 10000;
                                    this.Jd[kept] = this.Jd[g];
                                    this.yk[kept] = this.yk[g];
                                    this.Hj[kept] = this.Hj[g];
                                    this.Ng[kept] = this.Ng[g];
                                }
                                kept++;
                            }
                        }
                        this.hf = kept;
                    } else {                                   // add
                        this.Hh.a(y, itemId, dir, x, 11715);   // scene.placeGroundItem
                        this.rd[this.hf] = this.buildEntityModel(true, y, itemId, x, dir, this.hf);
                        this.Jd[this.hf] = x;
                        this.yk[this.hf] = y;
                        this.Ng[this.hf] = itemId;
                        this.Hj[this.hf++] = dir;
                    }
                }
                return;
            }

            // ---- 53 SEND_INVENTORY: full inventory contents ----
            case 53: {
                this.lc = this.mg.a((byte)104);                // item count
                for (int i = 0; i < this.lc; i++) {
                    int raw = this.mg.f(255);                  // bit15 = wielded, bits0..14 = item id
                    this.vf[i] = ib.a(raw, 32767);             // inventoryItems[i] = raw & 0x7FFF
                    this.Aj[i] = raw / 32768;                  // inventoryEquipped[i] = (raw >> 15)
                    if (fa.e[raw & 32767] != 0) {              // stackable/noted -> has a quantity
                        this.xe[i] = this.mg.c(103);           // inventoryCount[i] (var-length int)
                    } else {
                        this.xe[i] = 1;
                    }
                }
                return;
            }

            // ---- 90 SEND_INVENTORY_UPDATEITEM: single inventory slot changed ----
            case 90: {
                int amount = 1;
                int slot = this.mg.a((byte)104);
                int raw = this.mg.f(255);                      // bit15 = wielded
                if (fa.e[raw & 32767] != 0) {                  // stackable -> read amount
                    amount = this.mg.c(103);
                }
                this.vf[slot] = ib.a(raw, 32767);              // item id
                this.Aj[slot] = raw / 32768;                   // wielded flag
                this.xe[slot] = amount;
                if (slot >= this.lc) this.lc = slot + 1;       // grow inventory count if needed
                return;
            }

            // ---- 123 SEND_INVENTORY_REMOVE_ITEM: remove a slot, shift the rest down ----
            case 123: {
                int slot = this.mg.a((byte)104);
                this.lc--;
                for (int i = slot; i < this.lc; i++) {
                    this.vf[i] = this.vf[i + 1];
                    this.xe[i] = this.xe[i + 1];
                    this.Aj[i] = this.Aj[i + 1];
                }
                return;
            }

            // ---- 156 SEND_STATS: all skill levels + xp ----
            case 156: {
                for (int s = 0; s < 18; s++) this.oh[s] = this.mg.a((byte)104); // skillCurrent
                for (int s = 0; s < 18; s++) this.cg[s] = this.mg.a((byte)104); // skillBase
                for (int s = 0; s < 18; s++) this.Ak[s] = this.mg.b(-129);      // skillXp (signed 16-bit, *? — see 159)
                this.ii = this.mg.a((byte)104);               // quest points
                return;
            }

            // ---- 159 SEND_STAT: a single skill changed ----
            case 159: {
                int s = this.mg.a((byte)104);
                this.oh[s] = this.mg.a((byte)104);            // current level
                this.cg[s] = this.mg.a((byte)104);            // base level
                this.Ak[s] = this.mg.b(-129);                // xp
                return;
            }

            // ---- 33 SEND_EXPERIENCE: a single skill's raw xp ----
            case 33: {
                int s = this.mg.a((byte)104);
                this.Ak[s] = this.mg.b(-129);
                return;
            }

            // ---- 153 SEND_EQUIPMENT_STATS: armour/weapon aim/power bonuses ----
            case 153: {
                for (int i = 0; i < 5; i++) this.Fc[i] = this.mg.a((byte)104);
                return;
            }

            // ---- 59 SEND_APPEARANCE_SCREEN: open "design your character" appearance editor ----
            case 59:
                this.Kg = true;                               // show char-design screen
                return;

            // ---- 213 SEND_APPEARANCE_KEEPALIVE: no-op keepalive ----
            case 213:
                return;

            // ---- 245 SEND_OPTIONS_MENU_OPEN: an in-game multiple-choice question dialog ----
            case 245: {
                this.Ph = true;                               // options menu visible
                int n = this.Id = this.mg.a((byte)104);       // number of options
                for (int i = 0; i < n; i++) {
                    this.ah[i] = this.mg.c((byte)-44);         // option text
                }
                return;
            }

            // ---- 252 SEND_OPTIONS_MENU_CLOSE ----
            case 252:
                this.Ph = false;
                return;

            // ---- 25 SEND_WORLD_INFO: world/membership/region metadata at login ----
            case 25:
                this.Ub = true;
                this.Zc = this.mg.f(255);                     // local player server index
                this.Ki = this.mg.f(255);                     // plane width
                this.sk = this.mg.f(255);                     // plane index base
                this.bc = this.mg.f(255);                     // plane height
                this.rc = this.mg.f(255);                     // planes-per-region
                this.sk -= this.bc * this.rc;                 // compute origin plane
                return;

            // ---- 83 SEND_DEATH: player died ----
            case 83:
                this.rk = 250;                                // death animation timer
                return;

            // ---- 211 SEND_REMOVE_WORLD_ENTITY: bulk wall/scenery re-cull around a moved anchor ----
            // For each of (length-1)/4 entries the server gives a new anchor tile; the client
            // re-derives which boundary/scenery models are still in view and compacts the
            // Zf/Ni/Gj/Le (wall) and Se/ye/vc/bg (scenery) arrays accordingly. The exact
            // model add/remove geometry is the `scene` group's job; only the loop shape is
            // reproduced here.
            case 211: {
                int count = (length - 1) / 4;
                for (int u = 0; u < count; u++) {
                    int anchorX = this.Lf + this.mg.a(false) >> 3; // obf: mg.a(false) = read short
                    int anchorY = this.sh + this.mg.a(false) >> 3;
                    // ... re-cull walls/scenery/ground-items against (anchorX, anchorY),
                    //     removing entries whose region-tile delta is (0,0) and compacting
                    //     the parallel arrays (same pattern as opcodes 48 / 91 / 99 above).
                }
                return;
            }

            // NOTE (trade/duel/shop/bank cluster, opcodes 92/176/97/6/162/15/137/101/20/172/
            // 210/253/30/225/128): the per-field writes below were read directly from the
            // decompiled bodies and are faithful, but the opcode->window pairing in this
            // cluster differs from the OpenRSC generator's naming (this client uses the
            // `Hk`/`cj`/`Lk` field-set for the DUEL window and the `Pj`/`Lg`/`wj` set for the
            // TRADE window — the reverse of the generator's 92/176 labels). Treat the
            // SEND_* names here as best-effort; the field semantics are the ground truth.

            // ---- 92: open the DUEL window (Hk=duel-open) ----
            case 92: {
                int idx = this.mg.f(255);
                if (this.We[idx] != null) this.cj = this.We[idx].c; // opponent name hash
                this.Hk = true;                               // duel window open
                this.Lk = 0;                                  // their stake count
                this.mf = 0;
                this.Mi = false;                              // their-accepted
                this.md = false;                              // your-accepted
                return;
            }

            // ---- 128: close shop + duel windows (Xj/Hk) ----
            case 128:
                this.Xj = false;
                this.Hk = false;
                return;

            // ---- 97: the opponent's DUEL stake (zj/Dd) ----
            case 97: {
                this.Lk = this.mg.a((byte)104);               // their stake count
                for (int i = 0; i < this.Lk; i++) {
                    this.zj[i] = this.mg.f(255);              // item id
                    this.Dd[i] = this.mg.b(-129);            // amount
                }
                this.md = false;                              // reset accepted flags (stake changed)
                this.Mi = false;
                return;
            }

            // ---- 162: DUEL your-accepted flag (md) ----
            case 162: {
                this.md = this.mg.a((byte)104) == 1;
                return;
            }

            // ---- 101: options / bank-tab list (Rj/Jf/vi) ----
            case 101: {
                this.uk = true;
                int n = this.mg.a((byte)104);                 // entry count
                int mode = this.mg.h(20869);
                this.Nh = this.mg.a((byte)104);
                this.xk = this.mg.a((byte)104);
                this.Pf = this.mg.a((byte)104);
                for (int i = 0; i < 40; i++) this.Rj[i] = -1; // clear all 40 slots
                for (int i = 0; i < n; i++) {
                    this.Rj[i] = this.mg.f(255);
                    this.Jf[i] = this.mg.f(255);
                    this.vi[i] = this.mg.f(255);
                }
                this.uk = false;
                return;
            }

            // ---- 15: accepted flag for the open trade/duel (Mi) ----
            case 15: {
                this.Mi = this.mg.h(20869) == 1;
                return;
            }

            // ---- 20: open the SHOP window (Xj=shop-open; Lc/Bi stock, Vb/Me base amounts) ----
            case 20: {
                this.Hk = false;
                this.Xj = true;                               // shop open
                this.Vi = false;
                this.re = this.mg.c((byte)-44);              // shop header/flags string
                this.nh = this.mg.a((byte)104);              // stock size
                for (int i = 0; i < this.nh; i++) {
                    this.Lc[i] = this.mg.f(255);             // item id
                    this.Bi[i] = this.mg.b(-129);           // amount in stock
                }
                this.Ui = this.mg.a((byte)104);              // base-amount list size
                for (int i = 0; i < this.Ui; i++) {
                    this.Vb[i] = this.mg.f(255);
                    this.Me[i] = this.mg.b(-129);
                }
                return;
            }

            // ---- 30: the 4 trade-confirm boolean flags (fd/Yi/vd/ff) ----
            case 30:
                this.fd = this.mg.a((byte)104) == 2;
                this.Yi = this.mg.a((byte)104) == 2;
                this.vd = this.mg.a((byte)104) == 1;
                this.ff = this.mg.a((byte)104) == 1;
                this.ke = false;
                this.ki = false;
                return;

            // ---- 137: accepted flag for the open trade/duel (Mi) ----
            case 137: {
                this.Mi = this.mg.h(20869) == 1;
                return;
            }

            // ---- 42 SEND_BANK_OPEN: open the bank ----
            case 42: {
                this.Fe = true;                               // bank open
                this.fj = this.mg.a((byte)104);               // stored item count
                this.Gi = this.mg.a((byte)104);               // bank capacity
                for (int i = 0; i < this.fj; i++) {
                    this.ci[i] = this.mg.f(255);             // item id
                    this.Xe[i] = this.mg.c(103);            // amount (var-length)
                }
                this.drawHelpMenu(108);                       // obf: C(108) — refresh panel
                return;
            }

            // ---- 203 SEND_BANK_CLOSE ----
            case 203:
                this.Fe = false;
                return;

            // ---- 249 SEND_BANK_UPDATE: single bank slot changed ----
            case 249: {
                int slot = this.mg.a((byte)104);
                int itemId = this.mg.f(255);
                int amount = this.mg.c(103);
                if (amount == 0) {                            // removed -> shift down
                    this.fj--;
                    for (int i = slot; i < this.fj; i++) {
                        this.ci[i] = this.ci[i + 1];
                        this.Xe[i] = this.Xe[i + 1];
                    }
                } else {
                    this.ci[slot] = itemId;
                    this.Xe[slot] = amount;
                    if (slot >= this.fj) this.fj = slot + 1;
                }
                this.drawHelpMenu(-103);                       // obf: C(-103)
                return;
            }

            // ---- 176: open the TRADE window (Pj=trade-open; Lg=partner) ----
            case 176: {
                int idx = this.mg.f(255);
                if (this.We[idx] != null) this.Lg = this.We[idx].c; // trade partner name hash
                this.ke = false;                              // their-accepted
                this.vd = false;
                this.ki = false;                              // your-accepted
                this.ff = false;
                this.fd = false;
                this.Pj = true;                               // trade window open
                this.Yi = false;
                this.wj = 0;                                  // their offer count
                this.Ke = 0;                                  // your offer count
                return;
            }

            // ---- 6: the other player's TRADE offer (zc/of) ----
            case 6: {
                this.wj = this.mg.a((byte)104);               // their item count
                for (int i = 0; i < this.wj; i++) {
                    this.zc[i] = this.mg.f(255);              // item id
                    this.of[i] = this.mg.b(-129);            // amount
                }
                this.ke = false;                              // reset accepted flags (offer changed)
                this.ki = false;
                return;
            }

            // ---- 210: their-accepted flag (ke) ----
            case 210: {
                this.ke = this.mg.h(20869) == 1;
                return;
            }

            // ---- 253: your-accepted flag (ki) ----
            case 253: {
                this.ki = this.mg.h(20869) == 1;
                return;
            }

            // ---- 172: the second "confirm" window (your/their items + the 4 trade stats) ----
            case 172: {
                this.Cd = false;
                this.dd = true;                               // show confirm screen
                this.Pj = false;
                this.Uc = this.mg.c((byte)-44);              // confirmation text
                // your side
                this.Ve = this.mg.a((byte)104);
                for (int i = 0; i < this.Ve; i++) {
                    this.xj[i] = this.mg.f(255);
                    this.kf[i] = this.mg.b(-129);
                }
                // their side
                this.Nj = this.mg.a((byte)104);
                for (int i = 0; i < this.Nj; i++) {
                    this.xi[i] = this.mg.f(255);
                    this.th[i] = this.mg.b(-129);
                }
                this.Sh = this.mg.a((byte)104);
                this.gh = this.mg.a((byte)104);
                this.Cc = this.mg.a((byte)104);
                this.Rc = this.mg.a((byte)104);
                return;
            }

            // ---- 225: close the duel window (Hk) ----
            case 225:
                this.Hk = false;
                return;

            // ---- 240 SEND_GAME_SETTINGS: server-pushed privacy/mouse/sound toggles ----
            case 240:
                this.Kh = this.mg.a((byte)104) == 1;          // (e.g. auto-camera)
                this.Yh = this.mg.a((byte)104) == 1;          // (e.g. one-mouse-button)
                this.ne = this.mg.a((byte)104) == 1;          // (e.g. sound on)
                return;

            // ---- 111 SEND_ON_TUTORIAL: tutorial-island flag ----
            case 111:
                this.Kd = this.mg.a((byte)104) != 0;
                return;

            // ---- 206 SEND_PRAYERS_ACTIVE: which prayers are currently on ----
            case 206: {
                for (int i = 0; i < length - 1; i++) {
                    this.fi[i] = this.mg.h(20869) == 1;       // obf: fi[] re-used as prayer-on flags
                }
                return;
            }

            // ---- 5 SEND_QUESTS: quest-completion flags ----
            case 5: {
                for (int i = 0; i < length - 1; i++) {
                    this.bk[i] = this.mg.h(20869) == 1;       // obf: bk[] = quest-complete flags
                }
                return;
            }

            // ---- 204 SEND_PLAY_SOUND: play a named sound effect ----
            case 204: {
                String soundName = this.mg.c((byte)-44);
                this.playSound(-73, soundName);               // obf: a(-73, name)
                return;
            }

            // ---- 36 SEND_BUBBLE: teleport / telegrab / iban-magic bubble effect ----
            case 36: {
                int itemId = this.mg.a((byte)104);
                int x = this.mg.h(20869) + this.Lf;
                int y = this.mg.h(20869) + this.sh;
                this.Oc[this.el] = itemId;                    // bubble item
                this.oe[this.el] = 0;                         // bubble timer
                this.Sc[this.el] = x;
                this.gi[this.el] = y;
                this.el++;                                    // active bubble count
                return;
            }

            // ---- 182 SEND_WELCOME_INFO: "Welcome" box (last login IP/date + unread messages) ----
            case 182:
                if (this.Dc) return;                          // already shown
                this.ce = this.mg.b(-129);                    // days since last login
                this.hi = this.mg.f(255);                     // unread-messages count
                this.Sb = this.mg.a((byte)104);               // recovery-set days
                this.id = this.mg.f(255);                     // last-login IP (packed)
                this.Oh = true;
                this.ve = null;
                this.Dc = true;
                return;

            // ---- 89 SEND_BOX2: server message box (not closeable) ----
            case 89:
                this.Cj = this.mg.c((byte)-44);              // box text
                this.mh = true;                               // box visible
                this.Wk = false;                             // not "closeable" style
                return;

            // ---- 222 SEND_BOX: server message box (closeable) ----
            case 222:
                this.Cj = this.mg.c((byte)-44);
                this.mh = true;
                this.Wk = true;
                return;

            // ---- 114 SEND_FATIGUE: current fatigue value ----
            case 114:
                this.vg = this.mg.f(255);                     // fatigue (0..7500)
                return;

            // ---- 244 SEND_SLEEP_FATIGUE: fatigue while sleeping ----
            case 244:
                this.pg = this.mg.f(255);
                return;

            // ---- 117 SEND_SLEEPSCREEN: enter the sleep CAPTCHA screen ----
            case 117:
                if (!this.Qk) this.pg = this.vg;              // seed sleep-fatigue from current
                this.e = "";                                  // clear sleep-word input
                this.Qk = true;                               // sleeping
                this.Cb = "";
                this.li.a((byte)-118, this.mg.F, this.Eh + 1); // load CAPTCHA bitmap from packet bytes
                this.Zj = null;                               // clear "incorrect" prompt
                return;

            // ---- 84 SEND_STOPSLEEP: wake up ----
            case 84:
                this.Qk = false;
                return;

            // ---- 194 SEND_SLEEPWORD_INCORRECT ----
            case 194:
                this.Zj = il[55];                             // "...you entered the wrong word..."
                return;

            // ---- 52 SEND_SYSTEM_UPDATE: countdown to server restart ----
            case 52:
                this.kc = this.mg.f(255) * 32;                // ticks until update (×32)
                return;

            default:
                // Unknown opcode: original logs and drops it (no state change).
                return;
            }
        } catch (RuntimeException e) {
            // Authentic recovery: log the bad packet and force a clean disconnect.
            mb.a(0x1FFFFF, null, il[57] + opcode + il[60] + length);
            this.onStopGame(true);                            // obf: a(true,31)
        }
    }

    /**
     * Right-click MENU-ACTION dispatcher (the real role of this method; the skeleton's
     * "handleSceneUpdates" name is a misnomer). The selected context-menu entry encodes
     * an action code in {@code zh} (the menu MessageList) plus up to four operands; this
     * turns it into the matching OUTGOING action packet via {@code clientStream}.
     *
     * Action codes: 200/300/400 = object op (1st/2nd/examine), 600/610 = boundary op,
     * 650/660 = ground-item examine/take, 700/710/720/725/715 = player op (attack/trade/
     * follow/duel/cast), 800/810/805 = npc op, 2805/2806/2810/2820 = npc cast/use/talk,
     * 900/920 = walk/walk-to-tile, 1000 = item-on-object. The 28xx codes are social
     * (report/add-friend/add-ignore/send-message).
     *
     * obf: void b(boolean,int)   params: (signedShortFlag, menuIndex) — both consumed as
     * MessageList read selectors.
     */
    private void handleSceneUpdates(boolean signedFlag, int menuIndex) {
        try {
            // Pull the selected menu entry's action code + operands out of the menu list.
            int action = this.zh.a(-110, menuIndex);          // action code (200,300,…)
            int a1 = this.zh.a(true, menuIndex);              // operand 1 (id / dx)
            int a2 = this.zh.a((byte)97, menuIndex);          // operand 2 (dy)
            int a3 = this.zh.a(menuIndex, (byte)22);          // operand 3
            int a4 = this.zh.a(menuIndex, signedFlag);        // operand 4
            int a5 = this.zh.b(true, menuIndex);              // operand 5 (item slot)
            String str = this.zh.c(menuIndex, -4126);         // string operand (target name)

            if (action == 200) {                              // object: use 1st option
                this.walkTo((byte)10, this.sh, a2, a1, true, this.Lf); // obf: a(10,sh,a2,a1,true,Lf)
                this.Jh.b(249, 0);                            // -> opcode 249 (OP_OBJECT_1) [outgoing]
                this.Jh.f.e(393, a1 + this.Qg);
                this.Jh.f.e(393, a2 + this.zg);
                this.Jh.f.e(393, a3);
                this.Jh.f.e(393, a4);
                this.Jh.b(21294);                             // flush
                this.af = -1;
            }
            if (action == 210) {                              // object: use 2nd option
                this.walkTo((byte)10, this.sh, a2, a1, true, this.Lf);
                this.Jh.b(53, 0);                             // -> outgoing object-action 2
                this.Jh.f.e(393, this.Qg + a1);
                this.Jh.f.e(393, this.zg + a2);
                this.Jh.f.e(393, a3);
                this.Jh.f.e(393, a4);
                this.Jh.b(21294);
                this.Bh = -1;
            }
            if (action == 220) {                              // object: examine
                this.walkTo((byte)10, this.sh, a2, a1, true, this.Lf);
                this.Jh.b(247, 0);
                this.Jh.f.e(393, a1 + this.Qg);
                this.Jh.f.e(393, this.zg + a2);
                this.Jh.f.e(393, a3);
                this.Jh.b(21294);
            }
            if (action == 3600 || action == 3200) {           // object/scenery examine -> show def text
                this.addMessageToChat(false, null, 0, ga.b[a1], 0, 0, null, null);
            }
            if (action == 300) {                              // wall/boundary: use 1st option
                this.handleGameClick(false, a1, a2, a3);      // obf: a(false,a1,a2,a3)
                this.Jh.b(180, 0);
                this.Jh.f.e(393, this.Qg + a1);
                this.Jh.f.e(393, this.zg + a2);
                this.Jh.f.c(a3, 110);
                this.Jh.f.e(393, a4);
                this.Jh.b(21294);
                this.af = -1;
            }
            if (action == 310) {                              // wall: use 2nd option
                this.handleGameClick(false, a1, a2, a3);
                this.Jh.b(161, 0);
                this.Jh.f.e(393, a1 + this.Qg);
                this.Jh.f.e(393, a2 + this.zg);
                this.Jh.f.c(a3, -110);
                this.Jh.f.e(393, a4);
                this.Jh.b(21294);
                this.Bh = -1;
            }
            if (action == 320) {                              // wall: examine
                this.handleGameClick(signedFlag, a1, a2, a3);
                this.Jh.b(14, 0);
                this.Jh.f.e(393, a1 + this.Qg);
                this.Jh.f.e(393, a2 + this.zg);
                this.Jh.f.c(a3, 54);
                this.Jh.b(21294);
            }
            if (action == 2300) {                             // wall: use-item-on
                this.handleGameClick(false, a1, a2, a3);
                this.Jh.b(127, 0);
                this.Jh.f.e(393, this.Qg + a1);
                this.Jh.f.e(393, a2 + this.zg);
                this.Jh.f.c(a3, -60);
                this.Jh.b(21294);
            }
            if (action == 3300) {                             // wall examine -> show def text
                this.addMessageToChat(false, null, 0, ub.b[a1], 0, 0, null, null);
            }
            if (action == 400) {                              // ground item: pick up
                this.drawBox(5126, a4, a1, a2, a3);           // obf: b(5126,a4,a1,a2,a3) — path/select helper
                this.Jh.b(99, 0);
                this.Jh.f.e(393, a1 + this.Qg);
                this.Jh.f.e(393, this.zg + a2);
                this.Jh.f.e(393, a5);
                this.Jh.b(21294);
                this.af = -1;
            }
            if (action == 410) {                              // ground item: use-with-item
                this.drawBox(5126, a4, a1, a2, a3);
                this.Jh.b(115, 0);
                this.Jh.f.e(393, a1 + this.Qg);
                this.Jh.f.e(393, a2 + this.zg);
                this.Jh.f.e(393, a5);
                this.Jh.b(21294);
                this.Bh = -1;
            }
            if (action == 420) {                              // ground item: examine target
                this.drawBox(5126, a4, a1, a2, a3);
                this.Jh.b(136, 0);
                this.Jh.f.e(393, this.Qg + a1);
                this.Jh.f.e(393, a2 + this.zg);
                this.Jh.b(21294);
            }
            if (action == 2400) {                             // ground item: cast spell on
                this.drawBox(5126, a4, a1, a2, a3);
                this.Jh.b(79, 0);
                this.Jh.f.e(393, this.Qg + a1);
                this.Jh.f.e(393, this.zg + a2);
                this.Jh.b(21294);
            }
            if (action == 3400) {                             // ground item examine -> def text
                this.addMessageToChat(false, null, 0, la.f[a1], 0, 0, null, null);
            }
            if (action == 600) {                              // ground item (simple): pick up
                this.Jh.b(4, 0);
                this.Jh.f.e(393, a1);
                this.Jh.f.e(393, a2);
                this.Jh.b(21294);
                this.af = -1;
            }
            if (action == 610) {
                this.Jh.b(91, 0);
                this.Jh.f.e(393, a1);
                this.Jh.f.e(393, a2);
                this.Jh.b(21294);
                this.Bh = -1;
            }
            if (action == 620) {
                this.Jh.b(170, 0);
                this.Jh.f.e(393, a1);
                this.Jh.b(21294);
            }
            if (action == 630) {
                this.Jh.b(169, 0);
                this.Jh.f.e(393, a1);
                this.Jh.b(21294);
            }
            if (action == 640) {
                this.Jh.b(90, 0);
                this.Jh.f.e(393, a1);
                this.Jh.b(21294);
            }
            if (action == 650) {                              // ground item: examine (local def)
                this.Bh = a1;
                this.qc = 0;
                this.ig = ac.x[this.vf[this.Bh]];
            }
            if (action == 660) {                              // ground item: examine -> show text
                this.Jh.b(246, 0);
                this.Jh.f.e(393, a1);
                this.Jh.b(21294);
                this.qc = 0;
                this.Bh = -1;
                this.addMessageToChat(false, null, 0, il[511] + ac.x[this.vf[a1]], 7, 0, null, null);
            }
            // 700..725: player actions (attack/trade/follow/duel/cast). Walk toward the
            // target's tile first, then send the op.
            if (action == 700) {                              // attack player
                ta target = this.getPlayer(a1, (byte)-123);   // obf: b(a1,-123)
                int ty = (target.i - 64) / this.Ug;
                int tx = (target.K - 64) / this.Ug;
                this.walkToAction(tx, ty, this.sh, this.Lf, true, 8); // obf: a(tx,ty,sh,Lf,true,8)
                this.Jh.b(50, 0);
                this.Jh.f.e(393, a1);
                this.Jh.f.e(393, a2);
                this.Jh.b(21294);
                this.af = -1;
            }
            if (action == 710) {                              // trade player
                ta target = this.getPlayer(a1, (byte)-123);
                int ty = (target.i - 64) / this.Ug;
                int tx = (target.K - 64) / this.Ug;
                this.walkToAction(tx, ty, this.sh, this.Lf, true, 8);
                this.Jh.b(135, 0);
                this.Jh.f.e(393, a1);
                this.Jh.f.e(393, a2);
                this.Jh.b(21294);
                this.Bh = -1;
            }
            if (action == 720) {                              // follow player
                ta target = this.getPlayer(a1, (byte)-123);
                int ty = (target.i - 64) / this.Ug;
                int tx = (target.K - 64) / this.Ug;
                this.walkToAction(tx, ty, this.sh, this.Lf, true, 8);
                this.Jh.b(153, 0);
                this.Jh.f.e(393, a1);
                this.Jh.b(21294);
            }
            if (action == 725) {                              // duel player
                ta target = this.getPlayer(a1, (byte)-123);
                int ty = (target.i - 64) / this.Ug;
                int tx = (target.K - 64) / this.Ug;
                this.walkToAction(tx, ty, this.sh, this.Lf, true, 8);
                this.Jh.b(202, 0);
                this.Jh.f.e(393, a1);
                this.Jh.b(21294);
            }
            if (action == 2715 || action == 715) {            // cast spell / use item on player
                ta target = this.getPlayer(a1, (byte)-123);
                int ty = (target.i - 64) / this.Ug;
                int tx = (target.K - 64) / this.Ug;
                this.walkToAction(tx, ty, this.sh, this.Lf, true, 8);
                this.Jh.b(190, 0);
                this.Jh.f.e(393, a1);
                this.Jh.b(21294);
            }
            if (action == 3700) {                             // player examine -> def text
                this.addMessageToChat(false, null, 0, ba.ac[a1], 0, 0, null, null);
            }
            // 800..810: npc actions
            if (action == 800) {                              // attack npc
                ta npc = this.getNpc(a1, 220);                // obf: d(a1,220)
                int ty = (npc.i - 64) / this.Ug;
                int tx = (npc.K - 64) / this.Ug;
                this.walkToAction(tx, ty, this.sh, this.Lf, true, 8);
                this.Jh.b(229, 0);
                this.Jh.f.e(393, a1);
                this.Jh.f.e(393, a2);
                this.Jh.b(21294);
                this.af = -1;
            }
            if (action == 810) {                              // talk-to npc
                ta npc = this.getNpc(a1, 220);
                int ty = (npc.i - 64) / this.Ug;
                int tx = (npc.K - 64) / this.Ug;
                this.walkToAction(tx, ty, this.sh, this.Lf, true, 8);
                this.Jh.b(113, 0);
                this.Jh.f.e(393, a1);
                this.Jh.f.e(393, a2);
                this.Jh.b(21294);
                this.Bh = -1;
            }
            if (action == 2805 || action == 805) {            // cast spell / use item on npc
                ta npc = this.getNpc(a1, 220);
                int ty = (npc.i - 64) / this.Ug;
                int tx = (npc.K - 64) / this.Ug;
                this.walkToAction(tx, ty, this.sh, this.Lf, true, 8);
                this.Jh.b(171, 0);
                this.Jh.f.e(393, a1);
                this.Jh.b(21294);
            }
            if (action == 2806) {                             // npc: 1st option
                this.Jh.b(103, 0);
                this.Jh.f.e(393, a1);
                this.Jh.b(21294);
            }
            if (action == 2810) {                             // npc: 2nd option
                this.Jh.b(142, 0);
                this.Jh.f.e(393, a1);
                this.Jh.b(21294);
            }
            if (action == 2820) {                             // npc: examine
                this.Jh.b(165, 0);
                this.Jh.f.e(393, a1);
                this.Jh.b(21294);
            }
            // 28xx social actions (operate on the picked player name `str`)
            if (action == 2833) {                             // send public/quick message
                this.Cb = "";
                this.Vf = 1;
                this.e = str;
            }
            if (action == 2831) {                             // add friend
                this.sendAddFriend(97, str);                  // obf: b(97,str)
            }
            if (action == 2832) {                             // add ignore
                this.sendAddIgnore(str, (byte)5);             // obf: a(str,5)
            }
            if (action == 2830) {                             // send private message (open entry)
                this.Qd = str;
                this.x = "";
                this.Bj = 2;
                this.Ob = "";
            }
            if (action == 900) {                              // walk to clicked tile (then face)
                this.walkToAction(a2, a1, this.sh, this.Lf, true, 8);
                this.Jh.b(158, 0);
                this.Jh.f.e(393, a1 + this.Qg);
                this.Jh.f.e(393, this.zg + a2);
                this.Jh.f.e(393, a3);
                this.Jh.b(21294);
                this.af = -1;
            }
            if (action == 920) {                              // walk to clicked tile (no path-send)
                this.walkToAction(a2, a1, this.sh, this.Lf, false, 8);
                if (this.xh == -24) this.xh = 24;             // tutorial walk-acknowledged
            }
            if (action == 1000) {                             // close shop
                this.Jh.b(137, 0);
                this.Jh.f.e(393, a1);
                this.Jh.b(21294);
                this.af = -1;
            }
            if (action == 4000) {                             // cancel / clear pending action
                this.af = -1;
                this.Bh = -1;
            }
        } catch (RuntimeException e) {
            throw i.a(e, il[510] + signedFlag + ',' + menuIndex + ')');
        }
    }

    /**
     * Social / private-message packet sub-dispatcher (broadly matches the skeleton's
     * "onFriendUpdate" intent). Reads the opcode it was given ({@code opcode}) and applies
     * the friend-list, ignore-list, private-message and server-message packets; any opcode
     * it does not own is forwarded to {@link #handlePacket}.
     *
     * obf: void a(int,int,int)   params: (a, b, opcode). `a`/`b` are residual scratch
     * operands; the opcode is re-read from the stream header via {@code clientStream.a(507,opcode)}.
     */
    private void onFriendUpdate(int a, int b, int opcode) {
        try {
            opcode = this.Jh.a(507, opcode);                 // re-resolve opcode from stream header

            switch (opcode) {

            // ---- 131 SEND_SERVER_MESSAGE ----
            case 131: {
                int msgType = this.mg.a((byte)104);          // chat tab / message-type id
                int infoFlags = this.mg.a((byte)104);        // bit0 = has sender, bit1 = has colour
                String message = this.mg.c((byte)-44);
                String sender = null, senderDup = null, colour = null;
                if ((infoFlags & 1) != 0) sender = this.mg.c((byte)-44);
                if ((infoFlags & 1) != 0) senderDup = this.mg.c((byte)-44); // authentic duplicate
                if ((infoFlags & 2) != 0) colour = this.mg.c((byte)-44);
                this.addMessageToChat(false, sender, 0, message, msgType, 0, senderDup, colour);
                break;                                       // -> shared logout-check tail
            }

            // ---- 4 SEND_LOGOUT_REQUEST_CONFIRM: server allows logout -> send CONFIRM_LOGOUT + clean up ----
            case 4:
                this.sendConfirmLogoutAck(true, a - 56);     // obf: a(true, a-56) — sends opcode 31, tears down
                break;                                       // -> shared logout-check tail

            // ---- 183 SEND_CANT_LOGOUT: "can't log out right now" ----
            case 183:
                this.sendConfirmLogout((byte)-65);           // obf: g(-65)
                break;                                       // -> shared logout-check tail

            // ---- 189 SEND_28_BYTES_UNUSED: skip a fixed 28-byte block (verifies CRC) ----
            case 189:
                this.mg.w += 28;
                if (this.mg.e(-422797528)) {                 // CRC/length check
                    b.a(this.mg, 26628, this.mg.w - 28);
                }
                break;                                       // -> shared logout-check tail

            // ---- 165 SEND_LOGOUT: server forced/confirmed logout -> local cleanup (no packet) ----
            case 165:
                this.sendConfirmLogoutAck(false, 31);        // obf: a(false, 31) — reset session state only
                break;                                       // -> shared logout-check tail

            // ---- 51 SEND_PRIVACY_SETTINGS: the 4 chat-privacy toggle bytes ----
            // (block/public/private/trade-and-duel privacy levels)
            case 51:
                this.De = this.mg.a((byte)104);              // block-chat privacy
                this.dc = this.mg.a((byte)104);              // public-chat privacy
                this.Vg = this.mg.a((byte)104);              // private-chat privacy
                this.ui = this.mg.a((byte)104);              // trade/duel privacy
                break;                                       // -> shared logout-check tail

            // ---- 149 SEND_FRIEND_UPDATE: a friend logged in/out (and online world) ----
            case 149: {
                String name = this.mg.c((byte)-44);          // current name
                String formerName = this.mg.c((byte)-44);    // former name (for rename match)
                int flags = this.mg.a((byte)104);            // bit0: 0 => insert-if-missing; bit2: online
                boolean insertIfMissing = (flags & 1) == 0;
                boolean nowOnline = (flags & 4) != 0;
                String onlineWorld = null;
                if (nowOnline) onlineWorld = this.mg.c((byte)-44);
                for (int f = 0; f < n.g; f++) {              // n.g = friends count
                    if (insertIfMissing) {
                        if (ua.h[f].equals(name)) {          // matched by current name -> update status
                            if (ac.z[f] == null && nowOnline)
                                this.addMessageToChat(false, null, 0, name + il[9], 5, 0, null, null); // "has logged in"
                            if (ac.z[f] != null && !nowOnline)
                                this.addMessageToChat(false, null, 0, name + il[8], 5, 0, null, null); // "has logged out"
                            cb.c[f] = formerName;
                            ac.z[f] = onlineWorld;           // null = offline marker
                            Fj[f] = flags;
                            this.sortDrawList(51);           // obf: v(51) — re-sort friends list
                            return;
                        }
                    } else if (ua.h[f].equals(formerName)) {  // matched by former name -> rename in place
                        if (ac.z[f] == null && nowOnline)
                            this.addMessageToChat(false, null, 0, name + il[9], 5, 0, null, null);
                        if (ac.z[f] != null && !nowOnline)
                            this.addMessageToChat(false, null, 0, name + il[8], 5, 0, null, null);
                        ua.h[f] = name;
                        cb.c[f] = formerName;
                        ac.z[f] = onlineWorld;
                        Fj[f] = flags;
                        this.sortDrawList(50);               // obf: v(50)
                        return;
                    }
                }
                // not present yet -> append
                ua.h[n.g] = name;
                cb.c[n.g] = formerName;
                ac.z[n.g] = onlineWorld;
                Fj[n.g] = flags;
                n.g++;
                this.sortDrawList(66);                        // obf: v(66)
                break;                                       // -> shared logout-check tail
            }

            // ---- 237 SEND_UPDATE_IGNORE_LIST_BECAUSE_NAME_CHANGE ----
            case 237: {
                String newName = this.mg.c((byte)-44);
                String newName2 = this.mg.c((byte)-44);
                if (newName2.length() == 0) newName2 = newName;
                String oldWorld = this.mg.c((byte)-44);
                String oldName = this.mg.c((byte)-44);
                if (oldName.length() == 0) oldName = newName;
                boolean updateExisting = this.mg.a((byte)104) == 1;
                for (int i = 0; i < db.g; i++) {             // db.g = ignore count
                    if (updateExisting) {
                        if (ia.a[i].equals(oldName)) {        // rename an existing ignore entry
                            l.c[i] = newName;
                            ia.a[i] = newName2;
                            ia.g[i] = oldWorld;
                            ua.wb[i] = oldName;
                            return;
                        }
                    } else if (ia.a[i].equals(newName2)) {
                        return;                               // already present
                    }
                }
                // append a new ignore entry
                if (updateExisting) {
                    System.out.println(il[7] + oldName + il[5]);
                    return;
                }
                l.c[db.g] = newName;
                ia.a[db.g] = newName2;
                ia.g[db.g] = oldWorld;
                ua.wb[db.g] = oldName;
                db.g++;
                break;                                       // -> shared logout-check tail
            }

            // ---- 109 SEND_IGNORE_LIST: full ignore list refresh ----
            case 109: {
                db.g = this.mg.a((byte)104);                  // ignore count
                for (int i = 0; i < db.g; i++) {
                    l.c[i] = this.mg.c((byte)-44);
                    ia.a[i] = this.mg.c((byte)-44);
                    ia.g[i] = this.mg.c((byte)-44);
                    ua.wb[i] = this.mg.c((byte)-44);
                }
                break;                                       // -> shared logout-check tail
            }

            // ---- 120 SEND_PRIVATE_MESSAGE: incoming private message (de-duplicated) ----
            case 120: {
                String fromName = this.mg.c((byte)-44);
                String fromFormer = this.mg.c((byte)-44);
                int icon = this.mg.a((byte)104);              // moderator/icon sprite
                long messageId = this.mg.g(0);                // 8-byte message id (world+counter)
                String message = ia.a(this.mg, false);        // decode scrambled body
                // drop if we've already seen this exact message id (anti-duplicate ring)
                for (int i = 0; i < 100; i++) {
                    if (this.Zd[i] == messageId) return;
                }
                this.Zd[this.Ag] = messageId;                 // record id in the ring
                this.Ag = (this.Ag + 1) % 100;
                this.addMessageToChat(icon == 2, fromName, 0, message, 1, icon, fromFormer, null);
                break;                                       // -> shared logout-check tail
            }

            // ---- 87 SEND_PRIVATE_MESSAGE_SENT: echo of a PM you sent ----
            case 87: {
                String toName = this.mg.c((byte)-44);
                String message = ia.a(this.mg, false);
                this.addMessageToChat(false, toName, 0, message, 2, 0, toName, null);
                break;                                       // -> shared logout-check tail
            }

            // ---- everything else: hand off to the master dispatcher ----
            default:
                this.handlePacket(opcode, (byte)41, b);       // obf: b(opcode,41,b)
                break;                                       // -> shared logout-check tail
            }

            // Shared tail (faithful to bytecode): every path that did not early-return
            // falls here. Unless the caller's mode operand `a` is 87, request a logout.
            // (`a` is a mode selector set by the GameShell packet-read caller; the 87 sentinel
            //  suppresses the auto-logout for the private-message-flush path.)
            if (a != 87) {
                this.requestLogout(56);                       // obf: B(56) — send opcode 102 (LOGOUT)
            }
        } catch (RuntimeException e) {
            throw i.a(e, il[6] + a + ',' + b + ',' + opcode + ')');
        }
    }

    /**
     * Social-entry DIALOG renderer (the real role; the skeleton's "applyAppearanceUpdate"
     * name is a misnomer — this parses no packet). Draws the add-friend / add-ignore /
     * send-private-message popup: a titled box with the friends or ignore list, plus
     * "Cancel"/"OK" buttons, and handles clicks/typing inside it. {@code activeOnly}
     * selects friends- vs. ignore-list source; {@code suppressInput} pins the panel.
     *
     * Drawing uses {@code surface} (obf li) and the {@code panelLogin} widget container
     * (obf zk). Text comes from the {@code STRINGS} pool (obf il).
     *
     * obf: void a(boolean,boolean)   params: (handleInput, suppressInput)
     */
    private void applyAppearanceUpdate(boolean handleInput, boolean suppressInput) {
        try {
            int boxX = this.li.u - 199;                       // surface width - 199 (centre the box)
            int titleY = 36;
            this.li.b(-1, this.tg + 5, 3, boxX - 49);         // clear/frame backdrop

            int boxW = 196;
            int boxH = 182;
            if (suppressInput) this.Be = -88;

            // header gradient (highlight the active tab)
            int leftColour = o.a(160, 9570, 160, 160);
            int rightColour = leftColour;
            if (this.pk == 0) {                               // tab 0 active (add-friend/ignore)
                rightColour = o.a(220, 9570, 220, 220);
            } else {
                leftColour = o.a(220, 9570, 220, 220);
            }
            this.li.c(128, boxX, 24, 0, titleY, boxW / 2, leftColour);
            this.li.c(128, boxX + boxW / 2, 24, 0, titleY, boxW / 2, rightColour);
            this.li.c(128, boxX, boxH - 24, 0, titleY + 24, boxW, o.a(220, 9570, 220, 220));
            // borders
            this.li.b(boxW, 0, boxX, titleY + 24, (byte)95);
            this.li.b(boxX + boxW / 2, titleY, 0, 24, 0);
            this.li.b(boxW, 0, boxX, titleY + boxH - 16, (byte)-113);
            // tab captions
            this.li.a(boxX + boxW / 4, il[260], 0, 0, 4, titleY + 16);
            this.li.a(boxX + boxW / 4 + boxW / 2, il[258], 0, 0, 4, titleY + 16);

            this.zk.c((byte)-82, this.Hi);                    // reset the dialog's MessageList rows

            // --- populate the list with the appropriate names ---
            if (this.pk == 0) {                               // FRIENDS list
                for (int f = 0; f < n.g; f++) {               // n.g = friends count
                    boolean online = (Fj[f] & 2) != 0;
                    String statusColour;
                    if (online) statusColour = il[27];        // green
                    else if ((Fj[f] & 4) != 0) statusColour = il[20]; // (intermediate)
                    else statusColour = il[10];               // grey/offline
                    // truncate the name so the row fits 120px
                    String name = ua.h[f];
                    int len = ua.h[f].length();
                    int cut = 0;
                    while (this.li.a(1, 111, name) > 120) {
                        name = ua.h[f].substring(0, len - (++cut)) + il[261]; // "..."
                    }
                    this.zk.a(f, null, 49, 0, null, statusColour + name + il[262], this.Hi);
                }
            }
            if (this.pk == 1) {                               // IGNORE list
                for (int i = 0; i < db.g; i++) {              // db.g = ignore count
                    String name = l.c[i];
                    int len = l.c[i].length();
                    int cut = 0;
                    while (this.li.a(1, 100, name) > 120) {
                        name = l.c[i].substring(0, len - (++cut)) + il[261];
                    }
                    this.zk.a(i, null, 60, 0, null, il[20] + name + il[262], this.Hi);
                }
            }
            this.zk.a((byte)-43);                             // finalize list layout
            this.nj = -1;                                     // hovered friends row
            this.wk = -1;                                     // hovered ignore row

            // --- "ignore" tab caption + hover highlight ---
            int ignoreLabelColour;
            if (this.I >= boxX && this.I < boxX + boxW
                    && this.xb > boxH + (titleY - 16) && this.xb < boxH + titleY) {
                ignoreLabelColour = 0xFFFF00;                 // yellow when hovered
                if (this.pk == 0) {                           // friends tab active -> ignore-list rows clickable
                    int row = this.zk.b(this.Hi, 17050);
                    if (row >= 0 && this.I < 489 && this.I > 429) {
                        this.wk = -(row + 2);                 // mark ignore-row hovered
                    } else {
                        this.wk = row;
                    }
                }
            } else {
                ignoreLabelColour = 0xFFFFFF;
            }
            this.li.a(boxX + boxW / 2, il[259], ignoreLabelColour, 0, 1, boxH + titleY - 3); // "Ignore"

            // --- "friends" tab caption + hover highlight ---
            if (this.pk == 1) {                               // ignore tab active -> friends rows clickable
                int row = this.zk.b(this.Hi, 17050);
                if (row >= 0 && this.I < 489 && this.I > 429) {
                    this.nj = -(row + 2);
                } else {
                    this.nj = row;
                }
            }
            this.li.a(boxX + boxW / 2, il[266], 0xFFFFFF, 0, 1, titleY + 35); // (sub caption)

            int friendsLabelColour;
            if (this.I > boxX && this.I < boxX + boxW
                    && boxH + titleY - 16 < this.xb && this.xb < boxH + titleY) {
                friendsLabelColour = 0xFFFFFF;
            } else {
                friendsLabelColour = 0xFFFF00;
            }
            this.li.a(boxX + boxW / 2, il[265], friendsLabelColour, 0, 1, titleY + boxH - 3); // "Friends"
            this.li.a(boxX + boxW / 2, il[263], 0xFFFFFF, 0, 1, titleY + 35);

            // --- input handling (skipped when only redrawing) ---
            if (!handleInput) return;
            int my = this.xb - 36;                            // mouse Y relative to box
            int mx = this.I + 199 - this.li.u;                // mouse X relative to box
            if (mx < 0 || my < 0) return;
            if (mx > 196 || my > 182) return;

            this.zk.b(this.Bb, my + 36, -9989, this.Qb, mx + this.li.u - 199); // route mouse into panel

            // tab switching by clicking the header tabs
            if (mx <= 24 && this.Cf == 1) {                   // clicked left tab area
                if (my < 98 && this.pk == 1) {                // -> switch to friends tab
                    this.pk = 0;
                    this.zk.e(this.Hi, 14);
                }
                if (my > 98 && this.pk == 0) {                // -> switch to ignore tab
                    this.pk = 1;
                    this.zk.e(this.Hi, 14);
                }
            }
            // clicking a friends row -> open send-message entry for that name
            if (this.Cf == 1 && this.pk == 0) {
                int row = this.zk.b(this.Hi, 17050);
                if (row >= 0 && this.I < 489 && this.I > 429) {
                    this.sendRemoveFriend(ua.h[row], (byte)69); // obf: b(name,69) — remove on right area
                }
            }
            if ((Fj[/*selected*/0] & 4) != 0) {
                this.Bj = 2;                                  // open PM entry
                this.Qd = ua.h[0];
                this.Ob = "";
                this.x = "";
            }
            // clicking an ignore row -> remove that ignore
            if (this.Cf == 1 && this.pk == 1) {
                int row = this.zk.b(this.Hi, 17050);
                if (row >= 0 && this.I < 489 && this.I > 429) {
                    this.sendRemoveIgnore((byte)-15, ia.a[row]); // obf: a(-15, name)
                }
            }
            // bottom-left button -> open add-friend entry
            if (my > 166 && this.Cf == 1 && this.pk == 0) {
                this.Cb = "";
                this.e = "";
                this.Bj = 1;
            }
            // bottom-right button -> open add-ignore entry
            if (my > 166 && this.Cf == 1 && this.pk == 1) {
                this.Cb = "";
                this.Bj = 3;
                this.e = "";
            }
            this.Cf = 0;                                       // consume the click
        } catch (RuntimeException e) {
            throw i.a(e, il[264] + handleInput + ',' + suppressInput + ')');
        }
    }
