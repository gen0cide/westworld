/*
 * Decompiled with CFR 0.152.
 */
final class sb
extends va {
    private int t;
    private int x;
    private int o;
    private int p;
    private int n;
    private int r;
    private int y;
    private int s;
    private int v;
    private int l;
    private int m;
    private int w;
    private int k;
    private int q;
    private boolean u;

    @Override
    final int c() {
        int n2 = this.w * 3 >> 6;
        n2 = (n2 ^ n2 >> 31) + (n2 >>> 31);
        if (this.l == 0) {
            n2 -= n2 * this.v / (((vb)this.h).l.length << 8);
        } else if (this.l >= 0) {
            n2 -= n2 * this.x / ((vb)this.h).l.length;
        }
        return n2 > 255 ? 255 : n2;
    }

    private static final int b(byte[] byArray, int[] nArray, int n2, int n3, int n4, int n5, int n6, int n7, int n8, sb sb2) {
        n4 <<= 2;
        n5 <<= 2;
        n6 = n3 + (n2 >>= 8) - ((n8 >>= 8) - 1);
        if (n6 > n7) {
            n6 = n7;
        }
        sb2.k += sb2.p * (n6 - n3);
        sb2.m += sb2.o * (n6 - n3);
        n6 -= 3;
        while (n3 < n6) {
            int n9 = n3++;
            nArray[n9] = nArray[n9] + byArray[n2--] * n4;
            int n10 = n3++;
            nArray[n10] = nArray[n10] + byArray[n2--] * (n4 += n5);
            int n11 = n3++;
            nArray[n11] = nArray[n11] + byArray[n2--] * (n4 += n5);
            int n12 = n3++;
            nArray[n12] = nArray[n12] + byArray[n2--] * (n4 += n5);
            n4 += n5;
        }
        n6 += 3;
        while (n3 < n6) {
            int n13 = n3++;
            nArray[n13] = nArray[n13] + byArray[n2--] * n4;
            n4 += n5;
        }
        sb2.w = n4 >> 2;
        sb2.v = n2 << 8;
        return n3;
    }

    /*
     * Exception decompiling
     */
    @Override
    final synchronized void b(int[] var1_1, int var2_2, int var3_3) {
        /*
         * This method has failed to decompile.  When submitting a bug report, please provide this stack trace, and (if you hold appropriate legal rights) the relevant class file.
         * 
         * org.benf.cfr.reader.util.ConfusedCFRException: Tried to end blocks [3[DOLOOP]], but top level block is 33[SIMPLE_IF_TAKEN]
         *     at org.benf.cfr.reader.bytecode.analysis.opgraph.Op04StructuredStatement.processEndingBlocks(Op04StructuredStatement.java:435)
         *     at org.benf.cfr.reader.bytecode.analysis.opgraph.Op04StructuredStatement.buildNestedBlocks(Op04StructuredStatement.java:484)
         *     at org.benf.cfr.reader.bytecode.analysis.opgraph.Op03SimpleStatement.createInitialStructuredBlock(Op03SimpleStatement.java:736)
         *     at org.benf.cfr.reader.bytecode.CodeAnalyser.getAnalysisInner(CodeAnalyser.java:850)
         *     at org.benf.cfr.reader.bytecode.CodeAnalyser.getAnalysisOrWrapFail(CodeAnalyser.java:278)
         *     at org.benf.cfr.reader.bytecode.CodeAnalyser.getAnalysis(CodeAnalyser.java:201)
         *     at org.benf.cfr.reader.entities.attributes.AttributeCode.analyse(AttributeCode.java:94)
         *     at org.benf.cfr.reader.entities.Method.analyse(Method.java:531)
         *     at org.benf.cfr.reader.entities.ClassFile.analyseMid(ClassFile.java:1055)
         *     at org.benf.cfr.reader.entities.ClassFile.analyseTop(ClassFile.java:942)
         *     at org.benf.cfr.reader.Driver.doJarVersionTypes(Driver.java:257)
         *     at org.benf.cfr.reader.Driver.doJar(Driver.java:139)
         *     at org.benf.cfr.reader.CfrDriverImpl.analyse(CfrDriverImpl.java:76)
         *     at org.benf.cfr.reader.Main.main(Main.java:54)
         */
        throw new IllegalStateException("Decompilation failed");
    }

    static final sb a(vb vb2, int n2, int n3) {
        if (vb2.l == null || vb2.l.length == 0) {
            return null;
        }
        return new sb(vb2, (int)((long)vb2.i * 256L * (long)n2 / (long)(100 * sa.t)), n3 << 6);
    }

    private static final int a(int n2, int n3, byte[] byArray, int[] nArray, int n4, int n5, int n6, int n7, int n8, int n9, int n10, sb sb2, int n11, int n12) {
        if (n11 == 0 || (n8 = n5 + (n10 - n4 + n11 - 257) / n11) > n9) {
            n8 = n9;
        }
        n5 <<= 1;
        n8 <<= 1;
        while (n5 < n8) {
            n3 = n4 >> 8;
            n2 = byArray[n3];
            n2 = (n2 << 8) + (byArray[n3 + 1] - n2) * (n4 & 0xFF);
            int n13 = n5++;
            nArray[n13] = nArray[n13] + (n2 * n6 >> 6);
            int n14 = n5++;
            nArray[n14] = nArray[n14] + (n2 * n7 >> 6);
            n4 += n11;
        }
        if (n11 == 0 || (n8 = (n5 >> 1) + (n10 - n4 + n11 - 1) / n11) > n9) {
            n8 = n9;
        }
        n8 <<= 1;
        n3 = n12;
        while (n5 < n8) {
            n2 = byArray[n4 >> 8];
            n2 = (n2 << 8) + (n3 - n2) * (n4 & 0xFF);
            int n15 = n5++;
            nArray[n15] = nArray[n15] + (n2 * n6 >> 6);
            int n16 = n5++;
            nArray[n16] = nArray[n16] + (n2 * n7 >> 6);
            n4 += n11;
        }
        sb2.v = n4;
        return n5 >> 1;
    }

    private final boolean f() {
        int n2;
        int n3;
        int n4 = this.r;
        if (n4 == Integer.MIN_VALUE) {
            n3 = 0;
            n2 = 0;
            n4 = 0;
        } else {
            n2 = sb.b(n4, this.q);
            n3 = sb.c(n4, this.q);
        }
        if (this.w != n4 || this.k != n2 || this.m != n3) {
            if (this.w < n4) {
                this.s = 1;
                this.y = n4 - this.w;
            } else if (this.w > n4) {
                this.s = -1;
                this.y = this.w - n4;
            } else {
                this.s = 0;
            }
            if (this.k < n2) {
                this.p = 1;
                if (this.y == 0 || this.y > n2 - this.k) {
                    this.y = n2 - this.k;
                }
            } else if (this.k > n2) {
                this.p = -1;
                if (this.y == 0 || this.y > this.k - n2) {
                    this.y = this.k - n2;
                }
            } else {
                this.p = 0;
            }
            if (this.m < n3) {
                this.o = 1;
                if (this.y == 0 || this.y > n3 - this.m) {
                    this.y = n3 - this.m;
                }
            } else if (this.m > n3) {
                this.o = -1;
                if (this.y == 0 || this.y > this.m - n3) {
                    this.y = this.m - n3;
                }
            } else {
                this.o = 0;
            }
            return false;
        }
        if (this.r == Integer.MIN_VALUE) {
            this.r = 0;
            this.m = 0;
            this.k = 0;
            this.w = 0;
            this.a(-27331);
            return true;
        }
        this.e();
        return false;
    }

    private static final int a(int n2, byte[] byArray, int[] nArray, int n3, int n4, int n5, int n6, int n7, int n8, int n9, sb sb2) {
        n5 <<= 2;
        n6 <<= 2;
        n7 = n4 + (n9 >>= 8) - (n3 >>= 8);
        if (n7 > n8) {
            n7 = n8;
        }
        n4 <<= 1;
        n7 <<= 1;
        n7 -= 6;
        while (n4 < n7) {
            n2 = byArray[n3++];
            int n10 = n4++;
            nArray[n10] = nArray[n10] + n2 * n5;
            int n11 = n4++;
            nArray[n11] = nArray[n11] + n2 * n6;
            n2 = byArray[n3++];
            int n12 = n4++;
            nArray[n12] = nArray[n12] + n2 * n5;
            int n13 = n4++;
            nArray[n13] = nArray[n13] + n2 * n6;
            n2 = byArray[n3++];
            int n14 = n4++;
            nArray[n14] = nArray[n14] + n2 * n5;
            int n15 = n4++;
            nArray[n15] = nArray[n15] + n2 * n6;
            n2 = byArray[n3++];
            int n16 = n4++;
            nArray[n16] = nArray[n16] + n2 * n5;
            int n17 = n4++;
            nArray[n17] = nArray[n17] + n2 * n6;
        }
        n7 += 6;
        while (n4 < n7) {
            n2 = byArray[n3++];
            int n18 = n4++;
            nArray[n18] = nArray[n18] + n2 * n5;
            int n19 = n4++;
            nArray[n19] = nArray[n19] + n2 * n6;
        }
        sb2.v = n3 << 8;
        return n4 >> 1;
    }

    private final int b(int[] nArray, int n2, int n3, int n4, int n5) {
        while (this.y > 0) {
            int n6 = n2 + this.y;
            if (n6 > n4) {
                n6 = n4;
            }
            this.y += n2;
            n2 = this.n == -256 && (this.v & 0xFF) == 0 ? (sa.i ? sb.b(0, ((vb)this.h).l, nArray, this.v, n2, this.k, this.m, this.p, this.o, 0, n6, n3, this) : sb.b(((vb)this.h).l, nArray, this.v, n2, this.w, this.s, 0, n6, n3, this)) : (sa.i ? sb.b(0, 0, ((vb)this.h).l, nArray, this.v, n2, this.k, this.m, this.p, this.o, 0, n6, n3, this, this.n, n5) : sb.c(0, 0, ((vb)this.h).l, nArray, this.v, n2, this.w, this.s, 0, n6, n3, this, this.n, n5));
            this.y -= n2;
            if (this.y != 0) {
                return n2;
            }
            if (!this.f()) continue;
            return n4;
        }
        if (this.n == -256 && (this.v & 0xFF) == 0) {
            if (sa.i) {
                return sb.b(0, ((vb)this.h).l, nArray, this.v, n2, this.k, this.m, 0, n4, n3, this);
            }
            return sb.a(((vb)this.h).l, nArray, this.v, n2, this.w, 0, n4, n3, this);
        }
        if (sa.i) {
            return sb.d(0, 0, ((vb)this.h).l, nArray, this.v, n2, this.k, this.m, 0, n4, n3, this, this.n, n5);
        }
        return sb.b(0, 0, ((vb)this.h).l, nArray, this.v, n2, this.w, 0, n4, n3, this, this.n, n5);
    }

    private static final int c(int n2, int n3, byte[] byArray, int[] nArray, int n4, int n5, int n6, int n7, int n8, int n9, int n10, sb sb2, int n11, int n12) {
        sb2.k -= sb2.p * n5;
        sb2.m -= sb2.o * n5;
        if (n11 == 0 || (n8 = n5 + (n10 + 256 - n4 + n11) / n11) > n9) {
            n8 = n9;
        }
        while (n5 < n8) {
            n3 = n4 >> 8;
            n2 = byArray[n3 - 1];
            int n13 = n5++;
            nArray[n13] = nArray[n13] + (((n2 << 8) + (byArray[n3] - n2) * (n4 & 0xFF)) * n6 >> 6);
            n6 += n7;
            n4 += n11;
        }
        if (n11 == 0 || (n8 = n5 + (n10 - n4 + n11) / n11) > n9) {
            n8 = n9;
        }
        n2 = n12;
        n3 = n11;
        while (n5 < n8) {
            int n14 = n5++;
            nArray[n14] = nArray[n14] + (((n2 << 8) + (byArray[n4 >> 8] - n2) * (n4 & 0xFF)) * n6 >> 6);
            n6 += n7;
            n4 += n3;
        }
        sb2.k += sb2.p * n5;
        sb2.m += sb2.o * n5;
        sb2.w = n6;
        sb2.v = n4;
        return n5;
    }

    private static final int d(int n2, int n3, byte[] byArray, int[] nArray, int n4, int n5, int n6, int n7, int n8, int n9, int n10, sb sb2, int n11, int n12) {
        if (n11 == 0 || (n8 = n5 + (n10 + 256 - n4 + n11) / n11) > n9) {
            n8 = n9;
        }
        n5 <<= 1;
        n8 <<= 1;
        while (n5 < n8) {
            n3 = n4 >> 8;
            n2 = byArray[n3 - 1];
            n2 = (n2 << 8) + (byArray[n3] - n2) * (n4 & 0xFF);
            int n13 = n5++;
            nArray[n13] = nArray[n13] + (n2 * n6 >> 6);
            int n14 = n5++;
            nArray[n14] = nArray[n14] + (n2 * n7 >> 6);
            n4 += n11;
        }
        if (n11 == 0 || (n8 = (n5 >> 1) + (n10 - n4 + n11) / n11) > n9) {
            n8 = n9;
        }
        n8 <<= 1;
        n3 = n12;
        while (n5 < n8) {
            n2 = (n3 << 8) + (byArray[n4 >> 8] - n3) * (n4 & 0xFF);
            int n15 = n5++;
            nArray[n15] = nArray[n15] + (n2 * n6 >> 6);
            int n16 = n5++;
            nArray[n16] = nArray[n16] + (n2 * n7 >> 6);
            n4 += n11;
        }
        sb2.v = n4;
        return n5 >> 1;
    }

    private final void e() {
        this.w = this.r;
        this.k = sb.b(this.r, this.q);
        this.m = sb.c(this.r, this.q);
    }

    private static final int a(int n2, int n3, byte[] byArray, int[] nArray, int n4, int n5, int n6, int n7, int n8, int n9, int n10, int n11, int n12, sb sb2, int n13, int n14) {
        sb2.w -= sb2.s * n5;
        if (n13 == 0 || (n10 = n5 + (n12 - n4 + n13 - 257) / n13) > n11) {
            n10 = n11;
        }
        n5 <<= 1;
        n10 <<= 1;
        while (n5 < n10) {
            n3 = n4 >> 8;
            n2 = byArray[n3];
            n2 = (n2 << 8) + (byArray[n3 + 1] - n2) * (n4 & 0xFF);
            int n15 = n5++;
            nArray[n15] = nArray[n15] + (n2 * n6 >> 6);
            n6 += n8;
            int n16 = n5++;
            nArray[n16] = nArray[n16] + (n2 * n7 >> 6);
            n7 += n9;
            n4 += n13;
        }
        if (n13 == 0 || (n10 = (n5 >> 1) + (n12 - n4 + n13 - 1) / n13) > n11) {
            n10 = n11;
        }
        n10 <<= 1;
        n3 = n14;
        while (n5 < n10) {
            n2 = byArray[n4 >> 8];
            n2 = (n2 << 8) + (n3 - n2) * (n4 & 0xFF);
            int n17 = n5++;
            nArray[n17] = nArray[n17] + (n2 * n6 >> 6);
            n6 += n8;
            int n18 = n5++;
            nArray[n18] = nArray[n18] + (n2 * n7 >> 6);
            n7 += n9;
            n4 += n13;
        }
        sb2.w += sb2.s * (n5 >>= 1);
        sb2.k = n6;
        sb2.m = n7;
        sb2.v = n4;
        return n5;
    }

    @Override
    final va b() {
        return null;
    }

    private static final int a(byte[] byArray, int[] nArray, int n2, int n3, int n4, int n5, int n6, int n7, sb sb2) {
        n4 <<= 2;
        n5 = n3 + (n2 >>= 8) - ((n7 >>= 8) - 1);
        if (n5 > n6) {
            n5 = n6;
        }
        n5 -= 3;
        while (n3 < n5) {
            int n8 = n3++;
            nArray[n8] = nArray[n8] + byArray[n2--] * n4;
            int n9 = n3++;
            nArray[n9] = nArray[n9] + byArray[n2--] * n4;
            int n10 = n3++;
            nArray[n10] = nArray[n10] + byArray[n2--] * n4;
            int n11 = n3++;
            nArray[n11] = nArray[n11] + byArray[n2--] * n4;
        }
        n5 += 3;
        while (n3 < n5) {
            int n12 = n3++;
            nArray[n12] = nArray[n12] + byArray[n2--] * n4;
        }
        sb2.v = n2 << 8;
        return n3;
    }

    private static final int a(int n2, byte[] byArray, int[] nArray, int n3, int n4, int n5, int n6, int n7, int n8, int n9, int n10, int n11, sb sb2) {
        n5 <<= 2;
        n6 <<= 2;
        n7 <<= 2;
        n8 <<= 2;
        n9 = n4 + (n11 >>= 8) - (n3 >>= 8);
        if (n9 > n10) {
            n9 = n10;
        }
        sb2.w += sb2.s * (n9 - n4);
        n4 <<= 1;
        n9 <<= 1;
        n9 -= 6;
        while (n4 < n9) {
            n2 = byArray[n3++];
            int n12 = n4++;
            nArray[n12] = nArray[n12] + n2 * n5;
            n5 += n7;
            int n13 = n4++;
            nArray[n13] = nArray[n13] + n2 * n6;
            n6 += n8;
            n2 = byArray[n3++];
            int n14 = n4++;
            nArray[n14] = nArray[n14] + n2 * n5;
            n5 += n7;
            int n15 = n4++;
            nArray[n15] = nArray[n15] + n2 * n6;
            n6 += n8;
            n2 = byArray[n3++];
            int n16 = n4++;
            nArray[n16] = nArray[n16] + n2 * n5;
            n5 += n7;
            int n17 = n4++;
            nArray[n17] = nArray[n17] + n2 * n6;
            n6 += n8;
            n2 = byArray[n3++];
            int n18 = n4++;
            nArray[n18] = nArray[n18] + n2 * n5;
            n5 += n7;
            int n19 = n4++;
            nArray[n19] = nArray[n19] + n2 * n6;
            n6 += n8;
        }
        n9 += 6;
        while (n4 < n9) {
            n2 = byArray[n3++];
            int n20 = n4++;
            nArray[n20] = nArray[n20] + n2 * n5;
            n5 += n7;
            int n21 = n4++;
            nArray[n21] = nArray[n21] + n2 * n6;
            n6 += n8;
        }
        sb2.k = n5 >> 2;
        sb2.m = n6 >> 2;
        sb2.v = n3 << 8;
        return n4 >> 1;
    }

    private static final int b(int n2, byte[] byArray, int[] nArray, int n3, int n4, int n5, int n6, int n7, int n8, int n9, sb sb2) {
        n5 <<= 2;
        n6 <<= 2;
        n7 = n4 + (n3 >>= 8) - ((n9 >>= 8) - 1);
        if (n7 > n8) {
            n7 = n8;
        }
        n4 <<= 1;
        n7 <<= 1;
        n7 -= 6;
        while (n4 < n7) {
            n2 = byArray[n3--];
            int n10 = n4++;
            nArray[n10] = nArray[n10] + n2 * n5;
            int n11 = n4++;
            nArray[n11] = nArray[n11] + n2 * n6;
            n2 = byArray[n3--];
            int n12 = n4++;
            nArray[n12] = nArray[n12] + n2 * n5;
            int n13 = n4++;
            nArray[n13] = nArray[n13] + n2 * n6;
            n2 = byArray[n3--];
            int n14 = n4++;
            nArray[n14] = nArray[n14] + n2 * n5;
            int n15 = n4++;
            nArray[n15] = nArray[n15] + n2 * n6;
            n2 = byArray[n3--];
            int n16 = n4++;
            nArray[n16] = nArray[n16] + n2 * n5;
            int n17 = n4++;
            nArray[n17] = nArray[n17] + n2 * n6;
        }
        n7 += 6;
        while (n4 < n7) {
            n2 = byArray[n3--];
            int n18 = n4++;
            nArray[n18] = nArray[n18] + n2 * n5;
            int n19 = n4++;
            nArray[n19] = nArray[n19] + n2 * n6;
        }
        sb2.v = n3 << 8;
        return n4 >> 1;
    }

    private static final int b(byte[] byArray, int[] nArray, int n2, int n3, int n4, int n5, int n6, int n7, sb sb2) {
        n4 <<= 2;
        n5 = n3 + (n7 >>= 8) - (n2 >>= 8);
        if (n5 > n6) {
            n5 = n6;
        }
        n5 -= 3;
        while (n3 < n5) {
            int n8 = n3++;
            nArray[n8] = nArray[n8] + byArray[n2++] * n4;
            int n9 = n3++;
            nArray[n9] = nArray[n9] + byArray[n2++] * n4;
            int n10 = n3++;
            nArray[n10] = nArray[n10] + byArray[n2++] * n4;
            int n11 = n3++;
            nArray[n11] = nArray[n11] + byArray[n2++] * n4;
        }
        n5 += 3;
        while (n3 < n5) {
            int n12 = n3++;
            nArray[n12] = nArray[n12] + byArray[n2++] * n4;
        }
        sb2.v = n2 << 8;
        return n3;
    }

    private static final int b(int n2, int n3, byte[] byArray, int[] nArray, int n4, int n5, int n6, int n7, int n8, int n9, int n10, int n11, int n12, sb sb2, int n13, int n14) {
        sb2.w -= sb2.s * n5;
        if (n13 == 0 || (n10 = n5 + (n12 + 256 - n4 + n13) / n13) > n11) {
            n10 = n11;
        }
        n5 <<= 1;
        n10 <<= 1;
        while (n5 < n10) {
            n3 = n4 >> 8;
            n2 = byArray[n3 - 1];
            n2 = (n2 << 8) + (byArray[n3] - n2) * (n4 & 0xFF);
            int n15 = n5++;
            nArray[n15] = nArray[n15] + (n2 * n6 >> 6);
            n6 += n8;
            int n16 = n5++;
            nArray[n16] = nArray[n16] + (n2 * n7 >> 6);
            n7 += n9;
            n4 += n13;
        }
        if (n13 == 0 || (n10 = (n5 >> 1) + (n12 - n4 + n13) / n13) > n11) {
            n10 = n11;
        }
        n10 <<= 1;
        n3 = n14;
        while (n5 < n10) {
            n2 = (n3 << 8) + (byArray[n4 >> 8] - n3) * (n4 & 0xFF);
            int n17 = n5++;
            nArray[n17] = nArray[n17] + (n2 * n6 >> 6);
            n6 += n8;
            int n18 = n5++;
            nArray[n18] = nArray[n18] + (n2 * n7 >> 6);
            n7 += n9;
            n4 += n13;
        }
        sb2.w += sb2.s * (n5 >>= 1);
        sb2.k = n6;
        sb2.m = n7;
        sb2.v = n4;
        return n5;
    }

    private static final int b(int n2, byte[] byArray, int[] nArray, int n3, int n4, int n5, int n6, int n7, int n8, int n9, int n10, int n11, sb sb2) {
        n5 <<= 2;
        n6 <<= 2;
        n7 <<= 2;
        n8 <<= 2;
        n9 = n4 + (n3 >>= 8) - ((n11 >>= 8) - 1);
        if (n9 > n10) {
            n9 = n10;
        }
        sb2.w += sb2.s * (n9 - n4);
        n4 <<= 1;
        n9 <<= 1;
        n9 -= 6;
        while (n4 < n9) {
            n2 = byArray[n3--];
            int n12 = n4++;
            nArray[n12] = nArray[n12] + n2 * n5;
            n5 += n7;
            int n13 = n4++;
            nArray[n13] = nArray[n13] + n2 * n6;
            n6 += n8;
            n2 = byArray[n3--];
            int n14 = n4++;
            nArray[n14] = nArray[n14] + n2 * n5;
            n5 += n7;
            int n15 = n4++;
            nArray[n15] = nArray[n15] + n2 * n6;
            n6 += n8;
            n2 = byArray[n3--];
            int n16 = n4++;
            nArray[n16] = nArray[n16] + n2 * n5;
            n5 += n7;
            int n17 = n4++;
            nArray[n17] = nArray[n17] + n2 * n6;
            n6 += n8;
            n2 = byArray[n3--];
            int n18 = n4++;
            nArray[n18] = nArray[n18] + n2 * n5;
            n5 += n7;
            int n19 = n4++;
            nArray[n19] = nArray[n19] + n2 * n6;
            n6 += n8;
        }
        n9 += 6;
        while (n4 < n9) {
            n2 = byArray[n3--];
            int n20 = n4++;
            nArray[n20] = nArray[n20] + n2 * n5;
            n5 += n7;
            int n21 = n4++;
            nArray[n21] = nArray[n21] + n2 * n6;
            n6 += n8;
        }
        sb2.k = n5 >> 2;
        sb2.m = n6 >> 2;
        sb2.v = n3 << 8;
        return n4 >> 1;
    }

    private final void g() {
        if (this.y != 0) {
            if (this.r == Integer.MIN_VALUE) {
                this.r = 0;
            }
            this.y = 0;
            this.e();
        }
    }

    private static final int b(int n2, int n3, byte[] byArray, int[] nArray, int n4, int n5, int n6, int n7, int n8, int n9, sb sb2, int n10, int n11) {
        if (n10 == 0 || (n7 = n5 + (n9 + 256 - n4 + n10) / n10) > n8) {
            n7 = n8;
        }
        while (n5 < n7) {
            n3 = n4 >> 8;
            n2 = byArray[n3 - 1];
            int n12 = n5++;
            nArray[n12] = nArray[n12] + (((n2 << 8) + (byArray[n3] - n2) * (n4 & 0xFF)) * n6 >> 6);
            n4 += n10;
        }
        if (n10 == 0 || (n7 = n5 + (n9 - n4 + n10) / n10) > n8) {
            n7 = n8;
        }
        n2 = n11;
        n3 = n10;
        while (n5 < n7) {
            int n13 = n5++;
            nArray[n13] = nArray[n13] + (((n2 << 8) + (byArray[n4 >> 8] - n2) * (n4 & 0xFF)) * n6 >> 6);
            n4 += n3;
        }
        sb2.v = n4;
        return n5;
    }

    private static final int c(int n2, int n3) {
        return n3 < 0 ? -n2 : (int)((double)n2 * Math.sqrt((double)n3 * 1.220703125E-4) + 0.5);
    }

    private static final int a(byte[] byArray, int[] nArray, int n2, int n3, int n4, int n5, int n6, int n7, int n8, sb sb2) {
        n4 <<= 2;
        n5 <<= 2;
        n6 = n3 + (n8 >>= 8) - (n2 >>= 8);
        if (n6 > n7) {
            n6 = n7;
        }
        sb2.k += sb2.p * (n6 - n3);
        sb2.m += sb2.o * (n6 - n3);
        n6 -= 3;
        while (n3 < n6) {
            int n9 = n3++;
            nArray[n9] = nArray[n9] + byArray[n2++] * n4;
            int n10 = n3++;
            nArray[n10] = nArray[n10] + byArray[n2++] * (n4 += n5);
            int n11 = n3++;
            nArray[n11] = nArray[n11] + byArray[n2++] * (n4 += n5);
            int n12 = n3++;
            nArray[n12] = nArray[n12] + byArray[n2++] * (n4 += n5);
            n4 += n5;
        }
        n6 += 3;
        while (n3 < n6) {
            int n13 = n3++;
            nArray[n13] = nArray[n13] + byArray[n2++] * n4;
            n4 += n5;
        }
        sb2.w = n4 >> 2;
        sb2.v = n2 << 8;
        return n3;
    }

    private static final int a(int n2, int n3, byte[] byArray, int[] nArray, int n4, int n5, int n6, int n7, int n8, int n9, sb sb2, int n10, int n11) {
        if (n10 == 0 || (n7 = n5 + (n9 - n4 + n10 - 257) / n10) > n8) {
            n7 = n8;
        }
        while (n5 < n7) {
            n3 = n4 >> 8;
            n2 = byArray[n3];
            int n12 = n5++;
            nArray[n12] = nArray[n12] + (((n2 << 8) + (byArray[n3 + 1] - n2) * (n4 & 0xFF)) * n6 >> 6);
            n4 += n10;
        }
        if (n10 == 0 || (n7 = n5 + (n9 - n4 + n10 - 1) / n10) > n8) {
            n7 = n8;
        }
        n3 = n11;
        while (n5 < n7) {
            n2 = byArray[n4 >> 8];
            int n13 = n5++;
            nArray[n13] = nArray[n13] + (((n2 << 8) + (n3 - n2) * (n4 & 0xFF)) * n6 >> 6);
            n4 += n10;
        }
        sb2.v = n4;
        return n5;
    }

    @Override
    final va a() {
        return null;
    }

    /*
     * Enabled aggressive block sorting
     */
    @Override
    final synchronized void b(int n2) {
        int n3;
        block32: {
            int n4;
            int n5;
            block34: {
                block31: {
                    int n6;
                    block33: {
                        if (this.y > 0) {
                            if (n2 >= this.y) {
                                if (this.r == Integer.MIN_VALUE) {
                                    this.r = 0;
                                    this.m = 0;
                                    this.k = 0;
                                    this.w = 0;
                                    this.a(-27331);
                                    n2 = this.y;
                                }
                                this.y = 0;
                                this.e();
                            } else {
                                this.w += this.s * n2;
                                this.k += this.p * n2;
                                this.m += this.o * n2;
                                this.y -= n2;
                            }
                        }
                        vb vb2 = (vb)this.h;
                        n5 = this.x << 8;
                        n4 = this.t << 8;
                        n3 = vb2.l.length << 8;
                        n6 = n4 - n5;
                        if (n6 <= 0) {
                            this.l = 0;
                        }
                        if (this.v < 0) {
                            if (this.n <= 0) {
                                this.g();
                                this.a(-27331);
                                return;
                            }
                            this.v = 0;
                        }
                        if (this.v >= n3) {
                            if (this.n >= 0) {
                                this.g();
                                this.a(-27331);
                                return;
                            }
                            this.v = n3 - 1;
                        }
                        this.v += this.n * n2;
                        if (this.l < 0) {
                            if (this.u) {
                                if (this.n < 0) {
                                    if (this.v >= n5) {
                                        return;
                                    }
                                    this.v = n5 + n5 - 1 - this.v;
                                    this.n = -this.n;
                                }
                                break block31;
                            } else {
                                if (this.n < 0) {
                                    if (this.v >= n5) {
                                        return;
                                    }
                                    this.v = n4 - 1 - (n4 - 1 - this.v) % n6;
                                    return;
                                }
                                if (this.v < n4) {
                                    return;
                                }
                                this.v = n5 + (this.v - n5) % n6;
                                return;
                            }
                        }
                        if (this.l <= 0) break block32;
                        if (!this.u) break block33;
                        if (this.n >= 0) break block34;
                        if (this.v >= n5) {
                            return;
                        }
                        this.v = n5 + n5 - 1 - this.v;
                        this.n = -this.n;
                        if (--this.l != 0) break block34;
                        break block32;
                    }
                    if (this.n < 0) {
                        if (this.v >= n5) {
                            return;
                        }
                        int n7 = (n4 - 1 - this.v) / n6;
                        if (n7 < this.l) {
                            this.v += n6 * n7;
                            this.l -= n7;
                            return;
                        }
                        this.v += n6 * this.l;
                        this.l = 0;
                        break block32;
                    } else {
                        if (this.v < n4) {
                            return;
                        }
                        int n8 = (this.v - n5) / n6;
                        if (n8 < this.l) {
                            this.v -= n6 * n8;
                            this.l -= n8;
                            return;
                        }
                        this.v -= n6 * this.l;
                        this.l = 0;
                    }
                    break block32;
                }
                while (true) {
                    if (this.v < n4) {
                        return;
                    }
                    this.v = n4 + n4 - 1 - this.v;
                    this.n = -this.n;
                    if (this.v >= n5) {
                        return;
                    }
                    this.v = n5 + n5 - 1 - this.v;
                    this.n = -this.n;
                }
            }
            do {
                if (this.v < n4) {
                    return;
                }
                this.v = n4 + n4 - 1 - this.v;
                this.n = -this.n;
                if (--this.l == 0) break;
                if (this.v >= n5) {
                    return;
                }
                this.v = n5 + n5 - 1 - this.v;
                this.n = -this.n;
            } while (--this.l != 0);
        }
        if (this.n < 0) {
            if (this.v >= 0) return;
            this.v = -1;
            this.g();
            this.a(-27331);
            return;
        }
        if (this.v < n3) return;
        this.v = n3;
        this.g();
        this.a(-27331);
    }

    private static final int b(int n2, int n3) {
        return n3 < 0 ? n2 : (int)((double)n2 * Math.sqrt((double)(16384 - n3) * 1.220703125E-4) + 0.5);
    }

    private final int a(int[] nArray, int n2, int n3, int n4, int n5) {
        while (this.y > 0) {
            int n6 = n2 + this.y;
            if (n6 > n4) {
                n6 = n4;
            }
            this.y += n2;
            n2 = this.n == 256 && (this.v & 0xFF) == 0 ? (sa.i ? sb.a(0, ((vb)this.h).l, nArray, this.v, n2, this.k, this.m, this.p, this.o, 0, n6, n3, this) : sb.a(((vb)this.h).l, nArray, this.v, n2, this.w, this.s, 0, n6, n3, this)) : (sa.i ? sb.a(0, 0, ((vb)this.h).l, nArray, this.v, n2, this.k, this.m, this.p, this.o, 0, n6, n3, this, this.n, n5) : sb.b(0, 0, ((vb)this.h).l, nArray, this.v, n2, this.w, this.s, 0, n6, n3, this, this.n, n5));
            this.y -= n2;
            if (this.y != 0) {
                return n2;
            }
            if (!this.f()) continue;
            return n4;
        }
        if (this.n == 256 && (this.v & 0xFF) == 0) {
            if (sa.i) {
                return sb.a(0, ((vb)this.h).l, nArray, this.v, n2, this.k, this.m, 0, n4, n3, this);
            }
            return sb.b(((vb)this.h).l, nArray, this.v, n2, this.w, 0, n4, n3, this);
        }
        if (sa.i) {
            return sb.a(0, 0, ((vb)this.h).l, nArray, this.v, n2, this.k, this.m, 0, n4, n3, this, this.n, n5);
        }
        return sb.a(0, 0, ((vb)this.h).l, nArray, this.v, n2, this.w, 0, n4, n3, this, this.n, n5);
    }

    @Override
    final int d() {
        if (this.r == 0 && this.y == 0) {
            return 0;
        }
        return 1;
    }

    private sb(vb vb2, int n2, int n3) {
        this.h = vb2;
        this.x = vb2.h;
        this.t = vb2.k;
        this.u = vb2.j;
        this.n = n2;
        this.r = n3;
        this.q = 8192;
        this.v = 0;
        this.e();
    }

    private static final int b(int n2, int n3, byte[] byArray, int[] nArray, int n4, int n5, int n6, int n7, int n8, int n9, int n10, sb sb2, int n11, int n12) {
        sb2.k -= sb2.p * n5;
        sb2.m -= sb2.o * n5;
        if (n11 == 0 || (n8 = n5 + (n10 - n4 + n11 - 257) / n11) > n9) {
            n8 = n9;
        }
        while (n5 < n8) {
            n3 = n4 >> 8;
            n2 = byArray[n3];
            int n13 = n5++;
            nArray[n13] = nArray[n13] + (((n2 << 8) + (byArray[n3 + 1] - n2) * (n4 & 0xFF)) * n6 >> 6);
            n6 += n7;
            n4 += n11;
        }
        if (n11 == 0 || (n8 = n5 + (n10 - n4 + n11 - 1) / n11) > n9) {
            n8 = n9;
        }
        n3 = n12;
        while (n5 < n8) {
            n2 = byArray[n4 >> 8];
            int n14 = n5++;
            nArray[n14] = nArray[n14] + (((n2 << 8) + (n3 - n2) * (n4 & 0xFF)) * n6 >> 6);
            n6 += n7;
            n4 += n11;
        }
        sb2.k += sb2.p * n5;
        sb2.m += sb2.o * n5;
        sb2.w = n6;
        sb2.v = n4;
        return n5;
    }
}

