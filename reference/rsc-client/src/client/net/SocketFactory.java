package client.net;

import java.awt.image.ColorModel;
import java.io.IOException;
import java.net.Socket;

/**
 * SocketFactory — abstract socket factory base class (obf: {@code m}).
 *
 * <p>Holds the target hostname ({@link #host}) and port ({@link #port}) for the concrete
 * subclass {@code ProxySocketFactory} (obf: {@code gb}) to open a real TCP socket toward
 * the game server.  The concrete implementation is selected by
 * {@code StreamFactory.createSocketFactory(int, int, String)} (obf: {@code na.a(int,int,String)}),
 * which constructs a {@code ProxySocketFactory} and sets {@code host}/{@code port} before
 * returning the abstract reference.
 *
 * <p>In addition to the factory contract this class is a grab-bag of global static state that
 * was convenient to park here during obfuscation:
 * <ul>
 *   <li>{@link #fontGlyphData} — 50-slot raw glyph/sprite byte arrays used by the UI.</li>
 *   <li>{@link #itemSpriteIndex} — per-item sprite-index array (size = item count), used by the
 *       character-renderer to look up which sprite to draw for an equipped item.</li>
 *   <li>{@link #globalArchive} — the primary JAG archive reader ({@code ArchiveReader} / obf
 *       {@code ob}) used for lazy resource lookups by {@code StreamBase} (obf: {@code ib}).</li>
 *   <li>{@link #packetSizeByOpcode} — 256-entry profiling accumulator: for each packet opcode
 *       (0–255), the total byte count seen so far; updated in {@code Packet} (obf: {@code b}).</li>
 *   <li>{@link #globalColorModel} — the single shared 8-bit {@code IndexColorModel} built by
 *       {@code ImageLoader} (obf: {@code pa}) and consumed by {@code Scene} (obf: {@code lb}) and
 *       {@code SurfaceImageProducer} (obf: {@code fb}).</li>
 *   <li>Profiling counters {@link #profilingCounterA} and {@link #profilingCounterK} incremented
 *       by the obfuscator near every method entry — treated as dead weight here.</li>
 * </ul>
 *
 * <p>The large static method {@link #initGameData(byte[], byte, boolean)} allocates all per-entity
 * arrays for items, NPCs, objects, textures, prayers, spells, and sounds by reading the
 * {@code integer.dat} and {@code string.dat} data archives.  It is called exactly once from
 * {@code Mudclient} (obf: {@code client}) during startup.
 *
 * <p>Inheritance: {@code m} (SocketFactory) → {@code gb} (ProxySocketFactory).
 */
abstract class SocketFactory {

    // -------------------------------------------------------------------------
    // Static fields shared across the engine (parked here by the obfuscator)
    // -------------------------------------------------------------------------

    /**
     * Raw glyph/sprite byte arrays; slot 0..49.  Each entry is one font glyph's packed pixel data.
     * Used by {@code Surface} (obf: {@code ua}) and {@code Panel} (obf: {@code qa}).
     * obf: {@code m.b}
     */
    static byte[][] fontGlyphData = new byte[50][];

    /**
     * Per-item sprite-index array (length = total item count, set during {@link #initGameData}).
     * Indexed by {@code GameCharacter.equipSlot} (obf: {@code ta.t}) to resolve which equipped-item
     * sprite the character renderer should draw.
     * obf: {@code m.g}
     */
    static int[] itemSpriteIndex;

    /**
     * Profiling counter incremented near every call to {@link #readInt32BE}.
     * Dead weight left by the obfuscator — never read for game logic.
     * obf: {@code m.k}
     */
    static int profilingCounterK;

    /**
     * The hostname this factory will connect to.  Set by {@code StreamFactory} after construction.
     * obf: {@code m.h}
     */
    String host;

    /**
     * The primary JAG archive reader instance, used by {@code StreamBase} (obf: {@code ib}) for
     * lazy cache lookups.  Populated by {@code CacheUpdater} (obf: {@code cb}).
     * obf: {@code m.e}
     */
    static ArchiveReader globalArchive = null;   // type: ob → ArchiveReader

    /**
     * Per-opcode packet byte-size accumulator (256 entries, one per possible opcode byte).
     * Updated in {@code Packet.processIncomingPacket} for profiling.  Never reset during play.
     * obf: {@code m.i}
     */
    static int[] packetSizeByOpcode = new int[256];

    /**
     * The TCP port this factory will connect to.  Set by {@code StreamFactory} after construction.
     * obf: {@code m.f}
     */
    int port;

    /**
     * Profiling counter incremented near every call to {@link #initGameData}.
     * Dead weight left by the obfuscator — never read for game logic.
     * obf: {@code m.a}
     */
    static int profilingCounterA;

    /**
     * Shared 8-bit {@code IndexColorModel} built once by {@code ImageLoader} (obf: {@code pa})
     * and consumed by {@code Scene} (obf: {@code lb}) and
     * {@code SurfaceImageProducer} (obf: {@code fb}).
     * obf: {@code m.d}
     */
    static ColorModel globalColorModel;

    /**
     * Profiling counter incremented near every call to {@link #initGameData}.
     * Dead weight left by the obfuscator — never read for game logic.
     * obf: {@code m.c}
     */
    static int profilingCounterC;

    /**
     * Minimum Y value tracked during scene rendering (camera/frustum hint).
     * Used by {@code Scene} (obf: {@code lb}) and {@code GameModel} (obf: {@code ca}).
     * obf: {@code m.j}
     */
    static int minRenderY;

    // -------------------------------------------------------------------------
    // XOR-decoded string pool (obf: private static final String[] z)
    // -------------------------------------------------------------------------
    // Decoded via two-stage XOR: z(String) XORs with 0x5E if len<2; z(char[]) XORs with
    // the mod-5 key table {64, 20, 39, 57, 94}.
    //
    // z[0] = "null"           — null-array label in error messages
    // z[1] = "{...}"          — non-null array label in error messages
    // z[2] = "m.A("           — error prefix for readInt32BE
    // z[3] = "m.B("           — error prefix for initGameData
    // z[4] = "integer.dat"    — integer data archive filename / XOR key
    // z[5] = "You need to be a member to use this object"
    // z[6] = "Members object" — placeholder name for members-only items
    // z[7] = "string.dat"     — string data archive filename / XOR key
    // z[8] = "m.C("           — error prefix for openSocket

    private static final String[] STRINGS;

    static {
        STRINGS = new String[]{
            decodeString(decodeChars(".aKU")),                                  // [0] "null"
            decodeString(decodeChars(";:\t#")),                           // [1] "{...}"
            decodeString(decodeChars("-:f")),                             // [2] "m.A("
            decodeString(decodeChars("-:e")),                             // [3] "m.B("
            decodeString(decodeChars(")zS\\9%f\t]?4")),                        // [4] "integer.dat"
            decodeString(decodeChars("{R0%qC*/4E\\~!4J\\3\"qU*/4RJ;``OP-`{ES;#`")), // [5] "You need to be a member to use this object"
            decodeString(decodeChars("\rqJ[;2gV<*qDM")),                 // [6] "Members object"
            decodeString(decodeChars("3`UP0':CX*")),                           // [7] "string.dat"
            decodeString(decodeChars("-:d")),                             // [8] "m.C("
        };
        fontGlyphData = new byte[50][];
        packetSizeByOpcode = new int[256];
        globalArchive = null;
    }

    // -------------------------------------------------------------------------
    // Constructor
    // -------------------------------------------------------------------------

    /** Default constructor. obf: {@code m()} */
    SocketFactory() {
    }

    // -------------------------------------------------------------------------
    // Abstract factory method — implemented by ProxySocketFactory
    // -------------------------------------------------------------------------

    /**
     * Opens and returns a new TCP {@code Socket} to the server.  The concrete implementation
     * ({@code ProxySocketFactory}) routes the connection through an optional proxy, falling back
     * to a direct {@code new Socket(host, port)} if no proxy is configured.
     *
     * <p>The {@code dummy} byte parameter is present in the bytecode but carries no semantics;
     * it is an anti-tamper artifact inserted by the obfuscator.
     *
     * @param dummy  ignored obfuscation parameter (obf: {@code var1})
     * @return a connected {@code Socket}
     * @throws IOException if the connection fails
     * obf: {@code abstract Socket a(byte)}
     */
    abstract Socket openSocket(byte dummy) throws IOException;

    // -------------------------------------------------------------------------
    // Concrete methods
    // -------------------------------------------------------------------------

    /**
     * Opens a direct (non-proxy) TCP socket to {@link #host}:{@link #port}.
     *
     * <p>This is the base implementation available to all subclasses.  The concrete override in
     * {@code ProxySocketFactory} (obf: {@code gb}) first attempts proxy-aware connection and
     * falls back to this path on failure.
     *
     * <p>Obfuscation notes stripped:
     * <ul>
     *   <li>Opaque predicate {@code if(dummy) i=(int[])null;} — dead branch removed.</li>
     *   <li>Profiling {@code ++profilingCounterK;} — removed.</li>
     *   <li>Exception wrapper {@code catch(RuntimeException){throw ErrorHandler.a(e,"m.C("+dummy+')')}} — unwrapped.</li>
     * </ul>
     *
     * @param dummy  ignored obfuscation parameter; always false in practice (obf: {@code var1 / bl})
     * @return a new {@code Socket} connected to {@code host}:{@code port}
     * @throws IOException if the TCP connection cannot be established
     * obf: {@code final Socket a(boolean)}
     */
    final Socket openSocketDirect(boolean dummy) throws IOException {
        // obf stripped: ++profilingCounterK (profiling counter, dead)
        // obf stripped: if(dummy) packetSizeByOpcode=(int[])null; (opaque predicate, always false)
        return new Socket(this.host, this.port);
    }

    /**
     * Reads a big-endian 32-bit signed integer from a byte array at position {@code offset}.
     *
     * <p>Equivalent to:
     * <pre>
     *   ((data[offset]   &amp; 0xFF) &lt;&lt; 24)
     * | ((data[offset+1] &amp; 0xFF) &lt;&lt; 16)
     * | ((data[offset+2] &amp; 0xFF) &lt;&lt;  8)
     * | ( data[offset+3] &amp; 0xFF)
     * </pre>
     *
     * <p>The CFR decompiler renders the shifts as large negative constants because of J++ compiler
     * arithmetic on the shift amount; the Vineflower output shows
     * {@code var2[var1+3] & 0xFF + (var2[var1-(-2)] << 8 & 0xFF00) + ...}
     * which is the same big-endian assembly.
     *
     * <p>Obfuscation notes stripped:
     * <ul>
     *   <li>Opaque predicate {@code if(!dummy) fontGlyphData=null;} — dead branch removed.</li>
     *   <li>Profiling {@code ++profilingCounterA;} — removed.</li>
     *   <li>Exception wrapper — unwrapped.</li>
     * </ul>
     *
     * @param dummy   ignored obfuscation parameter (obf: {@code var0 / bl}); always false
     * @param offset  byte offset into {@code data}
     * @param data    source byte array
     * @return the 32-bit big-endian integer at {@code data[offset..offset+3]}
     * obf: {@code static final int a(boolean, int, byte[])}
     */
    static final int readInt32BE(boolean dummy, int offset, byte[] data) {
        // obf stripped: if(!dummy) fontGlyphData=null; (opaque predicate, always false)
        // obf stripped: ++profilingCounterA; (profiling counter, dead)

        // Big-endian reassembly; the obfuscator used negative shift constants to obscure this.
        // CFR renders: (data[offset+3]&0xFF) + ((data[offset+2]<<-1719048248)&0xFF00)
        //               + ((0xFF & data[offset])<<-523457256)
        //               - -((0xFF & data[offset+1])<<-728162096)
        // These odd-looking negated shift amounts are equivalent modulo 32 to 8, 24, 16:
        //   -1719048248 % 32 = 8,  -523457256 % 32 = 24,  -728162096 % 32 = 16
        return ((data[offset]     & 0xFF) << 24)
             | ((data[offset + 1] & 0xFF) << 16)
             | ((data[offset + 2] & 0xFF) <<  8)
             |  (data[offset + 3] & 0xFF);
    }

    /**
     * One-time game-data initializer — allocates all global per-entity arrays from the
     * {@code integer.dat} and {@code string.dat} data archives.
     *
     * <p>Called exactly once from {@code Mudclient} (obf: {@code client}) with
     * {@code m.a(byteArray, (byte)100, this.Pg)}.  The {@code rawData} array contains a
     * concatenation of the two data archives (XOR-decoded).  The second parameter {@code param}
     * carries the value {@code 100} in the only known call; when it is {@code < 10} the method
     * returns early after allocating only the first-tier arrays (items/NPCs/objects).
     *
     * <p>The method populates dozens of static arrays across the following classes.  Names below
     * use the canonical map; obfuscated field names are noted parenthetically.
     *
     * <h3>Tier 1 — item/NPC/object/texture tables (always allocated)</h3>
     * <ul>
     *   <li>{@code InputState.loadedStringData} (obf {@code kb.d}) — decoded string archive bytes
     *       from {@code string.dat} via {@code StreamFactory.decodeArchive}.</li>
     *   <li>{@code InputState.loadedIntData} (obf {@code b.v}) — decoded integer archive bytes
     *       from {@code integer.dat} via {@code StreamFactory.decodeArchive}.</li>
     *   <li>Item arrays (count = {@code EntityDef.itemCount} / obf {@code gb.p}):
     *     <ul>
     *       <li>{@code ClientIOException.itemY} (obf {@code fa.e}) — item Y-offsets.</li>
     *       <li>{@code DecodeBuffer.itemNames} (obf {@code ac.x}) — item name strings
     *           (pre-filled with empty-string sentinels via {@code ISAAC.emptyString}
     *            / obf {@code o.a(byte 38)}).</li>
     *       <li>{@code IntHolder.itemMembers} (obf {@code ka.c}) — members-only flag (1 = members).</li>
     *       <li>{@code TextEncoder.itemColors} (obf {@code h.c}) — item tint colors.</li>
     *       <li>{@code Scene.itemDescriptions} (obf {@code lb.ac}) — item examine strings.</li>
     *       <li>{@code Surface.itemBb} (obf {@code ua.Bb}) — item sprite widths.</li>
     *       <li>{@code InputState.itemFlags} (obf {@code kb.b}) — item flags.</li>
     *       <li>{@code Utility.itemK} (obf {@code mb.k}) — item unknown-K values.</li>
     *       <li>{@code InputState.itemC} (obf {@code kb.c}) — item team/category (1 = stackable).</li>
     *       <li>{@code CharTable.itemNotes} (obf {@code ga.b}) — item examine strings (alternate).</li>
     *       <li>{@code ProxySocketFactory.itemGsFlags} (obf {@code gb.s}) — item ground-sprite flags.</li>
     *     </ul>
     *   </li>
     *   <li>Members-only items: items whose {@code ka.c[i]} == 1 (the obfuscated test is
     *       {@code -2 == ~ka.c[i]}) get their name set to
     *       {@code "Members object"} (z[6]), examine set to
     *       {@code "You need to be a member to use this object"} (z[5]), flags zeroed,
     *       description cleared, and category set to 1.</li>
     * </ul>
     *
     * <h3>Tier 2 — entity/equipment tables (allocated after tier 1)</h3>
     * <p>Builds arrays of size {@code la.d} (entity/slot count returned by
     * {@code EntityDef.a(65525)} / obf {@code t.a}), covering:
     * {@code fb.d, b.h, jb.k, ob.h, la.a}, {@link #itemSpriteIndex} ({@code m.g}),
     * {@code v.e, o.a, ba.ac, fb.c, p.e, da.T, e.Mb, na.a, db.j, qb.d[][12], eb.b, ua.Ab}.
     *
     * <h3>Tier 3 — spell/prayer/sound tables</h3>
     * <p>NPCs ({@code jb.o}), prayers ({@code na.e}), spells ({@code ua.Db}),
     * equipment slots ({@code h.a}), render sizes ({@code v.h}), textures ({@code ia.h}),
     * objects/map ({@code n.c, fa.b}), and sounds ({@code t.g}).
     *
     * <p><b>Early return:</b> if {@code param < 10} the method returns after tier 1 and the
     * beginning of tier 2 (up to the {@code ua.Ab} loop).  The only known caller passes
     * {@code 100}, so all tiers are always executed in practice.
     *
     * <p>Obfuscation notes stripped:
     * <ul>
     *   <li>All {@code if(bl)/while(!bl)/break} opaque-predicate dead code removed;
     *       real loop bodies preserved.</li>
     *   <li>Profiling {@code ++m.c;} removed.</li>
     *   <li>Exception wrapper with {@code z[3]="m.B("} signature — unwrapped.</li>
     *   <li>Anti-tamper dummy parameter {@code dummy} (always {@code false}) noted but removed.</li>
     *   <li>At method end: {@code b.v = null; kb.d = null;} — clears the raw archive byte arrays
     *       now that all data has been parsed into the proper typed arrays.</li>
     * </ul>
     *
     * @param rawData  concatenated XOR-decoded archive bytes (obf: {@code param0 / var0})
     * @param param    controls depth of initialization; value 100 = full init (obf: {@code param1 / var1})
     * @param dummy    opaque-predicate flag, always false in practice (obf: {@code param2 / var2})
     * obf: {@code static final void a(byte[], byte, boolean)}
     */
    static final void initGameData(byte[] rawData, byte param, boolean dummy) {
        // obf stripped: boolean bl = client.vh; (opaque predicate, always false)
        // obf stripped: ++profilingCounterC; (profiling counter, dead)

        // --- Decode data archives ---
        // z[7] = "string.dat", key byte -124; result stored in InputState.loadedStringData (kb.d)
        // Calls StreamFactory.decodeArchive(String key, int offset, byte[] raw, int keyByte)
        InputState.loadedStringData = StreamFactory.decodeArchive(
            /* key: */ "string.dat", /* z[7] */
            /* offset: */ 0,
            rawData,
            /* keyByte: */ -124
        );
        // obf stripped: ++profilingCounterC;

        DownloadWorker.readCursor = 0; // obf: jb.p = 0  (reset the string-data read cursor)

        // z[4] = "integer.dat", key byte -125; result stored in Packet.loadedIntData (b.v)
        Packet.loadedIntData = StreamFactory.decodeArchive(
            /* key: */ "integer.dat", /* z[4] */
            /* offset: */ 0,
            rawData,
            /* keyByte: */ -125
        );

        // CRITICAL: reset the shared integer-data read cursor (IntHolder.bufferOffset / ka.b)
        // to 0 BEFORE the parsing loops below.  Every EntityDef.getCount/readSpriteId/decode call
        // advances this cursor as it consumes bytes from Packet.loadedIntData (b.v); without this
        // reset the whole game-data load reads from the wrong offset.  (clean base: ka.b = 0)
        IntHolder.bufferOffset = 0; // obf: ka.b = 0

        // --- Tier 1: allocate item arrays ---
        // ProxySocketFactory.itemCount = EntityDef.getCount(65525)  (obf: gb.p = t.a(65525))
        int itemCount = EntityDef.getCount(65525); // obf: gb.p = t.a(65525)
        ProxySocketFactory.itemCount = itemCount;

        // Allocate all per-item arrays (size = itemCount)
        ClientIOException.itemY          = new int[itemCount];         // obf: fa.e
        DecodeBuffer.itemNames           = new String[itemCount];      // obf: ac.x
        IntHolder.itemMembers            = new int[itemCount];         // obf: ka.c (members flag; ==1 → members-only)
        TextEncoder.itemColors           = new int[itemCount];         // obf: h.c
        Scene.itemDescriptions           = new String[itemCount];      // obf: lb.ac (Scene.diagStrings slot)
        Surface.itemBb                   = new int[itemCount];         // obf: ua.Bb
        InputState.itemFlags             = new int[itemCount];         // obf: kb.b
        Utility.itemK                    = new int[itemCount];         // obf: mb.k
        InputState.itemCategory          = new int[itemCount];         // obf: kb.c
        CharTable.itemNotes              = new String[itemCount];      // obf: ga.b
        ProxySocketFactory.itemGsFlags   = new int[itemCount];         // obf: gb.s

        // Pre-fill string arrays with empty-string sentinels
        // obf: o.a(byte 38) returns an interned empty string via ISAAC.emptyString
        for (int i = 0; i < itemCount; i++) {
            DecodeBuffer.itemNames[i] = ISAAC.emptyString((byte) 38); // obf: o.a((byte)38)
        }
        for (int i = 0; i < itemCount; i++) {
            CharTable.itemNotes[i] = ISAAC.emptyString((byte) 38);
        }
        for (int i = 0; i < itemCount; i++) {
            Scene.itemDescriptions[i] = ISAAC.emptyString((byte) 38);
        }

        // Fill per-item numeric arrays from decoded archives
        for (int i = 0; i < itemCount; i++) {
            Surface.itemBb[i] = EntityDef.readSpriteId(65525); // obf: ua.Bb[i] = t.a(65525)
            // Track the global sprite count maximum
            if (Utility.maxSpriteId < Surface.itemBb[i] + 1) { // obf: mb.l
                Utility.maxSpriteId = Surface.itemBb[i] + 1;
            }
        }
        for (int i = 0; i < itemCount; i++) {
            InputState.itemFlags[i] = NameTable.lookupName((byte) -105); // obf: ub.a
        }
        for (int i = 0; i < itemCount; i++) {
            ClientIOException.itemY[i] = ChatCipher.decode(-30504); // obf: v.a(-30504)
        }
        for (int i = 0; i < itemCount; i++) {
            ProxySocketFactory.itemGsFlags[i] = ChatCipher.decode(-30504);
        }
        for (int i = 0; i < itemCount; i++) {
            Utility.itemK[i] = EntityDef.readSpriteId(65525);
        }
        for (int i = 0; i < itemCount; i++) {
            TextEncoder.itemColors[i] = NameTable.lookupName((byte) -105);
        }
        // NOTE: kb.c (itemCategory) is read from the stream BEFORE ka.c (itemMembers) —
        // the order is load-bearing because both consume sequential bytes from v.a(-30504).
        for (int i = 0; i < itemCount; i++) {
            InputState.itemCategory[i] = ChatCipher.decode(-30504); // obf: kb.c[i]
        }
        for (int i = 0; i < itemCount; i++) {
            IntHolder.itemMembers[i] = ChatCipher.decode(-30504); // obf: ka.c[i] (members flag)
        }

        // Patch members-only items: if itemMembers[i] == 1, replace name/examine with member
        // strings.  The clean base writes this as "-2 == ~ka.c[i]"; since ~1 == -2, that is the
        // obfuscated form of "ka.c[i] == 1".  Matches the oracle (GameData.loadData):
        //   if (!isMembers && itemMembers[i] == 1) { ... }
        // (The leading "1 != var2" guard is the always-true dummy-param test, stripped.)
        for (int i = 0; i < itemCount; i++) {
            if (/* not: dummy==1 */ true && IntHolder.itemMembers[i] == 1) {
                DecodeBuffer.itemNames[i]     = STRINGS[6]; // "Members object"
                CharTable.itemNotes[i]        = STRINGS[5]; // "You need to be a member to use this object"
                InputState.itemFlags[i]       = 0;
                Scene.itemDescriptions[i]     = "";
                ProxySocketFactory.itemGsFlags[0] = 0;  // Note: index 0, not i — matches bytecode
                Utility.itemK[i]              = 0;
                InputState.itemCategory[i]    = 1;
            }
        }

        // --- Tier 2: entity/equipment tables ---
        // la.d = EntityDef.getCount(65525)  (second call, may return different count)
        int entityCount = EntityDef.getCount(65525); // obf: la.d = t.a(65525)

        // Allocate per-entity arrays
        SurfaceImageProducer.equipSpriteD     = new int[entityCount];    // obf: fb.d
        Packet.equipH                         = new int[entityCount];    // obf: b.h
        DownloadWorker.equipK                 = new int[entityCount];    // obf: jb.k
        ArchiveReader.equipH                  = new int[entityCount];    // obf: ob.h
        ClientRuntimeException.equipA         = new int[entityCount];    // obf: la.a (la = ClientRuntimeException)
        itemSpriteIndex                       = new int[entityCount];    // obf: m.g
        ChatCipher.equipE                     = new int[entityCount];    // obf: v.e
        ISAAC.equipA                          = new int[entityCount];    // obf: o.a  (note: conflicts with o.a static method — field)
        SurfaceSprite.equipAc                 = new String[entityCount]; // obf: ba.ac
        SurfaceImageProducer.equipC           = new int[entityCount];    // obf: fb.c
        Timer.equipE                          = new String[entityCount]; // obf: p.e
        ClientStream.equipT                   = new int[entityCount];    // obf: da.T
        GameShell.equipMb                     = new String[entityCount]; // obf: e.Mb
        StreamFactory.equipA                  = new int[entityCount];    // obf: na.a
        LinkedQueue.equipJ                    = new int[entityCount];    // obf: db.j
        GameFrame.equipD                      = new int[entityCount][12]; // obf: qb.d  — 12-slot equipment grid
        AudioMixer.equipB                     = new int[entityCount];    // obf: eb.b
        Surface.equipAb                       = new int[entityCount];    // obf: ua.Ab

        // Pre-fill String arrays with empty sentinels
        for (int i = 0; i < entityCount; i++) {
            GameShell.equipMb[i] = ISAAC.emptyString((byte) 38);
        }
        for (int i = 0; i < entityCount; i++) {
            SurfaceSprite.equipAc[i] = ISAAC.emptyString((byte) 38);
        }

        // Fill per-entity numeric arrays
        for (int i = 0; i < entityCount; i++) {
            ClientRuntimeException.equipA[i] = ChatCipher.decode(-30504); // obf: la.a[i]
        }
        for (int i = 0; i < entityCount; i++) {
            AudioMixer.equipB[i] = ChatCipher.decode(-30504);
        }
        for (int i = 0; i < entityCount; i++) {
            SurfaceImageProducer.equipD[i] = ChatCipher.decode(-30504);
        }
        for (int i = 0; i < entityCount; i++) {
            DownloadWorker.equipK[i] = ChatCipher.decode(-30504);
        }
        for (int i = 0; i < entityCount; i++) {
            ISAAC.equipA[i] = ChatCipher.decode(-30504);
        }

        // qb.d[i][j]: 12-slot equipment sprite grid per entity; value 255 → -1 (sentinel)
        for (int i = 0; i < entityCount; i++) {
            for (int j = 0; j < 12; j++) {
                GameFrame.equipD[i][j] = ChatCipher.decode(-30504);
                if (GameFrame.equipD[i][j] == 255) {
                    GameFrame.equipD[i][j] = -1; // 255 is the "no sprite" sentinel
                }
            }
        }

        for (int i = 0; i < entityCount; i++) {
            ClientStream.equipT[i] = NameTable.lookupName((byte) -105);
        }
        for (int i = 0; i < entityCount; i++) {
            itemSpriteIndex[i] = NameTable.lookupName((byte) -105); // obf: m.g[i] = ub.a(-105)
        }
        for (int i = 0; i < entityCount; i++) {
            Surface.equipAb[i] = NameTable.lookupName((byte) -105);
        }

        // Early-return gate: if param < 10, skip tiers 3+
        if (param < 10) {
            return;
        }

        for (int i = 0; i < entityCount; i++) {
            ChatCipher.equipE[i] = NameTable.lookupName((byte) -105);
        }
        for (int i = 0; i < entityCount; i++) {
            SurfaceImageProducer.equipC[i] = EntityDef.readSpriteId(65525);
        }
        for (int i = 0; i < entityCount; i++) {
            Packet.equipH[i] = EntityDef.readSpriteId(65525);
        }
        for (int i = 0; i < entityCount; i++) {
            ArchiveReader.equipH[i] = ChatCipher.decode(-30504);
        }
        for (int i = 0; i < entityCount; i++) {
            StreamFactory.equipA[i] = ChatCipher.decode(-30504);
        }
        for (int i = 0; i < entityCount; i++) {
            LinkedQueue.equipJ[i] = ChatCipher.decode(-30504);
        }
        for (int i = 0; i < entityCount; i++) {
            Timer.equipE[i] = ISAAC.emptyString((byte) 38);
        }

        // --- Tier 3a: NPC tables ---
        int npcCount = EntityDef.getCount(65525); // obf: jb.o = t.a(65525)
        Timer.npcNames       = new String[npcCount]; // obf: p.c
        Utility.npcG         = new String[npcCount]; // obf: mb.g

        for (int i = 0; i < npcCount; i++) {
            Utility.npcG[i] = ISAAC.emptyString((byte) 38);
        }
        for (int i = 0; i < npcCount; i++) {
            Timer.npcNames[i] = ISAAC.emptyString((byte) 38);
        }

        // --- Tier 3b: Prayer/spell tables ---
        int prayerCount = EntityDef.getCount(65525); // obf: na.e = t.a(65525)
        BZip.prayerC         = new int[prayerCount];    // obf: aa.c (aa = BZip; NOT ArrayUtil/ab)
        CacheUpdater.prayerE = new String[prayerCount]; // obf: cb.e
        DataStore.prayerD    = new int[prayerCount];    // obf: nb.d
        FontWidths.prayerM   = new int[prayerCount];    // obf: n.m
        WorldEntity.prayerG  = new int[prayerCount];    // obf: w.g
        LinkedQueue.prayerL  = new int[prayerCount];    // obf: db.l

        for (int i = 0; i < prayerCount; i++) {
            CacheUpdater.prayerE[i] = ISAAC.emptyString((byte) 38);
        }
        for (int i = 0; i < prayerCount; i++) {
            LinkedQueue.prayerL[i] = NameTable.lookupName((byte) -105);
        }
        for (int i = 0; i < prayerCount; i++) {
            FontWidths.prayerM[i] = ChatCipher.decode(-30504);
        }
        for (int i = 0; i < prayerCount; i++) {
            DataStore.prayerD[i] = ChatCipher.decode(-30504);
        }
        for (int i = 0; i < prayerCount; i++) {
            BZip.prayerC[i] = ChatCipher.decode(-30504); // obf: aa.c[i]
        }
        for (int i = 0; i < prayerCount; i++) {
            WorldEntity.prayerG[i] = ChatCipher.decode(-30504);
        }

        // --- Tier 3c: Spell/ability tables ---
        int spellCount = EntityDef.getCount(65525); // obf: ua.Db = t.a(65525)
        Utility.spellA    = new int[spellCount];    // obf: mb.a
        NameTable.spellG  = new int[spellCount];    // obf: ub.g
        ClientRuntimeException.spellF = new String[spellCount]; // obf: la.f (la = ClientRuntimeException)
        Globals.spellA    = new String[spellCount]; // obf: l.a  (l = Globals; NOT ListNode/g)
        RecordLoader.spellF = new int[spellCount];  // obf: f.f
        Timer.spellA      = new String[spellCount]; // obf: p.a
        FontBuilder.spellF = new String[spellCount];// obf: s.f
        SurfaceImageProducer.spellF = new int[spellCount]; // obf: fb.f
        TextEncoder.spellB = new int[spellCount];   // obf: h.b

        for (int i = 0; i < spellCount; i++) {
            Globals.spellA[i] = ISAAC.emptyString((byte) 38); // obf: l.a[i]
        }
        for (int i = 0; i < spellCount; i++) {
            ClientRuntimeException.spellF[i] = ISAAC.emptyString((byte) 38); // obf: la.f[i]
        }
        for (int i = 0; i < spellCount; i++) {
            FontBuilder.spellF[i] = ISAAC.emptyString((byte) 38);
        }
        for (int i = 0; i < spellCount; i++) {
            Timer.spellA[i] = ISAAC.emptyString((byte) 38);
        }
        for (int i = 0; i < spellCount; i++) {
            // obf: fb.f[i] = ca.a((byte)91, o.a((byte)38))
            // GameModel.textureId(byte 91, ISAAC.emptyString(byte 38)) — resolves/appends a
            // texture-name index (returns 0 for the empty/"Invisible" sentinel).
            SurfaceImageProducer.spellF[i] = GameModel.textureId((byte) 91, ISAAC.emptyString((byte) 38));
        }
        for (int i = 0; i < spellCount; i++) {
            RecordLoader.spellF[i] = ChatCipher.decode(-30504);
        }
        for (int i = 0; i < spellCount; i++) {
            NameTable.spellG[i] = ChatCipher.decode(-30504);
        }
        for (int i = 0; i < spellCount; i++) {
            Utility.spellA[i] = ChatCipher.decode(-30504);
        }
        for (int i = 0; i < spellCount; i++) {
            TextEncoder.spellB[i] = ChatCipher.decode(-30504);
        }

        // --- Tier 3d: Equipment-slot tables ---
        int equipSlotCount = EntityDef.getCount(65525); // obf: h.a = t.a(65525)
        StringCodec.equipSlotNames  = new String[equipSlotCount]; // obf: u.b
        StringCodec.equipSlotIds    = new int[equipSlotCount];    // obf: u.a
        ChatCipher.equipSlotV       = new int[equipSlotCount];    // obf: v.a
        RecordLoader.equipSlotE     = new String[equipSlotCount]; // obf: f.e
        Mudclient.equipSlotJk       = new int[equipSlotCount];    // obf: client.Jk
        Scene.equipSlotTb           = new int[equipSlotCount];    // obf: lb.Tb (Scene.diagScratch slot)
        NameTable.equipSlotB        = new String[equipSlotCount]; // obf: ub.b
        GameCharacter.equipSlotR    = new String[equipSlotCount]; // obf: ta.r
        StreamBase.equipSlotD       = new int[equipSlotCount];    // obf: ib.d  (ib = StreamBase; NOT FilterStage/hb)

        for (int i = 0; i < equipSlotCount; i++) {
            GameCharacter.equipSlotR[i] = ISAAC.emptyString((byte) 38);
        }
        for (int i = 0; i < equipSlotCount; i++) {
            NameTable.equipSlotB[i] = ISAAC.emptyString((byte) 38);
        }
        for (int i = 0; i < equipSlotCount; i++) {
            StringCodec.equipSlotNames[i] = ISAAC.emptyString((byte) 38);
        }
        for (int i = 0; i < equipSlotCount; i++) {
            RecordLoader.equipSlotE[i] = ISAAC.emptyString((byte) 38);
        }
        for (int i = 0; i < equipSlotCount; i++) {
            StreamBase.equipSlotD[i] = EntityDef.readSpriteId(65525); // obf: ib.d[i]
        }
        for (int i = 0; i < equipSlotCount; i++) {
            ChatCipher.equipSlotV[i] = NameTable.lookupName((byte) -105);
        }
        for (int i = 0; i < equipSlotCount; i++) {
            Mudclient.equipSlotJk[i] = NameTable.lookupName((byte) -105);
        }
        for (int i = 0; i < equipSlotCount; i++) {
            StringCodec.equipSlotIds[i] = ChatCipher.decode(-30504);
        }
        for (int i = 0; i < equipSlotCount; i++) {
            Scene.equipSlotTb[i] = ChatCipher.decode(-30504);
        }

        // --- Tier 3e: Render-size tables ---
        int renderSizeCount = EntityDef.getCount(65525); // obf: v.h = t.a(65525)
        CacheFile.renderG  = new int[renderSizeCount]; // obf: d.g
        ErrorHandler.renderG = new int[renderSizeCount]; // obf: i.g  (note: class i = ErrorHandler)

        for (int i = 0; i < renderSizeCount; i++) {
            ErrorHandler.renderG[i] = ChatCipher.decode(-30504);
        }
        for (int i = 0; i < renderSizeCount; i++) {
            CacheFile.renderG[i] = ChatCipher.decode(-30504);
        }

        // --- Tier 3f: Texture tables ---
        int textureCount = EntityDef.getCount(65525); // obf: ia.h = t.a(65525)
        ClientStream.texN  = new int[textureCount]; // obf: da.N
        DecodeBuffer.texL  = new int[textureCount]; // obf: ac.l
        Panel.texK         = new int[textureCount]; // obf: qa.K

        for (int i = 0; i < textureCount; i++) {
            Panel.texK[i] = NameTable.lookupName((byte) -105);
        }
        for (int i = 0; i < textureCount; i++) {
            ClientStream.texN[i] = ChatCipher.decode(-30504);
        }
        for (int i = 0; i < textureCount; i++) {
            DecodeBuffer.texL[i] = ChatCipher.decode(-30504);
        }

        // --- Tier 3g: Object/map tables ---
        // n.c = EntityDef.getCount(65525) but the value is unused directly here
        int _objectCountUnused = EntityDef.getCount(65525); // obf: n.c = t.a(65525)
        int mapEntityCount = EntityDef.getCount(65525);     // obf: fa.b = t.a(65525)

        NameHash.mapEntities     = new int[mapEntityCount][];    // obf: oa.d
        GameFrame.mapEntityE     = new int[mapEntityCount];      // obf: qb.e
        ClientStream.mapJ        = new int[mapEntityCount][];    // obf: da.J
        BitBuffer.mapEntityName  = new String[mapEntityCount];   // obf: ja.L
        NameHash.mapEntityNames  = new String[mapEntityCount];   // obf: oa.a
        ImageLoader.mapEntityF   = new int[mapEntityCount];      // obf: pa.f
        ISAAC.mapEntityP         = new int[mapEntityCount];      // obf: o.p

        for (int i = 0; i < mapEntityCount; i++) {
            BitBuffer.mapEntityName[i] = ISAAC.emptyString((byte) 38);
        }
        for (int i = 0; i < mapEntityCount; i++) {
            NameHash.mapEntityNames[i] = ISAAC.emptyString((byte) 38);
        }
        for (int i = 0; i < mapEntityCount; i++) {
            ImageLoader.mapEntityF[i] = ChatCipher.decode(-30504);
        }
        for (int i = 0; i < mapEntityCount; i++) {
            ISAAC.mapEntityP[i] = ChatCipher.decode(-30504);
        }
        for (int i = 0; i < mapEntityCount; i++) {
            GameFrame.mapEntityE[i] = ChatCipher.decode(-30504);
        }
        for (int i = 0; i < mapEntityCount; i++) {
            int subLen = ChatCipher.decode(-30504); // obf: v.a(-30504)
            NameHash.mapEntities[i] = new int[subLen];
            for (int j = 0; j < subLen; j++) {
                // plain assignment — clean base has NO 255→-1 sentinel here (unlike equipD)
                NameHash.mapEntities[i][j] = EntityDef.readSpriteId(65525);
            }
        }
        for (int i = 0; i < mapEntityCount; i++) {
            int subLen = ChatCipher.decode(-30504);
            ClientStream.mapJ[i] = new int[subLen];
            for (int j = 0; j < subLen; j++) {
                ClientStream.mapJ[i][j] = ChatCipher.decode(-30504);
            }
        }

        // --- Tier 3h: Sound tables ---
        int soundCount = EntityDef.getCount(65525); // obf: t.g = t.a(65525)
        EntityDef.soundNames  = new String[soundCount]; // obf: t.h
        GameModel.soundB      = new int[soundCount];    // obf: ca.B (GameModel.sharedScratch slot)
        ClientIOException.soundC = new int[soundCount]; // obf: fa.c
        TextEncoder.soundE    = new String[soundCount]; // obf: h.e

        for (int i = 0; i < soundCount; i++) {
            EntityDef.soundNames[i] = ISAAC.emptyString((byte) 38);
        }
        for (int i = 0; i < soundCount; i++) {
            TextEncoder.soundE[i] = ISAAC.emptyString((byte) 38);
        }
        for (int i = 0; i < soundCount; i++) {
            GameModel.soundB[i] = ChatCipher.decode(-30504);
        }
        for (int i = 0; i < soundCount; i++) {
            ClientIOException.soundC[i] = ChatCipher.decode(-30504);
        }

        // --- Cleanup: release raw archive byte arrays ---
        // Now that all data has been parsed into typed arrays, the raw archive bytes
        // can be GC'd to recover memory.
        Packet.loadedIntData    = null; // obf: b.v = null
        InputState.loadedStringData = null; // obf: kb.d = null
    }

    // -------------------------------------------------------------------------
    // XOR string-pool helpers (obfuscation infrastructure; present in every class)
    // -------------------------------------------------------------------------

    /**
     * First decode stage: converts a {@code String} to a {@code char[]} and XORs the first
     * character with {@code 0x5E} ('{@code ^}') if the string has fewer than 2 characters.
     * If {@code length >= 2} the array is returned unchanged (the caller's second stage handles it).
     *
     * @param s  the raw encoded string literal
     * @return   the partially-decoded char array
     * obf: {@code private static char[] z(String)}
     */
    private static char[] decodeChars(String s) {
        char[] chars = s.toCharArray();
        if (chars.length < 2) {
            chars[0] = (char) (chars[0] ^ 0x5E); // 0x5E = '^'
        }
        return chars;
    }

    /**
     * Second decode stage: XORs each character of the char array with the per-position key
     * from the 5-element table {@code {64, 20, 39, 57, 94}} (position {@code i % 5}), then
     * interns and returns the result as a {@code String}.
     *
     * @param chars  partially-decoded char array from {@link #decodeChars}
     * @return       the fully-decoded, interned {@code String}
     * obf: {@code private static String z(char[])}
     */
    private static String decodeString(char[] chars) {
        // Key table for mod-5 XOR: {64='@', 20=DC4, 39='\'', 57='9', 94='^'}
        int[] key = {64, 20, 39, 57, 94};
        for (int i = 0; i < chars.length; i++) {
            chars[i] = (char) (chars[i] ^ key[i % 5]);
        }
        return new String(chars).intern();
    }
}
