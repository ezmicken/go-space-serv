package main

import(
  "go-space-serv/internal/space/snet/udp"
  "go-space-serv/internal/space/player"
)

type SimPlayer struct {
  Stats     player.PlayerStats
  Udp       *udp.UDPPlayer
}
