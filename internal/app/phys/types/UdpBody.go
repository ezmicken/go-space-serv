package phys

import (
	"encoding/binary"
	"bytes"
	//"log"

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
	rot  int16
	xAcc float32
	yAcc float32
}

const LEFT byte = 0;
const RIGHT byte = 1;
const FORWARD byte = 2;
const BACKWARD byte = 3;

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
	b.rot = 0
	// xPos, yPos set when player spawn

	b.dead = false;

	return &b
}

// actions
//////////////////////

func (b *UdpBody) QueueInput(i *UdpInput) {
	b.inputs.Enqueue(i, 1)
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

func readInput(data []byte, offset int) (byte, int64, int64) {
	movementType := data[offset];
	start := snet.Read_int64(data[offset+1:offset+9])
	end := snet.Read_int64(data[offset+9:offset+17])

	return movementType, start, end
}

func applyStat(stat float32, frameStart int64, inputStart int64, inputEnd int64) float32 {
	result := float32(0.0)

	if inputEnd != 0 {
		result += (stat * (float32(inputEnd - inputStart) / float32(1000.0)))
	} else if (frameStart > inputStart) {
		result += (stat * (float32(frameStart - inputStart) / float32(1000.0)))
	}

	return result
}

func (b *UdpBody) ProcessInput(frameStart int64) {
	for i := b.DequeueInput(); i != nil; {
		if (i.GetType() == MOVE) {
			data := i.GetContent()
			dataLen := len(data)
			accel := float32(0.0)
			degrees := float32(0.0)
			stats := b.controllingPlayer.GetStats()
			remaining := dataLen

			for j := 0; remaining > 0; j++ {
				movementType, start, end := readInput(data, j)
				remaining -= 17

				if movementType == LEFT {
					//log.Printf("%d -- %d -- %d", frameStart, start, end);
					degrees += applyStat(stats.Rotation, frameStart, start, end)
				}
				if movementType == RIGHT {
					degrees -= applyStat(stats.Rotation, frameStart, start, end)
				}
				if movementType == FORWARD {
					accel += applyStat(stats.Thrust, frameStart, start, end)
				}
				if movementType == BACKWARD {
					accel -= applyStat(stats.Thrust, frameStart, start, end)
				}
			}

			// clamp b.rot to 0 - 360
      if degrees != 0 {
        b.rot += int16(degrees);
        if b.rot > 360 {
          b.rot -= 360;
        } else if b.rot < 0 {
          b.rot += 360;
        }
      }

      //log.Printf("%d", b.rot)

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

func (b *UdpBody) SetRot(r int16) {
	b.rot = r
}

func (b *UdpBody) GetRot() int16 {
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
