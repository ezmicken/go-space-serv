package msg

import (
  "encoding/binary"
  "github.com/google/uuid"
  "go-space-serv/internal/space/snet/udp"
)

type EnterMsg struct {
  PlayerId uuid.UUID // 9 - 11 characters
  BodyId uint16
  X uint32
  Y uint32
}

func (msg *EnterMsg) GetCmd() udp.UDPCmd { return udp.ENTER }

func (msg *EnterMsg) Serialize(bytes []byte) {
  offset := 0
  bytes[offset] = byte(udp.ENTER)
  offset++
  copy(bytes[offset:offset+16], msg.PlayerId[0:])
  offset += 16
  binary.LittleEndian.PutUint16(bytes[offset:offset+2], msg.BodyId)
  offset += 2
  binary.LittleEndian.PutUint32(bytes[offset:offset+4], msg.X)
  offset += 4
  binary.LittleEndian.PutUint32(bytes[offset:offset+4], msg.Y)

  return
}
func (msg *EnterMsg) GetSize() int {
  return 27
}

func (msg *EnterMsg) Deserialize(bytes []byte, head int) int { return head }
func (msg *EnterMsg) SetPlayerId(id uuid.UUID) {}
func (msg *EnterMsg) GetPlayerId() uuid.UUID { return uuid.UUID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0} }
