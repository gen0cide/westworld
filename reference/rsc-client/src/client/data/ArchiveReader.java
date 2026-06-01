package client.data;

import java.io.IOException;

/**
 * ArchiveReader — Jagex on-disk cache sector reader/writer for RSC rev ~233-235.
 *
 * <p>The Jagex RSC cache uses a flat-file block storage format:
 * <ul>
 *   <li>An <em>index file</em> stores fixed-width 6-byte records — one per file entry —
 *       holding the uncompressed data length and the index of the first 520-byte sector.</li>
 *   <li>A <em>data file</em> is partitioned into 520-byte sectors.  Each sector starts
 *       with a small header (8 bytes for file-IDs ≤ 65535, 10 bytes for extended IDs)
 *       followed by up to 512 or 510 bytes of payload.  The header contains the expected
 *       file ID, the chunk number within the file, the index of the next sector (or 0 for
 *       the last chunk), and the cache-type byte used as a cross-check.</li>
 * </ul>
 *
 * <p>Both files are accessed through {@link DataStore} ({@code nb}) wrapper objects that
 * layer buffered I/O over a {@link CacheFile} ({@code d}) RandomAccessFile.
 *
 * <p>The class is synchronized on {@code dataStore} so that concurrent threads sharing
 * the same data file do not interleave sector reads/writes.
 *
 * <p>Obfuscated class name: {@code ob}.  Lives in the {@code client.data} sub-package
 * (per the canonical name map).  The {@code CacheUpdater} ({@code cb}) constructs
 * instances of this class and stores downloaded archive bytes in the static
 * {@link #loadedArchives} and {@link #sectorAlloc} scratch fields.
 *
 * <h2>Sector layout (regular, fileId ≤ 65535)</h2>
 * <pre>
 *   offset  size  field
 *   0       2     fileId         (must equal the requested entry ID)
 *   2       2     chunkNumber    (0-based chunk index within this file)
 *   4       3     nextSector     (index of next 520-byte sector, 0 = end)
 *   7       1     cacheType      (must equal this.cacheType)
 *   8     512     payload
 * </pre>
 *
 * <h2>Sector layout (extended, fileId > 65535)</h2>
 * <pre>
 *   offset  size  field
 *   0       4     fileId         (must equal the requested entry ID)
 *   4       2     chunkNumber
 *   6       3     nextSector
 *   9       1     cacheType
 *   10    510     payload
 * </pre>
 *
 * <h2>Index entry layout (always 6 bytes)</h2>
 * <pre>
 *   offset  size  field
 *   0       3     dataLength     (total uncompressed size)
 *   3       3     firstSector    (first 520-byte sector index)
 * </pre>
 */
final class ArchiveReader {

    // -------------------------------------------------------------------------
    // Static scratch arrays used externally by CacheUpdater (cb)
    // -------------------------------------------------------------------------

    /** Loaded archive byte arrays (indexed by archive slot). obf: {@code j} */
    static byte[][] loadedArchives = new byte[1000][];

    /** Sector allocation / size scratch table; populated by CacheUpdater. obf: {@code h} */
    static int[] sectorAlloc;

    // -------------------------------------------------------------------------
    // Dead profiling counters — incremented near each method entry; never read.
    // Kept as fields to match the class layout; never write meaningful code against them.
    // -------------------------------------------------------------------------

    /** Dead profiling counter for write(int,byte[],boolean,int,int). obf: {@code e} */
    static int _profileWrite;          // obf: e

    /** Dead profiling counter for the public store method a(int,int,int,byte[]). obf: {@code d} */
    static int _profileStore;          // obf: d

    /** Dead profiling counter for toString(). obf: {@code g} */
    static int _profileToString;       // obf: g

    /** Dead profiling counter for load(int,int). obf: {@code i} */
    static int _profileLoad;           // obf: i

    // -------------------------------------------------------------------------
    // Instance fields
    // -------------------------------------------------------------------------

    /**
     * The cache-type byte written into every sector header and verified on
     * every read.  Corresponds to the on-disk file index (e.g. 0 for the main
     * cache data file, 1 for cache data 1, etc.).
     * obf: {@code c}
     */
    private int cacheType;             // obf: c

    /**
     * The data {@link DataStore} — the flat file divided into 520-byte sectors.
     * Seeks are made to {@code sectorIndex * SECTOR_SIZE}.
     * obf: {@code f}
     */
    private DataStore dataStore;       // obf: f (nb)

    /**
     * The index {@link DataStore} — the flat file of 6-byte records.
     * Seeks are made to {@code entryId * INDEX_ENTRY_SIZE}.
     * obf: {@code b}
     */
    private DataStore indexStore;      // obf: b (nb)

    /**
     * Maximum number of file entries this archive instance is allowed to hold.
     * Defaults to {@link #DEFAULT_MAX_ENTRIES} (65,000).
     * obf: {@code a}
     */
    private int maxEntries = DEFAULT_MAX_ENTRIES;  // obf: a

    // -------------------------------------------------------------------------
    // Constants
    // -------------------------------------------------------------------------

    /** Size of one on-disk sector in bytes. obf: literal {@code 520} */
    private static final int SECTOR_SIZE = 520;

    /**
     * Maximum data payload per sector for <em>regular</em> (8-byte header) sectors.
     * Regular sectors are used when fileId ≤ 65535.
     */
    private static final int DATA_PER_REGULAR_SECTOR = 512;

    /**
     * Maximum data payload per sector for <em>extended</em> (10-byte header) sectors.
     * Extended sectors are used when fileId > 65535.
     */
    private static final int DATA_PER_EXTENDED_SECTOR = 510;

    /** Size of one index entry in the index file (3 bytes length + 3 bytes firstSector). */
    private static final int INDEX_ENTRY_SIZE = 6;

    /** Default maximum number of entries the archive supports. obf: literal {@code 65000} */
    private static final int DEFAULT_MAX_ENTRIES = 65000;

    /**
     * Magic constant that must be passed as the first argument to {@link #load(int, int)}
     * to authorise a read.  Acts as a lightweight caller-authentication token.
     * Value: {@code 9395} (0x24B3).
     * obf: literal {@code 9395} in bytecode
     */
    private static final int READ_MAGIC = 9395;

    // -------------------------------------------------------------------------
    // XOR-encoded string pool (decoded at class-load time).
    // These are error-message fragments used by the obfuscated exception wrapper.
    // The originals are encoded with a two-pass XOR scheme:
    //   pass 1: if len < 2 → chars[0] ^= 127; else no-op
    //   pass 2: chars[i] ^= {126, 99, 115, 116, 127}[i % 5]
    // Decoded values are listed below.  obf field name: {@code z}
    // -------------------------------------------------------------------------
    private static final String[] ERROR_STRINGS = new String[]{
        "ob.A(",       // z[0] — prefix for load() error message
        "ob.toString()", // z[1] — toString() method signature
        "ob.B(",       // z[2] — prefix for store() error message
        "{...}",       // z[3] — non-null byte[] placeholder
        "null",        // z[4] — null byte[] placeholder
        "ob.C(",       // z[5] — prefix for private write() error message
        "ob.<init>(",  // z[6] — constructor error message prefix
    };

    // -------------------------------------------------------------------------
    // Constructor
    // -------------------------------------------------------------------------

    /**
     * Creates an ArchiveReader bound to an existing pair of DataStore files.
     *
     * @param cacheType  The cache-type index byte written into every sector
     *                   header and validated on reads (e.g. 0, 1, …).
     * @param dataStore  The 520-byte-sector data DataStore.    obf: {@code nb var2}
     * @param indexStore The 6-byte-record index DataStore.    obf: {@code nb var3}
     * @param maxEntries The maximum number of entries (use 0 for the default
     *                   of {@value #DEFAULT_MAX_ENTRIES}).    obf: {@code int var4}
     */
    ArchiveReader(int cacheType, DataStore dataStore, DataStore indexStore, int maxEntries) {
        // obf: ob(int var1, nb var2, nb var3, int var4)
        this.indexStore = null; // obf initialiser sets b=null before try block
        this.maxEntries   = maxEntries;
        this.cacheType    = cacheType;
        this.indexStore   = indexStore;   // obf: this.b = var3
        this.dataStore    = dataStore;    // obf: this.f = var2
    }

    // -------------------------------------------------------------------------
    // Public API
    // -------------------------------------------------------------------------

    /**
     * Stores (writes) {@code data[0..dataLen-1]} as a new or updated cache entry
     * identified by {@code entryId}.
     *
     * <p>The method first tries to overwrite an existing entry (reusing its first
     * sector).  If that fails (e.g. the entry does not exist yet), it allocates a
     * fresh sector chain.
     *
     * <p>The third parameter ({@code _junk}) is an obfuscated anti-tamper dummy that
     * is never used in the real logic; it has been dropped.
     *
     * @param entryId  The cache entry index to write.   obf: {@code param1}
     * @param dataLen  The number of bytes to write.    obf: {@code param2}
     * @param _junk    Anti-tamper dummy (ignored).     obf: {@code param3}  — DROPPED
     * @param data     The payload bytes.               obf: {@code param4}
     * @return {@code true} if the write succeeded; {@code false} otherwise.
     * obf: {@code final boolean a(int, int, int, byte[])}
     */
    final boolean store(int entryId, int dataLen, int _junk, byte[] data) {
        // obf: a(int param1, int param2, int param3, byte[] param4)
        // Anti-tamper/opaque-predicate stripped:
        //   int n5 = 94 % ((param3 - -61) / 35);   // junk — dropped

        // Increment dead profiling counter (++d).  Kept as a field write only.
        ++_profileStore;

        synchronized (dataStore) {
            // Validate entryId is in [0, maxEntries).
            if (dataLen < 0 || dataLen > this.maxEntries) {
                throw new IllegalArgumentException();
            }
            // First try to update an existing entry (reuse its sector chain).
            boolean ok = writeEntry(entryId, data, /*tryExisting=*/true, dataLen, /*cacheTypeHint=*/4);
            if (!ok) {
                // Entry didn't exist or its sector was unreachable; allocate fresh.
                ok = writeEntry(entryId, data, /*tryExisting=*/false, dataLen, /*cacheTypeHint=*/4);
            }
            return ok;
        }
    }

    /**
     * Reads and returns the full byte payload for cache entry {@code entryId}.
     *
     * <p>The caller <em>must</em> pass {@link #READ_MAGIC} (9395 / 0x24B3) as
     * {@code readMagic}; any other value causes an immediate {@code null} return.
     * This is a lightweight caller-authentication guard baked in by the original
     * obfuscation.
     *
     * @param readMagic Must equal {@link #READ_MAGIC} (9395).  obf: {@code param1}
     * @param entryId   The cache entry index to read.           obf: {@code param2}
     * @return The reconstructed payload bytes, or {@code null} if the entry does
     *         not exist, the index is corrupt, or {@code readMagic} is wrong.
     * obf: {@code final byte[] a(int, int)}
     */
    final byte[] load(int readMagic, int entryId) throws IOException {
        // obf: a(int param1, int param2)
        // Profiling counter stripped: ++ob.i

        synchronized (dataStore) {
            // ----------------------------------------------------------------
            // Step 1: Bounds-check the index file.
            // The index file must be large enough to contain a 6-byte record
            // for entryId.  indexStore.getSize() returns the file length.
            // ----------------------------------------------------------------
            long indexFileSize = indexStore.getSize();  // obf: this.b.a((byte)-111)
            // Check: indexFileSize < 6 + entryId*6  →  not enough data, bail out.
            if (indexFileSize < (long)(INDEX_ENTRY_SIZE + entryId * INDEX_ENTRY_SIZE)) {
                return null;
            }

            // ----------------------------------------------------------------
            // Step 2: Read the 6-byte index entry for this entryId.
            //   [0..2] = dataLength   (3 bytes, big-endian)
            //   [3..5] = firstSector  (3 bytes, big-endian)
            // ----------------------------------------------------------------
            // Seek to byte offset entryId*6 in the index file, then read 6 bytes
            // into the shared scratch buffer la.c.
            indexStore.seek(6L * entryId, /*opaque=*/12);            // obf: this.b.a(long, int)
            indexStore.readBytes(/*isRead=*/true, 6, 0, ScratchBuffer.DATA);  // obf: this.b.a(boolean, int, int, byte[])

            // Decode 3-byte big-endian dataLength from la.c[0..2].
            int dataLength = ((ScratchBuffer.DATA[0] & 0xFF) << 16)
                           | ((ScratchBuffer.DATA[1] & 0xFF) <<  8)
                           |  (ScratchBuffer.DATA[2] & 0xFF);

            // Decode 3-byte big-endian firstSector from la.c[3..5].
            int firstSector = ((ScratchBuffer.DATA[3] & 0xFF) << 16)
                            | ((ScratchBuffer.DATA[4] & 0xFF) <<  8)
                            |  (ScratchBuffer.DATA[5] & 0xFF);

            // Validate: dataLength must be non-negative and ≤ maxEntries.
            if (dataLength < 0 || dataLength > this.maxEntries) {
                return null;
            }

            // Validate: firstSector must be a legal sector index
            // (i.e. firstSector * 520 ≤ dataFile.length()).
            long dataFileSize = dataStore.getSize();               // obf: this.f.a((byte)-111)
            if (firstSector < 0 || (long)firstSector > dataFileSize / SECTOR_SIZE) {
                return null;
            }

            // ----------------------------------------------------------------
            // Step 3: Caller-authentication guard.
            //   Only proceed if readMagic == READ_MAGIC (9395).
            // ----------------------------------------------------------------
            if (readMagic != READ_MAGIC) {
                return null;
            }

            // ----------------------------------------------------------------
            // Step 4: Allocate output buffer and walk the sector chain.
            // ----------------------------------------------------------------
            byte[] output    = new byte[dataLength]; // obf: var6
            int    bytesRead = 0;                     // obf: var7 — bytes written to output
            int    chunkNum  = 0;                     // obf: var8 — expected chunk number

            while (bytesRead < dataLength) {
                // firstSector == 0 means no more sectors (broken chain).
                if (firstSector == 0) {
                    return null;
                }

                // Seek to the current sector in the data file.
                // obf: this.f.a(var5 * 520, 107)  — 107 is an opaque anti-tamper tag
                dataStore.seek((long)firstSector * SECTOR_SIZE, /*opaque=*/107);

                // Remaining bytes to read in this chunk.
                int remaining = dataLength - bytesRead;

                // Choose header size and max payload based on whether entryId is "extended".
                int headerSize; // number of header bytes before the payload
                int maxPayload; // maximum payload bytes in this sector
                int nextSector;
                int sectorEntryId;
                int sectorChunkNum;
                int sectorCacheType;

                if (entryId > 0xFFFF) {
                    // ---- Extended sector (10-byte header) ----
                    // Cap remaining at max extended payload (510 bytes).
                    if (remaining > DATA_PER_EXTENDED_SECTOR) {
                        remaining = DATA_PER_EXTENDED_SECTOR;
                    }
                    headerSize = 10;

                    // Read headerSize + remaining bytes into scratch buffer.
                    // obf: this.f.a(true, remaining + headerSize, 0, la.c)
                    dataStore.readBytes(true, remaining + headerSize, 0, ScratchBuffer.DATA);

                    // Decode extended header (10 bytes):
                    //   [0..3] fileId (4 bytes, big-endian)
                    //   [4..5] chunkNumber (2 bytes, big-endian)
                    //   [6..8] nextSector (3 bytes, big-endian)
                    //   [9]    cacheType (1 byte)
                    sectorEntryId =
                        ((ScratchBuffer.DATA[0] & 0xFF) << 24)
                      | ((ScratchBuffer.DATA[1] & 0xFF) << 16)
                      | ((ScratchBuffer.DATA[2] & 0xFF) <<  8)
                      |  (ScratchBuffer.DATA[3] & 0xFF);
                    sectorChunkNum =
                        ((ScratchBuffer.DATA[4] & 0xFF) <<  8)
                      |  (ScratchBuffer.DATA[5] & 0xFF);
                    nextSector =
                        ((ScratchBuffer.DATA[6] & 0xFF) << 16)
                      | ((ScratchBuffer.DATA[7] & 0xFF) <<  8)
                      |  (ScratchBuffer.DATA[8] & 0xFF);
                    sectorCacheType = ScratchBuffer.DATA[9] & 0xFF;
                } else {
                    // ---- Regular sector (8-byte header) ----
                    // Cap remaining at max regular payload (512 bytes).
                    if (remaining > DATA_PER_REGULAR_SECTOR) {
                        remaining = DATA_PER_REGULAR_SECTOR;
                    }
                    headerSize = 8;

                    // Read headerSize + remaining bytes into scratch buffer.
                    dataStore.readBytes(true, headerSize + remaining, 0, ScratchBuffer.DATA);

                    // Decode regular header (8 bytes):
                    //   [0..1] fileId (2 bytes, big-endian)
                    //   [2..3] chunkNumber (2 bytes, big-endian)
                    //   [4..6] nextSector (3 bytes, big-endian)
                    //   [7]    cacheType (1 byte)
                    sectorEntryId =
                        ((ScratchBuffer.DATA[0] & 0xFF) <<  8)
                      |  (ScratchBuffer.DATA[1] & 0xFF);
                    sectorChunkNum =
                        ((ScratchBuffer.DATA[2] & 0xFF) <<  8)
                      |  (ScratchBuffer.DATA[3] & 0xFF);
                    nextSector =
                        ((ScratchBuffer.DATA[4] & 0xFF) << 16)
                      | ((ScratchBuffer.DATA[5] & 0xFF) <<  8)
                      |  (ScratchBuffer.DATA[6] & 0xFF);
                    sectorCacheType = ScratchBuffer.DATA[7] & 0xFF;
                }

                // Validate the sector header fields.
                if (sectorEntryId != entryId
                        || sectorChunkNum != chunkNum
                        || sectorCacheType != this.cacheType) {
                    return null;
                }

                // Validate nextSector is a legal sector index (or 0 for end).
                if (nextSector < 0 || (long)nextSector > dataStore.getSize() / SECTOR_SIZE) {
                    return null;
                }

                // Copy the payload bytes (immediately after header in ScratchBuffer.DATA)
                // into the output buffer.
                System.arraycopy(ScratchBuffer.DATA, headerSize, output, bytesRead, remaining);
                bytesRead += remaining;

                // Advance to the next sector.
                firstSector = nextSector; // obf: var5 = var12_15
                chunkNum++;               // obf: var8++
            }

            return output;
        }
    }

    // -------------------------------------------------------------------------
    // Private helper — core read/write sector I/O
    // -------------------------------------------------------------------------

    /**
     * Low-level sector writer.  Writes the bytes in {@code data[0..dataLen-1]}
     * to the cache for entry {@code entryId}, either by updating an existing
     * sector chain ({@code tryExisting=true}) or by allocating a completely
     * new chain ({@code tryExisting=false}).
     *
     * <p>The fourth literal parameter {@code cacheTypeHint} (always passed as 4
     * by the caller) is used as an index into the scratch buffer when building
     * the 6-byte index entry and as a minor XOR tag on opaque seek calls; it is
     * not independently meaningful.
     *
     * @param entryId       The file entry to write.              obf: {@code n2}
     * @param data          The full data payload.                obf: {@code byArray}
     * @param tryExisting   {@code true} → try to reuse the existing first sector.
     *                                                            obf: {@code bl}
     * @param dataLen       Length of data to write.              obf: {@code n3}
     * @param cacheTypeHint Always 4 from the caller; used as buffer index.
     *                                                            obf: {@code n4}
     * @return {@code true} on success; {@code false} if the operation could not
     *         complete (e.g. the existing entry is absent or the chain is corrupt).
     * obf: {@code private final boolean a(int, byte[], boolean, int, int)}
     */
    private final boolean writeEntry(int entryId, byte[] data,
                                     boolean tryExisting, int dataLen,
                                     int cacheTypeHint) throws IOException {
        // obf: a(int n2, byte[] byArray, boolean bl, int n3, int n4)
        // Dead profiling: ++ob.e  — dropped.

        synchronized (dataStore) {
            // obf local: n5 = current/first sector index we will write into
            int currentSector;

            if (tryExisting) {
                // ------------------------------------------------------------------
                // Try to reuse the existing first sector for this entryId.
                // The index file must be large enough: 6 + 6*entryId bytes.
                // ------------------------------------------------------------------
                long indexFileSize = indexStore.getSize();  // obf: this.b.a((byte)-111)
                if (indexFileSize < (long)(INDEX_ENTRY_SIZE + entryId * INDEX_ENTRY_SIZE)) {
                    return false; // Index entry doesn't exist yet.
                }

                // Seek and read the 6-byte index entry.
                // obf: this.b.a(n2 * 6, -124)
                indexStore.seek((long)entryId * INDEX_ENTRY_SIZE, /*opaque=*/-124);
                // obf: this.b.a(true, 6, 0, la.c)
                indexStore.readBytes(true, INDEX_ENTRY_SIZE, 0, ScratchBuffer.DATA);

                // Decode: la.c[3..5] = existing firstSector (3 bytes, big-endian).
                // obf: n5 = (0xFF & la.c[5]) + (la.c[3] << 16 & 0xFF0000) - -((la.c[4] & 0xFF) << 8)
                currentSector = ((ScratchBuffer.DATA[3] & 0xFF) << 16)
                              | ((ScratchBuffer.DATA[4] & 0xFF) <<  8)
                              |  (ScratchBuffer.DATA[5] & 0xFF);

                // A firstSector of 0 means no existing entry.
                if (currentSector <= 0) {
                    return false;
                }
                // Validate: firstSector must be a legal sector index.
                if ((long)currentSector > dataStore.getSize() / SECTOR_SIZE) {
                    return false;
                }
                // currentSector is valid; fall through to write loop.

            } else {
                // ------------------------------------------------------------------
                // Allocate a brand-new first sector at the end of the data file.
                // Formula: ceil(fileLength / SECTOR_SIZE), minimum 1.
                // obf: n5 = (int)((this.f.a((byte)-111) - -519L) / 520L)
                //       if n5 == 0: n5 = 1
                // The -(-519) = +519 ensures ceiling division: (len+519)/520.
                // ------------------------------------------------------------------
                long dataFileSize = dataStore.getSize();
                currentSector = (int)((dataFileSize + (SECTOR_SIZE - 1)) / SECTOR_SIZE);
                if (currentSector == 0) {
                    currentSector = 1; // Sector 0 is reserved / invalid.
                }
            }

            // ------------------------------------------------------------------
            // Build and write the 6-byte index entry:
            //   [0..2] = dataLen   (3 bytes, big-endian)
            //   [3..5] = firstSector (3 bytes, big-endian)
            // Note: the index is written BEFORE the sector data so that a partial
            // write failure leaves the index pointing to a (possibly incomplete)
            // sector chain rather than dangling.
            // ------------------------------------------------------------------
            // obf: la.c[0] = (byte)(n3 >> 16); la.c[1] = (byte)(n3 >> 8); la.c[2] = (byte)n3
            ScratchBuffer.DATA[0] = (byte)(dataLen >> 16);
            ScratchBuffer.DATA[1] = (byte)(dataLen >>  8);
            ScratchBuffer.DATA[2] = (byte) dataLen;
            // obf: la.c[3] = (byte)(n5 >> 16); la.c[cacheTypeHint=4] = (byte)(n5 >> 8); la.c[5] = (byte)n5
            ScratchBuffer.DATA[3] = (byte)(currentSector >> 16);
            ScratchBuffer.DATA[cacheTypeHint] = (byte)(currentSector >> 8); // cacheTypeHint == 4
            ScratchBuffer.DATA[5] = (byte) currentSector;

            // Seek to this entry in the index file and write 6 bytes.
            // obf: this.b.a(6 * n2, 31)  ; this.b.a(6, -102, 0, la.c)
            indexStore.seek((long)entryId * INDEX_ENTRY_SIZE, /*opaque=*/31);
            indexStore.writeBytes(INDEX_ENTRY_SIZE, /*opaque=*/-102, 0, ScratchBuffer.DATA);

            // ------------------------------------------------------------------
            // Write loop: walk (or build) the sector chain, writing up to
            // DATA_PER_REGULAR_SECTOR or DATA_PER_EXTENDED_SECTOR bytes per
            // sector, until all dataLen bytes have been stored.
            // ------------------------------------------------------------------
            int bytesWritten = 0;  // obf: n6
            int chunkNum     = 0;  // obf: n7

            while (bytesWritten < dataLen) {
                // Seek to the current sector.
                // obf: this.f.a(520 * n5, n4 ^ 0x11)   (XOR with 4^17=21, opaque)
                dataStore.seek((long)currentSector * SECTOR_SIZE, cacheTypeHint ^ 0x11);

                boolean extended = (entryId > 0xFFFF);
                int headerSize;
                int maxPayload;

                if (extended) {
                    headerSize = 10;
                    maxPayload = DATA_PER_EXTENDED_SECTOR; // 510
                } else {
                    headerSize = 8;
                    maxPayload = DATA_PER_REGULAR_SECTOR;  // 512
                }

                // How many bytes remain to be written?
                int remaining = dataLen - bytesWritten;

                // Determine nextSector.
                // obf: n8 = 0 if nextSector comes from an empty/existing read;
                //          re-read from file or allocate fresh.
                int nextSector;
                if (tryExisting) {
                    // Read back the existing sector header to get the stored nextSector.
                    if (extended) {
                        // obf: this.f.a(true, 10, 0, la.c)
                        dataStore.readBytes(true, headerSize, 0, ScratchBuffer.DATA);
                        // nextSector is at [6..8]
                        nextSector = ((ScratchBuffer.DATA[6] & 0xFF) << 16)
                                   | ((ScratchBuffer.DATA[7] & 0xFF) <<  8)
                                   |  (ScratchBuffer.DATA[8] & 0xFF);
                    } else {
                        // obf: this.f.a(true, 8, 0, la.c)
                        dataStore.readBytes(true, headerSize, 0, ScratchBuffer.DATA);
                        // nextSector is at [4..6]
                        nextSector = ((ScratchBuffer.DATA[4] & 0xFF) << 16)
                                   | ((ScratchBuffer.DATA[5] & 0xFF) <<  8)
                                   |  (ScratchBuffer.DATA[6] & 0xFF);
                    }
                    // If nextSector == 0 (no more existing chain), switch to
                    // allocating new sectors.
                    if (nextSector == 0) {
                        tryExisting = false;
                        // Allocate next free sector.
                        // obf: n8 = (int)((this.f.a((byte)-111) - -519) / 520)
                        long sz = dataStore.getSize();
                        nextSector = (int)((sz + (SECTOR_SIZE - 1)) / SECTOR_SIZE);
                        if (nextSector == 0) nextSector = 1;
                        // Don't reuse the current sector.
                        if (nextSector == currentSector) nextSector++;
                    }
                } else {
                    // Allocating a fresh chain: compute the next free sector.
                    long sz = dataStore.getSize();
                    nextSector = (int)((sz + (SECTOR_SIZE - 1)) / SECTOR_SIZE);
                    if (nextSector == 0) nextSector = 1;
                    if (nextSector == currentSector) nextSector++;
                }

                // If this is the last chunk (remaining ≤ maxPayload), mark chain end.
                // obf: if (512 >= n3 - n6): n8 = 0
                if (remaining <= maxPayload) {
                    nextSector = 0; // End of chain.
                }

                // ---- Build and write the sector header ----
                if (extended) {
                    // 10-byte extended header:
                    //   [0..3] entryId (4 bytes, big-endian)
                    //   [4..5] chunkNum (2 bytes, big-endian)
                    //   [6..8] nextSector (3 bytes, big-endian)
                    //   [9]    cacheType
                    ScratchBuffer.DATA[0] = (byte)(entryId   >> 24);
                    ScratchBuffer.DATA[1] = (byte)(entryId   >> 16);
                    ScratchBuffer.DATA[2] = (byte)(entryId   >>  8);
                    ScratchBuffer.DATA[3] = (byte) entryId;
                    ScratchBuffer.DATA[4] = (byte)(chunkNum  >>  8);
                    ScratchBuffer.DATA[5] = (byte) chunkNum;
                    ScratchBuffer.DATA[6] = (byte)(nextSector >> 16);
                    ScratchBuffer.DATA[7] = (byte)(nextSector >>  8);
                    ScratchBuffer.DATA[8] = (byte) nextSector;
                    ScratchBuffer.DATA[9] = (byte) this.cacheType;

                    // Seek back to sector start and write the 10-byte header.
                    // obf: this.f.a(520 * n5, n4 ^ 0x21) ; this.f.a(10, -111, 0, la.c)
                    dataStore.seek((long)currentSector * SECTOR_SIZE, cacheTypeHint ^ 0x21);
                    dataStore.writeBytes(headerSize, /*opaque=*/-111, 0, ScratchBuffer.DATA);
                } else {
                    // 8-byte regular header:
                    //   [0..1] entryId (2 bytes, big-endian)
                    //   [2..3] chunkNum (2 bytes, big-endian)
                    //   [4..6] nextSector (3 bytes, big-endian)
                    //   [7]    cacheType
                    ScratchBuffer.DATA[0] = (byte)(entryId   >>  8);
                    ScratchBuffer.DATA[1] = (byte) entryId;
                    ScratchBuffer.DATA[2] = (byte)(chunkNum  >>  8);
                    ScratchBuffer.DATA[3] = (byte) chunkNum;
                    ScratchBuffer.DATA[4] = (byte)(nextSector >> 16);
                    ScratchBuffer.DATA[5] = (byte)(nextSector >>  8);
                    ScratchBuffer.DATA[6] = (byte) nextSector;
                    ScratchBuffer.DATA[7] = (byte) this.cacheType;

                    // Seek back to sector start and write the 8-byte header.
                    // obf: this.f.a(n5 * 520, n4 ^ 0x7F) ; this.f.a(8, -107, 0, la.c)
                    dataStore.seek((long)currentSector * SECTOR_SIZE, cacheTypeHint ^ 0x7F);
                    dataStore.writeBytes(headerSize, /*opaque=*/-107, 0, ScratchBuffer.DATA);
                }

                // ---- Write payload bytes ----
                // Clamp to maxPayload.
                int toWrite = Math.min(remaining, maxPayload);
                // obf: this.f.a(toWrite, n4 + (-119 or -125), bytesWritten, data)
                //   n4 + -119 = 4 - 119 = -115 (extended path)
                //   n4 + -125 = 4 - 125 = -121 (regular path)
                int opaque = cacheTypeHint + (extended ? -119 : -125);
                dataStore.writeBytes(toWrite, opaque, bytesWritten, data);

                bytesWritten   += toWrite;
                currentSector   = nextSector;  // obf: n5 = n8
                chunkNum++;                    // obf: ++n7
            }

            return true;
        }
    }

    // -------------------------------------------------------------------------
    // Object overrides
    // -------------------------------------------------------------------------

    /**
     * Returns a string representation consisting solely of the {@link #cacheType}
     * index byte.  This is used by the obfuscated error-handler when printing
     * context for exceptions.
     * obf: {@code public final String toString()}
     */
    @Override
    public final String toString() {
        // Dead profiling: ++g — dropped.
        return "" + this.cacheType;
    }
}
