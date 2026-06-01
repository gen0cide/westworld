import java.io.DataInputStream;
import java.net.URL;

final class jb implements Runnable {
   static int m;
   private c b;
   static int i;
   private g f;
   static int d;
   static int j;
   private DataInputStream q;
   private tb h;
   static int e;
   private URL g;
   private g c;
   static int n;
   static int[] k;
   static int p;
   private g l;
   private int a;
   static int o = 0;
   private static final String[] z = new String[]{
      z(z("o\u007f+$\u000f")),
      z(z("MZ4lZI\u00103")),
      z(z("my]Y}fz:")),
      z(z("MZ4\\\u0007")),
      z(z("-2")),
      z(z("MZ4Z\u0007")),
      z(z("\\\u001640R")),
      z(z("IMvr")),
      z(z("MZ4\"FIQn \u0007")),
      z(z("MZ4xFIYvwUB\u00103")),
      z(z("MZ4]\u0007"))
   };

   @Override
   public final void run() {
      boolean var4 = client.vh;

      try {
         m++;

         try {
            int var10000;
            int var10001;
            while (true) {
               label55:
               if (~this.h.w > ~this.h.F.length) {
                  int var1 = this.q.read(this.h.F, this.h.w, this.h.F.length - this.h.w);

                  try {
                     var10000 = ~var1;
                     var10001 = -1;
                     if (var4) {
                        break;
                     }

                     if (var10000 > -1) {
                        break label55;
                     }
                  } catch (Exception var8) {
                     throw var8;
                  }

                  try {
                     this.h.w += var1;
                     if (!var4) {
                        continue;
                     }
                  } catch (Exception var7) {
                     throw var7;
                  }
               }

               var10000 = this.h.w;
               var10001 = this.h.F.length;
               break;
            }

            try {
               if (var10000 == var10001) {
                  throw new Exception(z[0] + this.h.F.length + " " + this.g);
               }
            } catch (Exception var9) {
               throw var9;
            }

            synchronized (this) {
               this.finalize();
               this.a = 3;
            }
         } catch (Exception var10) {
            synchronized (this) {
               this.finalize();
               this.a++;
            }
         }
      } catch (RuntimeException var11) {
         throw i.a(var11, z[1]);
      }
   }

   static final void a(
      int[] param0,
      int param1,
      int param2,
      int param3,
      int param4,
      int param5,
      int param6,
      int param7,
      int param8,
      int param9,
      int[] param10,
      boolean param11,
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
      // 005: getstatic jb.d I
      // 008: bipush 1
      // 009: iadd
      // 00a: putstatic jb.d I
      // 00d: iload 7
      // 00f: bipush -1
      // 010: ixor
      // 011: bipush -1
      // 012: if_icmpge 018
      // 015: goto 019
      // 018: return
      // 019: bipush 0
      // 01a: istore 15
      // 01c: bipush 0
      // 01d: istore 16
      // 01f: iload 3
      // 020: bipush -1
      // 021: ixor
      // 022: bipush -1
      // 023: if_icmpne 02a
      // 026: goto 03c
      // 029: athrow
      // 02a: iload 13
      // 02c: iload 3
      // 02d: idiv
      // 02e: ldc 223739206
      // 030: ishl
      // 031: istore 16
      // 033: iload 8
      // 035: iload 3
      // 036: idiv
      // 037: ldc -902154106
      // 039: ishl
      // 03a: istore 15
      // 03c: iload 4
      // 03e: bipush 2
      // 03f: ishl
      // 040: istore 4
      // 042: iload 11
      // 044: ifeq 050
      // 047: bipush 21
      // 049: putstatic jb.o I
      // 04c: goto 050
      // 04f: athrow
      // 050: iload 15
      // 052: bipush -1
      // 053: ixor
      // 054: bipush -1
      // 055: if_icmple 060
      // 058: bipush 0
      // 059: istore 15
      // 05b: iload 22
      // 05d: ifeq 071
      // 060: sipush 4032
      // 063: iload 15
      // 065: if_icmpge 071
      // 068: goto 06c
      // 06b: athrow
      // 06c: sipush 4032
      // 06f: istore 15
      // 071: iload 7
      // 073: istore 19
      // 075: bipush 0
      // 076: iload 19
      // 078: if_icmpge 542
      // 07b: iload 3
      // 07c: iload 2
      // 07d: iadd
      // 07e: istore 3
      // 07f: iload 15
      // 081: istore 14
      // 083: iload 8
      // 085: iload 12
      // 087: iadd
      // 088: istore 8
      // 08a: iload 16
      // 08c: istore 9
      // 08e: iload 13
      // 090: iload 1
      // 091: iadd
      // 092: istore 13
      // 094: iload 22
      // 096: ifne 617
      // 099: iload 3
      // 09a: bipush -1
      // 09b: ixor
      // 09c: bipush -1
      // 09d: if_icmpne 0a8
      // 0a0: goto 0a4
      // 0a3: athrow
      // 0a4: goto 0ba
      // 0a7: athrow
      // 0a8: iload 8
      // 0aa: iload 3
      // 0ab: idiv
      // 0ac: ldc -213852602
      // 0ae: ishl
      // 0af: istore 15
      // 0b1: iload 13
      // 0b3: iload 3
      // 0b4: idiv
      // 0b5: ldc -474023130
      // 0b7: ishl
      // 0b8: istore 16
      // 0ba: iload 15
      // 0bc: bipush -1
      // 0bd: ixor
      // 0be: bipush -1
      // 0bf: if_icmpgt 0d8
      // 0c2: sipush 4032
      // 0c5: iload 15
      // 0c7: if_icmpge 0db
      // 0ca: goto 0ce
      // 0cd: athrow
      // 0ce: sipush 4032
      // 0d1: istore 15
      // 0d3: iload 22
      // 0d5: ifeq 0db
      // 0d8: bipush 0
      // 0d9: istore 15
      // 0db: iload 16
      // 0dd: iload 9
      // 0df: isub
      // 0e0: ldc -1841585212
      // 0e2: ishr
      // 0e3: istore 18
      // 0e5: iload 14
      // 0e7: ineg
      // 0e8: iload 15
      // 0ea: iadd
      // 0eb: ldc 166246532
      // 0ed: ishr
      // 0ee: istore 17
      // 0f0: iload 5
      // 0f2: ldc -1249879148
      // 0f4: ishr
      // 0f5: istore 20
      // 0f7: iload 14
      // 0f9: iload 5
      // 0fb: ldc 786432
      // 0fd: iand
      // 0fe: iadd
      // 0ff: istore 14
      // 101: iload 5
      // 103: iload 4
      // 105: iadd
      // 106: istore 5
      // 108: iload 19
      // 10a: bipush 16
      // 10c: if_icmplt 4c7
      // 10f: aload 0
      // 110: iload 6
      // 112: iinc 6 1
      // 115: aload 0
      // 116: iload 6
      // 118: iaload
      // 119: ldc -496952415
      // 11b: ishr
      // 11c: ldc 8355711
      // 11e: invokestatic ib.a (II)I
      // 121: aload 10
      // 123: sipush 4032
      // 126: iload 9
      // 128: invokestatic ib.a (II)I
      // 12b: iload 14
      // 12d: ldc 955305670
      // 12f: ishr
      // 130: iadd
      // 131: iaload
      // 132: iload 20
      // 134: iushr
      // 135: iadd
      // 136: iastore
      // 137: iload 14
      // 139: iload 17
      // 13b: iadd
      // 13c: istore 14
      // 13e: iload 9
      // 140: iload 18
      // 142: iadd
      // 143: istore 9
      // 145: aload 0
      // 146: iload 6
      // 148: iinc 6 1
      // 14b: aload 0
      // 14c: iload 6
      // 14e: iaload
      // 14f: ldc 16711423
      // 151: invokestatic ib.a (II)I
      // 154: ldc 393665345
      // 156: ishr
      // 157: aload 10
      // 159: sipush 4032
      // 15c: iload 9
      // 15e: invokestatic ib.a (II)I
      // 161: iload 14
      // 163: ldc 556791558
      // 165: ishr
      // 166: ineg
      // 167: isub
      // 168: iaload
      // 169: iload 20
      // 16b: iushr
      // 16c: iadd
      // 16d: iastore
      // 16e: iload 9
      // 170: iload 18
      // 172: iadd
      // 173: istore 9
      // 175: iload 14
      // 177: iload 17
      // 179: iadd
      // 17a: istore 14
      // 17c: aload 0
      // 17d: iload 6
      // 17f: iinc 6 1
      // 182: ldc 16711423
      // 184: aload 0
      // 185: iload 6
      // 187: iaload
      // 188: invokestatic ib.a (II)I
      // 18b: ldc -2007060127
      // 18d: ishr
      // 18e: aload 10
      // 190: iload 9
      // 192: sipush 4032
      // 195: invokestatic ib.a (II)I
      // 198: iload 14
      // 19a: ldc -1069632730
      // 19c: ishr
      // 19d: iadd
      // 19e: iaload
      // 19f: iload 20
      // 1a1: iushr
      // 1a2: iadd
      // 1a3: iastore
      // 1a4: iload 9
      // 1a6: iload 18
      // 1a8: iadd
      // 1a9: istore 9
      // 1ab: iload 14
      // 1ad: iload 17
      // 1af: iadd
      // 1b0: istore 14
      // 1b2: aload 0
      // 1b3: iload 6
      // 1b5: iinc 6 1
      // 1b8: aload 0
      // 1b9: iload 6
      // 1bb: iaload
      // 1bc: ldc 16711423
      // 1be: invokestatic ib.a (II)I
      // 1c1: ldc -526841663
      // 1c3: ishr
      // 1c4: aload 10
      // 1c6: iload 14
      // 1c8: ldc -1891324570
      // 1ca: ishr
      // 1cb: sipush 4032
      // 1ce: iload 9
      // 1d0: invokestatic ib.a (II)I
      // 1d3: iadd
      // 1d4: iaload
      // 1d5: iload 20
      // 1d7: iushr
      // 1d8: iadd
      // 1d9: iastore
      // 1da: iload 14
      // 1dc: iload 17
      // 1de: iadd
      // 1df: istore 14
      // 1e1: iload 9
      // 1e3: iload 18
      // 1e5: iadd
      // 1e6: istore 9
      // 1e8: iload 5
      // 1ea: ldc 786432
      // 1ec: iand
      // 1ed: sipush 4095
      // 1f0: iload 14
      // 1f2: iand
      // 1f3: iadd
      // 1f4: istore 14
      // 1f6: iload 5
      // 1f8: ldc -580603052
      // 1fa: ishr
      // 1fb: istore 20
      // 1fd: aload 0
      // 1fe: iload 6
      // 200: iinc 6 1
      // 203: aload 10
      // 205: iload 9
      // 207: sipush 4032
      // 20a: invokestatic ib.a (II)I
      // 20d: iload 14
      // 20f: ldc 1328606726
      // 211: ishr
      // 212: iadd
      // 213: iaload
      // 214: iload 20
      // 216: iushr
      // 217: aload 0
      // 218: iload 6
      // 21a: iaload
      // 21b: ldc 16711422
      // 21d: invokestatic ib.a (II)I
      // 220: ldc 604787489
      // 222: ishr
      // 223: iadd
      // 224: iastore
      // 225: iload 5
      // 227: iload 4
      // 229: iadd
      // 22a: istore 5
      // 22c: iload 14
      // 22e: iload 17
      // 230: iadd
      // 231: istore 14
      // 233: iload 9
      // 235: iload 18
      // 237: iadd
      // 238: istore 9
      // 23a: aload 0
      // 23b: iload 6
      // 23d: iinc 6 1
      // 240: aload 10
      // 242: iload 14
      // 244: ldc -830951482
      // 246: ishr
      // 247: sipush 4032
      // 24a: iload 9
      // 24c: invokestatic ib.a (II)I
      // 24f: iadd
      // 250: iaload
      // 251: iload 20
      // 253: iushr
      // 254: aload 0
      // 255: iload 6
      // 257: iaload
      // 258: ldc 16711423
      // 25a: invokestatic ib.a (II)I
      // 25d: ldc 310428257
      // 25f: ishr
      // 260: iadd
      // 261: iastore
      // 262: iload 14
      // 264: iload 17
      // 266: iadd
      // 267: istore 14
      // 269: iload 9
      // 26b: iload 18
      // 26d: iadd
      // 26e: istore 9
      // 270: aload 0
      // 271: iload 6
      // 273: iinc 6 1
      // 276: aload 10
      // 278: sipush 4032
      // 27b: iload 9
      // 27d: invokestatic ib.a (II)I
      // 280: iload 14
      // 282: ldc -1841159226
      // 284: ishr
      // 285: iadd
      // 286: iaload
      // 287: iload 20
      // 289: iushr
      // 28a: aload 0
      // 28b: iload 6
      // 28d: iaload
      // 28e: ldc 16711423
      // 290: invokestatic ib.a (II)I
      // 293: ldc -1760233471
      // 295: ishr
      // 296: ineg
      // 297: isub
      // 298: iastore
      // 299: iload 9
      // 29b: iload 18
      // 29d: iadd
      // 29e: istore 9
      // 2a0: iload 14
      // 2a2: iload 17
      // 2a4: iadd
      // 2a5: istore 14
      // 2a7: aload 0
      // 2a8: iload 6
      // 2aa: iinc 6 1
      // 2ad: aload 10
      // 2af: sipush 4032
      // 2b2: iload 9
      // 2b4: invokestatic ib.a (II)I
      // 2b7: iload 14
      // 2b9: ldc 1454319654
      // 2bb: ishr
      // 2bc: iadd
      // 2bd: iaload
      // 2be: iload 20
      // 2c0: iushr
      // 2c1: ldc 16711423
      // 2c3: aload 0
      // 2c4: iload 6
      // 2c6: iaload
      // 2c7: invokestatic ib.a (II)I
      // 2ca: ldc 1605358369
      // 2cc: ishr
      // 2cd: ineg
      // 2ce: isub
      // 2cf: iastore
      // 2d0: iload 14
      // 2d2: iload 17
      // 2d4: iadd
      // 2d5: istore 14
      // 2d7: iload 9
      // 2d9: iload 18
      // 2db: iadd
      // 2dc: istore 9
      // 2de: ldc 786432
      // 2e0: iload 5
      // 2e2: iand
      // 2e3: sipush 4095
      // 2e6: iload 14
      // 2e8: iand
      // 2e9: iadd
      // 2ea: istore 14
      // 2ec: iload 5
      // 2ee: ldc 1147218452
      // 2f0: ishr
      // 2f1: istore 20
      // 2f3: aload 0
      // 2f4: iload 6
      // 2f6: iinc 6 1
      // 2f9: aload 10
      // 2fb: sipush 4032
      // 2fe: iload 9
      // 300: invokestatic ib.a (II)I
      // 303: iload 14
      // 305: ldc 1983636742
      // 307: ishr
      // 308: ineg
      // 309: isub
      // 30a: iaload
      // 30b: iload 20
      // 30d: iushr
      // 30e: aload 0
      // 30f: iload 6
      // 311: iaload
      // 312: ldc 16711422
      // 314: invokestatic ib.a (II)I
      // 317: ldc 1637168449
      // 319: ishr
      // 31a: ineg
      // 31b: isub
      // 31c: iastore
      // 31d: iload 5
      // 31f: iload 4
      // 321: iadd
      // 322: istore 5
      // 324: iload 14
      // 326: iload 17
      // 328: iadd
      // 329: istore 14
      // 32b: iload 9
      // 32d: iload 18
      // 32f: iadd
      // 330: istore 9
      // 332: aload 0
      // 333: iload 6
      // 335: iinc 6 1
      // 338: aload 10
      // 33a: iload 9
      // 33c: sipush 4032
      // 33f: invokestatic ib.a (II)I
      // 342: iload 14
      // 344: ldc 1901625030
      // 346: ishr
      // 347: ineg
      // 348: isub
      // 349: iaload
      // 34a: iload 20
      // 34c: iushr
      // 34d: aload 0
      // 34e: iload 6
      // 350: iaload
      // 351: ldc -256795167
      // 353: ishr
      // 354: ldc 8355711
      // 356: invokestatic ib.a (II)I
      // 359: ineg
      // 35a: isub
      // 35b: iastore
      // 35c: iload 9
      // 35e: iload 18
      // 360: iadd
      // 361: istore 9
      // 363: iload 14
      // 365: iload 17
      // 367: iadd
      // 368: istore 14
      // 36a: aload 0
      // 36b: iload 6
      // 36d: iinc 6 1
      // 370: aload 10
      // 372: sipush 4032
      // 375: iload 9
      // 377: invokestatic ib.a (II)I
      // 37a: iload 14
      // 37c: ldc 1605754694
      // 37e: ishr
      // 37f: iadd
      // 380: iaload
      // 381: iload 20
      // 383: iushr
      // 384: aload 0
      // 385: iload 6
      // 387: iaload
      // 388: ldc 16711423
      // 38a: invokestatic ib.a (II)I
      // 38d: ldc -216359295
      // 38f: ishr
      // 390: ineg
      // 391: isub
      // 392: iastore
      // 393: iload 14
      // 395: iload 17
      // 397: iadd
      // 398: istore 14
      // 39a: iload 9
      // 39c: iload 18
      // 39e: iadd
      // 39f: istore 9
      // 3a1: aload 0
      // 3a2: iload 6
      // 3a4: iinc 6 1
      // 3a7: ldc 16711422
      // 3a9: aload 0
      // 3aa: iload 6
      // 3ac: iaload
      // 3ad: invokestatic ib.a (II)I
      // 3b0: ldc 791103809
      // 3b2: ishr
      // 3b3: aload 10
      // 3b5: iload 14
      // 3b7: ldc 371413222
      // 3b9: ishr
      // 3ba: iload 9
      // 3bc: sipush 4032
      // 3bf: invokestatic ib.a (II)I
      // 3c2: iadd
      // 3c3: iaload
      // 3c4: iload 20
      // 3c6: iushr
      // 3c7: iadd
      // 3c8: iastore
      // 3c9: iload 14
      // 3cb: iload 17
      // 3cd: iadd
      // 3ce: istore 14
      // 3d0: iload 9
      // 3d2: iload 18
      // 3d4: iadd
      // 3d5: istore 9
      // 3d7: iload 5
      // 3d9: ldc 786432
      // 3db: iand
      // 3dc: iload 14
      // 3de: sipush 4095
      // 3e1: iand
      // 3e2: iadd
      // 3e3: istore 14
      // 3e5: iload 5
      // 3e7: ldc 711720340
      // 3e9: ishr
      // 3ea: istore 20
      // 3ec: aload 0
      // 3ed: iload 6
      // 3ef: iinc 6 1
      // 3f2: aload 10
      // 3f4: iload 9
      // 3f6: sipush 4032
      // 3f9: invokestatic ib.a (II)I
      // 3fc: iload 14
      // 3fe: ldc -2780922
      // 400: ishr
      // 401: ineg
      // 402: isub
      // 403: iaload
      // 404: iload 20
      // 406: iushr
      // 407: ldc 16711422
      // 409: aload 0
      // 40a: iload 6
      // 40c: iaload
      // 40d: invokestatic ib.a (II)I
      // 410: ldc 962756193
      // 412: ishr
      // 413: ineg
      // 414: isub
      // 415: iastore
      // 416: iload 5
      // 418: iload 4
      // 41a: iadd
      // 41b: istore 5
      // 41d: iload 14
      // 41f: iload 17
      // 421: iadd
      // 422: istore 14
      // 424: iload 9
      // 426: iload 18
      // 428: iadd
      // 429: istore 9
      // 42b: aload 0
      // 42c: iload 6
      // 42e: iinc 6 1
      // 431: aload 0
      // 432: iload 6
      // 434: iaload
      // 435: ldc 16711423
      // 437: invokestatic ib.a (II)I
      // 43a: ldc -838985215
      // 43c: ishr
      // 43d: aload 10
      // 43f: iload 9
      // 441: sipush 4032
      // 444: invokestatic ib.a (II)I
      // 447: iload 14
      // 449: ldc 1389805702
      // 44b: ishr
      // 44c: ineg
      // 44d: isub
      // 44e: iaload
      // 44f: iload 20
      // 451: iushr
      // 452: iadd
      // 453: iastore
      // 454: iload 14
      // 456: iload 17
      // 458: iadd
      // 459: istore 14
      // 45b: iload 9
      // 45d: iload 18
      // 45f: iadd
      // 460: istore 9
      // 462: aload 0
      // 463: iload 6
      // 465: iinc 6 1
      // 468: aload 10
      // 46a: sipush 4032
      // 46d: iload 9
      // 46f: invokestatic ib.a (II)I
      // 472: iload 14
      // 474: ldc -1171869722
      // 476: ishr
      // 477: ineg
      // 478: isub
      // 479: iaload
      // 47a: iload 20
      // 47c: iushr
      // 47d: aload 0
      // 47e: iload 6
      // 480: iaload
      // 481: ldc 1072420929
      // 483: ishr
      // 484: ldc 8355711
      // 486: invokestatic ib.a (II)I
      // 489: ineg
      // 48a: isub
      // 48b: iastore
      // 48c: iload 9
      // 48e: iload 18
      // 490: iadd
      // 491: istore 9
      // 493: iload 14
      // 495: iload 17
      // 497: iadd
      // 498: istore 14
      // 49a: aload 0
      // 49b: iload 6
      // 49d: iinc 6 1
      // 4a0: aload 10
      // 4a2: iload 14
      // 4a4: ldc 227032774
      // 4a6: ishr
      // 4a7: sipush 4032
      // 4aa: iload 9
      // 4ac: invokestatic ib.a (II)I
      // 4af: iadd
      // 4b0: iaload
      // 4b1: iload 20
      // 4b3: iushr
      // 4b4: aload 0
      // 4b5: iload 6
      // 4b7: iaload
      // 4b8: ldc 16711423
      // 4ba: invokestatic ib.a (II)I
      // 4bd: ldc -454180287
      // 4bf: ishr
      // 4c0: iadd
      // 4c1: iastore
      // 4c2: iload 22
      // 4c4: ifeq 53a
      // 4c7: bipush 0
      // 4c8: istore 21
      // 4ca: iload 19
      // 4cc: iload 21
      // 4ce: if_icmple 53a
      // 4d1: aload 0
      // 4d2: iload 6
      // 4d4: iinc 6 1
      // 4d7: aload 10
      // 4d9: iload 14
      // 4db: ldc -208962138
      // 4dd: ishr
      // 4de: iload 9
      // 4e0: sipush 4032
      // 4e3: invokestatic ib.a (II)I
      // 4e6: iadd
      // 4e7: iaload
      // 4e8: iload 20
      // 4ea: iushr
      // 4eb: ldc 16711422
      // 4ed: aload 0
      // 4ee: iload 6
      // 4f0: iaload
      // 4f1: invokestatic ib.a (II)I
      // 4f4: ldc -1883934911
      // 4f6: ishr
      // 4f7: ineg
      // 4f8: isub
      // 4f9: iastore
      // 4fa: iload 9
      // 4fc: iload 18
      // 4fe: iadd
      // 4ff: istore 9
      // 501: iload 14
      // 503: iload 17
      // 505: iadd
      // 506: istore 14
      // 508: bipush 3
      // 509: iload 21
      // 50b: bipush 3
      // 50c: iand
      // 50d: iload 22
      // 50f: ifne 078
      // 512: if_icmpne 532
      // 515: iload 5
      // 517: ldc -2030987340
      // 519: ishr
      // 51a: istore 20
      // 51c: iload 14
      // 51e: sipush 4095
      // 521: iand
      // 522: ldc 786432
      // 524: iload 5
      // 526: iand
      // 527: ineg
      // 528: isub
      // 529: istore 14
      // 52b: iload 5
      // 52d: iload 4
      // 52f: iadd
      // 530: istore 5
      // 532: iinc 21 1
      // 535: iload 22
      // 537: ifeq 4ca
      // 53a: iinc 19 -16
      // 53d: iload 22
      // 53f: ifeq 075
      // 542: goto 617
      // 545: astore 15
      // 547: aload 15
      // 549: new java/lang/StringBuilder
      // 54c: dup
      // 54d: invokespecial java/lang/StringBuilder.<init> ()V
      // 550: getstatic jb.z [Ljava/lang/String;
      // 553: bipush 5
      // 554: aaload
      // 555: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 558: aload 0
      // 559: ifnull 566
      // 55c: getstatic jb.z [Ljava/lang/String;
      // 55f: bipush 6
      // 561: aaload
      // 562: goto 56c
      // 565: athrow
      // 566: getstatic jb.z [Ljava/lang/String;
      // 569: bipush 7
      // 56b: aaload
      // 56c: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 56f: bipush 44
      // 571: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 574: iload 1
      // 575: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 578: bipush 44
      // 57a: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 57d: iload 2
      // 57e: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 581: bipush 44
      // 583: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 586: iload 3
      // 587: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 58a: bipush 44
      // 58c: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 58f: iload 4
      // 591: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 594: bipush 44
      // 596: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 599: iload 5
      // 59b: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 59e: bipush 44
      // 5a0: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 5a3: iload 6
      // 5a5: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 5a8: bipush 44
      // 5aa: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 5ad: iload 7
      // 5af: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 5b2: bipush 44
      // 5b4: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 5b7: iload 8
      // 5b9: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 5bc: bipush 44
      // 5be: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 5c1: iload 9
      // 5c3: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 5c6: bipush 44
      // 5c8: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 5cb: aload 10
      // 5cd: ifnull 5da
      // 5d0: getstatic jb.z [Ljava/lang/String;
      // 5d3: bipush 6
      // 5d5: aaload
      // 5d6: goto 5e0
      // 5d9: athrow
      // 5da: getstatic jb.z [Ljava/lang/String;
      // 5dd: bipush 7
      // 5df: aaload
      // 5e0: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 5e3: bipush 44
      // 5e5: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 5e8: iload 11
      // 5ea: invokevirtual java/lang/StringBuilder.append (Z)Ljava/lang/StringBuilder;
      // 5ed: bipush 44
      // 5ef: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 5f2: iload 12
      // 5f4: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 5f7: bipush 44
      // 5f9: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 5fc: iload 13
      // 5fe: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 601: bipush 44
      // 603: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 606: iload 14
      // 608: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 60b: bipush 41
      // 60d: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 610: invokevirtual java/lang/StringBuilder.toString ()Ljava/lang/String;
      // 613: invokestatic i.a (Ljava/lang/Throwable;Ljava/lang/String;)Lla;
      // 616: athrow
      // 617: return
   }

   final synchronized boolean a(int param1) {
      // $VF: Couldn't be decompiled
      // Please report this to the Vineflower issue tracker, at https://github.com/Vineflower/vineflower/issues with a copy of the class file (if you have the rights to distribute it!)
      // java.lang.RuntimeException: parsing failure!
      //   at org.jetbrains.java.decompiler.modules.decompiler.decompose.DomHelper.parseGraph(DomHelper.java:211)
      //   at org.jetbrains.java.decompiler.main.rels.MethodProcessor.codeToJava(MethodProcessor.java:166)
      //
      // Bytecode:
      // 000: getstatic jb.e I
      // 003: bipush 1
      // 004: iadd
      // 005: putstatic jb.e I
      // 008: aload 0
      // 009: getfield jb.a I
      // 00c: bipush 2
      // 00d: if_icmpge 014
      // 010: goto 016
      // 013: athrow
      // 014: bipush 1
      // 015: ireturn
      // 016: aload 0
      // 017: getfield jb.a I
      // 01a: bipush -1
      // 01b: ixor
      // 01c: bipush -1
      // 01d: if_icmpeq 024
      // 020: goto 06d
      // 023: athrow
      // 024: aconst_null
      // 025: aload 0
      // 026: getfield jb.c Lg;
      // 029: if_acmpeq 030
      // 02c: goto 041
      // 02f: athrow
      // 030: aload 0
      // 031: aload 0
      // 032: getfield jb.b Lc;
      // 035: bipush 74
      // 037: aload 0
      // 038: getfield jb.g Ljava/net/URL;
      // 03b: invokevirtual c.a (BLjava/net/URL;)Lg;
      // 03e: putfield jb.c Lg;
      // 041: bipush 0
      // 042: aload 0
      // 043: getfield jb.c Lg;
      // 046: getfield g.b I
      // 049: if_icmpne 04e
      // 04c: bipush 0
      // 04d: ireturn
      // 04e: bipush -2
      // 050: aload 0
      // 051: getfield jb.c Lg;
      // 054: getfield g.b I
      // 057: bipush -1
      // 058: ixor
      // 059: if_icmpeq 06d
      // 05c: aload 0
      // 05d: aconst_null
      // 05e: putfield jb.c Lg;
      // 061: aload 0
      // 062: dup
      // 063: getfield jb.a I
      // 066: bipush 1
      // 067: iadd
      // 068: putfield jb.a I
      // 06b: bipush 0
      // 06c: ireturn
      // 06d: bipush -2
      // 06f: aload 0
      // 070: getfield jb.a I
      // 073: bipush -1
      // 074: ixor
      // 075: if_icmpne 0ca
      // 078: aconst_null
      // 079: aload 0
      // 07a: getfield jb.f Lg;
      // 07d: if_acmpeq 088
      // 080: goto 084
      // 083: athrow
      // 084: goto 09f
      // 087: athrow
      // 088: aload 0
      // 089: aload 0
      // 08a: getfield jb.b Lc;
      // 08d: aload 0
      // 08e: getfield jb.g Ljava/net/URL;
      // 091: invokevirtual java/net/URL.getHost ()Ljava/lang/String;
      // 094: sipush 443
      // 097: bipush -68
      // 099: invokevirtual c.a (Ljava/lang/String;II)Lg;
      // 09c: putfield jb.f Lg;
      // 09f: aload 0
      // 0a0: getfield jb.f Lg;
      // 0a3: getfield g.b I
      // 0a6: bipush -1
      // 0a7: ixor
      // 0a8: bipush -1
      // 0a9: if_icmpne 0ae
      // 0ac: bipush 0
      // 0ad: ireturn
      // 0ae: aload 0
      // 0af: getfield jb.f Lg;
      // 0b2: getfield g.b I
      // 0b5: bipush 1
      // 0b6: if_icmpeq 0ca
      // 0b9: aload 0
      // 0ba: aconst_null
      // 0bb: putfield jb.f Lg;
      // 0be: aload 0
      // 0bf: dup
      // 0c0: getfield jb.a I
      // 0c3: bipush 1
      // 0c4: iadd
      // 0c5: putfield jb.a I
      // 0c8: bipush 0
      // 0c9: ireturn
      // 0ca: aload 0
      // 0cb: getfield jb.q Ljava/io/DataInputStream;
      // 0ce: ifnull 0d5
      // 0d1: goto 16f
      // 0d4: athrow
      // 0d5: aload 0
      // 0d6: getfield jb.a I
      // 0d9: bipush -1
      // 0da: ixor
      // 0db: bipush -1
      // 0dc: if_icmpeq 0e3
      // 0df: goto 0f1
      // 0e2: athrow
      // 0e3: aload 0
      // 0e4: aload 0
      // 0e5: getfield jb.c Lg;
      // 0e8: getfield g.d Ljava/lang/Object;
      // 0eb: checkcast java/io/DataInputStream
      // 0ee: putfield jb.q Ljava/io/DataInputStream;
      // 0f1: aload 0
      // 0f2: getfield jb.a I
      // 0f5: bipush -1
      // 0f6: ixor
      // 0f7: bipush -2
      // 0f9: if_icmpne 155
      // 0fc: aload 0
      // 0fd: getfield jb.f Lg;
      // 100: getfield g.d Ljava/lang/Object;
      // 103: checkcast java/net/Socket
      // 106: astore 2
      // 107: aload 2
      // 108: sipush 10000
      // 10b: invokevirtual java/net/Socket.setSoTimeout (I)V
      // 10e: aload 2
      // 10f: invokevirtual java/net/Socket.getOutputStream ()Ljava/io/OutputStream;
      // 112: astore 3
      // 113: aload 3
      // 114: bipush 17
      // 116: invokevirtual java/io/OutputStream.write (I)V
      // 119: aload 3
      // 11a: new java/lang/StringBuilder
      // 11d: dup
      // 11e: invokespecial java/lang/StringBuilder.<init> ()V
      // 121: getstatic jb.z [Ljava/lang/String;
      // 124: bipush 2
      // 125: aaload
      // 126: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 129: aload 0
      // 12a: getfield jb.g Ljava/net/URL;
      // 12d: invokevirtual java/net/URL.getFile ()Ljava/lang/String;
      // 130: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 133: getstatic jb.z [Ljava/lang/String;
      // 136: bipush 4
      // 137: aaload
      // 138: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 13b: invokevirtual java/lang/StringBuilder.toString ()Ljava/lang/String;
      // 13e: bipush -104
      // 140: invokestatic h.a (Ljava/lang/CharSequence;B)[B
      // 143: invokevirtual java/io/OutputStream.write ([B)V
      // 146: aload 0
      // 147: new java/io/DataInputStream
      // 14a: dup
      // 14b: aload 2
      // 14c: invokevirtual java/net/Socket.getInputStream ()Ljava/io/InputStream;
      // 14f: invokespecial java/io/DataInputStream.<init> (Ljava/io/InputStream;)V
      // 152: putfield jb.q Ljava/io/DataInputStream;
      // 155: aload 0
      // 156: getfield jb.h Ltb;
      // 159: bipush 0
      // 15a: putfield tb.w I
      // 15d: goto 16f
      // 160: astore 2
      // 161: aload 0
      // 162: invokevirtual jb.finalize ()V
      // 165: aload 0
      // 166: dup
      // 167: getfield jb.a I
      // 16a: bipush 1
      // 16b: iadd
      // 16c: putfield jb.a I
      // 16f: aload 0
      // 170: getfield jb.l Lg;
      // 173: ifnull 17a
      // 176: goto 188
      // 179: athrow
      // 17a: aload 0
      // 17b: aload 0
      // 17c: getfield jb.b Lc;
      // 17f: bipush 1
      // 180: aload 0
      // 181: bipush 5
      // 182: invokevirtual c.a (ZLjava/lang/Runnable;I)Lg;
      // 185: putfield jb.l Lg;
      // 188: aload 0
      // 189: getfield jb.l Lg;
      // 18c: getfield g.b I
      // 18f: bipush -1
      // 190: ixor
      // 191: bipush -1
      // 192: if_icmpne 197
      // 195: bipush 0
      // 196: ireturn
      // 197: aload 0
      // 198: getfield jb.l Lg;
      // 19b: getfield g.b I
      // 19e: bipush -1
      // 19f: ixor
      // 1a0: iload 1
      // 1a1: if_icmpeq 1b6
      // 1a4: aload 0
      // 1a5: invokevirtual jb.finalize ()V
      // 1a8: aload 0
      // 1a9: dup
      // 1aa: getfield jb.a I
      // 1ad: bipush 1
      // 1ae: iadd
      // 1af: putfield jb.a I
      // 1b2: goto 1b6
      // 1b5: athrow
      // 1b6: bipush 0
      // 1b7: ireturn
      // 1b8: astore 2
      // 1b9: aload 2
      // 1ba: new java/lang/StringBuilder
      // 1bd: dup
      // 1be: invokespecial java/lang/StringBuilder.<init> ()V
      // 1c1: getstatic jb.z [Ljava/lang/String;
      // 1c4: bipush 3
      // 1c5: aaload
      // 1c6: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 1c9: iload 1
      // 1ca: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 1cd: bipush 41
      // 1cf: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 1d2: invokevirtual java/lang/StringBuilder.toString ()Ljava/lang/String;
      // 1d5: invokestatic i.a (Ljava/lang/Throwable;Ljava/lang/String;)Lla;
      // 1d8: athrow
   }

   final tb a(byte var1) {
      try {
         try {
            if (var1 > -110) {
               this.run();
            }
         } catch (RuntimeException var3) {
            throw var3;
         }

         i++;
         return 3 == this.a ? this.h : null;
      } catch (RuntimeException var4) {
         throw i.a(var4, z[10] + var1 + ')');
      }
   }

   static final String a(boolean var0, String var1, String var2, String var3) {
      boolean var5 = client.vh;

      try {
         try {
            if (!var0) {
               a((int[])null, 78, -46, -87, -87, -58, -96, -121, 50, -80, (int[])null, false, -54, -83, 52);
            }
         } catch (RuntimeException var6) {
            throw var6;
         }

         j++;
         int var4 = var3.indexOf(var2);

         String var10000;
         while (true) {
            if (~var4 != 0) {
               var3 = var3.substring(0, var4) + var1 + var3.substring(var2.length() + var4);
               var10000 = var3;
               if (var5) {
                  break;
               }

               var4 = var3.indexOf(var2, var4 - -var1.length());
               if (!var5) {
                  continue;
               }
            }

            var10000 = var3;
            break;
         }

         return var10000;
      } catch (RuntimeException var7) {
         throw var7;
      }
   }

   jb(c var1, URL var2, int var3) {
      try {
         this.g = var2;
         this.b = var1;
         this.h = new tb(var3);
      } catch (RuntimeException var7) {
         RuntimeException var4 = var7;

         RuntimeException var10000;
         StringBuilder var10001;
         String var10002;
         label34: {
            try {
               var10000 = var4;
               var10001 = new StringBuilder().append(z[8]);
               if (var1 != null) {
                  var10002 = z[6];
                  break label34;
               }
            } catch (RuntimeException var6) {
               throw var6;
            }

            var10002 = z[7];
         }

         try {
            var10001 = var10001.append(var10002).append(',');
            if (var2 != null) {
               throw i.a(var10000, var10001.append(z[6]).append(',').append(var3).append(')').toString());
            }
         } catch (RuntimeException var5) {
            throw var5;
         }

         throw i.a(var10000, var10001.append(z[7]).append(',').append(var3).append(')').toString());
      }
   }

   @Override
   protected final void finalize() {
      // $VF: Couldn't be decompiled
      // Please report this to the Vineflower issue tracker, at https://github.com/Vineflower/vineflower/issues with a copy of the class file (if you have the rights to distribute it!)
      // java.lang.RuntimeException: parsing failure!
      //   at org.jetbrains.java.decompiler.modules.decompiler.decompose.DomHelper.parseGraph(DomHelper.java:211)
      //   at org.jetbrains.java.decompiler.main.rels.MethodProcessor.codeToJava(MethodProcessor.java:166)
      //
      // Bytecode:
      // 00: getstatic jb.n I
      // 03: bipush 1
      // 04: iadd
      // 05: putstatic jb.n I
      // 08: aconst_null
      // 09: aload 0
      // 0a: getfield jb.c Lg;
      // 0d: if_acmpeq 35
      // 10: aconst_null
      // 11: aload 0
      // 12: getfield jb.c Lg;
      // 15: getfield g.d Ljava/lang/Object;
      // 18: if_acmpeq 30
      // 1b: goto 1f
      // 1e: athrow
      // 1f: aload 0
      // 20: getfield jb.c Lg;
      // 23: getfield g.d Ljava/lang/Object;
      // 26: checkcast java/io/DataInputStream
      // 29: invokevirtual java/io/DataInputStream.close ()V
      // 2c: goto 30
      // 2f: astore 1
      // 30: aload 0
      // 31: aconst_null
      // 32: putfield jb.c Lg;
      // 35: aconst_null
      // 36: aload 0
      // 37: getfield jb.f Lg;
      // 3a: if_acmpeq 65
      // 3d: aload 0
      // 3e: getfield jb.f Lg;
      // 41: getfield g.d Ljava/lang/Object;
      // 44: ifnonnull 4f
      // 47: goto 4b
      // 4a: athrow
      // 4b: goto 60
      // 4e: athrow
      // 4f: aload 0
      // 50: getfield jb.f Lg;
      // 53: getfield g.d Ljava/lang/Object;
      // 56: checkcast java/net/Socket
      // 59: invokevirtual java/net/Socket.close ()V
      // 5c: goto 60
      // 5f: astore 1
      // 60: aload 0
      // 61: aconst_null
      // 62: putfield jb.f Lg;
      // 65: aconst_null
      // 66: aload 0
      // 67: getfield jb.q Ljava/io/DataInputStream;
      // 6a: if_acmpeq 7d
      // 6d: aload 0
      // 6e: getfield jb.q Ljava/io/DataInputStream;
      // 71: invokevirtual java/io/DataInputStream.close ()V
      // 74: goto 78
      // 77: astore 1
      // 78: aload 0
      // 79: aconst_null
      // 7a: putfield jb.q Ljava/io/DataInputStream;
      // 7d: aload 0
      // 7e: aconst_null
      // 7f: putfield jb.l Lg;
      // 82: goto 91
      // 85: astore 1
      // 86: aload 1
      // 87: getstatic jb.z [Ljava/lang/String;
      // 8a: bipush 9
      // 8c: aaload
      // 8d: invokestatic i.a (Ljava/lang/Throwable;Ljava/lang/String;)Lla;
      // 90: athrow
      // 91: return
   }

   private static char[] z(String var0) {
      char[] var10000 = var0.toCharArray();
      if (var10000.length < 2) {
         var10000[0] = (char)(var10000[0] ^ '/');
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
               var10005 = 39;
               break;
            case 1:
               var10005 = 56;
               break;
            case 2:
               var10005 = 26;
               break;
            case 3:
               var10005 = 30;
               break;
            default:
               var10005 = 47;
         }

         var10001[var1] = (char)(var10004 ^ var10005);
      }

      return new String(var10001).intern();
   }
}
