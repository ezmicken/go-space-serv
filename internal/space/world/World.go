package world

import(
  "log"
  "net"

  "github.com/google/uuid"

  "go-space-serv/internal/space/world/msg"
  "go-space-serv/internal/space/snet"
)

type World struct {
  worldMap *WorldMap
  players  *WorldPlayers
  bodyToPlayer map[uint16]uuid.UUID
}

func NewWorld(wp *WorldPlayers) *World {
  var wld World
  wld.players = wp

  var wm WorldMap
  wm.W = 4096
  wm.H = 4096
  wm.ChunkSize = 16
  wm.Resolution = 32
  wm.Seed = 209323094
  wm.Generate()
  log.Printf("worldMap generated.")
  wld.worldMap = &wm

  wld.bodyToPlayer = make(map[uint16]uuid.UUID)

  return &wld
}

func (w *World) PlayerJoin(plr *WorldPlayer, physIp net.IP, physPort uint32) {
  // Tell this client his stats
  var playerInfoMsg msg.PlayerInfoMsg
  playerInfoMsg.Id = plr.Tcp.Id
  playerInfoMsg.Stats = plr.Stats
  plr.Tcp.Outgoing <- &playerInfoMsg

  // Tell this client about the world
  var worldInfoMsg msg.WorldInfoMsg
  worldInfoMsg.Size = uint32(w.worldMap.W)
  worldInfoMsg.Res = byte(w.worldMap.Resolution)
  plr.Tcp.Outgoing <- &worldInfoMsg

  // Tell this client about the physics server
  var simInfoMsg msg.SimInfoMsg
  simInfoMsg.Ip = physIp
  simInfoMsg.Port = physPort
  plr.Tcp.Outgoing <- &simInfoMsg
}

func (w *World) PlayerLeave(id uuid.UUID) {
  w.players.Remove(id)

  var leaveMsg msg.PlayerLeaveMsg
  leaveMsg.Id = id

  w.players.PushAll(&leaveMsg)
}

func (w *World) InterpretPhysics(bytes []byte) {
  if bytes[0] == byte(snet.ISpawn) {
    bodyId := snet.Read_uint16(bytes[1:3])
    playerId, err := uuid.FromBytes(bytes[3:19])
    if err == nil {
      w.bodyToPlayer[bodyId] = playerId
      log.Printf("%v spawned", playerId)
    }
  } else if bytes[0] == byte(snet.ISpec) {
    bodyId := snet.Read_uint16(bytes[1:3])
    log.Printf("%v specced", w.bodyToPlayer[bodyId])
    delete(w.bodyToPlayer, bodyId)
  } else if bytes[0] == byte(snet.IState) {
    head := 1
    l := len(bytes)
    for head < l {
      bodyId, x, y := deserializeState(bytes[head:head+6])
      head += 6

      plr := w.players.GetPlayer(w.bodyToPlayer[bodyId])
      if plr != nil {
        // TODO: update position and do polygon boolean stuff
        plr.Update(x, y, w.worldMap.Poly)
      }
    }
  }
}

func (w *World) SerializeMap() []byte {
  return w.worldMap.Serialize()
}

func deserializeState(bytes []byte) (id, x, y uint16) {
  id = snet.Read_uint16(bytes[:2])
  x = snet.Read_uint16(bytes[2:4])
  y = snet.Read_uint16(bytes[4:6])
  return
}
