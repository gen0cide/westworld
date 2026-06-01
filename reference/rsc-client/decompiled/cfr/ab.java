/*
 * Decompiled with CFR 0.152.
 */
final class ab {
    static final void a(int[] nArray, int n2, int n3) {
        n3 = n2 + n3 - 7;
        while (n2 < n3) {
            nArray[n2++] = 0;
            nArray[n2++] = 0;
            nArray[n2++] = 0;
            nArray[n2++] = 0;
            nArray[n2++] = 0;
            nArray[n2++] = 0;
            nArray[n2++] = 0;
            nArray[n2++] = 0;
        }
        n3 += 7;
        while (n2 < n3) {
            nArray[n2++] = 0;
        }
    }

    static final void a(byte[] byArray, int n2, byte[] byArray2, int n3, int n4) {
        if (byArray == byArray2) {
            if (n2 == n3) {
                return;
            }
            if (n3 > n2 && n3 < n2 + n4) {
                n2 += --n4;
                n3 += n4;
                n4 = n2 - n4;
                n4 += 7;
                while (n2 >= n4) {
                    byArray2[n3--] = byArray[n2--];
                    byArray2[n3--] = byArray[n2--];
                    byArray2[n3--] = byArray[n2--];
                    byArray2[n3--] = byArray[n2--];
                    byArray2[n3--] = byArray[n2--];
                    byArray2[n3--] = byArray[n2--];
                    byArray2[n3--] = byArray[n2--];
                    byArray2[n3--] = byArray[n2--];
                }
                n4 -= 7;
                while (n2 >= n4) {
                    byArray2[n3--] = byArray[n2--];
                }
                return;
            }
        }
        n4 += n2;
        n4 -= 7;
        while (n2 < n4) {
            byArray2[n3++] = byArray[n2++];
            byArray2[n3++] = byArray[n2++];
            byArray2[n3++] = byArray[n2++];
            byArray2[n3++] = byArray[n2++];
            byArray2[n3++] = byArray[n2++];
            byArray2[n3++] = byArray[n2++];
            byArray2[n3++] = byArray[n2++];
            byArray2[n3++] = byArray[n2++];
        }
        n4 += 7;
        while (n2 < n4) {
            byArray2[n3++] = byArray[n2++];
        }
    }
}

