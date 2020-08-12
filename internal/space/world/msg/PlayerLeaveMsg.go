package msg

import(
  "go-space-serv/internal/space/snet/tcp"
)

// Tell the client another client's player info.
type PlayerLeaveMsg struct {
  Id string
}

func (msg *PlayerLeaveMsg) GetCmd() tcp.TCPCmd { return tcp.LEAVE }
func (msg *PlayerLeaveMsg) Serialize(packet []byte, head int) int {
  packet[head] = byte(tcp.LEAVE)
  head++
  idBytes := []byte(msg.Id)
  idLen := byte(len(idBytes))
  packet[head] = idLen
  head++
  copy(packet[head:head+int(idLen)], idBytes)
  return head + int(idLen)
}
func (msg *PlayerLeaveMsg) Deserialize(packet []byte, head int) int { return 0 }


