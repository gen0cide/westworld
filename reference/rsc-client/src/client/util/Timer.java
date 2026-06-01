package client.util;

import client.Mudclient;
import client.net.StreamBase;
import client.ui.MessageList;

/**
 * Monotonic-clock helper for the RuneScape Classic client (rev ~233-235,
 * Microsoft J++ build). Obfuscated name: {@code p}.
 *
 * <p>The client never reads {@link System#currentTimeMillis()} directly for
 * frame/poll timing. Instead it funnels every read through
 * {@link #currentTimeMillisCorrected()}, which detects backwards jumps of the
 * wall clock (e.g. the user moving the system clock, NTP corrections, or a
 * laptop resuming from sleep) and accumulates a compensating offset so that the
 * value handed back to the game loop is always non-decreasing. Without this the
 * frame-rate governor and timeout logic in {@code GameShell} / the audio
 * channel scheduler ({@code AudioChannel}) would misbehave when the clock moved
 * backwards.</p>
 *
 * <p>The accumulated correction offset lives in {@link MessageList#w}
 * (obf {@code wb.w}) and the last observed raw timestamp lives in
 * {@link Mudclient#ze} (obf {@code client.ze}); both are static so the
 * correction is shared process-wide.</p>
 *
 * <p>This class also carries a second, unrelated static method,
 * {@link #renderAffineSpan}: a fixed-point affine/textured horizontal-span
 * blitter that the J++ obfuscator relocated here from the rasteriser. It is
 * documented and named but is logically part of the scene renderer, not the
 * timer. The original (obfuscated) method signatures were "p.A(" and "p.B(",
 * preserved below in {@link #ERROR_SIGNATURES} for crash context strings.</p>
 */
public final class Timer {

   // --- Original obfuscation artifacts, kept for documentation ----------------

   /**
    * Unused legacy flags/tables carried in the obfuscated class. The obfuscator
    * scattered string-table and boolean members across many classes; none of
    * these are referenced by the live timing path. Names preserved verbatim
    * from the obfuscated layout ({@code d}, {@code e}, {@code c}, {@code a}).
    */
   public static boolean legacyFlag = true;          // obf: d
   public static String[] legacyStringsE;            // obf: e
   public static String[] legacyStringsC;            // obf: c
   public static String[] legacyStringsA;            // obf: a

   /** Per-method invocation counter for {@link #currentTimeMillisCorrected()} (profiling artifact; harmless). obf: b */
   public static int timeCallCount;                  // obf: b
   /** Per-method invocation counter for {@link #renderAffineSpan}      (profiling artifact; harmless). obf: f */
   public static int spanCallCount;                  // obf: f

   /**
    * Decoded crash-context signatures, indexed by use site:
    * [0]="p.A("  -> {@link #currentTimeMillisCorrected()}'s original name,
    * [1]="null", [2]="{...}" -> placeholders for null / non-null array args,
    * [3]="p.B("  -> {@link #renderAffineSpan}'s original name.
    * These are XOR-obfuscated at class-init via {@link #xorPass}/{@link #xorPositional}.
    * obf: z
    */
   private static final String[] ERROR_SIGNATURES = new String[]{
      xorPositional(xorPass("/ 3\t")),
      xorPositional(xorPass("1{M")),
      xorPositional(xorPass("$ \\")),
      xorPositional(xorPass("/ 0\t"))
   };

   /**
    * Returns {@link System#currentTimeMillis()} adjusted by an accumulated
    * offset that cancels out any backwards movement of the wall clock, so the
    * result is monotonically non-decreasing across calls.
    *
    * @param mustBeZero anti-tamper guard parameter: every real caller passes 0;
    *                   any non-zero value returns the sentinel -57 instead of a
    *                   real timestamp. Dead in practice.
    * @return corrected, non-decreasing milliseconds.
    */
   public static final synchronized long currentTimeMillisCorrected(int mustBeZero) {
      // Anti-tamper guard: all live callers pass 0. Non-zero => sentinel.
      if (mustBeZero != 0) {
         return -57L;
      }

      long now = System.currentTimeMillis();

      // The raw obfuscated test was (~lastObservedTime < ~now), i.e. the
      // bitwise complement of a signed-less-than. ~a < ~b  <=>  a > b, so this
      // fires when the clock moved BACKWARDS since the previous reading.
      if (Mudclient.ze > now) {
         // Add the amount we slipped backwards (lastObserved - now) to the
         // running correction so future results don't go back in time.
         MessageList.w += Mudclient.ze - now;
      }

      Mudclient.ze = now;                 // remember this raw reading
      return MessageList.w + now;         // raw time + accumulated correction
   }

   /**
    * Fixed-point affine textured horizontal-span blitter (obf: {@code p.B} /
    * second {@code p.a}). Misplaced into the timer class by the obfuscator; it
    * belongs to the software rasteriser. Walks {@code spanCount} pixels (in
    * groups of 16, with a remainder loop), sampling a texture and writing
    * shaded ARGB-ish texels into the destination scanline.
    *
    * <p>Fixed-point conventions (recovered from the bytecode):
    * <ul>
    *   <li>{@code u}/{@code v} texture coordinates are 12.x fixed-point; the
    *       texture is masked to a row stride of 64 via {@code & 4032} (0xFC0)
    *       and a column of 0xFFF.</li>
    *   <li>{@code lightPacked} carries the per-pixel shade in its high bits
    *       ({@code 0xC0000} group) and the texture row in its low 12 bits;
    *       {@code shade = lightPacked >> ...} selects an unsigned right-shift
    *       amount used to dim the sampled texel ({@code texel >>> shade}).</li>
    * </ul>
    * The many large shift constants in the bytecode are taken mod 32 by the JVM;
    * they encode small shifts (typically 0..18) once normalised.</p>
    *
    * @param uStep        per-pixel U increment (pre-scaled by left-shift 2). obf param0 (n2)
    * @param texModeFlag  selects setup path; the sole live value short-circuits
    *                     a recursive self-call guard. obf param1 (n3)
    * @param uGradient    U gradient added each scanline group. obf param2 (n4)
    * @param uStart       starting U numerator. obf param3 (n5)
    * @param vGradient    V gradient added each scanline group. obf param4 (n6)
    * @param texture      source texture pixels. obf param5 (nArray)
    * @param lightPacked  packed shade + texture-row accumulator. obf param6 (n7)
    * @param u            current U fixed-point coordinate. obf param7 (n8)
    * @param vStart       starting V numerator. obf param8 (n9)
    * @param v            current V fixed-point coordinate. obf param9 (n10)
    * @param dest         destination scanline pixels. obf param10 (nArray2)
    * @param destIndex    write cursor into {@code dest}. obf param11 (n11)
    * @param depth        perspective divisor (0 => skip divide). obf param12 (n12)
    * @param depthStep    per-group divisor increment. obf param13 (n13)
    * @param spanCount    number of pixels to emit. obf param14 (n14)
    */
   public static final void renderAffineSpan(
      int uStep,
      int texModeFlag,
      int uGradient,
      int uStart,
      int vGradient,
      int[] texture,
      int lightPacked,
      int u,
      int vStart,
      int v,
      int[] dest,
      int destIndex,
      int depth,
      int depthStep,
      int spanCount
   ) {
      if (spanCount <= 0) {
         return;
      }

      // --- Initial perspective-correct U/V derivatives for this span ---------
      int vCoord = 0;
      int uCoord = 0;
      if (depth != 0) {
         // Shift constants are reduced mod 32 by the JVM; encode small shifts.
         uCoord = uStart / depth << (-258111322 & 31);
         vCoord = vStart / depth << (637317126 & 31);
      }
      uStep <<= 2;

      // Clamp the V coordinate into the valid texture-row range [0, 4032].
      if (~vCoord <= -1) {                       // i.e. vCoord >= 0
         if (vCoord > 4032) {
            vCoord = 4032;
         }
      } else {
         vCoord = 0;
      }

      // Self-recursion guard left by the obfuscator: never taken for live data
      // (texModeFlag is always 1121159302 on the real call path). Kept faithful
      // but it would only re-enter with all-sentinel arguments.
      if (texModeFlag != 1121159302) {
         renderAffineSpan(-69, 127, -20, -29, -78, null, 16, 2, -77, -5, null, 113, -57, 68, -87);
      }

      // --- Outer loop: process the span 16 pixels at a time ------------------
      int remaining = spanCount;
      do {
         int processed = 0;
         int limit = remaining;

         while (true) {
            // Advance the per-group gradients and reset the working coords.
            depth += depthStep;
            vStart += vGradient;
            uStart += uGradient;
            v = vCoord;
            u = uCoord;

            if (~depth != -1) {                  // depth != 0
               vCoord = vStart / depth << (25779686 & 31);
               uCoord = uStart / depth << (1121159302 & 31);
            }
            // Re-clamp vCoord (the V coordinate) into range.
            if (~vCoord >= -1) {                 // vCoord <= 0
               vCoord = 0;
            } else if (~vCoord < -4033) {        // vCoord > 4032
               vCoord = 4032;
            }

            int uInc = uCoord + -u >> (1397627332 & 31);   // per-pixel U step
            int vInc = -v + vCoord >> (1542826468 & 31);   // per-pixel V step
            int shade = lightPacked >> (-1668610924 & 31); // brightness shift
            v += 0xC0000 & lightPacked;          // carry shade group into V
            lightPacked += uStep;

            if (~remaining <= -17) {             // remaining >= 16: full block
               // 16 unrolled texel writes. Each: sample texture at
               // (V-row | U-col) masked to the 64-wide texture, dim by `shade`.
               dest[destIndex++] = texture[StreamBase.and(u, 4032) - -(v >> (-1148525818 & 31))] >>> shade;
               dest[destIndex++] = texture[((v += vInc) >> (1034190278 & 31)) + StreamBase.and(u += uInc, 4032)] >>> shade;
               dest[destIndex++] = texture[((v += vInc) >> (-385010618 & 31)) + StreamBase.and(4032, u += uInc)] >>> shade;
               dest[destIndex++] = texture[((v += vInc) >> (747209702 & 31)) + StreamBase.and(4032, u += uInc)] >>> shade;
               v += vInc;
               shade = lightPacked >> (597207284 & 31);
               v = (lightPacked & 0xC0000) + (0xFFF & v);
               lightPacked += uStep;

               dest[destIndex++] = texture[StreamBase.and(4032, u += uInc) + (v >> (831423910 & 31))] >>> shade;
               dest[destIndex++] = texture[StreamBase.and(u += uInc, 4032) - -((v += vInc) >> (-512409978 & 31))] >>> shade;
               dest[destIndex++] = texture[StreamBase.and(u += uInc, 4032) + ((v += vInc) >> (-783757370 & 31))] >>> shade;
               dest[destIndex++] = texture[((v += vInc) >> (-129948154 & 31)) + StreamBase.and(4032, u += uInc)] >>> shade;
               v += vInc;
               shade = lightPacked >> (92466196 & 31);
               v = (0xC0000 & lightPacked) + (0xFFF & v);
               lightPacked += uStep;

               dest[destIndex++] = texture[StreamBase.and(u += uInc, 4032) - -(v >> (-1989449594 & 31))] >>> shade;
               dest[destIndex++] = texture[((v += vInc) >> (-76155226 & 31)) + StreamBase.and(u += uInc, 4032)] >>> shade;
               dest[destIndex++] = texture[StreamBase.and(u += uInc, 4032) - -((v += vInc) >> (-158732986 & 31))] >>> shade;
               dest[destIndex++] = texture[StreamBase.and(4032, u += uInc) + ((v += vInc) >> (1960099526 & 31))] >>> shade;
               v += vInc;
               shade = lightPacked >> (1740031764 & 31);
               v = (0xFFF & v) - -(lightPacked & 0xC0000);
               lightPacked += uStep;

               dest[destIndex++] = texture[StreamBase.and(u += uInc, 4032) + (v >> (-1366214458 & 31))] >>> shade;
               dest[destIndex++] = texture[((v += vInc) >> (-1971655962 & 31)) + StreamBase.and(u += uInc, 4032)] >>> shade;
               dest[destIndex++] = texture[((v += vInc) >> (-674692218 & 31)) + StreamBase.and(u += uInc, 4032)] >>> shade;
               dest[destIndex++] = texture[StreamBase.and(4032, u += uInc) - -((v += vInc) >> (-1886734874 & 31))] >>> shade;
               break;                            // fall through to remainder handling
            }

            // --- Remainder loop: fewer than 16 pixels left ------------------
            for (int i = 0; i < remaining; i++) {
               dest[destIndex++] = texture[(v >> (-1102323802 & 31)) + StreamBase.and(4032, u)] >>> shade;
               u += uInc;
               v += vInc;
               // Every 4th pixel, advance the shade group and refresh `shade`.
               processed = 3;
               if (processed == (3 & i)) {
                  shade = lightPacked >> (-667502188 & 31);
                  v = (lightPacked & 0xC0000) + (0xFFF & v);
                  lightPacked += uStep;
               }
            }
            break;
         }

         remaining -= 16;
      } while (remaining > 0 || ~remaining <= -17);
      // NOTE: the do/while above mirrors the obfuscated control flow; the
      // group loop processes 16 pixels per iteration and the inner remainder
      // loop drains whatever is left once fewer than 16 remain. The outer
      // condition keeps iterating while there are full groups remaining.
   }

   private Timer() {
   }

   /**
    * First XOR stage of the static-string deobfuscator. Only flips the first
    * char (with key 'r') for single-char inputs; multi-char inputs pass through
    * unchanged. obf: z(String)
    *
    * @param encoded source string.
    * @return mutable char[] for {@link #xorPositional}.
    */
   private static char[] xorPass(String encoded) {
      char[] chars = encoded.toCharArray();
      if (chars.length < 2) {
         chars[0] = (char) (chars[0] ^ 'r');
      }
      return chars;
   }

   /**
    * Second XOR stage: decodes each char against a position-keyed pad
    * {95, 14, 114, 33, 114} cycling every 5 chars, then interns the result.
    * obf: z(char[])
    *
    * @param chars characters from {@link #xorPass}.
    * @return decoded, interned string.
    */
   private static String xorPositional(char[] chars) {
      for (int i = 0; i < chars.length; i++) {
         char c = chars[i];
         byte key;
         switch (i % 5) {
            case 0:
               key = 95;
               break;
            case 1:
               key = 14;
               break;
            case 2:
               key = 114;
               break;
            case 3:
               key = 33;
               break;
            default:
               key = 114;
         }
         chars[i] = (char) (c ^ key);
      }
      return new String(chars).intern();
   }
}
