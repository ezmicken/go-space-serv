package msg

import(
  "math"
  "encoding/binary"
  "go-space-serv/internal/space/player"
  "go-space-serv/internal/space/snet/tcp"
)

// Tell a client it's own player info.
type PlayerInfoMsg struct {
  Id string
  Stats player.PlayerStats

}

func (msg *PlayerInfoMsg) GetCmd() tcp.TCPCmd { return tcp.PLAYER_INFO }
func (msg *PlayerInfoMsg) Serialize(packet []byte, head int) int {
  packet[head] = byte(tcp.PLAYER_INFO)
  head++
  idBytes := []byte(msg.Id)
  idLen := byte(len(idBytes))
  packet[head] = idLen
  head++
  copy(packet[head:head+int(idLen)], idBytes)
  head += int(idLen)
  binary.LittleEndian.PutUint32(packet[head:head+4], math.Float32bits(msg.Stats.Thrust))
  head += 4
  binary.LittleEndian.PutUint32(packet[head:head+4], math.Float32bits(msg.Stats.MaxSpeed))
  head += 4
  binary.LittleEndian.PutUint32(packet[head:head+4], math.Float32bits(msg.Stats.Rotation))
  head += 4
  return head
}
func (msg *PlayerInfoMsg) Deserialize(packet []byte, head int) int { return 0 }


