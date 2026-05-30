package assets

import (
	"bytes"
	"compress/bzip2"
	"fmt"
	"io"
)

// Decompress reverses the Jagex bzip2 transform. Jagex strips the 4-byte
// stream magic "BZh1" from the front of each compressed block (BZLib.java
// hardcodes blocksize100k=1). The standard library compress/bzip2 reader
// expects that magic, so we prepend it and feed the rest to the decoder,
// reading exactly uncompLen bytes back out.
//
// in is the raw (magic-stripped) bzip2 payload; uncompLen is the expected
// decompressed size (from the archive catalog / outer header).
func Decompress(in []byte, uncompLen int) ([]byte, error) {
	var src bytes.Buffer
	src.WriteString("BZh1")
	src.Write(in)
	r := bzip2.NewReader(&src)
	out := make([]byte, uncompLen)
	n, err := io.ReadFull(r, out)
	if err != nil && err != io.ErrUnexpectedEOF && err != io.EOF {
		return nil, fmt.Errorf("bzip2 decompress: %w", err)
	}
	if n != uncompLen {
		return out[:n], fmt.Errorf("bzip2 short read: got %d want %d", n, uncompLen)
	}
	return out, nil
}
