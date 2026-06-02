package client;

import client.data.NameTable;
import client.data.RecordLoader;
import client.ui.MessageList;
import client.util.Utility;

/**
 * WidgetRenderer — the leaf 2D UI widget primitives extracted from Mudclient.
 *
 * <p>Contains the pure {@link client.scene.SurfaceSprite} (li) blitting methods plus
 * walkTo/walkToAction hit-test callbacks for the scrollbar widgets. These methods have
 * no this-identity, no synchronized, and no AWT Component identity.
 *
 * <p>All Mudclient fields and shared-routine callbacks are reached through the {@code m}
 * back-reference.
 *
 * <p>Methods extracted from Mudclient:
 * <ul>
 *   <li>{@link #drawBox} — sprite-slot border + fill via walkToAction</li>
 *   <li>{@link #clearScreen} — walkable-direction resolver for movement</li>
 *   <li>{@link #drawSprite} — UI-sprite blit via walkToAction</li>
 *   <li>{@link #drawIcon} — minimap/compass icon blit via li.setPixel</li>
 *   <li>{@link #drawScrollList} — configure generic scrollable menu/list</li>
 *   <li>{@link #drawMenuOptions} — right-click option list setup</li>
 *   <li>{@link #drawScrollbar} — primary scrollbar hit-test widget</li>
 *   <li>{@link #drawScrollbar2} — secondary scrollbar/slider widget</li>
 * </ul>
 */
class WidgetRenderer {
    final Mudclient m;

    WidgetRenderer(Mudclient m) {
        this.m = m;
    }

    // -------------------------------------------------------------------------
    // drawBox  (obf: final void b(int,int,int,int,int)  @clean L12511)
    // -------------------------------------------------------------------------

    /**
     * Draw a sprite-slot border + fill via two walkToAction passes.
     * Opaque-predicate guard: if magicKey != 5126, calls drawUiTabMagic (dead path).
     * obf: final void b(int magicKey, int styleIndex, int x, int y, int style)
     */
    final void drawBox(int magicKey, int styleIndex, int x, int y, int style) {
        // Opaque-predicate guard: if (magicKey != 5126) call a setup helper (dead path)
        if (magicKey != 5126) {
            m.drawUiTabMagic(true, (byte)-25);   // obf: this.b(boolean,byte)
        }

        int dimX, dimY;  // var6 = X-extent, var7 = Y-extent
        if (~style == -1 || ~style == -5) {   // style == 0 or style == 4
            dimY = NameTable.sortKeys[styleIndex];   // obf: var7 = ub.g[var2]
            dimX = RecordLoader.intArray[styleIndex];// obf: var6 = f.f[var2]
        } else {
            dimX = NameTable.sortKeys[styleIndex];   // obf: var6 = ub.g[var2]
            dimY = RecordLoader.intArray[styleIndex];// obf: var7 = f.f[var2]
        }

        // Skip outer border when the slot render-state is 2 or 3 (Utility.sizedPoolCounts[] = mb.a[])
        if (Utility.sizedPoolCounts[styleIndex] != 2 && Utility.sizedPoolCounts[styleIndex] != 3) {
            // Outer/border pass (flags: walk=true, mode -59)
            m.walkToAction(x, true, m.Lf, y, m.sh,
                           -1 + dimX + x, true, -1 + (y + dimY), -59);
        }

        // Style-direction adjustments
        if (style == 0) { ++dimX; --x; }
        if (style == 2) { ++dimY; }
        if (style == 6) { --y; ++dimY; }
        if (style == 4) { ++dimX; }

        // Inner/fill pass (flags: walk=false, mode -14)
        m.walkToAction(x, true, m.Lf, y, m.sh,
                       dimX + x - 1, false, dimY + y - 1, -14);
    }

    // -------------------------------------------------------------------------
    // clearScreen  (obf: private final void q(byte)  @clean L12555)
    // -------------------------------------------------------------------------

    /**
     * Resolve a clear walkable direction by cycling {@code m.si} through the available
     * terrain walkability slots (World collision via {@code m.isDirectionWalkable} probes).
     * {@code numExtraDirections > 7} enables the extended {±1,±2,±3,+4} search.
     *
     * CLEAN-BASE CORRECTION: the extended-search loop must NOT return when it finds a
     * clear direction (the defective base added an early {@code return}); after the loop the
     * code re-tests (si&1)==0 && b(91,si) and applies a secondary ±1 nudge, and the whole
     * extended block is nested inside {@code if (numExtraDirections > 7)}.
     * obf: void q(byte)  no label  (skeleton proposed name: clearScreen) (il[389])
     */
    final void clearScreen(byte numExtraDirections) {
        // Primary fast path: if (si&1)==1 and direction 90 is clear, done.
        if ((m.si & 1) == 1 && m.isDirectionWalkable((byte)90, m.si)) {
            return;
        }

        // Secondary: (si&1)==0 and direction 113 is clear → nudge si by ±1 within octet.
        if ((m.si & 1) == 0 && m.isDirectionWalkable((byte)113, m.si)) {
            if (!m.isDirectionWalkable((byte)-127, (1 + m.si) & 7)) {
                if (!m.isDirectionWalkable((byte)22, (7 + m.si) & 7)) {
                    return;
                }
                m.si = (7 + m.si) & 7;
                return;
            }
            m.si = (1 + m.si) & 7;
            return;
        }

        // Extended direction search (only when numExtraDirections > 7)
        int[] dirOffsets = new int[]{1, -1, 2, -2, 3, -3, 4};
        if (numExtraDirections <= 7) {
            return;
        }

        // Probe each offset; on a hit, set si toward it AND fall through (no early return).
        for (int d = 0; d < 7; ++d) {  // obf: -8 < ~var3 ⇔ var3 < 7
            if (m.isDirectionWalkable((byte)51, (8 + m.si + dirOffsets[d]) & 7)) {
                m.si = (m.si + dirOffsets[d] + 8) & 7;
                break;
            }
        }

        // Secondary nudge after the search: requires (si&1)==0 and direction 91 clear.
        if ((m.si & 1) == 0 && m.isDirectionWalkable((byte)91, m.si)) {
            if (m.isDirectionWalkable((byte)29, (1 + m.si) & 7)) {
                m.si = (1 + m.si) & 7;
                return;
            }
            if (m.isDirectionWalkable((byte)-125, (7 + m.si) & 7)) {
                m.si = (7 + m.si) & 7;
            }
        }
    }

    // -------------------------------------------------------------------------
    // drawSprite  (obf: private final void a(boolean,int,int,int)  @clean L12718)
    // -------------------------------------------------------------------------

    /**
     * Blit a UI sprite (by draw-list index) to the li at (x, y) via walkToAction,
     * with an optional screen-mode flag that sets {@code m.cl = 61}.
     *
     * CLEAN-BASE CORRECTION: drawMode 1 and 2 were SWAPPED in the defective base.
     *   drawMode==0 → mode -8  at (x, y-1)..(x, y)
     *   drawMode==1 → mode 126 at (x-1, y)..(x, y)   [fall-through case]
     *   drawMode==2 → mode 118 at (x, y)..(x, y)
     * obf: void a(boolean,int,int,int)  no label  (il[388])
     */
    final void drawSprite(boolean setScreenMode, int x, int y, int drawMode) {
        if (~drawMode == -1) {                 // drawMode == 0
            m.walkToAction(x, true, m.Lf, y - 1, m.sh, x, false, y, -8);
        } else if (~drawMode != -2) {          // drawMode != 1  → handles drawMode == 2
            m.walkToAction(x, true, m.Lf, y, m.sh, x, true, y, 118);
        } else {                               // drawMode == 1  (fall-through)
            m.walkToAction(x - 1, true, m.Lf, y, m.sh, x, false, y, 126);
        }

        if (setScreenMode) {
            m.cl = 61;
        }
    }

    // -------------------------------------------------------------------------
    // drawIcon  (obf: private final void a(int,int,byte,int)  @clean L7950)
    // -------------------------------------------------------------------------

    /**
     * Blit a compass/minimap icon: base sprite (76) + arrow overlay (111), and for
     * the large variant (size <= -32) three extra border sprites.
     * Skeleton labels this "setCamera" but the body is pure 2D sprite blitting; the
     * real Hh-camera positioning is inlined elsewhere.
     * obf: void a(int x, int y, byte size, int spriteBase)
     */
    final void drawIcon(int x, int y, byte size, int spriteBase) {
        m.li.setPixel(spriteBase, y, 76, x);
        m.li.setPixel(spriteBase, y - 1, 111, x);
        if (size <= -32) {
            m.li.setPixel(spriteBase, y + 1, 111, x);
            m.li.setPixel(spriteBase - 1, y, 60, x);
            m.li.setPixel(spriteBase + 1, y, 112, x);
        }
    }

    // -------------------------------------------------------------------------
    // drawScrollList  (obf: private void a(int,int,String[],boolean,String)  @clean L11202)
    // -------------------------------------------------------------------------

    /**
     * Configure the generic scrollable list/menu: stash the option strings, compute the
     * widest row (min 400) and the total height from the font metrics, and reset the
     * list state. Clears the bound message list unless x==3.
     * obf: void a(int,int,String[],boolean,String)
     */
    void drawScrollList(int x, int y, String[] options, boolean showBorder, String title) {
        m.menuOptionList = options;       // obf: od
        m.menuWidth = 400;                // obf: zi
        if (x != 3) {
            m.scrollMessageList = null;   // obf: Wf
        }
        for (int i = 0; i < options.length; i++) {
            int w = m.li.textWidth(1, x + 113, options[i]) + 10;
            if (m.menuWidth < w) {
                m.menuWidth = w;
            }
        }
        m.menuHeight = 15 + (m.li.textHeight(508305352, 1) + 2) * (1 + options.length) + m.li.textHeight(508305352, 4);   // obf: gl
        m.menuX = y;            // obf: gc
        m.menuTitle = title;    // obf: e
        m.menuOpenFlag = false; // obf: vk
        m.inputLine = "";       // obf: inputTextFinal
        m.showMenuBorder = showBorder;   // obf: Bd
    }

    // -------------------------------------------------------------------------
    // drawMenuOptions  (obf: private void a(String[],int,int,boolean)  @clean L11185)
    // -------------------------------------------------------------------------

    /**
     * Render the right-click "Choose option" list. Clears the screen first unless
     * called with sentinel 12, then delegates to drawScrollList with a 9px left margin.
     * obf: void a(String[],int,int,boolean)
     */
    void drawMenuOptions(String[] options, int x, int y, boolean rightClick) {
        if (x != 12) {
            // obf: this.e((byte)31) — NOT clearScreen (q(byte)@17480); e(byte)@12828 resets
            // entity counts + panel id + username/password buffers. Named per behaviour.
            m.resetMenuState((byte)31);
        }
        drawScrollList(x - 9, y, options, rightClick, "");
    }

    // -------------------------------------------------------------------------
    // drawScrollbar  (obf: private void a(byte,int,int,int,boolean,int)  @clean L11230)
    // -------------------------------------------------------------------------

    /**
     * Scrollbar widget (primary variant). Hit-tests via the walkTo dispatch helper; if
     * the first probe misses, runs the second probe and (unless sentinel==10) draws the
     * secondary slider.
     * obf: void a(byte,int,int,int,boolean,int)
     */
    void drawScrollbar(byte sentinel, int x, int y, int scrollPos, boolean animate, int trackLen) {
        if (!m.walkTo(x, trackLen, (byte)14, false, scrollPos, scrollPos, y, y, animate)) {
            // obf: this.a(var4,var5,var6,var3,var2,var4,true,var3,var1+107) — this 9-arg `a`
            // overload has boolean at pos2/pos7, so it is walkToAction (obf a(int,boolean,...,boolean,int,int)),
            // NOT walkTo (obf a(int,int,byte,boolean,...)). Resolved by arg-type signature.
            m.walkToAction(scrollPos, animate, trackLen, y, x, scrollPos, true, y, sentinel + 107);
            if (sentinel != 10) {
                drawScrollbar2(99, 113, -126, -87, true, 125);
            }
        }
    }

    // -------------------------------------------------------------------------
    // drawScrollbar2  (obf: private void a(int,int,int,int,boolean,int)  @clean L11249)
    // -------------------------------------------------------------------------

    /**
     * Secondary scrollbar/slider widget. Single walkTo probe; resets the activity timer
     * unless trackLen==8.
     * obf: void a(int,int,int,int,boolean,int)
     */
    void drawScrollbar2(int x, int y, int w, int h, boolean animate, int trackLen) {
        // obf: this.a(var2, var5, var4, var1, var3, var2, false, var1, 105)
        //   boolean at pos2 (var5=animate) and pos7 (false) ⇒ this 9-arg `a` overload is
        //   walkToAction (obf a(int,boolean,int,int,int,int,boolean,int,int)), NOT walkTo.
        m.walkToAction(y, animate, h, x, w, y, false, x, 105);
        if (trackLen != 8) {
            m.lastActionTime = -85L;   // obf: Wi
        }
    }
}
