import java.awt.Color;
import java.awt.Font;
import java.awt.FontMetrics;
import java.awt.Graphics;
import java.awt.Image;
import java.awt.image.PixelGrabber;
import java.math.BigInteger;

final class s {
   static int b;
   static String[] e = new String[]{
      z(z(".;\u0016f\u001cK;\u0017n\f\u000e'Bl\bK<\u0016f\u0003\u0018u\u0016lN\u00043\u0004f\u001cK4\fgN\u001b'\u0007p\u001dK0\fw\u000b\u0019"))
   };
   static String[] f;
   static BigInteger c = new BigInteger(z(z("ZeR3_")), 16);
   static nb a = null;
   static int d = 0;
   private static final String[] z = new String[]{z(z("\u0018{#+")), z(z("\u0010{L-\u0013")), z(z("\u0005 \u000eo"))};

   static final boolean a(int var0, Font var1, int var2, int var3, e var4, char var5, FontMetrics var6, boolean var7) {
      boolean var24 = client.vh;

      try {
         b++;
         int var8 = var6.charWidth(var5);
         int var9 = var8;
         if (var7) {
            try {
               if (var5 == 'f'
                  || var5 == 't'
                  || 'w' == var5
                  || ~var5 == -119
                  || var5 == 'k'
                  || ~var5 == -121
                  || 'y' == var5
                  || var5 == 'A'
                  || ~var5 == -87
                  || ~var5 == -88) {
                  var8++;
               }

               if (~var5 == -48) {
                  var7 = false;
               }
            } catch (Exception var26) {
            }
         }

         int var10 = var6.getMaxAscent();
         int var11 = var6.getMaxAscent() - -var6.getMaxDescent();
         int var12 = var6.getHeight();
         Image var13 = var4.createImage(var8, var11);
         if (null == var13) {
            return false;
         }

         Graphics var14 = var13.getGraphics();
         var14.setColor(Color.black);
         var14.fillRect(0, 0, var8, var11);
         var14.setColor(Color.white);
         var14.setFont(var1);
         var14.drawString(var5 + "", 0, var10);
         if (var7) {
            var14.drawString(var5 + "", 1, var10);
         }

         int[] var15 = new int[var11 * var8];
         PixelGrabber var16 = new PixelGrabber(var13, 0, 0, var8, var11, var15, 0, var8);

         try {
            var16.grabPixels();
         } catch (InterruptedException var25) {
            return false;
         }

         var13.flush();
         Object var28 = null;
         int var17 = 0;
         int var18 = 0;
         int var19 = var8;
         int var20 = var11;
         int var21 = 0;

         int var40;
         label289:
         while (true) {
            var40 = var21;
            int var10001 = var11;

            label286:
            while (var40 < var10001) {
               var40 = 0;
               if (var24) {
                  break label289;
               }

               int var22 = 0;

               while (~var8 < ~var22) {
                  int var23 = var15[var22 - -(var21 * var8)];
                  var40 = 0;
                  var10001 = 16777215 & var23;
                  if (var24) {
                     continue label286;
                  }

                  if (0 != var10001) {
                     var18 = var21;
                     if (!var24) {
                        break label286;
                     }
                  }

                  var22++;
                  if (var24) {
                     break;
                  }
               }

               var21++;
               if (!var24) {
                  continue label289;
               }
               break;
            }

            var40 = 0;
            break;
         }

         var21 = var40;

         label264:
         while (true) {
            var40 = var21;
            int var49 = var8;

            label261:
            while (var40 < var49) {
               var40 = 0;
               if (var24) {
                  break label264;
               }

               int var32 = 0;

               while (var11 > var32) {
                  int var36 = var15[var21 - -(var32 * var8)];
                  var40 = 0;
                  var49 = var36 & 16777215;
                  if (var24) {
                     continue label261;
                  }

                  if (0 != var49) {
                     var17 = var21;
                     if (!var24) {
                        break label261;
                     }
                  }

                  var32++;
                  if (var24) {
                     break;
                  }
               }

               var21++;
               if (!var24) {
                  continue label264;
               }
               break;
            }

            var21 = var11 - 1;
            var40 = var3;
            break;
         }

         if (var40 >= -86) {
            a(-60, (Font)null, 49, -85, (e)null, '\ufff8', (FontMetrics)null, true);
         }

         label239:
         while (true) {
            var40 = 0;
            int var50 = var21;

            label236:
            while (var40 <= var50) {
               var40 = 0;
               if (var24) {
                  break label239;
               }

               int var33 = 0;

               while (~var33 > ~var8) {
                  int var37 = var15[var33 + var21 * var8];
                  var40 = ~(var37 & 16777215);
                  var50 = -1;
                  if (var24) {
                     continue label236;
                  }

                  if (var40 != -1) {
                     var20 = 1 + var21;
                     if (!var24) {
                        break label236;
                     }
                  }

                  var33++;
                  if (var24) {
                     break;
                  }
               }

               var21--;
               if (!var24) {
                  continue label239;
               }
               break;
            }

            var40 = -1 + var8;
            break;
         }

         var21 = var40;

         label214:
         while (true) {
            byte var45 = -1;
            int var51 = ~var21;

            label211:
            while (var45 >= var51) {
               var40 = 0;
               if (var24) {
                  break label214;
               }

               int var34 = 0;

               while (~var34 > ~var11) {
                  int var38 = var15[var21 + var34 * var8];
                  var45 = -1;
                  var51 = ~(16777215 & var38);
                  if (var24) {
                     continue label211;
                  }

                  if (-1 != var51) {
                     var19 = var21 + 1;
                     if (!var24) {
                        break label211;
                     }
                  }

                  var34++;
                  if (var24) {
                     break;
                  }
               }

               var21--;
               if (!var24) {
                  continue label214;
               }
               break;
            }

            qb.k[0 + 9 * var2] = (byte)(b.c / 16384);
            qb.k[var2 * 9 - -1] = (byte)ib.a(b.c / 128, 127);
            qb.k[2 + 9 * var2] = (byte)ib.a(b.c, 127);
            qb.k[3 + var2 * 9] = (byte)(-var17 + var19);
            qb.k[var2 * 9 - -4] = (byte)(var20 + -var18);
            qb.k[9 * var2 - -5] = (byte)var17;
            qb.k[6 + var2 * 9] = (byte)(var10 + -var18);
            qb.k[7 + 9 * var2] = (byte)var9;
            qb.k[var2 * 9 - -8] = (byte)var12;
            var40 = var18;
            break;
         }

         var21 = var40;

         label189:
         while (true) {
            var40 = var20;
            int var52 = var21;

            label186:
            while (var40 > var52) {
               if (var24) {
                  return (boolean)var17;
               }

               int var35 = var17;

               while (~var35 > ~var19) {
                  int var39 = 0xFF & var15[var8 * var21 + var35];
                  var40 = 30;
                  var52 = var39;
                  if (var24) {
                     continue label186;
                  }

                  if (30 < var39 && ~var39 > -231) {
                     fb.k[var0] = true;
                  }

                  qb.k[b.c++] = (byte)var39;
                  var35++;
                  if (var24) {
                     break;
                  }
               }

               var21++;
               if (!var24) {
                  continue label189;
               }
               break;
            }

            return true;
         }
      } catch (RuntimeException var27) {
         throw i.a(
            var27,
            z[0]
               + var0
               + ','
               + (var1 != null ? z[1] : z[2])
               + ','
               + var2
               + ','
               + var3
               + ','
               + (var4 != null ? z[1] : z[2])
               + ','
               + var5
               + ','
               + (var6 != null ? z[1] : z[2])
               + ','
               + var7
               + ')'
         );
      }
   }

   private static char[] z(String var0) {
      char[] var10000 = var0.toCharArray();
      if (var10000.length < 2) {
         var10000[0] = (char)(var10000[0] ^ 110);
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
               var10005 = 107;
               break;
            case 1:
               var10005 = 85;
               break;
            case 2:
               var10005 = 98;
               break;
            case 3:
               var10005 = 3;
               break;
            default:
               var10005 = 110;
         }

         var10001[var1] = (char)(var10004 ^ var10005);
      }

      return new String(var10001).intern();
   }
}
