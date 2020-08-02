package msg

import(
  "go-space-serv/internal/space/snet/udp"
)

type CmdMsg struct {
  // local to server
  playerId string

  // sent
  cmd udp.UDPCmd
}

// UDPMsg interface
func (msg *CmdMsg) Deserialize(packet []byte, head int) int {
  msg.cmd = udp.UDPCmd(packet[head])
  return head+1
}

func (msg *CmdMsg) GetSize() int { return 1 }
func (msg *CmdMsg) GetCmd() udp.UDPCmd { return msg.cmd }

func (msg *CmdMsg) SetPlayerId(id string) { msg.playerId = id }
func (msg *CmdMsg) GetPlayerId() string { return msg.playerId }

func (msg *CmdMsg) Serialize([]byte) {}
