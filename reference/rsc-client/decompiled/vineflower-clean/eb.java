final class eb implements Runnable {
   volatile boolean a;
   volatile sa[] f = new sa[2];
   static int h;
   c g;
   volatile boolean i;
   static i e = new i(z(z("qp>I")), 0);
   static v d = new v(z(z("jm!")), z(z("R_\u000eekX")), z(z("bN\u001ce")), 5);
   static int[] b;
   private static char[] c = new char[256];
   private static final String z = z(z("X[F~}S\u0011A"));

   // $VF: Could not verify finally blocks. A semaphore variable has been added to preserve control flow.
   // Please report this to the Vineflower issue tracker, at https://github.com/Vineflower/vineflower/issues with a copy of the class file (if you have the rights to distribute it!)
   @Override
   public final void run() {
      boolean var4 = client.vh;

      try {
         this.i = true;
         h++;

         while (true) {
            boolean var8 = false /* VF: Semaphore variable */;

            label92: {
               try {
                  var8 = true;
                  if (this.a) {
                     var8 = false;
                  } else {
                     if (var4) {
                        var8 = false;
                        break;
                     }

                     int var1 = 0;

                     while (true) {
                        if (~var1 > -3) {
                           sa var2 = this.f[var1];
                           if (var4) {
                              break;
                           }

                           if (var2 != null) {
                              var2.a();
                           }

                           var1++;
                           if (!var4) {
                              continue;
                           }
                        }

                        mb.a(11200, 10L);
                        ba.a(null, 1, this.g);
                        break;
                     }

                     if (!var4) {
                        continue;
                     }

                     var8 = false;
                  }
                  break label92;
               } catch (Exception var9) {
                  mb.a(2097151, var9, null);
                  var8 = false;
               } finally {
                  if (var8) {
                     this.i = false;
                  }
               }

               this.i = false;
               break;
            }

            this.i = false;
            break;
         }
      } catch (RuntimeException var11) {
         throw i.a(var11, z);
      }
   }

   eb() {
      this.a = false;
      this.i = false;
   }

   static {
      for (int var0 = 0; 256 > var0; var0++) {
         c[var0] = (char)var0;
      }

      c[45] = '-';
      c[59] = ';';
      c[42] = '*';
      c[124] = '|';
      c[43] = '+';
      c[33] = '!';
      c[34] = '"';
      c[47] = '/';
      c[46] = '.';
      c[61] = '=';
      c[92] = '\\';
      c[44] = ',';
   }

   private static char[] z(String var0) {
      char[] var10000 = var0.toCharArray();
      if (var10000.length < 2) {
         var10000[0] = (char)(var10000[0] ^ 8);
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
               var10005 = 61;
               break;
            case 1:
               var10005 = 57;
               break;
            case 2:
               var10005 = 104;
               break;
            case 3:
               var10005 = 12;
               break;
            default:
               var10005 = 8;
         }

         var10001[var1] = (char)(var10004 ^ var10005);
      }

      return new String(var10001).intern();
   }
}
