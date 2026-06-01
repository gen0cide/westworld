/*
 * Decompiled with CFR 0.152.
 */
final class db {
    static int h;
    static int[] j;
    static int d;
    static int[] l;
    static byte[] i;
    static int a;
    static i f;
    ib k = new ib();
    static int e;
    static int g;
    static int c;
    private ib b;
    private static final String[] z;

    final void a(ib ib2, boolean bl) {
        try {
            ++h;
            if (null != ib2.e) {
                ib2.a(-27331);
            }
            ib2.e = this.k;
            ib2.a = this.k.a;
            ib2.e.a = ib2;
            ib2.a.e = ib2;
            if (bl) {
                this.b((byte)78);
            }
        }
        catch (RuntimeException runtimeException) {
            RuntimeException runtimeException2 = runtimeException;
            StringBuilder stringBuilder = new StringBuilder().append(z[4]);
            String string = ib2 != null ? z[2] : z[3];
            throw i.a(runtimeException2, stringBuilder.append(string).append(',').append(bl).append(')').toString());
        }
    }

    final ib a(byte by) {
        try {
            ++c;
            ib ib2 = this.k.a;
            if (this.k == ib2) {
                this.b = null;
                return null;
            }
            int n2 = 119 % ((by - -37) / 43);
            this.b = ib2.a;
            return ib2;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, z[6] + by + ')');
        }
    }

    final ib b(byte by) {
        try {
            ib ib2;
            block5: {
                block4: {
                    ++a;
                    int n2 = 81 % ((-37 - by) / 51);
                    ib2 = this.b;
                    if (this.k == ib2) break block4;
                    break block5;
                }
                this.b = null;
                return null;
            }
            this.b = ib2.a;
            return ib2;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, z[1] + by + ')');
        }
    }

    public db() {
        try {
            this.k.e = this.k;
            this.k.a = this.k;
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
    static final boolean a(int n2, char c2) {
        try {
            ++e;
            if (n2 != 32) {
                return false;
            }
            if (~c2 == -161) return true;
            if (c2 == ' ') return true;
            if (~c2 == -96) return true;
            if (~c2 == -46) return true;
            return false;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, z[0] + n2 + ',' + c2 + ')');
        }
    }

    static {
        z = new String[]{db.z(db.z("}\rZR?")), db.z(db.z("}\rZQ?")), db.z(db.z("bAZ>j")), db.z(db.z("w\u001a\u0018|")), db.z(db.z("}\rZS?")), db.z(db.z("}\rZ,~w\u0006\u0000.?0")), db.z(db.z("}\rZT?"))};
        g = 0;
    }

    private static char[] z(String string) {
        char[] cArray = string.toCharArray();
        if (cArray.length < 2) {
            cArray = cArray;
            cArray[0] = (char)(cArray[0] ^ 0x17);
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
                        n4 = 25;
                        break;
                    }
                    case 1: {
                        n4 = 111;
                        break;
                    }
                    case 2: {
                        n4 = 116;
                        break;
                    }
                    case 3: {
                        n4 = 16;
                        break;
                    }
                    default: {
                        n4 = 23;
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

