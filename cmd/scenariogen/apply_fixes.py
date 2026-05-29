#!/usr/bin/env python3
"""Deterministically apply triage find/replace fixes to scenarios.yaml.

Each fix is scoped to its scenario block (delimited by `^  - id: <id>`),
because substrings like `item 33 1` / `teleport 150 504` legitimately recur
across scenarios. Within a block the find must occur exactly once (else we
flag it rather than guess). Dry-run by default; pass --apply to write back.
"""
import json, re, sys, os

REPO = os.path.dirname(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))
YAML = os.path.join(REPO, "cmd/scenariogen/scenarios.yaml")

def load_fixes(path):
    d = json.load(open(path))
    return d["result"]["fixes"]

def split_blocks(text):
    """Return (preamble, [(id, block_text), ...]) preserving exact bytes."""
    # Block starts at a line matching `  - id: <id>` and runs to the next such line.
    idx = [m.start() for m in re.finditer(r"^  - id: (\S+)", text, re.M)]
    ids = [m.group(1) for m in re.finditer(r"^  - id: (\S+)", text, re.M)]
    preamble = text[:idx[0]]
    blocks = []
    for i, start in enumerate(idx):
        end = idx[i+1] if i+1 < len(idx) else len(text)
        blocks.append([ids[i], text[start:end]])
    return preamble, blocks

def main():
    apply = "--apply" in sys.argv
    fixes_path = [a for a in sys.argv[1:] if not a.startswith("-")][0]
    fixes = load_fixes(fixes_path)
    text = open(YAML).read()
    preamble, blocks = split_blocks(text)
    bymap = {bid: i for i, (bid, _) in enumerate(blocks)}

    applied = nomatch = multi = noid = 0
    for f in fixes:
        sid, find, repl, conf = f["scenario_id"], f["find"], f["replace"], f["confidence"]
        if sid not in bymap:
            print(f"  NO-ID   [{conf}] {sid}: scenario block not found")
            noid += 1
            continue
        bi = bymap[sid]
        block = blocks[bi][1]
        n = block.count(find)
        if n == 0:
            print(f"  NOMATCH [{conf}] {sid}: find not present")
            print(f"          find={find!r}")
            nomatch += 1
        elif n > 1:
            print(f"  MULTI({n}) [{conf}] {sid}: find occurs {n}x — replacing FIRST")
            blocks[bi][1] = block.replace(find, repl, 1)
            multi += 1
            applied += 1
        else:
            blocks[bi][1] = block.replace(find, repl, 1)
            applied += 1

    print(f"\nsummary: applied={applied} nomatch={nomatch} multi={multi} no-id={noid} / total={len(fixes)}")
    if apply:
        if nomatch or noid:
            print("REFUSING to write: unmatched fixes present — resolve first.")
            sys.exit(1)
        out = preamble + "".join(b for _, b in blocks)
        open(YAML, "w").write(out)
        print(f"WROTE {YAML}")
    else:
        print("(dry-run; pass --apply to write)")

if __name__ == "__main__":
    main()
