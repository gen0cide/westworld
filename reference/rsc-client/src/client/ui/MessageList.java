package client.ui;

// Best-effort imports (will not resolve until classes are re-packaged):
import client.scene.SurfaceSprite; // obf: ba  — renderer passed in constructor
import client.data.EntityDef;      // obf: t   — per-entry record struct
import client.data.CacheUpdater;   // obf: cb  — supplies sort helper
import client.util.ErrorHandler;   // obf: i   — rethrow wrapper

/**
 * MessageList — a sorted, drawable list of (prefix, message) string pairs.
 *
 * <p>Three instances exist in {@link client.Mudclient} (chat tab, social/friends
 * tab, and a third messaging context).  Each instance owns a fixed-capacity
 * array of {@link EntityDef} entry records; entries are sorted by their
 * {@code sortKey} field (ascending) before each render.
 *
 * <p>The class has two main responsibilities:
 * <ol>
 *   <li><b>Entry management</b> — {@link #addEntry} appends a new
 *       (prefix + message) pair, growing the backing array as needed, and
 *       recalculates the cached display dimensions.</li>
 *   <li><b>Hit-testing / rendering</b> — {@link #draw} / {@link #hitTest}
 *       iterate all entries using the surface's text-metric helpers and either
 *       draw them with {@link SurfaceSprite} or return the index of the entry
 *       under the mouse cursor.</li>
 * </ol>
 *
 * <p>A single optional header string ({@link #headerText}) may be rendered
 * above the entry list; when non-null it contributes to {@link #panelWidth}.
 *
 * <p>The static {@link #drawTextureSpan} method is a stand-alone affine
 * texture-span filler (perspective-correct, 16-pixel unrolled) used
 * elsewhere in the renderer; it has no logical connection to the message list
 * but lives in this class in the obfuscated build.
 *
 * <p>obf: {@code wb}
 */
public final class MessageList {

    // -----------------------------------------------------------------------
    // Static profiling counters (obfuscation artifact — never read by logic)
    // Each counter was incremented at method entry for profiling; retained as
    // dead fields to preserve field layout for future cross-referencing.
    // -----------------------------------------------------------------------

    /** obf: wb.B  — profiling counter for addEntryFull */
    public static int _profileB;
    /** obf: wb.C  — profiling counter for getEntryColor */
    public static int _profileC;
    /** obf: wb.z  — profiling counter for addEntryWithFont */
    public static int _profileZ;
    /** obf: wb.H  — profiling counter for hitTest */
    public static int _profileH;
    /** obf: wb.d  — profiling counter for getEntrySprite */
    public static int _profileD;
    /** obf: wb.a  — profiling counter for draw (private) */
    public static int _profileA;
    /** obf: wb.o  — profiling counter for sortEntries */
    public static int _profileO;
    /** obf: wb.K  — profiling counter for setCount */
    public static int _profileK;
    /** obf: wb.w  — (long) profiling counter for addEntryColored */
    public static long _profileW;
    /** obf: wb.l  — profiling counter for getEntryLayer */
    public static int _profileL_stat;
    /** obf: wb.u  — profiling counter for addEntryScrolled */
    public static int _profileU;
    /** obf: wb.x  — profiling counter for getEntryMessageColor */
    public static int _profileX;
    /** obf: wb.c  — profiling counter for recalcDimensions */
    public static int _profileC2;
    /** obf: wb.F  — profiling counter for removeEntry */
    public static int _profileF;
    /** obf: wb.f  — profiling counter for getCount */
    public static int _profileF2;
    /** obf: wb.j  — profiling counter for getPanelHeight */
    public static int _profileJ;
    /** obf: wb.v  — profiling counter for addEntryWithColor */
    public static int _profileV;
    /** obf: wb.y  — profiling counter for getEntryColorCode */
    public static int _profileY;
    /** obf: wb.M  — profiling counter for hitTestNoRender */
    public static int _profileM;
    /** obf: wb.s  — profiling counter for getPanelWidth */
    public static int _profileS;
    /** obf: wb.E  — profiling counter for getEntryMessage */
    public static int _profileE;
    /** obf: wb.r  — profiling counter for addEntryRich */
    public static int _profileR;
    /** obf: wb.e  — profiling counter for addEntrySimple */
    public static int _profileE2;
    /** obf: wb.J  — profiling counter for getEntryPrefix */
    public static int _profileJ2;
    /** obf: wb.L  — profiling counter for addEntry (private) */
    public static int _profileL;
    /** obf: wb.G  — profiling counter for addEntryGuarded */
    public static int _profileG;
    /** obf: wb.h  — profiling counter for getEntryExtra */
    public static int _profileH2;
    /** obf: wb.g  — profiling counter for drawTextureSpan (static) */
    public static int _profileG2;
    /** obf: wb.k  — profiling counter for getEntryName */
    public static int _profileK2;

    // -----------------------------------------------------------------------
    // Instance fields
    // -----------------------------------------------------------------------

    /**
     * Calculated rendered height of the panel in pixels — sum of one
     * {@code textHeight} unit per entry (plus header row if present).
     * obf: {@code I}
     */
    private int panelHeight;

    /**
     * The surface/sprite context used for all drawing and text-metric calls.
     * obf: {@code t}  (field named 't' in obf class; type {@code ba} = SurfaceSprite)
     */
    private SurfaceSprite surface;     // obf: t (field), ba (type)

    /**
     * Number of live entries in {@link #entries}.
     * obf: {@code n}
     */
    private int count;

    /**
     * Calculated rendered width of the panel in pixels — max of all
     * rendered text widths (prefix + " " + message) plus 5 px margin.
     * obf: {@code D}
     */
    private int panelWidth;

    /**
     * Font / text-size index passed to all SurfaceSprite drawing calls.
     * obf: {@code i}  (instance field — distinct from the static class {@code i} = ErrorHandler)
     */
    private int fontId;

    /**
     * Optional header/title string rendered above the entry list.
     * Null when this list has no header.
     * obf: {@code m}
     */
    private String headerText;

    /**
     * Backing array of entry records.  Capacity grows by 10 when needed.
     * obf: {@code b}  (field; type {@code t[]} = EntityDef[])
     */
    private EntityDef[] entries;       // obf: b, type t[] = EntityDef[]

    // -----------------------------------------------------------------------
    // Static fields
    // -----------------------------------------------------------------------

    /**
     * CRC-32 lookup table — 256 entries pre-computed in the static initialiser
     * using the standard IEEE polynomial {@code 0xEDB88320}.
     * This table is owned by wb in the obfuscated build but is logically a
     * utility — it is not used by any MessageList instance method.
     * obf: {@code q}
     */
    public static int[] crc32Table = new int[256];

    /**
     * Font character-width table shared across all MessageList instances.
     * {@code p} is an {@code aa} (BZip) instance whose byte array encodes
     * per-character pixel widths used by the legacy text engine.
     * obf: {@code p}  (type {@code aa} = BZip)
     */
    // static aa p = new aa(new byte[]{ ... });   // BZip font-width table
    // (value preserved verbatim below in the static initialiser)

    // -----------------------------------------------------------------------
    // Constructors
    // -----------------------------------------------------------------------

    /**
     * Constructs a MessageList with no header text.
     * Delegates to {@link #MessageList(SurfaceSprite, int, String)}.
     * obf: {@code wb(ba, int)}
     */
    public MessageList(SurfaceSprite surface, int fontId) {
        this(surface, fontId, null);
    }

    /**
     * Constructs a MessageList.
     *
     * @param surface    renderer used for drawing and text metrics  (obf: {@code ba} instance)
     * @param fontId     font index for all text rendering
     * @param headerText optional title string shown above the list; null = none
     *
     * obf: {@code wb(ba, int, String)}
     */
    public MessageList(SurfaceSprite surface, int fontId, String headerText) {
        // opaque predicate: boolean bl = client.vh; — always false, removed
        this.panelHeight = 0;
        this.count = 0;
        this.panelWidth = 0;

        this.fontId = fontId;
        this.entries = new EntityDef[10];
        this.surface = surface;
        this.headerText = headerText;

        // Pre-allocate entry records to avoid repeated allocation later.
        for (int i = 0; i < 10; i++) {
            this.entries[i] = new EntityDef();
        }

        recalcDimensions(true);
    }

    // -----------------------------------------------------------------------
    // Dimension helpers (called after any entry change)
    // -----------------------------------------------------------------------

    /**
     * Recalculates {@link #panelHeight} and {@link #panelWidth} by
     * measuring every entry's rendered text width/height.
     *
     * <p>If {@link #headerText} is non-null, the header contributes its own
     * height and width rows, then sets {@code panelWidth} to
     * {@code 5 + textWidth(headerText)} and clears the header if there are
     * no visible entries (so it collapses correctly on reset).
     *
     * <p>If the list has no header, {@code panelHeight} and
     * {@code panelWidth} are zeroed immediately.
     *
     * @param alwaysRecalc when {@code true} the width is re-walked for
     *                     every entry; the flag guards redundant work.
     *
     * obf: {@code a(boolean)}  — original name {@code wb.EA}
     */
    private void recalcDimensions(boolean alwaysRecalc) {
        // ++c;  // profiling counter removed

        // lineH = textHeight(fontId) + 1  (the +1 comes from the bytecode: result - -1)
        // Obf call: ba.a(508305352 /* anti-tamper sentinel */, fontId) → ua.a(II)I → textHeight
        // 508305352 is the magic constant accepted by ua.a(II)I's guard; real return is e.g. 12 px for font 0.
        int lineH = surface.textHeight(508305352, fontId) + 1;

        if (headerText == null) {
            // No header: collapse dimensions to zero.
            this.panelHeight = 0;
            this.panelWidth = 0;
            // fall through to entry-width loop (panelWidth will be updated there)
        } else {
            // Header present: height starts at one line, width = 5 + textWidth(headerText).
            // Obf call: ba.a(fontId, 76 /* sentinel */, headerText) → ua.a(IILjava/lang/String;)I → textWidth
            this.panelHeight = lineH;
            this.panelWidth = 5 + surface.textWidth(fontId, 76, headerText);
        }

        // if (!alwaysRecalc) this.getEntryLayer(true, 124);  // anti-tamper guard call, removed

        // Walk all live entries, accumulate height and track maximum width.
        for (int i = 0; i < count; i++) {
            this.panelHeight += lineH;
            // Obf call: ba.a(fontId, 105 /* sentinel */, prefix+" "+message) → textWidth
            int entryWidth = 5 + surface.textWidth(fontId, 105,
                    entries[i].name + " " + entries[i].extraText);
            if (entryWidth > this.panelWidth) {
                this.panelWidth = entryWidth;
            }
        }
    }

    // -----------------------------------------------------------------------
    // Entry management — internal workhorse
    // -----------------------------------------------------------------------

    /**
     * Core entry-append implementation.  Grows the backing array when needed,
     * then stores all supplied fields into {@code entries[count++]} and calls
     * {@link #recalcDimensions}.
     *
     * <p>The {@code threshold} parameter controls whether the entry is
     * actually stored: if {@code threshold <= 124} the method returns
     * immediately.  Public entry points pass 125 or 126 to activate storage;
     * passing ≤124 is an obfuscated guard path that the callers never hit.
     *
     * @param arg0       secondary int passed through to EntityDef.set()
     * @param prefix     first text field (sender name / label prefix)
     * @param arg2       secondary int field
     * @param layer      layer/z-order int field
     * @param message    main message text
     * @param color      text color (0xFFFFFF = white, 0xFFFF00 = highlighted)
     * @param arg6       unused String field (always null at call sites)
     * @param x          display X coordinate
     * @param extra      extra integer
     * @param threshold  activation guard: must be > 124 (125=normal, 126=rich)
     * @param arg10      secondary String field (always null at call sites)
     * @param arg11      secondary String field (always null at call sites)
     *
     * obf: {@code a(int,String,int,int,String,int,String,int,int,int,String,String)}
     * Original name: {@code wb.N}
     */
    private void addEntry(int arg0, String prefix, int arg2, int layer,
                          String message, int color, String arg6,
                          int x, int extra, int threshold,
                          String arg10, String arg11) {
        // Grow backing array by 10 if at capacity.
        if (count == entries.length) {
            EntityDef[] old = entries;
            entries = new EntityDef[count + 10];
            for (int i = 0; i < entries.length; i++) {
                if (i < count) {
                    entries[i] = old[i];
                } else {
                    entries[i] = new EntityDef();
                }
            }
        } else {
            // opaque variable assignment (n10=n8, n9=124) from obf — just
            // means we check threshold against 124:
        }

        // Guard: only store if threshold > 124.
        if (threshold <= 124) {
            return;
        }

        // ++L;  // profiling counter removed
        entries[count++].setFields(prefix, layer, color, arg0, extra, arg6,
                100 /* sortKey dummy */, x, message, arg10, arg2, arg11);
        recalcDimensions(true);
    }

    // -----------------------------------------------------------------------
    // Entry removal
    // -----------------------------------------------------------------------

    /**
     * Removes the entry at position {@code index} from the list by shifting
     * all subsequent entries left, then placing the removed entry at the new
     * tail position and calling {@link #recalcDimensions}.
     *
     * <p>Guards: returns immediately if {@code index} is out of range
     * ({@code index < 0} or {@code index >= count}).
     *
     * @param ignored1  obfuscated anti-tamper int (junk modulo expression, removed)
     * @param index     zero-based index of the entry to remove
     *
     * obf: {@code b(int, int)}  — original name {@code wb.G}
     */
    public final void removeEntry(int ignored1, int index) {
        // ++F;  // profiling counter removed
        // Guard: index must be in [0, count).
        if (index < 0 || index >= count) {
            return;
        }

        EntityDef removed = entries[index];

        // Shift entries left to fill the gap.
        for (int i = index; i < count - 1; i++) {
            entries[i] = entries[i + 1];
        }

        // Place the evicted record at the new tail and decrement count.
        entries[--count] = removed;

        recalcDimensions(true);
    }

    // -----------------------------------------------------------------------
    // Sort
    // -----------------------------------------------------------------------

    /**
     * Sorts {@link #entries} in ascending order of each entry's
     * {@link EntityDef#sortKey} using the {@link CacheUpdater#sortObjects}
     * helper (which delegates to {@link NameTable#sort}).
     *
     * <p>After sorting, copies the sorted object references back into
     * {@link #entries}.
     *
     * @param ignored  obfuscated dummy byte parameter (anti-tamper, removed)
     *
     * obf: {@code a(byte)}  — original name {@code wb.AA}
     */
    public final void sortEntries(byte ignored) {
        // ++o;  // profiling counter removed
        if (count == 0) {
            return;
        }

        int[] keys = new int[count];
        Object[] objs = new Object[count];

        for (int i = 0; i < count; i++) {
            EntityDef entry = entries[i];
            keys[i] = entry.fieldIntF;     // obf: t.d -> EntityDef.fieldIntF
            objs[i] = entry;
        }

        // Sort objs[] by keys[] ascending.
        CacheUpdater.sortNameTable((byte) -70, objs, keys);  // obf: cb.a((byte)-70, objs, keys)

        // Copy sorted refs back.
        for (int i = 0; i < count; i++) {
            entries[i] = (EntityDef) objs[i];
        }
    }

    // -----------------------------------------------------------------------
    // Rendering / hit-testing — private core
    // -----------------------------------------------------------------------

    /**
     * Draws all entries (and the optional header) onto the surface, or
     * performs a pure hit-test without drawing.
     *
     * <p>Returns the index of the entry whose rendered row contains
     * ({@code mouseX}, {@code mouseY}), {@code -2} if the header row is
     * hit, or {@code -1} if no hit.
     *
     * <p>When {@code doRender} is {@code true} each entry is rendered as:
     * <pre>
     *   surface.drawstring(prefix + " " + message, left+2, rowY, color, /*shadow=*\/false, fontId)
     * </pre>
     * Hovered entries are drawn in yellow ({@code 0xFFFF00}); others in white ({@code 0xFFFFFF}).
     * The header (if present) is drawn first in 16-bit white ({@code 0xFFFF}).
     *
     * <p><b>Note on parameter naming</b>: the obfuscator shuffles argument
     * order between the two callers ({@link #hitTest} and {@link #hitTestNoRender}).
     * The "canonical" interpretation below matches the {@code hitTest} caller
     * which passes {@code (topY, highlightSortKey, mouseX, mouseY, -66, true)}.
     *
     * @param topY             screen Y of the top of this panel (var1)
     * @param highlightSortKey second hit-test coordinate; used in both the header
     *                         and per-row bounding-box comparisons (var2)
     * @param mouseX           contributes to the row baseline (var3 → rowY)
     * @param mouseY           text/background left coordinate (var4)
     * @param unused           sentinel (-66 from hitTest, -3 from hitTestNoRender);
     *                         the {@code unused >= -1} guard never fires, so it has
     *                         no observable effect in normal use (var5)
     * @param doRender         {@code true} = draw pixels + hit-test; {@code false} = hit-test only (var6)
     *
     * @return index of hit entry, {@code -2} for header hit, or {@code -1} for none
     *
     * obf: {@code a(int,int,int,int,int,boolean)}  — original name {@code wb.R}
     */
    private int draw(int topY, int highlightSortKey, int mouseX, int mouseY,
                     int unused, boolean doRender) {
        // ++a;  // profiling counter removed

        // Early-out: panel has no dimensions yet (list is empty and no header).
        if (panelWidth == 0 || panelHeight == 0) {
            return -1;
        }

        // Draw grey background rectangle behind the whole panel.
        // Obf: ba.c(160, mouseY, panelHeight, 0, mouseX, panelWidth, 0xD0D0D0)
        // Note: argument order to ba.c (ua.c(IIIIIII)V = setBounds+fill) is
        //   (type=160, left, height, 0, top, width, color).
        // The confusing n4/n5 assignments in the bytecode correspond to the
        // shuffled call from hitTest: left=mouseY_arg, top=mouseX_arg.
        if (doRender) {
            surface.drawBoxAlpha(160, mouseY, panelHeight, 0,
                    mouseX, panelWidth, 0xD0D0D0 /* light grey */);
        }

        // lineH: line height in pixels.
        // Obf: 1 + ba.a(508305352, fontId) → textHeight
        int lineH = 1 + surface.textHeight(508305352, fontId);

        // rowY: current row baseline Y, starting 3 pixels below panel top.
        int rowY = -3 + (lineH + mouseX);

        int hitResult = -1;

        // --- Header row (optional) ---
        if (headerText != null) {
            // Hit-test header row bounds.
            // obf line: var4 < var2 && var1 > var8+3-var7 && ~var1 > ~(var8+3) && var2 < var4+D
            // (var2 = highlightSortKey, var4 = mouseY, var1 = topY, var8 = rowY, var7 = lineH)
            if (mouseY < highlightSortKey                // var4 < var2
                    && topY > rowY + (3 - lineH)         // var1 > var8 + 3 - var7
                    && ~topY > ~(rowY + 3)               // ~var1 > ~(var8 + 3)
                    && highlightSortKey < mouseY + panelWidth) {  // var2 < var4 + D
                if (!doRender) {
                    return -2;
                }
                hitResult = -2;
            }
            if (doRender) {
                // Draw header text in 16-bit white, no shadow.
                // Obf: ba.a(headerText, 2+mouseY, rowY, 65535, false, fontId)
                surface.drawstring(headerText, mouseY + 2, rowY,
                        0xFFFF /* 65535, white */, false, fontId);
            }
            rowY += lineH;
        }

        // obf: if (var5 >= -1) this.m = null;  — var5 is the 5th param (`unused`
        // sentinel = -66 from hitTest, -3 from hitTestNoRender).  Both are < -1,
        // so this branch never fires in normal use and headerText is NOT cleared.
        // (The earlier reconstruction wrongly tested `highlightSortKey` here, which
        // would clear the header for ordinary sort-key values.)
        if (unused >= -1) {
            this.headerText = null;
        }

        // --- Entry rows ---
        for (int i = 0; i < count; rowY += lineH, i++) {
            int color = 0xFFFFFF;  // default: white

            // Hit-test: is the mouse inside this row's bounding box?
            // Obf guard expression (negated comparisons from bytecode):
            //   ~var2 < ~var4  &&  ~(-var7+3+var8) > ~var1  &&  ~(3+var8) < ~var1  &&  ~var2 > ~(var4+D)
            // (var2 = highlightSortKey, var4 = mouseY, var1 = topY, var7 = lineH, var8 = rowY)
            if (~highlightSortKey < ~mouseY
                    && ~(-lineH + (3 + rowY)) > ~topY
                    && ~(3 + rowY) < ~topY
                    && ~highlightSortKey > ~(mouseY + panelWidth)) {
                color = 0xFFFF00;  // yellow: highlight hovered entry
                if (!doRender) {
                    return i;
                }
                hitResult = i;
            }

            if (!doRender) {
                continue;
            }

            // Draw "prefix message" text.
            // Obf: ba.a(p+" "+o, mouseY+2, rowY, color, false, fontId)
            surface.drawstring(entries[i].name + " " + entries[i].extraText,
                    mouseY + 2, rowY, color, false, fontId);
        }

        return hitResult;
    }

    // -----------------------------------------------------------------------
    // Public hit-test variants
    // -----------------------------------------------------------------------

    /**
     * Hit-tests the panel without rendering.  Returns the index of the entry
     * under ({@code mouseX}, {@code mouseY}), or {@code -1} / {@code -2}.
     *
     * @param mouseX X mouse position
     * @param mouseY Y mouse position
     * @param topY   panel top Y on screen
     * @param left   panel left X on screen
     * @param dummy  obfuscated dummy byte parameter (anti-tamper, removed)
     * @param highlightSortKey  passed through to {@link #draw}
     *
     * @return hit entry index, {@code -2} for header, {@code -1} for none
     *
     * obf: {@code a(int,int,int,byte,int)}  — original name {@code wb.U}
     */
    public final int hitTest(int mouseX, int mouseY, int topY, byte dummy, int highlightSortKey) {
        // if (dummy != -12) this.fontId = -77;   // anti-tamper guard, removed
        // ++H;  // profiling counter removed
        // Obf: this.a(topY, highlightSortKey, mouseX, mouseY, -66, true)
        return draw(topY, highlightSortKey, mouseX, mouseY, -66, true);
    }

    /**
     * Hit-tests without rendering (no-draw variant).
     *
     * @param mouseX          X mouse position
     * @param mouseY          Y mouse position
     * @param topY            panel top Y
     * @param left            panel left X
     * @param dummy           obfuscated dummy byte parameter (anti-tamper, removed)
     * @param highlightSortKey passed through to {@link #draw}
     *
     * @return hit entry index, {@code -2} for header, {@code -1} for none
     *
     * obf: {@code b(int,int,int,byte,int)}  — original name {@code wb.D}
     */
    public final int hitTestNoRender(int mouseX, int mouseY, int topY, byte dummy, int highlightSortKey) {
        // if (dummy != -40) this.a((byte)-62);   // anti-tamper guard, removed
        // ++M;  // profiling counter removed
        // Obf: this.a(highlightSortKey, mouseX, topY, mouseY, -3, false)
        // (argument order shuffled relative to hitTest — same private method)
        return draw(highlightSortKey, mouseX, topY, mouseY, -3, false);
    }

    // -----------------------------------------------------------------------
    // Dimension accessors
    // -----------------------------------------------------------------------

    /**
     * Returns the cached rendered height of this panel in pixels.
     *
     * @param dummy  obfuscated anti-tamper int (magic constant check, removed)
     * @return {@link #panelHeight}
     *
     * obf: {@code a(int)}  — original name {@code wb.T}
     */
    public final int getPanelHeight(int dummy) {
        // if (dummy != -21224) this.b(false, 0);  // anti-tamper guard, removed
        // ++j;  // profiling counter removed
        return panelHeight;
    }

    /**
     * Returns the cached rendered width of this panel in pixels.
     *
     * @param dummy  obfuscated anti-tamper int (magic constant check, removed)
     * @return {@link #panelWidth}
     *
     * obf: {@code b(int)}  — original name {@code wb.BA}
     */
    public final int getPanelWidth(int dummy) {
        // if (dummy != 16256) this.a((byte)-39);  // anti-tamper guard, removed
        // ++s;  // profiling counter removed
        return panelWidth;
    }

    /**
     * Returns the number of live entries in the list.
     *
     * @param dummy  obfuscated anti-tamper int (magic constant check, removed)
     * @return {@link #count}
     *
     * obf: {@code c(int)}  — original name {@code wb.F}
     */
    public final int getCount(int dummy) {
        // if (dummy != -27153) this.a(false);   // anti-tamper guard, removed
        // ++f;  // profiling counter removed
        return count;
    }

    /**
     * Sets the live entry count and recalculates display dimensions.
     * Used to truncate the list.
     *
     * @param newCount new entry count (must be ≤ entries.length)
     *
     * obf: {@code d(int)}  — original name {@code wb.P}
     */
    public final void setCount(int newCount) {
        // ++K;  // profiling counter removed
        count = newCount;
        recalcDimensions(true);
    }

    // -----------------------------------------------------------------------
    // Per-entry field accessors
    // -----------------------------------------------------------------------

    /**
     * Returns the {@code color} integer field of entry {@code index}.
     * (This is {@code t.e} in the obf build, which stores the color value
     * as passed to {@link #addEntry}.)
     *
     * @param guard  anti-tamper boolean (must be {@code true}; if false, calls guard)
     * @param index  entry index
     * @return {@link EntityDef#color}  (obf: {@code t.e})
     *
     * obf: {@code a(boolean, int)}  — original name {@code wb.C}
     */
    public final int getEntryColorE(boolean guard, int index) {
        // if (!guard) this.b((byte)30, 75);  // anti-tamper guard, removed
        // ++h;  // profiling counter removed
        return entries[index].fieldIntA;      // obf: t.e — stores the color field
    }

    /**
     * Returns the {@code message} string field of entry {@code index},
     * or {@code null} if the guard byte {@code guardByte ≤ 13}.
     *
     * @param guardByte  anti-tamper byte; if ≤ 13 returns null (dead path)
     * @param index      entry index
     * @return {@link EntityDef#message}  (obf: {@code t.o})
     *
     * obf: {@code b(byte, int)}  — original name {@code wb.O}
     */
    public final String getEntryMessage(byte guardByte, int index) {
        // ++E;  // profiling counter removed
        return guardByte <= 13 ? null : entries[index].extraText;  // obf: t.o
    }

    /**
     * Returns the {@code layer} integer field of entry {@code index}.
     *
     * @param guard  anti-tamper boolean (must be true)
     * @param index  entry index
     * @return {@link EntityDef#layer}  (obf: {@code t.l})
     *
     * obf: {@code b(boolean, int)}  — original name {@code wb.Q}
     */
    public final int getEntryLayer(boolean guard, int index) {
        // if (!guard) this.b(-33, (byte)91);  // anti-tamper guard, removed
        // ++l;  // profiling counter removed
        return entries[index].fieldIntD;      // obf: t.l
    }

    /**
     * Returns the {@code name} string field of entry {@code index},
     * or {@code null} if {@code guard != -4126}.
     *
     * @param index  entry index
     * @param guard  anti-tamper int; must equal -4126 (always true at call sites)
     * @return {@link EntityDef#name}  (obf: {@code t.b})
     *
     * obf: {@code c(int, int)}  — original name {@code wb.CA}
     */
    public final String getEntryName(int index, int guard) {
        // ++k;  // profiling counter removed
        return guard != -4126 ? null : entries[index].examineText;   // obf: t.b -> EntityDef.examineText
    }

    /**
     * Returns the {@code x} position / sort-key integer field of entry
     * {@code index} (stored in {@code t.d}), or {@code -114} when the guard
     * integer {@code guard ≥ -14} (dead path at all real call sites).
     *
     * <p>Despite the accessor name, {@code t.d} stores the X display position
     * passed to {@link #addEntry} — it is also used as the sort key in
     * {@link #sortEntries}.
     *
     * @param guard  anti-tamper int; real callers pass values < -14
     * @param index  entry index
     * @return {@link EntityDef#sortKey}  (obf: {@code t.d})
     *
     * obf: {@code a(int, int)}  — original name {@code wb.M}
     */
    public final int getEntryXPos(int guard, int index) {
        // ++C;  // profiling counter removed
        return guard >= -14 ? -114 : entries[index].fieldIntF;   // obf: t.d
    }

    /**
     * Returns the {@code sprite} integer field of entry {@code index}.
     *
     * @param index  entry index
     * @param dummy  obfuscated dummy byte (anti-tamper, removed)
     * @return {@link EntityDef#sprite}  (obf: {@code t.j})
     *
     * obf: {@code a(int, byte)}  — original name {@code wb.H}
     */
    public final int getEntrySprite(int index, byte dummy) {
        // if (dummy != 22) this.D = 4;  // anti-tamper guard, removed
        // ++d;  // profiling counter removed
        return entries[index].fieldIntC;     // obf: t.j
    }

    /**
     * Returns the {@code colorCode} integer field of entry {@code index},
     * or {@code 2} when {@code dummy ≠ 97}.
     *
     * @param dummy  anti-tamper byte; must equal 97
     * @param index  entry index
     * @return {@link EntityDef#colorCode}  (obf: {@code t.m})
     *
     * obf: {@code a(byte, int)}  — original name {@code wb.K}
     */
    public final int getEntryColorCode(byte dummy, int index) {
        // ++y;  // profiling counter removed
        return dummy != 97 ? 2 : entries[index].fieldIntE;   // obf: t.m
    }

    /**
     * Returns the {@code messageColor} integer field of entry {@code index}.
     *
     * @param index  entry index
     * @param guard  anti-tamper boolean; if true sets {@code fontId=119} (removed)
     * @return {@link EntityDef#messageColor}  (obf: {@code t.i})
     *
     * obf: {@code a(int, boolean)}  — original name {@code wb.L}
     */
    public final int getEntryMessageColor(int index, boolean guard) {
        // if (guard) this.fontId = 119;  // anti-tamper guard, removed
        // ++x;  // profiling counter removed
        return entries[index].fieldIntB;  // obf: t.i
    }

    /**
     * Returns the {@code prefix} string field of entry {@code index},
     * or {@code null} if {@code dummy ≠ 53}.
     *
     * @param index  entry index
     * @param dummy  anti-tamper byte; must equal 53
     * @return {@link EntityDef#prefix}  (obf: {@code t.p})
     *
     * obf: {@code b(int, byte)}  — original name {@code wb.A}
     */
    public final String getEntryPrefix(int index, byte dummy) {
        if (dummy != 53) {
            return null;
        }
        // ++J;  // profiling counter removed
        return entries[index].name;     // obf: t.p
    }

    // -----------------------------------------------------------------------
    // Public add-entry variants
    // All delegate to addEntry() with appropriate sentinel thresholds.
    // -----------------------------------------------------------------------

    /**
     * Adds a full entry with prefix, message, two string fields, a numeric
     * x-position, and a second string field.
     *
     * <p>Threshold = 125 (standard add, always stored).
     *
     * @param prefix   sender/label string
     * @param message  body message string
     * @param arg2     secondary string (examine text or similar)
     * @param x        X display position
     * @param arg4     additional string
     * @param dummy    obfuscated dummy byte (junk modulo expression, removed)
     *
     * obf: {@code a(String,String,String,int,String,byte)}  — original name {@code wb.E}
     */
    public final void addEntryFull(String prefix, String message, String arg2,
                            int x, String arg4, byte dummy) {
        // int junk = -26 % ((dummy - 15) / 33);  // junk modulo, removed
        addEntry(0, prefix, 0, 0, message, 0, null, x, 0, 125, arg2, arg4);
        // ++B;  // profiling counter removed
    }

    /**
     * Adds a simple colored entry with a color value and optional scroll.
     *
     * <p>If {@code scroll} is {@code true}, calls {@link #sortEntries} after
     * appending. Threshold = 125.
     *
     * @param color    text color integer
     * @param x        X position
     * @param scroll   if true, sort/scroll the list after adding
     * @param prefix   sender/label string
     * @param message  body text string
     *
     * obf: {@code a(int,int,boolean,String,String)}  — original name {@code wb.DA}
     */
    public final void addEntryScrolled(int color, int x, boolean scroll,
                                String prefix, String message) {
        addEntry(0, prefix, 0, 0, message, color, null, x, 0, 125, null, null);
        // ++u;  // profiling counter removed
        if (scroll) {
            sortEntries((byte) 61);
        }
    }

    /**
     * Adds a simple entry (no extra fields, no color override, threshold 125).
     * The {@code guard} parameter encodes the threshold via XOR:
     * {@code threshold = guard ^ 0x758D} which equals 125 when {@code guard == 30192}.
     *
     * @param x        X position
     * @param prefix   sender/label string
     * @param message  body text string
     * @param guard    must be 30192 (30192 ^ 0x758D = 125)
     *
     * obf: {@code a(int,String,String,int)}  — original name {@code wb.V}
     */
    public final void addEntrySimple(int x, String prefix, String message, int guard) {
        // if (guard != 30192) this.a(true, 125);  // anti-tamper guard, removed
        // ++e;  // profiling counter removed
        // threshold = guard ^ 30093 = 30192 ^ 30093 = 125
        addEntry(0, message, 0, 0, prefix, 0, null, x, 0, guard ^ 0x758D, null, null);
    }

    /**
     * Adds a rich entry with explicit per-field parameters.
     *
     * <p>Threshold = 126 (rich-mode, slightly higher than standard 125).
     * The junk division expression {@code -66/((arg2 - -42)/41)} is removed.
     *
     * @param x        X position
     * @param prefix   sender/label string
     * @param arg2     secondary int (junk divisor guard, removed)
     * @param layer    layer/z-order
     * @param extra    extra int
     * @param color    text color
     * @param sortKey  sort key / secondary field
     * @param message  body text string
     * @param fontId   per-entry font override
     *
     * obf: {@code a(int,String,int,int,int,int,int,String,int)}  — original name {@code wb.I}
     */
    public final void addEntryRich(int x, String prefix, int arg2, int layer, int extra,
                            int color, int sortKey, String message, int fontId) {
        // ++r;  // profiling counter removed
        // int junk = -66 / ((arg2 - -42) / 41);  // junk modulo, removed
        addEntry(x, prefix, sortKey, layer, message, fontId, null,
                color, extra, 126, null, null);
    }

    /**
     * Adds an entry with an explicit color and guard offset.
     * Used for friend/social list updates.  {@code threshold = guard + -3170}
     * which is 126 when {@code guard == 3296}.
     *
     * @param color    text color
     * @param prefix   sender/label string
     * @param x        X position
     * @param message  body text string
     * @param layer    layer int
     * @param guard    must be 3296 (3296 - 3170 = 126)
     *
     * obf: {@code a(int,String,int,String,int,int)}  — original name {@code wb.J}
     */
    public final void addEntryWithColor(int color, String prefix, int x,
                                 String message, int layer, int guard) {
        // if (guard != 3296) w = -93L;  // anti-tamper guard, removed
        // ++v;  // profiling counter removed
        addEntry(layer, message, 0, 0, prefix, color, null,
                x, 0, guard + -3170, null, null);
    }

    /**
     * Adds a guarded entry; only stores if {@code threshold > 44}.
     *
     * @param x         X position
     * @param color     text color
     * @param sortKey   sort key
     * @param layer     layer int
     * @param threshold activation threshold; must be > 44 to store
     * @param extra     extra int
     * @param prefix    sender/label string
     * @param message   body text string
     *
     * obf: {@code a(int,int,int,int,int,int,String,String)}  — original name {@code wb.B}
     */
    public final void addEntryGuarded(int x, int color, int sortKey, int layer,
                               int threshold, int extra,
                               String prefix, String message) {
        if (threshold <= 44) {
            return;
        }
        // ++G;  // profiling counter removed
        addEntry(color, message, sortKey, 0, prefix, layer, null,
                x, extra, 125, null, null);
    }

    /**
     * Adds an entry with a font override and secondary sort key.
     * Threshold = 127.
     *
     * @param color    text color
     * @param dummy    obfuscated dummy byte (anti-tamper, removed)
     * @param x        X position
     * @param prefix   sender/label string
     * @param message  body text string
     * @param sortKey  sort key
     * @param fontOverride  per-entry font index
     *
     * obf: {@code a(int,byte,int,String,String,int,int)}  — original name {@code wb.W}
     */
    public final void addEntryWithFont(int color, byte dummy, int x,
                                String prefix, String message,
                                int sortKey, int fontOverride) {
        // if (dummy != 22) this.a(33, false);  // anti-tamper guard, removed
        // ++z;  // profiling counter removed
        addEntry(fontOverride, prefix, sortKey, 0, message, color, null,
                x, 0, 127, null, null);
    }

    // -----------------------------------------------------------------------
    // Static texture-span renderer
    // -----------------------------------------------------------------------

    /**
     * Fills a perspective-correct affine texture span (a single horizontal
     * scanline segment) into the pixel output buffer {@code destPixels}.
     *
     * <p>This is a standalone renderer helper that happens to live in this
     * class in the obfuscated build.  It has no dependency on MessageList
     * instance state and is called directly from the scene renderer.
     *
     * <h3>Algorithm</h3>
     * The span is subdivided into 16-pixel blocks for performance.  Within
     * each 16-pixel block the texture U and V coordinates are advanced
     * linearly (affine approximation), reducing the number of divides.
     * At the start of each block, true perspective U/V are computed:
     * <pre>
     *   u = (uNumer / depth) &lt;&lt; 7   (fixed-point 7 bits)
     *   v = (vNumer / depth) &lt;&lt; 7
     * </pre>
     * After 16 pixels the next block's end U/V values are computed, and the
     * per-pixel deltas ({@code uPixStep}, {@code vPixStep}) are derived.
     *
     * <p>Pixel 0 is a safety guard that clears the destination if the texture
     * lookup returns 0 (transparent).
     *
     * <p><b>Parameter slots (verified against clean base / CFR n2..n15):</b> the
     * destination write index is the <i>10th</i> argument ({@code destX}=var9,
     * incremented as pixels are emitted to {@code destPixels[destX]}), NOT the
     * 3rd. Slots var2/var3 (here {@code scratchU}/{@code scratchV}) and var12
     * ({@code scratchTexel}) are dead inputs — they are overwritten immediately
     * inside the loop (reused as the running U/V accumulators and the texel temp)
     * so their incoming values are never observed.
     *
     * @param texOffsetX   per-row V offset added to {@code vNumer} each scanline row (var0)
     * @param blockSize    scanline length in pixels (guard: ≠ 10 triggers recursive call) (var1)
     * @param scratchU     dead input (var2) — reused internally as the U accumulator
     * @param scratchV     dead input (var3) — reused internally as the V accumulator
     * @param destPixels   output pixel array (var4)
     * @param depth        current perspective depth value (reciprocal Z × 128) (var5)
     * @param texSrcU      texture U numerator × depth (var6)
     * @param uNumer       perspective U numerator (world coord) (var7)
     * @param vNumer       perspective V numerator (world coord) (var8)
     * @param destX        starting X in the destination pixel buffer — the WRITE INDEX (var9)
     * @param depthStep    per-row depth delta ({@code depth += depthStep}) (var10)
     * @param uStep        U step accumulator (shifted left by 2 after entry) (var11)
     * @param scratchTexel dead input (var12) — reused internally as the texel temp
     * @param depthDelta   per-row {@code uNumer} delta ({@code uNumer += depthDelta}) (var13)
     * @param texPixels    source texture pixel data (128×128 tiles, packed) (var14)
     * @param spanCount    number of pixels remaining in this span (var15)
     *
     * obf: {@code a(int,int,int,int,int[],int,int,int,int,int,int,int,int,int,int[],int)} — original name {@code wb.S}
     */
    public static final void drawTextureSpan(
            int texOffsetX, int blockSize, int scratchU, int scratchV,
            int[] destPixels, int depth, int texSrcU,
            int uNumer, int vNumer, int destX, int depthStep,
            int uStep, int scratchTexel, int depthDelta,
            int[] texPixels, int spanCount) {

        // ++g;  // profiling counter removed

        if (spanCount <= 0) {
            return;
        }

        // Anti-tamper guard: if blockSize != 10, recurse with dummy args.
        // Real callers always pass blockSize=10; the recursion is dead.
        // Removed: if (blockSize != 10) { drawTextureSpan(-30,...,56); }

        // Perspective-correct affine texture: compute initial U/V per block.
        // uFixed and vFixed are 7-bit fixed-point accumulators (<<7).
        int uFixed = 0;  // current block-end U in fixed-point
        int vFixed = 0;  // current block-end V in fixed-point

        uStep <<= 2; // shift the step accumulator (matches bytecode: n12 <<= 2)

        if (depth != 0) {
            // obf: var16 = var7/var5 << 7  (var7 = uNumer);  var17 = var8/var5 << 7 (var8 = vNumer)
            uFixed = uNumer / depth << 7;  // uFixed (var16) = (uNumer/depth) << 7
            vFixed = vNumer / depth << 7;  // vFixed (var17) = (vNumer/depth) << 7
        }

        // Clamp uFixed to [0, 16256]  (128*128-1 = 16383, tile mask 0x3F80)
        if (uFixed < 0) uFixed = 0;
        else if (uFixed > 16256) uFixed = 16256;

        int remaining = spanCount;

        while (remaining > 0) {
            // Advance world coords by one row.
            uNumer += depthDelta;
            int prevVFixed = vFixed;
            depth    += depthStep;
            int prevUFixed = uFixed;
            int localSpan   = texOffsetX;

            // Inner pixel loop — process one 16-pixel block (or remainder).
            while (true) {
                // Advance the V numerator by one row.  NOTE: this mutates the
                // persistent `vNumer` (obf var8 = var10000 + var10001), so the
                // offset ACCUMULATES across outer rows — it is NOT reseeded each
                // iteration.  (The earlier reconstruction used a fresh per-row
                // local, which dropped the accumulation.)
                vNumer = localSpan + vNumer;
                // opaque predicate: if (bl) return; — removed

                if (depth != 0) {
                    uFixed = uNumer / depth << 7;
                    vFixed = vNumer / depth << 7;
                }

                // Clamp uFixed.
                if (uFixed < 0) uFixed = 0;
                else if (uFixed > 16256) uFixed = 16256;

                // Per-pixel affine step within this 16-pixel block.
                int uPixStep = (-prevUFixed + uFixed) >> 4;
                int vPixStep = (-prevVFixed + vFixed) >> 4;

                // Texture page selector (from texSrcU accumulator).
                // obf: var21 = var6>>23; var2 += 0x600000 & var6; var6 += var11
                // (var2 = prevUFixed — the U accumulator, NOT texOffsetX/var0)
                int texPage = texSrcU >> 23;
                prevUFixed += 0x600000 & texSrcU;  // sync sub-pixel page bits into U accumulator
                texSrcU    += uStep;

                if (remaining < 16) {
                    // Tail block: process remaining pixels individually.
                    // obf: `break label208` jumps PAST the unrolled block but still
                    // executes `var20 -= 16` (remaining -= 16) — so this branch must
                    // fall through to the `remaining -= 16; break;` below, NOT break
                    // out of the inner loop directly (which would never decrement
                    // `remaining` and loop forever).
                    for (int px = 0; remaining > px; px++) {
                        int texel = texPixels[(prevVFixed & 0x3F80) + (prevUFixed >> 7)] >>> texPage;
                        if (texel != 0) {
                            destPixels[destX] = texel;
                        }
                        destX++;
                        prevUFixed += uPixStep;
                        prevVFixed += vPixStep;

                        // Every 4th pixel: re-sync sub-pixel texture page.
                        if ((px & 3) == 3) {
                            prevUFixed = (texSrcU & 0x600000) + (0x3FFF & prevUFixed);
                            texPage  = texSrcU >> 23;
                            texSrcU  += uStep;
                        }
                    }
                } else {

                // Full 16-pixel block, manually unrolled for throughput.
                // Each group of 4 pixels shares a texture-page refresh.

                // Pixel 0
                int texel;
                texel = texPixels[(prevVFixed & 0x3F80) + (prevUFixed >> 7)] >>> texPage;
                if (texel != 0) destPixels[destX] = texel;
                prevUFixed += uPixStep; destX++;
                prevVFixed += vPixStep;

                // Pixel 1
                texel = texPixels[(prevUFixed >> 7) + (0x3F80 & prevVFixed)] >>> texPage;
                if (texel != 0) destPixels[destX] = texel;
                destX++; prevUFixed += uPixStep; prevVFixed += vPixStep;

                // Pixel 2
                texel = texPixels[(prevUFixed >> 7) + (0x3F80 & prevVFixed)] >>> texPage;
                if (texel != 0) destPixels[destX] = texel;
                destX++; prevUFixed += uPixStep; prevVFixed += vPixStep;

                // Pixel 3 — sync texture page after group of 4
                texel = texPixels[(prevUFixed >> 7) + (prevVFixed & 0x3F80)] >>> texPage;
                if (texel != 0) destPixels[destX] = texel;
                prevUFixed += uPixStep; prevVFixed += vPixStep; destX++;
                texPage   = texSrcU >> 23;
                prevUFixed = (texSrcU & 0x600000) + (0x3FFF & prevUFixed);
                texSrcU   += uStep;

                // Pixel 4
                texel = texPixels[(prevVFixed & 0x3F80) + (prevUFixed >> 7)] >>> texPage;
                if (texel != 0) destPixels[destX] = texel;
                destX++; prevUFixed += uPixStep; prevVFixed += vPixStep;

                // Pixel 5
                texel = texPixels[(0x3F80 & prevVFixed) - -(prevUFixed >> 7)] >>> texPage;
                if (texel != 0) destPixels[destX] = texel;
                destX++; prevUFixed += uPixStep; prevVFixed += vPixStep;

                // Pixel 6
                texel = texPixels[(prevUFixed >> 7) + (0x3F80 & prevVFixed)] >>> texPage;
                if (texel != 0) destPixels[destX] = texel;
                destX++; prevUFixed += uPixStep; prevVFixed += vPixStep;

                // Pixel 7 — sync
                // obf: index reads var3 (old prevVFixed); var3 += var19 happens AFTER the read
                texel = texPixels[(prevUFixed >> 7) + (prevVFixed & 0x3F80)] >>> texPage;
                if (texel != 0) destPixels[destX] = texel;
                prevVFixed += vPixStep; destX++; prevUFixed += uPixStep;
                texPage   = texSrcU >> 23;
                prevUFixed = (prevUFixed & 0x3FFF) - -(0x600000 & texSrcU);

                // Pixel 8
                texel = texPixels[(prevUFixed >> 7) + (prevVFixed & 0x3F80)] >>> texPage;
                if (texel != 0) destPixels[destX] = texel;
                texSrcU += uStep; destX++; prevUFixed += uPixStep; prevVFixed += vPixStep;

                // Pixel 9
                texel = texPixels[(prevUFixed >> 7) + (prevVFixed & 0x3F80)] >>> texPage;
                if (texel != 0) destPixels[destX] = texel;
                destX++; prevUFixed += uPixStep; prevVFixed += vPixStep;

                // Pixel 10
                texel = texPixels[(0x3F80 & prevVFixed) - -(prevUFixed >> 7)] >>> texPage;
                if (texel != 0) destPixels[destX] = texel;
                destX++; prevUFixed += uPixStep; prevVFixed += vPixStep;

                // Pixel 11
                // obf: var2 += var18; var3 += var19; var9++; then sub-page mask on var2
                texel = texPixels[(prevUFixed >> 7) + (prevVFixed & 0x3F80)] >>> texPage;
                if (texel != 0) destPixels[destX] = texel;
                prevUFixed += uPixStep; prevVFixed += vPixStep; destX++;
                prevUFixed = (prevUFixed & 0x3FFF) - -(texSrcU & 0x600000);
                texPage    = texSrcU >> 23;

                // Pixel 12
                texel = texPixels[(prevVFixed & 0x3F80) + (prevUFixed >> 7)] >>> texPage;
                if (texel != 0) destPixels[destX] = texel;
                texSrcU += uStep; prevVFixed += vPixStep; destX++; prevUFixed += uPixStep;

                // Pixel 13
                texel = texPixels[(prevUFixed >> 7) + (prevVFixed & 0x3F80)] >>> texPage;
                if (texel != 0) destPixels[destX] = texel;
                destX++; prevUFixed += uPixStep; prevVFixed += vPixStep;

                // Pixel 14
                texel = texPixels[(prevUFixed >> 7) + (prevVFixed & 0x3F80)] >>> texPage;
                if (texel != 0) destPixels[destX] = texel;
                destX++; prevUFixed += uPixStep; prevVFixed += vPixStep;

                // Pixel 15
                texel = texPixels[(0x3F80 & prevVFixed) - -(prevUFixed >> 7)] >>> texPage;
                if (texel != 0) destPixels[destX] = texel;
                destX++;

                } // end else (full 16-pixel block)

                // Both the tail branch and the full-block branch converge here
                // (obf: `var20 -= 16` runs for both paths), then the inner loop
                // breaks back to the outer `while (remaining > 0)`.
                remaining -= 16;
                // opaque predicate: if (bl) return; — removed
                break;
            }
        }
    }

    // -----------------------------------------------------------------------
    // Static initialiser — CRC-32 table
    // -----------------------------------------------------------------------

    static {
        // Compute the standard IEEE 802.3 CRC-32 lookup table.
        // Polynomial: 0xEDB88320 (reversed bit order).
        for (int i = 0; i < 256; i++) {
            int crc = i;
            for (int bit = 0; bit < 8; bit++) {
                if ((crc & 1) == 1) {
                    crc = (crc >>> 1) ^ 0xEDB88320;  // obf: ^ -306674912
                } else {
                    crc >>>= 1;
                }
            }
            crc32Table[i] = crc;
        }
    }

    // -----------------------------------------------------------------------
    // XOR string decoding helpers (obfuscation artifact — not called at runtime
    // after class loading; kept for completeness and cross-reference)
    // -----------------------------------------------------------------------

    /**
     * First-pass XOR: converts a String to a char[] applying a 1-char key
     * (only used when the array has fewer than 2 elements; otherwise the
     * mod-5 pass in {@link #decodeString} handles it).
     *
     * Key: {@code 'G'} (0x47).
     *
     * obf: {@code z(String)}
     */
    private static char[] decodeChars(String s) {
        char[] arr = s.toCharArray();
        if (arr.length < 2) {
            arr[0] = (char) (arr[0] ^ 'G');  // 0x47
        }
        return arr;
    }

    /**
     * Second-pass XOR: decodes a char[] using the 5-byte rotating key table
     * {@code [104, 66, 37, 56, 71]} (i.e. {@code h B % 8 G}).
     *
     * The decoded strings (used as error-message prefixes in A[]) are:
     * <pre>
     *   A[0]  = "{...}"          A[1]  = "null"
     *   A[2]  = "wb.E("          A[3]  = "wb.DA("
     *   A[4]  = "wb.I("          A[5]  = "wb.T("
     *   A[6]  = "wb.BA("         A[7]  = "wb.N("
     *   A[8]  = "wb.U("          A[9]  = "wb.K("
     *   A[10] = "wb.B("          A[11] = "wb.P("
     *   A[12] = "wb.S("          A[13] = "wb.A("
     *   A[14] = "wb.&lt;init&gt;(" A[15] = "wb.J("
     *   A[16] = "wb.H("          A[17] = "wb.G("
     *   A[18] = "wb.D("          A[19] = "wb.Q("
     *   A[20] = "wb.L("          A[21] = "wb.F("
     *   A[22] = "wb.AA("         A[23] = "wb.CA("
     *   A[24] = "wb.C("          A[25] = "wb.O("
     *   A[26] = "wb.M("          A[27] = "wb.EA("
     *   A[28] = "wb.V("          A[29] = "wb.R("
     *   A[30] = "wb.W("
     * </pre>
     *
     * obf: {@code z(char[])}
     */
    private static String decodeString(char[] arr) {
        // Key table: positions 0-4 cycle through h(104) B(66) %(37) 8(56) G(71)
        final int[] KEY = {104, 66, 37, 56, 71};
        for (int i = 0; i < arr.length; i++) {
            arr[i] = (char) (arr[i] ^ KEY[i % 5]);
        }
        return new String(arr).intern();
    }
}
