final class na {
   static int[] a;
   static String[] d = new String[]{z(z("q,l?/")), z(z("q8}7/")), z(z("q8}7/")), z(z("q,l?/")), z(z("q\"a:/")), z(z("q8}7/")), z(z("q,l?/")), z(z("q,l?/"))};
   static int b;
   static int c;
   static int e;
   private static final String[] z = new String[]{z(z("_.h:")), z(z("_:*\u0017G")), z(z("Ju*x\u0012")), z(z("_:*\u0014G"))};

   static final byte[] a(String var0, int var1, byte[] var2, int var3) {
      try {
         try {
            if (var3 > -117) {
               a = (int[])null;
            }
         } catch (RuntimeException var7) {
            throw var7;
         }

         c++;
         return t.a(var2, -127, var1, var0, null);
      } catch (RuntimeException var8) {
         RuntimeException var4 = var8;

         RuntimeException var10000;
         StringBuilder var10001;
         String var10002;
         label37: {
            try {
               var10000 = var4;
               var10001 = new StringBuilder().append(z[1]);
               if (var0 != null) {
                  var10002 = z[2];
                  break label37;
               }
            } catch (RuntimeException var6) {
               throw var6;
            }

            var10002 = z[0];
         }

         try {
            var10001 = var10001.append(var10002).append(',').append(var1).append(',');
            if (var2 != null) {
               throw i.a(var10000, var10001.append(z[2]).append(',').append(var3).append(')').toString());
            }
         } catch (RuntimeException var5) {
            throw var5;
         }

         throw i.a(var10000, var10001.append(z[0]).append(',').append(var3).append(')').toString());
      }
   }

   static final m a(int var0, int var1, String var2) {
      try {
         b++;
         gb var6 = new gb();
         var6.h = var2;
         var6.f = var1;
         return (m)(var0 != 4718 ? (m)null : var6);
      } catch (RuntimeException var5) {
         RuntimeException var3 = var5;

         RuntimeException var10000;
         StringBuilder var10001;
         try {
            var10000 = var3;
            var10001 = new StringBuilder().append(z[3]).append(var0).append(',').append(var1).append(',');
            if (var2 != null) {
               throw i.a(var3, var10001.append(z[2]).append(')').toString());
            }
         } catch (RuntimeException var4) {
            throw var4;
         }

         throw i.a(var10000, var10001.append(z[0]).append(')').toString());
      }
   }

   private static char[] z(String var0) {
      char[] var10000 = var0.toCharArray();
      if (var10000.length < 2) {
         var10000[0] = (char)(var10000[0] ^ 'o');
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
               var10005 = 49;
               break;
            case 1:
               var10005 = 91;
               break;
            case 2:
               var10005 = 4;
               break;
            case 3:
               var10005 = 86;
               break;
            default:
               var10005 = 111;
         }

         var10001[var1] = (char)(var10004 ^ var10005);
      }

      return new String(var10001).intern();
   }
}
