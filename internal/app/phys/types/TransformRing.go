package phys

import (
  "log"

  "go-space-serv/internal/app/helpers"
)

type TransformRing struct {
  nextFree int
  buffer []HistoricalTransform
  initialized bool
}

func NewTransformRing(length int) *TransformRing {
  var r TransformRing
  r.nextFree = 0
  r.buffer = make([]HistoricalTransform, length)
  r.initialized = false
  return &r
}

func (r *TransformRing) wrappedIndex(idx int) int {
  l := len(r.buffer)
  for idx >= l { idx -= l }
  for idx < 0 { idx += l }
  return idx;
}

func (r *TransformRing) initialize(ht HistoricalTransform) {
  if r.initialized { return }

  current := r.wrappedIndex(r.nextFree - 1)
  currentTime := ht.Timestamp
  timestep := helpers.GetConfiguredTimestep()

  for k := 0; k < len(r.buffer); k++ {
    ht.Timestamp = currentTime
    r.buffer[current] = ht
    current = r.wrappedIndex(current - 1)
    currentTime -= timestep
  }

  r.initialized = true
}

func (r *TransformRing) Advance(ht HistoricalTransform) {
  if (!r.initialized) {
    r.initialize(ht)
  } else {
    r.buffer[r.nextFree] = ht
    r.nextFree = r.wrappedIndex(r.nextFree + 1)
  }
}

func (r *TransformRing) Insert(ht HistoricalTransform) {
  current := r.wrappedIndex(r.nextFree - 1)
  l := len(r.buffer)

  if r.buffer[current].Timestamp == ht.Timestamp {
    r.buffer[current] = ht
    return
  }

  if r.buffer[current].Timestamp > ht.Timestamp {
    i := r.wrappedIndex(current - 1)
    k := 0

    for k < l {
      if r.buffer[i].Timestamp == ht.Timestamp {
        r.buffer[i] = ht
        return
      }
      k++
      i = r.wrappedIndex(i - 1)
    }
  }
}

func (r *TransformRing) GetTransformAt(time int64) (xPos, yPos, xVel, yVel, angle float32){
  i := r.nextFree
  k := 0
  var found bool = false
  l := len(r.buffer)

  for !found && k < l {
    i = r.wrappedIndex(i - 1);
    found = r.buffer[i].Timestamp <= time && r.buffer[i].Timestamp != 0
    k++
  }

  xPos = r.buffer[i].XPos;
  yPos = r.buffer[i].YPos;
  xVel = r.buffer[i].XVel;
  yVel = r.buffer[i].YVel;
  angle = r.buffer[i].Angle;

  return
}

func (r *TransformRing) IsConsistent() bool {
  current := r.wrappedIndex(r.nextFree - 1)
  currentTime := r.buffer[current].Timestamp
  k := 0
  l := len(r.buffer)
  for k < l && current != r.nextFree {
    if r.buffer[current].Timestamp != currentTime {
      return false
    }
    currentTime -= helpers.GetConfiguredTimestep()
    current = r.wrappedIndex(current - 1)
    k++
  }
  return true
}








