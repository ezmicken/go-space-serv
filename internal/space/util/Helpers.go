package helpers

import (
  "time"
)

func PerSecondOverTime(stat float32, dur int64) float32 {
  return stat * (float32(dur) / float32(1000))
}

func WrapFloat32(val, min, length float32) float32 {
  for val >= length { val -= length }
  for val < min { val += length }
  return val;
}

func WrapAngle(val float32) float32 {
  return WrapFloat32(val, 0, 360)
}

func WrapInt(val, min, length int) int {
  for val >= length { val -= length }
  for val < min { val += length }
  return val;
}

func NanosToMillis(nanos int64) int64 {
  return nanos / (int64(time.Millisecond) / int64(time.Nanosecond))
}

func NowMillis() int64 {
  return time.Now().UnixNano() / (int64(time.Millisecond)/int64(time.Nanosecond))
}

func SeqToMillis(seq uint16, lastSync int64) int64 {
  return NanosToMillis(lastSync + (int64(seq) * GetConfiguredTimestepNanos()))
}

// doesnt always work
func MillisToSeq(timestamp int64, lastSync int64) uint16 {
  return (uint16)((lastSync - timestamp) / GetConfiguredTimestep() * -1);
}

func BitOn(b byte, pos int) bool {
  return (b & (1 << pos)) != 0;
}

func BitString(bytes []byte) string {
  var dbg string = "["
  for l := 0; l < len(bytes); l++ {
    for m := 0; m < 8; m++ {
      if BitOn(bytes[l], m) {
        dbg += "1"
      } else {
        dbg += "0"
      }
    }
    dbg += " "
  }
  dbg += "]"

  return dbg
}

// Thanks Craig McQueen
// https://stackoverflow.com/questions/1100090/looking-for-an-efficient-integer-square-root-algorithm-for-arm-thumb2
func Sqrt_uint32(op uint32) uint32 {
  res := uint32(0)
  one64 := uint64(1)
  one := uint32(one64 << 30)

  for one > op {
    one >>= 2
  }

  for one != 0 {
    if op >= res + one {
      op = op - (res + one)
      res = res + 2 * one
    }
    res >>= 1
    one >>= 2
  }

  if op > res {
    res++
  }

  return res
}
