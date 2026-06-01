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

   final void b(int param1, int param2, int param3, int param4, int param5, int param6, int param7) {
      // $VF: Couldn't be decompiled
      // Please report this to the Vineflower issue tracker, at https://github.com/Vineflower/vineflower/issues with a copy of the class file (if you have the rights to distribute it!)
      // java.lang.RuntimeException: parsing failure!
      //   at org.jetbrains.java.decompiler.modules.decompiler.decompose.DomHelper.parseGraph(DomHelper.java:211)
      //   at org.jetbrains.java.decompiler.main.rels.MethodProcessor.codeToJava(MethodProcessor.java:166)
      //
      // Bytecode:
      // 000: getstatic client.vh Z
      // 003: istore 20
      // 005: getstatic ua.eb I
      // 008: bipush 1
      // 009: iadd
      // 00a: putstatic ua.eb I
      // 00d: aload 0
      // 00e: getfield ua.hb I
      // 011: iload 1
      // 012: if_icmple 024
      // 015: iload 3
      // 016: iload 1
      // 017: ineg
      // 018: aload 0
      // 019: getfield ua.hb I
      // 01c: iadd
      // 01d: isub
      // 01e: istore 3
      // 01f: aload 0
      // 020: getfield ua.hb I
      // 023: istore 1
      // 024: aload 0
      // 025: getfield ua.lb I
      // 028: iload 3
      // 029: iload 1
      // 02a: iadd
      // 02b: if_icmpge 036
      // 02e: iload 1
      // 02f: ineg
      // 030: aload 0
      // 031: getfield ua.lb I
      // 034: iadd
      // 035: istore 3
      // 036: iload 4
      // 038: ldc 16747974
      // 03a: iand
      // 03b: ldc 1599703024
      // 03d: ishr
      // 03e: istore 8
      // 040: iload 4
      // 042: ldc -272197656
      // 044: ishr
      // 045: sipush 255
      // 048: iand
      // 049: istore 9
      // 04b: iload 4
      // 04d: sipush 255
      // 050: iand
      // 051: istore 10
      // 053: iload 2
      // 054: ldc 16767844
      // 056: iand
      // 057: ldc -411014704
      // 059: ishr
      // 05a: istore 11
      // 05c: iload 2
      // 05d: ldc 442466760
      // 05f: ishr
      // 060: sipush 255
      // 063: iand
      // 064: istore 12
      // 066: iload 2
      // 067: sipush 255
      // 06a: iand
      // 06b: istore 13
      // 06d: iload 3
      // 06e: ineg
      // 06f: aload 0
      // 070: getfield ua.u I
      // 073: iadd
      // 074: istore 14
      // 076: bipush 1
      // 077: istore 15
      // 079: bipush 1
      // 07a: aload 0
      // 07b: getfield ua.i Z
      // 07e: ifne 086
      // 081: bipush 1
      // 082: goto 087
      // 085: athrow
      // 086: bipush 0
      // 087: if_icmpeq 0a8
      // 08a: iload 14
      // 08c: aload 0
      // 08d: getfield ua.u I
      // 090: iadd
      // 091: istore 14
      // 093: bipush 2
      // 094: istore 15
      // 096: bipush 0
      // 097: iload 6
      // 099: bipush 1
      // 09a: iand
      // 09b: if_icmpne 0a2
      // 09e: goto 0a8
      // 0a1: athrow
      // 0a2: iinc 6 1
      // 0a5: iinc 5 -1
      // 0a8: iload 7
      // 0aa: sipush 19020
      // 0ad: if_icmpeq 0c2
      // 0b0: aload 0
      // 0b1: bipush -124
      // 0b3: bipush 53
      // 0b5: bipush -53
      // 0b7: bipush -76
      // 0b9: bipush -44
      // 0bb: invokevirtual ua.a (IIIIB)V
      // 0be: goto 0c2
      // 0c1: athrow
      // 0c2: iload 1
      // 0c3: aload 0
      // 0c4: getfield ua.u I
      // 0c7: iload 6
      // 0c9: imul
      // 0ca: ineg
      // 0cb: isub
      // 0cc: istore 16
      // 0ce: bipush 0
      // 0cf: istore 17
      // 0d1: iload 5
      // 0d3: iload 17
      // 0d5: if_icmple 18c
      // 0d8: iload 20
      // 0da: ifne 1ed
      // 0dd: iload 17
      // 0df: iload 6
      // 0e1: iadd
      // 0e2: bipush -1
      // 0e3: ixor
      // 0e4: aload 0
      // 0e5: getfield ua.A I
      // 0e8: bipush -1
      // 0e9: ixor
      // 0ea: if_icmpgt 177
      // 0ed: goto 0f1
      // 0f0: athrow
      // 0f1: iload 6
      // 0f3: iload 17
      // 0f5: iadd
      // 0f6: aload 0
      // 0f7: getfield ua.Rb I
      // 0fa: if_icmpge 177
      // 0fd: goto 101
      // 100: athrow
      // 101: iload 9
      // 103: iload 17
      // 105: imul
      // 106: iload 12
      // 108: iload 17
      // 10a: ineg
      // 10b: iload 5
      // 10d: iadd
      // 10e: imul
      // 10f: iadd
      // 110: iload 5
      // 112: idiv
      // 113: ldc -1085162904
      // 115: ishl
      // 116: iload 11
      // 118: iload 17
      // 11a: ineg
      // 11b: iload 5
      // 11d: iadd
      // 11e: imul
      // 11f: iload 8
      // 121: iload 17
      // 123: imul
      // 124: iadd
      // 125: iload 5
      // 127: idiv
      // 128: ldc -1270717776
      // 12a: ishl
      // 12b: iadd
      // 12c: iload 17
      // 12e: iload 10
      // 130: imul
      // 131: iload 13
      // 133: iload 17
      // 135: ineg
      // 136: iload 5
      // 138: iadd
      // 139: imul
      // 13a: ineg
      // 13b: isub
      // 13c: iload 5
      // 13e: idiv
      // 13f: iadd
      // 140: istore 18
      // 142: iload 3
      // 143: ineg
      // 144: istore 19
      // 146: bipush -1
      // 147: iload 19
      // 149: bipush -1
      // 14a: ixor
      // 14b: if_icmpge 16b
      // 14e: aload 0
      // 14f: getfield ua.rb [I
      // 152: iload 16
      // 154: iinc 16 1
      // 157: iload 18
      // 159: iastore
      // 15a: iinc 19 1
      // 15d: iload 20
      // 15f: ifne 172
      // 162: iload 20
      // 164: ifeq 146
      // 167: goto 16b
      // 16a: athrow
      // 16b: iload 16
      // 16d: iload 14
      // 16f: iadd
      // 170: istore 16
      // 172: iload 20
      // 174: ifeq 180
      // 177: iload 16
      // 179: aload 0
      // 17a: getfield ua.u I
      // 17d: iadd
      // 17e: istore 16
      // 180: iload 17
      // 182: iload 15
      // 184: iadd
      // 185: istore 17
      // 187: iload 20
      // 189: ifeq 0d1
      // 18c: goto 1ed
      // 18f: astore 8
      // 191: aload 8
      // 193: new java/lang/StringBuilder
      // 196: dup
      // 197: invokespecial java/lang/StringBuilder.<init> ()V
      // 19a: getstatic ua.Yb [Ljava/lang/String;
      // 19d: bipush 10
      // 19f: aaload
      // 1a0: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 1a3: iload 1
      // 1a4: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 1a7: bipush 44
      // 1a9: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 1ac: iload 2
      // 1ad: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 1b0: bipush 44
      // 1b2: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 1b5: iload 3
      // 1b6: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 1b9: bipush 44
      // 1bb: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 1be: iload 4
      // 1c0: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 1c3: bipush 44
      // 1c5: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 1c8: iload 5
      // 1ca: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 1cd: bipush 44
      // 1cf: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 1d2: iload 6
      // 1d4: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 1d7: bipush 44
      // 1d9: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 1dc: iload 7
      // 1de: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 1e1: bipush 41
      // 1e3: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 1e6: invokevirtual java/lang/StringBuilder.toString ()Ljava/lang/String;
      // 1e9: invokestatic i.a (Ljava/lang/Throwable;Ljava/lang/String;)Lla;
      // 1ec: athrow
      // 1ed: return
   }

   // $VF: Handled exception range with multiple entry points by splitting it
   // $VF: Inserted dummy exception handlers to handle obfuscated exceptions
   private final void a(
      int[] var1, int var2, int var3, int var4, int var5, int[] var6, byte var7, int var8, int var9, int var10, int var11, int var12, int var13, int var14
   ) {
      boolean var20 = client.vh;

      try {
         O++;

         try {
            int var29 = var10;
            int var17 = -123 % ((-11 - var7) / 63);
            int var16 = -var9;

            while (~var16 > -1) {
               int var18 = (var5 >> -204204944) * var13;
               var5 += var8;
               if (var20) {
                  break;
               }

               int var19 = -var12;

               int var30;
               int var33;
               label94: {
                  while (-1 < ~var19) {
                     var4 = var1[(var10 >> 80730192) - -var18];
                     var10 += var3;

                     label112: {
                        label88: {
                           try {
                              var30 = 0;
                              var33 = var4;
                              if (var20) {
                                 break label94;
                              }

                              if (0 == var4) {
                                 break label88;
                              }
                           } catch (Exception var26) {
                              throw var26;
                           }

                           try {
                              try {
                                 var6[var14++] = var4;
                                 if (!var20) {
                                    break label112;
                                 }
                              } catch (Exception var23) {
                                 throw var23;
                              }
                           } catch (Exception var25) {
                              boolean var34 = false;
                              throw var25;
                           }
                        }

                        try {
                           var14++;
                        } catch (Exception var24) {
                           boolean var36 = false;
                           throw var24;
                        }
                     }

                     var19++;
                     if (var20) {
                        break;
                     }
                  }

                  var14 += var11;
                  var10 = var29;
                  var30 = var16;
                  var33 = var2;
               }

               var16 = var30 + var33;
               if (var20) {
                  break;
               }
            }
         } catch (Exception var27) {
            System.out.println(Yb[47]);
         }
      } catch (RuntimeException var28) {
         RuntimeException var15 = var28;

         RuntimeException var10000;
         StringBuilder var10001;
         String var10002;
         label61: {
            try {
               var10000 = var15;
               var10001 = new StringBuilder().append(Yb[48]);
               if (var1 != null) {
                  var10002 = Yb[1];
                  break label61;
               }
            } catch (Exception var22) {
               throw var22;
            }

            var10002 = Yb[0];
         }

         try {
            var10001 = var10001.append(var10002)
               .append(',')
               .append(var2)
               .append(',')
               .append(var3)
               .append(',')
               .append(var4)
               .append(',')
               .append(var5)
               .append(',');
            if (var6 != null) {
               throw i.a(
                  var10000,
                  var10001.append(Yb[1])
                     .append(',')
                     .append((int)var7)
                     .append(',')
                     .append(var8)
                     .append(',')
                     .append(var9)
                     .append(',')
                     .append(var10)
                     .append(',')
                     .append(var11)
                     .append(',')
                     .append(var12)
                     .append(',')
                     .append(var13)
                     .append(',')
                     .append(var14)
                     .append(')')
                     .toString()
               );
            }
         } catch (Exception var21) {
            throw var21;
         }

         throw i.a(
            var10000,
            var10001.append(Yb[0])
               .append(',')
               .append((int)var7)
               .append(',')
               .append(var8)
               .append(',')
               .append(var9)
               .append(',')
               .append(var10)
               .append(',')
               .append(var11)
               .append(',')
               .append(var12)
               .append(',')
               .append(var13)
               .append(',')
               .append(var14)
               .append(')')
               .toString()
         );
      }
   }

   private final void a(int var1, int var2, int var3, int[] var4, int var5, int var6, int var7, int[] var8, int var9, int var10, int var11, int var12) {
      boolean var17 = client.vh;

      try {
         c++;
         int var23 = 256 - var6;
         if (var9 <= -54) {
            int var14 = -var2;

            while (-1 < ~var14 && !var17) {
               int var15 = -var12;

               int var24;
               int var27;
               while (true) {
                  if (~var15 > -1) {
                     var5 = var4[var1++];

                     label78: {
                        label77: {
                           try {
                              var24 = 0;
                              var27 = var5;
                              if (var17) {
                                 break;
                              }

                              if (0 != var5) {
                                 break label77;
                              }
                           } catch (RuntimeException var21) {
                              throw var21;
                           }

                           try {
                              var3++;
                              if (!var17) {
                                 break label78;
                              }
                           } catch (RuntimeException var20) {
                              throw var20;
                           }
                        }

                        int var16 = var8[var3];
                        var8[var3++] = ib.a(16711680, var6 * ib.a(var5, 65280) + var23 * ib.a(65280, var16))
                              + ib.a(var23 * ib.a(var16, 16711935) + ib.a(var5, 16711935) * var6, -16711936)
                           >> -379053496;
                     }

                     var15++;
                     if (!var17) {
                        continue;
                     }
                  }

                  var1 += var10;
                  var3 += var11;
                  var24 = var14;
                  var27 = var7;
                  break;
               }

               var14 = var24 + var27;
               if (var17) {
                  break;
               }
            }
         }
      } catch (RuntimeException var22) {
         RuntimeException var13 = var22;

         RuntimeException var10000;
         StringBuilder var10001;
         String var10002;
         label57: {
            try {
               var10000 = var13;
               var10001 = new StringBuilder().append(Yb[91]).append(var1).append(',').append(var2).append(',').append(var3).append(',');
               if (var4 != null) {
                  var10002 = Yb[1];
                  break label57;
               }
            } catch (RuntimeException var19) {
               throw var19;
            }

            var10002 = Yb[0];
         }

         try {
            var10001 = var10001.append(var10002).append(',').append(var5).append(',').append(var6).append(',').append(var7).append(',');
            if (var8 != null) {
               throw i.a(
                  var10000,
                  var10001.append(Yb[1])
                     .append(',')
                     .append(var9)
                     .append(',')
                     .append(var10)
                     .append(',')
                     .append(var11)
                     .append(',')
                     .append(var12)
                     .append(')')
                     .toString()
               );
            }
         } catch (RuntimeException var18) {
            throw var18;
         }

         throw i.a(
            var10000,
            var10001.append(Yb[0])
               .append(',')
               .append(var9)
               .append(',')
               .append(var10)
               .append(',')
               .append(var11)
               .append(',')
               .append(var12)
               .append(')')
               .toString()
         );
      }
   }

   final void b(int var1, String var2, int var3, int var4, int var5, int var6) {
      try {
         int var10 = 24 % ((var5 - -11) / 58);
         s++;
         this.a(var1, var3, var2, var4, -12200, var6, 0);
      } catch (RuntimeException var9) {
         RuntimeException var7 = var9;

         RuntimeException var10000;
         StringBuilder var10001;
         try {
            var10000 = var7;
            var10001 = new StringBuilder().append(Yb[69]).append(var1).append(',');
            if (var2 != null) {
               throw i.a(
                  var7,
                  var10001.append(Yb[1])
                     .append(',')
                     .append(var3)
                     .append(',')
                     .append(var4)
                     .append(',')
                     .append(var5)
                     .append(',')
                     .append(var6)
                     .append(')')
                     .toString()
               );
            }
         } catch (RuntimeException var8) {
            throw var8;
         }

         throw i.a(
            var10000,
            var10001.append(Yb[0]).append(',').append(var3).append(',').append(var4).append(',').append(var5).append(',').append(var6).append(')').toString()
         );
      }
   }

   static final void a(int param0, int[] param1, int param2, int[] param3, int param4, int param5, int param6, int param7) {
      // $VF: Couldn't be decompiled
      // Please report this to the Vineflower issue tracker, at https://github.com/Vineflower/vineflower/issues with a copy of the class file (if you have the rights to distribute it!)
      // java.lang.RuntimeException: parsing failure!
      //   at org.jetbrains.java.decompiler.modules.decompiler.decompose.DomHelper.parseGraph(DomHelper.java:211)
      //   at org.jetbrains.java.decompiler.main.rels.MethodProcessor.codeToJava(MethodProcessor.java:166)
      //
      // Bytecode:
      // 000: getstatic client.vh Z
      // 003: istore 10
      // 005: getstatic ua.v I
      // 008: bipush 1
      // 009: iadd
      // 00a: putstatic ua.v I
      // 00d: bipush -1
      // 00e: iload 2
      // 00f: bipush -1
      // 010: ixor
      // 011: if_icmplt 015
      // 014: return
      // 015: aload 1
      // 016: ldc 65447
      // 018: iload 0
      // 019: iand
      // 01a: ldc 389407848
      // 01c: ishr
      // 01d: iaload
      // 01e: istore 4
      // 020: iload 5
      // 022: bipush 2
      // 023: ishl
      // 024: istore 5
      // 026: iload 0
      // 027: iload 5
      // 029: iadd
      // 02a: istore 0
      // 02b: iload 2
      // 02c: bipush 16
      // 02e: idiv
      // 02f: istore 8
      // 031: iload 8
      // 033: istore 9
      // 035: bipush -1
      // 036: iload 9
      // 038: bipush -1
      // 039: ixor
      // 03a: if_icmpge 1f4
      // 03d: aload 3
      // 03e: iload 6
      // 040: iinc 6 1
      // 043: iload 4
      // 045: ldc 8355711
      // 047: aload 3
      // 048: iload 6
      // 04a: iaload
      // 04b: ldc 980406721
      // 04d: ishr
      // 04e: invokestatic ib.a (II)I
      // 051: ineg
      // 052: isub
      // 053: iastore
      // 054: aload 3
      // 055: iload 6
      // 057: iinc 6 1
      // 05a: iload 4
      // 05c: ldc 8355711
      // 05e: aload 3
      // 05f: iload 6
      // 061: iaload
      // 062: ldc -33478591
      // 064: ishr
      // 065: invokestatic ib.a (II)I
      // 068: iadd
      // 069: iastore
      // 06a: aload 3
      // 06b: iload 6
      // 06d: iinc 6 1
      // 070: iload 4
      // 072: aload 3
      // 073: iload 6
      // 075: iaload
      // 076: ldc 16711423
      // 078: invokestatic ib.a (II)I
      // 07b: ldc 1438034561
      // 07d: ishr
      // 07e: iadd
      // 07f: iastore
      // 080: aload 3
      // 081: iload 6
      // 083: iinc 6 1
      // 086: ldc 16711423
      // 088: aload 3
      // 089: iload 6
      // 08b: iaload
      // 08c: invokestatic ib.a (II)I
      // 08f: ldc -1664277151
      // 091: ishr
      // 092: iload 4
      // 094: iadd
      // 095: iastore
      // 096: aload 1
      // 097: sipush 255
      // 09a: iload 0
      // 09b: ldc 1869141800
      // 09d: ishr
      // 09e: iand
      // 09f: iaload
      // 0a0: istore 4
      // 0a2: iload 0
      // 0a3: iload 5
      // 0a5: iadd
      // 0a6: istore 0
      // 0a7: aload 3
      // 0a8: iload 6
      // 0aa: iinc 6 1
      // 0ad: aload 3
      // 0ae: iload 6
      // 0b0: iaload
      // 0b1: ldc -651215775
      // 0b3: ishr
      // 0b4: ldc 8355711
      // 0b6: invokestatic ib.a (II)I
      // 0b9: iload 4
      // 0bb: iadd
      // 0bc: iastore
      // 0bd: aload 3
      // 0be: iload 6
      // 0c0: iinc 6 1
      // 0c3: aload 3
      // 0c4: iload 6
      // 0c6: iaload
      // 0c7: ldc 1567416321
      // 0c9: ishr
      // 0ca: ldc 8355711
      // 0cc: invokestatic ib.a (II)I
      // 0cf: iload 4
      // 0d1: iadd
      // 0d2: iastore
      // 0d3: aload 3
      // 0d4: iload 6
      // 0d6: iinc 6 1
      // 0d9: ldc 16711423
      // 0db: aload 3
      // 0dc: iload 6
      // 0de: iaload
      // 0df: invokestatic ib.a (II)I
      // 0e2: ldc -109945983
      // 0e4: ishr
      // 0e5: iload 4
      // 0e7: iadd
      // 0e8: iastore
      // 0e9: aload 3
      // 0ea: iload 6
      // 0ec: iinc 6 1
      // 0ef: aload 3
      // 0f0: iload 6
      // 0f2: iaload
      // 0f3: ldc -1634216127
      // 0f5: ishr
      // 0f6: ldc 8355711
      // 0f8: invokestatic ib.a (II)I
      // 0fb: iload 4
      // 0fd: iadd
      // 0fe: iastore
      // 0ff: aload 1
      // 100: iload 0
      // 101: ldc 1972579688
      // 103: ishr
      // 104: sipush 255
      // 107: iand
      // 108: iaload
      // 109: istore 4
      // 10b: aload 3
      // 10c: iload 6
      // 10e: iinc 6 1
      // 111: aload 3
      // 112: iload 6
      // 114: iaload
      // 115: ldc 16711422
      // 117: invokestatic ib.a (II)I
      // 11a: ldc 18481057
      // 11c: ishr
      // 11d: iload 4
      // 11f: iadd
      // 120: iastore
      // 121: iload 0
      // 122: iload 5
      // 124: iadd
      // 125: istore 0
      // 126: aload 3
      // 127: iload 6
      // 129: iinc 6 1
      // 12c: ldc 16711423
      // 12e: aload 3
      // 12f: iload 6
      // 131: iaload
      // 132: invokestatic ib.a (II)I
      // 135: ldc 1645567265
      // 137: ishr
      // 138: iload 4
      // 13a: iadd
      // 13b: iastore
      // 13c: aload 3
      // 13d: iload 6
      // 13f: iinc 6 1
      // 142: iload 4
      // 144: aload 3
      // 145: iload 6
      // 147: iaload
      // 148: ldc 363686529
      // 14a: ishr
      // 14b: ldc 8355711
      // 14d: invokestatic ib.a (II)I
      // 150: ineg
      // 151: isub
      // 152: iastore
      // 153: aload 3
      // 154: iload 6
      // 156: iinc 6 1
      // 159: iload 4
      // 15b: ldc 8355711
      // 15d: aload 3
      // 15e: iload 6
      // 160: iaload
      // 161: ldc -417782847
      // 163: ishr
      // 164: invokestatic ib.a (II)I
      // 167: iadd
      // 168: iastore
      // 169: aload 1
      // 16a: ldc 65302
      // 16c: iload 0
      // 16d: iand
      // 16e: ldc -491054904
      // 170: ishr
      // 171: iaload
      // 172: istore 4
      // 174: iload 0
      // 175: iload 5
      // 177: iadd
      // 178: istore 0
      // 179: aload 3
      // 17a: iload 6
      // 17c: iinc 6 1
      // 17f: aload 3
      // 180: iload 6
      // 182: iaload
      // 183: ldc 16711423
      // 185: invokestatic ib.a (II)I
      // 188: ldc -1655491807
      // 18a: ishr
      // 18b: iload 4
      // 18d: iadd
      // 18e: iastore
      // 18f: aload 3
      // 190: iload 6
      // 192: iinc 6 1
      // 195: iload 4
      // 197: aload 3
      // 198: iload 6
      // 19a: iaload
      // 19b: ldc 421283745
      // 19d: ishr
      // 19e: ldc 8355711
      // 1a0: invokestatic ib.a (II)I
      // 1a3: iadd
      // 1a4: iastore
      // 1a5: aload 3
      // 1a6: iload 6
      // 1a8: iinc 6 1
      // 1ab: iload 4
      // 1ad: aload 3
      // 1ae: iload 6
      // 1b0: iaload
      // 1b1: ldc 16711423
      // 1b3: invokestatic ib.a (II)I
      // 1b6: ldc 1309685921
      // 1b8: ishr
      // 1b9: ineg
      // 1ba: isub
      // 1bb: iastore
      // 1bc: aload 3
      // 1bd: iload 6
      // 1bf: iinc 6 1
      // 1c2: iload 4
      // 1c4: ldc 16711423
      // 1c6: aload 3
      // 1c7: iload 6
      // 1c9: iaload
      // 1ca: invokestatic ib.a (II)I
      // 1cd: ldc 1995672417
      // 1cf: ishr
      // 1d0: iadd
      // 1d1: iastore
      // 1d2: aload 1
      // 1d3: iload 0
      // 1d4: ldc 1728371944
      // 1d6: ishr
      // 1d7: sipush 255
      // 1da: iand
      // 1db: iaload
      // 1dc: istore 4
      // 1de: iload 0
      // 1df: iload 5
      // 1e1: iadd
      // 1e2: istore 0
      // 1e3: iinc 9 1
      // 1e6: iload 10
      // 1e8: ifne 1fb
      // 1eb: iload 10
      // 1ed: ifeq 035
      // 1f0: goto 1f4
      // 1f3: athrow
      // 1f4: iload 2
      // 1f5: bipush 16
      // 1f7: irem
      // 1f8: ineg
      // 1f9: istore 8
      // 1fb: iload 7
      // 1fd: istore 9
      // 1ff: iload 9
      // 201: bipush -1
      // 202: ixor
      // 203: iload 8
      // 205: bipush -1
      // 206: ixor
      // 207: if_icmple 252
      // 20a: aload 3
      // 20b: iload 6
      // 20d: iinc 6 1
      // 210: aload 3
      // 211: iload 6
      // 213: iaload
      // 214: ldc 1543799489
      // 216: ishr
      // 217: ldc 8355711
      // 219: invokestatic ib.a (II)I
      // 21c: iload 4
      // 21e: iadd
      // 21f: iastore
      // 220: iload 10
      // 222: ifne 2de
      // 225: bipush 3
      // 226: iload 9
      // 228: iand
      // 229: bipush 3
      // 22a: if_icmpeq 235
      // 22d: goto 231
      // 230: athrow
      // 231: goto 24a
      // 234: athrow
      // 235: aload 1
      // 236: iload 0
      // 237: ldc 65336
      // 239: iand
      // 23a: ldc -300394456
      // 23c: ishr
      // 23d: iaload
      // 23e: istore 4
      // 240: iload 0
      // 241: iload 5
      // 243: iadd
      // 244: istore 0
      // 245: iload 0
      // 246: iload 5
      // 248: iadd
      // 249: istore 0
      // 24a: iinc 9 1
      // 24d: iload 10
      // 24f: ifeq 1ff
      // 252: goto 2de
      // 255: astore 8
      // 257: aload 8
      // 259: new java/lang/StringBuilder
      // 25c: dup
      // 25d: invokespecial java/lang/StringBuilder.<init> ()V
      // 260: getstatic ua.Yb [Ljava/lang/String;
      // 263: bipush 63
      // 265: aaload
      // 266: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 269: iload 0
      // 26a: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 26d: bipush 44
      // 26f: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 272: aload 1
      // 273: ifnull 27f
      // 276: getstatic ua.Yb [Ljava/lang/String;
      // 279: bipush 1
      // 27a: aaload
      // 27b: goto 284
      // 27e: athrow
      // 27f: getstatic ua.Yb [Ljava/lang/String;
      // 282: bipush 0
      // 283: aaload
      // 284: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 287: bipush 44
      // 289: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 28c: iload 2
      // 28d: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 290: bipush 44
      // 292: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 295: aload 3
      // 296: ifnull 2a2
      // 299: getstatic ua.Yb [Ljava/lang/String;
      // 29c: bipush 1
      // 29d: aaload
      // 29e: goto 2a7
      // 2a1: athrow
      // 2a2: getstatic ua.Yb [Ljava/lang/String;
      // 2a5: bipush 0
      // 2a6: aaload
      // 2a7: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 2aa: bipush 44
      // 2ac: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 2af: iload 4
      // 2b1: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 2b4: bipush 44
      // 2b6: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 2b9: iload 5
      // 2bb: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 2be: bipush 44
      // 2c0: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 2c3: iload 6
      // 2c5: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 2c8: bipush 44
      // 2ca: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 2cd: iload 7
      // 2cf: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 2d2: bipush 41
      // 2d4: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 2d7: invokevirtual java/lang/StringBuilder.toString ()Ljava/lang/String;
      // 2da: invokestatic i.a (Ljava/lang/Throwable;Ljava/lang/String;)Lla;
      // 2dd: athrow
      // 2de: return
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
      } catch (RuntimeException var4) {
         RuntimeException var2 = var4;

         RuntimeException var10000;
         StringBuilder var10001;
         try {
            var10000 = var2;
            var10001 = new StringBuilder().append(Yb[78]);
            if (var1 != null) {
               throw i.a(var2, var10001.append(Yb[1]).append(')').toString());
            }
         } catch (RuntimeException var3) {
            throw var3;
         }

         throw i.a(var10000, var10001.append(Yb[0]).append(')').toString());
      }
   }

   final void b(int param1, int param2, int param3, int param4) {
      // $VF: Couldn't be decompiled
      // Please report this to the Vineflower issue tracker, at https://github.com/Vineflower/vineflower/issues with a copy of the class file (if you have the rights to distribute it!)
      // java.lang.RuntimeException: parsing failure!
      //   at org.jetbrains.java.decompiler.modules.decompiler.decompose.DomHelper.parseGraph(DomHelper.java:211)
      //   at org.jetbrains.java.decompiler.main.rels.MethodProcessor.codeToJava(MethodProcessor.java:166)
      //
      // Bytecode:
      // 000: iload 1
      // 001: bipush -1
      // 002: if_icmpeq 017
      // 005: aload 0
      // 006: bipush 41
      // 008: bipush 58
      // 00a: bipush 102
      // 00c: bipush 22
      // 00e: bipush 102
      // 010: invokevirtual ua.a (IIIIB)V
      // 013: goto 017
      // 016: athrow
      // 017: getstatic ua.Lb I
      // 01a: bipush 1
      // 01b: iadd
      // 01c: putstatic ua.Lb I
      // 01f: aload 0
      // 020: getfield ua.Qb [Z
      // 023: iload 2
      // 024: baload
      // 025: ifeq 03c
      // 028: iload 4
      // 02a: aload 0
      // 02b: getfield ua.Sb [I
      // 02e: iload 2
      // 02f: iaload
      // 030: iadd
      // 031: istore 4
      // 033: iload 3
      // 034: aload 0
      // 035: getfield ua.G [I
      // 038: iload 2
      // 039: iaload
      // 03a: iadd
      // 03b: istore 3
      // 03c: iload 3
      // 03d: aload 0
      // 03e: getfield ua.u I
      // 041: imul
      // 042: iload 4
      // 044: iadd
      // 045: istore 5
      // 047: bipush 0
      // 048: istore 6
      // 04a: aload 0
      // 04b: getfield ua.R [I
      // 04e: iload 2
      // 04f: iaload
      // 050: istore 7
      // 052: aload 0
      // 053: getfield ua.kb [I
      // 056: iload 2
      // 057: iaload
      // 058: istore 8
      // 05a: iload 8
      // 05c: ineg
      // 05d: aload 0
      // 05e: getfield ua.u I
      // 061: iadd
      // 062: istore 9
      // 064: bipush 0
      // 065: istore 10
      // 067: iload 3
      // 068: bipush -1
      // 069: ixor
      // 06a: aload 0
      // 06b: getfield ua.A I
      // 06e: bipush -1
      // 06f: ixor
      // 070: if_icmpgt 077
      // 073: goto 0a1
      // 076: athrow
      // 077: aload 0
      // 078: getfield ua.A I
      // 07b: iload 3
      // 07c: isub
      // 07d: istore 11
      // 07f: iload 7
      // 081: iload 11
      // 083: isub
      // 084: istore 7
      // 086: aload 0
      // 087: getfield ua.A I
      // 08a: istore 3
      // 08b: iload 5
      // 08d: aload 0
      // 08e: getfield ua.u I
      // 091: iload 11
      // 093: imul
      // 094: iadd
      // 095: istore 5
      // 097: iload 6
      // 099: iload 11
      // 09b: iload 8
      // 09d: imul
      // 09e: iadd
      // 09f: istore 6
      // 0a1: aload 0
      // 0a2: getfield ua.Rb I
      // 0a5: bipush -1
      // 0a6: ixor
      // 0a7: iload 3
      // 0a8: iload 7
      // 0aa: ineg
      // 0ab: isub
      // 0ac: bipush -1
      // 0ad: ixor
      // 0ae: if_icmplt 0c1
      // 0b1: iload 7
      // 0b3: bipush 1
      // 0b4: iload 7
      // 0b6: iload 3
      // 0b7: iadd
      // 0b8: aload 0
      // 0b9: getfield ua.Rb I
      // 0bc: isub
      // 0bd: iadd
      // 0be: isub
      // 0bf: istore 7
      // 0c1: aload 0
      // 0c2: getfield ua.hb I
      // 0c5: bipush -1
      // 0c6: ixor
      // 0c7: iload 4
      // 0c9: bipush -1
      // 0ca: ixor
      // 0cb: if_icmpge 101
      // 0ce: iload 4
      // 0d0: ineg
      // 0d1: aload 0
      // 0d2: getfield ua.hb I
      // 0d5: iadd
      // 0d6: istore 11
      // 0d8: iload 6
      // 0da: iload 11
      // 0dc: iadd
      // 0dd: istore 6
      // 0df: iload 9
      // 0e1: iload 11
      // 0e3: iadd
      // 0e4: istore 9
      // 0e6: iload 8
      // 0e8: iload 11
      // 0ea: isub
      // 0eb: istore 8
      // 0ed: iload 10
      // 0ef: iload 11
      // 0f1: iadd
      // 0f2: istore 10
      // 0f4: aload 0
      // 0f5: getfield ua.hb I
      // 0f8: istore 4
      // 0fa: iload 5
      // 0fc: iload 11
      // 0fe: iadd
      // 0ff: istore 5
      // 101: iload 4
      // 103: iload 8
      // 105: ineg
      // 106: isub
      // 107: aload 0
      // 108: getfield ua.lb I
      // 10b: if_icmpge 112
      // 10e: goto 137
      // 111: athrow
      // 112: iload 4
      // 114: iload 8
      // 116: ineg
      // 117: isub
      // 118: aload 0
      // 119: getfield ua.lb I
      // 11c: ineg
      // 11d: iadd
      // 11e: bipush -1
      // 11f: isub
      // 120: istore 11
      // 122: iload 8
      // 124: iload 11
      // 126: isub
      // 127: istore 8
      // 129: iload 10
      // 12b: iload 11
      // 12d: iadd
      // 12e: istore 10
      // 130: iload 9
      // 132: iload 11
      // 134: iadd
      // 135: istore 9
      // 137: bipush 0
      // 138: iload 8
      // 13a: if_icmpge 146
      // 13d: iload 7
      // 13f: ifgt 147
      // 142: goto 146
      // 145: athrow
      // 146: return
      // 147: bipush 1
      // 148: istore 11
      // 14a: aload 0
      // 14b: getfield ua.i Z
      // 14e: ifne 156
      // 151: bipush 1
      // 152: goto 157
      // 155: athrow
      // 156: bipush 0
      // 157: ifeq 15d
      // 15a: goto 189
      // 15d: iload 9
      // 15f: aload 0
      // 160: getfield ua.u I
      // 163: iadd
      // 164: istore 9
      // 166: bipush 1
      // 167: iload 3
      // 168: iand
      // 169: bipush -1
      // 16a: ixor
      // 16b: bipush -1
      // 16c: if_icmpeq 17b
      // 16f: iload 5
      // 171: aload 0
      // 172: getfield ua.u I
      // 175: iadd
      // 176: istore 5
      // 178: iinc 7 -1
      // 17b: bipush 2
      // 17c: istore 11
      // 17e: iload 10
      // 180: aload 0
      // 181: getfield ua.kb [I
      // 184: iload 2
      // 185: iaload
      // 186: iadd
      // 187: istore 10
      // 189: aload 0
      // 18a: getfield ua.ob [[I
      // 18d: iload 2
      // 18e: aaload
      // 18f: ifnull 1bb
      // 192: aload 0
      // 193: iload 8
      // 195: aload 0
      // 196: getfield ua.rb [I
      // 199: iload 11
      // 19b: iload 7
      // 19d: bipush 0
      // 19e: iload 6
      // 1a0: bipush 123
      // 1a2: iload 5
      // 1a4: aload 0
      // 1a5: getfield ua.ob [[I
      // 1a8: iload 2
      // 1a9: aaload
      // 1aa: iload 9
      // 1ac: iload 10
      // 1ae: invokespecial ua.a (I[IIIIIBI[III)V
      // 1b1: getstatic client.vh Z
      // 1b4: ifeq 1e2
      // 1b7: goto 1bb
      // 1ba: athrow
      // 1bb: aload 0
      // 1bc: iload 5
      // 1be: aload 0
      // 1bf: getfield ua.Y [[I
      // 1c2: iload 2
      // 1c3: aaload
      // 1c4: iload 6
      // 1c6: iload 10
      // 1c8: aload 0
      // 1c9: getfield ua.rb [I
      // 1cc: iload 11
      // 1ce: iload 7
      // 1d0: bipush 1
      // 1d1: aload 0
      // 1d2: getfield ua.gb [[B
      // 1d5: iload 2
      // 1d6: aaload
      // 1d7: iload 8
      // 1d9: iload 9
      // 1db: invokespecial ua.a (I[III[IIIZ[BII)V
      // 1de: goto 1e2
      // 1e1: athrow
      // 1e2: goto 225
      // 1e5: astore 5
      // 1e7: aload 5
      // 1e9: new java/lang/StringBuilder
      // 1ec: dup
      // 1ed: invokespecial java/lang/StringBuilder.<init> ()V
      // 1f0: getstatic ua.Yb [Ljava/lang/String;
      // 1f3: bipush 17
      // 1f5: aaload
      // 1f6: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 1f9: iload 1
      // 1fa: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 1fd: bipush 44
      // 1ff: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 202: iload 2
      // 203: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 206: bipush 44
      // 208: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 20b: iload 3
      // 20c: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 20f: bipush 44
      // 211: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 214: iload 4
      // 216: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 219: bipush 41
      // 21b: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 21e: invokevirtual java/lang/StringBuilder.toString ()Ljava/lang/String;
      // 221: invokestatic i.a (Ljava/lang/Throwable;Ljava/lang/String;)Lla;
      // 224: athrow
      // 225: return
   }

   @Override
   public final void requestTopDownLeftRightResend(ImageConsumer var1) {
      try {
         Ib++;
         System.out.println(Yb[59]);
      } catch (RuntimeException var4) {
         RuntimeException var2 = var4;

         RuntimeException var10000;
         StringBuilder var10001;
         try {
            var10000 = var2;
            var10001 = new StringBuilder().append(Yb[58]);
            if (var1 != null) {
               throw i.a(var2, var10001.append(Yb[1]).append(')').toString());
            }
         } catch (RuntimeException var3) {
            throw var3;
         }

         throw i.a(var10000, var10001.append(Yb[0]).append(')').toString());
      }
   }

   final void a(int param1, int param2, int param3, int param4, int param5, int param6, int param7) {
      // $VF: Couldn't be decompiled
      // Please report this to the Vineflower issue tracker, at https://github.com/Vineflower/vineflower/issues with a copy of the class file (if you have the rights to distribute it!)
      // java.lang.RuntimeException: parsing failure!
      //   at org.jetbrains.java.decompiler.modules.decompiler.decompose.DomHelper.parseGraph(DomHelper.java:211)
      //   at org.jetbrains.java.decompiler.main.rels.MethodProcessor.codeToJava(MethodProcessor.java:166)
      //
      // Bytecode:
      // 000: getstatic client.vh Z
      // 003: istore 17
      // 005: iload 4
      // 007: istore 8
      // 009: iload 4
      // 00b: iload 6
      // 00d: iadd
      // 00e: bipush -1
      // 00f: ixor
      // 010: iload 8
      // 012: bipush -1
      // 013: ixor
      // 014: if_icmpge 11d
      // 017: iload 3
      // 018: iload 17
      // 01a: ifne 127
      // 01d: istore 9
      // 01f: iload 9
      // 021: bipush -1
      // 022: ixor
      // 023: iload 3
      // 024: iload 1
      // 025: iadd
      // 026: bipush -1
      // 027: ixor
      // 028: if_icmple 115
      // 02b: bipush 0
      // 02c: istore 10
      // 02e: bipush 0
      // 02f: istore 11
      // 031: bipush 0
      // 032: istore 12
      // 034: bipush 0
      // 035: istore 13
      // 037: iload 7
      // 039: ineg
      // 03a: iload 8
      // 03c: iadd
      // 03d: iload 17
      // 03f: ifne 010
      // 042: istore 14
      // 044: iload 7
      // 046: iload 8
      // 048: iadd
      // 049: bipush -1
      // 04a: ixor
      // 04b: iload 14
      // 04d: bipush -1
      // 04e: ixor
      // 04f: if_icmpgt 0e7
      // 052: iload 14
      // 054: iload 17
      // 056: ifne 023
      // 059: iflt 0df
      // 05c: aload 0
      // 05d: getfield ua.u I
      // 060: iload 14
      // 062: if_icmpgt 06d
      // 065: goto 069
      // 068: athrow
      // 069: goto 0df
      // 06c: athrow
      // 06d: iload 9
      // 06f: iload 2
      // 070: isub
      // 071: istore 15
      // 073: iload 15
      // 075: iload 9
      // 077: iload 2
      // 078: iadd
      // 079: if_icmpgt 0df
      // 07c: bipush -1
      // 07d: iload 15
      // 07f: bipush -1
      // 080: ixor
      // 081: iload 17
      // 083: ifne 04f
      // 086: if_icmplt 0d7
      // 089: iload 15
      // 08b: bipush -1
      // 08c: ixor
      // 08d: aload 0
      // 08e: getfield ua.k I
      // 091: bipush -1
      // 092: ixor
      // 093: if_icmpgt 09e
      // 096: goto 09a
      // 099: athrow
      // 09a: goto 0d7
      // 09d: athrow
      // 09e: aload 0
      // 09f: getfield ua.rb [I
      // 0a2: aload 0
      // 0a3: getfield ua.u I
      // 0a6: iload 15
      // 0a8: imul
      // 0a9: iload 14
      // 0ab: iadd
      // 0ac: iaload
      // 0ad: istore 16
      // 0af: iload 12
      // 0b1: sipush 255
      // 0b4: iload 16
      // 0b6: iand
      // 0b7: iadd
      // 0b8: istore 12
      // 0ba: iinc 13 1
      // 0bd: iload 11
      // 0bf: iload 16
      // 0c1: ldc 65409
      // 0c3: iand
      // 0c4: ldc 743340392
      // 0c6: ishr
      // 0c7: iadd
      // 0c8: istore 11
      // 0ca: iload 10
      // 0cc: iload 16
      // 0ce: ldc 16737446
      // 0d0: iand
      // 0d1: ldc 483715504
      // 0d3: ishr
      // 0d4: iadd
      // 0d5: istore 10
      // 0d7: iinc 15 1
      // 0da: iload 17
      // 0dc: ifeq 073
      // 0df: iinc 14 1
      // 0e2: iload 17
      // 0e4: ifeq 044
      // 0e7: aload 0
      // 0e8: getfield ua.rb [I
      // 0eb: iload 8
      // 0ed: iload 9
      // 0ef: aload 0
      // 0f0: getfield ua.u I
      // 0f3: imul
      // 0f4: iadd
      // 0f5: iload 12
      // 0f7: iload 13
      // 0f9: idiv
      // 0fa: iload 10
      // 0fc: iload 13
      // 0fe: idiv
      // 0ff: ldc -148272656
      // 101: ishl
      // 102: iload 11
      // 104: iload 13
      // 106: idiv
      // 107: ldc 2002983304
      // 109: ishl
      // 10a: iadd
      // 10b: iadd
      // 10c: iastore
      // 10d: iinc 9 1
      // 110: iload 17
      // 112: ifeq 01f
      // 115: iinc 8 1
      // 118: iload 17
      // 11a: ifeq 009
      // 11d: getstatic ua.Vb I
      // 120: bipush 1
      // 121: iadd
      // 122: putstatic ua.Vb I
      // 125: iload 5
      // 127: ldc 16740352
      // 129: if_icmpeq 140
      // 12c: aload 0
      // 12d: bipush -18
      // 12f: bipush 79
      // 131: bipush -10
      // 133: bipush 106
      // 135: bipush -42
      // 137: bipush 27
      // 139: invokevirtual ua.a (IBIIII)V
      // 13c: goto 140
      // 13f: athrow
      // 140: goto 1a1
      // 143: astore 8
      // 145: aload 8
      // 147: new java/lang/StringBuilder
      // 14a: dup
      // 14b: invokespecial java/lang/StringBuilder.<init> ()V
      // 14e: getstatic ua.Yb [Ljava/lang/String;
      // 151: bipush 20
      // 153: aaload
      // 154: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 157: iload 1
      // 158: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 15b: bipush 44
      // 15d: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 160: iload 2
      // 161: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 164: bipush 44
      // 166: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 169: iload 3
      // 16a: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 16d: bipush 44
      // 16f: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 172: iload 4
      // 174: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 177: bipush 44
      // 179: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 17c: iload 5
      // 17e: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 181: bipush 44
      // 183: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 186: iload 6
      // 188: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 18b: bipush 44
      // 18d: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 190: iload 7
      // 192: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 195: bipush 41
      // 197: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 19a: invokevirtual java/lang/StringBuilder.toString ()Ljava/lang/String;
      // 19d: invokestatic i.a (Ljava/lang/Throwable;Ljava/lang/String;)Lla;
      // 1a0: athrow
      // 1a1: return
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

               try {
                  if (0 == var14 || -1 == ~var15) {
                     return;
                  }
               } catch (Exception var21) {
                  throw var21;
               }

               if (~(this.Sb[var3] * var7 % var14) != -1) {
                  var10 = (-(this.Sb[var3] * var7 % var14) + var14 << -1137453552) / var7;
               }

               var1 += (-1 + (var7 * this.Sb[var3] - -var14)) / var14;
               var12 = (var14 << -716906352) / var7;
               var4 += (var15 + var5 * this.G[var3] - 1) / var15;
               var13 = (var15 << 1305987728) / var5;

               label101: {
                  try {
                     if (-1 == ~(this.G[var3] * var5 % var15)) {
                        break label101;
                     }
                  } catch (Exception var20) {
                     throw var20;
                  }

                  var11 = (-(var5 * this.G[var3] % var15) + var15 << -1890544144) / var5;
               }

               var5 = var5 * (-(var11 >> 1020185680) + this.R[var3]) / var15;
               var7 = (-(var10 >> -839014480) + this.kb[var3]) * var7 / var14;
            }

            try {
               if (var6 <= 102) {
                  this.Y = (int[][])null;
               }
            } catch (Exception var19) {
               throw var19;
            }

            int var24 = var1 + var4 * this.u;
            int var25 = this.u - var7;
            if (~var4 > ~this.A) {
               int var16 = this.A + -var4;
               var4 = 0;
               var5 -= var16;
               var24 += var16 * this.u;
               var11 += var16 * var13;
            }

            if (~var1 > ~this.hb) {
               int var26 = this.hb - var1;
               var7 -= var26;
               var25 += var26;
               var10 += var26 * var12;
               var24 += var26;
               var1 = 0;
            }

            if (~(var4 - -var5) <= ~this.Rb) {
               var5 -= var4 - -var5 - (this.Rb + -1);
            }

            if (~(var1 + var7) <= ~this.lb) {
               int var27 = var1 - -var7 + (-this.lb - -1);
               var7 -= var27;
               var25 += var27;
            }

            byte var28 = 1;

            boolean var10000;
            label78: {
               try {
                  if (!this.i) {
                     var10000 = true;
                     break label78;
                  }
               } catch (Exception var18) {
                  throw var18;
               }

               var10000 = false;
            }

            if (!var10000) {
               var13 += var13;

               label69: {
                  try {
                     if ((1 & var4) == 0) {
                        break label69;
                     }
                  } catch (Exception var17) {
                     throw var17;
                  }

                  var5--;
                  var24 += this.u;
               }

               var28 = 2;
               var25 += this.u;
            }

            this.a(var11, var12, var7, var10, var25, this.ob[var3], var24, this.rb, 0, var8, false, var13, var5, var2, var28);
         } catch (Exception var22) {
            System.out.println(Yb[16]);
         }
      } catch (RuntimeException var23) {
         throw i.a(var23, Yb[57] + var1 + ',' + var2 + ',' + var3 + ',' + var4 + ',' + var5 + ',' + var6 + ',' + var7 + ')');
      }
   }

   final void b(int var1) {
      boolean var5 = client.vh;

      try {
         Z++;
         int var4 = this.k * this.u;

         try {
            if (var1 != 16316665) {
               wb = (String[])null;
            }
         } catch (RuntimeException var7) {
            throw var7;
         }

         int var3 = 0;

         while (var4 > var3) {
            int var2 = 16777215 & this.rb[var3];

            try {
               this.rb[var3] = ib.a(var2 >>> -93223452, 986895)
                  + ((ib.a(var2, 16316665) >>> 2097500387) + (ib.a(16711423, var2) >>> 1336934849) - -ib.a(-2143338689, var2 >>> 1527263298));
               var3++;
               if (var5 || var5) {
                  break;
               }
            } catch (RuntimeException var6) {
               throw var6;
            }
         }
      } catch (RuntimeException var8) {
         throw i.a(var8, Yb[55] + var1 + ')');
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

               try {
                  var3++;
                  if (var13) {
                     break;
                  }

                  if (!var13) {
                     continue;
                  }
               } catch (RuntimeException var17) {
                  throw var17;
               }
            }

            try {
               if (!var11) {
                  this.a(-59, -116, -115, true, 1, 118, 33, -46, -78, -30);
               }
            } catch (RuntimeException var16) {
               throw var16;
            }

            V++;
            break;
         }
      } catch (RuntimeException var18) {
         RuntimeException var12 = var18;

         RuntimeException var10000;
         StringBuilder var10001;
         String var10002;
         label47: {
            try {
               var10000 = var12;
               var10001 = new StringBuilder().append(Yb[68]).append(var1).append(',').append(var2).append(',').append(var3).append(',');
               if (var4 != null) {
                  var10002 = Yb[1];
                  break label47;
               }
            } catch (RuntimeException var15) {
               throw var15;
            }

            var10002 = Yb[0];
         }

         try {
            var10001 = var10001.append(var10002).append(',').append(var5).append(',').append(var6).append(',').append(var7).append(',');
            if (var8 != null) {
               throw i.a(var10000, var10001.append(Yb[1]).append(',').append(var9).append(',').append(var10).append(',').append(var11).append(')').toString());
            }
         } catch (RuntimeException var14) {
            throw var14;
         }

         throw i.a(var10000, var10001.append(Yb[0]).append(',').append(var9).append(',').append(var10).append(',').append(var11).append(')').toString());
      }
   }

   final void a(int param1, int param2, int param3, int param4) {
      // $VF: Couldn't be decompiled
      // Please report this to the Vineflower issue tracker, at https://github.com/Vineflower/vineflower/issues with a copy of the class file (if you have the rights to distribute it!)
      // java.lang.RuntimeException: parsing failure!
      //   at org.jetbrains.java.decompiler.modules.decompiler.decompose.DomHelper.parseGraph(DomHelper.java:211)
      //   at org.jetbrains.java.decompiler.main.rels.MethodProcessor.codeToJava(MethodProcessor.java:166)
      //
      // Bytecode:
      // 00: getstatic ua.db I
      // 03: bipush 1
      // 04: iadd
      // 05: putstatic ua.db I
      // 08: aload 0
      // 09: getfield ua.hb I
      // 0c: iload 2
      // 0d: if_icmpgt 44
      // 10: iload 1
      // 11: bipush -1
      // 12: ixor
      // 13: aload 0
      // 14: getfield ua.A I
      // 17: bipush -1
      // 18: ixor
      // 19: if_icmpgt 44
      // 1c: goto 20
      // 1f: athrow
      // 20: iload 2
      // 21: bipush -1
      // 22: ixor
      // 23: aload 0
      // 24: getfield ua.lb I
      // 27: bipush -1
      // 28: ixor
      // 29: if_icmple 44
      // 2c: goto 30
      // 2f: athrow
      // 30: iload 1
      // 31: bipush -1
      // 32: ixor
      // 33: aload 0
      // 34: getfield ua.Rb I
      // 37: bipush -1
      // 38: ixor
      // 39: if_icmple 44
      // 3c: goto 40
      // 3f: athrow
      // 40: goto 45
      // 43: athrow
      // 44: return
      // 45: iload 3
      // 46: bipush 44
      // 48: if_icmpgt 4c
      // 4b: return
      // 4c: aload 0
      // 4d: getfield ua.rb [I
      // 50: iload 2
      // 51: aload 0
      // 52: getfield ua.u I
      // 55: iload 1
      // 56: imul
      // 57: ineg
      // 58: isub
      // 59: iload 4
      // 5b: iastore
      // 5c: goto 9f
      // 5f: astore 5
      // 61: aload 5
      // 63: new java/lang/StringBuilder
      // 66: dup
      // 67: invokespecial java/lang/StringBuilder.<init> ()V
      // 6a: getstatic ua.Yb [Ljava/lang/String;
      // 6d: bipush 61
      // 6f: aaload
      // 70: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 73: iload 1
      // 74: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 77: bipush 44
      // 79: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 7c: iload 2
      // 7d: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 80: bipush 44
      // 82: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 85: iload 3
      // 86: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 89: bipush 44
      // 8b: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 8e: iload 4
      // 90: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 93: bipush 41
      // 95: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 98: invokevirtual java/lang/StringBuilder.toString ()Ljava/lang/String;
      // 9b: invokestatic i.a (Ljava/lang/Throwable;Ljava/lang/String;)Lla;
      // 9e: athrow
      // 9f: return
   }

   final void c(int param1, int param2, int param3, int param4, int param5, int param6, int param7) {
      // $VF: Couldn't be decompiled
      // Please report this to the Vineflower issue tracker, at https://github.com/Vineflower/vineflower/issues with a copy of the class file (if you have the rights to distribute it!)
      // java.lang.RuntimeException: parsing failure!
      //   at org.jetbrains.java.decompiler.modules.decompiler.decompose.DomHelper.parseGraph(DomHelper.java:211)
      //   at org.jetbrains.java.decompiler.main.rels.MethodProcessor.codeToJava(MethodProcessor.java:166)
      //
      // Bytecode:
      // 000: getstatic client.vh Z
      // 003: istore 21
      // 005: getstatic ua.z I
      // 008: bipush 1
      // 009: iadd
      // 00a: putstatic ua.z I
      // 00d: iload 5
      // 00f: aload 0
      // 010: getfield ua.A I
      // 013: if_icmplt 01a
      // 016: goto 02a
      // 019: athrow
      // 01a: iload 3
      // 01b: aload 0
      // 01c: getfield ua.A I
      // 01f: iload 5
      // 021: isub
      // 022: isub
      // 023: istore 3
      // 024: aload 0
      // 025: getfield ua.A I
      // 028: istore 5
      // 02a: iload 2
      // 02b: bipush -1
      // 02c: ixor
      // 02d: aload 0
      // 02e: getfield ua.hb I
      // 031: bipush -1
      // 032: ixor
      // 033: if_icmpgt 03a
      // 036: goto 04a
      // 039: athrow
      // 03a: iload 6
      // 03c: aload 0
      // 03d: getfield ua.hb I
      // 040: iload 2
      // 041: isub
      // 042: isub
      // 043: istore 6
      // 045: aload 0
      // 046: getfield ua.hb I
      // 049: istore 2
      // 04a: aload 0
      // 04b: getfield ua.lb I
      // 04e: iload 2
      // 04f: iload 6
      // 051: ineg
      // 052: isub
      // 053: if_icmpge 05e
      // 056: aload 0
      // 057: getfield ua.lb I
      // 05a: iload 2
      // 05b: isub
      // 05c: istore 6
      // 05e: aload 0
      // 05f: getfield ua.Rb I
      // 062: iload 3
      // 063: iload 5
      // 065: iadd
      // 066: if_icmplt 06d
      // 069: goto 076
      // 06c: athrow
      // 06d: iload 5
      // 06f: ineg
      // 070: aload 0
      // 071: getfield ua.Rb I
      // 074: iadd
      // 075: istore 3
      // 076: sipush 256
      // 079: iload 1
      // 07a: isub
      // 07b: istore 8
      // 07d: iload 1
      // 07e: iload 7
      // 080: ldc -2055680880
      // 082: ishr
      // 083: sipush 255
      // 086: iand
      // 087: imul
      // 088: istore 9
      // 08a: iload 7
      // 08c: ldc 65476
      // 08e: iand
      // 08f: ldc 1364192264
      // 091: ishr
      // 092: iload 1
      // 093: imul
      // 094: istore 10
      // 096: iload 1
      // 097: iload 7
      // 099: sipush 255
      // 09c: iand
      // 09d: imul
      // 09e: istore 11
      // 0a0: aload 0
      // 0a1: getfield ua.u I
      // 0a4: iload 6
      // 0a6: isub
      // 0a7: istore 15
      // 0a9: bipush 1
      // 0aa: istore 16
      // 0ac: bipush 0
      // 0ad: aload 0
      // 0ae: getfield ua.i Z
      // 0b1: ifne 0b9
      // 0b4: bipush 1
      // 0b5: goto 0ba
      // 0b8: athrow
      // 0b9: bipush 0
      // 0ba: if_icmpne 0e1
      // 0bd: bipush -1
      // 0be: iload 5
      // 0c0: bipush 1
      // 0c1: iand
      // 0c2: bipush -1
      // 0c3: ixor
      // 0c4: if_icmpne 0cf
      // 0c7: goto 0cb
      // 0ca: athrow
      // 0cb: goto 0d5
      // 0ce: athrow
      // 0cf: iinc 3 -1
      // 0d2: iinc 5 1
      // 0d5: iload 15
      // 0d7: aload 0
      // 0d8: getfield ua.u I
      // 0db: iadd
      // 0dc: istore 15
      // 0de: bipush 2
      // 0df: istore 16
      // 0e1: iload 2
      // 0e2: aload 0
      // 0e3: getfield ua.u I
      // 0e6: iload 5
      // 0e8: imul
      // 0e9: ineg
      // 0ea: isub
      // 0eb: istore 17
      // 0ed: iload 4
      // 0ef: istore 18
      // 0f1: iload 3
      // 0f2: bipush -1
      // 0f3: ixor
      // 0f4: iload 18
      // 0f6: bipush -1
      // 0f7: ixor
      // 0f8: if_icmpge 191
      // 0fb: iload 21
      // 0fd: ifne 1f2
      // 100: iload 6
      // 102: ineg
      // 103: istore 19
      // 105: bipush 0
      // 106: iload 19
      // 108: if_icmple 17e
      // 10b: iload 8
      // 10d: aload 0
      // 10e: getfield ua.rb [I
      // 111: iload 17
      // 113: iaload
      // 114: sipush 255
      // 117: iand
      // 118: imul
      // 119: istore 14
      // 11b: iload 8
      // 11d: ldc 16746549
      // 11f: aload 0
      // 120: getfield ua.rb [I
      // 123: iload 17
      // 125: iaload
      // 126: iand
      // 127: ldc 1661674448
      // 129: ishr
      // 12a: imul
      // 12b: istore 12
      // 12d: iload 8
      // 12f: ldc 65402
      // 131: aload 0
      // 132: getfield ua.rb [I
      // 135: iload 17
      // 137: iaload
      // 138: iand
      // 139: ldc 2108168104
      // 13b: ishr
      // 13c: imul
      // 13d: istore 13
      // 13f: iload 14
      // 141: iload 11
      // 143: iadd
      // 144: ldc -1220075704
      // 146: ishr
      // 147: iload 13
      // 149: iload 10
      // 14b: iadd
      // 14c: ldc -855772152
      // 14e: ishr
      // 14f: ldc -628820632
      // 151: ishl
      // 152: iadd
      // 153: iload 9
      // 155: iload 12
      // 157: iadd
      // 158: ldc -540786712
      // 15a: ishr
      // 15b: ldc -681889584
      // 15d: ishl
      // 15e: iadd
      // 15f: istore 20
      // 161: aload 0
      // 162: getfield ua.rb [I
      // 165: iload 17
      // 167: iinc 17 1
      // 16a: iload 20
      // 16c: iastore
      // 16d: iinc 19 1
      // 170: iload 21
      // 172: ifne 18c
      // 175: iload 21
      // 177: ifeq 105
      // 17a: goto 17e
      // 17d: athrow
      // 17e: iload 17
      // 180: iload 15
      // 182: iadd
      // 183: istore 17
      // 185: iload 18
      // 187: iload 16
      // 189: iadd
      // 18a: istore 18
      // 18c: iload 21
      // 18e: ifeq 0f1
      // 191: goto 1f2
      // 194: astore 8
      // 196: aload 8
      // 198: new java/lang/StringBuilder
      // 19b: dup
      // 19c: invokespecial java/lang/StringBuilder.<init> ()V
      // 19f: getstatic ua.Yb [Ljava/lang/String;
      // 1a2: bipush 62
      // 1a4: aaload
      // 1a5: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 1a8: iload 1
      // 1a9: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 1ac: bipush 44
      // 1ae: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 1b1: iload 2
      // 1b2: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 1b5: bipush 44
      // 1b7: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 1ba: iload 3
      // 1bb: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 1be: bipush 44
      // 1c0: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 1c3: iload 4
      // 1c5: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 1c8: bipush 44
      // 1ca: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 1cd: iload 5
      // 1cf: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 1d2: bipush 44
      // 1d4: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 1d7: iload 6
      // 1d9: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 1dc: bipush 44
      // 1de: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 1e1: iload 7
      // 1e3: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 1e6: bipush 41
      // 1e8: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 1eb: invokevirtual java/lang/StringBuilder.toString ()Ljava/lang/String;
      // 1ee: invokestatic i.a (Ljava/lang/Throwable;Ljava/lang/String;)Lla;
      // 1f1: athrow
      // 1f2: return
   }

   final int a(int param1, int param2, String param3) {
      // $VF: Couldn't be decompiled
      // Please report this to the Vineflower issue tracker, at https://github.com/Vineflower/vineflower/issues with a copy of the class file (if you have the rights to distribute it!)
      // java.lang.RuntimeException: parsing failure!
      //   at org.jetbrains.java.decompiler.modules.decompiler.decompose.DomHelper.parseGraph(DomHelper.java:211)
      //   at org.jetbrains.java.decompiler.main.rels.MethodProcessor.codeToJava(MethodProcessor.java:166)
      //
      // Bytecode:
      // 000: getstatic client.vh Z
      // 003: istore 8
      // 005: getstatic ua.Nb I
      // 008: bipush 1
      // 009: iadd
      // 00a: putstatic ua.Nb I
      // 00d: bipush 0
      // 00e: istore 4
      // 010: iload 2
      // 011: bipush 67
      // 013: if_icmpgt 020
      // 016: aload 0
      // 017: bipush 74
      // 019: putfield ua.Rb I
      // 01c: goto 020
      // 01f: athrow
      // 020: getstatic m.b [[B
      // 023: iload 1
      // 024: aaload
      // 025: astore 5
      // 027: bipush 0
      // 028: istore 6
      // 02a: iload 6
      // 02c: aload 3
      // 02d: invokevirtual java/lang/String.length ()I
      // 030: if_icmpge 0ef
      // 033: bipush 64
      // 035: iload 8
      // 037: ifne 0f1
      // 03a: aload 3
      // 03b: iload 6
      // 03d: invokevirtual java/lang/String.charAt (I)C
      // 040: if_icmpne 074
      // 043: goto 047
      // 046: athrow
      // 047: bipush 4
      // 048: iload 6
      // 04a: iadd
      // 04b: aload 3
      // 04c: invokevirtual java/lang/String.length ()I
      // 04f: if_icmpge 074
      // 052: goto 056
      // 055: athrow
      // 056: aload 3
      // 057: iload 6
      // 059: bipush -4
      // 05b: isub
      // 05c: invokevirtual java/lang/String.charAt (I)C
      // 05f: bipush 64
      // 061: if_icmpne 074
      // 064: goto 068
      // 067: athrow
      // 068: iinc 6 4
      // 06b: iload 8
      // 06d: ifeq 0e7
      // 070: goto 074
      // 073: athrow
      // 074: aload 3
      // 075: iload 6
      // 077: invokevirtual java/lang/String.charAt (I)C
      // 07a: bipush -1
      // 07b: ixor
      // 07c: bipush -127
      // 07e: if_icmpne 0a5
      // 081: goto 085
      // 084: athrow
      // 085: iload 6
      // 087: bipush 4
      // 088: iadd
      // 089: aload 3
      // 08a: invokevirtual java/lang/String.length ()I
      // 08d: if_icmpge 0a5
      // 090: goto 094
      // 093: athrow
      // 094: bipush 126
      // 096: aload 3
      // 097: iload 6
      // 099: bipush 4
      // 09a: iadd
      // 09b: invokevirtual java/lang/String.charAt (I)C
      // 09e: if_icmpeq 0e0
      // 0a1: goto 0a5
      // 0a4: athrow
      // 0a5: aload 3
      // 0a6: iload 6
      // 0a8: invokevirtual java/lang/String.charAt (I)C
      // 0ab: istore 7
      // 0ad: iload 7
      // 0af: bipush -1
      // 0b0: ixor
      // 0b1: bipush -1
      // 0b2: if_icmpgt 0c6
      // 0b5: getstatic n.a [I
      // 0b8: arraylength
      // 0b9: bipush -1
      // 0ba: ixor
      // 0bb: iload 7
      // 0bd: bipush -1
      // 0be: ixor
      // 0bf: if_icmplt 0ca
      // 0c2: goto 0c6
      // 0c5: athrow
      // 0c6: bipush 32
      // 0c8: istore 7
      // 0ca: iload 4
      // 0cc: aload 5
      // 0ce: getstatic n.a [I
      // 0d1: iload 7
      // 0d3: iaload
      // 0d4: bipush 7
      // 0d6: iadd
      // 0d7: baload
      // 0d8: iadd
      // 0d9: istore 4
      // 0db: iload 8
      // 0dd: ifeq 0e7
      // 0e0: iinc 6 4
      // 0e3: goto 0e7
      // 0e6: athrow
      // 0e7: iinc 6 1
      // 0ea: iload 8
      // 0ec: ifeq 02a
      // 0ef: iload 4
      // 0f1: ireturn
      // 0f2: astore 4
      // 0f4: aload 4
      // 0f6: new java/lang/StringBuilder
      // 0f9: dup
      // 0fa: invokespecial java/lang/StringBuilder.<init> ()V
      // 0fd: getstatic ua.Yb [Ljava/lang/String;
      // 100: bipush 11
      // 102: aaload
      // 103: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 106: iload 1
      // 107: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 10a: bipush 44
      // 10c: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 10f: iload 2
      // 110: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 113: bipush 44
      // 115: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 118: aload 3
      // 119: ifnull 125
      // 11c: getstatic ua.Yb [Ljava/lang/String;
      // 11f: bipush 1
      // 120: aaload
      // 121: goto 12a
      // 124: athrow
      // 125: getstatic ua.Yb [Ljava/lang/String;
      // 128: bipush 0
      // 129: aaload
      // 12a: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 12d: bipush 41
      // 12f: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 132: invokevirtual java/lang/StringBuilder.toString ()Ljava/lang/String;
      // 135: invokestatic i.a (Ljava/lang/Throwable;Ljava/lang/String;)Lla;
      // 138: athrow
   }

   @Override
   public final boolean imageUpdate(Image var1, int var2, int var3, int var4, int var5, int var6) {
      try {
         B++;
         return true;
      } catch (RuntimeException var9) {
         RuntimeException var7 = var9;

         RuntimeException var10000;
         StringBuilder var10001;
         try {
            var10000 = var7;
            var10001 = new StringBuilder().append(Yb[84]);
            if (var1 != null) {
               throw i.a(
                  var7,
                  var10001.append(Yb[1])
                     .append(',')
                     .append(var2)
                     .append(',')
                     .append(var3)
                     .append(',')
                     .append(var4)
                     .append(',')
                     .append(var5)
                     .append(',')
                     .append(var6)
                     .append(')')
                     .toString()
               );
            }
         } catch (RuntimeException var8) {
            throw var8;
         }

         throw i.a(
            var10000,
            var10001.append(Yb[0])
               .append(',')
               .append(var2)
               .append(',')
               .append(var3)
               .append(',')
               .append(var4)
               .append(',')
               .append(var5)
               .append(',')
               .append(var6)
               .append(')')
               .toString()
         );
      }
   }

   private final void a(
      int param1, int param2, int param3, int[] param4, int[] param5, int param6, int param7, int param8, int param9, int param10, byte param11
   ) {
      // $VF: Couldn't be decompiled
      // Please report this to the Vineflower issue tracker, at https://github.com/Vineflower/vineflower/issues with a copy of the class file (if you have the rights to distribute it!)
      // java.lang.RuntimeException: parsing failure!
      //   at org.jetbrains.java.decompiler.modules.decompiler.decompose.DomHelper.parseGraph(DomHelper.java:211)
      //   at org.jetbrains.java.decompiler.main.rels.MethodProcessor.codeToJava(MethodProcessor.java:166)
      //
      // Bytecode:
      // 000: getstatic client.vh Z
      // 003: istore 13
      // 005: getstatic ua.yb I
      // 008: bipush 1
      // 009: iadd
      // 00a: putstatic ua.yb I
      // 00d: iload 2
      // 00e: istore 12
      // 010: iload 11
      // 012: bipush 102
      // 014: if_icmpeq 018
      // 017: return
      // 018: bipush -1
      // 019: iload 12
      // 01b: bipush -1
      // 01c: ixor
      // 01d: if_icmpge 072
      // 020: aload 4
      // 022: iload 6
      // 024: ldc 930333777
      // 026: ishr
      // 027: iload 8
      // 029: imul
      // 02a: iload 3
      // 02b: ldc 98213361
      // 02d: ishr
      // 02e: iadd
      // 02f: iaload
      // 030: istore 10
      // 032: iload 3
      // 033: iload 9
      // 035: iadd
      // 036: istore 3
      // 037: iload 13
      // 039: ifne 11d
      // 03c: bipush -1
      // 03d: iload 10
      // 03f: bipush -1
      // 040: ixor
      // 041: if_icmpne 054
      // 044: goto 048
      // 047: athrow
      // 048: iinc 7 1
      // 04b: iload 13
      // 04d: ifeq 064
      // 050: goto 054
      // 053: athrow
      // 054: aload 0
      // 055: getfield ua.rb [I
      // 058: iload 7
      // 05a: iinc 7 1
      // 05d: iload 10
      // 05f: iastore
      // 060: goto 064
      // 063: athrow
      // 064: iload 6
      // 066: iload 1
      // 067: iadd
      // 068: istore 6
      // 06a: iinc 12 1
      // 06d: iload 13
      // 06f: ifeq 018
      // 072: goto 11d
      // 075: astore 12
      // 077: aload 12
      // 079: new java/lang/StringBuilder
      // 07c: dup
      // 07d: invokespecial java/lang/StringBuilder.<init> ()V
      // 080: getstatic ua.Yb [Ljava/lang/String;
      // 083: bipush 86
      // 085: aaload
      // 086: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 089: iload 1
      // 08a: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 08d: bipush 44
      // 08f: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 092: iload 2
      // 093: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 096: bipush 44
      // 098: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 09b: iload 3
      // 09c: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 09f: bipush 44
      // 0a1: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 0a4: aload 4
      // 0a6: ifnull 0b2
      // 0a9: getstatic ua.Yb [Ljava/lang/String;
      // 0ac: bipush 1
      // 0ad: aaload
      // 0ae: goto 0b7
      // 0b1: athrow
      // 0b2: getstatic ua.Yb [Ljava/lang/String;
      // 0b5: bipush 0
      // 0b6: aaload
      // 0b7: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 0ba: bipush 44
      // 0bc: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 0bf: aload 5
      // 0c1: ifnull 0cd
      // 0c4: getstatic ua.Yb [Ljava/lang/String;
      // 0c7: bipush 1
      // 0c8: aaload
      // 0c9: goto 0d2
      // 0cc: athrow
      // 0cd: getstatic ua.Yb [Ljava/lang/String;
      // 0d0: bipush 0
      // 0d1: aaload
      // 0d2: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 0d5: bipush 44
      // 0d7: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 0da: iload 6
      // 0dc: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 0df: bipush 44
      // 0e1: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 0e4: iload 7
      // 0e6: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 0e9: bipush 44
      // 0eb: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 0ee: iload 8
      // 0f0: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 0f3: bipush 44
      // 0f5: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 0f8: iload 9
      // 0fa: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 0fd: bipush 44
      // 0ff: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 102: iload 10
      // 104: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 107: bipush 44
      // 109: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 10c: iload 11
      // 10e: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 111: bipush 41
      // 113: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 116: invokevirtual java/lang/StringBuilder.toString ()Ljava/lang/String;
      // 119: invokestatic i.a (Ljava/lang/Throwable;Ljava/lang/String;)Lla;
      // 11c: athrow
      // 11d: return
   }

   private final void a(int var1, int var2, byte[] var3, int var4, boolean var5, int var6, int var7, int var8, int var9, int[] var10, int[] var11, int var12) {
      boolean var18 = client.vh;

      try {
         U++;
         int var25 = -var8 + 256;
         if (!var5) {
            int var14 = -var6;

            while (0 > var14 && !var18) {
               int var15 = -var9;

               int var27;
               int var31;
               while (true) {
                  if (var15 < 0) {
                     int var16 = var3[var1++];

                     label90: {
                        label89: {
                           try {
                              var27 = ~var16;
                              var31 = -1;
                              if (var18) {
                                 break;
                              }

                              if (var27 != -1) {
                                 break label89;
                              }
                           } catch (RuntimeException var23) {
                              throw var23;
                           }

                           try {
                              var12++;
                              if (!var18) {
                                 break label90;
                              }
                           } catch (RuntimeException var22) {
                              throw var22;
                           }
                        }

                        var16 = var11[0xFF & var16];
                        int var17 = var10[var12];
                        var10[var12++] = ib.a(var25 * ib.a(var17, 16711935) + ib.a(16711935, var16) * var8, -16711936)
                              + ib.a(16711680, var8 * ib.a(65280, var16) + var25 * ib.a(65280, var17))
                           >> 1273033224;
                     }

                     var15++;
                     if (!var18) {
                        continue;
                     }
                  }

                  var12 += var2;
                  var1 += var7;
                  var27 = var14;
                  var31 = var4;
                  break;
               }

               var14 = var27 + var31;
               if (var18) {
                  break;
               }
            }
         }
      } catch (RuntimeException var24) {
         RuntimeException var13 = var24;

         RuntimeException var10000;
         StringBuilder var10001;
         String var10002;
         label69: {
            try {
               var10000 = var13;
               var10001 = new StringBuilder().append(Yb[76]).append(var1).append(',').append(var2).append(',');
               if (var3 != null) {
                  var10002 = Yb[1];
                  break label69;
               }
            } catch (RuntimeException var21) {
               throw var21;
            }

            var10002 = Yb[0];
         }

         label62: {
            try {
               var10001 = var10001.append(var10002)
                  .append(',')
                  .append(var4)
                  .append(',')
                  .append(var5)
                  .append(',')
                  .append(var6)
                  .append(',')
                  .append(var7)
                  .append(',')
                  .append(var8)
                  .append(',')
                  .append(var9)
                  .append(',');
               if (var10 != null) {
                  var10002 = Yb[1];
                  break label62;
               }
            } catch (RuntimeException var20) {
               throw var20;
            }

            var10002 = Yb[0];
         }

         try {
            var10001 = var10001.append(var10002).append(',');
            if (var11 != null) {
               throw i.a(var10000, var10001.append(Yb[1]).append(',').append(var12).append(')').toString());
            }
         } catch (RuntimeException var19) {
            throw var19;
         }

         throw i.a(var10000, var10001.append(Yb[0]).append(',').append(var12).append(')').toString());
      }
   }

   private final int c(int var1, int var2) {
      try {
         try {
            o++;
            if (~var2 == -1) {
               return m.b[var2][8] - 2;
            }
         } catch (RuntimeException var5) {
            throw var5;
         }

         try {
            if (var1 < 49) {
               this.a(-22, 77, 112, -35, -44, (int[])null, -45, (int[])null, -39, -33, false, 50, 61, 37, -7);
            }
         } catch (RuntimeException var4) {
            throw var4;
         }

         return m.b[var2][8] + -1;
      } catch (RuntimeException var6) {
         throw i.a(var6, Yb[88] + var1 + 44 + var2 + 41);
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

               label44: {
                  while (~(var1 + var3) < ~var10) {
                     try {
                        this.ob[var5][var8++] = this.rb[var9 + this.u * var10];
                        var10++;
                        if (var11) {
                           break label44;
                        }

                        if (var11) {
                           break;
                        }
                     } catch (RuntimeException var13) {
                        throw var13;
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

         try {
            if (var10000 != -27966) {
               this.a(73, -62, -30, (byte)-113, 44, -64, -91, 100, -79, (int[])null, 117, 11, 127, -109, (int[])null);
            }
         } catch (RuntimeException var12) {
            throw var12;
         }
      } catch (RuntimeException var14) {
         throw i.a(var14, Yb[18] + var1 + ',' + var2 + ',' + var3 + ',' + var4 + ',' + var5 + ',' + var6 + ')');
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

   final void c(int param1, int param2, int param3, int param4, int param5, int param6) {
      // $VF: Couldn't be decompiled
      // Please report this to the Vineflower issue tracker, at https://github.com/Vineflower/vineflower/issues with a copy of the class file (if you have the rights to distribute it!)
      // java.lang.RuntimeException: parsing failure!
      //   at org.jetbrains.java.decompiler.modules.decompiler.decompose.DomHelper.parseGraph(DomHelper.java:211)
      //   at org.jetbrains.java.decompiler.main.rels.MethodProcessor.codeToJava(MethodProcessor.java:166)
      //
      // Bytecode:
      // 000: getstatic client.vh Z
      // 003: istore 25
      // 005: getstatic ua.X I
      // 008: bipush 1
      // 009: iadd
      // 00a: putstatic ua.X I
      // 00d: iload 1
      // 00e: ineg
      // 00f: sipush 256
      // 012: iadd
      // 013: istore 7
      // 015: iload 1
      // 016: sipush 255
      // 019: iload 5
      // 01b: ldc -25949072
      // 01d: ishr
      // 01e: iand
      // 01f: imul
      // 020: istore 8
      // 022: sipush 255
      // 025: iload 5
      // 027: ldc -1057205208
      // 029: ishr
      // 02a: iand
      // 02b: iload 1
      // 02c: imul
      // 02d: istore 9
      // 02f: iload 1
      // 030: iload 5
      // 032: sipush 255
      // 035: iand
      // 036: imul
      // 037: istore 10
      // 039: iload 2
      // 03a: ldc -1057205208
      // 03c: if_icmpeq 040
      // 03f: return
      // 040: iload 3
      // 041: ineg
      // 042: iload 4
      // 044: iadd
      // 045: istore 14
      // 047: iload 14
      // 049: bipush -1
      // 04a: ixor
      // 04b: bipush -1
      // 04c: if_icmpgt 053
      // 04f: goto 056
      // 052: athrow
      // 053: bipush 0
      // 054: istore 14
      // 056: iload 4
      // 058: iload 3
      // 059: ineg
      // 05a: isub
      // 05b: istore 15
      // 05d: aload 0
      // 05e: getfield ua.k I
      // 061: iload 15
      // 063: if_icmpgt 06e
      // 066: bipush -1
      // 067: aload 0
      // 068: getfield ua.k I
      // 06b: iadd
      // 06c: istore 15
      // 06e: bipush 1
      // 06f: istore 16
      // 071: aload 0
      // 072: getfield ua.i Z
      // 075: ifne 07d
      // 078: bipush 1
      // 079: goto 07e
      // 07c: athrow
      // 07d: bipush 0
      // 07e: ifeq 084
      // 081: goto 095
      // 084: bipush 1
      // 085: iload 14
      // 087: iand
      // 088: ifne 08f
      // 08b: goto 092
      // 08e: athrow
      // 08f: iinc 14 1
      // 092: bipush 2
      // 093: istore 16
      // 095: iload 14
      // 097: istore 17
      // 099: iload 15
      // 09b: iload 17
      // 09d: if_icmplt 18e
      // 0a0: iload 17
      // 0a2: iload 4
      // 0a4: ineg
      // 0a5: iadd
      // 0a6: istore 18
      // 0a8: iload 18
      // 0aa: iload 18
      // 0ac: imul
      // 0ad: ineg
      // 0ae: iload 3
      // 0af: iload 3
      // 0b0: imul
      // 0b1: iadd
      // 0b2: i2d
      // 0b3: invokestatic java/lang/Math.sqrt (D)D
      // 0b6: d2i
      // 0b7: istore 19
      // 0b9: iload 6
      // 0bb: iload 19
      // 0bd: ineg
      // 0be: iadd
      // 0bf: istore 20
      // 0c1: iload 25
      // 0c3: ifne 1e5
      // 0c6: iload 20
      // 0c8: bipush -1
      // 0c9: ixor
      // 0ca: bipush -1
      // 0cb: if_icmpgt 0d6
      // 0ce: goto 0d2
      // 0d1: athrow
      // 0d2: goto 0d9
      // 0d5: athrow
      // 0d6: bipush 0
      // 0d7: istore 20
      // 0d9: iload 6
      // 0db: iload 19
      // 0dd: ineg
      // 0de: isub
      // 0df: istore 21
      // 0e1: aload 0
      // 0e2: getfield ua.u I
      // 0e5: iload 21
      // 0e7: if_icmple 0ee
      // 0ea: goto 0f6
      // 0ed: athrow
      // 0ee: aload 0
      // 0ef: getfield ua.u I
      // 0f2: bipush -1
      // 0f3: iadd
      // 0f4: istore 21
      // 0f6: iload 20
      // 0f8: aload 0
      // 0f9: getfield ua.u I
      // 0fc: iload 17
      // 0fe: imul
      // 0ff: iadd
      // 100: istore 22
      // 102: iload 20
      // 104: istore 23
      // 106: iload 21
      // 108: iload 23
      // 10a: if_icmplt 182
      // 10d: aload 0
      // 10e: getfield ua.rb [I
      // 111: iload 22
      // 113: iaload
      // 114: sipush 255
      // 117: iand
      // 118: iload 7
      // 11a: imul
      // 11b: istore 13
      // 11d: iload 7
      // 11f: ldc 65493
      // 121: aload 0
      // 122: getfield ua.rb [I
      // 125: iload 22
      // 127: iaload
      // 128: iand
      // 129: ldc -1460723256
      // 12b: ishr
      // 12c: imul
      // 12d: istore 12
      // 12f: iload 7
      // 131: aload 0
      // 132: getfield ua.rb [I
      // 135: iload 22
      // 137: iaload
      // 138: ldc 16718009
      // 13a: iand
      // 13b: ldc -342059728
      // 13d: ishr
      // 13e: imul
      // 13f: istore 11
      // 141: iload 13
      // 143: iload 10
      // 145: iadd
      // 146: ldc 1268749032
      // 148: ishr
      // 149: iload 9
      // 14b: iload 12
      // 14d: ineg
      // 14e: isub
      // 14f: ldc -404063160
      // 151: ishr
      // 152: ldc -187266712
      // 154: ishl
      // 155: iload 8
      // 157: iload 11
      // 159: ineg
      // 15a: isub
      // 15b: ldc 296955496
      // 15d: ishr
      // 15e: ldc 1872374416
      // 160: ishl
      // 161: iadd
      // 162: iadd
      // 163: istore 24
      // 165: aload 0
      // 166: getfield ua.rb [I
      // 169: iload 22
      // 16b: iinc 22 1
      // 16e: iload 24
      // 170: iastore
      // 171: iinc 23 1
      // 174: iload 25
      // 176: ifne 189
      // 179: iload 25
      // 17b: ifeq 106
      // 17e: goto 182
      // 181: athrow
      // 182: iload 17
      // 184: iload 16
      // 186: iadd
      // 187: istore 17
      // 189: iload 25
      // 18b: ifeq 099
      // 18e: goto 1e5
      // 191: astore 7
      // 193: aload 7
      // 195: new java/lang/StringBuilder
      // 198: dup
      // 199: invokespecial java/lang/StringBuilder.<init> ()V
      // 19c: getstatic ua.Yb [Ljava/lang/String;
      // 19f: bipush 75
      // 1a1: aaload
      // 1a2: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 1a5: iload 1
      // 1a6: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 1a9: bipush 44
      // 1ab: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 1ae: iload 2
      // 1af: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 1b2: bipush 44
      // 1b4: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 1b7: iload 3
      // 1b8: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 1bb: bipush 44
      // 1bd: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 1c0: iload 4
      // 1c2: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 1c5: bipush 44
      // 1c7: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 1ca: iload 5
      // 1cc: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 1cf: bipush 44
      // 1d1: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 1d4: iload 6
      // 1d6: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 1d9: bipush 41
      // 1db: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 1de: invokevirtual java/lang/StringBuilder.toString ()Ljava/lang/String;
      // 1e1: invokestatic i.a (Ljava/lang/Throwable;Ljava/lang/String;)Lla;
      // 1e4: athrow
      // 1e5: return
   }

   private final void a(
      int[] param1,
      int[] param2,
      int param3,
      int param4,
      int param5,
      int param6,
      int param7,
      int param8,
      int param9,
      int param10,
      int param11,
      int param12,
      int param13,
      int param14,
      int param15,
      int param16,
      int param17
   ) {
      // $VF: Couldn't be decompiled
      // Please report this to the Vineflower issue tracker, at https://github.com/Vineflower/vineflower/issues with a copy of the class file (if you have the rights to distribute it!)
      // java.lang.RuntimeException: parsing failure!
      //   at org.jetbrains.java.decompiler.modules.decompiler.decompose.DomHelper.parseGraph(DomHelper.java:211)
      //   at org.jetbrains.java.decompiler.main.rels.MethodProcessor.codeToJava(MethodProcessor.java:166)
      //
      // Bytecode:
      // 000: getstatic client.vh Z
      // 003: istore 33
      // 005: getstatic ua.a I
      // 008: bipush 1
      // 009: iadd
      // 00a: putstatic ua.a I
      // 00d: iload 15
      // 00f: ldc 1196549680
      // 011: ishr
      // 012: sipush 255
      // 015: iand
      // 016: istore 21
      // 018: iload 15
      // 01a: ldc -340653432
      // 01c: ishr
      // 01d: sipush 255
      // 020: iand
      // 021: istore 22
      // 023: iload 15
      // 025: sipush 255
      // 028: iand
      // 029: istore 23
      // 02b: iload 8
      // 02d: ldc -922169040
      // 02f: ishr
      // 030: sipush 255
      // 033: iand
      // 034: istore 24
      // 036: iload 6
      // 038: ldc 1603920392
      // 03a: if_icmpeq 047
      // 03d: aload 0
      // 03e: bipush 29
      // 040: putfield ua.Rb I
      // 043: goto 047
      // 046: athrow
      // 047: iload 8
      // 049: ldc 5163400
      // 04b: ishr
      // 04c: sipush 255
      // 04f: iand
      // 050: istore 25
      // 052: sipush 255
      // 055: iload 8
      // 057: iand
      // 058: istore 26
      // 05a: iload 11
      // 05c: istore 27
      // 05e: iload 16
      // 060: ineg
      // 061: istore 28
      // 063: bipush -1
      // 064: iload 28
      // 066: bipush -1
      // 067: ixor
      // 068: if_icmpge 209
      // 06b: iload 13
      // 06d: ldc -1328879408
      // 06f: ishr
      // 070: iload 14
      // 072: imul
      // 073: istore 29
      // 075: iload 5
      // 077: ldc -299524176
      // 079: ishr
      // 07a: istore 30
      // 07c: iload 3
      // 07d: istore 31
      // 07f: iload 33
      // 081: ifne 21a
      // 084: iload 30
      // 086: bipush -1
      // 087: ixor
      // 088: aload 0
      // 089: getfield ua.hb I
      // 08c: bipush -1
      // 08d: ixor
      // 08e: if_icmple 0b6
      // 091: goto 095
      // 094: athrow
      // 095: aload 0
      // 096: getfield ua.hb I
      // 099: iload 30
      // 09b: ineg
      // 09c: iadd
      // 09d: istore 32
      // 09f: iload 31
      // 0a1: iload 32
      // 0a3: isub
      // 0a4: istore 31
      // 0a6: iload 11
      // 0a8: iload 32
      // 0aa: iload 10
      // 0ac: imul
      // 0ad: iadd
      // 0ae: istore 11
      // 0b0: aload 0
      // 0b1: getfield ua.hb I
      // 0b4: istore 30
      // 0b6: aload 0
      // 0b7: getfield ua.lb I
      // 0ba: iload 30
      // 0bc: iload 31
      // 0be: ineg
      // 0bf: isub
      // 0c0: if_icmple 0c7
      // 0c3: goto 0dc
      // 0c6: athrow
      // 0c7: aload 0
      // 0c8: getfield ua.lb I
      // 0cb: ineg
      // 0cc: iload 30
      // 0ce: iadd
      // 0cf: iload 31
      // 0d1: ineg
      // 0d2: isub
      // 0d3: istore 32
      // 0d5: iload 31
      // 0d7: iload 32
      // 0d9: isub
      // 0da: istore 31
      // 0dc: bipush 1
      // 0dd: iload 12
      // 0df: isub
      // 0e0: istore 12
      // 0e2: bipush 0
      // 0e3: iload 12
      // 0e5: if_icmpne 0ec
      // 0e8: goto 1e6
      // 0eb: athrow
      // 0ec: iload 30
      // 0ee: istore 32
      // 0f0: iload 30
      // 0f2: iload 31
      // 0f4: iadd
      // 0f5: iload 32
      // 0f7: if_icmple 1e6
      // 0fa: aload 2
      // 0fb: iload 29
      // 0fd: iload 11
      // 0ff: ldc -1009344688
      // 101: ishr
      // 102: iadd
      // 103: iaload
      // 104: istore 7
      // 106: bipush -1
      // 107: iload 7
      // 109: bipush -1
      // 10a: ixor
      // 10b: iload 33
      // 10d: ifne 1fe
      // 110: if_icmpne 11a
      // 113: goto 117
      // 116: athrow
      // 117: goto 1d7
      // 11a: sipush 255
      // 11d: iload 7
      // 11f: ldc 2008995472
      // 121: ishr
      // 122: iand
      // 123: istore 18
      // 125: sipush 255
      // 128: iload 7
      // 12a: ldc 1043792552
      // 12c: ishr
      // 12d: iand
      // 12e: istore 19
      // 130: iload 7
      // 132: sipush 255
      // 135: iand
      // 136: istore 20
      // 138: iload 18
      // 13a: iload 19
      // 13c: if_icmpne 17b
      // 13f: iload 20
      // 141: iload 19
      // 143: if_icmpne 17b
      // 146: goto 14a
      // 149: athrow
      // 14a: aload 1
      // 14b: iload 32
      // 14d: iload 17
      // 14f: ineg
      // 150: isub
      // 151: iload 20
      // 153: iload 23
      // 155: imul
      // 156: ldc -1147526104
      // 158: ishr
      // 159: iload 22
      // 15b: iload 19
      // 15d: imul
      // 15e: ldc 1601552776
      // 160: ishr
      // 161: ldc -566695064
      // 163: ishl
      // 164: iload 18
      // 166: iload 21
      // 168: imul
      // 169: ldc 1603920392
      // 16b: ishr
      // 16c: ldc -538385168
      // 16e: ishl
      // 16f: iadd
      // 170: iadd
      // 171: iastore
      // 172: iload 33
      // 174: ifeq 1d7
      // 177: goto 17b
      // 17a: athrow
      // 17b: sipush -256
      // 17e: iload 18
      // 180: bipush -1
      // 181: ixor
      // 182: if_icmpne 198
      // 185: goto 189
      // 188: athrow
      // 189: iload 20
      // 18b: bipush -1
      // 18c: ixor
      // 18d: iload 19
      // 18f: bipush -1
      // 190: ixor
      // 191: if_icmpeq 1aa
      // 194: goto 198
      // 197: athrow
      // 198: aload 1
      // 199: iload 32
      // 19b: iload 17
      // 19d: iadd
      // 19e: iload 7
      // 1a0: iastore
      // 1a1: iload 33
      // 1a3: ifeq 1d7
      // 1a6: goto 1aa
      // 1a9: athrow
      // 1aa: aload 1
      // 1ab: iload 32
      // 1ad: iload 17
      // 1af: ineg
      // 1b0: isub
      // 1b1: iload 26
      // 1b3: iload 20
      // 1b5: imul
      // 1b6: ldc -1345290488
      // 1b8: ishr
      // 1b9: iload 18
      // 1bb: iload 24
      // 1bd: imul
      // 1be: ldc 1685589832
      // 1c0: ishr
      // 1c1: ldc -1431216016
      // 1c3: ishl
      // 1c4: iadd
      // 1c5: iload 19
      // 1c7: iload 25
      // 1c9: imul
      // 1ca: ldc 508305352
      // 1cc: ishr
      // 1cd: ldc -483710840
      // 1cf: ishl
      // 1d0: ineg
      // 1d1: isub
      // 1d2: iastore
      // 1d3: goto 1d7
      // 1d6: athrow
      // 1d7: iload 11
      // 1d9: iload 10
      // 1db: iadd
      // 1dc: istore 11
      // 1de: iinc 32 1
      // 1e1: iload 33
      // 1e3: ifeq 0f0
      // 1e6: iload 13
      // 1e8: iload 9
      // 1ea: iadd
      // 1eb: istore 13
      // 1ed: iload 27
      // 1ef: istore 11
      // 1f1: iload 17
      // 1f3: aload 0
      // 1f4: getfield ua.u I
      // 1f7: iadd
      // 1f8: istore 17
      // 1fa: iload 5
      // 1fc: iload 4
      // 1fe: iadd
      // 1ff: istore 5
      // 201: iinc 28 1
      // 204: iload 33
      // 206: ifeq 063
      // 209: goto 21a
      // 20c: astore 27
      // 20e: getstatic java/lang/System.out Ljava/io/PrintStream;
      // 211: getstatic ua.Yb [Ljava/lang/String;
      // 214: bipush 65
      // 216: aaload
      // 217: invokevirtual java/io/PrintStream.println (Ljava/lang/String;)V
      // 21a: goto 301
      // 21d: astore 18
      // 21f: aload 18
      // 221: new java/lang/StringBuilder
      // 224: dup
      // 225: invokespecial java/lang/StringBuilder.<init> ()V
      // 228: getstatic ua.Yb [Ljava/lang/String;
      // 22b: bipush 66
      // 22d: aaload
      // 22e: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 231: aload 1
      // 232: ifnull 23e
      // 235: getstatic ua.Yb [Ljava/lang/String;
      // 238: bipush 1
      // 239: aaload
      // 23a: goto 243
      // 23d: athrow
      // 23e: getstatic ua.Yb [Ljava/lang/String;
      // 241: bipush 0
      // 242: aaload
      // 243: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 246: bipush 44
      // 248: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 24b: aload 2
      // 24c: ifnull 258
      // 24f: getstatic ua.Yb [Ljava/lang/String;
      // 252: bipush 1
      // 253: aaload
      // 254: goto 25d
      // 257: athrow
      // 258: getstatic ua.Yb [Ljava/lang/String;
      // 25b: bipush 0
      // 25c: aaload
      // 25d: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 260: bipush 44
      // 262: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 265: iload 3
      // 266: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 269: bipush 44
      // 26b: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 26e: iload 4
      // 270: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 273: bipush 44
      // 275: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 278: iload 5
      // 27a: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 27d: bipush 44
      // 27f: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 282: iload 6
      // 284: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 287: bipush 44
      // 289: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 28c: iload 7
      // 28e: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 291: bipush 44
      // 293: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 296: iload 8
      // 298: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 29b: bipush 44
      // 29d: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 2a0: iload 9
      // 2a2: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 2a5: bipush 44
      // 2a7: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 2aa: iload 10
      // 2ac: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 2af: bipush 44
      // 2b1: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 2b4: iload 11
      // 2b6: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 2b9: bipush 44
      // 2bb: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 2be: iload 12
      // 2c0: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 2c3: bipush 44
      // 2c5: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 2c8: iload 13
      // 2ca: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 2cd: bipush 44
      // 2cf: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 2d2: iload 14
      // 2d4: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 2d7: bipush 44
      // 2d9: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 2dc: iload 15
      // 2de: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 2e1: bipush 44
      // 2e3: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 2e6: iload 16
      // 2e8: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 2eb: bipush 44
      // 2ed: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 2f0: iload 17
      // 2f2: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 2f5: bipush 41
      // 2f7: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 2fa: invokevirtual java/lang/StringBuilder.toString ()Ljava/lang/String;
      // 2fd: invokestatic i.a (Ljava/lang/Throwable;Ljava/lang/String;)Lla;
      // 300: athrow
      // 301: return
   }

   private final void a(
      int param1,
      int[] param2,
      int param3,
      int param4,
      int param5,
      int param6,
      int param7,
      int param8,
      byte[] param9,
      int param10,
      byte param11,
      int param12,
      int param13,
      int param14,
      int param15,
      int param16,
      int[] param17,
      int param18
   ) {
      // $VF: Couldn't be decompiled
      // Please report this to the Vineflower issue tracker, at https://github.com/Vineflower/vineflower/issues with a copy of the class file (if you have the rights to distribute it!)
      // java.lang.RuntimeException: parsing failure!
      //   at org.jetbrains.java.decompiler.modules.decompiler.decompose.DomHelper.parseGraph(DomHelper.java:211)
      //   at org.jetbrains.java.decompiler.main.rels.MethodProcessor.codeToJava(MethodProcessor.java:166)
      //
      // Bytecode:
      // 000: getstatic client.vh Z
      // 003: istore 34
      // 005: getstatic ua.J I
      // 008: bipush 1
      // 009: iadd
      // 00a: putstatic ua.J I
      // 00d: iload 18
      // 00f: ldc 1698138864
      // 011: ishr
      // 012: sipush 255
      // 015: iand
      // 016: istore 22
      // 018: iload 11
      // 01a: bipush 8
      // 01c: if_icmpgt 020
      // 01f: return
      // 020: iload 18
      // 022: ldc 65381
      // 024: iand
      // 025: ldc 1758159464
      // 027: ishr
      // 028: istore 23
      // 02a: iload 18
      // 02c: sipush 255
      // 02f: iand
      // 030: istore 24
      // 032: iload 13
      // 034: ldc 34872048
      // 036: ishr
      // 037: sipush 255
      // 03a: iand
      // 03b: istore 25
      // 03d: sipush 255
      // 040: iload 13
      // 042: ldc 225029096
      // 044: ishr
      // 045: iand
      // 046: istore 26
      // 048: sipush 255
      // 04b: iload 13
      // 04d: iand
      // 04e: istore 27
      // 050: iload 12
      // 052: istore 28
      // 054: iload 1
      // 055: ineg
      // 056: istore 29
      // 058: bipush 0
      // 059: iload 29
      // 05b: if_icmple 203
      // 05e: iload 3
      // 05f: iload 8
      // 061: ldc -1095936688
      // 063: ishr
      // 064: imul
      // 065: istore 30
      // 067: iload 14
      // 069: ldc -2034604816
      // 06b: ishr
      // 06c: istore 31
      // 06e: iload 16
      // 070: istore 32
      // 072: iload 34
      // 074: ifne 214
      // 077: iload 31
      // 079: bipush -1
      // 07a: ixor
      // 07b: aload 0
      // 07c: getfield ua.hb I
      // 07f: bipush -1
      // 080: ixor
      // 081: if_icmpgt 08c
      // 084: goto 088
      // 087: athrow
      // 088: goto 0ad
      // 08b: athrow
      // 08c: aload 0
      // 08d: getfield ua.hb I
      // 090: iload 31
      // 092: ineg
      // 093: iadd
      // 094: istore 33
      // 096: iload 12
      // 098: iload 33
      // 09a: iload 15
      // 09c: imul
      // 09d: iadd
      // 09e: istore 12
      // 0a0: iload 32
      // 0a2: iload 33
      // 0a4: isub
      // 0a5: istore 32
      // 0a7: aload 0
      // 0a8: getfield ua.hb I
      // 0ab: istore 31
      // 0ad: iload 10
      // 0af: ineg
      // 0b0: bipush 1
      // 0b1: iadd
      // 0b2: istore 10
      // 0b4: iload 32
      // 0b6: iload 31
      // 0b8: iadd
      // 0b9: aload 0
      // 0ba: getfield ua.lb I
      // 0bd: if_icmpge 0c4
      // 0c0: goto 0d7
      // 0c3: athrow
      // 0c4: iload 31
      // 0c6: iload 32
      // 0c8: aload 0
      // 0c9: getfield ua.lb I
      // 0cc: isub
      // 0cd: iadd
      // 0ce: istore 33
      // 0d0: iload 32
      // 0d2: iload 33
      // 0d4: isub
      // 0d5: istore 32
      // 0d7: iload 10
      // 0d9: ifne 0e0
      // 0dc: goto 1e0
      // 0df: athrow
      // 0e0: iload 31
      // 0e2: istore 33
      // 0e4: iload 33
      // 0e6: iload 32
      // 0e8: iload 31
      // 0ea: iadd
      // 0eb: if_icmpge 1e0
      // 0ee: sipush 255
      // 0f1: aload 9
      // 0f3: iload 12
      // 0f5: ldc -91857424
      // 0f7: ishr
      // 0f8: iload 30
      // 0fa: ineg
      // 0fb: isub
      // 0fc: baload
      // 0fd: iand
      // 0fe: istore 5
      // 100: iload 5
      // 102: iload 34
      // 104: ifne 1f9
      // 107: ifeq 1d1
      // 10a: goto 10e
      // 10d: athrow
      // 10e: aload 2
      // 10f: iload 5
      // 111: iaload
      // 112: istore 5
      // 114: iload 5
      // 116: sipush 255
      // 119: iand
      // 11a: istore 21
      // 11c: sipush 255
      // 11f: iload 5
      // 121: ldc 1941812904
      // 123: ishr
      // 124: iand
      // 125: istore 20
      // 127: sipush 255
      // 12a: iload 5
      // 12c: ldc -309996368
      // 12e: ishr
      // 12f: iand
      // 130: istore 19
      // 132: iload 20
      // 134: iload 19
      // 136: if_icmpne 148
      // 139: iload 20
      // 13b: bipush -1
      // 13c: ixor
      // 13d: iload 21
      // 13f: bipush -1
      // 140: ixor
      // 141: if_icmpeq 1a5
      // 144: goto 148
      // 147: athrow
      // 148: sipush -256
      // 14b: iload 19
      // 14d: bipush -1
      // 14e: ixor
      // 14f: if_icmpne 192
      // 152: goto 156
      // 155: athrow
      // 156: iload 20
      // 158: iload 21
      // 15a: if_icmpne 192
      // 15d: goto 161
      // 160: athrow
      // 161: aload 17
      // 163: iload 7
      // 165: iload 33
      // 167: iadd
      // 168: iload 25
      // 16a: iload 19
      // 16c: imul
      // 16d: ldc 1700051816
      // 16f: ishr
      // 170: ldc -1733973648
      // 172: ishl
      // 173: iload 20
      // 175: iload 26
      // 177: imul
      // 178: ldc 269461512
      // 17a: ishr
      // 17b: ldc 971885320
      // 17d: ishl
      // 17e: iadd
      // 17f: iload 27
      // 181: iload 21
      // 183: imul
      // 184: ldc -1282776696
      // 186: ishr
      // 187: iadd
      // 188: iastore
      // 189: iload 34
      // 18b: ifeq 1d1
      // 18e: goto 192
      // 191: athrow
      // 192: aload 17
      // 194: iload 33
      // 196: iload 7
      // 198: iadd
      // 199: iload 5
      // 19b: iastore
      // 19c: iload 34
      // 19e: ifeq 1d1
      // 1a1: goto 1a5
      // 1a4: athrow
      // 1a5: aload 17
      // 1a7: iload 33
      // 1a9: iload 7
      // 1ab: iadd
      // 1ac: iload 24
      // 1ae: iload 21
      // 1b0: imul
      // 1b1: ldc -178220952
      // 1b3: ishr
      // 1b4: iload 19
      // 1b6: iload 22
      // 1b8: imul
      // 1b9: ldc -1022678840
      // 1bb: ishr
      // 1bc: ldc 2122310768
      // 1be: ishl
      // 1bf: iadd
      // 1c0: iload 23
      // 1c2: iload 20
      // 1c4: imul
      // 1c5: ldc 901102216
      // 1c7: ishr
      // 1c8: ldc -552193592
      // 1ca: ishl
      // 1cb: iadd
      // 1cc: iastore
      // 1cd: goto 1d1
      // 1d0: athrow
      // 1d1: iload 12
      // 1d3: iload 15
      // 1d5: iadd
      // 1d6: istore 12
      // 1d8: iinc 33 1
      // 1db: iload 34
      // 1dd: ifeq 0e4
      // 1e0: iload 8
      // 1e2: iload 4
      // 1e4: iadd
      // 1e5: istore 8
      // 1e7: iload 14
      // 1e9: iload 6
      // 1eb: iadd
      // 1ec: istore 14
      // 1ee: iload 28
      // 1f0: istore 12
      // 1f2: iload 7
      // 1f4: aload 0
      // 1f5: getfield ua.u I
      // 1f8: iadd
      // 1f9: istore 7
      // 1fb: iinc 29 1
      // 1fe: iload 34
      // 200: ifeq 058
      // 203: goto 214
      // 206: astore 28
      // 208: getstatic java/lang/System.out Ljava/io/PrintStream;
      // 20b: getstatic ua.Yb [Ljava/lang/String;
      // 20e: bipush 65
      // 210: aaload
      // 211: invokevirtual java/io/PrintStream.println (Ljava/lang/String;)V
      // 214: goto 316
      // 217: astore 19
      // 219: aload 19
      // 21b: new java/lang/StringBuilder
      // 21e: dup
      // 21f: invokespecial java/lang/StringBuilder.<init> ()V
      // 222: getstatic ua.Yb [Ljava/lang/String;
      // 225: bipush 85
      // 227: aaload
      // 228: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 22b: iload 1
      // 22c: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 22f: bipush 44
      // 231: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 234: aload 2
      // 235: ifnull 241
      // 238: getstatic ua.Yb [Ljava/lang/String;
      // 23b: bipush 1
      // 23c: aaload
      // 23d: goto 246
      // 240: athrow
      // 241: getstatic ua.Yb [Ljava/lang/String;
      // 244: bipush 0
      // 245: aaload
      // 246: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 249: bipush 44
      // 24b: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 24e: iload 3
      // 24f: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 252: bipush 44
      // 254: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 257: iload 4
      // 259: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 25c: bipush 44
      // 25e: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 261: iload 5
      // 263: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 266: bipush 44
      // 268: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 26b: iload 6
      // 26d: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 270: bipush 44
      // 272: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 275: iload 7
      // 277: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 27a: bipush 44
      // 27c: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 27f: iload 8
      // 281: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 284: bipush 44
      // 286: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 289: aload 9
      // 28b: ifnull 297
      // 28e: getstatic ua.Yb [Ljava/lang/String;
      // 291: bipush 1
      // 292: aaload
      // 293: goto 29c
      // 296: athrow
      // 297: getstatic ua.Yb [Ljava/lang/String;
      // 29a: bipush 0
      // 29b: aaload
      // 29c: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 29f: bipush 44
      // 2a1: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 2a4: iload 10
      // 2a6: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 2a9: bipush 44
      // 2ab: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 2ae: iload 11
      // 2b0: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 2b3: bipush 44
      // 2b5: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 2b8: iload 12
      // 2ba: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 2bd: bipush 44
      // 2bf: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 2c2: iload 13
      // 2c4: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 2c7: bipush 44
      // 2c9: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 2cc: iload 14
      // 2ce: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 2d1: bipush 44
      // 2d3: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 2d6: iload 15
      // 2d8: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 2db: bipush 44
      // 2dd: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 2e0: iload 16
      // 2e2: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 2e5: bipush 44
      // 2e7: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 2ea: aload 17
      // 2ec: ifnull 2f8
      // 2ef: getstatic ua.Yb [Ljava/lang/String;
      // 2f2: bipush 1
      // 2f3: aaload
      // 2f4: goto 2fd
      // 2f7: athrow
      // 2f8: getstatic ua.Yb [Ljava/lang/String;
      // 2fb: bipush 0
      // 2fc: aaload
      // 2fd: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 300: bipush 44
      // 302: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 305: iload 18
      // 307: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 30a: bipush 41
      // 30c: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 30f: invokevirtual java/lang/StringBuilder.toString ()Ljava/lang/String;
      // 312: invokestatic i.a (Ljava/lang/Throwable;Ljava/lang/String;)Lla;
      // 315: athrow
      // 316: return
   }

   private final void a(byte param1, boolean param2, byte[] param3, int param4, int param5, int param6, int param7) {
      // $VF: Couldn't be decompiled
      // Please report this to the Vineflower issue tracker, at https://github.com/Vineflower/vineflower/issues with a copy of the class file (if you have the rights to distribute it!)
      // java.lang.RuntimeException: parsing failure!
      //   at org.jetbrains.java.decompiler.modules.decompiler.decompose.DomHelper.parseGraph(DomHelper.java:211)
      //   at org.jetbrains.java.decompiler.main.rels.MethodProcessor.codeToJava(MethodProcessor.java:166)
      //
      // Bytecode:
      // 000: getstatic ua.P I
      // 003: bipush 1
      // 004: iadd
      // 005: putstatic ua.P I
      // 008: iload 4
      // 00a: aload 3
      // 00b: bipush 5
      // 00c: iload 6
      // 00e: iadd
      // 00f: baload
      // 010: ineg
      // 011: isub
      // 012: istore 8
      // 014: iload 1
      // 015: bipush 24
      // 017: if_icmpge 01b
      // 01a: return
      // 01b: iload 7
      // 01d: aload 3
      // 01e: bipush 6
      // 020: iload 6
      // 022: iadd
      // 023: baload
      // 024: ineg
      // 025: iadd
      // 026: istore 9
      // 028: aload 3
      // 029: iload 6
      // 02b: bipush -3
      // 02d: isub
      // 02e: baload
      // 02f: istore 10
      // 031: aload 3
      // 032: bipush 4
      // 033: iload 6
      // 035: iadd
      // 036: baload
      // 037: istore 11
      // 039: sipush 16384
      // 03c: aload 3
      // 03d: iload 6
      // 03f: baload
      // 040: imul
      // 041: aload 3
      // 042: iload 6
      // 044: bipush -1
      // 045: isub
      // 046: baload
      // 047: sipush 128
      // 04a: imul
      // 04b: iadd
      // 04c: aload 3
      // 04d: iload 6
      // 04f: bipush 2
      // 050: iadd
      // 051: baload
      // 052: iadd
      // 053: istore 12
      // 055: iload 8
      // 057: iload 9
      // 059: aload 0
      // 05a: getfield ua.u I
      // 05d: imul
      // 05e: ineg
      // 05f: isub
      // 060: istore 13
      // 062: aload 0
      // 063: getfield ua.u I
      // 066: iload 10
      // 068: ineg
      // 069: iadd
      // 06a: istore 14
      // 06c: aload 0
      // 06d: getfield ua.A I
      // 070: bipush -1
      // 071: ixor
      // 072: iload 9
      // 074: bipush -1
      // 075: ixor
      // 076: if_icmplt 07d
      // 079: goto 0a9
      // 07c: athrow
      // 07d: aload 0
      // 07e: getfield ua.A I
      // 081: iload 9
      // 083: isub
      // 084: istore 16
      // 086: iload 12
      // 088: iload 16
      // 08a: iload 10
      // 08c: imul
      // 08d: iadd
      // 08e: istore 12
      // 090: iload 13
      // 092: aload 0
      // 093: getfield ua.u I
      // 096: iload 16
      // 098: imul
      // 099: iadd
      // 09a: istore 13
      // 09c: iload 11
      // 09e: iload 16
      // 0a0: isub
      // 0a1: istore 11
      // 0a3: aload 0
      // 0a4: getfield ua.A I
      // 0a7: istore 9
      // 0a9: bipush 0
      // 0aa: istore 15
      // 0ac: aload 0
      // 0ad: getfield ua.Rb I
      // 0b0: bipush -1
      // 0b1: ixor
      // 0b2: iload 9
      // 0b4: iload 11
      // 0b6: iadd
      // 0b7: bipush -1
      // 0b8: ixor
      // 0b9: if_icmpge 0c0
      // 0bc: goto 0d2
      // 0bf: athrow
      // 0c0: iload 11
      // 0c2: bipush 1
      // 0c3: iload 9
      // 0c5: iload 11
      // 0c7: iadd
      // 0c8: iadd
      // 0c9: aload 0
      // 0ca: getfield ua.Rb I
      // 0cd: ineg
      // 0ce: iadd
      // 0cf: isub
      // 0d0: istore 11
      // 0d2: iload 8
      // 0d4: bipush -1
      // 0d5: ixor
      // 0d6: aload 0
      // 0d7: getfield ua.hb I
      // 0da: bipush -1
      // 0db: ixor
      // 0dc: if_icmpgt 0e3
      // 0df: goto 116
      // 0e2: athrow
      // 0e3: iload 8
      // 0e5: ineg
      // 0e6: aload 0
      // 0e7: getfield ua.hb I
      // 0ea: iadd
      // 0eb: istore 16
      // 0ed: iload 15
      // 0ef: iload 16
      // 0f1: iadd
      // 0f2: istore 15
      // 0f4: iload 10
      // 0f6: iload 16
      // 0f8: isub
      // 0f9: istore 10
      // 0fb: iload 12
      // 0fd: iload 16
      // 0ff: iadd
      // 100: istore 12
      // 102: aload 0
      // 103: getfield ua.hb I
      // 106: istore 8
      // 108: iload 14
      // 10a: iload 16
      // 10c: iadd
      // 10d: istore 14
      // 10f: iload 13
      // 111: iload 16
      // 113: iadd
      // 114: istore 13
      // 116: iload 10
      // 118: iload 8
      // 11a: iadd
      // 11b: bipush -1
      // 11c: ixor
      // 11d: aload 0
      // 11e: getfield ua.lb I
      // 121: bipush -1
      // 122: ixor
      // 123: if_icmpgt 14a
      // 126: aload 0
      // 127: getfield ua.lb I
      // 12a: ineg
      // 12b: iload 10
      // 12d: iload 8
      // 12f: iadd
      // 130: iadd
      // 131: bipush 1
      // 132: iadd
      // 133: istore 16
      // 135: iload 14
      // 137: iload 16
      // 139: iadd
      // 13a: istore 14
      // 13c: iload 15
      // 13e: iload 16
      // 140: iadd
      // 141: istore 15
      // 143: iload 10
      // 145: iload 16
      // 147: isub
      // 148: istore 10
      // 14a: bipush 0
      // 14b: iload 10
      // 14d: if_icmpge 1a4
      // 150: bipush -1
      // 151: iload 11
      // 153: bipush -1
      // 154: ixor
      // 155: if_icmpgt 160
      // 158: goto 15c
      // 15b: athrow
      // 15c: goto 1a4
      // 15f: athrow
      // 160: iload 2
      // 161: ifeq 187
      // 164: aload 0
      // 165: aload 3
      // 166: iload 5
      // 168: iload 10
      // 16a: iload 13
      // 16c: iload 11
      // 16e: iload 15
      // 170: ldc 1504725224
      // 172: iload 14
      // 174: aload 0
      // 175: getfield ua.rb [I
      // 178: iload 12
      // 17a: invokespecial ua.a ([BIIIIIII[II)V
      // 17d: getstatic client.vh Z
      // 180: ifeq 1a4
      // 183: goto 187
      // 186: athrow
      // 187: aload 0
      // 188: iload 5
      // 18a: aload 0
      // 18b: getfield ua.rb [I
      // 18e: iload 13
      // 190: bipush 37
      // 192: iload 14
      // 194: iload 11
      // 196: iload 10
      // 198: iload 12
      // 19a: aload 3
      // 19b: iload 15
      // 19d: invokespecial ua.a (I[IIBIIII[BI)V
      // 1a0: goto 1a4
      // 1a3: athrow
      // 1a4: goto 216
      // 1a7: astore 8
      // 1a9: aload 8
      // 1ab: new java/lang/StringBuilder
      // 1ae: dup
      // 1af: invokespecial java/lang/StringBuilder.<init> ()V
      // 1b2: getstatic ua.Yb [Ljava/lang/String;
      // 1b5: bipush 60
      // 1b7: aaload
      // 1b8: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 1bb: iload 1
      // 1bc: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 1bf: bipush 44
      // 1c1: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 1c4: iload 2
      // 1c5: invokevirtual java/lang/StringBuilder.append (Z)Ljava/lang/StringBuilder;
      // 1c8: bipush 44
      // 1ca: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 1cd: aload 3
      // 1ce: ifnull 1da
      // 1d1: getstatic ua.Yb [Ljava/lang/String;
      // 1d4: bipush 1
      // 1d5: aaload
      // 1d6: goto 1df
      // 1d9: athrow
      // 1da: getstatic ua.Yb [Ljava/lang/String;
      // 1dd: bipush 0
      // 1de: aaload
      // 1df: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 1e2: bipush 44
      // 1e4: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 1e7: iload 4
      // 1e9: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 1ec: bipush 44
      // 1ee: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 1f1: iload 5
      // 1f3: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 1f6: bipush 44
      // 1f8: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 1fb: iload 6
      // 1fd: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 200: bipush 44
      // 202: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 205: iload 7
      // 207: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 20a: bipush 41
      // 20c: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 20f: invokevirtual java/lang/StringBuilder.toString ()Ljava/lang/String;
      // 212: invokestatic i.a (Ljava/lang/Throwable;Ljava/lang/String;)Lla;
      // 215: athrow
      // 216: return
   }

   final void f(int param1, int param2, int param3, int param4, int param5, int param6) {
      // $VF: Couldn't be decompiled
      // Please report this to the Vineflower issue tracker, at https://github.com/Vineflower/vineflower/issues with a copy of the class file (if you have the rights to distribute it!)
      // java.lang.RuntimeException: parsing failure!
      //   at org.jetbrains.java.decompiler.modules.decompiler.decompose.DomHelper.parseGraph(DomHelper.java:211)
      //   at org.jetbrains.java.decompiler.main.rels.MethodProcessor.codeToJava(MethodProcessor.java:166)
      //
      // Bytecode:
      // 000: getstatic ua.cb I
      // 003: bipush 1
      // 004: iadd
      // 005: putstatic ua.cb I
      // 008: aload 0
      // 009: getfield ua.kb [I
      // 00c: iload 6
      // 00e: iaload
      // 00f: istore 7
      // 011: aload 0
      // 012: getfield ua.R [I
      // 015: iload 6
      // 017: iaload
      // 018: istore 8
      // 01a: bipush 0
      // 01b: istore 9
      // 01d: bipush 0
      // 01e: istore 10
      // 020: iload 7
      // 022: ldc -311879024
      // 024: ishl
      // 025: iload 4
      // 027: idiv
      // 028: istore 11
      // 02a: iload 8
      // 02c: ldc -1693975408
      // 02e: ishl
      // 02f: iload 3
      // 030: idiv
      // 031: istore 12
      // 033: aload 0
      // 034: getfield ua.Qb [Z
      // 037: iload 6
      // 039: baload
      // 03a: ifne 041
      // 03d: goto 125
      // 040: athrow
      // 041: aload 0
      // 042: getfield ua.Eb [I
      // 045: iload 6
      // 047: iaload
      // 048: istore 13
      // 04a: aload 0
      // 04b: getfield ua.qb [I
      // 04e: iload 6
      // 050: iaload
      // 051: istore 14
      // 053: iload 13
      // 055: bipush -1
      // 056: ixor
      // 057: bipush -1
      // 058: if_icmpeq 068
      // 05b: iload 14
      // 05d: ifeq 068
      // 060: goto 064
      // 063: athrow
      // 064: goto 069
      // 067: athrow
      // 068: return
      // 069: bipush -1
      // 06a: aload 0
      // 06b: getfield ua.G [I
      // 06e: iload 6
      // 070: iaload
      // 071: iload 3
      // 072: imul
      // 073: iload 14
      // 075: irem
      // 076: bipush -1
      // 077: ixor
      // 078: if_icmpeq 092
      // 07b: iload 14
      // 07d: iload 3
      // 07e: aload 0
      // 07f: getfield ua.G [I
      // 082: iload 6
      // 084: iaload
      // 085: imul
      // 086: iload 14
      // 088: irem
      // 089: isub
      // 08a: ldc_w 1500479664
      // 08d: ishl
      // 08e: iload 3
      // 08f: idiv
      // 090: istore 10
      // 092: iload 13
      // 094: ldc_w 1015741296
      // 097: ishl
      // 098: iload 4
      // 09a: idiv
      // 09b: istore 11
      // 09d: bipush 0
      // 09e: aload 0
      // 09f: getfield ua.Sb [I
      // 0a2: iload 6
      // 0a4: iaload
      // 0a5: iload 4
      // 0a7: imul
      // 0a8: iload 13
      // 0aa: irem
      // 0ab: if_icmpeq 0c7
      // 0ae: iload 13
      // 0b0: aload 0
      // 0b1: getfield ua.Sb [I
      // 0b4: iload 6
      // 0b6: iaload
      // 0b7: iload 4
      // 0b9: imul
      // 0ba: iload 13
      // 0bc: irem
      // 0bd: isub
      // 0be: ldc_w -1148345008
      // 0c1: ishl
      // 0c2: iload 4
      // 0c4: idiv
      // 0c5: istore 9
      // 0c7: iload 1
      // 0c8: bipush -1
      // 0c9: iload 4
      // 0cb: aload 0
      // 0cc: getfield ua.Sb [I
      // 0cf: iload 6
      // 0d1: iaload
      // 0d2: imul
      // 0d3: iload 13
      // 0d5: ineg
      // 0d6: isub
      // 0d7: iadd
      // 0d8: iload 13
      // 0da: idiv
      // 0db: iadd
      // 0dc: istore 1
      // 0dd: iload 14
      // 0df: ldc_w -959198096
      // 0e2: ishl
      // 0e3: iload 3
      // 0e4: idiv
      // 0e5: istore 12
      // 0e7: iload 2
      // 0e8: iload 14
      // 0ea: iload 3
      // 0eb: aload 0
      // 0ec: getfield ua.G [I
      // 0ef: iload 6
      // 0f1: iaload
      // 0f2: imul
      // 0f3: iadd
      // 0f4: bipush 1
      // 0f5: isub
      // 0f6: iload 14
      // 0f8: idiv
      // 0f9: iadd
      // 0fa: istore 2
      // 0fb: aload 0
      // 0fc: getfield ua.R [I
      // 0ff: iload 6
      // 101: iaload
      // 102: iload 10
      // 104: ldc_w 1117428208
      // 107: ishr
      // 108: isub
      // 109: iload 3
      // 10a: imul
      // 10b: iload 14
      // 10d: idiv
      // 10e: istore 3
      // 10f: iload 4
      // 111: aload 0
      // 112: getfield ua.kb [I
      // 115: iload 6
      // 117: iaload
      // 118: iload 9
      // 11a: ldc_w -725293584
      // 11d: ishr
      // 11e: isub
      // 11f: imul
      // 120: iload 13
      // 122: idiv
      // 123: istore 4
      // 125: iload 1
      // 126: aload 0
      // 127: getfield ua.u I
      // 12a: iload 2
      // 12b: imul
      // 12c: iadd
      // 12d: istore 13
      // 12f: aload 0
      // 130: getfield ua.A I
      // 133: bipush -1
      // 134: ixor
      // 135: iload 2
      // 136: bipush -1
      // 137: ixor
      // 138: if_icmpge 160
      // 13b: aload 0
      // 13c: getfield ua.A I
      // 13f: iload 2
      // 140: isub
      // 141: istore 15
      // 143: iload 10
      // 145: iload 12
      // 147: iload 15
      // 149: imul
      // 14a: iadd
      // 14b: istore 10
      // 14d: iload 3
      // 14e: iload 15
      // 150: isub
      // 151: istore 3
      // 152: iload 13
      // 154: aload 0
      // 155: getfield ua.u I
      // 158: iload 15
      // 15a: imul
      // 15b: iadd
      // 15c: istore 13
      // 15e: bipush 0
      // 15f: istore 2
      // 160: aload 0
      // 161: getfield ua.u I
      // 164: iload 4
      // 166: ineg
      // 167: iadd
      // 168: istore 14
      // 16a: aload 0
      // 16b: getfield ua.Rb I
      // 16e: bipush -1
      // 16f: ixor
      // 170: iload 2
      // 171: iload 3
      // 172: ineg
      // 173: isub
      // 174: bipush -1
      // 175: ixor
      // 176: if_icmpge 17d
      // 179: goto 18c
      // 17c: athrow
      // 17d: iload 3
      // 17e: aload 0
      // 17f: getfield ua.Rb I
      // 182: ineg
      // 183: iload 2
      // 184: iadd
      // 185: iload 3
      // 186: ineg
      // 187: isub
      // 188: bipush -1
      // 189: isub
      // 18a: isub
      // 18b: istore 3
      // 18c: iload 1
      // 18d: aload 0
      // 18e: getfield ua.hb I
      // 191: if_icmplt 198
      // 194: goto 1c2
      // 197: athrow
      // 198: iload 1
      // 199: ineg
      // 19a: aload 0
      // 19b: getfield ua.hb I
      // 19e: iadd
      // 19f: istore 15
      // 1a1: iload 4
      // 1a3: iload 15
      // 1a5: isub
      // 1a6: istore 4
      // 1a8: iload 14
      // 1aa: iload 15
      // 1ac: iadd
      // 1ad: istore 14
      // 1af: iload 13
      // 1b1: iload 15
      // 1b3: iadd
      // 1b4: istore 13
      // 1b6: bipush 0
      // 1b7: istore 1
      // 1b8: iload 9
      // 1ba: iload 11
      // 1bc: iload 15
      // 1be: imul
      // 1bf: iadd
      // 1c0: istore 9
      // 1c2: aload 0
      // 1c3: getfield ua.lb I
      // 1c6: iload 1
      // 1c7: iload 4
      // 1c9: iadd
      // 1ca: if_icmple 1d1
      // 1cd: goto 1ec
      // 1d0: athrow
      // 1d1: bipush 1
      // 1d2: iload 1
      // 1d3: iload 4
      // 1d5: aload 0
      // 1d6: getfield ua.lb I
      // 1d9: isub
      // 1da: iadd
      // 1db: iadd
      // 1dc: istore 15
      // 1de: iload 14
      // 1e0: iload 15
      // 1e2: iadd
      // 1e3: istore 14
      // 1e5: iload 4
      // 1e7: iload 15
      // 1e9: isub
      // 1ea: istore 4
      // 1ec: bipush 1
      // 1ed: istore 15
      // 1ef: bipush 1
      // 1f0: aload 0
      // 1f1: getfield ua.i Z
      // 1f4: ifne 1fc
      // 1f7: bipush 1
      // 1f8: goto 1fd
      // 1fb: athrow
      // 1fc: bipush 0
      // 1fd: if_icmpeq 22a
      // 200: bipush 0
      // 201: iload 2
      // 202: bipush 1
      // 203: iand
      // 204: if_icmpeq 217
      // 207: goto 20b
      // 20a: athrow
      // 20b: iinc 3 -1
      // 20e: iload 13
      // 210: aload 0
      // 211: getfield ua.u I
      // 214: iadd
      // 215: istore 13
      // 217: iload 14
      // 219: aload 0
      // 21a: getfield ua.u I
      // 21d: iadd
      // 21e: istore 14
      // 220: bipush 2
      // 221: istore 15
      // 223: iload 12
      // 225: iload 12
      // 227: iadd
      // 228: istore 12
      // 22a: aload 0
      // 22b: aload 0
      // 22c: getfield ua.ob [[I
      // 22f: iload 6
      // 231: aaload
      // 232: iload 15
      // 234: iload 11
      // 236: bipush 0
      // 237: iload 10
      // 239: aload 0
      // 23a: getfield ua.rb [I
      // 23d: bipush 78
      // 23f: iload 12
      // 241: iload 3
      // 242: iload 9
      // 244: iload 14
      // 246: iload 4
      // 248: iload 7
      // 24a: iload 13
      // 24c: invokespecial ua.a ([IIIII[IBIIIIIII)V
      // 24f: goto 260
      // 252: astore 7
      // 254: getstatic java/lang/System.out Ljava/io/PrintStream;
      // 257: getstatic ua.Yb [Ljava/lang/String;
      // 25a: bipush 16
      // 25c: aaload
      // 25d: invokevirtual java/io/PrintStream.println (Ljava/lang/String;)V
      // 260: iload 5
      // 262: sipush 5924
      // 265: if_icmpeq 272
      // 268: aload 0
      // 269: bipush -15
      // 26b: putfield ua.u I
      // 26e: goto 272
      // 271: athrow
      // 272: goto 2c9
      // 275: astore 7
      // 277: aload 7
      // 279: new java/lang/StringBuilder
      // 27c: dup
      // 27d: invokespecial java/lang/StringBuilder.<init> ()V
      // 280: getstatic ua.Yb [Ljava/lang/String;
      // 283: bipush 15
      // 285: aaload
      // 286: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 289: iload 1
      // 28a: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 28d: bipush 44
      // 28f: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 292: iload 2
      // 293: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 296: bipush 44
      // 298: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 29b: iload 3
      // 29c: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 29f: bipush 44
      // 2a1: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 2a4: iload 4
      // 2a6: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 2a9: bipush 44
      // 2ab: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 2ae: iload 5
      // 2b0: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 2b3: bipush 44
      // 2b5: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 2b8: iload 6
      // 2ba: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 2bd: bipush 41
      // 2bf: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 2c2: invokevirtual java/lang/StringBuilder.toString ()Ljava/lang/String;
      // 2c5: invokestatic i.a (Ljava/lang/Throwable;Ljava/lang/String;)Lla;
      // 2c8: athrow
      // 2c9: return
   }

   final void b(int param1, int param2, int param3, int param4, byte param5) {
      // $VF: Couldn't be decompiled
      // Please report this to the Vineflower issue tracker, at https://github.com/Vineflower/vineflower/issues with a copy of the class file (if you have the rights to distribute it!)
      // java.lang.RuntimeException: parsing failure!
      //   at org.jetbrains.java.decompiler.modules.decompiler.decompose.DomHelper.parseGraph(DomHelper.java:211)
      //   at org.jetbrains.java.decompiler.main.rels.MethodProcessor.codeToJava(MethodProcessor.java:166)
      //
      // Bytecode:
      // 00: getstatic client.vh Z
      // 03: istore 9
      // 05: getstatic ua.F I
      // 08: bipush 1
      // 09: iadd
      // 0a: putstatic ua.F I
      // 0d: aload 0
      // 0e: getfield ua.A I
      // 11: iload 4
      // 13: if_icmpgt 27
      // 16: iload 4
      // 18: aload 0
      // 19: getfield ua.Rb I
      // 1c: if_icmpge 27
      // 1f: goto 23
      // 22: athrow
      // 23: goto 28
      // 26: athrow
      // 27: return
      // 28: iload 3
      // 29: bipush -1
      // 2a: ixor
      // 2b: aload 0
      // 2c: getfield ua.hb I
      // 2f: bipush -1
      // 30: ixor
      // 31: if_icmpgt 38
      // 34: goto 47
      // 37: athrow
      // 38: iload 1
      // 39: iload 3
      // 3a: ineg
      // 3b: aload 0
      // 3c: getfield ua.hb I
      // 3f: iadd
      // 40: isub
      // 41: istore 1
      // 42: aload 0
      // 43: getfield ua.hb I
      // 46: istore 3
      // 47: aload 0
      // 48: getfield ua.lb I
      // 4b: bipush -1
      // 4c: ixor
      // 4d: iload 3
      // 4e: iload 1
      // 4f: ineg
      // 50: isub
      // 51: bipush -1
      // 52: ixor
      // 53: if_icmpgt 5a
      // 56: goto 62
      // 59: athrow
      // 5a: iload 3
      // 5b: ineg
      // 5c: aload 0
      // 5d: getfield ua.lb I
      // 60: iadd
      // 61: istore 1
      // 62: bipush -44
      // 64: iload 5
      // 66: bipush 15
      // 68: isub
      // 69: bipush 37
      // 6b: idiv
      // 6c: idiv
      // 6d: istore 6
      // 6f: bipush 0
      // 70: iload 1
      // 71: if_icmpge 78
      // 74: goto 79
      // 77: athrow
      // 78: return
      // 79: iload 3
      // 7a: aload 0
      // 7b: getfield ua.u I
      // 7e: iload 4
      // 80: imul
      // 81: ineg
      // 82: isub
      // 83: istore 7
      // 85: bipush 0
      // 86: istore 8
      // 88: iload 1
      // 89: iload 8
      // 8b: if_icmple ab
      // 8e: aload 0
      // 8f: getfield ua.rb [I
      // 92: iload 7
      // 94: iload 8
      // 96: ineg
      // 97: isub
      // 98: iload 2
      // 99: iastore
      // 9a: iinc 8 1
      // 9d: iload 9
      // 9f: ifne f8
      // a2: iload 9
      // a4: ifeq 88
      // a7: goto ab
      // aa: athrow
      // ab: goto f8
      // ae: astore 6
      // b0: aload 6
      // b2: new java/lang/StringBuilder
      // b5: dup
      // b6: invokespecial java/lang/StringBuilder.<init> ()V
      // b9: getstatic ua.Yb [Ljava/lang/String;
      // bc: bipush 72
      // be: aaload
      // bf: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // c2: iload 1
      // c3: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // c6: bipush 44
      // c8: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // cb: iload 2
      // cc: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // cf: bipush 44
      // d1: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // d4: iload 3
      // d5: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // d8: bipush 44
      // da: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // dd: iload 4
      // df: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // e2: bipush 44
      // e4: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // e7: iload 5
      // e9: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // ec: bipush 41
      // ee: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // f1: invokevirtual java/lang/StringBuilder.toString ()Ljava/lang/String;
      // f4: invokestatic i.a (Ljava/lang/Throwable;Ljava/lang/String;)Lla;
      // f7: athrow
      // f8: return
   }

   final void a(int param1, int param2, String param3, int param4, int param5, byte param6, int param7) {
      // $VF: Couldn't be decompiled
      // Please report this to the Vineflower issue tracker, at https://github.com/Vineflower/vineflower/issues with a copy of the class file (if you have the rights to distribute it!)
      // java.lang.RuntimeException: parsing failure!
      //   at org.jetbrains.java.decompiler.modules.decompiler.decompose.DomHelper.parseGraph(DomHelper.java:211)
      //   at org.jetbrains.java.decompiler.main.rels.MethodProcessor.codeToJava(MethodProcessor.java:166)
      //
      // Bytecode:
      // 000: getstatic client.vh Z
      // 003: istore 14
      // 005: iload 1
      // 006: bipush -1
      // 007: ixor
      // 008: bipush -1
      // 009: if_icmpge 042
      // 00c: bipush -1
      // 00d: iload 1
      // 00e: aload 0
      // 00f: getfield ua.D I
      // 012: iadd
      // 013: iadd
      // 014: istore 8
      // 016: aload 0
      // 017: getfield ua.gb [[B
      // 01a: iload 8
      // 01c: aaload
      // 01d: ifnull 042
      // 020: aload 0
      // 021: bipush -1
      // 022: iload 8
      // 024: aload 0
      // 025: getfield ua.R [I
      // 028: iload 8
      // 02a: iaload
      // 02b: ineg
      // 02c: iload 2
      // 02d: iadd
      // 02e: iload 4
      // 030: invokevirtual ua.b (IIII)V
      // 033: iload 4
      // 035: aload 0
      // 036: getfield ua.kb [I
      // 039: iload 8
      // 03b: iaload
      // 03c: bipush -5
      // 03e: isub
      // 03f: iadd
      // 040: istore 4
      // 042: bipush -5
      // 044: iload 6
      // 046: bipush -91
      // 048: isub
      // 049: bipush 33
      // 04b: idiv
      // 04c: irem
      // 04d: istore 9
      // 04f: getstatic m.b [[B
      // 052: iload 7
      // 054: aaload
      // 055: astore 8
      // 057: bipush 0
      // 058: istore 10
      // 05a: aload 3
      // 05b: invokevirtual java/lang/String.length ()I
      // 05e: iload 10
      // 060: if_icmple 4d5
      // 063: bipush -65
      // 065: aload 3
      // 066: iload 10
      // 068: invokevirtual java/lang/String.charAt (I)C
      // 06b: bipush -1
      // 06c: ixor
      // 06d: iload 14
      // 06f: ifne 501
      // 072: if_icmpne 351
      // 075: goto 079
      // 078: athrow
      // 079: aload 3
      // 07a: invokevirtual java/lang/String.length ()I
      // 07d: bipush -1
      // 07e: ixor
      // 07f: iload 10
      // 081: bipush -4
      // 083: isub
      // 084: bipush -1
      // 085: ixor
      // 086: if_icmpge 351
      // 089: goto 08d
      // 08c: athrow
      // 08d: aload 3
      // 08e: iload 10
      // 090: bipush -4
      // 092: isub
      // 093: invokevirtual java/lang/String.charAt (I)C
      // 096: bipush 64
      // 098: if_icmpne 351
      // 09b: goto 09f
      // 09e: athrow
      // 09f: aload 3
      // 0a0: iload 10
      // 0a2: bipush -1
      // 0a3: isub
      // 0a4: iload 10
      // 0a6: bipush -4
      // 0a8: isub
      // 0a9: invokevirtual java/lang/String.substring (II)Ljava/lang/String;
      // 0ac: getstatic ua.Yb [Ljava/lang/String;
      // 0af: bipush 38
      // 0b1: aaload
      // 0b2: invokevirtual java/lang/String.equalsIgnoreCase (Ljava/lang/String;)Z
      // 0b5: ifne 345
      // 0b8: goto 0bc
      // 0bb: athrow
      // 0bc: aload 3
      // 0bd: iload 10
      // 0bf: bipush -1
      // 0c0: isub
      // 0c1: iload 10
      // 0c3: bipush -4
      // 0c5: isub
      // 0c6: invokevirtual java/lang/String.substring (II)Ljava/lang/String;
      // 0c9: getstatic ua.Yb [Ljava/lang/String;
      // 0cc: bipush 28
      // 0ce: aaload
      // 0cf: invokevirtual java/lang/String.equalsIgnoreCase (Ljava/lang/String;)Z
      // 0d2: ifeq 0e3
      // 0d5: goto 0d9
      // 0d8: athrow
      // 0d9: ldc_w 16748608
      // 0dc: istore 5
      // 0de: iload 14
      // 0e0: ifeq 349
      // 0e3: aload 3
      // 0e4: iload 10
      // 0e6: bipush -1
      // 0e7: isub
      // 0e8: bipush 4
      // 0e9: iload 10
      // 0eb: iadd
      // 0ec: invokevirtual java/lang/String.substring (II)Ljava/lang/String;
      // 0ef: getstatic ua.Yb [Ljava/lang/String;
      // 0f2: bipush 37
      // 0f4: aaload
      // 0f5: invokevirtual java/lang/String.equalsIgnoreCase (Ljava/lang/String;)Z
      // 0f8: ifne 33b
      // 0fb: goto 0ff
      // 0fe: athrow
      // 0ff: aload 3
      // 100: bipush 1
      // 101: iload 10
      // 103: iadd
      // 104: iload 10
      // 106: bipush 4
      // 107: iadd
      // 108: invokevirtual java/lang/String.substring (II)Ljava/lang/String;
      // 10b: getstatic ua.Yb [Ljava/lang/String;
      // 10e: bipush 23
      // 110: aaload
      // 111: invokevirtual java/lang/String.equalsIgnoreCase (Ljava/lang/String;)Z
      // 114: ifne 332
      // 117: goto 11b
      // 11a: athrow
      // 11b: aload 3
      // 11c: iload 10
      // 11e: bipush -1
      // 11f: isub
      // 120: bipush 4
      // 121: iload 10
      // 123: iadd
      // 124: invokevirtual java/lang/String.substring (II)Ljava/lang/String;
      // 127: getstatic ua.Yb [Ljava/lang/String;
      // 12a: bipush 32
      // 12c: aaload
      // 12d: invokevirtual java/lang/String.equalsIgnoreCase (Ljava/lang/String;)Z
      // 130: ifne 328
      // 133: goto 137
      // 136: athrow
      // 137: aload 3
      // 138: bipush 1
      // 139: iload 10
      // 13b: iadd
      // 13c: bipush 4
      // 13d: iload 10
      // 13f: iadd
      // 140: invokevirtual java/lang/String.substring (II)Ljava/lang/String;
      // 143: getstatic ua.Yb [Ljava/lang/String;
      // 146: bipush 27
      // 148: aaload
      // 149: invokevirtual java/lang/String.equalsIgnoreCase (Ljava/lang/String;)Z
      // 14c: ifeq 15d
      // 14f: goto 153
      // 152: athrow
      // 153: ldc_w 65535
      // 156: istore 5
      // 158: iload 14
      // 15a: ifeq 349
      // 15d: aload 3
      // 15e: bipush 1
      // 15f: iload 10
      // 161: iadd
      // 162: iload 10
      // 164: bipush 4
      // 165: iadd
      // 166: invokevirtual java/lang/String.substring (II)Ljava/lang/String;
      // 169: getstatic ua.Yb [Ljava/lang/String;
      // 16c: bipush 25
      // 16e: aaload
      // 16f: invokevirtual java/lang/String.equalsIgnoreCase (Ljava/lang/String;)Z
      // 172: ifeq 182
      // 175: goto 179
      // 178: athrow
      // 179: ldc 16711935
      // 17b: istore 5
      // 17d: iload 14
      // 17f: ifeq 349
      // 182: aload 3
      // 183: iload 10
      // 185: bipush 1
      // 186: iadd
      // 187: bipush 4
      // 188: iload 10
      // 18a: iadd
      // 18b: invokevirtual java/lang/String.substring (II)Ljava/lang/String;
      // 18e: getstatic ua.Yb [Ljava/lang/String;
      // 191: bipush 33
      // 193: aaload
      // 194: invokevirtual java/lang/String.equalsIgnoreCase (Ljava/lang/String;)Z
      // 197: ifne 31f
      // 19a: goto 19e
      // 19d: athrow
      // 19e: aload 3
      // 19f: iload 10
      // 1a1: bipush 1
      // 1a2: iadd
      // 1a3: iload 10
      // 1a5: bipush -4
      // 1a7: isub
      // 1a8: invokevirtual java/lang/String.substring (II)Ljava/lang/String;
      // 1ab: getstatic ua.Yb [Ljava/lang/String;
      // 1ae: bipush 35
      // 1b0: aaload
      // 1b1: invokevirtual java/lang/String.equalsIgnoreCase (Ljava/lang/String;)Z
      // 1b4: ifne 317
      // 1b7: goto 1bb
      // 1ba: athrow
      // 1bb: aload 3
      // 1bc: bipush 1
      // 1bd: iload 10
      // 1bf: iadd
      // 1c0: iload 10
      // 1c2: bipush -4
      // 1c4: isub
      // 1c5: invokevirtual java/lang/String.substring (II)Ljava/lang/String;
      // 1c8: getstatic ua.Yb [Ljava/lang/String;
      // 1cb: bipush 31
      // 1cd: aaload
      // 1ce: invokevirtual java/lang/String.equalsIgnoreCase (Ljava/lang/String;)Z
      // 1d1: ifeq 1e2
      // 1d4: goto 1d8
      // 1d7: athrow
      // 1d8: ldc_w 12582912
      // 1db: istore 5
      // 1dd: iload 14
      // 1df: ifeq 349
      // 1e2: aload 3
      // 1e3: bipush 1
      // 1e4: iload 10
      // 1e6: iadd
      // 1e7: bipush 4
      // 1e8: iload 10
      // 1ea: iadd
      // 1eb: invokevirtual java/lang/String.substring (II)Ljava/lang/String;
      // 1ee: getstatic ua.Yb [Ljava/lang/String;
      // 1f1: bipush 41
      // 1f3: aaload
      // 1f4: invokevirtual java/lang/String.equalsIgnoreCase (Ljava/lang/String;)Z
      // 1f7: ifne 30d
      // 1fa: goto 1fe
      // 1fd: athrow
      // 1fe: aload 3
      // 1ff: iload 10
      // 201: bipush -1
      // 202: isub
      // 203: bipush 4
      // 204: iload 10
      // 206: iadd
      // 207: invokevirtual java/lang/String.substring (II)Ljava/lang/String;
      // 20a: getstatic ua.Yb [Ljava/lang/String;
      // 20d: bipush 39
      // 20f: aaload
      // 210: invokevirtual java/lang/String.equalsIgnoreCase (Ljava/lang/String;)Z
      // 213: ifne 2fe
      // 216: goto 21a
      // 219: athrow
      // 21a: aload 3
      // 21b: bipush 1
      // 21c: iload 10
      // 21e: iadd
      // 21f: bipush 4
      // 220: iload 10
      // 222: iadd
      // 223: invokevirtual java/lang/String.substring (II)Ljava/lang/String;
      // 226: getstatic ua.Yb [Ljava/lang/String;
      // 229: bipush 24
      // 22b: aaload
      // 22c: invokevirtual java/lang/String.equalsIgnoreCase (Ljava/lang/String;)Z
      // 22f: ifeq 240
      // 232: goto 236
      // 235: athrow
      // 236: ldc_w 16756736
      // 239: istore 5
      // 23b: iload 14
      // 23d: ifeq 349
      // 240: aload 3
      // 241: iload 10
      // 243: bipush 1
      // 244: iadd
      // 245: bipush 4
      // 246: iload 10
      // 248: iadd
      // 249: invokevirtual java/lang/String.substring (II)Ljava/lang/String;
      // 24c: getstatic ua.Yb [Ljava/lang/String;
      // 24f: bipush 40
      // 251: aaload
      // 252: invokevirtual java/lang/String.equalsIgnoreCase (Ljava/lang/String;)Z
      // 255: ifne 2f5
      // 258: goto 25c
      // 25b: athrow
      // 25c: aload 3
      // 25d: iload 10
      // 25f: bipush 1
      // 260: iadd
      // 261: iload 10
      // 263: bipush -4
      // 265: isub
      // 266: invokevirtual java/lang/String.substring (II)Ljava/lang/String;
      // 269: getstatic ua.Yb [Ljava/lang/String;
      // 26c: bipush 22
      // 26e: aaload
      // 26f: invokevirtual java/lang/String.equalsIgnoreCase (Ljava/lang/String;)Z
      // 272: ifne 2eb
      // 275: goto 279
      // 278: athrow
      // 279: aload 3
      // 27a: iload 10
      // 27c: bipush 1
      // 27d: iadd
      // 27e: bipush 4
      // 27f: iload 10
      // 281: iadd
      // 282: invokevirtual java/lang/String.substring (II)Ljava/lang/String;
      // 285: getstatic ua.Yb [Ljava/lang/String;
      // 288: bipush 29
      // 28a: aaload
      // 28b: invokevirtual java/lang/String.equalsIgnoreCase (Ljava/lang/String;)Z
      // 28e: ifeq 29f
      // 291: goto 295
      // 294: athrow
      // 295: ldc_w 12648192
      // 298: istore 5
      // 29a: iload 14
      // 29c: ifeq 349
      // 29f: aload 3
      // 2a0: bipush 1
      // 2a1: iload 10
      // 2a3: iadd
      // 2a4: bipush 4
      // 2a5: iload 10
      // 2a7: iadd
      // 2a8: invokevirtual java/lang/String.substring (II)Ljava/lang/String;
      // 2ab: getstatic ua.Yb [Ljava/lang/String;
      // 2ae: bipush 34
      // 2b0: aaload
      // 2b1: invokevirtual java/lang/String.equalsIgnoreCase (Ljava/lang/String;)Z
      // 2b4: ifeq 2c5
      // 2b7: goto 2bb
      // 2ba: athrow
      // 2bb: ldc_w 8453888
      // 2be: istore 5
      // 2c0: iload 14
      // 2c2: ifeq 349
      // 2c5: aload 3
      // 2c6: iload 10
      // 2c8: bipush -1
      // 2c9: isub
      // 2ca: bipush 4
      // 2cb: iload 10
      // 2cd: iadd
      // 2ce: invokevirtual java/lang/String.substring (II)Ljava/lang/String;
      // 2d1: getstatic ua.Yb [Ljava/lang/String;
      // 2d4: bipush 36
      // 2d6: aaload
      // 2d7: invokevirtual java/lang/String.equalsIgnoreCase (Ljava/lang/String;)Z
      // 2da: ifeq 349
      // 2dd: goto 2e1
      // 2e0: athrow
      // 2e1: ldc_w 4259584
      // 2e4: istore 5
      // 2e6: iload 14
      // 2e8: ifeq 349
      // 2eb: ldc_w 16723968
      // 2ee: istore 5
      // 2f0: iload 14
      // 2f2: ifeq 349
      // 2f5: ldc 16740352
      // 2f7: istore 5
      // 2f9: iload 14
      // 2fb: ifeq 349
      // 2fe: ldc2_w 1.6777215E7
      // 301: invokestatic java/lang/Math.random ()D
      // 304: dmul
      // 305: d2i
      // 306: istore 5
      // 308: iload 14
      // 30a: ifeq 349
      // 30d: ldc_w 16748608
      // 310: istore 5
      // 312: iload 14
      // 314: ifeq 349
      // 317: bipush 0
      // 318: istore 5
      // 31a: iload 14
      // 31c: ifeq 349
      // 31f: ldc 16777215
      // 321: istore 5
      // 323: iload 14
      // 325: ifeq 349
      // 328: sipush 255
      // 32b: istore 5
      // 32d: iload 14
      // 32f: ifeq 349
      // 332: ldc 65280
      // 334: istore 5
      // 336: iload 14
      // 338: ifeq 349
      // 33b: ldc_w 16776960
      // 33e: istore 5
      // 340: iload 14
      // 342: ifeq 349
      // 345: ldc 16711680
      // 347: istore 5
      // 349: iinc 10 4
      // 34c: iload 14
      // 34e: ifeq 4cd
      // 351: bipush -127
      // 353: aload 3
      // 354: iload 10
      // 356: invokevirtual java/lang/String.charAt (I)C
      // 359: bipush -1
      // 35a: ixor
      // 35b: if_icmpne 400
      // 35e: goto 362
      // 361: athrow
      // 362: aload 3
      // 363: invokevirtual java/lang/String.length ()I
      // 366: bipush -1
      // 367: ixor
      // 368: bipush 4
      // 369: iload 10
      // 36b: iadd
      // 36c: bipush -1
      // 36d: ixor
      // 36e: if_icmpge 400
      // 371: goto 375
      // 374: athrow
      // 375: bipush 126
      // 377: aload 3
      // 378: bipush 4
      // 379: iload 10
      // 37b: iadd
      // 37c: invokevirtual java/lang/String.charAt (I)C
      // 37f: if_icmpne 400
      // 382: goto 386
      // 385: athrow
      // 386: aload 3
      // 387: bipush 1
      // 388: iload 10
      // 38a: iadd
      // 38b: invokevirtual java/lang/String.charAt (I)C
      // 38e: istore 11
      // 390: aload 3
      // 391: bipush 2
      // 392: iload 10
      // 394: iadd
      // 395: invokevirtual java/lang/String.charAt (I)C
      // 398: istore 12
      // 39a: aload 3
      // 39b: iload 10
      // 39d: bipush 3
      // 39e: iadd
      // 39f: invokevirtual java/lang/String.charAt (I)C
      // 3a2: istore 13
      // 3a4: bipush -49
      // 3a6: iload 11
      // 3a8: bipush -1
      // 3a9: ixor
      // 3aa: if_icmplt 3f8
      // 3ad: iload 11
      // 3af: bipush -1
      // 3b0: ixor
      // 3b1: bipush -58
      // 3b3: if_icmplt 3f8
      // 3b6: goto 3ba
      // 3b9: athrow
      // 3ba: iload 12
      // 3bc: bipush 48
      // 3be: if_icmplt 3f8
      // 3c1: goto 3c5
      // 3c4: athrow
      // 3c5: bipush 57
      // 3c7: iload 12
      // 3c9: if_icmplt 3f8
      // 3cc: goto 3d0
      // 3cf: athrow
      // 3d0: iload 13
      // 3d2: bipush 48
      // 3d4: if_icmplt 3f8
      // 3d7: goto 3db
      // 3da: athrow
      // 3db: iload 13
      // 3dd: bipush 57
      // 3df: if_icmpgt 3f8
      // 3e2: goto 3e6
      // 3e5: athrow
      // 3e6: aload 3
      // 3e7: iload 10
      // 3e9: bipush 1
      // 3ea: iadd
      // 3eb: iload 10
      // 3ed: bipush -4
      // 3ef: isub
      // 3f0: invokevirtual java/lang/String.substring (II)Ljava/lang/String;
      // 3f3: invokestatic java/lang/Integer.parseInt (Ljava/lang/String;)I
      // 3f6: istore 4
      // 3f8: iinc 10 4
      // 3fb: iload 14
      // 3fd: ifeq 4cd
      // 400: aload 3
      // 401: iload 10
      // 403: invokevirtual java/lang/String.charAt (I)C
      // 406: istore 11
      // 408: iload 11
      // 40a: sipush 160
      // 40d: if_icmpeq 414
      // 410: goto 418
      // 413: athrow
      // 414: bipush 32
      // 416: istore 11
      // 418: iload 11
      // 41a: bipush -1
      // 41b: ixor
      // 41c: bipush -1
      // 41d: if_icmpgt 42d
      // 420: getstatic n.a [I
      // 423: arraylength
      // 424: iload 11
      // 426: if_icmpgt 431
      // 429: goto 42d
      // 42c: athrow
      // 42d: bipush 32
      // 42f: istore 11
      // 431: getstatic n.a [I
      // 434: iload 11
      // 436: iaload
      // 437: istore 12
      // 439: aload 0
      // 43a: getfield ua.xb Z
      // 43d: ifeq 471
      // 440: getstatic fb.k [Z
      // 443: iload 7
      // 445: baload
      // 446: ifne 471
      // 449: goto 44d
      // 44c: athrow
      // 44d: bipush 0
      // 44e: iload 5
      // 450: if_icmpeq 471
      // 453: goto 457
      // 456: athrow
      // 457: aload 0
      // 458: bipush 53
      // 45a: getstatic fb.k [Z
      // 45d: iload 7
      // 45f: baload
      // 460: aload 8
      // 462: bipush 1
      // 463: iload 4
      // 465: iadd
      // 466: bipush 0
      // 467: iload 12
      // 469: iload 2
      // 46a: invokespecial ua.a (BZ[BIIII)V
      // 46d: goto 471
      // 470: athrow
      // 471: aload 0
      // 472: getfield ua.xb Z
      // 475: ifeq 4ab
      // 478: getstatic fb.k [Z
      // 47b: iload 7
      // 47d: baload
      // 47e: ifne 4ab
      // 481: goto 485
      // 484: athrow
      // 485: iload 5
      // 487: bipush -1
      // 488: ixor
      // 489: bipush -1
      // 48a: if_icmpeq 4ab
      // 48d: goto 491
      // 490: athrow
      // 491: aload 0
      // 492: bipush 101
      // 494: getstatic fb.k [Z
      // 497: iload 7
      // 499: baload
      // 49a: aload 8
      // 49c: iload 4
      // 49e: bipush 0
      // 49f: iload 12
      // 4a1: iload 2
      // 4a2: bipush -1
      // 4a3: isub
      // 4a4: invokespecial ua.a (BZ[BIIII)V
      // 4a7: goto 4ab
      // 4aa: athrow
      // 4ab: aload 0
      // 4ac: bipush 73
      // 4ae: getstatic fb.k [Z
      // 4b1: iload 7
      // 4b3: baload
      // 4b4: aload 8
      // 4b6: iload 4
      // 4b8: iload 5
      // 4ba: iload 12
      // 4bc: iload 2
      // 4bd: invokespecial ua.a (BZ[BIIII)V
      // 4c0: iload 4
      // 4c2: aload 8
      // 4c4: iload 12
      // 4c6: bipush 7
      // 4c8: iadd
      // 4c9: baload
      // 4ca: iadd
      // 4cb: istore 4
      // 4cd: iinc 10 1
      // 4d0: iload 14
      // 4d2: ifeq 05a
      // 4d5: goto 4fd
      // 4d8: astore 8
      // 4da: getstatic java/lang/System.out Ljava/io/PrintStream;
      // 4dd: new java/lang/StringBuilder
      // 4e0: dup
      // 4e1: invokespecial java/lang/StringBuilder.<init> ()V
      // 4e4: getstatic ua.Yb [Ljava/lang/String;
      // 4e7: bipush 30
      // 4e9: aaload
      // 4ea: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 4ed: aload 8
      // 4ef: invokevirtual java/lang/StringBuilder.append (Ljava/lang/Object;)Ljava/lang/StringBuilder;
      // 4f2: invokevirtual java/lang/StringBuilder.toString ()Ljava/lang/String;
      // 4f5: invokevirtual java/io/PrintStream.println (Ljava/lang/String;)V
      // 4f8: aload 8
      // 4fa: invokevirtual java/lang/Exception.printStackTrace ()V
      // 4fd: getstatic ua.w I
      // 500: bipush 1
      // 501: iadd
      // 502: putstatic ua.w I
      // 505: goto 577
      // 508: astore 8
      // 50a: aload 8
      // 50c: new java/lang/StringBuilder
      // 50f: dup
      // 510: invokespecial java/lang/StringBuilder.<init> ()V
      // 513: getstatic ua.Yb [Ljava/lang/String;
      // 516: bipush 26
      // 518: aaload
      // 519: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 51c: iload 1
      // 51d: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 520: bipush 44
      // 522: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 525: iload 2
      // 526: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 529: bipush 44
      // 52b: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 52e: aload 3
      // 52f: ifnull 53b
      // 532: getstatic ua.Yb [Ljava/lang/String;
      // 535: bipush 1
      // 536: aaload
      // 537: goto 540
      // 53a: athrow
      // 53b: getstatic ua.Yb [Ljava/lang/String;
      // 53e: bipush 0
      // 53f: aaload
      // 540: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 543: bipush 44
      // 545: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 548: iload 4
      // 54a: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 54d: bipush 44
      // 54f: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 552: iload 5
      // 554: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 557: bipush 44
      // 559: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 55c: iload 6
      // 55e: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 561: bipush 44
      // 563: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 566: iload 7
      // 568: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 56b: bipush 41
      // 56d: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 570: invokevirtual java/lang/StringBuilder.toString ()Ljava/lang/String;
      // 573: invokestatic i.a (Ljava/lang/Throwable;Ljava/lang/String;)Lla;
      // 576: athrow
      // 577: return
   }

   @Override
   public final synchronized void removeConsumer(ImageConsumer var1) {
      try {
         label32: {
            try {
               if (this.fb != var1) {
                  break label32;
               }
            } catch (RuntimeException var4) {
               throw var4;
            }

            this.fb = null;
         }

         p++;
      } catch (RuntimeException var5) {
         RuntimeException var2 = var5;

         RuntimeException var10000;
         StringBuilder var10001;
         try {
            var10000 = var2;
            var10001 = new StringBuilder().append(Yb[42]);
            if (var1 != null) {
               throw i.a(var2, var10001.append(Yb[1]).append(')').toString());
            }
         } catch (RuntimeException var3) {
            throw var3;
         }

         throw i.a(var10000, var10001.append(Yb[0]).append(')').toString());
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

   final void b(int param1, int param2, int param3, int param4, int param5) {
      // $VF: Couldn't be decompiled
      // Please report this to the Vineflower issue tracker, at https://github.com/Vineflower/vineflower/issues with a copy of the class file (if you have the rights to distribute it!)
      // java.lang.RuntimeException: parsing failure!
      //   at org.jetbrains.java.decompiler.modules.decompiler.decompose.DomHelper.parseGraph(DomHelper.java:211)
      //   at org.jetbrains.java.decompiler.main.rels.MethodProcessor.codeToJava(MethodProcessor.java:166)
      //
      // Bytecode:
      // 00: getstatic client.vh Z
      // 03: istore 8
      // 05: getstatic ua.j I
      // 08: bipush 1
      // 09: iadd
      // 0a: putstatic ua.j I
      // 0d: iload 1
      // 0e: bipush -1
      // 0f: ixor
      // 10: aload 0
      // 11: getfield ua.hb I
      // 14: bipush -1
      // 15: ixor
      // 16: if_icmpgt 2d
      // 19: aload 0
      // 1a: getfield ua.lb I
      // 1d: bipush -1
      // 1e: ixor
      // 1f: iload 1
      // 20: bipush -1
      // 21: ixor
      // 22: if_icmpge 2d
      // 25: goto 29
      // 28: athrow
      // 29: goto 2e
      // 2c: athrow
      // 2d: return
      // 2e: aload 0
      // 2f: getfield ua.A I
      // 32: bipush -1
      // 33: ixor
      // 34: iload 2
      // 35: bipush -1
      // 36: ixor
      // 37: if_icmplt 3e
      // 3a: goto 4f
      // 3d: athrow
      // 3e: iload 4
      // 40: aload 0
      // 41: getfield ua.A I
      // 44: iload 2
      // 45: ineg
      // 46: iadd
      // 47: isub
      // 48: istore 4
      // 4a: aload 0
      // 4b: getfield ua.A I
      // 4e: istore 2
      // 4f: aload 0
      // 50: getfield ua.Rb I
      // 53: bipush -1
      // 54: ixor
      // 55: iload 2
      // 56: iload 4
      // 58: iadd
      // 59: bipush -1
      // 5a: ixor
      // 5b: if_icmpgt 62
      // 5e: goto 6b
      // 61: athrow
      // 62: iload 2
      // 63: ineg
      // 64: aload 0
      // 65: getfield ua.Rb I
      // 68: iadd
      // 69: istore 4
      // 6b: iload 4
      // 6d: iload 5
      // 6f: if_icmpgt 73
      // 72: return
      // 73: iload 1
      // 74: aload 0
      // 75: getfield ua.u I
      // 78: iload 2
      // 79: imul
      // 7a: ineg
      // 7b: isub
      // 7c: istore 6
      // 7e: bipush 0
      // 7f: istore 7
      // 81: iload 7
      // 83: bipush -1
      // 84: ixor
      // 85: iload 4
      // 87: bipush -1
      // 88: ixor
      // 89: if_icmple ae
      // 8c: aload 0
      // 8d: getfield ua.rb [I
      // 90: iload 6
      // 92: aload 0
      // 93: getfield ua.u I
      // 96: iload 7
      // 98: imul
      // 99: ineg
      // 9a: isub
      // 9b: iload 3
      // 9c: iastore
      // 9d: iinc 7 1
      // a0: iload 8
      // a2: ifne fb
      // a5: iload 8
      // a7: ifeq 81
      // aa: goto ae
      // ad: athrow
      // ae: goto fb
      // b1: astore 6
      // b3: aload 6
      // b5: new java/lang/StringBuilder
      // b8: dup
      // b9: invokespecial java/lang/StringBuilder.<init> ()V
      // bc: getstatic ua.Yb [Ljava/lang/String;
      // bf: bipush 43
      // c1: aaload
      // c2: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // c5: iload 1
      // c6: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // c9: bipush 44
      // cb: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // ce: iload 2
      // cf: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // d2: bipush 44
      // d4: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // d7: iload 3
      // d8: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // db: bipush 44
      // dd: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // e0: iload 4
      // e2: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // e5: bipush 44
      // e7: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // ea: iload 5
      // ec: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // ef: bipush 41
      // f1: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // f4: invokevirtual java/lang/StringBuilder.toString ()Ljava/lang/String;
      // f7: invokestatic i.a (Ljava/lang/Throwable;Ljava/lang/String;)Lla;
      // fa: athrow
      // fb: return
   }

   final void a(int param1, int param2, int param3, int param4, int param5) {
      // $VF: Couldn't be decompiled
      // Please report this to the Vineflower issue tracker, at https://github.com/Vineflower/vineflower/issues with a copy of the class file (if you have the rights to distribute it!)
      // java.lang.RuntimeException: parsing failure!
      //   at org.jetbrains.java.decompiler.modules.decompiler.decompose.DomHelper.parseGraph(DomHelper.java:211)
      //   at org.jetbrains.java.decompiler.main.rels.MethodProcessor.codeToJava(MethodProcessor.java:166)
      //
      // Bytecode:
      // 000: aload 0
      // 001: getfield ua.Qb [Z
      // 004: iload 1
      // 005: baload
      // 006: ifeq 01d
      // 009: iload 3
      // 00a: aload 0
      // 00b: getfield ua.Sb [I
      // 00e: iload 1
      // 00f: iaload
      // 010: iadd
      // 011: istore 3
      // 012: iload 5
      // 014: aload 0
      // 015: getfield ua.G [I
      // 018: iload 1
      // 019: iaload
      // 01a: iadd
      // 01b: istore 5
      // 01d: getstatic ua.sb I
      // 020: bipush 1
      // 021: iadd
      // 022: putstatic ua.sb I
      // 025: aload 0
      // 026: getfield ua.u I
      // 029: iload 5
      // 02b: imul
      // 02c: iload 3
      // 02d: iadd
      // 02e: istore 6
      // 030: iload 2
      // 031: istore 7
      // 033: aload 0
      // 034: getfield ua.R [I
      // 037: iload 1
      // 038: iaload
      // 039: istore 8
      // 03b: aload 0
      // 03c: getfield ua.kb [I
      // 03f: iload 1
      // 040: iaload
      // 041: istore 9
      // 043: aload 0
      // 044: getfield ua.u I
      // 047: iload 9
      // 049: isub
      // 04a: istore 10
      // 04c: bipush 0
      // 04d: istore 11
      // 04f: aload 0
      // 050: getfield ua.A I
      // 053: bipush -1
      // 054: ixor
      // 055: iload 5
      // 057: bipush -1
      // 058: ixor
      // 059: if_icmpge 089
      // 05c: iload 5
      // 05e: ineg
      // 05f: aload 0
      // 060: getfield ua.A I
      // 063: iadd
      // 064: istore 12
      // 066: aload 0
      // 067: getfield ua.A I
      // 06a: istore 5
      // 06c: iload 8
      // 06e: iload 12
      // 070: isub
      // 071: istore 8
      // 073: iload 7
      // 075: iload 12
      // 077: iload 9
      // 079: imul
      // 07a: iadd
      // 07b: istore 7
      // 07d: iload 6
      // 07f: iload 12
      // 081: aload 0
      // 082: getfield ua.u I
      // 085: imul
      // 086: iadd
      // 087: istore 6
      // 089: aload 0
      // 08a: getfield ua.Rb I
      // 08d: bipush -1
      // 08e: ixor
      // 08f: iload 8
      // 091: iload 5
      // 093: iadd
      // 094: bipush -1
      // 095: ixor
      // 096: if_icmplt 0ab
      // 099: iload 8
      // 09b: bipush 1
      // 09c: iload 8
      // 09e: iload 5
      // 0a0: aload 0
      // 0a1: getfield ua.Rb I
      // 0a4: ineg
      // 0a5: iadd
      // 0a6: iadd
      // 0a7: iadd
      // 0a8: isub
      // 0a9: istore 8
      // 0ab: iload 3
      // 0ac: aload 0
      // 0ad: getfield ua.hb I
      // 0b0: if_icmplt 0b7
      // 0b3: goto 0e8
      // 0b6: athrow
      // 0b7: aload 0
      // 0b8: getfield ua.hb I
      // 0bb: iload 3
      // 0bc: ineg
      // 0bd: iadd
      // 0be: istore 12
      // 0c0: iload 10
      // 0c2: iload 12
      // 0c4: iadd
      // 0c5: istore 10
      // 0c7: iload 11
      // 0c9: iload 12
      // 0cb: iadd
      // 0cc: istore 11
      // 0ce: iload 6
      // 0d0: iload 12
      // 0d2: iadd
      // 0d3: istore 6
      // 0d5: iload 7
      // 0d7: iload 12
      // 0d9: iadd
      // 0da: istore 7
      // 0dc: aload 0
      // 0dd: getfield ua.hb I
      // 0e0: istore 3
      // 0e1: iload 9
      // 0e3: iload 12
      // 0e5: isub
      // 0e6: istore 9
      // 0e8: aload 0
      // 0e9: getfield ua.lb I
      // 0ec: iload 3
      // 0ed: iload 9
      // 0ef: iadd
      // 0f0: if_icmple 0f7
      // 0f3: goto 11b
      // 0f6: athrow
      // 0f7: aload 0
      // 0f8: getfield ua.lb I
      // 0fb: ineg
      // 0fc: iload 3
      // 0fd: iload 9
      // 0ff: ineg
      // 100: bipush 1
      // 101: isub
      // 102: isub
      // 103: iadd
      // 104: istore 12
      // 106: iload 10
      // 108: iload 12
      // 10a: iadd
      // 10b: istore 10
      // 10d: iload 9
      // 10f: iload 12
      // 111: isub
      // 112: istore 9
      // 114: iload 11
      // 116: iload 12
      // 118: iadd
      // 119: istore 11
      // 11b: iload 9
      // 11d: bipush -1
      // 11e: ixor
      // 11f: bipush -1
      // 120: if_icmpge 133
      // 123: iload 8
      // 125: bipush -1
      // 126: ixor
      // 127: bipush -1
      // 128: if_icmpge 133
      // 12b: goto 12f
      // 12e: athrow
      // 12f: goto 134
      // 132: athrow
      // 133: return
      // 134: bipush 1
      // 135: istore 12
      // 137: bipush 0
      // 138: aload 0
      // 139: getfield ua.i Z
      // 13c: ifne 144
      // 13f: bipush 1
      // 140: goto 145
      // 143: athrow
      // 144: bipush 0
      // 145: if_icmpeq 14b
      // 148: goto 175
      // 14b: iload 11
      // 14d: aload 0
      // 14e: getfield ua.kb [I
      // 151: iload 1
      // 152: iaload
      // 153: iadd
      // 154: istore 11
      // 156: bipush 2
      // 157: istore 12
      // 159: bipush 1
      // 15a: iload 5
      // 15c: iand
      // 15d: ifeq 16c
      // 160: iinc 8 -1
      // 163: iload 6
      // 165: aload 0
      // 166: getfield ua.u I
      // 169: iadd
      // 16a: istore 6
      // 16c: iload 10
      // 16e: aload 0
      // 16f: getfield ua.u I
      // 172: iadd
      // 173: istore 10
      // 175: aload 0
      // 176: getfield ua.ob [[I
      // 179: iload 1
      // 17a: aaload
      // 17b: ifnonnull 1ad
      // 17e: aload 0
      // 17f: iload 7
      // 181: iload 10
      // 183: aload 0
      // 184: getfield ua.gb [[B
      // 187: iload 1
      // 188: aaload
      // 189: iload 12
      // 18b: bipush 0
      // 18c: iload 8
      // 18e: iload 11
      // 190: iload 4
      // 192: iload 9
      // 194: aload 0
      // 195: getfield ua.rb [I
      // 198: aload 0
      // 199: getfield ua.Y [[I
      // 19c: iload 1
      // 19d: aaload
      // 19e: iload 6
      // 1a0: invokespecial ua.a (II[BIZIIII[I[II)V
      // 1a3: getstatic client.vh Z
      // 1a6: ifeq 1d2
      // 1a9: goto 1ad
      // 1ac: athrow
      // 1ad: aload 0
      // 1ae: iload 7
      // 1b0: iload 8
      // 1b2: iload 6
      // 1b4: aload 0
      // 1b5: getfield ua.ob [[I
      // 1b8: iload 1
      // 1b9: aaload
      // 1ba: bipush 0
      // 1bb: iload 4
      // 1bd: iload 12
      // 1bf: aload 0
      // 1c0: getfield ua.rb [I
      // 1c3: bipush -107
      // 1c5: iload 11
      // 1c7: iload 10
      // 1c9: iload 9
      // 1cb: invokespecial ua.a (III[IIII[IIIII)V
      // 1ce: goto 1d2
      // 1d1: athrow
      // 1d2: goto 21f
      // 1d5: astore 6
      // 1d7: aload 6
      // 1d9: new java/lang/StringBuilder
      // 1dc: dup
      // 1dd: invokespecial java/lang/StringBuilder.<init> ()V
      // 1e0: getstatic ua.Yb [Ljava/lang/String;
      // 1e3: bipush 19
      // 1e5: aaload
      // 1e6: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 1e9: iload 1
      // 1ea: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 1ed: bipush 44
      // 1ef: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 1f2: iload 2
      // 1f3: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 1f6: bipush 44
      // 1f8: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 1fb: iload 3
      // 1fc: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 1ff: bipush 44
      // 201: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 204: iload 4
      // 206: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 209: bipush 44
      // 20b: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 20e: iload 5
      // 210: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 213: bipush 41
      // 215: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 218: invokevirtual java/lang/StringBuilder.toString ()Ljava/lang/String;
      // 21b: invokestatic i.a (Ljava/lang/Throwable;Ljava/lang/String;)Lla;
      // 21e: athrow
      // 21f: return
   }

   final void a(Graphics var1, int var2, int var3, int var4) {
      try {
         e++;
         this.b(true);
         var1.drawImage(this.Gb, var2, var4, this);
         if (var3 != 256) {
            Kb = (String[])null;
         }
      } catch (RuntimeException var7) {
         RuntimeException var5 = var7;

         RuntimeException var10000;
         StringBuilder var10001;
         try {
            var10000 = var5;
            var10001 = new StringBuilder().append(Yb[49]);
            if (var1 != null) {
               throw i.a(var5, var10001.append(Yb[1]).append(',').append(var2).append(',').append(var3).append(',').append(var4).append(')').toString());
            }
         } catch (RuntimeException var6) {
            throw var6;
         }

         throw i.a(var10000, var10001.append(Yb[0]).append(',').append(var2).append(',').append(var3).append(',').append(var4).append(')').toString());
      }
   }

   final void a(int param1, byte param2, int param3, int param4, int param5, int param6, int param7) {
      // $VF: Couldn't be decompiled
      // Please report this to the Vineflower issue tracker, at https://github.com/Vineflower/vineflower/issues with a copy of the class file (if you have the rights to distribute it!)
      // java.lang.RuntimeException: parsing failure!
      //   at org.jetbrains.java.decompiler.modules.decompiler.decompose.DomHelper.parseGraph(DomHelper.java:211)
      //   at org.jetbrains.java.decompiler.main.rels.MethodProcessor.codeToJava(MethodProcessor.java:166)
      //
      // Bytecode:
      // 000: getstatic ua.I I
      // 003: bipush 1
      // 004: iadd
      // 005: putstatic ua.I I
      // 008: aload 0
      // 009: getfield ua.kb [I
      // 00c: iload 1
      // 00d: iaload
      // 00e: istore 8
      // 010: aload 0
      // 011: getfield ua.R [I
      // 014: iload 1
      // 015: iaload
      // 016: istore 9
      // 018: bipush 0
      // 019: istore 10
      // 01b: bipush 0
      // 01c: istore 11
      // 01e: iload 8
      // 020: ldc_w 2133607216
      // 023: ishl
      // 024: iload 5
      // 026: idiv
      // 027: istore 12
      // 029: iload 9
      // 02b: ldc_w -1116806288
      // 02e: ishl
      // 02f: iload 3
      // 030: idiv
      // 031: istore 13
      // 033: aload 0
      // 034: getfield ua.Qb [Z
      // 037: iload 1
      // 038: baload
      // 039: ifeq 122
      // 03c: aload 0
      // 03d: getfield ua.Eb [I
      // 040: iload 1
      // 041: iaload
      // 042: istore 14
      // 044: aload 0
      // 045: getfield ua.qb [I
      // 048: iload 1
      // 049: iaload
      // 04a: istore 15
      // 04c: iload 14
      // 04e: ifeq 061
      // 051: bipush -1
      // 052: iload 15
      // 054: bipush -1
      // 055: ixor
      // 056: if_icmpeq 061
      // 059: goto 05d
      // 05c: athrow
      // 05d: goto 062
      // 060: athrow
      // 061: return
      // 062: iload 15
      // 064: ldc_w 1837802000
      // 067: ishl
      // 068: iload 3
      // 069: idiv
      // 06a: istore 13
      // 06c: iload 6
      // 06e: iload 15
      // 070: iload 3
      // 071: aload 0
      // 072: getfield ua.G [I
      // 075: iload 1
      // 076: iaload
      // 077: imul
      // 078: iadd
      // 079: bipush -1
      // 07a: iadd
      // 07b: iload 15
      // 07d: idiv
      // 07e: iadd
      // 07f: istore 6
      // 081: iload 4
      // 083: iload 14
      // 085: aload 0
      // 086: getfield ua.Sb [I
      // 089: iload 1
      // 08a: iaload
      // 08b: iload 5
      // 08d: imul
      // 08e: iadd
      // 08f: bipush -1
      // 090: iadd
      // 091: iload 14
      // 093: idiv
      // 094: iadd
      // 095: istore 4
      // 097: iload 14
      // 099: ldc_w -405443120
      // 09c: ishl
      // 09d: iload 5
      // 09f: idiv
      // 0a0: istore 12
      // 0a2: bipush -1
      // 0a3: iload 5
      // 0a5: aload 0
      // 0a6: getfield ua.Sb [I
      // 0a9: iload 1
      // 0aa: iaload
      // 0ab: imul
      // 0ac: iload 14
      // 0ae: irem
      // 0af: bipush -1
      // 0b0: ixor
      // 0b1: if_icmpeq 0cd
      // 0b4: aload 0
      // 0b5: getfield ua.Sb [I
      // 0b8: iload 1
      // 0b9: iaload
      // 0ba: iload 5
      // 0bc: imul
      // 0bd: iload 14
      // 0bf: irem
      // 0c0: ineg
      // 0c1: iload 14
      // 0c3: iadd
      // 0c4: ldc_w 1444777936
      // 0c7: ishl
      // 0c8: iload 5
      // 0ca: idiv
      // 0cb: istore 10
      // 0cd: bipush -1
      // 0ce: aload 0
      // 0cf: getfield ua.G [I
      // 0d2: iload 1
      // 0d3: iaload
      // 0d4: iload 3
      // 0d5: imul
      // 0d6: iload 15
      // 0d8: irem
      // 0d9: bipush -1
      // 0da: ixor
      // 0db: if_icmpne 0e2
      // 0de: goto 0f9
      // 0e1: athrow
      // 0e2: iload 15
      // 0e4: aload 0
      // 0e5: getfield ua.G [I
      // 0e8: iload 1
      // 0e9: iaload
      // 0ea: iload 3
      // 0eb: imul
      // 0ec: iload 15
      // 0ee: irem
      // 0ef: ineg
      // 0f0: iadd
      // 0f1: ldc_w -1176347888
      // 0f4: ishl
      // 0f5: iload 3
      // 0f6: idiv
      // 0f7: istore 11
      // 0f9: iload 5
      // 0fb: aload 0
      // 0fc: getfield ua.kb [I
      // 0ff: iload 1
      // 100: iaload
      // 101: iload 10
      // 103: ldc_w 7993200
      // 106: ishr
      // 107: isub
      // 108: imul
      // 109: iload 14
      // 10b: idiv
      // 10c: istore 5
      // 10e: iload 11
      // 110: ldc_w 826090000
      // 113: ishr
      // 114: ineg
      // 115: aload 0
      // 116: getfield ua.R [I
      // 119: iload 1
      // 11a: iaload
      // 11b: iadd
      // 11c: iload 3
      // 11d: imul
      // 11e: iload 15
      // 120: idiv
      // 121: istore 3
      // 122: iload 6
      // 124: aload 0
      // 125: getfield ua.u I
      // 128: imul
      // 129: iload 4
      // 12b: iadd
      // 12c: istore 14
      // 12e: iload 2
      // 12f: bipush -121
      // 131: if_icmple 135
      // 134: return
      // 135: iload 6
      // 137: bipush -1
      // 138: ixor
      // 139: aload 0
      // 13a: getfield ua.A I
      // 13d: bipush -1
      // 13e: ixor
      // 13f: if_icmple 16a
      // 142: iload 6
      // 144: ineg
      // 145: aload 0
      // 146: getfield ua.A I
      // 149: iadd
      // 14a: istore 16
      // 14c: iload 3
      // 14d: iload 16
      // 14f: isub
      // 150: istore 3
      // 151: bipush 0
      // 152: istore 6
      // 154: iload 14
      // 156: aload 0
      // 157: getfield ua.u I
      // 15a: iload 16
      // 15c: imul
      // 15d: iadd
      // 15e: istore 14
      // 160: iload 11
      // 162: iload 13
      // 164: iload 16
      // 166: imul
      // 167: iadd
      // 168: istore 11
      // 16a: aload 0
      // 16b: getfield ua.u I
      // 16e: iload 5
      // 170: ineg
      // 171: iadd
      // 172: istore 15
      // 174: aload 0
      // 175: getfield ua.hb I
      // 178: bipush -1
      // 179: ixor
      // 17a: iload 4
      // 17c: bipush -1
      // 17d: ixor
      // 17e: if_icmplt 185
      // 181: goto 1b1
      // 184: athrow
      // 185: iload 4
      // 187: ineg
      // 188: aload 0
      // 189: getfield ua.hb I
      // 18c: iadd
      // 18d: istore 16
      // 18f: bipush 0
      // 190: istore 4
      // 192: iload 10
      // 194: iload 16
      // 196: iload 12
      // 198: imul
      // 199: iadd
      // 19a: istore 10
      // 19c: iload 14
      // 19e: iload 16
      // 1a0: iadd
      // 1a1: istore 14
      // 1a3: iload 5
      // 1a5: iload 16
      // 1a7: isub
      // 1a8: istore 5
      // 1aa: iload 15
      // 1ac: iload 16
      // 1ae: iadd
      // 1af: istore 15
      // 1b1: aload 0
      // 1b2: getfield ua.Rb I
      // 1b5: bipush -1
      // 1b6: ixor
      // 1b7: iload 6
      // 1b9: iload 3
      // 1ba: iadd
      // 1bb: bipush -1
      // 1bc: ixor
      // 1bd: if_icmpge 1c4
      // 1c0: goto 1d3
      // 1c3: athrow
      // 1c4: iload 3
      // 1c5: bipush 1
      // 1c6: iload 3
      // 1c7: iadd
      // 1c8: iload 6
      // 1ca: aload 0
      // 1cb: getfield ua.Rb I
      // 1ce: ineg
      // 1cf: iadd
      // 1d0: iadd
      // 1d1: isub
      // 1d2: istore 3
      // 1d3: aload 0
      // 1d4: getfield ua.lb I
      // 1d7: bipush -1
      // 1d8: ixor
      // 1d9: iload 4
      // 1db: iload 5
      // 1dd: iadd
      // 1de: bipush -1
      // 1df: ixor
      // 1e0: if_icmplt 1ff
      // 1e3: bipush 1
      // 1e4: iload 4
      // 1e6: iadd
      // 1e7: iload 5
      // 1e9: aload 0
      // 1ea: getfield ua.lb I
      // 1ed: isub
      // 1ee: iadd
      // 1ef: istore 16
      // 1f1: iload 15
      // 1f3: iload 16
      // 1f5: iadd
      // 1f6: istore 15
      // 1f8: iload 5
      // 1fa: iload 16
      // 1fc: isub
      // 1fd: istore 5
      // 1ff: bipush 1
      // 200: istore 16
      // 202: bipush 1
      // 203: aload 0
      // 204: getfield ua.i Z
      // 207: ifne 20f
      // 20a: bipush 1
      // 20b: goto 210
      // 20e: athrow
      // 20f: bipush 0
      // 210: if_icmpeq 23a
      // 213: iload 13
      // 215: iload 13
      // 217: iadd
      // 218: istore 13
      // 21a: iload 15
      // 21c: aload 0
      // 21d: getfield ua.u I
      // 220: iadd
      // 221: istore 15
      // 223: bipush 0
      // 224: iload 6
      // 226: bipush 1
      // 227: iand
      // 228: if_icmpeq 237
      // 22b: iload 14
      // 22d: aload 0
      // 22e: getfield ua.u I
      // 231: iadd
      // 232: istore 14
      // 234: iinc 3 -1
      // 237: bipush 2
      // 238: istore 16
      // 23a: aload 0
      // 23b: iload 16
      // 23d: iload 11
      // 23f: iload 5
      // 241: bipush -61
      // 243: iload 13
      // 245: iload 8
      // 247: iload 12
      // 249: iload 3
      // 24a: iload 14
      // 24c: aload 0
      // 24d: getfield ua.ob [[I
      // 250: iload 1
      // 251: aaload
      // 252: bipush 0
      // 253: iload 10
      // 255: iload 15
      // 257: iload 7
      // 259: aload 0
      // 25a: getfield ua.rb [I
      // 25d: invokespecial ua.a (IIIBIIIII[IIIII[I)V
      // 260: goto 271
      // 263: astore 8
      // 265: getstatic java/lang/System.out Ljava/io/PrintStream;
      // 268: getstatic ua.Yb [Ljava/lang/String;
      // 26b: bipush 16
      // 26d: aaload
      // 26e: invokevirtual java/io/PrintStream.println (Ljava/lang/String;)V
      // 271: goto 2d2
      // 274: astore 8
      // 276: aload 8
      // 278: new java/lang/StringBuilder
      // 27b: dup
      // 27c: invokespecial java/lang/StringBuilder.<init> ()V
      // 27f: getstatic ua.Yb [Ljava/lang/String;
      // 282: bipush 54
      // 284: aaload
      // 285: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 288: iload 1
      // 289: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 28c: bipush 44
      // 28e: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 291: iload 2
      // 292: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 295: bipush 44
      // 297: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 29a: iload 3
      // 29b: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 29e: bipush 44
      // 2a0: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 2a3: iload 4
      // 2a5: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 2a8: bipush 44
      // 2aa: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 2ad: iload 5
      // 2af: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 2b2: bipush 44
      // 2b4: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 2b7: iload 6
      // 2b9: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 2bc: bipush 44
      // 2be: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 2c1: iload 7
      // 2c3: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 2c6: bipush 41
      // 2c8: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 2cb: invokevirtual java/lang/StringBuilder.toString ()Ljava/lang/String;
      // 2ce: invokestatic i.a (Ljava/lang/Throwable;Ljava/lang/String;)Lla;
      // 2d1: athrow
      // 2d2: return
   }

   private final void a(int param1, int[] param2, int param3, byte param4, int param5, int param6, int param7, int param8, byte[] param9, int param10) {
      // $VF: Couldn't be decompiled
      // Please report this to the Vineflower issue tracker, at https://github.com/Vineflower/vineflower/issues with a copy of the class file (if you have the rights to distribute it!)
      // java.lang.RuntimeException: parsing failure!
      //   at org.jetbrains.java.decompiler.modules.decompiler.decompose.DomHelper.parseGraph(DomHelper.java:211)
      //   at org.jetbrains.java.decompiler.main.rels.MethodProcessor.codeToJava(MethodProcessor.java:166)
      //
      // Bytecode:
      // 000: getstatic client.vh Z
      // 003: istore 14
      // 005: getstatic ua.g I
      // 008: bipush 1
      // 009: iadd
      // 00a: putstatic ua.g I
      // 00d: iload 7
      // 00f: ldc_w -1201305182
      // 012: ishr
      // 013: ineg
      // 014: istore 11
      // 016: iload 7
      // 018: bipush 3
      // 019: iand
      // 01a: ineg
      // 01b: istore 7
      // 01d: iload 6
      // 01f: ineg
      // 020: istore 12
      // 022: iload 12
      // 024: bipush -1
      // 025: ixor
      // 026: bipush -1
      // 027: if_icmple 133
      // 02a: iload 11
      // 02c: iload 14
      // 02e: ifne 166
      // 031: istore 13
      // 033: iload 13
      // 035: bipush -1
      // 036: ixor
      // 037: bipush -1
      // 038: if_icmple 0dc
      // 03b: bipush 0
      // 03c: aload 9
      // 03e: iload 8
      // 040: iinc 8 1
      // 043: baload
      // 044: iload 14
      // 046: ifne 0e5
      // 049: if_icmpeq 060
      // 04c: goto 050
      // 04f: athrow
      // 050: aload 2
      // 051: iload 3
      // 052: iinc 3 1
      // 055: iload 1
      // 056: iastore
      // 057: iload 14
      // 059: ifeq 067
      // 05c: goto 060
      // 05f: athrow
      // 060: iinc 3 1
      // 063: goto 067
      // 066: athrow
      // 067: aload 9
      // 069: iload 8
      // 06b: iinc 8 1
      // 06e: baload
      // 06f: bipush -1
      // 070: ixor
      // 071: bipush -1
      // 072: if_icmpne 081
      // 075: iinc 3 1
      // 078: iload 14
      // 07a: ifeq 08c
      // 07d: goto 081
      // 080: athrow
      // 081: aload 2
      // 082: iload 3
      // 083: iinc 3 1
      // 086: iload 1
      // 087: iastore
      // 088: goto 08c
      // 08b: athrow
      // 08c: bipush 0
      // 08d: aload 9
      // 08f: iload 8
      // 091: iinc 8 1
      // 094: baload
      // 095: if_icmpne 0a4
      // 098: iinc 3 1
      // 09b: iload 14
      // 09d: ifeq 0af
      // 0a0: goto 0a4
      // 0a3: athrow
      // 0a4: aload 2
      // 0a5: iload 3
      // 0a6: iinc 3 1
      // 0a9: iload 1
      // 0aa: iastore
      // 0ab: goto 0af
      // 0ae: athrow
      // 0af: aload 9
      // 0b1: iload 8
      // 0b3: iinc 8 1
      // 0b6: baload
      // 0b7: bipush -1
      // 0b8: ixor
      // 0b9: bipush -1
      // 0ba: if_icmpeq 0cd
      // 0bd: aload 2
      // 0be: iload 3
      // 0bf: iinc 3 1
      // 0c2: iload 1
      // 0c3: iastore
      // 0c4: iload 14
      // 0c6: ifeq 0d4
      // 0c9: goto 0cd
      // 0cc: athrow
      // 0cd: iinc 3 1
      // 0d0: goto 0d4
      // 0d3: athrow
      // 0d4: iinc 13 1
      // 0d7: iload 14
      // 0d9: ifeq 033
      // 0dc: iload 7
      // 0de: istore 13
      // 0e0: iload 13
      // 0e2: bipush -1
      // 0e3: ixor
      // 0e4: bipush -1
      // 0e5: if_icmple 11f
      // 0e8: aload 9
      // 0ea: iload 8
      // 0ec: iinc 8 1
      // 0ef: baload
      // 0f0: iload 14
      // 0f2: ifne 12a
      // 0f5: goto 0f9
      // 0f8: athrow
      // 0f9: ifne 10c
      // 0fc: goto 100
      // 0ff: athrow
      // 100: iinc 3 1
      // 103: iload 14
      // 105: ifeq 117
      // 108: goto 10c
      // 10b: athrow
      // 10c: aload 2
      // 10d: iload 3
      // 10e: iinc 3 1
      // 111: iload 1
      // 112: iastore
      // 113: goto 117
      // 116: athrow
      // 117: iinc 13 1
      // 11a: iload 14
      // 11c: ifeq 0e0
      // 11f: iload 8
      // 121: iload 10
      // 123: iadd
      // 124: istore 8
      // 126: iload 3
      // 127: iload 5
      // 129: iadd
      // 12a: istore 3
      // 12b: iinc 12 1
      // 12e: iload 14
      // 130: ifeq 022
      // 133: goto 15b
      // 136: astore 11
      // 138: getstatic java/lang/System.out Ljava/io/PrintStream;
      // 13b: new java/lang/StringBuilder
      // 13e: dup
      // 13f: invokespecial java/lang/StringBuilder.<init> ()V
      // 142: getstatic ua.Yb [Ljava/lang/String;
      // 145: bipush 14
      // 147: aaload
      // 148: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 14b: aload 11
      // 14d: invokevirtual java/lang/StringBuilder.append (Ljava/lang/Object;)Ljava/lang/StringBuilder;
      // 150: invokevirtual java/lang/StringBuilder.toString ()Ljava/lang/String;
      // 153: invokevirtual java/io/PrintStream.println (Ljava/lang/String;)V
      // 156: aload 11
      // 158: invokevirtual java/lang/Exception.printStackTrace ()V
      // 15b: bipush 82
      // 15d: bipush -45
      // 15f: iload 4
      // 161: isub
      // 162: bipush 48
      // 164: idiv
      // 165: irem
      // 166: istore 11
      // 168: goto 209
      // 16b: astore 11
      // 16d: aload 11
      // 16f: new java/lang/StringBuilder
      // 172: dup
      // 173: invokespecial java/lang/StringBuilder.<init> ()V
      // 176: getstatic ua.Yb [Ljava/lang/String;
      // 179: bipush 13
      // 17b: aaload
      // 17c: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 17f: iload 1
      // 180: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 183: bipush 44
      // 185: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 188: aload 2
      // 189: ifnull 195
      // 18c: getstatic ua.Yb [Ljava/lang/String;
      // 18f: bipush 1
      // 190: aaload
      // 191: goto 19a
      // 194: athrow
      // 195: getstatic ua.Yb [Ljava/lang/String;
      // 198: bipush 0
      // 199: aaload
      // 19a: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 19d: bipush 44
      // 19f: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 1a2: iload 3
      // 1a3: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 1a6: bipush 44
      // 1a8: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 1ab: iload 4
      // 1ad: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 1b0: bipush 44
      // 1b2: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 1b5: iload 5
      // 1b7: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 1ba: bipush 44
      // 1bc: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 1bf: iload 6
      // 1c1: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 1c4: bipush 44
      // 1c6: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 1c9: iload 7
      // 1cb: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 1ce: bipush 44
      // 1d0: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 1d3: iload 8
      // 1d5: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 1d8: bipush 44
      // 1da: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 1dd: aload 9
      // 1df: ifnull 1eb
      // 1e2: getstatic ua.Yb [Ljava/lang/String;
      // 1e5: bipush 1
      // 1e6: aaload
      // 1e7: goto 1f0
      // 1ea: athrow
      // 1eb: getstatic ua.Yb [Ljava/lang/String;
      // 1ee: bipush 0
      // 1ef: aaload
      // 1f0: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 1f3: bipush 44
      // 1f5: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 1f8: iload 10
      // 1fa: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 1fd: bipush 41
      // 1ff: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 202: invokevirtual java/lang/StringBuilder.toString ()Ljava/lang/String;
      // 205: invokestatic i.a (Ljava/lang/Throwable;Ljava/lang/String;)Lla;
      // 208: athrow
      // 209: return
   }

   private final void a(
      int param1, int[] param2, int param3, int param4, int[] param5, int param6, int param7, boolean param8, byte[] param9, int param10, int param11
   ) {
      // $VF: Couldn't be decompiled
      // Please report this to the Vineflower issue tracker, at https://github.com/Vineflower/vineflower/issues with a copy of the class file (if you have the rights to distribute it!)
      // java.lang.RuntimeException: parsing failure!
      //   at org.jetbrains.java.decompiler.modules.decompiler.decompose.DomHelper.parseGraph(DomHelper.java:211)
      //   at org.jetbrains.java.decompiler.main.rels.MethodProcessor.codeToJava(MethodProcessor.java:166)
      //
      // Bytecode:
      // 000: getstatic client.vh Z
      // 003: istore 16
      // 005: iload 8
      // 007: bipush 1
      // 008: if_icmpeq 014
      // 00b: aload 0
      // 00c: bipush 1
      // 00d: putfield ua.i Z
      // 010: goto 014
      // 013: athrow
      // 014: getstatic ua.Ub I
      // 017: bipush 1
      // 018: iadd
      // 019: putstatic ua.Ub I
      // 01c: iload 10
      // 01e: ldc_w -1243049534
      // 021: ishr
      // 022: ineg
      // 023: istore 12
      // 025: bipush 3
      // 026: iload 10
      // 028: iand
      // 029: ineg
      // 02a: istore 10
      // 02c: iload 7
      // 02e: ineg
      // 02f: istore 13
      // 031: iload 13
      // 033: bipush -1
      // 034: ixor
      // 035: bipush -1
      // 036: if_icmple 17c
      // 039: iload 16
      // 03b: ifne 238
      // 03e: iload 12
      // 040: istore 14
      // 042: bipush 0
      // 043: iload 14
      // 045: if_icmple 11b
      // 048: aload 9
      // 04a: iload 3
      // 04b: iinc 3 1
      // 04e: baload
      // 04f: istore 15
      // 051: iload 15
      // 053: bipush -1
      // 054: ixor
      // 055: bipush -1
      // 056: iload 16
      // 058: ifne 122
      // 05b: if_icmpne 06e
      // 05e: goto 062
      // 061: athrow
      // 062: iinc 1 1
      // 065: iload 16
      // 067: ifeq 083
      // 06a: goto 06e
      // 06d: athrow
      // 06e: aload 5
      // 070: iload 1
      // 071: iinc 1 1
      // 074: aload 2
      // 075: iload 15
      // 077: sipush 255
      // 07a: invokestatic ib.a (II)I
      // 07d: iaload
      // 07e: iastore
      // 07f: goto 083
      // 082: athrow
      // 083: aload 9
      // 085: iload 3
      // 086: iinc 3 1
      // 089: baload
      // 08a: istore 15
      // 08c: iload 15
      // 08e: ifne 09d
      // 091: iinc 1 1
      // 094: iload 16
      // 096: ifeq 0b2
      // 099: goto 09d
      // 09c: athrow
      // 09d: aload 5
      // 09f: iload 1
      // 0a0: iinc 1 1
      // 0a3: aload 2
      // 0a4: iload 15
      // 0a6: sipush 255
      // 0a9: invokestatic ib.a (II)I
      // 0ac: iaload
      // 0ad: iastore
      // 0ae: goto 0b2
      // 0b1: athrow
      // 0b2: aload 9
      // 0b4: iload 3
      // 0b5: iinc 3 1
      // 0b8: baload
      // 0b9: istore 15
      // 0bb: iload 15
      // 0bd: bipush -1
      // 0be: ixor
      // 0bf: bipush -1
      // 0c0: if_icmpeq 0dd
      // 0c3: aload 5
      // 0c5: iload 1
      // 0c6: iinc 1 1
      // 0c9: aload 2
      // 0ca: sipush 255
      // 0cd: iload 15
      // 0cf: invokestatic ib.a (II)I
      // 0d2: iaload
      // 0d3: iastore
      // 0d4: iload 16
      // 0d6: ifeq 0e4
      // 0d9: goto 0dd
      // 0dc: athrow
      // 0dd: iinc 1 1
      // 0e0: goto 0e4
      // 0e3: athrow
      // 0e4: aload 9
      // 0e6: iload 3
      // 0e7: iinc 3 1
      // 0ea: baload
      // 0eb: istore 15
      // 0ed: iload 15
      // 0ef: ifeq 10c
      // 0f2: aload 5
      // 0f4: iload 1
      // 0f5: iinc 1 1
      // 0f8: aload 2
      // 0f9: iload 15
      // 0fb: sipush 255
      // 0fe: invokestatic ib.a (II)I
      // 101: iaload
      // 102: iastore
      // 103: iload 16
      // 105: ifeq 113
      // 108: goto 10c
      // 10b: athrow
      // 10c: iinc 1 1
      // 10f: goto 113
      // 112: athrow
      // 113: iinc 14 1
      // 116: iload 16
      // 118: ifeq 042
      // 11b: iload 10
      // 11d: istore 14
      // 11f: bipush 0
      // 120: iload 14
      // 122: if_icmple 166
      // 125: aload 9
      // 127: iload 3
      // 128: iinc 3 1
      // 12b: baload
      // 12c: istore 15
      // 12e: bipush 0
      // 12f: iload 15
      // 131: iload 16
      // 133: ifne 174
      // 136: if_icmpeq 157
      // 139: goto 13d
      // 13c: athrow
      // 13d: aload 5
      // 13f: iload 1
      // 140: iinc 1 1
      // 143: aload 2
      // 144: iload 15
      // 146: sipush 255
      // 149: invokestatic ib.a (II)I
      // 14c: iaload
      // 14d: iastore
      // 14e: iload 16
      // 150: ifeq 15e
      // 153: goto 157
      // 156: athrow
      // 157: iinc 1 1
      // 15a: goto 15e
      // 15d: athrow
      // 15e: iinc 14 1
      // 161: iload 16
      // 163: ifeq 11f
      // 166: iload 1
      // 167: iload 11
      // 169: iadd
      // 16a: istore 1
      // 16b: iload 3
      // 16c: iload 4
      // 16e: iadd
      // 16f: istore 3
      // 170: iload 13
      // 172: iload 6
      // 174: iadd
      // 175: istore 13
      // 177: iload 16
      // 179: ifeq 031
      // 17c: goto 238
      // 17f: astore 12
      // 181: aload 12
      // 183: new java/lang/StringBuilder
      // 186: dup
      // 187: invokespecial java/lang/StringBuilder.<init> ()V
      // 18a: getstatic ua.Yb [Ljava/lang/String;
      // 18d: bipush 82
      // 18f: aaload
      // 190: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 193: iload 1
      // 194: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 197: bipush 44
      // 199: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 19c: aload 2
      // 19d: ifnull 1a9
      // 1a0: getstatic ua.Yb [Ljava/lang/String;
      // 1a3: bipush 1
      // 1a4: aaload
      // 1a5: goto 1ae
      // 1a8: athrow
      // 1a9: getstatic ua.Yb [Ljava/lang/String;
      // 1ac: bipush 0
      // 1ad: aaload
      // 1ae: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 1b1: bipush 44
      // 1b3: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 1b6: iload 3
      // 1b7: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 1ba: bipush 44
      // 1bc: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 1bf: iload 4
      // 1c1: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 1c4: bipush 44
      // 1c6: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 1c9: aload 5
      // 1cb: ifnull 1d7
      // 1ce: getstatic ua.Yb [Ljava/lang/String;
      // 1d1: bipush 1
      // 1d2: aaload
      // 1d3: goto 1dc
      // 1d6: athrow
      // 1d7: getstatic ua.Yb [Ljava/lang/String;
      // 1da: bipush 0
      // 1db: aaload
      // 1dc: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 1df: bipush 44
      // 1e1: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 1e4: iload 6
      // 1e6: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 1e9: bipush 44
      // 1eb: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 1ee: iload 7
      // 1f0: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 1f3: bipush 44
      // 1f5: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 1f8: iload 8
      // 1fa: invokevirtual java/lang/StringBuilder.append (Z)Ljava/lang/StringBuilder;
      // 1fd: bipush 44
      // 1ff: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 202: aload 9
      // 204: ifnull 210
      // 207: getstatic ua.Yb [Ljava/lang/String;
      // 20a: bipush 1
      // 20b: aaload
      // 20c: goto 215
      // 20f: athrow
      // 210: getstatic ua.Yb [Ljava/lang/String;
      // 213: bipush 0
      // 214: aaload
      // 215: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 218: bipush 44
      // 21a: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 21d: iload 10
      // 21f: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 222: bipush 44
      // 224: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 227: iload 11
      // 229: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 22c: bipush 41
      // 22e: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 231: invokevirtual java/lang/StringBuilder.toString ()Ljava/lang/String;
      // 234: invokestatic i.a (Ljava/lang/Throwable;Ljava/lang/String;)Lla;
      // 237: athrow
      // 238: return
   }

   // $VF: Handled exception range with multiple entry points by splitting it
   // $VF: Inserted dummy exception handlers to handle obfuscated exceptions
   final void b(int var1, int var2) {
      boolean var9 = client.vh;

      RuntimeException var10000;
      label82: {
         try {
            b++;
            if (null == this.gb[var1]) {
               return;
            }
         } catch (RuntimeException var15) {
            var10000 = var15;
            boolean var10001 = false;
            break label82;
         }

         try {
            int var16 = this.kb[var1] * this.R[var1];
            byte[] var4 = this.gb[var1];
            int[] var5 = this.Y[var1];
            int[] var6 = new int[var16];
            int var7 = 0;

            label71: {
               while (~var7 > ~var16) {
                  int var8 = var5[255 & var4[var7]];

                  label88: {
                     label89: {
                        try {
                           var17 = var8;
                           if (var9) {
                              break label71;
                           }

                           if (var8 != 0) {
                              break label89;
                           }
                        } catch (RuntimeException var13) {
                           throw var13;
                        }

                        var8 = 1;
                        boolean var18 = var9;

                        try {
                           if (!var18) {
                              break label88;
                           }
                        } catch (RuntimeException var12) {
                           boolean var21 = false;
                           throw var12;
                        }
                     }

                     try {
                        if (-16711936 != ~var8) {
                           break label88;
                        }
                     } catch (RuntimeException var11) {
                        boolean var22 = false;
                        throw var11;
                     }

                     var8 = 0;
                  }

                  var6[var7] = var8;
                  var7++;
                  if (var9) {
                     break;
                  }
               }

               var17 = var2;
            }

            try {
               if (var17 != -342059728) {
                  this.Eb = (int[])null;
               }
            } catch (RuntimeException var10) {
               throw var10;
            }

            this.ob[var1] = var6;
            this.gb[var1] = null;
            this.Y[var1] = null;
            return;
         } catch (RuntimeException var14) {
            var10000 = var14;
            boolean var20 = false;
         }
      }

      RuntimeException var3 = var10000;
      throw i.a(var3, Yb[77] + var1 + ',' + var2 + ')');
   }

   final void a(int param1, int param2, int param3, int param4, int param5, int param6) {
      // $VF: Couldn't be decompiled
      // Please report this to the Vineflower issue tracker, at https://github.com/Vineflower/vineflower/issues with a copy of the class file (if you have the rights to distribute it!)
      // java.lang.RuntimeException: parsing failure!
      //   at org.jetbrains.java.decompiler.modules.decompiler.decompose.DomHelper.parseGraph(DomHelper.java:211)
      //   at org.jetbrains.java.decompiler.main.rels.MethodProcessor.codeToJava(MethodProcessor.java:166)
      //
      // Bytecode:
      // 000: getstatic client.vh Z
      // 003: istore 48
      // 005: getstatic ua.S I
      // 008: bipush 1
      // 009: iadd
      // 00a: putstatic ua.S I
      // 00d: aload 0
      // 00e: getfield ua.u I
      // 011: istore 7
      // 013: aload 0
      // 014: getfield ua.k I
      // 017: istore 8
      // 019: aload 0
      // 01a: getfield ua.Hb [I
      // 01d: ifnull 024
      // 020: goto 079
      // 023: athrow
      // 024: aload 0
      // 025: sipush 512
      // 028: newarray 10
      // 02a: putfield ua.Hb [I
      // 02d: bipush 0
      // 02e: istore 9
      // 030: iload 9
      // 032: sipush 256
      // 035: if_icmpge 079
      // 038: aload 0
      // 039: getfield ua.Hb [I
      // 03c: iload 9
      // 03e: iload 9
      // 040: i2d
      // 041: ldc2_w 0.02454369
      // 044: dmul
      // 045: invokestatic java/lang/Math.sin (D)D
      // 048: ldc2_w 32768.0
      // 04b: dmul
      // 04c: d2i
      // 04d: iastore
      // 04e: aload 0
      // 04f: getfield ua.Hb [I
      // 052: sipush 256
      // 055: iload 9
      // 057: iadd
      // 058: iload 9
      // 05a: i2d
      // 05b: ldc2_w 0.02454369
      // 05e: dmul
      // 05f: invokestatic java/lang/Math.cos (D)D
      // 062: ldc2_w 32768.0
      // 065: dmul
      // 066: d2i
      // 067: iastore
      // 068: iinc 9 1
      // 06b: iload 48
      // 06d: ifne 084
      // 070: iload 48
      // 072: ifeq 030
      // 075: goto 079
      // 078: athrow
      // 079: aload 0
      // 07a: getfield ua.Eb [I
      // 07d: iload 1
      // 07e: iaload
      // 07f: ineg
      // 080: bipush 2
      // 081: idiv
      // 082: istore 9
      // 084: aload 0
      // 085: getfield ua.qb [I
      // 088: iload 1
      // 089: iaload
      // 08a: ineg
      // 08b: bipush 2
      // 08c: idiv
      // 08d: istore 10
      // 08f: aload 0
      // 090: getfield ua.Qb [Z
      // 093: iload 1
      // 094: baload
      // 095: ifne 09c
      // 098: goto 0b2
      // 09b: athrow
      // 09c: iload 9
      // 09e: aload 0
      // 09f: getfield ua.Sb [I
      // 0a2: iload 1
      // 0a3: iaload
      // 0a4: iadd
      // 0a5: istore 9
      // 0a7: iload 10
      // 0a9: aload 0
      // 0aa: getfield ua.G [I
      // 0ad: iload 1
      // 0ae: iaload
      // 0af: iadd
      // 0b0: istore 10
      // 0b2: aload 0
      // 0b3: getfield ua.kb [I
      // 0b6: iload 1
      // 0b7: iaload
      // 0b8: iload 9
      // 0ba: iadd
      // 0bb: istore 11
      // 0bd: aload 0
      // 0be: getfield ua.R [I
      // 0c1: iload 1
      // 0c2: iaload
      // 0c3: iload 10
      // 0c5: iadd
      // 0c6: istore 12
      // 0c8: iload 11
      // 0ca: istore 13
      // 0cc: iload 10
      // 0ce: istore 14
      // 0d0: iload 9
      // 0d2: istore 15
      // 0d4: iload 6
      // 0d6: sipush 255
      // 0d9: iand
      // 0da: istore 6
      // 0dc: iload 12
      // 0de: istore 16
      // 0e0: aload 0
      // 0e1: getfield ua.Hb [I
      // 0e4: iload 6
      // 0e6: iaload
      // 0e7: iload 5
      // 0e9: imul
      // 0ea: istore 17
      // 0ec: aload 0
      // 0ed: getfield ua.Hb [I
      // 0f0: iload 6
      // 0f2: sipush 256
      // 0f5: iadd
      // 0f6: iaload
      // 0f7: iload 5
      // 0f9: imul
      // 0fa: istore 18
      // 0fc: iload 3
      // 0fd: iload 18
      // 0ff: iload 9
      // 101: imul
      // 102: iload 10
      // 104: iload 17
      // 106: imul
      // 107: iadd
      // 108: ldc_w 2041339190
      // 10b: ishr
      // 10c: iadd
      // 10d: istore 19
      // 10f: iload 2
      // 110: iload 9
      // 112: iload 17
      // 114: imul
      // 115: ineg
      // 116: iload 10
      // 118: iload 18
      // 11a: imul
      // 11b: iadd
      // 11c: ldc_w -1101614986
      // 11f: ishr
      // 120: ineg
      // 121: isub
      // 122: istore 20
      // 124: iload 13
      // 126: iload 18
      // 128: imul
      // 129: iload 14
      // 12b: iload 17
      // 12d: imul
      // 12e: iadd
      // 12f: ldc_w -1412824714
      // 132: ishr
      // 133: iload 3
      // 134: iadd
      // 135: istore 21
      // 137: iload 2
      // 138: iload 18
      // 13a: iload 14
      // 13c: imul
      // 13d: iload 13
      // 13f: iload 17
      // 141: imul
      // 142: isub
      // 143: ldc_w -1277097834
      // 146: ishr
      // 147: iadd
      // 148: istore 22
      // 14a: iload 17
      // 14c: iload 12
      // 14e: imul
      // 14f: iload 18
      // 151: iload 11
      // 153: imul
      // 154: iadd
      // 155: ldc_w 1894119030
      // 158: ishr
      // 159: iload 3
      // 15a: iadd
      // 15b: istore 23
      // 15d: iload 17
      // 15f: iload 11
      // 161: imul
      // 162: ineg
      // 163: iload 18
      // 165: iload 12
      // 167: imul
      // 168: iadd
      // 169: ldc_w 195973654
      // 16c: ishr
      // 16d: iload 2
      // 16e: iadd
      // 16f: istore 24
      // 171: iload 15
      // 173: iload 18
      // 175: imul
      // 176: iload 17
      // 178: iload 16
      // 17a: imul
      // 17b: iadd
      // 17c: ldc_w -352508522
      // 17f: ishr
      // 180: iload 3
      // 181: iadd
      // 182: istore 25
      // 184: iload 2
      // 185: iload 16
      // 187: iload 18
      // 189: imul
      // 18a: iload 15
      // 18c: iload 17
      // 18e: imul
      // 18f: isub
      // 190: ldc_w -223287050
      // 193: ishr
      // 194: ineg
      // 195: isub
      // 196: istore 26
      // 198: iload 5
      // 19a: bipush -1
      // 19b: ixor
      // 19c: sipush -193
      // 19f: if_icmpne 1b8
      // 1a2: bipush 63
      // 1a4: getstatic client.ef I
      // 1a7: iand
      // 1a8: bipush -1
      // 1a9: ixor
      // 1aa: bipush 63
      // 1ac: iload 6
      // 1ae: iand
      // 1af: bipush -1
      // 1b0: ixor
      // 1b1: if_icmpeq 1e5
      // 1b4: goto 1b8
      // 1b7: athrow
      // 1b8: iload 5
      // 1ba: bipush -1
      // 1bb: ixor
      // 1bc: sipush -129
      // 1bf: if_icmpne 1d4
      // 1c2: goto 1c6
      // 1c5: athrow
      // 1c6: iload 6
      // 1c8: putstatic client.ef I
      // 1cb: iload 48
      // 1cd: ifeq 1f1
      // 1d0: goto 1d4
      // 1d3: athrow
      // 1d4: getstatic da.M I
      // 1d7: bipush 1
      // 1d8: iadd
      // 1d9: putstatic da.M I
      // 1dc: iload 48
      // 1de: ifeq 1f1
      // 1e1: goto 1e5
      // 1e4: athrow
      // 1e5: getstatic nb.g I
      // 1e8: bipush 1
      // 1e9: iadd
      // 1ea: putstatic nb.g I
      // 1ed: goto 1f1
      // 1f0: athrow
      // 1f1: iload 20
      // 1f3: istore 27
      // 1f5: iload 20
      // 1f7: istore 28
      // 1f9: iload 27
      // 1fb: iload 22
      // 1fd: if_icmple 209
      // 200: iload 22
      // 202: istore 27
      // 204: iload 48
      // 206: ifeq 21c
      // 209: iload 28
      // 20b: bipush -1
      // 20c: ixor
      // 20d: iload 22
      // 20f: bipush -1
      // 210: ixor
      // 211: if_icmple 21c
      // 214: goto 218
      // 217: athrow
      // 218: iload 22
      // 21a: istore 28
      // 21c: iload 24
      // 21e: bipush -1
      // 21f: ixor
      // 220: iload 27
      // 222: bipush -1
      // 223: ixor
      // 224: if_icmple 230
      // 227: iload 24
      // 229: istore 27
      // 22b: iload 48
      // 22d: ifeq 23f
      // 230: iload 24
      // 232: iload 28
      // 234: if_icmple 23f
      // 237: goto 23b
      // 23a: athrow
      // 23b: iload 24
      // 23d: istore 28
      // 23f: iload 26
      // 241: bipush -1
      // 242: ixor
      // 243: iload 27
      // 245: bipush -1
      // 246: ixor
      // 247: if_icmple 253
      // 24a: iload 26
      // 24c: istore 27
      // 24e: iload 48
      // 250: ifeq 266
      // 253: iload 28
      // 255: bipush -1
      // 256: ixor
      // 257: iload 26
      // 259: bipush -1
      // 25a: ixor
      // 25b: if_icmple 266
      // 25e: goto 262
      // 261: athrow
      // 262: iload 26
      // 264: istore 28
      // 266: iload 27
      // 268: bipush -1
      // 269: ixor
      // 26a: aload 0
      // 26b: getfield ua.A I
      // 26e: bipush -1
      // 26f: ixor
      // 270: if_icmpgt 277
      // 273: goto 27d
      // 276: athrow
      // 277: aload 0
      // 278: getfield ua.A I
      // 27b: istore 27
      // 27d: aload 0
      // 27e: getfield ua.Xb [I
      // 281: ifnull 29c
      // 284: aload 0
      // 285: getfield ua.Xb [I
      // 288: arraylength
      // 289: bipush -1
      // 28a: ixor
      // 28b: iload 8
      // 28d: bipush 1
      // 28e: iadd
      // 28f: bipush -1
      // 290: ixor
      // 291: if_icmpne 29c
      // 294: goto 298
      // 297: athrow
      // 298: goto 2d8
      // 29b: athrow
      // 29c: aload 0
      // 29d: iload 8
      // 29f: bipush 1
      // 2a0: iadd
      // 2a1: newarray 10
      // 2a3: putfield ua.tb [I
      // 2a6: aload 0
      // 2a7: bipush 1
      // 2a8: iload 8
      // 2aa: iadd
      // 2ab: newarray 10
      // 2ad: putfield ua.M [I
      // 2b0: aload 0
      // 2b1: iload 8
      // 2b3: bipush 1
      // 2b4: iadd
      // 2b5: newarray 10
      // 2b7: putfield ua.t [I
      // 2ba: aload 0
      // 2bb: bipush 1
      // 2bc: iload 8
      // 2be: iadd
      // 2bf: newarray 10
      // 2c1: putfield ua.Tb [I
      // 2c4: aload 0
      // 2c5: bipush 1
      // 2c6: iload 8
      // 2c8: iadd
      // 2c9: newarray 10
      // 2cb: putfield ua.Wb [I
      // 2ce: aload 0
      // 2cf: iload 8
      // 2d1: bipush -1
      // 2d2: isub
      // 2d3: newarray 10
      // 2d5: putfield ua.Xb [I
      // 2d8: aload 0
      // 2d9: getfield ua.Rb I
      // 2dc: bipush -1
      // 2dd: ixor
      // 2de: iload 28
      // 2e0: bipush -1
      // 2e1: ixor
      // 2e2: if_icmple 2eb
      // 2e5: aload 0
      // 2e6: getfield ua.Rb I
      // 2e9: istore 28
      // 2eb: iload 27
      // 2ed: istore 29
      // 2ef: iload 28
      // 2f1: bipush -1
      // 2f2: ixor
      // 2f3: iload 29
      // 2f5: bipush -1
      // 2f6: ixor
      // 2f7: if_icmpgt 316
      // 2fa: aload 0
      // 2fb: getfield ua.Xb [I
      // 2fe: iload 29
      // 300: ldc_w 99999999
      // 303: iastore
      // 304: aload 0
      // 305: getfield ua.t [I
      // 308: iload 29
      // 30a: ldc_w -99999999
      // 30d: iastore
      // 30e: iinc 29 1
      // 311: iload 48
      // 313: ifeq 2ef
      // 316: bipush 0
      // 317: istore 32
      // 319: bipush 0
      // 31a: istore 34
      // 31c: bipush 0
      // 31d: istore 36
      // 31f: aload 0
      // 320: getfield ua.kb [I
      // 323: iload 1
      // 324: iaload
      // 325: istore 37
      // 327: bipush -1
      // 328: iload 37
      // 32a: iadd
      // 32b: istore 11
      // 32d: aload 0
      // 32e: getfield ua.R [I
      // 331: iload 1
      // 332: iaload
      // 333: istore 38
      // 335: iload 37
      // 337: bipush -1
      // 338: iadd
      // 339: istore 13
      // 33b: bipush 0
      // 33c: istore 9
      // 33e: bipush 0
      // 33f: istore 10
      // 341: bipush 0
      // 342: istore 14
      // 344: bipush 0
      // 345: istore 15
      // 347: iload 38
      // 349: bipush -1
      // 34a: iadd
      // 34b: istore 12
      // 34d: bipush -1
      // 34e: iload 38
      // 350: iadd
      // 351: istore 16
      // 353: iload 20
      // 355: iload 26
      // 357: if_icmple 377
      // 35a: iload 20
      // 35c: istore 30
      // 35e: iload 16
      // 360: ldc_w 840930536
      // 363: ishl
      // 364: istore 35
      // 366: iload 26
      // 368: istore 29
      // 36a: iload 25
      // 36c: ldc_w -731169944
      // 36f: ishl
      // 370: istore 31
      // 372: iload 48
      // 374: ifeq 38f
      // 377: iload 26
      // 379: istore 30
      // 37b: iload 20
      // 37d: istore 29
      // 37f: iload 10
      // 381: ldc_w 813501160
      // 384: ishl
      // 385: istore 35
      // 387: iload 19
      // 389: ldc_w 1063907752
      // 38c: ishl
      // 38d: istore 31
      // 38f: iload 26
      // 391: bipush -1
      // 392: ixor
      // 393: iload 20
      // 395: bipush -1
      // 396: ixor
      // 397: if_icmpne 39e
      // 39a: goto 3c3
      // 39d: athrow
      // 39e: iload 10
      // 3a0: ineg
      // 3a1: iload 16
      // 3a3: iadd
      // 3a4: ldc_w 822762216
      // 3a7: ishl
      // 3a8: iload 26
      // 3aa: iload 20
      // 3ac: ineg
      // 3ad: iadd
      // 3ae: idiv
      // 3af: istore 36
      // 3b1: iload 19
      // 3b3: ineg
      // 3b4: iload 25
      // 3b6: iadd
      // 3b7: ldc_w -1960375352
      // 3ba: ishl
      // 3bb: iload 26
      // 3bd: iload 20
      // 3bf: isub
      // 3c0: idiv
      // 3c1: istore 32
      // 3c3: bipush -1
      // 3c4: iload 29
      // 3c6: bipush -1
      // 3c7: ixor
      // 3c8: if_icmplt 3cf
      // 3cb: goto 3e6
      // 3ce: athrow
      // 3cf: iload 31
      // 3d1: iload 32
      // 3d3: iload 29
      // 3d5: imul
      // 3d6: isub
      // 3d7: istore 31
      // 3d9: iload 35
      // 3db: iload 29
      // 3dd: iload 36
      // 3df: imul
      // 3e0: isub
      // 3e1: istore 35
      // 3e3: bipush 0
      // 3e4: istore 29
      // 3e6: iload 30
      // 3e8: bipush -1
      // 3e9: ixor
      // 3ea: bipush -1
      // 3eb: iload 8
      // 3ed: iadd
      // 3ee: bipush -1
      // 3ef: ixor
      // 3f0: if_icmpge 3f9
      // 3f3: iload 8
      // 3f5: bipush -1
      // 3f6: iadd
      // 3f7: istore 30
      // 3f9: iload 4
      // 3fb: ldc_w 842218000
      // 3fe: if_icmpeq 402
      // 401: return
      // 402: iload 29
      // 404: istore 39
      // 406: iload 39
      // 408: bipush -1
      // 409: ixor
      // 40a: iload 30
      // 40c: bipush -1
      // 40d: ixor
      // 40e: if_icmplt 46a
      // 411: aload 0
      // 412: getfield ua.Xb [I
      // 415: iload 39
      // 417: aload 0
      // 418: getfield ua.t [I
      // 41b: iload 39
      // 41d: iload 31
      // 41f: dup_x2
      // 420: iastore
      // 421: iastore
      // 422: iload 31
      // 424: iload 32
      // 426: iadd
      // 427: istore 31
      // 429: aload 0
      // 42a: getfield ua.M [I
      // 42d: astore 40
      // 42f: aload 0
      // 430: getfield ua.Tb [I
      // 433: iload 39
      // 435: bipush 0
      // 436: iastore
      // 437: iload 39
      // 439: istore 41
      // 43b: aload 40
      // 43d: iload 41
      // 43f: bipush 0
      // 440: iastore
      // 441: aload 0
      // 442: getfield ua.tb [I
      // 445: iload 39
      // 447: aload 0
      // 448: getfield ua.Wb [I
      // 44b: iload 39
      // 44d: iload 35
      // 44f: dup_x2
      // 450: iastore
      // 451: iastore
      // 452: iload 35
      // 454: iload 36
      // 456: iadd
      // 457: istore 35
      // 459: iinc 39 1
      // 45c: iload 48
      // 45e: ifne 497
      // 461: iload 48
      // 463: ifeq 406
      // 466: goto 46a
      // 469: athrow
      // 46a: iload 20
      // 46c: iload 22
      // 46e: if_icmpeq 497
      // 471: iload 21
      // 473: iload 19
      // 475: ineg
      // 476: iadd
      // 477: ldc_w -1951284024
      // 47a: ishl
      // 47b: iload 20
      // 47d: ineg
      // 47e: iload 22
      // 480: iadd
      // 481: idiv
      // 482: istore 32
      // 484: iload 9
      // 486: ineg
      // 487: iload 13
      // 489: iadd
      // 48a: ldc_w 1234137512
      // 48d: ishl
      // 48e: iload 20
      // 490: ineg
      // 491: iload 22
      // 493: iadd
      // 494: idiv
      // 495: istore 34
      // 497: iload 22
      // 499: iload 20
      // 49b: if_icmpge 4bb
      // 49e: iload 13
      // 4a0: ldc_w -2081746392
      // 4a3: ishl
      // 4a4: istore 33
      // 4a6: iload 20
      // 4a8: istore 30
      // 4aa: iload 22
      // 4ac: istore 29
      // 4ae: iload 21
      // 4b0: ldc_w 1385760008
      // 4b3: ishl
      // 4b4: istore 31
      // 4b6: iload 48
      // 4b8: ifeq 4d3
      // 4bb: iload 19
      // 4bd: ldc_w 743454440
      // 4c0: ishl
      // 4c1: istore 31
      // 4c3: iload 20
      // 4c5: istore 29
      // 4c7: iload 9
      // 4c9: ldc_w 705115528
      // 4cc: ishl
      // 4cd: istore 33
      // 4cf: iload 22
      // 4d1: istore 30
      // 4d3: iload 8
      // 4d5: bipush -1
      // 4d6: iadd
      // 4d7: iload 30
      // 4d9: if_icmpge 4e2
      // 4dc: iload 8
      // 4de: bipush 1
      // 4df: isub
      // 4e0: istore 30
      // 4e2: bipush 0
      // 4e3: iload 29
      // 4e5: if_icmple 4ff
      // 4e8: iload 33
      // 4ea: iload 34
      // 4ec: iload 29
      // 4ee: imul
      // 4ef: isub
      // 4f0: istore 33
      // 4f2: iload 31
      // 4f4: iload 32
      // 4f6: iload 29
      // 4f8: imul
      // 4f9: isub
      // 4fa: istore 31
      // 4fc: bipush 0
      // 4fd: istore 29
      // 4ff: iload 29
      // 501: istore 39
      // 503: iload 30
      // 505: iload 39
      // 507: if_icmplt 585
      // 50a: iload 31
      // 50c: bipush -1
      // 50d: ixor
      // 50e: aload 0
      // 50f: getfield ua.Xb [I
      // 512: iload 39
      // 514: iaload
      // 515: bipush -1
      // 516: ixor
      // 517: iload 48
      // 519: ifne 58d
      // 51c: if_icmple 541
      // 51f: goto 523
      // 522: athrow
      // 523: aload 0
      // 524: getfield ua.Xb [I
      // 527: iload 39
      // 529: iload 31
      // 52b: iastore
      // 52c: aload 0
      // 52d: getfield ua.M [I
      // 530: iload 39
      // 532: iload 33
      // 534: iastore
      // 535: aload 0
      // 536: getfield ua.tb [I
      // 539: iload 39
      // 53b: bipush 0
      // 53c: iastore
      // 53d: goto 541
      // 540: athrow
      // 541: aload 0
      // 542: getfield ua.t [I
      // 545: iload 39
      // 547: iaload
      // 548: bipush -1
      // 549: ixor
      // 54a: iload 31
      // 54c: bipush -1
      // 54d: ixor
      // 54e: if_icmpgt 555
      // 551: goto 56f
      // 554: athrow
      // 555: aload 0
      // 556: getfield ua.t [I
      // 559: iload 39
      // 55b: iload 31
      // 55d: iastore
      // 55e: aload 0
      // 55f: getfield ua.Tb [I
      // 562: iload 39
      // 564: iload 33
      // 566: iastore
      // 567: aload 0
      // 568: getfield ua.Wb [I
      // 56b: iload 39
      // 56d: bipush 0
      // 56e: iastore
      // 56f: iload 31
      // 571: iload 32
      // 573: iadd
      // 574: istore 31
      // 576: iload 33
      // 578: iload 34
      // 57a: iadd
      // 57b: istore 33
      // 57d: iinc 39 1
      // 580: iload 48
      // 582: ifeq 503
      // 585: iload 22
      // 587: bipush -1
      // 588: ixor
      // 589: iload 24
      // 58b: bipush -1
      // 58c: ixor
      // 58d: if_icmplt 5b5
      // 590: iload 13
      // 592: ldc_w -1289051512
      // 595: ishl
      // 596: istore 33
      // 598: iload 14
      // 59a: ldc_w -1525813880
      // 59d: ishl
      // 59e: istore 35
      // 5a0: iload 22
      // 5a2: istore 29
      // 5a4: iload 21
      // 5a6: ldc_w -952809592
      // 5a9: ishl
      // 5aa: istore 31
      // 5ac: iload 24
      // 5ae: istore 30
      // 5b0: iload 48
      // 5b2: ifeq 5d5
      // 5b5: iload 11
      // 5b7: ldc_w -794260248
      // 5ba: ishl
      // 5bb: istore 33
      // 5bd: iload 22
      // 5bf: istore 30
      // 5c1: iload 12
      // 5c3: ldc_w -606906680
      // 5c6: ishl
      // 5c7: istore 35
      // 5c9: iload 23
      // 5cb: ldc_w -1631775192
      // 5ce: ishl
      // 5cf: istore 31
      // 5d1: iload 24
      // 5d3: istore 29
      // 5d5: iload 22
      // 5d7: iload 24
      // 5d9: if_icmpeq 602
      // 5dc: iload 14
      // 5de: ineg
      // 5df: iload 12
      // 5e1: iadd
      // 5e2: ldc_w -665843768
      // 5e5: ishl
      // 5e6: iload 22
      // 5e8: ineg
      // 5e9: iload 24
      // 5eb: iadd
      // 5ec: idiv
      // 5ed: istore 36
      // 5ef: iload 21
      // 5f1: ineg
      // 5f2: iload 23
      // 5f4: iadd
      // 5f5: ldc_w -740697688
      // 5f8: ishl
      // 5f9: iload 24
      // 5fb: iload 22
      // 5fd: ineg
      // 5fe: iadd
      // 5ff: idiv
      // 600: istore 32
      // 602: iload 30
      // 604: iload 8
      // 606: bipush -1
      // 607: iadd
      // 608: if_icmple 611
      // 60b: iload 8
      // 60d: bipush -1
      // 60e: iadd
      // 60f: istore 30
      // 611: iload 29
      // 613: bipush -1
      // 614: ixor
      // 615: bipush -1
      // 616: if_icmpgt 61d
      // 619: goto 634
      // 61c: athrow
      // 61d: iload 31
      // 61f: iload 32
      // 621: iload 29
      // 623: imul
      // 624: isub
      // 625: istore 31
      // 627: iload 35
      // 629: iload 36
      // 62b: iload 29
      // 62d: imul
      // 62e: isub
      // 62f: istore 35
      // 631: bipush 0
      // 632: istore 29
      // 634: iload 29
      // 636: istore 39
      // 638: iload 30
      // 63a: bipush -1
      // 63b: ixor
      // 63c: iload 39
      // 63e: bipush -1
      // 63f: ixor
      // 640: if_icmpgt 6bb
      // 643: iload 31
      // 645: bipush -1
      // 646: ixor
      // 647: aload 0
      // 648: getfield ua.Xb [I
      // 64b: iload 39
      // 64d: iaload
      // 64e: bipush -1
      // 64f: ixor
      // 650: iload 48
      // 652: ifne 6c3
      // 655: if_icmpgt 65f
      // 658: goto 65c
      // 65b: athrow
      // 65c: goto 67a
      // 65f: aload 0
      // 660: getfield ua.Xb [I
      // 663: iload 39
      // 665: iload 31
      // 667: iastore
      // 668: aload 0
      // 669: getfield ua.M [I
      // 66c: iload 39
      // 66e: iload 33
      // 670: iastore
      // 671: aload 0
      // 672: getfield ua.tb [I
      // 675: iload 39
      // 677: iload 35
      // 679: iastore
      // 67a: iload 31
      // 67c: aload 0
      // 67d: getfield ua.t [I
      // 680: iload 39
      // 682: iaload
      // 683: if_icmpgt 68a
      // 686: goto 6a5
      // 689: athrow
      // 68a: aload 0
      // 68b: getfield ua.t [I
      // 68e: iload 39
      // 690: iload 31
      // 692: iastore
      // 693: aload 0
      // 694: getfield ua.Tb [I
      // 697: iload 39
      // 699: iload 33
      // 69b: iastore
      // 69c: aload 0
      // 69d: getfield ua.Wb [I
      // 6a0: iload 39
      // 6a2: iload 35
      // 6a4: iastore
      // 6a5: iload 31
      // 6a7: iload 32
      // 6a9: iadd
      // 6aa: istore 31
      // 6ac: iload 35
      // 6ae: iload 36
      // 6b0: iadd
      // 6b1: istore 35
      // 6b3: iinc 39 1
      // 6b6: iload 48
      // 6b8: ifeq 638
      // 6bb: iload 24
      // 6bd: bipush -1
      // 6be: ixor
      // 6bf: iload 26
      // 6c1: bipush -1
      // 6c2: ixor
      // 6c3: if_icmpne 6c9
      // 6c6: goto 6ef
      // 6c9: iload 25
      // 6cb: iload 23
      // 6cd: ineg
      // 6ce: iadd
      // 6cf: ldc_w -1317172024
      // 6d2: ishl
      // 6d3: iload 26
      // 6d5: iload 24
      // 6d7: ineg
      // 6d8: iadd
      // 6d9: idiv
      // 6da: istore 32
      // 6dc: iload 15
      // 6de: iload 11
      // 6e0: ineg
      // 6e1: iadd
      // 6e2: ldc_w -236389848
      // 6e5: ishl
      // 6e6: iload 24
      // 6e8: ineg
      // 6e9: iload 26
      // 6eb: iadd
      // 6ec: idiv
      // 6ed: istore 34
      // 6ef: iload 24
      // 6f1: bipush -1
      // 6f2: ixor
      // 6f3: iload 26
      // 6f5: bipush -1
      // 6f6: ixor
      // 6f7: if_icmpge 71f
      // 6fa: iload 15
      // 6fc: ldc_w 45836488
      // 6ff: ishl
      // 700: istore 33
      // 702: iload 24
      // 704: istore 30
      // 706: iload 26
      // 708: istore 29
      // 70a: iload 16
      // 70c: ldc_w 1400553864
      // 70f: ishl
      // 710: istore 35
      // 712: iload 25
      // 714: ldc_w 2055980168
      // 717: ishl
      // 718: istore 31
      // 71a: iload 48
      // 71c: ifeq 73f
      // 71f: iload 24
      // 721: istore 29
      // 723: iload 12
      // 725: ldc_w 222482728
      // 728: ishl
      // 729: istore 35
      // 72b: iload 23
      // 72d: ldc_w -523743608
      // 730: ishl
      // 731: istore 31
      // 733: iload 11
      // 735: ldc_w -171150040
      // 738: ishl
      // 739: istore 33
      // 73b: iload 26
      // 73d: istore 30
      // 73f: iload 29
      // 741: iflt 748
      // 744: goto 75f
      // 747: athrow
      // 748: iload 31
      // 74a: iload 29
      // 74c: iload 32
      // 74e: imul
      // 74f: isub
      // 750: istore 31
      // 752: iload 33
      // 754: iload 29
      // 756: iload 34
      // 758: imul
      // 759: isub
      // 75a: istore 33
      // 75c: bipush 0
      // 75d: istore 29
      // 75f: iload 8
      // 761: bipush -1
      // 762: iadd
      // 763: iload 30
      // 765: if_icmpge 76e
      // 768: bipush -1
      // 769: iload 8
      // 76b: iadd
      // 76c: istore 30
      // 76e: iload 29
      // 770: istore 39
      // 772: iload 30
      // 774: iload 39
      // 776: if_icmplt 7f6
      // 779: aload 0
      // 77a: getfield ua.Xb [I
      // 77d: iload 39
      // 77f: iaload
      // 780: bipush -1
      // 781: ixor
      // 782: iload 31
      // 784: bipush -1
      // 785: ixor
      // 786: iload 48
      // 788: ifne 7fa
      // 78b: if_icmpge 7b1
      // 78e: goto 792
      // 791: athrow
      // 792: aload 0
      // 793: getfield ua.Xb [I
      // 796: iload 39
      // 798: iload 31
      // 79a: iastore
      // 79b: aload 0
      // 79c: getfield ua.M [I
      // 79f: iload 39
      // 7a1: iload 33
      // 7a3: iastore
      // 7a4: aload 0
      // 7a5: getfield ua.tb [I
      // 7a8: iload 39
      // 7aa: iload 35
      // 7ac: iastore
      // 7ad: goto 7b1
      // 7b0: athrow
      // 7b1: aload 0
      // 7b2: getfield ua.t [I
      // 7b5: iload 39
      // 7b7: iaload
      // 7b8: bipush -1
      // 7b9: ixor
      // 7ba: iload 31
      // 7bc: bipush -1
      // 7bd: ixor
      // 7be: if_icmpgt 7c5
      // 7c1: goto 7e0
      // 7c4: athrow
      // 7c5: aload 0
      // 7c6: getfield ua.t [I
      // 7c9: iload 39
      // 7cb: iload 31
      // 7cd: iastore
      // 7ce: aload 0
      // 7cf: getfield ua.Tb [I
      // 7d2: iload 39
      // 7d4: iload 33
      // 7d6: iastore
      // 7d7: aload 0
      // 7d8: getfield ua.Wb [I
      // 7db: iload 39
      // 7dd: iload 35
      // 7df: iastore
      // 7e0: iload 31
      // 7e2: iload 32
      // 7e4: iadd
      // 7e5: istore 31
      // 7e7: iload 33
      // 7e9: iload 34
      // 7eb: iadd
      // 7ec: istore 33
      // 7ee: iinc 39 1
      // 7f1: iload 48
      // 7f3: ifeq 772
      // 7f6: iload 27
      // 7f8: iload 7
      // 7fa: imul
      // 7fb: istore 39
      // 7fd: aload 0
      // 7fe: getfield ua.ob [[I
      // 801: iload 1
      // 802: aaload
      // 803: astore 40
      // 805: iload 27
      // 807: istore 41
      // 809: iload 28
      // 80b: bipush -1
      // 80c: ixor
      // 80d: iload 41
      // 80f: bipush -1
      // 810: ixor
      // 811: if_icmpge 96d
      // 814: aload 0
      // 815: getfield ua.Xb [I
      // 818: iload 41
      // 81a: iaload
      // 81b: ldc_w -299165016
      // 81e: ishr
      // 81f: istore 42
      // 821: aload 0
      // 822: getfield ua.t [I
      // 825: iload 41
      // 827: iaload
      // 828: ldc_w -1707577976
      // 82b: ishr
      // 82c: istore 43
      // 82e: iload 48
      // 830: ifne 9c4
      // 833: iload 43
      // 835: iload 42
      // 837: ineg
      // 838: iadd
      // 839: bipush -1
      // 83a: ixor
      // 83b: bipush -1
      // 83c: if_icmpge 847
      // 83f: goto 843
      // 842: athrow
      // 843: goto 853
      // 846: athrow
      // 847: iload 39
      // 849: iload 7
      // 84b: iadd
      // 84c: istore 39
      // 84e: iload 48
      // 850: ifeq 965
      // 853: aload 0
      // 854: getfield ua.M [I
      // 857: iload 41
      // 859: iaload
      // 85a: ldc_w -122904183
      // 85d: ishl
      // 85e: istore 44
      // 860: aload 0
      // 861: getfield ua.Tb [I
      // 864: iload 41
      // 866: iaload
      // 867: ldc_w 1728443241
      // 86a: ishl
      // 86b: iload 44
      // 86d: isub
      // 86e: iload 43
      // 870: iload 42
      // 872: isub
      // 873: idiv
      // 874: istore 45
      // 876: aload 0
      // 877: getfield ua.tb [I
      // 87a: iload 41
      // 87c: iaload
      // 87d: ldc_w 1699710857
      // 880: ishl
      // 881: istore 46
      // 883: iload 46
      // 885: ineg
      // 886: aload 0
      // 887: getfield ua.Wb [I
      // 88a: iload 41
      // 88c: iaload
      // 88d: ldc_w -470306615
      // 890: ishl
      // 891: iadd
      // 892: iload 42
      // 894: ineg
      // 895: iload 43
      // 897: iadd
      // 898: idiv
      // 899: istore 47
      // 89b: aload 0
      // 89c: getfield ua.lb I
      // 89f: bipush -1
      // 8a0: ixor
      // 8a1: iload 43
      // 8a3: bipush -1
      // 8a4: ixor
      // 8a5: if_icmpgt 8ac
      // 8a8: goto 8b2
      // 8ab: athrow
      // 8ac: aload 0
      // 8ad: getfield ua.lb I
      // 8b0: istore 43
      // 8b2: iload 42
      // 8b4: aload 0
      // 8b5: getfield ua.hb I
      // 8b8: if_icmpge 8e1
      // 8bb: iload 46
      // 8bd: iload 42
      // 8bf: ineg
      // 8c0: aload 0
      // 8c1: getfield ua.hb I
      // 8c4: iadd
      // 8c5: iload 47
      // 8c7: imul
      // 8c8: iadd
      // 8c9: istore 46
      // 8cb: iload 44
      // 8cd: aload 0
      // 8ce: getfield ua.hb I
      // 8d1: iload 42
      // 8d3: ineg
      // 8d4: iadd
      // 8d5: iload 45
      // 8d7: imul
      // 8d8: iadd
      // 8d9: istore 44
      // 8db: aload 0
      // 8dc: getfield ua.hb I
      // 8df: istore 42
      // 8e1: bipush 1
      // 8e2: aload 0
      // 8e3: getfield ua.i Z
      // 8e6: ifne 8ee
      // 8e9: bipush 1
      // 8ea: goto 8ef
      // 8ed: athrow
      // 8ee: bipush 0
      // 8ef: if_icmpeq 900
      // 8f2: iload 41
      // 8f4: bipush 1
      // 8f5: iand
      // 8f6: bipush -1
      // 8f7: ixor
      // 8f8: bipush -1
      // 8f9: if_icmpne 95e
      // 8fc: goto 900
      // 8ff: athrow
      // 900: bipush 1
      // 901: aload 0
      // 902: getfield ua.Qb [Z
      // 905: iload 1
      // 906: baload
      // 907: if_icmpne 939
      // 90a: goto 90e
      // 90d: athrow
      // 90e: aload 0
      // 90f: iload 47
      // 911: iload 43
      // 913: ineg
      // 914: iload 42
      // 916: iadd
      // 917: iload 44
      // 919: aload 40
      // 91b: aload 0
      // 91c: getfield ua.rb [I
      // 91f: iload 46
      // 921: iload 42
      // 923: iload 39
      // 925: iadd
      // 926: iload 37
      // 928: iload 45
      // 92a: bipush 0
      // 92b: bipush 102
      // 92d: invokespecial ua.a (III[I[IIIIIIB)V
      // 930: iload 48
      // 932: ifeq 95e
      // 935: goto 939
      // 938: athrow
      // 939: aload 0
      // 93a: iload 43
      // 93c: ineg
      // 93d: iload 42
      // 93f: iadd
      // 940: iload 37
      // 942: bipush 0
      // 943: aload 0
      // 944: getfield ua.rb [I
      // 947: iload 45
      // 949: iload 46
      // 94b: iload 44
      // 94d: aload 40
      // 94f: iload 47
      // 951: iload 42
      // 953: iload 39
      // 955: iadd
      // 956: bipush 1
      // 957: invokespecial ua.a (III[IIII[IIIZ)V
      // 95a: goto 95e
      // 95d: athrow
      // 95e: iload 39
      // 960: iload 7
      // 962: iadd
      // 963: istore 39
      // 965: iinc 41 1
      // 968: iload 48
      // 96a: ifeq 809
      // 96d: goto 9c4
      // 970: astore 7
      // 972: aload 7
      // 974: new java/lang/StringBuilder
      // 977: dup
      // 978: invokespecial java/lang/StringBuilder.<init> ()V
      // 97b: getstatic ua.Yb [Ljava/lang/String;
      // 97e: bipush 79
      // 980: aaload
      // 981: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 984: iload 1
      // 985: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 988: bipush 44
      // 98a: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 98d: iload 2
      // 98e: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 991: bipush 44
      // 993: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 996: iload 3
      // 997: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 99a: bipush 44
      // 99c: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 99f: iload 4
      // 9a1: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 9a4: bipush 44
      // 9a6: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 9a9: iload 5
      // 9ab: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 9ae: bipush 44
      // 9b0: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 9b3: iload 6
      // 9b5: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 9b8: bipush 41
      // 9ba: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 9bd: invokevirtual java/lang/StringBuilder.toString ()Ljava/lang/String;
      // 9c0: invokestatic i.a (Ljava/lang/Throwable;Ljava/lang/String;)Lla;
      // 9c3: athrow
      // 9c4: getstatic e.Ab I
      // 9c7: ifeq 9dc
      // 9ca: iload 48
      // 9cc: ifeq 9d8
      // 9cf: goto 9d3
      // 9d2: athrow
      // 9d3: bipush 0
      // 9d4: goto 9d9
      // 9d7: athrow
      // 9d8: bipush 1
      // 9d9: putstatic client.vh Z
      // 9dc: return
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
                     try {
                        this.ob[var1][var8++] = this.rb[this.u * var9 + var10];
                        var10++;
                        if (var11) {
                           break;
                        }

                        if (!var11) {
                           continue;
                        }
                     } catch (RuntimeException var12) {
                        throw var12;
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
      } catch (RuntimeException var13) {
         throw i.a(var13, Yb[53] + var1 + ',' + var2 + ',' + var3 + ',' + var4 + ',' + var5 + ',' + var6 + ')');
      }
   }

   private final void a(
      int param1,
      int param2,
      int param3,
      int param4,
      int param5,
      int[] param6,
      int param7,
      int[] param8,
      int param9,
      int param10,
      boolean param11,
      int param12,
      int param13,
      int param14,
      int param15
   ) {
      // $VF: Couldn't be decompiled
      // Please report this to the Vineflower issue tracker, at https://github.com/Vineflower/vineflower/issues with a copy of the class file (if you have the rights to distribute it!)
      // java.lang.RuntimeException: parsing failure!
      //   at org.jetbrains.java.decompiler.modules.decompiler.decompose.DomHelper.parseGraph(DomHelper.java:211)
      //   at org.jetbrains.java.decompiler.main.rels.MethodProcessor.codeToJava(MethodProcessor.java:166)
      //
      // Bytecode:
      // 000: getstatic client.vh Z
      // 003: istore 26
      // 005: getstatic ua.vb I
      // 008: bipush 1
      // 009: iadd
      // 00a: putstatic ua.vb I
      // 00d: iload 14
      // 00f: ldc_w 16736117
      // 012: iand
      // 013: ldc_w -846694704
      // 016: ishr
      // 017: istore 16
      // 019: sipush 255
      // 01c: iload 14
      // 01e: ldc_w -1400710808
      // 021: ishr
      // 022: iand
      // 023: istore 17
      // 025: sipush 255
      // 028: iload 14
      // 02a: iand
      // 02b: istore 18
      // 02d: iload 4
      // 02f: istore 19
      // 031: iload 11
      // 033: ifeq 042
      // 036: aload 0
      // 037: aconst_null
      // 038: checkcast [I
      // 03b: putfield ua.Tb [I
      // 03e: goto 042
      // 041: athrow
      // 042: iload 13
      // 044: ineg
      // 045: istore 20
      // 047: iload 20
      // 049: bipush -1
      // 04a: ixor
      // 04b: bipush -1
      // 04c: if_icmple 13d
      // 04f: iload 1
      // 050: ldc_w -324011088
      // 053: ishr
      // 054: iload 10
      // 056: imul
      // 057: istore 21
      // 059: iload 26
      // 05b: ifne 14e
      // 05e: iload 3
      // 05f: ineg
      // 060: istore 22
      // 062: bipush -1
      // 063: iload 22
      // 065: bipush -1
      // 066: ixor
      // 067: if_icmpge 121
      // 06a: aload 6
      // 06c: iload 21
      // 06e: iload 4
      // 070: ldc_w 1856265008
      // 073: ishr
      // 074: iadd
      // 075: iaload
      // 076: istore 9
      // 078: iload 9
      // 07a: bipush -1
      // 07b: ixor
      // 07c: bipush -1
      // 07d: iload 26
      // 07f: ifne 135
      // 082: if_icmpeq 10c
      // 085: goto 089
      // 088: athrow
      // 089: iload 9
      // 08b: ldc_w -1835533520
      // 08e: ishr
      // 08f: sipush 255
      // 092: iand
      // 093: istore 23
      // 095: ldc_w 65376
      // 098: iload 9
      // 09a: iand
      // 09b: ldc_w 1180773672
      // 09e: ishr
      // 09f: istore 24
      // 0a1: sipush 255
      // 0a4: iload 9
      // 0a6: iand
      // 0a7: istore 25
      // 0a9: iload 24
      // 0ab: bipush -1
      // 0ac: ixor
      // 0ad: iload 23
      // 0af: bipush -1
      // 0b0: ixor
      // 0b1: if_icmpne 0c3
      // 0b4: iload 24
      // 0b6: bipush -1
      // 0b7: ixor
      // 0b8: iload 25
      // 0ba: bipush -1
      // 0bb: ixor
      // 0bc: if_icmpeq 0d6
      // 0bf: goto 0c3
      // 0c2: athrow
      // 0c3: aload 8
      // 0c5: iload 7
      // 0c7: iinc 7 1
      // 0ca: iload 9
      // 0cc: iastore
      // 0cd: iload 26
      // 0cf: ifeq 107
      // 0d2: goto 0d6
      // 0d5: athrow
      // 0d6: aload 8
      // 0d8: iload 7
      // 0da: iinc 7 1
      // 0dd: iload 23
      // 0df: iload 16
      // 0e1: imul
      // 0e2: ldc_w 1799197256
      // 0e5: ishr
      // 0e6: ldc_w -273346448
      // 0e9: ishl
      // 0ea: iload 17
      // 0ec: iload 24
      // 0ee: imul
      // 0ef: ldc_w -1796300920
      // 0f2: ishr
      // 0f3: ldc_w 1450486152
      // 0f6: ishl
      // 0f7: iload 25
      // 0f9: iload 18
      // 0fb: imul
      // 0fc: ldc_w 835792776
      // 0ff: ishr
      // 100: iadd
      // 101: iadd
      // 102: iastore
      // 103: goto 107
      // 106: athrow
      // 107: iload 26
      // 109: ifeq 113
      // 10c: iinc 7 1
      // 10f: goto 113
      // 112: athrow
      // 113: iload 4
      // 115: iload 2
      // 116: iadd
      // 117: istore 4
      // 119: iinc 22 1
      // 11c: iload 26
      // 11e: ifeq 062
      // 121: iload 1
      // 122: iload 12
      // 124: iadd
      // 125: istore 1
      // 126: iload 7
      // 128: iload 5
      // 12a: iadd
      // 12b: istore 7
      // 12d: iload 19
      // 12f: istore 4
      // 131: iload 20
      // 133: iload 15
      // 135: iadd
      // 136: istore 20
      // 138: iload 26
      // 13a: ifeq 047
      // 13d: goto 14e
      // 140: astore 19
      // 142: getstatic java/lang/System.out Ljava/io/PrintStream;
      // 145: getstatic ua.Yb [Ljava/lang/String;
      // 148: bipush 47
      // 14a: aaload
      // 14b: invokevirtual java/io/PrintStream.println (Ljava/lang/String;)V
      // 14e: goto 221
      // 151: astore 16
      // 153: aload 16
      // 155: new java/lang/StringBuilder
      // 158: dup
      // 159: invokespecial java/lang/StringBuilder.<init> ()V
      // 15c: getstatic ua.Yb [Ljava/lang/String;
      // 15f: bipush 52
      // 161: aaload
      // 162: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 165: iload 1
      // 166: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 169: bipush 44
      // 16b: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 16e: iload 2
      // 16f: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 172: bipush 44
      // 174: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 177: iload 3
      // 178: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 17b: bipush 44
      // 17d: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 180: iload 4
      // 182: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 185: bipush 44
      // 187: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 18a: iload 5
      // 18c: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 18f: bipush 44
      // 191: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 194: aload 6
      // 196: ifnull 1a2
      // 199: getstatic ua.Yb [Ljava/lang/String;
      // 19c: bipush 1
      // 19d: aaload
      // 19e: goto 1a7
      // 1a1: athrow
      // 1a2: getstatic ua.Yb [Ljava/lang/String;
      // 1a5: bipush 0
      // 1a6: aaload
      // 1a7: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 1aa: bipush 44
      // 1ac: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 1af: iload 7
      // 1b1: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 1b4: bipush 44
      // 1b6: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 1b9: aload 8
      // 1bb: ifnull 1c7
      // 1be: getstatic ua.Yb [Ljava/lang/String;
      // 1c1: bipush 1
      // 1c2: aaload
      // 1c3: goto 1cc
      // 1c6: athrow
      // 1c7: getstatic ua.Yb [Ljava/lang/String;
      // 1ca: bipush 0
      // 1cb: aaload
      // 1cc: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 1cf: bipush 44
      // 1d1: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 1d4: iload 9
      // 1d6: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 1d9: bipush 44
      // 1db: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 1de: iload 10
      // 1e0: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 1e3: bipush 44
      // 1e5: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 1e8: iload 11
      // 1ea: invokevirtual java/lang/StringBuilder.append (Z)Ljava/lang/StringBuilder;
      // 1ed: bipush 44
      // 1ef: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 1f2: iload 12
      // 1f4: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 1f7: bipush 44
      // 1f9: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 1fc: iload 13
      // 1fe: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 201: bipush 44
      // 203: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 206: iload 14
      // 208: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 20b: bipush 44
      // 20d: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 210: iload 15
      // 212: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 215: bipush 41
      // 217: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 21a: invokevirtual java/lang/StringBuilder.toString ()Ljava/lang/String;
      // 21d: invokestatic i.a (Ljava/lang/Throwable;Ljava/lang/String;)Lla;
      // 220: athrow
      // 221: return
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
      } catch (RuntimeException var10) {
         RuntimeException var8 = var10;

         RuntimeException var10000;
         StringBuilder var10001;
         try {
            var10000 = var8;
            var10001 = new StringBuilder().append(Yb[9]).append(var1).append(',').append(var2).append(',').append(var3).append(',').append(var4).append(',');
            if (var5 != null) {
               throw i.a(var8, var10001.append(Yb[1]).append(',').append(var6).append(',').append(var7).append(')').toString());
            }
         } catch (RuntimeException var9) {
            throw var9;
         }

         throw i.a(var10000, var10001.append(Yb[0]).append(',').append(var6).append(',').append(var7).append(')').toString());
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
         int var29 = 256 - var14;

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

               int var30;
               int var32;
               label89: {
                  while (-1 < ~var20) {
                     var11 = var10[var19 + (var12 >> -483039408)];
                     var12 += var7;

                     label85: {
                        label84: {
                           try {
                              var30 = ~var11;
                              var32 = -1;
                              if (var22) {
                                 break label89;
                              }

                              if (var30 != -1) {
                                 break label84;
                              }
                           } catch (Exception var26) {
                              throw var26;
                           }

                           try {
                              var9++;
                              if (!var22) {
                                 break label85;
                              }
                           } catch (Exception var25) {
                              throw var25;
                           }
                        }

                        int var21 = var15[var9];
                        var15[var9++] = ib.a(ib.a(65280, var21) * var29 + ib.a(65280, var11) * var14, 16711680)
                              + ib.a(ib.a(var11, 16711935) * var14 + var29 * ib.a(16711935, var21), -16711936)
                           >> -130221816;
                     }

                     var20++;
                     if (var22) {
                        break;
                     }
                  }

                  var9 += var13;
                  var12 = var17;
                  var30 = var18;
                  var32 = var1;
               }

               var18 = var30 + var32;
               if (var22) {
                  break;
               }
            }
         } catch (Exception var27) {
            System.out.println(Yb[45]);
         }
      } catch (RuntimeException var28) {
         RuntimeException var16 = var28;

         RuntimeException var10000;
         StringBuilder var10001;
         String var10002;
         label62: {
            try {
               var10000 = var16;
               var10001 = new StringBuilder()
                  .append(Yb[46])
                  .append(var1)
                  .append(',')
                  .append(var2)
                  .append(',')
                  .append(var3)
                  .append(',')
                  .append((int)var4)
                  .append(',')
                  .append(var5)
                  .append(',')
                  .append(var6)
                  .append(',')
                  .append(var7)
                  .append(',')
                  .append(var8)
                  .append(',')
                  .append(var9)
                  .append(',');
               if (var10 != null) {
                  var10002 = Yb[1];
                  break label62;
               }
            } catch (Exception var24) {
               throw var24;
            }

            var10002 = Yb[0];
         }

         try {
            var10001 = var10001.append(var10002)
               .append(',')
               .append(var11)
               .append(',')
               .append(var12)
               .append(',')
               .append(var13)
               .append(',')
               .append(var14)
               .append(',');
            if (var15 != null) {
               throw i.a(var10000, var10001.append(Yb[1]).append(')').toString());
            }
         } catch (Exception var23) {
            throw var23;
         }

         throw i.a(var10000, var10001.append(Yb[0]).append(')').toString());
      }
   }

   private final void a(
      int param1,
      int[] param2,
      int param3,
      int param4,
      int param5,
      int param6,
      int param7,
      int[] param8,
      int param9,
      int param10,
      int param11,
      int param12,
      int param13,
      int param14,
      int param15,
      int param16
   ) {
      // $VF: Couldn't be decompiled
      // Please report this to the Vineflower issue tracker, at https://github.com/Vineflower/vineflower/issues with a copy of the class file (if you have the rights to distribute it!)
      // java.lang.RuntimeException: parsing failure!
      //   at org.jetbrains.java.decompiler.modules.decompiler.decompose.DomHelper.parseGraph(DomHelper.java:211)
      //   at org.jetbrains.java.decompiler.main.rels.MethodProcessor.codeToJava(MethodProcessor.java:166)
      //
      // Bytecode:
      // 000: getstatic client.vh Z
      // 003: istore 29
      // 005: getstatic ua.L I
      // 008: bipush 1
      // 009: iadd
      // 00a: putstatic ua.L I
      // 00d: ldc_w 16717586
      // 010: iload 16
      // 012: iand
      // 013: ldc_w -1742839120
      // 016: ishr
      // 017: istore 20
      // 019: iload 16
      // 01b: ldc_w 65518
      // 01e: iand
      // 01f: ldc_w -734006424
      // 022: ishr
      // 023: istore 21
      // 025: sipush 255
      // 028: iload 16
      // 02a: iand
      // 02b: istore 22
      // 02d: iload 6
      // 02f: istore 23
      // 031: iload 9
      // 033: ineg
      // 034: istore 24
      // 036: bipush -1
      // 037: iload 24
      // 039: bipush -1
      // 03a: ixor
      // 03b: if_icmpge 199
      // 03e: iload 5
      // 040: ldc_w 1187405840
      // 043: ishr
      // 044: iload 14
      // 046: imul
      // 047: istore 25
      // 049: iload 13
      // 04b: ldc_w -1195642352
      // 04e: ishr
      // 04f: istore 26
      // 051: iload 7
      // 053: istore 27
      // 055: iload 26
      // 057: aload 0
      // 058: getfield ua.hb I
      // 05b: iload 29
      // 05d: ifne 1ad
      // 060: if_icmpge 088
      // 063: goto 067
      // 066: athrow
      // 067: aload 0
      // 068: getfield ua.hb I
      // 06b: iload 26
      // 06d: ineg
      // 06e: iadd
      // 06f: istore 28
      // 071: aload 0
      // 072: getfield ua.hb I
      // 075: istore 26
      // 077: iload 27
      // 079: iload 28
      // 07b: isub
      // 07c: istore 27
      // 07e: iload 6
      // 080: iload 12
      // 082: iload 28
      // 084: imul
      // 085: iadd
      // 086: istore 6
      // 088: iload 15
      // 08a: ineg
      // 08b: bipush 1
      // 08c: iadd
      // 08d: istore 15
      // 08f: aload 0
      // 090: getfield ua.lb I
      // 093: bipush -1
      // 094: ixor
      // 095: iload 26
      // 097: iload 27
      // 099: ineg
      // 09a: isub
      // 09b: bipush -1
      // 09c: ixor
      // 09d: if_icmpge 0a4
      // 0a0: goto 0b8
      // 0a3: athrow
      // 0a4: iload 27
      // 0a6: iload 26
      // 0a8: iadd
      // 0a9: aload 0
      // 0aa: getfield ua.lb I
      // 0ad: ineg
      // 0ae: iadd
      // 0af: istore 28
      // 0b1: iload 27
      // 0b3: iload 28
      // 0b5: isub
      // 0b6: istore 27
      // 0b8: bipush 0
      // 0b9: iload 15
      // 0bb: if_icmpne 0c2
      // 0be: goto 177
      // 0c1: athrow
      // 0c2: iload 26
      // 0c4: istore 28
      // 0c6: iload 28
      // 0c8: iload 27
      // 0ca: iload 26
      // 0cc: iadd
      // 0cd: if_icmpge 177
      // 0d0: aload 2
      // 0d1: iload 25
      // 0d3: iload 6
      // 0d5: ldc_w -1996606992
      // 0d8: ishr
      // 0d9: iadd
      // 0da: iaload
      // 0db: istore 4
      // 0dd: iload 4
      // 0df: iload 29
      // 0e1: ifne 18f
      // 0e4: ifne 0ee
      // 0e7: goto 0eb
      // 0ea: athrow
      // 0eb: goto 168
      // 0ee: iload 4
      // 0f0: sipush 255
      // 0f3: iand
      // 0f4: istore 19
      // 0f6: iload 4
      // 0f8: ldc_w 16769815
      // 0fb: iand
      // 0fc: ldc_w 1817608464
      // 0ff: ishr
      // 100: istore 17
      // 102: sipush 255
      // 105: iload 4
      // 107: ldc_w -7992056
      // 10a: ishr
      // 10b: iand
      // 10c: istore 18
      // 10e: iload 18
      // 110: bipush -1
      // 111: ixor
      // 112: iload 17
      // 114: bipush -1
      // 115: ixor
      // 116: if_icmpne 15a
      // 119: iload 19
      // 11b: iload 18
      // 11d: if_icmpne 15a
      // 120: goto 124
      // 123: athrow
      // 124: aload 8
      // 126: iload 11
      // 128: iload 28
      // 12a: iadd
      // 12b: iload 19
      // 12d: iload 22
      // 12f: imul
      // 130: ldc_w -1923074072
      // 133: ishr
      // 134: iload 21
      // 136: iload 18
      // 138: imul
      // 139: ldc_w -1585910200
      // 13c: ishr
      // 13d: ldc_w -1586887384
      // 140: ishl
      // 141: iadd
      // 142: iload 20
      // 144: iload 17
      // 146: imul
      // 147: ldc_w 2107059944
      // 14a: ishr
      // 14b: ldc_w -1441257008
      // 14e: ishl
      // 14f: iadd
      // 150: iastore
      // 151: iload 29
      // 153: ifeq 168
      // 156: goto 15a
      // 159: athrow
      // 15a: aload 8
      // 15c: iload 28
      // 15e: iload 11
      // 160: iadd
      // 161: iload 4
      // 163: iastore
      // 164: goto 168
      // 167: athrow
      // 168: iload 6
      // 16a: iload 12
      // 16c: iadd
      // 16d: istore 6
      // 16f: iinc 28 1
      // 172: iload 29
      // 174: ifeq 0c6
      // 177: iload 5
      // 179: iload 3
      // 17a: iadd
      // 17b: istore 5
      // 17d: iload 23
      // 17f: istore 6
      // 181: iload 13
      // 183: iload 10
      // 185: iadd
      // 186: istore 13
      // 188: iload 11
      // 18a: aload 0
      // 18b: getfield ua.u I
      // 18e: iadd
      // 18f: istore 11
      // 191: iinc 24 1
      // 194: iload 29
      // 196: ifeq 036
      // 199: goto 1aa
      // 19c: astore 23
      // 19e: getstatic java/lang/System.out Ljava/io/PrintStream;
      // 1a1: getstatic ua.Yb [Ljava/lang/String;
      // 1a4: bipush 65
      // 1a6: aaload
      // 1a7: invokevirtual java/io/PrintStream.println (Ljava/lang/String;)V
      // 1aa: iload 1
      // 1ab: bipush 20
      // 1ad: if_icmpge 1bc
      // 1b0: aload 0
      // 1b1: aconst_null
      // 1b2: checkcast [I
      // 1b5: putfield ua.t [I
      // 1b8: goto 1bc
      // 1bb: athrow
      // 1bc: goto 299
      // 1bf: astore 17
      // 1c1: aload 17
      // 1c3: new java/lang/StringBuilder
      // 1c6: dup
      // 1c7: invokespecial java/lang/StringBuilder.<init> ()V
      // 1ca: getstatic ua.Yb [Ljava/lang/String;
      // 1cd: bipush 89
      // 1cf: aaload
      // 1d0: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 1d3: iload 1
      // 1d4: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 1d7: bipush 44
      // 1d9: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 1dc: aload 2
      // 1dd: ifnull 1e9
      // 1e0: getstatic ua.Yb [Ljava/lang/String;
      // 1e3: bipush 1
      // 1e4: aaload
      // 1e5: goto 1ee
      // 1e8: athrow
      // 1e9: getstatic ua.Yb [Ljava/lang/String;
      // 1ec: bipush 0
      // 1ed: aaload
      // 1ee: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 1f1: bipush 44
      // 1f3: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 1f6: iload 3
      // 1f7: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 1fa: bipush 44
      // 1fc: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 1ff: iload 4
      // 201: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 204: bipush 44
      // 206: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 209: iload 5
      // 20b: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 20e: bipush 44
      // 210: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 213: iload 6
      // 215: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 218: bipush 44
      // 21a: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 21d: iload 7
      // 21f: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 222: bipush 44
      // 224: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 227: aload 8
      // 229: ifnull 235
      // 22c: getstatic ua.Yb [Ljava/lang/String;
      // 22f: bipush 1
      // 230: aaload
      // 231: goto 23a
      // 234: athrow
      // 235: getstatic ua.Yb [Ljava/lang/String;
      // 238: bipush 0
      // 239: aaload
      // 23a: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 23d: bipush 44
      // 23f: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 242: iload 9
      // 244: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 247: bipush 44
      // 249: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 24c: iload 10
      // 24e: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 251: bipush 44
      // 253: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 256: iload 11
      // 258: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 25b: bipush 44
      // 25d: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 260: iload 12
      // 262: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 265: bipush 44
      // 267: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 26a: iload 13
      // 26c: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 26f: bipush 44
      // 271: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 274: iload 14
      // 276: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 279: bipush 44
      // 27b: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 27e: iload 15
      // 280: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 283: bipush 44
      // 285: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 288: iload 16
      // 28a: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 28d: bipush 41
      // 28f: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 292: invokevirtual java/lang/StringBuilder.toString ()Ljava/lang/String;
      // 295: invokestatic i.a (Ljava/lang/Throwable;Ljava/lang/String;)Lla;
      // 298: athrow
      // 299: return
   }

   final int a(int var1, int var2) {
      try {
         try {
            if (var1 != 508305352) {
               this.G = (int[])null;
            }
         } catch (RuntimeException var4) {
            throw var4;
         }

         try {
            Pb++;
            if (var2 == 0) {
               return 12;
            }
         } catch (RuntimeException var9) {
            throw var9;
         }

         try {
            if (var2 == 1) {
               return 14;
            }
         } catch (RuntimeException var8) {
            throw var8;
         }

         if (var2 == 2) {
            return 14;
         } else if (-4 == ~var2) {
            return 15;
         } else {
            try {
               if (-5 == ~var2) {
                  return 15;
               }
            } catch (RuntimeException var7) {
               throw var7;
            }

            try {
               if (~var2 == -6) {
                  return 19;
               }
            } catch (RuntimeException var6) {
               throw var6;
            }

            if (6 == var2) {
               return 24;
            } else {
               try {
                  if (7 == var2) {
                     return 29;
                  }
               } catch (RuntimeException var5) {
                  throw var5;
               }

               return this.c(60, var2);
            }
         }
      } catch (RuntimeException var10) {
         throw i.a(var10, Yb[74] + var1 + 44 + var2 + 41);
      }
   }

   private final void a(
      byte[] param1,
      int param2,
      int param3,
      int param4,
      int[] param5,
      int param6,
      int param7,
      int param8,
      int param9,
      int param10,
      int[] param11,
      int param12,
      int param13,
      int param14,
      int param15,
      int param16,
      int param17
   ) {
      // $VF: Couldn't be decompiled
      // Please report this to the Vineflower issue tracker, at https://github.com/Vineflower/vineflower/issues with a copy of the class file (if you have the rights to distribute it!)
      // java.lang.RuntimeException: parsing failure!
      //   at org.jetbrains.java.decompiler.modules.decompiler.decompose.DomHelper.parseGraph(DomHelper.java:211)
      //   at org.jetbrains.java.decompiler.main.rels.MethodProcessor.codeToJava(MethodProcessor.java:166)
      //
      // Bytecode:
      // 000: getstatic client.vh Z
      // 003: istore 31
      // 005: getstatic ua.m I
      // 008: bipush 1
      // 009: iadd
      // 00a: putstatic ua.m I
      // 00d: ldc_w 16721450
      // 010: iload 2
      // 011: iand
      // 012: ldc_w 1137456112
      // 015: ishr
      // 016: istore 21
      // 018: ldc_w 65437
      // 01b: iload 2
      // 01c: iand
      // 01d: ldc_w -1550693080
      // 020: ishr
      // 021: istore 22
      // 023: bipush 82
      // 025: iload 13
      // 027: bipush -52
      // 029: isub
      // 02a: bipush 32
      // 02c: idiv
      // 02d: irem
      // 02e: istore 23
      // 030: iload 2
      // 031: sipush 255
      // 034: iand
      // 035: istore 24
      // 037: iload 14
      // 039: istore 25
      // 03b: iload 17
      // 03d: ineg
      // 03e: istore 26
      // 040: iload 26
      // 042: ifge 1b3
      // 045: iload 7
      // 047: ldc_w 160375440
      // 04a: ishr
      // 04b: iload 6
      // 04d: imul
      // 04e: istore 27
      // 050: iload 9
      // 052: ldc_w 698170864
      // 055: ishr
      // 056: istore 28
      // 058: iload 16
      // 05a: istore 29
      // 05c: iload 31
      // 05e: ifne 1c4
      // 061: aload 0
      // 062: getfield ua.hb I
      // 065: bipush -1
      // 066: ixor
      // 067: iload 28
      // 069: bipush -1
      // 06a: ixor
      // 06b: if_icmplt 076
      // 06e: goto 072
      // 071: athrow
      // 072: goto 097
      // 075: athrow
      // 076: aload 0
      // 077: getfield ua.hb I
      // 07a: iload 28
      // 07c: ineg
      // 07d: iadd
      // 07e: istore 30
      // 080: aload 0
      // 081: getfield ua.hb I
      // 084: istore 28
      // 086: iload 29
      // 088: iload 30
      // 08a: isub
      // 08b: istore 29
      // 08d: iload 14
      // 08f: iload 30
      // 091: iload 10
      // 093: imul
      // 094: iadd
      // 095: istore 14
      // 097: aload 0
      // 098: getfield ua.lb I
      // 09b: bipush -1
      // 09c: ixor
      // 09d: iload 29
      // 09f: iload 28
      // 0a1: iadd
      // 0a2: bipush -1
      // 0a3: ixor
      // 0a4: if_icmpge 0ab
      // 0a7: goto 0be
      // 0aa: athrow
      // 0ab: iload 28
      // 0ad: iload 29
      // 0af: aload 0
      // 0b0: getfield ua.lb I
      // 0b3: isub
      // 0b4: iadd
      // 0b5: istore 30
      // 0b7: iload 29
      // 0b9: iload 30
      // 0bb: isub
      // 0bc: istore 29
      // 0be: bipush 1
      // 0bf: iload 12
      // 0c1: ineg
      // 0c2: iadd
      // 0c3: istore 12
      // 0c5: iload 7
      // 0c7: iload 8
      // 0c9: iadd
      // 0ca: istore 7
      // 0cc: bipush -1
      // 0cd: iload 12
      // 0cf: bipush -1
      // 0d0: ixor
      // 0d1: if_icmpeq 197
      // 0d4: iload 28
      // 0d6: istore 30
      // 0d8: iload 29
      // 0da: iload 28
      // 0dc: iadd
      // 0dd: bipush -1
      // 0de: ixor
      // 0df: iload 30
      // 0e1: bipush -1
      // 0e2: ixor
      // 0e3: if_icmpge 197
      // 0e6: aload 1
      // 0e7: iload 14
      // 0e9: ldc_w -1390664752
      // 0ec: ishr
      // 0ed: iload 27
      // 0ef: ineg
      // 0f0: isub
      // 0f1: baload
      // 0f2: sipush 255
      // 0f5: iand
      // 0f6: istore 3
      // 0f7: bipush -1
      // 0f8: iload 3
      // 0f9: bipush -1
      // 0fa: ixor
      // 0fb: iload 31
      // 0fd: ifne 1a4
      // 100: if_icmpeq 188
      // 103: goto 107
      // 106: athrow
      // 107: aload 5
      // 109: iload 3
      // 10a: iaload
      // 10b: istore 3
      // 10c: iload 3
      // 10d: sipush 255
      // 110: iand
      // 111: istore 20
      // 113: iload 3
      // 114: ldc_w -692748912
      // 117: ishr
      // 118: sipush 255
      // 11b: iand
      // 11c: istore 18
      // 11e: iload 3
      // 11f: ldc_w 1006963688
      // 122: ishr
      // 123: sipush 255
      // 126: iand
      // 127: istore 19
      // 129: iload 19
      // 12b: bipush -1
      // 12c: ixor
      // 12d: iload 18
      // 12f: bipush -1
      // 130: ixor
      // 131: if_icmpne 143
      // 134: iload 19
      // 136: bipush -1
      // 137: ixor
      // 138: iload 20
      // 13a: bipush -1
      // 13b: ixor
      // 13c: if_icmpeq 155
      // 13f: goto 143
      // 142: athrow
      // 143: aload 11
      // 145: iload 15
      // 147: iload 30
      // 149: iadd
      // 14a: iload 3
      // 14b: iastore
      // 14c: iload 31
      // 14e: ifeq 188
      // 151: goto 155
      // 154: athrow
      // 155: aload 11
      // 157: iload 30
      // 159: iload 15
      // 15b: ineg
      // 15c: isub
      // 15d: iload 22
      // 15f: iload 19
      // 161: imul
      // 162: ldc_w 436099816
      // 165: ishr
      // 166: ldc_w -1418519128
      // 169: ishl
      // 16a: iload 21
      // 16c: iload 18
      // 16e: imul
      // 16f: ldc_w -2076941880
      // 172: ishr
      // 173: ldc_w 547107024
      // 176: ishl
      // 177: iadd
      // 178: iload 24
      // 17a: iload 20
      // 17c: imul
      // 17d: ldc_w -2023878008
      // 180: ishr
      // 181: ineg
      // 182: isub
      // 183: iastore
      // 184: goto 188
      // 187: athrow
      // 188: iload 14
      // 18a: iload 10
      // 18c: iadd
      // 18d: istore 14
      // 18f: iinc 30 1
      // 192: iload 31
      // 194: ifeq 0d8
      // 197: iload 9
      // 199: iload 4
      // 19b: iadd
      // 19c: istore 9
      // 19e: iload 15
      // 1a0: aload 0
      // 1a1: getfield ua.u I
      // 1a4: iadd
      // 1a5: istore 15
      // 1a7: iload 25
      // 1a9: istore 14
      // 1ab: iinc 26 1
      // 1ae: iload 31
      // 1b0: ifeq 040
      // 1b3: goto 1c4
      // 1b6: astore 25
      // 1b8: getstatic java/lang/System.out Ljava/io/PrintStream;
      // 1bb: getstatic ua.Yb [Ljava/lang/String;
      // 1be: bipush 65
      // 1c0: aaload
      // 1c1: invokevirtual java/io/PrintStream.println (Ljava/lang/String;)V
      // 1c4: goto 2bc
      // 1c7: astore 18
      // 1c9: aload 18
      // 1cb: new java/lang/StringBuilder
      // 1ce: dup
      // 1cf: invokespecial java/lang/StringBuilder.<init> ()V
      // 1d2: getstatic ua.Yb [Ljava/lang/String;
      // 1d5: bipush 67
      // 1d7: aaload
      // 1d8: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 1db: aload 1
      // 1dc: ifnull 1e8
      // 1df: getstatic ua.Yb [Ljava/lang/String;
      // 1e2: bipush 1
      // 1e3: aaload
      // 1e4: goto 1ed
      // 1e7: athrow
      // 1e8: getstatic ua.Yb [Ljava/lang/String;
      // 1eb: bipush 0
      // 1ec: aaload
      // 1ed: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 1f0: bipush 44
      // 1f2: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 1f5: iload 2
      // 1f6: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 1f9: bipush 44
      // 1fb: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 1fe: iload 3
      // 1ff: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 202: bipush 44
      // 204: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 207: iload 4
      // 209: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 20c: bipush 44
      // 20e: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 211: aload 5
      // 213: ifnull 21f
      // 216: getstatic ua.Yb [Ljava/lang/String;
      // 219: bipush 1
      // 21a: aaload
      // 21b: goto 224
      // 21e: athrow
      // 21f: getstatic ua.Yb [Ljava/lang/String;
      // 222: bipush 0
      // 223: aaload
      // 224: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 227: bipush 44
      // 229: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 22c: iload 6
      // 22e: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 231: bipush 44
      // 233: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 236: iload 7
      // 238: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 23b: bipush 44
      // 23d: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 240: iload 8
      // 242: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 245: bipush 44
      // 247: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 24a: iload 9
      // 24c: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 24f: bipush 44
      // 251: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 254: iload 10
      // 256: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 259: bipush 44
      // 25b: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 25e: aload 11
      // 260: ifnull 26c
      // 263: getstatic ua.Yb [Ljava/lang/String;
      // 266: bipush 1
      // 267: aaload
      // 268: goto 271
      // 26b: athrow
      // 26c: getstatic ua.Yb [Ljava/lang/String;
      // 26f: bipush 0
      // 270: aaload
      // 271: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 274: bipush 44
      // 276: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 279: iload 12
      // 27b: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 27e: bipush 44
      // 280: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 283: iload 13
      // 285: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 288: bipush 44
      // 28a: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 28d: iload 14
      // 28f: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 292: bipush 44
      // 294: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 297: iload 15
      // 299: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 29c: bipush 44
      // 29e: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 2a1: iload 16
      // 2a3: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 2a6: bipush 44
      // 2a8: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 2ab: iload 17
      // 2ad: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 2b0: bipush 41
      // 2b2: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 2b5: invokevirtual java/lang/StringBuilder.toString ()Ljava/lang/String;
      // 2b8: invokestatic i.a (Ljava/lang/Throwable;Ljava/lang/String;)Lla;
      // 2bb: athrow
      // 2bc: return
   }

   private final void a(int var1, int var2, String var3, int var4, int var5, int var6, int var7) {
      try {
         this.a(var7, var2, var3, var1 - this.a(var6, 114, var3), var4, (byte)123, var6);
         H++;
         if (var5 != -12200) {
            this.b(75, -128, -127, 3, 49, -124);
         }
      } catch (RuntimeException var10) {
         RuntimeException var8 = var10;

         RuntimeException var10000;
         StringBuilder var10001;
         try {
            var10000 = var8;
            var10001 = new StringBuilder().append(Yb[5]).append(var1).append(',').append(var2).append(',');
            if (var3 != null) {
               throw i.a(
                  var8,
                  var10001.append(Yb[1])
                     .append(',')
                     .append(var4)
                     .append(',')
                     .append(var5)
                     .append(',')
                     .append(var6)
                     .append(',')
                     .append(var7)
                     .append(')')
                     .toString()
               );
            }
         } catch (RuntimeException var9) {
            throw var9;
         }

         throw i.a(
            var10000,
            var10001.append(Yb[0]).append(',').append(var4).append(',').append(var5).append(',').append(var6).append(',').append(var7).append(')').toString()
         );
      }
   }

   // $VF: Handled exception range with multiple entry points by splitting it
   // $VF: Inserted dummy exception handlers to handle obfuscated exceptions
   private final void a(byte[] var1, int var2, int var3, int var4, int var5, int var6, int var7, int var8, int[] var9, int var10) {
      boolean var15 = client.vh;

      try {
         try {
            Fb++;
            if (var7 != 1504725224) {
               this.rb = (int[])null;
            }
         } catch (RuntimeException var24) {
            throw var24;
         }

         int var26 = -var5;

         while (-1 < ~var26 && !var15) {
            int var12 = -var3;

            int var27;
            int var31;
            label109: {
               while (true) {
                  if (var12 < 0) {
                     int var13 = 255 & var1[var10++];

                     label126: {
                        label102: {
                           try {
                              var27 = -31;
                              var31 = ~var13;
                              if (var15) {
                                 break label109;
                              }

                              if (-31 > var31) {
                                 break label102;
                              }
                           } catch (RuntimeException var23) {
                              throw var23;
                           }

                           try {
                              try {
                                 var4++;
                                 if (!var15) {
                                    break label126;
                                 }
                              } catch (RuntimeException var20) {
                                 throw var20;
                              }
                           } catch (RuntimeException var21) {
                              var28 = var21;
                              boolean var32 = false;
                              break;
                           }
                        }

                        label127: {
                           try {
                              if (-231 >= ~var13) {
                                 break label127;
                              }
                           } catch (RuntimeException var22) {
                              var28 = var22;
                              boolean var33 = false;
                              break;
                           }

                           int var14 = var9[var4];

                           try {
                              var9[var4++] = ib.a(-16711936, ib.a(16711935, var2) * var13 - -(ib.a(var14, 16711935) * (-var13 + 256)))
                                    - -ib.a((256 - var13) * ib.a(65280, var14) + var13 * ib.a(65280, var2), 16711680)
                                 >> 1504725224;
                              if (!var15) {
                                 break label126;
                              }
                           } catch (RuntimeException var19) {
                              boolean var34 = false;
                              throw var19;
                           }
                        }

                        try {
                           var9[var4++] = var2;
                        } catch (RuntimeException var18) {
                           boolean var36 = false;
                           throw var18;
                        }
                     }

                     var12++;
                     if (!var15) {
                        continue;
                     }
                  }

                  var10 += var6;
                  var27 = var4;
                  var31 = var8;
                  break label109;
               }

               throw var28;
            }

            var4 = var27 + var31;
            var26++;
            if (var15) {
               break;
            }
         }
      } catch (RuntimeException var25) {
         RuntimeException var11 = var25;

         RuntimeException var10000;
         StringBuilder var10001;
         String var10002;
         label70: {
            try {
               var10000 = var11;
               var10001 = new StringBuilder().append(Yb[64]);
               if (var1 != null) {
                  var10002 = Yb[1];
                  break label70;
               }
            } catch (RuntimeException var17) {
               throw var17;
            }

            var10002 = Yb[0];
         }

         try {
            var10001 = var10001.append(var10002)
               .append(',')
               .append(var2)
               .append(',')
               .append(var3)
               .append(',')
               .append(var4)
               .append(',')
               .append(var5)
               .append(',')
               .append(var6)
               .append(',')
               .append(var7)
               .append(',')
               .append(var8)
               .append(',');
            if (var9 != null) {
               throw i.a(var10000, var10001.append(Yb[1]).append(',').append(var10).append(')').toString());
            }
         } catch (RuntimeException var16) {
            throw var16;
         }

         throw i.a(var10000, var10001.append(Yb[0]).append(',').append(var10).append(')').toString());
      }
   }

   final void a(int var1, String var2, int var3, int var4, int var5, int var6) {
      try {
         this.a(11815, var3, var5, var4, var2, var1, var6);
         ab++;
      } catch (RuntimeException var9) {
         RuntimeException var7 = var9;

         RuntimeException var10000;
         StringBuilder var10001;
         try {
            var10000 = var7;
            var10001 = new StringBuilder().append(Yb[90]).append(var1).append(',');
            if (var2 != null) {
               throw i.a(
                  var7,
                  var10001.append(Yb[1])
                     .append(',')
                     .append(var3)
                     .append(',')
                     .append(var4)
                     .append(',')
                     .append(var5)
                     .append(',')
                     .append(var6)
                     .append(')')
                     .toString()
               );
            }
         } catch (RuntimeException var8) {
            throw var8;
         }

         throw i.a(
            var10000,
            var10001.append(Yb[0]).append(',').append(var3).append(',').append(var4).append(',').append(var5).append(',').append(var6).append(')').toString()
         );
      }
   }

   final void a(byte var1, byte[] var2, int var3) {
      boolean var12 = client.vh;

      try {
         Cb++;
         int[] var19 = this.ob[var3] = new int[10200];

         try {
            this.kb[var3] = 255;
            this.R[var3] = 40;
            this.Sb[var3] = 0;
            this.G[var3] = 0;
            if (var1 != -118) {
               this.a(82, -105, -7, 8, 9);
            }
         } catch (RuntimeException var17) {
            throw var17;
         }

         this.Eb[var3] = 255;
         this.qb[var3] = 40;
         this.Qb[var3] = false;
         int var5 = 0;
         int var6 = 1;
         int var7 = 0;

         byte var22;
         while (true) {
            if (var7 < 255) {
               int var8 = 255 & var2[var6++];
               var22 = 0;
               if (var12) {
                  break;
               }

               int var9 = 0;

               label115: {
                  while (~var8 < ~var9) {
                     try {
                        var19[var7++] = var5;
                        var9++;
                        if (var12) {
                           break label115;
                        }

                        if (var12) {
                           break;
                        }
                     } catch (RuntimeException var16) {
                        throw var16;
                     }
                  }

                  var5 = -var5 + 16777215;
               }

               if (!var12) {
                  continue;
               }
            }

            var22 = 1;
            break;
         }

         int var20 = var22;

         label99:
         while (true) {
            var22 = ~var20;

            label97:
            while (var22 > -41 && !var12) {
               int var21 = 0;

               while (~var21 > -256) {
                  int var10 = var2[var6++] & 255;
                  var22 = 0;
                  if (var12) {
                     continue label97;
                  }

                  int var11 = 0;

                  label91: {
                     while (true) {
                        if (~var11 > ~var10) {
                           try {
                              var19[var7] = var19[-255 + var7];
                              var7++;
                              var21++;
                              var11++;
                              if (var12) {
                                 break;
                              }

                              if (!var12) {
                                 continue;
                              }
                           } catch (RuntimeException var15) {
                              throw var15;
                           }
                        }

                        try {
                           if (~var21 <= -256) {
                              break label91;
                           }
                        } catch (RuntimeException var14) {
                           throw var14;
                        }

                        var19[var7] = 16777215 - var19[var7 - 255];
                        var7++;
                        break;
                     }

                     var21++;
                  }

                  if (var12) {
                     break;
                  }
               }

               var20++;
               if (var12) {
                  return;
               }
               continue label99;
            }

            return;
         }
      } catch (RuntimeException var18) {
         RuntimeException var4 = var18;

         RuntimeException var10000;
         StringBuilder var10001;
         try {
            var10000 = var4;
            var10001 = new StringBuilder().append(Yb[4]).append((int)var1).append(',');
            if (var2 != null) {
               throw i.a(var4, var10001.append(Yb[1]).append(',').append(var3).append(')').toString());
            }
         } catch (RuntimeException var13) {
            throw var13;
         }

         throw i.a(var10000, var10001.append(Yb[0]).append(',').append(var3).append(')').toString());
      }
   }

   // $VF: Handled exception range with multiple entry points by splitting it
   // $VF: Inserted dummy exception handlers to handle obfuscated exceptions
   final void a(boolean var1, int var2) {
      boolean var23 = client.vh;

      try {
         Ob++;
         int var3 = this.R[var2] * this.kb[var2];
         int[] var4 = this.ob[var2];
         int[] var5 = new int[32768];

         try {
            if (var1) {
               this.a(-63, 58, -7, -36, -99);
            }
         } catch (RuntimeException var30) {
            throw var30;
         }

         int var6 = 0;

         int[] var10000;
         while (true) {
            if (var6 < var3) {
               int var7 = var4[var6];

               try {
                  var10000 = var5;
                  if (var23) {
                     break;
                  }

                  var5[(31 & var7 >> 454314147) + (var7 >> 400635145 & 31744) + ((var7 & 63488) >> 303743686)]++;
                  var6++;
                  if (!var23) {
                     continue;
                  }
               } catch (RuntimeException var29) {
                  throw var29;
               }
            }

            var10000 = new int[256];
            break;
         }

         int[] var32 = var10000;
         var32[0] = 16711935;
         int[] var33 = new int[256];
         int var8 = 0;

         label151:
         while (true) {
            int var38 = var8;
            int var10001 = 32768;

            label148:
            while (var38 < var10001) {
               int var9 = var5[var8];

               label176: {
                  try {
                     var39 = var33[255];
                     if (var23) {
                        break label151;
                     }

                     if (var39 >= var9) {
                        break label176;
                     }
                  } catch (RuntimeException var25) {
                     throw var25;
                  }

                  int var10 = 1;

                  while (-257 < ~var10) {
                     var38 = var33[var10];
                     var10001 = var9;
                     if (var23) {
                        continue label148;
                     }

                     if (var38 < var9) {
                        int var11 = 255;

                        while (true) {
                           if (var10 < var11) {
                              try {
                                 var32[var11] = var32[-1 + var11];
                                 var33[var11] = var33[var11 - 1];
                                 var11--;
                                 if (var23) {
                                    break;
                                 }

                                 if (!var23) {
                                    continue;
                                 }
                              } catch (RuntimeException var28) {
                                 throw var28;
                              }
                           }

                           var32[var10] = 263172 + (ib.a(31, var8) << 257025667) + ib.a(63488, var8 << -1695574842) + ib.a(16252928, var8 << 986275945);
                           var33[var10] = var9;
                           break;
                        }

                        try {
                           if (!var23) {
                              break;
                           }
                        } catch (RuntimeException var27) {
                           boolean var42 = false;
                           throw var27;
                        }
                     }

                     try {
                        var10++;
                        if (var23) {
                           break;
                        }
                     } catch (RuntimeException var26) {
                        boolean var43 = false;
                        throw var26;
                     }
                  }
               }

               var5[var8] = -1;
               var8++;
               if (!var23) {
                  continue label151;
               }
               break;
            }

            var39 = var3;
            break;
         }

         byte[] var34 = new byte[var39];
         int var35 = 0;

         label102:
         while (true) {
            int var41 = ~var3;
            int var44 = ~var35;

            label99:
            while (var41 < var44) {
               int var36 = var4[var35];
               int var37 = (var36 >> 1087745987 & 31) + ((var36 & 63488) >> -1963155290) + ((16252928 & var36) >> 10645225);
               int var12 = var5[var37];

               label181: {
                  try {
                     if (var23) {
                        return;
                     }

                     if (-1 != var12) {
                        break label181;
                     }
                  } catch (RuntimeException var24) {
                     throw var24;
                  }

                  int var13 = 999999999;
                  int var14 = 0xFF & var36 >> 100065008;
                  int var15 = 0xFF & var36 >> 862974792;
                  int var16 = var36 & 0xFF;
                  int var17 = 0;

                  while (256 > var17) {
                     int var18 = var32[var17];
                     int var19 = (16719512 & var18) >> -112833712;
                     int var20 = var18 >> 1743095144 & 0xFF;
                     int var21 = 0xFF & var18;
                     int var22 = (-var21 + var16) * (-var21 + var16) + ((var14 - var19) * (var14 - var19) - -((-var20 + var15) * (-var20 + var15)));
                     var41 = var22;
                     var44 = var13;
                     if (var23) {
                        continue label99;
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

                  var5[var37] = var12;
               }

               var34[var35] = (byte)var12;
               var35++;
               if (!var23) {
                  continue label102;
               }
               break;
            }

            this.gb[var2] = var34;
            this.Y[var2] = var32;
            this.ob[var2] = null;
            return;
         }
      } catch (RuntimeException var31) {
         throw i.a(var31, Yb[44] + var1 + ',' + var2 + ')');
      }
   }

   final void a(int param1, String param2, int param3, int param4, int param5, int param6, boolean param7, int param8) {
      // $VF: Couldn't be decompiled
      // Please report this to the Vineflower issue tracker, at https://github.com/Vineflower/vineflower/issues with a copy of the class file (if you have the rights to distribute it!)
      // java.lang.RuntimeException: parsing failure!
      //   at org.jetbrains.java.decompiler.modules.decompiler.decompose.DomHelper.parseGraph(DomHelper.java:211)
      //   at org.jetbrains.java.decompiler.main.rels.MethodProcessor.codeToJava(MethodProcessor.java:166)
      //
      // Bytecode:
      // 000: getstatic client.vh Z
      // 003: istore 15
      // 005: getstatic ua.Q I
      // 008: bipush 1
      // 009: iadd
      // 00a: putstatic ua.Q I
      // 00d: bipush 0
      // 00e: istore 9
      // 010: getstatic m.b [[B
      // 013: iload 5
      // 015: aaload
      // 016: astore 10
      // 018: iload 4
      // 01a: bipush 44
      // 01c: if_icmpge 020
      // 01f: return
      // 020: bipush 0
      // 021: istore 11
      // 023: bipush 0
      // 024: istore 12
      // 026: bipush 0
      // 027: istore 13
      // 029: aload 2
      // 02a: invokevirtual java/lang/String.length ()I
      // 02d: iload 13
      // 02f: if_icmple 175
      // 032: bipush -65
      // 034: iload 15
      // 036: ifne 177
      // 039: aload 2
      // 03a: iload 13
      // 03c: invokevirtual java/lang/String.charAt (I)C
      // 03f: bipush -1
      // 040: ixor
      // 041: if_icmpne 06f
      // 044: goto 048
      // 047: athrow
      // 048: aload 2
      // 049: invokevirtual java/lang/String.length ()I
      // 04c: bipush -1
      // 04d: ixor
      // 04e: bipush 4
      // 04f: iload 13
      // 051: iadd
      // 052: bipush -1
      // 053: ixor
      // 054: if_icmpge 06f
      // 057: goto 05b
      // 05a: athrow
      // 05b: bipush -65
      // 05d: aload 2
      // 05e: iload 13
      // 060: bipush -4
      // 062: isub
      // 063: invokevirtual java/lang/String.charAt (I)C
      // 066: bipush -1
      // 067: ixor
      // 068: if_icmpeq 0ed
      // 06b: goto 06f
      // 06e: athrow
      // 06f: bipush -127
      // 071: aload 2
      // 072: iload 13
      // 074: invokevirtual java/lang/String.charAt (I)C
      // 077: bipush -1
      // 078: ixor
      // 079: if_icmpne 0a2
      // 07c: goto 080
      // 07f: athrow
      // 080: aload 2
      // 081: invokevirtual java/lang/String.length ()I
      // 084: bipush 4
      // 085: iload 13
      // 087: iadd
      // 088: if_icmple 0a2
      // 08b: goto 08f
      // 08e: athrow
      // 08f: aload 2
      // 090: bipush 4
      // 091: iload 13
      // 093: iadd
      // 094: invokevirtual java/lang/String.charAt (I)C
      // 097: bipush -1
      // 098: ixor
      // 099: bipush -127
      // 09b: if_icmpeq 0e1
      // 09e: goto 0a2
      // 0a1: athrow
      // 0a2: aload 2
      // 0a3: iload 13
      // 0a5: invokevirtual java/lang/String.charAt (I)C
      // 0a8: istore 14
      // 0aa: bipush -1
      // 0ab: iload 14
      // 0ad: bipush -1
      // 0ae: ixor
      // 0af: if_icmplt 0c7
      // 0b2: iload 14
      // 0b4: bipush -1
      // 0b5: ixor
      // 0b6: getstatic n.a [I
      // 0b9: arraylength
      // 0ba: bipush -1
      // 0bb: ixor
      // 0bc: if_icmple 0c7
      // 0bf: goto 0c3
      // 0c2: athrow
      // 0c3: goto 0cb
      // 0c6: athrow
      // 0c7: bipush 32
      // 0c9: istore 14
      // 0cb: iload 9
      // 0cd: aload 10
      // 0cf: bipush 7
      // 0d1: getstatic n.a [I
      // 0d4: iload 14
      // 0d6: iaload
      // 0d7: iadd
      // 0d8: baload
      // 0d9: iadd
      // 0da: istore 9
      // 0dc: iload 15
      // 0de: ifeq 0f4
      // 0e1: iinc 13 4
      // 0e4: iload 15
      // 0e6: ifeq 0f4
      // 0e9: goto 0ed
      // 0ec: athrow
      // 0ed: iinc 13 4
      // 0f0: goto 0f4
      // 0f3: athrow
      // 0f4: aload 2
      // 0f5: iload 13
      // 0f7: invokevirtual java/lang/String.charAt (I)C
      // 0fa: bipush 32
      // 0fc: if_icmpne 103
      // 0ff: iload 13
      // 101: istore 12
      // 103: aload 2
      // 104: iload 13
      // 106: invokevirtual java/lang/String.charAt (I)C
      // 109: bipush 37
      // 10b: if_icmpne 124
      // 10e: iload 7
      // 110: ifne 11b
      // 113: goto 117
      // 116: athrow
      // 117: goto 124
      // 11a: athrow
      // 11b: sipush 1000
      // 11e: istore 9
      // 120: iload 13
      // 122: istore 12
      // 124: iload 9
      // 126: iload 1
      // 127: if_icmple 16d
      // 12a: iload 12
      // 12c: iload 11
      // 12e: if_icmple 139
      // 131: goto 135
      // 134: athrow
      // 135: goto 13d
      // 138: athrow
      // 139: iload 13
      // 13b: istore 12
      // 13d: bipush 0
      // 13e: istore 9
      // 140: aload 0
      // 141: sipush 11815
      // 144: iload 8
      // 146: iload 5
      // 148: bipush 0
      // 149: aload 2
      // 14a: iload 11
      // 14c: iload 12
      // 14e: invokevirtual java/lang/String.substring (II)Ljava/lang/String;
      // 151: iload 3
      // 152: iload 6
      // 154: invokespecial ua.a (IIIILjava/lang/String;II)V
      // 157: bipush 1
      // 158: iload 12
      // 15a: iadd
      // 15b: dup
      // 15c: istore 13
      // 15e: istore 11
      // 160: iload 6
      // 162: aload 0
      // 163: ldc 508305352
      // 165: iload 5
      // 167: invokevirtual ua.a (II)I
      // 16a: iadd
      // 16b: istore 6
      // 16d: iinc 13 1
      // 170: iload 15
      // 172: ifeq 029
      // 175: iload 9
      // 177: ifle 193
      // 17a: aload 0
      // 17b: sipush 11815
      // 17e: iload 8
      // 180: iload 5
      // 182: bipush 0
      // 183: aload 2
      // 184: iload 11
      // 186: invokevirtual java/lang/String.substring (I)Ljava/lang/String;
      // 189: iload 3
      // 18a: iload 6
      // 18c: invokespecial ua.a (IIIILjava/lang/String;II)V
      // 18f: goto 193
      // 192: athrow
      // 193: goto 1bb
      // 196: astore 9
      // 198: getstatic java/lang/System.out Ljava/io/PrintStream;
      // 19b: new java/lang/StringBuilder
      // 19e: dup
      // 19f: invokespecial java/lang/StringBuilder.<init> ()V
      // 1a2: getstatic ua.Yb [Ljava/lang/String;
      // 1a5: bipush 81
      // 1a7: aaload
      // 1a8: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 1ab: aload 9
      // 1ad: invokevirtual java/lang/StringBuilder.append (Ljava/lang/Object;)Ljava/lang/StringBuilder;
      // 1b0: invokevirtual java/lang/StringBuilder.toString ()Ljava/lang/String;
      // 1b3: invokevirtual java/io/PrintStream.println (Ljava/lang/String;)V
      // 1b6: aload 9
      // 1b8: invokevirtual java/lang/Exception.printStackTrace ()V
      // 1bb: goto 237
      // 1be: astore 9
      // 1c0: aload 9
      // 1c2: new java/lang/StringBuilder
      // 1c5: dup
      // 1c6: invokespecial java/lang/StringBuilder.<init> ()V
      // 1c9: getstatic ua.Yb [Ljava/lang/String;
      // 1cc: bipush 80
      // 1ce: aaload
      // 1cf: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 1d2: iload 1
      // 1d3: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 1d6: bipush 44
      // 1d8: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 1db: aload 2
      // 1dc: ifnull 1e8
      // 1df: getstatic ua.Yb [Ljava/lang/String;
      // 1e2: bipush 1
      // 1e3: aaload
      // 1e4: goto 1ed
      // 1e7: athrow
      // 1e8: getstatic ua.Yb [Ljava/lang/String;
      // 1eb: bipush 0
      // 1ec: aaload
      // 1ed: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 1f0: bipush 44
      // 1f2: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 1f5: iload 3
      // 1f6: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 1f9: bipush 44
      // 1fb: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 1fe: iload 4
      // 200: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 203: bipush 44
      // 205: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 208: iload 5
      // 20a: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 20d: bipush 44
      // 20f: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 212: iload 6
      // 214: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 217: bipush 44
      // 219: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 21c: iload 7
      // 21e: invokevirtual java/lang/StringBuilder.append (Z)Ljava/lang/StringBuilder;
      // 221: bipush 44
      // 223: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 226: iload 8
      // 228: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 22b: bipush 41
      // 22d: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 230: invokevirtual java/lang/StringBuilder.toString ()Ljava/lang/String;
      // 233: invokestatic i.a (Ljava/lang/Throwable;Ljava/lang/String;)Lla;
      // 236: athrow
      // 237: return
   }

   final void a(int param1, int param2, int param3, boolean param4, int param5, int param6, int param7, int param8, int param9, int param10) {
      // $VF: Couldn't be decompiled
      // Please report this to the Vineflower issue tracker, at https://github.com/Vineflower/vineflower/issues with a copy of the class file (if you have the rights to distribute it!)
      // java.lang.RuntimeException: parsing failure!
      //   at org.jetbrains.java.decompiler.modules.decompiler.decompose.DomHelper.parseGraph(DomHelper.java:211)
      //   at org.jetbrains.java.decompiler.main.rels.MethodProcessor.codeToJava(MethodProcessor.java:166)
      //
      // Bytecode:
      // 000: getstatic client.vh Z
      // 003: istore 24
      // 005: getstatic ua.y I
      // 008: bipush 1
      // 009: iadd
      // 00a: putstatic ua.y I
      // 00d: iload 3
      // 00e: ifeq 014
      // 011: goto 017
      // 014: ldc 16777215
      // 016: istore 3
      // 017: bipush -1
      // 018: iload 2
      // 019: bipush -1
      // 01a: ixor
      // 01b: if_icmpeq 022
      // 01e: goto 025
      // 021: athrow
      // 022: ldc 16777215
      // 024: istore 2
      // 025: aload 0
      // 026: getfield ua.kb [I
      // 029: iload 6
      // 02b: iaload
      // 02c: istore 11
      // 02e: aload 0
      // 02f: getfield ua.R [I
      // 032: iload 6
      // 034: iaload
      // 035: istore 12
      // 037: bipush 0
      // 038: istore 13
      // 03a: bipush 0
      // 03b: istore 14
      // 03d: iload 5
      // 03f: ldc_w -822406960
      // 042: ishl
      // 043: istore 15
      // 045: iload 11
      // 047: ldc_w 264332976
      // 04a: ishl
      // 04b: iload 8
      // 04d: idiv
      // 04e: istore 16
      // 050: iload 12
      // 052: ldc_w -1079220976
      // 055: ishl
      // 056: iload 7
      // 058: idiv
      // 059: istore 17
      // 05b: iload 5
      // 05d: ldc_w -1507499888
      // 060: ishl
      // 061: ineg
      // 062: iload 7
      // 064: idiv
      // 065: istore 18
      // 067: aload 0
      // 068: getfield ua.Qb [Z
      // 06b: iload 6
      // 06d: baload
      // 06e: ifne 075
      // 071: goto 17c
      // 074: athrow
      // 075: aload 0
      // 076: getfield ua.Eb [I
      // 079: iload 6
      // 07b: iaload
      // 07c: istore 19
      // 07e: aload 0
      // 07f: getfield ua.qb [I
      // 082: iload 6
      // 084: iaload
      // 085: istore 20
      // 087: iload 19
      // 089: ifeq 098
      // 08c: iload 20
      // 08e: bipush -1
      // 08f: ixor
      // 090: bipush -1
      // 091: if_icmpne 099
      // 094: goto 098
      // 097: athrow
      // 098: return
      // 099: iload 19
      // 09b: ldc_w -179525200
      // 09e: ishl
      // 09f: iload 8
      // 0a1: idiv
      // 0a2: istore 16
      // 0a4: iload 20
      // 0a6: ldc_w 842218000
      // 0a9: ishl
      // 0aa: iload 7
      // 0ac: idiv
      // 0ad: istore 17
      // 0af: aload 0
      // 0b0: getfield ua.Sb [I
      // 0b3: iload 6
      // 0b5: iaload
      // 0b6: istore 21
      // 0b8: iload 4
      // 0ba: ifne 0c1
      // 0bd: goto 0d1
      // 0c0: athrow
      // 0c1: iload 19
      // 0c3: aload 0
      // 0c4: getfield ua.kb [I
      // 0c7: iload 6
      // 0c9: iaload
      // 0ca: ineg
      // 0cb: iadd
      // 0cc: iload 21
      // 0ce: isub
      // 0cf: istore 21
      // 0d1: aload 0
      // 0d2: getfield ua.G [I
      // 0d5: iload 6
      // 0d7: iaload
      // 0d8: istore 22
      // 0da: iload 9
      // 0dc: bipush -1
      // 0dd: iload 19
      // 0df: iload 21
      // 0e1: iload 8
      // 0e3: imul
      // 0e4: iadd
      // 0e5: iadd
      // 0e6: iload 19
      // 0e8: idiv
      // 0e9: iadd
      // 0ea: istore 9
      // 0ec: bipush -1
      // 0ed: iload 22
      // 0ef: iload 7
      // 0f1: imul
      // 0f2: iload 20
      // 0f4: iadd
      // 0f5: iadd
      // 0f6: iload 20
      // 0f8: idiv
      // 0f9: istore 23
      // 0fb: iload 21
      // 0fd: iload 8
      // 0ff: imul
      // 100: iload 19
      // 102: irem
      // 103: ifeq 11b
      // 106: iload 8
      // 108: iload 21
      // 10a: imul
      // 10b: iload 19
      // 10d: irem
      // 10e: ineg
      // 10f: iload 19
      // 111: iadd
      // 112: ldc_w 306741872
      // 115: ishl
      // 116: iload 8
      // 118: idiv
      // 119: istore 13
      // 11b: iload 1
      // 11c: iload 23
      // 11e: iadd
      // 11f: istore 1
      // 120: iload 15
      // 122: iload 23
      // 124: iload 18
      // 126: imul
      // 127: iadd
      // 128: istore 15
      // 12a: bipush 0
      // 12b: iload 22
      // 12d: iload 7
      // 12f: imul
      // 130: iload 20
      // 132: irem
      // 133: if_icmpeq 14a
      // 136: iload 20
      // 138: iload 7
      // 13a: iload 22
      // 13c: imul
      // 13d: iload 20
      // 13f: irem
      // 140: isub
      // 141: ldc_w -894050704
      // 144: ishl
      // 145: iload 7
      // 147: idiv
      // 148: istore 14
      // 14a: iload 16
      // 14c: aload 0
      // 14d: getfield ua.kb [I
      // 150: iload 6
      // 152: iaload
      // 153: ldc_w -1406651696
      // 156: ishl
      // 157: iload 13
      // 159: bipush 1
      // 15a: iadd
      // 15b: isub
      // 15c: iadd
      // 15d: iload 16
      // 15f: idiv
      // 160: istore 8
      // 162: aload 0
      // 163: getfield ua.R [I
      // 166: iload 6
      // 168: iaload
      // 169: ldc_w -1596145424
      // 16c: ishl
      // 16d: iload 14
      // 16f: ineg
      // 170: iadd
      // 171: iload 17
      // 173: ineg
      // 174: bipush 1
      // 175: iadd
      // 176: isub
      // 177: iload 17
      // 179: idiv
      // 17a: istore 7
      // 17c: aload 0
      // 17d: getfield ua.u I
      // 180: iload 1
      // 181: imul
      // 182: istore 19
      // 184: iload 15
      // 186: iload 9
      // 188: ldc_w 189764144
      // 18b: ishl
      // 18c: iadd
      // 18d: istore 15
      // 18f: iload 1
      // 190: aload 0
      // 191: getfield ua.A I
      // 194: if_icmplt 19b
      // 197: goto 1d0
      // 19a: athrow
      // 19b: aload 0
      // 19c: getfield ua.A I
      // 19f: iload 1
      // 1a0: ineg
      // 1a1: iadd
      // 1a2: istore 20
      // 1a4: iload 15
      // 1a6: iload 18
      // 1a8: iload 20
      // 1aa: imul
      // 1ab: iadd
      // 1ac: istore 15
      // 1ae: iload 7
      // 1b0: iload 20
      // 1b2: isub
      // 1b3: istore 7
      // 1b5: iload 14
      // 1b7: iload 20
      // 1b9: iload 17
      // 1bb: imul
      // 1bc: iadd
      // 1bd: istore 14
      // 1bf: iload 19
      // 1c1: aload 0
      // 1c2: getfield ua.u I
      // 1c5: iload 20
      // 1c7: imul
      // 1c8: iadd
      // 1c9: istore 19
      // 1cb: aload 0
      // 1cc: getfield ua.A I
      // 1cf: istore 1
      // 1d0: iload 1
      // 1d1: iload 7
      // 1d3: ineg
      // 1d4: isub
      // 1d5: aload 0
      // 1d6: getfield ua.Rb I
      // 1d9: if_icmpge 1e0
      // 1dc: goto 1f0
      // 1df: athrow
      // 1e0: iload 7
      // 1e2: bipush 1
      // 1e3: iload 1
      // 1e4: iload 7
      // 1e6: iadd
      // 1e7: iadd
      // 1e8: aload 0
      // 1e9: getfield ua.Rb I
      // 1ec: isub
      // 1ed: isub
      // 1ee: istore 7
      // 1f0: iload 19
      // 1f2: aload 0
      // 1f3: getfield ua.u I
      // 1f6: idiv
      // 1f7: iload 10
      // 1f9: iand
      // 1fa: istore 20
      // 1fc: aload 0
      // 1fd: getfield ua.i Z
      // 200: ifeq 207
      // 203: goto 20a
      // 206: athrow
      // 207: bipush 2
      // 208: istore 20
      // 20a: iload 3
      // 20b: bipush -1
      // 20c: ixor
      // 20d: ldc_w -16777216
      // 210: if_icmpne 333
      // 213: aconst_null
      // 214: aload 0
      // 215: getfield ua.ob [[I
      // 218: iload 6
      // 21a: aaload
      // 21b: if_acmpeq 2a5
      // 21e: goto 222
      // 221: athrow
      // 222: iload 4
      // 224: ifeq 270
      // 227: goto 22b
      // 22a: athrow
      // 22b: aload 0
      // 22c: iload 10
      // 22e: bipush 74
      // 230: ixor
      // 231: aload 0
      // 232: getfield ua.ob [[I
      // 235: iload 6
      // 237: aaload
      // 238: iload 17
      // 23a: bipush 0
      // 23b: iload 14
      // 23d: aload 0
      // 23e: getfield ua.kb [I
      // 241: iload 6
      // 243: iaload
      // 244: ldc_w 102617264
      // 247: ishl
      // 248: iload 13
      // 24a: ineg
      // 24b: bipush 1
      // 24c: isub
      // 24d: iadd
      // 24e: iload 8
      // 250: aload 0
      // 251: getfield ua.rb [I
      // 254: iload 7
      // 256: iload 18
      // 258: iload 19
      // 25a: iload 16
      // 25c: ineg
      // 25d: iload 15
      // 25f: iload 11
      // 261: iload 20
      // 263: iload 2
      // 264: invokespecial ua.a (I[IIIIII[IIIIIIIII)V
      // 267: iload 24
      // 269: ifeq 44c
      // 26c: goto 270
      // 26f: athrow
      // 270: aload 0
      // 271: iload 10
      // 273: bipush 89
      // 275: iadd
      // 276: aload 0
      // 277: getfield ua.ob [[I
      // 27a: iload 6
      // 27c: aaload
      // 27d: iload 17
      // 27f: bipush 0
      // 280: iload 14
      // 282: iload 13
      // 284: iload 8
      // 286: aload 0
      // 287: getfield ua.rb [I
      // 28a: iload 7
      // 28c: iload 18
      // 28e: iload 19
      // 290: iload 16
      // 292: iload 15
      // 294: iload 11
      // 296: iload 20
      // 298: iload 2
      // 299: invokespecial ua.a (I[IIIIII[IIIIIIIII)V
      // 29c: iload 24
      // 29e: ifeq 44c
      // 2a1: goto 2a5
      // 2a4: athrow
      // 2a5: iload 4
      // 2a7: ifne 2e7
      // 2aa: goto 2ae
      // 2ad: athrow
      // 2ae: aload 0
      // 2af: aload 0
      // 2b0: getfield ua.gb [[B
      // 2b3: iload 6
      // 2b5: aaload
      // 2b6: iload 2
      // 2b7: bipush 0
      // 2b8: iload 18
      // 2ba: aload 0
      // 2bb: getfield ua.Y [[I
      // 2be: iload 6
      // 2c0: aaload
      // 2c1: iload 11
      // 2c3: iload 14
      // 2c5: iload 17
      // 2c7: iload 15
      // 2c9: iload 16
      // 2cb: aload 0
      // 2cc: getfield ua.rb [I
      // 2cf: iload 20
      // 2d1: bipush -110
      // 2d3: iload 13
      // 2d5: iload 19
      // 2d7: iload 8
      // 2d9: iload 7
      // 2db: invokespecial ua.a ([BIII[IIIIII[IIIIIII)V
      // 2de: iload 24
      // 2e0: ifeq 44c
      // 2e3: goto 2e7
      // 2e6: athrow
      // 2e7: aload 0
      // 2e8: aload 0
      // 2e9: getfield ua.gb [[B
      // 2ec: iload 6
      // 2ee: aaload
      // 2ef: iload 2
      // 2f0: bipush 0
      // 2f1: iload 18
      // 2f3: aload 0
      // 2f4: getfield ua.Y [[I
      // 2f7: iload 6
      // 2f9: aaload
      // 2fa: iload 11
      // 2fc: iload 14
      // 2fe: iload 17
      // 300: iload 15
      // 302: iload 16
      // 304: ineg
      // 305: aload 0
      // 306: getfield ua.rb [I
      // 309: iload 20
      // 30b: iload 10
      // 30d: bipush -124
      // 30f: ixor
      // 310: bipush -1
      // 311: iload 13
      // 313: ineg
      // 314: aload 0
      // 315: getfield ua.kb [I
      // 318: iload 6
      // 31a: iaload
      // 31b: ldc_w -1997207152
      // 31e: ishl
      // 31f: iadd
      // 320: iadd
      // 321: iload 19
      // 323: iload 8
      // 325: iload 7
      // 327: invokespecial ua.a ([BIII[IIIIII[IIIIIII)V
      // 32a: iload 24
      // 32c: ifeq 44c
      // 32f: goto 333
      // 332: athrow
      // 333: aload 0
      // 334: getfield ua.ob [[I
      // 337: iload 6
      // 339: aaload
      // 33a: ifnonnull 3ce
      // 33d: goto 341
      // 340: athrow
      // 341: iload 4
      // 343: ifeq 394
      // 346: goto 34a
      // 349: athrow
      // 34a: aload 0
      // 34b: iload 7
      // 34d: aload 0
      // 34e: getfield ua.Y [[I
      // 351: iload 6
      // 353: aaload
      // 354: iload 11
      // 356: iload 17
      // 358: bipush 0
      // 359: iload 18
      // 35b: iload 19
      // 35d: iload 14
      // 35f: aload 0
      // 360: getfield ua.gb [[B
      // 363: iload 6
      // 365: aaload
      // 366: iload 20
      // 368: bipush 76
      // 36a: iload 13
      // 36c: ineg
      // 36d: aload 0
      // 36e: getfield ua.kb [I
      // 371: iload 6
      // 373: iaload
      // 374: ldc_w -1651772048
      // 377: ishl
      // 378: bipush -1
      // 379: iadd
      // 37a: iadd
      // 37b: iload 3
      // 37c: iload 15
      // 37e: iload 16
      // 380: ineg
      // 381: iload 8
      // 383: aload 0
      // 384: getfield ua.rb [I
      // 387: iload 2
      // 388: invokespecial ua.a (I[IIIIIII[BIBIIIII[II)V
      // 38b: iload 24
      // 38d: ifeq 44c
      // 390: goto 394
      // 393: athrow
      // 394: aload 0
      // 395: iload 7
      // 397: aload 0
      // 398: getfield ua.Y [[I
      // 39b: iload 6
      // 39d: aaload
      // 39e: iload 11
      // 3a0: iload 17
      // 3a2: bipush 0
      // 3a3: iload 18
      // 3a5: iload 19
      // 3a7: iload 14
      // 3a9: aload 0
      // 3aa: getfield ua.gb [[B
      // 3ad: iload 6
      // 3af: aaload
      // 3b0: iload 20
      // 3b2: bipush 78
      // 3b4: iload 13
      // 3b6: iload 3
      // 3b7: iload 15
      // 3b9: iload 16
      // 3bb: iload 8
      // 3bd: aload 0
      // 3be: getfield ua.rb [I
      // 3c1: iload 2
      // 3c2: invokespecial ua.a (I[IIIIIII[BIBIIIII[II)V
      // 3c5: iload 24
      // 3c7: ifeq 44c
      // 3ca: goto 3ce
      // 3cd: athrow
      // 3ce: iload 4
      // 3d0: ifeq 41e
      // 3d3: goto 3d7
      // 3d6: athrow
      // 3d7: aload 0
      // 3d8: aload 0
      // 3d9: getfield ua.rb [I
      // 3dc: aload 0
      // 3dd: getfield ua.ob [[I
      // 3e0: iload 6
      // 3e2: aaload
      // 3e3: iload 8
      // 3e5: iload 18
      // 3e7: iload 15
      // 3e9: iload 10
      // 3eb: ldc_w 1603920391
      // 3ee: iadd
      // 3ef: bipush 0
      // 3f0: iload 3
      // 3f1: iload 17
      // 3f3: iload 16
      // 3f5: ineg
      // 3f6: iload 13
      // 3f8: ineg
      // 3f9: aload 0
      // 3fa: getfield ua.kb [I
      // 3fd: iload 6
      // 3ff: iaload
      // 400: ldc_w -1212875536
      // 403: ishl
      // 404: iadd
      // 405: bipush -1
      // 406: iadd
      // 407: iload 20
      // 409: iload 14
      // 40b: iload 11
      // 40d: iload 2
      // 40e: iload 7
      // 410: iload 19
      // 412: invokespecial ua.a ([I[IIIIIIIIIIIIIIII)V
      // 415: iload 24
      // 417: ifeq 44c
      // 41a: goto 41e
      // 41d: athrow
      // 41e: aload 0
      // 41f: aload 0
      // 420: getfield ua.rb [I
      // 423: aload 0
      // 424: getfield ua.ob [[I
      // 427: iload 6
      // 429: aaload
      // 42a: iload 8
      // 42c: iload 18
      // 42e: iload 15
      // 430: ldc 1603920392
      // 432: bipush 0
      // 433: iload 3
      // 434: iload 17
      // 436: iload 16
      // 438: iload 13
      // 43a: iload 20
      // 43c: iload 14
      // 43e: iload 11
      // 440: iload 2
      // 441: iload 7
      // 443: iload 19
      // 445: invokespecial ua.a ([I[IIIIIIIIIIIIIIII)V
      // 448: goto 44c
      // 44b: athrow
      // 44c: goto 45d
      // 44f: astore 11
      // 451: getstatic java/lang/System.out Ljava/io/PrintStream;
      // 454: getstatic ua.Yb [Ljava/lang/String;
      // 457: bipush 16
      // 459: aaload
      // 45a: invokevirtual java/io/PrintStream.println (Ljava/lang/String;)V
      // 45d: goto 4dc
      // 460: astore 11
      // 462: aload 11
      // 464: new java/lang/StringBuilder
      // 467: dup
      // 468: invokespecial java/lang/StringBuilder.<init> ()V
      // 46b: getstatic ua.Yb [Ljava/lang/String;
      // 46e: bipush 71
      // 470: aaload
      // 471: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 474: iload 1
      // 475: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 478: bipush 44
      // 47a: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 47d: iload 2
      // 47e: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 481: bipush 44
      // 483: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 486: iload 3
      // 487: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 48a: bipush 44
      // 48c: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 48f: iload 4
      // 491: invokevirtual java/lang/StringBuilder.append (Z)Ljava/lang/StringBuilder;
      // 494: bipush 44
      // 496: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 499: iload 5
      // 49b: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 49e: bipush 44
      // 4a0: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 4a3: iload 6
      // 4a5: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 4a8: bipush 44
      // 4aa: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 4ad: iload 7
      // 4af: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 4b2: bipush 44
      // 4b4: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 4b7: iload 8
      // 4b9: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 4bc: bipush 44
      // 4be: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 4c1: iload 9
      // 4c3: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 4c6: bipush 44
      // 4c8: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 4cb: iload 10
      // 4cd: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 4d0: bipush 41
      // 4d2: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 4d5: invokevirtual java/lang/StringBuilder.toString ()Ljava/lang/String;
      // 4d8: invokestatic i.a (Ljava/lang/Throwable;Ljava/lang/String;)Lla;
      // 4db: athrow
      // 4dc: return
   }

   @Override
   public final void startProduction(ImageConsumer var1) {
      try {
         this.addConsumer(var1);
         mb++;
      } catch (RuntimeException var4) {
         RuntimeException var2 = var4;

         RuntimeException var10000;
         StringBuilder var10001;
         try {
            var10000 = var2;
            var10001 = new StringBuilder().append(Yb[12]);
            if (var1 != null) {
               throw i.a(var2, var10001.append(Yb[1]).append(')').toString());
            }
         } catch (RuntimeException var3) {
            throw var3;
         }

         throw i.a(var10000, var10001.append(Yb[0]).append(')').toString());
      }
   }

   @Override
   public final synchronized boolean isConsumer(ImageConsumer var1) {
      try {
         q++;
         return this.fb == var1;
      } catch (RuntimeException var4) {
         RuntimeException var2 = var4;

         RuntimeException var10000;
         StringBuilder var10001;
         try {
            var10000 = var2;
            var10001 = new StringBuilder().append(Yb[2]);
            if (var1 != null) {
               throw i.a(var2, var10001.append(Yb[1]).append(')').toString());
            }
         } catch (RuntimeException var3) {
            throw var3;
         }

         throw i.a(var10000, var10001.append(Yb[0]).append(')').toString());
      }
   }

   private final void a(
      int param1, int[] param2, int param3, int param4, int param5, int param6, byte param7, int param8, int[] param9, int param10, int param11
   ) {
      // $VF: Couldn't be decompiled
      // Please report this to the Vineflower issue tracker, at https://github.com/Vineflower/vineflower/issues with a copy of the class file (if you have the rights to distribute it!)
      // java.lang.RuntimeException: parsing failure!
      //   at org.jetbrains.java.decompiler.modules.decompiler.decompose.DomHelper.parseGraph(DomHelper.java:211)
      //   at org.jetbrains.java.decompiler.main.rels.MethodProcessor.codeToJava(MethodProcessor.java:166)
      //
      // Bytecode:
      // 000: getstatic client.vh Z
      // 003: istore 15
      // 005: getstatic ua.x I
      // 008: bipush 1
      // 009: iadd
      // 00a: putstatic ua.x I
      // 00d: iload 7
      // 00f: bipush 122
      // 011: if_icmpgt 028
      // 014: aload 0
      // 015: bipush 121
      // 017: bipush 54
      // 019: bipush -117
      // 01b: bipush -34
      // 01d: bipush 67
      // 01f: bipush -103
      // 021: invokevirtual ua.e (IIIIII)V
      // 024: goto 028
      // 027: athrow
      // 028: iload 1
      // 029: ldc_w -1677003518
      // 02c: ishr
      // 02d: ineg
      // 02e: istore 12
      // 030: bipush 3
      // 031: iload 1
      // 032: iand
      // 033: ineg
      // 034: istore 1
      // 035: iload 4
      // 037: ineg
      // 038: istore 13
      // 03a: iload 13
      // 03c: bipush -1
      // 03d: ixor
      // 03e: bipush -1
      // 03f: if_icmple 164
      // 042: iload 15
      // 044: ifne 20f
      // 047: iload 12
      // 049: istore 14
      // 04b: iload 14
      // 04d: ifge 106
      // 050: aload 9
      // 052: iload 6
      // 054: iinc 6 1
      // 057: iaload
      // 058: istore 5
      // 05a: iload 5
      // 05c: iload 15
      // 05e: ifne 107
      // 061: ifne 074
      // 064: goto 068
      // 067: athrow
      // 068: iinc 8 1
      // 06b: iload 15
      // 06d: ifeq 081
      // 070: goto 074
      // 073: athrow
      // 074: aload 2
      // 075: iload 8
      // 077: iinc 8 1
      // 07a: iload 5
      // 07c: iastore
      // 07d: goto 081
      // 080: athrow
      // 081: aload 9
      // 083: iload 6
      // 085: iinc 6 1
      // 088: iaload
      // 089: istore 5
      // 08b: iload 5
      // 08d: bipush -1
      // 08e: ixor
      // 08f: bipush -1
      // 090: if_icmpne 09f
      // 093: iinc 8 1
      // 096: iload 15
      // 098: ifeq 0ac
      // 09b: goto 09f
      // 09e: athrow
      // 09f: aload 2
      // 0a0: iload 8
      // 0a2: iinc 8 1
      // 0a5: iload 5
      // 0a7: iastore
      // 0a8: goto 0ac
      // 0ab: athrow
      // 0ac: aload 9
      // 0ae: iload 6
      // 0b0: iinc 6 1
      // 0b3: iaload
      // 0b4: istore 5
      // 0b6: bipush 0
      // 0b7: iload 5
      // 0b9: if_icmpeq 0ce
      // 0bc: aload 2
      // 0bd: iload 8
      // 0bf: iinc 8 1
      // 0c2: iload 5
      // 0c4: iastore
      // 0c5: iload 15
      // 0c7: ifeq 0d5
      // 0ca: goto 0ce
      // 0cd: athrow
      // 0ce: iinc 8 1
      // 0d1: goto 0d5
      // 0d4: athrow
      // 0d5: aload 9
      // 0d7: iload 6
      // 0d9: iinc 6 1
      // 0dc: iaload
      // 0dd: istore 5
      // 0df: bipush 0
      // 0e0: iload 5
      // 0e2: if_icmpne 0f1
      // 0e5: iinc 8 1
      // 0e8: iload 15
      // 0ea: ifeq 0fe
      // 0ed: goto 0f1
      // 0f0: athrow
      // 0f1: aload 2
      // 0f2: iload 8
      // 0f4: iinc 8 1
      // 0f7: iload 5
      // 0f9: iastore
      // 0fa: goto 0fe
      // 0fd: athrow
      // 0fe: iinc 14 1
      // 101: iload 15
      // 103: ifeq 04b
      // 106: iload 1
      // 107: istore 14
      // 109: bipush -1
      // 10a: iload 14
      // 10c: bipush -1
      // 10d: ixor
      // 10e: if_icmpge 14b
      // 111: aload 9
      // 113: iload 6
      // 115: iinc 6 1
      // 118: iaload
      // 119: istore 5
      // 11b: bipush 0
      // 11c: iload 5
      // 11e: iload 15
      // 120: ifne 15c
      // 123: if_icmpne 136
      // 126: goto 12a
      // 129: athrow
      // 12a: iinc 8 1
      // 12d: iload 15
      // 12f: ifeq 143
      // 132: goto 136
      // 135: athrow
      // 136: aload 2
      // 137: iload 8
      // 139: iinc 8 1
      // 13c: iload 5
      // 13e: iastore
      // 13f: goto 143
      // 142: athrow
      // 143: iinc 14 1
      // 146: iload 15
      // 148: ifeq 109
      // 14b: iload 6
      // 14d: iload 11
      // 14f: iadd
      // 150: istore 6
      // 152: iload 8
      // 154: iload 10
      // 156: iadd
      // 157: istore 8
      // 159: iload 13
      // 15b: iload 3
      // 15c: iadd
      // 15d: istore 13
      // 15f: iload 15
      // 161: ifeq 03a
      // 164: goto 20f
      // 167: astore 12
      // 169: aload 12
      // 16b: new java/lang/StringBuilder
      // 16e: dup
      // 16f: invokespecial java/lang/StringBuilder.<init> ()V
      // 172: getstatic ua.Yb [Ljava/lang/String;
      // 175: bipush 8
      // 177: aaload
      // 178: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 17b: iload 1
      // 17c: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 17f: bipush 44
      // 181: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 184: aload 2
      // 185: ifnull 191
      // 188: getstatic ua.Yb [Ljava/lang/String;
      // 18b: bipush 1
      // 18c: aaload
      // 18d: goto 196
      // 190: athrow
      // 191: getstatic ua.Yb [Ljava/lang/String;
      // 194: bipush 0
      // 195: aaload
      // 196: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 199: bipush 44
      // 19b: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 19e: iload 3
      // 19f: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 1a2: bipush 44
      // 1a4: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 1a7: iload 4
      // 1a9: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 1ac: bipush 44
      // 1ae: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 1b1: iload 5
      // 1b3: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 1b6: bipush 44
      // 1b8: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 1bb: iload 6
      // 1bd: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 1c0: bipush 44
      // 1c2: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 1c5: iload 7
      // 1c7: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 1ca: bipush 44
      // 1cc: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 1cf: iload 8
      // 1d1: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 1d4: bipush 44
      // 1d6: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 1d9: aload 9
      // 1db: ifnull 1e7
      // 1de: getstatic ua.Yb [Ljava/lang/String;
      // 1e1: bipush 1
      // 1e2: aaload
      // 1e3: goto 1ec
      // 1e6: athrow
      // 1e7: getstatic ua.Yb [Ljava/lang/String;
      // 1ea: bipush 0
      // 1eb: aaload
      // 1ec: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 1ef: bipush 44
      // 1f1: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 1f4: iload 10
      // 1f6: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 1f9: bipush 44
      // 1fb: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 1fe: iload 11
      // 200: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 203: bipush 41
      // 205: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 208: invokevirtual java/lang/StringBuilder.toString ()Ljava/lang/String;
      // 20b: invokestatic i.a (Ljava/lang/Throwable;Ljava/lang/String;)Lla;
      // 20e: athrow
      // 20f: return
   }

   final void a(int var1, int var2, int var3, int var4, byte var5) {
      try {
         label52: {
            try {
               if (this.k >= var3) {
                  break label52;
               }
            } catch (RuntimeException var10) {
               throw var10;
            }

            var3 = this.k;
         }

         if (0 > var4) {
            var4 = 0;
         }

         label46: {
            try {
               if (-1 >= ~var1) {
                  break label46;
               }
            } catch (RuntimeException var9) {
               throw var9;
            }

            var1 = 0;
         }

         label41: {
            try {
               n++;
               if (~this.u <= ~var2) {
                  break label41;
               }
            } catch (RuntimeException var8) {
               throw var8;
            }

            var2 = this.u;
         }

         try {
            if (var5 <= 15) {
               C = 109;
            }
         } catch (RuntimeException var7) {
            throw var7;
         }

         this.A = var4;
         this.Rb = var3;
         this.lb = var2;
         this.hb = var1;
      } catch (RuntimeException var11) {
         throw i.a(var11, Yb[6] + var1 + ',' + var2 + ',' + var3 + ',' + var4 + ',' + var5 + ')');
      }
   }

   final void a(boolean var1) {
      boolean var6 = client.vh;

      try {
         f++;
         int var2 = this.k * this.u;

         boolean var10000;
         boolean var10001;
         label77: {
            try {
               var10000 = var1;
               if (!this.i) {
                  var10001 = true;
                  break label77;
               }
            } catch (RuntimeException var9) {
               throw var9;
            }

            var10001 = false;
         }

         if (var10000 == var10001) {
            int var3 = 0;

            while (var2 > var3) {
               try {
                  this.rb[var3] = 0;
                  var3++;
                  if (var6) {
                     return;
                  }

                  if (var6) {
                     break;
                  }
               } catch (RuntimeException var8) {
                  throw var8;
               }
            }

            if (!var6) {
               return;
            }
         }

         int var11 = 0;
         int var4 = -this.k;

         while (~var4 > -1 && !var6) {
            int var5 = -this.u;

            label52: {
               while (0 > var5) {
                  try {
                     this.rb[var11++] = 0;
                     var5++;
                     if (var6) {
                        break label52;
                     }

                     if (var6) {
                        break;
                     }
                  } catch (RuntimeException var7) {
                     throw var7;
                  }
               }

               var11 += this.u;
               var4 += 2;
            }

            if (var6) {
               break;
            }
         }
      } catch (RuntimeException var10) {
         throw i.a(var10, Yb[51] + var1 + ')');
      }
   }

   final void a(int var1, byte var2, int var3, int var4, int var5, int var6) {
      boolean var13 = client.vh;

      try {
         if (var1 < this.hb) {
            var6 -= this.hb + -var1;
            var1 = this.hb;
         }

         label89: {
            try {
               if (this.A <= var4) {
                  break label89;
               }
            } catch (RuntimeException var18) {
               throw var18;
            }

            var5 -= -var4 + this.A;
            var4 = this.A;
         }

         label84: {
            try {
               K++;
               if (~(var4 + var5) >= ~this.Rb) {
                  break label84;
               }
            } catch (RuntimeException var17) {
               throw var17;
            }

            var5 = -var4 + this.Rb;
         }

         label79: {
            try {
               if (~(var6 + var1) >= ~this.lb) {
                  break label79;
               }
            } catch (RuntimeException var16) {
               throw var16;
            }

            var6 = -var1 + this.lb;
         }

         int var7 = this.u + -var6;
         byte var8 = 1;
         int var9 = -124 / ((-39 - var2) / 59);
         if (this.i) {
            var7 += this.u;

            label71: {
               try {
                  if (0 == (var4 & 1)) {
                     break label71;
                  }
               } catch (RuntimeException var15) {
                  throw var15;
               }

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
                  try {
                     this.rb[var10++] = var3;
                     var12++;
                     if (var13) {
                        break;
                     }

                     if (!var13) {
                        continue;
                     }
                  } catch (RuntimeException var14) {
                     throw var14;
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
      } catch (RuntimeException var19) {
         throw i.a(var19, Yb[3] + var1 + ',' + var2 + ',' + var3 + ',' + var4 + ',' + var5 + ',' + var6 + ')');
      }
   }

   final void a(String var1, int var2, int var3, int var4, boolean var5, int var6) {
      try {
         pb++;
         this.a(0, var3, var1, var2, var4, (byte)124, var6);
         if (var5) {
            this.a(-43, 36, -60, -88, 93, (int[])null, 114, (int[])null, -53, 59, true, 66, 34, 34, 70);
         }
      } catch (RuntimeException var9) {
         RuntimeException var7 = var9;

         RuntimeException var10000;
         StringBuilder var10001;
         try {
            var10000 = var7;
            var10001 = new StringBuilder().append(Yb[50]);
            if (var1 != null) {
               throw i.a(
                  var7,
                  var10001.append(Yb[1])
                     .append(',')
                     .append(var2)
                     .append(',')
                     .append(var3)
                     .append(',')
                     .append(var4)
                     .append(',')
                     .append(var5)
                     .append(',')
                     .append(var6)
                     .append(')')
                     .toString()
               );
            }
         } catch (RuntimeException var8) {
            throw var8;
         }

         throw i.a(
            var10000,
            var10001.append(Yb[0])
               .append(',')
               .append(var2)
               .append(',')
               .append(var3)
               .append(',')
               .append(var4)
               .append(',')
               .append(var5)
               .append(',')
               .append(var6)
               .append(')')
               .toString()
         );
      }
   }

   final void a(int var1) {
      try {
         try {
            this.lb = this.u;
            if (var1 != -1) {
               this.Sb = (int[])null;
            }
         } catch (RuntimeException var3) {
            throw var3;
         }

         this.hb = 0;
         this.A = 0;
         this.Rb = this.k;
         W++;
      } catch (RuntimeException var4) {
         throw i.a(var4, Yb[83] + var1 + ')');
      }
   }

   final void a(int param1, int param2, byte[] param3, int param4, byte[] param5) {
      // $VF: Couldn't be decompiled
      // Please report this to the Vineflower issue tracker, at https://github.com/Vineflower/vineflower/issues with a copy of the class file (if you have the rights to distribute it!)
      // java.lang.RuntimeException: parsing failure!
      //   at org.jetbrains.java.decompiler.modules.decompiler.decompose.DomHelper.parseGraph(DomHelper.java:211)
      //   at org.jetbrains.java.decompiler.main.rels.MethodProcessor.codeToJava(MethodProcessor.java:166)
      //
      // Bytecode:
      // 000: getstatic client.vh Z
      // 003: istore 17
      // 005: getstatic ua.jb I
      // 008: bipush 1
      // 009: iadd
      // 00a: putstatic ua.jb I
      // 00d: bipush 0
      // 00e: bipush 41
      // 010: aload 3
      // 011: invokestatic d.a (IB[B)I
      // 014: istore 6
      // 016: iload 6
      // 018: bipush 48
      // 01a: aload 5
      // 01c: invokestatic d.a (IB[B)I
      // 01f: istore 7
      // 021: iinc 6 2
      // 024: iload 4
      // 026: bipush 49
      // 028: if_icmpge 040
      // 02b: aload 0
      // 02c: bipush -97
      // 02e: bipush -40
      // 030: bipush 4
      // 031: bipush -57
      // 033: bipush 24
      // 035: bipush 50
      // 037: bipush 82
      // 039: invokevirtual ua.b (IIIIIII)V
      // 03c: goto 040
      // 03f: athrow
      // 040: iload 6
      // 042: bipush 15
      // 044: aload 5
      // 046: invokestatic d.a (IB[B)I
      // 049: istore 8
      // 04b: iinc 6 2
      // 04e: aload 5
      // 050: iload 6
      // 052: iinc 6 1
      // 055: baload
      // 056: sipush 255
      // 059: iand
      // 05a: istore 9
      // 05c: iload 9
      // 05e: newarray 10
      // 060: astore 10
      // 062: aload 10
      // 064: bipush 0
      // 065: ldc 16711935
      // 067: iastore
      // 068: bipush 0
      // 069: istore 11
      // 06b: iload 9
      // 06d: bipush -1
      // 06e: iadd
      // 06f: iload 11
      // 071: if_icmple 0be
      // 074: aload 10
      // 076: iload 11
      // 078: bipush -1
      // 079: isub
      // 07a: sipush 255
      // 07d: aload 5
      // 07f: iload 6
      // 081: baload
      // 082: invokestatic ib.a (II)I
      // 085: ldc_w -7887792
      // 088: ishl
      // 089: ldc 65280
      // 08b: aload 5
      // 08d: iload 6
      // 08f: bipush -1
      // 090: isub
      // 091: baload
      // 092: ldc_w -679902488
      // 095: ishl
      // 096: invokestatic ib.a (II)I
      // 099: ineg
      // 09a: isub
      // 09b: aload 5
      // 09d: iload 6
      // 09f: bipush 2
      // 0a0: iadd
      // 0a1: baload
      // 0a2: sipush 255
      // 0a5: invokestatic ib.a (II)I
      // 0a8: iadd
      // 0a9: iastore
      // 0aa: iinc 6 3
      // 0ad: iinc 11 1
      // 0b0: iload 17
      // 0b2: ifne 0c1
      // 0b5: iload 17
      // 0b7: ifeq 06b
      // 0ba: goto 0be
      // 0bd: athrow
      // 0be: bipush 2
      // 0bf: istore 11
      // 0c1: iload 1
      // 0c2: istore 12
      // 0c4: iload 12
      // 0c6: bipush -1
      // 0c7: ixor
      // 0c8: iload 1
      // 0c9: iload 2
      // 0ca: iadd
      // 0cb: bipush -1
      // 0cc: ixor
      // 0cd: if_icmple 28a
      // 0d0: aload 0
      // 0d1: getfield ua.Sb [I
      // 0d4: iload 12
      // 0d6: aload 5
      // 0d8: iload 6
      // 0da: iinc 6 1
      // 0dd: baload
      // 0de: sipush 255
      // 0e1: invokestatic ib.a (II)I
      // 0e4: iastore
      // 0e5: aload 0
      // 0e6: getfield ua.G [I
      // 0e9: iload 12
      // 0eb: aload 5
      // 0ed: iload 6
      // 0ef: iinc 6 1
      // 0f2: baload
      // 0f3: sipush 255
      // 0f6: invokestatic ib.a (II)I
      // 0f9: iastore
      // 0fa: aload 0
      // 0fb: getfield ua.kb [I
      // 0fe: iload 12
      // 100: iload 6
      // 102: bipush 32
      // 104: aload 5
      // 106: invokestatic d.a (IB[B)I
      // 109: iastore
      // 10a: iinc 6 2
      // 10d: aload 0
      // 10e: getfield ua.R [I
      // 111: iload 12
      // 113: iload 6
      // 115: bipush 83
      // 117: aload 5
      // 119: invokestatic d.a (IB[B)I
      // 11c: iastore
      // 11d: iinc 6 2
      // 120: aload 5
      // 122: iload 6
      // 124: iinc 6 1
      // 127: baload
      // 128: sipush 255
      // 12b: iand
      // 12c: istore 13
      // 12e: aload 0
      // 12f: getfield ua.R [I
      // 132: iload 12
      // 134: iaload
      // 135: aload 0
      // 136: getfield ua.kb [I
      // 139: iload 12
      // 13b: iaload
      // 13c: imul
      // 13d: istore 14
      // 13f: aload 0
      // 140: getfield ua.gb [[B
      // 143: iload 12
      // 145: iload 14
      // 147: newarray 8
      // 149: aastore
      // 14a: aload 0
      // 14b: getfield ua.Y [[I
      // 14e: iload 12
      // 150: aload 10
      // 152: aastore
      // 153: aload 0
      // 154: getfield ua.Eb [I
      // 157: iload 12
      // 159: iload 7
      // 15b: iastore
      // 15c: aload 0
      // 15d: getfield ua.qb [I
      // 160: iload 12
      // 162: iload 8
      // 164: iastore
      // 165: aload 0
      // 166: getfield ua.ob [[I
      // 169: iload 12
      // 16b: aconst_null
      // 16c: aastore
      // 16d: aload 0
      // 16e: getfield ua.Qb [Z
      // 171: iload 12
      // 173: bipush 0
      // 174: bastore
      // 175: iload 17
      // 177: ifne 2f9
      // 17a: bipush -1
      // 17b: aload 0
      // 17c: getfield ua.Sb [I
      // 17f: iload 12
      // 181: iaload
      // 182: bipush -1
      // 183: ixor
      // 184: if_icmpne 19a
      // 187: goto 18b
      // 18a: athrow
      // 18b: bipush 0
      // 18c: aload 0
      // 18d: getfield ua.G [I
      // 190: iload 12
      // 192: iaload
      // 193: if_icmpeq 1a6
      // 196: goto 19a
      // 199: athrow
      // 19a: aload 0
      // 19b: getfield ua.Qb [Z
      // 19e: iload 12
      // 1a0: bipush 1
      // 1a1: bastore
      // 1a2: goto 1a6
      // 1a5: athrow
      // 1a6: iload 13
      // 1a8: bipush -1
      // 1a9: ixor
      // 1aa: bipush -1
      // 1ab: if_icmpeq 240
      // 1ae: iload 13
      // 1b0: bipush -1
      // 1b1: ixor
      // 1b2: bipush -2
      // 1b4: if_icmpeq 1bf
      // 1b7: goto 1bb
      // 1ba: athrow
      // 1bb: goto 282
      // 1be: athrow
      // 1bf: bipush 0
      // 1c0: istore 15
      // 1c2: aload 0
      // 1c3: getfield ua.kb [I
      // 1c6: iload 12
      // 1c8: iaload
      // 1c9: bipush -1
      // 1ca: ixor
      // 1cb: iload 15
      // 1cd: bipush -1
      // 1ce: ixor
      // 1cf: if_icmpge 23b
      // 1d2: bipush 0
      // 1d3: iload 17
      // 1d5: ifne 0c8
      // 1d8: istore 16
      // 1da: aload 0
      // 1db: getfield ua.R [I
      // 1de: iload 12
      // 1e0: iaload
      // 1e1: iload 16
      // 1e3: if_icmple 233
      // 1e6: aload 0
      // 1e7: getfield ua.gb [[B
      // 1ea: iload 12
      // 1ec: aaload
      // 1ed: iload 16
      // 1ef: aload 0
      // 1f0: getfield ua.kb [I
      // 1f3: iload 12
      // 1f5: iaload
      // 1f6: imul
      // 1f7: iload 15
      // 1f9: iadd
      // 1fa: aload 3
      // 1fb: iload 11
      // 1fd: iinc 11 1
      // 200: baload
      // 201: bastore
      // 202: aload 0
      // 203: getfield ua.gb [[B
      // 206: iload 12
      // 208: aaload
      // 209: iload 15
      // 20b: iload 16
      // 20d: aload 0
      // 20e: getfield ua.kb [I
      // 211: iload 12
      // 213: iaload
      // 214: imul
      // 215: iadd
      // 216: baload
      // 217: iload 17
      // 219: ifne 1cb
      // 21c: ifne 22b
      // 21f: aload 0
      // 220: getfield ua.Qb [Z
      // 223: iload 12
      // 225: bipush 1
      // 226: bastore
      // 227: goto 22b
      // 22a: athrow
      // 22b: iinc 16 1
      // 22e: iload 17
      // 230: ifeq 1da
      // 233: iinc 15 1
      // 236: iload 17
      // 238: ifeq 1c2
      // 23b: iload 17
      // 23d: ifeq 282
      // 240: bipush 0
      // 241: istore 15
      // 243: iload 14
      // 245: iload 15
      // 247: if_icmple 282
      // 24a: aload 0
      // 24b: getfield ua.gb [[B
      // 24e: iload 12
      // 250: aaload
      // 251: iload 15
      // 253: aload 3
      // 254: iload 11
      // 256: iinc 11 1
      // 259: baload
      // 25a: bastore
      // 25b: bipush 0
      // 25c: aload 0
      // 25d: getfield ua.gb [[B
      // 260: iload 12
      // 262: aaload
      // 263: iload 15
      // 265: baload
      // 266: iload 17
      // 268: ifne 0cd
      // 26b: if_icmpne 27a
      // 26e: aload 0
      // 26f: getfield ua.Qb [Z
      // 272: iload 12
      // 274: bipush 1
      // 275: bastore
      // 276: goto 27a
      // 279: athrow
      // 27a: iinc 15 1
      // 27d: iload 17
      // 27f: ifeq 243
      // 282: iinc 12 1
      // 285: iload 17
      // 287: ifeq 0c4
      // 28a: goto 2f9
      // 28d: astore 6
      // 28f: aload 6
      // 291: new java/lang/StringBuilder
      // 294: dup
      // 295: invokespecial java/lang/StringBuilder.<init> ()V
      // 298: getstatic ua.Yb [Ljava/lang/String;
      // 29b: bipush 70
      // 29d: aaload
      // 29e: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 2a1: iload 1
      // 2a2: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 2a5: bipush 44
      // 2a7: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 2aa: iload 2
      // 2ab: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 2ae: bipush 44
      // 2b0: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 2b3: aload 3
      // 2b4: ifnull 2c0
      // 2b7: getstatic ua.Yb [Ljava/lang/String;
      // 2ba: bipush 1
      // 2bb: aaload
      // 2bc: goto 2c5
      // 2bf: athrow
      // 2c0: getstatic ua.Yb [Ljava/lang/String;
      // 2c3: bipush 0
      // 2c4: aaload
      // 2c5: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 2c8: bipush 44
      // 2ca: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 2cd: iload 4
      // 2cf: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 2d2: bipush 44
      // 2d4: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 2d7: aload 5
      // 2d9: ifnull 2e5
      // 2dc: getstatic ua.Yb [Ljava/lang/String;
      // 2df: bipush 1
      // 2e0: aaload
      // 2e1: goto 2ea
      // 2e4: athrow
      // 2e5: getstatic ua.Yb [Ljava/lang/String;
      // 2e8: bipush 0
      // 2e9: aaload
      // 2ea: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 2ed: bipush 41
      // 2ef: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 2f2: invokevirtual java/lang/StringBuilder.toString ()Ljava/lang/String;
      // 2f5: invokestatic i.a (Ljava/lang/Throwable;Ljava/lang/String;)Lla;
      // 2f8: athrow
      // 2f9: return
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
         if (~var1 < -2) {
            int var13 = var2;

            try {
               if (var13 <= 1 || var4 == null) {
                  return;
               }
            } catch (RuntimeException var10) {
               throw var10;
            }

            this.nb = new DirectColorModel(32, 16711680, 65280, 255);
            int var12 = this.k * this.u;
            int var6 = 0;

            label49: {
               while (~var6 > ~var12) {
                  try {
                     this.rb[var6] = 0;
                     var6++;
                     if (var7) {
                        break label49;
                     }

                     if (var7) {
                        break;
                     }
                  } catch (RuntimeException var9) {
                     throw var9;
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
      } catch (RuntimeException var11) {
         RuntimeException var5 = var11;

         RuntimeException var10000;
         StringBuilder var10001;
         try {
            var10000 = var5;
            var10001 = new StringBuilder().append(Yb[73]).append(var1).append(',').append(var2).append(',').append(var3).append(',');
            if (var4 != null) {
               throw i.a(var5, var10001.append(Yb[1]).append(')').toString());
            }
         } catch (RuntimeException var8) {
            throw var8;
         }

         throw i.a(var10000, var10001.append(Yb[0]).append(')').toString());
      }
   }

   private static char[] z(String var0) {
      char[] var10000 = var0.toCharArray();
      if (var10000.length < 2) {
         var10000[0] = (char)(var10000[0] ^ '\t');
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
