import java.awt.Component;
import java.awt.EventQueue;
import java.awt.Frame;
import java.awt.Point;
import java.awt.Toolkit;
import java.awt.datatransfer.Clipboard;
import java.awt.datatransfer.Transferable;
import java.io.DataInputStream;
import java.io.File;
import java.io.IOException;
import java.lang.reflect.Method;
import java.net.InetAddress;
import java.net.Socket;
import java.net.URL;

final class c implements Runnable {
   private boolean e;
   private q t;
   private Thread r;
   private Object x;
   private static String m;
   private static String p;
   private boolean j = false;
   private Object i;
   EventQueue n;
   static Method u;
   private static int b;
   private g a = null;
   private boolean c;
   private wa w;
   private static String g;
   static Method y;
   d f = null;
   d s = null;
   private static String h;
   d v;
   static String q;
   private d[] l;
   private static volatile long d = 0L;
   private g o;
   static String k;
   private static final String[] z = new String[]{
      z(z("lHu\u0004:Y")),
      z(z("e\u0013=\u0016+hM}\u00161)")),
      z(z(")[a\u0002#eAwN")),
      z(z("e\u0013=")),
      z(z("(Ms\u0015")),
      z(z("e\u0013=\u00131eHq\t')")),
      z(z("YY`\u0004$c[w\u000f!cZ")),
      z(z("Y^{\u0011lbHf")),
      z(z("t^")),
      z(z("e\u0013=\u0016+hGfN")),
      z(z(")]\u007f\u0011m")),
      z(z("Y[qO&g]")),
      z(z("LHu\u0004:&og\r.&zq\u0013'cG")),
      z(z("cQ{\u0015")),
      z(z("gKq\u0005'`Nz\b(mE\u007f\u000f-vX`\u00126s_e\u0019;|hP\"\u0006CoU)\u000bLb^,\fIyC3\u0011R|D6\u001a_s\"Pp5\u001d'Wu>\u0010-G\u007f*\u00077JoY\n(Nh")),
      z(z("nH")),
      z(z("n]f\u00111<\u0006=")),
      z(z("cGf\u00040")),
      z(z("n]f\u0011x)\u0006")),
      z(z("kFd\u0004/i\\a\u0004")),
      z(z("uLf\u00027u]}\f!s[a\u000e0")),
      z(z("eDvAme\ta\u0015#t]2C($\t0")),
      z(z("j@a\u0015/iMw\u0012")),
      z(z("q@|")),
      z(z("uA}\u0016!s[a\u000e0")),
      z(z("uLf'-e\\a50g_w\u00131gEY\u0004;ul|\u0000 jLv")),
      z(z("kH{\u000f\u001d`@~\u0004\u001deHq\t'(@v\u0019p3\u001c")),
      z(z("tK")),
      z(z("lHd\u0000lg^fO\u0001iGf\u0000+hL`")),
      z(z("iZ<\u000f#kL")),
      z(z("iZ<\u0017'tZ{\u000e,")),
      z(z("sZw\u0013lnF\u007f\u0004")),
      z(z("lHd\u0000lpL|\u0005-t")),
      z(z("kH{\u000f\u001d`@~\u0004\u001deHq\t'(Ms\u0015p")),
      z(z("iZ<\u00000eA")),
      z(z("tH|\u0005-k\u0007v\u00006")),
      z(z("lHd\u0000lpL`\u0012+iG")),
      z(z("x\u0006")),
      z(z("uLf'-e\\a\";eEw3-i]")),
      z(z("SGy\u000f-qG")),
      z(z("lHd\u0000lg^fO\u0001iDb\u000e,cGf")),
      z(z("kH{\u000f\u001d`@~\u0004\u001deHq\t'(@v\u0019")),
      z(z("7\u0007#")),
      z(z("k@q\u0013-uFt\u0015"))
   };

   final g a(String var1, int var2, int var3) {
      try {
         return var3 > -66 ? (g)null : this.a(var2, (byte)81, var1, false);
      } catch (RuntimeException var5) {
         throw var5;
      }
   }

   final g a(boolean var1, Runnable var2, int var3) {
      try {
         if (!var1) {
            this.a(-34, 71, (byte)60, 103, (Object)null);
         }

         return this.a(2, 0, (byte)-21, var3, var2);
      } catch (RuntimeException var5) {
         throw var5;
      }
   }

   private static final d a(int var0, String var1, boolean var2, String var3) {
      try {
         String var4;
         if (-34 != ~var0) {
            if (~var0 == -35) {
               var4 = z[0] + var1 + z[6] + var3 + z[7];
            } else {
               var4 = z[0] + var1 + z[6] + var3 + z[4];
            }
         } else {
            var4 = z[0] + var1 + z[6] + var3 + z[11];
         }

         if (var2) {
            return (d)null;
         }

         String[] var5 = new String[]{z[5], z[2], m, z[1], z[9], z[3], z[10], ""};

         for (int var6 = 0; var6 < var5.length; var6++) {
            String var7 = var5[var6];
            if (0 >= var7.length() || new File(var7).exists()) {
               try {
                  return new d(new File(var7, var4), z[8], 10000L);
               } catch (Exception var9) {
               }
            }
         }

         return null;
      } catch (RuntimeException var10) {
         throw var10;
      }
   }

   @Override
   public final void run() {
      try {
         while (true) {
            g var1;
            synchronized (this) {
               while (true) {
                  if (this.e) {
                     return;
                  }

                  if (null != this.a) {
                     var1 = this.a;
                     this.a = this.a.a;
                     if (null == this.a) {
                        this.o = null;
                     }
                     break;
                  }

                  try {
                     this.wait();
                  } catch (InterruptedException var9) {
                  }
               }
            }

            try {
               int var2 = var1.g;
               if (~var2 == -2) {
                  if (p.a(0) < d) {
                     throw new IOException();
                  }

                  var1.d = new Socket(InetAddress.getByName((String)var1.f), var1.e);
               } else if (var2 != 22) {
                  if (2 == var2) {
                     Thread var26 = new Thread((Runnable)var1.f);
                     var26.setDaemon(true);
                     var26.start();
                     var26.setPriority(var1.e);
                     var1.d = var26;
                  } else if (var2 != 4) {
                     if (8 == var2) {
                        Object[] var25 = (Object[])var1.f;
                        if (this.j && null == ((Class)var25[0]).getClassLoader()) {
                           throw new SecurityException();
                        }

                        var1.d = ((Class)var25[0]).getDeclaredMethod((String)var25[1], (Class<?>[])var25[2]);
                     } else if (-10 == ~var2) {
                        Object[] var24 = (Object[])var1.f;
                        if (this.j && ((Class)var24[0]).getClassLoader() == null) {
                           throw new SecurityException();
                        }

                        var1.d = ((Class)var24[0]).getDeclaredField((String)var24[1]);
                     } else if (var2 == 18) {
                        Clipboard var23 = Toolkit.getDefaultToolkit().getSystemClipboard();
                        var1.d = var23.getContents(null);
                     } else if (19 != var2) {
                        if (!this.j) {
                           throw new Exception("");
                        }

                        if (3 == var2) {
                           if (d > p.a(0)) {
                              throw new IOException();
                           }

                           String var22 = (0xFF & var1.e >> -182496008)
                              + "."
                              + (0xFF & var1.e >> 954736400)
                              + "."
                              + ((65472 & var1.e) >> -58046680)
                              + "."
                              + (0xFF & var1.e);
                           var1.d = InetAddress.getByName(var22).getHostName();
                        } else if (var2 == 21) {
                           if (p.a(0) < d) {
                              throw new IOException();
                           }

                           var1.d = InetAddress.getByName((String)var1.f).getAddress();
                        } else if (5 != var2) {
                           if (~var2 == -7) {
                              Frame var21 = new Frame(z[12]);
                              var1.d = var21;
                              var21.setResizable(false);
                              if (this.c) {
                                 this.w.a(var21, var1.c >> -747878896, var1.c & 65535, var1.e & 65535, var1.e >>> 831913136, (byte)77);
                              } else {
                                 Class.forName(z[15])
                                    .getMethod(z[17], Frame.class, int.class, int.class, int.class, int.class)
                                    .invoke(
                                       this.i,
                                       var21,
                                       new Integer(var1.e >>> -1397573296),
                                       new Integer(65535 & var1.e),
                                       new Integer(var1.c >> -1159913680),
                                       new Integer(var1.c & 65535)
                                    );
                              }
                           } else if (var2 == 7) {
                              if (!this.c) {
                                 Class.forName(z[15]).getMethod(z[13]).invoke(this.i);
                              } else {
                                 this.w.a((Frame)var1.f, 0);
                              }
                           } else if (-13 == ~var2) {
                              d var20 = a(b, p, false, (String)var1.f);
                              var1.d = var20;
                           } else if (-14 == ~var2) {
                              d var19 = a(b, "", false, (String)var1.f);
                              var1.d = var19;
                           } else if (this.j && var2 == 14) {
                              int var18 = var1.e;
                              int var29 = var1.c;
                              if (this.c) {
                                 this.t.a(23529, var29, var18);
                              } else {
                                 Class.forName("j").getDeclaredMethod(z[19], int.class, int.class).invoke(this.x, new Integer(var18), new Integer(var29));
                              }
                           } else if (this.j && ~var2 == -16) {
                              boolean var17 = ~var1.e != -1;
                              Component var28 = (Component)var1.f;
                              if (this.c) {
                                 this.t.a(-4, var28, var17);
                              } else {
                                 Class.forName("j").getDeclaredMethod(z[24], Component.class, boolean.class).invoke(this.x, var28, new Boolean(var17));
                              }
                           } else if (!this.c && var2 == 17) {
                              Object[] var16 = (Object[])var1.f;
                              Class.forName("j")
                                 .getDeclaredMethod(z[20], Component.class, int[].class, int.class, int.class, Point.class)
                                 .invoke(this.x, var16[0], var16[1], new Integer(var1.e), new Integer(var1.c), var16[2]);
                           } else {
                              if (16 != var2) {
                                 throw new Exception("");
                              }

                              try {
                                 if (!g.startsWith(z[23])) {
                                    throw new Exception();
                                 }

                                 String var15 = (String)var1.f;
                                 if (!var15.startsWith(z[18]) && !var15.startsWith(z[16])) {
                                    throw new Exception();
                                 }

                                 String var27 = z[14];

                                 for (int var5 = 0; var5 < var15.length(); var5++) {
                                    if (-1 == var27.indexOf(var15.charAt(var5))) {
                                       throw new Exception();
                                    }
                                 }

                                 Runtime.getRuntime().exec(z[21] + var15 + "\"");
                                 var1.d = null;
                              } catch (Exception var10) {
                                 var1.d = var10;
                                 throw var10;
                              }
                           }
                        } else if (this.c) {
                           var1.d = this.w.a((byte)-100);
                        } else {
                           var1.d = Class.forName(z[15]).getMethod(z[22]).invoke(this.i);
                        }
                     } else {
                        Transferable var3 = (Transferable)var1.f;
                        Clipboard var4 = Toolkit.getDefaultToolkit().getSystemClipboard();
                        var4.setContents(var3, null);
                     }
                  } else {
                     if (~p.a(0) > ~d) {
                        throw new IOException();
                     }

                     var1.d = new DataInputStream(((URL)var1.f).openStream());
                  }
               } else {
                  if (~d < ~p.a(0)) {
                     throw new IOException();
                  }

                  try {
                     var1.d = na.a(4718, var1.e, (String)var1.f).a((byte)50);
                  } catch (fa var8) {
                     var1.d = var8.getMessage();
                     throw var8;
                  }
               }

               var1.b = 1;
            } catch (ThreadDeath var11) {
               throw var11;
            } catch (Throwable var12) {
               var1.b = 2;
            }

            synchronized (var1) {
               var1.notify();
            }
         }
      } catch (RuntimeException var14) {
         throw var14;
      }
   }

   private final g a(int var1, byte var2, String var3, boolean var4) {
      try {
         if (var2 != 81) {
            this.a(3, (byte)-100, (String)null, true);
         }

         return this.a(var4 ? 22 : 1, 0, (byte)-21, var1, var3);
      } catch (RuntimeException var6) {
         throw var6;
      }
   }

   final g a(byte var1, URL var2) {
      try {
         if (var1 != 74) {
            d = -110L;
         }

         return this.a(4, 0, (byte)-21, 0, var2);
      } catch (RuntimeException var4) {
         throw var4;
      }
   }

   final g a(int var1, byte var2) {
      try {
         int var3 = 9 / ((-58 - var2) / 56);
         return this.a(3, 0, (byte)-21, var1, null);
      } catch (RuntimeException var4) {
         throw var4;
      }
   }

   private final g a(int var1, int var2, byte var3, int var4, Object var5) {
      try {
         g var6 = new g();
         var6.e = var4;
         var6.g = var1;
         var6.c = var2;
         var6.f = var5;
         synchronized (this) {
            if (null == this.o) {
               this.o = this.a = var6;
            } else {
               this.o.a = var6;
               this.o = var6;
            }

            this.notify();
         }

         if (var3 != -21) {
            k = (String)null;
         }

         return var6;
      } catch (RuntimeException var10) {
         throw var10;
      }
   }

   c(int var1, String var2, int var3, boolean var4) throws Exception {
      this.e = false;
      this.c = false;
      this.v = null;
      this.o = null;

      try {
         k = z[42];
         b = var1;
         q = z[39];
         p = var2;
         this.j = var4;

         try {
            q = System.getProperty(z[32]);
            k = System.getProperty(z[36]);
         } catch (Exception var16) {
         }

         if (~q.toLowerCase().indexOf(z[43]) != 0) {
            this.c = true;
         }

         try {
            h = System.getProperty(z[29]);
         } catch (Exception var15) {
            h = z[39];
         }

         g = h.toLowerCase();

         try {
            System.getProperty(z[34]).toLowerCase();
         } catch (Exception var14) {
         }

         try {
            System.getProperty(z[30]).toLowerCase();
         } catch (Exception var13) {
         }

         try {
            m = System.getProperty(z[31]);
            if (m != null) {
               m = m + "/";
            }
         } catch (Exception var12) {
         }

         if (null == m) {
            m = z[37];
         }

         try {
            this.n = Toolkit.getDefaultToolkit().getSystemEventQueue();
         } catch (Throwable var11) {
         }

         if (!this.c) {
            try {
               u = Class.forName(z[40]).getDeclaredMethod(z[25], boolean.class);
            } catch (Exception var10) {
            }

            try {
               y = Class.forName(z[28]).getDeclaredMethod(z[38], boolean.class);
            } catch (Exception var9) {
            }
         }

         r.a(b, (byte)101, p);
         if (this.j) {
            this.s = new d(r.a(b, null, z[35], 0), z[8], 25L);
            this.f = new d(r.a(2, z[33]), z[8], 314572800L);
            this.v = new d(r.a(2, z[26]), z[8], 1048576L);
            this.l = new d[var3];

            for (int var5 = 0; var3 > var5; var5++) {
               this.l[var5] = new d(r.a(2, z[41] + var5), z[8], 1048576L);
            }

            if (this.c) {
               try {
                  Class.forName(z[27]).newInstance();
               } catch (Throwable var8) {
               }
            }

            try {
               if (this.c) {
                  this.w = new wa();
               } else {
                  this.i = Class.forName(z[15]).newInstance();
               }
            } catch (Throwable var7) {
            }

            try {
               if (!this.c) {
                  this.x = Class.forName("j").newInstance();
               } else {
                  this.t = new q();
               }
            } catch (Throwable var6) {
            }
         }

         this.e = false;
         this.r = new Thread(this);
         this.r.setPriority(10);
         this.r.setDaemon(true);
         this.r.start();
      } catch (RuntimeException var17) {
         throw var17;
      }
   }

   private static char[] z(String var0) {
      char[] var10000 = var0.toCharArray();
      if (var10000.length < 2) {
         var10000[0] = (char)(var10000[0] ^ 66);
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
               var10005 = 6;
               break;
            case 1:
               var10005 = 41;
               break;
            case 2:
               var10005 = 18;
               break;
            case 3:
               var10005 = 97;
               break;
            default:
               var10005 = 66;
         }

         var10001[var1] = (char)(var10004 ^ var10005);
      }

      return new String(var10001).intern();
   }
}
