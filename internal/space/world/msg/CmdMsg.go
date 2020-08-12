package msg

import(
  "go-space-serv/internal/space/snet/tcp"
)

type CmdMsg struct {
  cmd tcp.TCPCmd
}

func (msg *CmdMsg) Deserialize(packet []byte, head int) int {
  msg.cmd = tcp.TCPCmd(packet[head])
  head++
  return head
}
func (msg *CmdMsg) Serialize(packet []byte, head int) int {
  packet[head] = byte(msg.cmd)
  head++
  return head
}
func (msg *CmdMsg) GetCmd() tcp.TCPCmd { return msg.cmd }
func (msg *CmdMsg) SetCmd(c tcp.TCPCmd) { msg.cmd = c  }


