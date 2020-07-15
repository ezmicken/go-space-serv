package helpers

import (
  "time"
)

func PerSecondOverTime(stat float32, dur ...int64) float32 {
  if len(dur) > 1 {
    return stat * float32(dur[1] - dur[0]) / float32(1000)
  } else {
    return stat * (float32(dur[0]) / float32(1000))
  }
}

func WrapFloat32(val, min, length float32) float32 {
  for val >= length { val -= length }
  for val < min { val += length }
  return val;
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
