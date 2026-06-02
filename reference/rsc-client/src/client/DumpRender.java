package client;

import client.scene.GameModel;
import client.scene.Scene;
import client.scene.Surface;
import client.world.World;

import java.io.File;
import java.io.FileInputStream;
import java.io.FileWriter;
import java.io.IOException;
import java.lang.reflect.Field;
import java.lang.reflect.Method;
import java.util.Arrays;
import java.util.Base64;

/**
 * Headless DEOB-engine render leg for the RSC 3-way render diff
 * (docs/build/RENDER_DIFF_DESIGN.md §5, the long-deferred "DEOB engine" leg).
 *
 * <p>This is the TRUSTWORTHY ground-truth oracle: it renders an {@code rscdump/1}
 * L1 fixture through the COMPILED, READABLE deobfuscated rev-235 client classes
 * — {@link client.scene.Surface}, {@link client.scene.Scene},
 * {@link client.world.World} — using their real names and the AUTHENTIC camera /
 * clip / fog setup the live client uses (Mudclient.java:2334 setBounds,
 * 6624-6638 the in-game camera). Unlike the obfuscated-jar oracle
 * (rscplus/dumprender/DumpRenderer.java), this drives the deob directly, so the
 * camera-look fields are observable and the projection is exactly the spec.
 *
 * <p>Pipeline (mirrors the live Mudclient frame, minus the network/UI):
 * <ol>
 *   <li>Parse the fixture JSON (terrain grids, window, camera).</li>
 *   <li>Construct {@code Surface(W,H,1,null)} (Component=null ⇒ truly headless,
 *       no AWT), {@code Scene(surface,15000,15000,1000)}, {@code World(scene,surface)}.</li>
 *   <li>Set the AUTHENTIC clip/fog (clipFar3d/2d=2400, fogZDistance=2300,
 *       fogZFalloff=1) and the screen bounds via setBounds.</li>
 *   <li>Synthesize the generic wood-door wall-def tables (front/back colour
 *       method305(120,90,55)=-15719, height 192) for every wall id the fixture
 *       references — matching the Go syntheticFacts + jar DumpRenderer.</li>
 *   <li>Inject the dumped terrain grids into World's private byte[][] grids,
 *       placed so the host tile (fixture window-local size/2) lands at World
 *       grid-local (48,48) — the SAME centre the orsc dump path uses
 *       (windowCentreTile=48), so all engines frame identically.</li>
 *   <li>{@code World.loadSection} (loadMapData no-ops because landscapePack==null,
 *       so injected grids survive — same effect the jar harness gets by patching
 *       loadMapData to RETURN).</li>
 *   <li>{@code Scene.setCameraOrientation} with the AUTHENTIC arg order
 *       (camX=cx, camY=cz, eyeLen=distance, yaw=912, mode=-12349, pitch=rotation,
 *       eyeY=-elev, roll=0) and {@code Scene.render}.</li>
 *   <li>Read {@code Surface.pixels} int[] → PNG; walk {@code Scene.models} →
 *       {@code <base>.faces.json} in the rscdump-faces/1 schema renderdiff consumes.</li>
 * </ol>
 *
 * <p>Run headless on JDK 17:
 * <pre>
 *   java -Djava.awt.headless=true -cp /tmp/deob-run client.DumpRender FIXTURE.json OUTDIR [BASENAME]
 * </pre>
 */
public final class DumpRender {

    // The window-local tile the host (fixture self / window-centre) maps to, in
    // the World 96x96 grid. orsc's dump path centres the requested tile at
    // windowCentreTile = worldWindowTiles/2 = 48 (render/orsc/harness.go), so the
    // DEOB leg must place the same host tile at grid-local 48 to align centroids.
    static final int HOST_GRID_LOCAL = 48;

    // method305(r,g,b) = -1 - (r/8)*1024 - (g/8)*32 - b/8  (Scene flat-fill encoding).
    // The generic wood door leaf colour = method305(120,90,55) = -15719, matching
    // the Go WallColourWood + the jar DumpRenderer.
    static int method305(int r, int g, int b) {
        return -1 - (r / 8) * 1024 - (g / 8) * 32 - b / 8;
    }

    public static void main(String[] args) throws Exception {
        if (args.length < 2) {
            System.err.println("usage: java client.DumpRender FIXTURE.json OUTDIR [BASENAME]");
            System.exit(2);
        }
        String fixturePath = args[0];
        String outDir = args[1];
        String outBase = args.length > 2 ? args[2] : "deob";

        // ---- 1. parse fixture ----
        String json = new String(readAll(fixturePath));
        Json fx = Json.parse(json);
        Json cam = fx.obj("camera");
        Json win = fx.obj("window");
        Json ter = fx.obj("terrain");

        int W = cam.i("screenW"), H = cam.i("screenH");
        int gsize = ter.i("size");
        int pitch = cam.i("pitch"), yaw = cam.i("yaw"), roll = cam.i("roll");
        int distance = cam.i("distance"), viewDist = cam.i("viewDist");

        int baseX = win.i("baseX"), baseY = win.i("baseY");
        int hostWorldX, hostWorldY;
        if (fx.has("self") && !fx.isNull("self")) {
            Json self = fx.obj("self");
            hostWorldX = self.i("x");
            hostWorldY = self.i("y");
        } else {
            hostWorldX = baseX + gsize / 2;
            hostWorldY = baseY + gsize / 2;
        }
        int fixtureHostLocalX = hostWorldX - baseX, fixtureHostLocalY = hostWorldY - baseY;
        // Place fixture window-local (i,j) at grid-local (gridBase+i, gridBase+j) so
        // the host fixture-local tile lands at grid-local HOST_GRID_LOCAL.
        int gridBaseX = HOST_GRID_LOCAL - fixtureHostLocalX;
        int gridBaseY = HOST_GRID_LOCAL - fixtureHostLocalY;
        System.out.println("placement: host world (" + hostWorldX + "," + hostWorldY + ") fixture-local ("
                + fixtureHostLocalX + "," + fixtureHostLocalY + ") -> gridBase (" + gridBaseX + "," + gridBaseY + ")");

        byte[] elevation = b64(ter, "elevation", gsize);
        byte[] groundColour = b64(ter, "groundColour", gsize);
        byte[] overlay = b64(ter, "overlay", gsize);
        byte[] roofGrid = b64(ter, "roof", gsize);
        byte[] wallH = b64(ter, "wallH", gsize);
        byte[] wallV = b64(ter, "wallV", gsize);
        int[] wallDiag = iarr(ter, "wallDiag", gsize);

        System.out.println("fixture: " + gsize + "x" + gsize + " grid, screen " + W + "x" + H
                + ", pitch=" + pitch + " yaw=" + yaw + " dist=" + distance + " viewDist=" + viewDist);

        // ---- 2. construct Surface / Scene / World (headless, Component=null) ----
        Surface surface = new Surface(W, H, 1, null);
        Scene scene = new Scene(surface, 15000, 15000, 1000);
        // Allocate the texture-cache backing arrays (Mudclient.java:2672
        // allocateTextures(0,11,7,texCount)). The hand-authored fixture uses only
        // flat NEGATIVE fills (terrainColours / wood door = -15719), so no texels are
        // ever materialised; we only need the arrays non-null so fillColour's
        // minimap-stroke path (World.method402 -> Scene.fillColour) doesn't NPE.
        scene.allocateTextures(0, 11, 7, 64);
        // Define every texture slot with a trivial 64px grey palette so any POSITIVE
        // fill id that reaches Scene.fillColour (the minimap-stroke path in
        // World.method402, and any textured-terrain face a fixture might carry)
        // resolves to a flat colour instead of NPE-ing on a null palette. The
        // hand-authored fixture uses only negative flat fills, so this is purely a
        // safety net — never visible in the 3D viewport.
        for (int t = 0; t < 64; t++) {
            scene.defineTexture(t, (byte) 0, new int[]{0x808080}, 0, new byte[64 * 64]);
        }
        World world = new World(scene, surface);
        // Preserve our injected grids across loadSection: with no landscapePack,
        // World.loadMapData would otherwise fall through to the .jm fallback, throw
        // IOException, and its catch would CLOBBER the injected grids (stamping
        // tileDecoration=-6 -> 250 -> grey "no map" terrain). This flag makes
        // loadMapData a no-op, the readable equivalent of the jar oracle's ASM
        // patch of k.loadMapData -> RETURN.
        world.dumpGridsInjected = true;
        // World.baseMediaSprite defaults to 750; buildSection's minimap blit
        // (Surface.drawSpriteMinimap into slot baseMediaSprite-1) would write slot
        // 749, out of range for our 1-slot headless Surface. Point it at slot 0
        // (the only slot) so the 2D minimap capture is harmless — it never touches
        // the 3D framebuffer the diff reads. (The live client allocates ~800 media
        // sprite slots; we don't render the minimap.)
        world.baseMediaSprite = 1;
        System.out.println("constructed Surface/Scene/World");

        // ---- 3. authentic clip/fog + screen bounds (Mudclient.java:2334-2341, 6624-6628) ----
        // setBounds(clipY, ignored, width, clipX, baseY, viewDistance, baseX)
        scene.setBounds(H / 2, true, W, W / 2, H / 2, viewDist, W / 2);
        scene.clipFar3d = 2400;
        scene.clipFar2d = 2400;
        scene.fogZDistance = 2300;
        scene.fogZFalloff = 1;
        scene.setLight(-50, -10, true, -50);

        // ---- 4. synthesize wall-def tables (generic wood door, matches Go syntheticFacts) ----
        // The grids store wall ids ONE-BASED (id 0 -> stored byte 1); method422 indexes
        // the static tables by the ZERO-BASED objectId (= storedByte - 1). Fill all N
        // entries so any referenced wall id renders the leaf identically to orsc.
        int doorColour = method305(120, 90, 55); // -15719
        int N = 16;
        client.net.ChatCipher.scratchA = fill(N, doorColour);        // v.a   wall front colour
        client.Mudclient.Jk = fill(N, doorColour);                   // client.Jk  wall back colour
        client.net.StreamBase._deadIntArray_d = fill(N, 192);        // ib.d  wall height
        client.scene.Scene.diagScratch = fill(N, 0);                 // lb.Tb visible flag
        System.out.println("synthesized wall-def tables (door colour " + doorColour + ", all " + N + " ids)");

        // GameData lookup tables the terrain/wall/roof/scenery mesher dereferences.
        // The hand-authored fixture carries no decoration/roof/scenery, so these are
        // only indexed on the deco>0 / roof>0 / object paths (never taken here); they
        // just need to be non-null. Size 256 covers any 1-byte id. (Matches the jar
        // DumpRenderer, which likewise synthesizes the tables the obf mesher reads.)
        int M = 256;
        client.net.ClientStream.sharedIntArrayN = new int[M];   // da.N  tile type
        client.ui.Panel.texK = new int[M];                      // qa.K  deco colour
        client.util.DecodeBuffer.landscapeFaceFlags = new int[M]; // landscape face flags
        client.net.StringCodec.DEAD_INT_ARRAY = new int[M];     // u.a   wall adjacency
        client.util.Utility.sizedPoolCounts = new int[M];       // mb.a  object type
        client.data.NameTable.sortKeys = new int[M];            // ub.g  object width
        client.data.RecordLoader.intArray = new int[M];         // f.f   object height
        client.data.CacheFile.sharedScratch = new int[M];       // d.g   roof texture
        client.util.ErrorHandler.unusedIntTable = new int[M];   // i.g   roof ridge height

        // ---- 5. inject grids into World's private byte[][] grids ----
        byte[][] terrainHeight = (byte[][]) gf(world, "terrainHeight"); // L
        byte[][] terrainColour = (byte[][]) gf(world, "terrainColour"); // eb
        byte[][] wallsNorthsouth = (byte[][]) gf(world, "wallsNorthsouth"); // f
        byte[][] wallsEastwest = (byte[][]) gf(world, "wallsEastwest"); // P
        byte[][] wallsRoof = (byte[][]) gf(world, "wallsRoof"); // A
        byte[][] tileDecoration = (byte[][]) gf(world, "tileDecoration"); // R
        int[][] wallsDiagonal = (int[][]) gf(world, "wallsDiagonal"); // s

        for (int q = 0; q < 4; q++) {
            Arrays.fill(terrainHeight[q], (byte) 0);
            Arrays.fill(terrainColour[q], (byte) 0);
            Arrays.fill(wallsNorthsouth[q], (byte) 0);
            Arrays.fill(wallsEastwest[q], (byte) 0);
            Arrays.fill(wallsRoof[q], (byte) 0);
            Arrays.fill(tileDecoration[q], (byte) 0);
            Arrays.fill(wallsDiagonal[q], 0);
        }

        // Fixture grids are row-major [i*size + j] with i = column = worldX - baseX.
        for (int i = 0; i < gsize; i++) {
            for (int j = 0; j < gsize; j++) {
                int gi = i * gsize + j;
                int gx = gridBaseX + i, gy = gridBaseY + j;
                // getTerrainHeight reads L[q][48*x + y]; getTerrainColour reads eb[q][y + 48*x].
                putByte(terrainHeight, gx, gy, true, elevation[gi]);
                putByte(terrainColour, gx, gy, true, groundColour[gi]);
                // Fixture VerticalWall spans the tile's X edge -> obf wallsEastwest
                // P (method422(wall-1, lx+1,ly, lx,ly), endpoints differ in X), which
                // is also where orsc's getVerticalWall -> place(w-1, x+1,z, x,z) goes.
                // Fixture HorizontalWall spans Z -> obf wallsNorthsouth f. (The earlier
                // f<->P swap put the door on the wrong, perpendicular edge.)
                putByte(wallsEastwest, gx, gy, true, wallV[gi]);
                putByte(wallsNorthsouth, gx, gy, true, wallH[gi]);
                // getTileDecoration reads R[q][48*x + y].
                putByte(tileDecoration, gx, gy, true, overlay[gi]);
                // getWallRoof reads A[q][x + y*48] with a SWAPPED quadrant rule.
                putRoof(wallsRoof, gx, gy, roofGrid[gi]);
                putInt(wallsDiagonal, gx, gy, wallDiag[gi]);
            }
        }
        System.out.println("injected grids at gridBase (" + gridBaseX + "," + gridBaseY + ")");

        // ---- 6. build the section (loadMapData no-ops: landscapePack==null) ----
        world.loadSection(HOST_GRID_LOCAL, (byte) -90, HOST_GRID_LOCAL, win.i("plane"));
        int modelCount = ((Integer) gf(scene, "modelCount")).intValue();
        System.out.println("loadSection OK; scene model count=" + modelCount);

        // ---- 7. camera (AUTHENTIC arg order, Mudclient.java:6637) + render ----
        // setCameraOrientation(camX, camY, eyeLen, yaw, mode, pitch, eyeY, roll)
        // Live call: setCameraOrientation(cx, cz, 2*ac, 912, -12349, ug*4, -elev, 0)
        //   => yaw arg carries the FIXED 912 downward tilt; pitch arg carries the
        //      orbit rotation ug*4 (the fixture's camera.yaw = orbit rotation = 512).
        int cx = HOST_GRID_LOCAL * 128 + 64; // 6208
        int cz = HOST_GRID_LOCAL * 128 + 64; // 6208
        int elev = world.getElevation(cx, cz);
        scene.setCameraOrientation(cx, cz, distance, pitch, -12349, yaw, -elev, roll);
        System.out.println("camera: center=(" + cx + "," + cz + ") eyeLen=" + distance
                + " yaw(arg)=" + pitch + " pitch(arg)=" + yaw + " eyeY=" + (-elev)
                + " -> lookX=" + gf(scene, "cameraLookX") + " lookY=" + gf(scene, "cameraLookY")
                + " lookZ=" + gf(scene, "cameraLookZ"));
        scene.render(-113);

        int[] pix = surface.pixels;
        int nz = 0;
        for (int p : pix) if ((p & 0xFFFFFF) != 0) nz++;
        System.out.println("render OK: nonzero px=" + nz + "/" + pix.length);

        // ---- 8a. PNG ----
        new File(outDir).mkdirs();
        java.awt.image.BufferedImage img =
                new java.awt.image.BufferedImage(W, H, java.awt.image.BufferedImage.TYPE_INT_RGB);
        img.setRGB(0, 0, W, H, pix, 0, W);
        File png = new File(outDir, outBase + ".png");
        javax.imageio.ImageIO.write(img, "png", png);
        System.out.println("wrote " + png + " (" + png.length() + " bytes)");

        // ---- 8b. faces.json (rscdump-faces/1) ----
        writeFaces(scene, modelCount, new File(outDir, outBase + ".faces.json"), doorColour);
    }

    static void writeFaces(Scene scene, int modelCount, File out, int doorColour) throws Exception {
        GameModel[] models = (GameModel[]) gf(scene, "models");
        int grid = 16; // render.CentroidGrid
        StringBuilder sb = new StringBuilder();
        sb.append("{\n  \"schema\": \"rscdump-faces/1\",\n  \"faces\": [\n");
        int faceTotal = 0, doorFaces = 0;
        boolean first = true;
        for (int mi = 0; mi < modelCount; mi++) {
            GameModel m = models[mi];
            if (m == null || m.numFaces == 0) continue;
            for (int fi = 0; fi < m.numFaces; fi++) {
                int[] fv = m.faceVertices[fi];
                if (fv == null || fv.length < 3) continue;
                long sx = 0, sy = 0, sz = 0;
                for (int vi : fv) {
                    sx += m.vertexX[vi];
                    sy += m.vertexY[vi];
                    sz += m.vertexZ[vi];
                }
                int n = fv.length;
                int cX = roundTo((int) (sx / n), grid);
                int cY = roundTo((int) (sy / n), grid);
                int cZ = roundTo((int) (sz / n), grid);
                int ff = m.faceFillFront[fi], fb = m.faceFillBack[fi];
                if (ff == doorColour || fb == doorColour) doorFaces++;
                if (!first) sb.append(",\n");
                first = false;
                sb.append("    { \"model\": ").append(mi)
                        .append(", \"centroid\": [").append(cX).append(", ").append(cY).append(", ").append(cZ).append("]")
                        .append(", \"numVerts\": ").append(n)
                        .append(", \"fillFront\": ").append(ff)
                        .append(", \"fillBack\": ").append(fb).append(" }");
                faceTotal++;
            }
        }
        sb.append("\n  ]\n}\n");
        try (FileWriter fw = new FileWriter(out)) {
            fw.write(sb.toString());
        }
        System.out.println("wrote " + out + " (" + out.length() + " bytes), faces=" + faceTotal + ", doorFaces=" + doorFaces);
    }

    // round v to nearest multiple of grid, symmetric for negatives (render.roundTo)
    static int roundTo(int v, int grid) {
        if (grid <= 1) return v;
        if (v >= 0) return ((v + grid / 2) / grid) * grid;
        return -(((-v + grid / 2) / grid) * grid);
    }

    static int[] fill(int n, int v) {
        int[] r = new int[n];
        Arrays.fill(r, v);
        return r;
    }

    // ---- grid quadrant writes mirroring World's accessors ----
    // Standard accessors (height/colour/walls/decoration) resolve quadrant the
    // same way; the index expression differs but the quadrant split is identical,
    // so a single helper writing the LOCAL (x,y) into the quadrant cell works for
    // all of them as long as we use the same index expression the reader uses.
    static void putByte(byte[][] g, int x, int y, boolean colourIndex, byte v) {
        int q = quad(x, y);
        int[] xy = qxy(x, y);
        // L uses 48*x+y, eb uses y+48*x (same), f/P use x*48+y, R uses 48*x+y — all
        // are 48*x + y. So one index expression covers them.
        g[q][48 * xy[0] + xy[1]] = v;
    }

    // getWallRoof uses a SWAPPED quadrant rule (q=1 when y>=48&&x<48) and indexes
    // A[q][x + y*48].
    static void putRoof(byte[][] g, int x, int y, byte v) {
        int qx = x, qy = y;
        int q = 0;
        if (qy >= 48 && qx < 48) { q = 1; qy -= 48; }
        else if (qy < 48 && qx >= 48) { q = 2; qx -= 48; }
        else if (qy >= 48 && qx >= 48) { q = 3; qx -= 48; qy -= 48; }
        g[q][qx + qy * 48] = v;
    }

    static void putInt(int[][] g, int x, int y, int v) {
        int q = quad(x, y);
        int[] xy = qxy(x, y);
        g[q][48 * xy[0] + xy[1]] = v;
    }

    static int quad(int x, int y) {
        if (x >= 48 && y < 48) return 1;
        if (x < 48 && y >= 48) return 2;
        if (x >= 48 && y >= 48) return 3;
        return 0;
    }

    static int[] qxy(int x, int y) {
        if (x >= 48 && y < 48) return new int[]{x - 48, y};
        if (x < 48 && y >= 48) return new int[]{x, y - 48};
        if (x >= 48 && y >= 48) return new int[]{x - 48, y - 48};
        return new int[]{x, y};
    }

    // ---- reflection helper for the few private fields we read ----
    static Object gf(Object o, String n) throws Exception {
        for (Class<?> k = o.getClass(); k != null; k = k.getSuperclass()) {
            try {
                Field f = k.getDeclaredField(n);
                f.setAccessible(true);
                return f.get(o);
            } catch (NoSuchFieldException ignore) {
            }
        }
        throw new NoSuchFieldException(n);
    }

    // ---- JSON field decoders ----
    static byte[] readAll(String path) throws IOException {
        try (FileInputStream in = new FileInputStream(path)) {
            return in.readAllBytes();
        }
    }

    static byte[] b64(Json ter, String key, int size) {
        int n = size * size;
        if (!ter.has(key) || ter.isNull(key)) return new byte[n];
        String s = ter.str(key);
        byte[] d = Base64.getDecoder().decode(s);
        if (d.length != n) {
            byte[] r = new byte[n];
            System.arraycopy(d, 0, r, 0, Math.min(n, d.length));
            return r;
        }
        return d;
    }

    static int[] iarr(Json ter, String key, int size) {
        int n = size * size;
        int[] r = new int[n];
        if (!ter.has(key) || ter.isNull(key)) return r;
        java.util.List<Object> a = ter.arr(key);
        for (int i = 0; i < Math.min(n, a.size()); i++) r[i] = ((Number) a.get(i)).intValue();
        return r;
    }
}
