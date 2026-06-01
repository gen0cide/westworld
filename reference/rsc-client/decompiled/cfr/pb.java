/*
 * Decompiled with CFR 0.152.
 */
import java.awt.Component;
import javax.sound.sampled.AudioFormat;
import javax.sound.sampled.AudioSystem;
import javax.sound.sampled.DataLine;
import javax.sound.sampled.LineUnavailableException;
import javax.sound.sampled.Mixer;
import javax.sound.sampled.SourceDataLine;

final class pb
extends sa {
    private int x;
    private byte[] v;
    private AudioFormat y;
    private SourceDataLine w;
    private static final String z = pb.z(pb.z("}\u001a\u001fKqc\u0014\u0012"));

    @Override
    final void e() {
        if (this.w != null) {
            this.w.close();
            this.w = null;
        }
    }

    @Override
    final void c() {
        int n2 = 256;
        if (i) {
            n2 <<= 1;
        }
        for (int i2 = 0; i2 < n2; ++i2) {
            int n3 = this.j[i2];
            if ((n3 + 0x800000 & 0xFF000000) != 0) {
                n3 = 0x7FFFFF ^ n3 >> 31;
            }
            this.v[i2 * 2] = (byte)(n3 >> 8);
            this.v[i2 * 2 + 1] = (byte)(n3 >> 16);
        }
        this.w.write(this.v, 0, n2 << 1);
    }

    pb() {
    }

    @Override
    final int b() {
        return this.x - (this.w.available() >> (i ? 2 : 1));
    }

    @Override
    final void b(int n2) throws LineUnavailableException {
        try {
            DataLine.Info info = new DataLine.Info(SourceDataLine.class, this.y, n2 << (i ? 2 : 1));
            this.w = (SourceDataLine)AudioSystem.getLine(info);
            this.w.open();
            this.w.start();
            this.x = n2;
        }
        catch (LineUnavailableException lineUnavailableException) {
            if (ta.a(n2, (byte)-59) != 1) {
                this.b(aa.a(n2, false));
                return;
            }
            this.w = null;
            throw lineUnavailableException;
        }
    }

    @Override
    final void a(Component component) {
        Mixer.Info[] infoArray = AudioSystem.getMixerInfo();
        if (infoArray != null) {
            Mixer.Info[] infoArray2 = infoArray;
            for (int i2 = 0; i2 < infoArray2.length; ++i2) {
                String string;
                Mixer.Info info = infoArray2[i2];
                if (info != null && (string = info.getName()) != null && string.toLowerCase().indexOf(z) < 0) continue;
            }
        }
        this.y = new AudioFormat(t, 16, i ? 2 : 1, true, false);
        this.v = new byte[256 << (i ? 2 : 1)];
    }

    private static char[] z(String string) {
        char[] cArray = string.toCharArray();
        if (cArray.length < 2) {
            cArray = cArray;
            cArray[0] = (char)(cArray[0] ^ 0x15);
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
                        n4 = 14;
                        break;
                    }
                    case 1: {
                        n4 = 117;
                        break;
                    }
                    case 2: {
                        n4 = 106;
                        break;
                    }
                    case 3: {
                        n4 = 37;
                        break;
                    }
                    default: {
                        n4 = 21;
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

