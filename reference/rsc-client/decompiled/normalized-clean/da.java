import java.applet.Applet;
import java.io.IOException;
import java.io.InputStream;
import java.io.OutputStream;
import java.net.Socket;
import java.net.URL;

final class da extends b implements Runnable {
   static int[][] J;
   private byte[] Z = new byte[1];
   static fb db = new fb();
   private int W;
   private boolean fb;
   static int S;
   private int G = 0;
   static int[] T;
   private OutputStream Q;
   static Applet gb;
   static int R;
   static int P;
   static int eb;
   private InputStream U;
   static int M = 0;
   private Socket ab;
   private boolean X;
   static int V;
   static int cb;
   static int L;
   static int I;
   static int H;
   static v O = new v(z(z("\u0003v:;")), z(z(";D\u000e\u0011Q1")), z(z("\u000bP\u000b")), 1);
   private byte[] Y;
   static int[] N;
   static int bb;
   static int K;
   private static final String[] hb = new String[]{
      z(z("0CF?\u001a")),
      z(z("\u0011P\u001a\u0017@tA\u0004\u0017A=L\u000fXA P\r\u0019_")),
      z(z(":W\u0004\u0014")),
      z(z("/\fFVO")),
      z(z("t\u000fH")),
      z(z("0CF=\u001a")),
      z(z("0CF0\u001a")),
      z(z("\u0017M\u001d\u0014V:\u0005\u001cXV;U\u0006\u0014]5FH\u001e[8G")),
      z(z("6W\u000e\u001eW&\u0002\u0007\u000eW&D\u0004\u0017E")),
      z(z("0CF<\u001a")),
      z(z("0CF:\u001a")),
      z(z("0CFD[:K\u001cF\u001a")),
      z(z("\u0011m.")),
      z(z("0CF>\u001a")),
      z(z("0CF9\u001a")),
      z(z("0CF\nG:\nA")),
      z(z("\u0000U\u001a\u0011F1PR")),
      z(z("0CF;\u001a"))
   };

   @Override
   public final void run() {
      boolean var6 = client.vh;

      try {
         R++;

         do {
            boolean var10000;
            if (!this.X) {
               var10000 = true;
               if (!var6) {
               }
            } else {
               var10000 = false;
            }

            if (!var10000) {
               break;
            }

            int var1;
            int var2;
            synchronized (this) {
               if (this.G == this.W) {
                  try {
                     this.wait();
                  } catch (InterruptedException var9) {
                  }
               }

               if (this.X) {
                  return;
               }

               label71: {
                  if (this.W > this.G) {
                     var1 = 5000 - this.W;
                     if (!var6) {
                        break label71;
                     }
                  }

                  var1 = this.G + -this.W;
               }

               var2 = this.W;
            }

            if (-1 > ~var1) {
               try {
                  this.Q.write(this.Y, var2, var1);
               } catch (IOException var8) {
                  this.k = true;
                  this.B = hb[16] + var8;
               }

               this.W = (this.W - -var1) % 5000;

               try {
                  if (this.W == this.G) {
                     this.Q.flush();
                  }
               } catch (IOException var7) {
                  this.k = true;
                  this.B = hb[16] + var7;
               }
            }
         } while (!var6);
      } catch (RuntimeException var11) {
         throw i.a(var11, hb[15]);
      }
   }

   static final void a(String var0, int var1, int var2) {
      try {
         if (var2 != 0) {
            a((String)null, -126, -28);
         }

         I++;
         d.h.a(nb.q, (byte)-101, var0 + o.l + hb[4] + var1 + "%");
      } catch (RuntimeException var4) {
         throw i.a(var4, hb[5] + (var0 != null ? hb[3] : hb[2]) + ',' + var1 + ',' + var2 + ')');
      }
   }

   static final byte[] a(URL var0, boolean var1, boolean var2) throws IOException {
      boolean var5 = client.vh;

      try {
         L++;
         jb var3 = new jb(pa.k, var0, 2000000);
         if (var1) {
            a("", 0, 0);
         }

         while (!var3.a(-2)) {
            mb.a(11200, 50L);
            if (var5) {
               break;
            }
         }

         tb var4 = var3.a((byte)-120);
         if (var4 != null) {
            if (!var2) {
               a((String)null, -15, -97);
            }

            if (var1) {
               a("", 100, 0);
            }

            return var4.d(0);
         } else {
            throw new IOException(hb[7]);
         }
      } catch (RuntimeException var6) {
         throw i.a(var6, hb[6] + (var0 != null ? hb[3] : hb[2]) + ',' + var1 + ',' + var2 + ')');
      }
   }

   // $VF: Could not create synchronized statement, marking monitor enters and exits
   // Please report this to the Vineflower issue tracker, at https://github.com/Vineflower/vineflower/issues with a copy of the class file (if you have the rights to distribute it!)
   @Override
   final void a(byte[] var1, int var2, int var3, byte var4) throws IOException {
      boolean var8 = client.vh;

      try {
         if (var4 == -67) {
            S++;
            if (!this.fb) {
               if (null == this.Y) {
                  this.Y = new byte[5000];
               }

               da var5 = this;
               synchronized (this) {
                  int var6 = 0;

                  da var10000;
                  while (true) {
                     if (var6 < var3) {
                        this.Y[this.G] = var1[var2 + var6];
                        this.G = (this.G - -1) % 5000;
                        var10000 = this;
                        if (var8) {
                           break;
                        }

                        if (this.G == (4900 + this.W) % 5000) {
                           throw new IOException(hb[8]);
                        }

                        var6++;
                        if (!var8) {
                           continue;
                        }
                     }

                     this.notify();
                     var10000 = var5;
                     break;
                  }

                  // $VF: monitorexit
               }
            }
         }
      } catch (RuntimeException var10) {
         throw i.a(var10, hb[9] + (var1 != null ? hb[3] : hb[2]) + ',' + var2 + ',' + var3 + ',' + var4 + ')');
      }
   }

   @Override
   final int b(byte var1) throws IOException {
      try {
         H++;
         if (this.fb) {
            return 0;
         }

         int var2 = -127 % ((var1 - -64) / 56);
         return this.U.available();
      } catch (RuntimeException var3) {
         throw i.a(var3, hb[10] + var1 + 41);
      }
   }

   @Override
   final void a(boolean var1) {
      try {
         V++;
         super.a(true);
         this.fb = var1;

         try {
            if (this.U != null) {
               this.U.close();
            }

            if (null != this.Q) {
               this.Q.close();
            }

            if (null != this.ab) {
               this.ab.close();
            }
         } catch (IOException var5) {
            System.out.println(hb[1]);
         }

         this.X = true;
         synchronized (this) {
            this.notify();
         }

         this.Y = null;
      } catch (RuntimeException var6) {
         throw i.a(var6, hb[0] + var1 + ')');
      }
   }

   static final int a(int var0, byte var1, int var2, int var3) {
      try {
         if (var1 != -66) {
            K = 35;
         }

         cb++;
         return -(var3 / 8 * 32) + -1 + -(var2 / 8 * 1024) - var0 / 8;
      } catch (RuntimeException var5) {
         throw i.a(var5, hb[17] + var0 + 44 + var1 + 44 + var2 + 44 + var3 + 41);
      }
   }

   @Override
   final int b(boolean var1) throws IOException {
      try {
         P++;
         if (!this.fb == var1) {
            this.a(this.Z, 1, 0, 123);
            return 0xFF & this.Z[0];
         } else {
            return 0;
         }
      } catch (RuntimeException var3) {
         throw i.a(var3, hb[14] + var1 + 41);
      }
   }

   @Override
   final void a(byte[] var1, int var2, int var3, int var4) throws IOException {
      boolean var8 = client.vh;

      try {
         eb++;
         if (!this.fb) {
            int var6 = -81 / ((-25 - var4) / 35);
            int var5 = 0;
            int var7 = 0;

            while (var5 < var2 && !var8) {
               if (~(var7 = this.U.read(var1, var3 + var5, -var5 + var2)) >= -1) {
                  throw new IOException(hb[12]);
               }

               var5 += var7;
               if (var8) {
                  break;
               }
            }
         }
      } catch (RuntimeException var9) {
         throw i.a(var9, hb[13] + (var1 != null ? hb[3] : hb[2]) + ',' + var2 + ',' + var3 + ',' + var4 + ')');
      }
   }

   da(Socket var1, e var2) throws IOException {
      this.W = 0;
      this.fb = false;
      this.X = true;

      try {
         this.ab = var1;
         this.U = var1.getInputStream();
         this.Q = var1.getOutputStream();
         this.X = false;
         var2.a(1, this);
      } catch (RuntimeException var4) {
         throw i.a(var4, hb[11] + (var1 != null ? hb[3] : hb[2]) + ',' + (var2 != null ? hb[3] : hb[2]) + ')');
      }
   }

   private static char[] z(String var0) {
      char[] var10000 = var0.toCharArray();
      if (var10000.length < 2) {
         var10000[0] = (char)(var10000[0] ^ 50);
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
               var10005 = 84;
               break;
            case 1:
               var10005 = 34;
               break;
            case 2:
               var10005 = 104;
               break;
            case 3:
               var10005 = 120;
               break;
            default:
               var10005 = 50;
         }

         var10001[var1] = (char)(var10004 ^ var10005);
      }

      return new String(var10001).intern();
   }
}
