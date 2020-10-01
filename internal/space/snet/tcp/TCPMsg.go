package tcp

type TCPMsg interface {
  GetCmd() TCPCmd
  Serialize(packet []byte, head int) int
  Deserialize(packet []byte, head int) int
}
