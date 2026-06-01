final class ac {
   byte[] j;
   int L;
   int E;
   int G;
   byte[] q;
   boolean[] n;
   static String[] z = new String[200];
   int[][] J = new int[6][258];
   static String[] x;
   int[] r;
   int h;
   int F;
   byte[] i;
   static int C;
   byte[] d;
   byte[] s;
   int f;
   byte[] A;
   int c;
   int[] D;
   static int[] l;
   int o;
   int y;
   int p;
   int[][] t;
   int[] w;
   int a;
   static char[] I = new char[]{'[', ']', '#'};
   int b;
   static int k;
   int[] m;
   int e;
   int H;
   boolean[] v;
   int K;
   int[][] u;
   byte g;
   byte[][] B;
   private static final String[] M = new String[]{z(z("h_og%")), z(z("gI-J")), z(z("h_od%")), z(z("r\u0012o\bp"))};

   static final void a(ib var0, byte var1, ib var2) {
      try {
         try {
            if (null != var0.e) {
               var0.a(-27331);
            }
         } catch (RuntimeException var4) {
            throw var4;
         }

         try {
            k++;
            if (var1 != 34) {
               x = (String[])null;
            }
         } catch (RuntimeException var7) {
            throw var7;
         }

         var0.a = var2;
         var0.e = var2.e;
         var0.e.a = var0;
         var0.a.e = var0;
      } catch (RuntimeException var8) {
         RuntimeException var3 = var8;

         RuntimeException var10000;
         StringBuilder var10001;
         String var10002;
         label45: {
            try {
               var10000 = var3;
               var10001 = new StringBuilder().append(M[2]);
               if (var0 != null) {
                  var10002 = M[3];
                  break label45;
               }
            } catch (RuntimeException var6) {
               throw var6;
            }

            var10002 = M[1];
         }

         try {
            var10001 = var10001.append(var10002).append(',').append((int)var1).append(',');
            if (var2 != null) {
               throw i.a(var10000, var10001.append(M[3]).append(')').toString());
            }
         } catch (RuntimeException var5) {
            throw var5;
         }

         throw i.a(var10000, var10001.append(M[1]).append(')').toString());
      }
   }

   static final char a(char param0, int param1) {
      // $VF: Couldn't be decompiled
      // Please report this to the Vineflower issue tracker, at https://github.com/Vineflower/vineflower/issues with a copy of the class file (if you have the rights to distribute it!)
      // java.lang.RuntimeException: parsing failure!
      //   at org.jetbrains.java.decompiler.modules.decompiler.decompose.DomHelper.parseGraph(DomHelper.java:211)
      //   at org.jetbrains.java.decompiler.main.rels.MethodProcessor.codeToJava(MethodProcessor.java:166)
      //
      // Bytecode:
      // 000: getstatic client.vh Z
      // 003: istore 3
      // 004: iload 1
      // 005: sipush -194
      // 008: if_icmpeq 017
      // 00b: ldc 65414
      // 00d: bipush 87
      // 00f: invokestatic ac.a (CI)C
      // 012: pop
      // 013: goto 017
      // 016: athrow
      // 017: getstatic ac.C I
      // 01a: bipush 1
      // 01b: iadd
      // 01c: putstatic ac.C I
      // 01f: iload 0
      // 020: istore 2
      // 021: bipush -33
      // 023: iload 2
      // 024: bipush -1
      // 025: ixor
      // 026: if_icmpne 031
      // 029: iload 3
      // 02a: ifeq 38f
      // 02d: goto 031
      // 030: athrow
      // 031: sipush 160
      // 034: iload 2
      // 035: if_icmpeq 38f
      // 038: goto 03c
      // 03b: athrow
      // 03c: bipush 95
      // 03e: iload 2
      // 03f: if_icmpeq 38f
      // 042: goto 046
      // 045: athrow
      // 046: bipush -46
      // 048: iload 2
      // 049: bipush -1
      // 04a: ixor
      // 04b: if_icmpne 05a
      // 04e: goto 052
      // 051: athrow
      // 052: iload 3
      // 053: ifeq 38f
      // 056: goto 05a
      // 059: athrow
      // 05a: iload 2
      // 05b: bipush -1
      // 05c: ixor
      // 05d: bipush -92
      // 05f: if_icmpne 06e
      // 062: goto 066
      // 065: athrow
      // 066: iload 3
      // 067: ifeq 392
      // 06a: goto 06e
      // 06d: athrow
      // 06e: bipush -94
      // 070: iload 2
      // 071: bipush -1
      // 072: ixor
      // 073: if_icmpeq 392
      // 076: goto 07a
      // 079: athrow
      // 07a: bipush 35
      // 07c: iload 2
      // 07d: if_icmpne 08c
      // 080: goto 084
      // 083: athrow
      // 084: iload 3
      // 085: ifeq 392
      // 088: goto 08c
      // 08b: athrow
      // 08c: sipush 224
      // 08f: iload 2
      // 090: if_icmpne 09f
      // 093: goto 097
      // 096: athrow
      // 097: iload 3
      // 098: ifeq 394
      // 09b: goto 09f
      // 09e: athrow
      // 09f: sipush 225
      // 0a2: iload 2
      // 0a3: if_icmpne 0b2
      // 0a6: goto 0aa
      // 0a9: athrow
      // 0aa: iload 3
      // 0ab: ifeq 394
      // 0ae: goto 0b2
      // 0b1: athrow
      // 0b2: iload 2
      // 0b3: bipush -1
      // 0b4: ixor
      // 0b5: sipush -227
      // 0b8: if_icmpne 0c7
      // 0bb: goto 0bf
      // 0be: athrow
      // 0bf: iload 3
      // 0c0: ifeq 394
      // 0c3: goto 0c7
      // 0c6: athrow
      // 0c7: iload 2
      // 0c8: sipush 228
      // 0cb: if_icmpne 0da
      // 0ce: goto 0d2
      // 0d1: athrow
      // 0d2: iload 3
      // 0d3: ifeq 394
      // 0d6: goto 0da
      // 0d9: athrow
      // 0da: iload 2
      // 0db: sipush 227
      // 0de: if_icmpeq 394
      // 0e1: goto 0e5
      // 0e4: athrow
      // 0e5: iload 2
      // 0e6: sipush 192
      // 0e9: if_icmpne 0f8
      // 0ec: goto 0f0
      // 0ef: athrow
      // 0f0: iload 3
      // 0f1: ifeq 394
      // 0f4: goto 0f8
      // 0f7: athrow
      // 0f8: sipush -194
      // 0fb: iload 2
      // 0fc: bipush -1
      // 0fd: ixor
      // 0fe: if_icmpne 10d
      // 101: goto 105
      // 104: athrow
      // 105: iload 3
      // 106: ifeq 394
      // 109: goto 10d
      // 10c: athrow
      // 10d: sipush 194
      // 110: iload 2
      // 111: if_icmpeq 394
      // 114: goto 118
      // 117: athrow
      // 118: iload 2
      // 119: sipush 196
      // 11c: if_icmpeq 394
      // 11f: goto 123
      // 122: athrow
      // 123: sipush 195
      // 126: iload 2
      // 127: if_icmpne 136
      // 12a: goto 12e
      // 12d: athrow
      // 12e: iload 3
      // 12f: ifeq 394
      // 132: goto 136
      // 135: athrow
      // 136: iload 2
      // 137: sipush 232
      // 13a: if_icmpeq 397
      // 13d: goto 141
      // 140: athrow
      // 141: sipush -234
      // 144: iload 2
      // 145: bipush -1
      // 146: ixor
      // 147: if_icmpne 156
      // 14a: goto 14e
      // 14d: athrow
      // 14e: iload 3
      // 14f: ifeq 397
      // 152: goto 156
      // 155: athrow
      // 156: sipush 234
      // 159: iload 2
      // 15a: if_icmpeq 397
      // 15d: goto 161
      // 160: athrow
      // 161: sipush 235
      // 164: iload 2
      // 165: if_icmpeq 397
      // 168: goto 16c
      // 16b: athrow
      // 16c: iload 2
      // 16d: bipush -1
      // 16e: ixor
      // 16f: sipush -201
      // 172: if_icmpeq 397
      // 175: goto 179
      // 178: athrow
      // 179: sipush 201
      // 17c: iload 2
      // 17d: if_icmpeq 397
      // 180: goto 184
      // 183: athrow
      // 184: iload 2
      // 185: sipush 202
      // 188: if_icmpeq 397
      // 18b: goto 18f
      // 18e: athrow
      // 18f: iload 2
      // 190: bipush -1
      // 191: ixor
      // 192: sipush -204
      // 195: if_icmpeq 397
      // 198: goto 19c
      // 19b: athrow
      // 19c: iload 2
      // 19d: sipush 237
      // 1a0: if_icmpeq 39a
      // 1a3: goto 1a7
      // 1a6: athrow
      // 1a7: iload 2
      // 1a8: sipush 238
      // 1ab: if_icmpne 1ba
      // 1ae: goto 1b2
      // 1b1: athrow
      // 1b2: iload 3
      // 1b3: ifeq 39a
      // 1b6: goto 1ba
      // 1b9: athrow
      // 1ba: iload 2
      // 1bb: bipush -1
      // 1bc: ixor
      // 1bd: sipush -240
      // 1c0: if_icmpeq 39a
      // 1c3: goto 1c7
      // 1c6: athrow
      // 1c7: iload 2
      // 1c8: bipush -1
      // 1c9: ixor
      // 1ca: sipush -206
      // 1cd: if_icmpne 1dc
      // 1d0: goto 1d4
      // 1d3: athrow
      // 1d4: iload 3
      // 1d5: ifeq 39a
      // 1d8: goto 1dc
      // 1db: athrow
      // 1dc: iload 2
      // 1dd: bipush -1
      // 1de: ixor
      // 1df: sipush -207
      // 1e2: if_icmpne 1f1
      // 1e5: goto 1e9
      // 1e8: athrow
      // 1e9: iload 3
      // 1ea: ifeq 39a
      // 1ed: goto 1f1
      // 1f0: athrow
      // 1f1: iload 2
      // 1f2: bipush -1
      // 1f3: ixor
      // 1f4: sipush -208
      // 1f7: if_icmpeq 39a
      // 1fa: goto 1fe
      // 1fd: athrow
      // 1fe: sipush -243
      // 201: iload 2
      // 202: bipush -1
      // 203: ixor
      // 204: if_icmpne 213
      // 207: goto 20b
      // 20a: athrow
      // 20b: iload 3
      // 20c: ifeq 39d
      // 20f: goto 213
      // 212: athrow
      // 213: sipush -244
      // 216: iload 2
      // 217: bipush -1
      // 218: ixor
      // 219: if_icmpeq 39d
      // 21c: goto 220
      // 21f: athrow
      // 220: sipush -245
      // 223: iload 2
      // 224: bipush -1
      // 225: ixor
      // 226: if_icmpne 235
      // 229: goto 22d
      // 22c: athrow
      // 22d: iload 3
      // 22e: ifeq 39d
      // 231: goto 235
      // 234: athrow
      // 235: sipush -247
      // 238: iload 2
      // 239: bipush -1
      // 23a: ixor
      // 23b: if_icmpeq 39d
      // 23e: goto 242
      // 241: athrow
      // 242: iload 2
      // 243: sipush 245
      // 246: if_icmpne 255
      // 249: goto 24d
      // 24c: athrow
      // 24d: iload 3
      // 24e: ifeq 39d
      // 251: goto 255
      // 254: athrow
      // 255: iload 2
      // 256: sipush 210
      // 259: if_icmpeq 39d
      // 25c: goto 260
      // 25f: athrow
      // 260: iload 2
      // 261: bipush -1
      // 262: ixor
      // 263: sipush -212
      // 266: if_icmpne 275
      // 269: goto 26d
      // 26c: athrow
      // 26d: iload 3
      // 26e: ifeq 39d
      // 271: goto 275
      // 274: athrow
      // 275: iload 2
      // 276: sipush 212
      // 279: if_icmpeq 39d
      // 27c: goto 280
      // 27f: athrow
      // 280: iload 2
      // 281: bipush -1
      // 282: ixor
      // 283: sipush -215
      // 286: if_icmpeq 39d
      // 289: goto 28d
      // 28c: athrow
      // 28d: sipush -214
      // 290: iload 2
      // 291: bipush -1
      // 292: ixor
      // 293: if_icmpeq 39d
      // 296: goto 29a
      // 299: athrow
      // 29a: sipush 249
      // 29d: iload 2
      // 29e: if_icmpne 2ad
      // 2a1: goto 2a5
      // 2a4: athrow
      // 2a5: iload 3
      // 2a6: ifeq 3a0
      // 2a9: goto 2ad
      // 2ac: athrow
      // 2ad: sipush -251
      // 2b0: iload 2
      // 2b1: bipush -1
      // 2b2: ixor
      // 2b3: if_icmpne 2c2
      // 2b6: goto 2ba
      // 2b9: athrow
      // 2ba: iload 3
      // 2bb: ifeq 3a0
      // 2be: goto 2c2
      // 2c1: athrow
      // 2c2: sipush -252
      // 2c5: iload 2
      // 2c6: bipush -1
      // 2c7: ixor
      // 2c8: if_icmpne 2d7
      // 2cb: goto 2cf
      // 2ce: athrow
      // 2cf: iload 3
      // 2d0: ifeq 3a0
      // 2d3: goto 2d7
      // 2d6: athrow
      // 2d7: sipush 252
      // 2da: iload 2
      // 2db: if_icmpeq 3a0
      // 2de: goto 2e2
      // 2e1: athrow
      // 2e2: iload 2
      // 2e3: bipush -1
      // 2e4: ixor
      // 2e5: sipush -218
      // 2e8: if_icmpeq 3a0
      // 2eb: goto 2ef
      // 2ee: athrow
      // 2ef: iload 2
      // 2f0: sipush 218
      // 2f3: if_icmpeq 3a0
      // 2f6: goto 2fa
      // 2f9: athrow
      // 2fa: iload 2
      // 2fb: sipush 219
      // 2fe: if_icmpne 30d
      // 301: goto 305
      // 304: athrow
      // 305: iload 3
      // 306: ifeq 3a0
      // 309: goto 30d
      // 30c: athrow
      // 30d: sipush 220
      // 310: iload 2
      // 311: if_icmpeq 3a0
      // 314: goto 318
      // 317: athrow
      // 318: sipush -232
      // 31b: iload 2
      // 31c: bipush -1
      // 31d: ixor
      // 31e: if_icmpeq 3a3
      // 321: goto 325
      // 324: athrow
      // 325: sipush 199
      // 328: iload 2
      // 329: if_icmpeq 3a3
      // 32c: goto 330
      // 32f: athrow
      // 330: sipush 255
      // 333: iload 2
      // 334: if_icmpne 343
      // 337: goto 33b
      // 33a: athrow
      // 33b: iload 3
      // 33c: ifeq 3a6
      // 33f: goto 343
      // 342: athrow
      // 343: sipush -377
      // 346: iload 2
      // 347: bipush -1
      // 348: ixor
      // 349: if_icmpne 358
      // 34c: goto 350
      // 34f: athrow
      // 350: iload 3
      // 351: ifeq 3a6
      // 354: goto 358
      // 357: athrow
      // 358: iload 2
      // 359: sipush 241
      // 35c: if_icmpne 36b
      // 35f: goto 363
      // 362: athrow
      // 363: iload 3
      // 364: ifeq 3a9
      // 367: goto 36b
      // 36a: athrow
      // 36b: sipush 209
      // 36e: iload 2
      // 36f: if_icmpne 37e
      // 372: goto 376
      // 375: athrow
      // 376: iload 3
      // 377: ifeq 3a9
      // 37a: goto 37e
      // 37d: athrow
      // 37e: sipush -224
      // 381: iload 2
      // 382: bipush -1
      // 383: ixor
      // 384: if_icmpeq 3ac
      // 387: goto 38b
      // 38a: athrow
      // 38b: goto 3af
      // 38e: athrow
      // 38f: bipush 95
      // 391: ireturn
      // 392: iload 0
      // 393: ireturn
      // 394: bipush 97
      // 396: ireturn
      // 397: bipush 101
      // 399: ireturn
      // 39a: bipush 105
      // 39c: ireturn
      // 39d: bipush 111
      // 39f: ireturn
      // 3a0: bipush 117
      // 3a2: ireturn
      // 3a3: bipush 99
      // 3a5: ireturn
      // 3a6: bipush 121
      // 3a8: ireturn
      // 3a9: bipush 110
      // 3ab: ireturn
      // 3ac: bipush 98
      // 3ae: ireturn
      // 3af: iload 0
      // 3b0: invokestatic java/lang/Character.toLowerCase (C)C
      // 3b3: ireturn
      // 3b4: astore 2
      // 3b5: aload 2
      // 3b6: new java/lang/StringBuilder
      // 3b9: dup
      // 3ba: invokespecial java/lang/StringBuilder.<init> ()V
      // 3bd: getstatic ac.M [Ljava/lang/String;
      // 3c0: bipush 0
      // 3c1: aaload
      // 3c2: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 3c5: iload 0
      // 3c6: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 3c9: bipush 44
      // 3cb: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 3ce: iload 1
      // 3cf: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 3d2: bipush 41
      // 3d4: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 3d7: invokevirtual java/lang/StringBuilder.toString ()Ljava/lang/String;
      // 3da: invokestatic i.a (Ljava/lang/Throwable;Ljava/lang/String;)Lla;
      // 3dd: athrow
   }

   ac() {
      this.j = new byte[18002];
      this.n = new boolean[256];
      this.m = new int[256];
      this.r = new int[16];
      this.s = new byte[18002];
      this.A = new byte[4096];
      this.o = 0;
      this.w = new int[257];
      this.u = new int[6][258];
      this.D = new int[6];
      this.d = new byte[256];
      this.a = 0;
      this.v = new boolean[16];
      this.B = new byte[6][258];
      this.t = new int[6][258];
   }

   private static char[] z(String var0) {
      char[] var10000 = var0.toCharArray();
      if (var10000.length < 2) {
         var10000[0] = (char)(var10000[0] ^ '\r');
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
               var10005 = 9;
               break;
            case 1:
               var10005 = 60;
               break;
            case 2:
               var10005 = 65;
               break;
            case 3:
               var10005 = 38;
               break;
            default:
               var10005 = 13;
         }

         var10001[var1] = (char)(var10004 ^ var10005);
      }

      return new String(var10001).intern();
   }
}
