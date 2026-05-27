// Package v235 implements the RuneScape Classic wire protocol as spoken by
// the OpenRSC server at the Payload235 revision (final-era RSC, mc234/235).
//
// The package contains:
//
//   - Frame encoding/decoding (framing/tail-byte/length quirks)
//   - The ISAAC stream cipher used to obfuscate opcode bytes post-login
//   - RSA encryption of the login credential block
//   - Big-endian read/write helpers for primitive types
//   - Opcode constants and per-packet encoders/decoders for the minimum
//     vocabulary needed by westworld's cradle bot runtime
//
// The reference implementation lives in OpenRSC's Java source tree under
// server/src/com/openrsc/server/net/. Every non-trivial function here has
// a doc comment citing the Java source file + line range it was ported from.
//
// Goals:
//
//   - Bit-for-bit wire compatibility with the OpenRSC server.
//   - Pure-function packet codecs (no I/O); session-level concerns live
//     in the session package.
//   - Testable in isolation: each codec has unit tests over captured
//     packet bytes.
//
// Non-goals for v1:
//
//   - Other RSC revisions (Payload38/69/115/140/177/196-203). If needed
//     later, they live in sibling packages (v203/, v177/, etc.).
//   - Bitpacked decoding of all inbound update packets — only the
//     minimum needed for Phase 0 (own player position) is implemented.
package v235
