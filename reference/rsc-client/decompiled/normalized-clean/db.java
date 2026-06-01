final class db {
   static int h;
   static int[] j;
   static int d;
   static int[] l;
   static byte[] i;
   static int a;
   static i f;
   ib k = new ib();
   static int e;
   static int g = 0;
   static int c;
   private ib b;
   private static final String[] z = new String[]{
      z(z("}\rZR?")), z(z("}\rZQ?")), z(z("bAZ>j")), z(z("w\u001a\u0018|")), z(z("}\rZS?")), z(z("}\rZ,~w\u0006\u0000.?0")), z(z("}\rZT?"))
   };

   final void a(ib var1, boolean var2) {
      try {
         h++;
         if (null != var1.e) {
            var1.a(-27331);
         }

         var1.e = this.k;
         var1.a = this.k.a;
         var1.e.a = var1;
         var1.a.e = var1;
         if (var2) {
            this.b((byte)78);
         }
      } catch (RuntimeException var4) {
         throw i.a(var4, z[4] + (var1 != null ? z[2] : z[3]) + ',' + var2 + ')');
      }
   }

   final ib a(byte var1) {
      try {
         c++;
         ib var2 = this.k.a;
         if (this.k == var2) {
            this.b = null;
            return null;
         } else {
            int var3 = 119 % ((var1 - -37) / 43);
            this.b = var2.a;
            return var2;
         }
      } catch (RuntimeException var4) {
         throw i.a(var4, z[6] + var1 + ')');
      }
   }

   final ib b(byte var1) {
      try {
         a++;
         int var3 = 81 % ((-37 - var1) / 51);
         ib var2 = this.b;
         if (this.k != var2) {
            this.b = var2.a;
            return var2;
         } else {
            this.b = null;
            return null;
         }
      } catch (RuntimeException var4) {
         throw i.a(var4, z[1] + var1 + ')');
      }
   }

   public db() {
      try {
         this.k.e = this.k;
         this.k.a = this.k;
      } catch (RuntimeException var2) {
         throw i.a(var2, z[5]);
      }
   }

   static final boolean a(int var0, char var1) {
      try {
         e++;
         return var0 != 32 ? false : ~var1 == -161 || var1 == ' ' || ~var1 == -96 || ~var1 == -46;
      } catch (RuntimeException var3) {
         throw i.a(var3, z[0] + var0 + ',' + var1 + ')');
      }
   }

   private static char[] z(String var0) {
      char[] var10000 = var0.toCharArray();
      if (var10000.length < 2) {
         var10000[0] = (char)(var10000[0] ^ 23);
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
               var10005 = 25;
               break;
            case 1:
               var10005 = 111;
               break;
            case 2:
               var10005 = 116;
               break;
            case 3:
               var10005 = 16;
               break;
            default:
               var10005 = 23;
         }

         var10001[var1] = (char)(var10004 ^ var10005);
      }

      return new String(var10001).intern();
   }
}
