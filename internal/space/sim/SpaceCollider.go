package sim

import(
  //"log"
  "github.com/go-gl/mathgl/mgl32"
  "github.com/chewxy/math32"
  "go-space-serv/internal/space/geom"
)

type SpaceCollider struct {
  broadSize       int
  narrowSize      int
  halfBroadSize   int
  halfNarrowSize  int

  Broad           geom.Rect
  Narrow          geom.Rect
}

type collision struct {
  Time float32
  Normal mgl32.Vec3
  Area float32
}

func NewSpaceCollider(broadSize, narrowSize int) *SpaceCollider {
  var c SpaceCollider
  c.broadSize = broadSize
  c.narrowSize = narrowSize
  c.halfBroadSize = broadSize/2
  c.halfNarrowSize = narrowSize/2

  c.Broad = geom.NewRect(float32(-c.halfBroadSize), float32(-c.halfBroadSize), float32(broadSize), float32(broadSize))
  c.Narrow = geom.NewRect(float32(-c.halfNarrowSize), float32(-c.halfNarrowSize), float32(narrowSize), float32(narrowSize))
  return &c
}

func (sc *SpaceCollider) Update(position, velocity mgl32.Vec3) {
  sc.Broad.X = position.X() - float32(sc.halfBroadSize)
  sc.Broad.Y = position.Y() - float32(sc.halfBroadSize)

  if velocity.X() > 0 { sc.Broad.X -= velocity.X() }
  if velocity.Y() > 0 { sc.Broad.Y -= velocity.Y() }
  sc.Broad.W = float32(sc.broadSize) + velocity.X()
  sc.Broad.H = float32(sc.broadSize) + velocity.Y()

  sc.Narrow.X = position.X() - float32(sc.halfNarrowSize)
  sc.Narrow.Y = position.Y() - float32(sc.halfNarrowSize)
}

func (sc *SpaceCollider) Check(ht HistoricalTransform, potentialCollisions []geom.Rect) (HistoricalTransform, int) {
  var closest collision
  closest.Time = float32(1.0)
  closest.Normal = mgl32.Vec3{0, 0, 0}
  closest.Area = 0.0
  closestIdx := -1

  for i := 0; i < len(potentialCollisions); i++ {
    col := sc.sweep(ht.Velocity, potentialCollisions[i])
    if col.Area > closest.Area {
      closest = col
      closestIdx = i
    }
  }

  pos := ht.Position
  vel := ht.Velocity

  remainingTime := float32(1.0) - closest.Time
  if remainingTime > 0.0 {
    if math32.Abs(closest.Normal[0]) > 0.001 {
      if math32.Abs(vel[0]) < 1.0 {
        pos[0] += (vel[0] * closest.Time)
        vel[0] = 0.0
      } else {
        pos[0] += (vel[0] * closest.Time)
        vel[0] *= float32(-0.5)
        pos[0] += (vel[0] * remainingTime)
      }
    }

    if math32.Abs(closest.Normal[1]) > 0.001 {
      if math32.Abs(vel[1]) < 1.0 {
        pos[1] += (vel[0] * closest.Time)
        vel[1] = 0.0
      } else {
        pos[1] += (vel[1] * closest.Time)
        vel[1] *= float32(-0.5)
        pos[1] += (vel[1] * remainingTime)
      }
    }

    ht.VelocityDelta = vel.Sub(ht.Velocity.Sub(ht.VelocityDelta))
    ht.Position = pos
    ht.Velocity = vel

    sc.Update(ht.Position, ht.Velocity)
  }

  return ht, closestIdx
}

func (sc *SpaceCollider) sweep(velocity mgl32.Vec3, block geom.Rect) collision {
  var dxEntry float32
  var dxExit float32
  var dyEntry float32
  var dyExit float32
  var result collision

  if velocity.X() > 0 {
    dxEntry = block.X - (sc.Narrow.X + sc.Narrow.W)
    dxExit = (block.X + block.W) - sc.Narrow.X
  } else {
    dxEntry = (block.X + block.W) - sc.Narrow.X
    dxExit = block.X - (sc.Narrow.X + sc.Narrow.W)
  }

  if velocity.Y() > 0 {
    dyEntry = block.Y - (sc.Narrow.Y + sc.Narrow.H)
    dyExit = (block.Y + block.H) - sc.Narrow.Y
  } else {
    dyEntry = (block.Y + block.H) - sc.Narrow.Y
    dyExit = block.Y - (sc.Narrow.Y + sc.Narrow.H)
  }

  var txEntry float32
  var txExit float32
  var tyEntry float32
  var tyExit float32

  if velocity.X() == 0 {
    txEntry = math32.Inf(-1)
    txExit = math32.Inf(1)
  } else {
    txEntry = dxEntry / velocity.X()
    txExit = dxExit / velocity.X()
  }

  if velocity.Y() == 0 {
    tyEntry = math32.Inf(-1)
    tyExit = math32.Inf(1)
  } else {
    tyEntry = dyEntry / velocity.Y()
    tyExit = dyExit / velocity.Y()
  }

  entryTime := math32.Max(txEntry, tyEntry)
  exitTime := math32.Min(txExit, tyExit)

  exiting := entryTime > exitTime
  negativeEntry := (txEntry < 0.0 && tyEntry < 0.0)
  futureEntry := txEntry > 1.0 || tyEntry > 1.0

  if exiting || negativeEntry || futureEntry {
    result.Time = 1.0
    result.Normal = mgl32.Vec3{0, 0, 0}
  } else {
    result.Time = entryTime
    movedBody := geom.NewRect(sc.Narrow.X + velocity[0], sc.Narrow.Y + velocity[1], sc.Narrow.W, sc.Narrow.H)
    result.Area = geom.RectOverlap(movedBody, block)
    if txEntry > tyEntry {
      if dxEntry < 0.0 {
        result.Normal = mgl32.Vec3{1, 0, 0}
      } else {
        result.Normal = mgl32.Vec3{-1, 0, 0}
      }
    } else {
      if dyEntry < 0.0 {
        result.Normal = mgl32.Vec3{0, 1, 0}
      } else {
        result.Normal = mgl32.Vec3{0, -1, 0}
      }
    }
  }

  return result
}
