import java.awt.Component;
import java.awt.Point;
import java.awt.Robot;
import java.awt.image.BufferedImage;

final class j {
   private Robot a;
   private Component b;

   public final void showcursor(Component var1, boolean var2) {
      try {
         if (var2) {
            var1 = null;
         } else if (var1 == null) {
            throw new NullPointerException();
         }

         if (var1 != this.b) {
            if (this.b != null) {
               this.b.setCursor(null);
               this.b = null;
            }

            if (var1 != null) {
               var1.setCursor(var1.getToolkit().createCustomCursor(new BufferedImage(1, 1, 2), new Point(0, 0), null));
               this.b = var1;
            }
         }
      } catch (RuntimeException var4) {
         throw var4;
      }
   }

   public final void setcustomcursor(Component var1, int[] var2, int var3, int var4, Point var5) {
      try {
         if (null != var2) {
            BufferedImage var6 = new BufferedImage(var3, var4, 2);
            var6.setRGB(0, 0, var3, var4, var2, 0, var3);
            var1.setCursor(var1.getToolkit().createCustomCursor(var6, var5, null));
         } else {
            var1.setCursor(null);
         }
      } catch (RuntimeException var7) {
         throw var7;
      }
   }

   public final void movemouse(int var1, int var2) {
      try {
         this.a.mouseMove(var1, var2);
      } catch (RuntimeException var4) {
         throw var4;
      }
   }

   public j() throws Exception {
      try {
         this.a = new Robot();
      } catch (RuntimeException var2) {
         throw var2;
      }
   }
}
