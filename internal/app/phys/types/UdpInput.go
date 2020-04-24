package phys

import(
  "fmt"

  "go-space-serv/internal/app/snet"
  . "go-space-serv/internal/app/snet/types"
)

type UdpInputType byte

const (
  HELLO UdpInputType = iota + 1

  // LIFECYCLE
  SPAWN
  SPECTATE

  // MOVEMENT
  MOVE

  // ABILITIES
  SHOOT_GUN
  SHOOT_BOMB
)

type UdpInput struct {
  seq uint16
  iType UdpInputType
  playerName string
  content []byte
}

func (i *UdpInput) GetType() UdpInputType {
  return i.iType
}

func (i *UdpInput) GetSeq() uint16 {
  return i.seq
}

func (i *UdpInput) GetName() string {
  return i.playerName
}

func (i *UdpInput) GetContent() []byte {
  return i.content
}

func (i *UdpInput) Deserialize (msg *NetworkMsg) {
  // name length: 2 bytes
  nameLengthBytes := msg.Data[:2]
  nameLength := snet.Read_uint16(nameLengthBytes)

  // name
  i.playerName = snet.Read_utf8(msg.Data[2:2+nameLength])

  // Sequence value: 2 byte
  i.seq = snet.Read_uint16(msg.Data[nameLength + 2:nameLength + 4])

  // Input type: 1 byte
  i.iType = UdpInputType(msg.Data[nameLength + 4])

  var contentLen = len(msg.Data) - 5 - int(nameLength)
  i.content = make([]byte, contentLen)
  copy(i.content[:], msg.Data[nameLength + 5:])
}

func (i *UdpInput) String() string {
  return fmt.Sprintf("player=%s seq=%d iType=%d content=%08b", i.playerName, i.seq, i.iType, i.content)
}
