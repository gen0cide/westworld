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
         int var8 = var1.indexOf(0);

         try {
            if (0 <= var8) {
               throw new IllegalArgumentException("");
            }
         } catch (RuntimeException var6) {
            throw var6;
         }

         this.F[this.w++] = 0;
         this.w = this.w + i.a(var1.length(), this.w, 0, var1, (byte)-118, this.F);
         int var4 = 53 / ((var2 - 45) / 55);
         this.F[this.w++] = 0;
      } catch (RuntimeException var7) {
         RuntimeException var3 = var7;

         RuntimeException var10000;
         StringBuilder var10001;
         try {
            var10000 = var3;
            var10001 = new StringBuilder().append(G[17]);
            if (var1 != null) {
               throw i.a(var3, var10001.append(G[3]).append(',').append(var2).append(')').toString());
            }
         } catch (RuntimeException var5) {
            throw var5;
         }

         throw i.a(var10000, var10001.append(G[4]).append(',').append(var2).append(')').toString());
      }
   }

   final void a(int var1, byte var2) {
      try {
         this.F[this.w++] = (byte)(var1 >> 16);
         q++;
         this.F[this.w++] = (byte)(var1 >> 8);
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
         } else {
            int var2 = w.a(this.w, 107, this.F, 0);
            int var3 = this.b(-129);
            return var3 == var2;
         }
      } catch (RuntimeException var4) {
         throw i.a(var4, G[19] + var1 + ')');
      }
   }

   private final void a(boolean var1, int var2, int var3, byte[] var4) {
      boolean var6 = client.vh;

      try {
         int var11 = var2;

         while (true) {
            if (~(var3 + var2) < ~var11) {
               try {
                  var4[var11] = this.F[this.w++];
                  var11++;
                  if (var6) {
                     break;
                  }

                  if (!var6) {
                     continue;
                  }
               } catch (RuntimeException var9) {
                  throw var9;
               }
            }

            i++;
            break;
         }

         try {
            if (var1) {
               this.a((BigInteger)null, -94, (BigInteger)null);
            }
         } catch (RuntimeException var7) {
            throw var7;
         }
      } catch (RuntimeException var10) {
         RuntimeException var5 = var10;

         RuntimeException var10000;
         StringBuilder var10001;
         try {
            var10000 = var5;
            var10001 = new StringBuilder().append(G[16]).append(var1).append(',').append(var2).append(',').append(var3).append(',');
            if (var4 != null) {
               throw i.a(var5, var10001.append(G[3]).append(')').toString());
            }
         } catch (RuntimeException var8) {
            throw var8;
         }

         throw i.a(var10000, var10001.append(G[4]).append(')').toString());
      }
   }

   final void a(int var1, int var2, int var3, byte[] var4) {
      boolean var6 = client.vh;

      try {
         int var11 = var1;

         while (true) {
            if (~(var3 + var1) < ~var11) {
               try {
                  this.F[this.w++] = var4[var11];
                  var11++;
                  if (var6) {
                     break;
                  }

                  if (!var6) {
                     continue;
                  }
               } catch (RuntimeException var9) {
                  throw var9;
               }
            }

            try {
               if (var2 >= -120) {
                  l = (int[])null;
               }
            } catch (RuntimeException var8) {
               throw var8;
            }

            D++;
            break;
         }
      } catch (RuntimeException var10) {
         RuntimeException var5 = var10;

         RuntimeException var10000;
         StringBuilder var10001;
         try {
            var10000 = var5;
            var10001 = new StringBuilder().append(G[5]).append(var1).append(',').append(var2).append(',').append(var3).append(',');
            if (var4 != null) {
               throw i.a(var5, var10001.append(G[3]).append(')').toString());
            }
         } catch (RuntimeException var7) {
            throw var7;
         }

         throw i.a(var10000, var10001.append(G[4]).append(')').toString());
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
         int var8 = var2.indexOf(0);

         try {
            if (~var8 <= -1) {
               throw new IllegalArgumentException("");
            }
         } catch (RuntimeException var6) {
            throw var6;
         }

         try {
            this.w = this.w + i.a(var2.length(), this.w, 0, var2, (byte)-112, this.F);
            this.F[this.w++] = 0;
            if (var1 != -39) {
               this.h(-74);
            }
         } catch (RuntimeException var4) {
            throw var4;
         }
      } catch (RuntimeException var7) {
         RuntimeException var3 = var7;

         RuntimeException var10000;
         StringBuilder var10001;
         try {
            var10000 = var3;
            var10001 = new StringBuilder().append(G[15]).append((int)var1).append(',');
            if (var2 != null) {
               throw i.a(var3, var10001.append(G[3]).append(')').toString());
            }
         } catch (RuntimeException var5) {
            throw var5;
         }

         throw i.a(var10000, var10001.append(G[4]).append(')').toString());
      }
   }

   final void e(int var1, int var2) {
      try {
         C++;
         this.F[this.w++] = (byte)(var2 >> 8);
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
         } else {
            v++;
            return this.F[this.w++];
         }
      } catch (RuntimeException var3) {
         throw i.a(var3, G[22] + var1 + ')');
      }
   }

   final int a(byte var1) {
      try {
         try {
            p++;
            if (var1 != 104) {
               this.b(111, (byte)-26);
            }
         } catch (RuntimeException var3) {
            throw var3;
         }

         return this.F[this.w++] & 0xFF;
      } catch (RuntimeException var4) {
         throw i.a(var4, G[18] + var1 + 41);
      }
   }

   final void a(byte var1, int var2, int[] var3, int var4) {
      boolean var13 = client.vh;

      try {
         m++;
         int var17 = this.w;
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

                  label57: {
                     while (~(var12--) < -1) {
                        var8 += (var9 << 4 ^ var9 >>> 5) + var9 ^ var10 + var3[var10 & 3];
                        var10 += var11;
                        var9 += var8 + (var8 >>> 5 ^ var8 << 4) ^ var10 - -var3[(7145 & var10) >>> 11];
                        boolean var19 = var13;

                        try {
                           if (var19) {
                              break label57;
                           }

                           if (var13) {
                              break;
                           }
                        } catch (RuntimeException var15) {
                           throw var15;
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

               this.w = var17;
               break;
            }
         }
      } catch (RuntimeException var16) {
         RuntimeException var5 = var16;

         RuntimeException var10000;
         StringBuilder var10001;
         try {
            var10000 = var5;
            var10001 = new StringBuilder().append(G[25]).append((int)var1).append(',').append(var2).append(',');
            if (var3 != null) {
               throw i.a(var5, var10001.append(G[3]).append(',').append(var4).append(')').toString());
            }
         } catch (RuntimeException var14) {
            throw var14;
         }

         throw i.a(var10000, var10001.append(G[4]).append(',').append(var4).append(')').toString());
      }
   }

   final void d(int var1, int var2) {
      try {
         this.F[-2 + -var1 + this.w] = (byte)(var1 >> 8);
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
         } else {
            y++;
            int var2 = 255 & this.F[this.w];
            return -129 < ~var2 ? this.a((byte)104) : -32768 + this.f(255);
         }
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
         } else {
            int var2 = (255 & this.F[this.w - 1]) + (this.F[-2 + this.w] << 8 & 0xFF00);
            if (-32768 > ~var2) {
               var2 -= 65536;
            }

            return var2;
         }
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
         try {
            o++;
            if (var1 != -44) {
               this.d(-84);
            }
         } catch (RuntimeException var7) {
            throw var7;
         }

         byte var2 = this.F[this.w++];

         try {
            if (var2 != 0) {
               throw new IllegalStateException("");
            }
         } catch (RuntimeException var6) {
            throw var6;
         }

         int var3 = this.w;

         try {
            while (this.F[this.w++] != 0) {
            }
         } catch (RuntimeException var5) {
            throw var5;
         }

         int var4 = -1 + (this.w - var3);
         return -1 == ~var4 ? "" : ga.a(var4, var1 + -68, var3, this.F);
      } catch (RuntimeException var8) {
         throw i.a(var8, G[7] + var1 + ')');
      }
   }

   final void b(int var1, byte var2) {
      try {
         if (var2 <= -62) {
            label49: {
               try {
                  h++;
                  if (0 <= var1 && 128 > var1) {
                     break label49;
                  }
               } catch (RuntimeException var5) {
                  throw var5;
               }

               label35: {
                  try {
                     if (var1 >= 0 && var1 < 32768) {
                        break label35;
                     }
                  } catch (RuntimeException var4) {
                     throw var4;
                  }

                  throw new IllegalArgumentException();
               }

               this.e(393, 32768 - -var1);
               return;
            }

            this.c(var1, 43);
         }
      } catch (RuntimeException var6) {
         throw i.a(var6, G[8] + var1 + ',' + var2 + ')');
      }
   }

   final void b(int var1, int var2) {
      try {
         try {
            s++;
            this.F[this.w++] = (byte)(var2 >> 24);
            if (var1 != -422797528) {
               this.c(-62, 1);
            }
         } catch (RuntimeException var4) {
            throw var4;
         }

         this.F[this.w++] = (byte)(var2 >> 16);
         this.F[this.w++] = (byte)(var2 >> 8);
         this.F[this.w++] = (byte)var2;
      } catch (RuntimeException var5) {
         throw i.a(var5, G[21] + var1 + ',' + var2 + ')');
      }
   }

   final int f(int var1) {
      try {
         try {
            B++;
            if (var1 != 255) {
               this.a((BigInteger)null, 71, (BigInteger)null);
            }
         } catch (RuntimeException var3) {
            throw var3;
         }

         this.w += 2;
         return ((this.F[-2 + this.w] & 0xFF) << 8) - -(0xFF & this.F[-1 + this.w]);
      } catch (RuntimeException var4) {
         throw i.a(var4, G[14] + var1 + 41);
      }
   }

   final int b(int var1) {
      try {
         this.w += 4;
         if (var1 != -129) {
            return 124;
         } else {
            r++;
            return (this.F[this.w - 3] << 16 & 0xFF0000)
               + (this.F[this.w + -4] << 24 & 0xFF000000)
               - (-(0xFF00 & this.F[this.w + -2] << 8) - (this.F[-1 + this.w] & 0xFF));
         }
      } catch (RuntimeException var3) {
         throw i.a(var3, G[24] + var1 + 41);
      }
   }

   final long g(int var1) {
      try {
         if (var1 != 0) {
            return -13L;
         } else {
            g++;
            long var2 = (long)this.b(-129) & 4294967295L;
            long var4 = (long)this.b(-129) & 4294967295L;
            return (var2 << 0) - -var4;
         }
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
               try {
                  var10000 = var2;
                  if (var4) {
                     break;
                  }

                  var2[var3] = this.F[var3];
                  var3++;
                  if (!var4) {
                     continue;
                  }
               } catch (RuntimeException var5) {
                  throw var5;
               }
            }

            var10000 = var2;
            break;
         }

         return var10000;
      } catch (RuntimeException var6) {
         throw i.a(var6, G[6] + var1 + ')');
      }
   }

   tb(byte[] var1) {
      try {
         this.F = var1;
         this.w = 0;
      } catch (RuntimeException var4) {
         RuntimeException var2 = var4;

         RuntimeException var10000;
         StringBuilder var10001;
         try {
            var10000 = var2;
            var10001 = new StringBuilder().append(G[10]);
            if (var1 != null) {
               throw i.a(var2, var10001.append(G[3]).append(')').toString());
            }
         } catch (RuntimeException var3) {
            throw var3;
         }

         throw i.a(var10000, var10001.append(G[4]).append(')').toString());
      }
   }

   final void a(BigInteger var1, int var2, BigInteger var3) {
      try {
         int var13 = -98 / ((var2 - 6) / 52);
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
      } catch (RuntimeException var12) {
         RuntimeException var4 = var12;

         RuntimeException var10000;
         StringBuilder var10001;
         String var10002;
         label33: {
            try {
               var10000 = var4;
               var10001 = new StringBuilder().append(G[9]);
               if (var1 != null) {
                  var10002 = G[3];
                  break label33;
               }
            } catch (RuntimeException var11) {
               throw var11;
            }

            var10002 = G[4];
         }

         try {
            var10001 = var10001.append(var10002).append(',').append(var2).append(',');
            if (var3 != null) {
               throw i.a(var10000, var10001.append(G[3]).append(')').toString());
            }
         } catch (RuntimeException var10) {
            throw var10;
         }

         throw i.a(var10000, var10001.append(G[4]).append(')').toString());
      }
   }

   private static char[] z(String var0) {
      char[] var10000 = var0.toCharArray();
      if (var10000.length < 2) {
         var10000[0] = (char)(var10000[0] ^ 'z');
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
