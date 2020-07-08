package phys

import(
  . "go-space-serv/internal/app/phys/types"
)

type SyncRequestMsg struct {
  // local to server
  playerId string

  // sent
  cmd UDPCmd
}

func (msg SyncRequestMsg) GetCmd() UDPCmd           { return SYNC }
func (msg SyncRequestMsg) Deserialize(bytes []byte) {}
func (msg SyncRequestMsg) Serialize([]byte)         {}
func (msg SyncRequestMsg) GetSize() int             { return 1 }

func (msg *SyncRequestMsg) SetPlayerId(id string)   { msg.playerId = id }
func (msg *SyncRequestMsg) GetPlayerId() string     { return msg.playerId }
