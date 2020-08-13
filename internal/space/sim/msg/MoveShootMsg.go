package msg

import(
  "encoding/binary"
  "go-space-serv/internal/space/snet"
  "go-space-serv/internal/space/snet/udp"
)

type MoveShootMsg struct {
  // local
  playerId string

  // outgoing
  BodyId    uint16

  // common
  Tick      uint16
  MoveShoot byte
}

func (msg *MoveShootMsg) GetCmd() udp.UDPCmd { return udp.MOVESHOOT }
func (msg *MoveShootMsg) Deserialize(packet []byte, head int) int {
  head++ // no need to read cmd.
  msg.Tick = snet.Read_uint16(packet[head:head+2])
  head += 2
  msg.MoveShoot = packet[head]

  return head + 1
}
func (msg *MoveShootMsg) GetSize() int {
  return 6
}

func (msg *MoveShootMsg) SetPlayerId(id string)   { msg.playerId = id }
func (msg *MoveShootMsg) GetPlayerId() string     { return msg.playerId }

func (msg *MoveShootMsg) Serialize(slice []byte) {
  slice[0] = byte(udp.MOVESHOOT)
  binary.LittleEndian.PutUint16(slice[1:3], msg.BodyId)
  binary.LittleEndian.PutUint16(slice[3:5], msg.Tick)
  slice[5] = msg.MoveShoot
}
