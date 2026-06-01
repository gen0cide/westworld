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
      } catch (RuntimeException var4) {
         RuntimeException var2 = var4;

         RuntimeException var10000;
         StringBuilder var10001;
         try {
            var10000 = var2;
            var10001 = new StringBuilder().append(z[6]);
            if (var1 != null) {
               throw i.a(var2, var10001.append(z[1]).append(')').toString());
            }
         } catch (RuntimeException var3) {
            throw var3;
         }

         throw i.a(var10000, var10001.append(z[0]).append(')').toString());
      }
   }

   @Override
   public final void startProduction(ImageConsumer var1) {
      try {
         i++;
         this.addConsumer(var1);
      } catch (RuntimeException var4) {
         RuntimeException var2 = var4;

         RuntimeException var10000;
         StringBuilder var10001;
         try {
            var10000 = var2;
            var10001 = new StringBuilder().append(z[7]);
            if (var1 != null) {
               throw i.a(var2, var10001.append(z[1]).append(')').toString());
            }
         } catch (RuntimeException var3) {
            throw var3;
         }

         throw i.a(var10000, var10001.append(z[0]).append(')').toString());
      }
   }

   @Override
   public final void requestTopDownLeftRightResend(ImageConsumer var1) {
      try {
         j++;
      } catch (RuntimeException var4) {
         RuntimeException var2 = var4;

         RuntimeException var10000;
         StringBuilder var10001;
         try {
            var10000 = var2;
            var10001 = new StringBuilder().append(z[4]);
            if (var1 != null) {
               throw i.a(var2, var10001.append(z[1]).append(')').toString());
            }
         } catch (RuntimeException var3) {
            throw var3;
         }

         throw i.a(var10000, var10001.append(z[0]).append(')').toString());
      }
   }

   @Override
   public final synchronized void removeConsumer(ImageConsumer var1) {
      try {
         try {
            g++;
            if (u.d != var1) {
               return;
            }
         } catch (RuntimeException var4) {
            throw var4;
         }

         u.d = null;
      } catch (RuntimeException var5) {
         RuntimeException var2 = var5;

         RuntimeException var10000;
         StringBuilder var10001;
         try {
            var10000 = var2;
            var10001 = new StringBuilder().append(z[3]);
            if (var1 != null) {
               throw i.a(var2, var10001.append(z[1]).append(')').toString());
            }
         } catch (RuntimeException var3) {
            throw var3;
         }

         throw i.a(var10000, var10001.append(z[0]).append(')').toString());
      }
   }

   @Override
   public final synchronized boolean isConsumer(ImageConsumer var1) {
      try {
         try {
            b++;
            if (var1 == u.d) {
               return true;
            }
         } catch (RuntimeException var4) {
            throw var4;
         }

         return false;
      } catch (RuntimeException var5) {
         RuntimeException var2 = var5;

         RuntimeException var10000;
         StringBuilder var10001;
         try {
            var10000 = var2;
            var10001 = new StringBuilder().append(z[5]);
            if (var1 != null) {
               throw i.a(var2, var10001.append(z[1]).append(')').toString());
            }
         } catch (RuntimeException var3) {
            throw var3;
         }

         throw i.a(var10000, var10001.append(z[0]).append(')').toString());
      }
   }

   @Override
   public final boolean imageUpdate(Image var1, int var2, int var3, int var4, int var5, int var6) {
      try {
         l++;
         return true;
      } catch (RuntimeException var9) {
         RuntimeException var7 = var9;

         RuntimeException var10000;
         StringBuilder var10001;
         try {
            var10000 = var7;
            var10001 = new StringBuilder().append(z[2]);
            if (var1 != null) {
               throw i.a(
                  var7,
                  var10001.append(z[1])
                     .append(',')
                     .append(var2)
                     .append(',')
                     .append(var3)
                     .append(',')
                     .append(var4)
                     .append(',')
                     .append(var5)
                     .append(',')
                     .append(var6)
                     .append(')')
                     .toString()
               );
            }
         } catch (RuntimeException var8) {
            throw var8;
         }

         throw i.a(
            var10000,
            var10001.append(z[0])
               .append(',')
               .append(var2)
               .append(',')
               .append(var3)
               .append(',')
               .append(var4)
               .append(',')
               .append(var5)
               .append(',')
               .append(var6)
               .append(')')
               .toString()
         );
      }
   }

   private static char[] z(String var0) {
      char[] var10000 = var0.toCharArray();
      if (var10000.length < 2) {
         var10000[0] = (char)(var10000[0] ^ 'J');
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
