package v235

import (
	"fmt"
	"math/big"
)

// RSAPublicKey holds the server's public key for encrypting the login
// credential block. The OpenRSC server uses "raw" RSA (no padding):
// ciphertext = plaintext^e mod n.
//
// Source: server/src/com/openrsc/server/net/rsc/Crypto.java:88-92
// (decryptRSA uses BigInteger.modPow directly with no PKCS#1 padding).
type RSAPublicKey struct {
	Modulus  *big.Int
	Exponent *big.Int // typically 65537 for RSC
}

// DefaultServerRSA returns the known OpenRSC server's RSA public key as
// observed in the running server's startup log. Both modulus and exponent
// are stable across restarts (persisted to server/{client,server}.pem).
//
// If the server is reset and a new key generated, this value must be
// updated. The current key was captured during the bootstrap session.
func DefaultServerRSA() *RSAPublicKey {
	mod, _ := new(big.Int).SetString(
		"7634250561283973106419144827843935010165327069935723928109242614288318739395804201883596278169185387687268116837066108754542364007806573724086207095863517",
		10,
	)
	return &RSAPublicKey{
		Modulus:  mod,
		Exponent: big.NewInt(65537),
	}
}

// Encrypt performs raw RSA encryption: plaintext^e mod n.
//
// Returns the ciphertext in the same byte layout Java's
// BigInteger.toByteArray() would: a big-endian two's-complement
// representation. If the high bit of the natural-form bytes is set, a
// leading 0x00 is prepended so Java's BigInteger(byte[]) interprets the
// value as positive.
//
// The plaintext is interpreted as a big-endian unsigned integer.
//
// For RSC's login block, plaintext is exactly 61 bytes (see login.go).
func (k *RSAPublicKey) Encrypt(plaintext []byte) ([]byte, error) {
	if len(plaintext) == 0 {
		return nil, fmt.Errorf("rsa: empty plaintext")
	}
	m := new(big.Int).SetBytes(plaintext)
	if m.Cmp(k.Modulus) >= 0 {
		return nil, fmt.Errorf("rsa: plaintext too large for modulus")
	}
	c := new(big.Int).Exp(m, k.Exponent, k.Modulus)
	b := c.Bytes() // no sign byte; high bit could be set

	// Match Java's BigInteger.toByteArray(): if high bit set, prepend 0.
	if len(b) > 0 && b[0]&0x80 != 0 {
		return append([]byte{0x00}, b...), nil
	}
	return b, nil
}
