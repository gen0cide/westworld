/*
 * Decompiled with CFR 0.152.
 */
import java.io.IOException;

final class fa
extends IOException {
    static int b;
    static String[] a;
    static int d;
    static int[] e;
    static int[] c;

    fa(String string) {
        super(string);
    }

    static {
        a = new String[]{fa.z(fa.z("F%:\u001f\u0014f4/ZZg1(\u001fF23,Z]f9'\t\u0014f3j\u0018Ak|+\u0014P2,8\u001fGa|/\u0014@w."))};
        d = 234;
    }

    private static char[] z(String string) {
        char[] cArray = string.toCharArray();
        if (cArray.length < 2) {
            cArray = cArray;
            cArray[0] = (char)(cArray[0] ^ 0x34);
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
                        n4 = 18;
                        break;
                    }
                    case 1: {
                        n4 = 92;
                        break;
                    }
                    case 2: {
                        n4 = 74;
                        break;
                    }
                    case 3: {
                        n4 = 122;
                        break;
                    }
                    default: {
                        n4 = 52;
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

