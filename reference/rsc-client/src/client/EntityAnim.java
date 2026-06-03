package client;

/**
 * EntityAnim — the AUTHENTIC entity-animation tables + frame/flip/offset math the
 * 3-engine NPC/player render-parity rig shares (Task A2). It is the DEOB-leg
 * counterpart to orsc's render/animdefs.go + render/entitysprite.go (the
 * authenticAnimDefs table, animationNumbers(), facingPose, layerFrame) and to the
 * inlined copy in the JAR oracle (rscplus/dumprender/DumpRenderer.java).
 *
 * <p>All values are transcribed VERBATIM from Mudclient.java:
 * <ul>
 *   <li>{@link #SF} = the scroll-frame step table {0,1,2,1} (Mudclient.java:1929).</li>
 *   <li>{@link #TG} = the 8x12 equipment-slot layer-order table (Mudclient.java:1865).</li>
 *   <li>{@link #ANIM_NAME}/{@link #ANIM_COLOUR}/{@link #ANIM_HASF} = the 229 authentic
 *       animation defs (name / charColour dye-marker / hasF F-frame flag), in
 *       EntityHandler.loadAnimationDefinitions order (== render/animdefs.go
 *       authenticAnimDefs). NpcDef.Sprites[] and player appearance ids index into
 *       these.</li>
 *   <li>{@link #spriteOffset(int)} = animationNumbers()[animID]: the 27-slot stride
 *       base (15 body + 3 'a' + 9 'f'), deduped by name, with the 1998->3300 jump.
 *       Computed once, lazily.</li>
 * </ul>
 *
 * <p>The frame within an animation's 27-slot block (Mudclient.java:5859-5866 npc /
 * 5663-5670 player):
 * <pre>
 *   walkAnim = dir &amp; 7
 *   flip = false; col = walkAnim
 *   if col==5 {flip; col=3} elif 6 {flip; col=2} elif 7 {flip; col=1}
 *   frame = SF[step &amp; 3] + 3*col       // step = the RAW walk-cycle index 0..3
 *   if (flip &amp;&amp; col in 1..3 &amp;&amp; hasF) frame += 15
 * </pre>
 * — the per-layer body part is {@code TG[walkAnim][layer]}, and the COMPUTED flip is
 * passed to Surface.spriteClipping (the mirror blit).
 */
public final class EntityAnim {
    private EntityAnim() {}

    // --- Scroll-frame step table (Mudclient.java:1929) ---
    public static final int[] SF = {0, 1, 2, 1};

    // --- Equipment-slot layer ordering (8 configs x 12 slots, Mudclient.java:1865) ---
    public static final int[][] TG = {
        {11, 2, 9, 7, 1, 6, 10, 0, 5, 8, 3, 4},
        {11, 2, 9, 7, 1, 6, 10, 0, 5, 8, 3, 4},
        {11, 3, 2, 9, 7, 1, 6, 10, 0, 5, 8, 4},
        {3, 4, 2, 9, 7, 1, 6, 10, 8, 11, 0, 5},
        {3, 4, 2, 9, 7, 1, 6, 10, 8, 11, 0, 5},
        {4, 3, 2, 9, 7, 1, 6, 10, 8, 11, 0, 5},
        {11, 4, 2, 9, 7, 1, 6, 10, 0, 5, 8, 3},
        {11, 2, 9, 7, 1, 6, 10, 0, 5, 8, 4, 3},
    };

    // --- 229 authentic animation defs (== render/animdefs.go authenticAnimDefs) ---
    public static final String[] ANIM_NAME = {
        "head1", "body1", "legs1", "fhead1", "fbody1", "head2", "head3", "head4", "chefshat", "apron", "apron", "boots",
        "fullhelm", "fullhelm", "fullhelm", "fullhelm", "fullhelm", "fullhelm", "fullhelm", "fullhelm", "chainmail", "chainmail", "chainmail", "chainmail",
        "chainmail", "chainmail", "chainmail", "platemailtop", "platemailtop", "platemailtop", "platemailtop", "platemailtop", "platemailtop", "platemailtop", "platemailtop", "platemailtop",
        "platemaillegs", "platemaillegs", "platemaillegs", "platemaillegs", "platemaillegs", "platemaillegs", "platemaillegs", "platemaillegs", "platemaillegs", "leatherarmour", "leathergloves", "sword",
        "sword", "sword", "sword", "sword", "sword", "sword", "fplatemailtop", "fplatemailtop", "fplatemailtop", "fplatemailtop", "fplatemailtop", "fplatemailtop",
        "fplatemailtop", "apron", "cape", "cape", "cape", "cape", "cape", "cape", "cape", "mediumhelm", "mediumhelm", "mediumhelm",
        "mediumhelm", "mediumhelm", "mediumhelm", "mediumhelm", "wizardsrobe", "wizardshat", "wizardshat", "necklace", "necklace", "skirt", "wizardsrobe", "wizardsrobe",
        "wizardsrobe", "wizardsrobe", "skirt", "skirt", "skirt", "skirt", "skirt", "skirt", "skirt", "skirt", "skirt", "skirt",
        "skirt", "squareshield", "squareshield", "squareshield", "squareshield", "squareshield", "squareshield", "squareshield", "squareshield", "squareshield", "crossbow", "longbow",
        "battleaxe", "battleaxe", "battleaxe", "battleaxe", "battleaxe", "battleaxe", "battleaxe", "mace", "mace", "mace", "mace", "mace",
        "mace", "mace", "staff", "rat", "demon", "spider", "spider", "camel", "cow", "sheep", "unicorn", "bear",
        "chicken", "skeleton", "skelweap", "zombie", "zombweap", "ghost", "bat", "goblin", "goblin", "goblin", "gobweap", "scorpion",
        "dragon", "dragon", "dragon", "wolf", "wolf", "partyhat", "partyhat", "partyhat", "partyhat", "partyhat", "partyhat", "leathergloves",
        "chicken", "fplatemailtop", "skirt", "wolf", "spider", "battleaxe", "sword", "eyepatch", "demon", "dragon", "spider", "wolf",
        "unicorn", "demon", "spider", "necklace", "rat", "mediumhelm", "chainmail", "wizardshat", "legs1", "gasmask", "mediumhelm", "spider",
        "spear", "halloweenmask", "wizardsrobe", "skirt", "halloweenmask", "halloweenmask", "skirt", "skirt", "skirt", "skirt", "skirt", "wizardshat",
        "wizardshat", "wizardshat", "wizardshat", "wizardshat", "wizardsrobe", "wizardsrobe", "wizardsrobe", "wizardsrobe", "wizardsrobe", "wizardsrobe", "skirt", "boots",
        "boots", "boots", "boots", "boots", "santahat", "ibanstaff", "souless", "boots", "legs1", "wizardsrobe", "skirt", "cape",
        "wolf", "bunnyears", "saradominstaff", "spear", "skirt", "wizardsrobe", "wolf", "chicken", "squareshield", "cape", "boots", "wizardsrobe",
        "scythe",
    };

    public static final int[] ANIM_COLOUR = {
        1, 2, 3, 1, 2, 1, 1, 1, 16777215, 16777215, 9789488, 5592405, 16737817, 15654365, 15658734, 10072780,
        11717785, 65535, 3158064, 16777215, 16737817, 15654365, 15658734, 10072780, 11717785, 65535, 3158064, 16737817, 15654365, 15658734, 10072780, 11717785,
        3158064, 65535, 16777215, 10083839, 16737817, 15654365, 15658734, 10072780, 11717785, 65535, 4210752, 16777215, 10083839, 0, 0, 16737817,
        15654365, 15658734, 10072780, 11717785, 65535, 3158064, 16737817, 15654365, 15658734, 10072780, 11717785, 65535, 3158064, 16777215, 16711680, 2434341,
        4210926, 4246592, 15658560, 15636736, 11141341, 16737817, 15654365, 15658734, 10072780, 11717785, 65535, 3158064, 255, 255, 4210752, 15658734,
        16763980, 255, 4210752, 10510400, 15609904, 16777215, 16777215, 10510400, 4210752, 16036851, 15609904, 8400921, 7824998, 7829367, 2245205, 4347170,
        26214, 16737817, 15654365, 15658734, 10072780, 11717785, 56797, 3158064, 16750896, 11363121, 0, 0, 16737817, 15654365, 15658734, 10072780,
        11717785, 65535, 3158064, 16737817, 15654365, 15658734, 10072780, 11717785, 65535, 3158064, 0, 4805259, 16384000, 13408576, 16728144, 0,
        0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 8969727, 16711680, 47872, 65535, 0,
        65280, 16711680, 21981, 0, 10066329, 16711680, 16776960, 255, 65280, 16711935, 16777215, 11202303, 16711680, 10083839, 1118481, 9789488,
        65535, 16711748, 16711748, 0, 3158064, 3158064, 14535680, 2236962, 2236962, 6291456, 2236962, 3158064, 11184810, 11250603, 11250603, 16711680,
        9785408, 0, 16711748, 3852326, 0, 52224, 1052688, 1052688, 16711680, 255, 16755370, 11206570, 11184895, 16777164, 13434879, 16755370,
        11206570, 11184895, 16777164, 13434879, 16755370, 11206570, 11184895, 16777164, 13434879, 3978097, 3978097, 16755370, 11206570, 11184895, 16777164, 13434879,
        0, 0, 0, 16777215, 16777215, 8421376, 8421376, 16777215, 13420580, 0, 0, 56797, 1392384, 1392384, 5585408, 6893315,
        13500416, 16777215, 1118481, 1118481, 0,
    };

    public static final int[] ANIM_HASF = {
        0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
        0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
        0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
        0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
        0, 0, 0, 0, 0, 0, 1, 0, 1, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
        0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
        0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
        0, 0, 0, 0, 0,
    };

    // hasF(animID): true when the animation has the 9-frame 'f' (front/flip) sub-block
    // (entityFlags[animID] == 1). Out-of-range -> false.
    public static boolean hasF(int animID) {
        return animID >= 0 && animID < ANIM_HASF.length && ANIM_HASF[animID] == 1;
    }

    // charColour(animID): the dye marker (1=hair, 2=top, 3=bottom) or a literal 24-bit
    // colour (animationCharacterColour[animID]). Out-of-range -> 0.
    public static int charColour(int animID) {
        return animID >= 0 && animID < ANIM_COLOUR.length ? ANIM_COLOUR[animID] : 0;
    }

    // --- spriteOffsets, computed once (animationNumbers / loadEntitiesAuthentic) ---
    private static int[] OFFSETS;

    public static synchronized int spriteOffset(int animID) {
        if (OFFSETS == null) {
            OFFSETS = new int[ANIM_NAME.length];
            java.util.HashMap<String, Integer> seen = new java.util.HashMap<>();
            int n = 0;
            for (int i = 0; i < ANIM_NAME.length; i++) {
                Integer prev = seen.get(ANIM_NAME[i]);
                if (prev != null) {
                    OFFSETS[i] = prev;
                    continue;
                }
                OFFSETS[i] = n;
                seen.put(ANIM_NAME[i], n);
                n += 27;
                if (n == 1998) {
                    n = 3300;
                }
            }
        }
        return (animID >= 0 && animID < OFFSETS.length) ? OFFSETS[animID] : -1;
    }

    // --- walkAnim/flip/col, the dir resolution (Mudclient.java:5859-5864) ---
    public static int walkAnim(int dir) { return dir & 7; }

    public static boolean flipForDir(int dir) {
        int d = dir & 7;
        return d == 5 || d == 6 || d == 7;
    }

    // dirColumn(dir): the flip-resolved column (0..4) — dirs 5/6/7 collapse onto 3/2/1.
    public static int dirColumn(int dir) {
        int d = dir & 7;
        if (d == 5) return 3;
        if (d == 6) return 2;
        if (d == 7) return 1;
        return d;
    }

    // layerFrame(animID, col, step, flip): the per-layer frame within the 27-slot block.
    // step = the RAW walk-cycle index 0..3 (resolved here via SF). +15 F-frame when the
    // pose is flipped (col in 1..3) and the animation has an F-frame.
    public static int layerFrame(int animID, int col, int step, boolean flip) {
        int frame = SF[step & 3] + 3 * col;
        if (flip && col >= 1 && col <= 3 && hasF(animID)) {
            frame += 15;
        }
        return frame;
    }
}
