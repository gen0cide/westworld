final class o {
   static int b;
   static int j;
   static int o;
   static int h;
   private int i;
   static int[] a;
   private int d;
   static String[] g = new String[]{
      z(z("}9~3\nS>;`\u0006N.;j\u001cIklz\u0000Tko|SO rcSH#~3\u0007I?ta\u001a]'")),
      z(z("]%\u007f3\u0007Y'~c\u001cN?;g\u001c\u001c\u0007n~\u0011N\"\u007ft\u0016\u0003"))
   };
   private int[] n;
   private int[] k;
   static int[] p;
   static int c;
   private int m;
   static int e;
   private int f;
   static String l = "";
   private static final String[] z = new String[]{
      z(z("SeY;")), z(z("SeZ;")), z(z("Se^;")), z(z("Ge5=\u000e")), z(z("R>w\u007f")), z(z("Se'z\u001dU?%;")), z(z("Se_;")), z(z("SeX;")), z(z("Se];"))
   };

   final int c(int var1) {
      try {
         try {
            if (-1 == ~(this.i--)) {
               this.b(-110);
               this.i = 255;
            }
         } catch (RuntimeException var3) {
            throw var3;
         }

         try {
            b++;
            if (var1 > -67) {
               this.d = 32;
            }
         } catch (RuntimeException var4) {
            throw var4;
         }

         return this.k[this.i];
      } catch (RuntimeException var5) {
         throw i.a(var5, z[7] + var1 + 41);
      }
   }

   private final void b(int param1) {
      // $VF: Couldn't be decompiled
      // Please report this to the Vineflower issue tracker, at https://github.com/Vineflower/vineflower/issues with a copy of the class file (if you have the rights to distribute it!)
      // java.lang.RuntimeException: parsing failure!
      //   at org.jetbrains.java.decompiler.modules.decompiler.decompose.DomHelper.parseGraph(DomHelper.java:211)
      //   at org.jetbrains.java.decompiler.main.rels.MethodProcessor.codeToJava(MethodProcessor.java:166)
      //
      // Bytecode:
      // 000: getstatic client.vh Z
      // 003: istore 6
      // 005: getstatic o.h I
      // 008: bipush 1
      // 009: iadd
      // 00a: putstatic o.h I
      // 00d: aload 0
      // 00e: dup
      // 00f: getfield o.d I
      // 012: aload 0
      // 013: dup
      // 014: getfield o.f I
      // 017: bipush 1
      // 018: iadd
      // 019: dup_x1
      // 01a: putfield o.f I
      // 01d: iadd
      // 01e: putfield o.d I
      // 021: iload 1
      // 022: bipush -100
      // 024: if_icmple 032
      // 027: aconst_null
      // 028: checkcast [Ljava/lang/String;
      // 02b: putstatic o.g [Ljava/lang/String;
      // 02e: goto 032
      // 031: athrow
      // 032: bipush 0
      // 033: istore 2
      // 034: sipush -257
      // 037: iload 2
      // 038: bipush -1
      // 039: ixor
      // 03a: if_icmpge 13c
      // 03d: aload 0
      // 03e: getfield o.n [I
      // 041: iload 2
      // 042: iaload
      // 043: istore 3
      // 044: iload 2
      // 045: bipush 3
      // 046: iand
      // 047: istore 5
      // 049: iload 6
      // 04b: ifne 161
      // 04e: iload 5
      // 050: ifne 070
      // 053: goto 057
      // 056: athrow
      // 057: aload 0
      // 058: dup
      // 059: getfield o.m I
      // 05c: aload 0
      // 05d: getfield o.m I
      // 060: ldc -402254995
      // 062: ishl
      // 063: ixor
      // 064: putfield o.m I
      // 067: iload 6
      // 069: ifeq 0da
      // 06c: goto 070
      // 06f: athrow
      // 070: bipush -2
      // 072: iload 5
      // 074: bipush -1
      // 075: ixor
      // 076: if_icmpeq 0c6
      // 079: goto 07d
      // 07c: athrow
      // 07d: bipush -3
      // 07f: iload 5
      // 081: bipush -1
      // 082: ixor
      // 083: if_icmpne 0a3
      // 086: goto 08a
      // 089: athrow
      // 08a: aload 0
      // 08b: dup
      // 08c: getfield o.m I
      // 08f: aload 0
      // 090: getfield o.m I
      // 093: ldc 2019129250
      // 095: ishl
      // 096: ixor
      // 097: putfield o.m I
      // 09a: iload 6
      // 09c: ifeq 0da
      // 09f: goto 0a3
      // 0a2: athrow
      // 0a3: iload 5
      // 0a5: bipush 3
      // 0a6: if_icmpeq 0b1
      // 0a9: goto 0ad
      // 0ac: athrow
      // 0ad: goto 0da
      // 0b0: athrow
      // 0b1: aload 0
      // 0b2: dup
      // 0b3: getfield o.m I
      // 0b6: aload 0
      // 0b7: getfield o.m I
      // 0ba: ldc -350927312
      // 0bc: iushr
      // 0bd: ixor
      // 0be: putfield o.m I
      // 0c1: iload 6
      // 0c3: ifeq 0da
      // 0c6: aload 0
      // 0c7: dup
      // 0c8: getfield o.m I
      // 0cb: aload 0
      // 0cc: getfield o.m I
      // 0cf: ldc -585344026
      // 0d1: iushr
      // 0d2: ixor
      // 0d3: putfield o.m I
      // 0d6: goto 0da
      // 0d9: athrow
      // 0da: aload 0
      // 0db: dup
      // 0dc: getfield o.m I
      // 0df: aload 0
      // 0e0: getfield o.n [I
      // 0e3: sipush 255
      // 0e6: sipush 128
      // 0e9: iload 2
      // 0ea: iadd
      // 0eb: iand
      // 0ec: iaload
      // 0ed: iadd
      // 0ee: putfield o.m I
      // 0f1: aload 0
      // 0f2: getfield o.n [I
      // 0f5: iload 2
      // 0f6: aload 0
      // 0f7: getfield o.m I
      // 0fa: aload 0
      // 0fb: getfield o.n [I
      // 0fe: sipush 1020
      // 101: iload 3
      // 102: invokestatic ib.a (II)I
      // 105: ldc -1542190526
      // 107: ishr
      // 108: iaload
      // 109: iadd
      // 10a: aload 0
      // 10b: getfield o.d I
      // 10e: ineg
      // 10f: isub
      // 110: dup
      // 111: istore 4
      // 113: iastore
      // 114: aload 0
      // 115: getfield o.k [I
      // 118: iload 2
      // 119: aload 0
      // 11a: aload 0
      // 11b: getfield o.n [I
      // 11e: sipush 255
      // 121: iload 4
      // 123: ldc 725943080
      // 125: ishr
      // 126: ldc -16506142
      // 128: ishr
      // 129: invokestatic ib.a (II)I
      // 12c: iaload
      // 12d: iload 3
      // 12e: iadd
      // 12f: dup_x1
      // 130: putfield o.d I
      // 133: iastore
      // 134: iinc 2 1
      // 137: iload 6
      // 139: ifeq 034
      // 13c: goto 161
      // 13f: astore 2
      // 140: aload 2
      // 141: new java/lang/StringBuilder
      // 144: dup
      // 145: invokespecial java/lang/StringBuilder.<init> ()V
      // 148: getstatic o.z [Ljava/lang/String;
      // 14b: bipush 6
      // 14d: aaload
      // 14e: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 151: iload 1
      // 152: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 155: bipush 41
      // 157: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 15a: invokevirtual java/lang/StringBuilder.toString ()Ljava/lang/String;
      // 15d: invokestatic i.a (Ljava/lang/Throwable;Ljava/lang/String;)Lla;
      // 160: athrow
      // 161: return
   }

   private final void a(int var1) {
      boolean var11 = client.vh;

      try {
         c++;
         int var5 = -1640531527;
         int var3 = -1640531527;
         int var10 = -1640531527;
         int var9 = -1640531527;
         int var7 = -1640531527;

         try {
            if (var1 != -2) {
               this.a(-15);
            }
         } catch (RuntimeException var15) {
            throw var15;
         }

         int var4 = -1640531527;
         int var8 = -1640531527;
         int var6 = -1640531527;
         int var2 = 0;

         while (true) {
            if (~var2 > -5) {
               var3 ^= var4 << -1205206741;
               int var41 = var6 + var3;
               int var25 = var4 + var5;
               int var26 = var25 ^ var5 >>> -890549438;
               int var49 = var7 + var26;
               var5 += var41;
               var5 ^= var41 << 2101945832;
               int var42 = var41 + var49;
               int var57 = var8 + var5;
               var6 = var42 ^ var49 >>> 581272400;
               int var50 = var49 + var57;
               int var65 = var9 + var6;
               var7 = var50 ^ var57 << -1718537238;
               int var73 = var10 + var7;
               int var58 = var57 + var65;
               var8 = var58 ^ var65 >>> 1908708324;
               int var66 = var65 + var73;
               var3 += var8;
               var9 = var66 ^ var73 << 297382696;
               int var74 = var73 + var3;
               var4 = var26 + var9;
               var10 = var74 ^ var3 >>> 1751843753;
               var3 += var4;
               var5 += var10;

               try {
                  var2++;
                  if (var11) {
                     break;
                  }

                  if (!var11) {
                     continue;
                  }
               } catch (RuntimeException var14) {
                  throw var14;
               }
            }

            var2 = 0;
            break;
         }

         while (true) {
            if (-257 < ~var2) {
               var3 += this.k[var2];
               int var51 = var7 + this.k[4 + var2];
               var5 += this.k[2 + var2];
               int var75 = var10 + this.k[var2 + 7];
               int var59 = var8 + this.k[var2 - -5];
               int var43 = var6 + this.k[3 + var2];
               int var67 = var9 + this.k[6 + var2];
               int var27 = var4 + this.k[var2 + 1];
               var3 ^= var27 << -536269493;
               int var44 = var43 + var3;
               int var28 = var27 + var5;
               int var29 = var28 ^ var5 >>> 186681986;
               var5 += var44;
               int var52 = var51 + var29;
               var5 ^= var44 << -1065768376;
               int var45 = var44 + var52;
               int var60 = var59 + var5;
               var6 = var45 ^ var52 >>> 21869104;
               int var68 = var67 + var6;
               int var53 = var52 + var60;
               var7 = var53 ^ var60 << -1346760726;
               int var76 = var75 + var7;
               int var61 = var60 + var68;
               var8 = var61 ^ var68 >>> -949961308;
               var3 += var8;
               int var69 = var68 + var76;
               var9 = var69 ^ var76 << 1673244968;
               var4 = var29 + var9;
               int var77 = var76 + var3;
               var10 = var77 ^ var3 >>> 409602633;
               var5 += var10;
               var3 += var4;

               try {
                  this.n[var2] = var3;
                  this.n[var2 - -1] = var4;
                  this.n[2 + var2] = var5;
                  this.n[3 + var2] = var6;
                  this.n[4 + var2] = var7;
                  this.n[var2 - -5] = var8;
                  this.n[var2 - -6] = var9;
                  this.n[var2 - -7] = var10;
                  var2 += 8;
                  if (var11) {
                     break;
                  }

                  if (!var11) {
                     continue;
                  }
               } catch (RuntimeException var13) {
                  throw var13;
               }
            }

            var2 = 0;
            break;
         }

         while (true) {
            if (~var2 > -257) {
               int var70 = var9 + this.n[6 + var2];
               int var30 = var4 + this.n[var2 + 1];
               int var78 = var10 + this.n[7 + var2];
               int var54 = var7 + this.n[4 + var2];
               int var62 = var8 + this.n[5 + var2];
               var5 += this.n[var2 - -2];
               int var46 = var6 + this.n[var2 + 3];
               var3 += this.n[var2];
               var3 ^= var30 << -1294322773;
               int var31 = var30 + var5;
               int var47 = var46 + var3;
               int var32 = var31 ^ var5 >>> -117514910;
               var5 += var47;
               int var55 = var54 + var32;
               var5 ^= var47 << -1087924184;
               int var63 = var62 + var5;
               int var48 = var47 + var55;
               var6 = var48 ^ var55 >>> 1052530928;
               int var56 = var55 + var63;
               int var71 = var70 + var6;
               var7 = var56 ^ var63 << -1448878102;
               int var64 = var63 + var71;
               int var79 = var78 + var7;
               var8 = var64 ^ var71 >>> 819293412;
               int var72 = var71 + var79;
               var3 += var8;
               var9 = var72 ^ var79 << -1355920056;
               int var80 = var79 + var3;
               var4 = var32 + var9;
               var10 = var80 ^ var3 >>> -2145912983;
               var3 += var4;
               var5 += var10;

               try {
                  this.n[var2] = var3;
                  this.n[var2 - -1] = var4;
                  this.n[var2 - -2] = var5;
                  this.n[3 + var2] = var6;
                  this.n[var2 + 4] = var7;
                  this.n[var2 + 5] = var8;
                  this.n[var2 - -6] = var9;
                  this.n[7 + var2] = var10;
                  var2 += 8;
                  if (var11) {
                     break;
                  }

                  if (!var11) {
                     continue;
                  }
               } catch (RuntimeException var12) {
                  throw var12;
               }
            }

            this.b(-105);
            this.i = 256;
            break;
         }
      } catch (RuntimeException var16) {
         throw i.a(var16, z[2] + var1 + ')');
      }
   }

   static final int a(int var0, int var1, int var2, int var3) {
      try {
         try {
            if (var1 != 9570) {
               a((byte)56);
            }
         } catch (RuntimeException var5) {
            throw var5;
         }

         e++;
         return (var0 << 905616656) + ((var3 << 81348360) - -var2);
      } catch (RuntimeException var6) {
         throw i.a(var6, z[1] + var0 + 44 + var1 + 44 + var2 + 44 + var3 + 41);
      }
   }

   static final String a(byte var0) {
      boolean var2 = client.vh;

      try {
         try {
            o++;
            if (var0 != 38) {
               a(67, 106, -48, 111);
            }
         } catch (RuntimeException var4) {
            throw var4;
         }

         String var1 = "";

         while (true) {
            if (-1 != ~kb.d[jb.p]) {
               var1 = var1 + (char)kb.d[jb.p++];
               boolean var10000 = var2;

               try {
                  if (var10000) {
                     break;
                  }

                  if (!var2) {
                     continue;
                  }
               } catch (RuntimeException var3) {
                  throw var3;
               }
            }

            jb.p++;
            break;
         }

         return var1;
      } catch (RuntimeException var5) {
         throw i.a(var5, z[0] + var0 + ')');
      }
   }

   static final int a(int param0, int param1, int param2, int param3, boolean param4, int param5, int param6, int param7) {
      // $VF: Couldn't be decompiled
      // Please report this to the Vineflower issue tracker, at https://github.com/Vineflower/vineflower/issues with a copy of the class file (if you have the rights to distribute it!)
      // java.lang.RuntimeException: parsing failure!
      //   at org.jetbrains.java.decompiler.modules.decompiler.decompose.DomHelper.parseGraph(DomHelper.java:211)
      //   at org.jetbrains.java.decompiler.main.rels.MethodProcessor.codeToJava(MethodProcessor.java:166)
      //
      // Bytecode:
      // 000: getstatic client.vh Z
      // 003: istore 12
      // 005: getstatic o.j I
      // 008: bipush 1
      // 009: iadd
      // 00a: putstatic o.j I
      // 00d: bipush 0
      // 00e: istore 8
      // 010: bipush 0
      // 011: istore 9
      // 013: iload 9
      // 015: bipush -1
      // 016: ixor
      // 017: iload 5
      // 019: bipush -1
      // 01a: ixor
      // 01b: if_icmple 088
      // 01e: iload 7
      // 020: iload 1
      // 021: iload 12
      // 023: ifne 08c
      // 026: iload 4
      // 028: ifne 036
      // 02b: goto 02f
      // 02e: athrow
      // 02f: iload 9
      // 031: ineg
      // 032: goto 038
      // 035: athrow
      // 036: iload 9
      // 038: iload 6
      // 03a: isub
      // 03b: iadd
      // 03c: imul
      // 03d: istore 10
      // 03f: iload 10
      // 041: bipush -100
      // 043: if_icmplt 05c
      // 046: bipush -101
      // 048: iload 10
      // 04a: bipush -1
      // 04b: ixor
      // 04c: if_icmple 060
      // 04f: goto 053
      // 052: athrow
      // 053: bipush 100
      // 055: istore 10
      // 057: iload 12
      // 059: ifeq 060
      // 05c: bipush -100
      // 05e: istore 10
      // 060: iload 2
      // 061: iload 10
      // 063: ineg
      // 064: isub
      // 065: istore 11
      // 067: bipush -11
      // 069: iload 11
      // 06b: bipush -1
      // 06c: ixor
      // 06d: if_icmpge 074
      // 070: bipush 10
      // 072: istore 11
      // 074: iload 8
      // 076: iload 0
      // 077: iload 11
      // 079: imul
      // 07a: bipush 100
      // 07c: idiv
      // 07d: iadd
      // 07e: istore 8
      // 080: iinc 9 1
      // 083: iload 12
      // 085: ifeq 013
      // 088: iload 3
      // 089: sipush -30910
      // 08c: if_icmpeq 099
      // 08f: bipush 106
      // 091: invokestatic o.a (B)Ljava/lang/String;
      // 094: pop
      // 095: goto 099
      // 098: athrow
      // 099: iload 8
      // 09b: ireturn
      // 09c: astore 8
      // 09e: aload 8
      // 0a0: new java/lang/StringBuilder
      // 0a3: dup
      // 0a4: invokespecial java/lang/StringBuilder.<init> ()V
      // 0a7: getstatic o.z [Ljava/lang/String;
      // 0aa: bipush 8
      // 0ac: aaload
      // 0ad: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 0b0: iload 0
      // 0b1: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 0b4: bipush 44
      // 0b6: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 0b9: iload 1
      // 0ba: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 0bd: bipush 44
      // 0bf: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 0c2: iload 2
      // 0c3: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 0c6: bipush 44
      // 0c8: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 0cb: iload 3
      // 0cc: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 0cf: bipush 44
      // 0d1: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 0d4: iload 4
      // 0d6: invokevirtual java/lang/StringBuilder.append (Z)Ljava/lang/StringBuilder;
      // 0d9: bipush 44
      // 0db: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 0de: iload 5
      // 0e0: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 0e3: bipush 44
      // 0e5: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 0e8: iload 6
      // 0ea: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 0ed: bipush 44
      // 0ef: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 0f2: iload 7
      // 0f4: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 0f7: bipush 41
      // 0f9: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 0fc: invokevirtual java/lang/StringBuilder.toString ()Ljava/lang/String;
      // 0ff: invokestatic i.a (Ljava/lang/Throwable;Ljava/lang/String;)Lla;
      // 102: athrow
   }

   o(int[] var1) {
      boolean var3 = client.vh;
      super();

      try {
         this.k = new int[256];
         this.n = new int[256];
         int var7 = 0;

         while (true) {
            if (~var1.length < ~var7) {
               try {
                  this.k[var7] = var1[var7];
                  var7++;
                  if (var3) {
                     break;
                  }

                  if (!var3) {
                     continue;
                  }
               } catch (RuntimeException var5) {
                  throw var5;
               }
            }

            this.a(-2);
            break;
         }
      } catch (RuntimeException var6) {
         RuntimeException var2 = var6;

         RuntimeException var10000;
         StringBuilder var10001;
         try {
            var10000 = var2;
            var10001 = new StringBuilder().append(z[5]);
            if (var1 != null) {
               throw i.a(var2, var10001.append(z[3]).append(')').toString());
            }
         } catch (RuntimeException var4) {
            throw var4;
         }

         throw i.a(var10000, var10001.append(z[4]).append(')').toString());
      }
   }

   private static char[] z(String var0) {
      char[] var10000 = var0.toCharArray();
      if (var10000.length < 2) {
         var10000[0] = (char)(var10000[0] ^ 's');
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
               var10005 = 60;
               break;
            case 1:
               var10005 = 75;
               break;
            case 2:
               var10005 = 27;
               break;
            case 3:
               var10005 = 19;
               break;
            default:
               var10005 = 115;
         }

         var10001[var1] = (char)(var10004 ^ var10005);
      }

      return new String(var10001).intern();
   }
}
