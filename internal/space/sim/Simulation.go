package sim

import (
  "log"
  "sync"
  "time"

  "go-space-serv/internal/space/util"
  "go-space-serv/internal/space/world"
  "go-space-serv/internal/space/sim/msg"
  "go-space-serv/internal/space/snet/udp"
)

type Simulation struct {
  controlledBodies    sync.Map
  allBodies           []*udp.UDPBody
  worldMap            *world.WorldMap
  players             *SimPlayers

  // Timing
  seq                 uint16      // incremented each simulation frame, sync when rolls over
  lastSync            int64       // unix nanos the last time sync was performed
  lastFrame           int64       // unix nanos since the last simulation frame
  framesSinceLastSync int64       // simulation frames since last sync

  fromPlayers chan    udp.UDPMsg  // incoming msgs from clients (UdpPlayer)
}

func (s *Simulation) Start(worldMap *world.WorldMap, players *SimPlayers) {
  s.players = players
  s.fromPlayers = make(chan udp.UDPMsg, 100)
  s.worldMap = worldMap
  s.seq = 0
  s.lastSync = 0
  s.framesSinceLastSync = 0
  s.worldMap.SpawnX = 1600
  s.worldMap.SpawnY = 0
  go s.loop()
}

// Run the simulation loop
//   - process incoming messages fom players
//   - produce outgoing messages for players
//   - instruct UDPPlayer to pack and send messages
func (s *Simulation) processFrame(frameStart int64) {
  // Process incoming messages from players
  for j := 0; j < 50; j++ {
    tmp := s.pullFromPlayers()
    if tmp == nil {
      break
    }

    switch t := tmp.(type) {
      case *msg.CmdMsg:
        m := t
        playerId := m.GetPlayerId()
        cmd := m.GetCmd()
        switch cmd {
          case udp.SYNC:
            syncTime := s.lastFrame / (int64(time.Millisecond) / int64(time.Nanosecond))

            var response msg.SyncMsg
            response.Seq = s.seq
            response.Time = uint64(syncTime)

            s.players.Push(playerId, &response)
          case udp.ENTER:
            // TODO: verify rtt/ploss limit
            player := s.players.GetPlayer(playerId)
            if player != nil && player.GetState() != udp.SPECTATING {
              log.Printf("player %s spawn when not spectating.", playerId)
              break
            }

            spawnX := s.worldMap.SpawnX
            spawnY := s.worldMap.SpawnY
            x, y := s.worldMap.GetCellCenter(spawnX, spawnY)
            pBod := udp.NewControlledBody(player)
            pBod.SetPos(x, y)
            s.addControlledBody(playerId, pBod)

            log.Printf("Spawning %s at %d/%d -- %f/%f", playerId, spawnX, spawnY, x, y)

            var response msg.EnterMsg
            response.PlayerId = player.GetName()
            response.BodyId = pBod.GetId();
            response.X = uint32(spawnX)
            response.Y = uint32(spawnY)
            s.players.PushAll(&response)
        }
      case *msg.MoveShootMsg:
        m := t
        playerId := m.GetPlayerId()
        bod, ok := s.controlledBodies.Load(playerId)
        if ok && bod != nil {
          m.BodyId = bod.(*udp.UDPBody).GetId();
          s.players.PushExcluding(playerId, m)
        }
        break
      default:
    }
  }

  // Update all bodies
  // flag dead bodies for removal
  // process live bodies
  // replace ps.bodies with filtered list
  filteredBodies := s.allBodies[:0]
  for i, b := range s.allBodies {
    if !b.IsDead() {
      filteredBodies = append(filteredBodies, b)
    } else {
      s.allBodies[i] = nil
      continue
    }
  }
  s.allBodies = filteredBodies

  s.players.PackAndSend()
}

// Simulation loop
// Determines when to process frames.
// Processes input not related to controlled bodies
func (s *Simulation) loop() {
  simulationStart := time.Now().UnixNano()
  s.lastSync = simulationStart
  s.seq = 0
  shouldSync := false
  timestepNano := helpers.GetConfiguredTimestepNanos()

  for {
    frameStartTime := time.Now()
    frameStart := frameStartTime.UnixNano()
    framesToProcess := ((frameStart - s.lastSync) / timestepNano) - int64(s.seq)
    if framesToProcess > 0 {
      for i := int64(0); i < framesToProcess; i++ {
        s.seq++
        s.lastFrame = s.lastSync + (int64(s.seq) * timestepNano)
        s.processFrame(s.lastFrame)

        if s.seq == 0 {
          shouldSync = true
          s.seq = 1
        }
      }
    } else if shouldSync {
      shouldSync = false
    }

    time.Sleep(32)
  }
}

// Actions
//////////////

func (s *Simulation) pullFromPlayers() interface{} {
  var tmp interface{}
  select {
  case tmp = <-s.fromPlayers:
  default:
    return nil
  }

  return tmp
}

// Modify
///////////

func (s *Simulation) addControlledBody(id string, bod *udp.UDPBody) {
  s.allBodies = append(s.allBodies, bod)
  s.controlledBodies.Store(id, bod)
}

func (s *Simulation) RemoveControlledBody(id string) {
  bod, ok := s.controlledBodies.Load(id)
  if ok && bod != nil {
    bod.(*udp.UDPBody).Kill()
    s.controlledBodies.Delete(id)
  }
}

// Access
/////////////

func (s *Simulation) GetLastSync() int64 {
  return s.lastSync
}

func (s *Simulation) GetSeq() uint16 {
  return s.seq
}

func (s *Simulation) GetPlayerChan() chan udp.UDPMsg {
  return s.fromPlayers
}
