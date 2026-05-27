// Package runtime composes the lower-level subsystems (session, world,
// event, action) into a single coherent agent: the Host.
//
// One Host runs in one goroutine. It owns:
//
//   - A session.Conn (the wire connection)
//   - A world.World (the state mirror)
//   - An event.Bus (downstream subscribers)
//
// The runtime's only loop reads decoded frames from the session,
// translates them into typed events via proto/v235.DecodeInbound,
// applies them to the world state, and publishes them to the bus.
// Other layers (the Brain, the script interpreter, observability)
// subscribe to the bus or directly read world state.
//
// Lifecycle:
//
//	host, err := runtime.New(opts)
//	err = host.Connect(ctx)        // dial + login
//	host.Run(ctx)                   // blocks until ctx cancel or error
//	host.Close()                    // tears down
package runtime
