import java.awt.Component;

class sa {
   static int t;
   private static eb n;
   private long u;
   int[] j;
   private static int o;
   static boolean i;
   private boolean b = false;
   private int r = 32;
   private va a;
   private int g;
   private int q;
   private int c;
   private int k;
   private va[] s;
   private long e;
   private int p;
   private int m;
   private va[] d;
   private long f;
   private int h;
   private boolean l;

   int b() throws Exception {
      return this.k;
   }

   void b(int var1) throws Exception {
   }

   void c() throws Exception {
   }

   final synchronized void d() {
      if (n != null) {
         boolean var1 = true;

         for (int var2 = 0; var2 < 2; var2++) {
            if (n.f[var2] == this) {
               n.f[var2] = null;
            }

            if (n.f[var2] != null) {
               var1 = false;
            }
         }

         if (var1) {
            n.a = true;

            while (n.i) {
               mb.a(11200, 50L);
            }

            n = null;
         }
      }

      this.e();
      this.j = null;
      this.b = true;
   }

   void e() {
   }

   private final void a(va var1, int var2) {
      int var3 = var2 >> 5;
      va var4 = this.s[var3];
      if (var4 == null) {
         this.d[var3] = var1;
      } else {
         var4.j = var1;
      }

      this.s[var3] = var1;
      var1.i = var2;
   }

   static final sa a(c var0, Component var1, int var2, int var3) {
      if (t == 0) {
         throw new IllegalStateException();
      }

      if (var2 >= 0 && var2 < 2) {
         if (var3 < 256) {
            var3 = 256;
         }

         try {
            pb var4 = new pb();
            var4.j = new int[256 * (i ? 2 : 1)];
            var4.q = var3;
            var4.a(var1);
            var4.k = (var3 & -1024) + 1024;
            if (var4.k > 16384) {
               var4.k = 16384;
            }

            var4.b(var4.k);
            if (o > 0 && n == null) {
               n = new eb();
               n.g = var0;
               var0.a(true, n, o);
            }

            if (n != null) {
               if (n.f[var2] != null) {
                  throw new IllegalArgumentException();
               }

               n.f[var2] = var4;
            }

            return var4;
         } catch (Throwable var5) {
            return new sa();
         }
      } else {
         throw new IllegalArgumentException();
      }
   }

   final synchronized void a() {
      if (!this.b) {
         long var1 = p.a(0);

         try {
            if (var1 > this.u + 6000L) {
               this.u = var1 - 6000L;
            }

            while (var1 > this.u + 5000L) {
               this.a(256);
               this.u = this.u + 256000 / t;
               var1 = p.a(0);
            }
         } catch (Exception var6) {
            this.u = var1;
         }

         if (this.j != null) {
            try {
               if (this.e != 0L) {
                  if (var1 < this.e) {
                     return;
                  }

                  this.b(this.k);
                  this.e = 0L;
                  this.l = true;
               }

               int var3 = this.b();
               if (this.h - var3 > this.m) {
                  this.m = this.h - var3;
               }

               int var4 = this.q + this.g;
               if (var4 + 256 > 16384) {
                  var4 = 16128;
               }

               if (var4 + 256 > this.k) {
                  this.k += 1024;
                  if (this.k > 16384) {
                     this.k = 16384;
                  }

                  this.e();
                  this.b(this.k);
                  var3 = 0;
                  this.l = true;
                  if (var4 + 256 > this.k) {
                     var4 = this.k - 256;
                     this.g = var4 - this.q;
                  }
               }

               while (var3 < var4) {
                  this.a(this.j, 256);
                  this.c();
                  var3 += 256;
               }

               if (var1 > this.f) {
                  if (!this.l) {
                     if (this.m == 0 && this.c == 0) {
                        this.e();
                        this.e = var1 + 2000L;
                        return;
                     }

                     this.g = Math.min(this.c, this.m);
                     this.c = this.m;
                  } else {
                     this.l = false;
                  }

                  this.m = 0;
                  this.f = var1 + 2000L;
               }

               this.h = var3;
            } catch (Exception var5) {
               this.e();
               this.e = var1 + 2000L;
            }
         }
      }
   }

   final synchronized void a(va var1) {
      this.a = var1;
   }

   static final void a(int var0, boolean var1, int var2) {
      if (var0 >= 8000 && var0 <= 48000) {
         t = var0;
         i = var1;
         o = var2;
      } else {
         throw new IllegalArgumentException();
      }
   }

   private final void a(int[] var1, int var2) {
      int var3 = var2;
      if (i) {
         var3 <<= 1;
      }

      ab.a(var1, 0, var3);
      this.p -= var2;
      if (this.a != null && this.p <= 0) {
         this.p = this.p + (t >> 4);
         b(this.a);
         this.a(this.a, this.a.c());
         int var4 = 0;
         int var5 = 255;

         label120:
         for (int var6 = 7; var5 != 0; var6--) {
            int var7;
            int var8;
            if (var6 < 0) {
               var7 = var6 & 3;
               var8 = -(var6 >> 2);
            } else {
               var7 = var6;
               var8 = 0;
            }

            for (int var9 = var5 >>> var7 & 286331153; var9 != 0; var9 >>>= 4) {
               if ((var9 & 1) != 0) {
                  var5 &= ~(1 << var7);
                  va var10 = null;
                  va var11 = this.d[var7];

                  while (var11 != null) {
                     bb var12 = var11.h;
                     if (var12 != null && var12.g > var8) {
                        var5 |= 1 << var7;
                        var10 = var11;
                        var11 = var11.j;
                     } else {
                        var11.g = true;
                        int var13 = var11.d();
                        var4 += var13;
                        if (var12 != null) {
                           var12.g += var13;
                        }

                        if (var4 >= this.r) {
                           break label120;
                        }

                        va var14 = var11.b();
                        if (var14 != null) {
                           int var15 = var11.i;

                           while (var14 != null) {
                              this.a(var14, var15 * var14.c() >> 8);
                              var14 = var11.a();
                           }
                        }

                        va var21 = var11.j;
                        var11.j = null;
                        if (var10 == null) {
                           this.d[var7] = var21;
                        } else {
                           var10.j = var21;
                        }

                        if (var21 == null) {
                           this.s[var7] = var10;
                        }

                        var11 = var21;
                     }
                  }
               }

               var7 += 4;
               var8++;
            }
         }

         for (int var16 = 0; var16 < 8; var16++) {
            va var17 = this.d[var16];
            va[] var18 = this.d;
            int var19 = var16;
            this.s[var16] = null;
            var18[var19] = null;

            while (var17 != null) {
               va var20 = var17.j;
               var17.j = null;
               var17 = var20;
            }
         }
      }

      if (this.p < 0) {
         this.p = 0;
      }

      if (this.a != null) {
         this.a.b(var1, 0, var2);
      }

      this.u = p.a(0);
   }

   private final void a(int var1) {
      this.p -= var1;
      if (this.p < 0) {
         this.p = 0;
      }

      if (this.a != null) {
         this.a.b(var1);
      }
   }

   private static final void b(va var0) {
      var0.g = false;
      if (var0.h != null) {
         var0.h.g = 0;
      }

      for (va var1 = var0.b(); var1 != null; var1 = var0.a()) {
         b(var1);
      }
   }

   void a(Component var1) throws Exception {
   }

   sa() {
      this.u = p.a(0);
      this.s = new va[8];
      this.c = 0;
      this.m = 0;
      this.e = 0L;
      this.p = 0;
      this.d = new va[8];
      this.f = 0L;
      this.h = 0;
      this.l = true;
   }
}
