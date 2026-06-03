import java.awt.Component;
import java.awt.Image;
import java.awt.image.IndexColorModel;

final class pa {
   static c b;
   static int[] g = new int[100];
   static int[] f;
   private static byte[] e = new byte[64];
   static int[] a = new int[512];
   static c k;
   static int i;
   static int c;
   static long h;
   static int[] j = new int[2048];
   static int[] d = new int[256];
   private static final String[] z = new String[]{z(z("J_\f\t%")), z(z("TKN'")), z(z("A\u0010\fep")), z(z("J_\f\n%"))};

   static final Image a(int var0, Component var1, byte[] var2) {
      try {
         k.o = var2[12] + var2[13] * 256;
         da.bb = var2[14] + 256 * var2[15];
         i++;
         byte[] var3 = new byte[256];
         byte[] var4 = new byte[256];
         byte[] var5 = new byte[256];

         for (int var6 = 0; -257 < ~var6; var6++) {
            var3[var6] = var2[20 + var6 * 3];
            var4[var6] = var2[3 * var6 + 19];
            var5[var6] = var2[3 * var6 + 18];
         }

         m.d = new IndexColorModel(8, 256, var3, var4, var5);
         byte[] var11 = new byte[k.o * da.bb];
         int var7 = 0;

         for (int var8 = -1 + da.bb; -1 >= ~var8; var8--) {
            for (int var9 = 0; k.o > var9; var9++) {
               var11[var7++] = var2[k.o * var8 + 786 + var9];
            }
         }

         Image var12 = var1.createImage(da.db);
         int var13 = 113 / ((var0 - 10) / 53);
         lb.a(true, var11);
         var1.prepareImage(var12, da.db);
         lb.a(true, var11);
         var1.prepareImage(var12, da.db);
         lb.a(true, var11);
         var1.prepareImage(var12, da.db);
         return var12;
      } catch (RuntimeException var10) {
         throw i.a(var10, z[0] + var0 + ',' + (var1 != null ? z[2] : z[1]) + ',' + (var2 != null ? z[2] : z[1]) + ')');
      }
   }

   static final byte[] a(int var0) {
      try {
         c++;
         byte[] var1 = new byte[256];
         int var2 = -128;
         if (var0 > -125) {
            a(-7, (Component)null, (byte[])null);
         }

         while (var2 < 127) {
            int var3 = (byte)var2;
            var3 = ~var3;
            int var4 = var3 & 128;
            int var5 = var3 >> 4 & 7;
            int var6 = 15 & var3;
            var6 |= 16;
            var6 = 1 + (var6 << 1);
            int var7 = var6 << var5 + 2;
            var7 = -132 + var7;
            if (~var4 != -1) {
               var7 = -var7;
            }

            var1[ib.a(var2, 255)] = (byte)(var7 / 256);
            var2++;
         }

         return var1;
      } catch (RuntimeException var8) {
         throw i.a(var8, z[3] + var0 + ')');
      }
   }

   static {
      for (int var0 = 0; 256 > var0; var0++) {
         a[var0] = (int)(32768.0 * Math.sin(0.02454369 * var0));
         a[256 + var0] = (int)(32768.0 * Math.cos(var0 * 0.02454369));
      }

      for (int var1 = 0; -1025 < ~var1; var1++) {
         j[var1] = (int)(Math.sin(var1 * 0.00613592315) * 32768.0);
         j[var1 - -1024] = (int)(Math.cos(var1 * 0.00613592315) * 32768.0);
      }

      for (int var2 = 0; -11 < ~var2; var2++) {
         e[var2] = (byte)(48 + var2);
      }

      for (int var3 = 0; ~var3 > -27; var3++) {
         e[var3 - -10] = (byte)(var3 + 65);
      }

      for (int var4 = 0; 26 > var4; var4++) {
         e[var4 - -36] = (byte)(97 + var4);
      }

      e[63] = 36;
      e[62] = -93;
      int var5 = 0;

      while (-11 < ~var5) {
         d[var5 + 48] = var5++;
      }

      for (int var6 = 0; -27 < ~var6; var6++) {
         d[var6 + 65] = var6 - -10;
      }

      for (int var7 = 0; var7 < 26; var7++) {
         d[var7 + 97] = 36 + var7;
      }

      d[36] = 63;
      d[163] = 62;
   }

   private static char[] z(String var0) {
      char[] var10000 = var0.toCharArray();
      if (var10000.length < 2) {
         var10000[0] = (char)(var10000[0] ^ 13);
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
               var10005 = 58;
               break;
            case 1:
               var10005 = 62;
               break;
            case 2:
               var10005 = 34;
               break;
            case 3:
               var10005 = 75;
               break;
            default:
               var10005 = 13;
         }

         var10001[var1] = (char)(var10004 ^ var10005);
      }

      return new String(var10001).intern();
   }
}
