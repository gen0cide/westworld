/*
 * Decompiled with CFR 0.152.
 */
import java.awt.Image;
import java.awt.image.ImageConsumer;
import java.awt.image.ImageObserver;
import java.awt.image.ImageProducer;

final class fb
implements ImageProducer,
ImageObserver {
    static i h;
    static int g;
    static aa a;
    static int[] f;
    static int i;
    static int b;
    static int l;
    static int j;
    static int e;
    static int[] d;
    static int[] c;
    static boolean[] k;
    private static final String[] z;

    @Override
    public final synchronized void addConsumer(ImageConsumer imageConsumer) {
        try {
            u.d = imageConsumer;
            ++e;
            imageConsumer.setDimensions(k.o, da.bb);
            imageConsumer.setProperties(null);
            imageConsumer.setColorModel(m.d);
            imageConsumer.setHints(14);
        }
        catch (RuntimeException runtimeException) {
            RuntimeException runtimeException2 = runtimeException;
            StringBuilder stringBuilder = new StringBuilder().append(z[6]);
            String string = imageConsumer != null ? z[1] : z[0];
            throw i.a(runtimeException2, stringBuilder.append(string).append(')').toString());
        }
    }

    @Override
    public final void startProduction(ImageConsumer imageConsumer) {
        try {
            ++i;
            this.addConsumer(imageConsumer);
        }
        catch (RuntimeException runtimeException) {
            RuntimeException runtimeException2 = runtimeException;
            StringBuilder stringBuilder = new StringBuilder().append(z[7]);
            String string = imageConsumer != null ? z[1] : z[0];
            throw i.a(runtimeException2, stringBuilder.append(string).append(')').toString());
        }
    }

    @Override
    public final void requestTopDownLeftRightResend(ImageConsumer imageConsumer) {
        try {
            ++j;
        }
        catch (RuntimeException runtimeException) {
            RuntimeException runtimeException2 = runtimeException;
            StringBuilder stringBuilder = new StringBuilder().append(z[4]);
            String string = imageConsumer != null ? z[1] : z[0];
            throw i.a(runtimeException2, stringBuilder.append(string).append(')').toString());
        }
    }

    fb() {
    }

    @Override
    public final synchronized void removeConsumer(ImageConsumer imageConsumer) {
        try {
            ++g;
            if (u.d == imageConsumer) {
                u.d = null;
            }
        }
        catch (RuntimeException runtimeException) {
            RuntimeException runtimeException2 = runtimeException;
            StringBuilder stringBuilder = new StringBuilder().append(z[3]);
            String string = imageConsumer != null ? z[1] : z[0];
            throw i.a(runtimeException2, stringBuilder.append(string).append(')').toString());
        }
    }

    @Override
    public final synchronized boolean isConsumer(ImageConsumer imageConsumer) {
        try {
            ++b;
            return imageConsumer == u.d;
        }
        catch (RuntimeException runtimeException) {
            RuntimeException runtimeException2 = runtimeException;
            StringBuilder stringBuilder = new StringBuilder().append(z[5]);
            String string = imageConsumer != null ? z[1] : z[0];
            throw i.a(runtimeException2, stringBuilder.append(string).append(')').toString());
        }
    }

    @Override
    public final boolean imageUpdate(Image image, int n2, int n3, int n4, int n5, int n6) {
        try {
            ++l;
            return true;
        }
        catch (RuntimeException runtimeException) {
            RuntimeException runtimeException2 = runtimeException;
            StringBuilder stringBuilder = new StringBuilder().append(z[2]);
            String string = image != null ? z[1] : z[0];
            throw i.a(runtimeException2, stringBuilder.append(string).append(',').append(n2).append(',').append(n3).append(',').append(n4).append(',').append(n5).append(',').append(n6).append(')').toString());
        }
    }

    static {
        z = new String[]{fb.z(fb.z("1{3R")), fb.z(fb.z("$ q\u00107")), fb.z(fb.z("9lqW'>i:k:;o+[b")), fb.z(fb.z("9lqL/2a)[\t0`,K':|w")), fb.z(fb.z("9lqL/.{:M>\u000ba/z%(`\u0013[,+\\6Y\"+\\:M/1jw")), fb.z(fb.z("9lqW9\u001ca1M?2k-\u0016")), fb.z(fb.z("9lq_.;M0P9*c:Lb")), fb.z(fb.z("9lqM>>|+n80j*]>6a1\u0016"))};
        k = new boolean[]{false, false, false, false, false, false, false, false, false, false, false, false};
        h = new i(fb.z(fb.z("\rM")), 1);
    }

    private static char[] z(String string) {
        char[] cArray = string.toCharArray();
        if (cArray.length < 2) {
            cArray = cArray;
            cArray[0] = (char)(cArray[0] ^ 0x4A);
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
                        n4 = 95;
                        break;
                    }
                    case 1: {
                        n4 = 14;
                        break;
                    }
                    case 2: {
                        n4 = 95;
                        break;
                    }
                    case 3: {
                        n4 = 62;
                        break;
                    }
                    default: {
                        n4 = 74;
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

