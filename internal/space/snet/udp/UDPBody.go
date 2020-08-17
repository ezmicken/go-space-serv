package udp

import (
  //"log"
  "github.com/go-gl/mathgl/mgl32"
)

// data
/////////////////////

type UDPBody struct {
  dead              bool

  Id                uint16
  Angle             float32
  TargetAngle       float32
  Position          mgl32.Vec3
  TargetPosition    mgl32.Vec3
  Velocity          mgl32.Vec3
}

func NewUDPBody(id uint16) *UDPBody {
  var bod UDPBody
  bod.dead = false
  bod.Id = id
  bod.Angle = 0
  bod.TargetAngle = 0
  bod.Position = mgl32.Vec3{0, 0, 0}
  bod.TargetPosition = mgl32.Vec3{0, 0, 0}
  bod.Velocity = mgl32.Vec3{0, 0, 0}

  return &bod
}

// access / modify
//////////////////////////////

func (b *UDPBody) Kill() {
  b.dead = true;
}

func (b *UDPBody) IsDead() bool {
  return b.dead
}
