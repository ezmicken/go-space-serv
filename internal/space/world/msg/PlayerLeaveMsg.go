package msg

import(
  "github.com/google/uuid"
  "go-space-serv/internal/space/snet/tcp"
)

// Tell the client another client's player info.
type PlayerLeaveMsg struct {
  Id uuid.UUID
}

func (msg *PlayerLeaveMsg) GetCmd() tcp.TCPCmd { return tcp.LEAVE }
func (msg *PlayerLeaveMsg) Serialize(packet []byte, head int) int {
  packet[head] = byte(tcp.LEAVE)
  head++
  copy(packet[head:head+16], msg.Id[0:])
  head += 16
  return head
}
func (msg *PlayerLeaveMsg) Deserialize(packet []byte, head int) int { return 0 }


