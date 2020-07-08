package phys

import (
  "encoding/binary"
  . "go-space-serv/internal/app/phys/types"
)

type SyncMsg struct {
  // sent
  Seq  uint16
  Time uint64
}

func (msg SyncMsg) GetCmd() UDPCmd { return SYNC }
func (msg SyncMsg) Deserialize(bytes []byte) {}
func (msg SyncMsg) Serialize(bytes []byte) {
  bytes[0] = byte(SYNC)

  binary.LittleEndian.PutUint16(bytes[1:3], msg.Seq)
  binary.LittleEndian.PutUint64(bytes[3:], msg.Time)

  return
}
func (msg SyncMsg) GetSize() int { return 11 }
