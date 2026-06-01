final class p {
   static boolean d = true;
   static String[] e;
   static int b;
   static String[] c;
   static String[] a;
   static int f;
   private static final String[] z = new String[]{z(z("/ 3\t")), z(z("1{\u001eM")), z(z("$ \\\u000f\u000f")), z(z("/ 0\t"))};

   static final synchronized long a(int var0) {
      try {
         b++;
         if (var0 != 0) {
            return -57L;
         } else {
            long var1 = System.currentTimeMillis();

            try {
               if (~client.ze < ~var1) {
                  wb.w = wb.w + client.ze + -var1;
               }
            } catch (RuntimeException var3) {
               throw var3;
            }

            client.ze = var1;
            return wb.w + var1;
         }
      } catch (RuntimeException var4) {
         throw i.a(var4, z[0] + var0 + ')');
      }
   }

   static final void a(
      int param0,
      int param1,
      int param2,
      int param3,
      int param4,
      int[] param5,
      int param6,
      int param7,
      int param8,
      int param9,
      int[] param10,
      int param11,
      int param12,
      int param13,
      int param14
   ) {
      // $VF: Couldn't be decompiled
      // Please report this to the Vineflower issue tracker, at https://github.com/Vineflower/vineflower/issues with a copy of the class file (if you have the rights to distribute it!)
      // java.lang.RuntimeException: parsing failure!
      //   at org.jetbrains.java.decompiler.modules.decompiler.decompose.DomHelper.parseGraph(DomHelper.java:211)
      //   at org.jetbrains.java.decompiler.main.rels.MethodProcessor.codeToJava(MethodProcessor.java:166)
      //
      // Bytecode:
      // 000: getstatic client.vh Z
      // 003: istore 22
      // 005: getstatic p.f I
      // 008: bipush 1
      // 009: iadd
      // 00a: putstatic p.f I
      // 00d: bipush 0
      // 00e: iload 14
      // 010: if_icmplt 014
      // 013: return
      // 014: bipush 0
      // 015: istore 15
      // 017: bipush 0
      // 018: istore 16
      // 01a: iload 12
      // 01c: ifeq 032
      // 01f: iload 3
      // 020: iload 12
      // 022: idiv
      // 023: ldc -258111322
      // 025: ishl
      // 026: istore 16
      // 028: iload 8
      // 02a: iload 12
      // 02c: idiv
      // 02d: ldc 637317126
      // 02f: ishl
      // 030: istore 15
      // 032: iload 0
      // 033: bipush 2
      // 034: ishl
      // 035: istore 0
      // 036: iload 15
      // 038: bipush -1
      // 039: ixor
      // 03a: bipush -1
      // 03b: if_icmpgt 058
      // 03e: iload 15
      // 040: sipush 4032
      // 043: if_icmpgt 04e
      // 046: goto 04a
      // 049: athrow
      // 04a: goto 05b
      // 04d: athrow
      // 04e: sipush 4032
      // 051: istore 15
      // 053: iload 22
      // 055: ifeq 05b
      // 058: bipush 0
      // 059: istore 15
      // 05b: iload 1
      // 05c: ldc 1121159302
      // 05e: if_icmpeq 089
      // 061: bipush -69
      // 063: bipush 127
      // 065: bipush -20
      // 067: bipush -29
      // 069: bipush -78
      // 06b: aconst_null
      // 06c: checkcast [I
      // 06f: bipush 16
      // 071: bipush 2
      // 072: bipush -77
      // 074: bipush -5
      // 076: aconst_null
      // 077: checkcast [I
      // 07a: bipush 113
      // 07c: bipush -57
      // 07e: bipush 68
      // 080: bipush -87
      // 082: invokestatic p.a (IIIII[IIIII[IIIII)V
      // 085: goto 089
      // 088: athrow
      // 089: iload 14
      // 08b: istore 19
      // 08d: bipush 0
      // 08e: iload 19
      // 090: if_icmpge 48f
      // 093: iload 12
      // 095: iload 13
      // 097: iadd
      // 098: istore 12
      // 09a: iload 8
      // 09c: iload 4
      // 09e: iadd
      // 09f: istore 8
      // 0a1: iload 3
      // 0a2: iload 2
      // 0a3: iadd
      // 0a4: istore 3
      // 0a5: iload 15
      // 0a7: istore 9
      // 0a9: iload 16
      // 0ab: istore 7
      // 0ad: iload 22
      // 0af: ifne 560
      // 0b2: bipush -1
      // 0b3: iload 12
      // 0b5: bipush -1
      // 0b6: ixor
      // 0b7: if_icmpne 0c2
      // 0ba: goto 0be
      // 0bd: athrow
      // 0be: goto 0d5
      // 0c1: athrow
      // 0c2: iload 8
      // 0c4: iload 12
      // 0c6: idiv
      // 0c7: ldc 25779686
      // 0c9: ishl
      // 0ca: istore 15
      // 0cc: iload 3
      // 0cd: iload 12
      // 0cf: idiv
      // 0d0: ldc 1121159302
      // 0d2: ishl
      // 0d3: istore 16
      // 0d5: bipush -1
      // 0d6: iload 15
      // 0d8: bipush -1
      // 0d9: ixor
      // 0da: if_icmpge 0e5
      // 0dd: bipush 0
      // 0de: istore 15
      // 0e0: iload 22
      // 0e2: ifeq 0fc
      // 0e5: iload 15
      // 0e7: bipush -1
      // 0e8: ixor
      // 0e9: sipush -4033
      // 0ec: if_icmplt 0f7
      // 0ef: goto 0f3
      // 0f2: athrow
      // 0f3: goto 0fc
      // 0f6: athrow
      // 0f7: sipush 4032
      // 0fa: istore 15
      // 0fc: iload 16
      // 0fe: iload 7
      // 100: ineg
      // 101: iadd
      // 102: ldc 1397627332
      // 104: ishr
      // 105: istore 18
      // 107: iload 9
      // 109: ineg
      // 10a: iload 15
      // 10c: iadd
      // 10d: ldc 1542826468
      // 10f: ishr
      // 110: istore 17
      // 112: iload 6
      // 114: ldc -1668610924
      // 116: ishr
      // 117: istore 20
      // 119: iload 9
      // 11b: ldc 786432
      // 11d: iload 6
      // 11f: iand
      // 120: iadd
      // 121: istore 9
      // 123: iload 6
      // 125: iload 0
      // 126: iadd
      // 127: istore 6
      // 129: iload 19
      // 12b: bipush -1
      // 12c: ixor
      // 12d: bipush -17
      // 12f: if_icmpgt 420
      // 132: aload 10
      // 134: iload 11
      // 136: iinc 11 1
      // 139: aload 5
      // 13b: iload 7
      // 13d: sipush 4032
      // 140: invokestatic ib.a (II)I
      // 143: iload 9
      // 145: ldc -1148525818
      // 147: ishr
      // 148: ineg
      // 149: isub
      // 14a: iaload
      // 14b: iload 20
      // 14d: iushr
      // 14e: iastore
      // 14f: iload 7
      // 151: iload 18
      // 153: iadd
      // 154: istore 7
      // 156: iload 9
      // 158: iload 17
      // 15a: iadd
      // 15b: istore 9
      // 15d: aload 10
      // 15f: iload 11
      // 161: iinc 11 1
      // 164: aload 5
      // 166: iload 9
      // 168: ldc 1034190278
      // 16a: ishr
      // 16b: iload 7
      // 16d: sipush 4032
      // 170: invokestatic ib.a (II)I
      // 173: iadd
      // 174: iaload
      // 175: iload 20
      // 177: iushr
      // 178: iastore
      // 179: iload 7
      // 17b: iload 18
      // 17d: iadd
      // 17e: istore 7
      // 180: iload 9
      // 182: iload 17
      // 184: iadd
      // 185: istore 9
      // 187: aload 10
      // 189: iload 11
      // 18b: iinc 11 1
      // 18e: aload 5
      // 190: iload 9
      // 192: ldc -385010618
      // 194: ishr
      // 195: sipush 4032
      // 198: iload 7
      // 19a: invokestatic ib.a (II)I
      // 19d: iadd
      // 19e: iaload
      // 19f: iload 20
      // 1a1: iushr
      // 1a2: iastore
      // 1a3: iload 9
      // 1a5: iload 17
      // 1a7: iadd
      // 1a8: istore 9
      // 1aa: iload 7
      // 1ac: iload 18
      // 1ae: iadd
      // 1af: istore 7
      // 1b1: aload 10
      // 1b3: iload 11
      // 1b5: iinc 11 1
      // 1b8: aload 5
      // 1ba: iload 9
      // 1bc: ldc 747209702
      // 1be: ishr
      // 1bf: sipush 4032
      // 1c2: iload 7
      // 1c4: invokestatic ib.a (II)I
      // 1c7: iadd
      // 1c8: iaload
      // 1c9: iload 20
      // 1cb: iushr
      // 1cc: iastore
      // 1cd: iload 9
      // 1cf: iload 17
      // 1d1: iadd
      // 1d2: istore 9
      // 1d4: iload 7
      // 1d6: iload 18
      // 1d8: iadd
      // 1d9: istore 7
      // 1db: iload 6
      // 1dd: ldc 597207284
      // 1df: ishr
      // 1e0: istore 20
      // 1e2: iload 6
      // 1e4: ldc 786432
      // 1e6: iand
      // 1e7: sipush 4095
      // 1ea: iload 9
      // 1ec: iand
      // 1ed: iadd
      // 1ee: istore 9
      // 1f0: iload 6
      // 1f2: iload 0
      // 1f3: iadd
      // 1f4: istore 6
      // 1f6: aload 10
      // 1f8: iload 11
      // 1fa: iinc 11 1
      // 1fd: aload 5
      // 1ff: sipush 4032
      // 202: iload 7
      // 204: invokestatic ib.a (II)I
      // 207: iload 9
      // 209: ldc 831423910
      // 20b: ishr
      // 20c: iadd
      // 20d: iaload
      // 20e: iload 20
      // 210: iushr
      // 211: iastore
      // 212: iload 7
      // 214: iload 18
      // 216: iadd
      // 217: istore 7
      // 219: iload 9
      // 21b: iload 17
      // 21d: iadd
      // 21e: istore 9
      // 220: aload 10
      // 222: iload 11
      // 224: iinc 11 1
      // 227: aload 5
      // 229: iload 7
      // 22b: sipush 4032
      // 22e: invokestatic ib.a (II)I
      // 231: iload 9
      // 233: ldc -512409978
      // 235: ishr
      // 236: ineg
      // 237: isub
      // 238: iaload
      // 239: iload 20
      // 23b: iushr
      // 23c: iastore
      // 23d: iload 9
      // 23f: iload 17
      // 241: iadd
      // 242: istore 9
      // 244: iload 7
      // 246: iload 18
      // 248: iadd
      // 249: istore 7
      // 24b: aload 10
      // 24d: iload 11
      // 24f: iinc 11 1
      // 252: aload 5
      // 254: iload 7
      // 256: sipush 4032
      // 259: invokestatic ib.a (II)I
      // 25c: iload 9
      // 25e: ldc -783757370
      // 260: ishr
      // 261: iadd
      // 262: iaload
      // 263: iload 20
      // 265: iushr
      // 266: iastore
      // 267: iload 7
      // 269: iload 18
      // 26b: iadd
      // 26c: istore 7
      // 26e: iload 9
      // 270: iload 17
      // 272: iadd
      // 273: istore 9
      // 275: aload 10
      // 277: iload 11
      // 279: iinc 11 1
      // 27c: aload 5
      // 27e: iload 9
      // 280: ldc -129948154
      // 282: ishr
      // 283: sipush 4032
      // 286: iload 7
      // 288: invokestatic ib.a (II)I
      // 28b: iadd
      // 28c: iaload
      // 28d: iload 20
      // 28f: iushr
      // 290: iastore
      // 291: iload 9
      // 293: iload 17
      // 295: iadd
      // 296: istore 9
      // 298: iload 7
      // 29a: iload 18
      // 29c: iadd
      // 29d: istore 7
      // 29f: iload 6
      // 2a1: ldc 92466196
      // 2a3: ishr
      // 2a4: istore 20
      // 2a6: ldc 786432
      // 2a8: iload 6
      // 2aa: iand
      // 2ab: sipush 4095
      // 2ae: iload 9
      // 2b0: iand
      // 2b1: iadd
      // 2b2: istore 9
      // 2b4: iload 6
      // 2b6: iload 0
      // 2b7: iadd
      // 2b8: istore 6
      // 2ba: aload 10
      // 2bc: iload 11
      // 2be: iinc 11 1
      // 2c1: aload 5
      // 2c3: iload 7
      // 2c5: sipush 4032
      // 2c8: invokestatic ib.a (II)I
      // 2cb: iload 9
      // 2cd: ldc -1989449594
      // 2cf: ishr
      // 2d0: ineg
      // 2d1: isub
      // 2d2: iaload
      // 2d3: iload 20
      // 2d5: iushr
      // 2d6: iastore
      // 2d7: iload 7
      // 2d9: iload 18
      // 2db: iadd
      // 2dc: istore 7
      // 2de: iload 9
      // 2e0: iload 17
      // 2e2: iadd
      // 2e3: istore 9
      // 2e5: aload 10
      // 2e7: iload 11
      // 2e9: iinc 11 1
      // 2ec: aload 5
      // 2ee: iload 9
      // 2f0: ldc -76155226
      // 2f2: ishr
      // 2f3: iload 7
      // 2f5: sipush 4032
      // 2f8: invokestatic ib.a (II)I
      // 2fb: iadd
      // 2fc: iaload
      // 2fd: iload 20
      // 2ff: iushr
      // 300: iastore
      // 301: iload 7
      // 303: iload 18
      // 305: iadd
      // 306: istore 7
      // 308: iload 9
      // 30a: iload 17
      // 30c: iadd
      // 30d: istore 9
      // 30f: aload 10
      // 311: iload 11
      // 313: iinc 11 1
      // 316: aload 5
      // 318: iload 7
      // 31a: sipush 4032
      // 31d: invokestatic ib.a (II)I
      // 320: iload 9
      // 322: ldc -158732986
      // 324: ishr
      // 325: ineg
      // 326: isub
      // 327: iaload
      // 328: iload 20
      // 32a: iushr
      // 32b: iastore
      // 32c: iload 7
      // 32e: iload 18
      // 330: iadd
      // 331: istore 7
      // 333: iload 9
      // 335: iload 17
      // 337: iadd
      // 338: istore 9
      // 33a: aload 10
      // 33c: iload 11
      // 33e: iinc 11 1
      // 341: aload 5
      // 343: sipush 4032
      // 346: iload 7
      // 348: invokestatic ib.a (II)I
      // 34b: iload 9
      // 34d: ldc 1960099526
      // 34f: ishr
      // 350: iadd
      // 351: iaload
      // 352: iload 20
      // 354: iushr
      // 355: iastore
      // 356: iload 7
      // 358: iload 18
      // 35a: iadd
      // 35b: istore 7
      // 35d: iload 9
      // 35f: iload 17
      // 361: iadd
      // 362: istore 9
      // 364: iload 6
      // 366: ldc 1740031764
      // 368: ishr
      // 369: istore 20
      // 36b: sipush 4095
      // 36e: iload 9
      // 370: iand
      // 371: iload 6
      // 373: ldc 786432
      // 375: iand
      // 376: ineg
      // 377: isub
      // 378: istore 9
      // 37a: iload 6
      // 37c: iload 0
      // 37d: iadd
      // 37e: istore 6
      // 380: aload 10
      // 382: iload 11
      // 384: iinc 11 1
      // 387: aload 5
      // 389: iload 7
      // 38b: sipush 4032
      // 38e: invokestatic ib.a (II)I
      // 391: iload 9
      // 393: ldc -1366214458
      // 395: ishr
      // 396: iadd
      // 397: iaload
      // 398: iload 20
      // 39a: iushr
      // 39b: iastore
      // 39c: iload 7
      // 39e: iload 18
      // 3a0: iadd
      // 3a1: istore 7
      // 3a3: iload 9
      // 3a5: iload 17
      // 3a7: iadd
      // 3a8: istore 9
      // 3aa: aload 10
      // 3ac: iload 11
      // 3ae: iinc 11 1
      // 3b1: aload 5
      // 3b3: iload 9
      // 3b5: ldc -1971655962
      // 3b7: ishr
      // 3b8: iload 7
      // 3ba: sipush 4032
      // 3bd: invokestatic ib.a (II)I
      // 3c0: iadd
      // 3c1: iaload
      // 3c2: iload 20
      // 3c4: iushr
      // 3c5: iastore
      // 3c6: iload 7
      // 3c8: iload 18
      // 3ca: iadd
      // 3cb: istore 7
      // 3cd: iload 9
      // 3cf: iload 17
      // 3d1: iadd
      // 3d2: istore 9
      // 3d4: aload 10
      // 3d6: iload 11
      // 3d8: iinc 11 1
      // 3db: aload 5
      // 3dd: iload 9
      // 3df: ldc -674692218
      // 3e1: ishr
      // 3e2: iload 7
      // 3e4: sipush 4032
      // 3e7: invokestatic ib.a (II)I
      // 3ea: iadd
      // 3eb: iaload
      // 3ec: iload 20
      // 3ee: iushr
      // 3ef: iastore
      // 3f0: iload 9
      // 3f2: iload 17
      // 3f4: iadd
      // 3f5: istore 9
      // 3f7: iload 7
      // 3f9: iload 18
      // 3fb: iadd
      // 3fc: istore 7
      // 3fe: aload 10
      // 400: iload 11
      // 402: iinc 11 1
      // 405: aload 5
      // 407: sipush 4032
      // 40a: iload 7
      // 40c: invokestatic ib.a (II)I
      // 40f: iload 9
      // 411: ldc -1886734874
      // 413: ishr
      // 414: ineg
      // 415: isub
      // 416: iaload
      // 417: iload 20
      // 419: iushr
      // 41a: iastore
      // 41b: iload 22
      // 41d: ifeq 487
      // 420: bipush 0
      // 421: istore 21
      // 423: iload 21
      // 425: iload 19
      // 427: if_icmpge 487
      // 42a: aload 10
      // 42c: iload 11
      // 42e: iinc 11 1
      // 431: aload 5
      // 433: iload 9
      // 435: ldc -1102323802
      // 437: ishr
      // 438: sipush 4032
      // 43b: iload 7
      // 43d: invokestatic ib.a (II)I
      // 440: iadd
      // 441: iaload
      // 442: iload 20
      // 444: iushr
      // 445: iastore
      // 446: iload 7
      // 448: iload 18
      // 44a: iadd
      // 44b: istore 7
      // 44d: iload 9
      // 44f: iload 17
      // 451: iadd
      // 452: istore 9
      // 454: bipush 3
      // 455: bipush 3
      // 456: iload 21
      // 458: iand
      // 459: iload 22
      // 45b: ifne 090
      // 45e: if_icmpeq 464
      // 461: goto 47f
      // 464: iload 6
      // 466: ldc -667502188
      // 468: ishr
      // 469: istore 20
      // 46b: iload 6
      // 46d: ldc 786432
      // 46f: iand
      // 470: sipush 4095
      // 473: iload 9
      // 475: iand
      // 476: iadd
      // 477: istore 9
      // 479: iload 6
      // 47b: iload 0
      // 47c: iadd
      // 47d: istore 6
      // 47f: iinc 21 1
      // 482: iload 22
      // 484: ifeq 423
      // 487: iinc 19 -16
      // 48a: iload 22
      // 48c: ifeq 08d
      // 48f: goto 560
      // 492: astore 15
      // 494: aload 15
      // 496: new java/lang/StringBuilder
      // 499: dup
      // 49a: invokespecial java/lang/StringBuilder.<init> ()V
      // 49d: getstatic p.z [Ljava/lang/String;
      // 4a0: bipush 3
      // 4a1: aaload
      // 4a2: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 4a5: iload 0
      // 4a6: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 4a9: bipush 44
      // 4ab: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 4ae: iload 1
      // 4af: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 4b2: bipush 44
      // 4b4: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 4b7: iload 2
      // 4b8: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 4bb: bipush 44
      // 4bd: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 4c0: iload 3
      // 4c1: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 4c4: bipush 44
      // 4c6: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 4c9: iload 4
      // 4cb: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 4ce: bipush 44
      // 4d0: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 4d3: aload 5
      // 4d5: ifnull 4e1
      // 4d8: getstatic p.z [Ljava/lang/String;
      // 4db: bipush 2
      // 4dc: aaload
      // 4dd: goto 4e6
      // 4e0: athrow
      // 4e1: getstatic p.z [Ljava/lang/String;
      // 4e4: bipush 1
      // 4e5: aaload
      // 4e6: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 4e9: bipush 44
      // 4eb: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 4ee: iload 6
      // 4f0: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 4f3: bipush 44
      // 4f5: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 4f8: iload 7
      // 4fa: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 4fd: bipush 44
      // 4ff: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 502: iload 8
      // 504: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 507: bipush 44
      // 509: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 50c: iload 9
      // 50e: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 511: bipush 44
      // 513: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 516: aload 10
      // 518: ifnull 524
      // 51b: getstatic p.z [Ljava/lang/String;
      // 51e: bipush 2
      // 51f: aaload
      // 520: goto 529
      // 523: athrow
      // 524: getstatic p.z [Ljava/lang/String;
      // 527: bipush 1
      // 528: aaload
      // 529: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 52c: bipush 44
      // 52e: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 531: iload 11
      // 533: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 536: bipush 44
      // 538: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 53b: iload 12
      // 53d: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 540: bipush 44
      // 542: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 545: iload 13
      // 547: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 54a: bipush 44
      // 54c: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 54f: iload 14
      // 551: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 554: bipush 41
      // 556: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 559: invokevirtual java/lang/StringBuilder.toString ()Ljava/lang/String;
      // 55c: invokestatic i.a (Ljava/lang/Throwable;Ljava/lang/String;)Lla;
      // 55f: athrow
      // 560: return
   }

   private static char[] z(String var0) {
      char[] var10000 = var0.toCharArray();
      if (var10000.length < 2) {
         var10000[0] = (char)(var10000[0] ^ 'r');
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
               var10005 = 95;
               break;
            case 1:
               var10005 = 14;
               break;
            case 2:
               var10005 = 114;
               break;
            case 3:
               var10005 = 33;
               break;
            default:
               var10005 = 114;
         }

         var10001[var1] = (char)(var10004 ^ var10005);
      }

      return new String(var10001).intern();
   }
}
