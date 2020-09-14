package world

import (
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
    {doubleX - 32, doubleY - 32},
    {doubleX + 32, doubleY - 32},
    {doubleX + 32, doubleY + 32},
    {doubleX - 32, doubleY + 32},
  }}

  // clip view to world map bounds
  view = worldMap.Poly.Construct(polyclip.INTERSECTION, view)

  // Get Contour(s) that exist in view but not in
  // the explored polygon.
  unexplored := view.Construct(polyclip.DIFFERENCE, p.explored)

  // Clamp these contours to chunk size.
  // Serialize the chunks and send to player.
  blocksMsgs := worldMap.SerializeChunks(unexplored)
  for _, m := range blocksMsgs {
    p.Tcp.Outgoing <- &m
  }

  // Add unexplored to explored.
  p.explored = p.explored.Construct(polyclip.UNION, view)
  return
}
