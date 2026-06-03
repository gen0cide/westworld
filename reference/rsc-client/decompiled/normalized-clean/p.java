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
         }

         long var1 = System.currentTimeMillis();
         if (~client.ze < ~var1) {
            wb.w = wb.w + client.ze + -var1;
         }

         client.ze = var1;
         return wb.w + var1;
      } catch (RuntimeException var3) {
         throw i.a(var3, z[0] + var0 + ')');
      }
   }

   static final void a(
      int var0,
      int var1,
      int var2,
      int var3,
      int var4,
      int[] var5,
      int var6,
      int var7,
      int var8,
      int var9,
      int[] var10,
      int var11,
      int var12,
      int var13,
      int var14
   ) {
      boolean var22 = client.vh;

      try {
         f++;
         if (0 < var14) {
            int var15 = 0;
            int var16 = 0;
            if (var12 != 0) {
               var16 = var3 / var12 << 6;
               var15 = var8 / var12 << 6;
            }

            label111: {
               var0 <<= 2;
               if (~var15 <= -1) {
                  if (var15 <= 4032) {
                     break label111;
                  }

                  var15 = 4032;
                  if (!var22) {
                     break label111;
                  }
               }

               var15 = 0;
            }

            if (var1 != 1121159302) {
               a(-69, 127, -20, -29, -78, (int[])null, 16, 2, -77, -5, (int[])null, 113, -57, 68, -87);
            }

            int var19 = var14;

            label103:
            while (true) {
               byte var10000 = 0;
               int var10001 = var19;

               label101:
               while (var10000 < var10001) {
                  var12 += var13;
                  var8 += var4;
                  var3 += var2;
                  var9 = var15;
                  var7 = var16;
                  if (var22) {
                     return;
                  }

                  if (-1 != ~var12) {
                     var15 = var8 / var12 << 6;
                     var16 = var3 / var12 << 6;
                  }

                  label98: {
                     if (-1 < ~var15) {
                        var15 = 0;
                        if (!var22) {
                           break label98;
                        }
                     }

                     if (~var15 < -4033) {
                        var15 = 4032;
                     }
                  }

                  label120: {
                     int var18 = var16 + -var7 >> 4;
                     int var17 = -var9 + var15 >> 4;
                     int var20 = var6 >> 20;
                     var9 += 786432 & var6;
                     var6 += var0;
                     if (~var19 <= -17) {
                        var10[var11++] = var5[ib.a(var7, 4032) - -(var9 >> 6)] >>> var20;
                        var7 += var18;
                        var9 += var17;
                        var10[var11++] = var5[(var9 >> 6) + ib.a(var7, 4032)] >>> var20;
                        var7 += var18;
                        var9 += var17;
                        var10[var11++] = var5[(var9 >> 6) + ib.a(4032, var7)] >>> var20;
                        var9 += var17;
                        var7 += var18;
                        var10[var11++] = var5[(var9 >> 6) + ib.a(4032, var7)] >>> var20;
                        var9 += var17;
                        var7 += var18;
                        int var24 = var6 >> 20;
                        var9 = (var6 & 786432) + (4095 & var9);
                        var6 += var0;
                        var10[var11++] = var5[ib.a(4032, var7) + (var9 >> 6)] >>> var24;
                        var7 += var18;
                        var9 += var17;
                        var10[var11++] = var5[ib.a(var7, 4032) - -(var9 >> 6)] >>> var24;
                        var9 += var17;
                        var7 += var18;
                        var10[var11++] = var5[ib.a(var7, 4032) + (var9 >> 6)] >>> var24;
                        var7 += var18;
                        var9 += var17;
                        var10[var11++] = var5[(var9 >> 6) + ib.a(4032, var7)] >>> var24;
                        var9 += var17;
                        var7 += var18;
                        int var25 = var6 >> 20;
                        var9 = (786432 & var6) + (4095 & var9);
                        var6 += var0;
                        var10[var11++] = var5[ib.a(var7, 4032) - -(var9 >> 6)] >>> var25;
                        var7 += var18;
                        var9 += var17;
                        var10[var11++] = var5[(var9 >> 6) + ib.a(var7, 4032)] >>> var25;
                        var7 += var18;
                        var9 += var17;
                        var10[var11++] = var5[ib.a(var7, 4032) - -(var9 >> 6)] >>> var25;
                        var7 += var18;
                        var9 += var17;
                        var10[var11++] = var5[ib.a(4032, var7) + (var9 >> 6)] >>> var25;
                        var7 += var18;
                        var9 += var17;
                        var20 = var6 >> 20;
                        var9 = (4095 & var9) - -(var6 & 786432);
                        var6 += var0;
                        var10[var11++] = var5[ib.a(var7, 4032) + (var9 >> 6)] >>> var20;
                        var7 += var18;
                        var9 += var17;
                        var10[var11++] = var5[(var9 >> 6) + ib.a(var7, 4032)] >>> var20;
                        var7 += var18;
                        var9 += var17;
                        var10[var11++] = var5[(var9 >> 6) + ib.a(var7, 4032)] >>> var20;
                        var9 += var17;
                        var7 += var18;
                        var10[var11++] = var5[ib.a(4032, var7) - -(var9 >> 6)] >>> var20;
                        if (!var22) {
                           break label120;
                        }
                     }

                     int var21 = 0;

                     while (var21 < var19) {
                        var10[var11++] = var5[(var9 >> 6) + ib.a(4032, var7)] >>> var20;
                        var7 += var18;
                        var9 += var17;
                        var10000 = 3;
                        var10001 = 3 & var21;
                        if (var22) {
                           continue label101;
                        }

                        if (3 == var10001) {
                           var20 = var6 >> 20;
                           var9 = (var6 & 786432) + (4095 & var9);
                           var6 += var0;
                        }

                        var21++;
                        if (var22) {
                           break;
                        }
                     }
                  }

                  var19 -= 16;
                  if (var22) {
                     return;
                  }
                  continue label103;
               }

               return;
            }
         }
      } catch (RuntimeException var23) {
         throw i.a(
            var23,
            z[3]
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
               + (var5 != null ? z[2] : z[1])
               + ','
               + var6
               + ','
               + var7
               + ','
               + var8
               + ','
               + var9
               + ','
               + (var10 != null ? z[2] : z[1])
               + ','
               + var11
               + ','
               + var12
               + ','
               + var13
               + ','
               + var14
               + ')'
         );
      }
   }

   private static char[] z(String var0) {
      char[] var10000 = var0.toCharArray();
      if (var10000.length < 2) {
         var10000[0] = (char)(var10000[0] ^ 114);
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
