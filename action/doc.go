// Package action provides high-level player actions that translate
// into one-or-more outbound packets via session.Conn. Each action
// performs its own validation (range, preconditions) against the
// world state and then sends.
//
// For Phase 0 only the Walk action is implemented. Phase 1 expands to
// the full vocabulary (attack, eat, talk, bank, etc.).
package action
