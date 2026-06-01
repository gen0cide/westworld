import java.awt.DisplayMode;
import java.awt.Frame;
import java.awt.GraphicsDevice;

final class ha {
   private DisplayMode a;
   private GraphicsDevice b;

   public final void exit() {
      // $VF: Couldn't be decompiled
      // Please report this to the Vineflower issue tracker, at https://github.com/Vineflower/vineflower/issues with a copy of the class file (if you have the rights to distribute it!)
      // java.lang.RuntimeException: parsing failure!
      //   at org.jetbrains.java.decompiler.modules.decompiler.decompose.DomHelper.parseGraph(DomHelper.java:211)
      //   at org.jetbrains.java.decompiler.main.rels.MethodProcessor.codeToJava(MethodProcessor.java:166)
      //
      // Bytecode:
      // 00: aconst_null
      // 01: aload 0
      // 02: getfield ha.a Ljava/awt/DisplayMode;
      // 05: if_acmpeq 3b
      // 08: aload 0
      // 09: getfield ha.b Ljava/awt/GraphicsDevice;
      // 0c: aload 0
      // 0d: getfield ha.a Ljava/awt/DisplayMode;
      // 10: invokevirtual java/awt/GraphicsDevice.setDisplayMode (Ljava/awt/DisplayMode;)V
      // 13: aload 0
      // 14: getfield ha.b Ljava/awt/GraphicsDevice;
      // 17: invokevirtual java/awt/GraphicsDevice.getDisplayMode ()Ljava/awt/DisplayMode;
      // 1a: aload 0
      // 1b: getfield ha.a Ljava/awt/DisplayMode;
      // 1e: invokevirtual java/awt/DisplayMode.equals (Ljava/awt/DisplayMode;)Z
      // 21: ifeq 2c
      // 24: goto 28
      // 27: athrow
      // 28: goto 36
      // 2b: athrow
      // 2c: new java/lang/RuntimeException
      // 2f: dup
      // 30: ldc ""
      // 32: invokespecial java/lang/RuntimeException.<init> (Ljava/lang/String;)V
      // 35: athrow
      // 36: aload 0
      // 37: aconst_null
      // 38: putfield ha.a Ljava/awt/DisplayMode;
      // 3b: aload 0
      // 3c: aconst_null
      // 3d: bipush 109
      // 3f: invokespecial ha.a (Ljava/awt/Frame;B)V
      // 42: goto 48
      // 45: astore 1
      // 46: aload 1
      // 47: athrow
      // 48: return
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
         int var3 = 0;

         try {
            while (~var1.length < ~var3) {
               var2[var3 << -186613982] = var1[var3].getWidth();
               var2[(var3 << -970847806) - -1] = var1[var3].getHeight();
               var2[(var3 << -921158878) - -2] = var1[var3].getBitDepth();
               var2[(var3 << -2140589790) - -3] = var1[var3].getRefreshRate();
               var3++;
            }
         } catch (RuntimeException var4) {
            throw var4;
         }

         return var2;
      } catch (RuntimeException var5) {
         throw var5;
      }
   }

   public ha() throws Exception {
      // $VF: Couldn't be decompiled
      // Please report this to the Vineflower issue tracker, at https://github.com/Vineflower/vineflower/issues with a copy of the class file (if you have the rights to distribute it!)
      // java.lang.RuntimeException: parsing failure!
      //   at org.jetbrains.java.decompiler.modules.decompiler.decompose.DomHelper.parseGraph(DomHelper.java:211)
      //   at org.jetbrains.java.decompiler.main.rels.MethodProcessor.codeToJava(MethodProcessor.java:166)
      //
      // Bytecode:
      // 00: aload 0
      // 01: invokespecial java/lang/Object.<init> ()V
      // 04: invokestatic java/awt/GraphicsEnvironment.getLocalGraphicsEnvironment ()Ljava/awt/GraphicsEnvironment;
      // 07: astore 1
      // 08: aload 0
      // 09: aload 1
      // 0a: invokevirtual java/awt/GraphicsEnvironment.getDefaultScreenDevice ()Ljava/awt/GraphicsDevice;
      // 0d: putfield ha.b Ljava/awt/GraphicsDevice;
      // 10: aload 0
      // 11: getfield ha.b Ljava/awt/GraphicsDevice;
      // 14: invokevirtual java/awt/GraphicsDevice.isFullScreenSupported ()Z
      // 17: ifne 5b
      // 1a: aload 1
      // 1b: invokevirtual java/awt/GraphicsEnvironment.getScreenDevices ()[Ljava/awt/GraphicsDevice;
      // 1e: astore 2
      // 1f: aload 2
      // 20: astore 3
      // 21: bipush 0
      // 22: istore 4
      // 24: aload 3
      // 25: arraylength
      // 26: iload 4
      // 28: if_icmple 53
      // 2b: aload 3
      // 2c: iload 4
      // 2e: aaload
      // 2f: astore 5
      // 31: aload 5
      // 33: ifnull 4d
      // 36: aload 5
      // 38: invokevirtual java/awt/GraphicsDevice.isFullScreenSupported ()Z
      // 3b: ifne 46
      // 3e: goto 42
      // 41: athrow
      // 42: goto 4d
      // 45: athrow
      // 46: aload 0
      // 47: aload 5
      // 49: putfield ha.b Ljava/awt/GraphicsDevice;
      // 4c: return
      // 4d: iinc 4 1
      // 50: goto 24
      // 53: new java/lang/Exception
      // 56: dup
      // 57: invokespecial java/lang/Exception.<init> ()V
      // 5a: athrow
      // 5b: goto 61
      // 5e: astore 1
      // 5f: aload 1
      // 60: athrow
      // 61: return
   }

   public final void enter(Frame param1, int param2, int param3, int param4, int param5) {
      // $VF: Couldn't be decompiled
      // Please report this to the Vineflower issue tracker, at https://github.com/Vineflower/vineflower/issues with a copy of the class file (if you have the rights to distribute it!)
      // java.lang.RuntimeException: parsing failure!
      //   at org.jetbrains.java.decompiler.modules.decompiler.decompose.DomHelper.parseGraph(DomHelper.java:211)
      //   at org.jetbrains.java.decompiler.main.rels.MethodProcessor.codeToJava(MethodProcessor.java:166)
      //
      // Bytecode:
      // 00: aload 0
      // 01: aload 0
      // 02: getfield ha.b Ljava/awt/GraphicsDevice;
      // 05: invokevirtual java/awt/GraphicsDevice.getDisplayMode ()Ljava/awt/DisplayMode;
      // 08: putfield ha.a Ljava/awt/DisplayMode;
      // 0b: aconst_null
      // 0c: aload 0
      // 0d: getfield ha.a Ljava/awt/DisplayMode;
      // 10: if_acmpeq 17
      // 13: goto 1f
      // 16: athrow
      // 17: new java/lang/NullPointerException
      // 1a: dup
      // 1b: invokespecial java/lang/NullPointerException.<init> ()V
      // 1e: athrow
      // 1f: aload 1
      // 20: bipush 1
      // 21: invokevirtual java/awt/Frame.setUndecorated (Z)V
      // 24: aload 1
      // 25: bipush 0
      // 26: invokevirtual java/awt/Frame.enableInputMethods (Z)V
      // 29: aload 0
      // 2a: aload 1
      // 2b: bipush -93
      // 2d: invokespecial ha.a (Ljava/awt/Frame;B)V
      // 30: iload 5
      // 32: ifeq 39
      // 35: goto dc
      // 38: athrow
      // 39: aload 0
      // 3a: getfield ha.a Ljava/awt/DisplayMode;
      // 3d: invokevirtual java/awt/DisplayMode.getRefreshRate ()I
      // 40: istore 6
      // 42: aload 0
      // 43: getfield ha.b Ljava/awt/GraphicsDevice;
      // 46: invokevirtual java/awt/GraphicsDevice.getDisplayModes ()[Ljava/awt/DisplayMode;
      // 49: astore 7
      // 4b: bipush 0
      // 4c: istore 8
      // 4e: bipush 0
      // 4f: istore 9
      // 51: iload 9
      // 53: aload 7
      // 55: arraylength
      // 56: if_icmpge cf
      // 59: iload 2
      // 5a: aload 7
      // 5c: iload 9
      // 5e: aaload
      // 5f: invokevirtual java/awt/DisplayMode.getWidth ()I
      // 62: if_icmpne c9
      // 65: goto 69
      // 68: athrow
      // 69: iload 3
      // 6a: bipush -1
      // 6b: ixor
      // 6c: aload 7
      // 6e: iload 9
      // 70: aaload
      // 71: invokevirtual java/awt/DisplayMode.getHeight ()I
      // 74: bipush -1
      // 75: ixor
      // 76: if_icmpne c9
      // 79: goto 7d
      // 7c: athrow
      // 7d: aload 7
      // 7f: iload 9
      // 81: aaload
      // 82: invokevirtual java/awt/DisplayMode.getBitDepth ()I
      // 85: bipush -1
      // 86: ixor
      // 87: iload 4
      // 89: bipush -1
      // 8a: ixor
      // 8b: if_icmpne c9
      // 8e: goto 92
      // 91: athrow
      // 92: aload 7
      // 94: iload 9
      // 96: aaload
      // 97: invokevirtual java/awt/DisplayMode.getRefreshRate ()I
      // 9a: istore 10
      // 9c: iload 8
      // 9e: ifeq c2
      // a1: iload 6
      // a3: ineg
      // a4: iload 10
      // a6: iadd
      // a7: invokestatic java/lang/Math.abs (I)I
      // aa: bipush -1
      // ab: ixor
      // ac: iload 6
      // ae: ineg
      // af: iload 5
      // b1: iadd
      // b2: invokestatic java/lang/Math.abs (I)I
      // b5: bipush -1
      // b6: ixor
      // b7: if_icmpgt c2
      // ba: goto be
      // bd: athrow
      // be: goto c9
      // c1: athrow
      // c2: iload 10
      // c4: istore 5
      // c6: bipush 1
      // c7: istore 8
      // c9: iinc 9 1
      // cc: goto 51
      // cf: iload 8
      // d1: ifeq d8
      // d4: goto dc
      // d7: athrow
      // d8: iload 6
      // da: istore 5
      // dc: aload 0
      // dd: getfield ha.b Ljava/awt/GraphicsDevice;
      // e0: new java/awt/DisplayMode
      // e3: dup
      // e4: iload 2
      // e5: iload 3
      // e6: iload 4
      // e8: iload 5
      // ea: invokespecial java/awt/DisplayMode.<init> (IIII)V
      // ed: invokevirtual java/awt/GraphicsDevice.setDisplayMode (Ljava/awt/DisplayMode;)V
      // f0: goto f8
      // f3: astore 6
      // f5: aload 6
      // f7: athrow
      // f8: return
   }
}
