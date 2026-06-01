final class ga {
   static v c = new v(z(z("\u0005D\\F")), z(z("=vknw7")), z(z("\ral")), 2);
   static int d;
   static char[] a = new char[]{
      ' ',
      ' ',
      '_',
      '-',
      'à',
      'á',
      'â',
      'ä',
      'ã',
      'À',
      'Á',
      'Â',
      'Ä',
      'Ã',
      'è',
      'é',
      'ê',
      'ë',
      'È',
      'É',
      'Ê',
      'Ë',
      'í',
      'î',
      'ï',
      'Í',
      'Î',
      'Ï',
      'ò',
      'ó',
      'ô',
      'ö',
      'õ',
      'Ò',
      'Ó',
      'Ô',
      'Ö',
      'Õ',
      'ù',
      'ú',
      'û',
      'ü',
      'Ù',
      'Ú',
      'Û',
      'Ü',
      'ç',
      'Ç',
      'ÿ',
      'Ÿ',
      'ñ',
      'Ñ',
      'ß'
   };
   static String[] b;
   private static final String[] z = new String[]{z(z("<eak")), z(z(")>#)i")), z(z("5q#F<"))};

   static final String a(int param0, int param1, int param2, byte[] param3) {
      // $VF: Couldn't be decompiled
      // Please report this to the Vineflower issue tracker, at https://github.com/Vineflower/vineflower/issues with a copy of the class file (if you have the rights to distribute it!)
      // java.lang.RuntimeException: parsing failure!
      //   at org.jetbrains.java.decompiler.modules.decompiler.decompose.DomHelper.parseGraph(DomHelper.java:211)
      //   at org.jetbrains.java.decompiler.main.rels.MethodProcessor.codeToJava(MethodProcessor.java:166)
      //
      // Bytecode:
      // 00: getstatic client.vh Z
      // 03: istore 9
      // 05: getstatic ga.d I
      // 08: bipush 1
      // 09: iadd
      // 0a: putstatic ga.d I
      // 0d: iload 0
      // 0e: newarray 5
      // 10: astore 4
      // 12: bipush 0
      // 13: istore 5
      // 15: bipush 0
      // 16: istore 6
      // 18: iload 0
      // 19: bipush -1
      // 1a: ixor
      // 1b: iload 6
      // 1d: bipush -1
      // 1e: ixor
      // 1f: if_icmpge 96
      // 22: sipush 255
      // 25: aload 3
      // 26: iload 2
      // 27: iload 6
      // 29: ineg
      // 2a: isub
      // 2b: baload
      // 2c: iand
      // 2d: istore 7
      // 2f: bipush 0
      // 30: iload 7
      // 32: iload 9
      // 34: ifne 9f
      // 37: if_icmpne 47
      // 3a: goto 3e
      // 3d: athrow
      // 3e: iload 9
      // 40: ifeq 8e
      // 43: goto 47
      // 46: athrow
      // 47: iload 7
      // 49: bipush -1
      // 4a: ixor
      // 4b: sipush -129
      // 4e: if_icmpgt 83
      // 51: goto 55
      // 54: athrow
      // 55: iload 7
      // 57: bipush -1
      // 58: ixor
      // 59: sipush -161
      // 5c: if_icmpgt 67
      // 5f: goto 63
      // 62: athrow
      // 63: goto 83
      // 66: athrow
      // 67: getstatic nb.f [C
      // 6a: iload 7
      // 6c: sipush 128
      // 6f: isub
      // 70: caload
      // 71: istore 8
      // 73: iload 8
      // 75: bipush -1
      // 76: ixor
      // 77: bipush -1
      // 78: if_icmpne 7f
      // 7b: bipush 63
      // 7d: istore 8
      // 7f: iload 8
      // 81: istore 7
      // 83: aload 4
      // 85: iload 5
      // 87: iinc 5 1
      // 8a: iload 7
      // 8c: i2c
      // 8d: castore
      // 8e: iinc 6 1
      // 91: iload 9
      // 93: ifeq 18
      // 96: bipush -103
      // 98: bipush -63
      // 9a: iload 1
      // 9b: isub
      // 9c: bipush 49
      // 9e: idiv
      // 9f: idiv
      // a0: istore 6
      // a2: new java/lang/String
      // a5: dup
      // a6: aload 4
      // a8: bipush 0
      // a9: iload 5
      // ab: invokespecial java/lang/String.<init> ([CII)V
      // ae: areturn
      // af: astore 4
      // b1: aload 4
      // b3: new java/lang/StringBuilder
      // b6: dup
      // b7: invokespecial java/lang/StringBuilder.<init> ()V
      // ba: getstatic ga.z [Ljava/lang/String;
      // bd: bipush 2
      // be: aaload
      // bf: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // c2: iload 0
      // c3: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // c6: bipush 44
      // c8: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // cb: iload 1
      // cc: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // cf: bipush 44
      // d1: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // d4: iload 2
      // d5: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // d8: bipush 44
      // da: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // dd: aload 3
      // de: ifnull ea
      // e1: getstatic ga.z [Ljava/lang/String;
      // e4: bipush 1
      // e5: aaload
      // e6: goto ef
      // e9: athrow
      // ea: getstatic ga.z [Ljava/lang/String;
      // ed: bipush 0
      // ee: aaload
      // ef: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // f2: bipush 41
      // f4: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // f7: invokevirtual java/lang/StringBuilder.toString ()Ljava/lang/String;
      // fa: invokestatic i.a (Ljava/lang/Throwable;Ljava/lang/String;)Lla;
      // fd: athrow
   }

   private static char[] z(String var0) {
      char[] var10000 = var0.toCharArray();
      if (var10000.length < 2) {
         var10000[0] = (char)(var10000[0] ^ 20);
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
               var10005 = 82;
               break;
            case 1:
               var10005 = 16;
               break;
            case 2:
               var10005 = 13;
               break;
            case 3:
               var10005 = 7;
               break;
            default:
               var10005 = 20;
         }

         var10001[var1] = (char)(var10004 ^ var10005);
      }

      return new String(var10001).intern();
   }
}
