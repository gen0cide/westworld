// Package session manages the TCP connection lifecycle to an OpenRSC
// server: dial, the session-id-then-login dance, ISAAC seeding, and
// duplex packet flow via channels.
//
// One Conn represents one live RSC server connection. It owns two
// goroutines:
//
//   - read loop: reads bytes from the socket, feeds the FrameDecoder,
//     pushes decoded frames to the Recv() channel.
//   - write loop: pulls packets from the Send() channel, ISAAC-encodes
//     the opcode, frames the bytes, writes to the socket.
//
// The session layer doesn't know about game semantics — it just shovels
// typed packets in both directions. Walk, attack, login etc. are
// composed at higher layers.
package session
