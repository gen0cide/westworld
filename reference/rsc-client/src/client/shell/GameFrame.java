package client.shell;

import java.awt.AWTEvent;
import java.awt.Frame;
import java.awt.Graphics;
import java.awt.event.MouseEvent;

/**
 * GameFrame — the top-level AWT window that hosts the RuneScape Classic client.
 *
 * <p>This is the heavyweight native window (an AWT {@link Frame}) that the
 * {@link GameShell} draws into when the client runs as a standalone application
 * rather than inside a browser applet. It is a thin host: all real rendering and
 * input handling lives in the {@code GameShell}, and this frame simply
 * forwards paint requests and mouse events to it.
 *
 * <p>Its only real responsibility is to compensate for the window's title bar /
 * decorations so that the game's logical (0,0) lines up with the top-left of the
 * drawable client area:
 * <ul>
 *   <li>{@link #resize(int, int)} grows the requested height by
 *       {@link #titleBarHeight} so the client area is the size the game asked for;</li>
 *   <li>{@link #getGraphics()} translates the returned {@link Graphics} so drawing
 *       at logical (0,0) lands below the title bar;</li>
 *   <li>{@link #processEvent(AWTEvent)} shifts incoming mouse Y coordinates up by
 *       the title-bar offset so the game sees client-area-relative positions.</li>
 * </ul>
 *
 * <p>Decompiled and deobfuscated from obfuscated class {@code qb} (Microsoft J++
 * build, rev ~233–235). The original was wrapped in the build's standard
 * obfuscation: per-method profiling counters, opaque {@code client.vh} predicates
 * (always false), and try/catch blocks that rethrow via {@code ErrorHandler.a}
 * with an encrypted method-signature string. All of that has been stripped here;
 * only the genuine logic remains.
 */
final class GameFrame extends Frame {

   // ---------------------------------------------------------------------------
   // Profiling counters (obfuscation artifact).
   //
   // In the obfuscated client each method incremented one of these static counters
   // on entry (e.g. paint() did ++b). They are pure telemetry/anti-analysis noise
   // and are NOT read by any game logic, so the methods below no longer touch them.
   // They are retained only to document the original class shape.
   // ---------------------------------------------------------------------------
   /** Profiling counter formerly incremented by {@link #paint(Graphics)} (obf {@code b}). */
   static int paintCallCount;
   /** Profiling counter formerly incremented by {@link #getGraphics()} (obf {@code f}). */
   static int getGraphicsCallCount;
   /** Profiling counter formerly incremented by {@link #resize(int, int)} (obf {@code i}). */
   static int resizeCallCount;
   /** Profiling counter formerly incremented by {@link #processEvent(AWTEvent)} (obf {@code a}). */
   static int processEventCallCount;

   // ---------------------------------------------------------------------------
   // Unused static scratch storage (obfuscation artifact).
   //
   // These static fields exist on the obfuscated class but are never referenced by
   // any GameFrame method; they are leftover/shared scratch buffers from the J++
   // build's class-merging and are kept here only for fidelity to the original.
   // ---------------------------------------------------------------------------
   /** Unused shared int buffer (obf {@code e}). */
   static int[] unusedIntBuffer;
   /** Unused shared 2D int buffer (obf {@code d}). */
   static int[][] unusedIntBuffer2d;
   /** Unused 100,000-byte shared scratch buffer (obf {@code k}). */
   static byte[] unusedByteBuffer = new byte[100000];

   /** The game shell hosted in this window; receives forwarded paint calls (obf {@code g}). */
   private GameShell gameShell;

   /**
    * Graphics translation mode (obf {@code c}).
    *
    * <p>Selects how {@link #getGraphics()} offsets the drawing origin:
    * <ul>
    *   <li>{@code 0} (default) — translate down by 24px to skip the title bar;</li>
    *   <li>non-zero — translate left by 5px instead (used for an alternate /
    *       borderless layout).</li>
    * </ul>
    * Always 0 in practice for this build, since the constructor never sets it.
    */
   private int translationMode;

   /** Requested client-area width, in pixels (obf {@code l}). */
   private int windowWidth;

   /**
    * Vertical padding added to the requested height to account for window
    * decorations (title bar), in pixels (obf {@code j}).
    *
    * <p>28px for a normal decorated window, 48px for the taller variant selected
    * by the constructor's {@code tallTitleBar} flag.
    */
   private int titleBarHeight = 28;

   /** Requested client-area height, in pixels (obf {@code h}). */
   private int windowHeight;

   /**
    * Vertical offset, in pixels, between the window's top edge and the start of
    * the drawable client area (below the title bar).
    *
    * <p>This is the magic constant the J++ build hard-codes for translating mouse
    * Y coordinates and the drawing origin. It corresponds to the visible title-bar
    * height on the target platform.
    */
   private static final int TITLE_BAR_Y_OFFSET = 24;

   /** Horizontal offset applied in the alternate translation mode, in pixels. */
   private static final int ALT_X_OFFSET = -5;

   /**
    * Decrypted method-signature strings used by the original obfuscated error
    * wrapper ({@code ErrorHandler.a(throwable, signature)}). The deobfuscated
    * methods below no longer wrap their bodies, so these are unused, but they are
    * preserved (decoded from the original two-pass XOR cipher) to document what
    * each catch block reported:
    * <pre>
    *   [0] = "{...}"               placeholder rendered when an argument is non-null
    *   [1] = "qb.&lt;init&gt;("    constructor
    *   [2] = "null"                rendered when an argument is null
    *   [3] = "qb.resize("          resize(int,int)
    *   [4] = "qb.getGraphics()"    getGraphics()
    *   [5] = "qb.processEvent("    processEvent(AWTEvent)
    *   [6] = "qb.paint("           paint(Graphics)
    * </pre>
    */
   private static final String[] ERROR_SIGNATURES = {
      "{...}",
      "qb.<init>(",
      "null",
      "qb.resize(",
      "qb.getGraphics()",
      "qb.processEvent(",
      "qb.paint(",
   };

   /**
    * Builds and shows the host window, then wires the given {@link GameShell} as
    * its painting/event target.
    *
    * @param shell        the game shell that owns this window and does the drawing (obf {@code var1})
    * @param width        requested client-area width in pixels (obf {@code var2})
    * @param height       requested client-area height in pixels (obf {@code var3})
    * @param title        window title bar text (obf {@code var4})
    * @param resizable    whether the user may resize the window (obf {@code var5})
    * @param tallTitleBar if {@code true} use the 48px title-bar padding, else 28px (obf {@code var6})
    */
   GameFrame(GameShell shell, int width, int height, String title, boolean resizable, boolean tallTitleBar) {
      // translationMode is left at its default 0 (down-by-24 offset).
      this.translationMode = 0;
      this.gameShell = shell;

      // Pick the decoration padding: taller windows reserve 48px, normal ones 28px.
      if (tallTitleBar) {
         this.titleBarHeight = 48;
      } else {
         this.titleBarHeight = 28;
      }

      this.windowWidth = width;
      this.windowHeight = height;
      this.setTitle(title);
      this.setResizable(resizable);
      this.show();        // make the native window visible (deprecated AWT API, original used show())
      this.toFront();     // raise it above other windows
      this.resize(this.windowWidth, this.windowHeight); // size client area (adds titleBarHeight)
      this.getGraphics(); // prime/realize the graphics context
   }

   /**
    * Intercepts AWT events before normal dispatch; for mouse events it rebases the
    * Y coordinate from window-relative to client-area-relative by subtracting the
    * title-bar offset, so the game sees mouse positions matching its drawing origin.
    *
    * @param event the incoming AWT event (obf {@code var1})
    */
   @Override
   protected final void processEvent(AWTEvent event) {
      if (event instanceof MouseEvent) {
         MouseEvent mouse = (MouseEvent) event;
         // Shift Y up by the title-bar height so (0,0) is the top of the client area.
         event = new MouseEvent(
            mouse.getComponent(),
            mouse.getID(),
            mouse.getWhen(),
            mouse.getModifiers(),
            mouse.getX(),
            mouse.getY() - TITLE_BAR_Y_OFFSET,
            mouse.getClickCount(),
            mouse.isPopupTrigger()
         );
      }

      super.processEvent(event);
   }

   /**
    * AWT paint callback; delegates all drawing to the hosted {@link GameShell}.
    *
    * @param graphics the graphics context to draw into (obf {@code var1})
    */
   @Override
   public final void paint(Graphics graphics) {
      this.gameShell.paint(graphics);
   }

   /**
    * Resizes the native window. Overrides the deprecated {@link Frame#resize(int, int)}
    * to add {@link #titleBarHeight} to the requested height, so that the drawable
    * client area ends up the size the caller actually asked for (the extra pixels
    * are consumed by the title bar / decorations).
    *
    * @param width  requested client-area width in pixels (obf {@code var1})
    * @param height requested client-area height in pixels (obf {@code var2})
    */
   @Override
   public final void resize(int width, int height) {
      super.resize(width, this.titleBarHeight + height);
   }

   /**
    * Returns a {@link Graphics} for the window whose origin has been translated so
    * that drawing at logical (0,0) lands at the top-left of the client area
    * (below the title bar).
    *
    * <ul>
    *   <li>{@link #translationMode} == 0 (the default): translate down by 24px;</li>
    *   <li>{@link #translationMode} != 0: translate left by 5px instead.</li>
    * </ul>
    *
    * @return the origin-translated graphics context
    */
   @Override
   public final Graphics getGraphics() {
      Graphics graphics = super.getGraphics();
      if (this.translationMode == 0) {
         // Normal case: shift the drawing origin below the title bar.
         graphics.translate(0, TITLE_BAR_Y_OFFSET);
      } else {
         // Alternate layout: shift left instead of down.
         graphics.translate(ALT_X_OFFSET, 0);
      }

      return graphics;
   }

   /**
    * First pass of the original string-obfuscation decode (obf {@code z(String)}).
    *
    * <p>Converts a literal to a char array and, only for single-character literals,
    * XORs the first char with {@code 'v'} (0x76). For all of this class's literals
    * (length &gt;= 2) it is a plain {@code toCharArray()}. Retained for fidelity;
    * the decoded results are baked into {@link #ERROR_SIGNATURES}.
    *
    * @param literal the obfuscated source literal (obf {@code var0})
    * @return the literal as a char array, with the conditional first-char flip applied
    */
   private static char[] decodeStringPass1(String literal) {
      char[] chars = literal.toCharArray();
      if (chars.length < 2) {
         chars[0] = (char) (chars[0] ^ 'v');
      }

      return chars;
   }

   /**
    * Second pass of the original string-obfuscation decode (obf {@code z(char[])}).
    *
    * <p>XORs each character with a 5-element repeating key {65, 52, 67, 40, 118}
    * (cycled by {@code index % 5}) and interns the result. This is the pass that
    * produces the human-readable method-signature strings in {@link #ERROR_SIGNATURES}.
    *
    * @param chars the char array from {@link #decodeStringPass1(String)} (obf {@code var0})
    * @return the decoded, interned string
    */
   private static String decodeStringPass2(char[] chars) {
      // Repeating XOR key, selected per-position by (index % 5).
      for (int i = 0; i < chars.length; i++) {
         byte key;
         switch (i % 5) {
            case 0:
               key = 65;
               break;
            case 1:
               key = 52;
               break;
            case 2:
               key = 67;
               break;
            case 3:
               key = 40;
               break;
            default:
               key = 118;
         }

         chars[i] = (char) (chars[i] ^ key);
      }

      return new String(chars).intern();
   }
}
