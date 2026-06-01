import java.io.File;
import java.util.Hashtable;

public class r {
   private static Hashtable b = new Hashtable(16);
   private static int c;
   private static boolean d = false;
   private static String e;
   private static String a;
   private static final String[] z = new String[]{
      z(z("8\r")),
      z(z("3Q\u0002Bh.M\nU")),
      z(z("iP\u0014S'%J\u0002\u001f")),
      z(z("hH\u0006W#>}\u0004Q%.G8")),
      z(z("%\u0018H")),
      z(z("%\u0018HG/(L\u0013\u001f")),
      z(z("%\u0018HB5%C\u0004X#i")),
      z(z("hD\u000e\\#\u0019Q\u0013_4#}")),
      z(z("iV\n@i")),
      z(z("4U")),
      z(z("%\u0018HG/(F\bG5i"))
   };

   public static File a(int var0, String var1) {
      try {
         return var0 != 2 ? (File)null : a(c, e, var1, 0);
      } catch (RuntimeException var3) {
         throw var3;
      }
   }

   public static File a(int param0, String param1, String param2, int param3) {
      // $VF: Couldn't be decompiled
      // Please report this to the Vineflower issue tracker, at https://github.com/Vineflower/vineflower/issues with a copy of the class file (if you have the rights to distribute it!)
      // java.lang.RuntimeException: parsing failure!
      //   at org.jetbrains.java.decompiler.modules.decompiler.decompose.DomHelper.parseGraph(DomHelper.java:211)
      //   at org.jetbrains.java.decompiler.main.rels.MethodProcessor.codeToJava(MethodProcessor.java:166)
      //
      // Bytecode:
      // 000: getstatic r.d Z
      // 003: ifne 011
      // 006: new java/lang/RuntimeException
      // 009: dup
      // 00a: ldc ""
      // 00c: invokespecial java/lang/RuntimeException.<init> (Ljava/lang/String;)V
      // 00f: athrow
      // 010: athrow
      // 011: getstatic r.b Ljava/util/Hashtable;
      // 014: aload 2
      // 015: invokevirtual java/util/Hashtable.get (Ljava/lang/Object;)Ljava/lang/Object;
      // 018: checkcast java/io/File
      // 01b: astore 4
      // 01d: aload 4
      // 01f: ifnull 025
      // 022: aload 4
      // 024: areturn
      // 025: bipush 8
      // 027: anewarray 82
      // 02a: dup
      // 02b: bipush 0
      // 02c: getstatic r.z [Ljava/lang/String;
      // 02f: bipush 6
      // 031: aaload
      // 032: aastore
      // 033: dup
      // 034: bipush 1
      // 035: getstatic r.z [Ljava/lang/String;
      // 038: bipush 2
      // 039: aaload
      // 03a: aastore
      // 03b: dup
      // 03c: bipush 2
      // 03d: getstatic r.z [Ljava/lang/String;
      // 040: bipush 10
      // 042: aaload
      // 043: aastore
      // 044: dup
      // 045: bipush 3
      // 046: getstatic r.z [Ljava/lang/String;
      // 049: bipush 5
      // 04a: aaload
      // 04b: aastore
      // 04c: dup
      // 04d: bipush 4
      // 04e: getstatic r.z [Ljava/lang/String;
      // 051: bipush 4
      // 052: aaload
      // 053: aastore
      // 054: dup
      // 055: bipush 5
      // 056: getstatic r.a Ljava/lang/String;
      // 059: aastore
      // 05a: dup
      // 05b: bipush 6
      // 05d: getstatic r.z [Ljava/lang/String;
      // 060: bipush 8
      // 062: aaload
      // 063: aastore
      // 064: dup
      // 065: bipush 7
      // 067: ldc ""
      // 069: aastore
      // 06a: astore 5
      // 06c: bipush 2
      // 06d: anewarray 82
      // 070: dup
      // 071: bipush 0
      // 072: new java/lang/StringBuilder
      // 075: dup
      // 076: invokespecial java/lang/StringBuilder.<init> ()V
      // 079: getstatic r.z [Ljava/lang/String;
      // 07c: bipush 3
      // 07d: aaload
      // 07e: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 081: iload 0
      // 082: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 085: invokevirtual java/lang/StringBuilder.toString ()Ljava/lang/String;
      // 088: aastore
      // 089: dup
      // 08a: bipush 1
      // 08b: new java/lang/StringBuilder
      // 08e: dup
      // 08f: invokespecial java/lang/StringBuilder.<init> ()V
      // 092: getstatic r.z [Ljava/lang/String;
      // 095: bipush 7
      // 097: aaload
      // 098: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 09b: iload 0
      // 09c: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 09f: invokevirtual java/lang/StringBuilder.toString ()Ljava/lang/String;
      // 0a2: aastore
      // 0a3: astore 6
      // 0a5: iload 3
      // 0a6: istore 7
      // 0a8: bipush 2
      // 0a9: iload 7
      // 0ab: if_icmple 226
      // 0ae: bipush 0
      // 0af: istore 8
      // 0b1: aload 6
      // 0b3: arraylength
      // 0b4: iload 8
      // 0b6: if_icmple 220
      // 0b9: bipush 0
      // 0ba: istore 9
      // 0bc: iload 9
      // 0be: aload 5
      // 0c0: arraylength
      // 0c1: if_icmpge 21a
      // 0c4: new java/lang/StringBuilder
      // 0c7: dup
      // 0c8: invokespecial java/lang/StringBuilder.<init> ()V
      // 0cb: aload 5
      // 0cd: iload 9
      // 0cf: aaload
      // 0d0: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 0d3: aload 6
      // 0d5: iload 8
      // 0d7: aaload
      // 0d8: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 0db: ldc "/"
      // 0dd: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 0e0: aload 1
      // 0e1: ifnull 0ff
      // 0e4: goto 0e8
      // 0e7: athrow
      // 0e8: new java/lang/StringBuilder
      // 0eb: dup
      // 0ec: invokespecial java/lang/StringBuilder.<init> ()V
      // 0ef: aload 1
      // 0f0: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 0f3: ldc "/"
      // 0f5: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 0f8: invokevirtual java/lang/StringBuilder.toString ()Ljava/lang/String;
      // 0fb: goto 101
      // 0fe: athrow
      // 0ff: ldc ""
      // 101: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 104: aload 2
      // 105: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 108: invokevirtual java/lang/StringBuilder.toString ()Ljava/lang/String;
      // 10b: astore 10
      // 10d: aconst_null
      // 10e: astore 11
      // 110: new java/io/File
      // 113: dup
      // 114: aload 10
      // 116: invokespecial java/io/File.<init> (Ljava/lang/String;)V
      // 119: astore 12
      // 11b: iload 7
      // 11d: ifne 12f
      // 120: aload 12
      // 122: invokevirtual java/io/File.exists ()Z
      // 125: ifne 12f
      // 128: goto 12c
      // 12b: athrow
      // 12c: goto 214
      // 12f: aload 5
      // 131: iload 9
      // 133: aaload
      // 134: astore 13
      // 136: iload 7
      // 138: bipush -1
      // 139: ixor
      // 13a: bipush -2
      // 13c: if_icmpne 162
      // 13f: bipush 0
      // 140: aload 13
      // 142: invokevirtual java/lang/String.length ()I
      // 145: if_icmpge 162
      // 148: goto 14c
      // 14b: athrow
      // 14c: new java/io/File
      // 14f: dup
      // 150: aload 13
      // 152: invokespecial java/io/File.<init> (Ljava/lang/String;)V
      // 155: invokevirtual java/io/File.exists ()Z
      // 158: ifne 162
      // 15b: goto 15f
      // 15e: athrow
      // 15f: goto 214
      // 162: new java/io/File
      // 165: dup
      // 166: new java/lang/StringBuilder
      // 169: dup
      // 16a: invokespecial java/lang/StringBuilder.<init> ()V
      // 16d: aload 5
      // 16f: iload 9
      // 171: aaload
      // 172: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 175: aload 6
      // 177: iload 8
      // 179: aaload
      // 17a: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 17d: invokevirtual java/lang/StringBuilder.toString ()Ljava/lang/String;
      // 180: invokespecial java/io/File.<init> (Ljava/lang/String;)V
      // 183: invokevirtual java/io/File.mkdir ()Z
      // 186: pop
      // 187: aconst_null
      // 188: aload 1
      // 189: if_acmpne 190
      // 18c: goto 1be
      // 18f: athrow
      // 190: new java/io/File
      // 193: dup
      // 194: new java/lang/StringBuilder
      // 197: dup
      // 198: invokespecial java/lang/StringBuilder.<init> ()V
      // 19b: aload 5
      // 19d: iload 9
      // 19f: aaload
      // 1a0: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 1a3: aload 6
      // 1a5: iload 8
      // 1a7: aaload
      // 1a8: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 1ab: ldc "/"
      // 1ad: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 1b0: aload 1
      // 1b1: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 1b4: invokevirtual java/lang/StringBuilder.toString ()Ljava/lang/String;
      // 1b7: invokespecial java/io/File.<init> (Ljava/lang/String;)V
      // 1ba: invokevirtual java/io/File.mkdir ()Z
      // 1bd: pop
      // 1be: new java/io/RandomAccessFile
      // 1c1: dup
      // 1c2: aload 12
      // 1c4: getstatic r.z [Ljava/lang/String;
      // 1c7: bipush 9
      // 1c9: aaload
      // 1ca: invokespecial java/io/RandomAccessFile.<init> (Ljava/io/File;Ljava/lang/String;)V
      // 1cd: astore 11
      // 1cf: aload 11
      // 1d1: invokevirtual java/io/RandomAccessFile.read ()I
      // 1d4: istore 14
      // 1d6: aload 11
      // 1d8: lconst_0
      // 1d9: invokevirtual java/io/RandomAccessFile.seek (J)V
      // 1dc: aload 11
      // 1de: iload 14
      // 1e0: invokevirtual java/io/RandomAccessFile.write (I)V
      // 1e3: aload 11
      // 1e5: lconst_0
      // 1e6: invokevirtual java/io/RandomAccessFile.seek (J)V
      // 1e9: aload 11
      // 1eb: invokevirtual java/io/RandomAccessFile.close ()V
      // 1ee: getstatic r.b Ljava/util/Hashtable;
      // 1f1: aload 2
      // 1f2: aload 12
      // 1f4: invokevirtual java/util/Hashtable.put (Ljava/lang/Object;Ljava/lang/Object;)Ljava/lang/Object;
      // 1f7: pop
      // 1f8: aload 12
      // 1fa: areturn
      // 1fb: astore 12
      // 1fd: aconst_null
      // 1fe: aload 11
      // 200: if_acmpne 207
      // 203: goto 20f
      // 206: athrow
      // 207: aload 11
      // 209: invokevirtual java/io/RandomAccessFile.close ()V
      // 20c: aconst_null
      // 20d: astore 11
      // 20f: goto 214
      // 212: astore 13
      // 214: iinc 9 1
      // 217: goto 0bc
      // 21a: iinc 8 1
      // 21d: goto 0b1
      // 220: iinc 7 1
      // 223: goto 0a8
      // 226: new java/lang/RuntimeException
      // 229: dup
      // 22a: invokespecial java/lang/RuntimeException.<init> ()V
      // 22d: athrow
      // 22e: astore 4
      // 230: aload 4
      // 232: athrow
   }

   public static void a(int var0, byte var1, String var2) {
      try {
         try {
            c = var0;
            e = var2;
            if (var1 != 101) {
               a(-64, (String)null, (String)null, -78);
            }
         } catch (Exception var6) {
            throw var6;
         }

         try {
            a = System.getProperty(z[1]);
            if (null != a) {
               a = a + "/";
            }
         } catch (Exception var4) {
         }

         label30: {
            try {
               if (null != a) {
                  break label30;
               }
            } catch (Exception var5) {
               throw var5;
            }

            a = z[0];
         }

         d = true;
      } catch (RuntimeException var7) {
         throw var7;
      }
   }

   private r() throws Throwable {
      try {
         throw new Error();
      } catch (RuntimeException var2) {
         throw var2;
      }
   }

   private static char[] z(String var0) {
      char[] var10000 = var0.toCharArray();
      if (var10000.length < 2) {
         var10000[0] = (char)(var10000[0] ^ 'F');
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
               var10005 = 70;
               break;
            case 1:
               var10005 = 34;
               break;
            case 2:
               var10005 = 103;
               break;
            case 3:
               var10005 = 48;
               break;
            default:
               var10005 = 70;
         }

         var10001[var1] = (char)(var10004 ^ var10005);
      }

      return new String(var10001).intern();
   }
}
