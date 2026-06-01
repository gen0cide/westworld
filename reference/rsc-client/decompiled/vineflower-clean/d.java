import java.io.EOFException;
import java.io.File;
import java.io.IOException;
import java.io.RandomAccessFile;

final class d {
   static int n;
   static int e;
   static int l = 0;
   static int k;
   static int d;
   static int i;
   static int f;
   static int[] g;
   static int a;
   private RandomAccessFile j;
   private long c;
   static e h;
   private long b;
   static String[] m = new String[]{
      z(z("(X\u001boB\u001d\u0014\u001b`E\u001dF^zY\u001d\u0014\u0010{\\\u001aQ\f.^\u001e\u0014\u0017zT\u0015G^z^XC\u0017zY\u001cF\u001fy")),
      z(z("\u0019Z\u001a.A\nQ\r}\u0011\u001dZ\nkC"))
   };
   private static final String[] z = new String[]{
      z(z("\u0003\u001aP L")),
      z(z("\u001c\u001a:&")),
      z(z("\u0016A\u0012b")),
      z(z("\u001c\u001a<&")),
      z(z("\u001c\u001a8&")),
      z(z("\u001c\u001aBg_\u0011@@&")),
      z(z("\u001c\u001a\u0018g_\u0019X\u0017tTP\u001d")),
      z(z("\u001c\u001a9&")),
      z(z("\u001c\u001a?&")),
      z(z("\u001c\u001a;&")),
      z(z("\u001c\u001a=&"))
   };

   private final void a(int var1) throws IOException {
      try {
         n++;
         if (null != this.j) {
            this.j.close();
            this.j = null;
         }

         if (var1 != 25291) {
            a(62, (byte)14, (byte[])null);
         }
      } catch (RuntimeException var3) {
         throw i.a(var3, z[9] + var1 + ')');
      }
   }

   final long a(byte var1) throws IOException {
      try {
         if (var1 != 47) {
            this.c = -52L;
         }

         i++;
         return this.j.length();
      } catch (RuntimeException var3) {
         throw i.a(var3, z[8] + var1 + ')');
      }
   }

   @Override
   protected final void finalize() throws Throwable {
      try {
         a++;
         if (null != this.j) {
            System.out.println("");
            this.a(25291);
         }
      } catch (RuntimeException var2) {
         throw i.a(var2, z[6]);
      }
   }

   static int a(int var0, int var1) {
      try {
         return var0 | var1;
      } catch (RuntimeException var3) {
         throw i.a(var3, z[3] + var0 + 44 + var1 + 41);
      }
   }

   final void b(byte[] var1, int var2, int var3, int var4) throws IOException {
      try {
         d++;
         if (~this.b > ~(this.c + var2)) {
            this.j.seek(this.b);
            this.j.write(1);
            throw new EOFException();
         }

         this.j.write(var1, var4, var2);
         this.c += var2;
         if (var3 != 1) {
            a(63, (byte)-101, (byte[])null);
         }
      } catch (RuntimeException var6) {
         throw i.a(var6, z[7] + (var1 != null ? z[0] : z[2]) + ',' + var2 + ',' + var3 + ',' + var4 + ')');
      }
   }

   static final int a(int var0, byte var1, byte[] var2) {
      try {
         if (var1 < 4) {
            h = (e)null;
         }

         f++;
         return (var2[1 + var0] & 0xFF) + ((0xFF & var2[var0]) << 922410888);
      } catch (RuntimeException var4) {
         throw i.a(var4, z[1] + var0 + 44 + var1 + 44 + (var2 != null ? z[0] : z[2]) + 41);
      }
   }

   final void a(int var1, long var2) throws IOException {
      try {
         if (var1 == 0) {
            this.j.seek(var2);
            k++;
            this.c = var2;
         }
      } catch (RuntimeException var5) {
         throw i.a(var5, z[4] + var1 + ',' + var2 + ')');
      }
   }

   final int a(byte[] var1, int var2, int var3, int var4) throws IOException {
      try {
         e++;
         int var5 = this.j.read(var1, var3, var2);
         if (~var5 < var4) {
            this.c += var5;
         }

         return var5;
      } catch (RuntimeException var6) {
         throw i.a(var6, z[10] + (var1 != null ? z[0] : z[2]) + 44 + var2 + 44 + var3 + 44 + var4 + 41);
      }
   }

   d(File var1, String var2, long var3) throws IOException {
      try {
         if (0L == ~var3) {
            var3 = Long.MAX_VALUE;
         }

         if (~var3 > ~var1.length()) {
            var1.delete();
         }

         this.j = new RandomAccessFile(var1, var2);
         this.b = var3;
         this.c = 0L;
         int var5 = this.j.read();
         if (0 != ~var5 && !var2.equals("r")) {
            this.j.seek(0L);
            this.j.write(var5);
         }

         this.j.seek(0L);
      } catch (RuntimeException var6) {
         throw i.a(var6, z[5] + (var1 != null ? z[0] : z[2]) + ',' + (var2 != null ? z[0] : z[2]) + ',' + var3 + ')');
      }
   }

   private static char[] z(String var0) {
      char[] var10000 = var0.toCharArray();
      if (var10000.length < 2) {
         var10000[0] = (char)(var10000[0] ^ 49);
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
               var10005 = 120;
               break;
            case 1:
               var10005 = 52;
               break;
            case 2:
               var10005 = 126;
               break;
            case 3:
               var10005 = 14;
               break;
            default:
               var10005 = 49;
         }

         var10001[var1] = (char)(var10004 ^ var10005);
      }

      return new String(var10001).intern();
   }
}
