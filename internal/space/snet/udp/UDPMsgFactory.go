package udp

import(
  "github.com/google/uuid"
)
type UDPMsgFactory interface {
  CreateAndPublishMsg([]byte, int, chan UDPMsg, uuid.UUID) int
}
