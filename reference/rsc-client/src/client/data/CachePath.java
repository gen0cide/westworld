package client.data;

import java.io.File;
import java.io.RandomAccessFile;
import java.util.Hashtable;

/**
 * CachePath — resolves and creates the on-disk cache directory/file for the RSC client.
 *
 * <p>On startup, {@link #initialize} is called once (with a sentinel byte == 101, a numeric
 * cache-slot suffix such as 32, and an optional sub-directory name).  It records the
 * {@code user.home} system property so the resolver can prefer the user's home directory.
 *
 * <p>When the client needs a cache {@link File} for a given filename it calls
 * {@link #resolveFile}.  The resolver tries every combination of:
 * <pre>
 *   baseDirs   × { ".jagex_cache_&lt;slot&gt;", ".file_store_&lt;slot&gt;" }
 * </pre>
 * in two passes:
 * <ol>
 *   <li><b>Pass 0</b> – only accept a directory whose base already exists on disk.</li>
 *   <li><b>Pass 1</b> – try any writable directory, creating it if necessary (but only
 *       if the base dir string is empty, i.e. the current working directory).</li>
 * </ol>
 * The first candidate that can be opened for read/write access wins; the result is cached
 * in a {@link Hashtable} so the same filename never re-scans.
 *
 * <p>Corresponds to obfuscated class {@code r} (rev ~233–235, Microsoft J++ build).
 *
 * <p>Oracle match: {@code mudclient.findcachedir()} in
 * {@code rscdump.com-runescape-classic-dump/LeadingBot/mudclient.java} (lines 9451-9471)
 * shows the same "try a list of base directories, mkdir, test write access" pattern.
 *
 * obf: r
 */
public final class CachePath {

    // -----------------------------------------------------------------------
    // XOR-encrypted string pool (decoded from static initializer)
    //
    // Original encoding: z(z("...")) where:
    //   z(String)  → char[]: if len < 2, XOR sole char with 0x46; else toCharArray()
    //   z(char[])  → String: XOR char[i] with key[i%5], key = {70,'F', 34,'"', 103,'g', 48,'0', 70,'F'}
    //
    // Decoded values (indices match bytecode aaload offsets):
    //   z[0]  = "~/"            — fallback home prefix when user.home unavailable   obf: z[0]
    //   z[1]  = "user.home"     — System property key                               obf: z[1]
    //   z[2]  = "/rscache/"     — Unix/Mac base dir                                 obf: z[2]
    //   z[3]  = ".jagex_cache_" — primary subdirectory name prefix                  obf: z[3]
    //   z[4]  = "c:/"           — bare Windows drive root                           obf: z[4]
    //   z[5]  = "c:/winnt/"     — Windows NT root                                   obf: z[5]
    //   z[6]  = "c:/rscache/"   — Windows rscache base                              obf: z[6]
    //   z[7]  = ".file_store_"  — secondary subdirectory name prefix                obf: z[7]
    //   z[8]  = "/tmp/"         — Unix temp base                                    obf: z[8]
    //   z[9]  = "rw"            — RandomAccessFile open mode                        obf: z[9]
    //   z[10] = "c:/windows/"   — Windows system dir root                           obf: z[10]
    // -----------------------------------------------------------------------

    /**
     * Ordered list of filesystem base directories to probe when searching for a writable
     * cache location.  Tried from index 0 to 7; the first one that is writable wins.
     *
     * <p>Index 5 is filled at runtime with {@link #homeDir} (user.home + "/" or "~/").
     *
     * obf: local var 5 inside a(int,String,String,int)
     */
    private static final String[] BASE_DIRS_TEMPLATE = {
        "c:/rscache/",  // [0] — Windows explicit rscache location       obf: z[6]
        "/rscache/",    // [1] — Unix/Mac /rscache/                      obf: z[2]
        "c:/windows/",  // [2] — Windows system dir (last resort Windows) obf: z[10]
        "c:/winnt/",    // [3] — Windows NT dir                          obf: z[5]
        "c:/",          // [4] — bare C drive root                       obf: z[4]
        null,           // [5] — placeholder; replaced with homeDir at runtime  obf: r.a
        "/tmp/",        // [6] — Unix /tmp fallback                      obf: z[8]
        "",             // [7] — current working directory (last resort) obf: literal ""
    };

    /**
     * Filename-to-File cache: once a file path is successfully resolved and verified
     * writable, its File is stored here so subsequent lookups skip the directory scan.
     * Key = filename string (param2 passed to the resolver).
     *
     * obf: r.b
     */
    private static final Hashtable<String, File> fileCache = new Hashtable<String, File>(16); // obf: b

    /**
     * The numeric cache slot suffix appended to the subdirectory name (e.g. 32 → ".jagex_cache_32").
     * Set by {@link #initialize}.
     *
     * obf: r.c
     */
    private static int cacheSlot; // obf: c

    /**
     * Flag set to {@code true} after {@link #initialize} completes successfully.
     * The core resolver throws {@link RuntimeException} if called before init.
     *
     * obf: r.d
     */
    private static boolean initialized = false; // obf: d

    /**
     * Optional client subdirectory name placed between the cache root and the filename
     * (e.g. "runescape" → ".jagex_cache_32/runescape/savegame.dat").  May be {@code null}.
     * Set by {@link #initialize}.
     *
     * obf: r.e
     */
    private static String clientSubDir; // obf: e

    /**
     * The user's home directory with a trailing "/", resolved from the {@code user.home}
     * system property.  Falls back to "~/" if the property is unavailable.
     * Set by {@link #initialize}.
     *
     * obf: r.a
     */
    private static String homeDir; // obf: a

    // -----------------------------------------------------------------------
    // Static utility-only class — no instances allowed.
    // The private constructor throws Error to enforce this even via reflection.
    // -----------------------------------------------------------------------

    /** obf: r() */
    private CachePath() throws Throwable {
        throw new Error();
    }

    // -----------------------------------------------------------------------
    // Public API
    // -----------------------------------------------------------------------

    /**
     * Initialise the cache path resolver.  Must be called exactly once before any call to
     * {@link #resolveFile}.
     *
     * <p>Anti-tamper: if {@code sentinel} is not exactly {@code 101} the method calls the
     * internal resolver with junk arguments ({@code cacheSlot=-64, pass=-78}), which always
     * fails and throws {@link RuntimeException} — crashing the client.
     *
     * @param slot      numeric suffix appended to the cache subdirectory name (e.g. 32).
     *                  obf: var0
     * @param sentinel  must equal {@code 101} (0x65); any other value triggers the
     *                  anti-tamper crash path.  obf: var1
     * @param subDir    optional sub-directory inside the cache dir (e.g. "runescape"), or
     *                  {@code null} for a flat layout.  obf: var2
     *
     * obf: r.a(int,byte,String)
     */
    public static void initialize(int slot, byte sentinel, String subDir) {
        // Store slot and subDir for later use by resolveFile().
        cacheSlot = slot;       // obf: c = var0
        clientSubDir = subDir;  // obf: e = var2

        // Anti-tamper sentinel check.  Byte value 101 == 'e'.
        // If the caller passes any other value the resolver below is invoked with nonsense
        // parameters; it exhausts all candidates without finding anything writable and throws.
        if (sentinel != 101) {
            // Deliberately crash: pass=-78 starts well below 0, so the outer while loop runs
            // many iterations; cacheSlot=-64 produces a directory name no legitimate filesystem
            // path can match; eventually pass reaches 2 and RuntimeException is thrown.
            resolveOrCreateCacheFile(-64, null, null, -78); // anti-tamper crash
        }

        // Resolve the user home directory.  If the JVM property is unavailable (e.g.
        // security manager blocks it, or applet sandbox) fall back to the literal "~/" string
        // which works on Unix shells but not all JVMs — it's a graceful-degradation fallback.
        try {
            homeDir = System.getProperty("user.home"); // obf: z[1] = "user.home"
            if (homeDir != null) {
                homeDir = homeDir + "/";
            }
        } catch (Exception ignored) {
            // Security manager blocked the property read — homeDir stays null, caught below.
        }

        if (homeDir == null) {
            homeDir = "~/"; // obf: z[0] = "~/" — Unix tilde fallback
        }

        initialized = true; // obf: d = true
    }

    /**
     * Returns a writable {@link File} for {@code filename} within the cache directory,
     * or {@code null} if {@code cacheType} is not {@code 2}.
     *
     * <p>This is the thin public wrapper used by the rest of the client (e.g. {@code CacheFile}).
     * It enforces that only cache-type 2 is supported in this revision.
     *
     * @param cacheType  must equal {@code 2}; any other value returns {@code null}.
     *                   obf: var0
     * @param filename   the filename to resolve inside the cache dir.  obf: var1
     * @return a writable {@link File}, or {@code null} if cacheType != 2.
     *
     * obf: r.a(int,String)
     */
    public static File resolveFile(int cacheType, String filename) {
        // Only cache type 2 is served; other types return null immediately.
        if (cacheType != 2) {
            return null;
        }
        return resolveOrCreateCacheFile(cacheSlot, clientSubDir, filename, 0);
    }

    // -----------------------------------------------------------------------
    // Internal resolver
    // -----------------------------------------------------------------------

    /**
     * Core cache-directory resolver.  Tries every combination of {@link #BASE_DIRS_TEMPLATE}
     * and subdirectory prefixes (".jagex_cache_&lt;slot&gt;" / ".file_store_&lt;slot&gt;") in
     * two passes, returning the first {@link File} that can be opened with {@code "rw"} access.
     *
     * <h3>Pass logic</h3>
     * <ul>
     *   <li><b>pass == 0</b>: the full candidate path must already exist on disk
     *       ({@code File.exists()}).  This is the fast path for clients that have already
     *       run before.</li>
     *   <li><b>pass == 1</b>: the candidate path may be newly created.  For non-empty
     *       base dirs the base directory itself must exist; for the empty-string base (CWD)
     *       no pre-existence check is done, so it can always be created.</li>
     *   <li><b>pass &ge; 2</b>: no writable location was found; throws
     *       {@link RuntimeException} to signal a fatal cache-init failure.</li>
     * </ul>
     *
     * <h3>Write-access probe</h3>
     * Once a candidate directory is created, the method opens the candidate file with a
     * {@link java.io.RandomAccessFile} in {@code "rw"} mode, reads one byte, seeks back to
     * position 0, writes the same byte back, seeks to 0 again, and closes the file.  This
     * confirms the file is both readable and writable before caching the result.  If any
     * {@link Exception} is thrown during this probe the candidate is skipped.
     *
     * <h3>Result caching</h3>
     * A successful {@link File} is stored in {@link #fileCache} keyed by {@code filename}
     * so that future calls with the same filename skip the scan entirely.
     *
     * @param slot       numeric suffix for the subdirectory name (e.g. 32).   obf: param0 / n2
     * @param subDir     optional intermediate subdirectory (e.g. "runescape"). obf: param1 / string
     * @param filename   the target filename to resolve.                        obf: param2 / string2
     * @param startPass  which pass to begin at (0 = normal; -78 = anti-tamper crash path).
     *                                                                          obf: param3 / n3
     * @return the resolved, writable {@link File}.
     * @throws RuntimeException if no writable cache location can be found after both passes.
     *
     * obf: r.a(int,String,String,int)
     */
    public static File resolveOrCreateCacheFile(int slot, String subDir, String filename, int startPass) {
        // Guard: must call initialize() first.
        if (!initialized) {
            throw new RuntimeException("");
        }

        // Fast path: return previously resolved File for this filename.
        File cached = (File) fileCache.get(filename); // obf: local 4
        if (cached != null) {
            return cached;
        }

        // Build the array of base directories to probe.
        // Index 5 is the runtime home directory (filled from static field homeDir / r.a).
        String[] baseDirs = new String[]{ // obf: local 5 (stringArray)
            "c:/rscache/",  // [0]  obf: z[6]
            "/rscache/",    // [1]  obf: z[2]
            "c:/windows/",  // [2]  obf: z[10]
            "c:/winnt/",    // [3]  obf: z[5]
            "c:/",          // [4]  obf: z[4]
            homeDir,        // [5]  obf: r.a — user.home+"/" or "~/"
            "/tmp/",        // [6]  obf: z[8]
            "",             // [7]  obf: literal "" — CWD / last resort
        };

        // Build the two subdirectory name candidates for this slot:
        //   ".jagex_cache_32"  — standard Jagex cache dir name
        //   ".file_store_32"   — alternative "file store" dir name (older convention)
        String[] subdirNames = new String[]{ // obf: local 6 (stringArray2)
            ".jagex_cache_" + slot, // [0]  obf: z[3] + n2
            ".file_store_"  + slot, // [1]  obf: z[7] + n2
        };

        // Two-pass outer loop.  pass=0: must exist; pass=1: try to create; pass>=2: give up.
        for (int pass = startPass; pass < 2; pass++) { // obf: i2; bytecode: iinc 7 1 / goto 0a8
            // Inner loop over subdirectory name candidates (jagex_cache_ vs file_store_).
            for (int subdirIdx = 0; subdirIdx < subdirNames.length; subdirIdx++) { // obf: i3
                // Innermost loop over base directories.
                for (int baseIdx = 0; baseIdx < baseDirs.length; baseIdx++) { // obf: n4
                    String baseDir  = baseDirs[baseIdx];   // obf: local 13 (string5)
                    String subdirName = subdirNames[subdirIdx]; // obf: part of stringArray2[i3]

                    // Build the full candidate file path:
                    //   <baseDir><subdirName>/<subDir?>/<filename>
                    // e.g. "c:/rscache/.jagex_cache_32/runescape/savegame.dat"
                    String candidatePath = baseDir + subdirName + "/"
                        + (subDir != null ? subDir + "/" : "")
                        + filename; // obf: local 10 (string4)

                    RandomAccessFile probe = null; // obf: local 11 (randomAccessFile)
                    try {
                        File candidateFile = new File(candidatePath); // obf: local 12 (file2)

                        // Pass 0: only accept already-existing files.
                        if (pass == 0 && !candidateFile.exists()) {
                            continue;
                        }

                        // Pass 1: for non-empty base dirs (real paths like "c:/windows/"),
                        // skip if the base directory itself does not exist.
                        // The bitwise trick "~pass == -2" is equivalent to "pass == 1".
                        //   ~1 == -2  →  true  →  apply existence check
                        //   ~0 == -1  →  false →  skip (pass 0 already handled above)
                        if ((~pass == -2) && baseDir.length() > 0 && !new File(baseDir).exists()) {
                            continue;
                        }

                        // Create the cache subdirectory (e.g. ".jagex_cache_32/").
                        // mkdir() is a no-op if it already exists, and we ignore failure here —
                        // the RandomAccessFile open below will catch any real I/O problem.
                        new File(baseDir + subdirName).mkdir();

                        // Create the optional intermediate sub-directory (e.g. "runescape/").
                        if (subDir != null) {
                            new File(baseDir + subdirName + "/" + subDir).mkdir();
                        }

                        // Write-access probe: open, read 1 byte, seek 0, write same byte,
                        // seek 0 again, close.  Confirms the file is both readable and writable.
                        // Mode "rw" (obf: z[9]) creates the file if it does not yet exist.
                        probe = new RandomAccessFile(candidateFile, "rw"); // obf: z[9] = "rw"
                        int probeByte = probe.read();  // read first byte (or -1 if empty)
                        probe.seek(0L);
                        probe.write(probeByte);        // write same byte back → confirms write access
                        probe.seek(0L);
                        probe.close();
                        probe = null;

                        // Cache the result and return.
                        fileCache.put(filename, candidateFile);
                        return candidateFile;

                    } catch (Exception ex) {
                        // This candidate failed (I/O error, permission denied, etc.).
                        // Close the probe file handle if it was opened, then try the next candidate.
                        try {
                            if (probe != null) {
                                probe.close();
                                probe = null;
                            }
                        } catch (Exception ignored) {
                            // Ignore close failure — just move on.
                        }
                        // continue to next baseDir
                    }
                } // end baseIdx loop
            } // end subdirIdx loop
        } // end pass loop

        // No writable cache location was found after both passes.
        // The client cannot continue without a cache directory.
        throw new RuntimeException("CachePath: no writable cache directory found for slot " + slot);
    }
}
