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

  serializedState     [1024]byte
  serializedStateSize int
}

var bodyIdsPerPlayer int = 512

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
  s.serializedState = [1024]byte{}
  s.serializedStateSize = 0
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
        s.space.Advance(int(s.seq))
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

// Run the simulation loop
//   - process incoming messages fom players
//   - produce outgoing messages for players
//   - instruct UDPPlayer to pack and send messages
func (s *Sim) processFrame(frameStart int64, seq int) {
  s.serializedStateSize = s.space.SerializeState(s.serializedState[0:1024], 0)
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
        player := s.players.GetPlayer(playerId)
        switch cmd {
          case udp.SYNC:
            syncTime := s.lastFrame / (int64(time.Millisecond) / int64(time.Nanosecond))

            var response msg.SyncMsg
            response.Seq = s.seq
            response.Time = uint64(syncTime)
            response.State = s.serializedState[:s.serializedStateSize]
            response.StateLen = s.serializedStateSize

            s.players.Push(playerId, &response)
          case udp.ENTER:
            // TODO: verify rtt/ploss limit
            if player != nil && !player.IsSpectating() {
              log.Printf("player %s spawn when not spectating.", playerId)
              break
            }

            s.spawnPlayer(player)
          case udp.EXIT:
            s.specPlayer(player)
        }
      case *msg.MoveShootMsg: // TODO: validate input
        m := t
        playerId := m.GetPlayerId()

        id, ok := s.bodyIdsByPlayer.Load(playerId)
        if ok {
          bodyId := id.(uint16)
          cb := s.space.GetControlledBody(bodyId)
          if cb != nil {
            //log.Printf("%v: input received for %v (%v)", seq, m.Tick, bodyId)
            m.Tick += 12
            cb.PushInput(m.Tick, m.MoveShoot)
            m.BodyId = bodyId
            s.players.PushExcluding(playerId, m)
            if helpers.ShouldEchoInput() {
              s.players.Push(playerId, m)
            }
          }
        }

        break
      default:
        log.Printf("Unknown message.")
    }
  }

  notifyWorld := seq % helpers.GetConfiguredWorldRate() == 0

  worldMsg := []byte{}
  head := 1

  if notifyWorld == true {
    worldMsg = append(worldMsg, byte(snet.IState))
  }

  // temporary storage for players that should be removed.
  var playersToSpec []*SimPlayer = nil

  // Advance the simulation one step for each controlled body
  s.bodyIdsByPlayer.Range(func(key, value interface{}) bool {
    id := value.(uint16)
    playerId := key.(uuid.UUID)
    player := s.players.GetPlayer(playerId)

    // spec users who do not respond fast enough.
    if player.Udp != nil {
      if player.Udp.IsTimedOut() || player.ShouldCull() {
        if playersToSpec == nil {
          playersToSpec = []*SimPlayer{ player }
        } else {
          playersToSpec = append(playersToSpec, player)
        }
      }
    }

    cb := s.space.GetControlledBody(id)
    if cb != nil {
      s.worldMap.PushBlockRects(seq, cb)
      cb.Advance(uint16(seq))
      nextPos := cb.GetBody().NextPos
      x := nextPos.X.Float()
      y := nextPos.Y.Float()
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

  if playersToSpec != nil {
    log.Printf("removing players");
    for i := 0; i < len(playersToSpec); i++ {
      s.specPlayer(playersToSpec[i])
    }
  }

  s.players.Cull()

  if notifyWorld && len(worldMsg) > 1 {
    s.toWorld <- worldMsg
  }

  s.players.PackAndSend()
}

// Actions
//////////////
func (s *Sim) spawnPlayer(player *SimPlayer) {
  playerId := player.Udp.Id
  x, y := s.worldMap.GetSpawnPoint()


  // Reserve some body ids for player projectiles.
  // TODO: handle these errors. @ 512 it should support 100 ish players.
  pBodId, err := uuint16.Rent()
  if err != nil { panic(err.Error()) }
  rangeStart, err2 := uuint16.Rent()
  if err2 != nil { panic(err2.Error()) }
  for i := 0; i < bodyIdsPerPlayer-1; i++ {
    _, err3 := uuint16.Rent()
    if err3 != nil { panic(err3.Error()) }
  }
  rangeEnd := rangeStart + uint16(bodyIdsPerPlayer)
  player.SetBodyIdRange(rangeStart, rangeEnd)

  bi := spacesim.BodyInfo {
    pBodId,
    48,
    0,
    -1,
    0.5,
    0.0,
    0.0,
  }
  s.space.AddControlledBody(pBodId, int32(world.SPAWNX), int32(world.SPAWNY), bi)
  s.bodyIdsByPlayer.Store(playerId, pBodId)
  player.OnEnter()

  log.Printf("Spawning %s at %d/%d -- %f/%f", playerId, world.SPAWNX, world.SPAWNY, x, y)

  // tell other players
  var response msg.EnterMsg
  response.PlayerId = player.Udp.Id
  response.BodyId = pBodId
  response.RangeStart = rangeStart
  response.RangeEnd = rangeEnd
  response.X = uint32(world.SPAWNX)
  response.Y = uint32(world.SPAWNY)
  s.players.PushAll(&response)

  // tell the map server
  worldSpawnMsg := []byte{byte(snet.ISpawn), 0, 0}
  binary.LittleEndian.PutUint16(worldSpawnMsg[1:3], pBodId)

  worldSpawnMsg = append(worldSpawnMsg,  playerId[0:]...)
  s.toWorld <- worldSpawnMsg
}

func (s *Sim) RemovePlayer(id uuid.UUID) {
  s.players.Remove(id)
}

func (s *Sim) specPlayer(player *SimPlayer) {
  if player != nil {
    playerId := player.Udp.Id
    pBodId, ok := s.bodyIdsByPlayer.Load(playerId)
    if ok {
      bodyId := pBodId.(uint16)
      s.space.RemoveControlledBody(bodyId)
      s.bodyIdsByPlayer.Delete(playerId)
      player.OnExit()

      rangeStart := player.GetBodyIdRangeStart()
      for i := 0; i < bodyIdsPerPlayer; i++ {
        uuint16.Return(rangeStart + uint16(i))
      }

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
