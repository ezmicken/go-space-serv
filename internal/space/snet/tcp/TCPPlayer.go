package tcp

import (
  "time"
  "log"
  "encoding/binary"
  "github.com/google/uuid"
  "github.com/panjf2000/gnet"
)

type TCPPlayerState byte

const (
  DISCONNECTED TCPPlayerState = iota
  AUTH
  CONNECTED
)

const PacketSize int = 1024

type TCPPlayer struct {
  Id              uuid.UUID
  Outgoing  chan  TCPMsg

  incoming  chan  TCPMsg
  connection      gnet.Conn
  factory         TCPMsgFactory
  state           TCPPlayerState
}

func NewPlayer(conn gnet.Conn, id uuid.UUID, factory TCPMsgFactory) *TCPPlayer {
  var p TCPPlayer
  p.Outgoing = make(chan TCPMsg, 100)
  p.incoming = make(chan TCPMsg, 100)
  p.connection = conn
  p.state = DISCONNECTED
  p.Id = id

  return &p
}

func (p *TCPPlayer) Connected() {
  p.state = CONNECTED
  go p.Tx()
}

func (p *TCPPlayer) Tx() {
  packet := make([]byte, PacketSize)
  head := 2
  ticker := time.NewTicker(time.Duration(100) * time.Millisecond)
  defer ticker.Stop()

  for {
    if p.state == DISCONNECTED {
      log.Printf("disconnected")
      break
    }
    var m TCPMsg
    select {
    case m = <- p.Outgoing:
      head = m.Serialize(packet, head)
    }

    if head > 2 {
      toSend := packet
      binary.LittleEndian.PutUint16(packet[0:2], uint16(head - 2))
      p.connection.AsyncWrite(toSend[:head])
      head = 2
      packet = make([]byte, PacketSize)
    }

    <- ticker.C
  }
}

func (p *TCPPlayer) Rx() {}

func (p *TCPPlayer) GetState() TCPPlayerState {
  return p.state
}

func (p *TCPPlayer) GetConnection() gnet.Conn {
  return p.connection
}
