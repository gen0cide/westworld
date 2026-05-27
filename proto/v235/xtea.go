package v235

import (
	"encoding/binary"
	"fmt"
)

// XTEA is a Feistel block cipher with 64-bit blocks (two uint32) and a
// 128-bit key (four uint32). OpenRSC uses 32 rounds and big-endian byte
// order for block I/O.
//
// Reference: standard XTEA spec by Needham/Wheeler (1997).
// Cross-referenced with the JS/Java implementations used by the
// historical RSC client family.

const (
	xteaRounds = 32
	xteaDelta  = uint32(0x9E3779B9) // (sqrt(5)-1) * 2^31

	// xteaDecryptInitSum = xteaDelta * xteaRounds (wraps mod 2^32).
	// Go's constant arithmetic is exact, so we precompute the wrapped
	// value here rather than relying on uint32 overflow at constant
	// evaluation time. Verified: 0x9E3779B9 * 32 mod 2^32 = 0xC6EF3720.
	xteaDecryptInitSum = uint32(0xC6EF3720)
)

// XTEAEncrypt encrypts plaintext in place using the given 4-word key.
// plaintext length must be a multiple of 8 bytes.
//
// Byte order within each 8-byte block: big-endian. Block layout:
// [4 bytes v0, big-endian][4 bytes v1, big-endian].
func XTEAEncrypt(plaintext []byte, keys [4]uint32) error {
	if len(plaintext)%8 != 0 {
		return fmt.Errorf("xtea: plaintext length %d is not a multiple of 8", len(plaintext))
	}
	for off := 0; off < len(plaintext); off += 8 {
		v0 := binary.BigEndian.Uint32(plaintext[off : off+4])
		v1 := binary.BigEndian.Uint32(plaintext[off+4 : off+8])
		var sum uint32 = 0
		for r := 0; r < xteaRounds; r++ {
			v0 += (((v1 << 4) ^ (v1 >> 5)) + v1) ^ (sum + keys[sum&3])
			sum += xteaDelta
			v1 += (((v0 << 4) ^ (v0 >> 5)) + v0) ^ (sum + keys[(sum>>11)&3])
		}
		binary.BigEndian.PutUint32(plaintext[off:off+4], v0)
		binary.BigEndian.PutUint32(plaintext[off+4:off+8], v1)
	}
	return nil
}

// XTEADecrypt decrypts ciphertext in place using the given key.
func XTEADecrypt(ciphertext []byte, keys [4]uint32) error {
	if len(ciphertext)%8 != 0 {
		return fmt.Errorf("xtea: ciphertext length %d is not a multiple of 8", len(ciphertext))
	}
	for off := 0; off < len(ciphertext); off += 8 {
		v0 := binary.BigEndian.Uint32(ciphertext[off : off+4])
		v1 := binary.BigEndian.Uint32(ciphertext[off+4 : off+8])
		var sum uint32 = xteaDecryptInitSum
		for r := 0; r < xteaRounds; r++ {
			v1 -= (((v0 << 4) ^ (v0 >> 5)) + v0) ^ (sum + keys[(sum>>11)&3])
			sum -= xteaDelta
			v0 -= (((v1 << 4) ^ (v1 >> 5)) + v1) ^ (sum + keys[sum&3])
		}
		binary.BigEndian.PutUint32(ciphertext[off:off+4], v0)
		binary.BigEndian.PutUint32(ciphertext[off+4:off+8], v1)
	}
	return nil
}
