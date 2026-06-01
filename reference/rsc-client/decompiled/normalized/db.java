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
         try {
            h++;
            if (null != var1.e) {
               var1.a(-27331);
            }
         } catch (RuntimeException var6) {
            throw var6;
         }

         try {
            var1.e = this.k;
            var1.a = this.k.a;
            var1.e.a = var1;
            var1.a.e = var1;
            if (var2) {
               this.b((byte)78);
            }
         } catch (RuntimeException var4) {
            throw var4;
         }
      } catch (RuntimeException var7) {
         RuntimeException var3 = var7;

         RuntimeException var10000;
         StringBuilder var10001;
         try {
            var10000 = var3;
            var10001 = new StringBuilder().append(z[4]);
            if (var1 != null) {
               throw i.a(var3, var10001.append(z[2]).append(',').append(var2).append(')').toString());
            }
         } catch (RuntimeException var5) {
            throw var5;
         }

         throw i.a(var10000, var10001.append(z[3]).append(',').append(var2).append(')').toString());
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

         label21: {
            try {
               if (this.k == var2) {
                  break label21;
               }
            } catch (RuntimeException var4) {
               throw var4;
            }

            this.b = var2.a;
            return var2;
         }

         this.b = null;
         return null;
      } catch (RuntimeException var5) {
         throw i.a(var5, z[1] + var1 + ')');
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

   static final boolean a(int param0, char param1) {
      // $VF: Couldn't be decompiled
      // Please report this to the Vineflower issue tracker, at https://github.com/Vineflower/vineflower/issues with a copy of the class file (if you have the rights to distribute it!)
      // java.lang.RuntimeException: parsing failure!
      //   at org.jetbrains.java.decompiler.modules.decompiler.decompose.DomHelper.parseGraph(DomHelper.java:211)
      //   at org.jetbrains.java.decompiler.main.rels.MethodProcessor.codeToJava(MethodProcessor.java:166)
      //
      // Bytecode:
      // 00: getstatic db.e I
      // 03: bipush 1
      // 04: iadd
      // 05: putstatic db.e I
      // 08: iload 0
      // 09: bipush 32
      // 0b: if_icmpeq 10
      // 0e: bipush 0
      // 0f: ireturn
      // 10: iload 1
      // 11: bipush -1
      // 12: ixor
      // 13: sipush -161
      // 16: if_icmpeq 3b
      // 19: iload 1
      // 1a: bipush 32
      // 1c: if_icmpeq 3b
      // 1f: goto 23
      // 22: athrow
      // 23: iload 1
      // 24: bipush -1
      // 25: ixor
      // 26: bipush -96
      // 28: if_icmpeq 3b
      // 2b: goto 2f
      // 2e: athrow
      // 2f: iload 1
      // 30: bipush -1
      // 31: ixor
      // 32: bipush -46
      // 34: if_icmpne 40
      // 37: goto 3b
      // 3a: athrow
      // 3b: bipush 1
      // 3c: goto 41
      // 3f: athrow
      // 40: bipush 0
      // 41: ireturn
      // 42: astore 2
      // 43: aload 2
      // 44: new java/lang/StringBuilder
      // 47: dup
      // 48: invokespecial java/lang/StringBuilder.<init> ()V
      // 4b: getstatic db.z [Ljava/lang/String;
      // 4e: bipush 0
      // 4f: aaload
      // 50: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 53: iload 0
      // 54: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 57: bipush 44
      // 59: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 5c: iload 1
      // 5d: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 60: bipush 41
      // 62: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 65: invokevirtual java/lang/StringBuilder.toString ()Ljava/lang/String;
      // 68: invokestatic i.a (Ljava/lang/Throwable;Ljava/lang/String;)Lla;
      // 6b: athrow
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
