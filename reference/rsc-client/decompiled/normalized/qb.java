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
         label32: {
            try {
               if (!(var1 instanceof MouseEvent)) {
                  break label32;
               }
            } catch (RuntimeException var4) {
               throw var4;
            }

            MouseEvent var6 = (MouseEvent)var1;
            var1 = new MouseEvent(
               var6.getComponent(),
               var6.getID(),
               var6.getWhen(),
               var6.getModifiers(),
               var6.getX(),
               var6.getY() - 24,
               var6.getClickCount(),
               var6.isPopupTrigger()
            );
         }

         a++;
         super.processEvent((AWTEvent)var1);
      } catch (RuntimeException var5) {
         RuntimeException var2 = var5;

         RuntimeException var10000;
         StringBuilder var10001;
         try {
            var10000 = var2;
            var10001 = new StringBuilder().append(z[5]);
            if (var1 != null) {
               throw i.a(var2, var10001.append(z[0]).append(')').toString());
            }
         } catch (RuntimeException var3) {
            throw var3;
         }

         throw i.a(var10000, var10001.append(z[2]).append(')').toString());
      }
   }

   @Override
   public final void paint(Graphics var1) {
      try {
         b++;
         this.g.paint(var1);
      } catch (RuntimeException var4) {
         RuntimeException var2 = var4;

         RuntimeException var10000;
         StringBuilder var10001;
         try {
            var10000 = var2;
            var10001 = new StringBuilder().append(z[6]);
            if (var1 != null) {
               throw i.a(var2, var10001.append(z[0]).append(')').toString());
            }
         } catch (RuntimeException var3) {
            throw var3;
         }

         throw i.a(var10000, var10001.append(z[2]).append(')').toString());
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
      // $VF: Couldn't be decompiled
      // Please report this to the Vineflower issue tracker, at https://github.com/Vineflower/vineflower/issues with a copy of the class file (if you have the rights to distribute it!)
      // java.lang.RuntimeException: parsing failure!
      //   at org.jetbrains.java.decompiler.modules.decompiler.decompose.DomHelper.parseGraph(DomHelper.java:211)
      //   at org.jetbrains.java.decompiler.main.rels.MethodProcessor.codeToJava(MethodProcessor.java:166)
      //
      // Bytecode:
      // 00: getstatic qb.f I
      // 03: bipush 1
      // 04: iadd
      // 05: putstatic qb.f I
      // 08: aload 0
      // 09: invokespecial java/awt/Frame.getGraphics ()Ljava/awt/Graphics;
      // 0c: astore 1
      // 0d: aload 0
      // 0e: getfield qb.c I
      // 11: ifeq 25
      // 14: aload 1
      // 15: bipush -5
      // 17: bipush 0
      // 18: invokevirtual java/awt/Graphics.translate (II)V
      // 1b: getstatic client.vh Z
      // 1e: ifeq 30
      // 21: goto 25
      // 24: athrow
      // 25: aload 1
      // 26: bipush 0
      // 27: bipush 24
      // 29: invokevirtual java/awt/Graphics.translate (II)V
      // 2c: goto 30
      // 2f: athrow
      // 30: aload 1
      // 31: areturn
      // 32: astore 1
      // 33: aload 1
      // 34: getstatic qb.z [Ljava/lang/String;
      // 37: bipush 4
      // 38: aaload
      // 39: invokestatic i.a (Ljava/lang/Throwable;Ljava/lang/String;)Lla;
      // 3c: athrow
   }

   // $VF: Handled exception range with multiple entry points by splitting it
   // $VF: Inserted dummy exception handlers to handle obfuscated exceptions
   qb(e var1, int var2, int var3, String var4, boolean var5, boolean var6) {
      this.c = 0;

      try {
         label62: {
            this.g = var1;
            if (!var6) {
               try {
                  this.j = 28;
                  if (!client.vh) {
                     break label62;
                  }
               } catch (RuntimeException var11) {
                  boolean var15 = false;
                  throw var11;
               }
            }

            try {
               this.j = 48;
            } catch (RuntimeException var10) {
               boolean var16 = false;
               throw var10;
            }
         }

         this.l = var2;
         this.h = var3;
         this.setTitle(var4);
         this.setResizable(var5);
         this.show();
         this.toFront();
         this.resize(this.l, this.h);
         this.getGraphics();
      } catch (RuntimeException var12) {
         RuntimeException var7 = var12;

         RuntimeException var10000;
         StringBuilder var10001;
         String var10002;
         label42: {
            try {
               var10000 = var7;
               var10001 = new StringBuilder().append(z[1]);
               if (var1 != null) {
                  var10002 = z[0];
                  break label42;
               }
            } catch (RuntimeException var9) {
               throw var9;
            }

            var10002 = z[2];
         }

         try {
            var10001 = var10001.append(var10002).append(',').append(var2).append(',').append(var3).append(',');
            if (var4 != null) {
               throw i.a(var10000, var10001.append(z[0]).append(',').append(var5).append(',').append(var6).append(')').toString());
            }
         } catch (RuntimeException var8) {
            throw var8;
         }

         throw i.a(var10000, var10001.append(z[2]).append(',').append(var5).append(',').append(var6).append(')').toString());
      }
   }

   private static char[] z(String var0) {
      char[] var10000 = var0.toCharArray();
      if (var10000.length < 2) {
         var10000[0] = (char)(var10000[0] ^ 'v');
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
