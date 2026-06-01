import java.math.BigInteger;

class tb extends ib {
   static int B;
   static int v;
   static int r;
   static int D;
   static int x;
   static int g;
   static int m;
   static int o;
   static int i;
   byte[] F;
   static int A;
   static int q;
   int w;
   static int n;
   static int E;
   static int k;
   static int[] l = new int[12];
   static int s;
   static int C;
   static int u;
   static int j;
   static int z;
   static int p;
   static int h;
   static int y;
   static int t;
   private static final String[] G = new String[]{
      z(z("#.09R")),
      z(z("#.0%R")),
      z(z("#.0!R")),
      z(z(",b0\\\u0007")),
      z(z("99r\u001e")),
      z(z("#.03;\u007f")),
      z(z("#.0&R")),
      z(z("#.0?R")),
      z(z("#.0:R")),
      z(z("#.0$R")),
      z(z("#.0N\u00139%jLR")),
      z(z("#.04R")),
      z(z("#.0;R")),
      z(z("#.0\"R")),
      z(z("#.05R")),
      z(z("#.08R")),
      z(z("#.0<R")),
      z(z("#.0=R")),
      z(z("#.07R")),
      z(z("#.00;\u007f")),
      z(z("#.06;\u007f")),
      z(z("#.0#R")),
      z(z("#.0>R")),
      z(z("#.0 R")),
      z(z("#.0'R")),
      z(z("#.01;\u007f"))
   };

   final void a(String var1, int var2) {
      try {
         x++;
         int var3 = var1.indexOf(0);
         if (0 > var3) {
            this.F[this.w++] = 0;
            this.w = this.w + i.a(var1.length(), this.w, 0, var1, (byte)-118, this.F);
            int var4 = 53 / ((var2 - 45) / 55);
            this.F[this.w++] = 0;
         } else {
            throw new IllegalArgumentException("");
         }
      } catch (RuntimeException var5) {
         throw i.a(var5, G[17] + (var1 != null ? G[3] : G[4]) + ',' + var2 + ')');
      }
   }

   final void a(int var1, byte var2) {
      try {
         this.F[this.w++] = (byte)(var1 >> -592188912);
         q++;
         this.F[this.w++] = (byte)(var1 >> -1114664312);
         if (var2 == -13) {
            this.F[this.w++] = (byte)var1;
         }
      } catch (RuntimeException var4) {
         throw i.a(var4, G[12] + var1 + ',' + var2 + ')');
      }
   }

   final boolean e(int var1) {
      try {
         this.w -= 4;
         t++;
         if (var1 != -422797528) {
            return false;
         }

         int var2 = w.a(this.w, 107, this.F, 0);
         int var3 = this.b(-129);
         return var3 == var2;
      } catch (RuntimeException var4) {
         throw i.a(var4, G[19] + var1 + ')');
      }
   }

   private final void a(boolean var1, int var2, int var3, byte[] var4) {
      boolean var6 = client.vh;

      try {
         int var5 = var2;

         while (true) {
            if (~(var3 + var2) < ~var5) {
               var4[var5] = this.F[this.w++];
               var5++;
               if (var6) {
                  break;
               }

               if (!var6) {
                  continue;
               }
            }

            i++;
            break;
         }

         if (var1) {
            this.a((BigInteger)null, -94, (BigInteger)null);
         }
      } catch (RuntimeException var7) {
         throw i.a(var7, G[16] + var1 + ',' + var2 + ',' + var3 + ',' + (var4 != null ? G[3] : G[4]) + ')');
      }
   }

   final void a(int var1, int var2, int var3, byte[] var4) {
      boolean var6 = client.vh;

      try {
         int var5 = var1;

         while (true) {
            if (~(var3 + var1) < ~var5) {
               this.F[this.w++] = var4[var5];
               var5++;
               if (var6) {
                  break;
               }

               if (!var6) {
                  continue;
               }
            }

            if (var2 >= -120) {
               l = (int[])null;
            }

            D++;
            break;
         }
      } catch (RuntimeException var7) {
         throw i.a(var7, G[5] + var1 + ',' + var2 + ',' + var3 + ',' + (var4 != null ? G[3] : G[4]) + ')');
      }
   }

   final void c(int var1, int var2) {
      try {
         this.F[this.w++] = (byte)var1;
         int var3 = 121 / ((-5 - var2) / 32);
         E++;
      } catch (RuntimeException var4) {
         throw i.a(var4, G[0] + var1 + ',' + var2 + ')');
      }
   }

   final void a(byte var1, String var2) {
      try {
         A++;
         int var3 = var2.indexOf(0);
         if (~var3 > -1) {
            this.w = this.w + i.a(var2.length(), this.w, 0, var2, (byte)-112, this.F);
            this.F[this.w++] = 0;
            if (var1 != -39) {
               this.h(-74);
            }
         } else {
            throw new IllegalArgumentException("");
         }
      } catch (RuntimeException var4) {
         throw i.a(var4, G[15] + var1 + ',' + (var2 != null ? G[3] : G[4]) + ')');
      }
   }

   final void e(int var1, int var2) {
      try {
         C++;
         this.F[this.w++] = (byte)(var2 >> 436303880);
         if (var1 == 393) {
            this.F[this.w++] = (byte)var2;
         }
      } catch (RuntimeException var4) {
         throw i.a(var4, G[11] + var1 + ',' + var2 + ')');
      }
   }

   final byte h(int var1) {
      try {
         if (var1 != 20869) {
            return 113;
         }

         v++;
         return this.F[this.w++];
      } catch (RuntimeException var3) {
         throw i.a(var3, G[22] + var1 + ')');
      }
   }

   final int a(byte var1) {
      try {
         p++;
         if (var1 != 104) {
            this.b(111, (byte)-26);
         }

         return this.F[this.w++] & 0xFF;
      } catch (RuntimeException var3) {
         throw i.a(var3, G[18] + var1 + 41);
      }
   }

   final void a(byte var1, int var2, int[] var3, int var4) {
      boolean var13 = client.vh;

      try {
         m++;
         int var5 = this.w;
         if (var1 == 87) {
            this.w = var2;
            int var6 = (-var2 + var4) / 8;
            int var7 = 0;

            while (true) {
               if (var7 < var6) {
                  int var8 = this.b(-129);
                  int var9 = this.b(-129);
                  int var10 = 0;
                  int var11 = -1640531527;
                  if (var13) {
                     break;
                  }

                  int var12 = 32;

                  label47: {
                     while (~(var12--) < -1) {
                        var8 += (var9 << 1853481540 ^ var9 >>> -1249842747) + var9 ^ var10 + var3[var10 & 3];
                        var10 += var11;
                        var9 += var8 + (var8 >>> -820868603 ^ var8 << 683776932) ^ var10 - -var3[(7145 & var10) >>> 2036143115];
                        if (var13) {
                           break label47;
                        }

                        if (var13) {
                           break;
                        }
                     }

                     this.w -= 8;
                     this.b(-422797528, var8);
                     this.b(-422797528, var9);
                     var7++;
                  }

                  if (!var13) {
                     continue;
                  }
               }

               this.w = var5;
               break;
            }
         }
      } catch (RuntimeException var14) {
         throw i.a(var14, G[25] + var1 + ',' + var2 + ',' + (var3 != null ? G[3] : G[4]) + ',' + var4 + ')');
      }
   }

   final void d(int var1, int var2) {
      try {
         this.F[-2 + -var1 + this.w] = (byte)(var1 >> 2065078440);
         k++;
         this.F[this.w + -var1 + -1] = (byte)var1;
         if (var2 != 1) {
            this.a((String)null, 53);
         }
      } catch (RuntimeException var4) {
         throw i.a(var4, G[23] + var1 + ',' + var2 + ')');
      }
   }

   final int b(byte var1) {
      try {
         if (var1 != 68) {
            return 53;
         }

         y++;
         int var2 = 255 & this.F[this.w];
         return -129 < ~var2 ? this.a((byte)104) : -32768 + this.f(255);
      } catch (RuntimeException var3) {
         throw i.a(var3, G[13] + var1 + 41);
      }
   }

   final int a(boolean var1) {
      try {
         this.w += 2;
         u++;
         if (var1) {
            return -8;
         }

         int var2 = (255 & this.F[this.w - 1]) + (this.F[-2 + this.w] << -1500474744 & 0xFF00);
         if (-32768 > ~var2) {
            var2 -= 65536;
         }

         return var2;
      } catch (RuntimeException var3) {
         throw i.a(var3, G[20] + var1 + 41);
      }
   }

   final int c(int var1) {
      try {
         n++;
         if (var1 != 103) {
            return 72;
         } else {
            return 0 > this.F[this.w] ? 2147483647 & this.b(-129) : this.f(var1 + 152);
         }
      } catch (RuntimeException var3) {
         throw i.a(var3, G[1] + var1 + 41);
      }
   }

   final String c(byte var1) {
      try {
         o++;
         if (var1 != -44) {
            this.d(-84);
         }

         byte var2 = this.F[this.w++];
         if (var2 != 0) {
            throw new IllegalStateException("");
         }

         int var3 = this.w;

         while (this.F[this.w++] != 0) {
         }

         int var4 = -1 + (this.w - var3);
         return -1 == ~var4 ? "" : ga.a(var4, var1 + -68, var3, this.F);
      } catch (RuntimeException var5) {
         throw i.a(var5, G[7] + var1 + ')');
      }
   }

   final void b(int var1, byte var2) {
      try {
         if (var2 <= -62) {
            h++;
            if (0 <= var1 && 128 > var1) {
               this.c(var1, 43);
            } else if (var1 >= 0 && var1 < 32768) {
               this.e(393, 32768 - -var1);
            } else {
               throw new IllegalArgumentException();
            }
         }
      } catch (RuntimeException var4) {
         throw i.a(var4, G[8] + var1 + ',' + var2 + ')');
      }
   }

   final void b(int var1, int var2) {
      try {
         s++;
         this.F[this.w++] = (byte)(var2 >> -2105201640);
         if (var1 != -422797528) {
            this.c(-62, 1);
         }

         this.F[this.w++] = (byte)(var2 >> -952226864);
         this.F[this.w++] = (byte)(var2 >> -422797528);
         this.F[this.w++] = (byte)var2;
      } catch (RuntimeException var4) {
         throw i.a(var4, G[21] + var1 + ',' + var2 + ')');
      }
   }

   final int f(int var1) {
      try {
         B++;
         if (var1 != 255) {
            this.a((BigInteger)null, 71, (BigInteger)null);
         }

         this.w += 2;
         return ((this.F[-2 + this.w] & 0xFF) << -958656888) - -(0xFF & this.F[-1 + this.w]);
      } catch (RuntimeException var3) {
         throw i.a(var3, G[14] + var1 + 41);
      }
   }

   final int b(int var1) {
      try {
         this.w += 4;
         if (var1 != -129) {
            return 124;
         }

         r++;
         return (this.F[this.w - 3] << 1172488496 & 0xFF0000)
            + (this.F[this.w + -4] << 2040727736 & 0xFF000000)
            - (-(0xFF00 & this.F[this.w + -2] << -1377058840) - (this.F[-1 + this.w] & 0xFF));
      } catch (RuntimeException var3) {
         throw i.a(var3, G[24] + var1 + 41);
      }
   }

   final long g(int var1) {
      try {
         if (var1 != 0) {
            return -13L;
         }

         g++;
         long var2 = this.b(-129) & 4294967295L;
         long var4 = this.b(-129) & 4294967295L;
         return (var2 << 1382465952) - -var4;
      } catch (RuntimeException var6) {
         throw i.a(var6, G[2] + var1 + ')');
      }
   }

   tb(int var1) {
      try {
         this.F = mb.a(var1, (byte)-104);
         this.w = 0;
      } catch (RuntimeException var3) {
         throw i.a(var3, G[10] + var1 + ')');
      }
   }

   final byte[] d(int var1) {
      boolean var4 = client.vh;

      try {
         z++;
         byte[] var2 = new byte[this.w];
         int var3 = var1;

         byte[] var10000;
         while (true) {
            if (~var3 > ~this.w) {
               var10000 = var2;
               if (var4) {
                  break;
               }

               var2[var3] = this.F[var3];
               var3++;
               if (!var4) {
                  continue;
               }
            }

            var10000 = var2;
            break;
         }

         return var10000;
      } catch (RuntimeException var5) {
         throw i.a(var5, G[6] + var1 + ')');
      }
   }

   tb(byte[] var1) {
      try {
         this.F = var1;
         this.w = 0;
      } catch (RuntimeException var3) {
         throw i.a(var3, G[10] + (var1 != null ? G[3] : G[4]) + ')');
      }
   }

   final void a(BigInteger var1, int var2, BigInteger var3) {
      try {
         int var4 = -98 / ((var2 - 6) / 52);
         j++;
         int var5 = this.w;
         this.w = 0;
         byte[] var6 = new byte[var5];
         this.a(false, 0, var5, var6);
         BigInteger var7 = new BigInteger(var6);
         BigInteger var8 = var7.modPow(var3, var1);
         byte[] var9 = var8.toByteArray();
         this.w = 0;
         this.e(393, var9.length);
         this.a(0, -127, var9.length, var9);
      } catch (RuntimeException var10) {
         throw i.a(var10, G[9] + (var1 != null ? G[3] : G[4]) + ',' + var2 + ',' + (var3 != null ? G[3] : G[4]) + ')');
      }
   }

   private static char[] z(String var0) {
      char[] var10000 = var0.toCharArray();
      if (var10000.length < 2) {
         var10000[0] = (char)(var10000[0] ^ 122);
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
               var10005 = 87;
               break;
            case 1:
               var10005 = 76;
               break;
            case 2:
               var10005 = 30;
               break;
            case 3:
               var10005 = 114;
               break;
            default:
               var10005 = 122;
         }

         var10001[var1] = (char)(var10004 ^ var10005);
      }

      return new String(var10001).intern();
   }
}
