import java.awt.DisplayMode;
import java.awt.Frame;
import java.awt.GraphicsDevice;
import java.awt.GraphicsEnvironment;

final class ha {
   private DisplayMode a;
   private GraphicsDevice b;

   public final void exit() {
      try {
         if (null != this.a) {
            this.b.setDisplayMode(this.a);
            if (!this.b.getDisplayMode().equals(this.a)) {
               throw new RuntimeException("");
            }

            this.a = null;
         }

         this.a(null, (byte)109);
      } catch (RuntimeException var2) {
         throw var2;
      }
   }

   private final void a(Frame var1, byte var2) {
      try {
         try {
            this.b.setFullScreenWindow(var1);
            int var3 = -113 % ((var2 - -39) / 47);
         } finally {
            ;
         }
      } catch (RuntimeException var8) {
         throw var8;
      }
   }

   public final int[] listmodes() {
      try {
         DisplayMode[] var1 = this.b.getDisplayModes();
         int[] var2 = new int[var1.length << -1704625342];

         for (int var3 = 0; ~var1.length < ~var3; var3++) {
            var2[var3 << -186613982] = var1[var3].getWidth();
            var2[(var3 << -970847806) - -1] = var1[var3].getHeight();
            var2[(var3 << -921158878) - -2] = var1[var3].getBitDepth();
            var2[(var3 << -2140589790) - -3] = var1[var3].getRefreshRate();
         }

         return var2;
      } catch (RuntimeException var4) {
         throw var4;
      }
   }

   public ha() throws Exception {
      try {
         GraphicsEnvironment var1 = GraphicsEnvironment.getLocalGraphicsEnvironment();
         this.b = var1.getDefaultScreenDevice();
         if (!this.b.isFullScreenSupported()) {
            GraphicsDevice[] var2 = var1.getScreenDevices();
            GraphicsDevice[] var3 = var2;

            for (int var4 = 0; var3.length > var4; var4++) {
               GraphicsDevice var5 = var3[var4];
               if (var5 != null && var5.isFullScreenSupported()) {
                  this.b = var5;
                  return;
               }
            }

            throw new Exception();
         }
      } catch (RuntimeException var6) {
         throw var6;
      }
   }

   public final void enter(Frame var1, int var2, int var3, int var4, int var5) {
      try {
         this.a = this.b.getDisplayMode();
         if (null == this.a) {
            throw new NullPointerException();
         }

         var1.setUndecorated(true);
         var1.enableInputMethods(false);
         this.a(var1, (byte)-93);
         if (var5 == 0) {
            int var6 = this.a.getRefreshRate();
            DisplayMode[] var7 = this.b.getDisplayModes();
            boolean var8 = false;

            for (int var9 = 0; var9 < var7.length; var9++) {
               if (var2 == var7[var9].getWidth() && ~var3 == ~var7[var9].getHeight() && ~var7[var9].getBitDepth() == ~var4) {
                  int var10 = var7[var9].getRefreshRate();
                  if (!var8 || ~Math.abs(-var6 + var10) > ~Math.abs(-var6 + var5)) {
                     var5 = var10;
                     var8 = true;
                  }
               }
            }

            if (!var8) {
               var5 = var6;
            }
         }

         this.b.setDisplayMode(new DisplayMode(var2, var3, var4, var5));
      } catch (RuntimeException var11) {
         throw var11;
      }
   }
}
