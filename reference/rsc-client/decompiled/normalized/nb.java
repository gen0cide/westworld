import java.io.IOException;
import java.io.InputStream;
import java.net.URL;

final class nb {
   private byte[] n;
   static int m;
   private long h = -1L;
   private long l;
   static int g = 0;
   static int p;
   private byte[] a;
   private long x;
   static int e;
   private d b;
   static int[] d;
   private int t;
   static int k;
   private int i = 0;
   static int y;
   static int r;
   static String[] u = new String[]{z(z("wFj_\nWW\u007f\u001aDVRx_X\u0003P|\u001aCWZwI\nWP:IOOS:[DG\u001fjHOPL:_DWZh"))};
   static int z;
   static URL s = null;
   static int w;
   static char[] f = new char[]{
      '€',
      '\u0000',
      '‚',
      'ƒ',
      '„',
      '…',
      '†',
      '‡',
      'ˆ',
      '‰',
      'Š',
      '‹',
      'Œ',
      '\u0000',
      'Ž',
      '\u0000',
      '\u0000',
      '‘',
      '’',
      '“',
      '”',
      '•',
      '–',
      '—',
      '˜',
      '™',
      'š',
      '›',
      'œ',
      '\u0000',
      'ž',
      'Ÿ'
   };
   private long o;
   private long A;
   private long j = -1L;
   static int c;
   static int q = 0;
   static int v;
   private static final String[] B = new String[]{
      z(z("M]4s\u0002")),
      z(z("MJvV")),
      z(z("X\u00114\u0014W")),
      z(z("M]4\u0006CMVn\u0004\u0002")),
      z(z("M]4~\u0002")),
      z(z("M]4|\u0002")),
      z(z("M]4r\u0002")),
      z(z("M]4x\u0002")),
      z(z("M]4}\u0002")),
      z(z("M]4{\u0002")),
      z(z("M]4\u007f\u0002")),
      z(z("M]4y\u0002"))
   };

   private final void a(int param1) throws IOException {
      // $VF: Couldn't be decompiled
      // Please report this to the Vineflower issue tracker, at https://github.com/Vineflower/vineflower/issues with a copy of the class file (if you have the rights to distribute it!)
      // java.lang.RuntimeException: parsing failure!
      //   at org.jetbrains.java.decompiler.modules.decompiler.decompose.DomHelper.parseGraph(DomHelper.java:211)
      //   at org.jetbrains.java.decompiler.main.rels.MethodProcessor.codeToJava(MethodProcessor.java:166)
      //
      // Bytecode:
      // 000: getstatic client.vh Z
      // 003: istore 7
      // 005: aload 0
      // 006: getfield nb.h J
      // 009: ldc2_w -1
      // 00c: lcmp
      // 00d: ifeq 1ca
      // 010: aload 0
      // 011: getfield nb.x J
      // 014: ldc2_w -1
      // 017: lxor
      // 018: aload 0
      // 019: getfield nb.h J
      // 01c: ldc2_w -1
      // 01f: lxor
      // 020: lcmp
      // 021: ifeq 040
      // 024: goto 028
      // 027: athrow
      // 028: aload 0
      // 029: getfield nb.b Ld;
      // 02c: bipush 0
      // 02d: aload 0
      // 02e: getfield nb.h J
      // 031: invokevirtual d.a (IJ)V
      // 034: aload 0
      // 035: aload 0
      // 036: getfield nb.h J
      // 039: putfield nb.x J
      // 03c: goto 040
      // 03f: athrow
      // 040: aload 0
      // 041: getfield nb.b Ld;
      // 044: aload 0
      // 045: getfield nb.a [B
      // 048: aload 0
      // 049: getfield nb.i I
      // 04c: bipush 1
      // 04d: bipush 0
      // 04e: invokevirtual d.b ([BIII)V
      // 051: aload 0
      // 052: dup
      // 053: getfield nb.x J
      // 056: aload 0
      // 057: getfield nb.i I
      // 05a: i2l
      // 05b: ladd
      // 05c: putfield nb.x J
      // 05f: aload 0
      // 060: getfield nb.A J
      // 063: ldc2_w -1
      // 066: lxor
      // 067: aload 0
      // 068: getfield nb.x J
      // 06b: ldc2_w -1
      // 06e: lxor
      // 06f: lcmp
      // 070: ifgt 077
      // 073: goto 07f
      // 076: athrow
      // 077: aload 0
      // 078: aload 0
      // 079: getfield nb.x J
      // 07c: putfield nb.A J
      // 07f: ldc2_w -1
      // 082: lstore 2
      // 083: ldc2_w -1
      // 086: lstore 4
      // 088: aload 0
      // 089: getfield nb.h J
      // 08c: aload 0
      // 08d: getfield nb.j J
      // 090: lcmp
      // 091: iflt 0b2
      // 094: aload 0
      // 095: getfield nb.h J
      // 098: ldc2_w -1
      // 09b: lxor
      // 09c: aload 0
      // 09d: getfield nb.j J
      // 0a0: aload 0
      // 0a1: getfield nb.t I
      // 0a4: i2l
      // 0a5: ladd
      // 0a6: ldc2_w -1
      // 0a9: lxor
      // 0aa: lcmp
      // 0ab: ifgt 0ea
      // 0ae: goto 0b2
      // 0b1: athrow
      // 0b2: aload 0
      // 0b3: getfield nb.j J
      // 0b6: ldc2_w -1
      // 0b9: lxor
      // 0ba: aload 0
      // 0bb: getfield nb.h J
      // 0be: ldc2_w -1
      // 0c1: lxor
      // 0c2: lcmp
      // 0c3: ifgt 0ef
      // 0c6: goto 0ca
      // 0c9: athrow
      // 0ca: aload 0
      // 0cb: getfield nb.j J
      // 0ce: aload 0
      // 0cf: getfield nb.i I
      // 0d2: i2l
      // 0d3: aload 0
      // 0d4: getfield nb.h J
      // 0d7: ladd
      // 0d8: lcmp
      // 0d9: ifge 0ef
      // 0dc: goto 0e0
      // 0df: athrow
      // 0e0: aload 0
      // 0e1: getfield nb.j J
      // 0e4: lstore 2
      // 0e5: iload 7
      // 0e7: ifeq 0ef
      // 0ea: aload 0
      // 0eb: getfield nb.h J
      // 0ee: lstore 2
      // 0ef: aload 0
      // 0f0: getfield nb.j J
      // 0f3: ldc2_w -1
      // 0f6: lxor
      // 0f7: aload 0
      // 0f8: getfield nb.h J
      // 0fb: aload 0
      // 0fc: getfield nb.i I
      // 0ff: i2l
      // 100: ladd
      // 101: ldc2_w -1
      // 104: lxor
      // 105: lcmp
      // 106: ifle 13e
      // 109: aload 0
      // 10a: getfield nb.i I
      // 10d: i2l
      // 10e: aload 0
      // 10f: getfield nb.h J
      // 112: ladd
      // 113: ldc2_w -1
      // 116: lxor
      // 117: aload 0
      // 118: getfield nb.j J
      // 11b: aload 0
      // 11c: getfield nb.t I
      // 11f: i2l
      // 120: ladd
      // 121: ldc2_w -1
      // 124: lxor
      // 125: lcmp
      // 126: iflt 13e
      // 129: goto 12d
      // 12c: athrow
      // 12d: aload 0
      // 12e: getfield nb.h J
      // 131: aload 0
      // 132: getfield nb.i I
      // 135: i2l
      // 136: ladd
      // 137: lstore 4
      // 139: iload 7
      // 13b: ifeq 17e
      // 13e: aload 0
      // 13f: getfield nb.t I
      // 142: i2l
      // 143: aload 0
      // 144: getfield nb.j J
      // 147: ladd
      // 148: aload 0
      // 149: getfield nb.h J
      // 14c: lcmp
      // 14d: ifle 17e
      // 150: goto 154
      // 153: athrow
      // 154: aload 0
      // 155: getfield nb.j J
      // 158: aload 0
      // 159: getfield nb.t I
      // 15c: i2l
      // 15d: lneg
      // 15e: lsub
      // 15f: aload 0
      // 160: getfield nb.i I
      // 163: i2l
      // 164: aload 0
      // 165: getfield nb.h J
      // 168: ladd
      // 169: lcmp
      // 16a: ifgt 17e
      // 16d: goto 171
      // 170: athrow
      // 171: aload 0
      // 172: getfield nb.j J
      // 175: aload 0
      // 176: getfield nb.t I
      // 179: i2l
      // 17a: lneg
      // 17b: lsub
      // 17c: lstore 4
      // 17e: ldc2_w -1
      // 181: lload 2
      // 182: lcmp
      // 183: ifge 1be
      // 186: lload 2
      // 187: ldc2_w -1
      // 18a: lxor
      // 18b: lload 4
      // 18d: ldc2_w -1
      // 190: lxor
      // 191: lcmp
      // 192: ifle 1be
      // 195: goto 199
      // 198: athrow
      // 199: lload 4
      // 19b: lload 2
      // 19c: lneg
      // 19d: ladd
      // 19e: l2i
      // 19f: istore 6
      // 1a1: aload 0
      // 1a2: getfield nb.a [B
      // 1a5: lload 2
      // 1a6: aload 0
      // 1a7: getfield nb.h J
      // 1aa: lneg
      // 1ab: ladd
      // 1ac: l2i
      // 1ad: aload 0
      // 1ae: getfield nb.n [B
      // 1b1: aload 0
      // 1b2: getfield nb.j J
      // 1b5: lneg
      // 1b6: lload 2
      // 1b7: ladd
      // 1b8: l2i
      // 1b9: iload 6
      // 1bb: invokestatic ab.a ([BI[BII)V
      // 1be: aload 0
      // 1bf: bipush 0
      // 1c0: putfield nb.i I
      // 1c3: aload 0
      // 1c4: ldc2_w -1
      // 1c7: putfield nb.h J
      // 1ca: getstatic nb.v I
      // 1cd: bipush 1
      // 1ce: iadd
      // 1cf: putstatic nb.v I
      // 1d2: iload 1
      // 1d3: sipush -14779
      // 1d6: if_icmpeq 1e5
      // 1d9: aload 0
      // 1da: aconst_null
      // 1db: checkcast [B
      // 1de: putfield nb.n [B
      // 1e1: goto 1e5
      // 1e4: athrow
      // 1e5: goto 20a
      // 1e8: astore 2
      // 1e9: aload 2
      // 1ea: new java/lang/StringBuilder
      // 1ed: dup
      // 1ee: invokespecial java/lang/StringBuilder.<init> ()V
      // 1f1: getstatic nb.B [Ljava/lang/String;
      // 1f4: bipush 9
      // 1f6: aaload
      // 1f7: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 1fa: iload 1
      // 1fb: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 1fe: bipush 41
      // 200: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 203: invokevirtual java/lang/StringBuilder.toString ()Ljava/lang/String;
      // 206: invokestatic i.a (Ljava/lang/Throwable;Ljava/lang/String;)Lla;
      // 209: athrow
      // 20a: return
   }

   final void a(boolean param1, int param2, int param3, byte[] param4) throws IOException {
      // $VF: Couldn't be decompiled
      // Please report this to the Vineflower issue tracker, at https://github.com/Vineflower/vineflower/issues with a copy of the class file (if you have the rights to distribute it!)
      // java.lang.RuntimeException: parsing failure!
      //   at org.jetbrains.java.decompiler.modules.decompiler.decompose.DomHelper.parseGraph(DomHelper.java:211)
      //   at org.jetbrains.java.decompiler.main.rels.MethodProcessor.codeToJava(MethodProcessor.java:166)
      //
      // Bytecode:
      // 000: getstatic client.vh Z
      // 003: istore 14
      // 005: getstatic nb.z I
      // 008: bipush 1
      // 009: iadd
      // 00a: putstatic nb.z I
      // 00d: iload 2
      // 00e: iload 3
      // 00f: iadd
      // 010: bipush -1
      // 011: ixor
      // 012: aload 4
      // 014: arraylength
      // 015: bipush -1
      // 016: ixor
      // 017: if_icmpge 02a
      // 01a: new java/lang/ArrayIndexOutOfBoundsException
      // 01d: dup
      // 01e: iload 3
      // 01f: iload 2
      // 020: ineg
      // 021: isub
      // 022: aload 4
      // 024: arraylength
      // 025: isub
      // 026: invokespecial java/lang/ArrayIndexOutOfBoundsException.<init> (I)V
      // 029: athrow
      // 02a: aload 0
      // 02b: getfield nb.h J
      // 02e: ldc2_w -1
      // 031: lcmp
      // 032: ifeq 080
      // 035: aload 0
      // 036: getfield nb.h J
      // 039: aload 0
      // 03a: getfield nb.l J
      // 03d: lcmp
      // 03e: ifgt 080
      // 041: goto 045
      // 044: athrow
      // 045: aload 0
      // 046: getfield nb.h J
      // 049: aload 0
      // 04a: getfield nb.i I
      // 04d: i2l
      // 04e: lneg
      // 04f: lsub
      // 050: iload 2
      // 051: i2l
      // 052: aload 0
      // 053: getfield nb.l J
      // 056: ladd
      // 057: lcmp
      // 058: iflt 080
      // 05b: goto 05f
      // 05e: athrow
      // 05f: aload 0
      // 060: getfield nb.a [B
      // 063: aload 0
      // 064: getfield nb.l J
      // 067: aload 0
      // 068: getfield nb.h J
      // 06b: lsub
      // 06c: l2i
      // 06d: aload 4
      // 06f: iload 3
      // 070: iload 2
      // 071: invokestatic ab.a ([BI[BII)V
      // 074: aload 0
      // 075: dup
      // 076: getfield nb.l J
      // 079: iload 2
      // 07a: i2l
      // 07b: ladd
      // 07c: putfield nb.l J
      // 07f: return
      // 080: aload 0
      // 081: getfield nb.l J
      // 084: lstore 5
      // 086: iload 3
      // 087: istore 7
      // 089: iload 2
      // 08a: istore 8
      // 08c: aload 0
      // 08d: getfield nb.l J
      // 090: ldc2_w -1
      // 093: lxor
      // 094: aload 0
      // 095: getfield nb.j J
      // 098: ldc2_w -1
      // 09b: lxor
      // 09c: lcmp
      // 09d: ifgt 10a
      // 0a0: aload 0
      // 0a1: getfield nb.j J
      // 0a4: aload 0
      // 0a5: getfield nb.t I
      // 0a8: i2l
      // 0a9: ladd
      // 0aa: ldc2_w -1
      // 0ad: lxor
      // 0ae: aload 0
      // 0af: getfield nb.l J
      // 0b2: ldc2_w -1
      // 0b5: lxor
      // 0b6: lcmp
      // 0b7: ifge 10a
      // 0ba: goto 0be
      // 0bd: athrow
      // 0be: aload 0
      // 0bf: getfield nb.j J
      // 0c2: aload 0
      // 0c3: getfield nb.l J
      // 0c6: lneg
      // 0c7: aload 0
      // 0c8: getfield nb.t I
      // 0cb: i2l
      // 0cc: ladd
      // 0cd: ladd
      // 0ce: l2i
      // 0cf: istore 9
      // 0d1: iload 2
      // 0d2: bipush -1
      // 0d3: ixor
      // 0d4: iload 9
      // 0d6: bipush -1
      // 0d7: ixor
      // 0d8: if_icmple 0de
      // 0db: iload 2
      // 0dc: istore 9
      // 0de: aload 0
      // 0df: getfield nb.n [B
      // 0e2: aload 0
      // 0e3: getfield nb.l J
      // 0e6: aload 0
      // 0e7: getfield nb.j J
      // 0ea: lsub
      // 0eb: l2i
      // 0ec: aload 4
      // 0ee: iload 3
      // 0ef: iload 9
      // 0f1: invokestatic ab.a ([BI[BII)V
      // 0f4: iload 2
      // 0f5: iload 9
      // 0f7: isub
      // 0f8: istore 2
      // 0f9: iload 3
      // 0fa: iload 9
      // 0fc: iadd
      // 0fd: istore 3
      // 0fe: aload 0
      // 0ff: dup
      // 100: getfield nb.l J
      // 103: iload 9
      // 105: i2l
      // 106: ladd
      // 107: putfield nb.l J
      // 10a: iload 1
      // 10b: bipush 1
      // 10c: if_icmpeq 110
      // 10f: return
      // 110: iload 2
      // 111: aload 0
      // 112: getfield nb.n [B
      // 115: arraylength
      // 116: if_icmpgt 166
      // 119: bipush 0
      // 11a: iload 2
      // 11b: if_icmpge 1d4
      // 11e: goto 122
      // 121: athrow
      // 122: aload 0
      // 123: bipush 34
      // 125: invokespecial nb.b (B)V
      // 128: iload 2
      // 129: istore 9
      // 12b: aload 0
      // 12c: getfield nb.t I
      // 12f: bipush -1
      // 130: ixor
      // 131: iload 9
      // 133: bipush -1
      // 134: ixor
      // 135: if_icmple 13e
      // 138: aload 0
      // 139: getfield nb.t I
      // 13c: istore 9
      // 13e: aload 0
      // 13f: getfield nb.n [B
      // 142: bipush 0
      // 143: aload 4
      // 145: iload 3
      // 146: iload 9
      // 148: invokestatic ab.a ([BI[BII)V
      // 14b: iload 3
      // 14c: iload 9
      // 14e: iadd
      // 14f: istore 3
      // 150: iload 2
      // 151: iload 9
      // 153: isub
      // 154: istore 2
      // 155: aload 0
      // 156: dup
      // 157: getfield nb.l J
      // 15a: iload 9
      // 15c: i2l
      // 15d: ladd
      // 15e: putfield nb.l J
      // 161: iload 14
      // 163: ifeq 1d4
      // 166: aload 0
      // 167: getfield nb.b Ld;
      // 16a: bipush 0
      // 16b: aload 0
      // 16c: getfield nb.l J
      // 16f: invokevirtual d.a (IJ)V
      // 172: aload 0
      // 173: aload 0
      // 174: getfield nb.l J
      // 177: putfield nb.x J
      // 17a: goto 17e
      // 17d: athrow
      // 17e: bipush -1
      // 17f: iload 2
      // 180: bipush -1
      // 181: ixor
      // 182: if_icmple 1d4
      // 185: aload 0
      // 186: getfield nb.b Ld;
      // 189: aload 4
      // 18b: iload 2
      // 18c: iload 3
      // 18d: bipush -1
      // 18e: invokevirtual d.a ([BIII)I
      // 191: istore 9
      // 193: bipush 0
      // 194: iload 9
      // 196: bipush -1
      // 197: ixor
      // 198: iload 14
      // 19a: ifne 3b0
      // 19d: if_icmpne 1ad
      // 1a0: goto 1a4
      // 1a3: athrow
      // 1a4: iload 14
      // 1a6: ifeq 1d4
      // 1a9: goto 1ad
      // 1ac: athrow
      // 1ad: iload 3
      // 1ae: iload 9
      // 1b0: iadd
      // 1b1: istore 3
      // 1b2: iload 2
      // 1b3: iload 9
      // 1b5: isub
      // 1b6: istore 2
      // 1b7: aload 0
      // 1b8: dup
      // 1b9: getfield nb.l J
      // 1bc: iload 9
      // 1be: i2l
      // 1bf: ladd
      // 1c0: putfield nb.l J
      // 1c3: aload 0
      // 1c4: dup
      // 1c5: getfield nb.x J
      // 1c8: iload 9
      // 1ca: i2l
      // 1cb: ladd
      // 1cc: putfield nb.x J
      // 1cf: iload 14
      // 1d1: ifeq 17e
      // 1d4: aload 0
      // 1d5: getfield nb.h J
      // 1d8: ldc2_w -1
      // 1db: lcmp
      // 1dc: ifne 1e3
      // 1df: goto 39f
      // 1e2: athrow
      // 1e3: aload 0
      // 1e4: getfield nb.l J
      // 1e7: ldc2_w -1
      // 1ea: lxor
      // 1eb: aload 0
      // 1ec: getfield nb.h J
      // 1ef: ldc2_w -1
      // 1f2: lxor
      // 1f3: lcmp
      // 1f4: ifle 252
      // 1f7: bipush 0
      // 1f8: iload 2
      // 1f9: if_icmplt 204
      // 1fc: goto 200
      // 1ff: athrow
      // 200: goto 252
      // 203: athrow
      // 204: aload 0
      // 205: getfield nb.h J
      // 208: aload 0
      // 209: getfield nb.l J
      // 20c: lsub
      // 20d: l2i
      // 20e: iload 3
      // 20f: iadd
      // 210: istore 9
      // 212: iload 9
      // 214: iload 3
      // 215: iload 2
      // 216: ineg
      // 217: isub
      // 218: if_icmpgt 21f
      // 21b: goto 225
      // 21e: athrow
      // 21f: iload 3
      // 220: iload 2
      // 221: ineg
      // 222: isub
      // 223: istore 9
      // 225: iload 9
      // 227: bipush -1
      // 228: ixor
      // 229: iload 3
      // 22a: bipush -1
      // 22b: ixor
      // 22c: if_icmpge 252
      // 22f: aload 4
      // 231: iload 3
      // 232: iinc 3 1
      // 235: bipush 0
      // 236: bastore
      // 237: iinc 2 -1
      // 23a: aload 0
      // 23b: dup
      // 23c: getfield nb.l J
      // 23f: lconst_1
      // 240: ladd
      // 241: putfield nb.l J
      // 244: iload 14
      // 246: ifne 39f
      // 249: iload 14
      // 24b: ifeq 225
      // 24e: goto 252
      // 251: athrow
      // 252: ldc2_w -1
      // 255: lstore 9
      // 257: lload 5
      // 259: ldc2_w -1
      // 25c: lxor
      // 25d: aload 0
      // 25e: getfield nb.h J
      // 261: ldc2_w -1
      // 264: lxor
      // 265: lcmp
      // 266: iflt 28e
      // 269: lload 5
      // 26b: iload 8
      // 26d: i2l
      // 26e: ladd
      // 26f: ldc2_w -1
      // 272: lxor
      // 273: aload 0
      // 274: getfield nb.h J
      // 277: ldc2_w -1
      // 27a: lxor
      // 27b: lcmp
      // 27c: ifge 28e
      // 27f: goto 283
      // 282: athrow
      // 283: aload 0
      // 284: getfield nb.h J
      // 287: lstore 9
      // 289: iload 14
      // 28b: ifeq 2bc
      // 28e: lload 5
      // 290: ldc2_w -1
      // 293: lxor
      // 294: aload 0
      // 295: getfield nb.h J
      // 298: ldc2_w -1
      // 29b: lxor
      // 29c: lcmp
      // 29d: ifgt 2bc
      // 2a0: goto 2a4
      // 2a3: athrow
      // 2a4: lload 5
      // 2a6: aload 0
      // 2a7: getfield nb.h J
      // 2aa: aload 0
      // 2ab: getfield nb.i I
      // 2ae: i2l
      // 2af: ladd
      // 2b0: lcmp
      // 2b1: ifge 2bc
      // 2b4: goto 2b8
      // 2b7: athrow
      // 2b8: lload 5
      // 2ba: lstore 9
      // 2bc: ldc2_w -1
      // 2bf: lstore 11
      // 2c1: lload 5
      // 2c3: aload 0
      // 2c4: getfield nb.h J
      // 2c7: aload 0
      // 2c8: getfield nb.i I
      // 2cb: i2l
      // 2cc: ladd
      // 2cd: lcmp
      // 2ce: ifge 2ea
      // 2d1: aload 0
      // 2d2: getfield nb.h J
      // 2d5: aload 0
      // 2d6: getfield nb.i I
      // 2d9: i2l
      // 2da: lneg
      // 2db: lsub
      // 2dc: lload 5
      // 2de: iload 8
      // 2e0: i2l
      // 2e1: ladd
      // 2e2: lcmp
      // 2e3: ifle 32a
      // 2e6: goto 2ea
      // 2e9: athrow
      // 2ea: lload 5
      // 2ec: iload 8
      // 2ee: i2l
      // 2ef: lneg
      // 2f0: lsub
      // 2f1: aload 0
      // 2f2: getfield nb.h J
      // 2f5: lcmp
      // 2f6: ifle 336
      // 2f9: goto 2fd
      // 2fc: athrow
      // 2fd: aload 0
      // 2fe: getfield nb.i I
      // 301: i2l
      // 302: aload 0
      // 303: getfield nb.h J
      // 306: ladd
      // 307: ldc2_w -1
      // 30a: lxor
      // 30b: iload 8
      // 30d: i2l
      // 30e: lload 5
      // 310: ladd
      // 311: ldc2_w -1
      // 314: lxor
      // 315: lcmp
      // 316: ifgt 336
      // 319: goto 31d
      // 31c: athrow
      // 31d: lload 5
      // 31f: iload 8
      // 321: i2l
      // 322: ladd
      // 323: lstore 11
      // 325: iload 14
      // 327: ifeq 336
      // 32a: aload 0
      // 32b: getfield nb.h J
      // 32e: aload 0
      // 32f: getfield nb.i I
      // 332: i2l
      // 333: ladd
      // 334: lstore 11
      // 336: lload 9
      // 338: ldc2_w -1
      // 33b: lcmp
      // 33c: ifle 39f
      // 33f: lload 9
      // 341: ldc2_w -1
      // 344: lxor
      // 345: lload 11
      // 347: ldc2_w -1
      // 34a: lxor
      // 34b: lcmp
      // 34c: ifle 39f
      // 34f: goto 353
      // 352: athrow
      // 353: lload 11
      // 355: lload 9
      // 357: lneg
      // 358: ladd
      // 359: l2i
      // 35a: istore 13
      // 35c: aload 0
      // 35d: getfield nb.a [B
      // 360: aload 0
      // 361: getfield nb.h J
      // 364: lneg
      // 365: lload 9
      // 367: ladd
      // 368: l2i
      // 369: aload 4
      // 36b: iload 7
      // 36d: lload 9
      // 36f: lload 5
      // 371: lneg
      // 372: ladd
      // 373: l2i
      // 374: ineg
      // 375: isub
      // 376: iload 13
      // 378: invokestatic ab.a ([BI[BII)V
      // 37b: aload 0
      // 37c: getfield nb.l J
      // 37f: ldc2_w -1
      // 382: lxor
      // 383: lload 11
      // 385: ldc2_w -1
      // 388: lxor
      // 389: lcmp
      // 38a: ifle 39f
      // 38d: iload 2
      // 38e: i2l
      // 38f: lload 11
      // 391: aload 0
      // 392: getfield nb.l J
      // 395: lsub
      // 396: lsub
      // 397: l2i
      // 398: istore 2
      // 399: aload 0
      // 39a: lload 11
      // 39c: putfield nb.l J
      // 39f: goto 3ae
      // 3a2: astore 5
      // 3a4: aload 0
      // 3a5: ldc2_w -1
      // 3a8: putfield nb.x J
      // 3ab: aload 5
      // 3ad: athrow
      // 3ae: bipush 0
      // 3af: iload 2
      // 3b0: if_icmplt 3b6
      // 3b3: goto 3be
      // 3b6: new java/io/EOFException
      // 3b9: dup
      // 3ba: invokespecial java/io/EOFException.<init> ()V
      // 3bd: athrow
      // 3be: goto 412
      // 3c1: astore 5
      // 3c3: aload 5
      // 3c5: new java/lang/StringBuilder
      // 3c8: dup
      // 3c9: invokespecial java/lang/StringBuilder.<init> ()V
      // 3cc: getstatic nb.B [Ljava/lang/String;
      // 3cf: bipush 6
      // 3d1: aaload
      // 3d2: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 3d5: iload 1
      // 3d6: invokevirtual java/lang/StringBuilder.append (Z)Ljava/lang/StringBuilder;
      // 3d9: bipush 44
      // 3db: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 3de: iload 2
      // 3df: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 3e2: bipush 44
      // 3e4: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 3e7: iload 3
      // 3e8: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 3eb: bipush 44
      // 3ed: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 3f0: aload 4
      // 3f2: ifnull 3fe
      // 3f5: getstatic nb.B [Ljava/lang/String;
      // 3f8: bipush 2
      // 3f9: aaload
      // 3fa: goto 403
      // 3fd: athrow
      // 3fe: getstatic nb.B [Ljava/lang/String;
      // 401: bipush 1
      // 402: aaload
      // 403: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 406: bipush 41
      // 408: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 40b: invokevirtual java/lang/StringBuilder.toString ()Ljava/lang/String;
      // 40e: invokestatic i.a (Ljava/lang/Throwable;Ljava/lang/String;)Lla;
      // 411: athrow
      // 412: return
   }

   final void a(byte var1, byte[] var2) throws IOException {
      try {
         int var6 = -12 % ((var1 - -22) / 54);
         this.a(true, var2.length, 0, var2);
         p++;
      } catch (RuntimeException var5) {
         RuntimeException var3 = var5;

         RuntimeException var10000;
         StringBuilder var10001;
         try {
            var10000 = var3;
            var10001 = new StringBuilder().append(B[4]).append((int)var1).append(',');
            if (var2 != null) {
               throw i.a(var3, var10001.append(B[2]).append(')').toString());
            }
         } catch (RuntimeException var4) {
            throw var4;
         }

         throw i.a(var10000, var10001.append(B[1]).append(')').toString());
      }
   }

   final long a(byte var1) {
      try {
         try {
            if (var1 != -111) {
               d = (int[])null;
            }
         } catch (RuntimeException var3) {
            throw var3;
         }

         r++;
         return this.o;
      } catch (RuntimeException var4) {
         throw i.a(var4, B[0] + var1 + ')');
      }
   }

   static final int a(int var0, byte var1) {
      try {
         try {
            e++;
            if (var0 != 255) {
               a(-35, (byte)126);
            }
         } catch (RuntimeException var3) {
            throw var3;
         }

         return var1 & 0xFF;
      } catch (RuntimeException var4) {
         throw i.a(var4, B[8] + var0 + 44 + var1 + 41);
      }
   }

   static final InputStream a(boolean param0, String param1) throws IOException {
      // $VF: Couldn't be decompiled
      // Please report this to the Vineflower issue tracker, at https://github.com/Vineflower/vineflower/issues with a copy of the class file (if you have the rights to distribute it!)
      // java.lang.RuntimeException: parsing failure!
      //   at org.jetbrains.java.decompiler.modules.decompiler.decompose.DomHelper.parseGraph(DomHelper.java:211)
      //   at org.jetbrains.java.decompiler.main.rels.MethodProcessor.codeToJava(MethodProcessor.java:166)
      //
      // Bytecode:
      // 00: iload 0
      // 01: bipush 1
      // 02: if_icmpeq 0b
      // 05: aconst_null
      // 06: checkcast java/io/InputStream
      // 09: areturn
      // 0a: athrow
      // 0b: getstatic nb.c I
      // 0e: bipush 1
      // 0f: iadd
      // 10: putstatic nb.c I
      // 13: aconst_null
      // 14: getstatic nb.s Ljava/net/URL;
      // 17: if_acmpeq 31
      // 1a: new java/net/URL
      // 1d: dup
      // 1e: getstatic nb.s Ljava/net/URL;
      // 21: aload 1
      // 22: invokespecial java/net/URL.<init> (Ljava/net/URL;Ljava/lang/String;)V
      // 25: astore 3
      // 26: aload 3
      // 27: invokevirtual java/net/URL.openStream ()Ljava/io/InputStream;
      // 2a: astore 2
      // 2b: getstatic client.vh Z
      // 2e: ifeq 41
      // 31: new java/io/BufferedInputStream
      // 34: dup
      // 35: new java/io/FileInputStream
      // 38: dup
      // 39: aload 1
      // 3a: invokespecial java/io/FileInputStream.<init> (Ljava/lang/String;)V
      // 3d: invokespecial java/io/BufferedInputStream.<init> (Ljava/io/InputStream;)V
      // 40: astore 2
      // 41: aload 2
      // 42: areturn
      // 43: astore 2
      // 44: aload 2
      // 45: new java/lang/StringBuilder
      // 48: dup
      // 49: invokespecial java/lang/StringBuilder.<init> ()V
      // 4c: getstatic nb.B [Ljava/lang/String;
      // 4f: bipush 5
      // 50: aaload
      // 51: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 54: iload 0
      // 55: invokevirtual java/lang/StringBuilder.append (Z)Ljava/lang/StringBuilder;
      // 58: bipush 44
      // 5a: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 5d: aload 1
      // 5e: ifnull 6a
      // 61: getstatic nb.B [Ljava/lang/String;
      // 64: bipush 2
      // 65: aaload
      // 66: goto 6f
      // 69: athrow
      // 6a: getstatic nb.B [Ljava/lang/String;
      // 6d: bipush 1
      // 6e: aaload
      // 6f: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 72: bipush 41
      // 74: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 77: invokevirtual java/lang/StringBuilder.toString ()Ljava/lang/String;
      // 7a: invokestatic i.a (Ljava/lang/Throwable;Ljava/lang/String;)Lla;
      // 7d: athrow
   }

   final void a(int param1, int param2, int param3, byte[] param4) throws IOException {
      // $VF: Couldn't be decompiled
      // Please report this to the Vineflower issue tracker, at https://github.com/Vineflower/vineflower/issues with a copy of the class file (if you have the rights to distribute it!)
      // java.lang.RuntimeException: parsing failure!
      //   at org.jetbrains.java.decompiler.modules.decompiler.decompose.DomHelper.parseGraph(DomHelper.java:211)
      //   at org.jetbrains.java.decompiler.main.rels.MethodProcessor.codeToJava(MethodProcessor.java:166)
      //
      // Bytecode:
      // 000: getstatic client.vh Z
      // 003: istore 10
      // 005: getstatic nb.w I
      // 008: bipush 1
      // 009: iadd
      // 00a: putstatic nb.w I
      // 00d: iload 2
      // 00e: bipush -80
      // 010: if_icmple 015
      // 013: return
      // 014: athrow
      // 015: aload 0
      // 016: getfield nb.l J
      // 019: iload 1
      // 01a: i2l
      // 01b: ladd
      // 01c: aload 0
      // 01d: getfield nb.o J
      // 020: lcmp
      // 021: ifgt 028
      // 024: goto 033
      // 027: athrow
      // 028: aload 0
      // 029: iload 1
      // 02a: i2l
      // 02b: aload 0
      // 02c: getfield nb.l J
      // 02f: ladd
      // 030: putfield nb.o J
      // 033: aload 0
      // 034: getfield nb.h J
      // 037: ldc2_w -1
      // 03a: lcmp
      // 03b: ifeq 077
      // 03e: aload 0
      // 03f: getfield nb.l J
      // 042: ldc2_w -1
      // 045: lxor
      // 046: aload 0
      // 047: getfield nb.h J
      // 04a: ldc2_w -1
      // 04d: lxor
      // 04e: lcmp
      // 04f: ifgt 070
      // 052: goto 056
      // 055: athrow
      // 056: aload 0
      // 057: getfield nb.i I
      // 05a: i2l
      // 05b: aload 0
      // 05c: getfield nb.h J
      // 05f: ladd
      // 060: aload 0
      // 061: getfield nb.l J
      // 064: lcmp
      // 065: iflt 070
      // 068: goto 06c
      // 06b: athrow
      // 06c: goto 077
      // 06f: athrow
      // 070: aload 0
      // 071: sipush -14779
      // 074: invokespecial nb.a (I)V
      // 077: lconst_0
      // 078: aload 0
      // 079: getfield nb.h J
      // 07c: ldc2_w -1
      // 07f: lxor
      // 080: lcmp
      // 081: ifeq 0f0
      // 084: iload 1
      // 085: i2l
      // 086: aload 0
      // 087: getfield nb.l J
      // 08a: ladd
      // 08b: aload 0
      // 08c: getfield nb.a [B
      // 08f: arraylength
      // 090: i2l
      // 091: aload 0
      // 092: getfield nb.h J
      // 095: ladd
      // 096: lcmp
      // 097: ifle 0f0
      // 09a: goto 09e
      // 09d: athrow
      // 09e: aload 0
      // 09f: getfield nb.l J
      // 0a2: lneg
      // 0a3: aload 0
      // 0a4: getfield nb.h J
      // 0a7: lneg
      // 0a8: lsub
      // 0a9: aload 0
      // 0aa: getfield nb.a [B
      // 0ad: arraylength
      // 0ae: i2l
      // 0af: ladd
      // 0b0: l2i
      // 0b1: istore 5
      // 0b3: aload 4
      // 0b5: iload 3
      // 0b6: aload 0
      // 0b7: getfield nb.a [B
      // 0ba: aload 0
      // 0bb: getfield nb.h J
      // 0be: lneg
      // 0bf: aload 0
      // 0c0: getfield nb.l J
      // 0c3: ladd
      // 0c4: l2i
      // 0c5: iload 5
      // 0c7: invokestatic ab.a ([BI[BII)V
      // 0ca: iload 1
      // 0cb: iload 5
      // 0cd: isub
      // 0ce: istore 1
      // 0cf: iload 3
      // 0d0: iload 5
      // 0d2: iadd
      // 0d3: istore 3
      // 0d4: aload 0
      // 0d5: dup
      // 0d6: getfield nb.l J
      // 0d9: iload 5
      // 0db: i2l
      // 0dc: ladd
      // 0dd: putfield nb.l J
      // 0e0: aload 0
      // 0e1: aload 0
      // 0e2: getfield nb.a [B
      // 0e5: arraylength
      // 0e6: putfield nb.i I
      // 0e9: aload 0
      // 0ea: sipush -14779
      // 0ed: invokespecial nb.a (I)V
      // 0f0: iload 1
      // 0f1: bipush -1
      // 0f2: ixor
      // 0f3: aload 0
      // 0f4: getfield nb.a [B
      // 0f7: arraylength
      // 0f8: bipush -1
      // 0f9: ixor
      // 0fa: if_icmpge 29e
      // 0fd: aload 0
      // 0fe: getfield nb.x J
      // 101: ldc2_w -1
      // 104: lxor
      // 105: aload 0
      // 106: getfield nb.l J
      // 109: ldc2_w -1
      // 10c: lxor
      // 10d: lcmp
      // 10e: ifne 119
      // 111: goto 115
      // 114: athrow
      // 115: goto 12d
      // 118: athrow
      // 119: aload 0
      // 11a: getfield nb.b Ld;
      // 11d: bipush 0
      // 11e: aload 0
      // 11f: getfield nb.l J
      // 122: invokevirtual d.a (IJ)V
      // 125: aload 0
      // 126: aload 0
      // 127: getfield nb.l J
      // 12a: putfield nb.x J
      // 12d: aload 0
      // 12e: getfield nb.b Ld;
      // 131: aload 4
      // 133: iload 1
      // 134: bipush 1
      // 135: iload 3
      // 136: invokevirtual d.b ([BIII)V
      // 139: aload 0
      // 13a: dup
      // 13b: getfield nb.x J
      // 13e: iload 1
      // 13f: i2l
      // 140: ladd
      // 141: putfield nb.x J
      // 144: aload 0
      // 145: getfield nb.x J
      // 148: aload 0
      // 149: getfield nb.A J
      // 14c: lcmp
      // 14d: ifgt 154
      // 150: goto 15c
      // 153: athrow
      // 154: aload 0
      // 155: aload 0
      // 156: getfield nb.x J
      // 159: putfield nb.A J
      // 15c: ldc2_w -1
      // 15f: lstore 5
      // 161: ldc2_w -1
      // 164: lstore 7
      // 166: aload 0
      // 167: getfield nb.j J
      // 16a: aload 0
      // 16b: getfield nb.l J
      // 16e: lcmp
      // 16f: ifgt 193
      // 172: aload 0
      // 173: getfield nb.l J
      // 176: aload 0
      // 177: getfield nb.j J
      // 17a: aload 0
      // 17b: getfield nb.t I
      // 17e: i2l
      // 17f: ladd
      // 180: lcmp
      // 181: ifge 193
      // 184: goto 188
      // 187: athrow
      // 188: aload 0
      // 189: getfield nb.l J
      // 18c: lstore 5
      // 18e: iload 10
      // 190: ifeq 1c1
      // 193: aload 0
      // 194: getfield nb.l J
      // 197: aload 0
      // 198: getfield nb.j J
      // 19b: lcmp
      // 19c: ifgt 1c1
      // 19f: goto 1a3
      // 1a2: athrow
      // 1a3: aload 0
      // 1a4: getfield nb.l J
      // 1a7: iload 1
      // 1a8: i2l
      // 1a9: lneg
      // 1aa: lsub
      // 1ab: aload 0
      // 1ac: getfield nb.j J
      // 1af: lcmp
      // 1b0: ifgt 1bb
      // 1b3: goto 1b7
      // 1b6: athrow
      // 1b7: goto 1c1
      // 1ba: athrow
      // 1bb: aload 0
      // 1bc: getfield nb.j J
      // 1bf: lstore 5
      // 1c1: aload 0
      // 1c2: getfield nb.j J
      // 1c5: aload 0
      // 1c6: getfield nb.l J
      // 1c9: iload 1
      // 1ca: i2l
      // 1cb: lneg
      // 1cc: lsub
      // 1cd: lcmp
      // 1ce: ifge 1f9
      // 1d1: aload 0
      // 1d2: getfield nb.t I
      // 1d5: i2l
      // 1d6: aload 0
      // 1d7: getfield nb.j J
      // 1da: ladd
      // 1db: aload 0
      // 1dc: getfield nb.l J
      // 1df: iload 1
      // 1e0: i2l
      // 1e1: lneg
      // 1e2: lsub
      // 1e3: lcmp
      // 1e4: iflt 1f9
      // 1e7: goto 1eb
      // 1ea: athrow
      // 1eb: aload 0
      // 1ec: getfield nb.l J
      // 1ef: iload 1
      // 1f0: i2l
      // 1f1: ladd
      // 1f2: lstore 7
      // 1f4: iload 10
      // 1f6: ifeq 24b
      // 1f9: aload 0
      // 1fa: getfield nb.l J
      // 1fd: ldc2_w -1
      // 200: lxor
      // 201: aload 0
      // 202: getfield nb.t I
      // 205: i2l
      // 206: aload 0
      // 207: getfield nb.j J
      // 20a: ladd
      // 20b: ldc2_w -1
      // 20e: lxor
      // 20f: lcmp
      // 210: ifle 24b
      // 213: goto 217
      // 216: athrow
      // 217: aload 0
      // 218: getfield nb.j J
      // 21b: aload 0
      // 21c: getfield nb.t I
      // 21f: i2l
      // 220: lneg
      // 221: lsub
      // 222: ldc2_w -1
      // 225: lxor
      // 226: aload 0
      // 227: getfield nb.l J
      // 22a: iload 1
      // 22b: i2l
      // 22c: lneg
      // 22d: lsub
      // 22e: ldc2_w -1
      // 231: lxor
      // 232: lcmp
      // 233: ifge 23e
      // 236: goto 23a
      // 239: athrow
      // 23a: goto 24b
      // 23d: athrow
      // 23e: aload 0
      // 23f: getfield nb.j J
      // 242: aload 0
      // 243: getfield nb.t I
      // 246: i2l
      // 247: lneg
      // 248: lsub
      // 249: lstore 7
      // 24b: lconst_0
      // 24c: lload 5
      // 24e: ldc2_w -1
      // 251: lxor
      // 252: lcmp
      // 253: ifle 292
      // 256: lload 7
      // 258: ldc2_w -1
      // 25b: lxor
      // 25c: lload 5
      // 25e: ldc2_w -1
      // 261: lxor
      // 262: lcmp
      // 263: ifge 292
      // 266: goto 26a
      // 269: athrow
      // 26a: lload 7
      // 26c: lload 5
      // 26e: lsub
      // 26f: l2i
      // 270: istore 9
      // 272: aload 4
      // 274: aload 0
      // 275: getfield nb.l J
      // 278: lneg
      // 279: lload 5
      // 27b: ladd
      // 27c: iload 3
      // 27d: i2l
      // 27e: ladd
      // 27f: l2i
      // 280: aload 0
      // 281: getfield nb.n [B
      // 284: aload 0
      // 285: getfield nb.j J
      // 288: lneg
      // 289: lload 5
      // 28b: ladd
      // 28c: l2i
      // 28d: iload 9
      // 28f: invokestatic ab.a ([BI[BII)V
      // 292: aload 0
      // 293: dup
      // 294: getfield nb.l J
      // 297: iload 1
      // 298: i2l
      // 299: ladd
      // 29a: putfield nb.l J
      // 29d: return
      // 29e: iload 1
      // 29f: ifgt 2a6
      // 2a2: goto 30d
      // 2a5: athrow
      // 2a6: aload 0
      // 2a7: getfield nb.h J
      // 2aa: ldc2_w -1
      // 2ad: lcmp
      // 2ae: ifne 2bd
      // 2b1: aload 0
      // 2b2: aload 0
      // 2b3: getfield nb.l J
      // 2b6: putfield nb.h J
      // 2b9: goto 2bd
      // 2bc: athrow
      // 2bd: aload 4
      // 2bf: iload 3
      // 2c0: aload 0
      // 2c1: getfield nb.a [B
      // 2c4: aload 0
      // 2c5: getfield nb.h J
      // 2c8: lneg
      // 2c9: aload 0
      // 2ca: getfield nb.l J
      // 2cd: ladd
      // 2ce: l2i
      // 2cf: iload 1
      // 2d0: invokestatic ab.a ([BI[BII)V
      // 2d3: aload 0
      // 2d4: dup
      // 2d5: getfield nb.l J
      // 2d8: iload 1
      // 2d9: i2l
      // 2da: ladd
      // 2db: putfield nb.l J
      // 2de: aload 0
      // 2df: getfield nb.i I
      // 2e2: i2l
      // 2e3: ldc2_w -1
      // 2e6: lxor
      // 2e7: aload 0
      // 2e8: getfield nb.l J
      // 2eb: aload 0
      // 2ec: getfield nb.h J
      // 2ef: lneg
      // 2f0: ladd
      // 2f1: ldc2_w -1
      // 2f4: lxor
      // 2f5: lcmp
      // 2f6: ifgt 2fd
      // 2f9: goto 30c
      // 2fc: athrow
      // 2fd: aload 0
      // 2fe: aload 0
      // 2ff: getfield nb.l J
      // 302: aload 0
      // 303: getfield nb.h J
      // 306: lneg
      // 307: ladd
      // 308: l2i
      // 309: putfield nb.i I
      // 30c: return
      // 30d: goto 31c
      // 310: astore 5
      // 312: aload 0
      // 313: ldc2_w -1
      // 316: putfield nb.x J
      // 319: aload 5
      // 31b: athrow
      // 31c: goto 370
      // 31f: astore 5
      // 321: aload 5
      // 323: new java/lang/StringBuilder
      // 326: dup
      // 327: invokespecial java/lang/StringBuilder.<init> ()V
      // 32a: getstatic nb.B [Ljava/lang/String;
      // 32d: bipush 10
      // 32f: aaload
      // 330: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 333: iload 1
      // 334: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 337: bipush 44
      // 339: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 33c: iload 2
      // 33d: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 340: bipush 44
      // 342: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 345: iload 3
      // 346: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 349: bipush 44
      // 34b: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 34e: aload 4
      // 350: ifnull 35c
      // 353: getstatic nb.B [Ljava/lang/String;
      // 356: bipush 2
      // 357: aaload
      // 358: goto 361
      // 35b: athrow
      // 35c: getstatic nb.B [Ljava/lang/String;
      // 35f: bipush 1
      // 360: aaload
      // 361: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 364: bipush 41
      // 366: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 369: invokevirtual java/lang/StringBuilder.toString ()Ljava/lang/String;
      // 36c: invokestatic i.a (Ljava/lang/Throwable;Ljava/lang/String;)Lla;
      // 36f: athrow
      // 370: return
   }

   final void a(long var1, int var3) throws IOException {
      try {
         try {
            m++;
            if (~var1 > -1L) {
               throw new IOException();
            }
         } catch (RuntimeException var5) {
            throw var5;
         }

         this.l = var1;
         int var4 = -39 / ((var3 - -66) / 55);
      } catch (RuntimeException var6) {
         throw i.a(var6, B[11] + var1 + ',' + var3 + ')');
      }
   }

   private final void b(byte param1) throws IOException {
      // $VF: Couldn't be decompiled
      // Please report this to the Vineflower issue tracker, at https://github.com/Vineflower/vineflower/issues with a copy of the class file (if you have the rights to distribute it!)
      // java.lang.RuntimeException: parsing failure!
      //   at org.jetbrains.java.decompiler.modules.decompiler.decompose.DomHelper.parseGraph(DomHelper.java:211)
      //   at org.jetbrains.java.decompiler.main.rels.MethodProcessor.codeToJava(MethodProcessor.java:166)
      //
      // Bytecode:
      // 00: getstatic client.vh Z
      // 03: istore 4
      // 05: aload 0
      // 06: bipush 0
      // 07: putfield nb.t I
      // 0a: getstatic nb.k I
      // 0d: bipush 1
      // 0e: iadd
      // 0f: putstatic nb.k I
      // 12: aload 0
      // 13: getfield nb.l J
      // 16: ldc2_w -1
      // 19: lxor
      // 1a: aload 0
      // 1b: getfield nb.x J
      // 1e: ldc2_w -1
      // 21: lxor
      // 22: lcmp
      // 23: ifne 2a
      // 26: goto 3e
      // 29: athrow
      // 2a: aload 0
      // 2b: getfield nb.b Ld;
      // 2e: bipush 0
      // 2f: aload 0
      // 30: getfield nb.l J
      // 33: invokevirtual d.a (IJ)V
      // 36: aload 0
      // 37: aload 0
      // 38: getfield nb.l J
      // 3b: putfield nb.x J
      // 3e: aload 0
      // 3f: aload 0
      // 40: getfield nb.l J
      // 43: putfield nb.j J
      // 46: aload 0
      // 47: getfield nb.t I
      // 4a: bipush -1
      // 4b: ixor
      // 4c: aload 0
      // 4d: getfield nb.n [B
      // 50: arraylength
      // 51: bipush -1
      // 52: ixor
      // 53: if_icmple b6
      // 56: aload 0
      // 57: getfield nb.t I
      // 5a: ineg
      // 5b: aload 0
      // 5c: getfield nb.n [B
      // 5f: arraylength
      // 60: iadd
      // 61: istore 2
      // 62: iload 2
      // 63: bipush -1
      // 64: ixor
      // 65: ldc -200000001
      // 67: iload 4
      // 69: ifne b9
      // 6c: if_icmpge 76
      // 6f: goto 73
      // 72: athrow
      // 73: ldc 200000000
      // 75: istore 2
      // 76: aload 0
      // 77: getfield nb.b Ld;
      // 7a: aload 0
      // 7b: getfield nb.n [B
      // 7e: iload 2
      // 7f: aload 0
      // 80: getfield nb.t I
      // 83: bipush -1
      // 84: invokevirtual d.a ([BIII)I
      // 87: istore 3
      // 88: bipush 0
      // 89: iload 3
      // 8a: bipush -1
      // 8b: ixor
      // 8c: if_icmpne 98
      // 8f: iload 4
      // 91: ifeq b6
      // 94: goto 98
      // 97: athrow
      // 98: aload 0
      // 99: dup
      // 9a: getfield nb.x J
      // 9d: iload 3
      // 9e: i2l
      // 9f: ladd
      // a0: putfield nb.x J
      // a3: aload 0
      // a4: dup
      // a5: getfield nb.t I
      // a8: iload 3
      // a9: iadd
      // aa: putfield nb.t I
      // ad: iload 4
      // af: ifeq 46
      // b2: goto b6
      // b5: athrow
      // b6: iload 1
      // b7: bipush 14
      // b9: if_icmpgt c8
      // bc: aload 0
      // bd: aconst_null
      // be: checkcast d
      // c1: putfield nb.b Ld;
      // c4: goto c8
      // c7: athrow
      // c8: goto ed
      // cb: astore 2
      // cc: aload 2
      // cd: new java/lang/StringBuilder
      // d0: dup
      // d1: invokespecial java/lang/StringBuilder.<init> ()V
      // d4: getstatic nb.B [Ljava/lang/String;
      // d7: bipush 7
      // d9: aaload
      // da: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // dd: iload 1
      // de: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // e1: bipush 41
      // e3: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // e6: invokevirtual java/lang/StringBuilder.toString ()Ljava/lang/String;
      // e9: invokestatic i.a (Ljava/lang/Throwable;Ljava/lang/String;)Lla;
      // ec: athrow
      // ed: return
   }

   nb(d var1, int var2, int var3) throws IOException {
      try {
         this.b = var1;
         this.o = this.A = var1.a((byte)47);
         this.a = new byte[var3];
         this.l = 0L;
         this.n = new byte[var2];
      } catch (RuntimeException var6) {
         RuntimeException var4 = var6;

         RuntimeException var10000;
         StringBuilder var10001;
         try {
            var10000 = var4;
            var10001 = new StringBuilder().append(B[3]);
            if (var1 != null) {
               throw i.a(var4, var10001.append(B[2]).append(',').append(var2).append(',').append(var3).append(')').toString());
            }
         } catch (RuntimeException var5) {
            throw var5;
         }

         throw i.a(var10000, var10001.append(B[1]).append(',').append(var2).append(',').append(var3).append(')').toString());
      }
   }

   private static char[] z(String var0) {
      char[] var10000 = var0.toCharArray();
      if (var10000.length < 2) {
         var10000[0] = (char)(var10000[0] ^ '*');
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
               var10005 = 35;
               break;
            case 1:
               var10005 = 63;
               break;
            case 2:
               var10005 = 26;
               break;
            case 3:
               var10005 = 58;
               break;
            default:
               var10005 = 42;
         }

         var10001[var1] = (char)(var10004 ^ var10005);
      }

      return new String(var10001).intern();
   }
}
