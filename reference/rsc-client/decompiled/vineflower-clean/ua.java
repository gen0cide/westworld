import java.awt.Component;
import java.awt.Graphics;
import java.awt.Image;
import java.awt.image.ColorModel;
import java.awt.image.DirectColorModel;
import java.awt.image.ImageConsumer;
import java.awt.image.ImageObserver;
import java.awt.image.ImageProducer;

class ua implements ImageProducer, ImageObserver {
   static int o;
   private int[] G;
   static int W;
   private int[] tb;
   private int[] Sb;
   static int S;
   boolean xb;
   int[] qb;
   static int d;
   static int db;
   static int sb;
   static int K;
   static int j;
   static int zb;
   static int N;
   static int e;
   static int Q;
   boolean i;
   static int Z;
   static int q;
   static int m;
   private int[] Xb;
   static int Cb;
   static int L;
   static v E = new v(z(z("sysC")), "", "", 0);
   private int[] t;
   int[] Eb;
   private int[] Tb;
   private int[] Wb;
   int[] rb;
   static int r;
   static int bb;
   static int n;
   static int b;
   private ColorModel nb;
   static int ib;
   int[] R;
   static int Vb;
   byte[][] gb;
   static int a;
   static int y;
   static int ab;
   int[] kb;
   static int U;
   static int pb;
   int u;
   static int v;
   static int yb;
   int[][] ob;
   private int A;
   static int V;
   static int w;
   private int hb;
   private int D;
   static int ub;
   static int s;
   int[][] Y;
   static int Ob;
   static int f;
   static int Jb;
   static int Pb;
   static int Ub;
   static int mb;
   static int Ib;
   static int O;
   static int H;
   int k;
   static int J;
   static int x;
   private boolean[] Qb;
   static String[] wb = new String[100];
   static int T;
   static int jb;
   static int B;
   private int lb;
   private ImageConsumer fb;
   static int c;
   static int Nb;
   static int[] Bb;
   static int X;
   static int F;
   static int z;
   static int Db;
   static int C = 114;
   private Image Gb;
   static int vb;
   static int l;
   static int Fb;
   static int eb;
   static int p;
   static int cb;
   static int Lb;
   private int[] M;
   static int I;
   static int[] Ab;
   private int Rb;
   static String[] Kb = new String[]{z(z("z^Qc{\u001f^PkkZB\u0005io\u001fYQcdL\u0010Qi)MUHi\u007fZ\u0010Dhm\u001f@WczL\u0010@h}ZB"))};
   static int P;
   private int[] Hb;
   static int g;
   static String[] h = new String[200];
   static int[] Mb;
   private static final String[] Yb = new String[]{
      z(z("QEIj")),
      z(z("D\u001e\u000b(t")),
      z(z("JQ\u000boz|_Ku|RUW.")),
      z(z("JQ\u000bJH\u0017")),
      z(z("JQ\u000bOH\u0017")),
      z(z("JQ\u000bIH\u0017")),
      z(z("JQ\u000bHH\u0017")),
      z(z("JQ\u000bS!")),
      z(z("JQ\u000bLK\u0017")),
      z(z("JQ\u000bA!")),
      z(z("JQ\u000b@!")),
      z(z("JQ\u000bM!")),
      z(z("JQ\u000bu}^BQV{PTPe}V_K.")),
      z(z("JQ\u000bVH\u0017")),
      z(z("O\\JreZDQc{\u0005\u0010")),
      z(z("JQ\u000bB!")),
      z(z("ZBWi{\u001fYK&zOBLrl\u001fSIoyOYKa)M_Pr`QU")),
      z(z("JQ\u000bW!")),
      z(z("JQ\u000bDK\u0017")),
      z(z("JQ\u000bR!")),
      z(z("JQ\u000bPH\u0017")),
      z(z("JQ\u000bEH\u0017")),
      z(z("PB\u0016")),
      z(z("XB@")),
      z(z("PB\u0014")),
      z(z("RQB")),
      z(z("JQ\u000b@K\u0017")),
      z(z("\\ID")),
      z(z("SB@")),
      z(z("XB\u0014")),
      z(z("[BDqzKBLhn\u0005\u0010")),
      z(z("[B@")),
      z(z("]\\P")),
      z(z("HXL")),
      z(z("XB\u0017")),
      z(z("]\\D")),
      z(z("XB\u0016")),
      z(z("FUI")),
      z(z("MUA")),
      z(z("MQK")),
      z(z("PB\u0017")),
      z(z("PBD")),
      z(z("JQ\u000btlR_ScJP^VsdZB\r")),
      z(z("JQ\u000bAK\u0017")),
      z(z("JQ\u000bBK\u0017")),
      z(z("ZBWi{\u001fYK&}MQKYz\\QIc")),
      z(z("JQ\u000bCH\u0017")),
      z(z("ZBWi{\u001fYK&yS_QYz\\QIc")),
      z(z("JQ\u000bBH\u0017")),
      z(z("JQ\u000bTH\u0017")),
      z(z("JQ\u000bDH\u0017")),
      z(z("JQ\u000bN!")),
      z(z("JQ\u000bCK\u0017")),
      z(z("JQ\u000bQ!")),
      z(z("JQ\u000bC!")),
      z(z("JQ\u000bP!")),
      z(z("JQ\u000bD!")),
      z(z("JQ\u000b@H\u0017")),
      z(z("JQ\u000btlNE@u}k_UBfH^icoKbLaaKb@ulQT\r")),
      z(z("ktiT")),
      z(z("JQ\u000bUH\u0017")),
      z(z("JQ\u000bEK\u0017")),
      z(z("JQ\u000bMK\u0017")),
      z(z("JQ\u000bU!")),
      z(z("JQ\u000bKH\u0017")),
      z(z("ZBWi{\u001fYK&}MQKuy^B@h}\u001fCUt`KU\u0005vePD\u0005tfJDLhl")),
      z(z("JQ\u000bGH\u0017")),
      z(z("JQ\u000bOK\u0017")),
      z(z("JQ\u000bK!")),
      z(z("JQ\u000bL!")),
      z(z("JQ\u000bT!")),
      z(z("JQ\u000bGK\u0017")),
      z(z("JQ\u000bJK\u0017")),
      z(z("JQ\u000b:`QYQ8!")),
      z(z("JQ\u000bWH\u0017")),
      z(z("JQ\u000bQH\u0017")),
      z(z("JQ\u000bJ!")),
      z(z("JQ\u000bV!")),
      z(z("JQ\u000bgm[sJhzJ]@t!")),
      z(z("JQ\u000bI!")),
      z(z("JQ\u000bNK\u0017")),
      z(z("\\UKr{Z@Dth\u0005\u0010")),
      z(z("JQ\u000bMH\u0017")),
      z(z("JQ\u000bLH\u0017")),
      z(z("JQ\u000bod^W@Sy[QQc!")),
      z(z("JQ\u000bKK\u0017")),
      z(z("JQ\u000bSH\u0017")),
      z(z("JQ\u000bO!")),
      z(z("JQ\u000bNH\u0017")),
      z(z("JQ\u000bAH\u0017")),
      z(z("JQ\u000bH!")),
      z(z("JQ\u000bRH\u0017"))
   };

   final void e(int var1, int var2, int var3, int var4, int var5, int var6) {
      try {
         this.b(var2, var6, var1, var3, (byte)115);
         if (var4 == 27785) {
            l++;
            this.b(var2, var6, var1, -1 + var5 + var3, (byte)-117);
            this.b(var1, var3, var6, var5, 0);
            this.b(var2 + var1 - 1, var3, var6, var5, 0);
         }
      } catch (RuntimeException var8) {
         throw i.a(var8, Yb[7] + var1 + ',' + var2 + ',' + var3 + ',' + var4 + ',' + var5 + ',' + var6 + ')');
      }
   }

   final void b(int var1, int var2, int var3, int var4, int var5, int var6, int var7) {
      boolean var20 = client.vh;

      try {
         eb++;
         if (this.hb > var1) {
            var3 -= -var1 + this.hb;
            var1 = this.hb;
         }

         if (this.lb < var3 + var1) {
            var3 = -var1 + this.lb;
         }

         int var8 = (var4 & 16747974) >> 1599703024;
         int var9 = var4 >> -272197656 & 0xFF;
         int var10 = var4 & 0xFF;
         int var11 = (var2 & 16767844) >> -411014704;
         int var12 = var2 >> 442466760 & 0xFF;
         int var13 = var2 & 0xFF;
         int var14 = -var3 + this.u;
         byte var15 = 1;
         if (this.i) {
            var14 += this.u;
            var15 = 2;
            if (0 != (var6 & 1)) {
               var6++;
               var5--;
            }
         }

         if (var7 != 19020) {
            this.a(-124, 53, -53, -76, (byte)-44);
         }

         int var16 = var1 - -(this.u * var6);
         byte var17 = 0;

         while (var5 > var17 && !var20) {
            label76: {
               if (~(var17 + var6) <= ~this.A && var6 + var17 < this.Rb) {
                  int var18 = ((var9 * var17 + var12 * (-var17 + var5)) / var5 << -1085162904)
                     + ((var11 * (-var17 + var5) + var8 * var17) / var5 << -1270717776)
                     + (var17 * var10 - -(var13 * (-var17 + var5))) / var5;
                  int var19 = -var3;

                  label71: {
                     while (-1 < ~var19) {
                        this.rb[var16++] = var18;
                        var19++;
                        if (var20) {
                           break label71;
                        }

                        if (var20) {
                           break;
                        }
                     }

                     var16 += var14;
                  }

                  if (!var20) {
                     break label76;
                  }
               }

               var16 += this.u;
            }

            var17 += var15;
            if (var20) {
               break;
            }
         }
      } catch (RuntimeException var21) {
         throw i.a(var21, Yb[10] + var1 + ',' + var2 + ',' + var3 + ',' + var4 + ',' + var5 + ',' + var6 + ',' + var7 + ')');
      }
   }

   private final void a(
      int[] var1, int var2, int var3, int var4, int var5, int[] var6, byte var7, int var8, int var9, int var10, int var11, int var12, int var13, int var14
   ) {
      boolean var20 = client.vh;

      try {
         O++;

         try {
            int var15 = var10;
            int var17 = -123 % ((-11 - var7) / 63);
            int var16 = -var9;

            while (~var16 > -1) {
               int var18 = (var5 >> -204204944) * var13;
               var5 += var8;
               if (var20) {
                  break;
               }

               int var19 = -var12;

               int var10000;
               int var10001;
               label64: {
                  while (-1 < ~var19) {
                     var4 = var1[(var10 >> 80730192) - -var18];
                     var10 += var3;
                     var10000 = 0;
                     var10001 = var4;
                     if (var20) {
                        break label64;
                     }

                     label60: {
                        if (0 != var4) {
                           var6[var14++] = var4;
                           if (!var20) {
                              break label60;
                           }
                        }

                        var14++;
                     }

                     var19++;
                     if (var20) {
                        break;
                     }
                  }

                  var14 += var11;
                  var10 = var15;
                  var10000 = var16;
                  var10001 = var2;
               }

               var16 = var10000 + var10001;
               if (var20) {
                  break;
               }
            }
         } catch (Exception var21) {
            System.out.println(Yb[47]);
         }
      } catch (RuntimeException var22) {
         throw i.a(
            var22,
            Yb[48]
               + (var1 != null ? Yb[1] : Yb[0])
               + ','
               + var2
               + ','
               + var3
               + ','
               + var4
               + ','
               + var5
               + ','
               + (var6 != null ? Yb[1] : Yb[0])
               + ','
               + var7
               + ','
               + var8
               + ','
               + var9
               + ','
               + var10
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

   private final void a(int var1, int var2, int var3, int[] var4, int var5, int var6, int var7, int[] var8, int var9, int var10, int var11, int var12) {
      boolean var17 = client.vh;

      try {
         c++;
         int var13 = 256 - var6;
         if (var9 <= -54) {
            int var14 = -var2;

            while (-1 < ~var14 && !var17) {
               int var15 = -var12;

               int var19;
               int var10000;
               while (true) {
                  if (~var15 > -1) {
                     var5 = var4[var1++];
                     var10000 = 0;
                     var19 = var5;
                     if (var17) {
                        break;
                     }

                     label58: {
                        if (0 == var5) {
                           var3++;
                           if (!var17) {
                              break label58;
                           }
                        }

                        int var16 = var8[var3];
                        var8[var3++] = ib.a(16711680, var6 * ib.a(var5, 65280) + var13 * ib.a(65280, var16))
                              + ib.a(var13 * ib.a(var16, 16711935) + ib.a(var5, 16711935) * var6, -16711936)
                           >> -379053496;
                     }

                     var15++;
                     if (!var17) {
                        continue;
                     }
                  }

                  var1 += var10;
                  var3 += var11;
                  var10000 = var14;
                  var19 = var7;
                  break;
               }

               var14 = var10000 + var19;
               if (var17) {
                  break;
               }
            }
         }
      } catch (RuntimeException var18) {
         throw i.a(
            var18,
            Yb[91]
               + var1
               + ','
               + var2
               + ','
               + var3
               + ','
               + (var4 != null ? Yb[1] : Yb[0])
               + ','
               + var5
               + ','
               + var6
               + ','
               + var7
               + ','
               + (var8 != null ? Yb[1] : Yb[0])
               + ','
               + var9
               + ','
               + var10
               + ','
               + var11
               + ','
               + var12
               + ')'
         );
      }
   }

   final void b(int var1, String var2, int var3, int var4, int var5, int var6) {
      try {
         int var7 = 24 % ((var5 - -11) / 58);
         s++;
         this.a(var1, var3, var2, var4, -12200, var6, 0);
      } catch (RuntimeException var8) {
         throw i.a(var8, Yb[69] + var1 + ',' + (var2 != null ? Yb[1] : Yb[0]) + ',' + var3 + ',' + var4 + ',' + var5 + ',' + var6 + ')');
      }
   }

   static final void a(int var0, int[] var1, int var2, int[] var3, int var4, int var5, int var6, int var7) {
      boolean var10 = client.vh;

      try {
         v++;
         if (-1 < ~var2) {
            var4 = var1[(65447 & var0) >> 389407848];
            var5 <<= 2;
            var0 += var5;
            int var8 = var2 / 16;
            int var9 = var8;

            while (true) {
               if (-1 < ~var9) {
                  var3[var6++] = var4 - -ib.a(8355711, var3[var6] >> 980406721);
                  var3[var6++] = var4 + ib.a(8355711, var3[var6] >> -33478591);
                  var3[var6++] = var4 + (ib.a(var3[var6], 16711423) >> 1438034561);
                  var3[var6++] = (ib.a(16711423, var3[var6]) >> -1664277151) + var4;
                  var4 = var1[0xFF & var0 >> 1869141800];
                  var0 += var5;
                  var3[var6++] = ib.a(var3[var6] >> -651215775, 8355711) + var4;
                  var3[var6++] = ib.a(var3[var6] >> 1567416321, 8355711) + var4;
                  var3[var6++] = (ib.a(16711423, var3[var6]) >> -109945983) + var4;
                  var3[var6++] = ib.a(var3[var6] >> -1634216127, 8355711) + var4;
                  var4 = var1[var0 >> 1972579688 & 0xFF];
                  var3[var6++] = (ib.a(var3[var6], 16711422) >> 18481057) + var4;
                  var0 += var5;
                  var3[var6++] = (ib.a(16711423, var3[var6]) >> 1645567265) + var4;
                  var3[var6++] = var4 - -ib.a(var3[var6] >> 363686529, 8355711);
                  var3[var6++] = var4 + ib.a(8355711, var3[var6] >> -417782847);
                  var4 = var1[(65302 & var0) >> -491054904];
                  var0 += var5;
                  var3[var6++] = (ib.a(var3[var6], 16711423) >> -1655491807) + var4;
                  var3[var6++] = var4 + ib.a(var3[var6] >> 421283745, 8355711);
                  var3[var6++] = var4 - -(ib.a(var3[var6], 16711423) >> 1309685921);
                  var3[var6++] = var4 + (ib.a(16711423, var3[var6]) >> 1995672417);
                  var4 = var1[var0 >> 1728371944 & 0xFF];
                  var0 += var5;
                  var9++;
                  if (var10) {
                     break;
                  }

                  if (!var10) {
                     continue;
                  }
               }

               var8 = -(var2 % 16);
               break;
            }

            var9 = var7;

            while (~var9 > ~var8) {
               var3[var6++] = ib.a(var3[var6] >> 1543799489, 8355711) + var4;
               if (var10) {
                  break;
               }

               if ((3 & var9) == 3) {
                  var4 = var1[(var0 & 65336) >> -300394456];
                  var0 += var5;
                  var0 += var5;
               }

               var9++;
               if (var10) {
                  break;
               }
            }
         }
      } catch (RuntimeException var11) {
         throw i.a(
            var11,
            Yb[63]
               + var0
               + ','
               + (var1 != null ? Yb[1] : Yb[0])
               + ','
               + var2
               + ','
               + (var3 != null ? Yb[1] : Yb[0])
               + ','
               + var4
               + ','
               + var5
               + ','
               + var6
               + ','
               + var7
               + ')'
         );
      }
   }

   @Override
   public final synchronized void addConsumer(ImageConsumer var1) {
      try {
         this.fb = var1;
         r++;
         var1.setDimensions(this.u, this.k);
         var1.setProperties(null);
         var1.setColorModel(this.nb);
         var1.setHints(14);
      } catch (RuntimeException var3) {
         throw i.a(var3, Yb[78] + (var1 != null ? Yb[1] : Yb[0]) + ')');
      }
   }

   final void b(int var1, int var2, int var3, int var4) {
      try {
         if (var1 != -1) {
            this.a(41, 58, 102, 22, (byte)102);
         }

         Lb++;
         if (this.Qb[var2]) {
            var4 += this.Sb[var2];
            var3 += this.G[var2];
         }

         int var5 = var3 * this.u + var4;
         int var6 = 0;
         int var7 = this.R[var2];
         int var8 = this.kb[var2];
         int var9 = -var8 + this.u;
         int var10 = 0;
         if (~var3 > ~this.A) {
            int var11 = this.A - var3;
            var7 -= var11;
            var3 = this.A;
            var5 += this.u * var11;
            var6 += var11 * var8;
         }

         if (~this.Rb >= ~(var3 - -var7)) {
            var7 -= 1 + (var7 + var3 - this.Rb);
         }

         if (~this.hb < ~var4) {
            int var13 = -var4 + this.hb;
            var6 += var13;
            var9 += var13;
            var8 -= var13;
            var10 += var13;
            var4 = this.hb;
            var5 += var13;
         }

         if (var4 - -var8 >= this.lb) {
            int var14 = var4 - -var8 + -this.lb - -1;
            var8 -= var14;
            var10 += var14;
            var9 += var14;
         }

         if (0 < var8 && var7 > 0) {
            byte var15 = 1;
            if (this.i) {
               var9 += this.u;
               if (~(1 & var3) != -1) {
                  var5 += this.u;
                  var7--;
               }

               var15 = 2;
               var10 += this.kb[var2];
            }

            if (this.ob[var2] != null) {
               this.a(var8, this.rb, var15, var7, 0, var6, (byte)123, var5, this.ob[var2], var9, var10);
               if (!client.vh) {
                  return;
               }
            }

            this.a(var5, this.Y[var2], var6, var10, this.rb, var15, var7, true, this.gb[var2], var8, var9);
         }
      } catch (RuntimeException var12) {
         throw i.a(var12, Yb[17] + var1 + ',' + var2 + ',' + var3 + ',' + var4 + ')');
      }
   }

   @Override
   public final void requestTopDownLeftRightResend(ImageConsumer var1) {
      try {
         Ib++;
         System.out.println(Yb[59]);
      } catch (RuntimeException var3) {
         throw i.a(var3, Yb[58] + (var1 != null ? Yb[1] : Yb[0]) + ')');
      }
   }

   final void a(int var1, int var2, int var3, int var4, int var5, int var6, int var7) {
      boolean var17 = client.vh;

      try {
         int var8 = var4;

         int var19;
         label96:
         while (true) {
            var19 = ~(var4 + var6);

            label93:
            while (var19 < ~var8) {
               var19 = var3;
               if (var17) {
                  break label96;
               }

               int var9 = var3;

               label90:
               while (true) {
                  var19 = ~var9;

                  label88:
                  while (true) {
                     if (var19 <= ~(var3 + var1)) {
                        break label90;
                     }

                     int var10 = 0;
                     int var11 = 0;
                     int var12 = 0;
                     int var13 = 0;
                     var19 = -var7 + var8;
                     if (var17) {
                        continue label93;
                     }

                     int var14 = var19;

                     label85:
                     while (true) {
                        var19 = ~(var7 + var8);
                        int var10001 = ~var14;

                        label83:
                        while (true) {
                           if (var19 > var10001) {
                              break label85;
                           }

                           var19 = var14;
                           if (var17) {
                              continue label88;
                           }

                           if (var14 >= 0 && this.u > var14) {
                              int var15 = var9 - var2;

                              while (var15 <= var9 + var2) {
                                 var19 = -1;
                                 var10001 = ~var15;
                                 if (var17) {
                                    continue label83;
                                 }

                                 if (-1 >= var10001 && ~var15 > ~this.k) {
                                    int var16 = this.rb[this.u * var15 + var14];
                                    var12 += 0xFF & var16;
                                    var13++;
                                    var11 += (var16 & 65409) >> 743340392;
                                    var10 += (var16 & 16737446) >> 483715504;
                                 }

                                 var15++;
                                 if (var17) {
                                    break;
                                 }
                              }
                           }

                           var14++;
                           if (var17) {
                              break label85;
                           }
                           break;
                        }
                     }

                     this.rb[var8 + var9 * this.u] = var12 / var13 + (var10 / var13 << -148272656) + (var11 / var13 << 2002983304);
                     var9++;
                     if (var17) {
                        break label90;
                     }
                     break;
                  }
               }

               var8++;
               if (!var17) {
                  continue label96;
               }
               break;
            }

            Vb++;
            var19 = var5;
            break;
         }

         if (var19 != 16740352) {
            this.a(-18, (byte)79, -10, 106, -42, 27);
         }
      } catch (RuntimeException var18) {
         throw i.a(var18, Yb[20] + var1 + ',' + var2 + ',' + var3 + ',' + var4 + ',' + var5 + ',' + var6 + ',' + var7 + ')');
      }
   }

   final void a(int var1, int var2, int var3, int var4, int var5, byte var6, int var7) {
      try {
         ub++;

         try {
            int var8 = this.kb[var3];
            int var9 = this.R[var3];
            int var10 = 0;
            int var11 = 0;
            int var12 = (var8 << -1350099408) / var7;
            int var13 = (var9 << 653525744) / var5;
            if (this.Qb[var3]) {
               int var14 = this.Eb[var3];
               int var15 = this.qb[var3];
               if (0 == var14 || -1 == ~var15) {
                  return;
               }

               if (~(this.Sb[var3] * var7 % var14) != -1) {
                  var10 = (-(this.Sb[var3] * var7 % var14) + var14 << -1137453552) / var7;
               }

               var1 += (-1 + (var7 * this.Sb[var3] - -var14)) / var14;
               var12 = (var14 << -716906352) / var7;
               var4 += (var15 + var5 * this.G[var3] - 1) / var15;
               var13 = (var15 << 1305987728) / var5;
               if (-1 != ~(this.G[var3] * var5 % var15)) {
                  var11 = (-(var5 * this.G[var3] % var15) + var15 << -1890544144) / var5;
               }

               var5 = var5 * (-(var11 >> 1020185680) + this.R[var3]) / var15;
               var7 = (-(var10 >> -839014480) + this.kb[var3]) * var7 / var14;
            }

            if (var6 <= 102) {
               this.Y = (int[][])null;
            }

            int var19 = var1 + var4 * this.u;
            int var20 = this.u - var7;
            if (~var4 > ~this.A) {
               int var16 = this.A + -var4;
               var4 = 0;
               var5 -= var16;
               var19 += var16 * this.u;
               var11 += var16 * var13;
            }

            if (~var1 > ~this.hb) {
               int var21 = this.hb - var1;
               var7 -= var21;
               var20 += var21;
               var10 += var21 * var12;
               var19 += var21;
               var1 = 0;
            }

            if (~(var4 - -var5) <= ~this.Rb) {
               var5 -= var4 - -var5 - (this.Rb + -1);
            }

            if (~(var1 + var7) <= ~this.lb) {
               int var22 = var1 - -var7 + (-this.lb - -1);
               var7 -= var22;
               var20 += var22;
            }

            byte var23 = 1;
            if (this.i) {
               var13 += var13;
               if ((1 & var4) != 0) {
                  var5--;
                  var19 += this.u;
               }

               var23 = 2;
               var20 += this.u;
            }

            this.a(var11, var12, var7, var10, var20, this.ob[var3], var19, this.rb, 0, var8, false, var13, var5, var2, var23);
         } catch (Exception var17) {
            System.out.println(Yb[16]);
         }
      } catch (RuntimeException var18) {
         throw i.a(var18, Yb[57] + var1 + ',' + var2 + ',' + var3 + ',' + var4 + ',' + var5 + ',' + var6 + ',' + var7 + ')');
      }
   }

   final void b(int var1) {
      boolean var5 = client.vh;

      try {
         Z++;
         int var4 = this.k * this.u;
         if (var1 != 16316665) {
            wb = (String[])null;
         }

         int var3 = 0;

         while (var4 > var3) {
            int var2 = 16777215 & this.rb[var3];
            this.rb[var3] = ib.a(var2 >>> -93223452, 986895)
               + ((ib.a(var2, 16316665) >>> 2097500387) + (ib.a(16711423, var2) >>> 1336934849) - -ib.a(-2143338689, var2 >>> 1527263298));
            var3++;
            if (var5 || var5) {
               break;
            }
         }
      } catch (RuntimeException var6) {
         throw i.a(var6, Yb[55] + var1 + ')');
      }
   }

   private final void a(int var1, int var2, int var3, int[] var4, int var5, int var6, int var7, int[] var8, int var9, int var10, boolean var11) {
      boolean var13 = client.vh;

      try {
         var3 = var1;

         while (true) {
            if (var3 < 0) {
               this.rb[var10++] = var8[(var6 >> -1782373679) * var2 + (var7 >> -324278255)];
               var7 += var5;
               var6 += var9;
               var3++;
               if (var13) {
                  break;
               }

               if (!var13) {
                  continue;
               }
            }

            if (!var11) {
               this.a(-59, -116, -115, true, 1, 118, 33, -46, -78, -30);
            }

            V++;
            break;
         }
      } catch (RuntimeException var14) {
         throw i.a(
            var14,
            Yb[68]
               + var1
               + ','
               + var2
               + ','
               + var3
               + ','
               + (var4 != null ? Yb[1] : Yb[0])
               + ','
               + var5
               + ','
               + var6
               + ','
               + var7
               + ','
               + (var8 != null ? Yb[1] : Yb[0])
               + ','
               + var9
               + ','
               + var10
               + ','
               + var11
               + ')'
         );
      }
   }

   final void a(int var1, int var2, int var3, int var4) {
      try {
         db++;
         if (this.hb <= var2 && ~var1 <= ~this.A && ~var2 > ~this.lb && ~var1 > ~this.Rb) {
            if (var3 > 44) {
               this.rb[var2 - -(this.u * var1)] = var4;
            }
         }
      } catch (RuntimeException var6) {
         throw i.a(var6, Yb[61] + var1 + ',' + var2 + ',' + var3 + ',' + var4 + ')');
      }
   }

   final void c(int var1, int var2, int var3, int var4, int var5, int var6, int var7) {
      boolean var21 = client.vh;

      try {
         z++;
         if (var5 < this.A) {
            var3 -= this.A - var5;
            var5 = this.A;
         }

         if (~var2 > ~this.hb) {
            var6 -= this.hb - var2;
            var2 = this.hb;
         }

         if (this.lb < var2 - -var6) {
            var6 = this.lb - var2;
         }

         if (this.Rb < var3 + var5) {
            var3 = -var5 + this.Rb;
         }

         int var8 = 256 - var1;
         int var9 = var1 * (var7 >> -2055680880 & 0xFF);
         int var10 = ((var7 & 65476) >> 1364192264) * var1;
         int var11 = var1 * (var7 & 0xFF);
         int var15 = this.u - var6;
         byte var16 = 1;
         if (this.i) {
            if (-1 != ~(var5 & 1)) {
               var3--;
               var5++;
            }

            var15 += this.u;
            var16 = 2;
         }

         int var17 = var2 - -(this.u * var5);
         int var18 = var4;

         while (~var3 < ~var18 && !var21) {
            int var19 = -var6;

            while (true) {
               if (0 > var19) {
                  int var14 = var8 * (this.rb[var17] & 0xFF);
                  int var12 = var8 * ((16746549 & this.rb[var17]) >> 1661674448);
                  int var13 = var8 * ((65402 & this.rb[var17]) >> 2108168104);
                  int var20 = (var14 + var11 >> -1220075704) + (var13 + var10 >> -855772152 << -628820632) + (var9 + var12 >> -540786712 << -681889584);
                  this.rb[var17++] = var20;
                  var19++;
                  if (var21) {
                     break;
                  }

                  if (!var21) {
                     continue;
                  }
               }

               var17 += var15;
               var18 += var16;
               break;
            }

            if (var21) {
               break;
            }
         }
      } catch (RuntimeException var22) {
         throw i.a(var22, Yb[62] + var1 + ',' + var2 + ',' + var3 + ',' + var4 + ',' + var5 + ',' + var6 + ',' + var7 + ')');
      }
   }

   final int a(int var1, int var2, String var3) {
      boolean var8 = client.vh;

      try {
         Nb++;
         byte var4 = 0;
         if (var2 <= 67) {
            this.Rb = 74;
         }

         byte[] var5 = m.b[var1];
         int var6 = 0;

         byte var10000;
         while (true) {
            if (var6 < var3.length()) {
               var10000 = 64;
               if (var8) {
                  break;
               }

               label93: {
                  if ('@' == var3.charAt(var6) && 4 + var6 < var3.length() && var3.charAt(var6 - -4) == '@') {
                     var6 += 4;
                     if (!var8) {
                        break label93;
                     }
                  }

                  if (~var3.charAt(var6) != -127 || var6 + 4 >= var3.length() || '~' != var3.charAt(var6 + 4)) {
                     char var7 = var3.charAt(var6);
                     if (~var7 > -1 || ~n.a.length >= ~var7) {
                        var7 = ' ';
                     }

                     var4 += var5[n.a[var7] + 7];
                     if (!var8) {
                        break label93;
                     }
                  }

                  var6 += 4;
               }

               var6++;
               if (!var8) {
                  continue;
               }
            }

            var10000 = var4;
            break;
         }

         return var10000;
      } catch (RuntimeException var9) {
         throw i.a(var9, Yb[11] + var1 + 44 + var2 + 44 + (var3 != null ? Yb[1] : Yb[0]) + 41);
      }
   }

   @Override
   public final boolean imageUpdate(Image var1, int var2, int var3, int var4, int var5, int var6) {
      try {
         B++;
         return true;
      } catch (RuntimeException var8) {
         throw i.a(var8, Yb[84] + (var1 != null ? Yb[1] : Yb[0]) + ',' + var2 + ',' + var3 + ',' + var4 + ',' + var5 + ',' + var6 + ')');
      }
   }

   private final void a(int var1, int var2, int var3, int[] var4, int[] var5, int var6, int var7, int var8, int var9, int var10, byte var11) {
      boolean var13 = client.vh;

      try {
         yb++;
         int var12 = var2;
         if (var11 == 102) {
            while (-1 < ~var12) {
               var10 = var4[(var6 >> 930333777) * var8 + (var3 >> 98213361)];
               var3 += var9;
               if (var13) {
                  break;
               }

               label47: {
                  if (-1 == ~var10) {
                     var7++;
                     if (!var13) {
                        break label47;
                     }
                  }

                  this.rb[var7++] = var10;
               }

               var6 += var1;
               var12++;
               if (var13) {
                  break;
               }
            }
         }
      } catch (RuntimeException var14) {
         throw i.a(
            var14,
            Yb[86]
               + var1
               + ','
               + var2
               + ','
               + var3
               + ','
               + (var4 != null ? Yb[1] : Yb[0])
               + ','
               + (var5 != null ? Yb[1] : Yb[0])
               + ','
               + var6
               + ','
               + var7
               + ','
               + var8
               + ','
               + var9
               + ','
               + var10
               + ','
               + var11
               + ')'
         );
      }
   }

   private final void a(int var1, int var2, byte[] var3, int var4, boolean var5, int var6, int var7, int var8, int var9, int[] var10, int[] var11, int var12) {
      boolean var18 = client.vh;

      try {
         U++;
         int var13 = -var8 + 256;
         if (!var5) {
            int var14 = -var6;

            while (0 > var14 && !var18) {
               int var15 = -var9;

               int var21;
               int var10000;
               while (true) {
                  if (var15 < 0) {
                     int var16 = var3[var1++];
                     var10000 = ~var16;
                     var21 = -1;
                     if (var18) {
                        break;
                     }

                     label64: {
                        if (var10000 == -1) {
                           var12++;
                           if (!var18) {
                              break label64;
                           }
                        }

                        var16 = var11[0xFF & var16];
                        int var17 = var10[var12];
                        var10[var12++] = ib.a(var13 * ib.a(var17, 16711935) + ib.a(16711935, var16) * var8, -16711936)
                              + ib.a(16711680, var8 * ib.a(65280, var16) + var13 * ib.a(65280, var17))
                           >> 1273033224;
                     }

                     var15++;
                     if (!var18) {
                        continue;
                     }
                  }

                  var12 += var2;
                  var1 += var7;
                  var10000 = var14;
                  var21 = var4;
                  break;
               }

               var14 = var10000 + var21;
               if (var18) {
                  break;
               }
            }
         }
      } catch (RuntimeException var19) {
         throw i.a(
            var19,
            Yb[76]
               + var1
               + ','
               + var2
               + ','
               + (var3 != null ? Yb[1] : Yb[0])
               + ','
               + var4
               + ','
               + var5
               + ','
               + var6
               + ','
               + var7
               + ','
               + var8
               + ','
               + var9
               + ','
               + (var10 != null ? Yb[1] : Yb[0])
               + ','
               + (var11 != null ? Yb[1] : Yb[0])
               + ','
               + var12
               + ')'
         );
      }
   }

   private final int c(int var1, int var2) {
      try {
         o++;
         if (~var2 != -1) {
            if (var1 < 49) {
               this.a(-22, 77, 112, -35, -44, (int[])null, -45, (int[])null, -39, -33, false, 50, 61, 37, -7);
            }

            return m.b[var2][8] + -1;
         } else {
            return m.b[var2][8] - 2;
         }
      } catch (RuntimeException var4) {
         throw i.a(var4, Yb[88] + var1 + 44 + var2 + 41);
      }
   }

   final void b(int var1, int var2, int var3, int var4, int var5, int var6) {
      boolean var11 = client.vh;

      try {
         ib++;
         this.kb[var5] = var6;
         this.R[var5] = var1;
         this.Qb[var5] = false;
         this.Sb[var5] = 0;
         this.G[var5] = 0;
         this.Eb[var5] = var6;
         this.qb[var5] = var1;
         int var7 = var1 * var6;
         int var8 = 0;
         this.ob[var5] = new int[var7];
         int var9 = var2;

         int var10000;
         while (true) {
            if (var2 + var6 > var9) {
               var10000 = var3;
               if (var11) {
                  break;
               }

               int var10 = var3;

               label39: {
                  while (~(var1 + var3) < ~var10) {
                     this.ob[var5][var8++] = this.rb[var9 + this.u * var10];
                     var10++;
                     if (var11) {
                        break label39;
                     }

                     if (var11) {
                        break;
                     }
                  }

                  var9++;
               }

               if (!var11) {
                  continue;
               }
            }

            var10000 = var4;
            break;
         }

         if (var10000 != -27966) {
            this.a(73, -62, -30, (byte)-113, 44, -64, -91, 100, -79, (int[])null, 117, 11, 127, -109, (int[])null);
         }
      } catch (RuntimeException var12) {
         throw i.a(var12, Yb[18] + var1 + ',' + var2 + ',' + var3 + ',' + var4 + ',' + var5 + ',' + var6 + ')');
      }
   }

   void a(int var1, int var2, int var3, int var4, int var5, int var6, byte var7, int var8) {
      try {
         this.f(var4, var5, var3, var6, 5924, var2);
         if (var7 == 29) {
            T++;
         }
      } catch (RuntimeException var10) {
         throw i.a(var10, Yb[56] + var1 + ',' + var2 + ',' + var3 + ',' + var4 + ',' + var5 + ',' + var6 + ',' + var7 + ',' + var8 + ')');
      }
   }

   final void c(int var1, int var2, int var3, int var4, int var5, int var6) {
      boolean var25 = client.vh;

      try {
         X++;
         int var7 = -var1 + 256;
         int var8 = var1 * (0xFF & var5 >> -25949072);
         int var9 = (0xFF & var5 >> -1057205208) * var1;
         int var10 = var1 * (var5 & 0xFF);
         if (var2 == -1057205208) {
            int var14 = -var3 + var4;
            if (~var14 > -1) {
               var14 = 0;
            }

            int var15 = var4 - -var3;
            if (this.k <= var15) {
               var15 = -1 + this.k;
            }

            byte var16 = 1;
            if (this.i) {
               if ((1 & var14) != 0) {
                  var14++;
               }

               var16 = 2;
            }

            int var17 = var14;

            while (var15 >= var17) {
               int var18 = var17 + -var4;
               int var19 = (int)Math.sqrt(-(var18 * var18) + var3 * var3);
               int var20 = var6 + -var19;
               if (var25) {
                  break;
               }

               if (~var20 > -1) {
                  var20 = 0;
               }

               int var21 = var6 - -var19;
               if (this.u <= var21) {
                  var21 = this.u + -1;
               }

               int var22 = var20 + this.u * var17;
               int var23 = var20;

               while (true) {
                  if (var21 >= var23) {
                     int var13 = (this.rb[var22] & 0xFF) * var7;
                     int var12 = var7 * ((65493 & this.rb[var22]) >> -1460723256);
                     int var11 = var7 * ((this.rb[var22] & 16718009) >> -342059728);
                     int var24 = (var13 + var10 >> 1268749032) + (var9 - -var12 >> -404063160 << -187266712) + (var8 - -var11 >> 296955496 << 1872374416);
                     this.rb[var22++] = var24;
                     var23++;
                     if (var25) {
                        break;
                     }

                     if (!var25) {
                        continue;
                     }
                  }

                  var17 += var16;
                  break;
               }

               if (var25) {
                  break;
               }
            }
         }
      } catch (RuntimeException var26) {
         throw i.a(var26, Yb[75] + var1 + ',' + var2 + ',' + var3 + ',' + var4 + ',' + var5 + ',' + var6 + ')');
      }
   }

   private final void a(
      int[] var1,
      int[] var2,
      int var3,
      int var4,
      int var5,
      int var6,
      int var7,
      int var8,
      int var9,
      int var10,
      int var11,
      int var12,
      int var13,
      int var14,
      int var15,
      int var16,
      int var17
   ) {
      boolean var33 = client.vh;

      try {
         a++;
         int var21 = var15 >> 1196549680 & 0xFF;
         int var22 = var15 >> -340653432 & 0xFF;
         int var23 = var15 & 0xFF;
         int var24 = var8 >> -922169040 & 0xFF;
         if (var6 != 1603920392) {
            this.Rb = 29;
         }

         int var25 = var8 >> 5163400 & 0xFF;
         int var26 = 0xFF & var8;

         try {
            int var27 = var11;
            int var28 = -var16;

            while (-1 < ~var28) {
               int var29 = (var13 >> -1328879408) * var14;
               int var30 = var5 >> -299524176;
               int var31 = var3;
               if (var33) {
                  break;
               }

               if (~var30 > ~this.hb) {
                  int var32 = this.hb + -var30;
                  var31 -= var32;
                  var11 += var32 * var10;
                  var30 = this.hb;
               }

               if (this.lb <= var30 - -var31) {
                  int var36 = -this.lb + var30 - -var31;
                  var31 -= var36;
               }

               int var10000;
               int var10001;
               label109: {
                  var12 = 1 - var12;
                  if (0 != var12) {
                     int var37 = var30;

                     while (var30 + var31 > var37) {
                        var7 = var2[var29 + (var11 >> -1009344688)];
                        var10000 = -1;
                        var10001 = ~var7;
                        if (var33) {
                           break label109;
                        }

                        label103:
                        if (-1 != var10001) {
                           int var18 = 0xFF & var7 >> 2008995472;
                           int var19 = 0xFF & var7 >> 1043792552;
                           int var20 = var7 & 0xFF;
                           if (var18 == var19 && var20 == var19) {
                              var1[var37 - -var17] = (var20 * var23 >> -1147526104)
                                 + (var22 * var19 >> 1601552776 << -566695064)
                                 + (var18 * var21 >> 1603920392 << -538385168);
                              if (!var33) {
                                 break label103;
                              }
                           }

                           if (-256 != ~var18 || ~var20 != ~var19) {
                              var1[var37 + var17] = var7;
                              if (!var33) {
                                 break label103;
                              }
                           }

                           var1[var37 - -var17] = (var26 * var20 >> -1345290488)
                              + (var18 * var24 >> 1685589832 << -1431216016)
                              - -(var19 * var25 >> 508305352 << -483710840);
                        }

                        var11 += var10;
                        var37++;
                        if (var33) {
                           break;
                        }
                     }
                  }

                  var13 += var9;
                  var11 = var27;
                  var17 += this.u;
                  var10000 = var5;
                  var10001 = var4;
               }

               var5 = var10000 + var10001;
               var28++;
               if (var33) {
                  break;
               }
            }
         } catch (Exception var34) {
            System.out.println(Yb[65]);
         }
      } catch (RuntimeException var35) {
         throw i.a(
            var35,
            Yb[66]
               + (var1 != null ? Yb[1] : Yb[0])
               + ','
               + (var2 != null ? Yb[1] : Yb[0])
               + ','
               + var3
               + ','
               + var4
               + ','
               + var5
               + ','
               + var6
               + ','
               + var7
               + ','
               + var8
               + ','
               + var9
               + ','
               + var10
               + ','
               + var11
               + ','
               + var12
               + ','
               + var13
               + ','
               + var14
               + ','
               + var15
               + ','
               + var16
               + ','
               + var17
               + ')'
         );
      }
   }

   private final void a(
      int var1,
      int[] var2,
      int var3,
      int var4,
      int var5,
      int var6,
      int var7,
      int var8,
      byte[] var9,
      int var10,
      byte var11,
      int var12,
      int var13,
      int var14,
      int var15,
      int var16,
      int[] var17,
      int var18
   ) {
      boolean var34 = client.vh;

      try {
         J++;
         int var22 = var18 >> 1698138864 & 0xFF;
         if (var11 > 8) {
            int var23 = (var18 & 65381) >> 1758159464;
            int var24 = var18 & 0xFF;
            int var25 = var13 >> 34872048 & 0xFF;
            int var26 = 0xFF & var13 >> 225029096;
            int var27 = 0xFF & var13;

            try {
               int var28 = var12;
               int var29 = -var1;

               while (0 > var29) {
                  int var30 = var3 * (var8 >> -1095936688);
                  int var31 = var14 >> -2034604816;
                  int var32 = var16;
                  if (var34) {
                     break;
                  }

                  if (~var31 > ~this.hb) {
                     int var33 = this.hb + -var31;
                     var12 += var33 * var15;
                     var32 -= var33;
                     var31 = this.hb;
                  }

                  var10 = -var10 + 1;
                  if (var32 + var31 >= this.lb) {
                     int var37 = var31 + (var32 - this.lb);
                     var32 -= var37;
                  }

                  int var10000;
                  label115: {
                     if (var10 != 0) {
                        int var38 = var31;

                        while (var38 < var32 + var31) {
                           var5 = 255 & var9[(var12 >> -91857424) - -var30];
                           var10000 = var5;
                           if (var34) {
                              break label115;
                           }

                           label109:
                           if (var5 != 0) {
                              var5 = var2[var5];
                              int var21 = var5 & 0xFF;
                              int var20 = 0xFF & var5 >> 1941812904;
                              int var19 = 0xFF & var5 >> -309996368;
                              if (var20 != var19 || ~var20 != ~var21) {
                                 if (-256 == ~var19 && var20 == var21) {
                                    var17[var7 + var38] = (var25 * var19 >> 1700051816 << -1733973648)
                                       + (var20 * var26 >> 269461512 << 971885320)
                                       + (var27 * var21 >> -1282776696);
                                    if (!var34) {
                                       break label109;
                                    }
                                 }

                                 var17[var38 + var7] = var5;
                                 if (!var34) {
                                    break label109;
                                 }
                              }

                              var17[var38 + var7] = (var24 * var21 >> -178220952)
                                 + (var19 * var22 >> -1022678840 << 2122310768)
                                 + (var23 * var20 >> 901102216 << -552193592);
                           }

                           var12 += var15;
                           var38++;
                           if (var34) {
                              break;
                           }
                        }
                     }

                     var8 += var4;
                     var14 += var6;
                     var12 = var28;
                     var10000 = var7 + this.u;
                  }

                  var7 = var10000;
                  var29++;
                  if (var34) {
                     break;
                  }
               }
            } catch (Exception var35) {
               System.out.println(Yb[65]);
            }
         }
      } catch (RuntimeException var36) {
         throw i.a(
            var36,
            Yb[85]
               + var1
               + ','
               + (var2 != null ? Yb[1] : Yb[0])
               + ','
               + var3
               + ','
               + var4
               + ','
               + var5
               + ','
               + var6
               + ','
               + var7
               + ','
               + var8
               + ','
               + (var9 != null ? Yb[1] : Yb[0])
               + ','
               + var10
               + ','
               + var11
               + ','
               + var12
               + ','
               + var13
               + ','
               + var14
               + ','
               + var15
               + ','
               + var16
               + ','
               + (var17 != null ? Yb[1] : Yb[0])
               + ','
               + var18
               + ')'
         );
      }
   }

   private final void a(byte var1, boolean var2, byte[] var3, int var4, int var5, int var6, int var7) {
      try {
         P++;
         int var8 = var4 - -var3[5 + var6];
         if (var1 >= 24) {
            int var9 = var7 + -var3[6 + var6];
            int var10 = var3[var6 - -3];
            int var11 = var3[4 + var6];
            int var12 = 16384 * var3[var6] + var3[var6 - -1] * 128 + var3[var6 + 2];
            int var13 = var8 - -(var9 * this.u);
            int var14 = this.u + -var10;
            if (~this.A < ~var9) {
               int var16 = this.A - var9;
               var12 += var16 * var10;
               var13 += this.u * var16;
               var11 -= var16;
               var9 = this.A;
            }

            int var15 = 0;
            if (~this.Rb >= ~(var9 + var11)) {
               var11 -= 1 + var9 + var11 + -this.Rb;
            }

            if (~var8 > ~this.hb) {
               int var18 = -var8 + this.hb;
               var15 += var18;
               var10 -= var18;
               var12 += var18;
               var8 = this.hb;
               var14 += var18;
               var13 += var18;
            }

            if (~(var10 + var8) <= ~this.lb) {
               int var19 = -this.lb + var10 + var8 + 1;
               var14 += var19;
               var15 += var19;
               var10 -= var19;
            }

            if (0 < var10 && -1 > ~var11) {
               if (var2) {
                  this.a(var3, var5, var10, var13, var11, var15, 1504725224, var14, this.rb, var12);
                  if (!client.vh) {
                     return;
                  }
               }

               this.a(var5, this.rb, var13, (byte)37, var14, var11, var10, var12, var3, var15);
            }
         }
      } catch (RuntimeException var17) {
         throw i.a(var17, Yb[60] + var1 + ',' + var2 + ',' + (var3 != null ? Yb[1] : Yb[0]) + ',' + var4 + ',' + var5 + ',' + var6 + ',' + var7 + ')');
      }
   }

   final void f(int var1, int var2, int var3, int var4, int var5, int var6) {
      try {
         cb++;

         try {
            int var7 = this.kb[var6];
            int var8 = this.R[var6];
            int var9 = 0;
            int var10 = 0;
            int var11 = (var7 << -311879024) / var4;
            int var12 = (var8 << -1693975408) / var3;
            if (this.Qb[var6]) {
               int var13 = this.Eb[var6];
               int var14 = this.qb[var6];
               if (~var13 == -1 || var14 == 0) {
                  return;
               }

               if (-1 != ~(this.G[var6] * var3 % var14)) {
                  var10 = (var14 - var3 * this.G[var6] % var14 << 1500479664) / var3;
               }

               var11 = (var13 << 1015741296) / var4;
               if (0 != this.Sb[var6] * var4 % var13) {
                  var9 = (var13 - this.Sb[var6] * var4 % var13 << -1148345008) / var4;
               }

               var1 += (-1 + (var4 * this.Sb[var6] - -var13)) / var13;
               var12 = (var14 << -959198096) / var3;
               var2 += (var14 + var3 * this.G[var6] - 1) / var14;
               var3 = (this.R[var6] - (var10 >> 1117428208)) * var3 / var14;
               var4 = var4 * (this.kb[var6] - (var9 >> -725293584)) / var13;
            }

            int var18 = var1 + this.u * var2;
            if (~this.A < ~var2) {
               int var15 = this.A - var2;
               var10 += var12 * var15;
               var3 -= var15;
               var18 += this.u * var15;
               var2 = 0;
            }

            int var19 = this.u + -var4;
            if (~this.Rb >= ~(var2 - -var3)) {
               var3 -= -this.Rb + var2 - -var3 - -1;
            }

            if (var1 < this.hb) {
               int var20 = -var1 + this.hb;
               var4 -= var20;
               var19 += var20;
               var18 += var20;
               var1 = 0;
               var9 += var11 * var20;
            }

            if (this.lb <= var1 + var4) {
               int var21 = 1 + var1 + (var4 - this.lb);
               var19 += var21;
               var4 -= var21;
            }

            byte var22 = 1;
            if (this.i) {
               if (0 != (var2 & 1)) {
                  var3--;
                  var18 += this.u;
               }

               var19 += this.u;
               var22 = 2;
               var12 += var12;
            }

            this.a(this.ob[var6], var22, var11, 0, var10, this.rb, (byte)78, var12, var3, var9, var19, var4, var7, var18);
         } catch (Exception var16) {
            System.out.println(Yb[16]);
         }

         if (var5 != 5924) {
            this.u = -15;
         }
      } catch (RuntimeException var17) {
         throw i.a(var17, Yb[15] + var1 + ',' + var2 + ',' + var3 + ',' + var4 + ',' + var5 + ',' + var6 + ')');
      }
   }

   final void b(int var1, int var2, int var3, int var4, byte var5) {
      boolean var9 = client.vh;

      try {
         F++;
         if (this.A <= var4 && var4 < this.Rb) {
            if (~var3 > ~this.hb) {
               var1 -= -var3 + this.hb;
               var3 = this.hb;
            }

            if (~this.lb > ~(var3 - -var1)) {
               var1 = -var3 + this.lb;
            }

            int var6 = -44 / ((var5 - 15) / 37);
            if (0 < var1) {
               int var7 = var3 - -(this.u * var4);
               int var8 = 0;

               while (var1 > var8) {
                  this.rb[var7 - -var8] = var2;
                  var8++;
                  if (var9 || var9) {
                     break;
                  }
               }
            }
         }
      } catch (RuntimeException var10) {
         throw i.a(var10, Yb[72] + var1 + ',' + var2 + ',' + var3 + ',' + var4 + ',' + var5 + ')');
      }
   }

   final void a(int var1, int var2, String var3, int var4, int var5, byte var6, int var7) {
      boolean var14 = client.vh;

      try {
         int var10000;
         int var10001;
         label271: {
            try {
               if (~var1 < -1) {
                  int var8 = -1 + var1 + this.D;
                  if (this.gb[var8] != null) {
                     this.b(-1, var8, -this.R[var8] + var2, var4);
                     var4 += this.kb[var8] - -5;
                  }
               }

               int var9 = -5 % ((var6 - -91) / 33);
               byte[] var17 = m.b[var7];
               int var10 = 0;

               while (var3.length() > var10) {
                  var10000 = -65;
                  var10001 = ~var3.charAt(var10);
                  if (var14) {
                     break label271;
                  }

                  label285: {
                     if (-65 == var10001 && ~var3.length() < ~(var10 - -4) && var3.charAt(var10 - -4) == '@') {
                        label260: {
                           if (!var3.substring(var10 - -1, var10 - -4).equalsIgnoreCase(Yb[38])) {
                              if (var3.substring(var10 - -1, var10 - -4).equalsIgnoreCase(Yb[28])) {
                                 var5 = 16748608;
                                 if (!var14) {
                                    break label260;
                                 }
                              }

                              if (!var3.substring(var10 - -1, 4 + var10).equalsIgnoreCase(Yb[37])) {
                                 if (!var3.substring(1 + var10, var10 + 4).equalsIgnoreCase(Yb[23])) {
                                    if (!var3.substring(var10 - -1, 4 + var10).equalsIgnoreCase(Yb[32])) {
                                       if (var3.substring(1 + var10, 4 + var10).equalsIgnoreCase(Yb[27])) {
                                          var5 = 65535;
                                          if (!var14) {
                                             break label260;
                                          }
                                       }

                                       if (var3.substring(1 + var10, var10 + 4).equalsIgnoreCase(Yb[25])) {
                                          var5 = 16711935;
                                          if (!var14) {
                                             break label260;
                                          }
                                       }

                                       if (!var3.substring(var10 + 1, 4 + var10).equalsIgnoreCase(Yb[33])) {
                                          if (!var3.substring(var10 + 1, var10 - -4).equalsIgnoreCase(Yb[35])) {
                                             if (var3.substring(1 + var10, var10 - -4).equalsIgnoreCase(Yb[31])) {
                                                var5 = 12582912;
                                                if (!var14) {
                                                   break label260;
                                                }
                                             }

                                             if (!var3.substring(1 + var10, 4 + var10).equalsIgnoreCase(Yb[41])) {
                                                if (!var3.substring(var10 - -1, 4 + var10).equalsIgnoreCase(Yb[39])) {
                                                   if (var3.substring(1 + var10, 4 + var10).equalsIgnoreCase(Yb[24])) {
                                                      var5 = 16756736;
                                                      if (!var14) {
                                                         break label260;
                                                      }
                                                   }

                                                   if (!var3.substring(var10 + 1, 4 + var10).equalsIgnoreCase(Yb[40])) {
                                                      if (!var3.substring(var10 + 1, var10 - -4).equalsIgnoreCase(Yb[22])) {
                                                         if (var3.substring(var10 + 1, 4 + var10).equalsIgnoreCase(Yb[29])) {
                                                            var5 = 12648192;
                                                            if (!var14) {
                                                               break label260;
                                                            }
                                                         }

                                                         if (var3.substring(1 + var10, 4 + var10).equalsIgnoreCase(Yb[34])) {
                                                            var5 = 8453888;
                                                            if (!var14) {
                                                               break label260;
                                                            }
                                                         }

                                                         if (!var3.substring(var10 - -1, 4 + var10).equalsIgnoreCase(Yb[36])) {
                                                            break label260;
                                                         }

                                                         var5 = 4259584;
                                                         if (!var14) {
                                                            break label260;
                                                         }
                                                      }

                                                      var5 = 16723968;
                                                      if (!var14) {
                                                         break label260;
                                                      }
                                                   }

                                                   var5 = 16740352;
                                                   if (!var14) {
                                                      break label260;
                                                   }
                                                }

                                                var5 = (int)(1.6777215E7 * Math.random());
                                                if (!var14) {
                                                   break label260;
                                                }
                                             }

                                             var5 = 16748608;
                                             if (!var14) {
                                                break label260;
                                             }
                                          }

                                          var5 = 0;
                                          if (!var14) {
                                             break label260;
                                          }
                                       }

                                       var5 = 16777215;
                                       if (!var14) {
                                          break label260;
                                       }
                                    }

                                    var5 = 255;
                                    if (!var14) {
                                       break label260;
                                    }
                                 }

                                 var5 = 65280;
                                 if (!var14) {
                                    break label260;
                                 }
                              }

                              var5 = 16776960;
                              if (!var14) {
                                 break label260;
                              }
                           }

                           var5 = 16711680;
                        }

                        var10 += 4;
                        if (!var14) {
                           break label285;
                        }
                     }

                     if (-127 == ~var3.charAt(var10) && ~var3.length() < ~(4 + var10) && '~' == var3.charAt(4 + var10)) {
                        char var11 = var3.charAt(1 + var10);
                        char var12 = var3.charAt(2 + var10);
                        char var13 = var3.charAt(var10 + 3);
                        if (-49 >= ~var11 && ~var11 >= -58 && var12 >= '0' && '9' >= var12 && var13 >= '0' && var13 <= '9') {
                           var4 = Integer.parseInt(var3.substring(var10 + 1, var10 - -4));
                        }

                        var10 += 4;
                        if (!var14) {
                           break label285;
                        }
                     }

                     char var18 = var3.charAt(var10);
                     if (var18 == 160) {
                        var18 = ' ';
                     }

                     if (~var18 > -1 || n.a.length <= var18) {
                        var18 = ' ';
                     }

                     int var19 = n.a[var18];
                     if (this.xb && !fb.k[var7] && 0 != var5) {
                        this.a((byte)53, fb.k[var7], var17, 1 + var4, 0, var19, var2);
                     }

                     if (this.xb && !fb.k[var7] && ~var5 != -1) {
                        this.a((byte)101, fb.k[var7], var17, var4, 0, var19, var2 - -1);
                     }

                     this.a((byte)73, fb.k[var7], var17, var4, var5, var19, var2);
                     var4 += var17[var19 + 7];
                  }

                  var10++;
                  if (var14) {
                     break;
                  }
               }
            } catch (Exception var15) {
               System.out.println(Yb[30] + var15);
               var15.printStackTrace();
            }

            var10000 = w;
            var10001 = 1;
         }

         w = var10000 + var10001;
      } catch (RuntimeException var16) {
         throw i.a(var16, Yb[26] + var1 + ',' + var2 + ',' + (var3 != null ? Yb[1] : Yb[0]) + ',' + var4 + ',' + var5 + ',' + var6 + ',' + var7 + ')');
      }
   }

   @Override
   public final synchronized void removeConsumer(ImageConsumer var1) {
      try {
         if (this.fb == var1) {
            this.fb = null;
         }

         p++;
      } catch (RuntimeException var3) {
         throw i.a(var3, Yb[42] + (var1 != null ? Yb[1] : Yb[0]) + ')');
      }
   }

   private final synchronized void b(boolean var1) {
      try {
         N++;
         if (null != this.fb) {
            this.fb.setPixels(0, 0, this.u, this.k, this.nb, this.rb, 0, this.u);
            this.fb.imageComplete(2);
            if (!var1) {
               this.startProduction((ImageConsumer)null);
            }
         }
      } catch (RuntimeException var3) {
         throw i.a(var3, Yb[21] + var1 + ')');
      }
   }

   final void b(int var1, int var2, int var3, int var4, int var5) {
      boolean var8 = client.vh;

      try {
         j++;
         if (~var1 <= ~this.hb && ~this.lb < ~var1) {
            if (~this.A < ~var2) {
               var4 -= this.A + -var2;
               var2 = this.A;
            }

            if (~this.Rb > ~(var2 + var4)) {
               var4 = -var2 + this.Rb;
            }

            if (var4 > var5) {
               int var6 = var1 - -(this.u * var2);
               int var7 = 0;

               while (~var7 > ~var4) {
                  this.rb[var6 - -(this.u * var7)] = var3;
                  var7++;
                  if (var8 || var8) {
                     break;
                  }
               }
            }
         }
      } catch (RuntimeException var9) {
         throw i.a(var9, Yb[43] + var1 + ',' + var2 + ',' + var3 + ',' + var4 + ',' + var5 + ')');
      }
   }

   final void a(int var1, int var2, int var3, int var4, int var5) {
      try {
         if (this.Qb[var1]) {
            var3 += this.Sb[var1];
            var5 += this.G[var1];
         }

         sb++;
         int var6 = this.u * var5 + var3;
         int var7 = var2;
         int var8 = this.R[var1];
         int var9 = this.kb[var1];
         int var10 = this.u - var9;
         int var11 = 0;
         if (~this.A < ~var5) {
            int var12 = -var5 + this.A;
            var5 = this.A;
            var8 -= var12;
            var7 += var12 * var9;
            var6 += var12 * this.u;
         }

         if (~this.Rb >= ~(var8 + var5)) {
            var8 -= 1 + var8 + var5 + -this.Rb;
         }

         if (var3 < this.hb) {
            int var14 = this.hb + -var3;
            var10 += var14;
            var11 += var14;
            var6 += var14;
            var7 += var14;
            var3 = this.hb;
            var9 -= var14;
         }

         if (this.lb <= var3 + var9) {
            int var15 = -this.lb + (var3 - (-var9 - 1));
            var10 += var15;
            var9 -= var15;
            var11 += var15;
         }

         if (~var9 < -1 && ~var8 < -1) {
            byte var16 = 1;
            if (this.i) {
               var11 += this.kb[var1];
               var16 = 2;
               if ((1 & var5) != 0) {
                  var8--;
                  var6 += this.u;
               }

               var10 += this.u;
            }

            if (this.ob[var1] == null) {
               this.a(var7, var10, this.gb[var1], var16, false, var8, var11, var4, var9, this.rb, this.Y[var1], var6);
               if (!client.vh) {
                  return;
               }
            }

            this.a(var7, var8, var6, this.ob[var1], 0, var4, var16, this.rb, -107, var11, var10, var9);
         }
      } catch (RuntimeException var13) {
         throw i.a(var13, Yb[19] + var1 + ',' + var2 + ',' + var3 + ',' + var4 + ',' + var5 + ')');
      }
   }

   final void a(Graphics var1, int var2, int var3, int var4) {
      try {
         e++;
         this.b(true);
         var1.drawImage(this.Gb, var2, var4, this);
         if (var3 != 256) {
            Kb = (String[])null;
         }
      } catch (RuntimeException var6) {
         throw i.a(var6, Yb[49] + (var1 != null ? Yb[1] : Yb[0]) + ',' + var2 + ',' + var3 + ',' + var4 + ')');
      }
   }

   final void a(int var1, byte var2, int var3, int var4, int var5, int var6, int var7) {
      try {
         I++;

         try {
            int var8 = this.kb[var1];
            int var9 = this.R[var1];
            int var10 = 0;
            int var11 = 0;
            int var12 = (var8 << 2133607216) / var5;
            int var13 = (var9 << -1116806288) / var3;
            if (this.Qb[var1]) {
               int var14 = this.Eb[var1];
               int var15 = this.qb[var1];
               if (var14 == 0 || -1 == ~var15) {
                  return;
               }

               var13 = (var15 << 1837802000) / var3;
               var6 += (var15 + var3 * this.G[var1] + -1) / var15;
               var4 += (var14 + this.Sb[var1] * var5 + -1) / var14;
               var12 = (var14 << -405443120) / var5;
               if (-1 != ~(var5 * this.Sb[var1] % var14)) {
                  var10 = (-(this.Sb[var1] * var5 % var14) + var14 << 1444777936) / var5;
               }

               if (-1 != ~(this.G[var1] * var3 % var15)) {
                  var11 = (var15 + -(this.G[var1] * var3 % var15) << -1176347888) / var3;
               }

               var5 = var5 * (this.kb[var1] - (var10 >> 7993200)) / var14;
               var3 = (-(var11 >> 826090000) + this.R[var1]) * var3 / var15;
            }

            int var19 = var6 * this.u + var4;
            if (var2 > -121) {
               return;
            }

            if (~var6 > ~this.A) {
               int var16 = -var6 + this.A;
               var3 -= var16;
               var6 = 0;
               var19 += this.u * var16;
               var11 += var13 * var16;
            }

            int var20 = this.u + -var5;
            if (~this.hb < ~var4) {
               int var21 = -var4 + this.hb;
               var4 = 0;
               var10 += var21 * var12;
               var19 += var21;
               var5 -= var21;
               var20 += var21;
            }

            if (~this.Rb >= ~(var6 + var3)) {
               var3 -= 1 + var3 + var6 + -this.Rb;
            }

            if (~this.lb >= ~(var4 + var5)) {
               int var22 = 1 + var4 + (var5 - this.lb);
               var20 += var22;
               var5 -= var22;
            }

            byte var23 = 1;
            if (this.i) {
               var13 += var13;
               var20 += this.u;
               if (0 != (var6 & 1)) {
                  var19 += this.u;
                  var3--;
               }

               var23 = 2;
            }

            this.a(var23, var11, var5, (byte)-61, var13, var8, var12, var3, var19, this.ob[var1], 0, var10, var20, var7, this.rb);
         } catch (Exception var17) {
            System.out.println(Yb[16]);
         }
      } catch (RuntimeException var18) {
         throw i.a(var18, Yb[54] + var1 + ',' + var2 + ',' + var3 + ',' + var4 + ',' + var5 + ',' + var6 + ',' + var7 + ')');
      }
   }

   // $VF: Irreducible bytecode was duplicated to produce valid code
   private final void a(int var1, int[] var2, int var3, byte var4, int var5, int var6, int var7, int var8, byte[] var9, int var10) {
      boolean var14 = client.vh;

      try {
         g++;

         try {
            int var11 = -(var7 >> -1201305182);
            var7 = -(var7 & 3);
            int var12 = -var6;

            while (~var12 > -1) {
               if (var14) {
                  return;
               }

               int var13 = var11;

               int var17;
               label130:
               while (true) {
                  byte var10001;
                  if (~var13 > -1) {
                     var17 = 0;
                     var10001 = var9[var8++];
                     if (!var14) {
                        if (0 != var10001) {
                           var2[var3++] = var1;
                           if (var14) {
                              var3++;
                           }
                        } else {
                           var3++;
                        }

                        label113: {
                           if (~var9[var8++] == -1) {
                              var3++;
                              if (!var14) {
                                 break label113;
                              }
                           }

                           var2[var3++] = var1;
                        }

                        label108: {
                           if (0 == var9[var8++]) {
                              var3++;
                              if (!var14) {
                                 break label108;
                              }
                           }

                           var2[var3++] = var1;
                        }

                        if (~var9[var8++] != -1) {
                           var2[var3++] = var1;
                           if (var14) {
                              var3++;
                           }
                        } else {
                           var3++;
                        }

                        var13++;
                        if (!var14) {
                           continue;
                        }

                        var13 = var7;
                        var17 = ~var13;
                        var10001 = -1;
                     }
                  } else {
                     var13 = var7;
                     var17 = ~var13;
                     var10001 = -1;
                  }

                  while (var17 > var10001) {
                     var17 = var9[var8++];
                     if (var14) {
                        break label130;
                     }

                     if (var17 == 0) {
                        var3++;
                        if (var14) {
                           var2[var3++] = var1;
                        }
                     } else {
                        var2[var3++] = var1;
                     }

                     var13++;
                     if (var14) {
                        var8 += var10;
                        var17 = var3 + var5;
                        break label130;
                     }

                     var17 = ~var13;
                     var10001 = -1;
                  }

                  var8 += var10;
                  var17 = var3 + var5;
                  break;
               }

               var3 = var17;
               var12++;
               if (var14) {
                  break;
               }
            }
         } catch (Exception var15) {
            System.out.println(Yb[14] + var15);
            var15.printStackTrace();
         }

         int var18 = 82 % ((-45 - var4) / 48);
      } catch (RuntimeException var16) {
         throw i.a(
            var16,
            Yb[13]
               + var1
               + ','
               + (var2 != null ? Yb[1] : Yb[0])
               + ','
               + var3
               + ','
               + var4
               + ','
               + var5
               + ','
               + var6
               + ','
               + var7
               + ','
               + var8
               + ','
               + (var9 != null ? Yb[1] : Yb[0])
               + ','
               + var10
               + ')'
         );
      }
   }

   // $VF: Irreducible bytecode was duplicated to produce valid code
   private final void a(int var1, int[] var2, int var3, int var4, int[] var5, int var6, int var7, boolean var8, byte[] var9, int var10, int var11) {
      boolean var16 = client.vh;

      try {
         if (!var8) {
            this.i = true;
         }

         Ub++;
         int var12 = -(var10 >> -1243049534);
         var10 = -(3 & var10);
         int var13 = -var7;

         while (~var13 > -1 && !var16) {
            int var14 = var12;

            int var22;
            int var32;
            label143: {
               while (true) {
                  if (0 > var14) {
                     byte var15 = var9[var3++];
                     var22 = ~var15;
                     var32 = -1;
                     if (var16) {
                        break;
                     }

                     label121: {
                        if (var22 == -1) {
                           var1++;
                           if (!var16) {
                              break label121;
                           }
                        }

                        var5[var1++] = var2[ib.a(var15, 255)];
                     }

                     label116: {
                        var15 = var9[var3++];
                        if (var15 == 0) {
                           var1++;
                           if (!var16) {
                              break label116;
                           }
                        }

                        var5[var1++] = var2[ib.a(var15, 255)];
                     }

                     var15 = var9[var3++];
                     if (~var15 != -1) {
                        var5[var1++] = var2[ib.a(255, var15)];
                        if (var16) {
                           var1++;
                        }
                     } else {
                        var1++;
                     }

                     var15 = var9[var3++];
                     if (var15 != 0) {
                        var5[var1++] = var2[ib.a(var15, 255)];
                        if (var16) {
                           var1++;
                        }
                     } else {
                        var1++;
                     }

                     var14++;
                     if (!var16) {
                        continue;
                     }

                     var14 = var10;
                  } else {
                     var14 = var10;
                  }

                  var22 = 0;
                  var32 = var14;
                  break;
               }

               while (var22 > var32) {
                  byte var21 = var9[var3++];
                  var22 = 0;
                  var32 = var21;
                  if (var16) {
                     break label143;
                  }

                  if (0 != var21) {
                     var5[var1++] = var2[ib.a(var21, 255)];
                     if (var16) {
                        var1++;
                     }
                  } else {
                     var1++;
                  }

                  var14++;
                  if (var16) {
                     var1 += var11;
                     var3 += var4;
                     var22 = var13;
                     var32 = var6;
                     break label143;
                  }

                  var22 = 0;
                  var32 = var14;
               }

               var1 += var11;
               var3 += var4;
               var22 = var13;
               var32 = var6;
            }

            var13 = var22 + var32;
            if (var16) {
               break;
            }
         }
      } catch (RuntimeException var17) {
         throw i.a(
            var17,
            Yb[82]
               + var1
               + ','
               + (var2 != null ? Yb[1] : Yb[0])
               + ','
               + var3
               + ','
               + var4
               + ','
               + (var5 != null ? Yb[1] : Yb[0])
               + ','
               + var6
               + ','
               + var7
               + ','
               + var8
               + ','
               + (var9 != null ? Yb[1] : Yb[0])
               + ','
               + var10
               + ','
               + var11
               + ')'
         );
      }
   }

   final void b(int var1, int var2) {
      boolean var9 = client.vh;

      try {
         b++;
         if (null != this.gb[var1]) {
            int var3 = this.kb[var1] * this.R[var1];
            byte[] var4 = this.gb[var1];
            int[] var5 = this.Y[var1];
            int[] var6 = new int[var3];
            int var7 = 0;

            int var10000;
            while (true) {
               if (~var7 > ~var3) {
                  int var8 = var5[0xFF & var4[var7]];
                  var10000 = var8;
                  if (var9) {
                     break;
                  }

                  label41: {
                     if (var8 == 0) {
                        var8 = 1;
                        if (!var9) {
                           break label41;
                        }
                     }

                     if (-16711936 == ~var8) {
                        var8 = 0;
                     }
                  }

                  var6[var7] = var8;
                  var7++;
                  if (!var9) {
                     continue;
                  }
               }

               var10000 = var2;
               break;
            }

            if (var10000 != -342059728) {
               this.Eb = (int[])null;
            }

            this.ob[var1] = var6;
            this.gb[var1] = null;
            this.Y[var1] = null;
         }
      } catch (RuntimeException var10) {
         throw i.a(var10, Yb[77] + var1 + ',' + var2 + ')');
      }
   }

   final void a(int var1, int var2, int var3, int var4, int var5, int var6) {
      boolean var48 = client.vh;

      try {
         int var7;
         int var8;
         int var9;
         label260: {
            S++;
            var7 = this.u;
            var8 = this.k;
            if (this.Hb == null) {
               this.Hb = new int[512];
               var9 = 0;

               while (var9 < 256) {
                  this.Hb[var9] = (int)(Math.sin(var9 * 0.02454369) * 32768.0);
                  this.Hb[256 + var9] = (int)(Math.cos(var9 * 0.02454369) * 32768.0);
                  var9++;
                  if (var48) {
                     break label260;
                  }

                  if (var48) {
                     break;
                  }
               }
            }

            var9 = -this.Eb[var1] / 2;
         }

         int var10 = -this.qb[var1] / 2;
         if (this.Qb[var1]) {
            var9 += this.Sb[var1];
            var10 += this.G[var1];
         }

         int var19;
         int var20;
         int var21;
         int var22;
         int var23;
         int var24;
         int var25;
         int var26;
         label422: {
            int var11 = this.kb[var1] + var9;
            int var12 = this.R[var1] + var10;
            int var13 = var11;
            int var14 = var10;
            int var15 = var9;
            var6 &= 255;
            int var16 = var12;
            int var17 = this.Hb[var6] * var5;
            int var18 = this.Hb[var6 + 256] * var5;
            var19 = var3 + (var18 * var9 + var10 * var17 >> 2041339190);
            var20 = var2 - -(-(var9 * var17) + var10 * var18 >> -1101614986);
            var21 = (var13 * var18 + var14 * var17 >> -1412824714) + var3;
            var22 = var2 + (var18 * var14 - var13 * var17 >> -1277097834);
            var23 = (var17 * var12 + var18 * var11 >> 1894119030) + var3;
            var24 = (-(var17 * var11) + var18 * var12 >> 195973654) + var2;
            var25 = (var15 * var18 + var17 * var16 >> -352508522) + var3;
            var26 = var2 - -(var16 * var18 - var15 * var17 >> -223287050);
            if (~var5 != -193 || ~(63 & client.ef) != ~(63 & var6)) {
               if (~var5 == -129) {
                  client.ef = var6;
                  if (!var48) {
                     break label422;
                  }
               }

               da.M++;
               if (!var48) {
                  break label422;
               }
            }

            nb.g++;
         }

         int var27;
         int var28;
         label276: {
            var27 = var20;
            var28 = var20;
            if (var27 > var22) {
               var27 = var22;
               if (!var48) {
                  break label276;
               }
            }

            if (~var28 > ~var22) {
               var28 = var22;
            }
         }

         label281: {
            if (~var24 > ~var27) {
               var27 = var24;
               if (!var48) {
                  break label281;
               }
            }

            if (var24 > var28) {
               var28 = var24;
            }
         }

         label286: {
            if (~var26 > ~var27) {
               var27 = var26;
               if (!var48) {
                  break label286;
               }
            }

            if (~var28 > ~var26) {
               var28 = var26;
            }
         }

         if (~var27 > ~this.A) {
            var27 = this.A;
         }

         if (this.Xb == null || ~this.Xb.length != ~(var8 + 1)) {
            this.tb = new int[var8 + 1];
            this.M = new int[1 + var8];
            this.t = new int[var8 + 1];
            this.Tb = new int[1 + var8];
            this.Wb = new int[1 + var8];
            this.Xb = new int[var8 - -1];
         }

         if (~this.Rb > ~var28) {
            var28 = this.Rb;
         }

         int var29 = var27;

         while (~var28 <= ~var29) {
            this.Xb[var29] = 99999999;
            this.t[var29] = -99999999;
            var29++;
            if (var48) {
               break;
            }
         }

         int var30;
         int var31;
         int var32;
         int var34;
         int var35;
         int var36;
         int var37;
         int var52;
         int var53;
         int var54;
         byte var55;
         byte var56;
         int var57;
         label304: {
            var32 = 0;
            var34 = 0;
            var36 = 0;
            var37 = this.kb[var1];
            var52 = -1 + var37;
            int var38 = this.R[var1];
            var54 = var37 + -1;
            var50 = 0;
            var51 = 0;
            var55 = 0;
            var56 = 0;
            var53 = var38 + -1;
            var57 = -1 + var38;
            if (var20 > var26) {
               var30 = var20;
               var35 = var57 << 840930536;
               var29 = var26;
               var31 = var25 << -731169944;
               if (!var48) {
                  break label304;
               }
            }

            var30 = var26;
            var29 = var20;
            var35 = var51 << 813501160;
            var31 = var19 << 1063907752;
         }

         if (~var26 != ~var20) {
            var36 = (-var51 + var57 << 822762216) / (var26 + -var20);
            var32 = (-var19 + var25 << -1960375352) / (var26 - var20);
         }

         if (-1 < ~var29) {
            var31 -= var32 * var29;
            var35 -= var29 * var36;
            var29 = 0;
         }

         if (~var30 < ~(-1 + var8)) {
            var30 = var8 + -1;
         }

         if (var4 != 842218000) {
            return;
         }

         int var39 = var29;

         label315: {
            while (~var39 >= ~var30) {
               this.Xb[var39] = this.t[var39] = var31;
               var31 += var32;
               int[] var40 = this.M;
               this.Tb[var39] = 0;
               int var41 = var39;
               var40[var41] = 0;
               this.tb[var39] = this.Wb[var39] = var35;
               var35 += var36;
               var39++;
               if (var48) {
                  break label315;
               }

               if (var48) {
                  break;
               }
            }

            if (var20 != var22) {
               var32 = (var21 + -var19 << -1951284024) / (-var20 + var22);
               var34 = (-var50 + var54 << 1234137512) / (-var20 + var22);
            }
         }

         int var33;
         label320: {
            if (var22 < var20) {
               var33 = var54 << -2081746392;
               var30 = var20;
               var29 = var22;
               var31 = var21 << 1385760008;
               if (!var48) {
                  break label320;
               }
            }

            var31 = var19 << 743454440;
            var29 = var20;
            var33 = var50 << 705115528;
            var30 = var22;
         }

         if (var8 + -1 < var30) {
            var30 = var8 - 1;
         }

         if (0 > var29) {
            var33 -= var34 * var29;
            var31 -= var32 * var29;
            var29 = 0;
         }

         var39 = var29;

         int var10000;
         int var10001;
         label334: {
            while (var30 >= var39) {
               var10000 = ~var31;
               var10001 = ~this.Xb[var39];
               if (var48) {
                  break label334;
               }

               if (var10000 > var10001) {
                  this.Xb[var39] = var31;
                  this.M[var39] = var33;
                  this.tb[var39] = 0;
               }

               if (~this.t[var39] > ~var31) {
                  this.t[var39] = var31;
                  this.Tb[var39] = var33;
                  this.Wb[var39] = 0;
               }

               var31 += var32;
               var33 += var34;
               var39++;
               if (var48) {
                  break;
               }
            }

            var10000 = ~var22;
            var10001 = ~var24;
         }

         label339: {
            if (var10000 >= var10001) {
               var33 = var54 << -1289051512;
               var35 = var55 << -1525813880;
               var29 = var22;
               var31 = var21 << -952809592;
               var30 = var24;
               if (!var48) {
                  break label339;
               }
            }

            var33 = var52 << -794260248;
            var30 = var22;
            var35 = var53 << -606906680;
            var31 = var23 << -1631775192;
            var29 = var24;
         }

         if (var22 != var24) {
            var36 = (-var55 + var53 << -665843768) / (-var22 + var24);
            var32 = (-var21 + var23 << -740697688) / (var24 + -var22);
         }

         if (var30 > var8 + -1) {
            var30 = var8 + -1;
         }

         if (~var29 > -1) {
            var31 -= var32 * var29;
            var35 -= var36 * var29;
            var29 = 0;
         }

         var39 = var29;

         label353: {
            while (~var30 <= ~var39) {
               var10000 = ~var31;
               var10001 = ~this.Xb[var39];
               if (var48) {
                  break label353;
               }

               if (var10000 > var10001) {
                  this.Xb[var39] = var31;
                  this.M[var39] = var33;
                  this.tb[var39] = var35;
               }

               if (var31 > this.t[var39]) {
                  this.t[var39] = var31;
                  this.Tb[var39] = var33;
                  this.Wb[var39] = var35;
               }

               var31 += var32;
               var35 += var36;
               var39++;
               if (var48) {
                  break;
               }
            }

            var10000 = ~var24;
            var10001 = ~var26;
         }

         if (var10000 != var10001) {
            var32 = (var25 + -var23 << -1317172024) / (var26 + -var24);
            var34 = (var56 + -var52 << -236389848) / (-var24 + var26);
         }

         label359: {
            if (~var24 < ~var26) {
               var33 = var56 << 45836488;
               var30 = var24;
               var29 = var26;
               var35 = var57 << 1400553864;
               var31 = var25 << 2055980168;
               if (!var48) {
                  break label359;
               }
            }

            var29 = var24;
            var35 = var53 << 222482728;
            var31 = var23 << -523743608;
            var33 = var52 << -171150040;
            var30 = var26;
         }

         if (var29 < 0) {
            var31 -= var29 * var32;
            var33 -= var29 * var34;
            var29 = 0;
         }

         if (var8 + -1 < var30) {
            var30 = -1 + var8;
         }

         var39 = var29;

         label373: {
            while (var30 >= var39) {
               var10000 = ~this.Xb[var39];
               var10001 = ~var31;
               if (var48) {
                  break label373;
               }

               if (var10000 < var10001) {
                  this.Xb[var39] = var31;
                  this.M[var39] = var33;
                  this.tb[var39] = var35;
               }

               if (~this.t[var39] > ~var31) {
                  this.t[var39] = var31;
                  this.Tb[var39] = var33;
                  this.Wb[var39] = var35;
               }

               var31 += var32;
               var33 += var34;
               var39++;
               if (var48) {
                  break;
               }
            }

            var10000 = var27;
            var10001 = var7;
         }

         var39 = var10000 * var10001;
         int[] var76 = this.ob[var1];
         int var77 = var27;

         while (~var28 < ~var77) {
            int var42 = this.Xb[var77] >> -299165016;
            int var43 = this.t[var77] >> -1707577976;
            if (var48) {
               break;
            }

            label419: {
               if (~(var43 + -var42) >= -1) {
                  var39 += var7;
                  if (!var48) {
                     break label419;
                  }
               }

               int var44 = this.M[var77] << -122904183;
               int var45 = ((this.Tb[var77] << 1728443241) - var44) / (var43 - var42);
               int var46 = this.tb[var77] << 1699710857;
               int var47 = (-var46 + (this.Wb[var77] << -470306615)) / (-var42 + var43);
               if (~this.lb > ~var43) {
                  var43 = this.lb;
               }

               if (var42 < this.hb) {
                  var46 += (-var42 + this.hb) * var47;
                  var44 += (this.hb + -var42) * var45;
                  var42 = this.hb;
               }

               label387:
               if (!this.i || ~(var77 & 1) == -1) {
                  if (this.Qb[var1]) {
                     this.a(var47, -var43 + var42, var44, var76, this.rb, var46, var42 + var39, var37, var45, 0, (byte)102);
                     if (!var48) {
                        break label387;
                     }
                  }

                  this.a(-var43 + var42, var37, 0, this.rb, var45, var46, var44, var76, var47, var42 + var39, true);
               }

               var39 += var7;
            }

            var77++;
            if (var48) {
               break;
            }
         }
      } catch (RuntimeException var49) {
         throw i.a(var49, Yb[79] + var1 + ',' + var2 + ',' + var3 + ',' + var4 + ',' + var5 + ',' + var6 + ')');
      }

      if (e.Ab != 0) {
         client.vh = !var48;
      }
   }

   final void d(int var1, int var2, int var3, int var4, int var5, int var6) {
      boolean var11 = client.vh;

      try {
         this.kb[var1] = var4;
         Jb++;
         this.R[var1] = var2;
         if (var3 > 108) {
            this.Qb[var1] = false;
            this.Sb[var1] = 0;
            this.G[var1] = 0;
            this.Eb[var1] = var4;
            this.qb[var1] = var2;
            int var7 = var4 * var2;
            this.ob[var1] = new int[var7];
            int var8 = 0;
            int var9 = var5;

            while (~(var2 + var5) < ~var9 && !var11) {
               int var10 = var6;

               while (true) {
                  if (var10 < var4 + var6) {
                     this.ob[var1][var8++] = this.rb[this.u * var9 + var10];
                     var10++;
                     if (var11) {
                        break;
                     }

                     if (!var11) {
                        continue;
                     }
                  }

                  var9++;
                  break;
               }

               if (var11) {
                  break;
               }
            }
         }
      } catch (RuntimeException var12) {
         throw i.a(var12, Yb[53] + var1 + ',' + var2 + ',' + var3 + ',' + var4 + ',' + var5 + ',' + var6 + ')');
      }
   }

   private final void a(
      int var1,
      int var2,
      int var3,
      int var4,
      int var5,
      int[] var6,
      int var7,
      int[] var8,
      int var9,
      int var10,
      boolean var11,
      int var12,
      int var13,
      int var14,
      int var15
   ) {
      boolean var26 = client.vh;

      try {
         vb++;
         int var16 = (var14 & 16736117) >> -846694704;
         int var17 = 0xFF & var14 >> -1400710808;
         int var18 = 0xFF & var14;

         try {
            int var19 = var4;
            if (var11) {
               this.Tb = (int[])null;
            }

            int var20 = -var13;

            while (~var20 > -1) {
               int var21 = (var1 >> -324011088) * var10;
               if (var26) {
                  break;
               }

               int var22 = -var3;

               int var10000;
               int var10001;
               label84: {
                  while (-1 < ~var22) {
                     var9 = var6[var21 + (var4 >> 1856265008)];
                     var10000 = ~var9;
                     var10001 = -1;
                     if (var26) {
                        break label84;
                     }

                     label80: {
                        if (var10000 != -1) {
                           label102: {
                              int var23 = var9 >> -1835533520 & 0xFF;
                              int var24 = (65376 & var9) >> 1180773672;
                              int var25 = 0xFF & var9;
                              if (~var24 != ~var23 || ~var24 != ~var25) {
                                 var8[var7++] = var9;
                                 if (!var26) {
                                    break label102;
                                 }
                              }

                              var8[var7++] = (var23 * var16 >> 1799197256 << -273346448)
                                 + (var17 * var24 >> -1796300920 << 1450486152)
                                 + (var25 * var18 >> 835792776);
                           }

                           if (!var26) {
                              break label80;
                           }
                        }

                        var7++;
                     }

                     var4 += var2;
                     var22++;
                     if (var26) {
                        break;
                     }
                  }

                  var1 += var12;
                  var7 += var5;
                  var4 = var19;
                  var10000 = var20;
                  var10001 = var15;
               }

               var20 = var10000 + var10001;
               if (var26) {
                  break;
               }
            }
         } catch (Exception var27) {
            System.out.println(Yb[47]);
         }
      } catch (RuntimeException var28) {
         throw i.a(
            var28,
            Yb[52]
               + var1
               + ','
               + var2
               + ','
               + var3
               + ','
               + var4
               + ','
               + var5
               + ','
               + (var6 != null ? Yb[1] : Yb[0])
               + ','
               + var7
               + ','
               + (var8 != null ? Yb[1] : Yb[0])
               + ','
               + var9
               + ','
               + var10
               + ','
               + var11
               + ','
               + var12
               + ','
               + var13
               + ','
               + var14
               + ','
               + var15
               + ')'
         );
      }
   }

   final void d(int var1, int var2) {
      try {
         this.D = var2;
         int var3 = -54 / ((-63 - var1) / 55);
         bb++;
      } catch (RuntimeException var4) {
         throw i.a(var4, Yb[87] + var1 + ',' + var2 + ')');
      }
   }

   private final void a(int var1, int var2, int var3, int var4, String var5, int var6, int var7) {
      try {
         if (var1 == 11815) {
            this.a(var4, var7, var5, var6 - this.a(var3, 92, var5) / 2, var2, (byte)-124, var3);
            zb++;
         }
      } catch (RuntimeException var9) {
         throw i.a(var9, Yb[9] + var1 + ',' + var2 + ',' + var3 + ',' + var4 + ',' + (var5 != null ? Yb[1] : Yb[0]) + ',' + var6 + ',' + var7 + ')');
      }
   }

   private final void a(
      int var1,
      int var2,
      int var3,
      byte var4,
      int var5,
      int var6,
      int var7,
      int var8,
      int var9,
      int[] var10,
      int var11,
      int var12,
      int var13,
      int var14,
      int[] var15
   ) {
      boolean var22 = client.vh;

      try {
         d++;
         int var16 = 256 - var14;

         try {
            int var17 = var12;
            if (var4 != -61) {
               return;
            }

            int var18 = -var8;

            while (~var18 > -1) {
               int var19 = var6 * (var2 >> -330929776);
               var2 += var5;
               if (var22) {
                  break;
               }

               int var20 = -var3;

               int var10000;
               int var10001;
               label69: {
                  while (-1 < ~var20) {
                     var11 = var10[var19 + (var12 >> -483039408)];
                     var12 += var7;
                     var10000 = ~var11;
                     var10001 = -1;
                     if (var22) {
                        break label69;
                     }

                     label65: {
                        if (var10000 == -1) {
                           var9++;
                           if (!var22) {
                              break label65;
                           }
                        }

                        int var21 = var15[var9];
                        var15[var9++] = ib.a(ib.a(65280, var21) * var16 + ib.a(65280, var11) * var14, 16711680)
                              + ib.a(ib.a(var11, 16711935) * var14 + var16 * ib.a(16711935, var21), -16711936)
                           >> -130221816;
                     }

                     var20++;
                     if (var22) {
                        break;
                     }
                  }

                  var9 += var13;
                  var12 = var17;
                  var10000 = var18;
                  var10001 = var1;
               }

               var18 = var10000 + var10001;
               if (var22) {
                  break;
               }
            }
         } catch (Exception var23) {
            System.out.println(Yb[45]);
         }
      } catch (RuntimeException var24) {
         throw i.a(
            var24,
            Yb[46]
               + var1
               + ','
               + var2
               + ','
               + var3
               + ','
               + var4
               + ','
               + var5
               + ','
               + var6
               + ','
               + var7
               + ','
               + var8
               + ','
               + var9
               + ','
               + (var10 != null ? Yb[1] : Yb[0])
               + ','
               + var11
               + ','
               + var12
               + ','
               + var13
               + ','
               + var14
               + ','
               + (var15 != null ? Yb[1] : Yb[0])
               + ')'
         );
      }
   }

   private final void a(
      int var1,
      int[] var2,
      int var3,
      int var4,
      int var5,
      int var6,
      int var7,
      int[] var8,
      int var9,
      int var10,
      int var11,
      int var12,
      int var13,
      int var14,
      int var15,
      int var16
   ) {
      boolean var29 = client.vh;

      try {
         L++;
         int var20 = (16717586 & var16) >> -1742839120;
         int var21 = (var16 & 65518) >> -734006424;
         int var22 = 0xFF & var16;

         int var10000;
         int var10001;
         label102: {
            try {
               int var23 = var6;
               int var24 = -var9;

               while (-1 < ~var24) {
                  int var25 = (var5 >> 1187405840) * var14;
                  int var26 = var13 >> -1195642352;
                  int var27 = var7;
                  var10000 = var26;
                  var10001 = this.hb;
                  if (var29) {
                     break label102;
                  }

                  if (var26 < this.hb) {
                     int var28 = this.hb + -var26;
                     var26 = this.hb;
                     var27 -= var28;
                     var6 += var12 * var28;
                  }

                  var15 = -var15 + 1;
                  if (~this.lb >= ~(var26 - -var27)) {
                     int var32 = var27 + var26 + -this.lb;
                     var27 -= var32;
                  }

                  label93: {
                     if (0 != var15) {
                        int var33 = var26;

                        while (var33 < var27 + var26) {
                           var4 = var2[var25 + (var6 >> -1996606992)];
                           var10000 = var4;
                           if (var29) {
                              break label93;
                           }

                           label87:
                           if (var4 != 0) {
                              int var19 = var4 & 0xFF;
                              int var17 = (var4 & 16769815) >> 1817608464;
                              int var18 = 0xFF & var4 >> -7992056;
                              if (~var18 == ~var17 && var19 == var18) {
                                 var8[var11 + var33] = (var19 * var22 >> -1923074072)
                                    + (var21 * var18 >> -1585910200 << -1586887384)
                                    + (var20 * var17 >> 2107059944 << -1441257008);
                                 if (!var29) {
                                    break label87;
                                 }
                              }

                              var8[var33 + var11] = var4;
                           }

                           var6 += var12;
                           var33++;
                           if (var29) {
                              break;
                           }
                        }
                     }

                     var5 += var3;
                     var6 = var23;
                     var13 += var10;
                     var10000 = var11 + this.u;
                  }

                  var11 = var10000;
                  var24++;
                  if (var29) {
                     break;
                  }
               }
            } catch (Exception var30) {
               System.out.println(Yb[65]);
            }

            var10000 = var1;
            var10001 = 20;
         }

         if (var10000 < var10001) {
            this.t = (int[])null;
         }
      } catch (RuntimeException var31) {
         throw i.a(
            var31,
            Yb[89]
               + var1
               + ','
               + (var2 != null ? Yb[1] : Yb[0])
               + ','
               + var3
               + ','
               + var4
               + ','
               + var5
               + ','
               + var6
               + ','
               + var7
               + ','
               + (var8 != null ? Yb[1] : Yb[0])
               + ','
               + var9
               + ','
               + var10
               + ','
               + var11
               + ','
               + var12
               + ','
               + var13
               + ','
               + var14
               + ','
               + var15
               + ','
               + var16
               + ')'
         );
      }
   }

   final int a(int var1, int var2) {
      try {
         if (var1 != 508305352) {
            this.G = (int[])null;
         }

         Pb++;
         if (var2 != 0) {
            if (var2 != 1) {
               if (var2 == 2) {
                  return 14;
               }

               if (-4 == ~var2) {
                  return 15;
               }

               if (-5 != ~var2) {
                  if (~var2 != -6) {
                     if (6 == var2) {
                        return 24;
                     } else {
                        return 7 != var2 ? this.c(60, var2) : 29;
                     }
                  } else {
                     return 19;
                  }
               } else {
                  return 15;
               }
            } else {
               return 14;
            }
         } else {
            return 12;
         }
      } catch (RuntimeException var4) {
         throw i.a(var4, Yb[74] + var1 + 44 + var2 + 41);
      }
   }

   private final void a(
      byte[] var1,
      int var2,
      int var3,
      int var4,
      int[] var5,
      int var6,
      int var7,
      int var8,
      int var9,
      int var10,
      int[] var11,
      int var12,
      int var13,
      int var14,
      int var15,
      int var16,
      int var17
   ) {
      boolean var31 = client.vh;

      try {
         m++;
         int var21 = (16721450 & var2) >> 1137456112;
         int var22 = (65437 & var2) >> -1550693080;
         int var23 = 82 % ((var13 - -52) / 32);
         int var24 = var2 & 0xFF;

         try {
            int var25 = var14;
            int var26 = -var17;

            while (var26 < 0) {
               int var27 = (var7 >> 160375440) * var6;
               int var28 = var9 >> 698170864;
               int var29 = var16;
               if (var31) {
                  break;
               }

               if (~this.hb < ~var28) {
                  int var30 = this.hb + -var28;
                  var28 = this.hb;
                  var29 -= var30;
                  var14 += var30 * var10;
               }

               if (~this.lb >= ~(var29 + var28)) {
                  int var34 = var28 + (var29 - this.lb);
                  var29 -= var34;
               }

               int var10000;
               int var10001;
               label97: {
                  var12 = 1 + -var12;
                  var7 += var8;
                  if (-1 != ~var12) {
                     int var35 = var28;

                     while (~(var29 + var28) < ~var35) {
                        var3 = var1[(var14 >> -1390664752) - -var27] & 255;
                        var10000 = -1;
                        var10001 = ~var3;
                        if (var31) {
                           break label97;
                        }

                        label91:
                        if (-1 != var10001) {
                           var3 = var5[var3];
                           int var20 = var3 & 0xFF;
                           int var18 = var3 >> -692748912 & 0xFF;
                           int var19 = var3 >> 1006963688 & 0xFF;
                           if (~var19 != ~var18 || ~var19 != ~var20) {
                              var11[var15 + var35] = var3;
                              if (!var31) {
                                 break label91;
                              }
                           }

                           var11[var35 - -var15] = (var22 * var19 >> 436099816 << -1418519128)
                              + (var21 * var18 >> -2076941880 << 547107024)
                              - -(var24 * var20 >> -2023878008);
                        }

                        var14 += var10;
                        var35++;
                        if (var31) {
                           break;
                        }
                     }
                  }

                  var9 += var4;
                  var10000 = var15;
                  var10001 = this.u;
               }

               var15 = var10000 + var10001;
               var14 = var25;
               var26++;
               if (var31) {
                  break;
               }
            }
         } catch (Exception var32) {
            System.out.println(Yb[65]);
         }
      } catch (RuntimeException var33) {
         throw i.a(
            var33,
            Yb[67]
               + (var1 != null ? Yb[1] : Yb[0])
               + ','
               + var2
               + ','
               + var3
               + ','
               + var4
               + ','
               + (var5 != null ? Yb[1] : Yb[0])
               + ','
               + var6
               + ','
               + var7
               + ','
               + var8
               + ','
               + var9
               + ','
               + var10
               + ','
               + (var11 != null ? Yb[1] : Yb[0])
               + ','
               + var12
               + ','
               + var13
               + ','
               + var14
               + ','
               + var15
               + ','
               + var16
               + ','
               + var17
               + ')'
         );
      }
   }

   private final void a(int var1, int var2, String var3, int var4, int var5, int var6, int var7) {
      try {
         this.a(var7, var2, var3, var1 - this.a(var6, 114, var3), var4, (byte)123, var6);
         H++;
         if (var5 != -12200) {
            this.b(75, -128, -127, 3, 49, -124);
         }
      } catch (RuntimeException var9) {
         throw i.a(var9, Yb[5] + var1 + ',' + var2 + ',' + (var3 != null ? Yb[1] : Yb[0]) + ',' + var4 + ',' + var5 + ',' + var6 + ',' + var7 + ')');
      }
   }

   private final void a(byte[] var1, int var2, int var3, int var4, int var5, int var6, int var7, int var8, int[] var9, int var10) {
      boolean var15 = client.vh;

      try {
         Fb++;
         if (var7 != 1504725224) {
            this.rb = (int[])null;
         }

         int var11 = -var5;

         while (-1 < ~var11 && !var15) {
            int var12 = -var3;

            int var10000;
            int var10001;
            while (true) {
               if (var12 < 0) {
                  int var13 = 255 & var1[var10++];
                  var10000 = -31;
                  var10001 = ~var13;
                  if (var15) {
                     break;
                  }

                  label82: {
                     if (-31 <= var10001) {
                        var4++;
                        if (!var15) {
                           break label82;
                        }
                     }

                     if (-231 < ~var13) {
                        int var14 = var9[var4];
                        var9[var4++] = ib.a(-16711936, ib.a(16711935, var2) * var13 - -(ib.a(var14, 16711935) * (-var13 + 256)))
                              - -ib.a((256 - var13) * ib.a(65280, var14) + var13 * ib.a(65280, var2), 16711680)
                           >> 1504725224;
                        if (!var15) {
                           break label82;
                        }
                     }

                     var9[var4++] = var2;
                  }

                  var12++;
                  if (!var15) {
                     continue;
                  }
               }

               var10 += var6;
               var10000 = var4;
               var10001 = var8;
               break;
            }

            var4 = var10000 + var10001;
            var11++;
            if (var15) {
               break;
            }
         }
      } catch (RuntimeException var16) {
         throw i.a(
            var16,
            Yb[64]
               + (var1 != null ? Yb[1] : Yb[0])
               + ','
               + var2
               + ','
               + var3
               + ','
               + var4
               + ','
               + var5
               + ','
               + var6
               + ','
               + var7
               + ','
               + var8
               + ','
               + (var9 != null ? Yb[1] : Yb[0])
               + ','
               + var10
               + ')'
         );
      }
   }

   final void a(int var1, String var2, int var3, int var4, int var5, int var6) {
      try {
         this.a(11815, var3, var5, var4, var2, var1, var6);
         ab++;
      } catch (RuntimeException var8) {
         throw i.a(var8, Yb[90] + var1 + ',' + (var2 != null ? Yb[1] : Yb[0]) + ',' + var3 + ',' + var4 + ',' + var5 + ',' + var6 + ')');
      }
   }

   final void a(byte var1, byte[] var2, int var3) {
      boolean var12 = client.vh;

      try {
         Cb++;
         int[] var4 = this.ob[var3] = new int[10200];
         this.kb[var3] = 255;
         this.R[var3] = 40;
         this.Sb[var3] = 0;
         this.G[var3] = 0;
         if (var1 != -118) {
            this.a(82, -105, -7, 8, 9);
         }

         this.Eb[var3] = 255;
         this.qb[var3] = 40;
         this.Qb[var3] = false;
         int var5 = 0;
         int var6 = 1;
         int var7 = 0;

         byte var10000;
         while (true) {
            if (var7 < 255) {
               int var8 = 255 & var2[var6++];
               var10000 = 0;
               if (var12) {
                  break;
               }

               int var9 = 0;

               label102: {
                  while (~var8 < ~var9) {
                     var4[var7++] = var5;
                     var9++;
                     if (var12) {
                        break label102;
                     }

                     if (var12) {
                        break;
                     }
                  }

                  var5 = -var5 + 16777215;
               }

               if (!var12) {
                  continue;
               }
            }

            var10000 = 1;
            break;
         }

         int var14 = var10000;

         label87:
         while (true) {
            var10000 = ~var14;

            label85:
            while (var10000 > -41 && !var12) {
               int var15 = 0;

               while (~var15 > -256) {
                  int var10 = var2[var6++] & 255;
                  var10000 = 0;
                  if (var12) {
                     continue label85;
                  }

                  int var11 = 0;

                  label79: {
                     while (true) {
                        if (~var11 > ~var10) {
                           var4[var7] = var4[-255 + var7];
                           var7++;
                           var15++;
                           var11++;
                           if (var12) {
                              break;
                           }

                           if (!var12) {
                              continue;
                           }
                        }

                        if (~var15 <= -256) {
                           break label79;
                        }

                        var4[var7] = 16777215 - var4[var7 - 255];
                        var7++;
                        break;
                     }

                     var15++;
                  }

                  if (var12) {
                     break;
                  }
               }

               var14++;
               if (var12) {
                  return;
               }
               continue label87;
            }

            return;
         }
      } catch (RuntimeException var13) {
         throw i.a(var13, Yb[4] + var1 + ',' + (var2 != null ? Yb[1] : Yb[0]) + ',' + var3 + ')');
      }
   }

   final void a(boolean var1, int var2) {
      boolean var23 = client.vh;

      try {
         Ob++;
         int var3 = this.R[var2] * this.kb[var2];
         int[] var4 = this.ob[var2];
         int[] var5 = new int[32768];
         if (var1) {
            this.a(-63, 58, -7, -36, -99);
         }

         int var6 = 0;

         int[] var10000;
         while (true) {
            if (var6 < var3) {
               int var7 = var4[var6];
               var10000 = var5;
               if (var23) {
                  break;
               }

               var5[(31 & var7 >> 454314147) + (var7 >> 400635145 & 31744) + ((var7 & 63488) >> 303743686)]++;
               var6++;
               if (!var23) {
                  continue;
               }
            }

            var10000 = new int[256];
            break;
         }

         int[] var25 = var10000;
         var25[0] = 16711935;
         int[] var26 = new int[256];
         int var8 = 0;

         label124:
         while (true) {
            int var31 = var8;
            int var10001 = 32768;

            label121:
            while (var31 < var10001) {
               int var9 = var5[var8];
               var32 = var26[255];
               if (var23) {
                  break label124;
               }

               if (var32 < var9) {
                  int var10 = 1;

                  while (-257 < ~var10) {
                     var31 = var26[var10];
                     var10001 = var9;
                     if (var23) {
                        continue label121;
                     }

                     if (var31 < var9) {
                        int var11 = 255;

                        while (true) {
                           if (var10 < var11) {
                              var25[var11] = var25[-1 + var11];
                              var26[var11] = var26[var11 - 1];
                              var11--;
                              if (var23) {
                                 break;
                              }

                              if (!var23) {
                                 continue;
                              }
                           }

                           var25[var10] = 263172 + (ib.a(31, var8) << 257025667) + ib.a(63488, var8 << -1695574842) + ib.a(16252928, var8 << 986275945);
                           var26[var10] = var9;
                           break;
                        }

                        if (!var23) {
                           break;
                        }
                     }

                     var10++;
                     if (var23) {
                        break;
                     }
                  }
               }

               var5[var8] = -1;
               var8++;
               if (!var23) {
                  continue label124;
               }
               break;
            }

            var32 = var3;
            break;
         }

         byte[] var27 = new byte[var32];
         int var28 = 0;

         label87:
         while (true) {
            int var33 = ~var3;
            int var34 = ~var28;

            label84:
            while (var33 < var34) {
               int var29 = var4[var28];
               int var30 = (var29 >> 1087745987 & 31) + ((var29 & 63488) >> -1963155290) + ((16252928 & var29) >> 10645225);
               int var12 = var5[var30];
               if (var23) {
                  return;
               }

               if (-1 == var12) {
                  int var13 = 999999999;
                  int var14 = 0xFF & var29 >> 100065008;
                  int var15 = 0xFF & var29 >> 862974792;
                  int var16 = var29 & 0xFF;
                  int var17 = 0;

                  while (256 > var17) {
                     int var18 = var25[var17];
                     int var19 = (16719512 & var18) >> -112833712;
                     int var20 = var18 >> 1743095144 & 0xFF;
                     int var21 = 0xFF & var18;
                     int var22 = (-var21 + var16) * (-var21 + var16) + ((var14 - var19) * (var14 - var19) - -((-var20 + var15) * (-var20 + var15)));
                     var33 = var22;
                     var34 = var13;
                     if (var23) {
                        continue label84;
                     }

                     if (var22 < var13) {
                        var12 = var17;
                        var13 = var22;
                     }

                     var17++;
                     if (var23) {
                        break;
                     }
                  }

                  var5[var30] = var12;
               }

               var27[var28] = (byte)var12;
               var28++;
               if (!var23) {
                  continue label87;
               }
               break;
            }

            this.gb[var2] = var27;
            this.Y[var2] = var25;
            this.ob[var2] = null;
            return;
         }
      } catch (RuntimeException var24) {
         throw i.a(var24, Yb[44] + var1 + ',' + var2 + ')');
      }
   }

   final void a(int var1, String var2, int var3, int var4, int var5, int var6, boolean var7, int var8) {
      boolean var15 = client.vh;

      try {
         Q++;

         try {
            short var9 = 0;
            byte[] var10 = m.b[var5];
            if (var4 < 44) {
               return;
            }

            int var11 = 0;
            int var12 = 0;
            int var13 = 0;

            short var10000;
            label123: {
               while (var2.length() > var13) {
                  var10000 = -65;
                  if (var15) {
                     break label123;
                  }

                  label141: {
                     if (-65 != ~var2.charAt(var13) || ~var2.length() >= ~(4 + var13) || -65 != ~var2.charAt(var13 - -4)) {
                        if (-127 != ~var2.charAt(var13) || var2.length() <= 4 + var13 || ~var2.charAt(4 + var13) != -127) {
                           char var14 = var2.charAt(var13);
                           if (-1 < ~var14 || ~var14 <= ~n.a.length) {
                              var14 = ' ';
                           }

                           var9 += var10[7 + n.a[var14]];
                           if (!var15) {
                              break label141;
                           }
                        }

                        var13 += 4;
                        if (!var15) {
                           break label141;
                        }
                     }

                     var13 += 4;
                  }

                  if (var2.charAt(var13) == ' ') {
                     var12 = var13;
                  }

                  if (var2.charAt(var13) == '%' && var7) {
                     var9 = 1000;
                     var12 = var13;
                  }

                  if (var9 > var1) {
                     if (var12 <= var11) {
                        var12 = var13;
                     }

                     var9 = 0;
                     this.a(11815, var8, var5, 0, var2.substring(var11, var12), var3, var6);
                     var11 = var13 = 1 + var12;
                     var6 += this.a(508305352, var5);
                  }

                  var13++;
                  if (var15) {
                     break;
                  }
               }

               var10000 = var9;
            }

            if (var10000 > 0) {
               this.a(11815, var8, var5, 0, var2.substring(var11), var3, var6);
            }
         } catch (Exception var16) {
            System.out.println(Yb[81] + var16);
            var16.printStackTrace();
         }
      } catch (RuntimeException var17) {
         throw i.a(
            var17, Yb[80] + var1 + ',' + (var2 != null ? Yb[1] : Yb[0]) + ',' + var3 + ',' + var4 + ',' + var5 + ',' + var6 + ',' + var7 + ',' + var8 + ')'
         );
      }
   }

   final void a(int var1, int var2, int var3, boolean var4, int var5, int var6, int var7, int var8, int var9, int var10) {
      boolean var24 = client.vh;

      try {
         y++;

         try {
            if (var3 == 0) {
               var3 = 16777215;
            }

            if (-1 == ~var2) {
               var2 = 16777215;
            }

            int var11 = this.kb[var6];
            int var12 = this.R[var6];
            int var13 = 0;
            int var14 = 0;
            int var15 = var5 << -822406960;
            int var16 = (var11 << 264332976) / var8;
            int var17 = (var12 << -1079220976) / var7;
            int var18 = -(var5 << -1507499888) / var7;
            if (this.Qb[var6]) {
               int var19 = this.Eb[var6];
               int var20 = this.qb[var6];
               if (var19 == 0 || ~var20 == -1) {
                  return;
               }

               var16 = (var19 << -179525200) / var8;
               var17 = (var20 << 842218000) / var7;
               int var21 = this.Sb[var6];
               if (var4) {
                  var21 = var19 + -this.kb[var6] - var21;
               }

               int var22 = this.G[var6];
               var9 += (-1 + var19 + var21 * var8) / var19;
               int var23 = (-1 + var22 * var7 + var20) / var20;
               if (var21 * var8 % var19 != 0) {
                  var13 = (-(var8 * var21 % var19) + var19 << 306741872) / var8;
               }

               var1 += var23;
               var15 += var23 * var18;
               if (0 != var22 * var7 % var20) {
                  var14 = (var20 - var7 * var22 % var20 << -894050704) / var7;
               }

               var8 = (var16 + ((this.kb[var6] << -1406651696) - (var13 + 1))) / var16;
               var7 = ((this.R[var6] << -1596145424) + -var14 - (-var17 + 1)) / var17;
            }

            int var28 = this.u * var1;
            var15 += var9 << 189764144;
            if (var1 < this.A) {
               int var29 = this.A + -var1;
               var15 += var18 * var29;
               var7 -= var29;
               var14 += var29 * var17;
               var28 += this.u * var29;
               var1 = this.A;
            }

            if (var1 - -var7 >= this.Rb) {
               var7 -= 1 + var1 + var7 - this.Rb;
            }

            int var30 = var28 / this.u & var10;
            if (!this.i) {
               var30 = 2;
            }

            if (~var3 == -16777216) {
               if (null != this.ob[var6]) {
                  if (var4) {
                     this.a(
                        var10 ^ 74,
                        this.ob[var6],
                        var17,
                        0,
                        var14,
                        (this.kb[var6] << 102617264) + (-var13 - 1),
                        var8,
                        this.rb,
                        var7,
                        var18,
                        var28,
                        -var16,
                        var15,
                        var11,
                        var30,
                        var2
                     );
                     if (!var24) {
                        return;
                     }
                  }

                  this.a(var10 + 89, this.ob[var6], var17, 0, var14, var13, var8, this.rb, var7, var18, var28, var16, var15, var11, var30, var2);
                  if (!var24) {
                     return;
                  }
               }

               if (!var4) {
                  this.a(this.gb[var6], var2, 0, var18, this.Y[var6], var11, var14, var17, var15, var16, this.rb, var30, -110, var13, var28, var8, var7);
                  if (!var24) {
                     return;
                  }
               }

               this.a(
                  this.gb[var6],
                  var2,
                  0,
                  var18,
                  this.Y[var6],
                  var11,
                  var14,
                  var17,
                  var15,
                  -var16,
                  this.rb,
                  var30,
                  var10 ^ -124,
                  -1 + -var13 + (this.kb[var6] << -1997207152),
                  var28,
                  var8,
                  var7
               );
               if (!var24) {
                  return;
               }
            }

            if (this.ob[var6] == null) {
               if (var4) {
                  this.a(
                     var7,
                     this.Y[var6],
                     var11,
                     var17,
                     0,
                     var18,
                     var28,
                     var14,
                     this.gb[var6],
                     var30,
                     (byte)76,
                     -var13 + (this.kb[var6] << -1651772048) + -1,
                     var3,
                     var15,
                     -var16,
                     var8,
                     this.rb,
                     var2
                  );
                  if (!var24) {
                     return;
                  }
               }

               this.a(var7, this.Y[var6], var11, var17, 0, var18, var28, var14, this.gb[var6], var30, (byte)78, var13, var3, var15, var16, var8, this.rb, var2);
               if (!var24) {
                  return;
               }
            }

            if (var4) {
               this.a(
                  this.rb,
                  this.ob[var6],
                  var8,
                  var18,
                  var15,
                  var10 + 1603920391,
                  0,
                  var3,
                  var17,
                  -var16,
                  -var13 + (this.kb[var6] << -1212875536) + -1,
                  var30,
                  var14,
                  var11,
                  var2,
                  var7,
                  var28
               );
               if (!var24) {
                  return;
               }
            }

            this.a(this.rb, this.ob[var6], var8, var18, var15, 1603920392, 0, var3, var17, var16, var13, var30, var14, var11, var2, var7, var28);
         } catch (Exception var25) {
            System.out.println(Yb[16]);
         }
      } catch (RuntimeException var26) {
         throw i.a(
            var26, Yb[71] + var1 + ',' + var2 + ',' + var3 + ',' + var4 + ',' + var5 + ',' + var6 + ',' + var7 + ',' + var8 + ',' + var9 + ',' + var10 + ')'
         );
      }
   }

   @Override
   public final void startProduction(ImageConsumer var1) {
      try {
         this.addConsumer(var1);
         mb++;
      } catch (RuntimeException var3) {
         throw i.a(var3, Yb[12] + (var1 != null ? Yb[1] : Yb[0]) + ')');
      }
   }

   @Override
   public final synchronized boolean isConsumer(ImageConsumer var1) {
      try {
         q++;
         return this.fb == var1;
      } catch (RuntimeException var3) {
         throw i.a(var3, Yb[2] + (var1 != null ? Yb[1] : Yb[0]) + ')');
      }
   }

   private final void a(int var1, int[] var2, int var3, int var4, int var5, int var6, byte var7, int var8, int[] var9, int var10, int var11) {
      boolean var15 = client.vh;

      try {
         x++;
         if (var7 <= 122) {
            this.e(121, 54, -117, -34, 67, -103);
         }

         int var12 = -(var1 >> -1677003518);
         var1 = -(3 & var1);
         int var13 = -var4;

         while (~var13 > -1 && !var15) {
            int var14 = var12;

            int var10000;
            while (true) {
               if (var14 < 0) {
                  var5 = var9[var6++];
                  var10000 = var5;
                  if (var15) {
                     break;
                  }

                  label119: {
                     if (var5 == 0) {
                        var8++;
                        if (!var15) {
                           break label119;
                        }
                     }

                     var2[var8++] = var5;
                  }

                  label114: {
                     var5 = var9[var6++];
                     if (~var5 == -1) {
                        var8++;
                        if (!var15) {
                           break label114;
                        }
                     }

                     var2[var8++] = var5;
                  }

                  label109: {
                     var5 = var9[var6++];
                     if (0 != var5) {
                        var2[var8++] = var5;
                        if (!var15) {
                           break label109;
                        }
                     }

                     var8++;
                  }

                  label104: {
                     var5 = var9[var6++];
                     if (0 == var5) {
                        var8++;
                        if (!var15) {
                           break label104;
                        }
                     }

                     var2[var8++] = var5;
                  }

                  var14++;
                  if (!var15) {
                     continue;
                  }
               }

               var10000 = var1;
               break;
            }

            var14 = var10000;

            int var27;
            while (true) {
               if (-1 < ~var14) {
                  var5 = var9[var6++];
                  var10000 = 0;
                  var27 = var5;
                  if (var15) {
                     break;
                  }

                  label90: {
                     if (0 == var5) {
                        var8++;
                        if (!var15) {
                           break label90;
                        }
                     }

                     var2[var8++] = var5;
                  }

                  var14++;
                  if (!var15) {
                     continue;
                  }
               }

               var6 += var11;
               var8 += var10;
               var10000 = var13;
               var27 = var3;
               break;
            }

            var13 = var10000 + var27;
            if (var15) {
               break;
            }
         }
      } catch (RuntimeException var16) {
         throw i.a(
            var16,
            Yb[8]
               + var1
               + ','
               + (var2 != null ? Yb[1] : Yb[0])
               + ','
               + var3
               + ','
               + var4
               + ','
               + var5
               + ','
               + var6
               + ','
               + var7
               + ','
               + var8
               + ','
               + (var9 != null ? Yb[1] : Yb[0])
               + ','
               + var10
               + ','
               + var11
               + ')'
         );
      }
   }

   final void a(int var1, int var2, int var3, int var4, byte var5) {
      try {
         if (this.k < var3) {
            var3 = this.k;
         }

         if (0 > var4) {
            var4 = 0;
         }

         if (-1 < ~var1) {
            var1 = 0;
         }

         n++;
         if (~this.u > ~var2) {
            var2 = this.u;
         }

         if (var5 <= 15) {
            C = 109;
         }

         this.A = var4;
         this.Rb = var3;
         this.lb = var2;
         this.hb = var1;
      } catch (RuntimeException var7) {
         throw i.a(var7, Yb[6] + var1 + ',' + var2 + ',' + var3 + ',' + var4 + ',' + var5 + ')');
      }
   }

   final void a(boolean var1) {
      boolean var6 = client.vh;

      try {
         f++;
         int var2 = this.k * this.u;
         if (var1 == !this.i) {
            int var3 = 0;

            while (var2 > var3) {
               this.rb[var3] = 0;
               var3++;
               if (var6) {
                  return;
               }

               if (var6) {
                  break;
               }
            }

            if (!var6) {
               return;
            }
         }

         int var8 = 0;
         int var4 = -this.k;

         while (~var4 > -1 && !var6) {
            int var5 = -this.u;

            label48: {
               while (0 > var5) {
                  this.rb[var8++] = 0;
                  var5++;
                  if (var6) {
                     break label48;
                  }

                  if (var6) {
                     break;
                  }
               }

               var8 += this.u;
               var4 += 2;
            }

            if (var6) {
               break;
            }
         }
      } catch (RuntimeException var7) {
         throw i.a(var7, Yb[51] + var1 + ')');
      }
   }

   final void a(int var1, byte var2, int var3, int var4, int var5, int var6) {
      boolean var13 = client.vh;

      try {
         if (var1 < this.hb) {
            var6 -= this.hb + -var1;
            var1 = this.hb;
         }

         if (this.A > var4) {
            var5 -= -var4 + this.A;
            var4 = this.A;
         }

         K++;
         if (~(var4 + var5) < ~this.Rb) {
            var5 = -var4 + this.Rb;
         }

         if (~(var6 + var1) < ~this.lb) {
            var6 = -var1 + this.lb;
         }

         int var7 = this.u + -var6;
         byte var8 = 1;
         int var9 = -124 / ((-39 - var2) / 59);
         if (this.i) {
            var7 += this.u;
            if (0 != (var4 & 1)) {
               var5--;
               var4++;
            }

            var8 = 2;
         }

         int var10 = var1 - -(this.u * var4);
         int var11 = -var5;

         while (var11 < 0 && !var13) {
            int var12 = -var6;

            while (true) {
               if (-1 < ~var12) {
                  this.rb[var10++] = var3;
                  var12++;
                  if (var13) {
                     break;
                  }

                  if (!var13) {
                     continue;
                  }
               }

               var10 += var7;
               var11 += var8;
               break;
            }

            if (var13) {
               break;
            }
         }
      } catch (RuntimeException var14) {
         throw i.a(var14, Yb[3] + var1 + ',' + var2 + ',' + var3 + ',' + var4 + ',' + var5 + ',' + var6 + ')');
      }
   }

   final void a(String var1, int var2, int var3, int var4, boolean var5, int var6) {
      try {
         pb++;
         this.a(0, var3, var1, var2, var4, (byte)124, var6);
         if (var5) {
            this.a(-43, 36, -60, -88, 93, (int[])null, 114, (int[])null, -53, 59, true, 66, 34, 34, 70);
         }
      } catch (RuntimeException var8) {
         throw i.a(var8, Yb[50] + (var1 != null ? Yb[1] : Yb[0]) + ',' + var2 + ',' + var3 + ',' + var4 + ',' + var5 + ',' + var6 + ')');
      }
   }

   final void a(int var1) {
      try {
         this.lb = this.u;
         if (var1 != -1) {
            this.Sb = (int[])null;
         }

         this.hb = 0;
         this.A = 0;
         this.Rb = this.k;
         W++;
      } catch (RuntimeException var3) {
         throw i.a(var3, Yb[83] + var1 + ')');
      }
   }

   final void a(int var1, int var2, byte[] var3, int var4, byte[] var5) {
      boolean var17 = client.vh;

      try {
         jb++;
         int var6 = d.a(0, (byte)41, var3);
         int var7 = d.a(var6, (byte)48, var5);
         var6 += 2;
         if (var4 < 49) {
            this.b(-97, -40, 4, -57, 24, 50, 82);
         }

         int var8 = d.a(var6, (byte)15, var5);
         var6 += 2;
         int var9 = var5[var6++] & 255;
         int[] var10 = new int[var9];
         var10[0] = 16711935;
         int var11 = 0;

         while (true) {
            if (var9 + -1 > var11) {
               var10[var11 - -1] = (ib.a(255, var5[var6]) << -7887792) - -ib.a(65280, var5[var6 - -1] << -679902488) + ib.a(var5[var6 + 2], 255);
               var6 += 3;
               var11++;
               if (var17) {
                  break;
               }

               if (!var17) {
                  continue;
               }
            }

            var11 = 2;
            break;
         }

         int var12 = var1;

         while (true) {
            int var10000 = ~var12;

            label129:
            while (true) {
               int var30 = ~(var1 + var2);

               label126:
               while (true) {
                  if (var10000 <= var30) {
                     return;
                  }

                  this.Sb[var12] = ib.a(var5[var6++], 255);
                  this.G[var12] = ib.a(var5[var6++], 255);
                  this.kb[var12] = d.a(var6, (byte)32, var5);
                  int var24 = var6 + 2;
                  this.R[var12] = d.a(var24, (byte)83, var5);
                  var6 = var24 + 2;
                  int var13 = var5[var6++] & 255;
                  int var14 = this.R[var12] * this.kb[var12];
                  this.gb[var12] = new byte[var14];
                  this.Y[var12] = var10;
                  this.Eb[var12] = var7;
                  this.qb[var12] = var8;
                  this.ob[var12] = null;
                  this.Qb[var12] = false;
                  if (var17) {
                     return;
                  }

                  if (-1 != ~this.Sb[var12] || 0 != this.G[var12]) {
                     this.Qb[var12] = true;
                  }

                  if (~var13 != -1) {
                     if (~var13 != -2) {
                        break;
                     }

                     int var15 = 0;

                     label120:
                     do {
                        var10000 = ~this.kb[var12];

                        label117:
                        while (true) {
                           if (var10000 >= ~var15) {
                              break label120;
                           }

                           var10000 = 0;
                           if (var17) {
                              continue label129;
                           }

                           int var16 = 0;

                           while (true) {
                              if (this.R[var12] <= var16) {
                                 break label117;
                              }

                              this.gb[var12][var16 * this.kb[var12] + var15] = var3[var11++];
                              var10000 = this.gb[var12][var15 + var16 * this.kb[var12]];
                              if (var17) {
                                 break;
                              }

                              if (var10000 == 0) {
                                 this.Qb[var12] = true;
                              }

                              var16++;
                              if (var17) {
                                 break label117;
                              }
                           }
                        }

                        var15++;
                     } while (!var17);

                     if (!var17) {
                        break;
                     }
                  }

                  int var26 = 0;

                  while (true) {
                     if (var14 <= var26) {
                        break label126;
                     }

                     this.gb[var12][var26] = var3[var11++];
                     var10000 = 0;
                     var30 = this.gb[var12][var26];
                     if (var17) {
                        break;
                     }

                     if (0 == var30) {
                        this.Qb[var12] = true;
                     }

                     var26++;
                     if (var17) {
                        break label126;
                     }
                  }
               }

               var12++;
               if (var17) {
                  return;
               }
               break;
            }
         }
      } catch (RuntimeException var18) {
         throw i.a(var18, Yb[70] + var1 + ',' + var2 + ',' + (var3 != null ? Yb[1] : Yb[0]) + ',' + var4 + ',' + (var5 != null ? Yb[1] : Yb[0]) + ')');
      }
   }

   ua(int var1, int var2, int var3, Component var4) {
      boolean var7 = client.vh;
      super();
      this.xb = false;
      this.A = 0;
      this.i = false;
      this.hb = 0;
      this.lb = 0;
      this.Rb = 0;

      try {
         this.Y = new int[var3][];
         this.ob = new int[var3][];
         this.Rb = var2;
         this.gb = new byte[var3][];
         this.Qb = new boolean[var3];
         this.lb = var1;
         this.Eb = new int[var3];
         this.rb = new int[var1 * var2];
         this.Sb = new int[var3];
         this.G = new int[var3];
         this.qb = new int[var3];
         this.R = new int[var3];
         this.kb = new int[var3];
         this.k = var2;
         this.u = var1;
         if (~var1 < -2 && var2 > 1 && var4 != null) {
            this.nb = new DirectColorModel(32, 16711680, 65280, 255);
            int var5 = this.k * this.u;
            int var6 = 0;

            label39: {
               while (~var6 > ~var5) {
                  this.rb[var6] = 0;
                  var6++;
                  if (var7) {
                     break label39;
                  }

                  if (var7) {
                     break;
                  }
               }

               this.Gb = var4.createImage(this);
               this.b(true);
               var4.prepareImage(this.Gb, var4);
               this.b(true);
               var4.prepareImage(this.Gb, var4);
               this.b(true);
            }

            var4.prepareImage(this.Gb, var4);
         }
      } catch (RuntimeException var8) {
         throw i.a(var8, Yb[73] + var1 + ',' + var2 + ',' + var3 + ',' + (var4 != null ? Yb[1] : Yb[0]) + ')');
      }
   }

   private static char[] z(String var0) {
      char[] var10000 = var0.toCharArray();
      if (var10000.length < 2) {
         var10000[0] = (char)(var10000[0] ^ 9);
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
               var10005 = 63;
               break;
            case 1:
               var10005 = 48;
               break;
            case 2:
               var10005 = 37;
               break;
            case 3:
               var10005 = 6;
               break;
            default:
               var10005 = 9;
         }

         var10001[var1] = (char)(var10004 ^ var10005);
      }

      return new String(var10001).intern();
   }
}
