final class h {
   static int[] c;
   static int d;
   static int[] b;
   static int a;
   static String[] e;
   private static final String[] z = new String[]{z(z("~\u0004$[5")), z(z("k_f\u0019")), z(z("m\u0004K]"))};

   static final byte[] a(CharSequence param0, byte param1) {
      // $VF: Couldn't be decompiled
      // Please report this to the Vineflower issue tracker, at https://github.com/Vineflower/vineflower/issues with a copy of the class file (if you have the rights to distribute it!)
      // java.lang.RuntimeException: parsing failure!
      //   at org.jetbrains.java.decompiler.modules.decompiler.decompose.DomHelper.parseGraph(DomHelper.java:211)
      //   at org.jetbrains.java.decompiler.main.rels.MethodProcessor.codeToJava(MethodProcessor.java:166)
      //
      // Bytecode:
      // 000: getstatic client.vh Z
      // 003: istore 7
      // 005: bipush -91
      // 007: iload 1
      // 008: bipush -26
      // 00a: isub
      // 00b: bipush 53
      // 00d: idiv
      // 00e: idiv
      // 00f: istore 3
      // 010: getstatic h.d I
      // 013: bipush 1
      // 014: iadd
      // 015: putstatic h.d I
      // 018: aload 0
      // 019: invokeinterface java/lang/CharSequence.length ()I 1
      // 01e: istore 2
      // 01f: iload 2
      // 020: newarray 8
      // 022: astore 4
      // 024: bipush 0
      // 025: istore 5
      // 027: iload 2
      // 028: bipush -1
      // 029: ixor
      // 02a: iload 5
      // 02c: bipush -1
      // 02d: ixor
      // 02e: if_icmpge 392
      // 031: aload 0
      // 032: iload 5
      // 034: invokeinterface java/lang/CharSequence.charAt (I)C 2
      // 039: istore 6
      // 03b: iload 6
      // 03d: ifle 04c
      // 040: iload 6
      // 042: sipush 128
      // 045: if_icmplt 37e
      // 048: goto 04c
      // 04b: athrow
      // 04c: sipush -161
      // 04f: iload 6
      // 051: bipush -1
      // 052: ixor
      // 053: if_icmplt 066
      // 056: goto 05a
      // 059: athrow
      // 05a: iload 6
      // 05c: sipush 255
      // 05f: if_icmple 37e
      // 062: goto 066
      // 065: athrow
      // 066: iload 6
      // 068: sipush 8364
      // 06b: if_icmpeq 36e
      // 06e: goto 072
      // 071: athrow
      // 072: iload 6
      // 074: sipush 8218
      // 077: if_icmpne 08e
      // 07a: goto 07e
      // 07d: athrow
      // 07e: aload 4
      // 080: iload 5
      // 082: bipush -126
      // 084: bastore
      // 085: iload 7
      // 087: ifeq 38a
      // 08a: goto 08e
      // 08d: athrow
      // 08e: sipush 402
      // 091: iload 6
      // 093: if_icmpeq 35e
      // 096: goto 09a
      // 099: athrow
      // 09a: iload 6
      // 09c: bipush -1
      // 09d: ixor
      // 09e: sipush -8223
      // 0a1: if_icmpeq 34e
      // 0a4: goto 0a8
      // 0a7: athrow
      // 0a8: sipush 8230
      // 0ab: iload 6
      // 0ad: if_icmpne 0c4
      // 0b0: goto 0b4
      // 0b3: athrow
      // 0b4: aload 4
      // 0b6: iload 5
      // 0b8: bipush -123
      // 0ba: bastore
      // 0bb: iload 7
      // 0bd: ifeq 38a
      // 0c0: goto 0c4
      // 0c3: athrow
      // 0c4: iload 6
      // 0c6: bipush -1
      // 0c7: ixor
      // 0c8: sipush -8225
      // 0cb: if_icmpeq 33e
      // 0ce: goto 0d2
      // 0d1: athrow
      // 0d2: sipush -8226
      // 0d5: iload 6
      // 0d7: bipush -1
      // 0d8: ixor
      // 0d9: if_icmpeq 32e
      // 0dc: goto 0e0
      // 0df: athrow
      // 0e0: sipush 710
      // 0e3: iload 6
      // 0e5: if_icmpeq 31e
      // 0e8: goto 0ec
      // 0eb: athrow
      // 0ec: sipush -8241
      // 0ef: iload 6
      // 0f1: bipush -1
      // 0f2: ixor
      // 0f3: if_icmpne 10a
      // 0f6: goto 0fa
      // 0f9: athrow
      // 0fa: aload 4
      // 0fc: iload 5
      // 0fe: bipush -119
      // 100: bastore
      // 101: iload 7
      // 103: ifeq 38a
      // 106: goto 10a
      // 109: athrow
      // 10a: iload 6
      // 10c: sipush 352
      // 10f: if_icmpeq 30e
      // 112: goto 116
      // 115: athrow
      // 116: iload 6
      // 118: sipush 8249
      // 11b: if_icmpne 132
      // 11e: goto 122
      // 121: athrow
      // 122: aload 4
      // 124: iload 5
      // 126: bipush -117
      // 128: bastore
      // 129: iload 7
      // 12b: ifeq 38a
      // 12e: goto 132
      // 131: athrow
      // 132: iload 6
      // 134: sipush 338
      // 137: if_icmpne 14e
      // 13a: goto 13e
      // 13d: athrow
      // 13e: aload 4
      // 140: iload 5
      // 142: bipush -116
      // 144: bastore
      // 145: iload 7
      // 147: ifeq 38a
      // 14a: goto 14e
      // 14d: athrow
      // 14e: iload 6
      // 150: sipush 381
      // 153: if_icmpeq 2fe
      // 156: goto 15a
      // 159: athrow
      // 15a: sipush 8216
      // 15d: iload 6
      // 15f: if_icmpeq 2ee
      // 162: goto 166
      // 165: athrow
      // 166: sipush -8218
      // 169: iload 6
      // 16b: bipush -1
      // 16c: ixor
      // 16d: if_icmpeq 2de
      // 170: goto 174
      // 173: athrow
      // 174: sipush 8220
      // 177: iload 6
      // 179: if_icmpne 190
      // 17c: goto 180
      // 17f: athrow
      // 180: aload 4
      // 182: iload 5
      // 184: bipush -109
      // 186: bastore
      // 187: iload 7
      // 189: ifeq 38a
      // 18c: goto 190
      // 18f: athrow
      // 190: sipush 8221
      // 193: iload 6
      // 195: if_icmpne 1ac
      // 198: goto 19c
      // 19b: athrow
      // 19c: aload 4
      // 19e: iload 5
      // 1a0: bipush -108
      // 1a2: bastore
      // 1a3: iload 7
      // 1a5: ifeq 38a
      // 1a8: goto 1ac
      // 1ab: athrow
      // 1ac: iload 6
      // 1ae: sipush 8226
      // 1b1: if_icmpne 1c8
      // 1b4: goto 1b8
      // 1b7: athrow
      // 1b8: aload 4
      // 1ba: iload 5
      // 1bc: bipush -107
      // 1be: bastore
      // 1bf: iload 7
      // 1c1: ifeq 38a
      // 1c4: goto 1c8
      // 1c7: athrow
      // 1c8: sipush 8211
      // 1cb: iload 6
      // 1cd: if_icmpne 1e4
      // 1d0: goto 1d4
      // 1d3: athrow
      // 1d4: aload 4
      // 1d6: iload 5
      // 1d8: bipush -106
      // 1da: bastore
      // 1db: iload 7
      // 1dd: ifeq 38a
      // 1e0: goto 1e4
      // 1e3: athrow
      // 1e4: sipush 8212
      // 1e7: iload 6
      // 1e9: if_icmpne 200
      // 1ec: goto 1f0
      // 1ef: athrow
      // 1f0: aload 4
      // 1f2: iload 5
      // 1f4: bipush -105
      // 1f6: bastore
      // 1f7: iload 7
      // 1f9: ifeq 38a
      // 1fc: goto 200
      // 1ff: athrow
      // 200: sipush -733
      // 203: iload 6
      // 205: bipush -1
      // 206: ixor
      // 207: if_icmpeq 2ce
      // 20a: goto 20e
      // 20d: athrow
      // 20e: iload 6
      // 210: bipush -1
      // 211: ixor
      // 212: sipush -8483
      // 215: if_icmpeq 2be
      // 218: goto 21c
      // 21b: athrow
      // 21c: sipush -354
      // 21f: iload 6
      // 221: bipush -1
      // 222: ixor
      // 223: if_icmpeq 2ae
      // 226: goto 22a
      // 229: athrow
      // 22a: sipush 8250
      // 22d: iload 6
      // 22f: if_icmpeq 29e
      // 232: goto 236
      // 235: athrow
      // 236: iload 6
      // 238: bipush -1
      // 239: ixor
      // 23a: sipush -340
      // 23d: if_icmpeq 28e
      // 240: goto 244
      // 243: athrow
      // 244: sipush 382
      // 247: iload 6
      // 249: if_icmpeq 27e
      // 24c: goto 250
      // 24f: athrow
      // 250: iload 6
      // 252: bipush -1
      // 253: ixor
      // 254: sipush -377
      // 257: if_icmpne 26e
      // 25a: goto 25e
      // 25d: athrow
      // 25e: aload 4
      // 260: iload 5
      // 262: bipush -97
      // 264: bastore
      // 265: iload 7
      // 267: ifeq 38a
      // 26a: goto 26e
      // 26d: athrow
      // 26e: aload 4
      // 270: iload 5
      // 272: bipush 63
      // 274: bastore
      // 275: iload 7
      // 277: ifeq 38a
      // 27a: goto 27e
      // 27d: athrow
      // 27e: aload 4
      // 280: iload 5
      // 282: bipush -98
      // 284: bastore
      // 285: iload 7
      // 287: ifeq 38a
      // 28a: goto 28e
      // 28d: athrow
      // 28e: aload 4
      // 290: iload 5
      // 292: bipush -100
      // 294: bastore
      // 295: iload 7
      // 297: ifeq 38a
      // 29a: goto 29e
      // 29d: athrow
      // 29e: aload 4
      // 2a0: iload 5
      // 2a2: bipush -101
      // 2a4: bastore
      // 2a5: iload 7
      // 2a7: ifeq 38a
      // 2aa: goto 2ae
      // 2ad: athrow
      // 2ae: aload 4
      // 2b0: iload 5
      // 2b2: bipush -102
      // 2b4: bastore
      // 2b5: iload 7
      // 2b7: ifeq 38a
      // 2ba: goto 2be
      // 2bd: athrow
      // 2be: aload 4
      // 2c0: iload 5
      // 2c2: bipush -103
      // 2c4: bastore
      // 2c5: iload 7
      // 2c7: ifeq 38a
      // 2ca: goto 2ce
      // 2cd: athrow
      // 2ce: aload 4
      // 2d0: iload 5
      // 2d2: bipush -104
      // 2d4: bastore
      // 2d5: iload 7
      // 2d7: ifeq 38a
      // 2da: goto 2de
      // 2dd: athrow
      // 2de: aload 4
      // 2e0: iload 5
      // 2e2: bipush -110
      // 2e4: bastore
      // 2e5: iload 7
      // 2e7: ifeq 38a
      // 2ea: goto 2ee
      // 2ed: athrow
      // 2ee: aload 4
      // 2f0: iload 5
      // 2f2: bipush -111
      // 2f4: bastore
      // 2f5: iload 7
      // 2f7: ifeq 38a
      // 2fa: goto 2fe
      // 2fd: athrow
      // 2fe: aload 4
      // 300: iload 5
      // 302: bipush -114
      // 304: bastore
      // 305: iload 7
      // 307: ifeq 38a
      // 30a: goto 30e
      // 30d: athrow
      // 30e: aload 4
      // 310: iload 5
      // 312: bipush -118
      // 314: bastore
      // 315: iload 7
      // 317: ifeq 38a
      // 31a: goto 31e
      // 31d: athrow
      // 31e: aload 4
      // 320: iload 5
      // 322: bipush -120
      // 324: bastore
      // 325: iload 7
      // 327: ifeq 38a
      // 32a: goto 32e
      // 32d: athrow
      // 32e: aload 4
      // 330: iload 5
      // 332: bipush -121
      // 334: bastore
      // 335: iload 7
      // 337: ifeq 38a
      // 33a: goto 33e
      // 33d: athrow
      // 33e: aload 4
      // 340: iload 5
      // 342: bipush -122
      // 344: bastore
      // 345: iload 7
      // 347: ifeq 38a
      // 34a: goto 34e
      // 34d: athrow
      // 34e: aload 4
      // 350: iload 5
      // 352: bipush -124
      // 354: bastore
      // 355: iload 7
      // 357: ifeq 38a
      // 35a: goto 35e
      // 35d: athrow
      // 35e: aload 4
      // 360: iload 5
      // 362: bipush -125
      // 364: bastore
      // 365: iload 7
      // 367: ifeq 38a
      // 36a: goto 36e
      // 36d: athrow
      // 36e: aload 4
      // 370: iload 5
      // 372: bipush -128
      // 374: bastore
      // 375: iload 7
      // 377: ifeq 38a
      // 37a: goto 37e
      // 37d: athrow
      // 37e: aload 4
      // 380: iload 5
      // 382: iload 6
      // 384: i2b
      // 385: bastore
      // 386: goto 38a
      // 389: athrow
      // 38a: iinc 5 1
      // 38d: iload 7
      // 38f: ifeq 027
      // 392: aload 4
      // 394: areturn
      // 395: astore 2
      // 396: aload 2
      // 397: new java/lang/StringBuilder
      // 39a: dup
      // 39b: invokespecial java/lang/StringBuilder.<init> ()V
      // 39e: getstatic h.z [Ljava/lang/String;
      // 3a1: bipush 2
      // 3a2: aaload
      // 3a3: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 3a6: aload 0
      // 3a7: ifnull 3b3
      // 3aa: getstatic h.z [Ljava/lang/String;
      // 3ad: bipush 0
      // 3ae: aaload
      // 3af: goto 3b8
      // 3b2: athrow
      // 3b3: getstatic h.z [Ljava/lang/String;
      // 3b6: bipush 1
      // 3b7: aaload
      // 3b8: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 3bb: bipush 44
      // 3bd: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 3c0: iload 1
      // 3c1: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 3c4: bipush 41
      // 3c6: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 3c9: invokevirtual java/lang/StringBuilder.toString ()Ljava/lang/String;
      // 3cc: invokestatic i.a (Ljava/lang/Throwable;Ljava/lang/String;)Lla;
      // 3cf: athrow
   }

   private static char[] z(String var0) {
      char[] var10000 = var0.toCharArray();
      if (var10000.length < 2) {
         var10000[0] = (char)(var10000[0] ^ 'H');
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
               var10005 = 5;
               break;
            case 1:
               var10005 = 42;
               break;
            case 2:
               var10005 = 10;
               break;
            case 3:
               var10005 = 117;
               break;
            default:
               var10005 = 72;
         }

         var10001[var1] = (char)(var10004 ^ var10005);
      }

      return new String(var10001).intern();
   }
}
