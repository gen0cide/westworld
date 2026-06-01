// ===== util =====
// Methods in the "util" group of Mudclient (obfuscated class "client").
// 4-space indented as if inside the Mudclient class body.
// Obfuscation artefacts stripped: opaque predicate (boolean bl = client.vh),
// dead if(bl)/while(!bl) branches, ++<counter> profiling increments,
// try{BODY}catch(RuntimeException e){throw ErrorHandler.a(e,"sig")} wrappers,
// anti-tamper guards and junk dead divisions.
// All field / class names from MUDCLIENT_SKELETON.md + NAMING.md.

    /**
     * Look up the display name for the given entity/item id from the ImageLoader cache;
     * falls back to SurfaceSprite.e() if not yet cached.
     * (Skeleton proposed name: formatNumber — actual implementation is a name cache lookup,
     *  not numeric formatting; Utility.formatNumber lives in mb.a(int,int).)
     * // obf: String c(int,int)  label: client.GC(
     */
    private final String formatNumber(int unusedGuard, int entityId) {
        // Look up ListNode for entityId in the ImageLoader static hash table
        ListNode node = ImageLoader.k.a(entityId, (byte)-121);

        // Spin-wait while the async loader hasn't populated the node yet
        // (node.b == 0  ↔  ~node.b == -1: slot is loading)
        while (~node.b == -1) {
            Utility.a(11200, 50L); // sleep 50 ms while resource loads
        }

        // node.b == 1  ↔  ~node.b == -2: slot is populated and payload is set
        if (~node.b == -2 && node.d != null) {
            return (String) node.d;
        }

        // Fallback: ask the surface/sprite subsystem for the name
        return surface.e(114, entityId);
    }

    /**
     * Advance per-tick scroll/visibility state for the friends-list MessageList panel.
     * If Cf (duel/trade state flag) is non-zero, commits a scroll position and clears
     * the flag; otherwise re-scrolls the chat to the current viewport position.
     * // obf: void i(byte)  no label
     */
    private final void updateTimers(byte param1) {
        if (this.Cf != 0) {
            // Cf == 0: commit the pending scroll offset to friendsList
            int scrollResult = this.friendsList.b(this.I, this.rh, this.fg, (byte)-40, this.xb);
            if (~scrollResult >= -1) {               // scrollResult >= 0: valid scroll
                this.drawGameSettings(false, scrollResult);
            }
            this.se = false;
            this.Cf = 0;
            return;
        }

        // param1 == -106: set tj = -11 (timer reset sentinel) before re-scroll
        if (param1 == -106) {
            this.tj = -11;
        }

        // Read current scroll extents from friendsList
        int scrollMax    = this.friendsList.b(16256);   // scroll upper bound
        int scrollOffset = this.friendsList.a(-21224);  // current offset

        // Check whether the current chat viewport is fully within bounds
        if (~this.I <= ~(-10 + this.rh)
                && this.fg - 10 <= this.xb
                && ~this.I >= ~(scrollMax + (this.rh + 10))
                && ~(10 + this.fg - -scrollOffset) <= ~this.xb) {
            // Viewport is valid: commit the scroll
            this.friendsList.a(this.fg, this.rh, this.xb, (byte)-12, this.I);
            return;
        }

        // Viewport is out of range: hide the friends panel
        this.se = false;
    }

    /**
     * Clear transient per-session game state on (re)entry to the game world:
     * resets entity-count fields, nulls entity caches, clears per-tick flags,
     * and resets the name-resolution tables for all classes that hold 100-slot caches.
     * // obf: void i(int)  no label
     */
    private final void resetGameState(int param1) {
        // Reset connection/state machine counters
        this.kc = 0;   // login/state stage
        this.Xd = 0;   // panel-open flag
        this.bj = 0;   // pending-logout countdown

        this.qg = 1;   // "game loaded" guard
        this.Fg = 0;   // fatigue-flash flag

        // Clear active wall/boundary model references from Scene
        this.resetChatInput((byte)-49);

        // Reinitialise the surface back-buffer
        this.surface.a(true);
        this.surface.a(this.graphics, this.Eb, 256, this.K);

        // Remove all active wall models from World
        for (int i = 0; i < this.eh; ++i) {
            this.world.a(this.wallModels[i], -1);
            // Also evict the wall's scene entry from Scene
            this.scene.a(this.vc[i], this.Se[i], this.ye[i], 4081);
        }

        // Remove all active NPC models from World
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

        // Null out previous-tick player array (500 entries, stored inverted)
        for (int i = 0; ~i > ~500; ++i) {  // i < 500
            this.playersLast[i] = null;
        }
        this.de = 0;   // "players last" count

        // Null out player server-index cache (5 000 entries)
        for (int i = 0; i < 5000; ++i) {
            this.playersCache[i] = null;
        }

        // Null out previous-tick NPC array (500 entries)
        for (int i = 0; 500 > i; ++i) {
            this.npcsLast[i] = null;
        }

        // Clear per-NPC sleeping flag array (50 entries)
        for (int i = 0; 50 > i; ++i) {
            this.bk[i] = false;
        }

        // Reset boolean flags and per-tick state
        this.uk  = false;   // sleeping flag
        this.Bb  = 0;       // bank-page index
        this.Qb  = 0;       // shop scroll
        this.Cf  = 0;       // trade/duel sub-state
        this.Qk  = false;   // quest-list open

        // Junk dead division: 58 / ((param1 - -46) / 51) — opaque predicate artifact, ignored

        this.Fe  = false;   // "first login" flag
        FontWidths.g = 0;   // glyph-width scratch counter
        this.Vf  = 0;       // fatigue bar value

        // Clear the 100-slot shared name-resolution caches used by several subsystems
        for (int i = 0; 100 > i; ++i) {
            BZip.k[i]             = null;   // BZip name table slot
            ImageLoader.g[i]      = 0;      // ImageLoader index slot
            World.G[i]            = null;   // World name slot
            BitBuffer.N[i]        = 0;      // BitBuffer index slot
            SurfaceSprite.Yb[i]   = null;   // SurfaceSprite name slot
            NameTable.a[i]        = null;   // NameTable slot
            FontWidths.j[i]       = 0;      // FontWidths width slot
        }

        // Re-apply shop/quest/inventory panel scroll positions
        this.panelShop.c((byte)-33, this.Fh);
        this.panelShop.c((byte)-33, this.ud);
        this.panelShop.c((byte)-76, this.mc);
    }

    /**
     * Draw the loading progress bar, including a 3D rotating-globe backdrop rendered
     * via World/Scene, the game logo sprite, and an orange-shadowed progress strip.
     * Renders up to 5 animation states depending on how far the glob has rotated
     * (state driven by this.tg and this.dg progress fields).
     * // obf: void y(int)  label: client.HC(
     */
    private final void drawProgressBar(int param1) {
        // Set up World camera at a fixed loading-screen vantage point
        byte   cameraLayer = 0;
        int    camX        = 48 * 50 + 23;  // tile 50, world-units
        int    camZ        = 48 * 50 + 23;
        this.scene.a(camX, (byte)-90, camZ, cameraLayer);  // position camera
        this.scene.a(this.objectModels, (byte)-113);        // attach model array

        // World camera parameters (fixed loading-screen position)
        int worldX = 9728, worldZ = 6400, worldY = 1100;
        this.world.Mb = 4100;   // render distance / fog far
        this.world.X  = 4100;   // render near clip
        this.world.P  = 1;      // render flags
        this.world.G  = 4000;   // second render distance

        // Position Scene camera and render the globe for the first pass
        this.world.a(worldX, worldZ, worldY * 2, 912, -12349, 888,
                     -this.scene.f(worldX, worldZ, 73), 0);
        this.world.c(-124); // render pass A

        // param1 >= -48: no local player during loading
        if (param1 >= -48) {
            this.localPlayer = null;
        }

        // Clear surface to off-white background (#F8F8F9)
        this.surface.b(0xF8F8F9);
        this.surface.b(0xF8F8F9);

        // Draw the top progress-bar frame and orange shadow (6 pixels of gradient border)
        this.surface.a(0, (byte)65, 0, 0, 6, 512);
        for (int shadow = 6; shadow >= 1; --shadow) {
            this.surface.a(8, shadow, shadow, 0, 0xFF7000 /* orange */, 512, 0);
        }
        // Draw the lower progress-bar border strip at y=194, height=20
        this.surface.a(0, (byte)-104, 0, 194, 20, 512);

        // Bottom-left shadow strip (descending from y=194)
        for (int shadow = 6; shadow >= 1; --shadow) {
            this.surface.a(8, shadow, 194 - shadow, 0, 0xFF7000, 512, 0);
        }

        // Blit the progress bar logo sprite and render second globe pass
        this.surface.b(-1, this.tg + 10, 15, 15);                       // white fill box
        this.surface.d(this.dg, 200, 123, 512, 0, 0);                   // logo sprite
        this.surface.a(false, this.dg);                                   // blit logo

        // Second World render pass (camera shifted to 9216, 9216)
        worldX = 9216; worldZ = 9216; worldY = 1100;
        this.world.Mb = 4100;
        this.world.P  = 1;
        this.world.G  = 4000;
        this.world.X  = 4100;
        this.world.a(worldX, worldZ, 2 * worldY, 912, -12349, 888,
                     -this.scene.f(worldX, worldZ, 117), 0);
        this.world.c(-114); // render pass B

        this.surface.b(0xF8F8F9);
        this.surface.b(0xF8F8F9);
        this.surface.a(0, (byte)59, 0, 0, 6, 512);

        // Third shadow + globe pass
        for (int shadow = 6; shadow >= 1; --shadow) {
            this.surface.a(8, shadow, shadow, 0, 0xFF7000, 512, 0);
        }
        this.surface.a(0, (byte)-128, 0, 194, 20, 512);

        for (int shadow = 6; shadow >= 1; --shadow) {
            this.surface.a(8, shadow, 194 - shadow, 0, 0xFF7000, 512, 0);
        }
        this.surface.b(-1, 10 + this.tg, 15, 15);
        this.surface.d(1 + this.dg, 200, 124, 512, 0, 0);
        this.surface.a(false, 1 + this.dg);

        // Fourth pass: wider view (camera at 11136, 10368)
        int camX4 = 11136, camZ4 = 10368, camY4 = 500, camPitch4 = 376;
        // Remove all 64 terrain-tile/roof models from Scene before re-rendering
        for (int t = 0; 64 > t; ++t) {
            this.world.a(this.scene.db[0][t], -1);
            this.world.a(this.scene.g[1][t],  -1);
            this.world.a(this.scene.db[1][t], -1);
            this.world.a(this.scene.g[2][t],  -1);
            this.world.a(this.scene.db[2][t], -1);
        }
        this.world.Mb = 4100;
        this.world.G  = 4000;
        this.world.P  = 1;
        this.world.X  = 4100;
        this.world.a(camX4, camZ4, camY4 * 2, 912, -12349, camPitch4,
                     -this.scene.f(camX4, camZ4, 115), 0);
        this.world.c(-111); // render pass C

        this.surface.b(0xF8F8F9);
        this.surface.b(0xF8F8F9);
        this.surface.a(0, (byte)84, 0, 0, 6, 512);

        for (int shadow = 6; shadow >= 1; --shadow) {
            this.surface.a(8, shadow, shadow, 0, 0xFF7000, 512, 0);
        }
        this.surface.a(0, (byte)-107, 0, 194, 20, 512);

        // Fifth pass: final strip at y=194 (no vertical offset)
        for (int shadow = 6; shadow >= 1; --shadow) {
            this.surface.a(8, shadow, 194, 0, 0xFF7000, 512, 0);
        }
        this.surface.b(-1, 10 + this.tg, 15, 15);
        this.surface.d(this.tg + 10, 200, 120, 512, 0, 0);
        this.surface.a(false, this.tg + 10);
    }

    /**
     * Draw a filled/bordered box on the surface.
     * Uses the SurfaceSprite.a(…) blit primitives keyed off two dimension lookup tables
     * (ub.g[] and f.f[]) selected by the colour/style parameter n6.
     * // obf: void b(int,int,int,int,int)  no label
     */
    private final void drawBox(int magicKey, int styleIndex, int x, int y, int style) {
        // Opaque-predicate guard: if (magicKey != 5126) call a screen setup helper
        if (magicKey != 5126) {
            this.drawScrollbar2(true, (byte)-25);  // opaque-predicate-gated setup
        }

        // Dimension lookup depends on whether style == 0 or style == 4
        int w, h;
        if (~style == -1 || ~style == -5) {  // style == 0 or style == 4
            w = NameTable.g[styleIndex];     // box width  from NameTable
            h = RecordLoader.f[styleIndex];  // box height from RecordLoader
        } else {
            w = NameTable.g[styleIndex];
            h = RecordLoader.f[styleIndex];
        }

        // Skip drawing if the utility state says this slot is state 2 or 3
        // (mb.a[] is a Utility-class static int array tracking slot render state)
        if (Utility.a[styleIndex] != 2 && Utility.a[styleIndex] != 3) {
            // Draw outer/shadow border first (flags: true, -59 blitter mode)
            this.walkTo(x, true, this.Lf, y, this.sh, w - 1 + x, true, y + h - 1, -59);
        }

        // Adjust box coordinates based on style direction code
        if (style == 0) { ++w; --x; }
        if (style == 2) { ++h; }
        if (style == 6) { --y; ++h; }
        if (style == 4) { ++w; }

        // Draw inner fill (flags: false, -14 blitter mode)
        this.walkTo(x, true, this.Lf, y, this.sh, w + (x - 1), false, h + y - 1, -14);
    }

    /**
     * Update the directional surface-buffer index (si) by cycling through the
     * available terrain walkability slots, checking collision flags in World.bb[][] .
     * The parameter controls how many extra direction offsets to attempt.
     * // obf: void q(byte)  no label  (skeleton proposed name: clearScreen)
     */
    private final void clearScreen(byte numExtraDirections) {
        // Fast paths for the two common aligned cases
        if ((this.si & 1) == 1) {
            if (this.b((byte)90, this.si)) return;  // direction 90 is clear
        }
        if ((this.si & 1) == 0 && this.b((byte)113, this.si)) {
            // Try shifting si by 1 within octet
            if (!this.b((byte)-127, (1 + this.si) & 7)) {
                if (!this.b((byte)22, (7 + this.si) & 7)) return;
                this.si = (7 + this.si) & 7;
                return;
            }
            this.si = (1 + this.si) & 7;
            return;
        }

        // Extended direction table: {1, -1, 2, -2, 3, -3, 4}
        int[] dirOffsets = new int[]{1, -1, 2, -2, 3, -3, 4};
        // Only run extended search when numExtraDirections > 7
        if (numExtraDirections <= 7) return;

        for (int d = 0; ~d > ~7; ++d) {  // d < 7
            int candidate = this.b((byte)51, (8 + this.si + dirOffsets[d]) & 7) ? 1 : 0;
            if (candidate != 0) {
                // Found a clear direction — step si toward it
                this.si = (8 + this.si - -dirOffsets[d]) & 7;
                return;
            }
        }

        // Check if current si bit-0 is clear before secondary tests
        if ((this.si & 1) != 0) return;
        if (!this.b((byte)91, this.si)) return;

        if (this.b((byte)29, (1 + this.si) & 7)) {
            this.si = (1 + this.si) & 7;
            return;
        }
        if (!this.b((byte)-125, (7 + this.si) & 7)) return;
        this.si = (7 + this.si) & 7;
    }

    /**
     * Tear down (clear) and re-init visible panel widgets.
     * Fills the letterbox regions around the game viewport with black using AWT Graphics,
     * so that any leftover pixels outside the logical 512×334 area are masked.
     * // obf: void p(byte)  no label
     */
    private final void resetPanels(byte param1) {
        // Compute the four letterbox strip widths/heights
        int viewW = this.Eb;                          // applet component X-offset (left strip width)
        int viewH = this.K;                           // applet component Y-offset (top strip height)
        int rightW = -this.Wd + (this.Rh - viewW);   // right strip width  (Rh = render width)
        int botH   = -viewH - this.Oi - 12 + this.Hf;// bottom strip height (Hf = render height)

        // Dead division — junk opaque predicate: -40 / ((6 - param1) / 38) — ignored

        // Early out: all strips are zero-width/height
        if (viewW <= 0 && rightW <= 0 && viewH <= 0 && botH <= 0) return;

        // Obtain AWT Graphics for the appropriate host container
        java.awt.Graphics g;
        java.awt.Component target;
        if (this.hj) {
            // hj: fullscreen / applet mode — use the applet component directly
            target = (da.gb != null) ? da.gb : this;
        } else {
            target = InputState.a;  // standalone frame: use GameFrame
        }
        try {
            g = target.getGraphics();
        } catch (Exception e) {
            return;
        }
        if (g == null) return;

        // Paint black strips
        g.setColor(java.awt.Color.black);
        if (viewW > 0) {
            g.fillRect(0, 0, viewW, this.Hf);               // left strip
        }
        if (viewH > 0) {
            g.fillRect(0, 0, this.Rh, viewH);               // top strip
        }
        if (rightW > 0) {
            g.fillRect(this.Rh - rightW, 0, rightW, this.Hf); // right strip
        }
        if (botH > 0) {
            g.fillRect(0, this.Hf - botH, this.Rh, botH);  // bottom strip
        }
    }

    /**
     * Run a Runnable on the GameShell deferred-event queue.
     * Delegates to ImageLoader.k.a(true, runnable, priority).
     * // obf: void a(int,Runnable)  label: client.S(
     */
    @Override
    final void runOnQueue(int priority, Runnable task) {
        ImageLoader.k.a(true, task, priority);
    }

    /**
     * Show or hide a panel widget by id, updating the Panel's scroll/visibility state.
     * Also handles the special case where no duel/shop/trade panel (Xd==2) is open.
     * // obf: void a(byte,int)  no label
     */
    @Override
    final void setPanelVisible(byte panelId, int scrollY) {
        // If the game is fully loaded (qg == 0) and a shop/quest panel is open (Xd == 0),
        // forward the scroll position to the quest panel
        if (this.qg == 0) {
            if (this.Xd == 0 && this.panelQuest != null) {
                this.panelQuest.a(-12, scrollY);       // Panel.a(int,int) = setScroll
            }
            // If duel panel is active (Xd == 2) and duel panel exists, scroll it too
            if (~this.Xd == -3 && this.panelDuel != null) {
                this.panelDuel.a(-12, scrollY);
            }
        }

        // Guard: if panelId <= 105, nothing further to do
        if (panelId <= 105) return;

        // Only proceed if qg == 1 (game-world view, not loading screen)
        if (~this.qg != -2) return;   // qg != 1

        if (this.Kg) {
            // Members server: show quest/option panel
            this.panelQuest.a(-12, scrollY);
            return;
        }

        // Non-members: only show the stat panel if no duel/bank/fatigue overlay is active
        if (~this.Bj == -1 && ~this.Vf == -1 && !this.Qk && this.gc == 0) {
            this.panelShop.a(-12, scrollY);   // panelShop used as stats panel in f2p
        }
    }

    /**
     * Blit a sprite (by draw-list index) to the surface at (x, y), with an optional
     * screen-mode flag that sets cl = 61.
     * // obf: void a(boolean,int,int,int)  no label
     */
    private final void drawSprite(boolean setScreenMode, int x, int y, int drawMode) {
        // drawMode == 0: draw sprite frame at (x, y-1) with blitter mode -8
        if (~drawMode == -1) {  // drawMode == 0
            this.walkTo(x, true, this.Lf, y - 1, this.sh, x, false, y, -8);
        }
        // drawMode == 1: draw sprite frame at (x, y) with blitter mode 118
        else if (~drawMode == -2) {  // drawMode == 1
            this.walkTo(x, true, this.Lf, y, this.sh, x, true, y, 118);
        }
        // drawMode == 2: draw sprite frame at (x-1, y) with blitter mode 126
        else {
            this.walkTo(x - 1, true, this.Lf, y, this.sh, x, false, y, 126);
        }

        // If setScreenMode == true, flag the screen for a full repaint
        if (setScreenMode) {
            this.cl = 61;
        }
    }

    /**
     * "Ready/loaded" guard check — always returns true; the dead-code-elimination
     * junk division is an opaque predicate artefact that prevents inlining.
     * // obf: boolean f(byte)  label: client.LC(
     */
    private final boolean isLoaded(byte param1) {
        // Junk dead division: 89 % ((param1 + 74) / 51) — opaque predicate, always executes
        // but the result is unused; the real effect is to always return true.
        int junk = 89 % ((param1 - -74) / 51);
        return true;
    }

    /**
     * XOR string-pool decoder stage 1: converts a String to a char[], XOR-ing
     * the sole character with '~' (0x7E) when the string is shorter than 2 chars.
     * Used as the outer wrapper in z(z("…")) double-decode of STRINGS[].
     * // obf: static char[] z(String)  no label
     */
    private static char[] xorDecode1(String encoded) {
        char[] chars = encoded.toCharArray();
        // Short strings use a single-byte XOR key of '~' (0x7E)
        if (chars.length < 2) {
            chars[0] = (char) (chars[0] ^ '~');
        }
        return chars;
    }

    /**
     * XOR string-pool decoder stage 2: XOR each character with a position-dependent
     * key from the 5-byte table {34, 7, 117, 116, 126} (index = pos % 5), then
     * intern the result.  Pair with xorDecode1 to decode STRINGS[] entries.
     * // obf: static String z(char[])  no label
     */
    private static String xorDecode2(char[] chars) {
        // Per-position XOR key table (mod-5 cycle)
        // index:  0    1    2    3    4
        // key:   34    7  117  116  126
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
