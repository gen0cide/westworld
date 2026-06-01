// ===== scene =====
// Group: Mudclient 3D scene/world build, entity menu-build & sprite render, camera, models.
// All methods are instance methods of class Mudclient (package client, extends GameShell),
// 4-space indented as if inside the class body.
//
// RE-AUDITED against the CLEAN Vineflower base (decompiled/normalized-clean/client.java).
// The previous version of this part file was written against a DEFECTIVE base (drawWorld /
// addSceneObject / buildTerrainTile / loadRegion were missing and reconstructed by hand);
// those reconstructions contained numerous logic errors that are corrected here.
//
// Stripping applied to every method:
//   - opaque predicate:  boolean bl = client.OPAQUE_FALSE;  (always false, dead) — removed
//   - profiling counters: ++<StaticCounter>;                 (dead) — removed
//   - exception wrapper:  catch(RuntimeException e){ throw ErrorHandler.a(e,"sig"); } — unwrapped
//   - anti-tamper guards: if(param != <magic>) <side-effect>; — kept verbatim where it touches
//                         a field (the JIT'd code really does execute the store), noted // guard
//   - junk before shifts: XOR masks on shift amounts        — removed
//   - ~x>~y / ~x==const sign idioms                          — rewritten to plain comparisons
//
// CROSS-CLASS NAMING — CORRECTED per docs/NAMING.md ("k=World, lb=Scene"):
//   The field types are:  private k Hh;   private lb Ek;   private ba li;
//   => Hh is type k  == World   (terrain/elevation/region/route)  -> read as `world`
//   => Ek is type lb == Scene   (3D renderer: model arrays, picking, camera) -> read as `scene`
//   The OLD version had these BACKWARDS (Hh=scene, Ek=world); fixed throughout this file.
//
//   Other key fields (MUDCLIENT_SKELETON.md):
//     li=surface(SurfaceSprite/ba), zh=menuList(wb; the right-click option accumulator —
//        NOT the friends list), Tb=playersInView, de=playerInViewCount,
//        rg=npcsInView, Yc=npcInViewCount, te=playersCache, We=npcsCache,
//        Ff/qj=knownPlayers/count, Zg/If=knownNpcs/count, Ug=tileSize,
//        af=selectedSpellOrItem, Bh=localPlayerServerIndex, ig=local player name.

    // -------------------------------------------------------------------------
    // drawWorld  (obf: void s(int)  @clean L10318)
    // -------------------------------------------------------------------------

    /**
     * Build the in-world right-click option menu for everything the 3D Scene
     * picked this frame: panel-element clicks first, then scenery objects,
     * boundary walls, ground items, NPCs and players, then a final
     * held-item / ground-tile fallthrough.
     *
     * Despite the skeleton label "drawWorld", this does NOT rasterise the scene
     * (that happens elsewhere); it walks the Scene's picked-entity lists
     * (Ek.b(124) = GameModel[], Ek.a(104) = encoded ids) and appends menu
     * options via the menu accumulator zh.  Each encoded id is
     *   E[idx] / 10000 = kind  (1=scenery, 2=ground item, 3=player), %10000 = local index;
     *   GameModels with rb >= 10000 are boundary walls (rb-10000 = wall index).
     *
     * obf: void s(int param)   (param: an interaction-pass selector; 2 = passive redraw)
     */
    private final void drawWorld(int param) {
        // --- panel-element click on the active sub-screen (Zh = screen/panel mode) ---
        if (Zh == 1 && yd.a((byte)-107, Fh) || Zh == 3 && yd.a((byte)-116, mc)) {
            int el = (Zh == 1) ? Fh : mc;                  // selected panel element
            int packed = yd.f(14458, el);
            if ((packed >> 16) == 2 || (Yh && (packed >> 16) == 1)) {
                int idx = packed & 0xFFFF;
                String actionA = yd.b(idx, 19680, el);
                String actionB = yd.a(idx, param ^ -122, el);
                if (this.a(actionA, param ^ 125, actionB)) {   // dispatch panel action
                    return;
                }
            }
        }

        // --- if on the login/world screen (Zh==0), hit-test the world list ---
        if (Zh == 0) {
            for (int w = 0; w < 100; w++) {
                if (pa.g[w] >= 0
                        && (n.j[w] == 4 || n.j[w] == 1 || n.j[w] == 5 || n.j[w] == 6)) {
                    String label = ub.a[w] + mb.a(aa.k[w], k.G[w], true, n.j[w]);
                    if (I > 7
                            && I < li.a(1, param ^ 114, label) + 7
                            && xb > Oi - 30 - w * 12
                            && xb < Oi - 18 - 12 * w
                            && (Cf == 2 || (Yh && Cf == 1))
                            && this.a(ba.Yb[w], 127, k.G[w])) {
                        return;
                    }
                }
            }
        } else {
            Hc = false;
        }

        // --- reset per-frame "already added to menu this tick" flags ---
        for (int i = 0; i < eh; i++) {
            Ed[i] = false;                                  // scenery objects
        }
        int clickedGroundSlot = -1;
        for (int i = 0; i < hf; i++) {
            Sj[i] = false;                                  // boundary walls
        }

        // --- iterate the Scene's picked entity list ---
        int entityCount = Ek.b(0);                          // Scene.pickedModelCount
        ca[] models = Ek.b((byte)124);                      // Scene.pickedModels (GameModel[])
        int[] faceTags = Ek.a((byte)104);                   // Scene.pickedFaceTags (int[])
        if (param != 2) {
            nk = -82;                                       // reset hover unless passive redraw
        }

        for (int pick = 0; pick < entityCount; pick++) {
            if (zh.c(param ^ -27155) > 200) {               // depth-cull: skip far-from-mouse polys
                continue;
            }
            int tag = faceTags[pick];
            ca model = models[pick];

            // valid face-tag bands: E[tag] in [0..0xFFFF], or [200000..300000]
            int eid = model.E[tag];
            if (!(eid <= 0xFFFF || (eid >= 200000 && eid <= 300000))) {
                continue;
            }

            // When the picked model is NOT the Scene's special target (Ek.T), it is a
            // scenery object (rb in [0..10000)) or a boundary wall (rb >= 10000); build
            // those menus and skip to the next entity.  When it IS the target, fall
            // through to the per-kind dispatch (ground item / player).
            if (Ek.T != model) {
                if (model == null || model.rb < 10000) {
                    // --- scenery object branch (also handles the ground-tile remap) ---
                    if (model != null && model.rb >= 0) {
                        int objSlot = model.rb;
                        int objId   = vc[objSlot];
                        if (!Ed[objSlot]) {
                            if (af < 0) {
                                if (Bh >= 0) {              // walk-to scenery
                                    zh.a(ye[objSlot], STRINGS[38] + ig + STRINGS[53], -104, Bh,
                                         vc[objSlot], 410, bg[objSlot],
                                         STRINGS[41] + l.a[objId], Se[objSlot]);
                                }
                                if (!s.f[objId].equalsIgnoreCase(STRINGS[33])) {  // command 1
                                    zh.a(420, ye[objSlot], bg[objSlot], Se[objSlot], param ^ 107,
                                         vc[objSlot], STRINGS[41] + l.a[objId], s.f[objId]);
                                }
                                if (!p.a[objId].equalsIgnoreCase(STRINGS[51])) {  // command 2
                                    zh.a(2400, ye[objSlot], bg[objSlot], Se[objSlot], param ^ 127,
                                         vc[objSlot], STRINGS[41] + l.a[objId], p.a[objId]);
                                }
                                zh.a(objId, 3400, false, STRINGS[51], STRINGS[41] + l.a[objId]); // Examine
                            } else if (qb.e[af] == 5) {     // cast held spell on scenery
                                zh.a(ye[objSlot], STRINGS[46] + ja.L[af] + STRINGS[50], param + 65,
                                     af, vc[objSlot], 400, bg[objSlot],
                                     STRINGS[41] + l.a[objId], Se[objSlot]);
                            }
                            Ed[objSlot] = true;
                        }
                        continue;                            // scenery handled
                    }
                    // model present but rb < 0: this is the ground-tile face — remember it
                    if (tag >= 0) {
                        tag = model.E[tag] - 200000;
                    }
                    if (tag < 0) {
                        continue;
                    }
                    clickedGroundSlot = tag;
                    continue;
                }

                // --- boundary wall branch (rb >= 10000) ---
                int wallIdx = model.rb - 10000;
                int wallId  = Ng[wallIdx];
                if (!Sj[wallIdx]) {
                    if (af >= 0 && qb.e[af] == 5) {          // cast held spell on wall
                        zh.a(300, yk[wallIdx], Hj[wallIdx], Jd[wallIdx], 60, af,
                             STRINGS[41] + ta.r[wallId], STRINGS[46] + ja.L[af] + STRINGS[50]);
                    }
                    if (Bh >= 0) {                           // walk-to wall
                        zh.a(310, yk[wallIdx], Hj[wallIdx], Jd[wallIdx], param ^ 66, Bh,
                             STRINGS[41] + ta.r[wallId], STRINGS[38] + ig + STRINGS[53]);
                    }
                    if (!u.b[wallId].equalsIgnoreCase(STRINGS[33])) {   // command 1
                        zh.a(Jd[wallIdx], (byte)22, 320, u.b[wallId],
                             STRINGS[41] + ta.r[wallId], Hj[wallIdx], yk[wallIdx]);
                    }
                    if (!f.e[wallId].equalsIgnoreCase(STRINGS[51])) {   // command 2
                        zh.a(Jd[wallIdx], (byte)22, 2300, f.e[wallId],
                             STRINGS[41] + ta.r[wallId], Hj[wallIdx], yk[wallIdx]);
                    }
                    zh.a(wallId, 3300, false, STRINGS[51], STRINGS[41] + ta.r[wallId]); // Examine
                    Sj[wallIdx] = true;
                }
                continue;                                    // wall handled
            }

            // --- Scene-target model: decode kind/local index and build mob menus ---
            int local = model.E[tag] % 10000;
            int kind  = model.E[tag] / 10000;
            if (kind != 1) {
                if (kind == 2) {
                    // ground item
                    if (af >= 0) {
                        if (qb.e[af] != 3) {                 // not a use-on-item spell/item
                            continue;
                        }
                        zh.a(200, Ni[local], Gj[local], Zf[local], param ^ 70, af,
                             STRINGS[34] + ac.x[Gj[local]], STRINGS[46] + ja.L[af] + STRINGS[50]);
                        continue;
                    }
                    if (Bh < 0) {                            // clean: ~Bh > -1  (Bh < 0)
                        zh.a(Zf[local], (byte)22, 220, STRINGS[52], STRINGS[34] + ac.x[Gj[local]],
                             Gj[local], Ni[local]);          // Pick up
                        zh.a(Gj[local], 3200, false, STRINGS[51], STRINGS[34] + ac.x[Gj[local]]); // Examine
                        continue;
                    }
                    zh.a(210, Ni[local], Gj[local], Zf[local], 68, Bh,
                         STRINGS[34] + ac.x[Gj[local]], STRINGS[38] + ig + STRINGS[53]);
                    continue;
                }
                if (kind != 3) {
                    continue;
                }

                // player
                String vsText = "";
                int combatDelta = -1;
                int charType = Tb[local].t;
                if (o.a[charType] > 0) {                     // combat-capable (PvP target)
                    int theirLevel = (eb.b[charType] + la.a[charType]
                                      + jb.k[charType] + fb.d[charType]) / 4;
                    int myLevel = (cg[3] + cg[2] + cg[1] + cg[0] + 27) / 4;
                    vsText = STRINGS[20];                    // baseline " @whi@" colour
                    combatDelta = myLevel - theirLevel;
                    if (combatDelta < 0)  vsText = STRINGS[40];   // they higher
                    if (combatDelta < -3) vsText = STRINGS[39];
                    if (combatDelta < -6) vsText = STRINGS[49];
                    if (combatDelta < -9) vsText = STRINGS[10];
                    if (combatDelta > 0)  vsText = STRINGS[35];   // you higher
                    if (combatDelta > 3)  vsText = STRINGS[37];
                    if (combatDelta > 6)  vsText = STRINGS[47];
                    if (combatDelta > 9)  vsText = STRINGS[27];
                    vsText = " " + vsText + STRINGS[42] + theirLevel + ")";
                }
                if (af >= 0) {
                    if (qb.e[af] != 2) {                     // not a cast-on-player spell
                        continue;
                    }
                    zh.a(Tb[local].b, STRINGS[20] + e.Mb[Tb[local].t], 700,
                         STRINGS[46] + ja.L[af] + STRINGS[50], af, 3296);
                    continue;
                }
                if (Bh < 0) {                                // clean: -1 < ~Bh  (Bh < 0)
                    if (o.a[charType] > 0) {                 // Attack
                        zh.a(Tb[local].b, combatDelta >= 0 ? 715 : 2715, false, STRINGS[48],
                             STRINGS[20] + e.Mb[Tb[local].t] + vsText);
                    }
                    zh.a(Tb[local].b, 720, false, STRINGS[45], STRINGS[20] + e.Mb[Tb[local].t]); // Trade
                    if (!p.e[charType].equals("")) {
                        zh.a(Tb[local].b, 725, false, p.e[charType], STRINGS[20] + e.Mb[Tb[local].t]);
                    }
                    zh.a(Tb[local].t, 3700, false, STRINGS[51], STRINGS[20] + e.Mb[Tb[local].t]); // Examine
                }
                zh.a(Tb[local].b, STRINGS[20] + e.Mb[Tb[local].t], 710,
                     STRINGS[38] + ig + STRINGS[53], Bh, param ^ 3298);  // Follow
            }

            this.a(local, -12);                              // walk-to entity tile
        }

        // --- global held-item fallback: "Use <item> with" when nothing else matched ---
        if (af >= 0 && qb.e[af] <= 1) {
            zh.a(af, 1000, false, STRINGS[46] + ja.L[af] + STRINGS[43], "");
        }

        // --- clicked ground tile (no entity): build a walk-here / use-on-ground option ---
        if (clickedGroundSlot != -1) {
            Hc = true;
            int slot = clickedGroundSlot;
            rf = Qg + Hh.q[slot];                            // world X = regionBaseX + scene tile X
            Cg = zg + Hh.E[slot];                            // world Z = regionBaseZ + scene tile Z
            if (af >= 0) {
                if (qb.e[af] != 6) {                         // not a "use-on-ground" spell
                    return;
                }
                zh.a(Hh.q[slot], (byte)22, 900, STRINGS[46] + ja.L[af] + STRINGS[44], "",
                     af, Hh.E[slot]);
                return;
            }
            if (Bh < 0) {                                    // clean: -1 < ~Bh  (Bh < 0)
                zh.a(Hh.q[slot], "", 920, STRINGS[54], Hh.E[slot], 3296);   // Walk here
            }
        }
    }

    // -------------------------------------------------------------------------
    // loadRegion  (obf: boolean a(int,int,boolean)  @clean L11537 — bytecode only)
    // -------------------------------------------------------------------------

    /**
     * (Re)load the 48x48 terrain region centred on the given world tile, snapping
     * the region origin to the chunk grid.  On a cache hit (player still inside the
     * loaded chunk and floor unchanged) returns false without reloading; otherwise
     * shows "Loading... Please wait", re-bases every wall/object/ground-item model
     * and every in-view player/npc by the region-origin delta, and returns true.
     *
     * Decoded directly from the method's bytecode (Vineflower could not decompile
     * this one method — the single residual failure in the clean base).
     *
     * obf: boolean a(int x, int z, boolean isUnderground)
     */
    private final boolean loadRegion(int x, int z, boolean isUnderground) {
        // disconnected / fatal stream error -> mark World not-ready, bail
        if (rk != 0) {
            Hh.Z = false;                                    // world.loaded = false
            return false;
        }
        this.Ub = isUnderground;

        // shift requested tile by the player's sub-region offset
        x += sk;
        z += Ki;

        // cache hit: same floor and still strictly inside the loaded window
        if (yj == bc && Jg < z && Rk > z && Fi < x && x < Ne) {
            Hh.Z = true;                                     // world.loaded = true
            return false;
        }

        // --- full region reload ---
        surface.a(256, STRINGS[676], 0xFFFFFF, 0, 1, 192);   // "Loading... Please wait"
        this.A(5);
        surface.a(graphics, Eb, 256, K);                     // flush to AWT

        int oldRegionX = Qg;
        int oldRegionZ = zg;

        // snap to 48-tile chunk grid; NOTE: Qg derives from z, zg derives from x
        int chunkZ = (z + 24) / 48;
        int chunkX = (x + 24) / 48;
        Qg = chunkZ * 48 - 48;
        Ne = chunkX * 48 + 32;
        Rk = chunkZ * 48 + 32;
        yj = bc;                                             // latch floor level
        Fi = chunkX * 48 - 32;
        zg = chunkX * 48 - 48;
        Jg = chunkZ * 48 - 32;

        // tell World its new origin/floor
        Hh.a(z, (byte)-90, x, yj);                           // world.setOrigin(z, x, floor)

        // subtract player sub-region offset so coords stay view-relative
        zg -= sk;
        Qg -= Ki;

        int deltaX = Qg - oldRegionX;                        // = -oldRegionX + Qg
        int deltaZ = zg - oldRegionZ;                        // = -oldRegionZ + zg

        // --- rebase wall/object models (slots [0..eh)) ---
        for (int i = 0; i < eh; i++) {
            Se[i] -= deltaX;
            ye[i] -= deltaZ;
            int tileX    = Se[i];
            int tileZ    = ye[i];
            int objType  = vc[i];
            ca model     = hg[i];
            int dir      = bg[i];

            try {
                int modelW, modelH;
                if (dir == 0 || dir == 4) {
                    modelW = f.f[objType];                   // obf: f.f  (RecordLoader.f = width table)
                    modelH = ub.g[objType];                  // obf: ub.g (NameTable.g  = height table)
                } else {
                    modelW = ub.g[objType];
                    modelH = f.f[objType];
                }
                int midX = (tileX + tileX + modelW) * Ug / 2;
                int midZ = Ug * (tileZ + tileZ + modelH) / 2;

                // cull if outside the visible 96x96 tile window
                if (tileX < 0 || tileZ < 0 || tileX >= 96 || tileZ >= 96) {
                    continue;
                }
                Ek.a(model, (byte)118);                       // scene.addModel
                model.c(-Hh.f(midX, midZ, 89), -123, midZ, midX); // translate to terrain height
                Hh.a(tileX, objType, isUnderground, tileZ);   // world.placeObject
                if (objType == 74) {                          // special: floats 480 up
                    model.a(0, 0, -480, true);
                }
            } catch (RuntimeException ex) {
                System.out.println(STRINGS[671] + ex.getMessage());
                System.out.println(STRINGS[672] + i + STRINGS[673] + model);
                ex.printStackTrace();
            }
        }

        // --- rebase door/diagonal entity models (slots [0..hf)) ---
        for (int i = 0; i < hf; i++) {
            Jd[i] -= deltaX;
            yk[i] -= deltaZ;
            int tileX   = Jd[i];
            int tileZ   = yk[i];
            int objType = Ng[i];
            int dir     = Hj[i];
            try {
                Hh.a(tileZ, objType, dir, tileX, 11715);      // world.placeBoundary
                rd[i] = this.buildEntityModel(!isUnderground, tileZ, objType, tileX, dir, i);
            } catch (RuntimeException ex) {
                System.out.println(STRINGS[674] + ex.getMessage());
                ex.printStackTrace();
            }
        }

        // --- rebase ground-item spots (slots [0..Ah)) ---
        for (int i = 0; i < Ah; i++) {
            Zf[i] -= deltaX;
            Ni[i] -= deltaZ;
        }

        // --- rebase in-view NPCs (rg[0..Yc)) ---
        for (int i = 0; i < Yc; i++) {
            ta npc = rg[i];
            npc.i -= Ug * deltaX;
            npc.K -= deltaZ * Ug;
            for (int wp = 0; wp <= npc.o; wp++) {
                npc.k[wp] -= Ug * deltaX;
                npc.F[wp] -= deltaZ * Ug;
            }
        }

        // --- rebase in-view players (Tb[0..de)) ---
        for (int i = 0; i < de; i++) {
            ta pl = Tb[i];
            pl.K -= Ug * deltaZ;
            pl.i -= Ug * deltaX;
            for (int wp = 0; wp <= pl.o; wp++) {
                pl.k[wp] -= Ug * deltaX;
                pl.F[wp] -= deltaZ * Ug;
            }
        }

        Hh.Z = true;                                         // world.loaded = true
        return true;
    }

    // -------------------------------------------------------------------------
    // buildEntityModel  (obf: ca a(boolean,int,int,int,int,int)  @clean L6204)
    // -------------------------------------------------------------------------

    /**
     * Build a single-quad GameModel for a boundary/door entity from World def data.
     * The quad spans from (tileX,tileZ) to a far corner chosen by the direction code
     * and is dropped to terrain elevation.  Tagged with rb = slot+10000 so the Scene
     * can recover its wall slot during picking.
     *
     * obf: ca a(boolean isUnderground, int tileX, int objType, int tileZ, int dir, int slot)
     */
    private final ca buildEntityModel(boolean isUnderground, int tileX, int objType,
                                      int tileZ, int dir, int slot) {
        int nearX = tileZ;          // var7  (near corner, tileZ axis copy)
        int nearZ = tileX;          // var8  (near corner, tileX axis copy)
        int farX  = tileZ;          // var9
        int farZ  = tileX;          // var10
        int colour = v.a[objType];  // obf: v.a  (ChatCipher.a = packed colour table)
        int bitmap = Jk[objType];   // obf: Jk   (texture/bitmap id table)
        int height = ib.d[objType]; // obf: ib.d (StreamBase.d = wall height)

        ca model = new ca(4, 1);
        if (dir == 1) {
            farZ = tileX + 1;
        }
        if (dir == 0) {
            farX = tileZ + 1;
        }
        if (dir == 2) {
            farZ = tileX + 1;
            nearX = tileZ + 1;
        }
        nearX *= Ug;
        if (dir == 3) {
            farZ = tileX + 1;
            farX = tileZ + 1;
        }
        nearZ *= Ug;
        farX  *= Ug;
        farZ  *= Ug;

        int v0 = model.e(nearX, nearZ, -Hh.f(nearX, nearZ, -35), -126);            // near bottom
        int v1 = model.e(nearX, nearZ, -Hh.f(nearX, nearZ, -103) - height, -126);  // near top
        if (!isUnderground) {
            this.a(119, 67, 26, 106, false, -100);           // internal flag setter (kept)
        }
        int v2 = model.e(farX, farZ, -height - Hh.f(farX, farZ, -77), -112);       // far top
        int v3 = model.e(farX, farZ, -Hh.f(farX, farZ, 96), 117);                  // far bottom

        model.a(4, new int[]{v0, v1, v2, v3}, colour, bitmap, false);
        model.a(-50, 60, -10, -50, false, 24, -95);          // lighting defaults

        if (tileZ >= 0 && tileX >= 0 && tileZ < 96 && tileX < 96) {
            Ek.a(model, (byte)118);                           // scene.addModel
        }
        model.rb = slot + 10000;                              // wall slot offset
        return model;
    }

    // -------------------------------------------------------------------------
    // addSceneObject  (obf: void a(int,int,int,int,int,int,int,int)  @clean L16874)
    // -------------------------------------------------------------------------

    /**
     * Render one PLAYER (Tb[id]) at screen position: blits each equipment/appearance
     * sprite layer, then queues its chat-message bubble, action bubble, health bar
     * and damage splat for later overlay passes.  (Oracle equivalent: drawPlayer.)
     *
     * obf: void a(int x, int y, int guard, int frameW, int scale, int id, int objW, int h)
     *   args: var1=x, var2=y, var3=guard, var4=frameW, var5=scale, var6=id, var7=objW, var8=h
     *   (objW renamed from obf var7=w to avoid clashing with the WorldEntity class `w`.)
     */
    final void addSceneObject(int x, int y, int guard, int frameW,
                              int scale, int id, int objW, int h) {
        if (guard != 20) {                                   // guard (kept; side-effect call)
            /* anti-tamper: 116 % ((69-guard)/35) — value discarded, no field touched */
        }
        ta player = Tb[id];

        // walk-cycle step: animationCurrent + (cameraRotation+16)/32, low 3 bits
        int walkAnim = (player.y + (ug + 16) / 32) & 7;
        boolean flip = false;
        int step = walkAnim;
        if (step == 5) { flip = true; step = 3; }
        else if (step == 6) { flip = true; step = 2; }
        else if (step == 7) { flip = true; step = 1; }

        int frame = sf[player.x / ob.h[player.t] % 4] + step * 3;   // base walk frame
        if (player.y == 8) {                                  // attacking
            flip = false;
            step = 5;
            h -= scale * db.j[player.t] / 100;
            walkAnim = 2;
            frame = 3 * step + Pc[jk / (na.a[player.t] - 1) % 8];
        }
        if (player.y == 9) {                                  // being hit
            step = 5;
            walkAnim = 2;
            flip = true;
            h += db.j[player.t] * scale / 100;
            frame = Og[jk / na.a[player.t] % 8] + 3 * step;
        }

        for (int layer = 0; layer < 12; layer++) {
            int bodyPart = Tg[walkAnim][layer];               // appearance layer order
            int itemId   = qb.d[player.t][bodyPart];          // appearance sprite id (-1 = none)
            if (itemId >= 0) {
                int dx = 0, dy = 0, f2 = frame;
                if (flip && step >= 1 && step <= 3 && aa.c[itemId] == 1) {
                    f2 += 15;
                }
                if (step != 5 || nb.d[itemId] == 1) {         // has a combat/idle frame
                    int sprite = f2 + w.g[itemId];            // obf: w.g (animationNumber)
                    int sw = li.Eb[sprite];                   // sprite full width
                    int sh = li.qb[sprite];                   // sprite full height
                    int baseW = li.Eb[w.g[itemId]];
                    if (sw != 0 && sh != 0 && baseW != 0) {
                        dy = frameW * dy / sh;
                        dx = dx * objW / sw;
                        int drawW = objW * li.Eb[sprite] / baseW;
                        dx -= (drawW - objW) / 2;
                        // appearance colour channel select (db.l = animationCharacterColour):
                        //   1 -> hair, 2 -> top, 3 -> bottom; default -> raw value (skin path)
                        int colourA = db.l[itemId];
                        int colourB = 0;
                        if (colourA == 1) {                   // hair
                            colourB = v.e[player.t];
                            colourA = da.T[player.t];
                        } else if (colourA == 2) {            // top
                            colourA = m.g[player.t];
                            colourB = v.e[player.t];
                        } else if (colourA == 3) {            // bottom
                            colourB = v.e[player.t];
                            colourA = ua.Ab[player.t];
                        }
                        // else: colourA keeps db.l value, colourB stays 0 (clean fall-through)
                        // obf arg order: li.a(dy+x, colourA, colourB, flip, y, sprite, frameW, drawW, dx+h, 1)
                        li.a(dy + x, colourA, colourB, flip, y, sprite, frameW, drawW, dx + h, 1);
                    }
                }
            }
        }

        // queue chat message bubble
        if (player.I > 0) {                                   // messageTimeout
            nf[Ef] = li.a(1, 120, player.n) / 2;             // mid-point (clamped to 150)
            if (nf[Ef] > 150) nf[Ef] = 150;
            uf[Ef] = li.a(1, 102, player.n) / 300 * li.a(508305352, 1);  // line count * lineH
            tf[Ef] = objW / 2 + h;
            ee[Ef] = x;
            Kc[Ef++] = player.n;
        }

        // queue health bar + damage splat during/after combat
        if (player.y == 8 || player.y == 9 || player.d != 0) {
            if (player.d > 0) {                              // combatTimer
                int barX = h;
                if (player.y == 9)      barX += scale * 20 / 100;
                else if (player.y == 8) barX -= scale * 20 / 100;
                int barLen = player.B * 30 / player.G;       // healthCurrent/healthMax *30
                gd[Bc] = objW / 2 + barX;
                Pk[Bc] = x;
                bf[Bc++] = barLen;
            }
            if (player.d > 150) {
                int dmgX = h;
                if (player.y == 9)      dmgX += scale * 10 / 100;
                else if (player.y == 8) dmgX -= scale * 10 / 100;
                li.b(-1, tg + 12, x + frameW / 2 - 12, dmgX - (12 - objW / 2));   // splat sprite
                li.a(objW / 2 + dmgX - 1, "" + player.u, 0xFFFFFF, 0, 3, 5 + x + frameW / 2); // dmg num
            }
        }
    }

    // -------------------------------------------------------------------------
    // addWallModel  (obf: void b(int,int,int,int,int,int,int)  @clean L13250  client.CA()
    // -------------------------------------------------------------------------

    /**
     * Blit a wall/boundary glyph sprite to the 2D surface (rev ~235 draws some
     * boundary art as 2D sprites over the 3D pass).
     * obf: void b(int x, int a2, int z, int spriteType, int height, int guard, int screenY) — 7 params
     */
    final void addWallModel(int x, int a2, int z, int spriteType,
                            int height, int guard, int screenY) {
        if (guard > -109) {                                   // guard (kept; touches tj)
            tj = 50;
        }
        int spriteIndex = ua.Bb[spriteType] + sg;             // base sprite + global offset
        int glyphWidth  = h.c[spriteType];                    // obf: h.c (TextEncoder width table)
        // obf arg order: li.a(screenY, glyphWidth, 0, false, 0, spriteIndex, x, height, z, 1)
        li.a(screenY, glyphWidth, 0, false, 0, spriteIndex, x, height, z, 1);
    }

    // -------------------------------------------------------------------------
    // addGroundObject  (obf: void a(int,int,int,int,int,int,int)  @clean L6064)
    // -------------------------------------------------------------------------

    /**
     * Draw a ground decoration marker (filled circle/oval) to the 2D surface — used
     * for fire/torch floor glow.
     * obf: void a(int a1, int y, int z, int w, int decorType, int h, int guard)  — 7 params
     */
    final void addGroundObject(int a1, int y, int z, int w,
                               int decorType, int h, int guard) {
        if (guard != 2) {                                     // guard (kept; sets Dc)
            Dc = true;
        }
        int shape = Oc[decorType];                            // 0 = circle, 1 = oval
        int size  = oe[decorType];                            // marker scale
        if (shape == 0) {
            int colour = 255 + size * 1280;
            li.c(255 - 5 * size, -1057205208, 20 + size * 2, w / 2 + z, colour, y + h / 2);
        }
        if (shape == 1) {
            int colour = 0xFF0000 + 1280 * size;
            li.c(255 - 5 * size, -1057205208, size + 10, z + w / 2, colour, y + h / 2);
        }
    }

    // -------------------------------------------------------------------------
    // buildTerrainTile  (obf: void b(int,int,int,int,int,int,int,int)  @clean L3544)
    // -------------------------------------------------------------------------

    /**
     * Render one NPC (rg[id]) at screen position: blits each appearance sprite layer
     * (with per-part walk offsets), then queues its chat bubble, action bubble,
     * health bar, damage splat and PK skull for later overlay passes.
     * (Oracle equivalent: drawNpc.)
     *
     * obf: void b(int objW, int frameW, int guard, int scale, int h, int x, int screenW, int id)
     *   args: var1=objW, var2=frameW, var3=guard, var4=scale, var5=h, var6=x, var7=screenW, var8=id
     *   (objW renamed from obf var1=w to avoid clashing with the WorldEntity class `w`.)
     */
    final void buildTerrainTile(int objW, int frameW, int guard, int scale,
                                int h, int x, int screenW, int id) {
        if (guard != 20) {                                    // guard (kept; side-effect call)
            this.e((byte)-115);
        }
        ta npc = rg[id];
        if (npc.A == 255) {                                   // colourBottom 255 -> invisible
            return;
        }

        int walkAnim = (npc.y + (ug + 16) / 32) & 7;
        boolean flip = false;
        int step = walkAnim;
        if (step == 5) { flip = true; step = 3; }
        else if (step == 6) { flip = true; step = 2; }
        else if (step == 7) { flip = true; step = 1; }

        int frame = sf[npc.x / 6 % 4] + 3 * step;
        if (npc.y == 8) {                                     // attacking
            h -= scale * 5 / 100;
            step = 5;
            flip = false;
            walkAnim = 2;
            frame = Pc[jk / 5 % 8] + 3 * step;
        }
        if (npc.y == 9) {                                     // being hit
            walkAnim = 2;
            step = 5;
            h += 5 * scale / 100;
            flip = true;
            frame = Og[jk / 6 % 8] + step * 3;
        }

        for (int layer = 0; layer < 12; layer++) {
            int bodyPart = Tg[walkAnim][layer];
            int itemId   = npc.m[bodyPart] - 1;               // equipped item id (-1 = none)
            if (itemId >= 0) {
                int dx = 0, dy = 0, f2 = frame;
                if (flip && step >= 1 && step <= 3) {
                    if (aa.c[itemId] != 1) {                  // per-part hand/shield offsets
                        if (bodyPart == 4 && step == 1) {
                            f2 = 3 * step + sf[(npc.x / 6 + 2) % 4];
                            dy = -3; dx = -22;
                        } else if (bodyPart == 4 && step == 2) {
                            dx = 0; dy = -8;
                            f2 = sf[(npc.x / 6 + 2) % 4] + 3 * step;
                        } else if (bodyPart == 4 && step == 3) {
                            dy = -5;
                            f2 = step * 3 + sf[(2 + npc.x / 6) % 4];
                            dx = 26;
                        } else if (bodyPart == 3 && step == 1) {
                            f2 = 3 * step + sf[(2 + npc.x / 6) % 4];
                            dx = 22; dy = 3;
                        } else if (bodyPart == 3 && step == 2) {
                            dy = 8;
                            f2 = 3 * step + sf[(npc.x / 6 + 2) % 4];
                            dx = 0;
                        } else if (bodyPart == 3 && step == 3) {
                            dx = -26;
                            f2 = sf[(2 + npc.x / 6) % 4] + step * 3;
                            dy = 5;
                        }
                    } else {
                        f2 += 15;
                    }
                }
                if (step != 5 || nb.d[itemId] == 1) {
                    int sprite = w.g[itemId] + f2;
                    int sw = li.Eb[sprite];
                    int sh = li.qb[sprite];
                    int baseW = li.Eb[w.g[itemId]];
                    if (sw != 0 && sh != 0 && baseW != 0) {
                        dy = dy * screenW / sh;
                        dx = dx * frameW / sw;
                        int drawW = sw * frameW / baseW;
                        dx -= (drawW - frameW) / 2;
                        int colourA = db.l[itemId];                   // animationCharacterColour
                        if (colourA == 1)      colourA = Dg[npc.p];   // hair
                        else if (colourA == 2) colourA = ei[npc.q];   // top
                        else if (colourA == 3) colourA = ei[npc.A];   // bottom
                        int colourB = Wh[npc.H];                      // skin
                        // obf arg order: li.a(x+dy, colourA, colourB, flip, objW, sprite, screenW, drawW, dx+h, 1)
                        li.a(x + dy, colourA, colourB, flip, objW, sprite, screenW, drawW, dx + h, 1);
                    }
                }
            }
        }

        // chat message bubble
        if (npc.I > 0) {                                       // messageTimeout
            nf[Ef] = li.a(1, 97, npc.n) / 2;
            if (nf[Ef] > 150) nf[Ef] = 150;
            uf[Ef] = li.a(1, 72, npc.n) / 300 * li.a(guard + 508305332, 1);
            tf[Ef] = frameW / 2 + h;
            ee[Ef] = x;
            Kc[Ef++] = npc.n;
        }

        // action bubble (item above head)
        if (npc.E > 0) {                                       // bubbleTimeout
            je[jc] = h + frameW / 2;
            pe[jc] = x;
            jd[jc] = scale;
            ak[jc++] = npc.j;                                  // bubbleItem
        }

        // health bar + damage splat
        if (npc.y == 8 || npc.y == 9 || npc.d != 0) {
            if (npc.d > 0) {                                  // combatTimer
                int barX = h;
                if (npc.y == 8)      barX -= scale * 20 / 100;
                else if (npc.y == 9) barX += 20 * scale / 100;
                int barLen = 30 * npc.B / npc.G;
                gd[Bc] = frameW / 2 + barX;
                Pk[Bc] = x;
                bf[Bc++] = barLen;
            }
            if (npc.d > 150) {
                int dmgX = h;
                if (npc.y == 8)      dmgX -= 10 * scale / 100;
                else if (npc.y == 9) dmgX += 10 * scale / 100;
                li.b(-1, tg + 11, screenW / 2 + x - 12, frameW / 2 + dmgX - 12);
                li.a(frameW / 2 + dmgX - 1, "" + npc.u, 0xFFFFFF, 0, 3, screenW / 2 + x + 5);
            }
        }

        // PK skull (skullVisible == 1 and no action bubble)
        if (npc.J == 1 && npc.E == 0) {
            int skullX = objW + h + frameW / 2;
            if (npc.y == 8)      skullX -= scale * 20 / 100;
            if (npc.y == 9)      skullX += 20 * scale / 100;
            int skullW = scale * 16 / 100;
            int skullH = 16 * scale / 100;
            li.f(skullX - skullW / 2, x - scale * 10 / 100 - skullH / 2, skullH, skullW,
                 5924, tg + 13);
        }
    }

    // -------------------------------------------------------------------------
    // getPlayer  (obf: ta b(int,byte)  @clean L3832  client.AC()
    // -------------------------------------------------------------------------

    /**
     * Resolve an in-view player GameCharacter (Tb[]) by server index; null if absent.
     * Side-effect: sets Bf when the sentinel byte != -123.
     * obf: ta b(int serverIndex, byte sentinel)
     */
    private final ta getPlayer(int serverIndex, byte sentinel) {
        if (sentinel != -123) {
            Bf = -116;
        }
        for (int i = 0; i < de; i++) {                        // de = in-view player count
            if (serverIndex == Tb[i].b) {
                return Tb[i];
            }
        }
        return null;
    }

    // -------------------------------------------------------------------------
    // addPlayer  (obf: ta a(int,int,int,byte,int,int)  @clean L13654  client.U()
    // -------------------------------------------------------------------------

    /**
     * Create or update a player entity and append it to the in-view list Tb[].
     * If the player was in the previous tick's known list (Ff[0..qj)) its animation
     * and waypoint ring are advanced; otherwise all state is freshly initialised.
     * obf: ta a(int animNext, int npcType, int tileX, byte sentinel, int tileZ, int serverIdx)
     */
    private final ta addPlayer(int animNext, int npcType, int tileX,
                               byte sentinel, int tileZ, int serverIdx) {
        if (te[serverIdx] == null) {
            te[serverIdx] = new ta();
            te[serverIdx].b = serverIdx;
        }
        ta player = te[serverIdx];                            // playersCache[serverIdx]

        boolean known = false;
        for (int i = 0; i < qj; i++) {                        // qj = known player count
            if (Ff[i].b == serverIdx) {
                known = true;
                break;
            }
        }

        if (sentinel != 127) {                                // emit event/sound unless default
            this.a((byte)-81, -15, (String)null);
        }

        if (known) {
            player.D = animNext;                              // animationNext
            player.t = npcType;
            int wp = player.o;                                // waypointCurrent
            if (player.k[wp] != tileX || tileZ != player.F[wp]) {
                player.o = wp = (wp + 1) % 10;
                player.k[wp] = tileX;
                player.F[wp] = tileZ;
            }
        } else {
            player.b = serverIdx;
            player.o = 0;
            player.e = 0;
            player.k[0] = player.i = tileX;
            player.D = player.y = animNext;
            player.x = 0;
            player.t = npcType;
            player.F[0] = player.K = tileZ;
        }

        Tb[de++] = player;                                    // append, de = in-view count
        return player;
    }

    // -------------------------------------------------------------------------
    // addNpc  (obf: ta d(int,int,int,int,int)  @clean L13871)
    // -------------------------------------------------------------------------

    /**
     * Create or update an NPC entity and append it to the in-view list rg[].
     * Known-check is against the previous tick's Zg[0..If).
     * obf: ta d(int tileZ, int serverIdx, int tileX, int junkGuard, int animNext)
     */
    private final ta addNpc(int tileZ, int serverIdx, int tileX,
                            int junkGuard, int animNext) {
        if (We[serverIdx] == null) {
            We[serverIdx] = new ta();
            We[serverIdx].b = serverIdx;
        }
        ta npc = We[serverIdx];                               // npcsCache[serverIdx]

        boolean known = false;
        for (int i = 0; i < If; i++) {                        // If = known npc count
            if (serverIdx == Zg[i].b) {
                known = true;
                break;
            }
        }

        if (known) {
            npc.D = animNext;
            int wp = npc.o;
            if (npc.k[wp] != tileX || tileZ != npc.F[wp]) {
                npc.o = wp = (wp + 1) % 10;
                npc.k[wp] = tileX;
                npc.F[wp] = tileZ;
            }
        } else {
            npc.b = serverIdx;
            npc.k[0] = npc.i = tileX;
            npc.o = 0;
            npc.e = 0;
            npc.x = 0;
            npc.D = npc.y = animNext;
            npc.F[0] = npc.K = tileZ;
        }
        // junkGuard: dead "-98 % ((0-junkGuard)/39)" expression — result discarded

        rg[Yc++] = npc;                                       // append, Yc = in-view count
        return npc;
    }

    // -------------------------------------------------------------------------
    // getNpc  (obf: ta d(int,int)  @clean L12247  client.K()
    // -------------------------------------------------------------------------

    /**
     * Resolve an in-view NPC GameCharacter (rg[]) by server index; null if absent.
     * Side-effect: clears the local-player reference (wi) when sentinel != 220.
     * obf: ta d(int serverIndex, int sentinel)
     */
    private final ta getNpc(int serverIndex, int sentinel) {
        for (int i = 0; i < Yc; i++) {                        // Yc = in-view npc count
            if (serverIndex == rg[i].b) {
                return rg[i];
            }
        }
        if (sentinel != 220) {                                // NOTE: != (clean: var2 != 220)
            wi = null;
        }
        return null;
    }

    // -------------------------------------------------------------------------
    // drawIcon  (obf: void a(int,int,byte,int)  @clean L9719  client.D()
    // -------------------------------------------------------------------------

    /**
     * Blit a compass/minimap icon: base sprite (76) + arrow overlay (111), and for
     * the large variant (size <= -32) three extra border sprites.
     * Skeleton labels this "setCamera" but the body is pure 2D sprite blitting; the
     * real scene-camera positioning is inlined elsewhere.
     * obf: void a(int x, int y, byte size, int spriteBase)
     */
    private final void drawIcon(int x, int y, byte size, int spriteBase) {
        li.a(spriteBase, y, 76, x);
        li.a(spriteBase, y - 1, 111, x);
        if (size <= -32) {
            li.a(spriteBase, y + 1, 111, x);
            li.a(spriteBase - 1, y, 60, x);
            li.a(spriteBase + 1, y, 112, x);
        }
    }

    // -------------------------------------------------------------------------
    // sendDuelItems  (obf: void b(int,int,int)  @clean L7479)
    // -------------------------------------------------------------------------

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
    private final void sendDuelItems(int p1, int delta, int invSlot) {
        boolean changed = false;
        int matched = 0;                                      // count of stackable duplicates seen
        int itemId = vf[invSlot];

        // pass over the existing offer slots looking for this item
        for (int i = 0; i < Ke; i++) {                        // Ke = current offer slot count
            if (itemId == Uf[i]) {
                if (fa.e[itemId] == 0) {                      // non-stackable
                    if (delta < 0) {                          // remove: tick df[] up to held count, Tk times
                        for (int n = 0; n < Tk; n++) {
                            if (df[i] < xe[invSlot]) {
                                df[i]++;
                            }
                            changed = true;
                        }
                    } else {                                  // add: bump df[] by delta, clamp to held
                        df[i] += delta;
                        if (xe[invSlot] < df[i]) {
                            df[i] = xe[invSlot];
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
            this.b((String)null, (byte)-34);                  // offer-confirmation callback
        }

        int slotsForItem = this.b(103, itemId);               // slots this item is allowed to occupy
        if (matched >= slotsForItem) {
            changed = true;
        }
        if (kb.c[itemId] == 1) {                               // item flagged non-offerable
            changed = true;
            this.a(false, null, 0, STRINGS[217], 0, 0, null, null);  // "cannot be added" message
        }

        // item not yet in the offer: add it
        if (!changed) {
            if (delta < 0) {
                // remove path with no existing slot: add a single df=1 slot (if room)
                if (Ke < 8) {
                    Uf[Ke] = itemId;
                    df[Ke] = 1;
                    Ke++;
                    changed = true;
                }
            } else {
                // add path: append df=1 slots while room remains and we are still under
                // the item's allowed slot count; the first slot of a non-stackable item
                // is clamped to min(heldCount, delta)
                for (int n = 0; delta > n; n++) {
                    if (Ke >= 8 || matched <= slotsForItem) {
                        break;
                    }
                    Uf[Ke] = itemId;
                    df[Ke] = 1;
                    matched++;
                    Ke++;
                    changed = true;
                    if (n == 0 && fa.e[itemId] == 0) {
                        df[Ke - 1] = Math.min(xe[invSlot], delta);
                        break;                                // clean: breaks after first clamp
                    }
                }
            }
        }

        if (!changed) {
            return;
        }

        // send opcode 33 (DUEL_OFFER_ITEM): count + (itemId, qty) pairs
        Jh.b(33, 0);
        Jh.f.c(Ke, -120);
        for (int j = 0; j < Ke; j++) {
            Jh.f.e(393, Uf[j]);
            Jh.f.b(-422797528, df[j]);
        }
        Jh.b(21294);
        ki = false;
        ke = false;
    }

    // -------------------------------------------------------------------------
    // resetChatInput  (obf: void o(byte)  @clean L6259  client.NA()
    // -------------------------------------------------------------------------

    /**
     * Clear the chat/text-entry buffers (x = current line, Ob = committed line).
     * obf: void o(byte sentinel)
     */
    private final void resetChatInput(byte sentinel) {
        x = "";
        if (sentinel != -49) {
            Nc = 13;
        }
        Ob = "";
    }

    // -------------------------------------------------------------------------
    // sortFriendsList  (obf: void v(int)  @clean L9656)
    // -------------------------------------------------------------------------

    /**
     * Bubble-sort the friends list so online (held) entries float up and recently
     * logged-in (pressed) entries follow; swaps display-name (ac.z), real-name
     * (ua.h), note (cb.c) and the Fj[] status bitfield in tandem.
     * Skeleton mislabels this "sortDrawList"; it sorts the social list, not a draw list.
     * obf: void v(int guard)
     */
    private final void sortFriendsList(int guard) {
        if (guard < 14) {                                     // guard (kept; scroll-state init)
            this.a(-44, 54, 119, 125, true, 30);
        }
        boolean swapped = true;
        while (swapped) {
            swapped = false;
            for (int i = 0; i < n.g - 1; i++) {               // n.g = friends list size
                boolean leftHeld     = (Fj[i]   & 2) != 0;
                boolean rightHeld    = (Fj[i+1] & 2) != 0;
                boolean leftPressed  = (Fj[i]   & 4) != 0;
                boolean rightPressed = (Fj[i+1] & 4) != 0;
                if ((!leftHeld && rightHeld) || (!leftPressed && rightPressed)) {
                    String tmp = ac.z[i];
                    ac.z[i]   = ac.z[i + 1];
                    ac.z[i+1] = tmp;
                    tmp = ua.h[i];
                    ua.h[i]   = ua.h[i + 1];
                    ua.h[i+1] = tmp;
                    tmp = cb.c[i];
                    cb.c[i]   = cb.c[i + 1];
                    cb.c[i+1] = tmp;
                    int tmpFj = Fj[i];
                    Fj[i]   = Fj[i + 1];
                    Fj[i+1] = tmpFj;
                    swapped = true;
                }
            }
        }
    }

    // -------------------------------------------------------------------------
    // findStringInData  (obf: static int a(byte[],String,int)  @clean L3450  client.ND()
    // -------------------------------------------------------------------------

    /**
     * Scan a name-table blob (count-prefixed array of 10-byte records) for a name and
     * return its 3-byte record offset; 0 if not found.  Record layout:
     *   [2..5] big-endian 4-byte name hash, [6..8] big-endian 3-byte offset.
     * Hash: 61-ary polynomial of (char - 32) over the uppercased name.
     *
     * obf: static int a(byte[] data, String name, int guard)
     *   guard: anti-tamper — when > -18 the real code returns 113 (kept, callers pass < -18).
     */
    private static final int findStringInData(byte[] data, String name, int guard) {
        int recordCount = d.a(0, (byte)127, data);
        name = name.toUpperCase();
        if (guard > -18) {                                    // guard (kept verbatim)
            return 113;
        }

        int hash = 0;
        for (int i = 0; i < name.length(); i++) {
            hash = 61 * hash + name.charAt(i) - 32;
        }

        for (int i = 0; i < recordCount; i++) {
            int storedHash = (data[i * 10 + 5] & 0xFF)
                           + (data[i * 10 + 2] & 0xFF) * 0x1000000
                           + (data[i * 10 + 3] & 0xFF) * 0x10000
                           + (data[i * 10 + 4] & 0xFF) * 0x100;
            int offset = (data[i * 10 + 6] & 0xFF) * 0x10000
                       + (data[i * 10 + 7] & 0xFF) * 0x100
                       + (data[i * 10 + 8] & 0xFF);
            if (storedHash == hash) {
                return offset;
            }
        }
        return 0;
    }

    // -------------------------------------------------------------------------
    // readDefString  (obf: static String a(int,tb,int)  @clean L5386  client.CD()
    // -------------------------------------------------------------------------

    /**
     * Read a length-prefixed definition string from a Buffer at the given offset,
     * honouring a caller minimum length; decode the raw bytes via CharTable.a().
     * Returns "Cabbage" (STRINGS[32]) on any error.
     * obf: static String a(int offset, tb buf, int minLen)
     */
    static final String readDefString(int offset, tb buf, int minLen) {
        try {
            int len = buf.b((byte)68);
            if (minLen > len) {
                len = minLen;
            }
            byte[] raw = new byte[len];
            buf.w = buf.w + fb.a.a(buf.F, raw, offset, buf.w, -1, len);
            return ga.a(len, offset ^ -124, 0, raw);
        } catch (Exception ex) {
            return STRINGS[32];                               // "Cabbage" — error sentinel
        }
    }
