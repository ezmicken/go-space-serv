package main

import (
  "log"
  "sync"
  "time"
  "encoding/binary"

  "github.com/ezmicken/fixpoint"
  "github.com/ezmicken/spacesim"
  "github.com/ezmicken/uuint16"
  "github.com/google/uuid"

  "go-space-serv/internal/space/util"
  "go-space-serv/internal/space/world"
  "go-space-serv/internal/space/sim/msg"
  "go-space-serv/internal/space/snet/udp"
  "go-space-serv/internal/space/snet"
)

type Sim struct {
  space               *spacesim.Simulation
  bodyIdsByPlayer     sync.Map
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

func (s *Sim) Start(worldMap *world.WorldMap, players *SimPlayers, worldChan chan []byte) {
  s.players = players
  s.toWorld = worldChan
  s.fromPlayers = make(chan udp.UDPMsg, 100)
  s.worldMap = worldMap
  s.seq = 0
  s.lastSync = 0
  s.framesSinceLastSync = 0
  s.ticker = time.NewTicker(time.Duration(helpers.GetConfiguredTimestepNanos()) * time.Nanosecond)
  ts := fixpoint.Q16FromInt32(int32(helpers.GetConfiguredTimestep()))
  scale := fixpoint.Q16FromInt32(32)
  s.space = spacesim.NewSimulation(ts, scale)
  go s.loop()
}

// Sim loop
// Determines when to process frames.
// Processes input not related to controlled bodies
func (s *Sim) loop() {
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

        s.space.Advance(int(s.seq))
      }
    } else if shouldSync {
      shouldSync = false
    }

    framesToProcess = 0
  }
}

// Run the simulation loop
//   - process incoming messages fom players
//   - produce outgoing messages for players
//   - instruct UDPPlayer to pack and send messages
func (s *Sim) processFrame(frameStart int64, seq int) {
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
            pBodId, err := uuint16.Rent()
            if err != nil { panic(err.Error()) }
            s.space.AddControlledBody(pBodId, int32(world.SPAWNX), int32(world.SPAWNY))
            s.bodyIdsByPlayer.Store(playerId, pBodId)
            player.Udp.SetState(udp.PLAYING)

            log.Printf("Spawning %s at %d/%d -- %f/%f", playerId, world.SPAWNX, world.SPAWNY, x, y)

            // tell other players
            var response msg.EnterMsg
            response.PlayerId = player.Udp.Id
            response.BodyId = pBodId
            response.X = uint32(world.SPAWNX)
            response.Y = uint32(world.SPAWNY)
            s.players.PushAll(&response)

            // tell the map server
            worldSpawnMsg := []byte{byte(snet.ISpawn), 0, 0}
            binary.LittleEndian.PutUint16(worldSpawnMsg[1:3], pBodId)

            worldSpawnMsg = append(worldSpawnMsg,  playerId[0:]...)
            s.toWorld <- worldSpawnMsg
          case udp.EXIT:
            player := s.players.GetPlayer(playerId)
            if player != nil && player.Udp.GetState() == udp.PLAYING {
              pBodId, ok := s.bodyIdsByPlayer.Load(playerId)
              if ok {
                bodyId := pBodId.(uint16)
                s.space.RemoveControlledBody(bodyId)
                s.bodyIdsByPlayer.Delete(playerId)
                player.Udp.SetState(udp.SPECTATING)

                // tell other playesr
                var response msg.ExitMsg
                response.BodyId = 0//bodyId
                s.players.PushAll(&response)

                // tell the map server
                worldSpecMsg := []byte{byte(snet.ISpec), 0, 0}
                binary.LittleEndian.PutUint16(worldSpecMsg[1:3], bodyId)
                s.toWorld <- worldSpecMsg
              }
            }
        }
      case *msg.MoveShootMsg: // TODO: validate input
        m := t
        playerId := m.GetPlayerId()

        id, ok := s.bodyIdsByPlayer.Load(playerId)
        if ok {
          cb := s.space.GetControlledBody(id.(uint16))
          if cb != nil {
            cb.InputToState((m.Tick + 12), m.MoveShoot)
            s.players.PushExcluding(playerId, m)
          }
        }

        break
      default:
    }
  }

  notifyWorld := seq % helpers.GetConfiguredWorldRate() == 0
  //x := float32(-1)
  //y := float32(-1)


  worldMsg := []byte{}
  head := 1

  if notifyWorld == true {
    worldMsg = append(worldMsg, byte(snet.IState))
  }

  // Advance the simulation one step for each controlled body
  s.bodyIdsByPlayer.Range(func(key, value interface{}) bool {
    id := value.(uint16)
    playerId := key.(uuid.UUID)
    player := s.players.GetPlayer(playerId)
    cb := s.space.GetControlledBody(id)
    if cb != nil {
      cb.Advance(uint16(seq))
      nextPos := cb.GetBody().NextPos
      x := nextPos.X.Float()
      y := nextPos.Y.Float()
      if player != nil && player.Udp.GetState() == udp.PLAYING {
        // send debug rect
      }
      if notifyWorld {
        xCoord, yCoord := s.worldMap.GetCellFromPosition(x, y)
        worldMsg = append(worldMsg, []byte{0, 0, 0, 0, 0, 0}...)
        binary.LittleEndian.PutUint16(worldMsg[head:head+2], id)
        head += 2
        binary.LittleEndian.PutUint16(worldMsg[head:head+2], uint16(xCoord))
        head += 2
        binary.LittleEndian.PutUint16(worldMsg[head:head+2], uint16(yCoord))
        head+= 2
      }
    }

    return true
  })

  if notifyWorld && len(worldMsg) > 1 {
    s.toWorld <- worldMsg
  }

  s.players.PackAndSend()
}

// Actions
//////////////

func (s *Sim) pullFromPlayers() interface{} {
  var tmp interface{}
  select {
  case tmp = <-s.fromPlayers:
  default:
    return nil
  }

  return tmp
}

// Access
/////////////

func (s *Sim) GetPlayerChan() chan udp.UDPMsg {
  return s.fromPlayers
}
