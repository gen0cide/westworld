// Package hostkv is a host's local key/value storage faculty: the small,
// process-local persistence a single host carries with it as it plays.
//
// Two stores, two lifetimes:
//
//   - Store is DURABLE. A versioned JSON document on disk (one file per
//     host, atomic temp+rename writes), the source of truth for things a
//     host must remember across logins: its relationship/trust ledger,
//     goal progress, learned world facts. It mirrors the conventions of
//     cognition/resolve.AliasStore — in-memory map of truth, opportunistic
//     write-through, a missing file is empty (not an error), a corrupt file
//     IS an error so a host never silently forgets.
//
//   - Scratch is EPHEMERAL. An in-memory, bounded (LRU), TTL'd cache for a
//     host's short-term working memory: "I already examined this offer",
//     "last place I stood", refractory timers. Never touches disk. It is
//     the same shape the pearl decision cache uses, kept here so both share
//     one well-tested LRU.
//
// The split is deliberate: durable writes (relationship updates, goal
// changes) are infrequent and worth a disk flush; scratch writes are hot
// and must never hit disk. Code reaches for Store when "a host would be
// wrong to forget this on logout" and Scratch otherwise.
package hostkv
