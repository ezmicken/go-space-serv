package world

import (
  "sync"
  "log"

  "go-space-serv/internal/space/player"
  "go-space-serv/internal/space/snet/tcp"
)

type WorldPlayers struct {
  Count int
  playerMap sync.Map
}

func (p *WorldPlayers) Add(id string, stats *player.PlayerStats) *tcp.TCPPlayer {
  // plr := tcp.NewTCPPlayer(id)
  // plr.SetStats(stats)

  // _, exists := p.playerMap.LoadOrStore(id, plr)
  // if !exists {
  //   p.Count += 1
  // }

  // return plr
  return nil
}

func (p *WorldPlayers) Remove(id string) {
  p.playerMap.Delete(id)
  log.Printf("%s left the simulation", id)
}

func (p *WorldPlayers) GetPlayer(id string) *tcp.TCPPlayer {
  plr, ok := p.playerMap.Load(id)
  if ok {
    return plr.(*tcp.TCPPlayer)
  }

  return nil
}

func (p *WorldPlayers) AddMsg(msg tcp.TCPMsg, playerId string) {
  plr := p.GetPlayer(playerId)
  if plr != nil && plr.GetState() >= tcp.CONNECTED {
    plr.AddMsg(msg)
  } else {
    log.Printf("Tried to add msg to player %s who doesnt exist", playerId)
  }
}

func (p *WorldPlayers) AddMsgAll(msg tcp.TCPMsg) {
  p.playerMap.Range(func(key, value interface{}) bool {
    plr := value.(*tcp.TCPPlayer)
    if plr.GetState() >= tcp.CONNECTED {
      plr.AddMsg(msg);
    }
    return true
  })
}

func (p *WorldPlayers) AddMsgExcluding(msg tcp.TCPMsg, playerId string) {
  p.playerMap.Range(func(key, value interface{}) bool {
    plr := value.(*tcp.TCPPlayer)
    if plr.GetState() >= tcp.CONNECTED && plr.GetPlayerId() != playerId {
      plr.AddMsg(msg);
    }
    return true
  })
}

func (p *WorldPlayers) PackAndSend() {
  p.playerMap.Range(func(key, value interface{}) bool {
    plr := value.(*tcp.TCPPlayer)
    if plr.GetState() >= tcp.CONNECTED {
      //plr.PackAndSend()
    }
    return true
  })
}
