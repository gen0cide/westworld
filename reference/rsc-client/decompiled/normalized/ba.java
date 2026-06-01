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
         return (var1 >> 24 & 0xFF) + "." + (0xFF & var1 >> 16) + "." + (var1 >> 8 & 0xFF) + "." + (var1 & 0xFF);
      } catch (RuntimeException var3) {
         throw i.a(var3, fc[5] + var0 + ',' + var1 + ')');
      }
   }

   ba(int var1, int var2, int var3, Component var4) {
      super(var1, var2, var3, var4);
   }

   @Override
   final void a(int param1, int param2, int param3, int param4, int param5, int param6, byte param7, int param8) {
      // $VF: Couldn't be decompiled
      // Please report this to the Vineflower issue tracker, at https://github.com/Vineflower/vineflower/issues with a copy of the class file (if you have the rights to distribute it!)
      // java.lang.RuntimeException: parsing failure!
      //   at org.jetbrains.java.decompiler.modules.decompiler.decompose.DomHelper.parseGraph(DomHelper.java:211)
      //   at org.jetbrains.java.decompiler.main.rels.MethodProcessor.codeToJava(MethodProcessor.java:166)
      //
      // Bytecode:
      // 000: getstatic client.vh Z
      // 003: istore 10
      // 005: getstatic ba.ec I
      // 008: bipush 1
      // 009: iadd
      // 00a: putstatic ba.ec I
      // 00d: iload 2
      // 00e: ldc 50000
      // 010: if_icmpge 0b1
      // 013: ldc 40000
      // 015: iload 2
      // 016: if_icmple 092
      // 019: goto 01d
      // 01c: athrow
      // 01d: iload 2
      // 01e: sipush 20000
      // 021: if_icmplt 049
      // 024: goto 028
      // 027: athrow
      // 028: aload 0
      // 029: getfield ba.dc Lclient;
      // 02c: iload 5
      // 02e: iload 8
      // 030: bipush 105
      // 032: iload 3
      // 033: iload 1
      // 034: sipush -20000
      // 037: iload 2
      // 038: iadd
      // 039: iload 6
      // 03b: iload 4
      // 03d: invokevirtual client.a (IIIIIIII)V
      // 040: iload 10
      // 042: ifeq 0ca
      // 045: goto 049
      // 048: athrow
      // 049: iload 2
      // 04a: bipush -1
      // 04b: ixor
      // 04c: sipush -5001
      // 04f: if_icmple 06e
      // 052: goto 056
      // 055: athrow
      // 056: aload 0
      // 057: iload 4
      // 059: iload 5
      // 05b: iload 3
      // 05c: iload 6
      // 05e: sipush 5924
      // 061: iload 2
      // 062: invokespecial ua.f (IIIIII)V
      // 065: iload 10
      // 067: ifeq 0ca
      // 06a: goto 06e
      // 06d: athrow
      // 06e: aload 0
      // 06f: getfield ba.dc Lclient;
      // 072: iload 8
      // 074: iload 6
      // 076: iload 7
      // 078: bipush 9
      // 07a: ixor
      // 07b: iload 1
      // 07c: iload 4
      // 07e: iload 5
      // 080: iload 3
      // 081: sipush -5000
      // 084: iload 2
      // 085: iadd
      // 086: invokevirtual client.b (IIIIIIII)V
      // 089: iload 10
      // 08b: ifeq 0ca
      // 08e: goto 092
      // 091: athrow
      // 092: aload 0
      // 093: getfield ba.dc Lclient;
      // 096: iload 3
      // 097: iload 8
      // 099: iload 4
      // 09b: iload 2
      // 09c: ldc 40000
      // 09e: isub
      // 09f: iload 6
      // 0a1: bipush -122
      // 0a3: iload 5
      // 0a5: invokevirtual client.b (IIIIIII)V
      // 0a8: iload 10
      // 0aa: ifeq 0ca
      // 0ad: goto 0b1
      // 0b0: athrow
      // 0b1: aload 0
      // 0b2: getfield ba.dc Lclient;
      // 0b5: iload 8
      // 0b7: iload 4
      // 0b9: iload 5
      // 0bb: iload 3
      // 0bc: iload 2
      // 0bd: ldc 50000
      // 0bf: isub
      // 0c0: iload 6
      // 0c2: bipush 2
      // 0c3: invokevirtual client.a (IIIIIII)V
      // 0c6: goto 0ca
      // 0c9: athrow
      // 0ca: iload 7
      // 0cc: bipush 29
      // 0ce: if_icmpeq 0dc
      // 0d1: aconst_null
      // 0d2: checkcast [Ljava/lang/String;
      // 0d5: putstatic ba.ac [Ljava/lang/String;
      // 0d8: goto 0dc
      // 0db: athrow
      // 0dc: goto 146
      // 0df: astore 9
      // 0e1: aload 9
      // 0e3: new java/lang/StringBuilder
      // 0e6: dup
      // 0e7: invokespecial java/lang/StringBuilder.<init> ()V
      // 0ea: getstatic ba.fc [Ljava/lang/String;
      // 0ed: bipush 4
      // 0ee: aaload
      // 0ef: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 0f2: iload 1
      // 0f3: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 0f6: bipush 44
      // 0f8: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 0fb: iload 2
      // 0fc: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 0ff: bipush 44
      // 101: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 104: iload 3
      // 105: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 108: bipush 44
      // 10a: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 10d: iload 4
      // 10f: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 112: bipush 44
      // 114: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 117: iload 5
      // 119: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 11c: bipush 44
      // 11e: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 121: iload 6
      // 123: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 126: bipush 44
      // 128: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 12b: iload 7
      // 12d: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 130: bipush 44
      // 132: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 135: iload 8
      // 137: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 13a: bipush 41
      // 13c: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 13f: invokevirtual java/lang/StringBuilder.toString ()Ljava/lang/String;
      // 142: invokestatic i.a (Ljava/lang/Throwable;Ljava/lang/String;)Lla;
      // 145: athrow
      // 146: return
   }

   static final void a(Object param0, int param1, c param2) {
      // $VF: Couldn't be decompiled
      // Please report this to the Vineflower issue tracker, at https://github.com/Vineflower/vineflower/issues with a copy of the class file (if you have the rights to distribute it!)
      // java.lang.RuntimeException: parsing failure!
      //   at org.jetbrains.java.decompiler.modules.decompiler.decompose.DomHelper.parseGraph(DomHelper.java:211)
      //   at org.jetbrains.java.decompiler.main.rels.MethodProcessor.codeToJava(MethodProcessor.java:166)
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
