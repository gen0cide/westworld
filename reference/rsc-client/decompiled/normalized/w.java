final class w {
   int l;
   int r;
   int m;
   int p;
   int b;
   int j;
   boolean c = false;
   int q;
   int i;
   int e;
   static int n;
   static int d;
   int k;
   int t;
   int s;
   static int[] g;
   int f;
   int h;
   static int a;
   int u;
   ca o;
   private static final String[] z = new String[]{z(z("\u0011S8B")), z(z("\u001dSTDf")), z(z("\b\b\u0016\u0006")), z(z("\u0011S9B")), z(z("\u0011S;B"))};

   static final String a(CharSequence param0, byte param1) {
      // $VF: Couldn't be decompiled
      // Please report this to the Vineflower issue tracker, at https://github.com/Vineflower/vineflower/issues with a copy of the class file (if you have the rights to distribute it!)
      // java.lang.RuntimeException: parsing failure!
      //   at org.jetbrains.java.decompiler.modules.decompiler.decompose.DomHelper.parseGraph(DomHelper.java:211)
      //   at org.jetbrains.java.decompiler.main.rels.MethodProcessor.codeToJava(MethodProcessor.java:166)
      //
      // Bytecode:
      // 000: getstatic client.vh Z
      // 003: istore 9
      // 005: getstatic w.a I
      // 008: bipush 1
      // 009: iadd
      // 00a: putstatic w.a I
      // 00d: iload 1
      // 00e: bipush 47
      // 010: if_icmpgt 018
      // 013: aconst_null
      // 014: checkcast java/lang/String
      // 017: areturn
      // 018: aload 0
      // 019: ifnonnull 01e
      // 01c: aconst_null
      // 01d: areturn
      // 01e: bipush 0
      // 01f: istore 2
      // 020: aload 0
      // 021: invokeinterface java/lang/CharSequence.length ()I 1
      // 026: istore 3
      // 027: iload 2
      // 028: bipush -1
      // 029: ixor
      // 02a: iload 3
      // 02b: bipush -1
      // 02c: ixor
      // 02d: if_icmple 054
      // 030: bipush 32
      // 032: aload 0
      // 033: iload 2
      // 034: invokeinterface java/lang/CharSequence.charAt (I)C 2
      // 039: invokestatic db.a (IC)Z
      // 03c: iload 9
      // 03e: ifne 055
      // 041: ifeq 054
      // 044: goto 048
      // 047: athrow
      // 048: iinc 2 1
      // 04b: iload 9
      // 04d: ifeq 027
      // 050: goto 054
      // 053: athrow
      // 054: iload 3
      // 055: iload 2
      // 056: if_icmple 083
      // 059: bipush 32
      // 05b: aload 0
      // 05c: bipush -1
      // 05d: iload 3
      // 05e: iadd
      // 05f: invokeinterface java/lang/CharSequence.charAt (I)C 2
      // 064: invokestatic db.a (IC)Z
      // 067: iload 9
      // 069: ifne 086
      // 06c: goto 070
      // 06f: athrow
      // 070: ifeq 083
      // 073: goto 077
      // 076: athrow
      // 077: iinc 3 -1
      // 07a: iload 9
      // 07c: ifeq 054
      // 07f: goto 083
      // 082: athrow
      // 083: iload 3
      // 084: iload 2
      // 085: isub
      // 086: istore 4
      // 088: bipush 1
      // 089: iload 4
      // 08b: if_icmpgt 09f
      // 08e: bipush -13
      // 090: iload 4
      // 092: bipush -1
      // 093: ixor
      // 094: if_icmpgt 09f
      // 097: goto 09b
      // 09a: athrow
      // 09b: goto 0a1
      // 09e: athrow
      // 09f: aconst_null
      // 0a0: areturn
      // 0a1: new java/lang/StringBuilder
      // 0a4: dup
      // 0a5: iload 4
      // 0a7: invokespecial java/lang/StringBuilder.<init> (I)V
      // 0aa: astore 5
      // 0ac: iload 2
      // 0ad: istore 6
      // 0af: iload 3
      // 0b0: bipush -1
      // 0b1: ixor
      // 0b2: iload 6
      // 0b4: bipush -1
      // 0b5: ixor
      // 0b6: if_icmpge 104
      // 0b9: aload 0
      // 0ba: iload 6
      // 0bc: invokeinterface java/lang/CharSequence.charAt (I)C 2
      // 0c1: istore 7
      // 0c3: iload 7
      // 0c5: bipush 0
      // 0c6: invokestatic f.a (CI)Z
      // 0c9: iload 9
      // 0cb: ifne 105
      // 0ce: ifne 0de
      // 0d1: goto 0d5
      // 0d4: athrow
      // 0d5: iload 9
      // 0d7: ifeq 0fc
      // 0da: goto 0de
      // 0dd: athrow
      // 0de: iload 7
      // 0e0: sipush -194
      // 0e3: invokestatic ac.a (CI)C
      // 0e6: istore 8
      // 0e8: iload 8
      // 0ea: bipush -1
      // 0eb: ixor
      // 0ec: bipush -1
      // 0ed: if_icmpeq 0fc
      // 0f0: aload 5
      // 0f2: iload 8
      // 0f4: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 0f7: pop
      // 0f8: goto 0fc
      // 0fb: athrow
      // 0fc: iinc 6 1
      // 0ff: iload 9
      // 101: ifeq 0af
      // 104: bipush 0
      // 105: aload 5
      // 107: invokevirtual java/lang/StringBuilder.length ()I
      // 10a: if_icmpeq 111
      // 10d: goto 113
      // 110: athrow
      // 111: aconst_null
      // 112: areturn
      // 113: aload 5
      // 115: invokevirtual java/lang/StringBuilder.toString ()Ljava/lang/String;
      // 118: areturn
      // 119: astore 2
      // 11a: aload 2
      // 11b: new java/lang/StringBuilder
      // 11e: dup
      // 11f: invokespecial java/lang/StringBuilder.<init> ()V
      // 122: getstatic w.z [Ljava/lang/String;
      // 125: bipush 4
      // 126: aaload
      // 127: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 12a: aload 0
      // 12b: ifnull 137
      // 12e: getstatic w.z [Ljava/lang/String;
      // 131: bipush 1
      // 132: aaload
      // 133: goto 13c
      // 136: athrow
      // 137: getstatic w.z [Ljava/lang/String;
      // 13a: bipush 2
      // 13b: aaload
      // 13c: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 13f: bipush 44
      // 141: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 144: iload 1
      // 145: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 148: bipush 41
      // 14a: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 14d: invokevirtual java/lang/StringBuilder.toString ()Ljava/lang/String;
      // 150: invokestatic i.a (Ljava/lang/Throwable;Ljava/lang/String;)Lla;
      // 153: athrow
   }

   static final int a(int var0, int var1, byte[] var2, int var3) {
      boolean var6 = client.vh;

      try {
         n++;
         int var10 = -1;
         int var5 = var3;

         while (true) {
            if (var0 > var5) {
               var10 = wb.q[(var2[var5] ^ var10) & 0xFF] ^ var10 >>> 8;

               try {
                  var5++;
                  if (var6) {
                     break;
                  }

                  if (!var6) {
                     continue;
                  }
               } catch (RuntimeException var8) {
                  throw var8;
               }
            }

            var5 = 123 / ((var1 - 23) / 63);
            var10 = ~var10;
            break;
         }

         return var10;
      } catch (RuntimeException var9) {
         RuntimeException var4 = var9;

         RuntimeException var10000;
         StringBuilder var10001;
         try {
            var10000 = var4;
            var10001 = new StringBuilder().append(z[3]).append(var0).append(',').append(var1).append(',');
            if (var2 != null) {
               throw i.a(var4, var10001.append(z[1]).append((char)44).append(var3).append((char)41).toString());
            }
         } catch (RuntimeException var7) {
            throw var7;
         }

         throw i.a(var10000, var10001.append(z[2]).append((char)44).append(var3).append((char)41).toString());
      }
   }

   static final int a(byte[] var0, int var1, int var2) {
      try {
         d++;
         if (var1 != -1) {
            return 71;
         } else {
            int var7 = 256 * nb.a(255, var0[var2]) + nb.a(255, var0[1 + var2]);

            try {
               if (-32768 <= ~var7) {
                  return var7;
               }
            } catch (RuntimeException var5) {
               throw var5;
            }

            var7 -= 65536;
            return var7;
         }
      } catch (RuntimeException var6) {
         RuntimeException var3 = var6;

         RuntimeException var10000;
         StringBuilder var10001;
         try {
            var10000 = var3;
            var10001 = new StringBuilder().append(z[0]);
            if (var0 != null) {
               throw i.a(var3, var10001.append(z[1]).append((char)44).append(var1).append((char)44).append(var2).append((char)41).toString());
            }
         } catch (RuntimeException var4) {
            throw var4;
         }

         throw i.a(var10000, var10001.append(z[2]).append((char)44).append(var1).append((char)44).append(var2).append((char)41).toString());
      }
   }

   w() {
      this.p = -1;
      this.f = 0;
   }

   private static char[] z(String var0) {
      char[] var10000 = var0.toCharArray();
      if (var10000.length < 2) {
         var10000[0] = (char)(var10000[0] ^ 27);
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
               var10005 = 102;
               break;
            case 1:
               var10005 = 125;
               break;
            case 2:
               var10005 = 122;
               break;
            case 3:
               var10005 = 106;
               break;
            default:
               var10005 = 27;
         }

         var10001[var1] = (char)(var10004 ^ var10005);
      }

      return new String(var10001).intern();
   }
}
