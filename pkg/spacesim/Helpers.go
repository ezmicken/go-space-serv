package main

import(
  //"log"
  "github.com/ezmicken/fixpoint"
)

func WrapQ16(val, min, length fixpoint.Q16) fixpoint.Q16 {
  for val.N >= length.N { val = val.Sub(length) }
  for val.N< min.N { val = val.Add(length) }
  return val;
}

func WrapAngle(val fixpoint.Q16) fixpoint.Q16 {
  return WrapQ16(val, fixpoint.Q16FromFloat(float32(0)), fixpoint.Q16FromFloat(float32(360)))
}

