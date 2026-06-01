package client.data;

import java.io.IOException;

import client.net.Packet;            // b  — its static DataStore field (Packet.outgoingTelemetry = Packet.q) is read here
import client.net.Buffer;            // tb — destination buffer for the 24 record bytes
import client.net.ChatCipher;        // v  — isLetterOrDigit(char) used by isValidChatChar
import client.ui.CharTable;          // ga — ALLOWED_CHARS accented/special table
import client.util.DecodeBuffer;     // ac — CHAT_MARKUP_CHARS {'[', ']', '#'}
import client.util.ErrorHandler;     // i  — the dead "WIP" instance field is an ErrorHandler; also the catch sink (stripped)

/**
 * RecordLoader — obf: {@code f}
 *
 * <p>Loads a fixed-size 24-byte cache record from the game's in-memory {@link DataStore}
 * ({@code nb}) into a {@link Buffer} ({@code tb}) packet for transmission or further
 * processing. In practice every call site passes record ID {@code 22607}
 * (byte-offset 15 inside the DataStore, since the store base is {@code 22592}).
 *
 * <p>This class also contains a stand-alone character-validation helper
 * ({@link #isValidChatChar}) used by text-input paths to decide whether a typed character
 * is acceptable (alphanumeric, accented, or RSC markup glyph).
 *
 * <p>The class holds two decoded string arrays that belong to the trade/deposit-dialog
 * UI ({@code DEPOSIT_PROMPT_STRINGS}), suggesting that the obfuscator co-located
 * unrelated string literals here.
 *
 * <h3>Static {@code DataStore} access (important)</h3>
 * The DataStore read/seek in {@link #loadRecord} goes through {@code b.q}, which is a
 * <b>static</b> field access on the {@code Packet} class ({@code b}): {@code Packet.q}
 * is a {@code static nb q} (DataStore) field. It is NOT an instance field of any object
 * declared in this class. Confirmed from bytecode: {@code getstatic Field b.q:Lnb;}.
 * The unfortunate symbol collision (the static field VARIABLE in this class is also
 * named {@code b}, but is type {@code i}/ErrorHandler — see {@link #unusedErrorHandler})
 * is what earlier confused the model into inventing a Packet instance field here.
 *
 * <h3>Obfuscation stripped</h3>
 * <ul>
 *   <li>Opaque predicate {@code boolean bl = client.vh} (always false) and all dead
 *       {@code if(bl)/while(!bl)/break} branches removed.
 *   <li>Per-method profiling counters ({@code ++d}, {@code ++a}) deleted.
 *   <li>{@code try { BODY } catch (RuntimeException e) { throw ErrorHandler.a(e,"…"); }}
 *       wrappers unwrapped.
 *   <li>Anti-tamper guard {@code if (recordId != 22607) return;} removed; the caller
 *       always passes 22607, so the guard is always false.
 *   <li>Junk {@code ~} arithmetic used to obfuscate loop bounds and comparison direction
 *       simplified to plain Java comparisons.
 *   <li>XOR string pool ({@code z(z("…"))}) decoded; literal values provided in comments.
 * </ul>
 *
 * <h3>Field-vs-class name collision (corrected)</h3>
 * An earlier pass mis-modelled the static field {@code b} as a {@code Packet} instance
 * named {@code packet} and treated {@code b.q} as {@code packet.dataStore}. That was
 * WRONG. The bytecode is unambiguous:
 * <ul>
 *   <li>{@code putstatic Field b:Li;} — the static field VARIABLE {@code b} is type
 *       {@code i} (ErrorHandler), constructed as {@code new i("WIP", 2)}. It is a
 *       decorative/dead instance: {@code i(String,int)} just stores the int and discards
 *       the string. Nothing in this class reads it. ({@link #unusedErrorHandler})</li>
 *   <li>{@code getstatic Field b.q:Lnb;} — every {@code b.q} in {@link #loadRecord}
 *       resolves to the {@code static nb q} (DataStore) field of the {@code Packet}
 *       CLASS ({@code b}), accessed by class name (canonical
 *       {@code Packet.outgoingTelemetry}). See {@link #loadRecord}.</li>
 * </ul>
 * The two uses of the symbol {@code b} (the field variable of type ErrorHandler vs. the
 * Packet class) are unrelated; the decompiler did not confuse the types.
 */
final class RecordLoader {

    // -------------------------------------------------------------------------
    // Fields
    // -------------------------------------------------------------------------

    /**
     * Externally-assigned int array; never written inside this class.
     * Likely a shared scratch buffer injected by the obfuscator or a dead field.
     * obf: {@code f}
     */
    static int[] intArray; // obf: f

    /**
     * Dead profiling counter incremented at the start of {@link #loadRecord}.
     * Not used for any game logic; left as a deleted obfuscation artefact.
     * obf: {@code d}
     */
    // static int loadRecordCallCount; // obf: d  — DELETED (profiling counter)

    /**
     * Dead profiling counter incremented at the start of {@link #isValidChatChar}.
     * obf: {@code a}
     */
    // static int isValidChatCharCallCount; // obf: a  — DELETED (profiling counter)

    /**
     * Externally-assigned string table; written by the client, read here as UI copy.
     * Exact content depends on which client locale/state populates it.
     * obf: {@code e}
     */
    static String[] stringTable; // obf: e

    /**
     * Two-element decoded string array used in the trade / deposit dialog UI.
     * Decoded from XOR pool:
     *   index 0 → "Please enter the number of items to deposit"
     *   index 1 → "and press enter"
     * obf: {@code c}
     */
    // obf: c
    static String[] DEPOSIT_PROMPT_STRINGS = new String[]{
        // z(z("c\fh!cV@h.dV\x12-4xV@c5}Q\x05\x7f`\x7fU@d4u^\x13-4\x7f\x13\x04h0\x7f@\ty"))
        "Please enter the number of items to deposit",
        // z(z("R\x0ei``A\x05~30V\x0ey%b"))
        "and press enter"
    };

    /**
     * Decorative / dead {@link ErrorHandler} instance ({@code i}).
     *
     * <p>The runtime type really is {@code i} (ErrorHandler), constructed as
     * {@code new i("WIP", 2)}. The {@code i(String,int)} constructor merely stores the
     * int into its {@code a} field and discards the string, so this object carries no
     * state anyone reads. Nothing in this class (or, as far as the bytecode shows, the
     * DataStore path) ever references it.
     *
     * <p><b>Do not confuse this with the DataStore access in {@link #loadRecord}.</b>
     * That goes through {@code b.q}, which is {@code Packet.q} — a STATIC field on the
     * {@code Packet} class ({@code b}) — and has nothing to do with this field variable,
     * which merely happens to share the obfuscated name {@code b}. Confirmed by bytecode
     * ({@code putstatic Field b:Li;} vs. {@code getstatic Field b.q:Lnb;}).
     *
     * <p>Constructor args: name {@code "WIP"} (decoded from XOR pool {@code z(z("d)]"))})
     * and int {@code 2}.
     * obf: {@code static i b}
     */
    // obf: b  — type i (ErrorHandler); a dead new i("WIP", 2) instance, never read.
    static ErrorHandler unusedErrorHandler = new ErrorHandler(
        /* name (discarded) = */ "WIP",   // z(z("d)]")) decoded
        /* value = */ 2
    );

    /**
     * Per-class XOR string-pool array used in exception-site error messages.
     * Decoded values (mod-5 key table {51,96,13,64,16}):
     *   z[0] → "f.A("    (method-signature prefix for isValidChatChar error)
     *   z[1] → "null"    (null-argument sentinel)
     *   z[2] → "f.B("    (method-signature prefix for loadRecord error)
     *   z[3] → "{...}"   (non-null object placeholder)
     * obf: {@code z}
     */
    // obf: z — private static final String[]
    // These are only used in the ErrorHandler.a() call at the bottom of each method
    // to produce a traceable call signature string on exception.


    // -------------------------------------------------------------------------
    // Public API
    // -------------------------------------------------------------------------

    /**
     * Determines whether a character is acceptable for RSC chat / text input.
     *
     * <p>A character is valid if it is any of:
     * <ol>
     *   <li>NOT a Java ISO control character (checked first; ISO controls → {@code false}).
     *   <li>An ASCII alphanumeric (A–Z, a–z, 0–9) checked via {@link ChatCipher#isLetterOrDigit}.
     *   <li>Present in the accented-character table {@code CharTable.ALLOWED_CHARS}
     *       ({@code ga.a}) at or after position {@code tableStartIndex}.
     *   <li>One of the three RSC chat-markup characters: {@code '['}, {@code ']'}, {@code '#'}
     *       stored in {@code DecodeBuffer.CHAT_MARKUP_CHARS} ({@code ac.I}).
     * </ol>
     *
     * <p>The {@code tableStartIndex} parameter controls which slice of the accented-char
     * table is searched, allowing callers to restrict valid accented characters by context.
     *
     * @param ch             the character to validate
     * @param tableStartIndex start index into {@code CharTable.ALLOWED_CHARS} for the search
     * @return {@code true} if the character is valid for chat input, {@code false} otherwise
     *
     * obf: {@code static final boolean a(char, int)}
     */
    // obf: a(char, int) -> isValidChatChar(char, int)
    static final boolean isValidChatChar(char ch, int tableStartIndex) {
        // Step 1: reject ISO control characters outright (Tab, LF, CR, DEL, etc.)
        if (Character.isISOControl(ch)) {
            return false;
        }

        // Step 2: accept ASCII alphanumeric characters (A-Z, a-z, 0-9).
        // Canonical ChatCipher.isLetterOrDigit(char) drops the obfuscator's dummy int
        // (the original v.a(ch, 115) had a dead anti-tamper 2nd param).
        if (ChatCipher.isLetterOrDigit(ch)) {  // obf: v.a(char, int) with dummy 115
            return true;
        }

        // Step 3: search the accented/special character table (CharTable.ALLOWED_CHARS / ga.a)
        // starting at tableStartIndex, scanning forward to the end of the array.
        char[] allowed = CharTable.ALLOWED_CHARS;  // obf: ga.a
        for (int i = tableStartIndex; i < allowed.length; i++) {
            if (ch == allowed[i]) {
                return true;
            }
        }

        // Step 4: check RSC chat markup characters: '[', ']', '#'  (ac.I)
        char[] markupChars = DecodeBuffer.CHAT_MARKUP_CHARS;  // obf: ac.I  = {'[', ']', '#'}
        for (int i = 0; i < markupChars.length; i++) {
            if (markupChars[i] == ch) {
                return true;
            }
        }

        return false;
    }

    /**
     * Reads the first 24 bytes of the shared {@code Packet.outgoingTelemetry}
     * {@link DataStore} ({@code Packet.q}) into the supplied output {@link Buffer}.
     *
     * <h3>Protocol / storage layout</h3>
     * <ul>
     *   <li>The DataStore is seeked to position {@code 0}. The second argument to
     *       {@code DataStore.seek(long,int)} — here {@code (recordId - RECORD_BASE_OFFSET)}
     *       — is the method's <b>discarded anti-tamper dummy</b>, NOT a byte offset.
     *       So {@code recordId} does not address anything; the read always starts at 0.
     *       (Confirmed against the clean DataStore source: {@code seek} stores only the
     *       long arg into {@code position} and divides the int arg into junk.)
     *   <li>24 bytes are read from position 0 into a local scratch buffer via
     *       {@code DataStore.write(byte,byte[])} (named "write" but it delegates to
     *       {@code read(true, len, 0, dest)} — it reads INTO {@code dest}).
     *   <li>If all 24 bytes are zero the record is considered missing/empty and an
     *       {@link IOException} is raised, which triggers the fallback path.
     *   <li>On any exception (read failure, seek failure, or the all-zeros sentinel) the
     *       scratch buffer is filled with {@code 0xFF} bytes — the conventional
     *       "not present" sentinel in this engine — and the method proceeds to write
     *       those sentinel bytes.
     *   <li>The 24 scratch bytes are then written into {@code outBuffer} via
     *       {@link Buffer#putBytes(int, int, int, byte[])}.
     * </ul>
     *
     * <h3>Obfuscation stripped</h3>
     * <ul>
     *   <li>Loop condition written using bitwise-NOT trick
     *       ({@code while(-25 < ~i)}) simplified to {@code while(i < 24)}.
     *   <li>Loop exit detection using {@code n4=-25; n3=~n5; if(n4>=n3)} simplified to
     *       {@code if(n5 >= 24)} (i.e. all bytes were zero → IOException).
     *   <li>Anti-tamper guard {@code if (recordId != 22607) return;} removed;
     *       the caller always passes 22607.
     * </ul>
     *
     * @param recordId  tamper/guard value; in practice always {@code 22607}. It does NOT
     *                  index the store (the seek goes to position 0 regardless); its only
     *                  live use is the {@code != 22607} guard that gates the final write.
     * @param outBuffer destination Buffer to receive the 24 record bytes
     *                  (typically a {@link BitBuffer}/ja instance from the active stream)
     *
     * obf: {@code static final void a(int, tb)}
     */
    // obf: a(int, tb) -> loadRecord(int, Buffer)
    static final void loadRecord(int recordId, Buffer outBuffer) {
        // Scratch buffer that receives raw bytes from the DataStore.
        byte[] recordBytes = new byte[RECORD_SIZE];

        // recordId - RECORD_BASE_OFFSET is passed as the DataStore.seek() dummy arg
        // (discarded by the store); it is NOT an offset. Kept only to mirror the original.
        final int seekDummyArg = recordId - RECORD_BASE_OFFSET;

        try {
            // b.q == Packet.outgoingTelemetry — a STATIC DataStore field on the Packet class.
            if (Packet.outgoingTelemetry != null) {  // obf: b.q != null  (Packet.q / nb)
                // Seek the DataStore to position 0. The second arg is the store's discarded
                // anti-tamper dummy (NOT a byte offset) — see DataStore.seek(long,int).
                Packet.outgoingTelemetry.seek(0L, seekDummyArg);  // obf: b.q.a(0L, var0 + -22592)

                // Read exactly RECORD_SIZE (24) bytes from position 0 into recordBytes.
                // DataStore.write(byte,byte[]) is misleadingly named: it delegates to
                // read(true,len,0,dest), copying INTO recordBytes. The -123 byte is a dummy.
                Packet.outgoingTelemetry.write((byte) -123, recordBytes);  // obf: b.q.a((byte)-123, var2)

                // Validate that the record is not all-zero (which would mean "not found").
                // The obfuscated code uses ~-trick:  n4=-25; n3=~n5; if(n4>=n3) → if(n5>=24).
                int firstNonZero = 0;
                while (firstNonZero < RECORD_SIZE) {
                    if (recordBytes[firstNonZero] != 0) {
                        break;  // found a non-zero byte → record has content
                    }
                    firstNonZero++;
                }
                if (firstNonZero >= RECORD_SIZE) {
                    // All 24 bytes are 0x00: treat as a missing / empty record.
                    throw new IOException();
                }
            }
            // If Packet.outgoingTelemetry == null: recordBytes stays all-zero, falls
            // through to the write below (writes zeroes — caller must handle this case).

        } catch (Exception ex) {
            // On any read failure (IOException, seek error, or all-zeros sentinel):
            // fill the scratch buffer with 0xFF (the "record absent" marker used
            // throughout the RSC engine).
            for (int i = 0; i < RECORD_SIZE; i++) {
                recordBytes[i] = -1;  // 0xFF
            }
        }

        // Write the 24 scratch bytes into the output Buffer.
        // Canonical Buffer.putBytes(srcOffset, length, src) drops the obfuscator's middle
        // dummy arg (-126). Copies recordBytes[0..23] into the buffer at its cursor.
        outBuffer.putBytes(0, RECORD_SIZE, recordBytes);  // obf: var1.a(0, -126, 24, var2)
    }


    // -------------------------------------------------------------------------
    // Constants
    // -------------------------------------------------------------------------

    /**
     * Number of bytes in each cache record handled by this class.
     * Every seek + read targets exactly 24 bytes.
     */
    private static final int RECORD_SIZE = 24;

    /**
     * Subtrahend applied to {@code recordId} to form the discarded {@code seek()} dummy
     * arg: {@code seekDummyArg = recordId - RECORD_BASE_OFFSET}. With
     * {@code recordId = 22607} this is {@code 15}, but the DataStore ignores it and the
     * read always starts at position 0 — so this does NOT address a byte offset.
     * Value: {@code 22592 = 0x5840}.
     */
    private static final int RECORD_BASE_OFFSET = 22592; // 0x5840


    // -------------------------------------------------------------------------
    // XOR string-pool helpers (obfuscation infrastructure — do not call)
    // -------------------------------------------------------------------------

    /**
     * XOR-decode step 1: converts a String to a char[] applying a 1-char key.
     * If the input has fewer than 2 characters the single character is XORed with 16
     * (key used across all classes in this build is class-specific; here it is 16).
     * For inputs of length ≥ 2 the array is returned unchanged; step 2 does the full decode.
     *
     * <p>This method exists only to support the static initialiser string pool;
     * all string literals have been decoded above and this helper is dead at runtime.
     *
     * obf: {@code private static char[] z(String)}
     */
    // obf: z(String) -> private helper; all strings already decoded as literals above
    @SuppressWarnings("unused")
    private static char[] xorDecodeStep1(String encoded) {
        // obf: z(String)
        char[] chars = encoded.toCharArray();
        if (chars.length < 2) {
            // Single-char case: XOR with per-class key 16
            chars[0] = (char) (chars[0] ^ 16);
        }
        // length >= 2: returned as-is; step-2 handles the full decode
        return chars;
    }

    /**
     * XOR-decode step 2: converts a char[] (from step 1) to the final String using
     * a cyclic 5-byte key table: {@code {51, 96, 13, 64, 16}} (indices 0–4 repeating).
     *
     * <p>Each character at position {@code i} is XORed with {@code key[i % 5]}.
     * The result is interned.
     *
     * <p>This method exists only to support the static initialiser string pool;
     * all string literals have been decoded above and this helper is dead at runtime.
     *
     * obf: {@code private static String z(char[])}
     */
    // obf: z(char[]) -> private helper; all strings already decoded as literals above
    @SuppressWarnings("unused")
    private static String xorDecodeStep2(char[] chars) {
        // obf: z(char[])
        // Key table (mod-5 cyclic XOR):  {51, 96, 13, 64, 16}
        final int[] KEY = {51, 96, 13, 64, 16};
        for (int i = 0; i < chars.length; i++) {
            chars[i] = (char) (chars[i] ^ KEY[i % 5]);
        }
        return new String(chars).intern();
    }
}
