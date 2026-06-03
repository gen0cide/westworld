abstract class va extends ib {
   va j;
   bb h;
   int i;
   volatile boolean g = true;

   abstract void b(int var1);

   abstract int d();

   abstract va a();

   abstract void b(int[] var1, int var2, int var3);

   abstract va b();

   int c() {
      return 255;
   }

   final void a(int[] var1, int var2, int var3) {
      if (this.g) {
         this.b(var1, var2, var3);
      } else {
         this.b(var3);
      }
   }
}
