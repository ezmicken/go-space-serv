package world

import (
  "sync"
  "log"

  "github.com/panjf2000/gnet"

  "go-space-serv/internal/space/snet/tcp"
)

type WorldPlayers struct {
  Count int
  playerMap sync.Map
}

func (p *WorldPlayers) Add(c gnet.Conn) (plr *tcp.TCPPlayer, id string) {
  plr = tcp.NewTCPPlayer(c)
  id = plr.GetPlayerId()

  _, exists := p.playerMap.LoadOrStore(id, plr)
  if !exists {
    p.Count += 1
  }

  return
}

func (p *WorldPlayers) Remove(id string) {
  p.playerMap.Delete(id)
  log.Printf("%s left the world", id)
}

func (p *WorldPlayers) GetPlayer(id string) *tcp.TCPPlayer {
  plr, ok := p.playerMap.Load(id)
  if ok {
    return plr.(*tcp.TCPPlayer)
  }

  return nil
}

func (p *WorldPlayers) Push(playerId string, msg tcp.TCPMsg) {
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

func (p *WorldPlayers) PushAllExcluding(playerId string, msg tcp.TCPMsg) {
  p.playerMap.Range(func(key, value interface{}) bool {
    plr := value.(*tcp.TCPPlayer)
    if plr.GetState() >= tcp.CONNECTED && plr.GetPlayerId() != playerId {
      plr.Push(msg);
    }
    return true
  })
}
