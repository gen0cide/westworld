import java.io.DataInputStream;
import java.io.IOException;
import java.io.OutputStream;
import java.net.Socket;
import java.net.URL;

final class jb implements Runnable {
   static int m;
   private c b;
   static int i;
   private g f;
   static int d;
   static int j;
   private DataInputStream q;
   private tb h;
   static int e;
   private URL g;
   private g c;
   static int n;
   static int[] k;
   static int p;
   private g l;
   private int a;
   static int o = 0;
   private static final String[] z = new String[]{
      z(z("o\u007f+$\u000f")),
      z(z("MZ4lZI\u00103")),
      z(z("my]Y}fz:")),
      z(z("MZ4\\\u0007")),
      z(z("-2")),
      z(z("MZ4Z\u0007")),
      z(z("\\\u001640R")),
      z(z("IMvr")),
      z(z("MZ4\"FIQn \u0007")),
      z(z("MZ4xFIYvwUB\u00103")),
      z(z("MZ4]\u0007"))
   };

   @Override
   public final void run() {
      boolean var4 = client.vh;

      try {
         m++;

         try {
            int var10000;
            int var10001;
            while (true) {
               if (~this.h.w > ~this.h.F.length) {
                  int var1 = this.q.read(this.h.F, this.h.w, this.h.F.length - this.h.w);
                  var10000 = ~var1;
                  var10001 = -1;
                  if (var4) {
                     break;
                  }

                  if (var10000 <= -1) {
                     this.h.w += var1;
                     if (!var4) {
                        continue;
                     }
                  }
               }

               var10000 = this.h.w;
               var10001 = this.h.F.length;
               break;
            }

            if (var10000 == var10001) {
               throw new Exception(z[0] + this.h.F.length + " " + this.g);
            }

            synchronized (this) {
               this.finalize();
               this.a = 3;
            }
         } catch (Exception var7) {
            synchronized (this) {
               this.finalize();
               this.a++;
            }
         }
      } catch (RuntimeException var8) {
         throw i.a(var8, z[1]);
      }
   }

   static final void a(
      int[] var0,
      int var1,
      int var2,
      int var3,
      int var4,
      int var5,
      int var6,
      int var7,
      int var8,
      int var9,
      int[] var10,
      boolean var11,
      int var12,
      int var13,
      int var14
   ) {
      boolean var22 = client.vh;

      try {
         d++;
         if (~var7 < -1) {
            int var15 = 0;
            int var16 = 0;
            if (~var3 != -1) {
               var16 = var13 / var3 << 223739206;
               var15 = var8 / var3 << -902154106;
            }

            var4 <<= 2;
            if (var11) {
               o = 21;
            }

            label108: {
               if (~var15 > -1) {
                  var15 = 0;
                  if (!var22) {
                     break label108;
                  }
               }

               if (4032 < var15) {
                  var15 = 4032;
               }
            }

            int var19 = var7;

            label102:
            while (true) {
               byte var10000 = 0;
               int var10001 = var19;

               label100:
               while (var10000 < var10001) {
                  var3 += var2;
                  var14 = var15;
                  var8 += var12;
                  var9 = var16;
                  var13 += var1;
                  if (var22) {
                     return;
                  }

                  if (~var3 != -1) {
                     var15 = var8 / var3 << -213852602;
                     var16 = var13 / var3 << -474023130;
                  }

                  label97: {
                     if (~var15 <= -1) {
                        if (4032 >= var15) {
                           break label97;
                        }

                        var15 = 4032;
                        if (!var22) {
                           break label97;
                        }
                     }

                     var15 = 0;
                  }

                  label118: {
                     int var18 = var16 - var9 >> -1841585212;
                     int var17 = -var14 + var15 >> 166246532;
                     int var20 = var5 >> -1249879148;
                     var14 += var5 & 786432;
                     var5 += var4;
                     if (var19 >= 16) {
                        var0[var6++] = ib.a(var0[var6] >> -496952415, 8355711) + (var10[ib.a(4032, var9) + (var14 >> 955305670)] >>> var20);
                        var14 += var17;
                        var9 += var18;
                        var0[var6++] = (ib.a(var0[var6], 16711423) >> 393665345) + (var10[ib.a(4032, var9) - -(var14 >> 556791558)] >>> var20);
                        var9 += var18;
                        var14 += var17;
                        var0[var6++] = (ib.a(16711423, var0[var6]) >> -2007060127) + (var10[ib.a(var9, 4032) + (var14 >> -1069632730)] >>> var20);
                        var9 += var18;
                        var14 += var17;
                        var0[var6++] = (ib.a(var0[var6], 16711423) >> -526841663) + (var10[(var14 >> -1891324570) + ib.a(4032, var9)] >>> var20);
                        var14 += var17;
                        var9 += var18;
                        var14 = (var5 & 786432) + (4095 & var14);
                        int var24 = var5 >> -580603052;
                        var0[var6++] = (var10[ib.a(var9, 4032) + (var14 >> 1328606726)] >>> var24) + (ib.a(var0[var6], 16711422) >> 604787489);
                        var5 += var4;
                        var14 += var17;
                        var9 += var18;
                        var0[var6++] = (var10[(var14 >> -830951482) + ib.a(4032, var9)] >>> var24) + (ib.a(var0[var6], 16711423) >> 310428257);
                        var14 += var17;
                        var9 += var18;
                        var0[var6++] = (var10[ib.a(4032, var9) + (var14 >> -1841159226)] >>> var24) - -(ib.a(var0[var6], 16711423) >> -1760233471);
                        var9 += var18;
                        var14 += var17;
                        var0[var6++] = (var10[ib.a(4032, var9) + (var14 >> 1454319654)] >>> var24) - -(ib.a(16711423, var0[var6]) >> 1605358369);
                        var14 += var17;
                        var9 += var18;
                        var14 = (786432 & var5) + (4095 & var14);
                        int var25 = var5 >> 1147218452;
                        var0[var6++] = (var10[ib.a(4032, var9) - -(var14 >> 1983636742)] >>> var25) - -(ib.a(var0[var6], 16711422) >> 1637168449);
                        var5 += var4;
                        var14 += var17;
                        var9 += var18;
                        var0[var6++] = (var10[ib.a(var9, 4032) - -(var14 >> 1901625030)] >>> var25) - -ib.a(var0[var6] >> -256795167, 8355711);
                        var9 += var18;
                        var14 += var17;
                        var0[var6++] = (var10[ib.a(4032, var9) + (var14 >> 1605754694)] >>> var25) - -(ib.a(var0[var6], 16711423) >> -216359295);
                        var14 += var17;
                        var9 += var18;
                        var0[var6++] = (ib.a(16711422, var0[var6]) >> 791103809) + (var10[(var14 >> 371413222) + ib.a(var9, 4032)] >>> var25);
                        var14 += var17;
                        var9 += var18;
                        var14 = (var5 & 786432) + (var14 & 4095);
                        var20 = var5 >> 711720340;
                        var0[var6++] = (var10[ib.a(var9, 4032) - -(var14 >> -2780922)] >>> var20) - -(ib.a(16711422, var0[var6]) >> 962756193);
                        var5 += var4;
                        var14 += var17;
                        var9 += var18;
                        var0[var6++] = (ib.a(var0[var6], 16711423) >> -838985215) + (var10[ib.a(var9, 4032) - -(var14 >> 1389805702)] >>> var20);
                        var14 += var17;
                        var9 += var18;
                        var0[var6++] = (var10[ib.a(4032, var9) - -(var14 >> -1171869722)] >>> var20) - -ib.a(var0[var6] >> 1072420929, 8355711);
                        var9 += var18;
                        var14 += var17;
                        var0[var6++] = (var10[(var14 >> 227032774) + ib.a(4032, var9)] >>> var20) + (ib.a(var0[var6], 16711423) >> -454180287);
                        if (!var22) {
                           break label118;
                        }
                     }

                     int var21 = 0;

                     while (var19 > var21) {
                        var0[var6++] = (var10[(var14 >> -208962138) + ib.a(var9, 4032)] >>> var20) - -(ib.a(16711422, var0[var6]) >> -1883934911);
                        var9 += var18;
                        var14 += var17;
                        var10000 = 3;
                        var10001 = var21 & 3;
                        if (var22) {
                           continue label100;
                        }

                        if (3 == var10001) {
                           var20 = var5 >> -2030987340;
                           var14 = (var14 & 4095) - -(786432 & var5);
                           var5 += var4;
                        }

                        var21++;
                        if (var22) {
                           break;
                        }
                     }
                  }

                  var19 -= 16;
                  if (var22) {
                     return;
                  }
                  continue label102;
               }

               return;
            }
         }
      } catch (RuntimeException var23) {
         throw i.a(
            var23,
            z[5]
               + (var0 != null ? z[6] : z[7])
               + ','
               + var1
               + ','
               + var2
               + ','
               + var3
               + ','
               + var4
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
               + ','
               + (var10 != null ? z[6] : z[7])
               + ','
               + var11
               + ','
               + var12
               + ','
               + var13
               + ','
               + var14
               + ')'
         );
      }
   }

   final synchronized boolean a(int var1) {
      try {
         e++;
         if (this.a < 2) {
            if (~this.a == -1) {
               if (null == this.c) {
                  this.c = this.b.a((byte)74, this.g);
               }

               if (0 == this.c.b) {
                  return false;
               }

               if (-2 != ~this.c.b) {
                  this.c = null;
                  this.a++;
                  return false;
               }
            }

            if (-2 == ~this.a) {
               if (null == this.f) {
                  this.f = this.b.a(this.g.getHost(), 443, -68);
               }

               if (~this.f.b == -1) {
                  return false;
               }

               if (this.f.b != 1) {
                  this.f = null;
                  this.a++;
                  return false;
               }
            }

            if (this.q == null) {
               try {
                  if (~this.a == -1) {
                     this.q = (DataInputStream)this.c.d;
                  }

                  if (~this.a == -2) {
                     Socket var2 = (Socket)this.f.d;
                     var2.setSoTimeout(10000);
                     OutputStream var3 = var2.getOutputStream();
                     var3.write(17);
                     var3.write(h.a(z[2] + this.g.getFile() + z[4], (byte)-104));
                     this.q = new DataInputStream(var2.getInputStream());
                  }

                  this.h.w = 0;
               } catch (IOException var4) {
                  this.finalize();
                  this.a++;
               }
            }

            if (this.l == null) {
               this.l = this.b.a(true, this, 5);
            }

            if (~this.l.b == -1) {
               return false;
            }

            if (~this.l.b != var1) {
               this.finalize();
               this.a++;
            }

            return false;
         } else {
            return true;
         }
      } catch (RuntimeException var5) {
         throw i.a(var5, z[3] + var1 + ')');
      }
   }

   final tb a(byte var1) {
      try {
         if (var1 > -110) {
            this.run();
         }

         i++;
         return 3 == this.a ? this.h : null;
      } catch (RuntimeException var3) {
         throw i.a(var3, z[10] + var1 + ')');
      }
   }

   static final String a(boolean var0, String var1, String var2, String var3) {
      boolean var5 = client.vh;

      try {
         if (!var0) {
            a((int[])null, 78, -46, -87, -87, -58, -96, -121, 50, -80, (int[])null, false, -54, -83, 52);
         }

         j++;
         int var4 = var3.indexOf(var2);

         String var10000;
         while (true) {
            if (~var4 != 0) {
               var3 = var3.substring(0, var4) + var1 + var3.substring(var2.length() + var4);
               var10000 = var3;
               if (var5) {
                  break;
               }

               var4 = var3.indexOf(var2, var4 - -var1.length());
               if (!var5) {
                  continue;
               }
            }

            var10000 = var3;
            break;
         }

         return var10000;
      } catch (RuntimeException var6) {
         throw var6;
      }
   }

   jb(c var1, URL var2, int var3) {
      try {
         this.g = var2;
         this.b = var1;
         this.h = new tb(var3);
      } catch (RuntimeException var5) {
         throw i.a(var5, z[8] + (var1 != null ? z[6] : z[7]) + ',' + (var2 != null ? z[6] : z[7]) + ',' + var3 + ')');
      }
   }

   @Override
   protected final void finalize() {
      try {
         n++;
         if (null != this.c) {
            if (null != this.c.d) {
               try {
                  ((DataInputStream)this.c.d).close();
               } catch (Exception var4) {
               }
            }

            this.c = null;
         }

         if (null != this.f) {
            if (this.f.d != null) {
               try {
                  ((Socket)this.f.d).close();
               } catch (Exception var3) {
               }
            }

            this.f = null;
         }

         if (null != this.q) {
            try {
               this.q.close();
            } catch (Exception var2) {
            }

            this.q = null;
         }

         this.l = null;
      } catch (RuntimeException var5) {
         throw i.a(var5, z[9]);
      }
   }

   private static char[] z(String var0) {
      char[] var10000 = var0.toCharArray();
      if (var10000.length < 2) {
         var10000[0] = (char)(var10000[0] ^ 47);
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
               var10005 = 39;
               break;
            case 1:
               var10005 = 56;
               break;
            case 2:
               var10005 = 26;
               break;
            case 3:
               var10005 = 30;
               break;
            default:
               var10005 = 47;
         }

         var10001[var1] = (char)(var10004 ^ var10005);
      }

      return new String(var10001).intern();
   }
}
