package main

import(
  "github.com/ezmicken/fixpoint"
)

type HistoricalTransform struct {
  Seq           int
  Angle         fixpoint.Q16
  AngleDelta    fixpoint.Q16
  Position      fixpoint.Vec3Q16
  Velocity      fixpoint.Vec3Q16
  VelocityDelta fixpoint.Vec3Q16
}

