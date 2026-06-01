final class na {
   static int[] a;
   static String[] d = new String[]{z(z("q,l?/")), z(z("q8}7/")), z(z("q8}7/")), z(z("q,l?/")), z(z("q\"a:/")), z(z("q8}7/")), z(z("q,l?/")), z(z("q,l?/"))};
   static int b;
   static int c;
   static int e;
   private static final String[] z = new String[]{z(z("_.h:")), z(z("_:*\u0017G")), z(z("Ju*x\u0012")), z(z("_:*\u0014G"))};

   static final byte[] a(String var0, int var1, byte[] var2, int var3) {
      try {
         if (var3 > -117) {
            a = (int[])null;
         }

         c++;
         return t.a(var2, -127, var1, var0, null);
      } catch (RuntimeException var5) {
         throw i.a(var5, z[1] + (var0 != null ? z[2] : z[0]) + ',' + var1 + ',' + (var2 != null ? z[2] : z[0]) + ',' + var3 + ')');
      }
   }

   static final m a(int var0, int var1, String var2) {
      try {
         b++;
         gb var3 = new gb();
         var3.h = var2;
         var3.f = var1;
         return var0 != 4718 ? (m)null : var3;
      } catch (RuntimeException var4) {
         throw i.a(var4, z[3] + var0 + ',' + var1 + ',' + (var2 != null ? z[2] : z[0]) + ')');
      }
   }

   private static char[] z(String var0) {
      char[] var10000 = var0.toCharArray();
      if (var10000.length < 2) {
         var10000[0] = (char)(var10000[0] ^ 111);
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
