final class aa {
   static int e;
   static int h;
   static int a;
   static int[] c;
   static int b;
   static int d = 114;
   private int[] j;
   private byte[] i;
   static int l;
   private int[] g;
   static int f;
   static String[] k = new String[100];
   private static final String[] z = new String[]{
      z(z("7L\u001d]e8DG_$")), z(z("-\u0003\u001dOq")), z(z("8X_\r")), z(z("7L\u001d#$")), z(z("7L\u001d $")), z(z("7L\u001d\"$"))
   };

   final int a(int var1, byte[] var2, int var3, byte[] var4, int var5, int var6) {
      boolean var15 = client.vh;

      try {
         try {
            h++;
            if (var6 < 99) {
               this.a(-58, (byte[])null, 69, (byte[])null, -39, 22);
            }
         } catch (RuntimeException var22) {
            throw var22;
         }

         int var24 = 0;
         var1 += var5;
         int var8 = var3 << 1494901059;

         int var35;
         int var40;
         while (true) {
            if (var5 < var1) {
               int var9 = var4[var5] & 255;
               int var10 = this.g[var9];
               byte var11 = this.i[var9];

               try {
                  var35 = ~var11;
                  var40 = -1;
                  if (var15) {
                     break;
                  }

                  if (var35 == -1) {
                     throw new RuntimeException("" + var9);
                  }
               } catch (RuntimeException var18) {
                  throw var18;
               }

               int var12 = var8 >> 1897076227;
               int var13 = var8 & 7;
               int var25 = var24 & -var13 >> -2036172737;
               int var14 = var12 + (var11 + var13 + -1 >> 1085645667);
               var8 += var11;
               var13 += 24;
               byte[] var36 = var2;
               var40 = var12;
               byte var45 = (byte)(var24 = d.a(var25, var10 >>> var13));

               label122: {
                  try {
                     var36[var40] = var45;
                     if (~var14 >= ~var12 && !var15) {
                        break label122;
                     }
                  } catch (RuntimeException var21) {
                     throw var21;
                  }

                  var12++;
                  var13 -= 8;
                  byte[] var37 = var2;
                  var40 = var12;
                  var45 = (byte)(var24 = var10 >>> var13);

                  try {
                     var37[var40] = var45;
                     if (~var12 <= ~var14 && !var15) {
                        break label122;
                     }
                  } catch (RuntimeException var20) {
                     throw var20;
                  }

                  var12++;
                  var13 -= 8;
                  byte[] var38 = var2;
                  var40 = var12;
                  var45 = (byte)(var24 = var10 >>> var13);

                  try {
                     var38[var40] = var45;
                     if (var12 >= var14 && !var15) {
                        break label122;
                     }
                  } catch (RuntimeException var19) {
                     throw var19;
                  }

                  var12++;
                  var13 -= 8;
                  var2[var12] = (byte)(var24 = var10 >>> var13);
                  if (var12 < var14) {
                     var12++;
                     var13 -= 8;
                     var2[var12] = (byte)(var24 = var10 << -var13);
                  }
               }

               var5++;
               if (!var15) {
                  continue;
               }
            }

            var35 = -var3;
            var40 = var8 + 7 >> -261766397;
            break;
         }

         return var35 + var40;
      } catch (RuntimeException var23) {
         RuntimeException var7 = var23;

         RuntimeException var10000;
         StringBuilder var10001;
         String var10002;
         label71: {
            try {
               var10000 = var7;
               var10001 = new StringBuilder().append(z[3]).append(var1).append(',');
               if (var2 != null) {
                  var10002 = z[1];
                  break label71;
               }
            } catch (RuntimeException var17) {
               throw var17;
            }

            var10002 = z[2];
         }

         try {
            var10001 = var10001.append(var10002).append(',').append(var3).append(',');
            if (var4 != null) {
               throw i.a(var10000, var10001.append(z[1]).append((char)44).append(var5).append((char)44).append(var6).append((char)41).toString());
            }
         } catch (RuntimeException var16) {
            throw var16;
         }

         throw i.a(var10000, var10001.append(z[2]).append((char)44).append(var5).append((char)44).append(var6).append((char)41).toString());
      }
   }

   final int a(byte[] param1, byte[] param2, int param3, int param4, int param5, int param6) {
      // $VF: Couldn't be decompiled
      // Please report this to the Vineflower issue tracker, at https://github.com/Vineflower/vineflower/issues with a copy of the class file (if you have the rights to distribute it!)
      // java.lang.RuntimeException: parsing failure!
      //   at org.jetbrains.java.decompiler.modules.decompiler.decompose.DomHelper.parseGraph(DomHelper.java:211)
      //   at org.jetbrains.java.decompiler.main.rels.MethodProcessor.codeToJava(MethodProcessor.java:166)
      //
      // Bytecode:
      // 000: getstatic client.vh Z
      // 003: istore 11
      // 005: getstatic aa.a I
      // 008: bipush 1
      // 009: iadd
      // 00a: putstatic aa.a I
      // 00d: bipush 0
      // 00e: iload 6
      // 010: if_icmpne 015
      // 013: bipush 0
      // 014: ireturn
      // 015: bipush 0
      // 016: istore 7
      // 018: iload 6
      // 01a: iload 3
      // 01b: iadd
      // 01c: istore 6
      // 01e: iload 5
      // 020: bipush -1
      // 021: if_icmpeq 03d
      // 024: aload 0
      // 025: bipush 105
      // 027: aconst_null
      // 028: checkcast [B
      // 02b: bipush 82
      // 02d: aconst_null
      // 02e: checkcast [B
      // 031: bipush 125
      // 033: bipush -45
      // 035: invokevirtual aa.a (I[BI[BII)I
      // 038: pop
      // 039: goto 03d
      // 03c: athrow
      // 03d: iload 4
      // 03f: istore 8
      // 041: aload 1
      // 042: iload 8
      // 044: baload
      // 045: istore 9
      // 047: iload 9
      // 049: iflt 05d
      // 04c: iinc 7 1
      // 04f: iload 11
      // 051: ifne 066
      // 054: iload 11
      // 056: ifeq 066
      // 059: goto 05d
      // 05c: athrow
      // 05d: aload 0
      // 05e: getfield aa.j [I
      // 061: iload 7
      // 063: iaload
      // 064: istore 7
      // 066: bipush 0
      // 067: aload 0
      // 068: getfield aa.j [I
      // 06b: iload 7
      // 06d: iaload
      // 06e: dup
      // 06f: istore 10
      // 071: if_icmple 08c
      // 074: aload 2
      // 075: iload 3
      // 076: iinc 3 1
      // 079: iload 10
      // 07b: bipush -1
      // 07c: ixor
      // 07d: i2b
      // 07e: bastore
      // 07f: iload 3
      // 080: iload 6
      // 082: if_icmpge 2b6
      // 085: goto 089
      // 088: athrow
      // 089: bipush 0
      // 08a: istore 7
      // 08c: bipush 64
      // 08e: iload 9
      // 090: iand
      // 091: ifeq 0a2
      // 094: aload 0
      // 095: getfield aa.j [I
      // 098: iload 7
      // 09a: iaload
      // 09b: istore 7
      // 09d: iload 11
      // 09f: ifeq 0a9
      // 0a2: iinc 7 1
      // 0a5: goto 0a9
      // 0a8: athrow
      // 0a9: aload 0
      // 0aa: getfield aa.j [I
      // 0ad: iload 7
      // 0af: iaload
      // 0b0: dup
      // 0b1: istore 10
      // 0b3: iflt 0b9
      // 0b6: goto 0d6
      // 0b9: aload 2
      // 0ba: iload 3
      // 0bb: iinc 3 1
      // 0be: iload 10
      // 0c0: bipush -1
      // 0c1: ixor
      // 0c2: i2b
      // 0c3: bastore
      // 0c4: iload 6
      // 0c6: iload 3
      // 0c7: if_icmpgt 0d3
      // 0ca: iload 11
      // 0cc: ifeq 2b6
      // 0cf: goto 0d3
      // 0d2: athrow
      // 0d3: bipush 0
      // 0d4: istore 7
      // 0d6: iload 9
      // 0d8: bipush 32
      // 0da: iand
      // 0db: bipush -1
      // 0dc: ixor
      // 0dd: bipush -1
      // 0de: if_icmpne 0ed
      // 0e1: iinc 7 1
      // 0e4: iload 11
      // 0e6: ifeq 0f6
      // 0e9: goto 0ed
      // 0ec: athrow
      // 0ed: aload 0
      // 0ee: getfield aa.j [I
      // 0f1: iload 7
      // 0f3: iaload
      // 0f4: istore 7
      // 0f6: aload 0
      // 0f7: getfield aa.j [I
      // 0fa: iload 7
      // 0fc: iaload
      // 0fd: dup
      // 0fe: istore 10
      // 100: bipush -1
      // 101: ixor
      // 102: bipush -1
      // 103: if_icmple 12b
      // 106: aload 2
      // 107: iload 3
      // 108: iinc 3 1
      // 10b: iload 10
      // 10d: bipush -1
      // 10e: ixor
      // 10f: i2b
      // 110: bastore
      // 111: iload 3
      // 112: bipush -1
      // 113: ixor
      // 114: iload 6
      // 116: bipush -1
      // 117: ixor
      // 118: if_icmpgt 128
      // 11b: goto 11f
      // 11e: athrow
      // 11f: iload 11
      // 121: ifeq 2b6
      // 124: goto 128
      // 127: athrow
      // 128: bipush 0
      // 129: istore 7
      // 12b: bipush 0
      // 12c: bipush 16
      // 12e: iload 9
      // 130: iand
      // 131: if_icmpeq 142
      // 134: aload 0
      // 135: getfield aa.j [I
      // 138: iload 7
      // 13a: iaload
      // 13b: istore 7
      // 13d: iload 11
      // 13f: ifeq 149
      // 142: iinc 7 1
      // 145: goto 149
      // 148: athrow
      // 149: aload 0
      // 14a: getfield aa.j [I
      // 14d: iload 7
      // 14f: iaload
      // 150: dup
      // 151: istore 10
      // 153: bipush -1
      // 154: ixor
      // 155: bipush -1
      // 156: if_icmple 17a
      // 159: aload 2
      // 15a: iload 3
      // 15b: iinc 3 1
      // 15e: iload 10
      // 160: bipush -1
      // 161: ixor
      // 162: i2b
      // 163: bastore
      // 164: iload 3
      // 165: iload 6
      // 167: if_icmplt 177
      // 16a: goto 16e
      // 16d: athrow
      // 16e: iload 11
      // 170: ifeq 2b6
      // 173: goto 177
      // 176: athrow
      // 177: bipush 0
      // 178: istore 7
      // 17a: bipush -1
      // 17b: iload 9
      // 17d: bipush 8
      // 17f: iand
      // 180: bipush -1
      // 181: ixor
      // 182: if_icmpeq 193
      // 185: aload 0
      // 186: getfield aa.j [I
      // 189: iload 7
      // 18b: iaload
      // 18c: istore 7
      // 18e: iload 11
      // 190: ifeq 19a
      // 193: iinc 7 1
      // 196: goto 19a
      // 199: athrow
      // 19a: bipush -1
      // 19b: aload 0
      // 19c: getfield aa.j [I
      // 19f: iload 7
      // 1a1: iaload
      // 1a2: dup
      // 1a3: istore 10
      // 1a5: bipush -1
      // 1a6: ixor
      // 1a7: if_icmplt 1ae
      // 1aa: goto 1c2
      // 1ad: athrow
      // 1ae: aload 2
      // 1af: iload 3
      // 1b0: iinc 3 1
      // 1b3: iload 10
      // 1b5: bipush -1
      // 1b6: ixor
      // 1b7: i2b
      // 1b8: bastore
      // 1b9: iload 3
      // 1ba: iload 6
      // 1bc: if_icmpge 2b6
      // 1bf: bipush 0
      // 1c0: istore 7
      // 1c2: bipush 4
      // 1c3: iload 9
      // 1c5: iand
      // 1c6: ifeq 1d7
      // 1c9: aload 0
      // 1ca: getfield aa.j [I
      // 1cd: iload 7
      // 1cf: iaload
      // 1d0: istore 7
      // 1d2: iload 11
      // 1d4: ifeq 1de
      // 1d7: iinc 7 1
      // 1da: goto 1de
      // 1dd: athrow
      // 1de: bipush 0
      // 1df: aload 0
      // 1e0: getfield aa.j [I
      // 1e3: iload 7
      // 1e5: iaload
      // 1e6: dup
      // 1e7: istore 10
      // 1e9: if_icmple 211
      // 1ec: aload 2
      // 1ed: iload 3
      // 1ee: iinc 3 1
      // 1f1: iload 10
      // 1f3: bipush -1
      // 1f4: ixor
      // 1f5: i2b
      // 1f6: bastore
      // 1f7: iload 6
      // 1f9: bipush -1
      // 1fa: ixor
      // 1fb: iload 3
      // 1fc: bipush -1
      // 1fd: ixor
      // 1fe: if_icmplt 20e
      // 201: goto 205
      // 204: athrow
      // 205: iload 11
      // 207: ifeq 2b6
      // 20a: goto 20e
      // 20d: athrow
      // 20e: bipush 0
      // 20f: istore 7
      // 211: bipush 0
      // 212: bipush 2
      // 213: iload 9
      // 215: iand
      // 216: if_icmpne 225
      // 219: iinc 7 1
      // 21c: iload 11
      // 21e: ifeq 22e
      // 221: goto 225
      // 224: athrow
      // 225: aload 0
      // 226: getfield aa.j [I
      // 229: iload 7
      // 22b: iaload
      // 22c: istore 7
      // 22e: aload 0
      // 22f: getfield aa.j [I
      // 232: iload 7
      // 234: iaload
      // 235: dup
      // 236: istore 10
      // 238: bipush -1
      // 239: ixor
      // 23a: bipush -1
      // 23b: if_icmpgt 242
      // 23e: goto 263
      // 241: athrow
      // 242: aload 2
      // 243: iload 3
      // 244: iinc 3 1
      // 247: iload 10
      // 249: bipush -1
      // 24a: ixor
      // 24b: i2b
      // 24c: bastore
      // 24d: iload 6
      // 24f: bipush -1
      // 250: ixor
      // 251: iload 3
      // 252: bipush -1
      // 253: ixor
      // 254: if_icmplt 260
      // 257: iload 11
      // 259: ifeq 2b6
      // 25c: goto 260
      // 25f: athrow
      // 260: bipush 0
      // 261: istore 7
      // 263: bipush -1
      // 264: bipush 1
      // 265: iload 9
      // 267: iand
      // 268: bipush -1
      // 269: ixor
      // 26a: if_icmpeq 27b
      // 26d: aload 0
      // 26e: getfield aa.j [I
      // 271: iload 7
      // 273: iaload
      // 274: istore 7
      // 276: iload 11
      // 278: ifeq 282
      // 27b: iinc 7 1
      // 27e: goto 282
      // 281: athrow
      // 282: aload 0
      // 283: getfield aa.j [I
      // 286: iload 7
      // 288: iaload
      // 289: dup
      // 28a: istore 10
      // 28c: bipush -1
      // 28d: ixor
      // 28e: bipush -1
      // 28f: if_icmple 2ae
      // 292: aload 2
      // 293: iload 3
      // 294: iinc 3 1
      // 297: iload 10
      // 299: bipush -1
      // 29a: ixor
      // 29b: i2b
      // 29c: bastore
      // 29d: iload 6
      // 29f: bipush -1
      // 2a0: ixor
      // 2a1: iload 3
      // 2a2: bipush -1
      // 2a3: ixor
      // 2a4: if_icmpge 2b6
      // 2a7: goto 2ab
      // 2aa: athrow
      // 2ab: bipush 0
      // 2ac: istore 7
      // 2ae: iinc 8 1
      // 2b1: iload 11
      // 2b3: ifeq 041
      // 2b6: iload 4
      // 2b8: ineg
      // 2b9: bipush 1
      // 2ba: iadd
      // 2bb: iload 8
      // 2bd: iadd
      // 2be: ireturn
      // 2bf: astore 7
      // 2c1: aload 7
      // 2c3: new java/lang/StringBuilder
      // 2c6: dup
      // 2c7: invokespecial java/lang/StringBuilder.<init> ()V
      // 2ca: getstatic aa.z [Ljava/lang/String;
      // 2cd: bipush 4
      // 2ce: aaload
      // 2cf: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 2d2: aload 1
      // 2d3: ifnull 2df
      // 2d6: getstatic aa.z [Ljava/lang/String;
      // 2d9: bipush 1
      // 2da: aaload
      // 2db: goto 2e4
      // 2de: athrow
      // 2df: getstatic aa.z [Ljava/lang/String;
      // 2e2: bipush 2
      // 2e3: aaload
      // 2e4: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 2e7: bipush 44
      // 2e9: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 2ec: aload 2
      // 2ed: ifnull 2f9
      // 2f0: getstatic aa.z [Ljava/lang/String;
      // 2f3: bipush 1
      // 2f4: aaload
      // 2f5: goto 2fe
      // 2f8: athrow
      // 2f9: getstatic aa.z [Ljava/lang/String;
      // 2fc: bipush 2
      // 2fd: aaload
      // 2fe: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 301: bipush 44
      // 303: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 306: iload 3
      // 307: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 30a: bipush 44
      // 30c: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 30f: iload 4
      // 311: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 314: bipush 44
      // 316: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 319: iload 5
      // 31b: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 31e: bipush 44
      // 320: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 323: iload 6
      // 325: invokevirtual java/lang/StringBuilder.append (I)Ljava/lang/StringBuilder;
      // 328: bipush 41
      // 32a: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 32d: invokevirtual java/lang/StringBuilder.toString ()Ljava/lang/String;
      // 330: invokestatic i.a (Ljava/lang/Throwable;Ljava/lang/String;)Lla;
      // 333: athrow
   }

   static final int a(int var0, boolean var1) {
      try {
         e++;
         var0 = --var0 | var0 >>> 311248929;
         var0 |= var0 >>> -1130998654;
         var0 |= var0 >>> -669289052;

         try {
            if (var1) {
               b = -4;
            }
         } catch (RuntimeException var3) {
            throw var3;
         }

         var0 |= var0 >>> -948655896;
         var0 |= var0 >>> 795067056;
         return var0 - -1;
      } catch (RuntimeException var4) {
         throw i.a(var4, z[5] + var0 + 44 + var1 + 41);
      }
   }

   aa(byte[] param1) {
      // $VF: Couldn't be decompiled
      // Please report this to the Vineflower issue tracker, at https://github.com/Vineflower/vineflower/issues with a copy of the class file (if you have the rights to distribute it!)
      // java.lang.RuntimeException: parsing failure!
      //   at org.jetbrains.java.decompiler.modules.decompiler.decompose.DomHelper.parseGraph(DomHelper.java:211)
      //   at org.jetbrains.java.decompiler.main.rels.MethodProcessor.codeToJava(MethodProcessor.java:166)
      //
      // Bytecode:
      // 000: aload 0
      // 001: invokespecial java/lang/Object.<init> ()V
      // 004: aload 1
      // 005: arraylength
      // 006: istore 2
      // 007: aload 0
      // 008: iload 2
      // 009: newarray 10
      // 00b: putfield aa.g [I
      // 00e: aload 0
      // 00f: aload 1
      // 010: putfield aa.i [B
      // 013: aload 0
      // 014: bipush 8
      // 016: newarray 10
      // 018: putfield aa.j [I
      // 01b: bipush 33
      // 01d: newarray 10
      // 01f: astore 3
      // 020: bipush 0
      // 021: istore 4
      // 023: bipush 0
      // 024: istore 5
      // 026: iload 5
      // 028: bipush -1
      // 029: ixor
      // 02a: iload 2
      // 02b: bipush -1
      // 02c: ixor
      // 02d: if_icmple 1b0
      // 030: aload 1
      // 031: iload 5
      // 033: baload
      // 034: istore 6
      // 036: bipush 0
      // 037: iload 6
      // 039: if_icmpeq 1aa
      // 03c: bipush 1
      // 03d: iload 6
      // 03f: ineg
      // 040: bipush 32
      // 042: iadd
      // 043: ishl
      // 044: istore 7
      // 046: aload 3
      // 047: iload 6
      // 049: iaload
      // 04a: istore 8
      // 04c: aload 0
      // 04d: getfield aa.g [I
      // 050: iload 5
      // 052: iload 8
      // 054: iastore
      // 055: bipush -1
      // 056: iload 7
      // 058: iload 8
      // 05a: iand
      // 05b: bipush -1
      // 05c: ixor
      // 05d: if_icmpeq 06b
      // 060: aload 3
      // 061: iload 6
      // 063: bipush 1
      // 064: isub
      // 065: iaload
      // 066: istore 9
      // 068: goto 0c2
      // 06b: iload 6
      // 06d: bipush 1
      // 06e: isub
      // 06f: istore 10
      // 071: iload 10
      // 073: bipush -1
      // 074: ixor
      // 075: bipush -2
      // 077: if_icmpgt 0bb
      // 07a: aload 3
      // 07b: iload 10
      // 07d: iaload
      // 07e: istore 11
      // 080: iload 8
      // 082: iload 11
      // 084: if_icmpeq 08b
      // 087: goto 0bb
      // 08a: athrow
      // 08b: bipush 1
      // 08c: bipush 32
      // 08e: iload 10
      // 090: isub
      // 091: ishl
      // 092: istore 12
      // 094: iload 11
      // 096: iload 12
      // 098: iand
      // 099: ifeq 0aa
      // 09c: aload 3
      // 09d: iload 10
      // 09f: aload 3
      // 0a0: iload 10
      // 0a2: bipush 1
      // 0a3: isub
      // 0a4: iaload
      // 0a5: iastore
      // 0a6: goto 0bb
      // 0a9: athrow
      // 0aa: aload 3
      // 0ab: iload 10
      // 0ad: iload 12
      // 0af: iload 11
      // 0b1: invokestatic d.a (II)I
      // 0b4: iastore
      // 0b5: iinc 10 -1
      // 0b8: goto 071
      // 0bb: iload 8
      // 0bd: iload 7
      // 0bf: ior
      // 0c0: istore 9
      // 0c2: aload 3
      // 0c3: iload 6
      // 0c5: iload 9
      // 0c7: iastore
      // 0c8: iload 6
      // 0ca: bipush -1
      // 0cb: isub
      // 0cc: istore 10
      // 0ce: iload 10
      // 0d0: bipush 32
      // 0d2: if_icmpgt 0f6
      // 0d5: iload 8
      // 0d7: bipush -1
      // 0d8: ixor
      // 0d9: aload 3
      // 0da: iload 10
      // 0dc: iaload
      // 0dd: bipush -1
      // 0de: ixor
      // 0df: if_icmpeq 0ea
      // 0e2: goto 0e6
      // 0e5: athrow
      // 0e6: goto 0f0
      // 0e9: athrow
      // 0ea: aload 3
      // 0eb: iload 10
      // 0ed: iload 9
      // 0ef: iastore
      // 0f0: iinc 10 1
      // 0f3: goto 0ce
      // 0f6: bipush 0
      // 0f7: istore 10
      // 0f9: bipush 0
      // 0fa: istore 11
      // 0fc: iload 6
      // 0fe: iload 11
      // 100: if_icmple 18a
      // 103: ldc -2147483648
      // 105: iload 11
      // 107: iushr
      // 108: istore 12
      // 10a: bipush -1
      // 10b: iload 12
      // 10d: iload 8
      // 10f: iand
      // 110: bipush -1
      // 111: ixor
      // 112: if_icmpne 11c
      // 115: iinc 10 1
      // 118: goto 13f
      // 11b: athrow
      // 11c: aload 0
      // 11d: getfield aa.j [I
      // 120: iload 10
      // 122: iaload
      // 123: bipush -1
      // 124: ixor
      // 125: bipush -1
      // 126: if_icmpne 136
      // 129: aload 0
      // 12a: getfield aa.j [I
      // 12d: iload 10
      // 12f: iload 4
      // 131: iastore
      // 132: goto 136
      // 135: athrow
      // 136: aload 0
      // 137: getfield aa.j [I
      // 13a: iload 10
      // 13c: iaload
      // 13d: istore 10
      // 13f: aload 0
      // 140: getfield aa.j [I
      // 143: arraylength
      // 144: iload 10
      // 146: if_icmpgt 17e
      // 149: aload 0
      // 14a: getfield aa.j [I
      // 14d: arraylength
      // 14e: bipush 2
      // 14f: imul
      // 150: newarray 10
      // 152: astore 13
      // 154: bipush 0
      // 155: istore 14
      // 157: aload 0
      // 158: getfield aa.j [I
      // 15b: arraylength
      // 15c: bipush -1
      // 15d: ixor
      // 15e: iload 14
      // 160: bipush -1
      // 161: ixor
      // 162: if_icmpge 178
      // 165: aload 13
      // 167: iload 14
      // 169: aload 0
      // 16a: getfield aa.j [I
      // 16d: iload 14
      // 16f: iaload
      // 170: iastore
      // 171: iinc 14 1
      // 174: goto 157
      // 177: athrow
      // 178: aload 0
      // 179: aload 13
      // 17b: putfield aa.j [I
      // 17e: iload 12
      // 180: bipush 1
      // 181: iushr
      // 182: istore 12
      // 184: iinc 11 1
      // 187: goto 0fc
      // 18a: iload 4
      // 18c: bipush -1
      // 18d: ixor
      // 18e: iload 10
      // 190: bipush -1
      // 191: ixor
      // 192: if_icmpge 199
      // 195: goto 19f
      // 198: athrow
      // 199: iload 10
      // 19b: bipush 1
      // 19c: iadd
      // 19d: istore 4
      // 19f: aload 0
      // 1a0: getfield aa.j [I
      // 1a3: iload 10
      // 1a5: iload 5
      // 1a7: bipush -1
      // 1a8: ixor
      // 1a9: iastore
      // 1aa: iinc 5 1
      // 1ad: goto 026
      // 1b0: goto 1e5
      // 1b3: astore 2
      // 1b4: aload 2
      // 1b5: new java/lang/StringBuilder
      // 1b8: dup
      // 1b9: invokespecial java/lang/StringBuilder.<init> ()V
      // 1bc: getstatic aa.z [Ljava/lang/String;
      // 1bf: bipush 0
      // 1c0: aaload
      // 1c1: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 1c4: aload 1
      // 1c5: ifnull 1d1
      // 1c8: getstatic aa.z [Ljava/lang/String;
      // 1cb: bipush 1
      // 1cc: aaload
      // 1cd: goto 1d6
      // 1d0: athrow
      // 1d1: getstatic aa.z [Ljava/lang/String;
      // 1d4: bipush 2
      // 1d5: aaload
      // 1d6: invokevirtual java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
      // 1d9: bipush 41
      // 1db: invokevirtual java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
      // 1de: invokevirtual java/lang/StringBuilder.toString ()Ljava/lang/String;
      // 1e1: invokestatic i.a (Ljava/lang/Throwable;Ljava/lang/String;)Lla;
      // 1e4: athrow
      // 1e5: return
   }

   private static char[] z(String var0) {
      char[] var10000 = var0.toCharArray();
      if (var10000.length < 2) {
         var10000[0] = (char)(var10000[0] ^ '\f');
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
               var10005 = 86;
               break;
            case 1:
               var10005 = 45;
               break;
            case 2:
               var10005 = 51;
               break;
            case 3:
               var10005 = 97;
               break;
            default:
               var10005 = 12;
         }

         var10001[var1] = (char)(var10004 ^ var10005);
      }

      return new String(var10001).intern();
   }
}
