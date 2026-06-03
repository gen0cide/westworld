import java.awt.Image;
import java.awt.image.ImageConsumer;
import java.awt.image.ImageObserver;
import java.awt.image.ImageProducer;

final class fb implements ImageProducer, ImageObserver {
   static i h = new i(z(z("\rM")), 1);
   static int g;
   static aa a;
   static int[] f;
   static int i;
   static int b;
   static int l;
   static int j;
   static int e;
   static int[] d;
   static int[] c;
   static boolean[] k = new boolean[]{false, false, false, false, false, false, false, false, false, false, false, false};
   private static final String[] z = new String[]{
      z(z("1{3R")),
      z(z("$ q\u00107")),
      z(z("9lqW'>i:k:;o+[b")),
      z(z("9lqL/2a)[\t0`,K':|w")),
      z(z("9lqL/.{:M>\u000ba/z%(`\u0013[,+\\6Y\"+\\:M/1jw")),
      z(z("9lqW9\u001ca1M?2k-\u0016")),
      z(z("9lq_.;M0P9*c:Lb")),
      z(z("9lqM>>|+n80j*]>6a1\u0016"))
   };

   @Override
   public final synchronized void addConsumer(ImageConsumer var1) {
      try {
         u.d = var1;
         e++;
         var1.setDimensions(k.o, da.bb);
         var1.setProperties(null);
         var1.setColorModel(m.d);
         var1.setHints(14);
      } catch (RuntimeException var3) {
         throw i.a(var3, z[6] + (var1 != null ? z[1] : z[0]) + ')');
      }
   }

   @Override
   public final void startProduction(ImageConsumer var1) {
      try {
         i++;
         this.addConsumer(var1);
      } catch (RuntimeException var3) {
         throw i.a(var3, z[7] + (var1 != null ? z[1] : z[0]) + ')');
      }
   }

   @Override
   public final void requestTopDownLeftRightResend(ImageConsumer var1) {
      try {
         j++;
      } catch (RuntimeException var3) {
         throw i.a(var3, z[4] + (var1 != null ? z[1] : z[0]) + ')');
      }
   }

   @Override
   public final synchronized void removeConsumer(ImageConsumer var1) {
      try {
         g++;
         if (u.d == var1) {
            u.d = null;
         }
      } catch (RuntimeException var3) {
         throw i.a(var3, z[3] + (var1 != null ? z[1] : z[0]) + ')');
      }
   }

   @Override
   public final synchronized boolean isConsumer(ImageConsumer var1) {
      try {
         b++;
         return var1 == u.d;
      } catch (RuntimeException var3) {
         throw i.a(var3, z[5] + (var1 != null ? z[1] : z[0]) + ')');
      }
   }

   @Override
   public final boolean imageUpdate(Image var1, int var2, int var3, int var4, int var5, int var6) {
      try {
         l++;
         return true;
      } catch (RuntimeException var8) {
         throw i.a(var8, z[2] + (var1 != null ? z[1] : z[0]) + ',' + var2 + ',' + var3 + ',' + var4 + ',' + var5 + ',' + var6 + ')');
      }
   }

   private static char[] z(String var0) {
      char[] var10000 = var0.toCharArray();
      if (var10000.length < 2) {
         var10000[0] = (char)(var10000[0] ^ 74);
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
               var10005 = 95;
               break;
            case 1:
               var10005 = 14;
               break;
            case 2:
               var10005 = 95;
               break;
            case 3:
               var10005 = 62;
               break;
            default:
               var10005 = 74;
         }

         var10001[var1] = (char)(var10004 ^ var10005);
      }

      return new String(var10001).intern();
   }
}
