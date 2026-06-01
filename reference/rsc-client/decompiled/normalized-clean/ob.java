import java.io.EOFException;
import java.io.IOException;

final class ob {
   static int e;
   static int d;
   static int g;
   private int c;
   static byte[][] j = new byte[1000][];
   private int a = 65000;
   private nb b;
   private nb f = null;
   static int i;
   static int[] h;
   private static final String[] z = new String[]{
      z(z("\u0011\u0001]5W")),
      z(z("\u0011\u0001]\u0000\u0010-\u0017\u0001\u001d\u0011\u0019KZ")),
      z(z("\u0011\u0001]6W")),
      z(z("\u0005M]Z\u0002")),
      z(z("\u0010\u0016\u001f\u0018")),
      z(z("\u0011\u0001]7W")),
      z(z("\u0011\u0001]H\u0016\u0010\n\u0007JW"))
   };

   final boolean a(int var1, int var2, int var3, byte[] var4) {
      try {
         int var5 = 94 % ((var3 - -61) / 35);
         d++;
         synchronized (this.f) {
            if (-1 >= ~var2 && this.a >= var2) {
               boolean var7 = this.a(var1, var4, true, var2, 4);
               if (!var7) {
                  var7 = this.a(var1, var4, false, var2, 4);
               }

               return var7;
            } else {
               throw new IllegalArgumentException();
            }
         }
      } catch (RuntimeException var10) {
         throw i.a(var10, z[2] + var1 + ',' + var2 + ',' + var3 + ',' + (var4 != null ? z[3] : z[4]) + ')');
      }
   }

   final byte[] a(int var1, int var2) {
      boolean var18 = client.vh;

      try {
         i++;
         synchronized (this.f) {
            byte[] var24;
            try {
               if (~this.b.a((byte)-111) > ~(6 + var2 * 6)) {
                  return null;
               }

               this.b.a(6 * var2, 12);
               this.b.a(true, 6, 0, la.c);
               int var4 = (0xFF00 & la.c[1] << 8) + ((255 & la.c[0]) << 16) + (255 & la.c[2]);
               int var5 = (la.c[4] << 8 & 0xFF00) + ((la.c[3] & 255) << 16) - -(255 & la.c[5]);
               if (var4 < 0 || ~var4 < ~this.a) {
                  return null;
               }

               if (~var5 >= -1 || ~(this.f.a((byte)-111) / 520L) > ~var5) {
                  return null;
               }

               if (var1 != 9395) {
                  return (byte[])null;
               }

               byte[] var6 = new byte[var4];
               int var7 = 0;
               int var8 = 0;

               while (var4 > var7) {
                  if (var5 == 0) {
                     return null;
                  }

                  int var9;
                  int var10;
                  int var11;
                  int var12;
                  int var13;
                  byte var14;
                  label194: {
                     this.f.a(var5 * 520, 107);
                     var9 = -var7 + var4;
                     if (var2 > 65535) {
                        if (~var9 < -511) {
                           var9 = 510;
                        }

                        var14 = 10;
                        this.f.a(true, var9 + var14, 0, la.c);
                        var10 = (la.c[3] & 255) + (la.c[0] << 24 & 0xFF000000) - -(0xFF0000 & la.c[1] << 16) + (0xFF00 & la.c[2] << 8);
                        var13 = la.c[9] & 255;
                        var11 = (la.c[5] & 255) + ((255 & la.c[4]) << 8);
                        var12 = (la.c[8] & 255) + (0xFF0000 & la.c[6] << 16) + (la.c[7] << 8 & 0xFF00);
                        if (!var18) {
                           break label194;
                        }
                     }

                     if (~var9 < -513) {
                        var9 = 512;
                     }

                     var14 = 8;
                     this.f.a(true, var14 + var9, 0, la.c);
                     var13 = la.c[7] & 255;
                     var10 = (la.c[0] << 8 & 0xFF00) + (255 & la.c[1]);
                     var11 = (la.c[2] << 8 & 0xFF00) + (la.c[3] & 255);
                     var12 = ((255 & la.c[4]) << 16) - -((la.c[5] & 255) << 8) - -(255 & la.c[6]);
                  }

                  if (~var2 == ~var10 && var8 == var11 && ~var13 == ~this.c) {
                     if (var12 >= 0 && ~(this.f.a((byte)-111) / 520L) <= ~var12) {
                        int var15 = var14 - -var9;
                        var5 = var12;
                        int var16 = var14;

                        while (true) {
                           if (var15 > var16) {
                              var6[var7++] = la.c[var16];
                              var16++;
                              if (var18) {
                                 break;
                              }

                              if (!var18) {
                                 continue;
                              }
                           }

                           var8++;
                           break;
                        }

                        if (var18) {
                           break;
                        }
                        continue;
                     }

                     return null;
                  }

                  return null;
               }

               var24 = var6;
            } catch (IOException var19) {
               return null;
            }

            return var24;
         }
      } catch (RuntimeException var21) {
         throw i.a(var21, z[0] + var1 + ',' + var2 + ')');
      }
   }

   private final boolean a(int var1, byte[] var2, boolean var3, int var4, int var5) {
      boolean var16 = client.vh;

      try {
         e++;
         synchronized (this.f) {
            boolean var10000;
            try {
               int var7;
               if (var3) {
                  if (~(6 + 6 * var1) < ~this.b.a((byte)-111)) {
                     return false;
                  }

                  this.b.a(var1 * 6, -124);
                  this.b.a(true, 6, 0, la.c);
                  var7 = (255 & la.c[5]) + (la.c[3] << 16 & 0xFF0000) - -((la.c[4] & 255) << 8);
                  if (var7 <= 0 || this.f.a((byte)-111) / 520L < var7) {
                     return false;
                  }
               } else {
                  var7 = (int)((this.f.a((byte)-111) - -519L) / 520L);
                  if (0 == var7) {
                     var7 = 1;
                  }
               }

               la.c[0] = (byte)(var4 >> 16);
               la.c[2] = (byte)var4;
               la.c[3] = (byte)(var7 >> 16);
               la.c[var5] = (byte)(var7 >> 8);
               la.c[1] = (byte)(var4 >> 8);
               la.c[5] = (byte)var7;
               this.b.a(6 * var1, 31);
               this.b.a(6, -102, 0, la.c);
               int var8 = 0;
               int var9 = 0;

               while (~var8 > ~var4) {
                  int var10 = 0;
                  if (var16) {
                     return var3;
                  }

                  if (var3) {
                     int var11;
                     int var12;
                     int var13;
                     label207: {
                        this.f.a(520 * var7, var5 ^ 17);
                        if (var1 <= 65535) {
                           try {
                              this.f.a(true, 8, 0, la.c);
                           } catch (EOFException var18) {
                              if (!var16) {
                                 break;
                              }
                           }

                           var11 = (0xFF00 & la.c[0] << 8) - -(la.c[1] & 255);
                           var10 = (la.c[6] & 255) + (0xFF0000 & la.c[4] << 16) - -((255 & la.c[5]) << 8);
                           var13 = 255 & la.c[7];
                           var12 = (la.c[3] & 255) + (la.c[2] << 8 & 0xFF00);
                           if (!var16) {
                              break label207;
                           }
                        }

                        try {
                           this.f.a(true, 10, 0, la.c);
                        } catch (EOFException var17) {
                           if (!var16) {
                              break;
                           }
                        }

                        var12 = (255 & la.c[5]) + (la.c[4] << 8 & 0xFF00);
                        var10 = (la.c[6] << 16 & 0xFF0000) - -((la.c[7] & 255) << 8) + (la.c[8] & 255);
                        var11 = (la.c[0] << 24 & 0xFF000000)
                           + ((la.c[1] << 16 & 0xFF0000) - -((255 & la.c[2]) << 8))
                           - -(255 & la.c[3]);
                        var13 = 255 & la.c[9];
                     }

                     if (~var1 != ~var11 || var12 != var9 || var13 != this.c) {
                        return false;
                     }

                     if (0 > var10 || ~var10 < ~(this.f.a((byte)-111) / 520L)) {
                        return false;
                     }
                  }

                  if (~var10 == -1) {
                     var3 = false;
                     var10 = (int)((this.f.a((byte)-111) - -519L) / 520L);
                     if (0 == var10) {
                        var10++;
                     }

                     if (var10 == var7) {
                        var10++;
                     }
                  }

                  if (512 >= var4 - var8) {
                     var10 = 0;
                  }

                  label208: {
                     if (-65536 > ~var1) {
                        la.c[0] = (byte)(var1 >> 24);
                        la.c[5] = (byte)var9;
                        la.c[2] = (byte)(var1 >> 8);
                        la.c[4] = (byte)(var9 >> 8);
                        la.c[7] = (byte)(var10 >> 8);
                        la.c[1] = (byte)(var1 >> 16);
                        la.c[8] = (byte)var10;
                        la.c[9] = (byte)this.c;
                        la.c[3] = (byte)var1;
                        la.c[6] = (byte)(var10 >> 16);
                        this.f.a(520 * var7, var5 ^ 33);
                        this.f.a(10, -111, 0, la.c);
                        int var22 = var4 + -var8;
                        if (-511 > ~var22) {
                           var22 = 510;
                        }

                        this.f.a(var22, var5 + -119, var8, var2);
                        var8 += var22;
                        if (!var16) {
                           break label208;
                        }
                     }

                     la.c[4] = (byte)(var10 >> 16);
                     la.c[0] = (byte)(var1 >> 8);
                     la.c[7] = (byte)this.c;
                     la.c[6] = (byte)var10;
                     la.c[3] = (byte)var9;
                     la.c[1] = (byte)var1;
                     la.c[2] = (byte)(var9 >> 8);
                     la.c[5] = (byte)(var10 >> 8);
                     this.f.a(var7 * 520, var5 ^ 127);
                     this.f.a(8, -107, 0, la.c);
                     int var23 = var4 - var8;
                     if (~var23 < -513) {
                        var23 = 512;
                     }

                     this.f.a(var23, var5 + -125, var8, var2);
                     var8 += var23;
                  }

                  var7 = var10;
                  var9++;
                  if (var16) {
                     break;
                  }
               }

               var10000 = true;
            } catch (IOException var19) {
               return false;
            }

            return var10000;
         }
      } catch (RuntimeException var21) {
         throw i.a(var21, z[5] + var1 + ',' + (var2 != null ? z[3] : z[4]) + ',' + var3 + ',' + var4 + ',' + var5 + ')');
      }
   }

   @Override
   public final String toString() {
      try {
         g++;
         return "" + this.c;
      } catch (RuntimeException var2) {
         throw i.a(var2, z[1]);
      }
   }

   ob(int var1, nb var2, nb var3, int var4) {
      this.b = null;

      try {
         this.a = var4;
         this.c = var1;
         this.b = var3;
         this.f = var2;
      } catch (RuntimeException var6) {
         throw i.a(var6, z[6] + var1 + ',' + (var2 != null ? z[3] : z[4]) + ',' + (var3 != null ? z[3] : z[4]) + ',' + var4 + ')');
      }
   }

   private static char[] z(String var0) {
      char[] var10000 = var0.toCharArray();
      if (var10000.length < 2) {
         var10000[0] = (char)(var10000[0] ^ 127);
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
               var10005 = 126;
               break;
            case 1:
               var10005 = 99;
               break;
            case 2:
               var10005 = 115;
               break;
            case 3:
               var10005 = 116;
               break;
            default:
               var10005 = 127;
         }

         var10001[var1] = (char)(var10004 ^ var10005);
      }

      return new String(var10001).intern();
   }
}
