package sim

import (
  "log"
  "sync"
  "time"
  "encoding/binary"

  "github.com/go-gl/mathgl/mgl32"
  "github.com/google/uuid"

  "go-space-serv/internal/space/util"
  "go-space-serv/internal/space/world"
  "go-space-serv/internal/space/sim/msg"
  "go-space-serv/internal/space/snet/udp"
  "go-space-serv/internal/space/snet"
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
  toWorld     chan    []byte
  ticker              *time.Ticker
}

func (s *Simulation) Start(worldMap *world.WorldMap, players *SimPlayers, worldChan chan []byte) {
  s.players = players
  s.toWorld = worldChan
  s.fromPlayers = make(chan udp.UDPMsg, 100)
  s.worldMap = worldMap
  s.seq = 0
  s.lastSync = 0
  s.framesSinceLastSync = 0
  s.ticker = time.NewTicker(time.Duration(helpers.GetConfiguredTimestepNanos()) * time.Nanosecond)
  go s.loop()
}

// Run the simulation loop
//   - process incoming messages fom players
//   - produce outgoing messages for players
//   - instruct UDPPlayer to pack and send messages
func (s *Simulation) processFrame(frameStart int64, seq int) {
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
            if player != nil && player.Udp.GetState() != udp.SPECTATING {
              log.Printf("player %s spawn when not spectating.", playerId)
              break
            }

            x, y := s.worldMap.GetSpawnPoint()
            pBod := NewControlledBody(player)
            var ht HistoricalTransform
            ht.Seq = int(s.seq)
            ht.Angle = 0
            ht.AngleDelta = 0
            ht.Position = mgl32.Vec3{x, y, 0}
            ht.Velocity = mgl32.Vec3{0, 0, 0}
            ht.VelocityDelta = mgl32.Vec3{0, 0, 0}
            pBod.Initialize(ht)
            s.addControlledBody(playerId, pBod)
            player.Udp.SetState(udp.PLAYING)

            log.Printf("Spawning %s at %d/%d -- %f/%f", playerId, world.SPAWNX, world.SPAWNY, x, y)

            // tell other players
            var response msg.EnterMsg
            response.PlayerId = player.Udp.Id
            response.BodyId = pBod.GetBody().Id
            response.X = uint32(world.SPAWNX)
            response.Y = uint32(world.SPAWNY)
            s.players.PushAll(&response)

            // tell the map server
            worldSpawnMsg := []byte{byte(snet.ISpawn), 0, 0}
            binary.LittleEndian.PutUint16(worldSpawnMsg[1:3], pBod.GetBody().Id)

            worldSpawnMsg = append(worldSpawnMsg,  playerId[0:]...)
            s.toWorld <- worldSpawnMsg
          case udp.EXIT:
            player := s.players.GetPlayer(playerId)
            if player != nil && player.Udp.GetState() == udp.PLAYING {
              cb, ok := s.controlledBodies.Load(playerId)
              if ok && cb != nil {
                bodyId := cb.(*ControlledBody).GetBody().Id
                s.RemoveControlledBody(playerId)
                player.Udp.SetState(udp.SPECTATING)

                // tell other playesr
                var response msg.ExitMsg
                response.BodyId = bodyId
                s.players.PushAll(&response)

                // tell the map server
                worldSpecMsg := []byte{byte(snet.ISpec), 0, 0}
                binary.LittleEndian.PutUint16(worldSpecMsg[1:3], bodyId)
                s.toWorld <- worldSpecMsg
              }
            }
        }
      case *msg.MoveShootMsg:
        m := t
        playerId := m.GetPlayerId()

        // TODO: validate input
        b, ok := s.controlledBodies.Load(playerId)
        if ok && b != nil {
          cb := b.(*ControlledBody)
          m.BodyId = cb.GetBody().Id;
          s.players.PushExcluding(playerId, m)
          cb.InputToState(int(m.Tick), m.MoveShoot)
        }

        break
      default:
    }
  }

  notifyWorld := seq % helpers.GetConfiguredWorldRate() == 0
  x := float32(-1)
  y := float32(-1)


  worldMsg := []byte{}
  head := 1

  if notifyWorld == true {
    worldMsg = append(worldMsg, byte(snet.IState))
  }

  // Advance the simulation by one step for each controlled body
  s.controlledBodies.Range(func(key, value interface{}) bool {
    cb := value.(*ControlledBody)
    x, y = cb.ProcessFrame(frameStart, seq, s.worldMap)
    player := cb.GetControllingPlayer()
    if player != nil && player.Udp.GetState() == udp.PLAYING {
      var debugMsg msg.DebugRectMsg
      debugMsg.R = cb.Collider.Narrow
      player.Udp.Outgoing <- &debugMsg
    }
    if notifyWorld && x != -1 && y != -1 {
      gridX, gridY := s.worldMap.GetCellFromPosition(x, y)
      bod := cb.GetBody()
      worldMsg = append(worldMsg, []byte{0, 0, 0, 0, 0, 0}...)
      binary.LittleEndian.PutUint16(worldMsg[head:head+2], bod.Id)
      head += 2
      binary.LittleEndian.PutUint16(worldMsg[head:head+2], uint16(gridX))
      head += 2
      binary.LittleEndian.PutUint16(worldMsg[head:head+2], uint16(gridY))
      head+= 2
    }

    return true
  })

  if notifyWorld && len(worldMsg) > 1 {
    s.toWorld <- worldMsg
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
  defer s.ticker.Stop()
  simulationStart := time.Now().UnixNano()
  s.lastSync = simulationStart
  s.seq = 0
  shouldSync := false
  timestepNano := helpers.GetConfiguredTimestepNanos()
  var frameStartTime time.Time
  var frameStart int64
  var framesToProcess int64

  for {
    frameStartTime = <- s.ticker.C
    frameStart = frameStartTime.UnixNano()
    framesToProcess = ((frameStart - s.lastSync) / timestepNano) - int64(s.seq)
    if framesToProcess > 0 {
      for i := int64(0); i < framesToProcess; i++ {
        s.seq++
        s.lastFrame = s.lastSync + (int64(s.seq) * timestepNano)
        s.processFrame(s.lastFrame, int(s.seq))

        if s.seq == 0 {
          shouldSync = true
          s.seq = 1
        }
      }
    } else if shouldSync {
      shouldSync = false
    }

    framesToProcess = 0
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

func (s *Simulation) addControlledBody(id uuid.UUID, cb *ControlledBody) {
  s.allBodies = append(s.allBodies, cb.GetBody())
  s.controlledBodies.Store(id, cb)
}

func (s *Simulation) RemoveControlledBody(id uuid.UUID) {
  cb, ok := s.controlledBodies.Load(id)
  if ok && cb != nil {
    cb.(*ControlledBody).GetBody().Kill()
    s.controlledBodies.Delete(id)
  }
}

// Access
/////////////

func (s *Simulation) GetPlayerChan() chan udp.UDPMsg {
  return s.fromPlayers
}
