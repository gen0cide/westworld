final class ub {
   static String[] c = new String[5000];
   static int f;
   static String[] b;
   static int e;
   static int[] g;
   static int d;
   static String[] a = new String[100];
   private static final String[] z = new String[]{
      z(z("\u0017\u000e\rJP")), z(z("\u0017\u000e\rKP")), z(z("\u0017\u000e\rIP")), z(z("\u0019B\r&\u0005")), z(z("\f\u0019Od"))
   };

   static final void a(int[] var0, byte var1, int var2, int var3, Object[] var4) {
      boolean var13 = client.vh;

      try {
         if (var1 <= -82) {
            int var10000;
            int var10001;
            label61: {
               if (~var2 > ~var3) {
                  int var5 = (var3 + var2) / 2;
                  int var6 = var2;
                  int var7 = var0[var5];
                  var0[var5] = var0[var3];
                  var0[var3] = var7;
                  Object var8 = var4[var5];
                  var4[var5] = var4[var3];
                  var4[var3] = var8;
                  int var9 = Integer.MIN_VALUE == ~var7 ? 0 : 1;
                  int var10 = var2;

                  while (var10 < var3) {
                     var10000 = (var10 & var9) + var7;
                     var10001 = var0[var10];
                     if (var13) {
                        break label61;
                     }

                     if (var10000 > var10001) {
                        int var11 = var0[var10];
                        var0[var10] = var0[var6];
                        var0[var6] = var11;
                        Object var12 = var4[var10];
                        var4[var10] = var4[var6];
                        var4[var6++] = var12;
                     }

                     var10++;
                     if (var13) {
                        break;
                     }
                  }

                  var0[var3] = var0[var6];
                  var0[var6] = var7;
                  var4[var3] = var4[var6];
                  var4[var6] = var8;
                  a(var0, (byte)-124, var2, var6 + -1, var4);
                  a(var0, (byte)-123, 1 + var6, var3, var4);
               }

               var10000 = e;
               var10001 = 1;
            }

            e = var10000 + var10001;
         }
      } catch (RuntimeException var14) {
         throw i.a(var14, z[2] + (var0 != null ? z[3] : z[4]) + ',' + var1 + ',' + var2 + ',' + var3 + ',' + (var4 != null ? z[3] : z[4]) + ')');
      }
   }

   static final v a(int var0, byte var1) {
      boolean var5 = client.vh;

      try {
         d++;
         if (var1 != 24) {
            c = (String[])null;
         }

         v[] var2 = i.a(var1 + -735);
         int var3 = 0;

         while (~var2.length < ~var3) {
            v var4 = var2[var3];
            if (var0 == var4.i) {
               return var4;
            }

            var3++;
            if (var5) {
               break;
            }
         }

         return null;
      } catch (RuntimeException var6) {
         throw i.a(var6, z[0] + var0 + ',' + var1 + ')');
      }
   }

   static final int a(byte var0) {
      try {
         f++;
         int var1 = m.a(true, ka.b, b.v);
         if (~var1 < -100000000) {
            var1 = -var1 + 99999999;
         }

         if (var0 != -105) {
            g = (int[])null;
         }

         ka.b += 4;
         return var1;
      } catch (RuntimeException var2) {
         throw i.a(var2, z[1] + var0 + 41);
      }
   }

   private static char[] z(String var0) {
      char[] var10000 = var0.toCharArray();
      if (var10000.length < 2) {
         var10000[0] = (char)(var10000[0] ^ 120);
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
               var10005 = 98;
               break;
            case 1:
               var10005 = 108;
               break;
            case 2:
               var10005 = 35;
               break;
            case 3:
               var10005 = 8;
               break;
            default:
               var10005 = 120;
         }

         var10001[var1] = (char)(var10004 ^ var10005);
      }

      return new String(var10001).intern();
   }
}
