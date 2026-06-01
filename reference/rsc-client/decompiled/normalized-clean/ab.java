final class ab {
   static final void a(int[] var0, int var1, int var2) {
      var2 = var1 + var2 - 7;

      while (var1 < var2) {
         var0[var1++] = 0;
         var0[var1++] = 0;
         var0[var1++] = 0;
         var0[var1++] = 0;
         var0[var1++] = 0;
         var0[var1++] = 0;
         var0[var1++] = 0;
         var0[var1++] = 0;
      }

      var2 += 7;

      while (var1 < var2) {
         var0[var1++] = 0;
      }
   }

   static final void a(byte[] var0, int var1, byte[] var2, int var3, int var4) {
      if (var0 == var2) {
         if (var1 == var3) {
            return;
         }

         if (var3 > var1 && var3 < var1 + var4) {
            var1 += --var4;
            var3 += var4;
            var4 = var1 - var4;

            for (var4 += 7; var1 >= var4; var2[var3--] = var0[var1--]) {
               var2[var3--] = var0[var1--];
               var2[var3--] = var0[var1--];
               var2[var3--] = var0[var1--];
               var2[var3--] = var0[var1--];
               var2[var3--] = var0[var1--];
               var2[var3--] = var0[var1--];
               var2[var3--] = var0[var1--];
            }

            var4 -= 7;

            while (var1 >= var4) {
               var2[var3--] = var0[var1--];
            }

            return;
         }
      }

      var4 += var1;

      for (var4 -= 7; var1 < var4; var2[var3++] = var0[var1++]) {
         var2[var3++] = var0[var1++];
         var2[var3++] = var0[var1++];
         var2[var3++] = var0[var1++];
         var2[var3++] = var0[var1++];
         var2[var3++] = var0[var1++];
         var2[var3++] = var0[var1++];
         var2[var3++] = var0[var1++];
      }

      var4 += 7;

      while (var1 < var4) {
         var2[var3++] = var0[var1++];
      }
   }
}
