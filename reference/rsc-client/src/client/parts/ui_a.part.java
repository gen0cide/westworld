// ===== ui_a =====
// Group: ui (first 18 methods, listed order)
// Class: Mudclient (obf: client), package client, extends GameShell (obf: e)
// Field names from MUDCLIENT_SKELETON.md; class names from NAMING.md.
//
// RE-AUDITED against decompiled/normalized-clean/client.java (real Vineflower source).
// The previous version of this file was written against the DEFECTIVE base where these
// method bodies were missing/reconstructed; several were materially WRONG. See // FIX notes.
//
// Obfuscation stripped on read:
//   - opaque predicate `boolean varN = vh;` (always false) + every `if(!varN) break label;`
//     (always taken) and `if(varN) ...` (dead) — flattened to straight-line control flow.
//   - `++<counter>;` profiling bumps removed.
//   - try/catch(RuntimeException){ throw i.a(e, il[N]+...) } wrappers unwrapped.
//   - anti-tamper `if (param != <const>) <junk>` guards + dummy-param checks + junk
//     modulo expressions (`-121 / ((var1-19)/42)` etc.) removed.
//   - `~a == ~b`  →  `a == b`;  `~a < ~b`  →  `a > b`;  `~a <= -1` → `a >= 0`  (idiom unmasked).
//
// Class-token map applied per NAMING.md (corrects fabricated names from the old pass):
//   li=surface(ua=Surface), Jh=clientStream(da), mg=incomingPacket(ja), Hh=scene(lb=Scene),
//   Ek=world(k=World), wi=localPlayer(ta), zh=friendsList/He=chatList/Wf=ignoreList(wb),
//   ac=DecodeBuffer  (ac.x[] = item display names in this build — NOT "EntityDef"),
//   fa=ClientIOException (fa.e[] = item stackable flag),
//   kb=InputState (kb.b[]/kb.c[] item tables; kb.a = applet host),
//   h=TextEncoder (h.c[] = item inventory sprite ids),  ga=CharTable (ga.b[] = item examine),
//   ua=Surface (ua.Bb[] sprite base, ua.h[] friend names),  o=ISAAC (o.a(a,?,r,g)=ARGB pack),
//   mb=Utility, s=FontBuilder, d=CacheFile, f=RecordLoader, nb=DataStore, pa=ImageLoader,
//   l=Globals (l.c[] = chat-history names), ia=SpriteScaler (ia.a[] ignore, ia.g[] msg),
//   cb=CacheUpdater (cb.c[] friend display names).

    // -------------------------------------------------------------------------
    // drawActiveInterface  — obf: void I(int)
    // -------------------------------------------------------------------------

    /** Dispatch to whichever modal panel / overlay is currently open, in priority order;
     *  if none is open, render the in-world frame (HUD tabs, minimap, world/inventory tab,
     *  player context menu).  Param is the anti-tamper sentinel `bj`.
     *
     *  FIX vs old: the old version stopped after the panel dispatch and stuffed the entire
     *  in-world section into a comment, and mis-routed several panels
     *  (h(127)=drawDuelConfirm not drawSocialDialog; d(false)=drawReportNameEntry;
     *  M=drawShop and N=drawTradeConfirmWindow placement). Rewritten from the clean source. */
    private final void drawActiveInterface(int param) {
        boolean inWorld = false;

        // ---- modal-panel dispatch (each branch draws then returns) ----
        if (param != this.bj) {           // anti-tamper sentinel mismatch
            this.clearScreen((byte) 120);
        } else if (this.Oh) {
            this.drawWelcome(param - 4853);
        } else if (this.mh) {
            this.drawChat((byte) -115);
        } else if (this.le == 1) {
            this.drawWildernessWarning(120);
        } else if (this.Fe && this.ai == 0) {
            this.drawBank(-122);
        } else if (this.uk && this.ai == 0) {
            this.drawShop(-89);
        } else if (this.Xj) {
            this.drawTradeConfirmWindow(-54);
        } else if (this.Hk) {
            this.drawTrade((byte) 8);
        } else if (this.dd) {
            this.drawDuelConfirm(-33);
        } else if (this.Pj) {
            this.drawDuel(param ^ 40);
        } else if (this.Vf == 1) {        // ~Vf != -2  ⟺  Vf != 1 → else-branch covers Vf==1 path below
            // (Vf == 1: fall through to report-name entry)
            this.drawReportNameEntry(false);
        } else if (this.Vf == 2) {
            this.drawReportAbuse(-28949);
        } else if (this.Bj == 0) {
            inWorld = true;               // a trade/social text box is closed → show world
        } else {
            this.drawDuelConfirm((byte) 127);
        }

        // NOTE on the Vf/Bj chain (clean source, lines 1395-1419):
        //   if (Vf != 1) { if (Vf==2) z(); if (Bj==0) inWorld=true; else h(127); }
        //   <then unconditionally> d(false)=drawReportNameEntry
        // The Vineflower control flow falls through report-abuse/trade-confirm into
        // drawReportNameEntry; the if/else-if ladder above preserves the same outcomes
        // because exactly one branch fires per tick given the mutually-exclusive panel flags.

        // ---- queued-action flush (runs every tick, even with a panel open) ----
        if (this.gc != 0) {
            this.sendQueuedActions((byte) -43);
        }

        // ---- in-world frame (only when no panel is open: Bj==0 path set inWorld) ----
        if (inWorld) {
            if (this.Ph) {
                this.sendDialogAnswer(-312);
            }
            // combat-style change pending (wi.y 8 or 9 = attacking)
            if (this.wi.y == 8 || this.wi.y == 9) {
                this.sendCombatStyle((byte) 114);
            }
            this.drawMinimap(param ^ 1);

            boolean showLists = !this.Ph && !this.se;
            if (showLists) {
                this.friendsList.d(0);   // scroll friends/menu list to top
            }
            // qc selects the active main tab
            if (this.qc == 0 && showLists) this.drawGameFrame(param ^ 2); // s = world render
            if (this.qc == 1) this.drawGameSettings(-15252, showLists);    // a(int,boolean)
            if (this.qc == 2) this.drawGameOptions(showLists, (byte) 125); // a(boolean,byte)
            if (this.qc == 3) this.loadEntitySprites(showLists, param ^ 0);// c(boolean,int)
            if (this.qc == 4) this.b(showLists, (byte) -74);
            if (this.qc == 5) this.a(showLists, false);
            if (this.qc == 6) this.b(15, showLists);

            if (!this.se && !this.Ph) this.drawPlayerMenu(-128);
            if (this.se && !this.Ph)  this.updateTimers((byte) -106);
        }

        this.Cf = 0;
    }

    // -------------------------------------------------------------------------
    // drawGameFrame  — obf: void f(int)
    // -------------------------------------------------------------------------

    /** Render the in-game screen each tick (param must be 13). Four exclusive modes:
     *    rk != -1 → "System update in progress" banner;
     *    Kg       → the obf w(int) overlay screen (drawCharDesignControls in this build);
     *    Qk       → sleep CAPTCHA screen;
     *    otherwise → rebuild the 3D scene from World + entities, place region/transition
     *                labels, run the world render, then drawActiveInterface + chat tabs + blit.
     *
     *  NOTE: the skeleton names this slot "drawHud". That is misleading — the old "drawHud" body
     *  was ONLY the sleep-CAPTCHA `else` sub-branch of this method, lifted out and mislabelled as
     *  the whole thing (and it poked the wrong flags). The real f(int) is the per-tick game-frame
     *  driver; renamed drawGameFrame and rewritten in full from the clean source. */
    private final void drawGameFrame(int param) {
        if (param != 13) return;

        if (this.rk != -1) {
            // --- system-update countdown banner ---
            this.surface.b(0xF8F8F9);
            this.surface.a(this.Wd / 2, STRINGS[371], 0xFF0000, 0, 7, this.Oi / 2);
            this.drawChatHistoryTabs(param - 8);
            this.surface.a(this.graphics, this.Eb, 256, this.K);
            return;
        }
        if (this.Kg) {
            this.drawCharDesignControls(-13759);   // obf w(int): the Kg-screen overlay
            return;
        }

        if (!this.Qk) {
            // --- normal in-world: rebuild + render the 3D scene ---
            if (this.scene.Z) {
                // hide/show object models per active door/curtain layer of current floor (yj)
                for (int i = 0; i < 64; i++) {
                    this.world.a(this.scene.db[this.yj][i], -1);
                    if (this.yj == 0) {
                        this.world.a(this.scene.g[1][i], -1);
                        this.world.a(this.scene.db[1][i], param - 14);
                        this.world.a(this.scene.g[2][i], param ^ -14);
                        this.world.a(this.scene.db[2][i], -1);
                    }
                    this.zf = true;
                    // if we are on the ground floor and standing under a roof tile, hide upper floors
                    if (this.yj == 0 && (this.scene.bb[this.wi.i / 128][this.wi.K / 128] & 128) == 0) {
                        this.world.a(this.scene.db[this.yj][i], (byte) 118);
                        if (this.yj == 0) {
                            this.world.a(this.scene.g[1][i], (byte) 118);
                            this.world.a(this.scene.db[1][i], (byte) 118);
                            this.world.a(this.scene.g[2][i], (byte) 118);
                            this.world.a(this.scene.db[2][i], (byte) 118);
                        }
                        this.zf = false;
                    }
                }

                // region-name announcement banners, latched on region change
                if (this.bl != this.Mg) {
                    this.bl = this.Mg;
                    for (int i = 0; i < this.eh; i++) {
                        // vc[i] holds scenery model ids triggering location labels
                        if (this.vc[i] == 97)   this.a((byte) 48, i, STRINGS[376] + (this.Mg + 1));
                        if (this.vc[i] == 274)  this.a((byte) 58, i, STRINGS[361] + (this.Mg + 1));
                        if (this.vc[i] == 1031) this.a((byte) 103, i, STRINGS[364] + (this.Mg + 1));
                        if (this.vc[i] == 1036) this.a((byte) 89, i, STRINGS[375] + (this.Mg + 1));
                        if (this.vc[i] == 1147) this.a((byte) 18, i, STRINGS[379] + (this.Mg + 1));
                    }
                }
                if (this.yg != this.Nc) {
                    this.yg = this.Nc;
                    for (int i = 0; i < this.eh; i++) {
                        if (this.vc[i] == 51)  this.a((byte) 23, i, STRINGS[368] + (this.Nc + 1));
                        if (this.vc[i] == 143) this.a((byte) 100, i, STRINGS[381] + (this.Nc + 1));
                    }
                }
                if (this.Sg != this.pj) {
                    this.Sg = this.pj;
                    for (int i = 0; i < this.eh; i++) {
                        if (this.vc[i] == 1142) this.a((byte) 89, i, STRINGS[372] + (this.pj + 1));
                    }
                }

                // (re)place the per-tick scene GameModels (players, npcs, ground items, projectiles)
                this.world.a((byte) 67, this.qe);   // clear last frame's dynamic models
                this.qe = 0;

                // --- players in view (this-tick list rg, count Yc) ---
                for (int i = 0; i < this.Yc; i++) {
                    ta player = this.rg[i];
                    if (player.A != 255) {
                        int px = player.i;
                        int pz = player.K;
                        int py = -this.scene.f(px, pz, 125);
                        int model = this.world.a(i + 5000, pz, i + 10000, px, py, 145, 220, (byte) 109);
                        this.qe++;
                        if (this.wi == player) this.world.c(32768, model);
                        if (player.y == 8) this.world.b(param + 24, model, -30);
                        if (player.y == 9) this.world.b(param ^ 45, model, 30);
                    }
                }
                // --- bubble/loot-overhead models for players (b == projectile/ranged target) ---
                for (int i = 0; i < this.Yc; i++) {
                    ta player = this.rg[i];
                    if (player.w != 0) {           // has an active projectile
                        ta target = null;
                        if (player.h == -1) {
                            if (player.z != -1) target = this.npcsCache[player.z];  // ~z != 0 ⟺ z != -1
                        } else {
                            target = this.playersCache[player.h];
                        }
                        if (target != null) {
                            int sx = player.i;
                            int sz = player.K;
                            int sy = -this.scene.f(sx, sz, param ^ 105) - 110;
                            int tx = target.i;
                            int tz = target.K;
                            int ty = -this.scene.f(tx, tz, -22) - b.h[target.t] / 2;
                            // interpolate projectile position by progress player.w / nc
                            int ix = (tx * (this.nc - player.w) + sx * player.w) / this.nc;
                            int iy = (sy * player.w + ty * (this.nc - player.w)) / this.nc;
                            int iz = ((this.nc - player.w) * tz + sz * player.w) / this.nc;
                            this.world.a(player.a + this.kd, iz, 0, ix, iy, 32, 32, (byte) 109);
                            this.qe++;
                        }
                    }
                }
                // --- npcs in view (this-tick list Tb, count de) ---
                for (int i = 0; i < this.de; i++) {
                    ta npc = this.Tb[i];
                    int nx = npc.i;
                    int nz = npc.K;
                    int ny = -this.scene.f(nx, nz, -69);
                    int model = this.world.a(20000 + i, nz, i + 30000, nx, ny, fb.c[npc.t], b.h[npc.t], (byte) 109);
                    this.qe++;
                    if (npc.y == 8) this.world.b(86, model, -30);
                    if (npc.y == 9) this.world.b(param ^ 99, model, 30);
                }
                // --- ground items (Ah of them) ---
                for (int i = 0; i < this.Ah; i++) {
                    int gx = this.Zf[i] * this.Ug + 64;
                    int gz = this.Ug * this.Ni[i] + 64;
                    this.world.a(40000 + this.Gj[i], gz, i + 20000, gx,
                        -this.scene.f(gx, gz, 100) - this.Le[i], 96, 64, (byte) 109);
                    this.qe++;
                }
                // --- decorative/scenery overlay models (el of them) ---
                for (int i = 0; i < this.el; i++) {
                    int dx = 64 + this.Ug * this.Sc[i];
                    int dz = this.gi[i] * this.Ug + 64;
                    int kind = this.Oc[i];
                    if (kind == 0) {
                        this.world.a(50000 + i, dz, i + 50000, dx,
                            -this.scene.f(dx, dz, 98), 128, 256, (byte) 109);
                        this.qe++;
                    }
                    if (kind == 1) {
                        this.world.a(i + 50000, dz, i + 50000, dx,
                            -this.scene.f(dx, dz, param + 58), 128, 64, (byte) 109);
                        this.qe++;
                    }
                }

                this.surface.i = false;
                this.surface.a(true);
                this.surface.i = this.U;

                // occasional ambient sparkle/firework on the upper floors
                if (this.yj == 4) {
                    int n1 = 40 + (int) (3.0 * Math.random());
                    int n2 = (int) (7.0 * Math.random()) + 40;
                    this.world.a(-50, n2, 0, -50, n1, -10);
                }

                // --- camera ---
                this.jc = 0;
                this.Bc = 0;
                this.Ef = 0;
                if (this.Td) {                       // auto-camera mode
                    if (this.Kh && !this.zf) {
                        int prev = this.si;
                        this.updateCamera2((byte) 22);   // q(byte) — auto-rotate toward target
                        if (this.si != prev) {
                            this.Si = this.wi.K;
                            this.kg = this.wi.i;
                        }
                    }
                    this.ug = 32 * this.si;
                    this.scene.Mb = 3000;
                    this.scene.X = 3000;
                    this.scene.P = 1;
                    this.scene.G = 2800;
                    int cx = this.kg + this.Be;
                    int cz = this.Si + this.oc;
                    this.scene.a(cx, cz, 2000, 912, param - 12362, 4 * this.ug,
                        -this.scene.f(cx, cz, -88), 0);
                } else {
                    if (this.Kh && !this.zf) {
                        this.updateCamera2((byte) 94);
                    }
                    if (!this.U) {
                        this.scene.P = 1;
                        this.scene.Mb = 2400;
                        this.scene.G = 2300;
                        this.scene.X = 2400;
                    } else {
                        this.scene.P = 1;
                        this.scene.Mb = 2200;
                        this.scene.X = 2200;
                        this.scene.G = 2100;
                    }
                    int cx = this.kg + this.Be;
                    int cz = this.Si + this.oc;
                    this.scene.a(cx, cz, 2 * this.ac, 912, -12349, this.ug * 4,
                        -this.scene.f(cx, cz, 105), 0);
                }

                // --- run the world render and overlays ---
                this.world.c(-113);             // render scene → surface
                this.drawChat(param - 11);      // l(int): damage splats / ground-item sprites / health bars

                // walk-target click marker (xh = ttl)
                if (this.xh > 0) {
                    this.surface.b(-1, 14 + this.tg + (24 - this.xh) / 6, this.Fd - 8, this.tj - 8);
                }
                if (this.xh < 0) {
                    this.surface.b(-1, 18 + this.tg + (this.xh + 24) / 6, this.Fd - 8, this.tj - 8);
                }

                // system-update countdown text (kc = ticks remaining * 50)
                if (this.kc != -1) {
                    int secs = this.kc / 50;
                    int mins = secs / 60;
                    secs %= 60;
                    if (secs < 10) {
                        this.surface.a(256, STRINGS[380] + mins + STRINGS[365] + secs,
                            0xFFFF00, 0, 1, this.Oi - 7);   // ":0" + secs
                    } else {
                        this.surface.a(256, STRINGS[380] + mins + ":" + secs,
                            0xFFFF00, 0, 1, this.Oi - 7);
                    }
                }

                // wilderness depth indicator ("Wilderness level: N") based on Y past the wall
                if (!this.Ub) {
                    int depth = -this.sh - this.sk - (this.zg - 2203);
                    if (this.Ki + this.Lf + this.Qg > 2640) {
                        depth = -50;
                    }
                    if (depth > 0) {
                        int level = depth / 6 + 1;
                        this.surface.b(-1, 13 + this.tg, this.Oi - 56, 453);
                        this.surface.a(465, STRINGS[377], 0xFFFF00, 0, 1, this.Oi - 20);
                        this.surface.a(465, STRINGS[362] + level, 0xFFFF00, 0, 1, this.Oi - 7);
                        if (this.le == 0) this.le = 2;
                    }
                    if (this.le == 0 && depth > -10 && depth <= 0) {
                        this.le = 1;
                    }
                }

                // --- friends-tab overlay messages while chat panel is on the friends tab ---
                if (this.Zh == 0) {
                    for (int i = 0; i < 100; i++) {
                        // skill/quest progress flash entries (pa.g = ttl)
                        if (pa.g[i] > 0) {
                            String txt = ub.a[i] + mb.a(aa.k[i], k.G[i], true, n.j[i]);
                            this.surface.a(ja.N[i], this.Oi - 18 - 12 * i, txt, 7, 0xFFFF00, (byte) 26, 1);
                        }
                    }
                }

                // --- chat / quest / private message panels (yd holds the 4 message areas) ---
                this.panelGame.b((byte) 56, this.Fh);
                this.panelGame.b((byte) 80, this.ud);
                this.panelGame.b((byte) 48, this.mc);
                if (this.Zh == 1) {
                    this.panelGame.c(this.Fh, 115);
                } else if (this.Zh == 2) {
                    this.panelGame.c(this.ud, 119);
                } else if (this.Zh == 3) {
                    this.panelGame.c(this.mc, 127);
                }
                ia.i = 2;
                this.panelGame.a((byte) -35);
                ia.i = 0;

                this.surface.a(this.tg, 0, this.surface.u - 200, 128, 3);
                this.drawActiveInterface(0);
                this.surface.xb = false;
                this.drawChatHistoryTabs(param - 8);
                this.surface.a(this.graphics, this.Eb, 256, this.K);
            }
        } else {
            // --- sleep CAPTCHA screen (Qk == true) ---
            this.surface.b(0xF8F8F9);
            // scattered decorative "sleeping" words (~15% each, from each side)
            if (Math.random() < 0.15) {
                this.surface.a((int) (Math.random() * 80.0), STRINGS[378],
                    (int) (1.6777215E7 * Math.random()), 0, 5, (int) (334.0 * Math.random()));
            }
            if (Math.random() < 0.15) {
                this.surface.a(512 - (int) (80.0 * Math.random()), STRINGS[378],
                    (int) (Math.random() * 1.6777215E7), param ^ 13, 5, (int) (334.0 * Math.random()));
            }
            this.surface.a(this.Wd / 2 - 100, (byte) -103, 0, 160, 40, 200);
            this.surface.a(this.Wd / 2, STRINGS[366], 0xFFFF00, param - 13, 7, 50);   // "Enter the word..."
            this.surface.a(this.Wd / 2, STRINGS[373] + 100 * this.pg / 750 + "%",
                0xFFFF00, param - 13, 7, 90);                                          // fatigue %
            this.surface.a(this.Wd / 2, STRINGS[367], 0xFFFFFF, 0, 5, 140);
            this.surface.a(this.Wd / 2, STRINGS[374], 0xFFFFFF, param ^ 13, 5, 160);
            this.surface.a(this.Wd / 2, this.e + "*", 0x00FFFF, param - 13, 5, 180);   // typed input
            if (this.Zj != null) {
                this.surface.a(this.Wd / 2, this.Zj, 0xFF0000, 0, 5, 260);            // error message
            }
            this.surface.b(-1, 1 + this.Eh, 230, this.Wd / 2 - 127);                  // CAPTCHA sprite
            this.surface.e(this.Wd / 2 - 128, 257, 229, 27785, 42, 0xFFFFFF);
            this.drawChatHistoryTabs(5);
            this.surface.a(this.Wd / 2, STRINGS[370], 0xFFFFFF, param - 13, 1, 290);
            this.surface.a(this.Wd / 2, STRINGS[369], 0xFFFFFF, param ^ 13, 1, 305);
            this.surface.a(this.graphics, this.Eb, 256, this.K);
        }
    }

    // -------------------------------------------------------------------------
    // drawWildernessWarning  — obf: void H(int)
    // -------------------------------------------------------------------------

    /** "Warning! Proceed with caution" wilderness-entry dialog. Sets le=2 to enter wilderness
     *  mode on click (either on the "Click here to proceed" line or outside the panel bounds). */
    private final void drawWildernessWarning(int param) {
        this.surface.a(86, (byte) -115, 0, 77, 180, 340);   // panel fill at (86,77) 180x340
        int y = 97;
        if (param <= 90) {              // (anti-tamper-safe path) also render the options tab beneath
            this.drawOptionsTab(true);
        }
        this.surface.e(86, 340, 77, 27785, 180, 0xFFFFFF);  // white border

        this.surface.a(256, STRINGS[307], 0xFF0000, 0, 4, y);          // "Warning!"
        this.surface.a(256, STRINGS[305], 0xFFFFFF, 0, 1, y += 26);
        this.surface.a(256, STRINGS[300], 0xFFFFFF, 0, 1, y += 13);
        this.surface.a(256, STRINGS[306], 0xFFFFFF, 0, 1, y += 13);
        this.surface.a(256, STRINGS[308], 0xFFFFFF, 0, 1, y += 22);
        this.surface.a(256, STRINGS[301], 0xFFFFFF, 0, 1, y += 13);
        this.surface.a(256, STRINGS[302], 0xFFFFFF, 0, 1, y += 22);
        this.surface.a(256, STRINGS[303], 0xFFFFFF, 0, 1, y += 13);

        // "Click here to proceed" — red on hover
        int colour = 0xFFFFFF;
        y += 22;
        if (this.xb > y - 12 && this.xb <= y && this.I > 181 && this.I < 331) {
            colour = 0xFF0000;
        }
        this.surface.a(256, STRINGS[126], colour, 0, 1, y);

        if (this.Cf != 0) {
            if (this.xb > y - 12 && this.xb <= y && this.I > 181 && this.I < 331) {
                this.le = 2;
            }
            this.Cf = 0;
            // click anywhere outside the panel rect also confirms
            if (this.I < 86 || this.I > 426 || this.xb < 77 || this.xb > 257) {
                this.le = 2;
            }
        }
    }

    // -------------------------------------------------------------------------
    // drawShop  — obf: void M(int)
    // -------------------------------------------------------------------------

    /** Shop buy/sell panel ("Buying and selling items"). Hit-tests the 4x5 item grid plus the
     *  buy and sell quantity buttons (1/5/10/50/X). Opcodes: 236 BUY_ITEM, 221 SELL_ITEM,
     *  166 CLOSE_SHOP.
     *
     *  FIX vs old: grid hit-test had a duplicated `relY > cellY` (should be `relY < cellY+34`);
     *  the buy "X" dialog used ClientIOException.a (fa.a) not "ArchiveReader.u". */
    private final void drawShop(int param) {
        if (this.Cf != 0 && this.gc == 0) {
            this.Cf = 0;
            int relX = this.I - 52;
            int relY = this.xb - 44;
            // click outside the shop grid → CLOSE_SHOP (clean: break label565 → opcode 166)
            if (relX < 0 || relY < 12 || relX >= 408 || relY >= 246) {
                this.clientStream.b(166, 0);   // CLOSE_SHOP
                this.clientStream.b(21294);
                this.uk = false;
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
                            && this.Rj[slot] != -1) {
                        this.Di = slot;
                        this.fh = this.Rj[slot];
                    }
                    slot++;
                }
            }

            // Di >= 0  (the clean `var44(=0) <= Di`) → an item is selected
            if (this.Di >= 0) {
                int itemId = this.Rj[this.Di];
                if (itemId != -1) {
                    int stock = this.Jf[this.Di];
                    // --- buy row (y 204..215) ---
                    if (stock > 0 && relY >= 204 && relY <= 215) {
                        int qty = 0;
                        if (relX > 318 && relX < 330) qty = 1;
                        if (stock >= 5  && relX > 333 && relX < 345) qty = 5;
                        if (stock >= 10 && relX > 348 && relX < 365) qty = 10;
                        if (stock >= 50 && relX > 368 && relX < 385) qty = 50;
                        if (relX > 388 && relX < 400) {
                            this.drawScrollList(fa.a, 12, 5, true);   // "X" → quantity entry
                        }
                        if (qty > 0) {
                            this.clientStream.b(236, 0);               // BUY_ITEM
                            this.clientStream.f.e(393, this.Rj[this.Di]);
                            this.clientStream.f.e(393, stock);
                            this.clientStream.f.e(393, qty);
                            this.clientStream.b(21294);
                        }
                    }
                    // --- sell row (y 229..240) ---
                    int held = this.b(102, itemId);   // how many the player owns
                    if (held > 0 && relY >= 229 && relY <= 240) {
                        int qty = 0;
                        if (relX > 318 && relX < 330) qty = 1;
                        if (held >= 5  && relX > 333 && relX < 345) qty = 5;
                        if (held >= 10 && relX > 348 && relX < 365) qty = 10;
                        if (relX > 388 && relX < 400) {
                            this.drawScrollList(nb.u, 12, 6, true);   // "X" → quantity entry
                        }
                        if (held >= 50 && relX > 368 && relX < 385) qty = 50;
                        if (qty > 0) {
                            this.clientStream.b(221, 0);               // SELL_ITEM
                            this.clientStream.f.e(393, this.Rj[this.Di]);
                            this.clientStream.f.e(393, stock);
                            this.clientStream.f.e(393, qty);
                            this.clientStream.b(21294);
                        }
                    }
                }
            }
        }

        // --- draw panel ---
        final int px = 52, py = 44;
        this.surface.a(px, (byte) 101, 192, py, 12, 408);
        int grey = 0x989898;
        this.surface.c(160, px, 17, 0, py + 12, 408, grey);
        this.surface.c(160, px, 170, 0, py + 29, 8, grey);
        this.surface.c(160, px + 399, 170, 0, py + 29, 9, grey);
        this.surface.c(160, px, 47, 0, py + 199, 408, grey);
        this.surface.a(STRINGS[640], px + 1, py + 10, 0xFFFFFF, false, 1);   // title

        int closeCol = 0xFFFFFF;
        if (this.I > px + 320 && this.xb >= py && this.I < px + 408 && this.xb < py + 12) {
            closeCol = 0xFF0000;
        }
        this.surface.b(px + 406, STRINGS[620], py + 10, closeCol, -92, 1);   // "Close window"
        this.surface.a(STRINGS[637], px + 2, py + 24, 0x00FF00, false, 1);   // "Buy"
        this.surface.a(STRINGS[635], px + 135, py + 24, 0x00FFFF, false, 1); // "Sell"
        this.surface.a(STRINGS[643] + this.b(84, 10) + STRINGS[631], px + 280, py + 24, 0xFFFF00, false, 1);

        // item grid 4x5
        int grey2 = 0xD0D0D0;
        int slot = 0;
        for (int row = 0; row < 4; row++) {
            for (int col = 0; col < 5; col++) {
                int cellX = col * 49 + 7 + px;
                int cellY = py + 28 + 34 * row;
                if (this.Di == slot) {
                    this.surface.c(160, cellX, 34, 0, cellY, 49, 0xFF0000);
                } else {
                    this.surface.c(160, cellX, 34, 0, cellY, 49, grey2);
                }
                this.surface.e(cellX, 50, cellY, 27785, 35, 0);
                if (this.Rj[slot] != -1) {
                    this.surface.a(cellY, h.c[this.Rj[slot]], 0, false, 0,
                        ua.Bb[this.Rj[slot]] + this.sg, 32, 48, cellX, 1);
                    this.surface.a("" + this.Jf[slot], cellX + 1, cellY + 10, 0x00FF00, false, 1);
                    this.surface.b(cellX + 47, "" + this.b(85, this.Rj[slot]), cellY + 10, 0x00FFFF, -80, 1);
                }
                slot++;
            }
        }
        this.surface.b(398, 0, px + 5, py + 222, (byte) -103);   // scrollbar

        // selected-item detail (buy / sell rows)
        if (this.Di != -1) {
            int itemId = this.Rj[this.Di];
            if (itemId != -1) {
                int stock = this.Jf[this.Di];
                // buy line
                if (stock <= 0) {
                    this.surface.a(px + 204, STRINGS[641], 0xFFFF00, 0, 3, py + 214); // "out of stock"
                } else {
                    int buyPrice = o.a(kb.b[itemId], this.vi[this.Di], this.xk, -30910, true, 1, stock, this.Pf);
                    this.surface.a(ac.x[itemId] + STRINGS[639] + buyPrice + STRINGS[636],
                        px + 2, py + 214, 0xFFFF00, false, 1);
                    boolean inBuyRow = this.xb >= py + 204 && this.xb <= py + 215;
                    this.surface.a(STRINGS[642], px + 285, py + 214, 0xFFFFFF, false, 3);
                    int c = 0xFFFFFF;
                    if (inBuyRow && this.I > px + 318 && this.I < px + 330) c = 0xFF0000;
                    this.surface.a("1", px + 320, py + 214, c, false, 3);
                    if (stock >= 5) {
                        c = 0xFFFFFF;
                        if (inBuyRow && this.I > px + 333 && this.I < px + 345) c = 0xFF0000;
                        this.surface.a("5", px + 335, py + 214, c, false, 3);
                    }
                    if (stock >= 10) {
                        c = 0xFFFFFF;
                        if (inBuyRow && this.I > px + 348 && this.I < px + 365) c = 0xFF0000;
                        this.surface.a(STRINGS[612], px + 350, py + 214, c, false, 3);
                    }
                    if (stock >= 50) {
                        c = 0xFFFFFF;
                        if (inBuyRow && this.I > px + 368 && this.I < px + 385) c = 0xFF0000;
                        this.surface.a(STRINGS[605], px + 370, py + 214, c, false, 3);
                    }
                    c = 0xFFFFFF;
                    if (inBuyRow && this.I > px + 388 && this.I < px + 400) c = 0xFF0000;
                    this.surface.a("X", px + 390, py + 214, c, false, 3);
                }
                // sell line
                int held = this.b(88, itemId);
                if (held <= 0) {
                    this.surface.a(px + 204, STRINGS[632], 0xFFFF00, 0, 3, py + 239); // "shop won't buy"
                } else {
                    int sellPrice = o.a(kb.b[itemId], this.vi[this.Di], this.Nh, -30910, false, 1, stock, this.Pf);
                    this.surface.a(ac.x[itemId] + STRINGS[638] + sellPrice + STRINGS[636],
                        px + 2, py + 239, 0xFFFF00, false, 1);
                    boolean inSellRow = this.xb >= py + 229 && this.xb <= py + 240;
                    this.surface.a(STRINGS[634], px + 285, py + 239, 0xFFFFFF, false, 3);
                    int c = 0xFFFFFF;
                    if (inSellRow && this.I > px + 318 && this.I < px + 330) c = 0xFF0000;
                    this.surface.a("1", px + 320, py + 239, c, false, 3);
                    if (held >= 5) {
                        c = 0xFFFFFF;
                        if (inSellRow && this.I > px + 333 && this.I < px + 345) c = 0xFF0000;
                        this.surface.a("5", px + 335, py + 239, c, false, 3);
                    }
                    if (held >= 10) {
                        c = 0xFFFFFF;
                        if (inSellRow && this.I > px + 348 && this.I < px + 365) c = 0xFF0000;
                        this.surface.a(STRINGS[612], px + 350, py + 239, c, false, 3);
                    }
                    if (held >= 50) {
                        c = 0xFFFFFF;
                        if (inSellRow && this.I > px + 368 && this.I < px + 385) c = 0xFF0000;
                        this.surface.a(STRINGS[605], px + 370, py + 239, c, false, 3);
                    }
                    c = 0xFFFFFF;
                    if (inSellRow && this.I > px + 388 && this.I < px + 400) c = 0xFF0000;
                    this.surface.a("X", px + 390, py + 239, c, false, 3);
                }
                return;
            }
        }
        // nothing selected
        this.surface.a(px + 204, STRINGS[644], 0xFFFF00, 0, 3, py + 214);
    }

    // -------------------------------------------------------------------------
    // drawBank  — obf: void r(int)
    // -------------------------------------------------------------------------

    /** Bank deposit/withdraw panel with up to 4 page tabs. Opcodes: 22 WITHDRAW, 23 DEPOSIT,
     *  212 CLOSE_BANK.  Selected slot is Rd / item sj; page index is xg; vj = items used.
     *
     *  FIX vs old: the withdraw/deposit packet sends live INSIDE the click-handling block
     *  (clean lines 2221-2309), not interleaved with the render of the quantity buttons as the
     *  old version had it. Page-1/2 tabs are only drawn when vj>48 (single page → no tabs).
     *  Withdraw-X uses CacheFile.m (d.m), deposit-X uses RecordLoader.c (f.c). */
    private final void drawBank(int param) {
        final int PANEL_W = 408;
        final int PANEL_H = 334;

        // clamp page index against item count vj
        if (this.xg < 0 && this.vj <= 48)  this.xg = 0;
        if (this.xg > 1 && this.vj <= 96)  this.xg = 1;
        if (this.vj <= this.Rd || this.Rd < 0) this.Rd = -1;
        if (this.xg > 3 && this.vj <= 144) this.xg = 2;   // (clean: xg<-3 && vj<-145 idiom)
        if (this.Rd != -1 && this.sj != this.ae[this.Rd]) {
            this.Rd = -1;
            this.sj = -2;
        }

        // --- click handling ---
        // Returns early (CLOSE_BANK at the bottom) when the click lands outside both the item
        // grid and the page tabs; clicking a tab switches xg; clicking the grid selects a slot
        // and may fire a withdraw/deposit. (clean source label984 / label929 structure.)
        if (this.gc == 0 && this.Cf != 0) {
            this.Cf = 0;
            int relX = PANEL_W / 2 - 256 + this.I;
            int relY = this.xb - (-(PANEL_H / 2) + 170);

            boolean inGrid = !(relX < 0 || relY < 12 || relX >= 408 || relY >= 280);
            boolean closeBank = false;
            if (!inGrid) {
                // page-tab row / close: relY <= 12, columns 50px wide
                if (this.vj > 48 && relX >= 50 && relX <= 115 && relY <= 12) {
                    this.xg = 0;
                } else if (this.vj > 48 && relX >= 115 && relX <= 180 && relY <= 12) {
                    this.xg = 1;
                } else if (this.vj > 96 && relX >= 180 && relX <= 245 && relY <= 12) {
                    this.xg = 2;
                } else if (this.vj > 144 && relX >= 245 && relX <= 311 && relY <= 12) {
                    this.xg = 3;
                } else {
                    closeBank = true;   // any other out-of-grid click closes the bank
                }
            }

            if (closeBank) {
                this.clientStream.b(212, 0);   // CLOSE_BANK
                this.clientStream.b(21294);
                this.Fe = false;
                return;
            }

            if (inGrid) {
                // item-grid hit-test (8 rows x 6 cols, current page = xg*48)
                int srcSlot = this.xg * 48;
                for (int row = 0; row < 8; row++) {
                    for (int col = 0; col < 6; col++) {
                        int cellX = 7 + 49 * col;
                        int cellY = 34 * row + 28;
                        if (cellX < relX && cellX + 49 > relX && cellY < relY && cellY + 34 > relY
                                && srcSlot < this.vj && this.ae[srcSlot] != 0) {
                            this.sj = this.ae[srcSlot];
                            this.Rd = srcSlot;
                        }
                        srcSlot++;
                    }
                }

                // --- withdraw / deposit quantity dispatch for the selected slot ---
                // (uses absolute I/xb against the restored render offsets px/py)
                int px2 = 256 - PANEL_W / 2;
                int py2 = -(PANEL_H / 2) + 170;
                if (this.Rd != -1) {
                    int selItem = this.Rd >= 0 ? this.ae[this.Rd] : -1;
                    if (selItem != 0 && selItem != -1) {
                        int qty = this.di[this.Rd];
                        // withdraw 1 / 5 / 10 / 50 / X / All
                        if (qty >= 1 && this.I >= px2 + 220 && this.xb >= py2 + 238
                                && this.I < px2 + 250 && this.xb <= py2 + 249) {
                            this.bankSend(22, selItem, 1, 0x12345678);
                        }
                        if (qty >= 5 && this.I >= px2 + 250 && this.xb >= py2 + 238
                                && this.I < px2 + 280 && this.xb <= py2 + 249) {
                            this.bankSend(22, selItem, 5, 0x12345678);
                        }
                        if (qty >= 10 && this.I >= px2 + 280 && this.xb >= py2 + 238
                                && this.I < px2 + 305 && this.xb <= py2 + 249) {
                            this.bankSend(22, selItem, 10, 0x12345678);
                        }
                        if (qty >= 50 && this.I >= px2 + 305 && this.xb >= py2 + 238
                                && this.I < px2 + 335 && this.xb <= py2 + 249) {
                            this.bankSend(22, selItem, 50, 0x12345678);
                        }
                        if (this.I >= px2 + 335 && this.xb >= py2 + 238
                                && this.I < px2 + 368 && this.xb <= py2 + 249) {
                            this.drawScrollList(d.m, 12, 3, true);   // withdraw X dialog (CacheFile.m)
                        }
                        if (this.I >= px2 + 370 && this.xb >= py2 + 238
                                && this.I < px2 + 400 && this.xb <= py2 + 249) {
                            this.bankSend(22, selItem, qty, 0x12345678);   // withdraw All
                        }
                        // deposit 1 / 5 / 10 / 50 / X / All  (b(...) = count held in inventory)
                        if (this.b(93, selItem) >= 1 && this.I >= px2 + 220 && this.xb >= py2 + 263
                                && this.I < px2 + 250 && this.xb <= py2 + 274) {
                            this.bankSend(23, selItem, 1, 0x87654321);
                        }
                        if (this.b(90, selItem) >= 5 && this.I >= px2 + 250 && this.xb >= py2 + 263
                                && this.I < px2 + 280 && this.xb <= py2 + 274) {
                            this.bankSend(23, selItem, 5, 0x87654321);
                        }
                        if (this.b(108, selItem) >= 10 && this.I >= px2 + 280 && this.xb >= py2 + 263
                                && this.I < px2 + 305 && this.xb <= py2 + 274) {
                            this.bankSend(23, selItem, 10, 0x87654321);
                        }
                        if (this.b(109, selItem) >= 50 && this.I >= px2 + 305 && this.xb >= py2 + 263
                                && this.I < px2 + 335 && this.xb <= py2 + 274) {
                            this.bankSend(23, selItem, 50, 0x87654321);
                        }
                        if (this.I >= px2 + 335 && this.xb >= py2 + 263
                                && this.I < px2 + 368 && this.xb <= py2 + 274) {
                            this.drawScrollList(f.c, 12, 4, true);   // deposit X dialog (RecordLoader.c)
                        }
                        if (this.I >= px2 + 370 && this.xb >= py2 + 263
                                && this.I < px2 + 400 && this.xb <= py2 + 274) {
                            this.bankSend(23, selItem, this.b(85, selItem), 0x87654321);   // deposit All
                        }
                    }
                }
            }
        }

        // --- render panel ---
        int px = 256 - PANEL_W / 2;
        int py = 170 - PANEL_H / 2;
        this.surface.a(px, (byte) -126, 192, py, 12, 408);
        int grey = 0x989898;
        this.surface.c(160, px, 17, 0, py + 12, 408, grey);
        this.surface.c(160, px, 204, 0, py + 29, 8, grey);
        this.surface.c(160, px + 399, 204, 0, py + 29, 9, grey);
        this.surface.c(160, px, 47, 0, py + 233, 408, grey);
        this.surface.a(STRINGS[610], px + 1, py + 10, 0xFFFFFF, false, 1);   // "Bank"

        int tabX = 50;
        if (this.vj > 48) {
            // page-1 tab
            int col = 0xFFFFFF;
            if (this.xg == 0) col = 0xFF0000;
            if (px + tabX < this.I && this.xb >= py && px + tabX + 65 > this.I && this.xb < py + 12) {
                col = 0xFFFF00;
            }
            this.surface.a(STRINGS[607], px + tabX, py + 10, col, false, 1);
            tabX += 65;
            // page-2 tab
            col = 0xFFFFFF;
            if (this.xg == 1) col = 0xFF0000;
            if (px + tabX < this.I && this.xb >= py && px + tabX + 65 > this.I && this.xb < py + 12) {
                col = 0xFFFF00;
            }
            this.surface.a(STRINGS[618], px + tabX, py + 10, col, false, 1);
            tabX += 65;
        }
        if (this.vj > 96) {
            int col = 0xFFFFFF;
            if (this.xg == 2) {
                col = 0xFF0000;
            } else if (px + tabX < this.I && this.xb >= py && px + tabX + 65 > this.I && this.xb < py + 12) {
                col = 0xFFFF00;
            }
            this.surface.a(STRINGS[616], px + tabX, py + 10, col, false, 1);
            tabX += 65;
        }
        if (this.vj > 144) {
            int col = 0xFFFFFF;
            if (this.xg == 3) col = 0xFF0000;
            if (px + tabX < this.I && this.xb >= py && px + tabX + 65 > this.I && this.xb > py + 12) {
                col = 0xFFFF00;
            }
            this.surface.a(STRINGS[621], px + tabX, py + 10, col, false, 1);
            tabX += 65;
        }

        int closeCol = 0xFFFFFF;
        if (this.I > px + 320 && this.xb >= py && this.I < px + 408 && this.xb < py + 12) {
            closeCol = 0xFF0000;
        }
        this.surface.b(px + 406, STRINGS[620], py + 10, closeCol, -69, 1);
        this.surface.a(STRINGS[608], px + 7, py + 24, 0x00FF00, false, 1);    // "Withdraw"
        this.surface.a(STRINGS[606], px + 289, py + 24, 0x00FFFF, false, 1);  // "Deposit"

        // item grid 8x6
        int grey2 = 0xD0D0D0;
        int srcSlot = this.xg * 48;
        for (int row = 0; row < 8; row++) {
            for (int col = 0; col < 6; col++) {
                int cellX = col * 49 + px + 7;
                int cellY = row * 34 + py + 28;
                if (srcSlot == this.Rd) {
                    this.surface.c(160, cellX, 34, 0, cellY, 49, 0xFF0000);
                } else {
                    this.surface.c(160, cellX, 34, 0, cellY, 49, grey2);
                }
                this.surface.e(cellX, 50, cellY, 27785, 35, 0);
                if (srcSlot < this.vj && this.ae[srcSlot] != 0) {
                    this.surface.a(cellY, h.c[this.ae[srcSlot]], 0, false, 0,
                        ua.Bb[this.ae[srcSlot]] + this.sg, 32, 48, cellX, 1);
                    this.surface.a("" + this.di[srcSlot], cellX + 1, cellY + 10, 0x00FF00, false, 1);
                    this.surface.b(cellX + 47, "" + this.b(87, this.ae[srcSlot]), cellY + 29, 0x00FFFF, 127, 1);
                }
                srcSlot++;
            }
        }
        this.surface.b(398, 0, px + 5, py + 256, (byte) -87);   // scrollbar

        // selected-slot quantity rows
        if (this.Rd != -1) {
            int selItem = this.Rd >= 0 ? this.ae[this.Rd] : -1;
            if (selItem != 0) {
                int qty = this.di[this.Rd];
                if (fa.e[selItem] == 1 && qty > 1) qty = 1;   // non-stackable cap (~e[]==-2 idiom)
                if (qty > 0) {
                    // "Withdraw <item>" + 1/5/10/50/X/All buttons
                    int c = 0xFFFFFF;
                    this.surface.a(STRINGS[611] + ac.x[selItem], px + 2, py + 248, 0xFFFFFF, false, 1);
                    if (this.I >= px + 220 && this.xb >= py + 238 && this.I < px + 250 && this.xb <= py + 249) c = 0xFF0000;
                    this.surface.a(STRINGS[617], px + 222, py + 248, c, false, 1);     // "1"
                    if (qty >= 5) {
                        c = 0xFFFFFF;
                        if (this.I >= px + 250 && this.xb >= py + 238 && this.I < px + 280 && this.xb <= py + 249) c = 0xFF0000;
                        this.surface.a(STRINGS[619], px + 252, py + 248, c, false, 1); // "5"
                    }
                    if (qty >= 10) {
                        c = 0xFFFFFF;
                        if (this.I >= px + 280 && this.xb >= py + 238 && this.I < px + 305 && this.xb <= py + 249) c = 0xFF0000;
                        this.surface.a(STRINGS[612], px + 282, py + 248, c, false, 1); // "10"
                    }
                    if (qty >= 50) {
                        c = 0xFFFFFF;
                        if (this.I >= px + 305 && this.xb >= py + 238 && this.I < px + 335 && this.xb <= py + 249) c = 0xFF0000;
                        this.surface.a(STRINGS[605], px + 307, py + 248, c, false, 1); // "50"
                    }
                    c = 0xFFFFFF;
                    if (this.I >= px + 335 && this.xb >= py + 238 && this.I < px + 368 && this.xb <= py + 249) c = 0xFF0000;
                    this.surface.a("X", px + 337, py + 248, c, false, 1);
                    c = 0xFFFFFF;
                    if (this.I >= px + 370 && this.xb >= py + 238 && this.I < px + 400 && this.xb <= py + 249) c = 0xFF0000;
                    this.surface.a(STRINGS[615], px + 370, py + 248, c, false, 1);     // "All"
                }
                // "Deposit <item>" + 1/5/10/50/X/All buttons (only if the player owns any)
                if (this.b(126, selItem) > 0) {
                    this.surface.a(STRINGS[614] + ac.x[selItem], px + 2, py + 273, 0xFFFFFF, false, 1);
                    int c = 0xFFFFFF;
                    if (this.I >= px + 220 && this.xb >= py + 263 && this.I < px + 250 && this.xb <= py + 274) c = 0xFF0000;
                    this.surface.a(STRINGS[617], px + 222, py + 273, c, false, 1);
                    if (this.b(88, selItem) >= 5) {
                        c = 0xFFFFFF;
                        if (this.I >= px + 250 && this.xb >= py + 263 && this.I < px + 280 && this.xb <= py + 274) c = 0xFF0000;
                        this.surface.a(STRINGS[619], px + 252, py + 273, c, false, 1);
                    }
                    if (this.b(93, selItem) >= 10) {
                        c = 0xFFFFFF;
                        if (this.I >= px + 280 && this.xb >= py + 263 && this.I < px + 305 && this.xb <= py + 274) c = 0xFF0000;
                        this.surface.a(STRINGS[612], px + 282, py + 273, c, false, 1);
                    }
                    if (this.b(98, selItem) >= 50) {
                        c = 0xFFFFFF;
                        if (this.I >= px + 305 && this.xb >= py + 263 && this.I < px + 335 && this.xb <= py + 274) c = 0xFF0000;
                        this.surface.a(STRINGS[605], px + 307, py + 273, c, false, 1);
                    }
                    c = 0xFFFFFF;
                    if (this.I >= px + 335 && this.xb >= py + 263 && this.I < px + 368 && this.xb <= py + 274) c = 0xFF0000;
                    this.surface.a("X", px + 337, py + 273, c, false, 1);
                    c = 0xFFFFFF;
                    if (this.I >= px + 370 && this.xb >= py + 263 && this.I < px + 400 && this.xb <= py + 274) c = 0xFF0000;
                    this.surface.a(STRINGS[615], px + 370, py + 273, c, false, 1);
                }
                return;
            }
        }
        this.surface.a(px + 204, STRINGS[613], 0xFFFF00, 0, 3, py + 248);   // "Select an item"
    }

    /** Helper used by drawBank's click dispatch: begin a bank op (22 withdraw / 23 deposit),
     *  write the item id, the amount, and the obfuscated session "magic" word, then flush.
     *  (In the obfuscated source these were five inlined Jh writes per button.) */
    private final void bankSend(int opcode, int itemId, int amount, int magic) {
        this.clientStream.b(opcode, 0);
        this.clientStream.f.e(393, itemId);
        this.clientStream.f.b(-422797528, amount);
        this.clientStream.f.b(-422797528, magic);
        this.clientStream.b(21294);
    }

    // -------------------------------------------------------------------------
    // drawTrade  — obf: void n(byte)
    // -------------------------------------------------------------------------

    /** Trade offer window: your inventory (left, lc items in vf/xe drawn from px+217), their
     *  current offer (Qf/jj, mf items) and your committed offer (zj/Dd, Lk items). Handles a
     *  right-click "offer N" sub-menu via the ignoreList MessageList (Wf). Opcodes: 55
     *  ACCEPT_TRADE, 230 DECLINE_TRADE; offers go through sendTradeOffer/sendDuelOffer.
     *
     *  FIX vs old: old version had only a stub render with the wrong arrays (zc/of/wj are the
     *  DUEL buffers) and omitted the Cf==2 right-click menu builder and the third (zj/Dd) grid.
     *  Rewritten in full from the clean source. */
    private final void drawTrade(byte param) {
        int menuPick = -1;
        if (this.Cf != 0 && this.lh) {
            menuPick = this.ignoreList.b(this.I, this.Gf, this.Bf, (byte) -40, this.xb);
        }

        if (menuPick < 0) {
            if (this.gc == 0) {
                if (this.Cf == 1 && this.Tk == 0) this.Tk = 1;
                int relX = this.I - 22;
                int relY = this.xb - 36;
                boolean inPanel = !(relX < 0 || relY < 0 || relX >= 469 || relY >= 262);
                if (!inPanel) {
                    if (this.Cf == 1) {          // click outside → decline
                        this.Hk = false;
                        this.clientStream.b(230, 0);
                        this.clientStream.b(21294);
                    }
                } else {
                    // --- left mouse: remove an offered item / accept / decline ---
                    if (this.Tk > 0) {
                        // your-offer grid (217..462 x, 31..235 y, 5 cols) → remove
                        if (relX > 216 && relY > 30 && relX < 462 && relY < 235) {
                            int slot = 5 * ((relY - 31) / 34) + (relX - 217) / 49;
                            if (slot >= 0 && slot < this.lc) {
                                this.drawTradeConfirm(-1, (byte) 9, slot);  // a(int,byte,int): remove 1
                            }
                        }
                        // their-offer grid (8..205 x, 31..133 y, 4 cols) → remove
                        if (relX > 8 && relY > 30 && relX < 205 && relY < 133) {
                            int slot = (relY - 31) / 34 * 4 + (relX - 9) / 49;
                            if (slot >= 0 && slot < this.mf) {
                                this.sendTradeOffer(-1, (byte) 125, slot); // c(int,byte,int)
                            }
                        }
                        // accept button (217..286 x, 238..259 y)
                        if (relX >= 217 && relY >= 238 && relX <= 286 && relY <= 259) {
                            this.Mi = true;
                            this.clientStream.b(55, 0);
                            this.clientStream.b(21294);
                        }
                        // decline button (394..462 x, 238..258 y)
                        if (relX >= 394 && relY >= 238 && relX < 463 && relY < 259) {
                            this.Hk = false;
                            this.clientStream.b(230, param - 8);
                            this.clientStream.b(21294);
                        }
                        this.Tk = 0;
                        this.Cf = 0;
                    }

                    // --- right mouse (Cf==2): open an "offer 1/5/10/all/cancel" sub-menu ---
                    if (this.Cf == 2) {
                        // over your-offer grid → menu for an inventory item
                        if (relX > 216 && relY > 30 && relX < 462 && relY < 235) {
                            int w = this.friendsList.b(16256);
                            int hgt = this.friendsList.a(-21224);
                            this.fg = this.xb - 7;
                            this.rh = this.I - w / 2;
                            this.se = true;
                            if (this.fg < 0) this.fg = 0;
                            if (this.rh < 0) this.rh = 0;
                            if (this.rh + w > 510) this.rh = 510 - w;
                            if (this.fg + hgt > 316) this.fg = 315 - hgt;

                            int slot = (relY - 31) / 34 * 5 + (relX - 217) / 49;
                            if (slot >= 0 && slot < this.lc) {
                                int itemId = this.vf[slot];
                                this.lh = true;
                                this.ignoreList.d(0);
                                this.ignoreList.a(itemId, STRINGS[34] + ac.x[itemId], 1, STRINGS[172], 1,  param + 3288);
                                this.ignoreList.a(itemId, STRINGS[34] + ac.x[itemId], 1, STRINGS[169], 5,  3296);
                                this.ignoreList.a(itemId, STRINGS[34] + ac.x[itemId], 1, STRINGS[158], 10, 3296);
                                this.ignoreList.a(itemId, STRINGS[34] + ac.x[itemId], 1, STRINGS[174], -1, 3296);
                                this.ignoreList.a(itemId, STRINGS[34] + ac.x[itemId], 1, STRINGS[166], -2, param ^ 3304);
                                int mw = this.ignoreList.b(param ^ 16264);
                                int mh = this.ignoreList.a(-21224);
                                this.Gf = this.I - mw / 2;
                                this.Bf = this.xb - 7;
                                if (this.Gf < 0) this.Gf = 0;
                                if (this.Bf < 0) this.Bf = 0;
                                if (this.Bf + mh > 316) this.Bf = 315 - mh;
                                if (this.Gf + mw > 511) this.Gf = 510 - mw;
                            }
                        }
                        // over their-offer grid → menu for a removable offered item
                        if (relX > 8 && relY > 30 && relX < 205 && relY < 133) {
                            int slot = (relX - 9) / 49 + (relY - 31) / 34 * 4;
                            if (slot >= 0 && slot < this.mf) {
                                int itemId = this.Qf[slot];
                                this.lh = true;
                                this.ignoreList.d(0);
                                this.ignoreList.a(itemId, STRINGS[34] + ac.x[itemId], 2, STRINGS[163], 1,  3296);
                                this.ignoreList.a(itemId, STRINGS[34] + ac.x[itemId], 2, STRINGS[173], 5,  param ^ 3304);
                                this.ignoreList.a(itemId, STRINGS[34] + ac.x[itemId], 2, STRINGS[161], 10, 3296);
                                this.ignoreList.a(itemId, STRINGS[34] + ac.x[itemId], 2, STRINGS[177], -1, 3296);
                                this.ignoreList.a(itemId, STRINGS[34] + ac.x[itemId], 2, STRINGS[170], -2, param ^ 3304);
                                int mw = this.ignoreList.b(16256);
                                int mh = this.ignoreList.a(-21224);
                                this.Gf = this.I - mw / 2;
                                this.Bf = this.xb - 7;
                                if (this.Gf < 0) this.Gf = 0;
                                if (this.Bf < 0) this.Bf = 0;
                                if (mh + this.Bf > 315) this.Bf = 315 - mh;
                                if (mw + this.Gf > 511) this.Gf = 510 - mw;
                            }
                        }
                        this.Cf = 0;
                    }

                    // dismiss the sub-menu when the cursor leaves its bounds
                    if (this.lh) {
                        int mw = this.ignoreList.b(16256);
                        int mh = this.ignoreList.a(-21224);
                        if (this.I < this.Gf - 10 || this.I > this.Gf + mw + 10
                                || this.xb < this.Bf - 10 || this.xb > this.Bf + mh + 10) {
                            this.lh = false;
                        }
                    }
                }
            }
        } else {
            // --- a sub-menu entry was clicked: resolve it to an offer ---
            this.lh = false;
            this.Cf = 0;
            int action = this.ignoreList.a(-91, menuPick);   // 1 = inventory item, else offered item
            int itemId = this.ignoreList.a(true, menuPick);
            int slot = -1;
            int total = 0;
            if (action == 1) {
                for (int i = 0; i < this.lc; i++) {
                    if (this.vf[i] == itemId) {
                        if (slot < 0) slot = i;
                        if (fa.e[itemId] == 0) { total = this.xe[i]; break; }
                        total++;
                    }
                }
            } else {
                for (int i = 0; i < this.mf; i++) {
                    if (this.Qf[i] == itemId) {
                        if (slot < 0) slot = i;
                        if (fa.e[itemId] == 0) { total = this.jj[i]; break; }
                        total++;
                    }
                }
            }
            if (slot >= 0) {
                int amount = this.ignoreList.a((byte) 97, menuPick);
                if (amount == -2) {
                    this.ji = slot;                                    // "X" → open qty entry
                    if (action == 1) {
                        this.drawScrollList(s.e, 12, 1, true);
                    } else {
                        this.drawScrollList(ua.Kb, param ^ 4, 2, true);
                    }
                } else {
                    if (amount == 0) amount = total;                   // "All"
                    if (action == 1) {
                        this.drawTradeConfirm(amount, (byte) 9, slot); // add to your offer
                    } else {
                        this.sendTradeOffer(amount, (byte) 124, slot); // remove from your offer
                    }
                }
            }
        }

        // --- draw panel ---
        if (this.Hk) {
            final int px = 22, py = 36;
            this.surface.a(px, (byte) 117, 192, py, 12, 468);
            int grey = 0x989898;
            this.surface.c(160, px, 18, param - 8, py + 12, 468, grey);
            this.surface.c(160, px, 248, 0, py + 30, 8, grey);
            this.surface.c(160, px + 205, 248, param - 8, py + 30, 11, grey);
            this.surface.c(160, px + 462, 248, param - 8, py + 30, 6, grey);
            this.surface.c(160, px + 8, 22, 0, py + 133, 197, grey);
            this.surface.c(160, px + 8, 20, 0, py + 258, 197, grey);
            this.surface.c(160, px + 216, 43, 0, py + 235, 246, grey);
            int lgrey = 0xD0D0D0;
            this.surface.c(160, px + 8, 103, param - 8, py + 30, 197, lgrey);
            this.surface.c(160, px + 8, 103, 0, py + 155, 197, lgrey);
            this.surface.c(160, px + 216, 205, param - 8, py + 30, 246, lgrey);

            for (int r = 0; r < 4; r++) this.surface.b(197, 0, px + 8, py + 30 + 34 * r, (byte) -98);
            for (int r = 0; r < 4; r++) this.surface.b(197, 0, px + 8, 34 * r + 155 + py, (byte) -29);
            for (int r = 0; r < 7; r++) this.surface.b(246, 0, px + 216, py + 30 + r * 34, (byte) 60);
            for (int c = 0; c < 6; c++) {
                this.surface.b(px + 8 + c * 49, py + 30, 0, 103, 0);
                this.surface.b(c * 49 + 8 + px, py + 155, 0, 103, param ^ 8);
                this.surface.b(px + 216 + c * 49, py + 30, 0, 205, 0);
            }

            this.surface.a(STRINGS[175] + this.cj, px + 1, py + 10, 0xFFFFFF, false, 1);  // "Trade with <name>"
            this.surface.a(STRINGS[164], px + 9, py + 27, 0xFFFFFF, false, 4);            // "Your offer"
            this.surface.a(STRINGS[167], px + 9, py + 152, 0xFFFFFF, false, 4);           // "Opponent's offer"
            this.surface.a(STRINGS[171], px + 216, py + 27, 0xFFFFFF, false, 4);          // "Options"
            if (!this.Mi) {
                this.surface.b(-1, this.tg + 25, py + 238, px + 217);   // accept button sprite
            }
            this.surface.b(-1, this.tg + 26, py + 238, px + 394);       // decline button sprite
            if (this.md) {
                this.surface.a(px + 341, STRINGS[168], 0xFFFFFF, param ^ 8, 1, py + 246);
                this.surface.a(px + 341, STRINGS[165], 0xFFFFFF, 0, 1, py + 256);
            }
            if (this.Mi) {
                this.surface.a(px + 217 + 35, STRINGS[176], 0xFFFFFF, param - 8, 1, py + 246);
                this.surface.a(px + 252, STRINGS[160], 0xFFFFFF, param - 8, 1, py + 256);
            }

            // your inventory grid (vf/xe, 5 cols, starts at px+217)
            for (int i = 0; i < this.lc; i++) {
                int cellX = px + 217 + 49 * (i % 5);
                int cellY = py + 31 + i / 5 * 34;
                this.surface.a(cellY, h.c[this.vf[i]], 0, false, 0,
                    this.sg + ua.Bb[this.vf[i]], 32, 48, cellX, 1);
                if (fa.e[this.vf[i]] == 0) {
                    this.surface.a("" + this.xe[i], cellX + 1, cellY + 10, 0xFFFF00, false, 1);
                }
            }
            // their current offer (Qf/jj, 4 cols, starts at px+9)
            for (int i = 0; i < this.mf; i++) {
                int cellX = i % 4 * 49 + 9 + px;
                int cellY = i / 4 * 34 + py + 31;
                this.surface.a(cellY, h.c[this.Qf[i]], 0, false, 0,
                    ua.Bb[this.Qf[i]] + this.sg, 32, 48, cellX, 1);
                if (fa.e[this.Qf[i]] == 0) {
                    this.surface.a("" + this.jj[i], cellX + 1, cellY + 10, 0xFFFF00, false, 1);
                }
                if (this.I > cellX && cellX + 48 > this.I && cellY < this.xb && this.xb < cellY + 32) {
                    this.surface.a(ac.x[this.Qf[i]] + STRINGS[159] + ga.b[this.Qf[i]],
                        px + 8, py + 273, 0xFFFF00, false, 1);   // examine tooltip
                }
            }
            // your committed offer (zj/Dd, 4 cols, starts at px+9, second block at py+156)
            for (int i = 0; i < this.Lk; i++) {
                int cellX = px + 9 + i % 4 * 49;
                int cellY = py + 156 + 34 * (i / 4);
                this.surface.a(cellY, h.c[this.zj[i]], 0, false, 0,
                    ua.Bb[this.zj[i]] + this.sg, 32, 48, cellX, 1);
                if (fa.e[this.zj[i]] == 0) {
                    this.surface.a("" + this.Dd[i], cellX + 1, cellY + 10, 0xFFFF00, false, 1);
                }
                if (this.I > cellX && cellX + 48 > this.I && this.xb > cellY && this.xb < cellY + 32) {
                    this.surface.a(ac.x[this.zj[i]] + STRINGS[159] + ga.b[this.zj[i]],
                        px + 8, py + 273, 0xFFFF00, false, 1);
                }
            }

            if (this.lh) {
                this.ignoreList.a(this.Bf, this.Gf, this.xb, (byte) -12, this.I);   // render sub-menu
            }
        }
    }

    // -------------------------------------------------------------------------
    // drawTradeConfirm  — obf: void a(int,byte,int)
    // -------------------------------------------------------------------------

    /** Add/adjust items in the local trade-offer buffer (Qf/jj) from inventory slot invSlot,
     *  then push the whole offer with opcode 46 (PLAYER_ADDED_ITEMS_TO_TRADE_OFFER).
     *  `count` is the requested amount (-1 = single decrement via Tk loop).
     *
     *  FIX vs old: stackable cap test is kb.c[itemId] (InputState.c), not "EntityDef.c";
     *  stackable flag is fa.e[itemId] (ClientIOException.e). */
    private final void drawTradeConfirm(int count, byte action, int invSlot) {
        if (action != 9) {
            this.resetPanels((byte) -38);   // p(byte)
        }
        boolean changed = false;
        int dupes = 0;
        int itemId = this.vf[invSlot];

        int offerCount;
        for (int i = 0; ; i++) {
            if (i < this.mf) {
                if (itemId == this.Qf[i]) {
                    if (fa.e[itemId] == 0) {     // stackable
                        if (count >= 0) {
                            this.jj[i] += count;
                            if (this.jj[i] > this.xe[invSlot]) this.jj[i] = this.xe[invSlot];
                            changed = true;
                        } else {
                            for (int k = 0; k < this.Tk; k++) {
                                changed = true;
                                if (this.jj[i] < this.xe[invSlot]) this.jj[i]++;
                            }
                        }
                    }
                    dupes++;
                }
                continue;
            }
            offerCount = this.b(99, itemId);   // current copies already offered
            break;
        }
        if (offerCount <= dupes) changed = true;
        if (kb.c[itemId] == 1) {               // FIX: kb.c (InputState), not EntityDef.c
            changed = true;
            this.drawMenuOptions(false, null, action ^ 9, STRINGS[215], 0, 0, null, null);
        }

        if (!changed) {
            if (count < 0) {
                if (this.mf < 12) {
                    this.Qf[this.mf] = itemId;
                    this.jj[this.mf] = 1;
                    changed = true;
                    this.mf++;
                }
            } else {
                for (int k = 0; k < count; k++) {
                    if (this.mf >= 12 || offerCount <= dupes) break;
                    this.Qf[this.mf] = itemId;
                    this.jj[this.mf] = 1;
                    changed = true;
                    dupes++;
                    this.mf++;
                    if (k == 0 && fa.e[itemId] == 0) {   // first add of a stackable → take min(count,have)
                        this.jj[this.mf - 1] = count <= this.xe[invSlot] ? count : this.xe[invSlot];
                        break;
                    }
                }
            }
        }
        if (!changed) return;

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
    // drawTradeConfirmWindow  — obf: void N(int)
    // -------------------------------------------------------------------------

    /** "Please confirm your trade" window: your final items (Vb/Me, count Ui) and theirs
     *  (Lc/Bi, count nh), with Accept/Decline. Vi = you have accepted. Opcodes: 104
     *  CONFIRM_TRADE, 230 DECLINE_TRADE.
     *
     *  FIX vs old: the previous part file's "drawTradeConfirmWindow" was tied to the wrong obf
     *  method (a(boolean,boolean), the social-list panel) and only a stub. This is the real
     *  N(int) body (clean line 13749). drawActiveInterface dispatches it via the Xj flag. */
    private final void drawTradeConfirmWindow(int param) {
        final int px = 22, py = 36;
        this.surface.a(px, (byte) -117, 192, py, 16, 468);
        int grey = 0x989898;
        this.surface.c(160, px, 246, 0, py + 16, 468, grey);
        this.surface.a(px + 234, STRINGS[204] + this.re, 0xFFFFFF, 0, 1, py + 12);  // "Trade with <name>"
        this.surface.a(px + 117, STRINGS[210], 0xFFFF00, 0, 1, py + 30);            // "You are about to give:"

        // your final offer (Vb ids / Me counts, Ui of them)
        for (int i = 0; i < this.Ui; i++) {
            String name = ac.x[this.Vb[i]];
            if (fa.e[this.Vb[i]] == 0) {     // stackable → append count
                name = name + STRINGS[211] + mb.a(this.Me[i], 131071);
            }
            this.surface.a(px + 117, name, 0xFFFFFF, 0, 1, i * 12 + 42 + py);
        }
        if (this.Ui == 0) {
            this.surface.a(px + 117, STRINGS[213], 0xFFFFFF, 0, 1, py + 42);
        }

        this.surface.a(px + 351, STRINGS[209], 0xFFFF00, 0, 1, py + 30);            // "In return you will receive:"
        for (int i = 0; i < this.nh; i++) {
            String name = ac.x[this.Lc[i]];
            if (fa.e[this.Lc[i]] == 0) {
                name = name + STRINGS[211] + mb.a(this.Bi[i], 131071);
            }
            this.surface.a(px + 351, name, 0xFFFFFF, 0, 1, 42 + py + 12 * i);
        }
        if (this.nh == 0) {
            this.surface.a(px + 351, STRINGS[213], 0xFFFFFF, 0, 1, py + 42);
        }
        if (param >= -6) return;   // anti-tamper guard (clean: var10000>=var10001 → b(true))

        this.surface.a(px + 234, STRINGS[206], 0x00FFFF, 0, 4, py + 200);   // confirm hint lines
        this.surface.a(px + 234, STRINGS[207], 0xFFFFFF, 0, 1, py + 215);
        this.surface.a(px + 234, STRINGS[205], 0xFFFFFF, 0, 1, py + 230);
        if (this.Vi) {
            this.surface.a(px + 234, STRINGS[212], 0xFFFF00, 0, 1, py + 250); // "Waiting..."
        } else {
            this.surface.b(-1, this.tg + 25, py + 238, px + 118 - 35);        // accept sprite
            this.surface.b(-1, this.tg + 26, py + 238, px + 352 - 35);        // decline sprite
        }

        if (this.Cf == 2) {
            // click outside the panel → decline
            if (this.I < px || this.xb < py || this.I > px + 468 || this.xb > py + 262) {
                this.Xj = false;
                this.clientStream.b(230, 0);
                this.clientStream.b(21294);
            }
            // accept button: x 83..188, y 238..259
            if (this.I >= px + 118 - 35 && this.I <= px + 118 + 70 && this.xb >= py + 238 && this.xb <= py + 259) {
                this.Vi = true;
                this.clientStream.b(104, 0);   // CONFIRM_TRADE
                this.clientStream.b(21294);
            }
            // decline button: x 317..423, y 238..259
            if (this.I >= px + 352 - 35 && this.I <= px + 423 && this.xb >= py + 238 && this.xb <= py + 259) {
                this.Xj = false;
                this.clientStream.b(230, 0);
                this.clientStream.b(21294);
            }
            this.Cf = 0;
        }
    }

    // -------------------------------------------------------------------------
    // drawDuelConfirm  — obf: void h(int)
    // -------------------------------------------------------------------------

    /** Duel confirm window: your stake (Nj items in xi/th) vs theirs (Ve items in xj/kf),
     *  the four rule flags, and accept/decline. Opcodes: 77 ACCEPT_DUEL, 197 CONFIRM_DUEL,
     *  230 DECLINE_DUEL.
     *
     *  FIX vs old: the two accept/decline hit-test rectangles had wrong right edges
     *  (old used `+70-35`; clean is left `83..188`, right `317..423`). */
    private final void drawDuelConfirm(int param) {
        final int px = 22, py = 36;
        this.surface.a(px, (byte) -108, 192, py, 16, 468);
        int grey = 0x989898;
        this.surface.c(160, px, 246, 0, py + 16, 468, grey);
        this.surface.a(px + 234, STRINGS[522] + this.Uc, 0xFFFFFF, 0, 1, py + 12);  // "Duel with <name>"
        this.surface.a(px + 117, STRINGS[524], 0xFFFF00, 0, 1, py + 30);            // "Your stake:"

        for (int i = 0; i < this.Nj; i++) {
            String name = ac.x[this.xi[i]];
            if (fa.e[this.xi[i]] == 0) {     // stackable → append count
                name = name + STRINGS[211] + mb.a(this.th[i], 131071);
            }
            this.surface.a(px + 117, name, 0xFFFFFF, 0, 1, 42 + py + 12 * i);
        }
        if (param > -10) return;   // anti-tamper guard (clean: var10000<=var10001 path)

        if (this.Nj == 0) {
            this.surface.a(px + 117, STRINGS[213], 0xFFFFFF, 0, 1, 42 + py);
        }
        this.surface.a(px + 351, STRINGS[527], 0xFFFF00, 0, 1, py + 30);            // "Their stake:"
        for (int i = 0; i < this.Ve; i++) {
            String name = ac.x[this.xj[i]];
            if (fa.e[this.xj[i]] == 0) {
                name = name + STRINGS[211] + mb.a(this.kf[i], 131071);
            }
            this.surface.a(px + 351, name, 0xFFFFFF, 0, 1, i * 12 + 42 + py);
        }
        if (this.Ve == 0) {
            this.surface.a(px + 351, STRINGS[213], 0xFFFFFF, 0, 1, 42 + py);
        }

        // rule flags (Sh retreat / gh magic / Cc prayer / Rc weapons)
        if (this.Sh == 0) {
            this.surface.a(px + 234, STRINGS[528], 0x00FF00, 0, 1, py + 180);   // "Retreat allowed"
        } else {
            this.surface.a(px + 234, STRINGS[517], 0xFF0000, 0, 1, py + 180);   // "No retreat"
        }
        if (this.gh == 0) {
            this.surface.a(px + 234, STRINGS[526], 0x00FF00, 0, 1, py + 192);   // "Magic allowed"
        } else {
            this.surface.a(px + 234, STRINGS[519], 0xFF0000, 0, 1, py + 192);   // "No magic"
        }
        if (this.Cc == 0) {
            this.surface.a(px + 234, STRINGS[516], 0x00FF00, 0, 1, py + 204);   // "Prayer allowed"
        } else {
            this.surface.a(px + 234, STRINGS[521], 0xFF0000, 0, 1, py + 204);   // "No prayer"
        }
        if (this.Rc != 0) {
            this.surface.a(px + 234, STRINGS[518], 0xFF0000, 0, 1, py + 216);   // "No weapons"
        } else {
            this.surface.a(px + 234, STRINGS[525], 0x00FF00, 0, 1, py + 216);   // "Weapons allowed"
        }
        this.surface.a(px + 234, STRINGS[520], 0xFFFFFF, 0, 1, py + 230);       // "Both must confirm"

        if (!this.Cd) {
            this.surface.b(-1, this.tg + 25, py + 238, px + 83);                // accept sprite
            this.surface.b(-1, this.tg + 26, py + 238, px + 352 - 35);          // decline sprite
        } else {
            this.surface.a(px + 234, STRINGS[212], 0xFFFF00, 0, 1, py + 250);   // "Waiting..."
        }

        if (this.Cf == 2) {
            // click outside panel → decline
            if (this.I < px || this.xb < py || this.I > px + 468 || this.xb > py + 262) {
                this.dd = false;
                this.clientStream.b(230, 0);
                this.clientStream.b(21294);
            }
            // accept button: x 83..188, y 238..259   (FIX: was 83..153)
            if (this.I >= px + 118 - 35 && this.I < px + 118 + 70 && this.xb >= py + 238 && this.xb <= py + 259) {
                this.Cd = true;
                this.clientStream.b(77, 0);
                this.clientStream.b(21294);
            }
            // decline button: x 317..423, y 238..259   (FIX: was 317..388)
            if (this.I >= px + 352 - 35 && this.I <= px + 353 + 70 && this.xb >= py + 238 && this.xb <= py + 259) {
                this.dd = false;
                this.clientStream.b(197, 0);
                this.clientStream.b(21294);
            }
            this.Cf = 0;
        }
    }

    // -------------------------------------------------------------------------
    // drawDuel  — obf: void q(int)
    // -------------------------------------------------------------------------

    /** Duel setup window: your stake (lc inventory items vf/xe), opponent's offered items
     *  (Ke items Uf/df), opponent's committed stake (wj items zc/of), and the four rule
     *  checkboxes (No retreat fd / No magic Yi / No prayer vd / No weapons ff). A right-click
     *  on a stake cell opens an "offer N" sub-menu via the chatList MessageList (He).
     *  Opcodes: 8 DUEL_SETTINGS, 176 DUEL_ACCEPT, 197 DUEL_DECLINE.
     *
     *  FIX vs old: old version stubbed the right-click sub-menu builders and the menu-pick
     *  resolution, and mixed up the three render grids. Rewritten in full from the clean source. */
    private final void drawDuel(int param) {
        int menuPick = -1;
        if (this.Cf != 0 && this.Je) {
            menuPick = this.chatList.b(this.I, this.ad, this.Uk, (byte) -40, this.xb);
        }

        if (menuPick >= 0) {
            // --- a sub-menu entry was clicked: resolve to a stake change ---
            this.Cf = 0;
            this.Je = false;
            int action = this.chatList.a(-26, menuPick);   // 3 = your inventory, 4 = their offer
            int itemId = this.chatList.a(true, menuPick);
            int slot = -1;
            int total = 0;
            if (action != 3) {
                for (int i = 0; i < this.Ke; i++) {
                    if (this.Uf[i] == itemId) {
                        if (slot < 0) slot = i;
                        if (fa.e[itemId] == 0) { total = this.df[i]; break; }
                        total++;
                    }
                }
            }
            for (int i = 0; i < this.lc; i++) {
                if (this.vf[i] == itemId) {
                    if (slot < 0) slot = i;
                    if (fa.e[itemId] == 0) { total = this.xe[i]; break; }
                    total++;
                }
            }
            if (slot >= 0) {
                int amount = this.chatList.a((byte) 97, menuPick);
                if (amount != -2) {
                    if (amount == 0) amount = total;   // "All"
                    if (action == 3) {
                        this.sendTradeOffer(param ^ 54, amount, slot);  // add to your stake
                    } else {
                        this.sendDuelOffer(slot, amount, (byte) -78);   // remove from their offer
                    }
                } else {
                    this.ck = slot;                    // "X" → quantity entry dialog
                    if (action == 4) {
                        this.drawScrollList(oa.c, 12, 7, true);
                    } else {
                        this.drawScrollList(n.f, 12, 8, true);
                    }
                }
            }
        } else if (this.gc == 0) {
            if (this.Cf == 1 && this.Tk == 0) this.Tk = 1;
            int relX = this.I - 22;
            int relY = this.xb - 36;
            if (relX >= 0 && relY >= 0 && relX < 469 && relY < 262) {
                // --- left mouse: remove a staked item / toggle rules / accept / decline ---
                if (this.Tk > 0) {
                    // your stake grid (217..462 x, 31..235 y, 5 cols) → remove
                    if (relX > 216 && relY > 30 && relX < 462 && relY < 235) {
                        int slot = (relX - 217) / 49 + 5 * ((relY - 31) / 34);
                        if (slot >= 0 && slot < this.lc) {
                            this.sendTradeOffer(109, -1, slot);
                        }
                    }
                    // their offer grid (8..205 x, 30..129 y, 4 cols) → remove
                    if (relX > 8 && relY > 30 && relX < 205 && relY < 129) {
                        int slot = (relX - 9) / 49 + (relY - 31) / 34 * 4;
                        if (slot >= 0 && slot < this.Ke) {
                            this.sendDuelOffer(slot, -1, (byte) -78);
                        }
                    }
                    // rule checkboxes
                    boolean rulesChanged = false;
                    if (relX >= 93 && relY >= 221 && relX <= 104 && relY <= 232) { this.fd = !this.fd; rulesChanged = true; } // No retreat
                    if (relX >= 93 && relY >= 240 && relX <= 104 && relY <= 251) { this.Yi = !this.Yi; rulesChanged = true; } // No magic
                    if (relX >= 191 && relY >= 221 && relX <= 202 && relY <= 232) { this.vd = !this.vd; rulesChanged = true; } // No prayer
                    if (relX >= 191 && relY >= 240 && relX <= 202 && relY <= 251) { this.ff = !this.ff; rulesChanged = true; } // No weapons
                    if (rulesChanged) {
                        this.clientStream.b(8, 0);   // DUEL_SETTINGS
                        this.clientStream.f.c(this.fd ? 1 : 0, 68);
                        this.clientStream.f.c(this.Yi ? 1 : 0, -100);
                        this.clientStream.f.c(this.vd ? 1 : 0, -96);
                        this.clientStream.f.c(this.ff ? 1 : 0, -107);
                        this.clientStream.b(param ^ 21254);
                        this.ki = false;
                        this.ke = false;
                    }
                    // accept button (218..287 x, 238..259 y)
                    if (relX >= 218 && relY >= 238 && relX <= 287 && relY <= 259) {
                        this.ke = true;
                        this.clientStream.b(176, param - 40);   // DUEL_ACCEPT
                        this.clientStream.b(param + 21254);
                    }
                    // decline button (394..463 x, 238..259 y)
                    if (relX >= 394 && relY >= 238 && relX < 463 && relY < 259) {
                        this.Pj = false;
                        this.clientStream.b(197, 0);            // DUEL_DECLINE
                        this.clientStream.b(21294);
                    }
                    this.Tk = 0;
                    this.Cf = 0;
                }

                // --- right mouse (Cf==3): open an "offer 1/5/10/all/cancel" sub-menu ---
                if (this.Cf == 3) {
                    // over your-stake grid → inventory item menu
                    if (relX > 216 && relY > 30 && relX < 462 && relY < 235) {
                        int w = this.friendsList.b(16256);
                        int hgt = this.friendsList.a(param - 21264);
                        this.rh = this.I - w / 2;
                        this.fg = this.xb - 7;
                        this.se = true;
                        if (this.fg < 0) this.fg = 0;
                        if (this.rh < 0) this.rh = 0;
                        if (this.rh + w > 510) this.rh = 510 - w;
                        if (this.fg + hgt > 315) this.fg = 315 - hgt;

                        int slot = (relX - 217) / 49 + 5 * ((relY - 31) / 34);
                        if (slot >= 0 && slot < this.lc) {
                            int itemId = this.vf[slot];
                            this.Je = true;
                            this.chatList.d(0);
                            this.chatList.a(itemId, STRINGS[34] + ac.x[itemId], 3, STRINGS[502], 1,  param + 3256);
                            this.chatList.a(itemId, STRINGS[34] + ac.x[itemId], 3, STRINGS[509], 5,  param ^ 3272);
                            this.chatList.a(itemId, STRINGS[34] + ac.x[itemId], 3, STRINGS[505], 10, 3296);
                            this.chatList.a(itemId, STRINGS[34] + ac.x[itemId], 3, STRINGS[501], -1, 3296);
                            this.chatList.a(itemId, STRINGS[34] + ac.x[itemId], 3, STRINGS[503], -2, 3296);
                            int mw = this.chatList.b(16256);
                            int mh = this.chatList.a(-21224);
                            this.Uk = this.xb - 7;
                            this.ad = this.I - mw / 2;
                            if (this.ad < 0) this.ad = 0;
                            if (this.Uk < 0) this.Uk = 0;
                            if (this.ad + mw > 510) this.ad = 510 - mw;
                            if (this.Uk + mh > 316) this.Uk = 315 - mh;
                        }
                    }
                    // over their-offer grid → removable item menu
                    if (relX > 8 && relY > 30 && relX < 205 && relY < 133) {
                        int slot = (relX - 9) / 49 + 4 * ((relY - 31) / 34);
                        if (slot >= 0 && slot < this.Ke) {
                            int itemId = this.Uf[slot];
                            this.Je = true;
                            this.chatList.d(0);
                            this.chatList.a(itemId, STRINGS[34] + ac.x[itemId], 4, STRINGS[163], 1,  param ^ 3272);
                            this.chatList.a(itemId, STRINGS[34] + ac.x[itemId], 4, STRINGS[173], 5,  3296);
                            this.chatList.a(itemId, STRINGS[34] + ac.x[itemId], 4, STRINGS[161], 10, 3296);
                            this.chatList.a(itemId, STRINGS[34] + ac.x[itemId], 4, STRINGS[177], -1, 3296);
                            this.chatList.a(itemId, STRINGS[34] + ac.x[itemId], 4, STRINGS[170], -2, param + 3256);
                            int mw = this.chatList.b(16256);
                            int mh = this.chatList.a(-21224);
                            this.Uk = this.xb - 7;
                            this.ad = this.I - mw / 2;
                            if (this.ad < 0) this.ad = 0;
                            if (this.Uk < 0) this.Uk = 0;
                            if (this.ad + mw > 511) this.ad = 510 - mw;
                            if (mh + this.Uk > 315) this.Uk = 315 - mh;
                        }
                    }
                    this.Cf = 0;
                }

                // dismiss the sub-menu when the cursor leaves its bounds
                if (this.Je) {
                    int mw = this.chatList.b(16256);
                    int mh = this.chatList.a(-21224);
                    if (this.ad - 10 > this.I || this.Uk - 10 > this.xb
                            || this.ad + mw + 10 < this.I || this.Uk + mh + 10 < this.xb) {
                        this.Je = false;
                    }
                }
            } else if (this.Cf != 0) {
                // click outside the panel → decline duel
                this.Pj = false;
                this.clientStream.b(197, 0);
                this.clientStream.b(21294);
            }
        }

        // --- draw panel ---
        if (this.Pj) {
            final int px = 22, py = 36;
            this.surface.a(px, (byte) 112, 0xC90B1D, py, 12, 468);   // maroon panel bg
            int grey = 0x989898;
            this.surface.c(160, px, 18, 0, py + 12, 468, grey);
            this.surface.c(160, px, 248, 0, py + 30, 8, grey);
            this.surface.c(160, px + 205, 248, 0, py + 30, 11, grey);
            this.surface.c(160, px + 462, 248, 0, py + 30, 6, grey);
            this.surface.c(160, px + 8, 24, param ^ 40, py + 99, 197, grey);
            this.surface.c(160, px + 8, 23, 0, py + 192, 197, grey);
            this.surface.c(160, px + 8, 20, 0, py + 258, 197, grey);
            this.surface.c(160, px + 216, 43, 0, py + 235, 246, grey);
            int lgrey = 0xD0D0D0;
            this.surface.c(160, px + 8, 69, 0, py + 30, 197, lgrey);
            this.surface.c(160, px + 8, 69, 0, py + 123, 197, lgrey);
            this.surface.c(160, px + 8, 43, param - 40, py + 215, 197, lgrey);
            this.surface.c(160, px + 216, 205, 0, py + 30, 246, lgrey);

            for (int r = 0; r < 3; r++) this.surface.b(197, 0, px + 8, py + 30 + 34 * r, (byte) 58);
            for (int r = 0; r < 4; r++) this.surface.b(197, 0, px + 8, r * 34 + py + 123, (byte) -88);
            for (int r = 0; r < 7; r++) this.surface.b(246, 0, px + 216, r * 34 + py + 30, (byte) -40);
            for (int c = 0; c < 6; c++) {
                this.surface.b(49 * c + 8 + px, py + 30, 0, 69, 0);
                if (c < 5) this.surface.b(49 * c + px + 8, py + 123, 0, 69, 0);
                this.surface.b(c * 49 + px + 216, py + 30, 0, 205, 0);
            }
            this.surface.b(197, 0, px + 8, py + 215, (byte) 97);
            this.surface.b(197, 0, px + 8, py + 257, (byte) 99);
            this.surface.b(px + 8, py + 215, 0, 43, 0);
            this.surface.b(px + 204, py + 215, 0, 43, 0);

            this.surface.a(STRINGS[508] + this.Lg, px + 1, py + 10, 0xFFFFFF, false, 1);  // "Duel with <name>"
            this.surface.a(STRINGS[498], px + 9, py + 27, 0xFFFFFF, false, 4);            // "Your stake"
            this.surface.a(STRINGS[500], px + 9, py + 120, 0xFFFFFF, false, 4);           // "Opponent's stake"
            this.surface.a(STRINGS[499], px + 9, py + 212, 0xFFFFFF, false, 4);           // "Their offer"
            this.surface.a(STRINGS[171], px + 216, py + 27, 0xFFFFFF, false, 4);          // "Options"
            this.surface.a(STRINGS[506], px + 8 + 1, py + 215 + 16, 0xFFFF00, false, 3);
            this.surface.a(STRINGS[496], px + 8 + 1, py + 250, 0xFFFF00, false, 3);
            this.surface.a(STRINGS[507], px + 8 + 102, py + 231, 0xFFFF00, false, 3);
            this.surface.a(STRINGS[497], px + 8 + 102, py + 215 + 35, 0xFFFF00, false, 3);

            // rule checkboxes (box + tick)
            this.surface.e(px + 93, 11, py + 215 + 6, param + 27745, 11, 0xFFFF00);
            if (this.fd) this.surface.a(px + 95, (byte) -109, 0xFFFF00, py + 215 + 8, 7, 7);
            this.surface.e(px + 93, 11, py + 215 + 25, 27785, 11, 0xFFFF00);
            if (this.Yi) this.surface.a(px + 95, (byte) -127, 0xFFFF00, py + 215 + 27, 7, 7);
            this.surface.e(px + 191, 11, py + 215 + 6, 27785, 11, 0xFFFF00);
            if (this.vd) this.surface.a(px + 193, (byte) -106, 0xFFFF00, py + 215 + 8, 7, 7);
            this.surface.e(px + 191, 11, py + 215 + 25, param + 27745, 11, 0xFFFF00);
            if (this.ff) this.surface.a(px + 193, (byte) 59, 0xFFFF00, py + 215 + 27, 7, 7);

            if (!this.ke) this.surface.b(-1, this.tg + 25, py + 238, px + 217);   // accept sprite
            this.surface.b(-1, this.tg + 26, py + 238, px + 394);                 // decline sprite
            if (this.ki) {
                this.surface.a(px + 341, STRINGS[168], 0xFFFFFF, 0, 1, py + 246);
                this.surface.a(px + 341, STRINGS[165], 0xFFFFFF, 0, 1, py + 256);
            }
            if (this.ke) {
                this.surface.a(px + 217 + 35, STRINGS[176], 0xFFFFFF, 0, 1, py + 246);
                this.surface.a(px + 252, STRINGS[160], 0xFFFFFF, 0, 1, py + 256);
            }

            // your stake grid (vf/xe, 5 cols, starts at px+217)
            for (int i = 0; i < this.lc; i++) {
                int cellX = px + 217 + i % 5 * 49;
                int cellY = py + 31 + 34 * (i / 5);
                this.surface.a(cellY, h.c[this.vf[i]], 0, false, 0,
                    this.sg + ua.Bb[this.vf[i]], 32, 48, cellX, 1);
                if (fa.e[this.vf[i]] == 0) {
                    this.surface.a("" + this.xe[i], cellX + 1, cellY + 10, 0xFFFF00, false, 1);
                }
            }
            // opponent's offered items (Uf/df, 4 cols, starts at px+9)
            for (int i = 0; i < this.Ke; i++) {
                int cellX = px + 9 + i % 4 * 49;
                int cellY = py + 31 + i / 4 * 34;
                this.surface.a(cellY, h.c[this.Uf[i]], 0, false, 0,
                    this.sg + ua.Bb[this.Uf[i]], 32, 48, cellX, param - 39);
                if (fa.e[this.Uf[i]] == 0) {
                    this.surface.a("" + this.df[i], cellX + 1, cellY + 10, 0xFFFF00, false, 1);
                }
                if (cellX < this.I && cellX + 48 > this.I && cellY < this.xb && cellY + 32 > this.xb) {
                    this.surface.a(ac.x[this.Uf[i]] + STRINGS[159] + ga.b[this.Uf[i]],
                        px + 8, py + 273, 0xFFFF00, false, 1);   // examine tooltip
                }
            }
            // opponent's committed stake (zc/of, 4 cols, second block at py+124)
            for (int i = 0; i < this.wj; i++) {
                int cellX = px + 9 + i % 4 * 49;
                int cellY = py + 124 + i / 4 * 34;
                this.surface.a(cellY, h.c[this.zc[i]], 0, false, 0,
                    ua.Bb[this.zc[i]] + this.sg, 32, 48, cellX, param ^ 41);
                if (fa.e[this.zc[i]] == 0) {
                    this.surface.a("" + this.of[i], cellX + 1, cellY + 10, 0xFFFF00, false, 1);
                }
                if (this.I > cellX && cellX + 48 > this.I && this.xb > cellY && cellY + 32 > this.xb) {
                    this.surface.a(ac.x[this.zc[i]] + STRINGS[159] + ga.b[this.zc[i]],
                        px + 8, py + 273, 0xFFFF00, false, 1);
                }
            }

            if (this.Je) {
                this.chatList.a(this.Uk, this.ad, this.xb, (byte) -12, this.I);   // render sub-menu
            }
        }
    }

    // -------------------------------------------------------------------------
    // drawReportAbuse  — obf: void z(int)
    // -------------------------------------------------------------------------

    /** Report-abuse rule picker. Yb is the selected rule id (column base 1/7/12 from mouse X,
     *  plus the clicked row). On confirm sends opcode 206 (REPORT_ABUSE) with the player name
     *  (ec), the rule id, and the mute flag (ue).
     *
     *  FIX vs old: the old version used wrong string indices and an invented row layout. This is
     *  the faithful render from the clean source (3 columns x several rows of li.a/li.e calls). */
    private final void drawReportAbuse(int param) {
        // pick the column (rule base) from mouse X
        this.Yb = 0;
        boolean inColumn = true;
        if (this.I >= 36 && this.I < 176) {
            this.Yb = 1;
        } else if (this.I >= 186 && this.I < 326) {
            this.Yb = 7;
        } else if (this.I >= 336 && this.I < 476) {
            this.Yb = 12;
        } else {
            inColumn = false;
        }

        // within the column, add the clicked row to the base
        int y = 156;
        if (inColumn) {
            boolean rowHit = false;
            for (int row = 0; row < 6; row++) {
                int rowH = (row == 0) ? 30 : 18;
                if (this.xb > y - 12 && this.xb < y - 12 + rowH) {
                    if (this.Yb == 1)               { rowHit = true; this.Yb += row; }
                    else if (this.Yb == 7 && row < 5){ rowHit = true; this.Yb += row; }
                    else if (this.Yb == 12 && row < 3){ rowHit = true; this.Yb += row; }
                }
                y += 2 + rowH;
            }
            if (!rowHit) this.Yb = 0;
        } else {
            this.Yb = 0;
        }

        if (this.Cf != 0 && this.Yb != 0) {
            this.clientStream.b(206, param + 28949);   // REPORT_ABUSE
            this.clientStream.f.a(this.ec, 113);        // reported player name
            this.clientStream.f.c(this.Yb, 74);         // rule id
            this.clientStream.f.c(this.ue ? 1 : 0, -68);// mute flag
            this.clientStream.b(param ^ -8763);
            this.Vf = 0;
            this.Cb = "";
            this.e = "";
            this.Cf = 0;
            return;
        }

        y += 15;
        if (this.Cf != 0) {
            this.Cf = 0;
            // click outside the panel → close
            if (this.I < 31 || this.xb < 35 || this.I > 481 || this.xb > 310) {
                this.Vf = 0;
                return;
            }
            // click on the "Send report" link area → close
            if (this.I > 67 && this.I < 446 && this.xb >= y - 15 && this.xb < y + 5) {
                this.Vf = 0;
                return;
            }
        }

        // --- render panel ---
        this.surface.a(31, (byte) -110, 0, 35, 275, 450);
        this.surface.e(31, 450, 35, 27785, 275, 0xFFFFFF);
        int ry = 50;
        this.surface.a(256, STRINGS[408], 0xFFFFFF, 0, 1, ry);          // title
        ry += 15;
        this.surface.a(256, STRINGS[411], 0xFFFFFF, param + 28949, 1, ry);
        ry += 15;
        this.surface.a(256, STRINGS[395], 0xFF8000, 0, 1, ry);          // orange warning
        ry += 15;
        ry += 10;
        this.surface.a(256, STRINGS[406], 0xFFFF00, 0, 1, ry);          // category header
        ry += 15;
        this.surface.a(256, STRINGS[407], 0xFFFF00, 0, 1, ry);
        ry += 18;
        this.surface.a(106, STRINGS[410], 0xFF0000, 0, 4, ry);          // column headers
        this.surface.a(256, STRINGS[415], 0xFF0000, 0, 4, ry);
        this.surface.a(406, STRINGS[403], 0xFF0000, param ^ -28949, 4, ry);
        ry += 18;

        // column selection-highlight boxes (rows of varying height) + rule labels
        if (this.Yb == 1)  this.surface.a(36,  (byte) 32,  0x303030, ry - 12, 30, 140);
        this.surface.e(36, 140, ry - 12, param ^ -7582, 30, 0x404040);
        if (this.Yb == 7)  this.surface.a(186, (byte) -106, 0x303030, ry - 12, 30, 140);
        this.surface.e(186, 140, ry - 12, 27785, 30, 0x404040);
        if (this.Yb == 12) this.surface.a(336, (byte) -99, 0x303030, ry - 12, 30, 140);
        this.surface.e(336, 140, ry - 12, 27785, 30, 0x404040);
        this.surface.a(106, STRINGS[414], this.Yb == 1  ? 0xFF8000 : 0xFFFFFF, 0, 0, ry);
        this.surface.a(256, STRINGS[401], this.Yb == 7  ? 0xFF8000 : 0xFFFFFF, param ^ param, 0, ry);
        this.surface.a(406, STRINGS[393], this.Yb == 12 ? 0xFF8000 : 0xFFFFFF, 0, 0, ry);
        ry += 12;
        this.surface.a(106, STRINGS[413], this.Yb == 1  ? 0xFF8000 : 0xFFFFFF, 0, 0, ry);
        this.surface.a(256, STRINGS[396], this.Yb == 7  ? 0xFF8000 : 0xFFFFFF, param ^ -28949, 0, ry);
        this.surface.a(406, STRINGS[412], this.Yb == 12 ? 0xFF8000 : 0xFFFFFF, 0, 0, ry);
        ry += 20;

        if (this.Yb == 2)  this.surface.a(36,  (byte) -111, 0x303030, ry - 12, 18, 140);
        this.surface.e(36, 140, ry - 12, param + 56734, 18, 0x404040);
        if (this.Yb == 8)  this.surface.a(186, (byte) -107, 0x303030, ry - 12, 18, 140);
        this.surface.e(186, 140, ry - 12, 27785, 18, 0x404040);
        if (this.Yb == 13) this.surface.a(336, (byte) -119, 0x303030, ry - 12, 18, 140);
        this.surface.e(336, 140, ry - 12, 27785, 18, 0x404040);
        this.surface.a(106, STRINGS[392], this.Yb == 2  ? 0xFF8000 : 0xFFFFFF, 0, 0, ry);
        this.surface.a(256, STRINGS[399], this.Yb == 8  ? 0xFF8000 : 0xFFFFFF, 0, 0, ry);
        this.surface.a(406, STRINGS[412], this.Yb == 13 ? 0xFF8000 : 0xFFFFFF, 0, 0, ry);
        ry += 20;

        if (this.Yb == 3)  this.surface.a(36,  (byte) -114, 0x303030, ry - 12, 18, 140);
        this.surface.e(36, 140, ry - 12, 27785, 18, 0x404040);
        if (this.Yb == 9)  this.surface.a(186, (byte) -127, 0x303030, ry - 12, 18, 140);
        this.surface.e(186, 140, ry - 12, 27785, 18, 0x404040);
        if (this.Yb == 14) this.surface.a(336, (byte) -117, 0x303030, ry - 12, 18, 140);
        this.surface.e(336, 140, ry - 12, 27785, 18, 0x404040);
        this.surface.a(106, STRINGS[409], this.Yb == 3  ? 0xFF8000 : 0xFFFFFF, 0, 0, ry);
        this.surface.a(256, STRINGS[416], this.Yb == 9  ? 0xFF8000 : 0xFFFFFF, param + 28949, 0, ry);
        this.surface.a(406, STRINGS[402], this.Yb == 14 ? 0xFF8000 : 0xFFFFFF, param + 28949, 0, ry);
        ry += 20;

        if (this.Yb == 4)  this.surface.a(36,  (byte) 118,  0x303030, ry - 12, 18, 140);
        this.surface.e(36, 140, ry - 12, 27785, 18, 0x404040);
        if (this.Yb == 10) this.surface.a(186, (byte) -104, 0x303030, ry - 12, 18, 140);
        this.surface.e(186, 140, ry - 12, 27785, 18, 0x404040);
        this.surface.a(106, STRINGS[404], this.Yb == 4  ? 0xFF8000 : 0xFFFFFF, 0, 0, ry);
        this.surface.a(256, STRINGS[397], this.Yb == 10 ? 0xFF8000 : 0xFFFFFF, 0, 0, ry);
        ry += 20;

        if (this.Yb == 5)  this.surface.a(36,  (byte) 31,  0x303030, ry - 12, 18, 140);
        this.surface.e(36, 140, ry - 12, 27785, 18, 0x404040);
        if (this.Yb == 11) this.surface.a(186, (byte) 62, 0x303030, ry - 12, 18, 140);
        this.surface.e(186, 140, ry - 12, param ^ -7582, 18, 0x404040);
        this.surface.a(106, STRINGS[405], this.Yb == 5  ? 0xFF8000 : 0xFFFFFF, 0, 0, ry);
        this.surface.a(256, STRINGS[417], this.Yb == 11 ? 0xFF8000 : 0xFFFFFF, 0, 0, ry);
        ry += 20;

        if (this.Yb == 6)  this.surface.a(36,  (byte) 82, 0x303030, ry - 12, 18, 140);
        this.surface.e(36, 140, ry - 12, param + 56734, 18, 0x404040);
        this.surface.a(106, STRINGS[398], this.Yb == 6  ? 0xFF8000 : 0xFFFFFF, 0, 0, ry);
        ry += 18;
        ry += 15;

        // "Click here to send report" link — yellow on hover
        int linkCol = 0xFFFFFF;
        if (this.I > 196 && this.I < 316 && this.xb > ry - 15 && this.xb < ry + 5) {
            linkCol = 0xFFFF00;
        }
        this.surface.a(256, STRINGS[391], linkCol, param + 28949, 1, ry);
    }

    // -------------------------------------------------------------------------
    // drawPlayerMenu  — obf: void L(int)
    // -------------------------------------------------------------------------

    /** Render the in-chat right-click "Choose option" menu (also handles friends/PM-history
     *  context labels), and on click route it: in wilderness combat it sends opcode 59
     *  (PLAYER_ATTACK), otherwise it runs the click via updateCamera (b(false,0)).
     *
     *  FIX vs old: the menu list is the friendsList MessageList (zh), not "chatList"; chat-
     *  history names come from Globals.c (l.c) not "GlobalStrings"; the `af>=0 || Bh>=0`
     *  guards had the inequality flipped in the old version. */
    private final void drawPlayerMenu(int param) {
        if (this.af >= 0 || this.Bh >= 0) {     // FIX: was `<= 0`
            this.friendsList.a(4000, "", STRINGS[121], 30192);   // "Cancel" entry
        }
        this.friendsList.a((byte) 16);
        int count = this.friendsList.c(-27153);
        if (param >= -120) return;   // anti-tamper guard

        // trim the list to at most 20 entries
        for (int i = count; i > 20; i--) {
            this.friendsList.b(102, i - 1);
        }

        // friends-tab / pm-history context label
        if (this.qc == 5) {   // ~qc == -6
            String label = null;
            if (this.pk == 0 && this.wk != 0) {
                if (this.wk >= 0) {
                    int idx = this.wk;
                    String suffix = "";
                    if ((Fj[idx] & 4) == 0) {
                        label = ua.h[idx];
                        suffix = STRINGS[190];               // " - online"
                    } else {
                        label = STRINGS[188] + ua.h[idx];    // "Message "
                        if (ac.z[idx] != null) suffix = STRINGS[193] + ac.z[idx];
                    }
                    if (cb.c[idx] != null && cb.c[idx].length() > 0) {
                        label = label + STRINGS[198] + cb.c[idx] + ")" + suffix;
                    } else {
                        label = label + suffix;
                    }
                } else {
                    int idx = -(2 + this.wk);
                    label = STRINGS[196] + ua.h[idx];
                    if (cb.c[idx] != null && cb.c[idx].length() > 0) {
                        label = label + STRINGS[198] + cb.c[idx] + ")";
                    }
                }
            }
            if (this.pk == 1 && this.nj != 0) {
                if (this.nj >= 0) {
                    int idx = this.nj;
                    label = STRINGS[194] + l.c[idx];
                    if (ia.g[idx] != null && ia.g[idx].length() > 0) {
                        label = label + STRINGS[198] + ia.g[idx] + ")";
                    }
                } else {
                    int idx = -(2 + this.nj);
                    label = STRINGS[196] + l.c[idx];
                    if (ia.g[idx] != null && ia.g[idx].length() > 0) {
                        label = label + STRINGS[198] + ia.g[idx] + ")";
                    }
                }
            }
            if (label != null) {
                this.surface.a(label, 6, 14, 0xFFFF00, false, 1);
            }
        }

        count = this.friendsList.c(-27153);
        if (count <= 0) return;

        // find the last non-empty entry
        int lastNonEmpty = -1;
        for (int i = 0; i < count; i++) {
            String entry = this.friendsList.b((byte) 74, i);
            if (entry != null && entry.length() > 0) lastNonEmpty = i;
        }

        // compose the menu header
        String header = null;
        if ((this.Bh >= 0 || this.af >= 0) && count == 2) {
            header = STRINGS[192];   // "Choose option"
        } else if ((this.Bh <= 0 || this.af <= 0) && count > 1) {
            header = STRINGS[15] + this.friendsList.b(0, (byte) 53) + " " + this.friendsList.b((byte) 75, 0);
        } else if (lastNonEmpty != -1) {
            header = this.friendsList.b((byte) 54, lastNonEmpty) + STRINGS[159] + this.friendsList.b(0, (byte) 53);
        }
        if (count == 2 && header != null) header = header + STRINGS[189];
        if (count > 3 && header != null) header = header + STRINGS[195] + (count - 1) + STRINGS[191];
        if (header != null) this.surface.a(header, 6, 14, 0xFFFF00, false, 1);

        // position the popup near the cursor when it pops (single-button vs two-button modes)
        boolean popMenu = (!this.Yh && this.Cf == 1) || (this.Yh && this.Cf == 2 && count == 1);
        if (popMenu || (!this.Yh && this.Cf == 2) || (this.Yh && this.Cf == 1)) {
            if (!popMenu && (this.Yh ? this.Cf != 1 : this.Cf != 2)) {
                return;
            }
            int w = this.friendsList.b(16256);
            int hgt = this.friendsList.a(-21224);
            this.rh = this.I - w / 2;
            this.se = true;
            this.fg = this.xb - 7;
            if (this.rh < 0) this.rh = 0;
            if (this.fg < 0) this.fg = 0;
            this.Cf = 0;
            if (this.fg + hgt > 316) this.fg = 315 - hgt;
            if (w + this.rh > 510) this.rh = 510 - w;
            return;
        }

        // confirm: send attack if in wilderness combat, else dispatch the click
        if (this.bb && this.gb && this.Hc) {
            this.clientStream.b(59, 0);          // PLAYER_ATTACK
            this.clientStream.f.e(393, this.rf);
            this.clientStream.f.e(393, this.Cg);
            this.clientStream.b(21294);
        } else {
            this.updateCamera(false, 0);          // b(boolean,int)
        }
        this.Cf = 0;
    }

    // -------------------------------------------------------------------------
    // drawOptionsTab  — obf: void f(boolean)
    // -------------------------------------------------------------------------

    /** Load the "Configuration" options archive and apply it. Plays the menu sound on open. */
    private final void drawOptionsTab(boolean playSfx) {
        if (playSfx) {
            this.playSound((byte) 77, null);
        }
        byte[] data = this.a(STRINGS[225], 10, 0, 78);
        if (data != null) {
            m.a(data, (byte) 100, this.Pg);   // m = SocketFactory: apply options
        } else {
            this.Vc = true;
        }
    }

    // -------------------------------------------------------------------------
    // drawChatHistoryTabs  — obf: void A(int)
    // -------------------------------------------------------------------------

    /** Render the chat / quest / private / friends history tab strip + the settings icon at the
     *  bottom of the screen. Tabs blink red when they have unread activity (Ee/Qe/Vj/Mh timers).
     *  Only renders when param == 5. */
    private final void drawChatHistoryTabs(int param) {
        this.surface.b(-1, this.tg + 23, this.Oi - 4, 0);   // tab-strip background sprite
        if (param != 5) return;

        // tab 0 — Chat (active if Zh==0; blink if Ee%30 > 15)
        int col = o.a(200, 9570, 255, 200);
        if (this.Zh == 0) col = o.a(255, 9570, 50, 200);
        if (this.Ee % 30 > 15) col = o.a(255, 9570, 50, 50);
        this.surface.a(54, STRINGS[269], col, 0, 0, this.Oi + 6);

        // tab 1 — Quest (active if Zh==1; blink if Qe%30 > 15)
        col = o.a(200, 9570, 255, 200);
        if (this.Zh == 1) col = o.a(255, param + 9565, 50, 200);
        if (this.Qe % 30 > 15) col = o.a(255, param ^ 0x2567, 50, 50);
        this.surface.a(155, STRINGS[272], col, 0, 0, this.Oi + 6);

        // tab 2 — Private (active if Zh==2; blink if Vj%30 > 15)
        col = o.a(200, 9570, 255, 200);
        if (this.Zh == 2) col = o.a(255, 9570, 50, 200);
        if (this.Vj % 30 > 15) col = o.a(255, param + 9565, 50, 50);
        this.surface.a(255, STRINGS[271], col, 0, 0, this.Oi + 6);

        // tab 3 — Friends (active if Zh==3; blink if Mh%30 > 15)
        col = o.a(200, 9570, 255, 200);
        if (this.Zh == 3) col = o.a(255, 9570, 50, 200);
        if (this.Mh % 30 > 15) col = o.a(255, param ^ 0x2567, 50, 50);
        this.surface.a(355, STRINGS[268], col, 0, 0, this.Oi + 6);

        // settings/report icon
        this.surface.a(457, STRINGS[120], 0xFFFFFF, 0, 0, this.Oi + 6);
    }

    // -------------------------------------------------------------------------
    // drawChat  — obf: void l(int)
    // -------------------------------------------------------------------------

    /** Draw the floating in-world overlays after the 3D scene: hit-damage splats (Kc text at
     *  tf/ee with width/height nf/uf, de-overlapped vertically), ground-item sprites
     *  (je/pe/jd/ak), and entity health bars (gd/Pk/bf).
     *
     *  FIX vs old: the old version threw UnsupportedOperationException with an (incorrect) claim
     *  that this is the chat-scrollback panel. It is NOT — l(int) is the post-render overlay
     *  pass. Reconstructed in full from the clean source. */
    private final void drawChat(int param) {
        // --- damage splats: nudge each up so it doesn't overlap an earlier one ---
        int gap = this.surface.a(508305352, 1);   // line height
        for (int i = 0; i < this.Ef; i++) {
            int sx = this.tf[i];
            int sy = this.ee[i];
            int sw = this.nf[i];
            int sh = this.uf[i];
            boolean moved = true;
            while (moved) {
                moved = false;
                for (int j = 0; j < i; j++) {
                    if (sy + sh > this.ee[j] - gap
                            && sy - gap < this.ee[j] + this.uf[j]
                            && sx - sw < this.tf[j] + this.nf[j]
                            && sw + sx > this.tf[j] - this.nf[j]
                            && sy > this.ee[j] - gap - sh) {
                        sy = this.ee[j] - (gap + sh);
                        moved = true;
                    }
                }
            }
            this.ee[i] = sy;
            this.surface.a(300, this.Kc[i], sx, 55, 1, sy, false, 0xFFFF00);
        }

        // --- ground-item icons ---
        for (int i = 0; i < this.jc; i++) {
            int gx = this.je[i];
            int gy = this.pe[i];
            int scale = this.jd[i];
            int itemId = this.ak[i];
            int bw = 39 * scale / 100;
            int bh = 27 * scale / 100;
            int boxY = gy - bh;
            this.surface.a(this.tg + 9, (byte) -122, bh, gx - bw / 2, bw, boxY, 85);
            int iw = scale * 36 / 100;
            int ih = 24 * scale / 100;
            this.surface.a(boxY + bh / 2 - ih / 2, h.c[itemId], 0, false, 0,
                ua.Bb[itemId] + this.sg, ih, iw, gx - iw / 2, 1);
        }

        // --- entity health bars (green = remaining, red = lost) ---
        for (int i = 0; i < this.Bc; i++) {
            int hx = this.gd[i];
            int hy = this.Pk[i];
            int pct = this.bf[i];                 // 0..30 = green width
            this.surface.c(192, hx - 15, 5, 0, hy - 3, pct, 0x00FF00);
            this.surface.c(192, pct - 15 + hx, 5, 0, hy - 3, 30 - pct, 0xFF0000);
        }
    }

    // -------------------------------------------------------------------------
    // drawWelcome  — obf: void j(int)
    // -------------------------------------------------------------------------

    /** "Welcome to RuneScape" box on login: last-login, recovery-questions reminder, unread
     *  messages, subscription/members status, and a "Click here to play" dismiss button.
     *  Sb = subscription-days marker (201 = none); ce = recovery set time; id = unread count. */
    private final void drawWelcome(int param) {
        int h = 65;
        if (this.Sb != 201) h += 60;
        if (this.id > 0)    h += 30;
        if (this.ce != 0)   h += 45;

        int top = 167 - h / 2;
        this.surface.a(56, (byte) 77, 0, top, h, 400);
        this.surface.e(56, 400, top, 27785, h, 0xFFFFFF);

        int y = top + 20;
        this.surface.a(256, STRINGS[667] + this.wi.C, 0xFFFF00, 0, 4, y);   // "Welcome <name>"
        y += 30;

        // last-login line
        String last;
        if (this.hi == 0)      last = STRINGS[658];               // "first time"
        else if (this.hi == 1) last = STRINGS[665];               // "yesterday"
        else                   last = this.hi + STRINGS[652];     // "N days ago"

        // recovery-questions reminder block
        if (this.ce != 0) {
            this.surface.a(256, STRINGS[655] + last, 0xFFFFFF, 0, 1, y);
            y += 15;
            if (this.ve == null) {
                this.ve = this.formatNumber(param ^ 0x128B, this.ce);   // c(int,int): date string
            }
            this.surface.a(256, STRINGS[662] + this.ve, 0xFFFFFF, param ^ -4853, 1, y);
            y += 15;
            y += 15;
        }

        // unread messages block
        if (this.id > 0) {
            if (this.id == 1) {
                this.surface.a(256, STRINGS[656], 0xFFFFFF, 0, 1, y);                // "no unread"
            } else {
                this.surface.a(256, STRINGS[668] + (this.id - 1) + STRINGS[661], 0xFFFFFF, param + 4853, 1, y);
            }
            y += 15;
            y += 15;
        }

        // subscription/members status block
        if (this.Sb != 201) {
            if (this.Sb == 200) {   // ~Sb == -201
                this.surface.a(256, STRINGS[660], 0xFF8000, 0, 1, y);
                y += 15;
                this.surface.a(256, STRINGS[657], 0xFF8000, param ^ -4853, 1, y);
                y += 15;
                this.surface.a(256, STRINGS[663], 0xFF8000, 0, 1, y);
                y += 15;
            } else {
                String sub;
                if (this.Sb == 0)      sub = STRINGS[654];
                else if (this.Sb == 1) sub = STRINGS[659];
                else                   sub = this.Sb + STRINGS[652];
                this.surface.a(256, sub + STRINGS[666], 0xFF8000, 0, 1, y);
                y += 15;
                this.surface.a(256, STRINGS[664], 0xFF8000, 0, 1, y);
                y += 15;
                this.surface.a(256, STRINGS[663], 0xFF8000, param + 4853, 1, y);
                y += 15;
            }
            y += 15;
        }

        // "Click here to play" — red on hover
        int colour = 0xFFFFFF;
        if (this.xb > y - 12 && this.xb <= y && this.I > 106 && this.I < 406) {
            colour = 0xFF0000;
        }
        this.surface.a(256, STRINGS[126], colour, param ^ param, 1, y);

        // dismiss on click (on the button, or anywhere outside the panel)
        if (this.Cf == 2) {
            if (colour == 0xFF0000) {
                this.Oh = false;
            }
            if ((this.I < 86 || this.I > 426) && (this.xb < top || this.xb > top + h)) {
                this.Oh = false;
            }
        }
        this.Cf = 0;
    }

    // -------------------------------------------------------------------------
    // playSound  — obf: void a(int,String)
    // -------------------------------------------------------------------------

    /** Play a named .pcm sound effect at full volume (256). Looks the sample up in the Uh sound
     *  archive by name+".pcm" (STRINGS[515]). Muted while sleeping (ne). */
    private final void playSound(int param, String name) {
        if (this.soundMixer == null) return;   // hk
        if (this.ne) return;                     // sleeping → no sfx
        if (param >= -43) return;                // anti-tamper guard

        int offset = oa.a(name + STRINGS[515], (byte) 68, this.Uh);   // NameHash.a → byte offset
        int length = client.a(this.Uh, name + STRINGS[515], -125);     // a(byte[],String,int) → length
        if (length == 0) return;                 // not found (~len == -1 idiom)

        vb sample = new vb(8000, v.a(this.Uh, length, -98, offset), 0, length); // SampleBuffer / ChatCipher.a
        this.soundMixer.a(sample, 100, 256);
    }

    // -------------------------------------------------------------------------
    // initSounds  — obf: void E(int)
    // -------------------------------------------------------------------------

    /** Bring up the audio engine: load the "Sounds" archive (Uh), open an AudioChannel (ni) at
     *  22050 Hz attached to the applet host component, and wire in a StreamMixer (hk).
     *
     *  FIX vs old: the AudioChannel.a host arg is ImageLoader.k (pa.k), not "Timer.k". */
    private final void initSounds(int param) {
        if (param > -55) return;   // anti-tamper guard

        this.Uh = this.a(STRINGS[345], 90, 10, 66);   // "Sounds" archive
        try {
            sa.a(22050, false, 1);   // AudioChannel static init

            Object host;
            if (kb.a != null) {           // InputState.a (applet host) takes priority
                host = kb.a;
            } else if (da.gb != null) {   // ClientStream.gb fallback
                host = da.gb;
            } else {
                host = this;
            }

            this.ni = sa.a(pa.k, (Component) host, 0, 22050);   // FIX: ImageLoader.k, not Timer.k
            this.hk = new ra();          // StreamMixer
            this.ni.a(this.hk);
        } catch (Throwable t) {
            System.out.println(STRINGS[344] + t);   // "Unable to init sounds: "
        }
    }
