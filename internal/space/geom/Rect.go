package geom

type Rect struct {
  X float32
  Y float32
  W float32
  H float32
}

func NewRect(x, y, w, h float32) Rect {
  var r Rect
  r.X = x
  r.Y = y
  r.W = w
  r.H = h
  return r
}
