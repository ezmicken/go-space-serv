package sim

import (
  "sync"
  "log"

  "github.com/google/uuid"

  "go-space-serv/internal/space/snet/udp"
  "go-space-serv/internal/space/player"
)

type SimPlayers struct {
  Count int
  playerMap sync.Map
}

func (p *SimPlayers) Add(udpPlayer *udp.UDPPlayer) {
  var plr SimPlayer
  plr.Stats = player.DefaultPlayerStats()
  plr.Udp = udpPlayer

  _, exists := p.playerMap.LoadOrStore(plr.Udp.Id, &plr)
  if !exists {
    p.Count += 1
    log.Printf("%v joined the simulation.", plr.Udp.Id)
  }
}

func (p *SimPlayers) Remove(id uuid.UUID) {
  p.playerMap.Delete(id)
  log.Printf("%s left the simulation", id)
}

func (p *SimPlayers) GetPlayer(id uuid.UUID) *SimPlayer {
  plr, ok := p.playerMap.Load(id)
  if ok {
    return plr.(*SimPlayer)
  }

  return nil
}

func (p *SimPlayers) Push(playerId uuid.UUID, msg udp.UDPMsg) {
  plr := p.GetPlayer(playerId)
  if plr != nil && plr.Udp.GetState() >= udp.CONNECTED {
    plr.Udp.Outgoing <- msg
  } else {
    log.Printf("Tried to add msg to player %s who doesnt exist", playerId)
  }
}

func (p *SimPlayers) PushAll(msg udp.UDPMsg) {
  p.playerMap.Range(func(key, value interface{}) bool {
    plr := value.(*SimPlayer)
    if plr.Udp.GetState() >= udp.CONNECTED {
      plr.Udp.Outgoing <- msg
    }
    return true
  })
}

func (p *SimPlayers) PushExcluding(playerId uuid.UUID, msg udp.UDPMsg) {
  p.playerMap.Range(func(key, value interface{}) bool {
    plr := value.(*SimPlayer)
    if plr.Udp.GetState() >= udp.CONNECTED && plr.Udp.Id != playerId {
      plr.Udp.Outgoing <- msg
    }
    return true
  })
}

func (p *SimPlayers) PackAndSend() {
  p.playerMap.Range(func(key, value interface{}) bool {
    plr := value.(*SimPlayer)
    if plr.Udp.GetState() >= udp.CONNECTED {
      plr.Udp.PackAndSend()
    }
    return true
  })
}
