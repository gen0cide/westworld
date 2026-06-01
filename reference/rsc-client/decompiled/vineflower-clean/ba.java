import java.awt.Component;

final class ba extends ua {
   client dc;
   static int bc;
   static int Zb;
   static int[] cc = new int[2048];
   static String[] Yb = new String[100];
   static int ec;
   static String[] ac;
   private static final String[] fc = new String[]{
      z(z("\t\u0013\u0007\u001c{")),
      z(z("\u0003\u0013\u0006\u001d")),
      z(z("\u0016HD_\u007f")),
      z(z("\u000f\u0007D0*")),
      z(z("\u000f\u0007D3*")),
      z(z("\u000f\u0007D2*"))
   };

   static final String e(int var0, int var1) {
      try {
         int var2 = 109 % ((var0 - -8) / 63);
         bc++;
         return (var1 >> 2086116408 & 0xFF) + "." + (0xFF & var1 >> 968554992) + "." + (var1 >> 1312213096 & 0xFF) + "." + (var1 & 0xFF);
      } catch (RuntimeException var3) {
         throw i.a(var3, fc[5] + var0 + ',' + var1 + ')');
      }
   }

   ba(int var1, int var2, int var3, Component var4) {
      super(var1, var2, var3, var4);
   }

   @Override
   final void a(int var1, int var2, int var3, int var4, int var5, int var6, byte var7, int var8) {
      boolean var10 = client.vh;

      try {
         label51: {
            ec++;
            if (var2 < 50000) {
               if (40000 > var2) {
                  if (var2 >= 20000) {
                     this.dc.a(var5, var8, 105, var3, var1, -20000 + var2, var6, var4);
                     if (!var10) {
                        break label51;
                     }
                  }

                  if (~var2 > -5001) {
                     super.f(var4, var5, var3, var6, 5924, var2);
                     if (!var10) {
                        break label51;
                     }
                  }

                  this.dc.b(var8, var6, var7 ^ 9, var1, var4, var5, var3, -5000 + var2);
                  if (!var10) {
                     break label51;
                  }
               }

               this.dc.b(var3, var8, var4, var2 - 40000, var6, -122, var5);
               if (!var10) {
                  break label51;
               }
            }

            this.dc.a(var8, var4, var5, var3, var2 - 50000, var6, 2);
         }

         if (var7 != 29) {
            ac = (String[])null;
         }
      } catch (RuntimeException var11) {
         throw i.a(var11, fc[4] + var1 + ',' + var2 + ',' + var3 + ',' + var4 + ',' + var5 + ',' + var6 + ',' + var7 + ',' + var8 + ')');
      }
   }

   static final void a(Object param0, int param1, c param2) {
      // $VF: Couldn't be decompiled
      // Please report this to the Vineflower issue tracker, at https://github.com/Vineflower/vineflower/issues with a copy of the class file (if you have the rights to distribute it!)
      // java.lang.RuntimeException: parsing failure!
      //   at org.jetbrains.java.decompiler.modules.decompiler.decompose.DomHelper.parseGraph(DomHelper.java:235)
      //   at org.jetbrains.java.decompiler.main.rels.MethodProcessor.codeToJava(MethodProcessor.java:174)
      //
      // Bytecode:
      // 00: getstatic client.vh Z
      // 03: istore 4
      // 05: getstatic ba.Zb I
      // 08: bipush 1
      // 09: iadd
      // 0a: putstatic ba.Zb I
      // 0d: aconst_null
      // 0e: aload 2
      // 0f: getfield c.n Ljava/awt/EventQueue;
      // 12: if_acmpeq 18
      // 15: goto 19
      // 18: return
      // 19: bipush 0
      // 1a: istore 3
      // 1b: bipush -51
      // 1d: iload 3
      // 1e: bipush -1
      // 1f: ixor
      // 20: if_icmpge 49
      // 23: aload 2
      // 24: getfield c.n Ljava/awt/EventQueue;
      // 27: invokevirtual java/awt/EventQueue.peekEvent ()Ljava/awt/AWTEvent;
      // 2a: iload 4
      // 2c: ifne 50
      // 2f: ifnull 49
      // 32: goto 36
      // 35: athrow
      // 36: sipush 11200
      // 39: lconst_1
      // 3a: invokestatic mb.a (IJ)V
      // 3d: iinc 3 1
      // 40: iload 4
      // 42: ifeq 1b
      // 45: goto 49
      // 48: athrow
      // 49: iload 1
      // 4a: bipush 1
      // 4b: if_icmpeq 4f
      // 4e: return
      // 4f: aload 0
      // 50: ifnull 6e
      // 53: aload 2
      // 54: getfield c.n Ljava/awt/EventQueue;
      // 57: new java/awt/event/ActionEvent
      // 5a: dup
      // 5b: aload 0
      // 5c: sipush 1001
      // 5f: getstatic ba.fc [Ljava/lang/String;
      // 62: bipush 0
      // 63: aaload
      // 64: invokespecial java/awt/event/ActionEvent.<init> (Ljava/lang/Object;ILjava/lang/String;)V
      // 67: invokevirtual java/awt/EventQueue.postEvent (Ljava/awt/AWTEvent;)V
      // 6a: goto 6e
      // 6d: athrow
      // 6e: goto 72
      // 71: astore 3
      // 72: goto ca
      // 75: astore 3
      // 76: aload 3
      // 77: new java/lang/StringBuilder
      // 7a: dup
      // 7b: invokespecial java/lang/StringBuilder.<init> ()V
      // 7e: getstatic ba.fc [Ljava/lang/String;
      // 81: bipush 3
      // 82: aaload
      // 83: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 86: aload 0
      // 87: ifnull 93
      // 8a: getstatic ba.fc [Ljava/lang/String;
      // 8d: bipush 2
      // 8e: aaload
      // 8f: goto 98
      // 92: athrow
      // 93: getstatic ba.fc [Ljava/lang/String;
      // 96: bipush 1
      // 97: aaload
      // 98: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 9b: bipush 44
      // 9d: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // a0: iload 1
      // a1: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // a4: bipush 44
      // a6: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // a9: aload 2
      // aa: ifnull b6
      // ad: getstatic ba.fc [Ljava/lang/String;
      // b0: bipush 2
      // b1: aaload
      // b2: goto bb
      // b5: athrow
      // b6: getstatic ba.fc [Ljava/lang/String;
      // b9: bipush 1
      // ba: aaload
      // bb: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // be: bipush 41
      // c0: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // c3: invokevirtual java/lang/StringBuilder.toString ()Ljava/lang/String;
      // c6: invokestatic i.a (Ljava/lang/Throwable;Ljava/lang/String;)Lla;
      // c9: athrow
      // ca: return
      // try (39 -> 54): 55 java/lang/Exception
      // try (2 -> 11): 57 java/lang/RuntimeException
      // try (12 -> 38): 57 java/lang/RuntimeException
      // try (39 -> 56): 57 java/lang/RuntimeException
   }

   private static char[] z(String var0) {
      char[] var10000 = var0.toCharArray();
      if (var10000.length < 2) {
         var10000[0] = (char)(var10000[0] ^ 2);
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
               var10005 = 109;
               break;
            case 1:
               var10005 = 102;
               break;
            case 2:
               var10005 = 106;
               break;
            case 3:
               var10005 = 113;
               break;
            default:
               var10005 = 2;
         }

         var10001[var1] = (char)(var10004 ^ var10005);
      }

      return new String(var10001).intern();
   }
}
