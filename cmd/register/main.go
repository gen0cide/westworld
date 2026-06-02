// Command register creates an OpenRSC account over the wire using the
// account-registration opcode (OutRegister = 2), which cradle itself does not
// implement (cradle only logs in). It speaks the authentic ">=204" register
// format the server expects from a v235 client, so we can mint a throwaway
// account to drive the remote client (cmd/cradle -client) for a live smoke test.
//
// Usage:
//
//	register -server localhost:43594 -username webclient -password smoketester1
//
// The wire format (server net/rsc/LoginPacketHandler.java REGISTER_ACCOUNT,
// clientVersion >= 204 branch), payload AFTER the framed opcode 2:
//
//	[2]  clientVersion        (unsigned short, big-endian)
//	[8]  usernameHash         (base-37, DataConversions.usernameToHash)
//	3×:  [1] blockLen, [blockLen] RSA(plaintext)   — plaintext is 15 bytes:
//	         [0:4]  nonce (ignored; byte 0 nonzero + high-bit-clear so the
//	                server's BigInteger round-trip yields exactly 15 bytes)
//	         [4:8]  session id (big-endian int; the server adopts ours when its
//	                own session is unset, else it must match what it sent us)
//	         [8:15] 7 password chars (3 blocks → 21 bytes, space-padded, trimmed)
//	[4]  hashed random.dat     (any int)
//
// The server replies with a RegisterLoginResponse byte; 2 == REGISTER_SUCCESSFUL.
package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"time"

	v235 "github.com/gen0cide/westworld/proto/v235"
)

// usernameToHash mirrors OpenRSC DataConversions.usernameToHash: lowercase,
// keep [a-z0-9] (everything else becomes a space), trim, cap at 12 chars, then
// fold into a base-37 long (a-z → 1..26, 0-9 → 27..36).
func usernameToHash(s string) uint64 {
	s = strings.ToLower(s)
	var b strings.Builder
	for _, c := range s {
		if (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') {
			b.WriteRune(c)
		} else {
			b.WriteByte(' ')
		}
	}
	t := strings.TrimSpace(b.String())
	if len(t) > 12 {
		t = t[:12]
	}
	var l uint64
	for i := 0; i < len(t); i++ {
		c := t[i]
		l *= 37
		switch {
		case c >= 'a' && c <= 'z':
			l += uint64(1 + c - 'a')
		case c >= '0' && c <= '9':
			l += uint64(27 + c - '0')
		}
	}
	return l
}

func main() {
	server := flag.String("server", "localhost:43594", "host:port of the OpenRSC server")
	username := flag.String("username", "", "account username (<=12 chars, [a-z0-9])")
	password := flag.String("password", "", "account password (no spaces); or set WESTWORLD_PASSWORD")
	version := flag.Int("version", 235, "client version to present")
	flag.Parse()
	if *password == "" {
		*password = os.Getenv("WESTWORLD_PASSWORD")
	}
	if *username == "" || *password == "" {
		fmt.Fprintln(os.Stderr, "register: need -username and -password (or WESTWORLD_PASSWORD)")
		os.Exit(2)
	}

	pub := v235.DefaultServerRSA()

	// Password → 21 bytes (3×7), space-padded; the server trims it.
	pwPad := make([]byte, 21)
	for i := range pwPad {
		pwPad[i] = ' '
	}
	copy(pwPad, []byte(*password))

	conn, err := net.DialTimeout("tcp", *server, 5*time.Second)
	if err != nil {
		fmt.Fprintln(os.Stderr, "register: dial:", err)
		os.Exit(1)
	}
	defer conn.Close()

	// Capture the optional 4-byte session id the server may push on connect
	// (it sets its own session to that value, so our blocks must echo it). If
	// none arrives, the server's session stays unset and adopts ours (0).
	var sessionID uint32
	_ = conn.SetReadDeadline(time.Now().Add(800 * time.Millisecond))
	sbuf := make([]byte, 4)
	if _, err := io.ReadFull(conn, sbuf); err == nil {
		sessionID = binary.BigEndian.Uint32(sbuf)
		fmt.Printf("register: captured session id %d\n", sessionID)
	} else {
		fmt.Println("register: no session id from server (using 0)")
	}
	_ = conn.SetReadDeadline(time.Time{})

	// Build the register payload.
	body := make([]byte, 0, 256)
	body = binary.BigEndian.AppendUint16(body, uint16(*version))
	body = binary.BigEndian.AppendUint64(body, usernameToHash(*username))
	for i := 0; i < 3; i++ {
		plain := make([]byte, 15)
		plain[0] = 0x01 // nonzero + high-bit-clear → exact 15-byte RSA round-trip
		binary.BigEndian.PutUint32(plain[4:8], sessionID)
		copy(plain[8:15], pwPad[i*7:i*7+7])
		ct, err := pub.Encrypt(plain)
		if err != nil {
			fmt.Fprintln(os.Stderr, "register: rsa encrypt:", err)
			os.Exit(1)
		}
		if len(ct) > 255 {
			fmt.Fprintln(os.Stderr, "register: rsa block too long for a byte length")
			os.Exit(1)
		}
		body = append(body, byte(len(ct)))
		body = append(body, ct...)
	}
	body = append(body, 0, 0, 0, 0) // hashed random.dat (unused)

	frame := v235.EncodeFrame(v235.OutRegister, body)
	if _, err := conn.Write(frame); err != nil {
		fmt.Fprintln(os.Stderr, "register: write:", err)
		os.Exit(1)
	}

	// Read the response. The server writes a leading byte (0) immediately, then
	// queues the account creation on its async login executor and writes the real
	// RegisterLoginResponse code a moment later. We MUST keep the connection open
	// and keep reading until that code arrives — closing after the leading byte
	// makes the server abort the async create ("Channel inactive"). Scan the
	// stream for 2 (REGISTER_SUCCESSFUL) / 3 (taken) / 100 (unsuccessful).
	_ = conn.SetReadDeadline(time.Now().Add(8 * time.Second))
	var got []byte
	buf := make([]byte, 64)
	for {
		n, err := conn.Read(buf)
		if n > 0 {
			got = append(got, buf[:n]...)
			for _, b := range buf[:n] {
				switch b {
				case 2:
					fmt.Printf("register: response %v -> REGISTER_SUCCESSFUL for %q\n", got, *username)
					return
				case 3:
					fmt.Printf("register: response %v -> USERNAME_TAKEN_OR_INVALID (try logging in)\n", got)
					return
				case 100:
					fmt.Fprintf(os.Stderr, "register: response %v -> UNSUCCESSFUL (decrypt/parse error); see server log\n", got)
					os.Exit(1)
				}
			}
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "register: stream ended; bytes=%v err=%v; check server log\n", got, err)
			os.Exit(1)
		}
	}
}
