package main

import(
  "C"
  "github.com/ezmicken/spacesim"
  "github.com/ezmicken/fixpoint"
)

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

//export AddControlledBody
func AddControlledBody(id uint16, x, y, d int32) {
  sim.AddControlledBody(id, x, y, d)
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

//export AddBlock
func AddBlock(id uint16, x, y int32) {
  cb := sim.GetControlledBody(id)
  if cb != nil {
    cb.AddBlock(x, y)
  }
}

func main(){}
