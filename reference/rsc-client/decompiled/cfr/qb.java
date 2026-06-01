/*
 * Decompiled with CFR 0.152.
 */
import java.awt.AWTEvent;
import java.awt.Frame;
import java.awt.Graphics;
import java.awt.event.MouseEvent;

final class qb
extends Frame {
    static int b;
    static int[] e;
    static byte[] k;
    private e g;
    static int f;
    static int[][] d;
    static int i;
    private int c = 0;
    private int l;
    private int j = 28;
    private int h;
    static int a;
    private static final String[] z;

    @Override
    protected final void processEvent(AWTEvent aWTEvent) {
        try {
            if (aWTEvent instanceof MouseEvent) {
                MouseEvent mouseEvent = (MouseEvent)aWTEvent;
                aWTEvent = new MouseEvent(mouseEvent.getComponent(), mouseEvent.getID(), mouseEvent.getWhen(), mouseEvent.getModifiers(), mouseEvent.getX(), mouseEvent.getY() - 24, mouseEvent.getClickCount(), mouseEvent.isPopupTrigger());
            }
            ++a;
            super.processEvent(aWTEvent);
        }
        catch (RuntimeException runtimeException) {
            RuntimeException runtimeException2 = runtimeException;
            StringBuilder stringBuilder = new StringBuilder().append(z[5]);
            String string = aWTEvent != null ? z[0] : z[2];
            throw i.a(runtimeException2, stringBuilder.append(string).append(')').toString());
        }
    }

    @Override
    public final void paint(Graphics graphics) {
        try {
            ++b;
            this.g.paint(graphics);
        }
        catch (RuntimeException runtimeException) {
            RuntimeException runtimeException2 = runtimeException;
            StringBuilder stringBuilder = new StringBuilder().append(z[6]);
            String string = graphics != null ? z[0] : z[2];
            throw i.a(runtimeException2, stringBuilder.append(string).append(')').toString());
        }
    }

    @Override
    public final void resize(int n2, int n3) {
        try {
            super.resize(n2, this.j + n3);
            ++i;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, z[3] + n2 + ',' + n3 + ')');
        }
    }

    /*
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    @Override
    public final Graphics getGraphics() {
        try {
            Graphics graphics;
            block7: {
                block6: {
                    ++f;
                    graphics = super.getGraphics();
                    if (this.c == 0) break block6;
                    graphics.translate(-5, 0);
                    if (!client.vh) break block7;
                }
                graphics.translate(0, 24);
            }
            return graphics;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, z[4]);
        }
    }

    /*
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    qb(e e2, int n2, int n3, String string, boolean bl, boolean bl2) {
        try {
            block10: {
                block9: {
                    this.g = e2;
                    if (bl2) break block9;
                    this.j = 28;
                    if (!client.vh) break block10;
                }
                this.j = 48;
            }
            this.l = n2;
            this.h = n3;
            this.setTitle(string);
            this.setResizable(bl);
            this.show();
            this.toFront();
            this.resize(this.l, this.h);
            this.getGraphics();
            return;
        }
        catch (RuntimeException runtimeException) {
            String string2;
            StringBuilder stringBuilder = new StringBuilder().append(z[1]).append(e2 != null ? z[0] : z[2]).append(',').append(n2).append(',').append(n3).append(',');
            if (string != null) {
                string2 = z[0];
                throw i.a(runtimeException, stringBuilder.append(string2).append(',').append(bl).append(',').append(bl2).append(')').toString());
            }
            string2 = z[2];
            throw i.a(runtimeException, stringBuilder.append(string2).append(',').append(bl).append(',').append(bl2).append(')').toString());
        }
    }

    static {
        z = new String[]{qb.z(qb.z(":\u001am\u0006\u000b")), qb.z(qb.z("0Vm\u0014\u001f/]7\u0016^")), qb.z(qb.z("/A/D")), qb.z(qb.z("0VmZ\u00132]9M^")), qb.z(qb.z("0VmO\u00135s1I\u0006)] [^h")), qb.z(qb.z("0VmX\u0004.W&[\u0005\u0004B&F\u0002i")), qb.z(qb.z("0VmX\u0017(Z7\u0000"))};
        k = new byte[100000];
    }

    private static char[] z(String string) {
        char[] cArray = string.toCharArray();
        if (cArray.length < 2) {
            cArray = cArray;
            cArray[0] = (char)(cArray[0] ^ 0x76);
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
                        n4 = 65;
                        break;
                    }
                    case 1: {
                        n4 = 52;
                        break;
                    }
                    case 2: {
                        n4 = 67;
                        break;
                    }
                    case 3: {
                        n4 = 40;
                        break;
                    }
                    default: {
                        n4 = 118;
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

