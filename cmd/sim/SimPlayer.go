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

  bodyIdPool      []uint16
  bodyIdIdx       int
  bodyIdRange     int
  rangeStart      uint16
  rangeEnd        uint16

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
  s.bodyIdPool = make([]uint16, s.rangeEnd - s.rangeStart)
  for i := 0; i < s.bodyIdRange; i++ {
    s.bodyIdPool[i] = s.rangeStart + uint16(i)
  }
  s.bodyIdIdx = 0
}

func (s *SimPlayer) OnExit() {
  s.state = SPECTATING
  s.bodyIdPool = nil
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

func (s *SimPlayer) SetBodyIdRange(start, end uint16) {
  s.rangeStart = start
  s.rangeEnd = end
  rnge := end - start
  s.bodyIdRange = int(rnge)
}

func (s *SimPlayer) GetBodyIdRangeStart() uint16 {
  return s.rangeStart
}
