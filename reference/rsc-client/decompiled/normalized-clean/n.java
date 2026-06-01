final class n {
   static int i;
   int e;
   static int[] a = new int[256];
   int k;
   static nb h = null;
   static String[] f = new String[]{
      z(z("9Z\u001e\"q\\Z\u001f*a\u0019FJ(e\\]\u001e\"n\u000f\u0014\u001e(#\u000eQ\u0007(u\u0019\u0014\u000b)g\\D\u0018\"p\u000f\u0014\u000f)w\u0019F"))
   };
   int d;
   int l;
   static int[] j;
   static int g;
   static int[] m;
   static int c;
   static int b;
   private static final String[] z = new String[]{z(z("Y\u0004\u000b")), z(z("9F\u0018(qF\u0014"))};

   static final void a(byte var0, String var1) {
      try {
         if (var0 != -93) {
            f = (String[])null;
         }

         i++;
         System.out.println(z[1] + jb.a(true, "\n", z[0], var1));
      } catch (RuntimeException var3) {
         throw var3;
      }
   }

   static {
      String var0 = z(
         z(
            "=v)\u0003F:s\"\u000eI7x'\tL,e8\u0014W)b=\u001fZ&U\b$g\u0019R\r/j\u0016_\u0006*m\u0013D\u001b5p\bA\u001c0{\u0005NZv1O\u0000_q4D\rKe X\u00114a)T\u001dG\u0018>Wo\u0011\u001a~G\u000eM\u0007 \u0002\u0018Vi=S\u000b6;#"
         )
      );

      for (int var1 = 0; var1 < 256; var1++) {
         int var2 = var0.indexOf(var1);
         if (~var2 == 0) {
            var2 = 74;
         }

         a[var1] = 9 * var2;
      }

      g = 0;
      j = new int[100];
      c = 0;
      b = 0;
   }

   private static char[] z(String var0) {
      char[] var10000 = var0.toCharArray();
      if (var10000.length < 2) {
         var10000[0] = (char)(var10000[0] ^ 3);
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
               var10005 = 124;
               break;
            case 1:
               var10005 = 52;
               break;
            case 2:
               var10005 = 106;
               break;
            case 3:
               var10005 = 71;
               break;
            default:
               var10005 = 3;
         }

         var10001[var1] = (char)(var10004 ^ var10005);
      }

      return new String(var10001).intern();
   }
}
