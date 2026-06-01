final class ia {
   static int e;
   static int b = 0;
   static String[] a = new String[100];
   static int[] d = new int[256];
   static int h;
   static int f;
   static int i = 0;
   static String[] g = new String[100];
   static int c;
   private static final String[] z = new String[]{
      z(z("3Wd\\*")), z(z("&\f&\u001e")), z(z("!\u0018d0\u007f")), z(z("!\u0018d1\u007f")), z(z("!\u0018d3\u007f"))
   };

   static final void a(int var0, int var1, int[] var2, int var3, int var4, int[] var5, int var6, byte var7) {
      boolean var11 = client.vh;

      try {
         e++;
         if (0 > var6) {
            var1 = var2[(65465 & var3) >> -1668975128];
            var0 <<= 1;
            var3 += var0;
            int var8 = var6 / 8;
            int var10 = 69 % ((var7 - 27) / 45);
            int var9 = var8;

            while (true) {
               if (0 > var9) {
                  var5[var4++] = var1;
                  var5[var4++] = var1;
                  var1 = var2[(65503 & var3) >> 2132179464];
                  var3 += var0;
                  var5[var4++] = var1;
                  var5[var4++] = var1;
                  var1 = var2[(var3 & 65391) >> -1202367672];
                  var5[var4++] = var1;
                  var3 += var0;
                  var5[var4++] = var1;
                  var1 = var2[(65302 & var3) >> 1020025064];
                  var3 += var0;
                  var5[var4++] = var1;
                  var5[var4++] = var1;
                  var1 = var2[(var3 & 65386) >> -31505560];
                  var3 += var0;
                  var9++;
                  if (var11) {
                     break;
                  }

                  if (!var11) {
                     continue;
                  }
               }

               var8 = -(var6 % 8);
               break;
            }

            var9 = 0;

            while (~var9 > ~var8) {
               var5[var4++] = var1;
               if (var11) {
                  break;
               }

               if (1 == (var9 & 1)) {
                  var1 = var2[var3 >> 867708808 & 0xFF];
                  var3 += var0;
               }

               var9++;
               if (var11) {
                  break;
               }
            }
         }
      } catch (RuntimeException var12) {
         throw i.a(
            var12,
            z[2]
               + var0
               + ','
               + var1
               + ','
               + (var2 != null ? z[0] : z[1])
               + ','
               + var3
               + ','
               + var4
               + ','
               + (var5 != null ? z[0] : z[1])
               + ','
               + var6
               + ','
               + var7
               + ')'
         );
      }
   }

   static final String a(tb var0, boolean var1) {
      try {
         if (var1) {
            a = (String[])null;
         }

         c++;
         return client.a(0, var0, 32767);
      } catch (RuntimeException var3) {
         throw i.a(var3, z[3] + (var0 != null ? z[0] : z[1]) + ',' + var1 + ')');
      }
   }

   static final boolean a(v var0, byte var1) {
      try {
         int var2 = 3 / ((-39 - var1) / 54);
         f++;
         return da.O == var0 || ga.c == var0 || ta.f == var0 || var0 == eb.d || gb.n == var0;
      } catch (RuntimeException var3) {
         throw i.a(var3, z[4] + (var0 != null ? z[0] : z[1]) + ',' + var1 + ')');
      }
   }

   private static char[] z(String var0) {
      char[] var10000 = var0.toCharArray();
      if (var10000.length < 2) {
         var10000[0] = (char)(var10000[0] ^ 87);
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
               var10005 = 72;
               break;
            case 1:
               var10005 = 121;
               break;
            case 2:
               var10005 = 74;
               break;
            case 3:
               var10005 = 114;
               break;
            default:
               var10005 = 87;
         }

         var10001[var1] = (char)(var10004 ^ var10005);
      }

      return new String(var10001).intern();
   }
}
