import com.ms.awt.WComponentPeer;
import com.ms.com.IUnknown;
import com.ms.directX.DDSurfaceDesc;
import com.ms.directX.DirectDraw;
import com.ms.directX.IEnumModesCallback;
import com.ms.win32.User32;
import java.awt.Frame;

final class wa implements IEnumModesCallback {
   private static int[] b;
   private DirectDraw c;
   private static int a;

   public final void callbackEnumModes(DDSurfaceDesc var1, IUnknown var2) {
      try {
         if (null == b) {
            a += 4;
         } else {
            b[a++] = var1.width;
            b[a++] = var1.height;
            b[a++] = var1.rgbBitCount;
            b[a++] = var1.refreshRate;
         }
      } catch (RuntimeException var4) {
         throw var4;
      }
   }

   final int[] a(byte var1) {
      try {
         this.c.enumDisplayModes(0, null, null, this);
         b = new int[a];
         a = 0;
         this.c.enumDisplayModes(0, null, null, this);
         int[] var2 = b;
         b = null;
         a = 0;
         int var3 = 55 % ((var1 - 22) / 53);
         return var2;
      } catch (RuntimeException var4) {
         throw var4;
      }
   }

   final void a(Frame var1, int var2) {
      try {
         if (var2 != 0) {
            this.a((Frame)null, 68, -102, 21, 73, (byte)97);
         }

         this.c.restoreDisplayMode();
         this.c.setCooperativeLevel(var1, 8);
      } catch (RuntimeException var4) {
         throw var4;
      }
   }

   final void a(Frame var1, int var2, int var3, int var4, int var5, byte var6) {
      try {
         if (var6 != 77) {
            this.c = (DirectDraw)null;
         }

         var1.setVisible(true);
         WComponentPeer var7 = (WComponentPeer)var1.getPeer();
         int var8 = var7.getHwnd();
         User32.SetWindowLong(var8, -16, Integer.MIN_VALUE);
         User32.SetWindowLong(var8, -20, 8);
         this.c.setCooperativeLevel(var1, 17);
         this.c.setDisplayMode(var5, var4, var2, var3, 0);
         var1.setBounds(0, 0, var5, var4);
         var1.toFront();
         var1.requestFocus();
      } catch (RuntimeException var9) {
         throw var9;
      }
   }

   public wa() {
      try {
         this.c = new DirectDraw();
         this.c.initialize(null);
      } catch (RuntimeException var2) {
         throw var2;
      }
   }
}
