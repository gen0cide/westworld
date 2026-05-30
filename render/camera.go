package render

import "math"

// Sine tables, ported bit-for-bit from GameModel.<clinit> (GameModel.java:1005).
// sine9 is a 512-entry table (256 step rotation) used by model applyRotation;
// sine11 is a 2048-entry table (1024 step rotation) used by project() and
// setCamera. Entry i = sin, entry i+half = cos, scaled by 32768 (>>15 fixed).
var (
	sine9  [512]int32
	sine11 [2048]int32
)

func init() {
	for i := 0; i < 256; i++ {
		sine9[i] = int32(math.Sin(float64(i)*0.02454369) * 32768)
		sine9[i+256] = int32(math.Cos(float64(i)*0.02454369) * 32768)
	}
	for j := 0; j < 1024; j++ {
		sine11[j] = int32(math.Sin(float64(j)*0.00613592315) * 32768)
		sine11[j+1024] = int32(math.Cos(float64(j)*0.00613592315) * 32768)
	}
}

// Camera holds the orbit state computed by SetCamera (Scene.setCamera,
// Scene.java:2290). All fields are integer fixed-point. CameraPitch/Yaw/Roll
// are the inverted angles fed to project(); CameraX/Y/Z is the eye position
// (look-at minus the orbit offset, i.e. the camera sits `distance` behind).
type Camera struct {
	CameraX, CameraY, CameraZ          int32
	CameraPitch, CameraYaw, CameraRoll int32
}

// SetCamera mirrors Scene.setCamera(x, z, y, pitch, yaw, roll, distance).
// NOTE the axis order: arg z is HEIGHT (vertical), arg y is world-south.
func SetCamera(x, z, y, pitch, yaw, roll, distance int32) Camera {
	pitch &= 0x3ff
	yaw &= 0x3ff
	roll &= 0x3ff
	var c Camera
	c.CameraPitch = (1024 - pitch) & 0x3ff
	c.CameraYaw = (1024 - yaw) & 0x3ff
	c.CameraRoll = (1024 - roll) & 0x3ff

	var l1, i2 int32
	j2 := distance
	if pitch != 0 {
		k2 := sine11[pitch]
		j3 := sine11[pitch+1024]
		i4 := (i2*j3 - j2*k2) >> 15
		j2 = (i2*k2 + j2*j3) >> 15
		i2 = i4
	}
	if yaw != 0 {
		l2 := sine11[yaw]
		k3 := sine11[yaw+1024]
		j4 := (j2*l2 + l1*k3) >> 15
		j2 = (j2*k3 - l1*l2) >> 15
		l1 = j4
	}
	if roll != 0 {
		i3 := sine11[roll]
		l3 := sine11[roll+1024]
		k4 := (i2*i3 + l1*l3) >> 15
		i2 = (i2*l3 - l1*i3) >> 15
		l1 = k4
	}
	c.CameraX = x - l1
	c.CameraY = z - i2
	c.CameraZ = y - j2
	return c
}
