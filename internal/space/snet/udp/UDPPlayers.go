package udp

import (
	"sync"
	"log"

	"go-space-serv/internal/space/player"
)

type UDPPlayers struct {
	Count int
	playerMap sync.Map
}

func (p *UDPPlayers) Add(id string, stats *player.PlayerStats) *UDPPlayer {
	plr := NewUdpPlayer(id)
	plr.SetStats(stats)

  _, exists := p.playerMap.LoadOrStore(id, plr)
  if !exists {
    p.Count += 1
  }

  return plr
}

func (p *UDPPlayers) Remove(id string) {
  p.playerMap.Delete(id)
  log.Printf("%s left the simulation", id)
}

func (p *UDPPlayers) GetPlayer(id string) *UDPPlayer {
	plr, ok := p.playerMap.Load(id)
	if ok {
		return plr.(*UDPPlayer)
	}

	return nil
}

func (p *UDPPlayers) AddMsg(msg UDPMsg, playerId string) {
  plr := p.GetPlayer(playerId)
  if plr != nil && plr.GetState() >= CONNECTED {
    plr.AddMsg(msg)
  } else {
    log.Printf("Tried to add msg to player %s who doesnt exist", playerId)
  }
}

func (p *UDPPlayers) AddMsgAll(msg UDPMsg) {
  p.playerMap.Range(func(key, value interface{}) bool {
    plr := value.(*UDPPlayer)
    if plr.GetState() >= CONNECTED {
      plr.AddMsg(msg);
    }
    return true
  })
}

func (p *UDPPlayers) AddMsgExcluding(msg UDPMsg, playerId string) {
  p.playerMap.Range(func(key, value interface{}) bool {
    plr := value.(*UDPPlayer)
    if plr.GetState() >= CONNECTED && plr.GetName() != playerId {
      plr.AddMsg(msg);
    }
    return true
  })
}

func (p *UDPPlayers) PackAndSend() {
  p.playerMap.Range(func(key, value interface{}) bool {
    plr := value.(*UDPPlayer)
    if plr.GetState() >= CONNECTED {
      plr.PackAndSend()
    }
    return true
  })
}
