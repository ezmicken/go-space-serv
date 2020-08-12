package msg

import(
  "encoding/binary"
  "go-space-serv/internal/space/snet/tcp"
)

// Tell a client it's own player info.
type BlocksMsg struct {
  Id uint16
  Data []byte
}

func (msg *BlocksMsg) GetCmd() tcp.TCPCmd { return tcp.BLOCKS }
func (msg *BlocksMsg) Serialize(packet []byte, head int) int {
  packet[head] = byte(tcp.BLOCKS)
  head++
  binary.LittleEndian.PutUint16(packet[head:head+2], msg.Id)
  head += 2
  dataLen := len(msg.Data)
  binary.LittleEndian.PutUint16(packet[head:head+2], uint16(dataLen))
  head += 2
  copy(packet[head:head+dataLen], msg.Data)
  head += dataLen
  return head
}
func (msg *BlocksMsg) Deserialize(packet []byte, head int) int { return 0 }


