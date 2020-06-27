package phys

import(
  "time"
  "log"

  "github.com/panjf2000/gnet"

  ."go-space-serv/internal/app/player/types"
  ."go-space-serv/internal/app/snet/types"
)
type UdpPlayerState byte

const (
  DISCONNECTED UdpPlayerState = iota
  CHALLENGED
  CONNECTED
  PLAYING
  SPECTATING
)

type UdpPlayer struct {
  spamChan chan struct{}
  name string
  active bool
  lastSync int64
  clientSalt int64
  serverSalt int64
  connection gnet.Conn
  output chan *NetworkMsg
  state UdpPlayerState
  stats *PlayerStats
}

func NewUdpPlayer(n string) *UdpPlayer{
  var p UdpPlayer
  p.name = n
  p.active = false
  p.state = DISCONNECTED
  p.lastSync = 0
  p.output = make(chan *NetworkMsg, 100)
  return &p
}

// Sends a packet every rate milliseconds count times
func (p *UdpPlayer) SendRepeating(msg []byte, rate, count int) {
  if p.spamChan != nil {
    close(p.spamChan)
  }

  log.Printf("sending first %d", msg[4])
  p.connection.SendTo(msg)

  // Use a local reference to the chan
  // for the case where it was closed
  // and re-opened during Sleep
  thisChan := make(chan struct{})
  p.spamChan = thisChan

  go func() {
    for i := 0; i < count; i++ {
      _, open := <-thisChan
      if (!open) {
        return
      }

      log.Printf("sending %d (iteration %d)", msg[4], i)
      p.connection.SendTo(msg)
      time.Sleep(time.Duration(rate) * time.Millisecond)
    }

    log.Printf("Player.SendRepeating finished all iterations.")
    close(thisChan)
  }()
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

func (p *UdpPlayer) SetClientSalt(salt int64) {
  p.clientSalt = salt
}
func (p *UdpPlayer) GetClientSalt() int64 {
  return p.clientSalt
}
func (p *UdpPlayer) SetServerSalt(salt int64) {
  p.serverSalt = salt
}
func (p *UdpPlayer) GetServerSalt() int64 {
  return p.serverSalt
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

