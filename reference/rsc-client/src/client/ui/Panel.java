package client.ui;

import java.awt.Font;
import java.awt.FontMetrics;
import client.shell.GameShell;
import client.scene.Surface;
import client.data.BZip;
import client.data.CacheFile;
import client.world.GameCharacter;
import client.shell.GameFrame;
import client.net.ISAAC;
import client.scene.ImageLoader;
import client.net.Packet;
import client.net.SocketFactory;
import client.scene.SpriteScaler;
import client.net.StringCodec;
import client.scene.SurfaceImageProducer;
import client.util.Timer;

/**
 * Panel — UI widget/panel container for the RSC client rev ~233-235 (Microsoft J++ build).
 *
 * Manages a flat array of up to {@code capacity} heterogeneous widgets, each identified
 * by a per-slot "widget type" code stored in {@link #widgetType}[]. Widgets are added via
 * the various {@code addXxx()} methods, which allocate the next free slot ({@code nextSlot})
 * and return its index. The single render pass {@link #render(byte)} iterates all active
 * slots and dispatches to private draw helpers based on the type code.
 *
 * Widget type codes:
 *   0  — label (text drawn dim/gray; passively displayed)
 *   1  — label-button (text, optionally password-masked, selected/focused state)
 *   2  — oval/circle button
 *   3  — plain rectangle region
 *   4  — scrollable list box (string items + colour codes + scroll tracking)
 *   5  — text-input field (single-line, can be password)
 *   6  — text-input field (variant: may have separate display string)
 *   7  — tab/button text label (horizontal strip, item centred)
 *   8  — icon/image sprite widget
 *   9  — list box (vertical, with scrollbar)
 *  10  — ellipsis/progress widget (centred)
 *  11  — centred image/sprite area
 *  12  — native sprite from Surface.kb[]/R[] dimensions
 *  14  — checkbox/toggle button (vb toggles on click)
 *  15  — invisible hit-test region (only sets D[] on click)
 *
 * The Panel holds a reference to {@link Surface} ({@code surface}) for all actual drawing.
 * Mouse position is injected each frame via {@link #handleMouseInput(int,int,int,int,int)}.
 * The font in use is "Helvetica" (loaded via {@link #loadFont(GameShell,String,int,int)}).
 *
 * Four instances are created in Mudclient; each covers a distinct HUD region
 * (inventory, chat, minimap overlay, main viewport controls).
 *
 * Obfuscated class name: {@code qa}.
 */
public final class Panel {

    // -------------------------------------------------------------------------
    // Per-slot widget data arrays (all parallel, indexed by slot index 0..nextSlot-1)
    // -------------------------------------------------------------------------

    /** Per-slot string[][] payloads (tab labels for type-8 icon lists, etc.). obf: Ab */
    private String[][] slotStringArrays;

    /** Per-slot integer array payloads (colour arrays for list items, etc.). obf: m */
    private int[][] slotIntArrays;

    /** Per-slot secondary string[][] (feedback/action strings). obf: Fb */
    private String[][] slotFeedback;

    /** Per-slot primary display string[] (list item labels). obf: p */
    private String[][] slotLabels;

    /** Per-slot text buffer (current text value of input fields / displayed string). obf: yb */
    private String[] slotText;

    /** Per-slot widget type code (see class doc). obf: U */
    private int[] widgetType;

    /** Per-slot X coordinate (left edge for most widgets). obf: kb */
    private int[] widgetX;

    /** Per-slot Y coordinate (top edge for most widgets). obf: B */
    private int[] widgetY;

    /** Per-slot width. obf: ob */
    private int[] widgetW;

    /** Per-slot height. obf: O */
    private int[] widgetH;

    /** Per-slot font/style index passed to Surface draw calls. obf: k */
    private int[] fontIndex;

    /** Per-slot max length (input fields) or item capacity (lists). obf: sb */
    private int[] maxLen;

    /** Per-slot scroll-position (list row offset, 0 = top). obf: j */
    public int[] scrollPos;

    /** Per-slot item count (list widgets). obf: pb */
    public int[] itemCount;

    /** Per-slot "visible / enabled" flag. obf: g */
    private boolean[] visible;

    /** Per-slot "mouse-activated / clicked" flag (cleared on read). obf: D */
    private boolean[] activated;

    /** Per-slot "password mask" flag: true → display text as 'XXXX'. obf: cb */
    private boolean[] passwordMask;

    /** Per-slot "highlighted / selected colour" flag (false=dim, true=white). obf: Y */
    private boolean[] highlighted;

    /** Per-slot "drag-scroll active" flag (set while user is dragging scrollbar). obf: d */
    private boolean[] dragging;

    /** Per-slot "selected item index" (list widget current selection). obf: vb */
    private int[] selectedItem;

    /** Per-slot "hovered item index" (which item mouse is over). obf: N */
    private int[] hoveredItem;

    // -------------------------------------------------------------------------
    // Instance state
    // -------------------------------------------------------------------------

    /** Backing Surface used for all drawing operations. obf: w */
    private Surface surface;

    /** Next free widget slot index. obf: eb */
    private int nextSlot = 0;

    /** Currently focused input-field slot index (-1 = none). obf: gb */
    private int focusedSlot = -1;

    /** Last mouse X (pixel column) seen in handleMouseInput. obf: bb */
    private int mouseX = 0;

    /** Last mouse Y (pixel row) seen in handleMouseInput. obf: hb */
    private int mouseY = 0;

    /**
     * Mouse-click mode injected by handleMouseInput:
     *   0 = no click, 1 = left-click, 2 = right-click.
     * obf: zb
     */
    private int clickMode = 0;

    /**
     * Mouse-button state for drag detection (1 = button held).
     * obf: Hb
     */
    private int mouseButtonState = 0;

    /** Click-drag counter (used for scroll-drag physics). obf: G */
    private int G = 0;

    /** Whether the selected-item highlight uses red (true = "active selection"). obf: t */
    private boolean activeSelectionHighlight = true;

    // -------------------------------------------------------------------------
    // Packed RGB colour constants (computed once in constructor via packRgb()).
    // Named after their visual role in the RSC "stone" UI palette.
    // -------------------------------------------------------------------------

    /** Mid rose-grey — used as border/highlight colour 0. obf: R */
    private int colorMidRose;          // RGB(176,114,114) = 0xB07272
    /** Deep dark rose — used for narrow accent bar. obf: qb */
    private int colorDarkRose;         // RGB(62,14,14)    = 0x3E0E0E
    /** Light stone cream — lightest panel face. obf: C */
    private int colorStoneLight;       // RGB(232,208,200) = 0xE8D0C8
    /** Warm sandstone — panel face mid-light. obf: i */
    private int colorSandstone;        // RGB(184,129,96)  = 0xB88160
    /** Dark brown — deep shadow / inset. obf: ib */
    private int colorDarkBrown;        // RGB(115,95,53)   = 0x735F35
    /** Medium warm tan — standard face. obf: fb */
    private int colorTan;              // RGB(171,142,117) = 0xAB8E75
    /** Dark stone — secondary shadow. obf: E */
    private int colorDarkStone;        // RGB(158,122,98)  = 0x9E7A62
    /** Darker stone — button recess. obf: f */
    private int colorButtonRecess;     // RGB(136,100,86)  = 0x886456
    /** Warm grey — inner panel background. obf: tb */
    private int colorInnerBg;          // RGB(179,146,135) = 0xB39287
    /** Mid warm grey — secondary inner. obf: F */
    private int colorMidWarmGrey;      // RGB(151,112,97)  = 0x977061
    /** Dark inner shadow. obf: Eb */
    private int colorInnerShadow;      // RGB(136,102,88)  = 0x886658
    /** Darkest shadow edge. obf: J */
    private int colorShadowEdge;       // RGB(120,93,84)   = 0x785D54

    // -------------------------------------------------------------------------
    // Static (shared across all Panel instances)
    // -------------------------------------------------------------------------

    /**
     * Glyph bitmap data for the current font, written by FontBuilder.rasterizeGlyph() and
     * used by Surface draw routines. Initialised to pa.a(-126) (the default
     * pixel/glyph table produced by ImageLoader). obf: l
     */
    public static byte[] fontGlyphData = ImageLoader.buildMuLawTable(-126);  // obf: l = pa.a(-126)

    /**
     * Per-texture numeric table (obf {@code qa.K}; clean qa.java:27 {@code static int[] K}).
     * Allocated and filled by {@link client.net.SocketFactory#initGameData} (GameData texture tier),
     * also consulted by {@code World} for tile-decoration colour lookups.
     */
    public static int[] texK;

    // -------------------------------------------------------------------------
    // Profiling counters (dead, stripped — listed here for traceability only)
    // These static int fields were incremented once per method entry as
    // per-method profiling counters; all reads are dead after stripping.
    // -------------------------------------------------------------------------
    /** obf: s  */ static int _cnt_s;
    /** obf: P  */ static int _cnt_P;
    /** obf: L  */ static int _cnt_L;
    /** obf: ub */ static int _cnt_ub;
    /** obf: x  */ static int _cnt_x;
    /** obf: Gb */ static int _cnt_Gb;
    /** obf: M  */ static int _cnt_M;
    /** obf: lb */ static int _cnt_lb;
    /** obf: r  */ static int _cnt_r;
    /** obf: nb */ static int _cnt_nb;
    /** obf: db */ static int _cnt_db;
    /** obf: mb */ static int _cnt_mb;
    /** obf: xb */ static int _cnt_xb;
    /** obf: H  */ static int _cnt_H;
    /** obf: y  */ static int _cnt_y;
    /** obf: a  */ static int _cnt_a;
    /** obf: o  */ static int _cnt_o;
    /** obf: wb */ static int _cnt_wb;
    /** obf: u  */ static int _cnt_u;
    /** obf: ab */ static int _cnt_ab;
    /** obf: h  */ static int _cnt_h;
    /** obf: X  */ static int _cnt_X;
    /** obf: Db */ static int _cnt_Db;
    /** obf: e  */ static int _cnt_e;
    /** obf: n  */ static int _cnt_n;
    /** obf: q  */ static int _cnt_q;
    /** obf: V  */ static int _cnt_V;
    /** obf: Bb */ static int _cnt_Bb;
    /** obf: v  */ static int _cnt_v;
    /** obf: jb */ static int _cnt_jb;
    /** obf: b  */ static int _cnt_b;
    /** obf: I  */ static int _cnt_I;
    /** obf: Z  */ static int _cnt_Z;
    /** obf: T  */ static int _cnt_T;
    /** obf: Q  */ static int _cnt_Q;
    /** obf: rb */ static int _cnt_rb;
    /** obf: S  */ static int _cnt_S;
    /** obf: A  */ static int _cnt_A;
    /** obf: c  */ static int _cnt_c;
    /** obf: Cb */ static int _cnt_Cb;
    /** obf: Bb2*/ // (Bb used for loadFont counter above)

    /**
     * Known-valid character set string; used as the iteration set when
     * rasterising glyphs via FontBuilder.  Decoded from XOR pool W[7].
     * Value: "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789!"£$%^&*()-_=+[{]};:'@#~,<.>/?\\| "
     */
    private static final String VALID_CHARS =
        "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz" +
        "0123456789!\"£$%^&*()-_=+[{]};:'@#~,<.>/?\\| ";
    // obf: W[7] decoded via XOR pool

    /** Font name constant.  Decoded from XOR pool W[31]. obf: W[31] */
    private static final String FONT_NAME_HELVETICA = "Helvetica";

    /** Font name prefix used in the font-spec parser. obf: W[32] */
    private static final String FONT_PREFIX = "helvetica";

    /** Font file suffix stripped during font-spec parse. obf: W[29] */
    private static final String FONT_SUFFIX = ".jf";

    /** Two-space separator string (used between tab labels). obf: W[20] */
    private static final String TWO_SPACES = "  ";

    // =========================================================================
    // Constructor
    // =========================================================================

    /**
     * Construct a new Panel backed by the given Surface, with capacity for
     * {@code capacity} widgets.
     *
     * Allocates all per-slot arrays and pre-computes the twelve UI colour
     * constants from packed RGB values.
     *
     * @param surface   the Surface to draw into (obf param: ua var1)
     * @param capacity  maximum number of widgets this panel can hold (obf: var2)
     */
    public Panel(Surface surface, int capacity) {
        // obf: qa(ua var1, int var2)
        this.slotFeedback  = new String[capacity][];
        this.highlighted   = new boolean[capacity];
        this.widgetY       = new int[capacity];
        this.passwordMask  = new boolean[capacity];
        this.fontIndex     = new int[capacity];
        this.itemCount     = new int[capacity];
        this.widgetW       = new int[capacity];
        this.visible       = new boolean[capacity];
        this.surface       = surface;
        this.hoveredItem   = new int[capacity];
        this.selectedItem  = new int[capacity];
        this.widgetType    = new int[capacity];
        this.slotIntArrays = new int[capacity][];
        this.slotText      = new String[capacity];
        this.dragging      = new boolean[capacity];
        this.scrollPos     = new int[capacity];
        this.slotStringArrays = new String[capacity][];
        this.widgetH       = new int[capacity];
        this.activated     = new boolean[capacity];
        this.slotLabels    = new String[capacity][];
        this.widgetX       = new int[capacity];
        this.maxLen        = new int[capacity];

        // Pre-compute the twelve UI palette colours.
        // packRgb(r, g, b, _) encodes RGB(r, g, b) → packed int via Surface colour math.
        this.colorMidRose      = packRgb(176, 114, 114, (byte)-98);
        this.colorDarkRose     = packRgb(62,  14,  14,  (byte)-75);
        this.colorStoneLight   = packRgb(232, 208, 200, (byte)-88);
        this.colorSandstone    = packRgb(184, 129, 96,  (byte)-79);
        this.colorDarkBrown    = packRgb(115, 95,  53,  (byte)-92);
        this.colorTan          = packRgb(171, 142, 117, (byte)-124);
        this.colorDarkStone    = packRgb(158, 122, 98,  (byte)-117);
        this.colorButtonRecess = packRgb(136, 100, 86,  (byte)-113);
        this.colorInnerBg      = packRgb(179, 146, 135, (byte)-89);
        this.colorMidWarmGrey  = packRgb(151, 112, 97,  (byte)-122);
        this.colorInnerShadow  = packRgb(136, 102, 88,  (byte)-124);
        this.colorShadowEdge   = packRgb(120, 93,  84,  (byte)-92);
    }

    // =========================================================================
    // Colour helpers
    // =========================================================================

    /**
     * Pack an RGB triple into a single colour int via the Surface colour pipeline.
     * Called only from the constructor; the dummy {@code by} param (always dead) was
     * an anti-tamper guard that conditionally cleared {@code widgetX} — removed.
     *
     * Implementation delegates to Surface.a(int,int,int,int) which does the
     * actual colour-space conversion used by the software renderer.
     *
     * obf: private final int a(int var1, int var2, int var3, byte var4)
     *   → return o.a(ua.C * n4 / 114, 9570, ta.g * n1 / 176, n2 * aa.d / 114)
     *
     * @param r  red component 0-255   (obf: var1 / n1)
     * @param g  green component 0-255 (obf: var2 / n2)
     * @param b  blue component 0-255  (obf: var3 / n4 — note swapped arg order in obf call)
     * @param _  dummy anti-tamper param (dropped)
     * @return   packed colour int
     */
    private int packRgb(int r, int g, int b, byte unusedGuard) {
        // obf: return o.a(ua.C * n4/114, 9570, ta.g * n1/176, n2 * aa.d/114)
        // where n4=b, n1=r, n2=g; ua.C, ta.g, aa.d are "255" scale factors from
        // Surface (ua), GameCharacter (ta), and BZip (aa) static fields.
        // Anti-tamper guard `if (var4 >= -70) this.kb = null` is dead; removed.
        return ISAAC.packColor(
            Surface.tamperMagic * b / 114,
            9570,
            GameCharacter.COLOUR_SPRITE_BASE * r / 176,
            g * BZip.VIEWPORT_HEIGHT_RATIO / 114
        );
    }

    // =========================================================================
    // Widget registration methods ("add…")
    // =========================================================================

    /**
     * Add a plain label (type 0): draws text at (x,y) in a dim colour.
     * Text is drawn without any focus/selection state.
     *
     * obf: final int a(boolean bl, byte by, int n2, int n3, String string, int n4)
     *   - widgetType = 1 in obf; BUT type-0 path in render() calls same draw helper
     *     as a non-focused label.  The dummy `by` guard calls junk helper; removed.
     *   - Returns slot index.
     *
     * @param highlighted  whether to draw in white (true) or grey (false)
     * @param font         font index for Surface.drawString
     * @param x            left X
     * @param y            top Y (baseline = y + ascent)
     * @param text         text to display
     * @return slot index
     */
    public final int addLabel(boolean highlighted, byte _guard, int font, int x, String text, int y) {
        // obf: a(boolean bl, byte by, int n2, int n3, String string, int n4)
        // Dummy guard: if (by >= -71) call junk helper — removed.
        widgetType[nextSlot] = 1;
        visible[nextSlot]    = true;
        activated[nextSlot]  = false;
        fontIndex[nextSlot]  = font;
        this.highlighted[nextSlot] = highlighted;
        widgetX[nextSlot]    = x;
        widgetY[nextSlot]    = y;
        slotText[nextSlot]   = text;
        return nextSlot++;
    }

    /**
     * Add an oval/circle widget (type 2) centred at (cx, cy) with given
     * width and height. A dummy guard triggers junk helper if cx > -56; removed.
     *
     * obf: final int c(int n2, int n3, int n4, int n5, int n6)
     *   n2=cx (guard), n3=w, n4=h, n5=cx, n6=cy
     *
     * @param _guard  anti-tamper (dropped; triggers junk when > -56)
     * @param w       oval width
     * @param h       oval height
     * @param cx      centre X
     * @param cy      centre Y
     * @return slot index
     */
    public final int addOval(int _guard, int w, int h, int cx, int cy) {
        // obf: c(int n2, int n3, int n4, int n5, int n6)
        widgetType[nextSlot] = 2;
        visible[nextSlot]    = true;
        activated[nextSlot]  = false;
        widgetX[nextSlot]    = cx - w / 2;
        widgetY[nextSlot]    = cy - h / 2;
        widgetW[nextSlot]    = w;
        widgetH[nextSlot]    = h;
        return nextSlot++;
    }

    /**
     * Add an "ellipsis/progress" widget (type 10) centred at (cx, cy).
     *
     * obf: final int d(int var1, int var2, int var3, int var4, int var5)
     *   var1=cx, var2=w, var3=cy, var4=guard, var5=h
     *
     * @param cx       centre X
     * @param w        width
     * @param cy       centre Y
     * @param _guard   anti-tamper (dropped; junk division on var4)
     * @param h        height
     * @return slot index
     */
    public final int addProgressWidget(int cx, int w, int cy, int _guard, int h) {
        // obf: d(int var1, int var2, int var3, int var4, int var5)
        widgetType[nextSlot] = 10;
        visible[nextSlot]    = true;
        activated[nextSlot]  = false;
        widgetX[nextSlot]    = cx - w / 2;
        widgetY[nextSlot]    = cy - h / 2;
        widgetW[nextSlot]    = w;
        widgetH[nextSlot]    = h;
        return nextSlot++;
    }

    /**
     * Add a text-input field (type 5) with optional password masking.
     *
     * Anti-tamper guard {@code n7 != 14179} causes early return of 2; removed
     * (all callers pass 14179).
     *
     * obf: final int a(int var1, int var2, boolean var3, int var4, int var5,
     *                   int var6, int var7, int var8, boolean var9)
     *   Assignments: cb=var3, k=var5, Y=var9, kb=var4, B=var6, ob=var8, O=var2, sb=var1.
     *
     * @param maxLen      maximum text length     (obf var1 → sb)
     * @param h           field height             (obf var2 → O)
     * @param password    true → mask with 'XXXX'  (obf var3 → cb)
     * @param x           left X                   (obf var4 → kb)
     * @param font        font index               (obf var5 → k)
     * @param y           top Y                    (obf var6 → B)
     * @param _antiTamper must equal 14179 in live code (obf var7; guard removed)
     * @param w           field width              (obf var8 → ob)
     * @param highlighted true = white text colour (obf var9 → Y)
     * @return slot index
     */
    public final int addTextInputField(int maxLen, int h, boolean password, int x, int font,
                                int y, int _antiTamper, int w, boolean highlighted) {
        // obf: a(int var1, int var2, boolean var3, int var4, int var5, int var6, int var7, int var8, boolean var9)
        // Guard: if (var7 != 14179) return 2; — removed (always 14179 in practice)
        widgetType[nextSlot]    = 5;
        visible[nextSlot]       = true;
        this.passwordMask[nextSlot] = password;   // cb = var3
        activated[nextSlot]     = false;
        fontIndex[nextSlot]     = font;            // k  = var5
        this.highlighted[nextSlot] = highlighted;  // Y  = var9
        widgetX[nextSlot]       = x;               // kb = var4
        widgetY[nextSlot]       = y;               // B  = var6
        widgetW[nextSlot]       = w;               // ob = var8
        widgetH[nextSlot]       = h;               // O  = var2
        this.maxLen[nextSlot]   = maxLen;          // sb = var1
        slotText[nextSlot]      = "";
        return nextSlot++;
    }

    /**
     * Add a password/text-input field variant (type 6) with optional password masking.
     * Differs from type-5 in that it also tracks an explicit width/height bounding box.
     *
     * obf: final int a(int n2, int n3, int n4, boolean bl, int n5, int n6, int n7, boolean bl2, int n8)
     *
     * @param _guard    anti-tamper dummy (dropped; triggers null-clear when != 0)
     * @param maxLen    max characters accepted
     * @param w         field width
     * @param password  true → mask display text
     * @param y         top Y
     * @param font      font index
     * @param h         field height
     * @param maskMode  true → cb[] masking mode (different star rendering)
     * @param x         left X
     * @return slot index
     */
    public final int addPasswordField(int _guard, int maxLen, int w, boolean password,
                               int y, int font, int h, boolean maskMode, int x) {
        // obf: a(int var1, int var2, int var3, boolean var4, int var5, int var6, int var7, boolean var8, int var9)
        // Guard: if (n2 != 0) this.B = null; — removed.
        widgetType[nextSlot]    = 6;
        visible[nextSlot]       = true;
        passwordMask[nextSlot]  = maskMode;
        activated[nextSlot]     = false;
        fontIndex[nextSlot]     = font;
        highlighted[nextSlot]   = password;
        widgetX[nextSlot]       = x;
        widgetY[nextSlot]       = y;
        widgetW[nextSlot]       = w;
        widgetH[nextSlot]       = h;
        this.maxLen[nextSlot]   = maxLen;
        slotText[nextSlot]      = "";
        return nextSlot++;
    }

    /**
     * Add a scrollable list-box (type 4).
     * Allocates per-slot string/int arrays for item labels, colours, secondary strings.
     *
     * Anti-tamper guard {@code n8 != 63} triggers junk helper; removed.
     *
     * obf: final int a(int n2, int n3, int n4, int n5, int n6, int n7, int n8, boolean bl)
     *
     * @param x        left X
     * @param y        top Y
     * @param visRows  number of visible rows
     * @param capacity max item count
     * @param h        height in pixels
     * @param font     font index
     * @param _guard   must equal 63 (anti-tamper removed)
     * @param canSelect true → items are selectable (highlighted colour allowed)
     * @return slot index
     */
    public final int addListBox(int x, int y, int visRows, int capacity, int h,
                         int font, int _guard, boolean canSelect) {
        // obf: a(int var1, int var2, int var3, int var4, int var5, int var6, int var7, boolean var8)
        widgetType[nextSlot]       = 4;
        visible[nextSlot]          = true;
        activated[nextSlot]        = false;
        widgetX[nextSlot]          = visRows;  // obf: kb[eb] = n4 (visRows used as col offset)
        widgetY[nextSlot]          = h;         // obf: B[eb]  = n6
        widgetW[nextSlot]          = x;         // obf: ob[eb] = n2
        widgetH[nextSlot]          = y;         // obf: O[eb]  = n3
        highlighted[nextSlot]      = canSelect;
        fontIndex[nextSlot]        = font;
        maxLen[nextSlot]           = capacity;
        itemCount[nextSlot]        = 0;
        scrollPos[nextSlot]        = 0;
        slotStringArrays[nextSlot] = new String[capacity];
        slotIntArrays[nextSlot]    = new int[capacity];
        slotFeedback[nextSlot]     = new String[capacity];
        slotLabels[nextSlot]       = new String[capacity];
        return nextSlot++;
    }

    /**
     * Add a list-box widget (type 9) — the vertical scrollable variety that tracks
     * selected item and scroll position. Returns 21 early if height < 40 (guard).
     *
     * obf: final int a(int n2, int n3, int n4, boolean bl, int n5, int n6, int n7, int n8)
     *
     * @param x        left X
     * @param w        width
     * @param h        height
     * @param canSelect true → items highlight on hover
     * @param minHeight must be >= 40 or method returns 21 immediately
     * @param capacity  max item capacity
     * @param y        top Y
     * @param font     font index
     * @return slot index, or 21 if minHeight < 40
     */
    public final int addScrollList(int x, int w, int h, boolean canSelect,
                            int minHeight, int capacity, int y, int font) {
        // obf: a(int var1, int var2, int var3, boolean var4, int var5, int var6, int var7, int var8)
        widgetType[nextSlot]       = 9;
        visible[nextSlot]          = true;
        activated[nextSlot]        = false;
        fontIndex[nextSlot]        = font;
        highlighted[nextSlot]      = canSelect;
        widgetX[nextSlot]          = x;
        if (minHeight < 40) {
            return 21;   // guard: list too small to scroll
        }
        widgetY[nextSlot]          = y;
        widgetW[nextSlot]          = w;
        widgetH[nextSlot]          = h;
        maxLen[nextSlot]           = capacity;
        slotStringArrays[nextSlot] = new String[capacity];
        slotIntArrays[nextSlot]    = new int[capacity];
        slotFeedback[nextSlot]     = new String[capacity];
        slotLabels[nextSlot]       = new String[capacity];
        itemCount[nextSlot]        = 0;
        scrollPos[nextSlot]        = 0;
        selectedItem[nextSlot]     = -1;
        hoveredItem[nextSlot]      = -1;
        return nextSlot++;
    }

    /**
     * Add a centred image/sprite area (type 11) centred at (cx, cy).
     *
     * Anti-tamper: if (n5 != 26531) return 59; — removed (always 26531).
     *
     * obf: final int a(int n2, int n3, int n4, int n5, int n6)
     *   n2=h, n3=cx, n4=w, n5=guard, n6=cy
     *
     * @param h      height
     * @param cx     centre X
     * @param w      width
     * @param _guard must be 26531
     * @param cy     centre Y
     * @return slot index, or 59 if guard fails
     */
    public final int addCentredSprite(int h, int cx, int w, int _guard, int cy) {
        // obf: a(int n2, int n3, int n4, int n5, int n6)
        // Guard n5 != 26531 → return 59; removed.
        widgetType[nextSlot] = 11;
        visible[nextSlot]    = true;
        activated[nextSlot]  = false;
        widgetX[nextSlot]    = cx - w / 2;
        widgetY[nextSlot]    = cy - h / 2;
        widgetW[nextSlot]    = w;
        widgetH[nextSlot]    = h;
        return nextSlot++;
    }

    /**
     * Add a native-sprite widget (type 12) whose dimensions are taken directly
     * from Surface arrays {@code surface.kb[spriteId]} and {@code surface.R[spriteId]}.
     *
     * Anti-tamper: if (n5 >= -52) this.d = null; — dead, removed.
     *
     * obf: final int c(int n2, int n3, int n4, int n5)
     *   n2=spriteId, n3=cy, n4=cx, n5=guard
     *
     * @param spriteId index into surface.kb[] / surface.R[]
     * @param cy       centre Y
     * @param cx       centre X
     * @param _guard   anti-tamper (dropped)
     * @return slot index
     */
    public final int addNativeSprite(int spriteId, int cy, int cx, int _guard) {
        // obf: c(int var1, int var2, int var3, int var4)
        int spriteW = surface.spriteWidth[spriteId];  // obf: ua.kb
        int spriteH = surface.spriteHeight[spriteId]; // obf: ua.R
        widgetType[nextSlot] = 12;
        visible[nextSlot]    = true;
        activated[nextSlot]  = false;
        widgetX[nextSlot]    = cx - spriteW / 2;
        widgetY[nextSlot]    = cy - spriteH / 2;
        widgetW[nextSlot]    = spriteW;
        widgetH[nextSlot]    = spriteH;
        fontIndex[nextSlot]  = spriteId;
        return nextSlot++;
    }

    /**
     * Add an icon/image-strip widget (type 8): a row of icon images with labels.
     *
     * obf: (private) a(int n2, int n3, byte by, String[] stringArray, int n4, int n5)
     *   Called from render() for U[i]==8.  Externally added via the type-8 path.
     *
     * NOTE: type-8 widget slots are set up externally by callers that write
     * directly to {@code slotStringArrays[slot]} and set widgetType=8 etc.
     * There is no dedicated addIconStrip() factory; the slot is pre-allocated.
     */
    // (No public addIconStrip — raw slot manipulation used by Mudclient.)

    // =========================================================================
    // Item management (for list widgets types 4, 7, 8, 9)
    // =========================================================================

    /**
     * Append or overwrite an item in a list widget's data arrays.
     * If the list is full (itemCount >= maxLen), the oldest item is dropped
     * from the front (circular buffer / sliding window).
     *
     * obf: final void a(String var1, boolean var2, int var3, String var4, String var5, byte var6, int var7)
     *   var1=label, var2=scrollToEnd, var3=colour, var4=primaryStr, var5=secondaryStr, var6=guard, var7=slot
     *
     * @param label        display label for the item
     * @param scrollToEnd  if true, force scroll position to end (shows latest items)
     * @param colour       packed colour int for this item
     * @param primaryStr   primary data string (stored in slotLabels[][])
     * @param secondaryStr feedback/secondary data (stored in slotFeedback[][])
     * @param _guard       anti-tamper (dropped; sets pb=null when > -39)
     * @param slot         target widget slot index
     */
    public final void addListItem(String label, boolean scrollToEnd, int colour,
                           String primaryStr, String secondaryStr, byte _guard, int slot) {
        // obf: a(String var1, boolean var2, int var3, String var4, String var5, byte var6, int var7)
        int idx = itemCount[slot]++;

        if (idx >= maxLen[slot]) {
            // Buffer full: drop oldest item (shift left)
            itemCount[slot]--;
            idx--;
            for (int i = 0; i < idx; i++) {
                slotStringArrays[slot][i] = slotStringArrays[slot][i + 1];
                slotIntArrays[slot][i]    = slotIntArrays[slot][i + 1];
                slotFeedback[slot][i]     = slotFeedback[slot][i + 1];
                slotLabels[slot][i]       = slotLabels[slot][i + 1];
            }
        }

        slotStringArrays[slot][idx] = label;
        slotIntArrays[slot][idx]    = colour;
        // Anti-tamper: if (var6 > -39) this.pb = null; — dead, removed.
        slotFeedback[slot][idx]     = secondaryStr;
        slotLabels[slot][idx]       = primaryStr;

        if (scrollToEnd) {
            scrollPos[slot] = 999999;  // force to bottom
        }
    }

    /**
     * Set a specific item in a list widget by index (direct write, no shifting).
     * Updates itemCount if the index exceeds the current count.
     *
     * obf: final void a(int n2, String string, int n3, int n4, String string2, String string3, int n5)
     *
     * @param idx           item index
     * @param primaryStr    display label  (stored in slotLabels[][])
     * @param _guard        junk modulo (dropped)
     * @param colour        packed colour int
     * @param secondaryStr  secondary/feedback string
     * @param label         tab/icon label (stored in slotStringArrays[][])
     * @param slot          widget slot index
     */
    public final void setListItem(int idx, String primaryStr, int _guard, int colour,
                           String secondaryStr, String label, int slot) {
        // obf: a(int var1, String var2, int var3, int var4, String var5, String var6, int var7)
        // Junk: int n6 = 78 % ((n3 - -14) / 54); — dead, removed.
        slotStringArrays[slot][idx] = label;
        slotIntArrays[slot][idx]    = colour;
        slotFeedback[slot][idx]     = secondaryStr;
        slotLabels[slot][idx]       = primaryStr;
        if (idx + 1 > itemCount[slot]) {
            itemCount[slot] = idx + 1;
        }
    }

    /**
     * Reset the scroll position and hovered-item index for a list widget.
     * (Note: does NOT clear the item count — see {@link #resetItemCount}.)
     *
     * Anti-tamper: if (n3 != 14) call junk helper d() — removed.
     *
     * obf: final void e(int var1, int var2)
     *
     * @param slot    widget slot index
     * @param _guard  anti-tamper (dropped)
     */
    public final void clearList(int slot, int _guard) {
        // obf: e(int var1, int var2)
        scrollPos[slot]   = 0;   // j[var1] = 0
        hoveredItem[slot] = -1;  // N[var1] = -1
    }

    /**
     * Reset the item count for a list widget to zero (empties the list; the
     * backing item arrays are overwritten lazily on next add).
     *
     * obf: final void c(byte var1, int var2)
     *   Junk: int var3 = 60 / ((19 - var1) / 44); — dead, removed.
     *
     * @param _guard  anti-tamper byte (dropped)
     * @param slot    widget slot index
     */
    public final void resetItemCount(byte _guard, int slot) {
        // obf: c(byte var1, int var2)
        itemCount[slot] = 0;   // pb[var2] = 0
    }

    // =========================================================================
    // Text field / focus management
    // =========================================================================

    /**
     * Set the text content of an input-field slot.
     *
     * Anti-tamper: if (n3 != 27642) return; — removed (always 27642).
     *
     * obf: final void a(int n2, String string, int n3)
     *
     * @param slot  widget slot index
     * @param text  new text value
     * @param _guard must equal 27642 in live code
     */
    public final void setFieldText(int slot, String text, int _guard) {
        // obf: a(int var1, String var2, int var3)
        slotText[slot] = text;
    }

    /**
     * Get the current text content of an input-field slot.
     * Returns "null" (the literal string from W[1]) if the text is null.
     *
     * Anti-tamper: if (n3 != 4) return null; — removed (always 4).
     * Anti-tamper: if (n3 >= -90) this.d = null; — dead, removed.
     *
     * obf: final String g(int n2, int n3)
     *
     * @param slot   widget slot
     * @param _guard must equal 4
     * @return current text or "null" if unset
     */
    public final String getFieldText(int slot, int _guard) {
        // obf: g(int var1, int var2)
        if (slotText[slot] == null) {
            return "null";  // W[1] = "null"
        }
        return slotText[slot];
    }

    /**
     * Set the focused input-field slot.  All keyboard events (via
     * {@link #handleKeyInput(int,int)}) will be routed to this slot.
     *
     * Anti-tamper: if (n3 > -70) call junk b() — removed.
     *
     * obf: final void d(int n2, int n3)
     *
     * @param slot   the slot to focus
     * @param _guard anti-tamper (dropped)
     */
    public final void setFocus(int slot, int _guard) {
        // obf: d(int var1, int var2)
        focusedSlot = slot;
    }

    /**
     * Route a key-press character into the focused input field.
     * Handles backspace (8), Enter/Return (10/13), and printable characters.
     * After typing a non-printable character that is not in the valid-char set,
     * focus cycles to the next type-5 or type-6 slot.
     *
     * Anti-tamper: if (n2 != -12) call junk d() — removed.
     *
     * obf: final void a(int param1, int param2)
     *   (This is the Vineflower-failed method; reconstructed from bytecode.)
     *
     * @param _guard  anti-tamper (dropped; was param1/-12 check)
     * @param keyCode character code of the key pressed
     */
    public final void handleKeyInput(int _guard, int keyCode) {
        // obf: a(int var1=guard, int var2=keyCode)
        // Guard: if (var1 != -12) this.d(-17, 7); — removed.
        // clean: if (-1 != ~var2) { ... }  ↔  proceed when var2 != 0  →  no-key sentinel is 0.
        if (keyCode == 0) return;                // no key
        if (focusedSlot == -1) return;           // clean: ~gb != 0 ↔ gb != -1
        if (slotText[focusedSlot] == null) return;
        if (!visible[focusedSlot]) return;

        int len = slotText[focusedSlot].length();

        // Backspace
        if (keyCode == 8 && len > 0) {
            slotText[focusedSlot] = slotText[focusedSlot].substring(0, len - 1);
        }

        // Enter/Return → mark field as "submitted"
        if ((keyCode == 10 || keyCode == 13) && len > 0) {
            activated[focusedSlot] = true;
        }

        // Append printable chars within capacity
        String validChars = VALID_CHARS;
        if (len < maxLen[focusedSlot]) {
            for (int i = 0; i < validChars.length(); i++) {
                if (validChars.charAt(i) == (char) keyCode) {
                    slotText[focusedSlot] = slotText[focusedSlot] + (char) keyCode;
                    break;
                }
            }
        }

        // Tab key (keyCode == 9): cycle focus to next text-input slot (type 5 or 6)
        // clean: ~var2 == -10  ↔  keyCode == 9
        if (~keyCode == -10) {
            do {
                focusedSlot = (1 + focusedSlot) % nextSlot;
                if (widgetType[focusedSlot] == 5) return;
            } while (widgetType[focusedSlot] != 6);
        }
    }

    // =========================================================================
    // Visibility control
    // =========================================================================

    /**
     * Show a widget slot (set visible flag to true).
     *
     * obf: final void c(int n2, int n3)
     *
     * @param slot   widget slot index
     * @param _guard anti-tamper (dropped; junk call when < 114)
     */
    public final void showWidget(int slot, int _guard) {
        // obf: c(int var1, int var2)
        visible[slot] = true;
    }

    /**
     * Hide a widget slot (set visible flag to false).
     *
     * obf: final void b(byte by, int n2)
     *
     * @param _guard  anti-tamper (dropped; sets colorMidRose = -86 when <= 33)
     * @param slot    widget slot index
     */
    public final void hideWidget(byte _guard, int slot) {
        // obf: b(byte var1, int var2)
        visible[slot] = false;
    }

    // =========================================================================
    // Query helpers
    // =========================================================================

    /**
     * Test whether the given widget slot was activated (clicked/submitted) since
     * the last call, and clear the flag.
     *
     * Anti-tamper: if (by > -95) return true unconditionally — removed (dead branch
     * since the guard param is never > -95 in practice; all real callers use by=-95).
     *
     * obf: final boolean a(byte by, int n2)
     *
     * @param _guard  anti-tamper (dropped)
     * @param slot    widget slot index
     * @return true if the widget was activated, clearing the flag
     */
    public final boolean wasActivated(byte _guard, int slot) {
        // obf: a(byte by, int n2)
        // Anti-tamper: if (by > -95) return true; — removed.
        if (visible[slot] && activated[slot]) {
            activated[slot] = false;
            return true;
        }
        return false;
    }

    /**
     * Return the selected-item index for a list widget.
     *
     * Anti-tamper: if (n2 != 14458) this.t = true; — removed.
     *
     * obf: final int f(int n2, int n3)
     *
     * @param _guard  anti-tamper (dropped)
     * @param slot    widget slot
     * @return selected item index, or -1 if none
     */
    public final int getSelectedItem(int _guard, int slot) {
        // obf: f(int var1, int var2)
        return selectedItem[slot];
    }

    /**
     * Return the hovered-item index for a list widget (most recently moused-over row).
     *
     * Anti-tamper: if (n3 != 17050) this.E = 56; — removed.
     *
     * obf: final int b(int n2, int n3)
     *
     * @param slot    widget slot
     * @param _guard  anti-tamper (dropped)
     * @return hovered item index, or -1 if none
     */
    public final int getHoveredItem(int slot, int _guard) {
        // obf: b(int var1, int var2)
        return hoveredItem[slot];
    }

    /**
     * Return the primary label string for a list item.
     *
     * Anti-tamper: if (n3 >= -90) this.d = null; — removed.
     *
     * obf: final String a(int n2, int n3, int n4)
     *   n2=itemIdx, n3=guard, n4=slot
     *
     * @param itemIdx item index within the list
     * @param _guard  anti-tamper (dropped)
     * @param slot    widget slot
     * @return the primary label string
     */
    public final String getItemLabel(int itemIdx, int _guard, int slot) {
        // obf: a(int n2, int n3, int n4)
        return slotLabels[slot][itemIdx];
    }

    /**
     * Return the feedback/secondary string for a list item.
     *
     * Anti-tamper: if (n3 != 19680) call junk scrollbar helper — removed.
     *
     * obf: final String b(int n2, int n3, int n4)
     *
     * @param itemIdx item index
     * @param _guard  anti-tamper (dropped)
     * @param slot    widget slot
     * @return feedback string
     */
    public final String getItemFeedback(int itemIdx, int _guard, int slot) {
        // obf: b(int var1, int var2, int var3)
        return slotFeedback[slot][itemIdx];
    }

    // =========================================================================
    // Mouse input injection
    // =========================================================================

    /**
     * Inject mouse state into the panel each frame.  Updates mouseX, mouseY,
     * clickMode, mouseButtonState.  If clickMode == 1 (left-click), checks each
     * visible widget for a hit and sets activated[] or toggles selectedItem[].
     *
     * This is the Vineflower-failed method, reconstructed from bytecode.
     *
     * obf: final void b(int n2, int n3, int n4, int n5, int n6)
     *   n2 = mouseButtonState (1=held), n3 = mouseY, n4 = n4 (unused slot, guard -9989),
     *   n5 = clickMode (0=none,1=left,2=right), n6 = mouseX
     *
     * @param mouseButtonState  1 if mouse button currently held, else 0
     * @param mouseY            current mouse Y pixel
     * @param _guard            internal sentinel (equals -9989 in normal call)
     * @param clickMode         0=no click, 1=left-click, 2=right-click
     * @param mouseX            current mouse X pixel
     */
    public final void handleMouseInput(int mouseButtonState, int mouseY, int _guard,
                                int clickMode, int mouseX) {
        // obf: b(int param1, int param2, int param3, int param4, int param5)
        // (Vineflower parse failure; reconstructed from bytecode above.)
        this.mouseY = mouseY;
        this.mouseX = mouseX;

        // clean: if (-1 != ~var4) this.zb = var4;  ↔  update only when clickMode != 0
        if (clickMode != 0) {
            this.clickMode = clickMode;
        }
        this.mouseButtonState = mouseButtonState;

        if (clickMode == 1) {
            // Left-click: scan for clickable widgets
            for (int i = 0; i < nextSlot; i++) {
                if (!visible[i]) continue;

                // Rectangular hit-test widgets (type 10 = ellipsis/progress)
                if (widgetType[i] == 10) {
                    if (widgetX[i] <= mouseX
                     && mouseY >= widgetY[i]
                     && widgetW[i] + widgetX[i] >= mouseX
                     && mouseY <= widgetH[i] + widgetY[i]) {
                        activated[i] = true;
                    }
                }

                // Checkbox/toggle (type 14): toggle selectedItem between 0 and 1
                if (visible[i] && widgetType[i] == 14
                 && mouseX >= widgetX[i]
                 && mouseY >= widgetY[i]
                 && widgetW[i] + widgetX[i] >= mouseX
                 && widgetH[i] + widgetY[i] >= mouseY) {
                    selectedItem[i] = 1 - selectedItem[i];
                }
            }
        }

        // Guard check: if (_guard != -9989) call junk a() — removed.

        // G counter (drag physics): increment while mouseButtonState==1, reset otherwise
        if (mouseButtonState == 1) {
            G++;
        } else {
            G = 0;
        }

        // Drag: if clickMode==1 (held-and-moving) or G > 20, check type-15 hit regions.
        // clean: if (-2 == ~var4 || ~this.G < -21) → (clickMode == 1 || G > 20)
        if (clickMode == 1 || ~G < -21) {
            for (int i = 0; i < nextSlot; i++) {
                if (!visible[i] || widgetType[i] != 15) continue;
                // clean: ~kb >= ~bb (widgetX <= mouseX) && hb >= B && ob+kb >= bb
                //        && ~(B+O) <= ~hb (mouseY <= widgetY+widgetH)
                if (widgetX[i] <= mouseX
                 && mouseY >= widgetY[i]
                 && widgetW[i] + widgetX[i] >= mouseX
                 && ~(widgetY[i] + widgetH[i]) <= ~mouseY) {
                    activated[i] = true;
                }
            }
            // Drain 5 drag ticks per frame when in drag mode
            G -= 5;
        }
    }

    // =========================================================================
    // Render
    // =========================================================================

    /**
     * Render all visible widgets for this frame.
     * Iterates {@code nextSlot} slots, dispatches each to the appropriate private
     * draw helper based on {@link #widgetType}[].
     * Resets {@code clickMode} to 0 at the end of the pass.
     *
     * Anti-tamper: if (param1 > 0) this.colorMidRose = -121; — dead, removed.
     * (This was a param-guarded colour corruption; always benign in callers.)
     *
     * This method failed Vineflower decompilation; the logic below is
     * reconstructed from bytecode.
     *
     * obf: final void a(byte param1)
     */
    public final void render(byte _guard) {
        // obf: a(byte param1)
        // Anti-tamper: if (param1 > 0) this.R = -121; — removed.
        for (int i = 0; i < nextSlot; i++) {
            if (!visible[i]) continue;

            switch (widgetType[i]) {
                case 0:
                    // Plain label: draw text at (widgetX, widgetY) in dim colour
                    // obf: a(yb, B, -65, k, kb, var2, 0) — 7-arg String-first overload
                    drawTextLabelCentred(slotText[i], widgetY[i], (byte) -65, fontIndex[i],
                                  widgetX[i], i, 0);
                    break;
                case 1:
                    // Label-button: draw with white or grey colour, centred
                    // obf: a(yb, B, 117, k, kb - w.a(k,108,yb)/2, var2, 0)
                    drawTextLabelCentred(slotText[i], widgetY[i], (byte) 117, fontIndex[i],
                                  widgetX[i] - surface.textWidth(fontIndex[i], 108, slotText[i]) / 2,
                                  i, 0);
                    break;
                case 2:
                    // Oval/circle
                    drawOval(widgetW[i], (byte) 69, widgetH[i], widgetX[i], widgetY[i]);
                    break;
                case 3:
                    // Plain rectangle
                    drawRect(widgetY[i], 0, widgetX[i], widgetW[i]);
                    break;
                case 4:
                    // Type-4 list: scrollable horizontal menu strip
                    drawHorizListBox(slotIntArrays[i], scrollPos[i], i,
                                     widgetX[i], widgetY[i], widgetW[i],
                                     widgetH[i], itemCount[i],
                                     slotStringArrays[i], fontIndex[i], false);
                    break;
                case 5:
                case 6:
                    // Text input field
                    drawTextField(fontIndex[i], slotText[i], widgetW[i], widgetH[i],
                                  true, widgetX[i], widgetY[i], i);
                    break;
                case 7:
                    // Tab/button strip (horizontal, centred labels)
                    drawTabStrip(widgetX[i], fontIndex[i], (byte) -73,
                                 slotStringArrays[i], i, widgetY[i]);
                    break;
                case 8:
                    // Icon/image widget
                    drawIconWidget(slotStringArrays[i], i, -121, fontIndex[i],
                                   widgetX[i], widgetY[i]);
                    break;
                case 9: {
                    // Type-9 scrollable list (vertical)
                    int u9 = ~widgetType[i];  // == ~9 == -10
                    // dispatch same as the `~U[i] == -10` check:
                    drawScrollList(widgetH[i], widgetX[i], i,
                                   slotStringArrays[i], itemCount[i],
                                   fontIndex[i], widgetY[i],
                                   slotIntArrays[i], 0, scrollPos[i], widgetW[i]);
                    break;
                }
                case 11:
                    // Centred sprite area (draws with scrollbar logic)
                    drawScrollBar(widgetY[i], widgetW[i], true, widgetX[i], widgetH[i]);
                    break;
                case 12:
                    // Native sprite
                    drawSpriteWidget((byte) -82, widgetY[i], widgetX[i], fontIndex[i]);
                    break;
                case 14:
                    // Checkbox/toggle
                    drawCheckbox((byte) 52, widgetX[i], widgetH[i], widgetY[i],
                                 widgetW[i], i);
                    break;
                // type 10, 15 etc. are hit-test only or handled above
                default:
                    break;
            }
        }
        clickMode = 0;
    }

    // =========================================================================
    // Private draw helpers
    // =========================================================================

    /**
     * Draw a text label string at (x, y) using the specified font.
     * Colour is white (0xFFFFFF) if {@link #highlighted}[slot] is true, else 0 (black/grey).
     * If this slot is the currently focused slot, appends '*' to the displayed text.
     *
     * obf: private final void a(int n2, int n3, boolean bl, int n4, int n5, int n6, String string)
     *
     * @param font    font index
     * @param slot    widget slot (for focus/highlight check)
     * @param active  true → widget is "active" (show cursor / special colour)
     * @param colour  base colour override (0 or 0xFFFFFF)
     * @param x       X position
     * @param y       Y position
     * @param text    string to draw
     */
    private void drawTextLabel(int font, int slot, boolean active, int colour,
                                int x, int y, String text) {
        // obf: a(int var1=font, int var2=slot, boolean var3=active, int var4=colour,
        //        int var5=x, int var6=y, String var7=text)
        // [Cb counter removed]
        // The actual draw colour is selected purely from highlighted[slot]
        // (the `colour` param is forwarded as the Surface style/prefix code, not the fill).
        // The dim/bright choice: highlighted[slot] ? white : black.
        int drawColour = highlighted[slot] ? 0xFFFFFF : 0;

        // NOTE: the focus '*' cursor is appended by the CALLERS (drawTextField /
        // drawTextLabelCentred dispatch), NOT here.  Clean source has no '*' logic
        // in this helper.

        // clean: w.a(var4=colour, var6=y, var7=text, var5=x, var8=drawColour, -54, var1=font)
        surface.drawstring(colour, y, text, x, drawColour, (byte) -54, font);
    }

    /**
     * Draw a text label at a computed baseline Y (y + ascent/3).
     * Delegates to {@link #drawTextLabel}.
     *
     * obf: private final void a(String var1, int var2, byte var3, int var4, int var5, int var6, int var7)
     *   (The byte param was an anti-tamper div guard: -109/((by-14)/62), always dead.)
     *
     * @param text    string to draw         (obf var1)
     * @param y       top Y; baseline = y + lineHeight/3  (obf var2)
     * @param _guard  anti-tamper junk division param (dropped) (obf var3)
     * @param font    font index              (obf var4)
     * @param x       X position              (obf var5)
     * @param slot    widget slot index       (obf var6)
     * @param colour  base colour override    (obf var7)
     */
    private void drawTextLabelCentred(String text, int y, byte _guard,
                                      int font, int x, int slot, int colour) {
        // obf: a(String var1, int var2, byte var3, int var4, int var5, int var6, int var7)
        // Junk: int n9 = -109/((by-14)/62); removed.
        // Delegate: a(var4=font, var6=slot, true, var7=colour, var5=x, var8=baselineY, var1=text)
        int baselineY = y + surface.textHeight(508305352, font) / 3;
        drawTextLabel(font, slot, true, colour, x, baselineY, text);
    }

    /**
     * Draw an oval (type 2) on the surface.
     *
     * obf: private final void a(int n2, byte n3, int n4, int n5, int n6)
     *   (The byte param was an anti-tamper guard: if (n3 != 69) l = null)
     *
     * @param w      oval width
     * @param _guard anti-tamper byte (dropped; zeroes fontGlyphData when != 69)
     * @param h      oval height
     * @param x      left X
     * @param y      top Y
     */
    private void drawOval(int w, byte _guard, int h, int x, int y) {
        // obf: a(int n2, byte n3, int n4, int n5, int n6)
        // Draws a tiled panel background + border frame.
        // [b counter removed]
        surface.setBounds(x, x + w, y + h, y, (byte) 30);  // obf: ua.a(I,I,I,I,B) — clip rect
        surface.drawGradientDirect(x, colorShadowEdge, w, colorInnerBg, h, y, 19020);

        // Optional grid-tile overlay (p.d = Timer.legacyFlag = debug/tile-grid flag)
        if (Timer.legacyFlag) {
            // clean: var6 = -(var5 & 63) + var4  → tileX = x - (y & 63)
            int tileX = x - (y & 0x3F);
            while (tileX < w + x) {
                int tileY = -(0x1F & y) + y;   // var7 = y - (y & 31)
                for (int iy = tileY; y + h > iy; iy += 128) {
                    surface.drawLineVert(6 + StringCodec.STATUS_NOT_FOUND, 0, tileX, 128, iy);
                }
                tileX += 128;
            }
        }

        // Border edges
        surface.drawLineHoriz(w, colorInnerBg, x, y, (byte) 111);
        // Anti-tamper: if (_guard != 69) l = null; — removed.

        surface.drawLineHoriz(w - 2, colorInnerBg, 1 + x, y + 1, (byte) -124);
        surface.drawLineHoriz(w - 4, colorMidWarmGrey, 2 + x, 2 + y, (byte) -99);
        surface.drawLineVert(x, y, colorInnerBg, h, 0);
        surface.drawLineVert(1 + x, y + 1, colorInnerBg, h - 2, 0);
        surface.drawLineVert(x + 2, y + 2, colorMidWarmGrey, h - 4, _guard ^ 0x45);
        surface.drawLineHoriz(w, colorShadowEdge, x, y + (h - 1), (byte) 100);
        surface.drawLineHoriz(w - 2, colorShadowEdge, x + 1, y + (h - 2), (byte) 103);
        surface.drawLineHoriz(w - 4, colorInnerShadow, 2 + x, h - 3 + y, (byte) 73);
        surface.drawLineVert(w + x - 1, y, colorShadowEdge, h, 0);
        surface.drawLineVert(w + x - 2, y + 1, colorShadowEdge, h - 2, 0);
        surface.drawLineVert(x + (w - 3), y + 2, colorInnerShadow, h - 4, 0);
        surface.resetBounds(-1);
    }

    /**
     * Draw a simple filled rectangle.
     *
     * obf: private final void a(int n2, int n3, int n4, int n5)
     *   n2=y, n3=0 (colour), n4=x, n5=w
     *
     * @param y       top Y
     * @param colour  fill colour (0 = black)
     * @param x       left X
     * @param w       width
     */
    private void drawRect(int y, int colour, int x, int w) {
        // obf: a(int n2, int n3, int n4, int n5)
        // [Q counter removed]
        surface.drawLineHoriz(w, colour, x, y, (byte) -119);
    }

    /**
     * Draw a sprite widget via Surface.drawSprite (used for type-12 native sprites).
     *
     * obf: private final void b(int n2, int n3, int n4, int n5)
     *   n2=junk (division guard), n3=y, n4=x, n5=spriteId
     *   Junk: int n5 = 71/((n2-67)/32); — removed.
     *
     * @param _guard   anti-tamper (dropped; junk division)
     * @param y        top Y
     * @param x        left X
     * @param spriteId sprite index passed to Surface
     */
    private void drawSpriteWidget(int _guard, int y, int x, int spriteId) {
        // obf: b(int var1, int var2, int var3, int var4)
        surface.drawSprite(-1, spriteId, y, x);
    }

    /**
     * Draw the scrollbar chrome for a scrollable list widget.
     * Renders the scrollbar track, thumb, up/down arrows.
     *
     * obf: private final void a(int n2, int n3, int n4, int n5, int n6, int n7, byte by)
     *   n2=thumbH, n3=thumbY, n4=x, n5=w, n6=h, n7=y, by=guard
     *   Junk: int n9 = -91%((58-by)/33); — removed.
     *
     * @param thumbH   scrollbar thumb height
     * @param thumbY   scrollbar thumb Y offset within track
     * @param x        left X of the panel region
     * @param w        panel width (track width derived from this)
     * @param h        panel height
     * @param y        top Y of the scrollable area
     * @param _guard   anti-tamper byte (dropped)
     */
    private void drawScrollBarChrome(int thumbH, int thumbY, int x, int w,
                                      int h, int y, byte _guard) {
        // obf: a(int n2, int n3, int n4, int n5, int n6, int n7, byte by)
        // [L counter removed]
        // trackX = x + w - 12  (right-aligned 12-px scrollbar track)
        int trackX = x + w - 12;   // var8 = var3 + var4 - 12
        surface.drawBoxEdge(trackX, 12, y, 27785, h, 0);
        surface.drawSprite(-1, StringCodec.STATUS_NOT_FOUND,     y + 1,         1 + trackX);
        surface.drawSprite(-1, StringCodec.STATUS_NOT_FOUND + 1, y + h - 12,    1 + trackX);
        surface.drawLineHoriz(12, 0, trackX, 13 + y, (byte) -49);
        surface.drawLineHoriz(12, 0, trackX, y - 13 + h, (byte) -119);
        // Scrollbar thumb (4th arg is var1 = thumbH, NOT w)
        surface.drawGradientDirect(1 + trackX, colorMidRose, 11, colorDarkRose,
                           h - 27, 14 + y, 19020);
        surface.drawBox(trackX + 3, (byte) -105, colorSandstone,
                           14 + y + thumbY, thumbH, 7);  // obf: ua.a(I,B,I,I,I,I) — filled box
        surface.drawLineVert(trackX + 2, y + thumbY + 14, colorStoneLight, thumbH, 0);
        surface.drawLineVert(trackX + 10, 14 + thumbY + y, colorDarkBrown, thumbH, 0);
    }

    /**
     * Draw a scrollable horizontal item strip (type 4 list — "tab bar" style).
     * Items are drawn left-to-right at the vertical centre.
     *
     * obf: private final void a(int[] nArray, int n2, int n3, int n4, int n5, int n6, int n7, int n8, String[] stringArray, int n9, boolean bl)
     *   nArray=colours, n2=scrollPos, n3=slot, n4=x, n5=y, n6=w(?), n7=h, n8=itemCount, stringArray=labels, n9=font, bl=false
     *
     * @param colours   per-item colour array
     * @param scrollPos current scroll offset (first visible item index)
     * @param slot      widget slot
     * @param x         left X
     * @param y         top Y
     * @param w         width
     * @param h         height
     * @param itemCount number of items
     * @param labels    item label strings
     * @param font      font index
     * @param _unused   always false in render dispatch
     */
    private void drawHorizListBox(int[] colours, int scrollPos, int slot,
                                  int x, int y, int w, int h, int itemCount,
                                  String[] labels, int font, boolean _unused) {
        // obf: a(int[] var1, int var2, int var3, int var4, int var5, int var6, int var7,
        //        int var8, String[] var9, int var10, boolean var11)
        //   var1=colours, var2=scrollPos, var3=slot, var4=x, var5=y, var6=w, var7=h,
        //   var8=itemCount, var9=labels, var10=font, var11=flag
        // [o counter removed]
        int visRows = h / surface.textHeight(508305352, font);  // var12 = visible rows
        if (scrollPos > itemCount - visRows) {
            scrollPos = itemCount - visRows;
        }
        if (scrollPos < 0) {   // clean: ~var2 > -1 ↔ var2 < 0
            scrollPos = 0;
        }
        // if (var11) junk a(null,-61,-120,-114,65,-8) — always false here; removed.
        this.scrollPos[slot] = scrollPos;

        if (itemCount > visRows) {   // clean: if (var8 > var12) → needs scrollbar
            int trackX = x + w - 12;                          // var13
            int thumbH = visRows * (h - 27) / itemCount;       // var14
            if (thumbH < 6) thumbH = 6;
            // var15 thumbY computed below (intermediate value discarded by clean too)

            // Up/down arrow click: mouseButtonState==1 and mouseX within [trackX, trackX+12]
            if (mouseButtonState == 1 && mouseX >= trackX && mouseX <= trackX + 12) {
                // mouseY in (y, y+12): scroll up
                if (mouseY > y && mouseY < y + 12 && scrollPos > 0) {
                    scrollPos--;
                }
                // mouseY in (y+h-12, y+h): scroll down
                if (y + h - 12 < mouseY && mouseY < y + h && scrollPos < itemCount - visRows) {
                    scrollPos++;
                }
                this.scrollPos[slot] = scrollPos;
            }

            // Thumb-drag scrolling (dragging[slot] = obf d[var3])
            // clean label194: drag active iff mouseButtonState==1 and mouseX in the thumb
            // column band; otherwise clear dragging.
            if (mouseButtonState == 1
             && ((mouseX >= trackX && mouseX <= trackX + 12)
                 || (mouseX >= trackX - 12 && mouseX <= trackX + 24 && dragging[slot]))) {
                // mouseY in (y+12, y+h-12): live drag
                if (mouseY > y + 12 && mouseY < y + h - 12) {
                    dragging[slot] = true;
                    int relY = mouseY - y - 12 - thumbH / 2;   // var16
                    scrollPos = relY * itemCount / (h - 24);
                    if (scrollPos > itemCount - visRows) scrollPos = itemCount - visRows;
                    if (scrollPos < 0) scrollPos = 0;
                    this.scrollPos[slot] = scrollPos;
                }
            } else {
                dragging[slot] = false;
            }

            int thumbY = scrollPos * (h - 27 - thumbH) / (itemCount - visRows);   // var15
            drawScrollBarChrome(thumbH, thumbY, x, w, h, y, (byte) 113);
        }

        // Item rendering
        int remainH = h - surface.textHeight(508305352, font) * visRows;  // var19
        int startY  = y + 5 * surface.textHeight(508305352, font) / 6 + remainH / 2;  // var20

        for (int row = scrollPos; row < itemCount; row++) {
            // Mouse hover/click detection
            // clean: zb!=0 && x+2<=bb && bb<=strW+x+2 && startY>=mouseY-2
            //        && mouseY-2 > startY-lineHeight
            if (clickMode != 0
             && x + 2 <= mouseX
             && mouseX <= surface.textWidth(font, 97, labels[row]) + x + 2
             && startY >= mouseY - 2
             && mouseY - 2 > startY - surface.textHeight(508305352, font)) {
                activated[slot]   = true;
                selectedItem[slot] = CacheFile.or(clickMode << 16, row);
            }

            // Draw item text
            drawTextLabel(font, slot, true, colours[row], 2 + x, startY, labels[row]);

            startY += surface.textHeight(508305352, font) - SpriteScaler.lineHeightOverride;
            if (startY >= y + h) return;
        }
    }

    /**
     * Draw a vertical scrollable list (type 9 / 4 variant).
     * Items are drawn top-to-bottom; selected item is highlighted red.
     *
     * obf: private final void a(int n2, int n3, int n4, String[] stringArray, int n5, int n6, int n7, int[] nArray, int n8, int n9, int n10)
     *
     * @param h         total height in pixels
     * @param x         left X
     * @param slot      widget slot
     * @param labels    item strings
     * @param itemCount total item count
     * @param font      font index
     * @param y         top Y
     * @param colours   per-item colours
     * @param _n8       unused (scrollbar mode selector, always 0 from render)
     * @param scrollPos current top-row offset
     * @param w         width
     */
    private void drawScrollList(int h, int x, int slot, String[] labels,
                                 int itemCount, int font, int y,
                                 int[] colours, int _n8, int scrollPos, int w) {
        // obf: a(int var1=h, int var2=x, int var3=slot, String[] var4=labels, int var5=itemCount,
        //        int var6=font, int var7=y, int[] var8=colours, int var9=flag, int var10=scrollPos, int var11=w)
        // [db counter removed]
        // if (var9 != 0) junk a(56,-19); — always 0 here; removed.
        int visRows = h / surface.textHeight(508305352, font);  // var12

        if (visRows >= itemCount) {
            // All items fit (clean: ~var12 <= ~var5 ↔ var12 >= var5) — reset scroll
            scrollPos = 0;
            this.scrollPos[slot] = 0;
        } else {
            // Scrollbar geometry
            int trackX = x + w - 12;                         // var13
            int thumbH = visRows * (h - 27) / itemCount;      // var14
            if (thumbH < 6) thumbH = 6;                       // clean: ~var14 > -7 ↔ var14 < 6

            // Up/down arrow click: mouseButtonState==1, mouseX in [trackX, trackX+12]
            if (mouseButtonState == 1 && mouseX >= trackX && mouseX <= trackX + 12) {
                // mouseY in (y, y+12): up
                if (y < mouseY && mouseY < y + 12 && scrollPos > 0) scrollPos--;
                // mouseY in (y+h-12, y+h): down
                if (mouseY > y + h - 12 && mouseY < y + h && scrollPos < itemCount - visRows) scrollPos++;
                this.scrollPos[slot] = scrollPos;
            }

            // Thumb-drag (clean label198). Active iff mouseButtonState==1 AND
            // (mouseX in [trackX,trackX+12]  OR  mouseX in [trackX-12,trackX+24] && dragging).
            if (mouseButtonState == 1
             && ((mouseX >= trackX && mouseX <= trackX + 12)
                 || (mouseX >= trackX - 12 && mouseX <= trackX + 24 && dragging[slot]))) {
                // Only drag while mouseY strictly inside (y+12, y+h-12); else leave dragging as-is
                if (mouseY > y + 12 && mouseY < y + h - 12) {
                    dragging[slot] = true;
                    int relY  = mouseY - y - thumbH / 2 - 12;        // var16
                    scrollPos = itemCount * relY / (h - 24);          // var5*var16/(var1-24)
                    if (scrollPos < 0) scrollPos = 0;                 // clean: -1 < ~var10 ↔ var10 < 0
                    if (scrollPos > itemCount - visRows) scrollPos = itemCount - visRows;
                    this.scrollPos[slot] = scrollPos;
                }
            } else {
                dragging[slot] = false;
            }

            int thumbY = (h - thumbH - 27) * scrollPos / (itemCount - visRows);  // var15
            drawScrollBarChrome(thumbH, thumbY, x, w, h, y, (byte) -25);
        }

        // Item rendering
        hoveredItem[slot] = -1;   // clean: N[var3] = -1
        int remainH = h - surface.textHeight(508305352, font) * visRows;  // var19
        int startY  = y + remainH / 2 + 5 * surface.textHeight(508305352, font) / 6;  // var20

        for (int row = scrollPos; row < itemCount; row++) {
            int colour = highlighted[slot] ? 0xFFFFFF : 0;

            // Mouse hover
            if (mouseX >= 2 + x
             && mouseX <= surface.textWidth(font, 95, labels[row]) + 2 + x
             && mouseY - 2 <= startY
             && mouseY - 2 > startY - surface.textHeight(508305352, font)) {
                if (highlighted[slot]) colour = 0x808080;
                else colour = 0xFFFFFF;
                hoveredItem[slot] = row;
                if (clickMode == 1) {
                    selectedItem[slot] = row;
                    activated[slot]    = true;
                }
            }

            // Selected item highlight
            if (selectedItem[slot] == row && activeSelectionHighlight) {
                colour = 0xFF0000;  // red = active selection
            }

            surface.drawstring(colours[row], startY, labels[row], x + 2, colour, (byte) -9, font);
            startY += surface.textHeight(508305352, font);
            if (startY >= y + h) return;
        }
    }

    /**
     * Draw a tab-strip (type 7): a horizontal row of text labels separated by
     * two-space gaps. Labels are centred above Y. The selected tab is highlighted.
     *
     * obf: private final void a(int n2, int n3, byte by, String[] stringArray, int n4, int n5)
     *   by must equal -73 or method returns early (anti-tamper guard).
     *
     * @param cx       centre X for the whole strip
     * @param font     font index
     * @param _guard   must be -73 (anti-tamper; early return if != -73) — kept as pass-through guard
     * @param tabs     tab label strings
     * @param slot     widget slot
     * @param y        Y position (baseline)
     */
    private void drawTabStrip(int cx, int font, byte _guard, String[] tabs,
                               int slot, int y) {
        // obf: a(int n2, int n3, byte by, String[] stringArray, int n4, int n5)
        // [M counter removed]
        if (_guard != -73) return;  // anti-tamper guard kept (always -73 in practice)

        // Compute total width
        int totalW = 0;
        for (int i = 0; i < tabs.length; i++) {
            totalW += surface.textWidth(font, 106, tabs[i]);
            if (i < tabs.length - 1) {
                totalW += surface.textWidth(font, 92, TWO_SPACES);
            }
        }

        int drawX  = cx - totalW / 2;
        int lineH  = surface.textHeight((byte) -73 + 508305425, font) / 3;
        int baseY  = y + lineH;

        for (int i = 0; i < tabs.length; i++) {
            int colour = highlighted[slot] ? 0xFFFFFF : 0;

            // Hover detection
            if (mouseX >= drawX && mouseX <= drawX + surface.textWidth(font, 73, tabs[i])
             && mouseY <= baseY && mouseY > baseY - surface.textHeight(508305352, font)) {
                if (highlighted[slot]) colour = 0x808080;
                else colour = 0xFFFFFF;
                if (clickMode == 1) {
                    selectedItem[slot] = i;
                    activated[slot]    = true;
                }
            }

            // Selected tab highlight
            if (selectedItem[slot] == i) {
                if (highlighted[slot]) colour = 0xFF0000;   // red = selected, enabled
                else colour = 0xC00000;                      // dark red = selected, disabled
            }

            surface.drawstring(0, baseY, tabs[i], drawX, colour, (byte) -53, font);
            drawX += surface.textWidth(font, 127, tabs[i] + TWO_SPACES);
        }
    }

    /**
     * Draw a type-8 widget: a vertical list of horizontally-centred text labels
     * (no scrollbar). Each label is centred about {@code cx}; the whole stack is
     * vertically centred about {@code cy}. Hover highlights the row; left-click
     * selects it. The selected row is drawn red (highlighted) or dark-red.
     *
     * obf: private final void a(String[] var1, int var2, int var3, int var4, int var5, int var6)
     *   var1=labels, var2=slot, var3=guard(junk -121), var4=font, var5=cx, var6=cy
     *   Dispatched from render for U[i]==8 as a(Ab, var2, -121, k, kb, B).
     *
     * @param labels   item label strings
     * @param slot     widget slot
     * @param _guard   junk modulo param (dropped; -121 in dispatch)
     * @param font     font index
     * @param cx       centre X (each label centred on this)
     * @param cy       centre Y of the whole stack
     */
    private void drawIconWidget(String[] labels, int slot, int _guard,
                                 int font, int cx, int cy) {
        // obf: a(String[] var1, int var2, int var3, int var4, int var5, int var6)
        // [u counter removed]; junk `80 % ((var3+55)/61)` removed.
        int count = labels.length;   // var7
        // var9 startY: top of vertically centred stack
        int startY = cy - (count - 1) * surface.textHeight(508305352, font) / 2;

        for (int row = 0; row < count; row++) {   // clean: ~var10 > ~var7 ↔ var10 < var7
            int colour = highlighted[slot] ? 0xFFFFFF : 0;   // var11
            int strW   = surface.textWidth(font, 112, labels[row]);  // var12

            // Hover: mouseX in [cx-strW/2, cx+strW/2] and mouseY-2 in (startY-lineH, startY]
            if (mouseX >= cx - strW / 2
             && mouseX <= cx + strW / 2
             && mouseY - 2 <= startY
             && mouseY - 2 > startY - surface.textHeight(508305352, font)) {
                colour = highlighted[slot] ? 0x808080 : 0xFFFFFF;
                if (clickMode == 1) {
                    selectedItem[slot] = row;
                    activated[slot]    = true;
                }
            }

            // Selected row highlight
            if (selectedItem[slot] == row) {   // clean: ~vb[var2] == ~var10
                colour = highlighted[slot] ? 0xFF0000 : 0xC00000;
            }

            // clean: w.a(0, var9, var1[var10], var5-var12/2, var11, -126, var4)
            surface.drawstring(0, startY, labels[row], cx - strW / 2, colour, (byte) -126, font);
            startY += surface.textHeight(508305352, font);
        }
    }

    /**
     * Draw a scrollbar thumb track (used for type-11 centred sprite / scroll area).
     *
     * obf: private final void a(int n2, int n3, boolean bl, int n4, int n5)
     *   n2=w, n3=h, bl=bordered, n4=x, n5=y
     *
     * @param w        width
     * @param h        height
     * @param bordered true → draw inner border rectangles
     * @param x        left X
     * @param y        top Y
     */
    private void drawScrollBar(int w, int h, boolean bordered, int x, int y) {
        // obf: a(int n2, int n3, boolean bl, int n4, int n5)
        // [n counter removed]
        surface.drawBox(x, (byte) -127, 0, w, y, h);
        surface.drawBoxEdge(x, h, w, 27785, y, colorTan);

        if (bordered) {
            surface.drawBoxEdge(1 + x, h - 2, 1 + w, 27785, y - 2, colorDarkStone);
            surface.drawBoxEdge(x + 2, h - 4, 2 + w, 27785, y - 4, colorButtonRecess);
            // Corner accent lines (u.g = StringCodec.STATUS_NOT_FOUND)
            surface.drawSprite(-1, StringCodec.STATUS_NOT_FOUND + 2, w, x);
            surface.drawSprite(-1, 3 + StringCodec.STATUS_NOT_FOUND, w, h - 7 + x);
            // clean: w.b(-1, 4+u.g, var1-(-var5- -7), var4) = (..., w + y - 7, x)
            surface.drawSprite(-1, 4 + StringCodec.STATUS_NOT_FOUND, w + y - 7, x);
            surface.drawSprite(-1, 5 + StringCodec.STATUS_NOT_FOUND, y - 7 + w, x - 7 + h);
        }
    }

    /**
     * Draw a checkbox/toggle (type 14): a filled rectangle with a border and,
     * if {@code selectedItem[slot] == 1}, a diagonal cross drawn inside.
     *
     * obf: private final void a(byte by, int n2, int n3, int n4, int n5, int n6)
     *   by must be 52 to draw the inner cross; otherwise only the base rect is drawn.
     *   Anti-tamper: `by ^ 0x34` and `var1 ^ 52` both evaluate to 0 when by==52.
     *
     * @param _must52  must equal 52 (byte) to draw toggle cross (anti-tamper sentinel)
     * @param x        left X
     * @param h        height
     * @param y        top Y
     * @param w        width
     * @param slot     widget slot
     */
    private void drawCheckbox(byte _must52, int x, int h, int y, int w, int slot) {
        // obf: a(byte var1=guard, int var2=x, int var3=h, int var4=y, int var5=w, int var6=slot)
        // [Db counter removed]
        // clean: w.a(var2=x, 100, 0xFFFFFF, var4=y, var3=h, var5=w)
        surface.drawBox(x, (byte) 100, 0xFFFFFF, y, h, w);
        if (_must52 != 52) return;  // sentinel guard preserved (var1 ^ 52 == 0)

        surface.drawLineHoriz(w, colorInnerBg, x, y, (byte) -116);       // w.b(var5, tb, var2, var4, -116)
        surface.drawLineVert(x, y, colorInnerBg, h, 0);                 // w.b(var2, var4, tb, var3, 0)
        surface.drawLineHoriz(w, colorShadowEdge, x, h + y - 1, (byte) -124);  // w.b(var5, J, var2, var3+var4-1, -124)
        surface.drawLineVert(w + x - 1, y, colorShadowEdge, h, 0);      // w.b(var5+var2-1, var4, J, var3, 0)

        // If checked (selectedItem==1), draw inner diagonal cross lines
        if (selectedItem[slot] == 1) {
            for (int d = 0; d < h; d++) {   // clean: ~var3 < ~var7 ↔ var7 < var3 (d < h)
                surface.drawLineHoriz(1, 0, d + x, y + d, (byte) 88);             // w.b(1,0,var7+var2,var4+var7,88)
                surface.drawLineHoriz(1, 0, w + x - 1 - d, d + y, (byte) 106);    // w.b(1,0,var5+var2-1-var7,var7+var4,106)
            }
        }
    }

    /**
     * Draw a text input field (types 5 and 6).
     * If password-masked ({@code passwordMask[slot]}), replaces text with 'X' characters.
     * Hover/click detection updates focusedSlot.
     *
     * obf: private final void a(int n2, String string, int n3, int n4, boolean bl, int n5, int n6, int n7)
     *   (Vineflower parse failure; reconstructed from bytecode.)
     *
     * @param font     font index
     * @param text     current text content
     * @param w        field width
     * @param h        field height
     * @param active   true → field is interactable (always true from render dispatch)
     * @param x        left X
     * @param y        top Y (vertical centre)
     * @param slot     widget slot
     */
    private void drawTextField(int font, String text, int w, int h,
                                boolean active, int x, int y, int slot) {
        // obf: a(int n2, String string, int n3, int n4, boolean bl, int n5, int n6, int n7)
        // [wb counter removed]

        // Password masking: replace each character with 'X' (clean: if (this.cb[var8]) ...)
        if (passwordMask[slot]) {
            int len = text.length();
            text = "";
            for (int i = 0; i < len; i++) {
                text = text + "X";
            }
        }

        // Focus detection for type-5 (text-input) fields.
        // clean focus region (negation of the break test): zb==1 && x<=bb && mouseY>=y-h/2
        //   && bb<=x+w && mouseY<=y+h/2
        if (~widgetType[slot] == -6) {  // widgetType == 5
            if (clickMode == 1
             && x <= mouseX
             && mouseY >= y - h / 2
             && mouseX <= x + w
             && mouseY <= y + h / 2) {
                focusedSlot = slot;
            }
        }

        // Focus detection for type-6 fields (button-text style)
        if (~widgetType[slot] == -7) {  // widgetType == 6
            if (clickMode == 1
             && ~mouseX <= ~(x - w / 2)
             && ~(-(h / 2) + y) >= ~mouseY
             && w / 2 + x >= mouseX
             && ~(y + h / 2) <= ~mouseY) {
                focusedSlot = slot;
            }
            // Centre text horizontally for type-6
            x -= surface.textWidth(font, 76, text) / 2;
        }

        // If this slot is focused, append '*' cursor
        if (focusedSlot == slot) {
            text = text + "*";
        }

        // Draw optional box border if not active
        if (!active) {
            drawOval(28, (byte) 94, -2, 23, 126);  // junk sentinel draw; always active in practice
        }

        int baselineY = y + surface.textHeight(508305352, font) / 3;
        drawTextLabel(font, slot, active, 0, x, baselineY, text);
    }

    // =========================================================================
    // Font loading
    // =========================================================================

    /**
     * Load and rasterize a Helvetica font into the global Surface font table.
     *
     * The font-spec string format (decoded from XOR pool):
     *   {@code "helvetica<N>[b][p]"} or {@code "helvetica<N>p"} or {@code "helvetica<N>b"}
     *   where N is the point size, 'b' = bold, 'p' = plain, 'd' = italic.
     *   Suffix ".jf" is stripped if present.
     *
     * Iterates {@link #VALID_CHARS} (95 characters) and calls {@link FontBuilder#a} for each,
     * writing glyph bitmaps into {@code SocketFactory.b[fontSlot][]} (the byte[50][] glyph table).
     * Then copies glyph data from {@code GameFrame.unusedByteBuffer[]} (pixel buffer) into that slot.
     *
     * After loading a bold font, recursively loads its plain ("f<N>p") variant.
     * After loading a font with flag {@code wantDouble}, recursively loads its
     * italic/double ("d<N>p") variant.
     *
     * obf: static final boolean a(e param0, String param1, int param2, int param3)
     *   param0 = applet shell (for getFontMetrics), param1 = fontSpec, param2 = fontSlot, param3 = srcOffset
     *
     * @param shell      GameShell applet (used to call getFontMetrics)
     * @param fontSpec   font descriptor string, e.g. "helvetica12b"
     * @param fontSlot   index into the font glyph table (SocketFactory.b[])
     * @param srcOffset  byte offset into GameFrame.unusedByteBuffer[] for glyph copy
     * @return true on success, false if any glyph rasterization fails
     */
    public static final boolean loadFont(GameShell shell, String fontSpec,
                                  int fontSlot, int srcOffset) {
        // obf: static final boolean a(e param0, String param1, int param2, int param3)
        // [Bb counter removed]
        boolean wantDouble = false;
        fontSpec = fontSpec.toLowerCase();
        boolean isBold = false;

        // Strip "helvetica" prefix variants
        if (fontSpec.startsWith(FONT_PREFIX)) {  // "helvetica"
            fontSpec = fontSpec.substring(9);
        }
        if (fontSpec.startsWith("h")) {
            fontSpec = fontSpec.substring(1);
        }
        if (fontSpec.startsWith("f")) {
            fontSpec = fontSpec.substring(1);
            wantDouble = true;
        }
        if (fontSpec.startsWith("d")) {
            fontSpec = fontSpec.substring(1);
            isBold = true;  // 'd' = italic/bold variant flag
        }
        if (fontSpec.endsWith(FONT_SUFFIX)) {  // ".jf"
            fontSpec = fontSpec.substring(0, fontSpec.length() - 3);
        }

        int style = 0;  // Font.PLAIN
        if (fontSpec.endsWith("b")) {
            style = 1;  // Font.BOLD
            fontSpec = fontSpec.substring(0, fontSpec.length() - 1);
        }
        if (fontSpec.endsWith("p")) {
            fontSpec = fontSpec.substring(0, fontSpec.length() - 1);
        }

        int pointSize = Integer.parseInt(fontSpec);
        Font font = new Font(FONT_NAME_HELVETICA, style, pointSize);
        FontMetrics metrics = shell.getFontMetrics(font);

        // Rasterize each glyph in VALID_CHARS (95 chars; clean loops while var11 < 95)
        Packet.writePos = 855;  // obf: b.c — reset glyph buffer size counter
        for (int i = 0; i < 95; i++) {   // clean: -96 < ~var11 ↔ var11 < 95
            if (!FontBuilder.rasterizeGlyph(fontSlot, font, i, -95, shell,
                               VALID_CHARS.charAt(i), metrics, isBold)) {
                return false;
            }
        }

        // Allocate the glyph byte array for this font slot
        // obf: m.b[var2] = new byte[b.c];  (m = SocketFactory, byte[50][] glyph table)
        SocketFactory.fontGlyphData[fontSlot] = new byte[Packet.writePos];

        // Copy glyph data from GameFrame's pixel buffer
        // obf: m.b[var2][var11] = qb.k[var11]  (m = SocketFactory, qb = GameFrame)
        for (int j = srcOffset; j < Packet.writePos; j++) {
            SocketFactory.fontGlyphData[fontSlot][j] = GameFrame.unusedByteBuffer[j];
        }

        // If bold style, recursively load the plain variant
        if (style == 1 && SurfaceImageProducer.k[fontSlot]) {
            SurfaceImageProducer.k[fontSlot] = false;
            if (!loadFont(shell, "f" + pointSize + "p", fontSlot, 0)) return false;
        }

        // If wantDouble, recursively load the italic/double-size variant
        if (wantDouble) {
            if (SurfaceImageProducer.k[fontSlot]) return true;
            SurfaceImageProducer.k[fontSlot] = false;
            if (!loadFont(shell, "d" + pointSize + "p", fontSlot, srcOffset)) return false;
        }

        return true;
    }

    // =========================================================================
    // XOR string-pool decoders (static utility; used only in static initializer)
    // =========================================================================

    /**
     * First-pass XOR: convert String → char[] and XOR byte 0 with 0x71 if length < 2.
     * Part of the two-pass XOR string pool used for obfuscated string literals.
     *
     * obf: private static char[] z(String var0)
     */
    private static char[] z(String s) {
        char[] arr = s.toCharArray();
        if (arr.length < 2) {
            arr[0] = (char)(arr[0] ^ 0x71);
        }
        return arr;
    }

    /**
     * Second-pass XOR: decode char[] with 5-byte rotating key {117,40,10,9,113} → String.
     * Part of the two-pass XOR string pool.
     *
     * obf: private static String z(char[] var0)
     */
    private static String z(char[] arr) {
        int[] key = {117, 40, 10, 9, 113};
        for (int i = 0; i < arr.length; i++) {
            arr[i] = (char)(arr[i] ^ key[i % 5]);
        }
        return new String(arr).intern();
    }

    /*
     * Decoded XOR string pool (W[]):
     *   W[0]  = "qa.J("           — error context for addCentredSprite
     *   W[1]  = "null"            — null placeholder in error strings
     *   W[2]  = "{...}"           — non-null placeholder in error strings
     *   W[3]  = "qa.KA("          — drawTextField error context
     *   W[4]  = "qa.B("           — addOval error context
     *   W[5]  = "qa.QA("          — packRgb error context
     *   W[6]  = "qa.H("           — handleKeyInput error context
     *   W[7]  = VALID_CHARS       — "ABCDEFGHIJKLMNOPQRSTUVWXYZabc...| "
     *   W[8]  = "qa.F("           — drawOval error context
     *   W[9]  = "qa.JA("          — addListItem error context
     *   W[10] = "qa.V("           — drawScrollBar error context
     *   W[11] = "qa.HA("          — render error context
     *   W[12] = "qa.O("           — handleMouseInput error context
     *   W[13] = "qa.G("           — addProgressWidget error context
     *   W[14] = "qa.PA("          — addNativeSprite error context
     *   W[15] = "qa.IA("          — drawTextLabel (private) error context
     *   W[16] = "qa.L("           — addTextInputField error context
     *   W[17] = "qa.EA("          — hideWidget error context
     *   W[18] = "qa.NA("          — showWidget error context
     *   W[19] = "qa.DA("          — drawTabStrip error context
     *   W[20] = "  "              — two-space separator (TWO_SPACES)
     *   W[21] = "qa.U("           — addPasswordField error context
     *   W[22] = "qa.<init>("      — constructor error context
     *   W[23] = "qa.C("           — getHoveredItem error context
     *   W[24] = "qa.FA("          — drawHorizListBox (private) error context
     *   W[25] = "qa.W("           — wasActivated error context
     *   W[26] = "qa.Q("           — setListItem error context
     *   W[27] = "qa.CA("          — getSelectedItem error context
     *   W[28] = "qa.S("           — addLabel error context
     *   W[29] = ".jf"             — font file suffix (FONT_SUFFIX)
     *   W[30] = "qa.GA("          — loadFont error context
     *   W[31] = "Helvetica"       — font name (FONT_NAME_HELVETICA)
     *   W[32] = "helvetica"       — font name prefix (FONT_PREFIX)
     *   W[33] = "qa.N("           — clearList error context
     *   W[34] = "qa.R("           — drawTextLabelCentred error context
     *   W[35] = "qa.D("           — drawScrollList error context
     *   W[36] = "qa.I("           — getItemLabel error context
     *   W[37] = "qa.M("           — drawHorizListBox variant error context
     *   W[38] = "qa.RA("          — drawRect error context
     *   W[39] = "qa.K("           — drawCheckbox error context
     *   W[40] = "qa.T("           — getItemFeedback error context
     *   W[41] = "qa.BA("          — getFieldText error context
     *   W[42] = "qa.OA("          — setFocus error context
     *   W[43] = "qa.AA("          — addScrollList error context
     *   W[44] = "qa.LA("          — drawSpriteWidget error context
     *   W[45] = "qa.E("           — addListBox error context
     *   W[46] = "qa.A("           — drawScrollBarChrome error context
     *   W[47] = "qa.P("           — clearList (c(byte,int)) error context
     *   W[48] = "qa.MA("          — setFieldText error context
     */
}
