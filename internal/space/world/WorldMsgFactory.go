package world

import(
  "github.com/google/uuid"
  //"go-space-serv/internal/space/world/msg"
  "go-space-serv/internal/space/snet/tcp"
)

// TODO: pooling
type WorldMsgFactory struct {}

// Create msg, deserialize it, publish it, return new head
func (mf *WorldMsgFactory) CreateAndPublishMsg(packet []byte, head int, target chan tcp.TCPMsg, playerId uuid.UUID) int {
  //cmd := tcp.TCPCmd(packet[head])
  return head
}
