package geom

type RectInt struct {
  X int
  Y int
  W int
  H int
}

func NewRectInt(x, y, w, h int) RectInt {
  var r RectInt
  r.X = x
  r.Y = y
  r.W = w
  r.H = h
  return r
}
