final class eb implements Runnable {
   volatile boolean a;
   volatile sa[] f = new sa[2];
   static int h;
   c g;
   volatile boolean i;
   static i e = new i(z(z("qp>I")), 0);
   static v d = new v(z(z("jm!")), z(z("R_\u000eekX")), z(z("bN\u001ce")), 5);
   static int[] b;
   private static char[] c = new char[256];
   private static final String z = z(z("X[F~}S\u0011A"));

   @Override
   public final void run() {
      // $VF: Couldn't be decompiled
      // Please report this to the Vineflower issue tracker, at https://github.com/Vineflower/vineflower/issues with a copy of the class file (if you have the rights to distribute it!)
      // java.lang.RuntimeException: parsing failure!
      //   at org.jetbrains.java.decompiler.modules.decompiler.decompose.DomHelper.parseGraph(DomHelper.java:211)
      //   at org.jetbrains.java.decompiler.main.rels.MethodProcessor.codeToJava(MethodProcessor.java:166)
      //
      // Bytecode:
      // 00: getstatic client.vh Z
      // 03: istore 4
      // 05: aload 0
      // 06: bipush 1
      // 07: putfield eb.i Z
      // 0a: getstatic eb.h I
      // 0d: bipush 1
      // 0e: iadd
      // 0f: putstatic eb.h I
      // 12: aload 0
      // 13: getfield eb.a Z
      // 16: ifne 63
      // 19: iload 4
      // 1b: ifne 68
      // 1e: bipush 0
      // 1f: istore 1
      // 20: iload 1
      // 21: bipush -1
      // 22: ixor
      // 23: bipush -3
      // 25: if_icmple 4c
      // 28: aload 0
      // 29: getfield eb.f [Lsa;
      // 2c: iload 1
      // 2d: aaload
      // 2e: astore 2
      // 2f: iload 4
      // 31: ifne 5e
      // 34: aload 2
      // 35: ifnull 44
      // 38: goto 3c
      // 3b: athrow
      // 3c: aload 2
      // 3d: invokevirtual sa.a ()V
      // 40: goto 44
      // 43: athrow
      // 44: iinc 1 1
      // 47: iload 4
      // 49: ifeq 20
      // 4c: sipush 11200
      // 4f: ldc2_w 10
      // 52: invokestatic mb.a (IJ)V
      // 55: aconst_null
      // 56: bipush 1
      // 57: aload 0
      // 58: getfield eb.g Lc;
      // 5b: invokestatic ba.a (Ljava/lang/Object;ILc;)V
      // 5e: iload 4
      // 60: ifeq 12
      // 63: aload 0
      // 64: bipush 0
      // 65: putfield eb.i Z
      // 68: goto 83
      // 6b: astore 1
      // 6c: ldc 2097151
      // 6e: aload 1
      // 6f: aconst_null
      // 70: invokestatic mb.a (ILjava/lang/Throwable;Ljava/lang/String;)V
      // 73: aload 0
      // 74: bipush 0
      // 75: putfield eb.i Z
      // 78: goto 83
      // 7b: astore 3
      // 7c: aload 0
      // 7d: bipush 0
      // 7e: putfield eb.i Z
      // 81: aload 3
      // 82: athrow
      // 83: goto 8f
      // 86: astore 1
      // 87: aload 1
      // 88: getstatic eb.z Ljava/lang/String;
      // 8b: invokestatic i.a (Ljava/lang/Throwable;Ljava/lang/String;)Lla;
      // 8e: athrow
      // 8f: return
   }

   eb() {
      this.a = false;
      this.i = false;
   }

   static {
      int var0 = 0;

      try {
         while (256 > var0) {
            c[var0] = (char)var0;
            var0++;
         }
      } catch (RuntimeException var1) {
         throw var1;
      }

      c[45] = '-';
      c[59] = ';';
      c[42] = '*';
      c[124] = '|';
      c[43] = '+';
      c[33] = '!';
      c[34] = '"';
      c[47] = '/';
      c[46] = '.';
      c[61] = '=';
      c[92] = '\\';
      c[44] = ',';
   }

   private static char[] z(String var0) {
      char[] var10000 = var0.toCharArray();
      if (var10000.length < 2) {
         var10000[0] = (char)(var10000[0] ^ '\b');
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
               var10005 = 61;
               break;
            case 1:
               var10005 = 57;
               break;
            case 2:
               var10005 = 104;
               break;
            case 3:
               var10005 = 12;
               break;
            default:
               var10005 = 8;
         }

         var10001[var1] = (char)(var10004 ^ var10005);
      }

      return new String(var10001).intern();
   }
}
