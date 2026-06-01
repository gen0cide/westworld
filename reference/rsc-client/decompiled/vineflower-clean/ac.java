final class ac {
   byte[] j;
   int L;
   int E;
   int G;
   byte[] q;
   boolean[] n;
   static String[] z = new String[200];
   int[][] J = new int[6][258];
   static String[] x;
   int[] r;
   int h;
   int F;
   byte[] i;
   static int C;
   byte[] d;
   byte[] s;
   int f;
   byte[] A;
   int c;
   int[] D;
   static int[] l;
   int o;
   int y;
   int p;
   int[][] t;
   int[] w;
   int a;
   static char[] I = new char[]{'[', ']', '#'};
   int b;
   static int k;
   int[] m;
   int e;
   int H;
   boolean[] v;
   int K;
   int[][] u;
   byte g;
   byte[][] B;
   private static final String[] M = new String[]{z(z("h_og%")), z(z("gI-J")), z(z("h_od%")), z(z("r\u0012o\bp"))};

   static final void a(ib var0, byte var1, ib var2) {
      try {
         if (null != var0.e) {
            var0.a(-27331);
         }

         k++;
         if (var1 != 34) {
            x = (String[])null;
         }

         var0.a = var2;
         var0.e = var2.e;
         var0.e.a = var0;
         var0.a.e = var0;
      } catch (RuntimeException var4) {
         throw i.a(var4, M[2] + (var0 != null ? M[3] : M[1]) + ',' + var1 + ',' + (var2 != null ? M[3] : M[1]) + ')');
      }
   }

   static final char a(char var0, int var1) {
      boolean var3 = client.vh;

      try {
         if (var1 != -194) {
            a('ﾆ', 87);
         }

         C++;
         char var2 = var0;
         if ((-33 != ~var2 || var3) && 160 != var2 && '_' != var2 && (-46 != ~var2 || var3)) {
            if ((~var2 != -92 || var3) && -94 != ~var2 && ('#' != var2 || var3)) {
               if ((224 != var2 || var3)
                  && (225 != var2 || var3)
                  && (~var2 != -227 || var3)
                  && (var2 != 228 || var3)
                  && var2 != 227
                  && (var2 != 192 || var3)
                  && (-194 != ~var2 || var3)
                  && 194 != var2
                  && var2 != 196
                  && (195 != var2 || var3)) {
                  if (var2 == 232 || -234 == ~var2 && !var3 || 234 == var2 || 235 == var2 || ~var2 == -201 || 201 == var2 || var2 == 202 || ~var2 == -204) {
                     return 'e';
                  }

                  if (var2 == 237 || var2 == 238 && !var3 || ~var2 == -240 || ~var2 == -206 && !var3 || ~var2 == -207 && !var3 || ~var2 == -208) {
                     return 'i';
                  }

                  if ((-243 != ~var2 || var3)
                     && -244 != ~var2
                     && (-245 != ~var2 || var3)
                     && -247 != ~var2
                     && (var2 != 245 || var3)
                     && var2 != 210
                     && (~var2 != -212 || var3)
                     && var2 != 212
                     && ~var2 != -215
                     && -214 != ~var2) {
                     if ((249 != var2 || var3)
                        && (-251 != ~var2 || var3)
                        && (-252 != ~var2 || var3)
                        && 252 != var2
                        && ~var2 != -218
                        && var2 != 218
                        && (var2 != 219 || var3)
                        && 220 != var2) {
                        if (-232 == ~var2 || 199 == var2) {
                           return 'c';
                        }

                        if ((255 != var2 || var3) && (-377 != ~var2 || var3)) {
                           if ((var2 != 241 || var3) && (209 != var2 || var3)) {
                              return -224 != ~var2 ? Character.toLowerCase(var0) : 'b';
                           } else {
                              return 'n';
                           }
                        } else {
                           return 'y';
                        }
                     } else {
                        return 'u';
                     }
                  } else {
                     return 'o';
                  }
               } else {
                  return 'a';
               }
            } else {
               return var0;
            }
         } else {
            return '_';
         }
      } catch (RuntimeException var4) {
         throw i.a(var4, M[0] + var0 + ',' + var1 + ')');
      }
   }

   ac() {
      this.j = new byte[18002];
      this.n = new boolean[256];
      this.m = new int[256];
      this.r = new int[16];
      this.s = new byte[18002];
      this.A = new byte[4096];
      this.o = 0;
      this.w = new int[257];
      this.u = new int[6][258];
      this.D = new int[6];
      this.d = new byte[256];
      this.a = 0;
      this.v = new boolean[16];
      this.B = new byte[6][258];
      this.t = new int[6][258];
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
               var10005 = 9;
               break;
            case 1:
               var10005 = 60;
               break;
            case 2:
               var10005 = 65;
               break;
            case 3:
               var10005 = 38;
               break;
            default:
               var10005 = 13;
         }

         var10001[var1] = (char)(var10004 ^ var10005);
      }

      return new String(var10001).intern();
   }
}
