package phys

import (
  "encoding/binary"
  "bytes"
  "log"

  "go-space-serv/internal/app/snet"
  "go-space-serv/internal/app/helpers"

  "github.com/go-gl/mathgl/mgl32"
  "github.com/bgadrian/data-structures/priorityqueue"
)

// data
/////////////////////

type UdpBody struct {
  // these are used only by server
  controllingPlayer *UdpPlayer
  owningPlayer      *UdpPlayer
  inputs            *priorityqueue.HierarchicalQueue
  dead              bool

  history           *TransformRing
  // these are sent to clients
  // these are listed by priority in ascending order
  id   uint16
  xPos float32
  yPos float32
  xVel float32
  yVel float32
  rot  float32
  xAcc float32
  yAcc float32
}

const LEFT byte = 1;
const RIGHT byte = 2;
const FORWARD byte = 3;
const BACKWARD byte = 4;

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
  bod.history = NewTransformRing(1024)
  bod.inputs = priorityqueue.NewHierarchicalQueue(255, true)

  bod.xPos = 0
  bod.yPos = 0
  bod.xVel = 0
  bod.yVel = 0
  bod.rot = 0

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

func (b *UdpBody) ProcessInput(seq uint16, frameStart int64) {
  if !b.history.Initialized {
    initHistory := b.GetHistoricalTransform()
    initHistory.Timestamp = frameStart
    b.history.Initialize(initHistory)
  }

  for i := b.DequeueInput(); i != nil; {
    if (i.GetType() == MOVE) {
      data := i.GetContent()
      dataLen := len(data)
      accel := float32(0.0)
      stats := b.controllingPlayer.GetStats()
      offset := 0
      redundancy := data[0];
      offset++;

      currentSeq := i.GetSeq() - uint16(redundancy)
      currentTime := helpers.SeqToMillis(currentSeq, b.controllingPlayer.lastSync)
      currentXPos, currentYPos, currentXVel, currentYVel, currentAngle := b.history.GetTransformAt(currentTime - helpers.GetConfiguredTimestep())
      var currentVel mgl32.Vec3
      currentVel[0] = currentXVel
      currentVel[1] = currentYVel
      var act byte
      var actDur int64

      // Process this input msg
      for offset < dataLen && currentSeq < seq {
        actCount := data[offset]
        offset++

        for a := byte(0); a < actCount; a++ {
          act = data[offset]
          offset++
          actDur = snet.Read_int64(data[offset:offset+8])
          offset += 8
          if act == LEFT {
            currentAngle += helpers.PerSecondOverTime(stats.Rotation, actDur)
          }
          if act == RIGHT {
            currentAngle -= helpers.PerSecondOverTime(stats.Rotation, actDur)
          }

          currentAngle = helpers.WrappedAngle(currentAngle)

          if act == FORWARD {
            accel += helpers.PerSecondOverTime(stats.Thrust, actDur)
          }
          if act == BACKWARD {
            accel -= helpers.PerSecondOverTime(stats.Thrust, actDur)
          }
        }

        if accel != 0 {
          q := mgl32.AnglesToQuat(mgl32.DegToRad(currentAngle), 0, 0, mgl32.ZYX).Normalize()

          // Apply force along the Y axis
          accVec := mgl32.Vec3{0, accel, 0}
          velVec := mgl32.Vec3{currentXVel, currentYVel, 0}

          // Rotate & Add the acceleration vector to velocity
          currentVel = velVec.Add(q.Rotate(accVec))
        }

        // Clamp velocity to max speed
        if currentVel.LenSqr() > (stats.MaxSpeed * stats.MaxSpeed) {
          currentVel = currentVel.Normalize().Mul(stats.MaxSpeed)
        }

        // Update velocity
        currentXPos += currentVel[0]
        currentYPos += currentVel[1]
        currentXVel = currentVel[0]
        currentYVel = currentVel[1]

        var crumb HistoricalTransform
        crumb.Timestamp = currentTime
        crumb.Angle = currentAngle
        crumb.XPos = currentXPos
        crumb.YPos = currentYPos
        crumb.XVel = currentXVel
        crumb.YVel = currentYVel
        b.history.Insert(crumb)

        accel = float32(0)

        // detect and handle world collision

        // detect and handle body collision
        // get objects near this one
        // add to list of object groups to check for collision
        currentSeq++
        currentTime = helpers.SeqToMillis(currentSeq, b.controllingPlayer.lastSync)
      }

      // apply changes to history
      for currentTime < frameStart {
        // Update velocity
        currentXPos += currentVel[0]
        currentYPos += currentVel[1]

        // collision check

        var crumb HistoricalTransform
        crumb.Timestamp = currentTime
        crumb.Angle = currentAngle
        crumb.XPos = currentXPos
        crumb.YPos = currentYPos
        crumb.XVel = currentXVel
        crumb.YVel = currentYVel
        b.history.Insert(crumb)

        currentSeq++
        currentTime = helpers.SeqToMillis(currentSeq, b.controllingPlayer.lastSync)
      }

      b.rot = currentAngle

      b.SetVel(currentVel[0], currentVel[1])

      i = b.DequeueInput()
    }
  }
}

// TODO: apply rotation?
func (b *UdpBody) ApplyTransform(frameStart int64) {
  b.xPos = b.xPos + b.xVel
  b.yPos = b.yPos + b.yVel

  crumb := b.GetHistoricalTransform()
  crumb.Timestamp = frameStart
  b.history.Advance(crumb)
}

func (b *UdpBody) GetHistoricalTransform() HistoricalTransform {
  var ht HistoricalTransform

  ht.Timestamp = helpers.NowMillis()
  ht.Angle = b.rot
  ht.XPos = b.xPos
  ht.YPos = b.yPos
  ht.XVel = b.xVel
  ht.YVel = b.yVel

  return ht
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

func (b *UdpBody) SetRot(r float32) {
  b.rot = r
}

func (b *UdpBody) GetRot() float32 {
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
