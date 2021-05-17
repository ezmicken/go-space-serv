//+build windows darwindele
package main

/*
  #include "main.h"
*/
import "C"

import(
  "github.com/ezmicken/spacesim"
  "github.com/ezmicken/fixpoint"
)

func convertBodyInfo(bi C.BodyInfo) spacesim.BodyInfo {
  var result spacesim.BodyInfo = spacesim.BodyInfo {
    Id: uint16(bi.Id),
    Size: int32(bi.Size),
    Proximity: int32(bi.Proximity),
    Lifetime: int32(bi.Lifetime),
    BounceCoefficient: float32(bi.BounceCoefficient),
    VelocityX: float32(bi.VelocityX),
    VelocityY: float32(bi.VelocityY),
  }
  return result
}

var sim *spacesim.Simulation

//export Initialize
func Initialize(ts , scale int32) {
  sim = spacesim.NewSimulation(fixpoint.Q16FromInt32(ts), fixpoint.Q16FromInt32(scale))
}

//export Reset
func Reset() {
  sim.Reset()
}

//export Rewind
func Rewind(seq uint16) {
  sim.Rewind(seq)
}

//export ClearInput
func ClearInput(id uint16) {
  cb := sim.GetControlledBody(id)
  if cb != nil {
    cb.ClearInput()
  }
}

//export PeekSeq
func PeekSeq(id uint16) uint16 {
  cb := sim.GetControlledBody(id)
  if cb != nil {
    ht := cb.PeekState()
    return ht.Seq
  }

  return 0
}

//export ControlBody
func ControlBody(id uint16, rotationSpeed int32, thrust, maxSpeed float32) {
  sim.ControlBody(id, rotationSpeed, thrust, maxSpeed)
}

//export AddBody
func AddBody(id uint16, x, y float32, bodyInfo C.BodyInfo) {
  sim.AddBody(id, x, y, convertBodyInfo(bodyInfo))
}

//export RemoveBody
func RemoveBody(id uint16) {
  sim.RemoveBody(id)
}

//export OverwriteState
func OverwriteState(seq, id, angle, angleDelta uint16, x, y, vx, vy, dvx, dvy int32) {
  sim.OverwriteState(seq, id, angle, angleDelta, x, y, vx, vy, dvx, dvy)
}

//export RemoveControlledBody
func RemoveControlledBody(id uint16) {
  sim.RemoveControlledBody(id)
}

//export PushInput
func PushInput(id uint16, tick uint16, moveshoot byte) {
  cb := sim.GetControlledBody(id)
  if cb != nil {
    cb.PushInput(tick, moveshoot)
  }
}

//export AdvanceSim
func AdvanceSim(seq uint16) {
  sim.Advance(int(seq))
}

//export AdvanceBody
func AdvanceBody(id, seq uint16) {
  sim.AdvanceBody(id, seq)
}

//export AdvanceControlledBody
func AdvanceControlledBody(id, seq uint16) {
  sim.AdvanceControlledBody(id, seq)
}

//export GetNextAngle
func GetNextAngle(id uint16) int32 {
  cb := sim.GetControlledBody(id)
  if cb != nil {
    return cb.GetBody().NextAngle
  }

  return int32(0)
}

//export GetNextPositionX
func GetNextPositionX(id uint16) float32 {
  b := sim.GetBody(id)
  if b != nil {
    return b.NextPos.X.Float()
  }

  return 0
}

//export GetNextPositionY
func GetNextPositionY(id uint16) float32 {
  b := sim.GetBody(id)
  if b != nil {
    return b.NextPos.Y.Float()
  }

  return 0
}

//export GetAngle
func GetAngle(id, seq uint16) int32 {
  cb := sim.GetControlledBody(id)
  if cb != nil {
    return cb.GetAngle(seq)
  }

  return 0
}

//export GetPositionX
func GetPositionX(id, seq uint16) float32 {
  cb := sim.GetControlledBody(id)
  if cb != nil {
    return cb.GetPositionX(seq).Float()
  }

  return 0
}

//export GetPositionY
func GetPositionY(id, seq uint16) float32 {
  cb := sim.GetControlledBody(id)
  if cb != nil {
    return cb.GetPositionY(seq).Float()
  }

  return 0
}

//export GetVelocityX
func GetVelocityX(id, seq uint16) float32 {
  cb := sim.GetControlledBody(id)
  if cb != nil {
    return cb.GetVelocityX(seq).Float()
  }

  return 0
}

//export GetVelocityY
func GetVelocityY(id, seq uint16) float32 {
  cb := sim.GetControlledBody(id)
  if cb != nil {
    return cb.GetVelocityY(seq).Float()
  }

  return 0
}

//export GetMovement
func GetMovement(id uint16) (posX, posY, velX, velY, time float32) {
  b := sim.GetBody(id)
  if b != nil {
    m := b.GetMovement()
    posX = m.PositionX
    posY = m.PositionY
    velX = m.VelocityX
    velY = m.VelocityY
    time = m.Time
    return
  }

  posX = 0.0
  posY = 0.0
  velX = 0.0
  velY = 0.0
  time = 0.0

  return
}

//export AddBlock
func AddBlock(id uint16, x, y int32) {
  b := sim.GetBody(id)
  if b != nil {
    b.AddBlock(x, y)
  }
}

func main(){}
