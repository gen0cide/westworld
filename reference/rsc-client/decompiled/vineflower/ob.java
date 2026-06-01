final class ob {
   static int e;
   static int d;
   static int g;
   private int c;
   static byte[][] j = new byte[1000][];
   private int a = 65000;
   private nb b;
   private nb f = null;
   static int i;
   static int[] h;
   private static final String[] z = new String[]{
      z(z("\u0011\u0001]5W")),
      z(z("\u0011\u0001]\u0000\u0010-\u0017\u0001\u001d\u0011\u0019KZ")),
      z(z("\u0011\u0001]6W")),
      z(z("\u0005M]Z\u0002")),
      z(z("\u0010\u0016\u001f\u0018")),
      z(z("\u0011\u0001]7W")),
      z(z("\u0011\u0001]H\u0016\u0010\n\u0007JW"))
   };

   final boolean a(int param1, int param2, int param3, byte[] param4) {
      // $VF: Couldn't be decompiled
      // Please report this to the Vineflower issue tracker, at https://github.com/Vineflower/vineflower/issues with a copy of the class file (if you have the rights to distribute it!)
      // java.lang.RuntimeException: parsing failure!
      //   at org.jetbrains.java.decompiler.modules.decompiler.decompose.DomHelper.parseGraph(DomHelper.java:211)
      //   at org.jetbrains.java.decompiler.main.rels.MethodProcessor.codeToJava(MethodProcessor.java:166)
      //
      // Bytecode:
      // 00: bipush 94
      // 02: iload 3
      // 03: bipush -61
      // 05: isub
      // 06: bipush 35
      // 08: idiv
      // 09: irem
      // 0a: istore 5
      // 0c: getstatic ob.d I
      // 0f: bipush 1
      // 10: iadd
      // 11: putstatic ob.d I
      // 14: aload 0
      // 15: getfield ob.f Lnb;
      // 18: dup
      // 19: astore 6
      // 1b: monitorenter
      // 1c: bipush -1
      // 1d: iload 2
      // 1e: bipush -1
      // 1f: ixor
      // 20: if_icmplt 2b
      // 23: aload 0
      // 24: getfield ob.a I
      // 27: iload 2
      // 28: if_icmpge 34
      // 2b: new java/lang/IllegalArgumentException
      // 2e: dup
      // 2f: invokespecial java/lang/IllegalArgumentException.<init> ()V
      // 32: athrow
      // 33: athrow
      // 34: aload 0
      // 35: iload 1
      // 36: aload 4
      // 38: bipush 1
      // 39: iload 2
      // 3a: bipush 4
      // 3b: invokespecial ob.a (I[BZII)Z
      // 3e: istore 7
      // 40: iload 7
      // 42: ifeq 49
      // 45: goto 55
      // 48: athrow
      // 49: aload 0
      // 4a: iload 1
      // 4b: aload 4
      // 4d: bipush 0
      // 4e: iload 2
      // 4f: bipush 4
      // 50: invokespecial ob.a (I[BZII)Z
      // 53: istore 7
      // 55: iload 7
      // 57: aload 6
      // 59: monitorexit
      // 5a: ireturn
      // 5b: astore 8
      // 5d: aload 6
      // 5f: monitorexit
      // 60: aload 8
      // 62: athrow
      // 63: astore 5
      // 65: aload 5
      // 67: new java/lang/StringBuilder
      // 6a: dup
      // 6b: invokespecial java/lang/StringBuilder.<init> ()V
      // 6e: getstatic ob.z [Ljava/lang/String;
      // 71: bipush 2
      // 72: aaload
      // 73: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 76: iload 1
      // 77: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 7a: bipush 44
      // 7c: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 7f: iload 2
      // 80: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 83: bipush 44
      // 85: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 88: iload 3
      // 89: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 8c: bipush 44
      // 8e: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 91: aload 4
      // 93: ifnull 9f
      // 96: getstatic ob.z [Ljava/lang/String;
      // 99: bipush 3
      // 9a: aaload
      // 9b: goto a4
      // 9e: athrow
      // 9f: getstatic ob.z [Ljava/lang/String;
      // a2: bipush 4
      // a3: aaload
      // a4: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // a7: bipush 41
      // a9: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // ac: invokevirtual java/lang/StringBuilder.toString ()Ljava/lang/String;
      // af: invokestatic i.a (Ljava/lang/Throwable;Ljava/lang/String;)Lla;
      // b2: athrow
   }

   final byte[] a(int param1, int param2) {
      // $VF: Couldn't be decompiled
      // Please report this to the Vineflower issue tracker, at https://github.com/Vineflower/vineflower/issues with a copy of the class file (if you have the rights to distribute it!)
      // java.lang.RuntimeException: parsing failure!
      //   at org.jetbrains.java.decompiler.modules.decompiler.decompose.DomHelper.parseGraph(DomHelper.java:211)
      //   at org.jetbrains.java.decompiler.main.rels.MethodProcessor.codeToJava(MethodProcessor.java:166)
      //
      // Bytecode:
      // 000: getstatic client.vh Z
      // 003: istore 18
      // 005: getstatic ob.i I
      // 008: bipush 1
      // 009: iadd
      // 00a: putstatic ob.i I
      // 00d: aload 0
      // 00e: getfield ob.f Lnb;
      // 011: dup
      // 012: astore 3
      // 013: monitorenter
      // 014: aload 0
      // 015: getfield ob.b Lnb;
      // 018: bipush -111
      // 01a: invokevirtual nb.a (B)J
      // 01d: ldc2_w -1
      // 020: lxor
      // 021: bipush 6
      // 023: iload 2
      // 024: bipush 6
      // 026: imul
      // 027: iadd
      // 028: i2l
      // 029: ldc2_w -1
      // 02c: lxor
      // 02d: lcmp
      // 02e: ifgt 034
      // 031: goto 038
      // 034: aconst_null
      // 035: aload 3
      // 036: monitorexit
      // 037: areturn
      // 038: aload 0
      // 039: getfield ob.b Lnb;
      // 03c: bipush 6
      // 03e: iload 2
      // 03f: imul
      // 040: i2l
      // 041: bipush 12
      // 043: invokevirtual nb.a (JI)V
      // 046: aload 0
      // 047: getfield ob.b Lnb;
      // 04a: bipush 1
      // 04b: bipush 6
      // 04d: bipush 0
      // 04e: getstatic la.c [B
      // 051: invokevirtual nb.a (ZII[B)V
      // 054: ldc 65280
      // 056: getstatic la.c [B
      // 059: bipush 1
      // 05a: baload
      // 05b: ldc 1647444456
      // 05d: ishl
      // 05e: iand
      // 05f: sipush 255
      // 062: getstatic la.c [B
      // 065: bipush 0
      // 066: baload
      // 067: iand
      // 068: ldc -463794928
      // 06a: ishl
      // 06b: sipush 255
      // 06e: getstatic la.c [B
      // 071: bipush 2
      // 072: baload
      // 073: iand
      // 074: iadd
      // 075: iadd
      // 076: istore 4
      // 078: getstatic la.c [B
      // 07b: bipush 4
      // 07c: baload
      // 07d: ldc 553924200
      // 07f: ishl
      // 080: ldc 65280
      // 082: iand
      // 083: getstatic la.c [B
      // 086: bipush 3
      // 087: baload
      // 088: sipush 255
      // 08b: iand
      // 08c: ldc -1798669456
      // 08e: ishl
      // 08f: iadd
      // 090: sipush 255
      // 093: getstatic la.c [B
      // 096: bipush 5
      // 097: baload
      // 098: iand
      // 099: ineg
      // 09a: isub
      // 09b: istore 5
      // 09d: iload 4
      // 09f: iflt 0b7
      // 0a2: iload 4
      // 0a4: bipush -1
      // 0a5: ixor
      // 0a6: aload 0
      // 0a7: getfield ob.a I
      // 0aa: bipush -1
      // 0ab: ixor
      // 0ac: if_icmplt 0b7
      // 0af: goto 0b3
      // 0b2: athrow
      // 0b3: goto 0bb
      // 0b6: athrow
      // 0b7: aconst_null
      // 0b8: aload 3
      // 0b9: monitorexit
      // 0ba: areturn
      // 0bb: iload 5
      // 0bd: bipush -1
      // 0be: ixor
      // 0bf: bipush -1
      // 0c0: if_icmpge 0e7
      // 0c3: aload 0
      // 0c4: getfield ob.f Lnb;
      // 0c7: bipush -111
      // 0c9: invokevirtual nb.a (B)J
      // 0cc: ldc2_w 520
      // 0cf: ldiv
      // 0d0: ldc2_w -1
      // 0d3: lxor
      // 0d4: iload 5
      // 0d6: i2l
      // 0d7: ldc2_w -1
      // 0da: lxor
      // 0db: lcmp
      // 0dc: ifgt 0e7
      // 0df: goto 0e3
      // 0e2: athrow
      // 0e3: goto 0eb
      // 0e6: athrow
      // 0e7: aconst_null
      // 0e8: aload 3
      // 0e9: monitorexit
      // 0ea: areturn
      // 0eb: iload 1
      // 0ec: sipush 9395
      // 0ef: if_icmpeq 0f9
      // 0f2: aconst_null
      // 0f3: checkcast [B
      // 0f6: aload 3
      // 0f7: monitorexit
      // 0f8: areturn
      // 0f9: iload 4
      // 0fb: newarray 8
      // 0fd: astore 6
      // 0ff: bipush 0
      // 100: istore 7
      // 102: bipush 0
      // 103: istore 8
      // 105: iload 4
      // 107: iload 7
      // 109: if_icmple 309
      // 10c: iload 5
      // 10e: ifeq 115
      // 111: goto 119
      // 114: athrow
      // 115: aconst_null
      // 116: aload 3
      // 117: monitorexit
      // 118: areturn
      // 119: aload 0
      // 11a: getfield ob.f Lnb;
      // 11d: iload 5
      // 11f: sipush 520
      // 122: imul
      // 123: i2l
      // 124: bipush 107
      // 126: invokevirtual nb.a (JI)V
      // 129: iload 7
      // 12b: ineg
      // 12c: iload 4
      // 12e: iadd
      // 12f: istore 9
      // 131: iload 2
      // 132: ldc 65535
      // 134: if_icmple 1e2
      // 137: iload 9
      // 139: bipush -1
      // 13a: ixor
      // 13b: sipush -511
      // 13e: if_icmplt 149
      // 141: goto 145
      // 144: athrow
      // 145: goto 14e
      // 148: athrow
      // 149: sipush 510
      // 14c: istore 9
      // 14e: bipush 10
      // 150: istore 14
      // 152: aload 0
      // 153: getfield ob.f Lnb;
      // 156: bipush 1
      // 157: iload 9
      // 159: iload 14
      // 15b: iadd
      // 15c: bipush 0
      // 15d: getstatic la.c [B
      // 160: invokevirtual nb.a (ZII[B)V
      // 163: getstatic la.c [B
      // 166: bipush 3
      // 167: baload
      // 168: sipush 255
      // 16b: iand
      // 16c: getstatic la.c [B
      // 16f: bipush 0
      // 170: baload
      // 171: ldc -854702312
      // 173: ishl
      // 174: ldc -16777216
      // 176: iand
      // 177: ldc 16711680
      // 179: getstatic la.c [B
      // 17c: bipush 1
      // 17d: baload
      // 17e: ldc -511327312
      // 180: ishl
      // 181: iand
      // 182: ineg
      // 183: isub
      // 184: ldc 65280
      // 186: getstatic la.c [B
      // 189: bipush 2
      // 18a: baload
      // 18b: ldc 2104943048
      // 18d: ishl
      // 18e: iand
      // 18f: iadd
      // 190: iadd
      // 191: istore 10
      // 193: getstatic la.c [B
      // 196: bipush 9
      // 198: baload
      // 199: sipush 255
      // 19c: iand
      // 19d: istore 13
      // 19f: getstatic la.c [B
      // 1a2: bipush 5
      // 1a3: baload
      // 1a4: sipush 255
      // 1a7: iand
      // 1a8: sipush 255
      // 1ab: getstatic la.c [B
      // 1ae: bipush 4
      // 1af: baload
      // 1b0: iand
      // 1b1: ldc 110659720
      // 1b3: ishl
      // 1b4: iadd
      // 1b5: istore 11
      // 1b7: getstatic la.c [B
      // 1ba: bipush 8
      // 1bc: baload
      // 1bd: sipush 255
      // 1c0: iand
      // 1c1: ldc 16711680
      // 1c3: getstatic la.c [B
      // 1c6: bipush 6
      // 1c8: baload
      // 1c9: ldc -1936015824
      // 1cb: ishl
      // 1cc: iand
      // 1cd: iadd
      // 1ce: getstatic la.c [B
      // 1d1: bipush 7
      // 1d3: baload
      // 1d4: ldc 2101716552
      // 1d6: ishl
      // 1d7: ldc 65280
      // 1d9: iand
      // 1da: iadd
      // 1db: istore 12
      // 1dd: iload 18
      // 1df: ifeq 270
      // 1e2: iload 9
      // 1e4: bipush -1
      // 1e5: ixor
      // 1e6: sipush -513
      // 1e9: if_icmplt 1f4
      // 1ec: goto 1f0
      // 1ef: athrow
      // 1f0: goto 1f9
      // 1f3: athrow
      // 1f4: sipush 512
      // 1f7: istore 9
      // 1f9: bipush 8
      // 1fb: istore 14
      // 1fd: aload 0
      // 1fe: getfield ob.f Lnb;
      // 201: bipush 1
      // 202: iload 14
      // 204: iload 9
      // 206: iadd
      // 207: bipush 0
      // 208: getstatic la.c [B
      // 20b: invokevirtual nb.a (ZII[B)V
      // 20e: getstatic la.c [B
      // 211: bipush 7
      // 213: baload
      // 214: sipush 255
      // 217: iand
      // 218: istore 13
      // 21a: getstatic la.c [B
      // 21d: bipush 0
      // 21e: baload
      // 21f: ldc 524989928
      // 221: ishl
      // 222: ldc 65280
      // 224: iand
      // 225: sipush 255
      // 228: getstatic la.c [B
      // 22b: bipush 1
      // 22c: baload
      // 22d: iand
      // 22e: iadd
      // 22f: istore 10
      // 231: getstatic la.c [B
      // 234: bipush 2
      // 235: baload
      // 236: ldc -1311007928
      // 238: ishl
      // 239: ldc 65280
      // 23b: iand
      // 23c: getstatic la.c [B
      // 23f: bipush 3
      // 240: baload
      // 241: sipush 255
      // 244: iand
      // 245: iadd
      // 246: istore 11
      // 248: sipush 255
      // 24b: getstatic la.c [B
      // 24e: bipush 4
      // 24f: baload
      // 250: iand
      // 251: ldc 862358512
      // 253: ishl
      // 254: getstatic la.c [B
      // 257: bipush 5
      // 258: baload
      // 259: sipush 255
      // 25c: iand
      // 25d: ldc 550121000
      // 25f: ishl
      // 260: ineg
      // 261: isub
      // 262: sipush 255
      // 265: getstatic la.c [B
      // 268: bipush 6
      // 26a: baload
      // 26b: iand
      // 26c: ineg
      // 26d: isub
      // 26e: istore 12
      // 270: iload 2
      // 271: bipush -1
      // 272: ixor
      // 273: iload 10
      // 275: bipush -1
      // 276: ixor
      // 277: if_icmpne 29a
      // 27a: iload 8
      // 27c: iload 11
      // 27e: if_icmpne 29a
      // 281: goto 285
      // 284: athrow
      // 285: iload 13
      // 287: bipush -1
      // 288: ixor
      // 289: aload 0
      // 28a: getfield ob.c I
      // 28d: bipush -1
      // 28e: ixor
      // 28f: if_icmpne 29a
      // 292: goto 296
      // 295: athrow
      // 296: goto 29e
      // 299: athrow
      // 29a: aconst_null
      // 29b: aload 3
      // 29c: monitorexit
      // 29d: areturn
      // 29e: iload 12
      // 2a0: iflt 2c7
      // 2a3: aload 0
      // 2a4: getfield ob.f Lnb;
      // 2a7: bipush -111
      // 2a9: invokevirtual nb.a (B)J
      // 2ac: ldc2_w 520
      // 2af: ldiv
      // 2b0: ldc2_w -1
      // 2b3: lxor
      // 2b4: iload 12
      // 2b6: i2l
      // 2b7: ldc2_w -1
      // 2ba: lxor
      // 2bb: lcmp
      // 2bc: ifgt 2c7
      // 2bf: goto 2c3
      // 2c2: athrow
      // 2c3: goto 2cb
      // 2c6: athrow
      // 2c7: aconst_null
      // 2c8: aload 3
      // 2c9: monitorexit
      // 2ca: areturn
      // 2cb: iload 14
      // 2cd: iload 9
      // 2cf: ineg
      // 2d0: isub
      // 2d1: istore 15
      // 2d3: iload 12
      // 2d5: istore 5
      // 2d7: iload 14
      // 2d9: istore 16
      // 2db: iload 15
      // 2dd: iload 16
      // 2df: if_icmple 301
      // 2e2: aload 6
      // 2e4: iload 7
      // 2e6: iinc 7 1
      // 2e9: getstatic la.c [B
      // 2ec: iload 16
      // 2ee: baload
      // 2ef: bastore
      // 2f0: iinc 16 1
      // 2f3: iload 18
      // 2f5: ifne 304
      // 2f8: iload 18
      // 2fa: ifeq 2db
      // 2fd: goto 301
      // 300: athrow
      // 301: iinc 8 1
      // 304: iload 18
      // 306: ifeq 105
      // 309: aload 6
      // 30b: aload 3
      // 30c: monitorexit
      // 30d: areturn
      // 30e: astore 4
      // 310: aconst_null
      // 311: aload 3
      // 312: monitorexit
      // 313: areturn
      // 314: astore 17
      // 316: aload 3
      // 317: monitorexit
      // 318: aload 17
      // 31a: athrow
      // 31b: astore 3
      // 31c: aload 3
      // 31d: new java/lang/StringBuilder
      // 320: dup
      // 321: invokespecial java/lang/StringBuilder.<init> ()V
      // 324: getstatic ob.z [Ljava/lang/String;
      // 327: bipush 0
      // 328: aaload
      // 329: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 32c: iload 1
      // 32d: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 330: bipush 44
      // 332: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 335: iload 2
      // 336: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 339: bipush 41
      // 33b: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 33e: invokevirtual java/lang/StringBuilder.toString ()Ljava/lang/String;
      // 341: invokestatic i.a (Ljava/lang/Throwable;Ljava/lang/String;)Lla;
      // 344: athrow
   }

   private final boolean a(int param1, byte[] param2, boolean param3, int param4, int param5) {
      // $VF: Couldn't be decompiled
      // Please report this to the Vineflower issue tracker, at https://github.com/Vineflower/vineflower/issues with a copy of the class file (if you have the rights to distribute it!)
      // java.lang.RuntimeException: parsing failure!
      //   at org.jetbrains.java.decompiler.modules.decompiler.decompose.DomHelper.parseGraph(DomHelper.java:211)
      //   at org.jetbrains.java.decompiler.main.rels.MethodProcessor.codeToJava(MethodProcessor.java:166)
      //
      // Bytecode:
      // 000: getstatic client.vh Z
      // 003: istore 16
      // 005: getstatic ob.e I
      // 008: bipush 1
      // 009: iadd
      // 00a: putstatic ob.e I
      // 00d: aload 0
      // 00e: getfield ob.f Lnb;
      // 011: dup
      // 012: astore 6
      // 014: monitorenter
      // 015: iload 3
      // 016: ifeq 0a2
      // 019: bipush 6
      // 01b: bipush 6
      // 01d: iload 1
      // 01e: imul
      // 01f: iadd
      // 020: i2l
      // 021: ldc2_w -1
      // 024: lxor
      // 025: aload 0
      // 026: getfield ob.b Lnb;
      // 029: bipush -111
      // 02b: invokevirtual nb.a (B)J
      // 02e: ldc2_w -1
      // 031: lxor
      // 032: lcmp
      // 033: ifge 03b
      // 036: bipush 0
      // 037: aload 6
      // 039: monitorexit
      // 03a: ireturn
      // 03b: aload 0
      // 03c: getfield ob.b Lnb;
      // 03f: iload 1
      // 040: bipush 6
      // 042: imul
      // 043: i2l
      // 044: bipush -124
      // 046: invokevirtual nb.a (JI)V
      // 049: aload 0
      // 04a: getfield ob.b Lnb;
      // 04d: bipush 1
      // 04e: bipush 6
      // 050: bipush 0
      // 051: getstatic la.c [B
      // 054: invokevirtual nb.a (ZII[B)V
      // 057: sipush 255
      // 05a: getstatic la.c [B
      // 05d: bipush 5
      // 05e: baload
      // 05f: iand
      // 060: getstatic la.c [B
      // 063: bipush 3
      // 064: baload
      // 065: ldc 4539376
      // 067: ishl
      // 068: ldc 16711680
      // 06a: iand
      // 06b: iadd
      // 06c: getstatic la.c [B
      // 06f: bipush 4
      // 070: baload
      // 071: sipush 255
      // 074: iand
      // 075: ldc 1519652136
      // 077: ishl
      // 078: ineg
      // 079: isub
      // 07a: istore 7
      // 07c: iload 7
      // 07e: ifle 09d
      // 081: aload 0
      // 082: getfield ob.f Lnb;
      // 085: bipush -111
      // 087: invokevirtual nb.a (B)J
      // 08a: ldc2_w 520
      // 08d: ldiv
      // 08e: iload 7
      // 090: i2l
      // 091: lcmp
      // 092: iflt 09d
      // 095: goto 099
      // 098: athrow
      // 099: goto 0bf
      // 09c: athrow
      // 09d: bipush 0
      // 09e: aload 6
      // 0a0: monitorexit
      // 0a1: ireturn
      // 0a2: aload 0
      // 0a3: getfield ob.f Lnb;
      // 0a6: bipush -111
      // 0a8: invokevirtual nb.a (B)J
      // 0ab: ldc2_w -519
      // 0ae: lsub
      // 0af: ldc2_w 520
      // 0b2: ldiv
      // 0b3: l2i
      // 0b4: istore 7
      // 0b6: bipush 0
      // 0b7: iload 7
      // 0b9: if_icmpne 0bf
      // 0bc: bipush 1
      // 0bd: istore 7
      // 0bf: getstatic la.c [B
      // 0c2: bipush 0
      // 0c3: iload 4
      // 0c5: ldc 1377278608
      // 0c7: ishr
      // 0c8: i2b
      // 0c9: bastore
      // 0ca: getstatic la.c [B
      // 0cd: bipush 2
      // 0ce: iload 4
      // 0d0: i2b
      // 0d1: bastore
      // 0d2: getstatic la.c [B
      // 0d5: bipush 3
      // 0d6: iload 7
      // 0d8: ldc -340974608
      // 0da: ishr
      // 0db: i2b
      // 0dc: bastore
      // 0dd: getstatic la.c [B
      // 0e0: iload 5
      // 0e2: iload 7
      // 0e4: ldc -1558555128
      // 0e6: ishr
      // 0e7: i2b
      // 0e8: bastore
      // 0e9: getstatic la.c [B
      // 0ec: bipush 1
      // 0ed: iload 4
      // 0ef: ldc -236170680
      // 0f1: ishr
      // 0f2: i2b
      // 0f3: bastore
      // 0f4: getstatic la.c [B
      // 0f7: bipush 5
      // 0f8: iload 7
      // 0fa: i2b
      // 0fb: bastore
      // 0fc: aload 0
      // 0fd: getfield ob.b Lnb;
      // 100: bipush 6
      // 102: iload 1
      // 103: imul
      // 104: i2l
      // 105: bipush 31
      // 107: invokevirtual nb.a (JI)V
      // 10a: aload 0
      // 10b: getfield ob.b Lnb;
      // 10e: bipush 6
      // 110: bipush -102
      // 112: bipush 0
      // 113: getstatic la.c [B
      // 116: invokevirtual nb.a (III[B)V
      // 119: bipush 0
      // 11a: istore 8
      // 11c: bipush 0
      // 11d: istore 9
      // 11f: iload 8
      // 121: bipush -1
      // 122: ixor
      // 123: iload 4
      // 125: bipush -1
      // 126: ixor
      // 127: if_icmple 47d
      // 12a: bipush 0
      // 12b: istore 10
      // 12d: iload 3
      // 12e: iload 16
      // 130: ifne 481
      // 133: ifeq 2c2
      // 136: goto 13a
      // 139: athrow
      // 13a: aload 0
      // 13b: getfield ob.f Lnb;
      // 13e: sipush 520
      // 141: iload 7
      // 143: imul
      // 144: i2l
      // 145: iload 5
      // 147: bipush 17
      // 149: ixor
      // 14a: invokevirtual nb.a (JI)V
      // 14d: iload 1
      // 14e: ldc 65535
      // 150: if_icmpgt 1d6
      // 153: goto 157
      // 156: athrow
      // 157: aload 0
      // 158: getfield ob.f Lnb;
      // 15b: bipush 1
      // 15c: bipush 8
      // 15e: bipush 0
      // 15f: getstatic la.c [B
      // 162: invokevirtual nb.a (ZII[B)V
      // 165: goto 170
      // 168: athrow
      // 169: astore 14
      // 16b: iload 16
      // 16d: ifeq 47d
      // 170: ldc 65280
      // 172: getstatic la.c [B
      // 175: bipush 0
      // 176: baload
      // 177: ldc -1162714328
      // 179: ishl
      // 17a: iand
      // 17b: getstatic la.c [B
      // 17e: bipush 1
      // 17f: baload
      // 180: sipush 255
      // 183: iand
      // 184: ineg
      // 185: isub
      // 186: istore 11
      // 188: getstatic la.c [B
      // 18b: bipush 6
      // 18d: baload
      // 18e: sipush 255
      // 191: iand
      // 192: ldc 16711680
      // 194: getstatic la.c [B
      // 197: bipush 4
      // 198: baload
      // 199: ldc -165446384
      // 19b: ishl
      // 19c: iand
      // 19d: iadd
      // 19e: sipush 255
      // 1a1: getstatic la.c [B
      // 1a4: bipush 5
      // 1a5: baload
      // 1a6: iand
      // 1a7: ldc -892700824
      // 1a9: ishl
      // 1aa: ineg
      // 1ab: isub
      // 1ac: istore 10
      // 1ae: sipush 255
      // 1b1: getstatic la.c [B
      // 1b4: bipush 7
      // 1b6: baload
      // 1b7: iand
      // 1b8: istore 13
      // 1ba: getstatic la.c [B
      // 1bd: bipush 3
      // 1be: baload
      // 1bf: sipush 255
      // 1c2: iand
      // 1c3: getstatic la.c [B
      // 1c6: bipush 2
      // 1c7: baload
      // 1c8: ldc -941949624
      // 1ca: ishl
      // 1cb: ldc 65280
      // 1cd: iand
      // 1ce: iadd
      // 1cf: istore 12
      // 1d1: iload 16
      // 1d3: ifeq 26c
      // 1d6: aload 0
      // 1d7: getfield ob.f Lnb;
      // 1da: bipush 1
      // 1db: bipush 10
      // 1dd: bipush 0
      // 1de: getstatic la.c [B
      // 1e1: invokevirtual nb.a (ZII[B)V
      // 1e4: goto 1ef
      // 1e7: athrow
      // 1e8: astore 14
      // 1ea: iload 16
      // 1ec: ifeq 47d
      // 1ef: sipush 255
      // 1f2: getstatic la.c [B
      // 1f5: bipush 5
      // 1f6: baload
      // 1f7: iand
      // 1f8: getstatic la.c [B
      // 1fb: bipush 4
      // 1fc: baload
      // 1fd: ldc -1447609528
      // 1ff: ishl
      // 200: ldc 65280
      // 202: iand
      // 203: iadd
      // 204: istore 12
      // 206: getstatic la.c [B
      // 209: bipush 6
      // 20b: baload
      // 20c: ldc 725235568
      // 20e: ishl
      // 20f: ldc 16711680
      // 211: iand
      // 212: getstatic la.c [B
      // 215: bipush 7
      // 217: baload
      // 218: sipush 255
      // 21b: iand
      // 21c: ldc 221297448
      // 21e: ishl
      // 21f: ineg
      // 220: isub
      // 221: getstatic la.c [B
      // 224: bipush 8
      // 226: baload
      // 227: sipush 255
      // 22a: iand
      // 22b: iadd
      // 22c: istore 10
      // 22e: getstatic la.c [B
      // 231: bipush 0
      // 232: baload
      // 233: ldc -2111158472
      // 235: ishl
      // 236: ldc -16777216
      // 238: iand
      // 239: getstatic la.c [B
      // 23c: bipush 1
      // 23d: baload
      // 23e: ldc -1296842608
      // 240: ishl
      // 241: ldc 16711680
      // 243: iand
      // 244: sipush 255
      // 247: getstatic la.c [B
      // 24a: bipush 2
      // 24b: baload
      // 24c: iand
      // 24d: ldc -1391218456
      // 24f: ishl
      // 250: ineg
      // 251: isub
      // 252: iadd
      // 253: sipush 255
      // 256: getstatic la.c [B
      // 259: bipush 3
      // 25a: baload
      // 25b: iand
      // 25c: ineg
      // 25d: isub
      // 25e: istore 11
      // 260: sipush 255
      // 263: getstatic la.c [B
      // 266: bipush 9
      // 268: baload
      // 269: iand
      // 26a: istore 13
      // 26c: iload 1
      // 26d: bipush -1
      // 26e: ixor
      // 26f: iload 11
      // 271: bipush -1
      // 272: ixor
      // 273: if_icmpne 292
      // 276: iload 12
      // 278: iload 9
      // 27a: if_icmpne 292
      // 27d: goto 281
      // 280: athrow
      // 281: iload 13
      // 283: aload 0
      // 284: getfield ob.c I
      // 287: if_icmpne 292
      // 28a: goto 28e
      // 28d: athrow
      // 28e: goto 297
      // 291: athrow
      // 292: bipush 0
      // 293: aload 6
      // 295: monitorexit
      // 296: ireturn
      // 297: bipush 0
      // 298: iload 10
      // 29a: if_icmpgt 2bd
      // 29d: iload 10
      // 29f: i2l
      // 2a0: ldc2_w -1
      // 2a3: lxor
      // 2a4: aload 0
      // 2a5: getfield ob.f Lnb;
      // 2a8: bipush -111
      // 2aa: invokevirtual nb.a (B)J
      // 2ad: ldc2_w 520
      // 2b0: ldiv
      // 2b1: ldc2_w -1
      // 2b4: lxor
      // 2b5: lcmp
      // 2b6: ifge 2c2
      // 2b9: goto 2bd
      // 2bc: athrow
      // 2bd: bipush 0
      // 2be: aload 6
      // 2c0: monitorexit
      // 2c1: ireturn
      // 2c2: iload 10
      // 2c4: bipush -1
      // 2c5: ixor
      // 2c6: bipush -1
      // 2c7: if_icmpne 2fb
      // 2ca: bipush 0
      // 2cb: istore 3
      // 2cc: aload 0
      // 2cd: getfield ob.f Lnb;
      // 2d0: bipush -111
      // 2d2: invokevirtual nb.a (B)J
      // 2d5: ldc2_w -519
      // 2d8: lsub
      // 2d9: ldc2_w 520
      // 2dc: ldiv
      // 2dd: l2i
      // 2de: istore 10
      // 2e0: bipush 0
      // 2e1: iload 10
      // 2e3: if_icmpeq 2ea
      // 2e6: goto 2ed
      // 2e9: athrow
      // 2ea: iinc 10 1
      // 2ed: iload 10
      // 2ef: iload 7
      // 2f1: if_icmpne 2fb
      // 2f4: iinc 10 1
      // 2f7: goto 2fb
      // 2fa: athrow
      // 2fb: sipush 512
      // 2fe: iload 4
      // 300: iload 8
      // 302: isub
      // 303: if_icmplt 309
      // 306: bipush 0
      // 307: istore 10
      // 309: ldc -65536
      // 30b: iload 1
      // 30c: bipush -1
      // 30d: ixor
      // 30e: if_icmple 3cf
      // 311: getstatic la.c [B
      // 314: bipush 0
      // 315: iload 1
      // 316: ldc -322435432
      // 318: ishr
      // 319: i2b
      // 31a: bastore
      // 31b: getstatic la.c [B
      // 31e: bipush 5
      // 31f: iload 9
      // 321: i2b
      // 322: bastore
      // 323: getstatic la.c [B
      // 326: bipush 2
      // 327: iload 1
      // 328: ldc -936822456
      // 32a: ishr
      // 32b: i2b
      // 32c: bastore
      // 32d: getstatic la.c [B
      // 330: bipush 4
      // 331: iload 9
      // 333: ldc 521649416
      // 335: ishr
      // 336: i2b
      // 337: bastore
      // 338: getstatic la.c [B
      // 33b: bipush 7
      // 33d: iload 10
      // 33f: ldc -876359160
      // 341: ishr
      // 342: i2b
      // 343: bastore
      // 344: getstatic la.c [B
      // 347: bipush 1
      // 348: iload 1
      // 349: ldc 527626448
      // 34b: ishr
      // 34c: i2b
      // 34d: bastore
      // 34e: getstatic la.c [B
      // 351: bipush 8
      // 353: iload 10
      // 355: i2b
      // 356: bastore
      // 357: getstatic la.c [B
      // 35a: bipush 9
      // 35c: aload 0
      // 35d: getfield ob.c I
      // 360: i2b
      // 361: bastore
      // 362: getstatic la.c [B
      // 365: bipush 3
      // 366: iload 1
      // 367: i2b
      // 368: bastore
      // 369: getstatic la.c [B
      // 36c: bipush 6
      // 36e: iload 10
      // 370: ldc -1755814992
      // 372: ishr
      // 373: i2b
      // 374: bastore
      // 375: aload 0
      // 376: getfield ob.f Lnb;
      // 379: sipush 520
      // 37c: iload 7
      // 37e: imul
      // 37f: i2l
      // 380: iload 5
      // 382: bipush 33
      // 384: ixor
      // 385: invokevirtual nb.a (JI)V
      // 388: aload 0
      // 389: getfield ob.f Lnb;
      // 38c: bipush 10
      // 38e: bipush -111
      // 390: bipush 0
      // 391: getstatic la.c [B
      // 394: invokevirtual nb.a (III[B)V
      // 397: iload 4
      // 399: iload 8
      // 39b: ineg
      // 39c: iadd
      // 39d: istore 11
      // 39f: sipush -511
      // 3a2: iload 11
      // 3a4: bipush -1
      // 3a5: ixor
      // 3a6: if_icmpgt 3ad
      // 3a9: goto 3b2
      // 3ac: athrow
      // 3ad: sipush 510
      // 3b0: istore 11
      // 3b2: aload 0
      // 3b3: getfield ob.f Lnb;
      // 3b6: iload 11
      // 3b8: iload 5
      // 3ba: bipush -119
      // 3bc: iadd
      // 3bd: iload 8
      // 3bf: aload 2
      // 3c0: invokevirtual nb.a (III[B)V
      // 3c3: iload 8
      // 3c5: iload 11
      // 3c7: iadd
      // 3c8: istore 8
      // 3ca: iload 16
      // 3cc: ifeq 471
      // 3cf: getstatic la.c [B
      // 3d2: bipush 4
      // 3d3: iload 10
      // 3d5: ldc 1803659792
      // 3d7: ishr
      // 3d8: i2b
      // 3d9: bastore
      // 3da: getstatic la.c [B
      // 3dd: bipush 0
      // 3de: iload 1
      // 3df: ldc -566551576
      // 3e1: ishr
      // 3e2: i2b
      // 3e3: bastore
      // 3e4: getstatic la.c [B
      // 3e7: bipush 7
      // 3e9: aload 0
      // 3ea: getfield ob.c I
      // 3ed: i2b
      // 3ee: bastore
      // 3ef: getstatic la.c [B
      // 3f2: bipush 6
      // 3f4: iload 10
      // 3f6: i2b
      // 3f7: bastore
      // 3f8: getstatic la.c [B
      // 3fb: bipush 3
      // 3fc: iload 9
      // 3fe: i2b
      // 3ff: bastore
      // 400: getstatic la.c [B
      // 403: bipush 1
      // 404: iload 1
      // 405: i2b
      // 406: bastore
      // 407: getstatic la.c [B
      // 40a: bipush 2
      // 40b: iload 9
      // 40d: ldc -447508120
      // 40f: ishr
      // 410: i2b
      // 411: bastore
      // 412: getstatic la.c [B
      // 415: bipush 5
      // 416: iload 10
      // 418: ldc -1738018904
      // 41a: ishr
      // 41b: i2b
      // 41c: bastore
      // 41d: aload 0
      // 41e: getfield ob.f Lnb;
      // 421: iload 7
      // 423: sipush 520
      // 426: imul
      // 427: i2l
      // 428: iload 5
      // 42a: bipush 127
      // 42c: ixor
      // 42d: invokevirtual nb.a (JI)V
      // 430: aload 0
      // 431: getfield ob.f Lnb;
      // 434: bipush 8
      // 436: bipush -107
      // 438: bipush 0
      // 439: getstatic la.c [B
      // 43c: invokevirtual nb.a (III[B)V
      // 43f: iload 4
      // 441: iload 8
      // 443: isub
      // 444: istore 11
      // 446: iload 11
      // 448: bipush -1
      // 449: ixor
      // 44a: sipush -513
      // 44d: if_icmplt 454
      // 450: goto 459
      // 453: athrow
      // 454: sipush 512
      // 457: istore 11
      // 459: aload 0
      // 45a: getfield ob.f Lnb;
      // 45d: iload 11
      // 45f: iload 5
      // 461: bipush -125
      // 463: iadd
      // 464: iload 8
      // 466: aload 2
      // 467: invokevirtual nb.a (III[B)V
      // 46a: iload 8
      // 46c: iload 11
      // 46e: iadd
      // 46f: istore 8
      // 471: iload 10
      // 473: istore 7
      // 475: iinc 9 1
      // 478: iload 16
      // 47a: ifeq 11f
      // 47d: bipush 1
      // 47e: aload 6
      // 480: monitorexit
      // 481: ireturn
      // 482: astore 7
      // 484: bipush 0
      // 485: aload 6
      // 487: monitorexit
      // 488: ireturn
      // 489: astore 15
      // 48b: aload 6
      // 48d: monitorexit
      // 48e: aload 15
      // 490: athrow
      // 491: astore 6
      // 493: aload 6
      // 495: new java/lang/StringBuilder
      // 498: dup
      // 499: invokespecial java/lang/StringBuilder.<init> ()V
      // 49c: getstatic ob.z [Ljava/lang/String;
      // 49f: bipush 5
      // 4a0: aaload
      // 4a1: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 4a4: iload 1
      // 4a5: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 4a8: bipush 44
      // 4aa: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 4ad: aload 2
      // 4ae: ifnull 4ba
      // 4b1: getstatic ob.z [Ljava/lang/String;
      // 4b4: bipush 3
      // 4b5: aaload
      // 4b6: goto 4bf
      // 4b9: athrow
      // 4ba: getstatic ob.z [Ljava/lang/String;
      // 4bd: bipush 4
      // 4be: aaload
      // 4bf: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 4c2: bipush 44
      // 4c4: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 4c7: iload 3
      // 4c8: invokevirtual java/lang/StringBuilder.append (Z)Ljava/lang/StringBuilder;
      // 4cb: bipush 44
      // 4cd: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 4d0: iload 4
      // 4d2: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 4d5: bipush 44
      // 4d7: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 4da: iload 5
      // 4dc: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 4df: bipush 41
      // 4e1: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 4e4: invokevirtual java/lang/StringBuilder.toString ()Ljava/lang/String;
      // 4e7: invokestatic i.a (Ljava/lang/Throwable;Ljava/lang/String;)Lla;
      // 4ea: athrow
   }

   @Override
   public final String toString() {
      try {
         g++;
         return "" + this.c;
      } catch (RuntimeException var2) {
         throw i.a(var2, z[1]);
      }
   }

   ob(int var1, nb var2, nb var3, int var4) {
      this.b = null;

      try {
         this.a = var4;
         this.c = var1;
         this.b = var3;
         this.f = var2;
      } catch (RuntimeException var8) {
         RuntimeException var5 = var8;

         RuntimeException var10000;
         StringBuilder var10001;
         String var10002;
         label34: {
            try {
               var10000 = var5;
               var10001 = new StringBuilder().append(z[6]).append(var1).append(',');
               if (var2 != null) {
                  var10002 = z[3];
                  break label34;
               }
            } catch (RuntimeException var7) {
               throw var7;
            }

            var10002 = z[4];
         }

         try {
            var10001 = var10001.append(var10002).append(',');
            if (var3 != null) {
               throw i.a(var10000, var10001.append(z[3]).append(',').append(var4).append(')').toString());
            }
         } catch (RuntimeException var6) {
            throw var6;
         }

         throw i.a(var10000, var10001.append(z[4]).append(',').append(var4).append(')').toString());
      }
   }

   private static char[] z(String var0) {
      char[] var10000 = var0.toCharArray();
      if (var10000.length < 2) {
         var10000[0] = (char)(var10000[0] ^ 127);
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
               var10005 = 126;
               break;
            case 1:
               var10005 = 99;
               break;
            case 2:
               var10005 = 115;
               break;
            case 3:
               var10005 = 116;
               break;
            default:
               var10005 = 127;
         }

         var10001[var1] = (char)(var10004 ^ var10005);
      }

      return new String(var10001).intern();
   }
}
