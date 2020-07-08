package phys

import(
  . "go-space-serv/internal/app/phys/types"
)

type CmdMsg struct {
  // local to server
  playerId string

  // sent
  cmd UDPCmd
}

func (msg CmdMsg) GetCmd() UDPCmd           { return msg.cmd }
func (msg CmdMsg) Deserialize(bytes []byte) {}
func (msg CmdMsg) Serialize([]byte)         {}
func (msg CmdMsg) GetSize() int             { return 1 }

func (msg *CmdMsg) SetCmd(val UDPCmd)       { msg.cmd = val }
func (msg *CmdMsg) SetPlayerId(id string)   { msg.playerId = id }
func (msg *CmdMsg) GetPlayerId() string     { return msg.playerId }
