import java.io.DataInputStream;
import java.io.EOFException;
import java.io.IOException;
import java.io.InputStream;

final class ta {
   int t;
   static v f = new v(z(z("\u001c\u000b\u0016\u0013m")), z(z("$9'3^.")), z(z("\u0014((*")), 3);
   int y;
   int s;
   int[] m = new int[12];
   int[] F = new int[10];
   int D;
   String n;
   int p;
   int q;
   int z;
   String C;
   int b;
   int x;
   int[] k = new int[10];
   int j;
   static int l;
   static String[] r;
   int u;
   int A;
   int h;
   int J;
   static int v;
   int H;
   int e;
   int i;
   static int g = 176;
   int o;
   String c;
   int I;
   int B;
   int a;
   int w;
   int G;
   int E;
   int K;
   int d;
   private static final String[] L = new String[]{z(z("?>o\u0018\u0015")), z(z("?>o\u001b\u0015")), z(z("%*-6")), z(z("0qot@"))};

   static final void a(String var0, int var1, byte[] var2, int var3) throws IOException {
      try {
         v++;
         if (var1 != -19675) {
            f = (v)null;
         }

         InputStream var4 = nb.a(true, var0);
         DataInputStream var5 = new DataInputStream(var4);

         try {
            var5.readFully(var2, 0, var3);
         } catch (EOFException var7) {
         }

         var5.close();
      } catch (RuntimeException var8) {
         throw i.a(var8, L[1] + (var0 != null ? L[3] : L[2]) + ',' + var1 + ',' + (var2 != null ? L[3] : L[2]) + ',' + var3 + ')');
      }
   }

   static final int a(int var0, byte var1) {
      try {
         int var2 = 110 / ((65 - var1) / 44);
         l++;
         var0 = (1431655765 & var0 >>> -2016269343) + (1431655765 & var0);
         var0 = ((var0 & -858993460) >>> 209493378) + (858993459 & var0);
         var0 = var0 - -(var0 >>> 1603286532) & 252645135;
         var0 += var0 >>> -843073400;
         var0 += var0 >>> -761392816;
         return 0xFF & var0;
      } catch (RuntimeException var3) {
         throw i.a(var3, L[0] + var0 + 44 + var1 + 41);
      }
   }

   ta() {
      this.s = -1;
      this.J = 0;
      this.z = 0;
      this.u = 0;
      this.h = 0;
      this.a = 0;
      this.w = 0;
      this.I = 0;
      this.B = 0;
      this.E = 0;
      this.G = 0;
      this.d = 0;
   }

   private static char[] z(String var0) {
      char[] var10000 = var0.toCharArray();
      if (var10000.length < 2) {
         var10000[0] = (char)(var10000[0] ^ 61);
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
               var10005 = 75;
               break;
            case 1:
               var10005 = 95;
               break;
            case 2:
               var10005 = 65;
               break;
            case 3:
               var10005 = 90;
               break;
            default:
               var10005 = 61;
         }

         var10001[var1] = (char)(var10004 ^ var10005);
      }

      return new String(var10001).intern();
   }
}
