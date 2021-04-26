package msg

import(
  "math"
  "encoding/binary"
  "github.com/google/uuid"
  "go-space-serv/internal/space/player"
  "go-space-serv/internal/space/snet/tcp"
)

// Tell a client it's own player info.
type PlayerInfoMsg struct {
  Id uuid.UUID
  X uint16
  Y uint16
  Stats player.PlayerStats
}

func (msg *PlayerInfoMsg) GetCmd() tcp.TCPCmd { return tcp.PLAYER_INFO }
func (msg *PlayerInfoMsg) Serialize(packet []byte, head int) int {
  packet[head] = byte(tcp.PLAYER_INFO)
  head++
  copy(packet[head:head+16], msg.Id[0:])
  head += 16
  binary.LittleEndian.PutUint16(packet[head:head+2], msg.X)
  head += 2
  binary.LittleEndian.PutUint16(packet[head:head+2], msg.Y)
  head += 2
  binary.LittleEndian.PutUint32(packet[head:head+4], math.Float32bits(msg.Stats.Thrust))
  head += 4
  binary.LittleEndian.PutUint32(packet[head:head+4], math.Float32bits(msg.Stats.MaxSpeed))
  head += 4
  binary.LittleEndian.PutUint32(packet[head:head+4], math.Float32bits(msg.Stats.Rotation))
  head += 4
  binary.LittleEndian.PutUint32(packet[head:head+4], math.Float32bits(msg.Stats.BombSpeed))
  head += 4
  binary.LittleEndian.PutUint32(packet[head:head+4], math.Float32bits(msg.Stats.BombRadius))
  head += 4
  binary.LittleEndian.PutUint32(packet[head:head+4], math.Float32bits(msg.Stats.BombProximity))
  head += 4
  binary.LittleEndian.PutUint16(packet[head:head+4], uint16(msg.Stats.BombRate))
  head += 2
  binary.LittleEndian.PutUint16(packet[head:head+4], uint16(msg.Stats.BombLife))
  head += 2
  packet[head] = byte(msg.Stats.BombBounces)
  head++
  return head
}
func (msg *PlayerInfoMsg) Deserialize(packet []byte, head int) int { return 0 }


