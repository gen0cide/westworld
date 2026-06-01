import com.ms.awt.WComponentPeer;
import com.ms.dll.Callback;
import com.ms.dll.Root;
import com.ms.win32.User32;
import java.awt.Component;

final class q extends Callback {
   private volatile int d;
   private volatile int c;
   private int e;
   private boolean a;
   private volatile boolean b = true;

   final void a(int var1, int var2, int var3) {
      try {
         if (var1 != 23529) {
            this.callback(-56, 122, 69, -57);
         }

         User32.SetCursorPos(var3, var2);
      } catch (RuntimeException var5) {
         throw var5;
      }
   }

   final void a(int var1, Component var2, boolean var3) {
      try {
         WComponentPeer var4 = (WComponentPeer)var2.getPeer();
         int var5 = var4.getTopHwnd();
         if (~var5 != ~this.d || !this.b == var3) {
            if (var1 != -4) {
               this.d = -1;
            }

            if (!this.a) {
               this.e = User32.LoadCursor(0, 32512);
               Root.alloc(this);
               this.a = true;
            }

            if (var5 != this.d) {
               if (this.d != 0) {
                  this.b = true;
                  User32.SendMessage(var5, 101024, 0, 0);
                  synchronized (this) {
                     User32.SetWindowLong(this.d, -4, this.c);
                  }
               }

               synchronized (this) {
                  this.d = var5;
                  this.c = User32.SetWindowLong(this.d, -4, this);
               }
            }

            this.b = var3;
            User32.SendMessage(var5, 101024, 0, 0);
         }
      } catch (RuntimeException var11) {
         throw var11;
      }
   }

   final synchronized int callback(int var1, int var2, int var3, int var4) {
      try {
         if (~this.d != ~var1) {
            int var7 = User32.GetWindowLong(var1, -4);
            return User32.CallWindowProc(var7, var1, var2, var3, var4);
         }

         if (-33 == ~var2) {
            int var5 = 65535 & var4;
            if (1 == var5) {
               User32.SetCursor(this.b ? this.e : 0);
               return 0;
            }
         }

         if (-101025 != ~var2) {
            if (-2 == ~var2) {
               this.d = 0;
               this.b = true;
            }

            return User32.CallWindowProc(this.c, var1, var2, var3, var4);
         } else {
            User32.SetCursor(this.b ? this.e : 0);
            return 0;
         }
      } catch (RuntimeException var6) {
         throw var6;
      }
   }
}
