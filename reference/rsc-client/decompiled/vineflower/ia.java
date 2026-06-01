final class ia {
   static int e;
   static int b = 0;
   static String[] a = new String[100];
   static int[] d = new int[256];
   static int h;
   static int f;
   static int i = 0;
   static String[] g = new String[100];
   static int c;
   private static final String[] z = new String[]{
      z(z("3Wd\\*")), z(z("&\f&\u001e")), z(z("!\u0018d0\u007f")), z(z("!\u0018d1\u007f")), z(z("!\u0018d3\u007f"))
   };

   static final void a(int var0, int var1, int[] var2, int var3, int var4, int[] var5, int var6, byte var7) {
      boolean var11 = client.vh;

      try {
         e++;
         if (0 > var6) {
            var1 = var2[(65465 & var3) >> -1668975128];
            var0 <<= 1;
            var3 += var0;
            int var17 = var6 / 8;
            int var10 = 69 % ((var7 - 27) / 45);
            int var9 = var17;

            while (true) {
               if (0 > var9) {
                  var5[var4++] = var1;
                  var5[var4++] = var1;
                  var1 = var2[(65503 & var3) >> 2132179464];
                  var3 += var0;
                  var5[var4++] = var1;
                  var5[var4++] = var1;
                  var1 = var2[(var3 & 65391) >> -1202367672];
                  var5[var4++] = var1;
                  var3 += var0;
                  var5[var4++] = var1;
                  var1 = var2[(65302 & var3) >> 1020025064];
                  var3 += var0;
                  var5[var4++] = var1;
                  var5[var4++] = var1;
                  var1 = var2[(var3 & 65386) >> -31505560];
                  var3 += var0;

                  try {
                     var9++;
                     if (var11) {
                        break;
                     }

                     if (!var11) {
                        continue;
                     }
                  } catch (RuntimeException var15) {
                     throw var15;
                  }
               }

               var17 = -(var6 % 8);
               break;
            }

            var9 = 0;

            while (~var9 > ~var17) {
               label67: {
                  try {
                     var5[var4++] = var1;
                     if (var11) {
                        break;
                     }

                     if (1 != (var9 & 1)) {
                        break label67;
                     }
                  } catch (RuntimeException var14) {
                     throw var14;
                  }

                  var1 = var2[var3 >> 867708808 & 0xFF];
                  var3 += var0;
               }

               var9++;
               if (var11) {
                  break;
               }
            }
         }
      } catch (RuntimeException var16) {
         RuntimeException var8 = var16;

         RuntimeException var10000;
         StringBuilder var10001;
         String var10002;
         label56: {
            try {
               var10000 = var8;
               var10001 = new StringBuilder().append(z[2]).append(var0).append(',').append(var1).append(',');
               if (var2 != null) {
                  var10002 = z[0];
                  break label56;
               }
            } catch (RuntimeException var13) {
               throw var13;
            }

            var10002 = z[1];
         }

         try {
            var10001 = var10001.append(var10002).append(',').append(var3).append(',').append(var4).append(',');
            if (var5 != null) {
               throw i.a(var10000, var10001.append(z[0]).append(',').append(var6).append(',').append((int)var7).append(')').toString());
            }
         } catch (RuntimeException var12) {
            throw var12;
         }

         throw i.a(var10000, var10001.append(z[1]).append(',').append(var6).append(',').append((int)var7).append(')').toString());
      }
   }

   static final String a(tb var0, boolean var1) {
      try {
         try {
            if (var1) {
               a = (String[])null;
            }
         } catch (RuntimeException var4) {
            throw var4;
         }

         c++;
         return client.a(0, var0, 32767);
      } catch (RuntimeException var5) {
         RuntimeException var2 = var5;

         RuntimeException var10000;
         StringBuilder var10001;
         try {
            var10000 = var2;
            var10001 = new StringBuilder().append(z[3]);
            if (var0 != null) {
               throw i.a(var2, var10001.append(z[0]).append(',').append(var1).append(')').toString());
            }
         } catch (RuntimeException var3) {
            throw var3;
         }

         throw i.a(var10000, var10001.append(z[1]).append(',').append(var1).append(')').toString());
      }
   }

   static final boolean a(v param0, byte param1) {
      // $VF: Couldn't be decompiled
      // Please report this to the Vineflower issue tracker, at https://github.com/Vineflower/vineflower/issues with a copy of the class file (if you have the rights to distribute it!)
      // java.lang.RuntimeException: parsing failure!
      //   at org.jetbrains.java.decompiler.modules.decompiler.decompose.DomHelper.parseGraph(DomHelper.java:211)
      //   at org.jetbrains.java.decompiler.main.rels.MethodProcessor.codeToJava(MethodProcessor.java:166)
      //
      // Bytecode:
      // 00: bipush 3
      // 01: bipush -39
      // 03: iload 1
      // 04: isub
      // 05: bipush 54
      // 07: idiv
      // 08: idiv
      // 09: istore 2
      // 0a: getstatic ia.f I
      // 0d: bipush 1
      // 0e: iadd
      // 0f: putstatic ia.f I
      // 12: getstatic da.O Lv;
      // 15: aload 0
      // 16: if_acmpeq 45
      // 19: getstatic ga.c Lv;
      // 1c: aload 0
      // 1d: if_acmpeq 45
      // 20: goto 24
      // 23: athrow
      // 24: getstatic ta.f Lv;
      // 27: aload 0
      // 28: if_acmpeq 45
      // 2b: goto 2f
      // 2e: athrow
      // 2f: aload 0
      // 30: getstatic eb.d Lv;
      // 33: if_acmpeq 45
      // 36: goto 3a
      // 39: athrow
      // 3a: getstatic gb.n Lv;
      // 3d: aload 0
      // 3e: if_acmpne 4a
      // 41: goto 45
      // 44: athrow
      // 45: bipush 1
      // 46: goto 4b
      // 49: athrow
      // 4a: bipush 0
      // 4b: ireturn
      // 4c: astore 2
      // 4d: aload 2
      // 4e: new java/lang/StringBuilder
      // 51: dup
      // 52: invokespecial java/lang/StringBuilder.<init> ()V
      // 55: getstatic ia.z [Ljava/lang/String;
      // 58: bipush 4
      // 59: aaload
      // 5a: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 5d: aload 0
      // 5e: ifnull 6a
      // 61: getstatic ia.z [Ljava/lang/String;
      // 64: bipush 0
      // 65: aaload
      // 66: goto 6f
      // 69: athrow
      // 6a: getstatic ia.z [Ljava/lang/String;
      // 6d: bipush 1
      // 6e: aaload
      // 6f: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 72: bipush 44
      // 74: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 77: iload 1
      // 78: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 7b: bipush 41
      // 7d: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 80: invokevirtual java/lang/StringBuilder.toString ()Ljava/lang/String;
      // 83: invokestatic i.a (Ljava/lang/Throwable;Ljava/lang/String;)Lla;
      // 86: athrow
   }

   private static char[] z(String var0) {
      char[] var10000 = var0.toCharArray();
      if (var10000.length < 2) {
         var10000[0] = (char)(var10000[0] ^ 'W');
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
               var10005 = 72;
               break;
            case 1:
               var10005 = 121;
               break;
            case 2:
               var10005 = 74;
               break;
            case 3:
               var10005 = 114;
               break;
            default:
               var10005 = 87;
         }

         var10001[var1] = (char)(var10004 ^ var10005);
      }

      return new String(var10001).intern();
   }
}
