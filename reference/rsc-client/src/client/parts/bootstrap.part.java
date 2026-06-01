// ===== bootstrap =====
// Methods in group "bootstrap" RE-AUDITED against the CLEAN Vineflower base
// (decompiled/normalized-clean/client.java). Microsoft J++ RSC rev ~233-235.
//
// Obfuscation stripped:
//   - Opaque predicate (client.OPAQUE_FALSE / vh) — always false; dead branches removed.
//   - Per-method profiling counter increments (++Pd; ++Gd; ...).
//   - try/catch(RuntimeException){throw ErrorHandler.a(e, il[label])} wrappers unwrapped.
//   - Anti-tamper dummy-param guards (if(p!=const)...) + junk modulo expressions removed.
//   - The ~x>~y / ~x==const sign idioms rewritten to ordinary comparisons.
//
// Naming honours docs/NAMING.md (CORRECTED): k=World, lb=Scene.
//   Therefore Ek (type lb=Scene) -> `scene`, Hh (type k=World) -> `world`.
//   (The MUDCLIENT_SKELETON field table had these two swapped; NAMING.md wins.)
//
// All il[] string indices below were decoded from the real XOR pool and are
// annotated with their true decoded text. The OLD part file's il[] comments
// were taken from the defective decompilation and were almost entirely wrong.
//
// Class context: package client; class Mudclient extends GameShell (e)
// STRINGS = client.il  (XOR-decoded string pool)

    // -------------------------------------------------------------------------
    /** Standalone entry point: parse args, select audio backend, create and start client frame.
     *  obf: static void main(String[])   obf-label: il[313]="client.main(" */
    public static final void main(String[] args) {
        try {
            // args[0] = nodeid (int), args[1] = mode string, args[2..] = optional flags.

            // GameShell.audioQueue (e.i) <- la.b (the static BZip2 reference).
            GameShell.audioQueue = BZip.instance;            // e.i = la.b

            // Parse the node/world id.
            BZip.nodeId = Integer.parseInt(args[0]);         // aa.l

            // Select audio/queue backend from the mode string.
            //   il[312]="live"    -> AudioMixer            (eb.e)
            //   il[317]="rc"      -> RecordLoader          (f.b)   [see note]
            //   il[318]="wip"     -> SurfaceImageProducer  (fb.h)
            // Control flow (from CFR) falls through: matching "wip" sets f.b then
            // falls into fb.h then eb.e; matching "rc" sets fb.h then eb.e;
            // matching "live" sets only eb.e. Net effect of the fall-through chain:
            //   "live"          -> db.f = eb.e
            //   "rc"  (il[317]) -> db.f = fb.h
            //   "wip" (il[318]) -> db.f = f.b
            if (args[1].equals(STRINGS[312])) {              // "live"
                LinkedQueue.audioFactory = AudioMixer.instance;            // db.f = eb.e
            } else if (args[1].equals(STRINGS[317])) {       // "rc"
                LinkedQueue.audioFactory = SurfaceImageProducer.audioInstance; // db.f = fb.h
            } else if (args[1].equals(STRINGS[318])) {       // "wip"
                LinkedQueue.audioFactory = RecordLoader.audioInstance;     // db.f = f.b
            }
            // else: leave db.f unchanged.

            // Construct the main client instance (applet flag false for standalone).
            Mudclient client = new Mudclient();
            client.isApplet = false;                         // hj = false

            // Parse optional flags from args[2..].
            for (int i = 2; i < args.length; i++) {
                if (args[i].equals(STRINGS[316])) {          // "members"
                    client.isMembersWorld = true;            // Pg
                }
                if (args[i].equals(STRINGS[315])) {          // "veterans"
                    client.isFreeWorld = true;               // cf
                }
            }

            // Create the AWT frame and start the game thread.
            //   il[314]="local.runescape.com"  (title/host param A)
            //   il[319]="classic"              (title/host param B)
            //   32 + db.f.a = frame/thread priority offset
            //   aa.l + 7000 = server port; (byte)112 flags; fa.d = display depth
            //   client.Wd = 512 (width); client.Oi - -12 = 334 + 12 = 346 (height)
            try {
                client.a(
                    false,                                       // standalone (not applet)
                    STRINGS[314],                                // "local.runescape.com"
                    32 + LinkedQueue.audioFactory.threadPriority,// 32 + db.f.a
                    STRINGS[319],                                // "classic"
                    BZip.nodeId + 7000,                          // aa.l + 7000 (port)
                    (byte)112,
                    ClientIOException.displayDepth,              // fa.d
                    client.screenWidth,                          // Wd = 512
                    client.screenHeight + 12                     // Oi - -12 = 346
                );
                client.ticksPerFrame = 10;                       // Q = 10
            } catch (Exception e) {
                Utility.reportError(0x1FFFFF, e, null);           // mb.a(2097151, e, null)
            }
        } catch (RuntimeException e) {
            // obf: throw i.a(e, il[313] + (args!=null ? il[29] : il[31]) + ')')
            throw ErrorHandler.a(e, STRINGS[313] + (args != null ? STRINGS[29] : STRINGS[31]) + ')');
        }
    }

    // -------------------------------------------------------------------------
    /** Applet init: read nodeid/modewhat/modewhere params, size window, kick off loading.
     *  obf: void init()   obf-label: il[183]="client.init()" */
    @Override
    public final void init() {
        try {
            // il[182]="nodeid" -> BZip.nodeId (aa.l)
            BZip.nodeId = Integer.parseInt(this.getParameter(STRINGS[182]));

            // il[185]="modewhere" here doubles as the font-size param; ub.a(size,(byte)24)
            // builds a NameTable font reference into GameShell.audioQueue? No — e.i (fontMetrics).
            GameShell.fontMetrics = NameTable.buildFont(
                Integer.parseInt(this.getParameter(STRINGS[185])), (byte)24
            );                                               // e.i = ub.a(...)
            if (GameShell.fontMetrics == null) {
                GameShell.fontMetrics = Surface.defaultFont; // e.i = ua.E
            }

            // il[184]="modewhat" -> u.a(false,id) picks the LinkedQueue/audio factory.
            LinkedQueue.audioFactory = StringCodec.buildQueue(
                false, Integer.parseInt(this.getParameter(STRINGS[184]))
            );                                               // db.f = u.a(false, ...)
            if (LinkedQueue.audioFactory == null) {
                LinkedQueue.audioFactory = AudioMixer.instance; // db.f = eb.e
            }

            // Start the GameShell loader thread:
            //   super.a(Oi+12, fa.d, db.f.a+32, 2, Wd)
            super.startLoaderThread(
                this.screenHeight + 12,                          // Oi + 12
                ClientIOException.displayDepth,                  // fa.d
                LinkedQueue.audioFactory.threadPriority + 32,    // db.f.a + 32
                2,
                this.screenWidth                                 // Wd
            );
        } catch (RuntimeException e) {
            throw ErrorHandler.a(e, STRINGS[183]);           // il[183]="client.init()"
        }
    }

    // -------------------------------------------------------------------------
    /** Constructor: allocate all state arrays and set field defaults.
     *  obf: client()  (no catch wrapper) */
    public Mudclient() {
        super();

        // --- Network ---
        incomingPacket = new BitBuffer(5000);    // mg

        // --- Basic scalars / cursors ---
        Nc = 0;
        Vg = 0;
        qd = 9;                                  // camera zoom default
        mouseClickTimes = new long[100];         // Zd
        Wc = 0;
        Oj = 0;
        loginAnimFrame = 0;                      // jk
        loginTimeout = 550;                      // ac
        isApplet = true;                         // hj
        yj = -1;
        De = 0;
        npcCount = 0;                            // If
        Xh = false;                              // domain-lock tripped flag
        si = 1;
        xh = 0;
        Ce = 0;
        cameraRotationY = 0;                     // oc (idle drift, cameraRotationY)
        bl = -1;
        qk = 0;
        Sg = -1;
        dc = 0;
        ug = 128;                                // cameraRotation seed
        Ug = 128;                                // cameraRotation seed (2nd axis)
        yg = -1;
        Kk = new int[8192];
        pf = new int[8000];
        isMembersWorld = false;                  // Pg
        Cf = 0;
        loginStage = 0;                          // Zb
        screenWidth = 512;                       // Wd
        kg = 0;
        bc = -1;
        screenHeight = 334;                      // Oi
        mouseButtonMode = 2;                     // eg: dual-purpose obf field — default 2.
                                                 //   In startGame it doubles as the cameraRotationX
                                                 //   idle-drift increment; setMouseButtonMode sets it to 77.
        rc = 0;
        qe = 0;
        isFreeWorld = false;                     // cf
        nk = 0;
        Yc = 0;
        outOfMemory = false;                     // Ue
        Si = 0;
        Ki = 0;
        fatalLoadError = false;                  // Vc
        pj = 0;
        uj = new int[8192];
        sk = 0;
        screenMode = 0;                          // qg: 0=login, 1=in-game
        zf = false;
        fpsCapBackground = 40;                   // nc
        cameraRotationYIncrement = 2;            // Ok (idle drift step for cameraRotationY)
        cameraRotationTime = 0;                  // oj
        Fd = 0;
        ui = 0;
        Ag = 0;

        // --- Entities in view ---
        players = new GameCharacter[500];        // Zg
        cameraRotationX = 0;                     // Be (idle drift, cameraRotationX)
        npcCountView = 0;                        // Mg
        playersLast = new GameCharacter[500];    // rg
        npcsCache = new GameCharacter[4000];     // We
        tj = 0;
        Rg = new int[8000];
        worldIndex = 0;                          // Vh
        localPlayer = new GameCharacter();       // wi
        bg = new int[1500];
        ci = new int[256];
        Cd = false;
        pmTarget = null;                         // Zj
        Rj = new int[256];
        dj = 0;

        // --- Chat colour palette (15 RGB entries) — verified against clean ei[] ---
        chatColors = new int[]{
            0xFF0000,  // red
            0xFF8000,  // orange
            0xFFE000,  // yellow
            0xA0E000,  // yellow-green
            0x00E000,  // green
            0x008000,  // dark green
            0x00A080,  // teal
            0x00B0FF,  // sky blue
            0x0080FF,  // blue
            0x0030F0,  // blue-violet (12528)
            0xE000E0,  // magenta
            0x303030,  // dark grey
            0x604000,  // brown
            0x805000,  // tan
            0xFFFFFF   // white
        };

        // --- Entity/object index arrays ---
        Zf = new int[5000];
        oe = new int[50];
        xk = 0;
        Uf = new int[8];
        Qe = 0;
        ae = new int[256];
        lh = false;
        skillXp = new int[14];                   // zj
        el = 0;
        Ui = 0;
        tf = new int[50];
        Bc = 0;
        zd = 0;
        Fg = 0;
        hi = 0;
        Zc = -1;

        // --- Item colour palette (10 RGB entries) — verified against clean Dg[] ---
        itemColors = new int[]{
            0xFFC030, 0xFFA040, 0x805030, 0x604020, 0x303030,
            0xFF6020, 0xFF4000, 0xFFFFFF, 0x00FF00, 0x00FFFF
        };

        oh = new int[18];                        // equipment bonus stats

        // --- Fatigue/sleep-bar colours (5 entries) — verified against clean Wh[] ---
        fatigueColors = new int[]{
            0xECDED0, 0xCCB366, 0xB38C40, 0x997326, 0x906020
        };

        ce = 0;
        objectModels = new GameModel[1000];      // kh
        Xe = new int[256];
        Pf = 0;
        Mi = false;

        // --- Head-slot draw-layer order (8 styles) ---
        Og = new int[]{0, 0, 0, 0, 0, 1, 2, 1};

        skillBase = new int[14];                 // Vb

        // --- Equipment-slot layer ordering tables (8 configs x 12 slots) ---
        Tg = new int[][] {
            {11, 2, 9, 7, 1, 6, 10, 0, 5, 8, 3, 4},
            {11, 2, 9, 7, 1, 6, 10, 0, 5, 8, 3, 4},
            {11, 3, 2, 9, 7, 1, 6, 10, 0, 5, 8, 4},
            {3, 4, 2, 9, 7, 1, 6, 10, 8, 11, 0, 5},
            {3, 4, 2, 9, 7, 1, 6, 10, 8, 11, 0, 5},
            {4, 3, 2, 9, 7, 1, 6, 10, 8, 11, 0, 5},
            {11, 4, 2, 9, 7, 1, 6, 10, 0, 5, 8, 3},
            {11, 2, 9, 7, 1, 6, 10, 0, 5, 8, 4, 3}
        };

        jd = new int[50];
        experienceTable = new int[99];           // ti (filled in loadGameConfig)
        Ng = new int[500];
        wg = 2;
        skillCurrent = new int[14];              // Me
        wj = 0;
        nf = new int[50];
        Rd = -1;
        Kd = false;
        Wg = 8;
        th = new int[8];
        zi = 0;

        // --- Skill names (short set) — il[48..580] ---
        skillNamesShort = new String[] {         // Vk
            STRINGS[48],  STRINGS[543], STRINGS[546], STRINGS[562], // Attack, Defense, Strength, Hits
            STRINGS[575], STRINGS[570], STRINGS[16],  STRINGS[548], // Ranged, Prayer, Magic, Cooking
            STRINGS[557], STRINGS[591], STRINGS[565], STRINGS[550], // Woodcut, Fletching, Fishing, Firemaking
            STRINGS[559], STRINGS[569], STRINGS[560], STRINGS[529], // Crafting, Smithing, Mining, Herblaw
            STRINGS[567], STRINGS[580]                              // Agility, Thieving
        };

        chatInputLine = "";                      // Lg
        Ee = 0;
        ke = false;
        Gi = 48;
        bf = new int[50];
        ff = false;
        Sc = new int[50];
        gc = 0;
        fg = 0;                                  // fg (distinct field; NOT screenMode)
        weaponBonuses = new int[18];             // cg
        Mh = 0;
        ee = new int[50];
        fd = false;
        gi = new int[50];
        playersCache = new GameCharacter[5000];  // te
        isMembersAccount = true;                 // Bd

        // --- Character animation slot order (8 entries) ---
        Pc = new int[]{0, 1, 2, 1, 0, 0, 0, 0};

        Ef = 0;
        Hk = false;
        wallModels = new GameModel[1500];        // hg
        Ve = 0;
        vj = 0;
        Zh = 0;
        chatInputUser = "";                      // cj
        ve = null;
        id = 0;

        // --- Scroll-frame step table ---
        sf = new int[]{0, 1, 2, 1};

        passwordInput = "";                      // wh
        mf = 0;

        // --- Skill names (long set) — same pool, idx 553 ("Woodcutting") vs short's 557 ---
        skillNamesLong = new String[] {          // Ej
            STRINGS[48],  STRINGS[543], STRINGS[546], STRINGS[562],
            STRINGS[575], STRINGS[570], STRINGS[16],  STRINGS[548],
            STRINGS[553], STRINGS[591], STRINGS[565], STRINGS[550],
            STRINGS[559], STRINGS[569], STRINGS[560], STRINGS[529],
            STRINGS[567], STRINGS[580]
        };

        vi = new int[256];
        ii = 0;
        vd = false;
        di = new int[256];
        Yh = false;
        od = null;
        de = 0;
        npcModelCache = new GameModel[500];      // rd
        Yk = true;
        qj = 0;
        Vf = 0;
        pe = new int[50];
        Ni = new int[5000];
        Le = new int[5000];
        gl = 0;
        Fc = new int[5];
        tradeOwnItems = new int[8];              // xi
        Gj = new int[5000];
        se = false;
        Wk = false;
        sessionBytes = null;                     // Uh
        je = new int[50];
        af = -1;
        skillXpGained = new int[14];             // Qf
        Ph = false;
        Jf = new int[256];
        objectTileX = new int[1500];             // ye
        Oc = new int[50];                        // menu/option scratch array
        questCompleteFlags = new boolean[500];   // Sj
        npcs = new GameCharacter[500];           // Ff
        Ji = 0;
        Lk = 0;
        bj = 0;
        Ed = new boolean[1500];
        serverMessage = "";                      // Cj
        eh = 0;
        Dc = false;
        bk = new boolean[50];
        Yb = 0;

        // --- Quest names (50 entries, il[529..598]) ---
        questNames = new String[] {              // Te
            STRINGS[596], STRINGS[542], STRINGS[554], STRINGS[598],
            STRINGS[586], STRINGS[573], STRINGS[584], STRINGS[590],
            STRINGS[541], STRINGS[551], STRINGS[545], STRINGS[535],
            STRINGS[561], STRINGS[547], STRINGS[566], STRINGS[589],
            STRINGS[568], STRINGS[540], STRINGS[555], STRINGS[594],
            STRINGS[595], STRINGS[583], STRINGS[536], STRINGS[588],
            STRINGS[579], STRINGS[544], STRINGS[587], STRINGS[578],
            STRINGS[564], STRINGS[534], STRINGS[585], STRINGS[572],
            STRINGS[556], STRINGS[577], STRINGS[576], STRINGS[538],
            STRINGS[582], STRINGS[531], STRINGS[539], STRINGS[563],
            STRINGS[593], STRINGS[537], STRINGS[533], STRINGS[549],
            STRINGS[558], STRINGS[574], STRINGS[592], STRINGS[530],
            STRINGS[597], STRINGS[532]
        };

        Ti = 0;
        Pj = false;
        rk = 0;
        wk = -1;
        ue = false;
        Id = 0;
        Vd = 0;
        fpsCapForeground = 30;                   // cl
        serverUpdateTick = 1;                    // Sf
        tradeTheirItems = new int[8];            // zc
        xg = 0;
        hf = 0;
        Nh = 0;
        chatEntry = "";                          // ec
        Kh = true;
        Kg = false;
        skillCurrentLevels = new int[14];        // Bi
        armorBonuses = new int[18];              // Ak
        fi = new boolean[50];
        Di = -1;
        uf = new int[50];
        md = false;
        combatStyleIndex = 14;                   // Lh
        Jd = new int[500];
        ai = 0;
        ki = false;
        fj = 0;
        duelOwnItems = new int[8];               // xj
        yk = new int[500];
        Sb = 0;
        menuOptionStrings = new String[50];      // Kc
        Hj = new int[500];
        menuOptionActions = new int[50];         // gd
        inventoryItems = new int[35];            // xe
        vk = false;
        Hc = false;
        skillBaseLevels = new int[14];           // Lc

        // --- Equipment-bonus stat labels (NOT combat-style names) ---
        //     il[552]="Armour", il[571]="WeaponAim", il[581]="WeaponPower",
        //     il[16]="Magic", il[570]="Prayer"
        combatStyleNames = new String[] {        // Ld
            STRINGS[552], STRINGS[571], STRINGS[581], STRINGS[16], STRINGS[570]
        };

        uk = false;
        Bj = 0;
        nj = -1;
        ne = false;
        username = "";                           // Xf
        bankOfferItems = new int[8];             // of
        Oh = false;
        duelTheirItems = new int[8];             // df
        fl = 0;
        ignoreListEntry = "";                    // ig
        qc = 0;
        pk = 0;
        rh = 0;
        Df = 0;
        Pk = new int[50];
        Vi = false;
        skillXpAccum = new int[14];              // jj
        Ah = 0;
        uc = 0;
        le = 0;
        dk = 1;
        bankSlotItems = new int[8];              // kf
        inventoryEquipped = new int[35];         // Aj
        Vj = 0;
        hh = 0;
        objectTileZ = new int[1500];             // vc
        fh = -2;
        Nj = 0;
        sd = 0;
        objectTileY = new int[1500];             // Se
        Yi = false;
        ld = 2;
        kc = 0;
        nh = 0;
        Xd = 0;
        Yd = 0;
        mh = false;
        Qk = false;
        Td = false;
        vg = 0;
        pg = 0;
        dd = false;
        sj = -2;
        Fe = false;
        Xj = false;
        Bh = -1;
        skillXpDeltas = new int[14];             // Dd
        Je = false;
        Ke = 0;
        Tk = 0;
        jc = 0;
        inventoryCount = new int[35];            // vf
        menuOptionTargets = new int[50];         // ak
        Ub = false;
        ah = new String[5];
        npcsLast = new GameCharacter[500];       // Tb
    }

    // -------------------------------------------------------------------------
    /** GameShell hook: per-login-screen tick — drive login/sleep screens + idle camera drift.
     *  Called from GameShell.run.   obf: void e(int)   obf-label: il[227]="client.MA(" */
    @Override
    final void startGame(int frameTick) {
        try {
            // Skip if domain-lock tripped or fatal load error already set.
            if (Xh) {
                return;
            }
            if (outOfMemory) {
                return;
            }

            // Clear the right-click menu/option buffer when not a forced tick (<64).
            if (frameTick < 64) {
                Oc = null;                       // Oc = (int[])null
            }

            if (fatalLoadError) {                // Vc
                return;
            }

            // Start the active audio voice if present.
            if (soundChannel != null) {          // ni
                soundChannel.startPlayback();    // ni.a()
            }

            try {
                // Advance the login-screen animation frame.
                loginAnimFrame++;                // jk++

                if (screenMode == 0) {           // login
                    loginFrameCount = 0;         // sb = 0
                    drawLoginInput(2);           // x(2)  [login.part: drawLoginInput]
                }

                if (screenMode == 1) {           // in-game
                    loginFrameCount++;           // sb++
                    handleGameInput(0);          // J(0)  [mainloop.part: handleGameInput / tick]
                }

                // Reset per-frame "mouse button seen" flag.
                Qb = 0;

                // Every 500 ticks, randomly nudge the idle camera drift.
                // NOTE: cameraRotationX = Be, cameraRotationY = oc; the per-axis drift
                //   increments are the dual-purpose fields mouseButtonMode (eg) for X and
                //   cameraRotationYIncrement (Ok) for Y. (Oracle: cameraRotationXIncrement,
                //   cameraRotationYIncrement.)
                cameraRotationTime++;            // oj++
                if (cameraRotationTime > 500) {
                    cameraRotationTime = 0;
                    int rnd = (int)(4.0 * Math.random());
                    if ((rnd & 2) == 2) {                          // ~(2 & rnd) == -3
                        cameraRotationY += cameraRotationYIncrement; // oc += Ok
                    }
                    if ((rnd & 1) == 1) {
                        cameraRotationX += mouseButtonMode;          // Be += eg
                    }
                }

                // Bounce the drift increments at +/-50.
                if (cameraRotationX < -50) mouseButtonMode = 2;            // Be<-50 -> eg=2
                if (cameraRotationY < -50) cameraRotationYIncrement = 2;   // oc<-50 -> Ok=2
                if (cameraRotationX > 50)  mouseButtonMode = -2;           // Be>50  -> eg=-2

                // Decrement the chat-tab flash timers.
                if (Mh > 0) Mh--;
                if (Vj > 0) Vj--;
                if (Ee > 0) Ee--;
                if (Qe > 0) Qe--;

                if (cameraRotationY > 50) cameraRotationYIncrement = -2;   // oc>50 -> Ok=-2
            } catch (OutOfMemoryError oom) {
                outOfMemory = true;              // Ue = true
            }
        } catch (RuntimeException e) {
            throw ErrorHandler.a(e, STRINGS[227] + frameTick + ')'); // il[227]="client.MA("
        }
    }

    // -------------------------------------------------------------------------
    /** Resolve world host/ports, build XP table, init surface/world/scene/UI, load all assets.
     *  Called from GameShell's loader thread.   obf: void a(byte)   obf-label: il[334]="client.KC(" */
    @Override
    final void loadGameConfig(byte dummy) {
        try {
            // Applet domain-lock: must be served from runescape.com (or local.runescape.com).
            if (isApplet) {                      // hj
                String host = this.getDocumentBase().getHost().toLowerCase();
                // il[333]="runescape.com", il[329]=".runescape.com"
                if (!host.equals(STRINGS[333]) && !host.endsWith(STRINGS[329])) {
                    Xh = true;                   // domain-lock tripped
                    return;
                }
            }

            this.pollInput(-113);                 // n(-113) [input.part: pollInput — window resize/setup]

            // Check/init the display surface; on failure flag a fatal load error.
            if (!this.checkDisplay(2)) {          // e.d(2)
                fatalLoadError = true;            // Vc
                return;
            }

            // Run the content CRC checker/downloader.
            CacheUpdater.initContent(BZip.staticRef, (byte)-72); // cb.a(wb.p, -72)

            // If the loader pre-fetched a cache file, wire it into the archive store.
            try {
                if (ImageLoader.loaderThread.cacheFile != null) {   // pa.k.s
                    Packet.archiveStore = new DataStore(            // b.q = new nb(...)
                        ImageLoader.loaderThread.cacheFile, 24, 0
                    );
                    ImageLoader.loaderThread.cacheFile = null;
                }
            } catch (IOException ex) {
                Packet.archiveStore = null;
            }

            // Build the experience-per-level table (levels 1..99):
            //   xpThis = (int)(300 * 2^(n/7) + n)  with n = lvl+1, accumulated, clamped to 0x0FFFFFFC.
            int xpAcc = 0;
            for (int lvl = 0; lvl < 99; lvl++) {
                int n = lvl + 1;
                int xpThis = (int)(300.0 * Math.pow(2.0, n / 7.0) + n);
                xpAcc += xpThis;
                experienceTable[lvl] = StreamBase.clampXp(xpAcc, 0x0FFFFFFC); // ib.a(sum, 268435452)
            }

            // il[332]="referid" applet param -> Yd.
            try {
                Yd = Integer.parseInt(this.getParameter(STRINGS[332]));
            } catch (Exception ignored) {}

            // il[331]="servertype" applet param -> members/free bits.
            try {
                String serverType = this.getParameter(STRINGS[331]);
                // dummy-param guard (anti-tamper): if dummy != -92, force narrow layout.
                if (dummy != -92) {
                    screenHeight = -6;            // Oi = -6
                }
                int modeVal = Integer.parseInt(serverType);
                if ((modeVal & 2) != 0) isFreeWorld = true;     // cf
                if ((modeVal & 1) != 0) isMembersWorld = true;  // Pg
            } catch (Exception ignored) {}

            // --- Server host / port selection (logic per CLEAN base) ---
            //   if fontMetrics != defaultFont:
            //       if SpriteScaler.a(fontMetrics,-117): codeBase host + (xd=nodeId+50000, fc=40000+nodeId)
            //       else if fontMetrics == BZip.instance: il[328] host + (xd=nodeId+50000, fc=40000+nodeId)
            //       else: leave host unset
            //   else (fontMetrics == defaultFont): codeBase host + (xd=443, fc=43594)
            // NOTE: fc = 40000 - -aa.l = 40000 + nodeId (PLUS, not minus).
            if (Surface.defaultFont != GameShell.fontMetrics) {   // ua.E != e.i
                if (SpriteScaler.canScale(GameShell.fontMetrics, (byte)-117)) { // ia.a(e.i,-117)
                    serverHost = this.getCodeBase().getHost();    // Dh
                    portB = 40000 + BZip.nodeId;                  // fc = 40000 - -aa.l
                    portA = BZip.nodeId + 50000;                  // xd
                } else if (BZip.instance == GameShell.fontMetrics) { // la.b == e.i
                    portA = BZip.nodeId + 50000;                  // xd
                    portB = 40000 + BZip.nodeId;                  // fc
                    serverHost = STRINGS[328];                    // Dh = "local.runescape.com"
                }
                // else: host left unset.
            } else {
                serverHost = this.getCodeBase().getHost();        // Dh
                portB = 43594;                                    // fc
                portA = 443;                                      // xd
            }

            CacheFile.cacheLimit = 1000;          // d.l = 1000

            // Load the "Configuration" options/data archive (ui_a.part defines f(boolean)
            // as drawOptionsTab — its body is actually the config loader). false = no SFX.
            drawOptionsTab(false);                 // f(false)  [ui_a.part: drawOptionsTab]
            if (fatalLoadError) return;

            // --- Sprite-slot base offsets in the SurfaceSprite map ---
            spriteBaseInventory   = 2000;                          // tg
            spriteBaseChars       = spriteBaseInventory + 100;     // hc
            spriteBaseNpcs        = spriteBaseChars + 50;          // sg
            spriteBaseObjects     = spriteBaseNpcs + 1000;         // dg
            spriteBaseGroundItems = spriteBaseObjects + 10;        // kd
            spriteBaseWalls       = spriteBaseGroundItems + 50;    // Eh
            spriteBaseBubbles     = spriteBaseWalls - -10;         // Wj
            spriteBaseTextures    = spriteBaseBubbles + 5;         // ij

            graphics = this.getGraphics();        // Xb

            // GameShell.a(int,byte): set the target frame interval (Ib = 1000/fps).
            this.setTargetFps(50, (byte)107);     // GameShell.a(50, 107)

            // Software renderer: Wd x (Oi+12), 4000 sprite slots.
            surface = new SurfaceSprite(screenWidth, screenHeight + 12, 4000, this); // li = new ba(...)
            surface.mudclient = this;             // li.dc = this
            surface.initSpriteRegion(0, screenWidth, screenHeight + 12, 0, (byte)54); // li.a(0,...)

            // Message lists: il[335]="Choose option" (chat), then friends + ignore.
            chatList    = new MessageList(surface, 1, STRINGS[335]); // zh
            friendsList = new MessageList(surface, 1);                // Wf
            ignoreList  = new MessageList(surface, 1);                // He

            Timer.displayEnabled = false;         // p.d = false
            StringCodec.glyphBase = spriteBaseChars; // u.g = hc

            panelLogin = new Panel(surface, 5);   // Mc
            int panelLeft = surface.width - 199;  // li.u - 199
            Ud = panelLogin.addScrollList(panelLeft, 196, 90, true, dummy ^ -12, 500, 24 + 36, 1);

            panelGame = new Panel(surface, 5);    // zk
            Hi = panelGame.addScrollList(panelLeft, 196, 126, true, dummy + 197, 500, 36 + 40, 1);

            panelQuest = new Panel(surface, 5);   // fe
            lk = panelQuest.addScrollList(panelLeft, 196, 251, true, 106, 500, 24 + 36, 1);

            loadMedia2d((byte)-49);               // m(-49)
            if (fatalLoadError) return;

            loadEntitySprites(true);              // c(true)
            if (fatalLoadError) return;

            // Scene renderer (lb=Scene): 15000x15000, 1000 models.
            scene = new Scene(surface, 15000, 15000, 1000); // Ek = new lb(li,15000,15000,1000)
            scene.initCamera(
                screenHeight / 2, true, screenWidth, screenWidth / 2, screenHeight / 2,
                qd, screenWidth / 2
            );                                    // Ek.a(...)
            scene.visibilityRadius = 2400;        // Ek.Mb
            scene.clipZ = 2400;                   // Ek.X
            scene.clipY = 2300;                   // Ek.G
            scene.renderMode = 1;                 // Ek.P
            scene.setFog(-50, -10, true, -50);    // Ek.a(-50,-10,true,-50)

            // World terrain engine (k=World) bound to the scene + surface.
            world = new World(scene, surface);    // Hh = new k(Ek, li)
            world.spriteBase = spriteBaseInventory; // Hh.x = tg

            loadTextures((byte)91);               // j(91)
            if (fatalLoadError) return;

            loadModelDefs(true);                  // e(true)
            if (fatalLoadError) return;

            loadMaps(5359);                       // m(5359)
            if (fatalLoadError) return;

            if (isMembersWorld) {
                initSounds(-90);                  // E(-90)
            }
            if (fatalLoadError) return;

            // GameShell.a(int,byte,String): loading-progress bar at 100% with the
            // message il[330]="Starting game...".
            this.drawLoadingProgress(100, (byte)-99, STRINGS[330]); // GameShell.a(100,-99,il[330])

            initShopPanel(56);                    // O(56)
            drawLoginScreen(3845);                // p(3845)  [login.part: drawLoginScreen]
            drawCharDesign(dummy ^ 24649);        // t(dummy ^ 0x6049)  [ui_b.part: drawCharDesign]
            resetMenuState((byte)-88);            // e(-88)  [ui_b.part: resetMenuState]
            this.gameShellTick(-77);              // GameShell.a(-77)  (near no-op)
            drawProgressBar(-116);                // y(-116)  [util.part: drawProgressBar]
        } catch (RuntimeException e) {
            throw ErrorHandler.a(e, STRINGS[334] + dummy + ')'); // il[334]="client.KC("
        }
    }

    // -------------------------------------------------------------------------
    /** Load the 3D model name/definition archives (*.ob2 / *.ob3); "3d models".
     *  obf: void e(boolean)   obf-label: il[286]="client.DA(" */
    private final void loadModelDefs(boolean membersContent) {
        try {
            // Base (free) model archives. ca.a((byte)91, name) = GameModel.loadArchive.
            GameModel.loadArchive((byte)91, STRINGS[287]); // "torcha2"
            GameModel.loadArchive((byte)91, STRINGS[284]); // "torcha3"
            GameModel.loadArchive((byte)91, STRINGS[295]); // "torcha4"
            GameModel.loadArchive((byte)91, STRINGS[294]); // "skulltorcha2"
            GameModel.loadArchive((byte)91, STRINGS[275]); // "skulltorcha3"
            GameModel.loadArchive((byte)91, STRINGS[278]); // "skulltorcha4"
            GameModel.loadArchive((byte)91, STRINGS[277]); // "firea2"
            GameModel.loadArchive((byte)91, STRINGS[273]); // "firea3"
            GameModel.loadArchive((byte)91, STRINGS[283]); // "fireplacea2"
            GameModel.loadArchive((byte)91, STRINGS[298]); // "fireplacea3"
            GameModel.loadArchive((byte)91, STRINGS[282]); // "firespell2"

            if (!membersContent) {
                return;                           // free-only set done
            }

            // Members-only model archives.
            GameModel.loadArchive((byte)91, STRINGS[280]); // "firespell3"
            GameModel.loadArchive((byte)91, STRINGS[276]); // "lightning2"
            GameModel.loadArchive((byte)91, STRINGS[289]); // "lightning3"
            GameModel.loadArchive((byte)91, STRINGS[299]); // "clawspell2"
            GameModel.loadArchive((byte)91, STRINGS[293]); // "clawspell3"
            GameModel.loadArchive((byte)91, STRINGS[292]); // "clawspell4"
            GameModel.loadArchive((byte)91, STRINGS[288]); // "clawspell5"
            GameModel.loadArchive((byte)91, STRINGS[291]); // "spellcharge2"
            GameModel.loadArchive((byte)91, STRINGS[281]); // "spellcharge3"

            // Standalone (no GameFrame): load model bodies from the "3d models" archive.
            if (GameFrame.instance == null) {     // kb.a == null
                byte[] modelData = this.fetchAsset(STRINGS[285], 60, 9, 84); // il[285]="3d models"
                if (modelData == null) {
                    fatalLoadError = true;        // Vc
                    return;
                }

                for (int i = 0; i < SpriteScaler.modelCount; i++) { // ia.b
                    // il[290]=".ob2"
                    int offset = NameHash.findOffset(
                        NameTable.modelNames[i] + STRINGS[290], (byte)68, modelData
                    );                            // oa.a(...)
                    if (offset == 0) {
                        objectModels[i] = new GameModel(1, 1);
                    } else {
                        objectModels[i] = new GameModel(modelData, offset, true);
                    }
                    // il[296]="giantcrystal" -> mark double-sided.
                    if (NameTable.modelNames[i].equals(STRINGS[296])) {
                        objectModels[i].isDoubleSided = true; // ca.cb
                    }
                }
            } else {
                // Applet: load each model body from a per-name .ob3 file.
                this.drawLoadingProgress(70, (byte)-98, STRINGS[274]); // GameShell.a(70,-98,il[274]="Loading 3d models")

                for (int i = 0; i < SpriteScaler.modelCount; i++) {
                    // il[297]="../content/src/models/", il[279]=".ob3"
                    objectModels[i] = new GameModel(
                        STRINGS[297] + NameTable.modelNames[i] + STRINGS[279]
                    );
                    if (NameTable.modelNames[i].equals(STRINGS[296])) {
                        objectModels[i].isDoubleSided = true;
                    }
                }
            }
        } catch (RuntimeException e) {
            throw ErrorHandler.a(e, STRINGS[286] + membersContent + ')'); // il[286]="client.DA("
        }
    }

    // -------------------------------------------------------------------------
    /** Load the 2D UI sprite archives ("2d graphics"); blit into SurfaceSprite slots.
     *  obf: void m(byte)   obf-label: il[104]="client.IA(" */
    private final void loadMedia2d(byte widescreen) {
        try {
            // il[110]="2d graphics" archive; il[103]="index.dat" shared index.
            byte[] data = this.fetchAsset(STRINGS[110], 20, 8, 76);
            if (data == null) {
                fatalLoadError = true;            // Vc
                return;
            }

            byte[] index = StreamFactory.extractEntry(STRINGS[103], 0, data, -128); // na.a("index.dat",...)

            // surface.loadSprites(slotBase, layerCount, entryData, spriteCount, index).
            // Sprite filenames are the real decoded il[] strings.
            surface.loadSprites(spriteBaseInventory,      1, StreamFactory.extractEntry(STRINGS[111], 0, data, -118), 120, index); // inv1.dat
            surface.loadSprites(spriteBaseInventory + 1,  6, StreamFactory.extractEntry(STRINGS[95],  0, data, -119),  52, index); // inv2.dat
            surface.loadSprites(spriteBaseInventory + 9,  1, StreamFactory.extractEntry(STRINGS[98],  0, data, -121), 101, index); // bubble.dat
            surface.loadSprites(spriteBaseInventory + 10, 1, StreamFactory.extractEntry(STRINGS[109], 0, data, -127),  86, index); // runescape.dat
            surface.loadSprites(spriteBaseInventory + 11, 3, StreamFactory.extractEntry(STRINGS[101], 0, data, -122),  84, index); // splat.dat
            surface.loadSprites(spriteBaseInventory + 14, 8, StreamFactory.extractEntry(STRINGS[99],  0, data, -120), 111, index); // icon.dat
            surface.loadSprites(spriteBaseInventory + 22, 1, StreamFactory.extractEntry(STRINGS[112], 0, data, -124), 112, index); // hbar.dat
            surface.loadSprites(spriteBaseInventory + 23, 1, StreamFactory.extractEntry(STRINGS[97],  0, data, -121), 104, index); // hbar2.dat
            surface.loadSprites(spriteBaseInventory + 24, 1, StreamFactory.extractEntry(STRINGS[96],  0, data, -128),  73, index); // compass.dat
            surface.loadSprites(spriteBaseInventory + 25, 2, StreamFactory.extractEntry(STRINGS[100], 0, data, -127), 100, index); // buttons.dat
            surface.loadSprites(spriteBaseChars,          2, StreamFactory.extractEntry(STRINGS[106], 0, data, -127), 125, index); // scrollbar.dat
            surface.loadSprites(spriteBaseChars + 2,      4, StreamFactory.extractEntry(STRINGS[93],  0, data, -125),  68, index); // corners.dat

            // Widescreen members layout: extend the status bar.
            if (widescreen > -1) {
                screenHeight = 24;                // Oi = 24
            }

            surface.loadSprites(spriteBaseChars + 6,  2,            StreamFactory.extractEntry(STRINGS[107], 0, data, -118),  74, index); // arrows.dat
            surface.loadSprites(spriteBaseGroundItems, FontWidths.charCount,
                                                                   StreamFactory.extractEntry(STRINGS[105], 0, data, -124),  83, index); // projectile.dat
            surface.loadSprites(spriteBaseBubbles,     2,          StreamFactory.extractEntry(STRINGS[108], 0, data, -123), 116, index); // crowns.dat

            surface.drawSeparator(-123, spriteBaseBubbles); // li.d(-123, Wj)

            // Numbered "objects"+N+".dat" NPC sprite sheets (30 per sheet), Utility.spriteSheetCount total.
            int remaining = Utility.spriteSheetCount;  // mb.l
            int sheet = 1;
            while (remaining > 0) {
                int count = (remaining <= 30) ? remaining : 30;
                remaining -= 30;
                surface.loadSprites(
                    spriteBaseNpcs + 30 * (sheet - 1), count,
                    StreamFactory.extractEntry(STRINGS[94] + sheet + STRINGS[102], 0, data, -122),
                    109, index
                );                                // il[94]="objects", il[102]=".dat"
                sheet++;
            }

            // Register chroma-key separators for each filled sprite slot.
            surface.drawSeparator2(spriteBaseInventory, -342059728);     // li.b(tg, colour)
            surface.drawSeparator2(spriteBaseInventory + 9, -342059728);
            for (int slot = 11; slot < 26; slot++) {
                surface.drawSeparator2(spriteBaseInventory + slot, -342059728);
            }
            for (int slot = 0; slot < FontWidths.charCount; slot++) {    // n.c
                surface.drawSeparator2(slot + spriteBaseGroundItems, -342059728);
            }
            for (int slot = 0; slot < Utility.spriteSheetCount; slot++) { // mb.l
                surface.drawSeparator2(slot + spriteBaseNpcs, -342059728);
            }
        } catch (RuntimeException e) {
            throw ErrorHandler.a(e, STRINGS[104] + widescreen + ')'); // il[104]="client.IA("
        }
    }

    // -------------------------------------------------------------------------
    /** Load player/NPC animation frame sprites ("people and monsters" / "a.dat" / "f.dat").
     *  obf: void c(boolean)   obf-label: il[325]="client.KD(" */
    private final void loadEntitySprites(boolean doLoad) {
        try {
            if (!doLoad) {
                return;
            }

            // il[324]="people and monsters" archive; il[103]="index.dat" index.
            byte[] data = this.fetchAsset(STRINGS[324], 30, 1, 88);
            if (data == null) {
                fatalLoadError = true;            // Vc
                return;
            }

            byte[] index = StreamFactory.extractEntry(STRINGS[103], 0, data, -120); // na.a("index.dat",...)

            // Members get an additional "member graphics" archive (il[326]).
            byte[] membersData = null;
            byte[] membersIndex = null;
            if (isMembersWorld) {                 // Pg
                membersData = this.fetchAsset(STRINGS[326], 45, 2, 68);
                if (membersData == null) {
                    fatalLoadError = true;        // Vc
                    return;
                }
                membersIndex = StreamFactory.extractEntry(STRINGS[103], 0, membersData, -121);
            }

            dj = 0;
            uc = dj;                              // running sprite cursor (li slot index)
            int frameCount = 0;

            int total = StreamFactory.animationCount; // na.e (= GameData.animationCount)
            label131:
            for (int idx = 0; idx < total; idx++) {
                String name = CacheUpdater.animationNames[idx]; // cb.e[idx]

                // De-dup: if an earlier entry shares this name, alias its sprite slot and
                // SKIP loading entirely — note uc is NOT advanced for duplicates (oracle-verified).
                for (int prev = 0; prev < idx; prev++) {
                    if (CacheUpdater.animationNames[prev].equalsIgnoreCase(name)) {
                        World.animationNumber[idx] = World.animationNumber[prev]; // w.g[idx]=w.g[prev]
                        continue label131;
                    }
                }

                // Body sprite (15 frames): from the base archive, else (members) the member archive.
                byte[] bodyData  = StreamFactory.extractEntry(name + STRINGS[102], 0, data, -124); // ".dat"
                byte[] bodyIndex = index;
                if (bodyData == null && isMembersWorld) {
                    bodyIndex = membersIndex;
                    bodyData  = StreamFactory.extractEntry(name + STRINGS[102], 0, membersData, -127);
                }

                if (bodyData != null) {
                    frameCount += 15;
                    surface.loadSprites(uc, 15, bodyData, 83, bodyIndex); // li.a(uc,15,...)

                    // "a.dat" extra-frame set (3 frames), when flagged (nb.d[idx] == 1).
                    if (NameTable.animationHasA[idx] == 1) { // ~nb.d[idx] == -2  <=>  nb.d[idx] == 1
                        byte[] animData  = StreamFactory.extractEntry(name + STRINGS[321], 0, data, -124); // "a.dat"
                        byte[] animIndex = index;
                        if (animData == null && isMembersWorld) {
                            animData  = StreamFactory.extractEntry(name + STRINGS[321], 0, membersData, -121);
                            animIndex = membersIndex;
                        }
                        frameCount += 3;
                        surface.loadSprites(uc + 15, 3, animData, 89, animIndex);
                    }

                    // "f.dat" front/face-frame set (9 frames), when flagged (aa.c[idx] == 1).
                    if (BZip.animationHasF[idx] == 1) {
                        byte[] faceData  = StreamFactory.extractEntry(name + STRINGS[323], 0, data, -123); // "f.dat"
                        byte[] faceIndex = index;
                        if (faceData == null && isMembersWorld) {
                            faceData  = StreamFactory.extractEntry(name + STRINGS[323], 0, membersData, -118);
                            faceIndex = membersIndex;
                        }
                        surface.loadSprites(uc + 18, 9, faceData, 76, faceIndex);
                        frameCount += 9;
                    }

                    // Register chroma separators across the 27-slot block (when n.m[idx] != 0).
                    if (FontWidths.animationSomething[idx] != 0) { // ~n.m[idx] != -1  <=>  n.m[idx] != 0
                        for (int slot = uc; slot < uc + 27; slot++) {
                            surface.drawSeparator2(slot, -342059728); // li.b(slot, colour)
                        }
                    }
                }

                World.animationNumber[idx] = uc; // w.g[idx] = uc
                uc += 27;
            }

            // il[322]="Loaded: ", il[327]=" frames of animation"
            System.out.println(STRINGS[322] + frameCount + STRINGS[327]);
        } catch (RuntimeException e) {
            throw ErrorHandler.a(e, STRINGS[325] + doLoad + ')'); // il[325]="client.KD("
        }
    }

    // -------------------------------------------------------------------------
    /** Load landscape/map archives into the World ("landscape"/"map"/members variants).
     *  obf: void m(int)   obf-label: il[603]="client.ED(" */
    private final void loadMaps(int dummy) {
        try {
            // il[602]="map" -> world.gb ; il[599]="landscape" -> world.Q
            // (Hh = world, type k = World per NAMING.md.)
            world.mapData = this.fetchAsset(STRINGS[602], 70, 4, 66); // Hh.gb

            if (isMembersWorld) {                 // Pg
                // il[601]="members map" -> world.m
                world.membersMapData = this.fetchAsset(STRINGS[601], 75, 5, 76); // Hh.m
            }

            // il[599]="landscape" -> world.Q
            world.landscapeData = this.fetchAsset(STRINGS[599], 80, 6, 54); // Hh.Q

            // Anti-tamper timing guard (dummy != 5359): drawSprite no-op — stripped.

            if (isMembersWorld) {
                // il[600]="members landscape" -> world.I ; (dummy ^ 5283) is the obf'd crc arg.
                world.membersLandscapeData = this.fetchAsset(STRINGS[600], 85, 7, dummy ^ 5283); // Hh.I
            }
        } catch (RuntimeException e) {
            throw ErrorHandler.a(e, STRINGS[603] + dummy + ')'); // il[603]="client.ED("
        }
    }

    // -------------------------------------------------------------------------
    /** Load the "Textures" archive into the Scene renderer.
     *  obf: void j(byte)   obf-label: il[241]="client.UB(" */
    private final void loadTextures(byte dummy) {
        try {
            // Junk: int j = -11 % ((-66 - dummy) / 55) — opaque, stripped.

            // il[240]="Textures" archive; il[103]="index.dat" index.
            byte[] data = this.fetchAsset(STRINGS[240], 50, 11, 111);
            if (data == null) {
                fatalLoadError = true;            // Vc
                return;
            }

            byte[] index = StreamFactory.extractEntry(STRINGS[103], 0, data, -122); // na.a("index.dat",...)

            // Allocate Scene texture slots (Ek = scene, type lb = Scene per NAMING.md).
            scene.initTextureSlots(0, 11, 7, DownloadWorker.textureCount); // Ek.a(0,11,7,jb.o)

            for (int i = 0; i < DownloadWorker.textureCount; i++) { // jb.o
                String texName = Utility.textureNames[i];           // mb.g[i]
                // il[102]=".dat"
                byte[] texEntry = StreamFactory.extractEntry(texName + STRINGS[102], 0, data, -125);

                // Parse the texture sprite into the Eh scratch slot, fill a 128x128 magenta
                // chroma box, then draw the sprite over it. (Clean order: parse, box, draw.)
                surface.loadSprites(spriteBaseWalls, 1, texEntry, 88, index);   // li.a(Eh,1,var7,88,var4)
                surface.drawBox(0, (byte)-117, 0xFF00FF, 0, 128, 128);         // li.a(0,-117,16711935,0,128,128)
                surface.trimSprite(-1, spriteBaseWalls, 0, 0);                  // li.b(-1,Eh,0,0)

                int texSize = surface.spriteSizes[spriteBaseWalls];             // li.Eb[Eh]

                // Optional overlay variant (Timer.altTextureNames[i] = p.c[i]).
                String altName = Timer.altTextureNames[i];                      // p.c[i]
                if (altName != null && altName.length() > 0) {
                    byte[] altEntry = StreamFactory.extractEntry(altName + STRINGS[102], 0, data, -121);
                    surface.loadSprites(spriteBaseWalls, 1, altEntry, 109, index);
                    surface.trimSprite(-1, spriteBaseWalls, 0, 0);
                }

                // Register the texture sprite at slot (ij + i): li.d(ij+i, size, 113, size, 0, 0).
                surface.addSpriteToScene(i + spriteBaseTextures, texSize, 113, texSize, 0, 0);

                int texSizeSq = texSize * texSize;
                // Chroma-key fix in the raw pixel buffer: pixel 0x00FF00 (green) -> 0xFF00FF (magenta).
                // clean: if (~li.ob[ij+i][px] == -65281) li.ob[ij+i][px] = 16711935;
                for (int px = 0; px < texSizeSq; px++) {
                    if (surface.spritePixels[spriteBaseTextures + i][px] == 0x00FF00) {
                        surface.spritePixels[spriteBaseTextures + i][px] = 0xFF00FF;
                    }
                }

                surface.unloadSprite(false, i + spriteBaseTextures);            // li.a(false, ij+i)

                // Register texture with the Scene: Ek.a(i, 74, pixelData, size/64-1, alphaData).
                scene.setTexture(
                    i, (byte)74,
                    surface.spriteData[spriteBaseTextures + i],   // li.Y[ij+i]
                    texSize / 64 - 1,
                    surface.spriteAlpha[spriteBaseTextures + i]   // li.gb[ij+i]
                );
            }
        } catch (RuntimeException e) {
            throw ErrorHandler.a(e, STRINGS[241] + dummy + ')'); // il[241]="client.UB("
        }
    }

    // -------------------------------------------------------------------------
    /** GameShell paint hook: fatal-error / domain-lock / out-of-memory screens, else normal render.
     *  obf: void b(boolean)   obf-label: il[481]="client.JD(" */
    @Override
    final void drawLoadError(boolean clearDomainLock) {
        try {
            // First-paint hook: bump the loader bar once.
            if (this.N) {                         // GameShell.N
                this.pollInput(-108);             // n(-108)  [input.part: pollInput]
                this.N = false;
            }

            if (!fatalLoadError) {                // !Vc
                if (Xh) {
                    // --- Domain-lock: "Error - unable to load game!" ---
                    Graphics g = this.getGraphics();
                    if (g != null) {
                        g.translate(this.Eb, this.K);
                        g.setColor(Color.black);
                        g.fillRect(0, 0, 512, 356);
                        g.setFont(new Font(STRINGS[477], Font.BOLD, 20)); // il[477]="Helvetica"
                        g.setColor(Color.white);
                        g.drawString(STRINGS[485], 50, 50);  // "Error - unable to load game!"
                        g.drawString(STRINGS[492], 50, 100); // "To play RuneScape make sure you play from"
                        g.drawString(STRINGS[495], 50, 150); // "http://www.runescape.com"
                        this.setTargetFps(1, (byte)111);     // GameShell.a(1, 111) — slow repaint (Ib=1000ms)
                    }
                } else if (outOfMemory) {         // Ue
                    // --- Out-of-memory screen ---
                    Graphics g = this.getGraphics();
                    if (g != null) {
                        g.translate(this.Eb, this.K);
                        g.setColor(Color.black);
                        g.fillRect(0, 0, 512, 356);
                        g.setFont(new Font(STRINGS[477], Font.BOLD, 20));
                        g.setColor(Color.white);
                        g.drawString(STRINGS[482], 50, 50);  // "Error - out of memory!"
                        g.drawString(STRINGS[488], 50, 100); // "Close ALL unnecessary programs"
                        g.drawString(STRINGS[494], 50, 150); // "and windows before loading the game"
                        g.drawString(STRINGS[491], 50, 200); // "RuneScape needs about 48meg of spare RAM"
                        this.setTargetFps(1, (byte)106);     // GameShell.a(1, 106) — slow repaint
                    }
                } else {
                    // --- Normal frame ---
                    try {
                        if (clearDomainLock) {
                            Xh = false;
                        }
                        if (surface == null) {    // li
                            return;
                        }
                        if (screenMode == 0) {    // ~qg == -1  ->  qg == 0 (login)
                            surface.minimap = false; // li.xb = false
                            drawMinimap(2540);    // k(2540)
                        }
                        if (screenMode == 1) {    // ~qg == -2  ->  qg == 1 (in-game)
                            surface.minimap = true;  // li.xb = true
                            drawGameFrame(13);    // f(13)  [ui_a.part: drawGameFrame]
                        }
                    } catch (OutOfMemoryError oom) {
                        outOfMemory = true;       // Ue = true
                    }
                }
            } else {
                // --- Fatal load error (Vc): the multi-step "Sorry, an error..." help screen ---
                Graphics g = this.getGraphics();
                if (g != null) {
                    g.translate(this.Eb, this.K);
                    g.setColor(Color.black);
                    g.fillRect(0, 0, 512, 356);
                    g.setFont(new Font(STRINGS[477], Font.BOLD, 16));
                    g.setColor(Color.yellow);
                    int y = 35;
                    g.drawString(STRINGS[493], 30, y); // "Sorry, an error has occured whilst loading RuneScape"
                    g.setColor(Color.white);
                    y += 50;
                    g.drawString(STRINGS[487], 30, y); // "To fix this try the following (in order):"
                    g.setColor(Color.white);
                    y += 50;
                    g.setFont(new Font(STRINGS[477], Font.BOLD, 12));
                    g.drawString(STRINGS[484], 30, y); // "1: Try closing ALL open web-browser windows..."
                    y += 30;
                    g.drawString(STRINGS[489], 30, y); // "2: Try clearing your web-browsers cache..."
                    y += 30;
                    g.drawString(STRINGS[483], 30, y); // "3: Try using a different game-world"
                    y += 30;
                    g.drawString(STRINGS[486], 30, y); // "4: Try rebooting your computer"
                    y += 30;
                    g.drawString(STRINGS[490], 30, y); // "5: Try selecting a different version of Java..."
                    this.setTargetFps(1, (byte)126);   // GameShell.a(1, 126) — slow repaint
                }
            }
        } catch (RuntimeException e) {
            throw ErrorHandler.a(e, STRINGS[481] + clearDomainLock + ')'); // il[481]="client.JD("
        }
    }

    // -------------------------------------------------------------------------
    /** String-pool XOR decoder stage 1: single-char strings XOR with 126 (0x7E).
     *  obf: static char[] z(String) */
    private static char[] xorDecode1(String s) {
        char[] chars = s.toCharArray();
        if (chars.length < 2) {
            chars[0] = (char)(chars[0] ^ 126);
        }
        return chars;
    }

    // -------------------------------------------------------------------------
    /** String-pool XOR decoder stage 2: rotating 5-byte XOR key {34,7,117,116,126}, interned.
     *  obf: static String z(char[]) */
    private static String xorDecode2(char[] chars) {
        for (int i = 0; i < chars.length; i++) {
            byte key;
            switch (i % 5) {
                case 0:  key = 34;  break;  // '"'
                case 1:  key = 7;   break;  // BEL
                case 2:  key = 117; break;  // 'u'
                case 3:  key = 116; break;  // 't'
                default: key = 126; break;  // '~'
            }
            chars[i] = (char)(chars[i] ^ key);
        }
        return new String(chars).intern();
    }
