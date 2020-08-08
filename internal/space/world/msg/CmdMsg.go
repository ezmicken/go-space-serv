package msg

import(
  "go-space-serv/internal/space/snet/tcp"
)

type CmdMsg struct {
  cmd tcp.TCPCmd
}

func (msg *CmdMsg) Deserialize(packet []byte, head int) int {
  msg.cmd = udp.UDPCmd(packet[head])
  return head++
}
func (msg *CmdMsg) Serialize(packet []byte, head) int {
  packet[head] = byte(msg.cmd)
  return head++
}
func (msg *CmdMsg) GetCmd() tcp.TCPCmd { return msg.cmd }
func (msg *CmdMsg) SetCmd(c tcp.TCPCmd) { msg.cmd = c  }


