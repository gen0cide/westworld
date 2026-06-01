/*
 * Decompiled with CFR 0.152.
 * 
 * Could not load the following classes:
 *  com.ms.awt.WComponentPeer
 *  com.ms.dll.Callback
 *  com.ms.dll.Root
 *  com.ms.win32.User32
 */
import com.ms.awt.WComponentPeer;
import com.ms.dll.Callback;
import com.ms.dll.Root;
import com.ms.win32.User32;
import java.awt.Component;

final class q
extends Callback {
    private volatile int d;
    private volatile int c;
    private int e;
    private boolean a;
    private volatile boolean b = true;

    final void a(int n2, int n3, int n4) {
        if (n2 != 23529) {
            this.callback(-56, 122, 69, -57);
        }
        User32.SetCursorPos((int)n4, (int)n3);
    }

    /*
     * WARNING - Removed try catching itself - possible behaviour change.
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    final void a(int n2, Component component, boolean bl) {
        WComponentPeer wComponentPeer = (WComponentPeer)component.getPeer();
        int n3 = wComponentPeer.getTopHwnd();
        if (~n3 == ~this.d) {
            if (!this.b != bl) return;
        }
        if (n2 != -4) {
            this.d = -1;
        }
        if (!this.a) {
            this.e = User32.LoadCursor((int)0, (int)32512);
            Root.alloc((Object)((Object)this));
            this.a = true;
        }
        if (n3 != this.d) {
            q q2;
            if (this.d != 0) {
                this.b = true;
                User32.SendMessage((int)n3, (int)101024, (int)0, (int)0);
                q2 = this;
                synchronized (q2) {
                    User32.SetWindowLong((int)this.d, (int)-4, (int)this.c);
                }
            }
            q2 = this;
            synchronized (q2) {
                this.d = n3;
                this.c = User32.SetWindowLong((int)this.d, (int)-4, (Object)((Object)this));
            }
        }
        this.b = bl;
        User32.SendMessage((int)n3, (int)101024, (int)0, (int)0);
        return;
    }

    q() {
    }

    /*
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    final synchronized int callback(int n2, int n3, int n4, int n5) {
        int n8;
        if (~this.d != ~n2) {
            int n7 = User32.GetWindowLong((int)n2, (int)-4);
            return User32.CallWindowProc((int)n7, (int)n2, (int)n3, (int)n4, (int)n5);
        }
        if (-33 == ~n3 && 1 == (n8 = 0xFFFF & n5)) {
            User32.SetCursor((int)(this.b ? this.e : 0));
            return 0;
        }
        if (-101025 != ~n3) {
            if (-2 != ~n3) return User32.CallWindowProc((int)this.c, (int)n2, (int)n3, (int)n4, (int)n5);
            this.d = 0;
            this.b = true;
            return User32.CallWindowProc((int)this.c, (int)n2, (int)n3, (int)n4, (int)n5);
        } else {
            User32.SetCursor((int)(this.b ? this.e : 0));
            return 0;
        }
    }
}

