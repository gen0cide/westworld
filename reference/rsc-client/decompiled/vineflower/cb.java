import java.io.IOException;
import java.net.URL;

final class cb {
   static int a;
   static int g;
   static String[] c = new String[200];
   static int f;
   static int b;
   static String[] e;
   static int d;
   private static final String[] z = new String[]{
      z(z("j`[\u0018W")),
      z(z("\u007f;\u0019Z")),
      z(z("r,[u\u0002")),
      z(z("r!\u001bBO\u007f:\u0016DIb")),
      z(z("r,[w\u0002")),
      z(z("R&\u0010UAx \u0012\u0016L~<UXOfn\u0016YDe+\u001bB")),
      z(z("X \u0003WFx*UuxRn\u001cX\nR\u001c6\u0016Iy+\u0016]\nw'\u0019S")),
      z(z("r,[s\u0002")),
      z(z("r,[t\u0002")),
      z(z("r,[r\u0002"))
   };

   static final void a(URL var0, e var1, int var2) throws IOException {
      boolean var7 = client.vh;

      try {
         f++;
         d.h = var1;
         ib.c = var0;
         URL var15 = new URL(ib.c, z[3] + Long.toHexString(p.a(0)));
         o.l = z[5];
         byte[] var4 = da.a(var15, true, true);
         tb var5 = new tb(var4);
         int var6 = 0;

         while (true) {
            if (-13 < ~var6) {
               try {
                  tb.l[var6] = var5.b(-129);
                  var6++;
                  if (var7) {
                     break;
                  }

                  if (!var7) {
                     continue;
                  }
               } catch (IOException var13) {
                  throw var13;
               }
            }

            var5.b(-129);
            break;
         }

         try {
            if (!var5.e(-422797528)) {
               throw new IOException(z[6]);
            }
         } catch (IOException var12) {
            throw var12;
         }

         try {
            var6 = 81 % ((0 - var2) / 54);

            try {
               if (pa.k.f != null) {
                  s.a = new nb(pa.k.f, 5200, 0);
                  n.h = new nb(pa.k.v, 6000, 0);
                  m.e = new ob(0, s.a, n.h, 1000000);
                  pa.k.f = null;
                  pa.k.v = null;
               }
            } catch (IOException var8) {
               throw var8;
            }
         } catch (IOException var9) {
            s.a = null;
            n.h = null;
         }
      } catch (RuntimeException var14) {
         RuntimeException var3 = var14;

         RuntimeException var10000;
         StringBuilder var10001;
         String var10002;
         label59: {
            try {
               var10000 = var3;
               var10001 = new StringBuilder().append(z[4]);
               if (var0 != null) {
                  var10002 = z[0];
                  break label59;
               }
            } catch (IOException var11) {
               throw var11;
            }

            var10002 = z[1];
         }

         try {
            var10001 = var10001.append(var10002).append(',');
            if (var1 != null) {
               throw i.a(var10000, var10001.append(z[0]).append(',').append(var2).append(')').toString());
            }
         } catch (IOException var10) {
            throw var10;
         }

         throw i.a(var10000, var10001.append(z[1]).append(',').append(var2).append(')').toString());
      }
   }

   // $VF: Handled exception range with multiple entry points by splitting it
   // $VF: Inserted dummy exception handlers to handle obfuscated exceptions
   static final void a(
      int var0,
      int var1,
      int var2,
      byte var3,
      int var4,
      int var5,
      int var6,
      int var7,
      int[] var8,
      int[] var9,
      int var10,
      int var11,
      int var12,
      int var13,
      int var14,
      int var15
   ) {
      boolean var23 = client.vh;

      RuntimeException var10000;
      label312: {
         try {
            a++;
            if (var0 <= 0) {
               return;
            }
         } catch (RuntimeException var42) {
            var10000 = var42;
            boolean var10001 = false;
            break label312;
         }

         int var16;
         int var17;
         try {
            var16 = 0;
            var17 = 0;

            label280: {
               try {
                  if (0 == var1) {
                     break label280;
                  }
               } catch (RuntimeException var38) {
                  throw var38;
               }

               var16 = var11 / var1 << -1240379354;
               var17 = var15 / var1 << 1095567590;
            }

            var7 <<= 2;

            label291: {
               label290: {
                  try {
                     if (0 > var16) {
                        break label290;
                     }

                     if (var16 <= 4032) {
                        break label291;
                     }
                  } catch (RuntimeException var39) {
                     throw var39;
                  }

                  var16 = 4032;
                  if (!var23) {
                     break label291;
                  }
               }

               var16 = 0;
            }

            if (var3 != 25) {
               return;
            }
         } catch (RuntimeException var41) {
            var10000 = var41;
            boolean var58 = false;
            break label312;
         }

         try {
            int var20 = var0;

            label274:
            while (true) {
               int var47 = ~var20;
               int var62 = -1;

               label272:
               while (var47 < var62) {
                  var4 = var17;
                  var12 = var16;
                  var11 += var5;
                  var1 += var6;
                  var15 += var13;

                  label269: {
                     try {
                        if (var23) {
                           return;
                        }

                        if (var1 == 0) {
                           break label269;
                        }
                     } catch (RuntimeException var37) {
                        throw var37;
                     }

                     var17 = var15 / var1 << -1050475738;
                     var16 = var11 / var1 << -700638746;
                  }

                  label331: {
                     if (~var16 > -1) {
                        var16 = 0;
                        boolean var48 = var23;

                        try {
                           if (!var48) {
                              break label331;
                           }
                        } catch (RuntimeException var36) {
                           boolean var63 = false;
                           throw var36;
                        }
                     }

                     try {
                        if (~var16 >= -4033) {
                           break label331;
                        }
                     } catch (RuntimeException var35) {
                        boolean var64 = false;
                        throw var35;
                     }

                     var16 = 4032;
                  }

                  label316: {
                     int var19 = var17 + -var4 >> -490767996;
                     int var18 = -var12 + var16 >> 461459556;
                     var12 += 786432 & var14;
                     int var21 = var14 >> 1111279860;
                     var14 += var7;
                     if (-17 >= ~var20) {
                        if (0 != (var2 = var8[(var12 >> 1781768774) + (4032 & var4)] >>> var21)) {
                           var9[var10] = var2;
                        }

                        var10++;
                        var4 += var19;
                        var12 += var18;
                        byte var50 = -1;
                        var62 = var2 = var8[(var12 >> 1983493318) + (var4 & 4032)] >>> var21;
                        byte var74 = -1;

                        try {
                           if (var50 != (var62 ^ var74)) {
                              var9[var10] = var2;
                           }
                        } catch (RuntimeException var34) {
                           throw var34;
                        }

                        var4 += var19;
                        var10++;
                        var12 += var18;
                        if (0 != (var2 = var8[(var12 >> 1627062566) + (4032 & var4)] >>> var21)) {
                           var9[var10] = var2;
                        }

                        var4 += var19;
                        var10++;
                        var12 += var18;
                        if (0 != (var2 = var8[(4032 & var4) + (var12 >> 387291942)] >>> var21)) {
                           var9[var10] = var2;
                        }

                        var10++;
                        var12 += var18;
                        var4 += var19;
                        var21 = var14 >> -1634170220;
                        var12 = (786432 & var14) + (4095 & var12);
                        var14 += var7;
                        byte var51 = -1;
                        var62 = var2 = var8[(var12 >> -291291162) + (4032 & var4)] >>> var21;
                        var74 = -1;

                        label234: {
                           try {
                              if (var51 == (var62 ^ var74)) {
                                 break label234;
                              }
                           } catch (RuntimeException var33) {
                              throw var33;
                           }

                           var9[var10] = var2;
                        }

                        var10++;
                        var12 += var18;
                        var4 += var19;
                        if ((var2 = var8[(var4 & 4032) + (var12 >> 1451268166)] >>> var21) != 0) {
                           try {
                              var9[var10] = var2;
                           } catch (RuntimeException var25) {
                              throw var25;
                           }
                        }

                        var10++;
                        var12 += var18;
                        var4 += var19;
                        byte var52 = -1;
                        var62 = var2 = var8[(var4 & 4032) + (var12 >> 258323942)] >>> var21;
                        var74 = -1;

                        label226: {
                           try {
                              if (var52 == (var62 ^ var74)) {
                                 break label226;
                              }
                           } catch (RuntimeException var32) {
                              throw var32;
                           }

                           var9[var10] = var2;
                        }

                        var10++;
                        var12 += var18;
                        var4 += var19;
                        if (0 != (var2 = var8[(var4 & 4032) + (var12 >> -1744130106)] >>> var21)) {
                           var9[var10] = var2;
                        }

                        var4 += var19;
                        var10++;
                        var12 += var18;
                        var12 = (var12 & 4095) + (var14 & 786432);
                        var21 = var14 >> -1353915596;
                        if ((var2 = var8[(var12 >> 1242019238) + (var4 & 4032)] >>> var21) != 0) {
                           var9[var10] = var2;
                        }

                        var14 += var7;
                        var10++;
                        var12 += var18;
                        var4 += var19;
                        if (0 != (var2 = var8[(var12 >> -980660250) + (4032 & var4)] >>> var21)) {
                           var9[var10] = var2;
                        }

                        var4 += var19;
                        var12 += var18;
                        var10++;
                        int var53 = var2 = var8[(var12 >> -1023624666) + (4032 & var4)] >>> var21;
                        int var68 = -1;

                        label215: {
                           try {
                              if ((var53 ^ var68) == -1) {
                                 break label215;
                              }
                           } catch (RuntimeException var31) {
                              throw var31;
                           }

                           var9[var10] = var2;
                        }

                        var12 += var18;
                        var4 += var19;
                        var10++;
                        int var54 = var2 = var8[(var4 & 4032) + (var12 >> -530730842)] >>> var21;
                        var68 = (byte)-1;

                        label209: {
                           try {
                              if ((var54 ^ var68) == -1) {
                                 break label209;
                              }
                           } catch (RuntimeException var30) {
                              throw var30;
                           }

                           var9[var10] = var2;
                        }

                        var10++;
                        var12 += var18;
                        var4 += var19;
                        var21 = var14 >> -352222028;
                        var12 = (var14 & 786432) + (var12 & 4095);
                        var14 += var7;
                        int var55 = var2 = var8[(var4 & 4032) - -(var12 >> -2119184058)] >>> var21;
                        var68 = (byte)-1;

                        try {
                           if ((var55 ^ var68) != -1) {
                              var9[var10] = var2;
                           }
                        } catch (RuntimeException var24) {
                           throw var24;
                        }

                        var4 += var19;
                        var12 += var18;
                        var10++;
                        if (0 != (var2 = var8[(var4 & 4032) - -(var12 >> 1682015462)] >>> var21)) {
                           var9[var10] = var2;
                        }

                        var10++;
                        var12 += var18;
                        var4 += var19;
                        byte var56 = -1;
                        var68 = var2 = var8[(var4 & 4032) + (var12 >> 971517030)] >>> var21;
                        var74 = -1;

                        label200: {
                           try {
                              if (var56 == (var68 ^ var74)) {
                                 break label200;
                              }
                           } catch (RuntimeException var29) {
                              throw var29;
                           }

                           var9[var10] = var2;
                        }

                        var4 += var19;
                        var10++;
                        var12 += var18;
                        byte var57 = -1;
                        var68 = var2 = var8[(4032 & var4) - -(var12 >> -761317434)] >>> var21;
                        var74 = -1;

                        label194: {
                           try {
                              if (var57 == (var68 ^ var74)) {
                                 break label194;
                              }
                           } catch (RuntimeException var28) {
                              throw var28;
                           }

                           var9[var10] = var2;
                        }

                        var10++;
                        if (!var23) {
                           break label316;
                        }
                     }

                     int var22 = 0;

                     while (var20 > var22) {
                        var47 = -1;
                        var62 = ~(var2 = var8[(var12 >> -1996690362) + (4032 & var4)] >>> var21);
                        if (var23) {
                           continue label272;
                        }

                        if (-1 != var62) {
                           var9[var10] = var2;
                        }

                        var10++;
                        var12 += var18;
                        var4 += var19;
                        if (3 == (3 & var22)) {
                           var21 = var14 >> 898327988;
                           var12 = (4095 & var12) + (var14 & 786432);
                           var14 += var7;
                        }

                        var22++;
                        if (var23) {
                           break;
                        }
                     }
                  }

                  var20 -= 16;
                  if (var23) {
                     return;
                  }
                  continue label274;
               }

               return;
            }
         } catch (RuntimeException var40) {
            var10000 = var40;
            boolean var59 = false;
         }
      }

      RuntimeException var43 = var10000;

      StringBuilder var60;
      String var10002;
      label169: {
         try {
            var10000 = var43;
            var60 = new StringBuilder()
               .append(z[8])
               .append(var0)
               .append(',')
               .append(var1)
               .append(',')
               .append(var2)
               .append(',')
               .append((int)var3)
               .append(',')
               .append(var4)
               .append(',')
               .append(var5)
               .append(',')
               .append(var6)
               .append(',')
               .append(var7)
               .append(',');
            if (var8 != null) {
               var10002 = z[0];
               break label169;
            }
         } catch (RuntimeException var27) {
            throw var27;
         }

         var10002 = z[1];
      }

      try {
         var60 = var60.append(var10002).append(',');
         if (var9 != null) {
            throw i.a(
               var10000,
               var60.append(z[0])
                  .append(',')
                  .append(var10)
                  .append(',')
                  .append(var11)
                  .append(',')
                  .append(var12)
                  .append(',')
                  .append(var13)
                  .append(',')
                  .append(var14)
                  .append(',')
                  .append(var15)
                  .append(')')
                  .toString()
            );
         }
      } catch (RuntimeException var26) {
         throw var26;
      }

      throw i.a(
         var10000,
         var60.append(z[1])
            .append(',')
            .append(var10)
            .append(',')
            .append(var11)
            .append(',')
            .append(var12)
            .append(',')
            .append(var13)
            .append(',')
            .append(var14)
            .append(',')
            .append(var15)
            .append(')')
            .toString()
      );
   }

   static final void a(aa var0, byte var1) {
      try {
         fb.a = var0;
         d++;
         int var5 = -87 % ((-31 - var1) / 41);
      } catch (RuntimeException var4) {
         RuntimeException var2 = var4;

         RuntimeException var10000;
         StringBuilder var10001;
         try {
            var10000 = var2;
            var10001 = new StringBuilder().append(z[2]);
            if (var0 != null) {
               throw i.a(var2, var10001.append(z[0]).append(',').append((int)var1).append(')').toString());
            }
         } catch (RuntimeException var3) {
            throw var3;
         }

         throw i.a(var10000, var10001.append(z[1]).append(',').append((int)var1).append(')').toString());
      }
   }

   static final void a(
      int param0,
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
      int param11,
      int param12,
      int[] param13,
      byte param14
   ) {
      // $VF: Couldn't be decompiled
      // Please report this to the Vineflower issue tracker, at https://github.com/Vineflower/vineflower/issues with a copy of the class file (if you have the rights to distribute it!)
      // java.lang.RuntimeException: parsing failure!
      //   at org.jetbrains.java.decompiler.modules.decompiler.decompose.DomHelper.parseGraph(DomHelper.java:211)
      //   at org.jetbrains.java.decompiler.main.rels.MethodProcessor.codeToJava(MethodProcessor.java:166)
      //
      // Bytecode:
      // 000: getstatic client.vh Z
      // 003: istore 21
      // 005: getstatic cb.b I
      // 008: bipush 1
      // 009: iadd
      // 00a: putstatic cb.b I
      // 00d: bipush 0
      // 00e: iload 11
      // 010: if_icmplt 014
      // 013: return
      // 014: bipush 0
      // 015: istore 15
      // 017: bipush 0
      // 018: istore 16
      // 01a: iload 14
      // 01c: bipush 97
      // 01e: if_icmpgt 04c
      // 021: bipush -65
      // 023: bipush -47
      // 025: bipush -42
      // 027: bipush -16
      // 029: bipush 62
      // 02b: bipush 50
      // 02d: bipush -59
      // 02f: bipush -91
      // 031: aconst_null
      // 032: checkcast [I
      // 035: aconst_null
      // 036: checkcast [I
      // 039: bipush 71
      // 03b: bipush -91
      // 03d: bipush -16
      // 03f: bipush -29
      // 041: bipush 110
      // 043: bipush 81
      // 045: invokestatic cb.a (IIIBIIII[I[IIIIIII)V
      // 048: goto 04c
      // 04b: athrow
      // 04c: bipush 0
      // 04d: istore 19
      // 04f: bipush 0
      // 050: iload 7
      // 052: if_icmpeq 066
      // 055: iload 1
      // 056: iload 7
      // 058: idiv
      // 059: ldc -987682713
      // 05b: ishl
      // 05c: istore 3
      // 05d: iload 2
      // 05e: iload 7
      // 060: idiv
      // 061: ldc -1456039769
      // 063: ishl
      // 064: istore 6
      // 066: iload 7
      // 068: iload 12
      // 06a: iadd
      // 06b: istore 7
      // 06d: bipush -1
      // 06e: iload 6
      // 070: bipush -1
      // 071: ixor
      // 072: if_icmplt 08d
      // 075: iload 6
      // 077: bipush -1
      // 078: ixor
      // 079: sipush -16257
      // 07c: if_icmpge 090
      // 07f: goto 083
      // 082: athrow
      // 083: sipush 16256
      // 086: istore 6
      // 088: iload 21
      // 08a: ifeq 090
      // 08d: bipush 0
      // 08e: istore 6
      // 090: iload 1
      // 091: iload 5
      // 093: iadd
      // 094: istore 1
      // 095: iload 2
      // 096: iload 8
      // 098: iadd
      // 099: istore 2
      // 09a: iload 7
      // 09c: bipush -1
      // 09d: ixor
      // 09e: bipush -1
      // 09f: if_icmpne 0a6
      // 0a2: goto 0b8
      // 0a5: athrow
      // 0a6: iload 2
      // 0a7: iload 7
      // 0a9: idiv
      // 0aa: ldc 1406785287
      // 0ac: ishl
      // 0ad: istore 15
      // 0af: iload 1
      // 0b0: iload 7
      // 0b2: idiv
      // 0b3: ldc -129167545
      // 0b5: ishl
      // 0b6: istore 16
      // 0b8: bipush -1
      // 0b9: iload 15
      // 0bb: bipush -1
      // 0bc: ixor
      // 0bd: if_icmplt 0dc
      // 0c0: sipush -16257
      // 0c3: iload 15
      // 0c5: bipush -1
      // 0c6: ixor
      // 0c7: if_icmpgt 0d2
      // 0ca: goto 0ce
      // 0cd: athrow
      // 0ce: goto 0df
      // 0d1: athrow
      // 0d2: sipush 16256
      // 0d5: istore 15
      // 0d7: iload 21
      // 0d9: ifeq 0df
      // 0dc: bipush 0
      // 0dd: istore 15
      // 0df: iload 15
      // 0e1: iload 6
      // 0e3: ineg
      // 0e4: iadd
      // 0e5: ldc -62706076
      // 0e7: ishr
      // 0e8: istore 17
      // 0ea: iload 3
      // 0eb: ineg
      // 0ec: iload 16
      // 0ee: iadd
      // 0ef: ldc 750561924
      // 0f1: ishr
      // 0f2: istore 18
      // 0f4: iload 11
      // 0f6: ldc -1736383548
      // 0f8: ishr
      // 0f9: istore 20
      // 0fb: iload 20
      // 0fd: bipush -1
      // 0fe: ixor
      // 0ff: bipush -1
      // 100: if_icmpge 51a
      // 103: iload 4
      // 105: ldc -719426313
      // 107: ishr
      // 108: istore 19
      // 10a: iload 6
      // 10c: iload 4
      // 10e: ldc 6291456
      // 110: iand
      // 111: iadd
      // 112: istore 6
      // 114: iload 4
      // 116: iload 9
      // 118: iadd
      // 119: istore 4
      // 11b: aload 13
      // 11d: iload 0
      // 11e: iinc 0 1
      // 121: aload 13
      // 123: iload 0
      // 124: iaload
      // 125: ldc -1882151871
      // 127: ishr
      // 128: ldc 8355711
      // 12a: invokestatic ib.a (II)I
      // 12d: aload 10
      // 12f: iload 6
      // 131: ldc -864795129
      // 133: ishr
      // 134: iload 3
      // 135: sipush 16256
      // 138: invokestatic ib.a (II)I
      // 13b: iadd
      // 13c: iaload
      // 13d: iload 19
      // 13f: iushr
      // 140: iadd
      // 141: iastore
      // 142: iload 3
      // 143: iload 18
      // 145: iadd
      // 146: istore 3
      // 147: iload 6
      // 149: iload 17
      // 14b: iadd
      // 14c: istore 6
      // 14e: aload 13
      // 150: iload 0
      // 151: iinc 0 1
      // 154: aload 13
      // 156: iload 0
      // 157: iaload
      // 158: ldc 1149821121
      // 15a: ishr
      // 15b: ldc 8355711
      // 15d: invokestatic ib.a (II)I
      // 160: aload 10
      // 162: iload 6
      // 164: ldc -1630381273
      // 166: ishr
      // 167: sipush 16256
      // 16a: iload 3
      // 16b: invokestatic ib.a (II)I
      // 16e: iadd
      // 16f: iaload
      // 170: iload 19
      // 172: iushr
      // 173: iadd
      // 174: iastore
      // 175: iload 6
      // 177: iload 17
      // 179: iadd
      // 17a: istore 6
      // 17c: iload 3
      // 17d: iload 18
      // 17f: iadd
      // 180: istore 3
      // 181: aload 13
      // 183: iload 0
      // 184: iinc 0 1
      // 187: aload 10
      // 189: sipush 16256
      // 18c: iload 3
      // 18d: invokestatic ib.a (II)I
      // 190: iload 6
      // 192: ldc 1709041255
      // 194: ishr
      // 195: ineg
      // 196: isub
      // 197: iaload
      // 198: iload 19
      // 19a: iushr
      // 19b: ldc 16711422
      // 19d: aload 13
      // 19f: iload 0
      // 1a0: iaload
      // 1a1: invokestatic ib.a (II)I
      // 1a4: ldc -783642367
      // 1a6: ishr
      // 1a7: ineg
      // 1a8: isub
      // 1a9: iastore
      // 1aa: iload 3
      // 1ab: iload 18
      // 1ad: iadd
      // 1ae: istore 3
      // 1af: iload 6
      // 1b1: iload 17
      // 1b3: iadd
      // 1b4: istore 6
      // 1b6: aload 13
      // 1b8: iload 0
      // 1b9: iinc 0 1
      // 1bc: ldc 16711422
      // 1be: aload 13
      // 1c0: iload 0
      // 1c1: iaload
      // 1c2: invokestatic ib.a (II)I
      // 1c5: ldc 2114203393
      // 1c7: ishr
      // 1c8: aload 10
      // 1ca: iload 3
      // 1cb: sipush 16256
      // 1ce: invokestatic ib.a (II)I
      // 1d1: iload 6
      // 1d3: ldc 1285028711
      // 1d5: ishr
      // 1d6: iadd
      // 1d7: iaload
      // 1d8: iload 19
      // 1da: iushr
      // 1db: iadd
      // 1dc: iastore
      // 1dd: iload 3
      // 1de: iload 18
      // 1e0: iadd
      // 1e1: istore 3
      // 1e2: iload 6
      // 1e4: iload 17
      // 1e6: iadd
      // 1e7: istore 6
      // 1e9: iload 4
      // 1eb: ldc 967448183
      // 1ed: ishr
      // 1ee: istore 19
      // 1f0: iload 6
      // 1f2: sipush 16383
      // 1f5: iand
      // 1f6: iload 4
      // 1f8: ldc 6291456
      // 1fa: iand
      // 1fb: iadd
      // 1fc: istore 6
      // 1fe: iload 4
      // 200: iload 9
      // 202: iadd
      // 203: istore 4
      // 205: aload 13
      // 207: iload 0
      // 208: iinc 0 1
      // 20b: aload 13
      // 20d: iload 0
      // 20e: iaload
      // 20f: ldc 237363553
      // 211: ishr
      // 212: ldc 8355711
      // 214: invokestatic ib.a (II)I
      // 217: aload 10
      // 219: sipush 16256
      // 21c: iload 3
      // 21d: invokestatic ib.a (II)I
      // 220: iload 6
      // 222: ldc 1720769863
      // 224: ishr
      // 225: ineg
      // 226: isub
      // 227: iaload
      // 228: iload 19
      // 22a: iushr
      // 22b: iadd
      // 22c: iastore
      // 22d: iload 3
      // 22e: iload 18
      // 230: iadd
      // 231: istore 3
      // 232: iload 6
      // 234: iload 17
      // 236: iadd
      // 237: istore 6
      // 239: aload 13
      // 23b: iload 0
      // 23c: iinc 0 1
      // 23f: aload 10
      // 241: iload 6
      // 243: ldc -1353166233
      // 245: ishr
      // 246: iload 3
      // 247: sipush 16256
      // 24a: invokestatic ib.a (II)I
      // 24d: iadd
      // 24e: iaload
      // 24f: iload 19
      // 251: iushr
      // 252: aload 13
      // 254: iload 0
      // 255: iaload
      // 256: ldc 464826369
      // 258: ishr
      // 259: ldc 8355711
      // 25b: invokestatic ib.a (II)I
      // 25e: ineg
      // 25f: isub
      // 260: iastore
      // 261: iload 3
      // 262: iload 18
      // 264: iadd
      // 265: istore 3
      // 266: iload 6
      // 268: iload 17
      // 26a: iadd
      // 26b: istore 6
      // 26d: aload 13
      // 26f: iload 0
      // 270: iinc 0 1
      // 273: aload 13
      // 275: iload 0
      // 276: iaload
      // 277: ldc 16711423
      // 279: invokestatic ib.a (II)I
      // 27c: ldc 76839841
      // 27e: ishr
      // 27f: aload 10
      // 281: iload 3
      // 282: sipush 16256
      // 285: invokestatic ib.a (II)I
      // 288: iload 6
      // 28a: ldc 2006644519
      // 28c: ishr
      // 28d: ineg
      // 28e: isub
      // 28f: iaload
      // 290: iload 19
      // 292: iushr
      // 293: iadd
      // 294: iastore
      // 295: iload 6
      // 297: iload 17
      // 299: iadd
      // 29a: istore 6
      // 29c: iload 3
      // 29d: iload 18
      // 29f: iadd
      // 2a0: istore 3
      // 2a1: aload 13
      // 2a3: iload 0
      // 2a4: iinc 0 1
      // 2a7: aload 10
      // 2a9: iload 6
      // 2ab: ldc -587801977
      // 2ad: ishr
      // 2ae: sipush 16256
      // 2b1: iload 3
      // 2b2: invokestatic ib.a (II)I
      // 2b5: iadd
      // 2b6: iaload
      // 2b7: iload 19
      // 2b9: iushr
      // 2ba: aload 13
      // 2bc: iload 0
      // 2bd: iaload
      // 2be: ldc 16711423
      // 2c0: invokestatic ib.a (II)I
      // 2c3: ldc -1787059871
      // 2c5: ishr
      // 2c6: ineg
      // 2c7: isub
      // 2c8: iastore
      // 2c9: iload 6
      // 2cb: iload 17
      // 2cd: iadd
      // 2ce: istore 6
      // 2d0: iload 3
      // 2d1: iload 18
      // 2d3: iadd
      // 2d4: istore 3
      // 2d5: sipush 16383
      // 2d8: iload 6
      // 2da: iand
      // 2db: iload 4
      // 2dd: ldc 6291456
      // 2df: iand
      // 2e0: iadd
      // 2e1: istore 6
      // 2e3: iload 4
      // 2e5: ldc 1022075575
      // 2e7: ishr
      // 2e8: istore 19
      // 2ea: aload 13
      // 2ec: iload 0
      // 2ed: iinc 0 1
      // 2f0: ldc 16711423
      // 2f2: aload 13
      // 2f4: iload 0
      // 2f5: iaload
      // 2f6: invokestatic ib.a (II)I
      // 2f9: ldc 263459617
      // 2fb: ishr
      // 2fc: aload 10
      // 2fe: iload 6
      // 300: ldc -1125486105
      // 302: ishr
      // 303: iload 3
      // 304: sipush 16256
      // 307: invokestatic ib.a (II)I
      // 30a: iadd
      // 30b: iaload
      // 30c: iload 19
      // 30e: iushr
      // 30f: iadd
      // 310: iastore
      // 311: iload 4
      // 313: iload 9
      // 315: iadd
      // 316: istore 4
      // 318: iload 3
      // 319: iload 18
      // 31b: iadd
      // 31c: istore 3
      // 31d: iload 6
      // 31f: iload 17
      // 321: iadd
      // 322: istore 6
      // 324: aload 13
      // 326: iload 0
      // 327: iinc 0 1
      // 32a: aload 13
      // 32c: iload 0
      // 32d: iaload
      // 32e: ldc 254571777
      // 330: ishr
      // 331: ldc 8355711
      // 333: invokestatic ib.a (II)I
      // 336: aload 10
      // 338: iload 6
      // 33a: ldc 201079751
      // 33c: ishr
      // 33d: sipush 16256
      // 340: iload 3
      // 341: invokestatic ib.a (II)I
      // 344: iadd
      // 345: iaload
      // 346: iload 19
      // 348: iushr
      // 349: iadd
      // 34a: iastore
      // 34b: iload 6
      // 34d: iload 17
      // 34f: iadd
      // 350: istore 6
      // 352: iload 3
      // 353: iload 18
      // 355: iadd
      // 356: istore 3
      // 357: aload 13
      // 359: iload 0
      // 35a: iinc 0 1
      // 35d: aload 10
      // 35f: iload 3
      // 360: sipush 16256
      // 363: invokestatic ib.a (II)I
      // 366: iload 6
      // 368: ldc 1856596775
      // 36a: ishr
      // 36b: ineg
      // 36c: isub
      // 36d: iaload
      // 36e: iload 19
      // 370: iushr
      // 371: ldc 8355711
      // 373: aload 13
      // 375: iload 0
      // 376: iaload
      // 377: ldc -2129743583
      // 379: ishr
      // 37a: invokestatic ib.a (II)I
      // 37d: iadd
      // 37e: iastore
      // 37f: iload 6
      // 381: iload 17
      // 383: iadd
      // 384: istore 6
      // 386: iload 3
      // 387: iload 18
      // 389: iadd
      // 38a: istore 3
      // 38b: aload 13
      // 38d: iload 0
      // 38e: iinc 0 1
      // 391: ldc 16711423
      // 393: aload 13
      // 395: iload 0
      // 396: iaload
      // 397: invokestatic ib.a (II)I
      // 39a: ldc 345902369
      // 39c: ishr
      // 39d: aload 10
      // 39f: sipush 16256
      // 3a2: iload 3
      // 3a3: invokestatic ib.a (II)I
      // 3a6: iload 6
      // 3a8: ldc 1601471175
      // 3aa: ishr
      // 3ab: ineg
      // 3ac: isub
      // 3ad: iaload
      // 3ae: iload 19
      // 3b0: iushr
      // 3b1: iadd
      // 3b2: iastore
      // 3b3: iload 3
      // 3b4: iload 18
      // 3b6: iadd
      // 3b7: istore 3
      // 3b8: iload 6
      // 3ba: iload 17
      // 3bc: iadd
      // 3bd: istore 6
      // 3bf: iload 6
      // 3c1: sipush 16383
      // 3c4: iand
      // 3c5: iload 4
      // 3c7: ldc 6291456
      // 3c9: iand
      // 3ca: iadd
      // 3cb: istore 6
      // 3cd: iload 4
      // 3cf: ldc -1943261385
      // 3d1: ishr
      // 3d2: istore 19
      // 3d4: aload 13
      // 3d6: iload 0
      // 3d7: iinc 0 1
      // 3da: ldc 8355711
      // 3dc: aload 13
      // 3de: iload 0
      // 3df: iaload
      // 3e0: ldc 1923875073
      // 3e2: ishr
      // 3e3: invokestatic ib.a (II)I
      // 3e6: aload 10
      // 3e8: iload 6
      // 3ea: ldc 965333095
      // 3ec: ishr
      // 3ed: iload 3
      // 3ee: sipush 16256
      // 3f1: invokestatic ib.a (II)I
      // 3f4: iadd
      // 3f5: iaload
      // 3f6: iload 19
      // 3f8: iushr
      // 3f9: iadd
      // 3fa: iastore
      // 3fb: iload 4
      // 3fd: iload 9
      // 3ff: iadd
      // 400: istore 4
      // 402: iload 6
      // 404: iload 17
      // 406: iadd
      // 407: istore 6
      // 409: iload 3
      // 40a: iload 18
      // 40c: iadd
      // 40d: istore 3
      // 40e: aload 13
      // 410: iload 0
      // 411: iinc 0 1
      // 414: aload 13
      // 416: iload 0
      // 417: iaload
      // 418: ldc 481724481
      // 41a: ishr
      // 41b: ldc 8355711
      // 41d: invokestatic ib.a (II)I
      // 420: aload 10
      // 422: iload 6
      // 424: ldc 1596705383
      // 426: ishr
      // 427: sipush 16256
      // 42a: iload 3
      // 42b: invokestatic ib.a (II)I
      // 42e: iadd
      // 42f: iaload
      // 430: iload 19
      // 432: iushr
      // 433: iadd
      // 434: iastore
      // 435: iload 6
      // 437: iload 17
      // 439: iadd
      // 43a: istore 6
      // 43c: iload 3
      // 43d: iload 18
      // 43f: iadd
      // 440: istore 3
      // 441: aload 13
      // 443: iload 0
      // 444: iinc 0 1
      // 447: aload 10
      // 449: iload 3
      // 44a: sipush 16256
      // 44d: invokestatic ib.a (II)I
      // 450: iload 6
      // 452: ldc 1419911623
      // 454: ishr
      // 455: ineg
      // 456: isub
      // 457: iaload
      // 458: iload 19
      // 45a: iushr
      // 45b: aload 13
      // 45d: iload 0
      // 45e: iaload
      // 45f: ldc 905849729
      // 461: ishr
      // 462: ldc 8355711
      // 464: invokestatic ib.a (II)I
      // 467: iadd
      // 468: iastore
      // 469: iload 6
      // 46b: iload 17
      // 46d: iadd
      // 46e: istore 6
      // 470: iload 3
      // 471: iload 18
      // 473: iadd
      // 474: istore 3
      // 475: aload 13
      // 477: iload 0
      // 478: iinc 0 1
      // 47b: aload 13
      // 47d: iload 0
      // 47e: iaload
      // 47f: ldc 1680585121
      // 481: ishr
      // 482: ldc 8355711
      // 484: invokestatic ib.a (II)I
      // 487: aload 10
      // 489: iload 6
      // 48b: ldc 1937899527
      // 48d: ishr
      // 48e: sipush 16256
      // 491: iload 3
      // 492: invokestatic ib.a (II)I
      // 495: iadd
      // 496: iaload
      // 497: iload 19
      // 499: iushr
      // 49a: iadd
      // 49b: iastore
      // 49c: iload 7
      // 49e: iload 12
      // 4a0: iadd
      // 4a1: istore 7
      // 4a3: iload 1
      // 4a4: iload 5
      // 4a6: iadd
      // 4a7: istore 1
      // 4a8: iload 2
      // 4a9: iload 8
      // 4ab: iadd
      // 4ac: istore 2
      // 4ad: iload 16
      // 4af: istore 3
      // 4b0: iload 15
      // 4b2: istore 6
      // 4b4: bipush -1
      // 4b5: iload 7
      // 4b7: bipush -1
      // 4b8: ixor
      // 4b9: iload 21
      // 4bb: ifne 528
      // 4be: if_icmpne 4c8
      // 4c1: goto 4c5
      // 4c4: athrow
      // 4c5: goto 4da
      // 4c8: iload 1
      // 4c9: iload 7
      // 4cb: idiv
      // 4cc: ldc 1649606631
      // 4ce: ishl
      // 4cf: istore 16
      // 4d1: iload 2
      // 4d2: iload 7
      // 4d4: idiv
      // 4d5: ldc 1163846599
      // 4d7: ishl
      // 4d8: istore 15
      // 4da: iload 15
      // 4dc: iflt 4fb
      // 4df: sipush -16257
      // 4e2: iload 15
      // 4e4: bipush -1
      // 4e5: ixor
      // 4e6: if_icmpgt 4f1
      // 4e9: goto 4ed
      // 4ec: athrow
      // 4ed: goto 4fe
      // 4f0: athrow
      // 4f1: sipush 16256
      // 4f4: istore 15
      // 4f6: iload 21
      // 4f8: ifeq 4fe
      // 4fb: bipush 0
      // 4fc: istore 15
      // 4fe: iload 16
      // 500: iload 3
      // 501: isub
      // 502: ldc 915302724
      // 504: ishr
      // 505: istore 18
      // 507: iload 6
      // 509: ineg
      // 50a: iload 15
      // 50c: iadd
      // 50d: ldc 1435337316
      // 50f: ishr
      // 510: istore 17
      // 512: iinc 20 -1
      // 515: iload 21
      // 517: ifeq 0fb
      // 51a: bipush 0
      // 51b: istore 20
      // 51d: iload 20
      // 51f: bipush -1
      // 520: ixor
      // 521: iload 11
      // 523: bipush 15
      // 525: iand
      // 526: bipush -1
      // 527: ixor
      // 528: if_icmple 59a
      // 52b: iload 21
      // 52d: ifne 66c
      // 530: goto 534
      // 533: athrow
      // 534: bipush -1
      // 535: iload 20
      // 537: bipush 3
      // 538: iand
      // 539: bipush -1
      // 53a: ixor
      // 53b: if_icmpne 55e
      // 53e: goto 542
      // 541: athrow
      // 542: iload 4
      // 544: ldc 6291456
      // 546: iand
      // 547: iload 6
      // 549: sipush 16383
      // 54c: iand
      // 54d: iadd
      // 54e: istore 6
      // 550: iload 4
      // 552: ldc -1240422601
      // 554: ishr
      // 555: istore 19
      // 557: iload 4
      // 559: iload 9
      // 55b: iadd
      // 55c: istore 4
      // 55e: aload 13
      // 560: iload 0
      // 561: iinc 0 1
      // 564: aload 10
      // 566: iload 3
      // 567: sipush 16256
      // 56a: invokestatic ib.a (II)I
      // 56d: iload 6
      // 56f: ldc 1938838919
      // 571: ishr
      // 572: ineg
      // 573: isub
      // 574: iaload
      // 575: iload 19
      // 577: iushr
      // 578: aload 13
      // 57a: iload 0
      // 57b: iaload
      // 57c: ldc 16711422
      // 57e: invokestatic ib.a (II)I
      // 581: ldc -616036735
      // 583: ishr
      // 584: iadd
      // 585: iastore
      // 586: iload 6
      // 588: iload 17
      // 58a: iadd
      // 58b: istore 6
      // 58d: iload 3
      // 58e: iload 18
      // 590: iadd
      // 591: istore 3
      // 592: iinc 20 1
      // 595: iload 21
      // 597: ifeq 51d
      // 59a: goto 66c
      // 59d: astore 15
      // 59f: aload 15
      // 5a1: new java/lang/StringBuilder
      // 5a4: dup
      // 5a5: invokespecial java/lang/StringBuilder.<init> ()V
      // 5a8: getstatic cb.z [Ljava/lang/String;
      // 5ab: bipush 7
      // 5ad: aaload
      // 5ae: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 5b1: iload 0
      // 5b2: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 5b5: bipush 44
      // 5b7: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 5ba: iload 1
      // 5bb: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 5be: bipush 44
      // 5c0: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 5c3: iload 2
      // 5c4: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 5c7: bipush 44
      // 5c9: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 5cc: iload 3
      // 5cd: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 5d0: bipush 44
      // 5d2: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 5d5: iload 4
      // 5d7: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 5da: bipush 44
      // 5dc: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 5df: iload 5
      // 5e1: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 5e4: bipush 44
      // 5e6: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 5e9: iload 6
      // 5eb: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 5ee: bipush 44
      // 5f0: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 5f3: iload 7
      // 5f5: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 5f8: bipush 44
      // 5fa: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 5fd: iload 8
      // 5ff: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 602: bipush 44
      // 604: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 607: iload 9
      // 609: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 60c: bipush 44
      // 60e: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 611: aload 10
      // 613: ifnull 61f
      // 616: getstatic cb.z [Ljava/lang/String;
      // 619: bipush 0
      // 61a: aaload
      // 61b: goto 624
      // 61e: athrow
      // 61f: getstatic cb.z [Ljava/lang/String;
      // 622: bipush 1
      // 623: aaload
      // 624: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 627: bipush 44
      // 629: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 62c: iload 11
      // 62e: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 631: bipush 44
      // 633: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 636: iload 12
      // 638: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 63b: bipush 44
      // 63d: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 640: aload 13
      // 642: ifnull 64e
      // 645: getstatic cb.z [Ljava/lang/String;
      // 648: bipush 0
      // 649: aaload
      // 64a: goto 653
      // 64d: athrow
      // 64e: getstatic cb.z [Ljava/lang/String;
      // 651: bipush 1
      // 652: aaload
      // 653: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 656: bipush 44
      // 658: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 65b: iload 14
      // 65d: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 660: bipush 41
      // 662: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 665: invokevirtual java/lang/StringBuilder.toString ()Ljava/lang/String;
      // 668: invokestatic i.a (Ljava/lang/Throwable;Ljava/lang/String;)Lla;
      // 66b: athrow
      // 66c: return
   }

   static final void a(byte var0, Object[] var1, int[] var2) {
      try {
         if (var0 == -70) {
            ub.a(var2, (byte)-128, 0, var2.length + -1, var1);
            g++;
         }
      } catch (RuntimeException var6) {
         RuntimeException var3 = var6;

         RuntimeException var10000;
         StringBuilder var10001;
         String var10002;
         label38: {
            try {
               var10000 = var3;
               var10001 = new StringBuilder().append(z[9]).append((int)var0).append(',');
               if (var1 != null) {
                  var10002 = z[0];
                  break label38;
               }
            } catch (RuntimeException var5) {
               throw var5;
            }

            var10002 = z[1];
         }

         try {
            var10001 = var10001.append(var10002).append(',');
            if (var2 != null) {
               throw i.a(var10000, var10001.append(z[0]).append(')').toString());
            }
         } catch (RuntimeException var4) {
            throw var4;
         }

         throw i.a(var10000, var10001.append(z[1]).append(')').toString());
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
               var10005 = 17;
               break;
            case 1:
               var10005 = 78;
               break;
            case 2:
               var10005 = 117;
               break;
            case 3:
               var10005 = 54;
               break;
            default:
               var10005 = 42;
         }

         var10001[var1] = (char)(var10004 ^ var10005);
      }

      return new String(var10001).intern();
   }
}
