package udp

import(
  "github.com/google/uuid"
)

type UDPMsg interface {
  GetCmd() UDPCmd

  // Outgoing messages implement this.
  GetSize() int // Safe to call before Serialize, Unsafe to call before Deserialize.
  Serialize([]byte)
  GetPlayerId() uuid.UUID
  SetPlayerId(uuid.UUID)

  // Incoming messages implement this.
  Deserialize(packet []byte, head int) int
}
