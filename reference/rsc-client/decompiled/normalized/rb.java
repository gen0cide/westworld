import com.ms.directX.DSBufferDesc;
import com.ms.directX.DSCursors;
import com.ms.directX.DirectSound;
import com.ms.directX.WaveFormatEx;

final class rb implements ma {
   private DSBufferDesc[] b = new DSBufferDesc[2];
   private DSCursors[] a = new DSCursors[2];

   public rb() throws Exception {
      try {
         new DirectSound();
         new WaveFormatEx();
         int var1 = 0;

         try {
            while (var1 < 2) {
               this.b[var1] = new DSBufferDesc();
               var1++;
            }
         } catch (RuntimeException var3) {
            throw var3;
         }

         var1 = 0;

         try {
            while (var1 < 2) {
               this.a[var1] = new DSCursors();
               var1++;
            }
         } catch (RuntimeException var2) {
            throw var2;
         }
      } catch (RuntimeException var4) {
         throw var4;
      }
   }
}
