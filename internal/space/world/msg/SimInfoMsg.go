package msg

import(
  "net"
  "encoding/binary"
  "go-space-serv/internal/space/snet/tcp"
)

type SimInfoMsg struct {
  Ip net.IP
  Port uint32
}

func (msg *SimInfoMsg) GetCmd() tcp.TCPCmd { return tcp.SIM_INFO }
func (msg *SimInfoMsg) Serialize(packet []byte, head int) int {
  packet[head] = byte(tcp.SIM_INFO)
  head++

  ipLen := len(msg.Ip)
  packet[head] = byte(ipLen)
  head++
  copy(packet[head:head+ipLen], msg.Ip)
  head += ipLen
  binary.LittleEndian.PutUint32(packet[head:head+4], msg.Port)
  head += 4
  return head
}
func (msg *SimInfoMsg) Deserialize(packet []byte, head int) int { return 0 }


