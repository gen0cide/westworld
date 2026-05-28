package v235

import "sync"

// RSC chat compression: a fixed Huffman-like prefix code over the
// byte alphabet 0-255, designed for compact English-text chat
// messages. Used in opcode 234 update-type 1/6/7 (player chat),
// opcode 120 (private messages), and elsewhere.
//
// Ported from OpenRSC's util/rsc/StringEncryption.java (which is
// itself derived from the authentic Jagex client + rscminus). The
// tables (init[], cipherBlock[], cipherDictionary[]) are built at
// package init time by porting the Java constructor verbatim.
//
// Algorithm summary:
//
// Encode: for each character c in the message, look up bitLength =
// init[c] and bitPattern = cipherBlock[c]. Pack those bits MSB-first
// into the output byte stream.
//
// Decode: walk the cipherDictionary tree bit-by-bit. The bits of each
// input byte are processed MSB-first (bit 7 first). A negative value
// at the current tree index means we've hit a terminal (the output
// character is ~node); reset to root.

// init is the per-character bit length (256 entries) — copied
// verbatim from StringEncryption.java:43-52.
var rscInit = [256]byte{
	22, 22, 22, 22, 22, 22, 21, 22, 22, 20, 22, 22, 22, 21, 22, 22,
	22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22,
	3, 8, 22, 16, 22, 16, 17, 7, 13, 13, 13, 16,
	7, 10, 6, 16, 10, 11, 12, 12, 12, 12, 13, 13, 14, 14, 11, 14, 19, 15, 17, 8, 11, 9, 10, 10, 10, 10, 11, 10,
	9, 7, 12, 11, 10, 10, 9, 10, 10, 12, 10, 9, 8, 12, 12, 9, 14, 8, 12, 17, 16, 17, 22, 13, 21, 4, 7, 6, 5, 3,
	6, 6, 5, 4, 10, 7, 5, 6, 4, 4, 6, 10, 5, 4, 4, 5, 7, 6, 10, 6, 10, 22, 19, 22, 14, 22, 22, 22, 22, 22, 22,
	22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22,
	22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22,
	22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22,
	22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22,
	22, 22, 22, 22, 22, 22, 21, 22, 21, 22, 22, 22, 21, 22, 22,
}

var (
	rscBuildOnce       sync.Once
	rscCipherBlock     [256]uint32
	rscCipherDictionary []int32 // grows dynamically during build
)

// buildRSCTables ports the Java constructor at StringEncryption.java:70-137.
func buildRSCTables() {
	rscCipherDictionary = make([]int32, 8)
	blockBuilder := make([]uint32, 33)
	cipherDictIndexTemp := int32(0)

	for initPos := 0; initPos < len(rscInit); initPos++ {
		initValue := int(rscInit[initPos])
		builderBitSelector := uint32(1) << (32 - uint(initValue))
		builderValue := blockBuilder[initValue]
		rscCipherBlock[initPos] = builderValue
		var builderValueBit uint32

		if (builderValue & builderBitSelector) == 0 {
			builderValueBit = builderValue | builderBitSelector
			for initValueCounter := initValue - 1; initValueCounter > 0; initValueCounter-- {
				builderValue2 := blockBuilder[initValueCounter]
				if builderValue != builderValue2 {
					break
				}
				builderValue2BitSelector := uint32(1) << (32 - uint(initValueCounter))
				if (builderValue2 & builderValue2BitSelector) == 0 {
					blockBuilder[initValueCounter] = builderValue2BitSelector | builderValue2
				} else {
					blockBuilder[initValueCounter] = blockBuilder[initValueCounter-1]
					break
				}
			}
		} else {
			builderValueBit = blockBuilder[initValue-1]
		}
		blockBuilder[initValue] = builderValueBit
		for initValueCounter := initValue + 1; initValueCounter <= 32; initValueCounter++ {
			if builderValue == blockBuilder[initValueCounter] {
				blockBuilder[initValueCounter] = builderValueBit
			}
		}

		cipherDictIndex := int32(0)
		for initValueCounter := 0; initValueCounter < initValue; initValueCounter++ {
			builderBitSelector2 := uint32(0x80000000) >> uint(initValueCounter)
			if (builderValue & builderBitSelector2) == 0 {
				cipherDictIndex++
			} else {
				if rscCipherDictionary[cipherDictIndex] == 0 {
					rscCipherDictionary[cipherDictIndex] = cipherDictIndexTemp
				}
				cipherDictIndex = rscCipherDictionary[cipherDictIndex]
			}
			if int(cipherDictIndex) >= len(rscCipherDictionary) {
				newDict := make([]int32, len(rscCipherDictionary)*2)
				copy(newDict, rscCipherDictionary)
				rscCipherDictionary = newDict
			}
		}
		// Terminal node: ~initPos. In Go we use bitwise complement.
		rscCipherDictionary[cipherDictIndex] = ^int32(initPos)
		if cipherDictIndex >= cipherDictIndexTemp {
			cipherDictIndexTemp = cipherDictIndex + 1
		}
	}
}

func ensureRSCTables() {
	rscBuildOnce.Do(buildRSCTables)
}

// EncipherRSCString compresses a plain-text message into the RSC
// wire format used by chat packets. Returns the compressed bytes;
// the caller is responsible for prefixing the smart-length char
// count and concatenating with surrounding packet fields.
//
// Ported from StringEncryption.java:144-189.
func EncipherRSCString(message string) []byte {
	ensureRSCTables()
	chatBuffer := convertMessageToBytes(message)

	// Output buffer sized for worst case: 22 bits per char.
	outBuf := make([]byte, (len(chatBuffer)*22)/8+2)

	encipheredByte := uint32(0)
	outputBitOffset := 0
	for messageIndex := 0; messageIndex < len(chatBuffer); messageIndex++ {
		messageCharacter := int(chatBuffer[messageIndex]) & 0xFF
		cipherBlockValue := rscCipherBlock[messageCharacter]
		initValue := int(rscInit[messageCharacter])

		outputByteOffset := outputBitOffset >> 3
		cipherBlockShifter := 7 & outputBitOffset
		// `encipheredByte &= -cipherBlockShifter >> 31` — Java arithmetic
		// shift on signed int. When cipherBlockShifter is 0, mask is 0,
		// clearing encipheredByte; otherwise mask is 0xFFFFFFFF, leaving
		// it unchanged.
		if cipherBlockShifter == 0 {
			encipheredByte = 0
		}
		outputByteOffset2 := outputByteOffset + ((cipherBlockShifter + initValue - 1) >> 3)
		outputBitOffset += initValue
		cipherBlockShifter += 24
		encipheredByte |= cipherBlockValue >> uint(cipherBlockShifter)
		outBuf[outputByteOffset] = byte(encipheredByte)

		if outputByteOffset2 > outputByteOffset {
			outputByteOffset++
			cipherBlockShifter -= 8
			encipheredByte = cipherBlockValue >> uint(cipherBlockShifter)
			outBuf[outputByteOffset] = byte(encipheredByte)
			if outputByteOffset < outputByteOffset2 {
				outputByteOffset++
				cipherBlockShifter -= 8
				encipheredByte = cipherBlockValue >> uint(cipherBlockShifter)
				outBuf[outputByteOffset] = byte(encipheredByte)
				if outputByteOffset2 > outputByteOffset {
					outputByteOffset++
					cipherBlockShifter -= 8
					encipheredByte = cipherBlockValue >> uint(cipherBlockShifter)
					outBuf[outputByteOffset] = byte(encipheredByte)
					if outputByteOffset2 > outputByteOffset {
						cipherBlockShifter -= 8
						outputByteOffset++
						// Java: cipherBlockValue << -cipherBlockShifter
						// where cipherBlockShifter is negative. In Go we
						// compute the left shift amount as the positive value.
						leftShift := -cipherBlockShifter
						encipheredByte = cipherBlockValue << uint(leftShift)
						outBuf[outputByteOffset] = byte(encipheredByte)
					}
				}
			}
		}
	}

	totalBytes := (outputBitOffset + 7) >> 3
	return outBuf[:totalBytes]
}

// DecipherRSCString decompresses an RSC-encoded chat body of the
// given expected character count. Returns just the decoded string.
// Use DecipherRSCStringWithLen if you need to know how many input
// bytes were consumed.
func DecipherRSCString(src []byte, count int) string {
	s, _ := DecipherRSCStringWithLen(src, count)
	return s
}

// DecipherRSCStringWithLen decompresses count chars from src and
// returns (decoded string, number of source bytes consumed). Use
// the byte count to advance a packet reader past the chat body so
// the next packet record can be parsed.
//
// Ported from StringEncryption.java:192-329.
func DecipherRSCStringWithLen(src []byte, count int) (string, int) {
	ensureRSCTables()
	if count <= 0 || len(src) == 0 {
		return "", 0
	}
	dest := make([]byte, 0, count)
	var var7 int32 = 0
	srcOffset := 0
	bits := [8]byte{0x80, 0x40, 0x20, 0x10, 0x08, 0x04, 0x02, 0x01}

	for len(dest) < count && srcOffset < len(src) {
		byteVal := src[srcOffset]
		for _, bit := range bits {
			if (byteVal & bit) != 0 {
				var7 = rscCipherDictionary[var7]
			} else {
				var7++
			}
			if int(var7) >= len(rscCipherDictionary) {
				return convertBytesToString(dest), srcOffset + 1
			}
			if rscCipherDictionary[var7] < 0 {
				dest = append(dest, byte(^rscCipherDictionary[var7]))
				if len(dest) >= count {
					return convertBytesToString(dest), srcOffset + 1
				}
				var7 = 0
			}
		}
		srcOffset++
	}
	return convertBytesToString(dest), srcOffset
}

// convertMessageToBytes maps a Go string to the byte-array RSC
// expects: 0-127 ASCII pass through; some Unicode characters get
// folded to special-byte slots 128-159; everything else becomes '?'.
func convertMessageToBytes(s string) []byte {
	runes := []rune(s)
	out := make([]byte, len(runes))
	for i, c := range runes {
		if (c <= 0 || c >= 0x80) && (c < 0xa0 || c > 0xff) {
			// Special-character fold (best effort; for chat the common
			// case is ASCII-only).
			out[i] = '?'
		} else {
			out[i] = byte(c)
		}
	}
	return out
}

// convertBytesToString does the inverse: takes decompressed bytes
// and emits a Go string, mapping special-byte slots 128-159 to their
// Unicode equivalents.
func convertBytesToString(b []byte) string {
	runes := make([]rune, 0, len(b))
	for _, by := range b {
		cp := int(by) & 0xFF
		if cp == 0 {
			continue // null padding
		}
		if cp >= 128 && cp < 160 {
			// Special character — map back to Unicode. We don't have
			// the table imported; fall back to '?'.
			runes = append(runes, '?')
		} else {
			runes = append(runes, rune(cp))
		}
	}
	return string(runes)
}
