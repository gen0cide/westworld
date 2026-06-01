import java.awt.Font;
import java.awt.FontMetrics;
import java.math.BigInteger;

final class s {
   static int b;
   static String[] e = new String[]{
      z(z(".;\u0016f\u001cK;\u0017n\f\u000e'Bl\bK<\u0016f\u0003\u0018u\u0016lN\u00043\u0004f\u001cK4\fgN\u001b'\u0007p\u001dK0\fw\u000b\u0019"))
   };
   static String[] f;
   static BigInteger c = new BigInteger(z(z("ZeR3_")), 16);
   static nb a = null;
   static int d = 0;
   private static final String[] z = new String[]{z(z("\u0018{#+")), z(z("\u0010{L-\u0013")), z(z("\u0005 \u000eo"))};

   static final boolean a(int param0, Font param1, int param2, int param3, e param4, char param5, FontMetrics param6, boolean param7) {
      // $VF: Couldn't be decompiled
      // Please report this to the Vineflower issue tracker, at https://github.com/Vineflower/vineflower/issues with a copy of the class file (if you have the rights to distribute it!)
      // java.lang.RuntimeException: parsing failure!
      //   at org.jetbrains.java.decompiler.modules.decompiler.decompose.DomHelper.parseGraph(DomHelper.java:211)
      //   at org.jetbrains.java.decompiler.main.rels.MethodProcessor.codeToJava(MethodProcessor.java:166)
      //
      // Bytecode:
      // 000: getstatic client.vh Z
      // 003: istore 24
      // 005: getstatic s.b I
      // 008: bipush 1
      // 009: iadd
      // 00a: putstatic s.b I
      // 00d: aload 6
      // 00f: iload 5
      // 011: invokevirtual java/awt/FontMetrics.charWidth (C)I
      // 014: istore 8
      // 016: iload 8
      // 018: istore 9
      // 01a: iload 7
      // 01c: ifne 023
      // 01f: goto 0ad
      // 022: athrow
      // 023: iload 5
      // 025: bipush 102
      // 027: if_icmpeq 095
      // 02a: iload 5
      // 02c: bipush 116
      // 02e: if_icmpeq 095
      // 031: goto 035
      // 034: athrow
      // 035: bipush 119
      // 037: iload 5
      // 039: if_icmpeq 095
      // 03c: goto 040
      // 03f: athrow
      // 040: iload 5
      // 042: bipush -1
      // 043: ixor
      // 044: bipush -119
      // 046: if_icmpeq 095
      // 049: goto 04d
      // 04c: athrow
      // 04d: iload 5
      // 04f: bipush 107
      // 051: if_icmpeq 095
      // 054: goto 058
      // 057: athrow
      // 058: iload 5
      // 05a: bipush -1
      // 05b: ixor
      // 05c: bipush -121
      // 05e: if_icmpeq 095
      // 061: goto 065
      // 064: athrow
      // 065: bipush 121
      // 067: iload 5
      // 069: if_icmpeq 095
      // 06c: goto 070
      // 06f: athrow
      // 070: iload 5
      // 072: bipush 65
      // 074: if_icmpeq 095
      // 077: goto 07b
      // 07a: athrow
      // 07b: iload 5
      // 07d: bipush -1
      // 07e: ixor
      // 07f: bipush -87
      // 081: if_icmpeq 095
      // 084: goto 088
      // 087: athrow
      // 088: iload 5
      // 08a: bipush -1
      // 08b: ixor
      // 08c: bipush -88
      // 08e: if_icmpne 09c
      // 091: goto 095
      // 094: athrow
      // 095: iinc 8 1
      // 098: goto 09c
      // 09b: athrow
      // 09c: iload 5
      // 09e: bipush -1
      // 09f: ixor
      // 0a0: bipush -48
      // 0a2: if_icmpne 0a8
      // 0a5: bipush 0
      // 0a6: istore 7
      // 0a8: goto 0ad
      // 0ab: astore 10
      // 0ad: aload 6
      // 0af: invokevirtual java/awt/FontMetrics.getMaxAscent ()I
      // 0b2: istore 10
      // 0b4: aload 6
      // 0b6: invokevirtual java/awt/FontMetrics.getMaxAscent ()I
      // 0b9: aload 6
      // 0bb: invokevirtual java/awt/FontMetrics.getMaxDescent ()I
      // 0be: ineg
      // 0bf: isub
      // 0c0: istore 11
      // 0c2: aload 6
      // 0c4: invokevirtual java/awt/FontMetrics.getHeight ()I
      // 0c7: istore 12
      // 0c9: aload 4
      // 0cb: iload 8
      // 0cd: iload 11
      // 0cf: invokevirtual e.createImage (II)Ljava/awt/Image;
      // 0d2: astore 13
      // 0d4: aconst_null
      // 0d5: aload 13
      // 0d7: if_acmpne 0dc
      // 0da: bipush 0
      // 0db: ireturn
      // 0dc: aload 13
      // 0de: invokevirtual java/awt/Image.getGraphics ()Ljava/awt/Graphics;
      // 0e1: astore 14
      // 0e3: aload 14
      // 0e5: getstatic java/awt/Color.black Ljava/awt/Color;
      // 0e8: invokevirtual java/awt/Graphics.setColor (Ljava/awt/Color;)V
      // 0eb: aload 14
      // 0ed: bipush 0
      // 0ee: bipush 0
      // 0ef: iload 8
      // 0f1: iload 11
      // 0f3: invokevirtual java/awt/Graphics.fillRect (IIII)V
      // 0f6: aload 14
      // 0f8: getstatic java/awt/Color.white Ljava/awt/Color;
      // 0fb: invokevirtual java/awt/Graphics.setColor (Ljava/awt/Color;)V
      // 0fe: aload 14
      // 100: aload 1
      // 101: invokevirtual java/awt/Graphics.setFont (Ljava/awt/Font;)V
      // 104: aload 14
      // 106: new java/lang/StringBuilder
      // 109: dup
      // 10a: invokespecial java/lang/StringBuilder.<init> ()V
      // 10d: iload 5
      // 10f: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 112: ldc ""
      // 114: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 117: invokevirtual java/lang/StringBuilder.toString ()Ljava/lang/String;
      // 11a: bipush 0
      // 11b: iload 10
      // 11d: invokevirtual java/awt/Graphics.drawString (Ljava/lang/String;II)V
      // 120: iload 7
      // 122: ifne 129
      // 125: goto 145
      // 128: athrow
      // 129: aload 14
      // 12b: new java/lang/StringBuilder
      // 12e: dup
      // 12f: invokespecial java/lang/StringBuilder.<init> ()V
      // 132: iload 5
      // 134: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 137: ldc ""
      // 139: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 13c: invokevirtual java/lang/StringBuilder.toString ()Ljava/lang/String;
      // 13f: bipush 1
      // 140: iload 10
      // 142: invokevirtual java/awt/Graphics.drawString (Ljava/lang/String;II)V
      // 145: iload 11
      // 147: iload 8
      // 149: imul
      // 14a: newarray 10
      // 14c: astore 15
      // 14e: new java/awt/image/PixelGrabber
      // 151: dup
      // 152: aload 13
      // 154: bipush 0
      // 155: bipush 0
      // 156: iload 8
      // 158: iload 11
      // 15a: aload 15
      // 15c: bipush 0
      // 15d: iload 8
      // 15f: invokespecial java/awt/image/PixelGrabber.<init> (Ljava/awt/Image;IIII[III)V
      // 162: astore 16
      // 164: aload 16
      // 166: invokevirtual java/awt/image/PixelGrabber.grabPixels ()Z
      // 169: pop
      // 16a: goto 171
      // 16d: astore 17
      // 16f: bipush 0
      // 170: ireturn
      // 171: aload 13
      // 173: invokevirtual java/awt/Image.flush ()V
      // 176: aconst_null
      // 177: astore 13
      // 179: bipush 0
      // 17a: istore 17
      // 17c: bipush 0
      // 17d: istore 18
      // 17f: iload 8
      // 181: istore 19
      // 183: iload 11
      // 185: istore 20
      // 187: bipush 0
      // 188: istore 21
      // 18a: iload 21
      // 18c: iload 11
      // 18e: if_icmpge 1e0
      // 191: bipush 0
      // 192: iload 24
      // 194: ifne 1e1
      // 197: istore 22
      // 199: iload 8
      // 19b: bipush -1
      // 19c: ixor
      // 19d: iload 22
      // 19f: bipush -1
      // 1a0: ixor
      // 1a1: if_icmpge 1d8
      // 1a4: aload 15
      // 1a6: iload 22
      // 1a8: iload 21
      // 1aa: iload 8
      // 1ac: imul
      // 1ad: ineg
      // 1ae: isub
      // 1af: iaload
      // 1b0: istore 23
      // 1b2: bipush 0
      // 1b3: ldc 16777215
      // 1b5: iload 23
      // 1b7: iand
      // 1b8: iload 24
      // 1ba: ifne 18e
      // 1bd: if_icmpne 1c3
      // 1c0: goto 1cc
      // 1c3: iload 21
      // 1c5: istore 18
      // 1c7: iload 24
      // 1c9: ifeq 1e0
      // 1cc: iinc 22 1
      // 1cf: iload 24
      // 1d1: ifeq 199
      // 1d4: goto 1d8
      // 1d7: athrow
      // 1d8: iinc 21 1
      // 1db: iload 24
      // 1dd: ifeq 18a
      // 1e0: bipush 0
      // 1e1: istore 21
      // 1e3: iload 21
      // 1e5: iload 8
      // 1e7: if_icmpge 235
      // 1ea: bipush 0
      // 1eb: iload 24
      // 1ed: ifne 23c
      // 1f0: istore 22
      // 1f2: iload 11
      // 1f4: iload 22
      // 1f6: if_icmple 22d
      // 1f9: aload 15
      // 1fb: iload 21
      // 1fd: iload 22
      // 1ff: iload 8
      // 201: imul
      // 202: ineg
      // 203: isub
      // 204: iaload
      // 205: istore 23
      // 207: bipush 0
      // 208: iload 23
      // 20a: ldc 16777215
      // 20c: iand
      // 20d: iload 24
      // 20f: ifne 1e7
      // 212: if_icmpne 218
      // 215: goto 221
      // 218: iload 21
      // 21a: istore 17
      // 21c: iload 24
      // 21e: ifeq 235
      // 221: iinc 22 1
      // 224: iload 24
      // 226: ifeq 1f2
      // 229: goto 22d
      // 22c: athrow
      // 22d: iinc 21 1
      // 230: iload 24
      // 232: ifeq 1e3
      // 235: iload 11
      // 237: bipush 1
      // 238: isub
      // 239: istore 21
      // 23b: iload 3
      // 23c: bipush -86
      // 23e: if_icmplt 25e
      // 241: bipush -60
      // 243: aconst_null
      // 244: checkcast java/awt/Font
      // 247: bipush 49
      // 249: bipush -85
      // 24b: aconst_null
      // 24c: checkcast e
      // 24f: ldc 65528
      // 251: aconst_null
      // 252: checkcast java/awt/FontMetrics
      // 255: bipush 1
      // 256: invokestatic s.a (ILjava/awt/Font;IILe;CLjava/awt/FontMetrics;Z)Z
      // 259: pop
      // 25a: goto 25e
      // 25d: athrow
      // 25e: bipush 0
      // 25f: iload 21
      // 261: if_icmpgt 2b3
      // 264: bipush 0
      // 265: iload 24
      // 267: ifne 2b7
      // 26a: istore 22
      // 26c: iload 22
      // 26e: bipush -1
      // 26f: ixor
      // 270: iload 8
      // 272: bipush -1
      // 273: ixor
      // 274: if_icmple 2ab
      // 277: aload 15
      // 279: iload 22
      // 27b: iload 21
      // 27d: iload 8
      // 27f: imul
      // 280: iadd
      // 281: iaload
      // 282: istore 23
      // 284: iload 23
      // 286: ldc 16777215
      // 288: iand
      // 289: bipush -1
      // 28a: ixor
      // 28b: bipush -1
      // 28c: iload 24
      // 28e: ifne 261
      // 291: if_icmpeq 29f
      // 294: bipush 1
      // 295: iload 21
      // 297: iadd
      // 298: istore 20
      // 29a: iload 24
      // 29c: ifeq 2b3
      // 29f: iinc 22 1
      // 2a2: iload 24
      // 2a4: ifeq 26c
      // 2a7: goto 2ab
      // 2aa: athrow
      // 2ab: iinc 21 -1
      // 2ae: iload 24
      // 2b0: ifeq 25e
      // 2b3: bipush -1
      // 2b4: iload 8
      // 2b6: iadd
      // 2b7: istore 21
      // 2b9: bipush -1
      // 2ba: iload 21
      // 2bc: bipush -1
      // 2bd: ixor
      // 2be: if_icmplt 310
      // 2c1: bipush 0
      // 2c2: iload 24
      // 2c4: ifne 3ad
      // 2c7: istore 22
      // 2c9: iload 22
      // 2cb: bipush -1
      // 2cc: ixor
      // 2cd: iload 11
      // 2cf: bipush -1
      // 2d0: ixor
      // 2d1: if_icmple 308
      // 2d4: aload 15
      // 2d6: iload 21
      // 2d8: iload 22
      // 2da: iload 8
      // 2dc: imul
      // 2dd: iadd
      // 2de: iaload
      // 2df: istore 23
      // 2e1: bipush -1
      // 2e2: ldc 16777215
      // 2e4: iload 23
      // 2e6: iand
      // 2e7: bipush -1
      // 2e8: ixor
      // 2e9: iload 24
      // 2eb: ifne 2be
      // 2ee: if_icmpeq 2fc
      // 2f1: iload 21
      // 2f3: bipush 1
      // 2f4: iadd
      // 2f5: istore 19
      // 2f7: iload 24
      // 2f9: ifeq 310
      // 2fc: iinc 22 1
      // 2ff: iload 24
      // 301: ifeq 2c9
      // 304: goto 308
      // 307: athrow
      // 308: iinc 21 -1
      // 30b: iload 24
      // 30d: ifeq 2b9
      // 310: getstatic qb.k [B
      // 313: bipush 0
      // 314: bipush 9
      // 316: iload 2
      // 317: imul
      // 318: iadd
      // 319: getstatic b.c I
      // 31c: sipush 16384
      // 31f: idiv
      // 320: i2b
      // 321: bastore
      // 322: getstatic qb.k [B
      // 325: iload 2
      // 326: bipush 9
      // 328: imul
      // 329: bipush -1
      // 32a: isub
      // 32b: getstatic b.c I
      // 32e: sipush 128
      // 331: idiv
      // 332: bipush 127
      // 334: invokestatic ib.a (II)I
      // 337: i2b
      // 338: bastore
      // 339: getstatic qb.k [B
      // 33c: bipush 2
      // 33d: bipush 9
      // 33f: iload 2
      // 340: imul
      // 341: iadd
      // 342: getstatic b.c I
      // 345: bipush 127
      // 347: invokestatic ib.a (II)I
      // 34a: i2b
      // 34b: bastore
      // 34c: getstatic qb.k [B
      // 34f: bipush 3
      // 350: iload 2
      // 351: bipush 9
      // 353: imul
      // 354: iadd
      // 355: iload 17
      // 357: ineg
      // 358: iload 19
      // 35a: iadd
      // 35b: i2b
      // 35c: bastore
      // 35d: getstatic qb.k [B
      // 360: iload 2
      // 361: bipush 9
      // 363: imul
      // 364: bipush -4
      // 366: isub
      // 367: iload 20
      // 369: iload 18
      // 36b: ineg
      // 36c: iadd
      // 36d: i2b
      // 36e: bastore
      // 36f: getstatic qb.k [B
      // 372: bipush 9
      // 374: iload 2
      // 375: imul
      // 376: bipush -5
      // 378: isub
      // 379: iload 17
      // 37b: i2b
      // 37c: bastore
      // 37d: getstatic qb.k [B
      // 380: bipush 6
      // 382: iload 2
      // 383: bipush 9
      // 385: imul
      // 386: iadd
      // 387: iload 10
      // 389: iload 18
      // 38b: ineg
      // 38c: iadd
      // 38d: i2b
      // 38e: bastore
      // 38f: getstatic qb.k [B
      // 392: bipush 7
      // 394: bipush 9
      // 396: iload 2
      // 397: imul
      // 398: iadd
      // 399: iload 9
      // 39b: i2b
      // 39c: bastore
      // 39d: getstatic qb.k [B
      // 3a0: iload 2
      // 3a1: bipush 9
      // 3a3: imul
      // 3a4: bipush -8
      // 3a6: isub
      // 3a7: iload 12
      // 3a9: i2b
      // 3aa: bastore
      // 3ab: iload 18
      // 3ad: istore 21
      // 3af: iload 20
      // 3b1: iload 21
      // 3b3: if_icmple 41f
      // 3b6: iload 17
      // 3b8: iload 24
      // 3ba: ifne 420
      // 3bd: istore 22
      // 3bf: iload 22
      // 3c1: bipush -1
      // 3c2: ixor
      // 3c3: iload 19
      // 3c5: bipush -1
      // 3c6: ixor
      // 3c7: if_icmple 417
      // 3ca: sipush 255
      // 3cd: aload 15
      // 3cf: iload 8
      // 3d1: iload 21
      // 3d3: imul
      // 3d4: iload 22
      // 3d6: iadd
      // 3d7: iaload
      // 3d8: iand
      // 3d9: istore 23
      // 3db: bipush 30
      // 3dd: iload 23
      // 3df: iload 24
      // 3e1: ifne 3b3
      // 3e4: if_icmpge 3ff
      // 3e7: iload 23
      // 3e9: bipush -1
      // 3ea: ixor
      // 3eb: sipush -231
      // 3ee: if_icmple 3ff
      // 3f1: goto 3f5
      // 3f4: athrow
      // 3f5: getstatic fb.k [Z
      // 3f8: iload 0
      // 3f9: bipush 1
      // 3fa: bastore
      // 3fb: goto 3ff
      // 3fe: athrow
      // 3ff: getstatic qb.k [B
      // 402: getstatic b.c I
      // 405: dup
      // 406: bipush 1
      // 407: iadd
      // 408: putstatic b.c I
      // 40b: iload 23
      // 40d: i2b
      // 40e: bastore
      // 40f: iinc 22 1
      // 412: iload 24
      // 414: ifeq 3bf
      // 417: iinc 21 1
      // 41a: iload 24
      // 41c: ifeq 3af
      // 41f: bipush 1
      // 420: ireturn
      // 421: astore 8
      // 423: aload 8
      // 425: new java/lang/StringBuilder
      // 428: dup
      // 429: invokespecial java/lang/StringBuilder.<init> ()V
      // 42c: getstatic s.z [Ljava/lang/String;
      // 42f: bipush 0
      // 430: aaload
      // 431: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 434: iload 0
      // 435: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 438: bipush 44
      // 43a: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 43d: aload 1
      // 43e: ifnull 44a
      // 441: getstatic s.z [Ljava/lang/String;
      // 444: bipush 1
      // 445: aaload
      // 446: goto 44f
      // 449: athrow
      // 44a: getstatic s.z [Ljava/lang/String;
      // 44d: bipush 2
      // 44e: aaload
      // 44f: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 452: bipush 44
      // 454: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 457: iload 2
      // 458: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 45b: bipush 44
      // 45d: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 460: iload 3
      // 461: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 464: bipush 44
      // 466: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 469: aload 4
      // 46b: ifnull 477
      // 46e: getstatic s.z [Ljava/lang/String;
      // 471: bipush 1
      // 472: aaload
      // 473: goto 47c
      // 476: athrow
      // 477: getstatic s.z [Ljava/lang/String;
      // 47a: bipush 2
      // 47b: aaload
      // 47c: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 47f: bipush 44
      // 481: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 484: iload 5
      // 486: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 489: bipush 44
      // 48b: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 48e: aload 6
      // 490: ifnull 49c
      // 493: getstatic s.z [Ljava/lang/String;
      // 496: bipush 1
      // 497: aaload
      // 498: goto 4a1
      // 49b: athrow
      // 49c: getstatic s.z [Ljava/lang/String;
      // 49f: bipush 2
      // 4a0: aaload
      // 4a1: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 4a4: bipush 44
      // 4a6: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 4a9: iload 7
      // 4ab: invokevirtual java/lang/StringBuilder.append (Z)Ljava/lang/StringBuilder;
      // 4ae: bipush 41
      // 4b0: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 4b3: invokevirtual java/lang/StringBuilder.toString ()Ljava/lang/String;
      // 4b6: invokestatic i.a (Ljava/lang/Throwable;Ljava/lang/String;)Lla;
      // 4b9: athrow
   }

   private static char[] z(String var0) {
      char[] var10000 = var0.toCharArray();
      if (var10000.length < 2) {
         var10000[0] = (char)(var10000[0] ^ 'n');
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
               var10005 = 107;
               break;
            case 1:
               var10005 = 85;
               break;
            case 2:
               var10005 = 98;
               break;
            case 3:
               var10005 = 3;
               break;
            default:
               var10005 = 110;
         }

         var10001[var1] = (char)(var10004 ^ var10005);
      }

      return new String(var10001).intern();
   }
}
