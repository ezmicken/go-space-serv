package sim

import(
  "fmt"
  "github.com/go-gl/mathgl/mgl32"
)

type HistoricalTransform struct {
  Seq           int
  Angle         float32
  AngleDelta    float32
  Position      mgl32.Vec3
  Velocity      mgl32.Vec3
  VelocityDelta mgl32.Vec3
}

func (ht *HistoricalTransform) String() string {
  return fmt.Sprintf("%v: %v - %v - %v - %v - %v", ht.Seq, ht.Position, ht.Angle, ht.AngleDelta, ht.Velocity, ht.VelocityDelta)
}
