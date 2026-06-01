final class oa {
   static int[][] d;
   static int e;
   static String[] c = new String[]{z(z("R\u00037^U7\u00036VEr\u001fcTA7\u00047^JdM7T\u0007d\u0019\"PB7\f-_\u0007g\u001f&HT7\b-OBe"))};
   static String[] a;
   static int b;
   private static final String[] z = new String[]{z(z("x\fmz\u000f")), z(z("y\u0018/W")), z(z("lCm\u0015Z"))};

   static final int a(String var0, byte var1, byte[] var2) {
      boolean var9 = client.vh;

      try {
         e++;
         int var3 = d.a(0, (byte)48, var2);
         int var4 = 0;
         var0 = var0.toUpperCase();
         int var5 = 0;

         while (true) {
            if (~var0.length() < ~var5) {
               var4 = -32 + var0.charAt(var5) + 61 * var4;
               var5++;
               if (var9) {
                  break;
               }

               if (!var9) {
                  continue;
               }
            }

            var5 = 2 + var3 * 10;
            break;
         }

         if (var1 != 68) {
            return 71;
         }

         int var6 = 0;

         int var10000;
         while (true) {
            if (~var6 > ~var3) {
               int var7 = (var2[5 + 10 * var6] & 255)
                  + (255 & var2[4 + 10 * var6]) * 256
                  + (255 & var2[2 + 10 * var6]) * 16777216
                  + 65536 * (var2[10 * var6 + 3] & 255);
               int var8 = (var2[10 * var6 + 11] & 255) + 256 * (var2[var6 * 10 - -10] & 255) + (255 & var2[9 + 10 * var6]) * 65536;
               var10000 = ~var4;
               if (var9) {
                  break;
               }

               if (var10000 == ~var7) {
                  return var5;
               }

               var5 += var8;
               var6++;
               if (!var9) {
                  continue;
               }
            }

            var10000 = 0;
            break;
         }

         return var10000;
      } catch (RuntimeException var10) {
         throw i.a(var10, z[0] + (var0 != null ? z[2] : z[1]) + 44 + var1 + 44 + (var2 != null ? z[2] : z[1]) + 41);
      }
   }

   private static char[] z(String var0) {
      char[] var10000 = var0.toCharArray();
      if (var10000.length < 2) {
         var10000[0] = (char)(var10000[0] ^ 39);
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
               var10005 = 23;
               break;
            case 1:
               var10005 = 109;
               break;
            case 2:
               var10005 = 67;
               break;
            case 3:
               var10005 = 59;
               break;
            default:
               var10005 = 39;
         }

         var10001[var1] = (char)(var10004 ^ var10005);
      }

      return new String(var10001).intern();
   }
}
