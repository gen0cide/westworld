import java.awt.AWTEvent;
import java.awt.Frame;
import java.awt.Graphics;
import java.awt.event.MouseEvent;

final class qb extends Frame {
   static int b;
   static int[] e;
   static byte[] k = new byte[100000];
   private e g;
   static int f;
   static int[][] d;
   static int i;
   private int c;
   private int l;
   private int j = 28;
   private int h;
   static int a;
   private static final String[] z = new String[]{
      z(z(":\u001am\u0006\u000b")),
      z(z("0Vm\u0014\u001f/]7\u0016^")),
      z(z("/A/D")),
      z(z("0VmZ\u00132]9M^")),
      z(z("0VmO\u00135s1I\u0006)] [^h")),
      z(z("0VmX\u0004.W&[\u0005\u0004B&F\u0002i")),
      z(z("0VmX\u0017(Z7\u0000"))
   };

   @Override
   protected final void processEvent(AWTEvent var1) {
      try {
         if (var1 instanceof MouseEvent) {
            MouseEvent var2 = (MouseEvent)var1;
            var1 = new MouseEvent(
               var2.getComponent(),
               var2.getID(),
               var2.getWhen(),
               var2.getModifiers(),
               var2.getX(),
               var2.getY() - 24,
               var2.getClickCount(),
               var2.isPopupTrigger()
            );
         }

         a++;
         super.processEvent(var1);
      } catch (RuntimeException var3) {
         throw i.a(var3, z[5] + (var1 != null ? z[0] : z[2]) + ')');
      }
   }

   @Override
   public final void paint(Graphics var1) {
      try {
         b++;
         this.g.paint(var1);
      } catch (RuntimeException var3) {
         throw i.a(var3, z[6] + (var1 != null ? z[0] : z[2]) + ')');
      }
   }

   @Override
   public final void resize(int var1, int var2) {
      try {
         super.resize(var1, this.j + var2);
         i++;
      } catch (RuntimeException var4) {
         throw i.a(var4, z[3] + var1 + ',' + var2 + ')');
      }
   }

   @Override
   public final Graphics getGraphics() {
      try {
         f++;
         Graphics var1 = super.getGraphics();
         if (this.c != 0) {
            var1.translate(-5, 0);
            if (!client.vh) {
               return var1;
            }
         }

         var1.translate(0, 24);
         return var1;
      } catch (RuntimeException var2) {
         throw i.a(var2, z[4]);
      }
   }

   qb(e var1, int var2, int var3, String var4, boolean var5, boolean var6) {
      this.c = 0;

      try {
         label32: {
            this.g = var1;
            if (!var6) {
               this.j = 28;
               if (!client.vh) {
                  break label32;
               }
            }

            this.j = 48;
         }

         this.l = var2;
         this.h = var3;
         this.setTitle(var4);
         this.setResizable(var5);
         this.show();
         this.toFront();
         this.resize(this.l, this.h);
         this.getGraphics();
      } catch (RuntimeException var8) {
         throw i.a(var8, z[1] + (var1 != null ? z[0] : z[2]) + ',' + var2 + ',' + var3 + ',' + (var4 != null ? z[0] : z[2]) + ',' + var5 + ',' + var6 + ')');
      }
   }

   private static char[] z(String var0) {
      char[] var10000 = var0.toCharArray();
      if (var10000.length < 2) {
         var10000[0] = (char)(var10000[0] ^ 118);
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
               var10005 = 65;
               break;
            case 1:
               var10005 = 52;
               break;
            case 2:
               var10005 = 67;
               break;
            case 3:
               var10005 = 40;
               break;
            default:
               var10005 = 118;
         }

         var10001[var1] = (char)(var10004 ^ var10005);
      }

      return new String(var10001).intern();
   }
}
