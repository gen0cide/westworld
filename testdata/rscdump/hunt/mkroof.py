#!/usr/bin/env python3
"""Author roof2 hunt fixtures by mutating a base roof fixture's grids.

Grids are row-major Size*Size, index = x*Size + y (x=col=worldX-baseX,
y=row=worldY-baseY). This matches rscdump.Terrain.idx and the deob's
lx (outer/X) , ly (inner/Y) convention once you account for getWallRoof's
swapped index (deob getWallRoof(ly,lx) -> wallsRoof[q][lx*48+ly], i.e. the
SAME physical cell as grid[lx*Size+ly]).
"""
import json, base64, sys, copy

SZ = 24
BASE = "/home/free/code/rsc-hacking/westworld/testdata/rscdump/hunt/roof_lone.json"

def load_base():
    return json.load(open(BASE))

def blank_byte_grid():
    return bytearray(SZ * SZ)

def idx(x, y):
    return x * SZ + y

def setb(grid, x, y, v):
    grid[idx(x, y)] = v

def emit(name, roof=None, diag=None, wallH=None, wallV=None, overlay=None,
         elevation=None, plane=0, self_xy=None):
    d = load_base()
    t = d["terrain"]
    # roof grid
    rg = blank_byte_grid()
    if roof:
        for (x, y, v) in roof:
            setb(rg, x, y, v)
    t["roof"] = base64.b64encode(bytes(rg)).decode()
    # walls
    wh = blank_byte_grid()
    if wallH:
        for (x, y, v) in wallH:
            setb(wh, x, y, v)
    t["wallH"] = base64.b64encode(bytes(wh)).decode()
    wv = blank_byte_grid()
    if wallV:
        for (x, y, v) in wallV:
            setb(wv, x, y, v)
    t["wallV"] = base64.b64encode(bytes(wv)).decode()
    # overlay
    if overlay is not None:
        og = blank_byte_grid()
        for (x, y, v) in overlay:
            setb(og, x, y, v)
        t["overlay"] = base64.b64encode(bytes(og)).decode()
    # elevation (default flat zero already in base)
    if elevation is not None:
        eg = blank_byte_grid()
        for (x, y, v) in elevation:
            setb(eg, x, y, v)
        t["elevation"] = base64.b64encode(bytes(eg)).decode()
    # diag (int array)
    dg = [0] * (SZ * SZ)
    if diag:
        for (x, y, v) in diag:
            dg[idx(x, y)] = v
    t["wallDiag"] = dg
    t["terrainSeed"] = 0
    d["window"]["plane"] = plane
    if self_xy:
        d["self"]["x"], d["self"]["y"] = self_xy
    out = "/home/free/code/rsc-hacking/westworld/testdata/rscdump/hunt/" + name + ".json"
    json.dump(d, open(out, "w"), indent=2)
    print("wrote", out)

if __name__ == "__main__":
    pass
