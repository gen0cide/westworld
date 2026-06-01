/*
 * Decompiled with CFR 0.152.
 */
final class i {
    static int b;
    static int e;
    static int c;
    int a;
    static int d;
    static int[] g;
    private static long[] h;
    static String f;
    private static final String[] z;

    public final String toString() {
        try {
            ++e;
            throw new IllegalStateException();
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, z[5]);
        }
    }

    /*
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    static final int a(int n2, int n3, int n4, CharSequence charSequence, byte by, byte[] byArray) {
        boolean bl = client.vh;
        try {
            int n5;
            ++b;
            int n6 = n2 + -n4;
            if (by >= -78) {
                i.a(null, null);
            }
            int n7 = 0;
            while (~n7 > ~n6) {
                block73: {
                    char c2;
                    block71: {
                        block74: {
                            block75: {
                                block77: {
                                    block80: {
                                        block82: {
                                            block83: {
                                                block84: {
                                                    block85: {
                                                        block86: {
                                                            block87: {
                                                                block91: {
                                                                    block92: {
                                                                        block93: {
                                                                            block96: {
                                                                                block98: {
                                                                                    block99: {
                                                                                        block97: {
                                                                                            block95: {
                                                                                                block94: {
                                                                                                    block90: {
                                                                                                        block89: {
                                                                                                            block88: {
                                                                                                                block81: {
                                                                                                                    block79: {
                                                                                                                        block78: {
                                                                                                                            block76: {
                                                                                                                                block72: {
                                                                                                                                    c2 = charSequence.charAt(n4 + n7);
                                                                                                                                    n5 = c2;
                                                                                                                                    if (bl) return n5;
                                                                                                                                    if (n5 > 0 && c2 < '\u0080' || ~c2 <= -161 && '\u00ff' >= c2) break block71;
                                                                                                                                    if (c2 != '\u20ac') break block72;
                                                                                                                                    byArray[n7 + n3] = -128;
                                                                                                                                    if (!bl) break block73;
                                                                                                                                }
                                                                                                                                if (c2 == '\u201a') break block74;
                                                                                                                                if (-403 == ~c2) break block75;
                                                                                                                                if (~c2 != -8223) break block76;
                                                                                                                                byArray[n3 - -n7] = -124;
                                                                                                                                if (!bl) break block73;
                                                                                                                            }
                                                                                                                            if ('\u2026' == c2) break block77;
                                                                                                                            if (c2 != '\u2020') break block78;
                                                                                                                            byArray[n3 - -n7] = -122;
                                                                                                                            if (!bl) break block73;
                                                                                                                        }
                                                                                                                        if (c2 != '\u2021') break block79;
                                                                                                                        byArray[n7 + n3] = -121;
                                                                                                                        if (!bl) break block73;
                                                                                                                    }
                                                                                                                    if (~c2 == -711) break block80;
                                                                                                                    if (~c2 != -8241) break block81;
                                                                                                                    byArray[n7 + n3] = -119;
                                                                                                                    if (!bl) break block73;
                                                                                                                }
                                                                                                                if ('\u0160' == c2) break block82;
                                                                                                                if ('\u2039' == c2) break block83;
                                                                                                                if (c2 == '\u0152') break block84;
                                                                                                                if (~c2 == -382) break block85;
                                                                                                                if (~c2 == -8217) break block86;
                                                                                                                if (~c2 == -8218) break block87;
                                                                                                                if ('\u201c' != c2) break block88;
                                                                                                                byArray[n3 - -n7] = -109;
                                                                                                                if (!bl) break block73;
                                                                                                            }
                                                                                                            if (-8222 != ~c2) break block89;
                                                                                                            byArray[n7 + n3] = -108;
                                                                                                            if (!bl) break block73;
                                                                                                        }
                                                                                                        if (c2 != '\u2022') break block90;
                                                                                                        byArray[n3 - -n7] = -107;
                                                                                                        if (!bl) break block73;
                                                                                                    }
                                                                                                    if (-8212 == ~c2) break block91;
                                                                                                    if (c2 == '\u2014') break block92;
                                                                                                    if ('\u02dc' == c2) break block93;
                                                                                                    if (~c2 != -8483) break block94;
                                                                                                    byArray[n7 + n3] = -103;
                                                                                                    if (!bl) break block73;
                                                                                                }
                                                                                                if (c2 != '\u0161') break block95;
                                                                                                byArray[n3 - -n7] = -102;
                                                                                                if (!bl) break block73;
                                                                                            }
                                                                                            if ('\u203a' == c2) break block96;
                                                                                            if (~c2 != -340) break block97;
                                                                                            byArray[n3 + n7] = -100;
                                                                                            if (!bl) break block73;
                                                                                        }
                                                                                        if ('\u017e' == c2) break block98;
                                                                                        if ('\u0178' == c2) break block99;
                                                                                        byArray[n7 + n3] = 63;
                                                                                        if (!bl) break block73;
                                                                                    }
                                                                                    byArray[n3 - -n7] = -97;
                                                                                    if (!bl) break block73;
                                                                                }
                                                                                byArray[n3 - -n7] = -98;
                                                                                if (!bl) break block73;
                                                                            }
                                                                            byArray[n7 + n3] = -101;
                                                                            if (!bl) break block73;
                                                                        }
                                                                        byArray[n7 + n3] = -104;
                                                                        if (!bl) break block73;
                                                                    }
                                                                    byArray[n3 + n7] = -105;
                                                                    if (!bl) break block73;
                                                                }
                                                                byArray[n3 + n7] = -106;
                                                                if (!bl) break block73;
                                                            }
                                                            byArray[n7 + n3] = -110;
                                                            if (!bl) break block73;
                                                        }
                                                        byArray[n7 + n3] = -111;
                                                        if (!bl) break block73;
                                                    }
                                                    byArray[n7 + n3] = -114;
                                                    if (!bl) break block73;
                                                }
                                                byArray[n7 + n3] = -116;
                                                if (!bl) break block73;
                                            }
                                            byArray[n7 + n3] = -117;
                                            if (!bl) break block73;
                                        }
                                        byArray[n7 + n3] = -118;
                                        if (!bl) break block73;
                                    }
                                    byArray[n3 + n7] = -120;
                                    if (!bl) break block73;
                                }
                                byArray[n7 + n3] = -123;
                                if (!bl) break block73;
                            }
                            byArray[n3 + n7] = -125;
                            if (!bl) break block73;
                        }
                        byArray[n7 + n3] = -126;
                        if (!bl) break block73;
                    }
                    byArray[n7 + n3] = (byte)c2;
                }
                ++n7;
                if (!bl) continue;
            }
            n5 = n6;
            return n5;
        }
        catch (RuntimeException runtimeException) {
            String string;
            StringBuilder stringBuilder = new StringBuilder().append(z[4]).append(n2).append(',').append(n3).append(',').append(n4).append(',').append(charSequence != null ? z[2] : z[1]).append(',').append(by).append(',');
            if (byArray != null) {
                string = z[2];
                throw i.a(runtimeException, stringBuilder.append(string).append(')').toString());
            }
            string = z[1];
            throw i.a(runtimeException, stringBuilder.append(string).append(')').toString());
        }
    }

    i(String string, int n2) {
        try {
            this.a = n2;
        }
        catch (RuntimeException runtimeException) {
            RuntimeException runtimeException2 = runtimeException;
            StringBuilder stringBuilder = new StringBuilder().append(z[0]);
            String string2 = string != null ? z[2] : z[1];
            throw i.a(runtimeException2, stringBuilder.append(string2).append(',').append(n2).append(')').toString());
        }
    }

    static final la a(Throwable throwable, String string) {
        la la2;
        ++c;
        if (throwable instanceof la) {
            la2 = (la)throwable;
            la2.h = la2.h + ' ' + string;
        } else {
            la2 = new la(throwable, string);
        }
        return la2;
    }

    static final v[] a(int n2) {
        try {
            ++d;
            if (n2 != -711) {
                i.a(null, null);
            }
            return new v[]{ua.E, da.O, ga.c, ta.f, la.b, eb.d, gb.n};
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, z[3] + n2 + ')');
        }
    }

    static {
        z = new String[]{i.z(i.z("s\u0000\u0011!|sZ\u0013`")), i.z(i.z("t[A$")), i.z(i.z("a\u0000\u0003fo")), i.z(i.z("s\u0000n`")), i.z(i.z("s\u0000o`")), i.z(i.z("s\u0000Y'An\\D&u2\u0007"))};
        h = new long[256];
        for (int i2 = 0; i2 < 256; ++i2) {
            long l2 = i2;
            int n2 = 0;
            while (true) {
                block5: {
                    block4: {
                        if (8 <= n2) break;
                        if ((1L & l2 ^ 0xFFFFFFFFFFFFFFFFL) != -2L) break block4;
                        l2 = l2 >>> -102854015 ^ 0xC96C5795D7870F42L;
                        break block5;
                    }
                    l2 >>>= 1;
                }
                ++n2;
            }
            i.h[i2] = l2;
        }
        f = i.z(i.z("[ln\fW\\ie\u0001XQb`\u0006]J\u007f\u007f\u001bFOxz\u0010K@OO+v\u007fHJ {pEA%|u^\\:an[[?jcT\u001dy )\u001a\u0018~%\"\u0017\fj\u00b1>\u000bsn82\u0007\u0000\u0017/1uV\u0015o!\u0014\n\b1d\u0002\u0011f,5\u0011q42"));
    }

    private static char[] z(String string) {
        char[] cArray = string.toCharArray();
        if (cArray.length < 2) {
            cArray = cArray;
            cArray[0] = (char)(cArray[0] ^ 0x12);
        }
        return cArray;
    }

    /*
     * Handled impossible loop by duplicating code
     * Enabled aggressive block sorting
     */
    private static String z(char[] cArray) {
        char[] cArray2;
        block9: {
            int n2;
            int n3;
            block8: {
                cArray2 = cArray;
                n3 = cArray.length;
                n2 = 0;
                if (!true) break block8;
                n3 = n3;
                if (n3 <= n2) break block9;
            }
            do {
                int n4;
                cArray2 = cArray2;
                int n5 = n2;
                char c2 = cArray2[n5];
                switch (n2 % 5) {
                    case 0: {
                        n4 = 26;
                        break;
                    }
                    case 1: {
                        n4 = 46;
                        break;
                    }
                    case 2: {
                        n4 = 45;
                        break;
                    }
                    case 3: {
                        n4 = 72;
                        break;
                    }
                    default: {
                        n4 = 18;
                    }
                }
                cArray2[n5] = (char)(c2 ^ n4);
                ++n2;
                n3 = n3;
            } while (n3 > n2);
        }
        return new String(cArray2).intern();
    }
}

