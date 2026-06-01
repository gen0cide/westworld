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
         URL var3 = new URL(ib.c, z[3] + Long.toHexString(p.a(0)));
         o.l = z[5];
         byte[] var4 = da.a(var3, true, true);
         tb var5 = new tb(var4);
         int var6 = 0;

         while (true) {
            if (-13 < ~var6) {
               tb.l[var6] = var5.b(-129);
               var6++;
               if (var7) {
                  break;
               }

               if (!var7) {
                  continue;
               }
            }

            var5.b(-129);
            break;
         }

         if (!var5.e(-422797528)) {
            throw new IOException(z[6]);
         }

         try {
            var6 = 81 % ((0 - var2) / 54);
            if (pa.k.f != null) {
               s.a = new nb(pa.k.f, 5200, 0);
               n.h = new nb(pa.k.v, 6000, 0);
               m.e = new ob(0, s.a, n.h, 1000000);
               pa.k.f = null;
               pa.k.v = null;
            }
         } catch (IOException var8) {
            s.a = null;
            n.h = null;
         }
      } catch (RuntimeException var9) {
         throw i.a(var9, z[4] + (var0 != null ? z[0] : z[1]) + ',' + (var1 != null ? z[0] : z[1]) + ',' + var2 + ')');
      }
   }

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

      try {
         a++;
         if (var0 > 0) {
            int var16 = 0;
            int var17 = 0;
            if (0 != var1) {
               var16 = var11 / var1 << 6;
               var17 = var15 / var1 << 6;
            }

            label197: {
               var7 <<= 2;
               if (0 <= var16) {
                  if (var16 <= 4032) {
                     break label197;
                  }

                  var16 = 4032;
                  if (!var23) {
                     break label197;
                  }
               }

               var16 = 0;
            }

            if (var3 == 25) {
               int var20 = var0;

               label189:
               while (true) {
                  int var10000 = ~var20;
                  int var10001 = -1;

                  label187:
                  while (var10000 < var10001) {
                     var4 = var17;
                     var12 = var16;
                     var11 += var5;
                     var1 += var6;
                     var15 += var13;
                     if (var23) {
                        return;
                     }

                     if (var1 != 0) {
                        var17 = var15 / var1 << 6;
                        var16 = var11 / var1 << 6;
                     }

                     label184: {
                        if (~var16 > -1) {
                           var16 = 0;
                           if (!var23) {
                              break label184;
                           }
                        }

                        if (~var16 < -4033) {
                           var16 = 4032;
                        }
                     }

                     label206: {
                        int var19 = var17 + -var4 >> 4;
                        int var18 = -var12 + var16 >> 4;
                        var12 += 786432 & var14;
                        int var21 = var14 >> 20;
                        var14 += var7;
                        if (-17 >= ~var20) {
                           if (0 != (var2 = var8[(var12 >> 6) + (4032 & var4)] >>> var21)) {
                              var9[var10] = var2;
                           }

                           var10++;
                           var4 += var19;
                           var12 += var18;
                           if (-1 != ~(var2 = var8[(var12 >> 6) + (var4 & 4032)] >>> var21)) {
                              var9[var10] = var2;
                           }

                           var4 += var19;
                           var10++;
                           var12 += var18;
                           if (0 != (var2 = var8[(var12 >> 6) + (4032 & var4)] >>> var21)) {
                              var9[var10] = var2;
                           }

                           var4 += var19;
                           var10++;
                           var12 += var18;
                           if (0 != (var2 = var8[(4032 & var4) + (var12 >> 6)] >>> var21)) {
                              var9[var10] = var2;
                           }

                           var10++;
                           var12 += var18;
                           var4 += var19;
                           var21 = var14 >> 20;
                           var12 = (786432 & var14) + (4095 & var12);
                           var14 += var7;
                           if (-1 != ~(var2 = var8[(var12 >> 6) + (4032 & var4)] >>> var21)) {
                              var9[var10] = var2;
                           }

                           var10++;
                           var12 += var18;
                           var4 += var19;
                           if ((var2 = var8[(var4 & 4032) + (var12 >> 6)] >>> var21) != 0) {
                              var9[var10] = var2;
                           }

                           var10++;
                           var12 += var18;
                           var4 += var19;
                           if (-1 != ~(var2 = var8[(var4 & 4032) + (var12 >> 6)] >>> var21)) {
                              var9[var10] = var2;
                           }

                           var10++;
                           var12 += var18;
                           var4 += var19;
                           if (0 != (var2 = var8[(var4 & 4032) + (var12 >> 6)] >>> var21)) {
                              var9[var10] = var2;
                           }

                           var4 += var19;
                           var10++;
                           var12 += var18;
                           var12 = (var12 & 4095) + (var14 & 786432);
                           var21 = var14 >> 20;
                           if ((var2 = var8[(var12 >> 6) + (var4 & 4032)] >>> var21) != 0) {
                              var9[var10] = var2;
                           }

                           var14 += var7;
                           var10++;
                           var12 += var18;
                           var4 += var19;
                           if (0 != (var2 = var8[(var12 >> 6) + (4032 & var4)] >>> var21)) {
                              var9[var10] = var2;
                           }

                           var4 += var19;
                           var12 += var18;
                           var10++;
                           if (~(var2 = var8[(var12 >> 6) + (4032 & var4)] >>> var21) != -1) {
                              var9[var10] = var2;
                           }

                           var12 += var18;
                           var4 += var19;
                           var10++;
                           if (~(var2 = var8[(var4 & 4032) + (var12 >> 6)] >>> var21) != -1) {
                              var9[var10] = var2;
                           }

                           var10++;
                           var12 += var18;
                           var4 += var19;
                           var21 = var14 >> 20;
                           var12 = (var14 & 786432) + (var12 & 4095);
                           var14 += var7;
                           if (~(var2 = var8[(var4 & 4032) - -(var12 >> 6)] >>> var21) != -1) {
                              var9[var10] = var2;
                           }

                           var4 += var19;
                           var12 += var18;
                           var10++;
                           if (0 != (var2 = var8[(var4 & 4032) - -(var12 >> 6)] >>> var21)) {
                              var9[var10] = var2;
                           }

                           var10++;
                           var12 += var18;
                           var4 += var19;
                           if (-1 != ~(var2 = var8[(var4 & 4032) + (var12 >> 6)] >>> var21)) {
                              var9[var10] = var2;
                           }

                           var4 += var19;
                           var10++;
                           var12 += var18;
                           if (-1 != ~(var2 = var8[(4032 & var4) - -(var12 >> 6)] >>> var21)) {
                              var9[var10] = var2;
                           }

                           var10++;
                           if (!var23) {
                              break label206;
                           }
                        }

                        int var22 = 0;

                        while (var20 > var22) {
                           var10000 = -1;
                           var10001 = ~(var2 = var8[(var12 >> 6) + (4032 & var4)] >>> var21);
                           if (var23) {
                              continue label187;
                           }

                           if (-1 != var10001) {
                              var9[var10] = var2;
                           }

                           var10++;
                           var12 += var18;
                           var4 += var19;
                           if (3 == (3 & var22)) {
                              var21 = var14 >> 20;
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
                     continue label189;
                  }

                  return;
               }
            }
         }
      } catch (RuntimeException var24) {
         throw i.a(
            var24,
            z[8]
               + var0
               + ','
               + var1
               + ','
               + var2
               + ','
               + var3
               + ','
               + var4
               + ','
               + var5
               + ','
               + var6
               + ','
               + var7
               + ','
               + (var8 != null ? z[0] : z[1])
               + ','
               + (var9 != null ? z[0] : z[1])
               + ','
               + var10
               + ','
               + var11
               + ','
               + var12
               + ','
               + var13
               + ','
               + var14
               + ','
               + var15
               + ')'
         );
      }
   }

   static final void a(aa var0, byte var1) {
      try {
         fb.a = var0;
         d++;
         int var2 = -87 % ((-31 - var1) / 41);
      } catch (RuntimeException var3) {
         throw i.a(var3, z[2] + (var0 != null ? z[0] : z[1]) + ',' + var1 + ')');
      }
   }

   // $VF: Irreducible bytecode was duplicated to produce valid code
   static final void a(
      int var0,
      int var1,
      int var2,
      int var3,
      int var4,
      int var5,
      int var6,
      int var7,
      int var8,
      int var9,
      int[] var10,
      int var11,
      int var12,
      int[] var13,
      byte var14
   ) {
      boolean var21 = client.vh;

      try {
         b++;
         if (0 < var11) {
            int var15 = 0;
            int var16 = 0;
            if (var14 <= 97) {
               a(-65, -47, -42, (byte)-16, 62, 50, -59, -91, (int[])null, (int[])null, 71, -91, -16, -29, 110, 81);
            }

            int var19 = 0;
            if (0 != var7) {
               var3 = var1 / var7 << 7;
               var6 = var2 / var7 << 7;
            }

            label126: {
               var7 += var12;
               if (-1 >= ~var6) {
                  if (~var6 >= -16257) {
                     break label126;
                  }

                  var6 = 16256;
                  if (!var21) {
                     break label126;
                  }
               }

               var6 = 0;
            }

            var1 += var5;
            var2 += var8;
            if (~var7 != -1) {
               var15 = var2 / var7 << 7;
               var16 = var1 / var7 << 7;
            }

            label118: {
               if (-1 >= ~var15) {
                  if (-16257 <= ~var15) {
                     break label118;
                  }

                  var15 = 16256;
                  if (!var21) {
                     break label118;
                  }
               }

               var15 = 0;
            }

            int var17 = var15 + -var6 >> 4;
            int var18 = -var3 + var16 >> 4;
            int var20 = var11 >> 4;

            int var41;
            int var10000;
            while (true) {
               if (~var20 < -1) {
                  int var23 = var4 >> 23;
                  var6 += var4 & 6291456;
                  var4 += var9;
                  var13[var0++] = ib.a(var13[var0] >> 1, 8355711) + (var10[(var6 >> 7) + ib.a(var3, 16256)] >>> var23);
                  var3 += var18;
                  var6 += var17;
                  var13[var0++] = ib.a(var13[var0] >> 1, 8355711) + (var10[(var6 >> 7) + ib.a(16256, var3)] >>> var23);
                  var6 += var17;
                  var3 += var18;
                  var13[var0++] = (var10[ib.a(16256, var3) - -(var6 >> 7)] >>> var23) - -(ib.a(16711422, var13[var0]) >> 1);
                  var3 += var18;
                  var6 += var17;
                  var13[var0++] = (ib.a(16711422, var13[var0]) >> 1) + (var10[ib.a(var3, 16256) + (var6 >> 7)] >>> var23);
                  var3 += var18;
                  var6 += var17;
                  int var24 = var4 >> 23;
                  var6 = (var6 & 16383) + (var4 & 6291456);
                  var4 += var9;
                  var13[var0++] = ib.a(var13[var0] >> 1, 8355711) + (var10[ib.a(16256, var3) - -(var6 >> 7)] >>> var24);
                  var3 += var18;
                  var6 += var17;
                  var13[var0++] = (var10[(var6 >> 7) + ib.a(var3, 16256)] >>> var24) - -ib.a(var13[var0] >> 1, 8355711);
                  var3 += var18;
                  var6 += var17;
                  var13[var0++] = (ib.a(var13[var0], 16711423) >> 1) + (var10[ib.a(var3, 16256) - -(var6 >> 7)] >>> var24);
                  var6 += var17;
                  var3 += var18;
                  var13[var0++] = (var10[(var6 >> 7) + ib.a(16256, var3)] >>> var24) - -(ib.a(var13[var0], 16711423) >> 1);
                  var6 += var17;
                  var3 += var18;
                  var6 = (16383 & var6) + (var4 & 6291456);
                  int var25 = var4 >> 23;
                  var13[var0++] = (ib.a(16711423, var13[var0]) >> 1) + (var10[(var6 >> 7) + ib.a(var3, 16256)] >>> var25);
                  var4 += var9;
                  var3 += var18;
                  var6 += var17;
                  var13[var0++] = ib.a(var13[var0] >> 1, 8355711) + (var10[(var6 >> 7) + ib.a(16256, var3)] >>> var25);
                  var6 += var17;
                  var3 += var18;
                  var13[var0++] = (var10[ib.a(var3, 16256) - -(var6 >> 7)] >>> var25) + ib.a(8355711, var13[var0] >> 1);
                  var6 += var17;
                  var3 += var18;
                  var13[var0++] = (ib.a(16711423, var13[var0]) >> 1) + (var10[ib.a(16256, var3) - -(var6 >> 7)] >>> var25);
                  var3 += var18;
                  var6 += var17;
                  var6 = (var6 & 16383) + (var4 & 6291456);
                  var19 = var4 >> 23;
                  var13[var0++] = ib.a(8355711, var13[var0] >> 1) + (var10[(var6 >> 7) + ib.a(var3, 16256)] >>> var19);
                  var4 += var9;
                  var6 += var17;
                  var3 += var18;
                  var13[var0++] = ib.a(var13[var0] >> 1, 8355711) + (var10[(var6 >> 7) + ib.a(16256, var3)] >>> var19);
                  var6 += var17;
                  var3 += var18;
                  var13[var0++] = (var10[ib.a(var3, 16256) - -(var6 >> 7)] >>> var19) + ib.a(var13[var0] >> 1, 8355711);
                  var6 += var17;
                  var3 += var18;
                  var13[var0++] = ib.a(var13[var0] >> 1, 8355711) + (var10[(var6 >> 7) + ib.a(16256, var3)] >>> var19);
                  var7 += var12;
                  var1 += var5;
                  var2 += var8;
                  var3 = var16;
                  var6 = var15;
                  var10000 = -1;
                  var41 = ~var7;
                  if (var21) {
                     break;
                  }

                  if (-1 != var41) {
                     var16 = var1 / var7 << 7;
                     var15 = var2 / var7 << 7;
                  }

                  label98: {
                     if (var15 >= 0) {
                        if (-16257 <= ~var15) {
                           break label98;
                        }

                        var15 = 16256;
                        if (!var21) {
                           break label98;
                        }
                     }

                     var15 = 0;
                  }

                  var18 = var16 - var3 >> 4;
                  var17 = -var6 + var15 >> 4;
                  var20--;
                  if (!var21) {
                     continue;
                  }

                  var20 = 0;
               } else {
                  var20 = 0;
               }

               var10000 = ~var20;
               var41 = ~(var11 & 15);
               break;
            }

            while (var10000 > var41 && !var21) {
               if (-1 == ~(var20 & 3)) {
                  var6 = (var4 & 6291456) + (var6 & 16383);
                  var19 = var4 >> 23;
                  var4 += var9;
               }

               var13[var0++] = (var10[ib.a(var3, 16256) - -(var6 >> 7)] >>> var19) + (ib.a(var13[var0], 16711422) >> 1);
               var6 += var17;
               var3 += var18;
               var20++;
               if (var21) {
                  break;
               }

               var10000 = ~var20;
               var41 = ~(var11 & 15);
            }
         }
      } catch (RuntimeException var22) {
         throw i.a(
            var22,
            z[7]
               + var0
               + ','
               + var1
               + ','
               + var2
               + ','
               + var3
               + ','
               + var4
               + ','
               + var5
               + ','
               + var6
               + ','
               + var7
               + ','
               + var8
               + ','
               + var9
               + ','
               + (var10 != null ? z[0] : z[1])
               + ','
               + var11
               + ','
               + var12
               + ','
               + (var13 != null ? z[0] : z[1])
               + ','
               + var14
               + ')'
         );
      }
   }

   static final void a(byte var0, Object[] var1, int[] var2) {
      try {
         if (var0 == -70) {
            ub.a(var2, (byte)-128, 0, var2.length + -1, var1);
            g++;
         }
      } catch (RuntimeException var4) {
         throw i.a(var4, z[9] + var0 + ',' + (var1 != null ? z[0] : z[1]) + ',' + (var2 != null ? z[0] : z[1]) + ')');
      }
   }

   private static char[] z(String var0) {
      char[] var10000 = var0.toCharArray();
      if (var10000.length < 2) {
         var10000[0] = (char)(var10000[0] ^ 42);
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
