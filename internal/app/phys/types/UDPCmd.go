package phys

type UDPCmd byte

const (
  NONE UDPCmd = iota
  HELLO
  SHUTUP
  DISCONNECT
  CHALLENGE
  WELCOME
  SYNC
  SPAWN
)

