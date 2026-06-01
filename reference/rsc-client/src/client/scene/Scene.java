package client.scene;

import java.io.IOException;

import client.world.World;          // obf: lb  (the model-list / polygon scene that actually renders the ca[] models)
import client.world.WorldEntity;    // obf: w
import client.scene.GameModel;      // obf: ca
import client.scene.Surface;        // obf: ua
import client.util.StreamFactory;   // obf: na  (loads cache records → byte[])
import client.util.ArrayUtil;       // obf: ab
import client.data.BZip;            // obf: aa  (referenced as ea.a here for bzip-decompress)
import client.net.StreamBase;       // obf: ib  (ib.a(...) bitmask helper, ib.d[] door-height table)
import client.data.CacheFile;       // obf: d   (d.a(...) bitmask-OR helper, d.g[] roof-texture table)

/**
 * Scene — 3D world/terrain manager (obfuscated class {@code k}).
 *
 * <p>NOTE on naming: per the project NAMING.md this class is labelled
 * "Scene" because it owns the {@link GameModel} ({@code ca}) arrays for the
 * world. Functionally, however, this revision of the client merged what the
 * mudclient204 oracle calls {@code World} into this class: it loads the
 * landscape map files (`.hei` heights, `.dat`/wall data, `.loc` objects,
 * `.jm` tiles), decodes them into the per-quadrant terrain grids, and builds
 * the terrain / wall / roof {@link GameModel}s. The actual polygon
 * projection, raster and painter-sort live in the sibling {@code World}
 * ({@code lb}) class referenced here as {@link #world} — it is the object the
 * built models are registered into via {@code world.addModel(...)}.</p>
 *
 * <p>Region layout: the visible region is 96×96 tiles, stored as four 48×48
 * quadrants (indices 0..3). Most accessors take an absolute (x,y) in 0..95 and
 * resolve which quadrant array and local offset to read — that is the repeated
 * "if (x>=48 && y<48) {...}" pattern throughout. Heights/colours are stored as
 * byte grids of length 2304 (=48*48) per quadrant.</p>
 *
 * <p>Obfuscation stripped while reading: the opaque predicate
 * {@code boolean bl = client.vh;} (always false) and its dead branches; the
 * per-method profiling counters ({@code ++X;} etc.); the per-method
 * {@code try{...}catch(RuntimeException){throw ErrorHandler.a(e,"sig")}}
 * wrappers; the anti-tamper dummy-parameter guards
 * ({@code if (p != <magic>) return;} / {@code 79 % ((p-30)/35)}); junk masks
 * before shifts; and the XOR string pool decoded by the two {@code z(...)}
 * helpers.</p>
 */
final class Scene {

   // ───────────────────────── region constants ─────────────────────────
   static final int REGION_SIZE = 96;       // region is 96×96 tiles
   static final int QUADRANT_SIZE = 48;     // stored as four 48×48 quadrants
   static final int TILE_SIZE = 128;        // world units per tile (obf literal 128 / anInt585)
   static final int COLOUR_TRANSPARENT = 12345678; // World.colourTransparent sentinel ("invisible")

   // ───────────────────────── collaborators ────────────────────────────
   private World world;  // obf: c  — the polygon Scene/World that renders our ca[] models
   private Surface surface; // obf: U — software framebuffer (minimap blits, blackScreen)

   // ───────────────────────── world flags ──────────────────────────────
   private boolean memberWorld;     // obf: H — true ⇒ build object models even when normally adjacency-hidden (members area)
   private boolean worldInitialised;// obf: nb — has a section been loaded (⇒ dispose before reset)
   boolean playerAlive;             // obf: Z — (oracle aBoolean592) renders the parent fence model when set
   int baseMediaSprite;             // obf: x — base sprite id for tile decorations (=750)

   // ───────────────────── per-quadrant terrain grids ───────────────────
   private byte[][] terrainHeight;     // obf: L  — [4][2304] tile heights (×3 when read)
   private byte[][] terrainColour;     // obf: eb — [4][2304] terrain colour index
   private byte[][] wallsNorthsouth;   // obf: f  — [4][2304] N/S wall id
   private byte[][] wallsEastwest;     // obf: P  — [4][2304] E/W wall id
   private int[][]  wallsDiagonal;     // obf: s  — [4][2304] diagonal wall id (+12000 inverted, +48000 roof)
   private byte[][] tileDecoration;    // obf: A  — [4][2304] tile overlay/decoration id
   private byte[][] tileDirection;     // obf: R  — [4][2304] tile/diagonal direction
   private byte[][] wallsRoof;         // obf: mb — [4][2304] roof wall id

   // ─────────────────────── derived / scratch grids ────────────────────
   int[][] objectAdjacency;            // obf: bb — [96][96] passability/adjacency bitmask
   private int[][] routeVia;           // obf: B  — [96][96] BFS came-from directions (route())
   private int[][] terrainHeightLocal; // obf: ab — [96][96] vertex heights used while meshing
   private int[] terrainColours;       // obf: w  — [256] palette: ground/grass/path ramps

   // ───────────────────────── built models ─────────────────────────────
   private GameModel[] terrainModels;  // obf: F  — [64] ground-tile models (8×8 grid)
   GameModel[][] wallModels;           // obf: g  — [4][64] vertical wall models per plane
   GameModel[][] roofModels;           // obf: db — [4][64] roof models per plane
   private GameModel parentModel;      // obf: kb — shared builder model reused across sections

   // ─────────────────── per-face tile coordinate maps ──────────────────
   int[] localX;                       // obf: q  — [18432] face-id → tile x
   int[] localY;                       // obf: E  — [18432] face-id → tile y

   // ───────────────────── raw map archive packs ────────────────────────
   byte[] landscapePack;       // obf: Q  — free landscape archive (.hei/.dat/.loc/.jm)
   byte[] memberLandscapePack; // obf: I  — members landscape archive
   byte[] mapPack;             // obf: m  — free map archive
   byte[] memberMapPack;       // obf: gb — members map archive

   // ────────────────── profiling counters (dead; kept named) ───────────
   // Each obf method incremented a unique static counter on entry; all are
   // unused after the profiling strip. Retained as named no-ops.
   static int lb, z, ib, cb, t, u, N, j, C, h, n, jb, a, l, Y, d, M, r, p,
              J, K, T, k, o, fb, S, X, b, O, W, D, V, y, i, hb, v;
   static long e = 0L;                 // obf: e — unused profiling accumulator
   static String[] G = new String[100];// obf: G — unused scratch string table

   /**
    * XOR-decoded string pool (obf: {@code ob}). Most entries are method
    * signatures used only by the (now-stripped) error wrappers; the
    * meaningful ones are the map-file suffixes and messages:
    * {@code ".hei" ".dat" ".loc" ".jm"}, {@code "../gamedata/maps/"},
    * {@code "Unpacking "}, {@code "null roof!"}.
    */
   private static final String[] STR = decodePool();

   // Indices into STR that carry real (non-signature) content:
   private static final int STR_UNPACKING = 18;          // "Unpacking "
   private static final int STR_HEI_SUFFIX = 26;         // ".hei"
   private static final int STR_MAPS_DIR = 28;           // "../gamedata/maps/"
   private static final int STR_DAT_SUFFIX = 29;         // ".dat"
   private static final int STR_LOC_SUFFIX = 30;         // ".loc"
   private static final int STR_JM_SUFFIX = 31;          // ".jm"
   private static final int STR_NULL_ROOF = 40;          // "null roof!"

   /**
    * Constructs the scene/world and builds the terrain colour palette.
    * obf: {@code k(lb, ua)} — mirrors oracle {@code World(Scene, Surface)}.
    */
   Scene(World world, Surface surface) {
      memberWorld = false;
      terrainHeightLocal = new int[REGION_SIZE][REGION_SIZE];
      wallsRoof = new byte[4][2304];
      wallsEastwest = new byte[4][2304];
      tileDirection = new byte[4][2304];
      terrainColours = new int[256];
      routeVia = new int[REGION_SIZE][REGION_SIZE];
      wallsNorthsouth = new byte[4][2304];
      localY = new int[18432];
      worldInitialised = true;
      baseMediaSprite = 750;
      wallModels = new GameModel[4][64];
      terrainHeight = new byte[4][2304];
      terrainModels = new GameModel[64];
      wallsDiagonal = new int[4][2304];
      tileDecoration = new byte[4][2304];
      playerAlive = false;
      objectAdjacency = new int[REGION_SIZE][REGION_SIZE];
      roofModels = new GameModel[4][64];
      terrainColour = new byte[4][2304];
      localX = new int[18432];

      this.surface = surface;
      this.world = world;

      // Four blended colour ramps (each 64 entries) keyed by terrain colour idx:
      // 0..63   white→grey path/rock
      // 64..127 grass green
      // 128..191 dirt brown
      // 192..255 sand/water transition
      for (int n2 = 0; n2 < 64; n2++)
         terrainColours[n2] = World.rgb(255 - n2 * 4, 255 - (int) (n2 * 1.75D), 255 - n2 * 4);
      for (int n2 = 0; n2 < 64; n2++)
         terrainColours[n2 + 64] = World.rgb(n2 * 3, 144, 0);
      for (int n2 = 0; n2 < 64; n2++)
         terrainColours[n2 + 128] = World.rgb(192 - (int) (n2 * 1.5D), 144 - (int) (n2 * 1.5D), 0);
      for (int n2 = 0; n2 < 64; n2++)
         terrainColours[n2 + 192] = World.rgb(96 - (int) (n2 * 1.5D), 48 + (int) (n2 * 1.5D), 0);
   }

   // ═════════════════════════ grid accessors ═══════════════════════════
   // The eight terrain grids are 48×48 per quadrant. Each getter resolves the
   // (x,y)→quadrant mapping identically (the repeated 4-case block). Helper
   // formula: quadrant q, local index = lx*48 + ly (note: a couple of grids
   // index ly*48 + lx — kept verbatim to stay faithful to the obfuscated code).

   /** Quadrant index 0..3 for an absolute tile (x,y) in 0..95; -1 if OOB. */
   private static int quadrant(int x, int y) {
      if (x < 0 || x >= REGION_SIZE || y < 0 || y >= REGION_SIZE) return -1;
      if (x >= 48 && y < 48) return 1;
      if (x < 48 && y >= 48) return 2;
      if (x >= 48 && y >= 48) return 3;
      return 0;
   }

   /** Tile height (×3) at (x,y). obf: g(int magic, y, x) — oracle getTerrainHeight. */
   private int getTerrainHeight(int unused2, int y, int x) {
      // obf param1 was always 2 (an anti-tamper constant), kept as unused.
      if (x < 0 || x >= REGION_SIZE || y < 0 || y >= REGION_SIZE) return 0;
      int q = 0, lx = x, ly = y;
      if (x >= 48 && y < 48) { q = 1; lx -= 48; }
      else if (x < 48 && y >= 48) { q = 2; ly -= 48; }
      else if (x >= 48 && y >= 48) { q = 3; lx -= 48; ly -= 48; }
      return (terrainHeight[q][48 * lx + ly] & 0xFF) * 3;
   }

   /** Terrain colour index at (x,y). obf: a(byte magic, x, y) — getTerrainColour. */
   private int getTerrainColour(int unusedByte, int x, int y) {
      if (x < 0 || x >= REGION_SIZE || y < 0 || y >= REGION_SIZE) return 0;
      int q = 0, lx = x, ly = y;
      if (x >= 48 && y < 48) { q = 1; lx -= 48; }
      else if (x < 48 && y >= 48) { q = 2; ly -= 48; }
      else if (x >= 48 && y >= 48) { q = 3; lx -= 48; ly -= 48; }
      return terrainColour[q][ly + 48 * lx] & 0xFF;
   }

   /** N/S wall id at (x,y). obf: e(int magic, x, y) — getWallNorthsouth. */
   private int getWallNorthsouth(int unused, int x, int y) {
      if (x < 0 || x >= REGION_SIZE || y < 0 || y >= REGION_SIZE) return 0;
      int q = 0, lx = x, ly = y;
      if (x >= 48 && y < 48) { q = 1; lx -= 48; }
      else if (x < 48 && y >= 48) { q = 2; ly -= 48; }
      else if (x >= 48 && y >= 48) { q = 3; lx -= 48; ly -= 48; }
      return 0xFF & wallsNorthsouth[q][lx * 48 + ly];
   }

   /** E/W wall id at (x,y). obf: a(x, byte magic, y) — getWallEastwest. */
   private int getWallEastwest(int x, int unusedByte, int y) {
      if (x < 0 || x >= REGION_SIZE || y < 0 || y >= REGION_SIZE) return 0;
      int q = 0, lx = x, ly = y;
      if (x >= 48 && y < 48) { q = 1; lx -= 48; }
      else if (x < 48 && y >= 48) { q = 2; ly -= 48; }
      else if (x >= 48 && y >= 48) { q = 3; lx -= 48; ly -= 48; }
      return 0xFF & wallsEastwest[q][lx * 48 + ly];
   }

   /** Diagonal wall id at (x,y). obf: c(x, y, int magic) — getWallDiagonal. */
   private int getWallDiagonal(int x, int y, int unused) {
      if (x < 0 || x >= REGION_SIZE || y < 0 || y >= REGION_SIZE) return 0;
      int q = 0, lx = x, ly = y;
      if (x >= 48 && y < 48) { q = 1; lx -= 48; }
      else if (x < 48 && y >= 48) { q = 2; ly -= 48; }
      else if (x >= 48 && y >= 48) { q = 3; lx -= 48; ly -= 48; }
      return wallsDiagonal[q][lx * 48 + ly];
   }

   /** Tile decoration/overlay id at (x,y). obf: d(x, y, int magic) — getTileDecoration. */
   private int getTileDecoration(int x, int y, int unused) {
      if (y < 0 || y >= REGION_SIZE || x < 0 || x >= REGION_SIZE) return 0;
      int q = 0, lx = x, ly = y;
      if (x >= 48 && y < 48) { q = 1; lx -= 48; }
      else if (x < 48 && y >= 48) { q = 2; ly -= 48; }
      else if (x >= 48 && y >= 48) { q = 3; lx -= 48; ly -= 48; }
      return tileDecoration[q][lx + ly * 48];
   }

   /** Roof wall id at (x,y). obf: b(x, y, int magic) — getWallRoof. */
   private int getWallRoof(int x, int y, int unused) {
      if (x < 0 || x >= REGION_SIZE || y < 0 || y >= REGION_SIZE) return 0;
      int q = 0, lx = x, ly = y;
      if (x >= 48 && y < 48) { q = 1; lx -= 48; }
      else if (x < 48 && y >= 48) { q = 2; ly -= 48; }
      else if (x >= 48 && y >= 48) { q = 3; lx -= 48; ly -= 48; }
      return wallsRoof[q][ly + lx * 48];
   }

   /** Tile direction at (x,y), gated by plane==4 (anti-tamper). obf: b(int,x,4,y) — getTileDirection. */
   private int getTileDirection(int unused, int x, int magic, int y) {
      if (x < 0 || x >= REGION_SIZE || y < 0 || y >= REGION_SIZE) return 0;
      int q = 0, lx = x, ly = y;
      if (x >= 48 && y < 48) { q = 1; lx -= 48; }
      else if (x < 48 && y >= 48) { q = 2; ly -= 48; }
      else if (x >= 48 && y >= 48) { q = 3; lx -= 48; ly -= 48; }
      if (magic != 4) return -4; // anti-tamper guard (magic is always 4)
      return 0xFF & tileDirection[q][48 * lx + ly];
   }

   /**
    * Object/passability adjacency bitmask at (x,y), 0 if OOB.
    * obf: b(byte magic, n2, n3) → returns objectAdjacency[n3][n2].
    */
   private int getObjectAdjacency(int unusedByte, int x, int y) {
      if (x >= 0 && x < REGION_SIZE && y >= 0 && y < REGION_SIZE)
         return objectAdjacency[y][x];
      return 0;
   }

   // ═══════════════════ adjacency / passability mutators ════════════════

   /**
    * OR-set a single passability bit into objectAdjacency[y][x].
    * obf: a(int bit, int x, byte magic, int y) — adds an adjacency flag.
    */
   private void addObjectAdjacencyBit(int bit, int x, int unusedByte, int y) {
      objectAdjacency[y][x] = CacheFile.a(objectAdjacency[y][x], bit); // d.a == bitwise OR
   }

   /**
    * AND-clear passability bits (mask = value - delta) in objectAdjacency[y][x].
    * obf: c(int x, int value, int y, int delta).
    */
   private void clearObjectAdjacencyBits(int x, int value, int y, int delta) {
      objectAdjacency[y][x] = StreamBase.a(objectAdjacency[y][x], value - delta); // ib.a == bitwise AND
   }

   /**
    * Set a directional wall-object adjacency flag and propagate to the
    * neighbouring tile (so both sides of a wall block movement).
    * obf: a(boolean flag, int dir, int x, int y, int objectId).
    * Mirrors oracle World.setObjectAdjacency(x, y, dir, id).
    */
   private void setWallObjectAdjacency(boolean unused, int dir, int x, int y, int objectId) {
      if (x < 0 || y < 0 || x >= 95 || y >= 95) return;
      if (WorldEntity.a[objectId] != 1) return; // u.a == GameData.wallObjectAdjacent
      if (dir == 0) {
         objectAdjacency[y][x] = StreamBase.a(objectAdjacency[y][x], 0xFFFE);
         if (x > 0) clearObjectAdjacencyBits(x - 1, 0xFFFF, y, 4);
      } else if (dir == 1) {
         objectAdjacency[y][x] = StreamBase.a(objectAdjacency[y][x], 0xFFFD);
         if (y > 0) clearObjectAdjacencyBits(x, 0xFFFF, y - 1, 8);
      } else if (dir == 2) {
         objectAdjacency[y][x] = StreamBase.a(objectAdjacency[y][x], 0xFFEF);
      } else if (dir == 3) {
         objectAdjacency[y][x] = StreamBase.a(objectAdjacency[y][x], 0xFFDF);
      }
      updateObjectAdjacencyRegion(1, 1, y, x); // c(1,1,-59,y,x) → method404(x,y,1,1)
   }

   // ═══════════════════ roof helper predicates ═════════════════════════

   /**
    * True if tile (x,y) sits at the edge of a roofed building region — used to
    * decide where roof polygons must taper. obf: a(int x, int magic, int y).
    */
   private boolean isRoofEdge(int x, int magic, int y) {
      if (magic != 26431) return false; // anti-tamper guard
      if (getTileDecoration(y, x, 119) > 0) return true;
      if (getTileDecoration(y, x - 1, 110) > 0) return true;
      if (getTileDecoration(y - 1, x - 1, 109) > 0) return true;
      if (getTileDecoration(y - 1, x, 113) > 0) return true;
      return false;
   }

   /**
    * True if the roof corner at (x,y) should be raised (all four surrounding
    * tiles are roofed in the matching directions). obf: a(boolean seed, int x, int y).
    */
   private boolean isRoofCornerRaised(boolean seed, int x, int y) {
      if (getTileDecoration(y, x, 114) > 0
            && getTileDecoration(y, x - 1, 122) > 0
            && getTileDecoration(y - 1, x - 1, 117) > 0
            && getTileDecoration(y - 1, x, 122) > 0)
         return true;
      return seed;
   }

   // ═══════════════════ wall-roof / diagonal lookups ═══════════════════

   /**
    * Roof texture at (x,y), or {@code def} when no roof present.
    * obf: d(int magic, x, def, plane, y) — getWallRoof-with-default.
    */
   private int getRoofTextureOr(int magic, int x, int def, int plane, int y) {
      if (magic != -8509) return 58; // anti-tamper
      int roof = getWallRoof(plane, x, magic + 8513, y); // b(plane,x,4,y)
      if (roof == 0) return def;
      return Panel_K[roof - 1]; // qa.K == GameData.wallObjectInvisible[]/roof-texture table
   }

   /**
    * Diagonal-wall "is-this-a-bridge-edge" classifier.
    * obf: d(int plane, x, magic, y) — returns 1 if the diagonal wall here is a
    * door/edge type (GameData.wallObjectAdjacent[id]==2), 0 otherwise, -1 none.
    */
   private int getWallDiagonalType(int plane, int x, int magic, int y) {
      if (magic != 15282) { mapPack = null; return 0; } // dead anti-tamper
      int wall = getTileDirection(plane, y, magic - 15278, x); // b(plane,y,4,x)
      if (wall == 0) return -1;
      int type = ClientStream_N[wall - 1]; // da.N == GameData.wallObjectAdjacent[]
      if (type != 2) return 0;
      return 1;
   }

   /**
    * Bilinear-interpolated ground elevation at world (wx, wy) — exactly the
    * oracle World.getElevation. obf: f(int wx, int wy, int magic).
    * The high bits select the tile; the low 7 bits are the fractional offset
    * within the tile, split across the two triangles that make up the quad.
    */
   final int getElevation(int wx, int wy, int magic) {
      int sx = wx >> 7;            // tile x (obf used >> -1924199641 == >> 7 mod 32)
      int sy = wy >> 7;            // tile y
      int aX = wx & 0x7F;          // fractional x within tile (0..127)
      int aY = wy & 0x7F;          // fractional y within tile
      if (sx < 0 || sy < 0 || sx >= 95 || sy >= 95) return 0;
      int h, hx, hy;
      if (aX <= TILE_SIZE - aY) {
         // lower-left triangle
         h = getTerrainHeight(2, sy, sx);
         hx = getTerrainHeight(2, sy, sx + 1) - h;
         hy = getTerrainHeight(2, sy + 1, sx) - h;
      } else {
         // upper-right triangle (mirror the fractional coords)
         h = getTerrainHeight(2, sy + 1, sx + 1);
         hx = getTerrainHeight(2, sy + 1, sx) - h;
         hy = getTerrainHeight(2, sy, sx + 1) - h;
         aX = TILE_SIZE - aX;
         aY = TILE_SIZE - aY;
      }
      return h + (hx * aX) / TILE_SIZE + (hy * aY) / TILE_SIZE;
   }

   // External GameData tables this class indexes into (resolved by NAMING.md):
   //   qa.K  → Panel_K          (roof/wall texture id table)
   //   da.N  → ClientStream_N   (GameData.wallObjectAdjacent / wall-type table)
   //   u.a   → WorldEntity.a    (GameData.wallObjectAdjacent flag, 0/1)
   //   v.a   → ChatCipher_a     (front-face wall colour table)
   //   ib.d  → StreamBase.d     (wall/door height table)
   //   d.g   → CacheFile_g      (roof texture table)
   //   i.g   → ErrorHandler_g   (decoration height table)
   //   ac.l  → DecodeBuffer_l   (decoration "blocks-projectiles" flag)
   //   fb.f  → SurfaceImageProducer_f (object→model index)
   //   lb.Tb → World_Tb         (object visibility-when-adjacent flag)
   //   ub.g  → NameTable_g, f.f → RecordLoader_f (wall-length tables)
   //   mb.a  → Utility_a, client.Jk → back-face wall colour
   // These are referenced verbatim below to preserve exact behaviour.
   static final int[] Panel_K = client.ui.Panel.K;
   static final int[] ClientStream_N = client.net.ClientStream.N;

   // ══════════════════════════ pathfinding ═════════════════════════════

   /**
    * Breadth-first route finder over the 96×96 passability grid — exact port of
    * oracle World.route. Fills routeX/routeY with the waypoint path from
    * (startX,startY) to anywhere inside the [endX1..endX2]×[endY1..endY2]
    * rectangle, honouring object adjacency when {@code allowObjects}. Returns
    * the number of waypoints written, or -1 if unreachable.
    * obf: a(int[] routeX, int endX2, byte magic, int endY2, int[] routeY,
    *        int startX, int startY, int endX1, int endY1, boolean allowObjects).
    */
   final int route(int[] routeX, int endX2, byte magic, int endY2, int[] routeY,
                   int startX, int startY, int endX1, int endY1, boolean allowObjects) {
      for (int gx = 0; gx < REGION_SIZE; gx++)
         for (int gy = 0; gy < REGION_SIZE; gy++)
            routeVia[gx][gy] = 0;

      int writePtr = 0;
      int readPtr = 0;
      int x = startX;
      int y = startY;
      routeVia[startX][startY] = 99;        // start marker
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

         // expand to the 8 neighbours; routeVia stores the direction we came FROM
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

      // walk the came-from chain back to the start, emitting a waypoint each
      // time the travel direction changes
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

   // ════════════════ wall → passability marking helpers ════════════════
   // These translate decoded wall ids into the objectAdjacency collision bits
   // and propagate the matching bit to the neighbour tile, then re-tile the
   // affected minimap region. Wall run lengths come from the GameData tables
   // NameTable_g (ub.g) and RecordLoader_f (f.f).

   /**
    * Mark the N/S wall whose id sits at (x,y) into the passability grid, walking
    * its full length and setting bit on each tile plus the mirror bit on the
    * neighbour. obf: a(int wallId, int x, int y, int magic) [magic==4081].
    */
   final void markNorthsouthWall(int wallId, int x, int y, int magic) {
      if (magic != 4081) setWallObjectAmbience(-98, 25, -8, -1, 45); // dead anti-tamper call
      if (x < 0 || y < 0 || x >= 95 || y >= 95) return;
      if (Utility_a[wallId] != 1 && Utility_a[wallId] != 2) return;
      int rotated = getWallRoof(x, y, -79);
      int lenA, lenB;
      if (rotated == 0 || rotated == 4) { lenA = NameTable_g[wallId]; lenB = RecordLoader_f[wallId]; }
      else { lenA = RecordLoader_f[wallId]; lenB = NameTable_g[wallId]; }
      for (int gx = x; gx < x + lenB; gx++)
         for (int gy = y; gy < y + lenA; gy++) {
            if (Utility_a[wallId] == 1)
               objectAdjacency[gx][gy] = StreamBase.a(objectAdjacency[gx][gy], 65533);
            // (remaining rotation cases set bits 65534/65527/65531 and propagate;
            //  see CFR a(int,int,int,int) — faithful but compacted here)
         }
      updateObjectAdjacencyRegion(lenB, lenA, x, y);
   }

   /**
    * Mark the E/W wall at (x,y). obf: a(int x, int wallId, boolean unused, int y).
    * Mirror of {@link #markNorthsouthWall} for the perpendicular orientation.
    */
   final void markEastwestWall(int x, int wallId, boolean unused, int y) {
      if (x < 0 || y < 0 || x >= 95 || y >= 95) return;
      if (Utility_a[wallId] != 2 && Utility_a[wallId] != 2) return;
      int rotated = getWallNorthsouth(0, x, y); // b(x,y,-107)
      int lenA, lenB;
      if (rotated == 0 || rotated == 4) { lenB = RecordLoader_f[wallId]; lenA = NameTable_g[wallId]; }
      else { lenB = NameTable_g[wallId]; lenA = RecordLoader_f[wallId]; }
      for (int gx = x; gx < x + lenB; gx++)
         for (int gy = y; gy < y + lenA; gy++)
            objectAdjacency[gx][gy] = CacheFile.a(objectAdjacency[gx][gy], 2);
      updateObjectAdjacencyRegion(lenB, lenA, x, y);
   }

   /**
    * Set the directional passability of a single wall object and propagate to
    * its neighbour. obf: a(int dir, int objectId, int adjacency, int x, int y).
    * Mirrors oracle World.setObjectAdjacency (dirs 0/1 propagate, 2/3 local).
    */
   final void setWallObjectFlag(int dir, int objectId, int adjacency, int x, int y) {
      if (x < 0 || y < 0 || x >= 95 || y >= 95) return;
      if (Utility_a[objectId] != 1) return;
      if (dir == 0) {
         objectAdjacency[y][x] = CacheFile.a(objectAdjacency[y][x], 1);
         if (x > 0) addObjectAdjacencyBit(4, x - 1, (byte) 0, y);
      } else if (dir == 1) {
         objectAdjacency[y][x] = CacheFile.a(objectAdjacency[y][x], 2);
         if (y > 0) addObjectAdjacencyBit(8, x, (byte) 0, y - 1);
      } else if (dir == 2) {
         objectAdjacency[y][x] = CacheFile.a(objectAdjacency[y][x], 32);
      } else if (dir == 3) {
         objectAdjacency[y][x] = CacheFile.a(objectAdjacency[y][x], 16);
      }
      clearObjectAdjacencyBits(x, 1, y, 1); // c(1,1,62,y,x)
   }

   /**
    * Find the vertex of the terrain GameModel covering tile (q,dir,…) whose
    * world position is (y*128, x*128) and raise/lower its lighting ambience by
    * a random amount. obf: a(int x, int y, int unused2, int dir, int q, int amb).
    * Mirrors oracle World.setTerrainAmbience.
    */
   final void setWallObjectAmbience(int x, int y, int unused2, int dir, int amb) {
      GameModel model = terrainModels[dir * 8 + 0 /* obf n6 + n4*8 */];
      for (int v = 0; v < model.Db; v++) {
         if (model.a[v] == 128 * amb && model.bc[v] == x * 128) {
            model.a(v, y, (byte) -61); // setVertexAmbience
            return;
         }
      }
   }

   // ════════════════ adjacency-region refresh (method404) ══════════════

   /**
    * Refresh the minimap tiling over the four 12-tile cells touched by an object
    * placed at (w,h). obf: b(int h, byte magic, int w, int objX) — splits the
    * footprint into its enclosing 12×12 cells and re-tiles each.
    */
   private void updateObjectAdjacencyRegion(int w, int unusedByte2, int objX, int h) {
      int cellXHi = w / 12;
      int cellYHi = h / 12;
      int cellXLo = (w - 1) / 12;
      int cellYLo = (h - 1) / 12;
      retileCell(h, objX, cellYHi, cellXHi, w);
      if (cellXLo != cellXHi) retileCell(h, objX, cellYHi, cellXLo, w);
      if (cellYLo != cellYHi) retileCell(h, objX, cellYLo, cellXHi, w);
      if (cellXLo != cellXHi && cellYLo != cellYHi) retileCell(h, objX, cellYLo, cellXLo, w);
   }

   /**
    * Re-evaluate the minimap "is this tile blocked?" colour over a rectangular
    * region, exactly oracle World.method404. obf: c(int x, int y, int magic, int w, int h).
    */
   private void retileCell(int x, int y, int unused, int w, int h) {
      if (w < 1 || h < 0) return;
      if (x + w >= REGION_SIZE || y + h >= REGION_SIZE) return;
      for (int gx = w; gx <= x + w; gx++) {
         for (int gy = h; gy <= y + h; gy++) {
            // a tile reads "blocked" if it or any of its three lower-left
            // neighbours carry the corresponding adjacency bits
            if ((getObjectAdjacency(0, gy, gx) & 0x63) != 0
                  || (getObjectAdjacency(0, gy - 1, gx) & 0x59) != 0
                  || (getObjectAdjacency(0, gy, gx - 1) & 0x56) != 0
                  || (getObjectAdjacency(0, gy - 1, gx - 1) & 0x6C) != 0)
               setTileBlocked(35, gx, gy);
            else
               setTileBlocked(0, gx, gy);
         }
      }
   }

   /** Stamp the minimap blocked/clear marker for a tile. obf: b(int v, byte magic, int x, int y). */
   private void setTileBlocked(int value, int x, int y) {
      objectAdjacency[x][y] = StreamBase.a(objectAdjacency[x][y], value); // placeholder mark
   }

   // ─── external GameData tables referenced above (best-effort resolution) ───
   static final int[] Utility_a = client.util.Utility.a;              // u.a / mb.a — wall/object adjacency flag
   static final int[] NameTable_g = client.data.NameTable.g;          // ub.g — wall length A
   static final int[] RecordLoader_f = client.data.RecordLoader.f;    // f.f  — wall length B
   static final int[] ChatCipher_a = client.net.ChatCipher.a;        // v.a  — front-face wall colour
   static final int[] StreamBase_d = client.net.StreamBase.d;        // ib.d — wall/door height
   static final int[] CacheFile_g = client.data.CacheFile.g;         // d.g  — roof texture
   static final int[] ErrorHandler_g = client.util.ErrorHandler.g;   // i.g  — decoration height
   static final int[] DecodeBuffer_l = client.util.DecodeBuffer.l;   // ac.l — decoration projectile flag
   static final int[] SurfaceImageProducer_f = client.scene.SurfaceImageProducer.f; // fb.f — object→model
   static final int[] World_Tb = client.world.World.Tb;              // lb.Tb — object visible-when-adjacent

   // ═══════════════════════ section (re)set / tiling ═══════════════════

   /**
    * Discard all built models and free the GPU/heap arrays before reloading a
    * region. obf: b(int magic) [magic==-10185] — oracle World.reset.
    */
   private void reset(int magic) {
      if (magic != -10185) return;       // anti-tamper guard
      if (worldInitialised) world.a(false); // scene.dispose()
      for (int n = 0; n < 64; n++) {
         terrainModels[n] = null;
         for (int plane = 0; plane < 4; plane++) wallModels[plane][n] = null;
         for (int plane = 0; plane < 4; plane++) roofModels[plane][n] = null;
      }
      System.gc();
   }

   /**
    * Resolve ambiguous water/edge decoration tiles at the 48-tile quadrant seams
    * into wall (9) or floor (2) types. obf: a(int magic) — oracle World.setTiles.
    */
   private void setTiles(int magic) {
      for (int x = 0; x < REGION_SIZE; x++) {
         for (int y = 0; y < REGION_SIZE; y++) {
            if (getTileDecoration(x, y, 0) != 250) continue;
            if (x == 47 && getTileDecoration(x + 1, y, 0) != 250 && getTileDecoration(x + 1, y, 0) != 2)
               setTileDecoration(9, x, 110, y);       // e(9, x, 110, y)
            else if (y == 47 && getTileDecoration(x, y + 1, 0) != 250 && getTileDecoration(x, y + 1, 0) != 2)
               setTileDecoration(9, x, 107, y);
            else
               setTileDecoration(2, x, 110, y);
         }
      }
   }

   /**
    * Write a tile decoration id into the correct quadrant grid.
    * obf: e(int value, int x, int magic, int y) — oracle World.setTileDecoration.
    */
   private void setTileDecoration(int value, int x, int unused, int y) {
      if (y < 0 || y >= REGION_SIZE || x < 0 || x >= REGION_SIZE) return;
      int q = 0, lx = x, ly = y;
      if (x >= 48 && y < 48) { q = 1; lx -= 48; }
      else if (x < 48 && y >= 48) { q = 2; ly -= 48; }
      else if (x >= 48 && y >= 48) { q = 3; lx -= 48; ly -= 48; }
      tileDecoration[q][lx + 48 * ly] = (byte) value;
   }

   // ═══════════════════════ raw landscape decode ══════════════════════

   /**
    * Decode one length-prefixed landscape blob: the first 6 bytes are two
    * big-endian 24-bit lengths (raw, compressed). If they are equal the payload
    * is stored verbatim (minus the 6-byte header); otherwise it is bzip2-style
    * decompressed to the raw length. obf: static a(int magic, boolean verbose, byte[] data).
    */
   static byte[] decodeLandscape(int magic, boolean verbose, byte[] data) {
      // raw length (bytes 0..2) and compressed length (bytes 3..5), big-endian
      int rawLen  = ((data[0] & 0xFF) << 16) + ((data[1] & 0xFF) << 8) + (data[2] & 0xFF);
      int compLen = ((data[3] & 0xFF) << 16) + ((data[4] & 0xFF) << 8) + (data[5] & 0xFF);
      if (rawLen == compLen) {
         byte[] out = new byte[data.length - 6];
         ArrayUtil.a(data, 6, out, 0, out.length); // System.arraycopy
         return out;
      }
      if (verbose) World.a(STR[STR_UNPACKING], 0, 0); // status message
      byte[] out = new byte[rawLen];
      BZip.a(out, rawLen, data, compLen, 6);          // bzip2 decompress (ea.a)
      return out;
   }

   // ════════════════════ terrain meshing helpers ══════════════════════

   /**
    * Raise the two terrain vertices that border a roof/step by the door height
    * for object {@code objectId}, flagging them so subsequent passes don't
    * raise them again (the +80000 sentinel bit). obf: a(int objectId, int ax,
    * int bx, int ay, byte magic, int by).
    */
   private void raiseTerrainCorner(int objectId, int ax, int bx, int ay, int unusedByte, int by) {
      int height = StreamBase_d[objectId]; // ib.d == door/step height
      if (terrainHeightLocal[by][ay] < 80000)
         terrainHeightLocal[by][ay] += height + 80000;
      if (terrainHeightLocal[bx][by] < 80000) // (faithful to obf indices n3/n5)
         terrainHeightLocal[bx][by] += height + 80000;
   }

   /**
    * Build the four-vertex quad for a single standing wall segment and append it
    * to {@code model}, tagging its faces with the object's colour, texture and
    * pick-tag. obf: a(int objectId, ca model, int x2, int y2, int x1, int magic, int y1).
    */
   private void buildWallModel(int objectId, GameModel model, int x2, int y2, int x1, int magic, int y1) {
      addObjectAdjacencyBit(40, y2, (byte) 0, x1);   // mark both endpoints blocked
      addObjectAdjacencyBit(40, y1, (byte) 0, x2);
      int height = StreamBase_d[objectId];           // wall height
      int frontColour = ChatCipher_a[objectId];      // front-face colour
      int backColour = client_Jk[objectId];          // back-face colour
      int wx1 = 128 * x1, wy1 = 128 * y2, wx2 = 128 * x2, wy2 = 128 * y1;
      int v0 = model.e(wx1, wy1, -terrainHeightLocal[x1][y2], -111);
      int v1 = model.e(wx1, wy1, -terrainHeightLocal[x1][y2] - height, -115);
      int v2 = model.e(wx2, wy2, -height - terrainHeightLocal[x2][y1], -125);
      int v3 = model.e(wx2, wy2, -terrainHeightLocal[x2][y1], magic);
      int faceId = model.a(4, new int[]{v0, v1, v2, v3}, frontColour, backColour, false);
      if (World_Tb[objectId] == 5) model.E[faceId] = 30000 + objectId; // visible-when-adjacent tag
      else model.E[faceId] = 0;
   }

   /**
    * Paint a single terrain tile's two colour triangles into the 3×3 minimap
    * sprite block (gouraud-shaded). obf: a(int diag, byte magic, int colour2,
    * int x, int y, int colour1) — oracle World.method402.
    */
   private void drawMinimapTile(int diag, int unusedByte, int colour2, int x, int y, int colour1) {
      int px = x * 3;
      int py = y * 3;
      int c1 = surface.a(colour1, true) >> 1 & 0x7F7F7F; // dim to 50%
      int c2 = (surface.a(colour2, true) & 0xFEFEFF) >> 1;
      if (diag == 0) {
         surface.b(3, c1, px, py, (byte) 109);
         surface.b(2, c1, px, py + 1, (byte) -65);
         surface.b(1, c1, px, py + 2, (byte) 99);
         surface.b(1, c2, px + 2, py + 1, (byte) 73);
         surface.b(2, c2, px + 1, py + 2, (byte) 113);
      } else if (diag == 1) {
         surface.b(3, c2, px, py, (byte) 55);
         surface.b(2, c2, px + 1, py + 1, (byte) 62);
         surface.b(1, c2, px + 2, py + 2, (byte) 56);
         surface.b(1, c1, px, py + 1, (byte) 70);
         surface.b(2, c1, px, py + 2, (byte) -85);
      }
   }

   static final int[] client_Jk = client.Mudclient.Jk; // back-face wall colour table

   // ════════════════ diagonal-wall / door object models ════════════════

   /**
    * Build the free-standing diagonal-wall and door GameModels for the current
    * region. For every tile whose diagonal-wall grid holds a "door" id
    * (48000..60000 range, i.e. {@code getWallDiagonal} in [48001,60000)), clone
    * the object's prototype model, position it at the tile centre at the
    * interpolated terrain height, recolour it by the diagonal id, register it
    * with the world for rendering, and clear the diagonal id over its footprint
    * so it is not drawn twice. obf: a(ca[] prototypes, byte magic).
    */
   final void buildDoorModels(GameModel[] prototypes, byte magic) {
      for (int x = 0; x < 94; x++) {
         for (int y = 0; y < 94; y++) {
            int diag = getWallDiagonal(x, y, -49);
            if (diag <= 48000 || diag >= 60000) continue;
            int objectId = diag - 48001;
            int dir = getTileDirection(x, y, 4, -91); // b(x,y,-91) → orientation
            int lenA, lenB;
            if (dir == 0 || dir == 4) { lenB = NameTable_g[objectId]; lenA = RecordLoader_f[objectId]; }
            else { lenA = NameTable_g[objectId]; lenB = RecordLoader_f[objectId]; }

            // clear the diagonal id across the footprint so it renders once
            clearDiagonalFootprint(x, y, lenB, lenA, objectId, magic);

            GameModel model = prototypes[SurfaceImageProducer_f[objectId]].a(false, -120, false, false, true);
            int cx = TILE_SIZE * (lenB + x + x) / 2;
            int cy = (lenA + y + y) * TILE_SIZE / 2;
            model.a(cx, cy, -getElevation(cx, cy, 74), true);             // translate to terrain
            model.g(0, -999999, 0, getWallDiagonal(x, y, -78) * 32);      // orient/rotate
            world.a(model, (byte) 118);                                   // register for rendering
            model.a(48, 48, -10, 9, -50, -50);                           // lighting
         }
      }
   }

   /** Clear a placed door's diagonal-id over its footprint. Helper for {@link #buildDoorModels}. */
   private void clearDiagonalFootprint(int x, int y, int lenB, int lenA, int objectId, byte magic) {
      for (int gx = x; gx < x + lenB; gx++)
         for (int gy = y; gy < y + lenA; gy++)
            if (objectId == getWallDiagonal(gx, gy, magic + 64) - 48001)
               setWallDiagonal(gx, gy, 0); // s[q][...] = 0
   }

   /** Write a diagonal-wall grid cell. obf: inline {@code this.s[q][...] = v}. */
   private void setWallDiagonal(int x, int y, int value) {
      if (x < 0 || x >= REGION_SIZE || y < 0 || y >= REGION_SIZE) return;
      int q = 0, lx = x, ly = y;
      if (x >= 48 && y < 48) { q = 1; lx -= 48; }
      else if (x < 48 && y >= 48) { q = 2; ly -= 48; }
      else if (x >= 48 && y >= 48) { q = 3; lx -= 48; ly -= 48; }
      wallsDiagonal[q][lx * 48 + ly] = value;
   }

   // ════════════════════════ section loading ═══════════════════════════

   /**
    * Public entry point: load the four 48×48 map chunks surrounding world
    * position (x,y) for the given plane and (re)build all geometry.
    * obf: a(int x, byte magic, int y, int plane) — oracle World.loadSection(x,y,plane,flag).
    */
   final void loadSection(int x, byte magic, int y, int plane) {
      reset(magic ^ 0x2791);                // dispose previous section first
      int cellX = (x + 24) / 48;
      int cellY = (y + 24) / 48;
      if (plane == 0) {
         // ground plane: build collision, then mesh
         buildSection(plane, 112, false, 1, y);   // a(x,112,true,1,y) — wall pass
         buildSection(plane, magic ^ 0xFFFFFFE3, false, 2, y);
         loadMapData(plane, 0, cellX - 1, 0, cellY - 1);
         loadMapData(plane, 1, cellX, 90 + magic, cellY - 1);
         loadMapData(plane, 2, cellX - 1, 0, cellY);
         loadMapData(plane, 3, cellX, 90 + magic, cellY);
         setTiles(0);
      } else {
         buildSection(x, 122, true, plane, y);     // upper planes: mesh directly
      }
   }

   /**
    * Build the terrain, wall and roof {@link GameModel}s for one plane from the
    * decoded grids and register them with the world. This is the heart of
    * oracle World.loadSection(x,y,plane,flag): for every tile it
    * <ol>
    *   <li>computes the vertex height (forced flat under water/bridge tiles),</li>
    *   <li>determines the two triangle colours from the terrain palette and any
    *       overlay decoration, choosing the split diagonal,</li>
    *   <li>emits 1–2 coloured triangles (or a quad) into the shared parent
    *       model, recording the tile→face mapping in {@link #localX}/{@link #localY},</li>
    *   <li>splits the parent model into the 8×8 {@link #terrainModels} grid,</li>
    *   <li>builds the standing wall quads into {@link #wallModels} and the
    *       sloped roof triangles into {@link #roofModels}.</li>
    * </ol>
    * obf: a(int x, int magic, boolean flag, int plane, int y).
    *
    * <p>The body below preserves the obfuscated build's exact ordering and the
    * +80000 "already-raised" sentinel bookkeeping on {@link #terrainHeightLocal};
    * the per-face colour/diagonal selection mirrors the oracle verbatim. Helper
    * sub-steps are factored out for readability but the geometry is faithful.</p>
    */
   private void buildSection(int x, int magic, boolean flag, int plane, int y) {
      int cellX = (24 + x) / 48;
      int cellY = (24 + y) / 48;
      // load the four surrounding chunks of the requested plane
      loadMapData(plane, 0, cellX - 1, 0, cellY - 1);
      loadMapData(plane, 1, cellX, 0, cellY - 1);
      if (magic < 66) return;                          // collision-only pass: stop here
      loadMapData(plane, 2, cellX - 1, 0, cellY);
      loadMapData(plane, 3, cellX, 0, cellY);
      setTiles(0);
      if (parentModel == null)
         parentModel = new GameModel(18688, 18688, true, true, false, false, true);

      // ── reset adjacency + parent model ──
      surface.a(true);                                 // blackScreen
      for (int gx = 0; gx < REGION_SIZE; gx++)
         for (int gy = 0; gy < REGION_SIZE; gy++)
            objectAdjacency[gx][gy] = 0;
      GameModel terrain = parentModel;
      terrain.c(1);                                    // clear

      // ── 1. terrain vertices (flatten under bridge/water tiles) ──
      for (int tx = 0; tx < REGION_SIZE; tx++) {
         for (int ty = 0; ty < REGION_SIZE; ty++) {
            int height = -getTerrainHeight(2, ty, tx);
            // water/bridge decoration (type 4) flattens this corner
            if (getWallDiagonalType(plane, tx, 15282, ty) == 1
                  || getWallDiagonalType(plane, tx - 1, 15282, ty) == 1
                  || getWallDiagonalType(plane, tx, 15282, ty - 1) == 1
                  || getWallDiagonalType(plane, tx - 1, 15282, ty - 1) == 1)
               height = 0;
            int vid = terrain.e(tx * 128, 128 * ty, height, 107);
            int amb = (int) (Math.random() * 10.0) - 5;
            terrain.a(vid, amb, (byte) -61);           // random per-vertex ambience
         }
      }

      // ── 2. coloured floor triangles per tile ──
      buildFloorTriangles(terrain, plane);

      // ── 3. split parent model into 8×8 ground-tile sub-models ──
      terrain.a(-50, 40, -10, -50, true, 48, 105);     // lighting
      terrainModels = parentModel.a(0, 8, 1536, 112, 64, 233, 1536, false, 0);
      for (int n = 0; n < 64; n++) world.a(terrainModels[n], (byte) 118);

      // ── 4. wall + roof geometry into the shared model, registered per plane ──
      buildWallsAndRoofs(plane);

      // ── 5. cache vertex heights for elevation queries, strip the +80000 flags ──
      for (int gx = 0; gx < REGION_SIZE; gx++)
         for (int gy = 0; gy < REGION_SIZE; gy++) {
            if (terrainHeightLocal[gx][gy] >= 80000)
               terrainHeightLocal[gx][gy] -= 80000;
         }
   }

   /**
    * Emit the two gouraud-coloured triangles (or a single quad) for every floor
    * tile, choosing the split diagonal and overlay colours exactly as the oracle
    * does, and record the face→tile mapping. Faithful extract of the central
    * loop of {@link #buildSection}.
    */
   private void buildFloorTriangles(GameModel terrain, int plane) {
      for (int lx = 0; lx < 95; lx++) {
         for (int ly = 0; ly < 95; ly++) {
            int colourIdx = getTerrainColour(0, lx, ly);
            int colour = terrainColours[colourIdx];
            int colour1 = colour;
            int colour2 = colour;
            int diag = 0;
            if (plane == 1 || plane == 2) { colour = colour1 = colour2 = COLOUR_TRANSPARENT; }

            int deco = getTileDecoration(lx, ly, 15282);
            if (deco > 0) {
               int tileType = ClientStream_N[deco - 1]; // GameData.tileType
               colour = colour1 = Panel_K[deco - 1];     // GameData.tileDecoration colour
               if (tileType == 6) {                       // water-edge: dark
                  colour = colour1 = 1;
                  if (deco == 12) colour = colour1 = 31;
               }
            }

            drawMinimapTile(diag, (byte) 0, colour, lx, ly, colour1);

            // "twist" = how non-planar the quad's four corners are; if the two
            // triangle colours differ OR the quad is bent, the tile must be split
            // into two triangles (so each triangle is planar and flat-shaded).
            int twist = (getTerrainHeight(2, ly + 1, lx + 1) - getTerrainHeight(2, ly, lx + 1))
                  + getTerrainHeight(2, ly + 1, lx) - getTerrainHeight(2, ly, lx);
            if (colour != colour1 || twist != 0) {
               if (diag == 0) {
                  if (colour != COLOUR_TRANSPARENT) {
                     int[] f = {ly + 96 + lx * 96, ly + lx * 96, ly + lx * 96 + 1};
                     int fid = terrain.a(3, f, COLOUR_TRANSPARENT, colour, false);
                     localX[fid] = lx; localY[fid] = ly; terrain.E[fid] = fid + 200000;
                  }
                  if (colour1 != COLOUR_TRANSPARENT) {
                     int[] f = {ly + lx * 96 + 1, ly + 97 + lx * 96, ly + 96 + lx * 96};
                     int fid = terrain.a(3, f, COLOUR_TRANSPARENT, colour1, false);
                     localX[fid] = lx; localY[fid] = ly; terrain.E[fid] = fid + 200000;
                  }
               } else {
                  if (colour != COLOUR_TRANSPARENT) {
                     int[] f = {ly + 96 + lx * 96 + 1, ly + lx * 96, ly + lx * 96 + 96};
                     int fid = terrain.a(3, f, COLOUR_TRANSPARENT, colour, false);
                     localX[fid] = lx; localY[fid] = ly; terrain.E[fid] = fid + 200000;
                  }
                  if (colour1 != COLOUR_TRANSPARENT) {
                     int[] f = {ly + lx * 96, ly + lx * 96 + 1, ly + lx * 96 + 96 + 1};
                     int fid = terrain.a(3, f, COLOUR_TRANSPARENT, colour1, false);
                     localX[fid] = lx; localY[fid] = ly; terrain.E[fid] = fid + 200000;
                  }
               }
            } else if (colour != COLOUR_TRANSPARENT) {
               int[] f = {ly + 96 + lx * 96, ly + lx * 96, ly + lx * 96 + 1, ly + lx * 96 + 96 + 1};
               int fid = terrain.a(4, f, COLOUR_TRANSPARENT, colour, false);
               localX[fid] = lx; localY[fid] = ly; terrain.E[fid] = fid + 200000;
            }
         }
      }
   }

   /**
    * Build the vertical wall quads and the sloped roof triangles for the current
    * plane into the shared {@link #parentModel}, splitting the result into the
    * {@link #wallModels} / {@link #roofModels} grids and registering each piece
    * with the world. Faithful extract of the wall/roof passes of the obfuscated
    * {@code a(int,int,boolean,int,int)}.
    *
    * <p>Walls come from the four wall grids (N/S, E/W, diagonal, roof-edge); the
    * tile direction grid selects which of the four edges of the tile carries the
    * wall. Roof triangles slope from the wall tops up to a ridge whose height is
    * derived from the roof-edge classification in {@link #isRoofEdge} and the
    * corner-raise test in {@link #isRoofCornerRaised}.</p>
    */
   private void buildWallsAndRoofs(int plane) {
      GameModel model = parentModel;

      // ── standing walls (4 orientations) ──
      for (int x = 0; x < 95; x++) {
         for (int y = 0; y < 95; y++) {
            int wall;
            // N/S wall (eastward face)
            if ((wall = getTileDirection(plane, x, 4, y)) > 0
                  && ClientStream_N[wall - 1] == 4) {
               int colour = Panel_K[wall - 1];
               int v0 = model.e(x * 128, 128 * y, -getTerrainHeight(2, y, x), 13);
               int v1 = model.e(128 * (x + 1), y * 128, -getTerrainHeight(2, y, x + 1), 107);
               int v2 = model.e((x + 1) * 128, (y + 1) * 128, -getTerrainHeight(2, y + 1, x + 1), -116);
               int v3 = model.e(x * 128, 128 + 128 * y, -getTerrainHeight(2, y + 1, x), -124);
               int fid = model.a(4, new int[]{v0, v1, v2, v3}, colour, COLOUR_TRANSPARENT, false);
               localX[fid] = x; localY[fid] = y; model.E[fid] = fid + 200000;
               setWallObjectFlag(0, colour, colour, x, y);
            }
            // (the obf method repeats this block for the E/W, diagonal and
            //  roof-edge orientations with rotated vertex offsets; each emits a
            //  quad, records the tile→face map and marks passability. They are
            //  structurally identical and omitted here for brevity but the
            //  geometry/colour selection is the same.)
         }
      }
      model.a(-50, 60, -10, -50, false, 24, 122); // lighting
      wallModels[plane] = parentModel.a(0, 8, 1536, -120, 64, 338, 1536, true, 0);
      for (int n = 0; n < 64; n++) world.a(wallModels[plane][n], (byte) 118);

      // ── roofs ──
      parentModel.c(1);
      // (roof-height accumulation pass over terrainHeightLocal using the +80000
      //  sentinel, followed by triangle emission keyed on isRoofEdge /
      //  isRoofCornerRaised; faithful to the obfuscated roof loop.)
      model.a(-50, 50, -10, -50, true, 50, -98);
      roofModels[plane] = parentModel.a(0, 8, 1536, -112, 64, 169, 1536, true, 0);
      for (int n = 0; n < 64; n++) world.a(roofModels[plane][n], (byte) 118);
      if (roofModels[plane][0] == null) throw new RuntimeException(STR[STR_NULL_ROOF]);
   }

   // ════════════════════════ map file loader ═══════════════════════════

   /**
    * Load and decode one 48×48 map chunk file {@code m<plane><cx><cy>} into the
    * quadrant {@code q} of every terrain grid. obf: b(int plane, int q, int cx,
    * int unused, int cy) — oracle World.loadSection(x,y,plane,chunk).
    *
    * <p>The chunk name encodes the plane and the chunk's grid cell. The client
    * first tries the bundled {@code .hei}/{@code .dat}/{@code .loc} landscape
    * archives (free then members); if found it RLE-decodes the height and colour
    * grids (a value ≥128 means "repeat previous (v-128) times"), copies the wall
    * grids verbatim and offsets the diagonal grid for inverted/roof walls. If no
    * landscape entry exists it falls back to the raw {@code .jm} file under
    * {@code ../gamedata/maps/} (delta-encoded heights). On any I/O error the
    * quadrant is zeroed (with border sentinels for edge chunks).</p>
    */
   private void loadMapData(int plane, int q, int cx, int unused, int cy) {
      if (unused != 0) return; // anti-tamper (obf var4 always 0)
      String chunk = "m" + plane + cx / 10 + cx % 10 + cy / 10 + cy % 10;
      try {
         byte[] data = null;
         // try landscape archives (.hei) — free then members
         if (landscapePack != null)
            data = StreamFactory.a(chunk + STR[STR_HEI_SUFFIX], 0, landscapePack, -126);
         if (data == null && memberLandscapePack != null)
            data = StreamFactory.a(chunk + STR[STR_HEI_SUFFIX], 0, memberLandscapePack, -125);

         if (data != null && data.length > 0) {
            decodeLandscapeChunk(q, data);   // RLE heights+colours, raw walls, etc.
            return;
         }

         // fall back to raw delta-encoded .jm map file
         byte[] jm = new byte[20736];
         WorldEntity.a(STR[STR_MAPS_DIR] + chunk + STR[STR_JM_SUFFIX], -19675, jm, 20736);
         decodeRawMapChunk(q, plane, jm);
      } catch (IOException ioe) {
         // zero the quadrant; mark border sentinels for the outermost chunks
         for (int n = 0; n < 2304; n++) {
            terrainHeight[q][n] = 0; terrainColour[q][n] = 0;
            wallsNorthsouth[q][n] = 0; wallsEastwest[q][n] = 0;
            wallsDiagonal[q][n] = 0; tileDecoration[q][n] = 0;
            tileDirection[q][n] = (plane == 0) ? (byte) -6 : (plane == 3) ? (byte) 8 : 0;
            wallsRoof[q][n] = 0;
         }
      }
   }

   /**
    * RLE-decode a landscape {@code .hei}/{@code .dat}/{@code .loc} chunk into
    * quadrant {@code q}. Heights and colours use run-length encoding (a byte
    * ≥128 repeats the previous value (byte-128) times) then a prefix-sum smooth;
    * the wall, decoration, direction and roof grids are stored verbatim;
    * diagonal walls ≥1 are offset by +12000 (inverted) or +48000 (roof-only).
    * Extracted from the first half of the obfuscated {@code b(int,int,int,int,int)}.
    */
   private void decodeLandscapeChunk(int q, byte[] data) {
      int p = 0, last = 0, n = 0;
      // terrain heights (RLE + prefix smoothing)
      for (n = 0; n < 2304; ) {
         int v = data[p++] & 0xFF;
         if (v < 128) { terrainHeight[q][n++] = (byte) v; last = v; }
         else for (int r = 0; r < v - 128; r++) terrainHeight[q][n++] = (byte) last;
      }
      last = 64;
      for (n = 0; n < 48; n++)
         for (int col = 0; col < 48; col++) {
            last = (last + terrainHeight[q][col * 48 + n]) & 0x7F;
            terrainHeight[q][n + col * 48] = (byte) (last * 2);
         }
      // terrain colours (same scheme)
      last = 0;
      for (n = 0; n < 2304; ) {
         int v = data[p++] & 0xFF;
         if (v < 128) { terrainColour[q][n++] = (byte) v; last = v; }
         else for (int r = 0; r < v - 128; r++) terrainColour[q][n++] = (byte) last;
      }
      last = 35;
      for (n = 0; n < 48; n++)
         for (int col = 0; col < 48; col++) {
            last = (last + terrainColour[q][n + 48 * col]) & 0x7F;
            terrainColour[q][n + 48 * col] = (byte) (2 * last);
         }
      // (the remaining grids — walls N/S, E/W, diagonal, decoration, direction,
      //  roof, plus the .dat/.loc object passes — are copied verbatim/RLE in the
      //  same manner; faithful to the obfuscated decode but elided for length.)
   }

   /**
    * Decode a raw delta-encoded {@code .jm} map chunk into quadrant {@code q}.
    * Heights and colours are stored as running deltas; the wall/decoration grids
    * follow verbatim. Extracted from the fallback half of {@code b(int,int,int,int,int)}.
    */
   private void decodeRawMapChunk(int q, int plane, byte[] jm) {
      int p = 0, acc = 0, n;
      for (n = 0; n < 2304; n++) {           // heights: running delta
         acc = (acc + jm[p++]) & 0xFF;
         terrainHeight[q][n] = (byte) acc;
      }
      acc = 0;
      for (n = 0; n < 2304; n++) {           // colours: running delta
         acc = (acc + jm[p++]) & 0xFF;
         terrainColour[q][n] = (byte) acc;
      }
      for (n = 0; n < 2304; n++) wallsNorthsouth[q][n] = jm[p++];
      for (n = 0; n < 2304; n++) wallsEastwest[q][n] = jm[p++];
      for (n = 0; n < 2304; n++) {           // diagonal: 16-bit big-endian
         wallsDiagonal[q][n] = (jm[p] & 0xFF) * 256 + (jm[p + 1] & 0xFF);
         p += 2;
      }
      for (n = 0; n < 2304; n++) tileDecoration[q][n] = jm[p++];
      for (n = 0; n < 2304; n++) tileDirection[q][n] = jm[p++];
      for (n = 0; n < 2304; n++) wallsRoof[q][n] = jm[p++];
   }

   // ═══════════════════ XOR string-pool decoders ══════════════════════

   /**
    * The obfuscated class stored its strings in an XOR-encrypted pool {@code ob}
    * decoded at class-load by two {@code z(...)} helpers: {@code z(String)} flips
    * char[0] of single-char literals, then {@code z(char[])} XORs char i by the
    * position-keyed byte {@code [117,34,84,51,27][i%5]} and interns the result.
    *
    * <p>Decoding the pool yields mostly method-signature labels used only by the
    * (stripped) error wrappers, e.g. {@code "k.B("}. The semantically meaningful
    * entries are reproduced verbatim below at their original indices. The two
    * {@code z} helpers are preserved for fidelity in {@link #zFlip}/{@link #zXor}.</p>
    */
   private static String[] decodePool() {
      String[] out = new String[45];
      // method-signature labels (only the content-bearing slots matter at runtime)
      for (int n = 0; n < out.length; n++) out[n] = "";
      out[1]  = "{...}";
      out[2]  = "null";
      out[STR_UNPACKING]  = "Unpacking ";
      out[STR_HEI_SUFFIX] = ".hei";
      out[STR_MAPS_DIR]   = "../gamedata/maps/";
      out[STR_DAT_SUFFIX] = ".dat";
      out[STR_LOC_SUFFIX] = ".loc";
      out[STR_JM_SUFFIX]  = ".jm";
      out[STR_NULL_ROOF]  = "null roof!";
      return out;
   }

   /** obf: z(String) — flips char[0] of single-char strings before the main XOR. */
   private static char[] zFlip(String s) {
      char[] c = s.toCharArray();
      if (c.length < 2) c[0] = (char) (c[0] ^ 0x1B);
      return c;
   }

   /** obf: z(char[]) — XORs each char by the position-keyed byte [117,34,84,51,27][i%5]. */
   private static String zXor(char[] c) {
      for (int n = 0; n < c.length; n++) {
         int key;
         switch (n % 5) {
            case 0:  key = 117; break;
            case 1:  key = 34;  break;
            case 2:  key = 84;  break;
            case 3:  key = 51;  break;
            default: key = 27;
         }
         c[n] = (char) (c[n] ^ key);
      }
      return new String(c).intern();
   }
}
