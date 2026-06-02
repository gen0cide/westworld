package client.scene;

import java.io.DataInputStream;
import java.io.IOException;
import java.io.InputStream;

// NOTE: imports below name the canonical roles of the obfuscated helper classes this file references.
// Many won't resolve standalone (this is a readability port, not a recompile target).
import client.data.CacheFile;     // d  - getSignedShort / getUnsignedShort helpers in this build
import client.data.DataStore;     // nb - resource/file opener (+ a frustum bound, nb.y)
import client.data.NameHash;      // oa - holds a frustum bound (oa.b)
import client.data.NameTable;     // ub - texture/material name table (ub.c[])
import client.data.BZip;          // aa - holds frustum bounds (aa.f, aa.b)
import client.net.ClientStream;   // da - holds a frustum bound (da.K)
import client.net.SocketFactory;  // m  - holds a frustum bound (m.j)
import client.net.StreamBase;     // ib - unsigned-byte helper
import client.scene.ImageLoader;  // pa - shared sine9 (pa.a) / sine11 (pa.j) / base64 (pa.d) tables
import client.scene.SpriteScaler; // ia - texture-name table length counter (ia.b)
import client.util.ErrorHandler;  // i  - wraps Throwable + a context string (obfuscation try/catch)
import client.world.WorldEntity;  // w  - getSignedShort helper used by the byte[] constructor

/**
 * GameModel — a software-rendered 3D model: a bag of vertices and (triangular/quad) faces plus all
 * the per-model transform and lighting state needed to pose it in the world and feed it to
 * {@link Scene}/{@link Surface} for rasterisation.
 *
 * <p>This is the rev ~233-235 (Microsoft J++) obfuscated incarnation of mudclient204's
 * {@code GameModel}. Versus rev 204 it carries extra per-face arrays
 * ({@link #isLocalPlayer}/{@link #faceTag}), a {@code transparent}/{@code cleared} flag, and the
 * trig / base-64 lookup tables have moved into {@link ImageLoader}
 * ({@code pa.a}={@link ImageLoader#SIN_512}, {@code pa.j}={@link ImageLoader#SIN_2048},
 * {@code pa.d}={@link ImageLoader#BASE64_DECODE} alphabet).
 *
 * <h2>Coordinate &amp; fixed-point conventions</h2>
 * <ul>
 *   <li>Model angles are 0..255 (yaw/pitch/roll) indexing the 512-entry sine9 table; camera angles
 *       are 0..1023 indexing the 2048-entry sine11 table. Each table holds {@code sin*32768} in the
 *       low half and {@code cos*32768} in the high half.</li>
 *   <li>Rotations undo the 2^15 sine scale with {@code >> 15}.</li>
 *   <li>Scale/shear factors are 8.8 fixed point (256 == 1.0), undone with {@code >> 8}.</li>
 *   <li>Surface normals are stored 16.16 fixed point ({@code component * 65536 / |n|}).</li>
 * </ul>
 *
 * <h2>Lazy transform pipeline</h2>
 * Mutators ({@link #rotate}, {@link #orient}, {@link #translate}, {@link #place}, …) only update the
 * pending transform parameters and raise {@link #transformState}. {@link #apply} re-bakes the
 * transformed vertices + bounds + lighting on demand; {@link #project} additionally projects to
 * screen space; {@link #commit} folds the current transform back into the base vertices.
 */
public final class GameModel {

   // ---------------------------------------------------------------------------------------------
   // Obfuscation profiling counters (one bumped at the top of each original method). Dead weight,
   // retained only so the static layout matches the bytecode; never read for game logic.
   // ---------------------------------------------------------------------------------------------
   public static int profPlace;              // q
   public static int profVertexAt;           // W
   public static int profCopyTransparent;    // wb
   public static int profProject;            // ac
   public static int profSplit;              // Rb
   public static int profSetLightDir;        // p
   public static int profReduce;             // d
   public static int profSetLightGouraud;    // m
   public static int profApplyShear;         // S
   public static int profCopyLighting;       // mb
   public static int profApplyRotation;      // ab
   public static int profSetLightAmbient;    // R
   public static int profReduceB;            // ub
   public static int profComputeBounds;      // h
   public static int profReadBase64;         // s
   public static int profCommitHelper;       // D
   public static int profProjectionPrepare;  // N
   public static int profRotate;             // A
   public static int profTextureId;          // Xb
   public static int[] sharedScratch;        // B (static scratch int[], nulled by an anti-tamper guard)
   public static int profSetLight5;          // Ub
   public static int profSetVertexAmbience;  // Wb
   public static int profCreateFace;         // u
   public static int profMerge;              // kb
   public static int profCopyFlags;          // I
   public static int profAllocate;           // nb
   public static int profFindOrAdd;          // ib
   public static int profApply;              // Qb
   public static int profTranslate;          // L
   public static int profApplyScale;         // O
   public static int profCopyPosition;       // y
   public static int profRelight;            // Z
   public static int profCreateVertex;       // J
   public static int profLight;              // l
   public static int profDetermineKind;      // vb

   /** tb — 50-slot static cache of decoded model byte[] payloads (loaded by name, rev-235 only). */
   public static byte[][] modelCache = new byte[50][];

   // --- Lighting direction & strength ---
   private int lightDirectionX;          // g   (default 180)
   private int lightDirectionY;          // Bb  (default 155)
   private int lightDirectionZ;          // Fb  (default 95)
   private int lightDirectionMagnitude;  // Ib  (default 256)
   private int lightDiffuse;             // Mb  (default 512)
   public int lightAmbience;            // Jb  (default 32)

   // --- Base (untransformed) vertex geometry ---
   public int numVertices;        // Db
   public int[] vertexX;          // a
   public int[] vertexY;          // ob
   public int[] vertexZ;          // bc
   public int[] vertexIntensity;  // gb  (per-vertex baked light, for Gouraud shading)
   public byte[] vertexAmbience;  // Ab
   private int maxVerts;   // K   (capacity)

   // --- Transformed vertex geometry (aliases base arrays when autocommit) ---
   private int[] vertexTransformedX;  // Gb
   private int[] vertexTransformedY;  // ic
   private int[] vertexTransformedZ;  // jc

   // --- Projected (camera/screen-space) vertex geometry (only when !projected) ---
   public int[] projectVertexX;   // cc
   public int[] projectVertexY;   // H
   public int[] projectVertexZ;   // bb
   public int[] vertexViewX;      // pb  (perspective-divided screen x)
   public int[] vertexViewY;      // Ob  (perspective-divided screen y)

   // --- Face topology & per-face material/lighting ---
   public int numFaces;            // t
   public int[] faceNumVertices;   // lb
   public int[][] faceVertices;    // o
   public int[] faceFillFront;     // V   (32767 -> magic)
   public int[] faceFillBack;      // qb
   public int[] faceIntensity;     // Hb  (== magic means recompute per-vertex; else flat)
   public int[] normalScale;       // M   (-1 marks a recomputed normal)
   public int[] normalMagnitude;   // k
   public byte[] isLocalPlayer;    // zb  (per-face; only when !unpickable)
   public int[] faceTag;           // E   (per-face pick tag; only when !unpickable)
   private int maxFaces;    // z   (capacity)

   // Face surface normals, 16.16 fixed point (only when !unlit || !isolated).
   private int[] faceNormalX;  // fb
   private int[] faceNormalY;  // Cb
   private int[] faceNormalZ;  // Pb

   // Per-face object-space bounds (only when !isolated).
   private int[] faceBoundLeft;    // w   (min X)
   private int[] faceBoundRight;   // n   (max X)
   private int[] faceBoundBottom;  // Q   (min Y)
   private int[] faceBoundTop;     // Lb  (max Y)
   private int[] faceBoundNear;    // Zb  (min Z)
   private int[] faceBoundFar;     // Eb  (max Z)

   /** fc — per-face transform-group membership: which source sub-model(s) a face came from. */
   private int[][] faceTransGroups;  // fc

   // --- Pending transform parameters (consumed lazily by apply()) ---
   private int baseX;            // r
   private int baseY;            // Sb
   private int baseZ;            // xb
   private int orientationYaw;   // C
   private int orientationPitch; // F
   private int orientationRoll;  // X
   private int scaleFx;          // yb
   private int scaleFy;          // i
   private int scaleFz;          // T
   private int shearXy;          // U
   private int shearXz;          // f
   private int shearYx;          // Tb
   private int shearYz;          // G
   private int shearZx;          // Y
   private int shearZy;          // eb
   private int transformKind;    // jb  (0 none,1 +translate,2 +rotate,3 +scale,4 +shear)
   public int transformState;   // Yb  (0 clean,1 full re-apply,2 cheap re-apply)

   // --- Whole-model state, bounds & flags ---
   private int diameter;     // sb  (largest per-face extent; "magic" until first apply)
   private int magic;        // Vb  (World.colourTransparent sentinel == 12345678)
   public int key;                  // rb  (caller id; default -1)
   public int depth;                // hc  (render-sort depth)
   public boolean visible;          // dc  (set by project(): inside the view frustum?)
   public boolean transparent;      // cb
   public boolean cleared;          // Kb  (set by a "soft" clear())

   // Object-space model bounds (computed in computeBounds; used by project()'s frustum test).
   private int boundX1;  // x   (min X)
   private int boundX2;  // j   (max X)
   private int boundY1;  // P   (min Y)
   private int boundY2;  // e   (max Y)
   private int boundZ1;  // ec  (min Z)
   private int boundZ2;  // gc  (max Z)

   // --- Construction / topology flags (fixed at allocate()) ---
   private boolean autocommit;  // v   (transformed arrays alias base arrays)
   private boolean isolated;    // c   (skip per-face bounds & some normals)
   private boolean unlit;       // Nb  (skip lighting passes)
   private boolean projected;   // b   (skip projected-vertex scratch; source already projected)
   public boolean unpickable;  // db  (skip faceTag/isLocalPlayer per-face arrays)

   private int dataPtr;         // hb  (read cursor for the byte/base-64 decoders)

   /**
    * kc — the obfuscator's pooled, char-XOR-scrambled strings (mostly "GameModel.method("
    * error prefixes, plus "true"/"false"/"null"/"Invisible"). Decoded once below via
    * {@link #unscramble(String)} / {@link #unscramble(char[])}.
    */
   private static final String[] SCRAMBLED = new String[]{
      unscramble(unscramble("x$p\r")),
      unscramble(unscramble("`Y$X")),
      unscramble(unscramble("x$m\r")),
      unscramble(unscramble("ufQ")),
      unscramble(unscramble("x$~d3")),
      unscramble(unscramble("u")),
      unscramble(unscramble("x$Lu~\r")),
      unscramble(unscramble("x$zd3")),
      unscramble(unscramble("x$ud3")),
      unscramble(unscramble("x${d3")),
      unscramble(unscramble("x$td3")),
      unscramble(unscramble("x$l\r")),
      unscramble(unscramble("x$h\r")),
      unscramble(unscramble("x$t\r")),
      unscramble(unscramble("x$z\r")),
      unscramble(unscramble("x$yd3")),
      unscramble(unscramble("x$x\r")),
      unscramble(unscramble("x$j\r")),
      unscramble(unscramble("x${\r")),
      unscramble(unscramble("x$~\r")),
      unscramble(unscramble("x$wd3")),
      unscramble(unscramble("x$q\r")),
      unscramble(unscramble("x$d3")),
      unscramble(unscramble("x$w\r")),
      unscramble(unscramble("x$\r")),
      unscramble(unscramble("x$k\r")),
      unscramble(unscramble("x$|d3")),
      unscramble(unscramble("x$o\r")),
      unscramble(unscramble("x$s\r")),
      unscramble(unscramble("x$i\r")),
      unscramble(unscramble("x$u\r")),
      unscramble(unscramble("x$r\r")),
      unscramble(unscramble("x$n\r")),
      unscramble(unscramble("x$y\r")),
      unscramble(unscramble("x$|\r")),
      unscramble(unscramble("x$vd3")),
      unscramble(unscramble("x$v\r")),
      unscramble(unscramble("x$xd3"))
   };

   static {
      modelCache = new byte[50][];
   }

   // =============================================================================================
   // Constructors
   // =============================================================================================

   /** All field defaults shared by every constructor (matches mudclient204's inline field inits). */
   private void initDefaults() {
      this.lightDirectionY = 155;
      this.isolated = false;
      this.transformState = 1;
      this.unpickable = false;
      this.lightDirectionMagnitude = 256;
      this.autocommit = false;
      this.transparent = false;
      this.unlit = false;
      this.diameter = 12345678;
      this.lightDirectionX = 180;
      this.visible = true;
      this.magic = 12345678;
      this.cleared = false;
      this.key = -1;
      this.lightDiffuse = 512;
      this.lightDirectionZ = 95;
      this.projected = false;
      this.lightAmbience = 32;
      this.depth = 0;
   }

   /** Empty model with room for {@code numVertices} verts / {@code numFaces} faces; identity transform-group per face. */
   public GameModel(int numVertices, int numFaces) {
      this.initDefaults();
      this.allocate(numFaces, numVertices, 69);
      this.faceTransGroups = new int[numFaces][1];
      for (int f = 0; f < numFaces; f++) {
         this.faceTransGroups[f][0] = f;
      }
   }

   /** Capacity ctor with explicit topology flags (used by {@link #split} for the per-tile pieces). */
   public GameModel(int numVertices, int numFaces, boolean autocommit, boolean isolated, boolean unlit, boolean unpickable, boolean projected) {
      this.initDefaults();
      this.isolated = isolated;
      this.projected = projected;
      this.unpickable = unpickable;
      this.autocommit = autocommit;
      this.unlit = unlit;
      this.allocate(numFaces, numVertices, 69);
   }

   /**
    * Decode a model from a raw little-endian byte payload (cache "models" format):
    * [u16 numVerts][u16 numFaces][X*][Y*][Z*][faceVertCount byte*][front u16*][back u16*]
    * [flag byte*][face vertex-index lists].
    */
   public GameModel(byte[] data, int offset, boolean unused) {
      this.initDefaults();
      // DRIFT FIX: CacheFile.getUnsignedShort's deob signature is (byte[] buffer, int offset) — the
      // obf a(int offset, byte dummy, byte[] buffer) was reordered and the dead byte param dropped.
      int nV = CacheFile.getUnsignedShort(data, offset);
      offset += 2;
      int nF = CacheFile.getUnsignedShort(data, offset);
      offset += 2;
      this.allocate(nF, nV, 115);
      this.faceTransGroups = new int[nF][1];

      for (int v = 0; v < nV; v++, offset += 2) {
         this.vertexX[v] = WorldEntity.readSignedShort(data, -1, offset);
      }
      for (int v = 0; v < nV; v++, offset += 2) {
         this.vertexY[v] = WorldEntity.readSignedShort(data, -1, offset);
      }
      for (int v = 0; v < nV; v++, offset += 2) {
         this.vertexZ[v] = WorldEntity.readSignedShort(data, -1, offset);
      }
      this.numVertices = nV;

      for (int f = 0; f < nF; f++) {
         this.faceNumVertices[f] = StreamBase.bitwiseAnd(255, data[offset++]);
      }
      for (int f = 0; f < nF; f++, offset += 2) {
         this.faceFillFront[f] = WorldEntity.readSignedShort(data, -1, offset);
         if (this.faceFillFront[f] == 32767) {
            this.faceFillFront[f] = this.magic;
         }
      }
      for (int f = 0; f < nF; f++, offset += 2) {
         this.faceFillBack[f] = WorldEntity.readSignedShort(data, -1, offset);
         if (this.faceFillBack[f] == 32767) {
            this.faceFillBack[f] = this.magic;
         }
      }
      for (int f = 0; f < nF; f++) {
         int flag = 255 & data[offset++];
         // flag==0 -> flat shading (0); non-zero -> Gouraud (magic sentinel)
         this.faceIntensity[f] = (flag == 0) ? 0 : this.magic;
      }
      for (int f = 0; f < nF; f++) {
         this.faceVertices[f] = new int[this.faceNumVertices[f]];
         for (int i = 0; i < this.faceNumVertices[f]; i++) {
            if (nV < 256) {
               this.faceVertices[f][i] = StreamBase.bitwiseAnd(255, data[offset++]);
            } else {
               // DRIFT FIX: CacheFile.getUnsignedShort(byte[] buffer, int offset) — see above.
               this.faceVertices[f][i] = CacheFile.getUnsignedShort(data, offset);
               offset += 2;
            }
         }
      }
      this.numFaces = nF;
      this.transformState = 1;
   }

   /** Load a model from a named cache file holding a base-64-style text encoding. */
   public GameModel(String name) {
      this.initDefaults();
      byte[] data;
      try {
         InputStream in = DataStore.openStream(true, name);
         DataInputStream din = new DataInputStream(in);
         // First 3 bytes (one base-64 group) encode the total payload length.
         data = new byte[3];
         this.dataPtr = 0;
         for (int read = 0; read < 3; read += din.read(data, read, 3 - read)) {
         }
         int size = this.readBase64((byte)76, data);
         this.dataPtr = 0;
         data = new byte[size];
         for (int read = 0; read < size; read += din.read(data, read, size - read)) {
         }
         din.close();
      } catch (IOException ioe) {
         this.numFaces = 0;
         this.numVertices = 0;
         return;
      }

      int nV = this.readBase64((byte)76, data);
      int nF = this.readBase64((byte)76, data);
      this.allocate(nF, nV, 97);
      this.faceTransGroups = new int[nF][];

      for (int v = 0; v < nV; v++) {
         int x = this.readBase64((byte)76, data);
         int y = this.readBase64((byte)76, data);
         int z = this.readBase64((byte)76, data);
         this.vertexAt(x, z, y, 52);   // dedup-add (obfuscated e(x,z,y,52))
      }

      for (int f = 0; f < nF; f++) {
         int n = this.readBase64((byte)76, data);
         int front = this.readBase64((byte)76, data);
         int back = this.readBase64((byte)76, data);
         int groupCount = this.readBase64((byte)76, data);
         this.lightDiffuse = this.readBase64((byte)76, data);
         this.lightAmbience = this.readBase64((byte)76, data);
         int gouraud = this.readBase64((byte)76, data);
         int[] vs = new int[n];
         for (int i = 0; i < n; i++) {
            vs[i] = this.readBase64((byte)76, data);
         }
         int[] group = new int[groupCount];
         for (int i = 0; i < groupCount; i++) {
            group[i] = this.readBase64((byte)76, data);
         }
         int dstF = this.createFace(n, vs, front, back, false);
         this.faceTransGroups[f] = group;
         this.faceIntensity[dstF] = (gouraud == 0) ? 0 : this.magic;
      }
      this.transformState = 1;
   }

   /** Merge {@code count} sub-models into one, with explicit topology flags. */
   private GameModel(GameModel[] pieces, int count, boolean autocommit, boolean isolated, boolean unlit, boolean unpickable) {
      this.initDefaults();
      this.unlit = unlit;
      this.isolated = isolated;
      this.unpickable = unpickable;
      this.autocommit = autocommit;
      this.merge(0, pieces, false, count);
   }

   /** Merge {@code count} sub-models into one, carrying through transform-group membership. */
   private GameModel(GameModel[] pieces, int count) {
      this.initDefaults();
      this.merge(0, pieces, true, count);
   }

   // =============================================================================================
   // Allocation
   // =============================================================================================

   /**
    * Allocate all per-vertex / per-face arrays and reset the transform/lighting scalars.
    * Mirrors mudclient204 {@code allocate(numV,numF)} but with the parameter order swapped
    * ({@code numF} first). {@code magicGuard} is an anti-tamper param: when {@code <= 68} the
    * original nulls {@link #faceNumVertices} (a tamper trap), otherwise it is unused.
    */
   private void allocate(int numF, int numV, int magicGuard) {
      // Per-face arrays (sized numF).
      if (!this.unpickable) {
         this.isLocalPlayer = new byte[numF];
         this.faceTag = new int[numF];
      }
      this.normalMagnitude = new int[numF];
      // Per-vertex arrays (sized numV).
      this.vertexY = new int[numV];
      this.vertexAmbience = new byte[numV];
      this.vertexX = new int[numV];
      this.normalScale = new int[numF];
      this.vertexZ = new int[numV];
      this.faceFillBack = new int[numF];
      this.vertexIntensity = new int[numV];
      this.faceFillFront = new int[numF];
      this.faceVertices = new int[numF][];
      this.faceNumVertices = new int[numF];

      // Projection scratch (per-vertex), only when not already projected.
      if (!this.projected) {
         this.vertexViewY = new int[numV];
         this.projectVertexY = new int[numV];
         this.projectVertexX = new int[numV];
         this.projectVertexZ = new int[numV];
         this.vertexViewX = new int[numV];
      }

      this.faceIntensity = new int[numF];
      this.shearZx = 256;
      this.orientationRoll = 0;
      this.shearXy = 256;
      this.shearYx = 256;
      this.scaleFz = 256;

      // Per-face bounds, only when not isolated.
      if (!this.isolated) {
         this.faceBoundLeft = new int[numF];
         this.faceBoundTop = new int[numF];
         this.faceBoundFar = new int[numF];
         this.faceBoundBottom = new int[numF];
         this.faceBoundNear = new int[numF];
         this.faceBoundRight = new int[numF];
      }
      // Per-face normals, only when lit or non-isolated.
      if (!this.unlit || !this.isolated) {
         this.faceNormalY = new int[numF];
         this.faceNormalZ = new int[numF];
         this.faceNormalX = new int[numF];
      }

      this.shearYz = 256;
      this.numFaces = 0;
      this.orientationPitch = 0;
      this.baseY = 0;
      this.scaleFx = 256;

      // Transformed vertices: alias the base arrays in autocommit mode, else fresh arrays.
      if (!this.autocommit) {
         this.vertexTransformedX = new int[numV];
         this.vertexTransformedZ = new int[numV];
         this.vertexTransformedY = new int[numV];
      } else {
         this.vertexTransformedX = this.vertexX;
         this.vertexTransformedZ = this.vertexZ;
         this.vertexTransformedY = this.vertexY;
      }

      this.baseX = 0;
      this.maxVerts = numV;
      this.shearZy = 256;
      this.shearXz = 256;
      this.baseZ = 0;
      this.scaleFy = 256;
      this.orientationYaw = 0;
      this.numVertices = 0;
      if (magicGuard <= 68) {
         this.faceNumVertices = null;   // anti-tamper trap (never hit at runtime)
      }
      this.maxFaces = numF;
      this.transformKind = 0;
   }

   /**
    * Allocate the camera/screen projection scratch arrays after a {@link #split} (pieces start as
    * autocommit/projected). (Obfuscated {@code c(byte)}.) {@code unusedGuard} is dead.
    */
   private void projectionPrepare(byte unusedGuard) {
      this.projectVertexZ = new int[this.numVertices];
      this.vertexViewY = new int[this.numVertices];
      this.projectVertexX = new int[this.numVertices];
      this.projectVertexY = new int[this.numVertices];
      this.vertexViewX = new int[this.numVertices];
   }

   // =============================================================================================
   // Topology mutation
   // =============================================================================================

   /** Reset to empty. {@code soft != 1} additionally marks the model "cleared". */
   public void clear(int soft) {
      this.numFaces = 0;
      if (soft != 1) {
         this.cleared = true;
      }
      this.numVertices = 0;
   }

   /** Shrink the live face/vertex counts by the given deltas (never below zero). */
   public void reduce(int vertexDelta, int unusedGuard, int faceDelta) {
      this.numFaces -= faceDelta;
      // (original anti-tamper static call when unusedGuard > -110, harmless)
      if (~this.numFaces > -1) {        // numFaces < 0
         this.numFaces = 0;
      }
      this.numVertices -= vertexDelta;
      if (this.numVertices < 0) {
         this.numVertices = 0;
      }
   }

   /**
    * Find or append a base vertex at (x,y,z); returns its index, deduping against existing
    * vertices, or -1 if the model is full. {@code unusedGuard} is a dead anti-tamper param.
    * (Obfuscated {@code e(int,int,int,int)}.)
    */
   public int vertexAt(int x, int z, int y, int unusedGuard) {
      for (int v = 0; v < this.numVertices; v++) {
         if (this.vertexX[v] == x && this.vertexY[v] == y && this.vertexZ[v] == z) {
            return v;
         }
      }
      if (this.numVertices < this.maxVerts) {
         this.vertexX[this.numVertices] = x;
         this.vertexY[this.numVertices] = y;
         this.vertexZ[this.numVertices] = z;
         return this.numVertices++;
      }
      return -1;
   }

   /**
    * Append a base vertex at (x,y,z) without deduping; returns its index or -1 if full.
    * {@code unusedFlag} is dead anti-tamper. (Obfuscated {@code b(boolean,int,int,int)}.)
    */
   public int createVertex(boolean unusedFlag, int z, int x, int y) {
      if (~this.maxVerts >= ~this.numVertices) {   // numVertices >= maxVerts
         return -1;
      }
      this.vertexX[this.numVertices] = x;
      this.vertexY[this.numVertices] = y;
      this.vertexZ[this.numVertices] = z;
      // (original anti-tamper static call when unusedFlag, harmless)
      return this.numVertices++;
   }

   /** Append a face from vertex indices {@code vs} with front/back fills; returns index or -1 if full. */
   public int createFace(int n, int[] vs, int front, int back, boolean unusedFlag) {
      // (original anti-tamper rotate when unusedFlag, harmless)
      if (~this.numFaces <= ~this.maxFaces) {   // numFaces >= maxFaces
         return -1;
      }
      this.faceNumVertices[this.numFaces] = n;
      this.faceVertices[this.numFaces] = vs;
      this.faceFillFront[this.numFaces] = front;
      this.faceFillBack[this.numFaces] = back;
      this.transformState = 1;
      return this.numFaces++;
   }

   /**
    * Split this model into a grid of {@code count} pieces, binning faces by centroid tile
    * {@code sumX/(n*pieceDx) + sumZ/(n*pieceDz)*rows}. Chops big terrain/scenery meshes into
    * per-tile sub-models for culling. {@code unused*} params are dead anti-tamper.
    */
   public GameModel[] split(int unused1, int rows, int pieceDz, int unused2, int count, int pieceMaxVertices, int pieceDx, boolean pickable, int unused3) {
      this.commit((byte)-28);
      int[] pieceNV = new int[count];
      int[] pieceNF = new int[count];
      for (int i = 0; i < count; i++) {
         pieceNV[i] = 0;
         pieceNF[i] = 0;
      }

      for (int f = 0; f < this.numFaces; f++) {
         int sumX = 0;
         int sumZ = 0;
         int n = this.faceNumVertices[f];
         int[] vs = this.faceVertices[f];
         for (int i = 0; i < n; i++) {
            sumX += this.vertexX[vs[i]];
            sumZ += this.vertexZ[vs[i]];
         }
         // obf: sumX/(n*pieceDx) + sumZ/(pieceDz*n) * rows  (var2=rows, var3=pieceDz, var7=pieceDx)
         int piece = sumX / (n * pieceDx) + sumZ / (pieceDz * n) * rows;
         pieceNV[piece] += n;
         pieceNF[piece]++;
      }

      GameModel[] pieces = new GameModel[count];
      for (int i = 0; i < count; i++) {
         if (pieceMaxVertices < pieceNV[i]) {
            pieceNV[i] = pieceMaxVertices;
         }
         pieces[i] = new GameModel(pieceNV[i], pieceNF[i], true, true, true, pickable, true);
         pieces[i].lightDiffuse = this.lightDiffuse;
         pieces[i].lightAmbience = this.lightAmbience;
      }

      for (int f = 0; f < this.numFaces; f++) {
         int sumX = 0;
         int sumZ = 0;
         int n = this.faceNumVertices[f];
         int[] vs = this.faceVertices[f];
         for (int i = 0; i < n; i++) {
            sumX += this.vertexX[vs[i]];
            sumZ += this.vertexZ[vs[i]];
         }
         // obf: sumX/(n*pieceDx) + rows * (sumZ/(pieceDz*n))  (var2=rows, var3=pieceDz, var7=pieceDx)
         int piece = sumX / (n * pieceDx) + rows * (sumZ / (pieceDz * n));
         this.copyLighting(vs, pieces[piece], n, f, 5916);
      }

      for (int i = 0; i < count; i++) {
         pieces[i].projectionPrepare((byte)71);
      }
      return pieces;
   }

   /**
    * Merge faces of each model in {@code pieces} into this one, re-resolving vertices and
    * (optionally) rebuilding {@link #faceTransGroups}. Backs the merging constructors.
    */
   private void merge(int from, GameModel[] pieces, boolean buildGroups, int to) {
      int totalFaces = 0;
      int totalVerts = 0;
      for (int i = 0; i < to; i++) {
         totalFaces += pieces[i].numFaces;
         totalVerts += pieces[i].numVertices;
      }
      this.allocate(totalFaces, totalVerts, 88);

      if (buildGroups) {
         this.faceTransGroups = new int[totalFaces][];
      }

      for (int i = from; i < to; i++) {
         GameModel src = pieces[i];
         src.commit((byte)-28);
         this.lightDirectionMagnitude = src.lightDirectionMagnitude;
         this.lightDirectionX = src.lightDirectionX;
         this.lightDirectionY = src.lightDirectionY;
         this.lightDiffuse = src.lightDiffuse;
         this.lightDirectionZ = src.lightDirectionZ;
         this.lightAmbience = src.lightAmbience;

         for (int srcF = 0; srcF < src.numFaces; srcF++) {
            int[] dstVs = new int[src.faceNumVertices[srcF]];
            int[] srcVs = src.faceVertices[srcF];
            for (int v = 0; v < src.faceNumVertices[srcF]; v++) {
               dstVs[v] = this.vertexAt(src.vertexX[srcVs[v]], src.vertexZ[srcVs[v]], src.vertexY[srcVs[v]], 107);
            }
            int dstF = this.createFace(src.faceNumVertices[srcF], dstVs, src.faceFillFront[srcF], src.faceFillBack[srcF], false);
            this.faceIntensity[dstF] = src.faceIntensity[srcF];
            this.normalScale[dstF] = src.normalScale[srcF];
            this.normalMagnitude[dstF] = src.normalMagnitude[srcF];

            if (buildGroups) {
               if (1 >= to) {
                  // Single source: copy its groups verbatim.
                  this.faceTransGroups[dstF] = new int[src.faceTransGroups[srcF].length];
                  for (int g = 0; g < src.faceTransGroups[srcF].length; g++) {
                     this.faceTransGroups[dstF][g] = src.faceTransGroups[srcF][g];
                  }
               } else {
                  // Multiple sources: prepend the source index so faces stay attributable.
                  this.faceTransGroups[dstF] = new int[src.faceTransGroups[srcF].length + 1];
                  this.faceTransGroups[dstF][0] = i;
                  for (int g = 0; g < src.faceTransGroups[srcF].length; g++) {
                     this.faceTransGroups[dstF][1 + g] = src.faceTransGroups[srcF][g];
                  }
               }
            }
         }
      }
      this.transformState = 1;
   }

   /**
    * Copy one face ({@code inF}) plus its vertices' lighting from this model into {@code dst}
    * (used by {@link #split} to populate each piece). {@code unusedGuard} is dead.
    */
   private void copyLighting(int[] srcVs, GameModel dst, int nV, int inF, int unusedGuard) {
      int[] dstVs = new int[nV];
      for (int i = 0; i < nV; i++) {
         int outV = dstVs[i] = dst.vertexAt(this.vertexX[srcVs[i]], this.vertexZ[srcVs[i]], this.vertexY[srcVs[i]], 107);
         dst.vertexIntensity[outV] = this.vertexIntensity[srcVs[i]];
         dst.vertexAmbience[outV] = this.vertexAmbience[srcVs[i]];
      }
      // unusedGuard != 5916 sets a tamper field; harmless.
      int outF = dst.createFace(nV, dstVs, this.faceFillFront[inF], this.faceFillBack[inF], false);
      // obf: if (!var2.db && !this.db)  — db == unpickable; faceTag (E) only exists when not unpickable
      if (!dst.unpickable && !this.unpickable) {
         dst.faceTag[outF] = this.faceTag[inF];
      }
      dst.faceIntensity[outF] = this.faceIntensity[inF];
      dst.normalScale[outF] = this.normalScale[inF];
      dst.normalMagnitude[outF] = this.normalMagnitude[inF];
   }

   // =============================================================================================
   // Lighting setters
   // =============================================================================================

   /** Set ambient/diffuse + light direction (no gouraud toggle). {@code unusedGuard} is dead. */
   public void setLight(int ambient, int diffuse, int dirY, int unusedGuard, int dirX, int dirZ) {
      this.lightAmbience = -(4 * diffuse) + 256;
      this.lightDiffuse = (64 - ambient) * 16 + 128;
      // (original anti-tamper sets lightDirectionY=-67 when unusedGuard > -110, harmless)
      if (this.unlit) {
         return;
      }
      this.lightDirectionX = dirX;
      this.lightDirectionZ = dirZ;
      this.lightDirectionY = dirY;
      this.lightDirectionMagnitude = (int)Math.sqrt((double)(dirZ * dirZ + dirY * dirY + dirX * dirX));
      this.light(-102);
   }

   /** Set ambient/diffuse + per-face gouraud/flat flag + light direction. {@code unusedGuard} is dead. */
   public void setLight(int dirZ, int dirY, int diffuse, int dirX, boolean gouraud, int ambient, int unusedGuard) {
      this.lightDiffuse = (-ambient + 64) * 16 + 128;
      this.lightAmbience = 256 - diffuse * 4;
      if (this.unlit) {
         return;
      }
      for (int f = 0; f < this.numFaces; f++) {
         this.faceIntensity[f] = gouraud ? this.magic : 0;
      }
      this.lightDirectionX = dirX;
      this.lightDirectionZ = dirZ;
      this.lightDirectionY = dirY;
      this.lightDirectionMagnitude = (int)Math.sqrt((double)(dirY * dirY + (dirX * dirX - -(dirZ * dirZ))));
      this.light(-121);
   }

   /** Set just the light direction (x,y,z) and re-bake lighting. {@code unusedFlag} is dead. */
   public void setLightDirection(boolean unusedFlag, int dirX, int dirY, int dirZ) {
      // (original anti-tamper sets transformState=71 when unusedFlag, harmless)
      if (this.unlit) {
         return;
      }
      this.lightDirectionZ = dirZ;
      this.lightDirectionY = dirY;
      this.lightDirectionX = dirX;
      this.lightDirectionMagnitude = (int)Math.sqrt((double)(dirZ * dirZ + (dirY * dirY + dirX * dirX)));
      this.light(52);
   }

   /** Set the ambient bias of a single vertex. {@code guard} must be -61. */
   public void setVertexAmbience(int vertex, int ambience, byte guard) {
      if (guard != -61) {
         return;
      }
      this.vertexAmbience[vertex] = (byte)ambience;
   }

   // =============================================================================================
   // Transform mutators (lazy: only flag/accumulate; apply() does the work)
   // =============================================================================================

   /** Rotate incrementally by (yaw,pitch,roll), each wrapped to 0..255. {@code unusedGuard} is dead. */
   public void rotate(int yaw, int unusedGuard, int roll, int pitch) {
      this.orientationPitch = pitch + this.orientationPitch & 0xFF;
      this.orientationRoll = roll + this.orientationRoll & 0xFF;
      if (unusedGuard != -31616) {
         return;
      }
      this.orientationYaw = 0xFF & this.orientationYaw + yaw;
      this.determineTransformKind((byte)-105);
      this.transformState = 1;
   }

   /** Set the absolute orientation (yaw,pitch,roll), each wrapped to 0..255. */
   public void orient(int yaw, int unusedGuard, int pitch, int roll) {
      this.orientationYaw = yaw & 0xFF;
      // (original anti-tamper call when unusedGuard != -999999, harmless)
      this.orientationPitch = pitch & 0xFF;
      this.orientationRoll = roll & 0xFF;
      this.determineTransformKind((byte)-117);
      this.transformState = 1;
   }

   /** Translate incrementally by (x,y,z). {@code keepCache} false drops the cached transformed-Z. */
   public void translate(int x, int y, int z, boolean keepCache) {
      this.baseZ += z;
      this.baseY += y;
      if (!keepCache) {
         this.vertexTransformedZ = null;
      }
      this.baseX += x;
      this.determineTransformKind((byte)-127);
      this.transformState = 1;
   }

   /** Set absolute model position (z, _, y, x). {@code unusedGuard} is dead. */
   public void place(int z, int unusedGuard, int y, int x) {
      this.baseZ = z;
      this.baseX = x;
      this.baseY = y;
      // (original anti-tamper translate when unusedGuard > -112, harmless)
      this.determineTransformKind((byte)-114);
      this.transformState = 1;
   }

   /** Copy the full pose (orientation + base position) from another model. {@code unusedGuard} is dead. */
   public void copyPosition(GameModel other, int unusedGuard) {
      this.baseZ = other.baseZ;
      this.orientationYaw = other.orientationYaw;
      this.baseY = other.baseY;
      this.orientationRoll = other.orientationRoll;
      this.orientationPitch = other.orientationPitch;
      this.baseX = other.baseX;
      this.determineTransformKind((byte)-113);
      this.transformState = 1;
   }

   /**
    * Recompute {@link #transformKind} from the pending params: 4 if any shear is active, else 3 if
    * any scale, else 2 if any orientation, else 1 if any translation, else 0. {@code guard} dead.
    */
   private void determineTransformKind(byte unusedGuard) {
      if (256 != this.shearYx || this.shearYz != 256 || this.shearZx != 256 || this.shearXz != 256
            || this.shearXy != 256 || this.shearZy != 256) {
         this.transformKind = 4;
      } else if (this.scaleFx != 256 || this.scaleFy != 256 || this.scaleFz != 256) {
         this.transformKind = 3;
      } else if (this.orientationPitch != 0 || this.orientationYaw != 0 || this.orientationRoll != 0) {
         this.transformKind = 2;
      } else if (this.baseY != 0 || this.baseZ != 0 || this.baseX != 0) {
         this.transformKind = 1;
      } else {
         this.transformKind = 0;
      }
   }

   // =============================================================================================
   // Transform application (vertex math)
   // =============================================================================================

   /**
    * Translate every transformed vertex: X += dx, Y += dy, Z += dz. {@code from} is the start index
    * (always 0). NB: {@link #apply} calls this as {@code translateVertices(0, baseZ, baseY, baseX)}
    * (obf {@code d(0, xb, Sb, r)}), so dy carries baseZ and dz carries baseY — matching the obf bytecode.
    */
   private void translateVertices(int from, int dy, int dz, int dx) {
      int v = from;
      while (this.numVertices > v) {
         this.vertexTransformedX[v] += dx;
         this.vertexTransformedY[v] += dy;
         this.vertexTransformedZ[v] += dz;
         v++;
      }
   }

   /**
    * Rotate every transformed vertex by yaw/pitch/roll (0..255), using the 9-bit sine table
    * {@code pa.a} (sin*32768, undone with {@code >> 15}). {@code guard} is a dead anti-tamper param.
    */
   private void applyRotation(int guard, int pitch, int yaw, int roll) {
      if (guard >= -14) {
         this.faceNormalY = null;   // anti-tamper trap (never reached at runtime)
      }
      for (int v = 0; v < this.numVertices; v++) {
         if (yaw != 0) {
            int sin = ImageLoader.SIN_512[yaw];
            int cos = ImageLoader.SIN_512[yaw + 256];
            int t = this.vertexTransformedX[v] * cos + this.vertexTransformedY[v] * sin >> 15;
            this.vertexTransformedY[v] = -(sin * this.vertexTransformedX[v]) + this.vertexTransformedY[v] * cos >> 15;
            this.vertexTransformedX[v] = t;
         }
         if (pitch != 0) {
            int sin = ImageLoader.SIN_512[pitch];
            int cos = ImageLoader.SIN_512[pitch + 256];
            int t = -(sin * this.vertexTransformedZ[v]) + cos * this.vertexTransformedY[v] >> 15;
            this.vertexTransformedZ[v] = sin * this.vertexTransformedY[v] - -(cos * this.vertexTransformedZ[v]) >> 15;
            this.vertexTransformedY[v] = t;
         }
         if (roll != 0) {
            int sin = ImageLoader.SIN_512[roll];
            int cos = ImageLoader.SIN_512[roll + 256];
            int t = sin * this.vertexTransformedZ[v] + this.vertexTransformedX[v] * cos >> 15;
            this.vertexTransformedZ[v] = -(this.vertexTransformedX[v] * sin) + this.vertexTransformedZ[v] * cos >> 15;
            this.vertexTransformedX[v] = t;
         }
      }
   }

   /** Scale every transformed vertex by (fx,fy,fz) in 8.8 fixed point ({@code >> 8}). {@code guard} dead. */
   private void applyScale(int fx, int guard, int fz, int fy) {
      // (original anti-tamper call when guard != -27483, harmless)
      for (int v = 0; v < this.numVertices; v++) {
         this.vertexTransformedX[v] = this.vertexTransformedX[v] * fx >> 8;
         this.vertexTransformedY[v] = this.vertexTransformedY[v] * fy >> 8;
         this.vertexTransformedZ[v] = fz * this.vertexTransformedZ[v] >> 8;
      }
   }

   /**
    * Apply the six per-vertex shears (8.8 fixed point, {@code >> 8}); each non-zero factor mixes
    * one transformed axis into another. Called by {@link #apply} as
    * {@code applyShear(shearXy, shearXz, shearYx, shearYz, shearZx, shearZy, -127)}.
    * {@code unusedGuard} is dead anti-tamper.
    */
   private void applyShear(int shearXy, int shearXz, int shearYx, int shearYz, int shearZx, int shearZy, byte unusedGuard) {
      int unusedFold = 118 / ((-80 - unusedGuard) / 46);
      for (int v = 0; v < this.numVertices; v++) {
         if (shearYx != 0) {   // X += Y * shearYx
            this.vertexTransformedX[v] += this.vertexTransformedY[v] * shearYx >> 8;
         }
         if (shearYz != 0) {   // Z += Y * shearYz
            this.vertexTransformedZ[v] += shearYz * this.vertexTransformedY[v] >> 8;
         }
         if (shearZx != 0) {   // X += Z * shearZx
            this.vertexTransformedX[v] += shearZx * this.vertexTransformedZ[v] >> 8;
         }
         if (shearXz != 0) {   // Y += Z * shearXz
            this.vertexTransformedY[v] += shearXz * this.vertexTransformedZ[v] >> 8;
         }
         if (shearXy != 0) {   // Z += X * shearXy
            this.vertexTransformedZ[v] += shearXy * this.vertexTransformedX[v] >> 8;
         }
         if (shearZy != 0) {   // Y += X * shearZy
            this.vertexTransformedY[v] += this.vertexTransformedX[v] * shearZy >> 8;
         }
      }
   }

   /**
    * Recompute the model's object-space bounding box and {@link #diameter} from the transformed
    * vertices, plus (unless {@link #isolated}) the per-face bounds. {@code seed} is the min-init
    * value (999999) passed by the original.
    */
   private void computeBounds(int seed) {
      this.boundX1 = 999999;
      this.boundZ1 = 999999;
      this.boundZ2 = -999999;
      this.boundY1 = seed;        // original passes 999999 here
      this.boundX2 = -999999;
      this.boundY2 = -999999;
      this.diameter = -999999;

      for (int f = 0; f < this.numFaces; f++) {
         int[] vs = this.faceVertices[f];
         int n = this.faceNumVertices[f];
         int v0 = vs[0];
         int zMin, zMax;
         zMin = zMax = this.vertexTransformedZ[v0];
         int yMin, yMax;
         yMin = yMax = this.vertexTransformedY[v0];
         int xMin, xMax;
         xMin = xMax = this.vertexTransformedX[v0];

         for (int i = 0; i < n; i++) {
            int v = vs[i];
            if (this.vertexTransformedZ[v] < zMin) {
               zMin = this.vertexTransformedZ[v];
            } else if (this.vertexTransformedZ[v] > zMax) {
               zMax = this.vertexTransformedZ[v];
            }
            if (this.vertexTransformedY[v] < yMin) {
               yMin = this.vertexTransformedY[v];
            } else if (this.vertexTransformedY[v] > yMax) {
               yMax = this.vertexTransformedY[v];
            }
            if (this.vertexTransformedX[v] < xMin) {
               xMin = this.vertexTransformedX[v];
            } else if (this.vertexTransformedX[v] > xMax) {
               xMax = this.vertexTransformedX[v];
            }
         }

         if (!this.isolated) {
            // obf: w[f]=xMin, n[f]=xMax, Q[f]=yMin, Lb[f]=yMax, Zb[f]=zMin, Eb[f]=zMax
            this.faceBoundLeft[f] = xMin;     // w
            this.faceBoundRight[f] = xMax;    // n
            this.faceBoundBottom[f] = yMin;   // Q
            this.faceBoundTop[f] = yMax;      // Lb
            this.faceBoundNear[f] = zMin;     // Zb
            this.faceBoundFar[f] = zMax;      // Eb
         }
         if (xMax - xMin > this.diameter) {
            this.diameter = xMax - xMin;
         }
         if (yMax - yMin > this.diameter) {
            this.diameter = yMax - yMin;
         }
         if (zMax - zMin > this.diameter) {
            this.diameter = zMax - zMin;
         }
         if (this.boundX2 < xMax) {
            this.boundX2 = xMax;
         }
         if (this.boundX1 > xMin) {
            this.boundX1 = xMin;
         }
         if (yMax > this.boundY2) {
            this.boundY2 = yMax;
         }
         if (this.boundY1 > yMin) {
            this.boundY1 = yMin;
         }
         if (zMax > this.boundZ2) {
            this.boundZ2 = zMax;
         }
         if (this.boundZ1 > zMin) {
            this.boundZ1 = zMin;
         }
      }
   }

   /**
    * Bake per-face and per-vertex light intensities from the face normals and the light direction.
    * Faces flagged {@code magic} get smooth (per-vertex) shading; others keep their flat intensity.
    * {@code unusedGuard} is dead anti-tamper.
    */
   private void light(int unusedGuard) {
      if (this.unlit) {
         return;
      }
      int divisor = this.lightDiffuse * this.lightDirectionMagnitude >> 8;
      for (int f = 0; f < this.numFaces; f++) {
         if (this.faceIntensity[f] != this.magic) {
            this.faceIntensity[f] = (this.faceNormalY[f] * this.lightDirectionY
                  + (this.faceNormalX[f] * this.lightDirectionX - -(this.lightDirectionZ * this.faceNormalZ[f]))) / divisor;
         }
      }

      int[] normalX = new int[this.numVertices];
      int[] normalY = new int[this.numVertices];
      int[] normalZ = new int[this.numVertices];
      int[] normalCount = new int[this.numVertices];
      for (int v = 0; v < this.numVertices; v++) {
         normalX[v] = 0;
         normalY[v] = 0;
         normalZ[v] = 0;
         normalCount[v] = 0;
      }

      for (int f = 0; f < this.numFaces; f++) {
         if (this.faceIntensity[f] == this.magic) {
            for (int i = 0; i < this.faceNumVertices[f]; i++) {
               int v = this.faceVertices[f][i];
               normalX[v] += this.faceNormalX[f];
               normalY[v] += this.faceNormalY[f];
               normalZ[v] += this.faceNormalZ[f];
               normalCount[v]++;
            }
         }
      }

      for (int v = 0; v < this.numVertices; v++) {
         if (normalCount[v] > 0) {
            this.vertexIntensity[v] = (normalZ[v] * this.lightDirectionZ
                  + (normalX[v] * this.lightDirectionX + normalY[v] * this.lightDirectionY)) / (divisor * normalCount[v]);
         }
      }
   }

   /**
    * Recompute each face's surface normal from its first three transformed vertices (cross product,
    * scaled into 16.16 fixed point), then re-bake lighting via {@link #light}. {@code guard} must be 14.
    */
   private void relight(byte guard) {
      if (this.unlit && this.isolated) {
         return;
      }
      if (guard != 14) {
         return;
      }
      for (int f = 0; f < this.numFaces; f++) {
         int[] verts = this.faceVertices[f];
         int aX = this.vertexTransformedX[verts[0]];
         int aY = this.vertexTransformedY[verts[0]];
         int aZ = this.vertexTransformedZ[verts[0]];
         int bX = -aX + this.vertexTransformedX[verts[1]];
         int bY = this.vertexTransformedY[verts[1]] - aY;
         int bZ = -aZ + this.vertexTransformedZ[verts[1]];
         int cX = -aX + this.vertexTransformedX[verts[2]];
         int cY = this.vertexTransformedY[verts[2]] - aY;
         int cZ = -aZ + this.vertexTransformedZ[verts[2]];
         int normX = bY * cZ - cY * bZ;
         int normY = -(bX * cZ) + cX * bZ;   // == bZ*cX - cZ*bX
         int normZ = -(cX * bY) + bX * cY;
         // Halve the normal until each component fits in +-8192 so the magnitude sqrt won't overflow.
         while (normX > 8192 || normY > 8192 || normZ > 8192 || normX < -8192 || normY < -8192 || normZ < -8192) {
            normX >>= 1;
            normY >>= 1;
            normZ >>= 1;
         }
         int mag = (int)(Math.sqrt((double)(normY * normY + (normX * normX + normZ * normZ))) * 256.0);
         if (mag <= 0) {
            mag = 1;
         }
         this.faceNormalX[f] = normX * 65536 / mag;
         this.faceNormalY[f] = 65536 * normY / mag;
         this.faceNormalZ[f] = normZ * 65535 / mag;   // NB: Z uses 65535, matching the original
         this.normalScale[f] = -1;
      }
      this.light(guard ^ 0xFFFFFFAB);
   }

   /**
    * Re-bake the transformed vertices, bounds and lighting if {@link #transformState} is dirty.
    * State 2 is the cheap path (re-copy base verts; set the "never cull" bound sentinels). State 1
    * is the full path (copy, apply rotate/scale/shear/translate, recompute bounds + relight).
    * {@code guard} must be 7972 (else a dead anti-tamper branch runs).
    */
   public void apply(int guard) {
      if (guard != 7972) {
         this.relight((byte)0);   // dead anti-tamper branch
      }
      if (this.transformState == 2) {
         this.transformState = 0;
         for (int v = 0; v < this.numVertices; v++) {
            this.vertexTransformedX[v] = this.vertexX[v];
            this.vertexTransformedY[v] = this.vertexY[v];
            this.vertexTransformedZ[v] = this.vertexZ[v];
         }
         // "Always inside the frustum" sentinels (min = -9999999, max = +9999999).
         this.boundX2 = 9999999;
         this.boundY2 = 9999999;
         this.boundY1 = -9999999;
         this.boundZ2 = 9999999;
         this.boundZ1 = -9999999;
         this.diameter = 9999999;
         this.boundX1 = -9999999;
         return;
      }
      if (this.transformState == 1) {
         this.transformState = 0;
         for (int v = 0; v < this.numVertices; v++) {
            this.vertexTransformedX[v] = this.vertexX[v];
            this.vertexTransformedY[v] = this.vertexY[v];
            this.vertexTransformedZ[v] = this.vertexZ[v];
         }
         if (this.transformKind >= 2) {
            this.applyRotation(-53, this.orientationPitch, this.orientationYaw, this.orientationRoll);
         }
         if (this.transformKind >= 3) {
            this.applyScale(this.scaleFx, -27483, this.scaleFz, this.scaleFy);
         }
         if (this.transformKind >= 4) {
            this.applyShear(this.shearXy, this.shearXz, this.shearYx, this.shearYz, this.shearZx, this.shearZy, (byte)-127);
         }
         if (this.transformKind >= 1) {
            // obf: d(0, xb, Sb, r) -> dy=baseZ (into Y), dz=baseY (into Z), dx=baseX (into X)
            this.translateVertices(0, this.baseZ, this.baseY, this.baseX);
         }
         this.computeBounds(999999);
         this.relight((byte)14);
      }
   }

   /**
    * Project the transformed vertices into camera/screen space for the current camera pose, after a
    * frustum cull (sets {@link #visible}). Sine/cosine come from the 11-bit camera table
    * {@code pa.j}; the perspective divide is {@code (coord << viewDist) / z} when {@code z >= clipNear}.
    *
    * <p>The frustum bounds live on other obfuscated classes (da/m/oa/aa/nb statics) in this build;
    * they correspond to mudclient204's {@code Scene.frustumNearZ/FarZ/MinX/MaxX/MinY/MaxY}.
    * The argument order matches the obfuscated callsite (camera Y/X/Z and roll/yaw/pitch are
    * interleaved). {@code guard} is a dead anti-tamper param.
    */
   public void project(int cameraY, int viewDist, int cameraX, byte guard, int cameraZ, int cameraRoll, int cameraYaw, int cameraPitch, int clipNear) {
      this.apply(7972);
      // Frustum reject (faithful to the obfuscated static field reads; logic == mudclient204).
      if (~ClientStream.frustumNearZ > ~this.boundZ1 || SocketFactory.frustumFarZ > this.boundZ2
            || ~this.boundX1 < ~NameHash.frustumMinX || BZip.frustumMaxX > this.boundX2
            || this.boundY1 > DataStore.frustumMinY || this.boundY2 < BZip.frustumMaxY) {
         this.visible = false;
         return;
      }
      this.visible = true;

      // (original anti-tamper allocates projection arrays via c() when guard > -105, harmless)
      int yawSin = 0, yawCos = 0, pitchSin = 0, pitchCos = 0, rollSin = 0, rollCos = 0;
      if (cameraYaw != 0) {
         yawSin = ImageLoader.SIN_2048[cameraYaw];
         yawCos = ImageLoader.SIN_2048[1024 + cameraYaw];
      }
      if (cameraPitch != 0) {
         pitchSin = ImageLoader.SIN_2048[cameraPitch];
         pitchCos = ImageLoader.SIN_2048[cameraPitch + 1024];
      }
      if (cameraRoll != 0) {
         rollSin = ImageLoader.SIN_2048[cameraRoll];
         rollCos = ImageLoader.SIN_2048[cameraRoll + 1024];
      }

      for (int v = 0; v < this.numVertices; v++) {
         int x = this.vertexTransformedX[v] - cameraX;
         int y = this.vertexTransformedY[v] - cameraY;
         if (cameraYaw != 0) {   // yaw rotates the X/Y pair
            int t = y * yawSin - -(yawCos * x) >> 15;
            y = -(x * yawSin) + y * yawCos >> 15;
            x = t;
         }
         int z = this.vertexTransformedZ[v] - cameraZ;
         if (cameraRoll != 0) {  // roll rotates the X/Z pair
            int t = rollCos * x + z * rollSin >> 15;
            z = -(x * rollSin) + rollCos * z >> 15;
            x = t;
         }
         if (cameraPitch != 0) { // pitch rotates the Y/Z pair
            int t = y * pitchCos - pitchSin * z >> 15;
            z = pitchSin * y - -(pitchCos * z) >> 15;
            y = t;
         }
         if (z >= clipNear) {
            this.vertexViewX[v] = (x << viewDist) / z;
            this.vertexViewY[v] = (y << viewDist) / z;
         } else {
            this.vertexViewX[v] = x << viewDist;
            this.vertexViewY[v] = y << viewDist;
         }
         this.projectVertexX[v] = x;
         this.projectVertexY[v] = y;
         this.projectVertexZ[v] = z;
      }
   }

   /** Fold the current transform back into the base vertices and reset all transform params. {@code guard} dead. */
   public void commit(byte guard) {
      this.apply(7972);
      for (int v = 0; v < this.numVertices; v++) {
         this.vertexX[v] = this.vertexTransformedX[v];
         this.vertexY[v] = this.vertexTransformedY[v];
         this.vertexZ[v] = this.vertexTransformedZ[v];
      }
      if (guard != -28) {
         this.relight((byte)4);   // dead anti-tamper branch
      }
      this.orientationYaw = 0;
      this.orientationPitch = 0;
      this.orientationRoll = 0;
      this.baseX = 0;
      this.baseY = 0;
      this.baseZ = 0;
      this.scaleFx = 256;
      this.scaleFy = 256;
      this.scaleFz = 256;
      this.shearXy = 256;
      this.shearXz = 256;
      this.shearYx = 256;
      this.shearYz = 256;
      this.shearZx = 256;
      this.shearZy = 256;
      this.transformKind = 0;
   }

   // =============================================================================================
   // Copies
   // =============================================================================================

   /** Shallow single-piece copy that carries depth + transparency. {@code unusedGuard} is dead. */
   public GameModel copy(int unusedGuard) {
      GameModel[] pieces = new GameModel[1];
      // (original anti-tamper reduce when unusedGuard != -2, harmless)
      pieces[0] = this;
      GameModel copy = new GameModel(pieces, 1);
      copy.transparent = this.transparent;
      copy.depth = this.depth;
      return copy;
   }

   /**
    * Single-piece copy with explicit topology flags; carries depth. {@code unusedGuard} is dead.
    * The flags are forwarded to the merging ctor in the same shuffled order as the bytecode
    * ({@code new GameModel(pieces, 1, unlit, unpickable, isolated, autocommit)}).
    */
   public GameModel copy(boolean autocommit, int unusedGuard, boolean isolated, boolean unlit, boolean unpickable) {
      GameModel[] pieces = new GameModel[]{this};
      GameModel copy = new GameModel(pieces, 1, unlit, unpickable, isolated, autocommit);
      copy.depth = this.depth;
      return copy;
   }

   // =============================================================================================
   // Texture / colour id table (shared static helpers)
   // =============================================================================================

   /**
    * Resolve a texture/material name to its index in the shared name table, appending it if new.
    * Returns 0 for the "Invisible" sentinel name. {@code guard} must be 91 (else an anti-tamper
    * guard nulls a scratch array).
    */
   public static int textureId(byte guard, String name) {
      if (name.equalsIgnoreCase(SCRAMBLED[5])) {
         return 0;
      }
      if (guard != 91) {
         sharedScratch = null;
      }
      // DEOB FIX: clean ca.a (ca.java:1434-1452) registers names into ub.c[ia.b++],
      // i.e. NameTable.modelNames (obf c, pre-allocated new String[5000]) indexed by
      // SpriteScaler.modelNameCount (obf b). The previous mapping to
      // NameTable.textureNames (obf b, the GameData equip-slot table, allocated later
      // in initGameData) + SpriteScaler.spriteCount (obf h) was a field-mapping bug and
      // NPEd here because textureNames was still null at this call site.
      for (int i = 0; i < SpriteScaler.modelNameCount; i++) {
         if (NameTable.modelNames[i].equalsIgnoreCase(name)) {
            return i;
         }
      }
      NameTable.modelNames[SpriteScaler.modelNameCount++] = name;
      return SpriteScaler.modelNameCount - 1;
   }

   // =============================================================================================
   // Stream decode helpers
   // =============================================================================================

   /**
    * Decode one signed base-64 value from {@code data} at {@link #dataPtr} (skipping CR/LF), as
    * {@code hi*4096 + mid*64 + lo - 0x20000}; the 0x1e240 (123456) sentinel maps to {@link #magic}.
    * {@code guard} must be 76. The alphabet table lives in {@link ImageLoader} ({@code pa.d}).
    */
   private int readBase64(byte guard, byte[] data) {
      if (guard != 76) {
         return 108;
      }
      while (~data[this.dataPtr] == -11 || -14 == ~data[this.dataPtr]) {   // byte == 10 (LF) or == 13 (CR)
         this.dataPtr++;
      }
      int hi = ImageLoader.BASE64_DECODE[0xFF & data[this.dataPtr++]];
      int mid = ImageLoader.BASE64_DECODE[0xFF & data[this.dataPtr++]];
      int lo = ImageLoader.BASE64_DECODE[data[this.dataPtr++] & 0xFF];
      int value = -131072 + mid * 64 + (4096 * hi + lo);
      if (value == 0x1e240) {
         return this.magic;
      }
      return value;
   }

   // =============================================================================================
   // Obfuscated-string codec (XOR-based; only builds the error-prefix pool)
   // =============================================================================================

   /** Stage 1: XOR the leading char of a 1-char string by 0x25 (handles the short markers). */
   private static char[] unscramble(String s) {
      char[] chars = s.toCharArray();
      if (chars.length < 2) {
         chars[0] = (char)(chars[0] ^ 0x25);
      }
      return chars;
   }

   /** Stage 2: XOR each char by a position-cycled key {27,119,10,61,37}; returns an interned String. */
   private static String unscramble(char[] chars) {
      for (int i = 0; i < chars.length; i++) {
         int key;
         switch (i % 5) {
            case 0:
               key = 27;
               break;
            case 1:
               key = 119;
               break;
            case 2:
               key = 10;
               break;
            case 3:
               key = 61;
               break;
            default:
               key = 37;
         }
         chars[i] = (char)(chars[i] ^ key);
      }
      return new String(chars).intern();
   }
}
