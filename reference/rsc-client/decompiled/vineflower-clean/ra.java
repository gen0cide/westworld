final class ra extends va {
   private db n = new db();
   private db l = new db();
   private int m = 0;
   private int k = -1;

   @Override
   final va b() {
      return (va)this.n.a((byte)35);
   }

   final void a(vb var1, int var2, int var3) {
      this.a(sb.a(var1, var2, var3));
   }

   private final void e() {
      if (this.m > 0) {
         for (hb var1 = (hb)this.l.a((byte)-87); var1 != null; var1 = (hb)this.l.b((byte)111)) {
            var1.g = var1.g - this.m;
         }

         this.k = this.k - this.m;
         this.m = 0;
      }
   }

   @Override
   final synchronized void b(int[] var1, int var2, int var3) {
      while (this.k >= 0) {
         if (this.m + var3 < this.k) {
            this.m += var3;
            this.c(var1, var2, var3);
            return;
         }

         int var4 = this.k - this.m;
         this.c(var1, var2, var4);
         var2 += var4;
         var3 -= var4;
         this.m += var4;
         this.e();
         hb var5 = (hb)this.l.a((byte)106);
         synchronized (var5) {
            int var7 = var5.a(this);
            if (var7 < 0) {
               var5.g = 0;
               this.a(var5);
            } else {
               var5.g = var7;
               this.a(var5.a, var5);
            }
         }

         if (var3 == 0) {
            return;
         }
      }

      this.c(var1, var2, var3);
   }

   private final void a(ib var1, hb var2) {
      while (var1 != this.l.k && ((hb)var1).g <= var2.g) {
         var1 = var1.a;
      }

      ac.a(var2, (byte)34, var1);
      this.k = ((hb)this.l.k.a).g;
   }

   @Override
   final int d() {
      return 0;
   }

   @Override
   final synchronized void b(int var1) {
      while (this.k >= 0) {
         if (this.m + var1 < this.k) {
            this.m += var1;
            this.c(var1);
            return;
         }

         int var2 = this.k - this.m;
         this.c(var2);
         var1 -= var2;
         this.m += var2;
         this.e();
         hb var3 = (hb)this.l.a((byte)-80);
         synchronized (var3) {
            int var5 = var3.a(this);
            if (var5 < 0) {
               var3.g = 0;
               this.a(var3);
            } else {
               var3.g = var5;
               this.a(var3.a, var3);
            }
         }

         if (var1 == 0) {
            return;
         }
      }

      this.c(var1);
   }

   private final void c(int[] var1, int var2, int var3) {
      for (va var4 = (va)this.n.a((byte)-88); var4 != null; var4 = (va)this.n.b((byte)122)) {
         var4.a(var1, var2, var3);
      }
   }

   @Override
   final va a() {
      return (va)this.n.b((byte)115);
   }

   private final synchronized void a(va var1) {
      this.n.a(var1, false);
   }

   private final void a(hb var1) {
      var1.a(-27331);
      var1.a();
      ib var2 = this.l.k.a;
      if (var2 == this.l.k) {
         this.k = -1;
      } else {
         this.k = ((hb)var2).g;
      }
   }

   private final void c(int var1) {
      for (va var2 = (va)this.n.a((byte)-123); var2 != null; var2 = (va)this.n.b((byte)80)) {
         var2.b(var1);
      }
   }

   public ra() {
   }
}
