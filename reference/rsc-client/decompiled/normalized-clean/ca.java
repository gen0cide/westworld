import java.io.DataInputStream;
import java.io.IOException;
import java.io.InputStream;

final class ca {
   private int f;
   private int ec;
   private int[] Eb;
   static int q;
   int[] H;
   private int[] jc;
   private int xb;
   static byte[][] tb = new byte[50][];
   private boolean c;
   static int W;
   private boolean v;
   int[] Hb;
   int[] Ob;
   private int Ib;
   int[] M;
   byte[] zb;
   boolean cb;
   int[] a;
   private int Bb = 155;
   private int C;
   static int wb;
   private int[] Pb;
   private int[] Gb;
   private int sb;
   private int z;
   static int ac;
   private int[] Cb;
   int[] bb;
   static int Rb;
   int[] E;
   static int p;
   private int eb;
   int[] pb;
   private boolean b;
   private int P;
   boolean Kb;
   private int Sb;
   private int[] Lb;
   private int jb;
   static int d;
   private int r;
   private int g;
   int Yb;
   private int Tb;
   static int m;
   private int e;
   static int S;
   int[] qb;
   static int mb;
   static int ab;
   private int x;
   private int hb;
   static int R;
   private int K;
   int t;
   static int ub;
   private int Mb;
   private int[] w;
   private boolean Nb;
   private int T;
   static int h;
   private int[] n;
   private int[] Q;
   private int i;
   private int G;
   static int s;
   private int[] ic;
   static int D;
   int[] gb;
   static int N;
   private int yb;
   static int A;
   static int Xb;
   static int[] B;
   int[] cc;
   static int Ub;
   boolean dc;
   static int Wb;
   int[][] o;
   static int u;
   int hc;
   static int kb;
   private int U;
   int rb;
   int[] bc;
   private int j;
   static int I;
   static int nb;
   private int Vb;
   int Db;
   static int ib;
   static int Qb;
   byte[] Ab;
   private int gc;
   private int Y;
   boolean db;
   private int[] Zb;
   int[] V;
   static int L;
   static int O;
   private int[] fb;
   private int F;
   private int Fb;
   private int[] ob;
   static int y;
   static int Z;
   int Jb;
   private int X;
   int[] lb;
   static int J;
   private int[][] fc;
   static int l;
   int[] k;
   static int vb;
   private static final String[] kc = new String[]{
      z(z("x\u0016$p\r")),
      z(z("`Y$\u0013X")),
      z(z("x\u0016$m\r")),
      z(z("u\u0002fQ")),
      z(z("x\u0016$~d3")),
      z(z("u\u0016")),
      z(z("x\u0016$\u0001Lu\u001e~\u0003\r")),
      z(z("x\u0016$zd3")),
      z(z("x\u0016$ud3")),
      z(z("x\u0016${d3")),
      z(z("x\u0016$td3")),
      z(z("x\u0016$l\r")),
      z(z("x\u0016$h\r")),
      z(z("x\u0016$t\r")),
      z(z("x\u0016$z\r")),
      z(z("x\u0016$yd3")),
      z(z("x\u0016$x\r")),
      z(z("x\u0016$j\r")),
      z(z("x\u0016${\r")),
      z(z("x\u0016$~\r")),
      z(z("x\u0016$wd3")),
      z(z("x\u0016$q\r")),
      z(z("x\u0016$\u007fd3")),
      z(z("x\u0016$w\r")),
      z(z("x\u0016$\u007f\r")),
      z(z("x\u0016$k\r")),
      z(z("x\u0016$|d3")),
      z(z("x\u0016$o\r")),
      z(z("x\u0016$s\r")),
      z(z("x\u0016$i\r")),
      z(z("x\u0016$u\r")),
      z(z("x\u0016$r\r")),
      z(z("x\u0016$n\r")),
      z(z("x\u0016$y\r")),
      z(z("x\u0016$|\r")),
      z(z("x\u0016$vd3")),
      z(z("x\u0016$v\r")),
      z(z("x\u0016$xd3"))
   };

   private final void a(int var1) {
      boolean var13 = client.vh;

      try {
         h++;
         this.x = 999999;
         this.ec = 999999;
         this.gc = -999999;
         this.P = var1;
         this.j = -999999;
         this.e = -999999;
         this.sb = -999999;
         int var3 = 0;

         while (~this.t < ~var3) {
            int[] var2 = this.o[var3];
            int var6 = this.lb[var3];
            int var5 = var2[0];
            int var11;
            int var12 = var11 = this.jc[var5];
            int var9;
            int var10 = var9 = this.ic[var5];
            int var7;
            int var8 = var7 = this.Gb[var5];
            if (var13) {
               break;
            }

            int var4 = 0;

            int var10000;
            int var10001;
            while (true) {
               if (var6 > var4) {
                  var5 = var2[var4];
                  var10000 = ~var11;
                  var10001 = ~this.jc[var5];
                  if (var13) {
                     break;
                  }

                  label124: {
                     if (var10000 >= var10001) {
                        if (~var12 <= ~this.jc[var5]) {
                           break label124;
                        }

                        var12 = this.jc[var5];
                        if (!var13) {
                           break label124;
                        }
                     }

                     var11 = this.jc[var5];
                  }

                  label117: {
                     if (this.ic[var5] < var9) {
                        var9 = this.ic[var5];
                        if (!var13) {
                           break label117;
                        }
                     }

                     if (this.ic[var5] > var10) {
                        var10 = this.ic[var5];
                     }
                  }

                  label112: {
                     if (~this.Gb[var5] <= ~var7) {
                        if (~var8 <= ~this.Gb[var5]) {
                           break label112;
                        }

                        var8 = this.Gb[var5];
                        if (!var13) {
                           break label112;
                        }
                     }

                     var7 = this.Gb[var5];
                  }

                  var4++;
                  if (!var13) {
                     continue;
                  }
               }

               if (!this.c) {
                  this.w[var3] = var7;
                  this.n[var3] = var8;
                  this.Q[var3] = var9;
                  this.Lb[var3] = var10;
                  this.Zb[var3] = var11;
                  this.Eb[var3] = var12;
               }

               var10000 = -var7 + var8;
               var10001 = this.sb;
               break;
            }

            if (var10000 > var10001) {
               this.sb = -var7 + var8;
            }

            if (var10 + -var9 > this.sb) {
               this.sb = -var9 + var10;
            }

            if (this.j < var8) {
               this.j = var8;
            }

            if (~this.gc > ~var12) {
               this.gc = var12;
            }

            if (-var11 + var12 > this.sb) {
               this.sb = -var11 + var12;
            }

            if (~var10 < ~this.e) {
               this.e = var10;
            }

            if (~this.x < ~var7) {
               this.x = var7;
            }

            if (~this.P < ~var9) {
               this.P = var9;
            }

            if (this.ec > var11) {
               this.ec = var11;
            }

            var3++;
            if (var13) {
               break;
            }
         }
      } catch (RuntimeException var14) {
         throw i.a(var14, kc[19] + var1 + ')');
      }
   }

   final void g(int var1, int var2, int var3, int var4) {
      try {
         this.C = var1 & 0xFF;
         m++;
         if (var2 != -999999) {
            this.a(115, -103, 21, -85, -116, -56);
         }

         this.F = var3 & 0xFF;
         this.X = var4 & 0xFF;
         this.b((byte)-117);
         this.Yb = 1;
      } catch (RuntimeException var6) {
         throw i.a(var6, kc[37] + var1 + ',' + var2 + ',' + var3 + ',' + var4 + ')');
      }
   }

   private final void a(int var1, ca[] var2, boolean var3, int var4) {
      boolean var14 = client.vh;

      try {
         kb++;
         int var5 = 0;
         int var6 = 0;
         int var7 = 0;

         while (true) {
            if (~var4 < ~var7) {
               var5 += var2[var7].t;
               var6 += var2[var7].Db;
               var7++;
               if (var14) {
                  break;
               }

               if (!var14) {
                  continue;
               }
            }

            this.a(var5, var6, 88);
            break;
         }

         if (var3) {
            this.fc = new int[var5][];
         }

         var7 = var1;

         label118:
         while (true) {
            int var10000 = var7;

            label115:
            while (var10000 < var4) {
               ca var8 = var2[var7];
               var8.a((byte)-28);
               this.Ib = var8.Ib;
               this.g = var8.g;
               this.Bb = var8.Bb;
               this.Mb = var8.Mb;
               this.Fb = var8.Fb;
               this.Jb = var8.Jb;
               if (var14) {
                  return;
               }

               int var9 = 0;

               while (~var9 > ~var8.t) {
                  int[] var10 = new int[var8.lb[var9]];
                  int[] var11 = var8.o[var9];
                  var10000 = 0;
                  if (var14) {
                     continue label115;
                  }

                  int var12 = 0;

                  while (true) {
                     if (~var12 > ~var8.lb[var9]) {
                        var10[var12] = this.e(var8.a[var11[var12]], var8.bc[var11[var12]], var8.ob[var11[var12]], -122);
                        var12++;
                        if (var14) {
                           break;
                        }

                        if (!var14) {
                           continue;
                        }
                     }

                     var12 = this.a(var8.lb[var9], var10, var8.V[var9], var8.qb[var9], false);
                     this.Hb[var12] = var8.Hb[var9];
                     this.M[var12] = var8.M[var9];
                     this.k[var12] = var8.k[var9];
                     break;
                  }

                  label101: {
                     label100:
                     if (var3) {
                        if (1 >= var4) {
                           this.fc[var12] = new int[var8.fc[var9].length];
                           int var13 = 0;

                           while (~var8.fc[var9].length < ~var13) {
                              this.fc[var12][var13] = var8.fc[var9][var13];
                              var13++;
                              if (var14) {
                                 break label101;
                              }

                              if (var14) {
                                 break;
                              }
                           }

                           if (!var14) {
                              break label100;
                           }
                        }

                        this.fc[var12] = new int[var8.fc[var9].length - -1];
                        this.fc[var12][0] = var7;
                        int var17 = 0;

                        while (var17 < var8.fc[var9].length) {
                           this.fc[var12][1 + var17] = var8.fc[var9][var17];
                           var17++;
                           if (var14) {
                              break label101;
                           }

                           if (var14) {
                              break;
                           }
                        }
                     }

                     var9++;
                  }

                  if (var14) {
                     break;
                  }
               }

               var7++;
               if (!var14) {
                  continue label118;
               }
               break;
            }

            this.Yb = 1;
            return;
         }
      } catch (RuntimeException var15) {
         throw i.a(var15, kc[35] + var1 + ',' + (var2 != null ? kc[1] : kc[3]) + ',' + var3 + ',' + var4 + ')');
      }
   }

   private final void a(int var1, int var2, int var3) {
      try {
         if (!this.db) {
            this.zb = new byte[var1];
            this.E = new int[var1];
         }

         this.k = new int[var1];
         this.ob = new int[var2];
         this.Ab = new byte[var2];
         this.a = new int[var2];
         this.M = new int[var1];
         this.bc = new int[var2];
         this.qb = new int[var1];
         this.gb = new int[var2];
         this.V = new int[var1];
         this.o = new int[var1][];
         this.lb = new int[var1];
         if (!this.b) {
            this.Ob = new int[var2];
            this.H = new int[var2];
            this.cc = new int[var2];
            this.bb = new int[var2];
            this.pb = new int[var2];
         }

         nb++;
         this.Hb = new int[var1];
         this.Y = 256;
         this.X = 0;
         this.U = 256;
         this.Tb = 256;
         this.T = 256;
         if (!this.c) {
            this.n = new int[var1];
            this.Q = new int[var1];
            this.Zb = new int[var1];
            this.Lb = new int[var1];
            this.Eb = new int[var1];
            this.w = new int[var1];
         }

         if (!this.Nb || !this.c) {
            this.Cb = new int[var1];
            this.Pb = new int[var1];
            this.fb = new int[var1];
         }

         label58: {
            this.G = 256;
            this.t = 0;
            this.F = 0;
            this.Sb = 0;
            this.yb = 256;
            if (!this.v) {
               this.Gb = new int[var2];
               this.jc = new int[var2];
               this.ic = new int[var2];
               if (!client.vh) {
                  break label58;
               }
            }

            this.Gb = this.a;
            this.jc = this.bc;
            this.ic = this.ob;
         }

         this.r = 0;
         this.K = var2;
         this.eb = 256;
         this.f = 256;
         this.xb = 0;
         this.jb = 0;
         this.Db = 0;
         this.i = 256;
         if (var3 <= 68) {
            this.lb = (int[])null;
         }

         this.z = var1;
         this.C = 0;
      } catch (RuntimeException var5) {
         throw i.a(var5, kc[29] + var1 + ',' + var2 + ',' + var3 + ')');
      }
   }

   final void a(int var1, int var2, int var3, int var4, int var5, int var6) {
      try {
         Ub++;
         this.Jb = -(4 * var2) + 256;
         this.Mb = (64 - var1) * 16 + 128;
         if (var4 > -110) {
            this.Bb = -67;
         }

         if (!this.Nb) {
            this.g = var5;
            this.Fb = var6;
            this.Bb = var3;
            this.Ib = (int)Math.sqrt(var6 * var6 + var3 * var3 + var5 * var5);
            this.e(-102);
         }
      } catch (RuntimeException var8) {
         throw i.a(var8, kc[14] + var1 + ',' + var2 + ',' + var3 + ',' + var4 + ',' + var5 + ',' + var6 + ')');
      }
   }

   // $VF: Irreducible bytecode was duplicated to produce valid code
   final ca[] a(int var1, int var2, int var3, int var4, int var5, int var6, int var7, boolean var8, int var9) {
      boolean var19 = client.vh;

      try {
         Rb++;
         this.a((byte)-28);
         int[] var10 = new int[var5];
         int[] var11 = new int[var5];
         int var12 = 0;

         while (true) {
            if (~var12 > ~var5) {
               var10[var12] = 0;
               var11[var12] = 0;
               var12++;
               if (var19) {
                  break;
               }

               if (!var19) {
                  continue;
               }
            }

            var12 = 0;
            break;
         }

         int var10000;
         while (true) {
            if (var12 < this.t) {
               int var13 = 0;
               int var14 = 0;
               int var15 = this.lb[var12];
               int[] var16 = this.o[var12];
               var10000 = 0;
               if (var19) {
                  break;
               }

               int var17 = 0;

               label115: {
                  while (~var17 > ~var15) {
                     var13 += this.a[var16[var17]];
                     var14 += this.bc[var16[var17]];
                     var17++;
                     if (var19) {
                        break label115;
                     }

                     if (var19) {
                        break;
                     }
                  }

                  var17 = var13 / (var15 * var7) + var14 / (var3 * var15) * var2;
                  var10[var17] += var15;
                  var11[var17]++;
                  var12++;
               }

               if (!var19) {
                  continue;
               }
            }

            var10000 = var5;
            break;
         }

         ca[] var21 = new ca[var10000];
         int var22 = 0;

         label100: {
            label99:
            while (true) {
               int var10001;
               if (~var22 > ~var5) {
                  var10000 = var6;
                  var10001 = var10[var22];
                  if (!var19) {
                     if (var6 < var10001) {
                        var10[var22] = var6;
                     }

                     var21[var22] = new ca(var10[var22], var11[var22], true, true, true, var8, true);
                     var21[var22].Mb = this.Mb;
                     var21[var22].Jb = this.Jb;
                     var22++;
                     if (!var19) {
                        continue;
                     }

                     var22 = 0;
                     var10000 = this.t;
                     var10001 = var22;
                  }
               } else {
                  var22 = 0;
                  var10000 = this.t;
                  var10001 = var22;
               }

               while (true) {
                  if (var10000 <= var10001) {
                     break label99;
                  }

                  int var23 = 0;
                  int var24 = 0;
                  int var25 = this.lb[var22];
                  int[] var27 = this.o[var22];
                  boolean var30 = false;
                  if (var19) {
                     break label100;
                  }

                  int var18 = 0;

                  label138: {
                     while (~var25 < ~var18) {
                        var23 += this.a[var27[var18]];
                        var24 += this.bc[var27[var18]];
                        var18++;
                        if (var19) {
                           if (var19) {
                              break label99;
                           }
                           break label138;
                        }

                        if (var19) {
                           break;
                        }
                     }

                     var18 = var23 / (var25 * var7) + var2 * (var24 / (var3 * var25));
                     this.a(var27, var21[var18], var25, var22, 5916);
                     var22++;
                     if (var19) {
                        break label99;
                     }
                  }

                  var10000 = this.t;
                  var10001 = var22;
               }
            }

            var22 = 0;
            var10000 = -50 / ((var4 - -33) / 60);
         }

         while (true) {
            if (~var22 > ~var5) {
               var32 = var21;
               if (var19) {
                  break;
               }

               var21[var22].c((byte)71);
               var22++;
               if (!var19) {
                  continue;
               }
            }

            var32 = var21;
            break;
         }

         return var32;
      } catch (RuntimeException var20) {
         throw i.a(var20, kc[23] + var1 + ',' + var2 + ',' + var3 + ',' + var4 + ',' + var5 + ',' + var6 + ',' + var7 + ',' + var8 + ',' + var9 + ')');
      }
   }

   final int a(int var1, int[] var2, int var3, int var4, boolean var5) {
      try {
         u++;
         if (var5) {
            this.f(30, -84, 10, 23);
         }

         if (~this.t > ~this.z) {
            this.lb[this.t] = var1;
            this.o[this.t] = var2;
            this.V[this.t] = var3;
            this.qb[this.t] = var4;
            this.Yb = 1;
            return this.t++;
         } else {
            return -1;
         }
      } catch (RuntimeException var7) {
         throw i.a(var7, kc[2] + var1 + 44 + (var2 != null ? kc[1] : kc[3]) + 44 + var3 + 44 + var4 + 44 + var5 + 41);
      }
   }

   final void a(int var1, int var2, int var3, boolean var4) {
      try {
         this.xb += var3;
         this.Sb += var2;
         if (!var4) {
            this.jc = (int[])null;
         }

         this.r += var1;
         L++;
         this.b((byte)-127);
         this.Yb = 1;
      } catch (RuntimeException var6) {
         throw i.a(var6, kc[31] + var1 + ',' + var2 + ',' + var3 + ',' + var4 + ')');
      }
   }

   private final void a(int var1, int var2, int var3, int var4, int var5, int var6, byte var7) {
      boolean var10 = client.vh;

      try {
         S++;
         int var8 = 118 / ((-80 - var7) / 46);
         int var9 = 0;

         while (var9 < this.Db && !var10) {
            if (-1 != ~var3) {
               this.Gb[var9] = this.Gb[var9] + (this.ic[var9] * var3 >> 8);
            }

            if (0 != var4) {
               this.jc[var9] = this.jc[var9] + (var4 * this.ic[var9] >> 8);
            }

            if (-1 != ~var5) {
               this.Gb[var9] = this.Gb[var9] + (var5 * this.jc[var9] >> 8);
            }

            if (-1 != ~var2) {
               this.ic[var9] = this.ic[var9] + (var2 * this.jc[var9] >> 8);
            }

            if (-1 != ~var1) {
               this.jc[var9] = this.jc[var9] + (var1 * this.Gb[var9] >> 8);
            }

            if (-1 != ~var6) {
               this.ic[var9] = this.ic[var9] + (this.Gb[var9] * var6 >> 8);
            }

            var9++;
            if (var10) {
               break;
            }
         }
      } catch (RuntimeException var11) {
         throw i.a(var11, kc[16] + var1 + ',' + var2 + ',' + var3 + ',' + var4 + ',' + var5 + ',' + var6 + ',' + var7 + ')');
      }
   }

   final void b(int var1, int var2, int var3) {
      try {
         ub++;
         this.t -= var3;
         if (var2 > -110) {
            a((byte)-3, (String)null);
         }

         if (~this.t > -1) {
            this.t = 0;
         }

         this.Db -= var1;
         if (this.Db < 0) {
            this.Db = 0;
         }
      } catch (RuntimeException var5) {
         throw i.a(var5, kc[18] + var1 + ',' + var2 + ',' + var3 + ')');
      }
   }

   final void c(int var1, int var2, int var3, int var4) {
      try {
         this.xb = var1;
         this.r = var4;
         this.Sb = var3;
         if (var2 > -112) {
            this.a(-96, (int[])null, -8, 42, true);
         }

         q++;
         this.b((byte)-114);
         this.Yb = 1;
      } catch (RuntimeException var6) {
         throw i.a(var6, kc[21] + var1 + ',' + var2 + ',' + var3 + ',' + var4 + ')');
      }
   }

   final void a(int var1, int var2, int var3, byte var4, int var5, int var6, int var7, int var8, int var9) {
      boolean var21 = client.vh;

      try {
         ac++;
         this.d(7972);
         if (~da.K <= ~this.ec && m.j <= this.gc && ~this.x >= ~oa.b && aa.f <= this.j && this.P <= nb.y && this.e >= aa.b) {
            this.dc = true;
            if (!var21) {
               int var11 = 0;
               if (var4 > -105) {
                  this.c((byte)-103);
               }

               int var12 = 0;
               int var13 = 0;
               int var14 = 0;
               int var15 = 0;
               if (var7 != 0) {
                  var12 = pa.j[1024 + var7];
                  var11 = pa.j[var7];
               }

               int var16 = 0;
               if (0 != var8) {
                  var13 = pa.j[var8];
                  var14 = pa.j[var8 - -1024];
               }

               if (var6 != 0) {
                  var15 = pa.j[var6];
                  var16 = pa.j[var6 - -1024];
               }

               int var17 = 0;

               while (this.Db > var17) {
                  int var18 = this.Gb[var17] + -var3;
                  int var19 = -var1 + this.ic[var17];
                  if (var21) {
                     break;
                  }

                  if (var7 != 0) {
                     int var10 = var19 * var11 - -(var12 * var18) >> 15;
                     var19 = -(var18 * var11) + var19 * var12 >> 15;
                     var18 = var10;
                  }

                  int var20 = this.jc[var17] - var5;
                  if (~var6 != -1) {
                     int var23 = var16 * var18 + var20 * var15 >> 15;
                     var20 = -(var18 * var15) + var16 * var20 >> 15;
                     var18 = var23;
                  }

                  if (~var8 != -1) {
                     int var24 = var19 * var14 + -(var13 * var20) >> 15;
                     var20 = var13 * var19 - -(var14 * var20) >> 15;
                     var19 = var24;
                  }

                  label95: {
                     if (var20 < var9) {
                        this.pb[var17] = var18 << var2;
                        if (!var21) {
                           break label95;
                        }
                     }

                     this.pb[var17] = (var18 << var2) / var20;
                  }

                  label90: {
                     if (var20 < var9) {
                        this.Ob[var17] = var19 << var2;
                        if (!var21) {
                           break label90;
                        }
                     }

                     this.Ob[var17] = (var19 << var2) / var20;
                  }

                  this.cc[var17] = var18;
                  this.H[var17] = var19;
                  this.bb[var17] = var20;
                  var17++;
                  if (var21) {
                     break;
                  }
               }

               return;
            }
         }

         this.dc = false;
      } catch (RuntimeException var22) {
         throw i.a(var22, kc[12] + var1 + ',' + var2 + ',' + var3 + ',' + var4 + ',' + var5 + ',' + var6 + ',' + var7 + ',' + var8 + ',' + var9 + ')');
      }
   }

   private final void b(byte var1) {
      try {
         vb++;
         if (var1 >= -103) {
            this.a(true, -115, true, true, false);
         }

         if (256 != this.Tb || this.G != 256 || ~this.Y != -257 || this.f != 256 || -257 != ~this.U || -257 != ~this.eb) {
            this.jb = 4;
         } else if (-257 != ~this.yb || this.i != 256 || -257 != ~this.T) {
            this.jb = 3;
         } else if (0 != this.F || 0 != this.X || 0 != this.C) {
            this.jb = 2;
         } else if (this.r == 0 && this.xb == 0 && -1 == ~this.Sb) {
            this.jb = 0;
         } else {
            this.jb = 1;
         }
      } catch (RuntimeException var3) {
         throw i.a(var3, kc[34] + var1 + ')');
      }
   }

   private final void a(byte var1) {
      boolean var3 = client.vh;

      try {
         this.d(7972);
         D++;
         int var2 = 0;

         while (true) {
            if (~this.Db < ~var2) {
               this.a[var2] = this.Gb[var2];
               this.ob[var2] = this.ic[var2];
               this.bc[var2] = this.jc[var2];
               var2++;
               if (var3) {
                  break;
               }

               if (!var3) {
                  continue;
               }
            }

            if (var1 != -28) {
               this.d(4);
            }

            this.C = 0;
            this.F = 0;
            this.U = 256;
            this.f = 256;
            this.xb = 0;
            this.X = 0;
            this.eb = 256;
            this.Y = 256;
            this.yb = 256;
            this.jb = 0;
            this.r = 0;
            this.i = 256;
            this.Tb = 256;
            this.T = 256;
            this.Sb = 0;
            this.G = 256;
            break;
         }
      } catch (RuntimeException var4) {
         throw i.a(var4, kc[27] + var1 + ')');
      }
   }

   final void a(int var1, int var2, int var3, int var4, boolean var5, int var6, int var7) {
      boolean var10 = client.vh;

      try {
         this.Mb = (-var6 + 64) * 16 + 128;
         int var8 = 76 / ((-8 - var7) / 49);
         this.Jb = 256 - var2 * 4;
         R++;
         if (!this.Nb) {
            int var9 = 0;

            while (true) {
               if (~this.t < ~var9) {
                  if (var10) {
                     break;
                  }

                  label34: {
                     if (var5) {
                        this.Hb[var9] = this.Vb;
                        if (!var10) {
                           break label34;
                        }
                     }

                     this.Hb[var9] = 0;
                  }

                  var9++;
                  if (!var10) {
                     continue;
                  }
               }

               this.g = var4;
               this.Fb = var1;
               this.Bb = var3;
               this.Ib = (int)Math.sqrt(var3 * var3 + (var4 * var4 - -(var1 * var1)));
               this.e(-121);
               break;
            }
         }
      } catch (RuntimeException var11) {
         throw i.a(var11, kc[28] + var1 + ',' + var2 + ',' + var3 + ',' + var4 + ',' + var5 + ',' + var6 + ',' + var7 + ')');
      }
   }

   private final void b(int var1, int var2, int var3, int var4) {
      boolean var6 = client.vh;

      try {
         O++;
         if (var2 != -27483) {
            this.e(-7, -82, -31, -24);
         }

         int var5 = 0;

         while (~var5 > ~this.Db) {
            this.Gb[var5] = this.Gb[var5] * var1 >> 8;
            this.ic[var5] = this.ic[var5] * var4 >> 8;
            this.jc[var5] = var3 * this.jc[var5] >> 8;
            var5++;
            if (var6 || var6) {
               break;
            }
         }
      } catch (RuntimeException var7) {
         throw i.a(var7, kc[22] + var1 + ',' + var2 + ',' + var3 + ',' + var4 + ')');
      }
   }

   final int e(int var1, int var2, int var3, int var4) {
      boolean var7 = client.vh;

      try {
         ib++;
         int var5 = 0;

         int var10000;
         int var10001;
         while (true) {
            if (~var5 > ~this.Db) {
               var10000 = ~this.a[var5];
               var10001 = ~var1;
               if (var7) {
                  break;
               }

               if (var10000 == var10001 && var3 == this.ob[var5] && var2 == this.bc[var5]) {
                  return var5;
               }

               var5++;
               if (!var7) {
                  continue;
               }
            }

            var10000 = 100;
            var10001 = (-46 - var4) / 58;
            break;
         }

         int var6 = var10000 / var10001;
         if (this.Db < this.K) {
            this.a[this.Db] = var1;
            this.ob[this.Db] = var3;
            this.bc[this.Db] = var2;
            return this.Db++;
         } else {
            return -1;
         }
      } catch (RuntimeException var8) {
         throw i.a(var8, kc[13] + var1 + 44 + var2 + 44 + var3 + 44 + var4 + 41);
      }
   }

   private final void a(int[] var1, ca var2, int var3, int var4, int var5) {
      boolean var9 = client.vh;

      try {
         mb++;
         int[] var6 = new int[var3];
         int var7 = 0;

         while (true) {
            if (~var3 < ~var7) {
               int var8 = var6[var7] = var2.e(this.a[var1[var7]], this.bc[var1[var7]], this.ob[var1[var7]], 107);
               var2.gb[var8] = this.gb[var1[var7]];
               var2.Ab[var8] = this.Ab[var1[var7]];
               var7++;
               if (var9) {
                  break;
               }

               if (!var9) {
                  continue;
               }
            }

            if (var5 != 5916) {
               this.yb = 77;
            }

            var7 = var2.a(var3, var6, this.V[var4], this.qb[var4], false);
            break;
         }

         if (!var2.db && !this.db) {
            var2.E[var7] = this.E[var4];
         }

         var2.Hb[var7] = this.Hb[var4];
         var2.M[var7] = this.M[var4];
         var2.k[var7] = this.k[var4];
      } catch (RuntimeException var10) {
         throw i.a(var10, kc[33] + (var1 != null ? kc[1] : kc[3]) + ',' + (var2 != null ? kc[1] : kc[3]) + ',' + var3 + ',' + var4 + ',' + var5 + ')');
      }
   }

   final void c(int var1) {
      try {
         this.t = 0;
         if (var1 != 1) {
            this.Kb = true;
         }

         this.Db = 0;
         W++;
      } catch (RuntimeException var3) {
         throw i.a(var3, kc[0] + var1 + ')');
      }
   }

   private final void a(int var1, int var2, int var3, int var4) {
      boolean var9 = client.vh;

      try {
         ab++;
         if (var1 >= -14) {
            this.Cb = (int[])null;
         }

         int var8 = 0;

         while (~this.Db < ~var8 && !var9) {
            if (var3 != 0) {
               int var6 = pa.a[var3 + 256];
               int var5 = pa.a[var3];
               int var7 = this.Gb[var8] * var6 + this.ic[var8] * var5 >> 15;
               this.ic[var8] = -(var5 * this.Gb[var8]) + this.ic[var8] * var6 >> 15;
               this.Gb[var8] = var7;
            }

            if (-1 != ~var2) {
               int var11 = pa.a[var2];
               int var13 = pa.a[256 + var2];
               int var15 = -(var11 * this.jc[var8]) + var13 * this.ic[var8] >> 15;
               this.jc[var8] = var11 * this.ic[var8] - -(var13 * this.jc[var8]) >> 15;
               this.ic[var8] = var15;
            }

            if (-1 != ~var4) {
               int var12 = pa.a[var4];
               int var14 = pa.a[256 + var4];
               int var16 = var12 * this.jc[var8] + this.Gb[var8] * var14 >> 15;
               this.jc[var8] = -(this.Gb[var8] * var12) + this.jc[var8] * var14 >> 15;
               this.Gb[var8] = var16;
            }

            var8++;
            if (var9) {
               break;
            }
         }
      } catch (RuntimeException var10) {
         throw i.a(var10, kc[17] + var1 + ',' + var2 + ',' + var3 + ',' + var4 + ')');
      }
   }

   // $VF: Irreducible bytecode was duplicated to produce valid code
   private final void e(int var1) {
      boolean var11 = client.vh;

      try {
         l++;
         if (!this.Nb) {
            int var2 = this.Mb * this.Ib >> 8;
            int var3 = 0;

            int var10000;
            while (true) {
               if (this.t > var3) {
                  var10000 = this.Hb[var3];
                  if (var11) {
                     break;
                  }

                  if (var10000 != this.Vb) {
                     this.Hb[var3] = (this.Cb[var3] * this.Bb + (this.fb[var3] * this.g - -(this.Fb * this.Pb[var3]))) / var2;
                  }

                  var3++;
                  if (!var11) {
                     continue;
                  }
               }

               var10000 = this.Db;
               break;
            }

            int[] var13 = new int[var10000];
            int[] var4 = new int[this.Db];
            int[] var5 = new int[this.Db];
            int[] var6 = new int[this.Db];
            int var7 = 0;

            while (true) {
               if (this.Db > var7) {
                  var13[var7] = 0;
                  var4[var7] = 0;
                  var5[var7] = 0;
                  var6[var7] = 0;
                  var7++;
                  if (var11) {
                     break;
                  }

                  if (!var11) {
                     continue;
                  }
               }

               var7 = -16 / ((var1 - -55) / 32);
               break;
            }

            int var8 = 0;

            int var10001;
            while (true) {
               if (var8 < this.t) {
                  var10000 = ~this.Hb[var8];
                  var10001 = ~this.Vb;
                  if (var11) {
                     break;
                  }

                  label81: {
                     if (var10000 == var10001) {
                        int var9 = 0;

                        while (this.lb[var8] > var9) {
                           int var10 = this.o[var8][var9];
                           var13[var10] += this.fb[var8];
                           var4[var10] += this.Cb[var8];
                           var5[var10] += this.Pb[var8];
                           var6[var10]++;
                           var9++;
                           if (var11) {
                              break label81;
                           }

                           if (var11) {
                              break;
                           }
                        }
                     }

                     var8++;
                  }

                  if (!var11) {
                     continue;
                  }

                  var8 = 0;
               } else {
                  var8 = 0;
               }

               var10000 = this.Db;
               var10001 = var8;
               break;
            }

            while (var10000 > var10001 && !var11) {
               if (0 < var6[var8]) {
                  this.gb[var8] = (var5[var8] * this.Fb + var13[var8] * this.g + var4[var8] * this.Bb) / (var2 * var6[var8]);
               }

               var8++;
               if (var11) {
                  break;
               }

               var10000 = this.Db;
               var10001 = var8;
            }
         }
      } catch (RuntimeException var12) {
         throw i.a(var12, kc[9] + var1 + ')');
      }
   }

   final ca a(boolean var1, int var2, boolean var3, boolean var4, boolean var5) {
      try {
         I++;
         ca[] var6 = new ca[]{this};
         int var7 = 122 / ((-56 - var2) / 59);
         ca var8 = new ca(var6, 1, var4, var5, var3, var1);
         var8.hc = this.hc;
         return var8;
      } catch (RuntimeException var9) {
         throw i.a(var9, kc[11] + var1 + ',' + var2 + ',' + var3 + ',' + var4 + ',' + var5 + ')');
      }
   }

   static final int a(byte var0, String var1) {
      boolean var3 = client.vh;

      try {
         Xb++;
         if (var1.equalsIgnoreCase(kc[5])) {
            return 0;
         }

         if (var0 != 91) {
            B = (int[])null;
         }

         int var2 = 0;

         int var10000;
         while (true) {
            if (var2 < ia.b) {
               var10000 = ub.c[var2].equalsIgnoreCase(var1);
               if (var3) {
                  break;
               }

               if (var10000 != 0) {
                  return var2;
               }

               var2++;
               if (!var3) {
                  continue;
               }
            }

            ub.c[ia.b++] = var1;
            var10000 = ia.b + -1;
            break;
         }

         return var10000;
      } catch (RuntimeException var4) {
         throw i.a(var4, kc[4] + var0 + 44 + (var1 != null ? kc[1] : kc[3]) + 41);
      }
   }

   final void a(ca var1, int var2) {
      try {
         if (var2 != 6029) {
            this.jb = -128;
         }

         this.xb = var1.xb;
         y++;
         this.C = var1.C;
         this.Sb = var1.Sb;
         this.r = var1.r;
         this.X = var1.X;
         this.F = var1.F;
         this.b((byte)-113);
         this.Yb = 1;
      } catch (RuntimeException var4) {
         throw i.a(var4, kc[26] + (var1 != null ? kc[1] : kc[3]) + ',' + var2 + ')');
      }
   }

   final ca b(int var1) {
      try {
         wb++;
         ca[] var2 = new ca[1];
         if (var1 != -2) {
            this.b(117, -93, -34, -34);
         }

         var2[0] = this;
         ca var3 = new ca(var2, 1);
         var3.cb = this.cb;
         var3.hc = this.hc;
         return var3;
      } catch (RuntimeException var4) {
         throw i.a(var4, kc[10] + var1 + ')');
      }
   }

   private final void c(byte var1) {
      try {
         if (var1 < 49) {
            this.a(40, 102, 104, 108, -20, -89);
         }

         this.bb = new int[this.Db];
         this.Ob = new int[this.Db];
         this.cc = new int[this.Db];
         this.H = new int[this.Db];
         this.pb = new int[this.Db];
         N++;
      } catch (RuntimeException var3) {
         throw i.a(var3, kc[30] + var1 + ')');
      }
   }

   final void a(int var1, int var2, byte var3) {
      try {
         if (var3 == -61) {
            Wb++;
            this.Ab[var1] = (byte)var2;
         }
      } catch (RuntimeException var5) {
         throw i.a(var5, kc[8] + var1 + ',' + var2 + ',' + var3 + ')');
      }
   }

   private final void d(int var1) {
      boolean var3 = client.vh;

      try {
         if (var1 != 7972) {
            this.a(120, 20, 57, true);
         }

         Qb++;
         if (2 == this.Yb) {
            this.Yb = 0;
            int var2 = 0;

            label79: {
               while (~this.Db < ~var2) {
                  this.Gb[var2] = this.a[var2];
                  this.ic[var2] = this.ob[var2];
                  this.jc[var2] = this.bc[var2];
                  var2++;
                  if (var3) {
                     break label79;
                  }

                  if (var3) {
                     break;
                  }
               }

               this.j = 9999999;
               this.e = 9999999;
               this.P = -9999999;
               this.gc = 9999999;
               this.ec = -9999999;
               this.sb = 9999999;
               this.x = -9999999;
            }

            if (!var3) {
               return;
            }
         }

         if (~this.Yb == -2) {
            this.Yb = 0;
            int var5 = 0;

            label65: {
               while (true) {
                  if (~var5 > ~this.Db) {
                     this.Gb[var5] = this.a[var5];
                     this.ic[var5] = this.ob[var5];
                     this.jc[var5] = this.bc[var5];
                     var5++;
                     if (var3) {
                        break;
                     }

                     if (!var3) {
                        continue;
                     }
                  }

                  if (2 > this.jb) {
                     break label65;
                  }
                  break;
               }

               this.a(-53, this.F, this.C, this.X);
            }

            if (this.jb >= 3) {
               this.b(this.yb, -27483, this.T, this.i);
            }

            if (4 <= this.jb) {
               this.a(this.U, this.f, this.Tb, this.G, this.Y, this.eb, (byte)-127);
            }

            if (1 <= this.jb) {
               this.d(var1 + -7972, this.xb, this.Sb, this.r);
            }

            this.a(999999);
            this.d((byte)14);
         }
      } catch (RuntimeException var4) {
         throw i.a(var4, kc[25] + var1 + ')');
      }
   }

   final int b(boolean var1, int var2, int var3, int var4) {
      try {
         J++;
         if (~this.K >= ~this.Db) {
            return -1;
         }

         this.a[this.Db] = var3;
         this.ob[this.Db] = var4;
         this.bc[this.Db] = var2;
         if (var1) {
            a((byte)52, (String)null);
         }

         return this.Db++;
      } catch (RuntimeException var6) {
         throw i.a(var6, kc[32] + var1 + 44 + var2 + 44 + var3 + 44 + var4 + 41);
      }
   }

   private final void d(byte var1) {
      boolean var17 = client.vh;

      try {
         Z++;
         if (!this.Nb || !this.c) {
            if (var1 == 14) {
               int var2 = 0;

               while (true) {
                  if (this.t > var2) {
                     int[] var3 = this.o[var2];
                     int var4 = this.Gb[var3[0]];
                     int var5 = this.ic[var3[0]];
                     int var6 = this.jc[var3[0]];
                     int var7 = -var4 + this.Gb[var3[1]];
                     int var8 = this.ic[var3[1]] + -var5;
                     int var9 = -var6 + this.jc[var3[1]];
                     int var10 = -var4 + this.Gb[var3[2]];
                     int var11 = this.ic[var3[2]] - var5;
                     int var12 = -var6 + this.jc[var3[2]];
                     int var13 = var12 * var8 - var9 * var11;
                     int var14 = -(var7 * var12) + var9 * var10;
                     if (var17) {
                        break;
                     }

                     int var15 = -(var10 * var8) + var7 * var11;

                     do {
                        int var10000;
                        short var10001;
                        label72: {
                           if (8192 >= var13) {
                              var10000 = ~var14;
                              var10001 = -8193;
                              if (var17) {
                                 break label72;
                              }

                              if (var10000 >= -8193 && var15 <= 8192 && var13 >= -8192 && var14 >= -8192 && -8192 <= var15) {
                                 break;
                              }
                           }

                           var14 >>= 1;
                           var15 >>= 1;
                           var10000 = var13;
                           var10001 = 1;
                        }

                        var13 = var10000 >> var10001;
                     } while (!var17);

                     int var16 = (int)(Math.sqrt(var14 * var14 + var13 * var13 + var15 * var15) * 256.0);
                     if (var16 <= 0) {
                        var16 = 1;
                     }

                     this.fb[var2] = var13 * 65536 / var16;
                     this.Cb[var2] = 65536 * var14 / var16;
                     this.Pb[var2] = var15 * 65535 / var16;
                     this.M[var2] = -1;
                     var2++;
                     if (!var17) {
                        continue;
                     }
                  }

                  this.e(var1 ^ -85);
                  break;
               }
            }
         }
      } catch (RuntimeException var18) {
         throw i.a(var18, kc[36] + var1 + ')');
      }
   }

   final void f(int var1, int var2, int var3, int var4) {
      try {
         this.F = var4 + this.F & 0xFF;
         A++;
         this.X = var3 + this.X & 0xFF;
         if (var2 == -31616) {
            this.C = 0xFF & this.C - -var1;
            this.b((byte)-105);
            this.Yb = 1;
         }
      } catch (RuntimeException var6) {
         throw i.a(var6, kc[7] + var1 + ',' + var2 + ',' + var3 + ',' + var4 + ')');
      }
   }

   private final void d(int var1, int var2, int var3, int var4) {
      boolean var6 = client.vh;

      try {
         d++;
         int var5 = var1;

         while (this.Db > var5) {
            this.Gb[var5] = this.Gb[var5] + var4;
            this.ic[var5] = this.ic[var5] + var2;
            this.jc[var5] = this.jc[var5] + var3;
            var5++;
            if (var6 || var6) {
               break;
            }
         }
      } catch (RuntimeException var7) {
         throw i.a(var7, kc[15] + var1 + ',' + var2 + ',' + var3 + ',' + var4 + ')');
      }
   }

   private final int a(byte var1, byte[] var2) {
      boolean var7 = client.vh;

      try {
         if (var1 != 76) {
            return 108;
         }

         s++;

         while (~var2[this.hb] == -11 || -14 == ~var2[this.hb]) {
            this.hb++;
            if (var7) {
               break;
            }
         }

         int var3 = pa.d[0xFF & var2[this.hb++]];
         int var4 = pa.d[0xFF & var2[this.hb++]];
         int var5 = pa.d[var2[this.hb++] & 0xFF];
         int var6 = -131072 + var4 * 64 + 4096 * var3 + var5;
         if (-123457 == ~var6) {
            var6 = this.Vb;
         }

         return var6;
      } catch (RuntimeException var8) {
         throw i.a(var8, kc[20] + var1 + 44 + (var2 != null ? kc[1] : kc[3]) + 41);
      }
   }

   final void a(boolean var1, int var2, int var3, int var4) {
      try {
         if (var1) {
            this.Yb = 71;
         }

         p++;
         if (!this.Nb) {
            this.Fb = var4;
            this.Bb = var3;
            this.g = var2;
            this.Ib = (int)Math.sqrt(var4 * var4 + var3 * var3 + var2 * var2);
            this.e(52);
         }
      } catch (RuntimeException var6) {
         throw i.a(var6, kc[24] + var1 + ',' + var2 + ',' + var3 + ',' + var4 + ')');
      }
   }

   ca(int var1, int var2) {
      boolean var4 = client.vh;
      super();
      this.Bb = 155;
      this.c = false;
      this.Yb = 1;
      this.b = false;
      this.Ib = 256;
      this.v = false;
      this.cb = false;
      this.Nb = false;
      this.sb = 12345678;
      this.g = 180;
      this.dc = true;
      this.Vb = 12345678;
      this.Kb = false;
      this.rb = -1;
      this.Mb = 512;
      this.Fb = 95;
      this.db = false;
      this.Jb = 32;
      this.hc = 0;

      try {
         this.a(var2, var1, 69);
         this.fc = new int[var2][1];
         int var3 = 0;

         while (var3 < var2) {
            this.fc[var3][0] = var3++;
            if (var4 || var4) {
               break;
            }
         }
      } catch (RuntimeException var5) {
         throw i.a(var5, kc[6] + var1 + ',' + var2 + ')');
      }
   }

   ca(int var1, int var2, boolean var3, boolean var4, boolean var5, boolean var6, boolean var7) {
      this.c = false;
      this.Yb = 1;
      this.b = false;
      this.Ib = 256;
      this.v = false;
      this.cb = false;
      this.Nb = false;
      this.sb = 12345678;
      this.g = 180;
      this.dc = true;
      this.Vb = 12345678;
      this.Kb = false;
      this.rb = -1;
      this.Mb = 512;
      this.Fb = 95;
      this.db = false;
      this.Jb = 32;
      this.hc = 0;

      try {
         this.c = var4;
         this.b = var7;
         this.db = var6;
         this.v = var3;
         this.Nb = var5;
         this.a(var2, var1, 69);
      } catch (RuntimeException var9) {
         throw i.a(var9, kc[6] + var1 + ',' + var2 + ',' + var3 + ',' + var4 + ',' + var5 + ',' + var6 + ',' + var7 + ')');
      }
   }

   // $VF: Irreducible bytecode was duplicated to produce valid code
   ca(byte[] var1, int var2, boolean var3) {
      boolean var8 = client.vh;
      super();
      this.Bb = 155;
      this.c = false;
      this.Yb = 1;
      this.b = false;
      this.Ib = 256;
      this.v = false;
      this.cb = false;
      this.Nb = false;
      this.sb = 12345678;
      this.g = 180;
      this.dc = true;
      this.Vb = 12345678;
      this.Kb = false;
      this.rb = -1;
      this.Mb = 512;
      this.Fb = 95;
      this.db = false;
      this.Jb = 32;
      this.hc = 0;

      try {
         int var4 = d.a(var2, (byte)7, var1);
         var2 += 2;
         int var5 = d.a(var2, (byte)8, var1);
         var2 += 2;
         this.a(var5, var4, 115);
         this.fc = new int[var5][1];
         int var6 = 0;

         while (true) {
            if (var4 > var6) {
               this.a[var6] = w.a(var1, -1, var2);
               var2 += 2;
               var6++;
               if (var8) {
                  break;
               }

               if (!var8) {
                  continue;
               }
            }

            var6 = 0;
            break;
         }

         while (true) {
            if (~var4 < ~var6) {
               this.ob[var6] = w.a(var1, -1, var2);
               var2 += 2;
               var6++;
               if (var8) {
                  break;
               }

               if (!var8) {
                  continue;
               }
            }

            var6 = 0;
            break;
         }

         while (true) {
            if (var4 > var6) {
               this.bc[var6] = w.a(var1, -1, var2);
               var2 += 2;
               var6++;
               if (var8) {
                  break;
               }

               if (!var8) {
                  continue;
               }
            }

            this.Db = var4;
            break;
         }

         var6 = 0;

         while (true) {
            if (~var5 < ~var6) {
               this.lb[var6] = ib.a(255, var1[var2++]);
               var6++;
               if (var8) {
                  break;
               }

               if (!var8) {
                  continue;
               }
            }

            var6 = 0;
            break;
         }

         int var14;
         label173: {
            label172:
            while (true) {
               int var10001;
               if (~var5 < ~var6) {
                  this.V[var6] = w.a(var1, -1, var2);
                  var14 = this.V[var6];
                  var10001 = 32767;
                  if (!var8) {
                     if (var14 == 32767) {
                        this.V[var6] = this.Vb;
                     }

                     var2 += 2;
                     var6++;
                     if (!var8) {
                        continue;
                     }

                     var6 = 0;
                     var14 = ~var6;
                     var10001 = ~var5;
                  }
               } else {
                  var6 = 0;
                  var14 = ~var6;
                  var10001 = ~var5;
               }

               while (true) {
                  if (var14 > var10001) {
                     this.qb[var6] = w.a(var1, -1, var2);
                     var2 += 2;
                     var14 = 32767;
                     var10001 = this.qb[var6];
                     if (!var8) {
                        if (32767 == var10001) {
                           this.qb[var6] = this.Vb;
                        }

                        var6++;
                        if (!var8) {
                           var14 = ~var6;
                           var10001 = ~var5;
                           continue;
                        }

                        var6 = 0;
                        var14 = ~var5;
                        var10001 = ~var6;
                     }
                  } else {
                     var6 = 0;
                     var14 = ~var5;
                     var10001 = ~var6;
                  }

                  while (true) {
                     if (var14 >= var10001) {
                        break label172;
                     }

                     int var7 = 255 & var1[var2++];
                     var14 = var7;
                     if (var8) {
                        break label173;
                     }

                     if (var7 != 0) {
                        this.Hb[var6] = this.Vb;
                        if (var8) {
                           this.Hb[var6] = 0;
                           var6++;
                           if (var8) {
                              break label172;
                           }
                        } else {
                           var6++;
                           if (var8) {
                              break label172;
                           }
                        }
                     } else {
                        this.Hb[var6] = 0;
                        var6++;
                        if (var8) {
                           break label172;
                        }
                     }

                     var14 = ~var5;
                     var10001 = ~var6;
                  }
               }
            }

            var14 = 0;
         }

         var6 = var14;

         label126:
         while (true) {
            var14 = ~var6;
            int var18 = ~var5;

            label123:
            while (var14 > var18) {
               this.o[var6] = new int[this.lb[var6]];
               if (var8) {
                  return;
               }

               int var12 = 0;

               while (this.lb[var6] > var12) {
                  var14 = var4;
                  var18 = 256;
                  if (var8) {
                     continue label123;
                  }

                  label117: {
                     if (var4 < 256) {
                        this.o[var6][var12] = ib.a(255, var1[var2++]);
                        if (!var8) {
                           break label117;
                        }
                     }

                     this.o[var6][var12] = d.a(var2, (byte)102, var1);
                     var2 += 2;
                  }

                  var12++;
                  if (var8) {
                     break;
                  }
               }

               var6++;
               if (!var8) {
                  continue label126;
               }
               break;
            }

            this.t = var5;
            this.Yb = 1;
            return;
         }
      } catch (RuntimeException var9) {
         throw i.a(var9, kc[6] + (var1 != null ? kc[1] : kc[3]) + ',' + var2 + ',' + var3 + ')');
      }
   }

   ca(String var1) {
      boolean var19 = client.vh;
      super();
      this.Bb = 155;
      this.c = false;
      this.Yb = 1;
      this.b = false;
      this.Ib = 256;
      this.v = false;
      this.cb = false;
      this.Nb = false;
      this.sb = 12345678;
      this.g = 180;
      this.dc = true;
      this.Vb = 12345678;
      this.Kb = false;
      this.rb = -1;
      this.Mb = 512;
      this.Fb = 95;
      this.db = false;
      this.Jb = 32;
      this.hc = 0;

      try {
         int var2 = 0;
         int var3 = 0;
         byte[] var4 = null;

         try {
            label135: {
               InputStream var5 = nb.a(true, var1);
               DataInputStream var6 = new DataInputStream(var5);
               var4 = new byte[3];
               var2 = 0;
               this.hb = 0;

               label116: {
                  while (~var2 > -4) {
                     var2 += var6.read(var4, var2, -var2 + 3);
                     if (var19) {
                        break label116;
                     }

                     if (var19) {
                        break;
                     }
                  }

                  var3 = this.a((byte)76, var4);
                  var2 = 0;
                  this.hb = 0;
               }

               var4 = new byte[var3];

               while (~var2 > ~var3) {
                  var2 += var6.read(var4, var2, var3 - var2);
                  if (var19) {
                     break label135;
                  }

                  if (var19) {
                     break;
                  }
               }

               var6.close();
            }
         } catch (IOException var20) {
            this.t = 0;
            this.Db = 0;
            return;
         }

         int var25 = this.a((byte)76, var4);
         int var26 = this.a((byte)76, var4);
         int var14 = 0;
         this.a(var26, var25, 97);
         this.fc = new int[var26][];
         int var15 = 0;

         while (true) {
            if (var25 > var15) {
               int var7 = this.a((byte)76, var4);
               int var8 = this.a((byte)76, var4);
               int var9 = this.a((byte)76, var4);
               this.e(var7, var9, var8, 52);
               var15++;
               if (var19) {
                  break;
               }

               if (!var19) {
                  continue;
               }
            }

            var15 = 0;
            break;
         }

         label97:
         while (true) {
            if (~var26 < ~var15) {
               int var10 = this.a((byte)76, var4);
               int var11 = this.a((byte)76, var4);
               int var12 = this.a((byte)76, var4);
               int var13 = this.a((byte)76, var4);
               this.Mb = this.a((byte)76, var4);
               this.Jb = this.a((byte)76, var4);
               var14 = this.a((byte)76, var4);
               int[] var16 = new int[var10];
               if (var19) {
                  break;
               }

               int var17 = 0;

               while (var17 < var10) {
                  var16[var17] = this.a((byte)76, var4);
                  var17++;
                  if (var19) {
                     continue label97;
                  }

                  if (var19) {
                     break;
                  }
               }

               int[] var28 = new int[var13];
               int var18 = 0;

               label84: {
                  while (~var18 > ~var13) {
                     var28[var18] = this.a((byte)76, var4);
                     var18++;
                     if (var19) {
                        break label84;
                     }

                     if (var19) {
                        break;
                     }
                  }

                  var18 = this.a(var10, var16, var11, var12, false);
                  this.fc[var15] = var28;
               }

               label74: {
                  if (0 == var14) {
                     this.Hb[var18] = 0;
                     if (!var19) {
                        break label74;
                     }
                  }

                  this.Hb[var18] = this.Vb;
               }

               var15++;
               if (!var19) {
                  continue;
               }
            }

            this.Yb = 1;
            break;
         }
      } catch (RuntimeException var21) {
         throw i.a(var21, kc[6] + (var1 != null ? kc[1] : kc[3]) + ')');
      }
   }

   private ca(ca[] var1, int var2, boolean var3, boolean var4, boolean var5, boolean var6) {
      this.c = false;
      this.Yb = 1;
      this.b = false;
      this.Ib = 256;
      this.v = false;
      this.cb = false;
      this.Nb = false;
      this.sb = 12345678;
      this.g = 180;
      this.dc = true;
      this.Vb = 12345678;
      this.Kb = false;
      this.rb = -1;
      this.Mb = 512;
      this.Fb = 95;
      this.db = false;
      this.Jb = 32;
      this.hc = 0;

      try {
         this.Nb = var5;
         this.c = var4;
         this.db = var6;
         this.v = var3;
         this.a(0, var1, false, var2);
      } catch (RuntimeException var8) {
         throw i.a(var8, kc[6] + (var1 != null ? kc[1] : kc[3]) + ',' + var2 + ',' + var3 + ',' + var4 + ',' + var5 + ',' + var6 + ')');
      }
   }

   private ca(ca[] var1, int var2) {
      this.c = false;
      this.Yb = 1;
      this.b = false;
      this.Ib = 256;
      this.v = false;
      this.cb = false;
      this.Nb = false;
      this.sb = 12345678;
      this.g = 180;
      this.dc = true;
      this.Vb = 12345678;
      this.Kb = false;
      this.rb = -1;
      this.Mb = 512;
      this.Fb = 95;
      this.db = false;
      this.Jb = 32;
      this.hc = 0;

      try {
         this.a(0, var1, true, var2);
      } catch (RuntimeException var4) {
         throw i.a(var4, kc[6] + (var1 != null ? kc[1] : kc[3]) + ',' + var2 + ')');
      }
   }

   private static char[] z(String var0) {
      char[] var10000 = var0.toCharArray();
      if (var10000.length < 2) {
         var10000[0] = (char)(var10000[0] ^ 37);
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
               var10005 = 27;
               break;
            case 1:
               var10005 = 119;
               break;
            case 2:
               var10005 = 10;
               break;
            case 3:
               var10005 = 61;
               break;
            default:
               var10005 = 37;
         }

         var10001[var1] = (char)(var10004 ^ var10005);
      }

      return new String(var10001).intern();
   }
}
