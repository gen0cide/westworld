// Package facts is the bot's read-only knowledge of static world data:
// every scenery type (well, tree, table, anvil), boundary (wall, fence,
// door), NPC species (Chicken, Goblin, Banker), and item type that
// exists in OpenRSC, plus the world's *placements* of those things —
// where each well, fence, NPC spawn, and ground-item respawn point is.
//
// Loaded once at host startup from OpenRSC's conf/server/defs/ files.
// Treat as immutable thereafter.
//
// # Two important nuances
//
// ## 1. This is KNOWLEDGE, not PERCEPTION
//
// The data here represents what a player *could in principle know*
// about the RSC world — the kind of information that exists on the
// wiki, in player heads, in collective folk knowledge. It is NOT the
// same as what the bot is currently looking at.
//
// A bot that just logged in at Lumbridge spawn has not yet *seen* the
// Falador bank. But it CAN reasonably know "there is a bank in
// Falador" — every player who has been around for a day knows that.
//
// So queries like NearestByName("Bank", pos) answer the question
// "where is the nearest bank in the world?" — a knowledge question.
// They do NOT answer "what do I currently see around me?" — that
// requires the live perception stream from session/world.
//
// The brain layer (when added) is responsible for distinguishing:
//   - "you see..." prompts use live world.* state
//   - "you know..." prompts use facts.*
//   - "you remember..." prompts use mesa-backed episodic memory
//
// For a fresh host (first session), reasonable believability says:
// the bot may know general lore ("Lumbridge has a furnace") but
// shouldn't claim to know obscure things ("the well in the southern
// corner of the cooking guild has a missing bucket") until it has
// observed them. The brain prompt design controls this.
//
// ## 2. ONE copy per process, shared across all hosts
//
// The full dataset is ~3MB of JSON + ~30k spatial index entries —
// roughly 50MB of resident memory once decoded and indexed. At 500
// hosts per process, loading per-host would be 25GB.
//
// The solution: load once per process (typically in main / delos /
// cmd/cradle startup), pass a *Facts pointer to every Host. Reads
// are concurrent-safe by virtue of immutability — no locks needed.
//
// This package provides no caching layer or invalidation — there's
// nothing to invalidate. If OpenRSC's def files change between
// runs, restart the process to pick them up.
//
// # Queries
//
//   facts.SceneryDef(id)               → *SceneryDef
//   facts.NpcDef(id)                   → *NpcDef
//   facts.ItemDef(id)                  → *ItemDef
//   facts.BoundaryDef(id)              → *BoundaryDef
//   facts.At(x, y)                     → all placements at this tile
//   facts.Near(x, y, radius)           → all placements in a square
//   facts.NearestByName(name, pos, r)  → closest matching placement
//
// # What's not here yet (Phase 2.5b)
//
//   - Walkability map: per-tile passability derived from boundary
//     and scenery type=1 placements. Will enable client-side path
//     validation and (eventually) local pathfinding.
//   - Multiple location-data variants: SceneryLocs14.json,
//     SceneryLocsCustomQuest.json, etc. The preservation server config
//     specifies which sets to include (location_data: 1 includes the
//     base set plus discontinued items).
package facts
