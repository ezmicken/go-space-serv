package msg

import (
  "encoding/binary"
  "go-space-serv/internal/space/snet/udp"
)

type SyncMsg struct {
  // sent
  Seq  uint16
  Time uint64
}

func (msg *SyncMsg) GetCmd() udp.UDPCmd { return udp.SYNC }
func (msg *SyncMsg) GetSize() int { return 11 }
func (msg *SyncMsg) Serialize(bytes []byte) {
  bytes[0] = byte(udp.SYNC)

  binary.LittleEndian.PutUint16(bytes[1:3], msg.Seq)
  binary.LittleEndian.PutUint64(bytes[3:], msg.Time)

  return
}

func (msg *SyncMsg) Deserialize(bytes []byte, head int) int { return head }
