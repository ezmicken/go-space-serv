package phys

import(
  "go-space-serv/internal/space/phys/msg"
  "go-space-serv/internal/space/snet/udp"
)

// TODO: pooling
type SimMsgFactory struct {}

// Create msg, deserialize it, publish it, return new head
func (mf *SimMsgFactory) CreateAndPublishMsg(packet []byte, head int, target chan udp.UDPMsg, playerId string) int {
  cmd := udp.UDPCmd(packet[head])
  if cmd == udp.SYNC {
    m := &msg.CmdMsg{}
    head = m.Deserialize(packet, head)
    m.SetPlayerId(playerId)
    target <- m
  } else if cmd == udp.ENTER {
    m := &msg.CmdMsg{}
    head = m.Deserialize(packet, head)
    m.SetPlayerId(playerId)
    target <- m
  } else if cmd == udp.EXIT {
    m := &msg.CmdMsg{}
    head = m.Deserialize(packet, head)
    m.SetPlayerId(playerId)
    target <- m
  } else if cmd == udp.MOVESHOOT {
    m := &msg.MoveShootMsg{}
    head = m.Deserialize(packet, head)
    m.SetPlayerId(playerId)
    target <- m
  }

  return head
}
