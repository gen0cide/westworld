import java.io.BufferedInputStream;
import java.io.EOFException;
import java.io.FileInputStream;
import java.io.IOException;
import java.io.InputStream;
import java.net.URL;

final class nb {
   private byte[] n;
   static int m;
   private long h = -1L;
   private long l;
   static int g = 0;
   static int p;
   private byte[] a;
   private long x;
   static int e;
   private d b;
   static int[] d;
   private int t;
   static int k;
   private int i = 0;
   static int y;
   static int r;
   static String[] u = new String[]{z(z("wFj_\nWW\u007f\u001aDVRx_X\u0003P|\u001aCWZwI\nWP:IOOS:[DG\u001fjHOPL:_DWZh"))};
   static int z;
   static URL s = null;
   static int w;
   static char[] f = new char[]{
      '€',
      '\u0000',
      '‚',
      'ƒ',
      '„',
      '…',
      '†',
      '‡',
      'ˆ',
      '‰',
      'Š',
      '‹',
      'Œ',
      '\u0000',
      'Ž',
      '\u0000',
      '\u0000',
      '‘',
      '’',
      '“',
      '”',
      '•',
      '–',
      '—',
      '˜',
      '™',
      'š',
      '›',
      'œ',
      '\u0000',
      'ž',
      'Ÿ'
   };
   private long o;
   private long A;
   private long j = -1L;
   static int c;
   static int q = 0;
   static int v;
   private static final String[] B = new String[]{
      z(z("M]4s\u0002")),
      z(z("MJvV")),
      z(z("X\u00114\u0014W")),
      z(z("M]4\u0006CMVn\u0004\u0002")),
      z(z("M]4~\u0002")),
      z(z("M]4|\u0002")),
      z(z("M]4r\u0002")),
      z(z("M]4x\u0002")),
      z(z("M]4}\u0002")),
      z(z("M]4{\u0002")),
      z(z("M]4\u007f\u0002")),
      z(z("M]4y\u0002"))
   };

   private final void a(int var1) throws IOException {
      boolean var7 = client.vh;

      try {
         if (this.h != -1L) {
            if (~this.x != ~this.h) {
               this.b.a(0, this.h);
               this.x = this.h;
            }

            this.b.b(this.a, this.i, 1, 0);
            this.x = this.x + this.i;
            if (~this.A > ~this.x) {
               this.A = this.x;
            }

            long var2;
            long var4;
            label90: {
               var2 = -1L;
               var4 = -1L;
               if (this.h < this.j || ~this.h <= ~(this.j + this.t)) {
                  if (~this.j > ~this.h || this.j >= this.i + this.h) {
                     break label90;
                  }

                  var2 = this.j;
                  if (!var7) {
                     break label90;
                  }
               }

               var2 = this.h;
            }

            label65: {
               if (~this.j > ~(this.h + this.i) && ~(this.i + this.h) >= ~(this.j + this.t)) {
                  var4 = this.h + this.i;
                  if (!var7) {
                     break label65;
                  }
               }

               if (this.t + this.j > this.h && this.j - -this.t <= this.i + this.h) {
                  var4 = this.j - -this.t;
               }
            }

            if (-1L < var2 && ~var2 > ~var4) {
               int var6 = (int)(var4 + -var2);
               ab.a(this.a, (int)(var2 + -this.h), this.n, (int)(-this.j + var2), var6);
            }

            this.i = 0;
            this.h = -1L;
         }

         v++;
         if (var1 != -14779) {
            this.n = (byte[])null;
         }
      } catch (RuntimeException var8) {
         throw i.a(var8, B[9] + var1 + ')');
      }
   }

   final void a(boolean var1, int var2, int var3, byte[] var4) throws IOException {
      boolean var14 = client.vh;

      try {
         z++;

         byte var10000;
         int var10001;
         label194: {
            try {
               if (~(var2 + var3) < ~var4.length) {
                  throw new ArrayIndexOutOfBoundsException(var3 - -var2 - var4.length);
               }

               if (this.h != -1L && this.h <= this.l && this.h - -this.i >= var2 + this.l) {
                  ab.a(this.a, (int)(this.l - this.h), var4, var3, var2);
                  this.l += var2;
                  return;
               }

               long var5 = this.l;
               int var7 = var3;
               int var8 = var2;
               if (~this.l <= ~this.j && ~(this.j + this.t) < ~this.l) {
                  int var9 = (int)(this.j + -this.l + this.t);
                  if (~var2 > ~var9) {
                     var9 = var2;
                  }

                  ab.a(this.n, (int)(this.l - this.j), var4, var3, var9);
                  var2 -= var9;
                  var3 += var9;
                  this.l += var9;
               }

               if (!var1) {
                  return;
               }

               label201: {
                  if (var2 <= this.n.length) {
                     if (0 >= var2) {
                        break label201;
                     }

                     this.b((byte)34);
                     int var17 = var2;
                     if (~this.t > ~var17) {
                        var17 = this.t;
                     }

                     ab.a(this.n, 0, var4, var3, var17);
                     var3 += var17;
                     var2 -= var17;
                     this.l += var17;
                     if (!var14) {
                        break label201;
                     }
                  }

                  this.b.a(0, this.l);
                  this.x = this.l;

                  while (-1 > ~var2) {
                     int var18 = this.b.a(var4, var2, var3, -1);
                     var10000 = 0;
                     var10001 = ~var18;
                     if (var14) {
                        break label194;
                     }

                     if (0 == var10001 && !var14) {
                        break;
                     }

                     var3 += var18;
                     var2 -= var18;
                     this.l += var18;
                     this.x += var18;
                     if (var14) {
                        break;
                     }
                  }
               }

               label190:
               if (this.h != -1L) {
                  if (~this.l > ~this.h && 0 < var2) {
                     int var19 = (int)(this.h - this.l) + var3;
                     if (var19 > var3 - -var2) {
                        var19 = var3 - -var2;
                     }

                     while (~var19 < ~var3) {
                        var4[var3++] = 0;
                        var2--;
                        this.l++;
                        if (var14) {
                           break label190;
                        }

                        if (var14) {
                           break;
                        }
                     }
                  }

                  long var20;
                  label158: {
                     var20 = -1L;
                     if (~var5 >= ~this.h && ~(var5 + var8) < ~this.h) {
                        var20 = this.h;
                        if (!var14) {
                           break label158;
                        }
                     }

                     if (~var5 <= ~this.h && var5 < this.h + this.i) {
                        var20 = var5;
                     }
                  }

                  long var11;
                  label210: {
                     var11 = -1L;
                     if (var5 >= this.h + this.i || this.h - -this.i > var5 + var8) {
                        if (var5 - -var8 <= this.h || ~(this.i + this.h) > ~(var8 + var5)) {
                           break label210;
                        }

                        var11 = var5 + var8;
                        if (!var14) {
                           break label210;
                        }
                     }

                     var11 = this.h + this.i;
                  }

                  if (var20 > -1L && ~var20 > ~var11) {
                     int var13 = (int)(var11 + -var20);
                     ab.a(this.a, (int)(-this.h + var20), var4, var7 - -((int)(var20 + -var5)), var13);
                     if (~this.l > ~var11) {
                        var2 = (int)(var2 - (var11 - this.l));
                        this.l = var11;
                     }
                  }
               }
            } catch (IOException var15) {
               this.x = -1L;
               throw var15;
            }

            var10000 = 0;
            var10001 = var2;
         }

         if (var10000 < var10001) {
            throw new EOFException();
         }
      } catch (RuntimeException var16) {
         throw i.a(var16, B[6] + var1 + ',' + var2 + ',' + var3 + ',' + (var4 != null ? B[2] : B[1]) + ')');
      }
   }

   final void a(byte var1, byte[] var2) throws IOException {
      try {
         int var3 = -12 % ((var1 - -22) / 54);
         this.a(true, var2.length, 0, var2);
         p++;
      } catch (RuntimeException var4) {
         throw i.a(var4, B[4] + var1 + ',' + (var2 != null ? B[2] : B[1]) + ')');
      }
   }

   final long a(byte var1) {
      try {
         if (var1 != -111) {
            d = (int[])null;
         }

         r++;
         return this.o;
      } catch (RuntimeException var3) {
         throw i.a(var3, B[0] + var1 + ')');
      }
   }

   static final int a(int var0, byte var1) {
      try {
         e++;
         if (var0 != 255) {
            a(-35, (byte)126);
         }

         return var1 & 0xFF;
      } catch (RuntimeException var3) {
         throw i.a(var3, B[8] + var0 + 44 + var1 + 41);
      }
   }

   static final InputStream a(boolean var0, String var1) throws IOException {
      try {
         if (!var0) {
            return (InputStream)null;
         }

         c++;
         if (null != s) {
            URL var3 = new URL(s, var1);
            InputStream var2 = var3.openStream();
            if (!client.vh) {
               return var2;
            }
         }

         return new BufferedInputStream(new FileInputStream(var1));
      } catch (RuntimeException var4) {
         throw i.a(var4, B[5] + var0 + ',' + (var1 != null ? B[2] : B[1]) + ')');
      }
   }

   final void a(int var1, int var2, int var3, byte[] var4) throws IOException {
      boolean var10 = client.vh;

      try {
         w++;
         if (var2 <= -80) {
            try {
               if (this.l + var1 > this.o) {
                  this.o = var1 + this.l;
               }

               if (this.h != -1L && (~this.l > ~this.h || this.i + this.h < this.l)) {
                  this.a(-14779);
               }

               if (0L != ~this.h && var1 + this.l > this.a.length + this.h) {
                  int var5 = (int)(-this.l - -this.h + this.a.length);
                  ab.a(var4, var3, this.a, (int)(-this.h + this.l), var5);
                  var1 -= var5;
                  var3 += var5;
                  this.l += var5;
                  this.i = this.a.length;
                  this.a(-14779);
               }

               if (~var1 >= ~this.a.length) {
                  if (var1 > 0) {
                     if (this.h == -1L) {
                        this.h = this.l;
                     }

                     ab.a(var4, var3, this.a, (int)(-this.h + this.l), var1);
                     this.l += var1;
                     if (~this.i > ~(this.l + -this.h)) {
                        this.i = (int)(this.l + -this.h);
                     }
                  }
               } else {
                  if (~this.x != ~this.l) {
                     this.b.a(0, this.l);
                     this.x = this.l;
                  }

                  this.b.b(var4, var1, 1, var3);
                  this.x += var1;
                  if (this.x > this.A) {
                     this.A = this.x;
                  }

                  long var7;
                  long var13;
                  label131: {
                     var13 = -1L;
                     var7 = -1L;
                     if (this.j <= this.l && this.l < this.j + this.t) {
                        var13 = this.l;
                        if (!var10) {
                           break label131;
                        }
                     }

                     if (this.l <= this.j && this.l - -var1 > this.j) {
                        var13 = this.j;
                     }
                  }

                  label125: {
                     if (this.j < this.l - -var1 && this.t + this.j >= this.l - -var1) {
                        var7 = this.l + var1;
                        if (!var10) {
                           break label125;
                        }
                     }

                     if (~this.l > ~(this.t + this.j) && ~(this.j - -this.t) >= ~(this.l - -var1)) {
                        var7 = this.j - -this.t;
                     }
                  }

                  if (0L > ~var13 && ~var7 < ~var13) {
                     int var9 = (int)(var7 - var13);
                     ab.a(var4, (int)(-this.l + var13 + var3), this.n, (int)(-this.j + var13), var9);
                  }

                  this.l += var1;
               }
            } catch (IOException var11) {
               this.x = -1L;
               throw var11;
            }
         }
      } catch (RuntimeException var12) {
         throw i.a(var12, B[10] + var1 + ',' + var2 + ',' + var3 + ',' + (var4 != null ? B[2] : B[1]) + ')');
      }
   }

   final void a(long var1, int var3) throws IOException {
      try {
         m++;
         if (~var1 > -1L) {
            throw new IOException();
         }

         this.l = var1;
         int var4 = -39 / ((var3 - -66) / 55);
      } catch (RuntimeException var5) {
         throw i.a(var5, B[11] + var1 + ',' + var3 + ')');
      }
   }

   private final void b(byte var1) throws IOException {
      boolean var4 = client.vh;

      try {
         this.t = 0;
         k++;
         if (~this.l != ~this.x) {
            this.b.a(0, this.l);
            this.x = this.l;
         }

         this.j = this.l;

         int var10000;
         int var10001;
         while (true) {
            if (~this.t > ~this.n.length) {
               int var2 = -this.t + this.n.length;
               var10000 = ~var2;
               var10001 = -200000001;
               if (var4) {
                  break;
               }

               if (var10000 < -200000001) {
                  var2 = 200000000;
               }

               int var3 = this.b.a(this.n, var2, this.t, -1);
               if (0 != ~var3 || var4) {
                  this.x += var3;
                  this.t += var3;
                  if (!var4) {
                     continue;
                  }
               }
            }

            var10000 = var1;
            var10001 = 14;
            break;
         }

         if (var10000 <= var10001) {
            this.b = (d)null;
         }
      } catch (RuntimeException var5) {
         throw i.a(var5, B[7] + var1 + ')');
      }
   }

   nb(d var1, int var2, int var3) throws IOException {
      try {
         this.b = var1;
         this.o = this.A = var1.a((byte)47);
         this.a = new byte[var3];
         this.l = 0L;
         this.n = new byte[var2];
      } catch (RuntimeException var5) {
         throw i.a(var5, B[3] + (var1 != null ? B[2] : B[1]) + ',' + var2 + ',' + var3 + ')');
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
               var10005 = 35;
               break;
            case 1:
               var10005 = 63;
               break;
            case 2:
               var10005 = 26;
               break;
            case 3:
               var10005 = 58;
               break;
            default:
               var10005 = 42;
         }

         var10001[var1] = (char)(var10004 ^ var10005);
      }

      return new String(var10001).intern();
   }
}
