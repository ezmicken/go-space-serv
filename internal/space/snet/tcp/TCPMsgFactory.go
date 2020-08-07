package tcp

type TCPMsgFactory interface {
  CreateAndPublishMsg([]byte, int, chan TCPMsg, string) int
}
