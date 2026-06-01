import java.awt.Component;

class sa {
   static int t;
   private static eb n;
   private long u;
   int[] j;
   private static int o;
   static boolean i;
   private boolean b = false;
   private int r = 32;
   private va a;
   private int g;
   private int q;
   private int c;
   private int k;
   private va[] s;
   private long e;
   private int p;
   private int m;
   private va[] d;
   private long f;
   private int h;
   private boolean l;

   int b() throws Exception {
      return this.k;
   }

   void b(int var1) throws Exception {
   }

   void c() throws Exception {
   }

   final synchronized void d() {
      // $VF: Couldn't be decompiled
      // Please report this to the Vineflower issue tracker, at https://github.com/Vineflower/vineflower/issues with a copy of the class file (if you have the rights to distribute it!)
      // java.lang.RuntimeException: parsing failure!
      //   at org.jetbrains.java.decompiler.modules.decompiler.decompose.DomHelper.parseGraph(DomHelper.java:211)
      //   at org.jetbrains.java.decompiler.main.rels.MethodProcessor.codeToJava(MethodProcessor.java:166)
      //
      // Bytecode:
      // 00: getstatic sa.n Leb;
      // 03: ifnull 68
      // 06: bipush 1
      // 07: istore 1
      // 08: bipush 0
      // 09: istore 2
      // 0a: iload 2
      // 0b: bipush 2
      // 0c: if_icmpge 3f
      // 0f: getstatic sa.n Leb;
      // 12: getfield eb.f [Lsa;
      // 15: iload 2
      // 16: aaload
      // 17: aload 0
      // 18: if_acmpne 2c
      // 1b: goto 1f
      // 1e: athrow
      // 1f: getstatic sa.n Leb;
      // 22: getfield eb.f [Lsa;
      // 25: iload 2
      // 26: aconst_null
      // 27: aastore
      // 28: goto 2c
      // 2b: athrow
      // 2c: getstatic sa.n Leb;
      // 2f: getfield eb.f [Lsa;
      // 32: iload 2
      // 33: aaload
      // 34: ifnull 39
      // 37: bipush 0
      // 38: istore 1
      // 39: iinc 2 1
      // 3c: goto 0a
      // 3f: iload 1
      // 40: ifeq 68
      // 43: getstatic sa.n Leb;
      // 46: bipush 1
      // 47: putfield eb.a Z
      // 4a: goto 4e
      // 4d: athrow
      // 4e: getstatic sa.n Leb;
      // 51: getfield eb.i Z
      // 54: ifeq 64
      // 57: sipush 11200
      // 5a: ldc2_w 50
      // 5d: invokestatic mb.a (IJ)V
      // 60: goto 4e
      // 63: athrow
      // 64: aconst_null
      // 65: putstatic sa.n Leb;
      // 68: aload 0
      // 69: invokevirtual sa.e ()V
      // 6c: aload 0
      // 6d: aconst_null
      // 6e: putfield sa.j [I
      // 71: aload 0
      // 72: bipush 1
      // 73: putfield sa.b Z
      // 76: return
   }

   void e() {
   }

   private final void a(va var1, int var2) {
      int var3 = var2 >> 5;
      va var4 = this.s[var3];

      label17: {
         try {
            if (var4 == null) {
               this.d[var3] = var1;
               break label17;
            }
         } catch (IllegalStateException var5) {
            throw var5;
         }

         var4.j = var1;
      }

      this.s[var3] = var1;
      var1.i = var2;
   }

   static final sa a(c param0, Component param1, int param2, int param3) {
      // $VF: Couldn't be decompiled
      // Please report this to the Vineflower issue tracker, at https://github.com/Vineflower/vineflower/issues with a copy of the class file (if you have the rights to distribute it!)
      // java.lang.RuntimeException: parsing failure!
      //   at org.jetbrains.java.decompiler.modules.decompiler.decompose.DomHelper.parseGraph(DomHelper.java:211)
      //   at org.jetbrains.java.decompiler.main.rels.MethodProcessor.codeToJava(MethodProcessor.java:166)
      //
      // Bytecode:
      // 00: getstatic sa.t I
      // 03: ifne 0f
      // 06: new java/lang/IllegalStateException
      // 09: dup
      // 0a: invokespecial java/lang/IllegalStateException.<init> ()V
      // 0d: athrow
      // 0e: athrow
      // 0f: iload 2
      // 10: iflt 1c
      // 13: iload 2
      // 14: bipush 2
      // 15: if_icmplt 25
      // 18: goto 1c
      // 1b: athrow
      // 1c: new java/lang/IllegalArgumentException
      // 1f: dup
      // 20: invokespecial java/lang/IllegalArgumentException.<init> ()V
      // 23: athrow
      // 24: athrow
      // 25: iload 3
      // 26: sipush 256
      // 29: if_icmpge 30
      // 2c: sipush 256
      // 2f: istore 3
      // 30: new pb
      // 33: dup
      // 34: invokespecial pb.<init> ()V
      // 37: astore 4
      // 39: aload 4
      // 3b: sipush 256
      // 3e: getstatic sa.i Z
      // 41: ifeq 49
      // 44: bipush 2
      // 45: goto 4a
      // 48: athrow
      // 49: bipush 1
      // 4a: imul
      // 4b: newarray 10
      // 4d: putfield sa.j [I
      // 50: aload 4
      // 52: iload 3
      // 53: putfield sa.q I
      // 56: aload 4
      // 58: aload 1
      // 59: invokevirtual sa.a (Ljava/awt/Component;)V
      // 5c: aload 4
      // 5e: iload 3
      // 5f: sipush -1024
      // 62: iand
      // 63: sipush 1024
      // 66: iadd
      // 67: putfield sa.k I
      // 6a: aload 4
      // 6c: getfield sa.k I
      // 6f: sipush 16384
      // 72: if_icmple 81
      // 75: aload 4
      // 77: sipush 16384
      // 7a: putfield sa.k I
      // 7d: goto 81
      // 80: athrow
      // 81: aload 4
      // 83: aload 4
      // 85: getfield sa.k I
      // 88: invokevirtual sa.b (I)V
      // 8b: getstatic sa.o I
      // 8e: ifle bc
      // 91: getstatic sa.n Leb;
      // 94: ifnonnull bc
      // 97: goto 9b
      // 9a: athrow
      // 9b: new eb
      // 9e: dup
      // 9f: invokespecial eb.<init> ()V
      // a2: putstatic sa.n Leb;
      // a5: getstatic sa.n Leb;
      // a8: aload 0
      // a9: putfield eb.g Lc;
      // ac: aload 0
      // ad: bipush 1
      // ae: getstatic sa.n Leb;
      // b1: getstatic sa.o I
      // b4: invokevirtual c.a (ZLjava/lang/Runnable;I)Lg;
      // b7: pop
      // b8: goto bc
      // bb: athrow
      // bc: getstatic sa.n Leb;
      // bf: ifnull e4
      // c2: getstatic sa.n Leb;
      // c5: getfield eb.f [Lsa;
      // c8: iload 2
      // c9: aaload
      // ca: ifnull da
      // cd: goto d1
      // d0: athrow
      // d1: new java/lang/IllegalArgumentException
      // d4: dup
      // d5: invokespecial java/lang/IllegalArgumentException.<init> ()V
      // d8: athrow
      // d9: athrow
      // da: getstatic sa.n Leb;
      // dd: getfield eb.f [Lsa;
      // e0: iload 2
      // e1: aload 4
      // e3: aastore
      // e4: aload 4
      // e6: areturn
      // e7: astore 4
      // e9: new sa
      // ec: dup
      // ed: invokespecial sa.<init> ()V
      // f0: areturn
   }

   final synchronized void a() {
      // $VF: Couldn't be decompiled
      // Please report this to the Vineflower issue tracker, at https://github.com/Vineflower/vineflower/issues with a copy of the class file (if you have the rights to distribute it!)
      // java.lang.RuntimeException: parsing failure!
      //   at org.jetbrains.java.decompiler.modules.decompiler.decompose.DomHelper.parseGraph(DomHelper.java:211)
      //   at org.jetbrains.java.decompiler.main.rels.MethodProcessor.codeToJava(MethodProcessor.java:166)
      //
      // Bytecode:
      // 000: aload 0
      // 001: getfield sa.b Z
      // 004: ifeq 009
      // 007: return
      // 008: athrow
      // 009: bipush 0
      // 00a: invokestatic p.a (I)J
      // 00d: lstore 1
      // 00e: lload 1
      // 00f: aload 0
      // 010: getfield sa.u J
      // 013: ldc2_w 6000
      // 016: ladd
      // 017: lcmp
      // 018: ifle 028
      // 01b: aload 0
      // 01c: lload 1
      // 01d: ldc2_w 6000
      // 020: lsub
      // 021: putfield sa.u J
      // 024: goto 028
      // 027: athrow
      // 028: lload 1
      // 029: aload 0
      // 02a: getfield sa.u J
      // 02d: ldc2_w 5000
      // 030: ladd
      // 031: lcmp
      // 032: ifle 054
      // 035: aload 0
      // 036: sipush 256
      // 039: invokespecial sa.a (I)V
      // 03c: aload 0
      // 03d: dup
      // 03e: getfield sa.u J
      // 041: ldc 256000
      // 043: getstatic sa.t I
      // 046: idiv
      // 047: i2l
      // 048: ladd
      // 049: putfield sa.u J
      // 04c: bipush 0
      // 04d: invokestatic p.a (I)J
      // 050: lstore 1
      // 051: goto 028
      // 054: goto 05d
      // 057: astore 3
      // 058: aload 0
      // 059: lload 1
      // 05a: putfield sa.u J
      // 05d: aload 0
      // 05e: getfield sa.j [I
      // 061: ifnonnull 066
      // 064: return
      // 065: athrow
      // 066: aload 0
      // 067: getfield sa.e J
      // 06a: lconst_0
      // 06b: lcmp
      // 06c: ifeq 08b
      // 06f: lload 1
      // 070: aload 0
      // 071: getfield sa.e J
      // 074: lcmp
      // 075: ifge 079
      // 078: return
      // 079: aload 0
      // 07a: aload 0
      // 07b: getfield sa.k I
      // 07e: invokevirtual sa.b (I)V
      // 081: aload 0
      // 082: lconst_0
      // 083: putfield sa.e J
      // 086: aload 0
      // 087: bipush 1
      // 088: putfield sa.l Z
      // 08b: aload 0
      // 08c: invokevirtual sa.b ()I
      // 08f: istore 3
      // 090: aload 0
      // 091: getfield sa.h I
      // 094: iload 3
      // 095: isub
      // 096: aload 0
      // 097: getfield sa.m I
      // 09a: if_icmple 0ab
      // 09d: aload 0
      // 09e: aload 0
      // 09f: getfield sa.h I
      // 0a2: iload 3
      // 0a3: isub
      // 0a4: putfield sa.m I
      // 0a7: goto 0ab
      // 0aa: athrow
      // 0ab: aload 0
      // 0ac: getfield sa.q I
      // 0af: aload 0
      // 0b0: getfield sa.g I
      // 0b3: iadd
      // 0b4: istore 4
      // 0b6: iload 4
      // 0b8: sipush 256
      // 0bb: iadd
      // 0bc: sipush 16384
      // 0bf: if_icmple 0c7
      // 0c2: sipush 16128
      // 0c5: istore 4
      // 0c7: iload 4
      // 0c9: sipush 256
      // 0cc: iadd
      // 0cd: aload 0
      // 0ce: getfield sa.k I
      // 0d1: if_icmple 12e
      // 0d4: aload 0
      // 0d5: dup
      // 0d6: getfield sa.k I
      // 0d9: sipush 1024
      // 0dc: iadd
      // 0dd: putfield sa.k I
      // 0e0: aload 0
      // 0e1: getfield sa.k I
      // 0e4: sipush 16384
      // 0e7: if_icmple 0f9
      // 0ea: goto 0ee
      // 0ed: athrow
      // 0ee: aload 0
      // 0ef: sipush 16384
      // 0f2: putfield sa.k I
      // 0f5: goto 0f9
      // 0f8: athrow
      // 0f9: aload 0
      // 0fa: invokevirtual sa.e ()V
      // 0fd: aload 0
      // 0fe: aload 0
      // 0ff: getfield sa.k I
      // 102: invokevirtual sa.b (I)V
      // 105: bipush 0
      // 106: istore 3
      // 107: aload 0
      // 108: bipush 1
      // 109: putfield sa.l Z
      // 10c: iload 4
      // 10e: sipush 256
      // 111: iadd
      // 112: aload 0
      // 113: getfield sa.k I
      // 116: if_icmple 12e
      // 119: aload 0
      // 11a: getfield sa.k I
      // 11d: sipush 256
      // 120: isub
      // 121: istore 4
      // 123: aload 0
      // 124: iload 4
      // 126: aload 0
      // 127: getfield sa.q I
      // 12a: isub
      // 12b: putfield sa.g I
      // 12e: iload 3
      // 12f: iload 4
      // 131: if_icmpge 14d
      // 134: aload 0
      // 135: aload 0
      // 136: getfield sa.j [I
      // 139: sipush 256
      // 13c: invokespecial sa.a ([II)V
      // 13f: aload 0
      // 140: invokevirtual sa.c ()V
      // 143: wide iinc 3 256
      // 149: goto 12e
      // 14c: athrow
      // 14d: lload 1
      // 14e: aload 0
      // 14f: getfield sa.f J
      // 152: lcmp
      // 153: ifle 1b2
      // 156: aload 0
      // 157: getfield sa.l Z
      // 15a: ifne 19f
      // 15d: goto 161
      // 160: athrow
      // 161: aload 0
      // 162: getfield sa.m I
      // 165: ifne 185
      // 168: goto 16c
      // 16b: athrow
      // 16c: aload 0
      // 16d: getfield sa.c I
      // 170: ifne 185
      // 173: goto 177
      // 176: athrow
      // 177: aload 0
      // 178: invokevirtual sa.e ()V
      // 17b: aload 0
      // 17c: lload 1
      // 17d: ldc2_w 2000
      // 180: ladd
      // 181: putfield sa.e J
      // 184: return
      // 185: aload 0
      // 186: aload 0
      // 187: getfield sa.c I
      // 18a: aload 0
      // 18b: getfield sa.m I
      // 18e: invokestatic java/lang/Math.min (II)I
      // 191: putfield sa.g I
      // 194: aload 0
      // 195: aload 0
      // 196: getfield sa.m I
      // 199: putfield sa.c I
      // 19c: goto 1a4
      // 19f: aload 0
      // 1a0: bipush 0
      // 1a1: putfield sa.l Z
      // 1a4: aload 0
      // 1a5: bipush 0
      // 1a6: putfield sa.m I
      // 1a9: aload 0
      // 1aa: lload 1
      // 1ab: ldc2_w 2000
      // 1ae: ladd
      // 1af: putfield sa.f J
      // 1b2: aload 0
      // 1b3: iload 3
      // 1b4: putfield sa.h I
      // 1b7: goto 1c8
      // 1ba: astore 3
      // 1bb: aload 0
      // 1bc: invokevirtual sa.e ()V
      // 1bf: aload 0
      // 1c0: lload 1
      // 1c1: ldc2_w 2000
      // 1c4: ladd
      // 1c5: putfield sa.e J
      // 1c8: return
   }

   final synchronized void a(va var1) {
      this.a = var1;
   }

   static final void a(int param0, boolean param1, int param2) {
      // $VF: Couldn't be decompiled
      // Please report this to the Vineflower issue tracker, at https://github.com/Vineflower/vineflower/issues with a copy of the class file (if you have the rights to distribute it!)
      // java.lang.RuntimeException: parsing failure!
      //   at org.jetbrains.java.decompiler.modules.decompiler.decompose.DomHelper.parseGraph(DomHelper.java:211)
      //   at org.jetbrains.java.decompiler.main.rels.MethodProcessor.codeToJava(MethodProcessor.java:166)
      //
      // Bytecode:
      // 00: iload 0
      // 01: sipush 8000
      // 04: if_icmplt 11
      // 07: iload 0
      // 08: ldc 48000
      // 0a: if_icmple 1a
      // 0d: goto 11
      // 10: athrow
      // 11: new java/lang/IllegalArgumentException
      // 14: dup
      // 15: invokespecial java/lang/IllegalArgumentException.<init> ()V
      // 18: athrow
      // 19: athrow
      // 1a: iload 0
      // 1b: putstatic sa.t I
      // 1e: iload 1
      // 1f: putstatic sa.i Z
      // 22: iload 2
      // 23: putstatic sa.o I
      // 26: return
   }

   private final void a(int[] var1, int var2) {
      int var3 = var2;
      if (i) {
         var3 = var2 << 1;
      }

      label178: {
         try {
            ab.a(var1, 0, var3);
            this.p -= var2;
            if (this.a == null || this.p > 0) {
               break label178;
            }
         } catch (IllegalStateException var26) {
            throw var26;
         }

         this.p = this.p + (t >> 4);
         b(this.a);
         this.a(this.a, this.a.c());
         int var4 = 0;
         int var5 = 255;
         int var6 = 7;

         label166:
         while (true) {
            int var7;
            int var8;
            label117: {
               label116: {
                  try {
                     if (var5 == 0) {
                        break;
                     }

                     if (var6 < 0) {
                        break label116;
                     }
                  } catch (IllegalStateException var20) {
                     throw var20;
                  }

                  var7 = var6;
                  var8 = 0;
                  break label117;
               }

               var7 = var6 & 3;
               var8 = -(var6 >> 2);
            }

            int var9 = var5 >>> var7 & 286331153;

            while (true) {
               label180: {
                  try {
                     if (var9 == 0) {
                        break;
                     }
                  } catch (IllegalStateException var21) {
                     throw var21;
                  }

                  try {
                     if ((var9 & 1) == 0) {
                        break label180;
                     }
                  } catch (IllegalStateException var22) {
                     throw var22;
                  }

                  var5 &= ~(1 << var7);
                  va var10 = null;
                  va var11 = this.d[var7];

                  label160:
                  while (true) {
                     while (true) {
                        if (var11 == null) {
                           break label160;
                        }

                        bb var12 = var11.h;

                        try {
                           if (var12 != null && var12.g > var8) {
                              break;
                           }
                        } catch (IllegalStateException var25) {
                           throw var25;
                        }

                        var11.g = true;
                        int var13 = var11.d();
                        var4 += var13;

                        try {
                           if (var12 != null) {
                              var12.g += var13;
                           }
                        } catch (IllegalStateException var17) {
                           throw var17;
                        }

                        try {
                           if (var4 >= this.r) {
                              break label166;
                           }
                        } catch (IllegalStateException var19) {
                           throw var19;
                        }

                        va var14 = var11.b();
                        if (var14 != null) {
                           for (int var15 = var11.i; var14 != null; var14 = var11.a()) {
                              this.a(var14, var15 * var14.c() >> 8);
                           }
                        }

                        va var31 = var11.j;

                        label141: {
                           try {
                              var11.j = null;
                              if (var10 == null) {
                                 this.d[var7] = var31;
                                 break label141;
                              }
                           } catch (IllegalStateException var24) {
                              throw var24;
                           }

                           var10.j = var31;
                        }

                        try {
                           if (var31 == null) {
                              this.s[var7] = var10;
                           }
                        } catch (IllegalStateException var23) {
                           throw var23;
                        }

                        var11 = var31;
                     }

                     var5 |= 1 << var7;
                     var10 = var11;
                     var11 = var11.j;
                  }
               }

               var7 += 4;
               var8++;
               var9 >>>= 4;
            }

            var6--;
         }

         for (int var27 = 0; var27 < 8; var27++) {
            va var28 = this.d[var27];
            va[] var29 = this.d;
            this.s[var27] = null;
            var29[var27] = null;

            while (var28 != null) {
               va var30 = var28.j;
               var28.j = null;
               var28 = var30;
            }
         }
      }

      try {
         if (this.p < 0) {
            this.p = 0;
         }
      } catch (IllegalStateException var16) {
         throw var16;
      }

      try {
         if (this.a != null) {
            this.a.b(var1, 0, var2);
         }
      } catch (IllegalStateException var18) {
         throw var18;
      }

      this.u = p.a(0);
   }

   private final void a(int var1) {
      try {
         this.p -= var1;
         if (this.p < 0) {
            this.p = 0;
         }
      } catch (IllegalStateException var3) {
         throw var3;
      }

      try {
         if (this.a != null) {
            this.a.b(var1);
         }
      } catch (IllegalStateException var2) {
         throw var2;
      }
   }

   private static final void b(va var0) {
      try {
         var0.g = false;
         if (var0.h != null) {
            var0.h.g = 0;
         }
      } catch (IllegalStateException var2) {
         throw var2;
      }

      for (va var1 = var0.b(); var1 != null; var1 = var0.a()) {
         b(var1);
      }
   }

   void a(Component var1) throws Exception {
   }

   sa() {
      this.u = p.a(0);
      this.s = new va[8];
      this.c = 0;
      this.m = 0;
      this.e = 0L;
      this.p = 0;
      this.d = new va[8];
      this.f = 0L;
      this.h = 0;
      this.l = true;
   }
}
