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
            return da.gb == null ? super.getParameter(var1) : da.gb.getParameter(var1);
         }
      } catch (RuntimeException var3) {
         throw i.a(var3, Sb[21] + (var1 != null ? Sb[4] : Sb[3]) + ')');
      }
   }

   private final void b(int var1) {
      try {
         fb++;
         this.vb = -2;
         System.out.println(Sb[28]);
         this.a(false);
         mb.a(11200, 1000L);
         if (var1 != 100) {
            this.e(27);
         }

         if (kb.a != null) {
            kb.a.dispose();
            System.exit(0);
         }
      } catch (RuntimeException var3) {
         throw i.a(var3, Sb[27] + var1 + ')');
      }
   }

   @Override
   public final void run() {
      boolean var10 = client.vh;

      try {
         qb++;

         try {
            if (1 == this.n) {
               this.n = 2;

               int var10000;
               while (true) {
                  if (!this.isDisplayable()) {
                     var10000 = this.vb;
                     if (var10) {
                        break;
                     }

                     if (this.vb >= 0) {
                        if (-1 > ~this.vb) {
                           this.vb--;
                           if (-1 == ~this.vb) {
                              this.b(100);
                              this.z = null;
                              return;
                           }
                        }

                        mb.a(11200, (long)this.Ib);
                        if (!var10) {
                           continue;
                        }
                     }
                  }

                  var10000 = 0;
                  break;
               }

               if (var10000 > this.vb) {
                  if (this.vb == -1) {
                     this.b(100);
                  }

                  this.z = null;
                  return;
               }

               if (!this.b((byte)118)) {
                  if (1 != ~this.vb) {
                     this.b(100);
                  }

                  this.z = null;
                  return;
               }

               this.a((byte)-92);
               this.n = 0;
            }

            label235: {
               if (null != kb.a) {
                  kb.a.addMouseListener(this);
                  kb.a.addMouseMotionListener(this);
                  kb.a.addKeyListener(this);
                  if (!var10) {
                     break label235;
                  }
               }

               if (null == da.gb) {
                  this.addMouseListener(this);
                  this.addMouseMotionListener(this);
                  this.addKeyListener(this);
                  if (!var10) {
                     break label235;
                  }
               }

               da.gb.addMouseListener(this);
               da.gb.addMouseMotionListener(this);
               da.gb.addKeyListener(this);
            }

            int var3 = 0;
            int var4 = 256;
            int var5 = 1;
            int var6 = 0;
            int var7 = 0;

            while (var7 < 10) {
               this.F[var7] = p.a(0);
               var7++;
               if (var10) {
                  return;
               }

               if (var10) {
                  break;
               }
            }

            long var1 = p.a(0);

            int var16;
            int var10001;
            label191: {
               while (0 <= this.vb) {
                  var16 = -1;
                  var10001 = ~this.vb;
                  if (var10) {
                     break label191;
                  }

                  if (-1 > var10001) {
                     this.vb--;
                     if (0 == this.vb) {
                        this.b(100);
                        this.z = null;
                        return;
                     }
                  }

                  label187: {
                     var7 = var4;
                     var4 = 300;
                     int var8 = var5;
                     var5 = 1;
                     var1 = p.a(0);
                     if (~this.F[var3] == -1L) {
                        var5 = var8;
                        var4 = var7;
                        if (!var10) {
                           break label187;
                        }
                     }

                     if (~var1 < ~this.F[var3]) {
                        var4 = (int)(this.Ib * 2560 / (var1 + -this.F[var3]));
                     }
                  }

                  if (~var4 > -26) {
                     var4 = 25;
                  }

                  if (-257 > ~var4) {
                     var4 = 256;
                     var5 = (int)(-((-this.F[var3] + var1) / 10L) + this.Ib);
                     if (var5 < this.Q) {
                        var5 = this.Q;
                     }
                  }

                  label179: {
                     mb.a(11200, (long)var5);
                     this.F[var3] = var1;
                     if (-2 > ~var5) {
                        int var9 = 0;

                        while (10 > var9) {
                           long var20;
                           var16 = (var20 = -1L - ~this.F[var9]) == 0L ? 0 : (var20 < 0L ? -1 : 1);
                           if (var10) {
                              break label179;
                           }

                           if (var16 != 0) {
                              this.F[var9] = this.F[var9] + var5;
                           }

                           var9++;
                           if (var10) {
                              break;
                           }
                        }
                     }

                     var3 = (1 + var3) % 10;
                     var16 = 0;
                  }

                  int var15 = var16;

                  while (true) {
                     if (-257 < ~var6) {
                        this.e(119);
                        var6 += var4;
                        if (~this.S <= ~(++var15)) {
                           continue;
                        }

                        var6 = 0;
                        this.b += 6;
                        var16 = 25;
                        var10001 = this.b;
                        if (var10) {
                           break;
                        }

                        if (25 < this.b) {
                           this.b = 0;
                           this.U = true;
                        }
                     }

                     this.b--;
                     var16 = var6;
                     var10001 = 255;
                     break;
                  }

                  var6 = var16 & var10001;
                  this.b(false);
                  if (var10) {
                     break;
                  }
               }

               var16 = -1;
               var10001 = this.vb;
            }

            if (var16 == var10001) {
               this.b(100);
            }

            this.z = null;
         } catch (Exception var11) {
            mb.a(2097151, var11, null);
            this.a(Sb[12], true);
         }
      } catch (RuntimeException var12) {
         throw i.a(var12, Sb[11]);
      }
   }

   @Override
   public final synchronized void keyReleased(KeyEvent var1) {
      try {
         this.a(var1, (byte)-128);
         w++;
         char var2 = var1.getKeyChar();
         int var3 = var1.getKeyCode();
         if (~((char)var2) != -33) {
         }

         if (40 == var3) {
         }

         if ('N' != (char)var2 && 'M' != (char)var2) {
         }

         if (~var3 == -40) {
            this.E = false;
         }

         if ((char)var2 != 'n' && 'm' != (char)var2) {
         }

         if ((char)var2 != '{') {
         }

         if (var3 == 37) {
            this.Z = false;
         }

         if (38 == var3) {
         }

         if (-126 == ~((char)var2)) {
         }
      } catch (RuntimeException var4) {
         throw i.a(var4, Sb[58] + (var1 != null ? Sb[4] : Sb[3]) + ')');
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

   public final void paint(Graphics var1) {
      try {
         this.N = true;
         l++;
         if (-3 != ~this.n || this.C == null) {
            if (0 != this.n) {
               return;
            }

            this.a(-89);
            if (!client.vh) {
               return;
            }
         }

         this.a(this.B, this.V, 126);
      } catch (RuntimeException var3) {
         throw i.a(var3, Sb[10] + (var1 != null ? Sb[4] : Sb[3]) + ')');
      }
   }

   private final void a(String var1, int var2, int var3) {
      try {
         try {
            int var4 = (-281 + this.m) / 2;
            int var5 = (-148 + this.a) / 2;
            this.u.setColor(Color.black);
            this.u.fillRect(0, 0, this.m, this.a);
            if (!this.hb) {
               this.u.drawImage(this.C, var4, var5, this);
            }

            var4 += 2;
            this.V = var2;
            var5 += 90;
            this.B = var1;
            if (var3 <= 97) {
               this.mouseReleased((MouseEvent)null);
            }

            this.u.setColor(new Color(132, 132, 132));
            if (this.hb) {
               this.u.setColor(new Color(220, 0, 0));
            }

            this.u.drawRect(var4 + -2, var5 + -2, 280, 23);
            this.u.fillRect(var4, var5, 277 * var2 / 100, 20);
            this.u.setColor(new Color(198, 198, 198));
            if (this.hb) {
               this.u.setColor(new Color(255, 255, 255));
            }

            label48: {
               this.a(this.tb, var1, 10 + var5, true, 138 + var4, this.u);
               if (!this.hb) {
                  this.a(this.X, Sb[24], 30 + var5, true, var4 - -138, this.u);
                  this.a(this.X, Sb[23], var5 + 44, true, var4 - -138, this.u);
                  if (!client.vh) {
                     break label48;
                  }
               }

               this.u.setColor(new Color(132, 132, 152));
               this.a(this.Jb, Sb[23], -20 + this.a, true, 138 + var4, this.u);
            }

            if (null != this.p) {
               this.u.setColor(Color.white);
               this.a(this.X, this.p, -120 + var5, true, var4 + 138, this.u);
            }
         } catch (Exception var6) {
         }

         r++;
      } catch (RuntimeException var7) {
         throw i.a(var7, Sb[22] + (var1 != null ? Sb[4] : Sb[3]) + ',' + var2 + ',' + var3 + ')');
      }
   }

   @Override
   public final synchronized void mouseReleased(MouseEvent var1) {
      try {
         g++;
         this.a(var1, (byte)-128);
         this.I = var1.getX() + -this.Eb;
         this.xb = var1.getY() + -this.K;
         this.Bb = 0;
      } catch (RuntimeException var3) {
         throw i.a(var3, Sb[19] + (var1 != null ? Sb[4] : Sb[3]) + ')');
      }
   }

   final void a(int var1, byte var2, String var3) {
      try {
         Lb++;

         try {
            int var4 = (-281 + this.m) / 2;
            var4 += 2;
            int var5 = (this.a - 148) / 2;
            this.B = var3;
            this.V = var1;
            var5 += 90;
            int var6 = 277 * var1 / 100;
            this.u.setColor(new Color(132, 132, 132));
            if (this.hb) {
               this.u.setColor(new Color(220, 0, 0));
            }

            this.u.fillRect(var4, var5, var6, 20);
            this.u.setColor(Color.black);
            this.u.fillRect(var6 + var4, var5, -var6 + 277, 20);
            this.u.setColor(new Color(198, 198, 198));
            if (this.hb) {
               this.u.setColor(new Color(255, 255, 255));
            }

            this.a(this.tb, var3, 10 + var5, true, 138 + var4, this.u);
         } catch (Exception var7) {
         }

         if (var2 > -96) {
            this.x = (String)null;
         }
      } catch (RuntimeException var8) {
         throw i.a(var8, Sb[57] + var1 + ',' + var2 + ',' + (var3 != null ? Sb[4] : Sb[3]) + ')');
      }
   }

   private final Image a(byte[] var1, byte var2) {
      try {
         if (var2 != -54) {
            this.Ob = (String)null;
         }

         cb++;
         return pa.a(79, this, var1);
      } catch (RuntimeException var4) {
         throw i.a(var4, Sb[68] + (var1 != null ? Sb[4] : Sb[3]) + ',' + var2 + ')');
      }
   }

   public static final void provideLoaderApplet(Applet var0) {
      try {
         da.gb = var0;
         t++;
      } catch (RuntimeException var2) {
         throw i.a(var2, Sb[54] + (var0 != null ? Sb[4] : Sb[3]) + ')');
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
      } catch (RuntimeException var14) {
         throw i.a(
            var14,
            Sb[18]
               + var1
               + ','
               + (var2 != null ? Sb[4] : Sb[3])
               + ','
               + var3
               + ','
               + (var4 != null ? Sb[4] : Sb[3])
               + ','
               + var5
               + ','
               + var6
               + ','
               + var7
               + ','
               + var8
               + ','
               + var9
               + ')'
         );
      }
   }

   private final boolean b(byte var1) {
      try {
         O++;
         byte[] var2 = this.a(Sb[32], 0, 3, 85);
         if (var2 != null) {
            if (var1 != 118) {
               this.B = (String)null;
            }

            byte[] var3 = na.a(Sb[30], 0, var2, -120);
            this.C = this.a(var3, (byte)-54);
            if (qa.a(this, Sb[29], 0, var1 + -118)) {
               if (!qa.a(this, Sb[39], 1, 0)) {
                  return false;
               }

               if (!qa.a(this, Sb[31], 2, 0)) {
                  return false;
               }

               if (!qa.a(this, Sb[36], 3, 0)) {
                  return false;
               }

               if (qa.a(this, Sb[38], 4, var1 + -118)) {
                  if (qa.a(this, Sb[34], 5, 0)) {
                     return !qa.a(this, Sb[35], 6, var1 ^ 118) ? false : qa.a(this, Sb[37], 7, var1 + -118);
                  } else {
                     return false;
                  }
               } else {
                  return false;
               }
            } else {
               return false;
            }
         } else {
            return false;
         }
      } catch (RuntimeException var4) {
         throw i.a(var4, Sb[33] + var1 + ')');
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

   private final void a(String var1, boolean var2) {
      try {
         P++;
         if (!this.Kb) {
            this.Kb = var2;
            System.out.println(Sb[63] + var1);

            try {
               label38: {
                  if (null != da.gb) {
                     a.a(Sb[59], (byte)82, da.gb);
                     if (!client.vh) {
                        break label38;
                     }
                  }

                  a.a(Sb[59], (byte)-73, this);
               }
            } catch (Throwable var5) {
            }

            try {
               this.getAppletContext().showDocument(new URL(this.getCodeBase(), Sb[63] + var1 + Sb[62]), Sb[60]);
            } catch (Exception var4) {
            }
         }
      } catch (RuntimeException var6) {
         throw i.a(var6, Sb[61] + (var1 != null ? Sb[4] : Sb[3]) + ',' + var2 + ')');
      }
   }

   @Override
   public final synchronized void mouseDragged(MouseEvent var1) {
      try {
         this.a(var1, (byte)-128);
         q++;
         this.I = var1.getX() - this.Eb;
         this.xb = var1.getY() - this.K;
         if (var1.isMetaDown()) {
            this.Bb = 2;
            if (!client.vh) {
               return;
            }
         }

         this.Bb = 1;
      } catch (RuntimeException var3) {
         throw i.a(var3, Sb[76] + (var1 != null ? Sb[4] : Sb[3]) + ')');
      }
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
               pa.b = pa.k = new c(var3, null, 0, null != da.gb);
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
         } catch (Exception var12) {
            mb.a(var4 ^ 2097149, var12, null);
            this.a(Sb[12], true);
         }
      } catch (RuntimeException var13) {
         throw i.a(var13, Sb[70] + var1 + ',' + var2 + ',' + var3 + ',' + var4 + ',' + var5 + ')');
      }
   }

   @Override
   public final void mouseEntered(MouseEvent var1) {
      try {
         this.a(var1, (byte)-128);
         Rb++;
      } catch (RuntimeException var3) {
         throw i.a(var3, Sb[20] + (var1 != null ? Sb[4] : Sb[3]) + ')');
      }
   }

   final boolean d(int var1) {
      try {
         ab++;
         Graphics var2 = this.getGraphics();
         if (var2 != null) {
            if (var1 != 2) {
               this.sb = -7;
            }

            this.u = var2.create();
            this.u.translate(this.Eb, this.K);
            this.u.setColor(Color.black);
            this.u.fillRect(0, 0, this.m, this.a);
            this.a(Sb[56], 0, var1 ^ 103);
            return true;
         } else {
            return false;
         }
      } catch (RuntimeException var3) {
         throw i.a(var3, Sb[55] + var1 + ')');
      }
   }

   private final void a(InputEvent var1, byte var2) {
      try {
         if (var2 <= -127) {
            d++;
            int var3 = var1.getModifiers();
            this.bb = 0 != (var3 & 2);
            this.gb = (var3 & 1) != 0;
         }
      } catch (RuntimeException var4) {
         throw i.a(var4, Sb[74] + (var1 != null ? Sb[4] : Sb[3]) + ',' + var2 + ')');
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
      } catch (RuntimeException var3) {
         throw i.a(var3, Sb[7] + (var1 != null ? Sb[4] : Sb[3]) + ')');
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
         Object var7;
         label41: {
            D++;
            if (null == kb.a) {
               var7 = this;
               if (!client.vh) {
                  break label41;
               }
            }

            var7 = kb.a;
         }

         FontMetrics var8 = var7.getFontMetrics(var1);
         var8.stringWidth(var2);
         if (!var4) {
            this.c(68);
         }

         var6.setFont(var1);
         var6.drawString(var2, var5 + -(var8.stringWidth(var2) / 2), var3 - -(var8.getHeight() / 4));
      } catch (RuntimeException var9) {
         throw i.a(
            var9,
            Sb[2]
               + (var1 != null ? Sb[4] : Sb[3])
               + ','
               + (var2 != null ? Sb[4] : Sb[3])
               + ','
               + var3
               + ','
               + var4
               + ','
               + var5
               + ','
               + (var6 != null ? Sb[4] : Sb[3])
               + ')'
         );
      }
   }

   public final void destroy() {
      try {
         this.vb = -1;
         Hb++;
         mb.a(11200, 5000L);
         if (-1 == this.vb) {
            System.out.println(Sb[0]);
            this.b(100);
            if (this.z != null) {
               this.z.stop();
               this.z = null;
            }
         }
      } catch (RuntimeException var2) {
         throw i.a(var2, Sb[1]);
      }
   }

   void a(byte var1, int var2) {
      try {
         if (var1 <= 105) {
            this.a((InputEvent)null, (byte)83);
         }

         c++;
      } catch (RuntimeException var4) {
         throw i.a(var4, Sb[46] + var1 + ',' + var2 + ')');
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
      } catch (RuntimeException var3) {
         throw i.a(var3, Sb[45] + (var1 != null ? Sb[4] : Sb[3]) + ')');
      }
   }

   final void c(int var1) {
      boolean var3 = client.vh;

      try {
         if (var1 == -28492) {
            int var2 = 0;

            while (true) {
               if (var2 < 10) {
                  this.F[var2] = 0L;
                  var2++;
                  if (var3) {
                     break;
                  }

                  if (!var3) {
                     continue;
                  }
               }

               R++;
               break;
            }
         }
      } catch (RuntimeException var4) {
         throw i.a(var4, Sb[53] + var1 + ')');
      }
   }

   final byte[] a(String var1, int var2, int var3, int var4) {
      try {
         J++;
         if (var4 <= 53) {
            this.c(15);
         }

         try {
            return ib.a(-101, var1, var2, var3);
         } catch (IOException var6) {
            mb.a(2097151, var6, Sb[65] + var3);
            return null;
         }
      } catch (RuntimeException var7) {
         throw i.a(var7, Sb[66] + (var1 != null ? Sb[4] : Sb[3]) + ',' + var2 + ',' + var3 + ',' + var4 + ')');
      }
   }

   @Override
   public final void keyTyped(KeyEvent var1) {
      try {
         Nb++;
         this.a(var1, (byte)-128);
      } catch (RuntimeException var3) {
         throw i.a(var3, Sb[6] + (var1 != null ? Sb[4] : Sb[3]) + ')');
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
      } catch (RuntimeException var3) {
         throw i.a(var3, Sb[64] + (var1 != null ? Sb[4] : Sb[3]) + ')');
      }
   }

   @Override
   public final synchronized void mousePressed(MouseEvent var1) {
      try {
         label27: {
            this.a(var1, (byte)-128);
            Pb++;
            this.I = var1.getX() - this.Eb;
            this.xb = var1.getY() + -this.K;
            if (!var1.isMetaDown()) {
               this.Bb = 1;
               if (!client.vh) {
                  break label27;
               }
            }

            this.Bb = 2;
         }

         this.Qb = this.Bb;
         this.sb = 0;
         this.a(this.I, 94, this.Bb, this.xb);
      } catch (RuntimeException var3) {
         throw i.a(var3, Sb[9] + (var1 != null ? Sb[4] : Sb[3]) + ')');
      }
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
   public final synchronized void keyPressed(KeyEvent var1) {
      boolean var6 = client.vh;

      try {
         L++;
         this.a(var1, (byte)-128);
         char var2 = var1.getKeyChar();
         int var3 = var1.getKeyCode();
         this.a((byte)126, var2);
         if (-113 == ~var3) {
            this.U = !this.U;
         }

         if (-79 != ~((char)var2) && 'M' != (char)var2) {
         }

         if (~((char)var2) == -33) {
         }

         if (-124 != ~((char)var2)) {
         }

         this.sb = 0;
         if (~((char)var2) != -111 && ~((char)var2) != -110) {
         }

         if (-41 != ~var3) {
         }

         if (~var3 == -40) {
            this.E = true;
         }

         if ((char)var2 == '}') {
         }

         if (~var3 == -39) {
         }

         if (~var3 == -38) {
            this.Z = true;
         }

         boolean var4 = false;
         int var5 = 0;

         int var10000;
         int var10001;
         while (true) {
            label138:
            if (var5 < i.f.length()) {
               var10000 = ~var2;
               var10001 = ~i.f.charAt(var5);
               if (var6) {
                  break;
               }

               if (var10000 == var10001) {
                  var4 = true;
                  if (!var6) {
                     break label138;
                  }
               }

               var5++;
               if (!var6) {
                  continue;
               }
            }

            if (var4 && ~this.e.length() > -21) {
               this.e = this.e + (char)var2;
            }

            if (var4 && 80 > this.x.length()) {
               this.x = this.x + (char)var2;
            }

            var10000 = 8;
            var10001 = var2;
            break;
         }

         if (var10000 == var10001 && -1 > ~this.e.length()) {
            this.e = this.e.substring(0, this.e.length() - 1);
         }

         if ('\b' == var2 && ~this.x.length() < -1) {
            this.x = this.x.substring(0, -1 + this.x.length());
         }

         if (-11 == ~var2 || '\r' == var2) {
            this.Cb = this.e;
            this.Ob = this.x;
         }
      } catch (RuntimeException var7) {
         throw i.a(var7, Sb[26] + (var1 != null ? Sb[4] : Sb[3]) + ')');
      }
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
      try {
         W++;
         if (null != kb.a) {
            return kb.a.getPeer() != null;
         } else {
            return null != da.gb ? null != da.gb.getPeer() : super.getPeer() != null;
         }
      } catch (RuntimeException var2) {
         throw i.a(var2, Sb[47]);
      }
   }

   public final void update(Graphics var1) {
      try {
         ub++;
         this.paint(var1);
      } catch (RuntimeException var3) {
         throw i.a(var3, Sb[40] + (var1 != null ? Sb[4] : Sb[3]) + ')');
      }
   }

   void a(int var1, Runnable var2) {
      try {
         A++;
         Thread var3 = new Thread(var2);
         if (var1 == 1) {
            var3.setDaemon(true);
            var3.start();
         }
      } catch (RuntimeException var4) {
         throw i.a(var4, Sb[5] + var1 + ',' + (var2 != null ? Sb[4] : Sb[3]) + ')');
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
         var10000[0] = (char)(var10000[0] ^ 104);
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
