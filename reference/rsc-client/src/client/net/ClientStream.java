package client.net;

import java.applet.Applet;
import java.io.IOException;
import java.io.InputStream;
import java.io.OutputStream;
import java.net.Socket;
import java.net.URL;

import client.shell.GameShell;
import client.scene.SurfaceImageProducer;
import client.scene.ImageLoader;
import client.scene.LoaderThread;
import client.data.CacheFile;
import client.data.DataStore;
import client.net.DownloadWorker;
import client.net.Buffer;
import client.net.ISAAC;
import client.net.ChatCipher;
import client.util.ErrorHandler;
import client.util.Utility;
import client.shell.LoaderThread;

/**
 * ClientStream — obfuscated class "da", a subclass of {@link Packet} ("b").
 *
 * <p>This is the concrete network endpoint for the RuneScape Classic client. It owns the live
 * {@link Socket} together with its {@link InputStream}/{@link OutputStream}, and supplies the raw
 * byte transport that the {@link Packet} base class layers the RSC packet framing on top of.</p>
 *
 * <p>Reads are performed synchronously on the caller's thread ({@link #readStream},
 * {@link #readStreamBytes}, {@link #availableStream}). Writes, however, are decoupled from the game loop:
 * {@link #writeStreamBytes} copies outgoing bytes into a 5000-byte ring buffer and signals a
 * dedicated writer thread (this class implements {@link Runnable}); the writer thread drains the
 * ring buffer to the socket in {@link #run}, so a slow/blocking socket never stalls the client.</p>
 *
 * <p>It also exposes a couple of unrelated static helpers that happened to be folded into this class
 * by the obfuscator: a small loading-progress reporter ({@link #reportProgress}), a synchronous
 * HTTP cache download ({@link #downloadFile}), and a world tile/region packing helper
 * ({@link #packRegion}).</p>
 *
 * <p>Deobfuscation notes: every method in the original was wrapped in an opaque-predicate
 * ({@code client.vh}, always false), a per-method profiling counter, an {@link ErrorHandler}
 * try/catch that only decorated the stack trace, and anti-tamper guards comparing a dummy parameter
 * against a magic constant. All of those have been stripped; the surviving logic matches the
 * deobfuscated RSC-204 {@code ClientStream}.</p>
 */
public final class ClientStream extends Packet implements Runnable {

    /** Ring-buffer capacity for outgoing bytes (must match the packet max length). */
    private static final int WRITE_BUFFER_SIZE = 5000;

    // --- Decoded string constants (original obfuscated array "hb"). ---------------------------
    /** Logged when closing the socket/streams fails. */
    private static final String MSG_ERROR_CLOSING = "Error closing stream";
    /** Prefix attached to socket-exception messages raised by the writer thread. */
    private static final String MSG_WRITER_PREFIX = "Twriter:";
    /** Thrown when a synchronous read hits end-of-stream prematurely. */
    private static final String MSG_EOF = "EOF";
    /** Thrown when the outgoing ring buffer would overrun the un-flushed region. */
    private static final String MSG_BUFFER_OVERFLOW = "buffer overflow";
    /** Thrown when {@link #downloadFile} cannot retrieve the requested resource. */
    private static final String MSG_DOWNLOAD_FAILED = "Couldn't download file";

    // --- Static fields carried over from the obfuscated class (mostly unrelated scratch). ------
    /** Obfuscated "J": shared 2-D int scratch table (unused here; kept for layout fidelity). */
    public static int[][] sharedIntTable2d;
    /** Obfuscated "T": shared int scratch array. */
    public static int[] sharedIntArrayT;
    /** Obfuscated "N": shared int scratch array. */
    public static int[] sharedIntArrayN;
    /** Obfuscated "db": shared off-screen image producer used by the scene renderer. */
    public static SurfaceImageProducer sharedImageProducer = new SurfaceImageProducer();
    /** Obfuscated "gb": applet handle used by static download/progress helpers. */
    public static Applet applet;
    /** Obfuscated "O": shared chat (de)cipher seeded with fixed key material ("WTRC","office","_rc"). */
    public static ChatCipher chatCipher = new ChatCipher("WTRC", "office", "_rc", 1);

    // Profiling counters (obfuscated S/R/P/eb/M/V/cb/L/I/H/bb/K), retained but no longer incremented.
    public static int profCountWriteStreamBytes; // S
    public static int profCountRun;              // R
    public static int profCountReadStream;       // P (obf b(boolean) single-byte read)
    public static int profCountReadStreamBytes;  // eb
    public static int profCountM;                // M
    public static int profCountClose;            // V
    public static int profCountPackRegion;       // cb
    public static int profCountDownload;         // L
    public static int profCountReportProgress;   // I
    public static int profCountAvailable;        // H
    public static int profCountBB;               // bb
    public static int profCountK;                // K

    // --- Instance state. ----------------------------------------------------------------------
    /** Single-byte scratch buffer for {@link #readStream}. (obf "Z") */
    private byte[] singleByteScratch = new byte[1];
    /** Read end of the socket. (obf "U") */
    private InputStream inputStream;
    /** Write end of the socket. (obf "Q") */
    private OutputStream outputStream;
    /** The underlying TCP socket. (obf "ab") */
    private Socket socket;
    /** Outgoing ring buffer drained by the writer thread; lazily allocated. (obf "Y") */
    private byte[] writeBuffer;
    /** Producer index — next free slot in {@link #writeBuffer}. (obf "G") */
    private int writeHead;
    /** Consumer index — next byte the writer thread will flush. (obf "W") */
    private int writeTail = 0;
    /** True once the stream is closing; short-circuits all read/write paths. (obf "fb") */
    private boolean closing = false;
    /** True before the socket is wired up and after close; stops the writer thread. (obf "X") */
    private boolean closed = true;

    /**
     * Wraps a connected socket, grabs its streams, and starts the background writer thread.
     * (obf "&lt;init&gt;(Socket, e)"; the GameShell starts the Runnable via {@code startThread}.)
     */
    public ClientStream(Socket socket, GameShell shell) throws IOException {
        this.socket = socket;
        this.inputStream = socket.getInputStream();
        this.outputStream = socket.getOutputStream();
        this.closed = false;
        // The leading "1" was a J++ thread-kind discriminator; semantically: shell.startThread(this).
        shell.startThread(1, this);
    }

    /**
     * Writer thread body: blocks until bytes are queued, then drains the ring buffer to the socket
     * and flushes when the buffer is fully consumed. (obf "run")
     */
    @Override
    public final void run() {
        while (!this.closed) {
            int flushOffset;
            int flushLength;
            synchronized (this) {
                // Sleep until writeStreamBytes() produces something (or close() wakes us).
                if (this.writeHead == this.writeTail) {
                    try {
                        this.wait();
                    } catch (InterruptedException ignored) {
                    }
                }
                if (this.closed) {
                    return;
                }
                flushOffset = this.writeTail;
                // obf: clean base branches on (writeTail > writeHead) -> (5000 - tail),
                // ELSE -> (head - tail). The two arms must keep this exact ordering: when
                // head == tail the else-arm yields a flush length of 0 (nothing to send),
                // not 5000 - tail. Mirror the original branch condition rather than flip it.
                if (this.writeTail > this.writeHead) {
                    // Producer wrapped around: flush only up to the end of the buffer this pass.
                    flushLength = WRITE_BUFFER_SIZE - this.writeTail;
                } else {
                    // Contiguous region: tail .. head (zero when head == tail).
                    flushLength = this.writeHead - this.writeTail;
                }
            }

            if (flushLength > 0) { // original wrote this as "-1 <= ~len", i.e. len > 0
                try {
                    this.outputStream.write(this.writeBuffer, flushOffset, flushLength);
                } catch (IOException ex) {
                    this.socketException = true;
                    this.socketExceptionMessage = MSG_WRITER_PREFIX + ex;
                }
                this.writeTail = (this.writeTail + flushLength) % WRITE_BUFFER_SIZE;
                try {
                    if (this.writeTail == this.writeHead) {
                        this.outputStream.flush();
                    }
                } catch (IOException ex) {
                    this.socketException = true;
                    this.socketExceptionMessage = MSG_WRITER_PREFIX + ex;
                }
            }
        }
    }

    /**
     * Queues {@code length} bytes for asynchronous transmission, copying them into the ring buffer
     * and waking the writer thread. (obf "a(byte[], off, len, (byte)-67)" — overrides Packet's
     * write hook; the {@code (byte)-67} arg was an anti-tamper magic constant, dropped here.)
     */
    @Override
    public final void writeStreamBytes(byte[] data, int offset, int length, byte ignoredMagic) throws IOException {
        if (this.closing) {
            return;
        }
        if (this.writeBuffer == null) {
            this.writeBuffer = new byte[WRITE_BUFFER_SIZE];
        }
        synchronized (this) {
            for (int i = 0; i < length; i++) {
                this.writeBuffer[this.writeHead] = data[offset + i];
                this.writeHead = (this.writeHead + 1) % WRITE_BUFFER_SIZE;
                // Keep a 100-byte guard band so the producer never laps the consumer.
                if (this.writeHead == (this.writeTail + 4900) % WRITE_BUFFER_SIZE) {
                    throw new IOException(MSG_BUFFER_OVERFLOW);
                }
            }
            this.notify();
        }
    }

    /**
     * Synchronously reads exactly {@code length} bytes into {@code dest} starting at {@code offset}.
     * (obf "a(byte[], len, off, int)" — overrides Packet's bulk-read hook; the trailing int was a
     * dead anti-tamper computation, dropped here.)
     */
    @Override
    public final void readStreamBytes(byte[] dest, int length, int offset, int ignoredMagic) throws IOException {
        if (this.closing) {
            return;
        }
        for (int read = 0, n = 0; read < length; read += n) {
            n = this.inputStream.read(dest, offset + read, length - read);
            if (n <= 0) {
                throw new IOException(MSG_EOF);
            }
        }
    }

    /**
     * Returns the number of bytes available without blocking, or 0 while closing. (obf "b(byte)" —
     * overrides Packet's {@link Packet#availableStream availableStream} hook; the byte arg fed a dead
     * anti-tamper division.) Method name must match the base hook so the override actually binds.
     */
    @Override
    public final int availableStream(byte ignoredMagic) throws IOException {
        if (this.closing) {
            return 0;
        }
        return this.inputStream.available();
    }

    /**
     * Reads a single byte (0..255) via the buffered bulk-read path. (obf "b(boolean)" — overrides
     * Packet's {@link Packet#readStream readStream} single-byte hook; the name must match the base
     * hook so the override actually binds.) The boolean selects this concrete implementation: when
     * the stream is open ({@code closing==false}) it must equal the caller's expected flag, otherwise
     * 0 is returned to indicate "no data on this code path".
     */
    @Override
    public final int readStream(boolean expectOpen) throws IOException {
        // Original guard: (!closing) != expectOpen  ->  return 0.
        if ((!this.closing) != expectOpen) {
            return 0;
        }
        this.readStreamBytes(this.singleByteScratch, 1, 0, 123);
        return this.singleByteScratch[0] & 0xFF;
    }

    /**
     * Closes the streams and socket, stops the writer thread, and releases the ring buffer.
     * (obf "a(boolean)" — overrides Packet's close hook; the flag becomes the {@code closing} state.)
     */
    @Override
    public final void closeStream(boolean closingFlag) {
        super.closeStream(true);
        this.closing = closingFlag;
        try {
            if (this.inputStream != null) {
                this.inputStream.close();
            }
            if (this.outputStream != null) {
                this.outputStream.close();
            }
            if (this.socket != null) {
                this.socket.close();
            }
        } catch (IOException ex) {
            System.out.println(MSG_ERROR_CLOSING);
        }
        this.closed = true;
        synchronized (this) {
            this.notify(); // wake the writer thread so run() can observe closed==true and exit
        }
        this.writeBuffer = null;
    }

    /**
     * Posts a loading-progress line ("&lt;label&gt; - &lt;percent&gt;%") to the game shell.
     * (obf static "a(String, int, int)"; the trailing int was an anti-tamper recursion guard.)
     */
    public static final void reportProgress(String label, int percent, int ignoredMagic) {
        // CacheFile.gameShell.showLoadingProgress(DataStore.progressContext, ...)
        CacheFile.gameShell.showLoadingProgress(DataStore.loadingProgress, (byte) -101,
                label + ISAAC.statusPrefix + " - " + percent + "%");
    }

    /**
     * Synchronously downloads a resource via a {@link DownloadWorker}, optionally reporting progress,
     * and returns the decompressed payload bytes. (obf static "a(URL, boolean, boolean)".)
     */
    public static final byte[] downloadFile(URL url, boolean withProgress, boolean ignoredMagic) throws IOException {
        DownloadWorker worker = new DownloadWorker(ImageLoader.loaderThread, url, 2000000);
        if (withProgress) {
            reportProgress("", 0, 0);
        }
        // Poll the worker until the download completes.
        while (!worker.isComplete(-2)) {
            Utility.sleep(11200, 50L);
        }
        Buffer payload = worker.getPayload((byte) -120);
        if (payload == null) {
            throw new IOException(MSG_DOWNLOAD_FAILED);
        }
        if (withProgress) {
            reportProgress("", 100, 0);
        }
        return payload.decompress(0);
    }

    /**
     * Packs a world (x, y, z) coordinate triple into the negative region/sector index RSC uses for
     * map addressing. (obf static "a(int, byte, int, int)"; the byte arg was an anti-tamper guard.)
     * Each axis is divided into 8-tile sectors and folded into one packed value.
     */
    public static final int packRegion(int x, byte ignoredMagic, int y, int z) {
        return -(z / 8 * 32) - 1 - (y / 8 * 1024) - x / 8;
    }

    // --- Obfuscation helpers retained for reference. ------------------------------------------

    /**
     * Per-character XOR key schedule used to decode the obfuscated string-constant pool ("z(char[])").
     * Kept for fidelity; the decoded values are inlined above as readable constants.
     */
    private static String decodeString(char[] chars) {
        for (int i = 0; i < chars.length; i++) {
            int key;
            switch (i % 5) {
                case 0:  key = 84;  break;
                case 1:  key = 34;  break;
                case 2:  key = 104; break;
                case 3:  key = 120; break;
                default: key = 50;
            }
            chars[i] = (char) (chars[i] ^ key);
        }
        return new String(chars).intern();
    }

    /**
     * First-stage transform of the string-decode pipeline ("z(String)"): for single-char literals it
     * XORs the lone character with '2'. Kept for fidelity.
     */
    private static char[] decodeStage1(String s) {
        char[] chars = s.toCharArray();
        if (chars.length < 2) {
            chars[0] = (char) (chars[0] ^ '2');
        }
        return chars;
    }
}
