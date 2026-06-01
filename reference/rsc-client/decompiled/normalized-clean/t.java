final class t {
   int e;
   int i;
   static int f;
   int j;
   int l;
   int m;
   static int c;
   String b;
   static int a;
   String p;
   String o = null;
   int d;
   static int k;
   static int g;
   static String[] h;
   static byte[][][] n;
   private static final String[] z = new String[]{
      z(z("9}_\u0015\t")), z(z(",&\u001dW")), z(z("6}3\u0013")), z(z("6}2\u0013")), z(z("6}5\u0013")), z(z("6}0\u0013"))
   };

   static final void a(int var0, int var1, int var2, int[] var3, int[] var4, int var5, int var6, int var7) {
      boolean var10 = client.vh;

      try {
         a++;
         if (var2 < 0) {
            var0 = var4[0xFF & var5 >> 8];
            var1 <<= 2;
            var5 += var1;
            int var8 = var2 / 16;
            int var9 = var8;

            while (true) {
               if (-1 < ~var9) {
                  var3[var6++] = var0;
                  var3[var6++] = var0;
                  var3[var6++] = var0;
                  var3[var6++] = var0;
                  var0 = var4[(var5 & 65525) >> 8];
                  var5 += var1;
                  var3[var6++] = var0;
                  var3[var6++] = var0;
                  var3[var6++] = var0;
                  var3[var6++] = var0;
                  var0 = var4[(65469 & var5) >> 8];
                  var3[var6++] = var0;
                  var5 += var1;
                  var3[var6++] = var0;
                  var3[var6++] = var0;
                  var3[var6++] = var0;
                  var0 = var4[0xFF & var5 >> 8];
                  var5 += var1;
                  var3[var6++] = var0;
                  var3[var6++] = var0;
                  var3[var6++] = var0;
                  var3[var6++] = var0;
                  var0 = var4[(65330 & var5) >> 8];
                  var5 += var1;
                  var9++;
                  if (var10) {
                     break;
                  }

                  if (!var10) {
                     continue;
                  }
               }

               var8 = -(var2 % 16);
               break;
            }

            if (var7 != 418609192) {
               a(76, -63, -59, (int[])null, (int[])null, 124, -66, -99);
            }

            var9 = 0;

            while (var8 > var9) {
               var3[var6++] = var0;
               if (var10) {
                  break;
               }

               if (~(3 & var9) == -4) {
                  var0 = var4[0xFF & var5 >> 8];
                  var5 += var1;
               }

               var9++;
               if (var10) {
                  break;
               }
            }
         }
      } catch (RuntimeException var11) {
         throw i.a(
            var11,
            z[3]
               + var0
               + ','
               + var1
               + ','
               + var2
               + ','
               + (var3 != null ? z[0] : z[1])
               + ','
               + (var4 != null ? z[0] : z[1])
               + ','
               + var5
               + ','
               + var6
               + ','
               + var7
               + ')'
         );
      }
   }

   static final byte[] a(byte[] var0, int var1, int var2, String var3, byte[] var4) {
      boolean var13 = client.vh;

      try {
         k++;
         int var5 = (255 & var0[1]) + (var0[0] & 255) * 256;
         var3 = var3.toUpperCase();
         int var6 = 0;
         int var7 = 0;

         while (true) {
            if (~var3.length() < ~var7) {
               int var10000 = 61 * var6 + (var3.charAt(var7) - ' ');
               if (var13) {
                  break;
               }

               var6 = var10000;
               var7++;
               if (!var13) {
                  continue;
               }
            }

            int var16 = 74 % ((var1 - -74) / 49);
            break;
         }

         var7 = 2 + 10 * var5;
         int var8 = 0;

         while (var8 < var5) {
            int var9 = (255 & var0[5 + var8 * 10])
               + (255 & var0[var8 * 10 + 4]) * 256
               + (var0[3 + 10 * var8] & 255) * 65536
               + (255 & var0[2 + 10 * var8]) * 16777216;
            int var10 = (var0[10 * var8 + 7] & 255) * 256 + (var0[var8 * 10 + 6] & 255) * 65536 + (255 & var0[var8 * 10 + 8]);
            int var11 = (var0[10 + 10 * var8] & 255) * 256 + 65536 * (255 & var0[var8 * 10 + 9]) + (var0[10 * var8 + 11] & 255);
            if (var9 == var6) {
               if (null == var4) {
                  var4 = new byte[var2 + var10];
               }

               if (var11 != var10) {
                  ea.a(var4, var10, var0, var11, var7);
                  if (!var13) {
                     return var4;
                  }
               }

               int var12 = 0;

               while (var12 < var10) {
                  if (var13) {
                     return var4;
                  }

                  var4[var12] = var0[var12 + var7];
                  var12++;
                  if (var13) {
                     break;
                  }
               }

               return var4;
            }

            var7 += var11;
            var8++;
            if (var13) {
               break;
            }
         }

         return null;
      } catch (RuntimeException var14) {
         throw i.a(
            var14,
            z[4] + (var0 != null ? z[0] : z[1]) + ',' + var1 + ',' + var2 + ',' + (var3 != null ? z[0] : z[1]) + ',' + (var4 != null ? z[0] : z[1]) + ')'
         );
      }
   }

   static final int a(int var0) {
      try {
         if (var0 != 65525) {
            h = (String[])null;
         }

         c++;
         int var1 = d.a(ka.b, (byte)100, b.v);
         ka.b += 2;
         return var1;
      } catch (RuntimeException var2) {
         throw i.a(var2, z[5] + var0 + 41);
      }
   }

   final void a(String var1, int var2, int var3, int var4, int var5, String var6, int var7, int var8, String var9, String var10, int var11, String var12) {
      try {
         f++;
         this.p = var1;
         this.l = var2;
         this.m = var4;
         if (var7 > 69) {
            this.d = var8;
            this.e = var3;
            this.o = var9;
            this.i = var5;
            this.j = var11;
            this.b = var12;
         }
      } catch (RuntimeException var14) {
         throw i.a(
            var14,
            z[2]
               + (var1 != null ? z[0] : z[1])
               + ','
               + var2
               + ','
               + var3
               + ','
               + var4
               + ','
               + var5
               + ','
               + (var6 != null ? z[0] : z[1])
               + ','
               + var7
               + ','
               + var8
               + ','
               + (var9 != null ? z[0] : z[1])
               + ','
               + (var10 != null ? z[0] : z[1])
               + ','
               + var11
               + ','
               + (var12 != null ? z[0] : z[1])
               + ')'
         );
      }
   }

   t() {
      this.p = null;
   }

   private static char[] z(String var0) {
      char[] var10000 = var0.toCharArray();
      if (var10000.length < 2) {
         var10000[0] = (char)(var10000[0] ^ 116);
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
               var10005 = 66;
               break;
            case 1:
               var10005 = 83;
               break;
            case 2:
               var10005 = 113;
               break;
            case 3:
               var10005 = 59;
               break;
            default:
               var10005 = 116;
         }

         var10001[var1] = (char)(var10004 ^ var10005);
      }

      return new String(var10001).intern();
   }
}
