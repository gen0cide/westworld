import java.io.IOException;

final class f {
   static int[] f;
   static int d;
   static int a;
   static String[] e;
   static String[] c = new String[]{
      z(z("c\fh!cV@h.dV\u0012-4xV@c5}Q\u0005\u007f`\u007fU@d4u^\u0013-4\u007f\u0013\u0004h0\u007f@\ty")), z(z("R\u000ei``A\u0005~30V\u000ey%b"))
   };
   static i b = new i(z(z("d)]")), 2);
   private static final String[] z = new String[]{z(z("UNLh")), z(z("]\u0015a,")), z(z("UNOh")), z(z("HN#nm"))};

   // $VF: Irreducible bytecode was duplicated to produce valid code
   static final boolean a(char var0, int var1) {
      boolean var5 = client.vh;

      try {
         a++;
         if (Character.isISOControl(var0)) {
            return false;
         }

         if (v.a(var0, 115)) {
            return true;
         }

         char[] var2 = ga.a;
         int var3 = var1;

         int var10000;
         int var10001;
         while (true) {
            if (~var3 > ~var2.length) {
               char var4 = var2[var3];
               var10000 = ~var0;
               var10001 = ~var4;
               if (var5) {
                  break;
               }

               if (var10000 == var10001) {
                  return true;
               }

               var3++;
               if (!var5) {
                  continue;
               }
            }

            var2 = ac.I;
            var3 = 0;
            var10000 = var2.length;
            var10001 = var3;
            break;
         }

         while (var10000 > var10001) {
            char var7 = var2[var3];
            if (var5) {
               return (boolean)var7;
            }

            if (var7 == var0) {
               return true;
            }

            var3++;
            if (var5) {
               break;
            }

            var10000 = var2.length;
            var10001 = var3;
         }

         return false;
      } catch (RuntimeException var6) {
         throw i.a(var6, z[0] + var0 + ',' + var1 + ')');
      }
   }

   static final void a(int var0, tb var1) {
      boolean var5 = client.vh;

      try {
         d++;
         byte[] var2 = new byte[24];
         if (b.q != null) {
            try {
               b.q.a(0L, var0 + -22592);
               b.q.a((byte)-123, var2);
               int var3 = 0;

               byte var10000;
               int var10001;
               label72: {
                  while (var3 < 24) {
                     var10000 = -1;
                     var10001 = ~var2[var3];
                     if (var5) {
                        break label72;
                     }

                     if (-1 != var10001 && !var5) {
                        break;
                     }

                     var3++;
                     if (var5) {
                        break;
                     }
                  }

                  var10000 = -25;
                  var10001 = ~var3;
               }

               if (var10000 >= var10001) {
                  throw new IOException();
               }
            } catch (Exception var6) {
               int var4 = 0;

               while (-25 < ~var4) {
                  var2[var4] = -1;
                  var4++;
                  if (var5) {
                     return;
                  }

                  if (var5) {
                     break;
                  }
               }
            }
         }

         if (var0 == 22607) {
            var1.a(0, -126, 24, var2);
         }
      } catch (RuntimeException var7) {
         throw i.a(var7, z[2] + var0 + ',' + (var1 != null ? z[3] : z[1]) + ')');
      }
   }

   private static char[] z(String var0) {
      char[] var10000 = var0.toCharArray();
      if (var10000.length < 2) {
         var10000[0] = (char)(var10000[0] ^ 16);
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
               var10005 = 51;
               break;
            case 1:
               var10005 = 96;
               break;
            case 2:
               var10005 = 13;
               break;
            case 3:
               var10005 = 64;
               break;
            default:
               var10005 = 16;
         }

         var10001[var1] = (char)(var10004 ^ var10005);
      }

      return new String(var10001).intern();
   }
}
