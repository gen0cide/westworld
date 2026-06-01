final class ea {
   private static ac a = new ac();

   private static final void b(ac var0) {
      boolean var4 = false;
      int var5 = 0;
      boolean var6 = false;
      int var7 = 0;
      int var8 = 0;
      int var9 = 0;
      int var10 = 0;
      int var11 = 0;
      int var12 = 0;
      int var13 = 0;
      int var14 = 0;
      int var15 = 0;
      byte var16 = 0;
      int var17 = 0;
      int var18 = 0;
      int var19 = 0;
      byte var20 = 0;
      byte var21 = 0;
      int var22 = 0;
      int[] var23 = null;
      int[] var24 = null;
      int[] var25 = null;
      var0.f = 1;
      if (ua.Mb == null) {
         ua.Mb = new int[var0.f * 100000];
      }

      boolean var26 = true;

      while (var26) {
         byte var1 = c(var0);
         if (var1 == 23) {
            return;
         }

         var1 = c(var0);
         var1 = c(var0);
         var1 = c(var0);
         var1 = c(var0);
         var1 = c(var0);
         var1 = c(var0);
         var1 = c(var0);
         var1 = c(var0);
         var1 = c(var0);
         var1 = d(var0);
         if (var1 != 0) {
         }

         var0.E = 0;
         var1 = c(var0);
         var0.E = var0.E << 8 | var1 & 255;
         var1 = c(var0);
         var0.E = var0.E << 8 | var1 & 255;
         var1 = c(var0);
         var0.E = var0.E << 8 | var1 & 255;

         for (int var56 = 0; var56 < 16; var56++) {
            var1 = d(var0);
            if (var1 == 1) {
               var0.v[var56] = true;
            } else {
               var0.v[var56] = false;
            }
         }

         for (int var57 = 0; var57 < 256; var57++) {
            var0.n[var57] = false;
         }

         for (int var58 = 0; var58 < 16; var58++) {
            if (var0.v[var58]) {
               for (int var67 = 0; var67 < 16; var67++) {
                  var1 = d(var0);
                  if (var1 == 1) {
                     var0.n[var58 * 16 + var67] = true;
                  }
               }
            }
         }

         a(var0);
         var7 = var0.K + 2;
         var8 = a(3, var0);
         var9 = a(15, var0);

         for (int var59 = 0; var59 < var9; var59++) {
            var5 = 0;

            while (true) {
               var1 = d(var0);
               if (var1 == 0) {
                  var0.s[var59] = (byte)var5;
                  break;
               }

               var5++;
            }
         }

         byte[] var27 = new byte[6];
         int var29 = 0;

         while (var29 < var8) {
            var27[var29] = var29++;
         }

         for (int var60 = 0; var60 < var9; var60++) {
            var29 = var0.s[var60];

            byte var28;
            for (var28 = var27[var29]; var29 > 0; var29--) {
               var27[var29] = var27[var29 - 1];
            }

            var27[0] = var28;
            var0.j[var60] = var28;
         }

         for (int var69 = 0; var69 < var8; var69++) {
            var17 = a(5, var0);

            for (int var61 = 0; var61 < var7; var61++) {
               while (true) {
                  var1 = d(var0);
                  if (var1 == 0) {
                     var0.B[var69][var61] = (byte)var17;
                     break;
                  }

                  var1 = d(var0);
                  if (var1 == 0) {
                     var17++;
                  } else {
                     var17--;
                  }
               }
            }
         }

         for (int var70 = 0; var70 < var8; var70++) {
            byte var2 = 32;
            byte var3 = 0;

            for (int var62 = 0; var62 < var7; var62++) {
               if (var0.B[var70][var62] > var3) {
                  var3 = var0.B[var70][var62];
               }

               if (var0.B[var70][var62] < var2) {
                  var2 = var0.B[var70][var62];
               }
            }

            a(var0.u[var70], var0.t[var70], var0.J[var70], var0.B[var70], var2, var3, var7);
            var0.D[var70] = var2;
         }

         var10 = var0.K + 1;
         var11 = -1;
         var12 = 0;

         for (int var63 = 0; var63 <= 255; var63++) {
            var0.m[var63] = 0;
         }

         var29 = 4095;

         for (int var96 = 15; var96 >= 0; var96--) {
            for (int var98 = 15; var98 >= 0; var98--) {
               var0.A[var29] = (byte)(var96 * 16 + var98);
               var29--;
            }

            var0.r[var96] = var29 + 1;
         }

         var14 = 0;
         if (var12 == 0) {
            var11++;
            var12 = 50;
            var21 = var0.j[var11];
            var22 = var0.D[var21];
            var23 = var0.u[var21];
            var25 = var0.J[var21];
            var24 = var0.t[var21];
         }

         var12--;
         var18 = var22;
         var19 = a(var22, var0);

         while (var19 > var23[var18]) {
            var18++;
            var20 = d(var0);
            var19 = var19 << 1 | var20;
         }

         var13 = var25[var19 - var24[var18]];

         while (var13 != var10) {
            if (var13 != 0 && var13 != 1) {
               int var33 = var13 - 1;
               if (var33 < 16) {
                  int var103 = var0.r[0];

                  for (var1 = var0.A[var103 + var33]; var33 > 3; var33 -= 4) {
                     int var34 = var103 + var33;
                     var0.A[var34] = var0.A[var34 - 1];
                     var0.A[var34 - 1] = var0.A[var34 - 2];
                     var0.A[var34 - 2] = var0.A[var34 - 3];
                     var0.A[var34 - 3] = var0.A[var34 - 4];
                  }

                  while (var33 > 0) {
                     var0.A[var103 + var33] = var0.A[var103 + var33 - 1];
                     var33--;
                  }

                  var0.A[var103] = var1;
               } else {
                  int var31 = var33 / 16;
                  int var32 = var33 % 16;
                  int var30 = var0.r[var31] + var32;

                  for (var1 = var0.A[var30]; var30 > var0.r[var31]; var30--) {
                     var0.A[var30] = var0.A[var30 - 1];
                  }

                  var0.r[var31]++;

                  while (var31 > 0) {
                     var0.r[var31]--;
                     var0.A[var0.r[var31]] = var0.A[var0.r[var31 - 1] + 16 - 1];
                     var31--;
                  }

                  var0.r[0]--;
                  var0.A[var0.r[0]] = var1;
                  if (var0.r[0] == 0) {
                     var29 = 4095;

                     for (int var97 = 15; var97 >= 0; var97--) {
                        for (int var99 = 15; var99 >= 0; var99--) {
                           var0.A[var29] = var0.A[var0.r[var97] + var99];
                           var29--;
                        }

                        var0.r[var97] = var29 + 1;
                     }
                  }
               }

               var0.m[var0.d[var1 & 255] & 255]++;
               ua.Mb[var14] = var0.d[var1 & 255] & 255;
               var14++;
               if (var12 == 0) {
                  var11++;
                  var12 = 50;
                  var21 = var0.j[var11];
                  var22 = var0.D[var21];
                  var23 = var0.u[var21];
                  var25 = var0.J[var21];
                  var24 = var0.t[var21];
               }

               var12--;
               var18 = var22;
               var19 = a(var22, var0);

               while (var19 > var23[var18]) {
                  var18++;
                  var20 = d(var0);
                  var19 = var19 << 1 | var20;
               }

               var13 = var25[var19 - var24[var18]];
            } else {
               var15 = -1;
               var16 = 1;

               do {
                  if (var13 == 0) {
                     var15 += 1 * var16;
                  } else if (var13 == 1) {
                     var15 += 2 * var16;
                  }

                  var16 *= 2;
                  if (var12 == 0) {
                     var11++;
                     var12 = 50;
                     var21 = var0.j[var11];
                     var22 = var0.D[var21];
                     var23 = var0.u[var21];
                     var25 = var0.J[var21];
                     var24 = var0.t[var21];
                  }

                  var12--;
                  var18 = var22;
                  var19 = a(var22, var0);

                  while (var19 > var23[var18]) {
                     var18++;
                     var20 = d(var0);
                     var19 = var19 << 1 | var20;
                  }

                  var13 = var25[var19 - var24[var18]];
               } while (var13 == 0 || var13 == 1);

               var15++;
               var1 = var0.d[var0.A[var0.r[0]] & 255];

               for (var0.m[var1 & 255] = var0.m[var1 & 255] + var15; var15 > 0; var15--) {
                  ua.Mb[var14] = var1 & 255;
                  var14++;
               }
            }
         }

         var0.F = 0;
         var0.g = 0;
         var0.w[0] = 0;

         for (int var64 = 1; var64 <= 256; var64++) {
            var0.w[var64] = var0.m[var64 - 1];
         }

         for (int var65 = 1; var65 <= 256; var65++) {
            var0.w[var65] = var0.w[var65] + var0.w[var65 - 1];
         }

         for (int var66 = 0; var66 < var14; var66++) {
            var1 = (byte)(ua.Mb[var66] & 0xFF);
            ua.Mb[var0.w[var1 & 255]] = ua.Mb[var0.w[var1 & 255]] | var66 << 8;
            var0.w[var1 & 255]++;
         }

         var0.H = ua.Mb[var0.E] >> 8;
         var0.L = 0;
         var0.H = ua.Mb[var0.H];
         var0.h = (byte)(var0.H & 0xFF);
         var0.H >>= 8;
         var0.L++;
         var0.b = var14;
         e(var0);
         if (var0.L == var0.b + 1 && var0.F == 0) {
            var26 = true;
         } else {
            var26 = false;
         }
      }
   }

   private static final byte c(ac var0) {
      return (byte)a(8, var0);
   }

   private static final byte d(ac var0) {
      return (byte)a(1, var0);
   }

   private static final void a(ac var0) {
      var0.K = 0;

      for (int var1 = 0; var1 < 256; var1++) {
         if (var0.n[var1]) {
            var0.d[var0.K] = (byte)var1;
            var0.K++;
         }
      }
   }

   private static final void a(int[] var0, int[] var1, int[] var2, byte[] var3, int var4, int var5, int var6) {
      int var7 = 0;

      for (int var8 = var4; var8 <= var5; var8++) {
         for (int var9 = 0; var9 < var6; var9++) {
            if (var3[var9] == var8) {
               var2[var7] = var9;
               var7++;
            }
         }
      }

      for (int var11 = 0; var11 < 23; var11++) {
         var1[var11] = 0;
      }

      for (int var12 = 0; var12 < var6; var12++) {
         var1[var3[var12] + 1]++;
      }

      for (int var13 = 1; var13 < 23; var13++) {
         var1[var13] += var1[var13 - 1];
      }

      for (int var14 = 0; var14 < 23; var14++) {
         var0[var14] = 0;
      }

      int var10 = 0;

      for (int var15 = var4; var15 <= var5; var15++) {
         var10 += var1[var15 + 1] - var1[var15];
         var0[var15] = var10 - 1;
         var10 <<= 1;
      }

      for (int var16 = var4 + 1; var16 <= var5; var16++) {
         var1[var16] = (var0[var16 - 1] + 1 << 1) - var1[var16];
      }
   }

   static final int a(byte[] var0, int var1, byte[] var2, int var3, int var4) {
      synchronized (a) {
         a.q = var2;
         a.a = var4;
         a.i = var0;
         a.o = 0;
         a.y = var1;
         a.p = 0;
         a.e = 0;
         a.c = 0;
         a.G = 0;
         b(a);
         var1 -= a.y;
         a.q = null;
         a.i = null;
         return var1;
      }
   }

   private static final void e(ac var0) {
      byte var2 = var0.g;
      int var3 = var0.F;
      int var4 = var0.L;
      int var5 = var0.h;
      int[] var6 = ua.Mb;
      int var7 = var0.H;
      byte[] var8 = var0.i;
      int var9 = var0.o;
      int var10 = var0.y;
      int var11 = var10;
      int var12 = var0.b + 1;

      label63:
      while (true) {
         if (var3 > 0) {
            while (true) {
               if (var10 == 0) {
                  break label63;
               }

               if (var3 == 1) {
                  if (var10 == 0) {
                     var3 = 1;
                     break label63;
                  }

                  var8[var9] = var2;
                  var9++;
                  var10--;
                  break;
               }

               var8[var9] = var2;
               var3--;
               var9++;
               var10--;
            }
         }

         while (var4 != var12) {
            var2 = (byte)var5;
            var7 = var6[var7];
            byte var1 = (byte)var7;
            var7 >>= 8;
            var4++;
            if (var1 != var5) {
               var5 = var1;
               if (var10 == 0) {
                  var3 = 1;
                  break label63;
               }

               var8[var9] = var2;
               var9++;
               var10--;
            } else {
               if (var4 != var12) {
                  var3 = 2;
                  var7 = var6[var7];
                  var1 = (byte)var7;
                  var7 >>= 8;
                  if (++var4 != var12) {
                     if (var1 != var5) {
                        var5 = var1;
                     } else {
                        var3 = 3;
                        var7 = var6[var7];
                        var1 = (byte)var7;
                        var7 >>= 8;
                        if (++var4 != var12) {
                           if (var1 != var5) {
                              var5 = var1;
                           } else {
                              var7 = var6[var7];
                              var1 = (byte)var7;
                              var7 >>= 8;
                              var4++;
                              var3 = (var1 & 255) + 4;
                              var7 = var6[var7];
                              var5 = (byte)var7;
                              var7 >>= 8;
                              var4++;
                           }
                        }
                     }
                  }
                  continue label63;
               }

               if (var10 == 0) {
                  var3 = 1;
                  break label63;
               }

               var8[var9] = var2;
               var9++;
               var10--;
            }
         }

         var3 = 0;
         break;
      }

      int var13 = var0.G;
      var0.G += var11 - var10;
      if (var0.G < var13) {
      }

      var0.g = var2;
      var0.F = var3;
      var0.L = var4;
      var0.h = var5;
      ua.Mb = var6;
      var0.H = var7;
      var0.i = var8;
      var0.o = var9;
      var0.y = var10;
   }

   private static final int a(int var0, ac var1) {
      while (var1.p < var0) {
         var1.e = var1.e << 8 | var1.q[var1.a] & 255;
         var1.p += 8;
         var1.a++;
         var1.c++;
         if (var1.c == 0) {
         }
      }

      int var2 = var1.e >> var1.p - var0 & (1 << var0) - 1;
      var1.p -= var0;
      return var2;
   }
}
