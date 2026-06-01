package client.scene;

import client.shell.GameShell;
import client.shell.LoaderThread;
import client.shell.Mudclient;
import client.util.ErrorHandler;
import client.util.Utility;
import java.awt.Component;
import java.awt.event.ActionEvent;

/**
 * Sprite-drawing extension of {@link Surface} (obfuscated {@code ba}, extends {@code ua}).
 *
 * <p>This is the rev ~233-235 J++ build's equivalent of RSC 204's {@code SurfaceSprite}.
 * It augments the base software renderer with a single dispatch entry point,
 * {@link #spriteClipping(int, int, int, int, int, int, byte, int)} (the entity-dispatch
 * override of {@link Surface}'s 8-arg {@code spriteClipping}), which decodes a "magic"
 * sprite id into the kind of game entity that should be rendered and calls back into
 * {@link Mudclient} to draw it.
 * The sprite id ranges encode the entity type:
 * <ul>
 *   <li>{@code id >= 50000} - teleport bubble  ({@code id - 50000})</li>
 *   <li>{@code id >= 40000} - ground/inventory item ({@code id - 40000})</li>
 *   <li>{@code id >= 20000} - NPC                ({@code id - 20000})</li>
 *   <li>{@code id >= 5000}  - player             ({@code id - 5000})</li>
 *   <li>otherwise           - a plain clipped sprite drawn by {@link Surface}</li>
 * </ul>
 *
 * <p>Besides rendering, this class is also the home of a handful of shared static
 * resources for the rest of the engine:
 * <ul>
 *   <li>{@link #sin2048Cache} - the global fixed-point sine/cosine lookup table
 *       (filled by {@code World}, read by {@code Mudclient}/{@code World}/{@code Scene}).</li>
 *   <li>{@link #recentMessages} - a rolling list of recently entered chat lines
 *       (managed by {@code Mudclient}).</li>
 *   <li>{@link #socialNames} - a string table populated from the social/friends stream.</li>
 *   <li>{@link #formatIpAddress(int, int)} - a packed-int to dotted-quad IP formatter.</li>
 *   <li>{@link #flushEventQueue(Object, int, LoaderThread)} - an AWT event-queue
 *       drain + synthetic {@link ActionEvent} poster used by the applet bootstrap.</li>
 * </ul>
 */
final class SurfaceSprite extends Surface {

   /** Back-reference to the owning game client; used to dispatch entity draws. */
   Mudclient client;

   /** Profiling counter for {@link #formatIpAddress(int, int)} (obfuscation artifact, unused logic). */
   static int formatIpCallCount;

   /** Profiling counter for {@link #flushEventQueue(Object, int, LoaderThread)} (obfuscation artifact). */
   static int flushEventQueueCallCount;

   /**
    * Global fixed-point sine/cosine table, indexed in 1/2048-of-a-turn units.
    * Entries {@code [0..1023]} hold {@code sin(i * 0.00613592315) * 32768} and entries
    * {@code [1024..2047]} hold {@code cos(i * 0.00613592315) * 32768} (i.e. cosine is the
    * sine table offset by +1024 / a quarter turn). The oracle (RSC 204) calls this
    * {@code sin2048Cache}. Populated by {@code World}; read with a {@code & 0x3FF} mask.
    */
   static int[] sin2048Cache = new int[2048];

   /** Rolling history of the most recently entered chat lines (newest at index 0). */
   static String[] recentMessages = new String[100];

   /** Profiling counter for the entity-dispatch {@link #spriteClipping(int, int, int, int, int, int, byte, int)} (obfuscation artifact). */
   static int drawEntitySpriteCallCount;

   /** Social/friends name table; (re)allocated and filled from the login/social stream. */
   static String[] socialNames;

   /**
    * Small pool of obfuscated string constants used only by the original error-context
    * wrappers and the synthetic AWT {@link ActionEvent} command. Decoded values are:
    * {@code {"dummy", "null", "{...}", "ba.A(", "ba.B(", "ba.C("}}.
    * Index 0 ("dummy") is the action command posted by {@link #flushEventQueue}.
    */
   private static final String[] OBF_STRINGS = new String[]{
      decode(decode("\t{")),   // "dummy"  - ActionEvent command
      decode(decode("")), // "null"   - error-context token
      decode(decode("HD_")),          // "{...}"  - error-context token
      decode(decode("D0*")),          // "ba.A("  - error-context prefix
      decode(decode("D3*")),          // "ba.B("  - error-context prefix
      decode(decode("D2*"))           // "ba.C("  - error-context prefix
   };

   /**
    * Formats a 32-bit packed IPv4 address into a dotted-quad string ("a.b.c.d").
    * The leading {@code dummyGuard} argument is an anti-tamper placeholder and is unused.
    */
   static final String formatIpAddress(int dummyGuard, int packedIp) {
      ++formatIpCallCount;
      // Big-endian extraction: byte 3 = high octet ... byte 0 = low octet.
      return (packedIp >> 24 & 0xFF) + "."
           + (packedIp >> 16 & 0xFF) + "."
           + (packedIp >> 8 & 0xFF) + "."
           + (packedIp & 0xFF);
   }

   /** Constructs the sprite surface with the given pixel dimensions and AWT host component. */
   SurfaceSprite(int width, int height, int spriteCount, Component component) {
      super(width, height, spriteCount, component);
   }

   /**
    * Renders one game entity by decoding {@code spriteId} into an entity kind and dispatching
    * the draw back into {@link Mudclient}. This mirrors RSC 204's
    * {@code SurfaceSprite.spriteClipping(x, y, w, h, id, tx, ty)}.
    *
    * @param tx        translate/anchor X passed through to the entity draw
    * @param spriteId  magic id encoding entity type + index (see class doc)
    * @param width     destination width
    * @param x         destination X
    * @param y         destination Y
    * @param height    destination height
    * @param flags     per-entity render flags (e.g. controls whether {@link #socialNames}
    *                  is cleared; value 29 suppresses that clear)
    * @param ty        translate/anchor Y passed through to the entity draw
    */
   // obf: a(int,int,int,int,int,int,byte,int) — overrides Surface's 8-arg spriteClipping
   //      (the oracle's SurfaceSprite.spriteClipping entity-dispatch entry point).
   @Override
   final void spriteClipping(int tx, int spriteId, int width, int x, int y, int height, byte flags, int ty) {
      ++drawEntitySpriteCallCount;
      if (spriteId >= 50000) {
         // Teleport bubble. Oracle: drawTeleportBubble(x, y, w, h, id - 50000, tx, ty).
         this.client.drawTeleportBubble(ty, x, y, width, spriteId - 50000, height, 2);
      } else if (spriteId >= 40000) {
         // Ground / inventory item. Oracle: drawItem(x, y, w, h, id - 40000, tx, ty).
         this.client.drawItem(width, ty, x, spriteId - 40000, height, -122, y);
      } else if (spriteId >= 20000) {
         // NPC. Oracle: drawNpc(x, y, w, h, id - 20000, tx, ty).
         this.client.drawNpc(y, ty, 105, width, tx, spriteId - 20000, height, x);
      } else if (spriteId >= 5000) {
         // Player. Oracle: drawPlayer(x, y, w, h, id - 5000, tx, ty).
         this.client.drawPlayer(ty, height, flags ^ 9, tx, x, y, width, spriteId - 5000);
      } else {
         // Plain sprite: clip and blit via the base Surface. The 5924 is an anti-tamper
         // dummy occupying the colour slot of Surface's 6-arg clipping overload.
         super.spriteClipping(x, y, width, height, 5924, spriteId);
      }

      // Original code clears the cached social name table unless flags == 29.
      if (flags != 29) {
         socialNames = null;
      }
   }

   /**
    * Drains the host AWT event queue and (optionally) posts a synthetic action event.
    *
    * <p>Spins the system {@link java.awt.EventQueue} up to 50 times, peeking events and
    * sleeping 1ms between iterations, stopping early once the queue is empty. If
    * {@code postAction == 1} and {@code target} is non-null, an {@link ActionEvent}
    * ({@code ACTION_PERFORMED}, command {@code "dummy"}) is posted to {@code target}.
    * Used by the applet/loader bootstrap to pump pending UI events. Any exception during
    * the post is swallowed.
    *
    * @param target      event source / post target (may be null)
    * @param postAction  if {@code 1}, post the synthetic action event after draining
    * @param loader      loader thread holding the cached system event queue
    */
   static final void flushEventQueue(Object target, int postAction, LoaderThread loader) {
      ++flushEventQueueCallCount;
      if (loader.systemEventQueue == null) {
         return;
      }
      // Pump up to 50 queued events (the original `while (var3 < 50)` with a 1ms yield).
      for (int i = 0; i < 50; i++) {
         if (loader.systemEventQueue.peekEvent() == null) {
            break;
         }
         Utility.sleep(11200, 1L); // first arg is an anti-tamper dummy; sleeps 1ms.
      }
      if (postAction != 1) {
         return;
      }
      try {
         if (target != null) {
            loader.systemEventQueue.postEvent(new ActionEvent(target, ActionEvent.ACTION_PERFORMED, OBF_STRINGS[0]));
         }
      } catch (Exception ignored) {
         // Original swallows any post failure.
      }
   }

   /**
    * First stage of the two-pass obfuscated string decode: returns the char array of the
    * literal, flipping bit 1 of the only char for length-1 strings (a length-dependent quirk
    * the original relies on so the round trip lands on the intended characters).
    */
   private static char[] decode(String s) {
      char[] chars = s.toCharArray();
      if (chars.length < 2) {
         chars[0] = (char)(chars[0] ^ 2);
      }
      return chars;
   }

   /**
    * Second stage of the obfuscated string decode: XORs each char with a position-dependent
    * key cycling every 5 chars ({@code 109, 102, 106, 113, 2}), then interns the result.
    */
   private static String decode(char[] chars) {
      for (int i = 0; i < chars.length; i++) {
         byte key;
         switch (i % 5) {
            case 0:
               key = 109;
               break;
            case 1:
               key = 102;
               break;
            case 2:
               key = 106;
               break;
            case 3:
               key = 113;
               break;
            default:
               key = 2;
         }
         chars[i] = (char)(chars[i] ^ key);
      }
      return new String(chars).intern();
   }
}
