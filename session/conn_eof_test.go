package session

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"
)

// dialPair spins a loopback listener, dials a Conn at it, and returns the
// session Conn + the server side of the socket.
func dialPair(t *testing.T) (*Conn, net.Conn) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	accepted := make(chan net.Conn, 1)
	go func() {
		s, aerr := ln.Accept()
		if aerr == nil {
			accepted <- s
		}
	}()
	c, err := Dial(context.Background(), ln.Addr().String(), Options{})
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	srv := <-accepted
	return c, srv
}

// TestServerEOFStoresErrServerClosed guards the supervisor-restart contract:
// a SERVER-initiated close (idle-kick / bounce) must store ErrServerClosed —
// the silent-nil behavior made host.Run return clean, the supervisor mark the
// host Stopped, and long soaks bleed the fleet to zero.
func TestServerEOFStoresErrServerClosed(t *testing.T) {
	c, srv := dialPair(t)
	defer c.Close()
	c.Start()
	srv.Close() // server hangs up
	select {
	case <-c.Done():
	case <-time.After(2 * time.Second):
		t.Fatal("conn did not tear down after server EOF")
	}
	if err := c.Err(); !errors.Is(err, ErrServerClosed) {
		t.Fatalf("Err() = %v, want ErrServerClosed", err)
	}
}

// TestLocalCloseStaysClean: an operator-initiated Close is a clean stop —
// no error, no restart.
func TestLocalCloseStaysClean(t *testing.T) {
	c, srv := dialPair(t)
	defer srv.Close()
	c.Start()
	c.Close()
	select {
	case <-c.Done():
	case <-time.After(2 * time.Second):
		t.Fatal("conn did not tear down after local Close")
	}
	time.Sleep(100 * time.Millisecond) // let the read loop observe the close
	if err := c.Err(); err != nil {
		t.Fatalf("Err() = %v, want nil for local Close", err)
	}
}
