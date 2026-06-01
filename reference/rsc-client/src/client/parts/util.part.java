// ===== util =====
// Methods in the "util" group of Mudclient (obfuscated class "client").
// 4-space indented as if inside the Mudclient class body.
// Re-audited against the CLEAN base (decompiled/normalized-clean/client.java).
// Obfuscation artefacts stripped: opaque predicate (boolean bl = client.vh),
// dead if(bl)/while(!bl)/break-label branches, ++<counter> profiling increments,
// try{BODY}catch(RuntimeException e){throw ErrorHandler.a(e,"sig")} wrappers,
// anti-tamper guards and junk dead divisions/modulos.
// All field / class names from MUDCLIENT_SKELETON.md + NAMING.md (k=World, lb=Scene).

    /**
     * Look up the display name for the given entity/item id from the ImageLoader cache;
     * spin-waits while the async loader populates the node, then returns the cached
     * String payload, else falls back to SurfaceSprite.e().
     * (Skeleton proposed name: formatNumber — actual implementation is a name cache lookup,
     *  not numeric formatting; Utility.formatNumber lives in mb.a(int,int).)
     * // obf: String c(int,int)  label: client.GC(
     */
    private final String formatNumber(int flag, int entityId) {
        // obf: if (var1 >= -7) this.Si = 126;  — guard side-effect the DEFECTIVE base dropped.
        // When flag >= -7, prime the Si scratch/state field to 126 before the lookup.
        if (flag >= -7) {
            this.Si = 126;
        }

        // Look up ListNode for entityId in the ImageLoader static hash table
        ListNode node = ImageLoader.k.a(entityId, (byte)-121);

        // Spin-wait while the async loader hasn't populated the node yet
        // (node.b == 0  ↔  ~node.b == -1: slot is loading)
        while (true) {
            if (~node.b == -1) {
                Utility.a(11200, 50L); // sleep 50 ms while resource loads
                continue;
            }
            // node.b == 1  ↔  ~node.b == -2: slot is populated and payload is set
            if (~node.b == -2 && node.d != null) {
                return (String) node.d;
            }
            break;
        }

        // Fallback: ask the surface/sprite subsystem for the name
        return surface.e(114, entityId);
    }

    /**
     * Advance per-tick scroll/visibility state for the friends-list MessageList panel.
     *
     * CLEAN-BASE CORRECTION: the two branches were SWAPPED in the defective base.
     *   - Cf == 0  : re-validate the chat viewport against the list's scroll extents and
     *                either commit the scroll (zh.a) or hide the panel (se=false).
     *   - Cf != 0  : commit the pending scroll position (zh.b), draw settings if valid,
     *                then clear se and Cf.
     * // obf: void i(byte)  no label
     */
    private final void updateTimers(byte param1) {
        if (this.Cf == 0) {
            // param1 != -106: arm the timer-reset sentinel before re-scroll
            if (param1 != -106) {
                this.tj = -11;
            }

            // Read current scroll extents from friendsList
            int scrollMax    = this.friendsList.b(16256);   // scroll upper bound
            int scrollOffset = this.friendsList.a(-21224);  // current offset

            // Check whether the current chat viewport is fully within bounds.
            // (~a <= ~b ⇔ a >= b ; ~a >= ~b ⇔ a <= b)
            if (~this.I <= ~(-10 + this.rh)
                    && this.fg - 10 <= this.xb
                    && ~this.I >= ~(scrollMax + this.rh + 10)
                    && ~(10 + this.fg - -scrollOffset) <= ~this.xb) {
                // Viewport is valid: commit the scroll
                this.friendsList.a(this.fg, this.rh, this.xb, (byte)-12, this.I);
            } else {
                // Viewport is out of range: hide the friends panel
                this.se = false;
            }
        } else {
            // Cf != 0: commit the pending scroll offset to friendsList
            int scrollResult = this.friendsList.b(this.I, this.rh, this.fg, (byte)-40, this.xb);
            if (~scrollResult <= -1) {               // ~scrollResult <= -1 ⇔ scrollResult >= 0
                this.drawGameSettings(false, scrollResult);
            }
            this.se = false;
            this.Cf = 0;
        }
    }

    /**
     * Clear transient per-session game state on (re)entry to the game world:
     * resets entity-count fields, nulls entity caches, clears per-tick flags,
     * and resets the 100-slot shared name-resolution tables across several classes.
     * // obf: void i(int)  label: client.<i(int)>  (il[115])
     */
    private final void resetGameState(int param1) {
        // Reset connection/state machine counters
        this.kc = 0;   // login/state stage
        this.Xd = 0;   // panel-open flag
        this.bj = 0;   // pending-logout countdown

        this.qg = 1;   // "game loaded" guard
        this.Fg = 0;   // fatigue-flash flag

        // Clear chat input buffers
        this.resetChatInput((byte)-49);

        // Reinitialise the surface back-buffer
        this.surface.a(true);
        this.surface.a(this.graphics, this.Eb, 256, this.K);

        // Remove all active wall/boundary models from World + Scene
        for (int i = 0; i < this.eh; ++i) {
            this.world.a(this.wallModels[i], -1);
            this.scene.a(this.vc[i], this.Se[i], this.ye[i], 4081);
        }

        // Remove all active NPC/anim models from World + Scene
        for (int i = 0; i < this.hf; ++i) {
            this.world.a(this.npcModelCache[i], -1);
            this.scene.a(true, this.Hj[i], this.yk[i], this.Jd[i], this.Ng[i]);
        }

        // Zero entity-count fields
        this.Ah = 0;   // wall/boundary count
        this.eh = 0;   // active wall-model count
        this.hf = 0;   // active NPC-model count
        this.Yc = 0;   // NPC view count

        // Null out NPC server-index cache (4 000 entries)
        for (int i = 0; i < 4000; ++i) {
            this.npcsCache[i] = null;
        }

        // Null out previous-tick player array (500 entries; obf: ~i > -501 ⇔ i < 500)
        for (int i = 0; ~i > -501; ++i) {
            this.playersLast[i] = null;
        }
        this.de = 0;   // "players last" count

        // Null out player server-index cache (5 000 entries)
        for (int i = 0; i < 5000; ++i) {
            this.playersCache[i] = null;
        }

        // Null out previous-tick NPC array (500 entries)
        for (int i = 0; i < 500; ++i) {
            this.npcsLast[i] = null;
        }

        // Clear per-NPC sleeping/transient flag array (50 entries)
        for (int i = 0; i < 50; ++i) {
            this.bk[i] = false;
        }

        // Reset boolean flags and per-tick state
        this.uk = false;   // sleeping flag
        this.Bb = 0;       // bank-page index
        this.Qb = 0;       // shop scroll
        this.Cf = 0;       // trade/duel sub-state
        this.Qk = false;   // quest-list open

        // obf: var2 = 58 / ((var1 - -46) / 51) — junk dead division, result discarded.

        this.Fe = false;   // "first login" flag
        FontWidths.g = 0;  // glyph-width scratch counter
        this.Vf = 0;       // fatigue bar value

        // Clear the 100-slot shared name-resolution caches used by several subsystems
        for (int i = 0; i < 100; ++i) {
            BZip.k[i]           = null;   // BZip name table slot   (aa.k)
            ImageLoader.g[i]    = 0;      // ImageLoader index slot (pa.g)
            World.G[i]          = null;   // World name slot        (k.G)
            BitBuffer.N[i]      = 0;      // BitBuffer index slot   (ja.N)
            SurfaceSprite.Yb[i] = null;   // SurfaceSprite name slot(ba.Yb)
            NameTable.a[i]      = null;   // NameTable slot         (ub.a)
            FontWidths.j[i]     = 0;      // FontWidths width slot  (n.j)
        }

        // Re-apply shop/quest/inventory panel scroll positions (yd = panelShop)
        this.panelShop.c((byte)-33, this.Fh);
        this.panelShop.c((byte)-33, this.ud);
        this.panelShop.c((byte)-76, this.mc);
    }

    /**
     * Draw the loading progress bar: a 3D rotating-globe backdrop rendered via
     * World/Scene across five camera passes, the game logo sprite, an off-white
     * background (#F8F8F9) and an orange-shadowed (#FF7000) progress strip.
     * // obf: void y(int)  label: client.HC(  (il[0])
     */
    private final void drawProgressBar(int param1) {
        // ---- Pass 1: camera at world tile (50,50) ----
        byte cameraLayer = 0;
        byte tileX = 50, tileZ = 50;
        this.scene.a(48 * tileX + 23, (byte)-90, 48 * tileZ + 23, cameraLayer);
        this.scene.a(this.objectModels, (byte)-113);

        int worldX = 9728, worldZ = 6400, worldY = 1100;
        int pitch = 888;
        this.world.Mb = 4100;
        this.world.X = 4100;
        this.world.P = 1;
        this.world.G = 4000;
        this.world.a(worldX, worldZ, worldY * 2, 912, -12349, pitch,
                     -this.scene.f(worldX, worldZ, 73), 0);
        this.world.c(-124); // render pass A

        // param1 >= -48: drop the local player ref during loading
        if (param1 >= -48) {
            this.localPlayer = null;
        }

        // Off-white background, then top progress-bar frame + orange shadow gradient
        this.surface.b(0xF8F8F9);
        this.surface.b(0xF8F8F9);
        this.surface.a(0, (byte)65, 0, 0, 6, 512);
        for (int s = 6; s >= 1; --s) {        // obf: ~var9 <= -2 ⇔ var9 >= 1
            this.surface.a(8, s, s, 0, 0xFF7000, 512, 0);
        }
        this.surface.a(0, (byte)-104, 0, 194, 20, 512);
        for (int s = 6; s >= 1; --s) {
            this.surface.a(8, s, 194 - s, 0, 0xFF7000, 512, 0);
        }
        this.surface.b(-1, this.tg + 10, 15, 15);
        this.surface.d(this.dg, 200, 123, 512, 0, 0);
        this.surface.a(false, this.dg);

        // ---- Pass 2: camera at (9216,9216) ----
        worldX = 9216; worldZ = 9216; worldY = 1100; pitch = 888;
        this.world.Mb = 4100;
        this.world.P = 1;
        this.world.G = 4000;
        this.world.X = 4100;
        this.world.a(worldX, worldZ, 2 * worldY, 912, -12349, pitch,
                     -this.scene.f(worldX, worldZ, 117), 0);
        this.world.c(-114); // render pass B

        this.surface.b(0xF8F8F9);
        this.surface.b(0xF8F8F9);
        this.surface.a(0, (byte)59, 0, 0, 6, 512);
        for (int s = 6; s >= 1; --s) {
            this.surface.a(8, s, s, 0, 0xFF7000, 512, 0);
        }
        this.surface.a(0, (byte)-128, 0, 194, 20, 512);
        for (int s = 6; s >= 1; --s) {
            this.surface.a(8, s, 194 - s, 0, 0xFF7000, 512, 0);
        }
        this.surface.b(-1, 10 + this.tg, 15, 15);
        this.surface.d(1 + this.dg, 200, 124, 512, 0, 0);
        this.surface.a(false, 1 + this.dg);

        // ---- Pass 3: wider view, camera at (11136,10368), y=500, pitch=376 ----
        worldX = 11136; worldZ = 10368; worldY = 500; pitch = 376;
        // Evict all 64 terrain-tile / roof models from World before re-rendering
        for (int t = 0; t < 64; ++t) {
            this.world.a(this.scene.db[0][t], -1);
            this.world.a(this.scene.g[1][t],  -1);
            this.world.a(this.scene.db[1][t], -1);
            this.world.a(this.scene.g[2][t],  -1);
            this.world.a(this.scene.db[2][t], -1);
        }
        this.world.Mb = 4100;
        this.world.G = 4000;
        this.world.P = 1;
        this.world.X = 4100;
        this.world.a(worldX, worldZ, worldY * 2, 912, -12349, pitch,
                     -this.scene.f(worldX, worldZ, 115), 0);
        this.world.c(-111); // render pass C

        this.surface.b(0xF8F8F9);
        this.surface.b(0xF8F8F9);
        this.surface.a(0, (byte)84, 0, 0, 6, 512);
        for (int s = 6; s >= 1; --s) {
            this.surface.a(8, s, s, 0, 0xFF7000, 512, 0);
        }
        this.surface.a(0, (byte)-107, 0, 194, 20, 512);
        // Final strip at y=194 (no vertical offset)
        for (int s = 6; s >= 1; --s) {
            this.surface.a(8, s, 194, 0, 0xFF7000, 512, 0);
        }
        this.surface.b(-1, 10 + this.tg, 15, 15);
        this.surface.d(this.tg + 10, 200, 120, 512, 0, 0);
        this.surface.a(false, this.tg + 10);
    }

    /**
     * Walk/box helper keyed off two dimension lookup tables (NameTable.g[] = X-extent,
     * RecordLoader.f[] = Y-extent) selected by index `styleIndex`.  Calls walkToAction
     * twice (outer border, then inner fill) with style-direction offsets.
     *
     * NOTE: the skeleton proposed name `drawBox`, but the body actually dispatches to
     * walkToAction (a(int,boolean,…) @5983, which sends WALK opcodes 16/187) — so the
     * "draw" framing is suspect; the body is transcribed faithfully. See uncertainties.
     *
     * CLEAN-BASE CORRECTION: the two table assignments differ per branch (the defective
     * base made both branches identical):
     *   style==0 || style==4 :  dimX = RecordLoader.f[i] ; dimY = NameTable.g[i]
     *   otherwise            :  dimX = NameTable.g[i]    ; dimY = RecordLoader.f[i]
     * // obf: void b(int,int,int,int,int)  no label  (il[216])
     */
    final void drawBox(int magicKey, int styleIndex, int x, int y, int style) {
        // Opaque-predicate guard: if (magicKey != 5126) call a setup helper (dead path)
        if (magicKey != 5126) {
            this.drawScrollbar2(true, (byte)-25);
        }

        int dimX, dimY;  // var6 = X-extent, var7 = Y-extent
        if (~style == -1 || ~style == -5) {   // style == 0 or style == 4
            dimY = NameTable.g[styleIndex];   // obf: var7 = ub.g[var2]
            dimX = RecordLoader.f[styleIndex];// obf: var6 = f.f[var2]
        } else {
            dimX = NameTable.g[styleIndex];   // obf: var6 = ub.g[var2]
            dimY = RecordLoader.f[styleIndex];// obf: var7 = f.f[var2]
        }

        // Skip outer border when the slot render-state is 2 or 3 (Utility.a[] = mb.a[])
        if (Utility.a[styleIndex] != 2 && Utility.a[styleIndex] != 3) {
            // Outer/border pass (flags: walk=true, mode -59)
            this.walkToAction(x, true, this.Lf, y, this.sh,
                              -1 + dimX + x, true, -1 + (y + dimY), -59);
        }

        // Style-direction adjustments
        if (style == 0) { ++dimX; --x; }
        if (style == 2) { ++dimY; }
        if (style == 6) { --y; ++dimY; }
        if (style == 4) { ++dimX; }

        // Inner/fill pass (flags: walk=false, mode -14)
        this.walkToAction(x, true, this.Lf, y, this.sh,
                          dimX + x - 1, false, dimY + y - 1, -14);
    }

    /**
     * Resolve a clear walkable direction by cycling `si` through the available
     * terrain walkability slots (World collision via this.b(byte,int) probes).
     * `numExtraDirections > 7` enables the extended {±1,±2,±3,+4} search.
     *
     * CLEAN-BASE CORRECTION: the extended-search loop must NOT return when it finds a
     * clear direction (the defective base added an early `return`); after the loop the
     * code re-tests (si&1)==0 && b(91,si) and applies a secondary ±1 nudge, and the whole
     * extended block is nested inside `if (numExtraDirections > 7)`.
     * // obf: void q(byte)  no label  (skeleton proposed name: clearScreen) (il[389])
     */
    private final void clearScreen(byte numExtraDirections) {
        // Primary fast path: if (si&1)==1 and direction 90 is clear, done.
        if ((this.si & 1) == 1 && this.b((byte)90, this.si)) {
            return;
        }

        // Secondary: (si&1)==0 and direction 113 is clear → nudge si by ±1 within octet.
        if ((this.si & 1) == 0 && this.b((byte)113, this.si)) {
            if (!this.b((byte)-127, (1 + this.si) & 7)) {
                if (!this.b((byte)22, (7 + this.si) & 7)) {
                    return;
                }
                this.si = (7 + this.si) & 7;
                return;
            }
            this.si = (1 + this.si) & 7;
            return;
        }

        // Extended direction search (only when numExtraDirections > 7)
        int[] dirOffsets = new int[]{1, -1, 2, -2, 3, -3, 4};
        if (numExtraDirections <= 7) {
            return;
        }

        // Probe each offset; on a hit, set si toward it AND fall through (no early return).
        for (int d = 0; d < 7; ++d) {  // obf: -8 < ~var3 ⇔ var3 < 7
            if (this.b((byte)51, (8 + this.si + dirOffsets[d]) & 7)) {
                this.si = (this.si + dirOffsets[d] + 8) & 7;
                break;
            }
        }

        // Secondary nudge after the search: requires (si&1)==0 and direction 91 clear.
        if ((this.si & 1) == 0 && this.b((byte)91, this.si)) {
            if (this.b((byte)29, (1 + this.si) & 7)) {
                this.si = (1 + this.si) & 7;
                return;
            }
            if (this.b((byte)-125, (7 + this.si) & 7)) {
                this.si = (7 + this.si) & 7;
            }
        }
    }

    /**
     * Tear down the letterbox regions around the game viewport: paints the four
     * black strips outside the logical 512×334 area via AWT Graphics.
     * // obf: void p(byte)  no label  (il[124])
     */
    private final void resetPanels(byte param1) {
        int leftW  = this.Eb;                              // left strip width  (component X-offset)
        int topH   = this.K;                               // top strip height  (component Y-offset)
        int rightW = -this.Wd + this.Rh + -leftW;          // right strip width
        int botH   = -topH - this.Oi - 12 + this.Hf;       // bottom strip height

        // obf: var6 = -40 / ((6 - var1) / 38) — junk dead division, result discarded.

        // Proceed if any strip is positive (obf: var2>0 || -1>~var4 || 0<var3 || var5>0)
        if (leftW > 0 || rightW > 0 || topH > 0 || botH > 0) {
            // Resolve the AWT host container for getGraphics()
            java.awt.Component target;
            if (this.hj) {
                target = (da.gb != null) ? da.gb : this;   // applet/fullscreen mode
            } else {
                target = InputState.a;                      // standalone frame (kb.a)
            }

            try {
                java.awt.Graphics g = target.getGraphics();
                if (g == null) {
                    return;
                }

                g.setColor(java.awt.Color.black);
                if (leftW > 0) {
                    g.fillRect(0, 0, leftW, this.Hf);                 // left strip
                }
                if (topH > 0) {                                       // obf: -1 > ~var3 ⇔ var3 > 0
                    g.fillRect(0, 0, this.Rh, topH);                  // top strip
                }
                if (rightW > 0) {
                    g.fillRect(-rightW + this.Rh, 0, rightW, this.Hf);// right strip
                }
                if (botH > 0) {                                       // obf: ~var5 < -1 ⇔ var5 > 0
                    g.fillRect(0, -botH + this.Hf, this.Rh, botH);    // bottom strip
                }
            } catch (Exception e) {
                // swallow (defective surface / not yet realised) — matches base
            }
        }
    }

    /**
     * Run a Runnable on the GameShell deferred-event queue.
     * Delegates to ImageLoader.k.a(true, runnable, priority).
     * // obf: void a(int,Runnable)  label: client.S(  (il[223])
     */
    @Override
    final void runOnQueue(int priority, Runnable task) {
        ImageLoader.k.a(true, task, priority);
    }

    /**
     * Forward a scroll position to whichever sub-panel is currently open, gated on the
     * load state (qg) and the open-panel id (Xd) / members flag (Kg).
     *
     * CLEAN-BASE CORRECTION vs the defective base:
     *   - (Xd==0) scrolls `ge`  (NOT panelQuest as the defective base claimed).
     *   - (Xd==2) scrolls `yi`.
     *   - members branch scrolls `Af`; f2p branch scrolls `yd`.
     * NAMING: the English panel names for ge/yi/Af/yd are CONTRADICTED across the
     * skeleton (Af=panelQuest, ge=panelTrade) vs the actual construction code
     * (clean base: Af = new qa(li,100) ⇒ panelDuel; ge,yi = new qa(li,50) built in
     * p(int)) and the existing part files. Until that is resolved in a dedicated
     * panel-naming pass, obf field names are kept here to avoid asserting a wrong
     * mapping; the `panelShop`/`yd` binding (stats panel in f2p) is the only one
     * already consistent across parts. See uncertainties.
     * // obf: void a(byte,int)  no label  (il[186])
     */
    @Override
    final void setPanelVisible(byte panelId, int scrollY) {
        // qg == 0: game fully loaded
        if (this.qg == 0) {
            if (this.Xd == 0 && this.ge != null) {
                this.ge.a(-12, scrollY);           // obf: this.ge.a(-12,var2)  (Panel.setScroll)
            }
            // Xd == 2 (obf: ~Xd == -3)
            if (~this.Xd == -3 && this.yi != null) {
                this.yi.a(-12, scrollY);           // obf: this.yi.a(-12,var2)
            }
        }

        if (panelId <= 105) {
            return;
        }

        // qg == 1: game-world view (obf: ~this.qg == -2)
        if (~this.qg == -2) {
            if (this.Kg) {
                // Members server: scroll the members-only panel (Af)
                this.Af.a(-12, scrollY);
                return;
            }
            // Non-members: only scroll the stats panel (yd = panelShop) when no
            // duel/fatigue/quest overlay is active (Bj==0 && Vf==0 && !Qk && gc==0).
            if (~this.Bj == -1 && ~this.Vf == -1 && !this.Qk && this.gc == 0) {
                this.panelShop.a(-12, scrollY);    // obf: this.yd.a(-12,var2)
            }
        }
    }

    /**
     * Blit a UI sprite (by draw-list index) to the surface at (x, y) via walkToAction,
     * with an optional screen-mode flag that sets cl = 61.
     *
     * CLEAN-BASE CORRECTION: drawMode 1 and 2 were SWAPPED in the defective base.
     *   drawMode==0 → mode -8  at (x, y-1)..(x, y)
     *   drawMode==1 → mode 126 at (x-1, y)..(x, y)   [fall-through case]
     *   drawMode==2 → mode 118 at (x, y)..(x, y)
     * // obf: void a(boolean,int,int,int)  no label  (il[388])
     */
    private final void drawSprite(boolean setScreenMode, int x, int y, int drawMode) {
        if (~drawMode == -1) {                 // drawMode == 0
            this.walkToAction(x, true, this.Lf, y - 1, this.sh, x, false, y, -8);
        } else if (~drawMode != -2) {          // drawMode != 1  → handles drawMode == 2
            this.walkToAction(x, true, this.Lf, y, this.sh, x, true, y, 118);
        } else {                               // drawMode == 1  (fall-through)
            this.walkToAction(x - 1, true, this.Lf, y, this.sh, x, false, y, 126);
        }

        if (setScreenMode) {
            this.cl = 61;
        }
    }

    /**
     * "Ready/loaded" guard check — always returns true; the junk modulo is an
     * opaque-predicate artefact whose result is discarded.
     * // obf: boolean f(byte)  label: client.LC(  (il[226])
     */
    private final boolean isLoaded(byte param1) {
        // obf: int var2 = 89 % ((param1 - -74) / 51); — dead, result unused.
        return true;
    }

    /**
     * XOR string-pool decoder stage 1: converts a String to a char[], XOR-ing the
     * sole character with 0x7E ('~') when the string is shorter than 2 chars.
     * Outer wrapper in z(z("…")) double-decode of STRINGS[].
     * // obf: static char[] z(String)  no label
     */
    private static char[] xorDecode1(String encoded) {
        char[] chars = encoded.toCharArray();
        if (chars.length < 2) {
            chars[0] = (char) (chars[0] ^ 126);
        }
        return chars;
    }

    /**
     * XOR string-pool decoder stage 2: XOR each char with a position-dependent key
     * from the 5-byte table {34, 7, 117, 116, 126} (index = pos % 5), then intern.
     * Pairs with xorDecode1 to decode STRINGS[] entries.
     * // obf: static String z(char[])  no label
     */
    private static String xorDecode2(char[] chars) {
        for (int i = 0; i < chars.length; i++) {
            byte key;
            switch (i % 5) {
                case 0:  key =  34; break;
                case 1:  key =   7; break;
                case 2:  key = 117; break;
                case 3:  key = 116; break;
                default: key = 126; break;
            }
            chars[i] = (char) (chars[i] ^ key);
        }
        return new String(chars).intern();
    }
