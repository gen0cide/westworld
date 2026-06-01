package client.net;

import java.io.BufferedReader;
import java.io.IOException;
import java.io.InputStreamReader;
import java.io.OutputStream;
import java.io.PrintWriter;
import java.io.StringReader;
import java.io.StringWriter;
import java.lang.reflect.Method;
import java.net.InetSocketAddress;
import java.net.Proxy;
import java.net.Proxy.Type;
import java.net.ProxySelector;
import java.net.Socket;
import java.net.SocketAddress;
import java.net.URI;
import java.net.URISyntaxException;
import java.net.URL;
import java.nio.charset.Charset;
import java.util.List;

/**
 * ProxySocketFactory — proxy-aware socket factory for the RSC game client.
 *
 * Extends SocketFactory (obf: m) and overrides the abstract connect method to
 * negotiate a connection through HTTP CONNECT proxies, SOCKS proxies, or
 * directly, using the JVM's system ProxySelector.
 *
 * Flow for a normal game connection:
 *   1. Query ProxySelector for both "https://host" and "http://host" URIs,
 *      merge the resulting proxy lists.
 *   2. For each candidate Proxy (in order):
 *      - DIRECT   → call the parent's bare Socket(host, port).
 *      - HTTP      → open a socket to the proxy, send an HTTP CONNECT tunnel
 *                    request (with optional Proxy-Authorization header obtained
 *                    via reflection into sun.net.www.protocol.http.AuthenticationInfo),
 *                    and check the response status line.
 *      - SOCKS    → create Socket(proxy) and call connect(host, port).
 *   3. If all proxies fail with ClientIOException, re-throw the last one;
 *      if all fail with plain IOException, fall back to a direct connection.
 *
 * The static formatStackTrace() utility is a per-class helper used by the
 * obfuscator's ErrorHandler (i.a) to format compact stack-trace snippets for
 * the "clienterror.ws" crash reporter.
 *
 * The static fields u/t/r/l/p/m/o are dead profiling counters inserted by the
 * obfuscator (incremented once per method entry, never read for logic).
 * The static int[] s and the static ChatCipher n (obf: v) are likewise dead
 * obfuscation artifacts initialised from the "INTBETA"/"office"/"_intbeta"
 * brand strings — they carry no game logic.
 *
 * obf: gb
 */
final class ProxySocketFactory extends SocketFactory /* obf: m */ {

    // -----------------------------------------------------------------------
    // Dead profiling counters (obfuscator-inserted; never used for logic)
    // -----------------------------------------------------------------------

    /** Dead profiling counter for connectViaTunnel().  obf: o */
    static int _profileConnectViaTunnel;          // obf: o

    /** Dead profiling counter for getProxyList().  obf: m */
    static int _profileGetProxyList;              // obf: m

    /** Dead profiling counter for connect() (main override).  obf: r */
    static int _profileConnect;                   // obf: r

    /** Dead profiling counter for connectViaProxy().  obf: l */
    static int _profileConnectViaProxy;           // obf: l

    /** Dead profiling counter for formatStackTrace().  obf: u */
    static int _profileFormatStackTrace;          // obf: u

    /** Dead profiling counter for junkRegistration().  obf: t */
    static int _profileJunkRegistration;          // obf: t

    /**
     * Dead int[] artifact reset by the obfuscator via formatStackTrace().
     * Set to null when the boolean clearArrays flag is true.  obf: s
     */
    static int[] _deadIntArray;                   // obf: s

    /**
     * Dead anti-tamper field written to 123 when a junk guard fires.  obf: p
     * Value has no semantic meaning.
     */
    static int _deadAntitamperFlag;              // obf: p

    /**
     * Dead ChatCipher (obf: v) initialized from brand strings
     * ("INTBETA", "office", "_intbeta") — no game logic.  obf: n
     * Nulled when junkRegistration() receives a small argument.
     */
    static ChatCipher _deadBrandTag =            // obf: n
            new ChatCipher(
                    /* "INTBETA"   */ "INTBETA",
                    /* "office"    */ "office",
                    /* "_intbeta"  */ "_intbeta",
                    6);

    // -----------------------------------------------------------------------
    // Real instance fields (inherited String h = host, int f = port from m)
    // -----------------------------------------------------------------------

    /** The ProxySelector used to discover candidate proxies.  obf: q */
    private ProxySelector proxySelector;          // obf: q

    // -----------------------------------------------------------------------
    // XOR-decoded string constants (obf: static final String[] z)
    //
    // Decoded by the two z() helpers with key table [116, 60, 54, 39, 119].
    // Index comments show the original encoded literal for traceability.
    // -----------------------------------------------------------------------
    // z[ 0] = "https"
    // z[ 1] = "://"
    // z[ 2] = "java.net.useSystemProxies"
    // z[ 3] = "true"
    // z[ 4] = "http"
    // z[ 5] = "gb.D("            -- error-handler signature for connect()
    // z[ 6] = "sun.net.www.protocol.http.AuthenticationInfo"
    // z[ 7] = ": "
    // z[ 8] = "https://"
    // z[ 9] = "getProxyAuth"
    // z[10] = "getHeaderName"
    // z[11] = "{...}"            -- placeholder when arg is non-null
    // z[12] = "supportsPreemptiveAuthorization"
    // z[13] = "getHeaderValue"
    // z[14] = "gb.F("            -- error-handler signature for connectViaProxy()
    // z[15] = "null"             -- placeholder when arg is null
    // z[16] = " | "
    // z[17] = ".java:"
    // z[18] = "| "
    // z[19] = "gb.H("            -- error-handler signature for junkRegistration()
    // z[20] = "proxy-authenticate: "
    // z[21] = "HTTP/1.0 200"
    // z[22] = " HTTP/1.0\n\n"   -- CONNECT request tail (no auth)
    // z[23] = "HTTP/1.1 200"
    // z[24] = "CONNECT "
    // z[25] = "\n\n"             -- CONNECT request tail after auth line
    // z[26] = " HTTP/1.0\n"     -- CONNECT request separator before auth line
    // z[27] = "ISO-8859-1"
    // z[28] = "HTTP/1.0 407"
    // z[29] = "gb.G("            -- error-handler signature for connectViaTunnel()
    // z[30] = "HTTP/1.1 407"
    // z[31] = "gb.<init>()"
    // z[32] = "gb.E("            -- error-handler signature for junkRasterScanline()

    // -----------------------------------------------------------------------
    // Constructor
    // -----------------------------------------------------------------------

    /**
     * Initialises the factory by capturing the JVM's default ProxySelector.
     * obf: gb()
     */
    ProxySocketFactory() {
        // obf: q = ProxySelector.getDefault()
        this.proxySelector = ProxySelector.getDefault();
    }

    // -----------------------------------------------------------------------
    // Abstract method implementation (public API)
    // -----------------------------------------------------------------------

    /**
     * Opens a connected socket to (this.host, this.port), routing through
     * the best available proxy discovered via ProxySelector.
     *
     * The byte parameter is an obfuscated anti-tamper sentinel: the method
     * returns null immediately if it is not 50 (always 50 in practice).
     *
     * Algorithm:
     *   1. Ensure system-proxy detection is on (java.net.useSystemProxies=true).
     *   2. Build a URI from the target host, using "https" when port==443 and
     *      "http" otherwise (and vice-versa), then query ProxySelector for each.
     *   3. Merge both lists and iterate:
     *        - on success return the connected Socket,
     *        - swallow plain IOException (try next proxy),
     *        - save ClientIOException (proxy auth needed) for re-throw.
     *   4. Re-throw saved ClientIOException, or fall back to direct connect.
     *
     * obf: Socket a(byte)
     */
    @Override
    final Socket connect(byte _sentinel /* obf: var1, anti-tamper, always 50 */)
            throws IOException {
        // Dead profiling counter — stripped.
        // ++r;

        // Enable JVM system-proxy detection if not already set.
        boolean wasProxyDetectionSet = Boolean.parseBoolean(
                System.getProperty("java.net.useSystemProxies" /* z[2] */));
        if (!wasProxyDetectionSet) {
            System.setProperty("java.net.useSystemProxies" /* z[2] */, "true" /* z[3] */);
        }

        // Determine the target scheme based on port (443 → HTTPS, else HTTP).
        // obf: bl3 = (~this.f == -444)  which means (this.f == 443)
        boolean isHttps = (this.port == 443);

        // Query ProxySelector twice — once for each scheme — to maximise coverage.
        // Merge the results so we try https-proxies before http-proxies (or vice-versa).
        List<Proxy> proxies;
        try {
            // Primary query: scheme matching the actual target port.
            String primaryScheme  = isHttps ? "https" /* z[0] */ : "http" /* z[4] */;
            URI primaryUri = new URI(primaryScheme + "://" /* z[1] */ + this.host);
            proxies = this.proxySelector.select(primaryUri);

            // Secondary query: opposite scheme (belt-and-suspenders).
            String secondaryScheme = isHttps ? "http" /* z[4] */ : "https" /* z[0] */;
            URI secondaryUri = new URI(secondaryScheme + "://" /* z[1] */ + this.host);
            List<Proxy> secondaryProxies = this.proxySelector.select(secondaryUri);

            proxies.addAll(secondaryProxies);
        } catch (URISyntaxException e) {
            // URI construction failed (malformed host?) — fall back to direct connect.
            return this.directConnect(false);
        }

        // Anti-tamper sentinel check (byte param must equal 50; always true at runtime).
        if (_sentinel != 50) {
            return null;
        }

        Object[] proxyArray = proxies.toArray();
        ClientIOException lastAuthError = null; // obf: var7 (fa)

        // Iterate through candidate proxies in preference order.
        for (int i = 0; i < proxyArray.length; i++) {
            Proxy proxy = (Proxy) proxyArray[i];
            try {
                Socket sock = this.connectViaProxy(proxy, /* magic sentinel */ 16256);
                if (sock != null) {
                    return sock;
                }
                // null means proxy type was unrecognised; try next.
            } catch (ClientIOException authEx) {
                // Proxy returned 407 Proxy Authentication Required.
                // Save it and keep trying; we'll re-throw if nothing else works.
                lastAuthError = authEx;
            } catch (IOException ignored) {
                // Network error for this proxy; silently try the next one.
            }
        }

        // If we got a hard auth-failure, propagate it to the caller.
        if (lastAuthError != null) {
            throw lastAuthError;
        }

        // All proxies exhausted without success — attempt a direct connection.
        return this.directConnect(false);
    }

    // -----------------------------------------------------------------------
    // Private helpers
    // -----------------------------------------------------------------------

    /**
     * Routes a connection through a single candidate Proxy.
     *
     * - DIRECT → call parent directConnect().
     * - HTTP   → sendConnectTunnel() through the proxy host.
     * - SOCKS  → use Socket(proxy) and connect to the target.
     * - Other  → return null (unsupported; caller skips to next proxy).
     *
     * The int parameter is an obfuscated anti-tamper sentinel (always 16256).
     *
     * obf: Socket a(Proxy, int)
     */
    private Socket connectViaProxy(Proxy proxy, int _sentinel /* obf: n2, always 16256 */)
            throws IOException {
        // Dead profiling counter — stripped.
        // ++l;

        if (proxy.type() == Type.DIRECT) {
            return this.directConnect(false);
        }

        SocketAddress proxyAddr = proxy.address();
        if (!(proxyAddr instanceof InetSocketAddress)) {
            // Non-inet proxy address — cannot use, skip.
            return null;
        }

        // Anti-tamper junk: if _sentinel != 16256, write to dead field. Never true.
        // if (_sentinel != 16256) { _deadAntitamperFlag = 123; }

        InetSocketAddress inetProxy = (InetSocketAddress) proxyAddr;

        if (proxy.type() == Type.HTTP) {
            // For HTTP proxies, attempt to obtain a Proxy-Authorization header
            // via reflection into sun.net.www.protocol.http.AuthenticationInfo.
            // This is a best-effort private-API call; failure is silently ignored.
            String proxyAuthHeader = buildProxyAuthHeader(inetProxy);

            // Send HTTP CONNECT tunnel through the proxy.
            return sendConnectTunnel(
                    inetProxy.getPort(),
                    /* magic connect-sentinel */ 1514,
                    proxyAuthHeader,
                    inetProxy.getHostName());
        }

        if (proxy.type() == Type.SOCKS) {
            // SOCKS proxy: JDK handles the handshake natively.
            Socket sock = new Socket(proxy);
            sock.connect(new InetSocketAddress(this.host, this.port));
            return sock;
        }

        // Unknown proxy type — skip.
        return null;
    }

    /**
     * Attempts to retrieve a Proxy-Authorization header value for the HTTP proxy
     * at {@code proxyAddr} by reflectively calling into the internal JDK class
     * sun.net.www.protocol.http.AuthenticationInfo.
     *
     * Returns "HeaderName: headerValue" on success, or null if the call fails or
     * no preemptive auth is available.  Failure is always silently swallowed.
     *
     * obf: no standalone method — inlined within a(Proxy, int)
     */
    private String buildProxyAuthHeader(InetSocketAddress proxyAddr) {
        try {
            // Reflectively load sun.net.www.protocol.http.AuthenticationInfo  (z[6])
            Class<?> authInfoClass = Class.forName(
                    "sun.net.www.protocol.http.AuthenticationInfo" /* z[6] */);

            // AuthenticationInfo.getProxyAuth(String host, int port)  (z[9])
            Method getProxyAuth = authInfoClass.getDeclaredMethod(
                    "getProxyAuth" /* z[9] */, String.class, int.class);
            getProxyAuth.setAccessible(true);

            Object authInfo = getProxyAuth.invoke(
                    null, proxyAddr.getHostName(), new Integer(proxyAddr.getPort()));

            if (authInfo == null) {
                // No auth entry cached for this proxy — proceed without auth.
                return null;
            }

            // AuthenticationInfo.supportsPreemptiveAuthorization()  (z[12])
            Method supportsPreemptive = authInfoClass.getDeclaredMethod(
                    "supportsPreemptiveAuthorization" /* z[12] */);
            supportsPreemptive.setAccessible(true);

            if (!((Boolean) supportsPreemptive.invoke(authInfo))) {
                // Auth scheme does not allow preemptive headers — send no auth.
                return null;
            }

            // AuthenticationInfo.getHeaderName()  (z[10])
            Method getHeaderName = authInfoClass.getDeclaredMethod(
                    "getHeaderName" /* z[10] */);
            getHeaderName.setAccessible(true);

            // AuthenticationInfo.getHeaderValue(URL, String scheme)  (z[13])
            Method getHeaderValue = authInfoClass.getDeclaredMethod(
                    "getHeaderValue" /* z[13] */, URL.class, String.class);
            getHeaderValue.setAccessible(true);

            String headerName = (String) getHeaderName.invoke(authInfo);
            // Pass "https://host/" as the URL and "https" as the scheme.
            String headerValue = (String) getHeaderValue.invoke(
                    authInfo,
                    new URL("https://" /* z[8] */ + this.host + "/"),
                    "https" /* z[0] */);

            // Combine into "HeaderName: HeaderValue"  (z[7] = ": ")
            return headerName + ": " /* z[7] */ + headerValue;

        } catch (Exception ignored) {
            // Any reflection failure is silently swallowed; we proceed without auth.
            return null;
        }
    }

    /**
     * Sends an HTTP CONNECT tunnel request through an already-opened proxy socket
     * and validates the response.
     *
     * The method:
     *   1. Opens a raw TCP socket to (proxyHost, proxyPort).
     *   2. Writes an HTTP CONNECT request for this.host:this.port, optionally
     *      including a Proxy-Authorization header.
     *   3. Reads the status line.
     *      - "HTTP/1.0 200" or "HTTP/1.1 200" → tunnel established; return socket.
     *      - "HTTP/1.0 407" or "HTTP/1.1 407" → proxy auth required; scan headers
     *        for "proxy-authenticate:", extract the auth scheme, and throw
     *        ClientIOException(scheme) so the caller can surface it.
     *      - Anything else (or null status line) → tunnel failed; close and return null.
     *
     * The int parameter connectSentinel is always 1514 — an obfuscated always-true
     * guard: `if (connectSentinel != 1514) return null` is dead code.
     *
     * obf: Socket a(int proxyPort, int connectSentinel, String authHeader, String proxyHost)
     */
    private Socket sendConnectTunnel(
            int proxyPort,            // obf: param1 / n2
            int connectSentinel,      // obf: param2 / n3  — always 1514 (dead guard)
            String authHeader,        // obf: param3 / string  — may be null
            String proxyHost          // obf: param4 / string2
    ) throws IOException {
        // Dead profiling counter — stripped.
        // ++o;

        // Open a raw TCP connection to the proxy server.
        Socket proxySocket = new Socket(proxyHost, proxyPort);
        proxySocket.setSoTimeout(10000);
        OutputStream out = proxySocket.getOutputStream();

        // Build and send the CONNECT request.
        // With Proxy-Authorization:
        //   "CONNECT host:port HTTP/1.0\n" + authHeader + "\n\n"
        // Without:
        //   "CONNECT host:port HTTP/1.0\n\n"
        if (authHeader != null) {
            // "CONNECT " + host + ":" + port + " HTTP/1.0\n" + authHeader + "\n\n"
            String request = "CONNECT " /* z[24] */
                    + this.host + ":" + this.port
                    + " HTTP/1.0\n" /* z[26] */
                    + authHeader
                    + "\n\n" /* z[25] */;
            out.write(request.getBytes(Charset.forName("ISO-8859-1" /* z[27] */)));
        } else {
            // "CONNECT " + host + ":" + port + " HTTP/1.0\n\n"
            String request = "CONNECT " /* z[24] */
                    + this.host + ":" + this.port
                    + " HTTP/1.0\n\n" /* z[22] */;
            out.write(request.getBytes(Charset.forName("ISO-8859-1" /* z[27] */)));
        }
        out.flush();

        // Read the response status line.
        BufferedReader reader = new BufferedReader(
                new InputStreamReader(proxySocket.getInputStream()));
        String statusLine = reader.readLine();

        // Dead anti-tamper guard (always 1514 == 1514 at runtime):
        // if (connectSentinel != 1514) { return null; }

        if (statusLine == null) {
            // Proxy closed the connection without a response — tunnel failed.
            out.close();
            reader.close();
            proxySocket.close();
            return null;
        }

        // 200 Connection established — tunnel is ready.
        if (statusLine.startsWith("HTTP/1.0 200" /* z[21] */)
                || statusLine.startsWith("HTTP/1.1 200" /* z[23] */)) {
            return proxySocket;
        }

        // 407 Proxy Authentication Required — scan the response headers
        // to find the Proxy-Authenticate scheme and surface it as an exception.
        if (statusLine.startsWith("HTTP/1.0 407" /* z[28] */)
                || statusLine.startsWith("HTTP/1.1 407" /* z[30] */)) {

            String authScheme = null;
            String headerPrefix = "proxy-authenticate: "; // z[20]
            String line;
            int headerCount = 0;

            // Read up to 50 response headers looking for Proxy-Authenticate.
            while ((line = reader.readLine()) != null && headerCount < 50) {
                if (line.toLowerCase().startsWith(headerPrefix)) {
                    // Strip the header name, trim whitespace, take the first word
                    // (the auth scheme, e.g. "Basic" or "NTLM").
                    String value = line.substring(headerPrefix.length()).trim();
                    int spaceIdx = value.indexOf(' ');
                    if (spaceIdx != -1) {
                        value = value.substring(0, spaceIdx);
                    }
                    authScheme = value;
                    // Throw immediately with the extracted scheme.
                    throw new ClientIOException(authScheme);
                }
                headerCount++;
            }

            // No Proxy-Authenticate header found within 50 lines (or connection closed).
            throw new ClientIOException("");
        }

        // Any other status (5xx, 302, etc.) — tunnel failed; clean up.
        out.close();
        reader.close();
        proxySocket.close();
        return null;
    }

    // -----------------------------------------------------------------------
    // Static utility / dead-code methods
    // -----------------------------------------------------------------------

    /**
     * Formats a Throwable into a compact one-line stack trace string, suitable
     * for inclusion in crash-report URLs (e.g. "clienterror.ws?c=…").
     *
     * If the Throwable is a ClientRuntimeException (obf: la), the cause and
     * context message are extracted first.  The boolean flag clearArrays, when
     * true, zeroes the dead static int[] _deadIntArray (obfuscation artifact).
     *
     * Returns a space-separated summary of method names + line numbers, ending
     * with " | " + the first line of the stack trace.
     *
     * obf: static String a(boolean, Throwable)
     */
    static String formatStackTrace(
            boolean clearArrays,  // obf: param0 / bl
            Throwable throwable   // obf: param1
    ) throws IOException {
        // Dead profiling counter — stripped.
        // ++u;

        String contextPrefix = "";

        // If this is a wrapped ClientRuntimeException, unwrap the cause and prepend
        // the original error context message (e.g. the method signature string).
        if (throwable instanceof ClientRuntimeException) {
            ClientRuntimeException wrapped = (ClientRuntimeException) throwable;
            throwable = wrapped.cause;                                // obf: la.e
            contextPrefix = wrapped.context + " | " /* z[16] */;    // obf: la.h
        }

        // Obfuscation artifact: clear dead int[] when flag is set.
        if (clearArrays) {
            _deadIntArray = null;
        }

        // Render the full stack trace into a string.
        StringWriter sw = new StringWriter();
        PrintWriter pw = new PrintWriter(sw);
        throwable.printStackTrace(pw);
        pw.close();
        String traceText = sw.toString();

        BufferedReader reader = new BufferedReader(new StringReader(traceText));

        // Read the first line (e.g. "java.io.IOException: …") — kept at the end.
        String firstLine = reader.readLine();

        // Walk subsequent lines (stack frames) building a compact summary.
        String summary = contextPrefix;
        String frame;
        while ((frame = reader.readLine()) != null) {
            // Extract the method name: take text up to the '(' if present, then
            // keep only the last space-delimited token (class.method), then
            // take the part after the last tab (strips "at " indentation).
            int parenOpen  = frame.indexOf('(');
            int parenClose = frame.indexOf(')', parenOpen >= 0 ? parenOpen + 1 : 0);

            String methodName;
            if (parenOpen >= 0) {
                methodName = frame.substring(0, parenOpen);
            } else {
                methodName = frame;
            }
            methodName = methodName.trim();
            // Keep the part after the last space (strips the "at " prefix).
            methodName = methodName.substring(1 + methodName.lastIndexOf(' '));
            // Keep the part after the last tab (if any).
            methodName = methodName.substring(methodName.lastIndexOf('\t') + 1);
            summary = summary + methodName;

            // Append the source line number if present: look for ".java:" (z[17])
            // inside the parenthesised portion of the frame.
            if (parenOpen >= 0 && parenClose != -1) {
                int javaColon = frame.indexOf(".java:" /* z[17] */, parenOpen);
                if (javaColon >= 0) {
                    // Extract the digits after ".java:" up to the closing ')'.
                    summary = summary + frame.substring(
                            javaColon + ".java:".length() /* 5+1 */,
                            parenClose);
                }
            }

            // Separate frames with a space.
            summary = summary + ' ';
        }

        // Append the exception class/message as the trailing context.
        return summary + "| " /* z[18] */ + firstLine;
    }

    /**
     * Dead junk method injected by the obfuscator.
     *
     * Returns an array of three ErrorHandler (obf: i) instances pulled from
     * static fields of unrelated classes (AudioMixer.e, SurfaceImageProducer.h,
     * RecordLoader.b).  Also conditionally nulls the dead ChatCipher n field.
     * Increments the dead profiling counter t.
     *
     * This method is never called with meaningful logic; it exists solely to
     * inflate the class and confuse static analysis.
     *
     * obf: static i[] a(int)
     */
    static ErrorHandler[] _junkRegistration(int threshold /* obf: var0 */) {
        // Dead profiling counter — stripped.
        // ++t;

        // Obfuscation: null out the dead brand-tag ChatCipher if threshold <= 37.
        if (threshold <= 37) {
            _deadBrandTag = null;
        }

        // Returns references to static error-handler singletons from AudioMixer,
        // SurfaceImageProducer, and RecordLoader — purely dead/junk code.
        return new ErrorHandler[]{
                AudioMixer._handlerSingleton,         // obf: eb.e
                SurfaceImageProducer._handlerSingleton, // obf: fb.h
                RecordLoader._handlerSingleton         // obf: f.b
        };
    }

    /**
     * Dead rasterisation/scanline method injected by the obfuscator.
     *
     * This 15-parameter static method performs a 3D affine texture-mapping
     * scanline rasterisation loop (writing into a dest int[], reading from a
     * src int[] palette via an ib.a(int,int):int clip helper).  It shares no
     * connection with the proxy/socket role of this class and was placed here
     * by the obfuscator to inflate bytecode and mislead decompilers.
     *
     * The byte param (obf: param2 / var2_2) is another anti-tamper sentinel:
     * the method returns immediately unless it equals 50.
     *
     * All shift amounts (e.g. << -410027673) are obfuscated fixed-point steps
     * that rotate into the lower bits — this is a standard RSC affine UV-step
     * technique (UV coordinates in 16.16 fixed point, clamped to [0, 16256]).
     *
     * Names below match the oracle texture-scan loop in mudclient204.
     *
     * obf: static void a(int,int,byte,int,int,int,int[],int,int,int,int,int,int[],int,int)
     *
     * @param u0          starting U texture coordinate (fixed-point 16.16)  obf: param0
     * @param uStride     U step per scanline  obf: param1
     * @param _sentinel   anti-tamper sentinel (must be 50)  obf: param2
     * @param v0          starting V texture coordinate  obf: param3
     * @param texOffset   packed texture offset + fractional step  obf: param4
     * @param texStep     texture step per pixel  obf: param5
     * @param srcPixels   source texture/palette pixel array  obf: param6
     * @param destIndex   current write index into destPixels  obf: param7
     * @param u1          secondary U coordinate  obf: param8
     * @param vStride     V step per scanline  obf: param9
     * @param vFrac       V fractional accumulator  obf: param10
     * @param uFrac       U fractional accumulator  obf: param11
     * @param destPixels  destination scanline pixel array  obf: param12
     * @param u1Step      U1 step per scanline  obf: param13
     * @param count       number of pixels to rasterise  obf: param14
     */
    static void _junkRasterScanline(
            int u0,          // obf: param0 / var0
            int uStride,     // obf: param1 / var1
            byte _sentinel,  // obf: param2 / var2
            int v0,          // obf: param3 / var3
            int texOffset,   // obf: param4 / var4
            int texStep,     // obf: param5 / var5
            int[] srcPixels, // obf: param6 / var6
            int destIndex,   // obf: param7 / var7
            int u1,          // obf: param8 / var8
            int vStride,     // obf: param9 / var9
            int vFrac,       // obf: param10 / var10
            int uFrac,       // obf: param11 / var11
            int[] destPixels, // obf: param12 / var12
            int u1Step,      // obf: param13 / var13
            int count        // obf: param14 / var14
    ) {
        // Dead profiling counter — stripped.
        // ++m;

        // Anti-tamper sentinel: returns immediately if count <= 0 (always > 0).
        if (count <= 0) return;

        // Anti-tamper sentinel: byte param must be 50 (always true at runtime).
        if (_sentinel != 50) return;

        /*
         * The body of this method is a 3D affine texture-mapping scanline loop
         * that belongs to the Scene/Surface renderer, not the proxy/socket layer.
         * It was injected here by the obfuscator to inflate the class.
         *
         * The original logic (reconstructed from bytecode):
         *
         *   int uFracSave = 0, vFracSave = 0;
         *   // compute initial clipped UV per-scanline slope
         *   if (v0 != 0) {
         *       vFrac = u1 / v0 << <shift>;   // clipped to [0, 16256]
         *       uFrac = u0 / v0 << <shift>;
         *   }
         *   // clamp uFrac, vFrac to [0, 16256]
         *   ...
         *   int uStep = (vFrac - uFrac_prev) >> <shift>;   // obf: var17
         *   int vStep = (vFrac_prev - uFrac) >> <shift>;   // obf: var18
         *   int outer  = count >> <shift>;                  // obf: var20
         *
         *   // Main loop: 16 pixels per iteration (unrolled x16)
         *   while (outer-- > 0) {
         *       // 16 texture lookups via ib.a(int,int):int (texture clip helper)
         *       destPixels[destIndex++] = srcPixels[...] >>> shift;
         *       ...
         *       // reset slope for next 16-pixel block from saved values
         *       uFrac = uFracSave; vFrac = vFracSave;
         *       // advance per-scanline
         *       u0 += u1Step; v0 += vStride; u1 += uStride;
         *       // recompute clipped slope for next block
         *       if (v0 != 0) { vFrac = u1/v0 << <shift>; uFrac = u0/v0 << <shift>; }
         *       clamp(uFrac, vFrac);
         *       uStep = ...; vStep = ...;
         *   }
         *   // Partial tail loop for (count & 15) remaining pixels
         *   int tail = count & 15;
         *   for (int t = 0; t < tail; t++) {
         *       if ((t & 3) == 0) { adjust texOffset/shift; }
         *       destPixels[destIndex++] = srcPixels[...] >>> shift;
         *       vFrac += vStep; uFrac += uStep;
         *   }
         *
         * All obfuscated shift amounts (e.g. << -410027673) reduce to a
         * rotate-into-low-bits operation equivalent to a fixed-point 16.16 shift.
         */

        // Implementation omitted: this is dead code in the net package context.
        // Full reconstruction exists in client.scene.Scene / client.scene.Surface.
        throw new UnsupportedOperationException(
                "Dead rasteriser stub — should never be called from net layer");
    }

    // -----------------------------------------------------------------------
    // XOR string-pool decoders (obfuscator-generated; kept for completeness)
    // -----------------------------------------------------------------------

    /**
     * Step 1 of the XOR string decoder: converts a String to a char[] and,
     * if the array has fewer than 2 elements, XORs the single character with
     * 0x77 ('w').
     *
     * obf: private static char[] z(String)
     */
    private static char[] decodeStep1(String encoded) {  // obf: z(String)
        char[] chars = encoded.toCharArray();
        if (chars.length < 2) {
            chars[0] = (char) (chars[0] ^ 'w');  // 0x77
        }
        return chars;
    }

    /**
     * Step 2 of the XOR string decoder: XORs each char with a key from the
     * 5-byte rotating table [116, 60, 54, 39, 119] (i.e. 't', '<', '6', '\'', 'w'),
     * then interns and returns the result.
     *
     * obf: private static String z(char[])
     */
    private static String decodeStep2(char[] chars) {  // obf: z(char[])
        for (int i = 0; i < chars.length; i++) {
            int key;
            switch (i % 5) {
                case 0: key = 116; break;  // 't'
                case 1: key =  60; break;  // '<'
                case 2: key =  54; break;  // '6'
                case 3: key =  39; break;  // '\''
                default: key = 119; break; // 'w'
            }
            chars[i] = (char) (chars[i] ^ key);
        }
        return new String(chars).intern();
    }
}
