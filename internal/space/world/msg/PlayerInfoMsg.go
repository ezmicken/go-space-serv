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

func (msg *PlayerJoinMsg) GetCmd() tcp.TCPCmd { return tcp.PLAYER_INFO }
func (msg *PlayerJoinMsg) Serialize(packet []byte, head int) int {
  packet[head] = byte(tcp.PLAYER_INFO)
  head++
  idBytes := []byte(msg.Id)
  idLen := byte(len(idBytes))
  packet[head] = idLen
  head++
  copy(packet[head:head+int(idLen)], stringBytes)
  head++ idLen
  binary.LittleEndian.PutUint32(packet[head:head+4], math.Float32bits(msg.Stats.Thrust))
  head += 4
  binary.LittleEndian.PutUint32(packet[head:head+4], math.Float32bits(msg.Stats.MaxSpeed))
  head += 4
  binary.LittleEndian.PutUint32(packet[head:head+4], math.Float32bits(msg.Stats.Rotation))
  head += 4
  packet[head] = msg.res
  return head++
}
func (msg *PlayerJoinMsg) Deserialize(packet []byte, head int) int {}


