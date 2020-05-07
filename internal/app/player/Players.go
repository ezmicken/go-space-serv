package player

import (
	"sync"
	"log"

	"go-space-serv/internal/app/snet"

	. "go-space-serv/internal/app/player/types"
	. "go-space-serv/internal/app/phys/types"
	. "go-space-serv/internal/app/snet/types"
)

type Players struct {
	Count int
	playerMap sync.Map
	bodyMap sync.Map
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

func (p *Players) SyncPlayers(time int64, msg *NetworkMsg) {
  p.playerMap.Range(func(key, value interface{}) bool {
    player := value.(*UdpPlayer)
    if player.IsActive() {
    	player.Sync(time)
    	player.AddMsg(msg)
    	log.Printf("Synced %s", player.GetName())
    }
    return true
  })
}

func (p *Players) QueueMsgAll(msg *NetworkMsg) {
	if msg != nil {
	  p.playerMap.Range(func(key, value interface{}) bool {
	    player := value.(*UdpPlayer)
	    if player.IsActive() {
	    	player.AddMsg(msg)
	    }
	    return true
	  })
	}
}

func (p *Players) SendQueuedMsgs() {
  p.playerMap.Range(func(key, value interface{}) bool {
    player := value.(*UdpPlayer)
    if player.IsActive() {
    	msg := player.GetMsg()
      if msg != nil {
        player.GetConnection().SendTo(snet.GetDataFromNetworkMsg(msg))
      }
    }
    return true
  })
}
