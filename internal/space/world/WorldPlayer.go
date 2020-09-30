package world

import (
  //"log"
  "github.com/akavel/polyclip-go"
  "github.com/google/uuid"
  "go-space-serv/internal/space/snet/tcp"
  "go-space-serv/internal/space/player"
)

type WorldPlayer struct {
  Tcp       *tcp.TCPPlayer
  Stats     player.PlayerStats
  Id        uuid.UUID
  X         uint16
  Y         uint16

  explored  polyclip.Polygon
  view      polyclip.Polygon
}

// TODO: make sure the previous update finished first!
func (p *WorldPlayer) Update(x, y uint16, worldMap *WorldMap) {
  p.X = x
  p.Y = y

  doubleX := float64(x)
  doubleY := float64(y)

  // Construct a square representing
  // the player's view.
  // TODO: use player view range stat
  view := polyclip.Polygon{{
    {doubleX - 128, doubleY - 128},
    {doubleX + 128, doubleY - 128},
    {doubleX + 128, doubleY + 128},
    {doubleX - 128, doubleY + 128},
  }}

  // clamp view to world
  view = worldMap.Poly.Construct(polyclip.INTERSECTION, view)
  if view.NumVertices() > 0 {
    viewRect := view.BoundingBox()
    viewRect = worldMap.ClampToChunks(viewRect)
    view = polyclip.Polygon{{
      {viewRect.Min.X, viewRect.Max.Y},
      {viewRect.Min.X, viewRect.Min.Y},
      {viewRect.Max.X, viewRect.Min.Y},
      {viewRect.Max.X, viewRect.Max.Y},
    }}
    unexplored := view.Construct(polyclip.DIFFERENCE, p.explored)
    p.explored = p.explored.Construct(polyclip.UNION, unexplored)

    if unexplored.NumVertices() >= 4 {
      blocksMsgs := worldMap.Explore(unexplored.BoundingBox())
      for _, m := range blocksMsgs {
        msg := m
        p.Tcp.Outgoing <- &msg
      }
    }
  }

  return
}
