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

func NewWorld(wp *WorldPlayers, mapName string) (*World, error) {
  var wld World
  wld.players = wp
  wm, err := NewWorldMap(mapName)
  if err != nil {
    return nil, err
  }
  wld.worldMap = wm
  wld.bodyToPlayer = make(map[uint16]uuid.UUID)
  return &wld, nil
}

func (w *World) PlayerJoin(plr *WorldPlayer, physIp net.IP, physPort uint32) {
  // Tell this client about the world
  worldInfoMsg := w.worldMap.GetWorldInfoMsg()
  plr.Tcp.Outgoing <- &worldInfoMsg

  // Tell this client his stats
  var playerInfoMsg msg.PlayerInfoMsg
  playerInfoMsg.Id = plr.Tcp.Id
  playerInfoMsg.Stats = plr.Stats
  playerInfoMsg.X = plr.X
  playerInfoMsg.Y = plr.Y
  plr.Tcp.Outgoing <- &playerInfoMsg

  // Tell this client about the physics server
  var simInfoMsg msg.SimInfoMsg
  simInfoMsg.Ip = physIp
  simInfoMsg.Port = physPort
  plr.Tcp.Outgoing <- &simInfoMsg

  plr.Update(plr.X, plr.Y, w.worldMap)
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
        plr.Update(x, y, w.worldMap)
      }
    }
  }
}

func deserializeState(bytes []byte) (id, x, y uint16) {
  id = snet.Read_uint16(bytes[:2])
  x = snet.Read_uint16(bytes[2:4])
  y = snet.Read_uint16(bytes[4:6])
  return
}
