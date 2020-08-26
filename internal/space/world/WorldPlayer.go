package world

import (
  "log"
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

func (p *WorldPlayer) Update(x, y uint16, worldMap *WorldMap) {
  p.X = x
  p.Y = y

  doubleX := float64(x)
  doubleY := float64(y)

  view := polyclip.Polygon{{
    {doubleX - 32, doubleY - 32},
    {doubleX + 32, doubleY - 32},
    {doubleX + 32, doubleY + 32},
    {doubleX - 32, doubleY + 32},
  }}

  // clip view to world map bounds
  view = worldMap.Poly.Construct(polyclip.INTERSECTION, view)

  unexplored := view.Construct(polyclip.DIFFERENCE, p.explored)

  blocksMsgs := worldMap.SerializeChunks(unexplored)
  log.Printf("%d", len(blocksMsgs))
  for _, m := range blocksMsgs {
    p.Tcp.Outgoing <- &m
  }

  p.explored = p.explored.Construct(polyclip.UNION, view)
  return
}
