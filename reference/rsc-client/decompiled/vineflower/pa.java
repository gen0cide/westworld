import java.awt.Component;
import java.awt.Image;
import java.awt.image.IndexColorModel;

final class pa {
   static c b;
   static int[] g = new int[100];
   static int[] f;
   private static byte[] e = new byte[64];
   static int[] a = new int[512];
   static c k;
   static int i;
   static int c;
   static long h;
   static int[] j = new int[2048];
   static int[] d = new int[256];
   private static final String[] z = new String[]{z(z("J_\f\t%")), z(z("TKN'")), z(z("A\u0010\fep")), z(z("J_\f\n%"))};

   static final Image a(int var0, Component var1, byte[] var2) {
      try {
         k.o = var2[12] + var2[13] * 256;
         da.bb = var2[14] + 256 * var2[15];
         i++;
         byte[] var15 = new byte[256];
         byte[] var4 = new byte[256];
         byte[] var5 = new byte[256];
         int var6 = 0;

         try {
            while (-257 < ~var6) {
               var15[var6] = var2[20 + var6 * 3];
               var4[var6] = var2[3 * var6 + 19];
               var5[var6] = var2[3 * var6 + 18];
               var6++;
            }
         } catch (RuntimeException var13) {
            throw var13;
         }

         m.d = new IndexColorModel(8, 256, var15, var4, var5);
         byte[] var16 = new byte[k.o * da.bb];
         int var7 = 0;

         for (int var8 = -1 + da.bb; -1 >= ~var8; var8--) {
            int var9 = 0;

            try {
               while (k.o > var9) {
                  var16[var7++] = var2[k.o * var8 + 786 + var9];
                  var9++;
               }
            } catch (RuntimeException var12) {
               throw var12;
            }
         }

         Image var17 = var1.createImage(da.db);
         int var18 = 113 / ((var0 - 10) / 53);
         lb.a(true, var16);
         var1.prepareImage(var17, da.db);
         lb.a(true, var16);
         var1.prepareImage(var17, da.db);
         lb.a(true, var16);
         var1.prepareImage(var17, da.db);
         return var17;
      } catch (RuntimeException var14) {
         RuntimeException var3 = var14;

         RuntimeException var10000;
         StringBuilder var10001;
         String var10002;
         label47: {
            try {
               var10000 = var3;
               var10001 = new StringBuilder().append(z[0]).append(var0).append(',');
               if (var1 != null) {
                  var10002 = z[2];
                  break label47;
               }
            } catch (RuntimeException var11) {
               throw var11;
            }

            var10002 = z[1];
         }

         try {
            var10001 = var10001.append(var10002).append(',');
            if (var2 != null) {
               throw i.a(var10000, var10001.append(z[2]).append(')').toString());
            }
         } catch (RuntimeException var10) {
            throw var10;
         }

         throw i.a(var10000, var10001.append(z[1]).append(')').toString());
      }
   }

   static final byte[] a(int var0) {
      try {
         c++;
         byte[] var1 = new byte[256];
         int var2 = -128;

         try {
            if (var0 > -125) {
               a(-7, (Component)null, (byte[])null);
            }
         } catch (RuntimeException var8) {
            throw var8;
         }

         while (var2 < 127) {
            int var3 = (byte)var2;
            var3 = ~var3;
            int var4 = var3 & 128;
            int var5 = var3 >> -1437861628 & 7;
            int var6 = 15 & var3;
            var6 |= 16;
            var6 = 1 + (var6 << -1014531647);
            int var7 = var6 << var5 + 2;
            var7 = -132 + var7;

            label32: {
               try {
                  if (~var4 == -1) {
                     break label32;
                  }
               } catch (RuntimeException var9) {
                  throw var9;
               }

               var7 = -var7;
            }

            var1[ib.a(var2, 255)] = (byte)(var7 / 256);
            var2++;
         }

         return var1;
      } catch (RuntimeException var10) {
         throw i.a(var10, z[3] + var0 + ')');
      }
   }

   static {
      int var0 = 0;

      try {
         while (256 > var0) {
            a[var0] = (int)(32768.0 * Math.sin(0.02454369 * (double)var0));
            a[256 + var0] = (int)(32768.0 * Math.cos((double)var0 * 0.02454369));
            var0++;
         }
      } catch (RuntimeException var8) {
         throw var8;
      }

      var0 = 0;

      try {
         while (-1025 < ~var0) {
            j[var0] = (int)(Math.sin((double)var0 * 0.00613592315) * 32768.0);
            j[var0 - -1024] = (int)(Math.cos((double)var0 * 0.00613592315) * 32768.0);
            var0++;
         }
      } catch (RuntimeException var7) {
         throw var7;
      }

      var0 = 0;

      try {
         while (-11 < ~var0) {
            e[var0] = (byte)(48 + var0);
            var0++;
         }
      } catch (RuntimeException var6) {
         throw var6;
      }

      var0 = 0;

      try {
         while (~var0 > -27) {
            e[var0 - -10] = (byte)(var0 + 65);
            var0++;
         }
      } catch (RuntimeException var5) {
         throw var5;
      }

      var0 = 0;

      try {
         while (26 > var0) {
            e[var0 - -36] = (byte)(97 + var0);
            var0++;
         }
      } catch (RuntimeException var4) {
         throw var4;
      }

      e[63] = 36;
      e[62] = -93;
      var0 = 0;

      try {
         while (-11 < ~var0) {
            d[var0 + 48] = var0++;
         }
      } catch (RuntimeException var3) {
         throw var3;
      }

      var0 = 0;

      try {
         while (-27 < ~var0) {
            d[var0 + 65] = var0 - -10;
            var0++;
         }
      } catch (RuntimeException var2) {
         throw var2;
      }

      var0 = 0;

      try {
         while (var0 < 26) {
            d[var0 + 97] = 36 + var0;
            var0++;
         }
      } catch (RuntimeException var1) {
         throw var1;
      }

      d[36] = 63;
      d[163] = 62;
   }

   private static char[] z(String var0) {
      char[] var10000 = var0.toCharArray();
      if (var10000.length < 2) {
         var10000[0] = (char)(var10000[0] ^ '\r');
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
               var10005 = 58;
               break;
            case 1:
               var10005 = 62;
               break;
            case 2:
               var10005 = 34;
               break;
            case 3:
               var10005 = 75;
               break;
            default:
               var10005 = 13;
         }

         var10001[var1] = (char)(var10004 ^ var10005);
      }

      return new String(var10001).intern();
   }
}
