package v235

// Isaac is a Bob Jenkins ISAAC (Indirection, Shift, Accumulate, Add,
// Count) stream cipher, used by RSC to obfuscate the opcode byte of
// every framed packet post-login.
//
// Ported from OpenRSC's login/ISAACCipher.java, which is itself based
// on Bob Jenkins's Java reference at http://www.burtleburtle.net/bob/java/rand/Rand.java
//
// Authors of the Java implementation (per source comment): Graham
// Edgecombe, rscminus contributors. License: GPL v3+.
//
// Implementation notes:
//
//   - Uses uint32 throughout (Java's int is 32-bit signed; the
//     arithmetic operations used by ISAAC are bit-identical between
//     signed/unsigned, but Go's >>> equivalent for unsigned is just >>
//     on uint32, which avoids sign-extension surprises).
//   - 256-word internal state (mem) and result buffer (results).
//   - Seeding is via SetKeys(k []uint32) where len(k) <= 256. Typical
//     seed for RSC: 4 uint32s extracted from the login block.
type Isaac struct {
	count   int
	results [isaacSize]uint32
	mem     [isaacSize]uint32
	a, b, c uint32
}

const (
	isaacSizeLog = 8
	isaacSize    = 1 << isaacSizeLog // 256
	isaacMask    = (isaacSize - 1) << 2
	isaacRatio   = uint32(0x9e3779b9) // golden ratio
)

// NewIsaac returns an Isaac in uninitialized state. Call SetKeys
// before using.
func NewIsaac() *Isaac { return &Isaac{} }

// SetKeys seeds the cipher with up to 256 uint32 values and runs the
// 2-pass initialization. RSC uses exactly 4 keys.
func (s *Isaac) SetKeys(keys []uint32) {
	for i := range s.results {
		s.results[i] = 0
	}
	for i, k := range keys {
		if i >= isaacSize {
			break
		}
		s.results[i] = k
	}
	s.init(true)
}

// NextValue returns the next 32-bit value from the ISAAC keystream.
func (s *Isaac) NextValue() uint32 {
	if s.count == 0 {
		s.isaac()
		s.count = isaacSize
	}
	s.count--
	return s.results[s.count]
}

// EncodeOpcode applies the cipher to an outbound opcode byte. Used
// when writing post-login frames.
func (s *Isaac) EncodeOpcode(op byte) byte {
	return byte(uint32(op)+s.NextValue()) & 0xFF
}

// DecodeOpcode reverses an inbound opcode byte. Used when reading
// post-login frames.
func (s *Isaac) DecodeOpcode(enc byte) byte {
	return byte(uint32(enc)-s.NextValue()) & 0xFF
}

// isaac runs one generator round, producing 256 new results.
//
// Ported verbatim from OpenRSC's ISAACCipher.isaac().
func (s *Isaac) isaac() {
	s.c++
	s.b += s.c
	for i, j := 0, isaacSize/2; i < isaacSize/2; {
		x := s.mem[i]
		s.a ^= s.a << 13
		s.a += s.mem[j]
		j++
		y := s.mem[(x&isaacMask)>>2] + s.a + s.b
		s.mem[i] = y
		s.b = s.mem[((y>>isaacSizeLog)&isaacMask)>>2] + x
		s.results[i] = s.b
		i++

		x = s.mem[i]
		s.a ^= s.a >> 6
		s.a += s.mem[j]
		j++
		y = s.mem[(x&isaacMask)>>2] + s.a + s.b
		s.mem[i] = y
		s.b = s.mem[((y>>isaacSizeLog)&isaacMask)>>2] + x
		s.results[i] = s.b
		i++

		x = s.mem[i]
		s.a ^= s.a << 2
		s.a += s.mem[j]
		j++
		y = s.mem[(x&isaacMask)>>2] + s.a + s.b
		s.mem[i] = y
		s.b = s.mem[((y>>isaacSizeLog)&isaacMask)>>2] + x
		s.results[i] = s.b
		i++

		x = s.mem[i]
		s.a ^= s.a >> 16
		s.a += s.mem[j]
		j++
		y = s.mem[(x&isaacMask)>>2] + s.a + s.b
		s.mem[i] = y
		s.b = s.mem[((y>>isaacSizeLog)&isaacMask)>>2] + x
		s.results[i] = s.b
		i++
	}
	for i, j := isaacSize/2, 0; j < isaacSize/2; {
		x := s.mem[i]
		s.a ^= s.a << 13
		s.a += s.mem[j]
		j++
		y := s.mem[(x&isaacMask)>>2] + s.a + s.b
		s.mem[i] = y
		s.b = s.mem[((y>>isaacSizeLog)&isaacMask)>>2] + x
		s.results[i] = s.b
		i++

		x = s.mem[i]
		s.a ^= s.a >> 6
		s.a += s.mem[j]
		j++
		y = s.mem[(x&isaacMask)>>2] + s.a + s.b
		s.mem[i] = y
		s.b = s.mem[((y>>isaacSizeLog)&isaacMask)>>2] + x
		s.results[i] = s.b
		i++

		x = s.mem[i]
		s.a ^= s.a << 2
		s.a += s.mem[j]
		j++
		y = s.mem[(x&isaacMask)>>2] + s.a + s.b
		s.mem[i] = y
		s.b = s.mem[((y>>isaacSizeLog)&isaacMask)>>2] + x
		s.results[i] = s.b
		i++

		x = s.mem[i]
		s.a ^= s.a >> 16
		s.a += s.mem[j]
		j++
		y = s.mem[(x&isaacMask)>>2] + s.a + s.b
		s.mem[i] = y
		s.b = s.mem[((y>>isaacSizeLog)&isaacMask)>>2] + x
		s.results[i] = s.b
		i++
	}
}

// init runs the 2-pass ISAAC initialization.
//
// Ported verbatim from OpenRSC's ISAACCipher.init().
func (s *Isaac) init(secondPass bool) {
	a, b, c, d, e, f, g, h := isaacRatio, isaacRatio, isaacRatio, isaacRatio,
		isaacRatio, isaacRatio, isaacRatio, isaacRatio

	mix := func() {
		a ^= b << 11
		d += a
		b += c
		b ^= c >> 2
		e += b
		c += d
		c ^= d << 8
		f += c
		d += e
		d ^= e >> 16
		g += d
		e += f
		e ^= f << 10
		h += e
		f += g
		f ^= g >> 4
		a += f
		g += h
		g ^= h << 8
		b += g
		h += a
		h ^= a >> 9
		c += h
		a += b
	}

	for i := 0; i < 4; i++ {
		mix()
	}

	for i := 0; i < isaacSize; i += 8 {
		if secondPass {
			a += s.results[i]
			b += s.results[i+1]
			c += s.results[i+2]
			d += s.results[i+3]
			e += s.results[i+4]
			f += s.results[i+5]
			g += s.results[i+6]
			h += s.results[i+7]
		}
		mix()
		s.mem[i] = a
		s.mem[i+1] = b
		s.mem[i+2] = c
		s.mem[i+3] = d
		s.mem[i+4] = e
		s.mem[i+5] = f
		s.mem[i+6] = g
		s.mem[i+7] = h
	}

	if secondPass {
		for i := 0; i < isaacSize; i += 8 {
			a += s.mem[i]
			b += s.mem[i+1]
			c += s.mem[i+2]
			d += s.mem[i+3]
			e += s.mem[i+4]
			f += s.mem[i+5]
			g += s.mem[i+6]
			h += s.mem[i+7]
			mix()
			s.mem[i] = a
			s.mem[i+1] = b
			s.mem[i+2] = c
			s.mem[i+3] = d
			s.mem[i+4] = e
			s.mem[i+5] = f
			s.mem[i+6] = g
			s.mem[i+7] = h
		}
	}

	s.isaac()
	s.count = isaacSize
}
