import java.awt.EventQueue;
import java.awt.Toolkit;
import java.io.File;
import java.lang.reflect.Method;
import java.net.URL;

final class c implements Runnable {
   private boolean e;
   private q t;
   private Thread r;
   private Object x;
   private static String m;
   private static String p;
   private boolean j = false;
   private Object i;
   EventQueue n;
   static Method u;
   private static int b;
   private g a = null;
   private boolean c;
   private wa w;
   private static String g;
   static Method y;
   d f = null;
   d s = null;
   private static String h;
   d v;
   static String q;
   private d[] l;
   private static volatile long d = 0L;
   private g o;
   static String k;
   private static final String[] z = new String[]{
      z(z("lHu\u0004:Y")),
      z(z("e\u0013=\u0016+hM}\u00161)")),
      z(z(")[a\u0002#eAwN")),
      z(z("e\u0013=")),
      z(z("(Ms\u0015")),
      z(z("e\u0013=\u00131eHq\t')")),
      z(z("YY`\u0004$c[w\u000f!cZ")),
      z(z("Y^{\u0011lbHf")),
      z(z("t^")),
      z(z("e\u0013=\u0016+hGfN")),
      z(z(")]\u007f\u0011m")),
      z(z("Y[qO&g]")),
      z(z("LHu\u0004:&og\r.&zq\u0013'cG")),
      z(z("cQ{\u0015")),
      z(z("gKq\u0005'`Nz\b(mE\u007f\u000f-vX`\u00126s_e\u0019;|hP\"\u0006CoU)\u000bLb^,\fIyC3\u0011R|D6\u001a_s\"Pp5\u001d'Wu>\u0010-G\u007f*\u00077JoY\n(Nh")),
      z(z("nH")),
      z(z("n]f\u00111<\u0006=")),
      z(z("cGf\u00040")),
      z(z("n]f\u0011x)\u0006")),
      z(z("kFd\u0004/i\\a\u0004")),
      z(z("uLf\u00027u]}\f!s[a\u000e0")),
      z(z("eDvAme\ta\u0015#t]2C($\t0")),
      z(z("j@a\u0015/iMw\u0012")),
      z(z("q@|")),
      z(z("uA}\u0016!s[a\u000e0")),
      z(z("uLf'-e\\a50g_w\u00131gEY\u0004;ul|\u0000 jLv")),
      z(z("kH{\u000f\u001d`@~\u0004\u001deHq\t'(@v\u0019p3\u001c")),
      z(z("tK")),
      z(z("lHd\u0000lg^fO\u0001iGf\u0000+hL`")),
      z(z("iZ<\u000f#kL")),
      z(z("iZ<\u0017'tZ{\u000e,")),
      z(z("sZw\u0013lnF\u007f\u0004")),
      z(z("lHd\u0000lpL|\u0005-t")),
      z(z("kH{\u000f\u001d`@~\u0004\u001deHq\t'(Ms\u0015p")),
      z(z("iZ<\u00000eA")),
      z(z("tH|\u0005-k\u0007v\u00006")),
      z(z("lHd\u0000lpL`\u0012+iG")),
      z(z("x\u0006")),
      z(z("uLf'-e\\a\";eEw3-i]")),
      z(z("SGy\u000f-qG")),
      z(z("lHd\u0000lg^fO\u0001iDb\u000e,cGf")),
      z(z("kH{\u000f\u001d`@~\u0004\u001deHq\t'(@v\u0019")),
      z(z("7\u0007#")),
      z(z("k@q\u0013-uFt\u0015"))
   };

   final g a(String var1, int var2, int var3) {
      try {
         return var3 > -66 ? (g)null : this.a(var2, (byte)81, var1, false);
      } catch (RuntimeException var5) {
         throw var5;
      }
   }

   final g a(boolean var1, Runnable var2, int var3) {
      try {
         try {
            if (!var1) {
               this.a(-34, 71, (byte)60, 103, (Object)null);
            }
         } catch (RuntimeException var5) {
            throw var5;
         }

         return this.a(2, 0, (byte)-21, var3, var2);
      } catch (RuntimeException var6) {
         throw var6;
      }
   }

   private static final d a(int var0, String var1, boolean var2, String var3) {
      try {
         String var4;
         label65: {
            label64: {
               label63: {
                  try {
                     if (-34 != ~var0) {
                        if (~var0 == -35) {
                           break label64;
                        }
                        break label63;
                     }
                  } catch (Exception var11) {
                     throw var11;
                  }

                  var4 = z[0] + var1 + z[6] + var3 + z[11];
                  break label65;
               }

               var4 = z[0] + var1 + z[6] + var3 + z[4];
               break label65;
            }

            var4 = z[0] + var1 + z[6] + var3 + z[7];
         }

         if (var2) {
            return (d)null;
         } else {
            String[] var5 = new String[]{z[5], z[2], m, z[1], z[9], z[3], z[10], ""};

            for (int var6 = 0; var6 < var5.length; var6++) {
               String var7 = var5[var6];

               try {
                  if (0 < var7.length() && !new File(var7).exists()) {
                     continue;
                  }
               } catch (Exception var10) {
                  throw var10;
               }

               try {
                  return new d(new File(var7, var4), z[8], 10000L);
               } catch (Exception var9) {
               }
            }

            return null;
         }
      } catch (RuntimeException var12) {
         throw var12;
      }
   }

   @Override
   public final void run() {
      // $VF: Couldn't be decompiled
      // Please report this to the Vineflower issue tracker, at https://github.com/Vineflower/vineflower/issues with a copy of the class file (if you have the rights to distribute it!)
      // java.lang.RuntimeException: parsing failure!
      //   at org.jetbrains.java.decompiler.modules.decompiler.decompose.DomHelper.parseGraph(DomHelper.java:211)
      //   at org.jetbrains.java.decompiler.main.rels.MethodProcessor.codeToJava(MethodProcessor.java:166)
      //
      // Bytecode:
      // 000: aload 0
      // 001: dup
      // 002: astore 2
      // 003: monitorenter
      // 004: aload 0
      // 005: getfield c.e Z
      // 008: ifeq 00e
      // 00b: aload 2
      // 00c: monitorexit
      // 00d: return
      // 00e: aconst_null
      // 00f: aload 0
      // 010: getfield c.a Lg;
      // 013: if_acmpeq 03a
      // 016: aload 0
      // 017: getfield c.a Lg;
      // 01a: astore 1
      // 01b: aload 0
      // 01c: aload 0
      // 01d: getfield c.a Lg;
      // 020: getfield g.a Lg;
      // 023: putfield c.a Lg;
      // 026: aconst_null
      // 027: aload 0
      // 028: getfield c.a Lg;
      // 02b: if_acmpeq 032
      // 02e: goto 045
      // 031: athrow
      // 032: aload 0
      // 033: aconst_null
      // 034: putfield c.o Lg;
      // 037: goto 045
      // 03a: aload 0
      // 03b: invokevirtual java/lang/Object.wait ()V
      // 03e: goto 004
      // 041: astore 3
      // 042: goto 004
      // 045: aload 2
      // 046: monitorexit
      // 047: goto 051
      // 04a: astore 4
      // 04c: aload 2
      // 04d: monitorexit
      // 04e: aload 4
      // 050: athrow
      // 051: aload 1
      // 052: getfield g.g I
      // 055: istore 2
      // 056: iload 2
      // 057: bipush -1
      // 058: ixor
      // 059: bipush -2
      // 05b: if_icmpne 095
      // 05e: bipush 0
      // 05f: invokestatic p.a (I)J
      // 062: getstatic c.d J
      // 065: lcmp
      // 066: iflt 071
      // 069: goto 06d
      // 06c: athrow
      // 06d: goto 079
      // 070: athrow
      // 071: new java/io/IOException
      // 074: dup
      // 075: invokespecial java/io/IOException.<init> ()V
      // 078: athrow
      // 079: aload 1
      // 07a: new java/net/Socket
      // 07d: dup
      // 07e: aload 1
      // 07f: getfield g.f Ljava/lang/Object;
      // 082: checkcast java/lang/String
      // 085: invokestatic java/net/InetAddress.getByName (Ljava/lang/String;)Ljava/net/InetAddress;
      // 088: aload 1
      // 089: getfield g.e I
      // 08c: invokespecial java/net/Socket.<init> (Ljava/net/InetAddress;I)V
      // 08f: putfield g.d Ljava/lang/Object;
      // 092: goto 6f0
      // 095: iload 2
      // 096: bipush 22
      // 098: if_icmpeq 6ac
      // 09b: bipush 2
      // 09c: iload 2
      // 09d: if_icmpne 0cc
      // 0a0: goto 0a4
      // 0a3: athrow
      // 0a4: new java/lang/Thread
      // 0a7: dup
      // 0a8: aload 1
      // 0a9: getfield g.f Ljava/lang/Object;
      // 0ac: checkcast java/lang/Runnable
      // 0af: invokespecial java/lang/Thread.<init> (Ljava/lang/Runnable;)V
      // 0b2: astore 3
      // 0b3: aload 3
      // 0b4: bipush 1
      // 0b5: invokevirtual java/lang/Thread.setDaemon (Z)V
      // 0b8: aload 3
      // 0b9: invokevirtual java/lang/Thread.start ()V
      // 0bc: aload 3
      // 0bd: aload 1
      // 0be: getfield g.e I
      // 0c1: invokevirtual java/lang/Thread.setPriority (I)V
      // 0c4: aload 1
      // 0c5: aload 3
      // 0c6: putfield g.d Ljava/lang/Object;
      // 0c9: goto 6f0
      // 0cc: iload 2
      // 0cd: bipush 4
      // 0ce: if_icmpeq 678
      // 0d1: bipush 8
      // 0d3: iload 2
      // 0d4: if_icmpne 126
      // 0d7: goto 0db
      // 0da: athrow
      // 0db: aload 1
      // 0dc: getfield g.f Ljava/lang/Object;
      // 0df: checkcast [Ljava/lang/Object;
      // 0e2: checkcast [Ljava/lang/Object;
      // 0e5: astore 3
      // 0e6: aload 0
      // 0e7: getfield c.j Z
      // 0ea: ifne 0f1
      // 0ed: goto 107
      // 0f0: athrow
      // 0f1: aconst_null
      // 0f2: aload 3
      // 0f3: bipush 0
      // 0f4: aaload
      // 0f5: checkcast java/lang/Class
      // 0f8: invokevirtual java/lang/Class.getClassLoader ()Ljava/lang/ClassLoader;
      // 0fb: if_acmpne 107
      // 0fe: new java/lang/SecurityException
      // 101: dup
      // 102: invokespecial java/lang/SecurityException.<init> ()V
      // 105: athrow
      // 106: athrow
      // 107: aload 1
      // 108: aload 3
      // 109: bipush 0
      // 10a: aaload
      // 10b: checkcast java/lang/Class
      // 10e: aload 3
      // 10f: bipush 1
      // 110: aaload
      // 111: checkcast java/lang/String
      // 114: aload 3
      // 115: bipush 2
      // 116: aaload
      // 117: checkcast [Ljava/lang/Class;
      // 11a: checkcast [Ljava/lang/Class;
      // 11d: invokevirtual java/lang/Class.getDeclaredMethod (Ljava/lang/String;[Ljava/lang/Class;)Ljava/lang/reflect/Method;
      // 120: putfield g.d Ljava/lang/Object;
      // 123: goto 6f0
      // 126: bipush -10
      // 128: iload 2
      // 129: bipush -1
      // 12a: ixor
      // 12b: if_icmpne 172
      // 12e: aload 1
      // 12f: getfield g.f Ljava/lang/Object;
      // 132: checkcast [Ljava/lang/Object;
      // 135: checkcast [Ljava/lang/Object;
      // 138: astore 3
      // 139: aload 0
      // 13a: getfield c.j Z
      // 13d: ifne 144
      // 140: goto 15c
      // 143: athrow
      // 144: aload 3
      // 145: bipush 0
      // 146: aaload
      // 147: checkcast java/lang/Class
      // 14a: invokevirtual java/lang/Class.getClassLoader ()Ljava/lang/ClassLoader;
      // 14d: ifnull 154
      // 150: goto 15c
      // 153: athrow
      // 154: new java/lang/SecurityException
      // 157: dup
      // 158: invokespecial java/lang/SecurityException.<init> ()V
      // 15b: athrow
      // 15c: aload 1
      // 15d: aload 3
      // 15e: bipush 0
      // 15f: aaload
      // 160: checkcast java/lang/Class
      // 163: aload 3
      // 164: bipush 1
      // 165: aaload
      // 166: checkcast java/lang/String
      // 169: invokevirtual java/lang/Class.getDeclaredField (Ljava/lang/String;)Ljava/lang/reflect/Field;
      // 16c: putfield g.d Ljava/lang/Object;
      // 16f: goto 6f0
      // 172: iload 2
      // 173: bipush 18
      // 175: if_icmpeq 665
      // 178: bipush 19
      // 17a: iload 2
      // 17b: if_icmpeq 64b
      // 17e: goto 182
      // 181: athrow
      // 182: aload 0
      // 183: getfield c.j Z
      // 186: ifeq 641
      // 189: goto 18d
      // 18c: athrow
      // 18d: bipush 3
      // 18e: iload 2
      // 18f: if_icmpne 20a
      // 192: goto 196
      // 195: athrow
      // 196: getstatic c.d J
      // 199: bipush 0
      // 19a: invokestatic p.a (I)J
      // 19d: lcmp
      // 19e: ifle 1ae
      // 1a1: goto 1a5
      // 1a4: athrow
      // 1a5: new java/io/IOException
      // 1a8: dup
      // 1a9: invokespecial java/io/IOException.<init> ()V
      // 1ac: athrow
      // 1ad: athrow
      // 1ae: new java/lang/StringBuilder
      // 1b1: dup
      // 1b2: invokespecial java/lang/StringBuilder.<init> ()V
      // 1b5: sipush 255
      // 1b8: aload 1
      // 1b9: getfield g.e I
      // 1bc: ldc -182496008
      // 1be: ishr
      // 1bf: iand
      // 1c0: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 1c3: ldc "."
      // 1c5: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 1c8: sipush 255
      // 1cb: aload 1
      // 1cc: getfield g.e I
      // 1cf: ldc 954736400
      // 1d1: ishr
      // 1d2: iand
      // 1d3: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 1d6: ldc "."
      // 1d8: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 1db: ldc 65472
      // 1dd: aload 1
      // 1de: getfield g.e I
      // 1e1: iand
      // 1e2: ldc -58046680
      // 1e4: ishr
      // 1e5: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 1e8: ldc "."
      // 1ea: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 1ed: sipush 255
      // 1f0: aload 1
      // 1f1: getfield g.e I
      // 1f4: iand
      // 1f5: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 1f8: invokevirtual java/lang/StringBuilder.toString ()Ljava/lang/String;
      // 1fb: astore 3
      // 1fc: aload 1
      // 1fd: aload 3
      // 1fe: invokestatic java/net/InetAddress.getByName (Ljava/lang/String;)Ljava/net/InetAddress;
      // 201: invokevirtual java/net/InetAddress.getHostName ()Ljava/lang/String;
      // 204: putfield g.d Ljava/lang/Object;
      // 207: goto 6f0
      // 20a: iload 2
      // 20b: bipush 21
      // 20d: if_icmpeq 619
      // 210: bipush 5
      // 211: iload 2
      // 212: if_icmpeq 5d9
      // 215: goto 219
      // 218: athrow
      // 219: iload 2
      // 21a: bipush -1
      // 21b: ixor
      // 21c: bipush -7
      // 21e: if_icmpeq 506
      // 221: goto 225
      // 224: athrow
      // 225: iload 2
      // 226: bipush 7
      // 228: if_icmpne 272
      // 22b: goto 22f
      // 22e: athrow
      // 22f: aload 0
      // 230: getfield c.c Z
      // 233: ifne 260
      // 236: goto 23a
      // 239: athrow
      // 23a: getstatic c.z [Ljava/lang/String;
      // 23d: bipush 15
      // 23f: aaload
      // 240: invokestatic java/lang/Class.forName (Ljava/lang/String;)Ljava/lang/Class;
      // 243: getstatic c.z [Ljava/lang/String;
      // 246: bipush 13
      // 248: aaload
      // 249: bipush 0
      // 24a: anewarray 257
      // 24d: invokevirtual java/lang/Class.getMethod (Ljava/lang/String;[Ljava/lang/Class;)Ljava/lang/reflect/Method;
      // 250: aload 0
      // 251: getfield c.i Ljava/lang/Object;
      // 254: bipush 0
      // 255: anewarray 261
      // 258: invokevirtual java/lang/reflect/Method.invoke (Ljava/lang/Object;[Ljava/lang/Object;)Ljava/lang/Object;
      // 25b: pop
      // 25c: goto 6f0
      // 25f: athrow
      // 260: aload 0
      // 261: getfield c.w Lwa;
      // 264: aload 1
      // 265: getfield g.f Ljava/lang/Object;
      // 268: checkcast java/awt/Frame
      // 26b: bipush 0
      // 26c: invokevirtual wa.a (Ljava/awt/Frame;I)V
      // 26f: goto 6f0
      // 272: bipush -13
      // 274: iload 2
      // 275: bipush -1
      // 276: ixor
      // 277: if_icmpeq 4ec
      // 27a: bipush -14
      // 27c: iload 2
      // 27d: bipush -1
      // 27e: ixor
      // 27f: if_icmpne 29f
      // 282: goto 286
      // 285: athrow
      // 286: getstatic c.b I
      // 289: ldc ""
      // 28b: bipush 0
      // 28c: aload 1
      // 28d: getfield g.f Ljava/lang/Object;
      // 290: checkcast java/lang/String
      // 293: invokestatic c.a (ILjava/lang/String;ZLjava/lang/String;)Ld;
      // 296: astore 3
      // 297: aload 1
      // 298: aload 3
      // 299: putfield g.d Ljava/lang/Object;
      // 29c: goto 6f0
      // 29f: aload 0
      // 2a0: getfield c.j Z
      // 2a3: ifeq 2b0
      // 2a6: iload 2
      // 2a7: bipush 14
      // 2a9: if_icmpeq 485
      // 2ac: goto 2b0
      // 2af: athrow
      // 2b0: aload 0
      // 2b1: getfield c.j Z
      // 2b4: ifeq 2c7
      // 2b7: goto 2bb
      // 2ba: athrow
      // 2bb: iload 2
      // 2bc: bipush -1
      // 2bd: ixor
      // 2be: bipush -16
      // 2c0: if_icmpeq 418
      // 2c3: goto 2c7
      // 2c6: athrow
      // 2c7: aload 0
      // 2c8: getfield c.c Z
      // 2cb: ifne 351
      // 2ce: goto 2d2
      // 2d1: athrow
      // 2d2: iload 2
      // 2d3: bipush 17
      // 2d5: if_icmpne 351
      // 2d8: goto 2dc
      // 2db: athrow
      // 2dc: aload 1
      // 2dd: getfield g.f Ljava/lang/Object;
      // 2e0: checkcast [Ljava/lang/Object;
      // 2e3: checkcast [Ljava/lang/Object;
      // 2e6: astore 3
      // 2e7: ldc "j"
      // 2e9: invokestatic java/lang/Class.forName (Ljava/lang/String;)Ljava/lang/Class;
      // 2ec: getstatic c.z [Ljava/lang/String;
      // 2ef: bipush 20
      // 2f1: aaload
      // 2f2: bipush 5
      // 2f3: anewarray 257
      // 2f6: dup
      // 2f7: bipush 0
      // 2f8: ldc java/awt/Component
      // 2fa: aastore
      // 2fb: dup
      // 2fc: bipush 1
      // 2fd: ldc [I
      // 2ff: aastore
      // 300: dup
      // 301: bipush 2
      // 302: getstatic java/lang/Integer.TYPE Ljava/lang/Class;
      // 305: aastore
      // 306: dup
      // 307: bipush 3
      // 308: getstatic java/lang/Integer.TYPE Ljava/lang/Class;
      // 30b: aastore
      // 30c: dup
      // 30d: bipush 4
      // 30e: ldc java/awt/Point
      // 310: aastore
      // 311: invokevirtual java/lang/Class.getDeclaredMethod (Ljava/lang/String;[Ljava/lang/Class;)Ljava/lang/reflect/Method;
      // 314: aload 0
      // 315: getfield c.x Ljava/lang/Object;
      // 318: bipush 5
      // 319: anewarray 261
      // 31c: dup
      // 31d: bipush 0
      // 31e: aload 3
      // 31f: bipush 0
      // 320: aaload
      // 321: aastore
      // 322: dup
      // 323: bipush 1
      // 324: aload 3
      // 325: bipush 1
      // 326: aaload
      // 327: aastore
      // 328: dup
      // 329: bipush 2
      // 32a: new java/lang/Integer
      // 32d: dup
      // 32e: aload 1
      // 32f: getfield g.e I
      // 332: invokespecial java/lang/Integer.<init> (I)V
      // 335: aastore
      // 336: dup
      // 337: bipush 3
      // 338: new java/lang/Integer
      // 33b: dup
      // 33c: aload 1
      // 33d: getfield g.c I
      // 340: invokespecial java/lang/Integer.<init> (I)V
      // 343: aastore
      // 344: dup
      // 345: bipush 4
      // 346: aload 3
      // 347: bipush 2
      // 348: aaload
      // 349: aastore
      // 34a: invokevirtual java/lang/reflect/Method.invoke (Ljava/lang/Object;[Ljava/lang/Object;)Ljava/lang/Object;
      // 34d: pop
      // 34e: goto 6f0
      // 351: bipush 16
      // 353: iload 2
      // 354: if_icmpeq 362
      // 357: new java/lang/Exception
      // 35a: dup
      // 35b: ldc ""
      // 35d: invokespecial java/lang/Exception.<init> (Ljava/lang/String;)V
      // 360: athrow
      // 361: athrow
      // 362: getstatic c.g Ljava/lang/String;
      // 365: getstatic c.z [Ljava/lang/String;
      // 368: bipush 23
      // 36a: aaload
      // 36b: invokevirtual java/lang/String.startsWith (Ljava/lang/String;)Z
      // 36e: ifeq 375
      // 371: goto 37d
      // 374: athrow
      // 375: new java/lang/Exception
      // 378: dup
      // 379: invokespecial java/lang/Exception.<init> ()V
      // 37c: athrow
      // 37d: aload 1
      // 37e: getfield g.f Ljava/lang/Object;
      // 381: checkcast java/lang/String
      // 384: astore 3
      // 385: aload 3
      // 386: getstatic c.z [Ljava/lang/String;
      // 389: bipush 18
      // 38b: aaload
      // 38c: invokevirtual java/lang/String.startsWith (Ljava/lang/String;)Z
      // 38f: ifne 3ac
      // 392: aload 3
      // 393: getstatic c.z [Ljava/lang/String;
      // 396: bipush 16
      // 398: aaload
      // 399: invokevirtual java/lang/String.startsWith (Ljava/lang/String;)Z
      // 39c: ifne 3ac
      // 39f: goto 3a3
      // 3a2: athrow
      // 3a3: new java/lang/Exception
      // 3a6: dup
      // 3a7: invokespecial java/lang/Exception.<init> ()V
      // 3aa: athrow
      // 3ab: athrow
      // 3ac: getstatic c.z [Ljava/lang/String;
      // 3af: bipush 14
      // 3b1: aaload
      // 3b2: astore 4
      // 3b4: bipush 0
      // 3b5: istore 5
      // 3b7: iload 5
      // 3b9: aload 3
      // 3ba: invokevirtual java/lang/String.length ()I
      // 3bd: if_icmpge 3e5
      // 3c0: bipush -1
      // 3c1: aload 4
      // 3c3: aload 3
      // 3c4: iload 5
      // 3c6: invokevirtual java/lang/String.charAt (I)C
      // 3c9: invokevirtual java/lang/String.indexOf (I)I
      // 3cc: if_icmpeq 3d7
      // 3cf: goto 3d3
      // 3d2: athrow
      // 3d3: goto 3df
      // 3d6: athrow
      // 3d7: new java/lang/Exception
      // 3da: dup
      // 3db: invokespecial java/lang/Exception.<init> ()V
      // 3de: athrow
      // 3df: iinc 5 1
      // 3e2: goto 3b7
      // 3e5: invokestatic java/lang/Runtime.getRuntime ()Ljava/lang/Runtime;
      // 3e8: new java/lang/StringBuilder
      // 3eb: dup
      // 3ec: invokespecial java/lang/StringBuilder.<init> ()V
      // 3ef: getstatic c.z [Ljava/lang/String;
      // 3f2: bipush 21
      // 3f4: aaload
      // 3f5: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 3f8: aload 3
      // 3f9: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 3fc: ldc "\""
      // 3fe: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 401: invokevirtual java/lang/StringBuilder.toString ()Ljava/lang/String;
      // 404: invokevirtual java/lang/Runtime.exec (Ljava/lang/String;)Ljava/lang/Process;
      // 407: pop
      // 408: aload 1
      // 409: aconst_null
      // 40a: putfield g.d Ljava/lang/Object;
      // 40d: goto 6f0
      // 410: astore 3
      // 411: aload 1
      // 412: aload 3
      // 413: putfield g.d Ljava/lang/Object;
      // 416: aload 3
      // 417: athrow
      // 418: aload 1
      // 419: getfield g.e I
      // 41c: bipush -1
      // 41d: ixor
      // 41e: bipush -1
      // 41f: if_icmpeq 427
      // 422: bipush 1
      // 423: goto 428
      // 426: athrow
      // 427: bipush 0
      // 428: istore 3
      // 429: aload 1
      // 42a: getfield g.f Ljava/lang/Object;
      // 42d: checkcast java/awt/Component
      // 430: astore 4
      // 432: aload 0
      // 433: getfield c.c Z
      // 436: ifeq 449
      // 439: aload 0
      // 43a: getfield c.t Lq;
      // 43d: bipush -4
      // 43f: aload 4
      // 441: iload 3
      // 442: invokevirtual q.a (ILjava/awt/Component;Z)V
      // 445: goto 482
      // 448: athrow
      // 449: ldc "j"
      // 44b: invokestatic java/lang/Class.forName (Ljava/lang/String;)Ljava/lang/Class;
      // 44e: getstatic c.z [Ljava/lang/String;
      // 451: bipush 24
      // 453: aaload
      // 454: bipush 2
      // 455: anewarray 257
      // 458: dup
      // 459: bipush 0
      // 45a: ldc java/awt/Component
      // 45c: aastore
      // 45d: dup
      // 45e: bipush 1
      // 45f: getstatic java/lang/Boolean.TYPE Ljava/lang/Class;
      // 462: aastore
      // 463: invokevirtual java/lang/Class.getDeclaredMethod (Ljava/lang/String;[Ljava/lang/Class;)Ljava/lang/reflect/Method;
      // 466: aload 0
      // 467: getfield c.x Ljava/lang/Object;
      // 46a: bipush 2
      // 46b: anewarray 261
      // 46e: dup
      // 46f: bipush 0
      // 470: aload 4
      // 472: aastore
      // 473: dup
      // 474: bipush 1
      // 475: new java/lang/Boolean
      // 478: dup
      // 479: iload 3
      // 47a: invokespecial java/lang/Boolean.<init> (Z)V
      // 47d: aastore
      // 47e: invokevirtual java/lang/reflect/Method.invoke (Ljava/lang/Object;[Ljava/lang/Object;)Ljava/lang/Object;
      // 481: pop
      // 482: goto 6f0
      // 485: aload 1
      // 486: getfield g.e I
      // 489: istore 3
      // 48a: aload 1
      // 48b: getfield g.c I
      // 48e: istore 4
      // 490: aload 0
      // 491: getfield c.c Z
      // 494: ifeq 4a8
      // 497: aload 0
      // 498: getfield c.t Lq;
      // 49b: sipush 23529
      // 49e: iload 4
      // 4a0: iload 3
      // 4a1: invokevirtual q.a (III)V
      // 4a4: goto 4e9
      // 4a7: athrow
      // 4a8: ldc "j"
      // 4aa: invokestatic java/lang/Class.forName (Ljava/lang/String;)Ljava/lang/Class;
      // 4ad: getstatic c.z [Ljava/lang/String;
      // 4b0: bipush 19
      // 4b2: aaload
      // 4b3: bipush 2
      // 4b4: anewarray 257
      // 4b7: dup
      // 4b8: bipush 0
      // 4b9: getstatic java/lang/Integer.TYPE Ljava/lang/Class;
      // 4bc: aastore
      // 4bd: dup
      // 4be: bipush 1
      // 4bf: getstatic java/lang/Integer.TYPE Ljava/lang/Class;
      // 4c2: aastore
      // 4c3: invokevirtual java/lang/Class.getDeclaredMethod (Ljava/lang/String;[Ljava/lang/Class;)Ljava/lang/reflect/Method;
      // 4c6: aload 0
      // 4c7: getfield c.x Ljava/lang/Object;
      // 4ca: bipush 2
      // 4cb: anewarray 261
      // 4ce: dup
      // 4cf: bipush 0
      // 4d0: new java/lang/Integer
      // 4d3: dup
      // 4d4: iload 3
      // 4d5: invokespecial java/lang/Integer.<init> (I)V
      // 4d8: aastore
      // 4d9: dup
      // 4da: bipush 1
      // 4db: new java/lang/Integer
      // 4de: dup
      // 4df: iload 4
      // 4e1: invokespecial java/lang/Integer.<init> (I)V
      // 4e4: aastore
      // 4e5: invokevirtual java/lang/reflect/Method.invoke (Ljava/lang/Object;[Ljava/lang/Object;)Ljava/lang/Object;
      // 4e8: pop
      // 4e9: goto 6f0
      // 4ec: getstatic c.b I
      // 4ef: getstatic c.p Ljava/lang/String;
      // 4f2: bipush 0
      // 4f3: aload 1
      // 4f4: getfield g.f Ljava/lang/Object;
      // 4f7: checkcast java/lang/String
      // 4fa: invokestatic c.a (ILjava/lang/String;ZLjava/lang/String;)Ld;
      // 4fd: astore 3
      // 4fe: aload 1
      // 4ff: aload 3
      // 500: putfield g.d Ljava/lang/Object;
      // 503: goto 6f0
      // 506: new java/awt/Frame
      // 509: dup
      // 50a: getstatic c.z [Ljava/lang/String;
      // 50d: bipush 12
      // 50f: aaload
      // 510: invokespecial java/awt/Frame.<init> (Ljava/lang/String;)V
      // 513: astore 3
      // 514: aload 1
      // 515: aload 3
      // 516: putfield g.d Ljava/lang/Object;
      // 519: aload 3
      // 51a: bipush 0
      // 51b: invokevirtual java/awt/Frame.setResizable (Z)V
      // 51e: aload 0
      // 51f: getfield c.c Z
      // 522: ifeq 54f
      // 525: aload 0
      // 526: getfield c.w Lwa;
      // 529: aload 3
      // 52a: aload 1
      // 52b: getfield g.c I
      // 52e: ldc -747878896
      // 530: ishr
      // 531: aload 1
      // 532: getfield g.c I
      // 535: ldc 65535
      // 537: iand
      // 538: aload 1
      // 539: getfield g.e I
      // 53c: ldc 65535
      // 53e: iand
      // 53f: aload 1
      // 540: getfield g.e I
      // 543: ldc 831913136
      // 545: iushr
      // 546: bipush 77
      // 548: invokevirtual wa.a (Ljava/awt/Frame;IIIIB)V
      // 54b: goto 5d6
      // 54e: athrow
      // 54f: getstatic c.z [Ljava/lang/String;
      // 552: bipush 15
      // 554: aaload
      // 555: invokestatic java/lang/Class.forName (Ljava/lang/String;)Ljava/lang/Class;
      // 558: getstatic c.z [Ljava/lang/String;
      // 55b: bipush 17
      // 55d: aaload
      // 55e: bipush 5
      // 55f: anewarray 257
      // 562: dup
      // 563: bipush 0
      // 564: ldc java/awt/Frame
      // 566: aastore
      // 567: dup
      // 568: bipush 1
      // 569: getstatic java/lang/Integer.TYPE Ljava/lang/Class;
      // 56c: aastore
      // 56d: dup
      // 56e: bipush 2
      // 56f: getstatic java/lang/Integer.TYPE Ljava/lang/Class;
      // 572: aastore
      // 573: dup
      // 574: bipush 3
      // 575: getstatic java/lang/Integer.TYPE Ljava/lang/Class;
      // 578: aastore
      // 579: dup
      // 57a: bipush 4
      // 57b: getstatic java/lang/Integer.TYPE Ljava/lang/Class;
      // 57e: aastore
      // 57f: invokevirtual java/lang/Class.getMethod (Ljava/lang/String;[Ljava/lang/Class;)Ljava/lang/reflect/Method;
      // 582: aload 0
      // 583: getfield c.i Ljava/lang/Object;
      // 586: bipush 5
      // 587: anewarray 261
      // 58a: dup
      // 58b: bipush 0
      // 58c: aload 3
      // 58d: aastore
      // 58e: dup
      // 58f: bipush 1
      // 590: new java/lang/Integer
      // 593: dup
      // 594: aload 1
      // 595: getfield g.e I
      // 598: ldc -1397573296
      // 59a: iushr
      // 59b: invokespecial java/lang/Integer.<init> (I)V
      // 59e: aastore
      // 59f: dup
      // 5a0: bipush 2
      // 5a1: new java/lang/Integer
      // 5a4: dup
      // 5a5: ldc 65535
      // 5a7: aload 1
      // 5a8: getfield g.e I
      // 5ab: iand
      // 5ac: invokespecial java/lang/Integer.<init> (I)V
      // 5af: aastore
      // 5b0: dup
      // 5b1: bipush 3
      // 5b2: new java/lang/Integer
      // 5b5: dup
      // 5b6: aload 1
      // 5b7: getfield g.c I
      // 5ba: ldc -1159913680
      // 5bc: ishr
      // 5bd: invokespecial java/lang/Integer.<init> (I)V
      // 5c0: aastore
      // 5c1: dup
      // 5c2: bipush 4
      // 5c3: new java/lang/Integer
      // 5c6: dup
      // 5c7: aload 1
      // 5c8: getfield g.c I
      // 5cb: ldc 65535
      // 5cd: iand
      // 5ce: invokespecial java/lang/Integer.<init> (I)V
      // 5d1: aastore
      // 5d2: invokevirtual java/lang/reflect/Method.invoke (Ljava/lang/Object;[Ljava/lang/Object;)Ljava/lang/Object;
      // 5d5: pop
      // 5d6: goto 6f0
      // 5d9: aload 0
      // 5da: getfield c.c Z
      // 5dd: ifeq 5f1
      // 5e0: aload 1
      // 5e1: aload 0
      // 5e2: getfield c.w Lwa;
      // 5e5: bipush -100
      // 5e7: invokevirtual wa.a (B)[I
      // 5ea: putfield g.d Ljava/lang/Object;
      // 5ed: goto 6f0
      // 5f0: athrow
      // 5f1: aload 1
      // 5f2: getstatic c.z [Ljava/lang/String;
      // 5f5: bipush 15
      // 5f7: aaload
      // 5f8: invokestatic java/lang/Class.forName (Ljava/lang/String;)Ljava/lang/Class;
      // 5fb: getstatic c.z [Ljava/lang/String;
      // 5fe: bipush 22
      // 600: aaload
      // 601: bipush 0
      // 602: anewarray 257
      // 605: invokevirtual java/lang/Class.getMethod (Ljava/lang/String;[Ljava/lang/Class;)Ljava/lang/reflect/Method;
      // 608: aload 0
      // 609: getfield c.i Ljava/lang/Object;
      // 60c: bipush 0
      // 60d: anewarray 261
      // 610: invokevirtual java/lang/reflect/Method.invoke (Ljava/lang/Object;[Ljava/lang/Object;)Ljava/lang/Object;
      // 613: putfield g.d Ljava/lang/Object;
      // 616: goto 6f0
      // 619: bipush 0
      // 61a: invokestatic p.a (I)J
      // 61d: getstatic c.d J
      // 620: lcmp
      // 621: ifge 62d
      // 624: new java/io/IOException
      // 627: dup
      // 628: invokespecial java/io/IOException.<init> ()V
      // 62b: athrow
      // 62c: athrow
      // 62d: aload 1
      // 62e: aload 1
      // 62f: getfield g.f Ljava/lang/Object;
      // 632: checkcast java/lang/String
      // 635: invokestatic java/net/InetAddress.getByName (Ljava/lang/String;)Ljava/net/InetAddress;
      // 638: invokevirtual java/net/InetAddress.getAddress ()[B
      // 63b: putfield g.d Ljava/lang/Object;
      // 63e: goto 6f0
      // 641: new java/lang/Exception
      // 644: dup
      // 645: ldc ""
      // 647: invokespecial java/lang/Exception.<init> (Ljava/lang/String;)V
      // 64a: athrow
      // 64b: aload 1
      // 64c: getfield g.f Ljava/lang/Object;
      // 64f: checkcast java/awt/datatransfer/Transferable
      // 652: astore 3
      // 653: invokestatic java/awt/Toolkit.getDefaultToolkit ()Ljava/awt/Toolkit;
      // 656: invokevirtual java/awt/Toolkit.getSystemClipboard ()Ljava/awt/datatransfer/Clipboard;
      // 659: astore 4
      // 65b: aload 4
      // 65d: aload 3
      // 65e: aconst_null
      // 65f: invokevirtual java/awt/datatransfer/Clipboard.setContents (Ljava/awt/datatransfer/Transferable;Ljava/awt/datatransfer/ClipboardOwner;)V
      // 662: goto 6f0
      // 665: invokestatic java/awt/Toolkit.getDefaultToolkit ()Ljava/awt/Toolkit;
      // 668: invokevirtual java/awt/Toolkit.getSystemClipboard ()Ljava/awt/datatransfer/Clipboard;
      // 66b: astore 3
      // 66c: aload 1
      // 66d: aload 3
      // 66e: aconst_null
      // 66f: invokevirtual java/awt/datatransfer/Clipboard.getContents (Ljava/lang/Object;)Ljava/awt/datatransfer/Transferable;
      // 672: putfield g.d Ljava/lang/Object;
      // 675: goto 6f0
      // 678: bipush 0
      // 679: invokestatic p.a (I)J
      // 67c: ldc2_w -1
      // 67f: lxor
      // 680: getstatic c.d J
      // 683: ldc2_w -1
      // 686: lxor
      // 687: lcmp
      // 688: ifle 694
      // 68b: new java/io/IOException
      // 68e: dup
      // 68f: invokespecial java/io/IOException.<init> ()V
      // 692: athrow
      // 693: athrow
      // 694: aload 1
      // 695: new java/io/DataInputStream
      // 698: dup
      // 699: aload 1
      // 69a: getfield g.f Ljava/lang/Object;
      // 69d: checkcast java/net/URL
      // 6a0: invokevirtual java/net/URL.openStream ()Ljava/io/InputStream;
      // 6a3: invokespecial java/io/DataInputStream.<init> (Ljava/io/InputStream;)V
      // 6a6: putfield g.d Ljava/lang/Object;
      // 6a9: goto 6f0
      // 6ac: getstatic c.d J
      // 6af: ldc2_w -1
      // 6b2: lxor
      // 6b3: bipush 0
      // 6b4: invokestatic p.a (I)J
      // 6b7: ldc2_w -1
      // 6ba: lxor
      // 6bb: lcmp
      // 6bc: ifge 6c8
      // 6bf: new java/io/IOException
      // 6c2: dup
      // 6c3: invokespecial java/io/IOException.<init> ()V
      // 6c6: athrow
      // 6c7: athrow
      // 6c8: aload 1
      // 6c9: sipush 4718
      // 6cc: aload 1
      // 6cd: getfield g.e I
      // 6d0: aload 1
      // 6d1: getfield g.f Ljava/lang/Object;
      // 6d4: checkcast java/lang/String
      // 6d7: invokestatic na.a (IILjava/lang/String;)Lm;
      // 6da: bipush 50
      // 6dc: invokevirtual m.a (B)Ljava/net/Socket;
      // 6df: putfield g.d Ljava/lang/Object;
      // 6e2: goto 6f0
      // 6e5: astore 3
      // 6e6: aload 1
      // 6e7: aload 3
      // 6e8: invokevirtual fa.getMessage ()Ljava/lang/String;
      // 6eb: putfield g.d Ljava/lang/Object;
      // 6ee: aload 3
      // 6ef: athrow
      // 6f0: aload 1
      // 6f1: bipush 1
      // 6f2: putfield g.b I
      // 6f5: goto 701
      // 6f8: astore 2
      // 6f9: aload 2
      // 6fa: athrow
      // 6fb: astore 2
      // 6fc: aload 1
      // 6fd: bipush 2
      // 6fe: putfield g.b I
      // 701: aload 1
      // 702: dup
      // 703: astore 2
      // 704: monitorenter
      // 705: aload 1
      // 706: invokevirtual java/lang/Object.notify ()V
      // 709: aload 2
      // 70a: monitorexit
      // 70b: goto 715
      // 70e: astore 6
      // 710: aload 2
      // 711: monitorexit
      // 712: aload 6
      // 714: athrow
      // 715: goto 000
      // 718: astore 1
      // 719: aload 1
      // 71a: athrow
   }

   private final g a(int var1, byte var2, String var3, boolean var4) {
      try {
         try {
            if (var2 != 81) {
               this.a(3, (byte)-100, (String)null, true);
            }
         } catch (RuntimeException var6) {
            throw var6;
         }

         c var10000;
         try {
            var10000 = this;
            if (var4) {
               return this.a(22, 0, (byte)-21, var1, var3);
            }
         } catch (RuntimeException var7) {
            throw var7;
         }

         return var10000.a(1, 0, (byte)-21, var1, var3);
      } catch (RuntimeException var8) {
         throw var8;
      }
   }

   final g a(byte var1, URL var2) {
      try {
         try {
            if (var1 != 74) {
               d = -110L;
            }
         } catch (RuntimeException var4) {
            throw var4;
         }

         return this.a(4, 0, (byte)-21, 0, var2);
      } catch (RuntimeException var5) {
         throw var5;
      }
   }

   final g a(int var1, byte var2) {
      try {
         int var3 = 9 / ((-58 - var2) / 56);
         return this.a(3, 0, (byte)-21, var1, null);
      } catch (RuntimeException var4) {
         throw var4;
      }
   }

   private final g a(int var1, int var2, byte var3, int var4, Object var5) {
      try {
         g var6 = new g();
         var6.e = var4;
         var6.g = var1;
         var6.c = var2;
         var6.f = var5;
         synchronized (this) {
            if (null == this.o) {
               this.o = this.a = var6;
            } else {
               this.o.a = var6;
               this.o = var6;
            }

            this.notify();
         }

         try {
            if (var3 != -21) {
               k = (String)null;
            }
         } catch (RuntimeException var10) {
            throw var10;
         }

         return var6;
      } catch (RuntimeException var11) {
         throw var11;
      }
   }

   c(int var1, String var2, int var3, boolean var4) throws Exception {
      this.e = false;
      this.c = false;
      this.v = null;
      this.o = null;

      try {
         k = z[42];
         b = var1;
         q = z[39];
         p = var2;
         this.j = var4;

         try {
            q = System.getProperty(z[32]);
            k = System.getProperty(z[36]);
         } catch (Exception var14) {
         }

         label144: {
            try {
               if (~q.toLowerCase().indexOf(z[43]) == 0) {
                  break label144;
               }
            } catch (Exception var22) {
               throw var22;
            }

            this.c = true;
         }

         try {
            h = System.getProperty(z[29]);
         } catch (Exception var13) {
            h = z[39];
         }

         g = h.toLowerCase();

         try {
            System.getProperty(z[34]).toLowerCase();
         } catch (Exception var12) {
         }

         try {
            System.getProperty(z[30]).toLowerCase();
         } catch (Exception var11) {
         }

         try {
            m = System.getProperty(z[31]);
            if (m != null) {
               m = m + "/";
            }
         } catch (Exception var10) {
         }

         label138: {
            try {
               if (null != m) {
                  break label138;
               }
            } catch (Exception var21) {
               throw var21;
            }

            m = z[37];
         }

         try {
            this.n = Toolkit.getDefaultToolkit().getSystemEventQueue();
         } catch (Throwable var9) {
         }

         label150: {
            try {
               if (this.c) {
                  break label150;
               }
            } catch (Exception var20) {
               throw var20;
            }

            try {
               u = Class.forName(z[40]).getDeclaredMethod(z[25], boolean.class);
            } catch (Exception var8) {
            }

            try {
               y = Class.forName(z[28]).getDeclaredMethod(z[38], boolean.class);
            } catch (Exception var7) {
            }
         }

         r.a(b, (byte)101, p);
         if (this.j) {
            this.s = new d(r.a(b, null, z[35], 0), z[8], 25L);
            this.f = new d(r.a(2, z[33]), z[8], 314572800L);
            this.v = new d(r.a(2, z[26]), z[8], 1048576L);
            this.l = new d[var3];
            int var5 = 0;

            try {
               while (var3 > var5) {
                  this.l[var5] = new d(r.a(2, z[41] + var5), z[8], 1048576L);
                  var5++;
               }
            } catch (Exception var19) {
               throw var19;
            }

            if (this.c) {
               try {
                  Class.forName(z[27]).newInstance();
               } catch (Throwable var6) {
               }
            }

            try {
               label116: {
                  try {
                     if (this.c) {
                        this.w = new wa();
                        break label116;
                     }
                  } catch (Exception var17) {
                     throw var17;
                  }

                  this.i = Class.forName(z[15]).newInstance();
               }
            } catch (Throwable var18) {
            }

            try {
               label109: {
                  try {
                     if (!this.c) {
                        this.x = Class.forName("j").newInstance();
                        break label109;
                     }
                  } catch (Exception var15) {
                     throw var15;
                  }

                  this.t = new q();
               }
            } catch (Throwable var16) {
            }
         }

         this.e = false;
         this.r = new Thread(this);
         this.r.setPriority(10);
         this.r.setDaemon(true);
         this.r.start();
      } catch (RuntimeException var23) {
         throw var23;
      }
   }

   private static char[] z(String var0) {
      char[] var10000 = var0.toCharArray();
      if (var10000.length < 2) {
         var10000[0] = (char)(var10000[0] ^ 'B');
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
               var10005 = 6;
               break;
            case 1:
               var10005 = 41;
               break;
            case 2:
               var10005 = 18;
               break;
            case 3:
               var10005 = 97;
               break;
            default:
               var10005 = 66;
         }

         var10001[var1] = (char)(var10004 ^ var10005);
      }

      return new String(var10001).intern();
   }
}
