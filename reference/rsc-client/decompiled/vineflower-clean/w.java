final class w {
   int l;
   int r;
   int m;
   int p;
   int b;
   int j;
   boolean c = false;
   int q;
   int i;
   int e;
   static int n;
   static int d;
   int k;
   int t;
   int s;
   static int[] g;
   int f;
   int h;
   static int a;
   int u;
   ca o;
   private static final String[] z = new String[]{z(z("\u0011S8B")), z(z("\u001dSTDf")), z(z("\b\b\u0016\u0006")), z(z("\u0011S9B")), z(z("\u0011S;B"))};

   // $VF: Irreducible bytecode was duplicated to produce valid code
   static final String a(CharSequence var0, byte var1) {
      boolean var9 = client.vh;

      try {
         a++;
         if (var1 <= 47) {
            return (String)null;
         }

         if (var0 == null) {
            return null;
         }

         int var2 = 0;
         int var3 = var0.length();

         int var11;
         label120: {
            label119:
            while (true) {
               if (~var2 > ~var3) {
                  var11 = db.a(32, var0.charAt(var2));
                  if (!var9) {
                     if (var11 != 0) {
                        var2++;
                        if (!var9) {
                           continue;
                        }
                     }

                     var11 = var3;
                  }
               } else {
                  var11 = var3;
               }

               while (true) {
                  if (var11 <= var2) {
                     break label119;
                  }

                  var11 = db.a(32, var0.charAt(-1 + var3));
                  if (var9) {
                     break label120;
                  }

                  if (var11 == 0) {
                     break label119;
                  }

                  var3--;
                  if (var9) {
                     break label119;
                  }

                  var11 = var3;
               }
            }

            var11 = var3 - var2;
         }

         int var4 = var11;
         if (1 <= var4 && -13 <= ~var4) {
            StringBuilder var5 = new StringBuilder(var4);
            int var6 = var2;

            while (true) {
               if (~var3 < ~var6) {
                  char var7 = var0.charAt(var6);
                  var12 = f.a(var7, 0);
                  if (var9) {
                     break;
                  }

                  if (var12 != 0 || var9) {
                     char var8 = ac.a(var7, -194);
                     if (~var8 != -1) {
                        var5.append(var8);
                     }
                  }

                  var6++;
                  if (!var9) {
                     continue;
                  }
               }

               var12 = 0;
               break;
            }

            return var12 != var5.length() ? var5.toString() : null;
         } else {
            return null;
         }
      } catch (RuntimeException var10) {
         throw i.a(var10, z[4] + (var0 != null ? z[1] : z[2]) + ',' + var1 + ')');
      }
   }

   static final int a(int var0, int var1, byte[] var2, int var3) {
      boolean var6 = client.vh;

      try {
         n++;
         int var4 = -1;
         int var5 = var3;

         while (true) {
            if (var0 > var5) {
               var4 = wb.q[(var2[var5] ^ var4) & 0xFF] ^ var4 >>> 1628768296;
               var5++;
               if (var6) {
                  break;
               }

               if (!var6) {
                  continue;
               }
            }

            var5 = 123 / ((var1 - 23) / 63);
            var4 = ~var4;
            break;
         }

         return var4;
      } catch (RuntimeException var7) {
         throw i.a(var7, z[3] + var0 + 44 + var1 + 44 + (var2 != null ? z[1] : z[2]) + 44 + var3 + 41);
      }
   }

   static final int a(byte[] var0, int var1, int var2) {
      try {
         d++;
         if (var1 != -1) {
            return 71;
         }

         int var3 = 256 * nb.a(255, var0[var2]) + nb.a(255, var0[1 + var2]);
         if (-32768 > ~var3) {
            var3 -= 65536;
         }

         return var3;
      } catch (RuntimeException var4) {
         throw i.a(var4, z[0] + (var0 != null ? z[1] : z[2]) + 44 + var1 + 44 + var2 + 41);
      }
   }

   w() {
      this.p = -1;
      this.f = 0;
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
               var10005 = 102;
               break;
            case 1:
               var10005 = 125;
               break;
            case 2:
               var10005 = 122;
               break;
            case 3:
               var10005 = 106;
               break;
            default:
               var10005 = 27;
         }

         var10001[var1] = (char)(var10004 ^ var10005);
      }

      return new String(var10001).intern();
   }
}
