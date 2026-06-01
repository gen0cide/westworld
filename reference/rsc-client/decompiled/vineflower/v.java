final class v {
   static int c;
   int i;
   static int[] a;
   static int f;
   static int h;
   static int d;
   static int b;
   static int[] g;
   static int[] e;
   private static final String[] z = new String[]{
      z(z("\"tv%")), z(z(":/& /%u$a")), z(z("7/4g<")), z(z(":/Ya")), z(z(":/[a")), z(z(":/n&\u00128ss'&d(")), z(z(":/Xa"))
   };

   static final byte[] a(byte[] var0, int var1, int var2, int var3) {
      boolean var6 = client.vh;

      try {
         try {
            c++;
            if (var2 != -98) {
               a = (int[])null;
            }
         } catch (RuntimeException var9) {
            throw var9;
         }

         byte[] var11 = new byte[var1];
         int var5 = 0;

         byte[] var12;
         while (true) {
            if (var5 < var1) {
               try {
                  var12 = var11;
                  if (var6) {
                     break;
                  }

                  var11[var5] = qa.l[ib.a(var0[var3 + var5], 255)];
                  var5++;
                  if (!var6) {
                     continue;
                  }
               } catch (RuntimeException var8) {
                  throw var8;
               }
            }

            var12 = var11;
            break;
         }

         return var12;
      } catch (RuntimeException var10) {
         RuntimeException var4 = var10;

         RuntimeException var10000;
         StringBuilder var10001;
         try {
            var10000 = var4;
            var10001 = new StringBuilder().append(z[3]);
            if (var0 != null) {
               throw i.a(var4, var10001.append(z[2]).append(',').append(var1).append(',').append(var2).append(',').append(var3).append(')').toString());
            }
         } catch (RuntimeException var7) {
            throw var7;
         }

         throw i.a(var10000, var10001.append(z[0]).append(',').append(var1).append(',').append(var2).append(',').append(var3).append(')').toString());
      }
   }

   static final int a(int var0) {
      try {
         d++;
         int var1 = b.v[ka.b] & 255;

         try {
            ka.b++;
            if (var0 != -30504) {
               a(113);
            }
         } catch (RuntimeException var2) {
            throw var2;
         }

         return var1;
      } catch (RuntimeException var3) {
         throw i.a(var3, z[4] + var0 + 41);
      }
   }

   @Override
   public final String toString() {
      try {
         f++;
         throw new IllegalStateException();
      } catch (RuntimeException var2) {
         throw i.a(var2, z[5]);
      }
   }

   static final boolean a(char param0, int param1) {
      // $VF: Couldn't be decompiled
      // Please report this to the Vineflower issue tracker, at https://github.com/Vineflower/vineflower/issues with a copy of the class file (if you have the rights to distribute it!)
      // java.lang.RuntimeException: parsing failure!
      //   at org.jetbrains.java.decompiler.modules.decompiler.decompose.DomHelper.parseGraph(DomHelper.java:211)
      //   at org.jetbrains.java.decompiler.main.rels.MethodProcessor.codeToJava(MethodProcessor.java:166)
      //
      // Bytecode:
      // 00: getstatic v.b I
      // 03: bipush 1
      // 04: iadd
      // 05: putstatic v.b I
      // 08: iload 1
      // 09: bipush 111
      // 0b: if_icmpgt 20
      // 0e: aconst_null
      // 0f: checkcast [B
      // 12: bipush 51
      // 14: bipush 127
      // 16: bipush 27
      // 18: invokestatic v.a ([BIII)[B
      // 1b: pop
      // 1c: goto 20
      // 1f: athrow
      // 20: iload 0
      // 21: bipush -1
      // 22: ixor
      // 23: bipush -49
      // 25: if_icmpgt 34
      // 28: iload 0
      // 29: bipush -1
      // 2a: ixor
      // 2b: bipush -58
      // 2d: if_icmpge 62
      // 30: goto 34
      // 33: athrow
      // 34: iload 0
      // 35: bipush 65
      // 37: if_icmplt 4a
      // 3a: goto 3e
      // 3d: athrow
      // 3e: iload 0
      // 3f: bipush -1
      // 40: ixor
      // 41: bipush -91
      // 43: if_icmpge 62
      // 46: goto 4a
      // 49: athrow
      // 4a: bipush -98
      // 4c: iload 0
      // 4d: bipush -1
      // 4e: ixor
      // 4f: if_icmplt 67
      // 52: goto 56
      // 55: athrow
      // 56: bipush -123
      // 58: iload 0
      // 59: bipush -1
      // 5a: ixor
      // 5b: if_icmpgt 67
      // 5e: goto 62
      // 61: athrow
      // 62: bipush 1
      // 63: goto 68
      // 66: athrow
      // 67: bipush 0
      // 68: ireturn
      // 69: astore 2
      // 6a: aload 2
      // 6b: new java/lang/StringBuilder
      // 6e: dup
      // 6f: invokespecial java/lang/StringBuilder.<init> ()V
      // 72: getstatic v.z [Ljava/lang/String;
      // 75: bipush 6
      // 77: aaload
      // 78: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 7b: iload 0
      // 7c: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 7f: bipush 44
      // 81: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 84: iload 1
      // 85: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 88: bipush 41
      // 8a: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 8d: invokevirtual java/lang/StringBuilder.toString ()Ljava/lang/String;
      // 90: invokestatic i.a (Ljava/lang/Throwable;Ljava/lang/String;)Lla;
      // 93: athrow
   }

   v(String var1, String var2, String var3, int var4) {
      try {
         this.i = var4;
      } catch (RuntimeException var9) {
         RuntimeException var5 = var9;

         RuntimeException var10000;
         StringBuilder var10001;
         String var10002;
         label46: {
            try {
               var10000 = var5;
               var10001 = new StringBuilder().append(z[1]);
               if (var1 != null) {
                  var10002 = z[2];
                  break label46;
               }
            } catch (RuntimeException var8) {
               throw var8;
            }

            var10002 = z[0];
         }

         label39: {
            try {
               var10001 = var10001.append(var10002).append(',');
               if (var2 != null) {
                  var10002 = z[2];
                  break label39;
               }
            } catch (RuntimeException var7) {
               throw var7;
            }

            var10002 = z[0];
         }

         try {
            var10001 = var10001.append(var10002).append(',');
            if (var3 != null) {
               throw i.a(var10000, var10001.append(z[2]).append(',').append(var4).append(')').toString());
            }
         } catch (RuntimeException var6) {
            throw var6;
         }

         throw i.a(var10000, var10001.append(z[0]).append(',').append(var4).append(')').toString());
      }
   }

   private static char[] z(String var0) {
      char[] var10000 = var0.toCharArray();
      if (var10000.length < 2) {
         var10000[0] = (char)(var10000[0] ^ 'A');
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
               var10005 = 76;
               break;
            case 1:
               var10005 = 1;
               break;
            case 2:
               var10005 = 26;
               break;
            case 3:
               var10005 = 73;
               break;
            default:
               var10005 = 65;
         }

         var10001[var1] = (char)(var10004 ^ var10005);
      }

      return new String(var10001).intern();
   }
}
