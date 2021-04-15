package msg

import (
  "encoding/binary"
  "github.com/google/uuid"
  "go-space-serv/internal/space/snet/udp"
)

type SyncMsg struct {
  // sent
  Seq       uint16
  Time      uint64
  State     []byte
  StateLen  int
}

func (msg *SyncMsg) GetCmd() udp.UDPCmd { return udp.SYNC }
func (msg *SyncMsg) GetSize() int { return 11 + msg.StateLen }
func (msg *SyncMsg) Serialize(bytes []byte) {
  bytes[0] = byte(udp.SYNC)

  binary.LittleEndian.PutUint16(bytes[1:3], msg.Seq)
  binary.LittleEndian.PutUint64(bytes[3:11], msg.Time)
  copy(bytes[11:], msg.State[:msg.StateLen])

  return
}

func (msg *SyncMsg) Deserialize(bytes []byte, head int) int { return head }
func (msg *SyncMsg) SetPlayerId(id uuid.UUID) {}
func (msg *SyncMsg) GetPlayerId() uuid.UUID { return uuid.UUID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}}
