package tcp

type TCPMsg interface {
  GetCmd() TCPCmd

  // Outgoing messages implement this.
  GetSize() int // Safe to call before Serialize, Unsafe to call before Deserialize.
  Serialize([]byte)

  // Incoming messages implement this.
  Deserialize(packet []byte, head int) int
}
