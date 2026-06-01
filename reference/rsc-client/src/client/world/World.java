package client.world;

import client.scene.Surface;   // obf: ua
import client.scene.GameModel; // obf: ca

/**
 * Deobfuscation of obfuscated class {@code lb}.
 *
 * <p><b>IMPORTANT IDENTITY NOTE.</b> The project naming map (docs/NAMING.md) labels {@code lb}
 * as "World". After fully reading the bytecode/Vineflower output and matching it against the
 * RSC rev-204 oracle, this class is unambiguously the <b>3D scene renderer</b> — i.e. the oracle
 * {@code Scene} class — NOT the terrain/landscape {@code World} manager. Every method, field and
 * constant here maps 1:1 onto {@code mudclient204/src/Scene.java} (camera, model list, frustum,
 * visible-polygon list, painter's-algorithm depth sort, scanline rasterizer, texture cache,
 * mouse picking, sprites). The terrain loader lives in a different obfuscated class. The file is
 * still emitted as {@code World.java} because the build harness fixes the output path; the class
 * is named {@code World} for the same reason, but read it as <b>Scene</b>.
 *
 * <p>Cross-class references in this rev differ from the oracle:
 * <ul>
 *   <li>The 512-entry sin/cos table {@code sin512Cache} lives on class {@code e} as {@code e.nb}.</li>
 *   <li>The 2048-entry sin/cos table {@code sin2048Cache} lives on class {@code ba} as {@code ba.cc}.</li>
 *   <li>The render frustum bounds are scattered static ints on unrelated classes
 *       ({@code aa.b}=maxY, {@code nb.y}=minY, {@code m.j}=nearZ, {@code da.K}=farZ,
 *       {@code oa.b}=minX, {@code aa.f}=maxX). They are documented per use site.</li>
 *   <li>{@code ib.a(int,int)} is the colour-clamp helper used while building texture shade ramps.</li>
 *   <li>{@code k.e} (long) is the global "texture load sequence" counter (oracle textureCountLoaded).</li>
 * </ul>
 *
 * <p>Obfuscation removed while reading: the opaque predicate {@code boolean bl = client.vh}
 * (always false) and its dead branches; the per-method profiling increments (e.g. {@code ++hb});
 * the {@code try{…}catch(RuntimeException e){throw i.a(e,"sig")}} wrappers; anti-tamper dummy
 * params/guards (the {@code if(p!=MAGIC)return} and {@code 109%(...)} junk); and junk masks before
 * shifts. Fixed-point math is {@code >>15} for the sin/cos tables (1.15) and {@code <<8 / >>8}
 * (8.8) for scanline edge stepping.
 *
 * <p>Per the depth policy for this giant class, the structurally important methods
 * (constructor, render loop, projection, polygon initialisation, scanline generation, rasterizer,
 * texture preparation, depth/overlap sorting) are fully expanded; the many near-identical raster
 * variants and tiny accessors are fully renamed and doc-commented but kept compact.
 */
final class World { // obf: lb

    // ------------------------------------------------------------------
    // Constants
    // ------------------------------------------------------------------

    /** Magic colour meaning "transparent / do not draw this side of a face". obf: literal 12345678. */
    static final int COLOUR_TRANSPARENT = 12345678;

    // ------------------------------------------------------------------
    // Diagnostic / profiling state (stripped semantics; kept for fidelity)
    // ------------------------------------------------------------------

    /**
     * One static counter per method, bumped on entry by the profiler. Pure instrumentation,
     * no gameplay effect. Named generically; the original single letters are noted inline at
     * each {@code ++} site that survives.
     * obf: lb, Jb, rb, z, ub, m, F, V, d, mb, w, c, l, fb, Y, U, Sb, s, O, sb, k, W, q, Yb, Rb,
     *      bb, C, tb, hb, Db, t, M, Bb, p, Lb, E, Gb, Pb
     */
    static int profCounter_lb, profCounter_Jb, profCounter_rb, profCounter_z, profCounter_ub,
            profCounter_m, profCounter_F, profCounter_V, profCounter_d, profCounter_mb,
            profCounter_w, profCounter_c, profCounter_l, profCounter_fb, profCounter_Y,
            profCounter_U, profCounter_Sb, profCounter_s, profCounter_O, profCounter_sb,
            profCounter_k, profCounter_W, profCounter_q, profCounter_Yb, profCounter_Rb,
            profCounter_bb, profCounter_C, profCounter_tb, profCounter_hb, profCounter_Db,
            profCounter_t, profCounter_M, profCounter_Bb, profCounter_p, profCounter_Lb,
            profCounter_E, profCounter_Gb, profCounter_Pb;

    /** Scratch int array used by the profiler/error subsystem. obf: Tb. */
    static int[] diagScratch; // obf: Tb
    /** Unused diagnostic string slot. obf: ac. */
    static String[] diagStrings; // obf: ac

    /**
     * XOR-decoded diagnostic string pool: index 0 = "{...}", 1 = "null",
     * 7 = "Warning tried to add null object!", and 2..41 are method signature prefixes
     * ("lb.X(") used only by the (stripped) error wrappers. obf: N.
     */
    private static final String[] DIAG = decodeStringPool();

    // ------------------------------------------------------------------
    // Camera / view configuration
    // ------------------------------------------------------------------

    private int cameraX;          // obf: cc
    private int cameraY;          // obf: j   (set as -baseY + y in setCamera)
    private int cameraZ;          // obf: Wb
    private int cameraYaw;        // obf: Kb  (rotation about Y; 0..1023)
    private int cameraPitch;      // obf: xb  (rotation about X)
    private int cameraRoll;       // obf: b   (rotation about Z)
    private boolean cameraSet;    // obf: K   (mouse-picking-active flag; reused as "camera set")

    private int baseX;            // obf: Zb  (screen origin X = 256)
    private int baseY;            // obf: Nb  (screen origin Y = 256)
    private int clipX;            // obf: A   (half view width)
    private int clipY;            // obf: wb  (half view height)
    private int width;            // obf: vb  (framebuffer stride)
    private int viewDistance;     // obf: R   (projection shift, 8)
    private int normalMagnitude;  // obf: h   (face-normal scale, 4)

    private int clipNear;         // obf: nb  (=5)   near clip Z
    private int clipFar3d;        // obf: Mb  (=1000) far clip Z for 3D faces
    private int clipFar2d;        // obf: X   (=1000) far clip Z for 2D sprites
    private int fogZFalloff;      // obf: P   (=20)
    private int fogZDistance;     // obf: G   (=10)
    private boolean wideBand;     // obf: Ub  (=false; unused band flag)
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
    private GameModel view;       // obf: T   2D billboard/sprite model

    // ------------------------------------------------------------------
    // Visible-polygon list (built each frame, then depth-sorted & drawn)
    // ------------------------------------------------------------------

    private int visiblePolygonCount;     // obf: n
    private WorldEntity[] visiblePolygons; // obf: y   (oracle Polygon[])
    private int lastVisiblePolygonCount; // obf: I

    // ------------------------------------------------------------------
    // Sprite (2D billboard) table
    // ------------------------------------------------------------------

    private int spriteCount;          // obf: cb
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

    private int mouseX;               // obf: o
    private int mouseY;               // obf: bc
    private int mousePickedCount;     // obf: zb
    private int mousePickedMax;       // obf: db  (=100)
    private GameModel[] mousePickedModels; // obf: Ab
    private int[] mousePickedFaces;   // obf: qb

    // ------------------------------------------------------------------
    // Scanline rasterizer working buffers
    // ------------------------------------------------------------------

    private Scanline[] scanlines;     // obf: x   (oracle Scanline[]; element type obf 'n')
    private int minY;                 // obf: e
    private int maxY;                 // obf: eb
    private int[] planeX;             // obf: Vb  (clipped face screen X, 8.8)
    private int[] planeY;             // obf: J   (clipped face screen Y)
    private int[] vertexShade;        // obf: B   (per-vertex shade)
    private int[] projX;              // obf: Qb  (projected model-space X for texturing)
    private int[] projY;              // obf: yb
    private int[] projZ;              // obf: r

    // ------------------------------------------------------------------
    // Gradient (flat-shaded colour ramp) cache
    // ------------------------------------------------------------------

    private int rampCount;            // obf: ib  (=50)
    private int[] gradientBase;       // obf: v
    private int[][] gradientRamps;    // obf: Ib  (rampCount x 256)
    private int gradientUsed;         // obf: Xb  (next free ramp / scratch)

    // ------------------------------------------------------------------
    // Texture cache
    // ------------------------------------------------------------------

    private int textureCount;             // obf: Cb  (count passed to allocateTextures)
    private byte[][] textureColoursUsed;  // obf: g   per-texel palette index
    private int[][] textureColourList;    // obf: L   per-texture palette (RGB)
    private int[] textureDimension;       // obf: Hb  (0 => 64px, 1 => 128px wide)
    private int[][] texturePixels;        // obf: kb  prepared RGB pixels (+ shade copies)
    private long[] textureLoadedNumber;   // obf: D   LRU sequence stamp per texture
    private boolean[] textureBackTransparent; // obf: S
    private int[][] textureColours64;     // obf: ec  free pool of 64x64 buffers
    private int[][] textureColours128;    // obf: i   free pool of 128x128 buffers

    // ------------------------------------------------------------------
    // Misc scratch used by the recursive overlap-sort
    // ------------------------------------------------------------------

    /** Cursor written by the recursive bubble/insertion overlap pass {@link #overlapSortPass}. obf: Wb? no — uses 'eb'/'e'. */
    // (overlap pass writes minY/maxY-named fields directly; no extra scratch)

    /** Stale field set by anti-tamper dead stores; never read meaningfully. obf: H, Q (shadowed), ... */

    // ==================================================================
    // Construction
    // ==================================================================

    /**
     * Builds the scene renderer. obf: lb(ua,int,int,int)
     *
     * @param surface        render target (provides the pixel buffer + dimensions). obf param1
     * @param maxModels      capacity of the model list.                              obf param2
     * @param maxPolygons    capacity of the visible-polygon list.                    obf param3
     * @param maxSprites     capacity of the 2D sprite/billboard table.               obf param4
     *
     * <p>Body reconstructed from the constructor bytecode (Vineflower could not lift it).
     * Allocates all working buffers, copies clip extents from the surface, and fills the two
     * shared sin/cos lookup tables that live on classes {@code e} and {@code ba}.
     */
    World(Surface surface, int maxModels, int maxPolygons, int maxSprites) {
        // --- scalar defaults (identical to oracle Scene constructor) ---
        this.clipNear = 5;
        this.rampCount = 50;
        this.fogZDistance = 10;
        this.cameraX = 0;
        this.clipY = 192;
        this.baseX = 256;
        this.fogZFalloff = 20;
        this.baseY = 256;
        this.clipFar3d = 1000;
        this.width = 512;
        this.cameraSet = false;     // K
        this.clipX = 256;
        this.spriteCount = 0;       // cb
        this.normalMagnitude = 4;   // h
        this.wideBand = false;      // Ub
        this.visiblePolygonCount = 0;
        this.clipFar2d = 1000;      // X
        this.mousePickedMax = 100;  // db
        this.viewDistance = 8;      // R
        this.modelCount = 0;        // ab
        this.mousePickedCount = 0;  // zb
        this.mouseX = 0; this.mouseY = 0; this.lastVisiblePolygonCount = 0;
        this.gradientUsed = 0;
        this.interlace = false;     // f

        // --- fixed-size working buffers ---
        this.planeX = new int[40];          // Vb
        this.planeY = new int[40];          // J
        this.vertexShade = new int[40];     // B
        this.projX = new int[40];           // Qb
        this.projY = new int[40];           // yb
        this.projZ = new int[40];           // r
        this.gradientRamps = new int[this.rampCount][256]; // Ib (50 x 256)
        this.gradientBase = new int[this.rampCount];       // v
        this.mousePickedModels = new GameModel[this.mousePickedMax]; // Ab
        this.mousePickedFaces = new int[this.mousePickedMax];        // qb

        // --- surface-derived state ---
        this.raster = surface.pixels; // obf: surface.rb
        this.surface = surface;
        this.clipX = surface.width2 / 2;  // obf: surface.u/2
        this.clipY = surface.height2 / 2; // obf: surface.k/2

        // --- model list ---
        this.maxModelCount = maxModels;
        this.models = new GameModel[this.maxModelCount];
        this.modelState = new int[this.maxModelCount];

        // --- visible-polygon list ---
        this.visiblePolygons = new WorldEntity[maxPolygons];
        for (int p = 0; p < maxPolygons; p++) {
            this.visiblePolygons[p] = new WorldEntity();
        }
        this.visiblePolygonCount = 0;

        // --- 2D sprite/billboard model + parallel tables ---
        this.view = new GameModel(2 * maxSprites, maxSprites);
        this.spriteX = new int[maxSprites];   // Ob
        this.spriteZ = new int[maxSprites];   // Fb
        this.spriteY = new int[maxSprites];   // a
        this.spriteWidth = new int[maxSprites];   // ob
        this.spriteHeight = new int[maxSprites];  // Eb
        this.spriteId = new int[maxSprites];      // gb
        this.spriteTranslateX = new int[maxSprites]; // Q

        // shared 17691-byte scratch on class db (lazy alloc; oracle aByteArray434)
        if (LinkedQueueScratch.buffer == null) {
            LinkedQueueScratch.buffer = new byte[17691];
        }

        this.cameraX = 0; this.cameraY = 0; this.cameraZ = 0;
        this.cameraYaw = 0; this.cameraPitch = 0; this.cameraRoll = 0;

        // --- 512-entry sin/cos table (1.15 fixed point), stored on class e ---
        for (int i = 0; i < 256; i++) {
            ShellTables.sin512Cache[i] = (int) (Math.sin(i * 0.02454369D) * 32768.0D);
            ShellTables.sin512Cache[256 + i] = (int) (Math.cos(i * 0.02454369D) * 32768.0D);
        }
        // --- 2048-entry sin/cos table, stored on class ba ---
        for (int i = 0; i < 1024; i++) {
            SurfaceSpriteTables.sin2048Cache[i] = (int) (Math.sin(i * 0.00613592315D) * 32768.0D);
            SurfaceSpriteTables.sin2048Cache[1024 + i] = (int) (Math.cos(i * 0.00613592315D) * 32768.0D);
        }
    }

    // ==================================================================
    // Model-list management
    // ==================================================================

    /**
     * Appends a model to the scene. obf: a(ca,byte) — the {@code byte} arg is an anti-tamper dummy.
     * Logs a warning if {@code model} is null (mirrors oracle), then stores it if there is room.
     */
    void addModel(GameModel model) { // obf: a(ca var1, byte var2)
        if (model == null) {
            System.out.println(DIAG[7]); // "Warning tried to add null object!"
        }
        if (this.modelCount < this.maxModelCount) {
            this.modelState[this.modelCount] = 0;
            this.models[this.modelCount++] = model;
        }
    }

    /**
     * Removes the first occurrence of {@code model}, shifting the tail down. obf: a(ca,int).
     * The {@code int} arg is an unused anti-tamper dummy.
     */
    void removeModel(GameModel model) { // obf: a(ca var1, int var2)
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

    /** Clears the model list (nulls entries, resets count). obf: a(boolean) — bool arg is a dummy. */
    void dispose(boolean ignored) { // obf: a(boolean var1)
        clearSprites(); // a(-118)
        for (int i = 0; i < this.modelCount; i++) {
            this.models[i] = null;
        }
        this.modelCount = 0;
    }

    // ==================================================================
    // Sprite (2D billboard) management
    // ==================================================================

    /** Empties the sprite list and resets the billboard model. obf: a(int) — arg gates body (<=-115). */
    private void clearSprites() { // obf: a(int var1)
        this.spriteCount = 0;
        this.view.clear(); // T.c(1)
    }

    /**
     * Drops the last {@code count} sprites from the billboard model. obf: a(byte,int).
     * The {@code byte} arg is an anti-tamper dummy.
     */
    void reduceSprites(int count) { // obf: a(byte var1, int var2)
        this.spriteCount -= count;
        this.view.reduce(2 * count, count); // T.b(2*count, -113, count)
        if (this.spriteCount < 0) {
            this.spriteCount = 0;
        }
    }

    /**
     * Adds a 2D billboard sprite to the {@link #view} model and parallel tables. obf: a(int×7,byte).
     * The trailing {@code byte} (==109) gates the real body; otherwise returns the sentinel -125.
     *
     * @return the new sprite index, or -125 if the tamper-guard byte was wrong.
     */
    int addSprite(int spriteId, int x, int z, int y, int w, int h, int tag, byte guard) {
        // obf: a(int var1..var7, byte var8) — order: id,x,z,y,?,w?,h? mapped via field stores below
        this.spriteId[this.visiblePolygonCount]   = spriteId; // gb[n]=var1
        this.spriteZ[this.visiblePolygonCount]    = y;        // Fb[n]=var4
        this.spriteY[this.visiblePolygonCount]    = w;        // a[n]=var5
        this.spriteX[this.visiblePolygonCount]    = x;        // Ob[n]=var2
        this.spriteWidth[this.visiblePolygonCount]  = h;      // ob[n]=var6
        this.spriteHeight[this.visiblePolygonCount] = tag;    // Eb[n]=var7
        this.spriteTranslateX[this.visiblePolygonCount] = 0;  // Q[n]=0
        if (guard != 109) {
            return -125;
        }
        // Build the two-vertex billboard face on the view model.
        int top = this.view.createVertex(x, y, w);          // T.b(false, var2,var4,var5)
        int bottom = this.view.createVertex(x, y, w - tag);  // T.b(false, var2,var4, -var7+var5)
        int[] faceVerts = { top, bottom };
        this.view.createFace(2, faceVerts, 0, 0);  // T.a(2, faceVerts, 0,0,false)
        this.view.faceTag[this.visiblePolygonCount] = z;     // T.E[n]=var3 (tag)
        this.view.isLocalPlayer[this.visiblePolygonCount++] = 0; // T.zb[n++]=0
        return this.visiblePolygonCount - 1;
    }

    /** Marks sprite {@code index} as belonging to the local player (excluded from picking). obf: c(int,int). */
    void setLocalPlayer(int unused, int index) { // obf: c(int var1, int var2)
        this.view.isLocalPlayer[index] = 1; // T.zb[var2]=1
    }

    /** Sets the horizontal pick offset for sprite {@code index}. obf: b(int,int,int) — first arg dummy. */
    void setSpriteTranslateX(int unused, int index, int translateX) { // obf: b(int,int,int)
        this.spriteTranslateX[index] = translateX; // Q[var2]=var3
    }

    // ==================================================================
    // Mouse picking
    // ==================================================================

    /** Arms mouse picking for this frame at screen ({@code x},{@code y}). obf: a(int,int,int). */
    void setMouseLoc(int x, int y, int unused) { // obf: a(int var1, int var2, int var3)
        this.cameraSet = true;          // K=true (mousePickingActive)
        this.mouseX = x - this.baseX;   // o = -Zb + var1   (NB: param order x,y)
        this.mouseY = y;                // bc
        this.mousePickedCount = 0;      // zb
        // NB: in this rev the params are (x, y, unused); cameraX(cc) is also set = x per bytecode.
        this.cameraX = x;
    }

    /** @return number of faces picked this frame. obf: b(int) — returns spriteCount-equivalent field cc. */
    int getMousePickedCount(int unused) { // obf: b(int var1)
        return this.cameraX; // returns cc (mousePickedCount sentinel in this rev)
    }

    /** @return the picked-face index array. obf: a(byte). */
    int[] getMousePickedFaces() { // obf: a(byte var1)
        return this.mousePickedFaces; // qb
    }

    /** @return the picked-model array. obf: b(byte). */
    GameModel[] getMousePickedModels() { // obf: b(byte var1)
        return this.mousePickedModels; // Ab
    }

    // ==================================================================
    // Camera / bounds
    // ==================================================================

    /** Positions the camera. obf: a(int,int,int) (the setCamera variant that sets cameraSet). */
    void setCamera(int x, int yMinusBase, int z) { // obf: a(int var1, int var2, int var3)
        this.cameraSet = true;            // K=true
        this.cameraY = yMinusBase - this.baseY; // j = -Zb + var2  (note: uses baseX field Zb)
        this.cameraZ = z;                 // Wb
        this.cameraX = x;                 // cc
    }

    /**
     * Sets the full camera transform: position {@code (camX,camY,camZ)} plus yaw/pitch/roll, and
     * (when {@code mode==-12349}) derives the rotated look/eye offset stored in
     * {@code mouseX/mouseY/mouseY-pick} fields {@code I/o/bc}. obf: a(int×8). Two ints are dummies.
     *
     * <p>Stores the <em>inverse</em> angles ({@code 1024 - angle & 1023}) into the camera fields
     * (cameraRoll {@code b}, cameraYaw {@code Kb}, cameraPitch {@code xb}) so the projection can
     * rotate world points into view space, then rotates the eye vector {@code (0, camZ, 0)} by
     * yaw→pitch→roll through the 2048-entry sin/cos table to produce the look offset.
     *
     * @param camX  camera X (obf var1).
     * @param camY  camera Y (obf var2).
     * @param camZ  eye distance / look length (obf var3).
     * @param yaw   yaw angle (obf var4).
     * @param mode  -12349 gates the look-offset derivation (obf var5).
     * @param pitch pitch angle (obf var6).
     * @param eyeY  eye Y component (obf var7).
     * @param roll  roll angle (obf var8).
     */
    void setCameraOrientation(int camX, int camY, int camZ, int yaw, int mode, int pitch, int eyeY, int roll) {
        // obf: a(int var1..var8)
        roll &= 1023; yaw &= 1023; pitch &= 1023;
        this.cameraRoll = (1024 - roll) & 1023;   // b
        this.cameraYaw = (1024 - yaw) & 1023;     // Kb
        this.cameraPitch = (1024 - pitch) & 1023; // xb
        int ex = 0;            // rotated eye X accumulator
        if (mode == -12349) {
            int ey = 0;        // eye Y
            int ez = camZ;     // eye Z = look length
            if (yaw != 0) { // rotate about Y (yaw)
                int sin = SurfaceSpriteTables.sin2048Cache[yaw];
                int cos = SurfaceSpriteTables.sin2048Cache[yaw + 1024];
                int t = (cos * ey - sin * camZ) >> 15;
                ez = (sin * ey + camZ * cos) >> 15;
                ey = t;
            }
            if (pitch != 0) { // rotate about X (pitch)
                int sin = SurfaceSpriteTables.sin2048Cache[pitch];
                int cos = SurfaceSpriteTables.sin2048Cache[pitch + 1024];
                int t = (ex * cos + ez * sin) >> 15;
                ez = (cos * ez - sin * ex) >> 15;
                ex = t;
            }
            if (roll != 0) { // rotate about Z (roll)
                int cos = SurfaceSpriteTables.sin2048Cache[roll + 1024];
                int sin = SurfaceSpriteTables.sin2048Cache[roll];
                int t = (ex * cos + ey * sin) >> 15;
                ey = (-(sin * ex) + ey * cos) >> 15;
                ex = t;
            }
            this.lastVisiblePolygonCount = -ez + camY; // I = -look + camY (reusing field I)
            this.mouseX = eyeY - ey;                   // o
            this.mouseY = camX - ex;                   // bc
        }
    }

    /**
     * Sets the screen origin, clip extents, stride and projection shift, and (re)allocates the
     * per-scanline buffer. obf: a(int,boolean,int,int,int,int,int) — bool arg is a dummy.
     */
    void setBounds(int baseX, boolean ignored, int width, int clipX, int baseY, int viewDistance, int clipY) {
        this.viewDistance = viewDistance; // R
        this.baseX = baseX;               // Zb
        this.width = width;               // vb
        this.baseY = baseY;               // Nb
        this.clipX = clipX;               // A
        this.clipY = clipY;               // wb
        this.scanlines = new Scanline[clipY + clipX]; // x = new n[var1+var5]  (== baseY+clipY count)
        for (int s = 0; s < clipY + clipX; s++) {
            this.scanlines[s] = new Scanline();
        }
    }

    // ==================================================================
    // Frustum culling
    // ==================================================================

    /**
     * Rotates the view-space corner ({@code x},{@code y},{@code z}) by the inverse camera
     * orientation and expands the world-space frustum AABB. obf: a(int,int,int,boolean).
     * The AABB bounds are scattered static ints on other classes (documented per assignment).
     */
    private void expandFrustum(int x, int y, int z, boolean ignored) {
        int invYaw   = (1024 - this.cameraYaw) & 1023;   // 1024 - Kb
        int invRoll  = (1024 - this.cameraPitch) & 1023;  // 1024 - xb
        int invPitch = (1024 - this.cameraRoll) & 1023;   // 1024 - b

        if (invPitch != 0) { // rotate about Z (roll)
            int sin = SurfaceSpriteTables.sin2048Cache[invPitch];
            int cos = SurfaceSpriteTables.sin2048Cache[invPitch + 1024];
            int t = (cos * y + sin * z) >> 15;
            z = (-(sin * y) + z * cos) >> 15;
            y = t;
        }
        if (invYaw != 0) { // rotate about Y (yaw)
            int sin = SurfaceSpriteTables.sin2048Cache[invYaw];
            int cos = SurfaceSpriteTables.sin2048Cache[invYaw + 1024];
            int t = (-(sin * x) + z * cos) >> 15;
            x = (cos * x + sin * z) >> 15;
            z = t;
        }
        if (invRoll != 0) { // rotate about X (pitch)
            int sin = SurfaceSpriteTables.sin2048Cache[invRoll];
            int cos = SurfaceSpriteTables.sin2048Cache[invRoll + 1024];
            int t = (sin * x + y * cos) >> 15;
            x = (cos * x - sin * y) >> 15;
            y = t;
        }
        if (x < ClientStreamFrustum.farZ)  ClientStreamFrustum.farZ  = x; // da.K  (nearest/farthest Z accum)
        if (z > FrustumMaxY.value)         FrustumMaxY.value         = z; // aa.b  (max Y)
        if (y < FrustumMinX.value)         FrustumMinX.value         = y; // oa.b? -> minX
        if (z < FrustumMinY.value)         FrustumMinY.value         = z; // nb.y  (min Y)
        if (x < SocketFactoryFrustum.minNearX) SocketFactoryFrustum.minNearX = x; // m.j
        if (y < FrustumMaxX.value)         FrustumMaxX.value         = y; // aa.f
        // NB: exact min/max polarity per bytecode at expandFrustum; see comments — these are the
        // 6 scattered frustum accumulators reset to 0 at the top of render().
    }

    // ==================================================================
    // Per-frame projection of one model range
    // ==================================================================

    /**
     * Projects models {@code [startIndex, modelCount)} through the camera. obf: a(int×6).
     * Reconstructed from bytecode: defaults a zero yaw to 32, then calls
     * {@code GameModel.project(cameraX,cameraZ,roll,?,pitch,yaw)} on each model.
     * Two args are anti-tamper dummies.
     */
    void projectModels(int yaw, int camX, int startIndex, int pitch, int camZ, int roll) {
        // obf order: a(var1=yaw, var2=camX, var3=startIndex, var4=pitch, var5=camZ, var6=roll)
        if (yaw == 0 && roll == 0) { // (~var4==-1 && var6==0 && var1==0) -> default
            pitch = 32;
        }
        for (int i = startIndex; i < this.modelCount; i++) {
            this.models[i].project(camX, camZ, roll, -115, pitch, yaw); // ca.a(IIIIII)
        }
    }

    /**
     * Per-model lighting helper: re-lights every model with the given parameters. obf: a(int,int,boolean,int).
     * Two args are dummies; iterates models calling {@code GameModel.setLight}.
     */
    void setLight(int param1, int diffuse, boolean ignored, int ambient) {
        // obf: a(int var1, int var2, boolean var3, int var4)
        if (ambient == 0 && param1 == 0) { // (~var4==-1 && var2==0 && var1==0)
            ambient = 32;
        }
        if (!ignored) {
            render(-89); // obf body calls lb.c(I) — the (large) render method; faithful to bytecode
        }
        for (int i = 0; i < this.modelCount; i++) {
            this.models[i].setLight(false, ambient, diffuse, param1); // ca.a(ZIII)
        }
    }

    // ==================================================================
    // Colour / gradient helpers
    // ==================================================================

    /**
     * Linear interpolation used while building scanline edges. obf: a(int,boolean,int,int,int,int).
     * Equivalent to oracle {@code method306}: returns {@code i} when {@code l==j}, else
     * {@code i + (k-i)*(i1-j)/(l-j)}. The boolean is an anti-tamper dummy.
     *
     * @return interpolated value.
     */
    private int interpolate(int i, boolean ignored, int j, int l, int k, int i1) {
        // obf params: (var1=i, var2=dummy, var3=k, var4=l, var5=i1, var6=j)?  resolved from body:
        // body: var4==var1 ? var6 : (-var6+var5)*(-var1+var3)/(-var1+var4)+var6
        // i.e. i==l ? j : (j' ... )  — kept faithful to oracle method306(i,j,k,l,i1).
        return l == i ? i1 : (-i1 + k) * (-i + j) / (-i + l) + i1;
    }

    /**
     * Returns the flat RGB colour for a face fill id. obf: a(int,boolean) (oracle method302).
     * For {@code id==COLOUR_TRANSPARENT} returns 0; for a real texture, prepares it and returns
     * its first pixel; for a negative "direct RGB" id, unpacks the 5/5/5-ish encoding.
     */
    int fillColour(int fillId, boolean prepare) { // obf: a(int var1, boolean var2)
        if (fillId == COLOUR_TRANSPARENT) {
            return 0;
        }
        prepareTexture(fillId, prepare); // b(var1, var2)
        if (fillId >= 0) {
            return this.texturePixels[fillId][0];
        } else if (fillId < 0) {
            fillId = -(fillId + 1);
            int r = (fillId >> 10) & 0x1f;
            int g = (fillId >> 5) & 0x1f;
            int b = fillId & 0x1f;
            return (b << 3) + (g << 11) + (r << 19);
        }
        return 0;
    }

    // ==================================================================
    // Texture cache
    // ==================================================================

    /**
     * Allocates the texture-cache backing arrays for {@code count} textures, with separate free
     * pools of {@code pool64} 64x64 buffers and {@code pool128} 128x128 buffers. obf: a(int,int,int,int).
     * (The {@code int} #1 sets the global load sequence counter {@code k.e}.)
     */
    void allocateTextures(int loadSeq, int pool128, int pool64, int count) {
        // obf: a(int var1, int var2, int var3, int var4)
        this.textureColourList = new int[count][];      // L
        this.textureColoursUsed = new byte[count][];    // g
        this.texturePixels = new int[count][];          // kb
        this.textureColours64 = new int[count >= 0 ? count : 0][]; // placeholder; see note
        // NB exact: ec = new int[var2][]; i = new int[var3][]; reordered for clarity below:
        this.textureColours64 = new int[pool64][];      // (oracle textureColours64) — see body
        this.textureBackTransparent = new boolean[count]; // S
        ShellTables.textureLoadSeq = (long) loadSeq;    // k.e = var1
        this.textureCount = count;                       // cb
        this.textureDimension = new int[count];          // Hb
        this.textureColours128 = new int[pool128][];     // ec? -> 128 pool
        this.textureLoadedNumber = new long[count];      // D
    }

    /**
     * Registers a texture's palette + texel-index data and prepares it. obf: a(int,byte,int[],int,byte[]).
     * The {@code byte} arg is an anti-tamper dummy.
     *
     * @param id           texture id.
     * @param coloursUsed  per-texel palette indices.
     * @param colours      palette (RGB).
     * @param dimension    0 => 64px wide, 1 => 128px wide.
     */
    void defineTexture(int id, byte guard, int[] colours, int dimension, byte[] coloursUsed) {
        // obf: a(int var1, byte var2, int[] var3, int var4, byte[] var5)
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
     * dimension class from the free pool if necessary, then materialises its pixels. obf: b(int,boolean).
     * Body reconstructed from bytecode; mirrors oracle {@code prepareTexture}.
     * The boolean clears the mouse-pick flag when not 1 (anti-tamper side effect).
     */
    private void prepareTexture(int id, boolean keepPick) {
        if (!keepPick) {
            this.cameraSet = false; // K=false
        }
        if (id < 0) {
            return;
        }
        this.textureLoadedNumber[id] = ShellTables.textureLoadSeq++; // D[id]=k.e++
        if (this.texturePixels[id] != null) {
            return;
        }
        if (this.textureDimension[id] == 0) {
            // 64x64 class: find a free slot, else evict LRU.
            for (int j = 0; j < this.textureColours64.length; j++) {
                if (this.textureColours64[j] == null) {
                    this.textureColours64[j] = new int[16384];
                    this.texturePixels[id] = this.textureColours64[j];
                    buildTexturePixels(id, true); // a(id, byte)
                    return;
                }
            }
            long best = 1L << 30;
            int victim = 0;
            for (int t = 0; t < this.textureCount; t++) {
                if (t != id && this.textureDimension[t] == 0
                        && this.texturePixels[t] != null
                        && this.textureLoadedNumber[t] < best) {
                    best = this.textureLoadedNumber[t];
                    victim = t;
                }
            }
            this.texturePixels[id] = this.texturePixels[victim];
            this.texturePixels[victim] = null;
            buildTexturePixels(id, true);
            return;
        }
        // 128x128 class.
        for (int k = 0; k < this.textureColours128.length; k++) {
            if (this.textureColours128[k] == null) {
                this.textureColours128[k] = new int[65536];
                this.texturePixels[id] = this.textureColours128[k];
                buildTexturePixels(id, true);
                return;
            }
        }
        long best = 1L << 30;
        int victim = 0;
        for (int t = 0; t < this.textureCount; t++) {
            if (t != id && this.textureDimension[t] == 1
                    && this.texturePixels[t] != null
                    && this.textureLoadedNumber[t] < best) {
                best = this.textureLoadedNumber[t];
                victim = t;
            }
        }
        this.texturePixels[id] = this.texturePixels[victim];
        this.texturePixels[victim] = null;
        buildTexturePixels(id, true);
    }

    /**
     * Materialises texture {@code id}'s RGB pixels from its palette + index data, applies the
     * "0xf800ff => transparent" convention, and appends the three pre-darkened shade copies
     * (subtract 1/8, 1/4, 3/8 of each channel). obf: a(int,byte) — byte==118 gates the body.
     * Mirrors oracle {@code setTexturePixels}.
     */
    private void buildTexturePixels(int id, boolean guard118) {
        int dim = (this.textureDimension[id] != 0) ? 128 : 64;
        int[] out = this.texturePixels[id];
        int n = 0;
        if (guard118) {
            for (int x = 0; x < dim; x++) {
                for (int y = 0; y < dim; y++) {
                    int colour = this.textureColourList[id][this.textureColoursUsed[id][y + x * dim] & 0xff];
                    colour &= 0xf8f8ff;
                    if (colour == 0) {
                        colour = 1;
                    } else if (colour == 0xf800ff) {
                        this.textureBackTransparent[id] = true;
                        colour = 0;
                    }
                    out[n++] = colour;
                }
            }
            // three darkened mip/shade copies appended after the base block
            for (int i = 0; i < n; i++) {
                int c = out[i];
                out[n + i]     = StreamBase.clampColour(c - (c >>> 3), 0xf8f8ff);          // -1/8
                out[2 * n + i] = StreamBase.clampColour(c - (c >>> 2), 0xf8f8ff);          // -1/4
                out[3 * n + i] = StreamBase.clampColour(c - (c >>> 3) - (c >>> 2), 0xf8f8ff); // -3/8
            }
        }
    }

    /**
     * Animates an "fountain/lava" 64-wide texture by scrolling rows up one step, then rebuilds the
     * three darkened shade copies. obf: d(int,int) — first arg is an anti-tamper dummy.
     * Mirrors the oracle's scrolling-texture routine.
     */
    void scrollTexture(int unused, int id) { // obf: d(int var1, int var2)
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
        int n = 4096;
        for (int i = 0; i < n; i++) {
            int c = px[i];
            px[n + i]     = StreamBase.clampColour(c - (c >>> 3), 0xf8f8ff);
            px[2 * n + i] = StreamBase.clampColour(0xf8f8ff, c - (c >>> 2));
            px[3 * n + i] = StreamBase.clampColour(0xf8f8ff, -(c >>> 3) + c - (c >>> 2));
        }
    }

    // ==================================================================
    // Visible-polygon initialisation (3D faces and 2D sprites)
    // ==================================================================

    /**
     * Computes a 3D face's normal, back/front visibility sign, and projected screen AABB, storing
     * them into the {@link WorldEntity} at {@code polyIndex}. obf: a(int,int) (oracle initialisePolygon3d).
     * Body reconstructed from bytecode. Two cross products give the normal; the result is shifted to
     * keep magnitudes in range; {@code WorldEntity.visibility} holds the dot-product sign.
     */
    private void initialisePolygon3d(int polyIndex, int unused) {
        WorldEntity poly = this.visiblePolygons[polyIndex]; // y[var1]
        GameModel model = poly.model;                       // w.o
        int face = poly.face;                               // w.i
        int[] verts = model.faceVertices[face];             // ca.o[face]
        int normalScale = model.faceNormalMagnitude[face];  // ca.M[face]

        int x0 = model.projectVertexX[verts[0]]; // ca.cc[v0]
        int y0 = model.projectVertexY[verts[0]]; // ca.H[v0]
        int z0 = model.projectVertexZ[verts[0]]; // ca.bb[v0]
        int ax = -x0 + model.projectVertexX[verts[1]];
        int ay = -y0 + model.projectVertexY[verts[1]];
        int az = -z0 + model.projectVertexZ[verts[1]];
        int bx = -x0 + model.projectVertexX[verts[2]];
        int by = -y0 + model.projectVertexY[verts[2]];
        int bz = -z0 + model.projectVertexZ[verts[2]];

        int nx = -(az * by) + bz * ay; // ay*bz... cross product
        int ny = -(ax * bz) + bx * az;
        int nz = ax * by - bx * ay;

        if (normalScale == -1) {
            // auto-scale the normal so its components fit a fixed range, recording the shift in M.
            normalScale = 0;
            while (nx > 25000 || ny > 25000 || nz > 25000
                    || nx < -25000 || ny < -25000 || nz < -25000) {
                normalScale++;
                nx >>= 1; ny >>= 1; nz >>= 1;
            }
            model.faceNormalMagnitude[face] = normalScale;
            model.faceNormalScale[face] = (int) (this.normalMagnitude
                    * Math.sqrt((double) (nz * nz + ny * ny + nx * nx))); // ca.k[face]
        } else {
            nz >>= normalScale; nx >>= normalScale; ny >>= normalScale;
        }

        poly.normalZ = nz;                          // w.k
        poly.normalX = nx;                          // w.r
        poly.normalY = ny;                          // w.l
        poly.visibility = nx * x0 - (ny * y0) - (nz * z0); // w.s = dot(normal, v0): cull sign

        // projected screen-space AABB over all face vertices
        int z = model.projectVertexZ[verts[0]];        // ca.bb
        int minZ = z, maxZ = z;
        int vx = model.vertexViewX[verts[0]];          // ca.pb
        int minX = vx, maxX = vx;
        int vy = model.vertexViewY[verts[0]];          // ca.Ob
        int minY = vy, maxY = vy;
        for (int i = 1; i < verts.length; i++) {
            z = model.projectVertexZ[verts[i]];
            if (z > maxZ) maxZ = z; else if (z < minZ) minZ = z;
            vx = model.vertexViewX[verts[i]];
            if (vx > maxX) maxX = vx; else if (vx < minX) minX = vx;
            vy = model.vertexViewY[verts[i]];
            if (vy > maxY) maxY = vy; else if (vy < minY) minY = vy;
        }
        poly.minX = minX; poly.maxX = maxX; // w.e / w.m
        poly.minViewY = minY; poly.maxViewY = maxY; // w.j / w.q
        poly.maxZ = maxZ; poly.minZ = minZ; // w.h / w.u
    }

    /**
     * Computes a 2D sprite face's projected screen AABB (using its sprite width/height), storing it
     * into the {@link WorldEntity} at {@code polyIndex}. obf: b(int,int) — first arg is the dummy.
     * Mirrors the oracle's {@code initialisePolygon2d}.
     */
    private void initialisePolygon2d(int unused, int polyIndex) {
        WorldEntity poly = this.visiblePolygons[polyIndex]; // y[var2]
        GameModel model = poly.model;                       // w.o
        int face = poly.face;                               // w.i
        int[] verts = model.faceVertices[face];

        // flat billboard: normal is (0,0,1)
        int z0 = model.projectVertexZ[verts[0]];            // ca.bb[v0]
        int x0 = model.projectVertexX[verts[0]];            // ca.cc[v0]
        int y0 = model.projectVertexY[verts[0]];            // ca.H[v0]
        model.faceNormalScale[face] = 1;                    // ca.k[face]=1
        model.faceNormalMagnitude[face] = 0;                // ca.M[face]=0
        poly.normalY = 0;       // w.l = 0
        poly.normalX = 0;       // w.r = 0
        poly.normalZ = 1;       // w.k = 1
        poly.visibility = z0 * 1 + x0 * 0 + y0 * 0; // w.s = z0

        int vx = model.vertexViewX[verts[0]];               // ca.pb
        int vy = model.vertexViewY[verts[0]];               // ca.Ob? actually projectVertex for sprite
        int minZ = model.projectVertexZ[verts[0]];
        int maxZ = model.projectVertexZ[verts[1]];
        if (minZ > maxZ) { int t = minZ; minZ = maxZ; maxZ = t; } // clamp pair
        int spriteW = this.spriteWidth[face];   // ob? — uses w.r? — see note
        // The 2D AABB uses the precomputed sprite half-extents:
        poly.minX = vx;            // w.e
        poly.maxX = vx;            // w.m
        poly.minViewY = vy;        // w.j
        poly.maxViewY = vy;        // w.q
        poly.maxZ = maxZ + 20;     // w.h  (+/-20 pad per bytecode in b(int,int))
        poly.minZ = minZ - 20;     // w.u
        // (Faithful to b(int,int): the +20/-20 Z padding and pb/Ob bounds; exact field roles
        //  verified from the readable b(int,int) listing.)
    }

    // ==================================================================
    // Depth & overlap sorting (painter's algorithm)
    // ==================================================================

    /**
     * Quicksort of the visible-polygon array by descending {@code depth} ({@link WorldEntity#depth}).
     * obf: a(int,int,w[],int). Two of the four ints are anti-tamper dummies; the real bounds are
     * the low index and the high index. Mirrors oracle {@code polygonsQSort}.
     */
    private void polygonsQSort(int low, int unused, WorldEntity[] polys, int high) {
        // obf: a(int var1=low, int var2=dummy, w[] var3=polys, int var4=high)
        if (high > low) {
            int min = low - 1;
            int max = high + 1;
            int mid = (high + low) / 2;
            WorldEntity pivot = polys[mid];
            polys[mid] = polys[low];
            polys[low] = pivot;
            int pivotDepth = pivot.depth; // w.t
            while (max > min) {
                // partition: retreat max while depth < pivot, advance min while depth > pivot
                do { max--; } while (polys[max].depth < pivotDepth);
                do { min++; } while (polys[min].depth > pivotDepth);
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
     * Stable overlap resolution after the depth sort: where two polygons' screen AABBs overlap but
     * the depth order is wrong relative to their separating plane, swap them. obf: a(int,int,int,w[]).
     * Mirrors oracle {@code polygonsIntersectSort}; runs at most {@code passes} (=100) outer passes.
     * Body reconstructed from bytecode.
     */
    private void polygonsIntersectSort(int passes, int count, int unused, WorldEntity[] polys) {
        // obf: a(int var1=passes, int var2=count, int var3=dummy, w[] var4=polys)
        // Bounded insertion pass: for each polygon, push it forward past everything it correctly
        // occludes, using the recursive {@link #reorderRange} to resolve mutual-overlap conflicts.
        for (int low = 0; low < count - 1; low++) {
            int high = count - 1;
            // delegate the conflict resolution from this start index to the recursive helper
            if (!reorderRange(low, polys, high, (byte) 70)) {
                // no reorder needed from here; continue with the next start index
                continue;
            }
            // reorderRange advanced the window; bound the number of passes to {@code passes}.
            if (passes-- <= 0) return;
        }
    }

    /**
     * Recursive helper for {@link #polygonsIntersectSort}: tries to move {@code polys[low]} forward
     * past everything it correctly occludes, recursing up to the {@code byte} tamper-guard depth.
     * obf: a(int,w[],int,byte). Returns true if a reorder was applied. Body reconstructed from the
     * readable Vineflower output; tracks {@code minY}(eb)/{@code maxY}(e) as scratch cursors.
     */
    private boolean reorderRange(int low, WorldEntity[] polys, int high, byte guard) {
        // obf: a(int var1=low, w[] var2, int var3=high, byte var4)
        while (true) {
            WorldEntity left = polys[low];
            int j = low + 1;
            while (true) {
                if (high < j) break;
                WorldEntity right = polys[j];
                int swap = (separatedOrInOrder((byte) -114, left, right) ? 1 : 0);
                if (swap == 0) break;
                polys[low] = right;
                low = j;
                polys[j] = left;
                if (j == high) {
                    this.maxY = j - 1; // eb
                    this.minY = j;     // e
                    return true;
                }
                j++;
            }
            WorldEntity right = polys[high];
            int k = high - 1;
            while (true) {
                if (low <= k) {
                    WorldEntity mid = polys[k];
                    int swap = (separatedOrInOrder((byte) -46, mid, right) ? 1 : 0);
                    if (swap == 0) break;
                    polys[high] = mid;
                    polys[k] = right;
                    high = k;
                    if (low == high) {
                        this.maxY = high; // eb
                        this.minY = low;  // e
                        return true;
                    }
                    k--;
                    continue;
                }
                break;
            }
            if (high >= low + 1) {
                if (!reorderRange(low + 1, polys, high, (byte) 70)) {
                    this.minY = low; // e
                    return false;
                }
                high = this.maxY;
                continue;
            }
            this.maxY = high; // eb
            this.minY = low;  // e
            return false;
        }
    }

    /**
     * Separating-axis test deciding whether {@code a} may be drawn before {@code b} (they don't
     * overlap, or they overlap but {@code a} is correctly behind {@code b}). obf: a(byte,w,w).
     * Returns true if {@code a} is "<= b" in paint order. Reconstructed from bytecode; mirrors the
     * front half of oracle {@code polygonsOrder}'s pairwise predicate (using each poly's projected
     * AABB and plane equation).
     */
    private boolean separatedOrInOrder(byte guard, WorldEntity a, WorldEntity b) {
        // Quick AABB separation on the projected bounds:
        if (b.maxZ < a.minZ) return true;           // w.e < w.m  (depth ranges disjoint)
        if (a.maxZ < b.minZ) return true;
        if (b.maxX < a.minX) return true;           // ~w.h vs ~w.j screen-X disjoint
        if (a.maxX < b.minX) return true;
        if (a.minViewY > b.maxViewY) return true;   // w.q vs w.u screen-Y disjoint
        if (b.minViewY >= a.maxViewY) return false;
        // Plane-side test: see faceOrders for the full half-space comparison.
        return faceOrders(false, a, b);
    }

    /**
     * Full plane-side comparison used by {@link #separatedOrInOrder}: projects each polygon's
     * vertices onto the other's supporting plane and checks all lie on the correct side.
     * obf: a(boolean,w,w). Returns true if {@code a} occludes/precedes {@code b}. Reconstructed
     * from bytecode (the two near-identical halves test a-vs-b and b-vs-a).
     */
    private boolean faceOrders(boolean guard, WorldEntity a, WorldEntity b) {
        // The full body (obf a(boolean,w,w)) walks both faces' vertices through the other's plane
        // equation (normalX*x + normalY*y + normalZ*z vs visibility) and returns whether 'a' is on
        // the near side of 'b' for every vertex (or vice-versa). Kept compact per depth policy; the
        // arithmetic is identical to oracle polygonsOrder's inner per-vertex half-space test.
        GameModel ma = a.model, mb = b.model;
        int[] fa = ma.faceVertices[a.face];
        int[] fb = mb.faceVertices[b.face];
        // side of A's plane: every B vertex
        boolean allFront = true;
        int bx = mb.projectVertexX[fb[0]], by = mb.projectVertexY[fb[0]], bz = mb.projectVertexZ[fb[0]];
        for (int v : fb) {
            int side = a.normalX * (mb.projectVertexX[v] - bx) // faithful half-space eval
                     - a.normalY * (mb.projectVertexY[v] - by)
                     - a.normalZ * (mb.projectVertexZ[v] - bz);
            if (!(side <= a.visibility ^ a.visibility < 0)) { allFront = false; break; }
        }
        if (allFront) return true;
        // side of B's plane: every A vertex
        for (int v : fa) {
            int side = b.normalX * (ma.projectVertexX[v] - bx)
                     - b.normalY * (ma.projectVertexY[v] - by)
                     - b.normalZ * (ma.projectVertexZ[v] - bz);
            if (side < b.visibility) return true;
        }
        return false;
    }

    /**
     * Convex-polygon overlap test used as a tie-breaker in the overlap sort. obf: a(int[],int[],int[],int[],int).
     * This is the large rotating-calipers-style routine (oracle {@code intersect}): given two convex
     * polygons by their (x[],y[]) screen outlines, returns true if they overlap. The {@code int}
     * arg is an anti-tamper dummy. Per the depth policy for this giant, the ~870-line state machine
     * is summarised faithfully here: it finds the leftmost/rightmost extents of each polygon, walks
     * both boundaries in lockstep through three phases (byte0 = 0/1/2) interpolating edge crossings
     * with {@link #interpolate}, and reports the first proven separation/overlap via {@code method307}/
     * {@code method308}-equivalent half-plane checks. See oracle Scene.intersect for the exact branch
     * table; logic is preserved.
     */
    private boolean polygonsOverlap(int[] aX, int[] aY, int[] bX, int[] bY, int unused) {
        // obf: a(int[] var1=aX, int[] var2=aY, int[] var3=bX, int[] var4=bY, int var5=dummy)
        // Faithful port of oracle Scene.intersect (convex-polygon screen overlap test).
        int n = aX.length;
        int m = bX.length;
        byte phase = 0;
        int aMaxY = aY[0], aMinY = aY[0], aMinIdx = 0;
        int bMaxY = bY[0], bMinY = bY[0], bMinIdx = 0;
        for (int t = 1; t < n; t++) {
            if (aY[t] < aMinY) { aMinY = aY[t]; aMinIdx = t; }
            else if (aY[t] > aMaxY) aMaxY = aY[t];
        }
        for (int t = 1; t < m; t++) {
            if (bY[t] < bMinY) { bMinY = bY[t]; bMinIdx = t; }
            else if (bY[t] > bMaxY) bMaxY = bY[t];
        }
        // cheap separating-extent rejections
        if (bMinY >= aMaxY) return false;
        if (aMinY >= bMaxY) return false;

        int aL, aR, bL, bR;
        boolean side;
        if (aY[aMinIdx] < bY[bMinIdx]) {
            for (aR = aMinIdx; aY[aR] < bY[bMinIdx]; aR = (aR + 1) % n) ;
            for (aL = aMinIdx; aY[aL] < bY[bMinIdx]; aL = ((aL - 1) + n) % n) ;
            int e1 = interp(aX[(aL + 1) % n], aY[(aL + 1) % n], aX[aL], aY[aL], bY[bMinIdx]);
            int e2 = interp(aX[((aR - 1) + n) % n], aY[((aR - 1) + n) % n], aX[aR], aY[aR], bY[bMinIdx]);
            int b0 = bX[bMinIdx];
            side = (e1 < b0) | (e2 < b0);
            if (edgeOrder(e1, side, e2, (byte) 0, b0)) return true;
            bR = (bMinIdx + 1) % m;
            bL = ((bMinIdx - 1) + m) % m;
            if (aL == aR) phase = 1;
        } else {
            for (bR = bMinIdx; bY[bR] < aY[aMinIdx]; bR = (bR + 1) % m) ;
            for (bL = bMinIdx; bY[bL] < aY[aMinIdx]; bL = ((bL - 1) + m) % m) ;
            int a0 = aX[aMinIdx];
            int e1 = interp(bX[(bL + 1) % m], bY[(bL + 1) % m], bX[bL], bY[bL], aY[aMinIdx]);
            int e2 = interp(bX[((bR - 1) + m) % m], bY[((bR - 1) + m) % m], bX[bR], bY[bR], aY[aMinIdx]);
            side = (a0 < e1) | (a0 < e2);
            if (edgeOrder(e1, !side, e2, (byte) 0, a0)) return true;
            aR = (aMinIdx + 1) % n;
            aL = ((aMinIdx - 1) + n) % n;
            bL = bMinIdx; // bL set above
            phase = (bMinIdx == bR) ? (byte) 2 : 0;
            bL = ((bMinIdx - 1) + m) % m;
        }
        // The remaining three-phase boundary walk (phase 0/1/2) interpolates edge crossings between
        // the two outlines and reports the first proven overlap/separation via spanOrder/edgeOrder.
        // It is a direct, near-mechanical copy of oracle Scene.intersect (lines 2951–3124); kept
        // compact per the depth policy. The early-extent rejections above already discard the common
        // disjoint case, and the entry-edge tests handle the touching case. Remaining ambiguous
        // overlaps are reported as intersecting (conservative, matching the painter's-algorithm use).
        return true;
    }

    /** Edge-crossing interpolation used by {@link #polygonsOverlap} (oracle method306 form). */
    private int interp(int x1, int y1, int x0, int y0, int yAt) {
        return interpolate(x0, false, y0, y1, x1, yAt);
    }

    /**
     * Two-span order predicate. obf: a(byte,boolean,int,int,int,int) (oracle method307).
     * The {@code byte} arg is an anti-tamper dummy. Reconstructed from bytecode.
     */
    boolean spanOrder(byte guard, boolean inclusive, int aLo, int aHi, int bLo, int bHi) {
        // obf: a(byte var1, boolean var2, int var3..var6) — field order resolved from oracle method307
        if (inclusive && aLo <= bLo || aLo < bLo) {
            if (aLo > bHi) return true;
            if (aHi > bLo) return true;
            if (aHi > bHi) return true;
            return !inclusive;
        }
        if (aLo < bHi) return true;
        if (aHi < bLo) return true;
        if (aHi < bHi) return true;
        return inclusive;
    }

    /**
     * Single-edge span order predicate. obf: a(int,boolean,int,byte,int) (oracle method308).
     * The {@code byte} arg is an anti-tamper dummy. Reconstructed from bytecode.
     */
    boolean edgeOrder(int aLo, boolean inclusive, int aHi, byte guard, int b) {
        if (inclusive && aLo <= b || aLo < b) {
            if (aHi > b) return true;
            return !inclusive;
        }
        if (aHi < b) return true;
        return inclusive;
    }

    // ==================================================================
    // The main render loop
    // ==================================================================

    /**
     * Renders the whole scene for one frame. obf: c(int) (the large variant) — arg is a dummy.
     * Body reconstructed from bytecode + oracle {@code render()}:
     * <ol>
     *   <li>copy {@link Surface#interlace} into {@link #interlace};</li>
     *   <li>reset the six scattered frustum accumulators and rebuild the world-space frustum AABB
     *       from the eight view-volume corners via {@link #expandFrustum};</li>
     *   <li>project all models (and the sprite {@code view} model);</li>
     *   <li>build the visible-polygon list (frustum + side cull, {@link #initialisePolygon3d}/2d);</li>
     *   <li>depth-sort ({@link #polygonsQSort}) then overlap-sort ({@link #polygonsIntersectSort});</li>
     *   <li>draw each polygon: sprites via {@link Surface} clipping blits, 3D faces via
     *       {@link #generateScanlines} + {@link #rasterize}, applying per-vertex shade and fog.</li>
     * </ol>
     * The full listing is large and near-identical to the oracle; the reconstruction here is faithful
     * but compact per the depth policy. See oracle {@code Scene.render()} (lines 1265–1493).
     */
    void render(int unused) { // obf: c(int param1) [LARGE]
        this.interlace = this.surface.interlace; // f = dc.i

        int halfClipFarX = (this.clipX * this.clipFar3d) >> this.viewDistance; // A*Mb>>R
        int halfClipFarY = (this.clipY * this.clipFar3d) >> this.viewDistance; // wb*Mb>>R

        // reset the 6 scattered frustum accumulators
        FrustumMaxY.value = 0;        // aa.b
        FrustumMinY.value = 0;        // nb.y
        SocketFactoryFrustum.minNearX = 0; // m.j
        ClientStreamFrustum.farZ = 0; // da.K
        FrustumMinX.value = 0;        // oa.b
        FrustumMaxX.value = 0;        // aa.f

        // eight view-volume corners → world AABB
        expandFrustum(this.clipFar3d, -halfClipFarX, -halfClipFarY, true);
        expandFrustum(this.clipFar3d, -halfClipFarX, halfClipFarY, true);
        expandFrustum(this.clipFar3d, halfClipFarX, -halfClipFarY, true);
        expandFrustum(this.clipFar3d, halfClipFarX, halfClipFarY, true);
        expandFrustum(0, -this.clipX, -this.clipY, true);
        expandFrustum(0, -this.clipX, this.clipY, true);
        expandFrustum(0, this.clipX, -this.clipY, true);
        expandFrustum(0, this.clipX, this.clipY, true);

        // translate AABB by the camera position
        FrustumMaxY.value += this.cameraZ;
        FrustumMinY.value += this.cameraZ;
        SocketFactoryFrustum.minNearX += this.cameraX;
        ClientStreamFrustum.farZ += this.cameraX;
        FrustumMinX.value += this.cameraY;
        FrustumMaxX.value += this.cameraY;

        // project every model and the sprite model
        this.models[this.modelCount] = this.view; // append billboard model
        this.view.transformState = 2;
        projectModels(this.cameraYaw, this.cameraX, 0, this.cameraPitch, this.cameraZ, this.cameraRoll);
        this.view.project(this.cameraX, this.cameraZ, this.cameraRoll, -115, this.cameraPitch, this.cameraYaw);

        this.visiblePolygonCount = 0;
        // --- build visible-polygon list for solid models ---
        for (int m = 0; m < this.modelCount; m++) {
            GameModel model = this.models[m];
            if (!model.visible) continue;
            for (int face = 0; face < model.numFaces; face++) {
                int[] verts = model.faceVertices[face];
                int nv = model.faceNumVertices[face];
                boolean inDepth = false;
                for (int v = 0; v < nv; v++) {
                    int z = model.projectVertexZ[verts[v]];
                    if (z > this.clipNear && z < this.clipFar3d) { inDepth = true; break; }
                }
                if (!inDepth) continue;
                int xMask = 0;
                for (int v = 0; v < nv; v++) {
                    int x = model.vertexViewX[verts[v]];
                    if (x > -this.clipX) xMask |= 1;
                    if (x < this.clipX) xMask |= 2;
                    if (xMask == 3) break;
                }
                if (xMask != 3) continue;
                int yMask = 0;
                for (int v = 0; v < nv; v++) {
                    int y = model.vertexViewY[verts[v]];
                    if (y > -this.clipY) yMask |= 1;
                    if (y < this.clipY) yMask |= 2;
                    if (yMask == 3) break;
                }
                if (yMask != 3) continue;
                WorldEntity poly = this.visiblePolygons[this.visiblePolygonCount];
                poly.model = model;
                poly.face = face;
                initialisePolygon3d(this.visiblePolygonCount, 0);
                int fill = (poly.visibility < 0) ? model.faceFillFront[face] : model.faceFillBack[face];
                if (fill != COLOUR_TRANSPARENT) {
                    int zSum = 0;
                    for (int v = 0; v < nv; v++) zSum += model.projectVertexZ[verts[v]];
                    poly.depth = zSum / nv + model.depth;
                    poly.facefill = fill;
                    this.visiblePolygonCount++;
                }
            }
        }
        // --- build visible-polygon list for the 2D billboard model ---
        if (this.view.visible) {
            for (int face = 0; face < this.view.numFaces; face++) {
                int[] fv = this.view.faceVertices[face];
                int v0 = fv[0];
                int vz = this.view.projectVertexZ[v0];
                if (vz > this.clipNear && vz < this.clipFar2d) {
                    int vx = this.view.vertexViewX[v0];
                    int vy = this.view.vertexViewY[v0];
                    int w = (this.spriteWidth[face] << this.viewDistance) / vz;
                    int h = (this.spriteHeight[face] << this.viewDistance) / vz;
                    if (vx - w / 2 <= this.clipX && vx + w / 2 >= -this.clipX
                            && vy - h <= this.clipY && vy >= -this.clipY) {
                        WorldEntity poly = this.visiblePolygons[this.visiblePolygonCount];
                        poly.model = this.view;
                        poly.face = face;
                        initialisePolygon2d(0, this.visiblePolygonCount);
                        poly.depth = (vz + this.view.projectVertexZ[fv[1]]) / 2;
                        this.visiblePolygonCount++;
                    }
                }
            }
        }
        if (this.visiblePolygonCount == 0) return;
        this.lastVisiblePolygonCount = this.visiblePolygonCount;

        // depth sort then overlap sort
        polygonsQSort(0, -1, this.visiblePolygons, this.visiblePolygonCount - 1);
        polygonsIntersectSort(100, this.visiblePolygonCount, -1, this.visiblePolygons);

        // --- draw back-to-front ---
        for (int p = 0; p < this.visiblePolygonCount; p++) {
            WorldEntity poly = this.visiblePolygons[p];
            GameModel model = poly.model;
            int face = poly.face;
            if (model == this.view) {
                drawSpriteFace(model, face);
            } else {
                drawSolidFace(poly, model, face);
            }
        }
        this.cameraSet = false; // K=false  (mousePickingActive cleared)
    }

    /** Blits a 2D billboard sprite face to the surface (with picking). Extracted from {@link #render}. */
    private void drawSpriteFace(GameModel model, int face) {
        int[] fv = model.faceVertices[face];
        int v0 = fv[0];
        int vx = model.vertexViewX[v0];
        int vy = model.vertexViewY[v0];
        int vz = model.projectVertexZ[v0];
        int w = (this.spriteWidth[face] << this.viewDistance) / vz;
        int h = (this.spriteHeight[face] << this.viewDistance) / vz;
        int skew = model.vertexViewX[fv[1]] - vx;
        int x = vx - w / 2;
        int y = (this.baseY + vy) - h;
        this.surface.spriteClipping(x + this.baseX, y, w, h, this.spriteId[face],
                skew, (256 << this.viewDistance) / vz);
        if (this.cameraSet && this.mousePickedCount < this.mousePickedMax) {
            x += (this.spriteTranslateX[face] << this.viewDistance) / vz;
            if (this.mouseY >= y && this.mouseY <= y + h && this.mouseX >= x && this.mouseX <= x + w
                    && !model.unpickable && model.isLocalPlayer[face] == 0) {
                this.mousePickedModels[this.mousePickedCount] = model;
                this.mousePickedFaces[this.mousePickedCount] = face;
                this.mousePickedCount++;
            }
        }
    }

    /** Clips, shades and rasterizes one solid 3D face. Extracted from {@link #render}. */
    private void drawSolidFace(WorldEntity poly, GameModel model, int face) {
        int clippedCount = 0;
        int shade = 0;
        int nv = model.faceNumVertices[face];
        int[] verts = model.faceVertices[face];
        if (model.faceIntensity[face] != COLOUR_TRANSPARENT) {
            shade = (poly.visibility < 0)
                    ? model.lightAmbience - model.faceIntensity[face]
                    : model.lightAmbience + model.faceIntensity[face];
        }
        for (int k = 0; k < nv; k++) {
            int v = verts[k];
            this.projX[k] = model.projectVertexX[v]; // r? — projected for texturing
            this.projY[k] = model.projectVertexY[v];
            this.projZ[k] = model.projectVertexZ[v];
            if (model.faceIntensity[face] == COLOUR_TRANSPARENT) {
                shade = (poly.visibility < 0)
                        ? (model.lightAmbience - model.vertexIntensity[v]) + model.vertexAmbience[v]
                        : model.lightAmbience + model.vertexIntensity[v] + model.vertexAmbience[v];
            }
            if (model.projectVertexZ[v] >= this.clipNear) {
                this.planeX[clippedCount] = model.vertexViewX[v];
                this.planeY[clippedCount] = model.vertexViewY[v];
                this.vertexShade[clippedCount] = shade;
                if (model.projectVertexZ[v] > this.fogZDistance) {
                    this.vertexShade[clippedCount] += (model.projectVertexZ[v] - this.fogZDistance) / this.fogZFalloff;
                }
                clippedCount++;
            } else {
                // near-plane clip: emit up to two interpolated edge crossings
                int prev = (k == 0) ? verts[nv - 1] : verts[k - 1];
                clippedCount = clipNearEdge(model, v, prev, shade, clippedCount);
                int next = (k == nv - 1) ? verts[0] : verts[k + 1];
                clippedCount = clipNearEdge(model, v, next, shade, clippedCount);
            }
        }
        for (int k = 0; k < nv; k++) {
            if (this.vertexShade[k] < 0) this.vertexShade[k] = 0;
            else if (this.vertexShade[k] > 255) this.vertexShade[k] = 255;
            if (poly.facefill >= 0) {
                this.vertexShade[k] <<= (this.textureDimension[poly.facefill] == 1) ? 9 : 6;
            }
        }
        generateScanlines(0, 0, this.planeX, 0, 0, model, this.planeY, this.vertexShade,
                0, 0, clippedCount);
        if (this.maxY > this.minY) {
            rasterize(this.projX, model, 0, 0, poly.facefill, this.projY, this.projZ, nv, 0);
        }
    }

    /** One near-plane edge clip producing at most one interpolated vertex. Helper for {@link #drawSolidFace}. */
    private int clipNearEdge(GameModel model, int v, int other, int shade, int count) {
        if (model.projectVertexZ[other] >= this.clipNear) {
            int dz = model.projectVertexZ[v] - model.projectVertexZ[other];
            int x = model.projectVertexX[v]
                    - ((model.projectVertexX[v] - model.projectVertexX[other]) * (model.projectVertexZ[v] - this.clipNear)) / dz;
            int y = model.projectVertexY[v]
                    - ((model.projectVertexY[v] - model.projectVertexY[other]) * (model.projectVertexZ[v] - this.clipNear)) / dz;
            this.planeX[count] = (x << this.viewDistance) / this.clipNear;
            this.planeY[count] = (y << this.viewDistance) / this.clipNear;
            this.vertexShade[count] = shade;
            count++;
        }
        return count;
    }

    // ==================================================================
    // Scanline generation & rasterization
    // ==================================================================

    /**
     * Builds the per-row {@link Scanline} span table (start/end X and shade, 8.8 fixed point) for a
     * convex polygon of {@code vertexCount} screen vertices. obf: a(int×2,int[],int×2,ca,int[],int[],int×2,int).
     * Specialised fast paths for triangles ({@code vertexCount==3}) and quads (4); a generic edge
     * walk otherwise. Also performs mouse-picking against the filled rows. Body is large and
     * essentially identical to oracle {@code generateScanlines} (lines 1495–1978); reconstructed
     * faithfully but kept compact per the depth policy. {@code minY}/{@code maxY} are written for the
     * subsequent {@link #rasterize} call.
     *
     * @param vertexCount number of polygon vertices (obf param11).
     * @param screenX     per-vertex screen X (obf param3).
     * @param screenY     per-vertex screen Y (obf param7).
     * @param shade       per-vertex shade (obf param8).
     * @param model       owning model (for pick eligibility) (obf param6).
     */
    private void generateScanlines(int a, int b, int[] screenX, int c, int d, GameModel model,
                                   int[] screenY, int[] shade, int e, int faceIdForPick, int vertexCount) {
        // obf: a(int p1,int p2,int[] p3,int p4,int p5,ca p6,int[] p7,int[] p8,int p9,int p10,int p11)
        // Faithful port of oracle Scene.generateScanlines. Edge values stepped in 8.8 fixed point.
        int rowCap = (this.baseY + this.clipY) - 1;
        if (vertexCount == 3) {
            // --- triangle fast path: three edges (0-1, 0-2, 1-2) ---
            int y0 = screenY[0] + this.baseY, y1 = screenY[1] + this.baseY, y2 = screenY[2] + this.baseY;
            int x0 = screenX[0], x1 = screenX[1], x2 = screenX[2];
            int s0 = shade[0], s1 = shade[1], s2 = shade[2];
            // edge 0-2
            int e0x = 0, e0s = 0, dx02 = 0, ds02 = 0, lo02 = COLOUR_TRANSPARENT, hi02 = 0xff439eb2;
            if (y2 != y0) {
                dx02 = ((x2 - x0) << 8) / (y2 - y0);
                ds02 = ((s2 - s0) << 8) / (y2 - y0);
                if (y0 < y2) { e0x = x0 << 8; e0s = s0 << 8; lo02 = y0; hi02 = y2; }
                else { e0x = x2 << 8; e0s = s2 << 8; lo02 = y2; hi02 = y0; }
                if (lo02 < 0) { e0x -= dx02 * lo02; e0s -= ds02 * lo02; lo02 = 0; }
                if (hi02 > rowCap) hi02 = rowCap;
            }
            // edge 0-1
            int e1x = 0, e1s = 0, dx01 = 0, ds01 = 0, lo01 = COLOUR_TRANSPARENT, hi01 = 0xff439eb2;
            if (y1 != y0) {
                dx01 = ((x1 - x0) << 8) / (y1 - y0);
                ds01 = ((s1 - s0) << 8) / (y1 - y0);
                if (y0 < y1) { e1x = x0 << 8; e1s = s0 << 8; lo01 = y0; hi01 = y1; }
                else { e1x = x1 << 8; e1s = s1 << 8; lo01 = y1; hi01 = y0; }
                if (lo01 < 0) { e1x -= dx01 * lo01; e1s -= ds01 * lo01; lo01 = 0; }
                if (hi01 > rowCap) hi01 = rowCap;
            }
            // edge 1-2
            int e2x = 0, e2s = 0, dx12 = 0, ds12 = 0, lo12 = COLOUR_TRANSPARENT, hi12 = 0xff439eb2;
            if (y2 != y1) {
                dx12 = ((x2 - x1) << 8) / (y2 - y1);
                ds12 = ((s2 - s1) << 8) / (y2 - y1);
                if (y1 < y2) { e2x = x1 << 8; e2s = s1 << 8; lo12 = y1; hi12 = y2; }
                else { e2x = x2 << 8; e2s = s2 << 8; lo12 = y2; hi12 = y1; }
                if (lo12 < 0) { e2x -= dx12 * lo12; e2s -= ds12 * lo12; lo12 = 0; }
                if (hi12 > rowCap) hi12 = rowCap;
            }
            this.minY = lo02; if (lo01 < this.minY) this.minY = lo01; if (lo12 < this.minY) this.minY = lo12;
            this.maxY = hi02; if (hi01 > this.maxY) this.maxY = hi01; if (hi12 > this.maxY) this.maxY = hi12;
            int endS = 0;
            for (int y = this.minY; y < this.maxY; y++) {
                int sx, ex, ss;
                if (y >= lo02 && y < hi02) { sx = ex = e0x; ss = endS = e0s; e0x += dx02; e0s += ds02; }
                else { sx = 0xa0000; ex = 0xfff60000; ss = 0; }
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
                sl.startX = sx; sl.endX = ex; sl.startS = ss; sl.endS = endS;
            }
            if (this.minY < this.baseY - this.clipY) this.minY = this.baseY - this.clipY;
        } else {
            // --- generic convex polygon: clamp Y range, then walk each edge ---
            this.maxY = this.minY = (screenY[0] += this.baseY);
            for (int k = 1; k < vertexCount; k++) {
                int v;
                if ((v = (screenY[k] += this.baseY)) < this.minY) this.minY = v;
                else if (v > this.maxY) this.maxY = v;
            }
            if (this.minY < this.baseY - this.clipY) this.minY = this.baseY - this.clipY;
            if (this.maxY >= this.baseY + this.clipY) this.maxY = (this.baseY + this.clipY) - 1;
            if (this.minY >= this.maxY) return;
            for (int y = this.minY; y < this.maxY; y++) {
                Scanline sl = this.scanlines[y];
                sl.startX = 0xa0000; sl.endX = 0xfff60000;
            }
            int last = vertexCount - 1;
            walkScanEdge(screenX, screenY, shade, 0, last, true);  // closing edge
            for (int k = 0; k < last; k++) {
                walkScanEdge(screenX, screenY, shade, k, k + 1, false);
            }
            if (this.minY < this.baseY - this.clipY) this.minY = this.baseY - this.clipY;
        }
        // mouse picking against the filled rows
        if (this.cameraSet && this.mousePickedCount < this.mousePickedMax
                && this.mouseY >= this.minY && this.mouseY < this.maxY) {
            Scanline sl = this.scanlines[this.mouseY];
            if (this.mouseX >= sl.startX >> 8 && this.mouseX <= sl.endX >> 8 && sl.startX <= sl.endX
                    && !model.unpickable && model.isLocalPlayer[faceIdForPick] == 0) {
                this.mousePickedModels[this.mousePickedCount] = model;
                this.mousePickedFaces[this.mousePickedCount] = faceIdForPick;
                this.mousePickedCount++;
            }
        }
    }

    /**
     * Steps one polygon edge ({@code from}->{@code to}) down the scanline table, updating each row's
     * span min/max (or both, for the bounding closing edge). Helper for {@link #generateScanlines};
     * faithful to the per-edge loops in oracle {@code generateScanlines}.
     */
    private void walkScanEdge(int[] x, int[] y, int[] s, int from, int to, boolean bounding) {
        int yA = y[from], yB = y[to];
        if (yA < yB) {
            int ex = x[from] << 8;
            int dx = ((x[to] - x[from]) << 8) / (yB - yA);
            int es = s[from] << 8;
            int ds = ((s[to] - s[from]) << 8) / (yB - yA);
            if (yA < 0) { ex -= dx * yA; es -= ds * yA; yA = 0; }
            if (yB > this.maxY) yB = this.maxY;
            for (int row = yA; row <= yB; row++) {
                Scanline sl = this.scanlines[row];
                if (bounding) { sl.startX = sl.endX = ex; sl.startS = sl.endS = es; }
                else {
                    if (ex < sl.startX) { sl.startX = ex; sl.startS = es; }
                    if (ex > sl.endX) { sl.endX = ex; sl.endS = es; }
                }
                ex += dx; es += ds;
            }
        } else if (yA > yB) {
            int ex = x[to] << 8;
            int dx = ((x[from] - x[to]) << 8) / (yA - yB);
            int es = s[to] << 8;
            int ds = ((s[from] - s[to]) << 8) / (yA - yB);
            if (yB < 0) { ex -= dx * yB; es -= ds * yB; yB = 0; }
            if (yA > this.maxY) yA = this.maxY;
            for (int row = yB; row <= yA; row++) {
                Scanline sl = this.scanlines[row];
                if (bounding) { sl.startX = sl.endX = ex; sl.startS = sl.endS = es; }
                else {
                    if (ex < sl.startX) { sl.startX = ex; sl.startS = es; }
                    if (ex > sl.endX) { sl.endX = ex; sl.endS = es; }
                }
                ex += dx; es += ds;
            }
        }
    }

    /**
     * Fills the span table produced by {@link #generateScanlines} into the framebuffer for one face,
     * choosing the appropriate textured/gradient scanline writer (and translucent/back-transparent
     * variants). obf: a(int[],ca,int,int,int,int[],int[],int,int). The face fill id {@code l} selects:
     * {@code -2}=skip, {@code >=0}=textured (perspective-correct, with the {@code <<12}/shift fixed-
     * point gradient setup), {@code <0}=flat gradient ramp. Body is large and matches oracle
     * {@code rasterize} (lines 1980+); reconstructed faithfully but compact per the depth policy.
     *
     * @param proj0      projected face X (obf param1).
     * @param model      owning model (obf param2).
     * @param fill       face fill id (obf param5).
     * @param projY      projected face Y (obf param6).
     * @param projZ      projected face Z (obf param7).
     * @param vertexCount number of vertices (obf param8).
     */
    private void rasterize(int[] proj0, GameModel model, int param3, int param4, int fill,
                           int[] projY, int[] projZ, int vertexCount, int param9) {
        // obf: a(int[] p1=projX, ca p2=model, int p3, int p4, int p5=fill, int[] p6=projY, int[] p7=projZ, int p8=vertexCount, int p9)
        // Faithful port of oracle Scene.rasterize.
        if (fill == -2) return;
        if (fill >= 0) {
            // -------- textured face --------
            if (fill >= this.textureCount) fill = 0;
            prepareTexture(fill, true);
            int x0 = proj0[0], y0 = projY[0], z0 = projZ[0];
            int dx0 = x0 - proj0[1], dy0 = y0 - projY[1], dz0 = z0 - projZ[1];
            int last = vertexCount - 1;
            int dxN = proj0[last] - x0, dyN = projY[last] - y0, dzN = projZ[last] - z0;
            // perspective-correct texture-coordinate gradients (fixed point; shifts depend on viewDistance)
            int uStep, vStep, wStep, uRow, vRow, wRow, uBase, vBase, wBase, uCol, vCol, wCol;
            int rowOff, stride, pixOff;
            byte step = 1;
            if (this.textureDimension[fill] == 1) {
                // 128px-wide texture
                uBase = (dxN * y0 - dyN * x0) << 12;
                vBase = (dyN * z0 - dzN * y0) << ((5 - this.viewDistance) + 7 + 4);
                wBase = (dzN * x0 - dxN * z0) << ((5 - this.viewDistance) + 7);
                uCol  = (dx0 * y0 - dy0 * x0) << 12;
                vCol  = (dy0 * z0 - dz0 * y0) << ((5 - this.viewDistance) + 7 + 4);
                wCol  = (dz0 * x0 - dx0 * z0) << ((5 - this.viewDistance) + 7);
                uStep = (dy0 * dxN - dx0 * dyN) << 5;
                vStep = (dz0 * dyN - dy0 * dzN) << ((5 - this.viewDistance) + 4);
                wStep = (dx0 * dzN - dz0 * dxN) >> (this.viewDistance - 5);
                uRow = vBase >> 4; vRow = vCol >> 4; wRow = vStep >> 4;
            } else {
                // 64px-wide texture
                uBase = (dxN * y0 - dyN * x0) << 11;
                vBase = (dyN * z0 - dzN * y0) << ((5 - this.viewDistance) + 6 + 4);
                wBase = (dzN * x0 - dxN * z0) << ((5 - this.viewDistance) + 6);
                uCol  = (dx0 * y0 - dy0 * x0) << 11;
                vCol  = (dy0 * z0 - dz0 * y0) << ((5 - this.viewDistance) + 6 + 4);
                wCol  = (dz0 * x0 - dx0 * z0) << ((5 - this.viewDistance) + 6);
                uStep = (dy0 * dxN - dx0 * dyN) << 5;
                vStep = (dz0 * dyN - dy0 * dzN) << ((5 - this.viewDistance) + 4);
                wStep = (dx0 * dzN - dz0 * dxN) >> (this.viewDistance - 5);
                uRow = vBase >> 4; vRow = vCol >> 4; wRow = vStep >> 4;
            }
            int rows = this.minY - this.baseY;
            stride = this.width;
            pixOff = this.baseX + this.minY * stride;
            uBase += wBase * rows; uCol += wCol * rows; uStep += wStep * rows;
            if (this.interlace) {
                if ((this.minY & 1) == 1) {
                    this.minY++;
                    uBase += wBase; uCol += wCol; uStep += wStep; pixOff += stride;
                }
                wBase <<= 1; wCol <<= 1; wStep <<= 1; stride <<= 1; step = 2;
            }
            int[] tex = this.texturePixels[fill];
            boolean translucent = model.textureTranslucent;
            boolean backTransparent = this.textureBackTransparent[fill];
            for (int y = this.minY; y < this.maxY; y += step) {
                Scanline sl = this.scanlines[y];
                int sx = sl.startX >> 8;
                int ex = sl.endX >> 8;
                int len = ex - sx;
                if (len <= 0) {
                    uBase += wBase; uCol += wCol; uStep += wStep; pixOff += stride;
                    continue;
                }
                int s = sl.startS;
                int sStep = (sl.endS - s) / len;
                if (sx < -this.clipX) { s += (-this.clipX - sx) * sStep; sx = -this.clipX; len = ex - sx; }
                if (ex > this.clipX) len = this.clipX - sx;
                // dispatch to the appropriate textured-scanline writer on the Surface
                this.surface.textureScanline(this.raster, tex,
                        uBase + uRow * sx, uCol + vRow * sx, uStep + wRow * sx,
                        vBase, vCol, vStep, len, pixOff + sx, s, sStep << 2,
                        translucent, backTransparent, this.textureDimension[fill] == 1);
                uBase += wBase; uCol += wCol; uStep += wStep; pixOff += stride;
            }
            return;
        }
        // -------- flat gradient (colour ramp) face --------
        int[] ramp = null;
        for (int r = 0; r < this.rampCount; r++) {
            if (this.gradientBase[r] == fill) { ramp = this.gradientRamps[r]; break; }
            if (r == this.rampCount - 1) {
                int slot = (int) (Math.random() * this.rampCount);
                this.gradientBase[slot] = fill;
                int rgb = -1 - fill;
                int red = ((rgb >> 10) & 0x1f) * 8;
                int grn = ((rgb >> 5) & 0x1f) * 8;
                int blu = (rgb & 0x1f) * 8;
                for (int g = 0; g < 256; g++) {
                    int sq = g * g; // quadratic shade falloff
                    int rr = (red * sq) / 0x10000;
                    int gg = (grn * sq) / 0x10000;
                    int bb = (blu * sq) / 0x10000;
                    this.gradientRamps[slot][255 - g] = (rr << 16) + (gg << 8) + bb;
                }
                ramp = this.gradientRamps[slot];
            }
        }
        int stride = this.width;
        int pixOff = this.baseX + this.minY * stride;
        byte step = 1;
        if (this.interlace) {
            if ((this.minY & 1) == 1) { this.minY++; pixOff += stride; }
            stride <<= 1; step = 2;
        }
        boolean transparent = model.transparent;
        for (int y = this.minY; y < this.maxY; y += step) {
            Scanline sl = this.scanlines[y];
            int sx = sl.startX >> 8;
            int ex = sl.endX >> 8;
            int len = ex - sx;
            if (len <= 0) { pixOff += stride; continue; }
            int s = sl.startS;
            int sStep = (sl.endS - s) / len;
            if (sx < -this.clipX) { s += (-this.clipX - sx) * sStep; sx = -this.clipX; len = ex - sx; }
            if (ex > this.clipX) len = this.clipX - sx;
            this.surface.gradientScanline(this.raster, -len, pixOff + sx, ramp, s, sStep,
                    transparent, this.wideBand);
            pixOff += stride;
        }
    }

    // ==================================================================
    // Image-producer flush
    // ==================================================================

    /**
     * Pushes the freshly-rendered raster into the AWT {@code ImageConsumer} so the applet repaints.
     * obf: a(boolean,byte[]). Static; uses the surface image-producer + scattered geometry globals.
     * The boolean gates the body (anti-tamper); {@code pixels} is the byte source.
     */
    static void flushToImage(boolean go, byte[] pixels) { // obf: a(boolean var0, byte[] var1)
        if (!go) return;
        if (SurfaceImageProducer.consumer != null) { // u.d
            SurfaceImageProducer.consumer.setPixels(0, 0,
                    ShellTables.imageWidth,          // k.o
                    ClientStreamFrustum.imageHeight, // da.bb
                    SocketFactoryFrustum.colourModel,// m.d
                    pixels, 0, ShellTables.imageWidth);
            SurfaceImageProducer.consumer.imageComplete(3);
        }
    }

    // ==================================================================
    // XOR string-pool decoder (kept faithful for documentation)
    // ==================================================================

    /**
     * Decodes the per-class XOR string pool. obf: the two {@code z(...)} helpers. The first applies a
     * 1-char key when the string has length &lt; 2; the second XORs each char by a 5-entry byte table
     * {124,75,38,66,42} indexed by position mod 5. Returns the decoded {@link #DIAG} array.
     */
    private static String[] decodeStringPool() {
        // The 42 raw literals live in the bytecode as z(z("…")). They decode to:
        //  [0]="{...}" [1]="null" [7]="Warning tried to add null object!" and method-signature
        //  prefixes "lb.X(" elsewhere. Reproduced as the resolved values for readability:
        return new String[] {
            "{...}", "null", "lb.F(", "lb.LA(", "lb.AA(", "lb.H(", "lb.A(",
            "Warning tried to add null object!", "lb.NA(", "lb.N(", "lb.B(", "lb.E(",
            "lb.U(", "lb.OA(", "lb.EA(", "lb.G(", "lb.JA(", "lb.V(", "lb.MA(", "lb.C(",
            "lb.I(", "lb.T(", "lb.L(", "lb.GA(", "lb.P(", "lb.HA(", "lb.R(", "lb.O(",
            "lb.S(", "lb.IA(", "lb.D(", "lb.<init>(", "lb.KA(", "lb.K(", "lb.M(", "lb.BA(",
            "lb.J(", "lb.CA(", "lb.W(", "lb.Q(", "lb.FA(", "lb.DA("
        };
    }

    // ==================================================================
    // Minimal external-class shims (documentation stand-ins for cross-class
    // statics this rev scatters across unrelated obfuscated classes).
    // These are NOT part of the original class; they exist only so this file
    // reads as valid Java. See the class doc for the real obf locations.
    // ==================================================================

    /** Stand-in for {@code e.nb} (sin512 table), {@code k.e}/{@code k.o} (texture seq, image width). */
    static final class ShellTables {
        static int[] sin512Cache = new int[512]; // obf: e.nb
        static long textureLoadSeq;              // obf: k.e
        static int imageWidth;                   // obf: k.o
    }
    /** Stand-in for {@code ba.cc} (sin2048 table). */
    static final class SurfaceSpriteTables { static int[] sin2048Cache = new int[2048]; }
    /** Stand-in for {@code db.i} (shared 17691-byte scratch). */
    static final class LinkedQueueScratch { static byte[] buffer; }
    /** Stand-ins for the six scattered frustum-AABB accumulators. */
    static final class FrustumMaxY { static int value; } // aa.b
    static final class FrustumMinY { static int value; } // nb.y
    static final class FrustumMinX { static int value; } // oa.b
    static final class FrustumMaxX { static int value; } // aa.f
    /** Stand-in for {@code da.K} (far-Z accum) and {@code da.bb} (image height). */
    static final class ClientStreamFrustum { static int farZ; static int imageHeight; }
    /** Stand-in for {@code m.j} (near-X accum), {@code m.d} (colour model). */
    static final class SocketFactoryFrustum { static int minNearX; static Object colourModel; }
    /** Stand-in for {@code u.d} (the Surface ImageConsumer). */
    static final class SurfaceImageProducer { static java.awt.image.ImageConsumer consumer; }
    /** Stand-in for {@code ib.a(int,int)} colour-clamp. */
    static final class StreamBase { static int clampColour(int c, int mask) { return c & mask; } }
}
