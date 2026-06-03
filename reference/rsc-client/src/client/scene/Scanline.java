package client.scene;

/**
 * Scanline (obf: {@code n}) — a single horizontal span entry in the software rasterizer's
 * per-row span table used by {@link Scene}.
 *
 * <p>{@link Scene} keeps an array {@code Scanline[] scanlines} (obf field {@code x}), one element
 * per screen row.  As a triangle/quad face is decomposed into horizontal spans by
 * {@code Scene.generateScanlines(...)}, each covered row's {@code Scanline} is filled with the
 * left/right X extents and the corresponding interpolated shade values (8.8 fixed point); the
 * rasterizer ({@code Scene.rasterize(...)} family) then walks those spans to fill the framebuffer.
 *
 * <p>The four instance fields keep the original single-letter obfuscated names so the field-access
 * sites in {@link Scene} read identically to the decompiled source (obf class {@code n}):
 * <ul>
 *   <li>{@link #d} — span start X (left edge)</li>
 *   <li>{@link #k} — span end X (right edge)</li>
 *   <li>{@link #e} — span start shade (left, 8.8 fixed point)</li>
 *   <li>{@link #l} — span end shade (right, 8.8 fixed point)</li>
 * </ul>
 *
 * <p>This per-row span struct was never emitted as its own file by the deobfuscation pass even
 * though {@link Scene} references it ({@code new Scanline()} / {@code Scanline[]} and the four
 * field accesses {@code sl.d/sl.k/sl.e/sl.l}); it is reconstructed here, faithful to the clean
 * decompile's obf class {@code n} (instance fields {@code int e; int k; int d; int l;}).
 *
 * obf: n
 */
public final class Scanline {

    /** Span start X — left edge of this row's covered span.  obf: {@code d} */
    public int d;

    /** Span end X — right edge of this row's covered span.  obf: {@code k} */
    public int k;

    /** Span start shade at the left edge (8.8 fixed point).  obf: {@code e} */
    public int e;

    /** Span end shade at the right edge (8.8 fixed point).  obf: {@code l} */
    public int l;
}
