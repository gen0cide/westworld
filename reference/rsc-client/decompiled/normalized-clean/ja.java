import java.math.BigInteger;

final class ja extends tb {
   static BigInteger K = new BigInteger(
      z(
         z(
            "SS\"\u0011v\u0004\u0005)E#\t\u0005-\u0011w\b\u0007yBt\t\u0002}Bs\u0004S#\u0016uR\u0003\u007f\u0016\u007fR\u0006-@%\u0003Q}\u0012q\u0006\u0000+\u0017$R\n,\u0015#VS)\u0013~T\u000bx\u0010\u007fU\u0003-@#VQ.\u0017 V\u0006,\u001du\u0000\u0007*\u0016u\u0004\u0007/\u0011v\u0005\u0002#\u0016 \u0004\u0005+\u0014$\u0000Vz\u0017~\u0001\u0002/\u0013 \u0005\u0003y\u001cq\u0002T\"F$UWz\u0012s\u0003T)\u0015 T\u0000/\u001c'\u0001\u0002}Bs\u0002\u0001\"Fu\u0000\u0000(\u0010'TV(\u0011\u007f\u0001\u0001xFp\u0000\u0004#@u\u0001\u0004~@\"\u0004\u0003#\u0012w\u0001\u0001(\u0010'U\u0002/\u0013 SV\"E%R\u0005y\u0014%\u0001\u0001y\u0017v\u0003\u000b(Et\u0006\u0000+\u0010\"S\n.\u0015~\u0003W+E\u007f\u0005\u0007.\u0011%\u0000\u0003yA#\b\u0002+\u0010r\u0000W\"\u0013rRP\"Fr\u0004\u0003}\u0010p\u0004T/\u0014s\u0007"
         )
      ),
      16
   );
   static int I;
   static int[] N = new int[100];
   static int J;
   static String[] L;
   static int G;
   private int M;
   static int H;
   private static final String[] O = new String[]{z(z("ZS5`n")), z(z("ZS5en")), z(z("ZS5gn")), z(z("ZS5fn"))};

   final int k(int var1) {
      try {
         if (var1 != -31874) {
            return 40;
         }

         H++;
         return this.M;
      } catch (RuntimeException var3) {
         throw i.a(var3, O[1] + var1 + 41);
      }
   }

   final void i(int var1) {
      try {
         this.M = 8 * this.w;
         G++;
         if (var1 != -2231) {
            this.k(-48);
         }
      } catch (RuntimeException var3) {
         throw i.a(var3, O[3] + var1 + ')');
      }
   }

   final int f(int var1, int var2) {
      boolean var6 = client.vh;

      try {
         J++;
         int var3 = this.M >> 3;
         int var4 = -(this.M & 7) + 8;
         if (var1 >= -67) {
            K = (BigInteger)null;
         }

         int var5 = 0;
         this.M += var2;

         while (true) {
            if (~var4 > ~var2) {
               var5 += (mb.i[var4] & this.F[var3++]) << -var4 + var2;
               var2 -= var4;
               var4 = 8;
               if (var6) {
                  break;
               }

               if (!var6) {
                  continue;
               }
            }

            if (~var4 != ~var2) {
               var5 += this.F[var3] >> -var2 + var4 & mb.i[var2];
               if (!var6) {
                  return var5;
               }
            }
            break;
         }

         var5 += this.F[var3] & mb.i[var4];
         return var5;
      } catch (RuntimeException var7) {
         throw i.a(var7, O[2] + var1 + 44 + var2 + 41);
      }
   }

   ja(int var1) {
      super(var1);
   }

   final void j(int var1) {
      try {
         if (var1 != 25505) {
            this.f(12, -68);
         }

         this.w = (7 + this.M) / 8;
         I++;
      } catch (RuntimeException var3) {
         throw i.a(var3, O[0] + var1 + ')');
      }
   }

   private static char[] z(String var0) {
      char[] var10000 = var0.toCharArray();
      if (var10000.length < 2) {
         var10000[0] = (char)(var10000[0] ^ 70);
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
               var10005 = 48;
               break;
            case 1:
               var10005 = 50;
               break;
            case 2:
               var10005 = 27;
               break;
            case 3:
               var10005 = 36;
               break;
            default:
               var10005 = 70;
         }

         var10001[var1] = (char)(var10004 ^ var10005);
      }

      return new String(var10001).intern();
   }
}
