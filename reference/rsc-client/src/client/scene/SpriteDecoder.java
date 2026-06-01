package client.scene;

/**
 * SpriteDecoder — BZip2 block decompressor for sprite/image data.
 *
 * <p>This class is a self-contained, static-only BZip2 decompressor. It is used
 * to inflate compressed sprite/image streams (stored in the RSC JAG archive) into
 * raw byte arrays. Its Burrows-Wheeler scratch array lives on {@link Surface} (ua)
 * as the static field {@code Surface.tt} ({@code ua.Mb}) — a 100 000-int decode
 * buffer that doubles as storage on the Surface class but is NOT the pixel
 * framebuffer.
 *
 * <p>The algorithm is a faithful port of Julian Seward's bzip2 block-decompressor:
 * <ol>
 *   <li>Parse the BZip2 block header (magic bytes, origPtr, inUse bitmap,
 *       Huffman selector MTF, per-group code-length tables).</li>
 *   <li>Decode the MTF+RLE2-encoded symbol stream using the per-group Huffman
 *       tables into a Burrows-Wheeler block ({@code Surface.tt} / the {@code tt[]}
 *       array, obf {@code ua.Mb}).</li>
 *   <li>Build the cumulative-frequency array ({@code cftab}) and apply the
 *       inverse Burrows-Wheeler transform, then run-length decode (RLE1) the
 *       output bytes.</li>
 * </ol>
 *
 * <p>The sole entry point is the package-visible static method
 * {@link #decompress(byte[], int, byte[], int, int)}, which is called by the
 * archive/image loader to decompress a single BZip2 stream.
 *
 * <p>A single shared {@link DecodeBuffer} instance ({@code sharedState}) is kept
 * as a static field and protected with a {@code synchronized} block so that
 * re-entrant calls from different threads are serialised.
 *
 * <p>Obfuscated class name: {@code ea} (package-private {@code final class}).
 *
 * @see client.util.DecodeBuffer  (ac)  — the BZip2 decompressor state record
 * @see client.scene.Surface      (ua)  — hosts the BWT scratch array {@code tt} (obf {@code Mb})
 */
final class SpriteDecoder {

    // -----------------------------------------------------------------------
    // Static state
    // -----------------------------------------------------------------------

    /**
     * Shared, reusable decompressor state.  Protected by the synchronized block
     * in {@link #decompress}.
     * obf: a
     */
    private static DecodeBuffer sharedState = new DecodeBuffer();

    // -----------------------------------------------------------------------
    // Public / package-visible entry point
    // -----------------------------------------------------------------------

    /**
     * Decompresses a BZip2-encoded block from {@code compressedData} (starting at
     * {@code compressedOffset}) into {@code outputBuffer}.
     *
     * <p>This matches oracle {@code BZLib.decompress(byte[], int, byte[], int, int)}.
     *
     * @param outputBuffer     destination byte array for decompressed data
     * @param availableOutput  number of bytes available in {@code outputBuffer}
     * @param compressedData   source BZip2-compressed bytes
     * @param compressedOffset byte offset into {@code compressedData} to start reading
     * @param compressedSize   total length of the compressed data available
     * @return number of bytes actually written to {@code outputBuffer}
     * obf: a(byte[], int, byte[], int, int) : int
     */
    static final int decompress(byte[] outputBuffer, int availableOutput,
                                byte[] compressedData, int compressedOffset,
                                int compressedSize) {
        synchronized (sharedState) {
            // Wire up the shared state for this decompression run
            sharedState.input        = compressedData;
            sharedState.nextIn       = compressedSize;       // obf: a.a — source byte index / limit
            sharedState.output       = outputBuffer;
            sharedState.availOut     = 0;                    // obf: a.o — write cursor into output
            sharedState.decompSize   = availableOutput;      // obf: a.y — bytes remaining in output
            sharedState.bsLive       = 0;                    // obf: a.p — bits in bit-buffer
            sharedState.bsBuff       = 0;                    // obf: a.e — bit-buffer accumulator
            sharedState.totalInLo32  = 0;                    // obf: a.c — input byte counter (lo)
            sharedState.totalOutLo32 = 0;                    // obf: a.G — output byte counter (lo)

            decompressBlocks(sharedState);

            // Return bytes consumed from availableOutput
            int bytesWritten = availableOutput - sharedState.decompSize;
            sharedState.input  = null;
            sharedState.output = null;
            return bytesWritten;
        }
    }

    // -----------------------------------------------------------------------
    // Core decompression loop  (oracle: BZLib.decompress / BZState loop)
    // -----------------------------------------------------------------------

    /**
     * Processes one or more BZip2 blocks from the input stream.  Each iteration
     * of the outer {@code while} loop handles one complete BZip2 block (header
     * parse → Huffman decode → BWT inverse → RLE1 output).  Returns when the
     * end-of-stream magic byte (0x17 / 23) is encountered.
     *
     * <p>Oracle: {@code BZLib.decompress(BZState)} (the private inner loop).
     * obf: b(ac)
     */
    private static final void decompressBlocks(DecodeBuffer s) {
        // Huffman decode tables for the current group — refreshed every 50 symbols
        int     minLen;                 // obf: var22 / n18 — min code length for current group
        int[]   limitTable  = null;     // obf: var23 / nArray  — limit[gSel]
        int[]   baseTable   = null;     // obf: var24 / nArray2 — base[gSel]
        int[]   permTable   = null;     // obf: var25 / nArray3 — perm[gSel]

        // Symbols-until-next-group-switch counter (reloaded every 50 syms = BZGSIZE)
        int groupPos;                   // obf: var12 / n10
        int groupNo  = -1;             // obf: var11 / n9  — current group index

        // Working Huffman decoder accumulators
        int zn;                         // obf: var18 / n16 — current code length tried
        int zvec;                       // obf: var19 / n17 — current code accumulator
        byte zj;                        // obf: var20 / by  — extra bit read during width search

        // Per-block working vars
        int nblock;                     // obf: var14 / n12 — symbols decoded so far (block size)
        byte gSel;                      // obf: var21 / by2 — selector for current group

        // -----------------------------------------------------------------------
        // blocksize100k: always 1 in this build (single 100 000-entry tt[] array)
        // -----------------------------------------------------------------------
        s.blocksize100k = 1;           // obf: var0.f = 1

        // Allocate the BWT transform array on first call; 100 000 ints per block
        if (Surface.tt == null) {       // obf: ua.Mb == null  (Surface hosts the bzip2 tt[] scratch array)
            Surface.tt = new int[s.blocksize100k * 100000];
        }

        boolean moreBlocks = true;
        while (moreBlocks) {

            // -------------------------------------------------------------------
            // 1. Read and validate BZip2 block header (6 magic bytes + 1 randomised bit)
            //    Magic: 0x31 0x41 0x59 0x26 0x53 0x59  (= "\x31AY&SY" = π digits in ASCII)
            //    EOS:   0x17 0x72 0x45 0x38 0x50 0x90  (first byte == 23 → end of stream)
            // -------------------------------------------------------------------
            byte headerByte = readByte(s);   // obf: var1 = c(var0)
            if (headerByte == 23) {          // EOS magic first byte
                return;
            }
            // Consume remaining 9 header bytes (magic + randomised flag):
            // bytes [1..5] of block magic, then the "block randomised" flag byte
            readByte(s);  // magic[1]
            readByte(s);  // magic[2]
            readByte(s);  // magic[3]
            readByte(s);  // magic[4]
            readByte(s);  // magic[5]
            readByte(s);  // (unused in this rev — would be block-no for multi-stream)
            readByte(s);
            readByte(s);
            readByte(s);
            // blockRandomised flag — if non-zero, oracle prints "PANIC! RANDOMISED BLOCK!"
            // This build silently ignores it (the empty `if` is the dead code).
            /* byte randomised = */ readBit(s);  // obf: d(var0); if (var1 != 0) {}

            // -------------------------------------------------------------------
            // 2. origPtr — 24-bit index of the original string in the BWT matrix
            // -------------------------------------------------------------------
            s.origPtr = 0;
            s.origPtr = s.origPtr << 8 | (readByte(s) & 0xFF);
            s.origPtr = s.origPtr << 8 | (readByte(s) & 0xFF);
            s.origPtr = s.origPtr << 8 | (readByte(s) & 0xFF);

            // -------------------------------------------------------------------
            // 3. inUse16[16] — which 16-symbol super-groups are present
            // -------------------------------------------------------------------
            for (int i = 0; i < 16; i++) {
                byte bit = readBit(s);
                s.inUse16[i] = (bit == 1);
            }

            // -------------------------------------------------------------------
            // 4. inUse[256] — per-symbol presence bitmap (derived from inUse16)
            // -------------------------------------------------------------------
            for (int i = 0; i < 256; i++) {
                s.inUse[i] = false;
            }
            for (int i = 0; i < 16; i++) {
                if (s.inUse16[i]) {
                    for (int j = 0; j < 16; j++) {
                        byte bit = readBit(s);
                        if (bit == 1) {
                            s.inUse[i * 16 + j] = true;
                        }
                    }
                }
            }

            // -------------------------------------------------------------------
            // 5. makeMaps: compress inUse[] → setToUnseq[] (nInUse symbols)
            // -------------------------------------------------------------------
            makeMaps(s);                         // obf: a(var0)
            int alphaSize = s.nInUse + 2;        // obf: var7 = var0.K + 2

            // -------------------------------------------------------------------
            // 6. Selector MTF stream
            //    nGroups  (3 bits)  — number of Huffman tables (up to 6)
            //    nSelectors (15 bits) — number of 50-symbol group assignments
            // -------------------------------------------------------------------
            int nGroups    = readBits(3,  s);    // obf: var8 = a(3, var0)
            int nSelectors = readBits(15, s);    // obf: var9 = a(15, var0)

            // Read selector MTF values: unary coded (count 1-bits before first 0)
            for (int i = 0; i < nSelectors; i++) {
                int v = 0;
                while (readBit(s) != 0) {       // obf: while(d(var0) != 0) v++;
                    v++;
                }
                s.selectorMtf[i] = (byte) v;
            }

            // Decode selector MTF → selector[] (move-to-front)
            byte[] pos = new byte[6];            // obf: var27
            for (int i = 0; i < nGroups; i++) {
                pos[i] = (byte) i;               // obf: while(var29 < var8) var27[var29] = var29++
            }
            for (int i = 0; i < nSelectors; i++) {
                int v = s.selectorMtf[i] & 0xFF;
                byte tmp = pos[v];
                for (; v > 0; v--) {
                    pos[v] = pos[v - 1];
                }
                pos[0] = tmp;
                s.selector[i] = tmp;
            }

            // -------------------------------------------------------------------
            // 7. Huffman code-length tables (for each of the nGroups tables)
            //    Each length delta-coded: start = getBits(5); then for each alpha
            //    symbol: while 1-bit { if next-bit==0 curr++ else curr-- }
            // -------------------------------------------------------------------
            for (int t = 0; t < nGroups; t++) {
                int curr = readBits(5, s);       // obf: var17 = a(5, var0)
                for (int i = 0; i < alphaSize; i++) {
                    while (readBit(s) != 0) {    // obf: while(d(var0) != 0)
                        byte dir = readBit(s);
                        if (dir == 0) {
                            curr++;
                        } else {
                            curr--;
                        }
                    }
                    s.len[t][i] = (byte) curr;
                }
            }

            // -------------------------------------------------------------------
            // 8. Build Huffman decode tables (limit/base/perm) for each group
            // -------------------------------------------------------------------
            for (int t = 0; t < nGroups; t++) {
                // Find min and max code lengths across all alpha symbols
                byte minCode = 32;
                byte maxCode = 0;
                for (int i = 0; i < alphaSize; i++) {
                    if (s.len[t][i] > maxCode) maxCode = s.len[t][i];
                    if (s.len[t][i] < minCode) minCode = s.len[t][i];
                }
                // Build the actual decode tables
                buildDecodeTables(s.limit[t], s.base[t], s.perm[t],
                                  s.len[t], minCode, maxCode, alphaSize);
                s.minLens[t] = minCode;
            }

            // -------------------------------------------------------------------
            // 9. Initialise block decode state
            // -------------------------------------------------------------------
            int eob     = s.nInUse + 1;  // end-of-block sentinel symbol  obf: var10 = var0.K + 1
            groupNo     = -1;
            groupPos    = 0;             // obf: var12 = 0

            // Clear unzftab (symbol frequency counts used for BWT inverse)
            for (int i = 0; i <= 255; i++) {
                s.unzftab[i] = 0;        // obf: var0.m[i] = 0
            }

            // Initialise the MTF array (mtfa) for Move-To-Front decoding.
            // mtfa is divided into 16 groups of 16 entries; each group holds the
            // current order of byte values [0..255].  Initialise in reverse so
            // that mtfa[0] = 0x0F, ..., mtfa[4095] = 0xFF  (big-endian order).
            int kk = 4095;              // obf: var29
            for (int ii = 15; ii >= 0; ii--) {
                for (int jj = 15; jj >= 0; jj--) {
                    s.mtfa[kk] = (byte) (ii * 16 + jj);
                    kk--;
                }
                s.mtfbase[ii] = kk + 1; // start-of-group pointer
            }

            nblock = 0;                 // obf: var14

            // -------------------------------------------------------------------
            // 10. Prime the Huffman decoder for the first group (GETMTFVAL)
            // -------------------------------------------------------------------
            if (groupPos == 0) {
                groupNo++;
                groupPos = 50;          // BZGSIZE — group switch every 50 symbols
                gSel      = s.selector[groupNo];
                minLen    = s.minLens[gSel];
                limitTable = s.limit[gSel];
                permTable  = s.perm[gSel];
                baseTable  = s.base[gSel];
            }
            groupPos--;

            // Read first Huffman codeword
            zn   = minLen;
            zvec = readBits(zn, s);
            while (zvec > limitTable[zn]) {
                zn++;
                zj = readBit(s);
                zvec = zvec << 1 | zj;
            }
            int nextSym = permTable[zvec - baseTable[zn]]; // obf: var13

            // -------------------------------------------------------------------
            // 11. Main symbol decode loop
            // -------------------------------------------------------------------
            while (nextSym != eob) {

                if (nextSym == 0 || nextSym == 1) {
                    // ---- BZRUNA (0) / BZRUNB (1): run-length coded zero-runs ----
                    // These encode repeated copies of the front-of-MTF symbol
                    // using a bijective base-2 representation: BZRUNA adds N,
                    // BZRUNB adds 2*N, where N starts at 1 and doubles each time.
                    int es = -1;       // obf: var15  — run length - 1 (delta)
                    int N  = 1;        // obf: var16  — positional weight (powers of 2)
                    do {
                        if (nextSym == 0) {
                            es += 1 * N;   // BZRUNA
                        } else {           // nextSym == 1
                            es += 2 * N;   // BZRUNB
                        }
                        N *= 2;

                        // GETMTFVAL — advance group if needed
                        if (groupPos == 0) {
                            groupNo++;
                            groupPos  = 50;
                            gSel      = s.selector[groupNo];
                            minLen    = s.minLens[gSel];
                            limitTable = s.limit[gSel];
                            permTable  = s.perm[gSel];
                            baseTable  = s.base[gSel];
                        }
                        groupPos--;
                        zn   = minLen;
                        zvec = readBits(zn, s);
                        while (zvec > limitTable[zn]) {
                            zn++;
                            zj = readBit(s);
                            zvec = zvec << 1 | zj;
                        }
                        nextSym = permTable[zvec - baseTable[zn]];
                    } while (nextSym == 0 || nextSym == 1);

                    es++;  // run length now correct (bijective decode adds 1)

                    // The symbol being repeated is the front of the MTF array
                    // mapped through setToUnseq[] to the actual byte value
                    byte uc = s.setToUnseq[s.mtfa[s.mtfbase[0]] & 0xFF];  // obf: var0.d[var0.A[var0.r[0]]]
                    s.unzftab[uc & 0xFF] += es;  // obf: var0.m[uc & 255] += var15+1

                    // Emit es copies of uc into the BWT block array
                    for (; es > 0; es--) {
                        Surface.tt[nblock] = uc & 0xFF;  // obf: ua.Mb[var14] = uc & 0xFF
                        nblock++;
                    }

                } else {
                    // ---- Literal symbol: MTF index nn = nextSym - 1 ----
                    int nn = nextSym - 1;  // obf: var33 / n25

                    if (nn < 16) {
                        // Fast path: symbol is in the first group (first 16 slots)
                        int pp  = s.mtfbase[0];    // obf: var103 / n23
                        byte uc = s.mtfa[pp + nn]; // obf: var1 = var0.A[var103 + var33]

                        // Shift [pp..pp+nn-1] right by one (unrolled by 4 for speed)
                        while (nn > 3) {
                            int z = pp + nn;
                            s.mtfa[z]     = s.mtfa[z - 1];
                            s.mtfa[z - 1] = s.mtfa[z - 2];
                            s.mtfa[z - 2] = s.mtfa[z - 3];
                            s.mtfa[z - 3] = s.mtfa[z - 4];
                            nn -= 4;
                        }
                        while (nn > 0) {
                            s.mtfa[pp + nn] = s.mtfa[pp + nn - 1];
                            nn--;
                        }
                        s.mtfa[pp] = uc;

                        // Map through setToUnseq and accumulate
                        int byteVal = s.setToUnseq[uc & 0xFF] & 0xFF;  // obf: var0.d[var1 & 255] & 255
                        s.unzftab[byteVal]++;
                        Surface.tt[nblock] = byteVal;
                        nblock++;

                    } else {
                        // Slow path: symbol spans across multiple groups in mtfa
                        int lno = nn / 16;   // obf: var31 / n27 — group index
                        int off = nn % 16;   // obf: var32 / n28 — offset within group

                        // Extract the symbol and shift within its group
                        int pp  = s.mtfbase[lno] + off;  // obf: var30
                        byte uc = s.mtfa[pp];             // obf: var1
                        while (pp > s.mtfbase[lno]) {
                            s.mtfa[pp] = s.mtfa[pp - 1];
                            pp--;
                        }
                        s.mtfbase[lno]++;

                        // Cascade the last element of each group into the front
                        // of the group below (maintain "virtual ring" invariant)
                        while (lno > 0) {
                            s.mtfbase[lno]--;
                            s.mtfa[s.mtfbase[lno]] = s.mtfa[s.mtfbase[lno - 1] + 16 - 1];
                            lno--;
                        }
                        s.mtfbase[0]--;
                        s.mtfa[s.mtfbase[0]] = uc;

                        // If the base pointer for group 0 has wrapped to 0,
                        // compact all groups back to the end of the mtfa array.
                        if (s.mtfbase[0] == 0) {
                            kk = 4095;
                            for (int ii = 15; ii >= 0; ii--) {
                                for (int jj = 15; jj >= 0; jj--) {
                                    s.mtfa[kk] = s.mtfa[s.mtfbase[ii] + jj];
                                    kk--;
                                }
                                s.mtfbase[ii] = kk + 1;
                            }
                        }

                        // Map and accumulate
                        int byteVal = s.setToUnseq[uc & 0xFF] & 0xFF;
                        s.unzftab[byteVal]++;
                        Surface.tt[nblock] = byteVal;
                        nblock++;
                    }

                    // GETMTFVAL — advance group switch counter and read next sym
                    if (groupPos == 0) {
                        groupNo++;
                        groupPos  = 50;
                        gSel      = s.selector[groupNo];
                        minLen    = s.minLens[gSel];
                        limitTable = s.limit[gSel];
                        permTable  = s.perm[gSel];
                        baseTable  = s.base[gSel];
                    }
                    groupPos--;
                    zn   = minLen;
                    zvec = readBits(zn, s);
                    while (zvec > limitTable[zn]) {
                        zn++;
                        zj = readBit(s);
                        zvec = zvec << 1 | zj;
                    }
                    nextSym = permTable[zvec - baseTable[zn]];
                }
            } // end while (nextSym != eob)

            // -------------------------------------------------------------------
            // 12. Build inverse BWT permutation (cftab)
            //     cftab[sym] = cumulative count of bytes < sym in the block,
            //     used to reconstruct the original string from the BWT output.
            // -------------------------------------------------------------------
            s.stateOutLen = 0;              // obf: var0.F = 0
            s.stateOutCh  = 0;             // obf: var0.g = 0
            s.cftab[0]    = 0;             // obf: var0.w[0] = 0
            for (int i = 1; i <= 256; i++) {
                s.cftab[i] = s.unzftab[i - 1];   // obf: var0.w[i] = var0.m[i-1]
            }
            // Compute prefix sums (cumulative frequencies)
            for (int i = 1; i <= 256; i++) {
                s.cftab[i] += s.cftab[i - 1];
            }

            // -------------------------------------------------------------------
            // 13. Annotate tt[] with forward-chain links (pack index into high 24 bits)
            //     After this loop tt[cftab[sym]] = sym | (originalPosition << 8)
            //     This is the standard bzip2 Inverse BWT setup.
            // -------------------------------------------------------------------
            for (int i = 0; i < nblock; i++) {
                byte sym = (byte) (Surface.tt[i] & 0xFF);  // obf: var1 = ua.Mb[var66]
                Surface.tt[s.cftab[sym & 0xFF]] |= i << 8;
                s.cftab[sym & 0xFF]++;
            }

            // -------------------------------------------------------------------
            // 14. Prime the BWT inverse traversal
            //     Start at the origPtr-th element; load first k0 byte.
            // -------------------------------------------------------------------
            s.tpos       = Surface.tt[s.origPtr] >> 8;  // obf: var0.H = ua.Mb[var0.E] >> 8
            s.nblockUsed = 0;
            s.tpos       = Surface.tt[s.tpos];          // first step of inverse BWT
            s.k0         = (byte) (s.tpos & 0xFF);      // obf: var0.h = (byte)(var0.H & 0xFF)
            s.tpos     >>= 8;
            s.nblockUsed++;
            s.saveNblock = nblock;

            // -------------------------------------------------------------------
            // 15. Write decompressed bytes to output (inverse BWT + RLE1 decode)
            // -------------------------------------------------------------------
            emitOutput(s);  // obf: e(var0)

            // Continue loop if the entire block was consumed with no pending output
            moreBlocks = (s.nblockUsed == s.saveNblock + 1 && s.stateOutLen == 0);
        }
    }

    // -----------------------------------------------------------------------
    // Output emission  (oracle: BZLib.nextHeader)
    // -----------------------------------------------------------------------

    /**
     * Applies the inverse BWT + RLE1 decode to produce output bytes.
     *
     * <p>This is the performance-critical innermost loop.  It follows the
     * {@code tt[]} chain from {@code tpos} to generate successive bytes, then
     * run-length decodes runs of identical bytes (RLE1 encoding used by bzip2
     * before the BWT stage).
     *
     * <p>The method is resumable: if the output buffer fills mid-block,
     * all state is saved back to {@code s} and the caller can call again.
     *
     * <p>Oracle: {@code BZLib.nextHeader(BZState)}.
     * obf: e(ac)
     */
    private static final void emitOutput(DecodeBuffer s) {
        // Cache all hot fields in locals for JIT/JVM performance (identical to
        // what the oracle and original obfuscated code do).
        byte    stateOutCh  = s.stateOutCh;   // obf: var2 / var2_1  — current run byte
        int     stateOutLen = s.stateOutLen;  // obf: var3 / var3_2  — remaining run length
        int     nblockUsed  = s.nblockUsed;  // obf: var4 / var4_3  — BWT symbols consumed
        int     k0          = s.k0;           // obf: var5 / var5_4  — previous BWT byte
        int[]   tt          = Surface.tt;     // obf: var6 / var6_5  — BWT chain array (ua.Mb)
        int     tpos        = s.tpos;         // obf: var7 / var7_6  — current BWT chain ptr
        byte[]  output      = s.output;       // obf: var8 / var8_7  — destination buffer
        int     availOut    = s.availOut;     // obf: var9 / var9_8  — write cursor
        int     decompSize  = s.decompSize;   // obf: var10/ var10_9 — remaining output space
        int     origDecomp  = decompSize;     // obf: var11/ var11_10 — snapshot for delta
        int     saveNblockPP = s.saveNblock + 1; // obf: var12/ var12_11

        // Main output loop label (matches oracle's "returnNotr" / Vineflower's label63)
        outer:
        while (true) {

            // ----- Drain any pending run (RLE1) -----
            if (stateOutLen > 0) {
                while (true) {
                    if (decompSize == 0) break outer;       // output full
                    if (stateOutLen == 1) {
                        if (decompSize == 0) {
                            stateOutLen = 1;
                            break outer;
                        }
                        output[availOut] = stateOutCh;
                        availOut++;
                        decompSize--;
                        break;  // run exhausted — fall through to BWT advance
                    }
                    output[availOut] = stateOutCh;
                    stateOutLen--;
                    availOut++;
                    decompSize--;
                }
            }

            // ----- Advance BWT chain and detect RLE1 runs -----
            while (nblockUsed != saveNblockPP) {
                stateOutCh = (byte) k0;          // emit previous byte later
                tpos       = tt[tpos];
                byte k1    = (byte) tpos;        // next BWT byte
                tpos     >>= 8;
                nblockUsed++;

                if (k1 != k0) {
                    // Different byte — emit stateOutCh and continue
                    k0 = k1;
                    if (decompSize == 0) {
                        stateOutLen = 1;
                        break outer;
                    }
                    output[availOut] = stateOutCh;
                    availOut++;
                    decompSize--;
                } else {
                    // Same byte — potential RLE1 run; look ahead
                    if (nblockUsed == saveNblockPP) {
                        // Hit end of block with run length 1 so far
                        if (decompSize == 0) {
                            stateOutLen = 1;
                            break outer;
                        }
                        output[availOut] = stateOutCh;
                        availOut++;
                        decompSize--;
                        // stateOutLen stays 0 — loop back to outer
                        continue;
                    }

                    // Run length >= 2: look ahead to determine actual run length
                    stateOutLen = 2;
                    tpos  = tt[tpos];
                    k1    = (byte) tpos;
                    tpos >>= 8;
                    if (++nblockUsed == saveNblockPP) continue outer;  // done
                    if (k1 != k0) {
                        k0 = k1;
                        continue outer;
                    }

                    stateOutLen = 3;
                    tpos  = tt[tpos];
                    k1    = (byte) tpos;
                    tpos >>= 8;
                    if (++nblockUsed == saveNblockPP) continue outer;
                    if (k1 != k0) {
                        k0 = k1;
                        continue outer;
                    }

                    // Four identical bytes in a row — read the RLE1 repeat count
                    // from the *next* BWT byte (this is the bzip2 RLE1 protocol:
                    // after 4 identical bytes the following byte gives 0..255 extra).
                    tpos  = tt[tpos];
                    k1    = (byte) tpos;
                    tpos >>= 8;
                    nblockUsed++;
                    stateOutLen = (k1 & 0xFF) + 4;  // run = 4 + extra_byte
                    tpos   = tt[tpos];
                    k0     = (byte) tpos;            // next non-run byte
                    tpos >>= 8;
                    nblockUsed++;
                    continue outer;
                }
            }

            // End of block reached with no pending run
            stateOutLen = 0;
            break;
        }

        // ---- Write-back all cached locals to the state object ----
        int prevTotalOut = s.totalOutLo32;
        s.totalOutLo32 += origDecomp - decompSize;  // account for bytes just written
        // (overflow check elided — the original has an empty if block here)

        s.stateOutCh  = stateOutCh;
        s.stateOutLen = stateOutLen;
        s.nblockUsed  = nblockUsed;
        s.k0          = k0;
        Surface.tt    = tt;        // write back in case array was replaced
        s.tpos        = tpos;
        s.output      = output;
        s.availOut    = availOut;
        s.decompSize  = decompSize;
    }

    // -----------------------------------------------------------------------
    // Huffman table construction
    // -----------------------------------------------------------------------

    /**
     * Builds the three Huffman decode tables ({@code limit}, {@code base},
     * {@code perm}) from an array of code lengths.
     *
     * <ul>
     *   <li>{@code perm[]} — symbols ordered by code length (for Huffman lookup)</li>
     *   <li>{@code base[]} — starting code value at each code length</li>
     *   <li>{@code limit[]} — maximum code value at each code length</li>
     * </ul>
     *
     * <p>Oracle: {@code BZLib.createDecodeTables(int[], int[], int[], byte[], int, int, int)}.
     * obf: a(int[], int[], int[], byte[], int, int, int)
     *
     * @param limit     output: upper code-value boundary for each bit-length
     * @param base      output: base code value (offset) for each bit-length
     * @param perm      output: symbol permutation sorted by ascending code length
     * @param lengths   input: per-symbol Huffman code lengths
     * @param minLen    minimum code length present
     * @param maxLen    maximum code length present
     * @param alphaSize number of symbols in the alphabet
     */
    private static final void buildDecodeTables(int[] limit, int[] base, int[] perm,
                                                byte[] lengths,
                                                int minLen, int maxLen, int alphaSize) {
        // ---- Build perm[]: symbols in ascending code-length order ----
        int pp = 0;
        for (int i = minLen; i <= maxLen; i++) {
            for (int j = 0; j < alphaSize; j++) {
                if (lengths[j] == i) {
                    perm[pp] = j;
                    pp++;
                }
            }
        }

        // ---- Build base[]: cumulative count of symbols with length < i ----
        for (int i = 0; i < 23; i++) {
            base[i] = 0;
        }
        for (int i = 0; i < alphaSize; i++) {
            base[lengths[i] + 1]++;
        }
        for (int i = 1; i < 23; i++) {
            base[i] += base[i - 1];
        }

        // ---- Build limit[]: maximum Huffman code value at each bit-length ----
        // limit[i] = last code value of length i.  Computed by iterating from
        // minLen upward: start with the count of codes at this length, add to a
        // running code value, then left-shift for the next length.
        for (int i = 0; i < 23; i++) {
            limit[i] = 0;
        }
        int vec = 0;
        for (int i = minLen; i <= maxLen; i++) {
            vec += base[i + 1] - base[i];  // number of codes of length i
            limit[i] = vec - 1;            // highest code of length i
            vec <<= 1;                     // shift to next length
        }

        // ---- Adjust base[] for decoder use ----
        // After this adjustment base[i] = first code of length i, offset such
        // that (code - base[i]) indexes directly into perm[].
        for (int i = minLen + 1; i <= maxLen; i++) {
            base[i] = ((limit[i - 1] + 1) << 1) - base[i];
        }
    }

    // -----------------------------------------------------------------------
    // makeMaps — build setToUnseq[] from inUse[]
    // -----------------------------------------------------------------------

    /**
     * Compresses the sparse {@code inUse[256]} bitmap into the dense
     * {@code setToUnseq[nInUse]} byte array, and records the count in
     * {@code nInUse}.  This mapping is used to translate MTF indices back to
     * actual byte values during block decode.
     *
     * <p>Oracle: {@code BZLib.makeMaps(BZState)}.
     * obf: a(ac)
     */
    private static final void makeMaps(DecodeBuffer s) {
        s.nInUse = 0;
        for (int i = 0; i < 256; i++) {
            if (s.inUse[i]) {
                s.setToUnseq[s.nInUse] = (byte) i;
                s.nInUse++;
            }
        }
    }

    // -----------------------------------------------------------------------
    // Bit-stream helpers
    // -----------------------------------------------------------------------

    /**
     * Reads exactly 8 bits from the input bit-stream and returns the result as
     * a signed byte.
     *
     * <p>Oracle: {@code BZLib.getUchar(BZState)}.
     * obf: c(ac) : byte
     */
    private static final byte readByte(DecodeBuffer s) {
        return (byte) readBits(8, s);
    }

    /**
     * Reads exactly 1 bit from the input bit-stream and returns it as a byte
     * (0 or 1).
     *
     * <p>Oracle: {@code BZLib.getBit(BZState)}.
     * obf: d(ac) : byte
     */
    private static final byte readBit(DecodeBuffer s) {
        return (byte) readBits(1, s);
    }

    /**
     * Reads {@code n} bits from the big-endian input bit-stream, refilling the
     * 32-bit accumulator ({@code s.bsBuff}) from the next input byte whenever
     * fewer than {@code n} bits are available in it.
     *
     * <p>Oracle: {@code BZLib.getBits(int, BZState)}.
     * obf: a(int, ac) : int
     *
     * @param n number of bits to read (1–24 in practice)
     * @param s decoder state (bit-buffer + input byte array)
     * @return the decoded value in the low {@code n} bits of the return value
     */
    private static final int readBits(int n, DecodeBuffer s) {
        // Refill the bit-buffer accumulator until it holds at least n bits.
        // Each input byte is consumed MSB-first (big-endian bit order).
        while (s.bsLive < n) {
            s.bsBuff = s.bsBuff << 8 | (s.input[s.nextIn] & 0xFF);
            s.bsLive += 8;
            s.nextIn++;
            s.totalInLo32++;
            // overflow of totalInLo32 into totalInHi32 is intentionally elided
            // (the original has an empty `if (s.c == 0) {}` block here)
        }
        // Extract the top n bits from the accumulator
        int value = (s.bsBuff >> (s.bsLive - n)) & ((1 << n) - 1);
        s.bsLive -= n;
        return value;
    }
}
