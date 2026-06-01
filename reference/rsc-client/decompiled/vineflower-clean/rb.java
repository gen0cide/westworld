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

         for (int var1 = 0; var1 < 2; var1++) {
            this.b[var1] = new DSBufferDesc();
         }

         for (int var3 = 0; var3 < 2; var3++) {
            this.a[var3] = new DSCursors();
         }
      } catch (RuntimeException var2) {
         throw var2;
      }
   }
}
