/*
 * Decompiled with CFR 0.152.
 */
final class aa {
    static int e;
    static int h;
    static int a;
    static int[] c;
    static int b;
    static int d;
    private int[] j;
    private byte[] i;
    static int l;
    private int[] g;
    static int f;
    static String[] k;
    private static final String[] z;

    final int a(int n2, byte[] byArray, int n3, byte[] byArray2, int n4, int n5) {
        boolean bl = client.vh;
        try {
            int n6;
            int n7;
            block17: {
                ++h;
                if (n5 < 99) {
                    this.a(-58, (byte[])null, 69, null, -39, 22);
                }
                int n8 = 0;
                n2 += n4;
                int n9 = n3 << 1494901059;
                while (n4 < n2) {
                    block20: {
                        byte by;
                        int n10;
                        block19: {
                            int n11;
                            block18: {
                                n11 = byArray2[n4] & 0xFF;
                                n10 = this.g[n11];
                                by = this.i[n11];
                                n7 = ~by;
                                n6 = -1;
                                if (bl) break block17;
                                if (n7 == n6) break block18;
                                break block19;
                            }
                            throw new RuntimeException("" + n11);
                        }
                        int n12 = n9 >> 1897076227;
                        int n13 = n9 & 7;
                        n8 &= -n13 >> -2036172737;
                        int n14 = n12 + (by + n13 + -1 >> 1085645667);
                        n9 += by;
                        n8 = d.a(n8, n10 >>> (n13 += 24));
                        byArray[n12] = (byte)n8;
                        if (~n14 >= ~n12 && !bl) break block20;
                        ++n12;
                        n8 = n10 >>> (n13 -= 8);
                        byArray[n12] = (byte)n8;
                        if (~n12 <= ~n14 && !bl) break block20;
                        ++n12;
                        n8 = n10 >>> (n13 -= 8);
                        byArray[n12] = (byte)n8;
                        if (n12 >= n14 && !bl) break block20;
                        n8 = n10 >>> (n13 -= 8);
                        byArray[++n12] = (byte)n8;
                        if (n12 < n14) {
                            n8 = n10 << -(n13 -= 8);
                            byArray[++n12] = (byte)n8;
                        }
                    }
                    ++n4;
                    if (!bl) continue;
                }
                n7 = -n3;
                n6 = n9 + 7 >> -261766397;
            }
            return n7 + n6;
        }
        catch (RuntimeException runtimeException) {
            RuntimeException runtimeException2 = runtimeException;
            StringBuilder stringBuilder = new StringBuilder().append(z[3]).append(n2).append(',');
            String string = byArray != null ? z[1] : z[2];
            StringBuilder stringBuilder2 = stringBuilder.append(string).append(',').append(n3).append(',');
            String string2 = byArray2 != null ? z[1] : z[2];
            throw i.a(runtimeException2, stringBuilder2.append(string2).append(',').append(n4).append(',').append(n5).append(')').toString());
        }
    }

    /*
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    final int a(byte[] byArray, byte[] byArray2, int n2, int n3, int n4, int n5) {
        boolean bl = client.vh;
        try {
            ++a;
            if (0 == n5) {
                return 0;
            }
            int n6 = 0;
            n5 += n2;
            if (n4 != -1) {
                this.a(105, (byte[])null, 82, null, 125, -45);
            }
            int n7 = n3;
            do {
                int n8;
                block80: {
                    block79: {
                        byte by;
                        block78: {
                            block77: {
                                block76: {
                                    block75: {
                                        block74: {
                                            block73: {
                                                block72: {
                                                    block71: {
                                                        block70: {
                                                            block69: {
                                                                block68: {
                                                                    block67: {
                                                                        block66: {
                                                                            block65: {
                                                                                if ((by = byArray[n7]) < 0) break block65;
                                                                                ++n6;
                                                                                if (bl || !bl) break block66;
                                                                            }
                                                                            n6 = this.j[n6];
                                                                        }
                                                                        if (0 > (n8 = this.j[n6])) {
                                                                            byArray2[n2++] = (byte)(~n8);
                                                                            if (n2 >= n5) return -n3 + 1 + n7;
                                                                            n6 = 0;
                                                                        }
                                                                        if ((0x40 & by) == 0) break block67;
                                                                        n6 = this.j[n6];
                                                                        if (!bl) break block68;
                                                                    }
                                                                    ++n6;
                                                                }
                                                                if ((n8 = this.j[n6]) < 0) {
                                                                    byArray2[n2++] = (byte)(~n8);
                                                                    if (n5 <= n2) {
                                                                        if (!bl) return -n3 + 1 + n7;
                                                                    }
                                                                    n6 = 0;
                                                                }
                                                                if (~(by & 0x20) != -1) break block69;
                                                                ++n6;
                                                                if (!bl) break block70;
                                                            }
                                                            n6 = this.j[n6];
                                                        }
                                                        if (~(n8 = this.j[n6]) > -1) {
                                                            byArray2[n2++] = (byte)(~n8);
                                                            if (~n2 <= ~n5) {
                                                                if (!bl) return -n3 + 1 + n7;
                                                            }
                                                            n6 = 0;
                                                        }
                                                        if (0 == (0x10 & by)) break block71;
                                                        n6 = this.j[n6];
                                                        if (!bl) break block72;
                                                    }
                                                    ++n6;
                                                }
                                                if (~(n8 = this.j[n6]) > -1) {
                                                    byArray2[n2++] = (byte)(~n8);
                                                    if (n2 >= n5) {
                                                        if (!bl) return -n3 + 1 + n7;
                                                    }
                                                    n6 = 0;
                                                }
                                                if (-1 == ~(by & 8)) break block73;
                                                n6 = this.j[n6];
                                                if (!bl) break block74;
                                            }
                                            ++n6;
                                        }
                                        if (-1 < ~(n8 = this.j[n6])) {
                                            byArray2[n2++] = (byte)(~n8);
                                            if (n2 >= n5) return -n3 + 1 + n7;
                                            n6 = 0;
                                        }
                                        if ((4 & by) == 0) break block75;
                                        n6 = this.j[n6];
                                        if (!bl) break block76;
                                    }
                                    ++n6;
                                }
                                if (0 > (n8 = this.j[n6])) {
                                    byArray2[n2++] = (byte)(~n8);
                                    if (~n5 >= ~n2) {
                                        if (!bl) return -n3 + 1 + n7;
                                    }
                                    n6 = 0;
                                }
                                if (0 != (2 & by)) break block77;
                                ++n6;
                                if (!bl) break block78;
                            }
                            n6 = this.j[n6];
                        }
                        if (~(n8 = this.j[n6]) > -1) {
                            byArray2[n2++] = (byte)(~n8);
                            if (~n5 >= ~n2) {
                                if (!bl) return -n3 + 1 + n7;
                            }
                            n6 = 0;
                        }
                        if (-1 == ~(1 & by)) break block79;
                        n6 = this.j[n6];
                        if (!bl) break block80;
                    }
                    ++n6;
                }
                if (~(n8 = this.j[n6]) > -1) {
                    byArray2[n2++] = (byte)(~n8);
                    if (~n5 >= ~n2) return -n3 + 1 + n7;
                    n6 = 0;
                }
                ++n7;
            } while (!bl);
            return -n3 + 1 + n7;
        }
        catch (RuntimeException runtimeException) {
            String string;
            StringBuilder stringBuilder = new StringBuilder().append(z[4]).append(byArray != null ? z[1] : z[2]).append(',');
            if (byArray2 != null) {
                string = z[1];
                throw i.a(runtimeException, stringBuilder.append(string).append(',').append(n2).append(',').append(n3).append(',').append(n4).append(',').append(n5).append(')').toString());
            }
            string = z[2];
            throw i.a(runtimeException, stringBuilder.append(string).append(',').append(n2).append(',').append(n3).append(',').append(n4).append(',').append(n5).append(')').toString());
        }
    }

    static final int a(int n2, boolean bl) {
        try {
            ++e;
            --n2;
            n2 |= n2 >>> 311248929;
            n2 |= n2 >>> -1130998654;
            n2 |= n2 >>> -669289052;
            if (bl) {
                b = -4;
            }
            n2 |= n2 >>> -948655896;
            n2 |= n2 >>> 795067056;
            return n2 - -1;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, z[5] + n2 + ',' + bl + ')');
        }
    }

    /*
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    aa(byte[] byArray) {
        try {
            int n2 = byArray.length;
            this.g = new int[n2];
            this.i = byArray;
            this.j = new int[8];
            int[] nArray = new int[33];
            int n3 = 0;
            int n4 = 0;
            while (~n4 > ~n2) {
                byte by = byArray[n4];
                if (0 != by) {
                    int n6;
                    int n7;
                    int n5;
                    int n8;
                    int n9;
                    int n10 = 1 << -by + 32;
                    this.g[n4] = n9 = nArray[by];
                    if (-1 != ~(n10 & n9)) {
                        n8 = nArray[by - 1];
                    } else {
                        n5 = by - 1;
                        while (~n5 <= -2 && n9 == (n7 = nArray[n5])) {
                            n6 = 1 << 32 - n5;
                            if ((n7 & n6) != 0) {
                                nArray[n5] = nArray[n5 - 1];
                                break;
                            }
                            nArray[n5] = d.a(n6, n7);
                            --n5;
                        }
                        n8 = n9 | n10;
                    }
                    nArray[by] = n8;
                    for (n5 = by - -1; n5 <= 32; ++n5) {
                        if (~n9 != ~nArray[n5]) continue;
                        nArray[n5] = n8;
                    }
                    n5 = 0;
                    for (n7 = 0; by > n7; n6 >>>= 1, ++n7) {
                        n6 = Integer.MIN_VALUE >>> n7;
                        if (-1 == ~(n6 & n9)) {
                            ++n5;
                        } else {
                            if (~this.j[n5] == -1) {
                                this.j[n5] = n3;
                            }
                            n5 = this.j[n5];
                        }
                        if (this.j.length > n5) continue;
                        int[] nArray2 = new int[this.j.length * 2];
                        int n11 = 0;
                        while (~this.j.length < ~n11) {
                            nArray2[n11] = this.j[n11];
                            ++n11;
                        }
                        this.j = nArray2;
                    }
                    if (~n3 >= ~n5) {
                        n3 = n5 + 1;
                    }
                    this.j[n5] = ~n4;
                }
                ++n4;
            }
            return;
        }
        catch (RuntimeException runtimeException) {
            String string;
            StringBuilder stringBuilder = new StringBuilder().append(z[0]);
            if (byArray != null) {
                string = z[1];
                throw i.a(runtimeException, stringBuilder.append(string).append(')').toString());
            }
            string = z[2];
            throw i.a(runtimeException, stringBuilder.append(string).append(')').toString());
        }
    }

    static {
        z = new String[]{aa.z(aa.z("7L\u001d]e8DG_$")), aa.z(aa.z("-\u0003\u001dOq")), aa.z(aa.z("8X_\r")), aa.z(aa.z("7L\u001d#$")), aa.z(aa.z("7L\u001d $")), aa.z(aa.z("7L\u001d\"$"))};
        d = 114;
        k = new String[100];
    }

    private static char[] z(String string) {
        char[] cArray = string.toCharArray();
        if (cArray.length < 2) {
            cArray = cArray;
            cArray[0] = (char)(cArray[0] ^ 0xC);
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
                        n4 = 86;
                        break;
                    }
                    case 1: {
                        n4 = 45;
                        break;
                    }
                    case 2: {
                        n4 = 51;
                        break;
                    }
                    case 3: {
                        n4 = 97;
                        break;
                    }
                    default: {
                        n4 = 12;
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

