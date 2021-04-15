package msg

import(
  "math"
  "encoding/binary"
  "github.com/google/uuid"
  "go-space-serv/internal/space/snet/udp"
  "go-space-serv/internal/space/geom"
)

type DebugRectMsg struct {
  R geom.Rect
  Seq uint16
}

func (msg *DebugRectMsg) GetCmd() udp.UDPCmd { return udp.DEBUGRECT }

func (msg *DebugRectMsg) Serialize(bytes []byte) {
  offset := 0
  bytes[offset] = byte(udp.DEBUGRECT)
  offset++

  binary.LittleEndian.PutUint16(bytes[offset:offset+2], msg.Seq)
  offset += 2
  binary.LittleEndian.PutUint32(bytes[offset:offset+4], math.Float32bits(msg.R.X))
  offset += 4
  binary.LittleEndian.PutUint32(bytes[offset:offset+4], math.Float32bits(msg.R.Y))
  offset += 4
  binary.LittleEndian.PutUint32(bytes[offset:offset+4], math.Float32bits(msg.R.W))
  offset += 4
  binary.LittleEndian.PutUint32(bytes[offset:offset+4], math.Float32bits(msg.R.H))
  offset += 4

  return
}

func (msg *DebugRectMsg) GetSize() int {
  return 19
}

func (msg *DebugRectMsg) Deserialize(bytes []byte, head int) int { return head }
func (msg *DebugRectMsg) SetPlayerId(id uuid.UUID) {}
func (msg *DebugRectMsg) GetPlayerId() uuid.UUID { return uuid.UUID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0} }
