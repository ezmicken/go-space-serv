package udp

import(
  "github.com/google/uuid"
)
type UDPMsgFactory interface {
  CreateAndPublishMsg(uint16, []byte, int, chan UDPMsg, uuid.UUID) int
}
