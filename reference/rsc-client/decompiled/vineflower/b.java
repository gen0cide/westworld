import java.io.IOException;

class b {
   static int w;
   static int b;
   static int i;
   static int s;
   static int o;
   String B;
   private o z;
   static int t;
   private int g;
   private int A;
   int d = 0;
   static int D;
   static int r;
   static int l;
   static byte[] v;
   private int e;
   static int C;
   static int a;
   static int n;
   static int x;
   static int u;
   private int m;
   static int y;
   private int E;
   static int c = 0;
   static int[] h;
   boolean k;
   static int F;
   ja f;
   static int p;
   private o j;
   static nb q = null;
   private static final String[] G = new String[]{
      z(z("]4R\u0012")),
      z(z("D4:\u0014+")),
      z(z("QoxV")),
      z(z("]4E\u0012")),
      z(z("]4(S8Vn*\u0012\u007f")),
      z(z("]4G\u0012")),
      z(z("]4Y\u0012")),
      z(z("]4P\u0012")),
      z(z("]4@\u0012")),
      z(z("]4X\u0012")),
      z(z("]4^\u0012")),
      z(z("]4Z\u0012")),
      z(z("]4]\u0012")),
      z(z("]4[\u0012")),
      z(z("Ksy_{Po`")),
      z(z("]4D\u0012")),
      z(z("]4_\u0012")),
      z(z("]4V\u0012")),
      z(z("]4U\u0012")),
      z(z("]4F\u0012")),
      z(z("]4S\u0012"))
   };

   final void a(int var1, boolean var2) throws IOException {
      try {
         try {
            if (!var2) {
               this.a(true);
            }
         } catch (RuntimeException var4) {
            throw var4;
         }

         try {
            i++;
            if (this.k) {
               this.f.w = 3;
               this.E = 0;
               this.k = false;
               throw new IOException(this.B);
            }
         } catch (RuntimeException var6) {
            throw var6;
         }

         this.e++;
         if (var1 <= this.e) {
            label33: {
               try {
                  if (-1 <= ~this.E) {
                     break label33;
                  }
               } catch (RuntimeException var5) {
                  throw var5;
               }

               this.e = 0;
               this.a(this.f.F, 0, this.E, (byte)-67);
            }

            this.f.w = 3;
            this.E = 0;
         }
      } catch (RuntimeException var7) {
         throw i.a(var7, G[19] + var1 + ',' + var2 + ')');
      }
   }

   final void a(int var1) throws IOException {
      try {
         try {
            this.b(21294);
            C++;
            this.a(0, true);
            if (var1 != -6924) {
               this.j = (o)null;
            }
         } catch (IOException var3) {
            throw var3;
         }
      } catch (RuntimeException var4) {
         throw i.a(var4, G[15] + var1 + ')');
      }
   }

   int b(boolean var1) throws IOException {
      try {
         try {
            if (!var1) {
               this.g = 126;
            }
         } catch (RuntimeException var3) {
            throw var3;
         }

         p++;
         return 0;
      } catch (RuntimeException var4) {
         throw i.a(var4, G[18] + var1 + 41);
      }
   }

   void a(byte[] var1, int var2, int var3, byte var4) throws IOException {
      try {
         try {
            F++;
            if (var4 != -67) {
               this.a(81, 91);
            }
         } catch (IOException var6) {
            throw var6;
         }
      } catch (RuntimeException var8) {
         RuntimeException var5 = var8;

         RuntimeException var10000;
         StringBuilder var10001;
         try {
            var10000 = var5;
            var10001 = new StringBuilder().append(G[7]);
            if (var1 != null) {
               throw i.a(var5, var10001.append(G[1]).append(',').append(var2).append(',').append(var3).append(',').append((int)var4).append(')').toString());
            }
         } catch (RuntimeException var7) {
            throw var7;
         }

         throw i.a(var10000, var10001.append(G[2]).append(',').append(var2).append(',').append(var3).append(',').append((int)var4).append(')').toString());
      }
   }

   final int a(int var1, int var2) {
      try {
         try {
            if (var1 != 507) {
               this.B = (String)null;
            }
         } catch (RuntimeException var4) {
            throw var4;
         }

         r++;
         return 0xFF & -this.j.c(var1 + -635) + var2;
      } catch (RuntimeException var5) {
         throw i.a(var5, G[16] + var1 + 44 + var2 + 41);
      }
   }

   private final int a(byte[] param1, int param2) {
      // $VF: Couldn't be decompiled
      // Please report this to the Vineflower issue tracker, at https://github.com/Vineflower/vineflower/issues with a copy of the class file (if you have the rights to distribute it!)
      // java.lang.RuntimeException: parsing failure!
      //   at org.jetbrains.java.decompiler.modules.decompiler.decompose.DomHelper.parseGraph(DomHelper.java:211)
      //   at org.jetbrains.java.decompiler.main.rels.MethodProcessor.codeToJava(MethodProcessor.java:166)
      //
      // Bytecode:
      // 000: getstatic b.s I
      // 003: bipush 1
      // 004: iadd
      // 005: putstatic b.s I
      // 008: aload 0
      // 009: dup
      // 00a: getfield b.g I
      // 00d: bipush 1
      // 00e: iadd
      // 00f: putfield b.g I
      // 012: bipush 0
      // 013: aload 0
      // 014: getfield b.d I
      // 017: if_icmpge 04b
      // 01a: aload 0
      // 01b: getfield b.d I
      // 01e: bipush -1
      // 01f: ixor
      // 020: aload 0
      // 021: getfield b.g I
      // 024: bipush -1
      // 025: ixor
      // 026: if_icmpgt 02d
      // 029: goto 04b
      // 02c: athrow
      // 02d: aload 0
      // 02e: bipush 1
      // 02f: putfield b.k Z
      // 032: aload 0
      // 033: getstatic b.G [Ljava/lang/String;
      // 036: bipush 14
      // 038: aaload
      // 039: putfield b.B Ljava/lang/String;
      // 03c: aload 0
      // 03d: dup
      // 03e: getfield b.d I
      // 041: aload 0
      // 042: getfield b.d I
      // 045: iadd
      // 046: putfield b.d I
      // 049: bipush 0
      // 04a: ireturn
      // 04b: aload 0
      // 04c: getfield b.A I
      // 04f: iload 2
      // 050: if_icmpne 093
      // 053: aload 0
      // 054: bipush -124
      // 056: invokevirtual b.b (B)I
      // 059: bipush 2
      // 05a: if_icmpge 065
      // 05d: goto 061
      // 060: athrow
      // 061: goto 093
      // 064: athrow
      // 065: aload 0
      // 066: aload 0
      // 067: bipush 1
      // 068: invokevirtual b.b (Z)I
      // 06b: putfield b.A I
      // 06e: aload 0
      // 06f: getfield b.A I
      // 072: sipush 160
      // 075: if_icmpge 07c
      // 078: goto 093
      // 07b: athrow
      // 07c: aload 0
      // 07d: sipush -160
      // 080: aload 0
      // 081: getfield b.A I
      // 084: iadd
      // 085: sipush 256
      // 088: imul
      // 089: aload 0
      // 08a: bipush 1
      // 08b: invokevirtual b.b (Z)I
      // 08e: ineg
      // 08f: isub
      // 090: putfield b.A I
      // 093: aload 0
      // 094: getfield b.A I
      // 097: bipush -1
      // 098: ixor
      // 099: bipush -1
      // 09a: if_icmpge 10d
      // 09d: aload 0
      // 09e: bipush -124
      // 0a0: invokevirtual b.b (B)I
      // 0a3: aload 0
      // 0a4: getfield b.A I
      // 0a7: if_icmpge 0b2
      // 0aa: goto 0ae
      // 0ad: athrow
      // 0ae: goto 10d
      // 0b1: athrow
      // 0b2: sipush 160
      // 0b5: aload 0
      // 0b6: getfield b.A I
      // 0b9: if_icmple 0ed
      // 0bc: aload 1
      // 0bd: bipush -1
      // 0be: aload 0
      // 0bf: getfield b.A I
      // 0c2: iadd
      // 0c3: aload 0
      // 0c4: bipush 1
      // 0c5: invokevirtual b.b (Z)I
      // 0c8: i2b
      // 0c9: bastore
      // 0ca: bipush 1
      // 0cb: aload 0
      // 0cc: getfield b.A I
      // 0cf: if_icmplt 0da
      // 0d2: goto 0d6
      // 0d5: athrow
      // 0d6: goto 0fc
      // 0d9: athrow
      // 0da: aload 0
      // 0db: aload 1
      // 0dc: bipush 126
      // 0de: bipush -1
      // 0df: aload 0
      // 0e0: getfield b.A I
      // 0e3: iadd
      // 0e4: invokespecial b.a ([BBI)V
      // 0e7: getstatic client.vh Z
      // 0ea: ifeq 0fc
      // 0ed: aload 0
      // 0ee: aload 1
      // 0ef: bipush 64
      // 0f1: aload 0
      // 0f2: getfield b.A I
      // 0f5: invokespecial b.a ([BBI)V
      // 0f8: goto 0fc
      // 0fb: athrow
      // 0fc: aload 0
      // 0fd: getfield b.A I
      // 100: istore 3
      // 101: aload 0
      // 102: bipush 0
      // 103: putfield b.g I
      // 106: aload 0
      // 107: bipush 0
      // 108: putfield b.A I
      // 10b: iload 3
      // 10c: ireturn
      // 10d: goto 11e
      // 110: astore 3
      // 111: aload 0
      // 112: bipush 1
      // 113: putfield b.k Z
      // 116: aload 0
      // 117: aload 3
      // 118: invokevirtual java/io/IOException.getMessage ()Ljava/lang/String;
      // 11b: putfield b.B Ljava/lang/String;
      // 11e: bipush 0
      // 11f: ireturn
      // 120: astore 3
      // 121: aload 3
      // 122: new java/lang/StringBuilder
      // 125: dup
      // 126: invokespecial java/lang/StringBuilder.<init> ()V
      // 129: getstatic b.G [Ljava/lang/String;
      // 12c: bipush 13
      // 12e: aaload
      // 12f: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 132: aload 1
      // 133: ifnull 13f
      // 136: getstatic b.G [Ljava/lang/String;
      // 139: bipush 1
      // 13a: aaload
      // 13b: goto 144
      // 13e: athrow
      // 13f: getstatic b.G [Ljava/lang/String;
      // 142: bipush 2
      // 143: aaload
      // 144: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 147: bipush 44
      // 149: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 14c: iload 2
      // 14d: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 150: bipush 41
      // 152: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 155: invokevirtual java/lang/StringBuilder.toString ()Ljava/lang/String;
      // 158: invokestatic i.a (Ljava/lang/Throwable;Ljava/lang/String;)Lla;
      // 15b: athrow
   }

   int b(byte var1) throws IOException {
      try {
         D++;
         int var2 = -6 % ((var1 - -64) / 56);
         return 0;
      } catch (RuntimeException var3) {
         throw i.a(var3, G[17] + var1 + 41);
      }
   }

   private final void a(byte[] param1, byte param2, int param3) throws IOException {
      // $VF: Couldn't be decompiled
      // Please report this to the Vineflower issue tracker, at https://github.com/Vineflower/vineflower/issues with a copy of the class file (if you have the rights to distribute it!)
      // java.lang.RuntimeException: parsing failure!
      //   at org.jetbrains.java.decompiler.modules.decompiler.decompose.DomHelper.parseGraph(DomHelper.java:211)
      //   at org.jetbrains.java.decompiler.main.rels.MethodProcessor.codeToJava(MethodProcessor.java:166)
      //
      // Bytecode:
      // 00: iload 2
      // 01: bipush 51
      // 03: if_icmpge 08
      // 06: return
      // 07: athrow
      // 08: getstatic b.l I
      // 0b: bipush 1
      // 0c: iadd
      // 0d: putstatic b.l I
      // 10: aload 0
      // 11: aload 1
      // 12: iload 3
      // 13: bipush 0
      // 14: bipush -112
      // 16: invokevirtual b.a ([BIII)V
      // 19: goto 63
      // 1c: astore 4
      // 1e: aload 4
      // 20: new java/lang/StringBuilder
      // 23: dup
      // 24: invokespecial java/lang/StringBuilder.<init> ()V
      // 27: getstatic b.G [Ljava/lang/String;
      // 2a: bipush 6
      // 2c: aaload
      // 2d: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 30: aload 1
      // 31: ifnull 3d
      // 34: getstatic b.G [Ljava/lang/String;
      // 37: bipush 1
      // 38: aaload
      // 39: goto 42
      // 3c: athrow
      // 3d: getstatic b.G [Ljava/lang/String;
      // 40: bipush 2
      // 41: aaload
      // 42: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 45: bipush 44
      // 47: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 4a: iload 2
      // 4b: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 4e: bipush 44
      // 50: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 53: iload 3
      // 54: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 57: bipush 41
      // 59: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 5c: invokevirtual java/lang/StringBuilder.toString ()Ljava/lang/String;
      // 5f: invokestatic i.a (Ljava/lang/Throwable;Ljava/lang/String;)Lla;
      // 62: athrow
      // 63: return
   }

   final boolean a(byte var1) {
      try {
         int var2 = -119 / ((-44 - var1) / 53);

         try {
            o++;
            if (0 < this.E) {
               return true;
            }
         } catch (RuntimeException var3) {
            throw var3;
         }

         return false;
      } catch (RuntimeException var4) {
         throw i.a(var4, G[5] + var1 + ')');
      }
   }

   final void b(int var1, int var2) {
      try {
         label29: {
            IOException var3;
            try {
               w++;
               if (~this.E >= ~(4 * this.m / 5)) {
                  break label29;
               }

               try {
                  this.a(0, true);
                  break label29;
               } catch (IOException var4) {
                  var3 = var4;
               }
            } catch (RuntimeException var5) {
               throw var5;
            }

            this.k = true;
            this.B = var3.getMessage();
         }

         this.f.w = this.E - -2;
         if (var2 == 0) {
            this.f.c(var1, 82);
         }
      } catch (RuntimeException var6) {
         throw i.a(var6, G[11] + var1 + ',' + var2 + ')');
      }
   }

   final int a(int var1, ja var2) {
      try {
         b++;
         var2.w = var1;
         return this.a(var2.F, var1 ^ 0);
      } catch (RuntimeException var5) {
         RuntimeException var3 = var5;

         RuntimeException var10000;
         StringBuilder var10001;
         try {
            var10000 = var3;
            var10001 = new StringBuilder().append(G[3]).append(var1).append(',');
            if (var2 != null) {
               throw i.a(var3, var10001.append(G[1]).append((char)41).toString());
            }
         } catch (RuntimeException var4) {
            throw var4;
         }

         throw i.a(var10000, var10001.append(G[2]).append((char)41).toString());
      }
   }

   void a(boolean var1) {
      try {
         n++;
         if (!var1) {
            this.a(116, (ja)null);
         }
      } catch (RuntimeException var3) {
         throw i.a(var3, G[20] + var1 + ')');
      }
   }

   final void a(byte var1, int[] var2) {
      try {
         try {
            if (var1 >= -68) {
               this.d = -84;
            }
         } catch (RuntimeException var5) {
            throw var5;
         }

         this.j = new o(var2);
         u++;
         this.z = new o(var2);
      } catch (RuntimeException var6) {
         RuntimeException var3 = var6;

         RuntimeException var10000;
         StringBuilder var10001;
         try {
            var10000 = var3;
            var10001 = new StringBuilder().append(G[12]).append((int)var1).append(',');
            if (var2 != null) {
               throw i.a(var3, var10001.append(G[1]).append(')').toString());
            }
         } catch (RuntimeException var4) {
            throw var4;
         }

         throw i.a(var10000, var10001.append(G[2]).append(')').toString());
      }
   }

   void a(byte[] var1, int var2, int var3, int var4) throws IOException {
      try {
         int var8 = -61 / ((-25 - var4) / 35);
         y++;
      } catch (RuntimeException var7) {
         RuntimeException var5 = var7;

         RuntimeException var10000;
         StringBuilder var10001;
         try {
            var10000 = var5;
            var10001 = new StringBuilder().append(G[0]);
            if (var1 != null) {
               throw i.a(var5, var10001.append(G[1]).append(',').append(var2).append(',').append(var3).append(',').append(var4).append(')').toString());
            }
         } catch (RuntimeException var6) {
            throw var6;
         }

         throw i.a(var10000, var10001.append(G[2]).append(',').append(var2).append(',').append(var3).append(',').append(var4).append(')').toString());
      }
   }

   final void b(int param1) {
      // $VF: Couldn't be decompiled
      // Please report this to the Vineflower issue tracker, at https://github.com/Vineflower/vineflower/issues with a copy of the class file (if you have the rights to distribute it!)
      // java.lang.RuntimeException: parsing failure!
      //   at org.jetbrains.java.decompiler.modules.decompiler.decompose.DomHelper.parseGraph(DomHelper.java:211)
      //   at org.jetbrains.java.decompiler.main.rels.MethodProcessor.codeToJava(MethodProcessor.java:166)
      //
      // Bytecode:
      // 000: aload 0
      // 001: getfield b.z Lo;
      // 004: ifnull 034
      // 007: sipush 255
      // 00a: aload 0
      // 00b: getfield b.f Lja;
      // 00e: getfield ja.F [B
      // 011: aload 0
      // 012: getfield b.E I
      // 015: bipush 2
      // 016: iadd
      // 017: baload
      // 018: iand
      // 019: istore 2
      // 01a: aload 0
      // 01b: getfield b.f Lja;
      // 01e: getfield ja.F [B
      // 021: bipush 2
      // 022: aload 0
      // 023: getfield b.E I
      // 026: iadd
      // 027: aload 0
      // 028: getfield b.z Lo;
      // 02b: bipush -83
      // 02d: invokevirtual o.c (I)I
      // 030: iload 2
      // 031: iadd
      // 032: i2b
      // 033: bastore
      // 034: getstatic b.x I
      // 037: bipush 1
      // 038: iadd
      // 039: putstatic b.x I
      // 03c: aload 0
      // 03d: getfield b.f Lja;
      // 040: getfield ja.w I
      // 043: aload 0
      // 044: getfield b.E I
      // 047: ineg
      // 048: iadd
      // 049: bipush 2
      // 04a: isub
      // 04b: istore 2
      // 04c: sipush -161
      // 04f: iload 2
      // 050: bipush -1
      // 051: ixor
      // 052: if_icmplt 08b
      // 055: aload 0
      // 056: getfield b.f Lja;
      // 059: getfield ja.F [B
      // 05c: aload 0
      // 05d: getfield b.E I
      // 060: sipush 160
      // 063: iload 2
      // 064: sipush 256
      // 067: idiv
      // 068: iadd
      // 069: i2b
      // 06a: bastore
      // 06b: aload 0
      // 06c: getfield b.f Lja;
      // 06f: getfield ja.F [B
      // 072: bipush 1
      // 073: aload 0
      // 074: getfield b.E I
      // 077: iadd
      // 078: iload 2
      // 079: sipush 255
      // 07c: invokestatic ib.a (II)I
      // 07f: i2b
      // 080: bastore
      // 081: getstatic client.vh Z
      // 084: ifeq 0c7
      // 087: goto 08b
      // 08a: athrow
      // 08b: aload 0
      // 08c: getfield b.f Lja;
      // 08f: getfield ja.F [B
      // 092: aload 0
      // 093: getfield b.E I
      // 096: iload 2
      // 097: i2b
      // 098: bastore
      // 099: aload 0
      // 09a: getfield b.f Lja;
      // 09d: dup
      // 09e: getfield ja.w I
      // 0a1: bipush 1
      // 0a2: isub
      // 0a3: putfield ja.w I
      // 0a6: aload 0
      // 0a7: getfield b.f Lja;
      // 0aa: getfield ja.F [B
      // 0ad: bipush 1
      // 0ae: aload 0
      // 0af: getfield b.E I
      // 0b2: iadd
      // 0b3: aload 0
      // 0b4: getfield b.f Lja;
      // 0b7: getfield ja.F [B
      // 0ba: aload 0
      // 0bb: getfield b.f Lja;
      // 0be: getfield ja.w I
      // 0c1: baload
      // 0c2: bastore
      // 0c3: goto 0c7
      // 0c6: athrow
      // 0c7: sipush -10001
      // 0ca: aload 0
      // 0cb: getfield b.m I
      // 0ce: bipush -1
      // 0cf: ixor
      // 0d0: if_icmple 0d7
      // 0d3: goto 108
      // 0d6: athrow
      // 0d7: aload 0
      // 0d8: getfield b.f Lja;
      // 0db: getfield ja.F [B
      // 0de: aload 0
      // 0df: getfield b.E I
      // 0e2: bipush 2
      // 0e3: iadd
      // 0e4: baload
      // 0e5: sipush 255
      // 0e8: iand
      // 0e9: istore 3
      // 0ea: getstatic ia.d [I
      // 0ed: iload 3
      // 0ee: dup2
      // 0ef: iaload
      // 0f0: bipush 1
      // 0f1: iadd
      // 0f2: iastore
      // 0f3: getstatic m.i [I
      // 0f6: iload 3
      // 0f7: dup2
      // 0f8: iaload
      // 0f9: aload 0
      // 0fa: getfield b.E I
      // 0fd: ineg
      // 0fe: aload 0
      // 0ff: getfield b.f Lja;
      // 102: getfield ja.w I
      // 105: iadd
      // 106: iadd
      // 107: iastore
      // 108: iload 1
      // 109: sipush 21294
      // 10c: if_icmpeq 118
      // 10f: bipush -78
      // 111: putstatic b.c I
      // 114: goto 118
      // 117: athrow
      // 118: aload 0
      // 119: aload 0
      // 11a: getfield b.f Lja;
      // 11d: getfield ja.w I
      // 120: putfield b.E I
      // 123: goto 148
      // 126: astore 2
      // 127: aload 2
      // 128: new java/lang/StringBuilder
      // 12b: dup
      // 12c: invokespecial java/lang/StringBuilder.<init> ()V
      // 12f: getstatic b.G [Ljava/lang/String;
      // 132: bipush 9
      // 134: aaload
      // 135: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 138: iload 1
      // 139: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 13c: bipush 41
      // 13e: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 141: invokevirtual java/lang/StringBuilder.toString ()Ljava/lang/String;
      // 144: invokestatic i.a (Ljava/lang/Throwable;Ljava/lang/String;)Lla;
      // 147: athrow
      // 148: return
   }

   static final String a(int param0, byte param1, String param2) {
      // $VF: Couldn't be decompiled
      // Please report this to the Vineflower issue tracker, at https://github.com/Vineflower/vineflower/issues with a copy of the class file (if you have the rights to distribute it!)
      // java.lang.RuntimeException: parsing failure!
      //   at org.jetbrains.java.decompiler.modules.decompiler.decompose.DomHelper.parseGraph(DomHelper.java:211)
      //   at org.jetbrains.java.decompiler.main.rels.MethodProcessor.codeToJava(MethodProcessor.java:166)
      //
      // Bytecode:
      // 000: getstatic client.vh Z
      // 003: istore 6
      // 005: getstatic b.t I
      // 008: bipush 1
      // 009: iadd
      // 00a: putstatic b.t I
      // 00d: ldc ""
      // 00f: astore 3
      // 010: iload 1
      // 011: bipush -5
      // 013: if_icmpeq 025
      // 016: aconst_null
      // 017: checkcast tb
      // 01a: bipush -63
      // 01c: bipush 17
      // 01e: invokestatic b.a (Ltb;II)V
      // 021: goto 025
      // 024: athrow
      // 025: bipush 0
      // 026: istore 4
      // 028: iload 4
      // 02a: iload 0
      // 02b: if_icmpge 109
      // 02e: iload 4
      // 030: bipush -1
      // 031: ixor
      // 032: aload 2
      // 033: invokevirtual java/lang/String.length ()I
      // 036: bipush -1
      // 037: ixor
      // 038: if_icmple 0ed
      // 03b: aload 2
      // 03c: iload 4
      // 03e: invokevirtual java/lang/String.charAt (I)C
      // 041: istore 5
      // 043: bipush -98
      // 045: iload 5
      // 047: bipush -1
      // 048: ixor
      // 049: if_icmplt 070
      // 04c: iload 5
      // 04e: bipush 122
      // 050: if_icmpgt 070
      // 053: goto 057
      // 056: athrow
      // 057: new java/lang/StringBuilder
      // 05a: dup
      // 05b: invokespecial java/lang/StringBuilder.<init> ()V
      // 05e: aload 3
      // 05f: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 062: iload 5
      // 064: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 067: invokevirtual java/lang/StringBuilder.toString ()Ljava/lang/String;
      // 06a: astore 3
      // 06b: iload 6
      // 06d: ifeq 0e8
      // 070: bipush 65
      // 072: iload 5
      // 074: if_icmpgt 0a1
      // 077: goto 07b
      // 07a: athrow
      // 07b: bipush -91
      // 07d: iload 5
      // 07f: bipush -1
      // 080: ixor
      // 081: if_icmpgt 0a1
      // 084: goto 088
      // 087: athrow
      // 088: new java/lang/StringBuilder
      // 08b: dup
      // 08c: invokespecial java/lang/StringBuilder.<init> ()V
      // 08f: aload 3
      // 090: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 093: iload 5
      // 095: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 098: invokevirtual java/lang/StringBuilder.toString ()Ljava/lang/String;
      // 09b: astore 3
      // 09c: iload 6
      // 09e: ifeq 0e8
      // 0a1: iload 5
      // 0a3: bipush -1
      // 0a4: ixor
      // 0a5: bipush -49
      // 0a7: if_icmpgt 0d4
      // 0aa: goto 0ae
      // 0ad: athrow
      // 0ae: iload 5
      // 0b0: bipush -1
      // 0b1: ixor
      // 0b2: bipush -58
      // 0b4: if_icmplt 0d4
      // 0b7: goto 0bb
      // 0ba: athrow
      // 0bb: new java/lang/StringBuilder
      // 0be: dup
      // 0bf: invokespecial java/lang/StringBuilder.<init> ()V
      // 0c2: aload 3
      // 0c3: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 0c6: iload 5
      // 0c8: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 0cb: invokevirtual java/lang/StringBuilder.toString ()Ljava/lang/String;
      // 0ce: astore 3
      // 0cf: iload 6
      // 0d1: ifeq 0e8
      // 0d4: new java/lang/StringBuilder
      // 0d7: dup
      // 0d8: invokespecial java/lang/StringBuilder.<init> ()V
      // 0db: aload 3
      // 0dc: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 0df: bipush 95
      // 0e1: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 0e4: invokevirtual java/lang/StringBuilder.toString ()Ljava/lang/String;
      // 0e7: astore 3
      // 0e8: iload 6
      // 0ea: ifeq 101
      // 0ed: new java/lang/StringBuilder
      // 0f0: dup
      // 0f1: invokespecial java/lang/StringBuilder.<init> ()V
      // 0f4: aload 3
      // 0f5: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 0f8: ldc " "
      // 0fa: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 0fd: invokevirtual java/lang/StringBuilder.toString ()Ljava/lang/String;
      // 100: astore 3
      // 101: iinc 4 1
      // 104: iload 6
      // 106: ifeq 028
      // 109: aload 3
      // 10a: areturn
      // 10b: astore 3
      // 10c: aload 3
      // 10d: new java/lang/StringBuilder
      // 110: dup
      // 111: invokespecial java/lang/StringBuilder.<init> ()V
      // 114: getstatic b.G [Ljava/lang/String;
      // 117: bipush 10
      // 119: aaload
      // 11a: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 11d: iload 0
      // 11e: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 121: bipush 44
      // 123: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 126: iload 1
      // 127: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 12a: bipush 44
      // 12c: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 12f: aload 2
      // 130: ifnull 13c
      // 133: getstatic b.G [Ljava/lang/String;
      // 136: bipush 1
      // 137: aaload
      // 138: goto 141
      // 13b: athrow
      // 13c: getstatic b.G [Ljava/lang/String;
      // 13f: bipush 2
      // 140: aaload
      // 141: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 144: bipush 41
      // 146: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 149: invokevirtual java/lang/StringBuilder.toString ()Ljava/lang/String;
      // 14c: invokestatic i.a (Ljava/lang/Throwable;Ljava/lang/String;)Lla;
      // 14f: athrow
   }

   static final void a(tb var0, int var1, int var2) {
      try {
         if (null != q) {
            try {
               q.a(0L, var1 ^ -26747);
               q.a(24, -107, var2, var0.F);
            } catch (Exception var4) {
            }
         }

         try {
            if (var1 != 26628) {
               v = (byte[])null;
            }
         } catch (Exception var6) {
            throw var6;
         }

         a++;
      } catch (RuntimeException var7) {
         RuntimeException var3 = var7;

         RuntimeException var10000;
         StringBuilder var10001;
         try {
            var10000 = var3;
            var10001 = new StringBuilder().append(G[8]);
            if (var0 != null) {
               throw i.a(var3, var10001.append(G[1]).append(',').append(var1).append(',').append(var2).append(')').toString());
            }
         } catch (Exception var5) {
            throw var5;
         }

         throw i.a(var10000, var10001.append(G[2]).append(',').append(var1).append(',').append(var2).append(')').toString());
      }
   }

   protected b() {
      this.B = "";
      this.e = 0;
      this.m = 5000;
      this.k = false;
      this.A = 0;
      this.E = 0;
      this.g = 0;

      try {
         this.f = new ja(this.m);
         this.f.w = 3;
      } catch (RuntimeException var2) {
         throw i.a(var2, G[4]);
      }
   }

   private static char[] z(String var0) {
      char[] var10000 = var0.toCharArray();
      if (var10000.length < 2) {
         var10000[0] = (char)(var10000[0] ^ 'V');
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
               var10005 = 63;
               break;
            case 1:
               var10005 = 26;
               break;
            case 2:
               var10005 = 20;
               break;
            case 3:
               var10005 = 58;
               break;
            default:
               var10005 = 86;
         }

         var10001[var1] = (char)(var10004 ^ var10005);
      }

      return new String(var10001).intern();
   }
}
