final class la extends RuntimeException {
   Throwable e;
   String h;
   static v b = new v(z(z("[\fDi<")), "", z(z("{,dI\u001c")), 4);
   static int[] a;
   static int d;
   static byte[] c = new byte[520];
   static byte[][] g = new byte[12][];
   static String[] f;

   la(Throwable var1, String var2) {
      try {
         this.e = var1;
         this.h = var2;
      } catch (RuntimeException var4) {
         throw var4;
      }
   }

   private static char[] z(String var0) {
      char[] var10000 = var0.toCharArray();
      if (var10000.length < 2) {
         var10000[0] = (char)(var10000[0] ^ 'p');
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
               var10005 = 67;
               break;
            case 2:
               var10005 = 7;
               break;
            case 3:
               var10005 = 40;
               break;
            default:
               var10005 = 112;
         }

         var10001[var1] = (char)(var10004 ^ var10005);
      }

      return new String(var10001).intern();
   }
}
