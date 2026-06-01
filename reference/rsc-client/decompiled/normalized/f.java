import java.io.IOException;

final class f {
   static int[] f;
   static int d;
   static int a;
   static String[] e;
   static String[] c = new String[]{
      z(z("c\fh!cV@h.dV\u0012-4xV@c5}Q\u0005\u007f`\u007fU@d4u^\u0013-4\u007f\u0013\u0004h0\u007f@\ty")), z(z("R\u000ei``A\u0005~30V\u000ey%b"))
   };
   static i b = new i(z(z("d)]")), 2);
   private static final String[] z = new String[]{z(z("UNLh")), z(z("]\u0015a,")), z(z("UNOh")), z(z("HN#nm"))};

   static final boolean a(char param0, int param1) {
      // $VF: Couldn't be decompiled
      // Please report this to the Vineflower issue tracker, at https://github.com/Vineflower/vineflower/issues with a copy of the class file (if you have the rights to distribute it!)
      // java.lang.RuntimeException: parsing failure!
      //   at org.jetbrains.java.decompiler.modules.decompiler.decompose.DomHelper.parseGraph(DomHelper.java:211)
      //   at org.jetbrains.java.decompiler.main.rels.MethodProcessor.codeToJava(MethodProcessor.java:166)
      //
      // Bytecode:
      // 00: getstatic client.vh Z
      // 03: istore 5
      // 05: getstatic f.a I
      // 08: bipush 1
      // 09: iadd
      // 0a: putstatic f.a I
      // 0d: iload 0
      // 0e: invokestatic java/lang/Character.isISOControl (C)Z
      // 11: ifne 18
      // 14: goto 1a
      // 17: athrow
      // 18: bipush 0
      // 19: ireturn
      // 1a: iload 0
      // 1b: bipush 115
      // 1d: invokestatic v.a (CI)Z
      // 20: ifne 27
      // 23: goto 29
      // 26: athrow
      // 27: bipush 1
      // 28: ireturn
      // 29: getstatic ga.a [C
      // 2c: astore 2
      // 2d: iload 1
      // 2e: istore 3
      // 2f: iload 3
      // 30: bipush -1
      // 31: ixor
      // 32: aload 2
      // 33: arraylength
      // 34: bipush -1
      // 35: ixor
      // 36: if_icmple 5b
      // 39: aload 2
      // 3a: iload 3
      // 3b: caload
      // 3c: istore 4
      // 3e: iload 0
      // 3f: bipush -1
      // 40: ixor
      // 41: iload 4
      // 43: bipush -1
      // 44: ixor
      // 45: iload 5
      // 47: ifne 64
      // 4a: if_icmpne 53
      // 4d: goto 51
      // 50: athrow
      // 51: bipush 1
      // 52: ireturn
      // 53: iinc 3 1
      // 56: iload 5
      // 58: ifeq 2f
      // 5b: getstatic ac.I [C
      // 5e: astore 2
      // 5f: bipush 0
      // 60: istore 3
      // 61: aload 2
      // 62: arraylength
      // 63: iload 3
      // 64: if_icmple 85
      // 67: aload 2
      // 68: iload 3
      // 69: caload
      // 6a: istore 4
      // 6c: iload 4
      // 6e: iload 5
      // 70: ifne 86
      // 73: iload 0
      // 74: if_icmpne 7d
      // 77: goto 7b
      // 7a: athrow
      // 7b: bipush 1
      // 7c: ireturn
      // 7d: iinc 3 1
      // 80: iload 5
      // 82: ifeq 61
      // 85: bipush 0
      // 86: ireturn
      // 87: astore 2
      // 88: aload 2
      // 89: new java/lang/StringBuilder
      // 8c: dup
      // 8d: invokespecial java/lang/StringBuilder.<init> ()V
      // 90: getstatic f.z [Ljava/lang/String;
      // 93: bipush 0
      // 94: aaload
      // 95: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 98: iload 0
      // 99: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 9c: bipush 44
      // 9e: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // a1: iload 1
      // a2: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // a5: bipush 41
      // a7: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // aa: invokevirtual java/lang/StringBuilder.toString ()Ljava/lang/String;
      // ad: invokestatic i.a (Ljava/lang/Throwable;Ljava/lang/String;)Lla;
      // b0: athrow
   }

   // $VF: Handled exception range with multiple entry points by splitting it
   // $VF: Inserted dummy exception handlers to handle obfuscated exceptions
   static final void a(int var0, tb var1) {
      boolean var5 = client.vh;

      RuntimeException var10000;
      label115: {
         label114: {
            byte[] var2;
            try {
               label120: {
                  d++;
                  var2 = new byte[24];

                  try {
                     if (b.q == null) {
                        break label120;
                     }
                  } catch (Exception var10) {
                     throw var10;
                  }

                  try {
                     b.q.a(0L, var0 + -22592);
                     b.q.a((byte)-123, var2);
                     int var3 = 0;

                     int var21;
                     label104: {
                        while (true) {
                           label101:
                           if (var3 < 24) {
                              label99: {
                                 try {
                                    var18 = -1;
                                    var21 = ~var2[var3];
                                    if (var5) {
                                       break label104;
                                    }

                                    if (-1 == var21) {
                                       break label99;
                                    }
                                 } catch (Exception var14) {
                                    throw var14;
                                 }

                                 try {
                                    try {
                                       if (!var5) {
                                          break label101;
                                       }
                                    } catch (Exception var12) {
                                       throw var12;
                                    }
                                 } catch (Exception var13) {
                                    var10000 = var13;
                                    boolean var22 = false;
                                    break;
                                 }
                              }

                              try {
                                 var3++;
                                 if (!var5) {
                                    continue;
                                 }
                              } catch (Exception var11) {
                                 var10000 = var11;
                                 boolean var23 = false;
                                 break;
                              }
                           }

                           var18 = -25;
                           var21 = ~var3;
                           break label104;
                        }

                        throw var10000;
                     }

                     if (var18 >= var21) {
                        throw new IOException();
                     }
                  } catch (Exception var15) {
                     int var4 = 0;

                     while (-25 < ~var4) {
                        try {
                           var2[var4] = -1;
                           var4++;
                           if (var5) {
                              break label114;
                           }

                           if (var5) {
                              break;
                           }
                        } catch (Exception var9) {
                           throw var9;
                        }
                     }
                  }
               }

               if (var0 != 22607) {
                  return;
               }
            } catch (RuntimeException var16) {
               var10000 = var16;
               boolean var10001 = false;
               break label115;
            }

            try {
               var1.a(0, -126, 24, var2);
            } catch (RuntimeException var8) {
               var10000 = var8;
               boolean var24 = false;
               break label115;
            }
         }

         try {
            return;
         } catch (RuntimeException var7) {
            var10000 = var7;
            boolean var25 = false;
         }
      }

      RuntimeException var17 = var10000;

      StringBuilder var26;
      try {
         var10000 = var17;
         var26 = new StringBuilder().append(z[2]).append(var0).append(',');
         if (var1 != null) {
            throw i.a(var17, var26.append(z[3]).append(')').toString());
         }
      } catch (Exception var6) {
         throw var6;
      }

      throw i.a(var10000, var26.append(z[1]).append(')').toString());
   }

   private static char[] z(String var0) {
      char[] var10000 = var0.toCharArray();
      if (var10000.length < 2) {
         var10000[0] = (char)(var10000[0] ^ 16);
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
               var10005 = 51;
               break;
            case 1:
               var10005 = 96;
               break;
            case 2:
               var10005 = 13;
               break;
            case 3:
               var10005 = 64;
               break;
            default:
               var10005 = 16;
         }

         var10001[var1] = (char)(var10004 ^ var10005);
      }

      return new String(var10001).intern();
   }
}
