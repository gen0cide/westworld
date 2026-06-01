import java.applet.Applet;
import java.io.IOException;
import java.io.InputStream;
import java.io.OutputStream;
import java.net.Socket;
import java.net.URL;

final class da extends b implements Runnable {
   static int[][] J;
   private byte[] Z = new byte[1];
   static fb db = new fb();
   private int W;
   private boolean fb;
   static int S;
   private int G = 0;
   static int[] T;
   private OutputStream Q;
   static Applet gb;
   static int R;
   static int P;
   static int eb;
   private InputStream U;
   static int M = 0;
   private Socket ab;
   private boolean X;
   static int V;
   static int cb;
   static int L;
   static int I;
   static int H;
   static v O = new v(z(z("\u0003v:;")), z(z(";D\u000e\u0011Q1")), z(z("\u000bP\u000b")), 1);
   private byte[] Y;
   static int[] N;
   static int bb;
   static int K;
   private static final String[] hb = new String[]{
      z(z("0CF?\u001a")),
      z(z("\u0011P\u001a\u0017@tA\u0004\u0017A=L\u000fXA P\r\u0019_")),
      z(z(":W\u0004\u0014")),
      z(z("/\fFVO")),
      z(z("t\u000fH")),
      z(z("0CF=\u001a")),
      z(z("0CF0\u001a")),
      z(z("\u0017M\u001d\u0014V:\u0005\u001cXV;U\u0006\u0014]5FH\u001e[8G")),
      z(z("6W\u000e\u001eW&\u0002\u0007\u000eW&D\u0004\u0017E")),
      z(z("0CF<\u001a")),
      z(z("0CF:\u001a")),
      z(z("0CFD[:K\u001cF\u001a")),
      z(z("\u0011m.")),
      z(z("0CF>\u001a")),
      z(z("0CF9\u001a")),
      z(z("0CF\nG:\nA")),
      z(z("\u0000U\u001a\u0011F1PR")),
      z(z("0CF;\u001a"))
   };

   @Override
   public final void run() {
      boolean var6 = client.vh;

      try {
         R++;

         while (true) {
            boolean var10000;
            if (!this.X) {
               var10000 = true;

               try {
                  if (!var6) {
                  }
               } catch (InterruptedException var7) {
                  throw var7;
               }
            } else {
               var10000 = false;
            }

            if (var10000) {
               int var1;
               int var2;
               synchronized (this) {
                  if (this.G == this.W) {
                     try {
                        this.wait();
                     } catch (InterruptedException var10) {
                     }
                  }

                  try {
                     if (this.X) {
                        return;
                     }
                  } catch (InterruptedException var12) {
                     throw var12;
                  }

                  label78: {
                     if (this.W > this.G) {
                        var1 = 5000 - this.W;
                        if (!var6) {
                           break label78;
                        }
                     }

                     var1 = this.G + -this.W;
                  }

                  var2 = this.W;
               }

               label105: {
                  try {
                     if (-1 <= ~var1) {
                        break label105;
                     }
                  } catch (InterruptedException var11) {
                     throw var11;
                  }

                  try {
                     this.Q.write(this.Y, var2, var1);
                  } catch (IOException var9) {
                     this.k = true;
                     this.B = hb[16] + var9;
                  }

                  this.W = (this.W - -var1) % 5000;

                  try {
                     if (this.W == this.G) {
                        this.Q.flush();
                     }
                  } catch (IOException var8) {
                     this.k = true;
                     this.B = hb[16] + var8;
                  }
               }

               if (!var6) {
                  continue;
               }
            }

            return;
         }
      } catch (RuntimeException var14) {
         throw i.a(var14, hb[15]);
      }
   }

   static final void a(String var0, int var1, int var2) {
      try {
         try {
            if (var2 != 0) {
               a((String)null, -126, -28);
            }
         } catch (RuntimeException var5) {
            throw var5;
         }

         I++;
         d.h.a(nb.q, (byte)-101, var0 + o.l + hb[4] + var1 + "%");
      } catch (RuntimeException var6) {
         RuntimeException var3 = var6;

         RuntimeException var10000;
         StringBuilder var10001;
         try {
            var10000 = var3;
            var10001 = new StringBuilder().append(hb[5]);
            if (var0 != null) {
               throw i.a(var3, var10001.append(hb[3]).append(',').append(var1).append(',').append(var2).append(')').toString());
            }
         } catch (RuntimeException var4) {
            throw var4;
         }

         throw i.a(var10000, var10001.append(hb[2]).append(',').append(var1).append(',').append(var2).append(')').toString());
      }
   }

   static final byte[] a(URL var0, boolean var1, boolean var2) throws IOException {
      boolean var5 = client.vh;

      try {
         L++;
         jb var12 = new jb(pa.k, var0, 2000000);

         try {
            if (var1) {
               a("", 0, 0);
            }
         } catch (RuntimeException var7) {
            throw var7;
         }

         while (!var12.a(-2)) {
            mb.a(11200, 50L);
            if (var5) {
               break;
            }
         }

         tb var4 = var12.a((byte)-120);

         try {
            if (var4 == null) {
               throw new IOException(hb[7]);
            }
         } catch (RuntimeException var10) {
            throw var10;
         }

         try {
            if (!var2) {
               a((String)null, -15, -97);
            }
         } catch (RuntimeException var6) {
            throw var6;
         }

         try {
            if (var1) {
               a("", 100, 0);
            }
         } catch (RuntimeException var9) {
            throw var9;
         }

         return var4.d(0);
      } catch (RuntimeException var11) {
         RuntimeException var3 = var11;

         RuntimeException var10000;
         StringBuilder var10001;
         try {
            var10000 = var3;
            var10001 = new StringBuilder().append(hb[6]);
            if (var0 != null) {
               throw i.a(var3, var10001.append(hb[3]).append(',').append(var1).append(',').append(var2).append(')').toString());
            }
         } catch (RuntimeException var8) {
            throw var8;
         }

         throw i.a(var10000, var10001.append(hb[2]).append(',').append(var1).append(',').append(var2).append(')').toString());
      }
   }

   @Override
   final void a(byte[] param1, int param2, int param3, byte param4) throws IOException {
      // $VF: Couldn't be decompiled
      // Please report this to the Vineflower issue tracker, at https://github.com/Vineflower/vineflower/issues with a copy of the class file (if you have the rights to distribute it!)
      // java.lang.RuntimeException: parsing failure!
      //   at org.jetbrains.java.decompiler.modules.decompiler.decompose.DomHelper.parseGraph(DomHelper.java:211)
      //   at org.jetbrains.java.decompiler.main.rels.MethodProcessor.codeToJava(MethodProcessor.java:166)
      //
      // Bytecode:
      // 000: getstatic client.vh Z
      // 003: istore 8
      // 005: iload 4
      // 007: bipush -67
      // 009: if_icmpeq 00e
      // 00c: return
      // 00d: athrow
      // 00e: getstatic da.S I
      // 011: bipush 1
      // 012: iadd
      // 013: putstatic da.S I
      // 016: bipush 1
      // 017: aload 0
      // 018: getfield da.fb Z
      // 01b: ifne 023
      // 01e: bipush 1
      // 01f: goto 024
      // 022: athrow
      // 023: bipush 0
      // 024: if_icmpne 02a
      // 027: goto 02b
      // 02a: return
      // 02b: aconst_null
      // 02c: aload 0
      // 02d: getfield da.Y [B
      // 030: if_acmpne 040
      // 033: aload 0
      // 034: sipush 5000
      // 037: newarray 8
      // 039: putfield da.Y [B
      // 03c: goto 040
      // 03f: athrow
      // 040: aload 0
      // 041: dup
      // 042: astore 5
      // 044: monitorenter
      // 045: bipush 0
      // 046: istore 6
      // 048: iload 6
      // 04a: iload 3
      // 04b: if_icmpge 09e
      // 04e: aload 0
      // 04f: getfield da.Y [B
      // 052: aload 0
      // 053: getfield da.G I
      // 056: aload 1
      // 057: iload 2
      // 058: iload 6
      // 05a: iadd
      // 05b: baload
      // 05c: bastore
      // 05d: aload 0
      // 05e: aload 0
      // 05f: getfield da.G I
      // 062: bipush -1
      // 063: isub
      // 064: sipush 5000
      // 067: irem
      // 068: putfield da.G I
      // 06b: aload 0
      // 06c: iload 8
      // 06e: ifne 0a4
      // 071: getfield da.G I
      // 074: sipush 4900
      // 077: aload 0
      // 078: getfield da.W I
      // 07b: iadd
      // 07c: sipush 5000
      // 07f: irem
      // 080: if_icmpne 096
      // 083: goto 087
      // 086: athrow
      // 087: new java/io/IOException
      // 08a: dup
      // 08b: getstatic da.hb [Ljava/lang/String;
      // 08e: bipush 8
      // 090: aaload
      // 091: invokespecial java/io/IOException.<init> (Ljava/lang/String;)V
      // 094: athrow
      // 095: athrow
      // 096: iinc 6 1
      // 099: iload 8
      // 09b: ifeq 048
      // 09e: aload 0
      // 09f: invokevirtual java/lang/Object.notify ()V
      // 0a2: aload 5
      // 0a4: monitorexit
      // 0a5: goto 0b0
      // 0a8: astore 7
      // 0aa: aload 5
      // 0ac: monitorexit
      // 0ad: aload 7
      // 0af: athrow
      // 0b0: goto 104
      // 0b3: astore 5
      // 0b5: aload 5
      // 0b7: new java/lang/StringBuilder
      // 0ba: dup
      // 0bb: invokespecial java/lang/StringBuilder.<init> ()V
      // 0be: getstatic da.hb [Ljava/lang/String;
      // 0c1: bipush 9
      // 0c3: aaload
      // 0c4: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 0c7: aload 1
      // 0c8: ifnull 0d4
      // 0cb: getstatic da.hb [Ljava/lang/String;
      // 0ce: bipush 3
      // 0cf: aaload
      // 0d0: goto 0d9
      // 0d3: athrow
      // 0d4: getstatic da.hb [Ljava/lang/String;
      // 0d7: bipush 2
      // 0d8: aaload
      // 0d9: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 0dc: bipush 44
      // 0de: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 0e1: iload 2
      // 0e2: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 0e5: bipush 44
      // 0e7: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 0ea: iload 3
      // 0eb: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 0ee: bipush 44
      // 0f0: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 0f3: iload 4
      // 0f5: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 0f8: bipush 41
      // 0fa: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 0fd: invokevirtual java/lang/StringBuilder.toString ()Ljava/lang/String;
      // 100: invokestatic i.a (Ljava/lang/Throwable;Ljava/lang/String;)Lla;
      // 103: athrow
      // 104: return
   }

   @Override
   final int b(byte param1) throws IOException {
      // $VF: Couldn't be decompiled
      // Please report this to the Vineflower issue tracker, at https://github.com/Vineflower/vineflower/issues with a copy of the class file (if you have the rights to distribute it!)
      // java.lang.RuntimeException: parsing failure!
      //   at org.jetbrains.java.decompiler.modules.decompiler.decompose.DomHelper.parseGraph(DomHelper.java:211)
      //   at org.jetbrains.java.decompiler.main.rels.MethodProcessor.codeToJava(MethodProcessor.java:166)
      //
      // Bytecode:
      // 00: getstatic da.H I
      // 03: bipush 1
      // 04: iadd
      // 05: putstatic da.H I
      // 08: bipush 1
      // 09: aload 0
      // 0a: getfield da.fb Z
      // 0d: if_icmpne 13
      // 10: bipush 0
      // 11: ireturn
      // 12: athrow
      // 13: bipush -127
      // 15: iload 1
      // 16: bipush -64
      // 18: isub
      // 19: bipush 56
      // 1b: idiv
      // 1c: irem
      // 1d: istore 2
      // 1e: aload 0
      // 1f: getfield da.U Ljava/io/InputStream;
      // 22: invokevirtual java/io/InputStream.available ()I
      // 25: ireturn
      // 26: astore 2
      // 27: aload 2
      // 28: new java/lang/StringBuilder
      // 2b: dup
      // 2c: invokespecial java/lang/StringBuilder.<init> ()V
      // 2f: getstatic da.hb [Ljava/lang/String;
      // 32: bipush 10
      // 34: aaload
      // 35: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 38: iload 1
      // 39: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 3c: bipush 41
      // 3e: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 41: invokevirtual java/lang/StringBuilder.toString ()Ljava/lang/String;
      // 44: invokestatic i.a (Ljava/lang/Throwable;Ljava/lang/String;)Lla;
      // 47: athrow
   }

   @Override
   final void a(boolean var1) {
      try {
         V++;
         super.a(true);
         this.fb = var1;

         try {
            if (this.U != null) {
               this.U.close();
            }

            try {
               if (null != this.Q) {
                  this.Q.close();
               }
            } catch (IOException var6) {
               throw var6;
            }

            try {
               if (null != this.ab) {
                  this.ab.close();
               }
            } catch (IOException var5) {
               throw var5;
            }
         } catch (IOException var7) {
            System.out.println(hb[1]);
         }

         this.X = true;
         synchronized (this) {
            this.notify();
         }

         this.Y = null;
      } catch (RuntimeException var8) {
         throw i.a(var8, hb[0] + var1 + ')');
      }
   }

   static final int a(int var0, byte var1, int var2, int var3) {
      try {
         try {
            if (var1 != -66) {
               K = 35;
            }
         } catch (RuntimeException var5) {
            throw var5;
         }

         cb++;
         return -(var3 / 8 * 32) + -1 + -(var2 / 8 * 1024) - var0 / 8;
      } catch (RuntimeException var6) {
         throw i.a(var6, hb[17] + var0 + 44 + var1 + 44 + var2 + 44 + var3 + 41);
      }
   }

   @Override
   final int b(boolean var1) throws IOException {
      try {
         boolean var10000;
         label32: {
            try {
               P++;
               if (!this.fb) {
                  var10000 = true;
                  break label32;
               }
            } catch (RuntimeException var4) {
               throw var4;
            }

            var10000 = false;
         }

         try {
            if (var10000 != var1) {
               return 0;
            }
         } catch (RuntimeException var3) {
            throw var3;
         }

         this.a(this.Z, 1, 0, 123);
         return 0xFF & this.Z[0];
      } catch (RuntimeException var5) {
         throw i.a(var5, hb[14] + var1 + 41);
      }
   }

   @Override
   final void a(byte[] param1, int param2, int param3, int param4) throws IOException {
      // $VF: Couldn't be decompiled
      // Please report this to the Vineflower issue tracker, at https://github.com/Vineflower/vineflower/issues with a copy of the class file (if you have the rights to distribute it!)
      // java.lang.RuntimeException: parsing failure!
      //   at org.jetbrains.java.decompiler.modules.decompiler.decompose.DomHelper.parseGraph(DomHelper.java:211)
      //   at org.jetbrains.java.decompiler.main.rels.MethodProcessor.codeToJava(MethodProcessor.java:166)
      //
      // Bytecode:
      // 00: getstatic client.vh Z
      // 03: istore 8
      // 05: getstatic da.eb I
      // 08: bipush 1
      // 09: iadd
      // 0a: putstatic da.eb I
      // 0d: aload 0
      // 0e: getfield da.fb Z
      // 11: bipush 1
      // 12: if_icmpne 17
      // 15: return
      // 16: athrow
      // 17: bipush -81
      // 19: bipush -25
      // 1b: iload 4
      // 1d: isub
      // 1e: bipush 35
      // 20: idiv
      // 21: idiv
      // 22: istore 6
      // 24: bipush 0
      // 25: istore 5
      // 27: bipush 0
      // 28: istore 7
      // 2a: iload 5
      // 2c: iload 2
      // 2d: if_icmpge 6a
      // 30: iload 8
      // 32: ifne be
      // 35: aload 0
      // 36: getfield da.U Ljava/io/InputStream;
      // 39: aload 1
      // 3a: iload 3
      // 3b: iload 5
      // 3d: iadd
      // 3e: iload 5
      // 40: ineg
      // 41: iload 2
      // 42: iadd
      // 43: invokevirtual java/io/InputStream.read ([BII)I
      // 46: dup
      // 47: istore 7
      // 49: bipush -1
      // 4a: ixor
      // 4b: bipush -1
      // 4c: if_icmplt 5e
      // 4f: new java/io/IOException
      // 52: dup
      // 53: getstatic da.hb [Ljava/lang/String;
      // 56: bipush 12
      // 58: aaload
      // 59: invokespecial java/io/IOException.<init> (Ljava/lang/String;)V
      // 5c: athrow
      // 5d: athrow
      // 5e: iload 5
      // 60: iload 7
      // 62: iadd
      // 63: istore 5
      // 65: iload 8
      // 67: ifeq 2a
      // 6a: goto be
      // 6d: astore 5
      // 6f: aload 5
      // 71: new java/lang/StringBuilder
      // 74: dup
      // 75: invokespecial java/lang/StringBuilder.<init> ()V
      // 78: getstatic da.hb [Ljava/lang/String;
      // 7b: bipush 13
      // 7d: aaload
      // 7e: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 81: aload 1
      // 82: ifnull 8e
      // 85: getstatic da.hb [Ljava/lang/String;
      // 88: bipush 3
      // 89: aaload
      // 8a: goto 93
      // 8d: athrow
      // 8e: getstatic da.hb [Ljava/lang/String;
      // 91: bipush 2
      // 92: aaload
      // 93: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 96: bipush 44
      // 98: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 9b: iload 2
      // 9c: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 9f: bipush 44
      // a1: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // a4: iload 3
      // a5: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // a8: bipush 44
      // aa: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // ad: iload 4
      // af: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // b2: bipush 41
      // b4: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // b7: invokevirtual java/lang/StringBuilder.toString ()Ljava/lang/String;
      // ba: invokestatic i.a (Ljava/lang/Throwable;Ljava/lang/String;)Lla;
      // bd: athrow
      // be: return
   }

   da(Socket var1, e var2) throws IOException {
      this.W = 0;
      this.fb = false;
      this.X = true;

      try {
         this.ab = var1;
         this.U = var1.getInputStream();
         this.Q = var1.getOutputStream();
         this.X = false;
         var2.a(1, this);
      } catch (RuntimeException var6) {
         RuntimeException var3 = var6;

         RuntimeException var10000;
         StringBuilder var10001;
         String var10002;
         label34: {
            try {
               var10000 = var3;
               var10001 = new StringBuilder().append(hb[11]);
               if (var1 != null) {
                  var10002 = hb[3];
                  break label34;
               }
            } catch (RuntimeException var5) {
               throw var5;
            }

            var10002 = hb[2];
         }

         try {
            var10001 = var10001.append(var10002).append(',');
            if (var2 != null) {
               throw i.a(var10000, var10001.append(hb[3]).append(')').toString());
            }
         } catch (RuntimeException var4) {
            throw var4;
         }

         throw i.a(var10000, var10001.append(hb[2]).append(')').toString());
      }
   }

   private static char[] z(String var0) {
      char[] var10000 = var0.toCharArray();
      if (var10000.length < 2) {
         var10000[0] = (char)(var10000[0] ^ '2');
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
               var10005 = 84;
               break;
            case 1:
               var10005 = 34;
               break;
            case 2:
               var10005 = 104;
               break;
            case 3:
               var10005 = 120;
               break;
            default:
               var10005 = 50;
         }

         var10001[var1] = (char)(var10004 ^ var10005);
      }

      return new String(var10001).intern();
   }
}
