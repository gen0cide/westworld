package client.scene;

import client.world.WorldEntity; // obf: w  — per-visible-polygon entry (oracle Polygon)

// View-frustum AABB accumulators are scattered statics on these obf classes; their deob
// names are UNIFIED here with GameModel.project()'s reads (SIG-DRIFT fix, RENDER_DEOB_GAPS):
//   da.K -> ClientStream.frustumNearZ   m.j  -> SocketFactory.frustumFarZ
//   oa.b -> NameHash.frustumMinX        aa.f -> BZip.frustumMaxX
//   nb.y -> DataStore.frustumMinY       aa.b -> BZip.frustumMaxY
import client.net.ClientStream;   // da  — frustumNearZ (+ imageHeight, imageProducer, rgb)
import client.net.SocketFactory;  // m   — frustumFarZ  (+ colorModel = m.d, font glyphs = m.b)
import client.data.NameHash;      // oa  — frustumMinX
import client.data.BZip;          // aa  — frustumMaxX / frustumMaxY
import client.data.DataStore;     // nb  — frustumMinY

// Other scattered statics / helpers the render path reads (OBF-DRIFT reconciliation):
import client.world.World;        // k   — texture-load seq (k.e) + surface/BMP width (k.o)
import client.net.StreamBase;     // ib  — bitwiseAnd(int,int) colour-clamp helper used in ramp build
import client.net.StringCodec;    // u   — active ImageConsumer (u.d = DEAD_IMAGE_CONSUMER)
import client.util.LinkedQueue;   // db  — shared 17691-byte scratch (db.i = sharedByteArray)
import client.shell.GameShell;    // e   — 512-entry sin/cos table (e.nb); see FLAG note in render init

// Scattered per-row span writers the rasteriser dispatches to (OBF-DRIFT reconciliation):
import client.ui.MessageList;        // wb  — drawTextureSpan (16-arg, back-transparent translucent)
import client.net.ProxySocketFactory;// gb  — _junkRasterScanline (15-arg, translucent)
import client.data.CacheUpdater;     // cb  — drawTexturedScanlinePerspective (16-arg, opaque)
import client.net.DownloadWorker;    // jb  — drawTexturedSpanUnrolled (15-arg, translucent-model)
import client.util.Timer;            // p   — renderAffineSpan (15-arg, opaque not back-transparent)
import client.data.EntityDef;        // t   — fillPixelColumns16 (8-arg opaque gradient)

/**
 * Deobfuscation of obfuscated class {@code lb} == <b>Scene</b> (oracle
 * {@code mudclient204/src/Scene.java}), the 3D scene renderer.
 *
 * <p>The scene owns a list of {@link GameModel}s (obf {@code ca}), a per-frame list of visible
 * polygons ({@link WorldEntity}, obf {@code w}, oracle {@code Polygon}) and a per-row span table of
 * {@code Scanline}s (element type obf {@code n}). Each frame {@link #render} rebuilds the world-space
 * view frustum, projects every model, gathers the visible faces, depth-sorts them (painter's
 * algorithm: {@link #polygonsQSort} then {@link #polygonsIntersectSort}/{@link #reorderRange}), and
 * rasterises each face — textured or flat-gradient — into the {@link Surface} pixel buffer, also
 * resolving mouse picks.
 *
 * <h2>Identity / history</h2>
 * An earlier pass mislabelled {@code lb} as "World" and wrote a (partly reconstructed, often summarised)
 * deob to {@code src/client/world/World.java}. Per {@code docs/NAMING.md} the correct mapping is
 * <b>{@code lb} = Scene</b> (and {@code k} = World). This file is the authoritative Scene, rewritten
 * faithfully from the clean Vineflower base {@code decompiled/normalized-clean/lb.java}. Several field
 * roles that the old file got wrong are corrected here (see the audit summary):
 * {@code zb}=visiblePolygonCount (was "mousePickedCount"); {@code cc}=mousePickedCount (was "cameraX");
 * {@code cb}=textureCount (was "spriteCount"); {@code n}=spriteCount (was "visiblePolygonCount");
 * {@code Xb}/{@code Cb}=scanline minY/maxY (was "gradientUsed"/"textureCount"); {@code e}/{@code eb}=
 * newStart/newEnd overlap cursors (was "minY"/"maxY"); {@code j}/{@code Wb}=mouseX/mouseY (were camera
 * fields); {@code I}/{@code o}/{@code bc}=camera Z/Y/X look offset (one was "lastVisiblePolygonCount").
 *
 * <h2>Cross-class statics (scattered by the obfuscator onto unrelated classes)</h2>
 * <ul>
 *   <li>{@code e.nb} — 512-entry sin/cos table (1.15 fixed point) on {@code e} (GameShell).</li>
 *   <li>{@code ba.cc} — 2048-entry sin/cos table on {@code ba} (SurfaceSprite).</li>
 *   <li>{@code k.e} (long) — global texture-load sequence stamp on {@code k} (World).</li>
 *   <li>{@code k.o} (int) — image width, on {@code k} (World).</li>
 *   <li>{@code db.i} (byte[]) — shared 17691-byte scratch on {@code db} (LinkedQueue).</li>
 *   <li>view-frustum AABB accumulators: {@code aa.b},{@code nb.y},{@code m.j},{@code da.K},{@code oa.b},
 *       {@code aa.f}. Reset to 0 at the top of {@link #render}, translated by the camera position, and
 *       consumed by {@link GameModel#project}'s frustum test.</li>
 *   <li>{@code u.d} — Surface's AWT {@code ImageConsumer}; {@code da.bb} image height; {@code m.d}
 *       colour model (used by {@link #flushToImage}).</li>
 *   <li>{@code ib.a(int,int)} — colour-clamp helper used while building texture shade ramps.</li>
 * </ul>
 *
 * <h2>Obfuscation stripped</h2>
 * Opaque predicate {@code boolean bl = client.vh} (always false) and its dead branches; per-method
 * profiling {@code ++<counter>}; {@code try{…}catch(RuntimeException e){throw i.a(e,"sig")}} wrappers;
 * anti-tamper dummy params / {@code if(p!=MAGIC)return} guards / junk modulo expressions; junk masks
 * before shifts. The {@code ~x > ~y} idiom is normalised back to {@code x < y}. Fixed point: {@code >>15}
 * for the sin/cos tables (1.15), {@code <<8 / >>8} (8.8) for scanline edge stepping, and the
 * {@code <<12}/shift family for perspective-correct texturing.
 */
public final class Scene { // obf: lb

    // ------------------------------------------------------------------
    // Constants
    // ------------------------------------------------------------------

    /** Magic colour meaning "transparent / recompute per-vertex". obf: literal 12345678. */
    public static final int COLOUR_TRANSPARENT = 12345678;

    // ------------------------------------------------------------------
    // Profiling / diagnostics (instrumentation only; no gameplay effect)
    // ------------------------------------------------------------------

    /**
     * One static counter per method, bumped on entry by the profiler. Pure instrumentation.
     * obf: lb,Jb,rb,z,ub,m,F,V,d,mb,w,c,l,fb,Y,U,Sb,s,O,sb,k,W,q,Yb,Rb,bb,C,tb,hb,Db,t,M,Bb,p,Lb,E,Gb,Pb
     */
    static int prof_lb, prof_Jb, prof_rb, prof_z, prof_ub, prof_m, prof_F, prof_V, prof_d, prof_mb,
            prof_w, prof_c, prof_l, prof_fb, prof_Y, prof_U, prof_Sb, prof_s, prof_O, prof_sb,
            prof_k, prof_W, prof_q, prof_Yb, prof_Rb, prof_bb, prof_C, prof_tb, prof_hb, prof_Db,
            prof_t, prof_M, prof_Bb, prof_p, prof_Lb, prof_E, prof_Gb, prof_Pb;

    /** Scratch int array used by the profiler/error subsystem. obf: Tb. */
    public static int[] diagScratch; // obf: Tb
    /** Unused diagnostic string slot. obf: ac. */
    public static String[] diagStrings; // obf: ac

    /**
     * XOR-decoded diagnostic string pool (used only by the stripped error wrappers). obf: N.
     * Index 0 = "{...}", 1 = "null", 7 = "Warning tried to add null object!"; the rest are "lb.X("
     * method-signature prefixes. Decoded by the two {@code z(...)} helpers.
     */
    private static final String[] DIAG = decodeStringPool();

    // ------------------------------------------------------------------
    // Camera look-offset (rotated eye vector, in world space)
    // ------------------------------------------------------------------
    // The camera position passed to GameModel.project each frame, and the values that translate the
    // frustum AABB. Set by setCameraOrientation (the a(int×8) variant).

    private int cameraLookZ;   // obf: I   (project: cameraZ arg; frustum: da.K/m.j += I)
    private int cameraLookY;   // obf: o   (project: cameraY arg; frustum: nb.y/aa.b += o)
    private int cameraLookX;   // obf: bc  (project: cameraX arg; frustum: aa.f/oa.b += bc)

    // ------------------------------------------------------------------
    // Camera orientation
    // ------------------------------------------------------------------
    // NOTE: this rev has no separate camera-position fields; obf cc/Wb are reused as mousePickedCount
    // and mouseY (declared in the mouse-picking section), and obf j as mouseX. The look offset above
    // (I/o/bc) carries the camera position into GameModel.project.

    private int cameraYaw;     // obf: b   (project: cameraYaw arg)
    private int cameraPitch;   // obf: Kb  (project: cameraPitch arg)
    private int cameraRoll;    // obf: xb  (project: cameraRoll arg)
    private int cameraClipNear;// obf: nb  (=5)   near clip Z, passed to project

    private int baseX;            // obf: Zb  (screen origin X = 256)
    private int baseY;            // obf: Nb  (screen origin Y = 256)
    private int clipX;            // obf: A   (half view width)
    private int clipY;            // obf: wb  (half view height)
    private int width;            // obf: vb  (framebuffer stride)
    private int viewDistance;     // obf: R   (projection shift, default 8)
    private int normalMagnitude;  // obf: h   (face-normal scale, default 4)

    public int clipFar3d;         // obf: Mb  (=1000) far clip Z for 3D faces
    public int clipFar2d;         // obf: X   (=1000) far clip Z for 2D sprites
    public int fogZFalloff;       // obf: P   (=20)
    public int fogZDistance;      // obf: G   (=10)
    private boolean wideBand;     // obf: Ub  (=false)
    private boolean interlace;    // obf: f   (mirrors Surface.interlace each frame)

    // ------------------------------------------------------------------
    // Model list (the scene graph)
    // ------------------------------------------------------------------

    private Surface surface;      // obf: dc  render target
    private int[] raster;         // obf: pb  == surface pixel buffer
    private int maxModelCount;    // obf: u
    private int modelCount;       // obf: ab
    private GameModel[] models;   // obf: Z
    private int[] modelState;     // obf: jb  (set-only state per model)
    public GameModel view;        // obf: T   2D billboard/sprite model (widened: read cross-package by Mudclient)

    // ------------------------------------------------------------------
    // Visible-polygon list (built each frame, depth-sorted & drawn)
    // ------------------------------------------------------------------

    private int visiblePolygonCount;        // obf: zb
    private WorldEntity[] visiblePolygons;  // obf: y  (oracle Polygon[])
    private int lastVisiblePolygonCount;    // obf: (no persistent obf field; oracle had one)

    // ------------------------------------------------------------------
    // Sprite (2D billboard) table
    // ------------------------------------------------------------------

    private int spriteCount;          // obf: n   (count of billboard faces / view-model parallel index)
    private int[] spriteId;           // obf: gb
    private int[] spriteX;            // obf: Ob
    private int[] spriteZ;            // obf: Fb
    private int[] spriteY;            // obf: a
    private int[] spriteWidth;        // obf: ob
    private int[] spriteHeight;       // obf: Eb
    private int[] spriteTranslateX;   // obf: Q

    // ------------------------------------------------------------------
    // Mouse picking
    // ------------------------------------------------------------------

    private boolean mousePickingActive;    // obf: K
    private int mouseX;                     // obf: j   (= screenX - baseX)
    private int mouseY;                     // obf: Wb  (storage shared with cameraZ; see setMouseLoc)
    private int mousePickedCount;           // obf: cc  (storage shared with cameraX; see setMouseLoc)
    private int mousePickedMax;             // obf: db  (=100)
    private GameModel[] mousePickedModels;  // obf: Ab
    private int[] mousePickedFaces;         // obf: qb

    // ------------------------------------------------------------------
    // Scanline rasterizer working buffers
    // ------------------------------------------------------------------

    private Scanline[] scanlines;     // obf: x   (element type obf 'n')
    private int minY;                 // obf: Xb  (current span top row)
    private int maxY;                 // obf: Cb  (current span bottom row)
    private int newStart;             // obf: e   (overlap-sort window start cursor)
    private int newEnd;               // obf: eb  (overlap-sort window end cursor)
    // Texture-coordinate buffers (one per face vertex; model-space projected coords):
    private int[] projVertX;          // obf: Qb  (== ca.cc, projectVertexX) texture U source
    private int[] projVertY;          // obf: Vb  (== ca.H,  projectVertexY) texture V source
    private int[] projVertZ;          // obf: J   (== ca.bb, projectVertexZ) texture W source
    // Clipped screen-space buffers (one per clipped vertex after near-plane clip):
    private int[] screenX;            // obf: yb  (== ca.pb, vertexViewX)
    private int[] screenY;            // obf: B   (== ca.Ob, vertexViewY)
    private int[] shadeBuf;           // obf: r   (per-clipped-vertex shade, + fog)

    // ------------------------------------------------------------------
    // Gradient (flat-shaded colour ramp) cache
    // ------------------------------------------------------------------

    private int rampCount;            // obf: ib  (=50)
    private int[] gradientBase;       // obf: v
    private int[][] gradientRamps;    // obf: Ib  (rampCount x 256)
    private int[] currentRamp;        // obf: H   (last ramp resolved by rasterize)

    // ------------------------------------------------------------------
    // Texture cache
    // ------------------------------------------------------------------

    private int textureCount;             // obf: cb
    private byte[][] textureColoursUsed;  // obf: g   per-texel palette index
    private int[][] textureColourList;    // obf: L   per-texture palette (RGB)
    private int[] textureDimension;       // obf: Hb  (0 => 64px wide, 1 => 128px wide)
    private int[][] texturePixels;        // obf: kb  prepared RGB pixels (+ 3 shade copies)
    private long[] textureLoadedNumber;   // obf: D   LRU sequence stamp per texture
    private boolean[] textureBackTransparent; // obf: S
    private int[][] textureColours128;    // obf: ec  free pool of 128px (65536-int) buffers
    private int[][] textureColours64;     // obf: i   free pool of 64px  (16384-int) buffers

    // ==================================================================
    // Construction
    // ==================================================================

    /**
     * Builds the scene renderer. obf: {@code lb(ua,int,int,int)}.
     *
     * @param surface     render target (provides the pixel buffer + dimensions). obf var1
     * @param maxModels   capacity of the model list.                              obf var2
     * @param maxPolygons capacity of the visible-polygon list.                    obf var3
     * @param maxSprites  capacity of the 2D sprite/billboard table.               obf var4
     */
    public Scene(Surface surface, int maxModels, int maxPolygons, int maxSprites) {
        // --- scalar defaults (identical to oracle Scene constructor) ---
        this.cameraClipNear = 5;     // nb
        this.rampCount = 50;         // ib
        this.fogZDistance = 10;      // G
        this.interlace = false;      // f
        this.clipY = 192;            // wb (overwritten from surface below)
        this.baseX = 256;            // Zb
        this.fogZFalloff = 20;       // P
        this.baseY = 256;            // Nb
        this.clipFar3d = 1000;       // Mb
        this.width = 512;            // vb
        this.mousePickingActive = false; // K
        this.clipX = 256;            // A (overwritten from surface below)
        this.spriteCount = 0;        // n
        this.normalMagnitude = 4;    // h
        this.wideBand = false;       // Ub
        this.clipFar2d = 1000;       // X
        this.mousePickedMax = 100;   // db
        this.viewDistance = 8;       // R
        this.modelCount = 0;         // ab
        this.visiblePolygonCount = 0;// zb
        this.cameraLookY = 0;        // o
        this.cameraLookX = 0;        // bc
        this.cameraLookZ = 0;        // I
        this.mousePickedCount = 0;   // cc (a.k.a. cameraX storage; init 0)

        // --- fixed-size working buffers (40-wide clipped-edge scratch) ---
        this.projVertY = new int[40];     // Vb
        this.projVertZ = new int[40];     // J
        this.screenY = new int[40];       // B
        this.projVertX = new int[40];     // Qb
        this.screenX = new int[40];       // yb
        this.shadeBuf = new int[40];      // r
        this.gradientRamps = new int[this.rampCount][256]; // Ib (50 x 256)
        this.gradientBase = new int[this.rampCount];       // v
        this.mousePickedModels = new GameModel[this.mousePickedMax]; // Ab
        this.mousePickedFaces = new int[this.mousePickedMax];        // qb

        // --- model list / surface-derived state ---
        this.maxModelCount = maxModels; // u
        this.raster = surface.pixels;   // pb = surface pixel buffer (obf ua.rb)
        this.surface = surface;         // dc
        this.clipX = surface.width / 2; // A = surface.width/2  (obf ua.u)
        this.clipY = surface.height / 2;// wb = surface.height/2 (obf ua.k)
        this.models = new GameModel[this.maxModelCount]; // Z
        this.modelState = new int[this.maxModelCount];   // jb

        // --- visible-polygon list ---
        this.visiblePolygons = new WorldEntity[maxPolygons]; // y
        for (int p = 0; p < maxPolygons; p++) {
            this.visiblePolygons[p] = new WorldEntity();
        }

        // --- 2D billboard model + parallel tables ---
        this.view = new GameModel(2 * maxSprites, maxSprites); // T
        this.spriteWidth = new int[maxSprites];   // ob
        this.spriteHeight = new int[maxSprites];  // Eb
        this.spriteZ = new int[maxSprites];       // Fb
        this.spriteX = new int[maxSprites];       // Ob
        this.spriteTranslateX = new int[maxSprites]; // Q
        this.spriteId = new int[maxSprites];      // gb
        this.spriteY = new int[maxSprites];       // a

        // shared 17691-byte scratch on class db (LinkedQueue); lazy alloc (oracle aByteArray434).
        if (LinkedQueue.sharedByteArray == null) { // obf: LinkedQueue.sharedByteArray
            LinkedQueue.sharedByteArray = new byte[17691];
        }

        // --- 512-entry sin/cos table (1.15 fixed point), stored on class e (GameShell) ---
        for (int a = 0; a < 256; a++) {
            GameShell.colorScratch[a] = (int) (Math.sin(a * 0.02454369D) * 32768.0D);        // obf: GameShell.colorScratch
            GameShell.colorScratch[256 + a] = (int) (Math.cos(a * 0.02454369D) * 32768.0D);
        }
        // --- 2048-entry sin/cos table, stored on class ba (SurfaceSprite) ---
        for (int a = 0; a < 1024; a++) {
            SurfaceSprite.sin2048Cache[a] = (int) (Math.sin(a * 0.00613592315D) * 32768.0D);    // obf: SurfaceSprite.sin2048Cache
            SurfaceSprite.sin2048Cache[1024 + a] = (int) (Math.cos(a * 0.00613592315D) * 32768.0D);
        }
    }

    // ==================================================================
    // Model-list management
    // ==================================================================

    /** Appends a model to the scene. obf: a(ca var1, byte var2) — byte arg is an anti-tamper dummy. */
    public void addModel(GameModel model) {
        if (model == null) {
            System.out.println(DIAG[7]); // "Warning tried to add null object!"
        }
        if (this.modelCount < this.maxModelCount) {
            this.modelState[this.modelCount] = 0;
            this.models[this.modelCount++] = model;
        }
    }

    /** Removes the first occurrence of {@code model}, shifting the tail down. obf: a(ca var1, int var2). */
    public void removeModel(GameModel model) {
        for (int i = 0; i < this.modelCount; i++) {
            if (this.models[i] == model) {
                this.modelCount--;
                for (int j = i; j < this.modelCount; j++) {
                    this.models[j] = this.models[j + 1];
                    this.modelState[j] = this.modelState[j + 1];
                }
            }
        }
    }

    /** Clears the model list (nulls entries, resets count). obf: a(boolean var1) — bool arg is a dummy. */
    public void clearModels() {
        clearSprites(); // a(-118)
        for (int i = 0; i < this.modelCount; i++) {
            this.models[i] = null;
        }
        this.modelCount = 0;
    }

    // ==================================================================
    // Sprite (2D billboard) management
    // ==================================================================

    /** Empties the sprite list and resets the billboard model. obf: a(int var1) — arg gates body (<=-115). */
    private void clearSprites() {
        this.spriteCount = 0;
        this.view.clear(1); // T.c(1)  (soft clear)
    }

    /** Drops the last {@code count} sprites from the billboard model. obf: a(byte var1, int var2). */
    public void reduceSprites(int count) {
        this.spriteCount -= count;
        this.view.reduce(2 * count, -113, count); // T.b(2*count, -113, count)
        if (this.spriteCount < 0) {
            this.spriteCount = 0;
        }
    }

    /**
     * Adds a 2D billboard sprite to the {@link #view} model and parallel tables. obf: a(int×7,byte var8)
     * (counter {@code t}). The trailing {@code byte} (==109) gates the real body; otherwise returns -125.
     *
     * <p>Faithful field order from the clean base:
     * {@code gb[n]=var1; Ob[n]=var2; Fb[n]=var4; a[n]=var5; ob[n]=var6; Eb[n]=var7; Q[n]=0}; the two
     * view vertices are {@code (var2,var4,var5)} and {@code (var2,var4,var5-var7)}; face tag {@code var3}.
     *
     * @return the new sprite index, or -125 if the tamper-guard byte was wrong.
     */
    public int addSprite(int id, int x, int tag, int z, int y, int w, int h, byte guard) {
        // obf params: var1=id, var2=x, var3=tag, var4=z, var5=y, var6=w, var7=h
        this.spriteId[this.spriteCount] = id;     // gb[n]=var1
        this.spriteZ[this.spriteCount] = z;       // Fb[n]=var4
        this.spriteY[this.spriteCount] = y;       // a[n]=var5
        this.spriteX[this.spriteCount] = x;       // Ob[n]=var2
        this.spriteWidth[this.spriteCount] = w;   // ob[n]=var6
        this.spriteHeight[this.spriteCount] = h;  // Eb[n]=var7
        this.spriteTranslateX[this.spriteCount] = 0; // Q[n]=0
        if (guard != 109) {
            return -125;
        }
        // positional clean call: T.b(false, var2=x, var4=z, var5=y) — GameModel.createVertex(flag,z,x,y)
        int top = this.view.createVertex(false, x, z, y);          // T.b(false, var2, var4, var5)
        int bottom = this.view.createVertex(false, x, z, -h + y);  // T.b(false, var2, var4, var5-var7)
        int[] faceVerts = { top, bottom };
        this.view.createFace(2, faceVerts, 0, 0, false);  // T.a(2, faceVerts, 0,0,false)
        this.view.faceTag[this.spriteCount] = tag;        // T.E[n]=var3
        this.view.isLocalPlayer[this.spriteCount++] = 0;  // T.zb[n++]=0
        return this.spriteCount - 1;
    }

    /** Marks sprite {@code index} as belonging to the local player (excluded from picking). obf: c(int,int). */
    public void setLocalPlayer(int unused, int index) {
        this.view.isLocalPlayer[index] = 1; // T.zb[var2]=1
    }

    /** Sets the horizontal pick offset for sprite {@code index}. obf: b(int,int,int) — first arg dummy. */
    public void setSpriteTranslateX(int unused, int index, int translateX) {
        this.spriteTranslateX[index] = translateX; // Q[var2]=var3
    }

    // ==================================================================
    // Mouse picking + camera position
    // ==================================================================

    /**
     * Arms mouse picking for this frame. obf: a(int var1, int var2, int var3) (counter {@code p}).
     * In this rev the mouse-pick fields share storage with the simple-camera fields:
     * {@code K(active)=true; j(mouseX)= -baseX + var2; Wb(mouseY)=var3; cc(mousePickedCount)=var1}.
     * Callers pass {@code var1=0} to begin a pick pass at count 0.
     */
    public void setMouseLoc(int startCount, int screenX, int screenY) {
        this.mousePickingActive = true;        // K=true
        this.mouseX = -this.baseX + screenX;   // j = screenX - baseX
        this.mouseY = screenY;                 // Wb
        this.mousePickedCount = startCount;     // cc
    }

    /** @return number of faces picked this frame. obf: b(int var1) (counter {@code rb}) — returns {@code cc}. */
    public int getMousePickedCount(int unused) {
        return this.mousePickedCount; // cc
    }

    /** @return the picked-face index array. obf: a(byte var1) (counter {@code q}). */
    public int[] getMousePickedFaces() {
        return this.mousePickedFaces; // qb
    }

    /** @return the picked-model array. obf: b(byte var1) (counter {@code Jb}). */
    public GameModel[] getMousePickedModels() {
        return this.mousePickedModels; // Ab
    }

    /**
     * Sets the full camera transform: position plus yaw/pitch/roll, and (when {@code mode == -12349})
     * derives the rotated look offset stored in {@code I/o/bc}. obf: a(int×8) (counter {@code U}).
     *
     * <p>Stores the <em>inverse</em> angles {@code (1024 - angle) & 1023} into {@code b/Kb/xb} so the
     * projection can rotate world points into view space, then rotates the eye vector by yaw→pitch→roll
     * through the 2048-entry sin/cos table to produce the look offset.
     *
     * @param camX  obf var1.   @param camY obf var2.   @param eyeLen obf var3 (eye Z / look length).
     * @param yaw   obf var4.   @param mode obf var5 (-12349 gates the derivation).
     * @param pitch obf var6.   @param eyeY obf var7.   @param roll obf var8.
     */
    public void setCameraOrientation(int camX, int camY, int eyeLen, int yaw, int mode, int pitch, int eyeY, int roll) {
        roll &= 1023; yaw &= 1023; pitch &= 1023;
        this.cameraYaw   = (1024 - roll)  & 1023; // b   = 1024 - var8(roll)
        this.cameraPitch = (1024 - yaw)   & 1023; // Kb  = 1024 - var4(yaw)
        this.cameraRoll  = (1024 - pitch) & 1023; // xb  = 1024 - var6(pitch)
        int rx = 0; // rotated eye X
        if (mode == -12349) {
            int ry = 0;        // eye Y
            int rz = eyeLen;   // eye Z = look length (clean: var11=eyeLen)
            if (yaw != 0) { // rotate about Y (yaw)
                int sin = SurfaceSprite.sin2048Cache[yaw];
                int cos = SurfaceSprite.sin2048Cache[yaw + 1024];
                int t = (cos * ry - sin * rz) >> 15;
                rz = (sin * ry + rz * cos) >> 15;
                ry = t;
            }
            if (pitch != 0) { // rotate about X (pitch)
                int sin = SurfaceSprite.sin2048Cache[pitch];
                int cos = SurfaceSprite.sin2048Cache[pitch + 1024];
                int t = (rx * cos + rz * sin) >> 15;
                rz = (cos * rz - sin * rx) >> 15;
                rx = t;
            }
            if (roll != 0) { // rotate about Z (roll)
                int cos = SurfaceSprite.sin2048Cache[roll + 1024];
                int sin = SurfaceSprite.sin2048Cache[roll];
                int t = (rx * cos + ry * sin) >> 15;
                ry = (-(sin * rx) + ry * cos) >> 15;
                rx = t;
            }
            this.cameraLookZ = -rz + camY; // I  = -var11 + var2(camY)
            this.cameraLookY = eyeY - ry;  // o  = var7(eyeY) - var10(ry)
            this.cameraLookX = camX - rx;  // bc = var1(camX) - var9(rx)
        }
    }

    /**
     * Sets the screen origin, clip extents, stride and projection shift, and (re)allocates the
     * per-scanline buffer. obf: a(int,boolean,int,int,int,int,int) (counter {@code fb}) — bool is a dummy.
     * Faithful param order from the clean base: {@code R=var6; Zb=var7; vb=var3; Nb=var5; wb=var1;
     * A=var4; x = new Scanline[var1 + var5]}.
     */
    public void setBounds(int clipY, boolean ignored, int width, int clipX, int baseY, int viewDistance, int baseX) {
        this.viewDistance = viewDistance; // R = var6
        this.baseX = baseX;               // Zb = var7
        this.width = width;               // vb = var3
        this.baseY = baseY;               // Nb = var5
        this.clipY = clipY;               // wb = var1
        this.clipX = clipX;               // A = var4
        int rows = clipY + baseY;         // x = new n[var1 + var5]
        this.scanlines = new Scanline[rows];
        for (int s = 0; s < rows; s++) {
            this.scanlines[s] = new Scanline();
        }
    }

    // ==================================================================
    // Frustum culling
    // ==================================================================

    /**
     * Rotates the view-space corner ({@code x},{@code y},{@code z}) by the inverse camera orientation and
     * expands the world-space frustum AABB. obf: a(int,int,int,boolean) (counter {@code M}). The AABB
     * bounds are scattered static ints on other classes; clean assignments reproduced verbatim.
     */
    private void expandFrustum(int x, int y, int z, boolean ignored) {
        // clean: var5 = (1024 - Kb) & 1023 = invPitch; var6 = (1024 - xb) & 1023 = invRoll;
        //        var7 = (1024 - b) & 1023 = invYaw. Rotations applied in the order yaw, pitch, roll.
        int invPitch = (1024 - this.cameraPitch) & 1023; // var5 = 1024 - Kb
        int invRoll  = (1024 - this.cameraRoll)  & 1023; // var6 = 1024 - xb
        int invYaw   = (1024 - this.cameraYaw)   & 1023; // var7 = 1024 - b

        if (invYaw != 0) { // rotate (y,z) about the yaw axis (clean var7 block on var2/var3)
            int sin = SurfaceSprite.sin2048Cache[invYaw];
            int cos = SurfaceSprite.sin2048Cache[1024 + invYaw];
            int t = (cos * y + sin * z) >> 15;
            z = (-(sin * y) + z * cos) >> 15;
            y = t;
        }
        if (invPitch != 0) { // rotate (x,z) about the pitch axis (clean var5 block on var1/var3)
            int sin = SurfaceSprite.sin2048Cache[invPitch];
            int cos = SurfaceSprite.sin2048Cache[1024 + invPitch];
            int t = (-(sin * x) + z * cos) >> 15; // clean var16 = -(sin*x) + z*cos
            x = (cos * x + sin * z) >> 15;
            z = t;
        }
        if (invRoll != 0) { // rotate (x,y) about the roll axis (clean var6 block on var1/var2)
            int sin = SurfaceSprite.sin2048Cache[invRoll];
            int cos = SurfaceSprite.sin2048Cache[1024 + invRoll];
            int t = (sin * x + y * cos) >> 15; // clean var17 = sin*x + y*cos
            x = (cos * x - sin * y) >> 15;
            y = t;
        }
        // expand the six scattered frustum accumulators (clean polarity decoded from the ~-idiom):
        if (ClientStream.frustumNearZ < x) ClientStream.frustumNearZ = x; // ClientStream.frustumNearZ = max(x)        (clean: if (ClientStream.frustumNearZ < var1) ClientStream.frustumNearZ = var1)
        if (z < BZip.frustumMaxY) BZip.frustumMaxY = z; // BZip.frustumMaxY = min(z)        (clean: if (~var3 > ~BZip.frustumMaxY) -> var3 < BZip.frustumMaxY)
        if (y > NameHash.frustumMinX) NameHash.frustumMinX = y; // NameHash.frustumMinX = max(y)        (clean: if (~NameHash.frustumMinX > ~var2) -> NameHash.frustumMinX < var2)
        if (z > DataStore.frustumMinY) DataStore.frustumMinY = z; // DataStore.frustumMinY = max(z)        (clean: if (~DataStore.frustumMinY > ~var3) -> DataStore.frustumMinY < var3)
        if (x < SocketFactory.frustumFarZ)  SocketFactory.frustumFarZ = x;  // SocketFactory.frustumFarZ  = min(x)        (clean: if (var1 < SocketFactory.frustumFarZ) SocketFactory.frustumFarZ = var1)
        if (y < BZip.frustumMaxX) BZip.frustumMaxX = y; // BZip.frustumMaxX = min(y)        (clean: if (var2 < BZip.frustumMaxX) BZip.frustumMaxX = var2)
    }

    // ==================================================================
    // Per-model lighting
    // ==================================================================

    /**
     * Sets ambient/diffuse + light direction on every model. obf: a(int,int,int,int,int,int)
     * (counter {@code mb}; oracle {@code setLight(i,j,k,l,i1)}). Defaults {@code var4} (the X direction)
     * to 32 when {@code (var4==0 && var6==0 && var1==0)}, then calls
     * {@code GameModel.setLight(ambient,diffuse,dirY,guard,dirX,dirZ)} on each model.
     *
     * @param var1 dirZ (obf var1).   @param var2 ambient (obf var2).   @param startIndex unused start
     *             cursor in this rev (the loop runs over all models). @param var4 dirX (obf var4).
     * @param var5 diffuse (obf var5).@param var6 dirY (obf var6).
     */
    public void setLightFull(int var1, int var2, int startIndex, int var4, int var5, int var6) {
        if (var4 == 0 && var6 == 0 && var1 == 0) { // clean: ~var4==-1 && var6==0 && var1==0
            var4 = 32;
        }
        for (int i = startIndex; i < this.modelCount; i++) {
            // ca.setLight(ambient=var2, diffuse=var5, dirY=var6, guard=-115, dirX=var4, dirZ=var1)
            this.models[i].setLight(var2, var5, var6, (byte) -115, var4, var1);
        }
    }

    /**
     * Sets just the light direction on every model. obf: a(int,int,boolean,int) (counter {@code Db};
     * oracle {@code setLight(i,j,k)}). Defaults {@code var4} (X) to 32 when {@code (var4==0 && var2==0 &&
     * var1==0)}, optionally rebuilds the scene (anti-tamper), then calls {@code GameModel.setLightDirection}.
     *
     * @param var1 dirZ. @param var2 dirY. @param ignored anti-tamper bool. @param var4 dirX.
     */
    public void setLight(int var1, int var2, boolean ignored, int var4) {
        if (var4 == 0 && var2 == 0 && var1 == 0) { // clean: ~var4==-1 && var2==0 && var1==0
            var4 = 32;
        }
        if (!ignored) {
            render(-89); // obf: lb.c(I)  (anti-tamper side effect, harmless)
        }
        for (int i = 0; i < this.modelCount; i++) {
            // ca.setLightDirection(false, dirX=var4, dirY=var2, dirZ=var1)
            this.models[i].setLightDirection(false, var4, var2, var1);
        }
    }

    // ==================================================================
    // Colour / gradient helpers
    // ==================================================================

    /**
     * Linear interpolation used while building polygon edges. obf: a(int,boolean,int,int,int,int)
     * (counter {@code hb}; oracle {@code method306}). Returns {@code var6} when {@code var4==var1}, else
     * {@code (var5-var6)*(var3-var1)/(var4-var1)+var6}. The boolean is an anti-tamper dummy.
     */
    private int interpolate(int var1, boolean ignored, int var3, int var4, int var5, int var6) {
        return var4 == var1 ? var6 : (-var6 + var5) * (-var1 + var3) / (-var1 + var4) + var6;
    }

    /**
     * Returns the flat RGB colour for a face fill id. obf: a(int,boolean) (counter {@code ub}).
     * For {@code COLOUR_TRANSPARENT} returns 0; for a real texture prepares it and returns its first
     * pixel; for a negative "direct RGB" id unpacks the 5/5/5 encoding (junk-masked shifts).
     */
    public int fillColour(int fillId, boolean prepare) {
        if (fillId == COLOUR_TRANSPARENT) {
            return 0;
        }
        prepareTexture(fillId, prepare); // b(var1, var2)
        if (fillId >= 0) { // (~var1 <= -1)
            return this.texturePixels[fillId][0]; // kb[var1][0]
        } else if (fillId < 0) {
            fillId = -(fillId + 1);
            int r = (fillId & 32041) >> 10; // junk-masked 5 bits
            int g = (fillId & 1012) >> 5;
            int b = fillId & 31;
            return (b << 3) + (g << 11) + (r << 19);
        }
        return 0;
    }

    // ==================================================================
    // Texture cache
    // ==================================================================

    /**
     * Allocates the texture-cache backing arrays for {@code count} textures, with separate free pools.
     * obf: a(int,int,int,int) (counter {@code sb}). Clean: {@code L,g,kb=new[var4][]; i=new[var3][];
     * S=new[var4]; World.e=var1; cb=var4; Hb=new[var4]; ec=new[var2][]; D=new long[var4]}.
     *
     * @param loadSeq  initial texture-load sequence (stored to {@code World.e}).                  obf var1
     * @param pool128  size of the 128px free pool {@code ec}.                                 obf var2
     * @param pool64   size of the 64px free pool {@code i}.                                   obf var3
     * @param count    number of textures.                                                     obf var4
     */
    public void allocateTextures(int loadSeq, int pool128, int pool64, int count) {
        this.textureColourList = new int[count][];        // L  = new[var4][]
        this.textureColoursUsed = new byte[count][];      // g  = new[var4][]
        this.texturePixels = new int[count][];            // kb = new[var4][]
        this.textureColours64 = new int[pool64][];        // i  = new[var3][]   (64px pool)
        this.textureBackTransparent = new boolean[count]; // S  = new[var4]
        World.e = (long) loadSeq;                              // World.e = var1  (texture load seq on World)
        this.textureCount = count;                        // cb = var4
        this.textureDimension = new int[count];           // Hb = new[var4]
        this.textureColours128 = new int[pool128][];      // ec = new[var2][]  (128px pool)
        this.textureLoadedNumber = new long[count];       // D  = new long[var4]
    }

    /**
     * Registers a texture's palette + texel-index data and prepares it. obf: a(int,byte,int[],int,byte[])
     * (counter {@code Rb}). The {@code byte} arg is an anti-tamper dummy.
     *
     * @param id          texture id.
     * @param colours     palette (RGB).             obf var3
     * @param dimension   0 => 64px wide, 1 => 128px.obf var4
     * @param coloursUsed per-texel palette indices. obf var5
     */
    public void defineTexture(int id, byte guard, int[] colours, int dimension, byte[] coloursUsed) {
        this.textureColoursUsed[id] = coloursUsed; // g[id]=var5
        this.textureColourList[id] = colours;      // L[id]=var3
        this.textureDimension[id] = dimension;     // Hb[id]=var4
        this.textureLoadedNumber[id] = 0L;         // D[id]=0
        this.textureBackTransparent[id] = false;   // S[id]=false
        this.texturePixels[id] = null;             // kb[id]=null
        prepareTexture(id, true);                  // b(id, true)
    }

    /**
     * Ensures texture {@code id} is resident, evicting the least-recently-used texture of the same
     * dimension class from the free pool if necessary, then materialises its pixels. obf: b(int,boolean)
     * (counter {@code W}). When {@code keepPick==false} clears {@code K} (anti-tamper side effect).
     *
     * <p>Clean dispatch: {@code Hb[id] != 0} -> 128px class draws a 65536-int buffer from the {@code ec}
     * pool; otherwise 64px class draws a 16384-int buffer from the {@code i} pool.
     */
    private void prepareTexture(int id, boolean keepPick) {
        if (!keepPick) {
            this.mousePickingActive = false; // K=false
        }
        if (id < 0) {
            return;
        }
        this.textureLoadedNumber[id] = World.e++; // D[id] = World.e++
        if (this.texturePixels[id] != null) {
            return;
        }
        if (this.textureDimension[id] != 0) {
            // 128px-wide class (Hb != 0): find a free 65536-int buffer in the ec pool, else evict LRU.
            for (int j = 0; j < this.textureColours128.length; j++) { // ec
                if (this.textureColours128[j] == null) {
                    this.textureColours128[j] = new int[65536];
                    this.texturePixels[id] = this.textureColours128[j];
                    buildTexturePixels(id, (byte) 118);
                    return;
                }
            }
            long best = 1073741824L;
            int victim = 0;
            for (int t = 0; t < this.textureCount; t++) { // cb
                if (id != t && this.textureDimension[t] == 1
                        && this.texturePixels[t] != null
                        && this.textureLoadedNumber[t] < best) {
                    best = this.textureLoadedNumber[t];
                    victim = t;
                }
            }
            this.texturePixels[id] = this.texturePixels[victim];
            this.texturePixels[victim] = null;
            buildTexturePixels(id, (byte) 118);
            return;
        }
        // 64px-wide class (Hb == 0): find a free 16384-int buffer in the i pool, else evict LRU.
        for (int j = 0; j < this.textureColours64.length; j++) { // i
            if (this.textureColours64[j] == null) {
                this.textureColours64[j] = new int[16384];
                this.texturePixels[id] = this.textureColours64[j];
                buildTexturePixels(id, (byte) 118);
                return;
            }
        }
        long best = 1073741824L;
        int victim = 0;
        for (int t = 0; t < this.textureCount; t++) {
            if (id != t && this.textureDimension[t] == 0
                    && this.texturePixels[t] != null
                    && this.textureLoadedNumber[t] < best) {
                best = this.textureLoadedNumber[t];
                victim = t;
            }
        }
        this.texturePixels[id] = this.texturePixels[victim];
        this.texturePixels[victim] = null;
        buildTexturePixels(id, (byte) 118);
    }

    /**
     * Materialises texture {@code id}'s RGB pixels from its palette + index data, applies the
     * "0xf800ff => back-transparent" convention, and appends three pre-darkened shade copies (subtract
     * 1/8, 1/4, 3/8 of each channel). obf: a(int,byte) (counter {@code tb}) — byte==118 gates the body.
     */
    private void buildTexturePixels(int id, byte guard118) {
        int dim = (this.textureDimension[id] != 0) ? 128 : 64;
        int[] out = this.texturePixels[id];
        int n = 0;
        if (guard118 == 118) {
            for (int x = 0; x < dim; x++) {
                for (int y = 0; y < dim; y++) {
                    int colour = this.textureColourList[id][this.textureColoursUsed[id][y + x * dim] & 0xFF];
                    colour &= 16316671; // 0xf8f8ff
                    if (colour == 0) {
                        colour = 1;
                    } else if (colour == 16253183) { // 0xf800ff (clean: ~colour == -16253184)
                        this.textureBackTransparent[id] = true;
                        colour = 0;
                    }
                    out[n++] = colour;
                }
            }
            for (int i = 0; i < n; i++) {
                int c = out[i];
                out[n + i]     = StreamBase.bitwiseAnd(c - (c >>> 3), 16316671);          // -1/8
                out[2 * n + i] = StreamBase.bitwiseAnd(c - (c >>> 2), 16316671);          // -1/4
                out[3 * n + i] = StreamBase.bitwiseAnd(c - (c >>> 3) - (c >>> 2), 16316671); // -3/8
            }
        }
    }

    /**
     * Animates a scrolling 64-wide texture by shifting rows up one step, then rebuilds the three darkened
     * shade copies. obf: d(int,int) (counter {@code d}) — first arg is an anti-tamper dummy.
     */
    public void scrollTexture(int guard, int id) {
        if (this.texturePixels[id] == null) {
            return;
        }
        int[] px = this.texturePixels[id];
        for (int col = 0; col < 64; col++) {
            int k = 4032 + col;
            int saved = px[k];
            for (int r = 0; r < 63; r++) {
                px[k] = px[k - 64];
                k -= 64;
            }
            this.texturePixels[id][k] = saved;
        }
        // anti-tamper: if (var1 != 25013) this.b = 60;  (dead store, omitted)
        int n = 4096;
        for (int i = 0; i < n; i++) {
            int c = px[i];
            px[n + i]     = StreamBase.bitwiseAnd(c - (c >>> 3), 16316671);
            px[2 * n + i] = StreamBase.bitwiseAnd(16316671, c - (c >>> 2));
            px[3 * n + i] = StreamBase.bitwiseAnd(16316671, -(c >>> 3) + c - (c >>> 2));
        }
    }

    // ==================================================================
    // Visible-polygon initialisation (3D faces and 2D sprites)
    // ==================================================================

    /**
     * Computes a 3D face's normal, back/front visibility sign, and projected screen AABB, storing them
     * into the {@link WorldEntity} at {@code polyIndex}. obf: a(int,int) (counter {@code l}). Two cross
     * products give the normal; the result is shifted to keep magnitudes in range (recording the shift in
     * {@code ca.M}); {@code w.s} holds the dot-product cull sign.
     */
    private void initialisePolygon3d(int polyIndex, int unused) {
        WorldEntity poly = this.visiblePolygons[polyIndex]; // y[var1]
        GameModel model = poly.model;                           // w.o (model)
        int face = poly.faceIndex;                                  // w.i (face)
        int[] verts = model.faceVertices[face];             // ca.o[face]
        int normalShift = model.normalScale[face];          // ca.M[face]

        int x0 = model.projectVertexX[verts[0]]; // ca.cc[v0]
        int y0 = model.projectVertexY[verts[0]]; // ca.H[v0]
        int z0 = model.projectVertexZ[verts[0]]; // ca.bb[v0]
        int ax = -x0 + model.projectVertexX[verts[1]];
        int ay = -y0 + model.projectVertexY[verts[1]];
        int az = -z0 + model.projectVertexZ[verts[1]];
        int bx = -x0 + model.projectVertexX[verts[2]];
        int by = -y0 + model.projectVertexY[verts[2]];
        int bz = -z0 + model.projectVertexZ[verts[2]];

        int nx = -(az * by) + bz * ay; // cross product (clean var19)
        int ny = -(ax * bz) + bx * az; // (clean var20)
        int nz = ax * by - bx * ay;    // (clean var21)

        if (normalShift == -1) {
            normalShift = 0;
            while (nx > 25000 || ny > 25000 || nz > 25000
                    || nx < -25000 || ny < -25000 || nz < -25000) {
                normalShift++;
                nx >>= 1; ny >>= 1; nz >>= 1;
            }
            model.normalScale[face] = normalShift; // ca.M[face]
            model.normalMagnitude[face] = (int) (this.normalMagnitude
                    * Math.sqrt((double) (nz * nz + ny * ny + nx * nx))); // ca.k[face]
        } else {
            nz >>= normalShift; nx >>= normalShift; ny >>= normalShift;
        }

        poly.normalZ = nz; // w.k = normalZ
        poly.normalX = nx; // w.r = normalX
        poly.normalY = ny; // w.l = normalY
        // w.s = nx*x0 - (-(ny*y0) + -(nz*z0))  (dot of normal with v0: cull sign)
        poly.normalDot = nx * x0 - (-(ny * y0) + -(nz * z0));

        // projected screen-space AABB over all face vertices (faithful min/max loop)
        int loZ, hiZ; loZ = hiZ = model.projectVertexZ[verts[0]]; // ca.bb ; clean var22=min,var23=max
        int hiX, loX; hiX = loX = model.vertexViewX[verts[0]];    // ca.pb ; clean var24=max,var25? -> see below
        int hiY, loY; hiY = loY = model.vertexViewY[verts[0]];    // ca.Ob
        for (int v = 1; v < verts.length; v++) {
            int z = model.projectVertexZ[verts[v]];
            if (z > hiZ) hiZ = z; else if (z < loZ) loZ = z;
            int vx = model.vertexViewX[verts[v]];
            if (vx > hiX) hiX = vx; else if (vx < loX) loX = vx;
            int vy = model.vertexViewY[verts[v]];
            if (vy > hiY) hiY = vy; else if (vy < loY) loY = vy;
        }
        // clean stores: w.e=var24(maxX? -> screen X bound), w.m=var25, w.j=var27, w.q=var23(maxZ),
        // w.h=var26, w.u=var22(minZ). Verified from b(int,int)/a(int,int) listings:
        poly.minX = loX;  // w.e  (min screen X)
        poly.maxX = hiX;  // w.m  (max screen X)
        poly.minZ = hiY;  // w.j  (max screen Y)
        poly.minDepth = hiZ;  // w.q  (max Z)
        poly.maxZ = loY;  // w.h  (min screen Y)
        poly.maxDepth = loZ;  // w.u  (min Z)
    }

    /**
     * Computes a 2D sprite face's projected screen AABB (using its billboard pair) and a fixed normal of
     * (0,0,1), storing them into the {@link WorldEntity} at {@code polyIndex}. obf: b(int,int)
     * (counter {@code s}) — first arg is dummy. Faithful to the clean base, including +/-20 Z padding.
     */
    private void initialisePolygon2d(int unused, int polyIndex) {
        WorldEntity poly = this.visiblePolygons[polyIndex]; // y[var2]
        GameModel model = poly.model;                           // w.o
        int face = poly.faceIndex;                                  // w.i
        int[] verts = model.faceVertices[face];             // ca.o[face]

        int cc0 = model.projectVertexX[verts[0]]; // var12 = cc[v0]
        int H0  = model.projectVertexY[verts[0]]; // var13 = H[v0]
        int bb0 = model.projectVertexZ[verts[0]]; // var14 = bb[v0]
        model.normalMagnitude[face] = 1; // ca.k[face]=1
        model.normalScale[face] = 0;     // ca.M[face]=0
        poly.normalY = 0; // w.l = var10 = 0  (normalY)
        poly.normalX = 0; // w.r = var9  = 0  (normalX)
        poly.normalZ = 1; // w.k = var11 = 1  (normalZ)
        poly.normalDot = bb0 * 1 + cc0 * 0 + H0 * 0; // w.s = var14*var11 + var12*var9 + var13*var10 = bb0

        // --- bb (Z) extents over the two billboard verts:  var15=max, var16=min ---
        int var15 = bb0, var16 = bb0;       // both seeded from bb[v0]
        // --- pb (screen X) extents:  var17=min, var18=max ---
        int var17 = model.vertexViewX[verts[0]], var18 = var17; // seeded from pb[v0]
        if (var17 <= model.vertexViewX[verts[1]]) { // clean: ~var17 >= ~pb[v1]
            var18 = model.vertexViewX[verts[1]];     // var18 = max(pb)
        } else {
            var17 = model.vertexViewX[verts[1]];     // var17 = min(pb)
        }
        // refine bb extents against v1 (clean label50):
        int bb1 = model.projectVertexZ[verts[1]];
        if (var16 >= bb1) { if (var15 > bb1) var15 = bb1; } else { var16 = bb1; }
        // (clean label43 re-clamps pb min/max against pb[v1]; equivalent to the above pb split)
        int pb1 = model.vertexViewX[verts[1]];
        if (var18 >= pb1) { if (var17 > pb1) var17 = pb1; } else { var18 = pb1; }

        // --- Ob (screen Y) extents:  var20 from Ob[v0], var19 from Ob[v1] (clean label36) ---
        int var19 = model.vertexViewY[verts[1]]; // Ob[v1]
        int var20 = model.vertexViewY[verts[0]]; // Ob[v0]
        int Ob1 = model.vertexViewY[verts[1]];
        poly.maxX = var18 + 20; // w.m = max(pb) + 20
        poly.minX = var17 - 20; // w.e = min(pb) - 20
        if (var20 < Ob1) var20 = Ob1;       // var20 = max(Ob[v0], Ob[v1])
        if (var19 > Ob1) var19 = Ob1;       // var19 (clean ~var19 < ~Ob1 == var19 > Ob1)

        poly.minDepth = var16; // w.q  (min bb)
        poly.maxDepth = var15; // w.u  (max bb)
        poly.minZ = var20; // w.j  (max Ob)
        poly.maxZ = var19; // w.h  (Ob[v1])
    }

    // ==================================================================
    // Depth & overlap sorting (painter's algorithm)
    // ==================================================================

    /**
     * Quicksort of the visible-polygon array by {@code depth} ({@code w.t}). obf: a(int,int,w[],int)
     * (counter {@code Pb}). Two of the four ints are anti-tamper dummies; the real bounds are {@code low}
     * and {@code high}. Partition matches the clean base (advance {@code min} while depth&lt;=pivot;
     * retreat {@code max} while depth&gt;pivot — see notes). Mirrors oracle {@code polygonsQSort}.
     */
    private void polygonsQSort(int low, int unused, WorldEntity[] polys, int high) {
        if (high > low) {
            int min = low - 1;
            int max = high + 1;
            int mid = (high + low) / 2;
            WorldEntity pivot = polys[mid];
            polys[mid] = polys[low];
            polys[low] = pivot;
            int pivotDepth = pivot.sortDepth; // w.t
            while (max > min) {
                // The min-advance comparison must be `> pivot`, matching BOTH ground truths:
                //   - clean lb.java:187-198 (decompiled flow): L_A re-iterates `++var5` while the
                //     innermost test `if (var3[var5].t > var9) break` is true, i.e. min advances while
                //     depth > pivot and stops at the first depth <= pivot.
                //   - OpenRSC three/Scene.java:1358-1360 (same algorithm, readable):
                //       do { ++var5; } while (var3[var5].m_t > var9);   // advance min while depth > pivot
                //       do { --var6; } while (var9 > var3[var6].m_t);   // retreat max while pivot > depth
                // i.e. min stops at depth <= pivot, max stops at depth >= pivot — a DESCENDING partition.
                // A prior pass inverted the min scan to `<= pivot` on a misread of the clean control flow
                // (and a non-existent "OpenRSC mirror"); that broke the painter order. Restored per-site
                // to the ground truth. The pivot at index `low` (polys[low].t == pivot) is the sentinel.
                // Cite: clean lb.java:188/196 (min) + 201/208 (max); OpenRSC Scene.java:1358-1364.
                do { min++; } while (polys[min].sortDepth > pivotDepth);
                do { max--; } while (pivotDepth > polys[max].sortDepth);
                if (min < max) {
                    WorldEntity tmp = polys[min];
                    polys[min] = polys[max];
                    polys[max] = tmp;
                }
            }
            polygonsQSort(low, -1, polys, max);
            polygonsQSort(max + 1, -1, polys, high);
        }
    }

    /**
     * Overlap-resolution pass after the depth sort. obf: a(int,int,int,w[]) (counter {@code lb}). Mirrors
     * oracle {@code polygonsIntersectSort(step, polygons, count)} exactly. The {@code int} var3 is an
     * anti-tamper dummy (its {@code >= -50} guard fires a harmless side-effect).
     *
     * <p>For each unprocessed polygon (tracked by {@code w.c}=skipSomething), it scans the {@code step}-
     * deep window before it from far to near; when the window polygon's screen AABB overlaps and a
     * separating/heuristic test says the order is wrong ({@code !separatedOrInOrder} &amp;&amp;
     * {@code faceOrders}), it slides the polygon forward via {@link #reorderRange}, advancing the window
     * start to {@code newStart} ({@code e}). {@code w.f}=index, {@code w.p}=index2 prevent re-processing.
     *
     * @param count  number of visible polygons (obf var1). @param step window depth (obf var2, =100).
     * @param dummy  anti-tamper (obf var3).                 @param polys the visible-polygon array (var4).
     */
    private void polygonsIntersectSort(int count, int step, int dummy, WorldEntity[] polys) {
        for (int i = 0; i <= count; i++) {
            polys[i].active = false; // skipSomething
            polys[i].slotIndex = i;     // index
            polys[i].prevSortIndex = -1;    // index2
        }
        int low = 0;
        do {
            while (polys[low].active) {  // skip already-processed
                low++;
            }
            if (low == count) {
                return;
            }
            WorldEntity poly = polys[low];
            poly.active = true;
            int start = low;                 // var7
            int end = low + step;            // var8
            if (end >= count) {
                end = count - 1;
            }
            for (int k = end; k >= start + 1; k--) { // var9 from end down to start+1
                WorldEntity other = polys[k];
                if (poly.minX < other.maxX && other.minX < poly.maxX   // minX/maxX overlap (w.e/w.m)
                        // DEOB FIX (class c — inverted comparison): the second-axis overlap test was
                        // `poly.minZ < other.maxZ`, but the clean base is `~var6.j < ~var10.h` which
                        // decodes to var6.j > var10.h == poly.minZ(j) > other.maxZ(h). Cross-checked vs
                        // OpenRSC three/Scene.java:695-696 (`var10.maxP2 > var6.minP2 && var10.minP2 <
                        // var6.maxP2`): both axis tests are `other.<hi> > poly.<lo> && poly.<hi> >
                        // other.<lo>` half-overlaps. The inverted `<` made the second-axis overlap gate
                        // almost always false, so the overlap sort rarely ran. Cite: clean lb.java:993-994.
                        && other.minZ > poly.maxZ && poly.minZ > other.maxZ // 2nd-axis overlap (w.j/w.h)
                        && other.prevSortIndex != poly.slotIndex                  // index2 != index
                        && !separatedOrInOrder((byte) -84, other, poly)
                        && faceOrders(false, other, poly)) {
                    reorderRange(start, polys, k, (byte) 34); // polygonsOrder(polys, start, k)
                    if (polys[k] != other) {
                        k++;
                    }
                    start = this.newStart; // e
                    other.prevSortIndex = poly.slotIndex;      // index2 = index
                }
            }
        } while (true);
    }

    /**
     * Recursive helper for the overlap sort. obf: a(int,w[],int,byte) (counter {@code bb}). Tries to move
     * {@code polys[low]} forward past everything it correctly occludes, recursing up to the byte-guard
     * depth. Returns true if a reorder was applied. Tracks {@code newEnd}(eb)/{@code newStart}(e).
     */
    private boolean reorderRange(int low, WorldEntity[] polys, int high, byte guard) {
        while (true) {
            WorldEntity left = polys[low];
            for (int j = low + 1; j <= high; j++) {
                WorldEntity right = polys[j];
                if (!separatedOrInOrder((byte) -114, left, right)) {
                    break;
                }
                polys[low] = right;
                polys[j] = left;
                low = j;
                if (low == high) {
                    this.newEnd = low - 1;   // eb
                    this.newStart = low;     // e
                    return true;
                }
            }
            WorldEntity rightEnd = polys[high];
            int k = high - 1;
            while (low <= k) {
                WorldEntity mid = polys[k];
                if (!separatedOrInOrder((byte) -46, mid, rightEnd)) {
                    break;
                }
                polys[high] = mid;
                polys[k] = rightEnd;
                high = k;
                if (low == high) {
                    this.newEnd = high;       // eb
                    this.newStart = high + 1; // e
                    return true;
                }
                k--;
            }
            if (high <= low + 1) { // clean: ~high >= ~(low+1)  ==  high <= low+1
                this.newEnd = high;  // eb
                this.newStart = low; // e
                return false;
            }
            if (!reorderRange(low + 1, polys, high, (byte) 70)) {
                this.newStart = low; // e
                return false;
            }
            high = this.newEnd; // eb
        }
    }

    /**
     * Quick separating-axis test deciding whether {@code a} may be painted before {@code b}.
     * obf: a(byte,w,w) (counter {@code z}). Returns true if they are separated or {@code a} is correctly
     * behind {@code b}; otherwise runs the per-vertex plane tests in both directions, falling back to the
     * convex screen-overlap test ({@link #polygonsOverlap}). The {@code byte} is an anti-tamper dummy.
     */
    private boolean separatedOrInOrder(byte guard, WorldEntity a, WorldEntity b) {
        // AABB rejects (clean lines 272-298; ~-idiom decoded):
        if (b.minX >= a.maxX) return true; // var3.e >= var2.m  (screen-X disjoint)
        if (a.minX >= b.maxX) return true; // var2.e >= var3.m
        if (b.maxZ >= a.minZ) return true; // ~var3.h <= ~var2.j  ==  var3.h >= var2.j  (screen-Y)
        if (a.maxZ >= b.minZ) return true; // ~var2.h <= ~var3.j  ==  var2.h >= var3.j
        if (a.minDepth <= b.maxDepth) return true; // var2.q <= var3.u    (depth ranges)
        if (b.minDepth < a.maxDepth) return false; // var3.q < var2.u

        GameModel mb = b.model, ma = a.model; // var4 = b.model, var5 = a.model
        int fb = b.faceIndex, fa = a.faceIndex;       // var6 = b.face, var7 = a.face
        int[] bVerts = mb.faceVertices[fb]; // var8 = var4.o[var6]
        int[] aVerts = ma.faceVertices[fa]; // var9 = var5.o[var7]
        int nb = mb.faceNumVertices[fb];    // var10
        int na = ma.faceNumVertices[fa];    // var11

        // first plane test: a's plane (normal a.normalX/l/k, origin = a's first vertex, tolerance a.normalZ-magnitude
        // var21 = ma.normalMagnitude[fa]) against EACH of b's vertices (var8).
        int ox = ma.projectVertexX[aVerts[0]]; // var15
        int oy = ma.projectVertexY[aVerts[0]]; // var16
        int oz = ma.projectVertexZ[aVerts[0]]; // var17
        int ar = a.normalX, al = a.normalY, ak = a.normalZ;      // var18/var19/var20
        int tolA = ma.normalMagnitude[fa];     // var21
        int avis = a.normalDot;                        // var22
        boolean bOutside = false;              // var14
        for (int i = 0; i < nb; i++) {
            int v = bVerts[i];
            int dot = (oz - mb.projectVertexZ[v]) * ak
                    + (oy - mb.projectVertexY[v]) * al
                    + ar * (ox - mb.projectVertexX[v]); // var13
            // clean: if (dot < -tolA && avis < 0 || dot > tolA && avis > 0) outside
            if (dot < -tolA && avis < 0 || dot > tolA && avis > 0) {
                bOutside = true;
                break;
            }
        }
        if (!bOutside) return true; // all of b within a's plane band -> a may precede b

        // second plane test: b's plane (normal b.normalX/l/k, origin = b's first vertex, tolerance
        // var21 = mb.normalMagnitude[fb]) against EACH of a's vertices (var9).
        int oz2 = mb.projectVertexZ[bVerts[0]]; // var17
        int oy2 = mb.projectVertexY[bVerts[0]]; // var16
        int ox2 = mb.projectVertexX[bVerts[0]]; // var15
        int br = b.normalX, bl = b.normalY, bk = b.normalZ;       // var18/var19/var20
        int tolB = mb.normalMagnitude[fb];      // var21 = var4.k[var6]
        int bvis = b.normalDot;                         // var22 = var3.s
        boolean aOutside = false;               // var35
        for (int i = 0; i < na; i++) {
            int v = aVerts[i];
            int dot = (oz2 - ma.projectVertexZ[v]) * bk
                    + (ox2 - ma.projectVertexX[v]) * br
                    + (oy2 - ma.projectVertexY[v]) * bl; // var34
            // clean: if (-tolB > dot && ~bvis<-1(bvis>0) || dot>tolB && ~bvis>-1(bvis<0)) outside
            if (-tolB > dot && bvis > 0 || dot > tolB && bvis < 0) {
                aOutside = true;
                break;
            }
        }
        if (!aOutside) return true;

        // ambiguous: convex screen-overlap fallback. Build b's outline (var45=X, var24=Y) from var4(mb)
        // and a's outline (var25=X, var26=Y) from var5(ma); 2-vert faces become a +/-20 padded quad.
        int[] bX, bY; // var45, var24
        if (nb != 2) {
            bX = new int[nb]; bY = new int[nb];
            for (int i = 0; i < nb; i++) { int v = bVerts[i]; bX[i] = mb.vertexViewX[v]; bY[i] = mb.vertexViewY[v]; }
        } else {
            bX = new int[4]; bY = new int[4];
            int v0 = bVerts[0], v1 = bVerts[1];
            bX[0] = mb.vertexViewX[v0] - 20; bX[1] = mb.vertexViewX[v1] - 20;
            bX[2] = mb.vertexViewX[v1] + 20; bX[3] = mb.vertexViewX[v0] + 20;
            bY[0] = bY[3] = mb.vertexViewY[v0]; bY[1] = bY[2] = mb.vertexViewY[v1];
        }
        int[] aX, aY; // var25, var26
        if (na != 2) {
            aX = new int[na]; aY = new int[na];
            for (int i = 0; i < na; i++) { int v = aVerts[i]; aX[i] = ma.vertexViewX[v]; aY[i] = ma.vertexViewY[v]; }
        } else {
            aX = new int[4]; aY = new int[4];
            int v0 = aVerts[0], v1 = aVerts[1];
            aX[0] = ma.vertexViewX[v0] - 20; aX[1] = ma.vertexViewX[v1] - 20;
            aX[2] = ma.vertexViewX[v1] + 20; aX[3] = ma.vertexViewX[v0] + 20;
            aY[0] = aY[3] = ma.vertexViewY[v0]; aY[1] = aY[2] = ma.vertexViewY[v1];
        }
        // clean: return !this.a(var25, var24, var45, var26, 1)  ==  !polygonsOverlap(aX, bY, bX, aY, 1)
        return !polygonsOverlap(aX, bY, bX, aY, 1);
    }

    /**
     * Full plane-side comparison used by the overlap-sort pass. obf: a(boolean,w,w) (counter {@code C}).
     * Returns true if {@code a} is on the correct side of {@code b} (or vice-versa) for every vertex.
     * Faithful to the clean base (two near-identical halves). The boolean is a dummy.
     */
    private boolean faceOrders(boolean guard, WorldEntity a, WorldEntity b) {
        GameModel ma = a.model, mb = b.model;          // var4 = a.model, var5 = b.model
        int fa = a.faceIndex, fb = b.faceIndex;                // var6 = a.face, var7 = b.face
        int[] aVerts = ma.faceVertices[fa];    // var8
        int[] bVerts = mb.faceVertices[fb];    // var9
        int na = ma.faceNumVertices[fa];       // var10
        int nb = mb.faceNumVertices[fb];       // var11

        // first half: b's plane (normal b.normalX/l/k, origin = b's first vertex, tolerance var21 =
        // mb.normalMagnitude[fb]) against EACH of a's vertices (var8, indexed into ma).
        int ox = mb.projectVertexX[bVerts[0]]; // var15
        int oy = mb.projectVertexY[bVerts[0]]; // var16
        int oz = mb.projectVertexZ[bVerts[0]]; // var17
        int br = b.normalX, bl = b.normalY, bk = b.normalZ;      // var18/var19/var20
        int tolB = mb.normalMagnitude[fb];     // var21 = var5.k[var7]
        int bvis = b.normalDot;                        // var22 = var3.s
        boolean someOutsideB = false;          // var14
        for (int i = 0; i < na; i++) {
            int v = aVerts[i];
            int dot = bl * (oy - ma.projectVertexY[v])
                    + (ox - ma.projectVertexX[v]) * br
                    + (oz - ma.projectVertexZ[v]) * bk; // var13
            // clean: if (dot < -tolB && bvis < 0 || dot > tolB && bvis > 0) outside
            if (dot < -tolB && bvis < 0 || dot > tolB && bvis > 0) {
                someOutsideB = true;
                break;
            }
        }
        if (!someOutsideB) return true; // all of a within b's band -> in order

        // second half: a's plane (normal a.normalX/l/k, origin = a's first vertex, tolerance var21 =
        // ma.normalMagnitude[fa]) against EACH of b's vertices (var9, indexed into mb). Returns true iff
        // none of b's vertices fall outside a's band.
        int ox2 = ma.projectVertexX[aVerts[0]]; // var15
        int oy2 = ma.projectVertexY[aVerts[0]]; // var16
        int oz2 = ma.projectVertexZ[aVerts[0]]; // var17
        int ar = a.normalX, al = a.normalY, ak = a.normalZ;       // var18/var19/var20
        int tolA = ma.normalMagnitude[fa];      // var21 = var4.k[var6]
        int avis = a.normalDot;                         // var22 = var2.s
        boolean someOutsideA = false;           // var14
        for (int i = 0; i < nb; i++) {
            int v = bVerts[i];
            int dot = ar * (ox2 - mb.projectVertexX[v])
                    + al * (oy2 - mb.projectVertexY[v])
                    + (oz2 - mb.projectVertexZ[v]) * ak; // var27 (clean factored as a - (- - -))
            // clean: if (dot < -tolA && avis > 0 || dot > tolA && avis < 0) outside
            if (dot < -tolA && avis > 0 || dot > tolA && avis < 0) {
                someOutsideA = true;
                break;
            }
        }
        return !someOutsideA;
    }

    // ==================================================================
    // Convex-polygon screen overlap (intersect)
    // ==================================================================

    /**
     * Convex-polygon overlap test used as a tie-breaker in the overlap sort. obf: a(int[],int[],int[],int[],int)
     * (counter {@code V}; oracle {@code Scene.intersect}). Given two convex polygons by their screen
     * outlines, returns true if they overlap. The trailing {@code int} is an anti-tamper dummy.
     *
     * <p>Clean parameter order: {@code var1, var2, var3, var4, var5(dummy)}, where {@code var3.length=var6}
     * and {@code var1.length=var7}. It finds the top vertex of each polygon, walks the left/right boundary
     * of each in lockstep through three phases ({@code var16} 0/1/2), interpolating edge crossings with
     * {@link #interpolate} and testing half-plane order with {@link #spanOrder}/{@link #edgeOrder}.
     *
     * <p>The clean base body is an irreducible state machine (Vineflower emitted it with "$VF: Irreducible
     * bytecode … not entirely decomposed"): the three-phase boundary walk is a single labelled loop whose
     * {@code var16}(=oracle {@code byte0}) state selects one of three sub-machines. It is reproduced here
     * FULLY, restructured into the three {@code while (state == N)} phases of oracle {@code Scene.intersect}
     * (semantically identical to the clean irreducible flow). Every {@link #interpolate}/{@link #spanOrder}/
     * {@link #edgeOrder} call preserves the clean base's exact argument order (clean lines ~4170-4630).
     *
     * <p>Cursors: {@code var8}=k (polygon-2 left boundary), {@code var9}=l (polygon-2 right boundary),
     * {@code var10}=i1 (polygon-4 left boundary), {@code var11}=j1 (polygon-4 right boundary); {@code var16}
     * is the walk state (0 = both polygons still bounding; 1 = polygon-2 boundaries met, walk polygon-4;
     * 2 = polygon-4 boundaries met, walk polygon-2); {@code var17} is the half-plane orientation flag.
     */
    private boolean polygonsOverlap(int[] var1, int[] var2, int[] var3, int[] var4, int dummy) {
        int var6 = var3.length;
        int var7 = var1.length;
        int var8 = 0;                 // index of min(var2) (top vertex of outline 2)
        int var18, var20;
        var20 = var18 = var2[0];      // var18 = min(var2), var20 = max(var2)
        int var10 = 0;                // index of min(var4) (top vertex of outline 4)
        for (int t = 1; t < var6; t++) {
            if (var2[t] < var18) { var18 = var2[t]; var8 = t; }      // clean else-branch: var18 > var2[t]
            else if (var2[t] > var20) { var20 = var2[t]; }           // clean if-branch: update max
        }
        int var19 = var4[0], var21 = var4[0];   // var19 = min(var4), var21 = max(var4)
        for (int t = 1; t < var7; t++) {
            if (var4[t] < var19) { var19 = var4[t]; var10 = t; }
            else if (var21 < var4[t]) { var21 = var4[t]; }
        }
        // cheap separating-extent rejections (clean: gate var20 > var19; reject var18 >= var21):
        if (var18 >= var21) return false;   // min(var2) >= max(var4): outline 2 below outline 4
        if (var19 >= var20) return false;   // min(var4) >= max(var2): outline 4 below outline 2

        int var9, var11;
        boolean var17;
        byte var16;
        if (~var2[var8] > ~var4[var10]) { // var2[var8] < var4[var10]: polygon-2 starts higher
            var9 = var8;
            while (~var4[var10] < ~var2[var8]) {       // var4[var10] > var2[var8]
                var8 = (var8 - (1 - var6)) % var6;
            }
            while (var2[var9] < var4[var10]) {
                var9 = (1 + var9) % var6;
            }
            int e1 = interpolate(var2[(var8 + 1) % var6], false, var4[var10], var2[var8], var3[var8], var3[(1 + var8) % var6]);
            int e2 = interpolate(var2[(var6 - 1 + var9) % var6], false, var4[var10], var2[var9], var3[var9], var3[(var6 - 1 + var9) % var6]);
            int b0 = var1[var10];
            var17 = e1 < b0 | ~b0 < ~e2;               // e1 < b0 | e2 > b0
            if (edgeOrder(b0, var17, e1, (byte) -71, e2)) {
                return true;
            }
            var11 = (var10 + 1) % var7;
            var10 = (var10 - 1 + var7) % var7;
            var16 = (var8 == var9) ? (byte) 1 : 0;
        } else {                          // polygon-4 starts higher (or equal): symmetric entry
            var11 = var10;
            while (var2[var8] > var4[var10]) {
                var10 = (var10 - 1 + var7) % var7;
            }
            int e0 = var3[var8];
            while (var2[var8] > var4[var11]) {
                var11 = (var11 + 1) % var7;
            }
            int e1 = interpolate(var4[(var10 + 1) % var7], false, var2[var8], var4[var10], var1[var10], var1[(var10 + 1) % var7]);
            int e2 = interpolate(var4[(var7 + var11 - 1) % var7], false, var2[var8], var4[var11], var1[var11], var1[(var11 - 1 + var7) % var7]);
            var17 = e0 < e1 | ~e2 < ~e0;
            if (edgeOrder(e0, !var17, e1, (byte) -71, e2)) {
                return true;
            }
            var9 = (1 + var8) % var6;
            var8 = (var6 + var8 - 1) % var6;
            var16 = (~var11 == ~var10) ? (byte) 2 : 0;
        }

        // --- three-phase boundary walk ------------------------------------------------------------
        // The clean base merges these three phases into one irreducible region (Vineflower: "Irreducible
        // bytecode … not entirely decomposed"). They are transcribed here FULLY and faithfully in the
        // fully-reduced three-loop form (oracle Scene.intersect, the canonical decomposition of the exact
        // same algorithm), with every {@link #interpolate}/{@link #spanOrder}/{@link #edgeOrder} call mapped
        // back to the clean helper's argument order. Cursor map: var8/var9 are the two boundary cursors of
        // the (var2=Y, var3=X) polygon; var10/var11 are the cursors of the (var4=Y, var1=X) polygon. var6
        // and var7 are those polygons' vertex counts. The interpolate/spanOrder/edgeOrder argument-order
        // translation was verified branch-for-branch against the clean base (e.g. clean lines 4509-4623).
        //
        // PHASE 0 (var16==0): both polygons still have a left (var8/var10) + right (var9/var11) boundary.
        // The 4-way outer split advances whichever boundary is highest; drop to phase 1 when var8==var9,
        // phase 2 when var10==var11. (oracle Scene.intersect byte0==0, lines 2951-3038.)
        while (var16 == 0) {
            if (var2[var8] < var2[var9]) {
                if (var2[var8] < var4[var10]) {
                    if (var2[var8] < var4[var11]) {
                        int i2 = var3[var8];
                        int l6 = interpolate(var2[(var9 - 1 + var6) % var6], false, var2[var8], var2[var9], var3[var9], var3[(var9 - 1 + var6) % var6]);
                        int j11 = interpolate(var4[(var10 + 1) % var7], false, var2[var8], var4[var10], var1[var10], var1[(var10 + 1) % var7]);
                        int i16 = interpolate(var4[(var11 - 1 + var7) % var7], false, var2[var8], var4[var11], var1[var11], var1[(var11 - 1 + var7) % var7]);
                        // method307(i2,l6,j11,i16,flag) -> spanOrder(_, flag, l6, i16, i2, j11)
                        if (spanOrder((byte) -127, var17, l6, i16, i2, j11)) {
                            return true;
                        }
                        var8 = (var8 - 1 + var6) % var6;
                        if (var8 == var9) {
                            var16 = 1;
                        }
                    } else {
                        int j2 = interpolate(var2[(var8 + 1) % var6], false, var4[var11], var2[var8], var3[var8], var3[(var8 + 1) % var6]);
                        int i7 = interpolate(var2[(var9 - 1 + var6) % var6], false, var4[var11], var2[var9], var3[var9], var3[(var9 - 1 + var6) % var6]);
                        int k11 = interpolate(var4[(var10 + 1) % var7], false, var4[var11], var4[var10], var1[var10], var1[(var10 + 1) % var7]);
                        int j16 = var1[var11];
                        if (spanOrder((byte) -113, var17, i7, j16, j2, k11)) {
                            return true;
                        }
                        var11 = (var11 + 1) % var7;
                        if (var10 == var11) {
                            var16 = 2;
                        }
                    }
                } else if (var4[var10] < var4[var11]) {
                    int k2 = interpolate(var2[(var8 + 1) % var6], false, var4[var10], var2[var8], var3[var8], var3[(var8 + 1) % var6]);
                    int j7 = interpolate(var2[(var9 - 1 + var6) % var6], false, var4[var10], var2[var9], var3[var9], var3[(var9 - 1 + var6) % var6]);
                    int l11 = var1[var10];
                    int k16 = interpolate(var4[(var11 - 1 + var7) % var7], false, var4[var10], var4[var11], var1[var11], var1[(var11 - 1 + var7) % var7]);
                    if (spanOrder((byte) -116, var17, j7, k16, k2, l11)) {
                        return true;
                    }
                    var10 = (var10 - 1 + var7) % var7;
                    if (var10 == var11) {
                        var16 = 2;
                    }
                } else {
                    int l2 = interpolate(var2[(var8 + 1) % var6], false, var4[var11], var2[var8], var3[var8], var3[(var8 + 1) % var6]);
                    int k7 = interpolate(var2[(var9 - 1 + var6) % var6], false, var4[var11], var2[var9], var3[var9], var3[(var9 - 1 + var6) % var6]);
                    int i12 = interpolate(var4[(var10 + 1) % var7], false, var4[var11], var4[var10], var1[var10], var1[(var10 + 1) % var7]);
                    int l16 = var1[var11];
                    if (spanOrder((byte) -123, var17, k7, l16, l2, i12)) {
                        return true;
                    }
                    var11 = (var11 + 1) % var7;
                    if (var10 == var11) {
                        var16 = 2;
                    }
                }
            } else if (var2[var9] < var4[var10]) {
                if (var2[var9] < var4[var11]) {
                    int i3 = interpolate(var2[(var8 + 1) % var6], false, var2[var9], var2[var8], var3[var8], var3[(var8 + 1) % var6]);
                    int l7 = var3[var9];
                    int j12 = interpolate(var4[(var10 + 1) % var7], false, var2[var9], var4[var10], var1[var10], var1[(var10 + 1) % var7]);
                    int i17 = interpolate(var4[(var11 - 1 + var7) % var7], false, var2[var9], var4[var11], var1[var11], var1[(var11 - 1 + var7) % var7]);
                    if (spanOrder((byte) -105, var17, l7, i17, i3, j12)) {
                        return true;
                    }
                    var9 = (var9 + 1) % var6;
                    if (var8 == var9) {
                        var16 = 1;
                    }
                } else {
                    int j3 = interpolate(var2[(var8 + 1) % var6], false, var4[var11], var2[var8], var3[var8], var3[(var8 + 1) % var6]);
                    int i8 = interpolate(var2[(var9 - 1 + var6) % var6], false, var4[var11], var2[var9], var3[var9], var3[(var9 - 1 + var6) % var6]);
                    int k12 = interpolate(var4[(var10 + 1) % var7], false, var4[var11], var4[var10], var1[var10], var1[(var10 + 1) % var7]);
                    int j17 = var1[var11];
                    if (spanOrder((byte) -123, var17, i8, j17, j3, k12)) {
                        return true;
                    }
                    var11 = (var11 + 1) % var7;
                    if (var10 == var11) {
                        var16 = 2;
                    }
                }
            } else if (var4[var10] < var4[var11]) {
                int k3 = interpolate(var2[(var8 + 1) % var6], false, var4[var10], var2[var8], var3[var8], var3[(var8 + 1) % var6]);
                int j8 = interpolate(var2[(var9 - 1 + var6) % var6], false, var4[var10], var2[var9], var3[var9], var3[(var9 - 1 + var6) % var6]);
                int l12 = var1[var10];
                int k17 = interpolate(var4[(var11 - 1 + var7) % var7], false, var4[var10], var4[var11], var1[var11], var1[(var11 - 1 + var7) % var7]);
                if (spanOrder((byte) -106, var17, j8, k17, k3, l12)) {
                    return true;
                }
                var10 = (var10 - 1 + var7) % var7;
                if (var10 == var11) {
                    var16 = 2;
                }
            } else {
                int l3 = interpolate(var2[(var8 + 1) % var6], false, var4[var11], var2[var8], var3[var8], var3[(var8 + 1) % var6]);
                int k8 = interpolate(var2[(var9 - 1 + var6) % var6], false, var4[var11], var2[var9], var3[var9], var3[(var9 - 1 + var6) % var6]);
                int i13 = interpolate(var4[(var10 + 1) % var7], false, var4[var11], var4[var10], var1[var10], var1[(var10 + 1) % var7]);
                int l17 = var1[var11];
                if (spanOrder((byte) -123, var17, k8, l17, l3, i13)) {
                    return true;
                }
                var11 = (var11 + 1) % var7;
                if (var10 == var11) {
                    var16 = 2;
                }
            }
        }

        // PHASE 1 (var16==1): var8/var9 (the var2/var3 polygon's cursors) have met; only var10/var11 still
        // walk. On exhaustion either resolve via edgeOrder or fall back to phase 0.
        while (var16 == 1) {
            if (var2[var8] < var4[var10]) {
                if (var2[var8] < var4[var11]) {
                    int i4 = var3[var8];
                    int j13 = interpolate(var4[(var10 + 1) % var7], false, var2[var8], var4[var10], var1[var10], var1[(var10 + 1) % var7]);
                    int i18 = interpolate(var4[(var11 - 1 + var7) % var7], false, var2[var8], var4[var11], var1[var11], var1[(var11 - 1 + var7) % var7]);
                    // method308(j13,i18,i4,!flag) -> edgeOrder(i4, !flag, j13, _, i18)
                    return edgeOrder(i4, !var17, j13, (byte) -71, i18);
                }
                int j4 = interpolate(var2[(var8 + 1) % var6], false, var4[var11], var2[var8], var3[var8], var3[(var8 + 1) % var6]);
                int l8 = interpolate(var2[(var9 - 1 + var6) % var6], false, var4[var11], var2[var9], var3[var9], var3[(var9 - 1 + var6) % var6]);
                int k13 = interpolate(var4[(var10 + 1) % var7], false, var4[var11], var4[var10], var1[var10], var1[(var10 + 1) % var7]);
                int j18 = var1[var11];
                if (spanOrder((byte) -118, var17, l8, j18, j4, k13)) {
                    return true;
                }
                var11 = (var11 + 1) % var7;
                if (var10 == var11) {
                    var16 = 0;
                }
            } else if (var4[var10] < var4[var11]) {
                int k4 = interpolate(var2[(var8 + 1) % var6], false, var4[var10], var2[var8], var3[var8], var3[(var8 + 1) % var6]);
                int i9 = interpolate(var2[(var9 - 1 + var6) % var6], false, var4[var10], var2[var9], var3[var9], var3[(var9 - 1 + var6) % var6]);
                int l13 = var1[var10];
                int k18 = interpolate(var4[(var11 - 1 + var7) % var7], false, var4[var10], var4[var11], var1[var11], var1[(var11 - 1 + var7) % var7]);
                if (spanOrder((byte) -98, var17, i9, k18, k4, l13)) {
                    return true;
                }
                var10 = (var10 - 1 + var7) % var7;
                if (var10 == var11) {
                    var16 = 0;
                }
            } else {
                int l4 = interpolate(var2[(var8 + 1) % var6], false, var4[var11], var2[var8], var3[var8], var3[(var8 + 1) % var6]);
                int j9 = interpolate(var2[(var9 - 1 + var6) % var6], false, var4[var11], var2[var9], var3[var9], var3[(var9 - 1 + var6) % var6]);
                int i14 = interpolate(var4[(var10 + 1) % var7], false, var4[var11], var4[var10], var1[var10], var1[(var10 + 1) % var7]);
                int l18 = var1[var11];
                if (spanOrder((byte) -118, var17, j9, l18, l4, i14)) {
                    return true;
                }
                var11 = (var11 + 1) % var7;
                if (var10 == var11) {
                    var16 = 0;
                }
            }
        }

        // PHASE 2 (var16==2): var10/var11 (the var4/var1 polygon's cursors) have met; only var8/var9 still
        // walk. On exhaustion either resolve via edgeOrder or fall back to phase 0.
        while (var16 == 2) {
            if (var4[var10] < var2[var8]) {
                if (var4[var10] < var2[var9]) {
                    int i5 = interpolate(var2[(var8 + 1) % var6], false, var4[var10], var2[var8], var3[var8], var3[(var8 + 1) % var6]);
                    int k9 = interpolate(var2[(var9 - 1 + var6) % var6], false, var4[var10], var2[var9], var3[var9], var3[(var9 - 1 + var6) % var6]);
                    int j14 = var1[var10];
                    // method308(i5,k9,j14,flag) -> edgeOrder(j14, flag, i5, _, k9)
                    return edgeOrder(j14, var17, i5, (byte) -71, k9);
                }
                int j5 = interpolate(var2[(var8 + 1) % var6], false, var2[var9], var2[var8], var3[var8], var3[(var8 + 1) % var6]);
                int l9 = var3[var9];
                int k14 = interpolate(var4[(var10 + 1) % var7], false, var2[var9], var4[var10], var1[var10], var1[(var10 + 1) % var7]);
                int i19 = interpolate(var4[(var11 - 1 + var7) % var7], false, var2[var9], var4[var11], var1[var11], var1[(var11 - 1 + var7) % var7]);
                if (spanOrder((byte) -102, var17, l9, i19, j5, k14)) {
                    return true;
                }
                var9 = (var9 + 1) % var6;
                if (var8 == var9) {
                    var16 = 0;
                }
            } else if (var2[var8] < var2[var9]) {
                int k5 = var3[var8];
                int i10 = interpolate(var2[(var9 - 1 + var6) % var6], false, var2[var8], var2[var9], var3[var9], var3[(var9 - 1 + var6) % var6]);
                int l14 = interpolate(var4[(var10 + 1) % var7], false, var2[var8], var4[var10], var1[var10], var1[(var10 + 1) % var7]);
                int j19 = interpolate(var4[(var11 - 1 + var7) % var7], false, var2[var8], var4[var11], var1[var11], var1[(var11 - 1 + var7) % var7]);
                if (spanOrder((byte) -116, var17, i10, j19, k5, l14)) {
                    return true;
                }
                var8 = (var8 - 1 + var6) % var6;
                if (var8 == var9) {
                    var16 = 0;
                }
            } else {
                int l5 = interpolate(var2[(var8 + 1) % var6], false, var2[var9], var2[var8], var3[var8], var3[(var8 + 1) % var6]);
                int j10 = var3[var9];
                int i15 = interpolate(var4[(var10 + 1) % var7], false, var2[var9], var4[var10], var1[var10], var1[(var10 + 1) % var7]);
                int k19 = interpolate(var4[(var11 - 1 + var7) % var7], false, var2[var9], var4[var11], var1[var11], var1[(var11 - 1 + var7) % var7]);
                if (spanOrder((byte) -102, var17, j10, k19, l5, i15)) {
                    return true;
                }
                var9 = (var9 + 1) % var6;
                if (var8 == var9) {
                    var16 = 0;
                }
            }
        }

        // Final resolution once both polygons are reduced to a single bounding edge (clean lines 4626-4629 /
        // oracle lines 3115-3124). var8 is the var2/var3 polygon's surviving cursor; var10/var11 the other's.
        if (var2[var8] < var4[var10]) {
            int i6 = var3[var8];
            int j15 = interpolate(var4[(var10 + 1) % var7], false, var2[var8], var4[var10], var1[var10], var1[(var10 + 1) % var7]);
            int l19 = interpolate(var4[(var11 - 1 + var7) % var7], false, var2[var8], var4[var11], var1[var11], var1[(var11 - 1 + var7) % var7]);
            // method308(j15,l19,i6,!flag) -> edgeOrder(i6, !flag, j15, _, l19)
            return edgeOrder(i6, !var17, j15, (byte) -71, l19);
        }
        int j6 = interpolate(var2[(var8 + 1) % var6], false, var4[var10], var2[var8], var3[var8], var3[(var8 + 1) % var6]);
        int k10 = interpolate(var2[(var9 - 1 + var6) % var6], false, var4[var10], var2[var9], var3[var9], var3[(var9 - 1 + var6) % var6]);
        int k15 = var1[var10];
        // method308(j6,k10,k15,flag) -> edgeOrder(k15, flag, j6, _, k10)
        return edgeOrder(k15, var17, j6, (byte) -71, k10);
    }

    /**
     * Two-span order predicate. obf: a(byte,boolean,int,int,int,int) (counter {@code Lb}; oracle method307).
     * The {@code byte} arg is an anti-tamper dummy. Faithful to the clean base.
     */
    public boolean spanOrder(byte guard, boolean inclusive, int var3, int var4, int var5, int var6) {
        if ((!inclusive || ~var6 > ~var5) && var5 >= var6) {
            if (var5 < var4) return true;
            if (var3 < var6) return true;
            return ~var4 < ~var3 ? true : inclusive;
        } else if (var5 <= var4) {
            if (var3 > var6) return true;
            return var3 <= var4 ? !inclusive : true;
        } else {
            return true;
        }
    }

    /**
     * Single-edge span order predicate. obf: a(int,boolean,int,byte,int) (counter {@code O}; oracle method308).
     * The {@code byte} arg is an anti-tamper dummy. Faithful to the clean base.
     */
    public boolean edgeOrder(int var1, boolean inclusive, int var3, byte guard, int var5) {
        if ((!inclusive || ~var1 > ~var3) && var3 >= var1) {
            return ~var1 < ~var5 ? true : inclusive;
        } else {
            return var1 >= var5 ? !inclusive : true;
        }
    }

    // ==================================================================
    // The main render loop
    // ==================================================================

    /**
     * Renders the whole scene for one frame. obf: c(int) (counter {@code Y}) — arg is a dummy.
     * <ol>
     *   <li>copy {@link Surface#i} (interlace) into {@link #interlace};</li>
     *   <li>reset the six scattered frustum accumulators and rebuild the world-space frustum AABB from
     *       the eight view-volume corners via {@link #expandFrustum}, then translate by the camera;</li>
     *   <li>append + project every model (and the billboard {@code view} model);</li>
     *   <li>build the visible-polygon list (frustum + side cull, {@link #initialisePolygon3d}/2d);</li>
     *   <li>depth-sort ({@link #polygonsQSort}) then overlap-sort ({@link #polygonsIntersectSort});</li>
     *   <li>draw each polygon back-to-front: sprites via {@link Surface} clipping blits, 3D faces via
     *       {@link #generateScanlines} + {@link #rasterize} with per-vertex shade and fog.</li>
     * </ol>
     */
    public void render(int unused) { // obf: c(int) [LARGE]
        this.interlace = this.surface.interlace; // f = dc.i

        int halfClipFarX = this.clipX * this.clipFar3d >> this.viewDistance; // A*Mb>>R
        BZip.frustumMaxY = 0; DataStore.frustumMinY = 0; SocketFactory.frustumFarZ = 0; ClientStream.frustumNearZ = 0; // reset 4 accumulators
        int halfClipFarY = this.clipFar3d * this.clipY >> this.viewDistance;  // Mb*wb>>R
        NameHash.frustumMinX = 0; BZip.frustumMaxX = 0;                    // reset remaining 2

        // eight view-volume corners -> world AABB
        expandFrustum(this.clipFar3d, -halfClipFarX, -halfClipFarY, true);
        expandFrustum(this.clipFar3d, -halfClipFarX,  halfClipFarY, true);
        expandFrustum(this.clipFar3d,  halfClipFarX, -halfClipFarY, true);
        expandFrustum(this.clipFar3d,  halfClipFarX,  halfClipFarY, true);
        expandFrustum(0, -this.clipX, -this.clipY, true);
        expandFrustum(0, -this.clipX,  this.clipY, true);
        expandFrustum(0,  this.clipX, -this.clipY, true);
        expandFrustum(0,  this.clipX,  this.clipY, true);

        // translate the AABB by the camera look offset (clean assignments verbatim)
        DataStore.frustumMinY += this.cameraLookY; ClientStream.frustumNearZ += this.cameraLookZ; BZip.frustumMaxY += this.cameraLookY;
        BZip.frustumMaxX += this.cameraLookX; SocketFactory.frustumFarZ += this.cameraLookZ; NameHash.frustumMinX += this.cameraLookX;

        // append the billboard model, then project everything
        this.models[this.modelCount] = this.view; // Z[ab] = T
        this.view.transformState = 2;             // T.Yb = 2
        for (int i = 0; i < this.modelCount; i++) {
            this.models[i].project(this.cameraLookY, this.viewDistance, this.cameraLookX, (byte) -122,
                    this.cameraLookZ, this.cameraRoll, this.cameraYaw, this.cameraPitch, this.cameraClipNear);
        }
        this.view.project(this.cameraLookY, this.viewDistance, this.cameraLookX, (byte) -114,
                this.cameraLookZ, this.cameraRoll, this.cameraYaw, this.cameraPitch, this.cameraClipNear);
        this.visiblePolygonCount = 0; // zb = 0

        // --- build visible-polygon list for solid models ---
        for (int mi = 0; mi < this.modelCount; mi++) {
            GameModel model = this.models[mi];
            if (!model.visible) continue; // ca.dc
            for (int face = 0; face < model.numFaces; face++) { // ca.t
                int nv = model.faceNumVertices[face]; // ca.lb
                int[] verts = model.faceVertices[face]; // ca.o
                boolean inDepth = false;
                for (int v = 0; v < nv; v++) {
                    int z = model.projectVertexZ[verts[v]]; // ca.bb
                    if (this.cameraClipNear < z && z < this.clipFar3d) { inDepth = true; break; }
                }
                if (!inDepth) continue;
                int xMask = 0;
                for (int v = 0; v < nv; v++) {
                    int x = model.vertexViewX[verts[v]]; // ca.pb
                    if (x > -this.clipX) xMask |= 1;
                    if (x < this.clipX) xMask |= 2;
                    if (xMask == 3) break;
                }
                if (xMask != 3) continue;
                int yMask = 0;
                for (int v = 0; v < nv; v++) {
                    int y = model.vertexViewY[verts[v]]; // ca.Ob
                    if (-this.clipY < y) yMask |= 1;
                    if (y < this.clipY) yMask |= 2;
                    if (yMask == 3) break;
                }
                if (yMask != 3) continue;
                WorldEntity poly = this.visiblePolygons[this.visiblePolygonCount]; // y[zb]
                poly.model = model;
                poly.faceIndex = face;
                initialisePolygon3d(this.visiblePolygonCount, -21875);
                int fill = (poly.normalDot < 0) ? model.faceFillFront[face] : model.faceFillBack[face]; // V / qb
                if (fill != COLOUR_TRANSPARENT) {
                    int zSum = 0;
                    for (int v = 0; v < nv; v++) zSum += model.projectVertexZ[verts[v]];
                    poly.sortDepth = model.depth + zSum / nv; // w.t = ca.hc + avgZ
                    poly.objectId = fill;                    // w.b = facefill
                    this.visiblePolygonCount++;       // zb++
                }
            }
        }

        // --- build visible-polygon list for the 2D billboard model ---
        GameModel viewModel = this.view;
        if (viewModel.visible) {
            for (int face = 0; face < viewModel.numFaces; face++) {
                int[] fv = viewModel.faceVertices[face];
                int v0 = fv[0];
                int vx = viewModel.vertexViewX[v0]; // pb
                int vy = viewModel.vertexViewY[v0]; // Ob
                int vz = viewModel.projectVertexZ[v0]; // bb
                if (this.cameraClipNear < vz && vz < this.clipFar2d) {
                    int w = (this.spriteWidth[face] << this.viewDistance) / vz;   // ob<<R / vz
                    int h = (this.spriteHeight[face] << this.viewDistance) / vz;  // Eb<<R / vz
                    if (this.clipX >= vx - w / 2 && vx + w / 2 >= -this.clipX
                            && vy - h <= this.clipY && -this.clipY <= vy) {
                        WorldEntity poly = this.visiblePolygons[this.visiblePolygonCount];
                        poly.faceIndex = face;
                        poly.model = viewModel;
                        initialisePolygon2d(-103, this.visiblePolygonCount);
                        poly.sortDepth = (viewModel.projectVertexZ[fv[1]] + vz) / 2; // w.t
                        this.visiblePolygonCount++;
                    }
                }
            }
        }
        if (this.visiblePolygonCount == 0) return;
        this.lastVisiblePolygonCount = this.visiblePolygonCount;

        // depth sort then overlap sort (clean: a(0,-1,y,zb-1); a(zb,100,-53,y))
        polygonsQSort(0, -1, this.visiblePolygons, this.visiblePolygonCount - 1);
        polygonsIntersectSort(this.visiblePolygonCount, 100, -53, this.visiblePolygons);

        // --- draw back-to-front ---
        for (int p = 0; p < this.visiblePolygonCount; p++) {
            WorldEntity poly = this.visiblePolygons[p];
            int face = poly.faceIndex;
            GameModel model = poly.model;
            if (model != this.view) {
                drawSolidFace(poly, model, face);
            } else {
                drawSpriteFace(model, face);
            }
        }
        this.mousePickingActive = false; // K = false
    }

    /**
     * Clips, shades and rasterises one solid 3D face. Extracted from {@link #render}'s draw loop.
     *
     * <p>Two parallel vertex buffer sets are produced:
     * <ul>
     *   <li>texture-coordinate buffers {@code projVertX/Y/Z} (obf Qb/Vb/J = cc/H/bb), one entry per face
     *       vertex {@code var17}, <em>unclipped</em>;</li>
     *   <li>screen-space buffers {@code screenX/screenY/shadeBuf} (obf yb/B/r = pb/Ob/shade), one entry per
     *       <em>clipped</em> vertex {@code var54} — vertices behind the near plane are replaced by their
     *       interpolated near-plane crossings (perspective-projected {@code (cc<<R)/nb}).</li>
     * </ul>
     * The shade clamp + texture-shift loop runs over the face's {@code var17} vertices, then
     * {@link #generateScanlines} consumes the clipped screen buffers and {@link #rasterize} the (unclipped)
     * texture buffers.
     */
    private void drawSolidFace(WorldEntity poly, GameModel model, int face) {
        int numVerts = model.faceNumVertices[face]; // var17 = lb[face]
        int clipped = 0;                            // var54
        int shade = 0;                              // var59
        int[] verts = model.faceVertices[face];     // var18 = o[face]
        if (model.faceIntensity[face] != COLOUR_TRANSPARENT) { // Hb[face] != magic -> flat shade
            shade = (poly.normalDot < 0)
                    ? model.lightAmbience - model.faceIntensity[face]  // Jb - Hb[face]
                    : model.lightAmbience + model.faceIntensity[face]; // Jb + Hb[face]
        }
        for (int i = 0; i < numVerts; i++) {        // var19
            int v = verts[i];                       // var33
            this.projVertX[i] = model.projectVertexX[v]; // Qb[i] = cc[v]
            this.projVertY[i] = model.projectVertexY[v]; // Vb[i] = H[v]
            this.projVertZ[i] = model.projectVertexZ[v]; // J[i]  = bb[v]
            if (model.faceIntensity[face] == COLOUR_TRANSPARENT) { // Gouraud (per-vertex) shade
                shade = (poly.normalDot < 0)
                        ? -model.vertexIntensity[v] + (model.lightAmbience + model.vertexAmbience[v]) // -gb + (Jb + Ab)
                        : model.vertexAmbience[v] + (model.lightAmbience + model.vertexIntensity[v]);  // Ab + (Jb + gb)
            }
            if (model.projectVertexZ[v] >= this.cameraClipNear) {
                // in front of the near plane: store the perspective-divided screen coords directly
                this.screenX[clipped] = model.vertexViewX[v]; // yb[var54] = pb[v]
                this.screenY[clipped] = model.vertexViewY[v]; // B[var54]  = Ob[v]
                this.shadeBuf[clipped] = shade;               // r[var54]  = shade
                if (model.projectVertexZ[v] > this.fogZDistance) {
                    this.shadeBuf[clipped] += (-this.fogZDistance + model.projectVertexZ[v]) / this.fogZFalloff;
                }
                clipped++;
            } else {
                // behind the near plane: emit near-plane crossings against the previous and next vertex
                int prev = (i != 0) ? verts[i - 1] : verts[numVerts - 1];   // var56 (label276)
                clipped = clipNearCrossing(model, v, prev, shade, clipped);
                int next = (i == numVerts - 1) ? verts[0] : verts[i + 1];   // var56 (label270)
                clipped = clipNearCrossing(model, v, next, shade, clipped);
            }
        }
        // clamp + texture-shift the shade buffer over the face's vertices (clean loops var19 in [0,var17))
        for (int i = 0; i < numVerts; i++) {
            if (this.shadeBuf[i] < 0) this.shadeBuf[i] = 0;
            else if (this.shadeBuf[i] > 255) this.shadeBuf[i] = 255;
            if (poly.objectId >= 0) { // facefill (w.b) is a texture id
                this.shadeBuf[i] = (this.textureDimension[poly.objectId] != 1)
                        ? this.shadeBuf[i] << 6   // 64px texture
                        : this.shadeBuf[i] << 9;  // 128px texture
            }
        }
        // clean: a(0, face, this.B, 0, 0, model, this.yb, this.r, 0, 5960, var54)
        generateScanlines(0, face, this.screenY, 0, 0, model, this.screenX, this.shadeBuf, 0, 5960, clipped);
        if (this.maxY > this.minY) { // Cb > Xb
            // clean: a(this.Vb, model, 1, var17, facefill, this.J, this.Qb, 0, 0)
            rasterize(this.projVertY, model, 1, numVerts, poly.objectId, this.projVertZ, this.projVertX, 0, 0);
        }
    }

    /**
     * Emits one near-plane crossing of the edge ({@code v}, {@code other}) into the clipped screen buffers
     * when {@code other} is in front of the near plane. The crossing's model-space (cc,H) at z==clipNear is
     * found by linear interpolation, then perspective-projected {@code (coord << viewDistance)/clipNear}.
     * Helper for {@link #drawSolidFace}; faithful to the clean base's label276/label270 blocks.
     */
    private int clipNearCrossing(GameModel model, int v, int other, int shade, int count) {
        if (this.cameraClipNear <= model.projectVertexZ[other]) { // ~nb >= ~bb[other]  ==  nb <= bb[other]
            int dz = model.projectVertexZ[v] - model.projectVertexZ[other]; // var50/var51
            int hClip = model.projectVertexY[v]
                    - (model.projectVertexY[v] - model.projectVertexY[other]) * (model.projectVertexZ[v] - this.cameraClipNear) / dz; // var46/var47
            int xClip = model.projectVertexX[v]
                    - (model.projectVertexX[v] - model.projectVertexX[other]) * (model.projectVertexZ[v] - this.cameraClipNear) / dz; // var39/var40
            this.screenX[count] = (xClip << this.viewDistance) / this.cameraClipNear; // yb[var54] = (cc<<R)/nb
            this.screenY[count] = (hClip << this.viewDistance) / this.cameraClipNear; // B[var54]  = (H<<R)/nb
            this.shadeBuf[count] = shade;                                             // r[var54]  = shade
            count++;
        }
        return count;
    }

    /** Blits a 2D billboard sprite face to the surface (with picking). Extracted from {@link #render}. */
    private void drawSpriteFace(GameModel model, int face) {
        int[] fv = model.faceVertices[face];
        int v0 = fv[0];
        int vx = model.vertexViewX[v0]; // pb
        int vy = model.vertexViewY[v0]; // Ob
        int vz = model.projectVertexZ[v0]; // bb
        int w = (this.spriteWidth[face] << this.viewDistance) / vz;  // ob
        int h = (this.spriteHeight[face] << this.viewDistance) / vz; // Eb
        int skew = model.vertexViewX[fv[1]] - vx;     // var63 = pb[fv[1]] - pb[fv[0]]
        int x = vx - w / 2;                            // var20
        int y = this.baseY + vy - h;                   // var21 = Nb - (-vy + h)
        // clean: this.dc.a(scale, spriteId, h, x+baseX, y, w, (byte)29, skew)  (ua.spriteClipping)
        this.surface.spriteClipping((256 << this.viewDistance) / vz, this.spriteId[face], h,
                x + this.baseX, y, w, (byte) 29, skew);
        if (this.mousePickingActive && this.mousePickedCount < this.mousePickedMax) {
            x += (this.spriteTranslateX[face] << this.viewDistance) / vz;
            if (this.mouseY >= y && this.mouseY <= y + h && this.mouseX >= x && this.mouseX <= x + w
                    && !model.unpickable && model.isLocalPlayer[face] == 0) { // ca.db / ca.zb
                this.mousePickedModels[this.mousePickedCount] = model; // Ab[cc]
                this.mousePickedFaces[this.mousePickedCount] = face;    // qb[cc]
                this.mousePickedCount++;                                 // cc++
            }
        }
    }

    // ==================================================================
    // Scanline generation & rasterization
    // ==================================================================

    /**
     * Builds the per-row {@link Scanline} span table (start/end X and shade, 8.8 fixed point) for a
     * convex polygon and performs mouse picking against the filled rows. obf: a(int×2,int[],int×2,ca,
     * int[],int[],int×2,int) (counter {@code Yb}). Triangle ({@code 3}) and quad ({@code 4}) fast paths,
     * generic edge walk otherwise. Writes {@code Xb}/{@code Cb} (minY/maxY) for {@link #rasterize}.
     *
     * <p>Clean parameter order: {@code var1, var2(faceForPick), var3(screenY plane), var4, var5,
     * var6(model), var7(screenX plane), var8(shade), var9, var10(==5960 picking-mode), var11(vertexCount)}.
     */
    private void generateScanlines(int var1, int faceForPick, int[] screenY, int var4, int var5,
                                   GameModel model, int[] screenX, int[] shade, int var9, int pickMode, int vertexCount) {
        int rowCap = this.baseY + this.clipY - 1; // var21 / var96
        if (vertexCount == 3) {
            triEdges(screenY, screenX, shade, rowCap);
        } else if (vertexCount != 4) {
            // --- generic convex polygon ---
            this.maxY = this.minY = screenY[0] += this.baseY; // Cb = Xb = var3[0]+Nb
            for (int k = 1; k < vertexCount; k++) {
                int v = screenY[k] += this.baseY;
                if (v < this.minY) this.minY = v;
                else if (v > this.maxY) this.maxY = v;
            }
            if (this.minY < this.baseY - this.clipY) this.minY = this.baseY - this.clipY;
            if (this.maxY >= this.baseY + this.clipY) this.maxY = this.baseY + this.clipY - 1;
            if (this.minY >= this.maxY) return;
            for (int y = this.minY; y < this.maxY; y++) {
                Scanline sl = this.scanlines[y];
                sl.k = -655360; // endX init to min extreme
                sl.d = 655360;  // startX init to max extreme
            }
            int last = vertexCount - 1;
            walkScanEdge(screenY, screenX, shade, 0, last, true);   // closing edge (bounding)
            for (int k = 0; k < last; k++) {
                walkScanEdge(screenY, screenX, shade, k, k + 1, false);
            }
            if (this.minY < this.baseY - this.clipY) this.minY = this.baseY - this.clipY;
        } else {
            quadEdges(screenY, screenX, shade, rowCap);
        }
        // mouse picking against the filled rows
        if (pickMode == 5960 && this.mousePickingActive && this.mousePickedCount < this.mousePickedMax
                && this.mouseY >= this.minY && this.mouseY < this.maxY) {
            Scanline sl = this.scanlines[this.mouseY];
            if (this.mouseX >= sl.d >> 8 && this.mouseX <= sl.k >> 8 && sl.d <= sl.k
                    && !model.unpickable && model.isLocalPlayer[faceForPick] == 0) {
                this.mousePickedModels[this.mousePickedCount] = model;
                this.mousePickedFaces[this.mousePickedCount] = faceForPick;
                this.mousePickedCount++;
            }
        }
    }

    /**
     * Triangle scanline span builder (the vertexCount==3 fast path of {@link #generateScanlines}).
     * Three edges (0-2, 0-1, 1-2) stepped in 8.8; writes {@code Scanline.d/k/e/l} = startX/endX/startS/endS.
     */
    private void triEdges(int[] screenY, int[] screenX, int[] shade, int rowCap) {
        int y0 = screenY[0] + this.baseY, y1 = screenY[1] + this.baseY, y2 = screenY[2] + this.baseY;
        int x0 = screenX[0], x1 = screenX[1], x2 = screenX[2];
        int s0 = shade[0], s1 = shade[1], s2 = shade[2];
        int e0x = 0, e0s = 0, dx02 = 0, ds02 = 0, lo02 = COLOUR_TRANSPARENT, hi02 = -12345678;
        if (y0 != y2) {
            dx02 = (x2 - x0 << 8) / (y2 - y0);
            ds02 = (s2 - s0 << 8) / (y2 - y0);
            if (y2 < y0) { e0x = x2 << 8; e0s = s2 << 8; lo02 = y2; hi02 = y0; }
            else { e0x = x0 << 8; e0s = s0 << 8; lo02 = y0; hi02 = y2; }
            if (lo02 < 0) { e0x -= dx02 * lo02; e0s -= ds02 * lo02; lo02 = 0; }
            if (hi02 > rowCap) hi02 = rowCap;
        }
        int e1x = 0, e1s = 0, dx01 = 0, ds01 = 0, lo01 = COLOUR_TRANSPARENT, hi01 = -12345678;
        if (y0 != y1) {
            dx01 = (x1 - x0 << 8) / (y1 - y0);
            ds01 = (s1 - s0 << 8) / (y1 - y0);
            if (y1 < y0) { e1x = x1 << 8; e1s = s1 << 8; lo01 = y1; hi01 = y0; }
            else { e1x = x0 << 8; e1s = s0 << 8; lo01 = y0; hi01 = y1; }
            if (lo01 < 0) { e1x -= dx01 * lo01; e1s -= ds01 * lo01; lo01 = 0; }
            if (hi01 > rowCap) hi01 = rowCap;
        }
        int e2x = 0, e2s = 0, dx12 = 0, ds12 = 0, lo12 = COLOUR_TRANSPARENT, hi12 = -12345678;
        if (y1 != y2) {
            dx12 = (x2 - x1 << 8) / (y2 - y1);
            ds12 = (s2 - s1 << 8) / (y2 - y1);
            if (y2 < y1) { e2x = x2 << 8; e2s = s2 << 8; lo12 = y2; hi12 = y1; }
            else { e2x = x1 << 8; e2s = s1 << 8; lo12 = y1; hi12 = y2; }
            if (lo12 < 0) { e2x -= dx12 * lo12; e2s -= ds12 * lo12; lo12 = 0; }
            if (hi12 > rowCap) hi12 = rowCap;
        }
        this.minY = lo02; if (lo01 < this.minY) this.minY = lo01; if (lo12 < this.minY) this.minY = lo12;
        this.maxY = hi02; if (hi01 > this.maxY) this.maxY = hi01; if (hi12 > this.maxY) this.maxY = hi12;
        int endS = 0;
        for (int y = this.minY; y < this.maxY; y++) {
            int sx, ex, ss;
            if (y >= lo02 && y < hi02) { sx = ex = e0x; ss = endS = e0s; e0x += dx02; e0s += ds02; }
            else { sx = 655360; ex = -655360; ss = 0; }
            if (y >= lo01 && y < hi01) {
                if (e1x < sx) { sx = e1x; ss = e1s; }
                if (e1x > ex) { ex = e1x; endS = e1s; }
                e1x += dx01; e1s += ds01;
            }
            if (y >= lo12 && y < hi12) {
                if (e2x < sx) { sx = e2x; ss = e2s; }
                if (e2x > ex) { ex = e2x; endS = e2s; }
                e2x += dx12; e2s += ds12;
            }
            Scanline sl = this.scanlines[y];
            sl.d = sx; sl.k = ex; sl.e = ss; sl.l = endS; // startX,endX,startS,endS
        }
        if (this.minY < this.baseY - this.clipY) this.minY = this.baseY - this.clipY;
    }

    /**
     * Quad scanline span builder (the vertexCount==4 fast path of {@link #generateScanlines}).
     * Four edges (0-3, 0-1, 1-2, 2-3) stepped in 8.8. Faithful to the clean base's second large branch.
     */
    private void quadEdges(int[] screenY, int[] screenX, int[] shade, int rowCap) {
        int y0 = screenY[0] + this.baseY, y1 = this.baseY + screenY[1];
        int y2 = this.baseY + screenY[2], y3 = this.baseY + screenY[3];
        int x0 = screenX[0], x1 = screenX[1], x2 = screenX[2], x3 = screenX[3];
        int s0 = shade[0], s1 = shade[1], s2 = shade[2], s3 = shade[3];
        // edge 0-3
        int e3x = 0, e3s = 0, dx03 = 0, ds03 = 0, lo03 = COLOUR_TRANSPARENT, hi03 = -12345678;
        if (y0 != y3) {
            dx03 = (x3 - x0 << 8) / (y3 - y0);
            ds03 = (s3 - s0 << 8) / (y3 - y0);
            if (y3 <= y0) { lo03 = y3; e3x = x3 << 8; e3s = s3 << 8; hi03 = y0; }
            else { hi03 = y3; e3x = x0 << 8; lo03 = y0; e3s = s0 << 8; }
            if (lo03 < 0) { e3s -= ds03 * lo03; e3x -= lo03 * dx03; lo03 = 0; }
            if (rowCap < hi03) hi03 = rowCap;
        }
        // edge 0-1
        int e1x = 0, e1s = 0, dx01 = 0, ds01 = 0, lo01 = COLOUR_TRANSPARENT, hi01 = -12345678;
        if (y0 != y1) {
            dx01 = (x1 - x0 << 8) / (y1 - y0);
            if (y0 <= y1) { hi01 = y1; e1x = x1 << 8; e1s = s1 << 8; lo01 = y0; }
            else { hi01 = y0; e1x = x0 << 8; lo01 = y1; e1s = s0 << 8; }
            ds01 = (s1 - s0 << 8) / (y1 - y0);
            if (rowCap < hi01) hi01 = rowCap;
            if (lo01 < 0) { e1x -= lo01 * dx01; e1s -= ds01 * lo01; lo01 = 0; }
        }
        // edge 1-2
        int e2x = 0, e2s = 0, dx12 = 0, ds12 = 0, lo12 = COLOUR_TRANSPARENT, hi12 = -12345678;
        if (y2 != y1) {
            dx12 = (x2 - x1 << 8) / (y2 - y1);
            if (y1 <= y2) { hi12 = y2; e2x = x2 << 8; e2s = s2 << 8; lo12 = y1; }
            else { hi12 = y1; e2x = x1 << 8; lo12 = y2; e2s = s1 << 8; }
            ds12 = (s2 - s1 << 8) / (y2 - y1);
            if (lo12 < 0) { e2x -= dx12 * lo12; e2s -= ds12 * lo12; lo12 = 0; }
            if (hi12 > rowCap) hi12 = rowCap;
        }
        // edge 2-3
        int e23x = 0, e23s = 0, dx23 = 0, ds23 = 0, lo23 = COLOUR_TRANSPARENT, hi23 = -12345678;
        if (y3 != y2) {
            dx23 = (x3 - x2 << 8) / (y3 - y2);
            if (y3 >= y2) { hi23 = y3; e23x = x3 << 8; lo23 = y2; e23s = s3 << 8; }
            else { e23x = x2 << 8; hi23 = y2; lo23 = y3; e23s = s2 << 8; }
            ds23 = (s3 - s2 << 8) / (y3 - y2);
            if (lo23 < 0) { e23x -= dx23 * lo23; e23s -= ds23 * lo23; lo23 = 0; }
            if (hi23 > rowCap) hi23 = rowCap;
        }
        this.minY = lo03;
        if (this.minY > lo01) this.minY = lo01;
        if (this.minY > lo12) this.minY = lo12;
        if (this.minY > lo23) this.minY = lo23;
        this.maxY = hi03;
        if (hi01 > this.maxY) this.maxY = hi01;
        if (hi12 > this.maxY) this.maxY = hi12;
        if (hi23 > this.maxY) this.maxY = hi23;
        int endS = 0;
        for (int y = this.minY; y < this.maxY; y++) {
            int sx, ex, ss;
            if (y >= lo03 && y < hi03) { ss = endS = e3s; sx = ex = e3x; e3x += dx03; e3s += ds03; }
            else { sx = 655360; ex = -655360; ss = 0; }
            if (y >= lo01 && y < hi01) {
                if (e1x < sx) { sx = e1x; ss = e1s; }
                if (e1x > ex) { ex = e1x; endS = e1s; }
                e1x += dx01; e1s += ds01;
            }
            if (y >= lo12 && y < hi12) {
                if (e2x < sx) { sx = e2x; ss = e2s; }
                if (e2x > ex) { ex = e2x; endS = e2s; }
                e2x += dx12; e2s += ds12;
            }
            if (y >= lo23 && y < hi23) {
                if (e23x < sx) { sx = e23x; ss = e23s; }
                if (e23x > ex) { ex = e23x; endS = e23s; }
                e23x += dx23; e23s += ds23;
            }
            Scanline sl = this.scanlines[y];
            sl.e = ss; sl.d = sx; sl.k = ex; sl.l = endS;
        }
        if (this.minY < this.baseY - this.clipY) this.minY = this.baseY - this.clipY;
    }

    /**
     * Steps one polygon edge ({@code from}->{@code to}) down the scanline table, updating each row's span
     * min/max (or both, for the bounding closing edge). Helper for the generic path of
     * {@link #generateScanlines}; faithful to the clean per-edge loops (scratch {@code Scanline.d}=startX,
     * {@code .k}=endX, {@code .e}=startS, {@code .l}=endS).
     */
    private void walkScanEdge(int[] y, int[] x, int[] s, int from, int to, boolean bounding) {
        int yA = y[from], yB = y[to];
        if (yA < yB) {
            int dx = (x[to] - x[from] << 8) / (yB - yA);
            int ex = x[from] << 8;
            int ds = (s[to] - s[from] << 8) / (yB - yA);
            int es = s[from] << 8;
            if (yA < 0) { ex -= dx * yA; es -= ds * yA; yA = 0; }
            if (yB > this.maxY) yB = this.maxY;
            for (int row = yA; row <= yB; row++) {
                Scanline sl = this.scanlines[row];
                if (bounding) { sl.d = sl.k = ex; sl.e = sl.l = es; }
                else {
                    if (ex < sl.d) { sl.d = ex; sl.e = es; }
                    if (ex > sl.k) { sl.k = ex; sl.l = es; }
                }
                ex += dx; es += ds;
            }
        } else if (yA > yB) {
            int dx = (x[from] - x[to] << 8) / (yA - yB);
            int ex = x[to] << 8;
            int ds = (s[from] - s[to] << 8) / (yA - yB);
            int es = s[to] << 8;
            if (yB < 0) { ex -= dx * yB; es -= ds * yB; yB = 0; }
            if (yA > this.maxY) yA = this.maxY;
            for (int row = yB; row <= yA; row++) {
                Scanline sl = this.scanlines[row];
                if (bounding) { sl.d = sl.k = ex; sl.e = sl.l = es; }
                else {
                    if (ex < sl.d) { sl.d = ex; sl.e = es; }
                    if (ex > sl.k) { sl.k = ex; sl.l = es; }
                }
                ex += dx; es += ds;
            }
        }
    }

    /**
     * Fills the span table produced by {@link #generateScanlines} into the framebuffer for one face,
     * choosing the textured / flat-gradient scanline writer. obf: a(int[],ca,int,int,int,int[],int[],int,int)
     * (counter {@code F}). The face fill id {@code fill}: {@code -2}=skip, {@code >=0}=textured
     * (perspective-correct), {@code <0}=flat gradient ramp.
     *
     * <p>Clean parameter order: {@code var1(=Vb=projVertY/H), var2(model), var3, var4(vertexCount),
     * var5(fill), var6(=J=projVertZ/bb), var7(=Qb=projVertX/cc), var8, var9}. Inside, {@code var10=var7[]}
     * (cc), {@code var11=var1[]} (H), {@code var12=var6[]} (bb).
     */
    private void rasterize(int[] hArr, GameModel model, int var3, int vertexCount, int fill,
                           int[] bbArr, int[] ccArr, int var8, int var9) {
        if (fill == -2) return;
        if (fill >= 0) {
            if (fill >= this.textureCount) fill = 0; // var5 >= cb -> 0
            prepareTexture(fill, true);
            textureRasterScanlines(model, fill, ccArr, hArr, bbArr, vertexCount);
            return;
        }
        // -------- flat gradient (colour ramp) face --------
        int[] ramp = null;
        for (int r = 0; r < this.rampCount; r++) {
            if (this.gradientBase[r] == fill) { this.currentRamp = this.gradientRamps[r]; ramp = this.currentRamp; break; }
            if (r == this.rampCount - 1) {
                int slot = (int) (Math.random() * this.rampCount);
                this.gradientBase[slot] = fill;
                int rgb = -1 + -fill; // -1 - fill
                int red = ((rgb & 32025) >> 10) * 8;
                int grn = ((rgb & 1019) >> 5) * 8;
                int blu = (rgb & 31) * 8;
                for (int g = 0; g < 256; g++) {
                    int sq = g * g;
                    int rr = red * sq / 65536;
                    int gg = sq * grn / 65536;
                    int bb = blu * sq / 65536;
                    this.gradientRamps[slot][255 - g] = bb + (gg << 8) + (rr << 16);
                }
                this.currentRamp = this.gradientRamps[slot];
                ramp = this.currentRamp;
            }
        }
        gradientRasterScanlines(model, ramp);
    }

    /**
     * Textured-scanline fill of the current span table (the {@code fill>=0} branch of {@link #rasterize}).
     * Computes the perspective-correct texture-coordinate gradients (the {@code <<12}/{@code <<(5-R+…)}
     * fixed-point family from the clean base) and walks each row of {@code [minY,maxY)}, clipping the span
     * to {@code [-clipX, clipX]} and dispatching to the appropriate scanline writer, with interlace row
     * stepping.
     *
     * <p>The clean base inlines a five-way dispatch (one per-row loop per writer): {@code wb.a} (translucent
     * AND back-transparent), {@code gb.a} (translucent), {@code jb.a} (opaque translucent-model variant),
     * {@code p.a} (opaque, not back-transparent) and {@code cb.a} (opaque default tail). The dispatch
     * predicate is {@code ca.Kb} (translucent — this rev's {@code textureTranslucent}; the GameModel deob
     * currently labels {@code Kb} "cleared") and {@code S[fill]} (back-transparent). All five branches are
     * written out below with their real external-writer calls; the writers themselves live on unrelated obf
     * classes ({@code wb}/{@code gb}/{@code jb}/{@code p}/{@code cb}) and are out of this file's scope, but
     * each call site and its argument expressions are reproduced faithfully. Per-row coordinate map (clean):
     * {@code uBase}=var22, {@code uCol}=var19, {@code uStep}=var25, {@code vBase}=var20, {@code vCol}=var23,
     * {@code vStep}=var26, {@code uRow}=var29, {@code vRow}=var28, {@code wRow}=var30, {@code pixOff}=var33.
     */
    private void textureRasterScanlines(GameModel model, int fill, int[] cc, int[] H, int[] bb, int vertexCount) {
        int R = this.viewDistance;
        int var10 = cc[0], var11 = H[0], var12 = bb[0];
        int var13 = var10 - cc[1];          // Δcc to vertex 1
        int var14 = -H[1] + var11;          // ΔH  to vertex 1
        int var15 = var12 - bb[1];          // Δbb to vertex 1
        int last = vertexCount - 1;         // var4-1
        int var16 = cc[last] - var10;       // Δcc to last vertex
        int var17 = -var11 + H[last];       // ΔH  to last vertex
        int var18 = bb[last] - var12;       // Δbb to last vertex
        int uBase, vBase, wBase, uCol, vCol, wCol, uStep, vStep, wStep, uRow, vRow, wRow;
        if (this.textureDimension[fill] == 1) {
            // 128px-wide texture (clean var19..var30; shift exprs verbatim)
            uBase = -(var17 * var10) + var11 * var16 << 12;          // var19
            int vBaseFull = var17 * var12 + -(var18 * var11) << 4 + -R + 5 - -7; // var20
            int wBaseFull = var18 * var10 + -(var16 * var12) << 7 + -R + 5;      // var21
            uCol = var11 * var13 - var10 * var14 << 12;              // var22
            int vColFull = var14 * var12 - var15 * var11 << -R + 5 - -11;        // var23
            int wColFull = -(var12 * var13) + var10 * var15 << 7 + (5 - R);      // var24
            uStep = var16 * var14 + -(var17 * var13) << 5;           // var25
            int vStepFull = var15 * var17 - var14 * var18 << 4 + -R + 5;         // var26
            wStep = var13 * var18 + -(var15 * var16) >> 27 + R;      // var27
            wBase = wBaseFull; wCol = wColFull;                                  // var21 / var24
            vBase = vBaseFull; vCol = vColFull; vStep = vStepFull;
            uRow = vBaseFull >> 4; vRow = vColFull >> 4; wRow = vStepFull >> 4;  // var28/29/30
        } else {
            // 64px-wide texture (clean var72..var89; shift exprs verbatim)
            uBase = var16 * var11 + -(var10 * var17) << 11;          // var72
            int vBaseFull = var12 * var17 + -(var18 * var11) << 4 + 6 + -R + 5;  // var75
            int wBaseFull = var18 * var10 + -(var16 * var12) << 11 - R;          // var76
            uCol = var11 * var13 + -(var14 * var10) << 11;           // var78
            int vColFull = var12 * var14 + -(var11 * var15) << 4 + -R + 11;      // var80
            int wColFull = var15 * var10 - var12 * var13 << 11 + -R;             // var81
            uStep = -(var17 * var13) + var16 * var14 << 5;           // var83
            int vStepFull = var17 * var15 - var14 * var18 << 4 + -R + 5;         // var85
            wStep = var18 * var13 + -(var16 * var15) >> 27 + R;      // var86
            wBase = wBaseFull; wCol = wColFull;                                  // var76 / var81
            vBase = vBaseFull; vCol = vColFull; vStep = vStepFull;
            uRow = vBaseFull >> 4; vRow = vColFull >> 4; wRow = vStepFull >> 4;  // var87/88/89
        }
        int rows = this.minY - this.baseY;          // -Nb + Xb
        int stride = this.width;                    // vb
        int pixOff = this.minY * this.width + this.baseX; // vb*Xb + Zb
        uBase += wBase * rows; uCol += wCol * rows; uStep += wStep * rows;
        byte step = 1;
        if (this.interlace) {
            if ((this.minY & 1) == 1) {
                this.minY++;
                uBase += wBase; uCol += wCol; uStep += wStep; pixOff += stride;
            }
            wBase <<= 1; wCol <<= 1; wStep <<= 1; stride <<= 1; step = 2;
        }
        int[] tex = this.texturePixels[fill];
        boolean translucent = model.cleared;                          // ca.Kb (textureTranslucent)
        boolean backTransparent = this.textureBackTransparent[fill];  // S[fill]

        // -------- writer dispatch (clean label501 for 128px / label499 for 64px), all branches written out --
        // The clean base emits a *separate* per-row span loop per external writer; each loop, once it has
        // drawn its rows, breaks out of the enclosing label so the others are skipped. The net per-row choice
        // depends on BOTH the texture size and the (translucent=ca.Kb, backTransparent=S[fill]) flags:
        //   128px:  !Kb &&  S -> wb.a ;  !Kb && !S -> gb.a ;  Kb -> cb.a
        //    64px:   Kb        -> jb.a ;  !Kb && !S -> p.a  ;  !Kb &&  S -> cb.a
        // All five writers therefore appear; the dim flag selects which trio is live. Writers wb/gb/jb/p/cb
        // are scattered statics on unrelated obf classes (out of this file's scope); each call's argument
        // expression is reproduced verbatim from the clean base.
        boolean is128 = this.textureDimension[fill] == 1;
        for (int y = this.minY; y < this.maxY; y += step) {
            Scanline sl = this.scanlines[y];
            int sx = sl.d >> 8;  // startX (var8)
            int ex = sl.k >> 8;  // endX
            int len = ex - sx;   // var37/var110/...
            if (len <= 0) {
                uBase += wBase; uCol += wCol; uStep += wStep; pixOff += stride;
                continue;
            }
            int s = sl.e;        // startS (var38/var115/...)
            int sStep = (sl.l - s) / len; // (endS-startS)/len (var39/var120/...)
            if (sx < -this.clipX) { s += (-this.clipX - sx) * sStep; sx = -this.clipX; len = ex - sx; }
            if (ex > this.clipX) len = this.clipX - sx;
            int texU = uBase + uRow * sx; // clean: var22 + var8*var29   (128px) / var8*var87 + var72  (64px)
            int texV = uCol + vRow * sx;  // clean: var8*var28 + var19   (128px) / var8*var88 + var78  (64px)
            int texW = uStep + wRow * sx; // clean: var25 + var8*var30   (128px) / var8*var89 + var83  (64px)
            int px = pixOff + sx;         // clean: var8 + var33 (128px) / var8 + var92 (64px)
            if (is128) {
                // 128px dispatch (clean label501)
                if (!translucent && backTransparent) {        // wb.a (back-transparent translucent)
                    // BEHAVIORAL FIX (RENDER_DEOB_GAPS Scene U/V transpose; clean lb.java:2066-2067):
                    // clean wb.a arg7 = var8*var28+var19 = texU, arg8 = var22+var8*var29 = texV.
                    // Deob previously passed (texV, texU) here -> texture sampled transposed.
                    // obf wb.a -> MessageList.drawTextureSpan (16-arg; obf arg order kept)
                    MessageList.drawTextureSpan(vCol, 10, 0, 0, this.raster, texW, s, texU, texV, px, vStep, sStep, 0, vBase, tex, len);
                } else if (!translucent) {                    // gb.a (translucent)
                    // BEHAVIORAL FIX (RENDER_DEOB_GAPS Scene U/V transpose; clean lb.java:2128/2136):
                    // clean gb.a arg0 = var22+var29*var8 = texV, arg8 = var8*var28+var19 = texU.
                    // Deob previously passed (texU,...,texV) -> transposed.
                    // obf gb.a -> ProxySocketFactory._junkRasterScanline (15-arg; obf arg order kept)
                    ProxySocketFactory._junkRasterScanline(texV, vBase, (byte) 50, texW, s, sStep << 2, tex, px, texU, vStep, 0, 0, this.raster, vCol, len);
                } else {                                       // cb.a (opaque)
                    // obf cb.a -> CacheUpdater.drawTexturedScanlinePerspective (16-arg; obf arg order kept)
                    CacheUpdater.drawTexturedScanlinePerspective(len, texW, 0, (byte) 25, 0, vBase, vStep, sStep << 2, tex, this.raster, px, texU, 0, vCol, s, texV);
                }
            } else {
                // 64px dispatch (clean label499)
                if (translucent) {                            // jb.a (translucent-model opaque)
                    // obf jb.a -> DownloadWorker.drawTexturedSpanUnrolled (15-arg; obf arg order kept)
                    DownloadWorker.drawTexturedSpanUnrolled(this.raster, vCol, vStep, texW, sStep, s, px, len, texU, 0, tex, false, vBase, texV, 0);
                } else if (!backTransparent) {                // p.a (opaque, not back-transparent)
                    // obf p.a -> Timer.renderAffineSpan (15-arg; obf arg order kept)
                    Timer.renderAffineSpan(sStep, 1121159302, vCol, texV, vBase, tex, s, 0, texU, 0, this.raster, px, texW, vStep, len);
                } else {                                       // cb.a (opaque)
                    // obf cb.a -> CacheUpdater.drawTexturedScanlinePerspective (16-arg; obf arg order kept)
                    CacheUpdater.drawTexturedScanlinePerspective(len, texW, 0, (byte) 25, 0, vBase, vStep, sStep << 2, tex, this.raster, px, texU, 0, vCol, s, texV);
                }
            }
            uBase += wBase; uCol += wCol; uStep += wStep; pixOff += stride;
        }
    }

    /**
     * Flat-gradient scanline fill of the current span table (the {@code fill<0} branch of
     * {@link #rasterize}). Walks each row of {@code [minY,maxY)}, clips the span to {@code [-clipX,clipX]}
     * and dispatches to one of the three scattered gradient-scanline writers, with interlace row stepping.
     * The clean base (lines ~2542-2677) emits three separate per-row loops gated by {@code ca.cb}
     * (model.transparent -> {@code ua.a}) and {@code Ub} (wideBand -> skip {@code t.a}); the default tail is
     * {@code ia.a}. All three writer call sites are reproduced here verbatim (the writers themselves are on
     * unrelated obf classes {@code ua}/{@code t}/{@code ia}, out of this file's scope). Clean per-row map:
     * {@code pixOff}=var44, {@code stride}=var42, {@code sx}=var8, {@code s}=var63/64/65, {@code len}=var59/60/61,
     * {@code sStep}=var67/68/69. {@code this.currentRamp} (obf H) is the resolved colour ramp.
     */
    private void gradientRasterScanlines(GameModel model, int[] ramp) {
        int stride = this.width;
        int pixOff = this.minY * this.width + this.baseX;
        byte step = 1;
        if (this.interlace) {
            if ((this.minY & 1) == 1) { this.minY++; pixOff += stride; }
            stride <<= 1; step = 2;
        }
        boolean transparent = model.transparent; // ca.cb
        for (int y = this.minY; y < this.maxY; y += step) {
            Scanline sl = this.scanlines[y];
            int sx = sl.d >> 8;
            int ex = sl.k >> 8;
            int len = ex - sx;
            if (len <= 0) { pixOff += stride; continue; }
            int s = sl.e;
            int sStep = (sl.l - s) / len;
            if (sx < -this.clipX) { s += (-this.clipX - sx) * sStep; sx = -this.clipX; len = ex - sx; }
            if (ex > this.clipX) len = this.clipX - sx;
            int px = pixOff + sx; // var8 + var44
            if (transparent) {                  // clean: if (var2.cb) -> ua.a (transparent gradient)
                // obf ua.a -> Surface.buildShadeRamp (8-arg gradient, transparent); arg order kept
                Surface.buildShadeRamp(s, this.currentRamp, -len, this.raster, 0, sStep, px, 0); // clean var3-1, var3==1 -> 0
            } else if (!this.wideBand) {        // clean: if (!this.Ub) -> t.a (opaque gradient)
                // obf t.a -> EntityDef.fillPixelColumns16 (8-arg gradient); arg order kept
                EntityDef.fillPixelColumns16(0, sStep, -len, this.raster, this.currentRamp, s, px, 418609192);
            } else {                            // clean default tail -> ia.a (wide-band gradient)
                // obf ia.a -> SpriteScaler.writePaletteScaledScanline (dead trailing byte param dropped)
                SpriteScaler.writePaletteScaledScanline(sStep, 0, this.currentRamp, s, px, this.raster, -len);
            }
            pixOff += stride;
        }
    }

    // ==================================================================
    // Image-producer flush
    // ==================================================================

    /**
     * Pushes the freshly-rendered raster into the AWT {@code ImageConsumer} so the applet repaints.
     * obf: a(boolean,byte[]) (counter {@code w}). Static; uses {@code StringCodec.DEAD_IMAGE_CONSUMER} (the Surface ImageConsumer),
     * {@code World.surfaceWidth} (image width on World), {@code ClientStream.imageHeight} (image height) and {@code SocketFactory.colorModel} (colour model).
     * The boolean gates the body (anti-tamper); {@code pixels} is the byte source.
     */
    public static void flushToImage(boolean go, byte[] pixels) {
        if (!go) return;
        if (StringCodec.DEAD_IMAGE_CONSUMER != null) { // obf: StringCodec.DEAD_IMAGE_CONSUMER (ImageConsumer)
            StringCodec.DEAD_IMAGE_CONSUMER.setPixels(0, 0, World.surfaceWidth, ClientStream.imageHeight, SocketFactory.globalColorModel, pixels, 0, World.surfaceWidth); // World.surfaceWidth width, ClientStream.imageHeight height, SocketFactory.globalColorModel colourModel
            StringCodec.DEAD_IMAGE_CONSUMER.imageComplete(3);
        }
    }

    // ==================================================================
    // XOR string-pool decoder (kept faithful for documentation)
    // ==================================================================

    /**
     * Decodes the per-class XOR string pool. obf: the two {@code z(...)} helpers. The first applies a
     * 1-char key (^42) when the string has length &lt; 2; the second XORs each char by the 5-entry byte
     * table {124,75,38,66,42} indexed by position mod 5. The resolved values are reproduced here.
     */
    private static String[] decodeStringPool() {
        return new String[] {
            "{...}", "null", "lb.F(", "lb.LA(", "lb.AA(", "lb.H(", "lb.A(",
            "Warning tried to add null object!", "lb.NA(", "lb.N(", "lb.B(", "lb.E(",
            "lb.U(", "lb.OA(", "lb.EA(", "lb.G(", "lb.JA(", "lb.V(", "lb.MA(", "lb.C(",
            "lb.I(", "lb.T(", "lb.L(", "lb.GA(", "lb.P(", "lb.HA(", "lb.R(", "lb.O(",
            "lb.S(", "lb.IA(", "lb.D(", "lb.<init>(", "lb.KA(", "lb.K(", "lb.M(", "lb.BA(",
            "lb.J(", "lb.CA(", "lb.W(", "lb.Q(", "lb.FA(", "lb.DA("
        };
    }
}
