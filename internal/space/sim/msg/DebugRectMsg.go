package msg

import(
  "math"
  "encoding/binary"
  "go-space-serv/internal/space/snet/udp"
  "go-space-serv/internal/space/geom"
)

type DebugRectMsg struct {
  R geom.Rect
}

func (msg *DebugRectMsg) GetCmd() udp.UDPCmd { return udp.DEBUGRECT }

func (msg *DebugRectMsg) Serialize(bytes []byte) {
  offset := 0
  bytes[offset] = byte(udp.DEBUGRECT)
  offset++

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
  return 17
}

func (msg *DebugRectMsg) Deserialize(bytes []byte, head int) int { return head }
