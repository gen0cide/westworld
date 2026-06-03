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

    // ── Phase-0 NPC/player 2D-sprite placement-sanity spec (the SHARED static-entity
    // spec, docs/build/NPC_SPRITE_PARITY_PLAN.md tool #3). The rat (content0 serverId
    // 19) billboard is 346x136 (entityIndexTableC[19]/legacyMaskTable[19]); all three
    // legs hardcode the SAME size for Phase 0 (Phase 1 loads real per-id sizes). The
    // debug fill is a solid cyan (0x00FFFF), distinct from every terrain/door colour,
    // matching orsc's debugBillboardColour so the projected rect is byte-identical.
    static final int RAT_BILLBOARD_W = 346;
    static final int RAT_BILLBOARD_H = 136;
    static final int DEBUG_BILLBOARD_COLOUR = 0x0000FFFF;

    // ── Phase-1 NPC-Rat entity spec (docs/build/NPC_SPRITE_PARITY_PLAN.md §"First
    // test entity"). The rat is content0 serverId 19, animID 123. Its sprite block
    // base is 837 (the 124th unique-named 27-stride block; Phase-0.5 IndexAssertion
    // PROVES spriteOffsets[123]==837). Frame 0 (dir 0/step 0) is the standing south
    // billboard. charColour 4805259 is a RAW 24-bit value, NOT a 1/2/3 dye marker, so
    // the dye-remap path is identity and colour1 is used verbatim; skin colour is 0
    // (white = single-tint fast path). hasF=false (no +15 F-frame).
    static final String RAT_ANIM_NAME = "rat";
    static final int RAT_SPRITE_OFFSET = 837;     // li slot of the rat frame-0 block
    static final int RAT_FRAME = 0;               // dir 0/step 0 => sf[0]+3*0 = 0
    static final int RAT_CHAR_COLOUR = 4805259;   // raw, no dye (animationCharacterColour[123])
    static final int RAT_SKIN_COLOUR = 0;         // 0 => white single-tint fast path

    // entityGateEngaged: the entity layer is active when RSC_MESH_NPC is set (mirrors
    // RSC_MESH_REALDEFS). When unset, terrain/scenery fixtures render exactly as before.
    static boolean entityGateEngaged() {
        String g = System.getenv("RSC_MESH_NPC");
        return g != null && !g.trim().isEmpty();
    }

    // method305(r,g,b) = -1 - (r/8)*1024 - (g/8)*32 - b/8  (Scene flat-fill encoding).
    // The generic wood door leaf colour = method305(120,90,55) = -15719, matching
    // the Go WallColourWood + the jar DumpRenderer.
    static int method305(int r, int g, int b) {
        return -1 - (r / 8) * 1024 - (g / 8) * 32 - b / 8;
    }

    // ── synthesized GameData overlay/roof def tables ──────────────────────────
    // The 25 GroundOverlay (tile-decoration) defs, VERBATIM from the orsc engine's
    // render/orsc/world.go tileDefs[] ({colour, tileType}, indexed by id-1). The
    // colour follows the face-fill convention (>=0 texture id, <0 flat 5:5:5 fill,
    // TRANSPARENT=12345678 the bridge sentinel). Both DEOB tables and the jar
    // DumpRenderer's obf tables are filled from THIS list so the three engines'
    // overlay flatten / neighbour-spread / type-2 split agree byte-for-byte.
    static final int TRANSPARENT = 12345678;
    // {colour, tileType} per overlay id 1..25 (== orsc tileDefs[id-1]).
    static final int[][] OVERLAY_DEFS = {
        {-16913, 1}, {1, 3}, {3, 2}, {3, 4}, {-16913, 2}, {-27685, 2}, {25, 3},
        {TRANSPARENT, 5}, {-26426, 1}, {-1, 5}, {31, 3}, {3, 4}, {-4534, 2},
        {32, 2}, {-9225, 2}, {-3172, 2}, {15, 2}, {-2, 2}, {-1, 3}, {-2, 4},
        {-2, 4}, {-2, 0}, {-17793, 2}, {-14594, 1}, {1, 3},
    };
    // The 6 roof (elevation) defs {rise, tex}, VERBATIM from orsc roofs.go
    // authenticElevationDefs, indexed by roofId-1.
    static final int[][] ROOF_DEFS = {
        {64, 6}, {64, 3}, {96, 2}, {80, 33}, {80, 15}, {90, 49},
    };

    static int[] overlayColourTable(int n) {
        int[] r = new int[n];
        for (int i = 0; i < OVERLAY_DEFS.length && i < n; i++) r[i] = OVERLAY_DEFS[i][0];
        return r;
    }
    static int[] overlayTypeTable(int n) {
        int[] r = new int[n];
        for (int i = 0; i < OVERLAY_DEFS.length && i < n; i++) r[i] = OVERLAY_DEFS[i][1];
        return r;
    }
    static int[] roofRiseTable(int n) {
        int[] r = new int[n];
        for (int i = 0; i < ROOF_DEFS.length && i < n; i++) r[i] = ROOF_DEFS[i][0];
        return r;
    }
    static int[] roofTextureTable(int n) {
        int[] r = new int[n];
        for (int i = 0; i < ROOF_DEFS.length && i < n; i++) r[i] = ROOF_DEFS[i][1];
        return r;
    }

    // The 55-entry rev-235 texture id -> {base, sub} binding, VERBATIM from OpenRSC
    // EntityHandler.loadTextureDefinitions (EntityHandler.java:252-306). Index == the
    // texture id a face fill references. base names the `<base>.dat` entry in
    // content11; a non-empty sub names a `<sub>.dat` overlay blended over the base.
    // The orsc (authenticTextureNames) + JAR (TEXTURE_NAMES) legs bake the SAME list.
    static final String[][] TEXTURE_NAMES = {
        {"wall", "door"}, {"water", ""}, {"wall", ""}, {"planks", ""},
        {"wall", "doorway"}, {"wall", "window"}, {"roof", ""}, {"wall", "arrowslit"},
        {"leafytree", ""}, {"treestump", ""}, {"fence", ""}, {"mossy", ""},
        {"railings", ""}, {"painting1", ""}, {"painting2", ""}, {"marble", ""},
        {"deadtree", ""}, {"fountain", ""}, {"wall", "stainedglass"}, {"target", ""},
        {"books", ""}, {"timbered", ""}, {"timbered", "timberwindow"}, {"mossybricks", ""},
        {"growingwheat", ""}, {"gungywater", ""}, {"web", ""}, {"wall", "desertwindow"},
        {"wall", "crumbled"}, {"cavern", ""}, {"cavern2", ""}, {"lava", ""},
        {"pentagram", ""}, {"mapletree", ""}, {"yewtree", ""}, {"helmet", ""},
        {"canvas", "tentbottom"}, {"Chainmail2", ""}, {"mummy", ""}, {"jungleleaf", ""},
        {"jungleleaf3", ""}, {"jungleleaf4", ""}, {"jungleleaf5", ""}, {"jungleleaf6", ""},
        {"mossybricks", "arrowslit"}, {"planks", "window"}, {"planks", "junglewindow"}, {"cargonet", ""},
        {"bark", ""}, {"canvas", ""}, {"canvas", "tentdoor"}, {"wall", "lowcrumbled"},
        {"cavern", "crumbled"}, {"cavern2", "crumbled"}, {"lava", "flames"},
    };

    // Scratch sprite slots for the texture decode (must be < the Surface spriteSlots,
    // and clear of the minimap-blit sink slot 0 = baseMediaSprite-1).
    static final int TEX_SCRATCH = 64; // li.Eh — parseSprite target / composite box
    static final int TEX_DST = 65;     // li.ij+i — captured texture sprite

    // loadTextures is the headless port of Mudclient.loadTextures (Mudclient.java:2658):
    // load content11 the SAME way the mesh block loads content9, then for each of the
    // 55 names parse the base (+ optional subname overlay) sprite, composite onto a
    // magenta 128x128 box, chroma-key green->magenta, re-quantise (drawWorld), and
    // register via Scene.defineTexture(id, 74, palette, size/64-1, indices). A
    // missing content11 pack leaves the slots at the allocateTextures default (flat
    // skip), so a fixture run without the cache still renders (geometry only).
    static void loadTextures(Scene scene, Surface surface) throws Exception {
        String cacheDir = System.getenv("RSC_MESH_CACHE");
        if (cacheDir == null || cacheDir.isEmpty()) cacheDir = "/tmp/rsc-run/cache";
        File pack;
        try {
            pack = findContentPack(cacheDir, "content11_");
        } catch (IOException noPack) {
            System.out.println("loadTextures: no content11 pack (" + noPack.getMessage() + "); textures left flat");
            return;
        }
        byte[] raw = readAll(pack.getPath());
        byte[] arc = World.unpackData(128, false, raw); // outer header strip + bzip inflate
        byte[] index = client.util.StreamFactory.lookupEntityDefRecord("index.dat", 0, arc);

        int loaded = 0;
        for (int id = 0; id < TEXTURE_NAMES.length; id++) {
            String base = TEXTURE_NAMES[id][0];
            String sub = TEXTURE_NAMES[id][1];
            byte[] baseEntry = client.util.StreamFactory.lookupEntityDefRecord(base + ".dat", 0, arc);
            if (baseEntry == null) continue;

            // parse -> magenta box -> blit base (Mudclient.java:2681-2683).
            surface.parseSprite(TEX_SCRATCH, 1, baseEntry, 88, index);
            surface.drawBox(0, (byte) -117, 0xFF00FF, 0, 128, 128);
            surface.drawSprite(-1, TEX_SCRATCH, 0, 0);
            int size = surface.spriteWidthFull[TEX_SCRATCH];

            // optional subname overlay (Mudclient.java:2688-2693).
            if (sub != null && sub.length() > 0) {
                byte[] subEntry = client.util.StreamFactory.lookupEntityDefRecord(sub + ".dat", 0, arc);
                if (subEntry != null) {
                    surface.parseSprite(TEX_SCRATCH, 1, subEntry, 109, index);
                    surface.drawSprite(-1, TEX_SCRATCH, 0, 0);
                }
            }

            // capture the size×size box into TEX_DST (Mudclient.java:2696).
            surface.drawSprite(TEX_DST, size, 113, size, 0, 0);

            // chroma-key fix: green 0x00FF00 -> magenta 0xFF00FF (Mudclient.java:2701-2705).
            int sizeSq = size * size;
            for (int px = 0; px < sizeSq; px++) {
                if (surface.spritePixels[TEX_DST][px] == 0x00FF00) {
                    surface.spritePixels[TEX_DST][px] = 0xFF00FF;
                }
            }

            // re-quantise to a 256-colour palette (Mudclient.java:2707).
            surface.drawWorld(false, TEX_DST);

            // register the texture with the Scene (Mudclient.java:2710-2715).
            scene.defineTexture(
                id, (byte) 74,
                surface.spritePalette[TEX_DST],
                size / 64 - 1,
                surface.spriteColourIndex[TEX_DST]);
            loaded++;
        }
        System.out.println("loadTextures: loaded " + loaded + "/" + TEXTURE_NAMES.length + " content11 textures");
    }

    // loadRatSprite decodes the NPC-Rat body sprite block (15 frames) from the
    // AUTHENTIC content1 "people and monsters" archive into the Surface sprite slots
    // starting at RAT_SPRITE_OFFSET (837) — exactly what Mudclient.loadEntitySprites
    // (Mudclient.java:2582) does for idx 123: `li.parseSprite(uc=837, 15, ratBodyData,
    // 83, index)`. content1 is loaded the SAME way content9/content11 are (read the
    // cache file -> World.unpackData strips the 6-byte header + bzip-inflates the JAG
    // archive), then StreamFactory.lookupEntityDefRecord pulls "rat.dat" + "index.dat".
    // Uses the REAL multi-frame Surface.parseSprite (which already retains fullWidth/
    // fullHeight and the per-frame trim) — the deob leg needs no plain-Java reimpl.
    // Returns true if the block decoded (frame 0 has a non-empty bitmap).
    static boolean loadRatSprite(Surface surface) throws Exception {
        String cacheDir = System.getenv("RSC_MESH_CACHE");
        if (cacheDir == null || cacheDir.isEmpty()) cacheDir = "/tmp/rsc-run/cache";
        File pack;
        try {
            pack = findContentPack(cacheDir, "content1_");
        } catch (IOException noPack) {
            System.out.println("loadRatSprite: no content1 pack (" + noPack.getMessage() + ")");
            return false;
        }
        byte[] raw = readAll(pack.getPath());
        byte[] arc = World.unpackData(128, false, raw); // outer header strip + bzip inflate
        byte[] index = client.util.StreamFactory.lookupEntityDefRecord("index.dat", 0, arc);
        byte[] body = client.util.StreamFactory.lookupEntityDefRecord(RAT_ANIM_NAME + ".dat", 0, arc);
        if (index == null || body == null) {
            System.out.println("loadRatSprite: missing index.dat or " + RAT_ANIM_NAME + ".dat in content1");
            return false;
        }
        // 15 body frames, the SAME stride the live client uses (the +27 block is
        // 15 body + 3 'a' + 9 'f'; the rat has no a/f, so frames 837..851).
        surface.parseSprite(RAT_SPRITE_OFFSET, 15, body, 83, index);
        int slot = RAT_SPRITE_OFFSET + RAT_FRAME;
        int sw = surface.spriteWidth[slot];
        int sh = surface.spriteHeight[slot];
        int fw = surface.spriteWidthFull[slot];
        int fh = surface.spriteHeightFull[slot];
        int tx = ((int[]) gf(surface, "spriteTranslateX"))[slot];
        int ty = ((int[]) gf(surface, "spriteTranslateY"))[slot];
        System.out.println("loadRatSprite: decoded '" + RAT_ANIM_NAME + "' frame " + RAT_FRAME
                + " into slot " + slot
                + " trimmed=" + sw + "x" + sh + " full=" + fw + "x" + fh
                + " trans=(" + tx + "," + ty + ")");
        return sw > 0 && sh > 0;
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
        // 66 sprite slots: slot 0 = the minimap-blit sink (baseMediaSprite-1, harmless);
        // slots TEX_SCRATCH(64)/TEX_DST(65) host the content11 texture-decode pipeline
        // (loadTextures, below). The live client allocates ~3000 slots; we need only
        // these few for the headless texture load + the minimap no-op. When the entity
        // gate is engaged we must also reach the rat block at slot 837 (RAT_SPRITE_OFFSET
        // + the 27-slot stride), so widen the bank to cover it.
        int spriteSlots = entityGateEngaged() ? (RAT_SPRITE_OFFSET + 27) : 66;
        Surface surface = new Surface(W, H, spriteSlots, null);
        Scene scene = new Scene(surface, 15000, 15000, 1000);
        // Allocate the texture-cache backing arrays (Mudclient.java:2672
        // allocateTextures(0,11,7,texCount)). The hand-authored fixture uses only
        // flat NEGATIVE fills (terrainColours / wood door = -15719), so no texels are
        // ever materialised; we only need the arrays non-null so fillColour's
        // minimap-stroke path (World.method402 -> Scene.fillColour) doesn't NPE.
        scene.allocateTextures(0, 11, 7, 64);
        // Load the AUTHENTIC rev-235 "Textures" content archive (content11) into the
        // Scene's texture slots — the SAME texel banks orsc (content11) + the JAR
        // (content11) load — so TEXTURED scenery faces (e.g. the well's 18 wall/planks
        // faces, texIds 2/3) and textured terrain overlays render with their real
        // texels instead of the old 64-magenta-transparent stub that skipped every
        // textured face. The decode pipeline (parseSprite -> magenta box -> drawSprite
        // -> chroma-key -> drawWorld -> defineTexture) is mudclient.loadTextures
        // (Mudclient.java:2658-2716); it re-quantises the composited RGB, so the texel
        // bank is byte-identical to orsc.quantizeRGB on the same source RGB. Slots not
        // covered by a content11 entry keep the allocateTextures default (untouched),
        // which the textured-fill path treats as a flat/transparent skip.
        loadTextures(scene, surface);
        // ---- Phase-1 entity sprite decode (RSC_MESH_NPC gated) ----
        // Decode the NPC-Rat frame-0 body block from the AUTHENTIC content1 archive into
        // the Surface sprite slots at 837 (the SAME slot the live loadEntitySprites writes
        // for animID 123), using the real multi-frame Surface.parseSprite. The projected
        // billboard rect (below) then blits this slot via the 10-arg Surface.spriteClipping.
        boolean ratLoaded = false;
        if (entityGateEngaged()) {
            ratLoaded = loadRatSprite(surface);
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
        // These MUST carry the SAME overlay/roof defs the orsc engine bakes in
        // (render/orsc/world.go tileDefs + roofs.go authenticElevationDefs) so all
        // three engines flatten/colour/peak identically — a fixture with a nonzero
        // overlay (water/bridge/lava) or roof id otherwise diverges (orsc uses its
        // table; a zeroed table renders tileType 0 / colour 0). Size 256 covers any
        // 1-byte id. (The jar DumpRenderer synthesizes the SAME values obf-side.)
        int M = 256;
        client.net.ClientStream.sharedIntArrayN = overlayTypeTable(M);   // da.N  tile type (getTileValue)
        client.ui.Panel.texK = overlayColourTable(M);                    // qa.K  deco colour (getColour)
        client.util.DecodeBuffer.landscapeFaceFlags = new int[M]; // landscape face flags
        client.net.StringCodec.DEAD_INT_ARRAY = new int[M];     // u.a   wall adjacency
        client.util.Utility.sizedPoolCounts = new int[M];       // mb.a  object type
        client.data.NameTable.sortKeys = new int[M];            // ub.g  object width
        client.data.RecordLoader.intArray = new int[M];         // f.f   object height
        client.data.CacheFile.sharedScratch = roofTextureTable(M);       // d.g   roof texture
        client.util.ErrorHandler.unusedIntTable = roofRiseTable(M);      // i.g   roof ridge height (rise)

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

        // The fixture's diagonal band carries objid 0 (value 48001). When driving by a
        // TRUE object id (RSC_MESH_OBJID != 0), remap any in-band value to 48001+objid so
        // World.addModels scans the band as that id and indexes the authentic def table at
        // it. orsc (-objid) + the JAR (RSC_MESH_OBJID) remap their bands identically.
        int meshObjIdBand = envInt("RSC_MESH_OBJID", 0);

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
                int diagVal = wallDiag[gi];
                if (meshObjIdBand != 0 && diagVal > 48000 && diagVal < 60000) {
                    diagVal = 48001 + meshObjIdBand;
                }
                putInt(wallsDiagonal, gx, gy, diagVal);
            }
        }
        System.out.println("injected grids at gridBase (" + gridBaseX + "," + gridBaseY + ")");

        // ---- 6. build the section (loadMapData no-ops: landscapePack==null) ----
        world.loadSection(HOST_GRID_LOCAL, (byte) -90, HOST_GRID_LOCAL, win.i("plane"));
        int modelCount = ((Integer) gf(scene, "modelCount")).intValue();
        System.out.println("loadSection OK; scene model count=" + modelCount);

        // ---- 6b. (optional) place a REAL scenery object from the "3d models" archive ----
        // Gated on RSC_MESH_MODEL so the terrain-only fixtures render exactly as before.
        // When set, load the authentic rev-235 "3d models" content archive (content9)
        // the SAME way the live client does (read file -> World.unpackData strips the
        // 6-byte header + bzip-inflates), register the model name -> prototype index
        // (GameModel.textureId), synthesize the object def the diagonal placer reads
        // (entityIndexTableF / sortKeys / intArray), decode the chosen model in place
        // (NameHash.getFileOffset + GameModel(byte[],offset,true)), and run the real
        // World.addModels — the DEOB leg of the 3-engine scenery-mesh parity check.
        // The fixture's wallDiag band (48001+objid) drives placement; orsc + the JAR
        // load the IDENTICAL .ob3 bytes and place by the SAME addModels math.
        // RSC_MESH_REALDEFS: when set, the object def (model name + W/H) is read from the
        // AUTHENTIC rev-235 def table in cache content0 (string.dat + integer.dat) instead
        // of the RSC_MESH_MODEL/W/H synthesized values. content0 is unpacked the same way
        // content9 is, then SocketFactory.initGameData parses it into the SAME static arrays
        // World.addModels reads (entityIndexTableF objId->modelIdx, NameTable.sortKeys W,
        // RecordLoader.intArray H, modelNames registered via GameModel.textureId). This is
        // the live client's own parse path (Mudclient.drawOptionsTab -> initGameData), so
        // the object surface is driven by TRUE map ids, not harness-synthesized defs.
        // Value "1" (or empty path) uses RSC_MESH_CACHE / the default cache dir.
        String realDefs = System.getenv("RSC_MESH_REALDEFS");
        boolean useRealDefs = realDefs != null && !realDefs.isEmpty();

        String meshModel = System.getenv("RSC_MESH_MODEL");
        if (useRealDefs || (meshModel != null && !meshModel.isEmpty())) {
            int meshObjId = envInt("RSC_MESH_OBJID", 0);
            int meshW = envInt("RSC_MESH_W", 1);
            int meshH = envInt("RSC_MESH_H", 1);
            String cacheDir = System.getenv("RSC_MESH_CACHE");
            if (cacheDir == null || cacheDir.isEmpty()) cacheDir = "/tmp/rsc-run/cache";
            // content9 = readDataFile("3d models",60,9,84); resolve by prefix since the
            // CRC suffix changes when the cache is rebuilt.
            File pack = findContentPack(cacheDir, "content9_");
            byte[] raw = readAll(pack.getPath());
            byte[] arc = World.unpackData(128, false, raw); // outer header strip + bzip inflate
            // (verbose=false: the verbose path reportProgress()s through a null headless shell)

            client.scene.SpriteScaler.modelNameCount = 0;
            int modelIdx;
            if (useRealDefs) {
                // Parse the authentic def archive (content0) ONCE, early, before any model
                // decode. initGameData fills entityIndexTableF/sortKeys/intArray/sizedPoolCounts
                // for ALL 1189 rev-235 objects; we then look the test id up directly.
                File defPack = findContentPack(cacheDir, "content0_");
                byte[] defRaw = readAll(defPack.getPath());
                byte[] defArc = World.unpackData(128, false, defRaw);
                client.net.SocketFactory.initGameData(defArc, (byte) 100, false);
                int[] tF = client.scene.SurfaceImageProducer.entityIndexTableF;
                if (meshObjId < 0 || meshObjId >= tF.length) {
                    throw new IllegalStateException("RSC_MESH_OBJID " + meshObjId
                            + " out of authentic def range [0," + tF.length + ")");
                }
                modelIdx = tF[meshObjId];
                meshModel = client.data.NameTable.modelNames[modelIdx];
                meshW = client.data.NameTable.sortKeys[meshObjId];   // ub.g object width
                meshH = client.data.RecordLoader.intArray[meshObjId]; // f.f object height
                // Match the harness "fresh scene" assumption: addModels' removeObject2 uses
                // the object TYPE (sizedPoolCounts/mb.a). The synthesized path pins it to 0;
                // pin it here too so a real non-zero type cannot take a different removal path.
                client.util.Utility.sizedPoolCounts[meshObjId] = 0;
                System.out.println("realdefs: id=" + meshObjId + " -> model='" + meshModel
                        + "' modelIdx=" + modelIdx + " W=" + meshW + " H=" + meshH);
            } else {
                modelIdx = GameModel.textureId((byte) 91, meshModel); // registers NameTable.modelNames[idx]
                client.scene.SurfaceImageProducer.entityIndexTableF = new int[256]; // objId -> model idx (fb.f)
                client.scene.SurfaceImageProducer.entityIndexTableF[meshObjId] = modelIdx;
                client.data.NameTable.sortKeys[meshObjId] = meshW;    // object width  (ub.g)
                client.data.RecordLoader.intArray[meshObjId] = meshH; // object height (f.f)
            }

            GameModel[] objectModels = new GameModel[5000];
            for (int i = 0; i < objectModels.length; i++) objectModels[i] = new GameModel(1, 1);
            int off = client.data.NameHash.getFileOffset(meshModel + ".ob3", (byte) 68, arc);
            objectModels[modelIdx] = new GameModel(arc, off, true);

            // Object heading: addModels reads getTileDirection (the mb grid) per tile and
            // rotates dir*32 about Y. The dump injects no direction grid, so fill it
            // uniformly from RSC_MESH_DIR (default 0). orsc + the JAR fill their grid the same.
            int meshDir = envInt("RSC_MESH_DIR", 0);
            byte[][] tileDirection = (byte[][]) gf(world, "tileDirection"); // obf mb
            for (int q = 0; q < tileDirection.length; q++) Arrays.fill(tileDirection[q], (byte) meshDir);

            // World.addModels: scan the diagonal band, clone prototypes[entityIndexTableF[id]],
            // translate to tile-centre at terrain height, orient by tileDirection, register,
            // and setLight(48,48,-10,magic^9,-50,-50). magic=-113 => arg4=-122 (== orsc -122).
            world.addModels(objectModels, (byte) -113);
            modelCount = ((Integer) gf(scene, "modelCount")).intValue();
            System.out.println("placed mesh object '" + meshModel + "' id=" + meshObjId
                    + " modelIdx=" + modelIdx + " off=" + off + " (nV=" + objectModels[modelIdx].numVertices
                    + " nF=" + objectModels[modelIdx].numFaces + "); scene model count=" + modelCount);
        }

        // ---- 7. camera (AUTHENTIC arg order, Mudclient.java:6637) + render ----
        // setCameraOrientation(camX, camY, eyeLen, yaw, mode, pitch, eyeY, roll)
        // Live call: setCameraOrientation(cx, cz, 2*ac, 912, -12349, ug*4, -elev, 0)
        //   => yaw arg carries the FIXED 912 downward tilt; pitch arg carries the
        //      orbit rotation ug*4 (the fixture's camera.yaw = orbit rotation = 512).
        int cx = HOST_GRID_LOCAL * 128 + 64; // 6208
        int cz = HOST_GRID_LOCAL * 128 + 64; // 6208
        int elev = world.getElevation(cx, cz);

        // ---- 6c. (optional) Phase-0 entity placement-sanity billboard ----
        // Gated on RSC_MESH_NPC so terrain/scenery fixtures render exactly as before.
        // Register the shared static-entity-spec billboard with the SAME camera as
        // terrain/scenery (Scene.addSprite -> the `view` GameModel billboard face),
        // so Scene.render PROJECTS it. addSprite's world args map (see GameModel
        // .createVertex(false,z,x,y)): vertexX <- z-arg, vertexZ <- x-arg, vertexY <-
        // y-arg. Foot at (cx,cz,-elev); cx==cz at the centre tile so x/z are equal.
        // After render we read back the PROJECTED foot vertex + the private projection
        // fields and recompute the rect (the projection-rect replicator, plan tool #5).
        int npcSpriteFace = -1;
        if (entityGateEngaged()) {
            // addSprite(id, x, tag, z, y, w, h, guard=109): x<-wz, z<-wx, y<-footY.
            npcSpriteFace = scene.addSprite(0, cz, 0, cx, -elev, RAT_BILLBOARD_W, RAT_BILLBOARD_H, (byte) 109);
            System.out.println("phase0 entity: registered billboard face=" + npcSpriteFace
                    + " at (" + cx + "," + cz + ",-elev=" + (-elev) + ") size " + RAT_BILLBOARD_W + "x" + RAT_BILLBOARD_H);
        }

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

        // ---- 7b. Phase-0 projection-rect replicator (plan tool #5) ----
        // The projected billboard rect (Scene.drawSpriteFace, Scene.java:1841) is a
        // method-local computed from PRIVATE fields (baseX/baseY/viewDistance). It is
        // impossible to read the rect itself, so we REPLICATE it: read the engine's
        // own projected foot vertex (view.vertexViewX/Y + projectVertexZ, populated by
        // Scene.render's this.view.project) and the private projection fields by
        // reflection, then recompute the canonical 5-line rect:
        //   w = (spriteWidth << viewDistance) / vz
        //   h = (spriteHeight << viewDistance) / vz
        //   x = (vx - w/2) + baseX
        //   y = baseY + vy - h
        // Then BLIT the rat sprite (Phase 1) into that rect via the 10-arg
        // Surface.spriteClipping (NOT through SurfaceSprite, whose .client is null
        // headless). For the rat the layer loop collapses to ONE blit (a single
        // non-empty Tg layer), and since the drawn layer IS the base frame the per-part
        // width i5 = (w * spriteWidthFull[frame]) / spriteWidthFull[base] == w, dx=dy=0.
        // The skew tx = vertexViewX[fv[1]] - vx (Scene.java:1841 / mudclient204
        // Scene.java:1404). Falls back to the Phase-0 solid fill if no sprite decoded.
        if (npcSpriteFace >= 0) {
            GameModel view = scene.view;
            int[] fv = view.faceVertices[npcSpriteFace];
            int v0 = fv[0];
            int vx = view.vertexViewX[v0];       // pb
            int vy = view.vertexViewY[v0];       // Ob
            int vz = view.projectVertexZ[v0];    // bb
            int tx = view.vertexViewX[fv[1]] - vx;             // skew (per-row shear)
            int sBaseX = ((Integer) gf(scene, "baseX")).intValue();        // Zb
            int sBaseY = ((Integer) gf(scene, "baseY")).intValue();        // Nb
            int viewDistance = ((Integer) gf(scene, "viewDistance")).intValue(); // R
            int w = (RAT_BILLBOARD_W << viewDistance) / vz;
            int h = (RAT_BILLBOARD_H << viewDistance) / vz;
            int scale = (256 << viewDistance) / vz;
            int rx = (vx - w / 2) + sBaseX;
            int ry = sBaseY + vy - h;
            System.out.println("phase1 rect-replicator: projVtx vx=" + vx + " vy=" + vy + " vz=" + vz
                    + " skew=" + tx + " | baseX=" + sBaseX + " baseY=" + sBaseY + " viewDistance=" + viewDistance
                    + " => rect x=" + rx + " y=" + ry + " w=" + w + " h=" + h + " scale=" + scale);
            if (ratLoaded) {
                // spriteClipping(x, y, w, flip, h, sprite, colour1, colour2, skew, dummy)
                // — the obf ua.a(int,int,int,boolean,int,int,int,int,int,int) order. The
                // rat's charColour is RAW (not a 1/2/3 dye marker), so colour1 is used
                // verbatim; skin colour 0 => white single-tint fast path.
                surface.spriteClipping(rx, ry, w, false, h, RAT_SPRITE_OFFSET + RAT_FRAME,
                        RAT_CHAR_COLOUR, RAT_SKIN_COLOUR, tx, 1);
                System.out.println("phase1 blit: rat slot " + (RAT_SPRITE_OFFSET + RAT_FRAME)
                        + " colour1=" + RAT_CHAR_COLOUR + " colour2=" + RAT_SKIN_COLOUR + " skew=" + tx);
            } else {
                // Phase-0 fallback: solid cyan rect (no sprite decoded).
                surface.drawBox(rx, (byte) 0, DEBUG_BILLBOARD_COLOUR, ry, h, w);
            }
        }

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

    // ---- (optional) real-scenery-object placement helpers (RSC_MESH_* gated) ----
    static int envInt(String name, int def) {
        String s = System.getenv(name);
        if (s == null || s.isEmpty()) return def;
        try { return Integer.parseInt(s.trim()); } catch (NumberFormatException e) { return def; }
    }

    // Resolve a content pack by filename prefix (e.g. "content9_"); the CRC suffix
    // changes whenever the cache is rebuilt, so match on prefix, not exact name.
    static File findContentPack(String dir, String prefix) throws IOException {
        File d = new File(dir);
        File[] fs = d.listFiles();
        if (fs != null) {
            for (File f : fs) {
                if (f.isFile() && f.getName().startsWith(prefix)) return f;
            }
        }
        throw new IOException("no " + prefix + "* content pack in " + dir);
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
