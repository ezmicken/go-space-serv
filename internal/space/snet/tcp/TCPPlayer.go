package tcp

import (
  "time"
  "log"
  "encoding/binary"
  "github.com/teris-io/shortid"
  "github.com/panjf2000/gnet"
)

type TCPPlayerState byte

const (
  DISCONNECTED TCPPlayerState = iota
  AUTH
  CONNECTED
)

const packetSize int = 1024

type TCPPlayer struct {
  playerId        string
  connection      gnet.Conn
  factory         TCPMsgFactory
  state           TCPPlayerState
  toClient chan   TCPMsg
  toWorld  chan   TCPMsg
}

var sid *shortid.Shortid

func NewTCPPlayer(conn gnet.Conn) *TCPPlayer {
  var p TCPPlayer
  p.toClient = make(chan TCPMsg, 100)
  p.toWorld = make(chan TCPMsg, 100)
  p.connection = conn
  p.state = DISCONNECTED

  if sid == nil {
    tempsid, err := shortid.New(0, shortid.DefaultABC, 447)
    if err != nil {
      panic(err)
    }
    sid = tempsid
  }

  id, err2 := sid.Generate()
  if err2 != nil {
    panic(err2)
  }

  p.playerId = id
  log.Printf("generated playerId: %s", id)

  return &p
}

func (p *TCPPlayer) Connected() {
  p.state = CONNECTED
  go p.Tx()
}

func (p *TCPPlayer) Tx() {
  packet := make([]byte, packetSize)
  head := 2
  ticker := time.NewTicker(time.Duration(500) * time.Millisecond)
  defer ticker.Stop()

  for {
    if p.state == DISCONNECTED {
      log.Printf("disconnected")
      break
    }
    for {
      var m TCPMsg
      fin := false
      select {
      case m = <- p.toClient:
        head = m.Serialize(packet, head)
        if head >= packetSize {
          log.Printf("packet is max size")
          fin = true
        }
      default:
        fin = true
      }

      if fin {
        break
      }
    }

    if head > 2 {
      binary.LittleEndian.PutUint16(packet[0:2], uint16(head - 2))
      p.connection.AsyncWrite(packet[:head])
      head = 2
      packet = make([]byte, packetSize)
    }

    <- ticker.C
  }
}

func (p *TCPPlayer) Rx() {

}

func (p *TCPPlayer) Push(m TCPMsg) {
  select {
  case p.toClient <- m:
  default:
    log.Printf("%s msg chan full. Discarding...", p.playerId)
  }
}

func (p *TCPPlayer) SetWorldChan(c chan TCPMsg) {
  p.toWorld = c
}

func (p *TCPPlayer) SetMsgFactory(f TCPMsgFactory) {
  p.factory = f
}

func (p *TCPPlayer) GetPlayerId() string {
  return p.playerId
}

func (p *TCPPlayer) GetState() TCPPlayerState {
  return p.state
}

func (p *TCPPlayer) GetConnection() gnet.Conn {
  return p.connection
}
