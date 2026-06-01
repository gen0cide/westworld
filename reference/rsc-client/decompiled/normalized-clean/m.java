import java.awt.image.ColorModel;
import java.io.IOException;
import java.net.Socket;

abstract class m {
   static byte[][] b = new byte[50][];
   static int[] g;
   static int k;
   String h;
   static ob e = null;
   static int[] i = new int[256];
   int f;
   static int a;
   static ColorModel d;
   static int c;
   static int j;
   private static final String[] z = new String[]{
      z(z(".aKU")),
      z(z(";:\t\u0017#")),
      z(z("-:f\u0011")),
      z(z("-:e\u0011")),
      z(z(")zS\\9%f\t]?4")),
      z(z("\u0019{R\u00190%qC\u0019*/4E\\~!4J\\3\"qU\u0019*/4RJ;``OP-`{ES;#`")),
      z(z("\rqJ[;2g\u0007V<*qDM")),
      z(z("3`UP0':CX*")),
      z(z("-:d\u0011"))
   };

   static final int a(boolean var0, int var1, byte[] var2) {
      try {
         if (!var0) {
            b = (byte[][])null;
         }

         a++;
         return (var2[var1 + 3] & 0xFF)
            + (var2[var1 - -2] << 8 & 0xFF00)
            + ((0xFF & var2[var1]) << 24)
            - -((0xFF & var2[1 + var1]) << 16);
      } catch (RuntimeException var4) {
         throw i.a(var4, z[2] + var0 + 44 + var1 + 44 + (var2 != null ? z[1] : z[0]) + 41);
      }
   }

   final Socket a(boolean var1) throws IOException {
      try {
         k++;
         if (var1) {
            i = (int[])null;
         }

         return new Socket(this.h, this.f);
      } catch (RuntimeException var3) {
         throw i.a(var3, z[8] + var1 + ')');
      }
   }

   // $VF: Irreducible bytecode was duplicated to produce valid code
   static final void a(byte[] var0, byte var1, boolean var2) {
      boolean var6 = client.vh;

      try {
         kb.d = na.a(z[7], 0, var0, -124);
         c++;
         jb.p = 0;
         b.v = na.a(z[4], 0, var0, -125);
         ka.b = 0;
         gb.p = t.a(65525);
         fa.e = new int[gb.p];
         ac.x = new String[gb.p];
         ka.c = new int[gb.p];
         h.c = new int[gb.p];
         lb.ac = new String[gb.p];
         ua.Bb = new int[gb.p];
         kb.b = new int[gb.p];
         mb.k = new int[gb.p];
         kb.c = new int[gb.p];
         ga.b = new String[gb.p];
         gb.s = new int[gb.p];
         int var3 = 0;

         while (true) {
            if (gb.p > var3) {
               ac.x[var3] = o.a((byte)38);
               var3++;
               if (var6) {
                  break;
               }

               if (!var6) {
                  continue;
               }
            }

            var3 = 0;
            break;
         }

         while (true) {
            if (~var3 > ~gb.p) {
               ga.b[var3] = o.a((byte)38);
               var3++;
               if (var6) {
                  break;
               }

               if (!var6) {
                  continue;
               }
            }

            var3 = 0;
            break;
         }

         while (true) {
            if (gb.p > var3) {
               lb.ac[var3] = o.a((byte)38);
               var3++;
               if (var6) {
                  break;
               }

               if (!var6) {
                  continue;
               }
            }

            var3 = 0;
            break;
         }

         label1221: {
            label1220:
            while (true) {
               if (gb.p > var3) {
                  ua.Bb[var3] = t.a(65525);
                  int var10000 = ~mb.l;
                  int var10001 = ~(ua.Bb[var3] + 1);
                  if (!var6) {
                     if (var10000 > var10001) {
                        mb.l = ua.Bb[var3] + 1;
                     }

                     var3++;
                     if (!var6) {
                        continue;
                     }

                     var3 = 0;
                  } else {
                     if (var10000 <= var10001) {
                        break;
                     }

                     kb.b[var3] = ub.a((byte)-105);
                     var3++;
                     if (var6) {
                        break label1221;
                     }

                     if (var6) {
                        break;
                     }
                  }
               } else {
                  var3 = 0;
               }

               while (true) {
                  if (~var3 <= ~gb.p) {
                     break label1220;
                  }

                  kb.b[var3] = ub.a((byte)-105);
                  var3++;
                  if (var6) {
                     break label1221;
                  }

                  if (var6) {
                     break label1220;
                  }
               }
            }

            var3 = 0;
         }

         while (true) {
            if (gb.p > var3) {
               fa.e[var3] = v.a(-30504);
               var3++;
               if (var6) {
                  break;
               }

               if (!var6) {
                  continue;
               }
            }

            var3 = 0;
            break;
         }

         while (true) {
            if (gb.p > var3) {
               gb.s[var3] = v.a(-30504);
               var3++;
               if (var6) {
                  break;
               }

               if (!var6) {
                  continue;
               }
            }

            var3 = 0;
            break;
         }

         while (true) {
            if (~gb.p < ~var3) {
               mb.k[var3] = t.a(65525);
               var3++;
               if (var6) {
                  break;
               }

               if (!var6) {
                  continue;
               }
            }

            var3 = 0;
            break;
         }

         while (true) {
            if (var3 < gb.p) {
               h.c[var3] = ub.a((byte)-105);
               var3++;
               if (var6) {
                  break;
               }

               if (!var6) {
                  continue;
               }
            }

            var3 = 0;
            break;
         }

         while (true) {
            if (gb.p > var3) {
               kb.c[var3] = v.a(-30504);
               var3++;
               if (var6) {
                  break;
               }

               if (!var6) {
                  continue;
               }
            }

            var3 = 0;
            break;
         }

         while (true) {
            if (gb.p > var3) {
               ka.c[var3] = v.a(-30504);
               var3++;
               if (var6) {
                  break;
               }

               if (!var6) {
                  continue;
               }
            }

            var3 = 0;
            break;
         }

         int var24;
         byte var29;
         while (true) {
            if (var3 < gb.p) {
               var24 = 1;
               var29 = var2;
               if (var6) {
                  break;
               }

               if (1 != var2 && -2 == ~ka.c[var3]) {
                  ac.x[var3] = z[6];
                  ga.b[var3] = z[5];
                  kb.b[var3] = 0;
                  lb.ac[var3] = "";
                  gb.s[0] = 0;
                  mb.k[var3] = 0;
                  kb.c[var3] = 1;
               }

               var3++;
               if (!var6) {
                  continue;
               }
            }

            la.d = t.a(65525);
            fb.d = new int[la.d];
            b.h = new int[la.d];
            jb.k = new int[la.d];
            ob.h = new int[la.d];
            la.a = new int[la.d];
            g = new int[la.d];
            v.e = new int[la.d];
            o.a = new int[la.d];
            ba.ac = new String[la.d];
            fb.c = new int[la.d];
            p.e = new String[la.d];
            da.T = new int[la.d];
            e.Mb = new String[la.d];
            na.a = new int[la.d];
            db.j = new int[la.d];
            var24 = la.d;
            var29 = 12;
            break;
         }

         qb.d = new int[var24][var29];
         eb.b = new int[la.d];
         ua.Ab = new int[la.d];
         var3 = 0;

         while (true) {
            if (~la.d < ~var3) {
               e.Mb[var3] = o.a((byte)38);
               var3++;
               if (var6) {
                  break;
               }

               if (!var6) {
                  continue;
               }
            }

            var3 = 0;
            break;
         }

         while (true) {
            if (~var3 > ~la.d) {
               ba.ac[var3] = o.a((byte)38);
               var3++;
               if (var6) {
                  break;
               }

               if (!var6) {
                  continue;
               }
            }

            var3 = 0;
            break;
         }

         while (true) {
            if (la.d > var3) {
               la.a[var3] = v.a(-30504);
               var3++;
               if (var6) {
                  break;
               }

               if (!var6) {
                  continue;
               }
            }

            var3 = 0;
            break;
         }

         while (true) {
            if (~la.d < ~var3) {
               eb.b[var3] = v.a(-30504);
               var3++;
               if (var6) {
                  break;
               }

               if (!var6) {
                  continue;
               }
            }

            var3 = 0;
            break;
         }

         while (true) {
            if (~la.d < ~var3) {
               fb.d[var3] = v.a(-30504);
               var3++;
               if (var6) {
                  break;
               }

               if (!var6) {
                  continue;
               }
            }

            var3 = 0;
            break;
         }

         while (true) {
            if (~var3 > ~la.d) {
               jb.k[var3] = v.a(-30504);
               var3++;
               if (var6) {
                  break;
               }

               if (!var6) {
                  continue;
               }
            }

            var3 = 0;
            break;
         }

         while (true) {
            if (la.d > var3) {
               o.a[var3] = v.a(-30504);
               var3++;
               if (var6) {
                  break;
               }

               if (!var6) {
                  continue;
               }
            }

            var3 = 0;
            break;
         }

         label1080:
         while (true) {
            var24 = ~la.d;
            var29 = ~var3;

            label1077:
            while (var24 < var29) {
               var26 = 0;
               if (var6) {
                  break label1080;
               }

               int var4 = 0;

               while (var4 < 12) {
                  qb.d[var3][var4] = v.a(-30504);
                  var24 = qb.d[var3][var4];
                  var29 = 255;
                  if (var6) {
                     continue label1077;
                  }

                  if (var24 == 255) {
                     qb.d[var3][var4] = -1;
                  }

                  var4++;
                  if (var6) {
                     break;
                  }
               }

               var3++;
               if (!var6) {
                  continue label1080;
               }
               break;
            }

            var26 = 0;
            break;
         }

         var3 = var26;

         while (true) {
            if (var3 < la.d) {
               da.T[var3] = ub.a((byte)-105);
               var3++;
               if (var6) {
                  break;
               }

               if (!var6) {
                  continue;
               }
            }

            var3 = 0;
            break;
         }

         while (true) {
            if (~var3 > ~la.d) {
               g[var3] = ub.a((byte)-105);
               var3++;
               if (var6) {
                  break;
               }

               if (!var6) {
                  continue;
               }
            }

            var3 = 0;
            break;
         }

         while (true) {
            if (~var3 > ~la.d) {
               ua.Ab[var3] = ub.a((byte)-105);
               var3++;
               if (var6) {
                  break;
               }

               if (!var6) {
                  continue;
               }
            }

            if (var1 < 10) {
               return;
            }
            break;
         }

         var3 = 0;

         while (true) {
            if (la.d > var3) {
               v.e[var3] = ub.a((byte)-105);
               var3++;
               if (var6) {
                  break;
               }

               if (!var6) {
                  continue;
               }
            }

            var3 = 0;
            break;
         }

         while (true) {
            if (~var3 > ~la.d) {
               fb.c[var3] = t.a(65525);
               var3++;
               if (var6) {
                  break;
               }

               if (!var6) {
                  continue;
               }
            }

            var3 = 0;
            break;
         }

         while (true) {
            if (~var3 > ~la.d) {
               b.h[var3] = t.a(65525);
               var3++;
               if (var6) {
                  break;
               }

               if (!var6) {
                  continue;
               }
            }

            var3 = 0;
            break;
         }

         while (true) {
            if (~var3 > ~la.d) {
               ob.h[var3] = v.a(-30504);
               var3++;
               if (var6) {
                  break;
               }

               if (!var6) {
                  continue;
               }
            }

            var3 = 0;
            break;
         }

         while (true) {
            if (~var3 > ~la.d) {
               na.a[var3] = v.a(-30504);
               var3++;
               if (var6) {
                  break;
               }

               if (!var6) {
                  continue;
               }
            }

            var3 = 0;
            break;
         }

         while (true) {
            if (var3 < la.d) {
               db.j[var3] = v.a(-30504);
               var3++;
               if (var6) {
                  break;
               }

               if (!var6) {
                  continue;
               }
            }

            var3 = 0;
            break;
         }

         while (true) {
            if (var3 < la.d) {
               p.e[var3] = o.a((byte)38);
               var3++;
               if (var6) {
                  break;
               }

               if (!var6) {
                  continue;
               }
            }

            jb.o = t.a(65525);
            p.c = new String[jb.o];
            mb.g = new String[jb.o];
            break;
         }

         var3 = 0;

         while (true) {
            if (var3 < jb.o) {
               mb.g[var3] = o.a((byte)38);
               var3++;
               if (var6) {
                  break;
               }

               if (!var6) {
                  continue;
               }
            }

            var3 = 0;
            break;
         }

         while (true) {
            if (jb.o > var3) {
               p.c[var3] = o.a((byte)38);
               var3++;
               if (var6) {
                  break;
               }

               if (!var6) {
                  continue;
               }
            }

            na.e = t.a(65525);
            aa.c = new int[na.e];
            cb.e = new String[na.e];
            nb.d = new int[na.e];
            n.m = new int[na.e];
            w.g = new int[na.e];
            db.l = new int[na.e];
            break;
         }

         var3 = 0;

         while (true) {
            if (var3 < na.e) {
               cb.e[var3] = o.a((byte)38);
               var3++;
               if (var6) {
                  break;
               }

               if (!var6) {
                  continue;
               }
            }

            var3 = 0;
            break;
         }

         while (true) {
            if (~var3 > ~na.e) {
               db.l[var3] = ub.a((byte)-105);
               var3++;
               if (var6) {
                  break;
               }

               if (!var6) {
                  continue;
               }
            }

            var3 = 0;
            break;
         }

         while (true) {
            if (~na.e < ~var3) {
               n.m[var3] = v.a(-30504);
               var3++;
               if (var6) {
                  break;
               }

               if (!var6) {
                  continue;
               }
            }

            var3 = 0;
            break;
         }

         while (true) {
            if (var3 < na.e) {
               nb.d[var3] = v.a(-30504);
               var3++;
               if (var6) {
                  break;
               }

               if (!var6) {
                  continue;
               }
            }

            var3 = 0;
            break;
         }

         while (true) {
            if (~var3 > ~na.e) {
               aa.c[var3] = v.a(-30504);
               var3++;
               if (var6) {
                  break;
               }

               if (!var6) {
                  continue;
               }
            }

            var3 = 0;
            break;
         }

         while (true) {
            if (na.e > var3) {
               w.g[var3] = v.a(-30504);
               var3++;
               if (var6) {
                  break;
               }

               if (!var6) {
                  continue;
               }
            }

            ua.Db = t.a(65525);
            mb.a = new int[ua.Db];
            ub.g = new int[ua.Db];
            la.f = new String[ua.Db];
            l.a = new String[ua.Db];
            f.f = new int[ua.Db];
            p.a = new String[ua.Db];
            s.f = new String[ua.Db];
            fb.f = new int[ua.Db];
            h.b = new int[ua.Db];
            break;
         }

         var3 = 0;

         while (true) {
            if (~ua.Db < ~var3) {
               l.a[var3] = o.a((byte)38);
               var3++;
               if (var6) {
                  break;
               }

               if (!var6) {
                  continue;
               }
            }

            var3 = 0;
            break;
         }

         while (true) {
            if (var3 < ua.Db) {
               la.f[var3] = o.a((byte)38);
               var3++;
               if (var6) {
                  break;
               }

               if (!var6) {
                  continue;
               }
            }

            var3 = 0;
            break;
         }

         while (true) {
            if (~var3 > ~ua.Db) {
               s.f[var3] = o.a((byte)38);
               var3++;
               if (var6) {
                  break;
               }

               if (!var6) {
                  continue;
               }
            }

            var3 = 0;
            break;
         }

         while (true) {
            if (~ua.Db < ~var3) {
               p.a[var3] = o.a((byte)38);
               var3++;
               if (var6) {
                  break;
               }

               if (!var6) {
                  continue;
               }
            }

            var3 = 0;
            break;
         }

         while (true) {
            if (var3 < ua.Db) {
               fb.f[var3] = ca.a((byte)91, o.a((byte)38));
               var3++;
               if (var6) {
                  break;
               }

               if (!var6) {
                  continue;
               }
            }

            var3 = 0;
            break;
         }

         while (true) {
            if (~var3 > ~ua.Db) {
               f.f[var3] = v.a(-30504);
               var3++;
               if (var6) {
                  break;
               }

               if (!var6) {
                  continue;
               }
            }

            var3 = 0;
            break;
         }

         while (true) {
            if (~ua.Db < ~var3) {
               ub.g[var3] = v.a(-30504);
               var3++;
               if (var6) {
                  break;
               }

               if (!var6) {
                  continue;
               }
            }

            var3 = 0;
            break;
         }

         while (true) {
            if (~ua.Db < ~var3) {
               mb.a[var3] = v.a(-30504);
               var3++;
               if (var6) {
                  break;
               }

               if (!var6) {
                  continue;
               }
            }

            var3 = 0;
            break;
         }

         while (true) {
            if (~var3 > ~ua.Db) {
               h.b[var3] = v.a(-30504);
               var3++;
               if (var6) {
                  break;
               }

               if (!var6) {
                  continue;
               }
            }

            h.a = t.a(65525);
            u.b = new String[h.a];
            u.a = new int[h.a];
            v.a = new int[h.a];
            f.e = new String[h.a];
            client.Jk = new int[h.a];
            lb.Tb = new int[h.a];
            ub.b = new String[h.a];
            ta.r = new String[h.a];
            ib.d = new int[h.a];
            break;
         }

         var3 = 0;

         while (true) {
            if (h.a > var3) {
               ta.r[var3] = o.a((byte)38);
               var3++;
               if (var6) {
                  break;
               }

               if (!var6) {
                  continue;
               }
            }

            var3 = 0;
            break;
         }

         while (true) {
            if (~var3 > ~h.a) {
               ub.b[var3] = o.a((byte)38);
               var3++;
               if (var6) {
                  break;
               }

               if (!var6) {
                  continue;
               }
            }

            var3 = 0;
            break;
         }

         while (true) {
            if (~h.a < ~var3) {
               u.b[var3] = o.a((byte)38);
               var3++;
               if (var6) {
                  break;
               }

               if (!var6) {
                  continue;
               }
            }

            var3 = 0;
            break;
         }

         while (true) {
            if (~h.a < ~var3) {
               f.e[var3] = o.a((byte)38);
               var3++;
               if (var6) {
                  break;
               }

               if (!var6) {
                  continue;
               }
            }

            var3 = 0;
            break;
         }

         while (true) {
            if (var3 < h.a) {
               ib.d[var3] = t.a(65525);
               var3++;
               if (var6) {
                  break;
               }

               if (!var6) {
                  continue;
               }
            }

            var3 = 0;
            break;
         }

         while (true) {
            if (~h.a < ~var3) {
               v.a[var3] = ub.a((byte)-105);
               var3++;
               if (var6) {
                  break;
               }

               if (!var6) {
                  continue;
               }
            }

            var3 = 0;
            break;
         }

         while (true) {
            if (var3 < h.a) {
               client.Jk[var3] = ub.a((byte)-105);
               var3++;
               if (var6) {
                  break;
               }

               if (!var6) {
                  continue;
               }
            }

            var3 = 0;
            break;
         }

         while (true) {
            if (~h.a < ~var3) {
               u.a[var3] = v.a(-30504);
               var3++;
               if (var6) {
                  break;
               }

               if (!var6) {
                  continue;
               }
            }

            var3 = 0;
            break;
         }

         while (true) {
            if (~var3 > ~h.a) {
               lb.Tb[var3] = v.a(-30504);
               var3++;
               if (var6) {
                  break;
               }

               if (!var6) {
                  continue;
               }
            }

            v.h = t.a(65525);
            d.g = new int[v.h];
            i.g = new int[v.h];
            break;
         }

         var3 = 0;

         while (true) {
            if (~var3 > ~v.h) {
               i.g[var3] = v.a(-30504);
               var3++;
               if (var6) {
                  break;
               }

               if (!var6) {
                  continue;
               }
            }

            var3 = 0;
            break;
         }

         while (true) {
            if (~v.h < ~var3) {
               d.g[var3] = v.a(-30504);
               var3++;
               if (var6) {
                  break;
               }

               if (!var6) {
                  continue;
               }
            }

            ia.h = t.a(65525);
            da.N = new int[ia.h];
            ac.l = new int[ia.h];
            qa.K = new int[ia.h];
            break;
         }

         var3 = 0;

         while (true) {
            if (ia.h > var3) {
               qa.K[var3] = ub.a((byte)-105);
               var3++;
               if (var6) {
                  break;
               }

               if (!var6) {
                  continue;
               }
            }

            var3 = 0;
            break;
         }

         while (true) {
            if (~var3 > ~ia.h) {
               da.N[var3] = v.a(-30504);
               var3++;
               if (var6) {
                  break;
               }

               if (!var6) {
                  continue;
               }
            }

            var3 = 0;
            break;
         }

         while (true) {
            if (var3 < ia.h) {
               ac.l[var3] = v.a(-30504);
               var3++;
               if (var6) {
                  break;
               }

               if (!var6) {
                  continue;
               }
            }

            n.c = t.a(65525);
            fa.b = t.a(65525);
            oa.d = new int[fa.b][];
            qb.e = new int[fa.b];
            da.J = new int[fa.b][];
            ja.L = new String[fa.b];
            oa.a = new String[fa.b];
            pa.f = new int[fa.b];
            o.p = new int[fa.b];
            break;
         }

         var3 = 0;

         while (true) {
            if (~var3 > ~fa.b) {
               ja.L[var3] = o.a((byte)38);
               var3++;
               if (var6) {
                  break;
               }

               if (!var6) {
                  continue;
               }
            }

            var3 = 0;
            break;
         }

         while (true) {
            if (fa.b > var3) {
               oa.a[var3] = o.a((byte)38);
               var3++;
               if (var6) {
                  break;
               }

               if (!var6) {
                  continue;
               }
            }

            var3 = 0;
            break;
         }

         while (true) {
            if (fa.b > var3) {
               pa.f[var3] = v.a(-30504);
               var3++;
               if (var6) {
                  break;
               }

               if (!var6) {
                  continue;
               }
            }

            var3 = 0;
            break;
         }

         while (true) {
            if (var3 < fa.b) {
               o.p[var3] = v.a(-30504);
               var3++;
               if (var6) {
                  break;
               }

               if (!var6) {
                  continue;
               }
            }

            var3 = 0;
            break;
         }

         while (true) {
            if (~fa.b < ~var3) {
               qb.e[var3] = v.a(-30504);
               var3++;
               if (var6) {
                  break;
               }

               if (!var6) {
                  continue;
               }
            }

            var3 = 0;
            break;
         }

         while (true) {
            if (var3 < fa.b) {
               int var21 = v.a(-30504);
               oa.d[var3] = new int[var21];
               var27 = 0;
               if (var6) {
                  break;
               }

               int var5 = 0;

               label676: {
                  while (~var21 < ~var5) {
                     oa.d[var3][var5] = t.a(65525);
                     var5++;
                     if (var6) {
                        break label676;
                     }

                     if (var6) {
                        break;
                     }
                  }

                  var3++;
               }

               if (!var6) {
                  continue;
               }
            }

            var27 = 0;
            break;
         }

         var3 = var27;

         while (true) {
            if (var3 < fa.b) {
               int var22 = v.a(-30504);
               da.J[var3] = new int[var22];
               var28 = 0;
               if (var6) {
                  break;
               }

               int var23 = 0;

               label656: {
                  while (~var23 > ~var22) {
                     da.J[var3][var23] = v.a(-30504);
                     var23++;
                     if (var6) {
                        break label656;
                     }

                     if (var6) {
                        break;
                     }
                  }

                  var3++;
               }

               if (!var6) {
                  continue;
               }
            }

            t.g = t.a(65525);
            t.h = new String[t.g];
            ca.B = new int[t.g];
            fa.c = new int[t.g];
            h.e = new String[t.g];
            var28 = 0;
            break;
         }

         var3 = var28;

         while (true) {
            if (var3 < t.g) {
               t.h[var3] = o.a((byte)38);
               var3++;
               if (var6) {
                  break;
               }

               if (!var6) {
                  continue;
               }
            }

            var3 = 0;
            break;
         }

         while (true) {
            if (var3 < t.g) {
               h.e[var3] = o.a((byte)38);
               var3++;
               if (var6) {
                  break;
               }

               if (!var6) {
                  continue;
               }
            }

            var3 = 0;
            break;
         }

         while (true) {
            if (t.g > var3) {
               ca.B[var3] = v.a(-30504);
               var3++;
               if (var6) {
                  break;
               }

               if (!var6) {
                  continue;
               }
            }

            var3 = 0;
            break;
         }

         while (true) {
            if (t.g > var3) {
               fa.c[var3] = v.a(-30504);
               var3++;
               if (var6) {
                  break;
               }

               if (!var6) {
                  continue;
               }
            }

            b.v = null;
            kb.d = null;
            break;
         }
      } catch (RuntimeException var7) {
         throw i.a(var7, z[3] + (var0 != null ? z[1] : z[0]) + ',' + var1 + ',' + var2 + ')');
      }
   }

   abstract Socket a(byte var1) throws IOException;

   private static char[] z(String var0) {
      char[] var10000 = var0.toCharArray();
      if (var10000.length < 2) {
         var10000[0] = (char)(var10000[0] ^ 94);
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
               var10005 = 64;
               break;
            case 1:
               var10005 = 20;
               break;
            case 2:
               var10005 = 39;
               break;
            case 3:
               var10005 = 57;
               break;
            default:
               var10005 = 94;
         }

         var10001[var1] = (char)(var10004 ^ var10005);
      }

      return new String(var10001).intern();
   }
}
