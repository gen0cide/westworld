package render

import "math"

// groundColour is the 256-entry ground-colour palette anIntArray578
// (World.<init>), each entry an encoded flat 5:5:5 fill from method305.
var groundColour [256]int32

// method305 encodes an (r,g,b) triple as a flat fill: -1 - (r/8)*1024 -
// (g/8)*32 - b/8 (Scene.method305).
func method305(r, g, b int32) int32 {
	return -1 - (r/8)*1024 - (g/8)*32 - b/8
}

func init() {
	for i := int32(0); i < 64; i++ {
		groundColour[i] = method305(255-i*4, 255-int32(float64(i)*1.75), 255-i*4)
	}
	for j := int32(0); j < 64; j++ {
		groundColour[j+64] = method305(j*3, 144, 0)
	}
	for k := int32(0); k < 64; k++ {
		groundColour[k+128] = method305(192-int32(float64(k)*1.5), 144-int32(float64(k)*1.5), 0)
	}
	for l := int32(0); l < 64; l++ {
		groundColour[l+192] = method305(96-int32(float64(l)*1.5), 48+int32(float64(l)*1.5), 0)
	}
	_ = math.Pi
}
