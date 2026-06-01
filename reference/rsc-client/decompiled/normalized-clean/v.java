final class v {
   static int c;
   int i;
   static int[] a;
   static int f;
   static int h;
   static int d;
   static int b;
   static int[] g;
   static int[] e;
   private static final String[] z = new String[]{
      z(z("\"tv%")), z(z(":/& /%u$a")), z(z("7/4g<")), z(z(":/Ya")), z(z(":/[a")), z(z(":/n&\u00128ss'&d(")), z(z(":/Xa"))
   };

   static final byte[] a(byte[] var0, int var1, int var2, int var3) {
      boolean var6 = client.vh;

      try {
         c++;
         if (var2 != -98) {
            a = (int[])null;
         }

         byte[] var4 = new byte[var1];
         int var5 = 0;

         byte[] var10000;
         while (true) {
            if (var5 < var1) {
               var10000 = var4;
               if (var6) {
                  break;
               }

               var4[var5] = qa.l[ib.a(var0[var3 + var5], 255)];
               var5++;
               if (!var6) {
                  continue;
               }
            }

            var10000 = var4;
            break;
         }

         return var10000;
      } catch (RuntimeException var7) {
         throw i.a(var7, z[3] + (var0 != null ? z[2] : z[0]) + ',' + var1 + ',' + var2 + ',' + var3 + ')');
      }
   }

   static final int a(int var0) {
      try {
         d++;
         int var1 = b.v[ka.b] & 255;
         ka.b++;
         if (var0 != -30504) {
            a(113);
         }

         return var1;
      } catch (RuntimeException var2) {
         throw i.a(var2, z[4] + var0 + 41);
      }
   }

   @Override
   public final String toString() {
      try {
         f++;
         throw new IllegalStateException();
      } catch (RuntimeException var2) {
         throw i.a(var2, z[5]);
      }
   }

   static final boolean a(char var0, int var1) {
      try {
         b++;
         if (var1 <= 111) {
            a((byte[])null, 51, 127, 27);
         }

         return ~var0 <= -49 && ~var0 >= -58 || var0 >= 'A' && ~var0 >= -91 || -98 >= ~var0 && -123 <= ~var0;
      } catch (RuntimeException var3) {
         throw i.a(var3, z[6] + var0 + ',' + var1 + ')');
      }
   }

   v(String var1, String var2, String var3, int var4) {
      try {
         this.i = var4;
      } catch (RuntimeException var6) {
         throw i.a(var6, z[1] + (var1 != null ? z[2] : z[0]) + ',' + (var2 != null ? z[2] : z[0]) + ',' + (var3 != null ? z[2] : z[0]) + ',' + var4 + ')');
      }
   }

   private static char[] z(String var0) {
      char[] var10000 = var0.toCharArray();
      if (var10000.length < 2) {
         var10000[0] = (char)(var10000[0] ^ 65);
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
               var10005 = 76;
               break;
            case 1:
               var10005 = 1;
               break;
            case 2:
               var10005 = 26;
               break;
            case 3:
               var10005 = 73;
               break;
            default:
               var10005 = 65;
         }

         var10001[var1] = (char)(var10004 ^ var10005);
      }

      return new String(var10001).intern();
   }
}
