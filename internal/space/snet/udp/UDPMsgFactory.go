package udp

type UDPMsgFactory interface {
  CreateAndPublishMsg([]byte, int, chan UDPMsg, string) int
}
