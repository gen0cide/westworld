import java.awt.image.ImageConsumer;

final class u {
   static int e;
   static int f;
   static int[] a;
   static ImageConsumer d;
   static int c;
   static String[] b;
   static int g = 0;
   private static final String[] z = new String[]{
      z(z("\u0005#{\u001a")), z(z("\u000b#\u0016\u001c/")), z(z("\u001exT^")), z(z("\u0005#z\u001a")), z(z("\u0005#y\u001a"))
   };

   static final int a(int var0, tb var1, String var2) {
      try {
         if (var0 <= 10) {
            a = (int[])null;
         }

         f++;
         int var3 = var1.w;
         byte[] var4 = h.a(var2, (byte)30);
         var1.b(var4.length, (byte)-88);
         var1.w = var1.w + fb.a.a(var4.length, var1.F, var1.w, var4, 0, 119);
         return var1.w - var3;
      } catch (RuntimeException var5) {
         throw i.a(var5, z[3] + var0 + 44 + (var1 != null ? z[1] : z[2]) + 44 + (var2 != null ? z[1] : z[2]) + 41);
      }
   }

   static final i a(boolean var0, int var1) {
      boolean var5 = client.vh;

      try {
         c++;
         i[] var2 = gb.a(69);
         int var3 = 0;

         int var10000;
         while (true) {
            if (var2.length > var3) {
               i var4 = var2[var3];
               var10000 = var4.a;
               if (var5) {
                  break;
               }

               if (var4.a == var1) {
                  return var4;
               }

               var3++;
               if (!var5) {
                  continue;
               }
            }

            var10000 = var0;
            break;
         }

         if (var10000 != 0) {
            g = -2;
         }

         return null;
      } catch (RuntimeException var6) {
         throw i.a(var6, z[0] + var0 + ',' + var1 + ')');
      }
   }

   static final void a(int var0, long var1) {
      try {
         try {
            if (var0 != 0) {
               return;
            }

            Thread.sleep(var1);
         } catch (InterruptedException var4) {
         }

         e++;
      } catch (RuntimeException var5) {
         throw i.a(var5, z[4] + var0 + ',' + var1 + ')');
      }
   }

   private static char[] z(String var0) {
      char[] var10000 = var0.toCharArray();
      if (var10000.length < 2) {
         var10000[0] = (char)(var10000[0] ^ 82);
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
               var10005 = 112;
               break;
            case 1:
               var10005 = 13;
               break;
            case 2:
               var10005 = 56;
               break;
            case 3:
               var10005 = 50;
               break;
            default:
               var10005 = 82;
         }

         var10001[var1] = (char)(var10004 ^ var10005);
      }

      return new String(var10001).intern();
   }
}
