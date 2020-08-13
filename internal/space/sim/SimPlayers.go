package sim

import (
  "sync"
  "log"

  "go-space-serv/internal/space/snet/udp"
  "go-space-serv/internal/space/player"
)

type SimPlayers struct {
  Count int
  playerMap sync.Map
}

func (p *SimPlayers) Add(id string, stats *player.PlayerStats) *udp.UDPPlayer {
  plr := udp.NewUdpPlayer(id)
  plr.SetStats(stats)

  _, exists := p.playerMap.LoadOrStore(id, plr)
  if !exists {
    p.Count += 1
  }

  return plr
}

func (p *SimPlayers) Remove(id string) {
  p.playerMap.Delete(id)
  log.Printf("%s left the simulation", id)
}

func (p *SimPlayers) GetPlayer(id string) *udp.UDPPlayer {
  plr, ok := p.playerMap.Load(id)
  if ok {
    return plr.(*udp.UDPPlayer)
  }

  return nil
}

func (p *SimPlayers) Push(playerId string, msg udp.UDPMsg) {
  plr := p.GetPlayer(playerId)
  if plr != nil && plr.GetState() >= udp.CONNECTED {
    plr.AddMsg(msg)
  } else {
    log.Printf("Tried to add msg to player %s who doesnt exist", playerId)
  }
}

func (p *SimPlayers) PushAll(msg udp.UDPMsg) {
  p.playerMap.Range(func(key, value interface{}) bool {
    plr := value.(*udp.UDPPlayer)
    if plr.GetState() >= udp.CONNECTED {
      plr.AddMsg(msg);
    }
    return true
  })
}

func (p *SimPlayers) PushExcluding(playerId string, msg udp.UDPMsg) {
  p.playerMap.Range(func(key, value interface{}) bool {
    plr := value.(*udp.UDPPlayer)
    if plr.GetState() >= udp.CONNECTED && plr.GetName() != playerId {
      plr.AddMsg(msg);
    }
    return true
  })
}

func (p *SimPlayers) PackAndSend() {
  p.playerMap.Range(func(key, value interface{}) bool {
    plr := value.(*udp.UDPPlayer)
    if plr.GetState() >= udp.CONNECTED {
      plr.PackAndSend()
    }
    return true
  })
}
