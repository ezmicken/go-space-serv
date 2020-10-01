package msg

import(
  "math"
  "encoding/binary"
  "github.com/google/uuid"
  "go-space-serv/internal/space/player"
  "go-space-serv/internal/space/snet/tcp"
)

// Tell the client another client's player info.
type PlayerJoinMsg struct {
  Id uuid.UUID
  Stats player.PlayerStats
}

func (msg *PlayerJoinMsg) GetCmd() tcp.TCPCmd { return tcp.JOIN }
func (msg *PlayerJoinMsg) Serialize(packet []byte, head int) int {
  packet[head] = byte(tcp.JOIN)
  head++
  copy(packet[head:head+16], msg.Id[0:])
  head += 16
  binary.LittleEndian.PutUint32(packet[head:head+4], math.Float32bits(msg.Stats.Thrust))
  head += 4
  binary.LittleEndian.PutUint32(packet[head:head+4], math.Float32bits(msg.Stats.MaxSpeed))
  head += 4
  binary.LittleEndian.PutUint32(packet[head:head+4], math.Float32bits(msg.Stats.Rotation))
  head += 4
  return head
}
func (msg *PlayerJoinMsg) Deserialize(packet []byte, head int) int { return 0 }


