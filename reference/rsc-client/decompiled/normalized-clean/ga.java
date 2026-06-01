final class ga {
   static v c = new v(z(z("\u0005D\\F")), z(z("=vknw7")), z(z("\ral")), 2);
   static int d;
   static char[] a = new char[]{
      ' ',
      ' ',
      '_',
      '-',
      'à',
      'á',
      'â',
      'ä',
      'ã',
      'À',
      'Á',
      'Â',
      'Ä',
      'Ã',
      'è',
      'é',
      'ê',
      'ë',
      'È',
      'É',
      'Ê',
      'Ë',
      'í',
      'î',
      'ï',
      'Í',
      'Î',
      'Ï',
      'ò',
      'ó',
      'ô',
      'ö',
      'õ',
      'Ò',
      'Ó',
      'Ô',
      'Ö',
      'Õ',
      'ù',
      'ú',
      'û',
      'ü',
      'Ù',
      'Ú',
      'Û',
      'Ü',
      'ç',
      'Ç',
      'ÿ',
      'Ÿ',
      'ñ',
      'Ñ',
      'ß'
   };
   static String[] b;
   private static final String[] z = new String[]{z(z("<eak")), z(z(")>#)i")), z(z("5q#F<"))};

   static final String a(int var0, int var1, int var2, byte[] var3) {
      boolean var9 = client.vh;

      try {
         d++;
         char[] var4 = new char[var0];
         int var5 = 0;
         int var6 = 0;

         byte var10000;
         int var10001;
         while (true) {
            if (~var0 < ~var6) {
               int var7 = 255 & var3[var2 - -var6];
               var10000 = 0;
               var10001 = var7;
               if (var9) {
                  break;
               }

               if (0 != var7 || var9) {
                  if (~var7 <= -129 && ~var7 > -161) {
                     char var8 = nb.f[var7 - 128];
                     if (~var8 == -1) {
                        var8 = '?';
                     }

                     var7 = var8;
                  }

                  var4[var5++] = (char)var7;
               }

               var6++;
               if (!var9) {
                  continue;
               }
            }

            var10000 = -103;
            var10001 = (-63 - var1) / 49;
            break;
         }

         var6 = var10000 / var10001;
         return new String(var4, 0, var5);
      } catch (RuntimeException var10) {
         throw i.a(var10, z[2] + var0 + ',' + var1 + ',' + var2 + ',' + (var3 != null ? z[1] : z[0]) + ')');
      }
   }

   private static char[] z(String var0) {
      char[] var10000 = var0.toCharArray();
      if (var10000.length < 2) {
         var10000[0] = (char)(var10000[0] ^ 20);
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
               var10005 = 82;
               break;
            case 1:
               var10005 = 16;
               break;
            case 2:
               var10005 = 13;
               break;
            case 3:
               var10005 = 7;
               break;
            default:
               var10005 = 20;
         }

         var10001[var1] = (char)(var10004 ^ var10005);
      }

      return new String(var10001).intern();
   }
}
