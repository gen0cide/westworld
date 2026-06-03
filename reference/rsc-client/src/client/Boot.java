package client;

import java.math.BigInteger;

import client.net.BitBuffer;
import client.ui.FontBuilder;

/**
 * Standalone launcher for the deobfuscated RSC client.
 *
 * Injects the runtime ADDs documented in
 * {@code docs/build/RUNTIME_BRINGUP_PLAN.md} / {@code RUNTIME_PREP_STATUS.md},
 * then delegates to {@link Mudclient#main}:
 *
 *   1. RSA: override {@link BitBuffer#RSA_MODULUS} + {@link FontBuilder#rsaPublicExponent}
 *      with the live OpenRSC server key (512-bit) BEFORE any login.
 *   2. Host/port: the deob's {@code setLoaderApplet} LOCAL_CIPHER branch is patched
 *      in-source to point {@code serverHost} at 127.0.0.1 and the login port at 43594.
 *   3. Content: {@code Mudclient.main} computes the content-host port as nodeId+7000,
 *      so passing nodeId "0" makes the client fetch content packs from
 *      http://127.0.0.1:7000/ (run the content host there) with doUpdate>20.
 *   4. Launch: {@code Mudclient.main(["0","live"])} sets the boot statics and calls
 *      startApplication, which runs CacheUpdater (fills _junkArray + BASE_URL),
 *      loadJagex (logo+fonts), then setLoaderApplet (title screen), then the game loop.
 */
public final class Boot {

    // --- ADD-2: live OpenRSC server RSA public key (see RUNTIME_PREP_STATUS.md §2) ---
    private static final BigInteger SERVER_RSA_MODULUS = new BigInteger(
        "8470727801174954902989859055344934434282083179399207801708507751976321325965228952554034824402302678046886295251980280826867546707365065713308009848924031");
    private static final BigInteger SERVER_RSA_EXPONENT = BigInteger.valueOf(65537);

    public static void main(String[] args) {
        // ADD-2: RSA override (both fields are public static, directly assignable).
        BitBuffer.RSA_MODULUS = SERVER_RSA_MODULUS;
        FontBuilder.rsaPublicExponent = SERVER_RSA_EXPONENT;

        System.out.println("[Boot] RSA modulus bitlen=" + SERVER_RSA_MODULUS.bitLength()
                + " exp=" + SERVER_RSA_EXPONENT);

        // ADD-1/3/4: nodeId "0" -> content host port 7000; mode "live".
        String[] delegate = (args != null && args.length >= 2)
                ? args
                : new String[] { "0", "live" };
        Mudclient.main(delegate);
    }

    private Boot() {}
}
