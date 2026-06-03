#!/usr/bin/env python3
"""Canonical 3-engine render-parity diff: per-pixel max-channel absolute delta.

This is the ONE diff method the whole render-parity rig uses (orsc / DEOB / JAR),
so every leg and every verifier measures identically. It loads two PNGs, converts
both to RGB, and reports the max per-channel absolute difference (maxR maxG maxB)
plus the count of pixels where ANY channel differs. Exit 0 iff byte-identical
(0/0/0, ndiff 0); exit 1 on any difference or a size mismatch.

usage:  pngdiff.py A.png B.png [--label NAME]
"""
import sys
from PIL import Image


def load_rgb(path):
    im = Image.open(path).convert("RGB")
    return im


def main():
    args = [a for a in sys.argv[1:] if not a.startswith("--")]
    label = ""
    for i, a in enumerate(sys.argv):
        if a == "--label" and i + 1 < len(sys.argv):
            label = sys.argv[i + 1]
    if len(args) < 2:
        print("usage: pngdiff.py A.png B.png [--label NAME]", file=sys.stderr)
        return 2
    a, b = load_rgb(args[0]), load_rgb(args[1])
    tag = f"[{label}] " if label else ""
    if a.size != b.size:
        print(f"{tag}SIZE MISMATCH {a.size} vs {b.size}")
        return 1
    pa, pb = a.load(), b.load()
    w, h = a.size
    maxr = maxg = maxb = 0
    ndiff = 0
    for y in range(h):
        for x in range(w):
            ra, ga, ba = pa[x, y]
            rb, gb, bb = pb[x, y]
            dr, dg, db = abs(ra - rb), abs(ga - gb), abs(ba - bb)
            if dr or dg or db:
                ndiff += 1
                if dr > maxr:
                    maxr = dr
                if dg > maxg:
                    maxg = dg
                if db > maxb:
                    maxb = db
    print(f"{tag}max(R,G,B)={maxr}/{maxg}/{maxb}  ndiff={ndiff}/{w*h}  {args[0]} vs {args[1]}")
    return 0 if ndiff == 0 else 1


if __name__ == "__main__":
    sys.exit(main())
