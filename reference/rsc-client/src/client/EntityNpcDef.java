package client;

/**
 * EntityNpcDef — the gated test NPCs' authentic defs (sprites / billboard camera /
 * colours), hardcoded from OpenRSC NpcDefs.json (the SAME table orsc's
 * facts.NpcDefs resolves). Task A2 def-drives the billboard W/H + sprite layers +
 * colours from the gated npcId so id45 resolves 216x234 with skelweap/body, not the
 * old hardcoded 346x136 rat billboard.
 *
 * <p>Per NpcDefs.json (proven: camera1/camera2 == billboard W x H; colours 0 = raw,
 * no dye markers):
 * <ul>
 *   <li>id 19 Rat:      sprites[123],        cam 346x136, colours 0 (hasF=false)</li>
 *   <li>id 45 skeleton: sprites[134,133],    cam 216x234, colours 0 (skelweap@slot0
 *       hasF, body@slot1) — the CANONICAL +15 F-frame entity</li>
 *   <li>id 4  Goblin:   sprites[142,139],    cam 219x206, colours 0 (gobweap hasF)</li>
 * </ul>
 *
 * <p>NPC colour fields (hairColour/topColour/bottomColour/skinColour) are RAW 24-bit
 * dye values (NOT palette indices like a player) — but for every test NPC they are 0
 * and no layer's charColour is a 1/2/3 marker, so the per-layer dye collapses to the
 * default branch (dye = animationCharacterColour, skin = 0), matching orsc's
 * compositeNPC(rawColours=true).
 */
public final class EntityNpcDef {

    public final int id;
    public final int[] sprites; // bodyPart -> animID, -1 = empty (12 slots)
    public final int camera1, camera2; // billboard W x H
    public final int hairColour, topColour, bottomColour, skinColour;

    private EntityNpcDef(int id, int[] sprites, int camera1, int camera2,
                         int hair, int top, int bottom, int skin) {
        this.id = id;
        this.sprites = sprites;
        this.camera1 = camera1;
        this.camera2 = camera2;
        this.hairColour = hair;
        this.topColour = top;
        this.bottomColour = bottom;
        this.skinColour = skin;
    }

    private static int[] s12(int... vals) {
        int[] r = new int[12];
        java.util.Arrays.fill(r, -1);
        for (int i = 0; i < vals.length && i < 12; i++) r[i] = vals[i];
        return r;
    }

    // Resolve the gated NPC def. Returns null for an id with no hardcoded def.
    public static EntityNpcDef def(int id) {
        switch (id) {
            case 19: return new EntityNpcDef(19, s12(123),      346, 136, 0, 0, 0, 0);
            case 45: return new EntityNpcDef(45, s12(134, 133), 216, 234, 0, 0, 0, 0);
            case 4:  return new EntityNpcDef(4,  s12(142, 139), 219, 206, 0, 0, 0, 0);
            default: return null;
        }
    }
}
