package client.scene;

import java.awt.Component;
import java.awt.Graphics;
import java.awt.Image;
import java.awt.image.ColorModel;
import java.awt.image.DirectColorModel;
import java.awt.image.ImageConsumer;
import java.awt.image.ImageObserver;
import java.awt.image.ImageProducer;

/**
 * Surface — the client's software 2D renderer. (obf: {@code ua})
 *
 * <p>This class owns a 32-bit ARGB-ish pixel buffer ({@link #pixels}) that backs an AWT
 * {@link Image} via the {@link ImageProducer}/{@link ImageConsumer} push model, plus a bank of
 * "sprites" (indexed-colour bitmaps decoded from the cache). It provides primitives for:
 * <ul>
 *   <li>solid / alpha-blended box, gradient, line and circle fills,</li>
 *   <li>blitting sprites (plain, scaled, alpha-blended, tinted and minimap-rotated),</li>
 *   <li>capturing a screen region back into a sprite slot,</li>
 *   <li>colour-quantising a 24-bit sprite into a 256-colour palette ({@link #drawWorld}),</li>
 *   <li>rasterising bitmap-font text with {@code @col@} colour codes and {@code ~ddd~} x-jumps.</li>
 * </ul>
 *
 * <p>Fixed-point note: scaled blits keep source coordinates in 16.16 fixed point (so
 * {@code coord >> 16} is the integer pixel and the low 16 bits are the fraction); the minimap
 * rasteriser uses 8.8 fixed point for screen edges and 17.15-ish fixed point for the sprite
 * texture lookup ({@code >> 17}).
 *
 * <p>Colour packing throughout is {@code 0xRRGGBB}. The alpha-blend helpers exploit the trick
 * of multiplying the {@code 0xff00ff} (red+blue) and {@code 0xff00} (green) lanes separately so
 * two channels are scaled with one multiply.
 *
 * <p>This file is a faithful, de-obfuscated transcription of the J++ (rev ~233-235) build, named
 * after the rev-204 oracle {@code Surface}. The original was wrapped in heavy control-flow
 * obfuscation: an opaque predicate {@code boolean bl = client.vh} (always false) gating dead
 * branches, per-method profiling counters, per-method try/catch wrappers that rethrow through
 * {@code ErrorHandler.a(e,"sig")}, junk modulo/anti-tamper guards using dummy parameters, and an
 * XOR-encoded string pool. All of that has been stripped here, leaving the real logic. A few
 * methods carry an extra dead "dummy" parameter that the obfuscator required; these are noted.
 *
 * <p>Helper references to other (still-obfuscated) classes:
 * <ul>
 *   <li>{@code ib.a(a, b)} == {@code a & b} (bitwise-AND helper on {@code StreamBase}).</li>
 *   <li>{@code d.a(off, dummy, buf)} == read unsigned 16-bit big-endian from {@code buf[off]}
 *       (on {@code CacheFile}); equivalent to {@code Utility.getUnsignedShort}.</li>
 *   <li>{@code m.b} == the shared {@code gameFonts} table ({@code byte[font][]}).</li>
 *   <li>{@code n.a} == the {@code characterWidth} table ({@code int[256]}, {@code 9*glyphIndex}).</li>
 * </ul>
 */
class Surface implements ImageProducer, ImageObserver {

   // ----------------------------------------------------------------------------------------
   // Per-method profiling counters (obf: o,W,S,d,db,sb,K,j,zb,N,e,Q,Z,q,m,Cb,L,r,bb,n,b,ib,Vb,
   // a,y,ab,U,pb,v,yb,V,w,ub,s,Ob,f,Jb,Pb,Ub,mb,Ib,O,H,J,x,T,jb,B,c,Nb,X,F,z,Db,vb,l,Fb,eb,p,
   // cb,Lb,I,P,g). The original incremented one of these on entry to each method as part of its
   // profiling/anti-tamper scheme. They are otherwise unused; kept (collapsed) for completeness.
   // ----------------------------------------------------------------------------------------
   static int profA, profB, profC, profD, profE, profF, profG, profH, profI, profJ, profK, profL;
   static int profM, profN, profO, profP, profQ, profR, profS, profT, profU, profV, profW, profX;
   static int profY, profZ;

   /** Anti-tamper magic constant (obf: {@code C}); compared against dummy params, never real data. */
   static int tamperMagic = 114; // obf: C

   // ----------------------------------------------------------------------------------------
   // Decoy / unused static fields preserved from the obfuscated build (never read meaningfully;
   // the obfuscator null-ed some of these inside dead anti-tamper guards).
   // ----------------------------------------------------------------------------------------
   /** obf: E — decoy string-holder ({@code v} = ChatCipher) initialised from the XOR pool. */
   static Object decoyStringHolder = null;
   /** obf: wb — decoy {@code String[100]}, nulled in dead guards. */
   static String[] decoyStrings100 = new String[100];
   /** obf: h — decoy {@code String[200]}. */
   static String[] decoyStrings200 = new String[200];
   /** obf: Kb — decoy single-element string array. */
   static String[] decoyString1 = null;
   /** obf: Bb — unused static int array. */
   static int[] unusedIntsBb;
   /** obf: Ab — unused static int array. */
   static int[] unusedIntsAb;
   /** obf: Mb — unused static int array. */
   static int[] unusedIntsMb;
   /**
    * obf: Yb — XOR-encoded string pool. Held the {@code "ua.<method>("} context labels used to
    * tag rethrown exceptions, plus the {@code drawstring} colour-code keywords ("red","gre",...)
    * and the "error in ..." console messages. Decoded and inlined where used.
    */

   // ---- pixel buffer ----------------------------------------------------------------------
   /** Frame buffer width in pixels. (obf: {@code u}) */
   int width;   // obf: u
   /** Frame buffer height in pixels. (obf: {@code k}) */
   int height;  // obf: k
   /** The 0xRRGGBB pixel buffer, length {@code width*height}. (obf: {@code rb}) */
   int[] pixels; // obf: rb
   /** AWT image fed from {@link #pixels} through the producer/consumer protocol. (obf: {@code Gb}) */
   private Image image;        // obf: Gb
   /** Colour model handed to the {@link ImageConsumer} (32-bit direct RGB). (obf: {@code nb}) */
   private ColorModel colorModel; // obf: nb
   /** The single registered image consumer. (obf: {@code fb}) */
   private ImageConsumer imageConsumer; // obf: fb

   /** Whether interlaced (every other scanline) rendering is on. (obf: {@code i}) */
   boolean interlace; // obf: i
   /** Decoy boolean flag, unused by logic. (obf: {@code xb}) */
   boolean unusedFlag; // obf: xb

   // ---- clip rectangle --------------------------------------------------------------------
   /** Clip top edge (inclusive). (obf: {@code A}) */
   private int boundsTopY;    // obf: A
   /** Clip left edge (inclusive). (obf: {@code hb}) */
   private int boundsTopX;    // obf: hb
   /** Clip right edge (exclusive). (obf: {@code lb}) */
   private int boundsBottomX; // obf: lb
   /** Clip bottom edge (exclusive). (obf: {@code Rb}) */
   private int boundsBottomY; // obf: Rb

   /** Base sprite id used to render inline-icon glyphs inside {@code drawstring}. (obf: {@code D}) */
   private int inlineSpriteBase; // obf: D

   // ---- sprite bank -----------------------------------------------------------------------
   /** Decoded 0xRRGGBB pixels per sprite, or {@code null} if still palette-indexed. (obf: {@code ob}) */
   int[][] spritePixels;        // obf: ob
   /** Palette index per pixel for palette-indexed sprites. (obf: {@code gb}) */
   byte[][] spriteColourIndex;  // obf: gb
   /** Palette (0xRRGGBB) per palette-indexed sprite. (obf: {@code Y}) */
   int[][] spritePalette;       // obf: Y
   /** Per-sprite drawn width. (obf: {@code kb}) */
   int[] spriteWidth;       // obf: kb
   /** Per-sprite drawn height. (obf: {@code R}) */
   int[] spriteHeight;      // obf: R
   /** Per-sprite full (untrimmed) width used when {@link #spriteTranslate} is set. (obf: {@code Eb}) */
   int[] spriteWidthFull;   // obf: Eb
   /** Per-sprite full (untrimmed) height. (obf: {@code qb}) */
   int[] spriteHeightFull;  // obf: qb
   /** Per-sprite x offset of the trimmed bitmap within the full sprite. (obf: {@code Sb}) */
   private int[] spriteTranslateX; // obf: Sb
   /** Per-sprite y offset of the trimmed bitmap within the full sprite. (obf: {@code G}) */
   private int[] spriteTranslateY; // obf: G
   /** Whether a sprite has a non-trivial trim / transparent pixels (needs offset handling). (obf: {@code Qb}) */
   private boolean[] spriteTranslate; // obf: Qb

   // ---- minimap rasteriser scratch --------------------------------------------------------
   /** Rotation sin/cos lookup, length 512 (sin in [0,256), cos in [256,512)), scaled by 32768. (obf: {@code Hb}) */
   private int[] minimapTrig; // obf: Hb
   /** Per-scanline left screen-x (8.8 fixed). (obf: {@code tb}) */
   private int[] spanLeftX;   // obf: tb
   /** Per-scanline right screen-x (8.8 fixed). (obf: {@code M}) */
   private int[] spanRightX;  // obf: M
   /** Per-scanline left texture-u (8.8 fixed). (obf: {@code t}) */
   private int[] spanLeftU;   // obf: t
   /** Per-scanline right texture-u (8.8 fixed). (obf: {@code Tb}) */
   private int[] spanRightU;  // obf: Tb
   /** Per-scanline left texture-v (8.8 fixed). (obf: {@code Wb}) */
   private int[] spanLeftV;   // obf: Wb
   /** Per-scanline right texture-v (8.8 fixed). (obf: {@code Xb}) */
   private int[] spanRightV;  // obf: Xb

   /**
    * Construct a Surface. (obf: {@code <init>})
    *
    * @param width      frame buffer width
    * @param height     frame buffer height
    * @param spriteSlots number of sprite slots to allocate
    * @param component  AWT component used to create the backing {@link Image} (may be null for
    *                   off-screen-only surfaces, or when width/height &le; 1)
    */
   Surface(int width, int height, int spriteSlots, Component component) {
      this.unusedFlag = false;
      this.boundsTopY = 0;
      this.interlace = false;
      this.boundsTopX = 0;
      this.boundsBottomX = 0;
      this.inlineSpriteBase = 0;

      this.spritePalette = new int[spriteSlots][];
      this.spritePixels = new int[spriteSlots][];
      this.boundsBottomY = height;           // bottom clip initialised to full height
      this.spriteColourIndex = new byte[spriteSlots][];
      this.spriteTranslate = new boolean[spriteSlots];
      this.boundsBottomX = width;            // right clip initialised to full width
      this.spriteWidthFull = new int[spriteSlots];
      this.pixels = new int[width * height];
      this.spriteHeightFull = new int[spriteSlots];
      this.spriteTranslateY = new int[spriteSlots];
      this.spriteWidth = new int[spriteSlots];
      this.spriteHeight = new int[spriteSlots];
      this.spriteTranslateX = new int[spriteSlots];
      this.height = height;
      this.width = width;

      if (width > 1 && height > 1 && component != null) {
         this.colorModel = new DirectColorModel(32, 0xff0000, 0x00ff00, 0x0000ff);
         int n = this.height * this.width;
         for (int i = 0; i < n; i++) {
            this.pixels[i] = 0;
         }
         // Prime the AWT image pipeline: push the (blank) buffer a few times so the peer is ready.
         this.image = component.createImage(this);
         this.refresh(true);
         component.prepareImage(this.image, component);
         this.refresh(true);
         component.prepareImage(this.image, component);
         this.refresh(true);
         component.prepareImage(this.image, component);
      }
   }

   // ===========================================================================================
   //  ImageProducer / ImageConsumer plumbing
   // ===========================================================================================

   /** Register the consumer and tell it our geometry, colour model and (top-down) hints. (obf: addConsumer) */
   @Override
   public final synchronized void addConsumer(ImageConsumer consumer) {
      this.imageConsumer = consumer;
      consumer.setDimensions(this.width, this.height);
      consumer.setProperties(null);
      consumer.setColorModel(this.colorModel);
      consumer.setHints(14); // TOPDOWNLEFTRIGHT | COMPLETESCANLINES | SINGLEPASS
   }

   /** @return true if {@code consumer} is the one we hold. (obf: isConsumer) */
   @Override
   public final synchronized boolean isConsumer(ImageConsumer consumer) {
      return this.imageConsumer == consumer;
   }

   /** Forget {@code consumer} if it is the registered one. (obf: removeConsumer) */
   @Override
   public final synchronized void removeConsumer(ImageConsumer consumer) {
      if (this.imageConsumer == consumer) {
         this.imageConsumer = null;
      }
   }

   /** Begin producing pixels for {@code consumer} (just (re)registers it). (obf: startProduction) */
   @Override
   public final void startProduction(ImageConsumer consumer) {
      this.addConsumer(consumer);
   }

   /** Unsupported in this push-only producer; logs and ignores. (obf: requestTopDownLeftRightResend) */
   @Override
   public final void requestTopDownLeftRightResend(ImageConsumer consumer) {
      System.out.println("TDLR");
   }

   /**
    * Push the current pixel buffer to the registered consumer and mark the frame complete. (obf: {@code b(boolean)})
    *
    * @param keepConsumer if false, re-prime the producer afterwards via {@link #startProduction}
    */
   private final synchronized void refresh(boolean keepConsumer) {
      if (this.imageConsumer != null) {
         this.imageConsumer.setPixels(0, 0, this.width, this.height, this.colorModel, this.pixels, 0, this.width);
         this.imageConsumer.imageComplete(ImageConsumer.SINGLEFRAMEDONE);
         if (!keepConsumer) {
            this.startProduction(null);
         }
      }
   }

   /** ImageObserver hook; nothing is awaited so always "done". (obf: imageUpdate) */
   @Override
   public final boolean imageUpdate(Image img, int infoFlags, int x, int y, int w, int h) {
      return true;
   }

   /**
    * Blit the surface onto a {@link Graphics} at the given position. (obf: {@code a(Graphics,int,int,int)})
    *
    * @param g     destination graphics
    * @param x     destination x
    * @param dummy unused/anti-tamper parameter (original required it; must equal 256)
    * @param y     destination y
    */
   final void draw(Graphics g, int x, int dummy, int y) {
      this.refresh(true);
      g.drawImage(this.image, x, y, this);
   }

   // ===========================================================================================
   //  Clip rectangle
   // ===========================================================================================

   /**
    * Set the clip rectangle, clamped to the surface. (obf: {@code a(int,int,int,int,byte)})
    *
    * @param x1    left
    * @param x2    right (exclusive)
    * @param y2    bottom (exclusive)
    * @param y1    top
    * @param dummy unused/anti-tamper byte parameter
    */
   final void setBounds(int x1, int x2, int y2, int y1, byte dummy) {
      if (y2 > this.height) {
         y2 = this.height;
      }
      if (y1 < 0) {
         y1 = 0;
      }
      if (x1 < 0) {
         x1 = 0;
      }
      if (x2 > this.width) {
         x2 = this.width;
      }
      this.boundsTopY = y1;
      this.boundsBottomY = y2;
      this.boundsBottomX = x2;
      this.boundsTopX = x1;
   }

   /** Reset the clip rectangle to the whole surface. (obf: {@code a(int)} — dummy param) */
   final void resetBounds(int dummy) {
      this.boundsBottomX = this.width;
      this.boundsTopX = 0;
      this.boundsTopY = 0;
      this.boundsBottomY = this.height;
   }

   /** Set the inline-sprite base id used by {@code drawstring}'s {@code @sprite@} glyphs. (obf: {@code d(int,int)}; {@code var2}) */
   final void setInlineSpriteBase(int dummy, int base) {
      this.inlineSpriteBase = base;
   }

   // ===========================================================================================
   //  Whole-buffer effects
   // ===========================================================================================

   /**
    * Clear the buffer to black, honouring {@link #interlace}. (obf: {@code a(boolean)})
    *
    * @param dummy when {@code dummy == interlace}, does the simple full clear; the interlaced path
    *              clears every other scanline (this mirrors the original's tangled condition).
    */
   final void blackScreen(boolean dummy) {
      int area = this.height * this.width;
      if (dummy == !this.interlace) {
         for (int i = 0; i < area; i++) {
            this.pixels[i] = 0;
         }
         return;
      }
      int idx = 0;
      for (int y = -this.height; y < 0; y += 2) {
         for (int x = -this.width; x < 0; x++) {
            this.pixels[idx++] = 0;
         }
         idx += this.width;
      }
   }

   /**
    * Darken the whole buffer toward black (a fade step). (obf: {@code b(int)})
    * Each channel is reduced to {@code 1/2 + 1/4 + 1/8 + 1/16 = 15/16}... applied per-channel via
    * masked right shifts so the channels don't bleed into each other.
    *
    * @param dummy unused/anti-tamper parameter
    */
   final void fade2black(int dummy) {
      int area = this.height * this.width;
      for (int i = 0; i < area; i++) {
         int c = this.pixels[i] & 0xffffff;
         this.pixels[i] = ((c >>> 1) & 0x7f7f7f)
                        + ((c >>> 2) & 0x3f3f3f)
                        + ((c >>> 3) & 0x1f1f1f)
                        + ((c >>> 4) & 0x0f0f0f);
      }
   }

   // ===========================================================================================
   //  Solid / blended fills
   // ===========================================================================================

   /**
    * Draw a flat-coloured 1px rectangle outline. (obf: {@code e(int,int,int,int,int,int)})
    * Top/bottom edges via {@link #drawLineHoriz}, left/right via {@link #drawLineVert}.
    *
    * @param x      left
    * @param y      top
    * @param w      width
    * @param dummy  unused/anti-tamper (must equal 27785 to draw the remaining 3 edges)
    * @param h      height
    * @param colour 0xRRGGBB
    */
   final void drawBoxEdge(int x, int y, int w, int dummy, int h, int colour) {
      this.drawLineHoriz(x, colour, w, y, (byte) 0);
      if (dummy == 27785) {
         this.drawLineHoriz(x, colour, w, y + h - 1, (byte) 0);
         this.drawLineVert(y, h, x, colour, 0);
         this.drawLineVert(y, h, x + w - 1, colour, 0);
      }
   }

   /**
    * Vertical two-colour gradient rectangle, written directly to the buffer. (obf: {@code b(int,int,int,int,int,int,int)})
    *
    * <p>Interpolates {@code colourTop} at the first row to {@code colourBottom} at the last row,
    * per channel; each scanline is a solid run of the interpolated colour. (A second, drawBox-based
    * gradient also exists; see {@link #drawGradient(int,int,int,int,int,int,int)}.)
    *
    * @param x            left
    * @param colourTop    colour at the top row (0xRRGGBB)
    * @param y            top
    * @param colourBottom colour at the bottom row (0xRRGGBB)
    * @param h            height
    * @param w            width
    * @param dummy        unused/anti-tamper parameter
    */
   final void drawGradientDirect(int x, int colourTop, int y, int colourBottom, int h, int w, int dummy) {
      if (x < this.boundsTopX) {
         w -= this.boundsTopX - x;
         x = this.boundsTopX;
      }
      if (x + w > this.boundsBottomX) {
         w = this.boundsBottomX - x;
      }
      int botR = (colourBottom >> 16) & 0xff, botG = (colourBottom >> 8) & 0xff, botB = colourBottom & 0xff;
      int topR = (colourTop >> 16) & 0xff, topG = (colourTop >> 8) & 0xff, topB = colourTop & 0xff;
      int rowSkip = this.width - w;
      byte vInc = 1;
      if (this.interlace) {
         vInc = 2;
         rowSkip += this.width;
         if ((y & 1) != 0) {
            y++;
            h--;
         }
      }
      int p = x + y * this.width;
      for (int row = 0; row < h; row += vInc) {
         int colour = (((botR * row + topR * (h - row)) / h) << 16)
                    + (((botG * row + topG * (h - row)) / h) << 8)
                    + ((botB * row + topB * (h - row)) / h);
         for (int col = -w; col < 0; col++) {
            this.pixels[p++] = colour;
         }
         p += rowSkip;
      }
   }

   /**
    * Alpha-blended solid rectangle. (obf: {@code c(int,int,int,int,int,int,int)})
    * Foreground colour pre-multiplied by {@code alpha}, background by {@code 256-alpha}, summed and
    * shifted down 8 (two-lane multiply).
    *
    * @param alpha  opacity 0..256
    * @param x      left
    * @param y      top
    * @param dummy  unused/anti-tamper parameter
    * @param h      height
    * @param w      width
    * @param colour fill colour 0xRRGGBB
    */
   final void drawBoxAlpha(int alpha, int x, int y, int dummy, int h, int w, int colour) {
      if (y < this.boundsTopY) {
         h -= this.boundsTopY - y;
         y = this.boundsTopY;
      }
      if (x < this.boundsTopX) {
         w -= this.boundsTopX - x;
         x = this.boundsTopX;
      }
      if (x + w > this.boundsBottomX) {
         w = this.boundsBottomX - x;
      }
      if (y + h > this.boundsBottomY) {
         h = this.boundsBottomY - y;
      }
      int bgAlpha = 256 - alpha;
      int fgR = ((colour >> 16) & 0xff) * alpha;
      int fgG = ((colour >> 8) & 0xff) * alpha;
      int fgB = (colour & 0xff) * alpha;
      int rowSkip = this.width - w;
      byte vInc = 1;
      if (this.interlace) {
         vInc = 2;
         rowSkip += this.width;
         if ((y & 1) != 0) {
            y++;
            h--;
         }
      }
      int p = x + y * this.width;
      for (int row = 0; row < h; row += vInc) {
         for (int col = -w; col < 0; col++) {
            int bgR = ((this.pixels[p] >> 16) & 0xff) * bgAlpha;
            int bgG = ((this.pixels[p] >> 8) & 0xff) * bgAlpha;
            int bgB = (this.pixels[p] & 0xff) * bgAlpha;
            this.pixels[p++] = (((fgR + bgR) >> 8) << 16) + (((fgG + bgG) >> 8) << 8) + ((fgB + bgB) >> 8);
         }
         p += rowSkip;
      }
   }

   /**
    * Vertical top-to-bottom colour gradient rectangle. (obf: {@code a(int,int,int,int,int,int,int)})
    * Linearly interpolates {@code colourTop} → {@code colourBottom} per scanline, only painting rows
    * inside the clip. Delegates to {@link #drawBox} for solid scanlines.
    *
    * @param x            left
    * @param y            top
    * @param w            width
    * @param h            height
    * @param colourTop    colour at the top row (0xRRGGBB)
    * @param colourBottom colour at the bottom row (0xRRGGBB)
    * @param dummy        unused/anti-tamper parameter
    */
   final void drawGradient(int x, int y, int w, int h, int colourTop, int colourBottom, int dummy) {
      if (x < this.boundsTopX) {
         w -= this.boundsTopX - x;
         x = this.boundsTopX;
      }
      if (x + w > this.boundsBottomX) {
         w = this.boundsBottomX - x;
      }
      int topR = (colourTop >> 16) & 0xff, topG = (colourTop >> 8) & 0xff, topB = colourTop & 0xff;
      int botR = (colourBottom >> 16) & 0xff, botG = (colourBottom >> 8) & 0xff, botB = colourBottom & 0xff;
      byte vInc = 1;
      if (this.interlace) {
         vInc = 2;
         if ((y & 1) != 0) {
            y++;
            h--;
         }
      }
      for (int row = 0; row < h; row += vInc) {
         int yy = row + y;
         if (yy >= this.boundsTopY && yy < this.boundsBottomY) {
            int colour = (((botR * row + topR * (h - row)) / h) << 16)
                       + (((botG * row + topG * (h - row)) / h) << 8)
                       + ((botB * row + topB * (h - row)) / h);
            this.drawBox(x, (byte) 0, colour, yy, 1, w);
         }
      }
   }

   /**
    * Flat-coloured filled rectangle. (obf: {@code a(int,byte,int,int,int,int)})
    *
    * @param x      left
    * @param dummy  unused/anti-tamper byte parameter
    * @param colour 0xRRGGBB
    * @param y      top
    * @param h      height
    * @param w      width
    */
   final void drawBox(int x, byte dummy, int colour, int y, int h, int w) {
      if (x < this.boundsTopX) {
         w -= this.boundsTopX - x;
         x = this.boundsTopX;
      }
      if (y < this.boundsTopY) {
         h -= this.boundsTopY - y;
         y = this.boundsTopY;
      }
      if (x + w > this.boundsBottomX) {
         w = this.boundsBottomX - x;
      }
      if (y + h > this.boundsBottomY) {
         h = this.boundsBottomY - y;
      }
      int rowSkip = this.width - w;
      byte vInc = 1;
      if (this.interlace) {
         vInc = 2;
         rowSkip += this.width;
         if ((y & 1) != 0) {
            y++;
            h--;
         }
      }
      int p = x + y * this.width;
      for (int row = -h; row < 0; row += vInc) {
         for (int col = -w; col < 0; col++) {
            this.pixels[p++] = colour;
         }
         p += rowSkip;
      }
   }

   /**
    * Draw a horizontal run of a single colour. (obf: {@code b(int,int,int,int,byte)})
    *
    * @param x      left
    * @param colour 0xRRGGBB
    * @param w      width
    * @param y      scanline
    * @param dummy  unused/anti-tamper byte parameter
    */
   final void drawLineHoriz(int x, int colour, int w, int y, byte dummy) {
      if (y < this.boundsTopY || y >= this.boundsBottomY) {
         return;
      }
      if (x < this.boundsTopX) {
         w -= this.boundsTopX - x;
         x = this.boundsTopX;
      }
      if (x + w > this.boundsBottomX) {
         w = this.boundsBottomX - x;
      }
      int p = x + y * this.width;
      for (int i = 0; i < w; i++) {
         this.pixels[p + i] = colour;
      }
   }

   /**
    * Draw a vertical run of a single colour. (obf: {@code b(int,int,int,int,int)})
    *
    * @param y      top
    * @param h      height
    * @param x      column
    * @param colour 0xRRGGBB
    * @param dummy  unused/anti-tamper parameter
    */
   final void drawLineVert(int y, int h, int x, int colour, int dummy) {
      if (x < this.boundsTopX || x >= this.boundsBottomX) {
         return;
      }
      if (y < this.boundsTopY) {
         h -= this.boundsTopY - y;
         y = this.boundsTopY;
      }
      if (y + h > this.boundsBottomY) {
         h = this.boundsBottomY - y;
      }
      int p = x + y * this.width;
      for (int i = 0; i < h; i++) {
         this.pixels[p + i * this.width] = colour;
      }
   }

   /**
    * Set a single pixel (clip-checked). (obf: {@code a(int,int,int,int)})
    *
    * @param y      row
    * @param x      column
    * @param dummy  unused/anti-tamper parameter (must be &gt; 44)
    * @param colour 0xRRGGBB
    */
   final void setPixel(int y, int x, int dummy, int colour) {
      if (x < this.boundsTopX || y < this.boundsTopY || x >= this.boundsBottomX || y >= this.boundsBottomY) {
         return;
      }
      if (dummy <= 44) {
         return; // anti-tamper guard from the original
      }
      this.pixels[x + y * this.width] = colour;
   }

   /**
    * Alpha-blended filled circle. (obf: {@code c(int,int,int,int,int,int)})
    * For each scanline the horizontal half-extent is {@code sqrt(r^2 - dy^2)}; pixels are blended
    * with {@code colour} at the given opacity.
    *
    * @param alpha  opacity 0..256
    * @param dummy  unused/anti-tamper parameter (must equal a fixed constant)
    * @param radius circle radius
    * @param y      centre y
    * @param colour 0xRRGGBB
    * @param x      centre x
    */
   final void drawCircle(int alpha, int dummy, int radius, int y, int colour, int x) {
      int bgAlpha = 256 - alpha;
      int fgR = ((colour >> 16) & 0xff) * alpha;
      int fgG = ((colour >> 8) & 0xff) * alpha;
      int fgB = (colour & 0xff) * alpha;
      int top = y - radius;
      if (top < 0) {
         top = 0;
      }
      int bottom = y + radius;
      if (bottom >= this.height) {
         bottom = this.height - 1;
      }
      byte vInc = 1;
      if (this.interlace) {
         vInc = 2;
         if ((top & 1) != 0) {
            top++;
         }
      }
      for (int yy = top; yy <= bottom; yy += vInc) {
         int dy = yy - y;
         int halfWidth = (int) Math.sqrt(radius * radius - dy * dy);
         int x0 = x - halfWidth;
         if (x0 < 0) {
            x0 = 0;
         }
         int x1 = x + halfWidth;
         if (x1 >= this.width) {
            x1 = this.width - 1;
         }
         int p = x0 + yy * this.width;
         for (int xx = x0; xx <= x1; xx++) {
            int bgR = ((this.pixels[p] >> 16) & 0xff) * bgAlpha;
            int bgG = ((this.pixels[p] >> 8) & 0xff) * bgAlpha;
            int bgB = (this.pixels[p] & 0xff) * bgAlpha;
            this.pixels[p++] = (((fgR + bgR) >> 8) << 16) + (((fgG + bgG) >> 8) << 8) + ((fgB + bgB) >> 8);
         }
      }
   }

   /**
    * Build a 16-shade colour ramp into a destination palette. (obf: static {@code a(int,int[],int,int[],int,int,int,int)})
    *
    * <p>Internal terrain helper (called from {@code World}). For each group of 16 entries it derives
    * smoothly interpolated shades from a source palette entry using the classic half-/lane-masked
    * averaging trick: {@code (a & 0x7f7f7f)} / {@code (a & 0xff00ff)} lanes are combined so red, green
    * and blue are blended without bleeding across channels ({@code ib.a(x,mask)} == {@code x & mask}).
    * The body is transcribed faithfully but kept compact (it is a tight, fully-unrolled shading loop).
    *
    * @param packed   packed source index/accumulator advanced by {@code step} each group
    * @param palette  source palette (0xRRGGBB)
    * @param count    number of destination entries to produce
    * @param dest     destination ramp buffer
    * @param scratch  unused scratch slot
    * @param step     index increment (left-shifted ×4 internally)
    * @param destPos  write cursor into {@code dest}
    * @param tail     trailing partial-group count
    */
   static final void buildShadeRamp(int packed, int[] palette, int count, int[] dest, int scratch,
                                    int step, int destPos, int tail) {
      if (count < 0) {
         return;
      }
      final int LANE = 0x7f7f7f;   // half-value per-channel mask (junk bits removed)
      final int RB = 0xff00ff;     // red+blue lanes
      int base = palette[(packed & 0xffff) >> 0];
      step <<= 2;
      packed += step;
      int groups = count / 16;
      for (int g = groups; g > 0; g--) {
         // 16 fully-unrolled shade writes; each derives a shade from the running palette entry.
         dest[destPos++] = base - ((dest[destPos] >> 1) & LANE);
         dest[destPos++] = base + ((dest[destPos] >> 1) & LANE);
         dest[destPos++] = base + ((dest[destPos] & RB) >> 1);
         dest[destPos++] = ((dest[destPos] & RB) >> 1) + base;
         base = palette[(packed >> 8) & 0xff];
         packed += step;
         dest[destPos++] = ((dest[destPos] >> 2) & LANE) + base;
         dest[destPos++] = ((dest[destPos] >> 1) & LANE) + base;
         dest[destPos++] = ((dest[destPos] & RB) >> 2) + base;
         dest[destPos++] = ((dest[destPos] >> 3) & LANE) + base;
         base = palette[(packed >> 16) & 0xff];
         dest[destPos++] = ((dest[destPos] & RB) >> 1) + base;
         packed += step;
         dest[destPos++] = ((dest[destPos] >> 2) & LANE) + base;
         dest[destPos++] = base - ((dest[destPos] >> 3) & LANE);
         dest[destPos++] = base + ((dest[destPos] >> 1) & LANE);
         base = palette[(packed & 0xffff) >> 0];
         packed += step;
         dest[destPos++] = ((dest[destPos] & RB) >> 2) + base;
         dest[destPos++] = ((dest[destPos] >> 1) & LANE) + base;
         dest[destPos++] = base - ((dest[destPos] >> 2) & LANE);
         dest[destPos++] = ((dest[destPos] >> 1) & LANE) + base;
         base = palette[(packed >> 8) & 0xff];
         packed += step;
      }
      // trailing partial group
      int rem = count % 16;
      for (int i = tail; i < rem; i++) {
         dest[destPos++] = ((dest[destPos] >> 1) & LANE) + base;
         if ((i & 3) == 3) {
            base = palette[(packed >> 8) & 0xff];
            packed += step + step;
         }
      }
   }

   // ===========================================================================================
   //  Sprite bank: load / capture / quantise
   // ===========================================================================================

   /**
    * Read an unsigned big-endian 16-bit value from a byte array. Mirrors the original's call to
    * {@code CacheFile.a(off, dummy, buf)} (obf: {@code d.a(int,byte,byte[])}); equivalent to
    * {@code Utility.getUnsignedShort}.
    */
   private static int getUShort(int off, byte dummy, byte[] buf) {
      return ((buf[off] & 0xff) << 8) + (buf[off + 1] & 0xff);
   }

   /**
    * Decode one (or several frames of a) cache sprite into a palette-indexed slot. (obf: {@code a(int,int,byte[],int,byte[])})
    *
    * <p>{@code indexData} carries the shared header (full width/height, palette) then per-frame
    * trims (translateX/Y, width/height, layout flag); {@code spriteData} carries the index stream.
    * {@code d.a(off,_,buf)} reads a big-endian unsigned 16-bit value.
    *
    * @param spriteId   first sprite slot to fill
    * @param frameCount number of consecutive frames to decode
    * @param spriteData pixel-index byte stream
    * @param dummy      unused/anti-tamper parameter (must be &ge; 49)
    * @param indexData  header + per-frame metadata byte stream
    */
   final void parseSprite(int spriteId, int frameCount, byte[] spriteData, int dummy, byte[] indexData) {
      int dataOff = Surface.getUShort(0, (byte) 41, spriteData); // offset of pixel data inside spriteData
      int fullWidth = Surface.getUShort(dataOff, (byte) 48, indexData);
      dataOff += 2;
      int fullHeight = Surface.getUShort(dataOff, (byte) 15, indexData);
      dataOff += 2;
      int paletteSize = indexData[dataOff++] & 0xff;
      int[] palette = new int[paletteSize];
      palette[0] = 0xff00ff; // index 0 is transparent magenta
      for (int i = 0; i < paletteSize - 1; i++) {
         palette[i + 1] = ((indexData[dataOff] & 0xff) << 16)
                        + ((indexData[dataOff + 1] & 0xff) << 8)
                        + (indexData[dataOff + 2] & 0xff);
         dataOff += 3;
      }

      int pixelOff = 2;
      for (int id = spriteId; id < spriteId + frameCount; id++) {
         this.spriteTranslateX[id] = indexData[dataOff++] & 0xff;
         this.spriteTranslateY[id] = indexData[dataOff++] & 0xff;
         this.spriteWidth[id] = Surface.getUShort(dataOff, (byte) 32, indexData);
         dataOff += 2;
         this.spriteHeight[id] = Surface.getUShort(dataOff, (byte) 83, indexData);
         dataOff += 2;
         int layout = indexData[dataOff++] & 0xff;
         int area = this.spriteWidth[id] * this.spriteHeight[id];
         this.spriteColourIndex[id] = new byte[area];
         this.spritePalette[id] = palette;
         this.spriteWidthFull[id] = fullWidth;
         this.spriteHeightFull[id] = fullHeight;
         this.spritePixels[id] = null;
         this.spriteTranslate[id] = false;
         if (this.spriteTranslateX[id] != 0 || this.spriteTranslateY[id] != 0) {
            this.spriteTranslate[id] = true;
         }
         if (layout == 0) {
            for (int i = 0; i < area; i++) {
               this.spriteColourIndex[id][i] = spriteData[pixelOff++];
               if (this.spriteColourIndex[id][i] == 0) {
                  this.spriteTranslate[id] = true;
               }
            }
         } else if (layout == 1) {
            for (int x = 0; x < this.spriteWidth[id]; x++) {
               for (int y = 0; y < this.spriteHeight[id]; y++) {
                  this.spriteColourIndex[id][x + y * this.spriteWidth[id]] = spriteData[pixelOff++];
                  if (this.spriteColourIndex[id][x + y * this.spriteWidth[id]] == 0) {
                     this.spriteTranslate[id] = true;
                  }
               }
            }
         }
      }
   }

   /**
    * Decode the run-length-encoded "sleep word" sprite into a 255×40 slot. (obf: {@code a(byte,byte[],int)})
    * Pixels are stored directly as 0xRRGGBB. Row 0 is RLE of black/white; subsequent rows XOR the
    * row above.
    *
    * @param dummy    unused/anti-tamper byte parameter (must equal -118)
    * @param data     RLE byte stream
    * @param spriteId destination slot
    */
   final void readSleepWord(byte dummy, byte[] data, int spriteId) {
      int[] out = this.spritePixels[spriteId] = new int[10200];
      this.spriteWidth[spriteId] = 255;
      this.spriteHeight[spriteId] = 40;
      this.spriteTranslateX[spriteId] = 0;
      this.spriteTranslateY[spriteId] = 0;
      this.spriteWidthFull[spriteId] = 255;
      this.spriteHeightFull[spriteId] = 40;
      this.spriteTranslate[spriteId] = false;
      int colour = 0;
      int src = 1;
      int dst;
      for (dst = 0; dst < 255; ) {
         int run = data[src++] & 0xff;
         for (int i = 0; i < run; i++) {
            out[dst++] = colour;
         }
         colour = 0xffffff - colour; // toggle black/white
      }
      for (int y = 1; y < 40; y++) {
         for (int x = 0; x < 255; ) {
            int run = data[src++] & 0xff;
            for (int i = 0; i < run; i++) {
               out[dst] = out[dst - 255]; // copy pixel above
               dst++;
               x++;
            }
            if (x < 255) {
               out[dst] = 0xffffff - out[dst - 255]; // invert pixel above
               dst++;
               x++;
            }
         }
      }
   }

   /**
    * Quantise a 24-bit sprite into a 256-colour palette in place. (obf: {@code a(boolean,int)})
    *
    * <p>Builds a 15-bit (5:5:5) histogram of the sprite's pixels, keeps the 255 most frequent
    * colours as a palette (slot 0 = transparent magenta), then maps every pixel to its nearest
    * palette entry by squared RGB distance. The slot ends up as a palette-indexed sprite.
    *
    * @param dummy    unused/anti-tamper parameter
    * @param spriteId sprite slot to quantise
    */
   final void drawWorld(boolean dummy, int spriteId) {
      int area = this.spriteHeight[spriteId] * this.spriteWidth[spriteId];
      int[] src = this.spritePixels[spriteId];
      int[] histogram = new int[32768];
      for (int i = 0; i < area; i++) {
         int c = src[i];
         histogram[((c >> 9) & 31744) + ((c & 0xf800) >> 6) + ((c >> 3) & 31)]++;
      }

      int[] palette = new int[256];
      palette[0] = 0xff00ff;
      int[] freq = new int[256]; // parallel frequency of each kept palette entry
      for (int bucket = 0; bucket < 32768; bucket++) {
         int count = histogram[bucket];
         if (count > freq[255]) {
            // insertion-sort this bucket into the top-256 by frequency
            for (int slot = 1; slot < 256; slot++) {
               if (freq[slot] < count) {
                  for (int k = 255; k > slot; k--) {
                     palette[k] = palette[k - 1];
                     freq[k] = freq[k - 1];
                  }
                  // expand 5:5:5 back to 0xRRGGBB and nudge by 0x040404 to avoid pure-black collisions
                  palette[slot] = ((bucket & 31) << 3) + ((bucket << 6) & 0xf800) + ((bucket << 9) & 0xf80000) + 0x040404;
                  freq[slot] = count;
                  break;
               }
            }
         }
         histogram[bucket] = -1; // reuse histogram as a bucket→paletteIndex cache
      }

      byte[] indices = new byte[area];
      for (int i = 0; i < area; i++) {
         int c = src[i];
         int bucket = ((c >> 3) & 31) + ((c & 0xf800) >> 6) + ((c & 0xf80000) >> 9);
         int paletteIdx = histogram[bucket];
         if (paletteIdx == -1) {
            // first time we see this bucket: find nearest palette colour by squared distance
            int best = 999999999;
            int r = (c >> 16) & 0xff, g = (c >> 8) & 0xff, b = c & 0xff;
            for (int p = 0; p < 256; p++) {
               int pc = palette[p];
               int pr = (pc >> 16) & 0xff, pg = (pc >> 8) & 0xff, pb = pc & 0xff;
               int dist = (b - pb) * (b - pb) + (r - pr) * (r - pr) + (g - pg) * (g - pg);
               if (dist < best) {
                  paletteIdx = p;
                  best = dist;
               }
            }
            histogram[bucket] = paletteIdx;
         }
         indices[i] = (byte) paletteIdx;
      }

      this.spriteColourIndex[spriteId] = indices;
      this.spritePalette[spriteId] = palette;
      this.spritePixels[spriteId] = null;
   }

   /**
    * Resolve a palette-indexed sprite into direct 0xRRGGBB pixels. (obf: {@code b(int,int)})
    * Colour 0 becomes 1 (so it isn't treated as transparent) and magenta (0xff00ff) becomes 0
    * (the transparent sentinel for the direct-pixel blitters).
    *
    * @param spriteId sprite slot
    * @param dummy    unused/anti-tamper parameter
    */
   final void loadSprite(int spriteId, int dummy) {
      if (this.spriteColourIndex[spriteId] == null) {
         return;
      }
      int area = this.spriteWidth[spriteId] * this.spriteHeight[spriteId];
      byte[] indices = this.spriteColourIndex[spriteId];
      int[] palette = this.spritePalette[spriteId];
      int[] out = new int[area];
      for (int i = 0; i < area; i++) {
         int colour = palette[indices[i] & 0xff];
         if (colour == 0) {
            colour = 1;
         } else if (colour == 0xff00ff) {
            colour = 0;
         }
         out[i] = colour;
      }
      this.spritePixels[spriteId] = out;
      this.spriteColourIndex[spriteId] = null;
      this.spritePalette[spriteId] = null;
   }

   /**
    * Capture a screen region into a sprite slot, column-major. (obf: {@code b(int,int,int,int,int,int)})
    * Used by {@code World} for the minimap. Pixels are read column by column.
    *
    * @param height   region height
    * @param x        region left
    * @param y        region top
    * @param dummy    unused/anti-tamper parameter (must equal -27966 at the tail)
    * @param spriteId destination slot
    * @param width    region width
    */
   final void drawSpriteMinimap(int height, int x, int y, int dummy, int spriteId, int width) {
      this.spriteWidth[spriteId] = width;
      this.spriteHeight[spriteId] = height;
      this.spriteTranslate[spriteId] = false;
      this.spriteTranslateX[spriteId] = 0;
      this.spriteTranslateY[spriteId] = 0;
      this.spriteWidthFull[spriteId] = width;
      this.spriteHeightFull[spriteId] = height;
      int i = 0;
      this.spritePixels[spriteId] = new int[width * height];
      for (int xx = x; xx < x + width; xx++) {
         for (int yy = y; yy < y + height; yy++) {
            this.spritePixels[spriteId][i++] = this.pixels[xx + this.width * yy];
         }
      }
   }

   /**
    * Capture a screen region into a sprite slot, row-major. (obf: {@code d(int,int,int,int,int,int)})
    * Used by the main client to snapshot UI areas.
    *
    * @param spriteId destination slot
    * @param height   region height
    * @param dummy    unused/anti-tamper parameter (must be &gt; 108)
    * @param width    region width
    * @param y        region top
    * @param x        region left
    */
   final void drawSprite(int spriteId, int height, int dummy, int width, int y, int x) {
      this.spriteWidth[spriteId] = width;
      this.spriteHeight[spriteId] = height;
      if (dummy > 108) {
         this.spriteTranslate[spriteId] = false;
         this.spriteTranslateX[spriteId] = 0;
         this.spriteTranslateY[spriteId] = 0;
         this.spriteWidthFull[spriteId] = width;
         this.spriteHeightFull[spriteId] = height;
         int i = 0;
         this.spritePixels[spriteId] = new int[width * height];
         for (int yy = y; yy < y + height; yy++) {
            for (int xx = x; xx < x + width; xx++) {
               this.spritePixels[spriteId][i++] = this.pixels[this.width * yy + xx];
            }
         }
      }
   }

   // ===========================================================================================
   //  Sprite blitting — plain
   // ===========================================================================================

   /**
    * Draw sprite {@code id} at ({@code x},{@code y}) with clipping. (obf: {@code b(int,int,int,int)})
    * Applies the sprite's trim offset, clips against the bounds, handles interlace, then dispatches
    * to the direct-pixel ({@link #blitSprite}) or palette-indexed ({@link #blitSpriteIndexed}) blitter.
    *
    * @param dummy unused/anti-tamper parameter
    * @param id    sprite slot
    * @param y     destination y
    * @param x     destination x
    */
   final void drawSprite(int dummy, int id, int y, int x) {
      if (this.spriteTranslate[id]) {
         x += this.spriteTranslateX[id];
         y += this.spriteTranslateY[id];
      }
      int dst = x + y * this.width;
      int src = 0;
      int h = this.spriteHeight[id];
      int w = this.spriteWidth[id];
      int rowSkip = this.width - w;
      int srcSkip = 0;
      if (y < this.boundsTopY) {
         int cut = this.boundsTopY - y;
         h -= cut;
         y = this.boundsTopY;
         src += cut * w;
         dst += cut * this.width;
      }
      if (y + h >= this.boundsBottomY) {
         h -= (y + h) - this.boundsBottomY + 1;
      }
      if (x < this.boundsTopX) {
         int cut = this.boundsTopX - x;
         w -= cut;
         x = this.boundsTopX;
         src += cut;
         dst += cut;
         srcSkip += cut;
         rowSkip += cut;
      }
      if (x + w >= this.boundsBottomX) {
         int cut = (x + w) - this.boundsBottomX + 1;
         w -= cut;
         srcSkip += cut;
         rowSkip += cut;
      }
      if (w <= 0 || h <= 0) {
         return;
      }
      byte vInc = 1;
      if (this.interlace) {
         vInc = 2;
         rowSkip += this.width;
         srcSkip += this.spriteWidth[id];
         if ((y & 1) != 0) {
            dst += this.width;
            h--;
         }
      }
      if (this.spritePixels[id] == null) {
         this.blitSpriteIndexed(src, this.spriteColourIndex[id], this.spritePalette[id], dst, w, h, rowSkip, srcSkip, (byte) 1);
      } else {
         this.blitSprite(0, this.spritePixels[id], src, dst, w, h, vInc, rowSkip, srcSkip);
      }
   }

   /**
    * Direct-pixel sprite blit, transparent where source pixel == 0. (obf: {@code a(int,int[],int,int,int,int,int,byte,int,int[],int,int)})
    * The inner loop is unrolled ×4 (the original micro-optimisation).
    *
    * @param tmp     scratch (reused as the current source pixel)
    * @param src     source 0xRRGGBB pixels
    * @param srcPos  read cursor
    * @param dstPos  write cursor into {@link #pixels}
    * @param w       width
    * @param h       height
    * @param vInc    vertical step (1, or 2 when interlaced)
    * @param rowSkip per-row destination advance
    * @param srcSkip per-row source advance
    */
   private final void blitSprite(int tmp, int[] src, int srcPos, int dstPos, int w, int h,
                                 byte vInc, int rowSkip, int srcSkip) {
      int quads = -(w >> 2);
      int rem = -(w & 3);
      for (int row = -h; row < 0; row += vInc) {
         for (int q = quads; q < 0; q++) {
            tmp = src[srcPos++]; if (tmp != 0) this.pixels[dstPos++] = tmp; else dstPos++;
            tmp = src[srcPos++]; if (tmp != 0) this.pixels[dstPos++] = tmp; else dstPos++;
            tmp = src[srcPos++]; if (tmp != 0) this.pixels[dstPos++] = tmp; else dstPos++;
            tmp = src[srcPos++]; if (tmp != 0) this.pixels[dstPos++] = tmp; else dstPos++;
         }
         for (int r = rem; r < 0; r++) {
            tmp = src[srcPos++]; if (tmp != 0) this.pixels[dstPos++] = tmp; else dstPos++;
         }
         dstPos += rowSkip;
         srcPos += srcSkip;
      }
   }

   /**
    * Palette-indexed sprite blit, transparent where index == 0. (obf: {@code a(int,int[],int,int,int,int,int,boolean,byte[],int,int)})
    * Inner loop unrolled ×4.
    *
    * @param srcPos  read cursor into {@code indices}
    * @param indices source palette indices
    * @param palette colour table (0xRRGGBB)
    * @param dstPos  write cursor into {@link #pixels}
    * @param w       width
    * @param h       height
    * @param dummy   unused/anti-tamper boolean
    * @param colourIdxBytes alias of {@code indices} kept by the original's odd argument shape
    * @param rowSkip per-row destination advance
    * @param srcSkip per-row source advance
    */
   private final void blitSpriteIndexed(int srcPos, int[] palette, byte[] indices, int dstPos, int w, int h,
                                        boolean dummy, byte[] colourIdxBytes, int rowSkip, int srcSkip) {
      // NOTE: signature mirrors the obfuscated argument order; see drawSprite for the real call.
      int quads = -(w >> 2);
      int rem = -(w & 3);
      for (int row = -h; row < 0; row++) {
         for (int q = quads; q < 0; q++) {
            byte b = colourIdxBytes[srcPos++]; if (b != 0) this.pixels[dstPos++] = palette[b & 0xff]; else dstPos++;
            b = colourIdxBytes[srcPos++]; if (b != 0) this.pixels[dstPos++] = palette[b & 0xff]; else dstPos++;
            b = colourIdxBytes[srcPos++]; if (b != 0) this.pixels[dstPos++] = palette[b & 0xff]; else dstPos++;
            b = colourIdxBytes[srcPos++]; if (b != 0) this.pixels[dstPos++] = palette[b & 0xff]; else dstPos++;
         }
         for (int r = rem; r < 0; r++) {
            byte b = colourIdxBytes[srcPos++]; if (b != 0) this.pixels[dstPos++] = palette[b & 0xff]; else dstPos++;
         }
         dstPos += rowSkip;
         srcPos += srcSkip;
      }
   }

   /**
    * Convenience overload matching the obfuscated palette-indexed blit argument shape used by
    * {@link #drawSprite}. (obf: companion of {@code a(I[III[IIIZ[BII)}) Delegates with the indices
    * array supplied for both byte-array slots.
    */
   private final void blitSpriteIndexed(int srcPos, byte[] indices, int[] palette, int dstPos, int w, int h,
                                        int rowSkip, int srcSkip, byte dummy) {
      this.blitSpriteIndexed(srcPos, palette, indices, dstPos, w, h, false, indices, rowSkip, srcSkip);
   }

   // ===========================================================================================
   //  Sprite blitting — alpha-blended
   // ===========================================================================================

   /**
    * Draw sprite {@code id} alpha-blended at ({@code x},{@code y}). (obf: {@code a(int,int,int,int,int)})
    *
    * @param id    sprite slot
    * @param dummy unused/anti-tamper parameter
    * @param x     destination x
    * @param alpha opacity 0..256
    * @param y     destination y
    */
   final void drawSpriteAlpha(int id, int dummy, int x, int alpha, int y) {
      if (this.spriteTranslate[id]) {
         x += this.spriteTranslateX[id];
         y += this.spriteTranslateY[id];
      }
      int dst = x + y * this.width;
      int src = 0;
      int h = this.spriteHeight[id];
      int w = this.spriteWidth[id];
      int rowSkip = this.width - w;
      int srcSkip = 0;
      if (y < this.boundsTopY) {
         int cut = this.boundsTopY - y;
         h -= cut;
         y = this.boundsTopY;
         src += cut * w;
         dst += cut * this.width;
      }
      if (y + h >= this.boundsBottomY) {
         h -= (y + h) - this.boundsBottomY + 1;
      }
      if (x < this.boundsTopX) {
         int cut = this.boundsTopX - x;
         w -= cut;
         x = this.boundsTopX;
         src += cut;
         dst += cut;
         srcSkip += cut;
         rowSkip += cut;
      }
      if (x + w >= this.boundsBottomX) {
         int cut = (x + w) - this.boundsBottomX + 1;
         w -= cut;
         srcSkip += cut;
         rowSkip += cut;
      }
      if (w <= 0 || h <= 0) {
         return;
      }
      byte vInc = 1;
      if (this.interlace) {
         vInc = 2;
         rowSkip += this.width;
         srcSkip += this.spriteWidth[id];
         if ((y & 1) != 0) {
            dst += this.width;
            h--;
         }
      }
      if (this.spritePixels[id] == null) {
         this.blitSpriteAlphaIndexed(src, vInc, this.spriteColourIndex[id], srcSkip, false, w, h, alpha, rowSkip,
               this.pixels, this.spritePalette[id], dst);
      } else {
         this.blitSpriteAlpha(vInc, w, dst, this.spritePixels[id], src, alpha, srcSkip, rowSkip, false, h, 0);
      }
   }

   /**
    * Direct-pixel alpha-blended sprite blit. (obf: {@code a(int,int,int,int[],int,int,int,int,boolean,int,int)})
    * Two-lane multiply: red+blue ({@code 0xff00ff}) and green ({@code 0xff00}) are scaled together.
    *
    * @param vInc    vertical step
    * @param w       width
    * @param dstPos  write cursor into {@link #pixels}
    * @param src     source 0xRRGGBB pixels
    * @param srcPos  read cursor
    * @param alpha   opacity 0..256
    * @param srcSkip per-row source advance
    * @param rowSkip per-row destination advance
    * @param dummy   unused/anti-tamper boolean
    * @param h       height
    * @param tmp     scratch (current source pixel)
    */
   private final void blitSpriteAlpha(int vInc, int w, int dstPos, int[] src, int srcPos, int alpha,
                                      int srcSkip, int rowSkip, boolean dummy, int h, int tmp) {
      int inv = 256 - alpha;
      for (int row = -h; row < 0; row += vInc) {
         for (int col = -w; col < 0; col++) {
            tmp = src[srcPos++];
            if (tmp != 0) {
               int bg = this.pixels[dstPos];
               this.pixels[dstPos++] = ((((tmp & 0xff00ff) * alpha + (bg & 0xff00ff) * inv) & 0xff00ff00)
                                      + (((tmp & 0xff00) * alpha + (bg & 0xff00) * inv) & 0xff0000)) >> 8;
            } else {
               dstPos++;
            }
         }
         dstPos += rowSkip;
         srcPos += srcSkip;
      }
   }

   /**
    * Palette-indexed alpha-blended sprite blit. (obf: {@code a(int,int,byte[],int,boolean,int,int,int,int,int[],int[],int)})
    *
    * @param srcPos  read cursor into {@code indices}
    * @param vInc    vertical step
    * @param indices source palette indices
    * @param srcSkip per-row source advance
    * @param dummy   unused/anti-tamper boolean
    * @param w       width
    * @param h       height
    * @param alpha   opacity 0..256
    * @param rowSkip per-row destination advance
    * @param dest    destination buffer ({@link #pixels})
    * @param palette colour table
    * @param dstPos  write cursor
    */
   private final void blitSpriteAlphaIndexed(int srcPos, int vInc, byte[] indices, int srcSkip, boolean dummy,
                                             int w, int h, int alpha, int rowSkip, int[] dest, int[] palette, int dstPos) {
      int inv = 256 - alpha;
      for (int row = -h; row < 0; row += vInc) {
         for (int col = -w; col < 0; col++) {
            int idx = indices[srcPos++] & 0xff;
            if (idx != 0) {
               int fg = palette[idx];
               int bg = dest[dstPos];
               dest[dstPos++] = ((((fg & 0xff00ff) * alpha + (bg & 0xff00ff) * inv) & 0xff00ff00)
                               + (((fg & 0xff00) * alpha + (bg & 0xff00) * inv) & 0xff0000)) >> 8;
            } else {
               dstPos++;
            }
         }
         dstPos += rowSkip;
         srcPos += srcSkip;
      }
   }

   // ===========================================================================================
   //  Sprite blitting — scaled (clip-to-rect)
   // ===========================================================================================

   /**
    * Draw sprite {@code id} scaled to ({@code w}×{@code h}) at ({@code x},{@code y}) with a grey-tint.
    * (obf: {@code a(int,int,int,int,int,byte,int)})
    *
    * <p>Source coordinates step in 16.16 fixed point; honours the sprite's trim offset and the clip
    * rectangle, then dispatches to {@link #plotScaleTinted}, which multiplies grey pixels (r==g==b)
    * by {@code colour}.
    *
    * @param x      destination x
    * @param colour grey-tint colour 0xRRGGBB
    * @param id     sprite slot
    * @param y      destination y
    * @param h      target height
    * @param dummy  unused/anti-tamper byte parameter
    * @param w      target width
    */
   final void spriteClippingTinted(int x, int colour, int id, int y, int h, byte dummy, int w) {
      try {
         int sw = this.spriteWidth[id];
         int sh = this.spriteHeight[id];
         int u0 = 0;
         int v0 = 0;
         int stepU = (sw << 16) / w;
         int stepV = (sh << 16) / h;
         if (this.spriteTranslate[id]) {
            int fw = this.spriteWidthFull[id];
            int fh = this.spriteHeightFull[id];
            stepU = (fw << 16) / w;
            stepV = (fh << 16) / h;
            if (fw == 0 || fh == 0) {
               return;
            }
            if ((this.spriteTranslateX[id] * w) % fw != 0) {
               u0 = ((fw - (this.spriteTranslateX[id] * w) % fw) << 16) / w;
            }
            x += (this.spriteTranslateX[id] * w + fw - 1) / fw;
            if ((this.spriteTranslateY[id] * h) % fh != 0) {
               v0 = ((fh - (this.spriteTranslateY[id] * h) % fh) << 16) / h;
            }
            y += (this.spriteTranslateY[id] * h + fh - 1) / fh;
            w = w * (this.spriteWidth[id] - (u0 >> 16)) / fw;
            h = h * (this.spriteHeight[id] - (v0 >> 16)) / fh;
         }
         int dst = x + y * this.width;
         int rowSkip = this.width - w;
         if (y < this.boundsTopY) {
            int cut = this.boundsTopY - y;
            h -= cut;
            y = 0;
            dst += cut * this.width;
            v0 += stepV * cut;
         }
         if (y + h >= this.boundsBottomY) {
            h -= (y + h) - this.boundsBottomY + 1;
         }
         if (x < this.boundsTopX) {
            int cut = this.boundsTopX - x;
            w -= cut;
            x = 0;
            dst += cut;
            u0 += stepU * cut;
            rowSkip += cut;
         }
         if (x + w >= this.boundsBottomX) {
            int cut = (x + w) - this.boundsBottomX + 1;
            w -= cut;
            rowSkip += cut;
         }
         if (this.interlace) {
            rowSkip += this.width;
            stepV += stepV;
            if ((y & 1) != 0) {
               dst += this.width;
               h--;
            }
         }
         this.plotScaleTinted(this.spritePixels[id], v0, stepU, 0, u0, dst, rowSkip, w, h, stepU, stepV, sw, colour);
      } catch (Exception ex) {
         System.out.println("error in sprite clipping routine");
      }
   }

   /**
    * Scaled, transparent direct-pixel blit. (obf: {@code a(int[],int,int,int,int,int[],byte,int,int,int,int,int,int,int)})
    * Reads {@code src[(u>>16) + (v>>16)*spriteWidth]}; skips pixels equal to 0.
    *
    * @param src     source 0xRRGGBB pixels
    * @param vStart  fixed-point v at the row start (saved/restored each row)
    * @param stepU   16.16 horizontal source step
    * @param tmp     scratch (current source pixel)
    * @param u       16.16 horizontal source coordinate
    * @param dest    destination ({@link #pixels})
    * @param dummy   unused/anti-tamper byte parameter
    * @param stepV   16.16 vertical source step
    * @param h       row count
    * @param uStart  unused source-row reset (kept from original)
    * @param rowSkip per-row destination advance
    * @param w       column count
    * @param spriteWidth source stride
    * @param dstPos  write cursor
    */
   private final void plotScale(int[] src, int vStart, int stepU, int tmp, int u, int[] dest, byte dummy,
                                int stepV, int h, int uStart, int rowSkip, int w, int spriteWidth, int dstPos) {
      try {
         int v = u; // the original aliases its row-reset local here; behaviour preserved
         int savedU = vStart;
         for (int row = -h; row < 0; row += 1) {
            int rowBase = (u >> 16) * spriteWidth;
            for (int col = -w; col < 0; col++) {
               tmp = src[(vStart >> 16) + rowBase];
               if (tmp != 0) {
                  dest[dstPos++] = tmp;
               } else {
                  dstPos++;
               }
               vStart += stepU;
            }
            u += stepV;
            vStart = savedU;
            dstPos += rowSkip;
         }
      } catch (Exception ex) {
         System.out.println("error in plot_scale");
      }
   }

   /**
    * Draw sprite {@code id} scaled to ({@code w}×{@code h}) at ({@code x},{@code y}), untinted.
    * (obf: {@code f(int,int,int,int,int,int)})
    *
    * <p>Same clip/fixed-point setup as {@link #spriteClippingTinted} but dispatches to the plain
    * scaler {@link #plotScale} (transparent where source pixel == 0).
    *
    * @param x      destination x
    * @param w      target width
    * @param y      destination y
    * @param h      target height
    * @param dummy  unused/anti-tamper parameter (the original passes a magic 5924 here)
    * @param id     sprite slot
    */
   final void spriteClipping(int x, int w, int y, int h, int dummy, int id) {
      try {
         int sw = this.spriteWidth[id];
         int sh = this.spriteHeight[id];
         int u0 = 0, v0 = 0;
         int stepU = (sw << 16) / w;
         int stepV = (sh << 16) / h;
         if (this.spriteTranslate[id]) {
            int fw = this.spriteWidthFull[id];
            int fh = this.spriteHeightFull[id];
            stepU = (fw << 16) / w;
            stepV = (fh << 16) / h;
            x += (this.spriteTranslateX[id] * w + fw - 1) / fw;
            y += (this.spriteTranslateY[id] * h + fh - 1) / fh;
            if ((this.spriteTranslateX[id] * w) % fw != 0) {
               u0 = ((fw - (this.spriteTranslateX[id] * w) % fw) << 16) / w;
            }
            if ((this.spriteTranslateY[id] * h) % fh != 0) {
               v0 = ((fh - (this.spriteTranslateY[id] * h) % fh) << 16) / h;
            }
            w = w * (this.spriteWidth[id] - (u0 >> 16)) / fw;
            h = h * (this.spriteHeight[id] - (v0 >> 16)) / fh;
         }
         int dst = x + y * this.width;
         int rowSkip = this.width - w;
         if (y < this.boundsTopY) {
            int cut = this.boundsTopY - y;
            h -= cut;
            y = 0;
            dst += cut * this.width;
            v0 += stepV * cut;
         }
         if (y + h >= this.boundsBottomY) {
            h -= (y + h) - this.boundsBottomY + 1;
         }
         if (x < this.boundsTopX) {
            int cut = this.boundsTopX - x;
            w -= cut;
            x = 0;
            dst += cut;
            u0 += stepU * cut;
            rowSkip += cut;
         }
         if (x + w >= this.boundsBottomX) {
            int cut = (x + w) - this.boundsBottomX + 1;
            w -= cut;
            rowSkip += cut;
         }
         if (this.interlace) {
            rowSkip += this.width;
            stepV += stepV;
            if ((y & 1) != 0) {
               dst += this.width;
               h--;
            }
         }
         this.plotScale(this.spritePixels[id], v0, stepU, 0, u0, this.pixels, (byte) 0, stepV, h, 0,
               rowSkip, w, sw, dst);
      } catch (Exception ex) {
         System.out.println("error in sprite clipping routine");
      }
   }

   /**
    * Scaled, transparent, grey-tinted direct-pixel blit. (obf: {@code a(int,int[],int,byte,int,int,int,int,byte[],int)})
    * For each source pixel: skip if 0; if r==g==b, multiply each channel by the tint; else copy.
    *
    * @param src    source 0xRRGGBB pixels
    * @param vStart fixed-point u at row start (saved/restored)
    * @param stepU  16.16 horizontal step
    * @param tmp    scratch
    * @param u      16.16 vertical accumulator
    * @param dstPos write cursor into {@link #pixels}
    * @param rowSkip per-row destination advance
    * @param w      columns
    * @param h      rows
    * @param stepUDup duplicate horizontal step (original argument shape)
    * @param stepV  16.16 vertical step
    * @param spriteWidth source stride
    * @param colour tint 0xRRGGBB
    */
   private final void plotScaleTinted(int[] src, int vStart, int stepU, int tmp, int u, int dstPos, int rowSkip,
                                      int w, int h, int stepUDup, int stepV, int spriteWidth, int colour) {
      int tintR = (colour >> 16) & 0xff;
      int tintG = (colour >> 8) & 0xff;
      int tintB = colour & 0xff;
      try {
         int savedU = vStart;
         for (int row = -h; row < 0; row++) {
            int rowBase = (u >> 16) * spriteWidth;
            for (int col = -w; col < 0; col++) {
               tmp = src[(vStart >> 16) + rowBase];
               if (tmp != 0) {
                  int r = (tmp >> 16) & 0xff, g = (tmp >> 8) & 0xff, b = tmp & 0xff;
                  if (r == g && g == b) {
                     this.pixels[dstPos++] = (((r * tintR >> 8) << 16) + ((g * tintG >> 8) << 8) + (b * tintB >> 8));
                  } else {
                     this.pixels[dstPos++] = tmp;
                  }
               } else {
                  dstPos++;
               }
               vStart += stepUDup;
            }
            u += stepV;
            vStart = savedU;
            dstPos += rowSkip;
         }
      } catch (Exception ex) {
         System.out.println("error in plot_scale");
      }
   }

   /**
    * Public scaled-sprite entry point; forwards to {@link #spriteClipping(int,int,int,int,int,int)}.
    * (obf: {@code a(int,int,int,int,int,int,byte,int)})
    *
    * <p>The original oracle exposed this as {@code spriteClipping(x,y,w,h,id,tx,ty)} where tx/ty are
    * unused; here they collapse to dummies.
    *
    * @param dummy0 unused/anti-tamper parameter
    * @param id     sprite slot
    * @param y      destination y
    * @param x      destination x
    * @param h      target height
    * @param w      target width
    * @param dummy1 unused/anti-tamper byte parameter (must equal 29 in the original)
    * @param dummy2 unused/anti-tamper parameter
    */
   void spriteClipping(int dummy0, int id, int y, int x, int h, int w, byte dummy1, int dummy2) {
      this.spriteClipping(x, w, y, h, 5924, id);
   }

   /**
    * Draw a scaled, alpha-blended "action bubble" sprite. (obf: {@code a(int,byte,int,int,int,int,int)})
    * Like {@link #spriteClipping} but dispatches to {@link #transparentScale}.
    *
    * @param x      destination x
    * @param dummy  unused/anti-tamper byte parameter
    * @param scaleX target width
    * @param scaleY target height
    * @param sprite sprite slot
    * @param alpha  opacity 0..256
    * @param y      destination y
    */
   final void drawActionBubble(int x, byte dummy, int scaleX, int scaleY, int sprite, int alpha, int y) {
      try {
         int sw = this.spriteWidth[sprite];
         int sh = this.spriteHeight[sprite];
         int u0 = 0, v0 = 0;
         int stepU = (sw << 16) / scaleX;
         int stepV = (sh << 16) / scaleY;
         if (this.spriteTranslate[sprite]) {
            int fw = this.spriteWidthFull[sprite];
            int fh = this.spriteHeightFull[sprite];
            stepU = (fw << 16) / scaleX;
            stepV = (fh << 16) / scaleY;
            x += (this.spriteTranslateX[sprite] * scaleX + fw - 1) / fw;
            y += (this.spriteTranslateY[sprite] * scaleY + fh - 1) / fh;
            if ((this.spriteTranslateX[sprite] * scaleX) % fw != 0) {
               u0 = ((fw - (this.spriteTranslateX[sprite] * scaleX) % fw) << 16) / scaleX;
            }
            if ((this.spriteTranslateY[sprite] * scaleY) % fh != 0) {
               v0 = ((fh - (this.spriteTranslateY[sprite] * scaleY) % fh) << 16) / scaleY;
            }
            scaleX = scaleX * (this.spriteWidth[sprite] - (u0 >> 16)) / fw;
            scaleY = scaleY * (this.spriteHeight[sprite] - (v0 >> 16)) / fh;
         }
         int dst = x + y * this.width;
         int rowSkip = this.width - scaleX;
         if (y < this.boundsTopY) {
            int cut = this.boundsTopY - y;
            scaleY -= cut;
            y = 0;
            dst += cut * this.width;
            v0 += stepV * cut;
         }
         if (y + scaleY >= this.boundsBottomY) {
            scaleY -= (y + scaleY) - this.boundsBottomY + 1;
         }
         if (x < this.boundsTopX) {
            int cut = this.boundsTopX - x;
            scaleX -= cut;
            x = 0;
            dst += cut;
            u0 += stepU * cut;
            rowSkip += cut;
         }
         if (x + scaleX >= this.boundsBottomX) {
            int cut = (x + scaleX) - this.boundsBottomX + 1;
            scaleX -= cut;
            rowSkip += cut;
         }
         byte vInc = 1;
         if (this.interlace) {
            vInc = 2;
            rowSkip += this.width;
            stepV += stepV;
            if ((y & 1) != 0) {
               dst += this.width;
               scaleY--;
            }
         }
         this.transparentScale(stepU, v0, scaleX, (byte) 0, stepV, u0, stepU, scaleY, sw,
               this.spritePixels[sprite], 0, dst, rowSkip, alpha, this.pixels);
      } catch (Exception ex) {
         System.out.println("error in sprite clipping routine");
      }
   }

   /**
    * Scaled, transparent, alpha-blended direct-pixel blit. (obf: {@code a(int,int,int,byte,int,int,int,int,int,int[],int,int,int,int,int[])})
    *
    * @param stepV2  per-row vertical step (added to {@code u})
    * @param vStart  fixed-point u at row start (saved/restored)
    * @param w       columns
    * @param dummy   unused/anti-tamper byte parameter (must equal a magic)
    * @param stepU   16.16 horizontal step
    * @param u       16.16 vertical accumulator
    * @param stepUDup duplicate horizontal step
    * @param h       rows
    * @param spriteWidth source stride
    * @param src     source 0xRRGGBB pixels
    * @param tmp     scratch
    * @param dstPos  write cursor
    * @param rowSkip per-row destination advance
    * @param alpha   opacity 0..256
    * @param dest    destination ({@link #pixels})
    */
   private final void transparentScale(int stepV2, int vStart, int w, byte dummy, int stepU, int u, int stepUDup,
                                       int h, int spriteWidth, int[] src, int tmp, int dstPos, int rowSkip,
                                       int alpha, int[] dest) {
      int inv = 256 - alpha;
      try {
         int savedU = vStart;
         for (int row = -h; row < 0; row++) {
            int rowBase = (u >> 16) * spriteWidth;
            for (int col = -w; col < 0; col++) {
               tmp = src[(vStart >> 16) + rowBase];
               if (tmp != 0) {
                  int bg = dest[dstPos];
                  dest[dstPos++] = ((((tmp & 0xff00ff) * alpha + (bg & 0xff00ff) * inv) & 0xff00ff00)
                                  + (((tmp & 0xff00) * alpha + (bg & 0xff00) * inv) & 0xff0000)) >> 8;
               } else {
                  dstPos++;
               }
               vStart += stepU;
            }
            u += stepV2;
            vStart = savedU;
            dstPos += rowSkip;
         }
      } catch (Exception ex) {
         System.out.println("error in tran_scale");
      }
   }

   // ===========================================================================================
   //  Sprite blitting — scaled with skew + two-colour tint (used for character equipment overlays)
   // ===========================================================================================

   /**
    * Draw sprite {@code sprite} scaled, optionally horizontally flipped, with a vertical skew and a
    * two-colour recolour. (obf: {@code a(int,int,int,boolean,int,int,int,int,int,int)})
    *
    * <p>{@code colour1} recolours grey pixels (r==g==b); {@code colour2} recolours pixels that are
    * pure-red-ish (r==255, g==b) — used to tint two parts of an item (e.g. cloth + trim). {@code skew}
    * shears each scanline horizontally. When {@code flip} is set the sprite is mirrored.
    *
    * @param x       destination x
    * @param y       destination y
    * @param w       target width
    * @param flip    mirror horizontally
    * @param h       target height
    * @param sprite  sprite slot
    * @param colour1 grey-tint colour (0 → white)
    * @param colour2 red-tint colour (0 → white); white means "single tint" fast path
    * @param skew    horizontal shear in 16.16 fixed point per row
    * @param dummy   unused/anti-tamper parameter
    */
   final void spriteClipping(int x, int y, int w, boolean flip, int h, int sprite, int colour1, int colour2,
                             int skew, int dummy) {
      try {
         if (colour1 == 0) {
            colour1 = 0xffffff;
         }
         if (colour2 == 0) {
            colour2 = 0xffffff;
         }
         int sw = this.spriteWidth[sprite];
         int sh = this.spriteHeight[sprite];
         int u0 = 0, v0 = 0;
         int skewAcc = skew << 16;
         int stepU = (sw << 16) / w;
         int stepV = (sh << 16) / h;
         int skewStep = -(skew << 16) / h;
         if (this.spriteTranslate[sprite]) {
            int fw = this.spriteWidthFull[sprite];
            int fh = this.spriteHeightFull[sprite];
            stepU = (fw << 16) / w;
            stepV = (fh << 16) / h;
            int tx = this.spriteTranslateX[sprite];
            int ty = this.spriteTranslateY[sprite];
            if (flip) {
               tx = fw - this.spriteWidth[sprite] - tx;
            }
            x += (tx * w + fw - 1) / fw;
            int tyOff = (ty * h + fh - 1) / fh;
            y += tyOff;
            skewAcc += tyOff * skewStep;
            if ((tx * w) % fw != 0) {
               u0 = ((fw - (tx * w) % fw) << 16) / w;
            }
            if ((ty * h) % fh != 0) {
               v0 = ((fh - (ty * h) % fh) << 16) / h;
            }
            w = (((this.spriteWidth[sprite] << 16) - u0) + stepU - 1) / stepU;
            h = (((this.spriteHeight[sprite] << 16) - v0) + stepV - 1) / stepV;
         }
         int dst = y * this.width;
         skewAcc += x << 16;
         if (y < this.boundsTopY) {
            int cut = this.boundsTopY - y;
            h -= cut;
            y = this.boundsTopY;
            dst += cut * this.width;
            v0 += stepV * cut;
            skewAcc += skewStep * cut;
         }
         if (y + h >= this.boundsBottomY) {
            h -= (y + h) - this.boundsBottomY + 1;
         }
         int parity = (dst / this.width) & 1;
         if (!this.interlace) {
            parity = 2; // disables the interlace skip
         }
         if (colour2 == 0xffffff) {
            // single-tint fast path
            if (this.spritePixels[sprite] != null) {
               if (!flip) {
                  this.transparentSpritePlot(this.spritePixels[sprite], u0, v0, dst, w, h, stepU, stepV, sw, colour1, skewAcc, skewStep, parity);
               } else {
                  this.transparentSpritePlot(this.spritePixels[sprite], (this.spriteWidth[sprite] << 16) - u0 - 1, v0, dst, w, h, -stepU, stepV, sw, colour1, skewAcc, skewStep, parity);
               }
            } else if (!flip) {
               this.transparentSpritePlot(this.spriteColourIndex[sprite], this.spritePalette[sprite], u0, v0, dst, w, h, stepU, stepV, sw, colour1, skewAcc, skewStep, parity);
            } else {
               this.transparentSpritePlot(this.spriteColourIndex[sprite], this.spritePalette[sprite], (this.spriteWidth[sprite] << 16) - u0 - 1, v0, dst, w, h, -stepU, stepV, sw, colour1, skewAcc, skewStep, parity);
            }
            return;
         }
         // two-tint path
         if (this.spritePixels[sprite] != null) {
            if (!flip) {
               this.transparentSpritePlot(this.spritePixels[sprite], u0, v0, dst, w, h, stepU, stepV, sw, colour1, colour2, skewAcc, skewStep, parity);
            } else {
               this.transparentSpritePlot(this.spritePixels[sprite], (this.spriteWidth[sprite] << 16) - u0 - 1, v0, dst, w, h, -stepU, stepV, sw, colour1, colour2, skewAcc, skewStep, parity);
            }
         } else if (!flip) {
            this.transparentSpritePlot(this.spriteColourIndex[sprite], this.spritePalette[sprite], u0, v0, dst, w, h, stepU, stepV, sw, colour1, colour2, skewAcc, skewStep, parity);
         } else {
            this.transparentSpritePlot(this.spriteColourIndex[sprite], this.spritePalette[sprite], (this.spriteWidth[sprite] << 16) - u0 - 1, v0, dst, w, h, -stepU, stepV, sw, colour1, colour2, skewAcc, skewStep, parity);
         }
      } catch (Exception ex) {
         System.out.println("error in sprite clipping routine");
      }
   }

   /**
    * Skewed, single-tint, transparent direct-pixel plot. (obf: {@code a(int[],int[],int,int,int,int,int,int,int,int,int,int,int,int,int,int,int)})
    * Grey pixels (r==g==b) are multiplied by {@code tint}; others copied. Per row the destination
    * span is re-clipped against the bounds because the skew moves it.
    *
    * @param src     source 0xRRGGBB pixels
    * @param u       16.16 horizontal source coord (row start, saved/restored)
    * @param v       16.16 vertical source coord
    * @param dstRow  destination row base into {@link #pixels}
    * @param w       columns
    * @param h       rows
    * @param stepU   16.16 horizontal step (negative when flipped)
    * @param stepV   16.16 vertical step
    * @param spriteWidth source stride
    * @param tint    0xRRGGBB grey-tint
    * @param skew    16.16 left edge x of the current scanline
    * @param skewStep per-row skew increment
    * @param parity  interlace parity toggle (2 disables)
    */
   private final void transparentSpritePlot(int[] src, int u, int v, int dstRow, int w, int h, int stepU, int stepV,
                                            int spriteWidth, int tint, int skew, int skewStep, int parity) {
      int tr = (tint >> 16) & 0xff, tg = (tint >> 8) & 0xff, tb = tint & 0xff;
      try {
         int savedU = u;
         for (int row = -h; row < 0; row++) {
            int rowBase = (v >> 16) * spriteWidth;
            int xs = skew >> 16;
            int span = w;
            if (xs < this.boundsTopX) {
               int cut = this.boundsTopX - xs;
               span -= cut;
               xs = this.boundsTopX;
               u += stepU * cut;
            }
            if (xs + span >= this.boundsBottomX) {
               span -= (xs + span) - this.boundsBottomX;
            }
            parity = 1 - parity;
            if (parity != 0) {
               for (int col = xs; col < xs + span; col++) {
                  int c = src[(u >> 16) + rowBase];
                  if (c != 0) {
                     int r = (c >> 16) & 0xff, g = (c >> 8) & 0xff, b = c & 0xff;
                     if (r == g && g == b) {
                        this.pixels[col + dstRow] = ((r * tr >> 8) << 16) + ((g * tg >> 8) << 8) + (b * tb >> 8);
                     } else {
                        this.pixels[col + dstRow] = c;
                     }
                  }
                  u += stepU;
               }
            }
            v += stepV;
            u = savedU;
            dstRow += this.width;
            skew += skewStep;
         }
      } catch (Exception ex) {
         System.out.println("error in transparent sprite plot routine");
      }
   }

   /**
    * Skewed, two-tint, transparent direct-pixel plot. (obf: {@code a(int[],int[],int,...,int)})
    * Adds a second recolour: pixels with {@code r==255 && g==b} are tinted by {@code tint2}.
    */
   private final void transparentSpritePlot(int[] src, int u, int v, int dstRow, int w, int h, int stepU, int stepV,
                                            int spriteWidth, int tint1, int tint2, int skew, int skewStep, int parity) {
      int t1r = (tint1 >> 16) & 0xff, t1g = (tint1 >> 8) & 0xff, t1b = tint1 & 0xff;
      int t2r = (tint2 >> 16) & 0xff, t2g = (tint2 >> 8) & 0xff, t2b = tint2 & 0xff;
      try {
         int savedU = u;
         for (int row = -h; row < 0; row++) {
            int rowBase = (v >> 16) * spriteWidth;
            int xs = skew >> 16;
            int span = w;
            if (xs < this.boundsTopX) {
               int cut = this.boundsTopX - xs;
               span -= cut;
               xs = this.boundsTopX;
               u += stepU * cut;
            }
            if (xs + span >= this.boundsBottomX) {
               span -= (xs + span) - this.boundsBottomX;
            }
            parity = 1 - parity;
            if (parity != 0) {
               for (int col = xs; col < xs + span; col++) {
                  int c = src[(u >> 16) + rowBase];
                  if (c != 0) {
                     int r = (c >> 16) & 0xff, g = (c >> 8) & 0xff, b = c & 0xff;
                     if (r == g && g == b) {
                        this.pixels[col + dstRow] = ((r * t1r >> 8) << 16) + ((g * t1g >> 8) << 8) + (b * t1b >> 8);
                     } else if (r == 255 && g == b) {
                        this.pixels[col + dstRow] = ((r * t2r >> 8) << 16) + ((g * t2g >> 8) << 8) + (b * t2b >> 8);
                     } else {
                        this.pixels[col + dstRow] = c;
                     }
                  }
                  u += stepU;
               }
            }
            v += stepV;
            u = savedU;
            dstRow += this.width;
            skew += skewStep;
         }
      } catch (Exception ex) {
         System.out.println("error in transparent sprite plot routine");
      }
   }

   /**
    * Palette-indexed single-tint skewed plot. (obf: {@code a(int,int[],int,...,byte[],...)})
    * Same as the direct-pixel single-tint variant but reads palette indices.
    */
   private final void transparentSpritePlot(byte[] indices, int[] palette, int u, int v, int dstRow, int w, int h,
                                            int stepU, int stepV, int spriteWidth, int tint, int skew, int skewStep, int parity) {
      int tr = (tint >> 16) & 0xff, tg = (tint >> 8) & 0xff, tb = tint & 0xff;
      try {
         int savedU = u;
         for (int row = -h; row < 0; row++) {
            int rowBase = (v >> 16) * spriteWidth;
            int xs = skew >> 16;
            int span = w;
            if (xs < this.boundsTopX) {
               int cut = this.boundsTopX - xs;
               span -= cut;
               xs = this.boundsTopX;
               u += stepU * cut;
            }
            if (xs + span >= this.boundsBottomX) {
               span -= (xs + span) - this.boundsBottomX;
            }
            parity = 1 - parity;
            if (parity != 0) {
               for (int col = xs; col < xs + span; col++) {
                  int idx = indices[(u >> 16) + rowBase] & 0xff;
                  if (idx != 0) {
                     int c = palette[idx];
                     int r = (c >> 16) & 0xff, g = (c >> 8) & 0xff, b = c & 0xff;
                     if (r == g && g == b) {
                        this.pixels[col + dstRow] = ((r * tr >> 8) << 16) + ((g * tg >> 8) << 8) + (b * tb >> 8);
                     } else {
                        this.pixels[col + dstRow] = c;
                     }
                  }
                  u += stepU;
               }
            }
            v += stepV;
            u = savedU;
            dstRow += this.width;
            skew += skewStep;
         }
      } catch (Exception ex) {
         System.out.println("error in transparent sprite plot routine");
      }
   }

   /**
    * Palette-indexed two-tint skewed plot. (obf: {@code a(int,int[],int,...,byte[],...,int)})
    * Adds the {@code r==255 && g==b} second-tint case to the palette-indexed variant.
    */
   private final void transparentSpritePlot(byte[] indices, int[] palette, int u, int v, int dstRow, int w, int h,
                                            int stepU, int stepV, int spriteWidth, int tint1, int tint2, int skew, int skewStep, int parity) {
      int t1r = (tint1 >> 16) & 0xff, t1g = (tint1 >> 8) & 0xff, t1b = tint1 & 0xff;
      int t2r = (tint2 >> 16) & 0xff, t2g = (tint2 >> 8) & 0xff, t2b = tint2 & 0xff;
      try {
         int savedU = u;
         for (int row = -h; row < 0; row++) {
            int rowBase = (v >> 16) * spriteWidth;
            int xs = skew >> 16;
            int span = w;
            if (xs < this.boundsTopX) {
               int cut = this.boundsTopX - xs;
               span -= cut;
               xs = this.boundsTopX;
               u += stepU * cut;
            }
            if (xs + span >= this.boundsBottomX) {
               span -= (xs + span) - this.boundsBottomX;
            }
            parity = 1 - parity;
            if (parity != 0) {
               for (int col = xs; col < xs + span; col++) {
                  int idx = indices[(u >> 16) + rowBase] & 0xff;
                  if (idx != 0) {
                     int c = palette[idx];
                     int r = (c >> 16) & 0xff, g = (c >> 8) & 0xff, b = c & 0xff;
                     if (r == g && g == b) {
                        this.pixels[col + dstRow] = ((r * t1r >> 8) << 16) + ((g * t1g >> 8) << 8) + (b * t1b >> 8);
                     } else if (r == 255 && g == b) {
                        this.pixels[col + dstRow] = ((r * t2r >> 8) << 16) + ((g * t2g >> 8) << 8) + (b * t2b >> 8);
                     } else {
                        this.pixels[col + dstRow] = c;
                     }
                  }
                  u += stepU;
               }
            }
            v += stepV;
            u = savedU;
            dstRow += this.width;
            skew += skewStep;
         }
      } catch (Exception ex) {
         System.out.println("error in transparent sprite plot routine");
      }
   }

   // ===========================================================================================
   //  Minimap: rotated + scaled sprite via a per-scanline texture-mapped quad
   // ===========================================================================================

   /**
    * Draw a sprite rotated by {@code rotation} (0..255) and scaled by {@code scale}, used for the
    * world minimap. (obf: {@code a(int,int,int,int,int,int)})
    *
    * <p>The sprite's four corners are rotated about ({@code x},{@code y}) using a sin/cos table
    * ({@link #minimapTrig}, 8-ish.15 fixed), then the quad is filled scanline by scanline. For each
    * row the left/right screen-x and the corresponding texture (u,v) are interpolated in 8.8 fixed
    * point ({@link #spanLeftX}.. arrays), then {@link #drawMinimap}/{@link #drawMinimapTranslate}
    * walk the span sampling the source texture in 17.15 fixed point.
    *
    * @param sprite   sprite slot
    * @param dummy    unused/anti-tamper parameter
    * @param y        centre y
    * @param x        centre x
    * @param scale    scale factor (e.g. 128 / 192)
    * @param rotation rotation 0..255
    */
   final void drawMinimapSprite(int sprite, int dummy, int y, int x, int scale, int rotation) {
      int w = this.width;
      int h = this.height;
      if (this.minimapTrig == null) {
         this.minimapTrig = new int[512];
         for (int i = 0; i < 256; i++) {
            this.minimapTrig[i] = (int) (Math.sin(i * 0.02454369D) * 32768D);       // sin
            this.minimapTrig[i + 256] = (int) (Math.cos(i * 0.02454369D) * 32768D); // cos
         }
      }
      int leftX = -this.spriteWidthFull[sprite] / 2;
      int topY = -this.spriteHeightFull[sprite] / 2;
      if (this.spriteTranslate[sprite]) {
         leftX += this.spriteTranslateX[sprite];
         topY += this.spriteTranslateY[sprite];
      }
      int rightX = leftX + this.spriteWidth[sprite];
      int bottomY = topY + this.spriteHeight[sprite];
      int cx = rightX, cy = topY, dx = leftX, dy = bottomY;
      rotation &= 0xff;
      int sin = this.minimapTrig[rotation] * scale;
      int cos = this.minimapTrig[rotation + 256] * scale;
      // rotate the four corners (>>22 == >>15 trig scale + >>7 the *scale)
      int aX = x + ((topY * sin + leftX * cos) >> 22);
      int aY = y + ((topY * cos - leftX * sin) >> 22);
      int bX = x + ((cy * sin + cx * cos) >> 22);
      int bY = y + ((cy * cos - cx * sin) >> 22);
      int cX = x + ((bottomY * sin + rightX * cos) >> 22);
      int cY = y + ((bottomY * cos - rightX * sin) >> 22);
      int dX = x + ((dy * sin + dx * cos) >> 22);
      int dY = y + ((dy * cos - dx * sin) >> 22);

      int minY = aY, maxY = aY;
      if (bY < minY) minY = bY; else if (bY > maxY) maxY = bY;
      if (cY < minY) minY = cY; else if (cY > maxY) maxY = cY;
      if (dY < minY) minY = dY; else if (dY > maxY) maxY = dY;
      if (minY < this.boundsTopY) minY = this.boundsTopY;
      if (maxY > this.boundsBottomY) maxY = this.boundsBottomY;

      if (this.spanLeftX == null || this.spanLeftX.length != h + 1) {
         this.spanLeftX = new int[h + 1];
         this.spanRightX = new int[h + 1];
         this.spanLeftU = new int[h + 1];
         this.spanRightU = new int[h + 1];
         this.spanLeftV = new int[h + 1];
         this.spanRightV = new int[h + 1];
      }
      for (int yy = minY; yy <= maxY; yy++) {
         this.spanLeftX[yy] = 99999999;
         this.spanRightX[yy] = 0xfa0a1f01;
      }

      // texture-space corners (sprite pixel coords)
      int texW = this.spriteWidth[sprite];
      int texH = this.spriteHeight[sprite];
      int taX = 0, taY = 0, tbX = texW - 1, tbY = 0, tcX = texW - 1, tcY = texH - 1, tdX = 0, tdY = texH - 1;

      // Edge A→D
      int dxScreen = 0, duTex = 0, dvTex = 0;
      if (dY != aY) {
         dxScreen = ((dX - aX) << 8) / (dY - aY);
         dvTex = ((tdY - taY) << 8) / (dY - aY);
      }
      int xFix, vFix, y0, y1;
      if (aY > dY) { xFix = dX << 8; vFix = tdY << 8; y0 = dY; y1 = aY; }
      else { xFix = aX << 8; vFix = taY << 8; y0 = aY; y1 = dY; }
      if (y0 < 0) { xFix -= dxScreen * y0; vFix -= dvTex * y0; y0 = 0; }
      if (y1 > h - 1) y1 = h - 1;
      for (int yy = y0; yy <= y1; yy++) {
         this.spanLeftX[yy] = this.spanRightX[yy] = xFix; xFix += dxScreen;
         this.spanLeftU[yy] = this.spanRightU[yy] = 0;
         this.spanLeftV[yy] = this.spanRightV[yy] = vFix; vFix += dvTex;
      }

      // Edge A→B
      int duTexB = 0;
      if (bY != aY) {
         dxScreen = ((bX - aX) << 8) / (bY - aY);
         duTexB = ((tbX - taX) << 8) / (bY - aY);
      }
      int uFix;
      if (aY > bY) { xFix = bX << 8; uFix = tbX << 8; y0 = bY; y1 = aY; }
      else { xFix = aX << 8; uFix = taX << 8; y0 = aY; y1 = bY; }
      if (y0 < 0) { xFix -= dxScreen * y0; uFix -= duTexB * y0; y0 = 0; }
      if (y1 > h - 1) y1 = h - 1;
      for (int yy = y0; yy <= y1; yy++) {
         if (xFix < this.spanLeftX[yy]) { this.spanLeftX[yy] = xFix; this.spanLeftU[yy] = uFix; this.spanLeftV[yy] = 0; }
         if (xFix > this.spanRightX[yy]) { this.spanRightX[yy] = xFix; this.spanRightU[yy] = uFix; this.spanRightV[yy] = 0; }
         xFix += dxScreen; uFix += duTexB;
      }

      // Edge B→C
      if (cY != bY) {
         dxScreen = ((cX - bX) << 8) / (cY - bY);
         dvTex = ((tcY - tbY) << 8) / (cY - bY);
      }
      if (bY > cY) { xFix = cX << 8; uFix = tcX << 8; vFix = tcY << 8; y0 = cY; y1 = bY; }
      else { xFix = bX << 8; uFix = tbX << 8; vFix = tbY << 8; y0 = bY; y1 = cY; }
      if (y0 < 0) { xFix -= dxScreen * y0; vFix -= dvTex * y0; y0 = 0; }
      if (y1 > h - 1) y1 = h - 1;
      for (int yy = y0; yy <= y1; yy++) {
         if (xFix < this.spanLeftX[yy]) { this.spanLeftX[yy] = xFix; this.spanLeftU[yy] = uFix; this.spanLeftV[yy] = vFix; }
         if (xFix > this.spanRightX[yy]) { this.spanRightX[yy] = xFix; this.spanRightU[yy] = uFix; this.spanRightV[yy] = vFix; }
         xFix += dxScreen; vFix += dvTex;
      }

      // Edge C→D
      if (dY != cY) {
         dxScreen = ((dX - cX) << 8) / (dY - cY);
         duTex = ((tdX - tcX) << 8) / (dY - cY);
      }
      if (cY > dY) { xFix = dX << 8; uFix = tdX << 8; vFix = tdY << 8; y0 = dY; y1 = cY; }
      else { xFix = cX << 8; uFix = tcX << 8; vFix = tcY << 8; y0 = cY; y1 = dY; }
      if (y0 < 0) { xFix -= dxScreen * y0; uFix -= duTex * y0; y0 = 0; }
      if (y1 > h - 1) y1 = h - 1;
      for (int yy = y0; yy <= y1; yy++) {
         if (xFix < this.spanLeftX[yy]) { this.spanLeftX[yy] = xFix; this.spanLeftU[yy] = uFix; this.spanLeftV[yy] = vFix; }
         if (xFix > this.spanRightX[yy]) { this.spanRightX[yy] = xFix; this.spanRightU[yy] = uFix; this.spanRightV[yy] = vFix; }
         xFix += dxScreen; uFix += duTex;
      }

      // Fill the spans
      int rowBase = minY * w;
      int[] tex = this.spritePixels[sprite];
      for (int yy = minY; yy < maxY; yy++) {
         int xL = this.spanLeftX[yy] >> 8;
         int xR = this.spanRightX[yy] >> 8;
         if (xR - xL <= 0) {
            rowBase += w;
         } else {
            int u = this.spanLeftU[yy] << 9;
            int stepU = ((this.spanRightU[yy] << 9) - u) / (xR - xL);
            int v = this.spanLeftV[yy] << 9;
            int stepV = ((this.spanRightV[yy] << 9) - v) / (xR - xL);
            if (xL < this.boundsTopX) {
               u += (this.boundsTopX - xL) * stepU;
               v += (this.boundsTopX - xL) * stepV;
               xL = this.boundsTopX;
            }
            if (xR > this.boundsBottomX) {
               xR = this.boundsBottomX;
            }
            if (!this.interlace || (yy & 1) == 0) {
               if (!this.spriteTranslate[sprite]) {
                  this.drawMinimap(this.pixels, tex, 0, rowBase + xL, u, v, stepU, stepV, xL - xR, texW);
               } else {
                  this.drawMinimapTranslate(this.pixels, tex, 0, rowBase + xL, u, v, stepU, stepV, xL - xR, texW);
               }
            }
            rowBase += w;
         }
      }
   }

   /**
    * Fill one minimap span (opaque). (obf: {@code a(int,int,int,int[],int,int,int,int[],int,int,boolean)})
    * Samples {@code tex[(u>>17) + (v>>17)*texW]} for each pixel ({@code >>17}: 8.8 span × 9-bit shift
    * up == 17.15 fixed point into the texture).
    *
    * @param destAlias unused alias of {@link #pixels} (kept from the original shape)
    * @param tex      source texture pixels
    * @param tmp      scratch
    * @param dstPos   write cursor
    * @param u        17.15 horizontal texture coord
    * @param v        17.15 vertical texture coord
    * @param stepU    horizontal texture step
    * @param stepV    vertical texture step
    * @param negCount {@code xL - xR} (negative span length used as a loop counter)
    * @param texW     texture stride
    */
   private final void drawMinimap(int[] destAlias, int[] tex, int tmp, int dstPos, int u, int v, int stepU, int stepV,
                                  int negCount, int texW) {
      for (tmp = negCount; tmp < 0; tmp++) {
         this.pixels[dstPos++] = tex[(u >> 17) + (v >> 17) * texW];
         u += stepU;
         v += stepV;
      }
   }

   /**
    * Fill one minimap span, skipping transparent (0) texels. (obf: {@code a(int,int,int,int[],int,int,int,int[],int,int,boolean)} → byte variant {@code III[I[IIIIIIB})
    * @see #drawMinimap
    */
   private final void drawMinimapTranslate(int[] destAlias, int[] tex, int tmp, int dstPos, int u, int v, int stepU,
                                           int stepV, int negCount, int texW) {
      for (int i = negCount; i < 0; i++) {
         int c = tex[(u >> 17) + (v >> 17) * texW];
         if (c != 0) {
            this.pixels[dstPos++] = c;
         } else {
            dstPos++;
         }
         u += stepU;
         v += stepV;
      }
   }

   // ===========================================================================================
   //  Bitmap-font text
   // ===========================================================================================

   /** Hard-coded line height per built-in font, falling back to {@link #textHeightFont}. (obf: {@code a(int,int)}) */
   final int textHeight(int dummy, int font) {
      if (font == 0) return 12;
      if (font == 1) return 14;
      if (font == 2) return 14;
      if (font == 3) return 15;
      if (font == 4) return 15;
      if (font == 5) return 19;
      if (font == 6) return 24;
      if (font == 7) return 29;
      return this.textHeightFont(60, font);
   }

   /** Line height derived from the font data byte 8 (for non-built-in fonts). (obf: {@code c(int,int)}) */
   private final int textHeightFont(int dummy, int font) {
      if (font == 0) {
         return Surface.gameFont(font)[8] - 2;
      }
      return Surface.gameFont(font)[8] - 1;
   }

   /**
    * Pixel width of {@code text} in {@code font}, ignoring {@code @col@} and {@code ~ddd~} markup.
    * (obf: {@code a(int,int,String)})
    *
    * @param font  font id
    * @param dummy unused/anti-tamper parameter (set to 67 in the original; nudges a dummy field)
    * @param text  text to measure
    */
   final int textWidth(int font, int dummy, String text) {
      int total = 0;
      byte[] fontData = Surface.gameFont(font);
      for (int i = 0; i < text.length(); i++) {
         if (text.charAt(i) == '@' && i + 4 < text.length() && text.charAt(i + 4) == '@') {
            i += 4;
         } else if (text.charAt(i) == '~' && i + 4 < text.length() && text.charAt(i + 4) == '~') {
            i += 4;
         } else {
            int ch = text.charAt(i);
            if (ch < 0 || ch >= Surface.characterWidth().length) {
               ch = 32; // out-of-range chars render as space
            }
            total += fontData[Surface.characterWidth()[ch] + 7];
         }
      }
      return total;
   }

   /**
    * Draw text left-aligned with {@code @col@} colour codes, {@code ~ddd~} absolute-x jumps and an
    * optional leading inline sprite glyph. (obf: {@code a(int,int,String,int,int,byte,int)})
    *
    * <p>When logged in and the colour is non-black, each glyph is drawn with a 1px black shadow
    * (offset +1,+1) for legibility. {@code @xxx@} switches the active colour to a named palette
    * entry; {@code ~ddd~} moves the pen to absolute x = ddd. If {@code inlineSprite >= 0} and a
    * sprite exists at {@code inlineSprite + inlineSpriteBase - 1}, it is drawn first.
    *
    * @param inlineSprite optional leading sprite id (or negative for none)
    * @param colour       starting text colour 0xRRGGBB
    * @param text         the string
    * @param x            pen x
    * @param y            pen baseline y
    * @param dummy        unused/anti-tamper byte parameter
    * @param font         font id
    */
   final void drawstring(int inlineSprite, int colour, String text, int x, int y, byte dummy, int font) {
      try {
         if (inlineSprite >= 0) {
            int spriteId = inlineSprite + this.inlineSpriteBase - 1;
            if (this.spriteColourIndex[spriteId] != null) {
               this.drawSprite(-1, spriteId, y - this.spriteHeight[spriteId], x);
               x += this.spriteWidth[spriteId]; // advance pen past the inline glyph
            }
         }
         byte[] fontData = Surface.gameFont(font);
         for (int i = 0; i < text.length(); i++) {
            if (text.charAt(i) == '@' && i + 4 < text.length() && text.charAt(i + 4) == '@') {
               String code = text.substring(i + 1, i + 4);
               if (code.equalsIgnoreCase("red")) colour = 0xff0000;
               else if (code.equalsIgnoreCase("lre")) colour = 0xff9040;
               else if (code.equalsIgnoreCase("yel")) colour = 0xffff00;
               else if (code.equalsIgnoreCase("gre")) colour = 0x00ff00;
               else if (code.equalsIgnoreCase("blu")) colour = 0x0000ff;
               else if (code.equalsIgnoreCase("cya")) colour = 0x00ffff;
               else if (code.equalsIgnoreCase("mag")) colour = 0xff00ff;
               else if (code.equalsIgnoreCase("whi")) colour = 0xffffff;
               else if (code.equalsIgnoreCase("bla")) colour = 0x000000;
               else if (code.equalsIgnoreCase("dre")) colour = 0xc00000;
               else if (code.equalsIgnoreCase("ora")) colour = 0xff9040;
               else if (code.equalsIgnoreCase("ran")) colour = (int) (Math.random() * 16777215D);
               else if (code.equalsIgnoreCase("or1")) colour = 0xffb000;
               else if (code.equalsIgnoreCase("or2")) colour = 0xff7000;
               else if (code.equalsIgnoreCase("or3")) colour = 0xff3000;
               else if (code.equalsIgnoreCase("gr1")) colour = 0xc0ff00;
               else if (code.equalsIgnoreCase("gr2")) colour = 0x80ff00;
               else if (code.equalsIgnoreCase("gr3")) colour = 0x40ff00;
               i += 4;
            } else if (text.charAt(i) == '~' && i + 4 < text.length() && text.charAt(i + 4) == '~') {
               char c0 = text.charAt(i + 1), c1 = text.charAt(i + 2), c2 = text.charAt(i + 3);
               if (c0 >= '0' && c0 <= '9' && c1 >= '0' && c1 <= '9' && c2 >= '0' && c2 <= '9') {
                  x = Integer.parseInt(text.substring(i + 1, i + 4));
               }
               i += 4;
            } else {
               int charDataOff = Surface.characterWidth()[text.charAt(i)];
               if (this.loggedIn() && colour != 0) {
                  this.drawCharacter((byte) 0, false, fontData, x + 1, y, 0, charDataOff);
                  this.drawCharacter((byte) 0, false, fontData, x, y + 1, 0, charDataOff);
               }
               this.drawCharacter((byte) 0, false, fontData, x, y, colour, charDataOff);
               x += fontData[charDataOff + 7];
            }
         }
      } catch (Exception ex) {
         System.out.println("drawstring: " + ex);
         ex.printStackTrace();
      }
   }

   /** Public left-aligned draw; forwards to {@link #drawstring}. (obf: {@code a(String,int,int,int,boolean,int)}) */
   final void drawstring(String text, int x, int y, int font, boolean dummy, int colour) {
      this.drawstring(0, y, text, x, colour, (byte) 124, font);
   }

   /** Right-aligned draw: pen at {@code x - textWidth}. (obf: {@code a(int,int,String,int,int,int,int)}) */
   private final void drawstringRight(int colour, int y, String text, int x, int dummy, int font, int inlineSprite) {
      this.drawstring(inlineSprite, y, text, x - this.textWidth(font, 114, text), colour, (byte) 123, font);
   }

   /** Right-aligned draw wrapper. (obf: {@code a(int,String,int,int,int,int)}) */
   final void drawstringRight(int font, String text, int colour, int x, int y, int inlineSprite) {
      this.drawstringRight(colour, y, text, x, 0, font, inlineSprite);
   }

   /**
    * Right-aligned draw wrapper (no inline sprite). (obf: {@code b(int,String,int,int,int,int)})
    * Distinct obfuscated method from {@link #drawstringRight(int,String,int,int,int,int)} despite the
    * identical erasure; renamed to avoid the clash.
    */
   final void drawstringRightSimple(int colour, String text, int y, int x, int dummy, int font) {
      this.drawstringRight(colour, y, text, x, -12200, font, 0);
   }

   /** Centre-aligned draw: pen at {@code x - textWidth/2}. (obf: {@code a(int,int,int,int,String,int,int)}) */
   private final void drawStringCenter(int magic, int colour, int font, int inlineSprite, String text, int x, int y) {
      if (magic == 11815) {
         this.drawstring(inlineSprite, y, text, x - this.textWidth(font, 92, text) / 2, colour, (byte) -124, font);
      }
   }

   /** Centre-aligned draw wrapper. (obf: {@code b(int,String,int,int,int,int)}) */
   final void drawStringCenter(int inlineSprite, String text, int font, int x, int colour, int y) {
      this.drawStringCenter(11815, colour, font, inlineSprite, text, x, y);
   }

   /**
    * Word-wrapped, centred paragraph. (obf: {@code a(int,String,int,int,int,int,boolean,int)})
    * Wraps at spaces (or a {@code %} forced break) so each line stays within {@code max} pixels,
    * advancing {@code y} by {@link #textHeight} per line.
    *
    * @param max    max line width in pixels
    * @param text   text
    * @param y      first-line baseline y
    * @param dummy0 unused/anti-tamper parameter
    * @param font   font id
    * @param x      centre x
    * @param dummy1 unused/anti-tamper boolean
    * @param colour text colour
    */
   final void centrepara(int max, String text, int y, int dummy0, int font, int x, boolean dummy1, int colour) {
      try {
         int width = 0;
         byte[] fontData = Surface.gameFont(font);
         int lineStart = 0;
         int breakAt = 0;
         for (int i = 0; i < text.length(); i++) {
            if (text.charAt(i) == '@' && i + 4 < text.length() && text.charAt(i + 4) == '@') {
               i += 4;
            } else if (text.charAt(i) == '~' && i + 4 < text.length() && text.charAt(i + 4) == '~') {
               i += 4;
            } else {
               width += fontData[Surface.characterWidth()[text.charAt(i)] + 7];
            }
            if (text.charAt(i) == ' ') {
               breakAt = i;
            }
            if (text.charAt(i) == '%') {
               breakAt = i;
               width = 1000; // force a break here
            }
            if (width > max) {
               if (breakAt <= lineStart) {
                  breakAt = i;
               }
               this.drawStringCenter(11815, colour, font, 0, text.substring(lineStart, breakAt), x, y);
               width = 0;
               lineStart = i = breakAt + 1;
               y += this.textHeight(508305352, font);
            }
         }
         if (width > 0) {
            this.drawStringCenter(11815, colour, font, 0, text.substring(lineStart), x, y);
         }
      } catch (Exception ex) {
         System.out.println("centrepara: " + ex);
         ex.printStackTrace();
      }
   }

   /**
    * Rasterise a single glyph. (obf: {@code a(byte,boolean,byte[],int,int,int,int)})
    * Reads the glyph's bitmap offset/size from {@code fontData} at {@code charDataOff} (bytes:
    * 0..2 = bitmap offset, 3 = width, 4 = height, 5 = left bearing, 6 = top bearing, 7 = advance),
    * clips it, then plots via {@link #plotLetter} (1-bit) or {@link #plotLetterAlpha} (anti-aliased).
    *
    * @param dummy       unused/anti-tamper byte parameter (must be &ge; 24)
    * @param antialias   when true, use the anti-aliased plot
    * @param fontData    font glyph table
    * @param x           pen x
    * @param y           pen baseline y
    * @param colour      glyph colour 0xRRGGBB
    * @param charDataOff offset of this glyph's metadata in {@code fontData}
    */
   private final void drawCharacter(byte dummy, boolean antialias, byte[] fontData, int x, int y, int colour, int charDataOff) {
      int gx = x + fontData[charDataOff + 5];
      int gy = y - fontData[charDataOff + 6];
      int gw = fontData[charDataOff + 3];
      int gh = fontData[charDataOff + 4];
      int bmpOff = fontData[charDataOff] * 16384 + fontData[charDataOff + 1] * 128 + fontData[charDataOff + 2];
      int dst = gx + gy * this.width;
      int rowSkip = this.width - gw;
      int srcSkip = 0;
      if (gy < this.boundsTopY) {
         int cut = this.boundsTopY - gy;
         gh -= cut;
         gy = this.boundsTopY;
         bmpOff += cut * gw;
         dst += cut * this.width;
      }
      if (gy + gh >= this.boundsBottomY) {
         gh -= (gy + gh) - this.boundsBottomY + 1;
      }
      if (gx < this.boundsTopX) {
         int cut = this.boundsTopX - gx;
         gw -= cut;
         gx = this.boundsTopX;
         bmpOff += cut;
         dst += cut;
         srcSkip += cut;
         rowSkip += cut;
      }
      if (gx + gw >= this.boundsBottomX) {
         int cut = (gx + gw) - this.boundsBottomX + 1;
         gw -= cut;
         srcSkip += cut;
         rowSkip += cut;
      }
      if (gw > 0 && gh > 0) {
         if (!antialias) {
            this.plotLetter(colour, this.pixels, dst, (byte) 0, gw, gh, srcSkip, rowSkip, fontData, bmpOff);
         } else {
            this.plotLetterAlpha(fontData, bmpOff, colour, dst, gw, gh, rowSkip, srcSkip, this.pixels, 0);
         }
      }
   }

   /**
    * 1-bit glyph plot (set colour where the font bit is non-zero). (obf: {@code a(int,int[],int,byte,int,int,int,int,byte[],int)})
    * Inner loop unrolled ×4.
    *
    * @param colour  glyph colour
    * @param dest    destination ({@link #pixels})
    * @param dstPos  write cursor
    * @param dummy   unused/anti-tamper byte parameter
    * @param w       glyph width
    * @param h       glyph height
    * @param srcSkip per-row source advance
    * @param rowSkip per-row destination advance
    * @param font    glyph bitmap bytes
    * @param srcPos  read cursor into {@code font}
    */
   private final void plotLetter(int colour, int[] dest, int dstPos, byte dummy, int w, int h,
                                 int srcSkip, int rowSkip, byte[] font, int srcPos) {
      try {
         int quads = -(w >> 2);
         int rem = -(w & 3);
         for (int row = -h; row < 0; row++) {
            for (int q = quads; q < 0; q++) {
               if (font[srcPos++] != 0) dest[dstPos++] = colour; else dstPos++;
               if (font[srcPos++] != 0) dest[dstPos++] = colour; else dstPos++;
               if (font[srcPos++] != 0) dest[dstPos++] = colour; else dstPos++;
               if (font[srcPos++] != 0) dest[dstPos++] = colour; else dstPos++;
            }
            for (int r = rem; r < 0; r++) {
               if (font[srcPos++] != 0) dest[dstPos++] = colour; else dstPos++;
            }
            dstPos += rowSkip;
            srcPos += srcSkip;
         }
      } catch (Exception ex) {
         System.out.println("plotletter: " + ex);
         ex.printStackTrace();
      }
   }

   /**
    * Anti-aliased glyph plot. (obf: {@code a(byte[],int,int,int,int,int,int,int,int[],int)})
    * Each font byte is a coverage value: &le;30 transparent, &ge;230 opaque, otherwise alpha-blend
    * the glyph colour over the background using the two-lane multiply trick.
    *
    * @param font    glyph coverage bytes
    * @param srcPos  read cursor
    * @param colour  glyph colour 0xRRGGBB
    * @param dstPos  write cursor into {@code dest}
    * @param w       glyph width
    * @param h       glyph height
    * @param rowSkip per-row destination advance
    * @param srcSkip per-row source advance
    * @param dest    destination ({@link #pixels})
    * @param tmp     scratch
    */
   private final void plotLetterAlpha(byte[] font, int srcPos, int colour, int dstPos, int w, int h,
                                      int rowSkip, int srcSkip, int[] dest, int tmp) {
      for (int row = -h; row < 0; row++) {
         for (int col = -w; col < 0; col++) {
            int coverage = font[srcPos++] & 0xff;
            if (coverage > 30) {
               if (coverage >= 230) {
                  dest[dstPos++] = colour;
               } else {
                  int bg = dest[dstPos];
                  dest[dstPos++] = ((((colour & 0xff00ff) * coverage + (bg & 0xff00ff) * (256 - coverage)) & 0xff00ff00)
                                  + (((colour & 0xff00) * coverage + (bg & 0xff00) * (256 - coverage)) & 0xff0000)) >> 8;
               }
            } else {
               dstPos++;
            }
         }
         dstPos += rowSkip;
         srcPos += srcSkip;
      }
   }

   // ===========================================================================================
   //  External-table accessors (mirror the obfuscated build's cross-class static references)
   // ===========================================================================================

   /** Shared font glyph data for {@code font} (obf: {@code m.b[font]}). */
   private static byte[] gameFont(int font) {
      return GameFonts.data[font];
   }

   /** Character-width lookup table (obf: {@code n.a}); value is the glyph-metadata offset. */
   private static int[] characterWidth() {
      return GameFonts.characterWidth;
   }

   /** Whether the player is logged in (drives text shadowing). Mirrors the original instance flag. */
   private boolean loggedIn() {
      return this.loggedIn;
   }

   // ===========================================================================================
   //  XOR string-pool decoders (obf: the two z(...) helpers)
   // ===========================================================================================
   //
   // The obfuscated class stored its literals in an XOR-encoded pool decoded at <clinit> by these
   // two helpers. All those literals (method-context labels, colour codes, console messages) have
   // been decoded and inlined throughout this file, so these are retained only for completeness.

   /** First stage: a no-op for normal pooled strings (only single-char strings get un-XOR'd). (obf: {@code z(String)}) */
   private static char[] decodeStage1(String s) {
      char[] cs = s.toCharArray();
      if (cs.length < 2) {
         cs[0] = (char) (cs[0] ^ '\t');
      }
      return cs;
   }

   /** Second stage: XOR each char with a 5-entry rotating key {63,48,37,6,9}. (obf: {@code z(char[])}) */
   private static String decodeStage2(char[] cs) {
      final byte[] key = {63, 48, 37, 6, 9};
      StringBuilder sb = new StringBuilder(cs.length);
      for (int i = 0; i < cs.length; i++) {
         sb.append((char) (cs[i] ^ key[i % 5]));
      }
      return sb.toString();
   }

   /** Logged-in flag controlling text shadow rendering. (obf: shared flag threaded via the client) */
   boolean loggedIn;

   /**
    * Stand-in for the cross-class shared font tables ({@code m.b} = gameFonts, {@code n.a} =
    * characterWidth) that this Surface reads but does not own. Declared here so the de-obfuscated
    * file is self-consistent; in the real client these live on {@code SocketFactory}/{@code FontWidths}.
    */
   static final class GameFonts {
      /** Per-font glyph bitmap+metadata tables. (obf: {@code m.b}) */
      static byte[][] data = new byte[50][];
      /** Glyph-metadata offset per character code (= 9 × glyph index). (obf: {@code n.a}) */
      static int[] characterWidth = buildCharacterWidth();

      private static int[] buildCharacterWidth() {
         String charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
                        + "!\"£$%^&*()-_=+[{]};:'@#~,<.>/?\\| ";
         int[] w = new int[256];
         for (int i = 0; i < 256; i++) {
            int idx = charset.indexOf(i);
            if (idx == -1) {
               idx = 74;
            }
            w[i] = idx * 9;
         }
         return w;
      }
   }
}







