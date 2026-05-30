package assets

// Magic is GameModel.magic / faceIntensity sentinel (0xbc614e = 12345678).
// faceFill 32767 in the .ob3 stream is replaced with Magic (the no-fill /
// gouraud sentinel). faceIntensity == Magic means per-vertex gouraud shading.
const Magic = 0xbc614e

// Model is a decoded .ob3 geometry blob. Coordinates are model-local; X/Y/Z
// are signed 16-bit. Fills encode flat colour as (-1 - colour) where colour is
// 5:5:5 RGB, or a texture id when >= 0, or Magic for "no fill".
type Model struct {
	NumVertices int
	NumFaces    int

	VertexX []int32
	VertexY []int32
	VertexZ []int32

	FaceNumVertices []int
	FaceVertices    [][]int
	FaceFillFront   []int32
	FaceFillBack    []int32
	FaceIntensity   []int32 // 0 = flat, Magic = per-vertex gouraud
}

func getUnsignedShort(b []byte, o int) int {
	return int(b[o]&0xff)<<8 | int(b[o+1]&0xff)
}

func getSignedShort(b []byte, o int) int32 {
	j := int(b[o]&0xff)*256 + int(b[o+1]&0xff)
	if j > 32767 {
		j -= 0x10000
	}
	return int32(j)
}

// DecodeModel parses a .ob3 blob (GameModel(byte[],offset) port). Big-endian,
// signed shorts for geometry. Per-face vertex indices are u8 when nV < 256
// else u16.
func DecodeModel(data []byte, offset int) *Model {
	nV := getUnsignedShort(data, offset)
	offset += 2
	nF := getUnsignedShort(data, offset)
	offset += 2

	m := &Model{
		NumVertices:     nV,
		NumFaces:        nF,
		VertexX:         make([]int32, nV),
		VertexY:         make([]int32, nV),
		VertexZ:         make([]int32, nV),
		FaceNumVertices: make([]int, nF),
		FaceVertices:    make([][]int, nF),
		FaceFillFront:   make([]int32, nF),
		FaceFillBack:    make([]int32, nF),
		FaceIntensity:   make([]int32, nF),
	}

	for i := 0; i < nV; i++ {
		m.VertexX[i] = getSignedShort(data, offset)
		offset += 2
	}
	for i := 0; i < nV; i++ {
		m.VertexY[i] = getSignedShort(data, offset)
		offset += 2
	}
	for i := 0; i < nV; i++ {
		m.VertexZ[i] = getSignedShort(data, offset)
		offset += 2
	}

	for i := 0; i < nF; i++ {
		m.FaceNumVertices[i] = int(data[offset] & 0xff)
		offset++
	}

	for i := 0; i < nF; i++ {
		f := getSignedShort(data, offset)
		offset += 2
		if f == 32767 {
			f = Magic
		}
		m.FaceFillFront[i] = f
	}
	for i := 0; i < nF; i++ {
		f := getSignedShort(data, offset)
		offset += 2
		if f == 32767 {
			f = Magic
		}
		m.FaceFillBack[i] = f
	}
	for i := 0; i < nF; i++ {
		flag := data[offset] & 0xff
		offset++
		if flag == 0 {
			m.FaceIntensity[i] = 0
		} else {
			m.FaceIntensity[i] = Magic
		}
	}

	for i := 0; i < nF; i++ {
		n := m.FaceNumVertices[i]
		vs := make([]int, n)
		for j := 0; j < n; j++ {
			if nV < 256 {
				vs[j] = int(data[offset] & 0xff)
				offset++
			} else {
				vs[j] = getUnsignedShort(data, offset)
				offset += 2
			}
		}
		m.FaceVertices[i] = vs
	}

	return m
}
