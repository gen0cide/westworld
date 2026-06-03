import java.io.DataInputStream;
import java.net.URL;

final class mb {
   static int c;
   private static int b = 0;
   static int[] i = new int[]{
      0,
      1,
      3,
      7,
      15,
      31,
      63,
      127,
      255,
      511,
      1023,
      2047,
      4095,
      8191,
      16383,
      32767,
      65535,
      131071,
      262143,
      524287,
      1048575,
      2097151,
      4194303,
      8388607,
      16777215,
      33554431,
      67108863,
      134217727,
      268435455,
      536870911,
      1073741823,
      Integer.MAX_VALUE,
      -1
   };
   static int[] a;
   static int d;
   static int e;
   static int[] k;
   static int l = 0;
   static String[] g;
   static int h;
   static int f;
   static int j;
   private static final String[] z = new String[]{
      z(z("U<HS\u001e")),
      z(z("U<HV\u001e")),
      z(z("U<HP\u001e")),
      z(z("V+\ny")),
      z(z("CpH;K")),
      z(z("a1\u00135B]2\n5")),
      z(z("\u0018*\u0003yZK~\u001fzC\u0002~")),
      z(z("\u0002~")),
      z(z("\u0018)\u000ff^]-FaY\u0018*\u0014tR]~\u0011|BP~\u001fzC\u0016")),
      z(z("U<HQ\u001e")),
      z(z("\u0018\"F")),
      z(z("[2\u000fpXL;\u0014gYJp\u0011f\t[c")),
      z(z("\u001djV")),
      z(z("\u001dlU")),
      z(z("\u001e;[")),
      z(z("\u001dm\u0007")),
      z(z("\u001dlP")),
      z(z("\u001e(W(")),
      z(z("\u001e(T(")),
      z(z("\u001e+[")),
      z(z("s~&b^Q\u001eN")),
      z(z("x=\u001ftv")),
      z(z("\u00183\u000fyZQ1\b5vO6\u000fU\u001e")),
      z(z("U<HW\u001e")),
      z(z("x9\u0014pv"))
   };

   static final synchronized byte[] a(int var0, byte var1) {
      boolean var4 = client.vh;

      try {
         j++;
         if (var0 == 100 && ~n.b < -1) {
            byte[] var8 = ob.j[--n.b];
            ob.j[n.b] = null;
            return var8;
         }

         if (5000 == var0 && ~s.d < -1) {
            byte[] var7 = e.kb[--s.d];
            e.kb[s.d] = null;
            return var7;
         }

         if (var1 > -97) {
            return (byte[])null;
         }

         if (-30001 == ~var0 && ~b < -1) {
            byte[] var6 = ca.tb[--b];
            ca.tb[b] = null;
            return var6;
         }

         if (null != t.n) {
            int var2 = 0;

            while (~e.wb.length < ~var2) {
               int var10000 = ~e.wb[var2];
               if (var4) {
                  return new byte[var10000];
               }

               if (var10000 == ~var0 && 0 < v.g[var2]) {
                  byte[] var3 = t.n[var2][--v.g[var2]];
                  t.n[var2][v.g[var2]] = null;
                  return var3;
               }

               var2++;
               if (var4) {
                  break;
               }
            }
         }

         return new byte[var0];
      } catch (RuntimeException var5) {
         throw i.a(var5, z[1] + var0 + ',' + var1 + ')');
      }
   }

   static final void a(int var0, long var1) {
      try {
         h++;
         if (var0 != 11200) {
            g = (String[])null;
         }

         if (0L < var1) {
            if (var1 % 10L == 0L) {
               u.a(var0 + -11200, -1L + var1);
               u.a(0, 1L);
               if (!client.vh) {
                  return;
               }
            }

            u.a(0, var1);
         }
      } catch (RuntimeException var4) {
         throw i.a(var4, z[0] + var0 + ',' + var1 + ')');
      }
   }

   static final String a(int var0, int var1) {
      boolean var4 = client.vh;

      try {
         f++;
         String var2 = "" + var0;
         int var3 = -3 + var2.length();

         while (true) {
            if (-1 > ~var3) {
               var2 = var2.substring(0, var3) + "," + var2.substring(var3);
               var3 -= 3;
               if (var4) {
                  break;
               }

               if (!var4) {
                  continue;
               }
            }

            if (var1 != 131071) {
               a((byte[])null, -74, 53);
            }
            break;
         }

         if (-9 > ~var2.length()) {
            var2 = z[24] + var2.substring(0, var2.length() + -8) + z[22] + var2 + ")";
            if (!var4) {
               return var2;
            }
         }

         if (~var2.length() < -5) {
            var2 = z[21] + var2.substring(0, -4 + var2.length()) + z[20] + var2 + ")";
         }

         return var2;
      } catch (RuntimeException var5) {
         throw i.a(var5, z[23] + var0 + ',' + var1 + ')');
      }
   }

   static final String a(String var0, String var1, boolean var2, int var3) {
      boolean var5 = client.vh;

      try {
         d++;
         if (!var2) {
            l = 90;
         }

         int var4 = var3;
         if (0 != var4 || var5) {
            if (var4 == 1) {
               if (var1 != null && var1.length() != 0) {
                  return var1 + z[6] + var0;
               }

               return var0;
            }

            if (var4 == 2 && !var5) {
               if (null != var1 && -1 != ~var1.length()) {
                  return z[5] + var1 + z[7] + var0;
               }

               return var0;
            }

            if (var4 == 3 && !var5) {
               if (null != var1 && ~var1.length() != -1) {
                  return var1 + z[7] + var0;
               }

               return var0;
            }

            if (~var4 == -5 && !var5) {
               if (null != var1 && ~var1.length() != -1) {
                  return var1 + z[7] + var0;
               }

               return var0;
            }

            if (~var4 == -6) {
               return var0;
            }

            if (6 == var4 && !var5) {
               return var1 + z[8];
            }

            if (-8 != ~var4) {
               return "";
            }

            if (!var5) {
               if (null != var1 && ~var1.length() != -1) {
                  return var1 + z[7] + var0;
               }

               return var0;
            }
         }

         return var1 != null && -1 != ~var1.length() ? var1 + z[7] + var0 : var0;
      } catch (RuntimeException var6) {
         throw i.a(var6, z[9] + (var0 != null ? z[4] : z[3]) + ',' + (var1 != null ? z[4] : z[3]) + ',' + var2 + ',' + var3 + ')');
      }
   }

   static final void a(int var0, Throwable var1, String var2) {
      boolean var6 = client.vh;

      try {
         e++;

         try {
            String var3 = "";
            if (var0 != 2097151) {
               a((String)null, (String)null, true, 27);
            }

            if (var1 != null) {
               var3 = gb.a(false, var1);
            }

            if (null != var2) {
               if (var1 != null) {
                  var3 = var3 + z[10];
               }

               var3 = var3 + var2;
            }

            n.a((byte)-93, var3);
            var3 = jb.a(true, z[15], ":", var3);
            var3 = jb.a(true, z[12], "@", var3);
            var3 = jb.a(true, z[16], "&", var3);
            var3 = jb.a(true, z[13], "#", var3);
            if (null == l.b) {
               return;
            }

            g var4 = pa.b
               .a((byte)74, new URL(l.b.getCodeBase(), z[11] + db.d + z[19] + (ka.a != null ? ka.a : "" + pa.h) + z[17] + c.q + z[18] + c.k + z[14] + var3));

            while (~var4.b == -1) {
               a(11200, 1L);
               if (var6) {
                  return;
               }

               if (var6) {
                  break;
               }
            }

            if (-2 == ~var4.b) {
               DataInputStream var5 = (DataInputStream)var4.d;
               var5.read();
               var5.close();
            }
         } catch (Exception var7) {
         }
      } catch (RuntimeException var8) {
         throw var8;
      }
   }

   static final int a(byte[] var0, int var1, int var2) {
      try {
         if (var2 != 0) {
            return 6;
         }

         c++;
         return w.a(var1, -49, var0, 0);
      } catch (RuntimeException var4) {
         throw i.a(var4, z[2] + (var0 != null ? z[4] : z[3]) + 44 + var1 + 44 + var2 + 41);
      }
   }

   private static char[] z(String var0) {
      char[] var10000 = var0.toCharArray();
      if (var10000.length < 2) {
         var10000[0] = (char)(var10000[0] ^ 54);
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
               var10005 = 56;
               break;
            case 1:
               var10005 = 94;
               break;
            case 2:
               var10005 = 102;
               break;
            case 3:
               var10005 = 21;
               break;
            default:
               var10005 = 54;
         }

         var10001[var1] = (char)(var10004 ^ var10005);
      }

      return new String(var10001).intern();
   }
}
