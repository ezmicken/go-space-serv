package world

import (
  "sync"
  "log"

  "github.com/panjf2000/gnet"
  "github.com/google/uuid"

  "go-space-serv/internal/space/snet/tcp"
  "go-space-serv/internal/space/world/msg"
  "go-space-serv/internal/space/player"
)

type WorldPlayers struct {
  Count int
  playerMap sync.Map
}

func (p *WorldPlayers) Add(c gnet.Conn) (plr *tcp.TCPPlayer, id uuid.UUID) {
  plr = tcp.NewTCPPlayer(c)
  id = plr.GetPlayerId()

  _, exists := p.playerMap.LoadOrStore(id, plr)
  if !exists {
    p.Count += 1

    var stats player.PlayerStats
    stats.Thrust = 12
    stats.MaxSpeed = 20
    stats.Rotation = 210

    var joinMsg msg.PlayerJoinMsg
    joinMsg.Id = id
    joinMsg.Stats = stats

    p.playerMap.Range(func(key, value interface{}) bool {
      otherPlr := value.(*tcp.TCPPlayer)
      if otherPlr.GetState() >= tcp.CONNECTED {
        if otherPlr.GetPlayerId() != id {
          otherPlr.Push(&joinMsg);
        } else {
          var existMsg msg.PlayerJoinMsg
          existMsg.Id = otherPlr.GetPlayerId()
          existMsg.Stats = stats
          plr.Push(&existMsg)
        }
      }
      return true
    })
  }

  return
}

func (p *WorldPlayers) Remove(id uuid.UUID) {
  p.playerMap.Delete(id)
  p.Count--

  var leaveMsg msg.PlayerLeaveMsg
  leaveMsg.Id = id
  p.PushAll(&leaveMsg)

  log.Printf("%s left the world", id)
}

func (p *WorldPlayers) GetPlayer(id uuid.UUID) *tcp.TCPPlayer {
  plr, ok := p.playerMap.Load(id)
  if ok {
    return plr.(*tcp.TCPPlayer)
  }

  return nil
}

func (p *WorldPlayers) Push(playerId uuid.UUID, msg tcp.TCPMsg) {
  plr := p.GetPlayer(playerId)
  if plr != nil && plr.GetState() >= tcp.CONNECTED {
    plr.Push(msg)
  } else {
    log.Printf("Tried to push msg to player %s who doesnt exist", playerId)
  }
}

func (p *WorldPlayers) PushAll(msg tcp.TCPMsg) {
  p.playerMap.Range(func(key, value interface{}) bool {
    plr := value.(*tcp.TCPPlayer)
    if plr.GetState() >= tcp.CONNECTED {
      plr.Push(msg);
    }
    return true
  })
}

func (p *WorldPlayers) PushAllExcluding(playerId uuid.UUID, msg tcp.TCPMsg) {
  p.playerMap.Range(func(key, value interface{}) bool {
    plr := value.(*tcp.TCPPlayer)
    if plr.GetState() >= tcp.CONNECTED && plr.GetPlayerId() != playerId {
      plr.Push(msg);
    }
    return true
  })
}
