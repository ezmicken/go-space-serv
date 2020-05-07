package phys

import(
  "github.com/panjf2000/gnet"
  "log"

  ."go-space-serv/internal/app/player/types"
  ."go-space-serv/internal/app/snet/types"
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
  output chan *NetworkMsg
  state UdpPlayerState
  stats *PlayerStats
}

func NewUdpPlayer(n string) *UdpPlayer{
  var p UdpPlayer
  p.name = n
  p.active = false
  p.state = CONNECTED
  p.lastSync = 0
  p.output = make(chan *NetworkMsg, 100)
  return &p
}

func (p *UdpPlayer) AddMsg(msg *NetworkMsg) {
	select {
		case p.output <- msg:
		default:
			log.Printf("%s msg queue full. Discarding...", p.name)
	}
}

func (p *UdpPlayer) GetMsg() *NetworkMsg {
	var msg *NetworkMsg
	select {
		case msg = <- p.output:
		default:
			msg = nil
	}

	return msg
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

