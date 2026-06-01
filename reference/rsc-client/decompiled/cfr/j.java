/*
 * Decompiled with CFR 0.152.
 */
import java.awt.Component;
import java.awt.Point;
import java.awt.Robot;
import java.awt.image.BufferedImage;

final class j {
    private Robot a = new Robot();
    private Component b;

    public final void showcursor(Component component, boolean bl) {
        block14: {
            block13: {
                if (bl) {
                    component = null;
                } else {
                    if (component == null) {
                        throw new NullPointerException();
                    }
                }
                if (component == this.b) break block13;
                break block14;
            }
            return;
        }
        if (this.b != null) {
            this.b.setCursor(null);
            this.b = null;
        }
        if (component != null) {
            component.setCursor(component.getToolkit().createCustomCursor(new BufferedImage(1, 1, 2), new Point(0, 0), null));
            this.b = component;
        }
    }

    public final void setcustomcursor(Component component, int[] nArray, int n2, int n3, Point point) {
        if (null != nArray) {
            BufferedImage bufferedImage = new BufferedImage(n2, n3, 2);
            bufferedImage.setRGB(0, 0, n2, n3, nArray, 0, n2);
            component.setCursor(component.getToolkit().createCustomCursor(bufferedImage, point, null));
        } else {
            component.setCursor(null);
        }
    }

    public final void movemouse(int n2, int n3) {
        this.a.mouseMove(n2, n3);
    }
}

