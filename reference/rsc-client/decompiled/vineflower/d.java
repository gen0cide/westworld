import java.io.EOFException;
import java.io.File;
import java.io.IOException;
import java.io.RandomAccessFile;

final class d {
   static int n;
   static int e;
   static int l = 0;
   static int k;
   static int d;
   static int i;
   static int f;
   static int[] g;
   static int a;
   private RandomAccessFile j;
   private long c;
   static e h;
   private long b;
   static String[] m = new String[]{
      z(z("(X\u001boB\u001d\u0014\u001b`E\u001dF^zY\u001d\u0014\u0010{\\\u001aQ\f.^\u001e\u0014\u0017zT\u0015G^z^XC\u0017zY\u001cF\u001fy")),
      z(z("\u0019Z\u001a.A\nQ\r}\u0011\u001dZ\nkC"))
   };
   private static final String[] z = new String[]{
      z(z("\u0003\u001aP L")),
      z(z("\u001c\u001a:&")),
      z(z("\u0016A\u0012b")),
      z(z("\u001c\u001a<&")),
      z(z("\u001c\u001a8&")),
      z(z("\u001c\u001aBg_\u0011@@&")),
      z(z("\u001c\u001a\u0018g_\u0019X\u0017tTP\u001d")),
      z(z("\u001c\u001a9&")),
      z(z("\u001c\u001a?&")),
      z(z("\u001c\u001a;&")),
      z(z("\u001c\u001a=&"))
   };

   private final void a(int var1) throws IOException {
      try {
         label25: {
            try {
               n++;
               if (null == this.j) {
                  break label25;
               }
            } catch (RuntimeException var4) {
               throw var4;
            }

            this.j.close();
            this.j = null;
         }

         try {
            if (var1 != 25291) {
               a(62, (byte)14, (byte[])null);
            }
         } catch (RuntimeException var3) {
            throw var3;
         }
      } catch (RuntimeException var5) {
         throw i.a(var5, z[9] + var1 + ')');
      }
   }

   final long a(byte var1) throws IOException {
      try {
         try {
            if (var1 != 47) {
               this.c = -52L;
            }
         } catch (RuntimeException var3) {
            throw var3;
         }

         i++;
         return this.j.length();
      } catch (RuntimeException var4) {
         throw i.a(var4, z[8] + var1 + ')');
      }
   }

   @Override
   protected final void finalize() throws Throwable {
      try {
         a++;
         if (null != this.j) {
            System.out.println("");
            this.a(25291);
         }
      } catch (RuntimeException var2) {
         throw i.a(var2, z[6]);
      }
   }

   static int a(int var0, int var1) {
      try {
         return var0 | var1;
      } catch (RuntimeException var3) {
         throw i.a(var3, z[3] + var0 + 44 + var1 + 41);
      }
   }

   final void b(byte[] var1, int var2, int var3, int var4) throws IOException {
      try {
         try {
            d++;
            if (~this.b > ~(this.c + (long)var2)) {
               this.j.seek(this.b);
               this.j.write(1);
               throw new EOFException();
            }
         } catch (RuntimeException var8) {
            throw var8;
         }

         try {
            this.j.write(var1, var4, var2);
            this.c += (long)var2;
            if (var3 != 1) {
               a(63, (byte)-101, (byte[])null);
            }
         } catch (RuntimeException var6) {
            throw var6;
         }
      } catch (RuntimeException var9) {
         RuntimeException var5 = var9;

         RuntimeException var10000;
         StringBuilder var10001;
         try {
            var10000 = var5;
            var10001 = new StringBuilder().append(z[7]);
            if (var1 != null) {
               throw i.a(var5, var10001.append(z[0]).append(',').append(var2).append(',').append(var3).append(',').append(var4).append(')').toString());
            }
         } catch (RuntimeException var7) {
            throw var7;
         }

         throw i.a(var10000, var10001.append(z[2]).append(',').append(var2).append(',').append(var3).append(',').append(var4).append(')').toString());
      }
   }

   static final int a(int var0, byte var1, byte[] var2) {
      try {
         try {
            if (var1 < 4) {
               h = (e)null;
            }
         } catch (RuntimeException var5) {
            throw var5;
         }

         f++;
         return (var2[1 + var0] & 0xFF) + ((0xFF & var2[var0]) << 922410888);
      } catch (RuntimeException var6) {
         RuntimeException var3 = var6;

         RuntimeException var10000;
         StringBuilder var10001;
         try {
            var10000 = var3;
            var10001 = new StringBuilder().append(z[1]).append(var0).append(',').append((int)var1).append(',');
            if (var2 != null) {
               throw i.a(var3, var10001.append(z[0]).append((char)41).toString());
            }
         } catch (RuntimeException var4) {
            throw var4;
         }

         throw i.a(var10000, var10001.append(z[2]).append((char)41).toString());
      }
   }

   final void a(int param1, long param2) throws IOException {
      // $VF: Couldn't be decompiled
      // Please report this to the Vineflower issue tracker, at https://github.com/Vineflower/vineflower/issues with a copy of the class file (if you have the rights to distribute it!)
      // java.lang.RuntimeException: parsing failure!
      //   at org.jetbrains.java.decompiler.modules.decompiler.decompose.DomHelper.parseGraph(DomHelper.java:211)
      //   at org.jetbrains.java.decompiler.main.rels.MethodProcessor.codeToJava(MethodProcessor.java:166)
      //
      // Bytecode:
      // 00: iload 1
      // 01: ifeq 06
      // 04: return
      // 05: athrow
      // 06: aload 0
      // 07: getfield d.j Ljava/io/RandomAccessFile;
      // 0a: lload 2
      // 0b: invokevirtual java/io/RandomAccessFile.seek (J)V
      // 0e: getstatic d.k I
      // 11: bipush 1
      // 12: iadd
      // 13: putstatic d.k I
      // 16: aload 0
      // 17: lload 2
      // 18: putfield d.c J
      // 1b: goto 4a
      // 1e: astore 4
      // 20: aload 4
      // 22: new java/lang/StringBuilder
      // 25: dup
      // 26: invokespecial java/lang/StringBuilder.<init> ()V
      // 29: getstatic d.z [Ljava/lang/String;
      // 2c: bipush 4
      // 2d: aaload
      // 2e: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 31: iload 1
      // 32: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 35: bipush 44
      // 37: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 3a: lload 2
      // 3b: invokevirtual java/lang/StringBuilder.append (J)Ljava/lang/StringBuilder;
      // 3e: bipush 41
      // 40: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 43: invokevirtual java/lang/StringBuilder.toString ()Ljava/lang/String;
      // 46: invokestatic i.a (Ljava/lang/Throwable;Ljava/lang/String;)Lla;
      // 49: athrow
      // 4a: return
   }

   final int a(byte[] var1, int var2, int var3, int var4) throws IOException {
      try {
         e++;
         int var9 = this.j.read(var1, var3, var2);

         try {
            if (~var9 >= var4) {
               return var9;
            }
         } catch (RuntimeException var7) {
            throw var7;
         }

         this.c += (long)var9;
         return var9;
      } catch (RuntimeException var8) {
         RuntimeException var5 = var8;

         RuntimeException var10000;
         StringBuilder var10001;
         try {
            var10000 = var5;
            var10001 = new StringBuilder().append(z[10]);
            if (var1 != null) {
               throw i.a(
                  var5,
                  var10001.append(z[0]).append((char)44).append(var2).append((char)44).append(var3).append((char)44).append(var4).append((char)41).toString()
               );
            }
         } catch (RuntimeException var6) {
            throw var6;
         }

         throw i.a(
            var10000,
            var10001.append(z[2]).append((char)44).append(var2).append((char)44).append(var3).append((char)44).append(var4).append((char)41).toString()
         );
      }
   }

   d(File param1, String param2, long param3) throws IOException {
      // $VF: Couldn't be decompiled
      // Please report this to the Vineflower issue tracker, at https://github.com/Vineflower/vineflower/issues with a copy of the class file (if you have the rights to distribute it!)
      // java.lang.RuntimeException: parsing failure!
      //   at org.jetbrains.java.decompiler.modules.decompiler.decompose.DomHelper.parseGraph(DomHelper.java:211)
      //   at org.jetbrains.java.decompiler.main.rels.MethodProcessor.codeToJava(MethodProcessor.java:166)
      //
      // Bytecode:
      // 00: aload 0
      // 01: invokespecial java/lang/Object.<init> ()V
      // 04: lconst_0
      // 05: lload 3
      // 06: ldc2_w -1
      // 09: lxor
      // 0a: lcmp
      // 0b: ifeq 11
      // 0e: goto 15
      // 11: ldc2_w 9223372036854775807
      // 14: lstore 3
      // 15: lload 3
      // 16: ldc2_w -1
      // 19: lxor
      // 1a: aload 1
      // 1b: invokevirtual java/io/File.length ()J
      // 1e: ldc2_w -1
      // 21: lxor
      // 22: lcmp
      // 23: ifgt 2a
      // 26: goto 2f
      // 29: athrow
      // 2a: aload 1
      // 2b: invokevirtual java/io/File.delete ()Z
      // 2e: pop
      // 2f: aload 0
      // 30: new java/io/RandomAccessFile
      // 33: dup
      // 34: aload 1
      // 35: aload 2
      // 36: invokespecial java/io/RandomAccessFile.<init> (Ljava/io/File;Ljava/lang/String;)V
      // 39: putfield d.j Ljava/io/RandomAccessFile;
      // 3c: aload 0
      // 3d: lload 3
      // 3e: putfield d.b J
      // 41: aload 0
      // 42: lconst_0
      // 43: putfield d.c J
      // 46: aload 0
      // 47: getfield d.j Ljava/io/RandomAccessFile;
      // 4a: invokevirtual java/io/RandomAccessFile.read ()I
      // 4d: istore 5
      // 4f: bipush 0
      // 50: iload 5
      // 52: bipush -1
      // 53: ixor
      // 54: if_icmpeq 79
      // 57: aload 2
      // 58: ldc "r"
      // 5a: invokevirtual java/lang/String.equals (Ljava/lang/Object;)Z
      // 5d: ifne 79
      // 60: goto 64
      // 63: athrow
      // 64: aload 0
      // 65: getfield d.j Ljava/io/RandomAccessFile;
      // 68: lconst_0
      // 69: invokevirtual java/io/RandomAccessFile.seek (J)V
      // 6c: aload 0
      // 6d: getfield d.j Ljava/io/RandomAccessFile;
      // 70: iload 5
      // 72: invokevirtual java/io/RandomAccessFile.write (I)V
      // 75: goto 79
      // 78: athrow
      // 79: aload 0
      // 7a: getfield d.j Ljava/io/RandomAccessFile;
      // 7d: lconst_0
      // 7e: invokevirtual java/io/RandomAccessFile.seek (J)V
      // 81: goto db
      // 84: astore 5
      // 86: aload 5
      // 88: new java/lang/StringBuilder
      // 8b: dup
      // 8c: invokespecial java/lang/StringBuilder.<init> ()V
      // 8f: getstatic d.z [Ljava/lang/String;
      // 92: bipush 5
      // 93: aaload
      // 94: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 97: aload 1
      // 98: ifnull a4
      // 9b: getstatic d.z [Ljava/lang/String;
      // 9e: bipush 0
      // 9f: aaload
      // a0: goto a9
      // a3: athrow
      // a4: getstatic d.z [Ljava/lang/String;
      // a7: bipush 2
      // a8: aaload
      // a9: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // ac: bipush 44
      // ae: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // b1: aload 2
      // b2: ifnull be
      // b5: getstatic d.z [Ljava/lang/String;
      // b8: bipush 0
      // b9: aaload
      // ba: goto c3
      // bd: athrow
      // be: getstatic d.z [Ljava/lang/String;
      // c1: bipush 2
      // c2: aaload
      // c3: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // c6: bipush 44
      // c8: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // cb: lload 3
      // cc: invokevirtual java/lang/StringBuilder.append (J)Ljava/lang/StringBuilder;
      // cf: bipush 41
      // d1: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // d4: invokevirtual java/lang/StringBuilder.toString ()Ljava/lang/String;
      // d7: invokestatic i.a (Ljava/lang/Throwable;Ljava/lang/String;)Lla;
      // da: athrow
      // db: return
   }

   private static char[] z(String var0) {
      char[] var10000 = var0.toCharArray();
      if (var10000.length < 2) {
         var10000[0] = (char)(var10000[0] ^ '1');
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
               var10005 = 120;
               break;
            case 1:
               var10005 = 52;
               break;
            case 2:
               var10005 = 126;
               break;
            case 3:
               var10005 = 14;
               break;
            default:
               var10005 = 49;
         }

         var10001[var1] = (char)(var10004 ^ var10005);
      }

      return new String(var10001).intern();
   }
}
