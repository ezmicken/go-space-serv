package msg

import (
  "encoding/binary"
  "github.com/google/uuid"
  "go-space-serv/internal/space/snet/udp"
)

type ExitMsg struct {
  BodyId uint16
}

func (msg *ExitMsg) GetCmd() udp.UDPCmd { return udp.EXIT }

func (msg *ExitMsg) Serialize(bytes []byte) {
  offset := 0
  bytes[offset] = byte(udp.EXIT)
  offset++
  binary.LittleEndian.PutUint16(bytes[offset:offset+2], msg.BodyId)
  offset += 2

  return
}
func (msg *ExitMsg) GetSize() int {
  return 26
}

func (msg *ExitMsg) Deserialize(bytes []byte, head int) int { return head }
func (msg *ExitMsg) SetPlayerId(id uuid.UUID) {}
func (msg *ExitMsg) GetPlayerId() uuid.UUID { return uuid.UUID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0} }
