package main

import(
  "testing"
  "github.com/ezmicken/fixpoint"
)

func TestWrapAngle(t *testing.T) {
  var v fixpoint.Q16
  for i := 1; i < 360; i++ {
    v = WrapAngle(fixpoint.Q16FromFloat(float32(360) + float32(i)))
    if v != fixpoint.Q16FromFloat(float32(i)) {
      t.Logf("Failed! %v is not %v", v, float32(i))
      t.Fail()
    }
    v = WrapAngle(fixpoint.Q16FromFloat(float32(0) - float32(i)))
    if v != fixpoint.Q16FromFloat(float32(360 - i)) {
      t.Logf("Failed! %v is not %v", v, float32(360 - i))
      t.Fail()
    }
  }
}
