package tcp

import (
  "time"
  "log"
  "encoding/binary"
  "github.com/google/uuid"
  "github.com/panjf2000/gnet"

  "go-space-serv/internal/space/player"
)

type TCPPlayerState byte

const (
  DISCONNECTED TCPPlayerState = iota
  AUTH
  CONNECTED
)

const packetSize int = 1024

type TCPPlayer struct {
  connection      gnet.Conn
  factory         TCPMsgFactory
  state           TCPPlayerState
  toClient chan   TCPMsg
  toWorld  chan   TCPMsg

  info            *player.Player
}

func NewTCPPlayer(conn gnet.Conn, info *player.Player) *TCPPlayer {
  var p TCPPlayer
  p.toClient = make(chan TCPMsg, 100)
  p.toWorld = make(chan TCPMsg, 100)
  p.connection = conn
  p.state = DISCONNECTED
  p.info = info

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
    log.Printf("%v msg chan full. Discarding...", p.info.Id)
  }
}

func (p *TCPPlayer) SetWorldChan(c chan TCPMsg) {
  p.toWorld = c
}

func (p *TCPPlayer) SetMsgFactory(f TCPMsgFactory) {
  p.factory = f
}

func (p *TCPPlayer) GetPlayerId() uuid.UUID {
  return p.info.Id
}

func (p *TCPPlayer) GetState() TCPPlayerState {
  return p.state
}

func (p *TCPPlayer) GetConnection() gnet.Conn {
  return p.connection
}
