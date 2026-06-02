package client.world;

import java.io.IOException;

import client.scene.Scene;          // obf: lb  — 3D model-list renderer (the ca[] models register here)
import client.scene.GameModel;      // obf: ca
import client.scene.Surface;        // obf: ua  — software 2D framebuffer (minimap blits, blackScreen)
import client.scene.SurfaceImageProducer; // obf: fb — holds object→model index table (fb.f)
import client.net.ClientStream;     // obf: da  — static (B,R,G) colour packer rgb (clean da.a) + GameData.tileType table (da.N)
import client.net.StreamBase;       // obf: ib  — StreamBase.bitwiseAnd(a,b)=a&b helper (obf ib.a); ib.d[]=wall/door height table
import client.net.StringCodec;      // obf: u   — u.a[] = GameData.wallObjectAdjacent flag table
import client.net.ChatCipher;       // obf: v   — v.a[] = wall front-face colour table
import client.util.StreamFactory;   // obf: na  — loads a cache record → byte[] (na.a)
import client.util.ArrayUtil;       // obf: ab  — System.arraycopy helper (ab.a)
import client.util.Utility;         // obf: mb  — Utility.a[] = GameData.objectType table
import client.util.ErrorHandler;    // obf: i   — i.g[] = roof ridge-height table
import client.util.DecodeBuffer;    // obf: ac  — ac.l[] = decoration "blocks-projectiles" flag table
import client.scene.SpriteDecoder;  // obf: ea  — bzip2 decompressor (clean k.java:528 ea.a -> decompress)
import client.data.CacheFile;       // obf: d   — CacheFile.or(a,b)=a|b helper (obf d.a); d.g[]=roof-texture table
import client.data.NameTable;       // obf: ub  — ub.g[] = GameData.objectWidth table
import client.data.RecordLoader;    // obf: f   — f.f[] = GameData.objectHeight table
// NOTE: the raw .jm reader is GameCharacter.readFromStore (clean ta.a), NOT EntityDef.a;
// EntityDef (obf t) is no longer referenced from this class. (RENDER_DEOB_GAPS world-terrain)
import client.ui.Panel;             // obf: qa  — qa.K[] = GameData.tileDecoration colour table

/**
 * World — landscape / terrain manager (obfuscated class {@code k}, RSC rev ~235).
 *
 * <p>This is the {@code World} of the mudclient204 oracle: it loads the map
 * files ({@code .hei} heights/colours, {@code .dat} walls/decoration,
 * {@code .loc} objects, raw {@code .jm} fallback), decodes them into the
 * per-quadrant terrain grids, builds the terrain / wall / roof
 * {@link GameModel}s, registers them with the {@link Scene} renderer, computes
 * ground {@link #getElevation elevation}, and runs the breadth-first
 * {@link #route} pathfinder over the passability grid.</p>
 *
 * <p>NAMING (per docs/NAMING.md, CORRECTED): obf {@code k}=World (this class,
 * package {@code client.world}); obf {@code lb}=Scene (the 3D renderer this
 * class feeds, field {@link #scene}). {@code da}=ClientStream owns the colour
 * packer (clean {@code da.a}), referenced as {@link ClientStream#rgb}; its
 * argument order is {@code (B,R,G)} (see ctor Javadoc).</p>
 *
 * <p>Region layout: the visible region is 96×96 tiles stored as four 48×48
 * quadrants (indices 0..3). Every grid accessor resolves an absolute (x,y) in
 * 0..95 to a quadrant + local offset via the repeated four-case block.</p>
 *
 * <h3>Field-identity correction</h3>
 * Verified from the {@code .dat} decode order in the (bytecode-only)
 * {@link #loadMapData} method and the IOException border sentinel: obf
 * {@code A}=wallsRoof, {@code R}=tileDecoration, {@code mb}=tileDirection
 * (an earlier pass had these three swapped).
 *
 * <h3>Obfuscation stripped while reading</h3>
 * the opaque predicate {@code boolean bl = client.vh;} (always false) and its
 * dead branches; the per-method profiling counters ({@code ++X;}); the
 * per-method {@code try{…}catch(RuntimeException){throw ErrorHandler.a(e,"sig")}}
 * wrappers; anti-tamper dummy-parameter guards ({@code if (p != <magic>)
 * return;} / junk-modulo expressions); and the XOR string pool (obf {@code ob}).
 * The {@code ~a > ~b} idioms are normalised to {@code a < b}.
 */
public final class World {

   // ───────────────────────── region constants ─────────────────────────
   public static final int colourTransparent = 12345678; // obf literal — "no triangle" sentinel colour
   public static final int regionWidth = 96;
   public static final int regionHeight = 96;
   public static final int anInt585 = 128;                // world units per tile

   // ───────────────────────── collaborators ────────────────────────────
   private Scene scene;     // obf: c  — 3D renderer the built models are added to
   private Surface surface; // obf: U  — software framebuffer (minimap, blackScreen)

   // ───────────────────────── world flags ──────────────────────────────
   private boolean memberWorld;      // obf: H  — build object models even when adjacency-hidden (members area)
   private boolean worldInitialised; // obf: nb — a section has been loaded (⇒ dispose before reset)
   public boolean playerAlive;              // obf: Z  — (oracle aBoolean592) render the parent fence model when set
   public int baseMediaSprite;              // obf: x  — base sprite id for tile decorations (=750)

   // ───────────────────── per-quadrant terrain grids ([4][2304]) ────────
   private byte[][] terrainHeight;   // obf: L  — tile heights (×3 when read)
   private byte[][] terrainColour;   // obf: eb — terrain colour index
   private byte[][] wallsNorthsouth; // obf: f  — N/S wall id
   private byte[][] wallsEastwest;   // obf: P  — E/W wall id
   private int[][]  wallsDiagonal;   // obf: s  — diagonal wall id (+12000 inverted, +48000 object)
   private byte[][] wallsRoof;       // obf: A  — roof wall id   [CORRECTED: was tileDecoration]
   private byte[][] tileDecoration;  // obf: R  — tile overlay/decoration id [CORRECTED: was tileDirection]
   private byte[][] tileDirection;   // obf: mb — tile/object direction [CORRECTED: was wallsRoof]

   // ─────────────────────── derived / scratch grids ────────────────────
   public int[][] objectAdjacency;          // obf: bb — [96][96] passability/adjacency bitmask
   private int[][] routeVia;         // obf: B  — [96][96] BFS came-from directions (route())
   private int[][] terrainHeightLocal; // obf: ab — [96][96] vertex heights used while meshing
   private int[] terrainColours;     // obf: w  — [256] palette: ground/grass/dirt/sand ramps

   // ───────────────────────── built models ─────────────────────────────
   private GameModel[] terrainModels;// obf: F  — [64] ground-tile models (8×8 grid)
   public GameModel[][] wallModels;         // obf: g  — [4][64] vertical wall models per plane
   public GameModel[][] roofModels;         // obf: db — [4][64] roof models per plane
   private GameModel parentModel;    // obf: kb — shared builder model reused across passes

   // ─────────────────── per-face tile coordinate maps ──────────────────
   public int[] localX;                     // obf: q  — [18432] face-id → tile x
   public int[] localY;                     // obf: E  — [18432] face-id → tile y

   // ───────────────────── raw map archive packs ────────────────────────
   public byte[] landscapePack;             // obf: Q  — free landscape archive (.hei)
   public byte[] memberLandscapePack;       // obf: I  — members landscape archive
   // FIELD-LABEL FIX (RENDER_DEOB_GAPS world-terrain; clean client.m()@client.java:5549-5560):
   // gb is loaded UNCONDITIONALLY (Hh.gb = a(il[602],...)) = the FREE map archive = mapPack;
   // m is loaded only `if (this.Pg)` (members flag) (Hh.m = a(il[601],...)) = memberMapPack.
   // The earlier labels (m=mapPack, gb=memberMapPack) were inverted vs the bytecode.
   public byte[] mapPack;                   // obf: gb — free map archive (.dat/.loc)
   public byte[] memberMapPack;             // obf: m  — members map archive

   /**
    * BMP/surface width, written by the BMP loader from header bytes 12-13.
    * BEHAVIORAL FIX (RENDER_DEOB_GAPS surface-sprites; map row "World ? -> surfaceWidth"):
    * clean k.java:50 `static int o`; pa.java:21 writes `k.o = var2[12] + var2[13]*256`
    * (ImageLoader.loadBmpImage) and lb.java:3770 reads `k.o` (Scene.flushToImage).
    * Extracted from the dead-counter list below (obf {@code o}) into a named LIVE field
    * so ImageLoader/Scene reference World.surfaceWidth (not a fabricated carrier).
    * obf: o
    */
   public static int surfaceWidth;

   // ────────────────── profiling counters (dead; kept named) ───────────
   // (obf `o` removed from this list -> promoted to the live field surfaceWidth above)
   static int lb, z, ib, cb, t, u, N, j, C, h, n, jb, a, l, Y, d, M, r, p,
              J, K, T, k, fb, S, X, b, O, W, D, V, y, i, hb, v;
   public static long e = 0L;                  // obf: e — unused profiling accumulator
   public static String[] G = new String[100]; // obf: G — unused scratch string table

   // ───────────────── decoded XOR string pool (obf: ob) ─────────────────
   private static final String STR_UNPACKING  = "Unpacking ";        // ob[18]
   private static final String STR_HEI        = ".hei";              // ob[26]
   private static final String STR_MAPS_DIR   = "../gamedata/maps/"; // ob[28]
   private static final String STR_DAT        = ".dat";              // ob[29]
   private static final String STR_LOC        = ".loc";              // ob[30]
   private static final String STR_JM         = ".jm";               // ob[31]
   private static final String STR_NULL_ROOF  = "null roof!";        // ob[40]

   // ── external GameData lookup tables (scattered across obf classes) ──
   // FIELD DRIFT: these obf static int[] tables were renamed (generically) in their
   // owning deob classes; each wrapper now references the owning class's DECLARED
   // deob field (obf letter kept in the trailing comment). The physical field is the
   // same; the deob name is the source of truth.

   /**
    * obf: {@code k(lb scene, ua surface)} — mirrors oracle {@code World(Scene,Surface)}.
    * Builds the four blended terrain colour ramps (each 64 entries).
    *
    * <p>CHANNEL-ORDER FIX (RENDER_DEOB_GAPS world-terrain ramp + OBF_MEMBER_MAP
    * ClientStream.rgb row): the obf packer is {@code da.a(int b, byte -66, int r,
    * int g)} whose body (clean da.java:288) is
    * {@code -1 - b/8 - (r/8)*1024 - (g/8)*32} — a {@code (B,R,G)}-order packer
    * (blue is the low channel, red the ×1024 channel). The earlier Javadoc claimed
    * {@code da.a(R,(byte)-66,B,G)} which was inverted. The four call sites below
    * pass {@code ClientStream.rgb(b, r, g)} in that verified {@code (B,R,G)} order,
    * byte-exact to clean k.java:4244/4261/4278/4294.</p>
    */
   public World(Scene scene, Surface surface) {
      this.memberWorld = false;
      this.terrainHeightLocal = new int[regionWidth][regionHeight];
      this.wallsRoof = new byte[4][2304];      // obf: A
      this.wallsEastwest = new byte[4][2304];  // obf: P
      this.tileDecoration = new byte[4][2304]; // obf: R
      this.terrainColours = new int[256];      // obf: w
      this.routeVia = new int[regionWidth][regionHeight];
      this.wallsNorthsouth = new byte[4][2304];// obf: f
      this.localY = new int[18432];            // obf: E
      this.worldInitialised = true;
      this.baseMediaSprite = 750;
      this.wallModels = new GameModel[4][64];
      this.terrainHeight = new byte[4][2304];  // obf: L
      this.terrainModels = new GameModel[64];
      this.wallsDiagonal = new int[4][2304];   // obf: s
      this.tileDirection = new byte[4][2304];  // obf: mb
      this.playerAlive = false;
      this.objectAdjacency = new int[regionWidth][regionHeight];
      this.roofModels = new GameModel[4][64];
      this.terrainColour = new byte[4][2304];  // obf: eb
      this.localX = new int[18432];            // obf: q

      this.surface = surface;
      this.scene = scene;

      for (int i = 0; i < 64; i++)
         terrainColours[i] = ClientStream.rgb(255 - i * 4, 255 - i * 4, 255 - (int) (i * 1.75));
      for (int i = 0; i < 64; i++)
         terrainColours[64 + i] = ClientStream.rgb(0, 3 * i, 144);
      for (int i = 0; i < 64; i++)
         terrainColours[128 + i] = ClientStream.rgb(0, 192 - (int) (i * 1.5), 144 - (int) (i * 1.5));
      for (int i = 0; i < 64; i++)
         terrainColours[192 + i] = ClientStream.rgb(0, 96 - (int) (i * 1.5), (int) (i * 1.5) + 48);
   }

   // ═════════════════════════ grid accessors ═══════════════════════════
   // Each getter resolves the (x,y)→quadrant mapping identically; the local
   // index expression is reproduced EXACTLY as in the clean base.

   /** obf: g(int magic=2, int y, int x) — getTerrainHeight; L[q][48*x + y]·3. */
   private int getTerrainHeight(int x, int y) {
      if (x < 0 || x >= regionWidth || y < 0 || y >= regionHeight) return 0;
      byte q = 0;
      if (x >= 48 && y < 48) { q = 1; x -= 48; }
      else if (x < 48 && y >= 48) { q = 2; y -= 48; }
      else if (x >= 48 && y >= 48) { q = 3; x -= 48; y -= 48; }
      return (terrainHeight[q][48 * x + y] & 0xFF) * 3;
   }

   /** obf: a(byte magic=104, int x, int y) — getTerrainColour; eb[q][y + 48*x]. */
   private int getTerrainColour(int x, int y) {
      if (x < 0 || x >= regionWidth || y < 0 || y >= regionHeight) return 0;
      byte q = 0;
      if (x >= 48 && y < 48) { q = 1; x -= 48; }
      else if (x < 48 && y >= 48) { q = 2; y -= 48; }
      else if (x >= 48 && y >= 48) { q = 3; x -= 48; y -= 48; }
      return terrainColour[q][y + 48 * x] & 0xFF;
   }

   /** obf: e(int magic, int x, int y) — getWallNorthsouth; f[q][x*48 + y]. */
   private int getWallNorthsouth(int x, int y) {
      if (x < 0 || x >= regionWidth || y < 0 || y >= regionHeight) return 0;
      byte q = 0;
      if (x >= 48 && y < 48) { q = 1; x -= 48; }
      else if (x < 48 && y >= 48) { q = 2; y -= 48; }
      else if (x >= 48 && y >= 48) { q = 3; x -= 48; y -= 48; }
      return 0xFF & wallsNorthsouth[q][x * 48 + y];
   }

   /** obf: a(int x, byte magic, int y) — getWallEastwest; P[q][x*48 + y]. */
   private int getWallEastwest(int x, int y) {
      if (x < 0 || x >= regionWidth || y < 0 || y >= regionHeight) return 0;
      byte q = 0;
      if (x >= 48 && y < 48) { q = 1; x -= 48; }
      else if (x < 48 && y >= 48) { q = 2; y -= 48; }
      else if (x >= 48 && y >= 48) { q = 3; x -= 48; y -= 48; }
      return 0xFF & wallsEastwest[q][x * 48 + y];
   }

   /** obf: c(int x, int y, int magic) — getWallDiagonal; s[q][x*48 + y]. */
   private int getWallDiagonal(int x, int y) {
      if (x < 0 || x >= regionWidth || y < 0 || y >= regionHeight) return 0;
      byte q = 0;
      if (x >= 48 && y < 48) { q = 1; x -= 48; }
      else if (x < 48 && y >= 48) { q = 2; y -= 48; }
      else if (x >= 48 && y >= 48) { q = 3; x -= 48; y -= 48; }
      return wallsDiagonal[q][x * 48 + y];
   }

   /**
    * obf: d(int x, int y, int magic) — oracle getWallRoof; reads A.
    * NOTE: the clean base resolves the quadrant with x/y roles SWAPPED relative
    * to the other accessors (q=1 when y>=48&&x<48), and indexes A[q][x + y*48].
    */
   private int getWallRoof(int x, int y) {
      if (x < 0 || x >= regionWidth || y < 0 || y >= regionHeight) return 0;
      byte q = 0;
      if (y >= 48 && x < 48) { q = 1; y -= 48; }
      else if (y < 48 && x >= 48) { q = 2; x -= 48; }
      else if (y >= 48 && x >= 48) { q = 3; x -= 48; y -= 48; }
      return wallsRoof[q][x + y * 48];
   }

   /** obf: b(int x, int y, int magic) — oracle getTileDirection; reads mb; mb[q][y + x*48]. */
   public int getTileDirection(int x, int y) {
      if (x < 0 || x >= regionWidth || y < 0 || y >= regionHeight) return 0;
      byte q = 0;
      if (x >= 48 && y < 48) { q = 1; x -= 48; }
      else if (x < 48 && y >= 48) { q = 2; y -= 48; }
      else if (x >= 48 && y >= 48) { q = 3; x -= 48; y -= 48; }
      return tileDirection[q][y + x * 48];
   }

   /**
    * obf: b(int magic, int x, int 4, int y) — oracle getTileDecoration(x,y,unused);
    * reads R; gated by the dummy plane==4 anti-tamper; index R[q][48*x + y].
    */
   private int getTileDecoration(int x, int y) {
      if (x < 0 || x >= regionWidth || y < 0 || y >= regionHeight) return 0;
      byte q = 0;
      if (x >= 48 && y < 48) { q = 1; x -= 48; }
      else if (x < 48 && y >= 48) { q = 2; y -= 48; }
      else if (x >= 48 && y >= 48) { q = 3; x -= 48; y -= 48; }
      return 0xFF & tileDecoration[q][48 * x + y];
   }

   /**
    * obf: d(int magic=-8509, int x, int def, int plane, int y) — oracle
    * getTileDecoration(x,y,unused,def): decoration colour, or {@code def} when none.
    */
   private int getTileDecorationOr(int x, int y, int def) {
      int deco = getTileDecoration(x, y);
      if (deco == 0) return def;
      return Panel.texK[deco - 1];           // qa.K — GameData.tileDecoration
   }

   /**
    * obf: d(int plane, int x, int magic=15282, int y) — oracle getTileType:
    * 1 if the decoration type is 2 (water), 0 if some other type, -1 if none.
    */
   private int getTileType(int x, int y) {
      int deco = getTileDecoration(x, y);
      if (deco == 0) return -1;
      return ClientStream.sharedIntArrayN[deco - 1] != 2 ? 0 : 1;     // da.N — GameData.tileType
   }

   /** obf: b(byte magic, int x, int y) — getObjectAdjacency; bb[y][x] (transposed), 0 if OOB. */
   private int getObjectAdjacency(int x, int y) {
      if (x >= 0 && x < regionWidth && y >= 0 && y < regionHeight)
         return objectAdjacency[y][x];
      return 0;
   }

   // ═══════════════════ adjacency / passability mutators ════════════════

   /** obf: a(int bit, int x, byte magic, int y) — objectAdjacency[y][x] |= bit (d.a == OR). */
   private void orObjectAdjacency(int bit, int x, int y) {
      objectAdjacency[y][x] = CacheFile.or(objectAdjacency[y][x], bit);
   }

   /** obf: c(int x, int mask, int y, int delta) — objectAdjacency[y][x] &= (mask-delta) (ib.a == AND). */
   private void andObjectAdjacency(int x, int mask, int y, int delta) {
      objectAdjacency[y][x] = StreamBase.bitwiseAnd(objectAdjacency[y][x], mask - delta);
   }

   /**
    * obf: a(int x, int objectId, int dir, int y, int adj) — oracle
    * World.setObjectAdjacency(x,y,dir,id): SET the directional wall-object
    * passability bit, propagating to the neighbour tile for dirs 0/1.
    * (objectAdjacency[y][x]; dir2 ⇒ 0x10, dir3 ⇒ 0x20.)
    */
   public final void setWallObjectAdjacency(int x, int objectId, int dir, int y, int adj) {
      if (x < 0 || y < 0 || x >= 95 || y >= 95) return;
      if (StringCodec.DEAD_INT_ARRAY[objectId] != 1) return;     // u.a — GameData.wallObjectAdjacent
      if (dir == 0) {
         objectAdjacency[y][x] = CacheFile.or(objectAdjacency[y][x], 1);
         if (x > 0) orObjectAdjacency(4, x - 1, y);     // [y][x-1] |= 4
      } else if (dir == 1) {
         objectAdjacency[y][x] = CacheFile.or(objectAdjacency[y][x], 2);
         if (y > 0) orObjectAdjacency(8, x, y - 1);     // [y-1][x] |= 8
      } else if (dir == 3) {
         objectAdjacency[y][x] = CacheFile.or(objectAdjacency[y][x], 32); // 0x20
      } else if (dir == 2) {
         objectAdjacency[y][x] = CacheFile.or(objectAdjacency[y][x], 16); // 0x10
      }
      method404(1, 1, x, y);                            // obf c(1,1,62,x,y)
   }

   /**
    * obf: a(boolean unused, int dir, int x, int y, int objectId) — oracle
    * World.setObjectAdjacency variant that AND-clears wall bits (used while
    * tearing down a placed wall object). objectType (u.a) must be 1.
    */
   public final void clearWallObjectAdjacency(boolean unused, int dir, int x, int y, int objectId) {
      if (x < 0 || y < 0 || x >= 95 || y >= 95) return;
      if (StringCodec.DEAD_INT_ARRAY[objectId] != 1) return;
      if (dir == 0) {
         objectAdjacency[y][x] = StreamBase.bitwiseAnd(objectAdjacency[y][x], 0xFFFE);  // &= ~1
         if (x > 0) andObjectAdjacency(x - 1, 0xFFFF, y, 4);                   // [y][x-1] &= ~4
      } else if (dir == 1) {
         objectAdjacency[y][x] = StreamBase.bitwiseAnd(objectAdjacency[y][x], 0xFFFD);  // &= ~2
         if (y > 0) andObjectAdjacency(x, 0xFFFF, y - 1, 8);                   // [y-1][x] &= ~8
      } else if (dir == 2) {
         objectAdjacency[y][x] = StreamBase.bitwiseAnd(objectAdjacency[y][x], 0xFFEF);  // &= ~16
      } else if (dir == 3) {
         objectAdjacency[y][x] = StreamBase.bitwiseAnd(objectAdjacency[y][x], 0xFFDF);  // &= ~32
      }
      method404(1, 1, x, y);                            // obf c(1,1,-59,x,y)
   }

   // ═══════════════════ roof helper predicates ═════════════════════════

   /**
    * obf: a(int var1, int magic=26431, int var3) — oracle method427: true if any of
    * the four tiles meeting at corner (x,y) carries a roof.
    *
    * <p>CORRECTNESS FIX (class b/c — un-swapped getWallRoof args): the clean body
    * (k.java:2655-2658) is {@code ~d(var3, var1, ..) < -1 || ~d(var3, var1-1, ..) <
    * -1 || 0 < d(var3-1, var1-1, ..) || -1 > ~d(var3-1, var1, ..)}, i.e. getWallRoof
    * is called {@code d(x=var3, y=var1)} — the THIRD positional arg is the
    * getWallRoof x-arg and the FIRST is the y-arg. With deob (x=var1, y=var3) the
    * getWallRoof calls must therefore be {@code getWallRoof(y, x)}, NOT
    * {@code getWallRoof(x, y)}. (getWallRoof itself is NOT symmetric — it indexes
    * {@code A[q][x + y*48]} — so the earlier un-swapped form queried the wrong
    * tiles for the roof-corner taper.) Cross-checked against the deliberately
    * swapped existence query in buildRoofs ({@code getWallRoof(ly, lx)}).</p>
    */
   private boolean isRoofCorner(int x, int y) {
      return getWallRoof(y, x) > 0
            || getWallRoof(y, x - 1) > 0
            || getWallRoof(y - 1, x - 1) > 0
            || getWallRoof(y - 1, x) > 0;
   }

   /**
    * obf: a(boolean var1=seed, int var2, int var3) — oracle hasRoof: returns true
    * only if ALL four tiles meeting at corner (x,y) are roofed, else {@code seed}.
    *
    * <p>CORRECTNESS FIX (class b/c — un-swapped getWallRoof args): the clean body
    * (k.java:4161-4164) calls {@code d(var3, var2, ..)} = getWallRoof(x=var3,
    * y=var2). With deob (x=var2, y=var3) the calls must be {@code getWallRoof(y, x)}.
    * Same swap as {@link #isRoofCorner}.</p>
    */
   private boolean hasRoof(boolean seed, int x, int y) {
      if (getWallRoof(y, x) > 0
            && getWallRoof(y, x - 1) > 0
            && getWallRoof(y - 1, x - 1) > 0
            && getWallRoof(y - 1, x) > 0)
         return true;
      return seed;
   }

   // ════════════════════════ ground elevation ══════════════════════════

   /**
    * obf: f(int wx, int wy, int magic) — oracle World.getElevation: bilinear
    * interpolation of the four tile-corner heights; the low 7 bits are the
    * fractional offset within the tile, split across the two triangles.
    */
   public final int getElevation(int wx, int wy) {
      int sx = wx >> 7;
      int sy = wy >> 7;
      int aX = wx & 0x7F;
      int aY = wy & 0x7F;
      if (sx < 0 || sy < 0 || sx >= 95 || sy >= 95) return 0;
      int h, hx, hy;
      if (aX <= anInt585 - aY) {
         h = getTerrainHeight(sx, sy);
         hx = getTerrainHeight(sx + 1, sy) - h;
         hy = getTerrainHeight(sx, sy + 1) - h;
      } else {
         h = getTerrainHeight(sx + 1, sy + 1);
         hx = getTerrainHeight(sx, sy + 1) - h;
         hy = getTerrainHeight(sx + 1, sy) - h;
         aX = anInt585 - aX;
         aY = anInt585 - aY;
      }
      return h + (hx * aX) / anInt585 + (hy * aY) / anInt585;
   }

   // ══════════════════════════ pathfinding ═════════════════════════════

   /**
    * obf: a(int[] routeX, int endX2, byte magic, int endY2, int[] routeY,
    *        int startX, int startY, int endX1, int endY1, boolean allowObjects)
    * — oracle World.route. Breadth-first search over the 96×96 passability grid;
    * fills routeX/routeY with the waypoint path from (startX,startY) to anywhere
    * inside [endX1..endX2]×[endY1..endY2], honouring object adjacency when
    * {@code allowObjects}. Returns the waypoint count, or -1 if unreachable.
    */
   public final int route(int[] routeX, int endX2, int endY2, int[] routeY,
                   int startX, int startY, int endX1, int endY1, boolean allowObjects) {
      for (int gx = 0; gx < regionWidth; gx++)
         for (int gy = 0; gy < regionHeight; gy++)
            routeVia[gx][gy] = 0;

      int writePtr = 0;
      int readPtr = 0;
      int x = startX;
      int y = startY;
      routeVia[startX][startY] = 99;
      routeX[writePtr] = startX;
      routeY[writePtr++] = startY;
      int size = routeX.length;
      boolean reached = false;

      while (readPtr != writePtr) {
         x = routeX[readPtr];
         y = routeY[readPtr];
         readPtr = (readPtr + 1) % size;

         if (x >= endX1 && x <= endX2 && y >= endY1 && y <= endY2) { reached = true; break; }

         if (allowObjects) {
            if (x > 0 && x - 1 >= endX1 && x - 1 <= endX2 && y >= endY1 && y <= endY2
                  && (objectAdjacency[x - 1][y] & 8) == 0) { reached = true; break; }
            if (x < 95 && x + 1 >= endX1 && x + 1 <= endX2 && y >= endY1 && y <= endY2
                  && (objectAdjacency[x + 1][y] & 2) == 0) { reached = true; break; }
            if (y > 0 && x >= endX1 && x <= endX2 && y - 1 >= endY1 && y - 1 <= endY2
                  && (objectAdjacency[x][y - 1] & 4) == 0) { reached = true; break; }
            if (y < 95 && x >= endX1 && x <= endX2 && y + 1 >= endY1 && y + 1 <= endY2
                  && (objectAdjacency[x][y + 1] & 1) == 0) { reached = true; break; }
         }

         if (x > 0 && routeVia[x - 1][y] == 0 && (objectAdjacency[x - 1][y] & 0x78) == 0) {
            routeX[writePtr] = x - 1; routeY[writePtr] = y; writePtr = (writePtr + 1) % size;
            routeVia[x - 1][y] = 2;
         }
         if (x < 95 && routeVia[x + 1][y] == 0 && (objectAdjacency[x + 1][y] & 0x72) == 0) {
            routeX[writePtr] = x + 1; routeY[writePtr] = y; writePtr = (writePtr + 1) % size;
            routeVia[x + 1][y] = 8;
         }
         if (y > 0 && routeVia[x][y - 1] == 0 && (objectAdjacency[x][y - 1] & 0x74) == 0) {
            routeX[writePtr] = x; routeY[writePtr] = y - 1; writePtr = (writePtr + 1) % size;
            routeVia[x][y - 1] = 1;
         }
         if (y < 95 && routeVia[x][y + 1] == 0 && (objectAdjacency[x][y + 1] & 0x71) == 0) {
            routeX[writePtr] = x; routeY[writePtr] = y + 1; writePtr = (writePtr + 1) % size;
            routeVia[x][y + 1] = 4;
         }
         if (x > 0 && y > 0 && (objectAdjacency[x][y - 1] & 0x74) == 0
               && (objectAdjacency[x - 1][y] & 0x78) == 0 && (objectAdjacency[x - 1][y - 1] & 0x7C) == 0
               && routeVia[x - 1][y - 1] == 0) {
            routeX[writePtr] = x - 1; routeY[writePtr] = y - 1; writePtr = (writePtr + 1) % size;
            routeVia[x - 1][y - 1] = 3;
         }
         if (x < 95 && y > 0 && (objectAdjacency[x][y - 1] & 0x74) == 0
               && (objectAdjacency[x + 1][y] & 0x72) == 0 && (objectAdjacency[x + 1][y - 1] & 0x76) == 0
               && routeVia[x + 1][y - 1] == 0) {
            routeX[writePtr] = x + 1; routeY[writePtr] = y - 1; writePtr = (writePtr + 1) % size;
            routeVia[x + 1][y - 1] = 9;
         }
         if (x > 0 && y < 95 && (objectAdjacency[x][y + 1] & 0x71) == 0
               && (objectAdjacency[x - 1][y] & 0x78) == 0 && (objectAdjacency[x - 1][y + 1] & 0x79) == 0
               && routeVia[x - 1][y + 1] == 0) {
            routeX[writePtr] = x - 1; routeY[writePtr] = y + 1; writePtr = (writePtr + 1) % size;
            routeVia[x - 1][y + 1] = 6;
         }
         if (x < 95 && y < 95 && (objectAdjacency[x][y + 1] & 0x71) == 0
               && (objectAdjacency[x + 1][y] & 0x72) == 0 && (objectAdjacency[x + 1][y + 1] & 0x73) == 0
               && routeVia[x + 1][y + 1] == 0) {
            routeX[writePtr] = x + 1; routeY[writePtr] = y + 1; writePtr = (writePtr + 1) % size;
            routeVia[x + 1][y + 1] = 12;
         }
      }

      if (!reached) return -1;

      readPtr = 0;
      routeX[readPtr] = x;
      routeY[readPtr++] = y;
      int stride;
      for (int step = stride = routeVia[x][y]; x != startX || y != startY; step = routeVia[x][y]) {
         if (step != stride) {
            stride = step;
            routeX[readPtr] = x;
            routeY[readPtr++] = y;
         }
         if ((step & 2) != 0) x++;
         else if ((step & 8) != 0) x--;
         if ((step & 1) != 0) y++;
         else if ((step & 4) != 0) y--;
      }
      return readPtr;
   }

   // ═══════════════ object footprint → passability marking ══════════════

   /**
    * obf: a(int objectId, int x, int y, int magic=4081) — oracle removeObject:
    * clear the directional passability bits of a multi-tile object placed at
    * (x,y) over its (objectWidth×objectHeight, rotated by tileDirection)
    * footprint, propagating the mirror bit to neighbours, then re-tile.
    */
   public final void removeObject(int objectId, int x, int y, int magic) {
      if (magic != 4081) method425(-98, 25, -8); // dead anti-tamper call
      if (x < 0 || y < 0 || x >= 95 || y >= 95) return;
      if (Utility.sizedPoolCounts[objectId] != 1 && Utility.sizedPoolCounts[objectId] != 2) return;
      int dir = getTileDirection(x, y);
      int w, hgt;
      if (dir == 0 || dir == 4) { hgt = NameTable.sortKeys[objectId]; w = RecordLoader.intArray[objectId]; }
      else { hgt = RecordLoader.intArray[objectId]; w = NameTable.sortKeys[objectId]; }
      for (int gx = x; gx < x + w; gx++) {
         for (int gy = y; gy < y + hgt; gy++) {
            if (Utility.sizedPoolCounts[objectId] == 1) {
               objectAdjacency[gx][gy] = StreamBase.bitwiseAnd(objectAdjacency[gx][gy], 0xFFBF); // &= ~0x40
            } else if (dir == 0) {
               objectAdjacency[gx][gy] = StreamBase.bitwiseAnd(objectAdjacency[gx][gy], 0xFFFD); // &= ~2
               if (gx > 0) andObjectAdjacency(gy, magic + 61454, gx - 1, 8);
            } else if (dir == 4) {
               objectAdjacency[gx][gy] = StreamBase.bitwiseAnd(objectAdjacency[gx][gy], 0xFFF7); // &= ~8
               if (gx < 95) andObjectAdjacency(gy, magic + 61454, gx + 1, 2);
            } else if (dir == 6) {
               objectAdjacency[gx][gy] = StreamBase.bitwiseAnd(objectAdjacency[gx][gy], 0xFFFE); // &= ~1
               if (gy > 0) andObjectAdjacency(gy - 1, magic + 61454, gx, 4);
            } else if (dir == 2) {
               objectAdjacency[gx][gy] = StreamBase.bitwiseAnd(objectAdjacency[gx][gy], 0xFFFB); // &= ~4
               if (gy < 95) andObjectAdjacency(gy + 1, magic + 61454, gx, 1);
            }
         }
      }
      method404(w, hgt, x, y);
   }

   /**
    * obf: a(int x, int objectId, boolean unused, int y) — oracle removeObject2:
    * SET the directional passability bits of an object placed at (x,y) over its
    * footprint and propagate the mirror bit to neighbours, then re-tile.
    * objectType==1 ⇒ block-all (0x40); objectType==2 ⇒ directional by tileDir.
    */
   public final void removeObject2(int x, int objectId, boolean unused, int y) {
      if (unused) return;
      if (x < 0 || y < 0 || x >= 95 || y >= 95) return;
      if (Utility.sizedPoolCounts[objectId] != 1 && Utility.sizedPoolCounts[objectId] != 2) return;
      int dir = getTileDirection(x, y);                  // obf b(x,y,-107)
      int w, hgt;
      if (dir == 0 || dir == 4) { w = RecordLoader.intArray[objectId]; hgt = NameTable.sortKeys[objectId]; }
      else { w = NameTable.sortKeys[objectId]; hgt = RecordLoader.intArray[objectId]; }
      for (int gx = x; gx < x + w; gx++) {
         for (int gy = y; gy < y + hgt; gy++) {
            if (Utility.sizedPoolCounts[objectId] == 1) {
               objectAdjacency[gx][gy] = CacheFile.or(objectAdjacency[gx][gy], 64);  // |= 0x40
            } else if (dir == 2) {
               objectAdjacency[gx][gy] = CacheFile.or(objectAdjacency[gx][gy], 4);   // |= 4
               if (gy < 95) orObjectAdjacency(1, gy + 1, gx);                       // [gx][gy+1] |= 1
            } else if (dir == 6) {
               objectAdjacency[gx][gy] = CacheFile.or(objectAdjacency[gx][gy], 1);   // |= 1
               if (gy > 0) orObjectAdjacency(4, gy - 1, gx);                        // [gx][gy-1] |= 4
            } else if (dir == 4) {
               objectAdjacency[gx][gy] = CacheFile.or(objectAdjacency[gx][gy], 8);   // |= 8
               if (gx < 95) orObjectAdjacency(2, gy, gx + 1);                       // [gx+1][gy] |= 2
            } else { // dir == 0
               objectAdjacency[gx][gy] = CacheFile.or(objectAdjacency[gx][gy], 2);   // |= 2
               if (gx > 0) orObjectAdjacency(8, gy, gx - 1);                        // [gx-1][gy] |= 8
            }
         }
      }
      method404(w, hgt, x, y);
   }

   /**
    * obf: a(int z2, int amb, int cellY, int magic, int cellX, int x2) — oracle
    * World.setTerrainAmbience: find the vertex of terrain model
    * {@code terrainModels[cellX + cellY*8]} whose world position is
    * (x2*128, z2*128) and set its lighting ambience to {@code amb}.
    * (obf model index = var5 + var3*8 = cellX + cellY*8; vertex match is
    * {@code vertexX==x2*128 && vertexZ==z2*128}.)
    */
   public final void setTerrainAmbience(int z2, int amb, int cellY, int magic, int cellX, int x2) {
      GameModel model = terrainModels[cellX + cellY * 8];
      for (int v = 0; v < model.numVertices; v++) {     // ca.Db -> numVertices
         if (model.vertexX[v] == 128 * x2 && model.vertexZ[v] == z2 * 128) { // ca.a -> vertexX, ca.bc -> vertexZ
            model.setVertexAmbience(v, amb, (byte) -61); // ca.a(IIB) -> setVertexAmbience
            return;
         }
      }
   }

   // ════════════════ minimap / ambience region refresh ═════════════════

   /**
    * obf: b(int k, byte magic, int i, int j) — oracle method425: split the (i,j)
    * point into its enclosing 12-tile cells and set each corner's terrain-model
    * ambience to {@code k} via {@link #setTerrainAmbience}.
    */
   private void method425(int i, int j, int k) {
      int cellXHi = i / 12;
      int cellYHi = j / 12;
      int cellXLo = (i - 1) / 12;
      int cellYLo = (j - 1) / 12;
      setTerrainAmbience(j, k, cellYHi, 2, cellXHi, i);
      if (cellXLo != cellXHi) setTerrainAmbience(j, k, cellYHi, 2, cellXLo, i);
      if (cellYLo != cellYHi) setTerrainAmbience(j, k, cellYLo, 2, cellXHi, i);
      if (cellXLo != cellXHi && cellYLo != cellYHi) setTerrainAmbience(j, k, cellYLo, 2, cellXLo, i);
   }

   /**
    * obf: c(int w, int h, int magic, int x, int y) — oracle method404: refresh the
    * minimap "is this tile blocked?" colour over the [x..x+w]×[y..y+h] region,
    * a tile reading blocked if it or any lower-left neighbour carries the bits.
    */
   private void method404(int w, int h, int x, int y) {
      if (x < 1 || y < 1 || x + w >= 96 || y + h >= 96) return;
      // obf reads getObjectAdjacency(gy,gx) etc — i.e. objectAdjacency[gx][gy].
      for (int gx = x; gx <= x + w; gx++) {
         for (int gy = y; gy <= y + h; gy++) {
            if ((getObjectAdjacency(gy, gx) & 0x63) != 0          // objectAdjacency[gx][gy]
                  || (getObjectAdjacency(gy, gx - 1) & 0x59) != 0  // [gx-1][gy]
                  || (getObjectAdjacency(gy - 1, gx) & 0x56) != 0  // [gx][gy-1]
                  || (getObjectAdjacency(gy - 1, gx - 1) & 0x6C) != 0) // [gx-1][gy-1]
               method425(gx, gy, 35);
            else
               method425(gx, gy, 0);
         }
      }
   }

   // ═══════════════════════ section (re)set / tiling ═══════════════════

   /** obf: b(int magic=-10185) — oracle World.reset: discard models, free heap, gc. */
   private void reset(int magic) {
      if (magic != -10185) return;             // anti-tamper guard
      if (worldInitialised) scene.clearModels(); // lb.a(boolean) -> clearModels (tear down model list)
      for (int n = 0; n < 64; n++) {
         terrainModels[n] = null;
         for (int plane = 0; plane < 4; plane++) wallModels[plane][n] = null;
         for (int plane = 0; plane < 4; plane++) roofModels[plane][n] = null;
      }
      System.gc();
   }

   /**
    * obf: a(int magic) — oracle World.setTiles: resolve ambiguous water/edge
    * decoration tiles (id 250) at the 48-tile quadrant seams into wall (9) or
    * floor (2) decoration.
    */
   private void setTiles(int magic) {
      for (int x = 0; x < regionWidth; x++) {
         for (int y = 0; y < regionHeight; y++) {
            if (getTileDecoration(x, y) != 250) continue;
            if (x == 47 && getTileDecoration(x + 1, y) != 250 && getTileDecoration(x + 1, y) != 2)
               setTileDecoration(9, x, y);
            else if (y == 47 && getTileDecoration(x, y + 1) != 250 && getTileDecoration(x, y + 1) != 2)
               setTileDecoration(9, x, y);
            else
               setTileDecoration(2, x, y);
         }
      }
   }

   /**
    * obf: e(int value, int x, int magic, int y) — oracle World.setTileDecoration:
    * write a decoration id into quadrant grid R; index R[q][y + 48*x].
    */
   private void setTileDecoration(int value, int x, int y) {
      if (x < 0 || x >= regionWidth || y < 0 || y >= regionHeight) return;
      byte q = 0;
      if (x >= 48 && y < 48) { q = 1; x -= 48; }
      else if (x < 48 && y >= 48) { q = 2; y -= 48; }
      else if (x >= 48 && y >= 48) { q = 3; x -= 48; y -= 48; }
      tileDecoration[q][y + 48 * x] = (byte) value;
   }

   // ═══════════════════════ landscape blob decode ══════════════════════

   /**
    * obf: static a(int magic=128, boolean verbose, byte[] data) — decode one
    * length-prefixed landscape blob: 6-byte header = two big-endian 24-bit
    * lengths (raw, compressed); equal ⇒ verbatim payload, else bzip2-decompress.
    */
   public static byte[] unpackData(int magic, boolean verbose, byte[] data) {
      int rawLen = ((data[0] << 16) & 0xFF0000) + ((data[1] & 0xFF) << 8) + (data[2] & 0xFF);
      int compLen = ((data[3] << 16) & 0xFF0000) + ((data[4] & 0xFF) << 8) + (data[5] & 0xFF);
      if (rawLen == compLen) {
         byte[] out = new byte[data.length - 6];
         ArrayUtil.copy(data, 6, out, 0, out.length);  // ab.a -> copy (System.arraycopy)
         return out;
      }
      if (verbose) ClientStream.reportProgress(STR_UNPACKING, 0, 0); // da.a(String,int,int) -> reportProgress
      byte[] out = new byte[rawLen];
      // WRONG-RECEIVER + DRIFT: clean k.java:528 is `ea.a(...)` where ea=SpriteDecoder
      // (NOT aa=BZip), and the deob renamed ea.a -> SpriteDecoder.decompress.
      SpriteDecoder.decompress(out, rawLen, data, compLen, 6);
      return out;
   }

   // ════════════════════ terrain / wall mesh helpers ══════════════════

   /**
    * obf: a(int objectId, int bx, int ay, int by, byte magic, int ax) — oracle
    * method428: raise the two terrain vertices bordering a wall by the wall's
    * height, flagging them with the +80000 "already raised" sentinel. The two
    * cells raised are terrainHeightLocal[ax][ay] and terrainHeightLocal[bx][by]
    * (obf ab[var6][var3] and ab[var2][var4]).
    */
   private void method428(int objectId, int bx, int ay, int by, int ax) {
      int height = StreamBase._deadIntArray_d[objectId];           // ib.d — wall/door height
      if (terrainHeightLocal[ax][ay] < 80000)
         terrainHeightLocal[ax][ay] += height + 80000;
      if (terrainHeightLocal[bx][by] < 80000)
         terrainHeightLocal[bx][by] += height + 80000;
   }

   /**
    * obf: a(int objectId, ca model, int ax, int ay, int bx, int magic, int by) —
    * oracle method422: build the four-vertex quad for one standing wall segment
    * from endpoint (bx,ay) to endpoint (ax,by) and append it to {@code model},
    * tagging faces with the wall colour/texture. (obf vars: var3=ax, var4=ay,
    * var5=bx, var7=by; endpoint A is (bx,ay), endpoint B is (ax,by).)
    */
   private void method422(int objectId, GameModel model, int ax, int ay, int bx, int magic, int by) {
      method425(bx, ay, 40);   // obf b(40,(byte)50,var5=bx,var4=ay)
      method425(ax, by, 40);   // obf b(40,(byte)109,var3=ax,var7=by)
      int height = StreamBase._deadIntArray_d[objectId];           // ib.d — wall height
      int frontColour = ChatCipher.scratchA[objectId];  // v.a
      int backColour = client.Mudclient.Jk[objectId];     // client.Jk
      int axw = 128 * bx, ayw = 128 * ay;               // endpoint A world coords
      int bxw = 128 * ax, byw = 128 * by;               // endpoint B world coords
      int v0 = model.vertexAt(axw, ayw, -terrainHeightLocal[bx][ay], -111);         // ca.e -> vertexAt
      int v1 = model.vertexAt(axw, ayw, -terrainHeightLocal[bx][ay] - height, -115);
      int v2 = model.vertexAt(bxw, byw, -height - terrainHeightLocal[ax][by], -125);
      int v3 = model.vertexAt(bxw, byw, -terrainHeightLocal[ax][by], magic);
      int faceId = model.createFace(4, new int[]{v0, v1, v2, v3}, frontColour, backColour, false); // ca.a(I[IIIZ) -> createFace
      if (Scene.diagScratch[objectId] == 5) model.faceTag[faceId] = 30000 + objectId; // ca.E -> faceTag; lb.Tb
      else model.faceTag[faceId] = 0;
   }

   /**
    * obf: a(int diag, byte magic, int colour2, int x, int y, int colour1) —
    * oracle method402: paint a tile's two colour triangles into the 3×3 minimap
    * sprite block (each colour dimmed to 50%).
    */
   private void method402(int diag, int colour2, int x, int y, int colour1) {
      int px = x * 3;
      int py = y * 3;
      // WRONG-RECEIVER FIX (OBF_MEMBER_MAP rows 717/779; clean k.java:3946/3949 are
      // `this.c.a(...,true)` where c=lb=Scene, NOT this.U=ua=Surface): the colour
      // fill is Scene.fillColour, not a (nonexistent) Surface.a(int,boolean).
      int c1 = scene.fillColour(colour1, true) >> 1 & 0x7F7F7F;
      int c2 = (scene.fillColour(colour2, true) & 0xFEFEFF) >> 1;
      // The minimap triangle strokes stay on Surface (this.U); ua.b(IIIIB) -> drawLineHoriz.
      if (diag == 0) {
         surface.drawLineHoriz(3, c1, px, py, (byte) 109);
         surface.drawLineHoriz(2, c1, px, py + 1, (byte) -65);
         surface.drawLineHoriz(1, c1, px, py + 2, (byte) 99);
         surface.drawLineHoriz(1, c2, px + 2, py + 1, (byte) 73);
         surface.drawLineHoriz(2, c2, px + 1, py + 2, (byte) 113);
      } else if (diag == 1) {
         surface.drawLineHoriz(3, c2, px, py, (byte) 55);
         surface.drawLineHoriz(2, c2, px + 1, py + 1, (byte) 62);
         surface.drawLineHoriz(1, c2, px + 2, py + 2, (byte) 56);
         surface.drawLineHoriz(1, c1, px, py + 1, (byte) 70);
         surface.drawLineHoriz(2, c1, px, py + 2, (byte) -85);
      }
   }

   // ════════════════ diagonal-wall / door object models ════════════════

   /**
    * obf: a(ca[] prototypes, byte magic) — oracle World.addModels: build the
    * free-standing diagonal-wall / door GameModels. For every tile whose diagonal
    * grid holds a "door" id (48001..59999), clone the prototype model, position it
    * at the tile centre at the interpolated terrain height, orient by
    * tileDirection, register it with the scene, and clear the diagonal id over its
    * footprint so it renders once.
    */
   public final void addModels(GameModel[] prototypes, byte magic) {
      for (int x = 0; x < 94; x++) {              // obf var3 (outer)
         for (int y = 0; y < 94; y++) {           // obf var4 (inner)
            int diag = getWallDiagonal(x, y);
            if (diag <= 48000 || diag >= 60000) continue;
            int objectId = diag - 48001;          // obf var5
            int dir = getTileDirection(x, y);     // obf var6 = b(x,y,-91)
            int w, hgt;                            // obf var7 (x-extent), var8 (y-extent)
            if (dir != 0 && dir != 4) { w = NameTable.sortKeys[objectId]; hgt = RecordLoader.intArray[objectId]; }
            else { hgt = NameTable.sortKeys[objectId]; w = RecordLoader.intArray[objectId]; }

            removeObject2(x, objectId, false, y);  // obf a(var3,var5,false,var4)
            GameModel model = prototypes[SurfaceImageProducer.entityIndexTableF[objectId]].copy(false, -120, false, false, true); // ca.a(Z,I,Z,Z,Z) -> copy
            int cx = anInt585 * (w + x + x) / 2;   // obf 128*(var7+2*var3)/2
            int cy = (hgt + y + y) * anInt585 / 2; // obf (var8+2*var4)*128/2
            model.translate(cx, cy, -getElevation(cx, cy), true);       // ca.a(IIIZ) -> translate
            model.orient(0, -999999, 0, getTileDirection(x, y) * 32);   // ca.g(IIII) -> orient
            scene.addModel(model);                                      // lb.a(ca,B) -> addModel (byte dropped)
            model.setLight(48, 48, -10, magic ^ 9, -50, -50);          // ca.a(IIIIII) -> setLight
            if (w <= 1 && hgt <= 1) continue;      // obf if(var7>1 || var8>1)
            for (int gx = x; gx < x + w; gx++) {
               for (int gy = y; gy < y + hgt; gy++) {
                  if ((gx > x || gy > y) && objectId == getWallDiagonal(gx, gy) - 48001)
                     setWallDiagonal(gx, gy, 0);
               }
            }
         }
      }
   }

   /** obf: inline {@code s[q][lx*48+ly] = 0} — clear a diagonal-wall grid cell. */
   private void setWallDiagonal(int x, int y, int value) {
      if (x < 0 || x >= regionWidth || y < 0 || y >= regionHeight) return;
      byte q = 0;
      if (x >= 48 && y < 48) { q = 1; x -= 48; }
      else if (x < 48 && y >= 48) { q = 2; y -= 48; }
      else if (x >= 48 && y >= 48) { q = 3; x -= 48; y -= 48; }
      wallsDiagonal[q][x * 48 + y] = value;
   }

   // ════════════════════════ section loading ═══════════════════════════

   /**
    * obf: a(int x, byte magic=-90, int y, int plane) — oracle
    * World.loadSection(x,y,plane): top-level entry. Resets the previous section,
    * meshes this plane, and (for plane 0) also loads the upper planes 1/2.
    */
   public final void loadSection(int x, byte magic, int y, int plane) {
      reset(magic ^ 0x2791);                 // obf b(magic^10129) → reset(-10185)
      int cellX = (24 + x) / 48;
      buildSection(x, 122, true, plane, y);  // mesh requested plane
      int cellY = (24 + y) / 48;
      if (plane == 0) {
         buildSection(x, 112, false, 1, y);                 // upper plane 1 (collision only)
         buildSection(x, magic ^ 0xFFFFFFE3, false, 2, y);  // upper plane 2 (collision only)
         loadMapData(plane, 0, cellX - 1, 0, cellY - 1);
         loadMapData(plane, 1, cellX, magic + 90, cellY - 1);
         loadMapData(plane, 2, cellX - 1, 0, cellY);
         loadMapData(plane, 3, cellX, magic + 90, cellY);
         setTiles(0);
      }
   }

   /**
    * obf: a(int x, int magic, boolean flag, int plane, int y) — oracle
    * World.loadSection(x,y,plane,flag): the heart of the mesher. Loads the four
    * surrounding chunks of {@code plane}; when {@code magic>=66} builds terrain,
    * wall and roof {@link GameModel}s and registers them with the scene.
    *
    * <p>Faithful to the clean base: per-tile vertex heights (flattened under
    * bridge/water tiles), two gouraud floor triangles (or a quad) per tile with
    * the split-diagonal/overlay colour selection, the 8×8 model split, the four
    * standing-wall orientations, and the roof-ridge pass using the +80000
    * "already-raised" sentinel on {@link #terrainHeightLocal}.</p>
    */
   private void buildSection(int x, int magic, boolean flag, int plane, int y) {
      int cellX = (24 + x) / 48;
      int cellY = (24 + y) / 48;
      loadMapData(plane, 0, cellX - 1, 0, cellY - 1);
      loadMapData(plane, 1, cellX, 0, cellY - 1);
      if (magic < 66) return;                  // collision-only pass stops here
      loadMapData(plane, 2, cellX - 1, 0, cellY);
      loadMapData(plane, 3, cellX, 0, cellY);
      setTiles(0);
      if (parentModel == null)
         parentModel = new GameModel(18688, 18688, true, true, false, false, true);

      if (flag) {
         surface.blackScreen(true);            // DRIFT FIX: obf surface.a(boolean) → Surface.blackScreen(boolean)
         for (int gx = 0; gx < regionWidth; gx++)
            for (int gy = 0; gy < regionHeight; gy++)
               objectAdjacency[gx][gy] = 0;

         GameModel terrain = parentModel;
         terrain.clear(1);                     // ca.c(I) -> clear

         // ── 1. terrain vertices (flatten under bridge/water type-4 tiles) ──
         for (int tx = 0; tx < regionWidth; tx++) {
            for (int ty = 0; ty < regionHeight; ty++) {
               int height = -getTerrainHeight(tx, ty);
               if (getTileTypeOnPlane(tx, ty) == 4) height = 0;
               if (getTileTypeOnPlane(tx - 1, ty) == 4) height = 0;
               if (getTileTypeOnPlane(tx, ty - 1) == 4) height = 0;
               if (getTileTypeOnPlane(tx - 1, ty - 1) == 4) height = 0;
               int vid = terrain.vertexAt(tx * 128, 128 * ty, height, 107); // ca.e -> vertexAt
               int amb = (int) (Math.random() * 10.0) - 5;
               terrain.setVertexAmbience(vid, amb, (byte) -61);             // ca.a(IIB) -> setVertexAmbience
            }
         }

         // ── 2. coloured floor triangles per tile ──
         for (int lx = 0; lx < 95; lx++) {
            for (int ly = 0; ly < 95; ly++) {
               int colour = terrainColours[getTerrainColour(lx, ly)];
               int colour1 = colour, colour2 = colour, diag = 0;
               if (plane == 1 || plane == 2) { colour = colour1 = colour2 = colourTransparent; }

               int deco = getTileDecoration(lx, ly);
               if (deco > 0) {
                  int decoType = ClientStream.sharedIntArrayN[deco - 1];                 // da.N
                  colour = colour1 = Panel.texK[deco - 1];      // qa.K
                  if (decoType == 4) {
                     colour = colour1 = 1;
                     if (deco == 12) colour = colour1 = 31;
                  } else if (decoType == 5) {
                     if (getWallDiagonal(lx, ly) > 0 && getWallDiagonal(lx, ly) < 24000) {
                        if (getTileDecorationOr(lx - 1, ly, colour2) != colourTransparent
                              && getTileDecorationOr(lx, ly - 1, colour2) != colourTransparent) {
                           colour = getTileDecorationOr(lx - 1, ly, colour2); diag = 0;
                        } else if (getTileDecorationOr(lx + 1, ly, colour2) != colourTransparent
                              && getTileDecorationOr(lx, ly + 1, colour2) != colourTransparent) {
                           colour1 = getTileDecorationOr(lx + 1, ly, colour2); diag = 0;
                        } else if (getTileDecorationOr(lx + 1, ly, colour2) != colourTransparent
                              && getTileDecorationOr(lx, ly - 1, colour2) != colourTransparent) {
                           colour1 = getTileDecorationOr(lx + 1, ly, colour2); diag = 1;
                        } else if (getTileDecorationOr(lx - 1, ly, colour2) != colourTransparent
                              && getTileDecorationOr(lx, ly + 1, colour2) != colourTransparent) {
                           colour = getTileDecorationOr(lx - 1, ly, colour2); diag = 1;
                        }
                     }
                  } else if (decoType != 2 || (getWallDiagonal(lx, ly) > 0 && getWallDiagonal(lx, ly) < 24000)) {
                     int tt = getTileType(lx, ly);
                     if (getTileType(lx - 1, ly) != tt && getTileType(lx, ly - 1) != tt) {
                        colour = colour2; diag = 0;
                     } else if (getTileType(lx + 1, ly) != tt && getTileType(lx, ly + 1) != tt) {
                        colour1 = colour2; diag = 0;
                     } else if (getTileType(lx + 1, ly) != tt && getTileType(lx, ly - 1) != tt) {
                        colour1 = colour2; diag = 1;
                     } else if (getTileType(lx - 1, ly) != tt && getTileType(lx, ly + 1) != tt) {
                        colour = colour2; diag = 1;
                     }
                  }
                  if (DecodeBuffer.landscapeFaceFlags[deco - 1] != 0) objectAdjacency[lx][ly] = CacheFile.or(objectAdjacency[lx][ly], 64);
                  if (ClientStream.sharedIntArrayN[deco - 1] == 2) objectAdjacency[lx][ly] = CacheFile.or(objectAdjacency[lx][ly], 128);
               }

               // obf method402(var15=diag, var82=colour1, lx, ly, var74=colour):
               // its "colour2" param receives colour1, its "colour1" param receives colour.
               method402(diag, colour1, lx, ly, colour);
               int twist = ((getTerrainHeight(lx + 1, ly + 1) - getTerrainHeight(lx + 1, ly))
                     + getTerrainHeight(lx, ly + 1)) - getTerrainHeight(lx, ly);
               if (colour != colour1 || twist != 0) {
                  if (diag == 0) {
                     if (colour != colourTransparent) {
                        int[] f = {ly + lx * 96 + 96, ly + lx * 96, ly + lx * 96 + 1};
                        int fid = terrain.createFace(3, f, colourTransparent, colour, false);
                        localX[fid] = lx; localY[fid] = ly; terrain.faceTag[fid] = fid + 200000;
                     }
                     if (colour1 != colourTransparent) {
                        int[] f = {ly + lx * 96 + 1, ly + lx * 96 + 97, ly + lx * 96 + 96};
                        int fid = terrain.createFace(3, f, colourTransparent, colour1, false);
                        localX[fid] = lx; localY[fid] = ly; terrain.faceTag[fid] = fid + 200000;
                     }
                  } else {
                     if (colour != colourTransparent) {
                        int[] f = {ly + lx * 96 + 1, ly + lx * 96 + 97, ly + lx * 96};
                        int fid = terrain.createFace(3, f, colourTransparent, colour, false);
                        localX[fid] = lx; localY[fid] = ly; terrain.faceTag[fid] = fid + 200000;
                     }
                     if (colour1 != colourTransparent) {
                        int[] f = {ly + lx * 96 + 96, ly + lx * 96, ly + lx * 96 + 97};
                        int fid = terrain.createFace(3, f, colourTransparent, colour1, false);
                        localX[fid] = lx; localY[fid] = ly; terrain.faceTag[fid] = fid + 200000;
                     }
                  }
               } else if (colour != colourTransparent) {
                  int[] f = {ly + lx * 96 + 96, ly + lx * 96, ly + lx * 96 + 1, ly + lx * 96 + 97};
                  int fid = terrain.createFace(4, f, colourTransparent, colour, false);
                  localX[fid] = lx; localY[fid] = ly; terrain.faceTag[fid] = fid + 200000;
               }
            }
         }

         // ── 3. overlay (type-4) decoration quads ──
         buildOverlayTriangles(terrain);

         // ── 4. split parent model into the 8×8 ground-tile grid ──
         terrain.setLight(-50, 40, -10, -50, true, 48, 105); // ca.a(IIIIZII) -> setLight
         terrainModels = parentModel.split(0, 8, 1536, 112, 64, 233, 1536, false, 0);
         for (int n = 0; n < 64; n++) scene.addModel(terrainModels[n]); // lb.a(ca,B) -> addModel

         // cache vertex heights for elevation queries
         for (int gx = 0; gx < regionWidth; gx++)
            for (int gy = 0; gy < regionHeight; gy++)
               terrainHeightLocal[gx][gy] = getTerrainHeight(gx, gy);
      }

      // ── 5. standing walls into the shared model, split + registered per plane ──
      parentModel.clear(1);  // ca.c(I) -> clear
      int minimapColour = 0x606060;
      for (int lx = 0; lx < 95; lx++) {
         for (int ly = 0; ly < 95; ly++) {
            int wall = getWallEastwest(lx, ly);
            if (wall > 0 && (Scene.diagScratch[wall - 1] == 0 || memberWorld)) {
               method422(wall - 1, parentModel, lx + 1, ly, lx, -14584, ly);
               if (flag && StringCodec.DEAD_INT_ARRAY[wall - 1] != 0) {
                  objectAdjacency[lx][ly] = CacheFile.or(objectAdjacency[lx][ly], 1);
                  if (ly > 0) orObjectAdjacency(4, ly - 1, lx);
               }
               if (flag) surface.drawLineHoriz(3, minimapColour, lx * 3, ly * 3, (byte) -109); // ua.b(IIIIB) -> drawLineHoriz
            }
            wall = getWallNorthsouth(lx, ly);
            if (wall > 0 && (Scene.diagScratch[wall - 1] == 0 || memberWorld)) {
               method422(wall - 1, parentModel, lx, ly, lx, -14584, ly + 1);
               if (flag && StringCodec.DEAD_INT_ARRAY[wall - 1] != 0) {
                  objectAdjacency[lx][ly] = CacheFile.or(objectAdjacency[lx][ly], 2);
                  if (lx > 0) orObjectAdjacency(8, ly, lx - 1);
               }
               if (flag) surface.drawLineVert(lx * 3, 3 * ly, minimapColour, 3, 0); // ua.b(IIIII) -> drawLineVert
            }
            wall = getWallDiagonal(lx, ly);
            if (wall > 0 && wall < 12000 && (Scene.diagScratch[wall - 1] == 0 || memberWorld)) {
               method422(wall - 1, parentModel, lx + 1, ly, lx, -14584, ly + 1);
               if (flag && StringCodec.DEAD_INT_ARRAY[wall - 1] != 0)
                  objectAdjacency[lx][ly] = CacheFile.or(objectAdjacency[lx][ly], 32);
               if (flag) { // setPixel(x, y, magic, colour)
                  surface.setPixel(3 * ly, lx * 3, 82, minimapColour);       // ua.a(IIII) -> setPixel
                  surface.setPixel(1 + 3 * ly, 1 + lx * 3, 69, minimapColour);
                  surface.setPixel(2 + 3 * ly, lx * 3 + 2, 65, minimapColour);
               }
            }
            if (wall > 12000 && wall < 24000 && (Scene.diagScratch[wall - 12001] == 0 || memberWorld)) {
               method422(wall - 12001, parentModel, lx, ly, lx + 1, -14584, ly + 1);
               if (flag && StringCodec.DEAD_INT_ARRAY[wall - 12001] != 0)
                  objectAdjacency[lx][ly] = CacheFile.or(objectAdjacency[lx][ly], 16);
               if (flag) {
                  surface.setPixel(3 * ly, 2 + 3 * lx, 116, minimapColour); // ua.a(IIII) -> setPixel
                  surface.setPixel(ly * 3 + 1, lx * 3 + 1, 99, minimapColour);
                  surface.setPixel(2 + 3 * ly, lx * 3, 90, minimapColour);
               }
            }
         }
      }
      if (flag) surface.drawSpriteMinimap(285, 0, 0, -27966, baseMediaSprite - 1, 285); // ua.b(IIIIII) -> drawSpriteMinimap
      parentModel.setLight(-50, 60, -10, -50, false, 24, 122); // ca.a(IIIIZII) -> setLight
      wallModels[plane] = parentModel.split(0, 8, 1536, -120, 64, 338, 1536, true, 0);
      for (int n = 0; n < 64; n++) scene.addModel(wallModels[plane][n]); // lb.a(ca,B) -> addModel

      // ── 6. method428: pre-pass that bumps wall-top vertices ──
      // obf call args map to method428(objectId, bx, ay, by, ax).
      for (int lx = 0; lx < 95; lx++) {
         for (int ly = 0; ly < 95; ly++) {
            int wall = getWallEastwest(lx, ly);
            if (wall > 0) method428(wall - 1, lx + 1, ly, ly, lx);
            wall = getWallNorthsouth(lx, ly);
            if (wall > 0) method428(wall - 1, lx, ly, ly + 1, lx);
            wall = getWallDiagonal(lx, ly);
            if (wall > 0 && wall < 12000) method428(wall - 1, lx + 1, ly, ly + 1, lx);
            if (wall > 12000 && wall < 24000) method428(wall - 12001, lx, ly, ly + 1, lx + 1);
         }
      }

      // ── 7. roof-top levelling: flatten each roofed cell's 4 corners to the
      //       max corner height (obf label909). Existence query is swapped (ly,lx). ──
      for (int lx = 1; lx < 95; lx++) {
         for (int ly = 1; ly < 95; ly++) {
            if (getWallRoof(ly, lx) <= 0) continue;
            int h0 = terrainHeightLocal[lx][ly];
            int h1 = terrainHeightLocal[lx + 1][ly];
            int h2 = terrainHeightLocal[lx + 1][ly + 1];
            int h3 = terrainHeightLocal[lx][ly + 1];
            if (h0 > 80000) h0 -= 80000;
            if (h1 > 80000) h1 -= 80000;
            if (h2 > 80000) h2 -= 80000;
            if (h3 > 80000) h3 -= 80000;
            int max = 0;
            if (h0 > max) max = h0;
            if (h1 > max) max = h1;
            if (h2 > max) max = h2;
            if (h3 > max) max = h3;
            if (max >= 80000) max -= 80000;
            terrainHeightLocal[lx][ly]         = h0 < 80000 ? max : terrainHeightLocal[lx][ly] - 80000;
            terrainHeightLocal[lx + 1][ly]     = h1 < 80000 ? max : terrainHeightLocal[lx + 1][ly] - 80000;
            terrainHeightLocal[lx + 1][ly + 1] = h2 < 80000 ? max : terrainHeightLocal[lx + 1][ly + 1] - 80000;
            terrainHeightLocal[lx][ly + 1]     = h3 < 80000 ? max : terrainHeightLocal[lx][ly + 1] - 80000;
         }
      }

      // ── 8. roof triangles into the shared model, split + registered ──
      parentModel.clear(1);  // ca.c(I) -> clear
      buildRoofs();
      parentModel.setLight(-50, 50, -10, -50, true, 50, -98); // ca.a(IIIIZII) -> setLight
      roofModels[plane] = parentModel.split(0, 8, 1536, -112, 64, 169, 1536, true, 0);
      for (int n = 0; n < 64; n++) scene.addModel(roofModels[plane][n]); // lb.a(ca,B) -> addModel
      if (roofModels[plane][0] == null) throw new RuntimeException(STR_NULL_ROOF);

      // ── 9. strip the +80000 sentinel from cached heights ──
      for (int gx = 0; gx < regionWidth; gx++)
         for (int gy = 0; gy < regionHeight; gy++)
            if (terrainHeightLocal[gx][gy] >= 80000)
               terrainHeightLocal[gx][gy] -= 80000;
   }

   /** Decoration type of tile (x,y), or -1 if no decoration: folds getTileDecoration+da.N. */
   private int getTileTypeOnPlane(int x, int y) {
      int deco = getTileDecoration(x, y);
      if (deco <= 0) return -1;
      return ClientStream.sharedIntArrayN[deco - 1];
   }

   /**
    * Emit the type-4 overlay (bridge/floor) quads. Faithful extract of the second
    * tile loop of the obfuscated mesher: each emits a flat quad across the four
    * corner vertices and re-tiles the minimap.
    */
   private void buildOverlayTriangles(GameModel terrain) {
      for (int lx = 1; lx < 95; lx++) {
         for (int ly = 1; ly < 95; ly++) {
            int d0 = getTileDecoration(lx, ly);
            if (d0 > 0 && ClientStream.sharedIntArrayN[d0 - 1] == 4) {
               emitOverlayQuad(terrain, lx, ly, Panel.texK[d0 - 1]);
            } else if (d0 == 0 || ClientStream.sharedIntArrayN[d0 - 1] != 3) {
               int dN = getTileDecoration(lx, ly + 1);
               if (dN > 0 && ClientStream.sharedIntArrayN[dN - 1] == 4)
                  emitOverlayQuad(terrain, lx, ly, Panel.texK[dN - 1]);
               int dS = getTileDecoration(lx, ly - 1);
               if (dS > 0 && ClientStream.sharedIntArrayN[dS - 1] == 4)
                  emitOverlayQuad(terrain, lx, ly, Panel.texK[dS - 1]);
               int dE = getTileDecoration(lx + 1, ly);
               if (dE > 0 && ClientStream.sharedIntArrayN[dE - 1] == 4)
                  emitOverlayQuad(terrain, lx, ly, Panel.texK[dE - 1]);
               int dW = getTileDecoration(lx - 1, ly);
               if (dW > 0 && ClientStream.sharedIntArrayN[dW - 1] == 4)
                  emitOverlayQuad(terrain, lx, ly, Panel.texK[dW - 1]);
            }
         }
      }
   }

   /** Helper for {@link #buildOverlayTriangles}: a flat type-4 quad at tile (lx,ly). */
   private void emitOverlayQuad(GameModel terrain, int lx, int ly, int colour) {
      int v0 = terrain.vertexAt(lx * 128, ly * 128, -getTerrainHeight(lx, ly), -116);
      int v1 = terrain.vertexAt((lx + 1) * 128, ly * 128, -getTerrainHeight(lx + 1, ly), -116);
      int v2 = terrain.vertexAt((lx + 1) * 128, (ly + 1) * 128, -getTerrainHeight(lx + 1, ly + 1), -116);
      int v3 = terrain.vertexAt(lx * 128, (ly + 1) * 128, -getTerrainHeight(lx, ly + 1), -116);
      int fid = terrain.createFace(4, new int[]{v0, v1, v2, v3}, colour, colourTransparent, false);
      localX[fid] = lx; localY[fid] = ly; terrain.faceTag[fid] = fid + 200000;
      method402(0, colour, lx, ly, colour);
   }

   /**
    * Build the sloped roof triangles — FAITHFUL transcription of the clean-base
    * roof loop (obf vars var44=outer/lx, var56=inner/ly). Reproduces the four
    * corner points P0..P3, the +80000 ridge-raise bookkeeping, the 16-unit corner
    * taper, and the seven-way face-emission cascade exactly as the obfuscated
    * code (which differs in vertex ordering from the rev-204 oracle).
    *
    * <p>The four roof corner points (world x, world z, −height):
    * P0=({@code cx0},{@code cy0},{@code h0}) NW, P1=({@code cx1},{@code cy1},{@code h1}) NE,
    * P2=({@code cx2},{@code cy2},{@code h2}) SE, P3=({@code cx3},{@code cy3},{@code h3}) SW.
    * Note the existence query is {@code getWallRoof(ly,lx)} (args swapped), while
    * the corner/taper queries use (lx,ly) directly — faithful to the clean base.</p>
    */
   private void buildRoofs() {
      for (int lx = 1; lx < 95; lx++) {              // obf var44
         for (int ly = 1; ly < 95; ly++) {           // obf var56
            int roof = getWallRoof(ly, lx);           // obf d(var56,var44,126) — args swapped
            if (roof <= 0) continue;

            int cx0 = 128 * lx, cy0 = ly * 128;       // obf var129,var130
            int cx1 = 128 + cx0, cy1 = cy0;           // obf var131,var25
            int cx2 = cx1, cy2 = 128 + cy0;           // obf var26,var132
            int cx3 = cx0, cy3 = cy2;                 // obf var133,var27
            int h0 = terrainHeightLocal[lx][ly];      // obf var28 = ab[var81][var89]
            int h1 = terrainHeightLocal[lx + 1][ly];  // obf var29 = ab[var96][var103]
            int h2 = terrainHeightLocal[lx + 1][ly + 1]; // obf var30 = ab[var111][var120]
            int h3 = terrainHeightLocal[lx][ly + 1];  // obf var31 = ab[var123][var128]
            int ridge = ErrorHandler.unusedIntTable[roof - 1];    // i.g — roof ridge height (var32)

            if (hasRoof(false, lx, ly) && h0 < 80000) { h0 += ridge + 80000; terrainHeightLocal[lx][ly] = h0; }
            if (hasRoof(false, lx + 1, ly) && h1 < 80000) { h1 += ridge + 80000; terrainHeightLocal[lx + 1][ly] = h1; }
            if (hasRoof(false, lx + 1, ly + 1) && h2 < 80000) { h2 += ridge + 80000; terrainHeightLocal[lx + 1][ly + 1] = h2; }
            if (h1 >= 80000) h1 -= 80000;
            if (h2 >= 80000) h2 -= 80000;
            if (hasRoof(false, lx, ly + 1) && h3 < 80000) { h3 += ridge + 80000; terrainHeightLocal[lx][ly + 1] = h3; }
            if (h0 >= 80000) h0 -= 80000;
            if (h3 >= 80000) h3 -= 80000;

            byte off = 16;
            if (!isRoofCorner(lx - 1, ly)) cx0 -= off;
            if (!isRoofCorner(lx + 1, ly)) cx0 += off;
            if (!isRoofCorner(lx, ly - 1)) cy0 -= off;
            if (!isRoofCorner(lx, ly + 1)) cy0 += off;
            if (!isRoofCorner(lx + 1 - 1, ly)) cx1 -= off;
            if (!isRoofCorner(lx + 1 + 1, ly)) cx1 += off;
            if (!isRoofCorner(lx + 1, ly - 1)) cy1 -= off;
            if (!isRoofCorner(lx + 1, ly + 1)) cy1 += off;
            if (!isRoofCorner(lx + 1 - 1, ly + 1)) cx2 -= off;
            if (!isRoofCorner(lx + 1 + 1, ly + 1)) cx2 += off;
            if (!isRoofCorner(lx + 1, ly + 1 - 1)) cy2 -= off;
            if (!isRoofCorner(lx + 1, ly + 1 + 1)) cy2 += off;
            if (!isRoofCorner(lx - 1, ly + 1)) cx3 -= off;
            if (!isRoofCorner(lx + 1, ly + 1)) cx3 += off;
            if (!isRoofCorner(lx, ly + 1 - 1)) cy3 -= off;
            if (!isRoofCorner(lx, ly + 1 + 1)) cy3 += off;

            int texture = CacheFile.sharedScratch[roof - 1];  // d.g — roof texture
            h0 = -h0; h1 = -h1; h2 = -h2; h3 = -h3;

            int diag = getWallDiagonal(lx, ly);       // obf c(var44,var56,-49)

            // Seven-way emission cascade, transcribed from the obf inverted-if tree.
            // OUTER: diag in (12000,24000) AND getWallRoof(ly-1,lx-1)==0 → fall to FACE_LAST.
            if (diag <= 12000 || diag >= 24000 || getWallRoof(ly - 1, lx - 1) != 0) {
               // CASE 1: diag in (12000,24000) AND getWallRoof(ly+1,lx+1)==0
               if (diag > 12000 && diag < 24000 && getWallRoof(ly + 1, lx + 1) == 0) {
                  int[] f = {parentModel.vertexAt(cx0, cy0, h0, -128), parentModel.vertexAt(cx1, cy1, h1, -122),
                             parentModel.vertexAt(cx3, cy3, h3, 12)};
                  parentModel.createFace(3, f, texture, colourTransparent, false);
               } else if (diag <= 0 || diag >= 12000 || getWallRoof(ly - 1, lx + 1) != 0) {
                  // INNER: diag in (0,12000) AND getWallRoof(ly-1,lx+1)==0 → fall to FACE_PRE_LAST
                  if (diag <= 0 || diag >= 12000 || getWallRoof(ly + 1, lx - 1) != 0) {
                     // not a (0,12000)+roof(ly+1,lx-1)==0 corner
                     if (h0 != h1 || h3 != h2) {
                        if (h0 != h3 || h2 != h1) {
                           // CASE 7: ridge split — two triangles, orientation by roof neighbours
                           boolean flag1 = !(getWallRoof(ly - 1, lx - 1) > 0 || getWallRoof(ly + 1, lx + 1) > 0);
                           if (!flag1) {
                              int[] fa = {parentModel.vertexAt(cx1, cy1, h1, -114), parentModel.vertexAt(cx2, cy2, h2, 101),
                                          parentModel.vertexAt(cx0, cy0, h0, -126)};
                              parentModel.createFace(3, fa, texture, colourTransparent, false);
                              int[] fb = {parentModel.vertexAt(cx3, cy3, h3, -107), parentModel.vertexAt(cx0, cy0, h0, 63),
                                          parentModel.vertexAt(cx2, cy2, h2, 44)};
                              parentModel.createFace(3, fb, texture, colourTransparent, false);
                           } else {
                              int[] fa = {parentModel.vertexAt(cx0, cy0, h0, -112), parentModel.vertexAt(cx1, cy1, h1, -118),
                                          parentModel.vertexAt(cx3, cy3, h3, 103)};
                              parentModel.createFace(3, fa, texture, colourTransparent, false);
                              int[] fb = {parentModel.vertexAt(cx2, cy2, h2, -128), parentModel.vertexAt(cx3, cy3, h3, -119),
                                          parentModel.vertexAt(cx1, cy1, h1, 52)};
                              parentModel.createFace(3, fb, texture, colourTransparent, false);
                           }
                        } else {
                           // CASE 6: h0==h3 && h1==h2 → quad (P3,P0,P1,P2)
                           int[] f = {parentModel.vertexAt(cx3, cy3, h3, -104), parentModel.vertexAt(cx0, cy0, h0, 23),
                                      parentModel.vertexAt(cx1, cy1, h1, 91), parentModel.vertexAt(cx2, cy2, h2, 13)};
                           parentModel.createFace(4, f, texture, colourTransparent, false);
                        }
                     } else {
                        // CASE 5: h0==h1 && h3==h2 → quad (P0,P1,P2,P3)
                        int[] f = {parentModel.vertexAt(cx0, cy0, h0, 78), parentModel.vertexAt(cx1, cy1, h1, 46),
                                   parentModel.vertexAt(cx2, cy2, h2, -113), parentModel.vertexAt(cx3, cy3, h3, -125)};
                        parentModel.createFace(4, f, texture, colourTransparent, false);
                     }
                  } else {
                     // CASE 4: diag in (0,12000) AND getWallRoof(ly+1,lx-1)==0 → tri (P1,P2,P0)
                     int[] f = {parentModel.vertexAt(cx1, cy1, h1, 121), parentModel.vertexAt(cx2, cy2, h2, 39),
                                parentModel.vertexAt(cx0, cy0, h0, 73)};
                     parentModel.createFace(3, f, texture, colourTransparent, false);
                  }
               } else {
                  // CASE 3: diag in (0,12000) AND getWallRoof(ly-1,lx+1)==0 → tri (P3,P0,P2)
                  int[] f = {parentModel.vertexAt(cx3, cy3, h3, -107), parentModel.vertexAt(cx0, cy0, h0, -122),
                             parentModel.vertexAt(cx2, cy2, h2, 35)};
                  parentModel.createFace(3, f, texture, colourTransparent, false);
               }
            } else {
               // CASE 2 (FACE_LAST): diag in (12000,24000) AND getWallRoof(ly-1,lx-1)==0 → tri (P2,P3,P1)
               int[] f = {parentModel.vertexAt(cx2, cy2, h2, -120), parentModel.vertexAt(cx3, cy3, h3, -116),
                          parentModel.vertexAt(cx1, cy1, h1, 117)};
               parentModel.createFace(3, f, texture, colourTransparent, false);
            }
         }
      }
   }

   // ════════════════════════ map file loader ═══════════════════════════

   /**
    * obf: b(int plane, int chunk, int cx, int unused=0, int cy) — oracle
    * World.loadSection(x,y,plane,chunk). RECONSTRUCTED from the bytecode (the
    * clean base bailed with "Couldn't be decompiled") cross-checked against the
    * oracle. Loads/decodes one 48×48 map chunk {@code m<plane><cx><cy>} into
    * quadrant {@code chunk} of every terrain grid:
    * <ul>
    *   <li>{@code .hei}: RLE heights then colours (value ≥128 ⇒ repeat previous
    *       (value-128) times) each followed by a column-major prefix-sum smooth
    *       (seed 64 for heights, 35 for colours, &amp;0x7F, ×2);</li>
    *   <li>{@code .dat}: walls N/S, E/W verbatim; diagonal &amp;0xFF then +12000
    *       if nonzero; roof RLE (repeat ⇒ 0); decoration RLE (repeat ⇒ lastVal);
    *       direction RLE (repeat ⇒ 0);</li>
    *   <li>{@code .loc}: diagonal +48000 for object placements;</li>
    *   <li>fallback raw {@code .jm}: delta-encoded heights/colours, verbatim
    *       walls, 16-bit big-endian diagonal, verbatim roof/decoration/direction.</li>
    * </ul>
    * On any I/O error the quadrant is zeroed; plane 0 seeds decoration −6 and
    * plane 3 seeds decoration 8 as border sentinels (into R = tileDecoration).
    */
   private void loadMapData(int plane, int chunk, int cx, int unused, int cy) {
      if (unused != 0) return;  // obf var4 anti-tamper guard (always 0)
      String name = "m" + plane + cx / 10 + cx % 10 + cy / 10 + cy % 10;
      try {
         if (landscapePack != null) {
            byte[] data = StreamFactory.lookupEntityDefRecord(name + STR_HEI, 0, landscapePack); // na.a -> lookupEntityDefRecord (4th dummy dropped)
            if (data == null && memberLandscapePack != null)
               data = StreamFactory.lookupEntityDefRecord(name + STR_HEI, 0, memberLandscapePack);

            if (data != null && data.length > 0) {
               int off = 0, last = 0, tile;
               for (tile = 0; tile < 2304; ) {                  // heights: RLE
                  int v = data[off++] & 0xFF;
                  if (v < 128) { terrainHeight[chunk][tile++] = (byte) v; last = v; }
                  else for (int r = 0; r < v - 128; r++) terrainHeight[chunk][tile++] = (byte) last;
               }
               last = 64;                                       // prefix smooth (column-major)
               for (int ty = 0; ty < 48; ty++)
                  for (int tx = 0; tx < 48; tx++) {
                     last = (terrainHeight[chunk][tx * 48 + ty] + last) & 0x7F;
                     terrainHeight[chunk][tx * 48 + ty] = (byte) (last * 2);
                  }
               last = 0;
               for (tile = 0; tile < 2304; ) {                  // colours: RLE
                  int v = data[off++] & 0xFF;
                  if (v < 128) { terrainColour[chunk][tile++] = (byte) v; last = v; }
                  else for (int r = 0; r < v - 128; r++) terrainColour[chunk][tile++] = (byte) last;
               }
               last = 35;
               for (int ty = 0; ty < 48; ty++)
                  for (int tx = 0; tx < 48; tx++) {
                     last = (terrainColour[chunk][tx * 48 + ty] + last) & 0x7F;
                     terrainColour[chunk][tx * 48 + ty] = (byte) (last * 2);
                  }
            } else {
               for (int tile = 0; tile < 2304; tile++) { terrainHeight[chunk][tile] = 0; terrainColour[chunk][tile] = 0; }
            }

            byte[] data2 = StreamFactory.lookupEntityDefRecord(name + STR_DAT, 0, mapPack); // na.a -> lookupEntityDefRecord
            if (data2 == null && memberMapPack != null)
               data2 = StreamFactory.lookupEntityDefRecord(name + STR_DAT, 0, memberMapPack);
            if (data2 == null || data2.length == 0) throw new IOException();

            int off = 0, tile;
            for (tile = 0; tile < 2304; tile++) wallsNorthsouth[chunk][tile] = data2[off++];
            for (tile = 0; tile < 2304; tile++) wallsEastwest[chunk][tile] = data2[off++];
            for (tile = 0; tile < 2304; tile++) wallsDiagonal[chunk][tile] = data2[off++] & 0xFF;
            for (tile = 0; tile < 2304; tile++) {
               int v = data2[off++] & 0xFF;
               if (v > 0) wallsDiagonal[chunk][tile] = v + 12000;
            }
            for (tile = 0; tile < 2304; ) {                     // roof: RLE (repeat ⇒ 0)
               int v = data2[off++] & 0xFF;
               if (v < 128) wallsRoof[chunk][tile++] = (byte) v;
               else for (int r = 0; r < v - 128; r++) wallsRoof[chunk][tile++] = 0;
            }
            int last = 0;
            for (tile = 0; tile < 2304; ) {                     // decoration: RLE (repeat ⇒ lastVal)
               int v = data2[off++] & 0xFF;
               if (v < 128) { tileDecoration[chunk][tile++] = (byte) v; last = v; }
               else for (int r = 0; r < v - 128; r++) tileDecoration[chunk][tile++] = (byte) last;
            }
            for (tile = 0; tile < 2304; ) {                     // direction: RLE (repeat ⇒ 0)
               int v = data2[off++] & 0xFF;
               if (v < 128) tileDirection[chunk][tile++] = (byte) v;
               else for (int r = 0; r < v - 128; r++) tileDirection[chunk][tile++] = 0;
            }

            byte[] loc = StreamFactory.lookupEntityDefRecord(name + STR_LOC, 0, mapPack); // na.a -> lookupEntityDefRecord
            if (loc != null && loc.length > 0) {
               int o2 = 0;
               for (tile = 0; tile < 2304; ) {
                  int v = loc[o2++] & 0xFF;
                  if (v < 128) wallsDiagonal[chunk][tile++] = v + 48000;
                  else tile += v - 128;
               }
            }
            return;
         }

         // fallback: raw delta-encoded .jm map file
         byte[] jm = new byte[20736];
         // BEHAVIORAL FIX (RENDER_DEOB_GAPS / OBF_MEMBER_MAP GameCharacter.readFromStore row):
         // clean k.java:3535 bytecode is `invokestatic ta.a (Ljava/lang/String;I[BI)V`,
         // i.e. ta(GameCharacter).readFromStore — NOT t(EntityDef).a (which has no such
         // (String,int,byte[],int) overload). Reads the .jm bytes into the buffer.
         GameCharacter.readFromStore(STR_MAPS_DIR + name + STR_JM, -19675, jm, 20736);
         int off = 0, acc = 0, tile;
         for (tile = 0; tile < 2304; tile++) { acc = (acc + jm[off++]) & 0xFF; terrainHeight[chunk][tile] = (byte) acc; }
         acc = 0;
         for (tile = 0; tile < 2304; tile++) { acc = (acc + jm[off++]) & 0xFF; terrainColour[chunk][tile] = (byte) acc; }
         for (tile = 0; tile < 2304; tile++) wallsNorthsouth[chunk][tile] = jm[off++];
         for (tile = 0; tile < 2304; tile++) wallsEastwest[chunk][tile] = jm[off++];
         for (tile = 0; tile < 2304; tile++) {
            wallsDiagonal[chunk][tile] = (jm[off] & 0xFF) * 256 + (jm[off + 1] & 0xFF);
            off += 2;
         }
         for (tile = 0; tile < 2304; tile++) wallsRoof[chunk][tile] = jm[off++];
         for (tile = 0; tile < 2304; tile++) tileDecoration[chunk][tile] = jm[off++];
         for (tile = 0; tile < 2304; tile++) tileDirection[chunk][tile] = jm[off++];
      } catch (IOException ex) {
         for (int tile = 0; tile < 2304; tile++) {
            terrainHeight[chunk][tile] = 0;
            terrainColour[chunk][tile] = 0;
            wallsNorthsouth[chunk][tile] = 0;
            wallsEastwest[chunk][tile] = 0;
            wallsDiagonal[chunk][tile] = 0;
            tileDecoration[chunk][tile] = 0;
            if (plane == 0) tileDecoration[chunk][tile] = -6;  // obf stores -6 into R (tileDecoration)
            if (plane == 3) tileDecoration[chunk][tile] = 8;
            wallsRoof[chunk][tile] = 0;
         }
      }
   }
}
