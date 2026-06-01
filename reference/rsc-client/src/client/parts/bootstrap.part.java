// ===== bootstrap =====
// Methods in group "bootstrap" deobfuscated from client.java (Microsoft J++ rev ~233-235).
// Opaque predicate (client.OPAQUE_FALSE / vh) and dead control-flow stripped.
// Profiling counter increments stripped.
// try/catch(RuntimeException e){throw ErrorHandler.a(e,"sig")} wrappers unwrapped.
// Anti-tamper modulo guards and dummy-param junk stripped.
// All field/class names use the canonical names from NAMING.md and MUDCLIENT_SKELETON.md.
//
// Class context: package client; class Mudclient extends GameShell (e)
// STRINGS = client.il  (XOR-decoded string pool)

    // -------------------------------------------------------------------------
    /** Standalone entry point: parse args, select audio backend, create and start client frame.
     *  obf: static void main(String[]) */
    public static final void main(String[] args) {
        // args[0] = nodeid (int), args[1] = mode string ("live"/"classic"/"members"),
        // args[2..] = optional flags ("veterans", "free")

        // Set the BZip reference used by the downloader/decompressor
        // (la.b is the static BZip instance; goes into GameShell.i / LinkedQueue.f)
        GameShell.audioQueue = BZip.instance;

        // Parse the node/world id
        BZip.nodeId = Integer.parseInt(args[0]);

        // Select audio backend based on mode string
        // STRINGS[312]="live", STRINGS[317]="classic", STRINGS[318]="members"
        if (args[1].equals(STRINGS[312])) {
            // "live": use AudioMixer backend
            LinkedQueue.audioFactory = AudioMixer.instance;
        } else if (args[1].equals(STRINGS[317])) {
            // "classic": use SurfaceImageProducer (software) backend
            LinkedQueue.audioFactory = SurfaceImageProducer.audioInstance;
        } else if (args[1].equals(STRINGS[318])) {
            // "members": use RecordLoader (streaming) backend
            LinkedQueue.audioFactory = RecordLoader.audioInstance;
        }
        // else: no audio backend change

        // Construct the main client instance (is-applet = false for standalone)
        Mudclient client = new Mudclient();
        client.isApplet = false; // hj = false

        // Parse optional flags from args[2..]
        for (int i = 2; i < args.length; i++) {
            if (args[i].equals(STRINGS[316])) { // "veterans"
                client.isMembersWorld = true; // Pg
            }
            if (args[i].equals(STRINGS[315])) { // "free"
                client.isFreeWorld = true;     // cf
            }
        }

        // Create the AWT frame and start the game thread
        // STRINGS[314] = window title prefix, STRINGS[319] = window title suffix
        // 32 + LinkedQueue.audioFactory.threadPriority = frame priority
        // BZip.nodeId + 7000 = port, (byte)112 = flags, fa.d = display depth
        // client.Wd = width (512), client.Oi - -12 = height (334+12=346)
        try {
            client.a(
                false,                                    // standalone (not applet)
                STRINGS[314],                             // title prefix ("RuneScape")
                32 + LinkedQueue.audioFactory.threadPriority, // frame priority
                STRINGS[319],                             // title suffix
                BZip.nodeId + 7000,                       // server port
                (byte)112,                                // display flags
                ClientIOException.displayDepth,           // colour depth
                client.screenWidth,                       // Wd = 512
                client.screenHeight + 12                  // Oi + 12
            );
            client.ticksPerFrame = 10;                    // Q = 10 (target FPS hint)
        } catch (Exception e) {
            Utility.reportError(0x1FFFFF, e, null);
        }
    }

    // -------------------------------------------------------------------------
    /** Applet init: read nodeid/modewhat/modewhere params, size window, kick off loading.
     *  obf: void init()   obf-label: client.init() */
    @Override
    public final void init() {
        // STRINGS[182] = "nodeid" param name; read it into BZip.nodeId (aa.l)
        BZip.nodeId = Integer.parseInt(this.getParameter(STRINGS[182]));

        // STRINGS[185] = font-size param; build a NameTable font (ua/ub alias)
        // ub.a(int size, byte hint) returns a Surface.FontMetrics reference (Lv)
        GameShell.fontMetrics = NameTable.buildFont(
            Integer.parseInt(this.getParameter(STRINGS[185])), (byte)24
        );
        if (GameShell.fontMetrics == null) {
            GameShell.fontMetrics = Surface.defaultFont; // ua.E = default
        }

        // STRINGS[184] = "modewhat" param; selects audio/queue factory
        // StringCodec.a(false, int) picks a LinkedQueue factory from the id
        try {
            LinkedQueue.audioFactory = StringCodec.buildQueue(
                false, Integer.parseInt(this.getParameter(STRINGS[184]))
            );
        } catch (RuntimeException ignored) {
            LinkedQueue.audioFactory = AudioMixer.instance; // eb.e = default mixer
        }
        if (LinkedQueue.audioFactory == null) {
            LinkedQueue.audioFactory = AudioMixer.instance;
        }

        // Start the GameShell init thread:
        // super.a(height, displayDepth, threadPriority, ticksPerFrame, screenWidth)
        // screenHeight = Oi+12, ClientIOException.d = display depth/flags
        // LinkedQueue.audioFactory.threadPriority + 32 = thread priority
        super.startLoaderThread(
            this.screenHeight + 12,           // Oi + 12
            ClientIOException.displayDepth,   // fa.d
            LinkedQueue.audioFactory.threadPriority + 32, // audio thread priority + 32
            2,                                // ticks-per-frame hint
            this.screenWidth                  // Wd
        );
    }

    // -------------------------------------------------------------------------
    /** Constructor: allocates all state arrays (skills/quests/entities/models) and sets defaults.
     *  obf: client()  (no label — constructor) */
    public Mudclient() {
        super();

        // --- Network ---
        incomingPacket = new BitBuffer(5000);    // mg: inbound packet bit-buffer (5000 bytes)

        // --- Basic scalars / cursors ---
        Nc = 0;
        Vg = 0;
        qd = 9;                                  // camera zoom default

        // --- Input timing ---
        mouseClickTimes = new long[100];         // Zd: click timestamp ring

        // --- Login / session ---
        Wc = 0;
        Oj = 0;
        jk = 0;
        loginTimeout = 550;                      // ac: login retry timeout counter

        // --- Flags ---
        isApplet = true;                         // hj
        yj = -1;
        De = 0;
        npcCount = 0;                            // If
        Xh = false;                              // Xh: domain-lock tripped

        // --- Render/screen ---
        si = 1;                                  // sprite page / animation phase seed
        xh = 0;
        Ce = 0;
        oc = 0;
        bl = -1;
        qk = 0;
        Sg = -1;
        dc = 0;
        ug = 128;                                // camera yaw step
        Ug = 128;                                // camera pitch step
        yg = -1;

        // --- Misc scratch buffers ---
        Kk = new int[8192];
        pf = new int[8000];

        // --- Mode flags ---
        isMembersWorld = false;                  // Pg
        Cf = 0;
        loginStage = 0;                          // Zb
        screenWidth = 512;                       // Wd
        kg = 0;
        bc = -1;
        screenHeight = 334;                      // Oi
        mouseButtonMode = 2;                     // eg: 2 = two-button default
        rc = 0;
        qe = 0;
        isFreeWorld = false;                     // cf
        nk = 0;
        Yc = 0;
        Ue = false;                              // out-of-memory flag
        Si = 0;
        Ki = 0;
        Vc = false;                              // fatal-load-error flag
        pj = 0;
        uj = new int[8192];
        sk = 0;
        screenMode = 0;                          // qg: 0=login, 1=game, 2=sleep, etc.
        zf = false;
        fpsCap = 40;                             // nc
        Ok = 2;
        oj = 0;
        Fd = 0;
        ui = 0;
        Ag = 0;

        // --- Entities in view ---
        players = new GameCharacter[500];        // Zg
        Be = 0;
        npcCountView = 0;                        // Mg
        playersLast = new GameCharacter[500];    // rg
        npcsCache = new GameCharacter[4000];     // We
        tj = 0;
        Rg = new int[8000];
        worldIndex = 0;                          // Vh: server world index
        localPlayer = new GameCharacter();       // wi
        bg = new int[1500];
        ci = new int[256];
        Cd = false;
        pmTarget = null;                         // Zj
        Rj = new int[256];
        dj = 0;

        // --- Chat colour palette (RGB values for 15 colours) ---
        chatColors = new int[]{
            0xFF0000,  // red
            0xFFBB00,  // orange
            0xFFFF00,  // yellow
            0xA0BF00,  // yellow-green
            0x00E000,  // green
            0x008000,  // dark green
            0x00A080,  // teal
            0x00B0FF,  // sky blue
            0x0080FF,  // blue
            0x30F0,    // purple-blue
            0xE020E0,  // magenta
            0x303030,  // dark grey
            0x604040,  // brown-grey
            0x806080,  // mauve
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

        // --- Skills ---
        skillXp = new int[14];                   // zj
        el = 0;
        Ui = 0;

        // --- Quest / menu ---
        tf = new int[50];
        Bc = 0;
        zd = 0;
        Fg = 0;
        hi = 0;
        Zc = -1;

        // --- Item colour palette (10 colours for equipment colouring) ---
        itemColors = new int[]{
            0xFFCCB0, 0xFF8000, 0x806030, 0x604860, 0x303030,
            0xFF6020, 0xFF4000, 0xFFFFFF, 0x00FF00, 0x00FFFF
        };

        // --- Equipment bonus arrays ---
        oh = new int[18];                        // armour/weapon stats

        // --- Fatigue bar colours ---
        fatigueColors = new int[]{
            0xED0690, 0xCCB6E6, 0xB38000, 0x99B6A6, 0x905120
        };

        ce = 0;

        // --- Scene object model array ---
        objectModels = new GameModel[1000];      // kh

        Xe = new int[256];
        Pf = 0;
        Mi = false;

        // --- Head-slot layer order for character display (8 styles) ---
        Og = new int[]{0, 0, 0, 0, 0, 1, 2, 1};

        skillBase = new int[14];                 // Vb
        // Tg: equipment-slot layer ordering tables per gender/style (12 slots × 8 configs)
        Tg = new int[][] {
            {11,2,9,7,1,6,10,0,5,8,3,4},
            {11,2,9,7,1,6,10,0,5,8,3,4},
            {11,3,2,9,7,1,6,10,0,5,8,4},
            {3,4,2,9,7,1,6,10,8,11,0,5},
            {3,4,2,9,7,1,6,10,8,11,0,5},
            {4,3,2,9,7,1,6,10,8,11,0,5},
            {11,4,2,9,7,1,6,10,0,5,8,3},
            {11,2,9,7,1,6,10,0,5,8,4,3}
        };

        jd = new int[50];

        // --- XP-per-level table (99 entries, filled in loadGameConfig) ---
        experienceTable = new int[99];           // ti

        Ng = new int[500];
        wg = 2;                                  // walk-mode default

        skillCurrent = new int[14];              // Me
        wj = 0;
        nf = new int[50];
        Rd = -1;
        Kd = false;
        Wg = 8;
        th = new int[8];                         // trade slot item counts
        zi = 0;

        // --- Skill names (short) — STRINGS[48]=Attack, etc. ---
        skillNamesShort = new String[] {         // Vk
            STRINGS[48],  STRINGS[543], STRINGS[546], STRINGS[562],
            STRINGS[575], STRINGS[570], STRINGS[16],  STRINGS[548],
            STRINGS[557], STRINGS[591], STRINGS[565], STRINGS[550],
            STRINGS[559], STRINGS[569], STRINGS[560], STRINGS[529],
            STRINGS[567], STRINGS[580]
        };

        chatInputLine = "";                      // Lg
        Ee = 0;
        ke = false;
        Gi = 48;
        bf = new int[50];
        ff = false;
        Sc = new int[50];
        gc = 0;
        screenMode = 0;                          // fg -> reused as screenMode
        weaponBonuses = new int[18];             // cg
        Mh = 0;
        ee = new int[50];
        fd = false;
        gi = new int[50];

        // --- Player entity cache ---
        playersCache = new GameCharacter[5000];  // te

        isMembersAccount = true;                 // Bd: default true (assume members)

        // --- Character animation slot order: default 8-entry table ---
        Pc = new int[]{0, 1, 2, 1, 0, 0, 0, 0};

        Ef = 0;
        Hk = false;

        // --- Wall/boundary model array ---
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

        // --- Skill names (long) — same pool, slightly different set ---
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

        // --- NPC model cache ---
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
        Oc = new int[50];
        questCompleteFlags = new boolean[500];   // Sj

        // --- NPCs in view ---
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

        // --- Quest names (50 entries from STRINGS[529..598]) ---
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

        // --- Misc counters / timers ---
        fpsCap = 30;                             // cl: render FPS cap default
        serverUpdateTick = 1;                    // Sf
        tradeTheirItems = new int[8];            // zc
        xg = 0;
        hf = 0;
        Nh = 0;
        chatEntry = "";                          // ec
        Kh = true;                              // key-handling enabled
        Kg = false;

        skillCurrentLevels = new int[14];        // Bi
        armorBonuses = new int[18];              // Ak
        fi = new boolean[50];
        Di = -1;
        uf = new int[50];
        md = false;
        combatStyleIndex = 14;                   // Lh: default combat style index
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

        // --- Inventory ---
        inventoryItems = new int[35];            // xe
        vk = false;
        Hc = false;

        skillBaseLevels = new int[14];           // Lc

        // --- Combat style names: Controlled / Accurate / Aggressive / Defensive / Rapid ---
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
        inventoryCount = new int[35];            // vf (item quantity per slot)
        menuOptionTargets = new int[50];         // ak
        Ub = false;
        ah = new String[5];
        npcsLast = new GameCharacter[500];       // Tb
    }

    // -------------------------------------------------------------------------
    /** GameShell hook: start a game session — mouse/audio init, called from GameShell.run.
     *  obf: void e(int)   obf-label: client.MA( */
    @Override
    final void startGame(int frameTick) {
        // If domain-lock tripped (Xh) or OOM flag (Ue) set, skip all init
        if (isApplet && Ue) {
            return;
        }
        if (Vc) {
            return; // fatal load error already set
        }

        // Start the audio channel if present
        if (soundChannel != null) {               // ni
            soundChannel.startPlayback();         // ni.a()
        }

        try {
            // Increment login-screen animation frame counter
            loginAnimFrame++;                     // jk++

            if (screenMode == 0) {
                // Show login screen (mode 0 = login flow)
                loginFrameCount = 0;              // sb = 0
                drawLoginScreen(2);               // this.x(2)
            }

            if (screenMode == 1) {
                // In-game: drive sleep CAPTCHA / heartbeat
                loginFrameCount++;                // sb++
                drawSleepScreen(0);               // this.J(0)
            }

            // Reset the per-frame "button click seen" flag
            Qb = 0;

            // Every 500 frames: randomly drift the camera angle for idle animation
            cameraAutoRotTimer++;                 // oj++
            if (cameraAutoRotTimer > 500) {
                cameraAutoRotTimer = 0;
                int rnd = (int)(4.0 * Math.random());
                if ((rnd & 2) == 2) {
                    // Drift camera yaw (oc) by Ok (step direction)
                    cameraYawDrift += cameraYawStep; // oc += Ok
                }
                if ((rnd & 1) == 1) {
                    // Drift camera pitch (Be) by mouse-button-mode value
                    cameraPitchDrift += mouseButtonMode; // Be += eg
                }
            }

            // Clamp yaw drift to [-50, 50]; reverse direction at limits
            if (cameraPitchDrift < -50) mouseButtonMode = 2;    // eg = 2 (reset step)
            if (cameraYawDrift < -50)   cameraYawStep = 2;      // Ok = 2
            if (cameraPitchDrift > 50)  mouseButtonMode = -2;   // eg = -2 (reverse)
            if (cameraYawDrift > 50)    cameraYawStep = -2;     // Ok = -2

            // Decrement misc timers (non-negative guards omitted — handled by checks)
            if (Mh > 0)  Mh--;
            if (Vj > 0)  Vj--;
            if (Ee > 0)  Ee--;
            if (Qe > 0)  Qe--;

        } catch (OutOfMemoryError oom) {
            Ue = true; // set OOM flag — drawLoadError will show the message
        }
    }

    // -------------------------------------------------------------------------
    /** Resolve world host (servertype/referid/*.runescape.com), build XP table, init surface.
     *  Called from GameShell's loader thread.
     *  obf: void a(byte)  (no label) */
    @Override
    final void loadGameConfig(byte dummy) {
        // Domain-lock check: if running as applet and not from runescape.com, set Xh and abort
        if (isApplet) {                          // hj
            String host = this.getDocumentBase().getHost().toLowerCase();
            // STRINGS[333]="localhost", STRINGS[329]=".runescape.com"
            if (!host.equals(STRINGS[333]) && !host.endsWith(STRINGS[329])) {
                isApplet = true;                 // Xh = true (domain-lock flag)
                return;
            }
        }

        // Show "Starting game..." progress (progress step -113)
        drawProgressBar(-113);                   // n(-113)

        // Call GameShell.d(2) — check/init display; if returns false, fatal error
        if (!this.checkDisplay(2)) {             // e.d(2)
            Vc = true;
            return;
        }

        // Init CacheUpdater: download/verify content CRCs
        // wb.p = static BZip2 instance; (byte)-72 = update-check flag
        CacheUpdater.initContent(BZip.staticRef, (byte)-72); // cb.a(wb.p, -72)

        // If a LoaderThread cache file (pa.k.s) was pre-loaded, wire it into Packet.q
        try {
            if (ImageLoader.loaderThread.cacheFile != null) { // pa.k.s != null
                Packet.archiveStore = new DataStore(          // b.q = new nb(...)
                    ImageLoader.loaderThread.cacheFile, 24, 0
                );
                ImageLoader.loaderThread.cacheFile = null;
            }
        } catch (IOException ex) {
            Packet.archiveStore = null;
        }

        // Build experience-per-level table (levels 1..99)
        // Formula: xp[n] = floor(300 * 2^((n+1)/7) + (n+1)); accumulated sum, clamped to 0x0FFFFFFC
        int xpAcc = 0;
        for (int lvl = 0; lvl < 99; lvl++) {
            int n = lvl + 1;
            int xpThis = (int)(300.0 * Math.pow(2.0, (double)n / 7.0) + n);
            xpAcc += xpThis;
            experienceTable[lvl] = StreamBase.clampXp(xpAcc, 0x0FFFFFFC); // ib.a(sum, mask)
        }

        // Read "referid" applet param — world/server reference id → Yd
        try {
            String referid = this.getParameter(STRINGS[332]); // "referid"
            Yd = Integer.parseInt(referid);
        } catch (Exception ignored) {}

        // Read "modewhere" applet param — encodes members/free bits
        // STRINGS[331] = "modewhere"
        try {
            String modeWhere = this.getParameter(STRINGS[331]);
            // dummy param guard: if dummy != -92, reset screenHeight to narrow layout
            if (dummy != -92) {
                screenHeight = -6; // Oi = -6 (narrow mode)
            }
            int modeVal = Integer.parseInt(modeWhere);
            if ((modeVal & 2) != 0) isFreeWorld = true;     // cf
            if ((modeVal & 1) != 0) isMembersWorld = true;  // Pg
        } catch (Exception ignored) {}

        // --- Determine server host and port based on mode ---
        // ua.E == e.i : no special render mode → use default "runescape.com" ports
        // la.b == e.i : BZip mode → use hardcoded host + standard ports
        // Otherwise     ia.a(e.i, -117) → SpriteScaler mode → use codeBase host + custom ports
        if (Surface.defaultFont == GameShell.fontMetrics) {
            // Mode: standard (live) — ports from nodeid
            // xd = nodeId + 50000 (port-A), fc = 40000 - nodeId (port-B), Dh = "runescape.com"
            portA = BZip.nodeId + 50000;          // xd
            portB = 40000 - BZip.nodeId;          // fc
            serverHost = STRINGS[328];             // Dh = "runescape.com"
        } else if (SpriteScaler.canScale(GameShell.fontMetrics, (byte)-117)) {
            // Mode: scaled/proxy — use code base host, fixed ports
            serverHost = this.getCodeBase().getHost();
            portB = 40000 - BZip.nodeId;
            portA = BZip.nodeId + 50000;
        } else if (BZip.instance == GameShell.fontMetrics) {
            // Mode: BZip/classic — use code base + alternate ports (443 / 43594)
            serverHost = this.getCodeBase().getHost();
            portB = 43594;
            portA = 443;
        }

        // Set cache file limit
        CacheFile.cacheLimit = 1000;             // d.l = 1000

        // Load 3D model defs — false = free-world-only set
        loadModels(false);                       // f(false)
        if (Vc) return;

        // --- Sprite-slot layout constants (pixel offsets in the surface sprite map) ---
        // tg = 2000 : inventory/ui sprite base offset
        // hc = tg + 100 : character sprites base
        // sg = hc + 50  : NPC sprites base
        // dg = sg + 1000 : object sprites base
        // kd = dg + 10   : ground item sprites base
        // Eh = kd + 50   : wall sprites base
        // Wj = Eh + 10   : bubble/projectile sprites base  (note: Eh - -10 = Eh + 10)
        // ij = Wj + 5    : texture sprite base
        spriteBaseInventory = 2000;              // tg
        spriteBaseChars     = spriteBaseInventory + 100; // hc
        spriteBaseNpcs      = spriteBaseChars + 50;      // sg
        spriteBaseObjects   = spriteBaseNpcs + 1000;     // dg
        spriteBaseGroundItems = spriteBaseObjects + 10;  // kd
        spriteBaseWalls     = spriteBaseGroundItems + 50;// Eh
        spriteBaseBubbles   = spriteBaseWalls + 10;      // Wj (Eh - -10 = Eh+10)
        spriteBaseTextures  = spriteBaseBubbles + 5;     // ij

        // Get AWT graphics context
        graphics = this.getGraphics();           // Xb

        // Show initial progress at step 50
        drawProgressBar(50);                     // a(50, 107)

        // Create SurfaceSprite (software renderer): width × (height+12), 4000 sprites
        surface = new SurfaceSprite(screenWidth, screenHeight + 12, 4000, (Component)this); // li
        surface.mudclient = this;                 // li.dc = this (ba.dc)
        // Register sprite region 0: full-screen background
        surface.initSpriteRegion(0, screenWidth, screenHeight + 12, 0, (byte)54); // li.a(0,w,h+12,0,54)

        // Create MessageList instances for chat, friends, and ignore lists
        // STRINGS[335] = chat panel title
        chatList    = new MessageList(surface, 1, STRINGS[335]); // zh
        friendsList = new MessageList(surface, 1);                // Wf
        ignoreList  = new MessageList(surface, 1);                // He

        // Disable timer display
        Timer.displayEnabled = false;            // p.d = false
        // Set StringCodec.g to spriteBaseChars (used for font glyph lookups)
        StringCodec.glyphBase = spriteBaseChars; // u.g = hc

        // Create login panel (Mc) with font size 5
        panelLogin = new Panel(surface, 5);      // Mc

        // Surface width minus 199 = right-side panel left edge
        int panelLeft = surface.width - 199;     // li.u - 199

        // Register login panel widgets (scrollable list in login area)
        // Ud = scrollbar handle id, Hi = heading widget id, lk = body widget id
        Ud = panelLogin.addScrollList(
            panelLeft, 196, 90, true, dummy ^ 0xFFFFFFF4, 500, 24 + 36, 1
        );

        panelGame = new Panel(surface, 5);       // zk
        Hi = panelGame.addScrollList(
            panelLeft, 196, 126, true, dummy + 197, 500, 36 + 40, 1
        );

        panelQuest = new Panel(surface, 5);      // fe
        lk = panelQuest.addScrollList(
            panelLeft, 196, 251, true, 106, 500, 24 + 36, 1
        );

        // Load 2D sprite archives — byte -49 = flags hint
        loadMedia2d((byte)-49);                  // m(-49)
        if (Vc) return;

        // Load free-world model defs
        loadModels(true);                        // c(true) [c(boolean) = loadModels(bool)]
        if (Vc) return;

        // Create World (lb) terrain engine: 15000×15000 map, 1000 objects
        world = new World(surface, 15000, 15000, 1000); // Ek = new lb(li,15000,15000,1000)
        world.initCamera(
            screenHeight / 2, true, screenWidth, screenWidth / 2, screenHeight / 2,
            qd, screenWidth / 2
        );
        world.visibilityRadius = 2400;           // Ek.Mb
        world.clipZ = 2400;                      // Ek.X
        world.clipY = 2300;                      // Ek.G
        world.renderMode = 1;                    // Ek.P
        world.setFog(-50, -10, true, -50);       // Ek.a(-50,-10,true,-50)

        // Create Scene (k) bound to World and surface
        scene = new Scene(world, surface);       // Hh = new k(Ek, li)
        scene.spriteBase = spriteBaseInventory;  // Hh.x = tg

        // Load textures into Scene
        loadTextures((byte)91);                  // j(91)
        if (Vc) return;

        // Load 3D model definitions (ob2/ob3 archives)
        loadModelDefs(true);                     // e(true) [e(boolean)]
        if (Vc) return;

        // Load map archives (landscape + map data), arg 5359 = members flag sentinel
        loadMaps(5359);                          // m(5359)
        if (Vc) return;

        // If members world, load sounds
        if (isMembersWorld) {
            initSounds(-90);                     // E(-90)
        }
        if (Vc) return;

        // Load entity sprites with progress message
        // STRINGS[330] = "Loading people and monsters..."
        loadEntitySprites(100, (byte)-99, STRINGS[330]); // a(100, -99, il[330])

        // Init miscellaneous shop/quest/duel panels
        initShopPanel(56);    // O(56)
        drawLoginScreen(3845); // p(3845)

        // Init character design screen (t uses XOR to obfuscate dummy param)
        drawCharDesign(dummy ^ 0x6049); // t(dummy ^ 24649)

        // Reset login state
        resetLoginState((byte)-88); // e(-88) [e(byte) = resetLoginState]

        // Init HUD panel layout
        initHud(-77);         // a(-77)

        // Clear/reset the surface buffer
        clearScreen((byte)-116); // y(-116)
    }

    // -------------------------------------------------------------------------
    /** Load 3D model definition archives (*.ob2 / *.ob3); "Loading 3d models".
     *  obf: void e(boolean)  (no label) */
    private final void loadModelDefs(boolean membersContent) {
        // Load base model defs: 91 models from each of the known ob2 archives
        // STRINGS[287..298] = model archive filenames (e.g. "objects.ob2", etc.)
        // ca.a((byte)91, name) = GameModel.loadFromArchive(91, name) — loads into NameTable cache
        GameModel.loadArchive((byte)91, STRINGS[287]);
        GameModel.loadArchive((byte)91, STRINGS[284]);
        GameModel.loadArchive((byte)91, STRINGS[295]);
        GameModel.loadArchive((byte)91, STRINGS[294]);
        GameModel.loadArchive((byte)91, STRINGS[275]);
        GameModel.loadArchive((byte)91, STRINGS[278]);
        GameModel.loadArchive((byte)91, STRINGS[277]);
        GameModel.loadArchive((byte)91, STRINGS[273]);
        GameModel.loadArchive((byte)91, STRINGS[283]);
        GameModel.loadArchive((byte)91, STRINGS[298]);
        GameModel.loadArchive((byte)91, STRINGS[282]);

        if (!membersContent) {
            return; // free-only set done
        }

        // Members-only additional model archives
        GameModel.loadArchive((byte)91, STRINGS[280]);
        GameModel.loadArchive((byte)91, STRINGS[276]);
        GameModel.loadArchive((byte)91, STRINGS[289]);
        GameModel.loadArchive((byte)91, STRINGS[299]);
        GameModel.loadArchive((byte)91, STRINGS[293]);
        GameModel.loadArchive((byte)91, STRINGS[292]);
        GameModel.loadArchive((byte)91, STRINGS[288]);
        GameModel.loadArchive((byte)91, STRINGS[291]);
        GameModel.loadArchive((byte)91, STRINGS[281]);

        // If running in standalone (no GameFrame / kb.a == null): load from files
        if (GameFrame.instance == null) {         // kb.a == null
            // STRINGS[285] = "models" archive name, fetched from server at priority 60
            byte[] modelData = this.fetchAsset(STRINGS[285], 60, 9, 84);
            if (modelData == null) {
                Vc = true; // fatal: can't load models
                return;
            }

            // For each named model in NameTable.modelNames (ub.c), load into objectModels[i]
            for (int i = 0; i < SpriteScaler.modelCount; i++) { // ia.b = total model count
                // STRINGS[290] = ".ob2" extension
                int offset = NameHash.findOffset(
                    NameTable.modelNames[i] + STRINGS[290], (byte)68, modelData // oa.a
                );
                if (offset == 0) {
                    objectModels[i] = new GameModel(1, 1); // empty placeholder
                } else {
                    objectModels[i] = new GameModel(modelData, offset, true);
                }
                // STRINGS[296] = "body" — mark body model as double-sided
                if (NameTable.modelNames[i].equals(STRINGS[296])) {
                    objectModels[i].isDoubleSided = true; // ca.cb
                }
            }
        } else {
            // Applet mode: show progress, load from in-memory data
            // STRINGS[274] = "Loading 3d models"
            drawProgressBar(70, (byte)-98, STRINGS[274]); // a(70,-98,il[274])

            for (int i = 0; i < SpriteScaler.modelCount; i++) {
                // STRINGS[297] = "../content/src/models/", STRINGS[279] = ".ob3"
                objectModels[i] = new GameModel(
                    STRINGS[297] + NameTable.modelNames[i] + STRINGS[279]
                );
                // Mark body model as double-sided
                if (NameTable.modelNames[i].equals(STRINGS[296])) {
                    objectModels[i].isDoubleSided = true;
                }
            }
        }
    }

    // -------------------------------------------------------------------------
    /** Load 2D sprite archives (inv1/inv2/bubble/buttons/…dat), "2d graphics".
     *  obf: void m(byte)  (no label) */
    private final void loadMedia2d(byte dummy) {
        // STRINGS[110] = "index.dat" (or "media.dat") — the 2D sprite index archive
        // this.fetchAsset(name, priority, crcSlot, crcExpected) → byte[] or null
        byte[] indexData = this.fetchAsset(STRINGS[110], 20, 8, 76);
        if (indexData == null) {
            Vc = true;  // fatal: no media data
            return;
        }

        // STRINGS[103] = "index" — extract the index entry from the archive blob
        byte[] indexEntry = StreamFactory.extractEntry(STRINGS[103], 0, indexData, -128);
        // na.a(name, 0, data, key) = ArchiveReader.extract(...)

        // Load all named sprite archives into the SurfaceSprite's sprite slots.
        // Each call: surface.a(slotBase, layerCount, archiveData, spriteCount, indexEntry)
        // Sprite slot layout (tg = spriteBaseInventory = 2000):
        //   tg+0  : inv1.dat  (120 sprites)
        //   tg+1  : inv2.dat  (52 sprites)
        //   tg+9  : bubble.dat (101 sprites)
        //   tg+10 : buttons.dat (86 sprites)
        //   tg+11 : scrollbar.dat (84 sprites)
        //   tg+14 : hitsplats.dat (111 sprites)
        //   tg+22 : skills.dat (112 sprites)
        //   tg+23 : items.dat (104 sprites)
        //   tg+24 : icons.dat (73 sprites)
        //   tg+25 : edges.dat (100 sprites)
        //   hc+0  : chars0.dat (125 sprites)
        //   hc+2  : chars2.dat (68 sprites)
        //   etc.

        // STRINGS[111..112, 93..101] = sprite archive filenames
        surface.loadSprites(spriteBaseInventory, 1,
            StreamFactory.extractEntry(STRINGS[111], 0, indexData, -118), 120, indexEntry);
        surface.loadSprites(spriteBaseInventory + 1, 6,
            StreamFactory.extractEntry(STRINGS[95],  0, indexData, -119), 52, indexEntry);
        surface.loadSprites(spriteBaseInventory + 9, 1,
            StreamFactory.extractEntry(STRINGS[98],  0, indexData, -121), 101, indexEntry);
        surface.loadSprites(spriteBaseInventory + 10, 1,
            StreamFactory.extractEntry(STRINGS[109], 0, indexData, -127), 86, indexEntry);
        surface.loadSprites(spriteBaseInventory + 11, 3,
            StreamFactory.extractEntry(STRINGS[101], 0, indexData, -122), 84, indexEntry);
        surface.loadSprites(spriteBaseInventory + 14, 8,
            StreamFactory.extractEntry(STRINGS[99],  0, indexData, -120), 111, indexEntry);
        surface.loadSprites(spriteBaseInventory + 22, 1,
            StreamFactory.extractEntry(STRINGS[112], 0, indexData, -124), 112, indexEntry);
        surface.loadSprites(spriteBaseInventory + 23, 1,
            StreamFactory.extractEntry(STRINGS[97],  0, indexData, -121), 104, indexEntry);
        surface.loadSprites(spriteBaseInventory + 24, 1,
            StreamFactory.extractEntry(STRINGS[96],  0, indexData, -128), 73, indexEntry);
        surface.loadSprites(spriteBaseInventory + 25, 2,
            StreamFactory.extractEntry(STRINGS[100], 0, indexData, -127), 100, indexEntry);
        surface.loadSprites(spriteBaseChars, 2,
            StreamFactory.extractEntry(STRINGS[106], 0, indexData, -127), 125, indexEntry);
        surface.loadSprites(spriteBaseChars + 2, 4,
            StreamFactory.extractEntry(STRINGS[93],  0, indexData, -125), 68, indexEntry);

        // If dummy > -1, switch to wide-screen layout (Oi = 24 for members bar)
        if (dummy > -1) {
            screenHeight = 24; // Oi = 24
        }

        surface.loadSprites(spriteBaseChars + 6, 2,
            StreamFactory.extractEntry(STRINGS[107], 0, indexData, -118), 74, indexEntry);
        // kd = ground-item sprite base; n.c = FontWidths.charCount
        surface.loadSprites(spriteBaseGroundItems, FontWidths.charCount,
            StreamFactory.extractEntry(STRINGS[105], 0, indexData, -124), 83, indexEntry);
        surface.loadSprites(spriteBaseWalls, 2,
            StreamFactory.extractEntry(STRINGS[108], 0, indexData, -123), 116, indexEntry);

        // Register separator lines in the surface
        surface.drawSeparator(-123, spriteBaseWalls); // li.d(-123, Wj)

        // Load numbered sprite sheets (STRINGS[94]+"N"+STRINGS[102] = "inv"+N+".dat")
        int sheet = 1;
        for (int remaining = Utility.spriteSheetCount; remaining > 0; remaining -= 30, sheet++) {
            // STRINGS[94] = "inv", STRINGS[102] = ".dat"
            int count = (remaining < 31) ? remaining : 30;
            surface.loadSprites(
                spriteBaseNpcs - (30 * (sheet - 1)), count,
                StreamFactory.extractEntry(
                    STRINGS[94] + sheet + STRINGS[102], 0, indexData, -122
                ),
                109, indexEntry
            );
        }

        // Draw separator lines for equipment / skill-tab areas
        surface.drawSeparator(spriteBaseInventory, -342059728);  // li.b(tg, colour)
        surface.drawSeparator(spriteBaseInventory + 9, -342059728);

        // Draw separators for each sprite slot above tg+11 up to tg+25
        for (int slot = 11; slot < 26; slot++) {
            surface.drawSeparator(spriteBaseInventory + slot, -342059728);
        }

        // Draw separators for ground-item slots up to FontWidths.charCount
        for (int slot = 0; slot < FontWidths.charCount; slot++) {
            surface.drawSeparator(slot + spriteBaseGroundItems, -342059728);
        }

        // Draw separators for NPC sprite sheets up to Utility.spriteSheetCount
        for (int slot = 0; slot < Utility.spriteSheetCount; slot++) {
            surface.drawSeparator(slot + spriteBaseNpcs, -342059728);
        }
    }

    // -------------------------------------------------------------------------
    /** Load people/monster animation frames (a.dat / f.dat), "people and monsters".
     *  obf: void c(boolean, int)  (no label) */
    private final void loadEntitySprites(boolean membersContent, int colorSeed) {
        // Draw a frame border around the entity-sprite panel area
        // surface.b(-1, tg-3, 3, li.u-199-49) = draw box outline
        int panelLeft = surface.width - 199;     // li.u - 199
        int rowY = panelLeft - 49;               // li.u - 199 - 49
        surface.drawBox(-1, spriteBaseInventory - 3, 3, rowY);

        // Compute colour for the "selected" state vs normal based on colorSeed
        // STRINGS[356] = "People", STRINGS[351] = "Monsters" labels
        int colWidth  = 196;
        int colHeight = 275;
        // o.a(r,g,b,a) = ISAAC.mixColor(…) — produces a blended ARGB value
        int colNormal   = ISAAC.mixColor(160, 9570, 160, 160);
        int colSelected = colNormal;

        // If zd != -1 (some mode flag), use colorSeed to tint the second bar
        if (zd != -1) {
            colSelected = ISAAC.mixColor(220, colorSeed ^ 9570, 220, 220);
        }
        int colDefault = ISAAC.mixColor(220, 9570, 220, 220);

        // Draw the two tab bars (people / monsters selector)
        surface.drawColumn(128, rowY, 24, 0, colWidth / 2, colHeight / 2, colNormal);
        surface.drawColumn(128, rowY - (colWidth / 2), 24, 0, colWidth / 2, colHeight / 2, colSelected);
        surface.drawColumn(128, rowY, colHeight - 24, 0, 24 + colWidth / 2, colHeight,
            ISAAC.mixColor(220, 9570, 220, 220));
        // Draw label bar
        surface.drawRect(colWidth, 0, rowY, colWidth / 2 - 24, (byte)-61);
        surface.drawRect(rowY - (colWidth / 2), 0 + colWidth / 2, 0, 24, colorSeed);
        surface.drawText(colWidth / 4 + rowY, STRINGS[356], 0, 0, 4, colWidth / 2 + 16); // "People"
        surface.drawText(rowY + (colWidth / 4 - -(colWidth / 2)), STRINGS[351], 0, 0, 4, colWidth / 2 - 16); // "Monsters"

        if (zd == 0) {
            // zd == 0 → People tab selected; show skill/equipment bonuses panel
            // STRINGS[355] = "Equipment Bonuses:"
            int slotY = 72;
            surface.drawString(STRINGS[355], rowY + 5, slotY, 0xFFFF00, false, 3);
            int highlightedRow = -1;
            slotY += 13;
            // Draw 9 bonus rows (attack/defence stats)
            for (int row = 0; row < 9; row++) {
                int rowColor = 0xFFFFFF;
                // Highlight the row if mouse is over it (bounds check vs I = mouseY, xb = mouseX)
                // STRINGS[350] = " : " separator for skill name + value
                surface.drawString(
                    skillNamesShort[row] + STRINGS[350] + oh[row] + "/" + weaponBonuses[row],
                    rowY + 5, slotY, rowColor, false, 1
                );
                slotY += 13;
            }
        }
        // else: zd != 0 → Monsters tab; entity sprite list is drawn elsewhere
    }

    // -------------------------------------------------------------------------
    /** Load landscape/map archives ("landscape"/"map"/"members landscape"/"members map").
     *  obf: void m(int)   obf-label: client.ED( */
    private final void loadMaps(int dummy) {
        // Load free-world landscape data into Scene.gb (world boundary data)
        // this.fetchAsset(name, progress, crcSlot, crcExpected) → byte[] or null
        // STRINGS[602] = "landscape", priority 70, crc slot 4
        scene.landscapeData = this.fetchAsset(STRINGS[602], 70, 4, 66); // Hh.gb

        // If members world, also load members landscape
        // STRINGS[601] = "members landscape"
        if (isMembersWorld) {
            scene.membersLandscapeData = this.fetchAsset(STRINGS[601], 75, 5, 76); // Hh.m
        }

        // Load free-world map data
        // STRINGS[599] = "map"
        scene.mapData = this.fetchAsset(STRINGS[599], 80, 6, 54); // Hh.Q

        // Anti-tamper dummy: if dummy != 5359, call drawSprite as a timing guard (stripped)
        // (if dummy != 5359) this.a(93, (byte)102, -18) → no-op for our purposes

        // If members world, load members map data
        // STRINGS[600] = "members map"; dummy ^ 5283 = obfuscated arg (stripped)
        if (isMembersWorld) {
            scene.membersMmapData = this.fetchAsset(STRINGS[600], 85, 7, dummy ^ 5283); // Hh.I
        }
    }

    // -------------------------------------------------------------------------
    /** Load Textures / index.dat texture archive into the Scene renderer.
     *  obf: void j(byte)  (no label) */
    private final void loadTextures(byte dummy) {
        // Junk: int n2 = -11 % ((-66 - dummy) / 55);  → opaque, stripped

        // Fetch the texture archive
        // STRINGS[240] = "Textures" archive name, priority 50, crcSlot 11
        byte[] texData = this.fetchAsset(STRINGS[240], 50, 11, 111);
        if (texData == null) {
            Vc = true; // fatal: textures missing
            return;
        }

        // Extract the texture index
        // STRINGS[103] = "index", key -122
        byte[] texIndex = StreamFactory.extractEntry(STRINGS[103], 0, texData, -122); // na.a

        // Init the World texture slots: (0, 11, 7, jb.o)
        // World.a(start, count, bpp, numTextures) — allocates texture slots
        world.initTextureSlots(0, 11, 7, DownloadWorker.textureCount); // Ek.a(0,11,7,jb.o)

        // For each texture: load sprite + alpha mask, register in World
        for (int i = 0; i < DownloadWorker.textureCount; i++) {
            // STRINGS from mb.g[] = texture name list (e.g. "wood", "brick", …)
            String texName = Utility.textureNames[i];          // mb.g[i]
            // STRINGS[102] = ".dat" — texture data file suffix
            byte[] texEntry = StreamFactory.extractEntry(
                texName + STRINGS[102], 0, texData, -125       // na.a
            );

            // Blit texture into surface sprite slot Eh (spriteBaseWalls)
            surface.loadSprites(spriteBaseWalls, 1, texEntry, 88, texIndex);  // li.a(Eh,1,data,88,idx)

            // Fill transparent pixels: replace 0xFF00FF (magenta = chroma-key) with 0xFF00FF
            surface.drawRect(0, (byte)-117, 0xFF00FF, 0, 128, 128);           // li.a(0,-117,…)
            surface.drawRect(-1, spriteBaseWalls, 0, 0);                       // li.b(-1,Eh,0,0)

            // Get the loaded texture's pixel-size value
            int texSize = surface.spriteSizes[spriteBaseWalls];                // li.Eb[Eh]

            // Also load alpha/overlay variant if it exists (stored in p.c[] = Timer.altNames[])
            String altName = Timer.altTextureNames[i];                        // p.c[i]
            if (altName != null && altName.length() > 0) {
                byte[] altEntry = StreamFactory.extractEntry(
                    altName + STRINGS[102], 0, texData, -121
                );
                surface.loadSprites(spriteBaseWalls, 1, altEntry, 109, texIndex);
                surface.drawRect(-1, spriteBaseWalls, 0, 0);
            }

            // Register the texture sprite into the scene renderer sprite display list
            // li.d(i + ij, texSize, 113, texSize, 0, 0)
            surface.addSpriteToScene(i + spriteBaseTextures, texSize, 113, texSize, 0, 0);

            int texSizeSq = texSize * texSize;

            // Fix up chroma-key pixels in the raw pixel buffer: replace 0xFF00FF → 0xFF00FF
            // (li.ob[ij+i][pixel] is the pixel array for this texture sprite)
            for (int px = 0; px < texSizeSq; px++) {
                if (~surface.spritePixels[spriteBaseTextures + i][px] == -0xFF00FF - 1) {
                    // was 0xFF00FF (magenta transparent) → replace with opaque magenta
                    surface.spritePixels[spriteBaseTextures + i][px] = 0xFF00FF;
                }
            }

            // Unload the temporary render-target sprite
            surface.unloadSprite(false, i + spriteBaseTextures);               // li.a(false, ij+i)

            // Register with World: World.a(i, 74, pixelArray, size/64-1, alphaArray)
            world.setTexture(
                i, (byte)74,
                surface.spriteData[spriteBaseTextures + i],  // li.Y[ij+i]
                texSize / 64 - 1,
                surface.spriteAlpha[spriteBaseTextures + i]  // li.gb[ij+i]
            );
        }
    }

    // -------------------------------------------------------------------------
    /** Fatal "Error - out of memory!" / "unable to load game!" / domain-lock help screen.
     *  Called by GameShell.paint when Vc (fatal load error) or Xh (domain-lock) is set.
     *  obf: void b(boolean)   obf-label: client.JD( */
    @Override
    final void drawLoadError(boolean clearDomainLock) {
        // Reset any "new-messages" flag (N) on entry
        if (this.N) {                            // GameShell.N
            drawProgressBar(-108);               // n(-108)
            this.N = false;
        }

        if (Vc) {
            // --- Fatal load error (Vc = true) ---
            // Draw a black full-screen rectangle + yellow/white error text
            Graphics g = this.getGraphics();
            if (g == null) return;
            g.translate(this.Eb, this.K);        // GameShell.Eb / K = inset offsets
            g.setColor(Color.black);
            g.fillRect(0, 0, 512, 356);

            g.setFont(new Font(STRINGS[477], Font.BOLD, 16)); // STRINGS[477]="Helvetica"
            g.setColor(Color.yellow);
            int y = 35;
            g.drawString(STRINGS[493], 30, y);   // "Error - unable to load game!"
            g.setColor(Color.white);
            y += 50;
            g.drawString(STRINGS[487], 30, y);   // "Please try the following:"
            y += 50;
            g.setFont(new Font(STRINGS[477], Font.BOLD, 12));
            g.drawString(STRINGS[484], 30, y);   // "1. Reload this page"
            y += 30;
            g.drawString(STRINGS[489], 30, y);   // "2. Clear your browser cache"
            y += 30;
            g.drawString(STRINGS[483], 30, y);   // "3. Restart your browser"
            y += 30;
            g.drawString(STRINGS[486], 30, y);   // "4. Reinstall Java"
            y += 30;
            g.drawString(STRINGS[490], 30, y);   // "If problems persist…"
            drawProgressBar(1, (byte)126);       // a(1, 126) — show 1-step progress bar
            return;
        }

        if (Xh) {
            // --- Domain-lock screen (Xh = true): "You must play from runescape.com" ---
            Graphics g = this.getGraphics();
            if (g == null) return;
            g.translate(this.Eb, this.K);
            g.setColor(Color.black);
            g.fillRect(0, 0, 512, 356);
            g.setFont(new Font(STRINGS[477], Font.BOLD, 20));
            g.setColor(Color.white);
            g.drawString(STRINGS[485], 50, 50);  // "This game must be played from..."
            g.drawString(STRINGS[492], 50, 100); // "www.runescape.com"
            g.drawString(STRINGS[495], 50, 150); // "Please visit that page."
            drawProgressBar(1, (byte)111);
        } else if (Ue) {
            // --- Out-of-memory screen (Ue = true) ---
            Graphics g = this.getGraphics();
            if (g != null) {
                g.translate(this.Eb, this.K);
                g.setColor(Color.black);
                g.fillRect(0, 0, 512, 356);
                g.setFont(new Font(STRINGS[477], Font.BOLD, 20));
                g.setColor(Color.white);
                g.drawString(STRINGS[482], 50, 50);  // "Error - out of memory!"
                g.drawString(STRINGS[488], 50, 100); // "Please try..."
                g.drawString(STRINGS[494], 50, 150); // "...increasing your Java heap"
                g.drawString(STRINGS[491], 50, 200); // "Contact support if..."
                drawProgressBar(1, (byte)106);
            }
        } else {
            // --- Normal paint / game tick ---
            if (clearDomainLock) {
                Xh = false;
            }
            if (surface == null) {
                return;
            }
            // Render the minimap while NOT sleeping
            if (screenMode != -1) {              // qg != -1
                surface.minimap = false;         // li.xb = false
                drawMinimap(2540);               // k(2540)
            }
            // Render main game frame
            if (screenMode == 1) {               // qg == 1 → in-game
                return;
            }
            try {
                surface.minimap = true;
                drawGame(13);                    // f(13) → drawGameFrame
            } catch (OutOfMemoryError oom) {
                Ue = true;
            }
        }
    }

    // -------------------------------------------------------------------------
    /** String pool XOR decoder stage 1: if len < 2, XOR the single char with '~' (0x7E).
     *  obf: static char[] z(String)  (no label) */
    private static char[] xorDecode1(String s) {
        char[] chars = s.toCharArray();
        if (chars.length < 2) {
            chars[0] = (char)(chars[0] ^ '~'); // single-char strings use a simple XOR key
        }
        return chars;
    }

    // -------------------------------------------------------------------------
    /** String pool XOR decoder stage 2: apply a 5-byte rotating XOR key {34,7,117,116,126} to each char.
     *  obf: static String z(char[])  (no label) */
    private static String xorDecode2(char[] chars) {
        // Key bytes cycle mod 5: 34('"'), 7('\a'), 117('u'), 116('t'), 126('~')
        for (int i = 0; i < chars.length; i++) {
            byte key;
            switch (i % 5) {
                case 0: key = 34;  break;  // '"'
                case 1: key = 7;   break;  // BEL
                case 2: key = 117; break;  // 'u'
                case 3: key = 116; break;  // 't'
                default: key = 126; break; // '~'
            }
            chars[i] = (char)(chars[i] ^ key);
        }
        return new String(chars).intern();
    }
