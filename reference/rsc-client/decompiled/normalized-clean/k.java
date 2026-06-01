final class k {
   private boolean H;
   byte[] Q;
   private lb c;
   private boolean nb;
   static int lb;
   static int z;
   static int ib;
   private int[][] ab;
   private int[] w;
   private byte[][] L;
   static int cb;
   boolean Z;
   static int t;
   private int[][] s;
   private byte[][] R;
   static int u;
   int[][] bb;
   static int N;
   static int j;
   private byte[][] f;
   private byte[][] P;
   static int C;
   private byte[][] mb;
   private ca kb;
   static int h;
   static int n;
   private ca[] F;
   static int jb;
   static int a;
   static int l;
   private ua U;
   static int Y;
   static int d;
   static int M;
   static int r;
   static int p;
   byte[] m;
   private int[][] B;
   static int J;
   int x;
   static int K;
   static int T;
   ca[][] g;
   int[] E;
   static String[] G = new String[100];
   byte[] gb;
   static int k;
   private byte[][] A;
   static int o;
   static int fb;
   static int S;
   static int X;
   static long e = 0L;
   static int b;
   byte[] I;
   static int O;
   static int W;
   static int D;
   static int V;
   static int y;
   static int i;
   private byte[][] eb;
   ca[][] db;
   static int hb;
   static int v;
   int[] q;
   private static final String[] ob = new String[]{
      z(z("\u001e\f\u0017r3")),
      z(z("\u000e\fz\u001df")),
      z(z("\u001bW8_")),
      z(z("\u001e\fhZu\u001cVj\u001b")),
      z(z("\u001e\f\u001cr3")),
      z(z("\u001e\f\u0012r3")),
      z(z("\u001e\f\u0016\u001b")),
      z(z("\u001e\f\u0007\u001b")),
      z(z("\u001e\f\u0015\u001b")),
      z(z("\u001e\f\u0010r3")),
      z(z("\u001e\f\u001dr3")),
      z(z("\u001e\f\u0016r3")),
      z(z("\u001e\f\u0013\u001b")),
      z(z("\u001e\f\u0006\u001b")),
      z(z("\u001e\f\u0011r3")),
      z(z("\u001e\f\u0012\u001b")),
      z(z("\u001e\f\u0019\u001b")),
      z(z("\u001e\f\u0004\u001b")),
      z(z(" L$Rx\u001eK:T;")),
      z(z("\u001e\f\u001f\u001b")),
      z(z("\u001e\f\u0005\u001b")),
      z(z("\u001e\f\u0018r3")),
      z(z("\u001e\f\u0002\u001b")),
      z(z("\u001e\f\u001e\u001b")),
      z(z("\u001e\f\u0001\u001b")),
      z(z("\u001e\f\u0003\u001b")),
      z(z("[J1Z")),
      z(z("\u001e\f\u001c\u001b")),
      z(z("[\f{Tz\u0018G0Ro\u0014\r9Rk\u0006\r")),
      z(z("[F5G")),
      z(z("[N;P")),
      z(z("[H9")),
      z(z("\u001e\f\u001a\u001b")),
      z(z("\u001e\f\u001er3")),
      z(z("\u001e\f\u0013r3")),
      z(z("\u001e\f\u0015r3")),
      z(z("\u001e\f\u0017\u001b")),
      z(z("\u001e\f\u001fr3")),
      z(z("\u001e\f\u001b\u001b")),
      z(z("\u001e\f\u001d\u001b")),
      z(z("\u001bW8_;\u0007M;U:")),
      z(z("\u001e\f\u0018\u001b")),
      z(z("\u001e\f\u0000\u001b")),
      z(z("\u001e\f\u0010\u001b")),
      z(z("\u001e\f\u0011\u001b"))
   };

   private final void a(int var1, int var2, int var3, int var4, int var5, int var6) {
      boolean var9 = client.vh;

      try {
         X++;
         if (var4 != 2) {
            this.Q = (byte[])null;
         }

         ca var7 = this.F[var5 - -(var3 * 8)];
         int var8 = 0;

         while (var7.Db > var8 && !var9) {
            if (~(128 * var6) == ~var7.a[var8] && ~var7.bc[var8] == ~(var1 * 128)) {
               var7.a(var8, var2, (byte)-61);
               return;
            }

            var8++;
            if (var9) {
               break;
            }
         }
      } catch (RuntimeException var10) {
         throw i.a(var10, ob[8] + var1 + ',' + var2 + ',' + var3 + ',' + var4 + ',' + var5 + ',' + var6 + ')');
      }
   }

   private final void a(int var1) {
      boolean var4 = client.vh;

      try {
         O++;
         int var2 = var1;

         label79:
         while (true) {
            int var10000 = 96;
            int var10001 = var2;

            label77:
            while (var10000 > var10001 && !var4) {
               int var3 = 0;

               while (var3 < 96) {
                  var10000 = ~this.b(0, var2, var1 ^ 4, var3);
                  var10001 = -251;
                  if (var4) {
                     continue label77;
                  }

                  label71:
                  if (var10000 == -251) {
                     if (47 != var2 || 250 == this.b(0, var2 - -1, var1 + 4, var3) || -3 == ~this.b(0, 1 + var2, 4, var3)) {
                        if (~var3 == -48 && 250 != this.b(0, var2, 4, var3 - -1) && ~this.b(0, var2, var1 ^ 4, 1 + var3) != -3) {
                           this.e(9, var3, 107, var2);
                           if (!var4) {
                              break label71;
                           }
                        }

                        this.e(2, var3, 110, var2);
                        if (!var4) {
                           break label71;
                        }
                     }

                     this.e(9, var3, var1 + 111, var2);
                  }

                  var3++;
                  if (var4) {
                     break;
                  }
               }

               var2++;
               if (var4) {
                  return;
               }
               continue label79;
            }

            return;
         }
      } catch (RuntimeException var5) {
         throw i.a(var5, ob[32] + var1 + ')');
      }
   }

   final void a(int var1, int var2, int var3, int var4) {
      boolean var10 = client.vh;

      try {
         h++;
         if (var4 != 4081) {
            this.a(-98, 25, -8, -1, (byte)-83, 45);
         }

         if (var2 >= 0 && var3 >= 0 && 95 > var2 && 95 > var3) {
            if (1 == mb.a[var1] || ~mb.a[var1] == -3) {
               int var5;
               int var6;
               int var7;
               label163: {
                  var5 = this.b(var2, var3, -79);
                  if (var5 == 0 || ~var5 == -5) {
                     var7 = ub.g[var1];
                     var6 = f.f[var1];
                     if (!var10) {
                        break label163;
                     }
                  }

                  var7 = f.f[var1];
                  var6 = ub.g[var1];
               }

               int var8 = var2;

               label130:
               while (true) {
                  int var10000 = ~var8;
                  int var10001 = ~(var2 - -var6);

                  label128:
                  while (true) {
                     if (var10000 <= var10001) {
                        break label130;
                     }

                     if (var10) {
                        return;
                     }

                     int var9 = var3;

                     while (var7 + var3 > var9) {
                        var10000 = -2;
                        var10001 = ~mb.a[var1];
                        if (var10) {
                           continue label128;
                        }

                        label122: {
                           if (-2 != var10001) {
                              if (var5 == 0) {
                                 this.bb[var8][var9] = ib.a(this.bb[var8][var9], 65533);
                                 if (var8 <= 0) {
                                    break label122;
                                 }

                                 this.c(var9, var4 + 61454, -1 + var8, 8);
                                 if (!var10) {
                                    break label122;
                                 }
                              }

                              if (~var5 != -3) {
                                 if (~var5 != -5) {
                                    if (-7 != ~var5) {
                                       break label122;
                                    }

                                    this.bb[var8][var9] = ib.a(this.bb[var8][var9], 65534);
                                    if (~var9 >= -1) {
                                       break label122;
                                    }

                                    this.c(-1 + var9, var4 ^ 61454, var8, 4);
                                    if (!var10) {
                                       break label122;
                                    }
                                 }

                                 this.bb[var8][var9] = ib.a(this.bb[var8][var9], 65527);
                                 if (var8 >= 95) {
                                    break label122;
                                 }

                                 this.c(var9, var4 + 61454, 1 + var8, 2);
                                 if (!var10) {
                                    break label122;
                                 }
                              }

                              this.bb[var8][var9] = ib.a(this.bb[var8][var9], 65531);
                              if (~var9 <= -96) {
                                 break label122;
                              }

                              this.c(var9 - -1, var4 + 61454, var8, 1);
                              if (!var10) {
                                 break label122;
                              }
                           }

                           this.bb[var8][var9] = ib.a(this.bb[var8][var9], 65471);
                        }

                        var9++;
                        if (var10) {
                           break;
                        }
                     }

                     var8++;
                     if (var10) {
                        break label130;
                     }
                     break;
                  }
               }

               this.c(var6, var7, -82, var2, var3);
            }
         }
      } catch (RuntimeException var11) {
         throw i.a(var11, ob[43] + var1 + ',' + var2 + ',' + var3 + ',' + var4 + ')');
      }
   }

   private final void e(int var1, int var2, int var3, int var4) {
      boolean var7 = client.vh;

      try {
         v++;
         if (~var4 <= -1 && -97 < ~var4 && ~var2 <= -1 && 96 > var2) {
            byte var5;
            label67: {
               var5 = 0;
               if (var4 < 48 || var2 >= 48) {
                  if (~var4 > -49 && ~var2 <= -49) {
                     var5 = 2;
                     var2 -= 48;
                     if (!var7) {
                        break label67;
                     }
                  }

                  if (~var4 > -49 || var2 < 48) {
                     break label67;
                  }

                  var4 -= 48;
                  var2 -= 48;
                  var5 = 3;
                  if (!var7) {
                     break label67;
                  }
               }

               var5 = 1;
               var4 -= 48;
            }

            int var6 = -76 % ((var3 - 53) / 53);
            this.R[var5][var2 + 48 * var4] = (byte)var1;
         }
      } catch (RuntimeException var8) {
         throw i.a(var8, ob[37] + var1 + ',' + var2 + ',' + var3 + ',' + var4 + ')');
      }
   }

   final void a(int var1, int var2, boolean var3, int var4) {
      boolean var10 = client.vh;

      try {
         if (!var3) {
            u++;
            if (~var1 <= -1 && -1 >= ~var4 && var1 < 95 && ~var4 > -96) {
               if (~mb.a[var2] == -2 || mb.a[var2] == 2) {
                  int var5;
                  int var6;
                  int var7;
                  label162: {
                     var5 = this.b(var1, var4, -107);
                     if (0 == var5 || ~var5 == -5) {
                        var6 = f.f[var2];
                        var7 = ub.g[var2];
                        if (!var10) {
                           break label162;
                        }
                     }

                     var7 = f.f[var2];
                     var6 = ub.g[var2];
                  }

                  int var8 = var1;

                  label128:
                  while (true) {
                     int var10000 = var8;
                     int var10001 = var6 + var1;

                     label126:
                     while (true) {
                        if (var10000 >= var10001) {
                           break label128;
                        }

                        if (var10) {
                           return;
                        }

                        int var9 = var4;

                        while (var4 + var7 > var9) {
                           var10000 = -2;
                           var10001 = ~mb.a[var2];
                           if (var10) {
                              continue label126;
                           }

                           label156: {
                              if (-2 == var10001) {
                                 this.bb[var8][var9] = d.a(this.bb[var8][var9], 64);
                                 if (!var10) {
                                    break label156;
                                 }
                              }

                              if (var5 != 0) {
                                 if (-3 == ~var5) {
                                    this.bb[var8][var9] = d.a(this.bb[var8][var9], 4);
                                    if (-96 >= ~var9) {
                                       break label156;
                                    }

                                    this.a(1, 1 + var9, (byte)-112, var8);
                                    if (!var10) {
                                       break label156;
                                    }
                                 }

                                 if (4 != var5) {
                                    if (6 != var5) {
                                       break label156;
                                    }

                                    this.bb[var8][var9] = d.a(this.bb[var8][var9], 1);
                                    if (~var9 >= -1) {
                                       break label156;
                                    }

                                    this.a(4, -1 + var9, (byte)-112, var8);
                                    if (!var10) {
                                       break label156;
                                    }
                                 }

                                 this.bb[var8][var9] = d.a(this.bb[var8][var9], 8);
                                 if (~var8 <= -96) {
                                    break label156;
                                 }

                                 this.a(2, var9, (byte)-56, var8 + 1);
                                 if (!var10) {
                                    break label156;
                                 }
                              }

                              this.bb[var8][var9] = d.a(this.bb[var8][var9], 2);
                              if (0 < var8) {
                                 this.a(8, var9, (byte)-109, var8 - 1);
                              }
                           }

                           var9++;
                           if (var10) {
                              break;
                           }
                        }

                        var8++;
                        if (var10) {
                           break label128;
                        }
                        break;
                     }
                  }

                  this.c(var6, var7, 94, var1, var4);
               }
            }
         }
      } catch (RuntimeException var11) {
         throw i.a(var11, ob[0] + var1 + ',' + var2 + ',' + var3 + ',' + var4 + ')');
      }
   }

   static final byte[] a(int var0, boolean var1, byte[] var2) {
      try {
         S++;
         if (var0 != 128) {
            o = 104;
         }

         int var3 = ((var2[1] & 255) << 8) + ((0xFF0000 & var2[0] << 16) - -(255 & var2[2]));
         int var4 = ((255 & var2[4]) << 8) + (0xFF0000 & var2[3] << 16) + (255 & var2[5]);
         if (~var3 == ~var4) {
            byte[] var7 = new byte[var2.length - 6];
            ab.a(var2, 6, var7, 0, var7.length);
            return var7;
         }

         if (var1) {
            da.a(ob[18], 0, 0);
         }

         byte[] var5 = new byte[var3];
         ea.a(var5, var3, var2, var4, 6);
         return var5;
      } catch (RuntimeException var6) {
         throw i.a(var6, ob[19] + var0 + ',' + var1 + ',' + (var2 != null ? ob[1] : ob[2]) + ')');
      }
   }

   private final void b(int var1) {
      boolean var4 = client.vh;

      try {
         if (var1 == -10185) {
            if (this.nb) {
               this.c.a(false);
            }

            N++;
            int var2 = 0;

            while (true) {
               if (~var2 > -65) {
                  this.F[var2] = null;
                  if (var4) {
                     break;
                  }

                  int var3 = 0;

                  label53: {
                     while (~var3 > -5) {
                        this.g[var3][var2] = null;
                        var3++;
                        if (var4) {
                           break label53;
                        }

                        if (var4) {
                           break;
                        }
                     }

                     var3 = 0;
                  }

                  label63: {
                     while (-5 < ~var3) {
                        this.db[var3][var2] = null;
                        var3++;
                        if (var4) {
                           break label63;
                        }

                        if (var4) {
                           break;
                        }
                     }

                     var2++;
                  }

                  if (!var4) {
                     continue;
                  }
               }

               System.gc();
               break;
            }
         }
      } catch (RuntimeException var5) {
         throw i.a(var5, ob[12] + var1 + ')');
      }
   }

   private final void a(int var1, int var2, boolean var3, int var4, int var5) {
      boolean var37 = client.vh;

      try {
         i++;
         int var6 = (24 + var1) / 48;
         int var7 = (24 + var5) / 48;
         this.b(var4, 0, var6 - 1, 0, -1 + var7);
         this.b(var4, 1, var6, 0, var7 - 1);
         if (var2 >= 66) {
            this.b(var4, 2, -1 + var6, 0, var7);
            this.b(var4, 3, var6, 0, var7);
            this.a(0);
            if (this.kb == null) {
               this.kb = new ca(18688, 18688, true, true, false, false, true);
            }

            int var8;
            int var10000;
            label1208: {
               label1207: {
                  if (var3) {
                     this.U.a(true);
                     var8 = 0;

                     while (96 > var8) {
                        var10000 = 0;
                        if (var37) {
                           break label1208;
                        }

                        int var9 = 0;

                        label1200: {
                           while (var9 < 96) {
                              this.bb[var8][var9] = 0;
                              var9++;
                              if (var37) {
                                 break label1200;
                              }

                              if (var37) {
                                 break;
                              }
                           }

                           var8++;
                        }

                        if (var37) {
                           break;
                        }
                     }

                     ca var39 = this.kb;
                     var39.c(1);
                     int var40 = 0;

                     label1185: {
                        label1184:
                        while (true) {
                           var10000 = var40;
                           byte var10001 = 96;

                           label1182:
                           while (true) {
                              if (var10000 >= var10001) {
                                 break label1184;
                              }

                              var147 = 0;
                              if (var37) {
                                 break label1185;
                              }

                              int var10 = 0;

                              while (96 > var10) {
                                 int var11 = -this.g(2, var10, var40);
                                 var10000 = ~this.b(var4, var40, 4, var10);
                                 var10001 = -1;
                                 if (var37) {
                                    continue label1182;
                                 }

                                 if (var10000 < -1 && 4 == da.N[-1 + this.b(var4, var40, 4, var10)]) {
                                    var11 = 0;
                                 }

                                 if (this.b(var4, -1 + var40, 4, var10) > 0 && -5 == ~da.N[this.b(var4, var40 - 1, 4, var10) + -1]) {
                                    var11 = 0;
                                 }

                                 if (this.b(var4, var40, 4, var10 + -1) > 0 && -5 == ~da.N[this.b(var4, var40, 4, var10 + -1) + -1]) {
                                    var11 = 0;
                                 }

                                 if (this.b(var4, var40 - 1, 4, var10 - 1) > 0 && 4 == da.N[this.b(var4, -1 + var40, 4, var10 - 1) - 1]) {
                                    var11 = 0;
                                 }

                                 int var12 = var39.e(var40 * 128, 128 * var10, var11, 107);
                                 int var13 = (int)(Math.random() * 10.0) - 5;
                                 var39.a(var12, var13, (byte)-61);
                                 var10++;
                                 if (var37) {
                                    break;
                                 }
                              }

                              var40++;
                              if (var37) {
                                 break label1184;
                              }
                              break;
                           }
                        }

                        var147 = 0;
                     }

                     var40 = var147;

                     label1156: {
                        label1155:
                        while (true) {
                           var10000 = ~var40;
                           int var162 = -96;

                           label1153:
                           while (true) {
                              if (var10000 <= var162) {
                                 break label1155;
                              }

                              var149 = 0;
                              if (var37) {
                                 break label1156;
                              }

                              int var50 = 0;

                              while (var50 < 95) {
                                 int var58 = this.a((byte)104, var40, var50);
                                 int var74 = this.w[var58];
                                 int var82 = var74;
                                 int var14 = var74;
                                 var10000 = -2;
                                 var162 = ~var4;
                                 if (var37) {
                                    continue label1153;
                                 }

                                 if (-2 == var162 || ~var4 == -3) {
                                    var82 = 12345678;
                                    var74 = 12345678;
                                    var14 = 12345678;
                                 }

                                 byte var15 = 0;
                                 if (this.b(var4, var40, 4, var50) > 0) {
                                    int var16 = this.b(var4, var40, 4, var50);
                                    var58 = da.N[var16 + -1];
                                    int var17 = this.d(var4, var50, 15282, var40);
                                    var74 = var82 = qa.K[-1 + var16];
                                    if (~var58 == -5) {
                                       var74 = 1;
                                       var82 = 1;
                                       if (~var16 == -13) {
                                          var74 = 31;
                                          var82 = 31;
                                       }
                                    }

                                    label1224: {
                                       if (~var58 == -6) {
                                          if (0 >= this.c(var40, var50, -49) || -24001 >= ~this.c(var40, var50, -49)) {
                                             break label1224;
                                          }

                                          if (~this.d(-8509, var40 + -1, var14, var4, var50) != -12345679
                                             && -12345679 != ~this.d(-8509, var40, var14, var4, -1 + var50)) {
                                             var15 = 0;
                                             var74 = this.d(-8509, var40 + -1, var14, var4, var50);
                                             if (!var37) {
                                                break label1224;
                                             }
                                          }

                                          if (this.d(-8509, 1 + var40, var14, var4, var50) != 12345678
                                             && 12345678 != this.d(-8509, var40, var14, var4, 1 + var50)) {
                                             var82 = this.d(-8509, var40 - -1, var14, var4, var50);
                                             var15 = 0;
                                             if (!var37) {
                                                break label1224;
                                             }
                                          }

                                          if (~this.d(-8509, 1 + var40, var14, var4, var50) == -12345679
                                             || ~this.d(-8509, var40, var14, var4, -1 + var50) == -12345679) {
                                             if (-12345679 == ~this.d(-8509, var40 - 1, var14, var4, var50)
                                                || this.d(-8509, var40, var14, var4, var50 - -1) == 12345678) {
                                                break label1224;
                                             }

                                             var15 = 1;
                                             var74 = this.d(-8509, var40 - 1, var14, var4, var50);
                                             if (!var37) {
                                                break label1224;
                                             }
                                          }

                                          var82 = this.d(-8509, var40 + 1, var14, var4, var50);
                                          var15 = 1;
                                          if (!var37) {
                                             break label1224;
                                          }
                                       }

                                       label1105:
                                       if (2 != var58 || 0 < this.c(var40, var50, -49) && -24001 < ~this.c(var40, var50, -49)) {
                                          if (var17 != this.d(var4, var50, 15282, -1 + var40) && this.d(var4, -1 + var50, 15282, var40) != var17) {
                                             var74 = var14;
                                             var15 = 0;
                                             if (!var37) {
                                                break label1105;
                                             }
                                          }

                                          if (var17 != this.d(var4, var50, 15282, var40 + 1) && ~var17 != ~this.d(var4, var50 - -1, 15282, var40)) {
                                             var15 = 0;
                                             var82 = var14;
                                             if (!var37) {
                                                break label1105;
                                             }
                                          }

                                          if (var17 == this.d(var4, var50, 15282, 1 + var40) || ~var17 == ~this.d(var4, -1 + var50, 15282, var40)) {
                                             if (var17 == this.d(var4, var50, 15282, -1 + var40) || var17 == this.d(var4, 1 + var50, 15282, var40)) {
                                                break label1105;
                                             }

                                             var74 = var14;
                                             var15 = 1;
                                             if (!var37) {
                                                break label1105;
                                             }
                                          }

                                          var82 = var14;
                                          var15 = 1;
                                       }
                                    }

                                    if (~ac.l[var16 - 1] != -1) {
                                       this.bb[var40][var50] = d.a(this.bb[var40][var50], 64);
                                    }

                                    if (-3 == ~da.N[var16 + -1]) {
                                       this.bb[var40][var50] = d.a(this.bb[var40][var50], 128);
                                    }
                                 }

                                 label1269: {
                                    this.a(var15, (byte)-122, var82, var40, var50, var74);
                                    int var104 = this.g(2, 1 + var50, var40 + 1)
                                       + (-this.g(2, var50, var40 - -1) - -this.g(2, var50 + 1, var40))
                                       - this.g(2, var50, var40);
                                    if (~var82 != ~var74 || -1 != ~var104) {
                                       label1229: {
                                          int[] var112 = new int[3];
                                          int[] var18 = new int[3];
                                          if (-1 == ~var15) {
                                             if (var74 != 12345678) {
                                                var112[1] = var40 * 96 + var50;
                                                var112[0] = 96 + var50 + 96 * var40;
                                                var112[2] = 1 + var50 + 96 * var40;
                                                int var19 = var39.a(3, var112, 12345678, var74, false);
                                                this.q[var19] = var40;
                                                this.E[var19] = var50;
                                                var39.E[var19] = var19 + 200000;
                                             }

                                             if (12345678 == var82) {
                                                break label1229;
                                             }

                                             var18[2] = var50 + (96 * var40 - -96);
                                             var18[1] = 97 + 96 * var40 + var50;
                                             var18[0] = 1 + var40 * 96 + var50;
                                             int var124 = var39.a(3, var18, 12345678, var82, false);
                                             this.q[var124] = var40;
                                             this.E[var124] = var50;
                                             var39.E[var124] = var124 + 200000;
                                             if (!var37) {
                                                break label1229;
                                             }
                                          }

                                          if (12345678 != var74) {
                                             var112[2] = var50 - -(96 * var40);
                                             var112[1] = 96 + 96 * var40 + (var50 - -1);
                                             var112[0] = 1 + 96 * var40 + var50;
                                             int var125 = var39.a(3, var112, 12345678, var74, false);
                                             this.q[var125] = var40;
                                             this.E[var125] = var50;
                                             var39.E[var125] = 200000 - -var125;
                                          }

                                          if (~var82 != -12345679) {
                                             var18[1] = var50 - -(96 * var40);
                                             var18[2] = var50 - (-(var40 * 96) - 97);
                                             var18[0] = 96 * var40 + var50 - -96;
                                             int var126 = var39.a(3, var18, 12345678, var82, false);
                                             this.q[var126] = var40;
                                             this.E[var126] = var50;
                                             var39.E[var126] = var126 + 200000;
                                          }
                                       }

                                       if (!var37) {
                                          break label1269;
                                       }
                                    }

                                    if (~var74 != -12345679) {
                                       int[] var113 = new int[]{
                                          var50 - (-(var40 * 96) + -96), var50 - -(var40 * 96), 1 + var40 * 96 + var50, var50 - (-(var40 * 96) + -96) - -1
                                       };
                                       int var121 = var39.a(4, var113, 12345678, var74, false);
                                       this.q[var121] = var40;
                                       this.E[var121] = var50;
                                       var39.E[var121] = var121 + 200000;
                                    }
                                 }

                                 var50++;
                                 if (var37) {
                                    break;
                                 }
                              }

                              var40++;
                              if (var37) {
                                 break label1155;
                              }
                              break;
                           }
                        }

                        var149 = 1;
                     }

                     var40 = var149;

                     label1059: {
                        label1058:
                        while (true) {
                           var10000 = 95;

                           label1056:
                           while (true) {
                              if (var10000 <= var40) {
                                 break label1058;
                              }

                              var151 = 1;
                              if (var37) {
                                 break label1059;
                              }

                              int var51 = 1;

                              while (var51 < 95) {
                                 var10000 = this.b(var4, var40, 4, var51);
                                 if (var37) {
                                    continue label1056;
                                 }

                                 label1233: {
                                    if (var10000 > 0 && da.N[this.b(var4, var40, 4, var51) + -1] == 4) {
                                       int var60 = qa.K[-1 + this.b(var4, var40, 4, var51)];
                                       int var75 = var39.e(var40 * 128, 128 * var51, -this.g(2, var51, var40), 13);
                                       int var83 = var39.e(128 * (var40 + 1), var51 * 128, -this.g(2, var51, 1 + var40), 107);
                                       int var90 = var39.e((1 + var40) * 128, (var51 - -1) * 128, -this.g(2, var51 + 1, var40 - -1), -116);
                                       int var97 = var39.e(var40 * 128, 128 + 128 * var51, -this.g(2, 1 + var51, var40), -124);
                                       int[] var105 = new int[]{var75, var83, var90, var97};
                                       int var114 = var39.a(4, var105, var60, 12345678, false);
                                       this.q[var114] = var40;
                                       this.E[var114] = var51;
                                       var39.E[var114] = var114 + 200000;
                                       this.a(0, (byte)-121, var60, var40, var51, var60);
                                       if (!var37) {
                                          break label1233;
                                       }
                                    }

                                    if (this.b(var4, var40, 4, var51) == 0 || da.N[this.b(var4, var40, 4, var51) - 1] != 3) {
                                       if (0 < this.b(var4, var40, 4, var51 - -1) && da.N[-1 + this.b(var4, var40, 4, 1 + var51)] == 4) {
                                          int var61 = qa.K[-1 + this.b(var4, var40, 4, var51 + 1)];
                                          int var76 = var39.e(128 * var40, 128 * var51, -this.g(2, var51, var40), -124);
                                          int var84 = var39.e(128 * (var40 + 1), 128 * var51, -this.g(2, var51, 1 + var40), -118);
                                          int var91 = var39.e(128 + var40 * 128, 128 * (var51 - -1), -this.g(2, var51 + 1, 1 + var40), -124);
                                          int var98 = var39.e(128 * var40, 128 * var51 + 128, -this.g(2, 1 + var51, var40), -116);
                                          int[] var106 = new int[]{var76, var84, var91, var98};
                                          int var115 = var39.a(4, var106, var61, 12345678, false);
                                          this.q[var115] = var40;
                                          this.E[var115] = var51;
                                          var39.E[var115] = var115 + 200000;
                                          this.a(0, (byte)34, var61, var40, var51, var61);
                                       }

                                       if (-1 > ~this.b(var4, var40, 4, var51 - 1) && ~da.N[this.b(var4, var40, 4, var51 - 1) - 1] == -5) {
                                          int var62 = qa.K[this.b(var4, var40, 4, -1 + var51) + -1];
                                          int var77 = var39.e(var40 * 128, var51 * 128, -this.g(2, var51, var40), -122);
                                          int var85 = var39.e(128 * (1 + var40), var51 * 128, -this.g(2, var51, var40 + 1), 123);
                                          int var92 = var39.e(128 * (1 + var40), 128 * (var51 - -1), -this.g(2, var51 - -1, var40 + 1), -104);
                                          int var99 = var39.e(128 * var40, 128 + 128 * var51, -this.g(2, var51 - -1, var40), -127);
                                          int[] var107 = new int[]{var77, var85, var92, var99};
                                          int var116 = var39.a(4, var107, var62, 12345678, false);
                                          this.q[var116] = var40;
                                          this.E[var116] = var51;
                                          var39.E[var116] = 200000 + var116;
                                          this.a(0, (byte)17, var62, var40, var51, var62);
                                       }

                                       if (this.b(var4, var40 - -1, 4, var51) > 0 && da.N[-1 + this.b(var4, var40 + 1, 4, var51)] == 4) {
                                          int var63 = qa.K[this.b(var4, 1 + var40, 4, var51) + -1];
                                          int var78 = var39.e(128 * var40, var51 * 128, -this.g(2, var51, var40), -113);
                                          int var86 = var39.e(128 + var40 * 128, var51 * 128, -this.g(2, var51, 1 + var40), 89);
                                          int var93 = var39.e(128 * (1 + var40), 128 * var51 - -128, -this.g(2, 1 + var51, 1 + var40), 124);
                                          int var100 = var39.e(var40 * 128, (1 + var51) * 128, -this.g(2, var51 - -1, var40), -112);
                                          int[] var108 = new int[]{var78, var86, var93, var100};
                                          int var117 = var39.a(4, var108, var63, 12345678, false);
                                          this.q[var117] = var40;
                                          this.E[var117] = var51;
                                          var39.E[var117] = var117 + 200000;
                                          this.a(0, (byte)-124, var63, var40, var51, var63);
                                       }

                                       if (-1 > ~this.b(var4, -1 + var40, 4, var51) && da.N[this.b(var4, -1 + var40, 4, var51) - 1] == 4) {
                                          int var64 = qa.K[this.b(var4, var40 + -1, 4, var51) - 1];
                                          int var79 = var39.e(var40 * 128, 128 * var51, -this.g(2, var51, var40), -123);
                                          int var87 = var39.e((var40 - -1) * 128, 128 * var51, -this.g(2, var51, 1 + var40), 106);
                                          int var94 = var39.e(128 + 128 * var40, var51 * 128 + 128, -this.g(2, 1 + var51, 1 + var40), 56);
                                          int var101 = var39.e(var40 * 128, 128 * (1 + var51), -this.g(2, var51 - -1, var40), 119);
                                          int[] var109 = new int[]{var79, var87, var94, var101};
                                          int var118 = var39.a(4, var109, var64, 12345678, false);
                                          this.q[var118] = var40;
                                          this.E[var118] = var51;
                                          var39.E[var118] = var118 + 200000;
                                          this.a(0, (byte)-127, var64, var40, var51, var64);
                                       }
                                    }
                                 }

                                 var51++;
                                 if (var37) {
                                    break;
                                 }
                              }

                              var40++;
                              if (var37) {
                                 break label1058;
                              }
                              break;
                           }
                        }

                        var39.a(-50, 40, -10, -50, true, 48, 105);
                        this.F = this.kb.a(0, 8, 1536, 112, 64, 233, 1536, false, 0);
                        var151 = 0;
                     }

                     var40 = var151;

                     label1010: {
                        while (~var40 > -65) {
                           this.c.a(this.F[var40], (byte)118);
                           var40++;
                           if (var37) {
                              break label1010;
                           }

                           if (var37) {
                              break;
                           }
                        }

                        var40 = 0;
                     }

                     while (~var40 > -97) {
                        var10000 = 0;
                        if (var37) {
                           break label1207;
                        }

                        int var52 = 0;

                        label1024: {
                           while (~var52 > -97) {
                              this.ab[var40][var52] = this.g(2, var52, var40);
                              var52++;
                              if (var37) {
                                 break label1024;
                              }

                              if (var37) {
                                 break;
                              }
                           }

                           var40++;
                        }

                        if (var37) {
                           break;
                        }
                     }
                  }

                  this.kb.c(1);
                  var10000 = 6316128;
               }

               var8 = var10000;
               var10000 = 0;
            }

            int var44 = var10000;

            label993:
            while (true) {
               var10000 = var44;
               int var163 = 95;

               label990:
               while (var10000 < var163) {
                  var154 = false;
                  if (var37) {
                     break label993;
                  }

                  int var53 = 0;

                  while (~var53 > -96) {
                     int var65 = this.a(var44, (byte)-124, var53);
                     var10000 = 0;
                     var163 = var65;
                     if (var37) {
                        continue label990;
                     }

                     if (0 < var65 && (lb.Tb[var65 + -1] == 0 || this.H)) {
                        this.a(-1 + var65, this.kb, 1 + var44, var53, var44, -14584, var53);
                        if (var3 && u.a[-1 + var65] != 0) {
                           this.bb[var44][var53] = d.a(this.bb[var44][var53], 1);
                           if (var53 > 0) {
                              this.a(4, -1 + var53, (byte)-125, var44);
                           }
                        }

                        if (var3) {
                           this.U.b(3, var8, var44 * 3, var53 * 3, (byte)-109);
                        }
                     }

                     var65 = this.e(95, var44, var53);
                     if (0 < var65 && (~lb.Tb[-1 + var65] == -1 || this.H)) {
                        this.a(var65 + -1, this.kb, var44, var53, var44, -14584, 1 + var53);
                        if (var3 && u.a[var65 + -1] != 0) {
                           this.bb[var44][var53] = d.a(this.bb[var44][var53], 2);
                           if (-1 > ~var44) {
                              this.a(8, var53, (byte)-72, var44 + -1);
                           }
                        }

                        if (var3) {
                           this.U.b(var44 * 3, 3 * var53, var8, 3, 0);
                        }
                     }

                     var65 = this.c(var44, var53, -49);
                     if (0 < var65 && ~var65 > -12001 && (lb.Tb[var65 + -1] == 0 || this.H)) {
                        this.a(-1 + var65, this.kb, var44 - -1, var53, var44, -14584, 1 + var53);
                        if (var3 && 0 != u.a[var65 - 1]) {
                           this.bb[var44][var53] = d.a(this.bb[var44][var53], 32);
                        }

                        if (var3) {
                           this.U.a(3 * var53, var44 * 3, 82, var8);
                           this.U.a(1 + 3 * var53, 1 + var44 * 3, 69, var8);
                           this.U.a(2 + 3 * var53, var44 * 3 - -2, 65, var8);
                        }
                     }

                     if (12000 < var65 && var65 < 24000 && (~lb.Tb[-12001 + var65] == -1 || this.H)) {
                        this.a(-12001 + var65, this.kb, var44, var53, var44 - -1, -14584, 1 + var53);
                        if (var3 && ~u.a[var65 + -12001] != -1) {
                           this.bb[var44][var53] = d.a(this.bb[var44][var53], 16);
                        }

                        if (var3) {
                           this.U.a(3 * var53, 2 + 3 * var44, 116, var8);
                           this.U.a(var53 * 3 + 1, var44 * 3 + 1, 99, var8);
                           this.U.a(2 + 3 * var53, var44 * 3, 90, var8);
                        }
                     }

                     var53++;
                     if (var37) {
                        break;
                     }
                  }

                  var44++;
                  if (!var37) {
                     continue label993;
                  }
                  break;
               }

               var154 = var3;
               break;
            }

            if (var154) {
               this.U.b(285, 0, 0, -27966, -1 + this.x, 285);
            }

            this.kb.a(-50, 60, -10, -50, false, 24, 122);
            this.g[var4] = this.kb.a(0, 8, 1536, -120, 64, 338, 1536, true, 0);
            var44 = 0;

            while (true) {
               if (var44 < 64) {
                  this.c.a(this.g[var4][var44], (byte)118);
                  var44++;
                  if (var37) {
                     break;
                  }

                  if (!var37) {
                     continue;
                  }
               }

               var44 = 0;
               break;
            }

            label937:
            while (true) {
               var10000 = 95;

               label934:
               while (var10000 > var44) {
                  var156 = 0;
                  if (var37) {
                     break label937;
                  }

                  int var54 = 0;

                  while (var54 < 95) {
                     int var68 = this.a(var44, (byte)-111, var54);
                     var10000 = var68;
                     if (var37) {
                        continue label934;
                     }

                     if (var68 > 0) {
                        this.a(-1 + var68, var44 + 1, var54, var54, (byte)-50, var44);
                     }

                     var68 = this.e(61, var44, var54);
                     if (0 < var68) {
                        this.a(var68 + -1, var44, var54, var54 - -1, (byte)-65, var44);
                     }

                     var68 = this.c(var44, var54, -49);
                     if (var68 > 0 && var68 < 12000) {
                        this.a(-1 + var68, var44 - -1, var54, var54 + 1, (byte)-118, var44);
                     }

                     if (-12001 > ~var68 && var68 < 24000) {
                        this.a(var68 + -12001, var44, var54, var54 + 1, (byte)82, var44 - -1);
                     }

                     var54++;
                     if (var37) {
                        break;
                     }
                  }

                  var44++;
                  if (!var37) {
                     continue label937;
                  }
                  break;
               }

               var156 = 1;
               break;
            }

            var44 = var156;

            label909:
            while (true) {
               var10000 = var44;
               byte var164 = 95;

               label906:
               while (var10000 < var164) {
                  var158 = 1;
                  if (var37) {
                     break label909;
                  }

                  int var55 = 1;

                  while (~var55 > -96) {
                     int var71 = this.d(var55, var44, 115);
                     var10000 = ~var71;
                     var164 = -1;
                     if (var37) {
                        continue label906;
                     }

                     label900:
                     if (var10000 < -1) {
                        int var80 = var44;
                        int var88 = var55;
                        int var95 = var44 + 1;
                        int var102 = var55;
                        int var110 = 1 + var44;
                        int var119 = var55 - -1;
                        int var122 = var44;
                        int var127 = 1 + var55;
                        int var20 = 0;
                        int var21 = this.ab[var80][var88];
                        int var22 = this.ab[var95][var102];
                        int var23 = this.ab[var110][var119];
                        if (-80001 > ~var22) {
                           var22 -= 80000;
                        }

                        if (80000 < var21) {
                           var21 -= 80000;
                        }

                        int var24 = this.ab[var122][var127];
                        if (-80001 > ~var23) {
                           var23 -= 80000;
                        }

                        if (80000 < var24) {
                           var24 -= 80000;
                        }

                        if (var21 > var20) {
                           var20 = var21;
                        }

                        if (~var22 < ~var20) {
                           var20 = var22;
                        }

                        if (~var23 < ~var20) {
                           var20 = var23;
                        }

                        if (~var24 < ~var20) {
                           var20 = var24;
                        }

                        if (-80001 >= ~var20) {
                           var20 -= 80000;
                        }

                        label885: {
                           if (~var21 > -80001) {
                              this.ab[var80][var88] = var20;
                              if (!var37) {
                                 break label885;
                              }
                           }

                           this.ab[var80][var88] = this.ab[var80][var88] - 80000;
                        }

                        label880: {
                           if (80000 > var22) {
                              this.ab[var95][var102] = var20;
                              if (!var37) {
                                 break label880;
                              }
                           }

                           this.ab[var95][var102] = this.ab[var95][var102] - 80000;
                        }

                        label875: {
                           if (-80001 < ~var23) {
                              this.ab[var110][var119] = var20;
                              if (!var37) {
                                 break label875;
                              }
                           }

                           this.ab[var110][var119] = this.ab[var110][var119] - 80000;
                        }

                        if (var24 < 80000) {
                           this.ab[var122][var127] = var20;
                           if (!var37) {
                              break label900;
                           }
                        }

                        this.ab[var122][var127] = this.ab[var122][var127] - 80000;
                     }

                     var55++;
                     if (var37) {
                        break;
                     }
                  }

                  var44++;
                  if (!var37) {
                     continue label909;
                  }
                  break;
               }

               this.kb.c(1);
               var158 = 1;
               break;
            }

            var44 = var158;

            label853:
            while (true) {
               var10000 = 95;

               label850:
               while (var10000 > var44) {
                  var160 = 1;
                  if (var37) {
                     break label853;
                  }

                  int var56 = 1;

                  while (95 > var56) {
                     int var72 = this.d(var56, var44, 126);
                     var10000 = var72;
                     if (var37) {
                        continue label850;
                     }

                     label844:
                     if (var72 > 0) {
                        int var81 = var44;
                        int var89 = var56;
                        int var96 = var44 + 1;
                        int var103 = var56;
                        int var111 = var44 - -1;
                        int var120 = 1 + var56;
                        int var123 = var44;
                        int var128 = var56 + 1;
                        int var129 = 128 * var44;
                        int var130 = var56 * 128;
                        int var131 = 128 + var129;
                        int var132 = 128 + var130;
                        int var133 = var129;
                        int var25 = var130;
                        int var26 = var131;
                        int var27 = var132;
                        int var28 = this.ab[var81][var89];
                        int var29 = this.ab[var96][var103];
                        int var30 = this.ab[var111][var120];
                        int var31 = this.ab[var123][var128];
                        int var32 = i.g[-1 + var72];
                        if (this.a(false, var81, var89) && ~var28 > -80001) {
                           var28 += var32 + 80000;
                           this.ab[var81][var89] = var28;
                        }

                        if (this.a(false, var96, var103) && ~var29 > -80001) {
                           var29 += var32 - -80000;
                           this.ab[var96][var103] = var29;
                        }

                        if (this.a(false, var111, var120) && -80001 < ~var30) {
                           var30 += 80000 + var32;
                           this.ab[var111][var120] = var30;
                        }

                        if (-80001 >= ~var29) {
                           var29 -= 80000;
                        }

                        if (80000 <= var30) {
                           var30 -= 80000;
                        }

                        if (this.a(false, var123, var128) && ~var31 > -80001) {
                           var31 += var32 + 80000;
                           this.ab[var123][var128] = var31;
                        }

                        if (var28 >= 80000) {
                           var28 -= 80000;
                        }

                        if (var31 >= 80000) {
                           var31 -= 80000;
                        }

                        byte var33 = 16;
                        if (!this.a(-1 + var81, 26431, var89)) {
                           var129 -= var33;
                        }

                        if (!this.a(var81 - -1, 26431, var89)) {
                           var129 += var33;
                        }

                        if (!this.a(var81, 26431, var89 + -1)) {
                           var130 -= var33;
                        }

                        if (!this.a(var81, 26431, 1 + var89)) {
                           var130 += var33;
                        }

                        if (!this.a(-1 + var96, 26431, var103)) {
                           var131 -= var33;
                        }

                        if (!this.a(var96 + 1, 26431, var103)) {
                           var131 += var33;
                        }

                        if (!this.a(var96, 26431, -1 + var103)) {
                           var25 -= var33;
                        }

                        if (!this.a(var96, 26431, var103 - -1)) {
                           var25 += var33;
                        }

                        if (!this.a(var111 + -1, 26431, var120)) {
                           var26 -= var33;
                        }

                        if (!this.a(1 + var111, 26431, var120)) {
                           var26 += var33;
                        }

                        if (!this.a(var111, 26431, var120 + -1)) {
                           var132 -= var33;
                        }

                        if (!this.a(var111, 26431, 1 + var120)) {
                           var132 += var33;
                        }

                        if (!this.a(var123 - 1, 26431, var128)) {
                           var133 -= var33;
                        }

                        if (!this.a(var123 + 1, 26431, var128)) {
                           var133 += var33;
                        }

                        if (!this.a(var123, 26431, var128 - 1)) {
                           var27 -= var33;
                        }

                        if (!this.a(var123, 26431, var128 - -1)) {
                           var27 += var33;
                        }

                        var29 = -var29;
                        var72 = d.g[-1 + var72];
                        var31 = -var31;
                        var30 = -var30;
                        var28 = -var28;
                        if (~this.c(var44, var56, -49) >= -12001 || this.c(var44, var56, -49) >= 24000 || 0 != this.d(var56 - 1, -1 + var44, 120)) {
                           if (12000 < this.c(var44, var56, -49) && ~this.c(var44, var56, -49) > -24001 && -1 == ~this.d(var56 - -1, 1 + var44, 115)) {
                              int[] var34 = new int[]{
                                 this.kb.e(var129, var130, var28, -128), this.kb.e(var131, var25, var29, -122), this.kb.e(var133, var27, var31, 12)
                              };
                              this.kb.a(3, var34, var72, 12345678, false);
                              if (!var37) {
                                 break label844;
                              }
                           }

                           if (~this.c(var44, var56, -49) >= -1 || this.c(var44, var56, -49) >= 12000 || -1 != ~this.d(var56 - 1, var44 + 1, 111)) {
                              if (0 >= this.c(var44, var56, -49) || ~this.c(var44, var56, -49) <= -12001 || ~this.d(1 + var56, var44 - 1, 103) != -1) {
                                 if (~var28 != ~var29 || ~var31 != ~var30) {
                                    if (var28 != var31 || var30 != var29) {
                                       byte var138 = 1;
                                       if (this.d(-1 + var56, var44 + -1, 117) > 0) {
                                          var138 = 0;
                                       }

                                       if (-1 > ~this.d(var56 - -1, var44 + 1, 110)) {
                                          var138 = 0;
                                       }

                                       label781: {
                                          if (-1 == ~var138) {
                                             int[] var35 = new int[]{
                                                this.kb.e(var131, var25, var29, -114),
                                                this.kb.e(var26, var132, var30, 101),
                                                this.kb.e(var129, var130, var28, -126)
                                             };
                                             this.kb.a(3, var35, var72, 12345678, false);
                                             int[] var36 = new int[]{
                                                this.kb.e(var133, var27, var31, -107),
                                                this.kb.e(var129, var130, var28, 63),
                                                this.kb.e(var26, var132, var30, 44)
                                             };
                                             this.kb.a(3, var36, var72, 12345678, false);
                                             if (!var37) {
                                                break label781;
                                             }
                                          }

                                          int[] var144 = new int[]{
                                             this.kb.e(var129, var130, var28, -112),
                                             this.kb.e(var131, var25, var29, -118),
                                             this.kb.e(var133, var27, var31, 103)
                                          };
                                          this.kb.a(3, var144, var72, 12345678, false);
                                          int[] var145 = new int[]{
                                             this.kb.e(var26, var132, var30, -128), this.kb.e(var133, var27, var31, -119), this.kb.e(var131, var25, var29, 52)
                                          };
                                          this.kb.a(3, var145, var72, 12345678, false);
                                       }

                                       if (!var37) {
                                          break label844;
                                       }
                                    }

                                    int[] var139 = new int[]{
                                       this.kb.e(var133, var27, var31, -104),
                                       this.kb.e(var129, var130, var28, 23),
                                       this.kb.e(var131, var25, var29, 91),
                                       this.kb.e(var26, var132, var30, 13)
                                    };
                                    this.kb.a(4, var139, var72, 12345678, false);
                                    if (!var37) {
                                       break label844;
                                    }
                                 }

                                 int[] var140 = new int[]{
                                    this.kb.e(var129, var130, var28, 78),
                                    this.kb.e(var131, var25, var29, 46),
                                    this.kb.e(var26, var132, var30, -113),
                                    this.kb.e(var133, var27, var31, -125)
                                 };
                                 this.kb.a(4, var140, var72, 12345678, false);
                                 if (!var37) {
                                    break label844;
                                 }
                              }

                              int[] var141 = new int[]{
                                 this.kb.e(var131, var25, var29, 121), this.kb.e(var26, var132, var30, 39), this.kb.e(var129, var130, var28, 73)
                              };
                              this.kb.a(3, var141, var72, 12345678, false);
                              if (!var37) {
                                 break label844;
                              }
                           }

                           int[] var142 = new int[]{
                              this.kb.e(var133, var27, var31, -107), this.kb.e(var129, var130, var28, -122), this.kb.e(var26, var132, var30, 35)
                           };
                           this.kb.a(3, var142, var72, 12345678, false);
                           if (!var37) {
                              break label844;
                           }
                        }

                        int[] var143 = new int[]{
                           this.kb.e(var26, var132, var30, -120), this.kb.e(var133, var27, var31, -116), this.kb.e(var131, var25, var29, 117)
                        };
                        this.kb.a(3, var143, var72, 12345678, false);
                     }

                     var56++;
                     if (var37) {
                        break;
                     }
                  }

                  var44++;
                  if (!var37) {
                     continue label853;
                  }
                  break;
               }

               this.kb.a(-50, 50, -10, -50, true, 50, -98);
               this.db[var4] = this.kb.a(0, 8, 1536, -112, 64, 169, 1536, true, 0);
               var160 = 0;
               break;
            }

            var44 = var160;

            while (true) {
               if (-65 < ~var44) {
                  this.c.a(this.db[var4][var44], (byte)118);
                  var44++;
                  if (var37) {
                     break;
                  }

                  if (!var37) {
                     continue;
                  }
               }

               if (this.db[var4][0] == null) {
                  throw new RuntimeException(ob[40]);
               }
               break;
            }

            var44 = 0;

            label737:
            while (true) {
               var10000 = -97;
               int var165 = ~var44;

               label735:
               while (var10000 < var165 && !var37) {
                  int var57 = 0;

                  while (96 > var57) {
                     var10000 = ~this.ab[var44][var57];
                     var165 = -80001;
                     if (var37) {
                        continue label735;
                     }

                     if (var10000 <= -80001) {
                        this.ab[var44][var57] = this.ab[var44][var57] - 80000;
                     }

                     var57++;
                     if (var37) {
                        break;
                     }
                  }

                  var44++;
                  if (var37) {
                     return;
                  }
                  continue label737;
               }

               return;
            }
         }
      } catch (RuntimeException var38) {
         throw i.a(var38, ob[39] + var1 + ',' + var2 + ',' + var3 + ',' + var4 + ',' + var5 + ')');
      }
   }

   private final int d(int var1, int var2, int var3) {
      boolean var5 = client.vh;

      try {
         a++;
         if (0 <= var2 && 96 > var2 && var1 >= 0 && 96 > var1) {
            byte var4 = 0;
            if (var3 < 99) {
               this.E = (int[])null;
            }

            if (~var2 <= -49 && var1 < 48) {
               var4 = 1;
               var2 -= 48;
               if (!var5) {
                  return this.A[var4][var1 + var2 * 48];
               }
            }

            if (var2 < 48 && var1 >= 48) {
               var4 = 2;
               var1 -= 48;
               if (!var5) {
                  return this.A[var4][var1 + var2 * 48];
               }
            }

            if (48 <= var2 && var1 >= 48) {
               var4 = 3;
               var2 -= 48;
               var1 -= 48;
            }

            return this.A[var4][var1 + var2 * 48];
         } else {
            return 0;
         }
      } catch (RuntimeException var6) {
         throw i.a(var6, ob[17] + var1 + 44 + var2 + 44 + var3 + 41);
      }
   }

   private final void b(int var1, byte var2, int var3, int var4) {
      try {
         lb++;
         int var5 = var3 / 12;
         int var6 = var4 / 12;
         int var7 = (var3 + -1) / 12;
         int var8 = (var4 - 1) / 12;
         this.a(var4, var1, var6, 2, var5, var3);
         if (~var7 != ~var5) {
            this.a(var4, var1, var6, 2, var7, var3);
         }

         if (var8 != var6) {
            this.a(var4, var1, var8, 2, var5, var3);
         }

         if (var7 != var5 && ~var8 != ~var6) {
            this.a(var4, var1, var8, 2, var7, var3);
         }

         if (var2 <= 23) {
            this.c(122, -121, -56, -127, -62);
         }
      } catch (RuntimeException var9) {
         throw i.a(var9, ob[24] + var1 + ',' + var2 + ',' + var3 + ',' + var4 + ')');
      }
   }

   private final int d(int var1, int var2, int var3, int var4) {
      try {
         hb++;
         if (var3 != 15282) {
            this.m = (byte[])null;
         }

         int var5 = this.b(var1, var4, var3 + -15278, var2);
         if (var5 == 0) {
            return -1;
         }

         int var6 = da.N[var5 - 1];
         return 2 != var6 ? 0 : 1;
      } catch (RuntimeException var7) {
         throw i.a(var7, ob[42] + var1 + 44 + var2 + 44 + var3 + 44 + var4 + 41);
      }
   }

   private final void c(int var1, int var2, int var3, int var4) {
      try {
         T++;
         this.bb[var3][var1] = ib.a(this.bb[var3][var1], var2 - var4);
      } catch (RuntimeException var6) {
         throw i.a(var6, ob[44] + var1 + ',' + var2 + ',' + var3 + ',' + var4 + ')');
      }
   }

   private final void a(int var1, int var2, byte var3, int var4) {
      try {
         J++;
         this.bb[var4][var2] = d.a(this.bb[var4][var2], var1);
         if (var3 >= -47) {
            this.eb = (byte[][])null;
         }
      } catch (RuntimeException var6) {
         throw i.a(var6, ob[38] + var1 + ',' + var2 + ',' + var3 + ',' + var4 + ')');
      }
   }

   final void a(ca[] var1, byte var2) {
      boolean var15 = client.vh;

      try {
         j++;
         int var3 = 0;

         int var19;
         label170:
         while (true) {
            var19 = var3;
            int var10001 = 94;

            label167:
            while (var19 < var10001) {
               var19 = 0;
               if (var15) {
                  break label170;
               }

               int var4 = 0;

               label164:
               while (true) {
                  var19 = ~var4;

                  label162:
                  while (true) {
                     if (var19 <= -95) {
                        break label164;
                     }

                     var19 = 48000;
                     var10001 = this.c(var3, var4, -49);
                     if (var15) {
                        continue label167;
                     }

                     if (48000 < var10001 && 60000 > this.c(var3, var4, -49)) {
                        int var5;
                        int var7;
                        int var8;
                        label156: {
                           var5 = this.c(var3, var4, -49) + -48001;
                           int var6 = this.b(var3, var4, -91);
                           if (0 != var6 && var6 != 4) {
                              var7 = ub.g[var5];
                              var8 = f.f[var5];
                              if (!var15) {
                                 break label156;
                              }
                           }

                           var8 = ub.g[var5];
                           var7 = f.f[var5];
                        }

                        this.a(var3, var5, false, var4);
                        ca var9 = var1[fb.f[var5]].a(false, -120, false, false, true);
                        int var10 = 128 * (var7 + var3 + var3) / 2;
                        int var11 = (var8 + var4 + var4) * 128 / 2;
                        var9.a(var10, var11, -this.f(var10, var11, 74), true);
                        var9.g(0, -999999, 0, this.b(var3, var4, -78) * 32);
                        this.c.a(var9, (byte)118);
                        var9.a(48, 48, -10, var2 ^ 9, -50, -50);
                        if (~var7 < -2 || var8 > 1) {
                           int var12 = var3;

                           label146:
                           do {
                              var19 = var3 + var7;
                              var10001 = var12;

                              label143:
                              while (true) {
                                 if (var19 <= var10001) {
                                    break label146;
                                 }

                                 var19 = var4;
                                 if (var15) {
                                    continue label162;
                                 }

                                 int var13 = var4;

                                 while (true) {
                                    if (var8 + var4 <= var13) {
                                       break label143;
                                    }

                                    var19 = var3;
                                    var10001 = var12;
                                    if (var15) {
                                       break;
                                    }

                                    if ((var3 < var12 || ~var13 < ~var4) && var5 == this.c(var12, var13, var2 + 64) + -48001) {
                                       byte var14;
                                       label190: {
                                          var11 = var13;
                                          var10 = var12;
                                          var14 = 0;
                                          if (-49 < ~var10 || var11 >= 48) {
                                             if (48 <= var10 || -49 < ~var11) {
                                                if (48 > var10 || var11 < 48) {
                                                   break label190;
                                                }

                                                var14 = 3;
                                                var11 -= 48;
                                                var10 -= 48;
                                                if (!var15) {
                                                   break label190;
                                                }
                                             }

                                             var14 = 2;
                                             var11 -= 48;
                                             if (!var15) {
                                                break label190;
                                             }
                                          }

                                          var10 -= 48;
                                          var14 = 1;
                                       }

                                       this.s[var14][var10 * 48 + var11] = 0;
                                    }

                                    var13++;
                                    if (var15) {
                                       break label143;
                                    }
                                 }
                              }

                              var12++;
                           } while (!var15);
                        }
                     }

                     var4++;
                     if (var15) {
                        break label164;
                     }
                     break;
                  }
               }

               var3++;
               if (!var15) {
                  continue label170;
               }
               break;
            }

            var19 = var2;
            break;
         }

         if (var19 != -113) {
            this.b(-116, 16, 84);
         }
      } catch (RuntimeException var16) {
         throw i.a(var16, ob[5] + (var1 != null ? ob[1] : ob[2]) + ',' + var2 + ')');
      }
   }

   private final int g(int var1, int var2, int var3) {
      boolean var5 = client.vh;

      try {
         y++;
         if (var3 >= 0 && 96 > var3 && 0 <= var2 && -97 < ~var2) {
            byte var4 = 0;
            if (var1 != 2) {
               return 79;
            }

            if (~var3 <= -49 && var2 < 48) {
               var4 = 1;
               var3 -= 48;
               if (!var5) {
                  return (0xFF & this.L[var4][48 * var3 - -var2]) * 3;
               }
            }

            if (var3 < 48 && -49 >= ~var2) {
               var2 -= 48;
               var4 = 2;
               if (!var5) {
                  return (0xFF & this.L[var4][48 * var3 - -var2]) * 3;
               }
            }

            if (-49 >= ~var3 && var2 >= 48) {
               var4 = 3;
               var2 -= 48;
               var3 -= 48;
            }

            return (0xFF & this.L[var4][48 * var3 - -var2]) * 3;
         } else {
            return 0;
         }
      } catch (RuntimeException var6) {
         throw i.a(var6, ob[7] + var1 + 44 + var2 + 44 + var3 + 41);
      }
   }

   final void a(boolean var1, int var2, int var3, int var4, int var5) {
      boolean var7 = client.vh;

      try {
         r++;
         if (0 <= var4 && 0 <= var3 && -96 < ~var4 && 95 > var3) {
            if (1 == u.a[var5]) {
               label84: {
                  if (var2 == 0) {
                     this.bb[var4][var3] = ib.a(this.bb[var4][var3], 65534);
                     if (var3 <= 0) {
                        break label84;
                     }

                     this.c(-1 + var3, 65535, var4, 4);
                     if (!var7) {
                        break label84;
                     }
                  }

                  if (-2 == ~var2) {
                     this.bb[var4][var3] = ib.a(this.bb[var4][var3], 65533);
                     if (~var4 >= -1) {
                        break label84;
                     }

                     this.c(var3, 65535, -1 + var4, 8);
                     if (!var7) {
                        break label84;
                     }
                  }

                  if (2 == var2) {
                     this.bb[var4][var3] = ib.a(this.bb[var4][var3], 65519);
                     if (!var7) {
                        break label84;
                     }
                  }

                  if (var2 == 3) {
                     this.bb[var4][var3] = ib.a(this.bb[var4][var3], 65503);
                  }
               }

               this.c(1, 1, -59, var4, var3);
            }

            if (!var1) {
               this.U = (ua)null;
            }
         }
      } catch (RuntimeException var8) {
         throw i.a(var8, ob[10] + var1 + ',' + var2 + ',' + var3 + ',' + var4 + ',' + var5 + ')');
      }
   }

   final void a(int var1, byte var2, int var3, int var4) {
      try {
         this.b(var2 ^ 10129);
         jb++;
         int var5 = (24 + var1) / 48;
         if (var2 != -90) {
            this.c(58, -126, -4);
         }

         this.a(var1, 122, true, var4, var3);
         int var6 = (24 + var3) / 48;
         if (var4 == 0) {
            this.a(var1, 112, false, 1, var3);
            this.a(var1, var2 ^ -29, false, 2, var3);
            this.b(var4, 0, var5 - 1, 0, -1 + var6);
            this.b(var4, 1, var5, var2 + 90, var6 + -1);
            this.b(var4, 2, var5 + -1, 0, var6);
            this.b(var4, 3, var5, var2 + 90, var6);
            this.a(0);
         }
      } catch (RuntimeException var7) {
         throw i.a(var7, ob[41] + var1 + ',' + var2 + ',' + var3 + ',' + var4 + ')');
      }
   }

   private final void a(int var1, int var2, int var3, int var4, byte var5, int var6) {
      try {
         fb++;
         int var7 = ib.d[var1];
         if (~this.ab[var6][var3] > -80001) {
            this.ab[var6][var3] = this.ab[var6][var3] + var7 + 80000;
         }

         if (80000 > this.ab[var2][var4]) {
            this.ab[var2][var4] = this.ab[var2][var4] + var7 + 80000;
         }

         int var8 = -39 / ((2 - var5) / 51);
      } catch (RuntimeException var9) {
         throw i.a(var9, ob[35] + var1 + ',' + var2 + ',' + var3 + ',' + var4 + ',' + var5 + ',' + var6 + ')');
      }
   }

   final int a(int[] var1, int var2, byte var3, int var4, int[] var5, int var6, int var7, int var8, int var9, boolean var10) {
      boolean var19 = client.vh;

      try {
         int var11 = 0;

         byte var10000;
         while (true) {
            if (-97 < ~var11) {
               var10000 = 0;
               if (var19) {
                  break;
               }

               int var12 = 0;

               label420: {
                  while (var12 < 96) {
                     this.B[var11][var12] = 0;
                     var12++;
                     if (var19) {
                        break label420;
                     }

                     if (var19) {
                        break;
                     }
                  }

                  var11++;
               }

               if (!var19) {
                  continue;
               }
            }

            cb++;
            var11 = 0;
            var10000 = 0;
            break;
         }

         int var22 = var10000;
         int var13 = var6;
         int var14 = var7;
         this.B[var6][var7] = 99;
         var1[var11] = var6;
         var5[var11++] = var7;
         int var15 = var1.length;
         boolean var16 = false;

         int var27;
         while (true) {
            label403:
            if (~var11 != ~var22) {
               var13 = var1[var22];
               var14 = var5[var22];
               var22 = (1 + var22) % var15;
               if (~var2 >= ~var13 && ~var8 <= ~var13 && var9 <= var14 && ~var14 >= ~var4) {
                  var16 = true;
                  if (!var19) {
                     break label403;
                  }
               }

               if (var10) {
                  if (-1 > ~var13 && var2 <= var13 + -1 && ~(var13 + -1) >= ~var8 && var9 <= var14 && var4 >= var14 && (this.bb[var13 - 1][var14] & 8) == 0) {
                     var16 = true;
                     if (!var19) {
                        break label403;
                     }
                  }

                  if (~var13 > -96
                     && ~var2 >= ~(1 + var13)
                     && var13 - -1 <= var8
                     && ~var9 >= ~var14
                     && ~var14 >= ~var4
                     && -1 == ~(2 & this.bb[var13 + 1][var14])) {
                     var16 = true;
                     if (!var19) {
                        break label403;
                     }
                  }

                  if (0 < var14
                     && ~var13 <= ~var2
                     && var8 >= var13
                     && ~var9 >= ~(var14 + -1)
                     && ~(-1 + var14) >= ~var4
                     && ~(4 & this.bb[var13][var14 + -1]) == -1) {
                     var16 = true;
                     if (!var19) {
                        break label403;
                     }
                  }

                  if (var14 < 95
                     && ~var13 <= ~var2
                     && var13 <= var8
                     && ~(var14 - -1) <= ~var9
                     && ~(var14 + 1) >= ~var4
                     && -1 == ~(1 & this.bb[var13][var14 - -1])) {
                     var16 = true;
                     if (!var19) {
                        break label403;
                     }
                  }
               }

               if (0 < var13 && ~this.B[-1 + var13][var14] == -1 && (this.bb[-1 + var13][var14] & 120) == 0) {
                  var1[var11] = -1 + var13;
                  var5[var11] = var14;
                  this.B[-1 + var13][var14] = 2;
                  var11 = (var11 + 1) % var15;
               }

               if (95 > var13 && this.B[1 + var13][var14] == 0 && (this.bb[1 + var13][var14] & 114) == 0) {
                  var1[var11] = 1 + var13;
                  var5[var11] = var14;
                  this.B[var13 + 1][var14] = 8;
                  var11 = (1 + var11) % var15;
               }

               if (var14 > 0 && 0 == this.B[var13][-1 + var14] && -1 == ~(116 & this.bb[var13][var14 + -1])) {
                  var1[var11] = var13;
                  var5[var11] = var14 + -1;
                  this.B[var13][-1 + var14] = 1;
                  var11 = (var11 + 1) % var15;
               }

               if (~var14 > -96 && this.B[var13][1 + var14] == 0 && -1 == ~(113 & this.bb[var13][1 + var14])) {
                  var1[var11] = var13;
                  var5[var11] = var14 - -1;
                  this.B[var13][var14 - -1] = 4;
                  var11 = (var11 - -1) % var15;
               }

               if (~var13 < -1
                  && ~var14 < -1
                  && (116 & this.bb[var13][-1 + var14]) == 0
                  && -1 == ~(120 & this.bb[var13 - 1][var14])
                  && 0 == (124 & this.bb[var13 - 1][var14 + -1])
                  && 0 == this.B[-1 + var13][-1 + var14]) {
                  var1[var11] = -1 + var13;
                  var5[var11] = var14 - 1;
                  this.B[-1 + var13][var14 + -1] = 3;
                  var11 = (1 + var11) % var15;
               }

               if (~var13 > -96
                  && ~var14 < -1
                  && 0 == (this.bb[var13][var14 + -1] & 116)
                  && 0 == (this.bb[1 + var13][var14] & 114)
                  && ~(this.bb[var13 - -1][-1 + var14] & 118) == -1
                  && 0 == this.B[1 + var13][var14 + -1]) {
                  var1[var11] = 1 + var13;
                  var5[var11] = -1 + var14;
                  this.B[var13 - -1][-1 + var14] = 9;
                  var11 = (1 + var11) % var15;
               }

               if (~var13 < -1
                  && ~var14 > -96
                  && -1 == ~(this.bb[var13][1 + var14] & 113)
                  && ~(this.bb[var13 + -1][var14] & 120) == -1
                  && ~(this.bb[var13 - 1][1 + var14] & 121) == -1
                  && -1 == ~this.B[var13 + -1][1 + var14]) {
                  var1[var11] = -1 + var13;
                  var5[var11] = 1 + var14;
                  var11 = (1 + var11) % var15;
                  this.B[-1 + var13][var14 - -1] = 6;
               }

               if (var13 >= 95) {
                  continue;
               }

               var10000 = (byte)-96;
               var27 = ~var14;
               if (var19) {
                  break;
               }

               if (-96 >= var27
                  || 0 != (113 & this.bb[var13][1 + var14])
                  || 0 != (this.bb[var13 + 1][var14] & 114)
                  || 0 != (115 & this.bb[var13 - -1][1 + var14])
                  || this.B[var13 - -1][1 + var14] != 0) {
                  continue;
               }

               var1[var11] = 1 + var13;
               var5[var11] = 1 + var14;
               this.B[1 + var13][1 + var14] = 12;
               var11 = (var11 - -1) % var15;
               if (!var19) {
                  continue;
               }
            }

            var10000 = var3;
            var27 = -48;
            break;
         }

         if (var10000 > var27) {
            return -42;
         }

         if (!var16) {
            return -1;
         }

         var22 = 0;
         var1[var22] = var13;
         var5[var22++] = var14;
         int var18;
         int var17 = var18 = this.B[var13][var14];

         do {
            label331: {
               if (~var6 == ~var13) {
                  var10000 = ~var7;
                  var27 = ~var14;
                  if (var19) {
                     break label331;
                  }

                  if (var10000 == var27) {
                     break;
                  }
               }

               var10000 = var18;
               var27 = var17;
            }

            if (var10000 != var27) {
               var18 = var17;
               var1[var22] = var13;
               var5[var22++] = var14;
            }

            label324: {
               if ((var17 & 1) == 0) {
                  if (-1 == ~(4 & var17)) {
                     break label324;
                  }

                  var14--;
                  if (!var19) {
                     break label324;
                  }
               }

               var14++;
            }

            label317: {
               if (~(2 & var17) == -1) {
                  if (-1 == ~(var17 & 8)) {
                     break label317;
                  }

                  var13--;
                  if (!var19) {
                     break label317;
                  }
               }

               var13++;
            }

            var17 = this.B[var13][var14];
         } while (!var19);

         return var22;
      } catch (RuntimeException var20) {
         throw i.a(
            var20,
            ob[20]
               + (var1 != null ? ob[1] : ob[2])
               + 44
               + var2
               + 44
               + var3
               + 44
               + var4
               + 44
               + (var5 != null ? ob[1] : ob[2])
               + 44
               + var6
               + 44
               + var7
               + 44
               + var8
               + 44
               + var9
               + 44
               + var10
               + 41
         );
      }
   }

   final int f(int var1, int var2, int var3) {
      try {
         k++;
         int var4 = var1 >> 7;
         int var5 = var2 >> 7;
         int var8 = 79 / ((var3 - 30) / 35);
         int var6 = 127 & var1;
         int var7 = 127 & var2;
         if (~var4 <= -1 && 0 <= var5 && 95 > var4 && 95 > var5) {
            if (var6 <= 128 + -var7) {
               int var9 = this.g(2, var5, var4);
               int var10 = this.g(2, var5, 1 + var4) + -var9;
               int var11 = -var9 + this.g(2, 1 + var5, var4);
               if (!client.vh) {
                  return var11 * var7 / 128 + (var9 - -(var10 * var6 / 128));
               }
            }

            int var16 = this.g(2, var5 - -1, 1 + var4);
            int var17 = -var16 + this.g(2, var5 + 1, var4);
            int var18 = -var16 + this.g(2, var5, 1 + var4);
            var6 = -var6 + 128;
            var7 = 128 + -var7;
            return var18 * var7 / 128 + (var16 - -(var17 * var6 / 128));
         } else {
            return 0;
         }
      } catch (RuntimeException var13) {
         throw i.a(var13, ob[34] + var1 + 44 + var2 + 44 + var3 + 41);
      }
   }

   private final void c(int var1, int var2, int var3, int var4, int var5) {
      boolean var9 = client.vh;

      try {
         int var6 = 116 % ((var3 - 15) / 39);
         D++;
         if (1 <= var4 && -2 >= ~var5 && 96 > var1 + var4 && 96 > var2 + var5) {
            int var7 = var4;

            label74:
            while (true) {
               int var10000 = var7;

               label72:
               while (var10000 <= var1 + var4 && !var9) {
                  int var8 = var5;

                  while (~var8 >= ~(var5 - -var2)) {
                     var10000 = 99 & this.b((byte)-38, var8, var7);
                     if (var9) {
                        continue label72;
                     }

                     label66: {
                        if (var10000 == 0
                           && ~(89 & this.b((byte)-38, var8, -1 + var7)) == -1
                           && (this.b((byte)-38, var8 - 1, var7) & 86) == 0
                           && ~(this.b((byte)-38, -1 + var8, var7 - 1) & 108) == -1) {
                           this.b(0, (byte)118, var7, var8);
                           if (!var9) {
                              break label66;
                           }
                        }

                        this.b(35, (byte)50, var7, var8);
                     }

                     var8++;
                     if (var9) {
                        break;
                     }
                  }

                  var7++;
                  if (var9) {
                     return;
                  }
                  continue label74;
               }

               return;
            }
         }
      } catch (RuntimeException var10) {
         throw i.a(var10, ob[6] + var1 + ',' + var2 + ',' + var3 + ',' + var4 + ',' + var5 + ')');
      }
   }

   private final int a(byte var1, int var2, int var3) {
      boolean var5 = client.vh;

      try {
         n++;
         if (var2 >= 0 && 96 > var2 && -1 >= ~var3 && ~var3 > -97) {
            byte var4 = 0;
            if (var1 != 104) {
               return -59;
            }

            if (~var2 <= -49 && 48 > var3) {
               var2 -= 48;
               var4 = 1;
               if (!var5) {
                  return this.eb[var4][var3 + 48 * var2] & 0xFF;
               }
            }

            if (-49 < ~var2 && ~var3 <= -49) {
               var4 = 2;
               var3 -= 48;
               if (!var5) {
                  return this.eb[var4][var3 + 48 * var2] & 0xFF;
               }
            }

            if (-49 >= ~var2 && 48 <= var3) {
               var2 -= 48;
               var4 = 3;
               var3 -= 48;
            }

            return this.eb[var4][var3 + 48 * var2] & 0xFF;
         } else {
            return 0;
         }
      } catch (RuntimeException var6) {
         throw i.a(var6, ob[22] + var1 + 44 + var2 + 44 + var3 + 41);
      }
   }

   private final boolean a(int var1, int var2, int var3) {
      try {
         if (var2 != 26431) {
            return false;
         }

         t++;
         return ~this.d(var3, var1, 119) < -1
            || ~this.d(var3, -1 + var1, 110) < -1
            || 0 < this.d(var3 + -1, var1 + -1, 109)
            || -1 > ~this.d(-1 + var3, var1, var2 + -26318);
      } catch (RuntimeException var5) {
         throw i.a(var5, ob[9] + var1 + ',' + var2 + ',' + var3 + ')');
      }
   }

   private final int d(int var1, int var2, int var3, int var4, int var5) {
      try {
         if (var1 != -8509) {
            return 58;
         }

         p++;
         int var6 = this.b(var4, var2, var1 + 8513, var5);
         return -1 != ~var6 ? qa.K[-1 + var6] : var3;
      } catch (RuntimeException var7) {
         throw i.a(var7, ob[16] + var1 + 44 + var2 + 44 + var3 + 44 + var4 + 44 + var5 + 41);
      }
   }

   private final int c(int var1, int var2, int var3) {
      boolean var5 = client.vh;

      try {
         z++;
         if (0 <= var1 && var1 < 96 && -1 >= ~var2 && ~var2 > -97) {
            byte var4 = 0;
            if (~var1 > var3 || ~var2 <= -49) {
               if (-49 >= ~var1 || ~var2 > -49) {
                  if (~var1 > -49 || ~var2 > -49) {
                     return this.s[var4][var1 * 48 + var2];
                  }

                  var2 -= 48;
                  var4 = 3;
                  var1 -= 48;
                  if (!var5) {
                     return this.s[var4][var1 * 48 + var2];
                  }
               }

               var4 = 2;
               var2 -= 48;
               if (!var5) {
                  return this.s[var4][var1 * 48 + var2];
               }
            }

            var1 -= 48;
            var4 = 1;
            return this.s[var4][var1 * 48 + var2];
         } else {
            return 0;
         }
      } catch (RuntimeException var6) {
         throw i.a(var6, ob[4] + var1 + 44 + var2 + 44 + var3 + 41);
      }
   }

   private final void b(int param1, int param2, int param3, int param4, int param5) {
      // $VF: Couldn't be decompiled
      // Please report this to the Vineflower issue tracker, at https://github.com/Vineflower/vineflower/issues with a copy of the class file (if you have the rights to distribute it!)
      // java.lang.RuntimeException: parsing failure!
      //   at org.jetbrains.java.decompiler.modules.decompiler.decompose.DomHelper.parseGraph(DomHelper.java:235)
      //   at org.jetbrains.java.decompiler.main.rels.MethodProcessor.codeToJava(MethodProcessor.java:174)
      //
      // Bytecode:
      // 000: getstatic client.vh Z
      // 003: istore 13
      // 005: iload 4
      // 007: ifeq 00b
      // 00a: return
      // 00b: getstatic k.l I
      // 00e: bipush 1
      // 00f: iadd
      // 010: putstatic k.l I
      // 013: new java/lang/StringBuilder
      // 016: dup
      // 017: invokespecial java/lang/StringBuilder.<init> ()V
      // 01a: ldc_w "m"
      // 01d: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 020: iload 1
      // 021: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 024: iload 3
      // 025: bipush 10
      // 027: idiv
      // 028: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 02b: iload 3
      // 02c: bipush 10
      // 02e: irem
      // 02f: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 032: iload 5
      // 034: bipush 10
      // 036: idiv
      // 037: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 03a: iload 5
      // 03c: bipush 10
      // 03e: irem
      // 03f: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 042: invokevirtual java/lang/StringBuilder.toString ()Ljava/lang/String;
      // 045: astore 6
      // 047: aconst_null
      // 048: aload 0
      // 049: getfield k.Q [B
      // 04c: if_acmpeq 610
      // 04f: new java/lang/StringBuilder
      // 052: dup
      // 053: invokespecial java/lang/StringBuilder.<init> ()V
      // 056: aload 6
      // 058: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 05b: getstatic k.ob [Ljava/lang/String;
      // 05e: bipush 26
      // 060: aaload
      // 061: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 064: invokevirtual java/lang/StringBuilder.toString ()Ljava/lang/String;
      // 067: bipush 0
      // 068: aload 0
      // 069: getfield k.Q [B
      // 06c: bipush -126
      // 06e: invokestatic na.a (Ljava/lang/String;I[BI)[B
      // 071: astore 7
      // 073: aconst_null
      // 074: aload 7
      // 076: if_acmpne 0ac
      // 079: aload 0
      // 07a: getfield k.I [B
      // 07d: ifnonnull 088
      // 080: goto 084
      // 083: athrow
      // 084: goto 0ac
      // 087: athrow
      // 088: new java/lang/StringBuilder
      // 08b: dup
      // 08c: invokespecial java/lang/StringBuilder.<init> ()V
      // 08f: aload 6
      // 091: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 094: getstatic k.ob [Ljava/lang/String;
      // 097: bipush 26
      // 099: aaload
      // 09a: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 09d: invokevirtual java/lang/StringBuilder.toString ()Ljava/lang/String;
      // 0a0: bipush 0
      // 0a1: aload 0
      // 0a2: getfield k.I [B
      // 0a5: bipush -125
      // 0a7: invokestatic na.a (Ljava/lang/String;I[BI)[B
      // 0aa: astore 7
      // 0ac: aconst_null
      // 0ad: aload 7
      // 0af: if_acmpeq 299
      // 0b2: bipush -1
      // 0b3: aload 7
      // 0b5: arraylength
      // 0b6: bipush -1
      // 0b7: ixor
      // 0b8: if_icmple 299
      // 0bb: goto 0bf
      // 0be: athrow
      // 0bf: bipush 0
      // 0c0: istore 8
      // 0c2: bipush 0
      // 0c3: istore 9
      // 0c5: bipush 0
      // 0c6: istore 10
      // 0c8: sipush 2304
      // 0cb: iload 10
      // 0cd: if_icmple 143
      // 0d0: aload 7
      // 0d2: iload 8
      // 0d4: iinc 8 1
      // 0d7: baload
      // 0d8: sipush 255
      // 0db: iand
      // 0dc: istore 11
      // 0de: sipush -129
      // 0e1: iload 11
      // 0e3: bipush -1
      // 0e4: ixor
      // 0e5: iload 13
      // 0e7: ifne 150
      // 0ea: if_icmpge 104
      // 0ed: goto 0f1
      // 0f0: athrow
      // 0f1: iload 11
      // 0f3: istore 9
      // 0f5: aload 0
      // 0f6: getfield k.L [[B
      // 0f9: iload 2
      // 0fa: aaload
      // 0fb: iload 10
      // 0fd: iinc 10 1
      // 100: iload 11
      // 102: i2b
      // 103: bastore
      // 104: iload 11
      // 106: sipush 128
      // 109: if_icmpge 110
      // 10c: goto 13e
      // 10f: athrow
      // 110: bipush 0
      // 111: istore 12
      // 113: iload 11
      // 115: sipush 128
      // 118: isub
      // 119: bipush -1
      // 11a: ixor
      // 11b: iload 12
      // 11d: bipush -1
      // 11e: ixor
      // 11f: if_icmpge 13e
      // 122: aload 0
      // 123: getfield k.L [[B
      // 126: iload 2
      // 127: aaload
      // 128: iload 10
      // 12a: iinc 10 1
      // 12d: iload 9
      // 12f: i2b
      // 130: bastore
      // 131: iinc 12 1
      // 134: iload 13
      // 136: ifne 0c8
      // 139: iload 13
      // 13b: ifeq 113
      // 13e: iload 13
      // 140: ifeq 0c8
      // 143: bipush 64
      // 145: istore 9
      // 147: bipush 0
      // 148: istore 10
      // 14a: iload 10
      // 14c: bipush -1
      // 14d: ixor
      // 14e: bipush -49
      // 150: if_icmple 1ab
      // 153: bipush 0
      // 154: iload 13
      // 156: ifne 1af
      // 159: goto 15d
      // 15c: athrow
      // 15d: istore 11
      // 15f: iload 11
      // 161: bipush 48
      // 163: if_icmpge 1a3
      // 166: bipush 127
      // 168: iload 9
      // 16a: aload 0
      // 16b: getfield k.L [[B
      // 16e: iload 2
      // 16f: aaload
      // 170: iload 11
      // 172: bipush 48
      // 174: imul
      // 175: iload 10
      // 177: ineg
      // 178: isub
      // 179: baload
      // 17a: iadd
      // 17b: iand
      // 17c: istore 9
      // 17e: aload 0
      // 17f: getfield k.L [[B
      // 182: iload 2
      // 183: aaload
      // 184: iload 10
      // 186: iload 11
      // 188: bipush 48
      // 18a: imul
      // 18b: iadd
      // 18c: iload 9
      // 18e: bipush 2
      // 18f: imul
      // 190: i2b
      // 191: bastore
      // 192: iinc 11 1
      // 195: iload 13
      // 197: ifne 1a6
      // 19a: iload 13
      // 19c: ifeq 15f
      // 19f: goto 1a3
      // 1a2: athrow
      // 1a3: iinc 10 1
      // 1a6: iload 13
      // 1a8: ifeq 14a
      // 1ab: bipush 0
      // 1ac: istore 9
      // 1ae: bipush 0
      // 1af: istore 10
      // 1b1: iload 10
      // 1b3: sipush 2304
      // 1b6: if_icmpge 22b
      // 1b9: sipush 255
      // 1bc: aload 7
      // 1be: iload 8
      // 1c0: iinc 8 1
      // 1c3: baload
      // 1c4: iand
      // 1c5: istore 11
      // 1c7: iload 11
      // 1c9: bipush -1
      // 1ca: ixor
      // 1cb: sipush -129
      // 1ce: iload 13
      // 1d0: ifne 238
      // 1d3: if_icmple 1ed
      // 1d6: goto 1da
      // 1d9: athrow
      // 1da: iload 11
      // 1dc: istore 9
      // 1de: aload 0
      // 1df: getfield k.eb [[B
      // 1e2: iload 2
      // 1e3: aaload
      // 1e4: iload 10
      // 1e6: iinc 10 1
      // 1e9: iload 11
      // 1eb: i2b
      // 1ec: bastore
      // 1ed: iload 11
      // 1ef: sipush 128
      // 1f2: if_icmpge 1f9
      // 1f5: goto 226
      // 1f8: athrow
      // 1f9: bipush 0
      // 1fa: istore 12
      // 1fc: iload 12
      // 1fe: bipush -1
      // 1ff: ixor
      // 200: bipush -128
      // 202: iload 11
      // 204: iadd
      // 205: bipush -1
      // 206: ixor
      // 207: if_icmple 226
      // 20a: aload 0
      // 20b: getfield k.eb [[B
      // 20e: iload 2
      // 20f: aaload
      // 210: iload 10
      // 212: iinc 10 1
      // 215: iload 9
      // 217: i2b
      // 218: bastore
      // 219: iinc 12 1
      // 21c: iload 13
      // 21e: ifne 1b1
      // 221: iload 13
      // 223: ifeq 1fc
      // 226: iload 13
      // 228: ifeq 1b1
      // 22b: bipush 35
      // 22d: istore 9
      // 22f: bipush 0
      // 230: istore 10
      // 232: bipush -49
      // 234: iload 10
      // 236: bipush -1
      // 237: ixor
      // 238: if_icmpge 294
      // 23b: bipush 0
      // 23c: iload 13
      // 23e: ifne 331
      // 241: goto 245
      // 244: athrow
      // 245: istore 11
      // 247: bipush -49
      // 249: iload 11
      // 24b: bipush -1
      // 24c: ixor
      // 24d: if_icmpge 28c
      // 250: bipush 127
      // 252: iload 9
      // 254: aload 0
      // 255: getfield k.eb [[B
      // 258: iload 2
      // 259: aaload
      // 25a: iload 10
      // 25c: bipush 48
      // 25e: iload 11
      // 260: imul
      // 261: iadd
      // 262: baload
      // 263: iadd
      // 264: iand
      // 265: istore 9
      // 267: aload 0
      // 268: getfield k.eb [[B
      // 26b: iload 2
      // 26c: aaload
      // 26d: iload 10
      // 26f: bipush 48
      // 271: iload 11
      // 273: imul
      // 274: iadd
      // 275: bipush 2
      // 276: iload 9
      // 278: imul
      // 279: i2b
      // 27a: bastore
      // 27b: iinc 11 1
      // 27e: iload 13
      // 280: ifne 28f
      // 283: iload 13
      // 285: ifeq 247
      // 288: goto 28c
      // 28b: athrow
      // 28c: iinc 10 1
      // 28f: iload 13
      // 291: ifeq 232
      // 294: iload 13
      // 296: ifeq 2cb
      // 299: bipush 0
      // 29a: istore 8
      // 29c: iload 8
      // 29e: bipush -1
      // 29f: ixor
      // 2a0: sipush -2305
      // 2a3: if_icmple 2cb
      // 2a6: aload 0
      // 2a7: getfield k.L [[B
      // 2aa: iload 2
      // 2ab: aaload
      // 2ac: iload 8
      // 2ae: bipush 0
      // 2af: bastore
      // 2b0: aload 0
      // 2b1: getfield k.eb [[B
      // 2b4: iload 2
      // 2b5: aaload
      // 2b6: iload 8
      // 2b8: bipush 0
      // 2b9: bastore
      // 2ba: iinc 8 1
      // 2bd: iload 13
      // 2bf: ifne 2ef
      // 2c2: iload 13
      // 2c4: ifeq 29c
      // 2c7: goto 2cb
      // 2ca: athrow
      // 2cb: new java/lang/StringBuilder
      // 2ce: dup
      // 2cf: invokespecial java/lang/StringBuilder.<init> ()V
      // 2d2: aload 6
      // 2d4: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 2d7: getstatic k.ob [Ljava/lang/String;
      // 2da: bipush 29
      // 2dc: aaload
      // 2dd: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 2e0: invokevirtual java/lang/StringBuilder.toString ()Ljava/lang/String;
      // 2e3: bipush 0
      // 2e4: aload 0
      // 2e5: getfield k.gb [B
      // 2e8: bipush -125
      // 2ea: invokestatic na.a (Ljava/lang/String;I[BI)[B
      // 2ed: astore 7
      // 2ef: aload 7
      // 2f1: ifnonnull 327
      // 2f4: aconst_null
      // 2f5: aload 0
      // 2f6: getfield k.m [B
      // 2f9: if_acmpeq 327
      // 2fc: goto 300
      // 2ff: athrow
      // 300: new java/lang/StringBuilder
      // 303: dup
      // 304: invokespecial java/lang/StringBuilder.<init> ()V
      // 307: aload 6
      // 309: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 30c: getstatic k.ob [Ljava/lang/String;
      // 30f: bipush 29
      // 311: aaload
      // 312: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 315: invokevirtual java/lang/StringBuilder.toString ()Ljava/lang/String;
      // 318: bipush 0
      // 319: aload 0
      // 31a: getfield k.m [B
      // 31d: iload 4
      // 31f: bipush -125
      // 321: ixor
      // 322: invokestatic na.a (Ljava/lang/String;I[BI)[B
      // 325: astore 7
      // 327: aload 7
      // 329: ifnull 339
      // 32c: bipush -1
      // 32d: goto 331
      // 330: athrow
      // 331: aload 7
      // 333: arraylength
      // 334: bipush -1
      // 335: ixor
      // 336: if_icmpne 342
      // 339: new java/io/IOException
      // 33c: dup
      // 33d: invokespecial java/io/IOException.<init> ()V
      // 340: athrow
      // 341: athrow
      // 342: bipush 0
      // 343: istore 8
      // 345: bipush 0
      // 346: istore 9
      // 348: sipush 2304
      // 34b: iload 9
      // 34d: if_icmple 372
      // 350: aload 0
      // 351: getfield k.f [[B
      // 354: iload 2
      // 355: aaload
      // 356: iload 9
      // 358: aload 7
      // 35a: iload 8
      // 35c: iinc 8 1
      // 35f: baload
      // 360: bastore
      // 361: iinc 9 1
      // 364: iload 13
      // 366: ifne 375
      // 369: iload 13
      // 36b: ifeq 348
      // 36e: goto 372
      // 371: athrow
      // 372: bipush 0
      // 373: istore 9
      // 375: iload 9
      // 377: sipush 2304
      // 37a: if_icmpge 3a3
      // 37d: aload 0
      // 37e: getfield k.P [[B
      // 381: iload 2
      // 382: aaload
      // 383: iload 9
      // 385: aload 7
      // 387: iload 8
      // 389: iinc 8 1
      // 38c: baload
      // 38d: bastore
      // 38e: iinc 9 1
      // 391: iload 13
      // 393: ifne 3a6
      // 396: goto 39a
      // 399: athrow
      // 39a: iload 13
      // 39c: ifeq 375
      // 39f: goto 3a3
      // 3a2: athrow
      // 3a3: bipush 0
      // 3a4: istore 9
      // 3a6: iload 9
      // 3a8: bipush -1
      // 3a9: ixor
      // 3aa: sipush -2305
      // 3ad: if_icmple 3dc
      // 3b0: aload 0
      // 3b1: getfield k.s [[I
      // 3b4: iload 2
      // 3b5: aaload
      // 3b6: iload 9
      // 3b8: sipush 255
      // 3bb: aload 7
      // 3bd: iload 8
      // 3bf: iinc 8 1
      // 3c2: baload
      // 3c3: invokestatic ib.a (II)I
      // 3c6: iastore
      // 3c7: iinc 9 1
      // 3ca: iload 13
      // 3cc: ifne 3df
      // 3cf: goto 3d3
      // 3d2: athrow
      // 3d3: iload 13
      // 3d5: ifeq 3a6
      // 3d8: goto 3dc
      // 3db: athrow
      // 3dc: bipush 0
      // 3dd: istore 9
      // 3df: iload 9
      // 3e1: bipush -1
      // 3e2: ixor
      // 3e3: sipush -2305
      // 3e6: if_icmple 422
      // 3e9: sipush 255
      // 3ec: aload 7
      // 3ee: iload 8
      // 3f0: iinc 8 1
      // 3f3: baload
      // 3f4: iand
      // 3f5: istore 10
      // 3f7: iload 10
      // 3f9: bipush -1
      // 3fa: ixor
      // 3fb: bipush -1
      // 3fc: iload 13
      // 3fe: ifne 42a
      // 401: if_icmplt 40b
      // 404: goto 408
      // 407: athrow
      // 408: goto 41a
      // 40b: aload 0
      // 40c: getfield k.s [[I
      // 40f: iload 2
      // 410: aaload
      // 411: iload 9
      // 413: iload 10
      // 415: sipush 12000
      // 418: iadd
      // 419: iastore
      // 41a: iinc 9 1
      // 41d: iload 13
      // 41f: ifeq 3df
      // 422: bipush 0
      // 423: istore 9
      // 425: sipush 2304
      // 428: iload 9
      // 42a: if_icmple 494
      // 42d: aload 7
      // 42f: iload 8
      // 431: iinc 8 1
      // 434: baload
      // 435: sipush 255
      // 438: iand
      // 439: istore 10
      // 43b: sipush 128
      // 43e: iload 10
      // 440: iload 13
      // 442: ifne 49f
      // 445: if_icmple 464
      // 448: goto 44c
      // 44b: athrow
      // 44c: aload 0
      // 44d: getfield k.A [[B
      // 450: iload 2
      // 451: aaload
      // 452: iload 9
      // 454: iinc 9 1
      // 457: iload 10
      // 459: i2b
      // 45a: bastore
      // 45b: iload 13
      // 45d: ifeq 48f
      // 460: goto 464
      // 463: athrow
      // 464: bipush 0
      // 465: istore 11
      // 467: bipush -128
      // 469: iload 10
      // 46b: iadd
      // 46c: bipush -1
      // 46d: ixor
      // 46e: iload 11
      // 470: bipush -1
      // 471: ixor
      // 472: if_icmpge 48f
      // 475: aload 0
      // 476: getfield k.A [[B
      // 479: iload 2
      // 47a: aaload
      // 47b: iload 9
      // 47d: iinc 9 1
      // 480: bipush 0
      // 481: bastore
      // 482: iinc 11 1
      // 485: iload 13
      // 487: ifne 425
      // 48a: iload 13
      // 48c: ifeq 467
      // 48f: iload 13
      // 491: ifeq 425
      // 494: bipush 0
      // 495: istore 9
      // 497: bipush 0
      // 498: istore 10
      // 49a: iload 10
      // 49c: sipush 2304
      // 49f: if_icmpge 509
      // 4a2: aload 7
      // 4a4: iload 8
      // 4a6: iinc 8 1
      // 4a9: baload
      // 4aa: sipush 255
      // 4ad: iand
      // 4ae: istore 11
      // 4b0: iload 11
      // 4b2: bipush -1
      // 4b3: ixor
      // 4b4: sipush -129
      // 4b7: iload 13
      // 4b9: ifne 511
      // 4bc: if_icmple 4db
      // 4bf: goto 4c3
      // 4c2: athrow
      // 4c3: aload 0
      // 4c4: getfield k.R [[B
      // 4c7: iload 2
      // 4c8: aaload
      // 4c9: iload 10
      // 4cb: iinc 10 1
      // 4ce: iload 11
      // 4d0: i2b
      // 4d1: bastore
      // 4d2: iload 11
      // 4d4: istore 9
      // 4d6: iload 13
      // 4d8: ifeq 504
      // 4db: bipush 0
      // 4dc: istore 12
      // 4de: bipush -128
      // 4e0: iload 11
      // 4e2: iadd
      // 4e3: iload 12
      // 4e5: if_icmple 504
      // 4e8: aload 0
      // 4e9: getfield k.R [[B
      // 4ec: iload 2
      // 4ed: aaload
      // 4ee: iload 10
      // 4f0: iinc 10 1
      // 4f3: iload 9
      // 4f5: i2b
      // 4f6: bastore
      // 4f7: iinc 12 1
      // 4fa: iload 13
      // 4fc: ifne 49a
      // 4ff: iload 13
      // 501: ifeq 4de
      // 504: iload 13
      // 506: ifeq 49a
      // 509: bipush 0
      // 50a: istore 10
      // 50c: sipush 2304
      // 50f: iload 10
      // 511: if_icmple 579
      // 514: sipush 255
      // 517: aload 7
      // 519: iload 8
      // 51b: iinc 8 1
      // 51e: baload
      // 51f: iand
      // 520: istore 11
      // 522: iload 11
      // 524: bipush -1
      // 525: ixor
      // 526: sipush -129
      // 529: iload 13
      // 52b: ifne 5aa
      // 52e: if_icmple 54d
      // 531: goto 535
      // 534: athrow
      // 535: aload 0
      // 536: getfield k.mb [[B
      // 539: iload 2
      // 53a: aaload
      // 53b: iload 10
      // 53d: iinc 10 1
      // 540: iload 11
      // 542: i2b
      // 543: bastore
      // 544: iload 13
      // 546: ifeq 574
      // 549: goto 54d
      // 54c: athrow
      // 54d: bipush 0
      // 54e: istore 12
      // 550: iload 12
      // 552: bipush -128
      // 554: iload 11
      // 556: iadd
      // 557: if_icmpge 574
      // 55a: aload 0
      // 55b: getfield k.mb [[B
      // 55e: iload 2
      // 55f: aaload
      // 560: iload 10
      // 562: iinc 10 1
      // 565: bipush 0
      // 566: bastore
      // 567: iinc 12 1
      // 56a: iload 13
      // 56c: ifne 50c
      // 56f: iload 13
      // 571: ifeq 550
      // 574: iload 13
      // 576: ifeq 50c
      // 579: new java/lang/StringBuilder
      // 57c: dup
      // 57d: invokespecial java/lang/StringBuilder.<init> ()V
      // 580: aload 6
      // 582: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 585: getstatic k.ob [Ljava/lang/String;
      // 588: bipush 30
      // 58a: aaload
      // 58b: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 58e: invokevirtual java/lang/StringBuilder.toString ()Ljava/lang/String;
      // 591: bipush 0
      // 592: aload 0
      // 593: getfield k.gb [B
      // 596: bipush -127
      // 598: invokestatic na.a (Ljava/lang/String;I[BI)[B
      // 59b: astore 7
      // 59d: aload 7
      // 59f: ifnull 60b
      // 5a2: bipush 0
      // 5a3: aload 7
      // 5a5: arraylength
      // 5a6: goto 5aa
      // 5a9: athrow
      // 5aa: if_icmplt 5b0
      // 5ad: goto 60b
      // 5b0: bipush 0
      // 5b1: istore 8
      // 5b3: bipush 0
      // 5b4: istore 10
      // 5b6: iload 10
      // 5b8: bipush -1
      // 5b9: ixor
      // 5ba: sipush -2305
      // 5bd: if_icmple 60b
      // 5c0: aload 7
      // 5c2: iload 8
      // 5c4: iinc 8 1
      // 5c7: baload
      // 5c8: sipush 255
      // 5cb: iand
      // 5cc: istore 11
      // 5ce: iload 13
      // 5d0: ifne 804
      // 5d3: sipush -129
      // 5d6: iload 11
      // 5d8: bipush -1
      // 5d9: ixor
      // 5da: if_icmpge 5fc
      // 5dd: goto 5e1
      // 5e0: athrow
      // 5e1: aload 0
      // 5e2: getfield k.s [[I
      // 5e5: iload 2
      // 5e6: aaload
      // 5e7: iload 10
      // 5e9: iinc 10 1
      // 5ec: ldc_w 48000
      // 5ef: iload 11
      // 5f1: iadd
      // 5f2: iastore
      // 5f3: iload 13
      // 5f5: ifeq 606
      // 5f8: goto 5fc
      // 5fb: athrow
      // 5fc: iload 10
      // 5fe: bipush -128
      // 600: iload 11
      // 602: iadd
      // 603: iadd
      // 604: istore 10
      // 606: iload 13
      // 608: ifeq 5b6
      // 60b: iload 13
      // 60d: ifeq 804
      // 610: sipush 20736
      // 613: newarray 8
      // 615: astore 7
      // 617: new java/lang/StringBuilder
      // 61a: dup
      // 61b: invokespecial java/lang/StringBuilder.<init> ()V
      // 61e: getstatic k.ob [Ljava/lang/String;
      // 621: bipush 28
      // 623: aaload
      // 624: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 627: aload 6
      // 629: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 62c: getstatic k.ob [Ljava/lang/String;
      // 62f: bipush 31
      // 631: aaload
      // 632: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 635: invokevirtual java/lang/StringBuilder.toString ()Ljava/lang/String;
      // 638: iload 4
      // 63a: sipush -19675
      // 63d: ixor
      // 63e: aload 7
      // 640: sipush 20736
      // 643: invokestatic ta.a (Ljava/lang/String;I[BI)V
      // 646: bipush 0
      // 647: istore 8
      // 649: bipush 0
      // 64a: istore 9
      // 64c: bipush 0
      // 64d: istore 10
      // 64f: sipush 2304
      // 652: iload 10
      // 654: if_icmple 686
      // 657: sipush 255
      // 65a: iload 8
      // 65c: aload 7
      // 65e: iload 9
      // 660: iinc 9 1
      // 663: baload
      // 664: ineg
      // 665: isub
      // 666: iand
      // 667: istore 8
      // 669: aload 0
      // 66a: getfield k.L [[B
      // 66d: iload 2
      // 66e: aaload
      // 66f: iload 10
      // 671: iload 8
      // 673: i2b
      // 674: bastore
      // 675: iinc 10 1
      // 678: iload 13
      // 67a: ifne 689
      // 67d: iload 13
      // 67f: ifeq 64f
      // 682: goto 686
      // 685: athrow
      // 686: bipush 0
      // 687: istore 8
      // 689: bipush 0
      // 68a: istore 10
      // 68c: sipush -2305
      // 68f: iload 10
      // 691: bipush -1
      // 692: ixor
      // 693: if_icmpge 6c4
      // 696: sipush 255
      // 699: iload 8
      // 69b: aload 7
      // 69d: iload 9
      // 69f: iinc 9 1
      // 6a2: baload
      // 6a3: iadd
      // 6a4: iand
      // 6a5: istore 8
      // 6a7: aload 0
      // 6a8: getfield k.eb [[B
      // 6ab: iload 2
      // 6ac: aaload
      // 6ad: iload 10
      // 6af: iload 8
      // 6b1: i2b
      // 6b2: bastore
      // 6b3: iinc 10 1
      // 6b6: iload 13
      // 6b8: ifne 6c7
      // 6bb: iload 13
      // 6bd: ifeq 68c
      // 6c0: goto 6c4
      // 6c3: athrow
      // 6c4: bipush 0
      // 6c5: istore 10
      // 6c7: iload 10
      // 6c9: sipush 2304
      // 6cc: if_icmpge 6f5
      // 6cf: aload 0
      // 6d0: getfield k.f [[B
      // 6d3: iload 2
      // 6d4: aaload
      // 6d5: iload 10
      // 6d7: aload 7
      // 6d9: iload 9
      // 6db: iinc 9 1
      // 6de: baload
      // 6df: bastore
      // 6e0: iinc 10 1
      // 6e3: iload 13
      // 6e5: ifne 6f8
      // 6e8: goto 6ec
      // 6eb: athrow
      // 6ec: iload 13
      // 6ee: ifeq 6c7
      // 6f1: goto 6f5
      // 6f4: athrow
      // 6f5: bipush 0
      // 6f6: istore 10
      // 6f8: iload 10
      // 6fa: sipush 2304
      // 6fd: if_icmpge 726
      // 700: aload 0
      // 701: getfield k.P [[B
      // 704: iload 2
      // 705: aaload
      // 706: iload 10
      // 708: aload 7
      // 70a: iload 9
      // 70c: iinc 9 1
      // 70f: baload
      // 710: bastore
      // 711: iinc 10 1
      // 714: iload 13
      // 716: ifne 729
      // 719: goto 71d
      // 71c: athrow
      // 71d: iload 13
      // 71f: ifeq 6f8
      // 722: goto 726
      // 725: athrow
      // 726: bipush 0
      // 727: istore 10
      // 729: sipush 2304
      // 72c: iload 10
      // 72e: if_icmple 76f
      // 731: aload 0
      // 732: getfield k.s [[I
      // 735: iload 2
      // 736: aaload
      // 737: iload 10
      // 739: sipush 255
      // 73c: aload 7
      // 73e: bipush 1
      // 73f: iload 9
      // 741: iadd
      // 742: baload
      // 743: invokestatic ib.a (II)I
      // 746: aload 7
      // 748: iload 9
      // 74a: baload
      // 74b: sipush 255
      // 74e: invokestatic ib.a (II)I
      // 751: sipush 256
      // 754: imul
      // 755: iadd
      // 756: iastore
      // 757: iinc 9 2
      // 75a: iinc 10 1
      // 75d: iload 13
      // 75f: ifne 772
      // 762: goto 766
      // 765: athrow
      // 766: iload 13
      // 768: ifeq 729
      // 76b: goto 76f
      // 76e: athrow
      // 76f: bipush 0
      // 770: istore 10
      // 772: sipush 2304
      // 775: iload 10
      // 777: if_icmple 7a0
      // 77a: aload 0
      // 77b: getfield k.A [[B
      // 77e: iload 2
      // 77f: aaload
      // 780: iload 10
      // 782: aload 7
      // 784: iload 9
      // 786: iinc 9 1
      // 789: baload
      // 78a: bastore
      // 78b: iinc 10 1
      // 78e: iload 13
      // 790: ifne 7a3
      // 793: goto 797
      // 796: athrow
      // 797: iload 13
      // 799: ifeq 772
      // 79c: goto 7a0
      // 79f: athrow
      // 7a0: bipush 0
      // 7a1: istore 10
      // 7a3: iload 10
      // 7a5: bipush -1
      // 7a6: ixor
      // 7a7: sipush -2305
      // 7aa: if_icmple 7d3
      // 7ad: aload 0
      // 7ae: getfield k.R [[B
      // 7b1: iload 2
      // 7b2: aaload
      // 7b3: iload 10
      // 7b5: aload 7
      // 7b7: iload 9
      // 7b9: iinc 9 1
      // 7bc: baload
      // 7bd: bastore
      // 7be: iinc 10 1
      // 7c1: iload 13
      // 7c3: ifne 7d6
      // 7c6: goto 7ca
      // 7c9: athrow
      // 7ca: iload 13
      // 7cc: ifeq 7a3
      // 7cf: goto 7d3
      // 7d2: athrow
      // 7d3: bipush 0
      // 7d4: istore 10
      // 7d6: sipush 2304
      // 7d9: iload 10
      // 7db: if_icmple 804
      // 7de: aload 0
      // 7df: getfield k.mb [[B
      // 7e2: iload 2
      // 7e3: aaload
      // 7e4: iload 10
      // 7e6: aload 7
      // 7e8: iload 9
      // 7ea: iinc 9 1
      // 7ed: baload
      // 7ee: bastore
      // 7ef: iinc 10 1
      // 7f2: iload 13
      // 7f4: ifne 8a1
      // 7f7: goto 7fb
      // 7fa: athrow
      // 7fb: iload 13
      // 7fd: ifeq 7d6
      // 800: goto 804
      // 803: athrow
      // 804: goto 8a1
      // 807: astore 7
      // 809: bipush 0
      // 80a: istore 8
      // 80c: iload 8
      // 80e: bipush -1
      // 80f: ixor
      // 810: sipush -2305
      // 813: if_icmple 8a1
      // 816: aload 0
      // 817: getfield k.L [[B
      // 81a: iload 2
      // 81b: aaload
      // 81c: iload 8
      // 81e: bipush 0
      // 81f: bastore
      // 820: aload 0
      // 821: getfield k.eb [[B
      // 824: iload 2
      // 825: aaload
      // 826: iload 8
      // 828: bipush 0
      // 829: bastore
      // 82a: aload 0
      // 82b: getfield k.f [[B
      // 82e: iload 2
      // 82f: aaload
      // 830: iload 8
      // 832: bipush 0
      // 833: bastore
      // 834: aload 0
      // 835: getfield k.P [[B
      // 838: iload 2
      // 839: aaload
      // 83a: iload 8
      // 83c: bipush 0
      // 83d: bastore
      // 83e: aload 0
      // 83f: getfield k.s [[I
      // 842: iload 2
      // 843: aaload
      // 844: iload 8
      // 846: bipush 0
      // 847: iastore
      // 848: aload 0
      // 849: getfield k.A [[B
      // 84c: iload 2
      // 84d: aaload
      // 84e: iload 8
      // 850: bipush 0
      // 851: bastore
      // 852: aload 0
      // 853: getfield k.R [[B
      // 856: iload 2
      // 857: aaload
      // 858: iload 8
      // 85a: bipush 0
      // 85b: bastore
      // 85c: iload 13
      // 85e: ifne 8ee
      // 861: iload 1
      // 862: bipush -1
      // 863: ixor
      // 864: bipush -1
      // 865: if_icmpne 87b
      // 868: goto 86c
      // 86b: athrow
      // 86c: aload 0
      // 86d: getfield k.R [[B
      // 870: iload 2
      // 871: aaload
      // 872: iload 8
      // 874: bipush -6
      // 876: bastore
      // 877: goto 87b
      // 87a: athrow
      // 87b: iload 1
      // 87c: bipush 3
      // 87d: if_icmpeq 884
      // 880: goto 88f
      // 883: athrow
      // 884: aload 0
      // 885: getfield k.R [[B
      // 888: iload 2
      // 889: aaload
      // 88a: iload 8
      // 88c: bipush 8
      // 88e: bastore
      // 88f: aload 0
      // 890: getfield k.mb [[B
      // 893: iload 2
      // 894: aaload
      // 895: iload 8
      // 897: bipush 0
      // 898: bastore
      // 899: iinc 8 1
      // 89c: iload 13
      // 89e: ifeq 80c
      // 8a1: goto 8ee
      // 8a4: astore 6
      // 8a6: aload 6
      // 8a8: new java/lang/StringBuilder
      // 8ab: dup
      // 8ac: invokespecial java/lang/StringBuilder.<init> ()V
      // 8af: getstatic k.ob [Ljava/lang/String;
      // 8b2: bipush 27
      // 8b4: aaload
      // 8b5: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 8b8: iload 1
      // 8b9: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 8bc: bipush 44
      // 8be: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 8c1: iload 2
      // 8c2: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 8c5: bipush 44
      // 8c7: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 8ca: iload 3
      // 8cb: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 8ce: bipush 44
      // 8d0: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 8d3: iload 4
      // 8d5: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 8d8: bipush 44
      // 8da: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 8dd: iload 5
      // 8df: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 8e2: bipush 41
      // 8e4: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 8e7: invokevirtual java/lang/StringBuilder.toString ()Ljava/lang/String;
      // 8ea: invokestatic i.a (Ljava/lang/Throwable;Ljava/lang/String;)Lla;
      // 8ed: athrow
      // 8ee: return
      // try (34 -> 1036): 1037 java/io/IOException
      // try (2 -> 4): 1135 java/lang/RuntimeException
      // try (5 -> 1134): 1135 java/lang/RuntimeException
   }

   final int b(int var1, int var2, int var3) {
      boolean var5 = client.vh;

      try {
         ib++;
         if (0 > var1 || ~var1 <= -97 || -1 < ~var2 || -97 >= ~var2) {
            return 0;
         }

         if (var3 > -68) {
            return -68;
         }

         byte var4 = 0;
         if (~var1 > -49 || var2 >= 48) {
            if (-49 >= ~var1 || ~var2 > -49) {
               if (48 > var1 || var2 < 48) {
                  return this.mb[var4][var2 + var1 * 48];
               }

               var1 -= 48;
               var2 -= 48;
               var4 = 3;
               if (!var5) {
                  return this.mb[var4][var2 + var1 * 48];
               }
            }

            var2 -= 48;
            var4 = 2;
            if (!var5) {
               return this.mb[var4][var2 + var1 * 48];
            }
         }

         var1 -= 48;
         var4 = 1;
         return this.mb[var4][var2 + var1 * 48];
      } catch (RuntimeException var6) {
         throw i.a(var6, ob[11] + var1 + 44 + var2 + 44 + var3 + 41);
      }
   }

   private final void a(int var1, byte var2, int var3, int var4, int var5, int var6) {
      try {
         Y++;
         int var7 = var4 * 3;
         int var8 = var5 * 3;
         int var9 = this.c.a(var6, true);
         int var10 = -57 % ((var2 - -64) / 57);
         var9 = var9 >> 1 & 8355711;
         int var11 = this.c.a(var3, true);
         var11 = (16711423 & var11) >> 1;
         if (var1 == 0) {
            this.U.b(3, var9, var7, var8, (byte)109);
            this.U.b(2, var9, var7, 1 + var8, (byte)-65);
            this.U.b(1, var9, var7, var8 - -2, (byte)99);
            this.U.b(1, var11, 2 + var7, var8 - -1, (byte)73);
            this.U.b(2, var11, var7 + 1, var8 + 2, (byte)113);
            if (!client.vh) {
               return;
            }
         }

         if (1 == var1) {
            this.U.b(3, var11, var7, var8, (byte)55);
            this.U.b(2, var11, 1 + var7, 1 + var8, (byte)62);
            this.U.b(1, var11, var7 + 2, var8 - -2, (byte)56);
            this.U.b(1, var9, var7, var8 + 1, (byte)70);
            this.U.b(2, var9, var7, 2 + var8, (byte)-85);
         }
      } catch (RuntimeException var12) {
         throw i.a(var12, ob[36] + var1 + ',' + var2 + ',' + var3 + ',' + var4 + ',' + var5 + ',' + var6 + ')');
      }
   }

   private final int b(byte var1, int var2, int var3) {
      try {
         if (var1 != -38) {
            this.B = (int[][])null;
         }

         d++;
         return 0 <= var3 && ~var2 <= -1 && 96 > var3 && ~var2 > -97 ? this.bb[var3][var2] : 0;
      } catch (RuntimeException var5) {
         throw i.a(var5, ob[33] + var1 + 44 + var2 + 44 + var3 + 41);
      }
   }

   private final int b(int var1, int var2, int var3, int var4) {
      boolean var6 = client.vh;

      try {
         K++;
         if (var2 >= 0 && ~var2 > -97 && var4 >= 0 && -97 < ~var4) {
            byte var5 = 0;
            if (-49 < ~var2 || var4 >= 48) {
               if (48 > var2 && 48 <= var4) {
                  var4 -= 48;
                  var5 = 2;
                  if (!var6) {
                     return var3 != 4 ? -4 : 0xFF & this.R[var5][48 * var2 + var4];
                  }
               }

               if (-49 < ~var2 || var4 < 48) {
                  return var3 != 4 ? -4 : 0xFF & this.R[var5][48 * var2 + var4];
               }

               var4 -= 48;
               var2 -= 48;
               var5 = 3;
               if (!var6) {
                  return var3 != 4 ? -4 : 0xFF & this.R[var5][48 * var2 + var4];
               }
            }

            var2 -= 48;
            var5 = 1;
            return var3 != 4 ? -4 : 0xFF & this.R[var5][48 * var2 + var4];
         } else {
            return 0;
         }
      } catch (RuntimeException var7) {
         throw i.a(var7, ob[23] + var1 + 44 + var2 + 44 + var3 + 44 + var4 + 41);
      }
   }

   private final int a(int var1, byte var2, int var3) {
      boolean var6 = client.vh;

      try {
         M++;
         if (0 <= var1 && var1 < 96 && var3 >= 0 && var3 < 96) {
            int var4 = -67 / ((var2 - -48) / 60);
            byte var5 = 0;
            if (-49 >= ~var1 && 48 > var3) {
               var1 -= 48;
               var5 = 1;
               if (!var6) {
                  return 0xFF & this.P[var5][var1 * 48 - -var3];
               }
            }

            if (-49 < ~var1 && ~var3 <= -49) {
               var3 -= 48;
               var5 = 2;
               if (!var6) {
                  return 0xFF & this.P[var5][var1 * 48 - -var3];
               }
            }

            if (~var1 <= -49 && -49 >= ~var3) {
               var3 -= 48;
               var1 -= 48;
               var5 = 3;
            }

            return 0xFF & this.P[var5][var1 * 48 - -var3];
         } else {
            return 0;
         }
      } catch (RuntimeException var7) {
         throw i.a(var7, ob[13] + var1 + 44 + var2 + 44 + var3 + 41);
      }
   }

   private final void a(int var1, ca var2, int var3, int var4, int var5, int var6, int var7) {
      try {
         C++;
         this.b(40, (byte)50, var5, var4);
         this.b(40, (byte)109, var3, var7);
         int var8 = ib.d[var1];
         int var9 = v.a[var1];
         if (var6 != -14584) {
            this.a((byte)-62, 104, -113);
         }

         int var10 = client.Jk[var1];
         int var11 = 128 * var5;
         int var12 = 128 * var4;
         int var13 = 128 * var3;
         int var14 = var7 * 128;
         int var15 = var2.e(var11, var12, -this.ab[var5][var4], -111);
         int var16 = var2.e(var11, var12, -this.ab[var5][var4] + -var8, -115);
         int var17 = var2.e(var13, var14, -var8 + -this.ab[var3][var7], -125);
         int var18 = var2.e(var13, var14, -this.ab[var3][var7], var6 ^ -14505);
         int[] var19 = new int[]{var15, var16, var17, var18};
         int var20 = var2.a(4, var19, var9, var10, false);
         if (-6 == ~lb.Tb[var1]) {
            var2.E[var20] = 30000 + var1;
            if (!client.vh) {
               return;
            }
         }

         var2.E[var20] = 0;
      } catch (RuntimeException var21) {
         throw i.a(var21, ob[15] + var1 + ',' + (var2 != null ? ob[1] : ob[2]) + ',' + var3 + ',' + var4 + ',' + var5 + ',' + var6 + ',' + var7 + ')');
      }
   }

   final void a(int var1, int var2, int var3, int var4, int var5) {
      boolean var7 = client.vh;

      try {
         W++;
         if (-1 >= ~var4 && ~var1 <= -1 && -96 < ~var4 && var1 < 95) {
            if (var5 != 11715) {
               this.s = (int[][])null;
            }

            if (u.a[var2] == 1) {
               label87: {
                  if (-1 == ~var3) {
                     this.bb[var4][var1] = d.a(this.bb[var4][var1], 1);
                     if (0 >= var1) {
                        break label87;
                     }

                     this.a(4, -1 + var1, (byte)-96, var4);
                     if (!var7) {
                        break label87;
                     }
                  }

                  if (1 == var3) {
                     this.bb[var4][var1] = d.a(this.bb[var4][var1], 2);
                     if (0 >= var4) {
                        break label87;
                     }

                     this.a(8, var1, (byte)-89, -1 + var4);
                     if (!var7) {
                        break label87;
                     }
                  }

                  if (-3 != ~var3) {
                     if (~var3 != -4) {
                        break label87;
                     }

                     this.bb[var4][var1] = d.a(this.bb[var4][var1], 32);
                     if (!var7) {
                        break label87;
                     }
                  }

                  this.bb[var4][var1] = d.a(this.bb[var4][var1], 16);
               }

               this.c(1, 1, 62, var4, var1);
            }
         }
      } catch (RuntimeException var8) {
         throw i.a(var8, ob[25] + var1 + ',' + var2 + ',' + var3 + ',' + var4 + ',' + var5 + ')');
      }
   }

   private final boolean a(boolean var1, int var2, int var3) {
      try {
         V++;
         return ~this.d(var3, var2, 114) < -1
               && this.d(var3, var2 - 1, 122) > 0
               && -1 > ~this.d(-1 + var3, -1 + var2, 117)
               && ~this.d(-1 + var3, var2, 122) < -1
            ? true
            : var1;
      } catch (RuntimeException var5) {
         throw i.a(var5, ob[14] + var1 + ',' + var2 + ',' + var3 + ')');
      }
   }

   private final int e(int var1, int var2, int var3) {
      boolean var6 = client.vh;

      try {
         b++;
         int var4 = -120 / ((var1 - 15) / 43);
         if (var2 >= 0 && ~var2 > -97 && -1 >= ~var3 && var3 < 96) {
            byte var5 = 0;
            if (var2 < 48 || -49 >= ~var3) {
               if (-49 >= ~var2 || ~var3 > -49) {
                  if (var2 < 48 || 48 > var3) {
                     return 0xFF & this.f[var5][var2 * 48 + var3];
                  }

                  var3 -= 48;
                  var2 -= 48;
                  var5 = 3;
                  if (!var6) {
                     return 0xFF & this.f[var5][var2 * 48 + var3];
                  }
               }

               var5 = 2;
               var3 -= 48;
               if (!var6) {
                  return 0xFF & this.f[var5][var2 * 48 + var3];
               }
            }

            var2 -= 48;
            var5 = 1;
            return 0xFF & this.f[var5][var2 * 48 + var3];
         } else {
            return 0;
         }
      } catch (RuntimeException var7) {
         throw i.a(var7, ob[21] + var1 + 44 + var2 + 44 + var3 + 41);
      }
   }

   k(lb var1, ua var2) {
      boolean var4 = client.vh;
      super();
      this.H = false;
      this.ab = new int[96][96];
      this.mb = new byte[4][2304];
      this.P = new byte[4][2304];
      this.R = new byte[4][2304];
      this.w = new int[256];
      this.B = new int[96][96];
      this.f = new byte[4][2304];
      this.E = new int[18432];
      this.nb = true;
      this.x = 750;
      this.g = new ca[4][64];
      this.L = new byte[4][2304];
      this.F = new ca[64];
      this.s = new int[4][2304];
      this.A = new byte[4][2304];
      this.Z = false;
      this.bb = new int[96][96];
      this.db = new ca[4][64];
      this.eb = new byte[4][2304];
      this.q = new int[18432];

      try {
         this.U = var2;
         this.c = var1;
         int var3 = 0;

         while (true) {
            if (var3 < 64) {
               this.w[var3] = da.a(255 + -(var3 * 4), (byte)-66, -(4 * var3) + 255, 255 - (int)(var3 * 1.75));
               var3++;
               if (var4) {
                  break;
               }

               if (!var4) {
                  continue;
               }
            }

            var3 = 0;
            break;
         }

         while (true) {
            if (~var3 > -65) {
               this.w[64 + var3] = da.a(0, (byte)-66, 3 * var3, 144);
               var3++;
               if (var4) {
                  break;
               }

               if (!var4) {
                  continue;
               }
            }

            var3 = 0;
            break;
         }

         while (true) {
            if (var3 < 64) {
               this.w[128 + var3] = da.a(0, (byte)-66, 192 - (int)(var3 * 1.5), -((int)(var3 * 1.5)) + 144);
               var3++;
               if (var4) {
                  break;
               }

               if (!var4) {
                  continue;
               }
            }

            var3 = 0;
            break;
         }

         while (64 > var3) {
            this.w[192 + var3] = da.a(0, (byte)-66, 96 - (int)(var3 * 1.5), (int)(var3 * 1.5) + 48);
            var3++;
            if (var4 || var4) {
               break;
            }
         }
      } catch (RuntimeException var5) {
         throw i.a(var5, ob[3] + (var1 != null ? ob[1] : ob[2]) + ',' + (var2 != null ? ob[1] : ob[2]) + ')');
      }
   }

   private static char[] z(String var0) {
      char[] var10000 = var0.toCharArray();
      if (var10000.length < 2) {
         var10000[0] = (char)(var10000[0] ^ 27);
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
               var10005 = 117;
               break;
            case 1:
               var10005 = 34;
               break;
            case 2:
               var10005 = 84;
               break;
            case 3:
               var10005 = 51;
               break;
            default:
               var10005 = 27;
         }

         var10001[var1] = (char)(var10004 ^ var10005);
      }

      return new String(var10001).intern();
   }
}
