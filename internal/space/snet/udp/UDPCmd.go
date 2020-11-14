package udp

type UDPCmd byte

const (
  NONE UDPCmd = iota
  HELLO
  SHUTUP
  DISCONNECT
  CHALLENGE
  WELCOME
  SYNC
  ENTER
  EXIT
  MOVESHOOT
  DEBUGRECT
)

