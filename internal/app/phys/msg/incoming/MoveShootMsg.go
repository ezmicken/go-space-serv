package phys

import(
  "go-space-serv/internal/app/snet"
  . "go-space-serv/internal/app/phys/types"
)

type MoveShootMsg struct {
  // local to server
  playerId string

  Tick      uint16
  MoveShoot byte
}

func (msg *MoveShootMsg) GetCmd() UDPCmd { return MOVESHOOT }
func (msg *MoveShootMsg) Deserialize(packet []byte, head int) int {
  head++ // no need to read cmd.
  msg.Tick = snet.Read_uint16(packet[head:head+2])
  head += 2
  msg.MoveShoot = packet[head]

  return head + 1
}
func (msg *MoveShootMsg) GetSize() int {
  return 3
}

func (msg *MoveShootMsg) SetPlayerId(id string)   { msg.playerId = id }
func (msg *MoveShootMsg) GetPlayerId() string     { return msg.playerId }

func (msg *MoveShootMsg) Serialize([]byte) {}
