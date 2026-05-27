package session

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"sync"
	"time"

	"github.com/gen0cide/westworld/proto/v235"
)

// Conn is a duplex RSC server connection with the wire protocol
// already established (session ID read, login completed, ISAAC seeded).
//
// Users push outbound packets via Send and receive inbound frames via
// the channel returned by Recv. Errors that terminate the connection
// (read EOF, write failure, etc.) are surfaced via Err.
type Conn struct {
	netConn net.Conn

	enc *v235.Isaac
	dec *v235.Isaac

	decoder *v235.FrameDecoder

	sendCh chan outgoing
	recvCh chan v235.Frame

	closeOnce sync.Once
	done      chan struct{}
	err       atomicErr

	log *slog.Logger
}

type outgoing struct {
	opcode  byte
	payload []byte
}

// Options for dialing a connection.
type Options struct {
	Logger *slog.Logger
	// DialTimeout is the TCP dial deadline; default 5s.
	DialTimeout time.Duration
	// PreLoginReadDelay is how long to wait before attempting to read
	// the optional 4-byte session ID from the server. Default 800ms
	// (the server's default SESSION_ID_SENDER_TIMER is 640ms; we add
	// margin). If the server sends within this window, we discard the
	// bytes. If it doesn't, we proceed to login.
	PreLoginReadDelay time.Duration
}

// Dial opens a TCP connection to the OpenRSC server and waits for the
// optional session ID. The returned Conn is NOT yet logged in — the
// caller must invoke Login.
//
// addr is in "host:port" form (e.g., "localhost:43596").
func Dial(ctx context.Context, addr string, opts Options) (*Conn, error) {
	if opts.Logger == nil {
		opts.Logger = slog.Default()
	}
	if opts.DialTimeout == 0 {
		opts.DialTimeout = 5 * time.Second
	}
	if opts.PreLoginReadDelay == 0 {
		opts.PreLoginReadDelay = 800 * time.Millisecond
	}

	dialer := net.Dialer{Timeout: opts.DialTimeout}
	netConn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("session: dial %s: %w", addr, err)
	}

	c := &Conn{
		netConn: netConn,
		enc:     v235.NewIsaac(),
		dec:     v235.NewIsaac(),
		decoder: v235.NewFrameDecoder(v235.PlainDecode),
		sendCh:  make(chan outgoing, 32),
		recvCh:  make(chan v235.Frame, 32),
		done:    make(chan struct{}),
		log:     opts.Logger,
	}

	// Read and discard the optional 4-byte session ID, if it arrives
	// within PreLoginReadDelay. We don't use the session ID for v235
	// login — the server may or may not send it depending on whether
	// our login packet beats the server's session-id timer.
	if err := c.consumeOptionalSessionID(opts.PreLoginReadDelay); err != nil {
		_ = netConn.Close()
		return nil, fmt.Errorf("session: consume session id: %w", err)
	}

	return c, nil
}

// consumeOptionalSessionID waits up to delay for the server to send a
// 4-byte session ID. If bytes arrive, they're discarded. If the
// deadline expires with no bytes, that's also fine.
func (c *Conn) consumeOptionalSessionID(delay time.Duration) error {
	if err := c.netConn.SetReadDeadline(time.Now().Add(delay)); err != nil {
		return err
	}
	defer c.netConn.SetReadDeadline(time.Time{}) // clear deadline

	buf := make([]byte, 8)
	n, err := io.ReadFull(c.netConn, buf[:4])
	if err == nil {
		c.log.Debug("consumed session id", "bytes", n, "first4", buf[:4])
		return nil
	}
	if nerr, ok := err.(net.Error); ok && nerr.Timeout() {
		c.log.Debug("no session id received (timeout); server likely suppressed it")
		return nil
	}
	if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
		// Connection closed before any data — unexpected
		return fmt.Errorf("connection closed prematurely")
	}
	return err
}

// Start launches the read and write goroutines. Must be called after
// any pre-login I/O (e.g., Login) has completed and ISAAC is seeded.
func (c *Conn) Start() {
	go c.readLoop()
	go c.writeLoop()
}

// SetIsaacKeys seeds the in/out ISAAC ciphers with the given keys.
// Call this immediately after a successful login response, BEFORE
// Start launches the I/O goroutines.
//
// After this is called, the decoder's opcode-decode function switches
// from PlainDecode to the ISAAC-decrypted variant.
func (c *Conn) SetIsaacKeys(keys [4]uint32) {
	c.enc.SetKeys(keys[:])
	c.dec.SetKeys(keys[:])
	c.decoder.SetDecode(c.dec.DecodeOpcode)
}

// SendRaw writes one framed packet to the network with the given
// PLAIN opcode (not ISAAC-encoded). Used for the login packet only.
// The write happens synchronously on the caller's goroutine.
//
// After Start has been called, prefer Send().
func (c *Conn) SendRaw(opcode byte, payload []byte) error {
	wire := v235.EncodeFrame(opcode, payload)
	_, err := c.netConn.Write(wire)
	return err
}

// Send queues a packet for sending via the write goroutine. The opcode
// will be ISAAC-encoded automatically. Returns an error only if the
// connection is closed.
func (c *Conn) Send(opcode byte, payload []byte) error {
	select {
	case c.sendCh <- outgoing{opcode: opcode, payload: payload}:
		return nil
	case <-c.done:
		return c.Err()
	}
}

// Recv returns the channel of decoded inbound frames. The channel is
// closed when the connection terminates.
func (c *Conn) Recv() <-chan v235.Frame { return c.recvCh }

// Done returns a channel that's closed when the connection has been
// torn down (by Close, error, or remote EOF).
func (c *Conn) Done() <-chan struct{} { return c.done }

// Err returns the terminal error, if any. Returns nil if the
// connection closed cleanly.
func (c *Conn) Err() error { return c.err.Load() }

// Close shuts down the connection. Idempotent.
func (c *Conn) Close() error {
	c.closeOnce.Do(func() {
		c.netConn.Close()
		close(c.done)
	})
	return nil
}

// ReadPreLoginBytes performs a synchronous, deadline-bounded read of
// exactly n bytes. Used during the login handshake before Start has
// launched the read goroutine.
func (c *Conn) ReadPreLoginBytes(n int, deadline time.Duration) ([]byte, error) {
	if err := c.netConn.SetReadDeadline(time.Now().Add(deadline)); err != nil {
		return nil, err
	}
	defer c.netConn.SetReadDeadline(time.Time{})
	buf := make([]byte, n)
	if _, err := io.ReadFull(c.netConn, buf); err != nil {
		return nil, err
	}
	return buf, nil
}

// readLoop reads bytes from the socket and feeds the FrameDecoder,
// emitting decoded frames on recvCh. Terminates on read error.
func (c *Conn) readLoop() {
	defer close(c.recvCh)
	defer c.Close()

	scratch := make([]byte, 4096)
	for {
		n, err := c.netConn.Read(scratch)
		if n > 0 {
			c.decoder.Feed(scratch[:n])
			for {
				frame, derr := c.decoder.Next()
				if errors.Is(derr, v235.ErrShortRead) {
					break
				}
				if derr != nil {
					c.err.Store(fmt.Errorf("decode: %w", derr))
					return
				}
				select {
				case c.recvCh <- frame:
				case <-c.done:
					return
				}
			}
		}
		if err != nil {
			if errors.Is(err, io.EOF) {
				c.log.Debug("read: server closed connection")
			} else {
				c.err.Store(fmt.Errorf("read: %w", err))
			}
			return
		}
	}
}

// writeLoop pulls outgoing packets from sendCh and writes them to the
// socket. ISAAC-encodes the opcode of each. Terminates on write error.
func (c *Conn) writeLoop() {
	for {
		select {
		case <-c.done:
			return
		case out := <-c.sendCh:
			encOp := c.enc.EncodeOpcode(out.opcode)
			wire := v235.EncodeFrame(encOp, out.payload)
			if _, err := c.netConn.Write(wire); err != nil {
				c.err.Store(fmt.Errorf("write: %w", err))
				c.Close()
				return
			}
		}
	}
}

// atomicErr is a tiny atomic wrapper for the connection's terminal
// error. We don't need a sync/atomic.Value because the error is set at
// most once.
type atomicErr struct {
	mu  sync.Mutex
	err error
}

func (a *atomicErr) Store(err error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.err == nil {
		a.err = err
	}
}

func (a *atomicErr) Load() error {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.err
}
