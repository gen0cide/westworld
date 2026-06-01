/*
 * Decompiled with CFR 0.152.
 */
abstract class va
extends ib {
    va j;
    bb h;
    int i;
    volatile boolean g = true;

    abstract void b(int var1);

    abstract int d();

    abstract va a();

    abstract void b(int[] var1, int var2, int var3);

    abstract va b();

    int c() {
        return 255;
    }

    final void a(int[] nArray, int n2, int n3) {
        if (this.g) {
            this.b(nArray, n2, n3);
        } else {
            this.b(n3);
        }
    }

    va() {
    }
}

