package client.world;

import java.io.DataInputStream;
import java.io.EOFException;
import java.io.IOException;
import java.io.InputStream;

/**
 * GameCharacter — a player or NPC entity in the RSC world (rev ~233-235, Microsoft J++ build).
 *
 * <p>One instance covers both players and NPCs: the same class is used for {@code client.wi}
 * (local player singleton), {@code client.We[]} / {@code client.Zg[]} (remote players),
 * and {@code client.Tb[]} / {@code client.rg[]} (NPCs / rendered entities).
 *
 * <p><b>Movement:</b> driven by a circular 10-slot waypoint queue
 * ({@link #waypointsX} / {@link #waypointsY}).  The head index is {@link #waypointCurrent};
 * each tick the entity interpolates toward {@code waypointsX[waypointCurrent]}, incrementing
 * {@link #movingStep} to select the correct walk-cycle animation frame ({@code movingStep/6 % 4}).
 * When the entity arrives, {@link #stepCount} advances the queue head.
 *
 * <p><b>Rendering overlay timers:</b>
 * <ul>
 *   <li>{@link #combatTimer} (200 ticks) — enables health-bar and skull overlay.
 *   <li>{@link #messageTimeout} (150 ticks) — shows overhead chat {@link #message}.
 *   <li>{@link #bubbleTimeout} (150 ticks) — shows overhead item-bubble {@link #bubbleItem}.
 *   <li>{@link #projectileRange} — projectile animation progress (counts down from
 *       {@code client.nc} to 0 while the spell/arrow travels).
 * </ul>
 *
 * <p><b>Appearance / discrimination:</b>
 * For NPC entities {@link #npcIdOrColourBottom} holds the NPC definition ID
 * (255 = sentinel "no NPC" / slot not yet initialised).
 * For player entities the same field receives the {@code colourBottom} appearance byte
 * from the player-update packet.  All other colour fields ({@link #colourHair},
 * {@link #colourTop}, {@link #colourSkin}) are player-only.
 *
 * <p><b>Oracle cross-reference:</b>
 * {@code mudclient204/src/GameCharacter.java} (rev 204, structurally equivalent).
 * LeadingBot {@code mudclient.java} field prefixes: {@code gm*} / {@code gn*} on its {@code f} class.
 *
 * <p><b>Obfuscation artefacts stripped:</b>
 * <ul>
 *   <li>Opaque predicate {@code boolean bl = client.vh;} / dead {@code if(bl)} branches removed.
 *   <li>Per-method profiling counters
 *       ({@link #PROFILING_COUNTER_readFromStore}, {@link #PROFILING_COUNTER_popcount}) noted.
 *   <li>Try/catch re-throw wrappers via {@code i.a(e, "sig...")} unwrapped.
 *   <li>Anti-tamper guard {@code if (param != -19675) f = null;} in
 *       {@link #readFromStore} removed.
 *   <li>Anti-tamper junk division {@code 110 / ((65 - dummyByte) / 44)} in
 *       {@link #popcount} removed.
 *   <li>XOR string pool decoded (key table mod-5: {0→75, 1→95, 2→65, 3→90, 4→61}).
 * </ul>
 *
 * <p><b>Decoded XOR strings</b> (used only in error-reporting / catch blocks):
 * <pre>
 *   L[0] = "ta.B("   // obf: z(z("?>o\x18\x15"))  — catch prefix for popcount()
 *   L[1] = "ta.A("   // obf: z(z("?>o\x1b\x15"))  — catch prefix for readFromStore()
 *   L[2] = "null"    // obf: z(z("%*-6"))
 *   L[3] = "{...}"   // obf: z(z("0qot@"))
 * </pre>
 *
 * <p>The static {@link #CHAT_CIPHER_METADATA} field ({@code v} = {@code client.net.ChatCipher})
 * is initialised with decoded strings {@code "WTWIP"}, {@code "office"}, {@code "_wip"} —
 * pure obfuscation / profiling metadata tags, never read at runtime for game logic.
 */
public final class GameCharacter {

    // =========================================================================
    // Static / class-level fields
    // =========================================================================

    /**
     * Obfuscation-injected ChatCipher instance holding profiling metadata tags
     * (decoded: "WTWIP", "office", "_wip", 3).  Never used for real game logic.
     * obf: f  (type: v = client.net.ChatCipher)
     */
    static client.net.ChatCipher CHAT_CIPHER_METADATA =
            new client.net.ChatCipher("WTWIP", "office", "_wip", 3);

    /**
     * Dead profiling counter incremented once per call to {@link #popcount}.
     * obf: l
     */
    static int PROFILING_COUNTER_popcount;

    /**
     * Dead profiling counter incremented once per call to {@link #readFromStore}.
     * obf: v
     */
    static int PROFILING_COUNTER_readFromStore;

    /**
     * Shared static name-string table — populated elsewhere in the engine;
     * exact role within this class is uncertain (never written here).
     * obf: r
     */
    static String[] sharedNameTable;

    /**
     * Constant 176 — likely a colour-palette offset or a sprite-table base used by
     * the rendering subsystem when indexing player appearance sprites.
     * obf: g
     */
    static int COLOUR_SPRITE_BASE = 176;

    // =========================================================================
    // XOR-decoded error-message string pool (private, class-local)
    // =========================================================================

    /**
     * Error-message string literals decoded from the per-class XOR pool.
     * <pre>
     *   [0] = "ta.B("   — catch prefix for popcount()
     *   [1] = "ta.A("   — catch prefix for readFromStore()
     *   [2] = "null"
     *   [3] = "{...}"
     * </pre>
     * obf: L
     */
    private static final String[] ERROR_STRINGS = new String[]{
            "ta.B(",   // obf: z(z("?>o\x18\x15"))
            "ta.A(",   // obf: z(z("?>o\x1b\x15"))
            "null",    // obf: z(z("%*-6"))
            "{...}"    // obf: z(z("0qot@"))
    };

    // =========================================================================
    // Instance identity fields
    // =========================================================================

    /**
     * Server-assigned entity index — the slot number in the server's player or
     * NPC table.  Set to the array index when the entity is first created/found.
     * Rendering uses this as the primary identity key for deduplication.
     * Corresponds to oracle {@code GameCharacter.serverIndex}.
     * LeadingBot: {@code .gmf}
     * obf: b
     */
    int serverIndex;

    /**
     * Per-frame server-assigned ID / hash used to track the same entity between
     * server ticks (e.g. for player-entity tracking across movement updates).
     * Also used as an index when looking up NPC sprite tables (see {@link #npcIdOrColourBottom}).
     * Corresponds to oracle {@code GameCharacter.serverId}.
     * LeadingBot: {@code .gmg}
     * obf: t
     *
     * <p>In the NPC rendering loop, {@code fb.c[ta.t]} (sprite array) and
     * {@code b.h[ta.t]} (model-height array) are indexed by this field,
     * confirming it carries the NPC's <em>visual</em> definition index for rendering.
     * For players this value is 0 (oracle: {@code serverId = 0}).
     */
    int serverId;

    // =========================================================================
    // Position / movement fields
    // =========================================================================

    /**
     * Current interpolated world X coordinate (fixed-point; divide by 128 to get
     * the tile column).  Advanced each tick toward {@code waypointsX[waypointCurrent]}.
     * Corresponds to oracle {@code GameCharacter.currentX}.
     * LeadingBot: {@code .gmh}
     * obf: i
     */
    int currentX;

    /**
     * Current interpolated world Y coordinate (fixed-point; divide by 128 to get
     * the tile row).  Advanced each tick toward {@code waypointsY[waypointCurrent]}.
     * Corresponds to oracle {@code GameCharacter.currentY}.
     * LeadingBot: {@code .gmi}
     * obf: K
     */
    int currentY;

    /**
     * Circular waypoint-queue head index — the slot in {@link #waypointsX} /
     * {@link #waypointsY} for the <em>next</em> destination tile.  Wraps modulo 10
     * as waypoints are consumed.
     * Corresponds to oracle {@code GameCharacter.waypointCurrent}.
     * LeadingBot: {@code .gna}
     * obf: o
     */
    int waypointCurrent;

    /**
     * Circular queue of destination X positions (10 slots).  The server pushes new
     * tiles onto the tail; the client pops from the head as the entity arrives.
     * Corresponds to oracle {@code GameCharacter.waypointsX}.
     * LeadingBot: {@code .gnb[]}
     * obf: k
     */
    int[] waypointsX = new int[10];

    /**
     * Circular queue of destination Y positions (10 slots).
     * Corresponds to oracle {@code GameCharacter.waypointsY}.
     * LeadingBot: {@code .gnc[]}
     * obf: F
     */
    int[] waypointsY = new int[10];

    /**
     * Sub-tile movement step counter.  Incremented each tick while the entity
     * walks toward its current waypoint.  Selects the walk-cycle frame:
     * {@code movingStep / 6 % 4}.  Reset to 0 when the entity is teleported.
     * Corresponds to oracle {@code GameCharacter.movingStep}.
     * LeadingBot: {@code .gmk}
     * obf: x
     */
    int movingStep;

    /**
     * Waypoint-arrival index.  Updated when the entity reaches its current waypoint
     * to advance {@link #waypointCurrent} to the next slot (modulo 10).
     * Corresponds to oracle {@code GameCharacter.stepCount}.
     * obf: e
     */
    int stepCount;

    // =========================================================================
    // Animation state
    // =========================================================================

    /**
     * Current animation ID (walk-direction + combat state).
     * <ul>
     *   <li>0–7 = walk/stand directions (N, NE, E, SE, S, SW, W, NW)
     *   <li>8   = combat attack pose
     *   <li>9   = combat block pose
     * </ul>
     * Updated each tick by the movement engine; checked by rendering to pick sprites.
     * Corresponds to oracle {@code GameCharacter.animationCurrent}.
     * LeadingBot: {@code .gml}
     * obf: y
     */
    int animationCurrent;

    /**
     * Pending animation ID — set by the server-update packet and latched into
     * {@link #animationCurrent} when the entity snaps to a new position.
     * Corresponds to oracle {@code GameCharacter.animationNext}.
     * LeadingBot: {@code .gmm}
     * obf: D
     */
    int animationNext;

    // =========================================================================
    // Entity type / appearance
    // =========================================================================

    /**
     * NPC definition ID (index into the NPC data table) <em>for NPC entities</em>,
     * or the {@code colourBottom} appearance byte for player entities.
     *
     * <p><b>NPC use:</b> the rendering loop checks {@code npcIdOrColourBottom == 255}
     * as a sentinel meaning "this slot has no valid NPC yet" and skips rendering.
     * The NPC model and sprite dimensions are looked up via {@link #serverId}
     * ({@code fb.c[serverId]}, {@code b.h[serverId]}); this field (the NPC def ID) is
     * used for data-table lookups (name strings, combat stats, etc.).
     *
     * <p><b>Player use:</b> the player-appearance packet writes the 3rd colour byte
     * (bottom/leg colour index) into this field.
     *
     * <p>Corresponds to oracle {@code GameCharacter.npcId} (NPC path) /
     * {@code GameCharacter.colourBottom} (player path).
     * LeadingBot: {@code .gmj} (npcId)
     * obf: A
     */
    int npcIdOrColourBottom;

    /**
     * Equipped-item / appearance slot array (12 slots, players only).
     * Each entry is a sprite/model ID for a worn-item slot
     * (head, cape, amulet, weapon, body, shield, legs, gloves, boots, …).
     * Zero means the slot is empty.  Read from the player-appearance packet
     * after {@link #chatSenderName} and {@link #message}.
     * Corresponds to oracle {@code GameCharacter.equippedItem[]}.
     * LeadingBot: {@code .gnd[]}
     * obf: m
     */
    int[] equippedItem = new int[12];

    /**
     * Hair colour index (palette index for the character's head/hair colour).
     * 1st colour byte in the player-appearance packet (after the equipment array).
     * Corresponds to oracle {@code GameCharacter.colourHair}.
     * LeadingBot: {@code .gnn}
     * obf: p
     */
    int colourHair;

    /**
     * Top/body colour index (2nd appearance colour byte).
     * Corresponds to oracle {@code GameCharacter.colourTop}.
     * LeadingBot: {@code .haa}
     * obf: q
     */
    int colourTop;

    /**
     * Skin colour index (4th appearance colour byte).
     * Corresponds to oracle {@code GameCharacter.colourSkin}.
     * LeadingBot: {@code .hac}
     * obf: H
     */
    int colourSkin;

    /**
     * Combat level of this entity (0–123; shown above the character in some contexts).
     * 5th byte in the player-appearance packet.  Initialised to -1 in the constructor
     * as a sentinel meaning "appearance not yet received".
     * Corresponds to oracle {@code GameCharacter.level}.
     * obf: s  (default -1)
     */
    int level = -1;   // obf: s; -1 = appearance not yet received

    /**
     * Server index of the <em>player</em> that is currently attacking this entity
     * (-1 if no player attacker).  The projectile renderer's companion to
     * {@link #attackingNpcServerIndex}.
     *
     * <p>In the projectile-attack sub-packets (entity-update block):
     * <ul>
     *   <li>opcode 3 ("NPC attacks this entity") writes -1 here, and the NPC index
     *       into {@link #attackingNpcServerIndex}.
     *   <li>opcode 4 ("player attacks this entity") writes the attacking player's
     *       server index here, and -1 into {@link #attackingNpcServerIndex}.
     * </ul>
     * The renderer branches: when {@code attackingNpcServerIndex == -1} and
     * {@code attackingPlayerServerIndex != -1}, the attacker (projectile source) is
     * looked up from the player array {@code client.We[attackingPlayerServerIndex]}.
     * (Verified in the clean-base projectile loop:
     * {@code if (var27.h == -1) { if (var27.z != -1) var37 = this.We[var27.z]; }
     *  else var37 = this.te[var27.h];}.)
     *
     * <p>Corresponds to oracle {@code GameCharacter.attackingPlayerServerIndex}.
     * obf: z
     */
    int attackingPlayerServerIndex;   // obf: z

    /**
     * Skull-visible flag.
     * <ul>
     *   <li>0 = not skulled (default)
     *   <li>positive = skulled (set from the 6th appearance byte in the player-
     *       appearance packet, opcode 5).
     * </ul>
     * Rendering checks {@code skullVisible != 0} (clean-base: {@code ~var9.J == -2},
     * i.e. {@code J == 1}) to draw the skull overlay above the head.
     *
     * <p>This is the 6th and final byte read in the appearance sub-packet, after
     * {@link #level} (5th byte), exactly matching the oracle's
     * {@code colourHair, colourTop, colourBottom, colourSkin, level, skullVisible}
     * ordering.  (The earlier mislabel as "combatDisplayLevel" was incorrect: obf
     * {@code J} is the skull byte, not a combat-level counter.)
     *
     * <p>Corresponds to oracle {@code GameCharacter.skullVisible}.
     * obf: J
     */
    int skullVisible;   // obf: J

    // =========================================================================
    // Name / overhead chat
    // =========================================================================

    /**
     * Entity display name (e.g. "Zezima").  For players, decoded via
     * {@code ia.a(buffer, false)} in the "send name" sub-packet (opcode 6 of the
     * entity-update block).  Also used as the sender name for overhead messages
     * when the entity is the local player.
     * Corresponds to oracle {@code GameCharacter.name}.
     * LeadingBot: {@code .gme}
     * obf: n
     */
    String name;

    /**
     * Secondary display name / clan string.  Received in the chat-message update
     * packet (opcode 5 of the entity-update block) alongside {@link #message}.
     * For the local player ({@code client.wi}), passed to {@code client.a()} as the
     * "sender" string.  May serve as a title or clan tag in this revision.
     *
     * <p>Uncertainty: this field has no direct oracle equivalent; the oracle uses
     * a single {@code name} field.  Its exact protocol role in rev ~233-235
     * is uncertain.
     *
     * obf: c
     */
    String chatSenderName;   // obf: c

    /**
     * Overhead chat message string.  Non-null while {@link #messageTimeout} > 0.
     * Corresponds to oracle {@code GameCharacter.message}.
     * LeadingBot: {@code .gne}
     * obf: C
     */
    String message;

    /**
     * Overhead message display countdown (ticks).  Set to 150 when a chat/overhead
     * message is received (clean-base: {@code var103.I = 150;}); decremented once per
     * tick in the entity-update loop alongside {@link #bubbleTimeout} and
     * {@link #combatTimer} (clean-base: {@code if (0 < var3.I) var3.I--;}); the
     * message is hidden when it reaches 0.
     *
     * <p><b>Naming correction:</b> obf {@code I}, NOT {@code h}.  The earlier deob
     * mislabelled this as {@code obf: h} (which is actually
     * {@link #attackingNpcServerIndex}) and mislabelled obf {@code I} as a phantom
     * "attackingNpcTimer".  Verified against the per-tick decrement loop and the
     * oracle's {@code messageTimeout--} ordering.
     *
     * Corresponds to oracle {@code GameCharacter.messageTimeout}.
     * LeadingBot: {@code .gnf}
     * obf: I
     */
    int messageTimeout;   // obf: I

    // =========================================================================
    // Item bubble (overhead item pop-up)
    // =========================================================================

    /**
     * Item ID shown in the overhead item-bubble sprite, or 0 if none active.
     * Corresponds to oracle {@code GameCharacter.bubbleItem}.
     * obf: j
     */
    int bubbleItem;

    /**
     * Item-bubble display countdown (ticks).  Set to 150 when a bubble is received;
     * the bubble sprite is drawn while this is greater than zero.
     * Corresponds to oracle {@code GameCharacter.bubbleTimeout}.
     * obf: E
     */
    int bubbleTimeout;

    // =========================================================================
    // Combat / projectile state
    // =========================================================================

    /**
     * Combat-engage timer (countdown ticks).  Set to 200 when this entity is hit
     * or attacks; enables health-bar and skull-overlay rendering while positive.
     * Decremented once per tick in the entity-update loop.
     * Corresponds to oracle {@code GameCharacter.combatTimer}.
     * LeadingBot: {@code .gnl}
     * obf: d
     */
    int combatTimer;

    /**
     * Damage taken in the most recent hit as reported by the server (raw value).
     * Rendered as a hit-splat number above the entity.
     * Corresponds to oracle {@code GameCharacter.damageTaken}.
     * LeadingBot: {@code .gni}
     * obf: u
     */
    int damageTaken;

    /**
     * Current HP (remaining health) as reported by the server.
     * Used in the health-bar width calculation: {@code barWidth = 30 * healthCurrent / healthMax}.
     * Corresponds to oracle {@code GameCharacter.healthCurrent}.
     * LeadingBot: {@code .gnj}
     * obf: B
     */
    int healthCurrent;

    /**
     * Maximum HP as reported by the server.  Divisor in the health-bar formula.
     * Corresponds to oracle {@code GameCharacter.healthMax}.
     * LeadingBot: {@code .gnk}
     * obf: G
     */
    int healthMax;

    /**
     * Sprite index of the incoming projectile / spell currently flying toward this
     * entity.  Added to a base sprite offset to select the correct projectile graphic.
     * Set by both the "NPC attacks player" (opcode 3) and "player attacks player"
     * (opcode 4) sub-packets.
     * Corresponds to oracle {@code GameCharacter.incomingProjectileSprite}.
     * obf: a
     */
    int incomingProjectileSprite;

    /**
     * Server index of the NPC that is currently attacking this entity (-1 if none).
     * Used by the projectile-rendering loop to locate the attacker's position
     * (source point) for the projectile interpolation.
     *
     * <p>In the protocol: opcode 3 ("NPC attacks this entity") writes the NPC
     * server index here; opcode 4 ("player attacks this entity") writes -1 here.
     * The renderer branches on this field: if {@code attackingNpcServerIndex != -1},
     * the attacker is looked up from {@code client.te[attackingNpcServerIndex]} (NPC array);
     * otherwise the attacker is looked up from
     * {@code client.We[attackingPlayerServerIndex]} (player array).
     *
     * <p><b>Naming correction:</b> obf {@code h} is this field (verified by the
     * clean-base assignments {@code var103.h = npcIdx;} in opcode 3 and
     * {@code var103.h = -1;} in opcode 4, and the projectile-loop NPC-array lookup
     * {@code var37 = this.te[var27.h];}).  The earlier deob also wrongly declared a
     * second {@code messageTimeout} field under {@code obf: h}; that duplicate has
     * been removed (real {@code messageTimeout} is obf {@code I}).
     *
     * Corresponds to oracle {@code GameCharacter.attackingNpcServerIndex}.
     * obf: h
     */
    int attackingNpcServerIndex;   // obf: h

    /**
     * Projectile animation progress counter.  Set to {@code client.nc}
     * (the total projectile step count) when an attack packet is received; decremented
     * each tick while the projectile animates between source and target.
     *
     * <p>Interpolated position formula (oracle equivalent):
     * <pre>
     *   rx = (sourceX * projectileRange + targetX * (nc - projectileRange)) / nc
     * </pre>
     * where {@code nc} is the total step count stored in {@code client.nc}.
     *
     * Corresponds to oracle {@code GameCharacter.projectileRange}.
     * obf: w
     */
    int projectileRange;

    // NOTE: obf field {@code I} is {@link #messageTimeout} (declared above in the
    // name/overhead-chat section), NOT an "attackingNpcTimer".  The earlier deob's
    // phantom {@code attackingNpcTimer} field (obf I, "set to 150") was a duplicate
    // misread of the {@code messageTimeout = 150} write and has been removed.

    // =========================================================================
    // Constructor
    // =========================================================================

    /**
     * Initialises a new GameCharacter with safe defaults.
     * {@link #level} is pre-set to -1 as a sentinel meaning "appearance data not
     * yet received from server".  All timer fields default to 0 (inactive).
     *
     * <p>Original explicit initialisations (obf names):
     * <pre>
     *   this.s  = -1;   // level = -1 (appearance sentinel)
     *   this.J  = 0;    // skullVisible = 0
     *   this.z  = 0;    // attackingPlayerServerIndex = 0
     *   this.u  = 0;    // damageTaken = 0
     *   this.h  = 0;    // attackingNpcServerIndex = 0 (initially "no attacker")
     *   this.a  = 0;    // incomingProjectileSprite = 0
     *   this.w  = 0;    // projectileRange = 0
     *   this.I  = 0;    // messageTimeout = 0
     *   this.B  = 0;    // healthCurrent = 0
     *   this.E  = 0;    // bubbleTimeout = 0
     *   this.G  = 0;    // healthMax = 0
     *   this.d  = 0;    // combatTimer = 0
     * </pre>
     * obf: ta()
     */
    GameCharacter() {
        // level = -1 is the only non-zero initial value; all others are Java defaults (0/null).
        // Java's field default initialisation already sets all int fields to 0, so the
        // explicit zeroing in the original constructor was purely obfuscator output.
    }

    // =========================================================================
    // Static utility methods (unrelated to entity state — helpers placed here
    // by the obfuscator; logically belong to other classes)
    // =========================================================================

    /**
     * Loads {@code length} bytes from a {@link client.data.DataStore} archive entry
     * named {@code resourceName} into {@code dest[0..length-1]}.
     *
     * <p>Note: this method has no logical connection to GameCharacter entity state.
     * The obfuscator placed it here along with the other static helper.
     *
     * <p><b>Stripped obfuscation:</b>
     * <ul>
     *   <li>Anti-tamper guard: {@code if (magicParam != -19675) CHAT_CIPHER_METADATA = null;} —
     *       removed.  All callers always supply {@code -19675}; the guard nulls the
     *       ChatCipher field (which is irrelevant) if tampered with.
     *   <li>Profiling counter: {@code ++PROFILING_COUNTER_readFromStore} — removed.
     *   <li>Outer {@code catch(RuntimeException)} error-reporter — unwrapped.
     *   <li>Inner {@code catch(EOFException)} around {@code readFully} is preserved:
     *       short reads are silently swallowed; only I/O errors propagate.
     * </ul>
     *
     * @param resourceName  archive entry name key (passed to {@code DataStore.openStream})
     * @param magicParam    anti-tamper constant (always {@code -19675}; param kept for
     *                      ABI compatibility but the guard has been removed)
     * @param dest          output buffer
     * @param length        number of bytes to read into {@code dest}
     * @throws IOException  on underlying I/O error
     * obf: a(String, int, byte[], int)
     */
    static void readFromStore(String resourceName, int magicParam, byte[] dest, int length)
            throws IOException {
        // obf: nb.a(true, resourceName) — DataStore.openStream(boolean, String)
        InputStream stream = client.data.DataStore.openStream(true, resourceName);
        DataInputStream dataIn = new DataInputStream(stream);
        try {
            dataIn.readFully(dest, 0, length);
        } catch (EOFException ignored) {
            // Short read: original silently ignores EOF before 'length' bytes.
        }
        dataIn.close();
    }

    /**
     * Counts the number of 1-bits in {@code value} (32-bit Hamming weight / popcount)
     * using the classic parallel-summation algorithm.
     *
     * <p>Note: this method has no logical connection to GameCharacter entity state.
     * The obfuscator placed it here.
     *
     * <p><b>Algorithm</b> (in readable form):
     * <pre>
     *   // Step 1: sum adjacent bit-pairs into 2-bit counts
     *   v = (v >>> 1 &amp; 0x55555555) + (v &amp; 0x55555555);
     *   // Step 2: sum adjacent 2-bit counts into 4-bit nibble counts
     *   v = ((v >>> 2) &amp; 0x33333333) + (v &amp; 0x33333333);
     *   // Step 3: sum nibble counts into byte counts
     *   //   Note: original writes "v - -(v >>> 4)" which equals "v + (v >>> 4)"
     *   v = (v + (v >>> 4)) &amp; 0x0F0F0F0F;
     *   // Step 4: sum bytes (parallel horizontal add)
     *   v += v >>> 8;
     *   v += v >>> 16;
     *   return v &amp; 0xFF;   // result in [0, 32]
     * </pre>
     *
     * <p><b>Stripped obfuscation:</b>
     * <ul>
     *   <li>Anti-tamper junk: {@code int junk = 110 / ((65 - dummyByte) / 44);} removed.
     *       The result is always discarded; ArithmeticException if {@code dummyByte == 65}.
     *       Callers always pass a safe constant.
     *   <li>Profiling counter: {@code ++PROFILING_COUNTER_popcount} — removed.
     *   <li>Outer {@code catch(RuntimeException)} error-reporter — unwrapped.
     *   <li>CFR decompiler reported shift amounts as large negative constants
     *       (e.g. {@code >>> -2016269343}) due to sign extension;
     *       Vineflower normalises these to the true shifts (1, 2, 4, 8, 16).
     * </ul>
     *
     * @param value      the 32-bit integer whose set bits are to be counted
     * @param dummyByte  anti-tamper dummy (always a safe non-65 constant; guard removed)
     * @return           number of 1-bits in {@code value}, in range [0, 32]
     * obf: a(int, byte)
     */
    static int popcount(int value, byte dummyByte) {
        // Step 1: sum adjacent bit-pairs → 2-bit counts  (mask 0x55555555 = 0101…)
        value = (0x55555555 & value >>> 1) + (0x55555555 & value);
        // Step 2: sum 2-bit groups → 4-bit nibble counts  (mask 0x33333333 = 0011…)
        value = ((value & 0xCCCCCCCC) >>> 2) + (0x33333333 & value);
        // Step 3: sum nibbles → byte counts; original uses subtraction of negation
        //   "value - -(value >>> 4)" == "value + (value >>> 4)"
        //   then mask to isolate the byte-wide sums  (mask 0x0F0F0F0F)
        value = (value + (value >>> 4)) & 0x0F0F0F0F;
        // Step 4: horizontal byte-add using shifts
        value += value >>> 8;
        value += value >>> 16;
        // Low byte holds the final popcount in [0..32]
        return 0xFF & value;
    }

    // =========================================================================
    // XOR string-pool helpers (obfuscation infrastructure — not real logic)
    // =========================================================================

    /**
     * Stage 1 of the per-class XOR string decoder: String → char[].
     * If the string has fewer than 2 characters the single character is XOR-ed
     * with {@code 0x3D} ('=') before being returned as a char array.
     * obf: z(String)
     */
    private static char[] z(String s) {
        char[] chars = s.toCharArray();
        if (chars.length < 2) {
            chars[0] = (char)(chars[0] ^ '=');   // 0x3D
        }
        return chars;
    }

    /**
     * Stage 2 of the per-class XOR string decoder: char[] → interned String.
     * XORs each character with the rotating 5-byte key table
     * {@code {75='K', 95='_', 65='A', 90='Z', 61='='}} (indexed by position mod 5).
     * obf: z(char[])
     */
    private static String z(char[] chars) {
        // Key table: positions 0–4 cycle through these XOR constants
        final int[] KEY = {75, 95, 65, 90, 61};
        for (int i = 0; i < chars.length; i++) {
            chars[i] = (char)(chars[i] ^ KEY[i % 5]);
        }
        return new String(chars).intern();
    }
}
