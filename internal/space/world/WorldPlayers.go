package world

import (
  "sync"
  "log"

  "github.com/google/uuid"

  "go-space-serv/internal/space/snet/tcp"
  "go-space-serv/internal/space/world/msg"
  "go-space-serv/internal/space/player"
)

type WorldPlayers struct {
  Count int
  playerMap sync.Map
}

func (p *WorldPlayers) Add(tcpPlr *tcp.TCPPlayer) *WorldPlayer{
  var plr WorldPlayer
  plr.Tcp = tcpPlr
  plr.Stats = player.DefaultPlayerStats()

  _, exists := p.playerMap.LoadOrStore(plr.Tcp.Id, &plr)
  if !exists {
    p.Count += 1

    var joinMsg msg.PlayerJoinMsg
    joinMsg.Id = plr.Tcp.Id
    joinMsg.Stats = plr.Stats

    p.playerMap.Range(func(key, value interface{}) bool {
      otherPlr := value.(*WorldPlayer)
      if otherPlr.Tcp.GetState() >= tcp.CONNECTED {
        if otherPlr.Tcp.Id != plr.Tcp.Id {
          otherPlr.Tcp.Outgoing <- &joinMsg;
        } else {
          var existMsg msg.PlayerJoinMsg
          existMsg.Id = otherPlr.Tcp.Id
          existMsg.Stats = otherPlr.Stats
          plr.Tcp.Outgoing <- &existMsg
        }
      }
      return true
    })
  }

  return &plr
}

func (p *WorldPlayers) Remove(id uuid.UUID) {
  p.playerMap.Delete(id)
  p.Count--

  var leaveMsg msg.PlayerLeaveMsg
  leaveMsg.Id = id
  p.PushAll(&leaveMsg)

  log.Printf("%s left the world", id)
}

func (p *WorldPlayers) GetPlayer(id uuid.UUID) *WorldPlayer {
  plr, ok := p.playerMap.Load(id)
  if ok {
    return plr.(*WorldPlayer)
  }

  return nil
}

func (p *WorldPlayers) Push(playerId uuid.UUID, msg tcp.TCPMsg) {
  plr := p.GetPlayer(playerId)
  if plr != nil && plr.Tcp.GetState() >= tcp.CONNECTED {
    plr.Tcp.Outgoing <- msg
  } else {
    log.Printf("Tried to push msg to player %s who doesnt exist", playerId)
  }
}

func (p *WorldPlayers) PushAll(msg tcp.TCPMsg) {
  p.playerMap.Range(func(key, value interface{}) bool {
    plr := value.(*WorldPlayer)
    if plr.Tcp.GetState() >= tcp.CONNECTED {
      plr.Tcp.Outgoing <- msg
    }
    return true
  })
}

func (p *WorldPlayers) PushAllExcluding(playerId uuid.UUID, msg tcp.TCPMsg) {
  p.playerMap.Range(func(key, value interface{}) bool {
    plr := value.(*WorldPlayer)
    if plr.Tcp.GetState() >= tcp.CONNECTED && plr.Tcp.Id != playerId {
      plr.Tcp.Outgoing <- msg
    }
    return true
  })
}
