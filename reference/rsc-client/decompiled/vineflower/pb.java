import java.awt.Component;
import javax.sound.sampled.AudioFormat;
import javax.sound.sampled.AudioSystem;
import javax.sound.sampled.LineUnavailableException;
import javax.sound.sampled.SourceDataLine;
import javax.sound.sampled.DataLine.Info;

final class pb extends sa {
   private int x;
   private byte[] v;
   private AudioFormat y;
   private SourceDataLine w;
   private static final String z = z(z("}\u001a\u001fKqc\u0014\u0012"));

   @Override
   final void e() {
      if (this.w != null) {
         this.w.close();
         this.w = null;
      }
   }

   @Override
   final void c() {
      short var1 = 256;
      if (i) {
         var1 <<= 1;
      }

      for (int var2 = 0; var2 < var1; var2++) {
         int var3 = this.j[var2];
         if ((var3 + 8388608 & 0xFF000000) != 0) {
            var3 = 8388607 ^ var3 >> 31;
         }

         this.v[var2 * 2] = (byte)(var3 >> 8);
         this.v[var2 * 2 + 1] = (byte)(var3 >> 16);
      }

      this.w.write(this.v, 0, var1 << 1);
   }

   @Override
   final int b() {
      return this.x - (this.w.available() >> (i ? 2 : 1));
   }

   @Override
   final void b(int var1) throws LineUnavailableException {
      try {
         Class<SourceDataLine> var10002;
         AudioFormat var10003;
         int var10004;
         byte var10005;
         label32: {
            try {
               var10002 = SourceDataLine.class;
               var10003 = this.y;
               var10004 = var1;
               if (i) {
                  var10005 = 2;
                  break label32;
               }
            } catch (LineUnavailableException var4) {
               throw var4;
            }

            var10005 = 1;
         }

         Info var10000 = new Info(var10002, var10003, var10004 << var10005);
         Info var2 = var10000;
         this.w = (SourceDataLine)AudioSystem.getLine(var2);
         this.w.open();
         this.w.start();
         this.x = var1;
      } catch (LineUnavailableException var5) {
         try {
            if (ta.a(var1, (byte)-59) != 1) {
               this.b(aa.a(var1, false));
               return;
            }
         } catch (LineUnavailableException var3) {
            throw var3;
         }

         this.w = null;
         throw var5;
      }
   }

   @Override
   final void a(Component var1) {
      javax.sound.sampled.Mixer.Info[] var2 = AudioSystem.getMixerInfo();
      if (var2 != null) {
         javax.sound.sampled.Mixer.Info[] var3 = var2;

         for (int var4 = 0; var4 < var3.length; var4++) {
            javax.sound.sampled.Mixer.Info var5 = var3[var4];
            if (var5 != null) {
               String var6 = var5.getName();
               if (var6 != null && var6.toLowerCase().indexOf(z) >= 0) {
               }
            }
         }
      }

      this.y = new AudioFormat((float)t, 16, i ? 2 : 1, true, false);
      this.v = new byte[256 << (i ? 2 : 1)];
   }

   private static char[] z(String var0) {
      char[] var10000 = var0.toCharArray();
      if (var10000.length < 2) {
         var10000[0] = (char)(var10000[0] ^ 21);
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
               var10005 = 14;
               break;
            case 1:
               var10005 = 117;
               break;
            case 2:
               var10005 = 106;
               break;
            case 3:
               var10005 = 37;
               break;
            default:
               var10005 = 21;
         }

         var10001[var1] = (char)(var10004 ^ var10005);
      }

      return new String(var10001).intern();
   }
}
