package phys

import (
  "encoding/binary"
  "bytes"
  //"log"

  "go-space-serv/internal/app/snet"

  //"github.com/go-gl/mathgl/mgl32"
  "github.com/bgadrian/data-structures/priorityqueue"

  . "go-space-serv/internal/app/snet/types"
  . "go-space-serv/internal/app/phys/types"
)

// data
/////////////////////

type UdpBody struct {
  // these are used only by server
  controllingPlayer *UdpPlayer
  owningPlayer      *UdpPlayer
  inputs            *priorityqueue.HierarchicalQueue
  dead              bool

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

func (b *UdpBody) HasInput() bool {
  return b.inputs.Len() > 0;
}

// Input data is as follows
// | redundancy | (actCount | (act id | act duration |) * actCount |) * redundancy |
// |   byte     |    byte   |  byte   |     int64    | ...
//
// Output is as follows
// | seq | bodyId | validated copy of input... |
func (b *UdpBody) ProcessInput(seq uint16, frameStart int64) *NetworkMsg {
  i := b.DequeueInput()
  if i != nil {
    var outputMsg NetworkMsg
    outputMsg.PutByte(byte(SFrame))
    outputMsg.PutUint16(b.id)
    outputMsg.PutBytes(i.GetContent())
    return &outputMsg
  }

  return nil
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
