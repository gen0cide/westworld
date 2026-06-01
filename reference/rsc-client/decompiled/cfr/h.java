/*
 * Decompiled with CFR 0.152.
 */
final class h {
    static int[] c;
    static int d;
    static int[] b;
    static int a;
    static String[] e;
    private static final String[] z;

    /*
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    static final byte[] a(CharSequence charSequence, byte by) {
        boolean bl = client.vh;
        try {
            int n2 = -91 / ((by - -26) / 53);
            ++d;
            int n3 = charSequence.length();
            byte[] byArray = new byte[n3];
            int n4 = 0;
            do {
                block70: {
                    char c2;
                    block67: {
                        block68: {
                            block71: {
                                block72: {
                                    block74: {
                                        block75: {
                                            block76: {
                                                block78: {
                                                    block81: {
                                                        block82: {
                                                            block83: {
                                                                block89: {
                                                                    block90: {
                                                                        block91: {
                                                                            block92: {
                                                                                block93: {
                                                                                    block94: {
                                                                                        block95: {
                                                                                            block88: {
                                                                                                block87: {
                                                                                                    block86: {
                                                                                                        block85: {
                                                                                                            block84: {
                                                                                                                block80: {
                                                                                                                    block79: {
                                                                                                                        block77: {
                                                                                                                            block73: {
                                                                                                                                block69: {
                                                                                                                                    if (~n3 >= ~n4) return byArray;
                                                                                                                                    c2 = charSequence.charAt(n4);
                                                                                                                                    if (c2 > '\u0000' && c2 < '\u0080' || -161 >= ~c2 && c2 <= '\u00ff') break block67;
                                                                                                                                    if (c2 == '\u20ac') break block68;
                                                                                                                                    if (c2 != '\u201a') break block69;
                                                                                                                                    byArray[n4] = -126;
                                                                                                                                    if (!bl) break block70;
                                                                                                                                }
                                                                                                                                if ('\u0192' == c2) break block71;
                                                                                                                                if (~c2 == -8223) break block72;
                                                                                                                                if ('\u2026' != c2) break block73;
                                                                                                                                byArray[n4] = -123;
                                                                                                                                if (!bl) break block70;
                                                                                                                            }
                                                                                                                            if (~c2 == -8225) break block74;
                                                                                                                            if (-8226 == ~c2) break block75;
                                                                                                                            if ('\u02c6' == c2) break block76;
                                                                                                                            if (-8241 != ~c2) break block77;
                                                                                                                            byArray[n4] = -119;
                                                                                                                            if (!bl) break block70;
                                                                                                                        }
                                                                                                                        if (c2 == '\u0160') break block78;
                                                                                                                        if (c2 != '\u2039') break block79;
                                                                                                                        byArray[n4] = -117;
                                                                                                                        if (!bl) break block70;
                                                                                                                    }
                                                                                                                    if (c2 != '\u0152') break block80;
                                                                                                                    byArray[n4] = -116;
                                                                                                                    if (!bl) break block70;
                                                                                                                }
                                                                                                                if (c2 == '\u017d') break block81;
                                                                                                                if ('\u2018' == c2) break block82;
                                                                                                                if (-8218 == ~c2) break block83;
                                                                                                                if ('\u201c' != c2) break block84;
                                                                                                                byArray[n4] = -109;
                                                                                                                if (!bl) break block70;
                                                                                                            }
                                                                                                            if ('\u201d' != c2) break block85;
                                                                                                            byArray[n4] = -108;
                                                                                                            if (!bl) break block70;
                                                                                                        }
                                                                                                        if (c2 != '\u2022') break block86;
                                                                                                        byArray[n4] = -107;
                                                                                                        if (!bl) break block70;
                                                                                                    }
                                                                                                    if ('\u2013' != c2) break block87;
                                                                                                    byArray[n4] = -106;
                                                                                                    if (!bl) break block70;
                                                                                                }
                                                                                                if ('\u2014' != c2) break block88;
                                                                                                byArray[n4] = -105;
                                                                                                if (!bl) break block70;
                                                                                            }
                                                                                            if (-733 == ~c2) break block89;
                                                                                            if (~c2 == -8483) break block90;
                                                                                            if (-354 == ~c2) break block91;
                                                                                            if ('\u203a' == c2) break block92;
                                                                                            if (~c2 == -340) break block93;
                                                                                            if ('\u017e' == c2) break block94;
                                                                                            if (~c2 != -377) break block95;
                                                                                            byArray[n4] = -97;
                                                                                            if (!bl) break block70;
                                                                                        }
                                                                                        byArray[n4] = 63;
                                                                                        if (!bl) break block70;
                                                                                    }
                                                                                    byArray[n4] = -98;
                                                                                    if (!bl) break block70;
                                                                                }
                                                                                byArray[n4] = -100;
                                                                                if (!bl) break block70;
                                                                            }
                                                                            byArray[n4] = -101;
                                                                            if (!bl) break block70;
                                                                        }
                                                                        byArray[n4] = -102;
                                                                        if (!bl) break block70;
                                                                    }
                                                                    byArray[n4] = -103;
                                                                    if (!bl) break block70;
                                                                }
                                                                byArray[n4] = -104;
                                                                if (!bl) break block70;
                                                            }
                                                            byArray[n4] = -110;
                                                            if (!bl) break block70;
                                                        }
                                                        byArray[n4] = -111;
                                                        if (!bl) break block70;
                                                    }
                                                    byArray[n4] = -114;
                                                    if (!bl) break block70;
                                                }
                                                byArray[n4] = -118;
                                                if (!bl) break block70;
                                            }
                                            byArray[n4] = -120;
                                            if (!bl) break block70;
                                        }
                                        byArray[n4] = -121;
                                        if (!bl) break block70;
                                    }
                                    byArray[n4] = -122;
                                    if (!bl) break block70;
                                }
                                byArray[n4] = -124;
                                if (!bl) break block70;
                            }
                            byArray[n4] = -125;
                            if (!bl) break block70;
                        }
                        byArray[n4] = -128;
                        if (!bl) break block70;
                    }
                    byArray[n4] = (byte)c2;
                }
                ++n4;
            } while (!bl);
            return byArray;
        }
        catch (RuntimeException runtimeException) {
            String string;
            StringBuilder stringBuilder = new StringBuilder().append(z[2]);
            if (charSequence != null) {
                string = z[0];
                throw i.a(runtimeException, stringBuilder.append(string).append(',').append(by).append(')').toString());
            }
            string = z[1];
            throw i.a(runtimeException, stringBuilder.append(string).append(',').append(by).append(')').toString());
        }
    }

    static {
        z = new String[]{h.z(h.z("~\u0004$[5")), h.z(h.z("k_f\u0019")), h.z(h.z("m\u0004K]"))};
    }

    private static char[] z(String string) {
        char[] cArray = string.toCharArray();
        if (cArray.length < 2) {
            cArray = cArray;
            cArray[0] = (char)(cArray[0] ^ 0x48);
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
                        n4 = 5;
                        break;
                    }
                    case 1: {
                        n4 = 42;
                        break;
                    }
                    case 2: {
                        n4 = 10;
                        break;
                    }
                    case 3: {
                        n4 = 117;
                        break;
                    }
                    default: {
                        n4 = 72;
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

