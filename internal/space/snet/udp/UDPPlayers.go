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
	player := NewUdpPlayer(id)
	player.SetStats(stats)

  _, exists := p.playerMap.LoadOrStore(id, player)
  if !exists {
    p.Count += 1
  }

  return player
}

func (p *UDPPlayers) Remove(id string) {
  p.playerMap.Delete(id)
  log.Printf("%s left the simulation", id)
}

func (p *UDPPlayers) GetPlayer(id string) *UDPPlayer {
	player, ok := p.playerMap.Load(id)
	if ok {
		return player.(*UDPPlayer)
	}

	return nil
}

func (p *UDPPlayers) AddMsg(msg UDPMsg, playerId string) {
  player := p.GetPlayer(playerId)
  if player != nil && player.GetState() >= CONNECTED {
    player.AddMsg(msg)
  } else {
    log.Printf("Tried to add msg to player %s who doesnt exist", playerId)
  }
}

func (p *UDPPlayers) AddMsgAll(msg UDPMsg) {
  p.playerMap.Range(func(key, value interface{}) bool {
    player := value.(*UDPPlayer)
    if player.GetState() >= CONNECTED {
      player.AddMsg(msg);
    }
    return true
  })
}

func (p *UDPPlayers) AddMsgExcluding(msg UDPMsg, playerId string) {
  p.playerMap.Range(func(key, value interface{}) bool {
    player := value.(*UDPPlayer)
    if player.GetState() >= CONNECTED && player.GetName() != playerId {
      player.AddMsg(msg);
    }
    return true
  })
}

func (p *UDPPlayers) PackAndSend() {
  p.playerMap.Range(func(key, value interface{}) bool {
    player := value.(*UDPPlayer)
    if player.GetState() >= CONNECTED {
      player.PackAndSend()
    }
    return true
  })
}
