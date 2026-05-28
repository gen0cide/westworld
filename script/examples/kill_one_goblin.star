# kill_one_goblin.star — Find the closest attackable Goblin, attack it,
# wait for the kill, then loot whatever it dropped.
#
# Demonstrates that you can compose primitives from script-land
# without recompiling cradle.

def manhattan(a, b):
    return abs(a[0] - b[0]) + abs(a[1] - b[1])

pos = position()
print("starting at", pos)

# Pick the nearest Goblin.
target = None
best = 1000
for n in nearby_npcs():
    if n["name"] != "Goblin":
        continue
    d = manhattan((n["x"], n["y"]), pos)
    if d < best:
        best = d
        target = n

if target == None:
    print("no goblin nearby")
else:
    print("engaging", target["name"], "at", (target["x"], target["y"]))
    attack_npc(target["index"])
    # Server takes ~5-10s to walk + kill a low-level goblin. Just wait.
    sleep(seconds = 10)

    # Loot any ground items within 3 tiles of where we ended up.
    pos = position()
    for g in ground_items():
        if manhattan((g["x"], g["y"]), pos) <= 3:
            print("looting", g["name"], "at", (g["x"], g["y"]))
            pick_up(x = g["x"], y = g["y"], item_id = g["item_id"])
            sleep(seconds = 1)

    # If we picked up bones, bury them.
    for s in inventory():
        if s["name"] == "Bones":
            print("burying slot", s["slot"])
            item_command(slot = s["slot"])
            break

print("done. final hp:", hp())
