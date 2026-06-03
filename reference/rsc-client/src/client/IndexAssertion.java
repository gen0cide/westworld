package client;

import client.data.BZip;
import client.data.CacheUpdater;
import client.data.DataStore;
import client.net.Packet;
import client.net.SocketFactory;
import client.scene.SurfaceImageProducer;
import client.shell.GameFrame;
import client.util.LinkedQueue;
import client.util.StreamFactory;
import client.world.World;
import client.world.WorldEntity;

import java.io.File;
import java.io.FileInputStream;
import java.io.IOException;
import java.lang.reflect.Field;
import java.util.HashMap;
import java.util.Map;

/**
 * Phase-0.5 sprite-index regression guard for the NPC/player 2D-sprite parity rig
 * (docs/build/NPC_SPRITE_PARITY_PLAN.md, Phase 0.5).
 *
 * <p>This LOCKS the index-alignment chain that the whole entity pipeline rides on, so a
 * cache rebuild or a {@code SocketFactory.initGameData} edit cannot silently shift the
 * sprite-block bases, recolour, or the NPC sprite grid out from under
 * {@code drawNpc}/{@code drawPlayer} (Mudclient.java:5845 / :5655) without this test going
 * red. It drives off the AUTHENTIC rev-235 cache content0 GameData — the same archive
 * the live client and the realdefs scenery path parse — so it is a real data assertion,
 * not a hand-mirror.
 *
 * <p>Pipeline (mirrors what Phase 1 will need before any pixel decode):
 * <ol>
 *   <li>Unpack cache content0 ("config"/GameData) and run {@code SocketFactory.initGameData}
 *       (param 100 = full init) — fills {@code CacheUpdater.contentNames} (animation names),
 *       {@code LinkedQueue.sharedIntArray2} (charColour / animationCharacterColour),
 *       {@code BZip.entityFlags} (the +9 f.dat "hasF" flag), {@code FontWidths.prayerM},
 *       {@code DataStore.tamperScratch}, the NPC equipped-sprite grid
 *       {@code GameFrame.unusedIntBuffer2d[serverId][bodyPart]}, and the NPC billboard size
 *       tables {@code SurfaceImageProducer.entityIndexTableC} / {@code Packet.legacyMaskTable},
 *       and allocates (but does NOT fill) {@code WorldEntity.spriteOffsets}.</li>
 *   <li>Reflection-set {@code StreamFactory.engineArraySize = CacheUpdater.contentNames.length}.
 *       The clean source (SocketFactory.java:596) lost the write of {@code na.e}; without this,
 *       {@code Mudclient.loadEntitySprites} loops 0 times and every sprite offset stays 0.
 *       (Plan tool #4; do NOT edit SocketFactory — this is harness-side.)</li>
 *   <li>Replicate the {@code Mudclient.loadEntitySprites} INDEX stride (Mudclient.java:2554-2618):
 *       walk {@code contentNames}, alias a repeated name's slot (no advance), else assign the
 *       running cursor {@code uc} and {@code uc += 27}. This is pure index math (the +27 / dedup
 *       loop) — it needs NO content1 pixel decode — and is exactly the chain
 *       {@code animationNumbers()} mirrors on the orsc side (render/animdefs.go).</li>
 *   <li>Assert the rat (content0 serverId 19, animID 123) alignment invariants.</li>
 * </ol>
 *
 * <p>Run headless on JDK 17:
 * <pre>
 *   java -Djava.awt.headless=true -cp /tmp/deob-run client.IndexAssertion [CACHE_DIR]
 * </pre>
 * Exits 0 on success, 1 on any failed assertion (so the build/diff script can gate on it).
 */
public final class IndexAssertion {

    // The first test entity: NPC Rat. content0 serverId 19 -> sprite grid slot 0 holds the
    // RAW animID 123; animID 123 -> sprite-block base 837 (the 124th unique-named 27-stride
    // block); charColour 4805259 (a raw 24-bit value, NOT a 1/2/3 dye marker, so the dye path
    // is identity); hasF false (no +9 f.dat block). The NPC draw path reads
    // equippedItem[bodyPart] - 1, so the harness synthesizes equippedItem[0] = animID + 1 = 124
    // and 124 - 1 recovers 123. The PLAYER path seeds the grid = animID directly (no -1).
    static final int RAT_SERVER_ID = 19;
    static final int RAT_ANIM_ID = 123;
    static final int RAT_SPRITE_OFFSET = 837;
    static final int RAT_CHAR_COLOUR = 4805259;
    static final boolean RAT_HAS_F = false;
    static final int RAT_EQUIPPED_ITEM = RAT_ANIM_ID + 1; // 124 (NPC path; -1 recovers 123)

    // The expected NPC billboard size for the rat (entityIndexTableC[19] x legacyMaskTable[19]),
    // = OpenRSC NpcDefs id 19 camera1/camera2. Locked here as a bonus alignment guard (Phase 0
    // already proved all three legs project this 346x136 rect identically).
    static final int RAT_BILLBOARD_W = 346;
    static final int RAT_BILLBOARD_H = 136;

    static int failures = 0;

    public static void main(String[] args) throws Exception {
        String cacheDir = args.length > 0 ? args[0] : System.getenv("RSC_MESH_CACHE");
        if (cacheDir == null || cacheDir.isEmpty()) cacheDir = "/tmp/rsc-run/cache";

        // ---- 1. unpack content0 + initGameData (full init) ----
        File defPack = findContentPack(cacheDir, "content0_");
        byte[] defRaw = readAll(defPack.getPath());
        byte[] defArc = World.unpackData(128, false, defRaw); // strip 6-byte header + bzip inflate
        SocketFactory.initGameData(defArc, (byte) 100, false);

        int prayerCount = CacheUpdater.contentNames.length;
        System.out.println("content0 GameData parsed: contentNames.length=" + prayerCount);

        // ---- 2. reflection-set engineArraySize (the lost na.e write; plan tool #4) ----
        int engineBefore = StreamFactory.engineArraySize;
        Field f = StreamFactory.class.getDeclaredField("engineArraySize");
        f.setAccessible(true);
        f.setInt(null, prayerCount);
        System.out.println("engineArraySize: before=" + engineBefore + " (lost na.e write) -> after="
                + StreamFactory.engineArraySize);
        check("engineArraySize was unassigned before the harness fix", engineBefore == 0);
        check("engineArraySize == contentNames.length after the fix",
                StreamFactory.engineArraySize == prayerCount);

        // ---- 3. replicate the loadEntitySprites +27 dedup-by-name INDEX stride ----
        // (Mudclient.loadEntitySprites, Mudclient.java:2554-2618 — index math only, no pixels.)
        int uc = 0;
        label:
        for (int idx = 0; idx < prayerCount; idx++) {
            String name = CacheUpdater.contentNames[idx];
            for (int prev = 0; prev < idx; prev++) {
                if (CacheUpdater.contentNames[prev].equalsIgnoreCase(name)) {
                    WorldEntity.spriteOffsets[idx] = WorldEntity.spriteOffsets[prev];
                    continue label;
                }
            }
            WorldEntity.spriteOffsets[idx] = uc;
            uc += 27;
        }

        // ---- 4. the rat alignment assertions ----
        int gridAnimId = GameFrame.unusedIntBuffer2d[RAT_SERVER_ID][0];
        int offsetFromCache = WorldEntity.spriteOffsets[RAT_ANIM_ID];
        int offsetFromAnimNumbers = animationNumbers(prayerCount)[RAT_ANIM_ID];
        int charColour = LinkedQueue.sharedIntArray2[RAT_ANIM_ID];
        boolean hasF = BZip.entityFlags[RAT_ANIM_ID] == 1;
        int billboardW = SurfaceImageProducer.entityIndexTableC[RAT_SERVER_ID];
        int billboardH = Packet.legacyMaskTable[RAT_SERVER_ID];
        int equippedRoundTrip = RAT_EQUIPPED_ITEM - 1; // NPC path: equippedItem[bodyPart] - 1

        System.out.println("rat: contentNames[" + RAT_ANIM_ID + "]='" + CacheUpdater.contentNames[RAT_ANIM_ID]
                + "' grid[" + RAT_SERVER_ID + "][0]=" + gridAnimId
                + " spriteOffsets[" + RAT_ANIM_ID + "]=" + offsetFromCache
                + " animationNumbers()[" + RAT_ANIM_ID + "]=" + offsetFromAnimNumbers
                + " charColour=" + charColour + " hasF=" + hasF
                + " billboard=" + billboardW + "x" + billboardH
                + " equipped(124-1)=" + equippedRoundTrip);

        check("content0 grid[serverId=19][0] == animID 123", gridAnimId == RAT_ANIM_ID);
        check("spriteOffsets[123] == 837", offsetFromCache == RAT_SPRITE_OFFSET);
        check("spriteOffsets[123] == animationNumbers()[123]", offsetFromCache == offsetFromAnimNumbers);
        check("animationNumbers()[123] == 837", offsetFromAnimNumbers == RAT_SPRITE_OFFSET);
        check("charColour[123] == 4805259 (raw, no dye)", charColour == RAT_CHAR_COLOUR);
        check("hasF[123] == false", hasF == RAT_HAS_F);
        check("equippedItem round-trip 124-1 == 123", equippedRoundTrip == RAT_ANIM_ID);
        check("rat billboard width (entityIndexTableC[19]) == 346", billboardW == RAT_BILLBOARD_W);
        check("rat billboard height (legacyMaskTable[19]) == 136", billboardH == RAT_BILLBOARD_H);

        if (failures == 0) {
            System.out.println("INDEX-ASSERTION PASS: all " + 11 + " sprite-index invariants hold.");
            System.exit(0);
        } else {
            System.out.println("INDEX-ASSERTION FAIL: " + failures + " invariant(s) broken — index alignment drifted.");
            System.exit(1);
        }
    }

    static void check(String label, boolean ok) {
        System.out.println((ok ? "  [PASS] " : "  [FAIL] ") + label);
        if (!ok) failures++;
    }

    /**
     * Self-contained replica of render/animdefs.go animationNumbers() — the orsc-side
     * sprite-block-base assignment (mudclient.loadEntitiesAuthentic): walk the 229 authentic
     * animation NAMES, dedup by name (a repeated name shares the earlier base), assigning +27
     * per unique name with the 1998 -> 3300 jump. Returns the same table the cache-driven
     * loadEntitySprites stride builds for indices below the jump; we cross-check
     * spriteOffsets[123] against animationNumbers()[123] so the two derivations of the rat
     * base (cache stride vs the authentic name list) must agree. Trimmed to the rat prefix
     * (the first 124 names) — that is all this guard needs, and it pins the rat to 837.
     */
    static int[] animationNumbers(int countCap) {
        String[] names = AUTHENTIC_ANIM_NAMES;
        int[] nums = new int[names.length];
        Map<String, Integer> seen = new HashMap<>();
        int n = 0;
        for (int i = 0; i < names.length; i++) {
            Integer prev = seen.get(names[i]);
            if (prev != null) {
                nums[i] = prev;
                continue;
            }
            nums[i] = n;
            seen.put(names[i], n);
            n += 27;
            if (n == 1998) n = 3300;
        }
        return nums;
    }

    // The authentic animation NAMES in EntityHandler order, indices 0..123 (head1 .. rat).
    // VERBATIM from render/animdefs.go authenticAnimDefs[].name (the rat prefix). The full
    // set is 229 long; only the first 124 are needed to pin the rat base (837), and the
    // 1998->3300 jump occurs far past index 123 so the prefix suffices for this guard.
    static final String[] AUTHENTIC_ANIM_NAMES = {
        "head1", "body1", "legs1", "fhead1", "fbody1", "head2", "head3", "head4", "chefshat",
        "apron", "apron", "boots", "fullhelm", "fullhelm", "fullhelm", "fullhelm", "fullhelm",
        "fullhelm", "fullhelm", "fullhelm", "chainmail", "chainmail", "chainmail", "chainmail",
        "chainmail", "chainmail", "chainmail", "platemailtop", "platemailtop", "platemailtop",
        "platemailtop", "platemailtop", "platemailtop", "platemailtop", "platemailtop",
        "platemailtop", "platemaillegs", "platemaillegs", "platemaillegs", "platemaillegs",
        "platemaillegs", "platemaillegs", "platemaillegs", "platemaillegs", "platemaillegs",
        "leatherarmour", "leathergloves", "sword", "sword", "sword", "sword", "sword", "sword",
        "sword", "fplatemailtop", "fplatemailtop", "fplatemailtop", "fplatemailtop",
        "fplatemailtop", "fplatemailtop", "fplatemailtop", "apron", "cape", "cape", "cape",
        "cape", "cape", "cape", "cape", "mediumhelm", "mediumhelm", "mediumhelm", "mediumhelm",
        "mediumhelm", "mediumhelm", "mediumhelm", "wizardsrobe", "wizardshat", "wizardshat",
        "necklace", "necklace", "skirt", "wizardsrobe", "wizardsrobe", "wizardsrobe",
        "wizardsrobe", "skirt", "skirt", "skirt", "skirt", "skirt", "skirt", "skirt", "skirt",
        "skirt", "skirt", "skirt", "squareshield", "squareshield", "squareshield",
        "squareshield", "squareshield", "squareshield", "squareshield", "squareshield",
        "squareshield", "crossbow", "longbow", "battleaxe", "battleaxe", "battleaxe",
        "battleaxe", "battleaxe", "battleaxe", "battleaxe", "mace", "mace", "mace", "mace",
        "mace", "mace", "mace", "staff", "rat",
    };

    static byte[] readAll(String path) throws IOException {
        try (FileInputStream in = new FileInputStream(path)) {
            return in.readAllBytes();
        }
    }

    static File findContentPack(String dir, String prefix) throws IOException {
        File d = new File(dir);
        File[] fs = d.listFiles();
        if (fs != null) {
            for (File ff : fs) {
                if (ff.isFile() && ff.getName().startsWith(prefix)) return ff;
            }
        }
        throw new IOException("no " + prefix + "* content pack in " + dir);
    }
}
