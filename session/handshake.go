package session

import (
	"context"
	"fmt"
	"time"

	"github.com/gen0cide/westworld/proto/v235"
)

// LoginParams are the inputs to a Login handshake.
type LoginParams struct {
	Username      string
	Password      string
	ClientVersion uint16          // 235 for our v235 target
	RSAPublicKey  *v235.RSAPublicKey
}

// LoginResult is what a successful login returns.
type LoginResult struct {
	ResponseCode byte
	// ISaacKeys are the four uint32s that seeded both directions of
	// the cipher. Returned for diagnostics / logging.
	IsaacKeys [4]uint32
}

// Login performs the post-dial handshake against the OpenRSC server:
//
//  1. Build the login packet with a fresh ISAAC seed
//  2. Send it (PLAIN opcode, since ISAAC isn't active yet)
//  3. Read the 1-byte response code
//  4. If success (high bit set per LoginResponse semantics), seed
//     both ISAAC ciphers with the committed keys and return
//
// Caller is expected to invoke c.Start() after a successful Login to
// launch the duplex I/O goroutines.
func (c *Conn) Login(ctx context.Context, p LoginParams) (*LoginResult, error) {
	if p.RSAPublicKey == nil {
		p.RSAPublicKey = v235.DefaultServerRSA()
	}
	if p.ClientVersion == 0 {
		p.ClientVersion = 235
	}

	payload := &v235.LoginPayload{
		Username:      p.Username,
		Password:      p.Password,
		ClientVersion: p.ClientVersion,
	}
	wire, err := payload.EncodeLoginFrame(p.RSAPublicKey)
	if err != nil {
		return nil, fmt.Errorf("session: build login frame: %w", err)
	}

	c.log.Debug("sending login packet",
		"username", p.Username,
		"version", p.ClientVersion,
		"wire_bytes", len(wire),
	)

	// Write the framed login packet synchronously. Opcode in the
	// frame is already plain (we used EncodeFrame directly).
	if _, err := c.netConn.Write(wire); err != nil {
		return nil, fmt.Errorf("session: write login: %w", err)
	}

	// Read the 1-byte response code with a reasonable deadline.
	rspBuf, err := c.ReadPreLoginBytes(1, 10*time.Second)
	if err != nil {
		return nil, fmt.Errorf("session: read login response: %w", err)
	}
	rsp := rspBuf[0]

	c.log.Info("login response received",
		"code", rsp,
		"successful", v235.LoginSuccessful(rsp),
	)

	if !v235.LoginSuccessful(rsp) {
		return &LoginResult{ResponseCode: rsp}, fmt.Errorf("session: login rejected (code %d)", rsp)
	}

	// Seed both ISAAC ciphers with the keys we committed to in the
	// RSA block.
	c.SetIsaacKeys(payload.IsaacKeys)

	return &LoginResult{
		ResponseCode: rsp,
		IsaacKeys:    payload.IsaacKeys,
	}, nil
}
