// Package world is the host's local mirror of game state. For Phase 0
// it tracks only the agent's own position — enough to verify walks
// succeed. Later phases extend it with inventory, stats, nearby
// entities, ground items, fatigue/poison, etc.
package world
