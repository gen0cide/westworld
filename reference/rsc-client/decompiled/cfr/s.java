/*
 * Decompiled with CFR 0.152.
 */
import java.awt.Color;
import java.awt.Font;
import java.awt.FontMetrics;
import java.awt.Graphics;
import java.awt.Image;
import java.awt.image.PixelGrabber;
import java.math.BigInteger;

final class s {
    static int b;
    static String[] e;
    static String[] f;
    static BigInteger c;
    static nb a;
    static int d;
    private static final String[] z;

    /*
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    static final boolean a(int n2, Font font, int n3, int n4, e e2, char c2, FontMetrics fontMetrics, boolean bl) {
        boolean bl2 = client.vh;
        try {
            int n5;
            int n6;
            int n7;
            int n8;
            int n9;
            int n10;
            int n11;
            int[] nArray;
            int n12;
            block56: {
                int n13;
                int n14;
                int n15;
                int n16;
                int n17;
                int n18;
                block55: {
                    int n19;
                    block54: {
                        int n20;
                        block53: {
                            ++b;
                            n18 = n12 = fontMetrics.charWidth(c2);
                            if (bl) {
                                try {
                                    if (c2 == 'f' || c2 == 't' || 'w' == c2 || ~c2 == '\uffffff89' || c2 == 'k' || ~c2 == '\uffffff87' || 'y' == c2 || c2 == 'A' || ~c2 == '\uffffffa9' || ~c2 == '\uffffffa8') {
                                        ++n12;
                                    }
                                    if (~c2 == '\uffffffd0') {
                                        bl = false;
                                    }
                                }
                                catch (Exception exception) {
                                    // empty catch block
                                }
                            }
                            n17 = fontMetrics.getMaxAscent();
                            n16 = fontMetrics.getMaxAscent() - -fontMetrics.getMaxDescent();
                            n15 = fontMetrics.getHeight();
                            Image image = e2.createImage(n12, n16);
                            if (null == image) {
                                return false;
                            }
                            Graphics graphics = image.getGraphics();
                            graphics.setColor(Color.black);
                            graphics.fillRect(0, 0, n12, n16);
                            graphics.setColor(Color.white);
                            graphics.setFont(font);
                            graphics.drawString(c2 + "", 0, n17);
                            if (bl) {
                                graphics.drawString(c2 + "", 1, n17);
                            }
                            nArray = new int[n16 * n12];
                            PixelGrabber pixelGrabber = new PixelGrabber(image, 0, 0, n12, n16, nArray, 0, n12);
                            try {
                                pixelGrabber.grabPixels();
                            }
                            catch (InterruptedException interruptedException) {
                                return false;
                            }
                            image.flush();
                            image = null;
                            n11 = 0;
                            n14 = 0;
                            n10 = n12;
                            n9 = n16;
                            n8 = 0;
                            block27: while (true) {
                                int n21 = n8;
                                int n22 = n16;
                                block28: while (n21 < n22) {
                                    n20 = 0;
                                    if (bl2) break block53;
                                    n7 = n20;
                                    while (~n12 < ~n7) {
                                        n6 = nArray[n7 - -(n8 * n12)];
                                        n21 = 0;
                                        n22 = 0xFFFFFF & n6;
                                        if (bl2) continue block28;
                                        if (n21 != n22) {
                                            n14 = n8;
                                            if (!bl2) break block27;
                                        }
                                        ++n7;
                                        if (!bl2) continue;
                                    }
                                    ++n8;
                                    if (!bl2) continue block27;
                                }
                                break;
                            }
                            n20 = 0;
                        }
                        n8 = n20;
                        block30: while (true) {
                            int n23 = n8;
                            int n24 = n12;
                            block31: while (n23 < n24) {
                                n19 = 0;
                                if (bl2) break block54;
                                for (n7 = v625190; n16 > n7; ++n7) {
                                    n6 = nArray[n8 - -(n7 * n12)];
                                    n23 = 0;
                                    n24 = n6 & 0xFFFFFF;
                                    if (bl2) continue block31;
                                    if (n23 == n24) continue;
                                    n11 = n8;
                                    if (!bl2) break block30;
                                    if (!bl2) continue;
                                }
                                ++n8;
                                if (!bl2) continue block30;
                            }
                            break;
                        }
                        n8 = n16 - 1;
                        n19 = n4;
                    }
                    if (n19 >= -86) {
                        s.a(-60, null, 49, -85, null, '\ufff8', null, true);
                    }
                    block33: while (true) {
                        int n25 = 0;
                        int n26 = n8;
                        block34: while (n25 <= n26) {
                            n13 = 0;
                            if (bl2) break block55;
                            n7 = n13;
                            while (~n7 > ~n12) {
                                n6 = nArray[n7 + n8 * n12];
                                n25 = ~(n6 & 0xFFFFFF);
                                n26 = -1;
                                if (bl2) continue block34;
                                if (n25 != n26) {
                                    n9 = 1 + n8;
                                    if (!bl2) break block33;
                                }
                                ++n7;
                                if (!bl2) continue;
                            }
                            --n8;
                            if (!bl2) continue block33;
                        }
                        break;
                    }
                    n13 = -1 + n12;
                }
                n8 = n13;
                block36: while (true) {
                    int n27 = -1;
                    int n28 = ~n8;
                    block37: while (n27 >= n28) {
                        n5 = 0;
                        if (bl2) break block56;
                        n7 = n5;
                        while (~n7 > ~n16) {
                            n6 = nArray[n8 + n7 * n12];
                            n27 = -1;
                            n28 = ~(0xFFFFFF & n6);
                            if (bl2) continue block37;
                            if (n27 != n28) {
                                n10 = n8 + 1;
                                if (!bl2) break block36;
                            }
                            ++n7;
                            if (!bl2) continue;
                        }
                        --n8;
                        if (!bl2) continue block36;
                    }
                    break;
                }
                qb.k[0 + 9 * n3] = (byte)(b.c / 16384);
                qb.k[n3 * 9 - -1] = (byte)ib.a(b.c / 128, 127);
                qb.k[2 + 9 * n3] = (byte)ib.a(b.c, 127);
                qb.k[3 + n3 * 9] = (byte)(-n11 + n10);
                qb.k[n3 * 9 - -4] = (byte)(n9 + -n14);
                qb.k[9 * n3 - -5] = (byte)n11;
                qb.k[6 + n3 * 9] = (byte)(n17 + -n14);
                qb.k[7 + 9 * n3] = (byte)n18;
                qb.k[n3 * 9 - -8] = (byte)n15;
                n5 = n14;
            }
            n8 = n5;
            do {
                int n29 = n9;
                int n30 = n8;
                block40: while (true) {
                    if (n29 <= n30) return 1 != 0;
                    int n31 = n11;
                    if (bl2) return n31 != 0;
                    n7 = n31;
                    while (~n7 > ~n10) {
                        n6 = 0xFF & nArray[n12 * n8 + n7];
                        n29 = 30;
                        n30 = n6;
                        if (bl2) continue block40;
                        if (n29 < n30 && ~n6 > -231) {
                            fb.k[n2] = true;
                        }
                        qb.k[b.c++] = (byte)n6;
                        ++n7;
                        if (!bl2) continue;
                    }
                    break;
                }
                ++n8;
            } while (!bl2);
            return 1 != 0;
        }
        catch (RuntimeException runtimeException) {
            String string;
            StringBuilder stringBuilder = new StringBuilder().append(z[0]).append(n2).append(',').append(font != null ? z[1] : z[2]).append(',').append(n3).append(',').append(n4).append(',').append(e2 != null ? z[1] : z[2]).append(',').append(c2).append(',');
            if (fontMetrics != null) {
                string = z[1];
                throw i.a(runtimeException, stringBuilder.append(string).append(',').append(bl).append(')').toString());
            }
            string = z[2];
            throw i.a(runtimeException, stringBuilder.append(string).append(',').append(bl).append(')').toString());
        }
    }

    static {
        z = new String[]{s.z(s.z("\u0018{#+")), s.z(s.z("\u0010{L-\u0013")), s.z(s.z("\u0005 \u000eo"))};
        e = new String[]{s.z(s.z(".;\u0016f\u001cK;\u0017n\f\u000e'Bl\bK<\u0016f\u0003\u0018u\u0016lN\u00043\u0004f\u001cK4\fgN\u001b'\u0007p\u001dK0\fw\u000b\u0019"))};
        c = new BigInteger(s.z(s.z("ZeR3_")), 16);
        d = 0;
        a = null;
    }

    private static char[] z(String string) {
        char[] cArray = string.toCharArray();
        if (cArray.length < 2) {
            cArray = cArray;
            cArray[0] = (char)(cArray[0] ^ 0x6E);
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
                        n4 = 107;
                        break;
                    }
                    case 1: {
                        n4 = 85;
                        break;
                    }
                    case 2: {
                        n4 = 98;
                        break;
                    }
                    case 3: {
                        n4 = 3;
                        break;
                    }
                    default: {
                        n4 = 110;
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

