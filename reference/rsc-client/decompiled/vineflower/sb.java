final class sb extends va {
   private int t;
   private int x;
   private int o;
   private int p;
   private int n;
   private int r;
   private int y;
   private int s;
   private int v;
   private int l;
   private int m;
   private int w;
   private int k;
   private int q;
   private boolean u;

   @Override
   final int c() {
      int var1 = this.w * 3 >> 6;
      var1 = (var1 ^ var1 >> 31) + (var1 >>> 31);
      if (this.l == 0) {
         var1 -= var1 * this.v / (((vb)this.h).l.length << 8);
      } else if (this.l >= 0) {
         var1 -= var1 * this.x / ((vb)this.h).l.length;
      }

      return var1 > 255 ? 255 : var1;
   }

   private static final int b(byte[] var0, int[] var1, int var2, int var3, int var4, int var5, int var6, int var7, int var8, sb var9) {
      var2 >>= 8;
      var8 >>= 8;
      var4 <<= 2;
      var5 <<= 2;
      if ((var6 = var3 + var2 - (var8 - 1)) > var7) {
         var6 = var7;
      }

      var9.k = var9.k + var9.p * (var6 - var3);
      var9.m = var9.m + var9.o * (var6 - var3);
      var6 -= 3;

      while (var3 < var6) {
         var1[var3++] += var0[var2--] * var4;
         var4 += var5;
         var1[var3++] += var0[var2--] * var4;
         var4 += var5;
         var1[var3++] += var0[var2--] * var4;
         var4 += var5;
         var1[var3++] += var0[var2--] * var4;
         var4 += var5;
      }

      for (int var24 = var6 + 3; var3 < var24; var4 += var5) {
         var1[var3++] += var0[var2--] * var4;
      }

      var9.w = var4 >> 2;
      var9.v = var2 << 8;
      return var3;
   }

   @Override
   final synchronized void b(int[] var1, int var2, int var3) {
      if (this.r == 0 && this.y == 0) {
         this.b(var3);
      } else {
         vb var4 = (vb)this.h;
         int var5 = this.x << 8;
         int var6 = this.t << 8;
         int var7 = var4.l.length << 8;
         int var8 = var6 - var5;
         if (var8 <= 0) {
            this.l = 0;
         }

         int var9 = var2;
         var3 += var2;
         if (this.v < 0) {
            if (this.n <= 0) {
               this.g();
               this.a(-27331);
               return;
            }

            this.v = 0;
         }

         if (this.v >= var7) {
            if (this.n >= 0) {
               this.g();
               this.a(-27331);
               return;
            }

            this.v = var7 - 1;
         }

         if (this.l < 0) {
            if (this.u) {
               if (this.n < 0) {
                  var9 = this.b(var1, var2, var5, var3, var4.l[this.x]);
                  if (this.v >= var5) {
                     return;
                  }

                  this.v = var5 + var5 - 1 - this.v;
                  this.n = -this.n;
               }

               while (true) {
                  var9 = this.a(var1, var9, var6, var3, var4.l[this.t - 1]);
                  if (this.v < var6) {
                     return;
                  }

                  this.v = var6 + var6 - 1 - this.v;
                  this.n = -this.n;
                  var9 = this.b(var1, var9, var5, var3, var4.l[this.x]);
                  if (this.v >= var5) {
                     return;
                  }

                  this.v = var5 + var5 - 1 - this.v;
                  this.n = -this.n;
               }
            } else if (this.n < 0) {
               while (true) {
                  var9 = this.b(var1, var9, var5, var3, var4.l[this.t - 1]);
                  if (this.v >= var5) {
                     return;
                  }

                  this.v = var6 - 1 - (var6 - 1 - this.v) % var8;
               }
            } else {
               while (true) {
                  var9 = this.a(var1, var9, var6, var3, var4.l[this.x]);
                  if (this.v < var6) {
                     return;
                  }

                  this.v = var5 + (this.v - var5) % var8;
               }
            }
         } else {
            if (this.l > 0) {
               label132:
               if (this.u) {
                  if (this.n < 0) {
                     var9 = this.b(var1, var2, var5, var3, var4.l[this.x]);
                     if (this.v >= var5) {
                        return;
                     }

                     this.v = var5 + var5 - 1 - this.v;
                     this.n = -this.n;
                     if (--this.l == 0) {
                        break label132;
                     }
                  }

                  do {
                     var9 = this.a(var1, var9, var6, var3, var4.l[this.t - 1]);
                     if (this.v < var6) {
                        return;
                     }

                     this.v = var6 + var6 - 1 - this.v;
                     this.n = -this.n;
                     if (--this.l == 0) {
                        break;
                     }

                     var9 = this.b(var1, var9, var5, var3, var4.l[this.x]);
                     if (this.v >= var5) {
                        return;
                     }

                     this.v = var5 + var5 - 1 - this.v;
                     this.n = -this.n;
                  } while (--this.l != 0);
               } else if (this.n < 0) {
                  while (true) {
                     var9 = this.b(var1, var9, var5, var3, var4.l[this.t - 1]);
                     if (this.v >= var5) {
                        return;
                     }

                     int var13 = (var6 - 1 - this.v) / var8;
                     if (var13 >= this.l) {
                        this.v = this.v + var8 * this.l;
                        this.l = 0;
                        break;
                     }

                     this.v += var8 * var13;
                     this.l -= var13;
                  }
               } else {
                  while (true) {
                     var9 = this.a(var1, var9, var6, var3, var4.l[this.x]);
                     if (this.v < var6) {
                        return;
                     }

                     int var10 = (this.v - var5) / var8;
                     if (var10 >= this.l) {
                        this.v = this.v - var8 * this.l;
                        this.l = 0;
                        break;
                     }

                     this.v -= var8 * var10;
                     this.l -= var10;
                  }
               }
            }

            if (this.n < 0) {
               this.b(var1, var9, 0, var3, 0);
               if (this.v < 0) {
                  this.v = -1;
                  this.g();
                  this.a(-27331);
               }
            } else {
               this.a(var1, var9, var7, var3, 0);
               if (this.v >= var7) {
                  this.v = var7;
                  this.g();
                  this.a(-27331);
               }
            }
         }
      }
   }

   static final sb a(vb var0, int var1, int var2) {
      return var0.l != null && var0.l.length != 0 ? new sb(var0, (int)((long)var0.i * 256L * (long)var1 / (long)(100 * sa.t)), var2 << 6) : null;
   }

   private static final int a(
      int var0, int var1, byte[] var2, int[] var3, int var4, int var5, int var6, int var7, int var8, int var9, int var10, sb var11, int var12, int var13
   ) {
      if (var12 == 0 || (var8 = var5 + (var10 - var4 + var12 - 257) / var12) > var9) {
         var8 = var9;
      }

      var5 <<= 1;

      for (int var24 = var8 << 1; var5 < var24; var4 += var12) {
         var1 = var4 >> 8;
         int var14 = var2[var1];
         var14 = (var14 << 8) + (var2[var1 + 1] - var14) * (var4 & 0xFF);
         var3[var5++] += var14 * var6 >> 6;
         var3[var5++] += var14 * var7 >> 6;
      }

      if (var12 == 0 || (var8 = (var5 >> 1) + (var10 - var4 + var12 - 1) / var12) > var9) {
         var8 = var9;
      }

      var8 <<= 1;

      for (int var19 = var13; var5 < var8; var4 += var12) {
         int var16 = var2[var4 >> 8];
         var16 = (var16 << 8) + (var19 - var16) * (var4 & 0xFF);
         var3[var5++] += var16 * var6 >> 6;
         var3[var5++] += var16 * var7 >> 6;
      }

      var11.v = var4;
      return var5 >> 1;
   }

   private final boolean f() {
      int var1 = this.r;
      int var2;
      int var3;
      if (var1 == Integer.MIN_VALUE) {
         var3 = 0;
         var2 = 0;
         var1 = 0;
      } else {
         var2 = b(var1, this.q);
         var3 = c(var1, this.q);
      }

      if (this.w == var1 && this.k == var2 && this.m == var3) {
         if (this.r == Integer.MIN_VALUE) {
            this.r = 0;
            this.m = 0;
            this.k = 0;
            this.w = 0;
            this.a(-27331);
            return true;
         } else {
            this.e();
            return false;
         }
      } else {
         if (this.w < var1) {
            this.s = 1;
            this.y = var1 - this.w;
         } else if (this.w > var1) {
            this.s = -1;
            this.y = this.w - var1;
         } else {
            this.s = 0;
         }

         if (this.k < var2) {
            this.p = 1;
            if (this.y == 0 || this.y > var2 - this.k) {
               this.y = var2 - this.k;
            }
         } else if (this.k > var2) {
            this.p = -1;
            if (this.y == 0 || this.y > this.k - var2) {
               this.y = this.k - var2;
            }
         } else {
            this.p = 0;
         }

         if (this.m < var3) {
            this.o = 1;
            if (this.y == 0 || this.y > var3 - this.m) {
               this.y = var3 - this.m;
            }
         } else if (this.m > var3) {
            this.o = -1;
            if (this.y == 0 || this.y > this.m - var3) {
               this.y = this.m - var3;
            }
         } else {
            this.o = 0;
         }

         return false;
      }
   }

   private static final int a(int var0, byte[] var1, int[] var2, int var3, int var4, int var5, int var6, int var7, int var8, int var9, sb var10) {
      var3 >>= 8;
      var9 >>= 8;
      var5 <<= 2;
      var6 <<= 2;
      if ((var7 = var4 + var9 - var3) > var8) {
         var7 = var8;
      }

      var4 <<= 1;
      var7 <<= 1;
      var7 -= 6;

      while (var4 < var7) {
         byte var11 = var1[var3++];
         var2[var4++] += var11 * var5;
         var2[var4++] += var11 * var6;
         var11 = var1[var3++];
         var2[var4++] += var11 * var5;
         var2[var4++] += var11 * var6;
         var11 = var1[var3++];
         var2[var4++] += var11 * var5;
         var2[var4++] += var11 * var6;
         var11 = var1[var3++];
         var2[var4++] += var11 * var5;
         var2[var4++] += var11 * var6;
      }

      var7 += 6;

      while (var4 < var7) {
         byte var15 = var1[var3++];
         var2[var4++] += var15 * var5;
         var2[var4++] += var15 * var6;
      }

      var10.v = var3 << 8;
      return var4 >> 1;
   }

   private final int b(int[] var1, int var2, int var3, int var4, int var5) {
      while (this.y > 0) {
         int var6 = var2 + this.y;
         if (var6 > var4) {
            var6 = var4;
         }

         this.y += var2;
         if (this.n == -256 && (this.v & 0xFF) == 0) {
            if (sa.i) {
               var2 = b(0, ((vb)this.h).l, var1, this.v, var2, this.k, this.m, this.p, this.o, 0, var6, var3, this);
            } else {
               var2 = b(((vb)this.h).l, var1, this.v, var2, this.w, this.s, 0, var6, var3, this);
            }
         } else if (sa.i) {
            var2 = b(0, 0, ((vb)this.h).l, var1, this.v, var2, this.k, this.m, this.p, this.o, 0, var6, var3, this, this.n, var5);
         } else {
            var2 = c(0, 0, ((vb)this.h).l, var1, this.v, var2, this.w, this.s, 0, var6, var3, this, this.n, var5);
         }

         this.y -= var2;
         if (this.y != 0) {
            return var2;
         }

         if (this.f()) {
            return var4;
         }
      }

      if (this.n == -256 && (this.v & 0xFF) == 0) {
         return sa.i
            ? b(0, ((vb)this.h).l, var1, this.v, var2, this.k, this.m, 0, var4, var3, this)
            : a(((vb)this.h).l, var1, this.v, var2, this.w, 0, var4, var3, this);
      } else {
         return sa.i
            ? d(0, 0, ((vb)this.h).l, var1, this.v, var2, this.k, this.m, 0, var4, var3, this, this.n, var5)
            : b(0, 0, ((vb)this.h).l, var1, this.v, var2, this.w, 0, var4, var3, this, this.n, var5);
      }
   }

   private static final int c(
      int var0, int var1, byte[] var2, int[] var3, int var4, int var5, int var6, int var7, int var8, int var9, int var10, sb var11, int var12, int var13
   ) {
      var11.k = var11.k - var11.p * var5;
      var11.m = var11.m - var11.o * var5;
      if (var12 == 0 || (var8 = var5 + (var10 + 256 - var4 + var12) / var12) > var9) {
         var8 = var9;
      }

      while (var5 < var8) {
         var1 = var4 >> 8;
         byte var14 = var2[var1 - 1];
         var3[var5++] += ((var14 << 8) + (var2[var1] - var14) * (var4 & 0xFF)) * var6 >> 6;
         var6 += var7;
         var4 += var12;
      }

      if (var12 == 0 || (var8 = var5 + (var10 - var4 + var12) / var12) > var9) {
         var8 = var9;
      }

      var0 = var13;

      for (int var17 = var12; var5 < var8; var4 += var17) {
         var3[var5++] += ((var0 << 8) + (var2[var4 >> 8] - var0) * (var4 & 0xFF)) * var6 >> 6;
         var6 += var7;
      }

      var11.k = var11.k + var11.p * var5;
      var11.m = var11.m + var11.o * var5;
      var11.w = var6;
      var11.v = var4;
      return var5;
   }

   private static final int d(
      int var0, int var1, byte[] var2, int[] var3, int var4, int var5, int var6, int var7, int var8, int var9, int var10, sb var11, int var12, int var13
   ) {
      if (var12 == 0 || (var8 = var5 + (var10 + 256 - var4 + var12) / var12) > var9) {
         var8 = var9;
      }

      var5 <<= 1;

      for (int var23 = var8 << 1; var5 < var23; var4 += var12) {
         var1 = var4 >> 8;
         int var14 = var2[var1 - 1];
         var14 = (var14 << 8) + (var2[var1] - var14) * (var4 & 0xFF);
         var3[var5++] += var14 * var6 >> 6;
         var3[var5++] += var14 * var7 >> 6;
      }

      if (var12 == 0 || (var8 = (var5 >> 1) + (var10 - var4 + var12) / var12) > var9) {
         var8 = var9;
      }

      var8 <<= 1;

      for (int var18 = var13; var5 < var8; var4 += var12) {
         var0 = (var18 << 8) + (var2[var4 >> 8] - var18) * (var4 & 0xFF);
         var3[var5++] += var0 * var6 >> 6;
         var3[var5++] += var0 * var7 >> 6;
      }

      var11.v = var4;
      return var5 >> 1;
   }

   private final void e() {
      this.w = this.r;
      this.k = b(this.r, this.q);
      this.m = c(this.r, this.q);
   }

   private static final int a(
      int var0,
      int var1,
      byte[] var2,
      int[] var3,
      int var4,
      int var5,
      int var6,
      int var7,
      int var8,
      int var9,
      int var10,
      int var11,
      int var12,
      sb var13,
      int var14,
      int var15
   ) {
      var13.w = var13.w - var13.s * var5;
      if (var14 == 0 || (var10 = var5 + (var12 - var4 + var14 - 257) / var14) > var11) {
         var10 = var11;
      }

      var5 <<= 1;

      for (int var27 = var10 << 1; var5 < var27; var4 += var14) {
         var1 = var4 >> 8;
         int var16 = var2[var1];
         var16 = (var16 << 8) + (var2[var1 + 1] - var16) * (var4 & 0xFF);
         var3[var5++] += var16 * var6 >> 6;
         var6 += var8;
         var3[var5++] += var16 * var7 >> 6;
         var7 += var9;
      }

      if (var14 == 0 || (var10 = (var5 >> 1) + (var12 - var4 + var14 - 1) / var14) > var11) {
         var10 = var11;
      }

      var10 <<= 1;

      for (int var21 = var15; var5 < var10; var4 += var14) {
         int var18 = var2[var4 >> 8];
         var18 = (var18 << 8) + (var21 - var18) * (var4 & 0xFF);
         var3[var5++] += var18 * var6 >> 6;
         var6 += var8;
         var3[var5++] += var18 * var7 >> 6;
         var7 += var9;
      }

      var5 >>= 1;
      var13.w = var13.w + var13.s * var5;
      var13.k = var6;
      var13.m = var7;
      var13.v = var4;
      return var5;
   }

   @Override
   final va b() {
      return null;
   }

   private static final int a(byte[] var0, int[] var1, int var2, int var3, int var4, int var5, int var6, int var7, sb var8) {
      var2 >>= 8;
      var7 >>= 8;
      var4 <<= 2;
      if ((var5 = var3 + var2 - (var7 - 1)) > var6) {
         var5 = var6;
      }

      var5 -= 3;

      while (var3 < var5) {
         var1[var3++] += var0[var2--] * var4;
         var1[var3++] += var0[var2--] * var4;
         var1[var3++] += var0[var2--] * var4;
         var1[var3++] += var0[var2--] * var4;
      }

      var5 += 3;

      while (var3 < var5) {
         var1[var3++] += var0[var2--] * var4;
      }

      var8.v = var2 << 8;
      return var3;
   }

   private static final int a(
      int var0, byte[] var1, int[] var2, int var3, int var4, int var5, int var6, int var7, int var8, int var9, int var10, int var11, sb var12
   ) {
      var3 >>= 8;
      var11 >>= 8;
      var5 <<= 2;
      var6 <<= 2;
      var7 <<= 2;
      var8 <<= 2;
      if ((var9 = var4 + var11 - var3) > var10) {
         var9 = var10;
      }

      var12.w = var12.w + var12.s * (var9 - var4);
      var4 <<= 1;
      var9 <<= 1;
      var9 -= 6;

      while (var4 < var9) {
         byte var13 = var1[var3++];
         var2[var4++] += var13 * var5;
         var5 += var7;
         var2[var4++] += var13 * var6;
         var6 += var8;
         var13 = var1[var3++];
         var2[var4++] += var13 * var5;
         var5 += var7;
         var2[var4++] += var13 * var6;
         var6 += var8;
         var13 = var1[var3++];
         var2[var4++] += var13 * var5;
         var5 += var7;
         var2[var4++] += var13 * var6;
         var6 += var8;
         var13 = var1[var3++];
         var2[var4++] += var13 * var5;
         var5 += var7;
         var2[var4++] += var13 * var6;
         var6 += var8;
      }

      for (int var44 = var9 + 6; var4 < var44; var6 += var8) {
         byte var17 = var1[var3++];
         var2[var4++] += var17 * var5;
         var5 += var7;
         var2[var4++] += var17 * var6;
      }

      var12.k = var5 >> 2;
      var12.m = var6 >> 2;
      var12.v = var3 << 8;
      return var4 >> 1;
   }

   private static final int b(int var0, byte[] var1, int[] var2, int var3, int var4, int var5, int var6, int var7, int var8, int var9, sb var10) {
      var3 >>= 8;
      var9 >>= 8;
      var5 <<= 2;
      var6 <<= 2;
      if ((var7 = var4 + var3 - (var9 - 1)) > var8) {
         var7 = var8;
      }

      var4 <<= 1;
      var7 <<= 1;
      var7 -= 6;

      while (var4 < var7) {
         byte var11 = var1[var3--];
         var2[var4++] += var11 * var5;
         var2[var4++] += var11 * var6;
         var11 = var1[var3--];
         var2[var4++] += var11 * var5;
         var2[var4++] += var11 * var6;
         var11 = var1[var3--];
         var2[var4++] += var11 * var5;
         var2[var4++] += var11 * var6;
         var11 = var1[var3--];
         var2[var4++] += var11 * var5;
         var2[var4++] += var11 * var6;
      }

      var7 += 6;

      while (var4 < var7) {
         byte var15 = var1[var3--];
         var2[var4++] += var15 * var5;
         var2[var4++] += var15 * var6;
      }

      var10.v = var3 << 8;
      return var4 >> 1;
   }

   private static final int b(byte[] var0, int[] var1, int var2, int var3, int var4, int var5, int var6, int var7, sb var8) {
      var2 >>= 8;
      var7 >>= 8;
      var4 <<= 2;
      if ((var5 = var3 + var7 - var2) > var6) {
         var5 = var6;
      }

      var5 -= 3;

      while (var3 < var5) {
         var1[var3++] += var0[var2++] * var4;
         var1[var3++] += var0[var2++] * var4;
         var1[var3++] += var0[var2++] * var4;
         var1[var3++] += var0[var2++] * var4;
      }

      var5 += 3;

      while (var3 < var5) {
         var1[var3++] += var0[var2++] * var4;
      }

      var8.v = var2 << 8;
      return var3;
   }

   private static final int b(
      int var0,
      int var1,
      byte[] var2,
      int[] var3,
      int var4,
      int var5,
      int var6,
      int var7,
      int var8,
      int var9,
      int var10,
      int var11,
      int var12,
      sb var13,
      int var14,
      int var15
   ) {
      var13.w = var13.w - var13.s * var5;
      if (var14 == 0 || (var10 = var5 + (var12 + 256 - var4 + var14) / var14) > var11) {
         var10 = var11;
      }

      var5 <<= 1;

      for (int var26 = var10 << 1; var5 < var26; var4 += var14) {
         var1 = var4 >> 8;
         int var16 = var2[var1 - 1];
         var16 = (var16 << 8) + (var2[var1] - var16) * (var4 & 0xFF);
         var3[var5++] += var16 * var6 >> 6;
         var6 += var8;
         var3[var5++] += var16 * var7 >> 6;
         var7 += var9;
      }

      if (var14 == 0 || (var10 = (var5 >> 1) + (var12 - var4 + var14) / var14) > var11) {
         var10 = var11;
      }

      var10 <<= 1;

      for (int var20 = var15; var5 < var10; var4 += var14) {
         var0 = (var20 << 8) + (var2[var4 >> 8] - var20) * (var4 & 0xFF);
         var3[var5++] += var0 * var6 >> 6;
         var6 += var8;
         var3[var5++] += var0 * var7 >> 6;
         var7 += var9;
      }

      var5 >>= 1;
      var13.w = var13.w + var13.s * var5;
      var13.k = var6;
      var13.m = var7;
      var13.v = var4;
      return var5;
   }

   private static final int b(
      int var0, byte[] var1, int[] var2, int var3, int var4, int var5, int var6, int var7, int var8, int var9, int var10, int var11, sb var12
   ) {
      var3 >>= 8;
      var11 >>= 8;
      var5 <<= 2;
      var6 <<= 2;
      var7 <<= 2;
      var8 <<= 2;
      if ((var9 = var4 + var3 - (var11 - 1)) > var10) {
         var9 = var10;
      }

      var12.w = var12.w + var12.s * (var9 - var4);
      var4 <<= 1;
      var9 <<= 1;
      var9 -= 6;

      while (var4 < var9) {
         byte var13 = var1[var3--];
         var2[var4++] += var13 * var5;
         var5 += var7;
         var2[var4++] += var13 * var6;
         var6 += var8;
         var13 = var1[var3--];
         var2[var4++] += var13 * var5;
         var5 += var7;
         var2[var4++] += var13 * var6;
         var6 += var8;
         var13 = var1[var3--];
         var2[var4++] += var13 * var5;
         var5 += var7;
         var2[var4++] += var13 * var6;
         var6 += var8;
         var13 = var1[var3--];
         var2[var4++] += var13 * var5;
         var5 += var7;
         var2[var4++] += var13 * var6;
         var6 += var8;
      }

      for (int var44 = var9 + 6; var4 < var44; var6 += var8) {
         byte var17 = var1[var3--];
         var2[var4++] += var17 * var5;
         var5 += var7;
         var2[var4++] += var17 * var6;
      }

      var12.k = var5 >> 2;
      var12.m = var6 >> 2;
      var12.v = var3 << 8;
      return var4 >> 1;
   }

   private final void g() {
      if (this.y != 0) {
         if (this.r == Integer.MIN_VALUE) {
            this.r = 0;
         }

         this.y = 0;
         this.e();
      }
   }

   private static final int b(
      int var0, int var1, byte[] var2, int[] var3, int var4, int var5, int var6, int var7, int var8, int var9, sb var10, int var11, int var12
   ) {
      if (var11 == 0 || (var7 = var5 + (var9 + 256 - var4 + var11) / var11) > var8) {
         var7 = var8;
      }

      while (var5 < var7) {
         var1 = var4 >> 8;
         byte var13 = var2[var1 - 1];
         var3[var5++] += ((var13 << 8) + (var2[var1] - var13) * (var4 & 0xFF)) * var6 >> 6;
         var4 += var11;
      }

      if (var11 == 0 || (var7 = var5 + (var9 - var4 + var11) / var11) > var8) {
         var7 = var8;
      }

      var0 = var12;

      for (int var16 = var11; var5 < var7; var4 += var16) {
         var3[var5++] += ((var0 << 8) + (var2[var4 >> 8] - var0) * (var4 & 0xFF)) * var6 >> 6;
      }

      var10.v = var4;
      return var5;
   }

   private static final int c(int var0, int var1) {
      return var1 < 0 ? -var0 : (int)((double)var0 * Math.sqrt((double)var1 * 1.2207031E-4F) + 0.5);
   }

   private static final int a(byte[] var0, int[] var1, int var2, int var3, int var4, int var5, int var6, int var7, int var8, sb var9) {
      var2 >>= 8;
      var8 >>= 8;
      var4 <<= 2;
      var5 <<= 2;
      if ((var6 = var3 + var8 - var2) > var7) {
         var6 = var7;
      }

      var9.k = var9.k + var9.p * (var6 - var3);
      var9.m = var9.m + var9.o * (var6 - var3);
      var6 -= 3;

      while (var3 < var6) {
         var1[var3++] += var0[var2++] * var4;
         var4 += var5;
         var1[var3++] += var0[var2++] * var4;
         var4 += var5;
         var1[var3++] += var0[var2++] * var4;
         var4 += var5;
         var1[var3++] += var0[var2++] * var4;
         var4 += var5;
      }

      for (int var24 = var6 + 3; var3 < var24; var4 += var5) {
         var1[var3++] += var0[var2++] * var4;
      }

      var9.w = var4 >> 2;
      var9.v = var2 << 8;
      return var3;
   }

   private static final int a(
      int var0, int var1, byte[] var2, int[] var3, int var4, int var5, int var6, int var7, int var8, int var9, sb var10, int var11, int var12
   ) {
      if (var11 == 0 || (var7 = var5 + (var9 - var4 + var11 - 257) / var11) > var8) {
         var7 = var8;
      }

      while (var5 < var7) {
         var1 = var4 >> 8;
         byte var13 = var2[var1];
         var3[var5++] += ((var13 << 8) + (var2[var1 + 1] - var13) * (var4 & 0xFF)) * var6 >> 6;
         var4 += var11;
      }

      if (var11 == 0 || (var7 = var5 + (var9 - var4 + var11 - 1) / var11) > var8) {
         var7 = var8;
      }

      for (int var16 = var12; var5 < var7; var4 += var11) {
         byte var14 = var2[var4 >> 8];
         var3[var5++] += ((var14 << 8) + (var16 - var14) * (var4 & 0xFF)) * var6 >> 6;
      }

      var10.v = var4;
      return var5;
   }

   @Override
   final va a() {
      return null;
   }

   @Override
   final synchronized void b(int var1) {
      if (this.y > 0) {
         if (var1 >= this.y) {
            if (this.r == Integer.MIN_VALUE) {
               this.r = 0;
               this.m = 0;
               this.k = 0;
               this.w = 0;
               this.a(-27331);
               var1 = this.y;
            }

            this.y = 0;
            this.e();
         } else {
            this.w = this.w + this.s * var1;
            this.k = this.k + this.p * var1;
            this.m = this.m + this.o * var1;
            this.y -= var1;
         }
      }

      vb var2 = (vb)this.h;
      int var3 = this.x << 8;
      int var4 = this.t << 8;
      int var5 = var2.l.length << 8;
      int var6 = var4 - var3;
      if (var6 <= 0) {
         this.l = 0;
      }

      if (this.v < 0) {
         if (this.n <= 0) {
            this.g();
            this.a(-27331);
            return;
         }

         this.v = 0;
      }

      if (this.v >= var5) {
         if (this.n >= 0) {
            this.g();
            this.a(-27331);
            return;
         }

         this.v = var5 - 1;
      }

      this.v = this.v + this.n * var1;
      if (this.l < 0) {
         if (!this.u) {
            if (this.n < 0) {
               if (this.v >= var3) {
                  return;
               }

               this.v = var4 - 1 - (var4 - 1 - this.v) % var6;
            } else {
               if (this.v < var4) {
                  return;
               }

               this.v = var3 + (this.v - var3) % var6;
            }
         } else {
            if (this.n < 0) {
               if (this.v >= var3) {
                  return;
               }

               this.v = var3 + var3 - 1 - this.v;
               this.n = -this.n;
            }

            while (this.v >= var4) {
               this.v = var4 + var4 - 1 - this.v;
               this.n = -this.n;
               if (this.v >= var3) {
                  return;
               }

               this.v = var3 + var3 - 1 - this.v;
               this.n = -this.n;
            }
         }
      } else {
         if (this.l > 0) {
            label131:
            if (this.u) {
               if (this.n < 0) {
                  if (this.v >= var3) {
                     return;
                  }

                  this.v = var3 + var3 - 1 - this.v;
                  this.n = -this.n;
                  if (--this.l == 0) {
                     break label131;
                  }
               }

               do {
                  if (this.v < var4) {
                     return;
                  }

                  this.v = var4 + var4 - 1 - this.v;
                  this.n = -this.n;
                  if (--this.l == 0) {
                     break;
                  }

                  if (this.v >= var3) {
                     return;
                  }

                  this.v = var3 + var3 - 1 - this.v;
                  this.n = -this.n;
               } while (--this.l != 0);
            } else if (this.n < 0) {
               if (this.v >= var3) {
                  return;
               }

               int var7 = (var4 - 1 - this.v) / var6;
               if (var7 < this.l) {
                  this.v += var6 * var7;
                  this.l -= var7;
                  return;
               }

               this.v = this.v + var6 * this.l;
               this.l = 0;
            } else {
               if (this.v < var4) {
                  return;
               }

               int var8 = (this.v - var3) / var6;
               if (var8 < this.l) {
                  this.v -= var6 * var8;
                  this.l -= var8;
                  return;
               }

               this.v = this.v - var6 * this.l;
               this.l = 0;
            }
         }

         if (this.n < 0) {
            if (this.v < 0) {
               this.v = -1;
               this.g();
               this.a(-27331);
            }
         } else if (this.v >= var5) {
            this.v = var5;
            this.g();
            this.a(-27331);
         }
      }
   }

   private static final int b(int var0, int var1) {
      return var1 < 0 ? var0 : (int)((double)var0 * Math.sqrt((double)(16384 - var1) * 1.2207031E-4F) + 0.5);
   }

   private final int a(int[] var1, int var2, int var3, int var4, int var5) {
      while (this.y > 0) {
         int var6 = var2 + this.y;
         if (var6 > var4) {
            var6 = var4;
         }

         this.y += var2;
         if (this.n == 256 && (this.v & 0xFF) == 0) {
            if (sa.i) {
               var2 = a(0, ((vb)this.h).l, var1, this.v, var2, this.k, this.m, this.p, this.o, 0, var6, var3, this);
            } else {
               var2 = a(((vb)this.h).l, var1, this.v, var2, this.w, this.s, 0, var6, var3, this);
            }
         } else if (sa.i) {
            var2 = a(0, 0, ((vb)this.h).l, var1, this.v, var2, this.k, this.m, this.p, this.o, 0, var6, var3, this, this.n, var5);
         } else {
            var2 = b(0, 0, ((vb)this.h).l, var1, this.v, var2, this.w, this.s, 0, var6, var3, this, this.n, var5);
         }

         this.y -= var2;
         if (this.y != 0) {
            return var2;
         }

         if (this.f()) {
            return var4;
         }
      }

      if (this.n == 256 && (this.v & 0xFF) == 0) {
         return sa.i
            ? a(0, ((vb)this.h).l, var1, this.v, var2, this.k, this.m, 0, var4, var3, this)
            : b(((vb)this.h).l, var1, this.v, var2, this.w, 0, var4, var3, this);
      } else {
         return sa.i
            ? a(0, 0, ((vb)this.h).l, var1, this.v, var2, this.k, this.m, 0, var4, var3, this, this.n, var5)
            : a(0, 0, ((vb)this.h).l, var1, this.v, var2, this.w, 0, var4, var3, this, this.n, var5);
      }
   }

   @Override
   final int d() {
      return this.r == 0 && this.y == 0 ? 0 : 1;
   }

   private sb(vb var1, int var2, int var3) {
      this.h = var1;
      this.x = var1.h;
      this.t = var1.k;
      this.u = var1.j;
      this.n = var2;
      this.r = var3;
      this.q = 8192;
      this.v = 0;
      this.e();
   }

   private static final int b(
      int var0, int var1, byte[] var2, int[] var3, int var4, int var5, int var6, int var7, int var8, int var9, int var10, sb var11, int var12, int var13
   ) {
      var11.k = var11.k - var11.p * var5;
      var11.m = var11.m - var11.o * var5;
      if (var12 == 0 || (var8 = var5 + (var10 - var4 + var12 - 257) / var12) > var9) {
         var8 = var9;
      }

      while (var5 < var8) {
         var1 = var4 >> 8;
         byte var14 = var2[var1];
         var3[var5++] += ((var14 << 8) + (var2[var1 + 1] - var14) * (var4 & 0xFF)) * var6 >> 6;
         var6 += var7;
         var4 += var12;
      }

      if (var12 == 0 || (var8 = var5 + (var10 - var4 + var12 - 1) / var12) > var9) {
         var8 = var9;
      }

      for (int var17 = var13; var5 < var8; var4 += var12) {
         byte var15 = var2[var4 >> 8];
         var3[var5++] += ((var15 << 8) + (var17 - var15) * (var4 & 0xFF)) * var6 >> 6;
         var6 += var7;
      }

      var11.k = var11.k + var11.p * var5;
      var11.m = var11.m + var11.o * var5;
      var11.w = var6;
      var11.v = var4;
      return var5;
   }
}
