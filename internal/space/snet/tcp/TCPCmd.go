package tcp

// one-byte indicator of client->server intent

type TCPCmd byte

const (
  PING TCPCmd = iota + 1
  PONG
  WORLD_INFO
  PLAYER_INFO
  SIM_INFO
  JOIN
  LEAVE
  BLOCKS
)
