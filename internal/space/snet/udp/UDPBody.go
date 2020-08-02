package udp

import (
  "encoding/binary"
  "bytes"
  //"log"

  "go-space-serv/internal/space/snet"

  //"github.com/go-gl/mathgl/mgl32"
)

// data
/////////////////////

type UDPBody struct {
  // these are used only by server
  controllingPlayer *UDPPlayer
  owningPlayer      *UDPPlayer
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

func (b *UDPBody) Serialize() []byte {
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

func NewControlledBody(player *UDPPlayer) (*UDPBody) {
  bod := NewUDPBody();
  bod.controllingPlayer = player
  bod.owningPlayer = player
  bod.dead = false;

  bod.xPos = 0
  bod.yPos = 0
  bod.xVel = 0
  bod.yVel = 0
  bod.rot = 0

  return bod
}

func NewUDPBody() (*UDPBody) {
  var b UDPBody
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

// access / modify
//////////////////////////////

func (b *UDPBody) SetControllingPlayer(player *UDPPlayer) {
  b.controllingPlayer = player
}

func (b *UDPBody) GetControllingPlayer() *UDPPlayer {
  return b.controllingPlayer
}

func (b *UDPBody) SetOwningPlayer(player *UDPPlayer) {
  b.owningPlayer = player
}

func (b *UDPBody) GetOwningPlayer() *UDPPlayer {
  return b.owningPlayer
}

func (b *UDPBody) Kill() {
  b.dead = true;
}

func (b *UDPBody) IsDead() bool {
  return b.dead
}

func (b *UDPBody) GetId() uint16 {
  return b.id
}

func (b *UDPBody) SetRot(r float32) {
  b.rot = r
}

func (b *UDPBody) GetRot() float32 {
  return b.rot
}

func (b *UDPBody) SetPos(x, y float32) {
  b.xPos = x
  b.yPos = y
}
func (b *UDPBody) GetPos() (x, y float32) {
  x = b.xPos
  y = b.yPos
  return
}

func (b *UDPBody) SetVel(x, y float32) {
  b.xVel = x
  b.yVel = y
}
func (b *UDPBody) GetVel() (x, y float32) {
  x = b.xVel
  y = b.yVel
  return
}

func (b *UDPBody) SetAcc(x, y float32) {
  b.xAcc = x
  b.yAcc = y
}
func (b *UDPBody) GetXAcc() (x, y float32) {
  x = b.xAcc
  y = b.yAcc
  return
}
