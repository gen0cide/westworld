final class ub {
   static String[] c = new String[5000];
   static int f;
   static String[] b;
   static int e;
   static int[] g;
   static int d;
   static String[] a = new String[100];
   private static final String[] z = new String[]{
      z(z("\u0017\u000e\rJP")), z(z("\u0017\u000e\rKP")), z(z("\u0017\u000e\rIP")), z(z("\u0019B\r&\u0005")), z(z("\f\u0019Od"))
   };

   static final void a(int[] var0, byte var1, int var2, int var3, Object[] var4) {
      boolean var13 = client.vh;

      try {
         if (var1 <= -82) {
            int var22;
            int var24;
            label90: {
               label96: {
                  try {
                     if (~var2 <= ~var3) {
                        break label96;
                     }
                  } catch (RuntimeException var18) {
                     throw var18;
                  }

                  int var20 = (var3 + var2) / 2;
                  int var6 = var2;
                  int var7 = var0[var20];
                  var0[var20] = var0[var3];
                  var0[var3] = var7;
                  Object var8 = var4[var20];

                  label83: {
                     try {
                        var4[var20] = var4[var3];
                        var4[var3] = var8;
                        if (Integer.MIN_VALUE == ~var7) {
                           var21 = 0;
                           break label83;
                        }
                     } catch (RuntimeException var17) {
                        throw var17;
                     }

                     var21 = 1;
                  }

                  byte var9 = var21;
                  int var10 = var2;

                  while (var10 < var3) {
                     label72: {
                        try {
                           var22 = (var10 & var9) + var7;
                           var24 = var0[var10];
                           if (var13) {
                              break label90;
                           }

                           if (var22 <= var24) {
                              break label72;
                           }
                        } catch (RuntimeException var16) {
                           throw var16;
                        }

                        int var11 = var0[var10];
                        var0[var10] = var0[var6];
                        var0[var6] = var11;
                        Object var12 = var4[var10];
                        var4[var10] = var4[var6];
                        var4[var6++] = var12;
                     }

                     var10++;
                     if (var13) {
                        break;
                     }
                  }

                  var0[var3] = var0[var6];
                  var0[var6] = var7;
                  var4[var3] = var4[var6];
                  var4[var6] = var8;
                  a(var0, (byte)-124, var2, var6 + -1, var4);
                  a(var0, (byte)-123, 1 + var6, var3, var4);
               }

               var22 = e;
               var24 = 1;
            }

            e = var22 + var24;
         }
      } catch (RuntimeException var19) {
         RuntimeException var5 = var19;

         RuntimeException var10000;
         StringBuilder var10001;
         String var10002;
         label59: {
            try {
               var10000 = var5;
               var10001 = new StringBuilder().append(z[2]);
               if (var0 != null) {
                  var10002 = z[3];
                  break label59;
               }
            } catch (RuntimeException var15) {
               throw var15;
            }

            var10002 = z[4];
         }

         try {
            var10001 = var10001.append(var10002).append(',').append((int)var1).append(',').append(var2).append(',').append(var3).append(',');
            if (var4 != null) {
               throw i.a(var10000, var10001.append(z[3]).append(')').toString());
            }
         } catch (RuntimeException var14) {
            throw var14;
         }

         throw i.a(var10000, var10001.append(z[4]).append(')').toString());
      }
   }

   static final v a(int var0, byte var1) {
      boolean var5 = client.vh;

      try {
         try {
            d++;
            if (var1 != 24) {
               c = (String[])null;
            }
         } catch (RuntimeException var6) {
            throw var6;
         }

         v[] var2 = i.a(var1 + -735);
         int var3 = 0;

         while (~var2.length < ~var3) {
            v var4 = var2[var3];
            if (var0 == var4.i) {
               return var4;
            }

            var3++;
            if (var5) {
               break;
            }
         }

         return null;
      } catch (RuntimeException var7) {
         throw i.a(var7, z[0] + var0 + ',' + var1 + ')');
      }
   }

   static final int a(byte var0) {
      try {
         f++;
         int var1 = m.a(true, ka.b, b.v);
         if (~var1 < -100000000) {
            var1 = -var1 + 99999999;
         }

         try {
            if (var0 != -105) {
               g = (int[])null;
            }
         } catch (RuntimeException var2) {
            throw var2;
         }

         ka.b += 4;
         return var1;
      } catch (RuntimeException var3) {
         throw i.a(var3, z[1] + var0 + 41);
      }
   }

   private static char[] z(String var0) {
      char[] var10000 = var0.toCharArray();
      if (var10000.length < 2) {
         var10000[0] = (char)(var10000[0] ^ 'x');
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
               var10005 = 98;
               break;
            case 1:
               var10005 = 108;
               break;
            case 2:
               var10005 = 35;
               break;
            case 3:
               var10005 = 8;
               break;
            default:
               var10005 = 120;
         }

         var10001[var1] = (char)(var10004 ^ var10005);
      }

      return new String(var10001).intern();
   }
}
