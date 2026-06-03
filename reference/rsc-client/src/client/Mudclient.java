package client;

// Best-effort imports. The decompiled support classes live under these packages;
// the original java.* / com.ms.* applet/J++ types will not resolve and that is expected.
import java.awt.Color;
import java.awt.Component;
import java.awt.Font;
import java.awt.Graphics;
import java.awt.Image;
import java.io.IOException;
import java.net.InetAddress;
import java.net.Socket;

import client.scene.*;     // Scene (lb), GameModel (ca), ImageLoader (pa), SurfaceSprite (ba) ...
import client.net.*;       // ClientStream (da), StreamBase (ib) ...
import client.world.*;     // World (k), WorldEntity (w) ...
import client.data.*;      // GameData, CacheUpdater (cb), DataStore (nb) ...
import client.ui.*;        // Panel (qa), FontBuilder (s), MessageList (wb) ...
import client.util.*;      // Utility (mb), Globals (l), StringCodec (u), Timer (p), ErrorHandler (i) ...
import client.shell.*;     // GameShell (e), GameFrame (qb), InputState (kb), LoaderThread (c) ...
import client.audio.*;     // SoundVoice (sa), SoundMixer (ra) ...
import client.nativeapi.*; // platform glue

/**
 * Mudclient — the main RuneScape Classic game class (obfuscated name {@code client}),
 * extending {@link GameShell} (obf {@code e}). Microsoft J++ build, RSC rev ~233-235.
 *
 * <p>This is the deobfuscated reassembly of the original ~55.6k-line obfuscated {@code client}
 * class: 123 declared methods (120 game methods + the constructor + the two {@code z} string-pool
 * decoders) and 484 fields. Field roles cross-reference {@code docs/NAMING.md} and
 * {@code docs/MUDCLIENT_SKELETON.md}; method bodies were re-audited against the clean Vineflower
 * base ({@code decompiled/normalized-clean/client.java}) and OpenRSC's Payload235 opcode maps.
 *
 * <p><b>Responsibilities.</b> Mudclient drives the entire client: it boots the applet/standalone
 * frame and loads every asset archive (models, 2D sprites, entity-animation frames, maps,
 * textures); performs the ISAAC-seed + RSA login handshake over a {@link client.net.ClientStream};
 * runs the per-tick game loop (input poll, mob/camera update, sub-system stepping); builds and
 * renders the 3D {@link client.Hh.Scene} from the {@link client.Ek.World}; renders all 2D
 * UI panels (login, game HUD, inventory, shop, bank, trade, duel, char-design, social dialogs,
 * chat) onto the {@link client.Hh.SurfaceSprite}; and serialises every outgoing action packet
 * while dispatching the ~42-opcode incoming server stream.
 *
 * <p><b>Deobfuscation notes.</b> The original obfuscation has been stripped from every method body:
 * <ul>
 *   <li>the opaque predicate {@code boolean bl = client.vh} (always {@code false}) and its dead
 *       branches;</li>
 *   <li>the per-method static {@code int} profiling-counter increments ({@code ++Pd; ++Gd; ...});</li>
 *   <li>the {@code try{ BODY }catch(RuntimeException e){ throw ErrorHandler.wrap(e, il[label]); }}
 *       wrappers (the {@code il[label]} signature strings are how each method's original obf name
 *       was recovered);</li>
 *   <li>anti-tamper dummy-parameter guards and junk modulo/mask expressions;</li>
 *   <li>the {@code ~x>~y} / {@code ~x==const} sign idioms (rewritten to ordinary comparisons).</li>
 * </ul>
 *
 * <p><b>Field naming.</b> The 484 fields split into two machine-generated bulk categories — ~119
 * per-method profiling counters and the opaque-predicate {@code vh} (collapsed into a single note
 * below, none are read by any kept method) — plus the genuinely-meaningful state declared in full
 * under the section banners that follow. Each declaration carries its obf name and role. NOTE on
 * {@code k}/{@code lb}: per {@code docs/NAMING.md}, {@code k=World} and {@code lb=Scene}; therefore
 * field {@code Hh} (type {@code k}) is the World and {@code Ek} (type {@code lb}) is the Scene —
 * the bodies use the readable aliases {@code Ek}/{@code Hh} accordingly.
 *
 * <p><b>Identifier note.</b> The pasted method bodies reference fields by a mix of readable aliases
 * (e.g. {@code Hh}, {@code Ek}, {@code li}, {@code Jh}, {@code wi},
 * {@code username}, {@code Qf} …) and still-obfuscated short names (e.g. {@code Ah},
 * {@code Bf}, {@code Cf}, {@code Di} …). Both forms are declared as fields here so every reference
 * resolves. A handful of GameShell-inherited fields are also referenced by their obf alias
 * ({@code mouseX}=mouseX, {@code mouseY}=mouseY, {@code originX}=originX, {@code originY}=originY, {@code mouseButtonDown}=
 * mouseButtonDown, {@code lastMouseButtonDown}=lastMouseButtonDown, {@code e}=inputTextCurrent, {@code x}=
 * inputPmCurrent, {@code inputTextFinal}=inputTextFinal, {@code inputPmFinal}=inputPmFinal, {@code interlace}=interlace,
 * {@code hasPainted}=hasPainted, {@code altDown}=altDown, {@code ctrlDown}=ctrlDown); these are NOT redeclared here
 * (they live in {@link GameShell}) so references to them resolve against the superclass.
 */
public class Mudclient extends GameShell {

    // ========================================================================
    // FIELDS — 484 total in the obf class. 397 named declarations follow (every real
    // field the kept method bodies reference: 258 by their original obf name, 53 readable
    // aliases, 82 fully-renamed game-state fields, plus Fj/Jk/ze/il). The remaining 122
    // per-method profiling counters + the opaque-predicate `vh` are obfuscation artefacts
    // that no kept method reads; they are collapsed into the single note just below. A few
    // GameShell-inherited fields referenced by obf alias (mouseX/mouseY/originX/originY/mouseButtonDown/lastMouseButtonDown/e/x/inputTextFinal/inputPmFinal/interlace/hasPainted/
    // altDown/ctrlDown) are intentionally NOT redeclared (they resolve against the superclass).
    // ========================================================================

    // ----- Opaque-predicate / special static fields -----
    // NOTE: the ~119 per-method static `int` profiling counters (Ic Pd Ci Xg lf hd Od Yg
    //   Hd Tj Bg Pi lj yh Yf Yj Ae gk Hg Th ej Dk Gc ic Kf Tf gj nd Qc bi hl ph wf Zi Ii
    //   pd wd Rf wc oi Gk Ec ih gf Ij qf Ei Zk mk Wb ie jh kk Bk jg Gh he Qj me aj dl tc
    //   al Gd ch Qh Lj ik rj ri Nk Xk pc Ck Tc Ie lg Ac bd Mf dh Nd Sd Mk Ye sc mj tk Sk
    //   Ch Fk Ad qi og ag Ik uh Uj qh ok cc Ri cd we Gg Nf kj yc Jc ef jf xc mi gg Li Oe
    //   yf Md fk Pe xf Ig) and the opaque predicate `public static boolean vh` (always
    //   false) are obfuscation artefacts, NOT game state. No kept method reads any of them,
    //   so the 122 counter scalars + vh are collapsed into this single note (the bytecode
    //   declares each as `static int <name>;` / `public static boolean vh;`).
    static int[] Fj;            // obf: Fj — keyState: per-keycode held/pressed bitfield
    public static int[] Jk;             // obf: Jk — loginScreenBg/wall-back-colour table (read cross-package by World)
    public static int[] equipSlotJk;    // obf: client.Jk (GameData equip table) — written cross-package by SocketFactory/StreamFactory
    public static long ze;              // obf: ze — tickMarker: scratch tick-hook timing marker (read cross-package by Timer)
    static final String[] il = new String[]{ // obf: il — STRINGS: XOR-encrypted string pool
        // RUNTIME FIX (Boot bring-up): the 683-entry initializer was previously omitted
        // (declared `new String[660]`, all null). Restored verbatim from the clean
        // decompile (client.java:493-1333); z(z(...)) -> xorDecode2(xorDecode1(...)).
      xorDecode2(xorDecode1("Ak\u001c\u0011\u0010V)=7V")),
      xorDecode2(xorDecode1("Ak\u001c\u0011\u0010V):7V")),
      xorDecode2(xorDecode1("Ak\u001c\u0011\u0010V)\"5V")),
      xorDecode2(xorDecode1("\u0005'\u001c\u0007^Lh\u0001T\u0011L'\u0013\u0006\u0017Gi\u0011T\u0012Kt\u0001")),
      xorDecode2(xorDecode1(
            "gu\u0007\u001b\f\u0018'\u0013\u0006\u0017Gi\u0011T\u001aKt\u0005\u0018\u001f['\u001b\u0015\u0013G'\u0016\u001c\u001fL`\u0010T\u000eCd\u001e\u0011\n\u0002u\u0010\u0017\u001bKq\u0010\u0010R\u0002e\u0000\u0000^Mk\u0011T\u0010Cj\u0010TY"
         )
      ),
      xorDecode2(xorDecode1("\u0005'\u001c\u0007^Lh\u0001T\u0011L'\u001c\u0013\u0010Mu\u0010T\u0012Kt\u0001")),
      xorDecode2(xorDecode1("Ak\u001c\u0011\u0010V)90V")),
      xorDecode2(xorDecode1(
            "gu\u0007\u001b\f\u0018'\u001c\u0013\u0010Mu\u0010T\u001aKt\u0005\u0018\u001f['\u001b\u0015\u0013G'\u0016\u001c\u001fL`\u0010T\u000eCd\u001e\u0011\n\u0002u\u0010\u0017\u001bKq\u0010\u0010R\u0002e\u0000\u0000^Mk\u0011T\u0010Cj\u0010TY"
         )
      ),
      xorDecode2(xorDecode1("\u0002o\u0014\u0007^Nh\u0012\u0013\u001bF'\u001a\u0001\n")),
      xorDecode2(xorDecode1("\u0002o\u0014\u0007^Nh\u0012\u0013\u001bF'\u001c\u001a")),
      xorDecode2(xorDecode1("bu\u0010\u0010>")),
      xorDecode2(xorDecode1("rh\u001c\u001a\n\u0002f\u0001T\u001f\u0002w\u0007\u0015\u0007GuU\u0012\u0011P'\u0014T\u001aGt\u0016\u0006\u0017Rs\u001c\u001b\u0010")),
      xorDecode2(xorDecode1("\u0018'")),
      xorDecode2(xorDecode1("Ak\u001c\u0011\u0010V)25V")),
      xorDecode2(xorDecode1("rh\u001c\u001a\n\u0002f\u0001T\u001f\u0002t\u0005\u0011\u0012N'\u0013\u001b\f\u0002fU\u0010\u001bQd\u0007\u001d\u000eVn\u001a\u001a")),
      xorDecode2(xorDecode1("bp\u001d\u001d>")),
      xorDecode2(xorDecode1("of\u0012\u001d\u001d")),
      xorDecode2(xorDecode1("Ru\u0014\r\u001bPh\u0013\u0012")),
      xorDecode2(xorDecode1("nb\u0003\u0011\u0012\u0002")),
      xorDecode2(xorDecode1("be\u0019\u0015>")),
      xorDecode2(xorDecode1("b~\u0010\u0018>")),
      xorDecode2(xorDecode1("ru\u0014\r\u001bPt")),
      xorDecode2(xorDecode1("Ru\u0014\r\u001bPh\u001b")),
      xorDecode2(xorDecode1(
            "{h\u0000\u0006^Ru\u0014\r\u001bP'\u0014\u0016\u0017Nn\u0001\r^KtU\u001a\u0011V'\u001d\u001d\u0019J'\u0010\u001a\u0011W`\u001dT\u0018MuU\u0000\u0016KtU\u0004\fC~\u0010\u0006"
         )
      ),
      xorDecode2(xorDecode1(
            "{h\u0000\u0006^Of\u0012\u001d\u001d\u0002f\u0017\u001d\u0012Ks\fT\u0017Q'\u001b\u001b\n\u0002o\u001c\u0013\u0016\u0002b\u001b\u001b\u000bEoU\u0012\u0011P'\u0001\u001c\u0017Q'\u0006\u0004\u001bNk"
         )
      ),
      xorDecode2(xorDecode1(
            "{h\u0000T\u001aMiR\u0000^Jf\u0003\u0011^Ck\u0019T\nJbU\u0006\u001bC`\u0010\u001a\nQ'\f\u001b\u000b\u0002i\u0010\u0011\u001a\u0002a\u001a\u0006^Vo\u001c\u0007^Qw\u0010\u0018\u0012"
         )
      ),
      xorDecode2(xorDecode1("fu\u0014\u001d\u0010\u0002u\u0014\u0000\u001b\u0018'")),
      xorDecode2(xorDecode1("b`\u0007\u0011>")),
      xorDecode2(xorDecode1(
            "{h\u0000T\u0016Cq\u0010T\fWiU\u001b\u000bV'\u001a\u0012^Ru\u0014\r\u001bP'\u0005\u001b\u0017Ls\u0006Z^pb\u0001\u0001\fL'\u0001\u001b^C'\u0016\u001c\u000bPd\u001dT\nM'\u0007\u0011\u001dJf\u0007\u0013\u001b"
         )
      ),
      xorDecode2(xorDecode1("Y)[Z\u0003")),
      xorDecode2(xorDecode1("Ak\u001c\u0011\u0010V)60V")),
      xorDecode2(xorDecode1("Lr\u0019\u0018")),
      xorDecode2(xorDecode1("af\u0017\u0016\u001fEb")),
      xorDecode2(xorDecode1("uf\u0019\u001f*M")),
      xorDecode2(xorDecode1("bk\u0007\u0011>")),
      xorDecode2(xorDecode1("b`\u0007E>")),
      xorDecode2(xorDecode1("Ak\u001c\u0011\u0010V)!5V")),
      xorDecode2(xorDecode1("b`\u0007F>")),
      xorDecode2(xorDecode1("wt\u0010T")),
      xorDecode2(xorDecode1("bh\u0007F>")),
      xorDecode2(xorDecode1("bh\u0007E>")),
      xorDecode2(xorDecode1("bd\f\u0015>")),
      xorDecode2(xorDecode1("\nk\u0010\u0002\u001bN*")),
      xorDecode2(xorDecode1("\u0002h\u001bT\rGk\u0013")),
      xorDecode2(xorDecode1("\u0002h\u001bT\u0019Ph\u0000\u001a\u001a")),
      xorDecode2(xorDecode1("vf\u0019\u001fSVh")),
      xorDecode2(xorDecode1("af\u0006\u0000^")),
      xorDecode2(xorDecode1("b`\u0007G>")),
      xorDecode2(xorDecode1("cs\u0001\u0015\u001dI")),
      xorDecode2(xorDecode1("bh\u0007G>")),
      xorDecode2(xorDecode1("\u0002h\u001b")),
      xorDecode2(xorDecode1("g\u007f\u0014\u0019\u0017Lb")),
      xorDecode2(xorDecode1("vf\u001e\u0011")),
      xorDecode2(xorDecode1("\u0002p\u001c\u0000\u0016")),
      xorDecode2(xorDecode1("uf\u0019\u001f^Jb\u0007\u0011")),
      xorDecode2(xorDecode1("ki\u0016\u001b\fPb\u0016\u0000^\u000f'%\u0018\u001bCt\u0010T\tCn\u0001ZP\f")),
      xorDecode2(xorDecode1("\u0002u\rN")),
      xorDecode2(xorDecode1("v6UY^")),
      xorDecode2(xorDecode1("\u0002u\fN")),
      xorDecode2(xorDecode1("v5UY^")),
      xorDecode2(xorDecode1("\u0002*U")),
      xorDecode2(xorDecode1("Ak\u001c\u0011\u0010V)11V")),
      xorDecode2(xorDecode1("\u0002i\u0000\u0019MN=")),
      xorDecode2(xorDecode1("Ak\u001c\u0011\u0010V)66V")),
      xorDecode2(xorDecode1("qh\u0007\u0006\u0007\u000e'\f\u001b\u000b\u0002d\u0014\u001aYV'\u0019\u001b\u0019Mr\u0001T\u001fV'\u0001\u001c\u001b\u0002j\u001a\u0019\u001bLs")),
      xorDecode2(xorDecode1("rk\u0010\u0015\rG'\u0010\u001a\nGuU\r\u0011WuU\u0001\rGu\u001b\u0015\u0013G'\u0014\u001a\u001a\u0002w\u0014\u0007\rUh\u0007\u0010")),
      xorDecode2(xorDecode1("Ak\u001c\u0011\u0010V)75V")),
      xorDecode2(xorDecode1("fu\u001a\u0004")),
      xorDecode2(xorDecode1("ub\u0014\u0006")),
      xorDecode2(xorDecode1("pb\u0018\u001b\bG")),
      xorDecode2(xorDecode1("Ak\u001c\u0011\u0010V)06V")),
      xorDecode2(xorDecode1("wt\u0010")),
      xorDecode2(xorDecode1("un\u0010\u0018\u001a")),
      xorDecode2(xorDecode1("Ak\u001c\u0011\u0010V)>\\")),
      xorDecode2(xorDecode1("Ak\u001c\u0011\u0010V);0V")),
      xorDecode2(xorDecode1("Ak\u001c\u0011\u0010V) 7V")),
      xorDecode2(xorDecode1("nh\u0006\u0000^Ah\u001b\u001a\u001bAs\u001c\u001b\u0010")),
      xorDecode2(xorDecode1("Ak\u001c\u0011\u0010V)67V")),
      xorDecode2(xorDecode1("Ak\u001c\u0011\u0010V)20V")),
      xorDecode2(xorDecode1("vh\u0005")),
      xorDecode2(xorDecode1("Ak\u001c\u0011\u0010V)?\\")),
      xorDecode2(xorDecode1("`f\u0016\u001f")),
      xorDecode2(xorDecode1("du\u001a\u001a\n")),
      xorDecode2(xorDecode1("ql\u001c\u001a")),
      xorDecode2(xorDecode1("jb\u0014\u0010")),
      xorDecode2(xorDecode1("jf\u001c\u0006")),
      xorDecode2(xorDecode1("ah\u0019\u001b\f")),
      xorDecode2(xorDecode1("rk\u0010\u0015\rG'\u0011\u0011\rK`\u001bT'Mr\u0007T=Jf\u0007\u0015\u001dVb\u0007")),
      xorDecode2(xorDecode1("v~\u0005\u0011")),
      xorDecode2(xorDecode1("`h\u0001\u0000\u0011O")),
      xorDecode2(xorDecode1("cd\u0016\u0011\u000eV")),
      xorDecode2(xorDecode1("eb\u001b\u0010\u001bP")),
      xorDecode2(xorDecode1("qn\u0011\u0011")),
      xorDecode2(xorDecode1("Ah\u0007\u001a\u001bPt[\u0010\u001fV")),
      xorDecode2(xorDecode1("Me\u001f\u0011\u001dVt")),
      xorDecode2(xorDecode1("Ki\u0003FPFf\u0001")),
      xorDecode2(xorDecode1("Ah\u0018\u0004\u001fQt[\u0010\u001fV")),
      xorDecode2(xorDecode1("Je\u0014\u0006L\fc\u0014\u0000")),
      xorDecode2(xorDecode1("@r\u0017\u0016\u0012G)\u0011\u0015\n")),
      xorDecode2(xorDecode1("Kd\u001a\u001aPFf\u0001")),
      xorDecode2(xorDecode1("@r\u0001\u0000\u0011Lt[\u0010\u001fV")),
      xorDecode2(xorDecode1("Qw\u0019\u0015\n\fc\u0014\u0000")),
      xorDecode2(xorDecode1("\fc\u0014\u0000")),
      xorDecode2(xorDecode1("Ki\u0011\u0011\u0006\fc\u0014\u0000")),
      xorDecode2(xorDecode1("Ak\u001c\u0011\u0010V)<5V")),
      xorDecode2(xorDecode1("Ru\u001a\u001e\u001bAs\u001c\u0018\u001b\fc\u0014\u0000")),
      xorDecode2(xorDecode1("Qd\u0007\u001b\u0012Ne\u0014\u0006PFf\u0001")),
      xorDecode2(xorDecode1("Cu\u0007\u001b\tQ)\u0011\u0015\n")),
      xorDecode2(xorDecode1("Au\u001a\u0003\u0010Q)\u0011\u0015\n")),
      xorDecode2(xorDecode1("Pr\u001b\u0011\rAf\u0005\u0011PFf\u0001")),
      xorDecode2(xorDecode1("\u0010cU\u0013\fCw\u001d\u001d\u001dQ")),
      xorDecode2(xorDecode1("Ki\u0003EPFf\u0001")),
      xorDecode2(xorDecode1("Je\u0014\u0006PFf\u0001")),
      xorDecode2(xorDecode1("Ak\u001c\u0011\u0010V)&5V")),
      xorDecode2(xorDecode1("Ak\u001c\u0011\u0010V)36V")),
      xorDecode2(xorDecode1("Ak\u001c\u0011\u0010V) 0V")),
      xorDecode2(xorDecode1("vu\u0014\u0010\u001b\u0002p\u001c\u0000\u0016")),
      xorDecode2(xorDecode1("Ak\u001c\u0011\u0010V);\\")),
      xorDecode2(xorDecode1("fr\u0010\u0018^Un\u0001\u001c")),
      xorDecode2(xorDecode1("dh\u0019\u0018\u0011U")),
      xorDecode2(xorDecode1("pb\u0005\u001b\fV'\u0014\u0016\u000bQb")),
      xorDecode2(xorDecode1("af\u001b\u0017\u001bN")),
      xorDecode2(xorDecode1("mL")),
      xorDecode2(xorDecode1("Ak\u001c\u0011\u0010V)%6V")),
      xorDecode2(xorDecode1("Ak\u001c\u0011\u0010V)\"\\")),
      xorDecode2(xorDecode1("Ak\u001c\u0011\u0010V) 5V")),
      xorDecode2(xorDecode1("ak\u001c\u0017\u0015\u0002o\u0010\u0006\u001b\u0002s\u001aT\u001dNh\u0006\u0011^Un\u001b\u0010\u0011U")),
      xorDecode2(xorDecode1("Ak\u001c\u0011\u0010V)#\\")),
      xorDecode2(xorDecode1("Ak\u001c\u0011\u0010V):5V")),
      xorDecode2(xorDecode1("`k\u001a\u0017\u0015\u0002c\u0000\u0011\u0012\u0002u\u0010\u0005\u000bGt\u0001\u0007D\u0002G\u0007\u0011\u001ab;\u001a\u0012\u0018\u001c")),
      xorDecode2(xorDecode1("Rk\u0010\u0015\rG'\u0006\u0011\u0012Gd\u0001TYCd\u0016\u001b\u000bLsU\u0019\u001fLf\u0012\u0011\u0013Gi\u0001S")),
      xorDecode2(xorDecode1("`k\u001a\u0017\u0015\u0002d\u001d\u0015\n\u0002j\u0010\u0007\rC`\u0010\u0007D\u0002G\u0007\u0011\u001ab;\u001a\u0012\u0018\u001c")),
      xorDecode2(xorDecode1("Du\u001a\u0019^Vo\u0010T\fWi\u0010\u0007\u001dCw\u0010T\u0018Ph\u001b\u0000^Ub\u0017\u0004\u001fEb")),
      xorDecode2(xorDecode1("Ck\u0019T\u000eGh\u0005\u0018\u001b\u0002i\u001a\u0000^MiU\r\u0011WuU\u0012\fKb\u001b\u0010\r\u0002k\u001c\u0007\n")),
      xorDecode2(xorDecode1("ql\u001c\u0004^Vo\u0010T\nWs\u001a\u0006\u0017Ck")),
      xorDecode2(xorDecode1("Du\u001a\u0019^Vo\u0010T\fWi\u0010\u0007\u001dCw\u0010Z\u001dMjU\u0012\fMi\u0001T\u000eC`\u0010")),
      xorDecode2(xorDecode1("af\u0018\u0011\fC'\u0014\u001a\u0019NbU\u0019\u0011FbUY^bu\u0010\u0010>of\u001b\u0001\u001fN")),
      xorDecode2(xorDecode1("Du\u001a\u0019^Vo\u0010T\u0012Ki\u001eT\u001cGk\u001a\u0003^Vo\u0010T\u0019Cj\u0010\u0003\u0017Lc\u001a\u0003")),
      xorDecode2(xorDecode1("ef\u0018\u0011^Mw\u0001\u001d\u0011LtUY^Ak\u001c\u0017\u0015\u0002s\u001aT\nM`\u0012\u0018\u001b")),
      xorDecode2(xorDecode1("ru\u001c\u0002\u001fA~U\u0007\u001bVs\u001c\u001a\u0019Q)U#\u0017NkU\u0016\u001b\u0002f\u0005\u0004\u0012Kb\u0011T\nM")),
      xorDecode2(xorDecode1("`k\u001a\u0017\u0015\u0002s\u0007\u0015\u001aG'\u0007\u0011\u000fWb\u0006\u0000\r\u0018'5\u0006\u001bFGI\u001b\u0018D9")),
      xorDecode2(xorDecode1("qh\u0000\u001a\u001a\u0002b\u0013\u0012\u001bAs\u0006TS\u0002G\u0012\u0006\u001bbh\u001b")),
      xorDecode2(xorDecode1("`k\u001a\u0017\u0015\u0002w\u0007\u001d\bCs\u0010T\u0013Gt\u0006\u0015\u0019GtOT>Pb\u00114BMa\u0013J")),
      xorDecode2(xorDecode1("Rf\u0006\u0007\tMu\u0011X^Pb\u0016\u001b\bGu\fT\u000fWb\u0006\u0000\u0017Mi\u0006X^Gs\u0016ZP")),
      xorDecode2(xorDecode1("oh\u0000\u0007\u001b\u0002e\u0000\u0000\nMi\u0006TS\u0002G\u0007\u0011\u001abH\u001b\u0011")),
      xorDecode2(xorDecode1("vhU\u0017\u0016Ci\u0012\u0011^[h\u0000\u0006^Ah\u001b\u0000\u001fAsU\u0010\u001bVf\u001c\u0018\r\u000e")),
      xorDecode2(xorDecode1("oh\u0000\u0007\u001b\u0002e\u0000\u0000\nMi\u0006TS\u0002G\u0012\u0006\u001bbS\u0002\u001b")),
      xorDecode2(xorDecode1("ck\u0002\u0015\u0007Q'\u0019\u001b\u0019Mr\u0001T\tJb\u001bT\u0007MrU\u0012\u0017Ln\u0006\u001c")),
      xorDecode2(xorDecode1("Ak\u001c\u0011\u0010V)77V")),
      xorDecode2(xorDecode1("ak\u001c\u0017\u0015\u0002o\u0010\u0006\u001b\u0002s\u001aT\u0012M`\u001a\u0001\n")),
      xorDecode2(xorDecode1("`k\u001a\u0017\u0015\u0002w\u0007\u001d\bCs\u0010T\u0013Gt\u0006\u0015\u0019GtOT>Eu\u00104BMiK")),
      xorDecode2(xorDecode1("af\u0018\u0011\fC'\u0014\u001a\u0019NbU\u0019\u0011FbUY^b`\u0007\u0011>cr\u0001\u001b")),
      xorDecode2(xorDecode1("`k\u001a\u0017\u0015\u0002s\u0007\u0015\u001aG'\u0007\u0011\u000fWb\u0006\u0000\r\u0018'5\u0013\fGGI\u001b\u0010\u001c")),
      xorDecode2(xorDecode1("`k\u001a\u0017\u0015\u0002d\u001d\u0015\n\u0002j\u0010\u0007\rC`\u0010\u0007D\u0002G\u0012\u0006\u001bb;\u001a\u001a@")),
      xorDecode2(xorDecode1("`k\u001a\u0017\u0015\u0002c\u0000\u0011\u0012\u0002u\u0010\u0005\u000bGt\u0001\u0007D\u0002G\u0012\u0006\u001bb;\u001a\u001a@")),
      xorDecode2(xorDecode1("qh\u0000\u001a\u001a\u0002b\u0013\u0012\u001bAs\u0006TS\u0002G\u0007\u0011\u001abh\u0013\u0012")),
      xorDecode2(xorDecode1("Ak\u001c\u0011\u0010V)07V")),
      xorDecode2(xorDecode1("Ak\u001c\u0011\u0010V)37V")),
      xorDecode2(xorDecode1("ma\u0013\u0011\f\u00026E")),
      xorDecode2(xorDecode1("\u0018'5\u0003\u0016KG")),
      xorDecode2(xorDecode1("Ms\u001d\u0011\f\u0002w\u0019\u0015\u0007Gu")),
      xorDecode2(xorDecode1("pb\u0018\u001b\bG'DD")),
      xorDecode2(xorDecode1("Ak\u001c\u0011\u0010V)2\\")),
      xorDecode2(xorDecode1("pb\u0018\u001b\bG'D")),
      xorDecode2(xorDecode1("{h\u0000\u0006^ma\u0013\u0011\f")),
      xorDecode2(xorDecode1("Jf\u0006T\u001fAd\u0010\u0004\nGc")),
      xorDecode2(xorDecode1("ma\u0013\u0011\f\u0002_")),
      xorDecode2(xorDecode1("mw\u0005\u001b\u0010Gi\u0001S\r\u0002H\u0013\u0012\u001bP")),
      xorDecode2(xorDecode1("ms\u001d\u0011\f\u0002w\u0019\u0015\u0007Gu")),
      xorDecode2(xorDecode1("ma\u0013\u0011\f\u00022")),
      xorDecode2(xorDecode1("pb\u0018\u001b\bG'-")),
      xorDecode2(xorDecode1("{h\u0000\u0006^ki\u0003\u0011\u0010Vh\u0007\r")),
      xorDecode2(xorDecode1("ma\u0013\u0011\f\u00026")),
      xorDecode2(xorDecode1("pb\u0018\u001b\bG'@")),
      xorDecode2(xorDecode1("ma\u0013\u0011\f\u0002F\u0019\u0018")),
      xorDecode2(xorDecode1("vu\u0014\u0010\u0017L`U\u0003\u0017VoOT")),
      xorDecode2(xorDecode1("uf\u001c\u0000\u0017L`U\u0012\u0011P")),
      xorDecode2(xorDecode1("pb\u0018\u001b\bG'4\u0018\u0012")),
      xorDecode2(xorDecode1("ob\u0006\u0007\u001fEb")),
      xorDecode2(xorDecode1("cc\u0011T\u0017Ei\u001a\u0006\u001b")),
      xorDecode2(xorDecode1("Ak\u001c\u0011\u0010V)$5V")),
      xorDecode2(xorDecode1("cc\u0011T\u0018Pn\u0010\u001a\u001a")),
      xorDecode2(xorDecode1("Lh\u0011\u0011\u0017F")),
      xorDecode2(xorDecode1("Ak\u001c\u0011\u0010V)\u001c\u001a\u0017V/\\")),
      xorDecode2(xorDecode1("Oh\u0011\u0011\tJf\u0001")),
      xorDecode2(xorDecode1("Oh\u0011\u0011\tJb\u0007\u0011")),
      xorDecode2(xorDecode1("Ak\u001c\u0011\u0010V)40V")),
      xorDecode2(xorDecode1("Ak\u001c\u0011\u0010V)4\\")),
      xorDecode2(xorDecode1("ak\u001c\u0017\u0015\u0002s\u001aT\u0013Gt\u0006\u0015\u0019G'")),
      xorDecode2(xorDecode1("bp\u001d\u001d>\u0002(UE^Oh\u0007\u0011^Mw\u0001\u001d\u0011L")),
      xorDecode2(xorDecode1("\u0002n\u0006T\u0011Da\u0019\u001d\u0010G")),
      xorDecode2(xorDecode1("\u0002j\u001a\u0006\u001b\u0002h\u0005\u0000\u0017Mi\u0006")),
      xorDecode2(xorDecode1("ao\u001a\u001b\rG'\u0014T\nCu\u0012\u0011\n")),
      xorDecode2(xorDecode1("\u0002h\u001bT")),
      xorDecode2(xorDecode1("k`\u001b\u001b\fKi\u0012T")),
      xorDecode2(xorDecode1("bp\u001d\u001d>\u0002(U")),
      xorDecode2(xorDecode1("ak\u001c\u0017\u0015\u0002s\u001aT\fGj\u001a\u0002\u001b\u0002")),
      xorDecode2(xorDecode1("Ak\u001c\u0011\u0010V)=\\")),
      xorDecode2(xorDecode1("\u0002/\u0013\u001b\fOb\u0007\u0018\u0007\u0002")),
      xorDecode2(xorDecode1("Ak\u001c\u0011\u0010V)27V")),
      xorDecode2(xorDecode1("Ak\u001c\u0011\u0010V)30V")),
      xorDecode2(xorDecode1("Ak\u001c\u0011\u0010V)'6V")),
      xorDecode2(xorDecode1("Ak\u001c\u0011\u0010V) \\")),
      xorDecode2(xorDecode1("Ak\u001c\u0011\u0010V)%\\")),
      xorDecode2(xorDecode1("rk\u0010\u0015\rG'\u0016\u001b\u0010Dn\u0007\u0019^[h\u0000\u0006^Vu\u0014\u0010\u001b\u0002p\u001c\u0000\u0016\u0002G\f\u0011\u0012b")),
      xorDecode2(xorDecode1("pb\u0018\u0011\u0013@b\u0007T\nJf\u0001T\u0010MsU\u0015\u0012N'\u0005\u0018\u001f[b\u0007\u0007^Cu\u0010T\nPr\u0006\u0000\tMu\u0001\u001c\u0007")),
      xorDecode2(xorDecode1("cu\u0010T\u0007MrU\u0007\u000bPbU\r\u0011W'\u0002\u0015\u0010V'\u0001\u001b^FhU\u0000\u0016KtJ")),
      xorDecode2(xorDecode1(
            "vo\u0010\u0006\u001b\u0002n\u0006T0m'\"5'\u0002s\u001aT\fGq\u0010\u0006\rG'\u0014T\nPf\u0011\u0011^KaU\r\u0011W'\u0016\u001c\u001fL`\u0010T\u0007Mr\u0007T\u0013Ki\u0011Z"
         )
      ),
      xorDecode2(xorDecode1("Ak\u001c\u0011\u0010V)\"0V")),
      xorDecode2(xorDecode1("kiU\u0006\u001bVr\u0007\u001a^[h\u0000T\tKk\u0019T\fGd\u0010\u001d\bG=")),
      xorDecode2(xorDecode1("{h\u0000T\u001fPbU\u0015\u001cMr\u0001T\nM'\u0012\u001d\bG=")),
      xorDecode2(xorDecode1("\u0002\u007fU")),
      xorDecode2(xorDecode1("uf\u001c\u0000\u0017L`U\u0012\u0011P'\u001a\u0000\u0016GuU\u0004\u0012C~\u0010\u0006P\f)")),
      xorDecode2(xorDecode1("lh\u0001\u001c\u0017L`T")),
      xorDecode2(xorDecode1("Ak\u001c\u0011\u0010V)%5V")),
      xorDecode2(xorDecode1(
            "vo\u001c\u0007^Me\u001f\u0011\u001dV'\u0016\u0015\u0010Lh\u0001T\u001cG'\u0001\u0006\u001fFb\u0011T\tKs\u001dT\u0011Vo\u0010\u0006^Rk\u0014\r\u001bPt"
         )
      ),
      xorDecode2(xorDecode1("Ak\u001c\u0011\u0010V):0V")),
      xorDecode2(xorDecode1(
            "vo\u001c\u0007^Me\u001f\u0011\u001dV'\u0016\u0015\u0010Lh\u0001T\u001cG'\u0014\u0010\u001aGcU\u0000\u0011\u0002fU\u0010\u000bGkU\u001b\u0018Db\u0007"
         )
      ),
      xorDecode2(xorDecode1("Ak\u001c\u0011\u0010V)6\\")),
      xorDecode2(xorDecode1("Ak\u001c\u0011\u0010V)<7V")),
      xorDecode2(xorDecode1("Ak\u001c\u0011\u0010V)>6V")),
      xorDecode2(xorDecode1("Ak\u001c\u0011\u0010V)<\\")),
      xorDecode2(xorDecode1("Ak\u001c\u0011\u0010V)\"6V")),
      xorDecode2(xorDecode1("Ak\u001c\u0011\u0010V)&\\")),
      xorDecode2(xorDecode1("Ak\u001c\u0011\u0010V)61V")),
      xorDecode2(xorDecode1("ah\u001b\u0012\u0017Er\u0007\u0015\nKh\u001b")),
      xorDecode2(xorDecode1("Ak\u001c\u0011\u0010V)97V")),
      xorDecode2(xorDecode1("Ak\u001c\u0011\u0010V)85V")),
      xorDecode2(xorDecode1("Ak\u001c\u0011\u0010V)70V")),
      xorDecode2(xorDecode1("Ak\u001c\u0011\u0010V)17V")),
      xorDecode2(xorDecode1("{h\u0000T\u0010Gb\u0011T\u001f\u0002j\u0010\u0019\u001cGu\u0006T\u001fAd\u001a\u0001\u0010V'\u0001\u001b^Wt\u0010T\nJn\u0006T\rGu\u0003\u0011\f")),
      xorDecode2(xorDecode1("ml")),
      xorDecode2(xorDecode1("ak\u001c\u0017\u0015\u0002o\u0010\u0006\u001b\u0002s\u001aT\u0012M`\u001c\u001a")),
      xorDecode2(xorDecode1(
            "{h\u0000T\u0010Gb\u0011T\u001f\u0002q\u0010\u0000\u001bPf\u001bT=Nf\u0006\u0007\u0017A'\u0018\u0011\u0013@b\u0007\u0007^Cd\u0016\u001b\u000bLsU\u0000\u0011\u0002r\u0006\u0011^Vo\u001c\u0007^Qb\u0007\u0002\u001bP"
         )
      ),
      xorDecode2(xorDecode1("rf\u0006\u0007\tMu\u0011N")),
      xorDecode2(xorDecode1("wt\u0010\u0006\u0010Cj\u0010N")),
      xorDecode2(xorDecode1("Ak\u001c\u0011\u0010V)7\\")),
      xorDecode2(xorDecode1("ub\u0019\u0017\u0011ObU\u0000\u0011\u0002U\u0000\u001a\u001bqd\u0014\u0004\u001b\u0002D\u0019\u0015\rQn\u0016")),
      xorDecode2(xorDecode1(
            "{h\u0000T\u0010Gb\u0011T\u001f\u0002q\u0010\u0000\u001bPf\u001bT=Nf\u0006\u0007\u0017A'\u0014\u0017\u001dMr\u001b\u0000^VhU\u0001\rG'\u0001\u001c\u0017Q'\u0006\u0011\fTb\u0007"
         )
      ),
      xorDecode2(xorDecode1("Ak\u001c\u0011\u0010V)\"7V")),
      xorDecode2(xorDecode1("vb\r\u0000\u000bPb\u0006")),
      xorDecode2(xorDecode1("Ak\u001c\u0011\u0010V) 6V")),
      xorDecode2(xorDecode1("Ak\u001c\u0011\u0010V)8\\")),
      xorDecode2(xorDecode1("Ak\u001c\u0011\u0010V)47V")),
      xorDecode2(xorDecode1("Ak\u001c\u0011\u0010V)96V")),
      xorDecode2(xorDecode1("Ak\u001c\u0011\u0010V)71V")),
      xorDecode2(xorDecode1("gi\u0001\u0011\f\u0002i\u0014\u0019\u001b\u0002s\u001aT\u001fFcU\u0000\u0011\u0002a\u0007\u001d\u001bLc\u0006T\u0012Kt\u0001")),
      xorDecode2(xorDecode1("Ak\u001c\u0011\u0010V)26V")),
      xorDecode2(xorDecode1("gi\u0001\u0011\f\u0002i\u0014\u0019\u001b\u0002s\u001aT\u001fFcU\u0000\u0011\u0002n\u0012\u001a\u0011PbU\u0018\u0017Qs")),
      xorDecode2(xorDecode1("gi\u0001\u0011\f\u0002j\u0010\u0007\rC`\u0010T\nM'\u0006\u0011\u0010F'\u0001\u001b^")),
      xorDecode2(xorDecode1("Ak\u001c\u0011\u0010V)87V")),
      xorDecode2(xorDecode1("rk\u0010\u0015\rG'\u0007\u0011\u0013Mq\u0010T")),
      xorDecode2(xorDecode1("\u0002n\u0006T\u001fNu\u0010\u0015\u001a['\u001a\u001a^[h\u0000\u0006^K`\u001b\u001b\fG'\u0019\u001d\rV")),
      xorDecode2(xorDecode1("{h\u0000T\u001dCiR\u0000^Cc\u0011T\u0007Mr\u0007\u0007\u001bNaU\u0000\u0011\u0002~\u001a\u0001\f\u0002n\u0012\u001a\u0011PbU\u0018\u0017Qs")),
      xorDecode2(xorDecode1("k`\u001b\u001b\fG'\u0019\u001d\rV'\u0013\u0001\u0012N")),
      xorDecode2(xorDecode1("\u0002a\u0007\u001b\u0013\u0002~\u001a\u0001\f\u0002a\u0007\u001d\u001bLc\u0006T\u0012Kt\u0001T\u0018Ku\u0006\u0000")),
      xorDecode2(xorDecode1("Ak\u001c\u0011\u0010V)#6V")),
      xorDecode2(xorDecode1("Ak\u001c\u0011\u0010V)=6V")),
      xorDecode2(xorDecode1("k`\u001b\u001b\fG")),
      xorDecode2(xorDecode1("ak\u001c\u0017\u0015\u0002o\u0010\u0006\u001b\u0002s\u001aT\u001fFcU\u0015^Du\u001c\u0011\u0010F")),
      xorDecode2(xorDecode1("du\u001c\u0011\u0010Ft")),
      xorDecode2(xorDecode1("\f)[")),
      xorDecode2(xorDecode1("\\3FM\u0000bp\u001d\u001d>pb\u0018\u001b\bG'UT^\u0002'UT^uP\"#)uP\"#)")),
      xorDecode2(xorDecode1("`k\u001a\u0017\u0015Ki\u0012T\u0013Gt\u0006\u0015\u0019GtU\u0012\fMjO")),
      xorDecode2(xorDecode1("Ak\u001c\u0011\u0010V)#0V")),
      xorDecode2(xorDecode1("ak\u001c\u0017\u0015\u0002o\u0010\u0006\u001b\u0002s\u001aT\u001fFcU\u0015^Lf\u0018\u0011")),
      xorDecode2(xorDecode1("ak\u001c\u0017\u0015\u0002fU\u001a\u001fObU\u0000\u0011\u0002t\u0010\u001a\u001a\u0002fU\u0019\u001bQt\u0014\u0013\u001b")),
      xorDecode2(xorDecode1("Ak\u001c\u0011\u0010V);7V")),
      xorDecode2(xorDecode1("ru\u001c\u0002\u001fVbU\u001c\u0017Qs\u001a\u0006\u0007")),
      xorDecode2(xorDecode1("ck\u0019T\u0013Gt\u0006\u0015\u0019Gt")),
      xorDecode2(xorDecode1("Ak\u001c\u0011\u0010V)$6V")),
      xorDecode2(xorDecode1("sr\u0010\u0007\n\u0002o\u001c\u0007\nMu\f")),
      xorDecode2(xorDecode1("ao\u0014\u0000^Jn\u0006\u0000\u0011P~")),
      xorDecode2(xorDecode1("Dn\u0007\u0011\u001f\u0011")),
      xorDecode2(xorDecode1("nh\u0014\u0010\u0017L`UG\u001a\u0002j\u001a\u0010\u001bNt")),
      xorDecode2(xorDecode1("Ql\u0000\u0018\u0012Vh\u0007\u0017\u0016C4")),
      xorDecode2(xorDecode1("Nn\u0012\u001c\nLn\u001b\u0013L")),
      xorDecode2(xorDecode1("Dn\u0007\u0011\u001f\u0010")),
      xorDecode2(xorDecode1("Ql\u0000\u0018\u0012Vh\u0007\u0017\u0016C3")),
      xorDecode2(xorDecode1("\fh\u0017F")),
      xorDecode2(xorDecode1("Dn\u0007\u0011\rRb\u0019\u0018M")),
      xorDecode2(xorDecode1("Qw\u0010\u0018\u0012Ao\u0014\u0006\u0019G4")),
      xorDecode2(xorDecode1("Dn\u0007\u0011\rRb\u0019\u0018L")),
      xorDecode2(xorDecode1("Dn\u0007\u0011\u000eNf\u0016\u0011\u001f\u0010")),
      xorDecode2(xorDecode1("Vh\u0007\u0017\u0016C4")),
      xorDecode2(xorDecode1("\u0011cU\u0019\u0011Fb\u0019\u0007")),
      xorDecode2(xorDecode1("Ak\u001c\u0011\u0010V)15V")),
      xorDecode2(xorDecode1("Vh\u0007\u0017\u0016C5")),
      xorDecode2(xorDecode1("Ak\u0014\u0003\rRb\u0019\u0018K")),
      xorDecode2(xorDecode1("Nn\u0012\u001c\nLn\u001b\u0013M")),
      xorDecode2(xorDecode1("\fh\u0017G")),
      xorDecode2(xorDecode1("Qw\u0010\u0018\u0012Ao\u0014\u0006\u0019G5")),
      xorDecode2(xorDecode1("Ak\u0014\u0003\rRb\u0019\u0018J")),
      xorDecode2(xorDecode1("Ak\u0014\u0003\rRb\u0019\u0018M")),
      xorDecode2(xorDecode1("Ql\u0000\u0018\u0012Vh\u0007\u0017\u0016C5")),
      xorDecode2(xorDecode1("Vh\u0007\u0017\u0016C3")),
      xorDecode2(xorDecode1("En\u0014\u001a\nAu\f\u0007\nCk")),
      xorDecode2(xorDecode1("\f)Z\u0017\u0011Ls\u0010\u001a\n\rt\u0007\u0017QOh\u0011\u0011\u0012Q(")),
      xorDecode2(xorDecode1("Dn\u0007\u0011\u000eNf\u0016\u0011\u001f\u0011")),
      xorDecode2(xorDecode1("Ak\u0014\u0003\rRb\u0019\u0018L")),
      xorDecode2(xorDecode1("Un\u0019\u0010\u001bPi\u0010\u0007\r\f'!\u001c\u0017Q'\u0014T\bGu\fT\u001aCi\u0012\u0011\fMr\u0006T\u001fPb\u0014T\tJb\u0007\u0011")),
      xorDecode2(xorDecode1(
            "@b\u0016\u001b\u0013GtYT\u001cWsU\u0000\u0016G'\u0018\u001b\fG'\u0001\u0006\u001bCt\u0000\u0006\u001b\u0002~\u001a\u0001^Un\u0019\u0018^Dn\u001b\u0010P"
         )
      ),
      xorDecode2(xorDecode1(
            "kiU\u0000\u0016G'\u0002\u001d\u0012Fb\u0007\u001a\u001bQtU\u0015\u0010\u0002n\u001b\u0010\u0017Af\u0001\u001b\f\u0002f\u0001T\nJbU\u0016\u0011Vs\u001a\u0019SPn\u0012\u001c\n"
         )
      ),
      xorDecode2(xorDecode1(
            "MaU\u0000\u0016G'\u0006\u0017\fGb\u001bT\tKk\u0019T\rJh\u0002T\nJbU\u0017\u000bPu\u0010\u001a\n\u0002k\u0010\u0002\u001bN'\u001a\u0012^Ff\u001b\u0013\u001bP"
         )
      ),
      xorDecode2(xorDecode1("Ak\u001c\u0011\u0010V)35V")),
      xorDecode2(xorDecode1("kaU\r\u0011W'\u0012\u001b^Or\u0016\u001c^Dr\u0007\u0000\u0016GuU\u001a\u0011Ps\u001dT\u0007MrU\u0003\u0017NkU\u0011\u0010Vb\u0007T\nJb")),
      xorDecode2(xorDecode1("Ms\u001d\u0011\f\u0002w\u0019\u0015\u0007Gu\u0006T\u001dCiU\u0015\nVf\u0016\u001f^[h\u0000U")),
      xorDecode2(xorDecode1("uf\u0007\u001a\u0017L`TT.Ph\u0016\u0011\u001bF'\u0002\u001d\nJ'\u0016\u0015\u000bVn\u001a\u001a")),
      xorDecode2(xorDecode1("vo\u0010T\u0018Wu\u0001\u001c\u001bP'\u001b\u001b\fVoU\r\u0011W'\u0012\u001b^Vo\u0010T\u0013Mu\u0010T\u001aCi\u0012\u0011\fMr\u0006T\u0017V")),
      xorDecode2(xorDecode1("Ak\u001c\u0011\u0010V)10V")),
      xorDecode2(xorDecode1("Ak\u001c\u0011\u0010V)65V")),
      xorDecode2(xorDecode1("Ak\u001c\u0011\u0010V)9\\")),
      xorDecode2(xorDecode1("Nn\u0003\u0011")),
      xorDecode2(xorDecode1("Ak\u001c\u0011\u0010V)\u0018\u0015\u0017L/")),
      xorDecode2(xorDecode1("pr\u001b\u0011-Af\u0005\u0011^ak\u0014\u0007\rKd")),
      xorDecode2(xorDecode1("Tb\u0001\u0011\fCi\u0006")),
      xorDecode2(xorDecode1("Ob\u0018\u0016\u001bPt")),
      xorDecode2(xorDecode1("Pd")),
      xorDecode2(xorDecode1("Un\u0005")),
      xorDecode2(xorDecode1("Ak\u0014\u0007\rKd")),
      xorDecode2(xorDecode1("Ak\u001c\u0011\u0010V)05V")),
      xorDecode2(xorDecode1("C)\u0011\u0015\n")),
      xorDecode2(xorDecode1("nh\u0014\u0010\u001bF=U")),
      xorDecode2(xorDecode1("D)\u0011\u0015\n")),
      xorDecode2(xorDecode1("Rb\u001a\u0004\u0012G'\u0014\u001a\u001a\u0002j\u001a\u001a\rVb\u0007\u0007")),
      xorDecode2(xorDecode1("Ak\u001c\u0011\u0010V)>0V")),
      xorDecode2(xorDecode1("Ob\u0018\u0016\u001bP'\u0012\u0006\u001fRo\u001c\u0017\r")),
      xorDecode2(xorDecode1("\u0002a\u0007\u0015\u0013GtU\u001b\u0018\u0002f\u001b\u001d\u0013Cs\u001c\u001b\u0010")),
      xorDecode2(xorDecode1("Nh\u0016\u0015\u0012\fu\u0000\u001a\u001bQd\u0014\u0004\u001b\fd\u001a\u0019")),
      xorDecode2(xorDecode1("\fu\u0000\u001a\u001bQd\u0014\u0004\u001b\fd\u001a\u0019")),
      xorDecode2(xorDecode1("qs\u0014\u0006\nKi\u0012T\u0019Cj\u0010ZP\f")),
      xorDecode2(xorDecode1("Qb\u0007\u0002\u001bPs\f\u0004\u001b")),
      xorDecode2(xorDecode1("Pb\u0013\u0011\fKc")),
      xorDecode2(xorDecode1("Pr\u001b\u0011\rAf\u0005\u0011PAh\u0018")),
      xorDecode2(xorDecode1("Ak\u001c\u0011\u0010V)>7V")),
      xorDecode2(xorDecode1("ao\u001a\u001b\rG'\u001a\u0004\nKh\u001b")),
      xorDecode2(xorDecode1("y_(")),
      xorDecode2(xorDecode1("\u0002T\u0000\u0013\u0019Gt\u0001T\u0013Ws\u0010")),
      xorDecode2(xorDecode1("Ak\u001c\u0011\u0010V)3\\")),
      xorDecode2(xorDecode1("y'(")),
      xorDecode2(xorDecode1(
            "gi\u0001\u0011\f\u0002s\u001d\u0011^Lf\u0018\u0011^MaU\u0000\u0016G'\u0005\u0018\u001f[b\u0007T\u0007MrU\u0003\u0017QoU\u0000\u0011\u0002u\u0010\u0004\u0011PsO"
         )
      ),
      xorDecode2(xorDecode1("\u0002J\u0000\u0000\u001b\u0002w\u0019\u0015\u0007Gu")),
      xorDecode2(xorDecode1("Ak\u001c\u0011\u0010V);6V")),
      xorDecode2(xorDecode1("Ak\u001c\u0011\u0010V)95V")),
      xorDecode2(xorDecode1("wi\u0014\u0016\u0012G'\u0001\u001b^Ki\u001c\u0000^Qh\u0000\u001a\u001aQ=")),
      xorDecode2(xorDecode1("qh\u0000\u001a\u001a\u0002b\u0013\u0012\u001bAs\u0006")),
      xorDecode2(xorDecode1("\u0002t\u001e\u001d\u0012N")),
      xorDecode2(xorDecode1("ql\u001c\u0018\u0012\u0002s\u001a\u0000\u001fN=U")),
      xorDecode2(xorDecode1("gv\u0000\u001d\u000eOb\u001b\u0000^qs\u0014\u0000\u000bQ")),
      xorDecode2(xorDecode1("Ak\u001c\u0011\u0010V)$7V")),
      xorDecode2(xorDecode1("\u0018G\f\u0011\u0012b")),
      xorDecode2(xorDecode1("sr\u0010\u0007\nQ")),
      xorDecode2(xorDecode1("mq\u0010\u0006\u001fNkU\u0018\u001bTb\u0019\u0007")),
      xorDecode2(xorDecode1("bp\u001d\u001d>sr\u0010\u0007\n\u000fk\u001c\u0007\n\u0002/\u0012\u0006\u001bGiH\u0017\u0011Ow\u0019\u0011\nGc\\")),
      xorDecode2(xorDecode1("ah\u0018\u0016\u001fV'\u0019\u0011\bGkOT")),
      xorDecode2(xorDecode1("ql\u001c\u0018\u0012Q")),
      xorDecode2(xorDecode1("qs\u0014\u0000\r")),
      xorDecode2(xorDecode1("vh\u0001\u0015\u0012\u0002\u007f\u0005N^")),
      xorDecode2(xorDecode1("sr\u0010\u0007\n\u0002W\u001a\u001d\u0010VtO4\u0007Gk5")),
      xorDecode2(xorDecode1("lb\r\u0000^Nb\u0003\u0011\u0012\u0002f\u0001N^")),
      xorDecode2(xorDecode1("df\u0001\u001d\u0019WbOT>[b\u00194")),
      xorDecode2(xorDecode1("Dn\u0007\u0011\u000eNf\u0016\u0011\u001f")),
      xorDecode2(xorDecode1("nb\u0003\u0011\u0012\u0018'")),
      xorDecode2(xorDecode1("Ak\u001c\u0011\u0010V)46V")),
      xorDecode2(xorDecode1("Nn\u0012\u001c\nLn\u001b\u0013")),
      xorDecode2(xorDecode1("\u00187")),
      xorDecode2(xorDecode1("{h\u0000T\u001fPbU\u0007\u0012Gb\u0005\u001d\u0010E")),
      xorDecode2(xorDecode1("uo\u0010\u001a^[h\u0000T\tCi\u0001T\nM'\u0002\u0015\u0015G'\u0000\u0004^Hr\u0006\u0000^Wt\u0010T\u0007Mr\u0007")),
      xorDecode2(xorDecode1("Vh\u0007\u0017\u0016C")),
      xorDecode2(xorDecode1(
            "b~\u0010\u0018>Ak\u001c\u0017\u0015\u0002o\u0010\u0006\u001bbp\u001d\u001d>\u0002s\u001aT\u0019GsU\u0015^Fn\u0013\u0012\u001bPb\u001b\u0000^Mi\u0010"
         )
      ),
      xorDecode2(xorDecode1("kaU\r\u0011W'\u0016\u0015\u0010\u0005sU\u0006\u001bCcU\u0000\u0016G'\u0002\u001b\fF")),
      xorDecode2(xorDecode1("moU\u0010\u001bCuTT'MrU\u0015\fG'\u0011\u0011\u001fF)[Z")),
      xorDecode2(xorDecode1("Ak\u0014\u0003\rRb\u0019\u0018")),
      xorDecode2(xorDecode1("df\u0001\u001d\u0019WbOT")),
      xorDecode2(xorDecode1("Ib\f\u0016\u0011Cu\u0011T\nM'\u0001\r\u000eG'\u0001\u001c\u001b\u0002p\u001a\u0006\u001a\u0002n\u001bT\nJbU\u0016\u0011Z'\u0017\u0011\u0012Mp")),
      xorDecode2(xorDecode1("Dn\u0007\u0011\rRb\u0019\u0018")),
      xorDecode2(xorDecode1("Dn\u0007\u0011\u001f")),
      xorDecode2(xorDecode1("un\u0019\u0010\u001bPi\u0010\u0007\r")),
      xorDecode2(xorDecode1("x]/")),
      xorDecode2(xorDecode1("Qw\u0010\u0018\u0012Ao\u0014\u0006\u0019G")),
      xorDecode2(xorDecode1("q~\u0006\u0000\u001bO'\u0000\u0004\u001aCs\u0010T\u0017L=U")),
      xorDecode2(xorDecode1("Ql\u0000\u0018\u0012Vh\u0007\u0017\u0016C")),
      xorDecode2(xorDecode1("Ak\u001c\u0011\u0010V)'5V")),
      xorDecode2(xorDecode1("\u0002a\u0007\u001b\u0013\u0002~\u001a\u0001\f\u0002n\u0012\u001a\u0011PbU\u0018\u0017QsU\u0012\u0017Pt\u0001Z")),
      xorDecode2(xorDecode1("du\u001c\u0011\u0010F'\u0019\u001d\rV'\u001c\u0007^Dr\u0019\u0018")),
      xorDecode2(xorDecode1(
            "{h\u0000T\u001dCiR\u0000^Cc\u0011T\u0007Mr\u0007\u0007\u001bNaU\u0000\u0011\u0002~\u001a\u0001\f\u0002h\u0002\u001a^Du\u001c\u0011\u0010F'\u0019\u001d\rV)"
         )
      ),
      xorDecode2(xorDecode1("\u0002n\u0006T\u001fNu\u0010\u0015\u001a['\u001a\u001a^[h\u0000\u0006^Du\u001c\u0011\u0010F'\u0019\u001d\rV)")),
      xorDecode2(xorDecode1("Ak\u001c\u0011\u0010V)%7V")),
      xorDecode2(xorDecode1("Ak\u001c\u0011\u0010V)$0V")),
      xorDecode2(xorDecode1("Ak\u001c\u0011\u0010V)41V")),
      xorDecode2(xorDecode1("Ak\u001c\u0011\u0010V)76V")),
      xorDecode2(xorDecode1("ak\u001c\u0017\u0015\u0002o\u0010\u0006\u001b\u0002s\u001aT\u001dCi\u0016\u0011\u0012")),
      xorDecode2(xorDecode1("gi\u0016\u001b\u000bPf\u0012\u001d\u0010E'\u0007\u0001\u0012G*\u0017\u0006\u001bCl\u001c\u001a\u0019")),
      xorDecode2(xorDecode1("ct\u001e\u001d\u0010E'\u0013\u001b\f\u0002h\u0007T\u000ePh\u0003\u001d\u001aKi\u0012")),
      xorDecode2(xorDecode1("Ak\u001c\u0011\u0010V)<0V")),
      xorDecode2(xorDecode1("kaU\r\u0011W'\u0018\u001d\rWt\u0010T\nJn\u0006T\u0018Mu\u0018X^[h\u0000T\tKk\u0019T\u001cG'\u0017\u0015\u0010Lb\u0011Z")),
      xorDecode2(xorDecode1("Nf\u001b\u0013\u000bC`\u0010")),
      xorDecode2(xorDecode1("ma\u0013\u0011\u0010Qn\u0003\u0011^Cd\u0016\u001b\u000bLsU\u001a\u001fOb")),
      xorDecode2(xorDecode1("g\u007f\u0005\u0018\u0011Ks\u001c\u001a\u0019\u0002fU\u0016\u000bE")),
      xorDecode2(xorDecode1("qh\u0019\u001d\u001dKs\u0014\u0000\u0017Mi")),
      xorDecode2(xorDecode1("Ah\u001b\u0000\u001fAsU\u001d\u0010Dh\u0007\u0019\u001fVn\u001a\u001a")),
      xorDecode2(xorDecode1("qb\u0007\u001d\u0011Wt\u0019\r^Ma\u0013\u0011\u0010Qn\u0003\u0011")),
      xorDecode2(xorDecode1("cc\u0003\u0011\fVn\u0006\u001d\u0010E'\u0002\u0011\u001cQn\u0001\u0011\r")),
      xorDecode2(xorDecode1("qb\u0016\u0001\fKs\f")),
      xorDecode2(xorDecode1("of\u0016\u0006\u0011Ki\u0012T\u0011P'\u0000\u0007\u001b\u0002h\u0013T\u001cMs\u0006")),
      xorDecode2(xorDecode1("qd\u0014\u0019\u0013Ki\u0012")),
      xorDecode2(xorDecode1(
            "ak\u001c\u0017\u0015\u0002h\u001bT\nJbU\u0019\u0011QsU\u0007\u000bKs\u0014\u0016\u0012G'\u001a\u0004\nKh\u001bT\u0018Ph\u0018T\nJbU&\u000bNb\u0006T\u0011D''\u0001\u0010GT\u0016\u0015\u000eG)"
         )
      ),
      xorDecode2(xorDecode1(
            "vo\u001c\u0007^Un\u0019\u0018^Qb\u001b\u0010^C'\u0007\u0011\u000eMu\u0001T\nM'\u001a\u0001\f\u0002W\u0019\u0015\u0007GuU'\u000bRw\u001a\u0006\n\u0002s\u0010\u0015\u0013\u0002a\u001a\u0006^Ki\u0003\u0011\rVn\u0012\u0015\nKh\u001bZ"
         )
      ),
      xorDecode2(xorDecode1(
            "vo\u001c\u0007^Dh\u0007\u0019^KtU\u0012\u0011P'\u0007\u0011\u000eMu\u0001\u001d\u0010E'\u0005\u0018\u001f[b\u0007\u0007^Uo\u001aT\u001fPbU\u0016\fGf\u001e\u001d\u0010E'\u001a\u0001\f\u0002u\u0000\u0018\u001bQ"
         )
      ),
      xorDecode2(xorDecode1("qs\u0014\u0012\u0018\u0002n\u0018\u0004\u001bPt\u001a\u001a\u001fVn\u001a\u001a")),
      xorDecode2(xorDecode1("jh\u001b\u001b\u000bP")),
      xorDecode2(xorDecode1(
            "wt\u001c\u001a\u0019\u0002n\u0001T\rGi\u0011\u0007^C'\u0006\u001a\u001fRt\u001d\u001b\n\u0002h\u0013T\nJbU\u0018\u001fQsUBN\u0002t\u0010\u0017\u0011Lc\u0006T\u0011D'\u0014\u0017\nKq\u001c\u0000\u0007\u0002s\u001aT\u000bQ"
         )
      ),
      xorDecode2(xorDecode1("`u\u0010\u0015\u0015Ki\u0012T\fGf\u0019Y\tMu\u0019\u0010^Nf\u0002\u0007")),
      xorDecode2(xorDecode1("Qb\u0019\u0018\u0017L`U\u0015\u0010\u0002f\u0016\u0017\u0011Wi\u0001")),
      xorDecode2(xorDecode1("`r\f\u001d\u0010E'\u001a\u0006")),
      xorDecode2(xorDecode1("pb\u0006\u0004\u001bAs")),
      xorDecode2(xorDecode1("fn\u0006\u0006\u000bRs\u001c\u0002\u001b\u0002e\u0010\u001c\u001fTn\u001a\u0001\f")),
      xorDecode2(xorDecode1("pb\u0014\u0018SNn\u0013\u0011^Vo\u0007\u0011\u001fVt")),
      xorDecode2(xorDecode1("Ak\u001c\u0011\u0010V):\\")),
      xorDecode2(xorDecode1("Ak\u001c\u0011\u0010V)!\\")),
      xorDecode2(xorDecode1("{h\u0000T\u001dCiR\u0000^Nh\u0012\u001b\u000bV'\u0013\u001b\f\u00026ET\rGd\u001a\u001a\u001aQ'\u0014\u0012\nGuU\u0017\u0011Oe\u0014\u0000")),
      xorDecode2(xorDecode1("{h\u0000T\u001dCiR\u0000^Nh\u0012\u001b\u000bV'\u0011\u0001\fKi\u0012T\u001dMj\u0017\u0015\n\u0003")),
      xorDecode2(xorDecode1("qh\u0007\u0006\u0007\u0003'!\u001c\u001b\u0002t\u0010\u0006\bGuU\u001d\r\u0002d\u0000\u0006\fGi\u0001\u0018\u0007\u0002a\u0000\u0018\u0012\f")),
      xorDecode2(xorDecode1("rk\u0010\u0015\rG'\u0001\u0006\u0007\u0002fU\u001a\u0011L*\u0003\u0011\nGu\u0014\u001a\r\u0002p\u001a\u0006\u0012F)")),
      xorDecode2(xorDecode1("rf\u0006\u0007\tMu\u0011T\rWt\u0005\u0011\u001dVb\u0011T\rVh\u0019\u0011\u0010\f")),
      xorDecode2(xorDecode1("vo\u0014\u0000^Wt\u0010\u0006\u0010Cj\u0010T\u0017Q'\u0014\u0018\fGf\u0011\r^KiU\u0001\rG)")),
      xorDecode2(xorDecode1("ao\u0010\u0017\u0015\u0002~\u001a\u0001\f\u0002j\u0010\u0007\rC`\u0010T\u0017Le\u001a\f^Dh\u0007T\u001aGs\u0014\u001d\u0012Q")),
      xorDecode2(xorDecode1("gu\u0007\u001b\f\u0002p\u001d\u001d\u0012G'\u0016\u001b\u0010Lb\u0016\u0000\u0017L`")),
      xorDecode2(xorDecode1(
            "ru\u0010\u0007\r\u0002 \u0007\u0011\u001dMq\u0010\u0006^C'\u0019\u001b\u001dIb\u0011T\u001fAd\u001a\u0001\u0010V U\u001b\u0010\u0002a\u0007\u001b\u0010V'\u0005\u0015\u0019G)"
         )
      ),
      xorDecode2(xorDecode1("gu\u0007\u001b\f\u0002r\u001b\u0015\u001cNbU\u0000\u0011\u0002k\u001a\u0013\u0017L)")),
      xorDecode2(xorDecode1("vo\u0010T\u001dNn\u0010\u001a\n\u0002o\u0014\u0007^@b\u0010\u001a^Ww\u0011\u0015\nGc[")),
      xorDecode2(xorDecode1("ki\u0003\u0015\u0012KcU\u0001\rGu\u001b\u0015\u0013G'\u001a\u0006^Rf\u0006\u0007\tMu\u0011Z")),
      xorDecode2(xorDecode1("ah\u001b\u001a\u001bAs\u001c\u001a\u0019\u0002s\u001aT\rGu\u0003\u0011\f")),
      xorDecode2(xorDecode1(
            "ao\u0010\u0017\u0015\u0002n\u001b\u0000\u001bPi\u0010\u0000^Qb\u0001\u0000\u0017L`\u0006T\u0011P'\u0001\u0006\u0007\u0002f\u001b\u001b\nJb\u0007T\tMu\u0019\u0010"
         )
      ),
      xorDecode2(xorDecode1("lh\u001b\u0011^MaU\r\u0011WuU\u0017\u0016Cu\u0014\u0017\nGu\u0006T\u001dCiU\u0018\u0011E'\u001c\u001aP")),
      xorDecode2(xorDecode1("ah\u001b\u0000\u001fAsU\u0017\u000bQs\u001a\u0019\u001bP'\u0006\u0001\u000eRh\u0007\u0000")),
      xorDecode2(xorDecode1("rk\u0010\u0015\rG'\u0002\u0015\u0017V)[Z")),
      xorDecode2(xorDecode1("Ak\u001c\u0011\u0010V)<6V")),
      xorDecode2(xorDecode1("nh\u0012\u001d\u0010\u0002f\u0001\u0000\u001bOw\u0001\u0007^G\u007f\u0016\u0011\u001bFb\u0011U")),
      xorDecode2(xorDecode1("Nh\u0012\u001d\u0010\u0002u\u0010\u0007\u000eMi\u0006\u0011D")),
      xorDecode2(xorDecode1("gu\u0007\u001b\f\u0002*U\u001a\u0011\u0002u\u0010\u0004\u0012['\u0013\u0006\u0011O'\u0019\u001b\u0019Ki\u0006\u0011\fTb\u0007Z")),
      xorDecode2(xorDecode1("qh\u0007\u0006\u0007\u0003' \u001a\u001f@k\u0010T\nM'\u0016\u001b\u0010Lb\u0016\u0000P")),
      xorDecode2(xorDecode1("qb\u0007\u0002\u001bP'\u0001\u001d\u0013GcU\u001b\u000bV")),
      xorDecode2(xorDecode1("vo\u0014\u0000^KtU\u001a\u0011V'\u0014T\bGs\u0010\u0006\u001fL'''Sak\u0014\u0007\rKdU\u0015\u001dAh\u0000\u001a\n\f")),
      xorDecode2(xorDecode1("qh\u0007\u0006\u0007\u0003'!\u001c\u0017Q'\u0002\u001b\fNcU\u001d\r\u0002d\u0000\u0006\fGi\u0001\u0018\u0007\u0002a\u0000\u0018\u0012\f")),
      xorDecode2(xorDecode1(
            "wi\u0011\u0011\f\u00026FT\u001fAd\u001a\u0001\u0010VtU\u0017\u001fLi\u001a\u0000^Cd\u0016\u0011\rQ''\u0001\u0010GT\u0016\u0015\u000eG'6\u0018\u001fQt\u001c\u0017"
         )
      ),
      xorDecode2(xorDecode1("ah\u001b\u001a\u001bAs\u001c\u001b\u0010\u0002k\u001a\u0007\n\u0003'%\u0018\u001bCt\u0010T\tCn\u0001ZP\f")),
      xorDecode2(xorDecode1("qb\u0007\u0002\u001bP'\u0007\u0011\u0014Gd\u0001\u0011\u001a\u0002t\u0010\u0007\rKh\u001b")),
      xorDecode2(xorDecode1("rk\u0010\u0015\rG'\u0006\u0011\u001b\u0002s\u001d\u0011^Nf\u0000\u001a\u001dJ'\u0005\u0015\u0019G'\u0013\u001b\f\u0002o\u0010\u0018\u000e")),
      xorDecode2(xorDecode1("rk\u0010\u0015\rG'\u0001\u0006\u0007\u0002fU\u0010\u0017Da\u0010\u0006\u001bLsU\u0003\u0011Pk\u0011")),
      xorDecode2(xorDecode1("vo\u0014\u0000^Wt\u0010\u0006\u0010Cj\u0010T\u0017Q'\u0014\u0018\fGf\u0011\r^Nh\u0012\u0013\u001bF'\u001c\u001aP")),
      xorDecode2(xorDecode1("cd\u0016\u001b\u000bLsU\u0007\u000bQw\u0010\u0017\nGcU\u0007\nMk\u0010\u001aP")),
      xorDecode2(xorDecode1("wi\u0007\u0011\u001dM`\u001b\u001d\rGcU\u0006\u001bQw\u001a\u001a\rG'\u0016\u001b\u001aG")),
      xorDecode2(xorDecode1("uf\u001c\u0000^\u00147U\u0007\u001bAh\u001b\u0010\r\u0002s\u001d\u0011\u0010\u0002u\u0010\u0000\f[")),
      xorDecode2(xorDecode1("rk\u0010\u0015\rG'\u0001\u0006\u0007\u0002f\u0012\u0015\u0017L'\u0019\u0015\nGu")),
      xorDecode2(xorDecode1("VhU\u0018\u0011En\u001bT\nM'\u0001\u001c\u0017Q'\u0002\u001b\fNc")),
      xorDecode2(xorDecode1(
            "rk\u0010\u0015\rG'\u0012\u001b^VhU\u0000\u0016G'4\u0017\u001dMr\u001b\u0000^of\u001b\u0015\u0019Gj\u0010\u001a\n\u0002w\u0014\u0013\u001b\u0002s\u001aT\u001aM'\u0001\u001c\u0017Q)"
         )
      ),
      xorDecode2(xorDecode1("cd\u0016\u001b\u000bLsU\u0000\u001bOw\u001a\u0006\u001fPn\u0019\r^Fn\u0006\u0015\u001cNb\u0011Z")),
      xorDecode2(xorDecode1("{h\u0000T\u0013C~U\u001b\u0010N~U\u0001\rG'DT\u001dJf\u0007\u0015\u001dVb\u0007T\u001fV'\u001a\u001a\u001dG)")),
      xorDecode2(xorDecode1("{h\u0000T\u0010Gb\u0011T\u001f\u0002j\u0010\u0019\u001cGu\u0006T\u001fAd\u001a\u0001\u0010V")),
      xorDecode2(xorDecode1("cs\u0001\u0011\u0013Rs\u001c\u001a\u0019\u0002s\u001aT\fG*\u0010\u0007\nCe\u0019\u001d\rJ")),
      xorDecode2(xorDecode1("{h\u0000T\u0010Gb\u0011T\nM'\u0006\u0011\n\u0002~\u001a\u0001\f\u0002c\u001c\u0007\u000eNf\fT\u0010Cj\u0010Z")),
      xorDecode2(xorDecode1("Nn\u0018\u001d\n\u00117")),
      xorDecode2(xorDecode1("gu\u0007\u001b\f\u0002*U\u0012\u001fKk\u0010\u0010^VhU\u0010\u001bAh\u0011\u0011^Ru\u001a\u0012\u0017Nb[")),
      xorDecode2(xorDecode1("gu\u0007\u001b\f\u0002*U\u0018\u0011En\u001b\u0007\u001bPq\u0010\u0006^On\u0006\u0019\u001fVd\u001d")),
      xorDecode2(xorDecode1("ru\u0010\u0007\r\u0002 \u0016\u001c\u001fL`\u0010T\u0007Mr\u0007T\u000eCt\u0006\u0003\u0011PcRT\u0011L'\u0013\u0006\u0011LsU\u0004\u001fEb[")),
      xorDecode2(xorDecode1("cd\u0016\u001b\u000bLsU\u0004\u001bPj\u0014\u001a\u001bLs\u0019\r^Fn\u0006\u0015\u001cNb\u0011Z")),
      xorDecode2(xorDecode1("rk\u0010\u0015\rG'\u0007\u0011\u0012Mf\u0011T\nJn\u0006T\u000eC`\u0010")),
      xorDecode2(xorDecode1("rk\u0010\u0015\rG'\u0001\u0006\u0007\u0002f\u0012\u0015\u0017L")),
      xorDecode2(xorDecode1("{h\u0000\u0006^KwX\u0015\u001aFu\u0010\u0007\r\u0002n\u0006T\u001fNu\u0010\u0015\u001a['\u001c\u001a^Wt\u0010")),
      xorDecode2(xorDecode1("rk\u0010\u0015\rG'\u0001\u0006\u0007\u0002f\u0012\u0015\u0017L'\u001c\u001a^\u0017'\u0018\u001d\u0010Ws\u0010\u0007")),
      xorDecode2(xorDecode1("Ci\u0011T\u001f\u0002w\u0014\u0007\rUh\u0007\u0010^\u000f'%\u0018\u001bCt\u0010T\nP~U\u0015\u0019Cn\u001b")),
      xorDecode2(xorDecode1("vo\u001c\u0007^Uh\u0007\u0018\u001a\u0002c\u001a\u0011\r\u0002i\u001a\u0000^Cd\u0016\u0011\u000eV'\u001b\u0011\t\u0002w\u0019\u0015\u0007Gu\u0006Z")),
      xorDecode2(xorDecode1("vu\fT\u001fEf\u001c\u001aR\u0002h\u0007T\u001dPb\u0014\u0000\u001b\u0002fU\u001a\u001bU'\u0014\u0017\u001dMr\u001b\u0000")),
      xorDecode2(xorDecode1("{h\u0000T\u0013Wt\u0001T\u001bLs\u0010\u0006^@h\u0001\u001c^C'\u0000\u0007\u001bPi\u0014\u0019\u001b")),
      xorDecode2(xorDecode1("Ak\u001c\u0011\u0010V)&6V")),
      xorDecode2(xorDecode1("Ak\u001c\u0011\u0010V)80V")),
      xorDecode2(xorDecode1("jb\u0019\u0002\u001bVn\u0016\u0015")),
      xorDecode2(xorDecode1("Ak\u001c\u0011\u0010V)?5V")),
      xorDecode2(xorDecode1("Ak\u001c\u0011\u0010V)$\\")),
      xorDecode2(xorDecode1("Ak\u001c\u0011\u0010V)?7V")),
      xorDecode2(xorDecode1("Ak\u001c\u0011\u0010V)?0V")),
      xorDecode2(xorDecode1("gu\u0007\u001b\f\u0002*U\u001b\u000bV'\u001a\u0012^Ob\u0018\u001b\f[&")),
      xorDecode2(xorDecode1("\u0011=U \f['\u0000\u0007\u0017L`U\u0015^Fn\u0013\u0012\u001bPb\u001b\u0000^Ef\u0018\u0011SUh\u0007\u0018\u001a")),
      xorDecode2(xorDecode1(
            "\u0013=U \f['\u0016\u0018\u0011Qn\u001b\u0013^cK9T\u0011Rb\u001bT\tGeX\u0016\fMp\u0006\u0011\f\u0002p\u001c\u001a\u001aMp\u0006X^Ci\u0011T\fGk\u001a\u0015\u001aKi\u0012"
         )
      ),
      xorDecode2(xorDecode1("gu\u0007\u001b\f\u0002*U\u0001\u0010Ce\u0019\u0011^VhU\u0018\u0011CcU\u0013\u001fObT")),
      xorDecode2(xorDecode1("\u0016=U \f['\u0007\u0011\u001cMh\u0001\u001d\u0010E'\f\u001b\u000bP'\u0016\u001b\u0013Rr\u0001\u0011\f")),
      xorDecode2(xorDecode1(
            "vhU\u0012\u0017Z'\u0001\u001c\u0017Q'\u0001\u0006\u0007\u0002s\u001d\u0011^Dh\u0019\u0018\u0011Un\u001b\u0013^\nn\u001bT\u0011Pc\u0010\u0006W\u0018"
         )
      ),
      xorDecode2(xorDecode1("ak\u001a\u0007\u001b\u0002F98^Wi\u001b\u0011\u001dGt\u0006\u0015\f['\u0005\u0006\u0011Eu\u0014\u0019\r")),
      xorDecode2(xorDecode1(
            "\u0010=U \f['\u0016\u0018\u001bCu\u001c\u001a\u0019\u0002~\u001a\u0001\f\u0002p\u0010\u0016S@u\u001a\u0003\rGu\u0006T\u001dCd\u001d\u0011^Du\u001a\u0019^Vh\u001a\u0018\r\u000f9\u001c\u001a\nGu\u001b\u0011\n\u0002h\u0005\u0000\u0017Mi\u0006"
         )
      ),
      xorDecode2(xorDecode1(
            "\u0017=U \f['\u0006\u0011\u0012Gd\u0001\u001d\u0010E'\u0014T\u001aKa\u0013\u0011\fGi\u0001T\bGu\u0006\u001d\u0011L'\u001a\u0012^hf\u0003\u0015^Du\u001a\u0019^Vo\u0010T\u000eNf\fY\u0019Cj\u0010T\u0013Gi\u0000"
         )
      ),
      xorDecode2(xorDecode1("pr\u001b\u0011-Af\u0005\u0011^Lb\u0010\u0010\r\u0002f\u0017\u001b\u000bV'AL\u0013G`U\u001b\u0018\u0002t\u0005\u0015\fG''53")),
      xorDecode2(xorDecode1("vhU\u0004\u0012C~U&\u000bLb&\u0017\u001fRbU\u0019\u001fIbU\u0007\u000bPbU\r\u0011W'\u0005\u0018\u001f['\u0013\u0006\u0011O")),
      xorDecode2(xorDecode1(
            "qh\u0007\u0006\u0007\u000e'\u0014\u001a^Gu\u0007\u001b\f\u0002o\u0014\u0007^Md\u0016\u0001\fGcU\u0003\u0016Kk\u0006\u0000^Nh\u0014\u0010\u0017L`U&\u000bLb&\u0017\u001fRb"
         )
      ),
      xorDecode2(xorDecode1("Ci\u0011T\tKi\u0011\u001b\tQ'\u0017\u0011\u0018Mu\u0010T\u0012Mf\u0011\u001d\u0010E'\u0001\u001c\u001b\u0002`\u0014\u0019\u001b")),
      xorDecode2(xorDecode1("Js\u0001\u0004D\r(\u0002\u0003\t\fu\u0000\u001a\u001bQd\u0014\u0004\u001b\fd\u001a\u0019")),
      xorDecode2(xorDecode1("lhU\u0019\u001fEn\u0016")),
      xorDecode2(xorDecode1("lhU\u0003\u001bCw\u001a\u001a\r")),
      xorDecode2(xorDecode1("{h\u0000\u0006^qs\u0014\u001f\u001b")),
      xorDecode2(xorDecode1("fr\u0010\u0018^mw\u0001\u001d\u0011Lt")),
      xorDecode2(xorDecode1("mw\u0005\u001b\u0010Gi\u0001S\r\u0002T\u0001\u0015\u0015G")),
      xorDecode2(xorDecode1("qs\u0014\u001f\u001b\u0002F\u0019\u0018")),
      xorDecode2(xorDecode1("qs\u0014\u001f\u001b\u00026")),
      xorDecode2(xorDecode1("qs\u0014\u001f\u001b\u0002_")),
      xorDecode2(xorDecode1("Ak\u001c\u0011\u0010V)!0V")),
      xorDecode2(xorDecode1("qs\u0014\u001f\u001b\u00026E")),
      xorDecode2(xorDecode1("lhU\u0006\u001bVu\u0010\u0015\nKi\u0012")),
      xorDecode2(xorDecode1("lhU\u0004\fC~\u0010\u0006")),
      xorDecode2(xorDecode1("ru\u0010\u0004\u001fPn\u001b\u0013^VhU\u0010\u000bGkU\u0003\u0017VoOT")),
      xorDecode2(xorDecode1("qs\u0014\u001f\u001b\u00022")),
      xorDecode2(xorDecode1("Ak\u001c\u0011\u0010V)>5V")),
      xorDecode2(xorDecode1("fu\u001a\u0004\u000eKi\u0012T")),
      xorDecode2(xorDecode1("Ak\u001c\u0011\u0010V)=0V")),
      xorDecode2(xorDecode1("Ak\u001c\u0011\u0010V)0\\")),
      xorDecode2(xorDecode1("Ak\u001c\u0011\u0010V)&7V")),
      xorDecode2(xorDecode1("\fw\u0016\u0019")),
      xorDecode2(xorDecode1("ru\u0014\r\u001bP'\u0018\u0015\u0007\u0002e\u0010T\u000bQb\u0011")),
      xorDecode2(xorDecode1("lhU\u0006\u001bVu\u0010\u0015\n\u0002n\u0006T\u000eMt\u0006\u001d\u001cNbT")),
      xorDecode2(xorDecode1("ub\u0014\u0004\u0011LtU\u0017\u001fLi\u001a\u0000^@bU\u0001\rGc")),
      xorDecode2(xorDecode1("of\u0012\u001d\u001d\u0002d\u0014\u001a\u0010MsU\u0016\u001b\u0002r\u0006\u0011\u001a")),
      xorDecode2(xorDecode1(
            "kaU\r\u0011W'\u0014\u0006\u001b\u0002t\u0000\u0006\u001b\u0002d\u0019\u001d\u001dI'R5\u001dAb\u0005\u0000Y\u0002s\u001aT\u001cG`\u001c\u001a^Vo\u0010T\u001aWb\u0019"
         )
      ),
      xorDecode2(xorDecode1("ru\u0014\r\u001bP'\u0016\u0015\u0010Lh\u0001T\u001cG'\u0000\u0007\u001bF")),
      xorDecode2(xorDecode1("rk\u0010\u0015\rG'\u0016\u001b\u0010Dn\u0007\u0019^[h\u0000\u0006^Fr\u0010\u0018^Un\u0001\u001c^b~\u0010\u0018>")),
      xorDecode2(xorDecode1("Ak\u001c\u0011\u0010V)#7V")),
      xorDecode2(xorDecode1("{h\u0000\u0006^Qs\u0014\u001f\u001b\u0018")),
      xorDecode2(xorDecode1("ub\u0014\u0004\u0011LtU\u0019\u001f['\u0017\u0011^Wt\u0010\u0010")),
      xorDecode2(xorDecode1("of\u0012\u001d\u001d\u0002j\u0014\r^@bU\u0001\rGc")),
      xorDecode2(xorDecode1("{h\u0000\u0006^Mw\u0005\u001b\u0010Gi\u0001S\r\u0002t\u0001\u0015\u0015G=")),
      xorDecode2(xorDecode1("{h\u0000T\u001dCiU\u0006\u001bVu\u0010\u0015\n\u0002a\u0007\u001b\u0013\u0002s\u001d\u001d\r\u0002c\u0000\u0011\u0012")),
      xorDecode2(xorDecode1("jb\u0007\u0016\u0012Cp")),
      xorDecode2(xorDecode1("fn\u0012\u0007\u0017VbU\\\u0013Gj\u0017\u0011\fQ.")),
      xorDecode2(xorDecode1("`n\u001a\u001c\u001fXf\u0007\u0010^\nj\u0010\u0019\u001cGu\u0006]")),
      xorDecode2(xorDecode1("nb\u0012\u0011\u0010F \u0006T/Wb\u0006\u0000^\nj\u0010\u0019\u001cGu\u0006]")),
      xorDecode2(xorDecode1("me\u0006\u0011\fTf\u0001\u001b\f['\u0004\u0001\u001bQsU\\\u0013Gj\u0017\u0011\fQ.")),
      xorDecode2(xorDecode1("vo\u0010T6Mk\fT9Pf\u001c\u0018^\nj\u0010\u0019\u001cGu\u0006]")),
      xorDecode2(xorDecode1("qo\u0010\u0011\u000e\u0002t\u001d\u0011\u001fPb\u0007")),
      xorDecode2(xorDecode1("qd\u001a\u0006\u000eKh\u001bT\u001dCs\u0016\u001c\u001bP']\u0019\u001bOe\u0010\u0006\r\u000b")),
      xorDecode2(xorDecode1("wi\u0011\u0011\fEu\u001a\u0001\u0010F'\u0005\u0015\rQ']\u0019\u001bOe\u0010\u0006\r\u000b")),
      xorDecode2(xorDecode1("qb\u0014T-Nr\u0012TVOb\u0018\u0016\u001bPt\\")),
      xorDecode2(xorDecode1("hr\u001b\u0013\u0012G'\u0005\u001b\nKh\u001bTVOb\u0018\u0016\u001bPt\\")),
      xorDecode2(xorDecode1("un\u0001\u0017\u0016\u0005tU\u001c\u0011Wt\u0010TVOb\u0018\u0016\u001bPt\\")),
      xorDecode2(xorDecode1("rn\u0007\u0015\nG \u0006T\nPb\u0014\u0007\u000bPb")),
      xorDecode2(xorDecode1("ah\u001a\u001fYQ'\u0014\u0007\rKt\u0001\u0015\u0010V")),
      xorDecode2(xorDecode1("fb\u0013\u0011\u0010Qb")),
      xorDecode2(xorDecode1("dn\u0006\u001c\u0017L`U\u0017\u0011Ls\u0010\u0007\n\u0002/\u0018\u0011\u0013@b\u0007\u0007W")),
      xorDecode2(xorDecode1("ph\u0018\u0011\u0011\u0002!U>\u000bNn\u0010\u0000")),
      xorDecode2(xorDecode1("qs\u0007\u0011\u0010Es\u001d")),
      xorDecode2(xorDecode1("vo\u0010T\u0015Ln\u0012\u001c\n\u0005tU\u0007\tMu\u0011")),
      xorDecode2(xorDecode1("ah\u001a\u001f\u0017L`")),
      xorDecode2(xorDecode1("vh\u0000\u0006\u0017QsU\u0000\fCwU\\\u0013Gj\u0017\u0011\fQ.")),
      xorDecode2(xorDecode1("dn\u0007\u0011\u0013Cl\u001c\u001a\u0019")),
      xorDecode2(xorDecode1("ru\u001c\u001a\u001dG'4\u0018\u0017\u0002u\u0010\u0007\u001dWb")),
      xorDecode2(xorDecode1("cu\u0018\u001b\u000bP")),
      xorDecode2(xorDecode1("uh\u001a\u0010\u001dWs\u0001\u001d\u0010E")),
      xorDecode2(xorDecode1("fb\u0018\u001b\u0010\u0002t\u0019\u0015\u0007Gu")),
      xorDecode2(xorDecode1("nh\u0006\u0000^An\u0001\r^\nj\u0010\u0019\u001cGu\u0006]")),
      xorDecode2(xorDecode1("vo\u0010T6C}\u0010\u0011\u0012\u0002D\u0000\u0018\n\u0002/\u0018\u0011\u0013@b\u0007\u0007W")),
      xorDecode2(xorDecode1("uh\u001a\u0010\u001dWs")),
      xorDecode2(xorDecode1("uf\u0001\u0017\u0016Vh\u0002\u0011\f\u0002/\u0018\u0011\u0013@b\u0007\u0007W")),
      xorDecode2(xorDecode1("au\u0014\u0012\nKi\u0012")),
      xorDecode2(xorDecode1("on\u001b\u001d\u0010E")),
      xorDecode2(xorDecode1("qo\u001c\u0011\u0012F'\u001a\u0012^cu\u0007\u0015\b")),
      xorDecode2(xorDecode1("jn\u0001\u0007")),
      xorDecode2(xorDecode1("eu\u0014\u001a\u001a\u0002s\u0007\u0011\u001b\u0002/\u0018\u0011\u0013@b\u0007\u0007W")),
      xorDecode2(xorDecode1("ak\u001a\u0017\u0015\u0002s\u001a\u0003\u001bP']\u0019\u001bOe\u0010\u0006\r\u000b")),
      xorDecode2(xorDecode1("dn\u0006\u001c\u0017L`")),
      xorDecode2(xorDecode1("tf\u0018\u0004\u0017PbU\u0007\u0012C~\u0010\u0006")),
      xorDecode2(xorDecode1("c`\u001c\u0018\u0017V~")),
      xorDecode2(xorDecode1("fu\u0014\u0013\u0011L'\u0006\u0018\u001f[b\u0007")),
      xorDecode2(xorDecode1("qj\u001c\u0000\u0016Ki\u0012")),
      xorDecode2(xorDecode1("ru\u0014\r\u001bP")),
      xorDecode2(xorDecode1("ub\u0014\u0004\u0011LF\u001c\u0019")),
      xorDecode2(xorDecode1("vu\u0010\u0011^ei\u001a\u0019\u001b\u0002Q\u001c\u0018\u0012C`\u0010TVOb\u0018\u0016\u001bPt\\")),
      xorDecode2(xorDecode1("eh\u0017\u0018\u0017L'\u0011\u001d\u000eNh\u0018\u0015\u001d[")),
      xorDecode2(xorDecode1("fp\u0014\u0006\u0018\u0002D\u0014\u001a\u0010MiU\\\u0013Gj\u0017\u0011\fQ.")),
      xorDecode2(xorDecode1("pf\u001b\u0013\u001bF")),
      xorDecode2(xorDecode1("rk\u0014\u0013\u000bG'6\u001d\n[']\u0019\u001bOe\u0010\u0006\r\u000b")),
      xorDecode2(xorDecode1("qo\u0010\u0011\u000e\u0002O\u0010\u0006\u001aGuU\\\u0013Gj\u0017\u0011\fQ.")),
      xorDecode2(xorDecode1("vb\u0018\u0004\u0012G'\u001a\u0012^kl\u001a\u0002^\nj\u0010\u0019\u001cGu\u0006]")),
      xorDecode2(xorDecode1("vu\u001c\u0016\u001fN'\u0001\u001b\nGjU\\\u0013Gj\u0017\u0011\fQ.")),
      xorDecode2(xorDecode1("vo\u001c\u0011\bKi\u0012")),
      xorDecode2(xorDecode1("ub\u0014\u0004\u0011LW\u001a\u0003\u001bP")),
      xorDecode2(xorDecode1("uf\u0001\u0011\fDf\u0019\u0018^Sr\u0010\u0007\n\u0002/\u0018\u0011\u0013@b\u0007\u0007W")),
      xorDecode2(xorDecode1("ob\u0007\u0018\u0017L \u0006T\u001dP~\u0006\u0000\u001fN']\u0019\u001bOe\u0010\u0006\r\u000b")),
      xorDecode2(xorDecode1("gu\u001b\u0011\rV'\u0001\u001c\u001b\u0002d\u001d\u001d\u001dIb\u001b")),
      xorDecode2(xorDecode1("dn\u0012\u001c\n\u0002F\u0007\u0011\u0010C']\u0019\u001bOe\u0010\u0006\r\u000b")),
      xorDecode2(xorDecode1("vo\u0010T\fGt\u0001\u0018\u001bQtU\u0013\u0016Mt\u0001")),
      xorDecode2(xorDecode1("oh\u001b\u001fYQ'\u0013\u0006\u0017Gi\u0011TVOb\u0018\u0016\u001bPt\\")),
      xorDecode2(xorDecode1("df\u0018\u001d\u0012['\u0016\u0006\u001bQsU\\\u0013Gj\u0017\u0011\fQ.")),
      xorDecode2(xorDecode1("un\u0001\u0017\u0016\u0005tU\u0004\u0011Vn\u001a\u001a")),
      xorDecode2(xorDecode1("kj\u0005T\u001dCs\u0016\u001c\u001bP")),
      xorDecode2(xorDecode1("dk\u0010\u0000\u001dJn\u001b\u0013")),
      xorDecode2(xorDecode1("or\u0007\u0010\u001bP'8\r\rVb\u0007\r^\nj\u0010\u0019\u001cGu\u0006]")),
      xorDecode2(xorDecode1("qo\u001c\u0018\u0011\u0002q\u001c\u0018\u0012C`\u0010TVOb\u0018\u0016\u001bPt\\")),
      xorDecode2(xorDecode1("jb\u0007\u001bYQ'\u0004\u0001\u001bQsU\\\u0013Gj\u0017\u0011\fQ.")),
      xorDecode2(xorDecode1("fu\u0000\u001d\u001aKdU\u0006\u0017Vr\u0014\u0018^\nj\u0010\u0019\u001cGu\u0006]")),
      xorDecode2(xorDecode1("`k\u0014\u0017\u0015\u0002l\u001b\u001d\u0019JsR\u0007^Dh\u0007\u0000\fGt\u0006")),
      xorDecode2(xorDecode1("eb\u0007\u0000\fWc\u0010S\r\u0002D\u0014\u0000^\nj\u0010\u0019\u001cGu\u0006]")),
      xorDecode2(xorDecode1("fh\u0007\u001d\u001d\u0005tU\u0005\u000bGt\u0001")),
      xorDecode2(xorDecode1("Nf\u001b\u0010\rAf\u0005\u0011")),
      xorDecode2(xorDecode1("Ob\u0018\u0016\u001bPtU\u0018\u001fLc\u0006\u0017\u001fRb")),
      xorDecode2(xorDecode1("Ob\u0018\u0016\u001bPtU\u0019\u001fR")),
      xorDecode2(xorDecode1("Of\u0005")),
      xorDecode2(xorDecode1("Ak\u001c\u0011\u0010V)00V")),
      xorDecode2(xorDecode1("Ak\u001c\u0011\u0010V):6V")),
      xorDecode2(xorDecode1("\u00177")),
      xorDecode2(xorDecode1("lr\u0018\u0016\u001bP'\u001d\u0011\u0012F'\u001c\u001a^@k\u0000\u0011")),
      xorDecode2(xorDecode1("\u001ew\u0014\u0013\u001b\u00026K")),
      xorDecode2(xorDecode1("lr\u0018\u0016\u001bP'\u001c\u001a^@f\u001b\u001f^KiU\u0013\fGb\u001b")),
      xorDecode2(xorDecode1("Ak\u001c\u0011\u0010V)!7V")),
      xorDecode2(xorDecode1("`f\u001b\u001f")),
      xorDecode2(xorDecode1("un\u0001\u001c\u001aPf\u0002T")),
      xorDecode2(xorDecode1("\u00137")),
      xorDecode2(xorDecode1("qb\u0019\u0011\u001dV'\u0014\u001a^Me\u001f\u0011\u001dV'\u0001\u001b^Un\u0001\u001c\u001aPf\u0002T\u0011P'\u0011\u0011\u000eMt\u001c\u0000")),
      xorDecode2(xorDecode1("fb\u0005\u001b\rKsU")),
      xorDecode2(xorDecode1("ck\u0019")),
      xorDecode2(xorDecode1("\u001ew\u0014\u0013\u001b\u00024K")),
      xorDecode2(xorDecode1("mi\u0010")),
      xorDecode2(xorDecode1("\u001ew\u0014\u0013\u001b\u00025K")),
      xorDecode2(xorDecode1("dn\u0003\u0011")),
      xorDecode2(xorDecode1("ak\u001a\u0007\u001b\u0002p\u001c\u001a\u001aMp")),
      xorDecode2(xorDecode1("\u001ew\u0014\u0013\u001b\u00023K")),
      xorDecode2(xorDecode1("Ak\u001c\u0011\u0010V)'7V")),
      xorDecode2(xorDecode1("\u0018=\u0016\u0018\u0011Qb\u0016\u001b\u0010")),
      xorDecode2(xorDecode1("Ak\u001c\u0011\u0010V)'0V")),
      xorDecode2(xorDecode1("\u000fi\u0000\u0018\u0012\u000f")),
      xorDecode2(xorDecode1("\u0018=\u0019\u001b\u0019Mr\u0001")),
      xorDecode2(xorDecode1("\u0018=")),
      xorDecode2(xorDecode1(
            "{h\u0000T\fGs\u0014\u001d\u0010\u0002~\u001a\u0001\f\u0002t\u001e\u001d\u0012Nt[T'Mr\u0007T\u0011@m\u0010\u0017\nQ'\u0019\u0015\u0010F'\u0002\u001c\u001bPbU\r\u0011W'\u0011\u001d\u001bF"
         )
      ),
      xorDecode2(xorDecode1(
            "{h\u0000T\u0016Cq\u0010T\u001cGb\u001bT\u0019Pf\u001b\u0000\u001bF'\u0014\u001a\u0011Vo\u0010\u0006^Nn\u0013\u0011P\u0002E\u0010T\u0013Mu\u0010T\u001dCu\u0010\u0012\u000bN'\u0001\u001c\u0017Q'\u0001\u001d\u0013G&"
         )
      ),
      xorDecode2(xorDecode1("\u0018=\u0019\u001b\rVd\u001a\u001a")),
      xorDecode2(xorDecode1("Ew")),
      xorDecode2(xorDecode1("{h\u0000T\u001aM'\u001b\u001b\n\u0002o\u0014\u0002\u001b\u0002f\u001b\r^MaU\u0000\u0016KtU\u001d\nGjU\u0000\u0011\u0002t\u0010\u0018\u0012")),
      xorDecode2(xorDecode1("Ak\u001c\u0011\u0010V)=5V")),
      xorDecode2(xorDecode1("qb\u0019\u0018D")),
      xorDecode2(xorDecode1("lr\u0018\u0016\u001bP'\f\u001b\u000b\u0002h\u0002\u001a^KiU\u0016\u0012Wb")),
      xorDecode2(xorDecode1("EwU\u0011\u001fAo")),
      xorDecode2(xorDecode1("qo\u001a\u0004\r\u0002t\u0001\u001b\u001dI'\u001c\u001a^Eu\u0010\u0011\u0010")),
      xorDecode2(xorDecode1("\u0018'\u0006\u0011\u0012N'\u0013\u001b\f\u0002")),
      xorDecode2(xorDecode1("\u0018'\u0017\u0001\u0007\u0002a\u001a\u0006^")),
      xorDecode2(xorDecode1("`r\f\u001d\u0010E'\u0014\u001a\u001a\u0002t\u0010\u0018\u0012Ki\u0012T\u0017Vb\u0018\u0007")),
      xorDecode2(xorDecode1("vo\u001c\u0007^Ks\u0010\u0019^KtU\u001a\u0011V'\u0016\u0001\fPb\u001b\u0000\u0012['\u0014\u0002\u001fKk\u0014\u0016\u0012G'\u0001\u001b^@r\f")),
      xorDecode2(xorDecode1("`r\fN")),
      xorDecode2(xorDecode1("{h\u0000\u0006^Oh\u001b\u0011\u0007\u0018'")),
      xorDecode2(xorDecode1("qb\u0019\u0011\u001dV'\u0014\u001a^Me\u001f\u0011\u001dV'\u0001\u001b^@r\fT\u0011P'\u0006\u0011\u0012N")),
      xorDecode2(xorDecode1("c`\u0012\u0006\u001bQt\u001c\u0002\u001b\u0002/^G^Qs\u0007\u0011\u0010Es\u001d]")),
      xorDecode2(xorDecode1("Ak\u001c\u0011\u0010V)!6V")),
      xorDecode2(xorDecode1("fb\u0013\u0011\u0010Qn\u0003\u0011^\u0002/^G^Fb\u0013\u0011\u0010Qb\\")),
      xorDecode2(xorDecode1("ah\u001b\u0000\fMk\u0019\u0011\u001a\u0002/^E^MaU\u0011\u001fAo\\")),
      xorDecode2(xorDecode1("cd\u0016\u0001\fCs\u0010T^\u0002/^G^Cs\u0001\u0015\u001dI.")),
      xorDecode2(xorDecode1("qb\u0019\u0011\u001dV'\u0016\u001b\u0013@f\u0001T\rV~\u0019\u0011")),
      xorDecode2(xorDecode1("Ak\u001c\u0011\u0010V)16V")),
      xorDecode2(xorDecode1("\u0002c\u0014\r\r\u0002f\u0012\u001b")),
      xorDecode2(xorDecode1("Ak\u001c\u0011\u0010V)?6V")),
      xorDecode2(xorDecode1("gf\u0007\u0018\u0017GuU\u0000\u0011Ff\f")),
      xorDecode2(xorDecode1("{h\u0000T\u0012Ct\u0001T\u0012M`\u0012\u0011\u001a\u0002n\u001bT")),
      xorDecode2(xorDecode1(
            "{h\u0000T\u0016Cq\u0010T>[b\u00194Nbp\u001d\u001d>\u0002r\u001b\u0006\u001bCcU\u0019\u001bQt\u0014\u0013\u001bQ'\u001c\u001a^[h\u0000\u0006^Ob\u0006\u0007\u001fEbX\u0017\u001bLs\u0007\u0011"
         )
      ),
      xorDecode2(xorDecode1(
            "ubU\u0007\nPh\u001b\u0013\u0012['\u0007\u0011\u001dMj\u0018\u0011\u0010F'\f\u001b\u000b\u0002c\u001aT\rM'\u001b\u001b\t\u0002s\u001aT\rGd\u0000\u0006\u001b\u0002~\u001a\u0001\f\u0002f\u0016\u0017\u0011Wi\u0001Z"
         )
      ),
      xorDecode2(xorDecode1("Gf\u0007\u0018\u0017GuU\u0000\u0011Ff\f")),
      xorDecode2(xorDecode1("{b\u0006\u0000\u001bPc\u0014\r")),
      xorDecode2(xorDecode1(
            "{h\u0000T\u0016Cq\u0010T\u0010MsU\r\u001bV'\u0006\u0011\n\u0002f\u001b\r^Rf\u0006\u0007\tMu\u0011T\fGd\u001a\u0002\u001bP~U\u0005\u000bGt\u0001\u001d\u0011Lt["
         )
      ),
      xorDecode2(xorDecode1(
            "\u0002r\u001b\u0006\u001bCcU\u0019\u001bQt\u0014\u0013\u001bQ'5\u0003\u0016KG\u001c\u001a^[h\u0000\u0006^Ob\u0006\u0007\u001fEbX\u0017\u001bLs\u0007\u0011"
         )
      ),
      xorDecode2(xorDecode1("Du\u001a\u0019D\u0002")),
      xorDecode2(xorDecode1(
            "fhU\u0000\u0016KtU\u0012\fMjU\u0000\u0016G'R\u0015\u001dAh\u0000\u001a\n\u0002j\u0014\u001a\u001fEb\u0018\u0011\u0010V U\u0015\fGfU\u001b\u0010\u0002h\u0000\u0006^Du\u001a\u001a\n\u0002p\u0010\u0016\u000eC`\u0010"
         )
      ),
      xorDecode2(xorDecode1(
            "kaU\r\u0011W'\u0011\u001b^Lh\u0001T\fGj\u0010\u0019\u001cGuU\u0019\u001fIn\u001b\u0013^Vo\u001c\u0007^Ao\u0014\u001a\u0019G'\u0001\u001c\u001bL'\u0016\u0015\u0010Ab\u0019T\u0017V'\u001c\u0019\u0013Gc\u001c\u0015\nGk\f"
         )
      ),
      xorDecode2(xorDecode1("[b\u0006\u0000\u001bPc\u0014\r")),
      xorDecode2(xorDecode1("\u0002~\u001a\u0001^Ao\u0014\u001a\u0019GcU\r\u0011WuU\u0006\u001bAh\u0003\u0011\f['\u0004\u0001\u001bQs\u001c\u001b\u0010Q")),
      xorDecode2(xorDecode1("ub\u0019\u0017\u0011ObU\u0000\u0011\u0002U\u0000\u001a\u001bqd\u0014\u0004\u001b\u0002")),
      xorDecode2(xorDecode1("{h\u0000T\u0016Cq\u0010T>Eu\u00104")),
      xorDecode2(xorDecode1("Ak\u001c\u0011\u0010V)1\\")),
      xorDecode2(xorDecode1("Ak\u001c\u0011\u0010V)#5V")),
      xorDecode2(xorDecode1("nh\u0016T;Pu\u001a\u0006D\u0002")),
      xorDecode2(xorDecode1("K=")),
      xorDecode2(xorDecode1("\u0002h\u0017\u001eD")),
      xorDecode2(xorDecode1("`h\u0000\u001a\u001a\u0002B\u0007\u0006\u0011P=U")),
      xorDecode2(xorDecode1("Ak\u001c\u0011\u0010V)86V")),
      xorDecode2(xorDecode1("nh\u0014\u0010\u0017L`[ZP\u0002W\u0019\u0011\u001fQbU\u0003\u001fKs")),
      xorDecode2(xorDecode1("Ak\u001c\u0011\u0010V)'\\")),
      xorDecode2(xorDecode1("Ak\u001c\u0011\u0010V)&0V")),
      xorDecode2(xorDecode1("nh\u0012\u0013\u0017L`U\u001b\u000bV)[Z")),
      xorDecode2(xorDecode1("Ak\u001c\u0011\u0010V)45V")),
      xorDecode2(xorDecode1("Ak\u001c\u0011\u0010V);5V")),
      xorDecode2(xorDecode1("Ak\u001c\u0011\u0010V)%0V"))
    };
    static final String[] STRINGS = il; // alias: method bodies reference the XOR pool as STRINGS[...]

    // ----- Delegate subsystems -----
    final ClientPackets packets = new ClientPackets(this);
    final ClientSound sound = new ClientSound(this);
    final TradeDuelBankPackets tradePackets = new TradeDuelBankPackets(this);
    final WidgetRenderer widgetRenderer = new WidgetRenderer(this);
    final GameInterface gameInterface = new GameInterface(this);
    final MenuController menus = new MenuController(this);
    final IncomingPackets incoming = new IncomingPackets(this);

    // ----- Network / streams -----
    ClientStream Jh; // obf: Jh — clientStream: outgoing packet stream (da)
    byte[] Uh; // obf: Uh — sessionBytes: handshake scratch bytes
    BitBuffer mg; // obf: mg — incomingPacket: inbound bit-buffer (ja)

    // ----- Scene / world / models -----
    int Ah; // obf: Ah — game-state scalar
    Scene Ek; // obf: Ek — type lb = Scene (3D renderer). NOTE: clean decompile declares `private lb Ek;` so Ek is the SCENE despite legacy "world" comment.
    int[] Gj; // obf: Gj — int buffer/table
    World Hh; // obf: Hh — type k = World (terrain/region). NOTE: clean decompile declares `private k Hh;` so Hh is the WORLD despite legacy "scene" comment.
    int[] Hj; // obf: Hj — int buffer/table
    int[] Jd; // obf: Jd — int buffer/table
    int[] Le; // obf: Le — int buffer/table
    int[] Ng; // obf: Ng — int buffer/table
    int[] Ni; // obf: Ni — int buffer/table
    int[] Se; // obf: Se — int buffer/table
    int[] Zf; // obf: Zf — int buffer/table
    int[] bg; // obf: bg — int buffer/table
    int eh; // obf: eh — game-state scalar
    int hf; // obf: hf — game-state scalar
    GameModel[] hg; // obf: hg — wallModels: wall GameModels (ca[1500])
    SurfaceSprite li; // obf: li — surface: 2D blitter (ba)
    GameModel[] rd; // obf: rd — npcModelCache: npc/anim GameModels (ca[500])
    int[] vc; // obf: vc — int buffer/table
    int[] vi; // obf: vi — int buffer/table
    int[] ye; // obf: ye — int buffer/table
    private int yg; // obf: yg — game-state scalar
    int[] yk; // obf: yk — int buffer/table

    // ----- Entities (players / NPCs) -----
    GameCharacter[] Ff; // obf: Ff — knownPlayers: players known this region (ta[500])
    GameCharacter[] Tb; // obf: Tb — playersInView: players in view (ta[500])
    GameCharacter[] We; // obf: We — npcsCache: id->npc cache (ta[4000])
    GameCharacter[] Zg; // obf: Zg — knownNpcs: npcs known this region (ta[500])
    GameCharacter[] rg; // obf: rg — npcsInView/playersLast: prev-tick entity buffer (ta[500])
    GameCharacter[] te; // obf: te — playersCache: id->player cache (ta[5000])
    GameCharacter wi; // obf: wi — localPlayer: local player (ta)

    // ----- UI panels & chat -----
    private Panel Af; // obf: Af — panelQuest: quest/char-design Panel (qa)
    private Panel ge; // obf: ge — panelGameAlt: game/trade Panel (qa)
    String ig; // obf: ig — selectedItemName: selected item name
    private Panel yi; // obf: yi — panelDuel: duel Panel (qa)
    MessageList zh; // obf: zh — friendsList: friends MessageList; reused as menu builder (wb)
    Panel zk; // obf: zk — panelLogin: login Panel (qa)

    // ----- Skills / stats / inventory / equipment -----
    int[] Aj; // obf: Aj — inventoryEquipped: inventory equip flags
    int[] Ak; // obf: Ak — equipBonusStats2: equip bonus
    int[] Bi; // obf: Bi — skillCurrent: skill current
    int[] Dd; // obf: Dd — skillStat: skill stat
    int[] Lc; // obf: Lc — skillXp: skill xp
    int[] Me; // obf: Me — skillBase: skill base
    int[] Qf; // obf: Qf — skillStat/tradeItems: skill stat / trade item ids (int[])
    int[] Vb; // obf: Vb — skillStat: skill stat
    int[] cg; // obf: cg — equipBonusStats3: equip bonus
    int[] jj; // obf: jj — skillStat/tradeItemCount: skill stat / trade item counts (int[])
    int[] oh; // obf: oh — equipBonusStats/equipBonusDisplay: equip bonus (int[18])
    int[] vf; // obf: vf — inventoryItems: inventory item ids
    int[] xe; // obf: xe — inventoryQty: inventory stack counts
    int[] zj; // obf: zj — skillStat: skill stat

    // ----- Audio -----
    StreamMixer hk; // obf: hk — soundMixer: audio mixer (ra)
    AudioChannel ni; // obf: ni — soundChannel: active audio voice (sa)

    // ----- Misc game-state scalars / flags / timers (obf-named) -----
    int Ag; // obf: Ag — game-state scalar
    private int Bc; // obf: Bc — game-state scalar
    int Be; // obf: Be — game-state scalar
    int Bf; // obf: Bf — game-state scalar
    int Bh; // obf: Bh — selectedItem: selected inv slot (-1=none)
    int Bj; // obf: Bj — socialDialogMode: 1=addFriend,2=PM,3=addIgnore
    int Cc; // obf: Cc — game-state scalar
    boolean Cd; // obf: Cd — state/feature flag
    int Cf; // obf: Cf — mouseButtonClick: 0/1/2 this tick
    private int Cg; // obf: Cg — game-state scalar
    String Cj; // obf: Cj — text buffer
    boolean Dc; // obf: Dc — state/feature flag
    int De; // obf: De — game-state scalar
    int Di; // obf: Di — game-state scalar
    private int Ee; // obf: Ee — game-state scalar
    private int Ef; // obf: Ef — game-state scalar
    int Eh; // obf: Eh — game-state scalar
    int[] Fc; // obf: Fc — int buffer/table
    private int Fd; // obf: Fd — game-state scalar
    boolean Fe; // obf: Fe — state/feature flag
    private int Fg; // obf: Fg — game-state scalar
    private int Fh; // obf: Fh — game-state scalar
    int Gf; // obf: Gf — game-state scalar
    int Gi; // obf: Gi — game-state scalar
    private boolean Hc; // obf: Hc — state/feature flag
    private int Hf; // obf: Hf — game-state scalar
    int Hi; // obf: Hi — game-state scalar
    boolean Hk; // obf: Hk — state/feature flag
    int Id; // obf: Id — game-state scalar
    int If; // obf: If — game-state scalar
    boolean Je; // obf: Je — state/feature flag
    int[] Jf; // obf: Jf — int buffer/table
    private String[] Kc; // obf: Kc — String table
    boolean Kd; // obf: Kd — state/feature flag
    int Ke; // obf: Ke — game-state scalar
    boolean Kg; // obf: Kg — state/feature flag
    boolean Kh; // obf: Kh — state/feature flag
    int Ki; // obf: Ki — game-state scalar
    int Lf; // obf: Lf — currentFloor: current floor/region offset
    String Lg; // obf: Lg — text buffer
    int Lk; // obf: Lk — game-state scalar
    private int Mg; // obf: Mg — game-state scalar
    private int Mh; // obf: Mh — game-state scalar
    boolean Mi; // obf: Mi — state/feature flag
    private int Nc; // obf: Nc — game-state scalar
    int Nh; // obf: Nh — game-state scalar
    int Nj; // obf: Nj — game-state scalar
    int[] Oc; // obf: Oc — int buffer/table
    boolean Oh; // obf: Oh — state/feature flag
    int Pf; // obf: Pf — game-state scalar
    boolean Pg; // obf: Pg — state/feature flag
    boolean Ph; // obf: Ph — state/feature flag
    boolean Pj; // obf: Pj — state/feature flag
    private int[] Pk; // obf: Pk — int buffer/table
    String Qd; // obf: Qd — text buffer
    private int Qe; // obf: Qe — game-state scalar
    int Qg; // obf: Qg — game-state scalar
    boolean Qk; // obf: Qk — state/feature flag
    int Rc; // obf: Rc — game-state scalar
    int Rd; // obf: Rd — game-state scalar
    private int Rh; // obf: Rh — game-state scalar
    int[] Rj; // obf: Rj — int buffer/table
    int Sb; // obf: Sb — game-state scalar
    int[] Sc; // obf: Sc — int buffer/table
    private int Sg; // obf: Sg — game-state scalar
    int Sh; // obf: Sh — game-state scalar
    private int Si; // obf: Si — game-state scalar
    private boolean Td; // obf: Td — state/feature flag
    int Tk; // obf: Tk — game-state scalar
    boolean Ub; // obf: Ub — state/feature flag
    String Uc; // obf: Uc — text buffer
    int[] Uf; // obf: Uf — int buffer/table
    int Ug; // obf: Ug — game-state scalar
    int Ui; // obf: Ui — game-state scalar
    int Uk; // obf: Uk — game-state scalar
    private boolean Vc; // obf: Vc — state/feature flag
    int Ve; // obf: Ve — game-state scalar
    int Vf; // obf: Vf — game-state scalar
    int Vg; // obf: Vg — game-state scalar
    boolean Vi; // obf: Vi — state/feature flag
    private int Vj; // obf: Vj — game-state scalar
    boolean Wk; // obf: Wk — state/feature flag
    private int Xd; // obf: Xd — activePanel: open panel id
    int[] Xe; // obf: Xe — int buffer/table
    boolean Xj; // obf: Xj — state/feature flag
    private int Yb; // obf: Yb — game-state scalar
    int Yc; // obf: Yc — game-state scalar
    boolean Yh; // obf: Yh — state/feature flag
    boolean Yi; // obf: Yi — state/feature flag
    int Zc; // obf: Zc — game-state scalar
    long[] Zd; // obf: Zd — mouseClickTimes: input timing ring (long[100])
    private int Zh; // obf: Zh — game-state scalar
    String Zj; // obf: Zj — text buffer
    private int ac; // obf: ac — game-state scalar
    int ad; // obf: ad — game-state scalar
    int[] ae; // obf: ae — int buffer/table
    int af; // obf: af — selectedSpell: selected spell (-1=none)
    String[] ah; // obf: ah — String table
    int ai; // obf: ai — game-state scalar
    private int[] ak; // obf: ak — int buffer/table
    int bc; // obf: bc — game-state scalar
    private int[] bf; // obf: bf — int buffer/table
    private int bh; // obf: bh — shop panel text-input control id (clean L255)
    int bj; // obf: bj — game-state scalar
    boolean[] bk; // obf: bk — flag array
    private int bl; // obf: bl — game-state scalar
    int ce; // obf: ce — game-state scalar
    int[] ci; // obf: ci — int buffer/table
    String cj; // obf: cj — text buffer
    int ck; // obf: ck — game-state scalar
    int cl; // obf: cl — inventorySize: inventory capacity
    int dc; // obf: dc — game-state scalar
    boolean dd; // obf: dd — state/feature flag
    int de; // obf: de — game-state scalar
    int[] df; // obf: df — int buffer/table
    private int dg; // obf: dg — game-state scalar
    int[] di; // obf: di — int buffer/table
    private String ec; // obf: ec — text buffer
    private int[] ee; // obf: ee — int buffer/table
    int el; // obf: el — game-state scalar
    boolean fd; // obf: fd — state/feature flag
    boolean ff; // obf: ff — state/feature flag
    int fg; // obf: fg — game-state scalar
    int fh; // obf: fh — game-state scalar
    boolean[] fi; // obf: fi — flag array
    int fj; // obf: fj — game-state scalar
    int gc; // obf: gc — game-state scalar
    private int[] gd; // obf: gd — int buffer/table
    int gh; // obf: gh — game-state scalar
    int[] gi; // obf: gi — int buffer/table
    int hi; // obf: hi — game-state scalar
    private boolean hj; // obf: hj — state/feature flag
    int id; // obf: id — game-state scalar
    int ii; // obf: ii — game-state scalar
    private int jc; // obf: jc — game-state scalar
    private int[] jd; // obf: jd — int buffer/table
    private int[] je; // obf: je — int buffer/table
    int ji; // obf: ji — game-state scalar
    int kc; // obf: kc — game-state scalar
    private int kd; // obf: kd — game-state scalar
    boolean ke; // obf: ke — state/feature flag
    int[] kf; // obf: kf — int buffer/table
    private int kg; // obf: kg — game-state scalar
    boolean ki; // obf: ki — state/feature flag
    int lc; // obf: lc — game-state scalar
    int le; // obf: le — game-state scalar
    boolean lh; // obf: lh — state/feature flag
    private int mc; // obf: mc — game-state scalar
    boolean md; // obf: md — state/feature flag
    int mf; // obf: mf — game-state scalar
    boolean mh; // obf: mh — showCloseWindow: close-window dialog visible
    int nc; // obf: nc — game-state scalar
    boolean ne; // obf: ne — state/feature flag
    private int[] nf; // obf: nf — int buffer/table
    int nh; // obf: nh — game-state scalar
    int nj; // obf: nj — game-state scalar
    private int oc; // obf: oc — game-state scalar
    int[] oe; // obf: oe — int buffer/table
    int[] of; // obf: of — int buffer/table
    private int[] pe; // obf: pe — int buffer/table
    int pg; // obf: pg — game-state scalar
    private int pj; // obf: pj — game-state scalar
    int pk; // obf: pk — game-state scalar
    int qc; // obf: qc — game-state scalar
    private int qe; // obf: qe — game-state scalar
    // SPLIT-FIELD FIX (class b): obf `qg` (the 0=login/1=in-game game-state scalar) is now the single
    // field `screenMode` (declared below). The duplicate `qg` and `loggedIn` ints that aliased it were
    // removed; they desynced (resetGameState wrote qg, the main loop read screenMode) -> login never
    // advanced to the in-game render.
    int qj; // obf: qj — game-state scalar
    int rc; // obf: rc — game-state scalar
    String re; // obf: re — text buffer
    private int rf; // obf: rf — game-state scalar
    int rh; // obf: rh — game-state scalar
    int rk; // obf: rk — game-state scalar
    boolean se; // obf: se — state/feature flag
    int sg; // obf: sg — game-state scalar
    int sh; // obf: sh — game-state scalar
    int si; // obf: si — game-state scalar
    int sj; // obf: sj — game-state scalar
    int sk; // obf: sk — game-state scalar
    private int[] tf; // obf: tf — int buffer/table
    int tg; // obf: tg — game-state scalar
    int[] th; // obf: th — int buffer/table
    private int tj; // obf: tj — game-state scalar
    private int ud; // obf: ud — game-state scalar
    private boolean ue; // obf: ue — state/feature flag
    private int[] uf; // obf: uf — int buffer/table
    private int ug; // obf: ug — game-state scalar
    int ui; // obf: ui — game-state scalar
    boolean uk; // obf: uk — state/feature flag
    boolean vd; // obf: vd — state/feature flag
    String ve; // obf: ve — text buffer
    int vg; // obf: vg — game-state scalar
    int vj; // obf: vj — game-state scalar
    int wj; // obf: wj — game-state scalar
    int wk; // obf: wk — game-state scalar
    int xg; // obf: xg — game-state scalar
    int xh; // obf: xh — game-state scalar
    int[] xi; // obf: xi — int buffer/table
    int[] xj; // obf: xj — int buffer/table
    int xk; // obf: xk — game-state scalar
    private int yj; // obf: yj — game-state scalar
    int[] zc; // obf: zc — int buffer/table
    private boolean zf; // obf: zf — state/feature flag
    int zg; // obf: zg — game-state scalar

    // ----- Renamed aliases (same underlying field referenced by a readable name in some
    //       method bodies AND by its obf name in others; both identifiers must exist) -----
    MessageList chatList; // renamed alias of obf He — chatList: chat MessageList (wb)
    private int fatigueControlId; // renamed alias of obf Qi — game-state scalar
    GameModel[] objectModels; // renamed alias of obf kh — objectModels: scene object GameModels (ca[1000])
    private String password; // renamed alias of obf wh — password: account password
    private int serverMsgControlId; // renamed alias of obf td — game-state scalar
    private int tradeItemMenu; // renamed alias of obf ? — game-state scalar
    private int tradeRecipientAccepted; // renamed alias of obf ? — game-state scalar
    String username; // renamed alias of obf Xf — username: account name
    private int[] walkPathX; // obf Rg — route() X-waypoint buffer (int[8000]); allocated in init
    private int[] walkPathY; // obf pf — route() Y-waypoint buffer (int[8000]); allocated in init

    // ----- Renamed game-state fields (readable names used throughout the bodies; type
    //       inferred from use / recovered obf in trailing note) -----
    private int accountFlags; // scalar
    private int appearance2Colour; // scalar
    private int appearanceBodyGender; // scalar
    private int appearanceBottomColour; // scalar
    private int appearanceGender; // scalar
    private int appearanceHairColour; // scalar
    private int appearanceHead; // scalar
    private int appearanceSkinColour; // scalar
    private int appearanceTopColour; // scalar
    private int charDesignAccept; // scalar
    private int charDesignBottomLeft; // scalar
    private int charDesignBottomRight; // scalar
    private int charDesignGenderLeft; // scalar
    private int charDesignGenderRight; // scalar
    private int charDesignHairLeft; // scalar
    private int charDesignHairRight; // scalar
    private int charDesignHeadLeft; // scalar
    private int charDesignHeadRight; // scalar
    private int charDesignSkinLeft; // scalar
    private int charDesignSkinRight; // scalar
    private int charDesignTopLeft; // scalar
    private int charDesignTopRight; // scalar
    private int[] charHairColours; // char-design hair palette
    private int[] charSkinColours; // char-design skin palette
    private int[] charTopBottomColours; // char-design top/bottom palette
    MessageList ignoreList; // ignore list (=Wf) (obf Wf)
    private int[] inventoryItemStackCount; // int array
    private int inventoryItemsCount; // scalar
    private int loggedInState; // scalar
    private int loginButton; // panel control id
    private int loginCancelButton; // panel control id
    private Panel loginEntryPanel; // Panel
    private int loginOkButton; // panel control id
    private int loginPort; // scalar
    private int loginPortAlt; // scalar
    private int loginPromptControl; // panel control id
    private int loginScreenMode; // scalar
    private boolean autoLoginDone; // M3 test-harness: one-shot env-driven auto-login latch
    private boolean loginScreenRedraw; // flag
    private int loginTitleControl; // panel control id
    private Panel loginWelcomePanel; // Panel
    boolean membersServer; // obf Pg — boolean (clean L194)
    private boolean membersWorld; // obf Pg — boolean (clean L194)
    private int menuOptionCount; // scalar
    private String[] menuOptions; // right-click menu option text
    private Panel messagePanel; // Panel
    private int messageTabSelected; // scalar
    private int moderatorLevel; // scalar
    private Panel panelShop; // shop panel (=yd) (obf yd)
    private int passwordField; // panel control id
    private int referId; // scalar
    // (removed orphan `regionX`/`regionY` — never-assigned split-field dups of the real region origin
    //  Qg/zg; walkTo/walkToAction were reading these 0s when building the absolute walk-start coords,
    //  so every WALK_TO_POINT/ENTITY packet carried a start tile ~144 tiles off and the server dropped it.)
    private int screenHeight; // scalar
    private int screenWidth; // scalar
    private String serverHost; // world host (=Dh) (obf Dh)
    private int spriteItem; // scalar
    private int spriteMedia; // scalar
    private int systemUpdateTimer; // scalar
    private int tabActivityChat; // scalar
    private int tabActivityGame; // scalar
    private int tabActivityPrivate; // scalar
    private int tabActivityQuest; // scalar
    private int tabChat; // scalar
    private int tabPrivate; // scalar
    private int tabQuest; // scalar
    private int tradeConfirmShown; // scalar
    int tradeItemsCount; // scalar
    private int[] tradeRecipientItemCount; // int array
    private int[] tradeRecipientItems; // partner trade item ids
    private int tradeRecipientItemsCount; // scalar
    private String tradeRecipientName; // trade partner name
    private int usernameField; // panel control id
    private boolean veteranWorld; // obf cf — boolean (clean L205)
    private int worldFullTimeout; // scalar
    private int worldIndex; // scalar (obf Vh)

    // ===== Residual game-state fields (readable names; types inferred from usage) =====
    private int animationCount;
    private int[] animationHasA;
    private int[] animationHasF;
    private String[] animationNames;
    private int[] animationNumber;
    private int[] animationSomething;
    private int appearanceCount;
    private int[] appearanceFlags;
    private boolean appletMode;
    private int armorBonuses;
    private int audioFactory;
    private int audioInstance;
    private int audioQueue;
    private int[] bankItems;
    private int bankOfferItems;
    private int bankSelectedSlot;
    private int bankSlotItems;
    private int cacheFile;
    private int cacheLimit;
    // NOTE: si, Td, kg, Si, Wc, ug, Be, oc, ac were duplicated here under descriptive
    // camera-* names by an earlier pass; they are the SAME obf fields declared above
    // (the camera-follow/rotate/zoom logic wrote these copies while the render read the
    // originals, so the camera stayed at world-origin 0,0). Unified to the obf
    // declarations; only the two genuinely-distinct idle-drift fields remain here.
    private int cameraRotationTime;        // obf: oj
    private int cameraRotationYIncrement;  // obf: Ok
    private int charCount;
    private int charDesignWobbleX;
    private int charDesignWobbleY;
    private int chatColors;
    private String chatEntry;
    private String chatInputLine;
    private int chatInputMode;
    private String chatInputUser;
    private int combatStyleIndex;
    private int combatStyleNames;
    private int combatTimeout;
    private int controlListChat;
    private int controlListInput;
    private int controlListMagic;
    private int controlListPrivate;
    private int controlListQuest;
    private int deathScreenTimeout;
    private int defaultFont;
    private int dialogItemId;
    private int dialogItemId2;
    private int displayDepth;
    private boolean displayEnabled;
    private int displayName;
    int drawListCount;
    int[] drawListCurrent;
    int[] drawListIds;
    int drawListSize;
    int[] drawListY;
    int[] drawListYShadow;
    private int duelOwnItems;
    private int duelTheirItems;
    private int[] experienceTable;
    private boolean fatalLoadError;
    private int fatigueColors;
    private boolean fogOfWar;
    private int fontMetrics;
    private int fpsCapBackground;
    private int fpsCapForeground;
    int friendListCount;
    String[] friendListFormerNames;
    String[] friendListNames; // obf ua.h — String[] (clean)
    int[] friendListOnline;
    String[] friendListWorlds;
    private int gameHeight;
    private int gameWidth;
    private int glyphBase;
    private int groundItemCount;
    private int[] groundItemX;
    private int[] groundItemY;
    // hasPainted: inherited boolean from GameShell (obf hasPainted) — NOT redeclared
    int ignoreListCount;
    String[] ignoreListDisplayNames; // obf ia.a — String[] (clean)
    private String ignoreListEntry;
    String[] ignoreListFormerNames;
    String[] ignoreListNames; // obf l.c — String[] (clean)
    int[] ignoreListWorlds;
    private boolean inputDialogConfirmed;
    private int inputDialogHeight;
    private String[] inputDialogLines; // String[] (clean)
    private boolean inputDialogMask;
    private int inputDialogType;
    private int inputDialogWidth;
    String inputLine; // obf Cb — String (clean)
    private int instance;
    private int[] inventoryGroundOverlay; // obf of — int[] (clean L398)
    private boolean isApplet;
    private boolean isDoubleSided;
    private boolean isFreeWorld;
    boolean isMember;
    private boolean isSleeping;
    private int itemColors;
    long lastActionTime; // obf Wi — long (clean L216)
    private int localRegionX;
    private int localRegionY;
    private int[] localX;
    private int[] localY;
    // (removed: `loggedIn` — was a desynced alias of obf qg; unified into `screenMode`. See class-b fix above.)
    private int loginAnimFrame;
    private int loginFrameCount;
    private int loginStage;
    // (was `loginTimeout` — actually obf `ac`, the camera zoom; unified to the `ac` field above)
    private int logoutTimeout;
    private int magicLoc;
    int menuHeight;
    boolean menuOpenFlag;
    private int menuOptionActions;
    String[] menuOptionList; // obf od — String[] (clean L378)
    private int menuOptionStrings;
    private int menuOptionTargets;
    String menuTitle; // obf e — String (clean)
    int menuWidth;
    int menuX;
    // (removed: `messageHistoryTimeout` — was a desynced, never-allocated alias of the static
    //  ImageLoader.scratchBuf[100] (obf pa.g); the decay loop now uses ImageLoader.scratchBuf. class-b fix.)
    private int minimapRandom1;
    private int minimapRandom2;
    private int modelCount;
    private int mouseButton;
    // mouseButtonDown: inherited int from GameShell (obf Bb) — NOT redeclared
    private int mouseButtonDownTime;
    private int mouseButtonItemCountIncrement;
    private int mouseButtonMode;
    private int mouseClickTimes;
    private int mouseLastButton;
    private int nodeId;
    private int npcCount;
    private int npcCountView;
    private int npcId;
    private int npcs;
    private int npcsLastCount;
    private int objectAnimationClaw;
    private int objectAnimationCount;
    private int objectAnimationFire;
    private int objectAnimationTorch;
    private int objectCount;
    private int objectTileX;
    private int objectTileY;
    private int objectTileZ;
    private int[] objectX;
    private int[] objectY;
    private int originY;
    private boolean outOfMemory;
    private Panel panelMagic; // obf Mc — Panel/qa (clean L404)
    private Panel panelMessageTabs; // Panel/qa (clean)
    private String passwordInput;
    private boolean playerAlive;
    private boolean autoAppearanceSent; // BOOT HOOK: one-shot guard for the RSC_AUTO_APPEARANCE bring-up hook
    private int fbufferInGameFrames; // BOOT HOOK: counts in-game render frames for the RSC_FBUFFER_DUMP hook
    private boolean fbufferDumped;   // BOOT HOOK: one-shot guard for the RSC_FBUFFER_DUMP hook
    private int fbufferDiag;         // BOOT HOOK: drawGameFrame entry counter for the RSC_FBUFFER_DUMP diagnostic
    private int autoWalkFrames;      // BOOT HOOK: in-game frame counter for the RSC_AUTO_WALK bring-up hook
    private int autoWalkStep;        // BOOT HOOK: index into the auto-walk tour (cycles the destination tile)
    private int autoTabFrames;       // BOOT HOOK: in-game frame counter for the RSC_AUTO_TABS hook
    private int autoTabStep;         // BOOT HOOK: index into the UI-tab cycle
    private int players;
    private String pmInput;
    private int portA;
    private int portB;
    private boolean[] prayerOn; // obf bk — boolean[] (clean L400)
    boolean privacyChatOn;
    boolean privacyMembersOn;
    boolean privacyTradeOn;
    private int questCompleteFlags;
    private int questNames;
    private int screenMode;
    MessageList scrollMessageList; // obf Wf — wb/MessageList (clean L401)
    private int selectedItemInventoryIndex;
    private String serverMessage;
    private int serverUpdateTick;
    private int sessionBytes;
    private int[] shopSelectedItemId;
    private int[] shopSelectedItemPrice;
    private int shopSelectedSlot;
    private boolean showDialogDuel;
    private int showDialogReportAbuseStep;
    private boolean showDialogTrade;
    boolean showMenuBorder; // obf Bd — boolean (clean L368)
    private int showUiTab;
    private int[] skillBase;
    private int skillBaseLevels;
    private int[] skillCurrent;
    private int skillCurrentLevels;
    private int skillNamesLong;
    private int skillNamesShort;
    private String[] Vk; // obf Vk — skillNamesShort String[] (clean L365)
    private String[] Ej; // obf Ej — skillNamesLong String[] (clean L325)
    private String[] Te; // obf Te — questNames String[] (clean L359)
    private String[] Ld; // obf Ld — combatStyleNames String[] (clean L475)
    private int skillXp;
    private int skillXpAccum;
    private int skillXpDeltas;
    private int skillXpGained;
    private boolean sleepWordDelay;
    private int sleepWordDelayTimer;
    private String sleepingStatusText;
    private int soundChannel;
    private int spriteBaseBubbles;
    private int spriteBaseChars;
    private int spriteBaseGroundItems;
    private int spriteBaseInventory;
    private int spriteBaseNpcs;
    private int spriteBaseObjects;
    private int spriteBaseTextures;
    private int spriteBaseWalls;
    private int spriteSheetCount;
    private int staticRef;
    private String submittedPmInput;
    private int systemUpdate;
    private int tabMagicPrayer;
    private int teleportBubbleCount;
    private int[] teleportBubbleTime;
    private int[] teleportBubbleType;
    private int[] teleportBubbleX;
    private int[] teleportBubbleY;
    private String tempInputString;
    private long tickMarker; // obf ze — long (clean L57)
    private int ticksPerFrame;
    private int tradeOwnItems;
    private int tradeQueuedAction;
    private int tradeTheirItems;
    private int view;
    private int[] wallModelX;
    private int[] wallModelZ;
    private int weaponBonuses;


    // ===== Residual obf-named fields (types from clean decompile client.java) =====
    private int Ai;
    private int Ce;
    private int Df;
    private int[] Dg;
    private int Dj;
    private boolean[] Ed;
    private int Eg;
    private int Fi;
    private int Ge;
    private int Jg;
    private int Ji;
    private int Kj;
    int[] Kk;
    private int Lh;
    private int Mj;
    private int Ne;
    private int Of;
    private int[] Og;
    private int Oj;
    private int[] Pc;
    private int Re;
    // (removed orphan `int[] Rg` — was the unallocated obf-name dup of walkPathX; see SPLIT-FIELD FIX in init)
    private int Rk;
    private boolean[] Sj;
    private int[][] Tg;
    private int Ti;
    private int Ud;
    private int Vd;
    private int Wc;
    private int Wg;
    private int[] Wh;
    private int Xc;
    private boolean Xh;
    int Yd;
    private boolean Yk;
    private int Ze;
    private int dj;
    private int dk;
    private int ed;
    private int[] ei;
    private int ek;
    private int fl;
    private int gl;
    private int hh;
    private int jk;
    private int ld;
    private int lk;
    int nk;
    private String[] od;
    // (removed orphan `int[] pf` — was the unallocated obf-name dup of walkPathY; see SPLIT-FIELD FIX in init)
    private int pi;
    int qd;
    private int qk;
    private int sd;
    private int[] sf;
    private int uc;
    int[] uj;
    private boolean vk;
    private int wg;
    private int zd;
    private int zi;



    // =========================================================================
    // ===== bootstrap =====
    // =========================================================================
// Methods in group "bootstrap" RE-AUDITED against the CLEAN Vineflower base
// (decompiled/normalized-clean/client.java). Microsoft J++ RSC rev ~233-235.
//
// Obfuscation stripped:
//   - Opaque predicate (client.OPAQUE_FALSE / vh) — always false; dead branches removed.
//   - Per-method profiling counter increments (++Pd; ++Gd; ...).
//   - try/catch(RuntimeException){throw ErrorHandler.wrap(e, il[label])} wrappers unwrapped.
//   - Anti-tamper dummy-param guards (if(p!=const)...) + junk modulo expressions removed.
//   - The ~x>~y / ~x==const sign idioms rewritten to ordinary comparisons.
//
// Naming honours docs/NAMING.md (CORRECTED): k=World, lb=Scene.
//   Therefore Ek (type lb=Scene) -> `scene`, Hh (type k=World) -> `world`.
//   (The MUDCLIENT_SKELETON field table had these two swapped; NAMING.md wins.)
//
// All il[] string indices below were decoded from the real XOR pool and are
// annotated with their true decoded text. The OLD part file's il[] comments
// were taken from the defective decompilation and were almost entirely wrong.
//
// Class context: package client; class Mudclient extends GameShell (e)
// STRINGS = client.il  (XOR-decoded string pool)

    // -------------------------------------------------------------------------
    /** Standalone entry point: parse args, select audio backend, create and start client frame.
     *  obf: static void main(String[])   obf-label: il[313]="client.main(" */
    public static final void main(String[] args) {
        try {
            // args[0] = nodeid (int), args[1] = mode string, args[2..] = optional flags.

            // GameShell.chatCipher (e.i, type v=ChatCipher) <- la.b = ClientRuntimeException.LOCAL_CIPHER.
            GameShell.chatCipher = ClientRuntimeException.LOCAL_CIPHER;            // e.i = la.b

            // Parse the node/world id.  aa.l = BZip.entityLimit (static int).
            BZip.entityLimit = Integer.parseInt(args[0]);         // aa.l

            // Select audio/queue backend from the mode string.
            //   il[312]="live"    -> AudioMixer            (eb.e)
            //   il[317]="rc"      -> RecordLoader          (f.b)   [see note]
            //   il[318]="wip"     -> SurfaceImageProducer  (fb.h)
            // Control flow (from CFR) falls through: matching "wip" sets f.b then
            // falls into fb.h then eb.e; matching "rc" sets fb.h then eb.e;
            // matching "live" sets only eb.e. Net effect of the fall-through chain:
            //   "live"          -> db.f = eb.e
            //   "rc"  (il[317]) -> db.f = fb.h
            //   "wip" (il[318]) -> db.f = f.b
            if (args[1].equals(STRINGS[312])) {              // "live"
                LinkedQueue.errorHandler = AudioMixer.errorHandlerTag;            // db.f = eb.e
            } else if (args[1].equals(STRINGS[317])) {       // "rc"
                LinkedQueue.errorHandler = SurfaceImageProducer.errorHandler; // db.f = fb.h
            } else if (args[1].equals(STRINGS[318])) {       // "wip"
                LinkedQueue.errorHandler = RecordLoader.unusedErrorHandler;     // db.f = f.b
            }
            // else: leave db.f unchanged.

            // Construct the main client instance (applet flag false for standalone).
            Mudclient client = new Mudclient();
            client.isApplet = false;                         // hj = false

            // Parse optional flags from args[2..].
            for (int i = 2; i < args.length; i++) {
                if (args[i].equals(STRINGS[316])) {          // "members"
                    client.Kd = true;            // Pg
                }
                if (args[i].equals(STRINGS[315])) {          // "veterans"
                    client.isFreeWorld = true;               // cf
                }
            }

            // Create the AWT frame and start the game thread.
            //   il[314]="local.runescape.com"  (title/host param A)
            //   il[319]="classic"              (title/host param B)
            //   32 + db.f.a = frame/thread priority offset
            //   aa.l + 7000 = server port; (byte)112 flags; fa.d = display depth
            //   client.Wd = 512 (width); client.Oi - -12 = 334 + 12 = 346 (height)
            try {
                client.startApplication(
                    false,                                       // standalone (not applet)
                    STRINGS[314],                                // "local.runescape.com"
                    32 + LinkedQueue.errorHandler.instanceValue,// 32 + db.f.a
                    STRINGS[319],                                // "classic"
                    BZip.entityLimit + 7000,                          // aa.l + 7000 (port)
                    (byte)112,
                    ClientIOException.BUILD_REVISION,              // fa.d
                    client.screenWidth,                          // Wd = 512
                    client.screenHeight + 12                     // Oi - -12 = 346
                );
                client.ticksPerFrame = 10;                       // Q = 10
            } catch (Exception e) {
                Utility.reportError(0x1FFFFF, e, null);           // mb.a(2097151, e, null)
            }
        } catch (RuntimeException e) {
            // obf: throw i.a(e, il[313] + (args!=null ? il[29] : il[31]) + ')')
            throw ErrorHandler.wrap(e, STRINGS[313] + (args != null ? STRINGS[29] : STRINGS[31]) + ')');
        }
    }

    // -------------------------------------------------------------------------
    /** Applet init: read nodeid/modewhat/modewhere params, size window, kick off loading.
     *  obf: void init()   obf-label: il[183]="client.init()" */
    @Override
    public final void init() {
        try {
            // il[182]="nodeid" -> BZip.entityLimit (aa.l)
            BZip.entityLimit = Integer.parseInt(this.getParameter(STRINGS[182]));

            // il[185]="modewhere" here doubles as the font-size param; ub.a(size,(byte)24)
            // builds a NameTable reference into e.i (GameShell.chatCipher, type v=ChatCipher).
            GameShell.chatCipher = NameTable.findById(
                Integer.parseInt(this.getParameter(STRINGS[185])), (byte)24
            );                                               // e.i = ub.a(...)
            if (GameShell.chatCipher == null) {
                GameShell.chatCipher = Surface.decoyStringHolder; // e.i = ua.E
            }

            // il[184]="modewhat" -> u.a(false,id) picks the LinkedQueue/audio factory (db.f, type i=ErrorHandler).
            LinkedQueue.errorHandler = StringCodec.lookupRegisteredEntry(
                false, Integer.parseInt(this.getParameter(STRINGS[184]))
            );                                               // db.f = u.a(false, ...)
            if (LinkedQueue.errorHandler == null) {
                LinkedQueue.errorHandler = AudioMixer.errorHandlerTag; // db.f = eb.e
            }

            // Start the GameShell loader thread:
            //   super.a(Oi+12, fa.d, db.f.a+32, 2, Wd)
            super.startApplet(
                this.screenHeight + 12,                          // Oi + 12
                ClientIOException.BUILD_REVISION,                  // fa.d
                LinkedQueue.errorHandler.instanceValue + 32,    // db.f.a + 32
                2,
                this.screenWidth                                 // Wd
            );
        } catch (RuntimeException e) {
            throw ErrorHandler.wrap(e, STRINGS[183]);           // il[183]="client.init()"
        }
    }

    // -------------------------------------------------------------------------
    /** Constructor: allocate all state arrays and set field defaults.
     *  obf: client()  (no catch wrapper) */
    public Mudclient() {
        super();

        // --- Network ---
        mg = new BitBuffer(5000);    // mg

        // --- Basic scalars / cursors ---
        Nc = 0;
        Vg = 0;
        qd = 9;                                  // camera zoom default
        Zd = new long[100];                      // Zd (mouseClickTimes)
        Wc = 0;
        Oj = 0;
        loginAnimFrame = 0;                      // jk
        ac = 550;                      // ac
        isApplet = true;                         // hj
        yj = -1;
        De = 0;
        npcCount = 0;                            // If
        Xh = false;                              // domain-lock tripped flag
        si = 1;
        xh = 0;
        Ce = 0;
        oc = 0;                     // oc (idle drift, oc)
        bl = -1;
        qk = 0;
        Sg = -1;
        dc = 0;
        ug = 128;                                // ug seed
        Ug = 128;                                // ug seed (2nd axis)
        yg = -1;
        Kk = new int[8192];
        walkPathY = new int[8000];   // SPLIT-FIELD FIX: obf pf == walkPathY (route() Y waypoints); alloc went to the orphan dup
        Kd = false;                  // Pg
        Cf = 0;
        loginStage = 0;                          // Zb
        screenWidth = 512;                       // Wd
        kg = 0;
        bc = -1;
        screenHeight = 334;                      // Oi
        mouseButtonMode = 2;                     // eg: dual-purpose obf field — default 2.
                                                 //   In startGame it doubles as the Be
                                                 //   idle-drift increment; setMouseButtonMode sets it to 77.
        rc = 0;
        qe = 0;
        isFreeWorld = false;                     // cf
        nk = 0;
        Yc = 0;
        outOfMemory = false;                     // Ue
        Si = 0;
        Ki = 0;
        fatalLoadError = false;                  // Vc
        pj = 0;
        uj = new int[8192];
        sk = 0;
        screenMode = 0;                          // qg: 0=login, 1=in-game
        zf = false;
        fpsCapBackground = 40;                   // nc
        cameraRotationYIncrement = 2;            // Ok (idle drift step for oc)
        cameraRotationTime = 0;                  // oj
        Fd = 0;
        ui = 0;
        Ag = 0;

        // --- Entities in view ---
        Zg = new GameCharacter[500];             // Zg (players)
        Be = 0;                     // Be (idle drift, Be)
        npcCountView = 0;                        // Mg
        rg = new GameCharacter[500];    // rg
        We = new GameCharacter[4000];     // We
        tj = 0;
        walkPathX = new int[8000];   // SPLIT-FIELD FIX: obf Rg == walkPathX (route() X waypoints); alloc went to the orphan dup
        worldIndex = 0;                          // Vh
        wi = new GameCharacter();       // wi
        bg = new int[1500];
        ci = new int[256];
        Cd = false;
        Qd = null;                         // Zj
        Rj = new int[256];
        dj = 0;

        // --- Chat colour palette (15 RGB entries) — verified against clean ei[] ---
        ei = new int[]{                          // ei (chatColors)
            0xFF0000,  // red
            0xFF8000,  // orange
            0xFFE000,  // yellow
            0xA0E000,  // yellow-green
            0x00E000,  // green
            0x008000,  // dark green
            0x00A080,  // teal
            0x00B0FF,  // sky blue
            0x0080FF,  // blue
            0x0030F0,  // blue-violet (12528)
            0xE000E0,  // magenta
            0x303030,  // dark grey
            0x604000,  // brown
            0x805000,  // tan
            0xFFFFFF   // white
        };

        // --- Entity/object index arrays ---
        Zf = new int[5000];
        oe = new int[50];
        xk = 0;
        Uf = new int[8];
        Qe = 0;
        ae = new int[256];
        lh = false;
        zj = new int[14];                        // zj (skillXp)
        el = 0;
        Ui = 0;
        tf = new int[50];
        Bc = 0;
        zd = 0;
        Fg = 0;
        hi = 0;
        Zc = -1;

        // --- Item colour palette (10 RGB entries) — verified against clean Dg[] ---
        Dg = new int[]{                          // Dg (itemColors)
            0xFFC030, 0xFFA040, 0x805030, 0x604020, 0x303030,
            0xFF6020, 0xFF4000, 0xFFFFFF, 0x00FF00, 0x00FFFF
        };

        oh = new int[18];                        // equipment bonus stats

        // --- Fatigue/sleep-bar colours (5 entries) — verified against clean Wh[] ---
        Wh = new int[]{                          // Wh (fatigueColors)
            0xECDED0, 0xCCB366, 0xB38C40, 0x997326, 0x906020
        };

        ce = 0;
        objectModels = new GameModel[1000];      // kh
        Xe = new int[256];
        Pf = 0;
        Mi = false;

        // --- Head-slot draw-layer order (8 styles) ---
        Og = new int[]{0, 0, 0, 0, 0, 1, 2, 1};

        skillBase = new int[14];                 // Vb

        // --- Equipment-slot layer ordering tables (8 configs x 12 slots) ---
        Tg = new int[][] {
            {11, 2, 9, 7, 1, 6, 10, 0, 5, 8, 3, 4},
            {11, 2, 9, 7, 1, 6, 10, 0, 5, 8, 3, 4},
            {11, 3, 2, 9, 7, 1, 6, 10, 0, 5, 8, 4},
            {3, 4, 2, 9, 7, 1, 6, 10, 8, 11, 0, 5},
            {3, 4, 2, 9, 7, 1, 6, 10, 8, 11, 0, 5},
            {4, 3, 2, 9, 7, 1, 6, 10, 8, 11, 0, 5},
            {11, 4, 2, 9, 7, 1, 6, 10, 0, 5, 8, 3},
            {11, 2, 9, 7, 1, 6, 10, 0, 5, 8, 4, 3}
        };

        jd = new int[50];
        experienceTable = new int[99];           // ti (filled in loadGameConfig)
        Ng = new int[500];
        wg = 2;
        skillCurrent = new int[14];              // Me
        wj = 0;
        nf = new int[50];
        Rd = -1;
        Kd = false;
        Wg = 8;
        th = new int[8];
        zi = 0;

        // --- Skill names (short set) — il[48..580] ---
        Vk = new String[] {                      // Vk (skillNamesShort)
            STRINGS[48],  STRINGS[543], STRINGS[546], STRINGS[562], // Attack, Defense, Strength, Hits
            STRINGS[575], STRINGS[570], STRINGS[16],  STRINGS[548], // Ranged, Prayer, Magic, Cooking
            STRINGS[557], STRINGS[591], STRINGS[565], STRINGS[550], // Woodcut, Fletching, Fishing, Firemaking
            STRINGS[559], STRINGS[569], STRINGS[560], STRINGS[529], // Crafting, Smithing, Mining, Herblaw
            STRINGS[567], STRINGS[580]                              // Agility, Thieving
        };

        chatInputLine = "";                      // Lg
        Ee = 0;
        ke = false;
        Gi = 48;
        bf = new int[50];
        ff = false;
        Sc = new int[50];
        gc = 0;
        fg = 0;                                  // fg (distinct field; NOT screenMode)
        cg = new int[18];                        // cg (weaponBonuses)
        Mh = 0;
        ee = new int[50];
        fd = false;
        gi = new int[50];
        te = new GameCharacter[5000];  // te
        Pg = true;                 // Bd

        // --- Character animation slot order (8 entries) ---
        Pc = new int[]{0, 1, 2, 1, 0, 0, 0, 0};

        Ef = 0;
        Hk = false;
        hg = new GameModel[1500];        // hg
        Ve = 0;
        vj = 0;
        Zh = 0;
        chatInputUser = "";                      // cj
        ve = null;
        id = 0;

        // --- Scroll-frame step table ---
        sf = new int[]{0, 1, 2, 1};

        passwordInput = "";                      // wh
        mf = 0;

        // --- Skill names (long set) — same pool, idx 553 ("Woodcutting") vs short's 557 ---
        Ej = new String[] {                      // Ej (skillNamesLong)
            STRINGS[48],  STRINGS[543], STRINGS[546], STRINGS[562],
            STRINGS[575], STRINGS[570], STRINGS[16],  STRINGS[548],
            STRINGS[553], STRINGS[591], STRINGS[565], STRINGS[550],
            STRINGS[559], STRINGS[569], STRINGS[560], STRINGS[529],
            STRINGS[567], STRINGS[580]
        };

        vi = new int[256];
        ii = 0;
        vd = false;
        di = new int[256];
        Yh = false;
        od = null;
        de = 0;
        rd = new GameModel[500];      // rd
        Yk = true;
        qj = 0;
        Vf = 0;
        pe = new int[50];
        Ni = new int[5000];
        Le = new int[5000];
        gl = 0;
        Fc = new int[5];
        xi = new int[8];                         // xi (tradeOwnItems)
        Gj = new int[5000];
        se = false;
        Wk = false;
        Uh = null;                               // Uh (sessionBytes)
        je = new int[50];
        af = -1;
        Qf = new int[14];                        // Qf (skillXpGained)
        Ph = false;
        Jf = new int[256];
        ye = new int[1500];                      // ye (objectTileX)
        Oc = new int[50];                        // menu/option scratch array
        Sj = new boolean[500];                   // Sj (questCompleteFlags)
        Ff = new GameCharacter[500];             // Ff (npcs)
        Ji = 0;
        Lk = 0;
        bj = 0;
        Ed = new boolean[1500];
        serverMessage = "";                      // Cj
        eh = 0;
        Dc = false;
        bk = new boolean[50];
        Yb = 0;

        // --- Quest names (50 entries, il[529..598]) ---
        Te = new String[] {                      // Te (questNames)
            STRINGS[596], STRINGS[542], STRINGS[554], STRINGS[598],
            STRINGS[586], STRINGS[573], STRINGS[584], STRINGS[590],
            STRINGS[541], STRINGS[551], STRINGS[545], STRINGS[535],
            STRINGS[561], STRINGS[547], STRINGS[566], STRINGS[589],
            STRINGS[568], STRINGS[540], STRINGS[555], STRINGS[594],
            STRINGS[595], STRINGS[583], STRINGS[536], STRINGS[588],
            STRINGS[579], STRINGS[544], STRINGS[587], STRINGS[578],
            STRINGS[564], STRINGS[534], STRINGS[585], STRINGS[572],
            STRINGS[556], STRINGS[577], STRINGS[576], STRINGS[538],
            STRINGS[582], STRINGS[531], STRINGS[539], STRINGS[563],
            STRINGS[593], STRINGS[537], STRINGS[533], STRINGS[549],
            STRINGS[558], STRINGS[574], STRINGS[592], STRINGS[530],
            STRINGS[597], STRINGS[532]
        };

        Ti = 0;
        Pj = false;
        rk = 0;
        wk = -1;
        ue = false;
        Id = 0;
        Vd = 0;
        fpsCapForeground = 30;                   // cl
        serverUpdateTick = 1;                    // Sf
        zc = new int[8];                         // zc (tradeTheirItems)
        xg = 0;
        hf = 0;
        Nh = 0;
        chatEntry = "";                          // ec
        Kh = true;
        Kg = false;
        Bi = new int[14];                        // Bi (skillCurrentLevels)
        Ak = new int[18];                        // Ak (armorBonuses)
        fi = new boolean[50];
        Di = -1;
        uf = new int[50];
        md = false;
        combatStyleIndex = 14;                   // Lh
        Jd = new int[500];
        ai = 0;
        ki = false;
        fj = 0;
        xj = new int[8];                         // xj (duelOwnItems)
        yk = new int[500];
        Sb = 0;
        Kc = new String[50];                     // Kc (menuOptionStrings)
        Hj = new int[500];
        gd = new int[50];                        // gd (menuOptionActions)
        xe = new int[35];            // xe (inventoryQty)
        vk = false;
        Hc = false;
        Lc = new int[14];                        // Lc (skillBaseLevels)

        // --- Equipment-bonus stat labels (NOT combat-style names) ---
        //     il[552]="Armour", il[571]="WeaponAim", il[581]="WeaponPower",
        //     il[16]="Magic", il[570]="Prayer"
        Ld = new String[] {                      // Ld (combatStyleNames)
            STRINGS[552], STRINGS[571], STRINGS[581], STRINGS[16], STRINGS[570]
        };

        uk = false;
        Bj = 0;
        nj = -1;
        ne = false;
        username = "";                           // Xf
        of = new int[8];                         // of (bankOfferItems)
        Oh = false;
        df = new int[8];                         // df (duelTheirItems)
        fl = 0;
        ignoreListEntry = "";                    // ig
        qc = 0;
        pk = 0;
        rh = 0;
        Df = 0;
        Pk = new int[50];
        Vi = false;
        jj = new int[14];                        // jj (skillXpAccum)
        Ah = 0;
        uc = 0;
        le = 0;
        dk = 1;
        kf = new int[8];                         // kf (bankSlotItems)
        Aj = new int[35];         // Aj
        Vj = 0;
        hh = 0;
        vc = new int[1500];                      // vc (objectTileZ)
        fh = -2;
        Nj = 0;
        sd = 0;
        Se = new int[1500];                      // Se (objectTileY)
        Yi = false;
        ld = 2;
        kc = 0;
        nh = 0;
        Xd = 0;
        Yd = 0;
        mh = false;
        Qk = false;
        Td = false;
        vg = 0;
        pg = 0;
        dd = false;
        sj = -2;
        Fe = false;
        Xj = false;
        Bh = -1;
        Dd = new int[14];                        // Dd (skillXpDeltas)
        Je = false;
        Ke = 0;
        Tk = 0;
        jc = 0;
        vf = new int[35];            // vf (inventoryItems)
        ak = new int[50];                        // ak (menuOptionTargets)
        Ub = false;
        ah = new String[5];
        Tb = new GameCharacter[500];       // Tb
    }

    // -------------------------------------------------------------------------
    /** GameShell hook: per-login-screen tick — drive login/sleep screens + idle camera drift.
     *  Called from GameShell.run.   obf: void e(int)   obf-label: il[227]="client.MA(" */
    @Override
    protected final void handleInputs(int frameTick) {
        try {
            // Skip if domain-lock tripped or fatal load error already set.
            if (Xh) {
                return;
            }
            if (outOfMemory) {
                return;
            }

            // Clear the right-click menu/option buffer when not a forced tick (<64).
            if (frameTick < 64) {
                Oc = null;                       // Oc = (int[])null
            }

            if (fatalLoadError) {                // Vc
                return;
            }

            // Start the active audio voice if present.
            if (ni != null) {                    // ni (soundChannel)
                ni.tick();                       // ni.a()
            }

            try {
                // Advance the login-screen animation frame.
                loginAnimFrame++;                // jk++

                if (screenMode == 0) {           // login
                    loginFrameCount = 0;         // sb = 0
                    drawLoginInput(2);           // x(2)  [login.part: drawLoginInput]
                }

                if (screenMode == 1) {           // in-game
                    loginFrameCount++;           // sb++
                    handleGameInput(0);          // J(0)  [mainloop.part: handleGameInput / tick]
                }

                // Reset per-frame "mouse button seen" flag.
                lastMouseButtonDown = 0;

                // Every 500 ticks, randomly nudge the idle camera drift.
                // NOTE: Be = Be, oc = oc; the per-axis drift
                //   increments are the dual-purpose fields mouseButtonMode (eg) for X and
                //   cameraRotationYIncrement (Ok) for Y. (Oracle: cameraRotationXIncrement,
                //   cameraRotationYIncrement.)
                cameraRotationTime++;            // oj++
                if (cameraRotationTime > 500) {
                    cameraRotationTime = 0;
                    int rnd = (int)(4.0 * Math.random());
                    if ((rnd & 2) == 2) {                          // ~(2 & rnd) == -3
                        oc += cameraRotationYIncrement; // oc += Ok
                    }
                    if ((rnd & 1) == 1) {
                        Be += mouseButtonMode;          // Be += eg
                    }
                }

                // Bounce the drift increments at +/-50.
                if (Be < -50) mouseButtonMode = 2;            // Be<-50 -> eg=2
                if (oc < -50) cameraRotationYIncrement = 2;   // oc<-50 -> Ok=2
                if (Be > 50)  mouseButtonMode = -2;           // Be>50  -> eg=-2

                // Decrement the chat-tab flash timers.
                if (Mh > 0) Mh--;
                if (Vj > 0) Vj--;
                if (Ee > 0) Ee--;
                if (Qe > 0) Qe--;

                if (oc > 50) cameraRotationYIncrement = -2;   // oc>50 -> Ok=-2
            } catch (OutOfMemoryError oom) {
                outOfMemory = true;              // Ue = true
            }
        } catch (RuntimeException e) {
            throw ErrorHandler.wrap(e, STRINGS[227] + frameTick + ')'); // il[227]="client.MA("
        }
    }

    // -------------------------------------------------------------------------
    /** Resolve Ek host/ports, build XP table, init li/Ek/Hh/UI, load all assets.
     *  Called from GameShell's loader thread.   obf: void a(byte)   obf-label: il[334]="client.KC("
     *  NOTE: obf `client.a(byte)` OVERRIDES `GameShell.a(byte)` (deob: GameShell.setLoaderApplet);
     *  the boot path dispatches here polymorphically via `this.setLoaderApplet((byte)-92)`.
     *  Renamed loadGameConfig -> setLoaderApplet so the override (and that dispatch) is valid. */
    @Override
    public final void setLoaderApplet(byte dummy) {
        try {
            // Applet domain-lock: must be served from runescape.com (or local.runescape.com).
            if (isApplet) {                      // hj
                String host = this.getDocumentBase().getHost().toLowerCase();
                // il[333]="runescape.com", il[329]=".runescape.com"
                if (!host.equals(STRINGS[333]) && !host.endsWith(STRINGS[329])) {
                    Xh = true;                   // domain-lock tripped
                    return;
                }
            }

            this.pollInput(-113);                 // n(-113) [input.part: pollInput — window resize/setup]

            // Check/init the display surface; on failure flag a fatal load error.
            if (!this.initGraphics(2)) {          // e.d(2)
                fatalLoadError = true;            // Vc
                return;
            }

            // Run the content CRC checker/downloader.
            CacheUpdater.setBzipRef(MessageList.p, (byte)-72); // cb.a(wb.p, -72)  [MessageList.p = static BZip font-width table; see ASSEMBLE flag]

            // If the loader pre-fetched a cache file, wire it into the archive store.
            try {
                if (ImageLoader.loaderThread.seedFile != null) {   // pa.k.s
                    Packet.outgoingTelemetry = new DataStore(            // b.q = new nb(...)
                        ImageLoader.loaderThread.seedFile, 24, 0
                    );
                    ImageLoader.loaderThread.seedFile = null;
                }
            } catch (IOException ex) {
                Packet.outgoingTelemetry = null;
            }

            // Build the experience-per-level table (levels 1..99):
            //   xpThis = (int)(300 * 2^(n/7) + n)  with n = lvl+1, accumulated, clamped to 0x0FFFFFFC.
            int xpAcc = 0;
            for (int lvl = 0; lvl < 99; lvl++) {
                int n = lvl + 1;
                int xpThis = (int)(300.0 * Math.pow(2.0, n / 7.0) + n);
                xpAcc += xpThis;
                experienceTable[lvl] = StreamBase.bitwiseAnd(xpAcc, 0x0FFFFFFC); // ib.a(sum, 268435452)
            }

            // il[332]="referid" applet param -> Yd.
            try {
                Yd = Integer.parseInt(this.getParameter(STRINGS[332]));
            } catch (Exception ignored) {}

            // il[331]="servertype" applet param -> members/free bits.
            try {
                String serverType = this.getParameter(STRINGS[331]);
                // dummy-param guard (anti-tamper): if dummy != -92, force narrow layout.
                if (dummy != -92) {
                    screenHeight = -6;            // Oi = -6
                }
                int modeVal = Integer.parseInt(serverType);
                if ((modeVal & 2) != 0) isFreeWorld = true;     // cf
                if ((modeVal & 1) != 0) Kd = true;  // Pg
            } catch (Exception ignored) {}

            // --- Server host / port selection (logic per CLEAN base) ---
            //   if fontMetrics != defaultFont:
            //       if SpriteScaler.a(fontMetrics,-117): codeBase host + (xd=nodeId+50000, fc=40000+nodeId)
            //       else if fontMetrics == BZip.instance: il[328] host + (xd=nodeId+50000, fc=40000+nodeId)
            //       else: leave host unset
            //   else (fontMetrics == defaultFont): codeBase host + (xd=443, fc=43594)
            // NOTE: fc = 40000 - -aa.l = 40000 + nodeId (PLUS, not minus).
            if (Surface.decoyStringHolder != GameShell.chatCipher) {   // ua.E != e.i
                if (SpriteScaler.isChatCipherKnown(GameShell.chatCipher)) { // ia.a(e.i,-117)
                    serverHost = this.getCodeBase().getHost();    // Dh
                    portB = 40000 + BZip.entityLimit;                  // fc = 40000 - -aa.l
                    portA = BZip.entityLimit + 50000;                  // xd
                } else if (ClientRuntimeException.LOCAL_CIPHER == GameShell.chatCipher) { // la.b == e.i
                    // RUNTIME ADD-3 (Boot): standalone has no codeBase; pin the live
                    // OpenRSC login endpoint directly. Original obf set
                    // xd=nodeId+50000, fc=40000+nodeId, Dh="local.runescape.com".
                    portA = 43594;                                // xd (worldIndex<=1 path)
                    portB = 43594;                                // fc (worldIndex>=2 path)
                    serverHost = "127.0.0.1";                     // Dh
                }
                // else: host left unset.
            } else {
                serverHost = this.getCodeBase().getHost();        // Dh
                portB = 43594;                                    // fc
                portA = 443;                                      // xd
            }

            CacheFile.cacheLimit = 1000;          // d.l = 1000

            // Load the "Configuration" options/data archive (ui_a.part defines f(boolean)
            // as drawOptionsTab — its body is actually the config loader). false = no SFX.
            drawOptionsTab(false);                 // f(false)  [ui_a.part: drawOptionsTab]
            if (fatalLoadError) return;

            // --- Sprite-slot base offsets in the SurfaceSprite map ---
            spriteBaseInventory   = 2000;                          // tg
            spriteBaseChars       = spriteBaseInventory + 100;     // hc
            spriteBaseNpcs        = spriteBaseChars + 50;          // sg
            spriteBaseObjects     = spriteBaseNpcs + 1000;         // dg
            spriteBaseGroundItems = spriteBaseObjects + 10;        // kd
            spriteBaseWalls       = spriteBaseGroundItems + 50;    // Eh
            spriteBaseBubbles     = spriteBaseWalls - -10;         // Wj
            spriteBaseTextures    = spriteBaseBubbles + 5;         // ij

            graphics = this.getGraphics();        // Xb

            // GameShell.a(int,byte): set the target frame interval (Ib = 1000/fps).
            this.setTargetFps(50, (byte)107);     // GameShell.a(50, 107)

            // Software renderer: Wd x (Oi+12), 4000 sprite slots.
            li = new SurfaceSprite(screenWidth, screenHeight + 12, 4000, this); // li = new ba(...)
            li.client = this;                // li.dc = this
            li.setBounds(0, screenWidth, screenHeight + 12, 0, (byte)54); // li.a(0,...)

            // Message lists: il[335]="Choose option" (chat), then friends + ignore.
            chatList    = new MessageList(li, 1, STRINGS[335]); // zh
            zh = new MessageList(li, 1);                // Wf
            ignoreList  = new MessageList(li, 1);                // He

            Timer.legacyFlag = false;             // p.d = false
            StringCodec.STATUS_NOT_FOUND = spriteBaseChars; // u.g = hc

            zk = new Panel(li, 5);   // Mc
            int panelLeft = li.width - 199;  // li.u - 199
            Ud = zk.addScrollList(panelLeft, 196, 90, true, dummy ^ -12, 500, 24 + 36, 1);

            ge = new Panel(li, 5);    // zk
            Hi = ge.addScrollList(panelLeft, 196, 126, true, dummy + 197, 500, 36 + 40, 1);

            Af = new Panel(li, 5);   // fe
            lk = Af.addScrollList(panelLeft, 196, 251, true, 106, 500, 24 + 36, 1);

            loadMedia2d((byte)-49);               // m(-49)
            if (fatalLoadError) return;

            loadEntitySprites(true);              // c(true)
            if (fatalLoadError) return;

            // Scene renderer (lb=Scene): 15000x15000, 1000 models. Field `world` is type lb=Scene (alias of Ek).
            Ek = new Scene(li, 15000, 15000, 1000); // Ek = new lb(li,15000,15000,1000)
            Ek.setBounds(
                screenHeight / 2, true, screenWidth, screenWidth / 2, screenHeight / 2,
                qd, screenWidth / 2
            );                                    // Ek.a(...)
            Ek.clipFar3d = 2400;               // Ek.Mb
            Ek.clipFar2d = 2400;               // Ek.X
            Ek.fogZDistance = 2300;            // Ek.G
            Ek.fogZFalloff = 1;                // Ek.P
            Ek.setLight(-50, -10, true, -50);  // Ek.a(-50,-10,true,-50)

            // World terrain engine (k=World) bound to the scene + surface. Field `scene` is type k=World (alias of Hh).
            Hh = new World(Ek, li);    // Hh = new k(Ek, li)
            Hh.baseMediaSprite = spriteBaseInventory; // Hh.x = tg

            loadTextures((byte)91);               // j(91)
            if (fatalLoadError) return;

            loadModelDefs(true);                  // e(true)
            if (fatalLoadError) return;

            loadMaps(5359);                       // m(5359)
            if (fatalLoadError) return;

            if (Kd) {
                sound.initSounds(-90);                  // E(-90)
            }
            if (fatalLoadError) return;

            // GameShell.a(int,byte,String): loading-progress bar at 100% with the
            // message il[330]="Starting game...".
            this.showLoadingProgress(100, (byte)-99, STRINGS[330]); // GameShell.a(100,-99,il[330])

            initShopPanel(56);                    // O(56)
            drawLoginScreen(3845);                // p(3845)  [login.part: drawLoginScreen]
            drawCharDesign(dummy ^ 24649);        // t(dummy ^ 0x6049)  [ui_b.part: drawCharDesign]
            resetMenuState((byte)-88);            // e(-88)  [ui_b.part: resetMenuState]
            this.profileA(-77);                   // GameShell.a(-77)  (near no-op)
            drawProgressBar(-116);                // y(-116)  [util.part: drawProgressBar]
        } catch (RuntimeException e) {
            throw ErrorHandler.wrap(e, STRINGS[334] + dummy + ')'); // il[334]="client.KC("
        }
    }

    // -------------------------------------------------------------------------
    /** Load the 3D model name/definition archives (*.ob2 / *.ob3); "3d models".
     *  obf: void e(boolean)   obf-label: il[286]="client.DA(" */
    private final void loadModelDefs(boolean membersContent) {
        try {
            // Base (free) model archives. ca.a((byte)91, name) = GameModel.textureId.
            GameModel.textureId((byte)91, STRINGS[287]); // "torcha2"
            GameModel.textureId((byte)91, STRINGS[284]); // "torcha3"
            GameModel.textureId((byte)91, STRINGS[295]); // "torcha4"
            GameModel.textureId((byte)91, STRINGS[294]); // "skulltorcha2"
            GameModel.textureId((byte)91, STRINGS[275]); // "skulltorcha3"
            GameModel.textureId((byte)91, STRINGS[278]); // "skulltorcha4"
            GameModel.textureId((byte)91, STRINGS[277]); // "firea2"
            GameModel.textureId((byte)91, STRINGS[273]); // "firea3"
            GameModel.textureId((byte)91, STRINGS[283]); // "fireplacea2"
            GameModel.textureId((byte)91, STRINGS[298]); // "fireplacea3"
            GameModel.textureId((byte)91, STRINGS[282]); // "firespell2"

            if (!membersContent) {
                return;                           // free-only set done
            }

            // Members-only model archives.
            GameModel.textureId((byte)91, STRINGS[280]); // "firespell3"
            GameModel.textureId((byte)91, STRINGS[276]); // "lightning2"
            GameModel.textureId((byte)91, STRINGS[289]); // "lightning3"
            GameModel.textureId((byte)91, STRINGS[299]); // "clawspell2"
            GameModel.textureId((byte)91, STRINGS[293]); // "clawspell3"
            GameModel.textureId((byte)91, STRINGS[292]); // "clawspell4"
            GameModel.textureId((byte)91, STRINGS[288]); // "clawspell5"
            GameModel.textureId((byte)91, STRINGS[291]); // "spellcharge2"
            GameModel.textureId((byte)91, STRINGS[281]); // "spellcharge3"

            // Standalone (no GameFrame): load model bodies from the "3d models" archive.
            if (InputState.gameFrame == null) {     // kb.a == null
                byte[] modelData = this.readDataFile(STRINGS[285], 60, 9, 84); // il[285]="3d models"
                if (modelData == null) {
                    fatalLoadError = true;        // Vc
                    return;
                }

                for (int i = 0; i < SpriteScaler.modelNameCount; i++) { // ia.b
                    // il[290]=".ob2"
                    int offset = NameHash.getFileOffset(
                        NameTable.modelNames[i] + STRINGS[290], (byte)68, modelData
                    );                            // oa.a(...)
                    if (offset == 0) {
                        objectModels[i] = new GameModel(1, 1);
                    } else {
                        objectModels[i] = new GameModel(modelData, offset, true);
                    }
                    // il[296]="giantcrystal" -> mark double-sided.
                    if (NameTable.modelNames[i].equals(STRINGS[296])) {
                        objectModels[i].transparent = true; // ca.cb
                    }
                }
            } else {
                // Applet: load each model body from a per-name .ob3 file.
                this.showLoadingProgress(70, (byte)-98, STRINGS[274]); // GameShell.a(70,-98,il[274]="Loading 3d models")

                for (int i = 0; i < SpriteScaler.modelNameCount; i++) {
                    // il[297]="../content/src/models/", il[279]=".ob3"
                    objectModels[i] = new GameModel(
                        STRINGS[297] + NameTable.modelNames[i] + STRINGS[279]
                    );
                    if (NameTable.modelNames[i].equals(STRINGS[296])) {
                        objectModels[i].transparent = true;
                    }
                }
            }
        } catch (RuntimeException e) {
            throw ErrorHandler.wrap(e, STRINGS[286] + membersContent + ')'); // il[286]="client.DA("
        }
    }

    // -------------------------------------------------------------------------
    /** Load the 2D UI sprite archives ("2d graphics"); blit into SurfaceSprite slots.
     *  obf: void m(byte)   obf-label: il[104]="client.IA(" */
    private final void loadMedia2d(byte widescreen) {
        try {
            // il[110]="2d graphics" archive; il[103]="index.dat" shared index.
            byte[] data = this.readDataFile(STRINGS[110], 20, 8, 76);
            if (data == null) {
                fatalLoadError = true;            // Vc
                return;
            }

            byte[] index = StreamFactory.lookupEntityDefRecord(STRINGS[103], 0, data); // na.a("index.dat",...)

            // surface.parseSprite(slotBase, layerCount, entryData, spriteCount, index).
            // Sprite filenames are the real decoded il[] strings.
            li.parseSprite(spriteBaseInventory,      1, StreamFactory.lookupEntityDefRecord(STRINGS[111], 0, data), 120, index); // inv1.dat
            li.parseSprite(spriteBaseInventory + 1,  6, StreamFactory.lookupEntityDefRecord(STRINGS[95],  0, data),  52, index); // inv2.dat
            li.parseSprite(spriteBaseInventory + 9,  1, StreamFactory.lookupEntityDefRecord(STRINGS[98],  0, data), 101, index); // bubble.dat
            li.parseSprite(spriteBaseInventory + 10, 1, StreamFactory.lookupEntityDefRecord(STRINGS[109], 0, data),  86, index); // runescape.dat
            li.parseSprite(spriteBaseInventory + 11, 3, StreamFactory.lookupEntityDefRecord(STRINGS[101], 0, data),  84, index); // splat.dat
            li.parseSprite(spriteBaseInventory + 14, 8, StreamFactory.lookupEntityDefRecord(STRINGS[99],  0, data), 111, index); // icon.dat
            li.parseSprite(spriteBaseInventory + 22, 1, StreamFactory.lookupEntityDefRecord(STRINGS[112], 0, data), 112, index); // hbar.dat
            li.parseSprite(spriteBaseInventory + 23, 1, StreamFactory.lookupEntityDefRecord(STRINGS[97],  0, data), 104, index); // hbar2.dat
            li.parseSprite(spriteBaseInventory + 24, 1, StreamFactory.lookupEntityDefRecord(STRINGS[96],  0, data),  73, index); // compass.dat
            li.parseSprite(spriteBaseInventory + 25, 2, StreamFactory.lookupEntityDefRecord(STRINGS[100], 0, data), 100, index); // buttons.dat
            li.parseSprite(spriteBaseChars,          2, StreamFactory.lookupEntityDefRecord(STRINGS[106], 0, data), 125, index); // scrollbar.dat
            li.parseSprite(spriteBaseChars + 2,      4, StreamFactory.lookupEntityDefRecord(STRINGS[93],  0, data),  68, index); // corners.dat

            // Widescreen members layout: extend the status bar.
            if (widescreen > -1) {
                screenHeight = 24;                // Oi = 24
            }

            li.parseSprite(spriteBaseChars + 6,  2,            StreamFactory.lookupEntityDefRecord(STRINGS[107], 0, data),  74, index); // arrows.dat
            li.parseSprite(spriteBaseGroundItems, FontWidths.scratchCounterC,
                                                                   StreamFactory.lookupEntityDefRecord(STRINGS[105], 0, data),  83, index); // projectile.dat
            li.parseSprite(spriteBaseBubbles,     2,          StreamFactory.lookupEntityDefRecord(STRINGS[108], 0, data), 116, index); // crowns.dat

            li.setInlineSpriteBase(-123, spriteBaseBubbles); // li.d(-123, Wj)

            // Numbered "objects"+hasPainted+".dat" NPC sprite sheets (30 per sheet), Utility.chatLineState total.
            int remaining = Utility.chatLineState;  // mb.l
            int sheet = 1;
            while (remaining > 0) {
                int count = (remaining <= 30) ? remaining : 30;
                remaining -= 30;
                li.parseSprite(
                    spriteBaseNpcs + 30 * (sheet - 1), count,
                    StreamFactory.lookupEntityDefRecord(STRINGS[94] + sheet + STRINGS[102], 0, data),
                    109, index
                );                                // il[94]="objects", il[102]=".dat"
                sheet++;
            }

            // Register chroma-key separators for each filled sprite slot.
            li.loadSprite(spriteBaseInventory, -342059728);     // li.b(tg, colour)
            li.loadSprite(spriteBaseInventory + 9, -342059728);
            for (int slot = 11; slot < 26; slot++) {
                li.loadSprite(spriteBaseInventory + slot, -342059728);
            }
            for (int slot = 0; slot < FontWidths.scratchCounterC; slot++) {    // n.c
                li.loadSprite(slot + spriteBaseGroundItems, -342059728);
            }
            for (int slot = 0; slot < Utility.chatLineState; slot++) { // mb.l
                li.loadSprite(slot + spriteBaseNpcs, -342059728);
            }
        } catch (RuntimeException e) {
            throw ErrorHandler.wrap(e, STRINGS[104] + widescreen + ')'); // il[104]="client.IA("
        }
    }

    // -------------------------------------------------------------------------
    /** Load player/NPC animation frame sprites ("people and monsters" / "a.dat" / "f.dat").
     *  obf: void c(boolean)   obf-label: il[325]="client.KD(" */
    private final void loadEntitySprites(boolean doLoad) {
        try {
            if (!doLoad) {
                return;
            }

            // il[324]="people and monsters" archive; il[103]="index.dat" index.
            byte[] data = this.readDataFile(STRINGS[324], 30, 1, 88);
            if (data == null) {
                fatalLoadError = true;            // Vc
                return;
            }

            byte[] index = StreamFactory.lookupEntityDefRecord(STRINGS[103], 0, data); // na.a("index.dat",...)

            // Members get an additional "member graphics" archive (il[326]).
            byte[] membersData = null;
            byte[] membersIndex = null;
            if (Kd) {                 // Pg
                membersData = this.readDataFile(STRINGS[326], 45, 2, 68);
                if (membersData == null) {
                    fatalLoadError = true;        // Vc
                    return;
                }
                membersIndex = StreamFactory.lookupEntityDefRecord(STRINGS[103], 0, membersData);
            }

            dj = 0;
            uc = dj;                              // running sprite cursor (li slot index)
            int frameCount = 0;

            int total = StreamFactory.engineArraySize; // na.e (= GameData.animationCount)
            label131:
            for (int idx = 0; idx < total; idx++) {
                String name = CacheUpdater.contentNames[idx]; // cb.e[idx]

                // De-dup: if an earlier entry shares this name, alias its sprite slot and
                // SKIP loading entirely — note uc is NOT advanced for duplicates (oracle-verified).
                for (int prev = 0; prev < idx; prev++) {
                    if (CacheUpdater.contentNames[prev].equalsIgnoreCase(name)) {
                        WorldEntity.spriteOffsets[idx] = WorldEntity.spriteOffsets[prev]; // w.g[idx]=w.g[prev]
                        continue label131;
                    }
                }

                // Body sprite (15 frames): from the base archive, else (members) the member archive.
                byte[] bodyData  = StreamFactory.lookupEntityDefRecord(name + STRINGS[102], 0, data); // ".dat"
                byte[] bodyIndex = index;
                if (bodyData == null && Kd) {
                    bodyIndex = membersIndex;
                    bodyData  = StreamFactory.lookupEntityDefRecord(name + STRINGS[102], 0, membersData);
                }

                if (bodyData != null) {
                    frameCount += 15;
                    li.parseSprite(uc, 15, bodyData, 83, bodyIndex); // li.a(uc,15,...)

                    // "a.dat" extra-frame set (3 frames), when flagged (nb.d[idx] == 1).
                    if (DataStore.tamperScratch[idx] == 1) { // ~nb.d[idx] == -2  <=>  nb.d[idx] == 1
                        byte[] animData  = StreamFactory.lookupEntityDefRecord(name + STRINGS[321], 0, data); // "a.dat"
                        byte[] animIndex = index;
                        if (animData == null && Kd) {
                            animData  = StreamFactory.lookupEntityDefRecord(name + STRINGS[321], 0, membersData);
                            animIndex = membersIndex;
                        }
                        frameCount += 3;
                        li.parseSprite(uc + 15, 3, animData, 89, animIndex);
                    }

                    // "f.dat" front/face-frame set (9 frames), when flagged (aa.c[idx] == 1).
                    if (BZip.entityFlags[idx] == 1) {
                        byte[] faceData  = StreamFactory.lookupEntityDefRecord(name + STRINGS[323], 0, data); // "f.dat"
                        byte[] faceIndex = index;
                        if (faceData == null && Kd) {
                            faceData  = StreamFactory.lookupEntityDefRecord(name + STRINGS[323], 0, membersData);
                            faceIndex = membersIndex;
                        }
                        li.parseSprite(uc + 18, 9, faceData, 76, faceIndex);
                        frameCount += 9;
                    }

                    // Register chroma separators across the 27-slot block (when n.m[idx] != 0).
                    if (FontWidths.prayerM[idx] != 0) { // ~n.m[idx] != -1  <=>  n.m[idx] != 0
                        for (int slot = uc; slot < uc + 27; slot++) {
                            li.loadSprite(slot, -342059728); // li.b(slot, colour)
                        }
                    }
                }

                WorldEntity.spriteOffsets[idx] = uc; // w.g[idx] = uc
                uc += 27;
            }

            // il[322]="Loaded: ", il[327]=" frames of animation"
            System.out.println(STRINGS[322] + frameCount + STRINGS[327]);
        } catch (RuntimeException e) {
            throw ErrorHandler.wrap(e, STRINGS[325] + doLoad + ')'); // il[325]="client.KD("
        }
    }

    // -------------------------------------------------------------------------
    /** Load landscape/map archives into the World ("landscape"/"map"/members variants).
     *  obf: void m(int)   obf-label: il[603]="client.ED(" */
    private final void loadMaps(int dummy) {
        try {
            // il[602]="map" -> world.ctrlDown ; il[599]="landscape" -> world.Q
            // (Hh = world, type k = World per NAMING.md.)
            Hh.mapPack = this.readDataFile(STRINGS[602], 70, 4, 66); // Hh.ctrlDown

            if (Kd) {                 // Pg
                // il[601]="members map" -> world.m
                Hh.memberMapPack = this.readDataFile(STRINGS[601], 75, 5, 76); // Hh.m
            }

            // il[599]="landscape" -> world.Q
            Hh.landscapePack = this.readDataFile(STRINGS[599], 80, 6, 54); // Hh.Q

            // Anti-tamper timing guard (dummy != 5359): drawSprite no-op — stripped.

            if (Kd) {
                // il[600]="members landscape" -> world.mouseX ; (dummy ^ 5283) is the obf'd crc arg.
                Hh.memberLandscapePack = this.readDataFile(STRINGS[600], 85, 7, dummy ^ 5283); // Hh.mouseX
            }
        } catch (RuntimeException e) {
            throw ErrorHandler.wrap(e, STRINGS[603] + dummy + ')'); // il[603]="client.ED("
        }
    }

    // -------------------------------------------------------------------------
    /** Load the "Textures" archive into the Scene renderer.
     *  obf: void j(byte)   obf-label: il[241]="client.UB(" */
    private final void loadTextures(byte dummy) {
        try {
            // Junk: int j = -11 % ((-66 - dummy) / 55) — opaque, stripped.

            // il[240]="Textures" archive; il[103]="index.dat" index.
            byte[] data = this.readDataFile(STRINGS[240], 50, 11, 111);
            if (data == null) {
                fatalLoadError = true;            // Vc
                return;
            }

            byte[] index = StreamFactory.lookupEntityDefRecord(STRINGS[103], 0, data); // na.a("index.dat",...)

            // Allocate Scene texture slots (Ek = scene, type lb = Scene per NAMING.md).
            Ek.allocateTextures(0, 11, 7, DownloadWorker.halfPixelModeFlag); // Ek.a(0,11,7,jb.o)

            for (int i = 0; i < DownloadWorker.halfPixelModeFlag; i++) { // jb.o
                String texName = Utility.chatLines[i];           // mb.g[i]
                // il[102]=".dat"
                byte[] texEntry = StreamFactory.lookupEntityDefRecord(texName + STRINGS[102], 0, data);

                // Parse the texture sprite into the Eh scratch slot, fill a 128x128 magenta
                // chroma box, then draw the sprite over it. (Clean order: parse, box, draw.)
                li.parseSprite(spriteBaseWalls, 1, texEntry, 88, index);   // li.a(Eh,1,var7,88,var4)
                li.drawBox(0, (byte)-117, 0xFF00FF, 0, 128, 128);         // li.a(0,-117,16711935,0,128,128)
                li.drawSprite(-1, spriteBaseWalls, 0, 0);                  // li.b(-1,Eh,0,0)

                int texSize = li.spriteWidthFull[spriteBaseWalls];             // li.originX[Eh]

                // Optional overlay variant (Timer.altTextureNames[i] = p.c[i]).
                String altName = Timer.legacyStringsC[i];                      // p.c[i]
                if (altName != null && altName.length() > 0) {
                    byte[] altEntry = StreamFactory.lookupEntityDefRecord(altName + STRINGS[102], 0, data);
                    li.parseSprite(spriteBaseWalls, 1, altEntry, 109, index);
                    li.drawSprite(-1, spriteBaseWalls, 0, 0);
                }

                // Register the texture sprite at slot (ij + i): li.d(ij+i, size, 113, size, 0, 0).
                li.drawSprite(i + spriteBaseTextures, texSize, 113, texSize, 0, 0);

                int texSizeSq = texSize * texSize;
                // Chroma-key fix in the raw pixel buffer: pixel 0x00FF00 (green) -> 0xFF00FF (magenta).
                // clean: if (~li.ob[ij+i][px] == -65281) li.ob[ij+i][px] = 16711935;
                for (int px = 0; px < texSizeSq; px++) {
                    if (li.spritePixels[spriteBaseTextures + i][px] == 0x00FF00) {
                        li.spritePixels[spriteBaseTextures + i][px] = 0xFF00FF;
                    }
                }

                li.drawWorld(false, i + spriteBaseTextures);            // li.a(false, ij+i)

                // Register texture with the Scene: Ek.a(i, 74, pixelData, size/64-1, alphaData).
                Ek.defineTexture(
                    i, (byte)74,
                    li.spritePalette[spriteBaseTextures + i],   // li.Y[ij+i]
                    texSize / 64 - 1,
                    li.spriteColourIndex[spriteBaseTextures + i]   // li.ctrlDown[ij+i]
                );
            }
        } catch (RuntimeException e) {
            throw ErrorHandler.wrap(e, STRINGS[241] + dummy + ')'); // il[241]="client.UB("
        }
    }

    // -------------------------------------------------------------------------
    /** GameShell paint hook: fatal-error / domain-lock / out-of-memory screens, else normal render.
     *  obf: void b(boolean)   obf-label: il[481]="client.JD(" */
    @Override
    protected final synchronized void draw(boolean clearDomainLock) {
        try {
            // First-paint hook: bump the loader bar once.
            if (super.hasPainted) {                         // GameShell.hasPainted (shadowed by Mudclient int[] hasPainted)
                this.pollInput(-108);             // n(-108)  [input.part: pollInput]
                super.hasPainted = false;
            }

            if (!fatalLoadError) {                // !Vc
                if (Xh) {
                    // --- Domain-lock: "Error - unable to load game!" ---
                    Graphics g = this.getGraphics();
                    if (g != null) {
                        g.translate(this.originX, this.originY);
                        g.setColor(Color.black);
                        g.fillRect(0, 0, 512, 356);
                        g.setFont(new Font(STRINGS[477], Font.BOLD, 20)); // il[477]="Helvetica"
                        g.setColor(Color.white);
                        g.drawString(STRINGS[485], 50, 50);  // "Error - unable to load game!"
                        g.drawString(STRINGS[492], 50, 100); // "To play RuneScape make sure you play from"
                        g.drawString(STRINGS[495], 50, 150); // "http://www.runescape.com"
                        this.setTargetFps(1, (byte)111);     // GameShell.a(1, 111) — slow repaint (Ib=1000ms)
                    }
                } else if (outOfMemory) {         // Ue
                    // --- Out-of-memory screen ---
                    Graphics g = this.getGraphics();
                    if (g != null) {
                        g.translate(this.originX, this.originY);
                        g.setColor(Color.black);
                        g.fillRect(0, 0, 512, 356);
                        g.setFont(new Font(STRINGS[477], Font.BOLD, 20));
                        g.setColor(Color.white);
                        g.drawString(STRINGS[482], 50, 50);  // "Error - out of memory!"
                        g.drawString(STRINGS[488], 50, 100); // "Close ALL unnecessary programs"
                        g.drawString(STRINGS[494], 50, 150); // "and windows before loading the game"
                        g.drawString(STRINGS[491], 50, 200); // "RuneScape needs about 48meg of spare RAM"
                        this.setTargetFps(1, (byte)106);     // GameShell.a(1, 106) — slow repaint
                    }
                } else {
                    // --- Normal frame ---
                    try {
                        if (clearDomainLock) {
                            Xh = false;
                        }
                        if (li == null) {    // li
                            return;
                        }
                        if (screenMode == 0) {    // ~qg == -1  ->  qg == 0 (login)
                            li.loggedIn = false; // li.mouseY = false
                            drawMinimap(2540);    // k(2540)
                        }
                        if (screenMode == 1) {    // ~qg == -2  ->  qg == 1 (in-game)
                            li.loggedIn = true;  // li.mouseY = true
                            drawGameFrame(13);    // f(13)  [ui_a.part: drawGameFrame]
                        }
                    } catch (OutOfMemoryError oom) {
                        outOfMemory = true;       // Ue = true
                    }
                }
            } else {
                // --- Fatal load error (Vc): the multi-step "Sorry, an error..." help screen ---
                Graphics g = this.getGraphics();
                if (g != null) {
                    g.translate(this.originX, this.originY);
                    g.setColor(Color.black);
                    g.fillRect(0, 0, 512, 356);
                    g.setFont(new Font(STRINGS[477], Font.BOLD, 16));
                    g.setColor(Color.yellow);
                    int y = 35;
                    g.drawString(STRINGS[493], 30, y); // "Sorry, an error has occured whilst loading RuneScape"
                    g.setColor(Color.white);
                    y += 50;
                    g.drawString(STRINGS[487], 30, y); // "To fix this try the following (in order):"
                    g.setColor(Color.white);
                    y += 50;
                    g.setFont(new Font(STRINGS[477], Font.BOLD, 12));
                    g.drawString(STRINGS[484], 30, y); // "1: Try closing ALL open web-browser windows..."
                    y += 30;
                    g.drawString(STRINGS[489], 30, y); // "2: Try clearing your web-browsers cache..."
                    y += 30;
                    g.drawString(STRINGS[483], 30, y); // "3: Try using a different game-world"
                    y += 30;
                    g.drawString(STRINGS[486], 30, y); // "4: Try rebooting your computer"
                    y += 30;
                    g.drawString(STRINGS[490], 30, y); // "5: Try selecting a different version of Java..."
                    this.setTargetFps(1, (byte)126);   // GameShell.a(1, 126) — slow repaint
                }
            }
        } catch (RuntimeException e) {
            throw ErrorHandler.wrap(e, STRINGS[481] + clearDomainLock + ')'); // il[481]="client.JD("
        }
    }


    // =========================================================================
    // ===== login =====
    // =========================================================================
    //
    // Session / login / registration block: opens the ClientStream, performs the
    // ISAAC-seed + RSA-encrypted login handshake, decodes the server's login
    // response code, renders the login UI (welcome screen + username/password
    // entry), and tears the connection down on loss.
    //
    // NOTE on STRINGS[] (`il[]`): the obfuscated client stores all UI / login-response
    // text XOR-encrypted in STRINGS[]. The plaintext beside each index below was
    // recovered by cross-referencing the canonical RSC login flow
    // (mudclient204 GameConnection.login, whose response-code -> message mapping is
    // byte-identical to this client). Indices are EXACT and were re-verified against
    // the clean Vineflower base: each recurring index resolves to the same plaintext
    // across every response code that uses it (e.g. il[453] = "Wait 60 seconds then
    // retry" for both code 4 and code 10; il[429] = "Error unable to login." for
    // codes -1/8/9/default), which cross-checks the mapping.
    //
    // Helper Mudclient methods called below that are outside the login group (obf -> name):
    //   b(byte,String,String)   -> showLoginScreenStatus  (two-line login banner)
    //   a(String,byte,String)   -> drawTextBox            (overlay box for reconnect)
    //   o(int)                  -> resetLoginState        (clears bj/Xd/qg, and kc when -2)
    //   i(int)                  -> resetGameState         (post-login scene/UI reset, qg=1)
    //   g(int)                  -> onSessionNeedsVerify    (resp==1 verify/sleep hook; mostly stub)
    //   c(byte)                 -> sendQueuedActions      (skeleton mainloop name)
    //   G(int)                  -> sendDialogAnswer       (skeleton packetout name)
    //   a(int,int,String)       -> createSocket           (open world socket)

    /**
     * Two-line login/welcome banner: writes {@code title}/{@code body} into the
     * login-status panel ({@code yi}=panelDuel reused as the message panel) when the
     * login screen is active ({@code Xd}=activePanel == 2). Deob of obf
     * {@code b(byte,String,String)} (clean client.java:12580); the leading anti-tamper
     * byte guard and its dead branch are stripped per the deob convention.
     */
    private final void showLoginScreenStatus(String title, String body) {
        if (Xd == 2) {                       // Xd == 2: login screen active
            if (body == null || body.length() < 1) {
                yi.setFieldText(serverMsgControlId, title, 27642);   // obf: yi.a(td, var2, 27642)
                return;
            }
            yi.setFieldText(fatigueControlId, title, 27642);         // obf: yi.a(Qi, var2, 27642)
            yi.setFieldText(serverMsgControlId, body, 27642);        // obf: yi.a(td, var3, 27642)
        }
    }

    /**
     * Connect to the Ek server and perform the full login handshake for
     * {@code username}/{@code password}.
     *
     * Flow: open a ClientStream socket, send opcode 0 (LOGIN) carrying an
     * RSA-encrypted block (4 random session-key words + username) plus an
     * XTEA-encrypted tail (UID + password), seed the ISAAC stream cipher, then read
     * and decode the one-byte login response. {@code reconnecting} drives the silent
     * auto-relogin path used after a dropped connection.
     *
     * @param dummy        anti-tamper constant (callers pass -12); only used masked,
     *                     e.g. {@code dummy ^ -12 == 0} yields the LOGIN opcode
     * @param username     account name (raw; trimmed + truncated to 20 chars)
     * @param password     account password (truncated to 20 chars)
     * @param reconnecting true for a silent re-establish after lost connection
     *
     * obf: void a(int,String,String,boolean)  [proposed: loginUser]
     */
    private final void loginUser(int dummy, String username, String password, boolean reconnecting) {
        // If this world reported "currently full" recently, refuse to even try and
        // show the full-world banner after a short pause.
        // obf: `Zb` doubles as the world-full cooldown timer here.
        if (this.worldFullTimeout > 0) {
            this.showLoginScreenStatus(STRINGS[436], STRINGS[432]); // "Please wait..." / "Connecting to server"
            try {
                Utility.sleepWithProfile(11200, 2000L);
            } catch (Exception ignored) {
            }
            this.showLoginScreenStatus(STRINGS[422], STRINGS[454]); // "Sorry! The server is currently full." / "Please try again later"
            return;
        }

        // Retry/attempt loop. NOTE: the obf field `worldIndex` (Vh) is overloaded —
        // it is both the world/port selector (see port pick below) and the remaining
        // auto-login attempt counter. The reconnect path (closeConnection) presets it
        // to 10; the entry-panel submit path sets it to 2; each failure decrements it.
        // obf: while (0 < this.Vh)
        while (this.worldIndex > 0) {
            try {
                this.password = password;
                this.username = username;
                // Truncate username to 20 chars / strip illegal characters for the auth packet.
                String authUsername = Packet.formatString(20, (byte) -5, username);

                // obf: ~this.wh.trim().length() == -1  <=>  length() == 0
                if (this.password.trim().length() == 0) {
                    this.showLoginScreenStatus(STRINGS[474], STRINGS[471]); // "You must enter both a username" / "and a password - Please try again"
                    return;
                }

                if (reconnecting) {
                    // Silent reconnect: overlay a "lost connection" box rather than the
                    // normal connecting status.
                    this.addChatMessage(STRINGS[460], (byte) -64, STRINGS[446]); // obf a(String,byte,String); "Connection lost! Please wait..." / "Attempting to re-establish"
                } else {
                    this.showLoginScreenStatus(STRINGS[436], STRINGS[432]); // "Please wait..." / "Connecting to server"
                }

                // World index <= 1 uses the primary login port, otherwise the alternate.
                // DEOB FIX: clean decompile (client.java:7881) reads `this.Vh <= 1 ? this.xd : this.fc`
                // i.e. portA/portB — the loginPort/loginPortAlt fields are phantom (never assigned).
                int port = this.worldIndex <= 1 ? this.portA : this.portB;
                this.Jh = new ClientStream(this.createSocket(dummy, port, this.serverHost), this);
                this.Jh.maxReadTries = CacheFile.unusedCounter; // max read-retry count

                // "limit30" applet param caps the frame rate to 30fps; flagged into the login block.
                int limit30 = 0;
                try {
                    if (InputState.gameFrame == null && this.getParameter(STRINGS[462]).equals("1")) {
                        limit30 = 1;
                    }
                } catch (Exception ignored) {
                }

                // Four random 32-bit words. These are both the ISAAC stream-cipher key
                // (applied to all subsequent traffic) and the XTEA key that encrypts the
                // login packet's plaintext tail (username/UID/password).
                int[] sessionKey = new int[]{
                    (int) (9.9999999E7 * Math.random()),
                    (int) (9.9999999E7 * Math.random()),
                    (int) (9.9999999E7 * Math.random()),
                    (int) (9.9999999E7 * Math.random()),
                };

                // ---- opcode 0: LOGIN ----  (obf: this.Jh.b(0, dummy ^ -12); the ^ -12 dummy reduces to 0)
                // (clientStream.f is the ClientStream's outgoing BitBuffer.)
                this.Jh.newPacket(0, dummy ^ -12);
                // reconnect flag: 1 = re-establish existing session, 0 = fresh login.
                if (reconnecting) {
                    this.Jh.outBuffer.putByte(1);
                } else {
                    this.Jh.outBuffer.putByte(0);
                }
                this.Jh.outBuffer.putInt(ClientIOException.BUILD_REVISION); // client/protocol version

                // Build the RSA block in a scratch Buffer: leading byte 10, the four
                // session-key words, then the password (20-byte space-padded), padding
                // ints and a random tail byte.
                //
                // PROTOCOL FIX (M3, clientVersion 235 > 204): the authentic mudclient this
                // jar was decompiled from puts the *username* in the RSA block and the
                // *password* in the XTEA tail (client.java:7922/7945 — var9.a((byte)-39,var5)
                // with var5=username, this.Jh.f.a((byte)-39,this.wh) with this.wh=password).
                // The OpenRSC server's clientVersion>204 path reads them the OTHER way:
                //   - password from RSA bytes [17,37)  (LoginPacketHandler.java:164
                //     `new String(loginBlock, 17, 20, "UTF8").trim()`)
                //   - username from XTEA bytes [25,..)  (LoginPacketHandler.java:188
                //     `new String(xteaBlock, 25, xteaBlock.length-25, "UTF8")`)
                // and proto/v235/login.go BuildRSABlock/BuildXTEABlock agree.
                // So to authenticate against the M3 server the two credentials must be
                // SWAPPED relative to the authentic client: password→RSA, username→XTEA.
                // The RSA credential must be a fixed 20-char (space-padded) field because
                // the server reads a fixed 20-byte window; formatString(20,...) pads it.
                // PROTOCOL FIX (M3, RSA block length): the v235 server reads a FIXED 61-byte
                // decrypted RSA block: marker[0], keys[1,17), password[17,37) (20 bytes, trimmed),
                // unused[37], nonces[0..4] [38,58), nonce[5] low 3 bytes [58,61)
                // (LoginPacketHandler.java:152-173). The previous build wrote the 20-char password
                // via putString() (which appends a NUL = 21 bytes) and only 1 trailing byte, giving a
                // 59-byte block -> server crashed reading loginBlock[59]/[60] (AIOOBE @173). We now
                // write the password as exactly 20 raw bytes and pad the block to a full 61 bytes so
                // the nonce window lands at the offsets the server reads.
                String rsaPassword = Packet.formatString(20, (byte) -5, this.password);
                Buffer rsaBlock = new Buffer(500);
                rsaBlock.putByte(10);                  // [0] RSA block marker
                rsaBlock.putInt(sessionKey[0]);        // [1,5)  ISAAC/XTEA key 0
                rsaBlock.putInt(sessionKey[1]);        // [5,9)  key 1
                rsaBlock.putInt(sessionKey[2]);        // [9,13) key 2
                rsaBlock.putInt(sessionKey[3]);        // [13,17) key 3
                // [17,37): 20 raw password bytes (server: new String(block,17,20).trim()).
                for (int i = 0; i < 20; i++) {
                    rsaBlock.putByte(rsaPassword.charAt(i));
                }
                rsaBlock.putByte((int) (9.9999999E7 * Math.random())); // [37] unused filler
                // [38,58): 5 session-nonce ints (anti-replay filler before encryption).
                // obf: while (~var10 > -6)  <=>  for (i = 0; i < 5; i++)
                for (int i = 0; i < 5; i++) {
                    rsaBlock.putInt((int) (Math.random() * 9.9999999E7));
                }
                // [58,61): nonce[5] low 3 bytes (server reads block[58],[59],[60]).
                rsaBlock.putByte((int) (9.9999999E7 * Math.random()));
                rsaBlock.putByte((int) (9.9999999E7 * Math.random()));
                rsaBlock.putByte((int) (9.9999999E7 * Math.random()));
                // RSA-encrypt the whole block in place (modulus, exponent).
                rsaBlock.rsaEncrypt(BitBuffer.RSA_MODULUS, -118, FontBuilder.rsaPublicExponent);

                // Emit the encrypted RSA block bytes, then a 2-byte length placeholder
                // for the XTEA tail that follows (patched after the tail is written).
                this.Jh.outBuffer.putBytes(0, rsaBlock.offset, rsaBlock.data);
                this.Jh.outBuffer.putShort(0); // placeholder for XTEA tail length
                int tailStart = this.Jh.outBuffer.offset;
                this.Jh.outBuffer.putByte(limit30);
                RecordLoader.loadRecord(22607, this.Jh.outBuffer); // append 24-byte client UID record (XTEA bytes [1,25))
                this.Jh.outBuffer.putString(authUsername);          // XTEA bytes [25,..): username (server reads [25,end))
                // XTEA-encrypt the plaintext tail [tailStart, position) with sessionKey.
                this.Jh.outBuffer.teaDecrypt((byte) 87, tailStart, sessionKey, this.Jh.outBuffer.offset);
                // Back-patch the placeholder with the XTEA tail's actual length.
                // obf: this.Jh.f.d(-tailStart + position, 1)
                this.Jh.outBuffer.patchShortBack(this.Jh.outBuffer.offset - tailStart);
                this.Jh.flushPacket(-6924);
                // Initialise the ISAAC stream cipher for all subsequent traffic.
                this.Jh.seedIsaac((byte) -119, sessionKey);

                // ---- read one-byte login response ----
                int response = this.Jh.readStream(true);
                System.out.println(STRINGS[439] + response); // "login response:"

                // obf: if (~(response & 64) == -1) { ...FAILURE... } else { ...SUCCESS... }
                // ~(response & 0x40) == -1  <=>  (response & 0x40) == 0  =>  login FAILED.
                if ((response & 0x40) == 0) {
                    // Response code 1: session needs verification (sleep word / recovery).
                    // Checked first inside the failure branch in the clean base.
                    if (response == 1) {
                        this.worldIndex = 0;
                        this.setMouseButtonMode(-16433); // obf g(int)
                        return;
                    }

                    // On the silent reconnect path, do not surface any error text — just
                    // give up this attempt and fall through to the reset below.
                    if (!reconnecting) {
                        // Decode the one-byte login response code -> on-screen message.
                        // Mapping re-derived from the clean base's nested if-cascade and
                        // cross-checked against the canonical RSC response messages.
                        if (response == -1) {
                            this.showLoginScreenStatus(STRINGS[429], STRINGS[442]); // "Error unable to login." / "Server timed out"
                        } else if (response == 3) {
                            this.showLoginScreenStatus(STRINGS[431], STRINGS[473]); // "Invalid username or password." / "Try again, or create a new account"
                        } else if (response == 4) {
                            this.showLoginScreenStatus(STRINGS[450], STRINGS[453]); // "That username is already logged in." / "Wait 60 seconds then retry"
                        } else if (response == 5) {
                            this.showLoginScreenStatus(STRINGS[430], STRINGS[467]); // "The client has been updated." / "Please reload this page"
                        } else if (response == 6) {
                            this.showLoginScreenStatus(STRINGS[458], STRINGS[469]); // "You may only use 1 character at once." / "Your ip-address is already in use"
                        } else if (response == 7) {
                            this.showLoginScreenStatus(STRINGS[438], STRINGS[470]); // "Login attempts exceeded!" / "Please try again in 5 minutes"
                        } else if (response == 8) {
                            this.showLoginScreenStatus(STRINGS[429], STRINGS[447]); // "Error unable to login." / "Server rejected session"
                        } else if (response == 9) {
                            this.showLoginScreenStatus(STRINGS[429], STRINGS[445]); // "Error unable to login." / "Loginserver rejected session"
                        } else if (response == 10) {
                            this.showLoginScreenStatus(STRINGS[425], STRINGS[453]); // "That username is already in use." / "Wait 60 seconds then retry"
                        } else if (response == 11) {
                            this.showLoginScreenStatus(STRINGS[457], STRINGS[426]); // "Account temporarily disabled." / "Check your message inbox for details"
                        } else if (response == 12) {
                            this.showLoginScreenStatus(STRINGS[466], STRINGS[426]); // "Account permanently disabled." / "Check your message inbox for details"
                        } else if (response == 14) {
                            this.showLoginScreenStatus(STRINGS[444], STRINGS[449]); // "Sorry! This world is currently full." / "Please try a different world"
                            this.worldFullTimeout = 1500;
                        } else if (response == 15) {
                            this.showLoginScreenStatus(STRINGS[459], STRINGS[455]); // "You need a members account" / "to login to this world"
                        } else if (response == 16) {
                            this.showLoginScreenStatus(STRINGS[440], STRINGS[468]); // "Error - no reply from loginserver." / "Please try again"
                        } else if (response == 17) {
                            this.showLoginScreenStatus(STRINGS[463], STRINGS[435]); // "Error - failed to decode profile." / "Contact customer support"
                        } else if (response == 18) {
                            this.showLoginScreenStatus(STRINGS[451], STRINGS[428]); // "Account suspected stolen." / "Press 'recover a locked account' on front page."
                        } else if (response == 20) {
                            this.showLoginScreenStatus(STRINGS[464], STRINGS[449]); // "Error - loginserver mismatch" / "Please try a different world"
                        } else if (response == 21) {
                            this.showLoginScreenStatus(STRINGS[443], STRINGS[423]); // "Unable to login." / "That is not an RS-Classic account"
                        } else if (response == 22) {
                            this.showLoginScreenStatus(STRINGS[424], STRINGS[465]); // "Password suspected stolen." / "Press 'change your password' on front page."
                        } else if (response == 23) {
                            this.showLoginScreenStatus(STRINGS[461], STRINGS[456]); // (extended code) "Unable to login." / account-support message
                        } else if (response == 24) {
                            this.showLoginScreenStatus(STRINGS[472], STRINGS[448]); // (extended code) account-support message
                        } else if (response == 25) {
                            this.showLoginScreenStatus(STRINGS[434], STRINGS[435]); // (extended code) account-support / "Contact customer support"
                        } else {
                            this.showLoginScreenStatus(STRINGS[429], STRINGS[452]); // "Error unable to login." / "Unrecognised response code"
                        }
                        return;
                    }

                    // Reconnect path with a non-success response: silently give up this attempt.
                    authUsername = "";
                    this.password = "";
                    this.resetTradeDuelState(-2); // obf o(int)
                    return;
                }

                // ---- SUCCESS (bit 0x40 set) ----  low bits carry account flags + rank.
                this.accountFlags = response & 3;            // low 2 bits = account flags
                this.worldIndex = 0;
                this.moderatorLevel = (response & 0x3F) >> 2; // bits 2..5 = staff rank
                this.resetGameState(-109);
                return;
            } catch (Exception e) {
                System.out.println("" + e);
                // On exception, decrement the retry counter and (if any left) loop to retry.
                // obf: if (-1 > ~this.Vh)  <=>  this.Vh > 0
                if (this.worldIndex > 0) {
                    try {
                        Utility.sleepWithProfile(11200, 5000L);
                    } catch (Exception ignored) {
                    }
                    --this.worldIndex;
                    continue;
                }
                if (reconnecting) {
                    // Reconnect ran out of retries: forget credentials, drop to login screen.
                    this.username = "";
                    this.password = "";
                    this.resetTradeDuelState(-2); // obf o(int)
                } else {
                    Utility.reportError(0x1FFFFF, e, STRINGS[427]);
                    this.showLoginScreenStatus(STRINGS[441], STRINGS[433]); // "Sorry! Unable to connect." / "Check internet settings or try another world"
                }
                continue;
            }
        }
        // Loop exited (worldIndex <= 0). For the normal call path (dummy == -12) we are
        // already done; otherwise flush any queued client actions before returning.
        if (dummy == -12) {
            return;
        }
        this.sendQueuedActions((byte) -97);
    }

    /**
     * Append a server/system message to the chat history and route it to the
     * correct message tab.
     *
     * NOTE: the skeleton labelled this `registerAccount`, but the actual bytecode
     * (identical to OpenRSC's {@code showMessage} / mudclient's {@code fdj}) is the
     * chat-message display routine, not account registration. It pushes the message
     * onto the 100-entry rolling history buffers and adds it to the chat/quest/
     * private message panels. There is no opcode-2 REGISTER traffic here. Named to
     * match true behaviour; original skeleton name noted for traceability.
     *
     * @param crownEnabled   show the sender's rank crown (else crownId forced to 0)
     * @param sender         display name of the sender (may be null for system text)
     * @param messageSlot    index into the message-text history array for the insert
     *                       (callers pass 0 = newest slot)
     * @param message        the message body
     * @param type           internal MessageType id (0 game, 1 private-recv,
     *                       2 private-send, 3 quest, 4 chat, 5 private-system,
     *                       6 friend-status, 7 inventory; NOT the OpenRSC enum order)
     * @param crownId        rank/crown id of the sender
     * @param formerName     clan/former display name; used for ignore filtering
     * @param colourOverride explicit colour code, or null to use the type's default
     *
     * obf: void a(boolean,String,int,String,int,int,String,String)  [skeleton: registerAccount; actual: showServerMessage]
     */
    final void showServerMessage(boolean crownEnabled, String sender, int messageSlot, String message,
                                         int type, int crownId, String formerName, String colourOverride) {
        // This client's internal MessageType ids (do NOT match OpenRSC's enum order):
        //   0 = GAME, 1 = PRIVATE_RECV, 2 = PRIVATE_SEND, 3 = QUEST, 4 = CHAT,
        //   5 = PRIVATE_SYSTEM, 6 = FRIEND_STATUS, 7 = INVENTORY.
        //
        // Ignore filtering: for a player-originated message (private-recv/chat/friend)
        // that is NOT showing a crown, drop it if the sender's display key is on the
        // ignore list. The display-key derived here is scratch used only for that
        // comparison; the actual render colour is always the per-type default below.
        // obf: (~type == -2 || -5 == ~type || 6 == type)  <=>  (type == 1 || type == 4 || type == 6)
        if ((type == 1 || type == 4 || type == 6) && formerName != null && !crownEnabled) {
            String senderKey = WorldEntity.trimAndValidateString(formerName, (byte) 93);
            if (senderKey == null) {
                return;
            }
            // obf: while (~i > ~db.g)  <=>  i < LinkedQueue.ignoreListCount
            for (int i = 0; i < LinkedQueue.DEAD_G; i++) {
                if (senderKey.equals(WorldEntity.trimAndValidateString(SpriteScaler.playerNames[i], (byte) 78))) {
                    return;
                }
            }
        }

        // Render colour for this message type (a "@xxx@" colour-code string).
        String colour = StreamFactory.CHAT_COLOR_CODES[type];

        // Flash the destination message tab (activity timer = 200) when the message
        // is NOT landing on the currently-viewed tab. `messageTabSelected` (Zh):
        // 0 = All/Game, 1 = Chat, 2 = Quest, 3 = Private.
        if (this.messageTabSelected != 0) {
            // obf: (5 == type || 1 == type || ~type == -3) && ~Zh != -4
            if ((type == 5 || type == 1 || type == 2) && this.messageTabSelected != 3) {
                this.tabActivityPrivate = 200;  // private tab
            }
            if (type == 4 && this.messageTabSelected != 1) {
                this.tabActivityChat = 200;     // chat tab
            }
            // obf: -4 == ~type  <=>  type == 3
            if (type == 3 && this.messageTabSelected != 2) {
                this.tabActivityQuest = 200;    // quest tab
            }
            if (type == 0 || type == 7) {
                this.tabActivityGame = 200;     // game/inventory -> "All/Game" tab
            }
            // obf: ~type == -1  <=>  type == 0  (NOT type==1; the defective base had this wrong)
            if (type == 0 && this.messageTabSelected != 0) {
                this.messageTabSelected = 0;
            }
            // obf: (~type == -6 || type == 1 || type == 2) && Zh != 3 && ~Zh != -1
            //  ~type == -6  <=>  type == 5
            if ((type == 5 || type == 1 || type == 2) && this.messageTabSelected != 3 && this.messageTabSelected != 0) {
                this.messageTabSelected = 0;
            }
        }

        // An explicit colour override replaces the per-type default.
        if (colourOverride != null) {
            colour = colourOverride;
        }

        // Shift the 100-entry rolling message history down by one and insert at slot 0
        // (except the message body itself, which is stored at index `messageSlot`).
        // NB: the obfuscator scatters this one logical record's parallel arrays across
        // unrelated classes as static storage (FontWidths/ImageLoader/BitBuffer/World/
        // SurfaceSprite/BZip/NameTable) — those host classes are just opaque slots here,
        // not their NAMING.md semantic roles.
        for (int i = 99; i > 0; i--) {
            FontWidths.entryTypes[i] = FontWidths.entryTypes[i - 1];
            ImageLoader.scratchBuf[i] = ImageLoader.scratchBuf[i - 1];
            BitBuffer.UNUSED_N[i] = BitBuffer.UNUSED_N[i - 1];
            World.G[i] = World.G[i - 1];
            SurfaceSprite.recentMessages[i] = SurfaceSprite.recentMessages[i - 1];
            BZip.entityNames[i] = BZip.entityNames[i - 1];
            NameTable.recentNames[i] = NameTable.recentNames[i - 1];
        }
        FontWidths.entryTypes[0] = type;
        ImageLoader.scratchBuf[0] = 300; // frames the message stays in the in-world overlay
        World.G[0] = sender;
        BitBuffer.UNUSED_N[0] = crownId;
        SurfaceSprite.recentMessages[0] = formerName;
        BZip.entityNames[messageSlot] = message;
        NameTable.recentNames[0] = colour;

        // Build the colour-prefixed, fully formatted message string.
        String formatted = colour + Utility.formatChatLine(message, sender, true, type);

        // Route into the chat tab list. type 4 (CHAT) auto-scrolls only if already at
        // the bottom (controlScrollAmount == controlListSize - 4); every other type is
        // appended with auto-scroll forced on.
        // obf: if (-5 == ~type)  <=>  type == 4
        if (type == 4) {
            boolean chatAtBottom =
                this.messagePanel.scrollPos[this.tabChat] == this.messagePanel.itemCount[this.tabChat] - 4;
            this.messagePanel.addListItem(formatted, chatAtBottom, crownId, sender, formerName, (byte) -100, this.tabChat);
        }

        // QUEST (type 3) goes to the quest tab.
        if (type == 3) {
            boolean questAtBottom =
                this.messagePanel.scrollPos[this.tabQuest] == this.messagePanel.itemCount[this.tabQuest] - 4;
            this.messagePanel.addListItem(formatted, questAtBottom, 0, null, null, (byte) -64, this.tabQuest);
        }

        // PRIVATE messages (type 1 = received, 2 = sent) go to the private tab.
        // The received-crown is only shown for received messages.
        // obf: if (-2 == ~type || 2 == type)  <=>  type == 1 || type == 2
        if (type == 1 || type == 2) {
            int privCrown = crownId;
            if (type != 1) {
                privCrown = 0;
            }
            boolean privAtBottom =
                this.messagePanel.scrollPos[this.tabPrivate] == this.messagePanel.itemCount[this.tabPrivate] - 4;
            this.messagePanel.addListItem(formatted, privAtBottom, privCrown, sender, formerName, (byte) -87, this.tabPrivate);
        }
    }

    /**
     * Build the login screen UI: the welcome panel (with account-type gating text
     * and a "Click here to login" button) and the username/password entry panel.
     *
     * Two {@code qa} (Panel) widgets are constructed each time the login screen is
     * shown:
     *   - {@code loginWelcomePanel} (obf {@code ge}) — title, gating text, login button.
     *   - {@code loginEntryPanel}   (obf {@code yi}) — username + password fields,
     *     Ok / Cancel buttons.
     *
     * obf: void p(int)  [proposed: drawLoginScreen]
     */
    private final void drawLoginScreen(int n) {
        // --- welcome panel ---
        this.loginWelcomePanel = new Panel(this.li, 50);
        int y = 40;
        // Centered title at (256, 240).  obf: 200 - -y == 200 + y
        this.loginWelcomePanel.addLabel(true, (byte) -79, 4, 256, STRINGS[237], 200 + y); // "Welcome to RuneScape"

        // Account-type gating sub-line. Selection depends on two flags:
        //   membersWorld (Pg) — this is a members world.
        //   veteranWorld (cf) — this is a veteran/classic world.
        // members && veteran -> STRINGS[233]; members && !veteran -> STRINGS[230];
        // !members && veteran -> STRINGS[238]; !members && !veteran -> no sub-line.
        String gatingText = null;
        if (this.membersWorld) {
            gatingText = this.veteranWorld ? STRINGS[233] : STRINGS[230];
        } else if (this.veteranWorld) {
            gatingText = STRINGS[238];
        }
        if (gatingText != null) {
            this.loginWelcomePanel.addLabel(true, (byte) -109, 4, 256, gatingText, 215 + y);
        }

        // "Click here to login" button at (256, 290).
        this.loginWelcomePanel.addOval(n - 3917, 200, 35, 256, y + 250);
        this.loginWelcomePanel.addLabel(false, (byte) -96, 5, 256, STRINGS[232], y + 250); // "Click here to login"
        this.loginButton = this.loginWelcomePanel.addProgressWidget(256, 200, 250 + y, 91, 35);

        // --- username / password entry panel ---
        this.loginEntryPanel = new Panel(this.li, 50);
        y = 230;
        this.loginTitleControl = this.loginEntryPanel.addLabel(true, (byte) -107, 4, 256, "", y - 30);
        // Instruction line: "Please enter your username and password".
        this.loginPromptControl = this.loginEntryPanel.addLabel(true, (byte) -125, 4, 256, STRINGS[65], y - 10);

        // First entry row. NOTE: field identity here is fixed by where drawLoginInput()
        // reads it back (obf ng -> password, Ih -> username); the displayed label index
        // and the masked flag in this obfuscated build do not follow the usual ordering.
        // The first input (obf ng) is the one read into the password.
        this.loginEntryPanel.addOval(-87, 200, 40, 140, y += 28);
        this.loginEntryPanel.addLabel(false, (byte) -126, 4, 140, STRINGS[235], y - 10);
        // addTextInput(..., masked = false, ...) — the password field is NOT masked in
        // this build (verified: the 8th arg `var8`, which sets Panel.cb[], is `false`).
        this.passwordField = this.loginEntryPanel.addPasswordField(n - 3845, 320, 200, false, 10 + y, 4, 40, false, 140);

        // Second entry row (obf Ih) -> read into the username.
        this.loginEntryPanel.addOval(-120, 200, 40, 190, y += 47);
        this.loginEntryPanel.addLabel(false, (byte) -93, 4, 190, STRINGS[234], y - 10);
        // addTextInput(..., masked = true, ...) — the username field IS masked here.
        this.usernameField = this.loginEntryPanel.addPasswordField(n - 3845, 20, 200, false, 10 + y, 4, 40, true, 190);

        // Ok button (back at the higher row).
        this.loginEntryPanel.addOval(-90, 120, 25, 410, y -= 55);
        this.loginEntryPanel.addLabel(false, (byte) -127, 4, 410, STRINGS[231], y); // "Ok"
        this.loginOkButton = this.loginEntryPanel.addProgressWidget(410, 120, y, -94, 25);

        // Cancel button.
        this.loginEntryPanel.addOval(n - 3952, 120, 25, 410, y += 30);
        this.loginEntryPanel.addLabel(false, (byte) -89, 4, 410, STRINGS[121], y); // "Cancel"
        this.loginCancelButton = this.loginEntryPanel.addProgressWidget(410, 120, y, -120, 25);

        // Give the first entry field (obf ng -> passwordField) initial keyboard focus.
        this.loginEntryPanel.setFocus(this.passwordField, -105);
        y += 30;
    }

    /**
     * Per-frame login-screen input handler: drives the welcome screen and the
     * username/password entry sub-state, and fires {@link #loginUser} when the user
     * submits credentials.
     *
     * {@code loginScreenMode} (obf {@code Xd}) selects the sub-screen:
     *   0 = welcome screen (click to begin), 2 = username/password entry.
     *
     * obf: void x(int)  [proposed: drawLoginInput]
     */
    private final void drawLoginInput(int n) {
        // M3 TEST HARNESS HOOK: one-shot auto-login driven by env vars
        // (RSC_AUTOLOGIN_USER / RSC_AUTOLOGIN_PASS). Lets the headless Xvfb bring-up
        // exercise the real loginUser() handshake without simulating keyboard/mouse.
        // Guarded so it fires exactly once and only when the env vars are present.
        if (!this.autoLoginDone) {
            String au = System.getenv("RSC_AUTOLOGIN_USER");
            String ap = System.getenv("RSC_AUTOLOGIN_PASS");
            if (au != null && ap != null) {
                this.autoLoginDone = true;
                this.username = au;
                this.password = ap;
                this.worldIndex = 2;
                this.loginUser(-12, this.username, this.password, false);
                return;
            }
        }
        // Mark login screen as needing a redraw (skip on the no-op param value).
        if (n != 2) {
            this.loginScreenRedraw = true;
        }
        // Count down the "world full" cooldown.
        if (this.worldFullTimeout > 0) {
            --this.worldFullTimeout;
        }

        // obf: if (~this.Xd != -1)  <=>  this.Xd != 0  (loginScreenMode != 0)
        if (this.loginScreenMode != 0) {
            // --- username/password entry sub-screen ---
            // obf: if (-3 != ~this.Xd) return;  <=>  if (loginScreenMode != 2) return;
            if (this.loginScreenMode != 2) {
                return;
            }
            // Forward mouse state to the entry panel.
            this.loginEntryPanel.handleMouseInput(this.lastMouseButtonDown, this.mouseX, -9989, this.Cf, this.mouseY);

            // Cancel button -> back to the welcome screen.
            if (this.loginEntryPanel.wasActivated((byte) -104, this.loginCancelButton)) {
                this.loginScreenMode = 0;
            }
            // Enter pressed in the first field -> advance focus to the second field.
            if (this.loginEntryPanel.wasActivated((byte) -100, this.passwordField)) {
                this.loginEntryPanel.setFocus(this.usernameField, -88);
            }
            // Submit when Enter is pressed in the second field, or the Ok button clicked.
            if (!this.loginEntryPanel.wasActivated((byte) -114, this.usernameField)
                && !this.loginEntryPanel.wasActivated((byte) -105, this.loginOkButton)) {
                return;
            }
            // Field identity comes from the bytecode's read targets (ng=password, Ih=username).
            this.password = this.loginEntryPanel.getFieldText(this.passwordField, n + 2);
            this.username = this.loginEntryPanel.getFieldText(this.usernameField, 4);
            this.worldIndex = 2; // try alternate-port world by default on manual login
            this.loginUser(-12, this.username, this.password, false);
            return;
        }

        // --- welcome sub-screen: wait for the "Click here to login" button ---
        this.loginWelcomePanel.handleMouseInput(this.lastMouseButtonDown, this.mouseX, -9989, this.Cf, this.mouseY);
        if (!this.loginWelcomePanel.wasActivated((byte) -98, this.loginButton)) {
            return;
        }
        // Enter the username/password entry sub-screen and clear all login fields.
        this.loginScreenMode = 2;
        this.loginEntryPanel.setFieldText(this.loginTitleControl, "", n ^ 27640);
        this.loginEntryPanel.setFieldText(this.loginPromptControl, STRINGS[65], n + 27640); // "Please enter your username and password"
        this.loginEntryPanel.setFieldText(this.passwordField, "", n ^ 27640);
        this.loginEntryPanel.setFieldText(this.usernameField, "", 27642);
        this.loginEntryPanel.setFocus(this.passwordField, n ^ -92);
    }

    /**
     * Handle a dropped/closed server connection (the SV_CLOSE_CONNECTION /
     * lost-socket path). Cancels the system-update countdown, then either resets to
     * the login screen (if mid-session) or starts a silent auto-relogin.
     *
     * If we are still logged in ({@code loggedInState} != 0) just reset the login
     * bookkeeping; otherwise log "Lost connection", arm 10 auto-relogin attempts and
     * call {@link #loginUser} in reconnect mode.
     *
     * @param dummy anti-tamper constant (callers pass 116/123, both > 59)
     *
     * obf: void u(int)  [proposed: closeConnection]
     */
    private final void closeConnection(int dummy) {
        this.systemUpdateTimer = 0; // cancel any pending "system update" countdown
        if (dummy <= 59) {
            // Dead for the real call paths (dummy is always > 59); kept for fidelity.
            this.sendDialogAnswer(-85);
        }
        // obf: if (~this.bj != -1)  <=>  this.bj != 0  (loggedInState != 0)
        if (this.loggedInState != 0) {
            // Mid-session: reset login state, do not auto-reconnect.
            // obf: this.o(-2)  [deob o(int) -> resetTradeDuelState; resets bj/Xd/qg]
            this.resetTradeDuelState(-2);
        } else {
            System.out.println(STRINGS[76]); // "Lost connection"
            this.worldIndex = 10;
            this.loginUser(-12, this.username, this.password, true);
        }
    }


    // =========================================================================
    // ===== mainloop =====
    // =========================================================================
    //
    // Per-tick game logic, connection keep-alive, logout flow, and the three
    // top-right UI tabs whose obfuscated bodies the skeleton grouped under
    // "mainloop".
    //
    // CORRECTNESS-AUDIT NOTES (re-verified method-by-method against the CLEAN
    // Vineflower base `decompiled/normalized-clean/client.java`, cross-checked
    // vs `decompiled/cfr/client.java`, the mudclient204 oracle, and the
    // OpenRSC Payload235 parser). Several reconstructed bodies in the previous
    // pass were WRONG; the fixes are flagged inline with `// FIX:`.
    //
    // CROSS-CLASS NAMING (per docs/NAMING.md — k=World, lb=Scene):
    //   * field `Ek` has type `lb` => Ek is the Scene  (client.scene).
    //   * field `Hh` has type `k`  => Hh is the World   (client.world).
    //   The skeleton's field table lists these swapped; NAMING.md is honored here.
    //
    // SKELETON-MISLABEL NOTES (verified against the oracle):
    //   * a(boolean,byte)@minimap the skeleton calls "tick"            -> drawUiTabMinimap.
    //   * b(boolean,byte)        the skeleton calls "updateGameState"  -> drawUiTabMagic.
    //   * J(int)                 the skeleton calls "drawSleepScreen"  -> handleGameInput (the real tick).
    // Methods are named for what they actually do; obf signatures are kept in comments.

    /**
     * GameShell stop hook. Tears down the game session: confirms logout to the
     * server (opcode 31) and stops the active sound voice.
     *
     * @param fromShell true when invoked by the shell's stop path (sets the tick marker).
     */
    // obf: final void a(boolean)   [client.SA(]   proposed: onStopGame
    final void onStopGame(boolean fromShell) {
        if (fromShell) {
            tickMarker = -103L;                 // ze: scratch timing marker
        }
        // a(true, 31) -> sendConfirmLogoutAck: opcode 31 (CONFIRM_LOGOUT) + stream teardown.
        sendConfirmLogoutAck(true, 31);
        if (ni == null) {             // ni: active audio voice
            return;
        }
        ni.release();                       // stop/close the sound channel
    }

    /**
     * Send a server keep-alive and pump one inbound packet. Called once per tick
     * by {@link #handleGameInput}. If no packet has arrived for >5s a PING
     * (opcode 67) is sent; then any pending writes are flushed and one inbound
     * packet is read and dispatched.
     *
     * The parameter is an obfuscation magic: it is called as {@code originY(0 - 26345)},
     * i.e. {@code magic == -26345}, which is what makes the embedded arg
     * arithmetic (read length {@code magic+26345 == 0}, dispatch tag
     * {@code magic ^ -26304 == 87}) resolve correctly.
     */
    // obf: void originY(int)   [client.SB(]   proposed: sendHeartbeat
    private final void sendHeartbeat(int magic) {
        long now = Timer.currentTimeMillisCorrected(0);                  // p.a(0) = System.currentTimeMillis()

        // Wi = packetLastRead timestamp (activity timer reused for net liveness).
        if (Jh.hasPacket((byte) 34)) {        // Jh.a(34) = hasPacket(): data arrived
            lastActionTime = now;
        }
        // clean: if (-5001 > ~(now - lastActionTime))  <=>  (now - lastActionTime) > 5000ms idle
        if ((now - lastActionTime) > 5000L) {
            lastActionTime = now;
            Jh.newPacket(67, 0);              // opcode 67 (HEARTBEAT / CL_PING)
            Jh.finishPacket(21294);              // flush packet
        }

        try {
            Jh.writePacket(20, true);           // writePacket(20): flush queued writes
        } catch (IOException ex) {
            closeConnection(123);               // u(...) = "Lost connection" teardown
            return;
        }

        if (!isLoaded((byte) -125)) {     // obf: this.f((byte)-125) [deob f(byte) -> isLoaded] = data ready to read?
            return;
        }
        // readPacket(incomingPacket); arg (magic+26345)==0 selects the real read path.
        int size = Jh.readPacketInto(magic + 26345, mg);
        if (size <= 0) {                        // clean: ~size < -1  <=>  size > 0
            return;
        }
        // Dispatch one server->client packet. obf: this.a(magic ^ -26304, size, mg.a((byte)104)).
        // This is onFriendUpdate (the a(int,int,int) FIRST-level dispatcher), NOT handlePacket directly:
        //   - arg0 = (magic ^ -26304) == 87  -> the auto-logout-suppress sentinel (a == 87).
        //   - arg1 = size                    -> packet body length (forwarded as `b`).
        //   - arg2 = mg.getUnsignedByte()    -> the RAW (still ISAAC-encrypted) opcode byte;
        //            onFriendUpdate decrypts it via clientStream.isaacCommand(507, opcode) before
        //            matching the social opcodes and forwarding the rest to handlePacket.
        // Calling handlePacket directly here was a defect: the opcode byte was never de-ISAAC'd
        // (so handlePacket saw garbage opcodes -> reportError -> disconnect ~1s after login), the
        // 11 social/PM/friend opcodes were skipped, and the a==87 logout-suppress was bypassed.
        incoming.onFriendUpdate(magic ^ -26304, size, mg.getUnsignedByte());
    }

    /**
     * Request a normal logout (opcode 102, LOGOUT). Refused while in / shortly
     * after combat. On success an internal logout timer is armed so the
     * "Logging out..." dialog shows until the server drops us.
     *
     * @param combatGrace usually 0 (the post-combat grace threshold to compare
     *        the combat timer against); passed through from the call site.
     */
    // obf: void B(int)   [client.T(]   proposed: requestLogout
    final void requestLogout(int combatGrace) {
        if (screenMode == 0) {                  // SPLIT-FIELD FIX (class b): clean ~qg==-1 => qg==0 => not logged in (was loggedIn)
            return;
        }
        if (combatTimeout > 450) {              // ai > 450: in combat
            // obf: this.a(false, null, 0, il[421], 0, 0, null, il[41])
            showServerMessage(false, null, 0, STRINGS[421], 0, 0, null, STRINGS[41]); // "@cya@You can't logout during combat!"
            return;
        }
        if (combatGrace < combatTimeout) {      // clean: var1 < ai; var1 is 0 -> within 10s grace
            // obf: this.a(false, null, var1 ^ 0, il[420], 0, 0, null, il[41])
            showServerMessage(false, null, combatGrace ^ 0, STRINGS[420], 0, 0, null, STRINGS[41]); // "@cya@You can't logout for 10 seconds after combat"
            return;
        }
        Jh.newPacket(102, 0);                 // opcode 102 (LOGOUT)
        Jh.finishPacket(21294);                  // flush
        logoutTimeout = 1000;                   // bj: arm "Logging out..." dialog
    }

    /**
     * Abort an in-progress logout: clears the logout timer and shows the
     * "can't logout" notice (server told us the request was rejected).
     */
    // obf: void g(byte)   [client.CB(]   proposed: sendConfirmLogout
    final void sendConfirmLogout(byte unused) {
        logoutTimeout = 0;                      // bj: cancel "Logging out..." dialog
        // obf: this.a(false, null, 0, il[64], 0, 0, null, il[41])
        showServerMessage(false, null, 0, STRINGS[64], 0, 0, null, STRINGS[41]);      // "@cya@Sorry, you can't logout at the moment"
    }

    /**
     * Draw the small modal "Logging out..." dialog box in the centre of the screen.
     */
    // obf: void d(byte)   [client.SD(]   proposed: doLogout
    private final void doLogout(byte unused) {
        li.drawBox(126, (byte) 52, 0, 137, 60, 260);          // drawBox(126,137,260,60, black)
        li.drawBoxEdge(126, 260, 137, 27785, 60, 16777215);       // drawBoxEdge(126,137,260,60, white)
        li.drawStringCenter(256, STRINGS[679], 16777215, 0, 5, 173);   // drawStringCenter("Logging out...",256,173)
    }

    /**
     * Render + service the "Map" (minimap) UI tab in the top-right panel.
     *
     * Draws the rotated minimap landscape sprite plus coloured dots for nearby
     * scenery (cyan), ground items (red), NPCs (yellow) and players (white, or
     * green for friends), then the white centre dot and compass. When menus are
     * enabled, a left-click inside the map walks the player to the clicked tile.
     *
     * NOTE: skeleton mislabels this as "tick". It is drawUiTabMinimap.
     * The byte param is an obf magic (called with 125); only the dead
     * {@code if (var2 <= 119)} anti-tamper branch reads it.
     */
    // obf: void a(boolean,byte)   proposed (skeleton): tick   actual: drawUiTabMinimap
    private final void drawUiTabMinimap(boolean handleMenus, byte unused) {
        int uiX = li.width - 199;              // li.u = surface.width2 (right edge)
        int uiWidth = 156;
        int uiHeight = 152;
        li.drawSprite(-1, 2 + spriteMedia, 3, -49 + uiX);        // drawSprite tab background
        uiX += 40;
        li.drawBox(uiX, (byte) -125, 0, 36, uiHeight, uiWidth);          // drawBox(uiX,36,w,h,black)
        li.setBounds(uiX, uiWidth + uiX, 36 + uiHeight, 36, (byte) 76);    // setBounds(clip rect)

        if (unused <= 119) {     // RENDER-BUG FIX (clean client.java:4462 if(var2<=119)): clear stale health bars
            this.bf = null;
        }

        // Rotation/zoom for the minimap. sd=minimapRandom2 (zoom jitter), ug=ug,
        // Df=minimapRandom1 (rotation offset). cc = SurfaceSprite.sin2048Cache (fixed-point sin/cos).
        // GameCharacter (ta) accessors: .currentX=obf .i, .currentY=obf .originY, .hash=obf .C.
        int zoom = 192 + minimapRandom2;
        int rot = (ug + minimapRandom1) & 255;
        int px = (wi.currentX - 6040) * zoom * 3 / 2048;
        int py = (wi.currentY - 6040) * zoom * 3 / 2048;
        int sinR = SurfaceSprite.sin2048Cache[(1024 - 4 * rot) & 0x3ff];
        int cosR = SurfaceSprite.sin2048Cache[((1024 - 4 * rot) & 0x3ff) + 1024];
        int rx = px * cosR + py * sinR >> 18;   // >>18: 2048*2048 -> divide back (junk shift masked to 18)
        py = -(px * sinR) + py * cosR >> 18;    // (2D rotate the point by -rot)
        px = rx;
        // FIX: landscape minimap sprite id is `spriteMedia - 1`, NOT `uiX - 1`.
        //      obf: this.li.a(-1 + this.tg, ...)   tg = spriteMedia.
        li.drawMinimapSprite(spriteMedia - 1, 36 - (-(uiHeight / 2) + -py), uiWidth / 2 + uiX - px, 842218000, zoom, (64 + rot) & 255);

        // Scenery dots (cyan = 0x00FFFF). eh=objectCount, ye/Se = objectX/objectY, Ug=magicLoc.
        for (int i = 0; i < objectCount; i++) {
            int dy = zoom * (64 + (magicLoc * objectY[i] - wi.currentY)) * 3 / 2048;
            int dx = 3 * ((magicLoc * objectX[i] - (-64 - -wi.currentX)) * zoom) / 2048;
            int rdx = cosR * dx + dy * sinR >> 18;
            dy = cosR * dy + -(sinR * dx) >> 18;
            dx = rdx;
            drawIcon(65535, dx + uiX + uiWidth / 2, (byte) -61, -dy + 36 - -(uiHeight / 2));
        }

        // Ground-item dots (red = 0xFF0000). Ah=groundItemCount, Zf/Ni = groundItemX/Y.
        for (int i = 0; i < groundItemCount; i++) {
            int dx = zoom * ((-wi.currentX + (64 + groundItemX[i] * magicLoc)) * 3) / 2048;
            int dy = zoom * 3 * (-wi.currentY + (64 + magicLoc * groundItemY[i])) / 2048;
            int rdx = cosR * dx + sinR * dy >> 18;
            dy = cosR * dy + -(dx * sinR) >> 18;
            dx = rdx;
            drawIcon(0xFF0000, uiX - (-(uiWidth / 2) + -dx), (byte) -53, uiHeight / 2 + 36 - dy);
        }

        // NPC dots (yellow = 0xFFFF00). Tb=npcsLast, de=npcsLastCount.
        for (int i = 0; i < npcsLastCount; i++) {
            GameCharacter npc = Tb[i];
            int dy = zoom * ((npc.currentY + -wi.currentY) * 3) / 2048;
            int dx = 3 * ((npc.currentX + -wi.currentX) * zoom) / 2048;
            int rdx = dy * sinR - -(dx * cosR) >> 18;
            dy = -(dx * sinR) + cosR * dy >> 18;
            dx = rdx;
            drawIcon(0xFFFF00, uiWidth / 2 + (uiX - -dx), (byte) -93, -dy + uiHeight / 2 + 36);
        }

        // Player dots (white = 0xFFFFFF, green = 0x00FF00 if on the friends list).
        // rg=players-in-view, Yc=in-view player count (clean bound: ~Yc < ~i == i < Yc).
        for (int i = 0; i < this.Yc; i++) {
            GameCharacter player = rg[i];
            int dx = 3 * ((-wi.currentX + player.currentX) * zoom) / 2048;
            int dy = zoom * (player.currentY + -wi.currentY) * 3 / 2048;
            int rdx = dx * cosR + sinR * dy >> 18;
            dy = cosR * dy - dx * sinR >> 18;
            dx = rdx;
            int colour = 0xFFFFFF;
            String name = WorldEntity.trimAndValidateString(player.message, (byte) 82);    // hashed name of this player
            if (name != null) {
                for (int f = 0; f < friendListCount; f++) {            // obf: n.g = friendListCount (Mudclient field)
                    boolean isFriend = name.equals(WorldEntity.trimAndValidateString(Surface.decoyStrings200[f], (byte) 107));
                    if (isFriend && (Fj[f] & 2) != 0) {   // Fj[f]&2 = friend online
                        colour = 0x00FF00;
                        break;
                    }
                }
            }
            drawIcon(colour, dx + (uiX - -(uiWidth / 2)), (byte) -67, -dy + 36 - -(uiHeight / 2));
        }

        // Centre marker (local player) + compass sprite, then restore the full-screen clip.
        li.drawCircle(255, -1057205208, 2, uiHeight / 2 + 36, 0xFFFFFF, uiX - -(uiWidth / 2));   // drawCircle
        li.drawMinimapSprite(spriteMedia + 24, 55, uiX - -19, 842218000, 128, (ug + 128) & 255);
        li.setBounds(0, gameWidth, gameHeight + 12, 0, (byte) 119);     // setBounds(full screen)

        if (!handleMenus) {
            return;
        }
        // Left-click inside the map area -> walk to the corresponding world tile.
        int mx = mouseX - (li.width - 199);
        int my = mouseY - 36;
        if (mx >= 40 && my >= 0 && mx < 196 && my < 152) {
            int z = 192 + minimapRandom2;
            int r = (ug + minimapRandom1) & 255;
            int base = (li.width - 199) + 40;
            // unproject screen offset -> world delta (16384 = 1<<14 fixed point; >>15 == /32768)
            int wy = 16384 * (mouseY - uiHeight / 2 - 36) / (z * 3);
            int wx = 16384 * (mouseX - (base + uiWidth / 2)) / (z * 3);
            int s2 = SurfaceSprite.sin2048Cache[(1024 - r * 4) & 0x3ff];
            int c2 = SurfaceSprite.sin2048Cache[((1024 - r * 4) & 0x3ff) + 1024];
            int rwx = wy * s2 - -(c2 * wx) >> 15;
            wy = c2 * wy - s2 * wx >> 15;
            wx = rwx + wi.currentX;
            wy = wi.currentY - wy;
            if (Cf == 1) {        // Cf == 1: a fresh left-click this tick
                // obf: a(worldY>>7, worldX>>7, localRegionX, localRegionY, false, 8)
                //   [deob a(int,int,int,int,boolean,int) -> drawScrollbar2 (walk-to-tile source)]
                drawScrollbar2(wy / 128, wx / 128, localRegionX, localRegionY, false, 8);
            }
            Cf = 0;
        }
    }

    /**
     * Render + service the "Magic" / "Prayers" UI tab (toggled by tabMagicPrayer).
     *
     * Lists castable spells (colour-coded by rune availability and level) or
     * prayers (colour-coded by level and prayer points), shows the hovered
     * entry's description, and on click either selects a spell to cast or
     * toggles a prayer, sending opcode 60 (PRAYER_ACTIVATED) / 254
     * (PRAYER_DEACTIVATED).
     *
     * NOTE: skeleton mislabels this as "updateGameState". It is drawUiTabMagic.
     * The byte param is an obf magic (called with -74): the embedded
     * {@code var2+74 == 0}, {@code var2^-74 == 0}, {@code var2+88 == 14},
     * {@code var2+17124 == 17050} arithmetic only resolves with -74.
     */
    // obf: void b(boolean,byte)   proposed (skeleton): updateGameState   actual: drawUiTabMagic
    final void drawUiTabMagic(boolean handleMenus, byte unused) {
        int uiX = -199 + li.width;
        int uiY = 36;
        li.drawSprite(-1, spriteMedia + 4, 3, -49 + uiX);        // drawSprite tab background
        int uiWidth = 196;
        int uiHeight = 182;

        // Highlight the active sub-tab header brighter (220) than the inactive (160).
        // leftShade = Magic header (bright when tabMagicPrayer==0); rightShade = Prayer header.
        int leftShade, rightShade;
        leftShade = rightShade = ISAAC.packColor(160, 9570, 160, 160);     // o.a(...) = Surface.rgb2long
        if (tabMagicPrayer != 0) {
            rightShade = ISAAC.packColor(220, 9570, 220, 220);             // prayers tab active
        } else {
            leftShade = ISAAC.packColor(220, 9570, 220, 220);              // magic tab active
        }
        li.drawBoxAlpha(128, uiX, 24, 0, uiY, uiWidth / 2, leftShade);
        li.drawBoxAlpha(128, uiWidth / 2 + uiX, 24, 0, uiY, uiWidth / 2, rightShade);
        li.drawBoxAlpha(128, uiX, 90, 0, uiY + 24, uiWidth, ISAAC.packColor(220, 9570, 220, 220));
        li.drawBoxAlpha(128, uiX, uiHeight - 24 - 90, 0, uiY + 24 + 90, uiWidth, ISAAC.packColor(160, 9570, 160, 160));
        li.drawLineHoriz(uiWidth, 0, uiX, uiY + 24, (byte) 70);     // drawLineHoriz under headers
        li.drawLineVert(uiX - -(uiWidth / 2), 0 + uiY, 0, 24, 0);  // drawLineVert between headers
        li.drawLineHoriz(uiWidth, 0, uiX, uiY + 113, (byte) -92);   // drawLineHoriz under list
        li.drawstringRight(uiWidth / 4 + uiX, STRINGS[16], 0, 0, 4, 16 + uiY);                 // "Magic"
        li.drawstringRight(uiX + uiWidth / 4 + uiWidth / 2, STRINGS[21], 0, 0, 4, 16 + uiY);   // "Prayers"

        if (tabMagicPrayer == 0) {
            // --- Spell list ---
            panelMagic.resetItemCount((byte) 118, controlListMagic);    // clearList
            int row = 0;
            // Scatter-slot field map (obf -> deob, per SocketFactory loader):
            //   spellCount(fa.b)=ImageLoader.sharedIntArray.length, spellLevel(pa.f)=ImageLoader.sharedIntArray,
            //   spellName(ja.L)=BitBuffer.UNUSED_L, spellDesc(oa.a)=NameHash.entryNames,
            //   spellRunesCount(o.p)=ISAAC.unusedP, spellRunesId(oa.d)=NameHash.idTable,
            //   spellRunesReq(da.J)=ClientStream.sharedIntTable2d, runeSprite(ua.Bb)=Surface.unusedIntsBb.
            for (int spell = 0; spell < ImageLoader.sharedIntArray.length; spell++) {   // obf bound fa.b == spellCount
                String colour = STRINGS[20];     // "@yel@" (have all runes)
                for (int rune = 0; rune < ISAAC.unusedP[spell]; rune++) {   // spellRunesRequired count (o.p)
                    int runeId = NameHash.idTable[spell][rune];             // spellRunesId (oa.d)
                    if (!menus.pointInPanel((byte) -70, ClientStream.sharedIntTable2d[spell][rune], runeId)) {
                        colour = STRINGS[15];     // "@whi@" (missing a rune)
                        break;
                    }
                }
                if (ImageLoader.sharedIntArray[spell] > skillCurrent[6]) {         // spellLevel > magic level (pa.f)
                    colour = STRINGS[19];        // "@bla@" (level too low)
                }
                panelMagic.setListItem(row++, null, -116, 0, null,
                        colour + STRINGS[18] + ImageLoader.sharedIntArray[spell] + STRINGS[12] + BitBuffer.UNUSED_L[spell], controlListMagic);
            }
            panelMagic.render((byte) -92);            // drawPanel
            int sel = panelMagic.getHoveredItem(controlListMagic, 17050);          // getListEntryIndex
            if (sel != -1) {
                li.drawstring(STRINGS[18] + ImageLoader.sharedIntArray[sel] + STRINGS[12] + BitBuffer.UNUSED_L[sel], 2 + uiX, uiY + 124, 0xFFFF00, false, 1);
                li.drawstring(NameHash.entryNames[sel], 2 + uiX, 136 + uiY, 0xFFFFFF, false, 0);       // spellDescription (oa.a)
                for (int rune = 0; rune < ISAAC.unusedP[sel]; rune++) {
                    int runeId = NameHash.idTable[sel][rune];
                    li.drawSprite(-1, Surface.unusedIntsBb[runeId] + spriteItem, uiY + 150, 2 + uiX + rune * 44);   // rune icon (ua.Bb)
                    int have = menus.menuHitTest(87, runeId);          // obf b(87, runeId) -> menuHitTest
                    int need = ClientStream.sharedIntTable2d[sel][rune];
                    String s = menus.pointInPanel((byte) -70, need, runeId) ? STRINGS[27] : STRINGS[10]; // "@gre@" : "@red@"
                    li.drawstring(s + have + "/" + need, 2 + (uiX + rune * 44), uiY + 150, 0xFFFFFF, false, 1);
                }
            } else {
                li.drawstring(STRINGS[14], uiX + 2, uiY + 124, 0, false, 1);   // "Point at a spell for a description"
            }
        }

        if (tabMagicPrayer == 1) {
            // --- Prayer list ---
            panelMagic.resetItemCount((byte) 90, controlListMagic);    // clearList
            int row = 0;
            // Scatter-slot field map (obf -> deob, per SocketFactory loader):
            //   prayerCount(t.g)=EntityDef.modelNames.length, prayerLevel(ca.B)=GameModel.sharedScratch,
            //   prayerName(t.h)=EntityDef.modelNames, prayerDesc(h.e)=TextEncoder.scratchStrings,
            //   prayerDrain(fa.c)=ClientIOException.soundC.
            for (int prayer = 0; prayer < EntityDef.modelNames.length; prayer++) {    // obf bound t.g == prayerCount
                String colour = STRINGS[15];     // "@whi@"
                if (skillBase[5] < GameModel.sharedScratch[prayer]) {             // prayer base < prayerLevel (ca.B)
                    colour = STRINGS[19];        // "@bla@"
                }
                if (prayerOn[prayer]) {          // bk[]: prayer currently active
                    colour = STRINGS[27];        // "@gre@"
                }
                panelMagic.setListItem(row++, null, -113, 0, null,
                        colour + STRINGS[18] + GameModel.sharedScratch[prayer] + STRINGS[12] + EntityDef.modelNames[prayer], controlListMagic);
            }
            panelMagic.render((byte) -7);             // drawPanel
            int sel = panelMagic.getHoveredItem(controlListMagic, 17050);
            if (sel != -1) {
                li.drawstringRight(uiX - -(uiWidth / 2), STRINGS[18] + GameModel.sharedScratch[sel] + STRINGS[12] + EntityDef.modelNames[sel], 0xFFFF00, 0, 1, uiY + 130);
                li.drawstringRight(uiX - -(uiWidth / 2), TextEncoder.scratchStrings[sel], 0xFFFFFF, 0, 0, 145 + uiY);    // prayerDescription (h.e)
                li.drawstringRight(uiX - -(uiWidth / 2), STRINGS[26] + ClientIOException.soundC[sel], 0, 0, 1, 160 + uiY);   // "Drain rate: " (fa.c)
            } else {
                li.drawstring(STRINGS[11], uiX - -2, uiY + 124, 0, false, 1);   // "Point at a prayer for a description"
            }
        }

        if (!handleMenus) {
            return;
        }
        int mx = mouseX - (li.width - 199);   // li.u = Surface.width
        int my = mouseY - 36;
        if (mx < 0 || my < 0 || mx >= 196 || my >= 182) {
            return;
        }
        // handleMouse(mouseButton, mouseY, junk, mouseLastButton, mouseX) on the magic panel.
        // obf: Mc.b(mouseButtonDown, my+36, -9989, lastMouseButtonDown, mx + (surface.u-199)).
        panelMagic.handleMouseInput(mouseButton, mouseY, -9989, mouseLastButton, mouseX);

        // Header click toggles between Magic (left) and Prayers (right).
        if (my <= 24 && Cf == 1) {
            if (mx < 98 && tabMagicPrayer == 1) {
                tabMagicPrayer = 0;
                panelMagic.clearList(controlListMagic, 14);    // resetListProps
            } else if (mx > 98 && tabMagicPrayer == 0) {
                tabMagicPrayer = 1;
                panelMagic.clearList(controlListMagic, 14);
            }
        }

        // Click a spell -> select it (level + rune checks first).
        if (Cf == 1 && tabMagicPrayer == 0) {
            int sel = panelMagic.getHoveredItem(controlListMagic, 17050);
            if (sel != -1) {
                if (skillCurrent[6] >= ImageLoader.sharedIntArray[sel]) {     // magic level OK (pa.f)
                    int rune;
                    for (rune = 0; rune < ISAAC.unusedP[sel]; rune++) {
                        int runeId = NameHash.idTable[sel][rune];
                        if (!menus.pointInPanel((byte) -70, ClientStream.sharedIntTable2d[sel][rune], runeId)) {
                            // FIX: missing reagents is il[25], not il[24] (indices were swapped).
                            // obf: this.a(false, null, 0, il[25], 0, 0, null, null)
                            showServerMessage(false, null, 0, STRINGS[25], 0, 0, null, null);   // "You don't have all the reagents you need for this spell"
                            rune = -1;
                            break;
                        }
                    }
                    if (rune == ISAAC.unusedP[sel]) {
                        af = sel;
                        selectedItemInventoryIndex = -1;
                    }
                } else {
                    // FIX: magic-level-too-low is il[24], not il[25].
                    // obf: this.a(false, null, var2 + 74, il[24], 0, 0, null, null)  [var2 == unused == -74]
                    showServerMessage(false, null, unused + 74, STRINGS[24], 0, 0, null, null);   // "Your magic ability is not high enough for this spell"
                }
            }
        }

        // Click a prayer -> toggle it; sends PRAYER_ACTIVATED (60) / PRAYER_DEACTIVATED (254).
        if (Cf == 1 && tabMagicPrayer == 1) {
            int sel = panelMagic.getHoveredItem(controlListMagic, 17050);
            if (sel != -1) {
                if (skillBase[5] < GameModel.sharedScratch[sel]) {   // ca.B = prayerLevel
                    // obf: this.a(false, null, 0, il[23], 0, 0, null, null)
                    showServerMessage(false, null, 0, STRINGS[23], 0, 0, null, null);   // "Your prayer ability is not high enough for this prayer"
                } else if (skillCurrent[5] == 0) {
                    // obf: this.a(false, null, 0, il[28], 0, 0, null, null)
                    showServerMessage(false, null, 0, STRINGS[28], 0, 0, null, null);   // "You have run out of prayer points..."
                } else if (!prayerOn[sel]) {
                    Jh.newPacket(60, 0);              // opcode 60 (PRAYER_ACTIVATED)
                    Jh.outBuffer.putByte(sel);          // putByte(prayerId)
                    Jh.finishPacket(21294);              // flush
                    prayerOn[sel] = true;
                    sound.playSound(-79, STRINGS[22]);         // obf a(-79, il[22]) -> playSound "prayeron"
                } else {
                    Jh.newPacket(254, 0);             // opcode 254 (PRAYER_DEACTIVATED)
                    Jh.outBuffer.putByte(sel);          // putByte(prayerId)
                    Jh.finishPacket(21294);              // flush
                    prayerOn[sel] = false;
                    sound.playSound(-117, STRINGS[17]);         // obf a(-117, il[17]) -> playSound "prayeroff"
                }
            }
        }
        Cf = 0;
    }

    // -------------------------------------------------------------------------
    // drawUiTabStats  — obf: void c(boolean,int)   (clean client.java:12938-13170)
    // -------------------------------------------------------------------------

    /**
     * Render + service the qc==3 stats/skills tab drawer (the right-edge panel).
     * Two sub-tabs selected by {@code zd}: 0 = Skills (per-skill cur/base, quest
     * points, fatigue, combat styles, total level, combat level, and a hovered
     * skill's XP / next-level XP), 1 = Quests (scroll list of quest names tinted
     * by completion flag {@code fi}).
     *
     * Faithful de-obfuscation of clean {@code c(boolean var1, int var2)} using
     * {@link #drawUiTabMagic} as the aliasing template.  Field map (obf -> deob):
     *   li.u->li.width; o.a(..)->ISAAC.packColor; il[N]->STRINGS[N]; tg->spriteMedia;
     *   zd->zd; Vk->Vk; oh->oh; cg->cg; ii->ii; vg->vg; Ld->Ld; Fc->Fc; Ej->Ej;
     *   Ak->Ak; ti->experienceTable; Te->Te; fi->fi; fe->Af; lk->lk; wi.s->wi.level;
     *   I->mouseX; xb->mouseY; Bb->mouseButtonDown; Qb->lastMouseButtonDown; Cf->Cf.
     * The opaque predicate {@code vh} (always false) and the per-method profiling
     * counter {@code Pe} are dropped, as in drawUiTabMagic.
     */
    private final void drawUiTabStats(boolean handleMenus, int param) {
        int uiX = li.width - 199;        // var3
        int uiY = 36;                     // var4
        li.drawSprite(-1, spriteMedia + 3, 3, uiX - 49);   // clean li.b(-1, tg-(-3), 3, var3-49)
        int uiWidth = 196;                // var5
        int uiHeight = 275;               // var6

        // Active sub-tab header brighter (220) than the inactive (160).
        int leftShade, rightShade;        // var7 (Skills), var8 (Quests)
        leftShade = rightShade = ISAAC.packColor(160, 9570, 160, 160);
        if (this.zd != 0) {               // ~zd != -1  → Quests active
            rightShade = ISAAC.packColor(220, param ^ 9570, 220, 220);
        } else {
            leftShade = ISAAC.packColor(220, 9570, 220, 220);
        }
        li.drawBoxAlpha(128, uiX, 24, 0, uiY, uiWidth / 2, leftShade);
        li.drawBoxAlpha(128, uiX + (uiWidth / 2), 24, 0, uiY, uiWidth / 2, rightShade);
        li.drawBoxAlpha(128, uiX, uiHeight - 24, 0, uiY + 24, uiWidth, ISAAC.packColor(220, 9570, 220, 220));
        li.drawLineHoriz(uiWidth, 0, uiX, uiY + 24, (byte) -61);
        li.drawLineVert(uiX + (uiWidth / 2), 0 + uiY, 0, 24, param);
        li.drawstringRight(uiWidth / 4 + uiX, STRINGS[356], 0, 0, 4, uiY + 16);                  // "Skills" header
        li.drawstringRight(uiX + (uiWidth / 4 + (uiWidth / 2)), STRINGS[351], 0, 0, 4, uiY + 16); // "Quests" header

        if (this.zd == 0) {
            // ---- Skills sub-tab ----
            int y = 72;                   // var9
            li.drawstring(STRINGS[355], uiX + 5, y, 0xFFFF00, false, 3);
            int hovered = -1;             // var10
            y += 13;

            // 9 rows × two columns (left i, right i+9).
            for (int i = 0; i < 9; i++) {
                int col = 0xFFFFFF;
                if (mouseX > uiX + 3 && y - 11 <= mouseY && y + 2 > mouseY && uiX + 90 > mouseX) {
                    col = 0xFF0000;
                    hovered = i;
                }
                li.drawstring(this.Vk[i] + STRINGS[350] + this.oh[i] + "/" + this.cg[i], uiX + 5, y, col, false, 1);
                col = 0xFFFFFF;
                if (mouseX >= uiX + 90 && mouseY >= y - 11 - 13 && y - 13 + 2 > mouseY && mouseX < uiX + 196) {
                    col = 0xFF0000;
                    hovered = 9 + i;
                }
                li.drawstring(this.Vk[9 + i] + STRINGS[350] + this.oh[9 + i] + "/" + this.cg[9 + i],
                        uiX + uiWidth / 2 - 5, y - 13, col, false, 1);
                y += 13;
            }

            li.drawstring(STRINGS[358] + this.ii, uiX - 5 + (uiWidth / 2), y - 13, 0xFFFFFF, false, 1);   // quest points
            y += 12;
            li.drawstring(STRINGS[360] + this.vg * 100 / 750 + "%", uiX + 5, y - 13, 0xFFFFFF, false, 1); // fatigue
            y += 8;
            li.drawstring(STRINGS[348], uiX + 5, y, 0xFFFF00, false, 3);
            y += 12;

            // combat styles: 3 rows × up to two columns (Ld[i] / Ld[i+3])
            for (int i = 0; i < 3; i++) {
                li.drawstring(this.Ld[i] + STRINGS[350] + this.Fc[i], uiX + 5, y, 0xFFFFFF, false, 1);
                if (i < 2) {
                    li.drawstring(this.Ld[i + 3] + STRINGS[350] + this.Fc[3 + i], uiWidth / 2 + (uiX + 25), y, 0xFFFFFF, false, 1);
                }
                y += 13;
            }

            y += 6;
            li.drawLineHoriz(uiWidth, 0, uiX, y - 15, (byte) 124);

            if (hovered == -1) {
                // nothing hovered: totals
                li.drawstring(STRINGS[352], uiX + 5, y, 0xFFFF00, false, 1);
                y += 12;
                int total = 0;            // var11
                for (int s = 0; s < 18; s++) {
                    total += this.cg[s];
                }
                li.drawstring(STRINGS[347] + total, uiX + 5, y, 0xFFFFFF, false, 1);   // total level
                y += 12;
                li.drawstring(STRINGS[354] + this.wi.level, uiX + 5, y, 0xFFFFFF, false, 1);   // combat level (clean wi.s)
                y += 12;
            } else {
                // a skill is hovered: show its XP / next-level XP
                li.drawstring(this.Ej[hovered] + STRINGS[346], uiX + 5, y, 0xFFFF00, false, 1);
                y += 12;
                int nextExp = this.experienceTable[0];   // var11
                for (int e = 0; e < 98; e++) {
                    if (this.experienceTable[e] <= this.Ak[hovered]) {
                        nextExp = this.experienceTable[e + 1];
                    }
                }
                li.drawstring(STRINGS[357] + this.Ak[hovered] / 4, uiX + 5, y, 0xFFFFFF, false, 1);   // current XP
                y += 12;
                li.drawstring(STRINGS[359] + nextExp / 4, uiX + 5, y, 0xFFFFFF, false, 1);            // next-level XP
            }
        }

        if (this.zd == 1) {
            // ---- Quests sub-tab ----  (~zd == -2)
            this.Af.resetItemCount((byte) 89, this.lk);
            this.Af.setListItem(0, null, -121, 0, null, STRINGS[353], this.lk);   // list header
            for (int q = 0; q < 50; q++) {                                       // ~var23 > -51  → q < 50
                this.Af.setListItem(q + 1, null, param - 82, 0, null,
                        (this.fi[q] ? STRINGS[27] : STRINGS[10]) + this.Te[q], this.lk);
            }
            this.Af.render((byte) -18);
        }

        if (handleMenus) {
            int my = mouseY - 36;                                 // var4
            int mx = -li.width - (-199 - mouseX);                 // var3 = mouseX - (li.width - 199)
            if (mx >= 0 && my >= 0 && mx < uiWidth && my < uiHeight) {
                if (this.zd == 1) {                               // ~zd == -2
                    this.Af.handleMouseInput(this.mouseButtonDown, 36 + my, -9989, this.lastMouseButtonDown, mx + li.width - 199);
                }
                if (my <= 24 && this.Cf == 1) {                   // ~Cf == -2  (left-click header)
                    if (mx >= 99) {                               // ~var3 <= -99
                        if (mx <= 99) {                           // ~var3 >= -99  → exactly the divider, no change
                            return;
                        }
                        this.zd = 1;                              // right half → Quests
                        return;
                    }
                    this.zd = 0;                                  // left half → Skills
                }
            }
        }
    }

    /**
     * Render + service the generic "enter amount / answer" modal dialog and,
     * once the player confirms, flush the queued client action it was gathering.
     *
     * The dialog kind is held in `inputDialogType` (gc). On confirm the typed
     * number is parsed and the matching packet is sent (gc -> action):
     *   3 -> bank withdraw (opcode 22), 4 -> bank deposit (23),
     *   5 -> shop buy (236), 6 -> shop sell (221), 9 -> skip tutorial (84);
     *   1/2/7/8 route through intra-class quantity/item-action wrappers.
     *
     * IMPORTANT: this method is dual-purpose.
     *   * Called as c(-43) from the panel render loop: renders the box AND
     *     services the Ok/Cancel buttons (the {@code if (var1 == -43)} gate).
     *   * Called as c(-97) elsewhere: only flushes a confirmed action; the
     *     bottom button hit-test is skipped.
     * The embedded packet keys (393 for putShort, -422797528 for putInt,
     * 21294 for flush) are anti-tamper guards in Buffer.e/Buffer.b and only
     * resolve to those exact values when var1 == -43 — see Buffer.e:
     * {@code if (n2 != 393) return;} and Buffer.b: {@code if (n2 != -422797528) inject byte}.
     */
    // obf: void c(byte)   [client.AB(]   proposed: sendQueuedActions   (dialogKind == gc)
    private final void sendQueuedActions(byte var1) {
        // --- Confirmation path: a value has been submitted (inputTextFinal non-empty) or OK was
        //     latched last tick (vk = inputDialogConfirmed). clean: !(inputTextFinal.length()<=0 && !vk). ---
        if (inputTextFinal.length() > 0 || inputDialogConfirmed) {
            String value = inputTextFinal.trim();
            inputTextCurrent = "";
            inputTextFinal = "";

            // gc (inputDialogType) selects which queued action to flush. The bare a()/b()/c()
            // wrappers build their own packets. ae[Rd]=bank slot item id;
            // Rj[Di]/Jf[Di]=selected shop slot item id / price.
            if (inputDialogType == 1) {             // generic "enter amount" wrapper
                try {
                    // obf: this.a(amount, (byte)9, this.ji)  [a(int,byte,int) -> drawTradeConfirm]
                    drawTradeConfirm(Integer.parseInt(value), (byte) 9, dialogItemId);
                } catch (NumberFormatException ignored) {
                }
            } else if (inputDialogType == 2) {      // -> c(amount, 124, itemId) wrapper
                try {
                    // obf: this.c(amount, (byte)124, this.ji)  [c(int,byte,int) -> sendTradeOffer]
                    sendTradeOffer(Integer.parseInt(value), (byte) 124, dialogItemId);
                } catch (NumberFormatException ignored) {
                }
            } else if (inputDialogType == 3) {      // Bank withdraw: opcode 22 (BANK_WITHDRAW).
                try {
                    int itemId = (bankSelectedSlot >= 0) ? bankItems[bankSelectedSlot] : -1;
                    int amount = Integer.parseInt(value);
                    Jh.newPacket(22, 0);
                    Jh.outBuffer.putShort(itemId);              // putShort(itemId)
                    Jh.outBuffer.putInt(amount);       // putInt(amount)
                    Jh.outBuffer.putInt(0x12345678);   // putInt(magic/checksum)
                    Jh.finishPacket(21294);                      // flush
                } catch (NumberFormatException ignored) {
                }
            } else if (inputDialogType == 4) {      // Bank deposit: opcode 23 (BANK_DEPOSIT).
                try {
                    // clean inverts the ternary: (Rd < 0) ? -1 : ae[Rd]  (same as withdraw).
                    int itemId = (bankSelectedSlot < 0) ? -1 : bankItems[bankSelectedSlot];
                    int amount = Integer.parseInt(value);
                    Jh.newPacket(23, 0);
                    Jh.outBuffer.putShort(itemId);              // putShort(itemId)   [var1+436 == 393]
                    Jh.outBuffer.putInt(amount);       // putInt(amount)
                    Jh.outBuffer.putInt(0x87654321);   // putInt(magic/checksum)
                    Jh.finishPacket(21294);                      // flush
                } catch (NumberFormatException ignored) {
                }
            } else if (inputDialogType == 6) {      // Shop sell: opcode 221 (SHOP_SELL).
                try {
                    // FIX: guard is `!= -1` (clean: ~Rj[Di] != 0), not `!= 0`.
                    if (shopSelectedItemId[shopSelectedSlot] != -1) {
                        int amount = Integer.parseInt(value);
                        Jh.newPacket(221, 0);
                        Jh.outBuffer.putShort(shopSelectedItemId[shopSelectedSlot]);    // item id
                        Jh.outBuffer.putShort(shopSelectedItemPrice[shopSelectedSlot]); // price
                        Jh.outBuffer.putShort(amount);          // amount   [var1+436 == 393]
                        Jh.finishPacket(21294);
                    }
                } catch (NumberFormatException ignored) {
                }
            } else if (inputDialogType == 7) {      // -> b(109, amount, itemId) wrapper
                try {
                    // obf: this.b(109, amount, this.ck)  [b(int,int,int) -> sendDuelItems]
                    sendDuelItems(109, Integer.parseInt(value), dialogItemId2);
                } catch (NumberFormatException ignored) {
                }
            } else if (inputDialogType == 8) {      // -> a(itemId, amount, -78) wrapper
                try {
                    // obf: this.a(this.ck, amount, (byte)-78)  [a(int,int,byte) -> sendDuelOffer]
                    sendDuelOffer(dialogItemId2, Integer.parseInt(value), (byte) -78);
                } catch (NumberFormatException ignored) {
                }
            } else if (inputDialogType == 9) {      // Skip tutorial: opcode 84 (SKIP_TUTORIAL), no payload.
                Jh.newPacket(84, 0);
                Jh.finishPacket(21294);
            } else {                                // case 5 (and clean's fall-through): Shop buy, opcode 236.
                try {
                    // FIX: guard is `!= -1` (clean: ~Rj[Di] != 0), not `!= 0`.
                    if (shopSelectedItemId[shopSelectedSlot] != -1) {
                        int amount = Integer.parseInt(value);
                        Jh.newPacket(236, 0);
                        Jh.outBuffer.putShort(shopSelectedItemId[shopSelectedSlot]);    // item id   [var1^-420 == 393]
                        Jh.outBuffer.putShort(shopSelectedItemPrice[shopSelectedSlot]); // price
                        Jh.outBuffer.putShort(amount);
                        Jh.finishPacket(21294);                  // flush   [var1+21337 == 21294]
                    }
                } catch (NumberFormatException ignored) {
                }
            }
            inputDialogType = 0;
            return;
        }

        // --- Render path (still waiting on input) ---
        // For numeric dialog kinds (gc 1..8) strip any non-digit chars from the live text.
        int boxX;
        if (inputDialogType >= 1 && inputDialogType <= 8) {
            StringBuilder digits = new StringBuilder();
            for (int n = 0; n < inputTextCurrent.length(); n++) {
                char ch = inputTextCurrent.charAt(n);
                if (Character.isDigit(ch)) {
                    digits.append(ch);
                }
            }
            inputTextCurrent = digits.toString();
        }
        boxX = 256 - inputDialogWidth / 2;

        // Draw the modal box, the prompt lines, and (only when called as c(-43)) Ok / Cancel.
        int boxY = 180 - inputDialogHeight / 2;
        li.drawBox(boxX, (byte) -103, 0, boxY, inputDialogHeight, inputDialogWidth);       // drawBox
        li.drawBoxEdge(boxX, inputDialogWidth, boxY, 27785, inputDialogHeight, 0xFFFFFF);       // drawBoxEdge   [var1^-27812 == 27785]
        int lineH = li.textHeight(1, 1);            // text height   [var1+508305395 == 508305352]
        int btnH = li.textHeight(4, 4);
        int step = lineH + 2;
        for (int n = 0; n < inputDialogLines.length; n++) {
            li.drawstringRight(256, inputDialogLines[n], 0xFFFF00, 0, 1, step * n + (5 + boxY) - -lineH);
        }
        if (inputDialogMask) {                   // Bd: password-style masking
            li.drawstringRight(256, inputTextCurrent + "*", 0xFFFFFF, 0, 4, boxY + (5 + step * inputDialogLines.length) - (-3 + -btnH));
        }

        // The Ok/Cancel buttons and the click-outside dismiss only run for the c(-43) call.
        if (var1 != -43) {
            return;
        }
        int btnY = lineH + (8 + boxY) - (-(inputDialogLines.length * step) + (-btnH - 2));
        // "Ok" button (left @ x=230..248).
        int colour = 0xFFFFFF;
        if (mouseX > 230 && mouseX < 248 && btnY - lineH < mouseY && btnY > mouseY) {
            colour = 0xFFFF00;
            if (Cf != 0) {
                inputDialogConfirmed = true;     // vk: latch confirm
                Cf = 0;
                inputTextFinal = inputTextCurrent;
            }
        }
        li.drawstring(STRINGS[122], 230, btnY, colour, false, 1);   // "Ok"
        // "Cancel" button (right @ x=264..304).
        colour = 0xFFFFFF;
        if (mouseX > 264 && mouseX < 304 && btnY - lineH < mouseY && btnY > mouseY) {
            colour = 0xFFFF00;
            if (Cf != 0) {
                Cf = 0;
                inputDialogType = 0;
            }
        }
        li.drawstring(STRINGS[121], 264, btnY, colour, false, 1);   // "Cancel"

        // A left-click outside the box also dismisses the dialog.
        if (Cf == 1
                && (mouseX < boxX || mouseX > inputDialogWidth + boxX || mouseY < boxY || mouseY > inputDialogHeight + boxY)) {
            inputDialogType = 0;
            Cf = 0;
        }
    }

    /**
     * The real per-tick game logic (one call per client tick, invoked as J(0)).
     * Drives the connection keep-alive, the logout/combat/idle timers, player &
     * NPC movement interpolation along their waypoint buffers, camera
     * auto-rotate and zoom, the sleep-CAPTCHA word entry (opcode 45, SLEEP_WORD),
     * the chat message tabs / chat-command parsing, mouse-button repeat
     * acceleration, and Ek object animations.
     *
     * NOTE: skeleton mislabels this as "drawSleepScreen". The sleep handling is
     * only one branch; this is handleGameInput (the tick).
     */
    // obf: void J(int)   [client.HD(]   proposed (skeleton): drawSleepScreen   actual: handleGameInput / tick
    private final void handleGameInput(int magic) {
        // 1) System-update countdown (server restart timer).
        if (systemUpdate > 1) {
            systemUpdate--;
        }
        // 2) Connection keep-alive + inbound packet pump.  originY(magic - 26345) == originY(-26345).
        sendHeartbeat(magic + -26345);
        // 3) Logout timer.
        if (logoutTimeout > 0) {
            logoutTimeout--;
        }
        // 4) Auto-logout after long inactivity (idle > 15000, not in combat / logging out).
        if (mouseActionTimeout > 15000 && combatTimeout == 0 && logoutTimeout == 0) {
            mouseActionTimeout -= 15000;
            requestLogout(magic ^ 0);           // B(0)
            return;
        }
        // 5) Local-player combat state: anim 8/9 means fighting -> hold combat timer high.
        if (wi.animationCurrent == 8 || wi.animationCurrent == 9) {
            combatTimeout = 500;
        }
        if (combatTimeout > 0) {
            combatTimeout--;
        }
        // 6) Character-design panel takes over input while open.
        //    F(86) services the panel and, on accept, sends opcode 235 (PLAYER_APPEARANCE_CHANGE).
        if (Kg) {             // Kg
            sendAppearance(86);                 // F(86)
            return;
        }

        // 7) Interpolate nearby players toward their next waypoint and tick their timers.
        //    GameCharacter (ta) fields: o=waypointCurrent, e=movingStep, y=animationCurrent,
        //    D=animationNext, i=currentX, originY=currentY, k[]=waypointsX, F[]=waypointsY, x=stepCount,
        //    E=messageTimeout, d=bubbleTimeout, mouseX=combatTimer, w=projectileRange.
        for (int i = 0; i < this.Yc; i++) {        // Yc over rg
            GameCharacter c = rg[i];
            int target = (c.waypointCurrent + 1) % 10;
            if (c.movingStep == target) {
                c.animationCurrent = c.animationNext;
            } else {
                int facing = -1;
                int step = c.movingStep;
                int remaining = (step < target) ? (target - step) : ((10 + target) - step);
                int speed = 4;
                if (remaining > 2) {
                    speed = (remaining - 1) * 4;
                }
                // Snap if the next waypoint is too far (teleport) or too many steps queued.
                if (c.waypointsX[step] - c.currentX > magicLoc * 3 || c.waypointsY[step] - c.currentY > magicLoc * 3
                        || c.waypointsX[step] - c.currentX < -magicLoc * 3 || c.waypointsY[step] - c.currentY < -magicLoc * 3
                        || remaining > 8) {
                    c.currentX = c.waypointsX[step];
                    c.currentY = c.waypointsY[step];
                } else {
                    if (c.currentX < c.waypointsX[step]) {
                        facing = 2; c.currentX += speed; c.stepCount++;
                    } else if (c.currentX > c.waypointsX[step]) {
                        c.stepCount++; facing = 6; c.currentX -= speed;
                    }
                    if (c.currentX - c.waypointsX[step] < speed && c.currentX - c.waypointsX[step] > -speed) {
                        c.currentX = c.waypointsX[step];
                    }
                    if (c.currentY < c.waypointsY[step]) {
                        facing = (facing == -1) ? 4 : (facing == 2 ? 3 : 5);
                        c.currentY += speed; c.stepCount++;
                    } else if (c.currentY > c.waypointsY[step]) {
                        c.stepCount++; c.currentY -= speed;
                        facing = (facing == -1) ? 0 : (facing == 2 ? 1 : 7);
                    }
                    if (c.currentY - c.waypointsY[step] < speed && c.currentY - c.waypointsY[step] > -speed) {
                        c.currentY = c.waypointsY[step];
                    }
                }
                if (facing != -1) {
                    c.animationCurrent = facing;
                }
                if (c.currentX == c.waypointsX[step] && c.currentY == c.waypointsY[step]) {
                    c.movingStep = (step + 1) % 10;
                }
            }
            if (c.messageTimeout > 0) c.messageTimeout--;
            if (c.bubbleTimeout > 0) c.bubbleTimeout--;
            if (c.combatTimer > 0) c.combatTimer--;
            // Death-screen respawn message (decremented inside this loop, matching the oracle).
            if (deathScreenTimeout > 0) {
                deathScreenTimeout--;
                if (deathScreenTimeout == 0) {
                    showServerMessage(false, null, 0, STRINGS[629], 0, 0, null, null);   // a(false,null,0,il[629],0,0,null,null): "You have been granted another life..."
                }
                if (deathScreenTimeout == 0) {
                    showServerMessage(false, null, 0, STRINGS[628], 0, 0, null, null);   // a(false,null,0,il[628],0,0,null,null): "You retain your skills..."
                }
            }
        }

        // 8) Interpolate nearby NPCs likewise (NPC id 43 spins continuously while idle).
        for (int i = 0; i < npcsLastCount; i++) {           // de over Tb
            GameCharacter c = Tb[i];
            int target = (c.waypointCurrent + 1) % 10;
            if (c.movingStep == target) {
                if (c.npcIdOrColourBottom == 43) {
                    c.stepCount++;
                }
                c.animationCurrent = c.animationNext;
            } else {
                int facing = -1;
                int step = c.movingStep;
                int remaining = (step < target) ? (target - step) : ((10 + target) - step);
                int speed = 4;
                if (remaining > 2) {
                    speed = (remaining - 1) * 4;
                }
                if (c.waypointsX[step] - c.currentX > magicLoc * 3 || c.waypointsY[step] - c.currentY > magicLoc * 3
                        || c.waypointsX[step] - c.currentX < -magicLoc * 3 || c.waypointsY[step] - c.currentY < -magicLoc * 3
                        || remaining > 8) {
                    c.currentX = c.waypointsX[step];
                    c.currentY = c.waypointsY[step];
                } else {
                    if (c.currentX < c.waypointsX[step]) {
                        c.stepCount++; c.currentX += speed; facing = 2;
                    } else if (c.currentX > c.waypointsX[step]) {
                        facing = 6; c.stepCount++; c.currentX -= speed;
                    }
                    if (c.currentX - c.waypointsX[step] < speed && c.currentX - c.waypointsX[step] > -speed) {
                        c.currentX = c.waypointsX[step];
                    }
                    if (c.currentY < c.waypointsY[step]) {
                        facing = (facing == -1) ? 4 : (facing == 2 ? 3 : 5);
                        c.currentY += speed; c.stepCount++;
                    } else if (c.currentY > c.waypointsY[step]) {
                        facing = (facing == -1) ? 0 : (facing == 2 ? 1 : 7);
                        c.currentY -= speed; c.stepCount++;
                    }
                    if (c.currentY - c.waypointsY[step] < speed && c.currentY - c.waypointsY[step] > -speed) {
                        c.currentY = c.waypointsY[step];
                    }
                }
                if (facing != -1) {
                    c.animationCurrent = facing;
                }
                if (c.currentX == c.waypointsX[step] && c.currentY == c.waypointsY[step]) {
                    c.movingStep = (step + 1) % 10;
                }
            }
            if (c.bubbleTimeout > 0) c.bubbleTimeout--;
            if (c.messageTimeout > 0) c.messageTimeout--;
            if (c.combatTimer > 0) c.combatTimer--;
        }

        // 9) Sleep-word delay bookkeeping (key-activity counters off the sleep tab).
        //    obf: nb.g = DataStore key-typed counter, da.M = ClientStream special-key counter.
        if (showUiTab != 2) {
            if (DataStore.unused_g > 0) sleepWordDelayTimer++;        // nb.g
            if (ClientStream.profCountM > 0) sleepWordDelayTimer = 0; // da.M
            DataStore.unused_g = 0;     // nb.g = 0
            ClientStream.profCountM = 0; // da.M = 0
        }
        // Tick projectile ranges on players.
        for (int i = 0; i < this.Yc; i++) {
            GameCharacter c = rg[i];
            if (c.projectileRange > 0) c.projectileRange--;
        }
        if (sleepWordDelayTimer > 20) {
            sleepWordDelayTimer = 0;
            sleepWordDelay = false;
        }

        // 10) Camera smooth-follow + auto-rotate of the local player.
        //     clean: if (!Td) { snap; autorotate; followY; followX } else { snap }.
        //     Td == Td (when set, only the hard snap happens).
        if (!Td) {
            if (Math.abs(kg - wi.currentX) > 500 || Math.abs(Si - wi.currentY) > 500) {
                kg = wi.currentX;
                Si = wi.currentY;
            }
            if (Kh) {
                int target = si * 32;
                int delta = target - ug;
                int dir = 1;
                if (delta != 0) {
                    Wc++;
                    if (delta > 128) {
                        dir = -1;
                        delta = 256 - delta;
                    } else if (delta > 0) {
                        dir = 1;
                    } else if (delta < -128) {
                        delta = 256 + delta;
                        dir = 1;
                    } else if (delta < 0) {
                        dir = -1;
                        delta = -delta;
                    }
                    ug += ((delta * Wc + 255) / 256) * dir;
                    ug &= 255;
                } else {
                    Wc = 0;
                }
            }
            if (wi.currentY != Si) {
                Si += (wi.currentY - Si) / ((ac - 500) / 15 + 16);
            }
            if (wi.currentX != kg) {
                kg += (wi.currentX - kg) / ((ac - 500) / 15 + 16);
            }
        } else if (kg - wi.currentX < -500 || kg - wi.currentX > 500
                || Si - wi.currentY < -500 || Si - wi.currentY > 500) {
            kg = wi.currentX;
            Si = wi.currentY;
        }

        if (!isSleeping) {                          // clean: if (!Qk)  (Qk = isSleeping)
            // 11) Chat message tab strip along the bottom of the screen.
            //     mouseX=mouseX, mouseY=mouseY, lastMouseButtonDown=mouseLastButton, mouseButtonDown=mouseButton, Oi=gameHeight,
            //     Zh=messageTabSelected, yd=panelMessageTabs, yd.j[]=controlFlashText.
            if (mouseY > gameHeight - 4) {
                if (mouseX > 15 && mouseX < 96 && mouseLastButton == 1) {
                    messageTabSelected = 0;
                }
                if (mouseX > 110 && mouseX < 194 && mouseLastButton == 1) {
                    messageTabSelected = 1;
                    panelMessageTabs.scrollPos[controlListChat] = 999999;
                }
                if (mouseX > 215 && mouseX < 295 && mouseLastButton == 1) {
                    messageTabSelected = 2;
                    panelMessageTabs.scrollPos[controlListQuest] = 999999;
                }
                if (mouseX > 315 && mouseX < 395 && mouseLastButton == 1) {
                    messageTabSelected = 3;
                    panelMessageTabs.scrollPos[controlListPrivate] = 999999;
                }
                if (mouseX > 417 && mouseX < 497 && mouseLastButton == 1) {
                    // FIX: clean (rev 235) sets only these three; the rev-204 `reportAbuseOffence = 0` was removed.
                    inputTextFinal = "";
                    showDialogReportAbuseStep = 1;
                    inputTextCurrent = "";
                }
                mouseLastButton = 0;
                mouseButton = 0;
            }
            // handleMouse(mouseButton, mouseY, junk, mouseLastButton, mouseX) on the chat-tabs panel.
            // obf: yd.b(mouseButtonDown, mouseY, magic-9989, lastMouseButtonDown, mouseX).
            panelMessageTabs.handleMouseInput(mouseButton, mouseY, magic + -9989, mouseLastButton, mouseX);

            if (messageTabSelected > 0 && mouseX >= 494 && mouseY >= gameHeight - 66) {
                mouseLastButton = 0;
            }

            // 12) A chat line was entered -> parse "::" commands or send as chat.
            if (panelMessageTabs.wasActivated((byte) -128, controlListInput)) {     // isClicked(bh)
                String text = panelMessageTabs.getFieldText(controlListInput, 4);   // getText
                panelMessageTabs.setFieldText(controlListInput, "", 27642);         // updateText("")
                if (text.startsWith(STRINGS[627])) {                     // "::"
                    // hj = appletMode; these debug commands are disabled in applet mode.
                    if (text.equalsIgnoreCase(STRINGS[626]) && !appletMode) {        // "::logout"
                        // FIX: clean sends a(true, 31) (sendConfirmLogoutAck), NOT closeConnection.
                        sendConfirmLogoutAck(true, magic ^ 31);   // a(true, 31)
                    } else if (text.equalsIgnoreCase(STRINGS[630]) && !appletMode) { // "::lostcon"
                        closeConnection(116);               // u(116)
                    } else if (text.equalsIgnoreCase(STRINGS[623]) && !appletMode) { // "::closecon"
                        Jh.closeStream(true);               // closeStream()
                    } else {
                        packets.sendCommand(text.substring(2), 120); // opcode 38 (COMMAND): "::" command
                    }
                } else {
                    packets.sendOpcodeString(text, magic + 216);     // b(text, magic+216) -> chat send
                }
            }

            // 13) Decay the chat-message fade timers (100-slot ring, obf pa.g[] @client.java:8945).
            // SPLIT-FIELD FIX (class b): obf `pa.g` is the static ImageLoader.scratchBuf[100] (allocated);
            // it had also been aliased to a never-allocated Mudclient.messageHistoryTimeout -> NPE here.
            for (int i = 0; i < 100; i++) {
                if (ImageLoader.scratchBuf[i] > 0) ImageLoader.scratchBuf[i]--;
            }
            if (deathScreenTimeout != 0) {              // rk != 0
                mouseLastButton = 0;                   // lastMouseButtonDown = 0
            }

            // 14) Trade/duel quantity buttons: accelerate the increment the longer held.
            //     Ti=mouseButtonDownTime, Tk=mouseButtonItemCountIncrement, mouseButtonDown=mouseButton.
            if (!showDialogTrade && !showDialogDuel) {  // !Hk && !Pj
                mouseButtonDownTime = 0;
                mouseButtonItemCountIncrement = 0;
            } else {
                if (mouseButton == 0) {
                    mouseButtonDownTime = 0;
                } else {
                    mouseButtonDownTime++;
                }
                if (mouseButtonDownTime <= 600) {
                    if (mouseButtonDownTime > 450) {
                        mouseButtonItemCountIncrement += 500;
                    } else if (mouseButtonDownTime > 300) {
                        mouseButtonItemCountIncrement += 50;
                    } else if (mouseButtonDownTime <= 150) {
                        if (mouseButtonDownTime <= 50) {
                            if (mouseButtonDownTime > 20 && (mouseButtonDownTime & 5) == 0) {
                                mouseButtonItemCountIncrement++;
                            }
                        } else {
                            mouseButtonItemCountIncrement++;          // 50 < t <= 150
                        }
                    } else {
                        mouseButtonItemCountIncrement += 5;           // 150 < t <= 300
                    }
                } else {
                    mouseButtonItemCountIncrement += 5000;            // t > 600
                }
            }

            // 15) Latch this tick's click (1 = left, 2 = right) for the UI handlers.
            if (mouseLastButton == 1) {                 // ~lastMouseButtonDown == -2
                Cf = 1;
            }
            if (mouseLastButton == 2) {                 // ~lastMouseButtonDown == -3
                Cf = 2;
            }
            Ek.setMouseLoc(0, mouseX, mouseY);             // Ek.a(0, mouseX, mouseY): setMouseLoc (Ek = Scene; deob alias 'world' holds the Scene)
            mouseLastButton = 0;                    // lastMouseButtonDown = 0

            // 16) Camera angle via arrow keys (auto mode steps the discrete 8-way angle,
            //     manual mode nudges the continuous rotation). Z=keyLeft, E=keyRight,
            //     si=si, ug=ug, zf=fogOfWar, Wc=Wc.
            if (Kh) {                 // Kh
                if (Wc == 0 || Td) {   // !(Wc!=0 && !Td)
                    if (keyLeft) {
                        keyLeft = false;
                        si = si + 1 & 7;
                        if (!fogOfWar) {
                            if ((si & 1) == 0) si = 1 + si & 7;
                            for (int i = 0; i < 8; i++) {
                                if (isValidCameraAngle((byte) -125, si)) break;
                                si = 1 + si & 7;
                            }
                        }
                    }
                    if (keyRight) {
                        keyRight = false;
                        si = 7 + si & 7;
                        if (!fogOfWar) {
                            if ((si & 1) == 0) si = si + 7 & 7;
                            for (int i = 0; i < 8; i++) {
                                if (isValidCameraAngle((byte) -116, si)) break;
                                si = si + 7 & 7;
                            }
                        }
                    }
                }
            } else {
                if (keyLeft) {
                    ug = 0xFF & ug + 2;
                }
                if (keyRight) {
                    ug = 0xFF & -2 + ug;
                }
            }

            // 17) Decay the minimap click-walk step counter toward zero (xh = mouseClickXStep).
            if (xh > 0) {
                xh--;
            } else if (xh < 0) {
                xh++;
            }

            // 18) Camera zoom drifts in (in fog-of-war / wilderness) or out otherwise (ac=ac).
            if (fogOfWar && ac > 550) {
                ac -= 4;
            } else if (!fogOfWar && ac < 750) {
                ac += 4;
            }

            // 19) Animated world scenery.
            Ek.scrollTexture(25013, 17);                     // Ek.d(25013, 17): animate fountain (Ek = Scene; deob alias 'world' holds the Scene)
            objectAnimationCount++;                 // qk
            if (objectAnimationCount > 5) {
                objectAnimationCount = 0;
                objectAnimationTorch = (objectAnimationTorch + 1) % 4;   // Nc %4
                objectAnimationFire = (objectAnimationFire + 1) % 3;     // Mg %3
                objectAnimationClaw = (objectAnimationClaw + 1) % 5;     // pj %5
            }
            for (int i = 0; i < eh; i++) {
                int ox = ye[i];                // ye
                int oy = Se[i];                // Se
                if (oy >= 0 && ox >= 0 && oy < 96 && ox < 96 && vc[i] == 74) {
                    hg[i].rotate(0, -31616, 0, 1);  // hg[i].f(...): rotate windmill sails (yaw += 1)
                }
            }

            // 20) Age out expired teleport "bubble" effects (compacting the parallel arrays).
            for (int i = 0; i < teleportBubbleCount; i++) {
                teleportBubbleTime[i]++;
                if (teleportBubbleTime[i] > 50) {
                    teleportBubbleCount--;
                    for (int j = i; j < teleportBubbleCount; j++) {
                        teleportBubbleX[j] = teleportBubbleX[j + 1];
                        teleportBubbleY[j] = teleportBubbleY[j + 1];
                        teleportBubbleTime[j] = teleportBubbleTime[j + 1];
                        teleportBubbleType[j] = teleportBubbleType[j + 1];
                    }
                }
            }
        } else {
            // 21) While asleep (Qk): handle the sleep-word CAPTCHA submit (opcode 45, SLEEP_WORD).
            //     Protocol (Payload235): putByte(delayFlag) THEN putString(word).
            if (inputTextFinal.length() > 0) {
                if (inputTextFinal.equalsIgnoreCase(STRINGS[630]) && !appletMode) {        // "::lostcon"
                    Jh.closeStream(true);               // closeStream()
                } else if (inputTextFinal.equalsIgnoreCase(STRINGS[623]) && !appletMode) { // "::closecon"
                    // FIX: clean sends a(true, 31) (sendConfirmLogoutAck), NOT closeConnection.
                    sendConfirmLogoutAck(true, magic + 31);   // a(true, 31)
                } else {
                    Jh.newPacket(45, 0);              // opcode 45 (SLEEP_WORD)
                    // FIX: delay byte is written FIRST (1 if delay engaged, else 0), then the word.
                    if (sleepWordDelay) {
                        Jh.outBuffer.putByte(1);       // putByte(1)
                    } else {
                        Jh.outBuffer.putByte(0);      // putByte(0)
                        sleepWordDelay = true;
                    }
                    Jh.outBuffer.putString(inputTextFinal);  // putString(word)
                    Jh.finishPacket(21294);
                    inputTextCurrent = "";
                    sleepingStatusText = STRINGS[436];  // "Please wait..."
                    inputTextFinal = "";
                }
            }
            // Clicking the "type the word" box submits "-null-".
            if (mouseLastButton == 1 && mouseY > 275 && mouseY < 310 && mouseX > 56 && mouseX < 456) {
                Jh.newPacket(45, 0);                  // opcode 45 (SLEEP_WORD)
                // FIX: write the delay byte first (0 the first time, 1 thereafter), then the word.
                if (!sleepWordDelay) {
                    Jh.outBuffer.putByte(0);            // putByte(0)
                    sleepWordDelay = true;
                } else {
                    Jh.outBuffer.putByte(1);           // putByte(1)
                }
                Jh.outBuffer.putString(STRINGS[625]);    // putString("-null-")
                Jh.finishPacket(21294);
                sleepingStatusText = STRINGS[436];      // "Please wait..."
                inputTextFinal = "";
                inputTextCurrent = "";
            }
            mouseLastButton = 0;
        }
    }


    // =========================================================================
    // ===== packetout =====
    // =========================================================================
    //
    // Outgoing (client -> server) packet builders. Opcodes are cited from the OpenRSC
    // Payload235Parser (the rev-235 client->server map). The low-level wire helpers used
    // throughout this group are:
    //
    //   clientStream.newPacket(op)             // begin a packet with opcode `op`  (da.b(int,int))
    //   clientStream.sendPacket()              // flush/queue the packet           (da.b(int) -> b(21294))
    //   clientStream.buffer.putByte(v)         // 1 byte                           (tb.c(int,int))
    //   clientStream.buffer.putShort(v)        // 2 bytes, big-endian              (tb.e(int,int))
    //   clientStream.buffer.putInt(v)          // 4 bytes, big-endian              (tb.b(int,int))
    //   clientStream.buffer.putString(s)       // length-prefixed + null-terminated(tb.a(String,int))
    //   StringCodec.encodeAndWrite(buffer, s)  // length byte + char-table-encoded (u.a(int,tb,String))
    //
    // The original methods are wrapped in J++ control-flow obfuscation (opaque predicate
    // `client.vh`, per-method profiling `++counter`, try/catch rethrow via ErrorHandler,
    // anti-tamper guards on the dummy `byte`/`int` params, and junk masks on the buffer
    // writes). All of that is stripped below; only the real logic remains.
    //
    // CROSS-CLASS NAMES (verified against docs/NAMING.md + clean base):
    //   Hh (type k=World)   -> this.world      [route()/pathfinding lives on World]
    //   Ek (type lb=Scene)  -> this.scene
    //   Cf                  -> mouseButtonClick (1=left, 2=right this tick; cleared after use)
    //   The social lists are scattered across static arrays on unrelated obf classes; the
    //   readable instance-style names below mirror the rest of the deob (see oracle GameConnection):
    //     friends: ua.h=names, cb.c=formerNames, ac.z=worlds(str), Fj=online/flags, n.g=count
    //     ignores: l.c=names, ia.a=displayNames, ia.g=formerNames, ua.wb=worlds(str), db.g=count
    //   fa.e (int[]) = GameData.itemStackable;  VALUE SEMANTICS: ==0 => STACKABLE, !=0 => not.

    /**
     * Walk to an explicit destination, optionally walking up to (rather than onto) the
     * target tile. Pathfinds with {@link World#route}, then streams the start tile plus
     * per-step deltas. Sends opcode 16 (WALK_TO_ENTITY) when {@code walkToAction} is set,
     * else opcode 187 (WALK_TO_POINT).
     *
     * @return true if a packet was sent (a route existed), false if unreachable.
     */
    // obf: private final boolean a(int,int,byte,boolean,int,int,int,int,boolean)  [byte param var3 is anti-tamper junk]
    boolean walkTo(int startX, int startY, byte unused, boolean checkObjects,
                           int x1, int y1, int x2, int y2, boolean walkToAction) {
        // route() fills walkPathX (Rg) / walkPathY (pf) and returns the waypoint count, or -1.
        int steps = this.Hh.route(this.walkPathX, x1, y2, this.walkPathY,
                                     startY, startX, y1, x2, checkObjects); // route() is a World method; deob alias 'scene' holds the World
        if (steps == -1) {            // obf: ~steps == 0  ⟺  steps == -1  (no path)
            return false;
        }

        // The last waypoint is our true starting tile this tick (read both arrays at the
        // same index, then drop into the per-step stream).
        int curX = this.walkPathX[--steps]; // obf: var2 = Rg[--var10]
        int curY = this.walkPathY[steps];   // obf: var1 = pf[var10]

        // opcode 16 = WALK_TO_ENTITY (walk-to-action), 187 = WALK_TO_POINT (plain walk)
        this.Jh.newPacket(walkToAction ? 16 : 187, 0);
        this.Jh.outBuffer.putShort(this.Qg + curX); // obf: Qg + var2 (absolute start X)  [SPLIT-FIELD FIX: was regionX (orphan, always 0)]
        this.Jh.outBuffer.putShort(this.zg + curY); // obf: zg + var1 (absolute start Y)  [SPLIT-FIELD FIX: was regionY (orphan, always 0)]

        steps--; // obf: var10-- (UNCONDITIONAL second decrement, before the loop bound)

        // Server-side anti-cheat quirk: for a zero-length action-walk on a tile whose
        // absolute X is a multiple of 5, emit a single (0,0) step.
        if (walkToAction && steps == -1 && (this.Qg + curX) % 5 == 0) {
            steps = 0;
        }
        // Stream waypoint deltas (at most 25), back-to-front, relative to the start tile.
        for (int i = steps; i >= 0 && i > steps - 25; i--) {
            this.Jh.outBuffer.putByte(this.walkPathX[i] - curX); // dx
            this.Jh.outBuffer.putByte(this.walkPathY[i] - curY); // dy
        }
        this.Jh.finishPacket(21294);

        // Remember the click so the yellow "X" walk marker can be drawn.
        this.tj = this.mouseX;   // obf: tj = mouseX
        this.Fd = this.mouseY;   // obf: Fd = mouseY
        this.xh = -24;        // obf: xh = -24
        return true;
    }

    /**
     * Walk-then-interact variant: like {@link #walkTo} but, when the path is blocked and
     * {@code walkToAction} is set, it still synthesises a single-step packet aimed at the
     * target tile so the queued interaction (talk/attack/use) can fire on arrival.
     * Sends opcode 16 (WALK_TO_ENTITY) or 187 (WALK_TO_POINT).
     *
     * @return true once a target is chosen (and a packet is sent).
     */
    // obf: private final boolean a(int,boolean,int,int,int,int,boolean,int,int)  [trailing int param var9 is anti-tamper junk]
    boolean walkToAction(int startX, boolean walkToAction, int destX, int destY,
                                 int x2, int y2, boolean checkObjects, int startY, int unused) {
        int steps = this.Hh.route(this.walkPathX, startX, startY, this.walkPathY,
                                     destX, x2, y2, destY, checkObjects); // route() is a World method; deob alias 'scene' holds the World
        if (steps == -1) {            // obf: ~steps == 0  ⟺  steps == -1  (blocked)
            if (!walkToAction) {
                return false; // plain walk to an unreachable tile: abort
            }
            // Action-walk to a blocked tile: synthesise a one-step path to the target so
            // the interaction still gets queued on the server.
            steps = 1;
            this.walkPathX[0] = startX; // obf: Rg[0] = var1
            this.walkPathY[0] = destY;  // obf: pf[0] = var4
        }
        int curY = this.walkPathY[--steps]; // obf: var5 = pf[--var10]
        int curX = this.walkPathX[steps];   // obf: var3 = Rg[var10]
        steps--;                            // obf: var10-- (unconditional)

        // opcode 16 = WALK_TO_ENTITY (action), 187 = WALK_TO_POINT (plain)
        this.Jh.newPacket(walkToAction ? 16 : 187, 0);
        this.Jh.outBuffer.putShort(this.Qg + curX); // obf: Qg + var3  [SPLIT-FIELD FIX: was regionX (orphan, always 0)]
        this.Jh.outBuffer.putShort(curY + this.zg); // obf: var5 + zg  [SPLIT-FIELD FIX: was regionY (orphan, always 0)]

        if (walkToAction && steps == -1 && (curX + this.Qg) % 5 == 0) {
            steps = 0;
        }
        for (int i = steps; i >= 0 && i > steps - 25; i--) {
            this.Jh.outBuffer.putByte(this.walkPathX[i] - curX); // dx
            this.Jh.outBuffer.putByte(this.walkPathY[i] - curY); // dy
        }
        this.Jh.finishPacket(21294);

        this.tj = this.mouseX;
        this.Fd = this.mouseY;
        this.xh = -24;
        return true;
    }

    /**
     * Question/menu dialog handler: renders the answer options and, on click, sends the
     * chosen answer index.
     * Sends opcode 116 (QUESTION_DIALOG_ANSWER): the selected option index (one byte).
     *
     * Hit-test (both render and click modes, matching oracle drawOptionMenu):
     *   mouseX < textWidth(option) && mouseY > 12*i && mouseY < 12*i + 12
     */
    // obf: private final void G(int)   [int param var1 is anti-tamper junk]
    private void sendDialogAnswer(int unused) {
        if (this.Cf == 0) { // obf: Cf == 0  -> render mode
            for (int i = 0; i < this.menuOptionCount; i++) { // obf: var6 < Id
                int colour = 0xFFFF; // yellow
                if (this.mouseX < this.li.textWidth(1, 125, this.menuOptions[i])
                        && this.mouseY > i * 12 && this.mouseY < i * 12 + 12) {
                    colour = 0xFF0000; // red highlight under cursor
                }
                this.li.drawstring(this.menuOptions[i], 6, i * 12 + 12, colour, false, 1);
            }
            return;
        }
        // Click mode: find the clicked option and send its index.
        for (int i = 0; i < this.menuOptionCount; i++) {
            if (this.mouseX < this.li.textWidth(1, 89, this.menuOptions[i])
                    && this.mouseY > i * 12 && this.mouseY < i * 12 + 12) {
                this.Jh.newPacket(116, 0); // QUESTION_DIALOG_ANSWER
                this.Jh.outBuffer.putByte(i);
                this.Jh.finishPacket(21294);
                break;
            }
        }
        this.Ph = false;    // obf: Ph = false
        this.Cf = 0;      // obf: Cf = 0
    }

    /**
     * Character-design screen controls: processes the Head / Hair / Top / Bottom / Skin and
     * gender arrow buttons (cycling the appearance indices, wrapping within each table) and,
     * when "Accept" is clicked, submits the chosen appearance and closes the screen.
     * Sends opcode 235 (PLAYER_APPEARANCE_CHANGE): gender, head, bodyGender, 2colour, hair,
     * top, bottom, skin (one byte each).
     *
     * obf class names for the appearance tables: n.m -> this.appearanceFlags (per-sprite
     * gender/slot flags), na.e -> this.appearanceCount (#sprites); the deob mirrors both obf
     * statics as Mudclient instance fields. The colour palettes
     * are Dg (hair), ei (top+bottom), Wh (skin) — see oracle GameData.character*Colours.
     */
    // obf: private final void F(int)   [int param var1 is anti-tamper junk]
    private void sendAppearance(int unused) {
        // BOOT HOOK (env-gated, headless): when RSC_AUTO_APPEARANCE is set, auto-submit the
        // default appearance once (sends opcode 235) so the server-opened "design your character"
        // screen (Kg) clears without a mouse click. Mirrors the committed RSC_AUTOLOGIN hook;
        // lets the headless bring-up advance past character creation into the live 3D world.
        if (this.Kg && System.getenv("RSC_AUTO_APPEARANCE") != null && !autoAppearanceSent) {
            autoAppearanceSent = true;
            this.Jh.newPacket(235, 0); // PLAYER_APPEARANCE_CHANGE
            this.Jh.outBuffer.putByte(this.appearanceGender);
            this.Jh.outBuffer.putByte(this.appearanceHead);
            this.Jh.outBuffer.putByte(this.appearanceBodyGender);
            this.Jh.outBuffer.putByte(this.appearance2Colour);
            this.Jh.outBuffer.putByte(this.appearanceHairColour);
            this.Jh.outBuffer.putByte(this.appearanceTopColour);
            this.Jh.outBuffer.putByte(this.appearanceBottomColour);
            this.Jh.outBuffer.putByte(this.appearanceSkinColour);
            this.Jh.finishPacket(21294);
            this.li.blackScreen(true);
            this.Kg = false;
            System.out.println("[RSC_AUTO_APPEARANCE] submitted default appearance; character-design screen cleared");
            return;
        }
        // panelCharDesign.handleMouse(lastMouseButtonDown, mouseY, junk, mouseButtonDown, mouseX).
        this.Af.handleMouseInput(this.lastMouseButtonDown, this.mouseY, -9989, this.mouseButtonDown, this.mouseX); // obf: Af.b(mouseButtonDown, mouseY, -9989, lastMouseButtonDown, mouseX)

        // Head arrows: cycle appearanceHead to the next sprite valid for the current gender
        // (flag&3 == 1 means "head" slot; flag & 4*gender must be set).
        if (this.Af.wasActivated((byte) -120, this.charDesignHeadLeft)) {   // obf: Af.a(.., Dj)
            do {
                this.appearanceHead = (this.appearanceCount + this.appearanceHead - 1) % this.appearanceCount;
            } while ((this.appearanceFlags[this.appearanceHead] & 3) != 1
                     || (this.appearanceFlags[this.appearanceHead] & (4 * this.appearanceGender)) == 0);
        }
        if (this.Af.wasActivated((byte) -118, this.charDesignHeadRight)) {  // obf: Af.a(.., pi)
            do {
                this.appearanceHead = (this.appearanceHead + 1) % this.appearanceCount;
            } while ((this.appearanceFlags[this.appearanceHead] & 3) != 1
                     || (this.appearanceFlags[this.appearanceHead] & (4 * this.appearanceGender)) == 0);
        }
        // Hair colour arrows.
        if (this.Af.wasActivated((byte) -111, this.charDesignHairLeft)) {   // obf: Af.a(.., Kj)
            this.appearanceHairColour = (this.charHairColours.length + this.appearanceHairColour - 1) % this.charHairColours.length;
        }
        if (this.Af.wasActivated((byte) -109, this.charDesignHairRight)) {  // obf: Af.a(.., ed)
            this.appearanceHairColour = (this.appearanceHairColour + 1) % this.charHairColours.length;
        }
        // Gender arrows: flip gender, then re-seek head (flag&3==1) and bodyGender (flag&3==2).
        if (this.Af.wasActivated((byte) -118, this.charDesignGenderLeft)   // obf: Af.a(.., Ge)
                || this.Af.wasActivated((byte) -117, this.charDesignGenderRight)) { // obf: Af.a(.., Of)
            this.appearanceGender = 3 - this.appearanceGender;
            while ((this.appearanceFlags[this.appearanceHead] & 3) != 1
                    || (this.appearanceFlags[this.appearanceHead] & (4 * this.appearanceGender)) == 0) {
                this.appearanceHead = (this.appearanceHead + 1) % this.appearanceCount;
            }
            while ((this.appearanceFlags[this.appearanceBodyGender] & 3) != 2
                    || (this.appearanceFlags[this.appearanceBodyGender] & (4 * this.appearanceGender)) == 0) {
                this.appearanceBodyGender = (this.appearanceBodyGender + 1) % this.appearanceCount;
            }
        }
        // Top colour arrows.
        if (this.Af.wasActivated((byte) -123, this.charDesignTopLeft)) {    // obf: Af.a(.., Xc)
            this.appearanceTopColour = (this.appearanceTopColour - 1 + this.charTopBottomColours.length) % this.charTopBottomColours.length;
        }
        if (this.Af.wasActivated((byte) -102, this.charDesignTopRight)) {   // obf: Af.a(.., ek)
            this.appearanceTopColour = (this.appearanceTopColour + 1) % this.charTopBottomColours.length;
        }
        // Skin colour arrows.
        if (this.Af.wasActivated((byte) -127, this.charDesignSkinLeft)) {   // obf: Af.a(.., Ze)
            this.appearanceSkinColour = (this.charSkinColours.length + this.appearanceSkinColour - 1) % this.charSkinColours.length;
        }
        if (this.Af.wasActivated((byte) -102, this.charDesignSkinRight)) {  // obf: Af.a(.., Mj)
            this.appearanceSkinColour = (this.appearanceSkinColour + 1) % this.charSkinColours.length;
        }
        // Bottom colour arrows.
        if (this.Af.wasActivated((byte) -101, this.charDesignBottomLeft)) { // obf: Af.a(.., Re)
            this.appearanceBottomColour = (this.charTopBottomColours.length + this.appearanceBottomColour - 1) % this.charTopBottomColours.length;
        }
        if (this.Af.wasActivated((byte) -122, this.charDesignBottomRight)) {// obf: Af.a(.., Ai)
            this.appearanceBottomColour = (this.appearanceBottomColour + 1) % this.charTopBottomColours.length;
        }

        // "Accept" button: submit the new appearance.
        if (!this.Af.wasActivated((byte) -118, this.charDesignAccept)) {    // obf: Af.a(.., Eg)
            return;
        }
        this.Jh.newPacket(235, 0); // PLAYER_APPEARANCE_CHANGE
        this.Jh.outBuffer.putByte(this.appearanceGender);      // obf: Sf
        this.Jh.outBuffer.putByte(this.appearanceHead);        // obf: Vd
        this.Jh.outBuffer.putByte(this.appearanceBodyGender);  // obf: dk
        this.Jh.outBuffer.putByte(this.appearance2Colour);     // obf: wg
        this.Jh.outBuffer.putByte(this.appearanceHairColour);  // obf: ld
        this.Jh.outBuffer.putByte(this.appearanceTopColour);   // obf: Wg
        this.Jh.outBuffer.putByte(this.appearanceBottomColour);// obf: Lh
        this.Jh.outBuffer.putByte(this.appearanceSkinColour);  // obf: hh
        this.Jh.finishPacket(21294);
        this.li.blackScreen(true);          // obf: li.a(true)
        this.Kg = false;   // obf: Kg = false
    }

    /**
     * Combat-style tab: renders the five rows (header + Controlled / Aggressive / Accurate /
     * Defensive) and, when one of the four style rows is clicked, selects it and informs the
     * server.
     * Sends opcode 29 (COMBAT_STYLE_CHANGED): the selected style index (one byte, 0..3).
     *
     * Hit-test (matching oracle drawDialogCombatStyle): row index 1..4 only,
     *   mouseX > boxX && mouseX < boxX+boxW && mouseY > boxY+20*row && mouseY < boxY+20*row+20
     */
    // obf: private final void k(byte)   [byte param var1 is anti-tamper junk]
    private void sendCombatStyle(byte unused) {
        int boxX = 7;    // obf: var2
        int boxY = 15;   // obf: var3
        int boxW = 175;  // obf: var4

        // Click mode: hit-test the four style rows (row 0 is the header).
        if (this.Cf != 0) { // obf: Cf != 0
            for (int row = 0; row < 5; row++) {
                if (row > 0
                        && this.mouseX > boxX && this.mouseX < boxX + boxW
                        && this.mouseY > boxY + row * 20 && this.mouseY < boxY + row * 20 + 20) {
                    this.Fg = row - 1; // obf: Fg = var5 - 1
                    this.Cf = 0;  // obf: Cf = 0
                    this.Jh.newPacket(29, 0); // COMBAT_STYLE_CHANGED
                    this.Jh.outBuffer.putByte(this.Fg);
                    this.Jh.finishPacket(21294);
                    break;
                }
            }
        }
        // Render the five rows (selected style row highlighted red) + labels.
        for (int row = 0; row < 5; row++) {
            int fill = (row == this.Fg + 1) ? ClientStream.rgb(255, 0, 0) : ClientStream.rgb(190, 190, 190);
            this.li.drawBoxAlpha(128, boxX, 20, 0, boxY + row * 20, boxW, fill);
            this.li.drawLineHoriz(boxW, 0, boxX, boxY + row * 20, (byte) 82);
            this.li.drawLineHoriz(boxW, 0, boxX, boxY + row * 20 + 20, (byte) -127);
        }
        // ARG-ORDER FIX (class c): clean order is (x, text, colour, inlineSprite, font, y)
        // (clean client.java:7833-7837). These were written against the old swapped signature.
        this.li.drawStringCenter(boxX + boxW / 2, STRINGS[650], 0xFFFFFF, 0, 3, boxY + 16); // header "Select combat style"
        this.li.drawStringCenter(boxX + boxW / 2, STRINGS[648], 0, 0, 3, boxY + 36);        // Controlled
        this.li.drawStringCenter(boxX + boxW / 2, STRINGS[645], 0, 0, 3, boxY + 56);        // Aggressive
        this.li.drawStringCenter(boxX + boxW / 2, STRINGS[649], 0, 0, 3, boxY + 76);        // Accurate
        this.li.drawStringCenter(boxX + boxW / 2, STRINGS[647], 0, 0, 3, boxY + 96);        // Defensive
    }

    // obf: private final void a(int,int,byte)   [byte param var3 is anti-tamper guard: send only if var3 == -78]
    // Moved to TradeDuelBankPackets.sendDuelOffer
    void sendDuelOffer(int slot, int qty, byte unused) {
        tradePackets.sendDuelOffer(slot, qty, unused);
    }

    // obf: private final void c(int,byte,int)   [byte param var2 is anti-tamper guard: send only if var2 > 120]
    // Moved to TradeDuelBankPackets.sendTradeOffer
    void sendTradeOffer(int qty, byte unused, int slot) {
        tradePackets.sendTradeOffer(qty, unused, slot);
    }

    /**
     * Camera-angle obstruction test for the auto-rotate logic: returns false if a wall blocks
     * the requested discrete 8-way camera angle {@code angle} from the local player's tile, true
     * otherwise. Checks the {@link client.Ek.World#objectAdjacency objectAdjacency} wall-flag
     * bit 0x80 on the tiles two and one steps away in the angle's direction.
     *
     * <p>Reconstructed from the clean decompile {@code client.b(byte,int)} (the obfuscator's
     * {@code -17/((-50-var1)/62)} dummy and the {@code try/catch(RuntimeException)} profiling
     * wrapper are dropped per the deob convention; {@code Hh.bb} = Hh.objectAdjacency since the
     * deob 'scene' alias holds the World).
     */
    // obf: private final boolean b(byte,int)   [byte param var1 is anti-tamper junk]
    private final boolean isValidCameraAngle(byte unused, int angle) {
        int tileX = this.wi.currentX / 128; // wi.i / 128
        int tileY = this.wi.currentY / 128; // wi.K / 128
        for (int s = 2; s >= 1; s--) {
            if (angle == 1
                    && ((this.Hh.objectAdjacency[tileX][tileY - s] & 128) == 128
                        || (this.Hh.objectAdjacency[tileX - s][tileY] & 128) == 128
                        || (this.Hh.objectAdjacency[tileX - s][tileY - s] & 128) == 128)) {
                return false;
            }
            if (angle == 3
                    && ((this.Hh.objectAdjacency[tileX][tileY + s] & 128) == 128
                        || (this.Hh.objectAdjacency[tileX - s][tileY] & 128) == 128
                        || (this.Hh.objectAdjacency[tileX - s][tileY + s] & 128) == 128)) {
                return false;
            }
            if (angle == 5
                    && ((this.Hh.objectAdjacency[tileX][tileY + s] & 128) == 128
                        || (this.Hh.objectAdjacency[tileX + s][tileY] & 128) == 128
                        || (this.Hh.objectAdjacency[tileX + s][tileY + s] & 128) == 128)) {
                return false;
            }
            if (angle == 7
                    && ((this.Hh.objectAdjacency[tileX][tileY - s] & 128) == 128
                        || (this.Hh.objectAdjacency[tileX + s][tileY] & 128) == 128
                        || (this.Hh.objectAdjacency[tileX + s][tileY - s] & 128) == 128)) {
                return false;
            }
            if (angle == 0 && (this.Hh.objectAdjacency[tileX][tileY - s] & 128) == 128) {
                return false;
            }
            if (angle == 2 && (this.Hh.objectAdjacency[tileX - s][tileY] & 128) == 128) {
                return false;
            }
            if (angle == 4 && (this.Hh.objectAdjacency[tileX][tileY + s] & 128) == 128) {
                return false;
            }
            if (angle == 6 && (this.Hh.objectAdjacency[tileX + s][tileY] & 128) == 128) {
                return false;
            }
        }
        return true;
    }

    /**
     * Final logout/close acknowledgement: when {@code send} is set, fires the close packet
     * and tears the connection down; then clears the cached credentials and resets the
     * username/password entry state.
     * Sends opcode 31 (CONFIRM_LOGOUT / CLOSE_CONNECTION).
     */
    // obf: private final void a(boolean,int)   [int param var2 is anti-tamper junk: if(var2!=31){sf=null}]
    void sendConfirmLogoutAck(boolean send, int unused) {
        if (send && this.Jh != null) { // obf: var1 && Jh != null
            // (clean wrapped these in try/catch(IOException); the deob newPacket/closeStream
            //  no longer declare throws IOException, so the dead catch is dropped.)
            this.Jh.newPacket(31, 0);     // CONFIRM_LOGOUT
            this.Jh.closeStream(true); // obf: Jh.a(-6924)  flush + close socket/writer thread
        }
        this.password = "";  // obf: wh = ""
        this.username = "";  // obf: Xf = ""
        this.resetTradeDuelState(-2); // obf: o(var2 ^ -31) -> o(-2): clears entry-cursor state (incl. kc)
    }




    // =========================================================================
    // ===== scene =====
    // =========================================================================
// Group: Mudclient 3D scene/world build, entity menu-build & sprite render, camera, models.
// All methods are instance methods of class Mudclient (package client, extends GameShell),
// 4-space indented as if inside the class body.
//
// RE-AUDITED against the CLEAN Vineflower base (decompiled/normalized-clean/client.java).
// The previous version of this part file was written against a DEFECTIVE base (drawWorld /
// addSceneObject / buildTerrainTile / loadRegion were missing and reconstructed by hand);
// those reconstructions contained numerous logic errors that are corrected here.
//
// Stripping applied to every method:
//   - opaque predicate:  boolean bl = client.OPAQUE_FALSE;  (always false, dead) — removed
//   - profiling counters: ++<StaticCounter>;                 (dead) — removed
//   - exception wrapper:  catch(RuntimeException e){ throw ErrorHandler.wrap(e,"sig"); } — unwrapped
//   - anti-tamper guards: if(param != <magic>) <side-effect>; — kept verbatim where it touches
//                         a field (the JIT'd code really does execute the store), noted // guard
//   - junk before shifts: XOR masks on shift amounts        — removed
//   - ~x>~y / ~x==const sign idioms                          — rewritten to plain comparisons
//
// CROSS-CLASS NAMING — CORRECTED per docs/NAMING.md ("k=World, lb=Scene"):
//   The field types are:  private k Hh;   private lb Ek;   private ba li;
//   => Hh is type k  == World   (terrain/elevation/region/route)  -> read as `world`
//   => Ek is type lb == Scene   (3D renderer: model arrays, picking, camera) -> read as `scene`
//   The OLD version had these BACKWARDS (Hh=scene, Ek=world); fixed throughout this file.
//
//   Other key fields (MUDCLIENT_SKELETON.md):
//     li=surface(SurfaceSprite/ba), zh=menuList(wb; the right-click option accumulator —
//        NOT the friends list), Tb=playersInView, de=playerInViewCount,
//        rg=npcsInView, Yc=npcInViewCount, te=playersCache, We=npcsCache,
//        Ff/qj=knownPlayers/count, Zg/If=knownNpcs/count, Ug=tileSize,
//        af=selectedSpellOrItem, Bh=localPlayerServerIndex, ig=local player name.

    // -------------------------------------------------------------------------
    // drawWorld  (obf: void s(int)  @clean L10318)
    // -------------------------------------------------------------------------

    /**
     * Build the in-Ek right-click option menu for everything the 3D Scene
     * picked this frame: panel-element clicks first, then scenery objects,
     * boundary walls, ground items, NPCs and players, then a final
     * held-item / ground-tile fallthrough.
     *
     * Despite the skeleton label "drawWorld", this does NOT rasterise the Hh
     * (that happens elsewhere); it walks the Scene's picked-entity lists
     * (Ek.b(124) = GameModel[], Ek.a(104) = encoded ids) and appends menu
     * options via the menu accumulator zh.  Each encoded id is
     *   E[idx] / 10000 = kind  (1=scenery, 2=ground item, 3=player), %10000 = local index;
     *   GameModels with rb >= 10000 are boundary walls (rb-10000 = wall index).
     *
     * obf: void s(int param)   (param: an interaction-pass selector; 2 = passive redraw)
     */
    private final void drawWorld(int param) {
        // --- panel-element click on the active sub-screen (Zh = screen/panel mode) ---
        if (Zh == 1 && panelShop.wasActivated((byte)-107, Fh) || Zh == 3 && panelShop.wasActivated((byte)-116, mc)) { // qa.a(byte,int)
            int el = (Zh == 1) ? Fh : mc;                  // selected panel element
            int packed = panelShop.getSelectedItem(14458, el);   // yd.f(int,int)
            if ((packed >> 16) == 2 || (Yh && (packed >> 16) == 1)) {
                int idx = packed & 0xFFFF;
                String actionA = panelShop.getItemFeedback(idx, 19680, el);   // yd.b(int,int,int)
                String actionB = panelShop.getItemLabel(idx, param ^ -122, el); // yd.a(int,int,int)
                if (this.a(actionA, param ^ 125, actionB)) {   // dispatch panel action (Mudclient a(String,int,String) — declared in another segment)
                    return;
                }
            }
        }

        // --- if on the login/world screen (Zh==0), hit-test the world list ---
        if (Zh == 0) {
            for (int w = 0; w < 100; w++) {
                if (ImageLoader.scratchBuf[w] >= 0
                        && (FontWidths.entryTypes[w] == 4 || FontWidths.entryTypes[w] == 1 || FontWidths.entryTypes[w] == 5 || FontWidths.entryTypes[w] == 6)) { // n.j
                    String label = NameTable.recentNames[w] + Utility.formatChatLine(BZip.entityNames[w], World.G[w], true, FontWidths.entryTypes[w]); // mb.a(String,String,bool,int)
                    if (mouseX > 7
                            && mouseX < li.textWidth(1, param ^ 114, label) + 7
                            && mouseY > screenHeight - 30 - w * 12
                            && mouseY < screenHeight - 18 - 12 * w
                            && (Cf == 2 || (Yh && Cf == 1))
                            && this.a(SurfaceSprite.recentMessages[w], 127, World.G[w])) { // ba.Yb
                        return;
                    }
                }
            }
        } else {
            Hc = false;
        }

        // --- reset per-frame "already added to menu this tick" flags ---
        for (int i = 0; i < eh; i++) {
            Ed[i] = false;                                  // scenery objects
        }
        int clickedGroundSlot = -1;
        for (int i = 0; i < hf; i++) {
            Sj[i] = false;                                  // boundary walls
        }

        // --- iterate the Scene's picked entity list (Ek=Scene -> `world` alias is the Scene-typed field) ---
        int entityCount = Ek.getMousePickedCount(0);        // Scene.b(int) picked count
        GameModel[] models = Ek.getMousePickedModels();     // Scene.b(byte) picked models (GameModel[])
        int[] faceTags = Ek.getMousePickedFaces();          // Scene.a(byte) picked face tags (int[])
        if (param != 2) {
            nk = -82;                                       // reset hover unless passive redraw
        }

        for (int pick = 0; pick < entityCount; pick++) {
            if (zh.getCount(param ^ -27155) > 200) {        // depth-cull: skip far-from-mouse polys (zh.c(int))
                continue;
            }
            int tag = faceTags[pick];
            GameModel model = models[pick];

            // valid face-tag bands: E[tag] in [0..0xFFFF], or [200000..300000]
            int eid = model.faceTag[tag];
            if (!(eid <= 0xFFFF || (eid >= 200000 && eid <= 300000))) {
                continue;
            }

            // When the picked model is NOT the Scene's special target (Ek.T), it is a
            // scenery object (rb in [0..10000)) or a boundary wall (rb >= 10000); build
            // those menus and skip to the next entity.  When it IS the target, fall
            // through to the per-kind dispatch (ground item / player).
            if (Ek.view != model) {                          // Scene.T (view); Ek=Scene -> `world` alias
                if (model == null || model.key < 10000) {
                    // --- scenery object branch (also handles the ground-tile remap) ---
                    if (model != null && model.key >= 0) {
                        int objSlot = model.key;
                        int objId   = vc[objSlot];
                        if (!Ed[objSlot]) {
                            if (af < 0) {
                                if (Bh >= 0) {              // walk-to scenery
                                    this.zh.addEntryRich(ye[objSlot], STRINGS[38] + ig + STRINGS[53], -104, Bh,
                                         vc[objSlot], 410, bg[objSlot],
                                         STRINGS[41] + Globals.paramNames[objId], Se[objSlot]);
                                }
                                if (!FontBuilder.injectedStrings2[objId].equalsIgnoreCase(STRINGS[33])) {  // command 1
                                    this.zh.addEntryGuarded(420, ye[objSlot], bg[objSlot], Se[objSlot], param ^ 107,
                                         vc[objSlot], STRINGS[41] + Globals.paramNames[objId], FontBuilder.injectedStrings2[objId]);
                                }
                                if (!Timer.legacyStringsA[objId].equalsIgnoreCase(STRINGS[51])) {  // command 2
                                    this.zh.addEntryGuarded(2400, ye[objSlot], bg[objSlot], Se[objSlot], param ^ 127,
                                         vc[objSlot], STRINGS[41] + Globals.paramNames[objId], Timer.legacyStringsA[objId]);
                                }
                                this.zh.addEntryScrolled(objId, 3400, false, STRINGS[51], STRINGS[41] + Globals.paramNames[objId]); // Examine
                            } else if (GameFrame.unusedIntBuffer[af] == 5) {     // cast held spell on scenery
                                this.zh.addEntryRich(ye[objSlot], STRINGS[46] + BitBuffer.UNUSED_L[af] + STRINGS[50], param + 65,
                                     af, vc[objSlot], 400, bg[objSlot],
                                     STRINGS[41] + Globals.paramNames[objId], Se[objSlot]);
                            }
                            Ed[objSlot] = true;
                        }
                        continue;                            // scenery handled
                    }
                    // model present but rb < 0: this is the ground-tile face — remember it
                    if (tag >= 0) {
                        tag = model.faceTag[tag] - 200000;
                    }
                    if (tag < 0) {
                        continue;
                    }
                    clickedGroundSlot = tag;
                    continue;
                }

                // --- boundary wall branch (rb >= 10000) ---
                int wallIdx = model.key - 10000;
                int wallId  = Ng[wallIdx];
                if (!Sj[wallIdx]) {
                    if (af >= 0 && GameFrame.unusedIntBuffer[af] == 5) {          // cast held spell on wall
                        this.zh.addEntryGuarded(300, yk[wallIdx], Hj[wallIdx], Jd[wallIdx], 60, af,
                             STRINGS[41] + GameCharacter.sharedNameTable[wallId], STRINGS[46] + BitBuffer.UNUSED_L[af] + STRINGS[50]);
                    }
                    if (Bh >= 0) {                           // walk-to wall
                        this.zh.addEntryGuarded(310, yk[wallIdx], Hj[wallIdx], Jd[wallIdx], param ^ 66, Bh,
                             STRINGS[41] + GameCharacter.sharedNameTable[wallId], STRINGS[38] + ig + STRINGS[53]);
                    }
                    if (!StringCodec.DEAD_STRING_ARRAY[wallId].equalsIgnoreCase(STRINGS[33])) {   // command 1
                        this.zh.addEntryWithFont(Jd[wallIdx], (byte)22, 320, StringCodec.DEAD_STRING_ARRAY[wallId],
                             STRINGS[41] + GameCharacter.sharedNameTable[wallId], Hj[wallIdx], yk[wallIdx]);
                    }
                    if (!RecordLoader.stringTable[wallId].equalsIgnoreCase(STRINGS[51])) {   // command 2
                        this.zh.addEntryWithFont(Jd[wallIdx], (byte)22, 2300, RecordLoader.stringTable[wallId],
                             STRINGS[41] + GameCharacter.sharedNameTable[wallId], Hj[wallIdx], yk[wallIdx]);
                    }
                    this.zh.addEntryScrolled(wallId, 3300, false, STRINGS[51], STRINGS[41] + GameCharacter.sharedNameTable[wallId]); // Examine
                    Sj[wallIdx] = true;
                }
                continue;                                    // wall handled
            }

            // --- Scene-target model: decode kind/local index and build mob menus ---
            int local = model.faceTag[tag] % 10000;
            int kind  = model.faceTag[tag] / 10000;
            if (kind != 1) {
                if (kind == 2) {
                    // ground item
                    if (af >= 0) {
                        if (GameFrame.unusedIntBuffer[af] != 3) {                 // not a use-on-item spell/item
                            continue;
                        }
                        this.zh.addEntryGuarded(200, Ni[local], Gj[local], Zf[local], param ^ 70, af,
                             STRINGS[34] + DecodeBuffer.chatFilterCache[Gj[local]], STRINGS[46] + BitBuffer.UNUSED_L[af] + STRINGS[50]);
                        continue;
                    }
                    if (Bh < 0) {                            // clean: ~Bh > -1  (Bh < 0)
                        this.zh.addEntryWithFont(Zf[local], (byte)22, 220, STRINGS[52], STRINGS[34] + DecodeBuffer.chatFilterCache[Gj[local]],
                             Gj[local], Ni[local]);          // Pick up
                        this.zh.addEntryScrolled(Gj[local], 3200, false, STRINGS[51], STRINGS[34] + DecodeBuffer.chatFilterCache[Gj[local]]); // Examine
                        continue;
                    }
                    this.zh.addEntryGuarded(210, Ni[local], Gj[local], Zf[local], 68, Bh,
                         STRINGS[34] + DecodeBuffer.chatFilterCache[Gj[local]], STRINGS[38] + ig + STRINGS[53]);
                    continue;
                }
                if (kind != 3) {
                    continue;
                }

                // player
                String vsText = "";
                int combatDelta = -1;
                int charType = Tb[local].serverId;
                if (ISAAC.unusedA[charType] > 0) {                     // combat-capable (PvP target)
                    int theirLevel = (AudioMixer.scratchBuffer[charType] + ClientRuntimeException.intScratch[charType]
                                      + DownloadWorker.unusedIntArray[charType] + SurfaceImageProducer.entityIndexTableD[charType]) / 4;
                    int myLevel = (cg[3] + cg[2] + cg[1] + cg[0] + 27) / 4;
                    vsText = STRINGS[20];                    // baseline " @whi@" colour
                    combatDelta = myLevel - theirLevel;
                    if (combatDelta < 0)  vsText = STRINGS[40];   // they higher
                    if (combatDelta < -3) vsText = STRINGS[39];
                    if (combatDelta < -6) vsText = STRINGS[49];
                    if (combatDelta < -9) vsText = STRINGS[10];
                    if (combatDelta > 0)  vsText = STRINGS[35];   // you higher
                    if (combatDelta > 3)  vsText = STRINGS[37];
                    if (combatDelta > 6)  vsText = STRINGS[47];
                    if (combatDelta > 9)  vsText = STRINGS[27];
                    vsText = " " + vsText + STRINGS[42] + theirLevel + ")";
                }
                if (af >= 0) {
                    if (GameFrame.unusedIntBuffer[af] != 2) {                     // not a cast-on-player spell
                        continue;
                    }
                    this.zh.addEntryWithColor(Tb[local].serverIndex, STRINGS[20] + GameShell.equipMb[Tb[local].serverId], 700,
                         STRINGS[46] + BitBuffer.UNUSED_L[af] + STRINGS[50], af, 3296);
                    continue;
                }
                if (Bh < 0) {                                // clean: -1 < ~Bh  (Bh < 0)
                    if (ISAAC.unusedA[charType] > 0) {                 // Attack
                        this.zh.addEntryScrolled(Tb[local].serverIndex, combatDelta >= 0 ? 715 : 2715, false, STRINGS[48],
                             STRINGS[20] + GameShell.equipMb[Tb[local].serverId] + vsText);
                    }
                    this.zh.addEntryScrolled(Tb[local].serverIndex, 720, false, STRINGS[45], STRINGS[20] + GameShell.equipMb[Tb[local].serverId]); // Trade
                    if (!Timer.legacyStringsE[charType].equals("")) {
                        this.zh.addEntryScrolled(Tb[local].serverIndex, 725, false, Timer.legacyStringsE[charType], STRINGS[20] + GameShell.equipMb[Tb[local].serverId]);
                    }
                    this.zh.addEntryScrolled(Tb[local].serverId, 3700, false, STRINGS[51], STRINGS[20] + GameShell.equipMb[Tb[local].serverId]); // Examine
                }
                this.zh.addEntryWithColor(Tb[local].serverIndex, STRINGS[20] + GameShell.equipMb[Tb[local].serverId], 710,
                     STRINGS[38] + ig + STRINGS[53], Bh, param ^ 3298);  // Follow
            }

            this.menus.buildClickMenu(local, -12);                              // walk-to entity tile
        }

        // --- global held-item fallback: "Use <item> with" when nothing else matched ---
        if (af >= 0 && GameFrame.unusedIntBuffer[af] <= 1) {
            this.zh.addEntryScrolled(af, 1000, false, STRINGS[46] + BitBuffer.UNUSED_L[af] + STRINGS[43], "");
        }

        // --- clicked ground tile (no entity): build a walk-here / use-on-ground option ---
        if (clickedGroundSlot != -1) {
            Hc = true;
            int slot = clickedGroundSlot;
            rf = Qg + Hh.localX[slot];                            // world X = regionBaseX + scene tile X
            Cg = zg + Hh.localY[slot];                            // world Z = regionBaseZ + scene tile Z
            if (af >= 0) {
                if (GameFrame.unusedIntBuffer[af] != 6) {                         // not a "use-on-ground" spell
                    return;
                }
                this.zh.addEntryWithFont(Hh.localX[slot], (byte)22, 900, STRINGS[46] + BitBuffer.UNUSED_L[af] + STRINGS[44], "",
                     af, Hh.localY[slot]);
                return;
            }
            if (Bh < 0) {                                    // clean: -1 < ~Bh  (Bh < 0)
                this.zh.addEntryWithColor(Hh.localX[slot], "", 920, STRINGS[54], Hh.localY[slot], 3296);   // Walk here
            }
        }
    }

    // -------------------------------------------------------------------------
    // loadRegion  (obf: boolean a(int,int,boolean)  @clean L11537 — bytecode only)
    // -------------------------------------------------------------------------

    /**
     * (Re)load the 48x48 terrain region centred on the given Ek tile, snapping
     * the region origin to the chunk grid.  On a cache hit (player still inside the
     * loaded chunk and floor unchanged) returns false without reloading; otherwise
     * shows "Loading... Please wait", re-bases every wall/object/ground-item model
     * and every in-view player/npc by the region-origin delta, and returns true.
     *
     * Decoded directly from the method's bytecode (Vineflower could not decompile
     * this one method — the single residual failure in the clean base).
     *
     * obf: boolean a(int x, int z, boolean isUnderground)
     */
    final boolean loadRegion(int x, int z, boolean isUnderground) {
        // disconnected / fatal stream error -> mark World not-ready, bail
        if (rk != 0) {
            Hh.playerAlive = false;                                    // World.loaded = false (scene alias = World)
            return false;
        }
        this.Ub = isUnderground;

        // shift requested tile by the player's sub-region offset
        x += sk;
        z += Ki;

        // cache hit: same floor and still strictly inside the loaded window
        if (yj == bc && Jg < z && Rk > z && Fi < x && x < Ne) {
            Hh.playerAlive = true;                                     // World.loaded = true (scene alias = World)
            return false;
        }

        // --- full region reload ---
        li.drawstringRight(256, STRINGS[676], 0xFFFFFF, 0, 1, 192);   // li.a(int,String,int,int,int,int): "Loading... Please wait"
        this.drawChatHistoryTabs(5);
        li.draw(graphics, originX, 256, originY);                     // li.a(Graphics,int,int,int): flush to AWT

        int oldRegionX = Qg;
        int oldRegionZ = zg;

        // snap to 48-tile chunk grid; NOTE: Qg derives from z, zg derives from x
        int chunkZ = (z + 24) / 48;
        int chunkX = (x + 24) / 48;
        Qg = chunkZ * 48 - 48;
        Ne = chunkX * 48 + 32;
        Rk = chunkZ * 48 + 32;
        yj = bc;                                             // latch floor level
        Fi = chunkX * 48 - 32;
        zg = chunkX * 48 - 48;
        Jg = chunkZ * 48 - 32;

        // tell World its new origin/floor
        Hh.loadSection(z, (byte)-90, x, yj);                 // World.loadSection (obf k.a(int,byte=-90,int,int)); scene alias = World

        // subtract player sub-region offset so coords stay view-relative
        zg -= sk;
        Qg -= Ki;

        int deltaX = Qg - oldRegionX;                        // = -oldRegionX + Qg
        int deltaZ = zg - oldRegionZ;                        // = -oldRegionZ + zg

        // --- rebase wall/object models (slots [0..eh)) ---
        for (int i = 0; i < eh; i++) {
            Se[i] -= deltaX;
            ye[i] -= deltaZ;
            int tileX    = Se[i];
            int tileZ    = ye[i];
            int objType  = vc[i];
            GameModel model     = hg[i];
            int dir      = bg[i];

            try {
                int modelW, modelH;
                if (dir == 0 || dir == 4) {
                    modelW = RecordLoader.intArray[objType];                   // obf: f.f  (RecordLoader.f = width table)
                    modelH = NameTable.sortKeys[objType];                  // obf: ub.g (NameTable.g  = height table)
                } else {
                    modelW = NameTable.sortKeys[objType];
                    modelH = RecordLoader.intArray[objType];
                }
                int midX = (tileX + tileX + modelW) * Ug / 2;
                int midZ = Ug * (tileZ + tileZ + modelH) / 2;

                // cull if outside the visible 96x96 tile window
                if (tileX < 0 || tileZ < 0 || tileX >= 96 || tileZ >= 96) {
                    continue;
                }
                Ek.addModel(model);                           // Scene.addModel (obf lb.a(ca,byte=118)); world alias = Scene
                model.place(-Hh.getElevation(midX, midZ), -123, midZ, midX); // World.getElevation (obf k.f); scene alias = World
                Hh.removeObject2(tileX, objType, isUnderground, tileZ);   // World.removeObject2 (obf k.a(int,int,boolean,int)); scene alias = World
                if (objType == 74) {                          // special: floats 480 up
                    model.translate(0, 0, -480, true);
                }
            } catch (RuntimeException ex) {
                System.out.println(STRINGS[671] + ex.getMessage());
                System.out.println(STRINGS[672] + i + STRINGS[673] + model);
                ex.printStackTrace();
            }
        }

        // --- rebase door/diagonal entity models (slots [0..hf)) ---
        for (int i = 0; i < hf; i++) {
            Jd[i] -= deltaX;
            yk[i] -= deltaZ;
            int tileX   = Jd[i];
            int tileZ   = yk[i];
            int objType = Ng[i];
            int dir     = Hj[i];
            try {
                Hh.setWallObjectAdjacency(tileZ, objType, dir, tileX, 11715);   // World.setWallObjectAdjacency (obf k.a(int,int,int,int,int)); scene alias = World
                rd[i] = this.buildEntityModel(!isUnderground, tileZ, objType, tileX, dir, i);
            } catch (RuntimeException ex) {
                System.out.println(STRINGS[674] + ex.getMessage());
                ex.printStackTrace();
            }
        }

        // --- rebase ground-item spots (slots [0..Ah)) ---
        for (int i = 0; i < Ah; i++) {
            Zf[i] -= deltaX;
            Ni[i] -= deltaZ;
        }

        // --- rebase in-view NPCs (rg[0..Yc)) ---
        for (int i = 0; i < Yc; i++) {
            GameCharacter npc = rg[i];
            npc.currentX -= Ug * deltaX;
            npc.currentY -= deltaZ * Ug;
            for (int wp = 0; wp <= npc.waypointCurrent; wp++) {
                npc.waypointsX[wp] -= Ug * deltaX;
                npc.waypointsY[wp] -= deltaZ * Ug;
            }
        }

        // --- rebase in-view players (Tb[0..de)) ---
        for (int i = 0; i < de; i++) {
            GameCharacter pl = Tb[i];
            pl.currentY -= Ug * deltaZ;
            pl.currentX -= Ug * deltaX;
            for (int wp = 0; wp <= pl.waypointCurrent; wp++) {
                pl.waypointsX[wp] -= Ug * deltaX;
                pl.waypointsY[wp] -= deltaZ * Ug;
            }
        }

        Hh.playerAlive = true;                                         // World.loaded = true (scene alias = World)
        return true;
    }

    // -------------------------------------------------------------------------
    // buildEntityModel  (obf: ca a(boolean,int,int,int,int,int)  @clean L6204)
    // -------------------------------------------------------------------------

    /**
     * Build a single-quad GameModel for a boundary/door entity from World def data.
     * The quad spans from (tileX,tileZ) to a far corner chosen by the direction code
     * and is dropped to terrain elevation.  Tagged with rb = slot+10000 so the Scene
     * can recover its wall slot during picking.
     *
     * obf: ca a(boolean isUnderground, int tileX, int objType, int tileZ, int dir, int slot)
     */
    final GameModel buildEntityModel(boolean isUnderground, int tileX, int objType,
                                      int tileZ, int dir, int slot) {
        int nearX = tileZ;          // var7  (near corner, tileZ axis copy)
        int nearZ = tileX;          // var8  (near corner, tileX axis copy)
        int farX  = tileZ;          // var9
        int farZ  = tileX;          // var10
        int colour = ChatCipher.scratchA[objType];  // obf: v.a (ChatCipher.a = packed colour table)
        int bitmap = Jk[objType];   // obf: Jk   (texture/bitmap id table)
        int height = StreamBase._deadIntArray_d[objType]; // obf: ib.d (StreamBase.d = wall height)

        GameModel model = new GameModel(4, 1);
        if (dir == 1) {
            farZ = tileX + 1;
        }
        if (dir == 0) {
            farX = tileZ + 1;
        }
        if (dir == 2) {
            farZ = tileX + 1;
            nearX = tileZ + 1;
        }
        nearX *= Ug;
        if (dir == 3) {
            farZ = tileX + 1;
            farX = tileZ + 1;
        }
        nearZ *= Ug;
        farX  *= Ug;
        farZ  *= Ug;

        int v0 = model.vertexAt(nearX, nearZ, -Hh.getElevation(nearX, nearZ), -126);            // near bottom (World.getElevation, obf k.f; scene alias = World)
        int v1 = model.vertexAt(nearX, nearZ, -Hh.getElevation(nearX, nearZ) - height, -126);  // near top
        if (!isUnderground) {
            this.drawScrollbar2(119, 67, 26, 106, false, -100);           // internal flag setter (kept)
        }
        int v2 = model.vertexAt(farX, farZ, -height - Hh.getElevation(farX, farZ), -112);       // far top
        int v3 = model.vertexAt(farX, farZ, -Hh.getElevation(farX, farZ), 117);                  // far bottom

        model.createFace(4, new int[]{v0, v1, v2, v3}, colour, bitmap, false);
        model.setLight(-50, 60, -10, -50, false, 24, -95);          // lighting defaults

        if (tileZ >= 0 && tileX >= 0 && tileZ < 96 && tileX < 96) {
            Ek.addModel(model);                               // Scene.addModel (obf lb.a(ca,byte=118)); world alias = Scene
        }
        model.key = slot + 10000;                              // wall slot offset
        return model;
    }

    // -------------------------------------------------------------------------
    // addSceneObject  (obf: void a(int,int,int,int,int,int,int,int)  @clean L16874)
    // -------------------------------------------------------------------------

    /**
     * Render one PLAYER (Tb[id]) at screen position: blits each equipment/appearance
     * sprite layer, then queues its chat-message bubble, action bubble, health bar
     * and damage splat for later overlay passes.  (Oracle equivalent: drawPlayer.)
     *
     * obf: void a(int x, int y, int guard, int frameW, int scale, int id, int objW, int h)
     *   args: var1=x, var2=y, var3=guard, var4=frameW, var5=scale, var6=id, var7=objW, var8=h
     *   (objW renamed from obf var7=w to avoid clashing with the WorldEntity class `w`.)
     */
    final void addSceneObject(int x, int y, int guard, int frameW,
                              int scale, int id, int objW, int h) {
        if (guard != 20) {                                   // guard (kept; side-effect call)
            /* anti-tamper: 116 % ((69-guard)/35) — value discarded, no field touched */
        }
        GameCharacter player = Tb[id];

        // walk-cycle step: animationCurrent + (ug+16)/32, low 3 bits
        int walkAnim = (player.animationCurrent + (ug + 16) / 32) & 7;
        boolean flip = false;
        int step = walkAnim;
        if (step == 5) { flip = true; step = 3; }
        else if (step == 6) { flip = true; step = 2; }
        else if (step == 7) { flip = true; step = 1; }

        int frame = sf[player.movingStep / ArchiveReader.sectorAlloc[player.serverId] % 4] + step * 3;   // base walk frame
        if (player.animationCurrent == 8) {                                  // attacking
            flip = false;
            step = 5;
            h -= scale * LinkedQueue.sharedIntArray[player.serverId] / 100;
            walkAnim = 2;
            frame = 3 * step + Pc[jk / (StreamFactory.deadGuardArray[player.serverId] - 1) % 8];
        }
        if (player.animationCurrent == 9) {                                  // being hit
            step = 5;
            walkAnim = 2;
            flip = true;
            h += LinkedQueue.sharedIntArray[player.serverId] * scale / 100;
            frame = Og[jk / StreamFactory.deadGuardArray[player.serverId] % 8] + 3 * step;
        }

        for (int layer = 0; layer < 12; layer++) {
            int bodyPart = Tg[walkAnim][layer];               // appearance layer order
            int itemId   = GameFrame.unusedIntBuffer2d[player.serverId][bodyPart];          // obf: qb.d (appearance sprite id, -1 = none)
            if (itemId >= 0) {
                int dx = 0, dy = 0, f2 = frame;
                if (flip && step >= 1 && step <= 3 && BZip.entityFlags[itemId] == 1) {
                    f2 += 15;
                }
                if (step != 5 || DataStore.tamperScratch[itemId] == 1) {         // obf: nb.d (has a combat/idle frame)
                    int sprite = f2 + WorldEntity.spriteOffsets[itemId];            // obf: w.g (animationNumber)
                    int sw = li.spriteWidthFull[sprite];                   // sprite full width
                    int sh = li.spriteHeightFull[sprite];                   // sprite full height
                    int baseW = li.spriteWidthFull[WorldEntity.spriteOffsets[itemId]];
                    if (sw != 0 && sh != 0 && baseW != 0) {
                        dy = frameW * dy / sh;
                        dx = dx * objW / sw;
                        int drawW = objW * li.spriteWidthFull[sprite] / baseW;
                        dx -= (drawW - objW) / 2;
                        // appearance colour channel select (db.l = animationCharacterColour):
                        //   1 -> hair, 2 -> top, 3 -> bottom; default -> raw value (skin path)
                        int colourA = LinkedQueue.sharedIntArray2[itemId];
                        int colourB = 0;
                        if (colourA == 1) {                   // hair
                            colourB = ChatCipher.unusedE[player.serverId];      // obf: v.e
                            colourA = ClientStream.sharedIntArrayT[player.serverId];  // obf: da.T
                        } else if (colourA == 2) {            // top
                            colourA = SocketFactory.itemSpriteIndex[player.serverId]; // obf: m.g
                            colourB = ChatCipher.unusedE[player.serverId];      // obf: v.e
                        } else if (colourA == 3) {            // bottom
                            colourB = ChatCipher.unusedE[player.serverId];      // obf: v.e
                            colourA = Surface.unusedIntsAb[player.serverId];    // obf: ua.Ab
                        }
                        // else: colourA keeps db.l value, colourB stays 0 (clean fall-through)
                        // obf arg order: li.a(dy+x, colourA, colourB, flip, y, sprite, frameW, drawW, dx+h, 1)
                        li.spriteClipping(dy + x, colourA, colourB, flip, y, sprite, frameW, drawW, dx + h, 1); // li.a(int×10)
                    }
                }
            }
        }

        // queue chat message bubble
        if (player.messageTimeout > 0) {                                   // messageTimeout
            nf[Ef] = li.textWidth(1, 120, player.name) / 2;             // mid-point (clamped to 150)
            if (nf[Ef] > 150) nf[Ef] = 150;
            uf[Ef] = li.textWidth(1, 102, player.name) / 300 * li.textHeight(508305352, 1);  // line count * lineH
            tf[Ef] = objW / 2 + h;
            ee[Ef] = x;
            Kc[Ef++] = player.name;
        }

        // queue health bar + damage splat during/after combat
        if (player.animationCurrent == 8 || player.animationCurrent == 9 || player.combatTimer != 0) {
            if (player.combatTimer > 0) {                              // combatTimer
                int barX = h;
                if (player.animationCurrent == 9)      barX += scale * 20 / 100;
                else if (player.animationCurrent == 8) barX -= scale * 20 / 100;
                int barLen = player.healthCurrent * 30 / player.healthMax;       // healthCurrent/healthMax *30
                gd[Bc] = objW / 2 + barX;
                Pk[Bc] = x;
                bf[Bc++] = barLen;
            }
            if (player.combatTimer > 150) {
                int dmgX = h;
                if (player.animationCurrent == 9)      dmgX += scale * 10 / 100;
                else if (player.animationCurrent == 8) dmgX -= scale * 10 / 100;
                li.drawSprite(-1, tg + 12, x + frameW / 2 - 12, dmgX - (12 - objW / 2));   // li.b(int×4): splat sprite
                li.drawstringRight(objW / 2 + dmgX - 1, "" + player.damageTaken, 0xFFFFFF, 0, 3, 5 + x + frameW / 2); // li.a(int,String,int,int,int,int): dmg num
            }
        }
    }

    // -------------------------------------------------------------------------
    // addWallModel  (obf: void b(int,int,int,int,int,int,int)  @clean L13250  client.CA()
    // -------------------------------------------------------------------------

    /**
     * Blit a wall/boundary glyph sprite to the 2D li (rev ~235 draws some
     * boundary art as 2D sprites over the 3D pass).
     * obf: void b(int x, int a2, int z, int spriteType, int height, int guard, int screenY) — 7 params
     */
    final void addWallModel(int x, int a2, int z, int spriteType,
                            int height, int guard, int screenY) {
        if (guard > -109) {                                   // guard (kept; touches tj)
            tj = 50;
        }
        int spriteIndex = Surface.unusedIntsBb[spriteType] + sg;           // obf: ua.Bb (base sprite + global offset)
        int glyphWidth  = TextEncoder.scratchIntArray2[spriteType];                    // obf: h.c (TextEncoder width table)
        // obf arg order: li.a(screenY, glyphWidth, 0, false, 0, spriteIndex, x, height, z, 1)
        li.spriteClipping(screenY, glyphWidth, 0, false, 0, spriteIndex, x, height, z, 1); // li.a(int×10)
    }

    // -------------------------------------------------------------------------
    // addGroundObject  (obf: void a(int,int,int,int,int,int,int)  @clean L6064)
    // -------------------------------------------------------------------------

    /**
     * Draw a ground decoration marker (filled circle/oval) to the 2D li — used
     * for fire/torch floor glow.
     * obf: void a(int a1, int y, int z, int w, int decorType, int h, int guard)  — 7 params
     */
    final void addGroundObject(int a1, int y, int z, int w,
                               int decorType, int h, int guard) {
        if (guard != 2) {                                     // guard (kept; sets Dc)
            Dc = true;
        }
        int shape = Oc[decorType];                            // 0 = circle, 1 = oval
        int size  = oe[decorType];                            // marker scale
        if (shape == 0) {
            int colour = 255 + size * 1280;
            li.drawCircle(255 - 5 * size, -1057205208, 20 + size * 2, w / 2 + z, colour, y + h / 2); // li.c(int×6)
        }
        if (shape == 1) {
            int colour = 0xFF0000 + 1280 * size;
            li.drawCircle(255 - 5 * size, -1057205208, size + 10, z + w / 2, colour, y + h / 2); // li.c(int×6)
        }
    }

    // -------------------------------------------------------------------------
    // Public entity-sprite entry points, called cross-package by SurfaceSprite
    // (ba) from spriteClipping. These map the readable oracle names to the
    // (renamed) implementation methods, preserving the obf method letter:
    //   obf a(int×8)=addSceneObject -> drawNpc;  obf b(int×8)=buildTerrainTile -> drawPlayer;
    //   obf b(int×7)=addWallModel  -> drawItem;  obf a(int×7)=addGroundObject -> drawTeleportBubble.
    // Arg order is 1:1 with clean ba.java's dc.a/dc.b calls.
    // -------------------------------------------------------------------------
    public final void drawNpc(int x, int y, int guard, int frameW,
                              int scale, int id, int objW, int h) {
        addSceneObject(x, y, guard, frameW, scale, id, objW, h);
    }

    public final void drawPlayer(int objW, int frameW, int guard, int scale,
                                 int h, int x, int screenW, int id) {
        buildTerrainTile(objW, frameW, guard, scale, h, x, screenW, id);
    }

    public final void drawItem(int x, int a2, int z, int spriteType,
                               int height, int guard, int screenY) {
        addWallModel(x, a2, z, spriteType, height, guard, screenY);
    }

    public final void drawTeleportBubble(int a1, int y, int z, int w,
                                         int decorType, int h, int guard) {
        addGroundObject(a1, y, z, w, decorType, h, guard);
    }

    // -------------------------------------------------------------------------
    // buildTerrainTile  (obf: void b(int,int,int,int,int,int,int,int)  @clean L3544)
    // -------------------------------------------------------------------------

    /**
     * Render one NPC (rg[id]) at screen position: blits each appearance sprite layer
     * (with per-part walk offsets), then queues its chat bubble, action bubble,
     * health bar, damage splat and PK skull for later overlay passes.
     * (Oracle equivalent: drawNpc.)
     *
     * obf: void b(int objW, int frameW, int guard, int scale, int h, int x, int screenW, int id)
     *   args: var1=objW, var2=frameW, var3=guard, var4=scale, var5=h, var6=x, var7=screenW, var8=id
     *   (objW renamed from obf var1=w to avoid clashing with the WorldEntity class `w`.)
     */
    final void buildTerrainTile(int objW, int frameW, int guard, int scale,
                                int h, int x, int screenW, int id) {
        if (guard != 20) {                                    // guard (kept; side-effect call)
            // FLAG-FOR-ASSEMBLE: clean calls this.e((byte)-115) — obf void e(byte) @clean L12828,
            // a reset that clears de/Yc/Xd/Oc/Xf/qg/wh. That method is NOT present in the deob
            // Mudclient (incomplete reconcile) and has no readable alias in mud_map. Cannot invoke
            // without inventing it. Left as a no-op pending Assemble re-adding the e(byte) method.
            // this.e((byte)-115);
        }
        GameCharacter npc = rg[id];
        if (npc.npcIdOrColourBottom == 255) {                                   // colourBottom 255 -> invisible
            return;
        }

        int walkAnim = (npc.animationCurrent + (ug + 16) / 32) & 7;
        boolean flip = false;
        int step = walkAnim;
        if (step == 5) { flip = true; step = 3; }
        else if (step == 6) { flip = true; step = 2; }
        else if (step == 7) { flip = true; step = 1; }

        int frame = sf[npc.movingStep / 6 % 4] + 3 * step;
        if (npc.animationCurrent == 8) {                                     // attacking
            h -= scale * 5 / 100;
            step = 5;
            flip = false;
            walkAnim = 2;
            frame = Pc[jk / 5 % 8] + 3 * step;
        }
        if (npc.animationCurrent == 9) {                                     // being hit
            walkAnim = 2;
            step = 5;
            h += 5 * scale / 100;
            flip = true;
            frame = Og[jk / 6 % 8] + step * 3;
        }

        for (int layer = 0; layer < 12; layer++) {
            int bodyPart = Tg[walkAnim][layer];
            int itemId   = npc.equippedItem[bodyPart] - 1;               // equipped item id (-1 = none)
            if (itemId >= 0) {
                int dx = 0, dy = 0, f2 = frame;
                if (flip && step >= 1 && step <= 3) {
                    if (BZip.entityFlags[itemId] != 1) {                  // per-part hand/shield offsets
                        if (bodyPart == 4 && step == 1) {
                            f2 = 3 * step + sf[(npc.movingStep / 6 + 2) % 4];
                            dy = -3; dx = -22;
                        } else if (bodyPart == 4 && step == 2) {
                            dx = 0; dy = -8;
                            f2 = sf[(npc.movingStep / 6 + 2) % 4] + 3 * step;
                        } else if (bodyPart == 4 && step == 3) {
                            dy = -5;
                            f2 = step * 3 + sf[(2 + npc.movingStep / 6) % 4];
                            dx = 26;
                        } else if (bodyPart == 3 && step == 1) {
                            f2 = 3 * step + sf[(2 + npc.movingStep / 6) % 4];
                            dx = 22; dy = 3;
                        } else if (bodyPart == 3 && step == 2) {
                            dy = 8;
                            f2 = 3 * step + sf[(npc.movingStep / 6 + 2) % 4];
                            dx = 0;
                        } else if (bodyPart == 3 && step == 3) {
                            dx = -26;
                            f2 = sf[(2 + npc.movingStep / 6) % 4] + step * 3;
                            dy = 5;
                        }
                    } else {
                        f2 += 15;
                    }
                }
                if (step != 5 || DataStore.tamperScratch[itemId] == 1) {         // obf: nb.d
                    int sprite = WorldEntity.spriteOffsets[itemId] + f2;
                    int sw = li.spriteWidthFull[sprite];
                    int sh = li.spriteHeightFull[sprite];
                    int baseW = li.spriteWidthFull[WorldEntity.spriteOffsets[itemId]];
                    if (sw != 0 && sh != 0 && baseW != 0) {
                        dy = dy * screenW / sh;
                        dx = dx * frameW / sw;
                        int drawW = sw * frameW / baseW;
                        dx -= (drawW - frameW) / 2;
                        int colourA = LinkedQueue.sharedIntArray2[itemId];                   // animationCharacterColour
                        if (colourA == 1)      colourA = Dg[npc.colourHair];   // hair
                        else if (colourA == 2) colourA = ei[npc.colourTop];   // top
                        else if (colourA == 3) colourA = ei[npc.npcIdOrColourBottom];   // bottom
                        int colourB = Wh[npc.colourSkin];                      // skin
                        // obf arg order: li.a(x+dy, colourA, colourB, flip, objW, sprite, screenW, drawW, dx+h, 1)
                        li.spriteClipping(x + dy, colourA, colourB, flip, objW, sprite, screenW, drawW, dx + h, 1); // li.a(int×10)
                    }
                }
            }
        }

        // chat message bubble
        if (npc.messageTimeout > 0) {                                       // messageTimeout
            nf[Ef] = li.textWidth(1, 97, npc.name) / 2;
            if (nf[Ef] > 150) nf[Ef] = 150;
            uf[Ef] = li.textWidth(1, 72, npc.name) / 300 * li.textHeight(guard + 508305332, 1);
            tf[Ef] = frameW / 2 + h;
            ee[Ef] = x;
            Kc[Ef++] = npc.name;
        }

        // action bubble (item above head)
        if (npc.bubbleTimeout > 0) {                                       // bubbleTimeout
            je[jc] = h + frameW / 2;
            pe[jc] = x;
            jd[jc] = scale;
            ak[jc++] = npc.bubbleItem;                                  // bubbleItem
        }

        // health bar + damage splat
        if (npc.animationCurrent == 8 || npc.animationCurrent == 9 || npc.combatTimer != 0) {
            if (npc.combatTimer > 0) {                                  // combatTimer
                int barX = h;
                if (npc.animationCurrent == 8)      barX -= scale * 20 / 100;
                else if (npc.animationCurrent == 9) barX += 20 * scale / 100;
                int barLen = 30 * npc.healthCurrent / npc.healthMax;
                gd[Bc] = frameW / 2 + barX;
                Pk[Bc] = x;
                bf[Bc++] = barLen;
            }
            if (npc.combatTimer > 150) {
                int dmgX = h;
                if (npc.animationCurrent == 8)      dmgX -= 10 * scale / 100;
                else if (npc.animationCurrent == 9) dmgX += 10 * scale / 100;
                li.drawSprite(-1, tg + 11, screenW / 2 + x - 12, frameW / 2 + dmgX - 12); // li.b(int×4)
                li.drawstringRight(frameW / 2 + dmgX - 1, "" + npc.damageTaken, 0xFFFFFF, 0, 3, screenW / 2 + x + 5); // li.a(int,String,int,int,int,int)
            }
        }

        // PK skull (skullVisible == 1 and no action bubble)
        if (npc.skullVisible == 1 && npc.bubbleTimeout == 0) {
            int skullX = objW + h + frameW / 2;
            if (npc.animationCurrent == 8)      skullX -= scale * 20 / 100;
            if (npc.animationCurrent == 9)      skullX += 20 * scale / 100;
            int skullW = scale * 16 / 100;
            int skullH = 16 * scale / 100;
            li.spriteClipping(skullX - skullW / 2, x - scale * 10 / 100 - skullH / 2, skullH, skullW,
                 5924, tg + 13);   // li.f(int×6) -> Surface.spriteClipping(int,int,int,int,int,int)
        }
    }

    // -------------------------------------------------------------------------
    // getPlayer  (obf: ta b(int,byte)  @clean L3832  client.AC()
    // -------------------------------------------------------------------------

    /**
     * Resolve an in-view player GameCharacter (Tb[]) by server index; null if absent.
     * Side-effect: sets Bf when the sentinel byte != -123.
     * obf: ta b(int serverIndex, byte sentinel)
     */
    final GameCharacter getPlayer(int serverIndex, byte sentinel) {
        if (sentinel != -123) {
            Bf = -116;
        }
        for (int i = 0; i < de; i++) {                        // de = in-view player count
            if (serverIndex == Tb[i].serverIndex) {           // obf: .b
                return Tb[i];
            }
        }
        return null;
    }

    // -------------------------------------------------------------------------
    // addPlayer  (obf: ta a(int,int,int,byte,int,int)  @clean L13654  client.interlace()
    // -------------------------------------------------------------------------

    /**
     * Create or update a player entity and append it to the in-view list Tb[].
     * If the player was in the previous tick's known list (Ff[0..qj)) its animation
     * and waypoint ring are advanced; otherwise all state is freshly initialised.
     * obf: ta a(int animNext, int npcType, int tileX, byte sentinel, int tileZ, int serverIdx)
     */
    final GameCharacter addPlayer(int animNext, int npcType, int tileX,
                               byte sentinel, int tileZ, int serverIdx) {
        if (te[serverIdx] == null) {
            te[serverIdx] = new GameCharacter();
            te[serverIdx].serverIndex = serverIdx;            // obf: .b
        }
        GameCharacter player = te[serverIdx];                            // playersCache[serverIdx]

        boolean known = false;
        for (int i = 0; i < qj; i++) {                        // qj = known player count
            if (Ff[i].serverIndex == serverIdx) {             // obf: .b
                known = true;
                break;
            }
        }

        if (sentinel != 127) {                                // emit event/sound unless default
            this.drawTextField((byte)-81, -15, (String)null);
        }

        if (known) {
            player.animationNext = animNext;                  // obf: .D
            player.serverId = npcType;                         // obf: .t
            int wp = player.waypointCurrent;                   // obf: .o
            if (player.waypointsX[wp] != tileX || tileZ != player.waypointsY[wp]) { // obf: .k / .F
                player.waypointCurrent = wp = (wp + 1) % 10;
                player.waypointsX[wp] = tileX;
                player.waypointsY[wp] = tileZ;
            }
        } else {
            player.serverIndex = serverIdx;                    // obf: .b
            player.waypointCurrent = 0;                        // obf: .o
            player.stepCount = 0;                              // obf: .e
            player.waypointsX[0] = player.currentX = tileX;    // obf: .k / .i
            player.animationNext = player.animationCurrent = animNext; // obf: .D / .y
            player.movingStep = 0;                             // obf: .x
            player.serverId = npcType;                         // obf: .t
            player.waypointsY[0] = player.currentY = tileZ;    // obf: .F
        }

        Tb[de++] = player;                                    // append, de = in-view count
        return player;
    }

    // -------------------------------------------------------------------------
    // addNpc  (obf: ta d(int,int,int,int,int)  @clean L13871)
    // -------------------------------------------------------------------------

    /**
     * Create or update an NPC entity and append it to the in-view list rg[].
     * Known-check is against the previous tick's Zg[0..If).
     * obf: ta d(int tileZ, int serverIdx, int tileX, int junkGuard, int animNext)
     */
    final GameCharacter addNpc(int tileZ, int serverIdx, int tileX,
                            int junkGuard, int animNext) {
        if (We[serverIdx] == null) {
            We[serverIdx] = new GameCharacter();
            We[serverIdx].serverIndex = serverIdx;            // obf: .b
        }
        GameCharacter npc = We[serverIdx];                               // npcsCache[serverIdx]

        boolean known = false;
        for (int i = 0; i < If; i++) {                        // If = known npc count
            if (serverIdx == Zg[i].serverIndex) {             // obf: .b
                known = true;
                break;
            }
        }

        if (known) {
            npc.animationNext = animNext;                      // obf: .D
            int wp = npc.waypointCurrent;                      // obf: .o
            if (npc.waypointsX[wp] != tileX || tileZ != npc.waypointsY[wp]) { // obf: .k / .F
                npc.waypointCurrent = wp = (wp + 1) % 10;
                npc.waypointsX[wp] = tileX;
                npc.waypointsY[wp] = tileZ;
            }
        } else {
            npc.serverIndex = serverIdx;                       // obf: .b
            npc.waypointsX[0] = npc.currentX = tileX;          // obf: .k / .i
            npc.waypointCurrent = 0;                           // obf: .o
            npc.stepCount = 0;                                 // obf: .e
            npc.movingStep = 0;                                // obf: .x
            npc.animationNext = npc.animationCurrent = animNext; // obf: .D / .y
            npc.waypointsY[0] = npc.currentY = tileZ;          // obf: .F
        }
        // junkGuard: dead "-98 % ((0-junkGuard)/39)" expression — result discarded

        rg[Yc++] = npc;                                       // append, Yc = in-view count
        return npc;
    }

    // -------------------------------------------------------------------------
    // getNpc  (obf: ta d(int,int)  @clean L12247  client.originY()
    // -------------------------------------------------------------------------

    /**
     * Resolve an in-view NPC GameCharacter (rg[]) by server index; null if absent.
     * Side-effect: clears the local-player reference (wi) when sentinel != 220.
     * obf: ta d(int serverIndex, int sentinel)
     */
    final GameCharacter getNpc(int serverIndex, int sentinel) {
        for (int i = 0; i < Yc; i++) {                        // Yc = in-view npc count
            if (serverIndex == rg[i].serverIndex) {           // obf: .b
                return rg[i];
            }
        }
        if (sentinel != 220) {                                // NOTE: != (clean: var2 != 220)
            wi = null;
        }
        return null;
    }

    // -------------------------------------------------------------------------
    // drawIcon  (obf: void a(int,int,byte,int)  @clean L9719  client.D()
    // -------------------------------------------------------------------------

    /**
     * Blit a compass/minimap icon: base sprite (76) + arrow overlay (111), and for
     * the large variant (size <= -32) three extra border sprites.
     * Skeleton labels this "setCamera" but the body is pure 2D sprite blitting; the
     * real Hh-camera positioning is inlined elsewhere.
     * obf: void a(int x, int y, byte size, int spriteBase)
     * Moved to WidgetRenderer.drawIcon — delegates here.
     */
    private final void drawIcon(int x, int y, byte size, int spriteBase) {
        widgetRenderer.drawIcon(x, y, size, spriteBase);
    }

    // -------------------------------------------------------------------------
    // sendDuelItems  (obf: void b(int,int,int)  @clean L7479)
    // Moved to TradeDuelBankPackets.sendDuelItems
    // -------------------------------------------------------------------------

    private final void sendDuelItems(int p1, int delta, int invSlot) {
        tradePackets.sendDuelItems(p1, delta, invSlot);
    }

    // -------------------------------------------------------------------------
    // resetChatInput  (obf: void o(byte)  @clean L6259  client.NA()
    // -------------------------------------------------------------------------

    /**
     * Clear the chat/text-entry buffers (x = current line, inputPmFinal = committed line).
     * obf: void o(byte sentinel)
     */
    private final void resetChatInput(byte sentinel) {
        inputPmCurrent = "";    // obf: x (GameShell.x -> inputPmCurrent); bare `x` collided with java.awt.Component.x
        if (sentinel != -49) {
            Nc = 13;
        }
        inputPmFinal = "";
    }

    // -------------------------------------------------------------------------
    // sortFriendsList  (obf: void v(int)  @clean L9656)
    // -------------------------------------------------------------------------

    /**
     * Bubble-sort the friends list so online (held) entries float up and recently
     * logged-in (pressed) entries follow; swaps display-name (ac.z), real-name
     * (ua.h), note (cb.c) and the Fj[] status bitfield in tandem.
     * Skeleton mislabels this "sortDrawList"; it sorts the social list, not a draw list.
     * obf: void v(int guard)
     */
    final void sortFriendsList(int guard) {
        if (guard < 14) {                                     // guard (kept; scroll-state init)
            this.drawScrollbar2(-44, 54, 119, 125, true, 30);
        }
        boolean swapped = true;
        while (swapped) {
            swapped = false;
            for (int i = 0; i < FontWidths.listEntryCount - 1; i++) {               // n.g = friends list size
                boolean leftHeld     = (Fj[i]   & 2) != 0;
                boolean rightHeld    = (Fj[i+1] & 2) != 0;
                boolean leftPressed  = (Fj[i]   & 4) != 0;
                boolean rightPressed = (Fj[i+1] & 4) != 0;
                if ((!leftHeld && rightHeld) || (!leftPressed && rightPressed)) {
                    String tmp = DecodeBuffer.stringPool[i];
                    DecodeBuffer.stringPool[i]   = DecodeBuffer.stringPool[i + 1];
                    DecodeBuffer.stringPool[i+1] = tmp;
                    tmp = Surface.decoyStrings200[i];
                    Surface.decoyStrings200[i]   = Surface.decoyStrings200[i + 1];
                    Surface.decoyStrings200[i+1] = tmp;
                    tmp = CacheUpdater.archiveNames[i];
                    CacheUpdater.archiveNames[i]   = CacheUpdater.archiveNames[i + 1];
                    CacheUpdater.archiveNames[i+1] = tmp;
                    int tmpFj = Fj[i];
                    Fj[i]   = Fj[i + 1];
                    Fj[i+1] = tmpFj;
                    swapped = true;
                }
            }
        }
    }

    // -------------------------------------------------------------------------
    // findStringInData  (obf: static int a(byte[],String,int)  @clean L3450  client.ND()
    // -------------------------------------------------------------------------

    /**
     * Scan a name-table blob (count-prefixed array of 10-byte records) for a name and
     * return its 3-byte record offset; 0 if not found.  Record layout:
     *   [2..5] big-endian 4-byte name hash, [6..8] big-endian 3-byte offset.
     * Hash: 61-ary polynomial of (char - 32) over the uppercased name.
     *
     * obf: static int a(byte[] data, String name, int guard)
     *   guard: anti-tamper — when > -18 the real code returns 113 (kept, callers pass < -18).
     */
    static final int findStringInData(byte[] data, String name, int guard) {
        int recordCount = CacheFile.getUnsignedShort(data, 0);
        name = name.toUpperCase();
        if (guard > -18) {                                    // guard (kept verbatim)
            return 113;
        }

        int hash = 0;
        for (int i = 0; i < name.length(); i++) {
            hash = 61 * hash + name.charAt(i) - 32;
        }

        for (int i = 0; i < recordCount; i++) {
            int storedHash = (data[i * 10 + 5] & 0xFF)
                           + (data[i * 10 + 2] & 0xFF) * 0x1000000
                           + (data[i * 10 + 3] & 0xFF) * 0x10000
                           + (data[i * 10 + 4] & 0xFF) * 0x100;
            int offset = (data[i * 10 + 6] & 0xFF) * 0x10000
                       + (data[i * 10 + 7] & 0xFF) * 0x100
                       + (data[i * 10 + 8] & 0xFF);
            if (storedHash == hash) {
                return offset;
            }
        }
        return 0;
    }

    // -------------------------------------------------------------------------
    // readDefString  (obf: static String a(int,tb,int)  @clean L5386  client.CD()
    // -------------------------------------------------------------------------

    /**
     * Read a length-prefixed definition string from a Buffer at the given offset,
     * honouring a caller minimum length; decode the raw bytes via CharTable.a().
     * Returns "Cabbage" (STRINGS[32]) on any error.
     * obf: static String a(int offset, tb buf, int minLen)
     */
    static final String readDefString(int offset, Buffer buf, int minLen) {
        try {
            int len = buf.getSmartUnsigned();
            if (minLen > len) {
                len = minLen;
            }
            byte[] raw = new byte[len];
            buf.offset = buf.offset + SurfaceImageProducer.bzip.decode(buf.data, raw, offset, buf.offset, len);
            return CharTable.decodeBytes(len, offset ^ -124, 0, raw);
        } catch (Exception ex) {
            return STRINGS[32];                               // "Cabbage" — error sentinel
        }
    }

    /**
     * Cross-package alias of {@link #readDefString} under its original obf name {@code a},
     * called by SpriteScaler (ea) as {@code Mudclient.a(0, buf, Short.MAX_VALUE)}.
     */
    public static final String a(int offset, Buffer buf, int minLen) {
        return readDefString(offset, buf, minLen);
    }


    // =========================================================================
    // ===== ui_a =====
    // =========================================================================
// Group: ui (first 18 methods, listed order)
// Class: Mudclient (obf: client), package client, extends GameShell (obf: e)
// Field names from MUDCLIENT_SKELETON.md; class names from NAMING.md.
//
// RE-AUDITED against decompiled/normalized-clean/client.java (real Vineflower source).
// The previous version of this file was written against the DEFECTIVE base where these
// method bodies were missing/reconstructed; several were materially WRONG. See // FIX notes.
//
// Obfuscation stripped on read:
//   - opaque predicate `boolean varN = vh;` (always false) + every `if(!varN) break label;`
//     (always taken) and `if(varN) ...` (dead) — flattened to straight-line control flow.
//   - `++<counter>;` profiling bumps removed.
//   - try/catch(RuntimeException){ throw i.a(e, il[hasPainted]+...) } wrappers unwrapped.
//   - anti-tamper `if (param != <const>) <junk>` guards + dummy-param checks + junk
//     modulo expressions (`-121 / ((var1-19)/42)` etc.) removed.
//   - `~a == ~b`  →  `a == b`;  `~a < ~b`  →  `a > b`;  `~a <= -1` → `a >= 0`  (idiom unmasked).
//
// Class-token map applied per NAMING.md (corrects fabricated names from the old pass):
//   li=surface(ua=Surface), Jh=clientStream(da), mg=incomingPacket(ja), Hh=scene(lb=Scene),
//   Ek=world(k=World), wi=wi(ta), zh=friendsList/He=chatList/Wf=ignoreList(wb),
//   ac=DecodeBuffer  (ac.x[] = item display names in this build — NOT "EntityDef"),
//   fa=ClientIOException (fa.e[] = item stackable flag),
//   kb=InputState (kb.b[]/kb.c[] item tables; kb.a = applet host),
//   h=TextEncoder (h.c[] = item inventory sprite ids),  ga=CharTable (ga.b[] = item examine),
//   ua=Surface (ua.mouseButtonDown[] sprite base, ua.h[] friend names),  o=ISAAC (o.a(a,?,r,g)=ARGB pack),
//   mb=Utility, s=FontBuilder, d=CacheFile, f=RecordLoader, nb=DataStore, pa=ImageLoader,
//   l=Globals (l.c[] = chat-history names), ia=SpriteScaler (ia.a[] ignore, ia.g[] msg),
//   cb=CacheUpdater (cb.c[] friend display names).

    // -------------------------------------------------------------------------
    // drawActiveInterface  — obf: void mouseX(int)
    // -------------------------------------------------------------------------

    /** Dispatch to whichever modal panel / overlay is currently open, in priority order;
     *  if none is open, render the in-Ek frame (HUD tabs, minimap, Ek/inventory tab,
     *  player context menu).  Param is the anti-tamper sentinel `bj`.
     *
     *  FIX vs old: the old version stopped after the panel dispatch and stuffed the entire
     *  in-Ek section into a comment, and mis-routed several panels
     *  (h(127)=drawDuelConfirm not drawSocialDialog; d(false)=drawReportNameEntry;
     *  M=drawShop and hasPainted=drawTradeConfirmWindow placement). Rewritten from the clean source. */
    private final void drawActiveInterface(int param) {
        boolean inWorld = false;

        // ---- modal-panel dispatch (each branch draws then returns) ----
        if (param != this.bj) {           // anti-tamper sentinel mismatch
            this.clearScreen((byte) 120);
        } else if (this.Oh) {
            this.drawWelcome(param - 4853);
        } else if (this.mh) {
            this.drawChat((byte) -115);
        } else if (this.le == 1) {
            this.drawWildernessWarning(120);
        } else if (this.Fe && this.ai == 0) {
            this.drawBank(-122);
        } else if (this.uk && this.ai == 0) {
            this.drawShop(-89);
        } else if (this.Xj) {
            this.drawTradeConfirmWindow(-54);
        } else if (this.Hk) {
            this.drawTrade((byte) 8);
        } else if (this.dd) {
            this.drawDuelConfirm(-33);
        } else if (this.Pj) {
            this.drawDuel(param ^ 40);
        } else if (this.Vf == 1) {        // ~Vf != -2  ⟺  Vf != 1 → else-branch covers Vf==1 path below
            // (Vf == 1: fall through to report-name entry)
            this.drawReportNameEntry(false);
        } else if (this.Vf == 2) {
            this.drawReportAbuse(-28949);
        } else if (this.Bj == 0) {
            inWorld = true;               // a trade/social text box is closed → show world
        } else {
            this.drawSocialDialog((byte) 127);   // RENDER-BUG FIX (clean client.java:1410 this.h((byte)127) = drawSocialDialog, NOT drawDuelConfirm/h(int))
        }

        // NOTE on the Vf/Bj chain (clean source, lines 1395-1419):
        //   if (Vf != 1) { if (Vf==2) z(); if (Bj==0) inWorld=true; else h(127); }
        //   <then unconditionally> d(false)=drawReportNameEntry
        // The Vineflower control flow falls through report-abuse/trade-confirm into
        // drawReportNameEntry; the if/else-if ladder above preserves the same outcomes
        // because exactly one branch fires per tick given the mutually-exclusive panel flags.

        // ---- queued-action flush (runs every tick, even with a panel open) ----
        if (this.gc != 0) {
            this.sendQueuedActions((byte) -43);
        }

        // ---- in-world frame (only when no panel is open: Bj==0 path set inWorld) ----
        if (inWorld) {
            if (this.Ph) {
                this.sendDialogAnswer(-312);
            }
            // combat-style change pending (wi.y 8 or 9 = attacking)
            if (this.wi.animationCurrent == 8 || this.wi.animationCurrent == 9) {  // obf: wi.y
                this.sendCombatStyle((byte) 114);
            }
            this.drawInventoryTab(param ^ 1);   // clean I(int) L1450: this.D(var1^1) — qc hover/leave tracker (was wrongly drawMinimap = k(int), which clean I never calls)

            boolean showLists = !this.Ph && !this.se;
            if (showLists) {
                this.zh.setCount(0);   // obf: zh.d(0) -> MessageList.setCount (truncate friends/menu list)
            }
            // qc selects the active main tab
            if (this.qc == 0 && showLists) this.drawGameFrame(param ^ 2); // s = world render
            if (this.qc == 1) this.menus.handleInventoryClick(-15252, showLists);    // clean I(int) L1461: this.a(-15252,var3) = handleInventoryClick (obf a(int,boolean)), NOT drawGameSettings
            if (this.qc == 2) this.drawUiTabMinimap(showLists, (byte) 125); // obf a(boolean,byte) -> drawUiTabMinimap
            if (this.qc == 3) this.drawUiTabStats(showLists, param ^ 0);     // clean I(int) L1469: this.c(var3, var1^0) = drawUiTabStats (obf c(boolean,int) @clean L12938)
            if (this.qc == 4) this.drawUiTabMagic(showLists, (byte) -74);   // obf b(boolean,byte) -> drawUiTabMagic
            if (this.qc == 5) this.incoming.applyAppearanceUpdate(showLists, false); // obf a(boolean,boolean) -> applyAppearanceUpdate
            if (this.qc == 6) this.drawGameSettings(15, showLists);         // obf b(int,boolean) -> drawGameSettings

            if (!this.se && !this.Ph) this.drawPlayerMenu(-128);
            if (this.se && !this.Ph)  this.updateTimers((byte) -106);
        }

        this.Cf = 0;
    }

    // -------------------------------------------------------------------------
    // drawGameFrame  — obf: void f(int)
    // -------------------------------------------------------------------------

    /** Render the in-game screen each tick (param must be 13). Four exclusive modes:
     *    rk != -1 → "System update in progress" banner;
     *    Kg       → the obf w(int) overlay screen (drawCharDesignControls in this build);
     *    Qk       → sleep CAPTCHA screen;
     *    otherwise → rebuild the 3D Hh from World + entities, place region/transition
     *                labels, run the Ek render, then drawActiveInterface + chat tabs + blit.
     *
     *  NOTE: the skeleton names this slot "drawHud". That is misleading — the old "drawHud" body
     *  was ONLY the sleep-CAPTCHA `else` sub-branch of this method, lifted out and mislabelled as
     *  the whole thing (and it poked the wrong flags). The real f(int) is the per-tick game-frame
     *  driver; renamed drawGameFrame and rewritten in full from the clean source. */
    private final void drawGameFrame(int param) {
        if (param != 13) return;

        // BOOT HOOK diagnostic (env-gated): trace the in-game gating state once per ~50 frames
        // so the headless bring-up can see whether the 3D render branch is reached.
        if (System.getenv("RSC_FBUFFER_DUMP") != null) {
            this.fbufferDiag++;
            if ((this.fbufferDiag % 50) == 1) {
                System.out.println("[DBGgf] frame=" + this.fbufferDiag + " rk=" + this.rk
                    + " Kg=" + this.Kg + " Qk=" + this.Qk
                    + " playerAlive=" + (this.Hh != null && this.Hh.playerAlive)
                    + " wi=" + (this.wi != null ? (this.wi.currentX + "," + this.wi.currentY) : "null"));
            }
        }

        // clean: if (~this.rk != -1)  ==  (this.rk != 0)  [system-update banner; rk inits to 0].
        // A prior transcription rendered this as `rk != -1`, which is true at startup (rk==0) and
        // wrongly short-circuited every in-game frame into the banner branch (3D scene never drawn).
        if (this.rk != 0) {
            // --- system-update countdown banner ---
            this.li.fade2black(0xF8F8F9);
            this.li.drawStringCenter(this.screenWidth / 2, STRINGS[371], 0xFF0000, 0, 7, this.screenHeight / 2);
            this.drawChatHistoryTabs(param - 8);
            this.li.draw(this.graphics, this.originX, 256, this.originY);
            return;
        }
        if (this.Kg) {
            this.drawCharDesignControls(-13759);   // obf w(int): the Kg-screen overlay
            return;
        }

        if (!this.Qk) {
            // --- normal in-world: rebuild + render the 3D scene ---
            if (this.Hh.playerAlive) {
                // hide/show object models per active door/curtain layer of current floor (yj)
                for (int i = 0; i < 64; i++) {
                    this.Ek.removeModel(this.Hh.roofModels[this.yj][i]);
                    if (this.yj == 0) {
                        this.Ek.removeModel(this.Hh.wallModels[1][i]);
                        this.Ek.removeModel(this.Hh.roofModels[1][i]);
                        this.Ek.removeModel(this.Hh.wallModels[2][i]);
                        this.Ek.removeModel(this.Hh.roofModels[2][i]);
                    }
                    this.zf = true;
                    // if we are on the ground floor and standing under a roof tile, hide upper floors
                    if (this.yj == 0 && (this.Hh.objectAdjacency[this.wi.currentX / 128][this.wi.currentY / 128] & 128) == 0) {
                        this.Ek.addModel(this.Hh.roofModels[this.yj][i]);
                        if (this.yj == 0) {
                            this.Ek.addModel(this.Hh.wallModels[1][i]);
                            this.Ek.addModel(this.Hh.roofModels[1][i]);
                            this.Ek.addModel(this.Hh.wallModels[2][i]);
                            this.Ek.addModel(this.Hh.roofModels[2][i]);
                        }
                        this.zf = false;
                    }
                }

                // region-name announcement banners, latched on region change
                if (this.bl != this.Mg) {
                    this.bl = this.Mg;
                    for (int i = 0; i < this.eh; i++) {
                        // vc[i] holds scenery model ids triggering location labels
                        if (this.vc[i] == 97)   this.drawTextField((byte) 48, i, STRINGS[376] + (this.Mg + 1));
                        if (this.vc[i] == 274)  this.drawTextField((byte) 58, i, STRINGS[361] + (this.Mg + 1));
                        if (this.vc[i] == 1031) this.drawTextField((byte) 103, i, STRINGS[364] + (this.Mg + 1));
                        if (this.vc[i] == 1036) this.drawTextField((byte) 89, i, STRINGS[375] + (this.Mg + 1));
                        if (this.vc[i] == 1147) this.drawTextField((byte) 18, i, STRINGS[379] + (this.Mg + 1));
                    }
                }
                if (this.yg != this.Nc) {
                    this.yg = this.Nc;
                    for (int i = 0; i < this.eh; i++) {
                        if (this.vc[i] == 51)  this.drawTextField((byte) 23, i, STRINGS[368] + (this.Nc + 1));
                        if (this.vc[i] == 143) this.drawTextField((byte) 100, i, STRINGS[381] + (this.Nc + 1));
                    }
                }
                if (this.Sg != this.pj) {
                    this.Sg = this.pj;
                    for (int i = 0; i < this.eh; i++) {
                        if (this.vc[i] == 1142) this.drawTextField((byte) 89, i, STRINGS[372] + (this.pj + 1));
                    }
                }

                // (re)place the per-tick scene GameModels (players, npcs, ground items, projectiles)
                this.Ek.reduceSprites(this.qe);   // clear last frame's dynamic models
                this.qe = 0;

                // --- players in view (this-tick list rg, count Yc) ---
                for (int i = 0; i < this.Yc; i++) {
                    GameCharacter player = this.rg[i];
                    if (player.npcIdOrColourBottom != 255) {
                        int px = player.currentX;
                        int pz = player.currentY;
                        int py = -this.Hh.getElevation(px, pz);
                        int model = this.Ek.addSprite(i + 5000, pz, i + 10000, px, py, 145, 220, (byte) 109);
                        this.qe++;
                        if (this.wi == player) this.Ek.setLocalPlayer(32768, model);
                        if (player.animationCurrent == 8) this.Ek.setSpriteTranslateX(param + 24, model, -30);
                        if (player.animationCurrent == 9) this.Ek.setSpriteTranslateX(param ^ 45, model, 30);
                    }
                }
                // --- bubble/loot-overhead models for players (b == projectile/ranged target) ---
                for (int i = 0; i < this.Yc; i++) {
                    GameCharacter player = this.rg[i];
                    if (player.projectileRange != 0) {           // has an active projectile
                        GameCharacter target = null;
                        if (player.attackingNpcServerIndex == -1) {
                            if (player.attackingPlayerServerIndex != -1) target = this.We[player.attackingPlayerServerIndex];  // ~z != 0 ⟺ z != -1
                        } else {
                            target = this.te[player.attackingNpcServerIndex];
                        }
                        if (target != null) {
                            int sx = player.currentX;
                            int sz = player.currentY;
                            int sy = -this.Hh.getElevation(sx, sz) - 110;
                            int tx = target.currentX;
                            int tz = target.currentY;
                            int ty = -this.Hh.getElevation(tx, tz) - Packet.legacyMaskTable[target.serverId] / 2;
                            // interpolate projectile position by progress player.projectileRange / nc
                            int ix = (tx * (this.nc - player.projectileRange) + sx * player.projectileRange) / this.nc;
                            int iy = (sy * player.projectileRange + ty * (this.nc - player.projectileRange)) / this.nc;
                            int iz = ((this.nc - player.projectileRange) * tz + sz * player.projectileRange) / this.nc;
                            this.Ek.addSprite(player.incomingProjectileSprite + this.kd, iz, 0, ix, iy, 32, 32, (byte) 109);
                            this.qe++;
                        }
                    }
                }
                // --- npcs in view (this-tick list Tb, count de) ---
                for (int i = 0; i < this.de; i++) {
                    GameCharacter npc = this.Tb[i];
                    int nx = npc.currentX;
                    int nz = npc.currentY;
                    int ny = -this.Hh.getElevation(nx, nz);
                    int model = this.Ek.addSprite(20000 + i, nz, i + 30000, nx, ny, SurfaceImageProducer.entityIndexTableC[npc.serverId], Packet.legacyMaskTable[npc.serverId], (byte) 109);
                    this.qe++;
                    if (npc.animationCurrent == 8) this.Ek.setSpriteTranslateX(86, model, -30);
                    if (npc.animationCurrent == 9) this.Ek.setSpriteTranslateX(param ^ 99, model, 30);
                }
                // --- ground items (Ah of them) ---
                for (int i = 0; i < this.Ah; i++) {
                    int gx = this.Zf[i] * this.Ug + 64;
                    int gz = this.Ug * this.Ni[i] + 64;
                    this.Ek.addSprite(40000 + this.Gj[i], gz, i + 20000, gx,
                        -this.Hh.getElevation(gx, gz) - this.Le[i], 96, 64, (byte) 109);
                    this.qe++;
                }
                // --- decorative/scenery overlay models (el of them) ---
                for (int i = 0; i < this.el; i++) {
                    int dx = 64 + this.Ug * this.Sc[i];
                    int dz = this.gi[i] * this.Ug + 64;
                    int kind = this.Oc[i];
                    if (kind == 0) {
                        this.Ek.addSprite(50000 + i, dz, i + 50000, dx,
                            -this.Hh.getElevation(dx, dz), 128, 256, (byte) 109);
                        this.qe++;
                    }
                    if (kind == 1) {
                        this.Ek.addSprite(i + 50000, dz, i + 50000, dx,
                            -this.Hh.getElevation(dx, dz), 128, 64, (byte) 109);
                        this.qe++;
                    }
                }

                this.li.interlace = false;
                this.li.blackScreen(true);
                this.li.interlace = this.interlace;

                // occasional ambient sparkle/firework on the upper floors
                if (this.yj == 4) {
                    int n1 = 40 + (int) (3.0 * Math.random());
                    int n2 = (int) (7.0 * Math.random()) + 40;
                    this.Ek.setLightFull(-50, n2, 0, -50, n1, -10);
                }

                // --- camera ---
                this.jc = 0;
                this.Bc = 0;
                this.Ef = 0;
                if (this.Td) {                       // auto-camera mode
                    if (this.Kh && !this.zf) {
                        int prev = this.si;
                        this.clearScreen((byte) 22);   // q(byte) — auto-rotate toward target
                        if (this.si != prev) {
                            this.Si = this.wi.currentY;
                            this.kg = this.wi.currentX;
                        }
                    }
                    this.ug = 32 * this.si;
                    this.Ek.clipFar3d = 3000;
                    this.Ek.clipFar2d = 3000;
                    this.Ek.fogZFalloff = 1;
                    this.Ek.fogZDistance = 2800;
                    int cx = this.kg + this.Be;
                    int cz = this.Si + this.oc;
                    this.Ek.setCameraOrientation(cx, cz, 2000, 912, param - 12362, 4 * this.ug,
                        -this.Hh.getElevation(cx, cz), 0);
                } else {
                    if (this.Kh && !this.zf) {
                        this.clearScreen((byte) 94);
                    }
                    if (!this.interlace) {
                        this.Ek.fogZFalloff = 1;
                        this.Ek.clipFar3d = 2400;
                        this.Ek.fogZDistance = 2300;
                        this.Ek.clipFar2d = 2400;
                    } else {
                        this.Ek.fogZFalloff = 1;
                        this.Ek.clipFar3d = 2200;
                        this.Ek.clipFar2d = 2200;
                        this.Ek.fogZDistance = 2100;
                    }
                    int cx = this.kg + this.Be;
                    int cz = this.Si + this.oc;
                    this.Ek.setCameraOrientation(cx, cz, 2 * this.ac, 912, -12349, this.ug * 4,
                        -this.Hh.getElevation(cx, cz), 0);
                }

                // --- run the world render and overlays ---
                this.Ek.render(-113);             // render scene → surface
                this.drawChat(param - 11);      // l(int): damage splats / ground-item sprites / health bars

                // walk-target click marker (xh = ttl)
                if (this.xh > 0) {
                    this.li.drawSprite(-1, 14 + this.tg + (24 - this.xh) / 6, this.Fd - 8, this.tj - 8);
                }
                if (this.xh < 0) {
                    this.li.drawSprite(-1, 18 + this.tg + (this.xh + 24) / 6, this.Fd - 8, this.tj - 8);
                }

                // system-update countdown text (kc = ticks remaining * 50)
                if (this.kc != -1) {
                    int secs = this.kc / 50;
                    int mins = secs / 60;
                    secs %= 60;
                    if (secs < 10) {
                        this.li.drawStringCenter(256, STRINGS[380] + mins + STRINGS[365] + secs,
                            0xFFFF00, 0, 1, this.screenHeight - 7);   // ":0" + secs
                    } else {
                        this.li.drawStringCenter(256, STRINGS[380] + mins + ":" + secs,
                            0xFFFF00, 0, 1, this.screenHeight - 7);
                    }
                }

                // wilderness depth indicator ("Wilderness level: hasPainted") based on Y past the wall
                if (!this.Ub) {
                    int depth = -this.sh - this.sk - (this.zg - 2203);
                    if (this.Ki + this.Lf + this.Qg > 2640) {
                        depth = -50;
                    }
                    if (depth > 0) {
                        int level = depth / 6 + 1;
                        this.li.drawSprite(-1, 13 + this.tg, this.screenHeight - 56, 453);
                        this.li.drawStringCenter(465, STRINGS[377], 0xFFFF00, 0, 1, this.screenHeight - 20);
                        this.li.drawStringCenter(465, STRINGS[362] + level, 0xFFFF00, 0, 1, this.screenHeight - 7);
                        if (this.le == 0) this.le = 2;
                    }
                    if (this.le == 0 && depth > -10 && depth <= 0) {
                        this.le = 1;
                    }
                }

                // --- friends-tab overlay messages while chat panel is on the friends tab ---
                if (this.Zh == 0) {
                    for (int i = 0; i < 100; i++) {
                        // skill/quest progress flash entries (pa.g = ttl)
                        if (ImageLoader.scratchBuf[i] > 0) {
                            String txt = NameTable.recentNames[i] + Utility.formatChatLine(BZip.entityNames[i], World.G[i], true, FontWidths.entryTypes[i]);
                            this.li.drawstring(BitBuffer.UNUSED_N[i], this.screenHeight - 18 - 12 * i, txt, 7, 0xFFFF00, (byte) 26, 1);
                        }
                    }
                }

                // --- chat / quest / private message panels (yd holds the 4 message areas) ---
                this.panelShop.hideWidget((byte) 56, this.Fh);
                this.panelShop.hideWidget((byte) 80, this.ud);
                this.panelShop.hideWidget((byte) 48, this.mc);
                if (this.Zh == 1) {
                    this.panelShop.showWidget(this.Fh, 115);
                } else if (this.Zh == 2) {
                    this.panelShop.showWidget(this.ud, 119);
                } else if (this.Zh == 3) {
                    this.panelShop.showWidget(this.mc, 127);
                }
                SpriteScaler.lineHeightOverride = 2;
                this.panelShop.render((byte) -35);
                SpriteScaler.lineHeightOverride = 0;

                this.li.drawSpriteAlpha(this.tg, 0, this.li.width - 200, 128, 3);
                this.drawActiveInterface(0);
                this.li.loggedIn = false;
                this.drawChatHistoryTabs(param - 8);

                // BOOT HOOK (env-gated, headless): when RSC_FBUFFER_DUMP is set, after the client
                // has been in-game for RSC_FBUFFER_FRAMES frames (default 100, so the region is
                // loaded + camera settled), write the Surface pixel buffer (li.pixels — obf ua.rb,
                // the int[] the software rasteriser/Scene.render draws into) DIRECTLY to a PNG via
                // BufferedImage TYPE_INT_RGB + ImageIO. This bypasses AWT/Xvfb entirely and shows
                // EXACTLY what the renderer produced, independent of any window-flush/exposure
                // artifact. Mirrors the RSC_AUTOLOGIN / RSC_AUTO_APPEARANCE hooks.
                //   Pass 1 (fbuffer.png): the composited frame as-is (server "Welcome" box still up).
                //   Pass 2 (live3d.png):  auto-dismiss the Welcome box (Oh=false, as the click handler
                //                         does) then dump again -> a clean view of the 3D viewport.
                // BOOT HOOK (env-gated, headless): when RSC_AUTO_WALK is set, periodically issue a
                // plain WALK_TO_POINT a few tiles around the spawn so the player MOVES — this
                // exercises World.route + region streaming + the camera-follow path + NPC/ground-item
                // delta updates that only run once the local player changes tile. It reuses the SAME
                // dispatch the object-walk menu uses (see handleSceneUpdates action==200:
                // drawScrollbar((byte)10, sh, dy, dx, true, Lf) -> walkTo(start=(sh,Lf), dest=(dx,dy))),
                // so it sends nothing the real client wouldn't. One walk every ~75 frames.
                if (System.getenv("RSC_AUTO_WALK") != null && this.wi != null) {
                    this.autoWalkFrames++;
                    if (this.autoWalkFrames % 100 == 0) {
                        // Walk a rectangular tour relative to the player's current authoritative tile
                        // (Lf,sh = scene-local tile X,Y from the last SEND_PLAYER_COORDS / op 191).
                        // Going several tiles per leg exercises World.route, region streaming, the
                        // camera-follow path, and the NPC/ground-item delta updates that only run once
                        // the local player changes tile. Uses walkTo directly with the proven
                        // object-walk arg convention (start=(sh,Lf)).
                        int[] dx = { 5, 5, 0, -5, -5, 0 };
                        int[] dy = { 0, 5, 5, 0, -5, -5 };
                        int k = this.autoWalkStep % dx.length;
                        this.autoWalkStep++;
                        int destX = this.Lf + dx[k];
                        int destY = this.sh + dy[k];
                        boolean sent = this.walkTo(this.sh, this.Lf, (byte) 0, false,
                                                   destX, destX, destY, destY, false);
                        System.out.println("[RSC_AUTO_WALK] step " + this.autoWalkStep
                            + " from tile(" + this.Lf + "," + this.sh + ") -> tile(" + destX + ","
                            + destY + ") sent=" + sent + " region(Qg,zg)=(" + this.Qg + "," + this.zg + ")");
                    }
                }

                // BOOT HOOK (env-gated, headless): when RSC_AUTO_TABS is set, cycle the active main UI
                // tab (qc: 0=world,1=inventory,2=map,4=magic/prayer,5=friends,6=options) every ~60 frames
                // so each tab's renderer is exercised over the session (surfacing any tab-draw crash).
                // qc==3 (stats) is skipped: its drawer method is absent from this deob (see qc==3 note below).
                if (System.getenv("RSC_AUTO_TABS") != null) {
                    this.autoTabFrames++;
                    if (this.autoTabFrames % 60 == 0) {
                        int[] tabs = { 1, 2, 4, 5, 6, 0 };
                        this.qc = tabs[this.autoTabStep % tabs.length];
                        this.autoTabStep++;
                        System.out.println("[RSC_AUTO_TABS] qc=" + this.qc);
                    }
                }

                if (!this.fbufferDumped && System.getenv("RSC_FBUFFER_DUMP") != null) {
                    this.fbufferInGameFrames++;
                    String fenv = System.getenv("RSC_FBUFFER_FRAMES");
                    int need = 100;
                    if (fenv != null) { try { need = Integer.parseInt(fenv.trim()); } catch (NumberFormatException nfe) {} }
                    if (this.fbufferInGameFrames >= need) {
                        this.dumpSurfaceToPng("fbuffer.png");
                        // dismiss the server-sent "Welcome to RuneScape" overlay so the next dump is
                        // an unobstructed 3D view (same effect as drawWelcome's click handler: Oh=false)
                        this.Oh = false;
                    }
                    // second dump a few frames after dismissing the Welcome box
                    if (this.fbufferInGameFrames >= need + 15) {
                        this.dumpSurfaceToPng("live3d.png");
                        this.fbufferDumped = true;
                    }
                }

                this.li.draw(this.graphics, this.originX, 256, this.originY);
            }
        } else {
            // --- sleep CAPTCHA screen (Qk == true) ---
            this.li.fade2black(0xF8F8F9);
            // scattered decorative "sleeping" words (~15% each, from each side)
            if (Math.random() < 0.15) {
                this.li.drawStringCenter((int) (Math.random() * 80.0), STRINGS[378],
                    (int) (1.6777215E7 * Math.random()), 0, 5, (int) (334.0 * Math.random()));
            }
            if (Math.random() < 0.15) {
                this.li.drawStringCenter(512 - (int) (80.0 * Math.random()), STRINGS[378],
                    (int) (Math.random() * 1.6777215E7), param ^ 13, 5, (int) (334.0 * Math.random()));
            }
            this.li.drawBox(this.screenWidth / 2 - 100, (byte) -103, 0, 160, 40, 200);
            this.li.drawStringCenter(this.screenWidth / 2, STRINGS[366], 0xFFFF00, param - 13, 7, 50);   // "Enter the word..."
            this.li.drawStringCenter(this.screenWidth / 2, STRINGS[373] + 100 * this.pg / 750 + "%",
                0xFFFF00, param - 13, 7, 90);                                          // fatigue %
            this.li.drawStringCenter(this.screenWidth / 2, STRINGS[367], 0xFFFFFF, 0, 5, 140);
            this.li.drawStringCenter(this.screenWidth / 2, STRINGS[374], 0xFFFFFF, param ^ 13, 5, 160);
            this.li.drawStringCenter(this.screenWidth / 2, this.inputTextCurrent + "*", 0x00FFFF, param - 13, 5, 180);   // typed input
            if (this.Zj != null) {
                this.li.drawStringCenter(this.screenWidth / 2, this.Zj, 0xFF0000, 0, 5, 260);            // error message
            }
            this.li.drawSprite(-1, 1 + this.Eh, 230, this.screenWidth / 2 - 127);                  // CAPTCHA sprite
            this.li.drawBoxEdge(this.screenWidth / 2 - 128, 257, 229, 27785, 42, 0xFFFFFF);
            this.drawChatHistoryTabs(5);
            this.li.drawStringCenter(this.screenWidth / 2, STRINGS[370], 0xFFFFFF, param - 13, 1, 290);
            this.li.drawStringCenter(this.screenWidth / 2, STRINGS[369], 0xFFFFFF, param ^ 13, 1, 305);
            this.li.draw(this.graphics, this.originX, 256, this.originY);
        }
    }

    /**
     * BOOT HOOK helper (env-gated bring-up only): dump the live Surface pixel buffer
     * ({@link client.scene.Surface#pixels}, obf {@code ua.rb} — the int[] the software
     * rasteriser and {@code Scene.render} write into) straight to a PNG via
     * {@link java.awt.image.BufferedImage} TYPE_INT_RGB + ImageIO, bypassing the AWT/Xvfb
     * window flush entirely. Used to prove what the renderer actually produced.
     */
    private void dumpSurfaceToPng(String name) {
        try {
            int w = this.li.width;
            int h = this.li.height;
            java.awt.image.BufferedImage bi =
                new java.awt.image.BufferedImage(w, h, java.awt.image.BufferedImage.TYPE_INT_RGB);
            // li.pixels is 0xRRGGBB, exactly TYPE_INT_RGB's layout — copy straight in.
            bi.setRGB(0, 0, w, h, this.li.pixels, 0, w);
            String outDir = System.getenv("RSC_FBUFFER_DIR");
            if (outDir == null) outDir = "/tmp/rsc-run";
            java.io.File f = new java.io.File(outDir, name);
            javax.imageio.ImageIO.write(bi, "png", f);
            // quick non-black sanity scan of the 3D viewport region (top portion)
            long nonBlack = 0;
            int vh = Math.min(h, 334);
            for (int yy = 0; yy < vh; yy++) {
                int row = yy * w;
                for (int xx = 0; xx < w; xx++) {
                    if ((this.li.pixels[row + xx] & 0xFFFFFF) != 0) nonBlack++;
                }
            }
            System.out.println("[RSC_FBUFFER_DUMP] wrote " + f.getAbsolutePath()
                + " (" + w + "x" + h + ") after " + this.fbufferInGameFrames
                + " in-game frames; non-black viewport pixels=" + nonBlack
                + " / " + ((long) w * vh));
        } catch (Throwable t) {
            System.out.println("[RSC_FBUFFER_DUMP] FAILED: " + t);
            t.printStackTrace();
        }
    }

    // -------------------------------------------------------------------------
    // drawWildernessWarning  — obf: void H(int)
    // -------------------------------------------------------------------------

    /** "Warning! Proceed with caution" wilderness-entry dialog. Sets le=2 to enter wilderness
     *  mode on click (either on the "Click here to proceed" line or outside the panel bounds). */
    // Moved to GameInterface.drawWildernessWarning — delegates here.
    private final void drawWildernessWarning(int param) {
        gameInterface.drawWildernessWarning(param);
    }

    // -------------------------------------------------------------------------
    // drawShop  — obf: void M(int)
    // -------------------------------------------------------------------------

    /** Shop buy/sell panel ("Buying and selling items"). Hit-tests the 4x5 item grid plus the
     *  buy and sell quantity buttons (1/5/10/50/X). Opcodes: 236 BUY_ITEM, 221 SELL_ITEM,
     *  166 CLOSE_SHOP.
     *
     *  FIX vs old: grid hit-test had a duplicated `relY > cellY` (should be `relY < cellY+34`);
     *  the buy "X" dialog used ClientIOException.a (fa.a) not "ArchiveReader.u". */
    // Moved to GameInterface.drawShop — delegates here.
    private final void drawShop(int param) {
        gameInterface.drawShop(param);
    }

    // -------------------------------------------------------------------------
    // drawBank  — obf: void r(int)
    // -------------------------------------------------------------------------

    /** Bank deposit/withdraw panel with up to 4 page tabs. Opcodes: 22 WITHDRAW, 23 DEPOSIT,
     *  212 CLOSE_BANK.  Selected slot is Rd / item sj; page index is xg; vj = items used.
     *
     *  FIX vs old: the withdraw/deposit packet sends live INSIDE the click-handling block
     *  (clean lines 2221-2309), not interleaved with the render of the quantity buttons as the
     *  old version had it. Page-1/2 tabs are only drawn when vj>48 (single page → no tabs).
     *  Withdraw-X uses CacheFile.m (d.m), deposit-X uses RecordLoader.c (f.c). */
    // Moved to GameInterface.drawBank — delegates here.
    private final void drawBank(int param) {
        gameInterface.drawBank(param);
    }

    // bankSend moved to TradeDuelBankPackets.bankSend
    final void bankSend(int opcode, int itemId, int amount, int magic) {
        tradePackets.bankSend(opcode, itemId, amount, magic);
    }

    // -------------------------------------------------------------------------
    // drawTrade  — obf: void n(byte)
    // -------------------------------------------------------------------------

    /** Trade offer window: your inventory (left, lc items in vf/xe drawn from px+217), their
     *  current offer (Qf/jj, mf items) and your committed offer (zj/Dd, Lk items). Handles a
     *  right-click "offer hasPainted" sub-menu via the ignoreList MessageList (Wf). Opcodes: 55
     *  ACCEPT_TRADE, 230 DECLINE_TRADE; offers go through sendTradeOffer/sendDuelOffer.
     *
     *  FIX vs old: old version had only a stub render with the wrong arrays (zc/of/wj are the
     *  DUEL buffers) and omitted the Cf==2 right-click menu builder and the third (zj/Dd) grid.
     *  Rewritten in full from the clean source. */
    // Moved to GameInterface.drawTrade — delegates here.
    private final void drawTrade(byte param) {
        gameInterface.drawTrade(param);
    }

    // -------------------------------------------------------------------------
    // drawTradeConfirm  — obf: void a(int,byte,int)
    // -------------------------------------------------------------------------

    /** Add/adjust items in the local trade-offer buffer (Qf/jj) from inventory slot invSlot,
     *  then push the whole offer with opcode 46 (PLAYER_ADDED_ITEMS_TO_TRADE_OFFER).
     *  `count` is the requested amount (-1 = single decrement via Tk loop).
     *
     *  FIX vs old: stackable cap test is kb.c[itemId] (InputState.c), not "EntityDef.c";
     *  stackable flag is fa.e[itemId] (ClientIOException.e). */
    // Moved to GameInterface.drawTradeConfirm — delegates here.
    private final void drawTradeConfirm(int count, byte action, int invSlot) {
        gameInterface.drawTradeConfirm(count, action, invSlot);
    }

    // -------------------------------------------------------------------------
    // drawTradeConfirmWindow  — obf: void hasPainted(int)
    // -------------------------------------------------------------------------

    /** "Please confirm your trade" window: your final items (Vb/Me, count Ui) and theirs
     *  (Lc/Bi, count nh), with Accept/Decline. Vi = you have accepted. Opcodes: 104
     *  CONFIRM_TRADE, 230 DECLINE_TRADE.
     *
     *  FIX vs old: the previous part file's "drawTradeConfirmWindow" was tied to the wrong obf
     *  method (a(boolean,boolean), the social-list panel) and only a stub. This is the real
     *  hasPainted(int) body (clean line 13749). drawActiveInterface dispatches it via the Xj flag. */
    // Moved to GameInterface.drawTradeConfirmWindow — delegates here.
    private final void drawTradeConfirmWindow(int param) {
        gameInterface.drawTradeConfirmWindow(param);
    }

    // -------------------------------------------------------------------------
    // drawDuelConfirm  — obf: void h(int)
    // -------------------------------------------------------------------------

    /** Duel confirm window: your stake (Nj items in xi/th) vs theirs (Ve items in xj/kf),
     *  the four rule flags, and accept/decline. Opcodes: 77 ACCEPT_DUEL, 197 CONFIRM_DUEL,
     *  230 DECLINE_DUEL.
     *
     *  FIX vs old: the two accept/decline hit-test rectangles had wrong right edges
     *  (old used `+70-35`; clean is left `83..188`, right `317..423`). */
    // Moved to GameInterface.drawDuelConfirm — delegates here.
    private final void drawDuelConfirm(int param) {
        gameInterface.drawDuelConfirm(param);
    }

    // -------------------------------------------------------------------------
    // drawDuel  — obf: void q(int)
    // -------------------------------------------------------------------------

    /** Duel setup window: your stake (lc inventory items vf/xe), opponent's offered items
     *  (Ke items Uf/df), opponent's committed stake (wj items zc/of), and the four rule
     *  checkboxes (No retreat fd / No magic Yi / No prayer vd / No weapons ff). A right-click
     *  on a stake cell opens an "offer hasPainted" sub-menu via the chatList MessageList (He).
     *  Opcodes: 8 DUEL_SETTINGS, 176 DUEL_ACCEPT, 197 DUEL_DECLINE.
     *
     *  FIX vs old: old version stubbed the right-click sub-menu builders and the menu-pick
     *  resolution, and mixed up the three render grids. Rewritten in full from the clean source. */
    // Moved to GameInterface.drawDuel — delegates here.
    private final void drawDuel(int param) {
        gameInterface.drawDuel(param);
    }

    // -------------------------------------------------------------------------
    // drawReportAbuse  — obf: void z(int)
    // -------------------------------------------------------------------------

    /** Report-abuse rule picker. Yb is the selected rule id (column base 1/7/12 from mouse X,
     *  plus the clicked row). On confirm sends opcode 206 (REPORT_ABUSE) with the player name
     *  (ec), the rule id, and the mute flag (ue).
     *
     *  FIX vs old: the old version used wrong string indices and an invented row layout. This is
     *  the faithful render from the clean source (3 columns x several rows of li.a/li.e calls). */
    private final void drawReportAbuse(int param) {
        // pick the column (rule base) from mouse X
        this.Yb = 0;
        boolean inColumn = true;
        if (this.mouseX >= 36 && this.mouseX < 176) {
            this.Yb = 1;
        } else if (this.mouseX >= 186 && this.mouseX < 326) {
            this.Yb = 7;
        } else if (this.mouseX >= 336 && this.mouseX < 476) {
            this.Yb = 12;
        } else {
            inColumn = false;
        }

        // within the column, add the clicked row to the base
        int y = 156;
        if (inColumn) {
            boolean rowHit = false;
            for (int row = 0; row < 6; row++) {
                int rowH = (row == 0) ? 30 : 18;
                if (this.mouseY > y - 12 && this.mouseY < y - 12 + rowH) {
                    if (this.Yb == 1)               { rowHit = true; this.Yb += row; }
                    else if (this.Yb == 7 && row < 5){ rowHit = true; this.Yb += row; }
                    else if (this.Yb == 12 && row < 3){ rowHit = true; this.Yb += row; }
                }
                y += 2 + rowH;
            }
            if (!rowHit) this.Yb = 0;
        } else {
            this.Yb = 0;
        }

        if (this.Cf != 0 && this.Yb != 0) {
            this.Jh.newPacket(206, param + 28949);   // REPORT_ABUSE
            this.Jh.outBuffer.putStringPrefixed(this.ec);        // reported player name
            this.Jh.outBuffer.putByte(this.Yb);         // rule id
            this.Jh.outBuffer.putByte(this.ue ? 1 : 0);// mute flag
            this.Jh.finishPacket(param ^ -8763);
            this.Vf = 0;
            this.inputTextFinal = "";
            this.inputTextCurrent = "";
            this.Cf = 0;
            return;
        }

        y += 15;
        if (this.Cf != 0) {
            this.Cf = 0;
            // click outside the panel → close
            if (this.mouseX < 31 || this.mouseY < 35 || this.mouseX > 481 || this.mouseY > 310) {
                this.Vf = 0;
                return;
            }
            // click on the "Send report" link area → close
            if (this.mouseX > 67 && this.mouseX < 446 && this.mouseY >= y - 15 && this.mouseY < y + 5) {
                this.Vf = 0;
                return;
            }
        }

        // --- render panel ---
        this.li.drawBox(31, (byte) -110, 0, 35, 275, 450);
        this.li.drawBoxEdge(31, 450, 35, 27785, 275, 0xFFFFFF);
        int ry = 50;
        this.li.drawstringRight(256, STRINGS[408], 0xFFFFFF, 0, 1, ry);          // title
        ry += 15;
        this.li.drawstringRight(256, STRINGS[411], 0xFFFFFF, param + 28949, 1, ry);
        ry += 15;
        this.li.drawstringRight(256, STRINGS[395], 0xFF8000, 0, 1, ry);          // orange warning
        ry += 15;
        ry += 10;
        this.li.drawstringRight(256, STRINGS[406], 0xFFFF00, 0, 1, ry);          // category header
        ry += 15;
        this.li.drawstringRight(256, STRINGS[407], 0xFFFF00, 0, 1, ry);
        ry += 18;
        this.li.drawstringRight(106, STRINGS[410], 0xFF0000, 0, 4, ry);          // column headers
        this.li.drawstringRight(256, STRINGS[415], 0xFF0000, 0, 4, ry);
        this.li.drawstringRight(406, STRINGS[403], 0xFF0000, param ^ -28949, 4, ry);
        ry += 18;

        // column selection-highlight boxes (rows of varying height) + rule labels
        if (this.Yb == 1)  this.li.drawBox(36,  (byte) 32,  0x303030, ry - 12, 30, 140);
        this.li.drawBoxEdge(36, 140, ry - 12, param ^ -7582, 30, 0x404040);
        if (this.Yb == 7)  this.li.drawBox(186, (byte) -106, 0x303030, ry - 12, 30, 140);
        this.li.drawBoxEdge(186, 140, ry - 12, 27785, 30, 0x404040);
        if (this.Yb == 12) this.li.drawBox(336, (byte) -99, 0x303030, ry - 12, 30, 140);
        this.li.drawBoxEdge(336, 140, ry - 12, 27785, 30, 0x404040);
        this.li.drawstringRight(106, STRINGS[414], this.Yb == 1  ? 0xFF8000 : 0xFFFFFF, 0, 0, ry);
        this.li.drawstringRight(256, STRINGS[401], this.Yb == 7  ? 0xFF8000 : 0xFFFFFF, param ^ param, 0, ry);
        this.li.drawstringRight(406, STRINGS[393], this.Yb == 12 ? 0xFF8000 : 0xFFFFFF, 0, 0, ry);
        ry += 12;
        this.li.drawstringRight(106, STRINGS[413], this.Yb == 1  ? 0xFF8000 : 0xFFFFFF, 0, 0, ry);
        this.li.drawstringRight(256, STRINGS[396], this.Yb == 7  ? 0xFF8000 : 0xFFFFFF, param ^ -28949, 0, ry);
        this.li.drawstringRight(406, STRINGS[412], this.Yb == 12 ? 0xFF8000 : 0xFFFFFF, 0, 0, ry);
        ry += 20;

        if (this.Yb == 2)  this.li.drawBox(36,  (byte) -111, 0x303030, ry - 12, 18, 140);
        this.li.drawBoxEdge(36, 140, ry - 12, param + 56734, 18, 0x404040);
        if (this.Yb == 8)  this.li.drawBox(186, (byte) -107, 0x303030, ry - 12, 18, 140);
        this.li.drawBoxEdge(186, 140, ry - 12, 27785, 18, 0x404040);
        if (this.Yb == 13) this.li.drawBox(336, (byte) -119, 0x303030, ry - 12, 18, 140);
        this.li.drawBoxEdge(336, 140, ry - 12, 27785, 18, 0x404040);
        this.li.drawstringRight(106, STRINGS[392], this.Yb == 2  ? 0xFF8000 : 0xFFFFFF, 0, 0, ry);
        this.li.drawstringRight(256, STRINGS[399], this.Yb == 8  ? 0xFF8000 : 0xFFFFFF, 0, 0, ry);
        this.li.drawstringRight(406, STRINGS[412], this.Yb == 13 ? 0xFF8000 : 0xFFFFFF, 0, 0, ry);
        ry += 20;

        if (this.Yb == 3)  this.li.drawBox(36,  (byte) -114, 0x303030, ry - 12, 18, 140);
        this.li.drawBoxEdge(36, 140, ry - 12, 27785, 18, 0x404040);
        if (this.Yb == 9)  this.li.drawBox(186, (byte) -127, 0x303030, ry - 12, 18, 140);
        this.li.drawBoxEdge(186, 140, ry - 12, 27785, 18, 0x404040);
        if (this.Yb == 14) this.li.drawBox(336, (byte) -117, 0x303030, ry - 12, 18, 140);
        this.li.drawBoxEdge(336, 140, ry - 12, 27785, 18, 0x404040);
        this.li.drawstringRight(106, STRINGS[409], this.Yb == 3  ? 0xFF8000 : 0xFFFFFF, 0, 0, ry);
        this.li.drawstringRight(256, STRINGS[416], this.Yb == 9  ? 0xFF8000 : 0xFFFFFF, param + 28949, 0, ry);
        this.li.drawstringRight(406, STRINGS[402], this.Yb == 14 ? 0xFF8000 : 0xFFFFFF, param + 28949, 0, ry);
        ry += 20;

        if (this.Yb == 4)  this.li.drawBox(36,  (byte) 118,  0x303030, ry - 12, 18, 140);
        this.li.drawBoxEdge(36, 140, ry - 12, 27785, 18, 0x404040);
        if (this.Yb == 10) this.li.drawBox(186, (byte) -104, 0x303030, ry - 12, 18, 140);
        this.li.drawBoxEdge(186, 140, ry - 12, 27785, 18, 0x404040);
        this.li.drawstringRight(106, STRINGS[404], this.Yb == 4  ? 0xFF8000 : 0xFFFFFF, 0, 0, ry);
        this.li.drawstringRight(256, STRINGS[397], this.Yb == 10 ? 0xFF8000 : 0xFFFFFF, 0, 0, ry);
        ry += 20;

        if (this.Yb == 5)  this.li.drawBox(36,  (byte) 31,  0x303030, ry - 12, 18, 140);
        this.li.drawBoxEdge(36, 140, ry - 12, 27785, 18, 0x404040);
        if (this.Yb == 11) this.li.drawBox(186, (byte) 62, 0x303030, ry - 12, 18, 140);
        this.li.drawBoxEdge(186, 140, ry - 12, param ^ -7582, 18, 0x404040);
        this.li.drawstringRight(106, STRINGS[405], this.Yb == 5  ? 0xFF8000 : 0xFFFFFF, 0, 0, ry);
        this.li.drawstringRight(256, STRINGS[417], this.Yb == 11 ? 0xFF8000 : 0xFFFFFF, 0, 0, ry);
        ry += 20;

        if (this.Yb == 6)  this.li.drawBox(36,  (byte) 82, 0x303030, ry - 12, 18, 140);
        this.li.drawBoxEdge(36, 140, ry - 12, param + 56734, 18, 0x404040);
        this.li.drawstringRight(106, STRINGS[398], this.Yb == 6  ? 0xFF8000 : 0xFFFFFF, 0, 0, ry);
        ry += 18;
        ry += 15;

        // "Click here to send report" link — yellow on hover
        int linkCol = 0xFFFFFF;
        if (this.mouseX > 196 && this.mouseX < 316 && this.mouseY > ry - 15 && this.mouseY < ry + 5) {
            linkCol = 0xFFFF00;
        }
        this.li.drawstringRight(256, STRINGS[391], linkCol, param + 28949, 1, ry);
    }

    // -------------------------------------------------------------------------
    // drawPlayerMenu  — obf: void L(int)
    // -------------------------------------------------------------------------

    /** Render the in-chat right-click "Choose option" menu (also handles friends/PM-history
     *  context labels), and on click route it: in wilderness combat it sends opcode 59
     *  (PLAYER_ATTACK), otherwise it runs the click via updateCamera (b(false,0)).
     *
     *  FIX vs old: the menu list is the zh MessageList (zh), not "chatList"; chat-
     *  history names come from Globals.c (l.c) not "GlobalStrings"; the `af>=0 || Bh>=0`
     *  guards had the inequality flipped in the old version. */
    private final void drawPlayerMenu(int param) {
        if (this.af >= 0 || this.Bh >= 0) {     // FIX: was `<= 0`
            this.zh.addEntrySimple(4000, "", STRINGS[121], 30192);   // "Cancel" entry
        }
        this.zh.sortEntries((byte) 16);
        int count = this.zh.getCount(-27153);
        if (param >= -120) return;   // anti-tamper guard

        // trim the list to at most 20 entries
        for (int i = count; i > 20; i--) {
            this.zh.removeEntry(102, i - 1);
        }

        // friends-tab / pm-history context label
        if (this.qc == 5) {   // ~qc == -6
            String label = null;
            if (this.pk == 0 && this.wk != 0) {
                if (this.wk >= 0) {
                    int idx = this.wk;
                    String suffix = "";
                    if ((Fj[idx] & 4) == 0) {
                        label = friendListNames[idx];        // clean L10747: var10 = ua.h[var5] (ua.h is String[]); SEG_00 mis-declares friendListNames as int[] -> field-type defect (also breaks siblings @3894/3930/3939/3983). FLAG FOR ASSEMBLE: friendListNames must be String[].
                        suffix = STRINGS[190];               // " - online"
                    } else {
                        label = STRINGS[188] + friendListNames[idx];    // "Message "
                        if (friendListWorlds[idx] != null) suffix = STRINGS[193] + friendListWorlds[idx];
                    }
                    if (friendListFormerNames[idx] != null && friendListFormerNames[idx].length() > 0) {
                        label = label + STRINGS[198] + friendListFormerNames[idx] + ")" + suffix;
                    } else {
                        label = label + suffix;
                    }
                } else {
                    int idx = -(2 + this.wk);
                    label = STRINGS[196] + friendListNames[idx];
                    if (friendListFormerNames[idx] != null && friendListFormerNames[idx].length() > 0) {
                        label = label + STRINGS[198] + friendListFormerNames[idx] + ")";
                    }
                }
            }
            if (this.pk == 1 && this.nj != 0) {
                if (this.nj >= 0) {
                    int idx = this.nj;
                    label = STRINGS[194] + ignoreListNames[idx];
                    if (ignoreListFormerNames[idx] != null && ignoreListFormerNames[idx].length() > 0) {
                        label = label + STRINGS[198] + ignoreListFormerNames[idx] + ")";
                    }
                } else {
                    int idx = -(2 + this.nj);
                    label = STRINGS[196] + ignoreListNames[idx];
                    if (ignoreListFormerNames[idx] != null && ignoreListFormerNames[idx].length() > 0) {
                        label = label + STRINGS[198] + ignoreListFormerNames[idx] + ")";
                    }
                }
            }
            if (label != null) {
                this.li.drawstring(label, 6, 14, 0xFFFF00, false, 1);
            }
        }

        count = this.zh.getCount(-27153);
        if (count <= 0) return;

        // find the last non-empty entry
        int lastNonEmpty = -1;
        for (int i = 0; i < count; i++) {
            String entry = this.zh.getEntryMessage((byte) 74, i);
            if (entry != null && entry.length() > 0) lastNonEmpty = i;
        }

        // compose the menu header
        String header = null;
        if ((this.Bh >= 0 || this.af >= 0) && count == 2) {
            header = STRINGS[192];   // "Choose option"
        } else if ((this.Bh <= 0 || this.af <= 0) && count > 1) {
            header = STRINGS[15] + this.zh.getEntryPrefix(0, (byte) 53) + " " + this.zh.getEntryMessage((byte) 75, 0);
        } else if (lastNonEmpty != -1) {
            header = this.zh.getEntryMessage((byte) 54, lastNonEmpty) + STRINGS[159] + this.zh.getEntryPrefix(0, (byte) 53);
        }
        if (count == 2 && header != null) header = header + STRINGS[189];
        if (count > 3 && header != null) header = header + STRINGS[195] + (count - 1) + STRINGS[191];
        if (header != null) this.li.drawstring(header, 6, 14, 0xFFFF00, false, 1);

        // position the popup near the cursor when it pops (single-button vs two-button modes)
        boolean popMenu = (!this.Yh && this.Cf == 1) || (this.Yh && this.Cf == 2 && count == 1);
        if (popMenu || (!this.Yh && this.Cf == 2) || (this.Yh && this.Cf == 1)) {
            if (!popMenu && (this.Yh ? this.Cf != 1 : this.Cf != 2)) {
                return;
            }
            int w = this.zh.getPanelWidth(16256);
            int hgt = this.zh.getPanelHeight(-21224);
            this.rh = this.mouseX - w / 2;
            this.se = true;
            this.fg = this.mouseY - 7;
            if (this.rh < 0) this.rh = 0;
            if (this.fg < 0) this.fg = 0;
            this.Cf = 0;
            if (this.fg + hgt > 316) this.fg = 315 - hgt;
            if (w + this.rh > 510) this.rh = 510 - w;
            return;
        }

        // confirm: send attack if in wilderness combat, else dispatch the click
        if (this.altDown && this.ctrlDown && this.Hc) {
            this.Jh.newPacket(59, 0);          // PLAYER_ATTACK
            this.Jh.outBuffer.putShort(this.rf);
            this.Jh.outBuffer.putShort(this.Cg);
            this.Jh.finishPacket(21294);
        } else {
            this.incoming.handleSceneUpdates(false, 0);          // b(boolean,int)
        }
        this.Cf = 0;
    }

    // -------------------------------------------------------------------------
    // drawOptionsTab  — obf: void f(boolean)
    // -------------------------------------------------------------------------

    /** Load the "Configuration" options archive and apply it. Plays the menu sound on open. */
    final void drawOptionsTab(boolean playSfx) {
        if (playSfx) {
            this.sound.playSound((byte) 77, null);
        }
        byte[] data = this.readDataFile(STRINGS[225], 10, 0, 78);
        if (data != null) {
            SocketFactory.initGameData(data, (byte) 100, this.Pg);   // m = SocketFactory: apply options
        } else {
            this.Vc = true;
        }
    }

    // -------------------------------------------------------------------------
    // drawChatHistoryTabs  — obf: void A(int)
    // -------------------------------------------------------------------------

    /** Render the chat / quest / private / friends history tab strip + the settings icon at the
     *  bottom of the screen. Tabs blink red when they have unread activity (Ee/Qe/Vj/Mh timers).
     *  Only renders when param == 5. */
    private final void drawChatHistoryTabs(int param) {
        this.li.drawSprite(-1, this.tg + 23, this.screenHeight - 4, 0);   // tab-strip background sprite
        if (param != 5) return;

        // tab 0 — Chat (active if Zh==0; blink if Ee%30 > 15)
        int col = ISAAC.packColor(200, 9570, 255, 200);
        if (this.Zh == 0) col = ISAAC.packColor(255, 9570, 50, 200);
        if (this.Ee % 30 > 15) col = ISAAC.packColor(255, 9570, 50, 50);
        this.li.drawstringRight(54, STRINGS[269], col, 0, 0, this.screenHeight + 6);

        // tab 1 — Quest (active if Zh==1; blink if Qe%30 > 15)
        col = ISAAC.packColor(200, 9570, 255, 200);
        if (this.Zh == 1) col = ISAAC.packColor(255, param + 9565, 50, 200);
        if (this.Qe % 30 > 15) col = ISAAC.packColor(255, param ^ 0x2567, 50, 50);
        this.li.drawstringRight(155, STRINGS[272], col, 0, 0, this.screenHeight + 6);

        // tab 2 — Private (active if Zh==2; blink if Vj%30 > 15)
        col = ISAAC.packColor(200, 9570, 255, 200);
        if (this.Zh == 2) col = ISAAC.packColor(255, 9570, 50, 200);
        if (this.Vj % 30 > 15) col = ISAAC.packColor(255, param + 9565, 50, 50);
        this.li.drawstringRight(255, STRINGS[271], col, 0, 0, this.screenHeight + 6);

        // tab 3 — Friends (active if Zh==3; blink if Mh%30 > 15)
        col = ISAAC.packColor(200, 9570, 255, 200);
        if (this.Zh == 3) col = ISAAC.packColor(255, 9570, 50, 200);
        if (this.Mh % 30 > 15) col = ISAAC.packColor(255, param ^ 0x2567, 50, 50);
        this.li.drawstringRight(355, STRINGS[268], col, 0, 0, this.screenHeight + 6);

        // settings/report icon
        this.li.drawstringRight(457, STRINGS[120], 0xFFFFFF, 0, 0, this.screenHeight + 6);
    }

    // -------------------------------------------------------------------------
    // drawChat  — obf: void l(int)
    // -------------------------------------------------------------------------

    /** Draw the floating in-Ek overlays after the 3D Hh: hit-damage splats (Kc text at
     *  tf/ee with width/height nf/uf, de-overlapped vertically), ground-item sprites
     *  (je/pe/jd/ak), and entity health bars (gd/Pk/bf).
     *
     *  FIX vs old: the old version threw UnsupportedOperationException with an (incorrect) claim
     *  that this is the chat-scrollback panel. It is NOT — l(int) is the post-render overlay
     *  pass. Reconstructed in full from the clean source. */
    private final void drawChat(int param) {
        // --- damage splats: nudge each up so it doesn't overlap an earlier one ---
        int gap = this.li.textHeight(508305352, 1);   // line height
        for (int i = 0; i < this.Ef; i++) {
            int sx = this.tf[i];
            int sy = this.ee[i];
            int sw = this.nf[i];
            int sh = this.uf[i];
            boolean moved = true;
            while (moved) {
                moved = false;
                for (int j = 0; j < i; j++) {
                    if (sy + sh > this.ee[j] - gap
                            && sy - gap < this.ee[j] + this.uf[j]
                            && sx - sw < this.tf[j] + this.nf[j]
                            && sw + sx > this.tf[j] - this.nf[j]
                            && sy > this.ee[j] - gap - sh) {
                        sy = this.ee[j] - (gap + sh);
                        moved = true;
                    }
                }
            }
            this.ee[i] = sy;
            this.li.centrepara(300, this.Kc[i], sx, 55, 1, sy, false, 0xFFFF00);
        }

        // RENDER-BUG FIX (clean client.java:5496-5498): restore the dropped ground-item reset.
        // Without it, stale ground-item icons (ak) persist across frames.
        if (param != 2) this.ak = null;

        // --- ground-item icons ---
        for (int i = 0; i < this.jc; i++) {
            int gx = this.je[i];
            int gy = this.pe[i];
            int scale = this.jd[i];
            int itemId = this.ak[i];
            int bw = 39 * scale / 100;
            int bh = 27 * scale / 100;
            int boxY = gy - bh;
            this.li.drawActionBubble(this.tg + 9, (byte) -122, bh, gx - bw / 2, bw, boxY, 85);
            int iw = scale * 36 / 100;
            int ih = 24 * scale / 100;
            this.li.spriteClipping(boxY + bh / 2 - ih / 2, TextEncoder.scratchIntArray2[itemId], 0, false, 0,
                Surface.unusedIntsBb[itemId] + this.sg, ih, iw, gx - iw / 2, 1);
        }

        // --- entity health bars (green = remaining, red = lost) ---
        for (int i = 0; i < this.Bc; i++) {
            int hx = this.gd[i];
            int hy = this.Pk[i];
            int pct = this.bf[i];                 // 0..30 = green width
            this.li.drawBoxAlpha(192, hx - 15, 5, 0, hy - 3, pct, 0x00FF00);
            this.li.drawBoxAlpha(192, pct - 15 + hx, 5, 0, hy - 3, 30 - pct, 0xFF0000);
        }
    }

    // -------------------------------------------------------------------------
    // drawWelcome  — obf: void j(int)
    // -------------------------------------------------------------------------

    /** "Welcome to RuneScape" box on login: last-login, recovery-questions reminder, unread
     *  messages, subscription/members status, and a "Click here to play" dismiss button.
     *  Sb = subscription-days marker (201 = none); ce = recovery set time; id = unread count. */
    // Moved to GameInterface.drawWelcome — delegates here.
    private final void drawWelcome(int param) {
        gameInterface.drawWelcome(param);
    }

    // -------------------------------------------------------------------------
    // playSound / initSounds  — moved to client.ClientSound (extract-delegate).
    // Invoke via the `sound` delegate field: sound.playSound(...), sound.initSounds(...).
    // -------------------------------------------------------------------------


    // =========================================================================
    // ===== ui_b =====
    // =========================================================================
// Methods 19–36 of the "ui" group from MUDCLIENT_SKELETON.md.
// RE-AUDITED against decompiled/normalized-clean/client.java (ground truth) +
// cfr/client.java cross-ref. The previous version was written against a DEFECTIVE
// decompilation (72 client methods missing) and had numerous reconstructed-logic
// bugs (inverted compass strips, scrambled inventory-tab bounds, wrong mouse-button
// constants in the settings panel, missing draw-list capacity guard, etc.). Those
// are now fixed against the clean base.
//
// All methods belong to class Mudclient (obf: client), extends GameShell (obf: e).
// Naming: fields from MUDCLIENT_SKELETON.md; other classes from NAMING.md.
// Obfuscation stripped: opaque-predicate guards (vh/bl), profiling ++counters,
// try/catch ErrorHandler wrappers, anti-tamper dummy params, junk masks, and the
// ~x>~y / ~x==-hasPainted sign idioms (rewritten to plain >,<,==).
//
// Field quick-ref used below:
//   li         = surface         (SurfaceSprite / ba)
//   Xb         = graphics        (java.awt.Graphics)
//   originX,originY       = screen offset x,y
//   tg,dg      = panel column offsets (tg = right panel x-base, dg = left panel x-base)
//   Oi         = inventoryPanelH  (inventory panel height / bottom border y)
//   jk         = compassAngle    (0..3071, compass rotation counter)
//   Xd         = activePanel     (0=none,1=options,2=quest/skill,3=...)
//   qc         = inventoryTab    (0-6 inventory sub-tab index)
//   mouseX          = mouseX          (inherited from GameShell)
//   mouseY         = mouseY
//   Cf         = mouseClickButton (0=none,1=left,2=right)
//   Bh         = selectedItem    (selected inventory item index, -1=none)
//   af         = selectedSpell   (selected spell index, -1=none)
//   Lf         = currentFloor
//   cl         = inventorySize   (inventory capacity, e.g. 30)
//   lc         = inventoryCount
//   vf         = inventoryItems  (int[] item ids)
//   xe         = inventoryQty    (int[] stack counts)
//   Aj         = inventoryEquipped (int[] equip flags)
//   eg         = mouseButtonMode (1=one-button, 77=two-button)
//   Jh         = clientStream    (ClientStream/da); Jh.f = outbound Buffer
//   Wf         = scrollMessageList (MessageList/wb) cleared by drawScrollList
//   zh         = friendsList     (MessageList/wb), used as menu builder
//   He         = chatList        (MessageList/wb)
//   ge         = panelGame       (Panel/qa)
//   yi         = panelDuel       (Panel/qa)  [reused: fatigue + server msgs]
//   Af         = panelQuest      (Panel/qa)  [reused as char-design widget container]
//   Qi,td      = panelDuel control ids (Qi=fatigue/title slot, td=server-msg slot)
//   Wd         = fatigueBarWidth
//   Ek         = world           (World/k)
//   Hh         = scene           (Scene/lb)
//   wi         = wi     (GameCharacter/ta)
//   Kh,Yh,ne   = privacy: chatPrivateOn, tradePrivateOn, membersPrivateOn
//   Pg         = isMembersAccount
//   Kd         = isMembersWorld
//   De         = autoRetaliateOn (toggled+sent in opcode 64)
//   Yd         = combatModeSetting (3-state display 0/1/2)
//   dc         = mouseButtonsOne  (0=two,1=one) game option
//   Vg         = cameraModeAuto   (0=manual,1=auto) game option
//   ui         = membersOption    (0/1) game option
//   Bd         = showMenuBorder
//   inputTextFinal         = inputLine        (current text entry buffer)
//   e          = tempInputString  (scratch label for dialogs/menus)
//   ec         = reportAbuseTarget
//   Yb         = reportAbusePage
//   Vf         = inputMode (0=none,2=report-offence-picker,...)
//   Oj         = reportAbuseOffence
//   Ce         = reportAbuseMuteFlag
//   ue         = reportAbuseMuteConfirmed
//   Bj         = socialDialogMode (1=addFriend,2=sendPM,3=addIgnore)
//   Qd         = pmTarget
//   x          = pmInput
//   inputPmFinal         = submittedPmInput
//   Wk         = helpMenuOpen     (controls close-window panel height)
//   mh         = showCloseWindow  (modal close-window dialog visible)
//   Cj         = helpMenuTitle
//   qd         = closeButtonSpriteId
//   pj         = closeButtonPressed
//   od,zi,gl,gc,vk = scroll-list: options, width, height, x, openFlag
//   vj,fj      = drawListCount, drawListSize
//   ae,ci      = drawListIds, drawListCurrent
//   di,Xe      = drawListY, drawListYShadow
//   Gi         = drawListCapacity
//   Se,ye,vc   = wall-model coordinate arrays (z, x, type) ; hg = wallModels
//   kh         = objectModels (GameModel[]) ; sg = spriteBaseSlot
//   ei,Dg,Wh   = char-design colour palettes (int[]) ; Lh,Wg,hh,ld = palette indices ;
//                wg,dk,Vd = equip-sprite slot bases (obf names kept — see drawCharDesignControls)
//   Sf         = charDesignGender
//   o          = ISAAC (o.a = colour/scratch int helper, o.g = static String[] menu list)
//   u          = StringCodec (u.g = scratch)
//   w          = WorldEntity (w.a = name normalize/format, w.g = equip-sprite slot table)
//   STRINGS    = il[] decoded string pool

// ---------------------------------------------------------------------------

    /**
     * Render the minimap/compass panel (and fatigue bar on the stats tab).
     * Only draws the rotating 3-segment compass ring when a side panel is open
     * (activePanel in {0,1,2,3}); otherwise the panel area is left to the Ek view.
     * obf: void k(int)
     */
    private void drawMinimap(int param) {
        li.interlace = false;        // clear sprite-clip flag (obf: li.i)
        Dc = false;               // clear dirty flag
        li.blackScreen(true);          // flush surface buffer

        // Compass ring only rotates while a side panel is showing.
        // obf: if (~Xd==-1 || ~Xd==-2 || Xd==2 || ~Xd==-4)  →  Xd in {0,1,2,3}
        if (Xd == 0 || Xd == 1 || Xd == 2 || Xd == 3) {
            // compassAngle (jk) runs 0..3071; doubled+wrapped into a [0,3072) phase.
            int compassPos = 2 * jk % 3072;
            // The ring is drawn in three 1024-wide slices; which slice we're in
            // selects one solid strip + an optional partial seam.
            if (compassPos < 1024) {
                li.drawSprite(-1, dg, 10, 0);                       // first strip
                if (compassPos > 768) {                         // seam into next slice
                    li.drawSpriteAlpha(1 + dg, 0, 0, compassPos - 768, 10);
                }
            } else if (compassPos < 2048) {
                li.drawSprite(-1, 1 + dg, 10, 0);                   // second strip
                if (compassPos > 1792) {
                    li.drawSpriteAlpha(tg - -10, 0, 0, compassPos - 1792, 10);
                }
            } else {
                li.drawSprite(-1, tg - -10, 10, 0);                 // third strip
                if (compassPos > 2816) {
                    li.drawSpriteAlpha(dg, 0, 0, compassPos - 2816, 10);
                }
            }
        }

        // Special token 2540: keep the ground-item overlay; otherwise clear it.
        if (param != 2540) {
            inventoryGroundOverlay = null;   // obf: of
        }

        // No active panel → tick the game panel (resets quest-list scroll).
        if (Xd == 0) {
            ge.render((byte)-63);          // obf: ge.a
        }

        // Stats/skills tab open → draw the fatigue bar.
        if (Xd == 2) {
            String fatStr = yi.getFieldText(fatigueControlId, 4);   // obf: yi.g(Qi,4)
            if (fatStr != null && fatStr.length() > 0) {
                li.drawBoxAlpha(100, 0, 30, 0, 185, screenWidth, 0);   // obf: li.c(...,Wd,0)
            }
            yi.render((byte)-52);          // obf: yi.a — tick panel
        }

        li.drawSprite(-1, tg + 22, screenHeight, 0);   // bottom border (obf: Oi)
        li.draw(graphics, originX, 256, originY);              // blit panel to AWT
    }

// ---------------------------------------------------------------------------

    /**
     * Inventory-area sub-tab hover/leave tracking (no drawing here — it only mutates
     * the active sub-tab index qc based on the cursor position over the right-side
     * tab strip). The tab strip lives at the right edge: x ∈ [li.u-200, li.u-3].
     *
     * All conditions below are the de-obfuscated forms of the clean ~-idioms, e.g.
     *   ~(li.u-35) >= ~mouseX        →  mouseX >= li.u-35
     *   ~mouseY <= -4               →  mouseY >= 3
     *   ~mouseX > ~(li.u-3)          →  mouseX < li.u-3
     * obf: void D(int)
     */
    private void drawInventoryTab(int param) {
        // --- enter a sub-tab from the closed state (qc==0) or the row-1 state (qc==1) ---
        if (qc == 0 && mouseX >= li.width - 35 && mouseY >= 3 && mouseX < li.width - 3 && mouseY < 35) {
            qc = 1;
        }
        if (qc == 0 && mouseX >= li.width - 68 && mouseY >= 3 && mouseX < li.width - 36 && mouseY < 35) {   // RENDER-BUG FIX (clean client.java:5314 ~qc==-1): qc==1 -> qc==0
            qc = 2;
            charDesignWobbleX = (int)(13.0 * Math.random()) - 6;    // obf: Df
            charDesignWobbleY = (int)(Math.random() * 23.0) - 11;   // obf: sd
        }
        if (qc == 0 && mouseX >= li.width - 101 && mouseY >= 3 && mouseX < li.width - 69 && mouseY < 35) {   // RENDER-BUG FIX (clean client.java:5320 ~qc==-1): qc==1 -> qc==0
            qc = 3;
        }
        if (qc == 0 && mouseX >= li.width - 134 && mouseY >= 3 && mouseX < li.width - 102 && mouseY < 35) {
            qc = 4;
        }
        if (qc == 0 && mouseX >= li.width - 167 && mouseY >= 3 && mouseX < li.width - 135 && mouseY < 35) {   // RENDER-BUG FIX (clean client.java:5328 ~qc==-1): qc==1 -> qc==0
            qc = 5;
        }
        if (param != 1) {
            Lf = -32;   // obf: Lf  (dummy reset when not called with sentinel 1)
        }
        if (qc == 0 && mouseX >= li.width - 200 && mouseY >= 3 && mouseX < li.width - 168 && mouseY < 35) {   // RENDER-BUG FIX (clean client.java:5336 ~qc==-1): qc==1 -> qc==0
            qc = 6;
        }
        // --- re-select from any open sub-tab when cursor is over the narrower (26px) header ---
        if (qc != 0 && mouseX >= li.width - 35 && mouseY >= 3 && mouseX < li.width - 3 && mouseY < 26) {
            qc = 1;
        }
        if (qc != 0 && qc != 2 && mouseX >= li.width - 68 && mouseY >= 3 && mouseX < li.width - 36 && mouseY < 26) {
            qc = 2;
            charDesignWobbleY = -11 + (int)(23.0 * Math.random());
            charDesignWobbleX = -6 + (int)(13.0 * Math.random());
        }
        if (qc != 0 && mouseX >= li.width - 101 && mouseY >= 3 && mouseX < li.width - 69 && mouseY < 26) {
            qc = 3;
        }
        if (qc != 0 && mouseX >= li.width - 134 && mouseY >= 3 && mouseX < li.width - 102 && mouseY < 26) {
            qc = 4;
        }
        if (qc != 0 && mouseX >= li.width - 167 && mouseY >= 3 && mouseX < li.width - 135 && mouseY < 26) {
            qc = 5;
        }
        if (qc != 0 && mouseX >= li.width - 200 && mouseY >= 3 && mouseX < li.width - 168 && mouseY < 26) {
            qc = 6;
        }
        // --- leave a sub-tab when the cursor drops out of its body region (→ qc=0) ---
        // Inventory grid (qc==1): below the item rows.
        if (qc == 1 && (mouseX < li.width - 248 || mouseY > 36 + 34 * (cl / 5))) {
            qc = 0;
        }
        // Stats body (qc==3): obf condition ~qc==-4 → qc==3.
        if (qc == 3 && (mouseX < li.width - 199 || mouseY > 316)) {
            qc = 0;
        }
        // Quest / friends / ignore bodies (qc==2 || qc==4 || qc==5).
        if ((qc == 2 || qc == 4 || qc == 5) && (mouseX < li.width - 199 || mouseY > 240)) {
            qc = 0;
        }
        // Settings body (qc==6).
        if (qc == 6 && (mouseX < li.width - 199 || mouseY > 311)) {
            qc = 0;
        }
    }

// ---------------------------------------------------------------------------

    /**
     * Game settings / privacy "Configuration" tab.
     * Renders the privacy toggles (private chat / trade / members) and the option
     * toggles (auto-retaliate / mouse buttons / camera / members-only), plus the
     * logout link(s). When processClicks is set, hit-tests the rows and, on a LEFT
     * click (Cf==1 for every row), flips the relevant flag and emits the matching
     * packet (opcode 111 GAME_SETTINGS_CHANGED for the privacy rows, opcode 64
     * PRIVACY_SETTINGS_CHANGED for the option rows).
     * obf: void b(int,boolean)
     */
    // Moved to GameInterface.drawGameSettings — delegates here.
    private void drawGameSettings(int param, boolean processClicks) {
        gameInterface.drawGameSettings(param, processClicks);
    }

// ---------------------------------------------------------------------------

    /**
     * Toggle to two-button mouse mode (eg=77). Called with a sentinel param so the
     * anti-tamper guard (if param != -16433) gates the assignment.
     * obf: void g(int)
     */
    private void setMouseButtonMode(int param) {
        if (param != -16433) {
            mouseButtonMode = 77;   // obf: eg
        }
    }

// ---------------------------------------------------------------------------

    /**
     * Reset the trade/duel/panel transient state. The secondary inventory-tab reset
     * only fires for the sentinel param (-2).
     * obf: void o(int)
     */
    private void resetTradeDuelState(int param) {
        tradeQueuedAction = 0;    // obf: bj
        Xd = 0;          // obf: Xd
        chatInputMode = 0;        // obf: qg
        if (param == -2) {
            qc = 0;     // obf: kc
        }
    }

// ---------------------------------------------------------------------------

    /**
     * Character-design arrow/colour-wheel control row renderer (hover highlight only).
     * Three categories per pass (body-part sprite + two colour wheels), drawn at three
     * column anchors (x-87, x-32, x+23 from the centred base x=256).
     * obf: void w(int)
     */
    private void drawCharDesignControls(int param) {
        li.interlace = false;
        li.blackScreen(true);
        Af.render((byte)-13);    // obf: Af.a — update panel hover state

        int baseX = 140 + 116;   // = 256 (centred column)
        int baseY = 50 - 25;     // = 25

        // Per-category indices into the colour palettes (ei/Wh/Dg) and the equip-sprite
        // slot table (WorldEntity.g): Lh,Wg = skin/colour ; hh = colour ; ld = colour ;
        // wg,dk,Vd = sprite-slot bases. (Obf field names kept — semantic split is approximate.)
        // Column 1 (baseX-87)
        li.spriteClippingTinted(baseX - 87, ei[Lh], WorldEntity.spriteOffsets[wg], baseY, 102, (byte)105, 64);
        li.spriteClipping(baseY, ei[Wg], Wh[hh], false, 0, WorldEntity.spriteOffsets[dk], 102, 64, baseX - 87, 1);
        li.spriteClipping(baseY, Dg[ld], Wh[hh], false, 0, WorldEntity.spriteOffsets[Vd], 102, 64, baseX - 87, param + 13760);
        // Column 2 (baseX-32)
        li.spriteClippingTinted(baseX - 32, ei[Lh], 6 + WorldEntity.spriteOffsets[wg], baseY, 102, (byte)105, 64);
        li.spriteClipping(baseY, ei[Wg], Wh[hh], false, 0, WorldEntity.spriteOffsets[dk] + 6, 102, 64, baseX - 32, 1);
        li.spriteClipping(baseY, Dg[ld], Wh[hh], false, 0, 6 + WorldEntity.spriteOffsets[Vd], 102, 64, baseX - 32, 1);
        // Column 3 (baseX+23)
        li.spriteClippingTinted(baseX - 32 + 55, ei[Lh], 12 + WorldEntity.spriteOffsets[wg], baseY, 102, (byte)110, 64);
        li.spriteClipping(baseY, ei[Wg], Wh[hh], false, 0, WorldEntity.spriteOffsets[dk] + 12, 102, 64, baseX - 32 + 55, param + 13760);
        li.spriteClipping(baseY, Dg[ld], Wh[hh], false, 0, WorldEntity.spriteOffsets[Vd] + 12, 102, 64, baseX - 32 + 55, 1);

        li.drawSprite(-1, tg + 22, screenHeight, 0);   // bottom border (obf: Oi)
        li.draw(graphics, originX, 256, originY);
        if (param != -13759) {
            drawCloseButton((byte)70);   // obf: l(70)
        }
    }

// ---------------------------------------------------------------------------

    /**
     * Build the "Please design your Character" appearance screen widgets into the
     * char-design panel (Af). Lays out the head/hair toggles, gender swatches and the
     * five colour-wheel arrow pairs (top/bottom/hair/skin), then the Accept button.
     * The recorded button ids (Dj/pi/Kj/ed/Ge/Of/Xc/ek/Ze/Mj/Re/Ai/Eg) drive
     * drawCharDesignControls + sendAppearance.
     * obf: void t(int)
     */
    private void drawCharDesign(int param) {
        Af = new Panel(li, 100);   // obf: Af = new qa(li, 100)
        Af.addLabel(true, (byte)-125, 4, 256, STRINGS[87], 10);   // title

        int x = 140 + 116;   // = 256
        int y = 34 - 10;     // = 24

        // Head-style toggle row (left / current / right)
        Af.addLabel(true, (byte)-104, 3, x - 55, STRINGS[82], y + 110);
        Af.addLabel(true, (byte)-91, 3, x, STRINGS[92], y + 110);
        Af.addLabel(true, (byte)-117, 3, x + 55, STRINGS[81], y + 110);
        y += 145;

        int s = 54;
        // Gender swatch (left)
        Af.addCentredSprite(41, x - s, 53, 26531, y);
        Af.addLabel(true, (byte)-81, 1, x - s, STRINGS[84], y - 8);
        Af.addLabel(true, (byte)-125, 1, x - s, STRINGS[88], y + 8);
        Af.addNativeSprite(StringCodec.STATUS_NOT_FOUND - -7, y, x - s - 40, -114);
        Dj = Af.addProgressWidget(x - s - 40, 20, y, param + 24525, 20);
        Af.addNativeSprite(6 + StringCodec.STATUS_NOT_FOUND, y, x - s + 40, -59);
        pi = Af.addProgressWidget(x - s + 40, 20, y, param ^ 24649, 20);
        // Gender swatch (right)
        Af.addCentredSprite(41, x - -s, 53, 26531, y);
        Af.addLabel(true, (byte)-85, 1, x - -s, STRINGS[85], y - 8);
        Af.addLabel(true, (byte)-102, 1, s + x, STRINGS[86], y + 8);
        Af.addNativeSprite(7 + StringCodec.STATUS_NOT_FOUND, y, s + (x - 40), -57);
        Kj = Af.addProgressWidget(x - -s - 40, 20, y, 64, 20);
        Af.addNativeSprite(6 + StringCodec.STATUS_NOT_FOUND, y, 40 + s + x, -127);
        ed = Af.addProgressWidget(40 + s + x, 20, y, param ^ -24650, 20);
        y += 50;

        // Hair colour (left swatch) + Top colour pair (right swatch)
        Af.addCentredSprite(41, x - s, 53, 26531, y);
        Af.addLabel(true, (byte)-102, 1, x - s, STRINGS[91], y);
        Af.addNativeSprite(StringCodec.STATUS_NOT_FOUND - -7, y, -40 + x - s, param + 24525);
        Ge = Af.addProgressWidget(x - s - 40, 20, y, -81, 20);
        Af.addNativeSprite(StringCodec.STATUS_NOT_FOUND - -6, y, 40 - s + x, param + 24521);
        Of = Af.addProgressWidget(40 - s + x, 20, y, 54, 20);
        Af.addCentredSprite(41, s + x, 53, param ^ -1970, y);
        Af.addLabel(true, (byte)-102, 1, s + x, STRINGS[79], y - 8);
        Af.addLabel(true, (byte)-79, 1, s + x, STRINGS[86], y + 8);
        Af.addNativeSprite(7 + StringCodec.STATUS_NOT_FOUND, y, s + x - 40, -104);
        Xc = Af.addProgressWidget(s + x - 40, 20, y, param + 24504, 20);
        Af.addNativeSprite(6 + StringCodec.STATUS_NOT_FOUND, y, 40 + s + x, -105);
        ek = Af.addProgressWidget(x - (-s - 40), 20, y, -91, 20);
        y += 50;
        if (param != -24595) {
            drawProgressBar(-127);   // obf: y(-127)
        }

        // Top colour pair (left swatch) + Bottom colour pair (right swatch)
        Af.addCentredSprite(41, x - s, 53, param ^ -1970, y);
        Af.addLabel(true, (byte)-81, 1, x - s, STRINGS[83], y - 8);
        Af.addLabel(true, (byte)-109, 1, x - s, STRINGS[86], y + 8);
        Af.addNativeSprite(7 + StringCodec.STATUS_NOT_FOUND, y, -40 + x - s, -59);
        Ze = Af.addProgressWidget(-40 + x - s, 20, y, param + 24468, 20);
        Af.addNativeSprite(StringCodec.STATUS_NOT_FOUND + 6, y, x - s - -40, -95);
        Mj = Af.addProgressWidget(x - s + 40, 20, y, param + 24637, 20);
        Af.addCentredSprite(41, s + x, 53, 26531, y);
        Af.addLabel(true, (byte)-108, 1, s + x, STRINGS[89], y - 8);
        Af.addLabel(true, (byte)-108, 1, s + x, STRINGS[86], y + 8);
        Af.addNativeSprite(StringCodec.STATUS_NOT_FOUND + 7, y, -40 + s + x, -90);
        Re = Af.addProgressWidget(s + x - 40, 20, y, 69, 20);
        Af.addNativeSprite(6 + StringCodec.STATUS_NOT_FOUND, y, x + s + 40, param + 24537);
        Ai = Af.addProgressWidget(40 + s + x, 20, y, -119, 20);
        y += 82;

        // Accept button
        y -= 35;
        Af.addOval(param ^ 24661, 200, 30, x, y);
        Af.addLabel(false, (byte)-74, 4, x, STRINGS[90], y);
        Eg = Af.addProgressWidget(x, 200, y, param ^ -24631, 30);   // Accept → opcode 235
    }

// ---------------------------------------------------------------------------

    /**
     * Overlay a centred two-line message box directly onto the AWT Graphics (used for
     * transient "saving/loading" notices that must paint outside the normal li
     * blit). Header text below, body text above, inside a black 280x50 box.
     * The colorCode sentinel (-64) is an anti-tamper guard.
     * obf: void a(String,byte,String)
     */
    private void addChatMessage(String header, byte colorCode, String body) {
        if (colorCode != -64) {
            return;
        }
        Graphics g = this.getGraphics();
        if (g == null) {
            return;
        }
        g.translate(originX, originY);
        Font font = new Font(STRINGS[477], 1, 15);   // STRINGS[477] = "Helvetica"
        int w = 512;
        g.setColor(Color.black);
        int h = 344;
        g.fillRect(w / 2 - 140, h / 2 - 25, 280, 50);
        g.setColor(Color.white);
        g.drawRect(w / 2 - 140, h / 2 - 25, 280, 50);
        this.drawCenteredString(font, body, h / 2 - 10, true, w / 2, g);
        this.drawCenteredString(font, header, h / 2 + 10, true, w / 2, g);
    }

// ---------------------------------------------------------------------------

    /**
     * Display a server/system message on the duel/stats panel (yi). When the body is
     * empty it goes to a single slot (td); otherwise the title goes to slot Qi and the
     * body to slot td. A trigger code < -11 forces a minimap redraw + action flush.
     * obf: void b(byte,String,String)
     */
    private void showServerMessage(byte triggerCode, String title, String body) {
        if (Xd == 2) {
            if (body == null || body.length() < 1) {
                yi.setFieldText(serverMsgControlId, title, 27642);   // obf: yi.a(td, ...)
            } else {
                yi.setFieldText(fatigueControlId, title, 27642);     // obf: yi.a(Qi, ...)
                yi.setFieldText(serverMsgControlId, body, 27642);    // obf: yi.a(td, ...)
            }
        }
        if (triggerCode < -11) {
            drawMinimap(2540);            // obf: k(2540)
            sendQueuedActions((byte) -28492);    // obf: c(-28492) — target is c(byte); dummy guard arg, narrowed to byte
        }
    }

// ---------------------------------------------------------------------------

    /**
     * Render the right-click "Choose option" list. Clears the screen first unless
     * called with sentinel 12, then delegates to drawScrollList with a 9px left margin.
     * obf: void a(String[],int,int,boolean)
     */
    // Moved to WidgetRenderer.drawMenuOptions — delegates here.
    void drawMenuOptions(String[] options, int x, int y, boolean rightClick) {
        widgetRenderer.drawMenuOptions(options, x, y, rightClick);
    }

// ---------------------------------------------------------------------------

    // Moved to WidgetRenderer.drawScrollList — delegates here.
    private void drawScrollList(int x, int y, String[] options, boolean showBorder, String title) {
        widgetRenderer.drawScrollList(x, y, options, showBorder, title);
    }

// ---------------------------------------------------------------------------

    // Moved to WidgetRenderer.drawScrollbar — delegates here.
    void drawScrollbar(byte sentinel, int x, int y, int scrollPos, boolean animate, int trackLen) {
        widgetRenderer.drawScrollbar(sentinel, x, y, scrollPos, animate, trackLen);
    }

// ---------------------------------------------------------------------------

    // Moved to WidgetRenderer.drawScrollbar2 — delegates here.
    void drawScrollbar2(int x, int y, int w, int h, boolean animate, int trackLen) {
        widgetRenderer.drawScrollbar2(x, y, w, h, animate, trackLen);
    }

// ---------------------------------------------------------------------------

    /**
     * Rebuild the inventory item draw-cache from the previous frame's draw list and the
     * current inventory, de-duplicating item ids and appending new ones up to the cache
     * capacity (Gi). (The trailing junk divide in the clean output is an opaque-predicate
     * artifact and is dropped.)
     * obf: void C(int)
     */
    // Moved to GameInterface.drawHelpMenu — delegates here.
    void drawHelpMenu(int param) {
        gameInterface.drawHelpMenu(param);
    }

// ---------------------------------------------------------------------------

    /**
     * Draw the "Click here to close window" modal button (and the help panel frame).
     * Highlights the close text on hover; on a LEFT click (Cf==1) over the text, or a
     * LEFT click fully outside the modal box, closes it (mh=false).
     * obf: void l(byte)
     */
    // Moved to GameInterface.drawCloseButton — delegates here.
    private void drawCloseButton(byte param) {
        gameInterface.drawCloseButton(param);
    }

// ---------------------------------------------------------------------------

    /**
     * Substitute a named wall/boundary GameModel into the Hh at wall-slot index
     * slotIndex, provided the slot's tile (Se/ye) is in-bounds and within 7 tiles of the
     * local player. Used to swap in special boundary geometry (e.g. doors). The sentinel
     * guard (var1 > 2) is anti-tamper.
     * obf: void a(byte,int,String)
     */
    private void drawTextField(byte sentinel, int slotIndex, String text) {
        int wallZ = wallModelZ[slotIndex];   // obf: Se[n2]
        int wallX = wallModelX[slotIndex];   // obf: ye[n2]
        int relX = wallZ - wi.currentX / 128;
        int relY = -(wi.currentY / 128) + wallX;
        int range = 7;

        if (sentinel <= 2) return;
        if (wallZ < 0) return;
        if (wallX < 0) return;
        if (wallZ > 95) return;
        if (wallX > 95) return;
        if (relX <= -range) return;
        if (relX >= range) return;
        if (relY <= -range) return;
        if (relY >= range) return;

        Ek.removeModel(hg[slotIndex]);             // remove existing wall model (obf: Ek.a)
        int modelIdx = GameModel.textureId((byte)91, text);     // resolve model by name
        GameModel newModel = objectModels[modelIdx].copy(-2);   // clone base model
        Ek.addModel(newModel);                   // register into scene
        newModel.setLight(-50, 48, -10, -50, true, 48, -74);   // transform/scale
        newModel.copyPosition(hg[slotIndex], 6029);        // copy placement from old model
        newModel.key = slotIndex;
        hg[slotIndex] = newModel;
    }

// ---------------------------------------------------------------------------

    /**
     * Add-friend / send-PM / add-ignore entry dialog renderer + submit handler.
     * socialDialogMode (Bj): 1=Add friend, 2=Send PM, 3=Add ignore. A click outside the
     * active dialog box (or on the Cancel link) dismisses it. Each mode renders its box
     * and, once the input line is committed, normalises the name (WorldEntity.a) and
     * fires the matching social packet — but never when the target equals the local
     * player's own name.
     * obf: void h(byte)
     */
    private void drawSocialDialog(byte sentinel) {
        if (Cf != 0) {
            Cf = 0;
            // Add-friend box: 106..406 x, 145..215 y
            if (Bj == 1 && (mouseX < 106 || mouseY < 145 || mouseX > 406 || mouseY > 215)) {
                Bj = 0;
                return;
            }
            // Send-PM box: 6..506 x, 145..215 y
            if (Bj == 2 && (mouseX < 6 || mouseY < 145 || mouseX > 506 || mouseY > 215)) {
                Bj = 0;
                return;
            }
            // Add-ignore box: 106..406 x, 145..215 y
            if (Bj == 3 && (mouseX < 106 || mouseY < 145 || mouseX > 406 || mouseY > 215)) {
                Bj = 0;
                return;
            }
            // Cancel link region
            if (mouseX > 236 && mouseX < 276 && mouseY > 193 && mouseY < 213) {
                Bj = 0;
                return;
            }
        }

        int y = 145;

        // Mode 1: Add friend
        if (Bj == 1) {
            li.drawBox(106, (byte)26, 0, y, 70, 300);
            li.drawBoxEdge(106, 300, y, 27785, 70, 0xFFFFFF);
            y += 20;
            li.drawstringRight(256, STRINGS[246], 0xFFFFFF, 0, 4, y);   // "Enter name to add to friends list"
            y += 20;
            li.drawstringRight(256, tempInputString + "*", 0xFFFFFF, 0, 4, y);
            String self = WorldEntity.trimAndValidateString(wi.message, (byte)50);   // normalise local name
            if (self != null && inputLine.length() > 0) {
                String typed = inputLine.trim();
                Bj = 0;
                tempInputString = "";
                inputLine = "";
                if (typed.length() > 0 && !self.equals(WorldEntity.trimAndValidateString(typed, (byte)100))) {
                    packets.sendAddFriend(114, typed);   // obf: b(114, typed) — opcode 195
                }
            }
        }

        // Mode 2: Send private message
        if (Bj == 2) {
            li.drawBox(6, (byte)110, 0, y, 70, 500);
            li.drawBoxEdge(6, 500, y, 27785, 70, 0xFFFFFF);
            y += 20;
            li.drawstringRight(256, STRINGS[249] + Qd, 0xFFFFFF, 0, 4, y);   // "Sending message to "
            y += 20;
            li.drawstringRight(256, pmInput + "*", 0xFFFFFF, 0, 4, y);
            if (submittedPmInput.length() > 0) {
                String msg = submittedPmInput;
                pmInput = "";
                Bj = 0;
                submittedPmInput = "";
                packets.sendPrivateMessage((byte)-76, Qd, msg);   // obf: a((byte)-76, ...) opcode 218
            }
        }

        // Mode 3: Add ignore
        if (Bj == 3) {
            li.drawBox(106, (byte)-115, 0, y, 70, 300);
            li.drawBoxEdge(106, 300, y, 27785, 70, 0xFFFFFF);
            y += 20;
            li.drawstringRight(256, STRINGS[248], 0xFFFFFF, 0, 4, y);   // "Enter name to add to ignore list"
            y += 20;
            li.drawstringRight(256, tempInputString + "*", 0xFFFFFF, 0, 4, y);
            String self = WorldEntity.trimAndValidateString(wi.message, (byte)59);
            if (self != null && inputLine.length() > 0) {
                String typed = inputLine.trim();
                tempInputString = "";
                Bj = 0;
                inputLine = "";
                if (typed.length() > 0 && !self.equals(WorldEntity.trimAndValidateString(typed, (byte)105))) {
                    packets.sendAddIgnore(typed, (byte)5);   // obf: a(typed, (byte)5) — opcode 132
                }
            }
        }

        // Cancel link (always drawn)
        int color = 0xFFFFFF;
        if (mouseX > 236 && mouseX < 276 && mouseY > 193 && mouseY < 213) {
            color = 0xFFFF00;
        }
        li.drawstringRight(256, STRINGS[121], color, 0, 1, 208);   // STRINGS[121] = "Cancel"

        if (sentinel <= 77) {
            pj = -42;   // obf: pj
        }
    }

// ---------------------------------------------------------------------------

    /**
     * "Enter the name of the player you wish to report" entry screen (report-abuse step 1).
     * If a name has been typed, accept it and advance to the offence picker (Vf=2).
     * Otherwise render the entry box (plus an optional mute toggle), the Submit link
     * (commits tempInputString into the input line) and the Cancel link. A LEFT click
     * outside the box cancels.
     *
     * Mute tier (var2): (Ce>=2 || Oj>=7) → 2 ; else Oj<5 → 0 ; else (Oj 5/6) → 1.
     * obf: void d(boolean)
     */
    private void drawReportNameEntry(boolean rightAlign) {
        if (inputLine.length() > 0) {
            ec = inputLine.trim();   // obf: ec
            Yb = 0;                    // obf: Yb
            Vf = 2;                          // obf: Vf
            return;
        }

        // mute-option tier
        int muteTier;
        if (Ce >= 2 || Oj >= 7) {
            muteTier = 2;
        } else if (Oj < 5) {        // obf: ~Oj > -6
            muteTier = 0;
        } else {
            muteTier = 1;
        }

        int fontH = li.textHeight(508305352, 1);
        int lineH = li.textHeight(508305352, 4);
        int panelW = 400;
        int panelH = (muteTier > 0 ? 5 + fontH : 0) + 70;
        int panelX = 256 - panelW / 2;
        int panelY = 180 - panelH / 2;

        li.drawBox(panelX, (byte)88, 0, panelY, panelH, panelW);
        li.drawBoxEdge(panelX, panelW, panelY, 27785, panelH, 0xFFFFFF);
        li.drawstringRight(256, STRINGS[340], 0xFFFF00, 0, 1, 5 + panelY + fontH);   // title prompt

        int inputPad = fontH + 2;   // obf: var9
        li.drawstringRight(256, tempInputString + "*", 0xFFFFFF, 0, 4, lineH + (panelY + 5) + (inputPad + 3));

        int nextY = fontH + lineH + (8 + panelY + inputPad + 2);   // obf: var10
        int color = 0xFFFFFF;

        // mute toggle row (only when tier > 0)
        if (muteTier > 0) {
            String muteLabel = ue ? STRINGS[336] : STRINGS[339];
            if (muteTier > 1) {
                muteLabel = muteLabel + STRINGS[341];
            }
            muteLabel = muteLabel + STRINGS[337];

            int muteW = li.textWidth(1, 72, muteLabel);
            if (mouseX > 256 - muteW / 2 && mouseX < 256 + muteW / 2 && mouseY > nextY - fontH && mouseY < nextY) {
                if (Cf != 0) {
                    ue = !ue;
                    Cf = 0;
                }
                color = 0xFFFF00;
            }
            li.drawstringRight(256, muteLabel, color, 0, 1, nextY);
            nextY += 10 + fontH;
        }

        // Submit link (obf: mouseX > 210 && mouseX < 228) — commits the typed name into the input line
        color = 0xFFFFFF;
        if (mouseX > 210 && mouseX < 228 && mouseY > nextY - fontH && mouseY < nextY) {
            if (Cf != 0) {
                inputLine = tempInputString;   // obf: inputTextFinal = e
                Cf = 0;
            }
            color = 0xFFFF00;
        }
        li.drawstring(STRINGS[122], 210, nextY, color, rightAlign, 1);   // "Submit"

        // Cancel link (obf: mouseX > 264 && mouseX < 304)
        color = 0xFFFFFF;
        if (mouseX > 264 && mouseX < 304 && mouseY > nextY - fontH && mouseY < nextY) {
            color = 0xFFFF00;
            if (Cf != 0) {
                Cf = 0;
                Vf = 0;   // obf: Vf = 0
            }
        }
        li.drawstring(STRINGS[121], 264, nextY, color, rightAlign, 1);   // "Cancel"

        // LEFT click fully outside the box → cancel
        if (Cf == 1) {   // obf: ~Cf == -2 → Cf == 1
            if (mouseX < panelX || mouseX > panelX + panelW || mouseY < panelY || mouseY > panelY + panelH) {
                Vf = 0;
                Cf = 0;
            }
        }
    }


    // =========================================================================
    // ===== input =====
    // =========================================================================
// Methods: handleGameClick, buildClickMenu, handleInventoryClick,
//          menuHitTest (getInventoryCount), pointInRect (isItemEquipped),
//          pointInPanel (isEquipSlotActive), pollInput
//
// Field naming follows MUDCLIENT_SKELETON.md; class names follow NAMING.md.
// Obfuscation strips applied to every method:
//   - `boolean bl = client.vh;` opaque-predicate removed (always false)
//   - `++<counter>;` profiling counter calls removed
//   - `if (bl) return;` / `while (!bl)` dead branches removed
//   - `try { BODY } catch (RuntimeException e) { throw ErrorHandler.wrap(e,"sig"); }` unwrapped
//   - anti-tamper `if (param != CONST) this.someOtherCall(junkArgs);` guards removed
//   - junk bitwise masks before arithmetic (`~x != ~y` → `x != y`, etc.) simplified
//
// Field notes (named here by behaviour, cross-checked vs oracle mudclient204):
//   Kk  (int[8192])  clickHistoryX    – ring buffer of recent click X values
//   uj  (int[8192])  clickHistoryY    – ring buffer of recent click Y values
//                                       (oracle: mouseClickXHistory/mouseClickYHistory)
//   nk  (int)        clickRingHead    – ring buffer write index (masked 0x1FFF)
//   oh  (int[])      equipBonusDisplay – equipment stat array (cleared on top-bar click)
//   ai  (int)        combatTimeout    – >0 while in/after combat; logout blocked unless 0
//                                       (oracle: combatTimeout; set to 500 by idle-logout path)
//   bj  (int)        logoutTimeout    – set to 1000 after requestLogout(0) fires
//   af  (int)        selectedSpell    – selected spell index (-1 = none)
//   Bh  (int)        selectedItemInventoryIndex – selected inventory slot (-1 = none)
//   ig  (String)     selectedItemName
//   Pg  (boolean)    isMember
//   rg  (ta[500])    playersLast      – player entities from previous tick
//   zg,sh,sk,Qg,Lf,Ki (int) – camera / region offset fields (localRegionX/Y, planeW/H, regionX/Y)
//
// Cross-class game-data tables (scattered across obf classes by the obfuscator;
// names reflect oracle GameData.*, obf traces kept for accuracy — NOTE these
// CONTRADICT a naive "EntityDef/BitBuffer holds everything" reading):
//   ac.x   (String[]) itemName       – oracle GameData.itemName    (NAMING: ac=DecodeBuffer)
//   fa.e   (int[])    itemStackable  – oracle GameData.itemStackable (NAMING: fa=ClientIOException; static table reused)
//   mb.k   (int[])    itemWearable   – oracle GameData.itemWearable  (NAMING: mb=Utility)
//   qb.e   (int[])    spellType      – oracle GameData.spellType     (NAMING: qb=GameFrame; static table reused)
//   ja.L   (String[]) spellName      – oracle GameData.spellName     (NAMING: ja=BitBuffer)
//   lb.ac  (String[]) itemCommand    – oracle GameData.itemCommand   (NAMING: lb=Scene)
//   h.c    (int[])    itemPicture    – oracle GameData.itemPicture   (NAMING: h=TextEncoder)
//   ua.mouseButtonDown  (int[])    itemMask       – oracle GameData.itemMask      (NAMING: ua=Surface)
//   ta.originY   (int)      currentY       – GameCharacter currentY (pairs with waypointsY F[]); ta.i = currentX
//   ta.s   (int)      level          – GameCharacter combat level
//   ta.b   (int)      serverIndex    – GameCharacter serverIndex
//   ta.c   (String)   name ; ta.C    (String) displayName
//
// Key method cross-references (using skeleton proposed names):
//   requestLogout(0)  →  void B(int)  (opcode 102 LOGOUT)
//   resetChatInput(x) →  void o(byte)
//   resetPanels(x)    →  void p(byte)
//   drawSocialDialog  →  void h(byte)
//   friendsList       →  zh  (wb / MessageList, reused as shared right-click option-menu list)

    // =========================================================================
    // ===== menu / hit-test / anti-bot click subsystem =====
    // The following six methods were extracted to MenuController (m.menus.*):
    //   handleGameClick, buildClickMenu, handleInventoryClick,
    //   menuHitTest (getInventoryCount), pointInRect (isItemEquipped),
    //   pointInPanel (isEquipSlotActive).
    // pointInRect is private to MenuController (only pointInPanel calls it).
    // All internal call sites now route through the m.menus delegate.
    // =========================================================================


    // -----------------------------------------------------------------
    // pollInput  obf: private final void n(int)
    // -----------------------------------------------------------------

    /**
     * Per-tick window-size poll and layout reset.  Detects host-window resize
     * and rebuilds the panel layout.
     *
     * <p>Selects the AWT {@link java.awt.Component} to query for dimensions:</p>
     * <ul>
     *   <li>{@code hj} set (desktop) + socket open → {@code ClientStream.socket} ({@code da.ctrlDown})</li>
     *   <li>{@code hj} set but no socket → {@code this} (the Applet itself)</li>
     *   <li>Applet mode → {@code InputState.applet} ({@code kb.a})</li>
     * </ul>
     *
     * <p>Sets {@code Rh} (screenWidth), {@code Hf} (screenHeight), zeroes the Y
     * origin {@code originY}, recomputes centering offset {@code originX = (screenWidth -
     * gameWidth)/2}, then calls {@link #resetPanels}({@code 49}).</p>
     *
     * obf: private final void n(int param1)
     */
    private final void pollInput(int dummy) {
        // Select which AWT Component hosts the display surface.
        Object hostComponent;
        if (hj) {                          // hj = isDesktopMode
            if (ClientStream.applet != null) {           // obf: da.gb = ClientStream.applet
                hostComponent = ClientStream.applet;
            } else {
                hostComponent = this;      // standalone without active socket
            }
        } else {
            hostComponent = InputState.gameFrame;          // applet mode: InputState.applet
        }

        // Anti-tamper: if (param1 > -77) Ee = 30; — keep side effect (brief debounce timer).
        if (dummy > -77) {
            Ee = 30;
        }

        // Query display dimensions from the host component.
        Rh = ((java.awt.Component) hostComponent).getSize().width;  // screenWidth
        Hf = ((java.awt.Component) hostComponent).getSize().height; // screenHeight
        originY  = 0;                                                     // screenOriginY
        // Horizontal centering: (screenWidth - gameWidth) / 2.  obf: (-Wd + Rh) / 2
        originX = (-screenWidth + Rh) / 2;
        // Rebuild all Panel widgets for the current window size.
        resetPanels((byte) 49); // obf: this.p((byte)49)
    }


    // =========================================================================
    // ===== util =====
    // =========================================================================
// Methods in the "util" group of Mudclient (obfuscated class "client").
// 4-space indented as if inside the Mudclient class body.
// Re-audited against the CLEAN base (decompiled/normalized-clean/client.java).
// Obfuscation artefacts stripped: opaque predicate (boolean bl = client.vh),
// dead if(bl)/while(!bl)/break-label branches, ++<counter> profiling increments,
// try{BODY}catch(RuntimeException e){throw ErrorHandler.wrap(e,"sig")} wrappers,
// anti-tamper guards and junk dead divisions/modulos.
// All field / class names from MUDCLIENT_SKELETON.md + NAMING.md (k=World, lb=Scene).

    /**
     * Look up the display name for the given entity/item id from the ImageLoader cache;
     * spin-waits while the async loader populates the node, then returns the cached
     * String payload, else falls back to SurfaceSprite.e().
     * (Skeleton proposed name: formatNumber — actual implementation is a name cache lookup,
     *  not numeric formatting; Utility.formatNumber lives in mb.a(int,int).)
     * // obf: String c(int,int)  label: client.GC(
     */
    final String formatNumber(int flag, int entityId) {
        // obf: if (var1 >= -7) this.Si = 126;  — guard side-effect the DEFECTIVE base dropped.
        // When flag >= -7, prime the Si scratch/state field to 126 before the lookup.
        if (flag >= -7) {
            this.Si = 126;
        }

        // Look up ListNode for entityId via the LoaderThread (obf: pa.k.a(int,byte))
        ListNode node = ImageLoader.imageWidthCarrier.reverseDns(entityId, (byte)-121);

        // Spin-wait while the async loader hasn't populated the node yet
        // (node.status == 0  ↔  ~node.status == -1: slot is loading)
        while (true) {
            if (~node.status == -1) {
                Utility.sleepWithProfile(11200, 50L); // sleep 50 ms while resource loads
                continue;
            }
            // node.status == 1  ↔  ~node.status == -2: slot is populated and payload is set
            if (~node.status == -2 && node.result != null) {
                return (String) node.result;
            }
            break;
        }

        // Fallback: ask the surface/sprite subsystem for the formatted IP/name
        return SurfaceSprite.formatIpAddress(114, entityId);
    }

    /**
     * Advance per-tick scroll/visibility state for the friends-list MessageList panel.
     *
     * CLEAN-BASE CORRECTION: the two branches were SWAPPED in the defective base.
     *   - Cf == 0  : re-validate the chat viewport against the list's scroll extents and
     *                either commit the scroll (zh.a) or hide the panel (se=false).
     *   - Cf != 0  : commit the pending scroll position (zh.b), draw settings if valid,
     *                then clear se and Cf.
     * // obf: void i(byte)  no label
     */
    private final void updateTimers(byte param1) {
        if (this.Cf == 0) {
            // param1 != -106: arm the timer-reset sentinel before re-scroll
            if (param1 != -106) {
                this.tj = -11;
            }

            // Read current scroll extents from friendsList
            int scrollMax    = this.zh.getPanelWidth(16256);   // scroll upper bound
            int scrollOffset = this.zh.getPanelHeight(-21224);  // current offset

            // Check whether the current chat viewport is fully within bounds.
            // (~a <= ~b ⇔ a >= b ; ~a >= ~b ⇔ a <= b)
            if (~this.mouseX <= ~(-10 + this.rh)
                    && this.fg - 10 <= this.mouseY
                    && ~this.mouseX >= ~(scrollMax + this.rh + 10)
                    && ~(10 + this.fg - -scrollOffset) <= ~this.mouseY) {
                // Viewport is valid: commit the scroll
                this.zh.hitTest(this.fg, this.rh, this.mouseY, (byte)-12, this.mouseX);
            } else {
                // Viewport is out of range: hide the friends panel
                this.se = false;
            }
        } else {
            // Cf != 0: commit the pending scroll offset to friendsList
            int scrollResult = this.zh.hitTestNoRender(this.mouseX, this.rh, this.fg, (byte)-40, this.mouseY);
            if (~scrollResult <= -1) {               // ~scrollResult <= -1 ⇔ scrollResult >= 0
                this.incoming.handleSceneUpdates(false, scrollResult);
            }
            this.se = false;
            this.Cf = 0;
        }
    }

    /**
     * Clear transient per-session game state on (re)entry to the game Ek:
     * resets entity-count fields, nulls entity caches, clears per-tick flags,
     * and resets the 100-slot shared name-resolution tables across several classes.
     * // obf: void i(int)  label: client.<i(int)>  (il[115])
     */
    private final void resetGameState(int param1) {
        // Reset connection/state machine counters
        this.kc = 0;   // login/state stage
        this.Xd = 0;   // panel-open flag
        this.bj = 0;   // pending-logout countdown

        this.screenMode = 1;   // SPLIT-FIELD FIX (class b): obf qg is ONE scalar (0=login,1=in-game);
                               // it had been split into screenMode/qg/loggedIn that desynced, so the
                               // main loop (which reads screenMode) never entered the in-game branch
                               // after a successful login. resetGameState sets qg=1 in clean client.java:12406.
        this.Fg = 0;   // fatigue-flash flag

        // Clear chat input buffers
        this.resetChatInput((byte)-49);

        // Reinitialise the surface back-buffer
        this.li.blackScreen(true);                                  // obf: li.a(boolean)
        this.li.draw(this.graphics, this.originX, 256, this.originY); // obf: li.a(Graphics,int,int,int)

        // Remove all active wall/boundary models (field 'world' is TYPE Scene = obf Ek; field 'scene' is TYPE World = obf Hh)
        for (int i = 0; i < this.eh; ++i) {
            this.Ek.removeModel(this.hg[i]);                  // Scene.removeModel (obf Ek.a(ca,int))
            this.Hh.removeObject(this.vc[i], this.Se[i], this.ye[i], 4081); // World.removeObject (obf Hh.a(int,int,int,int))
        }

        // Remove all active NPC/anim models
        for (int i = 0; i < this.hf; ++i) {
            this.Ek.removeModel(this.rd[i]);              // Scene.removeModel
            this.Hh.clearWallObjectAdjacency(true, this.Hj[i], this.yk[i], this.Jd[i], this.Ng[i]); // World (obf Hh.a(boolean,int,int,int,int))
        }

        // Zero entity-count fields
        this.Ah = 0;   // wall/boundary count
        this.eh = 0;   // active wall-model count
        this.hf = 0;   // active NPC-model count
        this.Yc = 0;   // NPC view count

        // Null out NPC server-index cache (4 000 entries)
        for (int i = 0; i < 4000; ++i) {
            this.We[i] = null;
        }

        // Null out previous-tick player array (500 entries; obf: ~i > -501 ⇔ i < 500)
        for (int i = 0; ~i > -501; ++i) {
            this.rg[i] = null;
        }
        this.de = 0;   // "players last" count

        // Null out player server-index cache (5 000 entries)
        for (int i = 0; i < 5000; ++i) {
            this.te[i] = null;
        }

        // Null out previous-tick NPC array (500 entries)
        for (int i = 0; i < 500; ++i) {
            this.Tb[i] = null;
        }

        // Clear per-NPC sleeping/transient flag array (50 entries)
        for (int i = 0; i < 50; ++i) {
            this.bk[i] = false;
        }

        // Reset boolean flags and per-tick state
        this.uk = false;   // sleeping flag
        this.mouseButtonDown = 0;       // obf: this.Bb (GameShell int, inherited)
        this.lastMouseButtonDown = 0;       // obf: this.Qb (GameShell int, inherited)
        this.Cf = 0;       // trade/duel sub-state
        this.Qk = false;   // quest-list open

        // obf: var2 = 58 / ((var1 - -46) / 51) — junk dead division, result discarded.

        this.Fe = false;   // "first login" flag
        FontWidths.listEntryCount = 0;  // obf: n.g — list entry count
        this.Vf = 0;       // fatigue bar value

        // Clear the 100-slot shared name-resolution caches used by several subsystems
        for (int i = 0; i < 100; ++i) {
            BZip.entityNames[i]      = null;   // obf: aa.k — entity name slot
            ImageLoader.scratchBuf[i] = 0;     // obf: pa.g — ImageLoader index slot
            World.G[i]               = null;   // obf: k.G — World name slot
            BitBuffer.UNUSED_N[i]    = 0;       // obf: ja.N — BitBuffer index slot
            SurfaceSprite.recentMessages[i] = null;   // obf: ba.Yb — SurfaceSprite name slot
            NameTable.recentNames[i] = null;   // obf: ub.a — NameTable slot
            FontWidths.entryTypes[i] = 0;      // obf: n.j — FontWidths width slot
        }

        // Reset item counts on shop/quest/inventory list widgets (yd = panelShop; obf yd.c(byte,int))
        this.panelShop.resetItemCount((byte)-33, this.Fh);
        this.panelShop.resetItemCount((byte)-33, this.ud);
        this.panelShop.resetItemCount((byte)-76, this.mc);
    }

    /**
     * Draw the loading progress bar: a 3D rotating-globe backdrop rendered via
     * World/Scene across five camera passes, the game logo sprite, an off-white
     * background (#F8F8F9) and an orange-shadowed (#FF7000) progress strip.
     * // obf: void y(int)  label: client.HC(  (il[0])
     */
    private final void drawProgressBar(int param1) {
        // ---- Pass 1: camera at world tile (50,50) ----
        byte cameraLayer = 0;
        byte tileX = 50, tileZ = 50;
        this.Hh.loadSection(48 * tileX + 23, (byte)-90, 48 * tileZ + 23, cameraLayer);
        this.Hh.addModels(this.objectModels, (byte)-113);

        int worldX = 9728, worldZ = 6400, worldY = 1100;
        int pitch = 888;
        this.Ek.clipFar3d = 4100;
        this.Ek.clipFar2d = 4100;
        this.Ek.fogZFalloff = 1;
        this.Ek.fogZDistance = 4000;
        this.Ek.setCameraOrientation(worldX, worldZ, worldY * 2, 912, -12349, pitch,
                     -this.Hh.getElevation(worldX, worldZ), 0);
        this.Ek.render(-124); // render pass A

        // param1 >= -48: drop the local player ref during loading
        if (param1 >= -48) {
            this.wi = null;
        }

        // Off-white background, then top progress-bar frame + orange shadow gradient
        this.li.fade2black(0xF8F8F9);
        this.li.fade2black(0xF8F8F9);
        this.li.drawBox(0, (byte)65, 0, 0, 6, 512);
        for (int s = 6; s >= 1; --s) {        // obf: ~var9 <= -2 ⇔ var9 >= 1
            this.li.blurRegion(8, s, s, 0, 0xFF7000, 512, 0);
        }
        this.li.drawBox(0, (byte)-104, 0, 194, 20, 512);
        for (int s = 6; s >= 1; --s) {
            this.li.blurRegion(8, s, 194 - s, 0, 0xFF7000, 512, 0);
        }
        this.li.drawSprite(-1, this.tg + 10, 15, 15);
        this.li.drawSprite(this.dg, 200, 123, 512, 0, 0);
        this.li.drawWorld(false, this.dg);

        // ---- Pass 2: camera at (9216,9216) ----
        worldX = 9216; worldZ = 9216; worldY = 1100; pitch = 888;
        this.Ek.clipFar3d = 4100;
        this.Ek.fogZFalloff = 1;
        this.Ek.fogZDistance = 4000;
        this.Ek.clipFar2d = 4100;
        this.Ek.setCameraOrientation(worldX, worldZ, 2 * worldY, 912, -12349, pitch,
                     -this.Hh.getElevation(worldX, worldZ), 0);
        this.Ek.render(-114); // render pass B

        this.li.fade2black(0xF8F8F9);
        this.li.fade2black(0xF8F8F9);
        this.li.drawBox(0, (byte)59, 0, 0, 6, 512);
        for (int s = 6; s >= 1; --s) {
            this.li.blurRegion(8, s, s, 0, 0xFF7000, 512, 0);
        }
        this.li.drawBox(0, (byte)-128, 0, 194, 20, 512);
        for (int s = 6; s >= 1; --s) {
            this.li.blurRegion(8, s, 194 - s, 0, 0xFF7000, 512, 0);
        }
        this.li.drawSprite(-1, 10 + this.tg, 15, 15);
        this.li.drawSprite(1 + this.dg, 200, 124, 512, 0, 0);
        this.li.drawWorld(false, 1 + this.dg);

        // ---- Pass 3: wider view, camera at (11136,10368), y=500, pitch=376 ----
        worldX = 11136; worldZ = 10368; worldY = 500; pitch = 376;
        // Evict all 64 terrain-tile / roof models from World before re-rendering
        for (int t = 0; t < 64; ++t) {
            this.Ek.removeModel(this.Hh.roofModels[0][t]);
            this.Ek.removeModel(this.Hh.wallModels[1][t]);
            this.Ek.removeModel(this.Hh.roofModels[1][t]);
            this.Ek.removeModel(this.Hh.wallModels[2][t]);
            this.Ek.removeModel(this.Hh.roofModels[2][t]);
        }
        this.Ek.clipFar3d = 4100;
        this.Ek.fogZDistance = 4000;
        this.Ek.fogZFalloff = 1;
        this.Ek.clipFar2d = 4100;
        this.Ek.setCameraOrientation(worldX, worldZ, worldY * 2, 912, -12349, pitch,
                     -this.Hh.getElevation(worldX, worldZ), 0);
        this.Ek.render(-111); // render pass C

        this.li.fade2black(0xF8F8F9);
        this.li.fade2black(0xF8F8F9);
        this.li.drawBox(0, (byte)84, 0, 0, 6, 512);
        for (int s = 6; s >= 1; --s) {
            this.li.blurRegion(8, s, s, 0, 0xFF7000, 512, 0);
        }
        this.li.drawBox(0, (byte)-107, 0, 194, 20, 512);
        // Final strip at y=194 (no vertical offset)
        for (int s = 6; s >= 1; --s) {
            this.li.blurRegion(8, s, 194, 0, 0xFF7000, 512, 0);
        }
        this.li.drawSprite(-1, 10 + this.tg, 15, 15);
        this.li.drawSprite(this.tg + 10, 200, 120, 512, 0, 0);
        this.li.drawWorld(false, this.tg + 10);
    }

    /**
     * Walk/box helper keyed off two dimension lookup tables (NameTable.g[] = X-extent,
     * RecordLoader.f[] = Y-extent) selected by index `styleIndex`.  Calls walkToAction
     * twice (outer border, then inner fill) with style-direction offsets.
     *
     * NOTE: the skeleton proposed name `drawBox`, but the body actually dispatches to
     * walkToAction (a(int,boolean,…) @5983, which sends WALK opcodes 16/187) — so the
     * "draw" framing is suspect; the body is transcribed faithfully. See uncertainties.
     *
     * CLEAN-BASE CORRECTION: the two table assignments differ per branch (the defective
     * base made both branches identical):
     *   style==0 || style==4 :  dimX = RecordLoader.f[i] ; dimY = NameTable.g[i]
     *   otherwise            :  dimX = NameTable.g[i]    ; dimY = RecordLoader.f[i]
     * // obf: void b(int,int,int,int,int)  no label  (il[216])
     */
    // Moved to WidgetRenderer.drawBox — delegates here.
    final void drawBox(int magicKey, int styleIndex, int x, int y, int style) {
        widgetRenderer.drawBox(magicKey, styleIndex, x, y, style);
    }

    /**
     * Resolve a clear walkable direction by cycling `si` through the available
     * terrain walkability slots (World collision via this.b(byte,int) probes).
     * `numExtraDirections > 7` enables the extended {±1,±2,±3,+4} search.
     *
     * CLEAN-BASE CORRECTION: the extended-search loop must NOT return when it finds a
     * clear direction (the defective base added an early `return`); after the loop the
     * code re-tests (si&1)==0 && b(91,si) and applies a secondary ±1 nudge, and the whole
     * extended block is nested inside `if (numExtraDirections > 7)`.
     * // obf: void q(byte)  no label  (skeleton proposed name: clearScreen) (il[389])
     */
    // Moved to WidgetRenderer.clearScreen — delegates here.
    private final void clearScreen(byte numExtraDirections) {
        widgetRenderer.clearScreen(numExtraDirections);
    }

    /**
     * Tear down the letterbox regions around the game viewport: paints the four
     * black strips outside the logical 512×334 area via AWT Graphics.
     * // obf: void p(byte)  no label  (il[124])
     */
    final void resetPanels(byte param1) {
        int leftW  = this.originX;                              // left strip width  (component X-offset)
        int topH   = this.originY;                               // top strip height  (component Y-offset)
        int rightW = -this.screenWidth + this.Rh + -leftW;          // right strip width
        int botH   = -topH - this.screenHeight - 12 + this.Hf;       // bottom strip height

        // obf: var6 = -40 / ((6 - var1) / 38) — junk dead division, result discarded.

        // Proceed if any strip is positive (obf: var2>0 || -1>~var4 || 0<var3 || var5>0)
        if (leftW > 0 || rightW > 0 || topH > 0 || botH > 0) {
            // Resolve the AWT host container for getGraphics()
            java.awt.Component target;
            if (this.hj) {
                target = (ClientStream.applet != null) ? ClientStream.applet : this;   // applet/fullscreen mode (obf: da.gb)
            } else {
                target = InputState.gameFrame;                      // standalone frame (kb.a)
            }

            try {
                java.awt.Graphics g = target.getGraphics();
                if (g == null) {
                    return;
                }

                g.setColor(java.awt.Color.black);
                if (leftW > 0) {
                    g.fillRect(0, 0, leftW, this.Hf);                 // left strip
                }
                if (topH > 0) {                                       // obf: -1 > ~var3 ⇔ var3 > 0
                    g.fillRect(0, 0, this.Rh, topH);                  // top strip
                }
                if (rightW > 0) {
                    g.fillRect(-rightW + this.Rh, 0, rightW, this.Hf);// right strip
                }
                if (botH > 0) {                                       // obf: ~var5 < -1 ⇔ var5 > 0
                    g.fillRect(0, -botH + this.Hf, this.Rh, botH);    // bottom strip
                }
            } catch (Exception e) {
                // swallow (defective surface / not yet realised) — matches base
            }
        }
    }

    /**
     * Run a Runnable on the GameShell deferred-event queue.
     * Delegates to ImageLoader.imageWidthCarrier.startThread(true, runnable, priority).
     * Overrides GameShell.startThread(int,Runnable) (obf {@code a(int,Runnable)}).
     * // obf: void a(int,Runnable)  label: client.S(  (il[223])
     */
    @Override
    public final void startThread(int priority, Runnable task) {
        ImageLoader.imageWidthCarrier.startThread(true, task, priority); // obf: pa.k.a(boolean,Runnable,int)
    }

    /**
     * Forward a scroll position to whichever sub-panel is currently open, gated on the
     * load state (qg) and the open-panel id (Xd) / members flag (Kg).
     *
     * CLEAN-BASE CORRECTION vs the defective base:
     *   - (Xd==0) scrolls `ge`  (NOT panelQuest as the defective base claimed).
     *   - (Xd==2) scrolls `yi`.
     *   - members branch scrolls `Af`; f2p branch scrolls `yd`.
     * NAMING: the English panel names for ge/yi/Af/yd are CONTRADICTED across the
     * skeleton (Af=panelQuest, ge=panelTrade) vs the actual construction code
     * (clean base: Af = new qa(li,100) ⇒ panelDuel; ge,yi = new qa(li,50) built in
     * p(int)) and the existing part files. Until that is resolved in a dedicated
     * panel-naming pass, obf field names are kept here to avoid asserting a wrong
     * mapping; the `panelShop`/`yd` binding (stats panel in f2p) is the only one
     * already consistent across parts. See uncertainties.
     * // obf: void a(byte,int)  no label  (il[186])
     */
    @Override
    protected final void handleKeyPress(byte panelId, int scrollY) {
        // qg == 0: login screen  (SPLIT-FIELD FIX class b: was this.qg, now the unified screenMode)
        if (this.screenMode == 0) {
            if (this.Xd == 0 && this.ge != null) {
                this.ge.handleKeyInput(-12, scrollY);           // obf: this.ge.a(-12,var2)  (Panel.handleKeyInput)
            }
            // Xd == 2 (obf: ~Xd == -3)
            if (~this.Xd == -3 && this.yi != null) {
                this.yi.handleKeyInput(-12, scrollY);           // obf: this.yi.a(-12,var2)
            }
        }

        if (panelId <= 105) {
            return;
        }

        // qg == 1: game-world view (obf: ~this.qg == -2)  (SPLIT-FIELD FIX class b: unified screenMode)
        if (this.screenMode == 1) {
            if (this.Kg) {
                // Members server: scroll the members-only panel (Af)
                this.Af.handleKeyInput(-12, scrollY);
                return;
            }
            // Non-members: only scroll the stats panel (yd = panelShop) when no
            // duel/fatigue/quest overlay is active (Bj==0 && Vf==0 && !Qk && gc==0).
            if (~this.Bj == -1 && ~this.Vf == -1 && !this.Qk && this.gc == 0) {
                this.panelShop.handleKeyInput(-12, scrollY);    // obf: this.yd.a(-12,var2)
            }
        }
    }

    /**
     * Blit a UI sprite (by draw-list index) to the li at (x, y) via walkToAction,
     * with an optional screen-mode flag that sets cl = 61.
     *
     * CLEAN-BASE CORRECTION: drawMode 1 and 2 were SWAPPED in the defective base.
     *   drawMode==0 → mode -8  at (x, y-1)..(x, y)
     *   drawMode==1 → mode 126 at (x-1, y)..(x, y)   [fall-through case]
     *   drawMode==2 → mode 118 at (x, y)..(x, y)
     * // obf: void a(boolean,int,int,int)  no label  (il[388])
     */
    // Moved to WidgetRenderer.drawSprite — delegates here.
    final void drawSprite(boolean setScreenMode, int x, int y, int drawMode) {
        widgetRenderer.drawSprite(setScreenMode, x, y, drawMode);
    }

    /**
     * "Ready/loaded" guard check — always returns true; the junk modulo is an
     * opaque-predicate artefact whose result is discarded.
     * // obf: boolean f(byte)  label: client.LC(  (il[226])
     */
    private final boolean isLoaded(byte param1) {
        // obf: int var2 = 89 % ((param1 - -74) / 51); — dead, result unused.
        return true;
    }

    /**
     * XOR string-pool decoder stage 1: converts a String to a char[], XOR-ing the
     * sole character with 0x7E ('~') when the string is shorter than 2 chars.
     * Outer wrapper in z(z("…")) double-decode of STRINGS[].
     * // obf: static char[] z(String)  no label
     */
    private static char[] xorDecode1(String encoded) {
        char[] chars = encoded.toCharArray();
        if (chars.length < 2) {
            chars[0] = (char) (chars[0] ^ 126);
        }
        return chars;
    }

    /**
     * XOR string-pool decoder stage 2: XOR each char with a position-dependent key
     * from the 5-byte table {34, 7, 117, 116, 126} (index = pos % 5), then intern.
     * Pairs with xorDecode1 to decode STRINGS[] entries.
     * // obf: static String z(char[])  no label
     */
    private static String xorDecode2(char[] chars) {
        for (int i = 0; i < chars.length; i++) {
            byte key;
            switch (i % 5) {
                case 0:  key =  34; break;
                case 1:  key =   7; break;
                case 2:  key = 117; break;
                case 3:  key = 116; break;
                default: key = 126; break;
            }
            chars[i] = (char) (chars[i] ^ key);
        }
        return new String(chars).intern();
    }

    /**
     * Wall/diagonal collision probe used by {@link #clearScreen(byte)} to test whether
     * a candidate move direction {@code dir} away from the local player's current tile
     * is blocked by an adjacent wall.  Reads {@code World.objectAdjacency} (obf
     * {@code Hh.bb}; the field named {@code Hh} is TYPE World) at the player tile and
     * its two predecessor tiles for the {@code &amp; 128} wall bit.
     *
     * <p>Returns {@code false} as soon as a blocking wall is found for the given
     * direction; {@code true} if the path stays clear.
     *
     * <p>Transcribed faithfully from the clean base (obf {@code private final boolean
     * b(byte var1, int var2)} @clean L17353); the {@code byte} first arg is an
     * anti-tamper dummy, the {@code Dk++} profiling counter and the opaque-predicate /
     * ErrorHandler wrappers are stripped per deob convention.
     *
     * obf: private final boolean b(byte param1, int dir)
     */
    final boolean isDirectionWalkable(byte dummy, int dir) {
        int tileX = this.wi.currentX / 128;   // obf: wi.i / 128
        int tileY = this.wi.currentY / 128;    // obf: wi.K / 128

        for (int d = 2; d >= 1; --d) {                  // obf: var5 = 2; while (1 <= var5) ... var5--
            // North (dir == 1)
            if (dir == 1
                    && ((128 & this.Hh.objectAdjacency[tileX][tileY - d]) == 128
                        || (128 & this.Hh.objectAdjacency[tileX - d][tileY]) == 128
                        || (this.Hh.objectAdjacency[tileX - d][tileY - d] & 128) == 128)) {
                return false;
            }
            // North-east (dir == 3)
            if (dir == 3
                    && ((128 & this.Hh.objectAdjacency[tileX][tileY + d]) == 128
                        || (this.Hh.objectAdjacency[tileX - d][tileY] & 128) == 128
                        || (128 & this.Hh.objectAdjacency[tileX - d][tileY + d]) == 128)) {
                return false;
            }
            // East (dir == 5)
            if (dir == 5
                    && ((this.Hh.objectAdjacency[tileX][tileY + d] & 128) == 128
                        || (this.Hh.objectAdjacency[tileX + d][tileY] & 128) == 128
                        || (this.Hh.objectAdjacency[tileX + d][tileY + d] & 128) == 128)) {
                return false;
            }
            // South-east (dir == 7)
            if (dir == 7
                    && ((this.Hh.objectAdjacency[tileX][tileY - d] & 128) == 128
                        || (128 & this.Hh.objectAdjacency[tileX + d][tileY]) == 128
                        || (128 & this.Hh.objectAdjacency[tileX + d][tileY - d]) == 128)) {
                return false;
            }
            // Cardinal single-edge probes
            if (dir == 0 && (this.Hh.objectAdjacency[tileX][tileY - d] & 128) == 128) {
                return false;
            }
            if (dir == 2 && (this.Hh.objectAdjacency[tileX - d][tileY] & 128) == 128) {
                return false;
            }
            if (dir == 4 && (128 & this.Hh.objectAdjacency[tileX][tileY + d]) == 128) {
                return false;
            }
            if (dir == 6 && (this.Hh.objectAdjacency[tileX + d][tileY] & 128) == 128) {
                return false;
            }
        }

        return true;
    }

    // =========================================================================
    // ===== methods absent from the reconcile, transcribed from clean base =====
    // =========================================================================

    /**
     * Resets the right-click menu / login-entry state. Clean: {@code e(byte)} @client.java:12828.
     * (obf fields de/Yc/Xd/Oc; Xf=username, qg=loggedIn, wh=password.)
     */
    final void resetMenuState(byte arg) {
        this.de = 0;
        this.Yc = 0;
        this.Xd = 0;
        if (arg != -88) {
            this.Oc = null;
        }
        this.username = "";
        this.screenMode = 0; // SPLIT-FIELD FIX (class b): obf qg=0 -> back to login screen (was this.loggedIn=0)
        this.password = "";
    }

    /**
     * Builds the shop panel controls. Clean: {@code O(int)} @client.java:1883.
     * (obf yd=panelShop; Fh/bh/ud/mc are control-id fields.)
     */
    private final void initShopPanel(int arg) {
        this.panelShop = new Panel(this.li, 10);
        this.Fh = this.panelShop.addListBox(502, arg, 5, 20, 269, 1, arg ^ 7, true);
        this.bh = this.panelShop.addTextInputField(80, 14, false, 7, 1, 324, 14179, 498, true);
        this.ud = this.panelShop.addListBox(502, 56, 5, 20, 269, 1, 63, true);
        this.mc = this.panelShop.addListBox(502, 56, 5, 20, 269, 1, 63, true);
        this.panelShop.setFocus(this.bh, -103);
        // SPLIT-FIELD FIX (class b): obf `yd` (created here by clean O(int) @client.java:1886) was
        // aliased into TWO deob Panel fields — `panelShop` (written here) and `panelMessageTabs`
        // (read by the in-world handleGameInput chat-tab handler but NEVER assigned -> NPE on entering
        // the game). Likewise its 4 controls Fh/bh/ud/mc were aliased to controlListChat/Input/Quest/
        // Private. They are the SAME panel/controls; bind the aliases so the in-game UI is non-null.
        this.panelMessageTabs   = this.panelShop;
        this.controlListChat    = this.Fh;   // yd.j[Fh]  tab 1 (chat)
        this.controlListInput   = this.bh;   // yd text input (clean bh)
        this.controlListQuest   = this.ud;   // yd.j[ud]  tab 2 (quest)
        this.controlListPrivate = this.mc;   // yd.j[mc]  tab 3 (private)
        // SPLIT-FIELD FIX (2): showServerMessage (clean a(boolean,String,...) @client.java:11104)
        // reads the SAME obf fields under a THIRD alias set — `messagePanel`/`tabChat`/`tabQuest`/
        // `tabPrivate` — which were also never assigned (NPE on the first server message: opcode 131
        // arrives right after login -> onFriendUpdate -> showServerMessage). Bind them to the same
        // obf yd/Fh/ud/mc so the chat/quest/private tabs are non-null in showServerMessage too.
        this.messagePanel = this.panelShop;  // obf yd
        this.tabChat      = this.Fh;         // obf Fh  (type 4)
        this.tabQuest     = this.ud;         // obf ud  (type 3)
        this.tabPrivate   = this.mc;         // obf mc  (type 1/2)
    }

    /**
     * Opens the Ek socket. Clean: {@code a(int,int,String)} @client.java:12289.
     * Uses the J++ LoaderThread async resolver when sandboxed (applet/gameFrame),
     * otherwise connects directly. (obf kb.a=InputState.gameFrame, da.gb=ClientStream.applet,
     * pa.k=ImageLoader.imageWidthCarrier, g=ListNode.)
     */
    /**
     * Panel-action dispatcher for the social/friends menu. Clean: {@code a(String,int,String)}
     * @client.java:9465. Builds the right-click report/PM menu entries for a player name,
     * returning true if the action was handled. (obf w.a=WorldEntity.trimAndValidateString,
     * wi.C=wi.message, n.g=friendListCount, db.g=ignoreListCount, ua.h=friendListNames,
     * ia.a=ignoreListDisplayNames, Fj=keyState/friend flags, zh=zh.)
     */
    private final boolean a(String name, int mode, String displayName) {
        String hashed = WorldEntity.trimAndValidateString(name, (byte) 92);
        if (hashed == null) {
            return false;
        }
        if (mode <= 126) {
            return true;
        }
        if (hashed.equals(WorldEntity.trimAndValidateString(this.wi.message, (byte) 93))) {
            return false;
        }

        boolean isFriend = false;
        boolean friendOnline = false;
        for (int i = 0; i < this.friendListCount; i++) {
            if (hashed.equals(WorldEntity.trimAndValidateString(this.friendListNames[i], (byte) 52))) {
                isFriend = true;
                if ((4 & Fj[i]) != 0) {
                    friendOnline = true;
                    break;
                }
            }
        }

        if (isFriend && friendOnline) {
            this.zh.addEntryFull(STRINGS[178], STRINGS[15] + displayName, displayName, 2830, name, (byte) -50);
        } else {
            boolean isIgnored = false;
            for (int i = 0; i < this.ignoreListCount; i++) {
                if (hashed.equals(WorldEntity.trimAndValidateString(this.ignoreListDisplayNames[i], (byte) 51))) {
                    isIgnored = true;
                    break;
                }
            }
            if (!isIgnored) {
                this.zh.addEntryFull(STRINGS[181], STRINGS[15] + displayName, displayName, 2831, name, (byte) 80);
                this.zh.addEntryFull(STRINGS[179], STRINGS[15] + displayName, displayName, 2832, name, (byte) -37);
            }
        }

        this.zh.addEntryFull(STRINGS[120], STRINGS[15] + displayName, displayName, 2833, name, (byte) 110);
        return true;
    }

    private final Socket createSocket(int dummy, int port, String host) throws IOException {
        Socket socket;
        if (InputState.gameFrame == null && ClientStream.applet != null) {
            ListNode node = ImageLoader.imageWidthCarrier.openSocket(host, port, -75);
            while (node.status == 0) {
                Utility.sleepWithProfile(dummy ^ -11212, 50L);
            }
            if (node.status != 1) {
                throw new IOException();
            }
            socket = (Socket) node.result;
            if (socket == null) {
                throw new IOException();
            }
        } else if (InputState.gameFrame == null) {
            socket = new Socket(InetAddress.getByName(this.getCodeBase().getHost()), port);
        } else {
            socket = new Socket(InetAddress.getByName(host), port);
        }
        socket.setSoTimeout(30000);
        socket.setTcpNoDelay(true);
        return socket;
    }

}
