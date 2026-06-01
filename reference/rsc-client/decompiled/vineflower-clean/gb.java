import java.io.BufferedReader;
import java.io.IOException;
import java.io.InputStreamReader;
import java.io.OutputStream;
import java.io.PrintWriter;
import java.io.StringReader;
import java.io.StringWriter;
import java.lang.reflect.Method;
import java.net.InetSocketAddress;
import java.net.Proxy;
import java.net.ProxySelector;
import java.net.Socket;
import java.net.SocketAddress;
import java.net.URI;
import java.net.URISyntaxException;
import java.net.URL;
import java.net.Proxy.Type;
import java.nio.charset.Charset;
import java.util.List;

final class gb extends m {
   static int o;
   static v n = new v(z(z("=rbe2 }")), z(z("\u001bZPN\u0014\u0011")), z(z("+UXS\u0015\u0011HW")), 6);
   private ProxySelector q;
   static int u;
   static int t;
   static int r;
   static int l;
   static int p;
   static int m;
   static int[] s;
   private static final String[] z = new String[]{
      z(z("\u001cHBW\u0004")),
      z(z("N\u0013\u0019")),
      z(z("\u001e]@FY\u001aYB\t\u0002\u0007Ye^\u0004\u0000Y[w\u0005\u001bD_B\u0004")),
      z(z("\u0000NCB")),
      z(z("\u001cHBW")),
      z(z("\u0013^\u0018c_")),
      z(z("\u0007IX\t\u0019\u0011H\u0018P\u0000\u0003\u0012FU\u0018\u0000SUH\u001bZTBS\u0007Z}CS\u001f\u0011RBN\u0014\u0015H_H\u0019=RPH")),
      z(z("N\u001c")),
      z(z("\u001cHBW\u0004N\u0013\u0019")),
      z(z("\u0013YBw\u0005\u001bDOf\u0002\u0000T")),
      z(z("\u0013YBo\u0012\u0015XSU9\u0015QS")),
      z(z("\u000f\u0012\u0018\t\n")),
      z(z("\u0007IFW\u0018\u0006HEw\u0005\u0011Y[W\u0003\u001dJSf\u0002\u0000TYU\u001e\u000e]BN\u0018\u001a")),
      z(z("\u0013YBo\u0012\u0015XSU!\u0015PCB")),
      z(z("\u0013^\u0018a_")),
      z(z("\u001aIZK")),
      z(z("T@\u0016")),
      z(z("ZVWQ\u0016N")),
      z(z("\b\u001c")),
      z(z("\u0013^\u0018o_")),
      z(z("\u0004NY_\u000eY]CS\u001f\u0011RBN\u0014\u0015HS\u001dW")),
      z(z("<hbwXE\u0012\u0006\u0007ED\f")),
      z(z("Ttbs'[\r\u0018\u0017}~")),
      z(z("<hbwXE\u0012\u0007\u0007ED\f")),
      z(z("7sxi27h\u0016")),
      z(z("~6")),
      z(z("Ttbs'[\r\u0018\u0017}")),
      z(z("=oy\nOL\t\u000f\nF")),
      z(z("<hbwXE\u0012\u0006\u0007CD\u000b")),
      z(z("\u0013^\u0018`_")),
      z(z("<hbwXE\u0012\u0007\u0007CD\u000b")),
      z(z("\u0013^\u0018\u001b\u001e\u001aUB\u0019_]")),
      z(z("\u0013^\u0018b_"))
   };

   gb() {
      try {
         this.q = ProxySelector.getDefault();
      } catch (RuntimeException var2) {
         throw i.a(var2, z[31]);
      }
   }

   private final Socket a(int var1, int var2, String var3, String var4) throws IOException {
      boolean var12 = client.vh;

      try {
         Socket var5;
         OutputStream var6;
         label99: {
            o++;
            var5 = new Socket(var4, var1);
            var5.setSoTimeout(10000);
            var6 = var5.getOutputStream();
            if (var3 != null) {
               var6.write((z[24] + this.h + ":" + this.f + z[26] + var3 + z[25]).getBytes(Charset.forName(z[27])));
               if (!var12) {
                  break label99;
               }
            }

            var6.write((z[24] + this.h + ":" + this.f + z[22]).getBytes(Charset.forName(z[27])));
         }

         var6.flush();
         BufferedReader var7 = new BufferedReader(new InputStreamReader(var5.getInputStream()));
         String var8 = var7.readLine();
         if (var2 != 1514) {
            return (Socket)null;
         }

         if (var8 != null) {
            if (var8.startsWith(z[21]) || var8.startsWith(z[23])) {
               return var5;
            }

            if (var8.startsWith(z[28]) || var8.startsWith(z[30])) {
               int var9 = 0;
               String var10 = z[20];
               var8 = var7.readLine();

               while (var8 != null && -51 < ~var9) {
                  if (var8.toLowerCase().startsWith(var10)) {
                     var8 = var8.substring(var10.length()).trim();
                     int var11 = var8.indexOf(32);
                     if (-1 != var11) {
                        var8 = var8.substring(0, var11);
                     }

                     throw new fa(var8);
                  }

                  var9++;
                  var8 = var7.readLine();
                  if (var12) {
                     break;
                  }
               }

               throw new fa("");
            }
         }

         var6.close();
         var7.close();
         var5.close();
         return null;
      } catch (RuntimeException var13) {
         throw i.a(var13, z[29] + var1 + ',' + var2 + ',' + (var3 != null ? z[11] : z[15]) + ',' + (var4 != null ? z[11] : z[15]) + ')');
      }
   }

   // $VF: Irreducible bytecode was duplicated to produce valid code
   static final void a(
      int var0,
      int var1,
      byte var2,
      int var3,
      int var4,
      int var5,
      int[] var6,
      int var7,
      int var8,
      int var9,
      int var10,
      int var11,
      int[] var12,
      int var13,
      int var14
   ) {
      boolean var21 = client.vh;

      try {
         m++;
         if (var14 > 0) {
            int var15 = 0;
            int var16 = 0;
            if (-1 != ~var3) {
               var11 = var8 / var3 << -410027673;
               var10 = var0 / var3 << -1978637785;
            }

            int var19;
            label129: {
               var19 = 0;
               if (~var11 > -1) {
                  var11 = 0;
                  if (!var21) {
                     break label129;
                  }
               }

               if (16256 < var11) {
                  var11 = 16256;
               }
            }

            if (var2 == 50) {
               var3 += var9;
               var0 += var13;
               var8 += var1;
               if (var3 != 0) {
                  var16 = var0 / var3 << -2053567033;
                  var15 = var8 / var3 << 983280871;
               }

               label121: {
                  if (var15 >= 0) {
                     if (~var15 >= -16257) {
                        break label121;
                     }

                     var15 = 16256;
                     if (!var21) {
                        break label121;
                     }
                  }

                  var15 = 0;
               }

               int var17 = var15 + -var11 >> -858142236;
               int var18 = var16 - var10 >> 1912830340;
               int var20 = var14 >> -1893613788;

               int var41;
               int var10000;
               while (true) {
                  if (~var20 < -1) {
                     var11 += var4 & 6291456;
                     int var23 = var4 >> -407765257;
                     var12[var7++] = var6[ib.a(16256, var10) + (var11 >> 897288903)] >>> var23;
                     var4 += var5;
                     var11 += var17;
                     var10 += var18;
                     var12[var7++] = var6[(var11 >> -1712778393) + ib.a(16256, var10)] >>> var23;
                     var10 += var18;
                     var11 += var17;
                     var12[var7++] = var6[(var11 >> -50303545) + ib.a(16256, var10)] >>> var23;
                     var10 += var18;
                     var11 += var17;
                     var12[var7++] = var6[(var11 >> -755517209) + ib.a(16256, var10)] >>> var23;
                     var10 += var18;
                     var11 += var17;
                     var11 = (6291456 & var4) + (16383 & var11);
                     int var24 = var4 >> 779290135;
                     var12[var7++] = var6[(var11 >> 1348344199) + ib.a(var10, 16256)] >>> var24;
                     var4 += var5;
                     var11 += var17;
                     var10 += var18;
                     var12[var7++] = var6[ib.a(16256, var10) + (var11 >> -1943849369)] >>> var24;
                     var11 += var17;
                     var10 += var18;
                     var12[var7++] = var6[(var11 >> -170574425) + ib.a(16256, var10)] >>> var24;
                     var10 += var18;
                     var11 += var17;
                     var12[var7++] = var6[ib.a(16256, var10) - -(var11 >> 583157351)] >>> var24;
                     var10 += var18;
                     var11 += var17;
                     var11 = (var4 & 6291456) + (16383 & var11);
                     int var25 = var4 >> -1894754601;
                     var4 += var5;
                     var12[var7++] = var6[ib.a(var10, 16256) - -(var11 >> 1007320743)] >>> var25;
                     var11 += var17;
                     var10 += var18;
                     var12[var7++] = var6[ib.a(16256, var10) - -(var11 >> 1276542663)] >>> var25;
                     var11 += var17;
                     var10 += var18;
                     var12[var7++] = var6[(var11 >> 646825255) + ib.a(16256, var10)] >>> var25;
                     var10 += var18;
                     var11 += var17;
                     var12[var7++] = var6[(var11 >> 1218318887) + ib.a(var10, 16256)] >>> var25;
                     var11 += var17;
                     var10 += var18;
                     var11 = (16383 & var11) + (6291456 & var4);
                     var19 = var4 >> 1764949655;
                     var12[var7++] = var6[(var11 >> -1446967161) + ib.a(16256, var10)] >>> var19;
                     var4 += var5;
                     var11 += var17;
                     var10 += var18;
                     var12[var7++] = var6[(var11 >> 1931622023) + ib.a(var10, 16256)] >>> var19;
                     var11 += var17;
                     var10 += var18;
                     var12[var7++] = var6[(var11 >> 821292839) + ib.a(16256, var10)] >>> var19;
                     var10 += var18;
                     var11 += var17;
                     var12[var7++] = var6[(var11 >> -698624601) + ib.a(16256, var10)] >>> var19;
                     var11 = var15;
                     var10 = var16;
                     var0 += var13;
                     var3 += var9;
                     var8 += var1;
                     var10000 = ~var3;
                     var41 = -1;
                     if (var21) {
                        break;
                     }

                     if (var10000 != -1) {
                        var16 = var0 / var3 << 204957255;
                        var15 = var8 / var3 << -586213497;
                     }

                     label101: {
                        if (-1 >= ~var15) {
                           if (16256 >= var15) {
                              break label101;
                           }

                           var15 = 16256;
                           if (!var21) {
                              break label101;
                           }
                        }

                        var15 = 0;
                     }

                     var18 = -var10 + var16 >> -1259225180;
                     var17 = -var11 + var15 >> 1647676292;
                     var20--;
                     if (!var21) {
                        continue;
                     }

                     var20 = 0;
                  } else {
                     var20 = 0;
                  }

                  var10000 = ~var20;
                  var41 = ~(15 & var14);
                  break;
               }

               while (var10000 > var41 && !var21) {
                  if ((var20 & 3) == 0) {
                     var19 = var4 >> -851479113;
                     var11 = (var4 & 6291456) + (16383 & var11);
                     var4 += var5;
                  }

                  var12[var7++] = var6[(var11 >> -1445030201) + ib.a(var10, 16256)] >>> var19;
                  var10 += var18;
                  var11 += var17;
                  var20++;
                  if (var21) {
                     break;
                  }

                  var10000 = ~var20;
                  var41 = ~(15 & var14);
               }
            }
         }
      } catch (RuntimeException var22) {
         throw i.a(
            var22,
            z[32]
               + var0
               + ','
               + var1
               + ','
               + var2
               + ','
               + var3
               + ','
               + var4
               + ','
               + var5
               + ','
               + (var6 != null ? z[11] : z[15])
               + ','
               + var7
               + ','
               + var8
               + ','
               + var9
               + ','
               + var10
               + ','
               + var11
               + ','
               + (var12 != null ? z[11] : z[15])
               + ','
               + var13
               + ','
               + var14
               + ')'
         );
      }
   }

   static final i[] a(int var0) {
      try {
         if (var0 <= 37) {
            n = (v)null;
         }

         t++;
         return new i[]{eb.e, fb.h, f.b};
      } catch (RuntimeException var2) {
         throw i.a(var2, z[19] + var0 + ')');
      }
   }

   @Override
   final Socket a(byte var1) throws IOException {
      boolean var13 = client.vh;

      try {
         r++;
         boolean var2 = Boolean.parseBoolean(System.getProperty(z[2]));
         if (!var2) {
            System.setProperty(z[2], z[3]);
         }

         boolean var5 = ~this.f == -444;

         List var3;
         List var4;
         try {
            var3 = this.q.select(new URI((var5 ? z[0] : z[4]) + z[1] + this.h));
            var4 = this.q.select(new URI((!var5 ? z[0] : z[4]) + z[1] + this.h));
         } catch (URISyntaxException var14) {
            return this.a(false);
         }

         var3.addAll(var4);
         Object[] var6 = var3.toArray();
         fa var7 = null;
         Object[] var8 = var6;
         if (var1 != 50) {
            return (Socket)null;
         }

         int var9 = 0;

         Socket var10000;
         while (true) {
            if (~var8.length < ~var9) {
               Object var10 = var8[var9];
               Proxy var11 = (Proxy)var10;

               try {
                  Socket var12 = this.a(var11, 16256);
                  var10000 = var12;
                  if (var13) {
                     break;
                  }

                  if (var12 != null) {
                     return var12;
                  }
               } catch (fa var15) {
                  var7 = var15;
               } catch (IOException var16) {
               }

               var9++;
               if (!var13) {
                  continue;
               }
            }

            if (null != var7) {
               throw var7;
            }

            var10000 = this.a(false);
            break;
         }

         return var10000;
      } catch (RuntimeException var17) {
         throw i.a(var17, z[5] + var1 + ')');
      }
   }

   static final String a(boolean var0, Throwable var1) throws IOException {
      boolean var13 = client.vh;

      try {
         String var2;
         label53: {
            u++;
            if (var1 instanceof la) {
               la var3 = (la)var1;
               var1 = var3.e;
               var2 = var3.h + z[16];
               if (!var13) {
                  break label53;
               }
            }

            var2 = "";
         }

         StringWriter var16 = new StringWriter();
         PrintWriter var4 = new PrintWriter(var16);
         var1.printStackTrace(var4);
         var4.close();
         String var5 = var16.toString();
         BufferedReader var6 = new BufferedReader(new StringReader(var5));
         if (var0) {
            s = (int[])null;
         }

         String var7 = var6.readLine();

         do {
            String var8 = var6.readLine();
            if (null == var8 && !var13) {
               break;
            }

            int var9;
            int var10;
            String var11;
            label44: {
               var9 = var8.indexOf(40);
               var10 = var8.indexOf(41, var9 - -1);
               if (0 != ~var9) {
                  var11 = var8.substring(0, var9);
                  if (!var13) {
                     break label44;
                  }
               }

               var11 = var8;
            }

            var11 = var11.trim();
            var11 = var11.substring(1 + var11.lastIndexOf(32));
            var11 = var11.substring(var11.lastIndexOf(9) + 1);
            var2 = var2 + var11;
            if (~var9 != 0 && -1 != var10) {
               int var12 = var8.indexOf(z[17], var9);
               if (var12 >= 0) {
                  var2 = var2 + var8.substring(5 + var12, var10);
               }
            }

            var2 = var2 + ' ';
         } while (!var13);

         return var2 + z[18] + var7;
      } catch (RuntimeException var14) {
         throw var14;
      }
   }

   private final Socket a(Proxy var1, int var2) throws IOException {
      try {
         l++;
         if (var1.type() != Type.DIRECT) {
            SocketAddress var3 = var1.address();
            if (var3 instanceof InetSocketAddress) {
               if (var2 != 16256) {
                  p = 123;
               }

               InetSocketAddress var4 = (InetSocketAddress)var3;
               if (var1.type() == Type.HTTP) {
                  String var16 = null;

                  try {
                     Class var6 = Class.forName(z[6]);
                     Method var7 = var6.getDeclaredMethod(z[9], String.class, int.class);
                     var7.setAccessible(true);
                     Object var8 = var7.invoke(null, var4.getHostName(), new Integer(var4.getPort()));
                     if (null != var8) {
                        Method var9 = var6.getDeclaredMethod(z[12]);
                        var9.setAccessible(true);
                        if ((Boolean)var9.invoke(var8)) {
                           Method var10 = var6.getDeclaredMethod(z[10]);
                           var10.setAccessible(true);
                           Method var11 = var6.getDeclaredMethod(z[13], URL.class, String.class);
                           var11.setAccessible(true);
                           String var12 = (String)var10.invoke(var8);
                           String var13 = (String)var11.invoke(var8, new URL(z[8] + this.h + "/"), z[0]);
                           var16 = var12 + z[7] + var13;
                        }
                     }
                  } catch (Exception var14) {
                  }

                  return this.a(var4.getPort(), 1514, var16, var4.getHostName());
               } else {
                  if (var1.type() != Type.SOCKS) {
                     return null;
                  }

                  Socket var5 = new Socket(var1);
                  var5.connect(new InetSocketAddress(this.h, this.f));
                  return var5;
               }
            } else {
               return null;
            }
         } else {
            return this.a(false);
         }
      } catch (RuntimeException var15) {
         throw i.a(var15, z[14] + (var1 != null ? z[11] : z[15]) + ',' + var2 + ')');
      }
   }

   private static char[] z(String var0) {
      char[] var10000 = var0.toCharArray();
      if (var10000.length < 2) {
         var10000[0] = (char)(var10000[0] ^ 119);
      }

      return var10000;
   }

   private static String z(char[] var0) {
      int var10002 = var0.length;
      char[] var10001 = var0;
      int var10000 = var10002;

      for (int var1 = 0; var10000 > var1; var1++) {
         char var10004 = var10001[var1];
         byte var10005;
         switch (var1 % 5) {
            case 0:
               var10005 = 116;
               break;
            case 1:
               var10005 = 60;
               break;
            case 2:
               var10005 = 54;
               break;
            case 3:
               var10005 = 39;
               break;
            default:
               var10005 = 119;
         }

         var10001[var1] = (char)(var10004 ^ var10005);
      }

      return new String(var10001).intern();
   }
}
