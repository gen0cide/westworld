final class i {
   static int b;
   static int e;
   static int c;
   int a;
   static int d;
   static int[] g;
   private static long[] h = new long[256];
   static String f;
   private static final String[] z = new String[]{
      z(z("s\u0000\u0011!|sZ\u0013`")), z(z("t[A$")), z(z("a\u0000\u0003fo")), z(z("s\u0000n`")), z(z("s\u0000o`")), z(z("s\u0000Y'An\\D&u2\u0007"))
   };

   @Override
   public final String toString() {
      try {
         e++;
         throw new IllegalStateException();
      } catch (RuntimeException var2) {
         throw a(var2, z[5]);
      }
   }

   static final int a(int param0, int param1, int param2, CharSequence param3, byte param4, byte[] param5) {
      // $VF: Couldn't be decompiled
      // Please report this to the Vineflower issue tracker, at https://github.com/Vineflower/vineflower/issues with a copy of the class file (if you have the rights to distribute it!)
      // java.lang.RuntimeException: parsing failure!
      //   at org.jetbrains.java.decompiler.modules.decompiler.decompose.DomHelper.parseGraph(DomHelper.java:211)
      //   at org.jetbrains.java.decompiler.main.rels.MethodProcessor.codeToJava(MethodProcessor.java:166)
      //
      // Bytecode:
      // 000: getstatic client.vh Z
      // 003: istore 9
      // 005: getstatic i.b I
      // 008: bipush 1
      // 009: iadd
      // 00a: putstatic i.b I
      // 00d: iload 0
      // 00e: iload 2
      // 00f: ineg
      // 010: iadd
      // 011: istore 6
      // 013: iload 4
      // 015: bipush -78
      // 017: if_icmplt 02a
      // 01a: aconst_null
      // 01b: checkcast java/lang/Throwable
      // 01e: aconst_null
      // 01f: checkcast java/lang/String
      // 022: invokestatic i.a (Ljava/lang/Throwable;Ljava/lang/String;)Lla;
      // 025: pop
      // 026: goto 02a
      // 029: athrow
      // 02a: bipush 0
      // 02b: istore 7
      // 02d: iload 7
      // 02f: bipush -1
      // 030: ixor
      // 031: iload 6
      // 033: bipush -1
      // 034: ixor
      // 035: if_icmple 3e7
      // 038: aload 3
      // 039: iload 2
      // 03a: iload 7
      // 03c: iadd
      // 03d: invokeinterface java/lang/CharSequence.charAt (I)C 2
      // 042: istore 8
      // 044: iload 8
      // 046: iload 9
      // 048: ifne 3e9
      // 04b: ifle 05e
      // 04e: goto 052
      // 051: athrow
      // 052: iload 8
      // 054: sipush 128
      // 057: if_icmplt 3d1
      // 05a: goto 05e
      // 05d: athrow
      // 05e: iload 8
      // 060: bipush -1
      // 061: ixor
      // 062: sipush -161
      // 065: if_icmpgt 078
      // 068: goto 06c
      // 06b: athrow
      // 06c: sipush 255
      // 06f: iload 8
      // 071: if_icmpge 3d1
      // 074: goto 078
      // 077: athrow
      // 078: iload 8
      // 07a: sipush 8364
      // 07d: if_icmpne 096
      // 080: goto 084
      // 083: athrow
      // 084: aload 5
      // 086: iload 7
      // 088: iload 1
      // 089: iadd
      // 08a: bipush -128
      // 08c: bastore
      // 08d: iload 9
      // 08f: ifeq 3df
      // 092: goto 096
      // 095: athrow
      // 096: iload 8
      // 098: sipush 8218
      // 09b: if_icmpeq 3bf
      // 09e: goto 0a2
      // 0a1: athrow
      // 0a2: sipush -403
      // 0a5: iload 8
      // 0a7: bipush -1
      // 0a8: ixor
      // 0a9: if_icmpeq 3ad
      // 0ac: goto 0b0
      // 0af: athrow
      // 0b0: iload 8
      // 0b2: bipush -1
      // 0b3: ixor
      // 0b4: sipush -8223
      // 0b7: if_icmpne 0d1
      // 0ba: goto 0be
      // 0bd: athrow
      // 0be: aload 5
      // 0c0: iload 1
      // 0c1: iload 7
      // 0c3: ineg
      // 0c4: isub
      // 0c5: bipush -124
      // 0c7: bastore
      // 0c8: iload 9
      // 0ca: ifeq 3df
      // 0cd: goto 0d1
      // 0d0: athrow
      // 0d1: sipush 8230
      // 0d4: iload 8
      // 0d6: if_icmpeq 39b
      // 0d9: goto 0dd
      // 0dc: athrow
      // 0dd: iload 8
      // 0df: sipush 8224
      // 0e2: if_icmpne 0fc
      // 0e5: goto 0e9
      // 0e8: athrow
      // 0e9: aload 5
      // 0eb: iload 1
      // 0ec: iload 7
      // 0ee: ineg
      // 0ef: isub
      // 0f0: bipush -122
      // 0f2: bastore
      // 0f3: iload 9
      // 0f5: ifeq 3df
      // 0f8: goto 0fc
      // 0fb: athrow
      // 0fc: iload 8
      // 0fe: sipush 8225
      // 101: if_icmpne 11a
      // 104: goto 108
      // 107: athrow
      // 108: aload 5
      // 10a: iload 7
      // 10c: iload 1
      // 10d: iadd
      // 10e: bipush -121
      // 110: bastore
      // 111: iload 9
      // 113: ifeq 3df
      // 116: goto 11a
      // 119: athrow
      // 11a: iload 8
      // 11c: bipush -1
      // 11d: ixor
      // 11e: sipush -711
      // 121: if_icmpeq 389
      // 124: goto 128
      // 127: athrow
      // 128: iload 8
      // 12a: bipush -1
      // 12b: ixor
      // 12c: sipush -8241
      // 12f: if_icmpne 148
      // 132: goto 136
      // 135: athrow
      // 136: aload 5
      // 138: iload 7
      // 13a: iload 1
      // 13b: iadd
      // 13c: bipush -119
      // 13e: bastore
      // 13f: iload 9
      // 141: ifeq 3df
      // 144: goto 148
      // 147: athrow
      // 148: sipush 352
      // 14b: iload 8
      // 14d: if_icmpeq 377
      // 150: goto 154
      // 153: athrow
      // 154: sipush 8249
      // 157: iload 8
      // 159: if_icmpeq 365
      // 15c: goto 160
      // 15f: athrow
      // 160: iload 8
      // 162: sipush 338
      // 165: if_icmpeq 353
      // 168: goto 16c
      // 16b: athrow
      // 16c: iload 8
      // 16e: bipush -1
      // 16f: ixor
      // 170: sipush -382
      // 173: if_icmpeq 341
      // 176: goto 17a
      // 179: athrow
      // 17a: iload 8
      // 17c: bipush -1
      // 17d: ixor
      // 17e: sipush -8217
      // 181: if_icmpeq 32f
      // 184: goto 188
      // 187: athrow
      // 188: iload 8
      // 18a: bipush -1
      // 18b: ixor
      // 18c: sipush -8218
      // 18f: if_icmpeq 31d
      // 192: goto 196
      // 195: athrow
      // 196: sipush 8220
      // 199: iload 8
      // 19b: if_icmpne 1b5
      // 19e: goto 1a2
      // 1a1: athrow
      // 1a2: aload 5
      // 1a4: iload 1
      // 1a5: iload 7
      // 1a7: ineg
      // 1a8: isub
      // 1a9: bipush -109
      // 1ab: bastore
      // 1ac: iload 9
      // 1ae: ifeq 3df
      // 1b1: goto 1b5
      // 1b4: athrow
      // 1b5: sipush -8222
      // 1b8: iload 8
      // 1ba: bipush -1
      // 1bb: ixor
      // 1bc: if_icmpne 1d5
      // 1bf: goto 1c3
      // 1c2: athrow
      // 1c3: aload 5
      // 1c5: iload 7
      // 1c7: iload 1
      // 1c8: iadd
      // 1c9: bipush -108
      // 1cb: bastore
      // 1cc: iload 9
      // 1ce: ifeq 3df
      // 1d1: goto 1d5
      // 1d4: athrow
      // 1d5: iload 8
      // 1d7: sipush 8226
      // 1da: if_icmpne 1f4
      // 1dd: goto 1e1
      // 1e0: athrow
      // 1e1: aload 5
      // 1e3: iload 1
      // 1e4: iload 7
      // 1e6: ineg
      // 1e7: isub
      // 1e8: bipush -107
      // 1ea: bastore
      // 1eb: iload 9
      // 1ed: ifeq 3df
      // 1f0: goto 1f4
      // 1f3: athrow
      // 1f4: sipush -8212
      // 1f7: iload 8
      // 1f9: bipush -1
      // 1fa: ixor
      // 1fb: if_icmpeq 30b
      // 1fe: goto 202
      // 201: athrow
      // 202: iload 8
      // 204: sipush 8212
      // 207: if_icmpeq 2f9
      // 20a: goto 20e
      // 20d: athrow
      // 20e: sipush 732
      // 211: iload 8
      // 213: if_icmpeq 2e7
      // 216: goto 21a
      // 219: athrow
      // 21a: iload 8
      // 21c: bipush -1
      // 21d: ixor
      // 21e: sipush -8483
      // 221: if_icmpne 23a
      // 224: goto 228
      // 227: athrow
      // 228: aload 5
      // 22a: iload 7
      // 22c: iload 1
      // 22d: iadd
      // 22e: bipush -103
      // 230: bastore
      // 231: iload 9
      // 233: ifeq 3df
      // 236: goto 23a
      // 239: athrow
      // 23a: iload 8
      // 23c: sipush 353
      // 23f: if_icmpne 259
      // 242: goto 246
      // 245: athrow
      // 246: aload 5
      // 248: iload 1
      // 249: iload 7
      // 24b: ineg
      // 24c: isub
      // 24d: bipush -102
      // 24f: bastore
      // 250: iload 9
      // 252: ifeq 3df
      // 255: goto 259
      // 258: athrow
      // 259: sipush 8250
      // 25c: iload 8
      // 25e: if_icmpeq 2d5
      // 261: goto 265
      // 264: athrow
      // 265: iload 8
      // 267: bipush -1
      // 268: ixor
      // 269: sipush -340
      // 26c: if_icmpne 285
      // 26f: goto 273
      // 272: athrow
      // 273: aload 5
      // 275: iload 1
      // 276: iload 7
      // 278: iadd
      // 279: bipush -100
      // 27b: bastore
      // 27c: iload 9
      // 27e: ifeq 3df
      // 281: goto 285
      // 284: athrow
      // 285: sipush 382
      // 288: iload 8
      // 28a: if_icmpeq 2c2
      // 28d: goto 291
      // 290: athrow
      // 291: sipush 376
      // 294: iload 8
      // 296: if_icmpeq 2af
      // 299: goto 29d
      // 29c: athrow
      // 29d: aload 5
      // 29f: iload 7
      // 2a1: iload 1
      // 2a2: iadd
      // 2a3: bipush 63
      // 2a5: bastore
      // 2a6: iload 9
      // 2a8: ifeq 3df
      // 2ab: goto 2af
      // 2ae: athrow
      // 2af: aload 5
      // 2b1: iload 1
      // 2b2: iload 7
      // 2b4: ineg
      // 2b5: isub
      // 2b6: bipush -97
      // 2b8: bastore
      // 2b9: iload 9
      // 2bb: ifeq 3df
      // 2be: goto 2c2
      // 2c1: athrow
      // 2c2: aload 5
      // 2c4: iload 1
      // 2c5: iload 7
      // 2c7: ineg
      // 2c8: isub
      // 2c9: bipush -98
      // 2cb: bastore
      // 2cc: iload 9
      // 2ce: ifeq 3df
      // 2d1: goto 2d5
      // 2d4: athrow
      // 2d5: aload 5
      // 2d7: iload 7
      // 2d9: iload 1
      // 2da: iadd
      // 2db: bipush -101
      // 2dd: bastore
      // 2de: iload 9
      // 2e0: ifeq 3df
      // 2e3: goto 2e7
      // 2e6: athrow
      // 2e7: aload 5
      // 2e9: iload 7
      // 2eb: iload 1
      // 2ec: iadd
      // 2ed: bipush -104
      // 2ef: bastore
      // 2f0: iload 9
      // 2f2: ifeq 3df
      // 2f5: goto 2f9
      // 2f8: athrow
      // 2f9: aload 5
      // 2fb: iload 1
      // 2fc: iload 7
      // 2fe: iadd
      // 2ff: bipush -105
      // 301: bastore
      // 302: iload 9
      // 304: ifeq 3df
      // 307: goto 30b
      // 30a: athrow
      // 30b: aload 5
      // 30d: iload 1
      // 30e: iload 7
      // 310: iadd
      // 311: bipush -106
      // 313: bastore
      // 314: iload 9
      // 316: ifeq 3df
      // 319: goto 31d
      // 31c: athrow
      // 31d: aload 5
      // 31f: iload 7
      // 321: iload 1
      // 322: iadd
      // 323: bipush -110
      // 325: bastore
      // 326: iload 9
      // 328: ifeq 3df
      // 32b: goto 32f
      // 32e: athrow
      // 32f: aload 5
      // 331: iload 7
      // 333: iload 1
      // 334: iadd
      // 335: bipush -111
      // 337: bastore
      // 338: iload 9
      // 33a: ifeq 3df
      // 33d: goto 341
      // 340: athrow
      // 341: aload 5
      // 343: iload 7
      // 345: iload 1
      // 346: iadd
      // 347: bipush -114
      // 349: bastore
      // 34a: iload 9
      // 34c: ifeq 3df
      // 34f: goto 353
      // 352: athrow
      // 353: aload 5
      // 355: iload 7
      // 357: iload 1
      // 358: iadd
      // 359: bipush -116
      // 35b: bastore
      // 35c: iload 9
      // 35e: ifeq 3df
      // 361: goto 365
      // 364: athrow
      // 365: aload 5
      // 367: iload 7
      // 369: iload 1
      // 36a: iadd
      // 36b: bipush -117
      // 36d: bastore
      // 36e: iload 9
      // 370: ifeq 3df
      // 373: goto 377
      // 376: athrow
      // 377: aload 5
      // 379: iload 7
      // 37b: iload 1
      // 37c: iadd
      // 37d: bipush -118
      // 37f: bastore
      // 380: iload 9
      // 382: ifeq 3df
      // 385: goto 389
      // 388: athrow
      // 389: aload 5
      // 38b: iload 1
      // 38c: iload 7
      // 38e: iadd
      // 38f: bipush -120
      // 391: bastore
      // 392: iload 9
      // 394: ifeq 3df
      // 397: goto 39b
      // 39a: athrow
      // 39b: aload 5
      // 39d: iload 7
      // 39f: iload 1
      // 3a0: iadd
      // 3a1: bipush -123
      // 3a3: bastore
      // 3a4: iload 9
      // 3a6: ifeq 3df
      // 3a9: goto 3ad
      // 3ac: athrow
      // 3ad: aload 5
      // 3af: iload 1
      // 3b0: iload 7
      // 3b2: iadd
      // 3b3: bipush -125
      // 3b5: bastore
      // 3b6: iload 9
      // 3b8: ifeq 3df
      // 3bb: goto 3bf
      // 3be: athrow
      // 3bf: aload 5
      // 3c1: iload 7
      // 3c3: iload 1
      // 3c4: iadd
      // 3c5: bipush -126
      // 3c7: bastore
      // 3c8: iload 9
      // 3ca: ifeq 3df
      // 3cd: goto 3d1
      // 3d0: athrow
      // 3d1: aload 5
      // 3d3: iload 7
      // 3d5: iload 1
      // 3d6: iadd
      // 3d7: iload 8
      // 3d9: i2b
      // 3da: bastore
      // 3db: goto 3df
      // 3de: athrow
      // 3df: iinc 7 1
      // 3e2: iload 9
      // 3e4: ifeq 02d
      // 3e7: iload 6
      // 3e9: ireturn
      // 3ea: astore 6
      // 3ec: aload 6
      // 3ee: new java/lang/StringBuilder
      // 3f1: dup
      // 3f2: invokespecial java/lang/StringBuilder.<init> ()V
      // 3f5: getstatic i.z [Ljava/lang/String;
      // 3f8: bipush 4
      // 3f9: aaload
      // 3fa: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 3fd: iload 0
      // 3fe: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 401: bipush 44
      // 403: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 406: iload 1
      // 407: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 40a: bipush 44
      // 40c: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 40f: iload 2
      // 410: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 413: bipush 44
      // 415: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 418: aload 3
      // 419: ifnull 425
      // 41c: getstatic i.z [Ljava/lang/String;
      // 41f: bipush 2
      // 420: aaload
      // 421: goto 42a
      // 424: athrow
      // 425: getstatic i.z [Ljava/lang/String;
      // 428: bipush 1
      // 429: aaload
      // 42a: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 42d: bipush 44
      // 42f: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 432: iload 4
      // 434: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 437: bipush 44
      // 439: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 43c: aload 5
      // 43e: ifnull 44a
      // 441: getstatic i.z [Ljava/lang/String;
      // 444: bipush 2
      // 445: aaload
      // 446: goto 44f
      // 449: athrow
      // 44a: getstatic i.z [Ljava/lang/String;
      // 44d: bipush 1
      // 44e: aaload
      // 44f: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 452: bipush 41
      // 454: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 457: invokevirtual java/lang/StringBuilder.toString ()Ljava/lang/String;
      // 45a: invokestatic i.a (Ljava/lang/Throwable;Ljava/lang/String;)Lla;
      // 45d: athrow
   }

   i(String var1, int var2) {
      try {
         this.a = var2;
      } catch (RuntimeException var5) {
         RuntimeException var3 = var5;

         RuntimeException var10000;
         StringBuilder var10001;
         try {
            var10000 = var3;
            var10001 = new StringBuilder().append(z[0]);
            if (var1 != null) {
               throw a(var3, var10001.append(z[2]).append(',').append(var2).append(')').toString());
            }
         } catch (RuntimeException var4) {
            throw var4;
         }

         throw a(var10000, var10001.append(z[1]).append(',').append(var2).append(')').toString());
      }
   }

   static final la a(Throwable var0, String var1) {
      try {
         c++;
         la var2;
         if (var0 instanceof la) {
            var2 = (la)var0;
            var2.h = var2.h + ' ' + var1;
         } else {
            var2 = new la(var0, var1);
         }

         return var2;
      } catch (RuntimeException var3) {
         throw var3;
      }
   }

   static final v[] a(int var0) {
      try {
         try {
            d++;
            if (var0 != -711) {
               a((Throwable)null, (String)null);
            }
         } catch (RuntimeException var2) {
            throw var2;
         }

         return new v[]{ua.E, da.O, ga.c, ta.f, la.b, eb.d, gb.n};
      } catch (RuntimeException var3) {
         throw a(var3, z[3] + var0 + ')');
      }
   }

   static {
      for (int var2 = 0; var2 < 256; var2++) {
         long var0 = (long)var2;
         int var3 = 0;

         while (true) {
            label28: {
               label27: {
                  try {
                     if (8 <= var3) {
                        break;
                     }

                     if (~(1L & var0) == -2L) {
                        break label27;
                     }
                  } catch (IllegalStateException var4) {
                     throw var4;
                  }

                  var0 >>>= 1;
                  break label28;
               }

               var0 = var0 >>> 1 ^ -3932672073523589310L;
            }

            var3++;
         }

         h[var2] = var0;
      }

      f = z(
         z(
            "[ln\fW\\ie\u0001XQb`\u0006]J\u007f\u007f\u001bFOxz\u0010K@OO+v\u007fHJ {pEA%|u^\\:an[[?jcT\u001dy )\u001a\u0018~%\"\u0017\fj±>\u000bsn82\u0007\u0000\u0017/1uV\u0015o!\u0014\n\b1d\u0002\u0011f,5\u0011q42"
         )
      );
   }

   private static char[] z(String var0) {
      char[] var10000 = var0.toCharArray();
      if (var10000.length < 2) {
         var10000[0] = (char)(var10000[0] ^ 18);
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
               var10005 = 26;
               break;
            case 1:
               var10005 = 46;
               break;
            case 2:
               var10005 = 45;
               break;
            case 3:
               var10005 = 72;
               break;
            default:
               var10005 = 18;
         }

         var10001[var1] = (char)(var10004 ^ var10005);
      }

      return new String(var10001).intern();
   }
}
