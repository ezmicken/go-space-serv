package phys

import(
  "github.com/panjf2000/gnet"

  ."go-space-serv/internal/app/player/types"
)
type UdpPlayerState byte

const (
  CONNECTED UdpPlayerState = iota
  PLAYING
  SPECTATING
)

type UdpPlayer struct {
  name string
  active bool
  lastSync int64
  connection gnet.Conn
  state UdpPlayerState
  stats *PlayerStats
}

func NewUdpPlayer(n string) *UdpPlayer{
  var p UdpPlayer
  p.name = n
  p.active = false
  p.state = CONNECTED
  p.lastSync = 0
  return &p
}

func (p *UdpPlayer) GetName() string {
  return p.name
}

func (p *UdpPlayer) Activate() {
  p.active = true
}

func (p *UdpPlayer) Deactivate() {
  p.active = false
}

func (p *UdpPlayer) IsActive() bool {
  return p.active
}

func (p *UdpPlayer) SetState(s UdpPlayerState) {
  p.state = s
}

func (p *UdpPlayer) SetStats(s *PlayerStats) {
  p.stats = s
}

func (p *UdpPlayer) GetStats() *PlayerStats {
  return p.stats
}

func (p *UdpPlayer) GetState() UdpPlayerState {
  return p.state
}

func (p *UdpPlayer) SetConnection(c gnet.Conn) {
  p.connection = c
}

func (p *UdpPlayer) GetConnection() gnet.Conn {
  return p.connection
}

func (p *UdpPlayer) Sync(time int64) {
  p.lastSync = time
}

func (p *UdpPlayer) GetLastSync() int64 {
  return p.lastSync;
}

