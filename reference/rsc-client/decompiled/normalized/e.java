import java.applet.Applet;
import java.applet.AppletContext;
import java.awt.Color;
import java.awt.Dimension;
import java.awt.Font;
import java.awt.FontMetrics;
import java.awt.Graphics;
import java.awt.Image;
import java.awt.event.InputEvent;
import java.awt.event.KeyEvent;
import java.awt.event.KeyListener;
import java.awt.event.MouseEvent;
import java.awt.event.MouseListener;
import java.awt.event.MouseMotionListener;
import java.io.IOException;
import java.lang.reflect.Method;
import java.net.URL;

public class e extends Applet implements Runnable, MouseListener, MouseMotionListener, KeyListener {
   int K;
   static int y;
   private int n;
   long[] F;
   static int Nb;
   static byte[][] kb = new byte[250][];
   String p = null;
   static int G;
   static int A;
   static int L;
   static int Y;
   static int c;
   static int Db;
   static int W;
   static int r;
   static int Rb;
   static int qb;
   static int Fb;
   static int cb;
   static String[] Mb;
   static int Pb;
   static int jb;
   Thread z;
   static int Lb;
   private int Ib;
   static int pb;
   Font tb;
   static int Gb;
   private boolean hb;
   static int j;
   static int ab;
   static int ub;
   static int D;
   String B;
   static int rb;
   private int a;
   boolean N;
   static int Hb;
   static int eb;
   static int[] nb = new int[512];
   static int k;
   static int q;
   static int f;
   static int w;
   int Eb;
   static v i;
   static int t;
   static int h;
   int sb;
   static int P;
   static int M;
   private int m;
   private int S;
   static int g;
   static int mb;
   private int b;
   static int fb;
   static int R;
   private int vb;
   static int s;
   static int ob;
   static int d;
   private int V;
   static int O;
   static int yb;
   static int o;
   static int[] wb;
   static int J;
   static int l;
   static int db;
   Font X;
   Font Jb;
   Graphics u;
   Image C;
   int I;
   String e;
   private boolean Kb;
   int xb;
   String x;
   boolean U;
   int Bb;
   boolean gb;
   String Cb;
   boolean E;
   int Q;
   boolean bb;
   boolean Z;
   int Qb;
   String Ob;
   public static int Ab;
   public static boolean T;
   public static boolean H;
   public static int zb;
   public static int v;
   public static boolean ib;
   public static int lb;
   private static final String[] Sb = new String[]{
      z(z("-#;m\u000bwm,{H}{8a\u001a}gd(\u000ewq+a\u0006\u007f##a\u0004t")),
      z(z("}-,m\u001blq'q@1")),
      z(z("}-\u0004M@")),
      z(z("vv$d")),
      z(z("c-f&\u0015")),
      z(z("}-\u001b ")),
      z(z("}-#m\u0011Lz8m\f0")),
      z(z("}-%g\u001dkf\u000bd\u0001{h-l@")),
      z(z("}-/m\u001c_q)x\u0000q`; A")),
      z(z("}-%g\u001dkf\u0018z\rkp-l@")),
      z(z("}-8i\u0001vw`")),
      z(z("}-:}\u00060*")),
      z(z("{q){\u0000")),
      z(z("}-\u0001M@")),
      z(z(")1\u007f&X63f9")),
      z(z("pw<x")),
      z(z("Kw)z\u001c}ghi\u0018ho!k\tlj'f")),
      z(z("kf<N\u0007{v;\\\u001ayu-z\u001byo\u0003m\u0011kF&i\ntf,")),
      z(z("}-\u0006M@")),
      z(z("}-%g\u001dkf\u001am\u0004}b;m\f0")),
      z(z("}-%g\u001dkf\rf\u001c}q-l@")),
      z(z("}-/m\u001cHb:i\u0005}w-z@")),
      z(z("}-\u000eM@")),
      z(z("±#z8X).z8Y-#\u0002i\u000f}{hD\u001c|")),
      z(z("[q-i\u001c}ghj\u00118I\tO\r@#e(\u001eqp!|Hot?&\u0002yd-pF{l%")),
      z(z("}-;|\tjw`!")),
      z(z("}-#m\u0011Hq-{\u001b}g`")),
      z(z("}-\u001aM@")),
      z(z("[o'{\u0001vdhx\u001awd:i\u0005")),
      z(z("p2yx")),
      z(z("tl/gFld)")),
      z(z("p2zx")),
      z(z("Rb/m\u00108o!j\u001ayq1")),
      z(z("}-\u0002M@")),
      z(z("p2~j")),
      z(z("p1xj")),
      z(z("p2{j")),
      z(z("p1|j")),
      z(z("p2|j")),
      z(z("p2zj")),
      z(z("}-=x\fyw- ")),
      z(z("}-/m\u001cYs8d\rl@'f\u001c}{< A")),
      z(z("Pf$~\rlj+i")),
      z(z("Lj%m\u001bJl%i\u0006")),
      z(z("Tl)l\u0001vd")),
      z(z("}-%g\u001dkf\u0005g\u001e}g`")),
      z(z("}-\tL@")),
      z(z("}-!{,qp8d\tab*d\r0*")),
      z(z("}-\u0002L@")),
      z(z("}-\u0005I@")),
      z(z("}-\u0019 ")),
      z(z("}-/m\u001c\\l+}\u0005}m<J\tkf`!")),
      z(z("}-;|\u0007h+a")),
      z(z("}-\u0018M@")),
      z(z("}-8z\u0007nj,m$wb,m\u001aYs8d\rl+")),
      z(z("}-\u0005M@")),
      z(z("Tl)l\u0001vdf&F")),
      z(z("}-\rM@")),
      z(z("}-#m\u0011Jf$m\tkf, ")),
      z(z("tl/o\r|l=|")),
      z(z("Gw'x")),
      z(z("}-\u0003M@")),
      z(z("6t;")),
      z(z("}q:g\u001aGd)e\rG")),
      z(z("}-%g\u001dkf\rp\u0001lf, ")),
      z(z("Mm)j\u0004}#<gHtl)lH{l&|\rvwhx\t{hh")),
      z(z("}-\u0000M@")),
      z(z("}-/m\u001cKj2m@1")),
      z(z("}-\u000fM@")),
      z(z("Kw)z\u001c}ghi\u0018ho-|")),
      z(z("}-\u0007M@")),
      z(z("}-+z\ryw-A\u0005yd- ")),
      z(z("}-\u001bI@")),
      z(z("}-\u0003K@")),
      z(z("}-\u001bM@")),
      z(z("}-\u0019M@")),
      z(z("}-%g\u001dkf\fz\t\u007fd-l@"))
   };

   public final String getParameter(String var1) {
      try {
         j++;
         if (null != kb.a) {
            return null;
         } else {
            try {
               if (da.gb != null) {
                  return da.gb.getParameter(var1);
               }
            } catch (RuntimeException var4) {
               throw var4;
            }

            return super.getParameter(var1);
         }
      } catch (RuntimeException var5) {
         RuntimeException var2 = var5;

         RuntimeException var10000;
         StringBuilder var10001;
         try {
            var10000 = var2;
            var10001 = new StringBuilder().append(Sb[21]);
            if (var1 != null) {
               throw i.a(var2, var10001.append(Sb[4]).append(')').toString());
            }
         } catch (RuntimeException var3) {
            throw var3;
         }

         throw i.a(var10000, var10001.append(Sb[3]).append(')').toString());
      }
   }

   private final void b(int var1) {
      try {
         try {
            fb++;
            this.vb = -2;
            System.out.println(Sb[28]);
            this.a(false);
            mb.a(11200, 1000L);
            if (var1 != 100) {
               this.e(27);
            }
         } catch (RuntimeException var3) {
            throw var3;
         }

         try {
            if (kb.a == null) {
               return;
            }
         } catch (RuntimeException var4) {
            throw var4;
         }

         kb.a.dispose();
         System.exit(0);
      } catch (RuntimeException var5) {
         throw i.a(var5, Sb[27] + var1 + ')');
      }
   }

   @Override
   public final void run() {
      // $VF: Couldn't be decompiled
      // Please report this to the Vineflower issue tracker, at https://github.com/Vineflower/vineflower/issues with a copy of the class file (if you have the rights to distribute it!)
      // java.lang.RuntimeException: parsing failure!
      //   at org.jetbrains.java.decompiler.modules.decompiler.decompose.DomHelper.parseGraph(DomHelper.java:211)
      //   at org.jetbrains.java.decompiler.main.rels.MethodProcessor.codeToJava(MethodProcessor.java:166)
      //
      // Bytecode:
      // 000: getstatic client.vh Z
      // 003: istore 10
      // 005: getstatic e.qb I
      // 008: bipush 1
      // 009: iadd
      // 00a: putstatic e.qb I
      // 00d: bipush 1
      // 00e: aload 0
      // 00f: getfield e.n I
      // 012: if_icmpne 0c9
      // 015: aload 0
      // 016: bipush 2
      // 017: putfield e.n I
      // 01a: aload 0
      // 01b: invokevirtual e.isDisplayable ()Z
      // 01e: ifne 073
      // 021: aload 0
      // 022: getfield e.vb I
      // 025: iload 10
      // 027: ifne 074
      // 02a: iflt 073
      // 02d: goto 031
      // 030: athrow
      // 031: bipush -1
      // 032: aload 0
      // 033: getfield e.vb I
      // 036: bipush -1
      // 037: ixor
      // 038: if_icmpgt 043
      // 03b: goto 03f
      // 03e: athrow
      // 03f: goto 063
      // 042: athrow
      // 043: aload 0
      // 044: dup
      // 045: getfield e.vb I
      // 048: bipush 1
      // 049: isub
      // 04a: putfield e.vb I
      // 04d: bipush -1
      // 04e: aload 0
      // 04f: getfield e.vb I
      // 052: bipush -1
      // 053: ixor
      // 054: if_icmpne 063
      // 057: aload 0
      // 058: bipush 100
      // 05a: invokespecial e.b (I)V
      // 05d: aload 0
      // 05e: aconst_null
      // 05f: putfield e.z Ljava/lang/Thread;
      // 062: return
      // 063: sipush 11200
      // 066: aload 0
      // 067: getfield e.Ib I
      // 06a: i2l
      // 06b: invokestatic mb.a (IJ)V
      // 06e: iload 10
      // 070: ifeq 01a
      // 073: bipush 0
      // 074: aload 0
      // 075: getfield e.vb I
      // 078: if_icmple 097
      // 07b: aload 0
      // 07c: getfield e.vb I
      // 07f: bipush -1
      // 080: if_icmpne 091
      // 083: goto 087
      // 086: athrow
      // 087: aload 0
      // 088: bipush 100
      // 08a: invokespecial e.b (I)V
      // 08d: goto 091
      // 090: athrow
      // 091: aload 0
      // 092: aconst_null
      // 093: putfield e.z Ljava/lang/Thread;
      // 096: return
      // 097: aload 0
      // 098: bipush 118
      // 09a: invokespecial e.b (B)Z
      // 09d: ifeq 0a4
      // 0a0: goto 0be
      // 0a3: athrow
      // 0a4: bipush 1
      // 0a5: aload 0
      // 0a6: getfield e.vb I
      // 0a9: bipush -1
      // 0aa: ixor
      // 0ab: if_icmpeq 0b8
      // 0ae: aload 0
      // 0af: bipush 100
      // 0b1: invokespecial e.b (I)V
      // 0b4: goto 0b8
      // 0b7: athrow
      // 0b8: aload 0
      // 0b9: aconst_null
      // 0ba: putfield e.z Ljava/lang/Thread;
      // 0bd: return
      // 0be: aload 0
      // 0bf: bipush -92
      // 0c1: invokevirtual e.a (B)V
      // 0c4: aload 0
      // 0c5: bipush 0
      // 0c6: putfield e.n I
      // 0c9: aconst_null
      // 0ca: getstatic kb.a Lqb;
      // 0cd: if_acmpeq 0ee
      // 0d0: getstatic kb.a Lqb;
      // 0d3: aload 0
      // 0d4: invokevirtual qb.addMouseListener (Ljava/awt/event/MouseListener;)V
      // 0d7: getstatic kb.a Lqb;
      // 0da: aload 0
      // 0db: invokevirtual qb.addMouseMotionListener (Ljava/awt/event/MouseMotionListener;)V
      // 0de: getstatic kb.a Lqb;
      // 0e1: aload 0
      // 0e2: invokevirtual qb.addKeyListener (Ljava/awt/event/KeyListener;)V
      // 0e5: iload 10
      // 0e7: ifeq 12a
      // 0ea: goto 0ee
      // 0ed: athrow
      // 0ee: aconst_null
      // 0ef: getstatic da.gb Ljava/applet/Applet;
      // 0f2: if_acmpne 111
      // 0f5: goto 0f9
      // 0f8: athrow
      // 0f9: aload 0
      // 0fa: aload 0
      // 0fb: invokevirtual e.addMouseListener (Ljava/awt/event/MouseListener;)V
      // 0fe: aload 0
      // 0ff: aload 0
      // 100: invokevirtual e.addMouseMotionListener (Ljava/awt/event/MouseMotionListener;)V
      // 103: aload 0
      // 104: aload 0
      // 105: invokevirtual e.addKeyListener (Ljava/awt/event/KeyListener;)V
      // 108: iload 10
      // 10a: ifeq 12a
      // 10d: goto 111
      // 110: athrow
      // 111: getstatic da.gb Ljava/applet/Applet;
      // 114: aload 0
      // 115: invokevirtual java/applet/Applet.addMouseListener (Ljava/awt/event/MouseListener;)V
      // 118: getstatic da.gb Ljava/applet/Applet;
      // 11b: aload 0
      // 11c: invokevirtual java/applet/Applet.addMouseMotionListener (Ljava/awt/event/MouseMotionListener;)V
      // 11f: getstatic da.gb Ljava/applet/Applet;
      // 122: aload 0
      // 123: invokevirtual java/applet/Applet.addKeyListener (Ljava/awt/event/KeyListener;)V
      // 126: goto 12a
      // 129: athrow
      // 12a: bipush 0
      // 12b: istore 3
      // 12c: sipush 256
      // 12f: istore 4
      // 131: bipush 1
      // 132: istore 5
      // 134: bipush 0
      // 135: istore 6
      // 137: bipush 0
      // 138: istore 7
      // 13a: iload 7
      // 13c: bipush 10
      // 13e: if_icmpge 15d
      // 141: aload 0
      // 142: getfield e.F [J
      // 145: iload 7
      // 147: bipush 0
      // 148: invokestatic p.a (I)J
      // 14b: lastore
      // 14c: iinc 7 1
      // 14f: iload 10
      // 151: ifne 351
      // 154: iload 10
      // 156: ifeq 13a
      // 159: goto 15d
      // 15c: athrow
      // 15d: bipush 0
      // 15e: invokestatic p.a (I)J
      // 161: lstore 1
      // 162: bipush 0
      // 163: aload 0
      // 164: getfield e.vb I
      // 167: if_icmpgt 324
      // 16a: bipush -1
      // 16b: aload 0
      // 16c: getfield e.vb I
      // 16f: bipush -1
      // 170: ixor
      // 171: iload 10
      // 173: ifne 329
      // 176: if_icmple 1a3
      // 179: goto 17d
      // 17c: athrow
      // 17d: aload 0
      // 17e: dup
      // 17f: getfield e.vb I
      // 182: bipush 1
      // 183: isub
      // 184: putfield e.vb I
      // 187: bipush 0
      // 188: aload 0
      // 189: getfield e.vb I
      // 18c: if_icmpeq 197
      // 18f: goto 193
      // 192: athrow
      // 193: goto 1a3
      // 196: athrow
      // 197: aload 0
      // 198: bipush 100
      // 19a: invokespecial e.b (I)V
      // 19d: aload 0
      // 19e: aconst_null
      // 19f: putfield e.z Ljava/lang/Thread;
      // 1a2: return
      // 1a3: iload 4
      // 1a5: istore 7
      // 1a7: sipush 300
      // 1aa: istore 4
      // 1ac: iload 5
      // 1ae: istore 8
      // 1b0: bipush 1
      // 1b1: istore 5
      // 1b3: bipush 0
      // 1b4: invokestatic p.a (I)J
      // 1b7: lstore 1
      // 1b8: aload 0
      // 1b9: getfield e.F [J
      // 1bc: iload 3
      // 1bd: laload
      // 1be: ldc2_w -1
      // 1c1: lxor
      // 1c2: ldc2_w -1
      // 1c5: lcmp
      // 1c6: ifne 1d6
      // 1c9: iload 8
      // 1cb: istore 5
      // 1cd: iload 7
      // 1cf: istore 4
      // 1d1: iload 10
      // 1d3: ifeq 203
      // 1d6: lload 1
      // 1d7: ldc2_w -1
      // 1da: lxor
      // 1db: aload 0
      // 1dc: getfield e.F [J
      // 1df: iload 3
      // 1e0: laload
      // 1e1: ldc2_w -1
      // 1e4: lxor
      // 1e5: lcmp
      // 1e6: ifge 203
      // 1e9: goto 1ed
      // 1ec: athrow
      // 1ed: aload 0
      // 1ee: getfield e.Ib I
      // 1f1: sipush 2560
      // 1f4: imul
      // 1f5: i2l
      // 1f6: lload 1
      // 1f7: aload 0
      // 1f8: getfield e.F [J
      // 1fb: iload 3
      // 1fc: laload
      // 1fd: lneg
      // 1fe: ladd
      // 1ff: ldiv
      // 200: l2i
      // 201: istore 4
      // 203: iload 4
      // 205: bipush -1
      // 206: ixor
      // 207: bipush -26
      // 209: if_icmpgt 210
      // 20c: goto 214
      // 20f: athrow
      // 210: bipush 25
      // 212: istore 4
      // 214: sipush -257
      // 217: iload 4
      // 219: bipush -1
      // 21a: ixor
      // 21b: if_icmpgt 222
      // 21e: goto 251
      // 221: athrow
      // 222: sipush 256
      // 225: istore 4
      // 227: aload 0
      // 228: getfield e.F [J
      // 22b: iload 3
      // 22c: laload
      // 22d: lneg
      // 22e: lload 1
      // 22f: ladd
      // 230: ldc2_w 10
      // 233: ldiv
      // 234: lneg
      // 235: aload 0
      // 236: getfield e.Ib I
      // 239: i2l
      // 23a: ladd
      // 23b: l2i
      // 23c: istore 5
      // 23e: iload 5
      // 240: aload 0
      // 241: getfield e.Q I
      // 244: if_icmplt 24b
      // 247: goto 251
      // 24a: athrow
      // 24b: aload 0
      // 24c: getfield e.Q I
      // 24f: istore 5
      // 251: sipush 11200
      // 254: iload 5
      // 256: i2l
      // 257: invokestatic mb.a (IJ)V
      // 25a: aload 0
      // 25b: getfield e.F [J
      // 25e: iload 3
      // 25f: lload 1
      // 260: lastore
      // 261: bipush -2
      // 263: iload 5
      // 265: bipush -1
      // 266: ixor
      // 267: if_icmple 2a7
      // 26a: bipush 0
      // 26b: istore 9
      // 26d: bipush 10
      // 26f: iload 9
      // 271: if_icmple 2a7
      // 274: ldc2_w -1
      // 277: aload 0
      // 278: getfield e.F [J
      // 27b: iload 9
      // 27d: laload
      // 27e: ldc2_w -1
      // 281: lxor
      // 282: lcmp
      // 283: iload 10
      // 285: ifne 2af
      // 288: ifne 292
      // 28b: goto 28f
      // 28e: athrow
      // 28f: goto 29f
      // 292: aload 0
      // 293: getfield e.F [J
      // 296: iload 9
      // 298: dup2
      // 299: laload
      // 29a: iload 5
      // 29c: i2l
      // 29d: ladd
      // 29e: lastore
      // 29f: iinc 9 1
      // 2a2: iload 10
      // 2a4: ifeq 26d
      // 2a7: bipush 1
      // 2a8: iload 3
      // 2a9: iadd
      // 2aa: bipush 10
      // 2ac: irem
      // 2ad: istore 3
      // 2ae: bipush 0
      // 2af: istore 9
      // 2b1: sipush -257
      // 2b4: iload 6
      // 2b6: bipush -1
      // 2b7: ixor
      // 2b8: if_icmpge 308
      // 2bb: aload 0
      // 2bc: bipush 119
      // 2be: invokevirtual e.e (I)V
      // 2c1: iload 6
      // 2c3: iload 4
      // 2c5: iadd
      // 2c6: istore 6
      // 2c8: iinc 9 1
      // 2cb: aload 0
      // 2cc: getfield e.S I
      // 2cf: bipush -1
      // 2d0: ixor
      // 2d1: iload 9
      // 2d3: bipush -1
      // 2d4: ixor
      // 2d5: if_icmple 2b1
      // 2d8: bipush 0
      // 2d9: istore 6
      // 2db: aload 0
      // 2dc: dup
      // 2dd: getfield e.b I
      // 2e0: bipush 6
      // 2e2: iadd
      // 2e3: putfield e.b I
      // 2e6: bipush 25
      // 2e8: aload 0
      // 2e9: getfield e.b I
      // 2ec: iload 10
      // 2ee: ifne 317
      // 2f1: if_icmplt 2fb
      // 2f4: goto 2f8
      // 2f7: athrow
      // 2f8: goto 308
      // 2fb: aload 0
      // 2fc: bipush 0
      // 2fd: putfield e.b I
      // 300: aload 0
      // 301: bipush 1
      // 302: putfield e.U Z
      // 305: goto 308
      // 308: aload 0
      // 309: dup
      // 30a: getfield e.b I
      // 30d: bipush 1
      // 30e: isub
      // 30f: putfield e.b I
      // 312: iload 6
      // 314: sipush 255
      // 317: iand
      // 318: istore 6
      // 31a: aload 0
      // 31b: bipush 0
      // 31c: invokevirtual e.b (Z)V
      // 31f: iload 10
      // 321: ifeq 162
      // 324: bipush -1
      // 325: aload 0
      // 326: getfield e.vb I
      // 329: if_icmpne 336
      // 32c: aload 0
      // 32d: bipush 100
      // 32f: invokespecial e.b (I)V
      // 332: goto 336
      // 335: athrow
      // 336: aload 0
      // 337: aconst_null
      // 338: putfield e.z Ljava/lang/Thread;
      // 33b: goto 351
      // 33e: astore 1
      // 33f: ldc 2097151
      // 341: aload 1
      // 342: aconst_null
      // 343: invokestatic mb.a (ILjava/lang/Throwable;Ljava/lang/String;)V
      // 346: aload 0
      // 347: getstatic e.Sb [Ljava/lang/String;
      // 34a: bipush 12
      // 34c: aaload
      // 34d: bipush 1
      // 34e: invokespecial e.a (Ljava/lang/String;Z)V
      // 351: goto 360
      // 354: astore 1
      // 355: aload 1
      // 356: getstatic e.Sb [Ljava/lang/String;
      // 359: bipush 11
      // 35b: aaload
      // 35c: invokestatic i.a (Ljava/lang/Throwable;Ljava/lang/String;)Lla;
      // 35f: athrow
      // 360: return
   }

   @Override
   public final synchronized void keyReleased(KeyEvent var1) {
      try {
         this.a(var1, (byte)-128);
         w++;
         char var10 = var1.getKeyChar();
         int var3 = var1.getKeyCode();
         if (~((char)var10) != -33) {
         }

         if (40 == var3) {
         }

         try {
            if ('N' != (char)var10 && 'M' != (char)var10) {
            }
         } catch (RuntimeException var8) {
            throw var8;
         }

         try {
            if (~var3 == -40) {
               this.E = false;
            }
         } catch (RuntimeException var4) {
            throw var4;
         }

         try {
            if ((char)var10 != 'n' && 'm' != (char)var10) {
            }
         } catch (RuntimeException var7) {
            throw var7;
         }

         if ((char)var10 != '{') {
         }

         try {
            if (var3 == 37) {
               this.Z = false;
            }
         } catch (RuntimeException var6) {
            throw var6;
         }

         if (38 == var3) {
         }

         if (-126 == ~((char)var10)) {
         }
      } catch (RuntimeException var9) {
         RuntimeException var2 = var9;

         RuntimeException var10000;
         StringBuilder var10001;
         try {
            var10000 = var2;
            var10001 = new StringBuilder().append(Sb[58]);
            if (var1 != null) {
               throw i.a(var2, var10001.append(Sb[4]).append(')').toString());
            }
         } catch (RuntimeException var5) {
            throw var5;
         }

         throw i.a(var10000, var10001.append(Sb[3]).append(')').toString());
      }
   }

   synchronized void b(boolean var1) {
      try {
         if (!var1) {
            M++;
         }
      } catch (RuntimeException var3) {
         throw i.a(var3, Sb[48] + var1 + ')');
      }
   }

   synchronized void e(int var1) {
      try {
         if (var1 >= 64) {
            ob++;
         }
      } catch (RuntimeException var3) {
         throw i.a(var3, Sb[49] + var1 + ')');
      }
   }

   public final void paint(Graphics param1) {
      // $VF: Couldn't be decompiled
      // Please report this to the Vineflower issue tracker, at https://github.com/Vineflower/vineflower/issues with a copy of the class file (if you have the rights to distribute it!)
      // java.lang.RuntimeException: parsing failure!
      //   at org.jetbrains.java.decompiler.modules.decompiler.decompose.DomHelper.parseGraph(DomHelper.java:211)
      //   at org.jetbrains.java.decompiler.main.rels.MethodProcessor.codeToJava(MethodProcessor.java:166)
      //
      // Bytecode:
      // 00: aload 0
      // 01: bipush 1
      // 02: putfield e.N Z
      // 05: getstatic e.l I
      // 08: bipush 1
      // 09: iadd
      // 0a: putstatic e.l I
      // 0d: bipush -3
      // 0f: aload 0
      // 10: getfield e.n I
      // 13: bipush -1
      // 14: ixor
      // 15: if_icmpne 23
      // 18: aload 0
      // 19: getfield e.C Ljava/awt/Image;
      // 1c: ifnonnull 3f
      // 1f: goto 23
      // 22: athrow
      // 23: bipush 0
      // 24: aload 0
      // 25: getfield e.n I
      // 28: if_icmpeq 33
      // 2b: goto 2f
      // 2e: athrow
      // 2f: goto 51
      // 32: athrow
      // 33: aload 0
      // 34: bipush -89
      // 36: invokevirtual e.a (I)V
      // 39: getstatic client.vh Z
      // 3c: ifeq 51
      // 3f: aload 0
      // 40: aload 0
      // 41: getfield e.B Ljava/lang/String;
      // 44: aload 0
      // 45: getfield e.V I
      // 48: bipush 126
      // 4a: invokespecial e.a (Ljava/lang/String;II)V
      // 4d: goto 51
      // 50: athrow
      // 51: goto 87
      // 54: astore 2
      // 55: aload 2
      // 56: new java/lang/StringBuilder
      // 59: dup
      // 5a: invokespecial java/lang/StringBuilder.<init> ()V
      // 5d: getstatic e.Sb [Ljava/lang/String;
      // 60: bipush 10
      // 62: aaload
      // 63: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 66: aload 1
      // 67: ifnull 73
      // 6a: getstatic e.Sb [Ljava/lang/String;
      // 6d: bipush 4
      // 6e: aaload
      // 6f: goto 78
      // 72: athrow
      // 73: getstatic e.Sb [Ljava/lang/String;
      // 76: bipush 3
      // 77: aaload
      // 78: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 7b: bipush 41
      // 7d: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 80: invokevirtual java/lang/StringBuilder.toString ()Ljava/lang/String;
      // 83: invokestatic i.a (Ljava/lang/Throwable;Ljava/lang/String;)Lla;
      // 86: athrow
      // 87: return
   }

   private final void a(String param1, int param2, int param3) {
      // $VF: Couldn't be decompiled
      // Please report this to the Vineflower issue tracker, at https://github.com/Vineflower/vineflower/issues with a copy of the class file (if you have the rights to distribute it!)
      // java.lang.RuntimeException: parsing failure!
      //   at org.jetbrains.java.decompiler.modules.decompiler.decompose.DomHelper.parseGraph(DomHelper.java:211)
      //   at org.jetbrains.java.decompiler.main.rels.MethodProcessor.codeToJava(MethodProcessor.java:166)
      //
      // Bytecode:
      // 000: sipush -281
      // 003: aload 0
      // 004: getfield e.m I
      // 007: iadd
      // 008: bipush 2
      // 009: idiv
      // 00a: istore 4
      // 00c: sipush -148
      // 00f: aload 0
      // 010: getfield e.a I
      // 013: iadd
      // 014: bipush 2
      // 015: idiv
      // 016: istore 5
      // 018: aload 0
      // 019: getfield e.u Ljava/awt/Graphics;
      // 01c: getstatic java/awt/Color.black Ljava/awt/Color;
      // 01f: invokevirtual java/awt/Graphics.setColor (Ljava/awt/Color;)V
      // 022: aload 0
      // 023: getfield e.u Ljava/awt/Graphics;
      // 026: bipush 0
      // 027: bipush 0
      // 028: aload 0
      // 029: getfield e.m I
      // 02c: aload 0
      // 02d: getfield e.a I
      // 030: invokevirtual java/awt/Graphics.fillRect (IIII)V
      // 033: aload 0
      // 034: getfield e.hb Z
      // 037: ifne 04f
      // 03a: aload 0
      // 03b: getfield e.u Ljava/awt/Graphics;
      // 03e: aload 0
      // 03f: getfield e.C Ljava/awt/Image;
      // 042: iload 4
      // 044: iload 5
      // 046: aload 0
      // 047: invokevirtual java/awt/Graphics.drawImage (Ljava/awt/Image;IILjava/awt/image/ImageObserver;)Z
      // 04a: pop
      // 04b: goto 04f
      // 04e: athrow
      // 04f: iinc 4 2
      // 052: aload 0
      // 053: iload 2
      // 054: putfield e.V I
      // 057: iinc 5 90
      // 05a: aload 0
      // 05b: aload 1
      // 05c: putfield e.B Ljava/lang/String;
      // 05f: iload 3
      // 060: bipush 97
      // 062: if_icmpgt 071
      // 065: aload 0
      // 066: aconst_null
      // 067: checkcast java/awt/event/MouseEvent
      // 06a: invokevirtual e.mouseReleased (Ljava/awt/event/MouseEvent;)V
      // 06d: goto 071
      // 070: athrow
      // 071: aload 0
      // 072: getfield e.u Ljava/awt/Graphics;
      // 075: new java/awt/Color
      // 078: dup
      // 079: sipush 132
      // 07c: sipush 132
      // 07f: sipush 132
      // 082: invokespecial java/awt/Color.<init> (III)V
      // 085: invokevirtual java/awt/Graphics.setColor (Ljava/awt/Color;)V
      // 088: aload 0
      // 089: getfield e.hb Z
      // 08c: ifeq 0a6
      // 08f: aload 0
      // 090: getfield e.u Ljava/awt/Graphics;
      // 093: new java/awt/Color
      // 096: dup
      // 097: sipush 220
      // 09a: bipush 0
      // 09b: bipush 0
      // 09c: invokespecial java/awt/Color.<init> (III)V
      // 09f: invokevirtual java/awt/Graphics.setColor (Ljava/awt/Color;)V
      // 0a2: goto 0a6
      // 0a5: athrow
      // 0a6: aload 0
      // 0a7: getfield e.u Ljava/awt/Graphics;
      // 0aa: iload 4
      // 0ac: bipush -2
      // 0ae: iadd
      // 0af: iload 5
      // 0b1: bipush -2
      // 0b3: iadd
      // 0b4: sipush 280
      // 0b7: bipush 23
      // 0b9: invokevirtual java/awt/Graphics.drawRect (IIII)V
      // 0bc: aload 0
      // 0bd: getfield e.u Ljava/awt/Graphics;
      // 0c0: iload 4
      // 0c2: iload 5
      // 0c4: sipush 277
      // 0c7: iload 2
      // 0c8: imul
      // 0c9: bipush 100
      // 0cb: idiv
      // 0cc: bipush 20
      // 0ce: invokevirtual java/awt/Graphics.fillRect (IIII)V
      // 0d1: aload 0
      // 0d2: getfield e.u Ljava/awt/Graphics;
      // 0d5: new java/awt/Color
      // 0d8: dup
      // 0d9: sipush 198
      // 0dc: sipush 198
      // 0df: sipush 198
      // 0e2: invokespecial java/awt/Color.<init> (III)V
      // 0e5: invokevirtual java/awt/Graphics.setColor (Ljava/awt/Color;)V
      // 0e8: aload 0
      // 0e9: getfield e.hb Z
      // 0ec: ifeq 10a
      // 0ef: aload 0
      // 0f0: getfield e.u Ljava/awt/Graphics;
      // 0f3: new java/awt/Color
      // 0f6: dup
      // 0f7: sipush 255
      // 0fa: sipush 255
      // 0fd: sipush 255
      // 100: invokespecial java/awt/Color.<init> (III)V
      // 103: invokevirtual java/awt/Graphics.setColor (Ljava/awt/Color;)V
      // 106: goto 10a
      // 109: athrow
      // 10a: aload 0
      // 10b: aload 0
      // 10c: getfield e.tb Ljava/awt/Font;
      // 10f: aload 1
      // 110: bipush 10
      // 112: iload 5
      // 114: iadd
      // 115: bipush 1
      // 116: sipush 138
      // 119: iload 4
      // 11b: iadd
      // 11c: aload 0
      // 11d: getfield e.u Ljava/awt/Graphics;
      // 120: invokevirtual e.a (Ljava/awt/Font;Ljava/lang/String;IZILjava/awt/Graphics;)V
      // 123: aload 0
      // 124: getfield e.hb Z
      // 127: ifne 170
      // 12a: aload 0
      // 12b: aload 0
      // 12c: getfield e.X Ljava/awt/Font;
      // 12f: getstatic e.Sb [Ljava/lang/String;
      // 132: bipush 24
      // 134: aaload
      // 135: bipush 30
      // 137: iload 5
      // 139: iadd
      // 13a: bipush 1
      // 13b: iload 4
      // 13d: sipush -138
      // 140: isub
      // 141: aload 0
      // 142: getfield e.u Ljava/awt/Graphics;
      // 145: invokevirtual e.a (Ljava/awt/Font;Ljava/lang/String;IZILjava/awt/Graphics;)V
      // 148: aload 0
      // 149: aload 0
      // 14a: getfield e.X Ljava/awt/Font;
      // 14d: getstatic e.Sb [Ljava/lang/String;
      // 150: bipush 23
      // 152: aaload
      // 153: iload 5
      // 155: bipush 44
      // 157: iadd
      // 158: bipush 1
      // 159: iload 4
      // 15b: sipush -138
      // 15e: isub
      // 15f: aload 0
      // 160: getfield e.u Ljava/awt/Graphics;
      // 163: invokevirtual e.a (Ljava/awt/Font;Ljava/lang/String;IZILjava/awt/Graphics;)V
      // 166: getstatic client.vh Z
      // 169: ifeq 1ab
      // 16c: goto 170
      // 16f: athrow
      // 170: aload 0
      // 171: getfield e.u Ljava/awt/Graphics;
      // 174: new java/awt/Color
      // 177: dup
      // 178: sipush 132
      // 17b: sipush 132
      // 17e: sipush 152
      // 181: invokespecial java/awt/Color.<init> (III)V
      // 184: invokevirtual java/awt/Graphics.setColor (Ljava/awt/Color;)V
      // 187: aload 0
      // 188: aload 0
      // 189: getfield e.Jb Ljava/awt/Font;
      // 18c: getstatic e.Sb [Ljava/lang/String;
      // 18f: bipush 23
      // 191: aaload
      // 192: bipush -20
      // 194: aload 0
      // 195: getfield e.a I
      // 198: iadd
      // 199: bipush 1
      // 19a: sipush 138
      // 19d: iload 4
      // 19f: iadd
      // 1a0: aload 0
      // 1a1: getfield e.u Ljava/awt/Graphics;
      // 1a4: invokevirtual e.a (Ljava/awt/Font;Ljava/lang/String;IZILjava/awt/Graphics;)V
      // 1a7: goto 1ab
      // 1aa: athrow
      // 1ab: aconst_null
      // 1ac: aload 0
      // 1ad: getfield e.p Ljava/lang/String;
      // 1b0: if_acmpeq 1dd
      // 1b3: aload 0
      // 1b4: getfield e.u Ljava/awt/Graphics;
      // 1b7: getstatic java/awt/Color.white Ljava/awt/Color;
      // 1ba: invokevirtual java/awt/Graphics.setColor (Ljava/awt/Color;)V
      // 1bd: aload 0
      // 1be: aload 0
      // 1bf: getfield e.X Ljava/awt/Font;
      // 1c2: aload 0
      // 1c3: getfield e.p Ljava/lang/String;
      // 1c6: bipush -120
      // 1c8: iload 5
      // 1ca: iadd
      // 1cb: bipush 1
      // 1cc: iload 4
      // 1ce: sipush 138
      // 1d1: iadd
      // 1d2: aload 0
      // 1d3: getfield e.u Ljava/awt/Graphics;
      // 1d6: invokevirtual e.a (Ljava/awt/Font;Ljava/lang/String;IZILjava/awt/Graphics;)V
      // 1d9: goto 1dd
      // 1dc: athrow
      // 1dd: goto 1e2
      // 1e0: astore 4
      // 1e2: getstatic e.r I
      // 1e5: bipush 1
      // 1e6: iadd
      // 1e7: putstatic e.r I
      // 1ea: goto 234
      // 1ed: astore 4
      // 1ef: aload 4
      // 1f1: new java/lang/StringBuilder
      // 1f4: dup
      // 1f5: invokespecial java/lang/StringBuilder.<init> ()V
      // 1f8: getstatic e.Sb [Ljava/lang/String;
      // 1fb: bipush 22
      // 1fd: aaload
      // 1fe: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 201: aload 1
      // 202: ifnull 20e
      // 205: getstatic e.Sb [Ljava/lang/String;
      // 208: bipush 4
      // 209: aaload
      // 20a: goto 213
      // 20d: athrow
      // 20e: getstatic e.Sb [Ljava/lang/String;
      // 211: bipush 3
      // 212: aaload
      // 213: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 216: bipush 44
      // 218: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 21b: iload 2
      // 21c: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 21f: bipush 44
      // 221: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 224: iload 3
      // 225: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 228: bipush 41
      // 22a: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 22d: invokevirtual java/lang/StringBuilder.toString ()Ljava/lang/String;
      // 230: invokestatic i.a (Ljava/lang/Throwable;Ljava/lang/String;)Lla;
      // 233: athrow
      // 234: return
   }

   @Override
   public final synchronized void mouseReleased(MouseEvent var1) {
      try {
         g++;
         this.a(var1, (byte)-128);
         this.I = var1.getX() + -this.Eb;
         this.xb = var1.getY() + -this.K;
         this.Bb = 0;
      } catch (RuntimeException var4) {
         RuntimeException var2 = var4;

         RuntimeException var10000;
         StringBuilder var10001;
         try {
            var10000 = var2;
            var10001 = new StringBuilder().append(Sb[19]);
            if (var1 != null) {
               throw i.a(var2, var10001.append(Sb[4]).append(')').toString());
            }
         } catch (RuntimeException var3) {
            throw var3;
         }

         throw i.a(var10000, var10001.append(Sb[3]).append(')').toString());
      }
   }

   final void a(int var1, byte var2, String var3) {
      try {
         Lb++;

         try {
            int var13 = (-281 + this.m) / 2;
            var13 += 2;
            int var5 = (this.a - 148) / 2;
            this.B = var3;
            this.V = var1;
            var5 += 90;
            int var6 = 277 * var1 / 100;

            try {
               this.u.setColor(new Color(132, 132, 132));
               if (this.hb) {
                  this.u.setColor(new Color(220, 0, 0));
               }
            } catch (Exception var8) {
               throw var8;
            }

            label53: {
               try {
                  this.u.fillRect(var13, var5, var6, 20);
                  this.u.setColor(Color.black);
                  this.u.fillRect(var6 + var13, var5, -var6 + 277, 20);
                  this.u.setColor(new Color(198, 198, 198));
                  if (!this.hb) {
                     break label53;
                  }
               } catch (Exception var10) {
                  throw var10;
               }

               this.u.setColor(new Color(255, 255, 255));
            }

            this.a(this.tb, var3, 10 + var5, true, 138 + var13, this.u);
         } catch (Exception var11) {
         }

         try {
            if (var2 > -96) {
               this.x = (String)null;
            }
         } catch (Exception var7) {
            throw var7;
         }
      } catch (RuntimeException var12) {
         RuntimeException var4 = var12;

         RuntimeException var10000;
         StringBuilder var10001;
         try {
            var10000 = var4;
            var10001 = new StringBuilder().append(Sb[57]).append(var1).append(',').append((int)var2).append(',');
            if (var3 != null) {
               throw i.a(var4, var10001.append(Sb[4]).append(')').toString());
            }
         } catch (Exception var9) {
            throw var9;
         }

         throw i.a(var10000, var10001.append(Sb[3]).append(')').toString());
      }
   }

   private final Image a(byte[] var1, byte var2) {
      try {
         try {
            if (var2 != -54) {
               this.Ob = (String)null;
            }
         } catch (RuntimeException var5) {
            throw var5;
         }

         cb++;
         return pa.a(79, this, var1);
      } catch (RuntimeException var6) {
         RuntimeException var3 = var6;

         RuntimeException var10000;
         StringBuilder var10001;
         try {
            var10000 = var3;
            var10001 = new StringBuilder().append(Sb[68]);
            if (var1 != null) {
               throw i.a(var3, var10001.append(Sb[4]).append(',').append((int)var2).append(')').toString());
            }
         } catch (RuntimeException var4) {
            throw var4;
         }

         throw i.a(var10000, var10001.append(Sb[3]).append(',').append((int)var2).append(')').toString());
      }
   }

   public static final void provideLoaderApplet(Applet var0) {
      try {
         da.gb = var0;
         t++;
      } catch (RuntimeException var3) {
         RuntimeException var1 = var3;

         RuntimeException var10000;
         StringBuilder var10001;
         try {
            var10000 = var1;
            var10001 = new StringBuilder().append(Sb[54]);
            if (var0 != null) {
               throw i.a(var1, var10001.append(Sb[4]).append(')').toString());
            }
         } catch (RuntimeException var2) {
            throw var2;
         }

         throw i.a(var10000, var10001.append(Sb[3]).append(')').toString());
      }
   }

   final void a(int var1) {
      try {
         if (var1 < -54) {
            Db++;
         }
      } catch (RuntimeException var3) {
         throw i.a(var3, Sb[75] + var1 + ')');
      }
   }

   public final URL getDocumentBase() {
      try {
         jb++;
         if (kb.a != null) {
            return null;
         } else {
            return null != da.gb ? da.gb.getDocumentBase() : super.getDocumentBase();
         }
      } catch (RuntimeException var2) {
         throw i.a(var2, Sb[51]);
      }
   }

   final void a(boolean var1, String var2, int var3, String var4, int var5, byte var6, int var7, int var8, int var9) {
      try {
         mb++;

         try {
            System.out.println(Sb[16]);
            this.m = var8;
            this.a = var9;
            kb.a = new qb(this, 800, 600, var2, var1, false);

            try {
               kb.a.getClass().getMethod(Sb[17], boolean.class).invoke(kb.a, Boolean.FALSE);
            } catch (Exception var12) {
            }

            db.d = var7;
            this.n = 1;
            pa.b = pa.k = new c(var3, var4, 0, true);

            try {
               if (var6 <= 20) {
                  return;
               }

               cb.a(new URL(Sb[15], Sb[14], var5, ""), this, -91);
            } catch (IOException var11) {
               mb.a(2097151, var11, null);
            }

            this.z = new Thread(this);
            this.z.start();
            this.z.setPriority(1);
         } catch (Exception var13) {
            mb.a(2097151, var13, null);
         }
      } catch (RuntimeException var16) {
         RuntimeException var10 = var16;

         RuntimeException var10000;
         StringBuilder var10001;
         String var10002;
         label63: {
            try {
               var10000 = var10;
               var10001 = new StringBuilder().append(Sb[18]).append(var1).append(',');
               if (var2 != null) {
                  var10002 = Sb[4];
                  break label63;
               }
            } catch (Exception var15) {
               throw var15;
            }

            var10002 = Sb[3];
         }

         try {
            var10001 = var10001.append(var10002).append(',').append(var3).append(',');
            if (var4 != null) {
               throw i.a(
                  var10000,
                  var10001.append(Sb[4])
                     .append(',')
                     .append(var5)
                     .append(',')
                     .append((int)var6)
                     .append(',')
                     .append(var7)
                     .append(',')
                     .append(var8)
                     .append(',')
                     .append(var9)
                     .append(')')
                     .toString()
               );
            }
         } catch (Exception var14) {
            throw var14;
         }

         throw i.a(
            var10000,
            var10001.append(Sb[3])
               .append(',')
               .append(var5)
               .append(',')
               .append((int)var6)
               .append(',')
               .append(var7)
               .append(',')
               .append(var8)
               .append(',')
               .append(var9)
               .append(')')
               .toString()
         );
      }
   }

   private final boolean b(byte var1) {
      try {
         O++;
         byte[] var2 = this.a(Sb[32], 0, 3, 85);

         try {
            if (var2 == null) {
               return false;
            }
         } catch (RuntimeException var8) {
            throw var8;
         }

         try {
            if (var1 != 118) {
               this.B = (String)null;
            }
         } catch (RuntimeException var7) {
            throw var7;
         }

         byte[] var3 = na.a(Sb[30], 0, var2, -120);

         try {
            this.C = this.a(var3, (byte)-54);
            if (!qa.a(this, Sb[29], 0, var1 + -118)) {
               return false;
            }
         } catch (RuntimeException var6) {
            throw var6;
         }

         if (!qa.a(this, Sb[39], 1, 0)) {
            return false;
         } else if (!qa.a(this, Sb[31], 2, 0)) {
            return false;
         } else if (!qa.a(this, Sb[36], 3, 0)) {
            return false;
         } else {
            try {
               if (!qa.a(this, Sb[38], 4, var1 + -118)) {
                  return false;
               }
            } catch (RuntimeException var5) {
               throw var5;
            }

            try {
               if (!qa.a(this, Sb[34], 5, 0)) {
                  return false;
               }
            } catch (RuntimeException var4) {
               throw var4;
            }

            return !qa.a(this, Sb[35], 6, var1 ^ 118) ? false : qa.a(this, Sb[37], 7, var1 + -118);
         }
      } catch (RuntimeException var9) {
         throw i.a(var9, Sb[33] + var1 + ')');
      }
   }

   void a(boolean var1) {
      try {
         k++;
         if (var1) {
            this.vb = 85;
         }
      } catch (RuntimeException var3) {
         throw i.a(var3, Sb[72] + var1 + ')');
      }
   }

   public final AppletContext getAppletContext() {
      try {
         f++;
         if (null != kb.a) {
            return null;
         } else {
            return da.gb != null ? da.gb.getAppletContext() : super.getAppletContext();
         }
      } catch (RuntimeException var2) {
         throw i.a(var2, Sb[41]);
      }
   }

   void a(int var1, int var2, int var3, int var4) {
      try {
         s++;
         if (var2 < 87) {
            this.z = (Thread)null;
         }
      } catch (RuntimeException var6) {
         throw i.a(var6, Sb[50] + var1 + ',' + var2 + ',' + var3 + ',' + var4 + ')');
      }
   }

   // $VF: Handled exception range with multiple entry points by splitting it
   // $VF: Inserted dummy exception handlers to handle obfuscated exceptions
   private final void a(String var1, boolean var2) {
      RuntimeException var10000;
      label66: {
         try {
            P++;
            if (this.Kb) {
               return;
            }
         } catch (RuntimeException var10) {
            var10000 = var10;
            boolean var10001 = false;
            break label66;
         }

         try {
            this.Kb = var2;
            System.out.println(Sb[63] + var1);

            try {
               label75: {
                  if (null != da.gb) {
                     try {
                        a.a(Sb[59], (byte)82, da.gb);
                        if (!client.vh) {
                           break label75;
                        }
                     } catch (Throwable var7) {
                        boolean var15 = false;
                        throw var7;
                     }
                  }

                  try {
                     a.a(Sb[59], (byte)-73, this);
                  } catch (Throwable var6) {
                     boolean var16 = false;
                     throw var6;
                  }
               }
            } catch (Throwable var8) {
            }

            try {
               this.getAppletContext().showDocument(new URL(this.getCodeBase(), Sb[63] + var1 + Sb[62]), Sb[60]);
            } catch (Exception var4) {
            }

            return;
         } catch (RuntimeException var9) {
            var10000 = var9;
            boolean var13 = false;
         }
      }

      RuntimeException var3 = var10000;

      StringBuilder var14;
      try {
         var10000 = var3;
         var14 = new StringBuilder().append(Sb[61]);
         if (var1 != null) {
            throw i.a(var3, var14.append(Sb[4]).append(',').append(var2).append(')').toString());
         }
      } catch (Throwable var5) {
         throw var5;
      }

      throw i.a(var10000, var14.append(Sb[3]).append(',').append(var2).append(')').toString());
   }

   @Override
   public final synchronized void mouseDragged(MouseEvent param1) {
      // $VF: Couldn't be decompiled
      // Please report this to the Vineflower issue tracker, at https://github.com/Vineflower/vineflower/issues with a copy of the class file (if you have the rights to distribute it!)
      // java.lang.RuntimeException: parsing failure!
      //   at org.jetbrains.java.decompiler.modules.decompiler.decompose.DomHelper.parseGraph(DomHelper.java:211)
      //   at org.jetbrains.java.decompiler.main.rels.MethodProcessor.codeToJava(MethodProcessor.java:166)
      //
      // Bytecode:
      // 00: aload 0
      // 01: aload 1
      // 02: bipush -128
      // 04: invokespecial e.a (Ljava/awt/event/InputEvent;B)V
      // 07: getstatic e.q I
      // 0a: bipush 1
      // 0b: iadd
      // 0c: putstatic e.q I
      // 0f: aload 0
      // 10: aload 1
      // 11: invokevirtual java/awt/event/MouseEvent.getX ()I
      // 14: aload 0
      // 15: getfield e.Eb I
      // 18: isub
      // 19: putfield e.I I
      // 1c: aload 0
      // 1d: aload 1
      // 1e: invokevirtual java/awt/event/MouseEvent.getY ()I
      // 21: aload 0
      // 22: getfield e.K I
      // 25: isub
      // 26: putfield e.xb I
      // 29: aload 1
      // 2a: invokevirtual java/awt/event/MouseEvent.isMetaDown ()Z
      // 2d: ifeq 3f
      // 30: aload 0
      // 31: bipush 2
      // 32: putfield e.Bb I
      // 35: getstatic client.vh Z
      // 38: ifeq 48
      // 3b: goto 3f
      // 3e: athrow
      // 3f: aload 0
      // 40: bipush 1
      // 41: putfield e.Bb I
      // 44: goto 48
      // 47: athrow
      // 48: goto 7e
      // 4b: astore 2
      // 4c: aload 2
      // 4d: new java/lang/StringBuilder
      // 50: dup
      // 51: invokespecial java/lang/StringBuilder.<init> ()V
      // 54: getstatic e.Sb [Ljava/lang/String;
      // 57: bipush 76
      // 59: aaload
      // 5a: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 5d: aload 1
      // 5e: ifnull 6a
      // 61: getstatic e.Sb [Ljava/lang/String;
      // 64: bipush 4
      // 65: aaload
      // 66: goto 6f
      // 69: athrow
      // 6a: getstatic e.Sb [Ljava/lang/String;
      // 6d: bipush 3
      // 6e: aaload
      // 6f: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 72: bipush 41
      // 74: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 77: invokevirtual java/lang/StringBuilder.toString ()Ljava/lang/String;
      // 7a: invokestatic i.a (Ljava/lang/Throwable;Ljava/lang/String;)Lla;
      // 7d: athrow
      // 7e: return
   }

   final void a(int var1, int var2, int var3, int var4, int var5) {
      try {
         G++;

         try {
            System.out.println(Sb[69]);
            this.n = 1;
            this.m = var5;
            this.a = var1;
            nb.s = this.getCodeBase();
            db.d = var2;
            if (pa.k == null) {
               int var10002;
               Object var10003;
               byte var10004;
               boolean var10005;
               label58: {
                  try {
                     var10002 = var3;
                     var10003 = null;
                     var10004 = 0;
                     if (null != da.gb) {
                        var10005 = true;
                        break label58;
                     }
                  } catch (Throwable var12) {
                     throw var12;
                  }

                  var10005 = false;
               }

               c var10000 = new c(var10002, (String)var10003, var10004, var10005);
               pa.k = var10000;
               pa.b = var10000;
            }

            if (var4 != 2) {
               return;
            }

            if (null != da.gb) {
               Method var6 = c.y;
               if (null != var6) {
                  try {
                     var6.invoke(da.gb, Boolean.TRUE);
                  } catch (Throwable var11) {
                  }
               }

               Method var7 = c.u;
               if (null != var7) {
                  try {
                     var7.invoke(da.gb, Boolean.FALSE);
                  } catch (Throwable var10) {
                  }
               }
            }

            try {
               cb.a(this.getCodeBase(), this, var4 + -110);
            } catch (IOException var9) {
               var9.printStackTrace();
            }

            this.a(var4 + -1, this);
         } catch (Exception var13) {
            mb.a(var4 ^ 2097149, var13, null);
            this.a(Sb[12], true);
         }
      } catch (RuntimeException var14) {
         throw i.a(var14, Sb[70] + var1 + ',' + var2 + ',' + var3 + ',' + var4 + ',' + var5 + ')');
      }
   }

   @Override
   public final void mouseEntered(MouseEvent var1) {
      try {
         this.a(var1, (byte)-128);
         Rb++;
      } catch (RuntimeException var4) {
         RuntimeException var2 = var4;

         RuntimeException var10000;
         StringBuilder var10001;
         try {
            var10000 = var2;
            var10001 = new StringBuilder().append(Sb[20]);
            if (var1 != null) {
               throw i.a(var2, var10001.append(Sb[4]).append(')').toString());
            }
         } catch (RuntimeException var3) {
            throw var3;
         }

         throw i.a(var10000, var10001.append(Sb[3]).append(')').toString());
      }
   }

   final boolean d(int var1) {
      try {
         ab++;
         Graphics var2 = this.getGraphics();

         try {
            if (var2 == null) {
               return false;
            }
         } catch (RuntimeException var4) {
            throw var4;
         }

         try {
            if (var1 != 2) {
               this.sb = -7;
            }
         } catch (RuntimeException var3) {
            throw var3;
         }

         this.u = var2.create();
         this.u.translate(this.Eb, this.K);
         this.u.setColor(Color.black);
         this.u.fillRect(0, 0, this.m, this.a);
         this.a(Sb[56], 0, var1 ^ 103);
         return true;
      } catch (RuntimeException var5) {
         throw i.a(var5, Sb[55] + var1 + ')');
      }
   }

   private final void a(InputEvent var1, byte var2) {
      try {
         if (var2 <= -127) {
            d++;
            int var8 = var1.getModifiers();

            e var9;
            boolean var11;
            label50: {
               try {
                  var9 = this;
                  if (0 != (var8 & 2)) {
                     var11 = true;
                     break label50;
                  }
               } catch (RuntimeException var6) {
                  throw var6;
               }

               var11 = false;
            }

            label43: {
               try {
                  var9.bb = var11;
                  var9 = this;
                  if ((var8 & 1) != 0) {
                     var11 = true;
                     break label43;
                  }
               } catch (RuntimeException var5) {
                  throw var5;
               }

               var11 = false;
            }

            var9.gb = var11;
         }
      } catch (RuntimeException var7) {
         RuntimeException var3 = var7;

         RuntimeException var10000;
         StringBuilder var10001;
         try {
            var10000 = var3;
            var10001 = new StringBuilder().append(Sb[74]);
            if (var1 != null) {
               throw i.a(var3, var10001.append(Sb[4]).append(',').append((int)var2).append(')').toString());
            }
         } catch (RuntimeException var4) {
            throw var4;
         }

         throw i.a(var10000, var10001.append(Sb[3]).append(',').append((int)var2).append(')').toString());
      }
   }

   public final Dimension getSize() {
      try {
         db++;
         if (kb.a != null) {
            return kb.a.getSize();
         } else {
            return null != da.gb ? da.gb.getSize() : super.getSize();
         }
      } catch (RuntimeException var2) {
         throw i.a(var2, Sb[67]);
      }
   }

   public final Graphics getGraphics() {
      try {
         y++;
         if (kb.a != null) {
            return kb.a.getGraphics();
         } else {
            return da.gb != null ? da.gb.getGraphics() : super.getGraphics();
         }
      } catch (RuntimeException var2) {
         throw i.a(var2, Sb[8]);
      }
   }

   public final URL getCodeBase() {
      try {
         Y++;
         if (null != kb.a) {
            return null;
         } else {
            return da.gb != null ? da.gb.getCodeBase() : super.getCodeBase();
         }
      } catch (RuntimeException var2) {
         throw var2;
      }
   }

   @Override
   public final void mouseClicked(MouseEvent var1) {
      try {
         h++;
         this.a(var1, (byte)-128);
      } catch (RuntimeException var4) {
         RuntimeException var2 = var4;

         RuntimeException var10000;
         StringBuilder var10001;
         try {
            var10000 = var2;
            var10001 = new StringBuilder().append(Sb[7]);
            if (var1 != null) {
               throw i.a(var2, var10001.append(Sb[4]).append(')').toString());
            }
         } catch (RuntimeException var3) {
            throw var3;
         }

         throw i.a(var10000, var10001.append(Sb[3]).append(')').toString());
      }
   }

   public final Image createImage(int var1, int var2) {
      try {
         Gb++;
         if (null != kb.a) {
            return kb.a.createImage(var1, var2);
         } else {
            return null != da.gb ? da.gb.createImage(var1, var2) : super.createImage(var1, var2);
         }
      } catch (RuntimeException var4) {
         throw i.a(var4, Sb[71] + var1 + ',' + var2 + ')');
      }
   }

   final void a(Font var1, String var2, int var3, boolean var4, int var5, Graphics var6) {
      try {
         Object var14;
         label66: {
            D++;
            if (null == kb.a) {
               var14 = this;
               if (!client.vh) {
                  break label66;
               }
            }

            var14 = kb.a;
         }

         FontMetrics var8 = var14.getFontMetrics(var1);

         try {
            var8.stringWidth(var2);
            if (!var4) {
               this.c(68);
            }
         } catch (RuntimeException var12) {
            throw var12;
         }

         var6.setFont(var1);
         var6.drawString(var2, var5 + -(var8.stringWidth(var2) / 2), var3 - -(var8.getHeight() / 4));
      } catch (RuntimeException var13) {
         RuntimeException var7 = var13;

         RuntimeException var10000;
         StringBuilder var10001;
         String var10002;
         label54: {
            try {
               var10000 = var7;
               var10001 = new StringBuilder().append(Sb[2]);
               if (var1 != null) {
                  var10002 = Sb[4];
                  break label54;
               }
            } catch (RuntimeException var11) {
               throw var11;
            }

            var10002 = Sb[3];
         }

         label47: {
            try {
               var10001 = var10001.append(var10002).append(',');
               if (var2 != null) {
                  var10002 = Sb[4];
                  break label47;
               }
            } catch (RuntimeException var10) {
               throw var10;
            }

            var10002 = Sb[3];
         }

         try {
            var10001 = var10001.append(var10002).append(',').append(var3).append(',').append(var4).append(',').append(var5).append(',');
            if (var6 != null) {
               throw i.a(var10000, var10001.append(Sb[4]).append(')').toString());
            }
         } catch (RuntimeException var9) {
            throw var9;
         }

         throw i.a(var10000, var10001.append(Sb[3]).append(')').toString());
      }
   }

   public final void destroy() {
      try {
         try {
            this.vb = -1;
            Hb++;
            mb.a(11200, 5000L);
            if (-1 != this.vb) {
               return;
            }
         } catch (RuntimeException var3) {
            throw var3;
         }

         try {
            System.out.println(Sb[0]);
            this.b(100);
            if (this.z != null) {
               this.z.stop();
               this.z = null;
            }
         } catch (RuntimeException var2) {
            throw var2;
         }
      } catch (RuntimeException var4) {
         throw i.a(var4, Sb[1]);
      }
   }

   void a(byte var1, int var2) {
      try {
         try {
            if (var1 <= 105) {
               this.a((InputEvent)null, (byte)83);
            }
         } catch (RuntimeException var4) {
            throw var4;
         }

         c++;
      } catch (RuntimeException var5) {
         throw i.a(var5, Sb[46] + var1 + ',' + var2 + ')');
      }
   }

   @Override
   public final synchronized void mouseMoved(MouseEvent var1) {
      try {
         this.a(var1, (byte)-128);
         yb++;
         this.I = var1.getX() - this.Eb;
         this.xb = var1.getY() + -this.K;
         this.sb = 0;
         this.Bb = 0;
      } catch (RuntimeException var4) {
         RuntimeException var2 = var4;

         RuntimeException var10000;
         StringBuilder var10001;
         try {
            var10000 = var2;
            var10001 = new StringBuilder().append(Sb[45]);
            if (var1 != null) {
               throw i.a(var2, var10001.append(Sb[4]).append(')').toString());
            }
         } catch (RuntimeException var3) {
            throw var3;
         }

         throw i.a(var10000, var10001.append(Sb[3]).append(')').toString());
      }
   }

   final void c(int var1) {
      boolean var3 = client.vh;

      try {
         if (var1 == -28492) {
            int var2 = 0;

            while (true) {
               if (var2 < 10) {
                  try {
                     this.F[var2] = 0L;
                     var2++;
                     if (var3) {
                        break;
                     }

                     if (!var3) {
                        continue;
                     }
                  } catch (RuntimeException var4) {
                     throw var4;
                  }
               }

               R++;
               break;
            }
         }
      } catch (RuntimeException var5) {
         throw i.a(var5, Sb[53] + var1 + ')');
      }
   }

   final byte[] a(String var1, int var2, int var3, int var4) {
      try {
         try {
            J++;
            if (var4 <= 53) {
               this.c(15);
            }
         } catch (IOException var8) {
            throw var8;
         }

         try {
            return ib.a(-101, var1, var2, var3);
         } catch (IOException var6) {
            mb.a(2097151, var6, Sb[65] + var3);
            return null;
         }
      } catch (RuntimeException var9) {
         RuntimeException var5 = var9;

         RuntimeException var10000;
         StringBuilder var10001;
         try {
            var10000 = var5;
            var10001 = new StringBuilder().append(Sb[66]);
            if (var1 != null) {
               throw i.a(var5, var10001.append(Sb[4]).append(',').append(var2).append(',').append(var3).append(',').append(var4).append(')').toString());
            }
         } catch (IOException var7) {
            throw var7;
         }

         throw i.a(var10000, var10001.append(Sb[3]).append(',').append(var2).append(',').append(var3).append(',').append(var4).append(')').toString());
      }
   }

   @Override
   public final void keyTyped(KeyEvent var1) {
      try {
         Nb++;
         this.a(var1, (byte)-128);
      } catch (RuntimeException var4) {
         RuntimeException var2 = var4;

         RuntimeException var10000;
         StringBuilder var10001;
         try {
            var10000 = var2;
            var10001 = new StringBuilder().append(Sb[6]);
            if (var1 != null) {
               throw i.a(var2, var10001.append(Sb[4]).append(')').toString());
            }
         } catch (RuntimeException var3) {
            throw var3;
         }

         throw i.a(var10000, var10001.append(Sb[3]).append(')').toString());
      }
   }

   public final void start() {
      try {
         Fb++;
         if (this.vb >= 0) {
            this.vb = 0;
         }
      } catch (RuntimeException var2) {
         throw i.a(var2, Sb[25]);
      }
   }

   @Override
   public final void mouseExited(MouseEvent var1) {
      try {
         eb++;
         this.a(var1, (byte)-128);
      } catch (RuntimeException var4) {
         RuntimeException var2 = var4;

         RuntimeException var10000;
         StringBuilder var10001;
         try {
            var10000 = var2;
            var10001 = new StringBuilder().append(Sb[64]);
            if (var1 != null) {
               throw i.a(var2, var10001.append(Sb[4]).append(')').toString());
            }
         } catch (RuntimeException var3) {
            throw var3;
         }

         throw i.a(var10000, var10001.append(Sb[3]).append(')').toString());
      }
   }

   @Override
   public final synchronized void mousePressed(MouseEvent param1) {
      // $VF: Couldn't be decompiled
      // Please report this to the Vineflower issue tracker, at https://github.com/Vineflower/vineflower/issues with a copy of the class file (if you have the rights to distribute it!)
      // java.lang.RuntimeException: parsing failure!
      //   at org.jetbrains.java.decompiler.modules.decompiler.decompose.DomHelper.parseGraph(DomHelper.java:211)
      //   at org.jetbrains.java.decompiler.main.rels.MethodProcessor.codeToJava(MethodProcessor.java:166)
      //
      // Bytecode:
      // 00: aload 0
      // 01: aload 1
      // 02: bipush -128
      // 04: invokespecial e.a (Ljava/awt/event/InputEvent;B)V
      // 07: getstatic e.Pb I
      // 0a: bipush 1
      // 0b: iadd
      // 0c: putstatic e.Pb I
      // 0f: aload 0
      // 10: aload 1
      // 11: invokevirtual java/awt/event/MouseEvent.getX ()I
      // 14: aload 0
      // 15: getfield e.Eb I
      // 18: isub
      // 19: putfield e.I I
      // 1c: aload 0
      // 1d: aload 1
      // 1e: invokevirtual java/awt/event/MouseEvent.getY ()I
      // 21: aload 0
      // 22: getfield e.K I
      // 25: ineg
      // 26: iadd
      // 27: putfield e.xb I
      // 2a: aload 1
      // 2b: invokevirtual java/awt/event/MouseEvent.isMetaDown ()Z
      // 2e: ifne 40
      // 31: aload 0
      // 32: bipush 1
      // 33: putfield e.Bb I
      // 36: getstatic client.vh Z
      // 39: ifeq 49
      // 3c: goto 40
      // 3f: athrow
      // 40: aload 0
      // 41: bipush 2
      // 42: putfield e.Bb I
      // 45: goto 49
      // 48: athrow
      // 49: aload 0
      // 4a: aload 0
      // 4b: getfield e.Bb I
      // 4e: putfield e.Qb I
      // 51: aload 0
      // 52: bipush 0
      // 53: putfield e.sb I
      // 56: aload 0
      // 57: aload 0
      // 58: getfield e.I I
      // 5b: bipush 94
      // 5d: aload 0
      // 5e: getfield e.Bb I
      // 61: aload 0
      // 62: getfield e.xb I
      // 65: invokevirtual e.a (IIII)V
      // 68: goto 9e
      // 6b: astore 2
      // 6c: aload 2
      // 6d: new java/lang/StringBuilder
      // 70: dup
      // 71: invokespecial java/lang/StringBuilder.<init> ()V
      // 74: getstatic e.Sb [Ljava/lang/String;
      // 77: bipush 9
      // 79: aaload
      // 7a: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 7d: aload 1
      // 7e: ifnull 8a
      // 81: getstatic e.Sb [Ljava/lang/String;
      // 84: bipush 4
      // 85: aaload
      // 86: goto 8f
      // 89: athrow
      // 8a: getstatic e.Sb [Ljava/lang/String;
      // 8d: bipush 3
      // 8e: aaload
      // 8f: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 92: bipush 41
      // 94: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 97: invokevirtual java/lang/StringBuilder.toString ()Ljava/lang/String;
      // 9a: invokestatic i.a (Ljava/lang/Throwable;Ljava/lang/String;)Lla;
      // 9d: athrow
      // 9e: return
   }

   void a(byte var1) {
      try {
         o++;
         if (var1 != -92) {
            provideLoaderApplet((Applet)null);
         }
      } catch (RuntimeException var3) {
         throw i.a(var3, Sb[73] + var1 + ')');
      }
   }

   @Override
   public final synchronized void keyPressed(KeyEvent param1) {
      // $VF: Couldn't be decompiled
      // Please report this to the Vineflower issue tracker, at https://github.com/Vineflower/vineflower/issues with a copy of the class file (if you have the rights to distribute it!)
      // java.lang.RuntimeException: parsing failure!
      //   at org.jetbrains.java.decompiler.modules.decompiler.decompose.DomHelper.parseGraph(DomHelper.java:211)
      //   at org.jetbrains.java.decompiler.main.rels.MethodProcessor.codeToJava(MethodProcessor.java:166)
      //
      // Bytecode:
      // 000: getstatic client.vh Z
      // 003: istore 6
      // 005: getstatic e.L I
      // 008: bipush 1
      // 009: iadd
      // 00a: putstatic e.L I
      // 00d: aload 0
      // 00e: aload 1
      // 00f: bipush -128
      // 011: invokespecial e.a (Ljava/awt/event/InputEvent;B)V
      // 014: aload 1
      // 015: invokevirtual java/awt/event/KeyEvent.getKeyChar ()C
      // 018: istore 2
      // 019: aload 1
      // 01a: invokevirtual java/awt/event/KeyEvent.getKeyCode ()I
      // 01d: istore 3
      // 01e: aload 0
      // 01f: bipush 126
      // 021: iload 2
      // 022: invokevirtual e.a (BI)V
      // 025: bipush -113
      // 027: iload 3
      // 028: bipush -1
      // 029: ixor
      // 02a: if_icmpeq 031
      // 02d: goto 042
      // 030: athrow
      // 031: aload 0
      // 032: aload 0
      // 033: getfield e.U Z
      // 036: ifne 03e
      // 039: bipush 1
      // 03a: goto 03f
      // 03d: athrow
      // 03e: bipush 0
      // 03f: putfield e.U Z
      // 042: bipush -79
      // 044: iload 2
      // 045: i2c
      // 046: bipush -1
      // 047: ixor
      // 048: if_icmpeq 056
      // 04b: bipush 77
      // 04d: iload 2
      // 04e: i2c
      // 04f: if_icmpeq 056
      // 052: goto 056
      // 055: athrow
      // 056: iload 2
      // 057: i2c
      // 058: bipush -1
      // 059: ixor
      // 05a: bipush -33
      // 05c: if_icmpne 05f
      // 05f: bipush -124
      // 061: iload 2
      // 062: i2c
      // 063: bipush -1
      // 064: ixor
      // 065: if_icmpeq 068
      // 068: aload 0
      // 069: bipush 0
      // 06a: putfield e.sb I
      // 06d: iload 2
      // 06e: i2c
      // 06f: bipush -1
      // 070: ixor
      // 071: bipush -111
      // 073: if_icmpeq 083
      // 076: iload 2
      // 077: i2c
      // 078: bipush -1
      // 079: ixor
      // 07a: bipush -110
      // 07c: if_icmpeq 083
      // 07f: goto 083
      // 082: athrow
      // 083: bipush -41
      // 085: iload 3
      // 086: bipush -1
      // 087: ixor
      // 088: if_icmpeq 08b
      // 08b: iload 3
      // 08c: bipush -1
      // 08d: ixor
      // 08e: bipush -40
      // 090: if_icmpne 09c
      // 093: aload 0
      // 094: bipush 1
      // 095: putfield e.E Z
      // 098: goto 09c
      // 09b: athrow
      // 09c: iload 2
      // 09d: i2c
      // 09e: bipush 125
      // 0a0: if_icmpne 0a3
      // 0a3: iload 3
      // 0a4: bipush -1
      // 0a5: ixor
      // 0a6: bipush -39
      // 0a8: if_icmpne 0ab
      // 0ab: iload 3
      // 0ac: bipush -1
      // 0ad: ixor
      // 0ae: bipush -38
      // 0b0: if_icmpne 0bc
      // 0b3: aload 0
      // 0b4: bipush 1
      // 0b5: putfield e.Z Z
      // 0b8: goto 0bc
      // 0bb: athrow
      // 0bc: bipush 0
      // 0bd: istore 4
      // 0bf: bipush 0
      // 0c0: istore 5
      // 0c2: iload 5
      // 0c4: getstatic i.f Ljava/lang/String;
      // 0c7: invokevirtual java/lang/String.length ()I
      // 0ca: if_icmpge 0fd
      // 0cd: iload 2
      // 0ce: bipush -1
      // 0cf: ixor
      // 0d0: getstatic i.f Ljava/lang/String;
      // 0d3: iload 5
      // 0d5: invokevirtual java/lang/String.charAt (I)C
      // 0d8: bipush -1
      // 0d9: ixor
      // 0da: iload 6
      // 0dc: ifne 168
      // 0df: if_icmpeq 0e9
      // 0e2: goto 0e6
      // 0e5: athrow
      // 0e6: goto 0f1
      // 0e9: bipush 1
      // 0ea: istore 4
      // 0ec: iload 6
      // 0ee: ifeq 0fd
      // 0f1: iinc 5 1
      // 0f4: iload 6
      // 0f6: ifeq 0c2
      // 0f9: goto 0fd
      // 0fc: athrow
      // 0fd: iload 4
      // 0ff: ifeq 132
      // 102: aload 0
      // 103: getfield e.e Ljava/lang/String;
      // 106: invokevirtual java/lang/String.length ()I
      // 109: bipush -1
      // 10a: ixor
      // 10b: bipush -21
      // 10d: if_icmple 132
      // 110: goto 114
      // 113: athrow
      // 114: new java/lang/StringBuilder
      // 117: dup
      // 118: invokespecial java/lang/StringBuilder.<init> ()V
      // 11b: aload 0
      // 11c: dup_x1
      // 11d: getfield e.e Ljava/lang/String;
      // 120: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 123: iload 2
      // 124: i2c
      // 125: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 128: invokevirtual java/lang/StringBuilder.toString ()Ljava/lang/String;
      // 12b: putfield e.e Ljava/lang/String;
      // 12e: goto 132
      // 131: athrow
      // 132: iload 4
      // 134: ifeq 165
      // 137: bipush 80
      // 139: aload 0
      // 13a: getfield e.x Ljava/lang/String;
      // 13d: invokevirtual java/lang/String.length ()I
      // 140: if_icmpgt 14b
      // 143: goto 147
      // 146: athrow
      // 147: goto 165
      // 14a: athrow
      // 14b: new java/lang/StringBuilder
      // 14e: dup
      // 14f: invokespecial java/lang/StringBuilder.<init> ()V
      // 152: aload 0
      // 153: dup_x1
      // 154: getfield e.x Ljava/lang/String;
      // 157: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 15a: iload 2
      // 15b: i2c
      // 15c: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 15f: invokevirtual java/lang/StringBuilder.toString ()Ljava/lang/String;
      // 162: putfield e.x Ljava/lang/String;
      // 165: bipush 8
      // 167: iload 2
      // 168: if_icmpne 195
      // 16b: bipush -1
      // 16c: aload 0
      // 16d: getfield e.e Ljava/lang/String;
      // 170: invokevirtual java/lang/String.length ()I
      // 173: bipush -1
      // 174: ixor
      // 175: if_icmple 195
      // 178: goto 17c
      // 17b: athrow
      // 17c: aload 0
      // 17d: aload 0
      // 17e: getfield e.e Ljava/lang/String;
      // 181: bipush 0
      // 182: aload 0
      // 183: getfield e.e Ljava/lang/String;
      // 186: invokevirtual java/lang/String.length ()I
      // 189: bipush 1
      // 18a: isub
      // 18b: invokevirtual java/lang/String.substring (II)Ljava/lang/String;
      // 18e: putfield e.e Ljava/lang/String;
      // 191: goto 195
      // 194: athrow
      // 195: bipush 8
      // 197: iload 2
      // 198: if_icmpne 1c5
      // 19b: aload 0
      // 19c: getfield e.x Ljava/lang/String;
      // 19f: invokevirtual java/lang/String.length ()I
      // 1a2: bipush -1
      // 1a3: ixor
      // 1a4: bipush -1
      // 1a5: if_icmpge 1c5
      // 1a8: goto 1ac
      // 1ab: athrow
      // 1ac: aload 0
      // 1ad: aload 0
      // 1ae: getfield e.x Ljava/lang/String;
      // 1b1: bipush 0
      // 1b2: bipush -1
      // 1b3: aload 0
      // 1b4: getfield e.x Ljava/lang/String;
      // 1b7: invokevirtual java/lang/String.length ()I
      // 1ba: iadd
      // 1bb: invokevirtual java/lang/String.substring (II)Ljava/lang/String;
      // 1be: putfield e.x Ljava/lang/String;
      // 1c1: goto 1c5
      // 1c4: athrow
      // 1c5: bipush -11
      // 1c7: iload 2
      // 1c8: bipush -1
      // 1c9: ixor
      // 1ca: if_icmpeq 1db
      // 1cd: bipush 13
      // 1cf: iload 2
      // 1d0: if_icmpeq 1db
      // 1d3: goto 1d7
      // 1d6: athrow
      // 1d7: goto 1eb
      // 1da: athrow
      // 1db: aload 0
      // 1dc: aload 0
      // 1dd: getfield e.e Ljava/lang/String;
      // 1e0: putfield e.Cb Ljava/lang/String;
      // 1e3: aload 0
      // 1e4: aload 0
      // 1e5: getfield e.x Ljava/lang/String;
      // 1e8: putfield e.Ob Ljava/lang/String;
      // 1eb: goto 221
      // 1ee: astore 2
      // 1ef: aload 2
      // 1f0: new java/lang/StringBuilder
      // 1f3: dup
      // 1f4: invokespecial java/lang/StringBuilder.<init> ()V
      // 1f7: getstatic e.Sb [Ljava/lang/String;
      // 1fa: bipush 26
      // 1fc: aaload
      // 1fd: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 200: aload 1
      // 201: ifnull 20d
      // 204: getstatic e.Sb [Ljava/lang/String;
      // 207: bipush 4
      // 208: aaload
      // 209: goto 212
      // 20c: athrow
      // 20d: getstatic e.Sb [Ljava/lang/String;
      // 210: bipush 3
      // 211: aaload
      // 212: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 215: bipush 41
      // 217: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 21a: invokevirtual java/lang/StringBuilder.toString ()Ljava/lang/String;
      // 21d: invokestatic i.a (Ljava/lang/Throwable;Ljava/lang/String;)Lla;
      // 220: athrow
      // 221: return
   }

   final void a(int var1, byte var2) {
      try {
         rb++;
         this.Ib = 1000 / var1;
         if (var2 <= 104) {
            this.Eb = 113;
         }
      } catch (RuntimeException var4) {
         throw i.a(var4, Sb[13] + var1 + ',' + var2 + ')');
      }
   }

   public final void stop() {
      try {
         pb++;
         if (~this.vb <= -1) {
            this.vb = 4000 / this.Ib;
         }
      } catch (RuntimeException var2) {
         throw i.a(var2, Sb[52]);
      }
   }

   public final boolean isDisplayable() {
      // $VF: Couldn't be decompiled
      // Please report this to the Vineflower issue tracker, at https://github.com/Vineflower/vineflower/issues with a copy of the class file (if you have the rights to distribute it!)
      // java.lang.RuntimeException: parsing failure!
      //   at org.jetbrains.java.decompiler.modules.decompiler.decompose.DomHelper.parseGraph(DomHelper.java:211)
      //   at org.jetbrains.java.decompiler.main.rels.MethodProcessor.codeToJava(MethodProcessor.java:166)
      //
      // Bytecode:
      // 00: getstatic e.W I
      // 03: bipush 1
      // 04: iadd
      // 05: putstatic e.W I
      // 08: aconst_null
      // 09: getstatic kb.a Lqb;
      // 0c: if_acmpeq 23
      // 0f: getstatic kb.a Lqb;
      // 12: invokevirtual qb.getPeer ()Ljava/awt/peer/ComponentPeer;
      // 15: ifnull 21
      // 18: goto 1c
      // 1b: athrow
      // 1c: bipush 1
      // 1d: goto 22
      // 20: athrow
      // 21: bipush 0
      // 22: ireturn
      // 23: aconst_null
      // 24: getstatic da.gb Ljava/applet/Applet;
      // 27: if_acmpeq 3f
      // 2a: aconst_null
      // 2b: getstatic da.gb Ljava/applet/Applet;
      // 2e: invokevirtual java/applet/Applet.getPeer ()Ljava/awt/peer/ComponentPeer;
      // 31: if_acmpeq 3d
      // 34: goto 38
      // 37: athrow
      // 38: bipush 1
      // 39: goto 3e
      // 3c: athrow
      // 3d: bipush 0
      // 3e: ireturn
      // 3f: aload 0
      // 40: invokespecial java/applet/Applet.getPeer ()Ljava/awt/peer/ComponentPeer;
      // 43: ifnull 4b
      // 46: bipush 1
      // 47: goto 4c
      // 4a: athrow
      // 4b: bipush 0
      // 4c: ireturn
      // 4d: astore 1
      // 4e: aload 1
      // 4f: getstatic e.Sb [Ljava/lang/String;
      // 52: bipush 47
      // 54: aaload
      // 55: invokestatic i.a (Ljava/lang/Throwable;Ljava/lang/String;)Lla;
      // 58: athrow
   }

   public final void update(Graphics var1) {
      try {
         ub++;
         this.paint(var1);
      } catch (RuntimeException var4) {
         RuntimeException var2 = var4;

         RuntimeException var10000;
         StringBuilder var10001;
         try {
            var10000 = var2;
            var10001 = new StringBuilder().append(Sb[40]);
            if (var1 != null) {
               throw i.a(var2, var10001.append(Sb[4]).append(')').toString());
            }
         } catch (RuntimeException var3) {
            throw var3;
         }

         throw i.a(var10000, var10001.append(Sb[3]).append(')').toString());
      }
   }

   void a(int var1, Runnable var2) {
      try {
         A++;
         Thread var6 = new Thread(var2);
         if (var1 == 1) {
            var6.setDaemon(true);
            var6.start();
         }
      } catch (RuntimeException var5) {
         RuntimeException var3 = var5;

         RuntimeException var10000;
         StringBuilder var10001;
         try {
            var10000 = var3;
            var10001 = new StringBuilder().append(Sb[5]).append(var1).append(',');
            if (var2 != null) {
               throw i.a(var3, var10001.append(Sb[4]).append(')').toString());
            }
         } catch (RuntimeException var4) {
            throw var4;
         }

         throw i.a(var10000, var10001.append(Sb[3]).append(')').toString());
      }
   }

   protected e() {
      this.F = new long[10];
      this.n = 1;
      this.hb = false;
      this.B = Sb[44];
      this.S = 1000;
      this.z = null;
      this.vb = 0;
      this.a = 384;
      this.Ib = 20;
      this.b = 0;
      this.m = 512;
      this.N = false;
      this.sb = 0;
      this.V = 0;
      this.tb = new Font(Sb[43], 0, 15);
      this.X = new Font(Sb[42], 1, 13);
      this.Jb = new Font(Sb[42], 0, 12);
      this.x = "";
      this.U = false;
      this.I = 0;
      this.e = "";
      this.bb = false;
      this.Q = 1;
      this.Qb = 0;
      this.Z = false;
      this.Bb = 0;
      this.xb = 0;
      this.gb = false;
      this.Kb = false;
      this.Cb = "";
      this.E = false;
      this.Ob = "";
   }

   private static char[] z(String var0) {
      char[] var10000 = var0.toCharArray();
      if (var10000.length < 2) {
         var10000[0] = (char)(var10000[0] ^ 'h');
      }

      return var10000;
   }

   private static String z(char[] var0) {
      int var10002 = var0.length;
      char[] var10001 = var0;
      int var10000 = var10002;

      for (int var1 = 0; var10000 > var1; var1++) {
         char var10004 = var10001[var1];
         byte var10005;
         switch (var1 % 5) {
            case 0:
               var10005 = 24;
               break;
            case 1:
               var10005 = 3;
               break;
            case 2:
               var10005 = 72;
               break;
            case 3:
               var10005 = 8;
               break;
            default:
               var10005 = 104;
         }

         var10001[var1] = (char)(var10004 ^ var10005);
      }

      return new String(var10001).intern();
   }
}
