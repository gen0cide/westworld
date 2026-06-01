import java.io.File;
import java.io.RandomAccessFile;
import java.util.Hashtable;

public class r {
   private static Hashtable b = new Hashtable(16);
   private static int c;
   private static boolean d = false;
   private static String e;
   private static String a;
   private static final String[] z = new String[]{
      z(z("8\r")),
      z(z("3Q\u0002Bh.M\nU")),
      z(z("iP\u0014S'%J\u0002\u001f")),
      z(z("hH\u0006W#>}\u0004Q%.G8")),
      z(z("%\u0018H")),
      z(z("%\u0018HG/(L\u0013\u001f")),
      z(z("%\u0018HB5%C\u0004X#i")),
      z(z("hD\u000e\\#\u0019Q\u0013_4#}")),
      z(z("iV\n@i")),
      z(z("4U")),
      z(z("%\u0018HG/(F\bG5i"))
   };

   public static File a(int var0, String var1) {
      try {
         return var0 != 2 ? (File)null : a(c, e, var1, 0);
      } catch (RuntimeException var3) {
         throw var3;
      }
   }

   public static File a(int var0, String var1, String var2, int var3) {
      try {
         if (!d) {
            throw new RuntimeException("");
         }

         File var4 = (File)b.get(var2);
         if (var4 != null) {
            return var4;
         }

         String[] var5 = new String[]{z[6], z[2], z[10], z[5], z[4], a, z[8], ""};
         String[] var6 = new String[]{z[3] + var0, z[7] + var0};

         for (int var7 = var3; 2 > var7; var7++) {
            for (int var8 = 0; var6.length > var8; var8++) {
               for (int var9 = 0; var9 < var5.length; var9++) {
                  String var10 = var5[var9] + var6[var8] + "/" + (var1 != null ? var1 + "/" : "") + var2;
                  RandomAccessFile var11 = null;

                  try {
                     File var12 = new File(var10);
                     if (var7 != 0 || var12.exists()) {
                        String var13 = var5[var9];
                        if (~var7 != -2 || 0 >= var13.length() || new File(var13).exists()) {
                           new File(var5[var9] + var6[var8]).mkdir();
                           if (null != var1) {
                              new File(var5[var9] + var6[var8] + "/" + var1).mkdir();
                           }

                           var11 = new RandomAccessFile(var12, z[9]);
                           int var14 = var11.read();
                           var11.seek(0L);
                           var11.write(var14);
                           var11.seek(0L);
                           var11.close();
                           b.put(var2, var12);
                           return var12;
                        }
                     }
                  } catch (Exception var16) {
                     try {
                        if (null != var11) {
                           var11.close();
                           Object var18 = null;
                        }
                     } catch (Exception var15) {
                     }
                  }
               }
            }
         }

         throw new RuntimeException();
      } catch (RuntimeException var17) {
         throw var17;
      }
   }

   public static void a(int var0, byte var1, String var2) {
      try {
         c = var0;
         e = var2;
         if (var1 != 101) {
            a(-64, (String)null, (String)null, -78);
         }

         try {
            a = System.getProperty(z[1]);
            if (null != a) {
               a = a + "/";
            }
         } catch (Exception var4) {
         }

         if (null == a) {
            a = z[0];
         }

         d = true;
      } catch (RuntimeException var5) {
         throw var5;
      }
   }

   private r() throws Throwable {
      try {
         throw new Error();
      } catch (RuntimeException var2) {
         throw var2;
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
               var10005 = 70;
               break;
            case 1:
               var10005 = 34;
               break;
            case 2:
               var10005 = 103;
               break;
            case 3:
               var10005 = 48;
               break;
            default:
               var10005 = 70;
         }

         var10001[var1] = (char)(var10004 ^ var10005);
      }

      return new String(var10001).intern();
   }
}
