import com.ms.dll.Callback;
import com.ms.win32.User32;
import java.awt.Component;

final class q extends Callback {
   private volatile int d;
   private volatile int c;
   private int e;
   private boolean a;
   private volatile boolean b = true;

   final void a(int var1, int var2, int var3) {
      try {
         try {
            if (var1 != 23529) {
               this.callback(-56, 122, 69, -57);
            }
         } catch (RuntimeException var5) {
            throw var5;
         }

         User32.SetCursorPos(var3, var2);
      } catch (RuntimeException var6) {
         throw var6;
      }
   }

   final void a(int param1, Component param2, boolean param3) {
      // $VF: Couldn't be decompiled
      // Please report this to the Vineflower issue tracker, at https://github.com/Vineflower/vineflower/issues with a copy of the class file (if you have the rights to distribute it!)
      // java.lang.RuntimeException: parsing failure!
      //   at org.jetbrains.java.decompiler.modules.decompiler.decompose.DomHelper.parseGraph(DomHelper.java:211)
      //   at org.jetbrains.java.decompiler.main.rels.MethodProcessor.codeToJava(MethodProcessor.java:166)
      //
      // Bytecode:
      // 00: aload 2
      // 01: invokevirtual java/awt/Component.getPeer ()Ljava/awt/peer/ComponentPeer;
      // 04: checkcast com/ms/awt/WComponentPeer
      // 07: astore 4
      // 09: aload 4
      // 0b: invokevirtual com/ms/awt/WComponentPeer.getTopHwnd ()I
      // 0e: istore 5
      // 10: iload 5
      // 12: bipush -1
      // 13: ixor
      // 14: aload 0
      // 15: getfield q.d I
      // 18: bipush -1
      // 19: ixor
      // 1a: if_icmpne 37
      // 1d: aload 0
      // 1e: getfield q.b Z
      // 21: ifne 2d
      // 24: goto 28
      // 27: athrow
      // 28: bipush 1
      // 29: goto 2e
      // 2c: athrow
      // 2d: bipush 0
      // 2e: iload 3
      // 2f: if_icmpne 36
      // 32: goto 37
      // 35: athrow
      // 36: return
      // 37: iload 1
      // 38: bipush -4
      // 3a: if_icmpeq 46
      // 3d: aload 0
      // 3e: bipush -1
      // 3f: putfield q.d I
      // 42: goto 46
      // 45: athrow
      // 46: aload 0
      // 47: getfield q.a Z
      // 4a: ifne 66
      // 4d: aload 0
      // 4e: bipush 0
      // 4f: sipush 32512
      // 52: invokestatic com/ms/win32/User32.LoadCursor (II)I
      // 55: putfield q.e I
      // 58: aload 0
      // 59: invokestatic com/ms/dll/Root.alloc (Ljava/lang/Object;)I
      // 5c: pop
      // 5d: aload 0
      // 5e: bipush 1
      // 5f: putfield q.a Z
      // 62: goto 66
      // 65: athrow
      // 66: iload 5
      // 68: aload 0
      // 69: getfield q.d I
      // 6c: if_icmpeq d1
      // 6f: aload 0
      // 70: getfield q.d I
      // 73: ifeq aa
      // 76: goto 7a
      // 79: athrow
      // 7a: aload 0
      // 7b: bipush 1
      // 7c: putfield q.b Z
      // 7f: iload 5
      // 81: ldc 101024
      // 83: bipush 0
      // 84: bipush 0
      // 85: invokestatic com/ms/win32/User32.SendMessage (IIII)I
      // 88: pop
      // 89: aload 0
      // 8a: dup
      // 8b: astore 6
      // 8d: monitorenter
      // 8e: aload 0
      // 8f: getfield q.d I
      // 92: bipush -4
      // 94: aload 0
      // 95: getfield q.c I
      // 98: invokestatic com/ms/win32/User32.SetWindowLong (III)I
      // 9b: pop
      // 9c: aload 6
      // 9e: monitorexit
      // 9f: goto aa
      // a2: astore 7
      // a4: aload 6
      // a6: monitorexit
      // a7: aload 7
      // a9: athrow
      // aa: aload 0
      // ab: dup
      // ac: astore 6
      // ae: monitorenter
      // af: aload 0
      // b0: iload 5
      // b2: putfield q.d I
      // b5: aload 0
      // b6: aload 0
      // b7: getfield q.d I
      // ba: bipush -4
      // bc: aload 0
      // bd: invokestatic com/ms/win32/User32.SetWindowLong (IILjava/lang/Object;)I
      // c0: putfield q.c I
      // c3: aload 6
      // c5: monitorexit
      // c6: goto d1
      // c9: astore 8
      // cb: aload 6
      // cd: monitorexit
      // ce: aload 8
      // d0: athrow
      // d1: aload 0
      // d2: iload 3
      // d3: putfield q.b Z
      // d6: iload 5
      // d8: ldc 101024
      // da: bipush 0
      // db: bipush 0
      // dc: invokestatic com/ms/win32/User32.SendMessage (IIII)I
      // df: pop
      // e0: goto e8
      // e3: astore 4
      // e5: aload 4
      // e7: athrow
      // e8: return
   }

   final synchronized int callback(int param1, int param2, int param3, int param4) {
      // $VF: Couldn't be decompiled
      // Please report this to the Vineflower issue tracker, at https://github.com/Vineflower/vineflower/issues with a copy of the class file (if you have the rights to distribute it!)
      // java.lang.RuntimeException: parsing failure!
      //   at org.jetbrains.java.decompiler.modules.decompiler.decompose.DomHelper.parseGraph(DomHelper.java:211)
      //   at org.jetbrains.java.decompiler.main.rels.MethodProcessor.codeToJava(MethodProcessor.java:166)
      //
      // Bytecode:
      // 00: aload 0
      // 01: getfield q.d I
      // 04: bipush -1
      // 05: ixor
      // 06: iload 1
      // 07: bipush -1
      // 08: ixor
      // 09: if_icmpeq 1f
      // 0c: iload 1
      // 0d: bipush -4
      // 0f: invokestatic com/ms/win32/User32.GetWindowLong (II)I
      // 12: istore 5
      // 14: iload 5
      // 16: iload 1
      // 17: iload 2
      // 18: iload 3
      // 19: iload 4
      // 1b: invokestatic com/ms/win32/User32.CallWindowProc (IIIII)I
      // 1e: ireturn
      // 1f: bipush -33
      // 21: iload 2
      // 22: bipush -1
      // 23: ixor
      // 24: if_icmpne 4e
      // 27: ldc 65535
      // 29: iload 4
      // 2b: iand
      // 2c: istore 5
      // 2e: bipush 1
      // 2f: iload 5
      // 31: if_icmpne 4e
      // 34: aload 0
      // 35: getfield q.b Z
      // 38: ifeq 47
      // 3b: goto 3f
      // 3e: athrow
      // 3f: aload 0
      // 40: getfield q.e I
      // 43: goto 48
      // 46: athrow
      // 47: bipush 0
      // 48: invokestatic com/ms/win32/User32.SetCursor (I)I
      // 4b: pop
      // 4c: bipush 0
      // 4d: ireturn
      // 4e: ldc -101025
      // 50: iload 2
      // 51: bipush -1
      // 52: ixor
      // 53: if_icmpeq 5a
      // 56: goto 70
      // 59: athrow
      // 5a: aload 0
      // 5b: getfield q.b Z
      // 5e: ifeq 69
      // 61: aload 0
      // 62: getfield q.e I
      // 65: goto 6a
      // 68: athrow
      // 69: bipush 0
      // 6a: invokestatic com/ms/win32/User32.SetCursor (I)I
      // 6d: pop
      // 6e: bipush 0
      // 6f: ireturn
      // 70: bipush -2
      // 72: iload 2
      // 73: bipush -1
      // 74: ixor
      // 75: if_icmpeq 7c
      // 78: goto 86
      // 7b: athrow
      // 7c: aload 0
      // 7d: bipush 0
      // 7e: putfield q.d I
      // 81: aload 0
      // 82: bipush 1
      // 83: putfield q.b Z
      // 86: aload 0
      // 87: getfield q.c I
      // 8a: iload 1
      // 8b: iload 2
      // 8c: iload 3
      // 8d: iload 4
      // 8f: invokestatic com/ms/win32/User32.CallWindowProc (IIIII)I
      // 92: ireturn
      // 93: astore 5
      // 95: aload 5
      // 97: athrow
   }
}
