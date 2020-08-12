package msg

import(
  "encoding/binary"
  "go-space-serv/internal/space/snet/tcp"
)

type WorldInfoMsg struct {
  Size uint32
  Res byte
}

func (msg *WorldInfoMsg) GetCmd() tcp.TCPCmd { return tcp.WORLD_INFO }
func (msg *WorldInfoMsg) Serialize(packet []byte, head int) int {
  packet[head] = byte(tcp.WORLD_INFO)
  head++
  binary.LittleEndian.PutUint32(packet[head:head+4], msg.Size)
  head += 4
  packet[head] = msg.Res
  head++
  return head
}
func (msg *WorldInfoMsg) Deserialize(packet []byte, head int) int { return 0 }


