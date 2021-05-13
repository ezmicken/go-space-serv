package main

import(
  "log"
  "github.com/google/uuid"
  "go-space-serv/internal/space/sim/msg"
  "go-space-serv/internal/space/snet/udp"
  "go-space-serv/internal/space/util"
)

// TODO: pooling
type SimMsgFactory struct {
  seq uint16
}

// Create msg, deserialize it, publish it, return new head
func (mf *SimMsgFactory) CreateAndPublishMsg(seq uint16, packet []byte, head int, target chan udp.UDPMsg, playerId uuid.UUID) int {
  cmd := udp.UDPCmd(packet[head])
  var m udp.UDPMsg = nil
  if cmd == udp.SYNC {
    m = &msg.CmdMsg{}
    head = m.Deserialize(packet, head)
  } else if cmd == udp.ENTER {
    m = &msg.CmdMsg{}
    head = m.Deserialize(packet, head)
  } else if cmd == udp.EXIT {
    m = &msg.CmdMsg{}
    head = m.Deserialize(packet, head)
  } else if cmd == udp.MOVESHOOT {
    m = &msg.MoveShootMsg{}
    head = m.Deserialize(packet, head)
  } else if cmd == udp.NONE {
    m = &msg.CmdMsg{}
    head = m.Deserialize(packet, head)
  } else {
    log.Printf("Unknown command %v %v - %v -- %v", cmd, seq, mf.seq, helpers.BitString(packet))
    head++
  }

  if m != nil {
    if seq == mf.seq + 1 {
      m.SetPlayerId(playerId)
      target <- m
      mf.seq++
    }
  }

  return head
}
