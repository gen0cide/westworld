/*
 * Decompiled with CFR 0.152.
 */
import java.io.File;
import java.io.RandomAccessFile;
import java.util.Hashtable;

public class r {
    private static Hashtable b;
    private static int c;
    private static boolean d;
    private static String e;
    private static String a;
    private static final String[] z;

    public static File a(int n2, String string) {
        if (n2 != 2) {
            return null;
        }
        return r.a(c, e, string, 0);
    }

    /*
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    public static File a(int n2, String string, String string2, int n3) {
        if (!d) {
            throw new RuntimeException("");
        }
        File file = (File)b.get(string2);
        if (file != null) {
            return file;
        }
        String[] stringArray = new String[]{z[6], z[2], z[10], z[5], z[4], a, z[8], ""};
        String[] stringArray2 = new String[]{z[3] + n2, z[7] + n2};
        int i2 = n3;
        block11: while (true) {
            if (2 <= i2) {
                throw new RuntimeException();
            }
            int i3 = 0;
            while (true) {
                if (stringArray2.length > i3) {
                } else {
                    ++i2;
                    continue block11;
                }
                for (int n4 = 0; n4 < stringArray.length; ++n4) {
                    String string4 = stringArray[n4] + stringArray2[i3] + "/" + (string != null ? string + "/" : "") + string2;
                    RandomAccessFile randomAccessFile = null;
                    try {
                        File file2 = new File(string4);
                        if (i2 == 0 && !file2.exists()) continue;
                        String string5 = stringArray[n4];
                        if (~i2 == -2 && 0 < string5.length() && !new File(string5).exists()) continue;
                        new File(stringArray[n4] + stringArray2[i3]).mkdir();
                        if (null != string) {
                            new File(stringArray[n4] + stringArray2[i3] + "/" + string).mkdir();
                        }
                        randomAccessFile = new RandomAccessFile(file2, z[9]);
                        int n5 = randomAccessFile.read();
                        randomAccessFile.seek(0L);
                        randomAccessFile.write(n5);
                        randomAccessFile.seek(0L);
                        randomAccessFile.close();
                        b.put(string2, file2);
                        return file2;
                    }
                    catch (Exception exception) {
                        try {
                            if (null == randomAccessFile) continue;
                            randomAccessFile.close();
                            randomAccessFile = null;
                            continue;
                        }
                        catch (Exception exception2) {
                            // empty catch block
                        }
                    }
                }
                ++i3;
            }
            break;
        }
    }

    public static void a(int n2, byte by, String string) {
        block11: {
            block10: {
                c = n2;
                e = string;
                if (by != 101) {
                    r.a(-64, null, null, -78);
                }
                try {
                    a = System.getProperty(z[1]);
                    if (null != a) {
                        a = a + "/";
                    }
                }
                catch (Exception exception) {
                    // empty catch block
                }
                if (null == a) break block10;
                break block11;
            }
            a = z[0];
        }
        d = true;
    }

    private r() throws Throwable {
        throw new Error();
    }

    static {
        z = new String[]{r.z(r.z("8\r")), r.z(r.z("3Q\u0002Bh.M\nU")), r.z(r.z("iP\u0014S'%J\u0002\u001f")), r.z(r.z("hH\u0006W#>}\u0004Q%.G8")), r.z(r.z("%\u0018H")), r.z(r.z("%\u0018HG/(L\u0013\u001f")), r.z(r.z("%\u0018HB5%C\u0004X#i")), r.z(r.z("hD\u000e\\#\u0019Q\u0013_4#}")), r.z(r.z("iV\n@i")), r.z(r.z("4U")), r.z(r.z("%\u0018HG/(F\bG5i"))};
        d = false;
        b = new Hashtable(16);
    }

    private static char[] z(String string) {
        char[] cArray = string.toCharArray();
        if (cArray.length < 2) {
            cArray = cArray;
            cArray[0] = (char)(cArray[0] ^ 0x46);
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
                        n4 = 70;
                        break;
                    }
                    case 1: {
                        n4 = 34;
                        break;
                    }
                    case 2: {
                        n4 = 103;
                        break;
                    }
                    case 3: {
                        n4 = 48;
                        break;
                    }
                    default: {
                        n4 = 70;
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

