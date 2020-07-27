package phys

import(
  . "go-space-serv/internal/app/phys/msg/incoming"
  . "go-space-serv/internal/app/phys/interface"
  . "go-space-serv/internal/app/phys/types"
)

// TODO: pooling

// Create msg, deserialize it, publish it, return new head
func CreateAndPublishMsg(packet []byte, head int, target chan UDPMsg, playerId string) int {
  cmd := UDPCmd(packet[head])
  if cmd == SYNC {
    msg := &CmdMsg{}
    head = msg.Deserialize(packet, head)
    msg.SetPlayerId(playerId)
    target <- msg
  } else if cmd == ENTER {
    msg := &CmdMsg{}
    head = msg.Deserialize(packet, head)
    msg.SetPlayerId(playerId)
    target <- msg
  } else if cmd == EXIT {
    msg := &CmdMsg{}
    head = msg.Deserialize(packet, head)
    msg.SetPlayerId(playerId)
    target <- msg
  } else if cmd == MOVESHOOT {
    msg := &MoveShootMsg{}
    head = msg.Deserialize(packet, head)
    msg.SetPlayerId(playerId)
    target <- msg
  }

  return head
}
