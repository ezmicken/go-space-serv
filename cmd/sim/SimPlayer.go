package main

import(
  "time"
  "go-space-serv/internal/space/snet/udp"
  "go-space-serv/internal/space/sim/msg"
  "go-space-serv/internal/space/player"
)

type SimPlayer struct {
  Stats           player.PlayerStats
  Udp             *udp.UDPPlayer

  lastActiveTime  int64
  lastPingTime    int64
  lastActiveState bool

  state           int
}

const PING_TIME   int64 = 200000000
const SPECTATING  int   = 0
const PLAYING     int   = 1
const CULL        int   = 2

func NewSimPlayer(stats player.PlayerStats, udpPlayer *udp.UDPPlayer) *SimPlayer {
  var plr SimPlayer
  plr.Stats = stats
  plr.Udp = udpPlayer

  plr.lastActiveTime = 0
  plr.lastActiveState = false
  plr.state = SPECTATING

  return &plr
}

func (s *SimPlayer) GetLastPingTime() int64 {
  return s.lastPingTime
}

func (s *SimPlayer) IsSpectating()  bool { return s.state == SPECTATING }
func (s *SimPlayer) IsPlaying()     bool { return s.state == PLAYING }
func (s *SimPlayer) ShouldCull()    bool { return s.state == CULL }

func (s *SimPlayer) OnEnter() {
  s.state = PLAYING
}

func (s *SimPlayer) OnExit() {
  s.state = SPECTATING
}

func (s *SimPlayer) OnDisconnect() {
  s.state = CULL
  s.Udp.Disconnect()
}

func (s *SimPlayer) OnPackAndSend() {
  // Send a ping after 500ms of inactivity.
  if s.Udp.IsActive() && !s.lastActiveState {
    s.lastActiveState = true
  } else if !s.Udp.IsActive() && s.lastActiveState {
    s.lastActiveState = false
    s.lastActiveTime = time.Now().UnixNano()
  } else if !s.Udp.IsActive() {
    now := time.Now().UnixNano()
    if now - s.lastActiveTime > PING_TIME {
      s.lastActiveTime = now
      s.lastPingTime = now
      var ping msg.CmdMsg
      ping.SetCmd(udp.PING)
      s.Udp.Outgoing <- &ping
    }
  }
}
