package tcp

// one-byte indicator of client->server intent

type TCPCmd byte

const (
  PING TCPCmd = iota
  PONG
  BLOCKS
)
