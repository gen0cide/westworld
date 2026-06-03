import java.io.IOException;

final class fa extends IOException {
   static int b;
   static String[] a = new String[]{z(z("F%:\u001f\u0014f4/ZZg1(\u001fF23,Z]f9'\t\u0014f3j\u0018Ak|+\u0014P2,8\u001fGa|/\u0014@w."))};
   static int d = 234;
   static int[] e;
   static int[] c;

   fa(String var1) {
      super(var1);
   }

   private static char[] z(String var0) {
      char[] var10000 = var0.toCharArray();
      if (var10000.length < 2) {
         var10000[0] = (char)(var10000[0] ^ 52);
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
               var10005 = 18;
               break;
            case 1:
               var10005 = 92;
               break;
            case 2:
               var10005 = 74;
               break;
            case 3:
               var10005 = 122;
               break;
            default:
               var10005 = 52;
         }

         var10001[var1] = (char)(var10004 ^ var10005);
      }

      return new String(var10001).intern();
   }
}
