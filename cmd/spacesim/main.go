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

//export AddControlledBody
func AddControlledBody(id uint16, x, y int32) {
  sim.AddControlledBody(id, x, y)
}

//export RemoveControlledBody
func RemoveControlledBody(id uint16) {
  sim.RemoveControlledBody(id)
}

//export PushInput
func PushInput(id uint16, tick uint16, moveshoot byte) {
  cb := sim.GetControlledBody(id)
  if cb != nil {
    cb.InputToState(tick, moveshoot)
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
func main(){}
