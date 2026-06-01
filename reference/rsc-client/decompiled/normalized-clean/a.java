import java.applet.Applet;
import netscape.javascript.JSObject;

final class a {
   static final Object a(String var0, byte var1, Applet var2) throws Throwable {
      try {
         int var3 = -38 % ((var1 - 14) / 53);
         return JSObject.getWindow(var2).call(var0, null);
      } catch (RuntimeException var4) {
         throw var4;
      }
   }
}
