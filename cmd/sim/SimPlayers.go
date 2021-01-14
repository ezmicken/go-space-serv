package main

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
  plr := NewSimPlayer(player.DefaultPlayerStats(), udpPlayer)

  _, exists := p.playerMap.LoadOrStore(plr.Udp.Id, plr)
  if !exists {
    p.Count += 1
    log.Printf("%v joined the simulation.", plr.Udp.Id)
  }
}

func (p *SimPlayers) Remove(id uuid.UUID) {
  plr, ok := p.playerMap.Load(id)
  if ok {
    player := plr.(*SimPlayer)
    player.OnDisconnect()
  }
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

func (p *SimPlayers) Cull() {
  var playersToCull []uuid.UUID = nil
  p.playerMap.Range(func(key, value interface{}) bool {
    plr := value.(*SimPlayer)
    if plr.Udp.IsTimedOut() || plr.ShouldCull() {
      if playersToCull == nil {
        playersToCull = []uuid.UUID{}
      }
      playersToCull = append(playersToCull, plr.Udp.Id)
    }
    return true
  })
  for i := 0; i < len(playersToCull); i++ {
    log.Printf("%s left the simulation", playersToCull[i])
    p.playerMap.Delete(playersToCull[i])
  }
}

func (p *SimPlayers) PackAndSend() {
  p.playerMap.Range(func(key, value interface{}) bool {
    plr := value.(*SimPlayer)
    if plr.Udp.GetState() >= udp.CONNECTED {
      plr.OnPackAndSend()
      plr.Udp.PackAndSend()
    }
    return true
  })
}
