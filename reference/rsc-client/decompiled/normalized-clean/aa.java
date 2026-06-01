final class aa {
   static int e;
   static int h;
   static int a;
   static int[] c;
   static int b;
   static int d = 114;
   private int[] j;
   private byte[] i;
   static int l;
   private int[] g;
   static int f;
   static String[] k = new String[100];
   private static final String[] z = new String[]{
      z(z("7L\u001d]e8DG_$")), z(z("-\u0003\u001dOq")), z(z("8X_\r")), z(z("7L\u001d#$")), z(z("7L\u001d $")), z(z("7L\u001d\"$"))
   };

   final int a(int var1, byte[] var2, int var3, byte[] var4, int var5, int var6) {
      boolean var15 = client.vh;

      try {
         h++;
         if (var6 < 99) {
            this.a(-58, (byte[])null, 69, (byte[])null, -39, 22);
         }

         int var7 = 0;
         var1 += var5;
         int var8 = var3 << 3;

         int var10000;
         int var10001;
         while (true) {
            if (var5 < var1) {
               int var9 = var4[var5] & 255;
               int var10 = this.g[var9];
               byte var11 = this.i[var9];
               var10000 = ~var11;
               var10001 = -1;
               if (var15) {
                  break;
               }

               if (var10000 == -1) {
                  throw new RuntimeException("" + var9);
               }

               int var12 = var8 >> 3;
               int var13 = var8 & 7;
               int var17 = var7 & -var13 >> 31;
               int var14 = var12 + (var11 + var13 + -1 >> 3);
               var8 += var11;
               var13 += 24;
               var2[var12] = (byte)(var7 = d.a(var17, var10 >>> var13));
               if (~var14 < ~var12 || var15) {
                  var12++;
                  var13 -= 8;
                  var2[var12] = (byte)(var7 = var10 >>> var13);
                  if (~var12 > ~var14 || var15) {
                     var12++;
                     var13 -= 8;
                     var2[var12] = (byte)(var7 = var10 >>> var13);
                     if (var12 < var14 || var15) {
                        var12++;
                        var13 -= 8;
                        var2[var12] = (byte)(var7 = var10 >>> var13);
                        if (var12 < var14) {
                           var12++;
                           var13 -= 8;
                           var2[var12] = (byte)(var7 = var10 << -var13);
                        }
                     }
                  }
               }

               var5++;
               if (!var15) {
                  continue;
               }
            }

            var10000 = -var3;
            var10001 = var8 + 7 >> 3;
            break;
         }

         return var10000 + var10001;
      } catch (RuntimeException var16) {
         throw i.a(var16, z[3] + var1 + 44 + (var2 != null ? z[1] : z[2]) + 44 + var3 + 44 + (var4 != null ? z[1] : z[2]) + 44 + var5 + 44 + var6 + 41);
      }
   }

   final int a(byte[] var1, byte[] var2, int var3, int var4, int var5, int var6) {
      boolean var11 = client.vh;

      try {
         a++;
         if (0 == var6) {
            return 0;
         }

         int var7 = 0;
         var6 += var3;
         if (var5 != -1) {
            this.a(105, (byte[])null, 82, (byte[])null, 125, -45);
         }

         int var8 = var4;

         do {
            byte var9;
            label132: {
               var9 = var1[var8];
               if (var9 >= 0) {
                  var7++;
                  if (var11 || !var11) {
                     break label132;
                  }
               }

               var7 = this.j[var7];
            }

            int var10;
            if (0 > (var10 = this.j[var7])) {
               var2[var3++] = (byte)(~var10);
               if (var3 >= var6) {
                  break;
               }

               var7 = 0;
            }

            label205: {
               if ((64 & var9) != 0) {
                  var7 = this.j[var7];
                  if (!var11) {
                     break label205;
                  }
               }

               var7++;
            }

            if ((var10 = this.j[var7]) < 0) {
               var2[var3++] = (byte)(~var10);
               if (var6 <= var3 && !var11) {
                  break;
               }

               var7 = 0;
            }

            label199: {
               if (~(var9 & 32) == -1) {
                  var7++;
                  if (!var11) {
                     break label199;
                  }
               }

               var7 = this.j[var7];
            }

            if (~(var10 = this.j[var7]) > -1) {
               var2[var3++] = (byte)(~var10);
               if (~var3 <= ~var6 && !var11) {
                  break;
               }

               var7 = 0;
            }

            label193: {
               if (0 != (16 & var9)) {
                  var7 = this.j[var7];
                  if (!var11) {
                     break label193;
                  }
               }

               var7++;
            }

            if (~(var10 = this.j[var7]) > -1) {
               var2[var3++] = (byte)(~var10);
               if (var3 >= var6 && !var11) {
                  break;
               }

               var7 = 0;
            }

            label187: {
               if (-1 != ~(var9 & 8)) {
                  var7 = this.j[var7];
                  if (!var11) {
                     break label187;
                  }
               }

               var7++;
            }

            if (-1 < ~(var10 = this.j[var7])) {
               var2[var3++] = (byte)(~var10);
               if (var3 >= var6) {
                  break;
               }

               var7 = 0;
            }

            label181: {
               if ((4 & var9) != 0) {
                  var7 = this.j[var7];
                  if (!var11) {
                     break label181;
                  }
               }

               var7++;
            }

            if (0 > (var10 = this.j[var7])) {
               var2[var3++] = (byte)(~var10);
               if (~var6 >= ~var3 && !var11) {
                  break;
               }

               var7 = 0;
            }

            label175: {
               if (0 == (2 & var9)) {
                  var7++;
                  if (!var11) {
                     break label175;
                  }
               }

               var7 = this.j[var7];
            }

            if (~(var10 = this.j[var7]) > -1) {
               var2[var3++] = (byte)(~var10);
               if (~var6 >= ~var3 && !var11) {
                  break;
               }

               var7 = 0;
            }

            label169: {
               if (-1 != ~(1 & var9)) {
                  var7 = this.j[var7];
                  if (!var11) {
                     break label169;
                  }
               }

               var7++;
            }

            if (~(var10 = this.j[var7]) > -1) {
               var2[var3++] = (byte)(~var10);
               if (~var6 >= ~var3) {
                  break;
               }

               var7 = 0;
            }

            var8++;
         } while (!var11);

         return -var4 + 1 + var8;
      } catch (RuntimeException var12) {
         throw i.a(var12, z[4] + (var1 != null ? z[1] : z[2]) + 44 + (var2 != null ? z[1] : z[2]) + 44 + var3 + 44 + var4 + 44 + var5 + 44 + var6 + 41);
      }
   }

   static final int a(int var0, boolean var1) {
      try {
         e++;
         var0 = --var0 | var0 >>> 1;
         var0 |= var0 >>> 2;
         var0 |= var0 >>> 4;
         if (var1) {
            b = -4;
         }

         var0 |= var0 >>> 8;
         var0 |= var0 >>> 16;
         return var0 - -1;
      } catch (RuntimeException var3) {
         throw i.a(var3, z[5] + var0 + 44 + var1 + 41);
      }
   }

   aa(byte[] var1) {
      try {
         int var2 = var1.length;
         this.g = new int[var2];
         this.i = var1;
         this.j = new int[8];
         int[] var3 = new int[33];
         int var4 = 0;

         for (int var5 = 0; ~var5 > ~var2; var5++) {
            byte var6 = var1[var5];
            if (0 != var6) {
               int var7 = 1 << -var6 + 32;
               int var8 = var3[var6];
               this.g[var5] = var8;
               int var9;
               if (-1 != ~(var7 & var8)) {
                  var9 = var3[var6 - 1];
               } else {
                  for (int var10 = var6 - 1; ~var10 <= -2; var10--) {
                     int var11 = var3[var10];
                     if (var8 != var11) {
                        break;
                     }

                     int var12 = 1 << 0 - var10;
                     if ((var11 & var12) != 0) {
                        var3[var10] = var3[var10 - 1];
                        break;
                     }

                     var3[var10] = d.a(var12, var11);
                  }

                  var9 = var8 | var7;
               }

               var3[var6] = var9;

               for (int var16 = var6 - -1; var16 <= 32; var16++) {
                  if (~var8 == ~var3[var16]) {
                     var3[var16] = var9;
                  }
               }

               int var17 = 0;

               for (int var18 = 0; var6 > var18; var18++) {
                  int var19 = Integer.MIN_VALUE >>> var18;
                  if (-1 == ~(var19 & var8)) {
                     var17++;
                  } else {
                     if (~this.j[var17] == -1) {
                        this.j[var17] = var4;
                     }

                     var17 = this.j[var17];
                  }

                  if (this.j.length <= var17) {
                     int[] var13 = new int[this.j.length * 2];

                     for (int var14 = 0; ~this.j.length < ~var14; var14++) {
                        var13[var14] = this.j[var14];
                     }

                     this.j = var13;
                  }

                  var19 >>>= 1;
               }

               if (~var4 >= ~var17) {
                  var4 = var17 + 1;
               }

               this.j[var17] = ~var5;
            }
         }
      } catch (RuntimeException var15) {
         throw i.a(var15, z[0] + (var1 != null ? z[1] : z[2]) + ')');
      }
   }

   private static char[] z(String var0) {
      char[] var10000 = var0.toCharArray();
      if (var10000.length < 2) {
         var10000[0] = (char)(var10000[0] ^ 12);
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
               var10005 = 86;
               break;
            case 1:
               var10005 = 45;
               break;
            case 2:
               var10005 = 51;
               break;
            case 3:
               var10005 = 97;
               break;
            default:
               var10005 = 12;
         }

         var10001[var1] = (char)(var10004 ^ var10005);
      }

      return new String(var10001).intern();
   }
}
