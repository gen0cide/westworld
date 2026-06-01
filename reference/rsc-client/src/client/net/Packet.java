package client.net;

import java.io.IOException;

import client.scene.SpriteScaler;      // ia — hosts one of the two per-opcode stat arrays (ia.d = opcodeFrequencyTable)
import client.net.SocketFactory;       // m  — hosts the other per-opcode stat array (m.i = packetSizeByOpcode)
import client.net.StreamBase;          // ib — base stream helper; StreamBase.bitwiseAnd(v, 0xFF) == v & 0xFF
import client.net.Buffer;              // tb — Buffer base; telemetry takes one (obf a(tb,int,int))
import client.net.BitBuffer;           // ja — the concrete outgoing buffer (extends Buffer); obf field f
import client.data.DataStore;          // nb — optional outgoing-packet telemetry sink
import client.util.ErrorHandler;       // i  — obfuscation's try/catch context wrapper (stripped here)

/**
 * Packet — base network packet/stream for the RuneScape Classic protocol (rev ~233-235).
 *
 * <p>This is the read/write framing layer that sits underneath {@link client.net.ClientStream}
 * (obf {@code da}), which adds the actual socket + background writer thread. {@code Packet}
 * itself only knows how to:
 * <ul>
 *   <li>Assemble an <b>outgoing</b> packet into a {@link client.net.BitBuffer} (obf {@code ja},
 *       a {@link client.net.Buffer} subclass): reserve a header, write a body via the
 *       {@code put*} helpers (defined on {@code Buffer}),
 *       then {@link #finishPacket()} which back-patches the 1- or 2-byte length header and
 *       encrypts the opcode byte with the outgoing ISAAC cipher.</li>
 *   <li>Reassemble an <b>incoming</b> packet from the raw byte stream: read the length header
 *       (1 byte, or 2 bytes when {@code >= 160}), then the body ({@link #readPacket}).</li>
 *   <li>Hold the two {@link client.net.ISAAC} keystream ciphers — one per direction — seeded
 *       from a shared key after login ({@link #seedIsaac}). The keystream advance call is
 *       {@code ISAAC.getNextValue()} (no args; the obf {@code c(int)} junk param was dropped).</li>
 * </ul>
 *
 * <p>RSC's wire framing: every packet is {@code [length][opcode][payload...]}. The length field
 * is one byte for bodies under 160, otherwise two bytes encoded as
 * {@code hi = 160 + len/256}, {@code lo = len & 0xFF}. The opcode byte is offset by the next
 * ISAAC keystream value (per direction) so opcodes are not transmitted in the clear.
 *
 * <p>Most of the low-level stream methods here are no-op stubs returning 0 / doing nothing;
 * {@link client.net.ClientStream} overrides them with real socket I/O. Original obfuscated name
 * was {@code b}; the per-method profiling counters, {@code client.vh} opaque predicates, and
 * {@code ErrorHandler.a(...)} context wrappers have been stripped for readability.
 */
public class Packet {

    // ------------------------------------------------------------------
    // Profiling counters (obf: static int w,b,i,s,o,t,D,r,l,C,a,n,x,u,y,F,p).
    // Each method bumped one of these on entry; pure obfuscation/telemetry,
    // never read for logic. Kept only so member count matches the original.
    // ------------------------------------------------------------------
    public static int profWritePacket;       // w
    public static int profReadPacketInto;    // b
    public static int profWritePacketBump;   // i
    public static int profReadPacketBody;    // s
    public static int profHasPacket;         // o
    public static int profFormatString;      // t
    public static int profAvailableStub;     // D
    public static int profIsaacCommand;      // r
    public static int profReadBytes;         // l
    public static int profFlushPacket;       // C
    public static int profTelemetry;         // a
    public static int profCloseStream;       // n
    public static int profFinishPacket;      // x
    public static int profSeedIsaac;         // u
    public static int profReadStreamBytes;   // y
    public static int profReadStreamDispatch;// F
    public static int profReadStreamStub;    // p

    /** Unused legacy scratch buffer (obf {@code v}); only ever nulled by telemetry. */
    public static byte[] legacyBytes;
    /** Unused legacy bit-mask table (obf {@code h}); the powers-of-two table, never populated. */
    public static int[] legacyMaskTable;

    /**
     * Font-glyph buffer write cursor / size counter (obf {@code static int c = 0}; clean b.java:28).
     * Used by the UI font loader ({@link client.ui.Panel}/{@link client.ui.FontBuilder}) as the
     * running write position into {@link client.shell.GameFrame#unusedByteBuffer} while rasterizing
     * glyphs, and as the final length when slicing each glyph block out of the staging buffer.
     */
    public static int writePos = 0;

    /** Optional telemetry sink for outgoing packets (obf {@code q}); null in normal play. */
    public static DataStore outgoingTelemetry = null;

    /** Human-readable message captured when the socket faults (obf {@code B}). */
    public String socketExceptionMessage;
    /** Incoming ISAAC keystream cipher — decrypts received opcodes (obf {@code j}). */
    private ISAAC isaacIncoming;
    /** Outgoing ISAAC keystream cipher — encrypts sent opcodes (obf {@code z}). */
    private ISAAC isaacOutgoing;

    /** Number of consecutive read attempts since the last full packet (obf {@code g}). */
    private int readTries;
    /** Length (bytes) of the incoming packet currently being reassembled (obf {@code A}). */
    private int incomingLength;
    /** Max read attempts before declaring a time-out; 0 = unlimited (obf {@code d}). */
    public int maxReadTries = 0;
    /** Outgoing-flush throttle: counts {@link #writePacket} calls (obf {@code e}). */
    private int writeDelay;
    /** Capacity of the outgoing {@link Buffer} (obf {@code m}); also gates telemetry. */
    private int packetMaxLength;
    /** Start offset of the current outgoing packet within the buffer (obf {@code E}). */
    private int packetStart;

    /** True once a socket read/write has failed; surfaced on the next flush (obf {@code k}). */
    public boolean socketException;

    /**
     * Outgoing write buffer (obf {@code f}, declared/constructed as {@code ja} = {@link BitBuffer},
     * which extends {@link Buffer}): {@code outBuffer.data} is the byte array (obf {@code tb.F}) and
     * {@code outBuffer.offset} is the write cursor / packet end (obf {@code tb.w}). The {@code put*}
     * body writers live on {@link Buffer}; bit-level writers (used elsewhere, e.g. {@code Jh.f.d(...)})
     * live on {@link BitBuffer}, which is why the declared type is {@code BitBuffer}, not {@code Buffer}.
     */
    public BitBuffer outBuffer;

    /**
     * Obfuscated error-context labels (obf {@code G}). Decoded at class-load via {@link #deobf}.
     * Used only to build exception messages in the original; the readable logic ignores them.
     * Decoded values: signature stubs ("b.X(") plus G[1]="{...}", G[2]="null", G[14]="time-out".
     */
    private static final String[] ERR = new String[]{
        deobf(deobf("]4R")),            //  0
        deobf(deobf("D4:+")),           //  1  "{...}"
        deobf(deobf("QoxV")),                 //  2  "null"
        deobf(deobf("]4E")),            //  3
        deobf(deobf("]4(S8Vn*")), //  4  "b.<init>()"
        deobf(deobf("]4G")),            //  5
        deobf(deobf("]4Y")),            //  6
        deobf(deobf("]4P")),            //  7
        deobf(deobf("]4@")),            //  8
        deobf(deobf("]4X")),            //  9
        deobf(deobf("]4^")),            // 10
        deobf(deobf("]4Z")),            // 11
        deobf(deobf("]4]")),            // 12
        deobf(deobf("]4[")),            // 13
        deobf(deobf("Ksy_{Po`")),             // 14  "time-out"
        deobf(deobf("]4D")),            // 15
        deobf(deobf("]4_")),            // 16
        deobf(deobf("]4V")),            // 17
        deobf(deobf("]4U")),            // 18
        deobf(deobf("]4F")),            // 19
        deobf(deobf("]4S"))             // 20
    };

    /** Build the base packet stream with a 5000-byte outgoing buffer; header reserves 3 bytes. */
    protected Packet() {
        this.socketExceptionMessage = "";
        this.writeDelay = 0;
        this.packetMaxLength = 5000;
        this.socketException = false;
        this.incomingLength = 0;
        this.packetStart = 0;
        this.readTries = 0;

        this.outBuffer = new BitBuffer(this.packetMaxLength); // obf: new ja(this.m)
        this.outBuffer.offset = 3; // leave room for [len-hi][len-lo][opcode]
    }

    /**
     * Flush any queued outgoing bytes if {@code writeDelay} has reached {@code maxDelay},
     * surfacing a pending socket fault as an {@link IOException} (obf {@code a(int,boolean)}).
     *
     * @param maxDelay     flush threshold (0 = flush now)
     * @param skipPreFlush when false, runs {@link #closeStream}-style pre-flush hook first
     */
    public final void writePacket(int maxDelay, boolean skipPreFlush) throws IOException {
        if (!skipPreFlush) {
            this.closeStream(true);
        }

        if (this.socketException) {
            this.outBuffer.offset = 3;
            this.packetStart = 0;
            this.socketException = false;
            throw new IOException(this.socketExceptionMessage);
        }

        this.writeDelay++;
        if (maxDelay <= this.writeDelay) {
            if (this.packetStart > 0) { // (~packetStart <= -1) i.e. packetStart >= 1
                this.writeDelay = 0;
                // Send the assembled bytes [0, packetStart) to the socket.
                this.writeStreamBytes(this.outBuffer.data, 0, this.packetStart, (byte) -67);
            }
            this.outBuffer.offset = 3;
            this.packetStart = 0;
        }
    }

    /**
     * Finalize then immediately flush the current outgoing packet (obf {@code a(int)}).
     *
     * @param guardCode anti-tamper guard; live callers always pass {@code -6924}
     */
    public final void flushPacket(int guardCode) throws IOException {
        this.finishPacket(21294);
        this.writePacket(0, true);
        // obf: `if (guardCode != -6924) this.isaacIncoming = null;` — every live caller
        // passes -6924 (client.Jh.a(-6924)), so the guard is always false and the cipher
        // is NEVER nulled. Keeping the guard verbatim so the dead branch stays inert.
        if (guardCode != -6924) {
            this.isaacIncoming = null;
        }
    }

    /**
     * Read a single byte from the underlying stream (obf {@code b(boolean)}).
     * Base stub returns 0; {@link client.net.ClientStream} overrides with real socket reads.
     */
    public int readStream(boolean unusedFlag) throws IOException {
        if (!unusedFlag) {
            this.readTries = 126; // dead store on a constant-false path; preserved for fidelity
        }
        return 0;
    }

    /**
     * Block-read {@code len} bytes into {@code dest} from the socket (obf {@code a(byte[],int,int,byte)}).
     * Base stub no-ops; overridden by {@link client.net.ClientStream}. The trailing tag byte
     * is an anti-tamper discriminator (-67 on the real call path).
     */
    public void writeStreamBytes(byte[] dest, int off, int len, byte tag) throws IOException {
        if (tag != -67) {
            this.newPacket(81, 91); // dead anti-tamper branch
        }
    }

    /**
     * Decrypt a received opcode/byte with the incoming ISAAC keystream (obf {@code a(int,int)}).
     *
     * @param unusedGuard anti-tamper guard value (ignored)
     * @param value       the encrypted opcode byte
     * @return {@code (value - isaacIncoming.next()) & 0xFF}
     */
    public final int isaacCommand(int unusedGuard, int value) {
        // -isaacIncoming.getNextValue() + value, masked to a byte. obf passed the junk arg
        // (unusedGuard + -635) into ISAAC's c(int); that param was dead and is dropped here.
        return 0xFF & (-this.isaacIncoming.getNextValue() + value);
    }

    /**
     * Reassemble one incoming packet into {@code dest} (obf {@code a(byte[],int)}).
     *
     * <p>Mirrors the oracle {@code readPacket}: enforces the read-attempt time-out, reads the
     * 1- or 2-byte length header once available, then the body. Returns the body length on a
     * complete packet, otherwise 0 (more bytes still pending).
     *
     * @param dest             destination for the packet body
     * @param expectedZeroSync sentinel compared against {@code incomingLength} (0 on the real path)
     */
    private int readPacket(byte[] dest, int expectedZeroSync) {
        try {
            this.readTries++;

            // Time-out: too many empty read attempts without completing a packet.
            if (this.maxReadTries > 0 && this.readTries > this.maxReadTries) {
                this.socketException = true;
                this.socketExceptionMessage = ERR[14]; // "time-out"
                this.maxReadTries += this.maxReadTries;
                return 0;
            }

            // Read the length header once at least 2 bytes are buffered.
            if (this.incomingLength == expectedZeroSync && this.availableStream((byte) -124) >= 2) {
                this.incomingLength = this.readStream(true);
                if (this.incomingLength >= 160) {
                    // 2-byte length: hi = 160 + len/256, lo = next byte.
                    this.incomingLength = (this.incomingLength - 160) * 256 + this.readStream(true);
                }
            }

            // Read the body once the whole packet is buffered.
            if (this.incomingLength > 0 && this.availableStream((byte) -124) >= this.incomingLength) {
                if (this.incomingLength >= 160) {
                    this.readBytes(dest, (byte) 64, this.incomingLength);
                } else {
                    // Short packets stash the last byte first, then bulk-read the rest.
                    dest[this.incomingLength - 1] = (byte) this.readStream(true);
                    if (this.incomingLength > 1) {
                        this.readBytes(dest, (byte) 126, this.incomingLength - 1);
                    }
                }
                int bodyLength = this.incomingLength;
                this.readTries = 0;
                this.incomingLength = 0;
                return bodyLength;
            }
        } catch (IOException ioException) {
            this.socketException = true;
            this.socketExceptionMessage = ioException.getMessage();
        }
        return 0;
    }

    /**
     * Bytes available on the underlying stream (obf {@code b(byte)}).
     * Base stub returns 0; overridden by {@link client.net.ClientStream}.
     */
    public int availableStream(byte unusedTag) throws IOException {
        return 0;
    }

    /**
     * Read {@code len} bytes into {@code dest} starting at offset 0 (obf {@code a(byte[],byte,int)}).
     * Thin wrapper over {@link #readStreamBytes}; the tag byte is an anti-tamper guard (>= 51 on
     * the live path).
     */
    private void readBytes(byte[] dest, byte tag, int len) throws IOException {
        if (tag < 51) {
            return; // dead anti-tamper guard
        }
        this.readStreamBytes(dest, len, 0, -112);
    }

    /** True when an outgoing packet has bytes queued ahead of the cursor (obf {@code a(byte)}). */
    public final boolean hasPacket(byte unusedTag) {
        return this.packetStart > 0;
    }

    /**
     * Begin a new outgoing packet with the given opcode (obf {@code b(int,int)}).
     *
     * <p>Auto-flushes if the buffer is over 80% full, lazily allocates the buffer, writes the
     * opcode into the reserved header slot, and positions the write cursor past the 3-byte header.
     *
     * @param opcode    protocol opcode for this packet
     * @param sizeFlag  0 to write the opcode now (the normal "fixed/var packet start" path)
     */
    public final void newPacket(int opcode, int sizeFlag) {
        // Flush early if we're past 80% of capacity to avoid overflow.
        if (this.packetStart > 4 * this.packetMaxLength / 5) {
            try {
                this.writePacket(0, true);
            } catch (IOException ioException) {
                this.socketException = true;
                this.socketExceptionMessage = ioException.getMessage();
            }
        }
        this.outBuffer.offset = this.packetStart + 2; // cursor at the opcode slot
        if (sizeFlag == 0) {
            this.outBuffer.putByte(opcode); // obf: this.f.c(opcode, 82) — store opcode at [packetStart + 2]
        }
    }

    /**
     * Convenience: set the buffer cursor and reassemble a packet into it (obf {@code a(int,ja)}).
     * @param cursor    write cursor to install on {@code buf}
     * @param buf       buffer to read the packet into (obf type {@code ja} = {@link BitBuffer})
     */
    public final int readPacketInto(int cursor, BitBuffer buf) {
        buf.offset = cursor;
        return this.readPacket(buf.data, cursor ^ 0);
    }

    /**
     * Close/finish the underlying stream (obf {@code a(boolean)}).
     * Base stub; {@link client.net.ClientStream} closes the socket. The false branch re-syncs
     * the read cursor via {@link #readPacketInto}.
     */
    public void closeStream(boolean unusedFlag) {
        if (!unusedFlag) {
            this.readPacketInto(116, null);
        }
    }

    /**
     * Seed both ISAAC ciphers from a shared key (obf {@code a(byte,int[])}).
     * Called once after the login handshake establishes the session key.
     *
     * @param tag anti-tamper guard (ignored)
     * @param key 4-int (128-bit) shared session key
     */
    public final void seedIsaac(byte tag, int[] key) {
        if (tag >= -68) {
            this.maxReadTries = -84; // dead store on the live path; preserved
        }
        this.isaacIncoming = new ISAAC(key);
        this.isaacOutgoing = new ISAAC(key);
    }

    /**
     * Block-read into a buffer region (obf {@code a(byte[],int,int,int)}).
     * Base stub no-ops; overridden by {@link client.net.ClientStream} with real socket reads.
     */
    public void readStreamBytes(byte[] dest, int len, int off, int tag) throws IOException {
        // overridden in ClientStream
    }

    /**
     * Finalize the current outgoing packet: encrypt its opcode, back-patch the length header,
     * and advance the packet boundary (obf {@code b(int)}). RSC wire equivalent of "sendPacket".
     *
     * @param unusedGuard anti-tamper guard value (21294 on the live path)
     */
    public final void finishPacket(int unusedGuard) {
        byte[] data = this.outBuffer.data;

        // Encrypt the opcode byte (at packetStart+2) by adding the next outgoing keystream value.
        if (this.isaacOutgoing != null) {
            int opcodeByte = 0xFF & data[this.packetStart + 2];
            data[this.packetStart + 2] = (byte) (this.isaacOutgoing.getNextValue() + opcodeByte);
        }

        // Body length = (write cursor) - (packet start) - 2-byte length prefix space.
        int bodyLength = this.outBuffer.offset - this.packetStart - 2;
        if (bodyLength >= 160) { // (~(-161) < ~bodyLength)  <=>  bodyLength >= 160
            // 2-byte length header: hi = 160 + len/256, lo = len & 0xFF.
            data[this.packetStart] = (byte) (160 + bodyLength / 256);
            data[this.packetStart + 1] = (byte) StreamBase.bitwiseAnd(bodyLength, 255);
        } else {
            // 1-byte length header; reclaim the unused 2nd header byte by moving the last
            // body byte forward, then back the write cursor up by one.
            data[this.packetStart] = (byte) bodyLength;
            this.outBuffer.offset--;
            data[this.packetStart + 1] = data[this.outBuffer.offset];
        }

        // Per-opcode telemetry: count sends + bytes for buffers small enough to track.
        if (this.packetMaxLength <= 10000) { // (-10001 <= ~packetMaxLength)
            int opcode = data[this.packetStart + 2] & 0xFF;
            SpriteScaler.opcodeFrequencyTable[opcode]++;            // obf: ia.d[opcode]++
            SocketFactory.packetSizeByOpcode[opcode] += this.outBuffer.offset - this.packetStart; // obf: m.i[opcode]
        }

        // (obf wrote a dead static on a constant guard; dropped.)
        this.packetStart = this.outBuffer.offset; // advance boundary to the new end
    }

    /**
     * Sanitize a string to a restricted character set (obf static {@code a(int,byte,String)}).
     *
     * <p>Keeps lowercase {@code a-z}, uppercase {@code A-Z}, and digit-range punctuation
     * {@code '0'-'9'} (chars 48-57); maps everything else to {@code '_'}, padding the result to
     * {@code maxLen} with spaces. Used to scrub names/identifiers before transmission.
     *
     * @param maxLen      number of source characters to scan
     * @param tag         anti-tamper guard (-5 on the live path)
     * @param source      the input string
     */
    public static String formatString(int maxLen, byte tag, String source) {
        if (tag != -5) {
            telemetry(null, -63, 17); // dead anti-tamper side effect
        }
        String out = "";
        for (int idx = 0; idx < maxLen; idx++) {
            if (idx >= source.length()) {
                out = out + " "; // pad past end of input
                continue;
            }
            char ch = source.charAt(idx);
            if (ch >= 'a' && ch <= 'z') {            // (-98 < ~ch) => ch >= 97
                out = out + ch;
            } else if (ch >= 'A' && ch <= 'Z') {     // ('A' <= ch) && (~ch >= -91) => ch <= 90
                out = out + ch;
            } else if (ch >= '0' && ch <= '9') {     // (~ch <= -49 && ~ch >= -58) => 48..57
                out = out + ch;
            } else {
                out = out + '_';
            }
        }
        return out;
    }

    /**
     * Report a just-built outgoing packet to the optional telemetry sink (obf static {@code a(tb,int,int)}).
     * No-op when {@link #outgoingTelemetry} is null (the normal case).
     *
     * @param buf       the source buffer (its {@code data} array is forwarded)
     * @param idCode    obfuscated event id (XOR-masked before use)
     * @param length    number of bytes in the packet
     */
    public static void telemetry(Buffer buf, int idCode, int length) {
        if (outgoingTelemetry != null) {
            try {
                outgoingTelemetry.seek(0L, idCode ^ -26747);            // obf: q.a(0L, idCode ^ -26747)
                outgoingTelemetry.write(24, -107, length, buf.data);   // obf: q.a(24, -107, length, buf.data)
            } catch (Exception ignored) {
            }
        }
        if (idCode != 26628) {
            legacyBytes = null; // dead guard clears unused scratch
        }
    }

    /**
     * String deobfuscator stage 1 (obf {@code z(String)}): for the single-char wrapper inputs
     * used by {@link #ERR} this is effectively identity; XORs the lone char with 'V' otherwise.
     */
    private static char[] deobf(String s) {
        char[] chars = s.toCharArray();
        if (chars.length < 2) {
            chars[0] = (char) (chars[0] ^ 'V');
        }
        return chars;
    }

    /**
     * String deobfuscator stage 2 (obf {@code z(char[])}): XORs each char by a 5-key cycle
     * {63, 26, 20, 58, 86} keyed on its index, returning an interned String.
     */
    private static String deobf(char[] chars) {
        for (int i = 0; i < chars.length; i++) {
            byte key;
            switch (i % 5) {
                case 0:  key = 63; break;
                case 1:  key = 26; break;
                case 2:  key = 20; break;
                case 3:  key = 58; break;
                default: key = 86;
            }
            chars[i] = (char) (chars[i] ^ key);
        }
        return new String(chars).intern();
    }
}
