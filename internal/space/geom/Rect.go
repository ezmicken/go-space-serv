package geom

import(
  "github.com/chewxy/math32"
)

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

func RectOverlap(a, b Rect) float32 {
  axMax := a.X + a.W
  ayMax := a.Y + a.H
  bxMax := b.X + b.W
  byMax := b.Y + b.H

  overlapX := math32.Max(a.X, b.X)
  overlapY := math32.Max(a.Y, b.Y)
  overlapW := (math32.Min(axMax, bxMax) - overlapX)
  overlapH := (math32.Min(ayMax, byMax) - overlapY)

  return overlapW * overlapH
}
