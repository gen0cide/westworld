# look.star — Bot perception dump. Useful for verifying the bot sees
# what we expect before running other scripts.

print("--- self ---")
print(self_state())

print("--- look around (radius=8) ---")
print(look_around(radius = 8))

print("--- inventory ---")
inv = inventory()
if not inv:
    print("(empty)")
else:
    for s in inv:
        print("  slot %d: %s x%d" % (s["slot"], s["name"], s["amount"]))

print("--- nearby attackable NPCs ---")
for n in nearby_npcs():
    if n["attackable"]:
        print("  %s (idx=%d) at (%d, %d) — %s" % (n["name"], n["index"], n["x"], n["y"], n["description"]))
