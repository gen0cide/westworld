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
         try {
            v++;
            if (var1 != -19675) {
               f = (v)null;
            }
         } catch (EOFException var10) {
            throw var10;
         }

         InputStream var12 = nb.a(true, var0);
         DataInputStream var5 = new DataInputStream(var12);

         try {
            var5.readFully(var2, 0, var3);
         } catch (EOFException var7) {
         }

         var5.close();
      } catch (RuntimeException var11) {
         RuntimeException var4 = var11;

         RuntimeException var10000;
         StringBuilder var10001;
         String var10002;
         label46: {
            try {
               var10000 = var4;
               var10001 = new StringBuilder().append(L[1]);
               if (var0 != null) {
                  var10002 = L[3];
                  break label46;
               }
            } catch (EOFException var9) {
               throw var9;
            }

            var10002 = L[2];
         }

         try {
            var10001 = var10001.append(var10002).append(',').append(var1).append(',');
            if (var2 != null) {
               throw i.a(var10000, var10001.append(L[3]).append(',').append(var3).append(')').toString());
            }
         } catch (EOFException var8) {
            throw var8;
         }

         throw i.a(var10000, var10001.append(L[2]).append(',').append(var3).append(')').toString());
      }
   }

   static final int a(int var0, byte var1) {
      try {
         int var2 = 110 / ((65 - var1) / 44);
         l++;
         var0 = (1431655765 & var0 >>> 1) + (1431655765 & var0);
         var0 = ((var0 & -858993460) >>> 2) + (858993459 & var0);
         var0 = var0 - -(var0 >>> 4) & 252645135;
         var0 += var0 >>> 8;
         var0 += var0 >>> 16;
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
         var10000[0] = (char)(var10000[0] ^ '=');
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
