package world

import(
  //"go-space-serv/internal/space/world/msg"
  "go-space-serv/internal/space/snet/tcp"
)

// TODO: pooling
type MsgFactory struct {}

// Create msg, deserialize it, publish it, return new head
func (mf *MsgFactory) CreateAndPublishMsg(packet []byte, head int, target chan tcp.TCPMsg, playerId string) int {
  //cmd := tcp.TCPCmd(packet[head])
  return head
}
