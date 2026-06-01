/*
 * Decompiled with CFR 0.152.
 */
import java.io.EOFException;
import java.io.IOException;

final class ob {
    static int e;
    static int d;
    static int g;
    private int c;
    static byte[][] j;
    private int a = 65000;
    private nb b = null;
    private nb f = null;
    static int i;
    static int[] h;
    private static final String[] z;

    /*
     * WARNING - Removed try catching itself - possible behaviour change.
     * Enabled force condition propagation
     * Lifted jumps to return sites
     */
    final boolean a(int n2, int n3, int n4, byte[] byArray) {
        try {
            int n5 = 94 % ((n4 - -61) / 35);
            ++d;
            nb nb2 = this.f;
            synchronized (nb2) {
                if (-1 < ~n3) throw new IllegalArgumentException();
                if (this.a < n3) {
                    throw new IllegalArgumentException();
                }
                boolean bl = this.a(n2, byArray, true, n3, 4);
                if (!bl) return this.a(n2, byArray, false, n3, 4);
                return bl;
            }
        }
        catch (RuntimeException runtimeException) {
            RuntimeException runtimeException2 = runtimeException;
            StringBuilder stringBuilder = new StringBuilder().append(z[2]).append(n2).append(',').append(n3).append(',').append(n4).append(',');
            String string = byArray != null ? z[3] : z[4];
            throw i.a(runtimeException2, stringBuilder.append(string).append(')').toString());
        }
    }

    /*
     * WARNING - Removed try catching itself - possible behaviour change.
     * Unable to fully structure code
     */
    final byte[] a(int var1_1, int var2_2) {
        var18_3 = client.vh;
        try {
            ++ob.i;
            var3_4 = this.f;
            synchronized (var3_4) {
                try {
                    block34: {
                        block39: {
                            block38: {
                                block37: {
                                    block36: {
                                        block35: {
                                            block41: {
                                                block40: {
                                                    block33: {
                                                        block32: {
                                                            block31: {
                                                                block30: {
                                                                    block29: {
                                                                        block28: {
                                                                            if ((this.b.a((byte)-111) ^ -1L) > ((long)(6 + var2_2 * 6) ^ -1L)) {
                                                                                return null;
                                                                            }
                                                                            this.b.a(6 * var2_2, 12);
                                                                            this.b.a(true, 6, 0, la.c);
                                                                            var4_6 = (65280 & la.c[1] << 1647444456) + (((255 & la.c[0]) << -463794928) + (255 & la.c[2]));
                                                                            var5_8 = (la.c[4] << 553924200 & 65280) + ((la.c[3] & 255) << -1798669456) - -(255 & la.c[5]);
                                                                            if (var4_6 < 0 || ~var4_6 < ~this.a) break block28;
                                                                            break block29;
                                                                        }
                                                                        return null;
                                                                    }
                                                                    if (~var5_8 >= -1 || (this.f.a((byte)-111) / 520L ^ -1L) > ((long)var5_8 ^ -1L)) break block30;
                                                                    break block31;
                                                                }
                                                                return null;
                                                            }
                                                            if (var1_1 != 9395) {
                                                                return null;
                                                            }
                                                            var6_9 = new byte[var4_6];
                                                            var7_10 = 0;
                                                            var8_11 = 0;
lbl36:
                                                            // 2 sources

                                                            while (var4_6 > var7_10) {
                                                                if (var5_8 == 0) break block32;
                                                                break block33;
                                                            }
                                                            break block34;
                                                        }
                                                        return null;
                                                    }
                                                    this.f.a(var5_8 * 520, 107);
                                                    var9_12 = -var7_10 + var4_6;
                                                    if (var2_2 <= 65535) break block40;
                                                    if (~var9_12 < -511) {
                                                        var9_12 = 510;
                                                    }
                                                    var14_17 = 10;
                                                    this.f.a(true, var9_12 + var14_17, 0, la.c);
                                                    var10_13 = (la.c[3] & 255) + ((la.c[0] << -854702312 & -16777216) - -(0xFF0000 & la.c[1] << -511327312) + (65280 & la.c[2] << 2104943048));
                                                    var13_16 = la.c[9] & 255;
                                                    var11_14 = (la.c[5] & 255) + ((255 & la.c[4]) << 0x6988888);
                                                    var12_15 = (la.c[8] & 255) + (0xFF0000 & la.c[6] << -1936015824) + (la.c[7] << 2101716552 & 65280);
                                                    if (!var18_3) break block41;
                                                }
                                                if (~var9_12 < -513) {
                                                    var9_12 = 512;
                                                }
                                                var14_17 = 8;
                                                this.f.a(true, var14_17 + var9_12, 0, la.c);
                                                var13_16 = la.c[7] & 255;
                                                var10_13 = (la.c[0] << 524989928 & 65280) + (255 & la.c[1]);
                                                var11_14 = (la.c[2] << -1311007928 & 65280) + (la.c[3] & 255);
                                                var12_15 = ((255 & la.c[4]) << 862358512) - -((la.c[5] & 255) << 550121000) - -(255 & la.c[6]);
                                            }
                                            if (~var2_2 != ~var10_13 || var8_11 != var11_14 || ~var13_16 != ~this.c) break block35;
                                            break block36;
                                        }
                                        return null;
                                    }
                                    if (var12_15 < 0 || (this.f.a((byte)-111) / 520L ^ -1L) > ((long)var12_15 ^ -1L)) break block37;
                                    break block38;
                                }
                                return null;
                            }
                            var15_18 = var14_17 - -var9_12;
                            var5_8 = var12_15;
                            for (var16_19 = var14_17; var15_18 > var16_19; ++var16_19) {
                                var6_9[var7_10++] = la.c[var16_19];
                                if (!var18_3) {
                                    if (!var18_3) continue;
                                }
                                break block39;
                            }
                            ++var8_11;
                        }
                        if (!var18_3) ** GOTO lbl36
                    }
                    return var6_9;
                }
                catch (IOException var4_7) {
                    return null;
                }
            }
        }
        catch (RuntimeException var3_5) {
            throw i.a(var3_5, ob.z[0] + var1_1 + ',' + var2_2 + ')');
        }
    }

    /*
     * WARNING - Removed try catching itself - possible behaviour change.
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    private final boolean a(int n2, byte[] byArray, boolean bl, int n3, int n4) {
        boolean bl2 = client.vh;
        try {
            ++e;
            nb nb2 = this.f;
            synchronized (nb2) {
                try {
                    int n5;
                    if (bl) {
                        if (((long)(6 + 6 * n2) ^ 0xFFFFFFFFFFFFFFFFL) < (this.b.a((byte)-111) ^ 0xFFFFFFFFFFFFFFFFL)) {
                            return false;
                        }
                        this.b.a(n2 * 6, -124);
                        this.b.a(true, 6, 0, la.c);
                        n5 = (0xFF & la.c[5]) + (la.c[3] << 4539376 & 0xFF0000) - -((la.c[4] & 0xFF) << 1519652136);
                        if (n5 <= 0) return false;
                        if (this.f.a((byte)-111) / 520L < (long)n5) return false;
                    } else {
                        n5 = (int)((this.f.a((byte)-111) - -519L) / 520L);
                        if (0 == n5) {
                            n5 = 1;
                        }
                    }
                    la.c[0] = (byte)(n3 >> 1377278608);
                    la.c[2] = (byte)n3;
                    la.c[3] = (byte)(n5 >> -340974608);
                    la.c[n4] = (byte)(n5 >> -1558555128);
                    la.c[1] = (byte)(n3 >> -236170680);
                    la.c[5] = (byte)n5;
                    this.b.a(6 * n2, 31);
                    this.b.a(6, -102, 0, la.c);
                    int n6 = 0;
                    int n7 = 0;
                    do {
                        int n8;
                        block46: {
                            int n9;
                            block45: {
                                block42: {
                                    int n10;
                                    int n11;
                                    block44: {
                                        block43: {
                                            if (~n6 <= ~n3) return true;
                                            n8 = 0;
                                            boolean bl3 = bl;
                                            if (bl2) return bl3;
                                            if (!bl3) break block42;
                                            this.f.a(520 * n5, n4 ^ 0x11);
                                            if (n2 > 65535) break block43;
                                            try {
                                                this.f.a(true, 8, 0, la.c);
                                            }
                                            catch (EOFException eOFException) {
                                                if (!bl2) return true;
                                            }
                                            n9 = (0xFF00 & la.c[0] << -1162714328) - -(la.c[1] & 0xFF);
                                            n8 = (la.c[6] & 0xFF) + (0xFF0000 & la.c[4] << -165446384) - -((0xFF & la.c[5]) << -892700824);
                                            n11 = 0xFF & la.c[7];
                                            n10 = (la.c[3] & 0xFF) + (la.c[2] << -941949624 & 0xFF00);
                                            if (!bl2) break block44;
                                        }
                                        try {
                                            this.f.a(true, 10, 0, la.c);
                                        }
                                        catch (EOFException eOFException) {
                                            if (!bl2) return true;
                                        }
                                        n10 = (0xFF & la.c[5]) + (la.c[4] << -1447609528 & 0xFF00);
                                        n8 = (la.c[6] << 725235568 & 0xFF0000) - -((la.c[7] & 0xFF) << 221297448) + (la.c[8] & 0xFF);
                                        n9 = (la.c[0] << -2111158472 & 0xFF000000) + ((la.c[1] << -1296842608 & 0xFF0000) - -((0xFF & la.c[2]) << -1391218456)) - -(0xFF & la.c[3]);
                                        n11 = 0xFF & la.c[9];
                                    }
                                    if (~n2 != ~n9) return false;
                                    if (n10 != n7) return false;
                                    if (n11 != this.c) return false;
                                    if (0 > n8) return false;
                                    if (((long)n8 ^ 0xFFFFFFFFFFFFFFFFL) < (this.f.a((byte)-111) / 520L ^ 0xFFFFFFFFFFFFFFFFL)) {
                                        return false;
                                    }
                                }
                                if (~n8 == -1) {
                                    bl = false;
                                    n8 = (int)((this.f.a((byte)-111) - -519L) / 520L);
                                    if (0 == n8) {
                                        ++n8;
                                    }
                                    if (n8 == n5) {
                                        ++n8;
                                    }
                                }
                                if (512 >= n3 - n6) {
                                    n8 = 0;
                                }
                                if (-65536 <= ~n2) break block45;
                                la.c[0] = (byte)(n2 >> -322435432);
                                la.c[5] = (byte)n7;
                                la.c[2] = (byte)(n2 >> -936822456);
                                la.c[4] = (byte)(n7 >> 521649416);
                                la.c[7] = (byte)(n8 >> -876359160);
                                la.c[1] = (byte)(n2 >> 527626448);
                                la.c[8] = (byte)n8;
                                la.c[9] = (byte)this.c;
                                la.c[3] = (byte)n2;
                                la.c[6] = (byte)(n8 >> -1755814992);
                                this.f.a(520 * n5, n4 ^ 0x21);
                                this.f.a(10, -111, 0, la.c);
                                n9 = n3 + -n6;
                                if (-511 > ~n9) {
                                    n9 = 510;
                                }
                                this.f.a(n9, n4 + -119, n6, byArray);
                                n6 += n9;
                                if (!bl2) break block46;
                            }
                            la.c[4] = (byte)(n8 >> 1803659792);
                            la.c[0] = (byte)(n2 >> -566551576);
                            la.c[7] = (byte)this.c;
                            la.c[6] = (byte)n8;
                            la.c[3] = (byte)n7;
                            la.c[1] = (byte)n2;
                            la.c[2] = (byte)(n7 >> -447508120);
                            la.c[5] = (byte)(n8 >> -1738018904);
                            this.f.a(n5 * 520, n4 ^ 0x7F);
                            this.f.a(8, -107, 0, la.c);
                            n9 = n3 - n6;
                            if (~n9 < -513) {
                                n9 = 512;
                            }
                            this.f.a(n9, n4 + -125, n6, byArray);
                            n6 += n9;
                        }
                        n5 = n8;
                        ++n7;
                    } while (!bl2);
                    return true;
                }
                catch (IOException iOException) {
                    return false;
                }
            }
        }
        catch (RuntimeException runtimeException) {
            String string;
            StringBuilder stringBuilder = new StringBuilder().append(z[5]).append(n2).append(',');
            if (byArray != null) {
                string = z[3];
                throw i.a(runtimeException, stringBuilder.append(string).append(',').append(bl).append(',').append(n3).append(',').append(n4).append(')').toString());
            }
            string = z[4];
            throw i.a(runtimeException, stringBuilder.append(string).append(',').append(bl).append(',').append(n3).append(',').append(n4).append(')').toString());
        }
    }

    public final String toString() {
        try {
            ++g;
            return "" + this.c;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, z[1]);
        }
    }

    ob(int n2, nb nb2, nb nb3, int n3) {
        try {
            this.a = n3;
            this.c = n2;
            this.b = nb3;
            this.f = nb2;
        }
        catch (RuntimeException runtimeException) {
            RuntimeException runtimeException2 = runtimeException;
            StringBuilder stringBuilder = new StringBuilder().append(z[6]).append(n2).append(',');
            String string = nb2 != null ? z[3] : z[4];
            StringBuilder stringBuilder2 = stringBuilder.append(string).append(',');
            String string2 = nb3 != null ? z[3] : z[4];
            throw i.a(runtimeException2, stringBuilder2.append(string2).append(',').append(n3).append(')').toString());
        }
    }

    static {
        z = new String[]{ob.z(ob.z("\u0011\u0001]5W")), ob.z(ob.z("\u0011\u0001]\u0000\u0010-\u0017\u0001\u001d\u0011\u0019KZ")), ob.z(ob.z("\u0011\u0001]6W")), ob.z(ob.z("\u0005M]Z\u0002")), ob.z(ob.z("\u0010\u0016\u001f\u0018")), ob.z(ob.z("\u0011\u0001]7W")), ob.z(ob.z("\u0011\u0001]H\u0016\u0010\n\u0007JW"))};
        j = new byte[1000][];
    }

    private static char[] z(String string) {
        char[] cArray = string.toCharArray();
        if (cArray.length < 2) {
            cArray = cArray;
            cArray[0] = (char)(cArray[0] ^ 0x7F);
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
                        n4 = 126;
                        break;
                    }
                    case 1: {
                        n4 = 99;
                        break;
                    }
                    case 2: {
                        n4 = 115;
                        break;
                    }
                    case 3: {
                        n4 = 116;
                        break;
                    }
                    default: {
                        n4 = 127;
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

