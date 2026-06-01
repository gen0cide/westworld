// ===== scene =====
// Group: Mudclient 3D scene/world build & render, camera, models.
// All methods are instance methods of class Mudclient (package client, extends GameShell),
// 4-space indented as if inside the class body.
//
// Stripping applied to every method:
//   - opaque predicate:  boolean bl = client.OPAQUE_FALSE;  (always false, dead)
//   - profiling counters: ++<StaticCounter>;                 (dead, removed)
//   - exception wrapper:  catch(RuntimeException e){ throw ErrorHandler.a(e,"sig"); }
//   - anti-tamper guards: if(param != <magic>) return;      (dead, removed)
//   - junk before shifts: XOR masks on shift amounts        (removed)
//
// Field naming follows MUDCLIENT_SKELETON.md; class naming follows NAMING.md.
//   Key: Hh=scene(Scene/lb), Ek=world(World/k), li=surface(SurfaceSprite/ba),
//        Zg=players, rg=npcsInView(*), Tb=playersInView(*), te=playersCache,
//        We=npcsCache, de=playerCount(*), Yc=npcCount(*), Ug=tileSize.
//
// (*) NOTE: The skeleton's Tb/rg and de/Yc field assignments appear to be
//     swapped relative to actual usage.  addPlayer writes to Tb[de++], and
//     addNpc writes to rg[Yc++]; getPlayer searches Tb[0..de], getNpc searches
//     rg[0..Yc].  Names from the skeleton are kept as-is; see SKELETON.md for
//     the field table (field names Tb, rg, de, Yc may need revision).

    // -------------------------------------------------------------------------
    // drawWorld
    // -------------------------------------------------------------------------

    /**
     * Main 3D world render: build Scene from World, place entity models,
     * set camera, project polygons, blit to surface.
     * obf: void s(int)
     */
    private final void drawWorld(int param) {
        // This method failed to decompile in Vineflower (parsing failure).
        // CFR reconstruction follows; heavy block-label nesting stripped.
        //
        // High-level flow (from CFR output and oracle mudclient204.drawAboveHead /
        // drawGameScreen):
        //
        //   1. Dead/sleeping/out-of-area early-outs (rk, Hh.Z checks).
        //   2. Reset per-frame arrays (Ed[], Sj[]).
        //   3. Pull entity lists from World (objectModels[], wallModels[]).
        //   4. For each entity slot (n5 = 0..n6):
        //        a. Resolve the encoded entity kind from ca2.E[n3]:
        //             kind/10000 = type  (1=scenery, 2=wall/boundary, 3=npc)
        //             kind%10000 = local index
        //        b. Scenery (type==1): build right-click "Examine" option via
        //             scene.a(ye[n], ig+name, priority, Bh, vc[n], 410, bg[n],
        //                     "@"+l.a[id], Se[n])         [walk-here action slot]
        //             If held item (af>=0) and item has "use on" action:
        //               scene.a(420, ye[n], bg[n], Se[n], af, vc[n], label, useText)
        //             If combat NPC (Bh>=0):
        //               scene.a(2400, ...)                 [attack option]
        //             scene.a(id, 3400, false, il[51], label)  [examine]
        //             If spellbook open (af->spell, qb.e[af]==5):
        //               scene.a(ye[n], "Cast "+spell+" on", +65, af, vc[n], 400, ...)
        //             Mark Ed[n] = true (seen this tick).
        //        c. Wall/boundary (type==2): similar right-click menu build;
        //             "Open"/"Close" actions; Sj[n] = true.
        //        d. Ground item (type==3):
        //             Decode combat level delta vs local player for skull/vs text.
        //             "Attack"/"Talk-to"/"Pick up"/"Examine" options built.
        //        e. a(n, -12) — walk-to entity path request.
        //   5. If held item (af>=0, non-spell) add global "Drop" or "Cancel" option.
        //   6. If clicked ground entity (Hc=true):
        //        rf = Qg + Hh.q[slot]  (scene x → world x)
        //        Cg = zg + Hh.E[slot]  (scene z → world z)
        //        Build walk-to or interaction option at that tile.
        //
        // The outer container `zh` is the menu-option accumulator (Panel yd
        // or a separate menu list); `af` = currently selected inventory slot
        // (attack/use), `Bh` = local player server-index.
        //
        // CFR reconstruction (abbreviated, dead code stripped):

        if (scene.Z == false) return;  // scene not ready

        // iterate entity slots emitted by World
        int objectCount = world.b(0);           // number of placed entities
        GameModel[] caArray = world.b((byte)124);  // scene-object GameModels
        int[] objEntityIds = world.a((byte)104);   // encoded entity id per slot

        if (param != 2) nk = -82;              // reset hover target (not an overloaded redraw)

        for (int slot = 0; slot < objectCount; slot++) {
            // --- decode entity type and local index ---
            int entityId = objEntityIds[slot];
            GameModel ca2 = caArray[slot];

            // distance cull: skip if too far (>200 units from camera centre)
            int dist = scene.c(param ^ 0xFFFF95ED); // c() = polygon dist to eye
            if (dist > 200) continue;

            if (ca2 == null || !ca2.visible /* E[idx] bounds-check */) continue;

            // entity rb encodes: rb<10000 → scenery, rb>=10000 → wall/boundary
            int rb = ca2.rb;
            if (rb < 0) continue;  // not registered

            if (rb >= 10000) {
                // --- wall / boundary entity ---
                int wallIdx = rb - 10000;
                int npcId   = Ng[wallIdx];     // wallObjectId[]
                if (!Sj[wallIdx]) {            // not yet processed this frame
                    if (af >= 0 && EntityDef.commandTwo[npcId] != null) {
                        // "Use <item> with" interaction option
                        scene.a(300, yk[wallIdx], Hj[wallIdx], Jd[wallIdx],
                                60, af,
                                "@" + GameCharacter.r[npcId],
                                "Use " + BitBuffer.L[af] + " with");
                    }
                    if (Bh >= 0) {
                        scene.a(310, yk[wallIdx], Hj[wallIdx], Jd[wallIdx],
                                param ^ 0x42, Bh,
                                "@" + GameCharacter.r[npcId],
                                ig + STRINGS[38] + STRINGS[53]);
                    }
                    if (!EntityDef.commandOne[npcId].equalsIgnoreCase(STRINGS[33])) {
                        scene.a(Jd[wallIdx], (byte)22, 320,
                                EntityDef.commandOne[npcId],
                                "@" + GameCharacter.r[npcId],
                                Hj[wallIdx], yk[wallIdx]);
                    }
                    if (!EntityDef.commandTwo[npcId].equalsIgnoreCase(STRINGS[51])) {
                        scene.a(Jd[wallIdx], (byte)22, 2300,
                                EntityDef.commandTwo[npcId],
                                "@" + GameCharacter.r[npcId],
                                Hj[wallIdx], yk[wallIdx]);
                    }
                    scene.a(npcId, 3300, false, STRINGS[51], "@" + GameCharacter.r[npcId]);
                    Sj[wallIdx] = true;
                }
            } else {
                // --- scenery / object entity ---
                int sceneryIdx = rb;
                int objType    = vc[sceneryIdx];
                if (!Ed[sceneryIdx]) {  // not yet processed this frame
                    if (af < 0) {
                        // no held item: build standard "Examine" option
                        if (Bh >= 0) {
                            scene.a(ye[sceneryIdx],
                                    STRINGS[38] + ig + STRINGS[53],
                                    -104, Bh, vc[sceneryIdx], 410,
                                    bg[sceneryIdx],
                                    "@" + NameTable.a[objType],
                                    Se[sceneryIdx]);
                        }
                        // item-specific: "Examine"
                        if (!EntityDef.commandOne[objType].equalsIgnoreCase(STRINGS[33])) {
                            scene.a(420, ye[sceneryIdx], bg[sceneryIdx], Se[sceneryIdx],
                                    param ^ 0x6B, vc[sceneryIdx],
                                    "@" + NameTable.a[objType],
                                    EntityDef.commandOne[objType]);
                        }
                        if (!EntityDef.commandTwo[objType].equalsIgnoreCase(STRINGS[51])) {
                            scene.a(2400, ye[sceneryIdx], bg[sceneryIdx], Se[sceneryIdx],
                                    param ^ 0x7F, vc[sceneryIdx],
                                    "@" + NameTable.a[objType],
                                    EntityDef.commandTwo[objType]);
                        }
                        scene.a(objType, 3400, false, STRINGS[51], "@" + NameTable.a[objType]);
                    } else {
                        // spell or item held
                        if (EntityDef.commandOne[af] == 5) { // cast spell
                            scene.a(ye[sceneryIdx],
                                    "Cast " + BitBuffer.L[af] + " on",
                                    param + 65, af, vc[sceneryIdx], 400,
                                    bg[sceneryIdx], "@" + NameTable.a[objType],
                                    Se[sceneryIdx]);
                        }
                    }
                    Ed[sceneryIdx] = true;
                }
            }

            // walk-to: schedule path to this entity tile
            this.walkTo(slot, -12);  // -12 = internal sentinel
        }

        // if no held item and a valid target player exists, add attack option
        if (af >= 0 && EntityDef.commandOne[af] != -2) {
            scene.a(af, 1000, false, "Use " + BitBuffer.L[af] + " with", "");
        }

        // if a ground entity was clicked (Hc flag), set walk destination
        if (!Hc) return;
        int slot = /* clicked slot */0;  // from Hh.q/E
        rf = Qg + scene.q[slot];        // world-x = regionBase + scene-x
        Cg = zg + scene.E[slot];        // world-z = regionBase + scene-z
        if (af < 0) return;
        if (EntityDef.commandOne[af] != -7) return;
        // "Walk here" on scene ground
        scene.a(scene.q[slot], (byte)22, 900,
                "Use " + BitBuffer.L[af] + " on",
                "", af, scene.E[slot]);
        if (Bh < 0) return;
        scene.a(scene.q[slot], "", 920, STRINGS[54], scene.E[slot], 3296);
    }

    // -------------------------------------------------------------------------
    // loadRegion
    // -------------------------------------------------------------------------

    /**
     * (Re)load the terrain region centred on the given world tile.
     * Displays "Loading... Please wait" on first load or region shift.
     * Adjusts all entity coordinates when the region origin moves.
     * Returns true when a full load was performed, false if already in region.
     * obf: boolean a(int x, int z, boolean isUnderground)
     */
    private final boolean loadRegion(int x, int z, boolean isUnderground) {
        // Check for stream-down / fatal error
        if (~rk != -1) {              // rk != 0 → disconnected or error
            scene.Z = false;
            return false;
        }
        this.Ub = isUnderground;

        // --- region-cache hit: player is still in the same 48×48 chunk ---
        // yj == bc: floor level unchanged; bounds check against [Jg..Rk) × [Fi..Ne)
        z += Ki;   // apply region base offset
        x += sk;
        if (yj == bc
                && Jg < z && z < Rk
                && Fi < x && x < Ne) {
            scene.Z = true;
            return false;   // no reload needed
        }

        // --- full region reload ---
        // Show "Loading... Please wait"
        surface.a(256, STRINGS[676], 0xFFFFFF, 0, 1, 192);
        this.A(5);                               // repaint progress
        surface.a(graphics, Eb, 256, K);         // flush to AWT

        // save old region base
        int oldRegionX = Qg;
        int oldRegionZ = zg;

        // compute new 48×48 chunk origin (snap to grid)
        int chunkZ = (z + 24) / 48;
        int chunkX = (x + 24) / 48;
        Qg = chunkX * 48 - 48;
        zg = chunkZ * 48 - 48;

        yj = bc;   // latch current floor level

        // region boundary in world coords (±32 tiles from centre)
        Ne = chunkX * 48 + 32;
        Rk = chunkZ * 48 + 32;
        Fi = chunkX * 48 - 32;
        Jg = chunkZ * 48 - 32;

        // tell Scene its new world-space origin and floor
        scene.a(z, (byte)-90, x, yj);   // Scene.setOrigin(z, x, floor)

        // subtract player's sub-tile offset so coords stay relative to view
        zg -= sk;
        Qg -= Ki;

        // delta from old region base to new
        int deltaX = Qg - oldRegionX;
        int deltaZ = zg - oldRegionZ;

        // --- rebase all wall/boundary models ---
        // wallModels[] at indices [0..eh)
        for (int i = 0; i < eh; i++) {
            Se[i]       -= deltaX;   // wall X coord (scene-relative)
            ye[i]       -= deltaZ;   // wall Z coord
            int tileX    = Se[i];
            int tileZ    = ye[i];
            int wallType = vc[i];
            GameModel wallGm = hg[i];

            try {
                // look up model width/height for this wall type
                // bg[i]==0 or bg[i]==4: use normal width/height; else swap
                int modelW, modelH;
                if (bg[i] == 0 || bg[i] == 4) {
                    modelW = NameTable.g[wallType];   // obf: ub.g  = objectWidth
                    modelH = TextEncoder.f[wallType]; // obf: f.f   = objectHeight
                } else {
                    modelH = NameTable.g[wallType];
                    modelW = TextEncoder.f[wallType];
                }

                // mid-point in scene units (Ug = tileSize, typically 128)
                int midX = (tileX + tileX + modelW) * Ug / 2;
                int midZ = (tileZ + tileZ + modelH) * Ug / 2;

                // cull if outside the 96×96-tile view window
                if (tileX < 0 || tileZ < 0 || tileX >= 96 || tileZ >= 96) continue;

                // re-anchor model in World and reposition it
                world.a(wallGm, (byte)118);  // World.placeModel
                // translate model to terrain elevation
                wallGm.c(-scene.f(midX, midZ, 89), -123, midZ, midX);
                // register model with Scene at its wall slot
                scene.a(tileX, wallType, isUnderground, tileZ);
                // special: object 74 (ladder?) floats 480 units above ground
                if (wallType == 74) {
                    wallGm.a(0, 0, -480, true);
                }
            } catch (RuntimeException ex) {
                System.out.println(STRINGS[671] + ex.getMessage());
                System.out.println(STRINGS[672] + i + STRINGS[673] + wallGm);
                ex.printStackTrace();
            }
        }

        // --- rebase all NPC/mob models ---
        // npcModelCache[] at indices [0..hf)
        for (int i = 0; i < hf; i++) {
            Jd[i] -= deltaX;   // npc tile X
            yk[i] -= deltaZ;   // npc tile Z
            int tileX    = Jd[i];
            int tileZ    = yk[i];
            int npcType  = Ng[i];
            int npcAnim  = Hj[i];
            try {
                scene.a(tileZ, npcType, npcAnim, tileX, 11715); // Scene.placeNpc
                npcModelCache[i] = this.buildEntityModel(
                        isUnderground == false, tileZ, npcType, tileX, npcAnim, i);
            } catch (RuntimeException ex) {
                System.out.println(STRINGS[674] + ex.getMessage());
                ex.printStackTrace();
            }
        }

        // --- rebase all ground-item spots ---
        // Ah = groundItemCount
        for (int i = 0; i < Ah; i++) {
            Zf[i] -= deltaX;
            Ni[i] -= deltaZ;
        }

        // --- rebase NPCs (rg array, Yc count) ---
        // NOTE: In practice rg[] holds NPCs-in-view (see field note at top).
        for (int i = 0; i < Yc; i++) {
            GameCharacter npc = rg[i];
            npc.i -= Ug * deltaX;   // currentX (fixed-point, scale by tileSize)
            npc.K -= deltaZ * Ug;   // currentZ
            for (int wp = 0; wp <= npc.o; wp++) {
                npc.k[wp] -= Ug * deltaX;  // waypointsX
                npc.F[wp] -= deltaZ * Ug;  // waypointsZ
            }
        }

        // --- rebase players (Tb array, de count) ---
        // NOTE: In practice Tb[] holds players-in-view (see field note at top).
        for (int i = 0; i < de; i++) {
            GameCharacter pl = Tb[i];
            pl.K -= Ug * deltaZ;
            pl.i -= Ug * deltaX;
            for (int wp = 0; wp <= pl.o; wp++) {
                pl.k[wp] -= Ug * deltaX;
                pl.F[wp] -= deltaZ * Ug;
            }
        }

        scene.Z = true;
        return true;
    }

    // -------------------------------------------------------------------------
    // buildEntityModel
    // -------------------------------------------------------------------------

    /**
     * Build a GameModel for a wall/door entity from World definition data.
     * The model is a single quad (4 vertices) spanning from tile (x,z) to
     * (x+1,z+1) depending on direction; placed at terrain elevation.
     * obf: ca a(boolean isUnderground, int tileX, int wallType, int tileZ, int dir, int slot)
     */
    private final GameModel buildEntityModel(boolean isUnderground, int tileX, int wallType,
                                              int tileZ, int dir, int slot) {
        int modelColor  = ChatCipher.a[wallType];   // packed colour
        int modelBitmap = loginScreenBg[wallType];  // texture/bitmap id (Jk)
        int modelHeight = StreamBase.d[wallType];   // wall height (ib.d)

        GameModel ca2 = new GameModel(4, 1);

        // adjust tile extents based on wall direction (0-3):
        //   0/1 = E-W wall  (x → x+1)
        //   2/3 = N-S wall  (z → z+1)
        int x0 = tileX, x1 = tileX;
        int z0 = tileZ, z1 = tileZ;
        if (~dir == -2) x1 = tileX + 1;                          // dir==1: extend X
        if (dir  ==  0) z1 = tileZ + 1;                          // dir==0: extend Z
        if (~dir == -3) { x1 = tileX + 1; z1 = tileZ + 1; }     // dir==2: diagonal
        if (~dir == -4) { x1 = tileX + 1; z1 = tileZ + 1; }     // dir==3: diagonal alt

        // convert tile coords to scene units (Ug = tileSize, e.g. 128)
        x0 *= Ug;  x1 *= Ug;
        z0 *= Ug;  z1 *= Ug;

        // add 4 vertices: bottom-near, top-near, top-far, bottom-far
        int v0 = ca2.e(x0, z0, -scene.f(x0, z0, -35),  -126);  // bottom near
        int v1 = ca2.e(x0, z0, -scene.f(x0, z0, -103) - modelHeight, -126); // top near
        int v2 = ca2.e(x1, z1, -modelHeight + -scene.f(x1, z1, -77), -112); // top far
        int v3 = ca2.e(x1, z1, -scene.f(x1, z1, 96),   117);   // bottom far

        // if not underground, notify World to attach model
        if (!isUnderground) {
            this.a(119, 67, 26, 106, false, -100);  // internal flag setter
        }

        // build the single quad face with the wall colour/texture
        ca2.a(4, new int[]{v0, v1, v2, v3}, modelColor, modelBitmap, false);
        // set lighting / shading defaults
        ca2.a(-50, 60, -10, -50, false, 24, -95);

        // register with World if in visible tile range (0..95 × 0..95)
        if (tileZ >= 0 && tileX >= 0 && tileZ < 96 && tileX < 96) {
            world.a(ca2, (byte)118);   // World.addModel
        }
        // tag with slot id so the scene can look up this wall's slot
        ca2.rb = slot + 10000;   // wall slot uses +10000 offset (vs scenery at <10000)
        return ca2;
    }

    // -------------------------------------------------------------------------
    // addSceneObject
    // -------------------------------------------------------------------------

    /**
     * Register a scenery/object GameModel into the Scene at the given tile.
     * Handles animated objects (walking, combat attacks) for the local player.
     * obf: void a(int, int, int, int, int, int, int, int)
     *      (Vineflower failed; CFR reconstruction used)
     */
    final void addSceneObject(int tileX, int tileZ, int z2, int n3,
                               int n4, int n5, int n6, int playerIdx) {
        // playerIdx → Tb[playerIdx] = GameCharacter (player entity)
        GameCharacter player = Tb[playerIdx];  // playersInView[playerIdx]
        if (player.A == 255) return;           // entity slot unused / dead

        // compute walk-cycle frame from world clock and speed (ug)
        // n16 = (player.y - (ug+16)/32) & 7  — 3-bit walk step
        int walkStep = (player.y - (ug + 16) / 32) & 7;
        boolean flip = false;
        int frameBase = walkStep;

        // normalise diagonal walk steps (5→1, 6→2, 7→3)
        if (walkStep == 5) { flip = true; frameBase = 1; }
        else if (walkStep == 6) { flip = true; frameBase = 2; }
        else if (walkStep == 7) { flip = true; frameBase = 3; }

        // frame within sprite-row (sf[]=combat frame offsets, Pc[]=run, Og[]=?)
        int frameIdx = sf[player.x / 6 % 4] + 3 * frameBase;

        // combat animations override walk frame
        if (player.y == 8) {
            // combat: attack swing
            flip    = false;
            frameBase = 5;
            n6 -= n5 * EntityDef.combatDelay[player.t] / 100; // adjust Z by anim
            frameIdx = Pc[jk / EntityDef.attackSpeed[player.t] % 8] + 3 * frameBase;
        }
        if (player.y == 9) {
            // combat: block/recoil
            frameBase = 5;
            flip = true;
            n6 += EntityDef.combatDelay[player.t] * n5 / 100;
            frameIdx = Og[jk / EntityDef.attackSpeed[player.t] % 8] + 3 * frameBase;
        }

        // iterate over the 12 equipment slots (armour layers)
        for (int layer = 0; layer < 12; layer++) {
            int equipSlot = Tg[/* direction index */2][layer]; // body-part slot mapping
            int itemId    = player.m[equipSlot] - 1;          // equipped item id (-1 = none)
            if (itemId < 0) continue;                          // slot empty

            // look up sprite dimensions for this item/layer
            int spriteW = li.Eb[frameIdx - -(WorldEntity.g[itemId])]; // sprite width
            int spriteH = li.qb[frameIdx];                             // sprite height
            if (spriteW <= 0 || spriteH == 0) continue;

            // scale offset: centre sprite at entity position
            int offX = n5 * 0 / spriteH;   // horizontal alignment (centred)
            int offZ = 0  * n6 / spriteW;   // vertical alignment
            int drawW = n6 * li.Eb[frameIdx] / li.Eb[WorldEntity.g[itemId]];
            offX -= (n6 - drawW) / 2;

            // blit the equipment sprite to the surface at screen (tileX, tileZ)
            surface.a(tileZ, h.c[itemId], 0, false, 0,
                      sg + ua.Bb[itemId],
                      tileX, n6, n4, 1);
        }
    }

    // -------------------------------------------------------------------------
    // addWallModel
    // -------------------------------------------------------------------------

    /**
     * Add a wall/boundary sprite overlay to the 2D surface at a tile position.
     * In rev ~235 boundary walls are blitted as 2D sprites over the 3D render,
     * not as standalone 3D GameModels.
     * obf: void b(int x, int y, int z, int wallType, int height, int dir, int w, int screenY)
     *      client.CA(
     */
    final void addWallModel(int x, int y, int z, int wallType,
                             int height, int dir, int w, int screenY) {
        if (dir > -109) tj = 50;  // anti-tamper guard stripped; field touched but harmless

        // look up sprite data for this wall type
        int spriteOffset = ua.Bb[wallType] + sg;   // base sprite index + global sprite offset
        int charWidth    = h.c[wallType];           // TextEncoder char-width used as glyph width

        // blit the wall sprite to the 2D surface
        // li.a(screenY, charWidth, 0, false, 0, spriteOffset, x, height, z, 1)
        surface.a(screenY, charWidth, 0, false, 0, spriteOffset, x, height, z, 1);
    }

    // -------------------------------------------------------------------------
    // addGroundObject
    // -------------------------------------------------------------------------

    /**
     * Add a ground/floor decoration oval marker to the 2D surface.
     * Used for campfires, torches, and similar floor decorations.
     * obf: void a(int, int, int, int, int, int, int)
     *      (note: 7 params, different from 8-param addSceneObject)
     */
    final void addGroundObject(int x, int y, int z, int w, int h,
                                int decorType, int screenY, int n8) {
        if (n8 != 2) Dc = true;   // anti-tamper guard stripped; flag side-effect kept

        int decorShape = Oc[decorType];  // 0=circle, 1=oval
        int decorSize  = oe[decorType];  // radius/scale of marker

        if (decorShape == 0) {
            // Draw a green circle (ground ring) — e.g. campfire base
            // li.c(alpha, colour, radiusMinor, cx, radiusMajor, cy)
            int radiusMajor = 255 - decorSize * 1280;
            surface.c(255 - 5 * decorSize, -1057205208 /* olive-green packed */,
                      20 + decorSize * 2,
                      w / 2 + z,   // centre X
                      radiusMajor,
                      y + h / 2);  // centre Y
        }

        if (decorShape == 1) {
            // Draw a red oval marker — e.g. fire/lava pit
            int colourR = 0xFF0000 + 1280 * decorSize;
            surface.c(255 - 5 * decorSize, -1057205208,
                      decorSize + 10,
                      z + w / 2,
                      colourR,
                      y + h / 2);
        }
    }

    // -------------------------------------------------------------------------
    // buildTerrainTile
    // -------------------------------------------------------------------------

    /**
     * Build/render the player-sprite overlay for a single entity tile.
     * Despite the name, this actually renders the sprite layers (equipment,
     * body, head, etc.) for one player at their screen tile position.
     * obf: void b(int, int, int, int, int, int, int, int)
     *      (Vineflower failed; CFR reconstruction used)
     */
    final void buildTerrainTile(int n2, int n3, int n4, int n5,
                                 int n6, int n7, int n8, int playerIdx) {
        if (n4 != 20) this.e((byte)-115);  // anti-tamper guard stripped; side-effect kept

        GameCharacter player = rg[playerIdx]; // entity from in-view array
        if (player.A == 255) return;          // invisible/unused slot

        // compute walk-cycle step from world clock and speed
        int walkStep = (player.y - (ug + 16) / 32) & 7;
        boolean flip = false;
        int frameBase = walkStep;

        // normalise diagonal walk (5→3, 6→1, 7→2 with flip)
        if (~walkStep == -7) { flip = true; frameBase = 3; }
        else if (walkStep == 6) { flip = true; frameBase = 1; }
        else if (~walkStep != -8) { /* 7 */ frameBase = 2; flip = true; }

        // base frame index in sprite sheet
        int frameIdx = sf[player.x / 6 % 4] + 3 * frameBase;

        // combat animation overrides
        if (player.y == 8) {
            // attack swing
            flip     = false;
            frameBase = 5;
            n6 -= n5 * 5 / 100;   // adjust height for forward lean
            frameIdx = Pc[jk / 5 % 8] + 3 * frameBase;
        }
        if (player.y == 9) {
            // block / recoil
            flip     = true;
            frameBase = 5;
            n6 += 5 * n5 / 100;
            frameIdx = Og[jk / 6 % 8] + 3 * frameBase;
        }

        // iterate equipment slots and blit each layer
        for (int layer = 0; layer < 12; layer++) {
            int bodyPart = Tg[/* dir/row index */2][layer]; // body-part slot lookup
            int itemId   = player.m[bodyPart] - 1;         // 0-based item id
            if (itemId < 0) continue;                       // empty slot

            // compute per-item sprite dims and offsets
            int spriteFrameA = frameIdx + WorldEntity.g[itemId];
            int spriteFrameB = WorldEntity.g[itemId];
            int wA = li.Eb[spriteFrameA];
            int wB = li.qb[spriteFrameA];
            int wC = li.Eb[spriteFrameB];
            if (wA <= 0 || wB == 0 || wC == 0) continue;

            int offX = n5 * 0 / wB;
            int offZ = 0  * n6 / wA;
            int drawW = n6 * li.Eb[spriteFrameA] / wC;
            offX -= (n6 - drawW) / 2;

            // blit this equipment-layer sprite
            surface.a(n8, h.c[itemId], 0, false, 0,
                      sg + ua.Bb[itemId],
                      n3, n6, n2, 1);
        }
    }

    // -------------------------------------------------------------------------
    // getPlayer
    // -------------------------------------------------------------------------

    /**
     * Look up a player GameCharacter in the current-tick players array by server index.
     * Returns null if no player with that index is in view this tick.
     * Side-effect: sets Bf = -116 when called with a sentinel byte != -123.
     * obf: ta b(int serverIndex, byte sentinel)   client.AC(
     */
    private final GameCharacter getPlayer(int serverIndex, byte sentinel) {
        if (sentinel != -123) {
            Bf = -116;   // marks "fresh allocation needed" for the caller
        }
        // Tb[] = players in view this tick (de = count)
        for (int i = 0; i < de; i++) {
            if (serverIndex == Tb[i].b) {   // ta.b = serverIndex field
                return Tb[i];
            }
        }
        return null;
    }

    // -------------------------------------------------------------------------
    // addPlayer
    // -------------------------------------------------------------------------

    /**
     * Create or update a player entity at the given tile and add it to the
     * current-tick players-in-view list (Tb[]).
     * If the player is already in the known-players list the waypoint ring is
     * advanced; otherwise all state is freshly initialised.
     * obf: ta a(int animNext, int npcType, int tileX, byte sentinel, int tileZ, int serverIdx)
     *      client.U(
     */
    private final GameCharacter addPlayer(int animNext, int npcType, int tileX,
                                           byte sentinel, int tileZ, int serverIdx) {
        // allocate from cache if needed
        if (te[serverIdx] == null) {
            te[serverIdx]   = new GameCharacter();
            te[serverIdx].b = serverIdx;   // .b = serverIndex
        }
        GameCharacter player = te[serverIdx];   // playersCache[serverIdx]

        // check whether this player was already in the previous tick's list
        boolean known = false;
        for (int i = 0; i < qj /* knownPlayerCount */; i++) {
            if (Ff[i].b == serverIdx) {
                known = true;
                break;
            }
        }

        // sentinel byte = 127 means "use default anim"; otherwise emit sound/event
        if (sentinel != 127) {
            this.a((byte)-81, -15, (String)null);  // internal event dispatch
        }

        if (known) {
            // already visible: update animation and advance waypoint ring
            player.D = animNext;   // .D = animationNext
            player.t = npcType;    // .t = npcId / type
            int wp = player.o;     // .o = waypointCurrent
            if (player.k[wp] != tileX || tileZ != player.F[wp]) {
                player.o = wp = (wp + 1) % 10;
                player.k[wp] = tileX;   // waypointsX
                player.F[wp] = tileZ;   // waypointsZ (F=Y in oracle)
            }
        } else {
            // first appearance: full initialise
            player.b  = serverIdx;
            player.o  = 0;
            player.e  = 0;          // movingStep
            player.k[0] = player.i = tileX;   // waypointsX[0] = currentX
            player.D  = player.y = animNext;   // animNext = animCurrent
            player.x  = 0;          // stepCount
            player.t  = npcType;
            player.F[0] = player.K = tileZ;    // waypointsZ[0] = currentZ
        }

        // append to current-tick list
        Tb[de++] = player;   // de = playerCount
        return player;
    }

    // -------------------------------------------------------------------------
    // addNpc
    // -------------------------------------------------------------------------

    /**
     * Create or update an NPC entity and add it to the current-tick
     * NPCs-in-view list (rg[]).
     * obf: ta d(int tileZ, int serverIdx, int tileX, int opaqueSentinel, int animNext)
     */
    private final GameCharacter addNpc(int tileZ, int serverIdx, int tileX,
                                        int opaqueSentinel, int animNext) {
        // allocate from NPC cache if needed
        if (We[serverIdx] == null) {
            We[serverIdx]   = new GameCharacter();
            We[serverIdx].b = serverIdx;
        }
        GameCharacter npc = We[serverIdx];   // npcsCache[serverIdx]

        // check whether this NPC was already in the previous tick's Zg[] list
        boolean known = false;
        for (int i = 0; i < If /* currentNpcCount */; i++) {
            if (~serverIdx == ~Zg[i].b) {
                known = true;
                break;
            }
        }

        if (known) {
            // update animation and advance waypoint ring
            npc.D = animNext;
            int wp = npc.o;
            if (npc.k[wp] != tileX || ~tileZ != ~npc.F[wp]) {
                npc.o = wp = (wp + 1) % 10;
                npc.k[wp] = tileX;
                npc.F[wp] = tileZ;
            }
        } else {
            // first appearance: full initialise
            npc.b    = serverIdx;
            npc.k[0] = npc.i = tileX;
            npc.o    = 0;
            npc.e    = 0;
            npc.x    = 0;
            npc.D    = npc.y  = animNext;
            npc.F[0] = npc.K  = tileZ;
        }

        // opaqueSentinel is a junk anti-tamper expression; discard:
        //   n8 = -98; n7 = (0 - opaqueSentinel) / 39; n9 = n8 % n7; → dead

        rg[Yc++] = npc;   // Yc = npcCount
        return npc;
    }

    // -------------------------------------------------------------------------
    // getNpc
    // -------------------------------------------------------------------------

    /**
     * Look up an NPC GameCharacter in the current-tick NPC array by server index.
     * Returns null if not found; also clears localPlayer (wi) if sentinel==220.
     * obf: ta d(int serverIndex, int sentinel)   client.K(
     */
    private final GameCharacter getNpc(int serverIndex, int sentinel) {
        // rg[] = NPCs in view (Yc = count)
        for (int i = 0; i < Yc; i++) {
            if (serverIndex == rg[i].b) {
                return rg[i];
            }
        }
        // not found; sentinel 220 signals "clear local player reference"
        if (sentinel == 220) {
            wi = null;
        }
        return null;
    }

    // -------------------------------------------------------------------------
    // setCamera  (skeleton: void a(int,int,byte,int) client.D( — actually drawIcon)
    // -------------------------------------------------------------------------

    /**
     * Draw a compass/direction arrow icon sprite onto the minimap panel.
     * NOTE: The skeleton labels this "setCamera", but the actual body
     * performs 2D sprite blitting, not camera positioning.  The true
     * scene-camera call (World.setCamera / Ek.a(...,912,...)) is inlined
     * inside drawWorld.  Renamed drawIcon for accuracy.
     * obf: void a(int x, int y, byte size, int spriteBase)   client.D(
     */
    private final void drawIcon(int x, int y, byte size, int spriteBase) {
        // blit main compass rose sprite (sprite 76)
        surface.a(spriteBase, y,   76, x);
        // blit directional arrow overlay (sprite 111)
        surface.a(spriteBase, y - 1, 111, x);
        if (size <= -32) {
            // large icon variant: add extra border/arrow sprites
            surface.a(spriteBase, y + 1, 111, x);
            surface.a(spriteBase - 1, y, 60, x);
            surface.a(spriteBase + 1, y, 112, x);
        }
    }

    // -------------------------------------------------------------------------
    // updateCamera  (skeleton: void b(int,int,int) — actually sendDuelItems)
    // -------------------------------------------------------------------------

    /**
     * Send updated duel-offer item list to the server (opcode 33 DUEL_OFFER_ITEM).
     * NOTE: The skeleton labels this "updateCamera" at L:23985.  However the
     * method body sends a network packet (opc 33) and manages the Uf[]/df[]
     * inventory-slot arrays; it has nothing to do with camera positioning.
     * The true per-tick camera follow/auto-rotate is inlined in drawWorld (f(int)=13).
     * Skeleton mapping error flagged as uncertainty.
     * obf: void b(int, int, int)
     */
    private final void updateCamera(int param1, int param2, int itemSlot) {
        // itemId for the slot being offered
        int itemId = vf[itemSlot];   // vf[] = inventory item id array

        boolean changed = false;
        int matchCount  = 0;

        // scan existing offer slots for this item
        for (int i = 0; i < Ke; i++) {
            if (~itemId != ~Uf[i]) continue;        // Uf[i] = offered item id
            // found: increment quantity (stackable items, fa.e[id]==0 → stackable)
            if (EntityDef.isStackable[itemId] == 0) {
                if (param2 > 0) {
                    // increase count, clamp to max (xe[itemSlot])
                    df[i] += param2;
                    if (xe[itemSlot] < df[i]) df[i] = xe[itemSlot];
                }
                changed = true;
            } else {
                df[i] += param2;
                if (xe[itemSlot] < df[i]) df[i] = xe[itemSlot];
                changed = true;
            }
            matchCount++;
        }

        // if not found and adding (param2 > 0): add new offer slot (up to 8)
        if (!changed) {
            if (param2 > 0 && ~Ke > -9) {
                Uf[Ke] = itemId;
                df[Ke] = 1;
                Ke++;
                changed = true;
                // for non-stackable: clamp first quantity to min(xe, param2)
                if (EntityDef.isStackable[itemId] == 0) {
                    df[Ke - 1] = Math.min(xe[itemSlot], param2);
                }
            }
        }

        // auto-accept check: if this item's category is "always accept"
        if (matchCount >= Tk || EntityDef.alwaysAccept[itemId] == 1) {
            changed = true;
            this.b(null, (byte)-34);   // trigger auto-accept callback
        }

        if (!changed) {
            ke = false;   // no offer change
            return;
        }

        // send opcode 33 (DUEL_OFFER_ITEM): count + list of (itemId, qty)
        clientStream.b(33, 0);         // begin packet opcode 33
        clientStream.f.c(Ke, -120);    // write count (8-bit)
        for (int i = 0; i < Ke; i++) {
            clientStream.f.e(393, Uf[i]);           // write item id (16-bit)
            clientStream.f.b(-422797528, df[i]);    // write quantity (16-bit)
        }
        clientStream.b(21294);         // flush packet

        ki = false;   // clear "items changed" flag
        ke = false;
    }

    // -------------------------------------------------------------------------
    // resetChatInput
    // -------------------------------------------------------------------------

    /**
     * Clear the chat/text-entry input buffers (current line and scroll buffer).
     * obf: void o(byte)   client.NA(
     */
    private final void resetChatInput(byte param) {
        x  = "";   // current input text (x = inputTextCurrent)
        if (param != -49) {
            Nc = 13;   // reset cursor / line counter
        }
        Ob = "";   // second input buffer (Ob = inputTextFinal or scroll buffer)
    }

    // -------------------------------------------------------------------------
    // sortDrawList
    // -------------------------------------------------------------------------

    /**
     * Bubble-sort the friends/social list display order by online status and
     * key-state flags.  Swaps ac.z[], ua.h[], cb.c[], and Fj[] in tandem.
     * Despite the skeleton's name "sortDrawList", the sorted arrays are the
     * friends-list state arrays, not a 3D polygon draw list.
     * obf: void v(int)
     */
    private final void sortDrawList(int param) {
        if (param < 14) {
            this.a(-44, 54, 119, 125, true, 30);  // guard: init some scroll state
        }

        // bubble-sort pass: repeat until no swap
        boolean swapped = true;
        while (swapped) {
            swapped = false;
            for (int i = 0; i < n.g - 1 /* friendListSize - 1 */; i++) {
                // compare adjacent entries by keyState (held vs not) and pressed flag
                boolean leftHeld    = (Fj[i]   & 2) != 0;
                boolean rightHeld   = (Fj[i+1] & 2) != 0;
                boolean leftPressed = (Fj[i]   & 4) != 0;
                boolean rightPressed= (Fj[i+1] & 4) != 0;

                if ((!leftHeld && rightHeld) || (leftPressed && !rightPressed)) {
                    // swap display-name (ac.z)
                    String tmpAcZ = ac.z[i];
                    ac.z[i]   = ac.z[i + 1];
                    ac.z[i+1] = tmpAcZ;

                    // swap real-name (ua.h)
                    String tmpUaH = ua.h[i];
                    ua.h[i]   = ua.h[i + 1];
                    ua.h[i+1] = tmpUaH;

                    // swap clan/note field (cb.c)
                    String tmpCbC = cb.c[i];
                    cb.c[i]   = cb.c[i + 1];
                    cb.c[i+1] = tmpCbC;

                    // swap keyState bitfield
                    int tmpFj = Fj[i];
                    Fj[i]   = Fj[i + 1];
                    Fj[i+1] = tmpFj;

                    swapped = true;
                }
            }
        }
    }

    // -------------------------------------------------------------------------
    // findStringInData
    // -------------------------------------------------------------------------

    /**
     * Scan a name-table data blob for a string and return its associated record offset.
     * The blob is a flat array of 10-byte records:
     *   bytes [5..8] = 4-byte name hash (big-endian int)
     *   bytes [6..8] = 3-byte record offset (big-endian int, upper 24 bits)
     * Hash is computed as: hash = Σ (61 * hash + (char - 32)) for each char in uppercase name.
     * Returns 0 if not found.
     * obf: static int a(byte[] data, String name, int param)   client.ND(
     */
    private static final int findStringInData(byte[] data, String name, int param) {
        // param is an anti-tamper sentinel (if param > -18, return 113); stripped:
        // if (param > -18) return 113;  ← dead guard, removed

        int recordCount = DecodeBuffer.a(0, (byte)127, data); // read record count from blob
        name = name.toUpperCase();

        // compute name hash: 61-ary polynomial (same as NameHash.a)
        int hash = 0;
        for (int i = 0; i < name.length(); i++) {
            hash = 61 * hash + name.charAt(i) - 32;
        }

        // scan records (each 10 bytes wide)
        for (int i = 0; i < recordCount; i++) {
            // bytes [5..8]: name hash as 4-byte big-endian int
            int storedHash = (data[i * 10 + 5] & 0xFF)
                           + (data[i * 10 + 2] & 0xFF) * 0x1000000
                           + (data[i * 10 + 3] & 0xFF) * 65536
                           + (data[i * 10 + 4] & 0xFF) * 256;
            // bytes [6..8]: record offset as 3-byte big-endian int
            int offset = (data[i * 10 + 6] & 0xFF) * 65536
                       + (data[i * 10 + 7] & 0xFF) * 256
                       + (data[i * 10 + 8] & 0xFF);
            if (storedHash == hash) {
                return offset;
            }
        }
        return 0;
    }

    // -------------------------------------------------------------------------
    // readDefString
    // -------------------------------------------------------------------------

    /**
     * Read a length-prefixed definition string from a Buffer at the current cursor.
     * If the string is empty or an exception occurs, returns "Cabbage" (il[32]).
     * The length is read as a byte (tb.b(byte)), then the raw bytes are decoded
     * via CharTable.a() into a display string.
     * obf: static String a(int offset, tb buf, int maxLen)   client.CD(
     */
    static final String readDefString(int offset, Buffer buf, int maxLen) {
        try {
            int len = buf.b((byte)68);   // read 1-byte length prefix
            if (maxLen > len) len = maxLen;  // honour caller's minimum length
            byte[] raw = new byte[len];
            // copy from buffer at current cursor position, advancing cursor
            buf.w = buf.w + BZip.a.a(buf.F, raw, offset, buf.w, -1, len);
            // decode raw bytes → display string (accent/special-char handling)
            return CharTable.a(len, offset ^ -124, 0, raw);
        } catch (Exception ex) {
            return STRINGS[32];  // "Cabbage" — default/error sentinel string
        }
    }
