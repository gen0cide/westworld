import java.io.IOException;
import java.lang.reflect.Method;
import java.net.InetSocketAddress;
import java.net.Proxy;
import java.net.ProxySelector;
import java.net.Socket;
import java.net.SocketAddress;
import java.net.URI;
import java.net.URISyntaxException;
import java.net.URL;
import java.net.Proxy.Type;
import java.util.List;

final class gb extends m {
   static int o;
   static v n = new v(z(z("=rbe2 }")), z(z("\u001bZPN\u0014\u0011")), z(z("+UXS\u0015\u0011HW")), 6);
   private ProxySelector q;
   static int u;
   static int t;
   static int r;
   static int l;
   static int p;
   static int m;
   static int[] s;
   private static final String[] z = new String[]{
      z(z("\u001cHBW\u0004")),
      z(z("N\u0013\u0019")),
      z(z("\u001e]@FY\u001aYB\t\u0002\u0007Ye^\u0004\u0000Y[w\u0005\u001bD_B\u0004")),
      z(z("\u0000NCB")),
      z(z("\u001cHBW")),
      z(z("\u0013^\u0018c_")),
      z(z("\u0007IX\t\u0019\u0011H\u0018P\u0000\u0003\u0012FU\u0018\u0000SUH\u001bZTBS\u0007Z}CS\u001f\u0011RBN\u0014\u0015H_H\u0019=RPH")),
      z(z("N\u001c")),
      z(z("\u001cHBW\u0004N\u0013\u0019")),
      z(z("\u0013YBw\u0005\u001bDOf\u0002\u0000T")),
      z(z("\u0013YBo\u0012\u0015XSU9\u0015QS")),
      z(z("\u000f\u0012\u0018\t\n")),
      z(z("\u0007IFW\u0018\u0006HEw\u0005\u0011Y[W\u0003\u001dJSf\u0002\u0000TYU\u001e\u000e]BN\u0018\u001a")),
      z(z("\u0013YBo\u0012\u0015XSU!\u0015PCB")),
      z(z("\u0013^\u0018a_")),
      z(z("\u001aIZK")),
      z(z("T@\u0016")),
      z(z("ZVWQ\u0016N")),
      z(z("\b\u001c")),
      z(z("\u0013^\u0018o_")),
      z(z("\u0004NY_\u000eY]CS\u001f\u0011RBN\u0014\u0015HS\u001dW")),
      z(z("<hbwXE\u0012\u0006\u0007ED\f")),
      z(z("Ttbs'[\r\u0018\u0017}~")),
      z(z("<hbwXE\u0012\u0007\u0007ED\f")),
      z(z("7sxi27h\u0016")),
      z(z("~6")),
      z(z("Ttbs'[\r\u0018\u0017}")),
      z(z("=oy\nOL\t\u000f\nF")),
      z(z("<hbwXE\u0012\u0006\u0007CD\u000b")),
      z(z("\u0013^\u0018`_")),
      z(z("<hbwXE\u0012\u0007\u0007CD\u000b")),
      z(z("\u0013^\u0018\u001b\u001e\u001aUB\u0019_]")),
      z(z("\u0013^\u0018b_"))
   };

   gb() {
      try {
         this.q = ProxySelector.getDefault();
      } catch (RuntimeException var2) {
         throw i.a(var2, z[31]);
      }
   }

   private final Socket a(int param1, int param2, String param3, String param4) throws IOException {
      // $VF: Couldn't be decompiled
      // Please report this to the Vineflower issue tracker, at https://github.com/Vineflower/vineflower/issues with a copy of the class file (if you have the rights to distribute it!)
      // java.lang.RuntimeException: parsing failure!
      //   at org.jetbrains.java.decompiler.modules.decompiler.decompose.DomHelper.parseGraph(DomHelper.java:211)
      //   at org.jetbrains.java.decompiler.main.rels.MethodProcessor.codeToJava(MethodProcessor.java:166)
      //
      // Bytecode:
      // 000: getstatic client.vh Z
      // 003: istore 12
      // 005: getstatic gb.o I
      // 008: bipush 1
      // 009: iadd
      // 00a: putstatic gb.o I
      // 00d: new java/net/Socket
      // 010: dup
      // 011: aload 4
      // 013: iload 1
      // 014: invokespecial java/net/Socket.<init> (Ljava/lang/String;I)V
      // 017: astore 5
      // 019: aload 5
      // 01b: sipush 10000
      // 01e: invokevirtual java/net/Socket.setSoTimeout (I)V
      // 021: aload 5
      // 023: invokevirtual java/net/Socket.getOutputStream ()Ljava/io/OutputStream;
      // 026: astore 6
      // 028: aload 3
      // 029: ifnull 082
      // 02c: aload 6
      // 02e: new java/lang/StringBuilder
      // 031: dup
      // 032: invokespecial java/lang/StringBuilder.<init> ()V
      // 035: getstatic gb.z [Ljava/lang/String;
      // 038: bipush 24
      // 03a: aaload
      // 03b: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 03e: aload 0
      // 03f: getfield gb.h Ljava/lang/String;
      // 042: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 045: ldc ":"
      // 047: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 04a: aload 0
      // 04b: getfield gb.f I
      // 04e: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 051: getstatic gb.z [Ljava/lang/String;
      // 054: bipush 26
      // 056: aaload
      // 057: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 05a: aload 3
      // 05b: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 05e: getstatic gb.z [Ljava/lang/String;
      // 061: bipush 25
      // 063: aaload
      // 064: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 067: invokevirtual java/lang/StringBuilder.toString ()Ljava/lang/String;
      // 06a: getstatic gb.z [Ljava/lang/String;
      // 06d: bipush 27
      // 06f: aaload
      // 070: invokestatic java/nio/charset/Charset.forName (Ljava/lang/String;)Ljava/nio/charset/Charset;
      // 073: invokevirtual java/lang/String.getBytes (Ljava/nio/charset/Charset;)[B
      // 076: invokevirtual java/io/OutputStream.write ([B)V
      // 079: iload 12
      // 07b: ifeq 0c6
      // 07e: goto 082
      // 081: athrow
      // 082: aload 6
      // 084: new java/lang/StringBuilder
      // 087: dup
      // 088: invokespecial java/lang/StringBuilder.<init> ()V
      // 08b: getstatic gb.z [Ljava/lang/String;
      // 08e: bipush 24
      // 090: aaload
      // 091: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 094: aload 0
      // 095: getfield gb.h Ljava/lang/String;
      // 098: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 09b: ldc ":"
      // 09d: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 0a0: aload 0
      // 0a1: getfield gb.f I
      // 0a4: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 0a7: getstatic gb.z [Ljava/lang/String;
      // 0aa: bipush 22
      // 0ac: aaload
      // 0ad: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 0b0: invokevirtual java/lang/StringBuilder.toString ()Ljava/lang/String;
      // 0b3: getstatic gb.z [Ljava/lang/String;
      // 0b6: bipush 27
      // 0b8: aaload
      // 0b9: invokestatic java/nio/charset/Charset.forName (Ljava/lang/String;)Ljava/nio/charset/Charset;
      // 0bc: invokevirtual java/lang/String.getBytes (Ljava/nio/charset/Charset;)[B
      // 0bf: invokevirtual java/io/OutputStream.write ([B)V
      // 0c2: goto 0c6
      // 0c5: athrow
      // 0c6: aload 6
      // 0c8: invokevirtual java/io/OutputStream.flush ()V
      // 0cb: new java/io/BufferedReader
      // 0ce: dup
      // 0cf: new java/io/InputStreamReader
      // 0d2: dup
      // 0d3: aload 5
      // 0d5: invokevirtual java/net/Socket.getInputStream ()Ljava/io/InputStream;
      // 0d8: invokespecial java/io/InputStreamReader.<init> (Ljava/io/InputStream;)V
      // 0db: invokespecial java/io/BufferedReader.<init> (Ljava/io/Reader;)V
      // 0de: astore 7
      // 0e0: aload 7
      // 0e2: invokevirtual java/io/BufferedReader.readLine ()Ljava/lang/String;
      // 0e5: astore 8
      // 0e7: iload 2
      // 0e8: sipush 1514
      // 0eb: if_icmpeq 0f3
      // 0ee: aconst_null
      // 0ef: checkcast java/net/Socket
      // 0f2: areturn
      // 0f3: aload 8
      // 0f5: ifnonnull 0fc
      // 0f8: goto 1cb
      // 0fb: athrow
      // 0fc: aload 8
      // 0fe: getstatic gb.z [Ljava/lang/String;
      // 101: bipush 21
      // 103: aaload
      // 104: invokevirtual java/lang/String.startsWith (Ljava/lang/String;)Z
      // 107: ifne 120
      // 10a: aload 8
      // 10c: getstatic gb.z [Ljava/lang/String;
      // 10f: bipush 23
      // 111: aaload
      // 112: invokevirtual java/lang/String.startsWith (Ljava/lang/String;)Z
      // 115: ifne 120
      // 118: goto 11c
      // 11b: athrow
      // 11c: goto 123
      // 11f: athrow
      // 120: aload 5
      // 122: areturn
      // 123: aload 8
      // 125: getstatic gb.z [Ljava/lang/String;
      // 128: bipush 28
      // 12a: aaload
      // 12b: invokevirtual java/lang/String.startsWith (Ljava/lang/String;)Z
      // 12e: ifne 147
      // 131: aload 8
      // 133: getstatic gb.z [Ljava/lang/String;
      // 136: bipush 30
      // 138: aaload
      // 139: invokevirtual java/lang/String.startsWith (Ljava/lang/String;)Z
      // 13c: ifne 147
      // 13f: goto 143
      // 142: athrow
      // 143: goto 1cb
      // 146: athrow
      // 147: bipush 0
      // 148: istore 9
      // 14a: getstatic gb.z [Ljava/lang/String;
      // 14d: bipush 20
      // 14f: aaload
      // 150: astore 10
      // 152: aload 7
      // 154: invokevirtual java/io/BufferedReader.readLine ()Ljava/lang/String;
      // 157: astore 8
      // 159: aload 8
      // 15b: ifnull 1c1
      // 15e: bipush -51
      // 160: iload 9
      // 162: bipush -1
      // 163: ixor
      // 164: if_icmpge 1c1
      // 167: aload 8
      // 169: invokevirtual java/lang/String.toLowerCase ()Ljava/lang/String;
      // 16c: aload 10
      // 16e: invokevirtual java/lang/String.startsWith (Ljava/lang/String;)Z
      // 171: ifne 17c
      // 174: goto 178
      // 177: athrow
      // 178: goto 1b2
      // 17b: athrow
      // 17c: aload 8
      // 17e: aload 10
      // 180: invokevirtual java/lang/String.length ()I
      // 183: invokevirtual java/lang/String.substring (I)Ljava/lang/String;
      // 186: invokevirtual java/lang/String.trim ()Ljava/lang/String;
      // 189: astore 8
      // 18b: aload 8
      // 18d: bipush 32
      // 18f: invokevirtual java/lang/String.indexOf (I)I
      // 192: istore 11
      // 194: bipush -1
      // 195: iload 11
      // 197: if_icmpne 19e
      // 19a: goto 1a8
      // 19d: athrow
      // 19e: aload 8
      // 1a0: bipush 0
      // 1a1: iload 11
      // 1a3: invokevirtual java/lang/String.substring (II)Ljava/lang/String;
      // 1a6: astore 8
      // 1a8: new fa
      // 1ab: dup
      // 1ac: aload 8
      // 1ae: invokespecial fa.<init> (Ljava/lang/String;)V
      // 1b1: athrow
      // 1b2: iinc 9 1
      // 1b5: aload 7
      // 1b7: invokevirtual java/io/BufferedReader.readLine ()Ljava/lang/String;
      // 1ba: astore 8
      // 1bc: iload 12
      // 1be: ifeq 159
      // 1c1: new fa
      // 1c4: dup
      // 1c5: ldc ""
      // 1c7: invokespecial fa.<init> (Ljava/lang/String;)V
      // 1ca: athrow
      // 1cb: aload 6
      // 1cd: invokevirtual java/io/OutputStream.close ()V
      // 1d0: aload 7
      // 1d2: invokevirtual java/io/BufferedReader.close ()V
      // 1d5: aload 5
      // 1d7: invokevirtual java/net/Socket.close ()V
      // 1da: aconst_null
      // 1db: areturn
      // 1dc: astore 5
      // 1de: aload 5
      // 1e0: new java/lang/StringBuilder
      // 1e3: dup
      // 1e4: invokespecial java/lang/StringBuilder.<init> ()V
      // 1e7: getstatic gb.z [Ljava/lang/String;
      // 1ea: bipush 29
      // 1ec: aaload
      // 1ed: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 1f0: iload 1
      // 1f1: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 1f4: bipush 44
      // 1f6: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 1f9: iload 2
      // 1fa: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 1fd: bipush 44
      // 1ff: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 202: aload 3
      // 203: ifnull 210
      // 206: getstatic gb.z [Ljava/lang/String;
      // 209: bipush 11
      // 20b: aaload
      // 20c: goto 216
      // 20f: athrow
      // 210: getstatic gb.z [Ljava/lang/String;
      // 213: bipush 15
      // 215: aaload
      // 216: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 219: bipush 44
      // 21b: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 21e: aload 4
      // 220: ifnull 22d
      // 223: getstatic gb.z [Ljava/lang/String;
      // 226: bipush 11
      // 228: aaload
      // 229: goto 233
      // 22c: athrow
      // 22d: getstatic gb.z [Ljava/lang/String;
      // 230: bipush 15
      // 232: aaload
      // 233: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 236: bipush 41
      // 238: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 23b: invokevirtual java/lang/StringBuilder.toString ()Ljava/lang/String;
      // 23e: invokestatic i.a (Ljava/lang/Throwable;Ljava/lang/String;)Lla;
      // 241: athrow
   }

   static final void a(
      int param0,
      int param1,
      byte param2,
      int param3,
      int param4,
      int param5,
      int[] param6,
      int param7,
      int param8,
      int param9,
      int param10,
      int param11,
      int[] param12,
      int param13,
      int param14
   ) {
      // $VF: Couldn't be decompiled
      // Please report this to the Vineflower issue tracker, at https://github.com/Vineflower/vineflower/issues with a copy of the class file (if you have the rights to distribute it!)
      // java.lang.RuntimeException: parsing failure!
      //   at org.jetbrains.java.decompiler.modules.decompiler.decompose.DomHelper.parseGraph(DomHelper.java:211)
      //   at org.jetbrains.java.decompiler.main.rels.MethodProcessor.codeToJava(MethodProcessor.java:166)
      //
      // Bytecode:
      // 000: getstatic client.vh Z
      // 003: istore 21
      // 005: getstatic gb.m I
      // 008: bipush 1
      // 009: iadd
      // 00a: putstatic gb.m I
      // 00d: iload 14
      // 00f: ifle 015
      // 012: goto 016
      // 015: return
      // 016: bipush 0
      // 017: istore 15
      // 019: bipush 0
      // 01a: istore 16
      // 01c: bipush -1
      // 01d: iload 3
      // 01e: bipush -1
      // 01f: ixor
      // 020: if_icmpne 027
      // 023: goto 038
      // 026: athrow
      // 027: iload 8
      // 029: iload 3
      // 02a: idiv
      // 02b: ldc -410027673
      // 02d: ishl
      // 02e: istore 11
      // 030: iload 0
      // 031: iload 3
      // 032: idiv
      // 033: ldc -1978637785
      // 035: ishl
      // 036: istore 10
      // 038: bipush 0
      // 039: istore 19
      // 03b: iload 11
      // 03d: bipush -1
      // 03e: ixor
      // 03f: bipush -1
      // 040: if_icmple 04b
      // 043: bipush 0
      // 044: istore 11
      // 046: iload 21
      // 048: ifeq 060
      // 04b: sipush 16256
      // 04e: iload 11
      // 050: if_icmplt 05b
      // 053: goto 057
      // 056: athrow
      // 057: goto 060
      // 05a: athrow
      // 05b: sipush 16256
      // 05e: istore 11
      // 060: iload 2
      // 061: bipush 50
      // 063: if_icmpeq 067
      // 066: return
      // 067: iload 3
      // 068: iload 9
      // 06a: iadd
      // 06b: istore 3
      // 06c: iload 0
      // 06d: iload 13
      // 06f: iadd
      // 070: istore 0
      // 071: iload 8
      // 073: iload 1
      // 074: iadd
      // 075: istore 8
      // 077: iload 3
      // 078: ifne 07f
      // 07b: goto 090
      // 07e: athrow
      // 07f: iload 0
      // 080: iload 3
      // 081: idiv
      // 082: ldc -2053567033
      // 084: ishl
      // 085: istore 16
      // 087: iload 8
      // 089: iload 3
      // 08a: idiv
      // 08b: ldc 983280871
      // 08d: ishl
      // 08e: istore 15
      // 090: iload 15
      // 092: iflt 0ad
      // 095: iload 15
      // 097: bipush -1
      // 098: ixor
      // 099: sipush -16257
      // 09c: if_icmpge 0b0
      // 09f: goto 0a3
      // 0a2: athrow
      // 0a3: sipush 16256
      // 0a6: istore 15
      // 0a8: iload 21
      // 0aa: ifeq 0b0
      // 0ad: bipush 0
      // 0ae: istore 15
      // 0b0: iload 15
      // 0b2: iload 11
      // 0b4: ineg
      // 0b5: iadd
      // 0b6: ldc -858142236
      // 0b8: ishr
      // 0b9: istore 17
      // 0bb: iload 16
      // 0bd: iload 10
      // 0bf: isub
      // 0c0: ldc 1912830340
      // 0c2: ishr
      // 0c3: istore 18
      // 0c5: iload 14
      // 0c7: ldc -1893613788
      // 0c9: ishr
      // 0ca: istore 20
      // 0cc: iload 20
      // 0ce: bipush -1
      // 0cf: ixor
      // 0d0: bipush -1
      // 0d1: if_icmpge 454
      // 0d4: iload 11
      // 0d6: iload 4
      // 0d8: ldc 6291456
      // 0da: iand
      // 0db: iadd
      // 0dc: istore 11
      // 0de: iload 4
      // 0e0: ldc -407765257
      // 0e2: ishr
      // 0e3: istore 19
      // 0e5: aload 12
      // 0e7: iload 7
      // 0e9: iinc 7 1
      // 0ec: aload 6
      // 0ee: sipush 16256
      // 0f1: iload 10
      // 0f3: invokestatic ib.a (II)I
      // 0f6: iload 11
      // 0f8: ldc 897288903
      // 0fa: ishr
      // 0fb: iadd
      // 0fc: iaload
      // 0fd: iload 19
      // 0ff: iushr
      // 100: iastore
      // 101: iload 4
      // 103: iload 5
      // 105: iadd
      // 106: istore 4
      // 108: iload 11
      // 10a: iload 17
      // 10c: iadd
      // 10d: istore 11
      // 10f: iload 10
      // 111: iload 18
      // 113: iadd
      // 114: istore 10
      // 116: aload 12
      // 118: iload 7
      // 11a: iinc 7 1
      // 11d: aload 6
      // 11f: iload 11
      // 121: ldc -1712778393
      // 123: ishr
      // 124: sipush 16256
      // 127: iload 10
      // 129: invokestatic ib.a (II)I
      // 12c: iadd
      // 12d: iaload
      // 12e: iload 19
      // 130: iushr
      // 131: iastore
      // 132: iload 10
      // 134: iload 18
      // 136: iadd
      // 137: istore 10
      // 139: iload 11
      // 13b: iload 17
      // 13d: iadd
      // 13e: istore 11
      // 140: aload 12
      // 142: iload 7
      // 144: iinc 7 1
      // 147: aload 6
      // 149: iload 11
      // 14b: ldc -50303545
      // 14d: ishr
      // 14e: sipush 16256
      // 151: iload 10
      // 153: invokestatic ib.a (II)I
      // 156: iadd
      // 157: iaload
      // 158: iload 19
      // 15a: iushr
      // 15b: iastore
      // 15c: iload 10
      // 15e: iload 18
      // 160: iadd
      // 161: istore 10
      // 163: iload 11
      // 165: iload 17
      // 167: iadd
      // 168: istore 11
      // 16a: aload 12
      // 16c: iload 7
      // 16e: iinc 7 1
      // 171: aload 6
      // 173: iload 11
      // 175: ldc -755517209
      // 177: ishr
      // 178: sipush 16256
      // 17b: iload 10
      // 17d: invokestatic ib.a (II)I
      // 180: iadd
      // 181: iaload
      // 182: iload 19
      // 184: iushr
      // 185: iastore
      // 186: iload 10
      // 188: iload 18
      // 18a: iadd
      // 18b: istore 10
      // 18d: iload 11
      // 18f: iload 17
      // 191: iadd
      // 192: istore 11
      // 194: ldc 6291456
      // 196: iload 4
      // 198: iand
      // 199: sipush 16383
      // 19c: iload 11
      // 19e: iand
      // 19f: iadd
      // 1a0: istore 11
      // 1a2: iload 4
      // 1a4: ldc 779290135
      // 1a6: ishr
      // 1a7: istore 19
      // 1a9: aload 12
      // 1ab: iload 7
      // 1ad: iinc 7 1
      // 1b0: aload 6
      // 1b2: iload 11
      // 1b4: ldc 1348344199
      // 1b6: ishr
      // 1b7: iload 10
      // 1b9: sipush 16256
      // 1bc: invokestatic ib.a (II)I
      // 1bf: iadd
      // 1c0: iaload
      // 1c1: iload 19
      // 1c3: iushr
      // 1c4: iastore
      // 1c5: iload 4
      // 1c7: iload 5
      // 1c9: iadd
      // 1ca: istore 4
      // 1cc: iload 11
      // 1ce: iload 17
      // 1d0: iadd
      // 1d1: istore 11
      // 1d3: iload 10
      // 1d5: iload 18
      // 1d7: iadd
      // 1d8: istore 10
      // 1da: aload 12
      // 1dc: iload 7
      // 1de: iinc 7 1
      // 1e1: aload 6
      // 1e3: sipush 16256
      // 1e6: iload 10
      // 1e8: invokestatic ib.a (II)I
      // 1eb: iload 11
      // 1ed: ldc -1943849369
      // 1ef: ishr
      // 1f0: iadd
      // 1f1: iaload
      // 1f2: iload 19
      // 1f4: iushr
      // 1f5: iastore
      // 1f6: iload 11
      // 1f8: iload 17
      // 1fa: iadd
      // 1fb: istore 11
      // 1fd: iload 10
      // 1ff: iload 18
      // 201: iadd
      // 202: istore 10
      // 204: aload 12
      // 206: iload 7
      // 208: iinc 7 1
      // 20b: aload 6
      // 20d: iload 11
      // 20f: ldc -170574425
      // 211: ishr
      // 212: sipush 16256
      // 215: iload 10
      // 217: invokestatic ib.a (II)I
      // 21a: iadd
      // 21b: iaload
      // 21c: iload 19
      // 21e: iushr
      // 21f: iastore
      // 220: iload 10
      // 222: iload 18
      // 224: iadd
      // 225: istore 10
      // 227: iload 11
      // 229: iload 17
      // 22b: iadd
      // 22c: istore 11
      // 22e: aload 12
      // 230: iload 7
      // 232: iinc 7 1
      // 235: aload 6
      // 237: sipush 16256
      // 23a: iload 10
      // 23c: invokestatic ib.a (II)I
      // 23f: iload 11
      // 241: ldc 583157351
      // 243: ishr
      // 244: ineg
      // 245: isub
      // 246: iaload
      // 247: iload 19
      // 249: iushr
      // 24a: iastore
      // 24b: iload 10
      // 24d: iload 18
      // 24f: iadd
      // 250: istore 10
      // 252: iload 11
      // 254: iload 17
      // 256: iadd
      // 257: istore 11
      // 259: iload 4
      // 25b: ldc 6291456
      // 25d: iand
      // 25e: sipush 16383
      // 261: iload 11
      // 263: iand
      // 264: iadd
      // 265: istore 11
      // 267: iload 4
      // 269: ldc -1894754601
      // 26b: ishr
      // 26c: istore 19
      // 26e: iload 4
      // 270: iload 5
      // 272: iadd
      // 273: istore 4
      // 275: aload 12
      // 277: iload 7
      // 279: iinc 7 1
      // 27c: aload 6
      // 27e: iload 10
      // 280: sipush 16256
      // 283: invokestatic ib.a (II)I
      // 286: iload 11
      // 288: ldc 1007320743
      // 28a: ishr
      // 28b: ineg
      // 28c: isub
      // 28d: iaload
      // 28e: iload 19
      // 290: iushr
      // 291: iastore
      // 292: iload 11
      // 294: iload 17
      // 296: iadd
      // 297: istore 11
      // 299: iload 10
      // 29b: iload 18
      // 29d: iadd
      // 29e: istore 10
      // 2a0: aload 12
      // 2a2: iload 7
      // 2a4: iinc 7 1
      // 2a7: aload 6
      // 2a9: sipush 16256
      // 2ac: iload 10
      // 2ae: invokestatic ib.a (II)I
      // 2b1: iload 11
      // 2b3: ldc 1276542663
      // 2b5: ishr
      // 2b6: ineg
      // 2b7: isub
      // 2b8: iaload
      // 2b9: iload 19
      // 2bb: iushr
      // 2bc: iastore
      // 2bd: iload 11
      // 2bf: iload 17
      // 2c1: iadd
      // 2c2: istore 11
      // 2c4: iload 10
      // 2c6: iload 18
      // 2c8: iadd
      // 2c9: istore 10
      // 2cb: aload 12
      // 2cd: iload 7
      // 2cf: iinc 7 1
      // 2d2: aload 6
      // 2d4: iload 11
      // 2d6: ldc 646825255
      // 2d8: ishr
      // 2d9: sipush 16256
      // 2dc: iload 10
      // 2de: invokestatic ib.a (II)I
      // 2e1: iadd
      // 2e2: iaload
      // 2e3: iload 19
      // 2e5: iushr
      // 2e6: iastore
      // 2e7: iload 10
      // 2e9: iload 18
      // 2eb: iadd
      // 2ec: istore 10
      // 2ee: iload 11
      // 2f0: iload 17
      // 2f2: iadd
      // 2f3: istore 11
      // 2f5: aload 12
      // 2f7: iload 7
      // 2f9: iinc 7 1
      // 2fc: aload 6
      // 2fe: iload 11
      // 300: ldc 1218318887
      // 302: ishr
      // 303: iload 10
      // 305: sipush 16256
      // 308: invokestatic ib.a (II)I
      // 30b: iadd
      // 30c: iaload
      // 30d: iload 19
      // 30f: iushr
      // 310: iastore
      // 311: iload 11
      // 313: iload 17
      // 315: iadd
      // 316: istore 11
      // 318: iload 10
      // 31a: iload 18
      // 31c: iadd
      // 31d: istore 10
      // 31f: sipush 16383
      // 322: iload 11
      // 324: iand
      // 325: ldc 6291456
      // 327: iload 4
      // 329: iand
      // 32a: iadd
      // 32b: istore 11
      // 32d: iload 4
      // 32f: ldc 1764949655
      // 331: ishr
      // 332: istore 19
      // 334: aload 12
      // 336: iload 7
      // 338: iinc 7 1
      // 33b: aload 6
      // 33d: iload 11
      // 33f: ldc -1446967161
      // 341: ishr
      // 342: sipush 16256
      // 345: iload 10
      // 347: invokestatic ib.a (II)I
      // 34a: iadd
      // 34b: iaload
      // 34c: iload 19
      // 34e: iushr
      // 34f: iastore
      // 350: iload 4
      // 352: iload 5
      // 354: iadd
      // 355: istore 4
      // 357: iload 11
      // 359: iload 17
      // 35b: iadd
      // 35c: istore 11
      // 35e: iload 10
      // 360: iload 18
      // 362: iadd
      // 363: istore 10
      // 365: aload 12
      // 367: iload 7
      // 369: iinc 7 1
      // 36c: aload 6
      // 36e: iload 11
      // 370: ldc 1931622023
      // 372: ishr
      // 373: iload 10
      // 375: sipush 16256
      // 378: invokestatic ib.a (II)I
      // 37b: iadd
      // 37c: iaload
      // 37d: iload 19
      // 37f: iushr
      // 380: iastore
      // 381: iload 11
      // 383: iload 17
      // 385: iadd
      // 386: istore 11
      // 388: iload 10
      // 38a: iload 18
      // 38c: iadd
      // 38d: istore 10
      // 38f: aload 12
      // 391: iload 7
      // 393: iinc 7 1
      // 396: aload 6
      // 398: iload 11
      // 39a: ldc 821292839
      // 39c: ishr
      // 39d: sipush 16256
      // 3a0: iload 10
      // 3a2: invokestatic ib.a (II)I
      // 3a5: iadd
      // 3a6: iaload
      // 3a7: iload 19
      // 3a9: iushr
      // 3aa: iastore
      // 3ab: iload 10
      // 3ad: iload 18
      // 3af: iadd
      // 3b0: istore 10
      // 3b2: iload 11
      // 3b4: iload 17
      // 3b6: iadd
      // 3b7: istore 11
      // 3b9: aload 12
      // 3bb: iload 7
      // 3bd: iinc 7 1
      // 3c0: aload 6
      // 3c2: iload 11
      // 3c4: ldc -698624601
      // 3c6: ishr
      // 3c7: sipush 16256
      // 3ca: iload 10
      // 3cc: invokestatic ib.a (II)I
      // 3cf: iadd
      // 3d0: iaload
      // 3d1: iload 19
      // 3d3: iushr
      // 3d4: iastore
      // 3d5: iload 15
      // 3d7: istore 11
      // 3d9: iload 16
      // 3db: istore 10
      // 3dd: iload 0
      // 3de: iload 13
      // 3e0: iadd
      // 3e1: istore 0
      // 3e2: iload 3
      // 3e3: iload 9
      // 3e5: iadd
      // 3e6: istore 3
      // 3e7: iload 8
      // 3e9: iload 1
      // 3ea: iadd
      // 3eb: istore 8
      // 3ed: iload 3
      // 3ee: bipush -1
      // 3ef: ixor
      // 3f0: bipush -1
      // 3f1: iload 21
      // 3f3: ifne 462
      // 3f6: if_icmpne 400
      // 3f9: goto 3fd
      // 3fc: athrow
      // 3fd: goto 411
      // 400: iload 0
      // 401: iload 3
      // 402: idiv
      // 403: ldc 204957255
      // 405: ishl
      // 406: istore 16
      // 408: iload 8
      // 40a: iload 3
      // 40b: idiv
      // 40c: ldc -586213497
      // 40e: ishl
      // 40f: istore 15
      // 411: bipush -1
      // 412: iload 15
      // 414: bipush -1
      // 415: ixor
      // 416: if_icmplt 433
      // 419: sipush 16256
      // 41c: iload 15
      // 41e: if_icmplt 429
      // 421: goto 425
      // 424: athrow
      // 425: goto 436
      // 428: athrow
      // 429: sipush 16256
      // 42c: istore 15
      // 42e: iload 21
      // 430: ifeq 436
      // 433: bipush 0
      // 434: istore 15
      // 436: iload 10
      // 438: ineg
      // 439: iload 16
      // 43b: iadd
      // 43c: ldc -1259225180
      // 43e: ishr
      // 43f: istore 18
      // 441: iload 11
      // 443: ineg
      // 444: iload 15
      // 446: iadd
      // 447: ldc 1647676292
      // 449: ishr
      // 44a: istore 17
      // 44c: iinc 20 -1
      // 44f: iload 21
      // 451: ifeq 0cc
      // 454: bipush 0
      // 455: istore 20
      // 457: iload 20
      // 459: bipush -1
      // 45a: ixor
      // 45b: bipush 15
      // 45d: iload 14
      // 45f: iand
      // 460: bipush -1
      // 461: ixor
      // 462: if_icmple 4c7
      // 465: iload 21
      // 467: ifne 59d
      // 46a: goto 46e
      // 46d: athrow
      // 46e: iload 20
      // 470: bipush 3
      // 471: iand
      // 472: ifne 495
      // 475: goto 479
      // 478: athrow
      // 479: iload 4
      // 47b: ldc -851479113
      // 47d: ishr
      // 47e: istore 19
      // 480: iload 4
      // 482: ldc 6291456
      // 484: iand
      // 485: sipush 16383
      // 488: iload 11
      // 48a: iand
      // 48b: iadd
      // 48c: istore 11
      // 48e: iload 4
      // 490: iload 5
      // 492: iadd
      // 493: istore 4
      // 495: aload 12
      // 497: iload 7
      // 499: iinc 7 1
      // 49c: aload 6
      // 49e: iload 11
      // 4a0: ldc -1445030201
      // 4a2: ishr
      // 4a3: iload 10
      // 4a5: sipush 16256
      // 4a8: invokestatic ib.a (II)I
      // 4ab: iadd
      // 4ac: iaload
      // 4ad: iload 19
      // 4af: iushr
      // 4b0: iastore
      // 4b1: iload 10
      // 4b3: iload 18
      // 4b5: iadd
      // 4b6: istore 10
      // 4b8: iload 11
      // 4ba: iload 17
      // 4bc: iadd
      // 4bd: istore 11
      // 4bf: iinc 20 1
      // 4c2: iload 21
      // 4c4: ifeq 457
      // 4c7: goto 59d
      // 4ca: astore 15
      // 4cc: aload 15
      // 4ce: new java/lang/StringBuilder
      // 4d1: dup
      // 4d2: invokespecial java/lang/StringBuilder.<init> ()V
      // 4d5: getstatic gb.z [Ljava/lang/String;
      // 4d8: bipush 32
      // 4da: aaload
      // 4db: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 4de: iload 0
      // 4df: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 4e2: bipush 44
      // 4e4: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 4e7: iload 1
      // 4e8: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 4eb: bipush 44
      // 4ed: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 4f0: iload 2
      // 4f1: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 4f4: bipush 44
      // 4f6: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 4f9: iload 3
      // 4fa: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 4fd: bipush 44
      // 4ff: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 502: iload 4
      // 504: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 507: bipush 44
      // 509: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 50c: iload 5
      // 50e: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 511: bipush 44
      // 513: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 516: aload 6
      // 518: ifnull 525
      // 51b: getstatic gb.z [Ljava/lang/String;
      // 51e: bipush 11
      // 520: aaload
      // 521: goto 52b
      // 524: athrow
      // 525: getstatic gb.z [Ljava/lang/String;
      // 528: bipush 15
      // 52a: aaload
      // 52b: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 52e: bipush 44
      // 530: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 533: iload 7
      // 535: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 538: bipush 44
      // 53a: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 53d: iload 8
      // 53f: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 542: bipush 44
      // 544: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 547: iload 9
      // 549: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 54c: bipush 44
      // 54e: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 551: iload 10
      // 553: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 556: bipush 44
      // 558: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 55b: iload 11
      // 55d: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 560: bipush 44
      // 562: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 565: aload 12
      // 567: ifnull 574
      // 56a: getstatic gb.z [Ljava/lang/String;
      // 56d: bipush 11
      // 56f: aaload
      // 570: goto 57a
      // 573: athrow
      // 574: getstatic gb.z [Ljava/lang/String;
      // 577: bipush 15
      // 579: aaload
      // 57a: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 57d: bipush 44
      // 57f: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 582: iload 13
      // 584: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 587: bipush 44
      // 589: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 58c: iload 14
      // 58e: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 591: bipush 41
      // 593: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 596: invokevirtual java/lang/StringBuilder.toString ()Ljava/lang/String;
      // 599: invokestatic i.a (Ljava/lang/Throwable;Ljava/lang/String;)Lla;
      // 59c: athrow
      // 59d: return
   }

   static final i[] a(int var0) {
      try {
         try {
            if (var0 <= 37) {
               n = (v)null;
            }
         } catch (RuntimeException var2) {
            throw var2;
         }

         t++;
         return new i[]{eb.e, fb.h, f.b};
      } catch (RuntimeException var3) {
         throw i.a(var3, z[19] + var0 + ')');
      }
   }

   @Override
   final Socket a(byte var1) throws IOException {
      boolean var13 = client.vh;

      try {
         r++;
         boolean var2 = Boolean.parseBoolean(System.getProperty(z[2]));

         try {
            if (!var2) {
               System.setProperty(z[2], z[3]);
            }
         } catch (URISyntaxException var14) {
            throw var14;
         }

         boolean var10000;
         label121: {
            try {
               if (~this.f == -444) {
                  var10000 = true;
                  break label121;
               }
            } catch (URISyntaxException var22) {
               throw var22;
            }

            var10000 = false;
         }

         boolean var5 = var10000;

         List var3;
         List var4;
         try {
            URI var10001;
            StringBuilder var10003;
            String var10004;
            label102: {
               try {
                  var24 = this.q;
                  var10003 = new StringBuilder();
                  if (var5) {
                     var10004 = z[0];
                     break label102;
                  }
               } catch (URISyntaxException var19) {
                  throw var19;
               }

               var10004 = z[4];
            }

            var10001 = new URI(var10003.append(var10004).append(z[1]).append(this.h).toString());
            var3 = var24.select(var10001);

            label110: {
               try {
                  var25 = this.q;
                  var10003 = new StringBuilder();
                  if (!var5) {
                     var10004 = z[0];
                     break label110;
                  }
               } catch (URISyntaxException var20) {
                  throw var20;
               }

               var10004 = z[4];
            }

            var10001 = new URI(var10003.append(var10004).append(z[1]).append(this.h).toString());
            var4 = var25.select(var10001);
         } catch (URISyntaxException var21) {
            return this.a(false);
         }

         var3.addAll(var4);
         Object[] var6 = var3.toArray();
         fa var7 = null;
         Object[] var8 = var6;
         if (var1 != 50) {
            return (Socket)null;
         } else {
            int var9 = 0;

            while (~var8.length < ~var9) {
               Object var10 = var8[var9];
               Proxy var11 = (Proxy)var10;

               try {
                  label130: {
                     Socket var12 = this.a(var11, 16256);

                     try {
                        if (var13) {
                           return var12;
                        }

                        if (var12 == null) {
                           break label130;
                        }
                     } catch (URISyntaxException var16) {
                        throw var16;
                     }

                     return var12;
                  }
               } catch (fa var17) {
                  var7 = var17;
               } catch (IOException var18) {
               }

               var9++;
               if (var13) {
                  break;
               }
            }

            try {
               if (null != var7) {
                  throw var7;
               }
            } catch (URISyntaxException var15) {
               throw var15;
            }

            return this.a(false);
         }
      } catch (RuntimeException var23) {
         throw i.a(var23, z[5] + var1 + ')');
      }
   }

   static final String a(boolean param0, Throwable param1) throws IOException {
      // $VF: Couldn't be decompiled
      // Please report this to the Vineflower issue tracker, at https://github.com/Vineflower/vineflower/issues with a copy of the class file (if you have the rights to distribute it!)
      // java.lang.RuntimeException: parsing failure!
      //   at org.jetbrains.java.decompiler.modules.decompiler.decompose.DomHelper.parseGraph(DomHelper.java:211)
      //   at org.jetbrains.java.decompiler.main.rels.MethodProcessor.codeToJava(MethodProcessor.java:166)
      //
      // Bytecode:
      // 000: getstatic client.vh Z
      // 003: istore 13
      // 005: getstatic gb.u I
      // 008: bipush 1
      // 009: iadd
      // 00a: putstatic gb.u I
      // 00d: aload 1
      // 00e: instanceof la
      // 011: ifeq 03e
      // 014: aload 1
      // 015: checkcast la
      // 018: astore 3
      // 019: aload 3
      // 01a: getfield la.e Ljava/lang/Throwable;
      // 01d: astore 1
      // 01e: new java/lang/StringBuilder
      // 021: dup
      // 022: invokespecial java/lang/StringBuilder.<init> ()V
      // 025: aload 3
      // 026: getfield la.h Ljava/lang/String;
      // 029: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 02c: getstatic gb.z [Ljava/lang/String;
      // 02f: bipush 16
      // 031: aaload
      // 032: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 035: invokevirtual java/lang/StringBuilder.toString ()Ljava/lang/String;
      // 038: astore 2
      // 039: iload 13
      // 03b: ifeq 041
      // 03e: ldc ""
      // 040: astore 2
      // 041: new java/io/StringWriter
      // 044: dup
      // 045: invokespecial java/io/StringWriter.<init> ()V
      // 048: astore 3
      // 049: new java/io/PrintWriter
      // 04c: dup
      // 04d: aload 3
      // 04e: invokespecial java/io/PrintWriter.<init> (Ljava/io/Writer;)V
      // 051: astore 4
      // 053: aload 1
      // 054: aload 4
      // 056: invokevirtual java/lang/Throwable.printStackTrace (Ljava/io/PrintWriter;)V
      // 059: aload 4
      // 05b: invokevirtual java/io/PrintWriter.close ()V
      // 05e: aload 3
      // 05f: invokevirtual java/io/StringWriter.toString ()Ljava/lang/String;
      // 062: astore 5
      // 064: new java/io/BufferedReader
      // 067: dup
      // 068: new java/io/StringReader
      // 06b: dup
      // 06c: aload 5
      // 06e: invokespecial java/io/StringReader.<init> (Ljava/lang/String;)V
      // 071: invokespecial java/io/BufferedReader.<init> (Ljava/io/Reader;)V
      // 074: astore 6
      // 076: iload 0
      // 077: ifeq 085
      // 07a: aconst_null
      // 07b: checkcast [I
      // 07e: putstatic gb.s [I
      // 081: goto 085
      // 084: athrow
      // 085: aload 6
      // 087: invokevirtual java/io/BufferedReader.readLine ()Ljava/lang/String;
      // 08a: astore 7
      // 08c: aload 6
      // 08e: invokevirtual java/io/BufferedReader.readLine ()Ljava/lang/String;
      // 091: astore 8
      // 093: aconst_null
      // 094: aload 8
      // 096: if_acmpne 09e
      // 099: iload 13
      // 09b: ifeq 169
      // 09e: aload 8
      // 0a0: bipush 40
      // 0a2: invokevirtual java/lang/String.indexOf (I)I
      // 0a5: istore 9
      // 0a7: aload 8
      // 0a9: bipush 41
      // 0ab: iload 9
      // 0ad: bipush -1
      // 0ae: isub
      // 0af: invokevirtual java/lang/String.indexOf (II)I
      // 0b2: istore 10
      // 0b4: bipush 0
      // 0b5: iload 9
      // 0b7: bipush -1
      // 0b8: ixor
      // 0b9: if_icmpeq 0cb
      // 0bc: aload 8
      // 0be: bipush 0
      // 0bf: iload 9
      // 0c1: invokevirtual java/lang/String.substring (II)Ljava/lang/String;
      // 0c4: astore 11
      // 0c6: iload 13
      // 0c8: ifeq 0cf
      // 0cb: aload 8
      // 0cd: astore 11
      // 0cf: aload 11
      // 0d1: invokevirtual java/lang/String.trim ()Ljava/lang/String;
      // 0d4: astore 11
      // 0d6: aload 11
      // 0d8: bipush 1
      // 0d9: aload 11
      // 0db: bipush 32
      // 0dd: invokevirtual java/lang/String.lastIndexOf (I)I
      // 0e0: iadd
      // 0e1: invokevirtual java/lang/String.substring (I)Ljava/lang/String;
      // 0e4: astore 11
      // 0e6: aload 11
      // 0e8: aload 11
      // 0ea: bipush 9
      // 0ec: invokevirtual java/lang/String.lastIndexOf (I)I
      // 0ef: bipush 1
      // 0f0: iadd
      // 0f1: invokevirtual java/lang/String.substring (I)Ljava/lang/String;
      // 0f4: astore 11
      // 0f6: new java/lang/StringBuilder
      // 0f9: dup
      // 0fa: invokespecial java/lang/StringBuilder.<init> ()V
      // 0fd: aload 2
      // 0fe: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 101: aload 11
      // 103: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 106: invokevirtual java/lang/StringBuilder.toString ()Ljava/lang/String;
      // 109: astore 2
      // 10a: iload 9
      // 10c: bipush -1
      // 10d: ixor
      // 10e: ifeq 150
      // 111: bipush -1
      // 112: iload 10
      // 114: if_icmpne 11f
      // 117: goto 11b
      // 11a: athrow
      // 11b: goto 150
      // 11e: athrow
      // 11f: aload 8
      // 121: getstatic gb.z [Ljava/lang/String;
      // 124: bipush 17
      // 126: aaload
      // 127: iload 9
      // 129: invokevirtual java/lang/String.indexOf (Ljava/lang/String;I)I
      // 12c: istore 12
      // 12e: iload 12
      // 130: iflt 150
      // 133: new java/lang/StringBuilder
      // 136: dup
      // 137: invokespecial java/lang/StringBuilder.<init> ()V
      // 13a: aload 2
      // 13b: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 13e: aload 8
      // 140: bipush 5
      // 141: iload 12
      // 143: iadd
      // 144: iload 10
      // 146: invokevirtual java/lang/String.substring (II)Ljava/lang/String;
      // 149: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 14c: invokevirtual java/lang/StringBuilder.toString ()Ljava/lang/String;
      // 14f: astore 2
      // 150: new java/lang/StringBuilder
      // 153: dup
      // 154: invokespecial java/lang/StringBuilder.<init> ()V
      // 157: aload 2
      // 158: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 15b: bipush 32
      // 15d: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 160: invokevirtual java/lang/StringBuilder.toString ()Ljava/lang/String;
      // 163: astore 2
      // 164: iload 13
      // 166: ifeq 08c
      // 169: new java/lang/StringBuilder
      // 16c: dup
      // 16d: invokespecial java/lang/StringBuilder.<init> ()V
      // 170: aload 2
      // 171: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 174: getstatic gb.z [Ljava/lang/String;
      // 177: bipush 18
      // 179: aaload
      // 17a: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 17d: aload 7
      // 17f: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 182: invokevirtual java/lang/StringBuilder.toString ()Ljava/lang/String;
      // 185: astore 2
      // 186: aload 2
      // 187: areturn
      // 188: astore 2
      // 189: aload 2
      // 18a: athrow
   }

   private final Socket a(Proxy var1, int var2) throws IOException {
      try {
         try {
            l++;
            if (var1.type() == Type.DIRECT) {
               return this.a(false);
            }
         } catch (Exception var20) {
            throw var20;
         }

         SocketAddress var22 = var1.address();

         try {
            if (!(var22 instanceof InetSocketAddress)) {
               return null;
            }
         } catch (Exception var19) {
            throw var19;
         }

         try {
            if (var2 != 16256) {
               p = 123;
            }
         } catch (Exception var14) {
            throw var14;
         }

         InetSocketAddress var4 = (InetSocketAddress)var22;
         if (var1.type() == Type.HTTP) {
            String var23 = null;

            try {
               Class var6 = Class.forName(z[6]);
               Method var7 = var6.getDeclaredMethod(z[9], String.class, int.class);
               var7.setAccessible(true);
               Object var8 = var7.invoke(null, var4.getHostName(), new Integer(var4.getPort()));

               try {
                  if (null == var8) {
                     return this.a(var4.getPort(), 1514, var23, var4.getHostName());
                  }
               } catch (Exception var16) {
                  throw var16;
               }

               Method var9 = var6.getDeclaredMethod(z[12]);
               var9.setAccessible(true);
               if ((Boolean)var9.invoke(var8)) {
                  Method var10 = var6.getDeclaredMethod(z[10]);
                  var10.setAccessible(true);
                  Method var11 = var6.getDeclaredMethod(z[13], URL.class, String.class);
                  var11.setAccessible(true);
                  String var12 = (String)var10.invoke(var8);
                  String var13 = (String)var11.invoke(var8, new URL(z[8] + this.h + "/"), z[0]);
                  var23 = var12 + z[7] + var13;
               }
            } catch (Exception var17) {
            }

            return this.a(var4.getPort(), 1514, var23, var4.getHostName());
         } else {
            label82: {
               try {
                  if (var1.type() == Type.SOCKS) {
                     break label82;
                  }
               } catch (Exception var18) {
                  throw var18;
               }

               return null;
            }

            Socket var5 = new Socket(var1);
            var5.connect(new InetSocketAddress(this.h, this.f));
            return var5;
         }
      } catch (RuntimeException var21) {
         RuntimeException var3 = var21;

         RuntimeException var10000;
         StringBuilder var10001;
         try {
            var10000 = var3;
            var10001 = new StringBuilder().append(z[14]);
            if (var1 != null) {
               throw i.a(var3, var10001.append(z[11]).append(',').append(var2).append(')').toString());
            }
         } catch (Exception var15) {
            throw var15;
         }

         throw i.a(var10000, var10001.append(z[15]).append(',').append(var2).append(')').toString());
      }
   }

   private static char[] z(String var0) {
      char[] var10000 = var0.toCharArray();
      if (var10000.length < 2) {
         var10000[0] = (char)(var10000[0] ^ 'w');
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
               var10005 = 116;
               break;
            case 1:
               var10005 = 60;
               break;
            case 2:
               var10005 = 54;
               break;
            case 3:
               var10005 = 39;
               break;
            default:
               var10005 = 119;
         }

         var10001[var1] = (char)(var10004 ^ var10005);
      }

      return new String(var10001).intern();
   }
}
