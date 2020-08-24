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

func (p *WorldPlayer) Update(x, y uint16, wld polyclip.Polygon) {
  p.X = x
  p.Y = y
  return
}
