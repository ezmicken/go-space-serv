package phys

import (
	"sync"
	"log"

  . "go-space-serv/internal/app/phys/interface"
	. "go-space-serv/internal/app/player/types"
)

type Players struct {
	Count int
	playerMap sync.Map
}

func (p *Players) Add(id string, stats *PlayerStats) {
	player := NewUdpPlayer(id)
	player.SetStats(stats)

  _, exists := p.playerMap.LoadOrStore(id, player)
  if !exists {
    p.Count += 1
  }
}

func (p *Players) Remove(id string) {
  p.playerMap.Delete(id)
  log.Printf("%s left the simulation", id)
}

func (p *Players) GetPlayer(id string) *UdpPlayer {
	player, ok := p.playerMap.Load(id)
	if ok {
		return player.(*UdpPlayer)
	}

	return nil
}

func (p *Players) AddMsg(msg UDPMsg, playerId string) {
  player := p.GetPlayer(playerId)
  if player != nil && player.GetState() >= CONNECTED {
    player.AddMsg(msg)
  } else {
    log.Printf("Tried to add msg to player %s who doesnt exist", playerId)
  }
}

func (p *Players) AddMsgAll(msg UDPMsg) {
  p.playerMap.Range(func(key, value interface{}) bool {
    player := value.(*UdpPlayer)
    if player.GetState() >= CONNECTED {
      player.AddMsg(msg);
    }
    return true
  })
}

func (p *Players) AddMsgExcluding(msg UDPMsg, playerId string) {
  p.playerMap.Range(func(key, value interface{}) bool {
    player := value.(*UdpPlayer)
    if player.GetState() >= CONNECTED && player.GetName() != playerId {
      player.AddMsg(msg);
    }
    return true
  })
}

func (p *Players) PackAndSend() {
  p.playerMap.Range(func(key, value interface{}) bool {
    player := value.(*UdpPlayer)
    if player.GetState() >= CONNECTED {
      player.PackAndSend()
    }
    return true
  })
}
