package msg

import(
  "math"
  "encoding/binary"
  "go-space-serv/internal/space/snet/tcp"
)

type WorldInfoMsg struct {
  ChunksPerFile   uint32
  ChunkSize       uint32
  Size            uint32
  Seed            uint64
  Threshold       float64
}

func (msg *WorldInfoMsg) GetCmd() tcp.TCPCmd { return tcp.WORLD_INFO }
func (msg *WorldInfoMsg) Serialize(packet []byte, head int) int {
  packet[head] = byte(tcp.WORLD_INFO)
  head++
  binary.LittleEndian.PutUint32(packet[head:head+4], msg.ChunksPerFile)
  head += 4
  binary.LittleEndian.PutUint32(packet[head:head+4], msg.ChunkSize)
  head += 4
  binary.LittleEndian.PutUint32(packet[head:head+4], msg.Size)
  head += 4
  binary.LittleEndian.PutUint64(packet[head:head+8], msg.Seed)
  head += 8
  binary.LittleEndian.PutUint64(packet[head:head+8], math.Float64bits(msg.Threshold))
  head += 8
  return head
}
func (msg *WorldInfoMsg) Deserialize(packet []byte, head int) int { return 0 }
