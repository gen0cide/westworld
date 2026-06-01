/*
 * Decompiled with CFR 0.152.
 */
import java.io.BufferedReader;
import java.io.IOException;
import java.io.InputStreamReader;
import java.io.OutputStream;
import java.io.PrintWriter;
import java.io.StringReader;
import java.io.StringWriter;
import java.io.Writer;
import java.lang.reflect.Method;
import java.net.InetSocketAddress;
import java.net.Proxy;
import java.net.ProxySelector;
import java.net.Socket;
import java.net.SocketAddress;
import java.net.URI;
import java.net.URISyntaxException;
import java.net.URL;
import java.nio.charset.Charset;
import java.util.List;

final class gb
extends m {
    static int o;
    static v n;
    private ProxySelector q;
    static int u;
    static int t;
    static int r;
    static int l;
    static int p;
    static int m;
    static int[] s;
    private static final String[] z;

    gb() {
        try {
            this.q = ProxySelector.getDefault();
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, z[31]);
        }
    }

    /*
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    private final Socket a(int n2, int n3, String string, String string2) throws IOException {
        boolean bl = client.vh;
        try {
            OutputStream outputStream;
            Socket socket;
            block24: {
                block23: {
                    ++o;
                    socket = new Socket(string2, n2);
                    socket.setSoTimeout(10000);
                    outputStream = socket.getOutputStream();
                    if (string == null) break block23;
                    outputStream.write((z[24] + this.h + ":" + this.f + z[26] + string + z[25]).getBytes(Charset.forName(z[27])));
                    if (!bl) break block24;
                }
                outputStream.write((z[24] + this.h + ":" + this.f + z[22]).getBytes(Charset.forName(z[27])));
            }
            outputStream.flush();
            BufferedReader bufferedReader = new BufferedReader(new InputStreamReader(socket.getInputStream()));
            String string3 = bufferedReader.readLine();
            if (n3 != 1514) {
                return null;
            }
            if (string3 != null) {
                if (string3.startsWith(z[21])) return socket;
                if (string3.startsWith(z[23])) return socket;
                if (string3.startsWith(z[28]) || string3.startsWith(z[30])) {
                    String string4;
                    block22: {
                        int n4 = 0;
                        string4 = z[20];
                        string3 = bufferedReader.readLine();
                        do {
                            if (string3 == null) throw new fa("");
                            if (-51 >= ~n4) throw new fa("");
                            if (string3.toLowerCase().startsWith(string4)) break block22;
                            ++n4;
                            string3 = bufferedReader.readLine();
                        } while (!bl);
                        throw new fa("");
                    }
                    int n5 = (string3 = string3.substring(string4.length()).trim()).indexOf(32);
                    if (-1 == n5) {
                        throw new fa(string3);
                    }
                    string3 = string3.substring(0, n5);
                    throw new fa(string3);
                }
            }
            outputStream.close();
            bufferedReader.close();
            socket.close();
            return null;
        }
        catch (RuntimeException runtimeException) {
            String string5;
            StringBuilder stringBuilder = new StringBuilder().append(z[29]).append(n2).append(',').append(n3).append(',').append(string != null ? z[11] : z[15]).append(',');
            if (string2 != null) {
                string5 = z[11];
                throw i.a(runtimeException, stringBuilder.append(string5).append(')').toString());
            }
            string5 = z[15];
            throw i.a(runtimeException, stringBuilder.append(string5).append(')').toString());
        }
    }

    /*
     * Unable to fully structure code
     */
    static final void a(int var0, int var1_1, byte var2_2, int var3_3, int var4_4, int var5_5, int[] var6_6, int var7_7, int var8_8, int var9_9, int var10_10, int var11_11, int[] var12_12, int var13_13, int var14_14) {
        block48: {
            var21_15 = client.vh;
            try {
                block43: {
                    block42: {
                        block41: {
                            block40: {
                                block39: {
                                    block38: {
                                        block37: {
                                            block36: {
                                                ++gb.m;
                                                if (var14_14 <= 0) {
                                                    return;
                                                }
                                                var15_16 = 0;
                                                var16_18 = 0;
                                                if (-1 != ~var3_3) break block36;
                                                break block37;
                                            }
                                            var11_11 = var8_8 / var3_3 << -410027673;
                                            var10_10 = var0 / var3_3 << -1978637785;
                                        }
                                        var19_19 = 0;
                                        if (~var11_11 <= -1) ** GOTO lbl23
                                        var11_11 = 0;
                                        if (!var21_15) break block39;
lbl23:
                                        // 3 sources

                                        if (16256 < var11_11) break block38;
                                        break block39;
                                        catch (RuntimeException v1) {
                                            throw v1;
                                        }
                                    }
                                    var11_11 = 16256;
                                }
                                if (var2_2 != 50) {
                                    return;
                                }
                                var3_3 += var9_9;
                                var0 += var13_13;
                                var8_8 += var1_1;
                                if (var3_3 != 0) break block40;
                                break block41;
                            }
                            var16_18 = var0 / var3_3 << -2053567033;
                            var15_16 = var8_8 / var3_3 << 983280871;
                        }
                        if (var15_16 < 0) break block42;
                        if (~var15_16 >= -16257) break block43;
                        var15_16 = 16256;
                        if (!var21_15) break block43;
                    }
                    var15_16 = 0;
                }
                var17_20 = var15_16 + -var11_11 >> -858142236;
                var18_21 = var16_18 - var10_10 >> 1912830340;
                var20_22 = var14_14 >> -1893613788;
                while (~var20_22 < -1) {
                    block47: {
                        block50: {
                            block46: {
                                block45: {
                                    block44: {
                                        var19_19 = var4_4 >> -407765257;
                                        var12_12[var7_7++] = var6_6[ib.a(16256, var10_10) + ((var11_11 += var4_4 & 0x600000) >> 897288903)] >>> var19_19;
                                        var12_12[var7_7++] = var6_6[((var11_11 += var17_20) >> -1712778393) + ib.a(16256, var10_10 += var18_21)] >>> var19_19;
                                        var12_12[var7_7++] = var6_6[((var11_11 += var17_20) >> -50303545) + ib.a(16256, var10_10 += var18_21)] >>> var19_19;
                                        var12_12[var7_7++] = var6_6[((var11_11 += var17_20) >> -755517209) + ib.a(16256, var10_10 += var18_21)] >>> var19_19;
                                        var11_11 += var17_20;
                                        var11_11 = (0x600000 & (var4_4 += var5_5)) + (16383 & var11_11);
                                        var19_19 = var4_4 >> 779290135;
                                        var12_12[var7_7++] = var6_6[(var11_11 >> 1348344199) + ib.a(var10_10 += var18_21, 16256)] >>> var19_19;
                                        var12_12[var7_7++] = var6_6[ib.a(16256, var10_10 += var18_21) + ((var11_11 += var17_20) >> -1943849369)] >>> var19_19;
                                        var12_12[var7_7++] = var6_6[((var11_11 += var17_20) >> -170574425) + ib.a(16256, var10_10 += var18_21)] >>> var19_19;
                                        var12_12[var7_7++] = var6_6[ib.a(16256, var10_10 += var18_21) - -((var11_11 += var17_20) >> 583157351)] >>> var19_19;
                                        var11_11 += var17_20;
                                        var11_11 = ((var4_4 += var5_5) & 0x600000) + (16383 & var11_11);
                                        var19_19 = var4_4 >> -1894754601;
                                        var4_4 += var5_5;
                                        var12_12[var7_7++] = var6_6[ib.a(var10_10 += var18_21, 16256) - -(var11_11 >> 1007320743)] >>> var19_19;
                                        var12_12[var7_7++] = var6_6[ib.a(16256, var10_10 += var18_21) - -((var11_11 += var17_20) >> 1276542663)] >>> var19_19;
                                        var12_12[var7_7++] = var6_6[((var11_11 += var17_20) >> 646825255) + ib.a(16256, var10_10 += var18_21)] >>> var19_19;
                                        var12_12[var7_7++] = var6_6[((var11_11 += var17_20) >> 1218318887) + ib.a(var10_10 += var18_21, 16256)] >>> var19_19;
                                        var11_11 += var17_20;
                                        var11_11 = (16383 & var11_11) + (0x600000 & var4_4);
                                        var19_19 = var4_4 >> 1764949655;
                                        var12_12[var7_7++] = var6_6[(var11_11 >> -1446967161) + ib.a(16256, var10_10 += var18_21)] >>> var19_19;
                                        var4_4 += var5_5;
                                        var12_12[var7_7++] = var6_6[((var11_11 += var17_20) >> 1931622023) + ib.a(var10_10 += var18_21, 16256)] >>> var19_19;
                                        var12_12[var7_7++] = var6_6[((var11_11 += var17_20) >> 821292839) + ib.a(16256, var10_10 += var18_21)] >>> var19_19;
                                        var12_12[var7_7++] = var6_6[((var11_11 += var17_20) >> -698624601) + ib.a(16256, var10_10 += var18_21)] >>> var19_19;
                                        var11_11 = var15_16;
                                        var10_10 = var16_18;
                                        var0 += var13_13;
                                        var3_3 += var9_9;
                                        var8_8 += var1_1;
                                        v5 = ~var3_3;
                                        v6 = -1;
                                        if (!var21_15) {
                                            if (v5 != v6) break block44;
                                            break block45;
                                        }
                                        ** GOTO lbl131
                                    }
                                    var16_18 = var0 / var3_3 << 204957255;
                                    var15_16 = var8_8 / var3_3 << -586213497;
                                }
                                if (-1 < ~var15_16) break block50;
                                if (16256 < var15_16) break block46;
                                break block47;
                                catch (RuntimeException v8) {
                                    throw v8;
                                }
                            }
                            var15_16 = 16256;
                            if (!var21_15) break block47;
                        }
                        var15_16 = 0;
                    }
                    var18_21 = -var10_10 + var16_18 >> -1259225180;
                    var17_20 = -var11_11 + var15_16 >> 1647676292;
                    --var20_22;
                    if (!var21_15) continue;
                }
                var20_22 = 0;
                do {
                    block49: {
                        block51: {
                            v5 = ~var20_22;
                            v6 = ~(15 & var14_14);
lbl131:
                            // 3 sources

                            if (v5 <= v6) break block48;
                            if (var21_15) break block48;
                            break block51;
                            catch (RuntimeException v10) {
                                throw v10;
                            }
                        }
                        if ((var20_22 & 3) != 0) break block49;
                        var19_19 = var4_4 >> -851479113;
                        var11_11 = (var4_4 & 0x600000) + (16383 & var11_11);
                        var4_4 += var5_5;
                    }
                    var12_12[var7_7++] = var6_6[(var11_11 >> -1445030201) + ib.a(var10_10, 16256)] >>> var19_19;
                    var10_10 += var18_21;
                    var11_11 += var17_20;
                    ++var20_22;
                } while (!var21_15);
            }
            catch (RuntimeException var15_17) {
                v12 = var15_17;
                v13 = new StringBuilder().append(gb.z[32]).append(var0).append(',').append(var1_1).append(',').append(var2_2).append(',').append(var3_3).append(',').append(var4_4).append(',').append(var5_5).append(',');
                v14 = var6_6 != null ? gb.z[11] : gb.z[15];
                v16 = v13.append(v14).append(',').append(var7_7).append(',').append(var8_8).append(',').append(var9_9).append(',').append(var10_10).append(',').append(var11_11).append(',');
                v17 = var12_12 != null ? gb.z[11] : gb.z[15];
                throw i.a(v12, v16.append(v17).append(',').append(var13_13).append(',').append(var14_14).append(')').toString());
            }
        }
    }

    static final i[] a(int n2) {
        try {
            if (n2 <= 37) {
                n = null;
            }
            ++t;
            return new i[]{eb.e, fb.h, f.b};
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, z[19] + n2 + ')');
        }
    }

    @Override
    final Socket a(byte by) throws IOException {
        boolean bl = client.vh;
        try {
            Socket socket;
            block20: {
                block23: {
                    fa fa2;
                    block22: {
                        List<Proxy> list;
                        List<Proxy> list2;
                        ++r;
                        boolean bl2 = Boolean.parseBoolean(System.getProperty(z[2]));
                        if (!bl2) {
                            System.setProperty(z[2], z[3]);
                        }
                        boolean bl3 = ~this.f == -444;
                        boolean bl4 = bl3;
                        try {
                            URI uRI;
                            list2 = this.q.select(new URI((bl4 ? z[0] : z[4]) + z[1] + this.h));
                            ProxySelector proxySelector = this.q;
                            URI uRI2 = uRI;
                            URI uRI3 = uRI;
                            StringBuilder stringBuilder = new StringBuilder();
                            String string = !bl4 ? z[0] : z[4];
                            uRI2(stringBuilder.append(string).append(z[1]).append(this.h).toString());
                            list = proxySelector.select(uRI3);
                        }
                        catch (URISyntaxException uRISyntaxException) {
                            return this.a(false);
                        }
                        list2.addAll(list);
                        Object[] objectArray = list2.toArray();
                        fa2 = null;
                        Object[] objectArray2 = objectArray;
                        if (by != 50) {
                            return null;
                        }
                        int n2 = 0;
                        while (~objectArray2.length < ~n2) {
                            block21: {
                                Object object = objectArray2[n2];
                                Proxy proxy = (Proxy)object;
                                try {
                                    Socket socket2 = this.a(proxy, 16256);
                                    socket = socket2;
                                    if (bl) break block20;
                                    if (socket == null) break block21;
                                    return socket2;
                                }
                                catch (fa fa3) {
                                    fa2 = fa3;
                                }
                                catch (IOException iOException) {
                                    // empty catch block
                                }
                            }
                            ++n2;
                            if (!bl) continue;
                        }
                        if (null != fa2) break block22;
                        break block23;
                    }
                    throw fa2;
                }
                socket = this.a(false);
            }
            return socket;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, z[5] + by + ')');
        }
    }

    /*
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    static final String a(boolean bl, Throwable throwable) throws IOException {
        String string2;
        Object object;
        boolean bl2;
        block16: {
            block15: {
                bl2 = client.vh;
                ++u;
                if (!(throwable instanceof la)) break block15;
                object = (la)throwable;
                throwable = ((la)object).e;
                string2 = ((la)object).h + z[16];
                if (!bl2) break block16;
            }
            string2 = "";
        }
        object = new StringWriter();
        PrintWriter printWriter = new PrintWriter((Writer)object);
        throwable.printStackTrace(printWriter);
        printWriter.close();
        String string3 = ((StringWriter)object).toString();
        BufferedReader bufferedReader = new BufferedReader(new StringReader(string3));
        if (bl) {
            s = null;
        }
        String string4 = bufferedReader.readLine();
        do {
            int n4;
            String string5;
            int n2;
            int n3;
            String string;
            block18: {
                block17: {
                    if (null == (string = bufferedReader.readLine())) {
                        if (!bl2) return string2 + z[18] + string4;
                    }
                    n3 = string.indexOf(40);
                    n2 = string.indexOf(41, n3 - -1);
                    if (0 == ~n3) break block17;
                    string5 = string.substring(0, n3);
                    if (!bl2) break block18;
                }
                string5 = string;
            }
            string5 = string5.trim();
            string5 = string5.substring(1 + string5.lastIndexOf(32));
            string5 = string5.substring(string5.lastIndexOf(9) + 1);
            string2 = string2 + string5;
            if (~n3 != 0 && -1 != n2 && (n4 = string.indexOf(z[17], n3)) >= 0) {
                string2 = string2 + string.substring(5 + n4, n2);
            }
            string2 = string2 + ' ';
        } while (!bl2);
        return string2 + z[18] + string4;
    }

    private final Socket a(Proxy proxy, int n2) throws IOException {
        try {
            block23: {
                block22: {
                    SocketAddress socketAddress;
                    block21: {
                        block20: {
                            block19: {
                                block18: {
                                    ++l;
                                    if (proxy.type() == Proxy.Type.DIRECT) break block18;
                                    break block19;
                                }
                                return this.a(false);
                            }
                            socketAddress = proxy.address();
                            if (!(socketAddress instanceof InetSocketAddress)) break block20;
                            break block21;
                        }
                        return null;
                    }
                    if (n2 != 16256) {
                        p = 123;
                    }
                    InetSocketAddress inetSocketAddress = (InetSocketAddress)socketAddress;
                    if (proxy.type() == Proxy.Type.HTTP) {
                        String string = null;
                        try {
                            Class<?> clazz = Class.forName(z[6]);
                            Method method = clazz.getDeclaredMethod(z[9], String.class, Integer.TYPE);
                            method.setAccessible(true);
                            Object object = method.invoke(null, inetSocketAddress.getHostName(), new Integer(inetSocketAddress.getPort()));
                            if (null != object) {
                                Method method2 = clazz.getDeclaredMethod(z[12], new Class[0]);
                                method2.setAccessible(true);
                                if (((Boolean)method2.invoke(object, new Object[0])).booleanValue()) {
                                    Method method3 = clazz.getDeclaredMethod(z[10], new Class[0]);
                                    method3.setAccessible(true);
                                    Method method4 = clazz.getDeclaredMethod(z[13], URL.class, String.class);
                                    method4.setAccessible(true);
                                    String string2 = (String)method3.invoke(object, new Object[0]);
                                    String string3 = (String)method4.invoke(object, new URL(z[8] + this.h + "/"), z[0]);
                                    string = string2 + z[7] + string3;
                                }
                            }
                        }
                        catch (Exception exception) {
                            // empty catch block
                        }
                        return this.a(inetSocketAddress.getPort(), 1514, string, inetSocketAddress.getHostName());
                    }
                    if (proxy.type() == Proxy.Type.SOCKS) break block22;
                    break block23;
                }
                Socket socket = new Socket(proxy);
                socket.connect(new InetSocketAddress(this.h, this.f));
                return socket;
            }
            return null;
        }
        catch (RuntimeException runtimeException) {
            RuntimeException runtimeException2 = runtimeException;
            StringBuilder stringBuilder = new StringBuilder().append(z[14]);
            String string = proxy != null ? z[11] : z[15];
            throw i.a(runtimeException2, stringBuilder.append(string).append(',').append(n2).append(')').toString());
        }
    }

    static {
        z = new String[]{gb.z(gb.z("\u001cHBW\u0004")), gb.z(gb.z("N\u0013\u0019")), gb.z(gb.z("\u001e]@FY\u001aYB\t\u0002\u0007Ye^\u0004\u0000Y[w\u0005\u001bD_B\u0004")), gb.z(gb.z("\u0000NCB")), gb.z(gb.z("\u001cHBW")), gb.z(gb.z("\u0013^\u0018c_")), gb.z(gb.z("\u0007IX\t\u0019\u0011H\u0018P\u0000\u0003\u0012FU\u0018\u0000SUH\u001bZTBS\u0007Z}CS\u001f\u0011RBN\u0014\u0015H_H\u0019=RPH")), gb.z(gb.z("N\u001c")), gb.z(gb.z("\u001cHBW\u0004N\u0013\u0019")), gb.z(gb.z("\u0013YBw\u0005\u001bDOf\u0002\u0000T")), gb.z(gb.z("\u0013YBo\u0012\u0015XSU9\u0015QS")), gb.z(gb.z("\u000f\u0012\u0018\t\n")), gb.z(gb.z("\u0007IFW\u0018\u0006HEw\u0005\u0011Y[W\u0003\u001dJSf\u0002\u0000TYU\u001e\u000e]BN\u0018\u001a")), gb.z(gb.z("\u0013YBo\u0012\u0015XSU!\u0015PCB")), gb.z(gb.z("\u0013^\u0018a_")), gb.z(gb.z("\u001aIZK")), gb.z(gb.z("T@\u0016")), gb.z(gb.z("ZVWQ\u0016N")), gb.z(gb.z("\b\u001c")), gb.z(gb.z("\u0013^\u0018o_")), gb.z(gb.z("\u0004NY_\u000eY]CS\u001f\u0011RBN\u0014\u0015HS\u001dW")), gb.z(gb.z("<hbwXE\u0012\u0006\u0007ED\f")), gb.z(gb.z("Ttbs'[\r\u0018\u0017}~")), gb.z(gb.z("<hbwXE\u0012\u0007\u0007ED\f")), gb.z(gb.z("7sxi27h\u0016")), gb.z(gb.z("~6")), gb.z(gb.z("Ttbs'[\r\u0018\u0017}")), gb.z(gb.z("=oy\nOL\t\u000f\nF")), gb.z(gb.z("<hbwXE\u0012\u0006\u0007CD\u000b")), gb.z(gb.z("\u0013^\u0018`_")), gb.z(gb.z("<hbwXE\u0012\u0007\u0007CD\u000b")), gb.z(gb.z("\u0013^\u0018\u001b\u001e\u001aUB\u0019_]")), gb.z(gb.z("\u0013^\u0018b_"))};
        n = new v(gb.z(gb.z("=rbe2 }")), gb.z(gb.z("\u001bZPN\u0014\u0011")), gb.z(gb.z("+UXS\u0015\u0011HW")), 6);
    }

    private static char[] z(String string) {
        char[] cArray = string.toCharArray();
        if (cArray.length < 2) {
            cArray = cArray;
            cArray[0] = (char)(cArray[0] ^ 0x77);
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
                        n4 = 116;
                        break;
                    }
                    case 1: {
                        n4 = 60;
                        break;
                    }
                    case 2: {
                        n4 = 54;
                        break;
                    }
                    case 3: {
                        n4 = 39;
                        break;
                    }
                    default: {
                        n4 = 119;
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

