package phys

import (
	"encoding/binary"
	"bytes"
	"log"

	"go-space-serv/internal/app/snet"

	//"github.com/stojg/vector"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/bgadrian/data-structures/priorityqueue"
)

// data
/////////////////////

type UdpBody struct {
	// these are used only by server
	controllingPlayer *UdpPlayer
	owningPlayer 			*UdpPlayer
	inputs            *priorityqueue.HierarchicalQueue
	dead              bool

	// these are sent to clients
	// these are listed by priority in ascending order
	id   uint16
	xPos float32
	yPos float32
	xVel float32
	yVel float32
	rot  uint16
	xAcc float32
	yAcc float32
}

func (b *UdpBody) Serialize() []byte {
	var buf bytes.Buffer

	binary.Write(&buf, binary.LittleEndian, b.id)
	binary.Write(&buf, binary.LittleEndian, b.xPos)
	binary.Write(&buf, binary.LittleEndian, b.yPos)
	binary.Write(&buf, binary.LittleEndian, b.xVel)
	binary.Write(&buf, binary.LittleEndian, b.yVel)

	return buf.Bytes()
}

// instantiation
///////////////////////

func NewControlledUdpBody(player *UdpPlayer) (*UdpBody) {
	bod := NewUdpBody();
	bod.controllingPlayer = player
	bod.owningPlayer = player
	bod.dead = false;

	bod.inputs = priorityqueue.NewHierarchicalQueue(255, true)
	return bod
}

func NewUdpBody() (*UdpBody) {
	var b UdpBody
	b.id = snet.GetNextId()
	b.xVel = 0
	b.yVel = 0
	b.xAcc = 0
	b.yAcc = 0
	// xPos, yPos set when player spawn

	b.dead = false;

	return &b
}

// actions
//////////////////////

func (b *UdpBody) QueueInput(i *UdpInput) {
	b.inputs.Enqueue(i, i.seq)
}

func (b *UdpBody) DequeueInput() *UdpInput {
	if b.inputs.Len() > 0 {
		i, err := b.inputs.Dequeue()
		if err != nil {
			panic(err)
		}
		input, ok := i.(*UdpInput)
		if !ok {
			return nil
		}

		return input
	}

	return nil
}

func (b *UdpBody) ProcessInput() {
	// TODO: figure out how priority comes into play here
	for i := b.DequeueInput(); i != nil; {
		// handle input
		if (i.GetType() == MOVE) {
			// MOVE = type | value | type2 | value2 | ...
			// 0 + 0 = thrust forward
			// 0 + 1 = thrust backward
			// 1 + 0 = rotate left
			// 1 + 1 = rotate right
			data := i.GetContent()
			dataLen := len(data)
			accel := float32(0.0)
			stats := b.controllingPlayer.GetStats()
			remaining := dataLen

			for j := 0; remaining > 0; j++ {
				movementType := data[j]

				if movementType == 0 {
					movementVal := data[j+1]
					movementAngle := data[j+2:j+4]

					if movementVal == 0 {
						accel = stats.Acceleration
					} else if movementVal == 1 {
						accel = -stats.Acceleration
					}

					b.rot = snet.Read_uint16(movementAngle)

					remaining -= 4
				} else {
					log.Printf("Received movementType=%d from %s", movementType, b.controllingPlayer.GetName())
				}
			}

			// TODO: apply rotation
			if accel != 0 {
				q := mgl32.AnglesToQuat(mgl32.DegToRad(float32(b.rot)), 0, 0, mgl32.ZYX).Normalize()

				// Apply force along the Y axis
				accVec := mgl32.Vec3{0, accel, 0}
				velVec := mgl32.Vec3{b.xVel, b.yVel, 0}

				// Rotate & Add the acceleration vector to velocity
				newVel := velVec.Add(q.Rotate(accVec))

				// Clamp velocity to max speed
				if newVel.LenSqr() > (stats.MaxSpeed * stats.MaxSpeed) {
					newVel = newVel.Normalize().Mul(stats.MaxSpeed)
				}
				// Update velocity
				b.SetVel(newVel[0], newVel[1])

				// detect and handle world collision

				// detect and handle body collision
					// get objects near this one
					// add to list of object groups to check for collision
			}

			i = b.DequeueInput()
		}
	}
}

// TODO: apply rotation?
func (b *UdpBody) ApplyTransform() {
	b.xPos = b.xPos + b.xVel
	b.yPos = b.yPos + b.yVel
}

// access / modify
//////////////////////////////

func (b *UdpBody) SetControllingPlayer(player *UdpPlayer) {
	b.controllingPlayer = player
}

func (b *UdpBody) GetControllingPlayer() *UdpPlayer {
	return b.controllingPlayer
}

func (b *UdpBody) SetOwningPlayer(player *UdpPlayer) {
	b.owningPlayer = player
}

func (b *UdpBody) GetOwningPlayer() *UdpPlayer {
	return b.owningPlayer
}

func (b *UdpBody) Kill() {
	b.dead = true;
}

func (b *UdpBody) IsDead() bool {
	return b.dead
}

func (b *UdpBody) GetId() uint16 {
	return b.id
}

func (b *UdpBody) SetRot(r uint16) {
	b.rot = r
}

func (b *UdpBody) GetRot() uint16 {
	return b.rot
}

func (b *UdpBody) SetPos(x, y float32) {
	b.xPos = x
	b.yPos = y
}
func (b *UdpBody) GetPos() (x, y float32) {
	x = b.xPos
	y = b.yPos
	return
}

func (b *UdpBody) SetVel(x, y float32) {
	b.xVel = x
	b.yVel = y
}
func (b *UdpBody) GetVel() (x, y float32) {
	x = b.xVel
	y = b.yVel
	return
}

func (b *UdpBody) SetAcc(x, y float32) {
	b.xAcc = x
	b.yAcc = y
}
func (b *UdpBody) GetXAcc() (x, y float32) {
	x = b.xAcc
	y = b.yAcc
	return
}
