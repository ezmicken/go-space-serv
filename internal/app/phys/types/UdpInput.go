package phys

type UdpInputType byte

const (
	// MOVEMENT
	ACCEL UdpInputType = iota + 1
	ORIENT

	// ABILITIES
	SHOOT_GUN
	SHOOT_BOMB
)

type UdpInput struct {
	seq byte
	size int
	iType UdpInputType
	content []byte
}
