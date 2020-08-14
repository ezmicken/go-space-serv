package msg

import (
  "encoding/binary"
  "go-space-serv/internal/space/snet/udp"
)

type EnterMsg struct {
  PlayerId string // 9 - 11 characters
  BodyId uint16
  X uint32
  Y uint32
}

func (msg *EnterMsg) GetCmd() udp.UDPCmd { return udp.ENTER }

func (msg *EnterMsg) Serialize(bytes []byte) {
  offset := 0
  bytes[offset] = byte(udp.ENTER)
  offset++
  stringBytes := []byte(msg.PlayerId)
  playerIdLen := len(stringBytes)
  bytes[offset] = byte(playerIdLen)
  offset++
  copy(bytes[offset:offset+playerIdLen], stringBytes)
  offset += playerIdLen
  binary.LittleEndian.PutUint16(bytes[offset:offset+2], msg.BodyId)
  offset += 2
  binary.LittleEndian.PutUint32(bytes[offset:offset+4], msg.X)
  offset += 4
  binary.LittleEndian.PutUint32(bytes[offset:offset+4], msg.Y)

  return
}
func (msg *EnterMsg) GetSize() int {
  return 12 + len(msg.PlayerId)
}

func (msg *EnterMsg) Deserialize(bytes []byte, head int) int { return head }