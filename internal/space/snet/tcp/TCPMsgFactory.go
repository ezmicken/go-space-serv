package tcp

import(
  "github.com/google/uuid"
)

type TCPMsgFactory interface {
  CreateAndPublishMsg([]byte, int, chan TCPMsg, uuid.UUID) int
}
