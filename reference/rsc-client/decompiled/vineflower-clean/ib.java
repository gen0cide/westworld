import java.io.IOException;
import java.net.URL;

class ib {
   ib a;
   static URL c;
   static int[] d;
   ib e;
   static int b;
   static int f;
   private static final String[] z = new String[]{
      z(z("8\u00108\u001d#y")),
      z(z("*\\8|\u001f")),
      z(z("q\u001es<_")),
      z(z("kRu \u0001l")),
      z(z("\u0012\u001dc>\u0006?Ubr\u0006>\u0005x>\r0\u001664\u000b=\u00176q")),
      z(z("2\u001dx&\u0007?\u0006")),
      z(z("?\u0007z>")),
      z(z("8\u00108\u0002#y")),
      z(z("8\u00108\u0003#y"))
   };

   static int a(int var0, int var1) {
      try {
         return var0 & var1;
      } catch (RuntimeException var3) {
         throw i.a(var3, z[8] + var0 + 44 + var1 + 41);
      }
   }

   final void a(int var1) {
      try {
         f++;
         if (var1 == -27331) {
            if (null != this.e) {
               this.e.a = this.a;
               this.a.e = this.e;
               this.a = null;
               this.e = null;
            }
         }
      } catch (RuntimeException var3) {
         throw i.a(var3, z[7] + var1 + ')');
      }
   }

   protected ib() {
   }

   static final byte[] a(int param0, String param1, int param2, int param3) throws IOException {
      // $VF: Couldn't be decompiled
      // Please report this to the Vineflower issue tracker, at https://github.com/Vineflower/vineflower/issues with a copy of the class file (if you have the rights to distribute it!)
      // java.lang.RuntimeException: parsing failure!
      //   at org.jetbrains.java.decompiler.modules.decompiler.decompose.DomHelper.parseGraph(DomHelper.java:235)
      //   at org.jetbrains.java.decompiler.main.rels.MethodProcessor.codeToJava(MethodProcessor.java:174)
      //
      // Bytecode:
      // 000: getstatic client.vh Z
      // 003: istore 8
      // 005: getstatic ib.b I
      // 008: bipush 1
      // 009: iadd
      // 00a: putstatic ib.b I
      // 00d: aconst_null
      // 00e: getstatic la.g [[B
      // 011: iload 3
      // 012: aaload
      // 013: if_acmpeq 01d
      // 016: getstatic la.g [[B
      // 019: iload 3
      // 01a: aaload
      // 01b: areturn
      // 01c: athrow
      // 01d: iload 0
      // 01e: bipush -73
      // 020: if_icmple 029
      // 023: aconst_null
      // 024: checkcast [B
      // 027: areturn
      // 028: athrow
      // 029: iload 2
      // 02a: putstatic nb.q I
      // 02d: aload 1
      // 02e: putstatic o.l Ljava/lang/String;
      // 031: getstatic m.e Lob;
      // 034: ifnonnull 03b
      // 037: goto 07a
      // 03a: athrow
      // 03b: getstatic m.e Lob;
      // 03e: sipush 9395
      // 041: iload 3
      // 042: invokevirtual ob.a (II)[B
      // 045: astore 4
      // 047: aconst_null
      // 048: aload 4
      // 04a: if_acmpne 051
      // 04d: goto 07a
      // 050: athrow
      // 051: aload 4
      // 053: aload 4
      // 055: arraylength
      // 056: bipush 0
      // 057: invokestatic mb.a ([BII)I
      // 05a: getstatic tb.l [I
      // 05d: iload 3
      // 05e: iaload
      // 05f: if_icmpeq 066
      // 062: goto 07a
      // 065: athrow
      // 066: getstatic la.g [[B
      // 069: iload 3
      // 06a: sipush 128
      // 06d: bipush 1
      // 06e: aload 4
      // 070: invokestatic k.a (IZ[B)[B
      // 073: aastore
      // 074: getstatic la.g [[B
      // 077: iload 3
      // 078: aaload
      // 079: areturn
      // 07a: new java/net/URL
      // 07d: dup
      // 07e: getstatic ib.c Ljava/net/URL;
      // 081: new java/lang/StringBuilder
      // 084: dup
      // 085: invokespecial java/lang/StringBuilder.<init> ()V
      // 088: getstatic ib.z [Ljava/lang/String;
      // 08b: bipush 5
      // 08c: aaload
      // 08d: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 090: iload 3
      // 091: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 094: ldc "_"
      // 096: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 099: getstatic tb.l [I
      // 09c: iload 3
      // 09d: iaload
      // 09e: i2l
      // 09f: invokestatic java/lang/Long.toHexString (J)Ljava/lang/String;
      // 0a2: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 0a5: invokevirtual java/lang/StringBuilder.toString ()Ljava/lang/String;
      // 0a8: invokespecial java/net/URL.<init> (Ljava/net/URL;Ljava/lang/String;)V
      // 0ab: astore 4
      // 0ad: aconst_null
      // 0ae: astore 5
      // 0b0: bipush 0
      // 0b1: istore 6
      // 0b3: iload 6
      // 0b5: bipush -1
      // 0b6: ixor
      // 0b7: bipush -4
      // 0b9: if_icmple 130
      // 0bc: aload 4
      // 0be: bipush 1
      // 0bf: bipush 1
      // 0c0: invokestatic da.a (Ljava/net/URL;ZZ)[B
      // 0c3: astore 5
      // 0c5: iload 8
      // 0c7: ifne 1d1
      // 0ca: aload 5
      // 0cc: aload 5
      // 0ce: arraylength
      // 0cf: bipush 0
      // 0d0: invokestatic mb.a ([BII)I
      // 0d3: bipush -1
      // 0d4: ixor
      // 0d5: getstatic tb.l [I
      // 0d8: iload 3
      // 0d9: iaload
      // 0da: bipush -1
      // 0db: ixor
      // 0dc: if_icmpeq 0e8
      // 0df: goto 0e3
      // 0e2: athrow
      // 0e3: iload 8
      // 0e5: ifeq 128
      // 0e8: aconst_null
      // 0e9: getstatic m.e Lob;
      // 0ec: if_acmpeq 102
      // 0ef: getstatic m.e Lob;
      // 0f2: iload 3
      // 0f3: aload 5
      // 0f5: arraylength
      // 0f6: bipush -97
      // 0f8: aload 5
      // 0fa: invokevirtual ob.a (III[B)Z
      // 0fd: pop
      // 0fe: goto 102
      // 101: athrow
      // 102: getstatic la.g [[B
      // 105: iload 3
      // 106: sipush 128
      // 109: bipush 1
      // 10a: aload 5
      // 10c: invokestatic k.a (IZ[B)[B
      // 10f: aastore
      // 110: getstatic la.g [[B
      // 113: iload 3
      // 114: aaload
      // 115: areturn
      // 116: astore 7
      // 118: iload 6
      // 11a: bipush -1
      // 11b: ixor
      // 11c: bipush -3
      // 11e: if_icmpeq 125
      // 121: goto 128
      // 124: athrow
      // 125: aload 7
      // 127: athrow
      // 128: iinc 6 1
      // 12b: iload 8
      // 12d: ifeq 0b3
      // 130: aconst_null
      // 131: aload 5
      // 133: if_acmpeq 1d1
      // 136: new java/lang/StringBuilder
      // 139: dup
      // 13a: new java/lang/StringBuilder
      // 13d: dup
      // 13e: invokespecial java/lang/StringBuilder.<init> ()V
      // 141: getstatic ib.z [Ljava/lang/String;
      // 144: bipush 4
      // 145: aaload
      // 146: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 149: iload 3
      // 14a: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 14d: getstatic ib.z [Ljava/lang/String;
      // 150: bipush 3
      // 151: aaload
      // 152: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 155: getstatic tb.l [I
      // 158: iload 3
      // 159: iaload
      // 15a: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 15d: invokevirtual java/lang/StringBuilder.toString ()Ljava/lang/String;
      // 160: invokespecial java/lang/StringBuilder.<init> (Ljava/lang/String;)V
      // 163: astore 6
      // 165: aload 6
      // 167: new java/lang/StringBuilder
      // 16a: dup
      // 16b: invokespecial java/lang/StringBuilder.<init> ()V
      // 16e: getstatic ib.z [Ljava/lang/String;
      // 171: bipush 2
      // 172: aaload
      // 173: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 176: aload 5
      // 178: arraylength
      // 179: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 17c: invokevirtual java/lang/StringBuilder.toString ()Ljava/lang/String;
      // 17f: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 182: pop
      // 183: bipush 0
      // 184: istore 7
      // 186: aload 5
      // 188: arraylength
      // 189: bipush -1
      // 18a: ixor
      // 18b: iload 7
      // 18d: bipush -1
      // 18e: ixor
      // 18f: if_icmpge 1c4
      // 192: bipush -6
      // 194: iload 7
      // 196: bipush -1
      // 197: ixor
      // 198: if_icmpge 1c4
      // 19b: aload 6
      // 19d: new java/lang/StringBuilder
      // 1a0: dup
      // 1a1: invokespecial java/lang/StringBuilder.<init> ()V
      // 1a4: ldc " "
      // 1a6: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 1a9: aload 5
      // 1ab: iload 7
      // 1ad: baload
      // 1ae: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 1b1: invokevirtual java/lang/StringBuilder.toString ()Ljava/lang/String;
      // 1b4: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 1b7: pop
      // 1b8: iinc 7 1
      // 1bb: iload 8
      // 1bd: ifeq 186
      // 1c0: goto 1c4
      // 1c3: athrow
      // 1c4: new java/io/IOException
      // 1c7: dup
      // 1c8: aload 6
      // 1ca: invokevirtual java/lang/StringBuilder.toString ()Ljava/lang/String;
      // 1cd: invokespecial java/io/IOException.<init> (Ljava/lang/String;)V
      // 1d0: athrow
      // 1d1: new java/io/IOException
      // 1d4: dup
      // 1d5: new java/lang/StringBuilder
      // 1d8: dup
      // 1d9: invokespecial java/lang/StringBuilder.<init> ()V
      // 1dc: getstatic ib.z [Ljava/lang/String;
      // 1df: bipush 4
      // 1e0: aaload
      // 1e1: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 1e4: iload 3
      // 1e5: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 1e8: getstatic ib.z [Ljava/lang/String;
      // 1eb: bipush 3
      // 1ec: aaload
      // 1ed: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 1f0: getstatic tb.l [I
      // 1f3: iload 3
      // 1f4: iaload
      // 1f5: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 1f8: invokevirtual java/lang/StringBuilder.toString ()Ljava/lang/String;
      // 1fb: invokespecial java/io/IOException.<init> (Ljava/lang/String;)V
      // 1fe: athrow
      // 1ff: astore 4
      // 201: aload 4
      // 203: new java/lang/StringBuilder
      // 206: dup
      // 207: invokespecial java/lang/StringBuilder.<init> ()V
      // 20a: getstatic ib.z [Ljava/lang/String;
      // 20d: bipush 0
      // 20e: aaload
      // 20f: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 212: iload 0
      // 213: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 216: bipush 44
      // 218: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 21b: aload 1
      // 21c: ifnull 228
      // 21f: getstatic ib.z [Ljava/lang/String;
      // 222: bipush 1
      // 223: aaload
      // 224: goto 22e
      // 227: athrow
      // 228: getstatic ib.z [Ljava/lang/String;
      // 22b: bipush 6
      // 22d: aaload
      // 22e: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 231: bipush 44
      // 233: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 236: iload 2
      // 237: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 23a: bipush 44
      // 23c: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 23f: iload 3
      // 240: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 243: bipush 41
      // 245: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 248: invokevirtual java/lang/StringBuilder.toString ()Ljava/lang/String;
      // 24b: invokestatic i.a (Ljava/lang/Throwable;Ljava/lang/String;)Lla;
      // 24e: athrow
      // try (95 -> 117): 143 java/io/IOException
      // try (119 -> 142): 143 java/io/IOException
      // try (2 -> 14): 256 java/lang/RuntimeException
      // try (16 -> 21): 256 java/lang/RuntimeException
      // try (23 -> 62): 256 java/lang/RuntimeException
      // try (63 -> 142): 256 java/lang/RuntimeException
      // try (143 -> 256): 256 java/lang/RuntimeException
   }

   private static char[] z(String var0) {
      char[] var10000 = var0.toCharArray();
      if (var10000.length < 2) {
         var10000[0] = (char)(var10000[0] ^ 98);
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
               var10005 = 81;
               break;
            case 1:
               var10005 = 114;
               break;
            case 2:
               var10005 = 22;
               break;
            case 3:
               var10005 = 82;
               break;
            default:
               var10005 = 98;
         }

         var10001[var1] = (char)(var10004 ^ var10005);
      }

      return new String(var10001).intern();
   }
}
