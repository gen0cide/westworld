#!/usr/bin/env python3
"""Normalize /tmp/ww_passages.json into worldmap/transport.json.

Output schema (the durable, version-controlled transport ground truth):
{
  "edges":   [ TransportEdge, ... ],   // resolved standing edges
  "skipped": int,                       // records that could not be resolved
  "skipped_reasons": { reason: count }
}

TransportEdge {
  "kind":  "teleport" | "gate",        // teleport = jump from->to; gate = conditional barrier line
  "category": str,                      // ferry/toll/gate/ladder/...
  "name":  str,                         // human label
  "source_file": str,                   // provenance
  "from":  [x,y]  (optional)            // board / approach tile (plane-0)
  "to":    [x,y]  (optional)            // teleport target tile (plane-0)
  "barrier": {"axis":"x"|"y","line":int,"lo":int,"hi":int} (optional, gate only)
  "req":   Requirement
  "traversability": str
}
Requirement = exactly one populated discriminant (others zero/empty):
  {"coins":int, "item":str, "skill_name":str, "skill_level":int,
   "quest_done":str, "members":bool, "none":bool}
"""
import json, re, collections

SRC = "/tmp/ww_passages.json"
OUT = "/Users/flint/Code/westworld/worldmap/transport.json"

records = json.load(open(SRC))

COORD = re.compile(r'\(\s*(\d{1,4})\s*,\s*(\d{1,4})\s*\)')

def plane0_coords(s):
    out = []
    for a, b in COORD.findall(s or ''):
        x, y = int(a), int(b)
        if 0 <= x <= 1007 and 0 <= y <= 1007:
            out.append((x, y))
    return out

# ---- requirement parsing -------------------------------------------------
SKILL_NAMES = ["agility","mining","fishing","crafting","cooking","magic",
               "woodcutting","thieving","strength","prayer","attack","defense",
               "ranged","herblaw","smithing","firemaking","fletching"]

def parse_requirement(req, notes):
    """Return (Requirement dict, ok, skip_reason).

    ok=False means this record is NOT a clean host-capability gate we model
    (lever/puzzle/cache/disguise/position/random/mid-quest-stage); skip it.
    """
    r = (req or '').strip()
    low = r.lower()

    # Pure 'none' => always-traversable standing edge.
    if low == 'none' or low.startswith('none ') or low.startswith('none(') or low == 'none.':
        return ({"none": True}, True, None)

    # ---- HIGH-CONFIDENCE CAPABILITIES (matched BEFORE the reject list so a
    #      routine "(MEMBER_WORLD config)" annotation does not knock out a clean
    #      members/coins gate). ----

    # Shantay: members + consumable desert pass -> model as Item (the pass).
    if 'shantay desert pass' in low or ('members' in low and 'shantay' in low):
        return ({"item": "Shantay Desert Pass", "members": True}, True, None)

    # Coins toll/fare (e.g. "10 coins toll", "30 coins", "500 coins AND quest ...").
    # We model the coin capability; quest-free clauses are handled by the caller
    # for the toll gate specifically.
    mcoins = re.search(r'(\d+)\s*coins?', low)
    if mcoins:
        return ({"coins": int(mcoins.group(1))}, True, None)

    # Members gate (e.g. "members (MEMBER_WORLD config) + carrying NO blocked
    # weapons/armour"). The no-weapons clause is an equip restriction, not a
    # host capability we track in v1; the binding gate is membership.
    if low.startswith('members'):
        return ({"members": True}, True, None)

    # ---- Reject classes we do NOT model as a standing host capability:
    #   cache flags (mid-state), levers/puzzle tiles, disguises, random rolls,
    #   position/side-dependent, "stage N" mid-quest (only stage==-1/complete is
    #   a capability), specific keys/disguise-sets. ----
    reject_markers = [
        "cache ", "lever", "combination", "dials ", "puzzle", "random",
        "succeed(", "success roll", "calcproduction", "failcalculation",
        "disguise", "robe", "apron", "chefs_hat", "chef hat", "gnome_ball",
        "side-dependent", "side/position", "position dependent", "x>=", "x<=",
        "y>=", "y<=", "from x", "from y", "trip wire", "trap", "trips",
        "pick lock", "lockpick", "menu confirm", "confirmation", "dialogue confirm",
        "dialogue 'teleport", "blocked during", "blocked if carrying",
        "ana barrel", "ana in a barrel", "must not carry", "must not be carrying",
        "wilderness>=", "fisher king", "mod room", "non-admin",
        "stage 0", "stage>=2 +", "stage>=3 +", "stage>=5;", "stage 4 or 5",
        "stage 6-7", "stage 8/9", "stage 13", "stage 15", "stage>=11 /",
        "gang", "phoenix", "black arm", "amulet (4 charges", "dragonstone amulet",
        "disk of returning", "agility cape", "crafting cape", "fishing cape",
        "talisman", "orb spell", "charge fire", "magical_fire_pass", "blessed_golden_bowl",
        "skavid map", "beads of the dead", "bone key", "metal key", "wrought iron key",
        "little_key", "jail_key", "dusty_key", "miscellaneous key", "bunch of keys",
        "an_old_key", "closet key", "front door key", "cell door key", "keep_key",
        "matching key", "key (", "key used", "key when", "key to unlock", "key found",
        "glarial", "dramen_staff", "paramaya_rest_ticket", "disk_of_returning",
    ]
    for m in reject_markers:
        if m in low:
            return (None, False, "non-capability/transient requirement")

    # ---- Skill gates: "skill: Fishing current level >= 68", "skill: agility 32",
    #      "skill: Mining current level >= 60", "skill: Crafting >= 40 AND ..." ----
    if low.startswith('skill:') or 'current level >=' in low or re.search(r'skill:\s*\w+\s*>?=?\s*\d+', low):
        for name in SKILL_NAMES:
            if name in low:
                mlvl = re.search(r'(?:>=\s*|level\s*>=\s*|\b' + name + r'\s+)(\d+)', low)
                # generic: any number after the skill name
                if not mlvl:
                    mlvl = re.search(r'(\d+)', low)
                if mlvl:
                    return ({"skill_name": name.capitalize(), "skill_level": int(mlvl.group(1))}, True, None)
        # skill mentioned but no clean level (formula/soft) -> reject
        return (None, False, "soft/formula skill check")

    # ---- Quest complete (the only quest predicate that is a standing capability):
    #      "quest X complete (stage == -1)" / "quest X == -1" ----
    if 'quest' in low and ('== -1' in low or '==-1' in low or 'complete' in low or 'completed' in low):
        # only treat as standing if it is a COMPLETE predicate, not a mid-stage one
        if re.search(r'stage\s*(?:in|==)?\s*\{?\s*\d', low) and 'stage == -1' not in low and 'stage==-1' not in low:
            return (None, False, "mid-quest-stage gate")
        # extract quest name between 'quest' and the predicate
        mq = re.search(r'quest\s+([A-Za-z][A-Za-z _\'-]+?)\s*(?:complete|==|stage|\()', r, re.IGNORECASE)
        qname = (mq.group(1).strip() if mq else '').strip()
        qname = re.sub(r'\s+', ' ', qname)
        # reject junk captures (the words "stage"/"points"/"slayer stage"/empty):
        # those mean the quest name did not parse cleanly -> not a clean gate.
        junk = {"", "stage", "points", "stage ==", "slayer stage"}
        if qname.lower() in junk or len(qname) < 3:
            return (None, False, "unparseable quest name")
        return ({"quest_done": qname}, True, None)

    # ---- Plain item consumable (e.g. "item: Ship ticket (consumed each trip)",
    #      "item: DIAMOND", "20 Limpwurt Roots") ----
    if low.startswith('item:') or 'limpwurt' in low or low.startswith('item '):
        # only model SIMPLE single-item passes (not disguise sets / keys already rejected)
        mit = re.search(r'item:\s*([A-Za-z][A-Za-z _]+?)\s*(?:\(|$|;| via| AND)', r, re.IGNORECASE)
        if 'limpwurt' in low:
            return ({"item": "Limpwurt Roots"}, True, None)
        if mit:
            item = mit.group(1).strip()
            return ({"item": item}, True, None)
        return (None, False, "unparseable item requirement")

    return (None, False, "unrecognized requirement")


# ---- traversability filter ----------------------------------------------
def is_standing_candidate(rec):
    t = rec['traversability']
    if t in ('repeatable', 'post_quest_permanent'):
        return True
    if t == 'conditional_other':
        # only keep conditional_other whose requirement is a host capability
        return True  # parse_requirement decides; non-capability ones get rejected there
    return False  # one_time_quest skipped


# ---- non-passage filter --------------------------------------------------
NON_PASSAGE_MARKERS = ["no passage", "n/a", "herb-search", "npc-self", "punitive",
                       "knockback", "self-teleport", "not a passage"]
def is_non_passage(rec):
    blob = (rec.get('notes','') + ' ' + rec.get('effect','')).lower()
    return any(m in blob for m in NON_PASSAGE_MARKERS)


edges = []
skip_reasons = collections.Counter()

for rec in records:
    if not is_standing_candidate(rec):
        skip_reasons['one_time_quest'] += 1
        continue
    if is_non_passage(rec):
        skip_reasons['non-passage record'] += 1
        continue

    req_obj, ok, reason = parse_requirement(rec['requirement'], rec.get('notes',''))
    if not ok:
        skip_reasons[reason or 'unparseable requirement'] += 1
        continue

    src_coords = plane0_coords(rec['location'])
    dst_coords = plane0_coords(rec['effect'])

    cat = rec['category']
    is_toll = (cat == 'toll') or ('toll' in (rec['location'] or '').lower())
    is_border_guard = 'borderguard' in (rec['source_file'] or '').lower()

    edge = {
        "category": cat,
        "name": rec['location'][:80],
        "source_file": rec['source_file'],
        "req": req_obj,
        "traversability": rec['traversability'],
    }

    # ---- THE AL-KHARID TOLL GATE: a conditional BARRIER, not a teleport. ----
    # Collision data has NO border wall here (Lumbridge & the mine are one
    # walking component -- the locked TollGateCorrection test). The gate is
    # pure server game-logic, so we overlay it as a conditional cut on the
    # authentic Al-Kharid border line x=108 (derived by min-cut probing the
    # real collision flood: a full vertical wall at x=108 separates the mine
    # (74,583) from Lumbridge (135,654); opening it reconnects them). Crossing
    # is permitted iff Coins>=10 OR Prince Ali Rescue complete.
    if is_border_guard:
        edge["kind"] = "gate"
        edge["barrier"] = {"axis": "x", "line": 108, "lo": 0, "hi": 1007}
        # free if Prince Ali Rescue complete -> encode the quest-free clause.
        edge["req"] = {"coins": 10, "quest_free": "Prince Ali Rescue"}
        if dst_coords:
            edge["to"] = list(dst_coords[0])
        edges.append(edge)
        continue

    # ---- Teleport links: need a TARGET (to). Without it, skip (no coords). ----
    if dst_coords:
        edge["kind"] = "teleport"
        edge["to"] = list(dst_coords[0])
        if src_coords:
            edge["from"] = list(src_coords[0])
        edges.append(edge)
        continue

    # ---- "opens in place" gates/doors with NO teleport target but a RESOLVED
    #      source tile. These open in the collision map already (BlocksMovement
    #      false, exactly like the toll gate), so they are NOT flood-cuts and we
    #      do NOT synthesize a barrier line for them (we cannot derive which
    #      side is "inside"). We still record them as informational `gate` edges
    #      at their tile (with their capability requirement) so a later layer can
    #      surface "this door needs <skill/quest>". They have no flood effect in
    #      v1 -- only the explicit axis-line gates (the toll gate) cut the flood. ----
    if src_coords:
        if not req_obj.get("none"):
            edge["kind"] = "gate"
            x, y = src_coords[0]
            edge["from"] = [x, y]
            edges.append(edge)
            continue
        skip_reasons['in-place open, no capability gate'] += 1
        continue

    skip_reasons['unresolved coords (loc/bound id only)'] += 1

out = {
    "_comment": "Normalized standing transport edges for the worldmap WorldOracle. "
                "Generated from the 431 OpenRSC passage records by tools/normalize_transport.py. "
                "Durable, version-controlled transport ground truth. v1: plane-0, static world.",
    "edges": edges,
    "skipped": sum(skip_reasons.values()),
    "skipped_reasons": dict(skip_reasons),
}
json.dump(out, open(OUT, 'w'), indent=1)
print("wrote", OUT)
print("edges:", len(edges))
print("skipped:", out['skipped'])
for k, v in skip_reasons.most_common():
    print(f"  {v:4d}  {k}")
